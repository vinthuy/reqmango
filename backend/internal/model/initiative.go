package model

import "time"

type Initiative struct {
	BaseModel
	WorkspaceID uint64     `gorm:"index" json:"workspace_id"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	Color       string     `gorm:"size:20" json:"color,omitempty"`
	Status      string     `gorm:"size:20;default:'active'" json:"status"` // active, completed, paused, at_risk, off_track
	TargetDate  *time.Time `json:"target_date,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedByID uint64     `json:"created_by_id"`

	Projects []Project `gorm:"many2many:initiative_projects;" json:"projects,omitempty"`
}

type InitiativeProject struct {
	InitiativeID uint64 `gorm:"primaryKey" json:"initiative_id"`
	ProjectID    uint64 `gorm:"primaryKey" json:"project_id"`
}
