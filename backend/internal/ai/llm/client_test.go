package llm

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestToolSchema_JSONSerialization(t *testing.T) {
	schema := ToolSchema{
		Type: "object",
		Properties: map[string]SchemaProp{
			"query": {Type: "string", Description: "search query"},
			"limit": {Type: "integer", Description: "max results"},
		},
		Required: []string{"query"},
	}
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}
	if result["type"] != "object" {
		t.Errorf("type = %v, want 'object'", result["type"])
	}
	if result["required"] == nil {
		t.Error("required field should be present")
	}
}

func TestMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "simple user message",
			msg:  Message{Role: "user", Content: "Hello"},
		},
		{
			name: "assistant with tool calls",
			msg: Message{
				Role:    "assistant",
				Content: "Let me search...",
				ToolCalls: []ToolCall{
					{ID: "call_1", Name: "search_issues", Input: json.RawMessage(`{"query":"bug"}`)},
				},
			},
		},
		{
			name: "tool result",
			msg: Message{
				Role:       "tool",
				Content:    `[{"id":1,"title":"Bug A"}]`,
				ToolCallID: "call_1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Role != tt.msg.Role {
				t.Errorf("Role = %q, want %q", decoded.Role, tt.msg.Role)
			}
			if decoded.Content != tt.msg.Content {
				t.Errorf("Content = %q, want %q", decoded.Content, tt.msg.Content)
			}
		})
	}
}

func TestStreamEvent_Types(t *testing.T) {
	eventTypes := []string{"text", "tool_call", "tool_result", "thinking", "done", "error"}
	for _, typ := range eventTypes {
		evt := StreamEvent{Type: typ, Content: "test content"}
		data, err := json.Marshal(evt)
		if err != nil {
			t.Errorf("failed to marshal %s event: %v", typ, err)
		}
		var decoded StreamEvent
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("failed to unmarshal %s event: %v", typ, err)
		}
		if decoded.Type != typ {
			t.Errorf("Type = %q, want %q", decoded.Type, typ)
		}
	}
}

func TestChatResponse_JSONSerialization(t *testing.T) {
	resp := ChatResponse{
		Content:    "I found 3 issues.",
		StopReason: "end_turn",
		ToolCalls:  []ToolCall{{ID: "tc1", Name: "search", Input: json.RawMessage(`{}`)}},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ChatResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Content != resp.Content {
		t.Errorf("Content = %q, want %q", decoded.Content, resp.Content)
	}
	if decoded.StopReason != resp.StopReason {
		t.Errorf("StopReason = %q, want %q", decoded.StopReason, resp.StopReason)
	}
}

func TestTool_ToAnthropicFormat(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: &ToolSchema{
			Type: "object",
			Properties: map[string]SchemaProp{
				"param1": {Type: "string", Description: "A parameter"},
			},
		},
	}
	// Verify the tool is structured correctly for API consumption
	if tool.Name != "test_tool" {
		t.Errorf("Name = %q, want 'test_tool'", tool.Name)
	}
	if tool.InputSchema == nil {
		t.Error("InputSchema should not be nil")
	}
	if tool.InputSchema.Properties["param1"].Type != "string" {
		t.Errorf("param1 type = %q, want 'string'", tool.InputSchema.Properties["param1"].Type)
	}
}

func TestNewLLMClient_Defaults(t *testing.T) {
	client := NewLLMClient("test-api-key", "", "", "deepseek")
	if client == nil {
		t.Fatal("NewLLMClient should not return nil")
	}
}

func TestNewLLMClient_WithAllParams(t *testing.T) {
	client := NewLLMClient("key-123", "gpt-4", "https://custom.api.com/v1", "openai")
	if client == nil {
		t.Fatal("NewLLMClient should not return nil")
	}
}

func TestNewLLMClient_DefaultModel(t *testing.T) {
	client := NewLLMClient("key", "", "", "deepseek")
	if client == nil {
		t.Fatal("NewLLMClient should not return nil")
	}
}

func TestNewLLMClient_AnthropicProvider(t *testing.T) {
	client := NewLLMClient("key", "claude-sonnet-4-6", "https://api.anthropic.com", "anthropic")
	if client == nil {
		t.Fatal("NewLLMClient should not return nil")
	}
}

func TestNewLLMClient_XiaomiProvider(t *testing.T) {
	client := NewLLMClient("test-key", "mimo-v2.5", "", "xiaomi")
	if client == nil {
		t.Fatal("NewLLMClient should not return nil for xiaomi provider")
	}
	if client.provider != ProviderXiaomi {
		t.Errorf("provider = %q, want %q", client.provider, ProviderXiaomi)
	}
}

