package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns middleware that logs each request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

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
