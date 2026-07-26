package request

import "encoding/json"

// Squad DTOs
type SquadCreate struct {
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	LeaderAgentID  *uint64                `json:"leader_agent_id"`
	ProjectID      *uint64                `json:"project_id"`
	Goal           string                 `json:"goal"`
	Config         map[string]interface{} `json:"config"`
	Members        []SquadMemberCreate    `json:"members"`
}

type SquadMemberCreate struct {
	AgentID       uint64 `json:"agent_id"`
	Role          string `json:"role"`
	AgentConfigID uint64 `json:"agent_config_id"`
}

type SquadUpdate struct {
	Name           *string                 `json:"name"`
	Description    *string                 `json:"description"`
	LeaderAgentID  *uint64                `json:"leader_agent_id"`
	Goal           *string                 `json:"goal"`
	Config         map[string]interface{} `json:"config"`
}

type SquadMemberAdd struct {
	AgentID       uint64 `json:"agent_id"`
	Role          string `json:"role"`
	AgentConfigID uint64 `json:"agent_config_id"`
}

type SquadExecutionStart struct {
	Goal      string                 `json:"goal" binding:"required"`
	InputData map[string]interface{} `json:"input_data"`
}

// Autopilot DTOs
type AutopilotTaskCreate struct {
	Name               string                 `json:"name" binding:"required"`
	Description        string                 `json:"description"`
	TriggerType        string                 `json:"trigger_type" binding:"required"`
	CronExpression     string                 `json:"cron_expression"`
	TaskType           string                 `json:"task_type" binding:"required"`
	AgentTemplateID    *uint64                `json:"agent_template_id"`
	AgentConfigID      *uint64                `json:"agent_config_id"`
	ProjectID          *uint64                `json:"project_id"`
	InputData          map[string]interface{} `json:"input_data"`
	Config             map[string]interface{} `json:"config"`
	NotificationConfig map[string]interface{} `json:"notification_config"`
	TimeoutSeconds     int                    `json:"timeout_seconds"`
	RetryCount         int                    `json:"retry_count"`
	Enabled            bool                   `json:"enabled"`
}

type AutopilotTaskUpdate struct {
	Name            *string                 `json:"name"`
	Description     *string                 `json:"description"`
	CronExpression  *string                 `json:"cron_expression"`
	InputData       map[string]interface{} `json:"input_data"`
	Config          map[string]interface{} `json:"config"`
	Enabled         *bool                   `json:"enabled"`
}

type AgentTemplateCreate struct {
	Name            string          `json:"name" binding:"required"`
	Description     *string         `json:"description"`
	Icon            string          `json:"icon"`
	SystemPrompt    string          `json:"system_prompt" binding:"required"`
	AvailableSkills json.RawMessage `json:"available_skills"`
	AvailableTools  json.RawMessage `json:"available_tools"`
	DefaultConfig   json.RawMessage `json:"default_config"`
	Version         string          `json:"version"`
}

type AgentTemplateUpdate struct {
	Name            *string          `json:"name"`
	Description     *string         `json:"description"`
	Icon            *string          `json:"icon"`
	SystemPrompt    *string          `json:"system_prompt"`
	AvailableSkills *json.RawMessage `json:"available_skills"`
	AvailableTools  *json.RawMessage `json:"available_tools"`
	DefaultConfig   *json.RawMessage `json:"default_config"`
	Version         *string          `json:"version"`
	Status          *string          `json:"status"`
}

