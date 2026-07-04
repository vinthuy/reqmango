package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
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

	// Convert model to response DTO
	var result []response.SavedReportResponse
	for _, r := range reports {
		result = append(result, response.SavedReportResponse{
			ID:         r.ID,
			Name:       r.Name,
			ReportType: r.ReportType,
			GroupBy:    r.GroupBy,
			ChartType:  r.ChartType,
			RQL:        r.RQL,
			Interval:   r.Interval,
			DateFrom:   r.DateFrom,
			DateTo:     r.DateTo,
			ProjectID:  r.ProjectID,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		})
	}
	common.RespondOK(c, result)
}

func (h *SavedReportHandler) Create(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var req request.SavedReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}

	report, err := h.svc.Create(pid, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, response.SavedReportResponse{
		ID:         report.ID,
		Name:       report.Name,
		ReportType: report.ReportType,
		GroupBy:    report.GroupBy,
		ChartType:  report.ChartType,
		RQL:        report.RQL,
		Interval:   report.Interval,
		DateFrom:   report.DateFrom,
		DateTo:     report.DateTo,
		ProjectID:  report.ProjectID,
		CreatedAt:  report.CreatedAt,
		UpdatedAt:  report.UpdatedAt,
	})
}

func (h *SavedReportHandler) Update(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req request.SavedReportUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest(err.Error()))
		return
	}

	report, err := h.svc.Update(id, pid, &req)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	common.RespondOK(c, response.SavedReportResponse{
		ID:         report.ID,
		Name:       report.Name,
		ReportType: report.ReportType,
		GroupBy:    report.GroupBy,
		ChartType:  report.ChartType,
		RQL:        report.RQL,
		Interval:   report.Interval,
		DateFrom:   report.DateFrom,
		DateTo:     report.DateTo,
		ProjectID:  report.ProjectID,
		CreatedAt:  report.CreatedAt,
		UpdatedAt:  report.UpdatedAt,
	})
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
