package model

import (
	"encoding/json"
	"time"
)

// AgentSession provides unified observability across all agent activities.
type AgentSession struct {
	ID            string          `gorm:"primaryKey;size:64" json:"id"`
	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	AgentType     string          `gorm:"size:50;not null" json:"agent_type"`
	AgentRef      *string         `gorm:"size:255" json:"agent_ref,omitempty"`
	Status        string          `gorm:"size:50;default:running" json:"status"`
	ModelUsed     *string         `gorm:"size:100" json:"model_used,omitempty"`
	InputSummary  *string         `gorm:"type:text" json:"input_summary,omitempty"`
	OutputSummary *string         `gorm:"type:text" json:"output_summary,omitempty"`
	TokensInput   int             `gorm:"default:0" json:"tokens_input"`
	TokensOutput  int             `gorm:"default:0" json:"tokens_output"`
	CostUSD       float64         `gorm:"type:decimal(10,4);default:0" json:"cost_usd"`
	ToolsCalled   json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"tools_called"`
	ErrorMessage  *string         `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt     time.Time       `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Metadata      json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"metadata"`
}

func (AgentSession) TableName() string { return "agent_sessions" }
