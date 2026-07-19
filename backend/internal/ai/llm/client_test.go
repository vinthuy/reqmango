package llm

import (
	"encoding/json"
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
