package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

type WorkItemTemplateHandler struct {
	svc *service.WorkItemTemplateService
}

func NewWorkItemTemplateHandler(svc *service.WorkItemTemplateService) *WorkItemTemplateHandler {
	return &WorkItemTemplateHandler{svc: svc}
}

func (h *WorkItemTemplateHandler) getProjectID(c *gin.Context) (uint64, uint64) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	return projectID, workspaceID
}

func (h *WorkItemTemplateHandler) List(c *gin.Context) {
	projectID, _ := h.getProjectID(c)
	if projectID == 0 {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}
	r, err := h.svc.List(projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, r)
}

func (h *WorkItemTemplateHandler) Get(c *gin.Context) {
	projectID, _ := h.getProjectID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		common.RespondError(c, common.BadRequest("Invalid ID"))
		return
	}
	r, err := h.svc.Get(id, projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, r)
}

func (h *WorkItemTemplateHandler) Create(c *gin.Context) {
	projectID, workspaceID := h.getProjectID(c)
	if projectID == 0 || workspaceID == 0 {
		common.RespondError(c, common.BadRequest("Invalid project/workspace ID"))
		return
	}

	var req request.WorkItemTemplateCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}

	r, err := h.svc.Create(projectID, workspaceID, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondCreated(c, r)
}

func (h *WorkItemTemplateHandler) Update(c *gin.Context) {
	projectID, _ := h.getProjectID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		common.RespondError(c, common.BadRequest("Invalid ID"))
		return
	}

	var req request.WorkItemTemplateUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}

	r, err := h.svc.Update(id, projectID, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, r)
}

func (h *WorkItemTemplateHandler) Delete(c *gin.Context) {
	projectID, _ := h.getProjectID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		common.RespondError(c, common.BadRequest("Invalid ID"))
		return
	}

	if err := h.svc.Delete(id, projectID); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Work item template deleted"})
}

func (h *WorkItemTemplateHandler) GetDefault(c *gin.Context) {
	projectID, _ := h.getProjectID(c)
	if projectID == 0 {
		common.RespondError(c, common.BadRequest("Invalid project ID"))
		return
	}
	r, err := h.svc.GetDefault(projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, r)
}

func (h *WorkItemTemplateHandler) getCurrentUserID(c *gin.Context) uint64 {
	u, _ := c.Get("currentUser")
	if u, ok := u.(*model.User); ok {
		return u.ID
	}
	return 0
}
