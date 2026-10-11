package model

import "time"

// DataMigration marks a one-time data repair as applied so it does not rerun
// on every startup.
type DataMigration struct {
	Name      string    `gorm:"primaryKey;size:128" json:"name"`
	AppliedAt time.Time `gorm:"not null" json:"applied_at"`
}
