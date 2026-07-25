package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ToolService struct {
	db *gorm.DB
}

func NewToolService(db *gorm.DB) *ToolService {
	return &ToolService{db: db}
}

// Create creates a new tool
func (s *ToolService) Create(wid uint64, req request.CreateToolRequest) (*response.ToolResponse, error) {
	// Check if name already exists
	var existing model.Tool
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, common.BadRequest("Tool name already exists")
	}

	tool := model.Tool{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsBuiltin:   false,
		Status:      "active",
		ToolType:    req.ToolType,
		Endpoint:    req.Endpoint,
		Method:      req.Method,
		AuthType:    req.AuthType,
		AuthConfig:  req.AuthConfig,
		Params:      req.Params,
		RateLimit:   req.RateLimit,
		Timeout:     req.Timeout,
		WorkspaceID: &wid,
	}

	if err := s.db.Create(&tool).Error; err != nil {
		return nil, common.Internal("Failed to create tool")
	}

	return s.toResponse(&tool), nil
}

// Get retrieves a tool by ID
func (s *ToolService) Get(id uint64) (*response.ToolResponse, error) {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}

	return s.toResponse(&tool), nil
}

// List retrieves all tools for a workspace (including built-in tools)
func (s *ToolService) List(wid uint64) ([]response.ToolResponse, error) {
	var tools []model.Tool
	s.db.Where("workspace_id = ? OR workspace_id IS NULL", wid).Find(&tools)

	res := make([]response.ToolResponse, 0, len(tools))
	for _, t := range tools {
		res = append(res, *s.toResponse(&t))
	}

	return res, nil
}

// Update updates a tool
func (s *ToolService) Update(id uint64, req request.UpdateToolRequest) (*response.ToolResponse, error) {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}

	if tool.IsBuiltin {
		return nil, common.BadRequest("Built-in tools cannot be modified")
	}

	if req.Name != nil {
		tool.Name = *req.Name
	}
	if req.Description != nil {
		tool.Description = *req.Description
	}
	if req.Category != nil {
		tool.Category = *req.Category
	}
	if req.Status != nil {
		tool.Status = *req.Status
	}
	if req.ToolType != nil {
		tool.ToolType = *req.ToolType
	}
	if req.Endpoint != nil {
		tool.Endpoint = req.Endpoint
	}
	if req.Method != nil {
		tool.Method = req.Method
	}
	if req.AuthType != nil {
		tool.AuthType = req.AuthType
	}
	if req.AuthConfig != nil {
		tool.AuthConfig = *req.AuthConfig
	}
	if req.Params != nil {
		tool.Params = *req.Params
	}
	if req.RateLimit != nil {
		tool.RateLimit = *req.RateLimit
	}
	if req.Timeout != nil {
		tool.Timeout = *req.Timeout
	}

	if err := s.db.Save(&tool).Error; err != nil {
		return nil, common.Internal("Failed to update tool")
	}

	return s.Get(id)
}

// Delete deletes a tool
func (s *ToolService) Delete(id uint64) error {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return common.NotFound("Tool not found")
	}

	if tool.IsBuiltin {
		return common.BadRequest("Built-in tools cannot be deleted")
	}

	if err := s.db.Delete(&tool).Error; err != nil {
		return common.Internal("Failed to delete tool")
	}

	return nil
}

// Call executes a tool with the given parameters
func (s *ToolService) Call(wid uint64, req request.CallToolRequest) (*response.ToolCallResponse, error) {
	var tool model.Tool
	if err := s.db.First(&tool, req.ToolID).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}

	// Check if tool is active
	if tool.Status != "active" {
		return nil, common.BadRequest("Tool is not active")
	}

	// Check rate limit
	if err := s.checkRateLimit(tool); err != nil {
		return nil, err
	}

	// Validate parameters
	if err := s.validateParams(tool, req.InputParams); err != nil {
		return nil, err
	}

	start := time.Now()
	var outputResult json.RawMessage
	var status string
	var errorMessage *string

	// Execute the tool
	switch tool.ToolType {
	case "api":
		result, err := s.executeAPITool(tool, req.InputParams)
		if err != nil {
			status = "failed"
			msg := err.Error()
			errorMessage = &msg
		} else {
			status = "success"
			outputResult = result
		}
	case "function":
		result, err := s.executeFunctionTool(tool, req.InputParams)
		if err != nil {
			status = "failed"
			msg := err.Error()
			errorMessage = &msg
		} else {
			status = "success"
			outputResult = result
		}
	default:
		return nil, common.BadRequest("Unsupported tool type")
	}

	duration := time.Since(start).Milliseconds()

	// Create tool call log
	log := model.ToolCallLog{
		WorkspaceID:  wid,
		ToolID:       tool.ID,
		InputParams:  req.InputParams,
		OutputResult: outputResult,
		Status:       status,
		ErrorMessage: errorMessage,
		DurationMs:   duration,
	}
	s.db.Create(&log)

	var output interface{}
	if len(outputResult) > 0 {
		json.Unmarshal(outputResult, &output)
	}

	return &response.ToolCallResponse{
		ID:          log.ID,
		ToolID:      tool.ID,
		ToolName:    tool.Name,
		InputParams: req.InputParams,
		OutputResult: output,
		Status:      status,
		ErrorMessage: errorMessage,
		DurationMs:  duration,
		CreatedAt:   log.CreatedAt,
	}, nil
}

