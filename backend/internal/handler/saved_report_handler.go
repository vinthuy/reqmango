package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

type SavedReportHandler struct{ svc *service.SavedReportService }

func NewSavedReportHandler(svc *service.SavedReportService) *SavedReportHandler {
	return &SavedReportHandler{svc: svc}
}

func (h *SavedReportHandler) List(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	reports, err := h.svc.List(pid)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, reports)
}

func (h *SavedReportHandler) Create(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var report model.SavedReport
	if err := c.ShouldBindJSON(&report); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	report.ProjectID = pid
	if err := h.svc.Create(&report); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, report)
}

func (h *SavedReportHandler) Update(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.svc.Update(id, pid, updates); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Updated"})
}

func (h *SavedReportHandler) Delete(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id, pid); err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, gin.H{"message": "Deleted"})
}
