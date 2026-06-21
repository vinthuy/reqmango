package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/service"
)

type CycleHandler struct {
	svc *service.CycleService
}

func NewCycleHandler(svc *service.CycleService) *CycleHandler {
	return &CycleHandler{svc: svc}
}

// List handles GET /projects/:projectId/cycles
func (h *CycleHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	cycles, err := h.svc.ListByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, cycles)
}