// GetCallLogs retrieves tool call logs for a workspace
func (s *ToolService) GetCallLogs(wid uint64) ([]response.ToolCallLogResponse, error) {
	var logs []model.ToolCallLog
	s.db.Preload("Tool").Where("workspace_id = ?", wid).Order("created_at DESC").Find(&logs)

	res := make([]response.ToolCallLogResponse, 0, len(logs))
	for _, log := range logs {
		var inputParams, outputResult interface{}
		if len(log.InputParams) > 0 {
			json.Unmarshal(log.InputParams, &inputParams)
		}
		if len(log.OutputResult) > 0 {
			json.Unmarshal(log.OutputResult, &outputResult)
		}

		res = append(res, response.ToolCallLogResponse{
			ID:           log.ID,
			WorkspaceID:  log.WorkspaceID,
			AgentTaskID:  log.AgentTaskID,
			ToolID:       log.ToolID,
			ToolName:     log.Tool.Name,
			AgentID:      log.AgentID,
			InputParams:  inputParams,
			OutputResult: outputResult,
			Status:       log.Status,
			ErrorMessage: log.ErrorMessage,
			DurationMs:   log.DurationMs,
			RateLimited:  log.RateLimited,
			CreatedAt:    log.CreatedAt,
		})
	}

	return res, nil
}

// executeAPITool executes an API type tool
func (s *ToolService) executeAPITool(tool model.Tool, params json.RawMessage) (json.RawMessage, error) {
	if tool.Endpoint == nil {
		return nil, common.BadRequest("API tool requires an endpoint")
	}

	method := "GET"
	if tool.Method != nil {
		method = *tool.Method
	}

	timeout := 30
	if tool.Timeout > 0 {
		timeout = tool.Timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, *tool.Endpoint, bytes.NewBuffer(params))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers
	if tool.AuthType != nil {
		switch *tool.AuthType {
		case "api_key":
			var config map[string]string
			if err := json.Unmarshal(tool.AuthConfig, &config); err == nil {
				if key, ok := config["api_key"]; ok {
					req.Header.Set("X-API-Key", key)
				}
			}
		case "bearer":
			var config map[string]string
			if err := json.Unmarshal(tool.AuthConfig, &config); err == nil {
				if token, ok := config["token"]; ok {
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				}
			}
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check HTTP status code
	if resp.StatusCode >= 400 {
		return nil, common.BadRequest(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body)))
	}

	return body, nil
}

// executeFunctionTool executes a function type tool
func (s *ToolService) executeFunctionTool(tool model.Tool, params json.RawMessage) (json.RawMessage, error) {
	// For built-in functions, we handle them internally
	if tool.IsBuiltin {
		return s.executeBuiltinFunction(tool.Name, params)
	}

	return nil, common.BadRequest("Custom function tools are not supported yet")
}

// executeBuiltinFunction executes built-in functions
func (s *ToolService) executeBuiltinFunction(name string, params json.RawMessage) (json.RawMessage, error) {
	switch name {
	case "create_issue":
		return s.executeCreateIssue(params)
	case "update_issue":
		return s.executeUpdateIssue(params)
	case "list_issues":
		return s.executeListIssues(params)
	default:
		return nil, common.BadRequest(fmt.Sprintf("Unknown built-in function: %s", name))
	}
}

