package response

type ReleaseResponse struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Version     string  `json:"version"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	ReleaseDate *string `json:"release_date"`
	ProjectID   uint64  `json:"project_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ReleaseProgressResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	TotalIssues int    `json:"total_issues"`
	DoneIssues  int    `json:"done_issues"`
	Progress    int    `json:"progress"`
}