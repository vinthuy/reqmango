package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/service"
)

type GitIntegrationHandler struct {
	gitSvc *service.GitService
}

func NewGitIntegrationHandler(gitSvc *service.GitService) *GitIntegrationHandler {
	return &GitIntegrationHandler{gitSvc: gitSvc}
}

func (h *GitIntegrationHandler) CreateIntegration(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	var req struct {
		Provider      string `json:"provider"`
		RepoURL       string `json:"repo_url"`
		RepoName      string `json:"repo_name"`
		AccessToken   string `json:"access_token"`
		WebhookSecret string `json:"webhook_secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	integration, err := h.gitSvc.CreateIntegration(projectID, req.Provider, req.RepoURL, req.RepoName, req.AccessToken, req.WebhookSecret)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create git integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              integration.ID,
		"project_id":      integration.ProjectID,
		"provider":        integration.Provider,
		"repo_url":        integration.RepoURL,
		"repo_name":       integration.RepoName,
		"active":          integration.Active,
		"sync_prs":        integration.SyncPRs,
		"sync_commits":    integration.SyncCommits,
		"sync_branches":   integration.SyncBranches,
		"created_at":      integration.CreatedAt,
		"updated_at":      integration.UpdatedAt,
	})
}

func (h *GitIntegrationHandler) GetIntegration(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	integration, err := h.gitSvc.GetIntegration(projectID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch git integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              integration.ID,
		"project_id":      integration.ProjectID,
		"provider":        integration.Provider,
		"repo_url":        integration.RepoURL,
		"repo_name":       integration.RepoName,
		"active":          integration.Active,
		"sync_prs":        integration.SyncPRs,
		"sync_commits":    integration.SyncCommits,
		"sync_branches":   integration.SyncBranches,
		"created_at":      integration.CreatedAt,
		"updated_at":      integration.UpdatedAt,
	})
}

func (h *GitIntegrationHandler) UpdateIntegration(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	integration, err := h.gitSvc.UpdateIntegration(projectID, req)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update git integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              integration.ID,
		"project_id":      integration.ProjectID,
		"provider":        integration.Provider,
		"repo_url":        integration.RepoURL,
		"repo_name":       integration.RepoName,
		"active":          integration.Active,
		"sync_prs":        integration.SyncPRs,
		"sync_commits":    integration.SyncCommits,
		"sync_branches":   integration.SyncBranches,
		"created_at":      integration.CreatedAt,
		"updated_at":      integration.UpdatedAt,
	})
}

func (h *GitIntegrationHandler) DeleteIntegration(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	err = h.gitSvc.DeleteIntegration(projectID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete git integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Git integration deleted"})
}

func (h *GitIntegrationHandler) GetIssueGitLinks(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue_id"})
		return
	}

	links, err := h.gitSvc.GetIssueGitLinks(issueID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch git links"})
		return
	}

	c.JSON(http.StatusOK, links)
}