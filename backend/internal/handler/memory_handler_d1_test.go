package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/reqmango/backend/internal/ai/llm"
)

// TestSemanticSearch_NoLLMKey_Returns400 verifies D1: when no valid LLM API
// key is configured, the semantic-search endpoint returns 400 (not 500).
//
// With an empty API key the LLM client's SupportsEmbedding() is false, so
// SemanticSearchByText returns "semantic search requires an LLM client with
// embedding support" before touching the DB. The handler maps that to 400.
// A numeric :wsParam keeps parseWorkspaceID off the DB (nil db is safe).
func TestSemanticSearch_NoLLMKey_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	llmClient := llm.NewLLMClient("", "deepseek-chat", "https://api.deepseek.com/v1", "deepseek")
	h := NewMemoryHandler(nil, llmClient)

	r := gin.New()
	r.POST("/api/v1/workspaces/:wsParam/memories/semantic-search", h.SemanticSearch)

	body := bytes.NewBufferString(`{"query":"test","limit":5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/1/memories/semantic-search", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}
