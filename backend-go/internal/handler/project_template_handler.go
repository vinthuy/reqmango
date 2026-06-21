package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/service"
)

type ProjectTemplateHandler struct {
	svc *service.ProjectTemplateService
}

func NewProjectTemplateHandler(svc *service.ProjectTemplateService) *ProjectTemplateHandler {
	return &ProjectTemplateHandler{svc: svc}
}

func (h *ProjectTemplateHandler) parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("templateId"), 10, 64)
}

func (h *ProjectTemplateHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}
	var req request.ProjectTemplateCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	resp, svcErr := h.svc.Create(workspaceID, user.ID, req)
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

func (h *ProjectTemplateHandler) List(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}
	resp, svcErr := h.svc.List(workspaceID)
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

func (h *ProjectTemplateHandler) Get(c *gin.Context) {
	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	resp, svcErr := h.svc.Get(id)
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

func (h *ProjectTemplateHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	var req request.ProjectTemplateUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	resp, svcErr := h.svc.Update(id, user.ID, req)
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

func (h *ProjectTemplateHandler) Delete(c *gin.Context) {
	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	if svcErr := h.svc.Delete(id); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

func (h *ProjectTemplateHandler) AddType(c *gin.Context) {
	templateID, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	var req request.ProjectTemplateAddType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	resp, svcErr := h.svc.AddType(templateID, req.TypeTemplateID, req.IsRequired, req.DefaultStateID, req.Sequence)
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

func (h *ProjectTemplateHandler) RemoveType(c *gin.Context) {
	templateID, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	typeID, err := strconv.ParseUint(c.Param("typeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid type ID"})
		return
	}
	if svcErr := h.svc.RemoveType(templateID, typeID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Type removed from template"})
}

func (h *ProjectTemplateHandler) Apply(c *gin.Context) {
	templateID, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid template ID"})
		return
	}
	var req request.ProjectTemplateApply
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	// Apply: set project's template_id
	resp, svcErr := h.svc.Get(templateID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template applied", "template": resp})
}
