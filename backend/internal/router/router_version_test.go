package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/version"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// /version is how an operator finds out which build is deployed, so it has to
// answer without a token and stay outside the rate limiter. Registration order
// is the part that is easy to break: move it below the limiter and a burst of
// polls can start returning 429.
func TestVersionRouteIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	r := gin.New()
	SetupRoutes(r, db, config.Load())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/version", nil))

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, version.Version, body["version"])
	require.Equal(t, version.Commit, body["commit"])
	require.Equal(t, version.BuildDate, body["build_date"])
	require.NotEmpty(t, body["go_version"])
}
