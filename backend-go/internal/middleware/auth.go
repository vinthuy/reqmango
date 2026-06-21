package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

// AuthMiddleware validates JWT tokens and sets the current user in context.
func AuthMiddleware(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			c.Abort()
			return
		}

		// Extract user ID from "sub" claim
		sub, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token subject"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Look up user
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			c.Abort()
			return
		}

		if !user.IsActive || user.DeletedAt.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Account is disabled"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("currentUser", &user)
		c.Next()
	}
}

// GetCurrentUser extracts the current user from the Gin context.
// Panics if the user is not set (should be called only in routes using AuthMiddleware).
func GetCurrentUser(c *gin.Context) *model.User {
	user, exists := c.Get("currentUser")
	if !exists {
		panic("currentUser not found in context - ensure AuthMiddleware is applied to this route")
	}
	return user.(*model.User)
}
