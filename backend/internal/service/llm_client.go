package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ==================== LLM Types ====================

// Message represents a chat message.
type Message struct {
	Role    string     `json:"role"` // "user" | "assistant" | "system" | "tool"
	Content string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Tool defines a function the LLM can call.
type Tool struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	InputSchema *ToolSchema  `json:"input_schema"`
}

// ToolSchema is the JSON Schema for tool parameters.
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]SchemaProp  `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// SchemaProp describes a single parameter.
type SchemaProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall represents a tool call request from the LLM.
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
}

// StreamEvent is emitted during streaming responses.
type StreamEvent struct {
	Type       string      `json:"type"`       // "text" | "tool_call" | "tool_result" | "thinking" | "done" | "error"
	Content    string      `json:"content,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
	ThreadID   uint64      `json:"thread_id,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// ChatResponse is the final non-streaming response.
type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	StopReason string    `json:"stop_reason"`
}

// ==================== Anthropic API Types ====================

type anthropicRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []anthropicMsg  `json:"messages"`
	System      string          `json:"system,omitempty"`
	Tools       []anthropicTool `json:"tools,omitempty"`
	Stream      bool            `json:"stream"`
}

type anthropicMsg struct {
	Role    string           `json:"role"`
	Content json.RawMessage  `json:"content"`
}

type anthropicTool struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	InputSchema *ToolSchema  `json:"input_schema"`
}

type anthropicContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
}

type anthropicStreamEvent struct {
	Type          string                    `json:"type"`
	Delta         *anthropicDelta           `json:"delta,omitempty"`
	ContentBlock  *anthropicContentBlock    `json:"content_block,omitempty"`
	Message       *anthropicMsg             `json:"message,omitempty"`
}

type anthropicDelta struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	PartialJSON string        `json:"partial_json,omitempty"`
}

// ==================== LLMClient ====================

// LLMProvider identifies the LLM API protocol.
type LLMProvider string

const (
	ProviderAnthropic LLMProvider = "anthropic"
	ProviderOpenAI    LLMProvider = "openai"
	ProviderDeepSeek  LLMProvider = "deepseek" // OpenAI-compatible
)

// LLMClient wraps LLM API calls. Supports Anthropic and OpenAI-compatible protocols.
type LLMClient struct {
	apiKey   string
	model    string
	baseURL  string
	provider LLMProvider
	client   *http.Client
}

// NewLLMClient creates a new LLM client.
func NewLLMClient(apiKey, model, baseURL, provider string) *LLMClient {
	if model == "" {
		model = "deepseek-chat"
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	p := LLMProvider(provider)
	switch p {
	case ProviderAnthropic, ProviderOpenAI, ProviderDeepSeek:
	default:
		p = ProviderDeepSeek
	}
	return &LLMClient{
		apiKey:   apiKey,
		model:    model,
		baseURL:  baseURL,
		provider: p,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *LLMClient) isOpenAIProtocol() bool {
	return c.provider == ProviderOpenAI || c.provider == ProviderDeepSeek
}

// hasValidAPIKey returns false for obviously invalid keys.
func (c *LLMClient) hasValidAPIKey() bool {
	return c.apiKey != "" &&
		!strings.HasPrefix(c.apiKey, "sk-test") &&
		!strings.HasPrefix(c.apiKey, "sk-your-") &&
		!strings.Contains(c.apiKey, "infrastructure-testing") &&
		!strings.Contains(c.apiKey, "change-me")
}

// ChatSync sends a synchronous (non-streaming) chat request.
func (c *LLMClient) ChatSync(ctx context.Context, systemPrompt string, messages []Message, tools []Tool) (*ChatResponse, error) {
	if !c.hasValidAPIKey() {
		return nil, fmt.Errorf("AI 服务未配置：请在 工作空间设置 → AI 中配置有效的 API Key（当前 key 无效或为测试 key）")
	}
	req, err := c.buildRequest(systemPrompt, messages, tools, false)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if c.isOpenAIProtocol() {
		return c.parseOpenAIResponse(body)
	}
	return c.parseAnthropicResponse(body)
}

func (c *LLMClient) parseOpenAIResponse(body []byte) (*ChatResponse, error) {
	var result struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse OpenAI response: %w", err)
	}
	chatResp := &ChatResponse{}
	if len(result.Choices) > 0 {
		chatResp.Content = result.Choices[0].Message.Content
		chatResp.ToolCalls = result.Choices[0].Message.ToolCalls
		chatResp.StopReason = result.Choices[0].FinishReason
	}
	return chatResp, nil
}

func (c *LLMClient) parseAnthropicResponse(body []byte) (*ChatResponse, error) {
	var result struct {
		Content    []anthropicContentBlock `json:"content"`
		StopReason string                  `json:"stop_reason"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse Anthropic response: %w", err)
	}
	chatResp := &ChatResponse{StopReason: result.StopReason}
	for _, block := range result.Content {
		switch block.Type {
		case "text":
			chatResp.Content += block.Text
		case "tool_use":
			chatResp.ToolCalls = append(chatResp.ToolCalls, ToolCall{
				ID:   block.ID,
				Name: block.Name,
				Input: block.Input,
			})
		}
	}
	return chatResp, nil
}

