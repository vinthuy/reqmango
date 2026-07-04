package model

// SavedDashboard represents a persisted dashboard configuration for a project.
type SavedDashboard struct {
	BaseModel

	Name        string  `gorm:"size:100;not null" json:"name"`
	Description *string `gorm:"size:500" json:"description"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	IsShared    bool    `gorm:"default:false" json:"is_shared"`
	OwnerID     uint64  `gorm:"not null;index" json:"owner_id"`
	ProjectID   uint64  `gorm:"not null;index" json:"project_id"`
	// Global date range filter applied to all widgets
	DateFrom *string `gorm:"size:20" json:"date_from"`
	DateTo   *string `gorm:"size:20" json:"date_to"`
	// Grid columns: 6 | 12 (default 12)
	Columns int `gorm:"default:12" json:"columns"`

	// Relationships
	Owner   User              `gorm:"foreignKey:OwnerID" json:"-"`
	Project Project           `gorm:"foreignKey:ProjectID" json:"-"`
	Widgets []DashboardWidget `gorm:"foreignKey:DashboardID" json:"widgets,omitempty"`
}

func (SavedDashboard) TableName() string {
	return "saved_dashboards"
}