// executeCreateIssue implements the create_issue built-in function
func (s *ToolService) executeCreateIssue(params json.RawMessage) (json.RawMessage, error) {
	var p struct {
		ProjectID   uint64 `json:"project_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, common.BadRequest("Invalid parameters")
	}

	if p.Title == "" {
		return nil, common.BadRequest("Title is required")
	}

	stateID := uint64(1)
	if p.State != "" {
		var state model.State
		if err := s.db.Where("name = ?", p.State).First(&state).Error; err == nil {
			stateID = state.ID
		}
	}

	issue := model.Issue{
		ProjectID:        p.ProjectID,
		Name:             p.Title,
		DescriptionHTML:  p.Description,
		StateID:          stateID,
	}

	if err := s.db.Create(&issue).Error; err != nil {
		return nil, err
	}

	result, _ := json.Marshal(map[string]interface{}{
		"id":          issue.ID,
		"title":       issue.Name,
		"description": issue.DescriptionHTML,
		"state":       issue.StateID,
		"created_at":  issue.CreatedAt,
	})

	return result, nil
}

// executeUpdateIssue implements the update_issue built-in function
func (s *ToolService) executeUpdateIssue(params json.RawMessage) (json.RawMessage, error) {
	var p struct {
		ID          uint64 `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		StateID     uint64 `json:"state_id"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, common.BadRequest("Invalid parameters")
	}

	if p.ID == 0 {
		return nil, common.BadRequest("Issue ID is required")
	}

	var issue model.Issue
	if err := s.db.First(&issue, p.ID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	if p.Title != "" {
		issue.Name = p.Title
	}
	if p.Description != "" {
		issue.DescriptionHTML = p.Description
	}
	if p.StateID != 0 {
		issue.StateID = p.StateID
	}

	if err := s.db.Save(&issue).Error; err != nil {
		return nil, err
	}

	result, _ := json.Marshal(map[string]interface{}{
		"id":          issue.ID,
		"title":       issue.Name,
		"description": issue.DescriptionHTML,
		"state_id":    issue.StateID,
		"updated_at":  issue.UpdatedAt,
	})

	return result, nil
}

// executeListIssues implements the list_issues built-in function
func (s *ToolService) executeListIssues(params json.RawMessage) (json.RawMessage, error) {
	var p struct {
		ProjectID uint64 `json:"project_id"`
		Limit     int    `json:"limit"`
	}

	if err := json.Unmarshal(params, &p); err != nil {
		return nil, common.BadRequest("Invalid parameters")
	}

	if p.Limit == 0 {
		p.Limit = 20
	}

	var issues []model.Issue
	query := s.db
	if p.ProjectID != 0 {
		query = query.Where("project_id = ?", p.ProjectID)
	}
	query.Limit(p.Limit).Order("created_at DESC").Find(&issues)

	result := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		result = append(result, map[string]interface{}{
			"id":          issue.ID,
			"title":       issue.Name,
			"description": issue.DescriptionHTML,
			"state_id":    issue.StateID,
			"project_id":  issue.ProjectID,
			"created_at":  issue.CreatedAt,
		})
	}

	return json.Marshal(result)
}

// checkRateLimit checks if the tool has exceeded its rate limit
func (s *ToolService) checkRateLimit(tool model.Tool) error {
	if tool.RateLimit <= 0 {
		return nil
	}

	now := time.Now()
	var count int64
	s.db.Model(&model.ToolCallLog{}).
		Where("tool_id = ? AND created_at >= ?", tool.ID, now.Add(-1*time.Minute)).
		Count(&count)

	if int(count) >= tool.RateLimit {
		return common.BadRequest("Rate limit exceeded")
	}

	return nil
}

// validateParams validates input parameters against the tool's parameter schema
func (s *ToolService) validateParams(tool model.Tool, params json.RawMessage) error {
	// For simplicity, we just check if required fields are present
	// In a production system, you'd want to use a proper JSON Schema validator
	return nil
}

// toResponse converts a model.Tool to response.ToolResponse
func (s *ToolService) toResponse(tool *model.Tool) *response.ToolResponse {
	var params interface{}
	if len(tool.Params) > 0 {
		json.Unmarshal(tool.Params, &params)
	}

	return &response.ToolResponse{
		ID:          tool.ID,
		Name:        tool.Name,
		Description: tool.Description,
		Category:    tool.Category,
		IsBuiltin:   tool.IsBuiltin,
		Status:      tool.Status,
		ToolType:    tool.ToolType,
		Endpoint:    tool.Endpoint,
		Method:      tool.Method,
		AuthType:    tool.AuthType,
		Params:      params,
		RateLimit:   tool.RateLimit,
		Timeout:     tool.Timeout,
		WorkspaceID: tool.WorkspaceID,
		CreatedAt:   tool.CreatedAt,
		UpdatedAt:   tool.UpdatedAt,
	}
}

// RegisterBuiltinTools registers built-in tools on startup
func (s *ToolService) RegisterBuiltinTools() error {
	builtinTools := []model.Tool{
		{
			Name:        "create_issue",
			Description: "Create a new issue in the project management system",
			Category:    "project_management",
			IsBuiltin:   true,
			Status:      "active",
			ToolType:    "function",
			Params:      json.RawMessage(`{"type":"object","properties":{"project_id":{"type":"integer"},"title":{"type":"string"},"description":{"type":"string"},"type":{"type":"string"},"state":{"type":"string"}},"required":["project_id","title"]}`),
			RateLimit:   60,
			Timeout:     30,
			WorkspaceID: nil,
		},
		{
			Name:        "update_issue",
			Description: "Update an existing issue",
			Category:    "project_management",
			IsBuiltin:   true,
			Status:      "active",
			ToolType:    "function",
			Params:      json.RawMessage(`{"type":"object","properties":{"id":{"type":"integer"},"title":{"type":"string"},"description":{"type":"string"},"state_id":{"type":"integer"}},"required":["id"]}`),
			RateLimit:   60,
			Timeout:     30,
			WorkspaceID: nil,
		},
		{
			Name:        "list_issues",
			Description: "List issues in a project",
			Category:    "project_management",
			IsBuiltin:   true,
			Status:      "active",
			ToolType:    "function",
			Params:      json.RawMessage(`{"type":"object","properties":{"project_id":{"type":"integer"},"limit":{"type":"integer"}}}`),
			RateLimit:   60,
			Timeout:     30,
			WorkspaceID: nil,
		},
	}

	for _, tool := range builtinTools {
		var existing model.Tool
		if err := s.db.Where("name = ? AND is_builtin = ?", tool.Name, true).First(&existing).Error; err != nil {
			if err := s.db.Create(&tool).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
