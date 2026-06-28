package response

import (
	"encoding/json"
	"time"
)

type WorkItemTemplateResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IssueTypeID *uint64         `json:"issue_type_id"`
	IssueType   *IssueTypeLite  `json:"issue_type,omitempty"`
	Defaults    json.RawMessage `json:"defaults"`
	IsDefault   bool            `json:"is_default"`
	ProjectID   uint64          `json:"project_id"`
	WorkspaceID uint64          `json:"workspace_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
