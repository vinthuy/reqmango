package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAISearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/5/ai/search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["query"] != "auth bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"rql": `search "auth"`, "explanation": "matched", "issues": []map[string]any{{"id": 1}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	res, err := c.AISearch(context.Background(), 5, "auth bugs")
	if err != nil || res.RQL == "" || len(res.Issues) != 1 {
		t.Fatalf("AISearch: %v %+v", err, res)
	}
}

func TestAIChat_AggregatesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// backend ai/chat SSE: data-only lines, no event: field
		chunks := []map[string]any{
			{"type": "text", "content": "Hel"},
			{"type": "text", "content": "lo"},
			{"type": "tool_call", "tool_call": map[string]any{"name": "get_issue", "arguments": `{"id":1}`}},
			{"type": "done", "thread_id": 99},
		}
		for _, ch := range chunks {
			b, _ := json.Marshal(ch)
			fmt.Fprintf(w, "data: %s\n\n", b)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	reply, err := c.AIChat(context.Background(), 5, "hi", nil)
	if err != nil {
		t.Fatalf("AIChat: %v", err)
	}
	if reply.Text != "Hello" {
		t.Fatalf("expected aggregated text %q", reply.Text)
	}
	if reply.ThreadID != 99 || len(reply.ToolCalls) != 1 {
		t.Fatalf("unexpected reply %+v", reply)
	}
}

func TestAIChat_StreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":\"llm down\"}\n\n")
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	_, err := c.AIChat(context.Background(), 5, "hi", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	// The stub handler writes no explicit status (implicit 200): the error comes
	// from the SSE error event, which AIChat maps to *APIError with status 502.
	if apiErr := AsAPIError(err); apiErr == nil || apiErr.StatusCode != 502 {
		t.Fatalf("expected APIError with status 502, got %v", err)
	}
}
