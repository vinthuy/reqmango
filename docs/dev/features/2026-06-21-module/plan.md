# Module（模块）管理功能 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Module CRUD with AppError pattern, ModuleIssue multi-many join table, progress/stats/tree API, Pinia store, side detail panel with add/remove issues, and inline form modal �?all integrated into Project tabs.

**Architecture:** Go backend (Gin + GORM) adds ModuleIssue join table + 6 new endpoints, refactors existing service/handler to use AppError. Vue 3 frontend adds Pinia store driving ModuleDetailPanel and ModuleFormModal, both within the Project tab layout (no new routes).

**Tech Stack:** Go + Gin + GORM + PostgreSQL (backend), Vue 3 + TypeScript + Pinia + Tailwind CSS (frontend)

**Spec reference:** `docs/pages/2026-06-21-module-feature-design.md`

---

## File Structure Map

### Go Backend
| File | Responsibility |
|------|---------------|
| `internal/model/module.go` | Module model + ModuleIssue join table |
| `internal/service/module_service.go` | CRUD + Issue association + progress/stats/tree |
| `internal/handler/module_handler.go` | HTTP handlers with AppError |
| `internal/router/router.go` | Route registration |

### Vue Frontend
| File | Responsibility |
|------|---------------|
| `stores/module.ts` | Pinia store |
| `components/ModuleDetailPanel.vue` | Side panel with issue management |
| `components/ModuleFormModal.vue` | Create/edit inline modal |
| `components/ModuleList.vue` | Store-driven list (card/tree) |
| `api/module.ts` | API client adjustments |
| `views/Project.vue` | Panel integration + event wiring |

---

### Task 1: Go Model �?ModuleIssue Join Table

**Files:**
- Modify: `backend/internal/model/module.go`

- [ ] **Step 1: Add ModuleIssue struct and update Module**

Replace `backend/internal/model/module.go`:

```go
package model

import "time"

// Module represents a project module (supports hierarchical grouping via ParentID).
type Module struct {
	BaseModel
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null;index" json:"workspace_id"`
	ParentID    *uint64    `gorm:"index" json:"parent_id"`
	Order       int        `gorm:"default:0" json:"order"`
	ArchivedAt  *time.Time `json:"archived_at"`
	IsArchived  bool       `gorm:"default:false" json:"is_archived"`

	// Relationships
	Project    Project        `gorm:"foreignKey:ProjectID" json:"-"`
	IssueLinks []ModuleIssue  `gorm:"foreignKey:ModuleID" json:"-"`
}

func (Module) TableName() string {
	return "modules"
}

// ModuleIssue is a join table for many-to-many module-issue association.
type ModuleIssue struct {
	ModuleID uint64 `gorm:"primaryKey;autoIncrement:false" json:"module_id"`
	IssueID  uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`

	Module Module `gorm:"foreignKey:ModuleID;constraint:OnDelete:CASCADE" json:"-"`
	Issue  Issue  `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ModuleIssue) TableName() string {
	return "module_issues"
}
```

- [ ] **Step 2: Register ModuleIssue in AutoMigrate**

Read `backend/cmd/server/main.go`, find the `AutoMigrate` call, add `&model.ModuleIssue{}` to the list.

- [ ] **Step 3: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/module.go backend/cmd/server/main.go
git commit -m "feat(module): add ModuleIssue join table for many-to-many association"
```

---

### Task 2: Go ModuleService �?Refactor + New Methods

**Files:**
- Modify: `backend/internal/service/module_service.go`

- [ ] **Step 1: Rewrite ModuleService**

Replace the entire content of `backend/internal/service/module_service.go`:

