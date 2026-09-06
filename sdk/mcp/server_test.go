package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mark3labs/mcp-go/server/servertest"

	"github.com/reqmango/tools/client"
)

// mcpPost sends one JSON-RPC message over stateful streamable HTTP and
// returns the decoded response body plus the session id ("" on first call).
func mcpPost(t *testing.T, baseURL, sessionID string, msg map[string]any) (map[string]any, string) {
	t.Helper()
	body, _ := json.Marshal(msg)
	req, _ := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set(server.HeaderKeySessionID, sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST status %d: %s", resp.StatusCode, data)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out, resp.Header.Get(server.HeaderKeySessionID)
}

// newTestMCPServer starts a stateful streamable HTTP test server, runs the
// initialize handshake and returns the base URL + session id.
func newTestMCPServer(t *testing.T, cli *client.Client) (*httptest.Server, string) {
	t.Helper()
	httpSrv := servertest.NewTestStreamableHTTPServer(New(cli), server.WithStateful(true))
	t.Cleanup(httpSrv.Close)
	_, sessionID := mcpPost(t, httpSrv.URL, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": mcp.LATEST_PROTOCOL_VERSION,
			"clientInfo":      map[string]any{"name": "reqmango-test", "version": "1.0.0"},
			"capabilities":    map[string]any{},
		},
	})
	if sessionID == "" {
		t.Fatal("initialize did not return a session id")
	}
	return httpSrv, sessionID
}

// callTool invokes tools/call and returns (text contents, isError).
func callTool(t *testing.T, baseURL, sessionID, name string, args map[string]any) ([]string, bool) {
	t.Helper()
	resp, _ := mcpPost(t, baseURL, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params":  map[string]any{"name": name, "arguments": args},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call %s: no result in %v", name, resp)
	}
	isError, _ := result["isError"].(bool)
	var texts []string
	if content, ok := result["content"].([]any); ok {
		for _, c := range content {
			if cm, ok := c.(map[string]any); ok {
				if s, ok := cm["text"].(string); ok {
					texts = append(texts, s)
				}
			}
		}
	}
	return texts, isError
}

func listToolNames(t *testing.T, baseURL, sessionID string) []string {
	t.Helper()
	resp, _ := mcpPost(t, baseURL, sessionID, map[string]any{
		"jsonrpc": "2.0", "id": 3, "method": "tools/list", "params": map[string]any{},
	})
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/list: no result in %v", resp)
	}
	var names []string
	for _, tool := range result["tools"].([]any) {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	return names
}

func TestNew_Initialize(t *testing.T) {
	httpSrv, sessionID := newTestMCPServer(t, client.New("", "reqmango_pat_test"))
	_ = httpSrv

	resp, _ := mcpPost(t, httpSrv.URL, sessionID, map[string]any{
		"jsonrpc": "2.0", "id": 4, "method": "ping", "params": map[string]any{},
	})
	if _, ok := resp["result"]; !ok {
		t.Fatalf("ping failed: %v", resp)
	}
}

func TestNew_ToolsListCoreRegistered(t *testing.T) {
	httpSrv, sessionID := newTestMCPServer(t, client.New("", "reqmango_pat_test"))
	names := listToolNames(t, httpSrv.URL, sessionID)
	// >=（不是 ==）：Task 12 还会注册 5 个 AI 工具，总数变 24。
	if len(names) < 19 {
		t.Fatalf("expected at least 19 core tools registered, got %v", names)
	}
}

func TestBearerAuth(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := BearerAuth("reqmango_pat_secret", inner)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without header, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_wrong")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer reqmango_pat_secret")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 with correct token, got %d", rr.Code)
	}
}
