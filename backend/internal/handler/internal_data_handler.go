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
	if err := h.db.Preload("Project").Preload("State").First(&issue, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "issue not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get issue"})
		return
	}
	description := ""
	if issue.DescriptionStripped != nil {
		description = *issue.DescriptionStripped
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                 issue.ID,
		"title":              issue.Name,
		"description":        description,
		"status":             issue.State.Name,
		"state_group":        issue.State.Group,
		"priority":           issue.Priority,
		"project_id":         issue.ProjectID,
		"workspace_id":       issue.Project.WorkspaceID,
		"sequence_id":        issue.SequenceID,
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
		"id":           project.ID,
		"name":         project.Name,
		"identifier":   project.Identifier,
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

// SearchIssues searches issues. Supports a simple text search via POST with JSON body {"query": "..."}.
func (h *InternalDataHandler) SearchIssues(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	var issues []model.Issue
	q := h.db.Preload("Project").Preload("State")
	if req.Query != "" {
		like := "%" + req.Query + "%"
		q = q.Where("name ILIKE ? OR description_stripped ILIKE ?", like, like)
	}
	if err := q.Limit(50).Find(&issues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to search issues"})
		return
	}
	result := make([]gin.H, len(issues))
	for i, issue := range issues {
		description := ""
		if issue.DescriptionStripped != nil {
			description = *issue.DescriptionStripped
		}
		result[i] = gin.H{
			"id":                 issue.ID,
			"title":              issue.Name,
			"description":        description,
			"status":             issue.State.Name,
			"state_group":        issue.State.Group,
			"priority":           issue.Priority,
			"project_id":         issue.ProjectID,
			"workspace_id":       issue.Project.WorkspaceID,
			"sequence_id":        issue.SequenceID,
			"project_identifier": issue.Project.Identifier,
		}
	}
	c.JSON(http.StatusOK, result)
}
