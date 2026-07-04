package response

import "time"

// SavedReportResponse is the API response for a saved report.
type SavedReportResponse struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	ReportType string    `json:"report_type"`
	GroupBy    string    `json:"group_by"`
	ChartType  string    `json:"chart_type"`
	RQL        string    `json:"rql"`
	Interval   string    `json:"interval"`
	DateFrom   string    `json:"date_from"`
	DateTo     string    `json:"date_to"`
	ProjectID  uint64    `json:"project_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
