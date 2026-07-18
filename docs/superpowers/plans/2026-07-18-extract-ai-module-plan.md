# Extract AI Module into agent-service — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move all AI/Agent backend code from `backend/internal/` into the existing `agent-service/` microservice over 4 phases, making agent-service the single AI owner. Frontend stays unchanged.

**Architecture:** Main backend keeps thin reverse-proxy handlers for all AI routes (no frontend URL changes). agent-service gains all AI logic (handlers, services, LLM client, models) and communicates with main backend via HTTP for issue/project/user data queries. Both services share the same Postgres database.

**Tech Stack:** Go 1.22, Gin, GORM, Postgres, JWT auth

---

## Phase 1: Infrastructure

### Task 1.1: Add error helpers to agent-service common package

**Files:**
- Modify: `agent-service/internal/common/errors.go`

- [ ] **Step 1: Add Validation and NotFound constructors compatible with agent-service's AppError**

agent-service's `common` package has `Validation(msg)` but its signature takes only a string (no ErrorCode). It also lacks `AgentNotFound()`. Add them:

```go
// Append to agent-service/internal/common/errors.go after the existing Unauthorized function:

// Validation returns a validation error (422).
func Validation(msg string) *AppError {
	return &AppError{Code: 422, Message: msg}
}

// AgentNotFound returns a 404 for missing agents.
func AgentNotFound() *AppError {
	return &AppError{Code: 404, Message: "Agent not found"}
}
```

- [ ] **Step 2: Commit**

```bash
git add agent-service/internal/common/errors.go
git commit -m "feat(agent-service): add Validation and AgentNotFound error helpers"
```

### Task 1.2: Add model definitions to agent-service

**Files:**
- Create: `agent-service/internal/model/agent.go`
- Create: `agent-service/internal/model/ai_config.go`

- [ ] **Step 1: Create agent model**

```go
// agent-service/internal/model/agent.go
package model

import (
	"encoding/json"
	"time"
)

// Agent represents an AI agent that can be assigned to work items and @mentioned in comments.
type Agent struct {
	ID           uint64          `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `gorm:"index" json:"-"`
	CreatedByID  *uint64         `json:"created_by_id"`
	UpdatedByID  *uint64         `json:"updated_by_id"`

	WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
	Name          string          `gorm:"size:128;not null" json:"name"`
	Avatar        string          `gorm:"size:10;default:🤖" json:"avatar"`
	AgentType     string          `gorm:"size:20;default:builtin" json:"agent_type"`
	Capabilities  json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"capabilities"`
	Status        string          `gorm:"size:20;default:active" json:"status"`
	ModelOverride *string         `gorm:"size:50" json:"model_override,omitempty"`
	SystemPrompt  *string         `gorm:"type:text" json:"system_prompt,omitempty"`
}

func (Agent) TableName() string { return "agents" }

// AgentActivity records every action an AI agent performs for audit trail purposes.
type AgentActivity struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
	CreatedByID  *uint64    `json:"created_by_id"`
	UpdatedByID  *uint64    `json:"updated_by_id"`

	AgentID       uint64    `gorm:"not null;index" json:"agent_id"`
	IssueID       *uint64   `gorm:"index" json:"issue_id"`
	Action        string    `gorm:"size:50;not null" json:"action"`
	ResultSummary string    `gorm:"type:text" json:"result_summary"`
	Rating        *int      `gorm:"default:null" json:"rating,omitempty"`
	ExecutedAt    time.Time `gorm:"autoCreateTime" json:"executed_at"`
	AgentName     string    `gorm:"size:128" json:"agent_name"`
	TaskContext   *string   `gorm:"type:text" json:"task_context,omitempty"`
}

func (AgentActivity) TableName() string { return "agent_activities" }
```

- [ ] **Step 2: Create ai_config model**

```go
// agent-service/internal/model/ai_config.go
package model

import (
	"encoding/json"
	"time"
)

// AIConfig stores workspace-level AI configuration.
type AIConfig struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	Provider    string `gorm:"size:20;default:deepseek" json:"provider"`
	Model       string `gorm:"size:50;default:deepseek-chat" json:"model"`
	APIKey      string `gorm:"size:500;not null;column:api_key" json:"-"`
	MaxTokens   int    `gorm:"default:4096" json:"max_tokens"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	WorkspaceID uint64 `gorm:"not null;uniqueIndex" json:"workspace_id"`
}

func (AIConfig) TableName() string { return "ai_configs" }

// AIThread stores a conversation thread.
type AIThread struct {
	ID        uint64  `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Title       string  `gorm:"size:255" json:"title"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	UserID      uint64  `gorm:"not null;index" json:"user_id"`

	Messages []AIMessage `gorm:"foreignKey:ThreadID" json:"messages,omitempty"`
}

func (AIThread) TableName() string { return "ai_threads" }

// AIMessage stores a single message in a conversation.
type AIMessage struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ThreadID  uint64           `gorm:"not null;index" json:"thread_id"`
	Role      string           `gorm:"size:20;not null" json:"role"`
	Content   string           `gorm:"type:text;not null" json:"content"`
	ToolCalls json.RawMessage  `gorm:"type:jsonb;column:tool_calls" json:"tool_calls,omitempty"`
	ToolName  *string          `gorm:"size:50" json:"tool_name,omitempty"`
}

func (AIMessage) TableName() string { return "ai_messages" }
```

- [ ] **Step 3: Commit**

```bash
git add agent-service/internal/model/agent.go agent-service/internal/model/ai_config.go
git commit -m "feat(agent-service): add Agent, AgentActivity, AIConfig, AIThread, AIMessage models"
```

### Task 1.3: Move llm_client.go to agent-service

**Files:**
- Copy: `backend/internal/service/llm_client.go` → `agent-service/internal/llm/client.go`
- Remove: `backend/internal/service/llm_client.go` (removed later in Phase 4 after main backend stops using it)

- [ ] **Step 1: Create the directory and copy the file**

```bash
mkdir -p agent-service/internal/llm
cp backend/internal/service/llm_client.go agent-service/internal/llm/client.go
```

- [ ] **Step 2: Update package declaration**

Edit `agent-service/internal/llm/client.go` — change `package service` to `package llm`.

- [ ] **Step 3: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

Expected: builds successfully (llm package has zero imports from backend)

- [ ] **Step 4: Commit**

```bash
git add agent-service/internal/llm/client.go
git commit -m "feat(agent-service): copy llm_client.go as internal/llm/client.go"
```

### Task 1.4: Add BackendClient to agent-service

**Files:**
- Create: `agent-service/internal/client/backend_client.go`

- [ ] **Step 1: Create BackendClient for querying main backend data APIs**

```go
// agent-service/internal/client/backend_client.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BackendClient calls the main Reqmango backend for project/issue/user data.
type BackendClient struct {
	baseURL    string
	httpClient *http.Client
}

