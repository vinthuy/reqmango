package request

// PageTemplateCreateRequest creates a page template.
type PageTemplateCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Content     string  `json:"content"`
	ContentJSON *string `json:"content_json"`
	IsDefault   bool    `json:"is_default"`
	ProjectID   *uint64 `json:"project_id"`
}

// PageTemplateUpdateRequest updates a page template.
type PageTemplateUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	ContentJSON *string `json:"content_json"`
	IsDefault   *bool   `json:"is_default"`
}
