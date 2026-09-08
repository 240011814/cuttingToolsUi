package model

import "time"

type Reminder struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"userId"`
	Title           string     `gorm:"size:255;not null" json:"title"`
	Content         string     `gorm:"type:text" json:"content"`
	RemindAt        time.Time  `gorm:"not null;index" json:"remindAt"`
	Notified        bool       `gorm:"default:false;index" json:"notified"`
	RepeatType      string     `gorm:"size:20;default:'none'" json:"repeatType"`
	RepeatInterval  int        `gorm:"default:1" json:"repeatInterval"`
	RepeatEndAt     *time.Time `json:"repeatEndAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (Reminder) TableName() string {
	return "reminders"
}

type CreateReminderRequest struct {
	Title          string     `json:"title" binding:"required"`
	Content        string     `json:"content"`
	RemindAt       time.Time  `json:"remindAt" binding:"required"`
	RepeatType     string     `json:"repeatType"`
	RepeatInterval int        `json:"repeatInterval"`
	RepeatEndAt    *time.Time `json:"repeatEndAt"`
}

type UpdateReminderRequest struct {
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	RemindAt       time.Time  `json:"remindAt"`
	RepeatType     string     `json:"repeatType"`
	RepeatInterval int        `json:"repeatInterval"`
	RepeatEndAt    *time.Time `json:"repeatEndAt"`
}
