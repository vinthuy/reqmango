package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// sseEvent is one parsed SSE frame. Backend ai/chat emits data-only lines.
type sseEvent struct {
	Event string
	Data  []byte
}

// postSSE POSTs and streams SSE frames to fn until the response ends or ctx
// is done. Handles both `data:`-only streams (ai/chat) and `event:`-tagged
// streams (chats/:id/stream).
func (c *Client) postSSE(ctx context.Context, path string, query url.Values, body any, fn func(sseEvent) error) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// hcLong (Timeout: 0) — the 5-minute ctx is the sole governor for streams.
	resp, err := c.hcLong.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: map[string]any{}}
		var m map[string]any
		if json.Unmarshal(data, &m) == nil {
			apiErr.Body = m
			if msg, ok := m["message"].(string); ok {
				apiErr.Message = msg
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(data))
		}
		return apiErr
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var cur sseEvent
	flush := func() error {
		if len(cur.Data) > 0 {
			if err := fn(cur); err != nil {
				return err
			}
			cur = sseEvent{}
		}
		return nil
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		case strings.HasPrefix(line, "event:"):
			cur.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if cur.Data == nil {
				cur.Data = []byte(payload)
			} else {
				cur.Data = append(cur.Data, '\n')
				cur.Data = append(cur.Data, []byte(payload)...)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read SSE: %w", err)
	}
	return flush()
}

// AISearchRequest is the body for POST /projects/:id/ai/search.
type AISearchRequest struct {
	Query string `json:"query"`
}

// AISearchResponse carries the AI-generated RQL and matched issues.
type AISearchResponse struct {
	RQL         string                   `json:"rql"`
	Explanation string                   `json:"explanation"`
	Issues      []map[string]interface{} `json:"issues"`
}

// AISearch converts natural language into an issue search.
func (c *Client) AISearch(ctx context.Context, projectID uint64, query string) (*AISearchResponse, error) {
	var out AISearchResponse
	_, err := c.PostJSON(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/ai/search", nil,
		AISearchRequest{Query: query}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AIChatReply is the aggregated single-turn chat result.
type AIChatReply struct {
	Text      string
	ThreadID  uint64
	ToolCalls []string // "name(args)" summaries
}

// streamEvent is the backend llm.StreamEvent wire shape.
type streamEvent struct {
	Type       string `json:"type"` // text|tool_call|tool_result|thinking|done|error
	Content    string `json:"content"`
	ToolCall   *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"tool_call"`
	ThreadID uint64 `json:"thread_id"`
	Error    string `json:"error"`
}

// AIChat sends one message to the project AI chat and aggregates the SSE
// stream into a single reply. Long-running: uses a 5-minute context timeout.
func (c *Client) AIChat(ctx context.Context, projectID uint64, message string, threadID *uint64) (*AIChatReply, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	body := map[string]any{"message": message}
	if threadID != nil {
		body["thread_id"] = *threadID
	}

	reply := &AIChatReply{}
	err := c.postSSE(ctx, "/projects/"+strconv.FormatUint(projectID, 10)+"/ai/chat", nil, body,
		func(ev sseEvent) error {
			var se streamEvent
			if err := json.Unmarshal(ev.Data, &se); err != nil {
				return nil // ignore non-JSON keepalives
			}
			switch se.Type {
			case "text", "thinking":
				reply.Text += se.Content
			case "tool_call":
				if se.ToolCall != nil {
					reply.ToolCalls = append(reply.ToolCalls,
						se.ToolCall.Name+"("+se.ToolCall.Arguments+")")
				}
			case "error":
				return &APIError{StatusCode: 502, Message: se.Error, Body: map[string]any{"error": se.Error}}
			}
			if se.ThreadID != 0 {
				reply.ThreadID = se.ThreadID
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
