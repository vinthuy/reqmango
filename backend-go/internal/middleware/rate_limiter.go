package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/model"
)

// RateLimiter implements a per-key token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	limit    int           // max requests per window
	window   time.Duration // time window
	cleanupInterval time.Duration
}

type bucket struct {
	tokens   int
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter.
// limit: max requests per window (0 = unlimited)
func NewRateLimiter(limit int, windowSec int) *RateLimiter {
	if limit <= 0 {
		return &RateLimiter{limit: 0}
	}
	if windowSec <= 0 { windowSec = 60 }
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		limit:    limit,
		window:   time.Duration(windowSec) * time.Second,
		cleanupInterval: 5 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, b := range rl.buckets {
			if now.Sub(b.lastRefill) > rl.window*2 {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin middleware that rate limits requests.
// If limit is 0, rate limiting is disabled (no-op).
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	// If limit is 0, skip rate limiting entirely
	if rl.limit <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		// Get rate limit key: user ID if authenticated, otherwise IP
		key := c.ClientIP()
		if user, exists := c.Get("currentUser"); exists {
			if u, ok := user.(*model.User); ok {
				key = "user:" + strconv.FormatUint(u.ID, 10)
			}
		}

		rl.mu.Lock()
		b, ok := rl.buckets[key]
		if !ok {
			b = &bucket{tokens: rl.limit, lastRefill: time.Now()}
			rl.buckets[key] = b
		}

		// Refill tokens based on elapsed time
		elapsed := time.Since(b.lastRefill)
		refill := int(float64(rl.limit) * elapsed.Seconds() / rl.window.Seconds())
		if refill > 0 {
			b.tokens += refill
			if b.tokens > rl.limit { b.tokens = rl.limit }
			b.lastRefill = time.Now()
		}

		b.tokens--
		allowed := b.tokens >= 0
		remaining := b.tokens
		if remaining < 0 { remaining = 0 }

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(b.lastRefill.Add(rl.window).Unix(), 10))

		rl.mu.Unlock()

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error_code": "RATE_LIMIT_EXCEEDED",
				"message":    "Too many requests. Please retry after " + rl.window.String(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Skipper allows skipping rate limiting for certain paths.
func RateLimitSkipper(paths ...string) gin.HandlerFunc {
	skip := make(map[string]bool, len(paths))
	for _, p := range paths { skip[p] = true }
	return func(c *gin.Context) {
		if skip[c.Request.URL.Path] {
			c.Next()
			return
		}
	}
}
