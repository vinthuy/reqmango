package request

type WorkItemTemplateCreate struct {
	Name        string                 `json:"name" binding:"required,max=100"`
	Description string                 `json:"description"`
	IssueTypeID *uint64                `json:"issue_type_id"`
	Defaults    map[string]interface{} `json:"defaults"`
	IsDefault   bool                   `json:"is_default"`
}

type WorkItemTemplateUpdate struct {
	Name        *string                 `json:"name"`
	Description *string                 `json:"description"`
	IssueTypeID *uint64                 `json:"issue_type_id"`
	Defaults    *map[string]interface{} `json:"defaults"`
	IsDefault   *bool                   `json:"is_default"`
}
