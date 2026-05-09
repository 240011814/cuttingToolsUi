package service

import (
	"backend/model"
)

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

func (s *HistoryService) ListHistory(userID uint, page, pageSize int, title string, isFavorite *bool) ([]model.TrainingHistory, int64, error) {
	var histories []model.TrainingHistory
	var total int64

	query := DB.Model(&model.TrainingHistory{}).Where("user_id = ?", userID)

	if title != "" {
		fuzzy := "%" + title + "%"
		query = query.Where("title LIKE ? OR messages LIKE ?", fuzzy, fuzzy)
	}
	if isFavorite != nil {
		query = query.Where("is_favorite = ?", *isFavorite)
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

func (s *HistoryService) GetHistoryByID(userID uint, historyID uint) (*model.TrainingHistory, error) {
	var history model.TrainingHistory
	if err := DB.Where("id = ? AND user_id = ?", historyID, userID).First(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (s *HistoryService) UpdateFavorite(userID uint, historyID uint, isFavorite bool) error {
	return DB.Model(&model.TrainingHistory{}).Where("id = ? AND user_id = ?", historyID, userID).Update("is_favorite", isFavorite).Error
}

func (s *HistoryService) UpdateTitle(userID uint, historyID uint, title string) error {
	return DB.Model(&model.TrainingHistory{}).Where("id = ? AND user_id = ?", historyID, userID).Update("title", title).Error
}

func (s *HistoryService) DeleteHistory(userID uint, historyID uint) error {
	return DB.Where("id = ? AND user_id = ?", historyID, userID).Delete(&model.TrainingHistory{}).Error
}

func (s *HistoryService) SaveHistory(userID uint, historyID uint, trainingType string, title string, messages string, isFavorite bool) (uint, error) {
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
		history.IsFavorite = isFavorite
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
		IsFavorite:   isFavorite,
		Messages:     messages,
	}
	if err := DB.Create(&history).Error; err != nil {
		return 0, err
	}
	return history.ID, nil
}
