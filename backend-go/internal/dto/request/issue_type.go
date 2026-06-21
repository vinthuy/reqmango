package request

// IssueTypeCreate request for creating a new issue type.
type IssueTypeCreate struct {
	Name      string  `json:"name" binding:"required"`
	Color     string  `json:"color"`
	Icon      string  `json:"icon"`
	IsDefault bool    `json:"is_default"`
	Sequence  int     `json:"sequence"`
	ProjectID *uint64 `json:"project_id"` // nil = workspace-shared
}

// IssueTypeUpdate request for updating an issue type.
// Uses pointer types to distinguish between "not provided" and "set to zero".
type IssueTypeUpdate struct {
	Name      *string  `json:"name"`
	Color     *string  `json:"color"`
	Icon      *string  `json:"icon"`
	IsDefault *bool    `json:"is_default"`
	Sequence  *int     `json:"sequence"`
	IsActive  *bool    `json:"is_active"`
	ProjectID *uint64  `json:"project_id"`
}

// IssueTypeDisable request for toggling active status.
type IssueTypeDisable struct {
	IsActive bool `json:"is_active"`
}

// IssueTypeFieldCreate request for associating a custom field to an issue type.
type IssueTypeFieldCreate struct {
	FieldID    uint64 `json:"field_id" binding:"required"`
	IsRequired bool   `json:"is_required"`
	Sequence   int    `json:"sequence"`
}

// IssueTypeFieldUpdate request for updating a field association.
type IssueTypeFieldUpdate struct {
	IsRequired *bool `json:"is_required"`
	Sequence   *int  `json:"sequence"`
}