// IssueInfo is the minimal issue data returned by the main backend.
type IssueInfo struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	ProjectID   uint64 `json:"project_id"`
	WorkspaceID uint64 `json:"workspace_id"`
	SequenceID  int    `json:"sequence_id"`
	ProjectIdentifier string `json:"project_identifier"`
}

// ProjectInfo is the minimal project data.
type ProjectInfo struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// UserInfo is the minimal user data.
type UserInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *BackendClient) get(token, path string, result interface{}) error {
	req, _ := http.NewRequest("GET", c.baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *BackendClient) post(token, path string, reqBody, result interface{}) error {
	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// GetIssue fetches a single issue by ID.
func (c *BackendClient) GetIssue(issueID uint64, token string) (*IssueInfo, error) {
	var issue IssueInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/issues/%d", issueID), &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetProject fetches a single project by ID.
func (c *BackendClient) GetProject(projectID uint64, token string) (*ProjectInfo, error) {
	var project ProjectInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/projects/%d", projectID), &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetUser fetches a single user by ID.
func (c *BackendClient) GetUser(userID uint64, token string) (*UserInfo, error) {
	var user UserInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/users/%d", userID), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SearchIssues searches issues via RQL.
func (c *BackendClient) SearchIssues(token, query string) ([]IssueInfo, error) {
	var issues []IssueInfo
	if err := c.post(token, "/api/internal/issues/search", map[string]string{"query": query}, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add agent-service/internal/client/backend_client.go
git commit -m "feat(agent-service): add BackendClient for main backend data API calls"
```

### Task 1.5: Add internal data API to main backend

**Files:**
- Create: `backend/internal/handler/internal_data_handler.go`
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Create internal data handler**

```go
// backend/internal/handler/internal_data_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// InternalDataHandler serves internal data queries for the agent-service.
type InternalDataHandler struct {
	db *gorm.DB
}

func NewInternalDataHandler(db *gorm.DB) *InternalDataHandler {
	return &InternalDataHandler{db: db}
}

// GetIssue returns issue details by ID.
func (h *InternalDataHandler) GetIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var issue model.Issue
	if err := h.db.Preload("Project").First(&issue, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "issue not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get issue"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            issue.ID,
		"title":         issue.Name,
		"description":   issue.Description,
		"status":        issue.Status,
		"priority":      issue.Priority,
		"project_id":    issue.ProjectID,
		"workspace_id":  issue.Project.WorkspaceID,
		"sequence_id":   issue.SequenceID,
		"project_identifier": issue.Project.Identifier,
	})
}

// GetProject returns project details by ID.
func (h *InternalDataHandler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var project model.Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get project"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          project.ID,
		"name":        project.Name,
		"identifier":  project.Identifier,
		"workspace_id": project.WorkspaceID,
	})
}

// GetUser returns user details by ID.
func (h *InternalDataHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.DisplayName,
		"email": user.Email,
	})
}

// SearchIssues searches issues via RQL query.
func (h *InternalDataHandler) SearchIssues(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	// Delegate to the existing issue service search — for now use simple LIKE search
	var issues []model.Issue
	q := h.db.Preload("Project")
	if req.Query != "" {
		q = q.Where("name ILIKE ? OR description ILIKE ?", "%"+req.Query+"%", "%"+req.Query+"%")
	}
	if err := q.Limit(50).Find(&issues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to search issues"})
		return
	}
	result := make([]gin.H, len(issues))
	for i, issue := range issues {
		result[i] = gin.H{
			"id":            issue.ID,
			"title":         issue.Name,
			"description":   issue.Description,
			"status":        issue.Status,
			"priority":      issue.Priority,
			"project_id":    issue.ProjectID,
			"workspace_id":  issue.Project.WorkspaceID,
			"sequence_id":   issue.SequenceID,
			"project_identifier": issue.Project.Identifier,
		}
	}
	c.JSON(http.StatusOK, result)
}
```

- [ ] **Step 2: Register internal routes in router.go**

In `backend/internal/router/router.go`, after the `v1` group section, add:

```go
// ---- Internal Data API (for agent-service) ----
internalH := handler.NewInternalDataHandler(db)
internal := r.Group("/api/internal", authMiddleware)
{
    internal.GET("/issues/:id", internalH.GetIssue)
    internal.POST("/issues/search", internalH.SearchIssues)
    internal.GET("/projects/:id", internalH.GetProject)
    internal.GET("/users/:id", internalH.GetUser)
}
```

- [ ] **Step 3: Verify main backend compiles**

```bash
cd backend && go build ./...
```

Expected: builds successfully

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/internal_data_handler.go backend/internal/router/router.go
git commit -m "feat(backend): add /api/internal data API for agent-service"
```

### Task 1.6: Register new models in agent-service DB migration

**Files:**
- Modify: `agent-service/cmd/server/main.go`

- [ ] **Step 1: Add new models to AutoMigrate and wire BackendClient**

In `agent-service/cmd/server/main.go`, update the `db.AutoMigrate` call to include the new models:

```go
// Auto-migrate agent-specific tables (IF NOT EXISTS — safe with shared DB)
if err := db.AutoMigrate(
    &model.Loop{},
    &model.LoopRun{},
    &model.LoopIteration{},
    &model.AgentSession{},
    &model.Pipeline{},
    &model.PipelineRun{},
    &registry.AgentEntry{},
    // New AI models (tables already exist in shared DB, AutoMigrate adds columns only)
    &model.Agent{},
    &model.AgentActivity{},
    &model.AIConfig{},
    &model.AIThread{},
    &model.AIMessage{},
); err != nil {
    log.Fatalf("Failed to auto-migrate: %v", err)
}
```

And add BackendClient creation:

```go
// Create BackendClient for main backend data queries
backendClient := client.NewBackendClient(cfg.MainBackendURL)
_ = backendClient // will be used by services in Phases 2-3
```

- [ ] **Step 2: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

Expected: builds successfully

- [ ] **Step 3: Commit**

```bash
git add agent-service/cmd/server/main.go
git commit -m "feat(agent-service): register AI models in migration and wire BackendClient"
```

---

## Phase 2: Agent CRUD + Dispatch

### Task 2.1: Move and adapt agent_service.go to agent-service

**Files:**
- Create: `agent-service/internal/service/agent_service.go`
- The file is a modified copy of `backend/internal/service/agent_service.go`

- [ ] **Step 1: Copy the file and update package/imports**

```bash
cp backend/internal/service/agent_service.go agent-service/internal/service/agent_service.go
```

Edit `agent-service/internal/service/agent_service.go`:
- Change `package service` → keep as `package service`
- Replace imports: remove `"github.com/reqmango/backend/internal/common"` and `"github.com/reqmango/backend/internal/model"`
- Add: `"github.com/reqmango/agent-service/internal/common"`, `"github.com/reqmango/agent-service/internal/model"`, `"github.com/reqmango/agent-service/internal/llm"`
- Add: `"github.com/reqmango/agent-service/internal/client"`

Replace key references:
- `*model.Agent` → `*model.Agent` (same type name, different package)
- `model.Agent` → `model.Agent`
- `model.AgentActivity` → `model.AgentActivity`
- `model.IssueActivity` → Remove (not available in agent-service). The `recordActivity` function's IssueActivity creation must be removed — that table belongs to main backend.
- `*LLMClient` → `*llm.LLMClient`
- `Tool` / `Message` → `llm.Tool` / `llm.Message`
- `*IssueService` → Remove (replace with BackendClient HTTP calls)
- `*AIService` → Remove (will be added in Phase 3, set to nil initially)
- `common.Internal(...)` → `common.Internal(...)` (same package name, different module)
- `common.AgentNotFound()` → `common.AgentNotFound()`
- `common.Validation(...)` → `common.Validation(...)`
- `common.NotFound(...)` → `common.NotFound(...)`

Updated struct:

```go
type AgentService struct {
    db       *gorm.DB
    llm      *llm.LLMClient
    backend  *client.BackendClient
    aiSvc    *AIService // nil until Phase 3
}

func NewAgentService(db *gorm.DB, llm *llm.LLMClient, backend *client.BackendClient) *AgentService {
    return &AgentService{
        db:      db,
        llm:     llm,
        backend: backend,
    }
}
```

In `AutoTriage`, `AutoAssign`, `SummarizeCycle` — replace `s.issueSvc` calls with `s.backend.GetIssue(issueID, token)` HTTP calls. Add a `token` field or parameter. For now, in Phase 2, these methods can log that they need a token and return an error.


In `DispatchAgent`: remove the `s.aiSvc.ExecuteTool(...)` call and add a stub that returns an error until AIService arrives in Phase 3.

In `DispatchAgent`: remove the `model.IssueActivity` creation block (lines 314-324) — agent-service doesn't own that table.

- [ ] **Step 2: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

Expected: builds successfully

- [ ] **Step 3: Commit**

```bash
git add agent-service/internal/service/agent_service.go
git commit -m "feat(agent-service): move agent_service.go with adapted imports"
```

### Task 2.2: Create agent handler in agent-service

**Files:**
- Create: `agent-service/internal/handler/agent_handler.go`

- [ ] **Step 1: Create agent handler adapted for agent-service**

Copy from `backend/internal/handler/agent_handler.go`, adapting:

```go
// agent-service/internal/handler/agent_handler.go
package handler

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/reqmango/agent-service/internal/common"
    "github.com/reqmango/agent-service/internal/middleware"
    "github.com/reqmango/agent-service/internal/service"
)

type AgentHandler struct {
    svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
    return &AgentHandler{svc: svc}
}

func (h *AgentHandler) resolveWorkspaceID(c *gin.Context) uint64 {
    wsParam := c.Param("wsParam")
    var id uint64
    wss := strings.Split(wsParam, "-")
    if len(wss) >= 2 {
        strconv.ParseUint(wss[1], 10, 64) // skip, fallback below
    }
    strconv.ParseUint(wsParam, 10, 64) // try direct numeric
    return id
}

func (h *AgentHandler) getToken(c *gin.Context) string {
    auth := c.GetHeader("Authorization")
    return strings.TrimPrefix(auth, "Bearer ")
}

func appError(c *gin.Context, err error) bool {
    if err == nil {
        return false
    }
    if ae, ok := err.(*common.AppError); ok {
        c.JSON(ae.Code, ae)
        return true
    }
    c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    return true
}

// List — GET /api/v1/workspaces/:wsParam/agents
func (h *AgentHandler) List(c *gin.Context) {
    wsID := h.resolveWorkspaceID(c)
    if wsID == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
        return
    }
    agents, err := h.svc.ListByWorkspace(wsID)
    if appError(c, err) {
        return
    }
    c.JSON(http.StatusOK, agents)
}

// Create — POST /api/v1/workspaces/:wsParam/agents
func (h *AgentHandler) Create(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    wsID := h.resolveWorkspaceID(c)
    if wsID == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
        return
    }
    var req service.AgentCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    agent, svcErr := h.svc.Create(wsID, user.ID, &req)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusCreated, agent)
}

// Update — PUT /api/v1/workspaces/:wsParam/agents/:id
func (h *AgentHandler) Update(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    var req service.AgentUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    agent, svcErr := h.svc.Update(agentID, user.ID, &req)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, agent)
}

// GetByID — GET /api/v1/workspaces/:wsParam/agents/:id
func (h *AgentHandler) GetByID(c *gin.Context) {
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    agent, svcErr := h.svc.GetByID(agentID)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, agent)
}

// Delete — DELETE /api/v1/workspaces/:wsParam/agents/:id
func (h *AgentHandler) Delete(c *gin.Context) {
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    if appError(c, h.svc.Delete(agentID)) {
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Dispatch — POST /api/v1/workspaces/:wsParam/agents/:id/dispatch
func (h *AgentHandler) Dispatch(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    wsID := h.resolveWorkspaceID(c)
    var req struct {
        Task        string  `json:"task"`
        IssueID     *uint64 `json:"issue_id"`
        ProjectID   *uint64 `json:"project_id"`
        TriggeredBy string  `json:"triggered_by"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    if req.TriggeredBy == "" {
        req.TriggeredBy = "manual"
    }
    if req.Task == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "task is required"})
        return
    }
    dispCtx := &service.DispatchContext{
        IssueID:     req.IssueID,
        ProjectID:   req.ProjectID,
        WorkspaceID: wsID,
        TriggeredBy: req.TriggeredBy,
    }
    activity, svcErr := h.svc.DispatchAgent(agentID, user.ID, req.Task, dispCtx)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activity)
}

