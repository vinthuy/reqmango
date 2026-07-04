package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

type FieldPermissionHandler struct{ svc *service.FieldPermissionService }

func NewFieldPermissionHandler(svc *service.FieldPermissionService) *FieldPermissionHandler {
	return &FieldPermissionHandler{svc: svc}
}

func (h *FieldPermissionHandler) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("project_id"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	perms, err := h.svc.List(projectID, workspaceID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(200, perms)
}

func (h *FieldPermissionHandler) Create(c *gin.Context) {
	var perm model.FieldPermission
	if err := c.ShouldBindJSON(&perm); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if err := h.svc.Create(&perm); err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(201, perm)
}

func (h *FieldPermissionHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		CanRead  *bool `json:"can_read"`
		CanWrite *bool `json:"can_write"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if err := h.svc.Update(id, req.CanRead, req.CanWrite); err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Field permission updated"})
}

func (h *FieldPermissionHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Field permission deleted"})
}

func (h *FieldPermissionHandler) CheckAccess(c *gin.Context) {
	resource := c.Query("resource")
	fieldName := c.Query("field_name")
	roleID, _ := strconv.ParseUint(c.Query("role_id"), 10, 64)
	projectID, _ := strconv.ParseUint(c.Query("project_id"), 10, 64)
	canRead, canWrite := h.svc.CheckFieldAccess(resource, fieldName, roleID, projectID)
	c.JSON(200, gin.H{"can_read": canRead, "can_write": canWrite})
}
