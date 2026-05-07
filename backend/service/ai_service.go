package service

import (
	"backend/model"
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// AIService 提供与 AI 相关的业务逻辑
type AIService struct {
	mu             sync.RWMutex
	activeProvider *model.AIProvider
	activeModel    *model.AIModel
	enabledModels  []model.AIModel
	client         *openai.Client
}

// NewAIService 初始化 AIService
func NewAIService() (*AIService, error) {
	s := &AIService{}
	if err := s.ReloadConfig(); err != nil {
		// 初始加载失败不阻塞启动，但记录日志
		return s, nil
	}
	return s, nil
}

// ReloadConfig 重新从数据库加载配置到缓存
func (s *AIService) ReloadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

// getActiveConfig 从缓存获取当前启用的 Provider 和默认 Model
func (s *AIService) getActiveConfig() (*model.AIProvider, *model.AIModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeProvider == nil {
		log.Printf("未找到启用的 AI 提供商配置，请联系管理员在后台配置")
		return nil, nil, errors.New("未找到启用的 AI 提供商配置，请联系管理员在后台配置")
	}
	if s.activeModel == nil {
		log.Printf("该 AI 提供商下未配置任何模型")
		return s.activeProvider, nil, errors.New("该 AI 提供商下未配置任何模型")
	}
	log.Printf("s.activeProviderURL: %s, model_code: %s", s.activeProvider.BaseURL, s.activeModel.ModelCode)

	return s.activeProvider, s.activeModel, nil
}

// getClient 根据配置获取或创建 OpenAI 客户端
func (s *AIService) getClient(provider *model.AIProvider) *openai.Client {
	config := openai.DefaultConfig(provider.APIKey)
	if provider.BaseURL != "" {
		config.BaseURL = provider.BaseURL
	}
	return openai.NewClientWithConfig(config)
}

// ChatStream 向 AI 发起请求并返回流
func (s *AIService) ChatStream(ctx context.Context, messages []openai.ChatCompletionMessage, modelOverride string) (*openai.ChatCompletionStream, error) {
	s.mu.RLock()
	provider := s.activeProvider
	m := s.activeModel
	client := s.client
	s.mu.RUnlock()

	if provider == nil || client == nil {
		return nil, errors.New("未找到启用的 AI 提供商配置")
	}
	if m == nil && modelOverride == "" {
		return nil, errors.New("未配置可用模型")
	}

	selectedModel := m.ModelCode
	if modelOverride != "" {
		selectedModel = modelOverride
	}

	req := openai.ChatCompletionRequest{
		Model:    selectedModel,
		Messages: messages,
		Stream:   true,
	}

	// 解析配置中的运行参数
	if m != nil && m.ConfigJSON != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(m.ConfigJSON), &configMap); err == nil {
			if t, ok := configMap["temperature"].(float64); ok {
				req.Temperature = float32(t)
			}
			if topP, ok := configMap["top_p"].(float64); ok {
				req.TopP = float32(topP)
			}
			if maxTokens, ok := configMap["max_tokens"].(float64); ok {
				req.MaxTokens = int(maxTokens)
			}
		}
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// ListEnabledModels 从缓存获取所有已启用 Provider 的模型列表
func (s *AIService) ListEnabledModels() ([]model.AIModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.enabledModels == nil {
		return []model.AIModel{}, nil
	}
	return s.enabledModels, nil
}
