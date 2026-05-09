package service

import (
	"backend/model"
)

type CustomTrainingService struct{}

func NewCustomTrainingService() *CustomTrainingService {
	return &CustomTrainingService{}
}

func (s *CustomTrainingService) ListCustomTrainings(userID uint) ([]model.CustomTraining, error) {
	var trainings []model.CustomTraining
	if err := DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&trainings).Error; err != nil {
		return nil, err
	}
	return trainings, nil
}

func (s *CustomTrainingService) GetCustomTrainingByID(userID uint, trainingID uint) (*model.CustomTraining, error) {
	var training model.CustomTraining
	if err := DB.Where("id = ? AND user_id = ?", trainingID, userID).First(&training).Error; err != nil {
		return nil, err
	}
	return &training, nil
}

func (s *CustomTrainingService) CreateCustomTraining(userID uint, req model.CreateCustomTrainingRequest) (*model.CustomTraining, error) {
	training := model.CustomTraining{
		UserID:           userID,
		Title:            req.Title,
		Description:      req.Description,
		SystemPrompt:     req.SystemPrompt,
		Icon:             req.Icon,
		Color:            req.Color,
		InitialMessage:   req.InitialMessage,
		InputPlaceholder: req.InputPlaceholder,
		SpeechLang:       req.SpeechLang,
		SpeechRate:       req.SpeechRate,
	}

	if training.Icon == "" {
		training.Icon = "mdi:robot-outline"
	}
	if training.Color == "" {
		training.Color = "#2080f0"
	}
	if training.InputPlaceholder == "" {
		training.InputPlaceholder = "输入消息... (回车发送，Shift + 回车换行)"
	}
	if training.SpeechLang == "" {
		training.SpeechLang = "zh-CN"
	}
	if training.SpeechRate == 0 {
		training.SpeechRate = 0.95
	}

	if err := DB.Create(&training).Error; err != nil {
		return nil, err
	}
	return &training, nil
}

func (s *CustomTrainingService) UpdateCustomTraining(userID uint, trainingID uint, req model.UpdateCustomTrainingRequest) error {
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

	return DB.Model(&model.CustomTraining{}).Where("id = ? AND user_id = ?", trainingID, userID).Updates(updates).Error
}

func (s *CustomTrainingService) DeleteCustomTraining(userID uint, trainingID uint) error {
	return DB.Where("id = ? AND user_id = ?", trainingID, userID).Delete(&model.CustomTraining{}).Error
}
