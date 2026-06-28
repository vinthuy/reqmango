package request

type ReleaseCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Version     string  `json:"version" binding:"required"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	ReleaseDate *string `json:"release_date"`
}

type ReleaseUpdateRequest struct {
	Name        string  `json:"name"`
	Version     string  `json:"version"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	ReleaseDate *string `json:"release_date"`
}

type ReleaseIssueRequest struct {
	IssueIDs []uint64 `json:"issue_ids" binding:"required"`
}