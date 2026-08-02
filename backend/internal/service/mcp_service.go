package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// MCPService manages MCP server connections and tool execution.
type MCPService struct {
	db *gorm.DB
}

func NewMCPService(db *gorm.DB) *MCPService {
	return &MCPService{db: db}
}

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards mutations against privilege escalation.
func (s *MCPService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage MCP configs")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage MCP configs")
	}
	return nil
}

// ======== Request/Response types ========

type MCPCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	ServerURL     string `json:"server_url" binding:"required"`
	TransportType string `json:"transport_type"`
	APIKey        string `json:"api_key"`
	IsEnabled     *bool  `json:"is_enabled"`
}

type MCPUpdateRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	ServerURL     *string `json:"server_url"`
	TransportType *string `json:"transport_type"`
	APIKey        *string `json:"api_key"`
	IsEnabled     *bool   `json:"is_enabled"`
}

type MCPResponse struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	WorkspaceID   uint64  `json:"workspace_id"`
	ServerURL     string  `json:"server_url"`
	TransportType string  `json:"transport_type"`
	IsEnabled     bool    `json:"is_enabled"`
	ToolsCount    int     `json:"tools_count"`
	LastSyncAt    *string `json:"last_sync_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"` // JSON schema string
}

type MCPExecuteRequest struct {
	ToolName    string                 `json:"tool_name" binding:"required"`
	Arguments   map[string]interface{} `json:"arguments"`
}

// ======== CRUD ========

func (s *MCPService) List(workspaceID uint64) ([]MCPResponse, error) {
	var configs []model.MCPConfig
	if err := s.db.Where("workspace_id = ?", workspaceID).Order("name ASC").Find(&configs).Error; err != nil {
		return nil, common.Internal("Failed to list MCP configs")
	}
	res := make([]MCPResponse, len(configs))
	for i, c := range configs {
		res[i] = s.toResponse(&c)
	}
	if res == nil {
		res = []MCPResponse{}
	}
	return res, nil
}

func (s *MCPService) Get(id uint64) (*MCPResponse, error) {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("MCP config not found")
		}
		return nil, common.Internal("Failed to get MCP config")
	}
	r := s.toResponse(&cfg)
	return &r, nil
}

func (s *MCPService) Create(workspaceID uint64, callerID uint64, req *MCPCreateRequest) (*MCPResponse, error) {
	if err := s.checkWorkspaceAdmin(workspaceID, callerID); err != nil {
		return nil, err
	}
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	transport := req.TransportType
	if transport == "" {
		transport = "sse"
	}

	cfg := model.MCPConfig{
		Name:          req.Name,
		Description:   req.Description,
		WorkspaceID:   workspaceID,
		ServerURL:     req.ServerURL,
		TransportType: transport,
		APIKey:        req.APIKey,
		IsEnabled:     enabled,
	}

	if err := s.db.Create(&cfg).Error; err != nil {
		return nil, common.Internal("Failed to create MCP config")
	}
	r := s.toResponse(&cfg)
	return &r, nil
}

func (s *MCPService) Update(id uint64, callerID uint64, req *MCPUpdateRequest) (*MCPResponse, error) {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("MCP config not found")
		}
		return nil, common.Internal("Failed to get MCP config")
	}
	if err := s.checkWorkspaceAdmin(cfg.WorkspaceID, callerID); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ServerURL != nil {
		updates["server_url"] = *req.ServerURL
	}
	if req.TransportType != nil {
		updates["transport_type"] = *req.TransportType
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if err := s.db.Model(&cfg).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update MCP config")
	}

	s.db.First(&cfg, id)
	r := s.toResponse(&cfg)
	return &r, nil
}

func (s *MCPService) Delete(id uint64, callerID uint64) error {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("MCP config not found")
		}
		return common.Internal("Failed to get MCP config")
	}
	if err := s.checkWorkspaceAdmin(cfg.WorkspaceID, callerID); err != nil {
		return err
	}
	if err := s.db.Delete(&cfg).Error; err != nil {
		return common.Internal("Failed to delete MCP config")
	}
	return nil
}

// ======== Tool Operations ========

// DiscoverTools fetches available tools from the remote MCP server via SSE.
func (s *MCPService) DiscoverTools(id uint64) ([]MCPTool, error) {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		return nil, common.NotFound("MCP config not found")
	}

	tools, err := s.fetchToolsFromServer(cfg.ServerURL, cfg.APIKey)
	if err != nil {
		return nil, common.Internal(fmt.Sprintf("Failed to discover tools: %v", err))
	}

	// Cache tools in config
	toolsJSON, _ := json.Marshal(tools)
	now := time.Now()
	s.db.Model(&cfg).Updates(map[string]interface{}{
		"tools_config": string(toolsJSON),
		"last_sync_at": now,
	})

	return tools, nil
}

// GetTools returns cached tools from the MCP config.
func (s *MCPService) GetTools(id uint64) ([]MCPTool, error) {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		return nil, common.NotFound("MCP config not found")
	}

	if cfg.ToolsConfig == "" {
		return []MCPTool{}, nil
	}

	var tools []MCPTool
	if err := json.Unmarshal([]byte(cfg.ToolsConfig), &tools); err != nil {
		return []MCPTool{}, nil
	}
	return tools, nil
}

// ExecuteTool calls a specific tool on the remote MCP server.
func (s *MCPService) ExecuteTool(id uint64, req *MCPExecuteRequest) (map[string]interface{}, error) {
	var cfg model.MCPConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		return nil, common.NotFound("MCP config not found")
	}

	result, err := s.callToolOnServer(cfg.ServerURL, cfg.APIKey, req.ToolName, req.Arguments)
	if err != nil {
		return nil, common.Internal(fmt.Sprintf("Tool execution failed: %v", err))
	}
	return result, nil
}

// ======== Helpers ========

func (s *MCPService) toResponse(cfg *model.MCPConfig) MCPResponse {
	toolCount := 0
	if cfg.ToolsConfig != "" {
		var tools []MCPTool
		if json.Unmarshal([]byte(cfg.ToolsConfig), &tools) == nil {
			toolCount = len(tools)
		}
	}

	var lastSync *string
	if cfg.LastSyncAt != nil {
		ts := cfg.LastSyncAt.Format("2006-01-02T15:04:05Z")
		lastSync = &ts
	}

	return MCPResponse{
		ID:            cfg.ID,
		Name:          cfg.Name,
		Description:   cfg.Description,
		WorkspaceID:   cfg.WorkspaceID,
		ServerURL:     cfg.ServerURL,
		TransportType: cfg.TransportType,
		IsEnabled:     cfg.IsEnabled,
		ToolsCount:    toolCount,
		LastSyncAt:    lastSync,
		CreatedAt:     cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// fetchToolsFromServer calls the MCP server's tools/list endpoint.
func (s *MCPService) fetchToolsFromServer(serverURL, apiKey string) ([]MCPTool, error) {
	url := strings.TrimRight(serverURL, "/") + "/tools/list"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		// Try alternate format: direct array of tools
		var tools []MCPTool
		if err2 := json.Unmarshal(body, &tools); err2 != nil {
			return nil, fmt.Errorf("failed to parse tools response: %s", string(body))
		}
		return tools, nil
	}
	return result.Tools, nil
}

// callToolOnServer calls the MCP server's tools/call endpoint.
func (s *MCPService) callToolOnServer(serverURL, apiKey, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	url := strings.TrimRight(serverURL, "/") + "/tools/call"

	payload := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{"raw": string(respBody)}, nil
	}
	return result, nil
}
