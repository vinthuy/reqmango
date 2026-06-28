package model

import "encoding/json"

// SavedView represents a persisted filter/layout configuration for a project.
type SavedView struct {
	BaseModel

	Name        string  `gorm:"size:100;not null" json:"name"`
	Description *string `gorm:"size:255" json:"description"`

	// View type: "list" | "kanban"
	ViewType string `gorm:"size:20;default:list" json:"view_type"`

	// Filters: JSON blob storing the filter conditions
	Filters json.RawMessage `gorm:"type:jsonb" json:"filters"`

	// Sort configuration: JSON array of {field, direction}
	SortConfig json.RawMessage `gorm:"type:jsonb" json:"sort_config"`

	// Column visibility/order for list view: JSON array of field names
	Columns json.RawMessage `gorm:"type:jsonb" json:"columns"`

	// Group-by field for kanban view: e.g., "state_id" or "priority" or "assignee_id"
	GroupBy *string `gorm:"size:50" json:"group_by"`

	IsDefault bool   `gorm:"default:false" json:"is_default"`
	IsShared  bool   `gorm:"default:false" json:"is_shared"`
	OwnerID   uint64 `gorm:"not null;index" json:"owner_id"`
	ProjectID uint64 `gorm:"not null;index" json:"project_id"`

	// Relationships
	Owner   User    `gorm:"foreignKey:OwnerID" json:"-"`
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (SavedView) TableName() string {
	return "saved_views"
}
