package model

import "time"

type ProjectUpdate struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ProjectID uint64    `gorm:"index" json:"project_id"`
	AuthorID  uint64    `json:"author_id"`
	Author    User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Status    string    `gorm:"size:20;not null" json:"status"` // on_track, at_risk, off_track
	Content   string    `gorm:"type:text" json:"content"`
	Recap     string    `gorm:"type:text" json:"recap,omitempty"`     // what was accomplished
	Plan      string    `gorm:"type:text" json:"plan,omitempty"`      // what's next
	Blockers  string    `gorm:"type:text" json:"blockers,omitempty"`  // obstacles
	Metrics   string    `gorm:"type:text" json:"metrics,omitempty"`   // JSON string of key metrics
	CreatedAt time.Time `json:"created_at"`
}