// GetActivity — GET /api/v1/workspaces/:wsParam/agents/:id/activity
func (h *AgentHandler) GetActivity(c *gin.Context) {
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    activities, svcErr := h.svc.GetActivity(agentID)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activities)
}

// ListWorkspaceActivity — GET /api/v1/workspaces/:wsParam/agents/activity
func (h *AgentHandler) ListWorkspaceActivity(c *gin.Context) {
    wsID := h.resolveWorkspaceID(c)
    if wsID == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid workspace"})
        return
    }
    var agentID *uint64
    if aid, err := strconv.ParseUint(c.Query("agent_id"), 10, 64); err == nil && aid > 0 {
        agentID = &aid
    }
    action := c.Query("action")
    limit := 50
    if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
        limit = l
    }
    activities, svcErr := h.svc.ListWorkspaceActivity(wsID, agentID, action, limit)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activities)
}

// UpdateActivityFeedback — PATCH /api/v1/workspaces/:wsParam/agents/activity/:id/feedback
func (h *AgentHandler) UpdateActivityFeedback(c *gin.Context) {
    activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid activity id"})
        return
    }
    var req struct {
        Rating int `json:"rating"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    if appError(c, h.svc.UpdateActivityFeedback(activityID, req.Rating)) {
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// AutoTriage — POST /api/v1/workspaces/:wsParam/agents/:id/auto-triage
func (h *AgentHandler) AutoTriage(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    var req struct {
        IssueID uint64 `json:"issue_id"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    activity, svcErr := h.svc.AutoTriage(agentID, req.IssueID, user.ID)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activity)
}