type AgentConfigCreate struct {
	Name           string  `json:"name" binding:"required"`
	Description    *string `json:"description"`
	Provider       string  `json:"provider" binding:"required"`
	Model          string  `json:"model" binding:"required"`
	APIKey         string  `json:"api_key" binding:"required"`
	APIEndpoint    *string `json:"api_endpoint"`
	InferenceLevel string  `json:"inference_level"`
	ServiceLevel   string  `json:"service_level"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
	TopP           float64 `json:"top_p"`
	IsDefault      bool    `json:"is_default"`
}

type AgentConfigUpdate struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	Provider       *string `json:"provider"`
	Model          *string `json:"model"`
	APIKey         *string `json:"api_key"`
	APIEndpoint    *string `json:"api_endpoint"`
	InferenceLevel *string `json:"inference_level"`
	ServiceLevel   *string `json:"service_level"`
	MaxTokens      *int    `json:"max_tokens"`
	Temperature    *float64 `json:"temperature"`
	TopP           *float64 `json:"top_p"`
	IsDefault      *bool   `json:"is_default"`
	IsActive       *bool   `json:"is_active"`
}

type RuntimeCreate struct {
	Name        string          `json:"name" binding:"required"`
	RuntimeType string          `json:"runtime_type" binding:"required"`
	RuntimeMode string          `json:"runtime_mode"`
	Endpoint    *string         `json:"endpoint"`
	Capacity    int             `json:"capacity"`
	Metadata    json.RawMessage `json:"metadata"`
}

type RuntimeUpdate struct {
	Name        *string          `json:"name"`
	RuntimeType *string          `json:"runtime_type"`
	RuntimeMode *string          `json:"runtime_mode"`
	Endpoint    *string         `json:"endpoint"`
	Capacity    *int             `json:"capacity"`
	Metadata    *json.RawMessage `json:"metadata"`
}

type RuntimeHeartbeat struct {
	Version     string          `json:"version"`
	HostInfo    json.RawMessage `json:"host_info"`
	CurrentLoad int             `json:"current_load"`
}

type SkillCreate struct {
	Name        string          `json:"name" binding:"required"`
	Description *string         `json:"description"`
	SkillType   string          `json:"skill_type"`
	SkillMD     string          `json:"skill_md" binding:"required"`
	Parameters  json.RawMessage `json:"parameters"`
	Tags        json.RawMessage `json:"tags"`
	IsShared    bool            `json:"is_shared"`
}

type SkillUpdate struct {
	Name        *string          `json:"name"`
	Description *string         `json:"description"`
	SkillType   *string          `json:"skill_type"`
	SkillMD     *string          `json:"skill_md"`
	Parameters  *json.RawMessage `json:"parameters"`
	Tags        *json.RawMessage `json:"tags"`
	IsShared    *bool            `json:"is_shared"`
	Status      *string          `json:"status"`
}

type SkillExecute struct {
	Parameters json.RawMessage `json:"parameters"`
}

type AgentTaskCreate struct {
	Title             string          `json:"title" binding:"required"`
	Description       *string         `json:"description"`
	Priority          string          `json:"priority"`
	TaskType          string          `json:"task_type"`
	InputData         json.RawMessage `json:"input_data"`
	AgentTemplateID   *uint64         `json:"agent_template_id"`
	AgentConfigID     *uint64         `json:"agent_config_id"`
	ProjectID         *uint64         `json:"project_id"`
	IssueID           *uint64         `json:"issue_id"`
	EstimatedTime     *int            `json:"estimated_time"`
}

type AgentTaskUpdate struct {
	Title       *string          `json:"title"`
	Description *string         `json:"description"`
	Priority    *string          `json:"priority"`
	Status      *string          `json:"status"`
	Progress    *int             `json:"progress"`
	OutputData  *json.RawMessage `json:"output_data"`
	ErrorInfo   *string         `json:"error_info"`
}

type AgentTaskClaim struct {
	RuntimeID uint64 `json:"runtime_id"`
}

type AgentTaskStart struct {
}

type AgentTaskComplete struct {
	OutputData json.RawMessage `json:"output_data"`
	ActualTime int             `json:"actual_time"`
}

type AgentTaskFail struct {
	ErrorInfo      string `json:"error_info"`
	FailureReason  string `json:"failure_reason"` // Optional: "agent_error", "timeout", "runtime_offline", "invalid_input", "model_error", "rate_limit", "unknown"
}
