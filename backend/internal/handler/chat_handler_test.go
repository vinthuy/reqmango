package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChatHandler_Stream_RejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &ChatHandler{secret: "test-secret"}
	r.GET("/chats/:chatId/stream", h.Stream)

	req := httptest.NewRequest(http.MethodGet, "/chats/1/stream", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
