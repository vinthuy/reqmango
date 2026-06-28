package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/service"
)

// SavedViewHandler handles saved view HTTP endpoints.
type SavedViewHandler struct {
	svc *service.SavedViewService
}

// NewSavedViewHandler creates a new SavedViewHandler.
func NewSavedViewHandler(svc *service.SavedViewService) *SavedViewHandler {
	return &SavedViewHandler{svc: svc}
}

func (h *SavedViewHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func (h *SavedViewHandler) getUserID(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists {
		return 0
	}
	if u, ok := user.(*model.User); ok {
		return u.ID
	}
	return 0
}

// List handles GET /projects/:projectId/views
func (h *SavedViewHandler) List(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)

	views, svcErr := h.svc.List(projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, views)
}

// Get handles GET /projects/:projectId/views/:viewId
func (h *SavedViewHandler) Get(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	viewID, err := strconv.ParseUint(c.Param("viewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid view ID"})
		return
	}

	view, svcErr := h.svc.Get(viewID, projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// Create handles POST /projects/:projectId/views
func (h *SavedViewHandler) Create(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)

	var req request.SavedViewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	view, svcErr := h.svc.Create(&req, projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, view)
}

// Update handles PUT /projects/:projectId/views/:viewId
func (h *SavedViewHandler) Update(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	viewID, err := strconv.ParseUint(c.Param("viewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid view ID"})
		return
	}

	var req request.SavedViewUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	view, svcErr := h.svc.Update(viewID, projectID, userID, &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// Delete handles DELETE /projects/:projectId/views/:viewId
func (h *SavedViewHandler) Delete(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	viewID, err := strconv.ParseUint(c.Param("viewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid view ID"})
		return
	}

	if svcErr := h.svc.Delete(viewID, projectID, userID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Saved view deleted"})
}

// SetDefault handles POST /projects/:projectId/views/:viewId/set-default
func (h *SavedViewHandler) SetDefault(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	viewID, err := strconv.ParseUint(c.Param("viewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid view ID"})
		return
	}

	view, svcErr := h.svc.SetDefault(viewID, projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// Duplicate handles POST /projects/:projectId/views/:viewId/duplicate
func (h *SavedViewHandler) Duplicate(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	userID := h.getUserID(c)
	viewID, err := strconv.ParseUint(c.Param("viewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid view ID"})
		return
	}

	view, svcErr := h.svc.Duplicate(viewID, projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, view)
}