func TestBuildAnthropicRequest_XiaomiBearerAuth(t *testing.T) {
	client := NewLLMClient("test-key", "mimo-v2.5", "https://api.xiaomimimo.com/anthropic/v1", "xiaomi")
	req, err := client.buildAnthropicRequest("system prompt", []Message{
		{Role: "user", Content: "hello"},
	}, nil, false)
	if err != nil {
		t.Fatalf("buildAnthropicRequest: %v", err)
	}
	// Xiaomi provider should set Bearer auth in addition to x-api-key.
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Errorf("Authorization header = %q, want Bearer prefix", auth)
	}
	if req.Header.Get("x-api-key") != "test-key" {
		t.Errorf("x-api-key = %q, want %q", req.Header.Get("x-api-key"), "test-key")
	}
}

func TestParseOpenAIResponse_ToolCalls(t *testing.T) {
	client := NewLLMClient("k", "deepseek-chat", "", "deepseek")
	body := `{"choices":[{"finish_reason":"tool_calls","message":{"content":"先查一下",
		"tool_calls":[{"id":"call_1","type":"function","function":{"name":"get_issue","arguments":"{\"issue_id\":6444}"}}]}}]}`
	resp, err := client.parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("tool calls = %d, want 1", len(resp.ToolCalls))
	}
	tc := resp.ToolCalls[0]
	if tc.ID != "call_1" || tc.Name != "get_issue" || string(tc.Input) != `{"issue_id":6444}` {
		t.Errorf("tool call = %+v (input %s)", tc, tc.Input)
	}
	if resp.Content != "先查一下" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestParseOpenAIResponse_EmptyArgumentsBecomeObject(t *testing.T) {
	client := NewLLMClient("k", "deepseek-chat", "", "deepseek")
	body := `{"choices":[{"message":{"tool_calls":[{"id":"c","type":"function","function":{"name":"list_states","arguments":""}}]}}]}`
	resp, err := client.parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if string(resp.ToolCalls[0].Input) != "{}" {
		t.Errorf("input = %q, want {}", resp.ToolCalls[0].Input)
	}
}

func TestBuildAnthropicRequest_ToolRoundTrip(t *testing.T) {
	client := NewLLMClient("k", "claude-sonnet-4-6", "https://api.anthropic.com", "anthropic")
	req, err := client.buildAnthropicRequest("sys", []Message{
		{Role: "user", Content: "分诊"},
		{Role: "assistant", Content: "", ToolCalls: []ToolCall{
			{ID: "tu_1", Name: "get_issue", Input: json.RawMessage(`{"issue_id":1}`)},
			{ID: "tu_2", Name: "list_states", Input: json.RawMessage(`{}`)},
		}},
		{Role: "tool", Content: `{"name":"x"}`, ToolCallID: "tu_1"},
		{Role: "tool", Content: `{"error":"boom"}`, ToolCallID: "tu_2"},
	}, nil, false)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var body struct {
		Messages []struct {
			Role    string `json:"role"`
			Content []struct {
				Type      string          `json:"type"`
				Text      string          `json:"text"`
				ID        string          `json:"id"`
				Name      string          `json:"name"`
				Input     json.RawMessage `json:"input"`
				ToolUseID string          `json:"tool_use_id"`
				Content   string          `json:"content"`
			} `json:"content"`
		} `json:"messages"`
	}
	raw, _ := io.ReadAll(req.Body)
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Messages) != 3 {
		t.Fatalf("messages = %d, want 3 (user, assistant tool_use, merged tool_result): %s", len(body.Messages), raw)
	}
	asst := body.Messages[1]
	if asst.Role != "assistant" || len(asst.Content) != 2 || asst.Content[0].Type != "tool_use" || asst.Content[0].Name != "get_issue" {
		t.Errorf("assistant message = %+v", asst)
	}
	results := body.Messages[2]
	if results.Role != "user" || len(results.Content) != 2 {
		t.Fatalf("tool results message = %+v", results)
	}
	if results.Content[0].Type != "tool_result" || results.Content[0].ToolUseID != "tu_1" || results.Content[1].ToolUseID != "tu_2" {
		t.Errorf("tool results = %+v", results.Content)
	}
}

func TestBuildAnthropicRequest_NonXiaomi_NoBearerAuth(t *testing.T) {
	client := NewLLMClient("test-key", "claude-sonnet-4-6", "https://api.anthropic.com", "anthropic")
	req, err := client.buildAnthropicRequest("system prompt", []Message{
		{Role: "user", Content: "hello"},
	}, nil, false)
	if err != nil {
		t.Fatalf("buildAnthropicRequest: %v", err)
	}
	// Non-Xiaomi providers should NOT set Bearer auth.
	if req.Header.Get("Authorization") != "" {
		t.Errorf("Authorization header should be empty for anthropic provider, got %q", req.Header.Get("Authorization"))
	}
}
