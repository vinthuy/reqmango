package response

import (
	"encoding/json"
	"time"
)

// Squad DTOs
type SquadResponse struct {
	ID             uint64     `json:"id"`
	WorkspaceID    uint64     `json:"workspace_id"`
	ProjectID      *uint64    `json:"project_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	LeaderAgentID  *uint64    `json:"leader_agent_id"`
	Status         string     `json:"status"`
	Goal           string     `json:"goal"`
	Members        []SquadMemberResponse `json:"members"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type SquadMemberResponse struct {
	ID             uint64     `json:"id"`
	SquadID        uint64     `json:"squad_id"`
	AgentID        uint64     `json:"agent_id"`
	Role           string     `json:"role"`
	AgentConfigID  uint64     `json:"agent_config_id"`
	Status         string     `json:"status"`
	AssignedAt     time.Time  `json:"assigned_at"`
	RemovedAt      *time.Time `json:"removed_at"`
}

type SquadExecutionResponse struct {
	ID           uint64          `json:"id"`
	SquadID      uint64          `json:"squad_id"`
	Status       string          `json:"status"`
	Goal         string          `json:"goal"`
	InputData    json.RawMessage `json:"input_data"`
	OutputData   json.RawMessage `json:"output_data"`
	Logs         json.RawMessage `json:"logs"`
	StartedAt    *time.Time      `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at"`
	FailedAt     *time.Time      `json:"failed_at"`
	ErrorInfo    string          `json:"error_info"`
	CreatedAt    time.Time       `json:"created_at"`
}

// Autopilot DTOs
type AutopilotTaskResponse struct {
	ID                uint64          `json:"id"`
	WorkspaceID       uint64          `json:"workspace_id"`
	ProjectID         *uint64         `json:"project_id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	TriggerType       string          `json:"trigger_type"`
	CronExpression    string          `json:"cron_expression"`
	TriggerURL        string          `json:"trigger_url"`
	TaskType          string          `json:"task_type"`
	AgentTemplateID   *uint64         `json:"agent_template_id"`
	AgentConfigID     *uint64         `json:"agent_config_id"`
	InputData         json.RawMessage `json:"input_data"`
	Status            string          `json:"status"`
	LastRunAt         *time.Time      `json:"last_run_at"`
	NextRunAt         *time.Time      `json:"next_run_at"`
	Config            json.RawMessage `json:"config"`
	NotificationConfig json.RawMessage `json:"notification_config"`
	TimeoutSeconds    int             `json:"timeout_seconds"`
	RetryCount        int             `json:"retry_count"`
	Enabled           bool            `json:"enabled"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type AutopilotExecutionResponse struct {
	ID          uint64          `json:"id"`
	TaskID      uint64          `json:"task_id"`
	Status      string          `json:"status"`
	TriggerType string          `json:"trigger_type"`
	InputData   json.RawMessage `json:"input_data"`
	OutputData  json.RawMessage `json:"output_data"`
	ErrorInfo   string          `json:"error_info"`
	Logs        json.RawMessage `json:"logs"`
	StartedAt   *time.Time      `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	FailedAt    *time.Time      `json:"failed_at"`
	DurationMs  int64           `json:"duration_ms"`
	RetryCount  int             `json:"retry_count"`
	CreatedAt   time.Time       `json:"created_at"`
}

type AgentTemplateResponse struct {
	ID              uint64          `json:"id"`
	Name            string          `json:"name"`
	Description     *string         `json:"description"`
	IsPreset        bool            `json:"is_preset"`
	Icon            string          `json:"icon"`
	SystemPrompt    string          `json:"system_prompt"`
	AvailableSkills json.RawMessage `json:"available_skills"`
	AvailableTools  json.RawMessage `json:"available_tools"`
	DefaultConfig   json.RawMessage `json:"default_config"`
	Version         string          `json:"version"`
	Status          string          `json:"status"`
	WorkspaceID     *uint64         `json:"workspace_id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type AgentConfigResponse struct {
	ID             uint64  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	APIEndpoint    *string `json:"api_endpoint"`
	InferenceLevel string  `json:"inference_level"`
	ServiceLevel   string  `json:"service_level"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
	TopP           float64 `json:"top_p"`
	IsDefault      bool    `json:"is_default"`
	IsActive       bool    `json:"is_active"`
	WorkspaceID    uint64  `json:"workspace_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RuntimeResponse struct {
	ID            uint64          `json:"id"`
	Name          string          `json:"name"`
	RuntimeType   string          `json:"runtime_type"`
	RuntimeMode   string          `json:"runtime_mode"`
	Endpoint      *string         `json:"endpoint"`
	Status        string          `json:"status"`
	Capacity      int             `json:"capacity"`
	CurrentLoad   int             `json:"current_load"`
	Version       string          `json:"version"`
	HostInfo      json.RawMessage `json:"host_info"`
	LastHeartbeat *time.Time      `json:"last_heartbeat"`
	Metadata      json.RawMessage `json:"metadata"`
	WorkspaceID   uint64          `json:"workspace_id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type SkillResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	SkillType   string          `json:"skill_type"`
	Version     string          `json:"version"`
	Status      string          `json:"status"`
	SkillMD     string          `json:"skill_md"`
	Parameters  json.RawMessage `json:"parameters"`
	Tags        json.RawMessage `json:"tags"`
	UsageCount  int             `json:"usage_count"`
	IsShared    bool            `json:"is_shared"`
	WorkspaceID uint64          `json:"workspace_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AgentTaskResponse struct {
	ID               uint64          `json:"id"`
	Title            string          `json:"title"`
	Description      *string         `json:"description"`
	Status           string          `json:"status"`
	Priority         string          `json:"priority"`
	Progress         int             `json:"progress"`
	TaskType         string          `json:"task_type"`
	InputData        json.RawMessage `json:"input_data"`
	OutputData       json.RawMessage `json:"output_data"`
	ErrorInfo        *string         `json:"error_info"`
	Logs             json.RawMessage `json:"logs"`
	AgentTemplateID  *uint64         `json:"agent_template_id"`
	AgentConfigID    *uint64         `json:"agent_config_id"`
	RuntimeID        *uint64         `json:"runtime_id"`
	WorkspaceID      uint64          `json:"workspace_id"`
	ProjectID        *uint64         `json:"project_id"`
	IssueID          *uint64         `json:"issue_id"`
	EnqueuedAt       time.Time       `json:"enqueued_at"`
	ClaimedAt        *time.Time      `json:"claimed_at"`
	StartedAt        *time.Time      `json:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at"`
	CancelledAt      *time.Time      `json:"cancelled_at"`
	EstimatedTime    *int            `json:"estimated_time"`
	ActualTime       *int            `json:"actual_time"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type TaskLogResponse struct {
	ID        uint64          `json:"id"`
	TaskID    uint64          `json:"task_id"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}
