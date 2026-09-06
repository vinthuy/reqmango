package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetIssueChat verifies path assembly and decoding of the issue chat shape.
// (The task brief lists no chat test file; these two cases follow the same
// pattern as the other brief-derived tests and cover the chat methods that
// the MCP/CLI phases depend on.)
func TestGetIssueChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues/11/chat" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": 8, "workspace_id": 2, "issue_id": 11, "type": "issue", "title": "Login broken",
			"messages": []map[string]any{{
				"id": 1, "chat_id": 8, "sender_type": "user", "content": "hi",
			}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	chat, err := c.GetIssueChat(context.Background(), 11)
	if err != nil || chat.ID != 8 || len(chat.Messages) != 1 || chat.Messages[0].Content != "hi" {
		t.Fatalf("GetIssueChat: %v %+v", err, chat)
	}
}

func TestSendMessage_PostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/api/v1/chats/8/messages" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["content"] != "hello there" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": 9, "chat_id": 8, "sender_type": "user", "content": "hello there",
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	msg, err := c.SendMessage(context.Background(), 8, "hello there")
	if err != nil || msg.ID != 9 || msg.Content != "hello there" {
		t.Fatalf("SendMessage: %v %+v", err, msg)
	}
}
