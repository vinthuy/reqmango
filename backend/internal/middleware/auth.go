package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/i18n"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

// AuthMiddleware validates JWT tokens and sets the current user in context.
// msg returns a translated message for the given i18n key.
func msg(c *gin.Context, key, fallback string) string {
	lang := i18n.DetectLanguage(c.GetHeader("Accept-Language"))
	if ql := c.Query("lang"); ql != "" { lang = ql }
	if translated := i18n.T(lang, key); translated != key { return translated }
	return fallback
}

// AuthMiddleware validates JWT tokens and sets the current user in context.
func AuthMiddleware(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "missing_auth_header", "Missing authorization header")})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "invalid_format", "Invalid authorization header format")})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Personal access tokens take precedence: long-lived credentials
		// stored as SHA-256 hashes (used by CLI / MCP / CI).
		if strings.HasPrefix(tokenStr, service.PATPrefix) {
			patSvc := service.NewPATService(db)
			user, err := patSvc.Authenticate(tokenStr)
			if err != nil {
				if appErr, ok := err.(*common.AppError); ok {
					c.JSON(appErr.Code, gin.H{"message": appErr.Message})
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid token")})
				}
				c.Abort()
				return
			}
			c.Set("currentUser", user)
			c.Next()
			return
		}

		// Parse and validate JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid or expired token")})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid token claims")})
			c.Abort()
			return
		}

		// Extract user ID from "sub" claim
		sub, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid token subject")})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "unauthorized", "Invalid user ID in token")})
			c.Abort()
			return
		}

		// Look up user
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "user_not_found", "User not found")})
			c.Abort()
			return
		}

		if !user.IsActive || user.DeletedAt.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": msg(c, "forbidden", "Account is disabled")})
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

// GetUserID extracts the current user's ID from the Gin context.
func GetUserID(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists { return 0 }
	if u, ok := user.(*model.User); ok { return u.ID }
	return 0
}
