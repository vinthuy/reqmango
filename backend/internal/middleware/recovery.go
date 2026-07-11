package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Recovery] panic recovered:\n%s\n%s", err, debug.Stack())

				c.Header("Content-Type", "application/json")
				c.Status(http.StatusInternalServerError)

				response := map[string]string{
					"message": "Internal server error",
				}
				json.NewEncoder(c.Writer).Encode(response)
			}
		}()

		c.Next()
	}
}
