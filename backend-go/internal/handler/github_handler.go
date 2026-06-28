package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/service"
)

type GitHubHandler struct {
	svc *service.GitHubService
}

func NewGitHubHandler(svc *service.GitHubService) *GitHubHandler {
	return &GitHubHandler{svc: svc}
}

// List handles GET /workspaces/:workspaceId/github
func (h *GitHubHandler) List(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	result, err := h.svc.List(workspaceID)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Get handles GET /workspaces/:workspaceId/github/:id
func (h *GitHubHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := h.svc.Get(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Create handles POST /workspaces/:workspaceId/github
func (h *GitHubHandler) Create(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	var req service.GitHubCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Create(workspaceID, &req)
	if appError(c, err) {
		return
	}
	c.JSON(201, result)
}

// Update handles PUT /workspaces/:workspaceId/github/:id
func (h *GitHubHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.GitHubUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Update(id, &req)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Delete handles DELETE /workspaces/:workspaceId/github/:id
func (h *GitHubHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		appError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "deleted"})
}

// SyncIssues handles POST /workspaces/:workspaceId/github/:id/sync
func (h *GitHubHandler) SyncIssues(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	issues, err := h.svc.SyncIssues(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, gin.H{"synced": len(issues), "issues": issues})
}

// Webhook handles POST /api/v1/webhook/github/:id (public endpoint)
func (h *GitHubHandler) Webhook(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	eventType := c.GetHeader("X-GitHub-Event")
	body, _ := io.ReadAll(c.Request.Body)

	result, err := h.svc.HandleWebhook(id, body, eventType)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}
