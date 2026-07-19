package model

import "time"

// Workflow defines a state transition rule set scoped to an issue type.
type Workflow struct {
	BaseModel
	Name         string  `gorm:"type:varchar(100);not null" json:"name"`
	Description  string  `gorm:"type:text" json:"description"`
	ProjectID    *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID  uint64  `gorm:"index" json:"workspace_id"`
	IssueTypeID  *uint64 `gorm:"index" json:"issue_type_id"`
	IsActive     bool    `gorm:"default:true" json:"is_active"`

	Transitions []StateTransition `gorm:"foreignKey:WorkflowID" json:"-"`
	Project     Project           `gorm:"foreignKey:ProjectID" json:"-"`
}

func (Workflow) TableName() string { return "workflows" }

// AutomationRule defines an automated trigger→condition→action rule.
type AutomationRule struct {
	BaseModel
	Name           string `gorm:"type:varchar(100);not null" json:"name"`
	Description    string `gorm:"type:text" json:"description"`
	ProjectID      uint64 `gorm:"index" json:"project_id"`
	WorkspaceID    uint64 `gorm:"index" json:"workspace_id"`
	IsEnabled      bool   `gorm:"default:true" json:"is_enabled"`
	Sequence       int    `gorm:"default:1" json:"sequence"`
	ExecutionCount int    `gorm:"default:0" json:"execution_count"`

	TriggerType string `gorm:"type:varchar(50);not null" json:"trigger_type"`
	Conditions  string `gorm:"type:text" json:"conditions"`
	Actions     string `gorm:"type:text" json:"actions"`

	// Workspace-level rule project scope: "all" = all projects, or JSON array "[1,2,3]"
	Scope string `gorm:"type:varchar(50);default:'all'" json:"scope"`

	// Scheduled trigger configuration (JSON): {"frequency":"daily","time":"09:00","days":["mon","wed","fri"]}
	ScheduleConfig  string     `gorm:"type:text" json:"schedule_config,omitempty"`
	LastTriggeredAt *time.Time `gorm:"" json:"last_triggered_at,omitempty"`

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

}

func (AutomationExecution) TableName() string { return "automation_executions" }
