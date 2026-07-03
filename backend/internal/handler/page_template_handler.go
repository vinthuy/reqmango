package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

// PageTemplateHandler handles page template endpoints.
type PageTemplateHandler struct {
	svc *service.PageTemplateService
}

// NewPageTemplateHandler creates a new PageTemplateHandler.
func NewPageTemplateHandler(svc *service.PageTemplateService) *PageTemplateHandler {
	return &PageTemplateHandler{svc: svc}
}

// List handles GET /projects/:projectId/page-templates
func (h *PageTemplateHandler) List(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var projectID *uint64
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err == nil {
			projectID = &pid
		}
	}

	templates, svcErr := h.svc.List(workspaceID, projectID)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// Get handles GET /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Get(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}

	template, svcErr := h.svc.Get(templateID)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, template)
}

// Create handles POST /projects/:projectId/page-templates
func (h *PageTemplateHandler) Create(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.PageTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	template, svcErr := h.svc.Create(&req, workspaceID, getUserIDFromContext(c))
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusCreated, template)
}

// Update handles PUT /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Update(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}

	var req request.PageTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	template, svcErr := h.svc.Update(templateID, getUserIDFromContext(c), &req)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, template)
}

// Delete handles DELETE /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Delete(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}

	if svcErr := h.svc.Delete(templateID); svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}
