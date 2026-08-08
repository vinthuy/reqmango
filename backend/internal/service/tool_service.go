package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// derefUint64 returns the value of a *uint64, or 0 if nil.
func derefUint64(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

// ==================== ToolService ====================

type ToolService struct {
	db     *gorm.DB
	mcpSvc *MCPService
}

func NewToolService(db *gorm.DB) *ToolService {
	return &ToolService{db: db}
}

// SetMCPService sets the MCPService reference to avoid circular dependency.
func (s *ToolService) SetMCPService(mcpSvc *MCPService) {
	s.mcpSvc = mcpSvc
}

// ==================== T3: Dual-dimension sliding window rate limiter ====================

type rateLimitEntry struct {
	mu    sync.Mutex
	calls []time.Time
}

type rateLimiter struct {
	mu    sync.RWMutex
	store sync.Map // string key -> *rateLimitEntry
}

var globalRateLimiter = &rateLimiter{}

func (rl *rateLimiter) getOrCreateEntry(key string) *rateLimitEntry {
	if v, ok := rl.store.Load(key); ok {
		return v.(*rateLimitEntry)
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if v, ok := rl.store.Load(key); ok {
		return v.(*rateLimitEntry)
	}
	e := &rateLimitEntry{calls: make([]time.Time, 0, 16)}
	rl.store.Store(key, e)
	return e
}

// tryAcquire attempts to acquire a rate-limit token for both global and per-caller dimensions.
// callerID=0 means skip the per-caller check (anonymous).
func (rl *rateLimiter) tryAcquire(toolID, callerID uint64, perMinute int) bool {
	if perMinute <= 0 {
		return true
	}
	window := 60 * time.Second
	now := time.Now()
	cutoff := now.Add(-window)

	gKey := fmt.Sprintf("g:%d", toolID)
	cKey := ""
	if callerID > 0 {
		cKey = fmt.Sprintf("c:%d:%d", toolID, callerID)
	}
	gEntry := rl.getOrCreateEntry(gKey)
	var cEntry *rateLimitEntry
	if cKey != "" {
		cEntry = rl.getOrCreateEntry(cKey)
	}

	gEntry.mu.Lock()
	if cEntry != nil {
		cEntry.mu.Lock()
		defer cEntry.mu.Unlock()
	}
	defer gEntry.mu.Unlock()

	filter := func(src []time.Time) []time.Time {
		dst := src[:0]
		for _, t := range src {
			if t.After(cutoff) {
				dst = append(dst, t)
			}
		}
		return dst
	}
	gEntry.calls = filter(gEntry.calls)
	if cEntry != nil {
		cEntry.calls = filter(cEntry.calls)
	}

	callerLimit := int(math.Ceil(float64(perMinute) / 5.0))

	if len(gEntry.calls) >= perMinute {
		return false
	}
	if cEntry != nil && len(cEntry.calls) >= callerLimit {
		return false
	}

	gEntry.calls = append(gEntry.calls, now)
	if cEntry != nil {
		cEntry.calls = append(cEntry.calls, now)
	}
	return true
}

// checkRateLimit uses the global rate limiter instead of DB count.
func (s *ToolService) checkRateLimit(toolID, callerID uint64, perMinute int) bool {
	return globalRateLimiter.tryAcquire(toolID, callerID, perMinute)
}

// ==================== T2: Three-step permission checks ====================

func (s *ToolService) checkMemberOfWorkspace(wid, uid uint64) (*model.WorkspaceMember, error) {
	var m model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?",
		wid, uid, true).First(&m).Error; err != nil {
		return nil, common.Forbidden("Workspace member required")
	}
	return &m, nil
}

func (s *ToolService) checkPermissions(wid uint64, tool *model.Tool, req *request.CallToolRequest) (*model.WorkspaceMember, error) {
	// Step 1: Workspace membership
	member, err := s.checkMemberOfWorkspace(wid, req.CallerUserID)
	if err != nil {
		return nil, err
	}
	// Step 2: Dangerous tools require Admin role
	if tool.Category == "dangerous" && member.Role < common.RoleAdmin {
		return nil, common.Forbidden("Admin required for dangerous tools")
	}
	// Step 3: ToolPermission whitelist/blacklist
	if req.AgentTemplateID != nil {
		var tp model.ToolPermission
		// Try exact match first
		err := s.db.Where("tool_id = ? AND agent_template_id = ?", tool.ID, *req.AgentTemplateID).First(&tp).Error
		if err == nil && !tp.Allowed {
			return nil, common.Forbidden("Tool permission denied")
		}
		// Fallback to wildcard (agent_template_id IS NULL)
		err = s.db.Where("tool_id = ? AND agent_template_id IS NULL", tool.ID).First(&tp).Error
		if err == nil && !tp.Allowed {
			return nil, common.Forbidden("Tool permission denied")
		}
	}
	return member, nil
}

// ==================== CRUD ====================

func (s *ToolService) Create(wid uint64, req request.CreateToolRequest) (*response.ToolResponse, error) {
	tool := model.Tool{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		ToolType:    req.ToolType,
		Endpoint:    req.Endpoint,
		Method:      req.Method,
		AuthType:    req.AuthType,
		AuthConfig:  model.FromRawMessage(req.AuthConfig),
		Params:      model.FromRawMessage(req.Params),
		RateLimit:   req.RateLimit,
		Timeout:     req.Timeout,
		WorkspaceID: &wid,
		Status:      "active",
	}
	if tool.Category == "" {
		tool.Category = "general"
	}
	if tool.Timeout == 0 {
		tool.Timeout = 30
	}
	if err := s.db.Create(&tool).Error; err != nil {
		return nil, common.Internal("Failed to create tool")
	}
	return s.toResponse(&tool), nil
}

func (s *ToolService) Get(id uint64) (*response.ToolResponse, error) {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}
	return s.toResponse(&tool), nil
}

