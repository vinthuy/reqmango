package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/service"
)

// AutomationHandler handles HTTP requests for automation rules.
type AutomationHandler struct {
	svc *service.AutomationService
}

func NewAutomationHandler(svc *service.AutomationService) *AutomationHandler {
	return &AutomationHandler{svc: svc}
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

	var req service.AutomationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	rule, svcErr := h.svc.Update(id, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, rule)
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
