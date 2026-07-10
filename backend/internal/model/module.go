package model

import "time"

// Module represents a project module (supports hierarchical grouping via ParentID).
type Module struct {
	BaseModel
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null;index" json:"workspace_id"`
	ParentID    *uint64    `gorm:"index" json:"parent_id"`
	Order       int        `gorm:"default:0" json:"order"`
	ArchivedAt  *time.Time `json:"archived_at"`
	IsArchived  bool       `gorm:"default:false" json:"is_archived"`

	// Relationships
	Project    Project        `gorm:"foreignKey:ProjectID" json:"-"`
	IssueLinks []ModuleIssue  `gorm:"foreignKey:ModuleID" json:"-"`
}

func (Module) TableName() string {
	return "modules"
}

// ModuleIssue is a join table for many-to-many module-issue association.
type ModuleIssue struct {
	ModuleID uint64 `gorm:"primaryKey;autoIncrement:false" json:"module_id"`
	IssueID  uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`

	Module Module `gorm:"foreignKey:ModuleID;constraint:OnDelete:CASCADE" json:"-"`
	Issue  Issue  `gorm:"foreignKey:IssueID" json:"-"`
}

func (ModuleIssue) TableName() string {
	return "module_issues"
}
