package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type SSEHandler struct {
	db     *gorm.DB
	secret string
}

func NewSSEHandler(db *gorm.DB, cfg *config.Config) *SSEHandler {
	return &SSEHandler{db: db, secret: cfg.SecretKey}
}

// Connect handles GET /api/v1/sse — Server-Sent Events stream.
func (h *SSEHandler) Connect(c *gin.Context) {
	user := h.authenticate(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	client := service.SSE.Register(user.ID)
	defer service.SSE.Unregister(client)

	// Send connected event
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"message\":\"SSE connected\"}\n\n")
	c.Writer.Flush()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-client.Ch:
			if !ok { return }
			io.WriteString(c.Writer, msg)
			c.Writer.Flush()
		}
	}
}

func (h *SSEHandler) authenticate(c *gin.Context) *model.User {
	// Try header first
	tokenStr := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			tokenStr = parts[1]
		}
	}
	// Fallback to query param (for EventSource API which doesn't support custom headers)
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	if tokenStr == "" {
		return nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.secret), nil
	})
	if err != nil || !token.Valid {
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil
	}

	userID, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return nil
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return nil
	}
	return &user
}