// ChatStream sends a streaming chat request. Returns a channel of StreamEvent.
func (c *LLMClient) ChatStream(ctx context.Context, systemPrompt string, messages []Message, tools []Tool) (<-chan StreamEvent, error) {
	if !c.hasValidAPIKey() {
		return nil, fmt.Errorf("AI 服务未配置：请在 工作空间设置 → AI 中配置有效的 API Key（当前 key 无效或为测试 key）")
	}
	req, err := c.buildRequest(systemPrompt, messages, tools, true)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan StreamEvent, 64)
	if c.isOpenAIProtocol() {
		go c.readOpenAISSE(resp.Body, ch)
	} else {
		go c.readAnthropicSSE(resp.Body, ch)
	}
	return ch, nil
}

// Complete sends a simple prompt (no tools, no streaming).
func (c *LLMClient) Complete(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	resp, err := c.ChatSync(ctx, systemPrompt, []Message{{Role: "user", Content: userMessage}}, nil)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ==================== Internal ====================

func (c *LLMClient) buildRequest(systemPrompt string, messages []Message, tools []Tool, stream bool) (*http.Request, error) {
	if c.isOpenAIProtocol() {
		return c.buildOpenAIRequest(systemPrompt, messages, tools, stream)
	}
	return c.buildAnthropicRequest(systemPrompt, messages, tools, stream)
}

func (c *LLMClient) buildOpenAIRequest(systemPrompt string, messages []Message, tools []Tool, stream bool) (*http.Request, error) {
	openaiMsgs := make([]map[string]interface{}, 0, len(messages)+1)
	if systemPrompt != "" {
		openaiMsgs = append(openaiMsgs, map[string]interface{}{"role": "system", "content": systemPrompt})
	}
	for _, m := range messages {
		msg := map[string]interface{}{"role": m.Role, "content": m.Content}
		openaiMsgs = append(openaiMsgs, msg)
	}

	body := map[string]interface{}{
		"model":       c.model,
		"max_tokens":  4096,
		"messages":    openaiMsgs,
		"stream":      stream,
	}
	if len(tools) > 0 {
		openaiTools := make([]map[string]interface{}, len(tools))
		for i, t := range tools {
			openaiTools[i] = map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        t.Name,
					"description": t.Description,
					"parameters":  t.InputSchema,
				},
			}
		}
		body["tools"] = openaiTools
	}

	bodyBytes, _ := json.Marshal(body)
	url := strings.TrimRight(c.baseURL, "/") + "/chat/completions"
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	return httpReq, nil
}

func (c *LLMClient) buildAnthropicRequest(systemPrompt string, messages []Message, tools []Tool, stream bool) (*http.Request, error) {
	anthropicMsgs := make([]anthropicMsg, 0, len(messages))
	for _, m := range messages {
		content, _ := json.Marshal([]anthropicContentBlock{{Type: "text", Text: m.Content}})
		anthropicMsgs = append(anthropicMsgs, anthropicMsg{Role: m.Role, Content: content})
	}
	anthropicTools := make([]anthropicTool, len(tools))
	for i, t := range tools {
		anthropicTools[i] = anthropicTool{Name: t.Name, Description: t.Description, InputSchema: t.InputSchema}
	}

	body := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages:  anthropicMsgs,
		System:    systemPrompt,
		Tools:     anthropicTools,
		Stream:    stream,
	}
	bodyBytes, _ := json.Marshal(body)
	httpReq, _ := http.NewRequest("POST", c.baseURL+"/messages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	return httpReq, nil
}