func (s *ToolService) List(wid uint64) ([]response.ToolResponse, error) {
	var tools []model.Tool
	if err := s.db.Where("workspace_id = ?", wid).Order("name ASC").Find(&tools).Error; err != nil {
		return nil, common.Internal("Failed to list tools")
	}
	result := make([]response.ToolResponse, len(tools))
	for i := range tools {
		result[i] = *s.toResponse(&tools[i])
	}
	return result, nil
}

func (s *ToolService) Update(id uint64, req request.UpdateToolRequest) (*response.ToolResponse, error) {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ToolType != nil {
		updates["tool_type"] = *req.ToolType
	}
	if req.Endpoint != nil {
		updates["endpoint"] = *req.Endpoint
	}
	if req.Method != nil {
		updates["method"] = *req.Method
	}
	if req.AuthType != nil {
		updates["auth_type"] = *req.AuthType
	}
	if req.AuthConfig != nil {
		updates["auth_config"] = *req.AuthConfig
	}
	if req.Params != nil {
		updates["params"] = *req.Params
	}
	if req.RateLimit != nil {
		updates["rate_limit"] = *req.RateLimit
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if len(updates) > 0 {
		if err := s.db.Model(&tool).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update tool")
		}
	}
	s.db.First(&tool, id)
	return s.toResponse(&tool), nil
}

func (s *ToolService) Delete(id uint64) error {
	var tool model.Tool
	if err := s.db.First(&tool, id).Error; err != nil {
		return common.NotFound("Tool not found")
	}
	if tool.IsBuiltin {
		return common.Forbidden("Built-in tools cannot be deleted")
	}
	if err := s.db.Delete(&tool).Error; err != nil {
		return common.Internal("Failed to delete tool")
	}
	return nil
}

// ==================== Existing builtin functions ====================

