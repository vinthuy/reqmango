package response

import "time"

// IssueTypeResponse is the API response for a single issue type.
type IssueTypeResponse struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	Icon         string    `json:"icon"`
	Description  string    `json:"description"`
	Level        int       `json:"level"`
	ParentTypeID *uint64   `json:"parent_type_id"`
	IsDefault    bool      `json:"is_default"`
	Sequence     int       `json:"sequence"`
	IsActive     bool      `json:"is_active"`
	ProjectID    *uint64   `json:"project_id"`
	WorkspaceID  uint64    `json:"workspace_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsInherited  bool      `json:"is_inherited"`
}

// IssueTypeLite is a minimal issue type reference used in issue responses.
type IssueTypeLite struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// IssueTypeFieldResponse represents a custom field linked to an issue type.
type IssueTypeFieldResponse struct {
	FieldID    uint64 `json:"field_id"`
	TypeID     uint64 `json:"type_id"`
	IsRequired bool   `json:"is_required"`
	Sequence   int    `json:"sequence"`
	// Embedded custom field info (populated from join)
	Name        string                      `json:"name,omitempty"`
	FieldType   string                      `json:"field_type,omitempty"`
	Description string                      `json:"description,omitempty"`
	Options     []CustomFieldOptionResponse `json:"options,omitempty"`
}