func (c *LLMClient) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return c.client.Do(req)
}

// --- OpenAI-compatible SSE ---

func (c *LLMClient) readOpenAISSE(body io.ReadCloser, ch chan<- StreamEvent) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentToolID, currentToolName string
	var currentToolArgs strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- StreamEvent{Type: "done"}
			return
		}

		var evt struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						Index    int             `json:"index"`
						ID       string          `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		if len(evt.Choices) == 0 {
			continue
		}
		delta := evt.Choices[0].Delta

		// Text content
		if delta.Content != "" {
			ch <- StreamEvent{Type: "text", Content: delta.Content}
		}

		// Tool calls
		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				if tc.ID != "" {
					// Flush previous tool call if any
					if currentToolID != "" {
						rawInput := currentToolArgs.String()
						var input json.RawMessage
						if json.Valid([]byte(rawInput)) {
							input = json.RawMessage(rawInput)
						} else {
							input = json.RawMessage("{}")
						}
						ch <- StreamEvent{Type: "tool_call", ToolCall: &ToolCall{ID: currentToolID, Name: currentToolName, Input: input}}
					}
					currentToolID = tc.ID
					currentToolName = tc.Function.Name
					currentToolArgs.Reset()
				}
				if tc.Function.Arguments != "" {
					currentToolArgs.WriteString(tc.Function.Arguments)
				}
			}
		}

		// Finish — emit pending tool call
		if evt.Choices[0].FinishReason != "" {
			if currentToolID != "" {
				rawInput := currentToolArgs.String()
				var input json.RawMessage
				if json.Valid([]byte(rawInput)) {
					input = json.RawMessage(rawInput)
				} else {
					input = json.RawMessage("{}")
				}
				ch <- StreamEvent{Type: "tool_call", ToolCall: &ToolCall{ID: currentToolID, Name: currentToolName, Input: input}}
				currentToolID = ""
			}
			if evt.Choices[0].FinishReason == "stop" || evt.Choices[0].FinishReason == "tool_calls" {
				ch <- StreamEvent{Type: "done"}
			}
		}
	}
}

// --- Anthropic SSE ---

func (c *LLMClient) readAnthropicSSE(body io.ReadCloser, ch chan<- StreamEvent) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentToolID, currentToolName string
	var currentToolInput strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var evt anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}

		switch evt.Type {
		case "content_block_start":
			if evt.ContentBlock != nil && evt.ContentBlock.Type == "tool_use" {
				currentToolID = evt.ContentBlock.ID
				currentToolName = evt.ContentBlock.Name
				currentToolInput.Reset()
			}
		case "content_block_delta":
			if evt.Delta != nil {
				switch evt.Delta.Type {
				case "text_delta":
					ch <- StreamEvent{Type: "text", Content: evt.Delta.Text}
				case "input_json_delta":
					currentToolInput.WriteString(evt.Delta.PartialJSON)
				}
			}
		case "content_block_stop":
			if currentToolID != "" {
				var input json.RawMessage
				rawInput := currentToolInput.String()
				if json.Valid([]byte(rawInput)) {
					input = json.RawMessage(rawInput)
				} else {
					input = json.RawMessage("{}")
				}
				ch <- StreamEvent{Type: "tool_call", ToolCall: &ToolCall{ID: currentToolID, Name: currentToolName, Input: input}}
				currentToolID = ""
				currentToolName = ""
				currentToolInput.Reset()
			}
		case "message_stop":
			ch <- StreamEvent{Type: "done"}
		case "error":
			errMsg := "unknown error"
			if evt.Delta != nil {
				errMsg = evt.Delta.Text
			}
			ch <- StreamEvent{Type: "error", Error: errMsg}
		}
	}
}
