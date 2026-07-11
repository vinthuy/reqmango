package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type RoleHandler struct {
	roleService *service.RoleService
	db          *gorm.DB
}

func NewRoleHandler(roleService *service.RoleService, db *gorm.DB) *RoleHandler {
	return &RoleHandler{roleService: roleService, db: db}
}

func (h *RoleHandler) resolveWorkspaceID(c *gin.Context) (uint64, error) {
	wsParam := c.Param("wsParam")
	if id, err := strconv.ParseUint(wsParam, 10, 64); err == nil {
		return id, nil
	}
	var workspace model.Workspace
	if err := h.db.Where("slug = ?", wsParam).First(&workspace).Error; err != nil {
		return 0, err
	}
	return workspace.ID, nil
}

// ListRoles returns roles for a workspace or project.
func (h *RoleHandler) ListRoles(c *gin.Context) {
	workspaceID := queryUint64(c, "workspace_id")
	projectID := queryUint64(c, "project_id")
	scope := c.DefaultQuery("scope", "workspace")

	roles, err := h.roleService.List(scope, workspaceID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// ListPermissions returns all available permissions.
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.roleService.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

// CreateRole creates a new custom role.
func (h *RoleHandler) CreateRole(c *gin.Context) {
	workspaceID, err := h.resolveWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req request.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.WorkspaceID = &workspaceID
	role, err := h.roleService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": role})
}

// UpdateRole updates a role.
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.roleService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": role})
}

// DeleteRole deletes a custom role.
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	if err := h.roleService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

func queryUint64(c *gin.Context, key string) *uint64 {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	n, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}
