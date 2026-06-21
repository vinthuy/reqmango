package model

import "time"

// Cycle represents an iteration/sprint cycle.
// Minimal model for Phase 1 - full Cycle CRUD comes in Phase 2.
type Cycle struct {
	BaseModel

	Name        string     `gorm:"size:255;not null" json:"name"`
	Description *string    `gorm:"size:1000" json:"description"`
	StartDate   time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate     *time.Time `gorm:"type:date" json:"end_date"`
	CompletedAt *time.Time `json:"completed_at"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null" json:"workspace_id"`

	// Relationships
	Project    Project      `gorm:"foreignKey:ProjectID" json:"-"`
	IssueLinks []IssueCycle `gorm:"foreignKey:CycleID" json:"-"`
}

func (Cycle) TableName() string {
	return "cycles"
}
