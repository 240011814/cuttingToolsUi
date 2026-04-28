package service

import (
	"context"
	"errors"

	"backend/config"
	"github.com/sashabaranov/go-openai"
)

// AIService 提供与 AI 相关的业务逻辑
type AIService struct {
	client *openai.Client
	model  string
}

// NewAIService 初始化 AIService
func NewAIService(cfg *config.Config) (*AIService, error) {
	apiKey := cfg.DeepSeek.APIKey
	if apiKey == "" {
		// 允许不报错以便前端调试UI
	}

	baseURL := cfg.DeepSeek.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.BaseURL = baseURL

	client := openai.NewClientWithConfig(openaiConfig)

	model := cfg.DeepSeek.Model
	if model == "" {
		model = "deepseek-chat"
	}

	return &AIService{
		client: client,
		model:  model,
	}, nil
}

// ChatStream 向 DeepSeek 发起请求并返回流
func (s *AIService) ChatStream(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionStream, error) {
	if s.client == nil {
		return nil, errors.New("AI Client is not initialized")
	}

	req := openai.ChatCompletionRequest{
		Model:    s.model,
		Messages: messages,
		Stream:   true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	return stream, nil
}
