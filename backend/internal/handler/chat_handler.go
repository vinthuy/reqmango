package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

// ChatHandler exposes chat HTTP + SSE endpoints.
type ChatHandler struct {
	svc    *service.ChatService
	db     *gorm.DB
	secret string
}

// NewChatHandler creates a ChatHandler. db and secret are used for SSE JWT auth.
func NewChatHandler(svc *service.ChatService, db *gorm.DB, secret string) *ChatHandler {
	return &ChatHandler{svc: svc, db: db, secret: secret}
}

// GetOrCreateForIssue: GET /issues/:issueId/chat
func (h *ChatHandler) GetOrCreateForIssue(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	resp, err := h.svc.GetOrCreateForIssue(issueID, user.ID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetChat: GET /chats/:chatId
func (h *ChatHandler) GetChat(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	// Reuse GetOrCreateForIssue's shape minus messages: verify membership then return
	if err := h.svc.GetChatMembershipCheck(chatID, user.ID); err != nil {
		h.writeError(c, err)
		return
	}
	resp, err := h.svc.GetChat(chatID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListMessages: GET /chats/:chatId/messages
func (h *ChatHandler) ListMessages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	var q request.ListMessagesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query"})
		return
	}
	resp, err := h.svc.ListMessages(chatID, user.ID, q)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendMessage: POST /chats/:chatId/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	var req request.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.SendMessage(chatID, user.ID, req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// EditMessage: PUT /messages/:messageId
func (h *ChatHandler) EditMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.EditMessage(messageID, user.ID, req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteMessage: DELETE /messages/:messageId
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	if err := h.svc.DeleteMessage(messageID, user.ID); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// AddReaction: POST /messages/:messageId/reactions
func (h *ChatHandler) AddReaction(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	if err := h.svc.AddReaction(messageID, user.ID, req.Emoji); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
}

// RemoveReaction: DELETE /messages/:messageId/reactions
func (h *ChatHandler) RemoveReaction(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	messageID, _ := strconv.ParseUint(c.Param("messageId"), 10, 64)
	var req request.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	if err := h.svc.RemoveReaction(messageID, user.ID, req.Emoji); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

// Stream: GET /chats/:chatId/stream — SSE long connection.
// Auth via JWT in ?token= query param (EventSource cannot set headers).
func (h *ChatHandler) Stream(c *gin.Context) {
	userID, chatID, ok := h.authSSE(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
		return
	}
	// Verify chat membership before subscribing
	if err := h.svc.GetChatMembershipCheck(chatID, userID); err != nil {
		h.writeError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	client := service.SSE.RegisterChat(chatID, userID)
	defer service.SSE.UnregisterChat(chatID, client)

	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"chat_id\":%d}\n\n", chatID)
	c.Writer.Flush()

	// Heartbeat ticker (30s) to keep proxies from closing idle connections
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			_, _ = io.WriteString(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		case msg, ok := <-client.Ch:
			if !ok {
				return
			}
			_, _ = io.WriteString(c.Writer, msg)
			c.Writer.Flush()
		}
	}
}

// authSSE extracts the user ID from a JWT in ?token= or Authorization header.
func (h *ChatHandler) authSSE(c *gin.Context) (uint64, uint64, bool) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			tokenStr = parts[1]
		}
	}
	if tokenStr == "" {
		return 0, 0, false
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.secret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, false
	}
	sub, _ := claims["sub"].(string)
	uid, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return 0, 0, false
	}
	chatID, _ := strconv.ParseUint(c.Param("chatId"), 10, 64)
	return uid, chatID, true
}

func (h *ChatHandler) writeError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(appErr.Code, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}