```go
package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ModuleService struct {
	db *gorm.DB
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{db: db}
}

// ==================== Helpers ====================

func (s *ModuleService) buildResponse(m model.Module) *response.ModuleResponse {
	return &response.ModuleResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		ProjectID:   m.ProjectID,
		WorkspaceID: m.WorkspaceID,
		ParentID:    m.ParentID,
		Order:       m.Order,
		IsArchived:  m.IsArchived,
		ArchivedAt:  m.ArchivedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (s *ModuleService) countIssues(moduleID uint64) (total, completed int64, err error) {
	if err = s.db.Model(&model.ModuleIssue{}).Where("module_id = ?", moduleID).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group = ?", moduleID, common.StateGroupCompleted).
		Count(&completed).Error; err != nil {
		return total, 0, err
	}
	return total, completed, nil
}

// ==================== CRUD ====================

func (s *ModuleService) Create(workspaceID, userID uint64, req request.ModuleCreate) (*response.ModuleResponse, error) {
	var project model.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	module := model.Module{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		WorkspaceID: workspaceID,
		ParentID:    req.ParentID,
	}
	module.CreatedByID = &userID

	if err := s.db.Create(&module).Error; err != nil {
		return nil, common.Internal("Failed to create module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) List(projectID, workspaceID uint64, includeArchived bool) ([]response.ModuleResponse, int64, error) {
	query := s.db.Model(&model.Module{}).Where("project_id = ? AND workspace_id = ?", projectID, workspaceID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Failed to count modules")
	}

	var modules []model.Module
	if err := query.Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, 0, common.Internal("Failed to list modules")
	}

	result := make([]response.ModuleResponse, len(modules))
	for i, m := range modules {
		result[i] = *s.buildResponse(m)
	}

	return result, total, nil
}

func (s *ModuleService) Get(moduleID uint64) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}
	return s.buildResponse(module), nil
}

func (s *ModuleService) Update(moduleID, userID uint64, req request.ModuleUpdate) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	if req.Name != nil {
		module.Name = *req.Name
	}
	if req.Description != nil {
		module.Description = *req.Description
	}
	if req.ParentID != nil {
		module.ParentID = req.ParentID
	}

	module.UpdatedByID = &userID
	if err := s.db.Save(&module).Error; err != nil {
		return nil, common.Internal("Failed to update module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) Delete(moduleID uint64) error {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return common.NotFound("Module not found")
	}
	s.db.Where("module_id = ?", moduleID).Delete(&model.ModuleIssue{})
	return s.db.Delete(&module).Error
}

// ==================== Issue Association ====================

func (s *ModuleService) AddIssue(moduleID, issueID uint64) error {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return common.NotFound("Module not found")
	}

	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if issue.ProjectID != module.ProjectID {
		return common.BadRequest("Issue does not belong to this module's project")
	}

	var count int64
	s.db.Model(&model.ModuleIssue{}).Where("module_id = ? AND issue_id = ?", moduleID, issueID).Count(&count)
	if count > 0 {
		return common.BadRequest("Issue is already in this module")
	}

	if err := s.db.Create(&model.ModuleIssue{ModuleID: moduleID, IssueID: issueID}).Error; err != nil {
		return common.Internal("Failed to add issue to module")
	}

	return nil
}

func (s *ModuleService) RemoveIssue(moduleID, issueID uint64) error {
	result := s.db.Where("module_id = ? AND issue_id = ?", moduleID, issueID).Delete(&model.ModuleIssue{})
	if result.RowsAffected == 0 {
		return common.NotFound("Issue is not in this module")
	}
	return nil
}

func (s *ModuleService) ListIssues(moduleID uint64, stateID *uint64, priority string, limit, offset int) ([]response.IssueResponse, int64, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, 0, common.NotFound("Module not found")
	}

	baseQuery := s.db.Model(&model.Issue{}).
		Joins("JOIN module_issues ON module_issues.issue_id = issues.id").
		Where("module_issues.module_id = ?", moduleID)

	if stateID != nil {
		baseQuery = baseQuery.Where("issues.state_id = ?", *stateID)
	}
	if priority != "" {
		baseQuery = baseQuery.Where("issues.priority = ?", priority)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var issues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("Project").
		Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").
		Order("issues.sort_order ASC, issues.sequence_id DESC").
		Limit(limit).Offset(offset).
		Find(&issues).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	issueSvc := &IssueService{db: s.db}
	result := make([]response.IssueResponse, len(issues))
	for i, issue := range issues {
		resp, err := issueSvc.BuildIssueResponse(&issue)
		if err != nil {
			return nil, 0, err
		}
		result[i] = *resp
	}

	return result, total, nil
}

// ==================== Progress & Statistics ====================

func (s *ModuleService) GetProgress(moduleID uint64) (*response.ModuleProgress, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	total, completed, _ := s.countIssues(moduleID)
	progress := 0
	if total > 0 {
		progress = int(float64(completed) / float64(total) * 100)
	}

	return &response.ModuleProgress{
		ModuleID:    moduleID,
		ModuleName:  module.Name,
		TotalIssues: total,
		Completed:   completed,
		Progress:    progress,
	}, nil
}

func (s *ModuleService) GetStatistics(moduleID uint64) (*response.ModuleStatistics, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	total, completed, _ := s.countIssues(moduleID)

	// Active
	var active int64
	s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group IN ?", moduleID, []string{common.StateGroupBacklog, common.StateGroupUnstarted, common.StateGroupStarted}).
		Count(&active)

	// Cancelled
	var cancelled int64
	s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group = ?", moduleID, common.StateGroupCancelled).
		Count(&cancelled)

	// By priority
	type priorityRow struct {
		Priority string
		Count    int64
	}
	var pRows []priorityRow
	s.db.Table("module_issues").
		Select("issues.priority, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Where("module_issues.module_id = ?", moduleID).
		Group("issues.priority").
		Scan(&pRows)
	byPriority := make(map[string]int64)
	for _, r := range pRows {
		byPriority[r.Priority] = r.Count
	}

	// By state
	type stateRow struct {
		Name  string
		Count int64
	}
	var sRows []stateRow
	s.db.Table("module_issues").
		Select("states.name, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ?", moduleID).
		Group("states.name").
		Scan(&sRows)
	byState := make(map[string]int64)
	for _, r := range sRows {
		byState[r.Name] = r.Count
	}

	return &response.ModuleStatistics{
		ModuleID:     moduleID,
		ModuleName:   module.Name,
		TotalIssues:  total,
		ActiveIssues: active,
		Completed:    completed,
		Cancelled:    cancelled,
		ByPriority:   byPriority,
		ByState:      byState,
	}, nil
}

// ==================== Tree ====================

func (s *ModuleService) BuildTree(projectID uint64) ([]*response.ModuleTreeNode, error) {
	var modules []model.Module
	if err := s.db.Where("project_id = ? AND is_archived = ?", projectID, false).
		Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to load modules")
	}

	// Group children by parent
	childrenMap := make(map[uint64][]*response.ModuleTreeNode)
	var roots []*response.ModuleTreeNode

	for _, m := range modules {
		total, completed, _ := s.countIssues(m.ID)
		progress := 0
		if total > 0 {
			progress = int(float64(completed) / float64(total) * 100)
		}

		node := &response.ModuleTreeNode{
			ModuleResponse: *s.buildResponse(m),
			Children:       nil,
			TotalIssues:    total,
			CompletedIssues: completed,
			Progress:       progress,
		}

		if m.ParentID == nil {
			roots = append(roots, node)
		} else {
			childrenMap[*m.ParentID] = append(childrenMap[*m.ParentID], node)
		}
	}

	// Recursively attach children
	var attachChildren func(nodes []*response.ModuleTreeNode)
	attachChildren = func(nodes []*response.ModuleTreeNode) {
		for _, n := range nodes {
			n.Children = childrenMap[n.ID]
			if n.Children == nil {
				n.Children = []*response.ModuleTreeNode{}
			}
			attachChildren(n.Children)
		}
	}
	attachChildren(roots)

	if roots == nil {
		roots = []*response.ModuleTreeNode{}
	}
	return roots, nil
}
```

