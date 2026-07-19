package model

// IssueTypeImport records a project's reference to a workspace-level issue type.
//
// Plane v3-style "Import" model: a project imports workspace-level types
// by reference (link) rather than making independent copies. The project holds
// a reference to the workspace type, so any workspace admin update is
// automatically reflected in every project that imported it.
//
// Legacy auto-inherit (where all workspace-level types were automatically
// visible to every project) has been removed. Now only explicitly imported
// workspace types appear in a project's type list.
//
// Custom fields attached to an imported type are visible in the project
// automatically (they "follow" the type), without requiring separate enrollment.
type IssueTypeImport struct {
	BaseModel
	ProjectID       uint64 `gorm:"not null;uniqueIndex:idx_project_workspace_type" json:"project_id"`
	WorkspaceTypeID uint64 `gorm:"not null;uniqueIndex:idx_project_workspace_type" json:"workspace_type_id"`
	WorkspaceID     uint64 `gorm:"not null;index" json:"workspace_id"`

	// Relationships
	Project       Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	WorkspaceType IssueType `gorm:"foreignKey:WorkspaceTypeID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueTypeImport) TableName() string {
	return "issue_type_imports"
}
