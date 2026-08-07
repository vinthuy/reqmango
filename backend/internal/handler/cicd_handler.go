package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// CICDHandler exposes the CI/CD HTTP endpoints
// (PRD P4-004: CI/CD 配置、触发构建、查看状态).
//
// Routes are grouped under /workspaces/:wsParam/cicd:
//
//	GET    /configs                list configs (optional ?project_id=)
//	POST   /configs                create config
//	GET    /configs/:configId      get config
//	PATCH  /configs/:configId      update config
//	DELETE /configs/:configId      delete config
//	GET    /builds                 list builds (filters: config_id, status, project_id, limit)
//	POST   /builds                 trigger a new build
//	GET    /builds/:buildId        get build
//	POST   /builds/:buildId/cancel cancel a running/pending build
//	DELETE /builds/:buildId        delete a build record
type CICDHandler struct {
	svc *service.CICDService
}

func NewCICDHandler(svc *service.CICDService) *CICDHandler {
	return &CICDHandler{svc: svc}
}

func (h *CICDHandler) respond(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(500, gin.H{"message": "Internal server error"})
	return true
}

func (h *CICDHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *CICDHandler) parseOptionalUint(c *gin.Context, key string) *uint64 {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

// ======== Config endpoints ========

// ListConfigs handles GET /workspaces/:wsParam/cicd/configs
// Query params: project_id (optional)
func (h *CICDHandler) ListConfigs(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var projectID *uint64
	if v := h.parseOptionalUint(c, "project_id"); v != nil {
		projectID = v
	}
	resp, e := h.svc.ListConfigs(wid, projectID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// GetConfig handles GET /workspaces/:wsParam/cicd/configs/:configId
func (h *CICDHandler) GetConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("configId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid config id"})
		return
	}
	resp, e := h.svc.GetConfig(configID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// CreateConfig handles POST /workspaces/:wsParam/cicd/configs
func (h *CICDHandler) CreateConfig(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var req service.CICDConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	resp, e := h.svc.CreateConfig(wid, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

// UpdateConfig handles PATCH /workspaces/:wsParam/cicd/configs/:configId
func (h *CICDHandler) UpdateConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("configId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid config id"})
		return
	}
	var req service.CICDConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	resp, e := h.svc.UpdateConfig(configID, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// DeleteConfig handles DELETE /workspaces/:wsParam/cicd/configs/:configId
func (h *CICDHandler) DeleteConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("configId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid config id"})
		return
	}
	if e := h.svc.DeleteConfig(configID, middleware.GetCurrentUser(c).ID); h.respond(c, e) {
		return
	}
	c.JSON(204, gin.H{"message": "deleted"})
}

// ======== Build endpoints ========

// ListBuilds handles GET /workspaces/:wsParam/cicd/builds
// Query params: config_id (optional), status (optional), project_id (optional), limit (optional)
func (h *CICDHandler) ListBuilds(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var configID *uint64
	if v := h.parseOptionalUint(c, "config_id"); v != nil {
		configID = v
	}
	var projectID *uint64
	if v := h.parseOptionalUint(c, "project_id"); v != nil {
		projectID = v
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	resp, e := h.svc.ListBuilds(wid, configID, c.Query("status"), projectID, limit)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// GetBuild handles GET /workspaces/:wsParam/cicd/builds/:buildId
func (h *CICDHandler) GetBuild(c *gin.Context) {
	buildID, err := strconv.ParseUint(c.Param("buildId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid build id"})
		return
	}
	resp, e := h.svc.GetBuild(buildID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// TriggerBuild handles POST /workspaces/:wsParam/cicd/builds
// Starts a new build for the supplied config. The workflow runs
// asynchronously; the response returns the pending build so the client can
// poll or subscribe to SSE events (cicd_build.updated) for progress.
func (h *CICDHandler) TriggerBuild(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var req service.BuildTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	resp, e := h.svc.TriggerBuild(wid, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

// CancelBuild handles POST /workspaces/:wsParam/cicd/builds/:buildId/cancel
func (h *CICDHandler) CancelBuild(c *gin.Context) {
	buildID, err := strconv.ParseUint(c.Param("buildId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid build id"})
		return
	}
	resp, e := h.svc.CancelBuild(buildID, middleware.GetCurrentUser(c).ID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// DeleteBuild handles DELETE /workspaces/:wsParam/cicd/builds/:buildId
func (h *CICDHandler) DeleteBuild(c *gin.Context) {
	buildID, err := strconv.ParseUint(c.Param("buildId"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid build id"})
		return
	}
	if e := h.svc.DeleteBuild(buildID, middleware.GetCurrentUser(c).ID); h.respond(c, e) {
		return
	}
	c.JSON(204, gin.H{"message": "deleted"})
}
