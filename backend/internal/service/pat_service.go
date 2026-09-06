package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
)

// PATPrefix marks a personal access token in the Authorization header.
const PATPrefix = "reqmango_pat_"

// PATService manages personal access tokens.
type PATService struct {
	db *gorm.DB
}

func NewPATService(db *gorm.DB) *PATService { return &PATService{db: db} }

// HashToken returns the hex SHA-256 hash of a plaintext token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create issues a new PAT for the user. The plaintext token is returned
// exactly once and is never stored.
func (s *PATService) Create(userID uint64, name string, expiresAt *time.Time) (string, *model.PersonalAccessToken, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, common.Internal("failed to generate token")
	}
	token := PATPrefix + hex.EncodeToString(raw)

	pat := &model.PersonalAccessToken{
		UserID:      userID,
		Name:        name,
		TokenPrefix: token[:len(PATPrefix)+4],
		TokenHash:   HashToken(token),
		ExpiresAt:   expiresAt,
	}
	if err := s.db.Create(pat).Error; err != nil {
		return "", nil, common.Internal("failed to save token")
	}
	return token, pat, nil
}

// List returns the user's PATs. TokenHash is already hidden by its json:"-" tag.
func (s *PATService) List(userID uint64) ([]model.PersonalAccessToken, error) {
	var pats []model.PersonalAccessToken
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&pats).Error; err != nil {
		return nil, common.Internal("failed to list tokens")
	}
	return pats, nil
}

// Revoke marks a token revoked. Returns ErrNotFound if it is not owned by the user.
func (s *PATService) Revoke(userID, id uint64) error {
	res := s.db.Model(&model.PersonalAccessToken{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("revoked_at", time.Now())
	if res.Error != nil {
		return common.Internal("failed to revoke token")
	}
	if res.RowsAffected == 0 {
		return common.NotFound("token not found")
	}
	return nil
}

// Authenticate resolves a plaintext PAT to its user, checking revocation,
// expiry and account status.
func (s *PATService) Authenticate(tokenStr string) (*model.User, error) {
	if !strings.HasPrefix(tokenStr, PATPrefix) {
		return nil, common.Unauthorized("invalid token")
	}

	var pat model.PersonalAccessToken
	if err := s.db.Where("token_hash = ? AND revoked_at IS NULL", HashToken(tokenStr)).First(&pat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.Unauthorized("invalid token")
		}
		return nil, common.Internal("failed to look up token")
	}
	if pat.ExpiresAt != nil && time.Now().After(*pat.ExpiresAt) {
		return nil, common.Unauthorized("token expired")
	}

	var user model.User
	if err := s.db.First(&user, pat.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.Unauthorized("user not found")
		}
		return nil, common.Internal("failed to load user")
	}
	if !user.IsActive || user.DeletedAt.Valid {
		return nil, common.Unauthorized("account is disabled")
	}

	// Fire the last_used_at refresh (blocking keeps sqlmock tests simple).
	if err := s.db.Model(&pat).Update("last_used_at", time.Now()).Error; err != nil {
		return nil, common.Internal("failed to refresh token")
	}
	return &user, nil
}
