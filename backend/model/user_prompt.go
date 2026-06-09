package model

import "time"

// UserPrompt 用户自定义提示词模型
type UserPrompt struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"not null;index:idx_user_module" json:"user_id"`
	ModuleKey         string    `gorm:"size:50;not null;index:idx_user_module" json:"module_key"`
	CustomPrompt      string    `gorm:"type:text;not null" json:"custom_prompt"`
	MemorySearchQuery string    `gorm:"size:500;default:''" json:"memory_search_query"`
	Version           int       `gorm:"not null;default:1" json:"version"`
	IsActive          bool      `gorm:"default:false" json:"is_active"`
	Remark            string    `gorm:"size:255" json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
