package request

// WorkspaceCreateRequest is the request body for creating a workspace.
type WorkspaceCreateRequest struct {
	Name             string `json:"name" binding:"required,min=1,max=255"`
	Slug             string `json:"slug" binding:"required,min=1,max=50"`
	OrganizationSize string `json:"organization_size"`
	Timezone         string `json:"timezone"`
}

// WorkspaceUpdateRequest is the request body for updating a workspace.
type WorkspaceUpdateRequest struct {
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
	Timezone string `json:"timezone"`
}
