package model

import (
	"encoding/json"
)

// AgentDecision records the decision-making process of an agent for explainability.
type AgentDecision struct {
	BaseModel

	AgentID       uint64          `gorm:"not null;index" json:"agent_id"`
	IssueID       *uint64         `gorm:"index" json:"issue_id"`
	AgentTaskID   *uint64         `gorm:"index" json:"agent_task_id"`
	WorkflowRunID *uint64         `gorm:"index" json:"workflow_run_id"`
	NodeType      string          `gorm:"size:50" json:"node_type"` // requirement_analysis|design|coding|testing
	Thinking      string          `gorm:"type:text" json:"thinking"` // thinking process
	Decision      string          `gorm:"type:text" json:"decision"` // decision result
	Reasoning     string          `gorm:"type:text" json:"reasoning"` // reasoning basis
	Alternatives  json.RawMessage `gorm:"type:jsonb" json:"alternatives"` // alternative options (JSON)
	Confidence    float64         `json:"confidence"` // confidence level

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID" json:"-"`
}

func (AgentDecision) TableName() string {
	return "agent_decisions"
}
