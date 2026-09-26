package model

import "encoding/json"

// DashboardWidget represents a single widget within a dashboard.
type DashboardWidget struct {
	BaseModel

	DashboardID uint64 `gorm:"not null;index" json:"dashboard_id"`
	// WidgetType: number_card | bar_chart | pie_chart | doughnut_chart | line_chart |
	// bubble_chart | scatter_chart | mixed_chart | metric_chart | burndown | table | recent_list | saved_report | ai_summary
	WidgetType string `gorm:"size:30;not null" json:"widget_type"`
	Title      string `gorm:"size:100" json:"title"`
	// Description is an optional tooltip/help text
	Description *string `gorm:"size:255" json:"description"`

	// Config stores widget-type-specific configuration (JSONB).
	// number_card: {"metric":"total|completed|in_progress|overdue","label":"...","color":"#xxx"}
	// metric_chart: {"metric_chart_id":123}
	// bar/pie/doughnut/table: {"report_type":"distribution","group_by":"state","interval":"week","rql":"..."}
	// line/bubble/scatter/mixed: {"report_type":"created_trend","group_by":"state","interval":"week","rql":"..."}
	// burndown: {"cycle_id":123}
	// recent_list: {"limit":10}
	Config json.RawMessage `gorm:"type:jsonb" json:"config"`

	// Position stores grid position (JSONB): {"x":0,"y":0,"w":6,"h":4}
	Position json.RawMessage `gorm:"type:jsonb" json:"position"`

	SortOrder int `gorm:"default:0" json:"sort_order"`

	// Relationship
	Dashboard SavedDashboard `gorm:"foreignKey:DashboardID" json:"-"`
}

func (DashboardWidget) TableName() string {
	return "dashboard_widgets"
}
