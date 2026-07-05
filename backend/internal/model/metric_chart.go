package model

import "time"

// MetricChart 度量图表配置
type MetricChart struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  uint64    `gorm:"not null;index" json:"project_id"`
	CreatorID  uint64    `gorm:"not null" json:"creator_id"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	TemplateID string    `gorm:"size:100" json:"template_id"`
	ChartType  string    `gorm:"size:50;not null" json:"chart_type"`
	XAxis      string    `gorm:"size:100;not null" json:"x_axis"`
	YAxis      string    `gorm:"size:100;not null" json:"y_axis"`
	Filters    string    `gorm:"type:jsonb;default:'{}'" json:"filters"`
	Config     string    `gorm:"type:jsonb;default:'{}'" json:"config"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	IsVisible  bool      `gorm:"default:true" json:"is_visible"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (MetricChart) TableName() string {
	return "metric_charts"
}
