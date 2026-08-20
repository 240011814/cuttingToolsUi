package model

import "time"

type AITool struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:100"`
	DisplayName string    `json:"display_name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Enabled     bool      `json:"enabled"`
	ConfigJSON  string    `json:"config_json" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (AITool) TableName() string {
	return "ai_tools"
}
