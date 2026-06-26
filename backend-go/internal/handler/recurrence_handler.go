package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/service"
)

type RecurrenceHandler struct{ svc *service.RecurrenceService }

func NewRecurrenceHandler(svc *service.RecurrenceService) *RecurrenceHandler {
	return &RecurrenceHandler{svc: svc}
}

func (h *RecurrenceHandler) Create(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req request.RecurrenceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message": err.Error()}); return }
	r, err := h.svc.Create(&req, issueID)
	if err != nil { ae(c, err); return }
	c.JSON(201, r)
}

func (h *RecurrenceHandler) Get(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	r, err := h.svc.Get(issueID)
	if err != nil { ae(c, err); return }
	c.JSON(200, r)
}

func (h *RecurrenceHandler) Update(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req request.RecurrenceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message": err.Error()}); return }
	r, err := h.svc.Update(issueID, &req)
	if err != nil { ae(c, err); return }
	c.JSON(200, r)
}

func (h *RecurrenceHandler) Delete(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err := h.svc.Delete(issueID); err != nil { ae(c, err); return }
	c.JSON(200, gin.H{"message": "Recurrence rule deleted"})
}

func ae(c *gin.Context, err error) {
	if ae, ok := err.(*common.AppError); ok { c.JSON(ae.Code, gin.H{"message": ae.Message}); return }
	c.JSON(500, gin.H{"message": "Internal server error"})
}
