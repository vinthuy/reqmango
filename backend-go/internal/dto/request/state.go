package request

// StateCreateRequest is the request body for creating a state.
type StateCreateRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=255"`
	Color     string `json:"color"`
	Group     string `json:"group"`
	Sequence  *int   `json:"sequence"`
	IsDefault *bool  `json:"is_default"`
}

// StateUpdateRequest is the request body for updating a state.
type StateUpdateRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	Group     *string `json:"group"`
	Sequence  *int    `json:"sequence"`
	IsDefault *bool   `json:"is_default"`
	IsActive  *bool   `json:"is_active"`
}
