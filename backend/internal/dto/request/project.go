package request

// ProjectCreateRequest is the request body for creating a project.
type ProjectCreateRequest struct {
	Name              string  `json:"name" binding:"required,min=1,max=255"`
	Identifier        string  `json:"identifier" binding:"required,min=1,max=10"`
	Description       *string `json:"description"`
	IsPublic          *bool   `json:"is_public"`
	Timezone          string  `json:"timezone"`
	DefaultAssigneeID *uint64 `json:"default_assignee_id"`
	TemplateID        *uint64 `json:"template_id"`
}

// ProjectUpdateRequest is the request body for updating a project.
type ProjectUpdateRequest struct {
	Name              string     `json:"name"`
	Description       *string    `json:"description"`
	IsPublic          *bool      `json:"is_public"`
	ArchivedAt        *string    `json:"archived_at"`        // ISO date string, null to unarchive
	DefaultAssigneeID *uint64    `json:"default_assignee_id"`
}
