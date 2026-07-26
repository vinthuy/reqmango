package model

import (
	"encoding/json"
	"time"
)

// AgentPresenceInterface defines the interface for updating agent presence.
// Used to avoid circular dependency between internal/service and internal/ai/service.
type AgentPresenceInterface interface {
	UpdatePresenceOnTaskStateChange(agentID, taskID uint64, taskStatus string) error
	CreateSnapshot(agentID, taskID uint64, taskTitle, taskStatus string, progress int, currentStep *string) error
}

// AgentPermissionMode defines who may trigger/assign/@mention/chat an agent.
type AgentPermissionMode string

const (
	PermissionModePrivate  AgentPermissionMode = "private"   // Only owner can invoke
	PermissionModePublicTo AgentPermissionMode = "public_to" // Invocation targets define who can invoke
)

// AgentVisibility defines the visibility of an agent.
type AgentVisibility string

const (
	VisibilityWorkspace AgentVisibility = "workspace" // Visible to all workspace members
	VisibilityPrivate   AgentVisibility = "private"   // Only visible to owner
)

// AgentAvailability defines the availability status of an agent.
type AgentAvailability string

const (
	AvailabilityOnline   AgentAvailability = "online"    // Agent is available and responsive
	AvailabilityUnstable AgentAvailability = "unstable"  // Agent is experiencing issues
	AvailabilityOffline  AgentAvailability = "offline"   // Agent is not available
	AvailabilityArchived AgentAvailability = "archived"  // Agent has been archived
)

// AgentWorkload defines the workload status of an agent.
type AgentWorkload string

const (
	WorkloadWorking AgentWorkload = "working" // Agent is actively working on tasks
	WorkloadQueued  AgentWorkload = "queued"  // Agent has tasks waiting
	WorkloadIdle    AgentWorkload = "idle"    // Agent has no pending tasks
)

// AgentInvocationTarget represents a single invocation grant on an agent.
type AgentInvocationTarget struct {
	TargetType string `json:"target_type"` // "workspace" | "member" | "team"
	TargetID   string `json:"target_id,omitempty"` // null for workspace target
}

// Agent represents an AI agent that can be assigned to work items and @mentioned in comments.
type Agent struct {
	BaseModel

	WorkspaceID       uint64                `gorm:"not null;index" json:"workspace_id"`
	Name              string                `gorm:"size:128;not null" json:"name"`
	Avatar            string                `gorm:"size:10;default:🤖" json:"avatar"`
	AgentType         string                `gorm:"size:20;default:builtin" json:"agent_type"`       // "builtin" | "custom"
	Capabilities      json.RawMessage       `gorm:"type:jsonb;default:'[]'" json:"capabilities"`     // JSON array of tool names
	Status            string                `gorm:"size:20;default:active" json:"status"`             // "active" | "inactive"
	ModelOverride     *string               `gorm:"size:50" json:"model_override,omitempty"`          // override workspace LLM model
	SystemPrompt      *string               `gorm:"type:text" json:"system_prompt,omitempty"`         // custom system prompt
	
	// Permission fields (MUL-3963)
	PermissionMode    AgentPermissionMode   `gorm:"size:20;default:private" json:"permission_mode"`   // "private" | "public_to"
	InvocationTargets json.RawMessage       `gorm:"type:jsonb;default:'[]'" json:"invocation_targets"` // Array of AgentInvocationTarget
	Visibility        AgentVisibility       `gorm:"size:20;default:private" json:"visibility"`        // "workspace" | "private"
	
	// Presence fields (MUL-3963)
	Availability      AgentAvailability     `gorm:"size:20;default:offline" json:"availability"`      // "online" | "unstable" | "offline" | "archived"
	Workload          AgentWorkload         `gorm:"size:20;default:idle" json:"workload"`              // "working" | "queued" | "idle"
	LastActiveAt      *time.Time            `json:"last_active_at,omitempty"`                        // Last activity timestamp
	RunningTaskID     *uint64               `gorm:"index" json:"running_task_id,omitempty"`          // Currently running task
	QueuedTaskCount   int                   `gorm:"default:0" json:"queued_task_count"`             // Number of queued tasks

	// Relationships
	Activities []AgentActivity `gorm:"foreignKey:AgentID" json:"-"`
	Workspace  Workspace       `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Agent) TableName() string { return "agents" }

// AgentActivity records every action an AI agent performs for audit trail purposes.
type AgentActivity struct {
	BaseModel

	AgentID       uint64    `gorm:"not null;index" json:"agent_id"`
	IssueID       *uint64   `gorm:"index" json:"issue_id"`            // optional — may be nil for workspace-level actions
	Action        string    `gorm:"size:50;not null" json:"action"`   // "dispatch" | "auto_triage" | "auto_assign" | "mention" | "summarize" | "custom"
	ResultSummary string    `gorm:"type:text" json:"result_summary"` // human-readable summary of what happened
	Rating        *int      `gorm:"default:null" json:"rating,omitempty"` // 1=positive, -1=negative, null=no feedback
	ExecutedAt    time.Time `gorm:"autoCreateTime" json:"executed_at"`
	AgentName     string    `gorm:"size:128" json:"agent_name"`      // denormalized for audit readability
	TaskContext   *string   `gorm:"type:text" json:"task_context,omitempty"` // the prompt/task sent to the LLM

	// Relationships
	Agent Agent  `gorm:"foreignKey:AgentID" json:"-"`
	Issue *Issue `gorm:"foreignKey:IssueID" json:"-"`
}

func (AgentActivity) TableName() string { return "agent_activities" }
