package response

import "time"

// PageTemplateResponse is the API response for a page template.
type PageTemplateResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	ContentJSON *string   `json:"content_json"`
	IsDefault   bool      `json:"is_default"`
	WorkspaceID uint64    `json:"workspace_id"`
	ProjectID   *uint64   `json:"project_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
