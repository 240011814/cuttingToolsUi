package model

import "time"

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TrainingHistory struct {
	ID               uint              `json:"id" gorm:"primaryKey"`
	UserID           uint              `json:"user_id"`
	TrainingType     string            `json:"training_type"`
	CustomTrainingID *uint             `json:"custom_training_id"`
	Title            string            `json:"title"`
	IsFavorite       bool              `json:"is_favorite"`
	LastMessage      string            `json:"last_message"`
	Messages         []TrainingMessage `json:"messages" gorm:"foreignKey:HistoryID"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type TrainingMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	HistoryID uint      `json:"history_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (TrainingMessage) TableName() string {
	return "training_messages"
}

type ListHistoryRequest struct {
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
	Title      string `form:"title"`
	IsFavorite *bool  `form:"is_favorite"`
}

type ListHistoryResponse struct {
	Total int64             `json:"total"`
	Items []TrainingHistory `json:"items"`
}

type UpdateFavoriteRequest struct {
	IsFavorite bool `json:"is_favorite"`
}

type UpdateTitleRequest struct {
	Title string `json:"title" binding:"required"`
}
