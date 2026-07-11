package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type CycleHandler struct {
	svc *service.CycleService
}

func NewCycleHandler(svc *service.CycleService) *CycleHandler {
	return &CycleHandler{svc: svc}
}

func (h *CycleHandler) parseCycleID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("cycleId"), 10, 64)
}

func (h *CycleHandler) parseProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

// appError sends an AppError response if err is one. Returns true if an error was handled.
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

// Create handles POST /projects/:projectId/cycles?workspace_id=uint
func (h *CycleHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.CycleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	fmt.Println("CycleCreateRequest:", req.Name, "AutoAddEnabled:", req.AutoAddEnabled, "AutoAddRQL:", req.AutoAddRQL)

	projectID, err := h.parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	req.ProjectID = projectID

	resp, svcErr := h.svc.Create(workspaceID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List handles GET /projects/:projectId/cycles?status=&limit=&offset=
func (h *CycleHandler) List(c *gin.Context) {
	projectID, err := h.parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)
	status := c.Query("status")

	cycles, total, svcErr := h.svc.ListByProject(projectID, status, p.Limit, p.Offset)
	if appError(c, svcErr) {
		return
	}

	if cycles == nil {
		cycles = []response.CycleResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  cycles,
		"total":  total,
		"limit":  p.Limit,
		"offset": p.Offset,
	})
}

// Search handles GET /projects/:projectId/cycles/search?q=xxx
func (h *CycleHandler) Search(c *gin.Context) {
	projectID, err := h.parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query parameter 'q' is required"})
		return
	}
	cycles, svcErr := h.svc.Search(projectID, query)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cycles})
}

// Get handles GET /cycles/:cycleId
func (h *CycleHandler) Get(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetByID(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /cycles/:cycleId
func (h *CycleHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	var req request.CycleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(cycleID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /cycles/:cycleId
func (h *CycleHandler) Delete(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	if appError(c, h.svc.Delete(cycleID)) {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ==================== Status Transitions ====================

// Start handles POST /cycles/:cycleId/start
func (h *CycleHandler) Start(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.Start(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// End handles POST /cycles/:cycleId/end
func (h *CycleHandler) End(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.End(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Cancel handles POST /cycles/:cycleId/cancel
func (h *CycleHandler) Cancel(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.Cancel(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== Issue Association ====================

// AddIssue handles POST /cycles/:cycleId/issues?issue_id=uint
func (h *CycleHandler) AddIssue(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	issueID, err := strconv.ParseUint(c.Query("issue_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue_id"})
		return
	}

	if appError(c, h.svc.AddIssue(cycleID, issueID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cycle_id": cycleID,
		"issue_id": issueID,
		"action":   "added",
	})
}

// RemoveIssue handles DELETE /cycles/:cycleId/issues/:issueId
func (h *CycleHandler) RemoveIssue(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	if appError(c, h.svc.RemoveIssue(cycleID, issueID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cycle_id": cycleID,
		"issue_id": issueID,
		"action":   "removed",
	})
}

// ListIssues handles GET /cycles/:cycleId/issues?state_id=&priority=&limit=&offset=
func (h *CycleHandler) ListIssues(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)

	var stateID *uint64
	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			stateID = &id
		}
	}

	priority := c.Query("priority")

	issues, _, svcErr := h.svc.ListIssues(cycleID, stateID, priority, p.Limit, p.Offset)
	if appError(c, svcErr) {
		return
	}

	if issues == nil {
		issues = []response.IssueResponse{}
	}

	c.JSON(http.StatusOK, issues)
}

// ==================== Analysis ====================

// GetProgress handles GET /cycles/:cycleId/progress
func (h *CycleHandler) GetProgress(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetProgress(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStatistics handles GET /cycles/:cycleId/statistics
func (h *CycleHandler) GetStatistics(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetStatistics(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetBurndown handles GET /cycles/:cycleId/burndown
func (h *CycleHandler) GetBurndown(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetBurndown(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== Automation ====================

// ApplyAutoAddRules handles POST /cycles/:cycleId/apply-auto-add
func (h *CycleHandler) ApplyAutoAddRules(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	if appError(c, h.svc.ApplyAutoAddRules(cycleID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Auto-add rules applied successfully"})
}

// ApplyAutoCloseRules handles POST /cycles/:cycleId/apply-auto-close
func (h *CycleHandler) ApplyAutoCloseRules(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	if appError(c, h.svc.ApplyAutoCloseRules(cycleID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Auto-close rules applied successfully"})
}
