package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserInfo is the minimal user info extracted from JWT.
type UserInfo struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	WorkspaceID uint64 `json:"workspace_id,omitempty"`
}

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token claims"})
			c.Abort()
			return
		}

		// Extract user info from JWT claims. The main backend issues tokens with
		// the user ID in the "sub" claim as a decimal string (see backend auth_service).
		var userID uint64
		if sub, ok := claims["sub"].(string); ok {
			userID, _ = strconv.ParseUint(sub, 10, 64)
		}
		email, _ := claims["email"].(string)
		wsID, _ := claims["workspace_id"].(float64)

		user := &UserInfo{
			ID:          uint64(userID),
			Email:       email,
			WorkspaceID: uint64(wsID),
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

// GetCurrentUser extracts the user from gin context.
func GetCurrentUser(c *gin.Context) *UserInfo {
	if user, exists := c.Get("currentUser"); exists {
		if u, ok := user.(*UserInfo); ok {
			return u
		}
	}
	return nil
}
