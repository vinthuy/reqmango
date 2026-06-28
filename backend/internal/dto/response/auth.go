package response

import "time"

// UserResponse is the public representation of a user (no password).
type UserResponse struct {
	ID              uint64     `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	DisplayName     string     `json:"display_name"`
	FirstName       *string    `json:"first_name"`
	LastName        *string    `json:"last_name"`
	Avatar          *string    `json:"avatar"`
	AvatarURL       *string    `json:"avatar_url"`
	CoverImageURL   *string    `json:"cover_image_url"`
	IsActive        bool       `json:"is_active"`
	IsSuperuser     bool       `json:"is_superuser"`
	IsEmailVerified bool       `json:"is_email_verified"`
	UserTimezone    string     `json:"user_timezone"`
	LastActive      *time.Time `json:"last_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedByID     *uint64    `json:"created_by"`
	UpdatedByID     *uint64    `json:"updated_by"`
	DeletedAt       *time.Time `json:"deleted_at"`
	IsDeleted       bool       `json:"is_deleted"`
}

// TokenResponse is returned after successful login.
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// UserLite is a compact user representation for lists/relations.
type UserLite struct {
	ID          uint64  `json:"id"`
	DisplayName string  `json:"display_name"`
	Email       string  `json:"email"`
	AvatarURL   *string `json:"avatar_url"`
}
