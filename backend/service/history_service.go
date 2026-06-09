package service

import (
	"backend/model"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	htmlTagRegex      = regexp.MustCompile(`<[^>]*>`)
	markdownCharRegex = regexp.MustCompile("[#*`~_]")
)

// truncateUTF8 按 UTF-8 字符边界安全截断，不会切碎多字节字符
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	for maxBytes > 0 && !utf8.RuneStart(s[maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

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

	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (s *HistoryService) GetHistoryByID(userID uint, historyID uint) (*model.TrainingHistory, error) {
	var history model.TrainingHistory
	if err := DB.Where("id = ? AND user_id = ?", historyID, userID).First(&history).Error; err != nil {
		return nil, err
	}

	var messages []model.TrainingMessage
	if err := DB.Where("history_id = ?", historyID).Order("sort_order ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	history.Messages = messages

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

func extractLastMessage(messages []model.OpenAIMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "system" && messages[i].Content != "" {
			content := messages[i].Content
			content = htmlTagRegex.ReplaceAllString(content, "")
			content = markdownCharRegex.ReplaceAllString(content, "")
			content = strings.Join(strings.Fields(content), " ")
			if len(content) > 200 {
				content = truncateUTF8(content, 200)
			}
			return content
		}
	}
	return ""
}

func (s *HistoryService) SaveHistory(userID uint, historyID uint, trainingType string, customTrainingID *uint, title string, messages []model.OpenAIMessage, isFavorite bool) (uint, error) {
	lastMessage := extractLastMessage(messages)

	if historyID > 0 {
		var history model.TrainingHistory
		if err := DB.First(&history, historyID).Error; err != nil {
			return 0, err
		}
		if history.UserID != userID {
			return 0, nil
		}
		history.Title = title
		history.IsFavorite = isFavorite
		history.LastMessage = lastMessage
		if err := DB.Save(&history).Error; err != nil {
			return 0, err
		}

		if err := DB.Where("history_id = ?", historyID).Delete(&model.TrainingMessage{}).Error; err != nil {
			return 0, err
		}

		msgs := make([]model.TrainingMessage, len(messages))
		for i, m := range messages {
			msgs[i] = model.TrainingMessage{
				HistoryID: historyID,
				Role:      m.Role,
				Content:   m.Content,
				SortOrder: i,
			}
		}
		if len(msgs) > 0 {
			if err := DB.Create(&msgs).Error; err != nil {
				return 0, err
			}
		}

		return historyID, nil
	}

	history := model.TrainingHistory{
		UserID:           userID,
		TrainingType:     trainingType,
		CustomTrainingID: customTrainingID,
		Title:            title,
		IsFavorite:       isFavorite,
		LastMessage:      lastMessage,
	}
	if err := DB.Create(&history).Error; err != nil {
		return 0, err
	}

	msgs := make([]model.TrainingMessage, len(messages))
	for i, m := range messages {
		msgs[i] = model.TrainingMessage{
			HistoryID: history.ID,
			Role:      m.Role,
			Content:   m.Content,
			SortOrder: i,
		}
	}
	if len(msgs) > 0 {
		if err := DB.Create(&msgs).Error; err != nil {
			return 0, err
		}
	}

	return history.ID, nil
}
