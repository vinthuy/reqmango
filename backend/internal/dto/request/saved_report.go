package request

// SavedReportCreateRequest represents the request to create a saved report.
type SavedReportCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	ReportType string `json:"report_type" binding:"required"` // distribution | created_vs_resolved | avg_age | current_age | created_trend
	GroupBy    string `json:"group_by"`                        // state | priority | assignee | type | label | cycle | module
	ChartType  string `json:"chart_type"`                      // bar | pie | doughnut | table | line
	RQL        string `json:"rql"`
	Interval   string `json:"interval"`    // day | week | month
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

// SavedReportUpdateRequest represents the request to update a saved report.
type SavedReportUpdateRequest struct {
	Name       *string `json:"name"`
	ReportType *string `json:"report_type"`
	GroupBy    *string `json:"group_by"`
	ChartType  *string `json:"chart_type"`
	RQL        *string `json:"rql"`
	Interval   *string `json:"interval"`
	DateFrom   *string `json:"date_from"`
	DateTo     *string `json:"date_to"`
}
