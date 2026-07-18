package model

import (
	"encoding/json"
	"time"
)

// Agent represents an AI agent that can be assigned to work items and @mentioned in comments.
type Agent struct {
	ID           uint64          `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `gorm:"index" json:"-"`
	CreatedByID  *uint64         `json:"created_by_id"`
	UpdatedByID  *uint64         `json:"updated_by_id"`

	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	Name          string          `gorm:"size:128;not null" json:"name"`
	Avatar        string          `gorm:"size:10;default:🤖" json:"avatar"`
	AgentType     string          `gorm:"size:20;default:builtin" json:"agent_type"`
	Capabilities  json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"capabilities"`
	Status        string          `gorm:"size:20;default:active" json:"status"`
	ModelOverride *string         `gorm:"size:50" json:"model_override,omitempty"`
	SystemPrompt  *string         `gorm:"type:text" json:"system_prompt,omitempty"`
}

func (Agent) TableName() string { return "agents" }

// AgentActivity records every action an AI agent performs for audit trail purposes.
type AgentActivity struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
	CreatedByID  *uint64    `json:"created_by_id"`
	UpdatedByID  *uint64    `json:"updated_by_id"`

	AgentID       uint64    `gorm:"not null;index" json:"agent_id"`
	IssueID       *uint64   `gorm:"index" json:"issue_id"`
	Action        string    `gorm:"size:50;not null" json:"action"`
	ResultSummary string    `gorm:"type:text" json:"result_summary"`
	Rating        *int      `gorm:"default:null" json:"rating,omitempty"`
	ExecutedAt    time.Time `gorm:"autoCreateTime" json:"executed_at"`
	AgentName     string    `gorm:"size:128" json:"agent_name"`
	TaskContext   *string   `gorm:"type:text" json:"task_context,omitempty"`
}

func (AgentActivity) TableName() string { return "agent_activities" }
