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
	Config         json.RawMessage `json:"config"`
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
	CancelledAt  *time.Time      `json:"cancelled_at,omitempty"`
	CancelReason string          `json:"cancel_reason"`
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

type RuntimeStatsResponse struct {
	Total             int64 `json:"total"`
	Online            int64 `json:"online"`
	Offline           int64 `json:"offline"`
	Busy              int64 `json:"busy"`
	TotalCapacity     int   `json:"total_capacity"`
	TotalCurrentLoad  int   `json:"total_current_load"`
	AvailableCapacity int   `json:"available_capacity"`
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

type SkillStepResponse struct {
	Step     int                 `json:"step"`
	Action   string              `json:"action"`
	Tool     string              `json:"tool,omitempty"`
	Input    map[string]interface{} `json:"input,omitempty"`
	Output   interface{}         `json:"output,omitempty"`
	Error    string              `json:"error,omitempty"`
	Status   string              `json:"status"`
}

type SkillExecutionResponse struct {
	SkillID    uint64              `json:"skill_id"`
	SkillName  string              `json:"skill_name"`
	Steps      []SkillStepResponse `json:"steps"`
	FinalResult string             `json:"final_result"`
	Error      string              `json:"error,omitempty"`
	TokensUsed int                 `json:"tokens_used"`
}

type SkillExecutionLogResponse struct {
	ID          uint64         `json:"id"`
	SkillID     uint64         `json:"skill_id"`
	WorkspaceID uint64         `json:"workspace_id"`
	InputParams json.RawMessage `json:"input_params"`
	OutputResult json.RawMessage `json:"output_result"`
	Status      string         `json:"status"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	TokensUsed  int            `json:"tokens_used"`
	DurationMs  int64          `json:"duration_ms"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type PaginationResponse struct {
	Items       interface{} `json:"items"`
	Total       int64       `json:"total"`
	Page        int         `json:"page"`
	PageSize    int         `json:"page_size"`
	TotalPages  int         `json:"total_pages"`
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
	
	// Attribution fields (MUL-3963)
	FailureReason    string          `json:"failure_reason,omitempty"`   // Why the task failed
	Attribution      json.RawMessage `json:"attribution,omitempty"`      // TaskAttribution JSON
	ParentTaskID     *uint64         `json:"parent_task_id,omitempty"`   // Parent task for subtasks
	RetryOfTaskID    *uint64         `json:"retry_of_task_id,omitempty"` // Task this is a retry of
	RerunOfTaskID    *uint64         `json:"rerun_of_task_id,omitempty"` // Task this is a rerun of
	
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

// Monitoring Dashboard DTOs

// AgentPresenceSummary contains agent presence statistics.
type AgentPresenceSummary struct {
	Total      int64 `json:"total"`
	Online     int64 `json:"online"`
	Unstable   int64 `json:"unstable"`
	Offline    int64 `json:"offline"`
	Archived   int64 `json:"archived"`
	Working    int64 `json:"working"`
	Queued     int64 `json:"queued"`
	Idle       int64 `json:"idle"`
}

// TaskExecutionStats contains task execution statistics.
type TaskExecutionStats struct {
	Total       int64 `json:"total"`
	Completed   int64 `json:"completed"`
	Failed      int64 `json:"failed"`
	Cancelled   int64 `json:"cancelled"`
	Running     int64 `json:"running"`
	Enqueued    int64 `json:"enqueued"`
	SuccessRate float64 `json:"success_rate"`
	AvgDurationMs int64 `json:"avg_duration_ms"`
}

// ToolCallStats contains tool call statistics.
type ToolCallStats struct {
	Total       int64 `json:"total"`
	Success     int64 `json:"success"`
	Failed      int64 `json:"failed"`
	AvgDurationMs int64 `json:"avg_duration_ms"`
	TopTools    []ToolCallFrequency `json:"top_tools"`
}

// ToolCallFrequency contains tool call frequency data.
type ToolCallFrequency struct {
	ToolName   string `json:"tool_name"`
	ToolID     uint64 `json:"tool_id"`
	CallCount  int64  `json:"call_count"`
	SuccessRate float64 `json:"success_rate"`
}

// SkillUsageStats contains skill usage statistics.
type SkillUsageStats struct {
	TotalExecutions int64 `json:"total_executions"`
	ActiveSkills    int64 `json:"active_skills"`
	TopSkills       []SkillUsageFrequency `json:"top_skills"`
}

// SkillUsageFrequency contains skill usage frequency data.
type SkillUsageFrequency struct {
	SkillName    string `json:"skill_name"`
	SkillID      uint64 `json:"skill_id"`
	ExecutionCount int64 `json:"execution_count"`
	AvgDurationMs int64 `json:"avg_duration_ms"`
}

// AgentMonitoringResponse is the main monitoring dashboard response.
type AgentMonitoringResponse struct {
	WorkspaceID        uint64               `json:"workspace_id"`
	AgentPresence      AgentPresenceSummary `json:"agent_presence"`
	TaskExecution      TaskExecutionStats   `json:"task_execution"`
	ToolCalls          ToolCallStats        `json:"tool_calls"`
	SkillUsage         SkillUsageStats      `json:"skill_usage"`
	GeneratedAt        time.Time            `json:"generated_at"`
}