func (s *ToolService) executeCreateIssue(params json.RawMessage) (interface{}, error) {
	var input struct {
		WorkspaceID   uint64 `json:"workspace_id"`
		ProjectID     uint64 `json:"project_id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		Priority      string `json:"priority"`
		StateID       *uint64 `json:"state_id"`
		IssueTypeID   *uint64 `json:"issue_type_id"`
		ParentID      *uint64 `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.WorkspaceID == 0 || input.ProjectID == 0 || input.Name == "" {
		return nil, common.BadRequest("workspace_id, project_id, and name are required")
	}
	issue := model.Issue{
		Name:          input.Name,
		DescriptionHTML: input.Description,
		Priority:      input.Priority,
		ProjectID:     input.ProjectID,
		WorkspaceID:   input.WorkspaceID,
		StateID:       derefUint64(input.StateID),
		IssueTypeID:   input.IssueTypeID,
		ParentID:      input.ParentID,
	}
	if issue.Priority == "" {
		issue.Priority = "none"
	}
	if err := s.db.Create(&issue).Error; err != nil {
		return nil, common.Internal("Failed to create issue: " + err.Error())
	}
	return map[string]interface{}{"id": issue.ID, "name": issue.Name}, nil
}

func (s *ToolService) executeUpdateIssue(params json.RawMessage) (interface{}, error) {
	var input struct {
		IssueID     uint64  `json:"issue_id"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Priority    *string `json:"priority"`
		StateID     *uint64 `json:"state_id"`
		IssueTypeID *uint64 `json:"issue_type_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.IssueID == 0 {
		return nil, common.BadRequest("issue_id is required")
	}
	var issue model.Issue
	if err := s.db.First(&issue, input.IssueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Priority != nil {
		updates["priority"] = *input.Priority
	}
	if input.StateID != nil {
		updates["state_id"] = *input.StateID
	}
	if input.IssueTypeID != nil {
		updates["issue_type_id"] = *input.IssueTypeID
	}
	if len(updates) > 0 {
		s.db.Model(&issue).Updates(updates)
	}
	return map[string]interface{}{"id": issue.ID, "updated": true}, nil
}

func (s *ToolService) executeListIssues(params json.RawMessage) (interface{}, error) {
	var input struct {
		ProjectID   uint64 `json:"project_id"`
		StateID     *uint64 `json:"state_id"`
		Priority    string `json:"priority"`
		AssigneeID  *uint64 `json:"assignee_id"`
		Limit       int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.ProjectID == 0 {
		return nil, common.BadRequest("project_id is required")
	}
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 50
	}
	q := s.db.Where("project_id = ?", input.ProjectID)
	if input.StateID != nil {
		q = q.Where("state_id = ?", *input.StateID)
	}
	if input.Priority != "" {
		q = q.Where("priority = ?", input.Priority)
	}
	var issues []model.Issue
	q.Order("created_at DESC").Limit(input.Limit).Find(&issues)
	return issues, nil
}

// ==================== githubJSON helper ====================

func githubJSON(method, url string, body interface{}) (map[string]interface{}, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN not set")
	}
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API %d: %s", resp.StatusCode, string(raw))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return map[string]interface{}{"raw": string(raw)}, nil
	}
	return result, nil
}

// ==================== 4 New code review builtin functions ====================

