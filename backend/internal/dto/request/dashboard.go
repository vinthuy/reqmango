package request

import "encoding/json"

// CreateDashboardRequest represents the request to create a saved dashboard.
type CreateDashboardRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description *string               `json:"description"`
	IsDefault   bool                  `json:"is_default"`
	IsShared    bool                  `json:"is_shared"`
	DateFrom    *string               `json:"date_from"`
	DateTo      *string               `json:"date_to"`
	Columns     int                   `json:"columns"`
	Widgets     []CreateWidgetRequest `json:"widgets"`
}

// UpdateDashboardRequest represents the request to update a saved dashboard.
type UpdateDashboardRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"is_default"`
	IsShared    *bool   `json:"is_shared"`
	DateFrom    *string `json:"date_from"`
	DateTo      *string `json:"date_to"`
	Columns     *int    `json:"columns"`
}

// CreateWidgetRequest represents the request to add a widget.
type CreateWidgetRequest struct {
	WidgetType  string          `json:"widget_type" binding:"required"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Config      json.RawMessage `json:"config"`
	Position    json.RawMessage `json:"position"`
	SortOrder   int             `json:"sort_order"`
}

// UpdateWidgetRequest represents the request to update a widget.
type UpdateWidgetRequest struct {
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	Config      *json.RawMessage `json:"config"`
	Position    *json.RawMessage `json:"position"`
	SortOrder   *int             `json:"sort_order"`
}

// ReorderWidgetsRequest represents the request to reorder widgets.
type ReorderWidgetsRequest struct {
	WidgetIDs []uint64 `json:"widget_ids" binding:"required"`
}
