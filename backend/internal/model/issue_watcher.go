package model

import "time"

type IssueWatcher struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID   uint64     `gorm:"index;not null" json:"issue_id"`
	UserID    uint64     `gorm:"index;not null" json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (IssueWatcher) TableName() string {
	return "issue_watchers"
}