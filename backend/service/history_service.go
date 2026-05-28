package service

import (
	"backend/model"
	"encoding/json"
	"regexp"
	"strings"
)

var (
	htmlTagRegex      = regexp.MustCompile(`<[^>]*>`)
	markdownCharRegex = regexp.MustCompile("[#*`~_]")
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
		query = query.Where("title LIKE ?", fuzzy)
	}
	if isFavorite != nil {
		query = query.Where("is_favorite = ?", *isFavorite)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Select("id", "user_id", "training_type", "custom_training_id", "title", "is_favorite", "last_message", "created_at", "updated_at").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error; err != nil {
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

// extractLastMessage extracts the last non-system message content from a JSON messages array,
// strips HTML/Markdown tags, and truncates to 200 characters.
func extractLastMessage(messagesJSON string) string {
	var msgs []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(messagesJSON), &msgs); err != nil || len(msgs) == 0 {
		return ""
	}
	// Find last non-system message
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != "system" && msgs[i].Content != "" {
			content := msgs[i].Content
			// Strip HTML tags
			content = htmlTagRegex.ReplaceAllString(content, "")
			// Strip Markdown formatting characters
			content = markdownCharRegex.ReplaceAllString(content, "")
			// Normalize whitespace
			content = strings.Join(strings.Fields(content), " ")
			if len(content) > 200 {
				content = content[:200]
			}
			return content
		}
	}
	return ""
}

func (s *HistoryService) SaveHistory(userID uint, historyID uint, trainingType string, customTrainingID *uint, title string, messages string, isFavorite bool) (uint, error) {
	lastMessage := extractLastMessage(messages)

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
		history.LastMessage = lastMessage
		if err := DB.Save(&history).Error; err != nil {
			return 0, err
		}
		return historyID, nil
	}

	// Create new
	history := model.TrainingHistory{
		UserID:           userID,
		TrainingType:     trainingType,
		CustomTrainingID: customTrainingID,
		Title:            title,
		IsFavorite:       isFavorite,
		Messages:         messages,
		LastMessage:      lastMessage,
	}
	if err := DB.Create(&history).Error; err != nil {
		return 0, err
	}
	return history.ID, nil
}
