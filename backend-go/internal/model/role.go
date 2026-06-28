package model

// Role represents a custom role for a workspace or project.
// Scope determines if it's workspace-level or project-level.
type Role struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	// Scope: "workspace" or "project"
	Scope       string `gorm:"size:20;default:project" json:"scope"`
	// WorkspaceID is set for workspace-level roles; NULL for project-level roles that belong to a specific workspace's project
	WorkspaceID *uint64 `gorm:"index" json:"workspace_id"`
	// ProjectID is set for project-specific custom roles; NULL for workspace-level roles
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	// IsSystem indicates if this is a system-defined role (cannot be deleted)
	IsSystem    bool    `gorm:"default:false" json:"is_system"`
	// SortOrder for UI ordering
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	// Level: 5=guest, 15=member, 20=admin (for quick comparison without loading permissions)
	Level       int     `gorm:"default:15;index" json:"level"`
	// Permissions loaded via join table
	Permissions []Permission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID" json:"permissions,omitempty"`
}
