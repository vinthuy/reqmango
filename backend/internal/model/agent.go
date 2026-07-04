package model

import (
	"encoding/json"
	"time"
)

// Agent represents an AI agent that can be assigned to work items and @mentioned in comments.
type Agent struct {
	BaseModel

	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	Name          string          `gorm:"size:128;not null" json:"name"`
	Avatar        string          `gorm:"size:10;default:🤖" json:"avatar"`
	AgentType     string          `gorm:"size:20;default:builtin" json:"agent_type"`       // "builtin" | "custom"
	Capabilities  json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"capabilities"`     // JSON array of tool names
	Status        string          `gorm:"size:20;default:active" json:"status"`             // "active" | "inactive"
	ModelOverride *string         `gorm:"size:50" json:"model_override,omitempty"`          // override workspace LLM model
	SystemPrompt  *string         `gorm:"type:text" json:"system_prompt,omitempty"`         // custom system prompt

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
