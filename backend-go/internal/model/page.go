package model

import "time"

// Page represents a documentation page within a project.
type Page struct {
	BaseModel

	Title       string  `gorm:"size:255;not null" json:"title"`
	Content     string  `gorm:"type:text;not null;default:''" json:"content"`
	ContentJSON *string `gorm:"type:jsonb" json:"content_json"` // TipTap JSON format

	Published  bool       `gorm:"default:true" json:"published"`
	ArchivedAt *time.Time `json:"archived_at"`
	Sequence   int        `gorm:"default:1" json:"sequence"`

	// Hierarchy
	ParentID *uint64 `gorm:"index" json:"parent_id"`
	Depth    int     `gorm:"default:0" json:"depth"` // max 5

	ProjectID   uint64 `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64 `gorm:"not null" json:"workspace_id"`

	// Relationships
	Project  Project `gorm:"foreignKey:ProjectID" json:"-"`
	Parent   *Page   `gorm:"foreignKey:ParentID" json:"-"`
	Children []Page  `gorm:"foreignKey:ParentID" json:"-"`
}

func (Page) TableName() string {
	return "pages"
}
