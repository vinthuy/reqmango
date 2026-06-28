package request

// CreateRoleRequest for creating a custom role.
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Scope       string   `json:"scope" binding:"required,oneof=workspace project"`
	WorkspaceID *uint64  `json:"workspace_id"`
	ProjectID   *uint64  `json:"project_id"`
	Level       int      `json:"level"`
	Permissions []uint64 `json:"permissions"` // permission IDs
}

// UpdateRoleRequest for updating a role.
type UpdateRoleRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Level       *int     `json:"level"`
	Permissions []uint64 `json:"permissions"` // replaces all permissions
}
