package service

import (
	"backend/model"
	"backend/service/tools"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

type AIAgentService struct{
		ctx 					context.Context
	activeProvider *model.AIProvider
	activeModel    *model.AIModel
	enabledModels  []model.AIModel
	timeout        time.Duration
	timeoutConfig  TimeoutConfig
	runnerCache    map[string]*adk.Runner
	sysCfgService  *SystemConfigService
}

func NewAIAgentService(timeoutMinutes int, sysCfgService *SystemConfigService) (*AIAgentService, error) {
		if timeoutMinutes <= 0 {
		timeoutMinutes = 5
	}
	s := &AIAgentService{
		ctx:          context.Background(),
		timeout:      time.Duration(timeoutMinutes) * time.Minute,
		runnerCache:  make(map[string]*adk.Runner),
		sysCfgService: sysCfgService,
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

func (s *AIAgentService) ListAvailableAgents(userID uint) ([]model.AIAgent, error) {
	var agents []model.AIAgent
	if err := DB.Where("is_public = ? OR user_id = ?", true, userID).
		Order("is_public DESC, created_at DESC").
		Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (s *AIAgentService) GetAIAgentByID(userID uint, agentID uint) (*model.AIAgent, error) {
	var agent model.AIAgent
	if err := DB.Where("(is_public = ? OR user_id = ?) AND id = ?", true, userID, agentID).
		First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *AIAgentService) CreateAIAgent(userID uint, req model.CreateAIAgentRequest) (*model.AIAgent, error) {
	agent := model.AIAgent{
		UserID:           userID,
		IsPublic:         false,
		Title:            req.Title,
		Description:      req.Description,
		Code:             req.Code,
		SystemPrompt:     req.SystemPrompt,
		Icon:             req.Icon,
		Color:            req.Color,
		InitialMessage:   req.InitialMessage,
		InputPlaceholder: req.InputPlaceholder,
		SpeechLang:       req.SpeechLang,
		SpeechRate:       req.SpeechRate,
	}

	if agent.Icon == "" {
		agent.Icon = "mdi:robot-outline"
	}
	if agent.Color == "" {
		agent.Color = "#2080f0"
	}
	if agent.InputPlaceholder == "" {
		agent.InputPlaceholder = "输入消息... (回车发送，Shift + 回车换行)"
	}
	if agent.SpeechLang == "" {
		agent.SpeechLang = "zh-CN"
	}
	if agent.SpeechRate == 0 {
		agent.SpeechRate = 0.95
	}

	if err := DB.Create(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *AIAgentService) UpdateAIAgent(userID uint, agentID uint, req model.UpdateAIAgentRequest) error {
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SystemPrompt != "" {
		updates["system_prompt"] = req.SystemPrompt
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}
	if req.InitialMessage != "" {
		updates["initial_message"] = req.InitialMessage
	}
	if req.InputPlaceholder != "" {
		updates["input_placeholder"] = req.InputPlaceholder
	}
	if req.SpeechLang != "" {
		updates["speech_lang"] = req.SpeechLang
	}
	if req.SpeechRate != 0 {
		updates["speech_rate"] = req.SpeechRate
	}

	err := DB.Model(&model.AIAgent{}).Where("id = ? AND user_id = ?", agentID, userID).Updates(updates).Error
	if err == nil {
		s.clearRunnerCache(userID, agentID)
	}
	return err
}

func (s *AIAgentService) DeleteAIAgent(userID uint, agentID uint) error {
	err := DB.Where("id = ? AND user_id = ? AND is_public = ?", agentID, userID, false).
		Delete(&model.AIAgent{}).Error
	if err == nil {
		s.clearRunnerCache(userID, agentID)
	}
	return err
}

func (s *AIAgentService) clearRunnerCache(userID uint, agentID uint) {
	prefix := fmt.Sprintf("%d_%d", userID, agentID)
	for key := range s.runnerCache {
		if strings.HasPrefix(key, prefix) {
			delete(s.runnerCache, key)
		}
	}
}


func (s *AIAgentService) ReloadConfig() error {

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

	s.runnerCache = make(map[string]*adk.Runner)
	return nil
}

func (s *AIAgentService) ListEnabledModels() ([]model.AIModel, error) {
	var providers []model.AIProvider
	if err := DB.Where("is_active = ?", true).Find(&providers).Error; err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return []model.AIModel{}, nil
	}
	providerIDs := make([]int, 0, len(providers))
	for _, p := range providers {
		providerIDs = append(providerIDs, p.ID)
	}
	var models []model.AIModel
	if err := DB.Where("provider_id IN ?", providerIDs).Order("is_default DESC, id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *AIAgentService) TestConnection(apiKey, baseURL, modelCode string) error {
	if apiKey == "" {
		return errors.New("API Key 不能为空")
	}
	testModel := modelCode
	if testModel == "" {
		testModel = "gpt-3.5-turbo"
	}
	chatModel, err := ark.NewChatModel(s.ctx, &ark.ChatModelConfig{
		Model:   testModel,
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return err
	}
	testAgent, err := adk.NewChatModelAgent(s.ctx, &adk.ChatModelAgentConfig{
		Name:        "test-connection",
		Description: "test connection agent",
		Instruction: "You are a test bot. Reply OK.",
		Model:       chatModel,
	})
	if err != nil {
		return err
	}
	runner := adk.NewRunner(s.ctx, adk.RunnerConfig{
		Agent:           testAgent,
		EnableStreaming: false,
	})
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()
	iter := runner.Run(ctx, []*schema.Message{
		{Role: schema.User, Content: "Hello"},
	})
	for {
		_, ok := iter.Next()
		if !ok {
			break
		}
	}
	return nil
}

func (s *AIAgentService) buildTools() []tool.BaseTool {
	var toolsList []tool.BaseTool

	if s.sysCfgService != nil {
		cfg := s.sysCfgService.GetWebSearchConfig()
		if cfg.APIKey != "" {
			webSearchTool, err := tools.NewWebSearchTool(tools.WebSearchConfig{
				APIKey: cfg.APIKey,
			})
			if err != nil {
				log.Printf("Failed to create web search tool: %v", err)
			} else {
				toolsList = append(toolsList, webSearchTool)
			}
		}
	}

	return toolsList
}

func (s *AIAgentService)GetRunerByAgentID(userID uint, agentID uint, modelOverride string) (*adk.Runner, error) {
	cacheKey := fmt.Sprintf("%d_%d_%s", userID, agentID, modelOverride)
	if runner, ok := s.runnerCache[cacheKey]; ok {
		return runner, nil
	}
	
	chatModel, err := s.getModel(modelOverride)
	if err != nil {
		log.Printf("Failed to get model: %v", err)
		return nil, err
	}
	userAgent, err := s.GetAIAgentByID(userID, agentID)
	if err != nil {
		log.Printf("Failed to get AI agent: %v", err)
		return nil, err
	}
	var userPrompt model.UserPrompt
	systemPrompt := userAgent.SystemPrompt
	err = DB.Where("user_id = ? AND agent_id = ? AND is_active = ?", userID, agentID, true).First(&userPrompt).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		systemPrompt = userPrompt.CustomPrompt
	}
  agent, err := adk.NewChatModelAgent(s.ctx, &adk.ChatModelAgentConfig{
        Name:        userAgent.Code,
        Description: userAgent.Description,
        Instruction: systemPrompt,
        Model:       chatModel,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: s.buildTools(),
            },
        },
        // Handlers: []adk.ChatModelAgentMiddleware{...}, // 注册 Middleware
    })
    if err != nil {
        log.Fatal(err)
    }

    runner := adk.NewRunner(s.ctx, adk.RunnerConfig{
        Agent:           agent,
        EnableStreaming: true,
    })
    s.runnerCache[cacheKey] = runner
		return runner, nil
}

func (s *AIAgentService) getModel(modelOverride string) (*ark.ChatModel, error) {
	modelCode := s.activeModel.ModelCode
	if modelOverride != "" {
		modelCode = modelOverride
	}
	chatConfig :=&ark.ChatModelConfig{
		Model:  modelCode,
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



func (s *AIAgentService) ChatStream(userID uint, agentID uint, messages []*schema.Message, modelOverride string) (*adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]], error) {
	runner, err := s.GetRunerByAgentID(userID, agentID, modelOverride)
	if err != nil {
		return nil, err
	}
	logHandler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if mi := einomodel.ConvCallbackInput(input); mi != nil {
				log.Printf("[eino] %s start: %d messages", info.Name, len(mi.Messages))
				for _, m := range mi.Messages {
					log.Printf("[eino]   [%s] %s", m.Role, m.Content)
				}
			} else {
				log.Printf("[eino] %s start", info.Name)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if mo := einomodel.ConvCallbackOutput(output); mo != nil && mo.Message != nil {
				if mo.Message.ResponseMeta != nil && mo.Message.ResponseMeta.Usage.TotalTokens > 0 {
					log.Printf("[eino] %s end: prompt_tokens=%d completion_tokens=%d total_tokens=%d",
						info.Name,
						mo.Message.ResponseMeta.Usage.PromptTokens,
						mo.Message.ResponseMeta.Usage.CompletionTokens,
						mo.Message.ResponseMeta.Usage.TotalTokens)
				}
				if mo.Message.Content != "" {
					log.Printf("[eino] %s end: reply=%s", info.Name, mo.Message.Content)
				}
			} else {
				log.Printf("[eino] %s end", info.Name)
			}
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			log.Printf("[eino] %s error: %v", info.Name, err)
			return ctx
		}).
		Build()
	return runner.Run(s.ctx, messages, adk.WithCallbacks(logHandler)), nil
}
