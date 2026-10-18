package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	h := NewHealthHandler(db)
	r := gin.New()
	r.GET("/health", h.Live)
	r.GET("/readyz", h.Ready)

	get := func(path string) (int, string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code, w.Body.String()
	}

	code, body := get("/health")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, `"ok"`)

	code, body = get("/readyz")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, `"ready"`)
	assert.Contains(t, body, `"database":"ok"`)

	// A closed pool must fail readiness but keep liveness green.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	code, _ = get("/readyz")
	assert.Equal(t, http.StatusServiceUnavailable, code)

	code, _ = get("/health")
	assert.Equal(t, http.StatusOK, code)
}
