package model

import (
	"encoding/json"
	"time"
)

const (
	JobRunStatusRunning = "running"
	JobRunStatusSuccess = "success"
	JobRunStatusFailed  = "failed"
	JobRunStatusSkipped = "skipped"

	JobTriggerScheduler = "scheduler"
	JobTriggerManual    = "manual"

	// ScheduleType 调度类型
	ScheduleTypeCron   = "cron"   // cron 表达式周期调度(系统任务)
	ScheduleTypeOnce   = "once"   // 指定时间点执行一次(不重复的备忘)
	ScheduleTypeRepeat = "repeat" // 指定锚点时间 + 重复周期(重复备忘)
)

// 用户备忘专用任务
const TaskNameReminder = "reminder.notify"

// JobDefinition 定时任务定义
type JobDefinition struct {
	ID             uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ChainID        uint            `gorm:"column:chain_id;index" json:"chainId"`
	UserID         *uint           `gorm:"index" json:"userId"`
	Name           string          `gorm:"size:100;not null;uniqueIndex" json:"name"`
	TaskName       string          `gorm:"size:100;not null" json:"taskName"`
	ScheduleType   string          `gorm:"column:schedule_type;size:20;not null;default:cron" json:"scheduleType"`
	CronExpr       string          `gorm:"column:cron_expr;size:64" json:"cronExpr"`
	RunAt          *time.Time      `gorm:"column:run_at" json:"runAt"`
	AdvanceMinutes int             `gorm:"default:0" json:"advanceMinutes"`
	RepeatType     string          `gorm:"size:20" json:"repeatType"`
	RepeatInterval int             `gorm:"default:1" json:"repeatInterval"`
	RepeatEndAt    *time.Time      `gorm:"column:repeat_end_at" json:"repeatEndAt"`
	Params         json.RawMessage `gorm:"type:json" json:"params"`
	Enabled        bool            `gorm:"default:false" json:"enabled"`
	MaxRetries     int             `gorm:"default:0" json:"maxRetries"`
	Remark         string          `gorm:"size:255" json:"remark"`
	CreatedBy      *uint           `json:"createdBy"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func (JobDefinition) TableName() string {
	return "job_definitions"
}

// JobRun 定时任务执行历史
type JobRun struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Definition  uint       `gorm:"column:definition_id;not null;index" json:"definitionId"`
	Attempt     int        `gorm:"default:1" json:"attempt"`
	Status      string     `gorm:"size:20;default:running" json:"status"`
	TriggerType string     `gorm:"size:20;default:scheduler" json:"triggerType"`
	StartedAt   *time.Time `json:"startedAt"`
	FinishedAt  *time.Time `json:"finishedAt"`
	Error       string     `gorm:"type:text" json:"error"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func (JobRun) TableName() string {
	return "job_runs"
}

// TaskMeta 已注册的可调度任务
type TaskMeta struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	ParamsExample json.RawMessage `json:"paramsExample"`
}

type CreateJobDefinitionRequest struct {
	Name       string          `json:"name" binding:"required"`
	TaskName   string          `json:"taskName" binding:"required"`
	CronExpr   string          `json:"cronExpr" binding:"required"`
	Params     json.RawMessage `json:"params"`
	Enabled    *bool           `json:"enabled"`
	MaxRetries int             `json:"maxRetries"`
	Remark     string          `json:"remark"`
}

type UpdateJobDefinitionRequest struct {
	Name       *string         `json:"name"`
	TaskName   *string         `json:"taskName"`
	CronExpr   *string         `json:"cronExpr"`
	Params     json.RawMessage `json:"params"`
	Enabled    *bool           `json:"enabled"`
	MaxRetries *int            `json:"maxRetries"`
	Remark     *string         `json:"remark"`
}
