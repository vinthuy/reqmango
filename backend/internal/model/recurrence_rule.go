package model

import "time"

// RecurrenceRule defines auto-creation schedule for an issue.
type RecurrenceRule struct {
	BaseModel

	IssueID   uint64     `gorm:"not null;uniqueIndex" json:"issue_id"`
	Frequency string     `gorm:"size:20;not null" json:"frequency"` // daily | weekly | monthly | cron
	Interval  int        `gorm:"default:1" json:"interval"`          // every N periods
	CronExpr  *string    `gorm:"size:100" json:"cron_expr"`
	NextRun   time.Time  `gorm:"not null" json:"next_run"`
	EndDate   *time.Time `json:"end_date"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`

	Issue *Issue `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
}

func (RecurrenceRule) TableName() string { return "recurrence_rules" }
