package model

import "time"

// Project represents a project within a workspace.
type Project struct {
	BaseModel

	Name              string     `gorm:"size:255;not null" json:"name"`
	Identifier        string     `gorm:"size:10;not null" json:"identifier"`
	Description       *string    `gorm:"size:1000" json:"description"`
	IsPublic          bool       `gorm:"default:false" json:"is_public"`
	Timezone          string     `gorm:"size:255;default:UTC" json:"timezone"`
	ArchivedAt        *time.Time `json:"archived_at"`
	WorkspaceID       uint64     `gorm:"not null;index" json:"workspace_id"`
	DefaultAssigneeID *uint64    `json:"default_assignee_id"`
	Color             string     `gorm:"size:20;default:'#6366F1'" json:"color"`
	TemplateID        *uint64    `json:"template_id"`

	// Relationships
	Workspace       Workspace       `gorm:"foreignKey:WorkspaceID" json:"-"`
	DefaultAssignee *User           `gorm:"foreignKey:DefaultAssigneeID" json:"-"`
	Members         []ProjectMember `gorm:"foreignKey:ProjectID" json:"-"`
	Issues          []Issue         `gorm:"foreignKey:ProjectID" json:"-"`
	States          []State         `gorm:"foreignKey:ProjectID" json:"-"`
	Labels          []Label         `gorm:"foreignKey:ProjectID" json:"-"`
	Cycles          []Cycle         `gorm:"foreignKey:ProjectID" json:"-"`
}

func (Project) TableName() string {
	return "projects"
}

// ProjectMember represents a user's membership in a project.
type ProjectMember struct {
	BaseModel

	ProjectID uint64 `gorm:"not null;uniqueIndex:idx_proj_member_user" json:"project_id"`
	UserID    uint64 `gorm:"not null;uniqueIndex:idx_proj_member_user" json:"user_id"`
	Role      int    `gorm:"default:15" json:"role"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    User    `gorm:"foreignKey:UserID" json:"-"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
