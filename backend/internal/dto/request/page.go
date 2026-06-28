package request

// PageCreateRequest represents the request to create a page.
type PageCreateRequest struct {
	Title       string  `json:"title" binding:"required"`
	Content     string  `json:"content"`
	ContentJSON *string `json:"content_json"`
	ParentID    *uint64 `json:"parent_id"`
	Sequence    int     `json:"sequence"`
}

// PageUpdateRequest represents the request to update a page.
type PageUpdateRequest struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	ContentJSON *string `json:"content_json"`
	Published   *bool   `json:"published"`
	Sequence    *int    `json:"sequence"`
}

// PageMoveRequest represents the request to move a page in the hierarchy.
type PageMoveRequest struct {
	ParentID *uint64 `json:"parent_id"`
	Sequence int     `json:"sequence"`
}
