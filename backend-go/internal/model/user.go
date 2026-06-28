package model

import "time"

// User represents a registered user in the system.
type User struct {
	BaseModel

	Email           string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username        string     `gorm:"uniqueIndex;size:128;not null" json:"username"`
	DisplayName     string     `gorm:"size:255;default:''" json:"display_name"`
	FirstName       *string    `gorm:"size:255" json:"first_name"`
	LastName        *string    `gorm:"size:255" json:"last_name"`
	Avatar          *string    `gorm:"type:text" json:"avatar"`
	PasswordHash    string     `gorm:"size:255;not null" json:"-"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	IsSuperuser     bool       `gorm:"default:false" json:"is_superuser"`
	IsEmailVerified bool       `gorm:"default:false" json:"is_email_verified"`
	UserTimezone    string     `gorm:"size:255;default:UTC" json:"user_timezone"`
	LastActive      *time.Time `json:"last_active"`

	// Relationships
	Workspaces      []WorkspaceMember `gorm:"foreignKey:UserID" json:"-"`
	Projects        []ProjectMember   `gorm:"foreignKey:UserID" json:"-"`
	AssignedIssues  []IssueAssignee   `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// GetID returns the user ID (used by middleware).
func (u User) GetID() uint64 { return u.ID }

// IsSuper returns true if the user is a superuser (used by authorization middleware).
func (u User) IsSuper() bool { return u.IsSuperuser }
