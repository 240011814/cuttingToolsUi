package service

import (
	"backend/model"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/sashabaranov/go-openai"
)

type AgentService struct {
	ctx 					context.Context
	activeProvider *model.AIProvider
	activeModel    *model.AIModel
	enabledModels  []model.AIModel
	client         *openai.Client
	timeout        time.Duration
	timeoutConfig  TimeoutConfig
}

func NewAgentService(timeoutMinutes int, sysCfgService *SystemConfigService) (*AgentService, error) {
	if timeoutMinutes <= 0 {
		timeoutMinutes = 5
	}
	s := &AgentService{
		timeout: time.Duration(timeoutMinutes) * time.Minute,
	}
	// 从数据库读取超时配置
	if sysCfgService != nil {
		s.timeoutConfig = sysCfgService.GetTimeoutConfig()
		// 如果数据库中有AI请求超时配置，使用数据库值覆盖
		if s.timeoutConfig.AIRequestTimeout > 0 {
			s.timeout = time.Duration(s.timeoutConfig.AIRequestTimeout) * time.Minute
		}
	}
	if err := s.ReloadConfig(); err != nil {
		// 初始加载失败不阻塞启动，但记录日志
		return s, nil
	}
	return s, nil
}

func (s *AgentService) ReloadConfig() error {

	var provider model.AIProvider
	if err := DB.Where("is_active = ?", true).First(&provider).Error; err != nil {
		s.activeProvider = nil
		s.activeModel = nil
		s.enabledModels = nil
		return err
	}
	s.activeProvider = &provider

	var m model.AIModel
	if err := DB.Where("provider_id = ? AND is_default = ?", provider.ID, true).First(&m).Error; err != nil {
		if err := DB.Where("provider_id = ?", provider.ID).First(&m).Error; err != nil {
			s.activeModel = nil
		} else {
			s.activeModel = &m
		}
	} else {
		s.activeModel = &m
	}

	// 更新客户端缓存
	if s.activeProvider != nil {
		config := openai.DefaultConfig(s.activeProvider.APIKey)
		if s.activeProvider.BaseURL != "" {
			config.BaseURL = s.activeProvider.BaseURL
		}
		config.HTTPClient = newHTTPClient(s.timeout, s.timeoutConfig)
		s.client = openai.NewClientWithConfig(config)
	}

	// 加载所有启用 Provider 的模型列表
	var providers []model.AIProvider
	if err := DB.Where("is_active = ?", true).Find(&providers).Error; err == nil {
		providerIDs := make([]int, 0, len(providers))
		for _, p := range providers {
			providerIDs = append(providerIDs, p.ID)
		}
		var models []model.AIModel
		if err := DB.Where("provider_id IN ?", providerIDs).Order("is_default DESC, id ASC").Find(&models).Error; err == nil {
			s.enabledModels = models
		}
	}

	return nil
}

func (s *AgentService)GetRunerByCode(agentCode string) (*adk.Runner, error) {
	model, err := s.getModel()
	if err != nil {
		log.Printf("Failed to get model: %v", err)
		return nil, err
	}
  agent, err := adk.NewChatModelAgent(s.ctx, &adk.ChatModelAgentConfig{
        Name:        "my-assistant",
        Description: "一个可以使用工具回答问题的助手。",
        Instruction: "你是一个有帮助的助手。请根据可用工具回答用户问题。",
        Model:       model,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{
                    // 注册你的工具，例如 webSearchTool
                },
            },
        },
        // Handlers: []adk.ChatModelAgentMiddleware{...}, // 注册 Middleware
    })
    if err != nil {
        log.Fatal(err)
    }

    // 3. 通过 Runner 执行 Agent
    runner := adk.NewRunner(s.ctx, adk.RunnerConfig{
        Agent:           agent,
        EnableStreaming: true,
    })
		return runner, nil
}

func (s *AgentService) getModel() (*ark.ChatModel, error) {
	chatConfig :=&ark.ChatModelConfig{
		Model:  s.activeModel.ModelCode,
		APIKey: s.activeProvider.APIKey,
		BaseURL: s.activeProvider.BaseURL, 
	}
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(s.activeModel.ConfigJSON), &configMap); err == nil {
			if t, ok := configMap["temperature"].(float64); ok {
				 temperature := float32(t)
				 chatConfig.Temperature = &temperature
			}
			if topP, ok := configMap["top_p"].(float64); ok {
				topP := float32(topP)
				chatConfig.TopP = &topP
			}
			if maxTokens, ok := configMap["max_tokens"].(float64); ok {
				maxTokens := int(maxTokens)
				chatConfig.MaxTokens = &maxTokens
			}
			if frequencyPenalty, ok := configMap["frequency_penalty"].(float64); ok {
				frequencyPenalty := float32(frequencyPenalty)
				chatConfig.FrequencyPenalty = &frequencyPenalty
			}
			if presencePenalty, ok := configMap["presence_penalty"].(float64); ok {
				presencePenalty := float32(presencePenalty)
				chatConfig.PresencePenalty = &presencePenalty
			}
		} else {
			log.Printf("AI model config_json parse failed model=%s config_json=%s err=%v", s.activeModel.ModelCode, s.activeModel.ConfigJSON, err)
		}
	return ark.NewChatModel(s.ctx, chatConfig)
}