package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/service"
)

type IssueHandler struct {
	svc *service.IssueService
}

func NewIssueHandler(svc *service.IssueService) *IssueHandler {
	return &IssueHandler{svc: svc}
}

func (h *IssueHandler) parseIssueID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("issueId"), 10, 64)
}

// ==================== CRUD ====================

// Create handles POST /issues/?project_id=int&workspace_id=int
func (h *IssueHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

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

	var req request.IssueCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Create(&req, projectID, workspaceID, user.ID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List handles GET /issues/?project_id=int&workspace_id=int&filters...
func (h *IssueHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)

	filters := make(map[string]interface{})

	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filters["state_id"] = id
		}
	}
	if v := c.Query("priority"); v != "" {
		filters["priority"] = v
	}
	if v := c.Query("assignee_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filters["assignee_id"] = id
		}
	}
	if v := c.Query("parent_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filters["parent_id"] = id
		}
	}
	if v := c.Query("cycle_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filters["cycle_id"] = id
		}
	}
	if v := c.Query("module_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filters["module_id"] = id
		}
	}
	if v := c.Query("search"); v != "" {
		filters["search"] = v
	}
	if v := c.Query("is_draft"); v != "" {
		filters["is_draft"] = v == "true"
	}

	issues, total, svcErr := h.svc.List(projectID, filters, p.Limit, p.Offset)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	if issues == nil {
		issues = make([]response.IssueResponse, 0)
	}

	c.Header("X-Total-Count", fmt.Sprintf("%d", total))
	c.JSON(http.StatusOK, issues)
}

// Get handles GET /issues/:id
func (h *IssueHandler) Get(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	resp, svcErr := h.svc.GetByID(issueID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /issues/:id
func (h *IssueHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	var req request.IssueUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(issueID, &req, user.ID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /issues/:id
func (h *IssueHandler) Delete(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	if svcErr := h.svc.Delete(issueID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Archive handles POST /issues/:id/archive
func (h *IssueHandler) Archive(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	if svcErr := h.svc.Archive(issueID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Restore handles POST /issues/:id/restore
func (h *IssueHandler) Restore(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	resp, svcErr := h.svc.Restore(issueID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== Activities ====================

// GetActivities handles GET /issues/:id/activities?limit=&offset=
func (h *IssueHandler) GetActivities(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)

	activities, svcErr := h.svc.GetActivities(issueID, p.Limit, p.Offset)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// ==================== Statistics & Search ====================

// GetStatistics handles GET /issues/statistics?project_id=int
func (h *IssueHandler) GetStatistics(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	stats, svcErr := h.svc.GetStatistics(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Search handles GET /issues/search?workspace_id=int&query=str&project_id=&limit=
func (h *IssueHandler) Search(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query is required"})
		return
	}

	var projectID *uint64
	if v := c.Query("project_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			projectID = &id
		}
	}

	p := common.ParsePagination(c.Query("limit"), "", 10, 50)

	results, svcErr := h.svc.Search(workspaceID, query, projectID, p.Limit)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ==================== Bulk Operations ====================

// BulkUpdate handles POST /issues/bulk/update?project_id=int
func (h *IssueHandler) BulkUpdate(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	var req request.BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	results, svcErr := h.svc.BulkUpdate(projectID, &req, user.ID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// BulkDelete handles POST /issues/bulk/delete
func (h *IssueHandler) BulkDelete(c *gin.Context) {
	var req request.BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if svcErr := h.svc.BulkDelete(req.IssueIDs); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ==================== Assignee Management ====================

// AddAssignee handles POST /issues/:id/assignees?user_id=int
func (h *IssueHandler) AddAssignee(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user_id"})
		return
	}

	if svcErr := h.svc.AddAssignee(issueID, userID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"user_id":  userID,
		"action":   "added",
	})
}

// RemoveAssignee handles DELETE /issues/:id/assignees/:userId
func (h *IssueHandler) RemoveAssignee(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	if svcErr := h.svc.RemoveAssignee(issueID, userID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"user_id":  userID,
		"action":   "removed",
	})
}

// ==================== Label Management ====================

// AddLabel handles POST /issues/:id/labels?label_id=int
func (h *IssueHandler) AddLabel(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Query("label_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label_id"})
		return
	}

	if svcErr := h.svc.AddLabel(issueID, labelID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"label_id": labelID,
		"action":   "added",
	})
}

// RemoveLabel handles DELETE /issues/:id/labels/:labelId
func (h *IssueHandler) RemoveLabel(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	if svcErr := h.svc.RemoveLabel(issueID, labelID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"label_id": labelID,
		"action":   "removed",
	})
}

// ==================== Cycle Management ====================

// SetCycle handles POST /issues/:id/cycle?cycle_id=int
func (h *IssueHandler) SetCycle(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	cycleID, err := strconv.ParseUint(c.Query("cycle_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle_id"})
		return
	}

	if svcErr := h.svc.SetCycle(issueID, cycleID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"cycle_id": cycleID,
		"action":   "set",
	})
}

// RemoveCycle handles DELETE /issues/:id/cycle
func (h *IssueHandler) RemoveCycle(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)

	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	if svcErr := h.svc.RemoveCycle(issueID, actor.ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue_id": issueID,
		"cycle_id": nil,
		"action":   "removed",
	})
}
