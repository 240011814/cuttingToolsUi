package service

import (
	"backend/model"
)

type AIAgentService struct{}

func NewAIAgentService() *AIAgentService {
	return &AIAgentService{}
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

	return DB.Model(&model.AIAgent{}).Where("id = ? AND user_id = ?", agentID, userID).Updates(updates).Error
}

func (s *AIAgentService) DeleteAIAgent(userID uint, agentID uint) error {
	return DB.Where("id = ? AND user_id = ? AND is_public = ?", agentID, userID, false).
		Delete(&model.AIAgent{}).Error
}
