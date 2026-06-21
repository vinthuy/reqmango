package model

// State represents an issue state (e.g., Backlog, Todo, In Progress, Done).
type State struct {
	BaseModel

	Name        string `gorm:"size:255;not null" json:"name"`
	Color       string `gorm:"size:50;default:#6B7280" json:"color"`
	Group       string `gorm:"size:50;default:backlog" json:"group"`
	Sequence    int    `gorm:"default:1" json:"sequence"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	ProjectID   uint64 `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64 `gorm:"not null" json:"workspace_id"`

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
// Minimal model for Phase 1 - full workflow comes in Phase 2.
type StateTransition struct {
	BaseModel

	Name          string  `gorm:"size:255;not null" json:"name"`
	Description   *string `gorm:"size:500" json:"description"`
	SourceStateID uint64  `gorm:"not null" json:"source_state_id"`
	TargetStateID uint64  `gorm:"not null" json:"target_state_id"`
	IssueTypeID   *uint64 `json:"issue_type_id"`
	IsAuto        bool    `gorm:"default:false" json:"is_auto"`
	ProjectID     uint64  `gorm:"not null" json:"project_id"`
	WorkspaceID   uint64  `gorm:"not null" json:"workspace_id"`

	SourceState State `gorm:"foreignKey:SourceStateID" json:"-"`
	TargetState State `gorm:"foreignKey:TargetStateID" json:"-"`
}

func (StateTransition) TableName() string {
	return "state_transitions"
}
