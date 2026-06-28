package request

// LabelCreateRequest is the request body for creating a label.
type LabelCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Color       string  `json:"color"`
	Description *string `json:"description"`
}

// LabelUpdateRequest is the request body for updating a label.
type LabelUpdateRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}
