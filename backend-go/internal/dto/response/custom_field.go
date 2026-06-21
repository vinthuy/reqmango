package response

import "time"

// CustomFieldResponse is the API response for a single custom field.
type CustomFieldResponse struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	FieldType    string    `json:"field_type"`
	IsRequired   bool      `json:"is_required"`
	DefaultValue string    `json:"default_value"`
	Placeholder  string    `json:"placeholder"`
	IsActive     bool      `json:"is_active"`
	ProjectID    *uint64   `json:"project_id"`
	WorkspaceID  uint64    `json:"workspace_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Options      []CustomFieldOptionResponse `json:"options,omitempty"`
}

// CustomFieldLite is a minimal custom field reference.
type CustomFieldLite struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	FieldType string `json:"field_type"`
}

// CustomFieldOptionResponse represents a dropdown option.
type CustomFieldOptionResponse struct {
	ID       uint64 `json:"id"`
	FieldID  uint64 `json:"field_id"`
	Value    string `json:"value"`
	Color    string `json:"color"`
	Sequence int    `json:"sequence"`
}

// IssueCustomFieldValueResponse represents a single custom field value on an issue.
type IssueCustomFieldValueResponse struct {
	IssueID   uint64   `json:"issue_id"`
	FieldID   uint64   `json:"field_id"`
	Value     string   `json:"value"`
	FieldName string   `json:"field_name,omitempty"`
	FieldType string   `json:"field_type,omitempty"`
}

// IssueCustomFieldsResponse bundles field definitions with their values for an issue.
type IssueCustomFieldsResponse struct {
	IssueID uint64                           `json:"issue_id"`
	Fields  []FieldWithValue                 `json:"fields"`
}

// FieldWithValue pairs a custom field definition with its current value for an issue.
type FieldWithValue struct {
	CustomFieldResponse
	Value string `json:"value"`
}
