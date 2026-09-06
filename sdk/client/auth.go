package client

import (
	"context"
	"strconv"
	"time"
)

// TokenResponse is the backend login response.
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Login exchanges credentials for a JWT (used once to mint a PAT).
func (c *Client) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	var out TokenResponse
	_, err := c.PostJSON(ctx, "/auth/login", nil,
		map[string]string{"email": email, "password": password}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// User is the current-user shape returned by /auth/me.
type User struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// Me returns the user behind the configured token.
func (c *Client) Me(ctx context.Context) (*User, error) {
	var out User
	if _, err := c.GetJSON(ctx, "/auth/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePATRequest is the body for POST /auth/tokens.
type CreatePATRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// PATResponse is the safe view of a token.
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

// CreatePATResponse carries the one-time plaintext token.
type CreatePATResponse struct {
	Token string `json:"token"`
	PATResponse
}

// CreatePAT mints a PAT using the current (JWT) token.
func (c *Client) CreatePAT(ctx context.Context, req CreatePATRequest) (*CreatePATResponse, error) {
	var out CreatePATResponse
	_, err := c.PostJSON(ctx, "/auth/tokens", nil, req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPATs returns the user's tokens.
func (c *Client) ListPATs(ctx context.Context) ([]PATResponse, error) {
	var out []PATResponse
	if _, err := c.GetJSON(ctx, "/auth/tokens", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RevokePAT revokes a token by ID.
func (c *Client) RevokePAT(ctx context.Context, id uint64) error {
	return c.DeleteJSON(ctx, "/auth/tokens/"+strconv.FormatUint(id, 10), nil, nil)
}
