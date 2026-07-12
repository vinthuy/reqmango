package model

// State represents an issue state (e.g., Backlog, Todo, In Progress, Done).
type State struct {
	BaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Color       string  `gorm:"size:50;default:#6B7280" json:"color"`
	Group       string  `gorm:"size:50;default:backlog" json:"group"`
	Sequence    int     `gorm:"default:1" json:"sequence"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null" json:"workspace_id"`

	// Relationships
	Project            Project           `gorm:"foreignKey:ProjectID" json:"-"`
	Issues             []Issue           `gorm:"foreignKey:StateID" json:"-"`
	SourceTransitions  []StateTransition `gorm:"foreignKey:SourceStateID" json:"-"`
	TargetTransitions  []StateTransition `gorm:"foreignKey:TargetStateID" json:"-"`
}

func (State) TableName() string {
	return "states"
}

// StateTransition represents a workflow transition between states.
type StateTransition struct {
	BaseModel

	Name          string  `gorm:"size:255;not null" json:"name"`
	Description   *string `gorm:"size:500" json:"description"`
	WorkflowID    uint64  `gorm:"not null;index" json:"workflow_id"`
	SourceStateID uint64  `gorm:"not null" json:"source_state_id"`
	TargetStateID uint64  `gorm:"not null" json:"target_state_id"`
	IssueTypeID   *uint64 `json:"issue_type_id"`
	IsAuto        bool    `gorm:"default:false" json:"is_auto"`
	RuleType      string  `gorm:"default:'allow'" json:"rule_type"`
	ApproverIDs   *string `gorm:"type:text" json:"approver_ids"`
	RoleAllowed   string  `gorm:"size:50" json:"role_allowed"`
	ProjectID     *uint64 `json:"project_id"`
	WorkspaceID   uint64  `gorm:"not null" json:"workspace_id"`

	Workflow    Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	SourceState State    `gorm:"foreignKey:SourceStateID" json:"-"`
	TargetState State    `gorm:"foreignKey:TargetStateID" json:"-"`
}

func (StateTransition) TableName() string {
	return "state_transitions"
}
