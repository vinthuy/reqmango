package middleware

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// sensitiveQueryKeys lists the query parameters whose values must never reach the
// logs. SSE/chat clients authenticate with "?token=<jwt>" because EventSource
// cannot set an Authorization header, so a naive access log would persist raw
// credentials to disk.
var sensitiveQueryKeys = map[string]struct{}{
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"api_key":       {},
	"apikey":        {},
	"key":           {},
	"secret":        {},
	"password":      {},
	"pat":           {},
	"authorization": {},
	"auth":          {},
}

// RedactQuery replaces the value of every sensitive query parameter with
// "REDACTED" so credentials are never written to a log sink. Non-sensitive
// queries are returned untouched.
func RedactQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		// Never echo a query string that cannot be parsed: it may hold a credential.
		return "REDACTED"
	}

	redacted := false
	for key := range values {
		if _, ok := sensitiveQueryKeys[strings.ToLower(key)]; ok {
			values[key] = []string{"REDACTED"}
			redacted = true
		}
	}
	if !redacted {
		return rawQuery
	}
	return values.Encode()
}

// Logger returns middleware that logs each request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := RedactQuery(c.Request.URL.RawQuery)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[%d] %s %s | %v",
			statusCode,
			c.Request.Method,
			path,
			latency,
		)
	}
}
