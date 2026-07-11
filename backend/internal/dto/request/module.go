package request

type ModuleCreate struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ProjectID   uint64 `json:"project_id" binding:"required"`
	WorkspaceID uint64 `json:"workspace_id,omitempty"`
	ParentID    *uint64 `json:"parent_id"`
}

type ModuleUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *uint64 `json:"parent_id"`
}
