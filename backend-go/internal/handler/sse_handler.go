package handler

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/service"
)

type SSEHandler struct{}

func NewSSEHandler() *SSEHandler { return &SSEHandler{} }

// Connect handles GET /api/v1/sse — Server-Sent Events stream.
func (h *SSEHandler) Connect(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists { c.JSON(401, gin.H{"message":"Authentication required"}); return }
	u := user.(*model.User)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	client := service.SSE.Register(u.ID)
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
