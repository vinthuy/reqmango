package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type InitiativeHandler struct {
	svc *service.InitiativeService
}

func NewInitiativeHandler(db *gorm.DB) *InitiativeHandler {
	return &InitiativeHandler{svc: service.NewInitiativeService(db)}
}

func (h *InitiativeHandler) Create(c *gin.Context) {
	var req request.CreateInitiativeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wsID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if wsID == 0 {
		wsID = req.WorkspaceID
	}
	initiative, err := h.svc.Create(wsID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": initiative})
}

func (h *InitiativeHandler) List(c *gin.Context) {
	wsID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	initiatives, err := h.svc.List(wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": initiatives})
}

func (h *InitiativeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("initiativeId"), 10, 64)
	i, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": i})
}

func (h *InitiativeHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("initiativeId"), 10, 64)
	var req request.UpdateInitiativeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	i, err := h.svc.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": i})
}

func (h *InitiativeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("initiativeId"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *InitiativeHandler) GetProgress(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("initiativeId"), 10, 64)
	progress, err := h.svc.GetProgress(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": progress})
}