func (s *ToolService) builtinGetPRDiff(params json.RawMessage) (interface{}, error) {
	var input struct {
		RepoOwner string `json:"repo_owner"`
		RepoName  string `json:"repo_name"`
		PRNumber  int    `json:"pr_number"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.RepoOwner == "" {
		return nil, common.BadRequest("repo_owner is required")
	}
	if input.RepoName == "" {
		return nil, common.BadRequest("repo_name is required")
	}
	if input.PRNumber == 0 {
		return nil, common.BadRequest("pr_number is required")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", input.RepoOwner, input.RepoName, input.PRNumber)
	return githubJSON("GET", url, nil)
}

func (s *ToolService) builtinAddReviewComment(params json.RawMessage) (interface{}, error) {
	var input struct {
		RepoOwner string `json:"repo_owner"`
		RepoName  string `json:"repo_name"`
		PRNumber  int    `json:"pr_number"`
		Body      string `json:"body"`
		Path      string `json:"path"`
		Position  int    `json:"position"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.RepoOwner == "" || input.RepoName == "" || input.PRNumber == 0 {
		return nil, common.BadRequest("repo_owner, repo_name, and pr_number are required")
	}
	if input.Body == "" {
		return nil, common.BadRequest("body is required")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments", input.RepoOwner, input.RepoName, input.PRNumber)
	payload := map[string]interface{}{
		"body":     input.Body,
		"path":     input.Path,
		"position": input.Position,
	}
	return githubJSON("POST", url, payload)
}

func (s *ToolService) builtinListPRCommits(params json.RawMessage) (interface{}, error) {
	var input struct {
		RepoOwner string `json:"repo_owner"`
		RepoName  string `json:"repo_name"`
		PRNumber  int    `json:"pr_number"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.RepoOwner == "" || input.RepoName == "" || input.PRNumber == 0 {
		return nil, common.BadRequest("repo_owner, repo_name, and pr_number are required")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/commits", input.RepoOwner, input.RepoName, input.PRNumber)
	return githubJSON("GET", url, nil)
}

func (s *ToolService) builtinCreatePRReview(params json.RawMessage) (interface{}, error) {
	var input struct {
		RepoOwner string `json:"repo_owner"`
		RepoName  string `json:"repo_name"`
		PRNumber  int    `json:"pr_number"`
		Event     string `json:"event"`
		Body      string `json:"body"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, common.BadRequest("Invalid params")
	}
	if input.RepoOwner == "" || input.RepoName == "" || input.PRNumber == 0 {
		return nil, common.BadRequest("repo_owner, repo_name, and pr_number are required")
	}
	validEvents := map[string]bool{
		"APPROVE": true, "REQUEST_CHANGES": true, "COMMENT": true,
	}
	if !validEvents[input.Event] {
		return nil, common.BadRequest("event must be one of: APPROVE, REQUEST_CHANGES, COMMENT")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews", input.RepoOwner, input.RepoName, input.PRNumber)
	payload := map[string]interface{}{
		"event": input.Event,
		"body":  input.Body,
	}
	return githubJSON("POST", url, payload)
}

// ==================== Call: permission check → rate limit → validate → execute → audit → SSE ====================

func (s *ToolService) Call(wid uint64, req request.CallToolRequest) (*response.ToolCallResponse, error) {
	// Look up tool
	var tool model.Tool
	if err := s.db.First(&tool, req.ToolID).Error; err != nil {
		return nil, common.NotFound("Tool not found")
	}
	if tool.Status != "active" {
		return nil, common.BadRequest("Tool is not active")
	}

	// Step 1-3: Permission checks
	if _, err := s.checkPermissions(wid, &tool, &req); err != nil {
		return nil, err
	}

	// Rate limit
	if tool.RateLimit > 0 {
		if !s.checkRateLimit(tool.ID, req.CallerUserID, tool.RateLimit) {
			return nil, common.BadRequest("Rate limit exceeded")
		}
	}

	// Validate params
	if err := s.validateParams(&tool, req.InputParams); err != nil {
		return nil, err
	}

	// Execute
	start := time.Now()
	var result interface{}
	var status = "success"
	var errorMessage *string

	switch tool.ToolType {
	case "api":
		r, err := s.executeAPITool(&tool, req.InputParams)
		if err != nil {
			status = "failed"
			msg := err.Error()
			errorMessage = &msg
		} else {
			result = r
		}
	case "function":
		r, err := s.executeFunctionTool(&tool, req.InputParams)
		if err != nil {
			status = "failed"
			msg := err.Error()
			errorMessage = &msg
		} else {
			result = r
		}
	case "mcp":
		if tool.MCPConfigID == nil {
			status = "failed"
			msg := "MCP tool missing mcp_config_id"
			errorMessage = &msg
		} else if s.mcpSvc == nil {
			status = "failed"
			msg := "MCP service not configured"
			errorMessage = &msg
		} else {
			var args map[string]interface{}
			if len(req.InputParams) > 0 {
				_ = json.Unmarshal(req.InputParams, &args)
			}
			r, err := s.mcpSvc.CallTool(wid, *tool.MCPConfigID, tool.Name, args)
			if err != nil {
				status = "failed"
				msg := err.Error()
				errorMessage = &msg
			} else {
				result = r
			}
		}
	default:
		return nil, common.BadRequest("Unsupported tool type")
	}

	duration := time.Since(start).Milliseconds()

	// Build output
	var outputResult json.RawMessage
	if result != nil {
		outputResult, _ = json.Marshal(result)
	}

	// Audit log
	log := model.ToolCallLog{
		WorkspaceID:  wid,
		ToolID:       tool.ID,
		AgentTaskID:  req.AgentTaskID,
		InputParams:  model.FromRawMessage(req.InputParams),
		OutputResult: model.FromRawMessage(outputResult),
		Status:       status,
		ErrorMessage: errorMessage,
		DurationMs:   duration,
	}
	s.db.Create(&log)

	// SSE broadcast
	evt := "tool_call.completed"
	payload := map[string]interface{}{
		"id":             log.ID,
		"workspace_id":   wid,
		"tool_id":        tool.ID,
		"tool_name":      tool.Name,
		"status":         status,
		"duration_ms":    duration,
		"caller_user_id": req.CallerUserID,
	}
	if req.AgentTaskID != nil {
		payload["agent_task_id"] = *req.AgentTaskID
	}
	SSE.BroadcastEvent(evt, payload)

	if status == "failed" {
		failPayload := map[string]interface{}{
			"id":      log.ID,
			"tool_id": tool.ID,
			"error":   errorMessage,
		}
		SSE.BroadcastEvent("tool_call.failed", failPayload)
	}

	return &response.ToolCallResponse{
		ID:           log.ID,
		ToolID:       tool.ID,
		ToolName:     tool.Name,
		InputParams:  req.InputParams,
		OutputResult: result,
		Status:       status,
		ErrorMessage: errorMessage,
		DurationMs:   duration,
		CreatedAt:    log.CreatedAt,
	}, nil
}

// ==================== Execute helpers ====================

func (s *ToolService) executeAPITool(tool *model.Tool, params json.RawMessage) (interface{}, error) {
	if tool.Endpoint == nil || *tool.Endpoint == "" {
		return nil, fmt.Errorf("tool has no endpoint configured")
	}
	method := "POST"
	if tool.Method != nil && *tool.Method != "" {
		method = strings.ToUpper(*tool.Method)
	}

	req, err := http.NewRequest(method, *tool.Endpoint, bytes.NewReader(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if tool.AuthType != nil {
		switch *tool.AuthType {
		case "bearer", "api_key":
			var authCfg map[string]string
			if err := json.Unmarshal(tool.AuthConfig, &authCfg); err == nil {
				if token, ok := authCfg["token"]; ok {
					req.Header.Set("Authorization", "Bearer "+token)
				}
			}
		}
	}

	timeout := time.Duration(tool.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{"raw": string(body)}, nil
	}
	return result, nil
}

func (s *ToolService) executeFunctionTool(tool *model.Tool, params json.RawMessage) (interface{}, error) {
	return s.executeBuiltinFunction(tool.Name, params)
}

func (s *ToolService) executeBuiltinFunction(name string, params json.RawMessage) (interface{}, error) {
	switch name {
	case "create_issue":
		return s.executeCreateIssue(params)
	case "update_issue":
		return s.executeUpdateIssue(params)
	case "list_issues":
		return s.executeListIssues(params)
	case "get_pr_diff":
		return s.builtinGetPRDiff(params)
	case "add_review_comment":
		return s.builtinAddReviewComment(params)
	case "list_pr_commits":
		return s.builtinListPRCommits(params)
	case "create_pr_review":
		return s.builtinCreatePRReview(params)
	default:
		return nil, fmt.Errorf("unknown builtin function: %s", name)
	}
}

// ==================== validateParams ====================

func (s *ToolService) validateParams(tool *model.Tool, params json.RawMessage) error {
	if len(tool.Params) == 0 || string(tool.Params) == "{}" || string(tool.Params) == "null" {
		return nil
	}
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(tool.Params, &schema); err != nil {
		return nil
	}
	if len(schema.Required) == 0 {
		return nil
	}
	var input map[string]interface{}
	if len(params) == 0 || string(params) == "null" {
		input = map[string]interface{}{}
	} else if err := json.Unmarshal(params, &input); err != nil {
		return common.BadRequest("Invalid input params JSON")
	}
	for _, field := range schema.Required {
		if _, ok := input[field]; !ok {
			return common.BadRequest(fmt.Sprintf("Missing required parameter: %s", field))
		}
	}
	return nil
}

// ==================== toResponse ====================

func (s *ToolService) toResponse(tool *model.Tool) *response.ToolResponse {
	var params interface{}
	if len(tool.Params) > 0 && string(tool.Params) != "null" {
		_ = json.Unmarshal(tool.Params, &params)
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

// ==================== GetCallLogs ====================

func (s *ToolService) GetCallLogs(wid uint64) ([]response.ToolCallLogResponse, error) {
	var logs []model.ToolCallLog
	if err := s.db.Where("workspace_id = ?", wid).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, common.Internal("Failed to get call logs")
	}
	result := make([]response.ToolCallLogResponse, len(logs))
	for i, l := range logs {
		var toolName string
		s.db.Model(&model.Tool{}).Where("id = ?", l.ToolID).Pluck("name", &toolName)
		result[i] = response.ToolCallLogResponse{
			ID:           l.ID,
			WorkspaceID:  l.WorkspaceID,
			AgentTaskID:  l.AgentTaskID,
			ToolID:       l.ToolID,
			ToolName:     toolName,
			AgentID:      l.AgentID,
			InputParams:  l.InputParams,
			OutputResult: l.OutputResult,
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
			DurationMs:   l.DurationMs,
			RateLimited:  l.RateLimited,
			CreatedAt:    l.CreatedAt,
		}
	}
	return result, nil
}

// ==================== RegisterBuiltinTools ====================

// RegisterBuiltinTools registers built-in tools. If wid > 0, registers only for that workspace;
// if wid == 0, registers for every workspace in the database.
func (s *ToolService) RegisterBuiltinTools(wid uint64) error {
	builtins := []struct {
		Name        string
		Description string
		Category    string
		ToolType    string
	}{
		{"create_issue", "创建一个新的 Issue", "project_management", "function"},
		{"update_issue", "更新已有的 Issue", "project_management", "function"},
		{"list_issues", "列出项目下的 Issue", "project_management", "function"},
		{"get_pr_diff", "获取 GitHub PR 的 diff 内容", "code", "function"},
		{"add_review_comment", "在 GitHub PR 上添加 review comment", "code", "function"},
		{"list_pr_commits", "列出 GitHub PR 的 commits", "code", "function"},
		{"create_pr_review", "在 GitHub PR 上创建 review", "code", "function"},
	}
	// Collect target workspace IDs
	var wids []uint64
	if wid != 0 {
		wids = []uint64{wid}
	} else {
		s.db.Model(&model.Workspace{}).Pluck("id", &wids)
	}
	for _, w := range wids {
		for _, b := range builtins {
			var count int64
			s.db.Model(&model.Tool{}).Where("name = ? AND workspace_id = ?", b.Name, w).Count(&count)
			if count == 0 {
				s.db.Create(&model.Tool{
					Name:        b.Name,
					Description: b.Description,
					Category:    b.Category,
					ToolType:    b.ToolType,
					IsBuiltin:   true,
					Status:      "active",
					WorkspaceID: &w,
				})
			}
		}
	}
	return nil
}