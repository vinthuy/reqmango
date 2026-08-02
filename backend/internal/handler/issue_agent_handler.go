package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

// IssueAgentHandler handles issue-agent assignment related requests.
type IssueAgentHandler struct {
	issueAgentSvc *service.IssueAgentService
}

// NewIssueAgentHandler creates a new IssueAgentHandler.
func NewIssueAgentHandler(issueAgentSvc *service.IssueAgentService) *IssueAgentHandler {
	return &IssueAgentHandler{issueAgentSvc: issueAgentSvc}
}

// AssignAgent assigns an issue to an agent.
func (h *IssueAgentHandler) AssignAgent(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue ID"})
		return
	}

	var req service.AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.issueAgentSvc.Assign(issueID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UnassignAgent removes agent assignment from an issue.
func (h *IssueAgentHandler) UnassignAgent(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue ID"})
		return
	}

	if err := h.issueAgentSvc.Unassign(issueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent unassigned"})
}

// GetAgentStatus returns the agent execution status for an issue.
func (h *IssueAgentHandler) GetAgentStatus(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue ID"})
		return
	}

	status, err := h.issueAgentSvc.GetStatus(issueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// PreviewExecution returns a preview of what the agent will do.
func (h *IssueAgentHandler) PreviewExecution(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue ID"})
		return
	}

	agentID, err := strconv.ParseUint(c.Query("agent_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent ID"})
		return
	}

	preview, err := h.issueAgentSvc.PreviewExecution(issueID, agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// BulkAssign assigns multiple issues to an agent.
func (h *IssueAgentHandler) BulkAssign(c *gin.Context) {
	var req struct {
		IssueIDs []uint64 `json:"issue_ids" binding:"required"`
		AgentID  uint64   `json:"agent_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tasks, err := h.issueAgentSvc.BulkAssign(req.IssueIDs, req.AgentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks})
}
