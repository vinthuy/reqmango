package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/service"
)

type ReportHandler struct{ svc *service.ReportService }

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{svc: svc} }

func (h *ReportHandler) Generate(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var req service.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.Generate(pid, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}

func (h *ReportHandler) GenerateV2(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var req service.ReportV2Request
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.GenerateV2(pid, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}
