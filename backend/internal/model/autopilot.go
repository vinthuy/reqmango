package model

import (
	"encoding/json"
	"time"
)

// AutopilotTask represents a scheduled/automatic task
type AutopilotTask struct {
	BaseModel
	WorkspaceID    uint64         `gorm:"not null;index" json:"workspace_id"`
	ProjectID      *uint64        `gorm:"index" json:"project_id,omitempty"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	TriggerType    string         `gorm:"size:20;not null" json:"trigger_type"` // cron, webhook, manual
	CronExpression string         `gorm:"size:100" json:"cron_expression"`
	TriggerURL     string         `gorm:"size:255" json:"trigger_url"`
	TaskType       string         `gorm:"size:50;not null" json:"task_type"` // report, scan, sync, etc.
	AgentTemplateID *uint64       `gorm:"index" json:"agent_template_id,omitempty"`
	AgentConfigID   *uint64       `gorm:"index" json:"agent_config_id,omitempty"`
	InputData      json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_data"`
	Status         string         `gorm:"size:20;not null;default:active" json:"status"`
	LastRunAt      *time.Time     `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time     `json:"next_run_at,omitempty"`
	Config         json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config"`
	NotificationConfig json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"notification_config"`
	TimeoutSeconds int            `gorm:"default:300" json:"timeout_seconds"`
	RetryCount     int            `gorm:"default:0" json:"retry_count"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
}

// AutopilotExecution represents an execution record of an autopilot task
type AutopilotExecution struct {
	BaseModel
	TaskID         uint64         `gorm:"not null;index" json:"task_id"`
	Status         string         `gorm:"size:20;not null;default:pending" json:"status"`
	TriggerType    string         `gorm:"size:20" json:"trigger_type"`
	InputData      json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_data"`
	OutputData     json.RawMessage `gorm:"type:jsonb" json:"output_data"`
	ErrorInfo      string         `json:"error_info"`
	Logs           json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"logs"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	FailedAt       *time.Time     `json:"failed_at,omitempty"`
	DurationMs     int64          `json:"duration_ms"`
	RetryCount     int            `gorm:"default:0" json:"retry_count"`
}

// AutopilotWebhook represents a webhook endpoint for triggering autopilot tasks
type AutopilotWebhook struct {
	BaseModel
	WorkspaceID    uint64 `gorm:"not null;index" json:"workspace_id"`
	TaskID         uint64 `gorm:"not null;index" json:"task_id"`
	Endpoint       string `gorm:"size:255;not null;unique" json:"endpoint"`
	Secret         string `gorm:"size:100" json:"secret"`
	Status         string `gorm:"size:20;not null;default:active" json:"status"`
	AllowedIPs     json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"allowed_ips"`
	MaxRequests    int    `gorm:"default:100" json:"max_requests"`
	RequestCount   int    `gorm:"default:0" json:"request_count"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}