- [ ] **Step 2: Update ModuleUpdate DTO for pointer fields**

Read `backend/internal/dto/request/module.go`. Change `ModuleUpdate` to use pointer fields:

```go
type ModuleUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *uint64 `json:"parent_id"`
}
```

- [ ] **Step 3: Add TotalIssues/CompletedIssues/Progress to ModuleTreeNode DTO**

Read `backend/internal/dto/response/module.go`. Update `ModuleTreeNode`:

```go
type ModuleTreeNode struct {
	ModuleResponse
	Children         []*ModuleTreeNode `json:"children"`
	TotalIssues      int64             `json:"total_issues"`
	CompletedIssues  int64             `json:"completed_issues"`
	Progress         int               `json:"progress"`
}
```

- [ ] **Step 4: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/module_service.go backend/internal/dto/request/module.go backend/internal/dto/response/module.go
git commit -m "feat(module): rewrite ModuleService with AppError, add issue/progress/stats/tree methods"
```

---

### Task 3: Go ModuleHandler �?Refactor + New Handlers

**Files:**
- Modify: `backend/internal/handler/module_handler.go`

- [ ] **Step 1: Rewrite ModuleHandler**

Replace the entire content of `backend/internal/handler/module_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type ModuleHandler struct {
	svc *service.ModuleService
}

