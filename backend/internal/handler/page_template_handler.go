package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
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

func parseProjectAndTemplate(c *gin.Context, withTemplate bool) (projectID, templateID uint64, ok bool) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return 0, 0, false
	}
	if withTemplate {
		templateID, err = strconv.ParseUint(c.Param("templateId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
			return 0, 0, false
		}
	}
	return projectID, templateID, true
}

// List handles GET /projects/:projectId/page-templates
func (h *PageTemplateHandler) List(c *gin.Context) {
	projectID, _, ok := parseProjectAndTemplate(c, false)
	if !ok {
		return
	}
	templates, err := h.svc.List(projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, templates)
}

// Get handles GET /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Get(c *gin.Context) {
	projectID, templateID, ok := parseProjectAndTemplate(c, true)
	if !ok {
		return
	}
	template, err := h.svc.Get(templateID, projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, template)
}

// Create handles POST /projects/:projectId/page-templates
func (h *PageTemplateHandler) Create(c *gin.Context) {
	projectID, _, ok := parseProjectAndTemplate(c, false)
	if !ok {
		return
	}
	var req request.PageTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	template, err := h.svc.Create(&req, projectID, getUserIDFromContext(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, template)
}

// Update handles PUT /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Update(c *gin.Context) {
	projectID, templateID, ok := parseProjectAndTemplate(c, true)
	if !ok {
		return
	}
	var req request.PageTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	template, err := h.svc.Update(templateID, projectID, getUserIDFromContext(c), &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, template)
}

// Delete handles DELETE /projects/:projectId/page-templates/:templateId
func (h *PageTemplateHandler) Delete(c *gin.Context) {
	projectID, templateID, ok := parseProjectAndTemplate(c, true)
	if !ok {
		return
	}
	if err := h.svc.Delete(templateID, projectID, getUserIDFromContext(c)); err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}
