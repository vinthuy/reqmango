package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

// BudgetHandler handles budget related requests.
type BudgetHandler struct {
	budgetSvc *service.AgentCostBudgetService
}

// NewBudgetHandler creates a new BudgetHandler.
func NewBudgetHandler(budgetSvc *service.AgentCostBudgetService) *BudgetHandler {
	return &BudgetHandler{budgetSvc: budgetSvc}
}

// Get returns the budget configuration for a project.
func (h *BudgetHandler) Get(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	budget, err := h.budgetSvc.Get(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// Update updates the budget configuration for a project.
func (h *BudgetHandler) Update(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req service.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.budgetSvc.Update(projectID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "budget updated"})
}

// SLAHandler handles SLA related requests.
type SLAHandler struct {
	slaSvc *service.AgentSLAService
}

// NewSLAHandler creates a new SLAHandler.
func NewSLAHandler(slaSvc *service.AgentSLAService) *SLAHandler {
	return &SLAHandler{slaSvc: slaSvc}
}

// Get returns the SLA configuration for a project.
func (h *SLAHandler) Get(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	sla, err := h.slaSvc.Get(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sla)
}

// Update updates the SLA configuration for a project.
func (h *SLAHandler) Update(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req service.UpdateSLARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.slaSvc.Update(projectID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SLA updated"})
}

// DecisionHandler handles decision related requests.
type DecisionHandler struct {
	decisionSvc *service.AgentDecisionService
}

// NewDecisionHandler creates a new DecisionHandler.
func NewDecisionHandler(decisionSvc *service.AgentDecisionService) *DecisionHandler {
	return &DecisionHandler{decisionSvc: decisionSvc}
}

// ListByIssue returns all decision records for an issue.
func (h *DecisionHandler) ListByIssue(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue ID"})
		return
	}

	decisions, err := h.decisionSvc.ListByIssue(issueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": decisions})
}

// ListByTask returns all decision records for a task.
func (h *DecisionHandler) ListByTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	decisions, err := h.decisionSvc.ListByTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": decisions})
}

// ListByProject returns all decision records for a project.
func (h *DecisionHandler) ListByProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	decisions, err := h.decisionSvc.ListByProject(projectID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": decisions, "total": len(decisions)})
}
