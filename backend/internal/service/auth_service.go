package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// Register creates a new user account.
func (s *AuthService) Register(req *request.RegisterRequest) (*response.UserResponse, error) {
	// Check email uniqueness
	var count int64
	s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Email already registered")
	}

	// Check username uniqueness
	s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.Internal("Failed to hash password")
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	timezone := req.UserTimezone
	if timezone == "" {
		timezone = "UTC"
	}

	user := &model.User{
		Email:        req.Email,
		Username:     req.Username,
		DisplayName:  displayName,
		PasswordHash: string(hashedPassword),
		UserTimezone: timezone,
	}

	if req.FirstName != "" {
		user.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		user.LastName = &req.LastName
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, common.Internal("Failed to create user")
	}

	return userToResponse(user), nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(req *request.LoginRequest) (*response.TokenResponse, error) {
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.Unauthorized("Invalid email or password")
		}
		return nil, common.Internal("Database error")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, common.Unauthorized("Invalid email or password")
	}

	if !user.IsActive {
		return nil, common.Unauthorized("Account is disabled")
	}

	// Generate JWT
	accessToken, expiresAt, err := s.generateToken(user.ID)
	if err != nil {
		return nil, common.Internal("Failed to generate token")
	}

	// Update last active
	now := time.Now()
	user.LastActive = &now
	s.db.Model(&user).Update("last_active", now)

	return &response.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

// GetCurrentUser returns the current user's profile.
func (s *AuthService) GetCurrentUser(userID uint64) (*response.UserResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("User not found")
		}
		return nil, common.Internal("Database error")
	}
	return userToResponse(&user), nil
}

// FindUserByID looks up a user by ID and returns the model.
func (s *AuthService) FindUserByID(userID uint64) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GenerateTokenForTest is a helper for tests to generate tokens without login.
func (s *AuthService) GenerateTokenForTest(userID uint64) (string, time.Time, error) {
	return s.generateToken(userID)
}

func (s *AuthService) generateToken(userID uint64) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.AccessTokenExpireMin) * time.Minute)

	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(userID, 10),
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

// userToResponse converts a User model to a UserResponse DTO.
func userToResponse(user *model.User) *response.UserResponse {
	resp := &response.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Avatar:          user.Avatar,
		IsActive:        user.IsActive,
		IsSuperuser:     user.IsSuperuser,
		IsEmailVerified: user.IsEmailVerified,
		UserTimezone:    user.UserTimezone,
		LastActive:      user.LastActive,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		CreatedByID:     user.CreatedByID,
		UpdatedByID:     user.UpdatedByID,
	}

	if user.DeletedAt.Valid {
		resp.DeletedAt = &user.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp
}