// AutoAssign — POST /api/v1/workspaces/:wsParam/agents/:id/auto-assign
func (h *AgentHandler) AutoAssign(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    var req struct {
        IssueID uint64 `json:"issue_id"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    activity, svcErr := h.svc.AutoAssign(agentID, req.IssueID, user.ID)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activity)
}

// HandleMention — POST /api/v1/workspaces/:wsParam/agents/:id/mention
func (h *AgentHandler) HandleMention(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid agent id"})
        return
    }
    var req struct {
        CommentID   uint64  `json:"comment_id"`
        IssueID     *uint64 `json:"issue_id"`
        CommentBody string  `json:"comment_body"`
        IssueName   string  `json:"issue_name"`
        WorkspaceID uint64  `json:"workspace_id"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    activity, svcErr := h.svc.HandleMention(agentID, req.CommentID, user.ID, req.CommentBody, req.IssueName, req.IssueID)
    if appError(c, svcErr) {
        return
    }
    c.JSON(http.StatusOK, activity)
}

// AutoTriageProject — POST /api/v1/projects/:projectId/agent/auto-triage
func (h *AgentHandler) AutoTriageProject(c *gin.Context) {
    // stub for project-level route — Phase 2 proxy fallback
    c.JSON(http.StatusOK, gin.H{"message": "auto-triage proxied to agent-service"})
}

// AutoAssignProject — POST /api/v1/projects/:projectId/agent/auto-assign
func (h *AgentHandler) AutoAssignProject(c *gin.Context) {
    // stub for project-level route — Phase 2 proxy fallback
    c.JSON(http.StatusOK, gin.H{"message": "auto-assign proxied to agent-service"})
}
```

- [ ] **Step 2: Commit**

```bash
git add agent-service/internal/handler/agent_handler.go
git commit -m "feat(agent-service): add AgentHandler with all CRUD and dispatch endpoints"
```

### Task 2.3: Register agent routes in agent-service main.go

**Files:**
- Modify: `agent-service/cmd/server/main.go`

- [ ] **Step 1: Add agent routes to agent-service router**

After the registry routes section in `agent-service/cmd/server/main.go`, add:

```go
// Initialize agent service and handler
llmClient := llm.NewLLMClient(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, cfg.AIProvider)
agentSvc := service.NewAgentService(db, llmClient, backendClient)
agentH := handler.NewAgentHandler(agentSvc)

// Agent routes (same URLs as main backend, so proxy just forwards path unchanged)
agents := api.Group("/agents")
{
    agents.GET("", agentH.List)
    agents.POST("", agentH.Create)
    agents.GET("/activity", agentH.ListWorkspaceActivity)
    agents.PATCH("/activity/:id/feedback", agentH.UpdateActivityFeedback)
    agents.GET("/:id", agentH.GetByID)
    agents.PUT("/:id", agentH.Update)
    agents.DELETE("/:id", agentH.Delete)
    agents.POST("/:id/dispatch", agentH.Dispatch)
    agents.GET("/:id/activity", agentH.GetActivity)
    agents.POST("/:id/auto-triage", agentH.AutoTriage)
    agents.POST("/:id/auto-assign", agentH.AutoAssign)
    agents.POST("/:id/mention", agentH.HandleMention)
}
```

Add `AIAPIKey`, `AIModel`, `AIBaseURL`, `AIProvider` fields to `config.Config`:

```go
type Config struct {
    Port           string
    DatabaseURL    string
    MainBackendURL string
    SecretKey      string
    AIAPIKey       string
    AIModel        string
    AIBaseURL      string
    AIProvider     string
}
```

Update `config.Load()` to read these from env:

```go
return &Config{
    Port:           port,
    DatabaseURL:    os.Getenv("DATABASE_URL"),
    MainBackendURL: getEnv("MAIN_BACKEND_URL", "http://localhost:8000"),
    SecretKey:      getEnv("SECRET_KEY", "change-me-in-production"),
    AIAPIKey:       os.Getenv("AI_API_KEY"),
    AIModel:        getEnv("AI_MODEL", "deepseek-chat"),
    AIBaseURL:      getEnv("AI_BASE_URL", "https://api.deepseek.com/v1"),
    AIProvider:     getEnv("AI_PROVIDER", "deepseek"),
}
```

- [ ] **Step 2: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add agent-service/cmd/server/main.go agent-service/config/config.go
git commit -m "feat(agent-service): register agent routes and wire AgentService"
```

### Task 2.4: Replace main backend agent routes with reverse proxy

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Replace agent handler routes with proxy**

In `backend/internal/router/router.go`, remove the agent handler creation and replace all agent routes (non-loop/non-pipeline/non-registry) with a single reverse proxy:

Remove:
```go
agentH := handler.NewAgentHandler(agentSvc)
```

Replace the workspace agent routes block:

```go
// Old (remove):
workspaces.GET("/:wsParam/agents", agentH.List)
workspaces.POST("/:wsParam/agents", agentH.Create)
workspaces.GET("/:wsParam/agents/activity", agentH.ListWorkspaceActivity)
workspaces.PATCH("/:wsParam/agents/activity/:id/feedback", agentH.UpdateActivityFeedback)
workspaces.GET("/:wsParam/agents/:id", agentH.GetByID)
workspaces.PUT("/:wsParam/agents/:id", agentH.Update)
workspaces.DELETE("/:wsParam/agents/:id", agentH.Delete)
workspaces.POST("/:wsParam/agents/:id/dispatch", agentH.Dispatch)
workspaces.GET("/:wsParam/agents/:id/activity", agentH.GetActivity)
workspaces.POST("/:wsParam/agents/:id/auto-triage", agentH.AutoTriage)
workspaces.POST("/:wsParam/agents/:id/auto-assign", agentH.AutoAssign)

// New:
// AI Agents — proxied to agent-service:8001
workspaces.Any("/:wsParam/agents/*path", agentProxy)
```

Also replace project-level agent routes:

```go
// Old (remove):
projects.POST("/:projectId/agent/auto-triage", agentH.AutoTriageProject)
projects.POST("/:projectId/agent/auto-assign", agentH.AutoAssignProject)

// New:
projects.Any("/:projectId/agent/*path", agentProxy)
```

- [ ] **Step 2: Remove agentSvc/agentH references from router**

Remove `agentSvc := service.NewAgentService(db, llmClient, issueSvc, aiSvc)` and the `SetAgentService` calls:

```go
// Remove these lines:
agentSvc := service.NewAgentService(db, llmClient, issueSvc, aiSvc)
automationSvc.SetAgentService(agentSvc)
commentSvc.SetAgentService(agentSvc)
```

- [ ] **Step 3: Verify main backend compiles (it won't yet — see Task 2.5)**

```bash
cd backend && go build ./...
```

Expected: compilation errors in automation_service.go and comment_service.go — these will be fixed in Task 2.5.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/router/router.go
git commit -m "refactor(backend): replace agent routes with reverse proxy to agent-service"
```

### Task 2.5: Refactor automation_service and comment_service to use HTTP client

**Files:**
- Modify: `backend/internal/service/automation_service.go`
- Modify: `backend/internal/service/comment_service.go`

- [ ] **Step 1: Create AgentClient in main backend for calling agent-service**

Create `backend/internal/client/agent_client.go`:

```go
// backend/internal/client/agent_client.go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    "gorm.io/gorm"
)

type AgentClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewAgentClient(baseURL string) *AgentClient {
    return &AgentClient{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *AgentClient) do(method, path string, body interface{}, result interface{}) error {
    var reqBody io.Reader
    if body != nil {
        b, _ := json.Marshal(body)
        reqBody = bytes.NewReader(b)
    }
    req, _ := http.NewRequest(method, c.baseURL+path, reqBody)
    req.Header.Set("Content-Type", "application/json")
    // Auth handled by internal network — no JWT needed for internal calls
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("agent-service request failed: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("agent-service %d: %s", resp.StatusCode, string(respBody))
    }
    if result != nil {
        return json.NewDecoder(resp.Body).Decode(result)
    }
    return nil
}

// DispatchAgent dispatches an agent task.
func (c *AgentClient) DispatchAgent(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, triggeredBy string) error {
    reqBody := map[string]interface{}{
        "task":         task,
        "issue_id":     issueID,
        "project_id":   projectID,
        "triggered_by": triggeredBy,
    }
    return c.do("POST", fmt.Sprintf("/api/v1/workspaces/%d/agents/%d/dispatch", workspaceID, agentID), reqBody, nil)
}

// HandleMention triggers agent mention handling.
func (c *AgentClient) HandleMention(workspaceID, agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) error {
    reqBody := map[string]interface{}{
        "comment_id":   commentID,
        "issue_id":     issueID,
        "comment_body": commentBody,
        "issue_name":   issueName,
        "workspace_id": workspaceID,
    }
    return c.do("POST", fmt.Sprintf("/api/v1/workspaces/%d/agents/%d/mention", workspaceID, agentID), reqBody, nil)
}
```



- [ ] **Step 1 (revised): Create AgentClient in main backend**

Create `backend/internal/client/agent_client.go` with the code above.

- [ ] **Step 2: Update automation_service.go**

Replace `*AgentService` with `*client.AgentClient` in `DefaultActionExecutor`:

```go
// In DefaultActionExecutor struct:
type DefaultActionExecutor struct {
    handlers   map[string]ActionHandler
    db         *gorm.DB
    agentClient *client.AgentClient // was: agentSvc *AgentService
}

func NewDefaultActionExecutor(db *gorm.DB, agentClient *client.AgentClient) *DefaultActionExecutor {
    executor := &DefaultActionExecutor{
        handlers:    make(map[string]ActionHandler),
        db:          db,
        agentClient: agentClient,
    }
    executor.registerBuiltinActions()
    return executor
}
```

Update `handleDispatchAgent`:
```go
func (e *DefaultActionExecutor) handleDispatchAgent(action Action, context map[string]interface{}, db *gorm.DB) error {
    if e.agentClient == nil {
        return fmt.Errorf("agent service not available")
    }
    // ... same extraction of issueID, projectID, workspaceID, agentID ...

    return e.agentClient.DispatchAgent(workspaceID, agentID, 1, task, projectIDPtr, workspaceID, "automation")
}
```

```go
_, err := e.agentSvc.DispatchAgent(agentID, 1, task, dispatchCtx)
```


Update `SetAgentService`:
```go
func (s *AutomationService) SetAgentService(agentClient *client.AgentClient) {
    if exec, ok := s.actionExecutor.(*DefaultActionExecutor); ok {
        exec.agentClient = agentClient
    }
}
```

And in `NewAutomationService`:
```go
actionExecutor := NewDefaultActionExecutor(db, nil) // agentClient set via SetAgentService
```

- [ ] **Step 3: Update comment_service.go**

Replace `*AgentService` with `*client.AgentClient`:

```go
type CommentService struct {
    db              *gorm.DB
    notificationSvc *NotificationService
    agentClient     *client.AgentClient // was: agentSvc *AgentService
    automationSvc   *AutomationService
}

func (s *CommentService) SetAgentService(agentClient *client.AgentClient) {
    s.agentClient = agentClient
}
```

Update the mention handling (two places):
```go
if s.agentClient != nil {
    var agents []model.Agent
    s.db.Where("workspace_id = ? AND name IN ? AND status = 'active'", issue.Project.WorkspaceID, mentioned).Find(&agents)
    for _, agent := range agents {
        go func(a model.Agent) {
            s.agentClient.HandleMention(a.WorkspaceID, a.ID, c.ID, authorID, body, issue.Name, &issueID)
        }(agent)
    }
}
```


```go
s.agentSvc.HandleMention(a.ID, issue.Project.WorkspaceID, authorID, body, issue.Name, &issueID)
```

```go
func (s *AgentService) HandleMention(agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64)
```


```go
func (c *AgentClient) HandleMention(workspaceID, agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) error
```

- [ ] **Step 4: Update router.go to wire AgentClient**

In `backend/internal/router/router.go`, add:
```go
agentClient := client.NewAgentClient(cfg.AgentServiceURL)
automationSvc.SetAgentService(agentClient)
commentSvc.SetAgentService(agentClient)
```

- [ ] **Step 5: Verify main backend compiles**

```bash
cd backend && go build ./...
```

Fix any compilation errors.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/client/agent_client.go backend/internal/service/automation_service.go backend/internal/service/comment_service.go backend/internal/router/router.go
git commit -m "refactor(backend): replace direct AgentService calls with AgentClient HTTP calls"
```

### Task 2.6: Phase 2 verification

- [ ] **Step 1: Run agent-service tests**

```bash
cd agent-service && go test ./...
```

- [ ] **Step 2: Run backend tests**

```bash
cd backend && go test ./... 2>&1 | head -60
```

- [ ] **Step 3: Commit any fixes**

---

## Phase 3: AI Features

### Task 3.1: Adapt and move ai_service.go (the 57KB migration)

**Files:**
- Create: `agent-service/internal/service/ai_service.go`

This is the largest change. The file is 57KB with deep coupling to `issueSvc` and `projectSvc` via direct GORM queries.

- [ ] **Step 1: Identify all DB queries that must become BackendClient calls**

Key areas in ai_service.go that need rewriting:
- `isProjectMember(projectID, userID)` — queries `project_members` table → replace with BackendClient check or direct DB (table exists in shared DB)
- `buildSystemPrompt` — uses AIContext fields, no DB → no change
- `GetTools()` / `getTools()` — returns tool definitions, no DB → no change
- `Chat()` — uses s.db for AIThread/AIMessage CRUD → direct DB OK (tables are in shared DB)
- `Search()` — calls s.issueSvc/RQL → replace with BackendClient.SearchIssues
- `CreatePreview()` — uses DB for issue types/states/labels → shared DB or BackendClient
- `Analyze()` — uses DB for project stats → BackendClient or shared DB
- `SuggestLabels()` — uses DB for labels → shared DB
- `SprintPlan()` — uses DB for cycles/issues → shared DB or BackendClient
- `Chart()` — uses DB for metrics → shared DB or BackendClient
- `PageAI()` — page content processing → mostly self-contained
- `TriageAnalyze()` — similar to Analyze
- `AssistComment()` — similar to Chat
- `ExecuteTool()` — executes AI tools → needs DB access for CRUD tools

Strategy: For tables that agent-service owns (ai_threads, ai_messages) or shares (agents, agent_activities), use direct GORM. For tables owned by main backend (issues, projects, project_members, users, labels, cycles, modules), use BackendClient HTTP calls through the issue/project internal data API.

- [ ] **Step 2: Copy and adapt the file**

```bash
cp backend/internal/service/ai_service.go agent-service/internal/service/ai_service.go
```

Edit `agent-service/internal/service/ai_service.go`:
- Change `package service` → keep as `package service`
- Replace `"github.com/reqmango/backend/internal/model"` → `"github.com/reqmango/agent-service/internal/model"`
- Replace `"github.com/reqmango/backend/internal/common"` → `"github.com/reqmango/agent-service/internal/common"`
- Add `"github.com/reqmango/agent-service/internal/llm"` (for llm.LLMClient, llm.Message, etc.)
- Add `"github.com/reqmango/agent-service/internal/client"` (for BackendClient)
- Remove `*IssueService` and `*ProjectService` fields, replace with `*client.BackendClient`
- Update `NewAIService` to take `*client.BackendClient` instead of issue/project services
- All `*LLMClient` → `*llm.LLMClient`
- All `Message`, `Tool`, etc. → `llm.Message`, `llm.Tool`, etc.
- All `model.Issue` queries → `s.backend.GetIssue(id, token)` — but need token parameter propagation

The token problem: `ai_service.go` methods don't receive tokens. Solution: add a `token string` parameter to public methods, or store it in AIContext. The handler already has access to the request, so the cleanest approach is to pass token through AIContext:

```go
type AIContext struct {
    // ... existing fields ...
    Token string // JWT token for backend API calls (not serialized)
}
```

This minimizes signature changes across all AI service methods.

- [ ] **Step 3: Rewrite DB-dependent methods**

For methods that directly query main backend tables (issues, projects, members, labels, cycles), replace with BackendClient calls or use the shared DB directly.

Since both services share the same Postgres, the simplest approach for Phase 3 is: **keep direct DB queries for read-only access to main backend tables** (issues, projects, labels, etc.). The shared DB approach avoids the massive token-passing refactor. Write operations (issue update, create) must go through BackendClient to ensure business logic runs in main backend.

Concrete rule:
- **Read queries on main backend tables** → use `s.db` directly (shared DB)
- **Write operations on main backend data** → use `s.backend` HTTP calls
- **Tables owned by agent-service (ai_threads, ai_messages, agents)** → use `s.db` directly

This dramatically reduces the migration scope while maintaining correctness.

- [ ] **Step 4: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

Fix any compilation errors iteratively.

- [ ] **Step 5: Commit**

```bash
git add agent-service/internal/service/ai_service.go
git commit -m "feat(agent-service): move and adapt ai_service.go with shared DB access"
```

### Task 3.2: Create AI handler in agent-service

**Files:**
- Create: `agent-service/internal/handler/ai_handler.go`

- [ ] **Step 1: Create AI handler adapted for agent-service**

Copy from `backend/internal/handler/ai_handler.go`, adapting imports and references as done for agent_handler. Keep the same route handler method signatures.

Key adaptations:
- `getUserID` → use `middleware.GetCurrentUser(c).ID`
- `buildContext` → add token to AIContext
- All `*model.User` → `*middleware.UserInfo`

```go
// agent-service/internal/handler/ai_handler.go
package handler

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/reqmango/agent-service/internal/service"
    "gorm.io/gorm"
)

type AIHandler struct {
    svc *service.AIService
    db  *gorm.DB
}

func NewAIHandler(svc *service.AIService, db *gorm.DB) *AIHandler {
    return &AIHandler{svc: svc, db: db}
}

func (h *AIHandler) getToken(c *gin.Context) string {
    auth := c.GetHeader("Authorization")
    return strings.TrimPrefix(auth, "Bearer ")
}

// GetAIConfig — GET /api/v1/workspaces/:wsParam/ai-config
func (h *AIHandler) GetAIConfig(c *gin.Context) {
    wsID := resolveWorkspaceIDFromParam(c.Param("wsParam"))
    config, err := h.svc.GetAIConfig(wsID)
    if appError(c, err) {
        return
    }
    c.JSON(http.StatusOK, config)
}

// UpdateAIConfig — PUT /api/v1/workspaces/:wsParam/ai-config
func (h *AIHandler) UpdateAIConfig(c *gin.Context) {
    wsID := resolveWorkspaceIDFromParam(c.Param("wsParam"))
    var req service.AIConfigUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    config, err := h.svc.UpdateAIConfig(wsID, &req)
    if appError(c, err) {
        return
    }
    c.JSON(http.StatusOK, config)
}

// Chat — POST /api/v1/projects/:projectId/ai/chat
func (h *AIHandler) Chat(c *gin.Context) {
    token := h.getToken(c)
    actx := h.buildContext(c, token)
    var req service.AIChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }
    ch, err := h.svc.Chat(c.Request.Context(), &req, actx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
        return
    }
    // SSE streaming
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    for evt := range ch {
        data, _ := json.Marshal(evt)
        fmt.Fprintf(c.Writer, "data: %s\n\n", data)
        c.Writer.Flush()
    }
}

// Search, CreatePreview, Analyze, SuggestLabels, SprintPlan, Chart, PageAI,
// TriageAnalyze, AssistComment follow the same pattern as the original ai_handler.go.
// Each builds context, calls the corresponding service method, and returns JSON.
// For brevity, the full handler file is ~400 lines; the crucial point is that
// each method calls h.svc with an AIContext that includes the JWT token.
```

The full handler is a copy of `backend/internal/handler/ai_handler.go` with:
- Package changed to `handler`
- Imports adapted to agent-service module path
- Token extracted from `Authorization` header and passed into `AIContext.Token`

- [ ] **Step 2: Commit**

```bash
git add agent-service/internal/handler/ai_handler.go
git commit -m "feat(agent-service): add AIHandler for all AI endpoints"
```

### Task 3.3: Register AI routes in agent-service

**Files:**
- Modify: `agent-service/cmd/server/main.go`

- [ ] **Step 1: Add AI routes**

```go
// Initialize AI service and handler
aiSvc := service.NewAIService(db, llmClient, backendClient)
aiH := handler.NewAIHandler(aiSvc, db)

// AI config routes
api.Any("/ai-config", func(c *gin.Context) {
    // Route by method
    switch c.Request.Method {
    case "GET":
        aiH.GetAIConfig(c)
    case "PUT":
        aiH.UpdateAIConfig(c)
    default:
        c.JSON(405, gin.H{"message": "method not allowed"})
    }
})

// AI routes (project-level)
// These come via proxy from main backend's /api/v1/projects/:projectId/ai/*
ai2 := api.Group("/ai")
{
    ai2.POST("/chat", aiH.Chat)
    ai2.POST("/search", aiH.Search)
    ai2.POST("/create", aiH.CreatePreview)
    ai2.POST("/analyze", aiH.Analyze)
    ai2.POST("/suggest-labels", aiH.SuggestLabels)
    ai2.POST("/sprint-plan", aiH.SprintPlan)
    ai2.POST("/chart", aiH.Chart)
}
```


Cleanest approach: add a second route group in agent-service for project-level routes:

```go
// Project-level AI routes (proxied from main backend)
proj := r.Group("/api/v1/projects/:projectId", auth)
{
    proj.POST("/ai/chat", aiH.Chat)
    proj.POST("/ai/search", aiH.Search)
    proj.POST("/ai/create", aiH.CreatePreview)
    proj.POST("/ai/analyze", aiH.Analyze)
    proj.POST("/ai/suggest-labels", aiH.SuggestLabels)
    proj.POST("/ai/sprint-plan", aiH.SprintPlan)
    proj.POST("/ai/chart", aiH.Chart)
    proj.POST("/intake/:issueId/ai-analyze", aiH.TriageAnalyze)
    proj.POST("/agent/auto-triage", agentH.AutoTriageProject)
    proj.POST("/agent/auto-assign", agentH.AutoAssignProject)
}

// Issue-level AI routes (proxied from main backend)
issues := r.Group("/api/v1/issues/:issueId", auth)
{
    issues.POST("/ai/comment", aiH.AssistComment)
    issues.POST("/agents/:agentId/mention", agentH.HandleMention)
}

// Page AI
pages := r.Group("/api/v1/pages/:pageId", auth)
{
    pages.POST("/ai", aiH.PageAI)
}
```

Also need to add `strconv` import for `resolveWorkspaceIDFromParam`:

```go
func resolveWorkspaceIDFromParam(wsParam string) uint64 {
    parts := strings.Split(wsParam, "-")
    if len(parts) >= 2 {
        if id, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
            return id
        }
    }
    id, _ := strconv.ParseUint(wsParam, 10, 64)
    return id
}
```

- [ ] **Step 2: Verify agent-service compiles**

```bash
cd agent-service && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add agent-service/cmd/server/main.go
git commit -m "feat(agent-service): register AI routes"
```

### Task 3.4: Replace main backend AI routes with reverse proxy

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Replace all AI routes with proxy**

In `backend/internal/router/router.go`, replace the AI handler setup and routes:

Remove `aiH := handler.NewAIHandler(aiSvc, db)` and `aiSvc := service.NewAIService(db, llmClient, issueSvc, projectSvc)`.

Replace AI routes with proxy (the existing `agentProxy` already points to agent-service):

```go
// Workspace AI config — proxy
workspaces.Any("/:wsParam/ai-config/*path", agentProxy)

// Project AI — proxy
projects.Any("/:projectId/ai/*path", agentProxy)
projects.Any("/:projectId/intake/:issueId/ai-analyze", agentProxy)

// Issue AI — proxy  
issues.Any("/:issueId/ai/*path", agentProxy)
issues.Any("/:issueId/agents/:agentId/mention", agentProxy)

// Page AI — proxy
v1.Any("/pages/:pageId/ai", authMiddleware, agentProxy)
```

Note: we need to ensure the `agentProxy` is available earlier in the route setup. Currently it's defined inside the workspaces group. Move it to a package-level or top-of-function declaration.


```go
agentProxy := reverseProxy(cfg.AgentServiceURL)
```

Move this line before any route group that uses it.

- [ ] **Step 2: Verify main backend compiles**

```bash
cd backend && go build ./...
```

Fix issues: remove aiSvc/aiH initialization, remove llmClient creation, remove unused imports.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/router/router.go
git commit -m "refactor(backend): replace AI routes with reverse proxy to agent-service"
```

---

## Phase 4: Cleanup

### Task 4.1: Remove agent/AI code from main backend

**Files:**
- Remove: `backend/internal/handler/agent_handler.go`
- Remove: `backend/internal/handler/ai_handler.go`
- Remove: `backend/internal/service/agent_service.go`
- Remove: `backend/internal/service/ai_service.go`
- Remove: `backend/internal/service/llm_client.go`
- Keep models: `backend/internal/model/agent.go` and `backend/internal/model/ai_config.go` (comment_service still imports model.Agent)

- [ ] **Step 1: Remove handler and service files**

```bash
rm backend/internal/handler/agent_handler.go
rm backend/internal/handler/ai_handler.go
rm backend/internal/service/agent_service.go
rm backend/internal/service/ai_service.go
rm backend/internal/service/ai_service.go.bak
rm backend/internal/service/llm_client.go
```

- [ ] **Step 2: Verify main backend compiles**

```bash
cd backend && go build ./...
```

Fix any remaining import errors.

- [ ] **Step 3: Commit**

```bash
git rm backend/internal/handler/agent_handler.go backend/internal/handler/ai_handler.go backend/internal/service/agent_service.go backend/internal/service/ai_service.go backend/internal/service/ai_service.go.bak backend/internal/service/llm_client.go
git commit -m "refactor(backend): remove agent/AI handler and service files (migrated to agent-service)"
```

### Task 4.2: Clean up model references in main backend

**Files:**
- Modify: `backend/internal/model/agent.go` (reduce to minimal struct used by comment_service)
- Or keep as-is if comment_service still needs it

- [ ] **Step 1: Assess remaining model usage**

```bash
cd backend && go build ./...
```

If the build passes without `model/agent.go` and `model/ai_config.go`, they can be removed. Otherwise, keep them as they are used by remaining services.

The `comment_service.go` still queries `model.Agent` directly (for finding agents by name). So keep `model/agent.go`. The `model/ai_config.go` is no longer used by main backend (AIConfig is accessed via agent-service proxy), so it can be removed.

- [ ] **Step 2: Remove ai_config.go model**

```bash
rm backend/internal/model/ai_config.go
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git rm backend/internal/model/ai_config.go
git commit -m "refactor(backend): remove ai_config model (migrated to agent-service)"
```

### Task 4.3: Simplify agent_client.go in agent-service

**Files:**
- Modify: `agent-service/internal/client/agent_client.go`

- [ ] **Step 1: Remove agent API calls from agent_client.go**

The existing `agent_client.go` calls main backend's agent API (ListByWorkspace, DispatchAgent). Since agent-service now owns these endpoints, remove these methods and keep only the BackendClient data query methods.

```go
// agent-service/internal/client/agent_client.go
// After migration, this file only contains BackendClient (data queries).
// Agent dispatch/CRUD is now in-process via AgentService.

// DELETE: AgentInfo, DispatchResult, DispatchRequest types
// DELETE: ListByWorkspace(), DispatchAgent() methods
// DELETE: NewAgentClient() constructor
// KEEP: (nothing — the file can be removed if BackendClient is in backend_client.go)
```


- [ ] **Step 2: Update loop_service.go to use AgentService directly**

In `agent-service/internal/service/loop_service.go`, replace `AgentClient.DispatchAgent()` calls with direct `AgentService.DispatchAgent()` calls.

- [ ] **Step 3: Remove old agent_client.go**

```bash
rm agent-service/internal/client/agent_client.go
cd agent-service && go build ./...
```

Fix any compilation errors (update loop_service.go).

- [ ] **Step 4: Commit**

```bash
git rm agent-service/internal/client/agent_client.go
git add agent-service/internal/service/loop_service.go
git commit -m "refactor(agent-service): remove redundant agent_client, use AgentService directly"
```

### Task 4.4: Full regression test

- [ ] **Step 1: Run all backend tests**

```bash
cd backend && go test ./...
```

- [ ] **Step 2: Run all agent-service tests**

```bash
cd agent-service && go test ./...
```

- [ ] **Step 3: Run frontend tests**

```bash
cd frontend && npm test
```

- [ ] **Step 4: Manual smoke test checklist**

- [ ] Agent CRUD: Create, list, update, delete an agent via proxy
- [ ] Agent dispatch: Dispatch an agent with a task
- [ ] Agent activity: View activity log
- [ ] AI chat: Send a chat message
- [ ] AI config: Get and update AI config
- [ ] @mention: Comment with @agent-name triggers agent
- [ ] Automation: Automation rule with dispatch_agent action works

- [ ] **Step 5: Commit any final fixes**

---

## Summary of File Changes

| Action | File |
|--------|------|
| CREATE | `agent-service/internal/common/errors.go` (append) |
| CREATE | `agent-service/internal/model/agent.go` |
| CREATE | `agent-service/internal/model/ai_config.go` |
| CREATE | `agent-service/internal/llm/client.go` |
| CREATE | `agent-service/internal/client/backend_client.go` |
| CREATE | `agent-service/internal/service/agent_service.go` |
| CREATE | `agent-service/internal/handler/agent_handler.go` |
| CREATE | `agent-service/internal/service/ai_service.go` |
| CREATE | `agent-service/internal/handler/ai_handler.go` |
| CREATE | `backend/internal/handler/internal_data_handler.go` |
| CREATE | `backend/internal/client/agent_client.go` |
| MODIFY | `agent-service/cmd/server/main.go` |
| MODIFY | `agent-service/config/config.go` |
| MODIFY | `backend/internal/router/router.go` |
| MODIFY | `backend/internal/service/automation_service.go` |
| MODIFY | `backend/internal/service/comment_service.go` |
| REMOVE | `backend/internal/handler/agent_handler.go` |
| REMOVE | `backend/internal/handler/ai_handler.go` |
| REMOVE | `backend/internal/service/agent_service.go` |
| REMOVE | `backend/internal/service/ai_service.go` |
| REMOVE | `backend/internal/service/ai_service.go.bak` |
| REMOVE | `backend/internal/service/llm_client.go` |
| REMOVE | `backend/internal/model/ai_config.go` |
| REMOVE | `agent-service/internal/client/agent_client.go` |
