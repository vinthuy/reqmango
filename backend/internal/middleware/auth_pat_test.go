package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/reqmango/backend/internal/testutil"
)

func TestAuthMiddleware_PAT_Valid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ping", AuthMiddleware(db, "secret"), func(c *gin.Context) {
		u := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"uid": u.ID})
	})

	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}).
			AddRow(1, 7, service.HashToken("reqmango_pat_testtoken"), nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(7, true))
	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_testtoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":7`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthMiddleware_PAT_Invalid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ping", AuthMiddleware(db, "secret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_bogus")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var _ = model.User{} // keep model import for future tests
