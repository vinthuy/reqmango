package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// DatabaseToolExecutor executes tools using database lookups and HTTP calls.
type DatabaseToolExecutor struct {
	db           *gorm.DB
	functionRegistry *FunctionRegistry
}

func NewDatabaseToolExecutor(db *gorm.DB) *DatabaseToolExecutor {
	return &DatabaseToolExecutor{
		db:           db,
		functionRegistry: NewFunctionRegistry(),
	}
}

// SetFunctionRegistry sets a custom function registry.
func (e *DatabaseToolExecutor) SetFunctionRegistry(registry *FunctionRegistry) {
	e.functionRegistry = registry
}

// FunctionRegistry manages registered built-in functions.
type FunctionRegistry struct {
	functions map[string]func(ctx context.Context, input json.RawMessage) (interface{}, error)
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]func(ctx context.Context, input json.RawMessage) (interface{}, error)),
	}
}

// Register registers a function with the given name.
func (r *FunctionRegistry) Register(name string, fn func(ctx context.Context, input json.RawMessage) (interface{}, error)) {
	r.functions[name] = fn
}

// Execute executes a registered function.
func (r *FunctionRegistry) Execute(ctx context.Context, name string, input json.RawMessage) (interface{}, error) {
	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("function %s not registered", name)
	}
	return fn(ctx, input)
}

// Execute executes a tool by name and records the call log.
func (e *DatabaseToolExecutor) Execute(ctx context.Context, toolName string, input json.RawMessage) (interface{}, error) {
	// Find the tool by name
	var tool model.Tool
	if err := e.db.Where("name = ? AND status = ?", toolName, "active").First(&tool).Error; err != nil {
		return nil, fmt.Errorf("tool not found or inactive: %s", toolName)
	}

	startTime := time.Now()
	var output interface{}
	var err error

	// Execute based on tool type
	switch tool.ToolType {
	case "api":
		output, err = e.executeAPI(ctx, &tool, input)
	case "function":
		output, err = e.executeFunction(ctx, &tool, input)
	default:
		err = fmt.Errorf("unsupported tool type: %s", tool.ToolType)
	}

	durationMs := time.Since(startTime).Milliseconds()

	// Record tool call log
	e.recordToolCallLog(&tool, input, output, err, durationMs)

	return output, err
}

// recordToolCallLog records the tool call in the database.
func (e *DatabaseToolExecutor) recordToolCallLog(tool *model.Tool, input json.RawMessage, output interface{}, err error, durationMs int64) {
	log := model.ToolCallLog{
		ToolID:      tool.ID,
		InputParams: model.FromRawMessage(input),
		DurationMs:  durationMs,
	}

	// Set workspace ID if available
	if tool.WorkspaceID != nil {
		log.WorkspaceID = *tool.WorkspaceID
	}

	// Set output and status
	if err != nil {
		log.Status = "failed"
		errorMsg := err.Error()
		log.ErrorMessage = &errorMsg
	} else {
		log.Status = "success"
		if output != nil {
			outputJSON, jsonErr := json.Marshal(output)
			if jsonErr == nil {
				log.OutputResult = outputJSON
			}
		}
	}

	// Save log (fire and forget)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic recovery to prevent crash
			}
		}()
		if e.db == nil {
			return
		}
		e.db.Create(&log)
	}()
}

// executeAPI executes an API-type tool.
func (e *DatabaseToolExecutor) executeAPI(ctx context.Context, tool *model.Tool, input json.RawMessage) (interface{}, error) {
	if tool.Endpoint == nil || *tool.Endpoint == "" {
		return nil, fmt.Errorf("tool %s has no endpoint configured", tool.Name)
	}

	method := "POST"
	if tool.Method != nil {
		method = *tool.Method
	}

	client := &http.Client{
		Timeout: time.Duration(tool.Timeout) * time.Second,
	}

	req, err := http.NewRequest(method, *tool.Endpoint, bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add auth if configured
	if tool.AuthType != nil && tool.AuthConfig != nil {
		if err := e.addAuthHeaders(req, *tool.AuthType, tool.AuthConfig.ToRawMessage()); err != nil {
			return nil, fmt.Errorf("auth config error: %w", err)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// Return raw string if parsing fails
		return string(body), nil
	}

	return result, nil
}

// executeFunction executes a function-type tool using the function registry.
func (e *DatabaseToolExecutor) executeFunction(ctx context.Context, tool *model.Tool, input json.RawMessage) (interface{}, error) {
	if e.functionRegistry == nil {
		return nil, fmt.Errorf("function registry not configured")
	}

	// Try to execute using registry
	result, err := e.functionRegistry.Execute(ctx, tool.Name, input)
	if err != nil {
		// Return explicit error for unregistered functions to avoid masking issues in production
		return nil, fmt.Errorf("function %s not registered in function registry", tool.Name)
	}

	return map[string]interface{}{
		"tool":       tool.Name,
		"result":     result,
		"registered": true,
	}, nil
}

// addAuthHeaders adds authentication headers to the request.
func (e *DatabaseToolExecutor) addAuthHeaders(req *http.Request, authType string, authConfig json.RawMessage) error {
	var config map[string]string
	if err := json.Unmarshal(authConfig, &config); err != nil {
		return err
	}

	switch authType {
	case "api_key":
		if key, ok := config["api_key"]; ok {
			req.Header.Set("X-API-Key", key)
		}
	case "bearer":
		if token, ok := config["token"]; ok {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "oauth2":
		if token, ok := config["access_token"]; ok {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return nil
}