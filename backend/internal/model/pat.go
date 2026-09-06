package model

import "time"

// PersonalAccessToken is a long-lived API credential issued to a user.
// The plaintext token is shown only once at creation; only the SHA-256
// hash is stored.
type PersonalAccessToken struct {
	BaseModel

	UserID      uint64     `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	TokenPrefix string     `gorm:"size:20;not null;index" json:"token_prefix"` // 展示用前缀
	TokenHash   string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	Scopes      string     `gorm:"type:text;default:''" json:"scopes"` // 预留，首版空
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"` // null = 永不过期
	RevokedAt   *time.Time `json:"revoked_at"`
}

func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }
