package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// DashboardHandler handles dashboard HTTP endpoints.
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func (h *DashboardHandler) getUserID(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists {
		return 0
	}
	if u, ok := user.(*model.User); ok {
		return u.ID
	}
	return 0
}

// ==================== Dashboard CRUD ====================

// List handles GET /projects/:projectId/dashboards
func (h *DashboardHandler) List(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)

	dashboards, svcErr := h.svc.List(projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, dashboards)
}

// Get handles GET /projects/:projectId/dashboards/:id
func (h *DashboardHandler) Get(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	dashboard, svcErr := h.svc.Get(id, projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, dashboard)
}

// Create handles POST /projects/:projectId/dashboards
func (h *DashboardHandler) Create(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)

	var req request.CreateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	dashboard, svcErr := h.svc.Create(&req, projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondCreated(c, dashboard)
}

// Update handles PUT /projects/:projectId/dashboards/:id
func (h *DashboardHandler) Update(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req request.UpdateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	dashboard, svcErr := h.svc.Update(id, projectID, userID, &req)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, dashboard)
}

// Delete handles DELETE /projects/:projectId/dashboards/:id
func (h *DashboardHandler) Delete(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	if svcErr := h.svc.Delete(id, projectID, userID); svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, gin.H{"message": "Dashboard deleted"})
}

// SetDefault handles POST /projects/:projectId/dashboards/:id/set-default
func (h *DashboardHandler) SetDefault(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	dashboard, svcErr := h.svc.SetDefault(id, projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, dashboard)
}

// Duplicate handles POST /projects/:projectId/dashboards/:id/duplicate
func (h *DashboardHandler) Duplicate(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	dashboard, svcErr := h.svc.Duplicate(id, projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondCreated(c, dashboard)
}

// ==================== Widget CRUD ====================

// AddWidget handles POST /projects/:projectId/dashboards/:id/widgets
func (h *DashboardHandler) AddWidget(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	dashboardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req request.CreateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	widget, svcErr := h.svc.AddWidget(dashboardID, projectID, userID, &req)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondCreated(c, widget)
}

// UpdateWidget handles PUT /projects/:projectId/dashboards/:id/widgets/:widgetId
func (h *DashboardHandler) UpdateWidget(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	dashboardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}
	widgetID, err := strconv.ParseUint(c.Param("widgetId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid widget ID"})
		return
	}

	var req request.UpdateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	widget, svcErr := h.svc.UpdateWidget(dashboardID, widgetID, projectID, userID, &req)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, widget)
}

// DeleteWidget handles DELETE /projects/:projectId/dashboards/:id/widgets/:widgetId
func (h *DashboardHandler) DeleteWidget(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	dashboardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}
	widgetID, err := strconv.ParseUint(c.Param("widgetId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid widget ID"})
		return
	}

	if svcErr := h.svc.DeleteWidget(dashboardID, widgetID, projectID, userID); svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, gin.H{"message": "Widget deleted"})
}

// ReorderWidgets handles PUT /projects/:projectId/dashboards/:id/widgets/reorder
func (h *DashboardHandler) ReorderWidgets(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	dashboardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req request.ReorderWidgetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if svcErr := h.svc.ReorderWidgets(dashboardID, projectID, userID, &req); svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, gin.H{"message": "Widgets reordered"})
}

// ==================== Full Data ====================

// GetFull handles GET /projects/:projectId/dashboards/:id/full
func (h *DashboardHandler) GetFull(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	full, svcErr := h.svc.GetFull(id, projectID, userID)
	if svcErr != nil {
		common.RespondError(c, svcErr)
		return
	}
	common.RespondOK(c, full)
}
