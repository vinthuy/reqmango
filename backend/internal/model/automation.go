package model

import "time"

// AutomationRule represents an automation rule for a project.
type AutomationRule struct {
	BaseModel

	Name            string     `gorm:"size:100;not null" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	ProjectID       uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID     uint64     `gorm:"not null;index" json:"workspace_id"`
	IsEnabled       bool       `gorm:"default:true" json:"is_enabled"`
	Sequence        int        `gorm:"default:1" json:"sequence"`
	ExecutionCount  int        `gorm:"default:0" json:"execution_count"`
	TriggerType     string     `gorm:"size:50;not null" json:"trigger_type"`
	Conditions      string     `gorm:"type:text" json:"conditions"`
	Actions         string     `gorm:"type:text" json:"actions"`
	Scope           string     `gorm:"size:20;default:'all'" json:"scope"` // all, project, specific
	ScheduleConfig  string     `gorm:"type:text" json:"schedule_config"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedByID     *uint64    `json:"created_by_id"`
	UpdatedByID     *uint64    `json:"updated_by_id"`
}

func (AutomationRule) TableName() string {
	return "automation_rules"
}

// AutomationRuleOverride represents a per-project override of a workspace automation rule.
type AutomationRuleOverride struct {
	BaseModel

	RuleID     uint64  `gorm:"not null;index" json:"rule_id"`
	ProjectID  uint64  `gorm:"not null;index" json:"project_id"`
	IsEnabled  *bool   `json:"is_enabled"`
	CreatedByID *uint64 `json:"created_by_id"`
	UpdatedByID *uint64 `json:"updated_by_id"`
}

func (AutomationRuleOverride) TableName() string {
	return "automation_rule_overrides"
}

// AutomationExecution represents the execution of an automation rule.
type AutomationExecution struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	RuleID       uint64    `gorm:"not null;index" json:"rule_id"`
	IssueID      uint64    `gorm:"not null;index" json:"issue_id"`
	TriggerType  string    `gorm:"size:50;not null" json:"trigger_type"`
	ContextJSON  string    `gorm:"type:jsonb" json:"context_json"`
	ActionsTaken string    `gorm:"type:jsonb" json:"actions_taken"`
	Status       string    `gorm:"size:20;not null" json:"status"`
	Error        string    `gorm:"type:text" json:"error"`
	Duration     int64     `gorm:"default:0" json:"duration"`
	ExecutedAt   time.Time `gorm:"not null" json:"executed_at"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

func (AutomationExecution) TableName() string {
	return "automation_executions"
}
