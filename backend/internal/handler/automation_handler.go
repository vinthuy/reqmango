package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

// AutomationHandler handles HTTP requests for automation rules.
type AutomationHandler struct {
	svc *service.AutomationService
	db  *gorm.DB
}

func NewAutomationHandler(svc *service.AutomationService, db *gorm.DB) *AutomationHandler {
	return &AutomationHandler{svc: svc, db: db}
}

func (h *AutomationHandler) resolveWorkspaceID(c *gin.Context) (uint64, error) {
	wsParam := c.Param("wsParam")
	if id, err := strconv.ParseUint(wsParam, 10, 64); err == nil {
		return id, nil
	}
	var workspace model.Workspace
	if err := h.db.Where("slug = ?", wsParam).First(&workspace).Error; err != nil {
		return 0, err
	}
	return workspace.ID, nil
}

// List handles GET /projects/:projectId/automations
func (h *AutomationHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	rules, svcErr := h.svc.List(projectID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rules)
}

// Get handles GET /projects/:projectId/automations/:id
func (h *AutomationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	rule, svcErr := h.svc.Get(id)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rule)
}

// Create handles POST /projects/:projectId/automations
func (h *AutomationHandler) Create(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var req service.AutomationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.Create(projectID, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// Update handles PUT /projects/:projectId/automations/:id
func (h *AutomationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var req service.AutomationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.Update(id, projectID, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rule)
}

// GetExecutionHistory handles GET /issues/:issueId/automation-history
func (h *AutomationHandler) GetExecutionHistory(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	history, total, svcErr := h.svc.GetExecutionHistory(issueID, limit, offset)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history, "total": total, "limit": limit, "offset": offset})
}

// GetRuleExecutionHistory handles GET /automations/:ruleId/execution-history
func (h *AutomationHandler) GetRuleExecutionHistory(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("ruleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid rule ID"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var startTime, endTime *time.Time
	if s := c.Query("start_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			startTime = &t
		}
	}
	if e := c.Query("end_time"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			endTime = &t
		}
	}

	history, total, svcErr := h.svc.GetRuleExecutionHistory(ruleID, limit, offset, startTime, endTime)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history, "total": total, "limit": limit, "offset": offset})
}

// GetProjectExecutionHistory handles GET /projects/:projectId/automation-executions
func (h *AutomationHandler) GetProjectExecutionHistory(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var startTime, endTime *time.Time
	if s := c.Query("start_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			startTime = &t
		}
	}
	if e := c.Query("end_time"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			endTime = &t
		}
	}

	history, total, svcErr := h.svc.GetProjectExecutionHistory(projectID, limit, offset, startTime, endTime)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history, "total": total, "limit": limit, "offset": offset})
}

// Delete handles DELETE /projects/:projectId/automations/:id
func (h *AutomationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	if svcErr := h.svc.Delete(id); appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Automation rule deleted"})
}

// Execute handles POST /projects/:projectId/automations/:id/execute (manual trigger for testing)
func (h *AutomationHandler) Execute(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	var req struct {
		IssueID uint64                 `json:"issue_id" binding:"required"`
		Context map[string]interface{} `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.Get(id)
	if svcErr != nil {
		if appError(c, svcErr) {
			return
		}
	}

	results := h.svc.ExecuteTrigger(projectID, rule.TriggerType, req.IssueID, req.Context)
	c.JSON(http.StatusOK, gin.H{"rule": rule, "results": results})
}

// ======== 工作区级自动化规则 API ========

// ListWorkspace handles GET /workspaces/:wsParam/automations
func (h *AutomationHandler) ListWorkspace(c *gin.Context) {
	workspaceID, err := h.resolveWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace identifier"})
		return
	}

	rules, svcErr := h.svc.ListWorkspace(workspaceID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rules)
}

// CreateWorkspace handles POST /workspaces/:wsParam/automations
func (h *AutomationHandler) CreateWorkspace(c *gin.Context) {
	workspaceID, err := h.resolveWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace identifier"})
		return
	}

	var req service.AutomationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.CreateWorkspace(workspaceID, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// UpdateWorkspace handles PUT /workspaces/:workspaceId/automations/:id
func (h *AutomationHandler) UpdateWorkspace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	var req service.AutomationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.UpdateWorkspace(id, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rule)
}

// DeleteWorkspace handles DELETE /workspaces/:workspaceId/automations/:id
func (h *AutomationHandler) DeleteWorkspace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid automation ID"})
		return
	}

	if svcErr := h.svc.DeleteWorkspace(id); appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Automation rule deleted"})
}
