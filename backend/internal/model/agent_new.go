package model

import (
	"encoding/json"
	"time"
)

// AgentTemplate represents a preset agent role template.
// Preset templates include: Requirements Analyst, Developer, Tester, etc.
type AgentTemplate struct {
	BaseModel

	Name           string          `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description    *string         `gorm:"type:text" json:"description,omitempty"`
	IsPreset       bool            `gorm:"default:false" json:"is_preset"`
	Icon           string          `gorm:"size:10;default:🤖" json:"icon"`
	SystemPrompt   string          `gorm:"type:text;not null" json:"system_prompt"`
	AvailableSkills json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"available_skills"`  // Array of skill IDs
	AvailableTools json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"available_tools"`    // Array of tool IDs
	DefaultConfig  json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"default_config"`    // Default model config
	Version        string          `gorm:"size:50;default:1.0" json:"version"`
	Status         string          `gorm:"size:20;default:active" json:"status"`              // "active" | "deprecated"
	WorkspaceID    *uint64         `gorm:"index" json:"workspace_id,omitempty"`               // Null for preset templates

	// Relationships
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (AgentTemplate) TableName() string { return "agent_templates" }

// AgentConfig represents the configuration for an AI agent model.
// Supports multiple providers (Claude Code, Codex, CodeBuddy, etc.)
type AgentConfig struct {
	BaseModel

	Name        string `gorm:"size:128;not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	Provider    string `gorm:"size:30;not null" json:"provider"`    // "deepseek", "anthropic", "openai", "codellama", "codebuddy", etc.
	Model       string `gorm:"size:100;not null" json:"model"`       // Model name
	APIKey      string `gorm:"size:500;not null" json:"-"`           // API key (not returned in JSON)
	APIEndpoint *string `gorm:"size:255" json:"api_endpoint,omitempty"` // Custom endpoint
	InferenceLevel string `gorm:"size:20;default:normal" json:"inference_level"` // "normal", "advanced", "thinking-2"
	ServiceLevel string `gorm:"size:20;default:standard" json:"service_level"`   // "standard", "premium", "turbo"
	MaxTokens   int    `gorm:"default:4096" json:"max_tokens"`
	Temperature float64 `gorm:"type:decimal(3,2);default:0.7" json:"temperature"`
	TopP        float64 `gorm:"type:decimal(3,2);default:1.0" json:"top_p"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (AgentConfig) TableName() string { return "agent_configs" }

// RuntimeHealth defines the health status of a runtime.
type RuntimeHealth string

const (
	RuntimeHealthOnline        RuntimeHealth = "online"         // Runtime is online and responsive
	RuntimeHealthRecentlyLost  RuntimeHealth = "recently_lost"  // Runtime recently lost connection
	RuntimeHealthOffline       RuntimeHealth = "offline"        // Runtime is offline
	RuntimeHealthAboutToGC     RuntimeHealth = "about_to_gc"    // Runtime is about to be garbage collected
)

// Runtime represents a local daemon or cloud runtime for executing agent tasks.
type Runtime struct {
	BaseModel

	Name           string          `gorm:"size:128;not null" json:"name"`
	RuntimeType    string          `gorm:"size:20;not null" json:"runtime_type"`       // "local_daemon", "cloud"
	RuntimeMode    string          `gorm:"size:20;default:pull" json:"runtime_mode"`   // "pull", "push"
	Endpoint       *string         `gorm:"size:255" json:"endpoint,omitempty"`         // WebSocket/HTTP endpoint
	Status         string          `gorm:"size:20;default:offline" json:"status"`      // "online", "offline", "busy"
	Health         RuntimeHealth   `gorm:"size:20;default:offline" json:"health"`      // "online" | "recently_lost" | "offline" | "about_to_gc"
	Capacity       int             `gorm:"default:1" json:"capacity"`                 // Number of concurrent tasks
	CurrentLoad    int             `gorm:"default:0" json:"current_load"`              // Currently running tasks
	Version        string          `gorm:"size:50" json:"version,omitempty"`           // Agent CLI version
	HostInfo       json.RawMessage `gorm:"type:jsonb" json:"host_info,omitempty"`      // CPU, memory, OS info
	LastHeartbeat  *time.Time      `json:"last_heartbeat,omitempty"`                   // Last heartbeat timestamp
	Metadata       json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`    // Additional metadata
	WorkspaceID    uint64          `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

// AgentTaskSnapshot represents a snapshot of an agent's current task state.
type AgentTaskSnapshot struct {
	BaseModel

	AgentID        uint64          `gorm:"not null;index" json:"agent_id"`
	TaskID         uint64          `gorm:"not null;index" json:"task_id"`
	TaskTitle      string          `gorm:"size:255" json:"task_title"`
	TaskStatus     string          `gorm:"size:20" json:"task_status"`
	Progress       int             `gorm:"default:0" json:"progress"`
	CurrentStep    *string         `json:"current_step,omitempty"`    // Current execution step
	StepProgress   *int            `json:"step_progress,omitempty"`   // Progress within current step
	EstimatedTime  *int            `json:"estimated_time,omitempty"`  // Estimated remaining minutes
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID" json:"-"`
	Task  AgentTask `gorm:"foreignKey:TaskID" json:"-"`
}

func (AgentTaskSnapshot) TableName() string { return "agent_task_snapshots" }

func (Runtime) TableName() string { return "runtimes" }

// Skill represents a reusable AI operation guide in SKILL.md format.
type Skill struct {
	BaseModel

	Name        string          `gorm:"size:128;not null" json:"name"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	SkillType   string          `gorm:"size:30;default:custom" json:"skill_type"`      // "builtin", "custom", "shared"
	Version     string          `gorm:"size:50;default:1.0" json:"version"`
	Status      string          `gorm:"size:20;default:active" json:"status"`          // "active", "deprecated"
	SkillMD     string          `gorm:"type:text;not null" json:"skill_md"`           // SKILL.md content
	Parameters  json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"parameters"`    // Parameter definitions
	Tags        json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"tags"`          // Tags for categorization
	UsageCount  int             `gorm:"default:0" json:"usage_count"`                 // How many times this skill was used
	IsShared    bool            `gorm:"default:false" json:"is_shared"`               // Whether shared to workspace
	IsPreset    bool            `gorm:"default:false" json:"is_preset"`               // Whether this is a preset skill
	WorkspaceID uint64          `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Skill) TableName() string { return "skills" }

// SkillExecutionLog represents a log entry for a skill execution.
type SkillExecutionLog struct {
	BaseModel

	SkillID     uint64         `gorm:"not null;index" json:"skill_id"`
	WorkspaceID uint64         `gorm:"not null;index" json:"workspace_id"`
	InputParams json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"input_params"`
	OutputResult json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"output_result"`
	Status      string         `gorm:"size:20;default:running" json:"status"` // "running", "completed", "failed"
	ErrorMessage *string        `gorm:"type:text" json:"error_message,omitempty"`
	TokensUsed  int            `gorm:"default:0" json:"tokens_used"`
	DurationMs  int64          `gorm:"default:0" json:"duration_ms"`

	// Relationships
	Skill   Skill   `gorm:"foreignKey:SkillID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (SkillExecutionLog) TableName() string { return "skill_execution_logs" }

// TaskFailureReason defines the reason why a task failed.
type TaskFailureReason string

const (
	FailureReasonAgentError      TaskFailureReason = "agent_error"      // Agent execution error
	FailureReasonTimeout         TaskFailureReason = "timeout"          // Task timed out
	FailureReasonRuntimeOffline  TaskFailureReason = "runtime_offline"  // Runtime is offline
	FailureReasonInvalidInput    TaskFailureReason = "invalid_input"    // Invalid input parameters
	FailureReasonModelError      TaskFailureReason = "model_error"      // LLM model error
	FailureReasonRateLimit       TaskFailureReason = "rate_limit"       // API rate limit exceeded
	FailureReasonUnknown         TaskFailureReason = "unknown"          // Unknown failure reason
)

// TaskEvidence represents evidence related to a task.
type TaskEvidence struct {
	Kind  string `json:"kind"`   // "issue", "comment", "file", "url", "message"
	RefID string `json:"ref_id"` // Reference ID for the evidence
}

// TaskAttribution represents the attribution information for a task.
type TaskAttribution struct {
	InitiatorID   uint64 `json:"initiator_id"`    // User who initiated the task
	OriginatorID  uint64 `json:"originator_id"`   // Agent/entity that originated the task
	Evidence      []TaskEvidence `json:"evidence,omitempty"` // Evidence supporting the task
}

// AgentTask represents a task assigned to an AI agent for execution.
// Status flow: enqueue -> claim -> start -> complete/fail
type AgentTask struct {
	BaseModel

	Title       string          `gorm:"size:255;not null" json:"title"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	Status      string          `gorm:"size:20;default:enqueue" json:"status"`          // "enqueue", "claimed", "running", "completed", "failed", "cancelled"
	Priority    string          `gorm:"size:20;default:normal" json:"priority"`         // "low", "normal", "high", "urgent"
	Progress    int             `gorm:"default:0" json:"progress"`                     // 0-100
	TaskType    string          `gorm:"size:50" json:"task_type,omitempty"`            // "generate_code", "analyze_requirement", "write_test", etc.
	InputData   json.RawMessage `gorm:"type:jsonb" json:"input_data,omitempty"`        // Input parameters
	OutputData  json.RawMessage `gorm:"type:jsonb" json:"output_data,omitempty"`       // Execution result
	ErrorInfo   *string         `gorm:"type:text" json:"error_info,omitempty"`          // Error message if failed
	Logs        json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"logs"`            // Execution logs

	// Attribution fields (MUL-3963)
	FailureReason TaskFailureReason `gorm:"size:30" json:"failure_reason,omitempty"`   // Why the task failed
	Attribution   json.RawMessage   `gorm:"type:jsonb" json:"attribution,omitempty"`  // TaskAttribution JSON
	ParentTaskID  *uint64           `gorm:"index" json:"parent_task_id,omitempty"`     // Parent task for subtasks
	RetryOfTaskID *uint64           `gorm:"index" json:"retry_of_task_id,omitempty"`  // Task this is a retry of
	RerunOfTaskID *uint64           `gorm:"index" json:"rerun_of_task_id,omitempty"`  // Task this is a rerun of

	// Agent configuration
	AgentTemplateID *uint64 `gorm:"index" json:"agent_template_id,omitempty"`           // Template used
	AgentConfigID   *uint64 `gorm:"index" json:"agent_config_id,omitempty"`             // Model config used
	RuntimeID       *uint64 `gorm:"index" json:"runtime_id,omitempty"`                   // Runtime executing the task

	// Relationships
	WorkspaceID    uint64          `gorm:"not null;index" json:"workspace_id"`
	ProjectID      *uint64         `gorm:"index" json:"project_id,omitempty"`
	IssueID        *uint64         `gorm:"index" json:"issue_id,omitempty"`
	EnqueuedAt     time.Time       `gorm:"autoCreateTime" json:"enqueued_at"`
	ClaimedAt      *time.Time      `json:"claimed_at,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CancelledAt    *time.Time      `json:"cancelled_at,omitempty"`
	EstimatedTime  *int            `json:"estimated_time,omitempty"`                   // Estimated minutes
	ActualTime     *int            `json:"actual_time,omitempty"`                      // Actual minutes

	// Relationships
	Workspace      Workspace      `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project        *Project       `gorm:"foreignKey:ProjectID" json:"-"`
	Issue          *Issue         `gorm:"foreignKey:IssueID" json:"-"`
	AgentTemplate  *AgentTemplate `gorm:"foreignKey:AgentTemplateID" json:"-"`
	AgentConfig    *AgentConfig   `gorm:"foreignKey:AgentConfigID" json:"-"`
	Runtime        *Runtime       `gorm:"foreignKey:RuntimeID" json:"-"`
}

func (AgentTask) TableName() string { return "agent_tasks" }

// TaskLog represents a single log entry for an agent task execution.
type TaskLog struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	TaskID    uint64         `gorm:"not null;index" json:"task_id"`
	Level     string         `gorm:"size:20;not null" json:"level"`      // "debug", "info", "warn", "error"
	Message   string         `gorm:"type:text;not null" json:"message"`  // Log message
	Metadata  json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional context

	// Relationships
	Task AgentTask `gorm:"foreignKey:TaskID" json:"-"`
}

func (TaskLog) TableName() string { return "task_logs" }
