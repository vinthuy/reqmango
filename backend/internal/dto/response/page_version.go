package response

import "time"

// PageVersionResponse is the API response for a page version.
type PageVersionResponse struct {
	ID            uint64     `json:"id"`
	PageID        uint64     `json:"page_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	ContentJSON   *string    `json:"content_json"`
	VersionNumber int        `json:"version_number"`
	ChangeSummary *string    `json:"change_summary"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedByID   *uint64    `json:"created_by_id"`
	CreatedByName string     `json:"created_by_name,omitempty"`
}
