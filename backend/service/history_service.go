package service

import (
	"backend/model"
)

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

func (s *HistoryService) ListHistory(userID uint, page, pageSize int, title, recordType string) ([]model.TrainingHistory, int64, error) {
	var histories []model.TrainingHistory
	var total int64

	query := DB.Model(&model.TrainingHistory{}).Where("user_id = ?", userID)

	if title != "" {
		fuzzy := "%" + title + "%"
		query = query.Where("title LIKE ? OR messages LIKE ?", fuzzy, fuzzy)
	}
	if recordType != "" {
		query = query.Where("record_type = ?", recordType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (s *HistoryService) SaveHistory(userID uint, historyID uint, trainingType string, title string, messages string, recordType string) (uint, error) {
	if historyID > 0 {
		// Update existing
		var history model.TrainingHistory
		if err := DB.First(&history, historyID).Error; err != nil {
			return 0, err
		}
		if history.UserID != userID {
			return 0, nil // unauthorized, ignore
		}
		history.Messages = messages
		history.Title = title
		history.RecordType = recordType
		if err := DB.Save(&history).Error; err != nil {
			return 0, err
		}
		return historyID, nil
	}

	// Create new
	history := model.TrainingHistory{
		UserID:       userID,
		TrainingType: trainingType,
		Title:        title,
		RecordType:   recordType,
		Messages:     messages,
	}
	if err := DB.Create(&history).Error; err != nil {
		return 0, err
	}
	return history.ID, nil
}
