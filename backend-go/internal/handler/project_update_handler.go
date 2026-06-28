package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/service"
	"gorm.io/gorm"
)

type ProjectUpdateHandler struct{ svc *service.ProjectUpdateService }

func NewProjectUpdateHandler(db *gorm.DB) *ProjectUpdateHandler {
	return &ProjectUpdateHandler{svc: service.NewProjectUpdateService(db)}
}

func (h *ProjectUpdateHandler) Create(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var body struct {
		Status  string `json:"status" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Create(projectID, 1, body.Status, body.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": u})
}

func (h *ProjectUpdateHandler) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	updates, err := h.svc.List(projectID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updates})
}
