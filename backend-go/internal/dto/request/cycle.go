package request

// CycleCreateRequest is the request body for creating a cycle.
// ProjectID is set from the URL path by the handler, not from JSON body.
type CycleCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	StartDate   string  `json:"start_date" binding:"required"`   // RFC3339 date
	EndDate     *string `json:"end_date"`                         // RFC3339 date, nullable
	Timezone    string  `json:"timezone"`
	ProjectID   uint64  `json:"-"`                                // Set from URL path by handler
}

// CycleUpdateRequest is the request body for updating a cycle.
type CycleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`   // RFC3339 date
	EndDate     *string `json:"end_date"`     // RFC3339 date
}
