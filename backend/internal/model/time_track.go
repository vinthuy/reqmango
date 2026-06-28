package model

import "time"

// TimeTrack records time spent on an issue.
type TimeTrack struct {
	BaseModel

	IssueID     uint64     `gorm:"not null;index" json:"issue_id"`
	UserID      uint64     `gorm:"not null;index" json:"user_id"`
	Description *string    `gorm:"size:500" json:"description"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Duration    int64      `json:"duration"` // seconds, computed when stopped

	Issue *Issue `gorm:"foreignKey:IssueID" json:"-"`
	User  *User  `gorm:"foreignKey:UserID" json:"-"`
}

func (TimeTrack) TableName() string {
	return "time_tracks"
}
