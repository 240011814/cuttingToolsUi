package model

import (
	"time"
)

// Note 笔记实体
type Note struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Category  string    `gorm:"size:100;not null;index" json:"category"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Note) TableName() string {
	return "notes"
}

// CreateNoteRequest 创建笔记请求
type CreateNoteRequest struct {
	Category string `json:"category" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// UpdateNoteRequest 更新笔记请求
type UpdateNoteRequest struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}
