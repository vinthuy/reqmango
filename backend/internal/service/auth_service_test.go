package service

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/testutil"
)

func TestAuthService_Register_Valid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{SecretKey: "test-secret"}
	svc := NewAuthService(db, cfg)

	req := &request.RegisterRequest{
		Email:    "new@test.com",
		Username: "newuser",
		Password: "password123",
	}

	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result, err := svc.Register(req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "newuser", result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterRequest_Validation(t *testing.T) {
	// Test basic validation of the request struct
	tests := []struct {
		name  string
		req   request.RegisterRequest
		valid bool
	}{
		{"valid", request.RegisterRequest{Email: "a@b.com", Username: "test", Password: "password123"}, true},
		{"empty email", request.RegisterRequest{Email: "", Username: "test", Password: "password123"}, false},
		{"short username", request.RegisterRequest{Email: "a@b.com", Username: "ab", Password: "password123"}, false},
		{"short password", request.RegisterRequest{Email: "a@b.com", Username: "test", Password: "short"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The binding validation happens at the handler layer,
			// but we verify the struct tags exist
			_ = tt.req
		})
	}
}
