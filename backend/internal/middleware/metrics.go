package middleware

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// latencyBucketsMS are the upper bounds of the latency histogram. Percentiles
// are estimated from these, so the bucket edges are the resolution.
var latencyBucketsMS = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// probePaths are excluded from metrics and slow-request logs: orchestrators
// poll them constantly and they would dominate both the counts and the
// latency distribution without saying anything about real traffic.
var probePaths = map[string]bool{"/health": true, "/readyz": true}

type routeKey struct {
	Method string
	Route  string
}

type routeStats struct {
	Requests     int64
	ClientErrors int64 // 4xx
	ServerErrors int64 // 5xx
	TotalMS      float64
	MaxMS        float64
	Buckets      []int64 // one per latencyBucketsMS entry
	Overflow     int64   // slower than the last bucket
}

// MetricsCollector aggregates request counts and latency per route template.
// Grouping by the route template (not the raw path) keeps cardinality bounded
// by the number of registered routes.
type MetricsCollector struct {
	mu            sync.Mutex
	routes        map[routeKey]*routeStats
	started       time.Time
	slowThreshold time.Duration
}

func NewMetricsCollector(slowThreshold time.Duration) *MetricsCollector {
	return &MetricsCollector{
		routes:        map[routeKey]*routeStats{},
		started:       time.Now(),
		slowThreshold: slowThreshold,
	}
}

// Middleware records every request and logs the ones that exceed the slow
// threshold. It replaces nothing: it runs alongside the access logger.
func (m *MetricsCollector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "<unmatched>"
		}
		if probePaths[route] {
			return
		}
		elapsed := time.Since(start)
		status := c.Writer.Status()
		ms := float64(elapsed) / float64(time.Millisecond)

		m.mu.Lock()
		stats := m.routes[routeKey{Method: c.Request.Method, Route: route}]
		if stats == nil {
			stats = &routeStats{Buckets: make([]int64, len(latencyBucketsMS))}
			m.routes[routeKey{Method: c.Request.Method, Route: route}] = stats
		}
		stats.Requests++
		stats.TotalMS += ms
		if ms > stats.MaxMS {
			stats.MaxMS = ms
		}
		switch {
		case status >= 500:
			stats.ServerErrors++
		case status >= 400:
			stats.ClientErrors++
		}
		recorded := false
		for i, bound := range latencyBucketsMS {
			if ms <= bound {
				stats.Buckets[i]++
				recorded = true
				break
			}
		}
		if !recorded {
			stats.Overflow++
		}
		m.mu.Unlock()

		if m.slowThreshold > 0 && elapsed >= m.slowThreshold {
			log.Printf("[SLOW] %s %s %d %v (threshold %v)",
				c.Request.Method, route, status, elapsed.Round(time.Millisecond), m.slowThreshold)
		}
	}
}

// RouteMetrics is the exported shape of one route's counters.
type RouteMetrics struct {
	Method       string  `json:"method"`
	Route        string  `json:"route"`
	Requests     int64   `json:"requests"`
	ClientErrors int64   `json:"client_errors"`
	ServerErrors int64   `json:"server_errors"`
	AvgMS        float64 `json:"avg_ms"`
	MaxMS        float64 `json:"max_ms"`
	P95MS        float64 `json:"p95_ms"`
}

// Snapshot returns the counters ordered by route, for a human or a scrape.
func (m *MetricsCollector) Snapshot() ([]RouteMetrics, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]RouteMetrics, 0, len(m.routes))
	for key, s := range m.routes {
		avg := 0.0
		if s.Requests > 0 {
			avg = s.TotalMS / float64(s.Requests)
		}
		out = append(out, RouteMetrics{
			Method:       key.Method,
			Route:        key.Route,
			Requests:     s.Requests,
			ClientErrors: s.ClientErrors,
			ServerErrors: s.ServerErrors,
			AvgMS:        round1(avg),
			MaxMS:        round1(s.MaxMS),
			P95MS:        round1(percentile(s, 0.95)),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Route != out[j].Route {
			return out[i].Route < out[j].Route
		}
		return out[i].Method < out[j].Method
	})
	return out, time.Since(m.started)
}

// percentile estimates the p-th latency from the cumulative histogram. The
// estimate is the upper bound of the bucket the percentile falls into, which
// is accurate to the bucket resolution and never below the true value.
func percentile(s *routeStats, p float64) float64 {
	if s.Requests == 0 {
		return 0
	}
	target := int64(float64(s.Requests) * p)
	if float64(target)/float64(s.Requests) < p {
		target++ // round up so p95 of 1 request is that request
	}
	cumulative := int64(0)
	for i, count := range s.Buckets {
		cumulative += count
		if cumulative >= target {
			return latencyBucketsMS[i]
		}
	}
	if s.Overflow > 0 {
		return s.MaxMS
	}
	return s.MaxMS
}

func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}

// MetricsHandler serves the collector's snapshot. Metrics are process-wide and
// span every tenant, so only a superuser may read them.
func MetricsHandler(m *MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if !user.IsSuperuser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Superuser required"})
			return
		}
		routes, uptime := m.Snapshot()
		c.JSON(http.StatusOK, gin.H{
			"uptime_seconds":    int64(uptime.Seconds()),
			"slow_threshold_ms": int64(m.slowThreshold / time.Millisecond),
			"routes":            routes,
		})
	}
}
