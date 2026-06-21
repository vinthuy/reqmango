package model

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
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ProjectID   uint64 `gorm:"not null;index" json:"project_id"`
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`
	Sequence    int    `gorm:"default:1" json:"sequence"`

	// Trigger: issue_created, issue_updated, state_changed, assignee_changed, comment_added, scheduled
	TriggerType string `gorm:"type:varchar(50);not null" json:"trigger_type"`
	// Conditions: JSON array of {field, operator, value}
	Conditions string `gorm:"type:text" json:"conditions"`
	// Actions: JSON array of {type, field, value}
	Actions string `gorm:"type:text" json:"actions"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (AutomationRule) TableName() string { return "automation_rules" }
