package model

import "time"

type TrainingHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`
	TrainingType string    `json:"training_type"`
	Title        string    `json:"title"`
	RecordType   string    `json:"record_type"`
	Messages     string    `json:"messages"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListHistoryRequest struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=100"`
	Title        string `form:"title"`
	RecordType   string `form:"record_type"`
}

type ListHistoryResponse struct {
	Total int64             `json:"total"`
	Items []TrainingHistory `json:"items"`
}

type ArchiveHistoryRequest struct {
	TrainingType string `json:"training_type" binding:"required"`
	Title        string `json:"title"`
	Messages     string `json:"messages" binding:"required"`
}
