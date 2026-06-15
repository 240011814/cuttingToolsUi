package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONValue 自定义 JSON 类型，支持数据库 JSON 列的读写
type JSONValue json.RawMessage

func (j JSONValue) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONValue) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = append((*j)[0:0], []byte(v)...)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}

func (j JSONValue) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONValue) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONValue: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// UserPreference 用户偏好设置实体
type UserPreference struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_user_key" json:"userId"`
	PrefKey   string    `gorm:"size:100;not null;uniqueIndex:uk_user_key" json:"prefKey"`
	PrefValue JSONValue `gorm:"type:json" json:"prefValue"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (UserPreference) TableName() string {
	return "user_preferences"
}

// SavePreferenceRequest 保存偏好设置请求
type SavePreferenceRequest struct {
	PrefValue interface{} `json:"prefValue" binding:"required"`
}
