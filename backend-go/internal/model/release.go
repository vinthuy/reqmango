package model

import (
	"time"
)

type Release struct {
	BaseModel
	Name        string     `gorm:"size:100;not null" json:"name"`
	Version     string     `gorm:"size:50;not null" json:"version"`
	Description string     `gorm:"type:text" json:"description"`
	Status      string     `gorm:"size:30;default:planned" json:"status"`
	ReleaseDate *time.Time `json:"release_date"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
}

type ReleaseIssue struct {
	ReleaseID uint64 `gorm:"primaryKey;autoIncrement:false" json:"release_id"`
	IssueID   uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
}