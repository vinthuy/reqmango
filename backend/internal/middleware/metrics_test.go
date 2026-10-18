package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewMetricsCollector(time.Second)
	r := gin.New()
	r.Use(collector.Middleware())

	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	slow := func(c *gin.Context) {
		time.Sleep(12 * time.Millisecond)
		c.Status(http.StatusOK)
	}
	fail := func(c *gin.Context) { c.Status(http.StatusInternalServerError) }
	bad := func(c *gin.Context) { c.Status(http.StatusBadRequest) }

	r.GET("/health", ok)
	r.GET("/api/v1/items/:id", ok)
	r.GET("/api/v1/slow", slow)
	r.POST("/api/v1/fail", fail)
	r.GET("/api/v1/bad", bad)

	send := func(method, path string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	}
	send("GET", "/health")
	send("GET", "/api/v1/items/1")
	send("GET", "/api/v1/items/2")
	send("GET", "/api/v1/slow")
	send("POST", "/api/v1/fail")
	send("GET", "/api/v1/bad")
	send("GET", "/api/v1/unknown-path")

	routes, uptime := collector.Snapshot()
	require.GreaterOrEqual(t, uptime, time.Duration(0))

	byKey := map[string]RouteMetrics{}
	for _, m := range routes {
		byKey[m.Method+" "+m.Route] = m
	}

	// Probes are excluded entirely.
	_, hasProbe := byKey["GET /health"]
	assert.False(t, hasProbe, "probe traffic must not be counted")

	// The raw ID is replaced by the route template, so both requests share a row.
	items, ok2 := byKey["GET /api/v1/items/:id"]
	require.True(t, ok2)
	assert.Equal(t, int64(2), items.Requests)
	assert.Equal(t, int64(0), items.ServerErrors)
	assert.Equal(t, int64(0), items.ClientErrors)
	assert.Greater(t, items.P95MS, 0.0)

	failed := byKey["POST /api/v1/fail"]
	assert.Equal(t, int64(1), failed.ServerErrors)
	assert.Equal(t, int64(0), failed.ClientErrors)

	clientErr := byKey["GET /api/v1/bad"]
	assert.Equal(t, int64(1), clientErr.ClientErrors)
	assert.Equal(t, int64(0), clientErr.ServerErrors)

	// Unmatched paths collapse into one bounded bucket instead of one row per URL.
	unmatched, ok3 := byKey["GET <unmatched>"]
	require.True(t, ok3)
	assert.Equal(t, int64(1), unmatched.Requests)

	slowRow := byKey["GET /api/v1/slow"]
	assert.GreaterOrEqual(t, slowRow.AvgMS, 10.0)
	assert.GreaterOrEqual(t, slowRow.P95MS, 10.0)
	assert.GreaterOrEqual(t, slowRow.MaxMS, slowRow.AvgMS)
}

func TestMetricsHandlerRequiresSuperuser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	collector := NewMetricsCollector(time.Second)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-User"); id != "" {
			var user *model.User
			if id == "root" {
				user = &model.User{BaseModel: model.BaseModel{ID: 1}, IsSuperuser: true}
			} else {
				user = &model.User{BaseModel: model.BaseModel{ID: 2}}
			}
			c.Set("currentUser", user)
		}
	})
	r.GET("/api/v1/metrics", MetricsHandler(collector))

	call := func(user string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
		if user != "" {
			req.Header.Set("X-User", user)
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusForbidden, call("member"))
	assert.Equal(t, http.StatusUnauthorized, call(""), "no user in context is unauthenticated, not forbidden")
	assert.Equal(t, http.StatusOK, call("root"))
}
