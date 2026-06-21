package request

// CustomFieldCreate request for creating a new custom field.
type CustomFieldCreate struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	FieldType    string  `json:"field_type" binding:"required"` // text/number/dropdown/boolean/date/member/url
	IsRequired   bool    `json:"is_required"`
	DefaultValue string  `json:"default_value"`
	Placeholder  string  `json:"placeholder"`
	ProjectID    *uint64 `json:"project_id"` // nil = workspace-shared
}

// CustomFieldUpdate request for updating a custom field.
type CustomFieldUpdate struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	FieldType    *string `json:"field_type"`
	IsRequired   *bool   `json:"is_required"`
	DefaultValue *string `json:"default_value"`
	Placeholder  *string `json:"placeholder"`
	IsActive     *bool   `json:"is_active"`
	ProjectID    *uint64 `json:"project_id"`
}

// CustomFieldOptionCreate request for adding a dropdown option.
type CustomFieldOptionCreate struct {
	Value    string `json:"value" binding:"required"`
	Color    string `json:"color"`
	Sequence int    `json:"sequence"`
}

// CustomFieldOptionUpdate request for updating a dropdown option.
type CustomFieldOptionUpdate struct {
	Value    *string `json:"value"`
	Color    *string `json:"color"`
	Sequence *int    `json:"sequence"`
}

// IssueCustomFieldValueCreate request for setting a single custom field value on an issue.
type IssueCustomFieldValueCreate struct {
	FieldID uint64 `json:"field_id" binding:"required"`
	Value   string `json:"value"`
}

// IssueCustomFieldValueUpdate request for updating a custom field value.
type IssueCustomFieldValueUpdate struct {
	Value string `json:"value"`
}

// BulkCustomFieldValueUpdate request for batch-updating custom field values.
type BulkCustomFieldValueUpdate struct {
	IssueID uint64                           `json:"issue_id"`
	Values  []IssueCustomFieldValueUpdateItem `json:"values"`
}

// IssueCustomFieldValueUpdateItem is a single entry in a bulk update.
type IssueCustomFieldValueUpdateItem struct {
	FieldID uint64 `json:"field_id"`
	Value   string `json:"value"`
}
