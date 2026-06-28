package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type ReleaseHandler struct {
	service *service.ReleaseService
}

func NewReleaseHandler(service *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{service: service}
}

func (h *ReleaseHandler) Create(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)

	var req request.ReleaseCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	release, err := h.service.Create(projectID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) List(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)

	releases, err := h.service.List(projectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": releases})
}

func (h *ReleaseHandler) Get(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	release, err := h.service.Get(projectID, releaseID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) Update(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	var req request.ReleaseUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	release, err := h.service.Update(projectID, releaseID, &req)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, release)
}

func (h *ReleaseHandler) Delete(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	if err := h.service.Delete(projectID, releaseID); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *ReleaseHandler) AddIssues(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	var req request.ReleaseIssueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddIssues(projectID, releaseID, req.IssueIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *ReleaseHandler) RemoveIssues(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	var req request.ReleaseIssueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RemoveIssues(projectID, releaseID, req.IssueIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *ReleaseHandler) GetProgress(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("projectId"), 10, 64)
	releaseID, _ := strconv.ParseUint(ctx.Param("releaseId"), 10, 64)

	progress, err := h.service.GetProgress(projectID, releaseID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, progress)
}

func formatReleaseDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}