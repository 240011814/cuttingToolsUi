package model

import "time"

// UserExperience 用户经历, 与聊天会话一对一(history_id 唯一)
type UserExperience struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"index"`
	HistoryID    *uint      `json:"history_id" gorm:"uniqueIndex"`
	Category     string     `json:"category" gorm:"size:32"`
	Title        string     `json:"title" gorm:"size:255"`
	Content      string     `json:"content" gorm:"type:text"`
	OccurredAt   *time.Time `json:"occurred_at"`
	Tags         string     `json:"tags" gorm:"type:json"` // JSON array
	Confidence   float64    `json:"confidence"`
	Status       string     `json:"status" gorm:"size:20;default:active"`
	IsUserEdited bool       `json:"is_user_edited"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (UserExperience) TableName() string {
	return "user_experiences"
}

type CreateUserExperienceRequest struct {
	Category   string     `json:"category"`
	Title      string     `json:"title" binding:"required"`
	Content    string     `json:"content"`
	OccurredAt *time.Time `json:"occurred_at"`
	Tags       []string   `json:"tags"`
}

type UpdateUserExperienceRequest struct {
	Category   *string     `json:"category"`
	Title      *string     `json:"title"`
	Content    *string     `json:"content"`
	OccurredAt *time.Time  `json:"occurred_at"`
	Tags       *[]string   `json:"tags"`
	Status     *string     `json:"status"`
}