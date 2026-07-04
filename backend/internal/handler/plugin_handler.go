package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/plugin"
	"github.com/reqmango/backend/internal/service"
)

// PluginHandler handles plugin HTTP endpoints.
type PluginHandler struct {
	svc *service.PluginService
}

// NewPluginHandler creates a new PluginHandler.
func NewPluginHandler(svc *service.PluginService) *PluginHandler {
	return &PluginHandler{svc: svc}
}

func (h *PluginHandler) getWorkspaceID(c *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListCatalog handles GET /workspaces/:wsParam/plugins/catalog
func (h *PluginHandler) ListCatalog(c *gin.Context) {
	catalog := plugin.BuiltinCatalog()
	c.JSON(http.StatusOK, catalog)
}

// ListInstalled handles GET /workspaces/:wsParam/plugins
func (h *PluginHandler) ListInstalled(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	plugins, svcErr := h.svc.ListInstalled(workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, plugins)
}

// Install handles POST /workspaces/:wsParam/plugins
func (h *PluginHandler) Install(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	user := middleware.GetCurrentUser(c)

	var req request.PluginInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	p, svcErr := h.svc.Install(workspaceID, user.ID, &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Get handles GET /workspaces/:wsParam/plugins/:id
func (h *PluginHandler) Get(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	p, svcErr := h.svc.Get(id, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Update handles PUT /workspaces/:wsParam/plugins/:id
func (h *PluginHandler) Update(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	var req request.PluginUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	p, svcErr := h.svc.Update(id, workspaceID, &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Uninstall handles DELETE /workspaces/:wsParam/plugins/:id
func (h *PluginHandler) Uninstall(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	if svcErr := h.svc.Uninstall(id, workspaceID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plugin uninstalled"})
}

// Enable handles POST /workspaces/:wsParam/plugins/:id/enable
func (h *PluginHandler) Enable(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	p, svcErr := h.svc.Enable(id, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Disable handles POST /workspaces/:wsParam/plugins/:id/disable
func (h *PluginHandler) Disable(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	p, svcErr := h.svc.Disable(id, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GetEventLogs handles GET /workspaces/:wsParam/plugins/:id/logs
func (h *PluginHandler) GetEventLogs(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, svcErr := h.svc.GetEventLogs(id, workspaceID, limit)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	// Return as array directly (or wrap)
	type logEntry struct {
		ID           uint64 `json:"id"`
		PluginID     uint64 `json:"plugin_id"`
		EventType    string `json:"event_type"`
		Status       string `json:"status"`
		RequestBody  string `json:"request_body"`
		ResponseBody string `json:"response_body"`
		StatusCode   int    `json:"status_code"`
		DurationMs   int64  `json:"duration_ms"`
		CreatedAt    string `json:"created_at"`
	}

	entries := make([]logEntry, len(logs))
	for i, l := range logs {
		entries[i] = logEntry{
			ID:           l.ID,
			PluginID:     l.PluginID,
			EventType:    l.EventType,
			Status:       l.Status,
			RequestBody:  l.RequestBody,
			ResponseBody: l.ResponseBody,
			StatusCode:   l.StatusCode,
			DurationMs:   l.DurationMs,
			CreatedAt:    l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, entries)
}

// TestExecute handles POST /workspaces/:wsParam/plugins/:id/test
func (h *PluginHandler) TestExecute(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid plugin ID"})
		return
	}

	user := middleware.GetCurrentUser(c)

	// Get plugin
	p, svcErr := h.svc.Get(id, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	// Build a test payload
	testPayload := map[string]interface{}{
		"event_type":    "test.execute",
		"workspace_id":  workspaceID,
		"actor_id":      user.ID,
		"message":       "This is a test execution",
		"timestamp":     "now",
	}

	payloadBytes, _ := json.Marshal(testPayload)
	c.JSON(http.StatusOK, gin.H{
		"message":       "Test payload would be dispatched",
		"plugin_id":     p.ID,
		"plugin_slug":   p.Slug,
		"plugin_type":   p.Type,
		"payload":       string(payloadBytes),
	})
}
