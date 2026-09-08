package model

import (
	"encoding/json"
	"time"
)

type Job struct {
	ID             uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	JobType        string          `gorm:"size:50;not null;index" json:"jobType"`
	JobID          uint            `gorm:"not null;index" json:"jobId"`
	UserID         *uint           `gorm:"index" json:"userId"`
	ScheduledAt    time.Time       `gorm:"not null;index" json:"scheduledAt"`
	Status         string          `gorm:"size:20;default:'pending';index" json:"status"`
	RetryCount     int             `gorm:"default:0" json:"retryCount"`
	MaxRetries     int             `gorm:"default:3" json:"maxRetries"`
	LastError      string          `gorm:"type:text" json:"lastError"`
	Params         json.RawMessage `gorm:"type:json" json:"params"`
	RepeatType     string          `gorm:"size:20;default:'none'" json:"repeatType"`
	RepeatInterval int             `gorm:"default:1" json:"repeatInterval"`
	RepeatEndAt    *time.Time      `json:"repeatEndAt"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func (Job) TableName() string {
	return "jobs"
}

const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"

	JobTypeReminder = "reminder"
)

type JobCallback func(job Job) error

type ReminderParams struct {
	Title   string `json:"title"`
	Content string `json:"content"`
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