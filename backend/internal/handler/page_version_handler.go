package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

// PageVersionHandler handles page version endpoints.
type PageVersionHandler struct {
	svc *service.PageVersionService
}

// NewPageVersionHandler creates a new PageVersionHandler.
func NewPageVersionHandler(svc *service.PageVersionService) *PageVersionHandler {
	return &PageVersionHandler{svc: svc}
}

// List handles GET /projects/:projectId/pages/:pageId/versions
func (h *PageVersionHandler) List(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	versions, svcErr := h.svc.List(pageID)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, versions)
}

// Get handles GET /projects/:projectId/pages/:pageId/versions/:versionNumber
func (h *PageVersionHandler) Get(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}
	versionNumber, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid version number"})
		return
	}

	version, svcErr := h.svc.GetByVersion(pageID, versionNumber)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, version)
}

// Restore handles POST /projects/:projectId/pages/:pageId/versions/:versionNumber/restore
func (h *PageVersionHandler) Restore(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}
	versionNumber, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid version number"})
		return
	}

	userID := getUserIDFromContext(c)
	if svcErr := h.svc.Restore(pageID, versionNumber, userID); svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Page restored to version " + strconv.Itoa(versionNumber)})
}
