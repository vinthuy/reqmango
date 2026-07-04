package response

import (
	"encoding/json"
	"time"
)

// DashboardResponse is the API response for a saved dashboard.
type DashboardResponse struct {
	ID          uint64           `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	IsDefault   bool             `json:"is_default"`
	IsShared    bool             `json:"is_shared"`
	OwnerID     uint64           `json:"owner_id"`
	ProjectID   uint64           `json:"project_id"`
	DateFrom    *string          `json:"date_from"`
	DateTo      *string          `json:"date_to"`
	Columns     int              `json:"columns"`
	Widgets     []WidgetResponse `json:"widgets"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// WidgetResponse is the API response for a dashboard widget.
type WidgetResponse struct {
	ID          uint64          `json:"id"`
	DashboardID uint64          `json:"dashboard_id"`
	WidgetType  string          `json:"widget_type"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Config      json.RawMessage `json:"config"`
	Position    json.RawMessage `json:"position"`
	SortOrder   int             `json:"sort_order"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WidgetDataResponse carries the rendered data for a single widget.
type WidgetDataResponse struct {
	WidgetID uint64          `json:"widget_id"`
	Data     json.RawMessage `json:"data"`
}

// DashboardFullResponse returns the dashboard metadata plus rendered widget data.
type DashboardFullResponse struct {
	Dashboard  DashboardResponse    `json:"dashboard"`
	WidgetData []WidgetDataResponse `json:"widget_data"`
}
