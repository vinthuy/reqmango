package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/service"
)

type SearchTemplateHandler struct {
	svc *service.SearchTemplateService
}

func NewSearchTemplateHandler(svc *service.SearchTemplateService) *SearchTemplateHandler {
	return &SearchTemplateHandler{svc: svc}
}

func (h *SearchTemplateHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func (h *SearchTemplateHandler) getTemplateID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("templateId"), 10, 64)
}

func (h *SearchTemplateHandler) List(c *gin.Context) {
	pid, err := h.getProjectID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}

	userID := c.GetUint64("user_id")
	templates, err := h.svc.List(pid, userID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, templates)
}

func (h *SearchTemplateHandler) Get(c *gin.Context) {
	pid, err := h.getProjectID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}
	tid, err := h.getTemplateID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid template ID"))
		return
	}

	userID := c.GetUint64("user_id")
	template, err := h.svc.Get(tid, pid, userID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, template)
}

type CreateSearchTemplateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	RQLTemplate string                 `json:"rql_template"`
	ViewType    string                 `json:"view_type"`
	SortConfig  map[string]interface{} `json:"sort_config"`
	GroupBy     *string                `json:"group_by"`
	Columns     []string               `json:"columns"`
}

func (h *SearchTemplateHandler) Create(c *gin.Context) {
	pid, err := h.getProjectID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}

	var req CreateSearchTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	userID := c.GetUint64("user_id")

	template, err := h.svc.Create(pid, userID, req.Name, req.Description, req.Icon, req.RQLTemplate, req.ViewType, nil, nil, req.GroupBy)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, template)
}

func (h *SearchTemplateHandler) Delete(c *gin.Context) {
	pid, err := h.getProjectID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}
	tid, err := h.getTemplateID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid template ID"))
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.svc.Delete(tid, pid, userID); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Template deleted"})
}

func (h *SearchTemplateHandler) Apply(c *gin.Context) {
	pid, err := h.getProjectID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}
	tid, err := h.getTemplateID(c)
	if err != nil {
		common.RespondError(c, common.BadRequest("Invalid template ID"))
		return
	}

	userID := c.GetUint64("user_id")
	template, err := h.svc.ApplyTemplate(tid, pid, userID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, template)
}