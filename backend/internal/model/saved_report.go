package model

// SavedReport 保存的自定义报表配置
type SavedReport struct {
	BaseModel

	Name       string  `gorm:"size:100;not null" json:"name"`
	ReportType string  `gorm:"size:30;not null" json:"report_type"` // distribution | created_vs_resolved | avg_age | current_age | created_trend
	GroupBy    string  `gorm:"size:50" json:"group_by"`             // state | priority | assignee | type | label | cycle | module
	ChartType  string  `gorm:"size:20;default:bar" json:"chart_type"` // bar | pie | doughnut | table | line
	RQL        string  `gorm:"type:text" json:"rql"`                // RQL 筛选条件
	Interval   string  `gorm:"size:10" json:"interval"`             // day | week | month
	DateFrom   string  `gorm:"size:20" json:"date_from"`
	DateTo     string  `gorm:"size:20" json:"date_to"`
	ProjectID  uint64  `gorm:"not null;index" json:"project_id"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (SavedReport) TableName() string {
	return "saved_reports"
}
