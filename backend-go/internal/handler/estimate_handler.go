package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/service"
	"net/http"
	"strconv"
)

type EstimateHandler struct {
	service *service.EstimateService
}

func NewEstimateHandler(service *service.EstimateService) *EstimateHandler {
	return &EstimateHandler{service: service}
}

func (h *EstimateHandler) GetSettings(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	settings, err := h.service.GetSettings(projectID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *EstimateHandler) UpdateSettings(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	settings, err := h.service.UpdateSettings(projectID, workspaceID, model.EstimateMode(req.Mode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *EstimateHandler) ListPoints(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	points, err := h.service.ListPoints(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (h *EstimateHandler) GetPoint(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	pointID, _ := strconv.ParseUint(c.Param("pointId"), 10, 64)
	point, err := h.service.GetPoint(projectID, pointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, point)
}

func (h *EstimateHandler) CreatePoint(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	var req struct {
		Name      string `json:"name"`
		Value     int    `json:"value"`
		IsDefault bool   `json:"is_default"`
		Sequence  int    `json:"sequence"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	point, err := h.service.CreatePoint(projectID, workspaceID, req.Name, req.Value, req.IsDefault, req.Sequence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, point)
}

func (h *EstimateHandler) UpdatePoint(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	pointID, _ := strconv.ParseUint(c.Param("pointId"), 10, 64)

	var req struct {
		Name      *string `json:"name"`
		Value     *int    `json:"value"`
		IsDefault *bool   `json:"is_default"`
		Sequence  *int    `json:"sequence"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	point, err := h.service.UpdatePoint(projectID, pointID, req.Name, req.Value, req.IsDefault, req.Sequence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, point)
}

func (h *EstimateHandler) DeletePoint(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	pointID, _ := strconv.ParseUint(c.Param("pointId"), 10, 64)

	err := h.service.DeletePoint(projectID, pointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Estimate point deleted"})
}

func (h *EstimateHandler) ReorderPoints(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)

	var req struct {
		PointIDs []uint64 `json:"point_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.service.ReorderPoints(projectID, req.PointIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Estimate points reordered"})
}

func (h *EstimateHandler) CreateDefaultPoints(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	points, err := h.service.CreateDefaultPoints(projectID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (h *EstimateHandler) BulkCreatePoints(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	var req struct {
		Points []struct {
			Name      string `json:"name"`
			Value     int    `json:"value"`
			IsDefault bool   `json:"is_default"`
			Sequence  int    `json:"sequence"`
		} `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	points, err := h.service.BulkCreatePoints(projectID, workspaceID, req.Points)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (h *EstimateHandler) ListCategories(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	categories, err := h.service.ListCategories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *EstimateHandler) CreateCategory(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	var req struct {
		Name      string `json:"name"`
		IsDefault bool   `json:"is_default"`
		Sequence  int    `json:"sequence"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	category, err := h.service.CreateCategory(projectID, workspaceID, req.Name, req.IsDefault, req.Sequence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *EstimateHandler) CreateDefaultCategories(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	categories, err := h.service.CreateDefaultCategories(projectID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *EstimateHandler) ListTime(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	times, err := h.service.ListTime(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, times)
}

func (h *EstimateHandler) CreateTime(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	var req struct {
		Name      string `json:"name"`
		Minutes   int    `json:"minutes"`
		IsDefault bool   `json:"is_default"`
		Sequence  int    `json:"sequence"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	etime, err := h.service.CreateTime(projectID, workspaceID, req.Name, req.Minutes, req.IsDefault, req.Sequence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, etime)
}

func (h *EstimateHandler) CreateDefaultTime(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)

	times, err := h.service.CreateDefaultTime(projectID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, times)
}