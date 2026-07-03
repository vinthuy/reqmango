package model

import "time"

// Workflow defines a state transition rule set scoped to an issue type.
type Workflow struct {
	BaseModel
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	ProjectID   uint64  `gorm:"not null;index" json:"project_id"`
	IssueTypeID *uint64 `gorm:"index" json:"issue_type_id"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`

	Transitions []StateTransition `gorm:"foreignKey:WorkflowID" json:"-"`
	Project     Project           `gorm:"foreignKey:ProjectID" json:"-"`
}

func (Workflow) TableName() string { return "workflows" }

// AutomationRule defines an automated trigger→condition→action rule.
type AutomationRule struct {
	BaseModel
	Name           string `gorm:"type:varchar(100);not null" json:"name"`
	Description    string `gorm:"type:text" json:"description"`
	ProjectID      uint64 `gorm:"not null;index" json:"project_id"`
	IsEnabled      bool   `gorm:"default:true" json:"is_enabled"`
	Sequence       int    `gorm:"default:1" json:"sequence"`
	ExecutionCount int    `gorm:"default:0" json:"execution_count"`

	// Trigger: issue_created, issue_updated, state_changed, assignee_changed, comment_added, scheduled
	TriggerType string `gorm:"type:varchar(50);not null" json:"trigger_type"`
	// Conditions: JSON array of {field, operator, value}
	Conditions string `gorm:"type:text" json:"conditions"`
	// Actions: JSON array of {type, field, value}
	Actions string `gorm:"type:text" json:"actions"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (AutomationRule) TableName() string { return "automation_rules" }

// AutomationExecution records the execution history of automation rules.
type AutomationExecution struct {
	BaseModel
	RuleID       uint64    `gorm:"index" json:"rule_id"`
	IssueID      uint64    `gorm:"index" json:"issue_id"`
	TriggerType  string    `gorm:"type:varchar(50)" json:"trigger_type"`
	ContextJSON  string    `gorm:"type:jsonb" json:"context_json"`
	ActionsTaken string    `gorm:"type:jsonb" json:"actions_taken"`
	Status       string    `gorm:"type:varchar(20)" json:"status"`
	Error        string    `gorm:"type:text" json:"error,omitempty"`
	Duration     int64     `json:"duration"`
	ExecutedAt   time.Time `gorm:"index" json:"executed_at"`

	// Relationships
	Rule  AutomationRule `gorm:"foreignKey:RuleID" json:"-"`
	Issue Issue           `gorm:"foreignKey:IssueID" json:"-"`
}

func (AutomationExecution) TableName() string { return "automation_executions" }
