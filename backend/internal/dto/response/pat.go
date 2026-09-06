package response

import "time"

// PATResponse is the safe (hash-less) view of a personal access token.
type PATResponse struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      string     `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PATCreateResponse is returned once when a token is created.
type PATCreateResponse struct {
	Token string `json:"token"`
	PATResponse
}