func NewModuleHandler(svc *service.ModuleService) *ModuleHandler {
	return &ModuleHandler{svc: svc}
}

func (h *ModuleHandler) parseModuleID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("moduleId"), 10, 64)
}

func appError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	return true
}

// ==================== CRUD ====================

func (h *ModuleHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.ModuleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Create(workspaceID, user.ID, req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *ModuleHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}
	includeArchived := c.DefaultQuery("include_archived", "false") == "true"

	modules, _, svcErr := h.svc.List(projectID, workspaceID, includeArchived)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, modules)
}

func (h *ModuleHandler) Get(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	resp, svcErr := h.svc.Get(moduleID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	var req request.ModuleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(moduleID, user.ID, req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) Delete(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	if appError(c, h.svc.Delete(moduleID)) {
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ==================== Issue Association ====================

func (h *ModuleHandler) AddIssue(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	issueID, err := strconv.ParseUint(c.Query("issue_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue_id"})
		return
	}
	if appError(c, h.svc.AddIssue(moduleID, issueID)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_id": moduleID, "issue_id": issueID, "action": "added"})
}

func (h *ModuleHandler) RemoveIssue(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}
	if appError(c, h.svc.RemoveIssue(moduleID, issueID)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_id": moduleID, "issue_id": issueID, "action": "removed"})
}

func (h *ModuleHandler) ListIssues(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)
	var stateID *uint64
	if v := c.Query("state_id"); v != "" {
		if id, e := strconv.ParseUint(v, 10, 64); e == nil {
			stateID = &id
		}
	}
	issues, _, svcErr := h.svc.ListIssues(moduleID, stateID, c.Query("priority"), p.Limit, p.Offset)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, issues)
}

// ==================== Analysis ====================

func (h *ModuleHandler) GetProgress(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	resp, svcErr := h.svc.GetProgress(moduleID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) GetStatistics(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	resp, svcErr := h.svc.GetStatistics(moduleID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ModuleHandler) GetTree(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}
	resp, svcErr := h.svc.BuildTree(projectID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly (note: handler's `appError` is package-level, won't conflict with cycle_handler's)

Note: Both `cycle_handler.go` and `module_handler.go` define `appError` at package level. This WILL cause a compilation error. Fix: rename the cycle_handler's `appError` to `cycleAppError` or move it to a shared helper.

- [ ] **Step 3: Fix appError conflict**

In `backend/internal/handler/cycle_handler.go`, rename `func appError` �?`func cycleAppError` and update all call sites.

OR: delete the standalone `appError` from both handlers, create a shared helper file `backend/internal/handler/helpers.go`:

```go
package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
)

func appError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	return true
}
```

Then remove `appError` from both `cycle_handler.go` and `module_handler.go`.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/
git commit -m "feat(module): rewrite ModuleHandler with AppError, add issue/progress/stats/tree endpoints"
```

---

### Task 4: Go Router �?Register New Module Routes

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Add new routes to modules group**

In `backend/internal/router/router.go`, update the modules group to add 6 new routes:

```go
		modules := v1.Group("/modules", authMiddleware)
		{
			// CRUD
			modules.POST("", moduleH.Create)                      // ?workspace_id=
			modules.GET("", moduleH.List)                         // ?project_id=&workspace_id=&include_archived=
			modules.GET("/:moduleId", moduleH.Get)
			modules.PUT("/:moduleId", moduleH.Update)
			modules.DELETE("/:moduleId", moduleH.Delete)

			// Issues
			modules.POST("/:moduleId/issues", moduleH.AddIssue)         // ?issue_id=
			modules.DELETE("/:moduleId/issues/:issueId", moduleH.RemoveIssue)
			modules.GET("/:moduleId/issues", moduleH.ListIssues)

			// Analysis
			modules.GET("/:moduleId/progress", moduleH.GetProgress)
			modules.GET("/:moduleId/statistics", moduleH.GetStatistics)

			// Tree
			modules.GET("/tree", moduleH.GetTree)                       // ?project_id=
		}
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 3: Commit**

```bash
git add backend/internal/router/router.go
git commit -m "feat(module): register issue/progress/stats/tree routes"
```

---

### Task 5: Frontend �?Pinia Store

**Files:**
- Create: `frontend/src/stores/module.ts`

- [ ] **Step 1: Create Module Pinia store**

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import moduleApi from '@/api/module'
import type { ModuleResponse, ModuleCreate, ModuleUpdate, ModuleProgress, ModuleTreeNode } from '@/types/module'

export const useModuleStore = defineStore('module', () => {
  const modules = ref<ModuleResponse[]>([])
  const moduleTree = ref<ModuleTreeNode[]>([])
  const currentModule = ref<ModuleResponse | null>(null)
  const progress = ref<ModuleProgress | null>(null)
  const moduleIssues = ref<any[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function fetchModules(projectId: number, workspaceId: number) {
    isLoading.value = true
    error.value = null
    try {
      modules.value = await moduleApi.listModules(projectId, workspaceId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  async function fetchModuleTree(projectId: number) {
    try {
      moduleTree.value = await moduleApi.getModuleTree(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function createModule(workspaceId: number, data: ModuleCreate) {
    error.value = null
    try {
      const created = await moduleApi.createModule(workspaceId, data)
      modules.value.unshift(created)
      return created
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function updateModuleAction(id: number, data: ModuleUpdate) {
    error.value = null
    try {
      const updated = await moduleApi.updateModule(id, data)
      const idx = modules.value.findIndex(m => m.id === id)
      if (idx !== -1) modules.value[idx] = updated
      if (currentModule.value?.id === id) currentModule.value = updated
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function deleteModuleAction(id: number) {
    error.value = null
    try {
      await moduleApi.deleteModule(id)
      modules.value = modules.value.filter(m => m.id !== id)
      if (currentModule.value?.id === id) currentModule.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function addIssueToModule(moduleId: number, issueId: number) {
    error.value = null
    try {
      const result = await moduleApi.addIssueToModule(moduleId, issueId)
      await fetchModuleIssues(moduleId)
      await fetchProgress(moduleId)
      return result
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function removeIssueFromModule(moduleId: number, issueId: number) {
    error.value = null
    try {
      await moduleApi.removeIssueFromModule(moduleId, issueId)
      moduleIssues.value = moduleIssues.value.filter((i: any) => i.id !== issueId)
      await fetchProgress(moduleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchModuleIssues(moduleId: number, filters?: { state_id?: number; priority?: string }) {
    isLoading.value = true
    try {
      moduleIssues.value = await moduleApi.getModuleIssues(moduleId, filters)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  async function fetchProgress(moduleId: number) {
    try {
      progress.value = await moduleApi.getModuleProgress(moduleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  return {
    modules, moduleTree, currentModule, progress, moduleIssues, isLoading, error,
    fetchModules, fetchModuleTree,
    createModule, updateModuleAction, deleteModuleAction,
    addIssueToModule, removeIssueFromModule, fetchModuleIssues,
    fetchProgress,
  }
})
```

- [ ] **Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`
Ignore pre-existing errors in unrelated files.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/module.ts
git commit -m "feat(module): add Pinia store for module management"
```

---

### Task 6: Frontend �?Update ModuleList to use Store

**Files:**
- Modify: `frontend/src/components/ModuleList.vue`

- [ ] **Step 1: Update ModuleList to use store**

In the `<script setup>`, replace manual API calls with store:

```typescript
import { ref, computed, onMounted } from 'vue'
import { useModuleStore } from '@/stores/module'
import ModuleCard from './ModuleCard.vue'
import ModuleTree from './ModuleTree.vue'
import type { ModuleResponse, ModuleTreeNode } from '@/types/module'

const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  select: [module: ModuleResponse | ModuleTreeNode]
  delete: [module: ModuleResponse | ModuleTreeNode]
  create: []
}>()

const moduleStore = useModuleStore()
const viewMode = ref<'card' | 'tree'>('card')

const modules = computed(() => moduleStore.modules)
const moduleTree = computed(() => moduleStore.moduleTree)
const loading = computed(() => moduleStore.isLoading)

onMounted(async () => {
  await Promise.all([
    moduleStore.fetchModules(props.projectId, props.workspaceId),
    moduleStore.fetchModuleTree(props.projectId),
  ])
})
```

Remove the local `modules`, `moduleTree`, `loading` refs and the `loadModules` function.

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ModuleList.vue
git commit -m "feat(module): update ModuleList to use Pinia store"
```

---

### Task 7: Frontend �?ModuleFormModal

**Files:**
- Create: `frontend/src/components/ModuleFormModal.vue`

- [ ] **Step 1: Create ModuleFormModal**

```vue
<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 bg-black bg-opacity-30 z-50 flex items-center justify-center" @click.self="$emit('close')">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ isEdit ? '编辑模块' : '新建模块' }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">名称 <span class="text-red-500">*</span></label>
            <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="模块名称" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">描述</label>
            <textarea v-model="form.description" rows="3" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="描述..."></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">父模�?/label>
            <select v-model="form.parent_id" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500">
              <option :value="undefined">无（顶级模块�?/option>
              <option v-for="m in moduleStore.modules" :key="m.id" :value="m.id" :disabled="m.id === editModule?.id">
                {{ m.name }}
              </option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="$emit('close')" class="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">取消</button>
          <button @click="handleSubmit" :disabled="submitting" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700 disabled:opacity-50">
            {{ submitting ? '保存�?..' : (isEdit ? '保存' : '创建') }}
          </button>
        </div>
        <div v-if="moduleStore.error" class="mt-3 text-sm text-red-600">{{ moduleStore.error }}</div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useModuleStore } from '@/stores/module'
import type { ModuleResponse } from '@/types/module'

const props = defineProps<{
  visible: boolean
  editModule: ModuleResponse | null
  workspaceId: number
  projectId: number
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const moduleStore = useModuleStore()
const isEdit = computed(() => !!props.editModule)
const submitting = ref(false)

const form = ref({
  name: '',
  description: '',
  parent_id: undefined as number | undefined,
})

watch(() => props.visible, (v) => {
  if (v) {
    if (props.editModule) {
      form.value.name = props.editModule.name
      form.value.description = props.editModule.description || ''
      form.value.parent_id = props.editModule.parent_id as number | undefined
    } else {
      form.value.name = ''
      form.value.description = ''
      form.value.parent_id = undefined
    }
  }
})

async function handleSubmit() {
  if (!form.value.name.trim()) return
  submitting.value = true

  const data = {
    name: form.value.name,
    description: form.value.description,
    project_id: props.projectId,
    workspace_id: props.workspaceId,
    parent_id: form.value.parent_id || undefined,
  }

  if (isEdit.value) {
    const result = await moduleStore.updateModuleAction(props.editModule!.id, {
      name: form.value.name,
      description: form.value.description,
      parent_id: form.value.parent_id || undefined,
    } as any)
    if (!result) { submitting.value = false; return }
  } else {
    const result = await moduleStore.createModule(props.workspaceId, data as any)
    if (!result) { submitting.value = false; return }
  }

  submitting.value = false
  emit('saved')
  emit('close')
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ModuleFormModal.vue
git commit -m "feat(module): add ModuleFormModal for create/edit"
```

---

### Task 8: Frontend �?ModuleDetailPanel + Project.vue Integration

**Files:**
- Create: `frontend/src/components/ModuleDetailPanel.vue`
- Modify: `frontend/src/views/Project.vue`

- [ ] **Step 1: Create ModuleDetailPanel.vue**

```vue
<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between z-10">
        <h3 class="text-lg font-semibold text-gray-900 truncate">{{ module?.name }}</h3>
        <div class="flex items-center space-x-1">
          <button @click="$emit('edit', module)" class="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50">编辑</button>
          <button @click="handleDelete" class="px-2 py-1 text-xs border border-red-300 text-red-600 rounded hover:bg-red-50">删除</button>
          <button @click="$emit('close')" class="p-1 text-gray-400 hover:text-gray-600 ml-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <svg class="animate-spin h-6 w-6 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <div v-else-if="module" class="p-4 space-y-4">
        <p v-if="module.description" class="text-sm text-gray-500">{{ module.description }}</p>

        <!-- Progress -->
        <div v-if="moduleStore.progress" class="grid grid-cols-3 gap-3">
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-gray-900">{{ moduleStore.progress.total_issues }}</div>
            <div class="text-xs text-gray-500">总数</div>
          </div>
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-green-600">{{ moduleStore.progress.completed }}</div>
            <div class="text-xs text-gray-500">完成</div>
          </div>
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-indigo-600">{{ moduleStore.progress.progress }}%</div>
            <div class="text-xs text-gray-500">进度</div>
          </div>
        </div>

        <!-- Issues -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-medium text-gray-700">模块工作�?({{ moduleStore.moduleIssues.length }})</h4>
            <button @click="toggleAddIssue" class="px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700">+ 添加</button>
          </div>

          <!-- Add issue search -->
          <div v-if="showAddIssue" class="mb-3 border border-gray-200 rounded-md p-2">
            <input v-model="searchQuery" type="text" placeholder="搜索工作�?.."
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded mb-2 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              @input="searchIssues" />
            <div class="max-h-40 overflow-y-auto space-y-1">
              <div v-for="issue in availableIssues" :key="issue.id" @click="handleAddIssue(issue.id)"
                class="flex items-center p-1.5 hover:bg-indigo-50 rounded cursor-pointer text-sm">
                <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
                <span class="text-xs text-gray-400 ml-2">#{{ issue.sequence_id }}</span>
              </div>
              <div v-if="availableIssues.length === 0 && searched" class="text-xs text-gray-400 py-2 text-center">没有可添加的工作�?/div>
            </div>
          </div>

          <div v-if="moduleStore.moduleIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无工作�?/div>
          <div v-else class="space-y-2">
            <div v-for="issue in moduleStore.moduleIssues" :key="issue.id" class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
              <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
              <button @click="handleRemoveIssue(issue.id)" class="ml-2 text-gray-400 hover:text-red-500" title="移除">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useModuleStore } from '@/stores/module'
import { issueApi } from '@/api/issue'
import type { ModuleResponse } from '@/types/module'

const props = defineProps<{
  module: ModuleResponse | null
  visible: boolean
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  close: []
  edit: [module: ModuleResponse]
}>()

const moduleStore = useModuleStore()
const loading = computed(() => moduleStore.isLoading)

const showAddIssue = ref(false)
const searchQuery = ref('')
const searched = ref(false)
const availableIssues = ref<any[]>([])

watch(() => props.visible, async (v) => {
  if (v && props.module) {
    showAddIssue.value = false
    await Promise.all([
      moduleStore.fetchProgress(props.module.id),
      moduleStore.fetchModuleIssues(props.module.id),
    ])
  }
})

function toggleAddIssue() {
  showAddIssue.value = !showAddIssue.value
  if (showAddIssue.value) {
    searchQuery.value = ''
    searchIssues()
  }
}

async function searchIssues() {
  if (!props.module) return
  try {
    const result = await issueApi.listIssues(props.projectId, props.workspaceId, {
      search: searchQuery.value || undefined,
    })
    const allIssues = (result as any)?.items || result || []
    const currentIds = new Set(moduleStore.moduleIssues.map((i: any) => i.id))
    availableIssues.value = allIssues.filter(
      (i: any) => !currentIds.has(i.id) && i.state_group !== 'completed' && i.state_group !== 'cancelled'
    )
    searched.value = true
  } catch {
    availableIssues.value = []
    searched.value = true
  }
}

async function handleAddIssue(issueId: number) {
  if (!props.module) return
  await moduleStore.addIssueToModule(props.module.id, issueId)
  showAddIssue.value = false
}

async function handleRemoveIssue(issueId: number) {
  if (!props.module) return
  await moduleStore.removeIssueFromModule(props.module.id, issueId)
}

async function handleDelete() {
  if (!props.module) return
  if (!confirm(`确定要删除模�?"${props.module.name}" 吗？此操作不可撤销。`)) return
  await moduleStore.deleteModuleAction(props.module.id)
  emit('close')
}
</script>

<style scoped>
.slide-enter-active, .slide-leave-active { transition: transform 0.3s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
</style>
```

- [ ] **Step 2: Update Project.vue to integrate ModuleDetailPanel and ModuleFormModal**

In `frontend/src/views/Project.vue`:

Add imports:
```typescript
import ModuleDetailPanel from '@/components/ModuleDetailPanel.vue'
import ModuleFormModal from '@/components/ModuleFormModal.vue'
import type { ModuleResponse } from '@/types/module'
```

Add state:
```typescript
const selectedModule = ref<ModuleResponse | null>(null)
const modulePanelVisible = ref(false)
const moduleFormVisible = ref(false)
const editingModule = ref<ModuleResponse | null>(null)
```

Add handlers:
```typescript
function openModulePanel(module: ModuleResponse | any) {
  selectedModule.value = module as ModuleResponse
  modulePanelVisible.value = true
}

function handleModuleEdit(module: ModuleResponse) {
  editingModule.value = module
  moduleFormVisible.value = true
}

function handleModuleDelete(module: ModuleResponse | any) {
  const m = module as ModuleResponse
  if (confirm(`确定要删除模�?"${m.name}" 吗？`)) {
    // Will be handled by the detail panel
  }
}
```

Update the template's ModuleList to wire events:
```vue
<ModuleList
  :project-id="projectId"
  :workspace-id="workspaceId"
  @select="openModulePanel"
  @create="moduleFormVisible = true"
  @delete="handleModuleDelete"
/>
```

Add panel and modal after CycleDetailPanel:
```vue
      <ModuleDetailPanel
        :module="selectedModule"
        :visible="modulePanelVisible"
        :project-id="projectId"
        :workspace-id="workspaceId"
        @close="modulePanelVisible = false"
        @edit="handleModuleEdit"
      />

      <ModuleFormModal
        :visible="moduleFormVisible"
        :edit-module="editingModule"
        :workspace-id="workspaceId"
        :project-id="projectId"
        @close="moduleFormVisible = false; editingModule = null"
        @saved="moduleFormVisible = false; editingModule = null"
      />
```

Update the `watch(activeTab)` to also close modulePanelVisible:
```typescript
watch(activeTab, () => {
  detailPanelVisible.value = false
  cyclePanelVisible.value = false
  modulePanelVisible.value = false
})
```

- [ ] **Step 3: Verify**

Run: `cd frontend && npx vue-tsc --noEmit`
Ignore pre-existing errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ModuleDetailPanel.vue frontend/src/views/Project.vue
git commit -m "feat(module): add ModuleDetailPanel with issue management, integrate into Project tab"
```

---

### Task 9: End-to-End Smoke Test

- [ ] **Step 1: Rebuild Go backend**

Run: `cd backend && go build -o server.exe ./cmd/server/`
Expected: Build OK

- [ ] **Step 2: Restart services and test API**

```bash
# Restart backend
taskkill /f /im server.exe
cd backend && ./server.exe &

# Test module tree
curl -s "http://localhost:8000/api/v1/modules/tree?project_id=1" -H "Authorization: Bearer <token>"

# Test add issue to module
curl -s -X POST "http://localhost:8000/api/v1/modules/1/issues?issue_id=1" -H "Authorization: Bearer <token>"

# Test progress
curl -s "http://localhost:8000/api/v1/modules/1/progress" -H "Authorization: Bearer <token>"
```

- [ ] **Step 3: Run frontend type check**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: No new errors

- [ ] **Step 4: Commit any fixes**

```bash
git add -A && git commit -m "fix(module): smoke test fixes"
```

---

## Plan Summary

| Task | Area | Files | Steps |
|------|------|-------|-------|
| 1 | Go Model | 2 modify | 4 |
| 2 | Go Service | 3 modify | 5 |
| 3 | Go Handler + helpers | 3 modify/create | 4 |
| 4 | Go Router | 1 modify | 3 |
| 5 | Frontend Store | 1 create | 3 |
| 6 | Frontend ModuleList | 1 modify | 2 |
| 7 | Frontend Modal | 1 create | 2 |
| 8 | Frontend Panel + Project | 2 create/modify | 4 |
| 9 | Smoke Test | all | 4 |

**Total: 9 tasks, ~31 steps**
