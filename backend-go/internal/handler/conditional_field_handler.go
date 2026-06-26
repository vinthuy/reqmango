package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/service"
)

// ConditionalFieldHandler handles conditional field API requests.
type ConditionalFieldHandler struct {
	svc *service.ConditionalFieldService
}

// NewConditionalFieldHandler creates a new ConditionalFieldHandler.
func NewConditionalFieldHandler(svc *service.ConditionalFieldService) *ConditionalFieldHandler {
	return &ConditionalFieldHandler{svc: svc}
}

// Create handles POST /conditional-fields
func (h *ConditionalFieldHandler) Create(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.ConditionalFieldCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	req.WorkspaceID = workspaceID

	result, err := h.svc.Create(&req)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// List handles GET /conditional-fields
func (h *ConditionalFieldHandler) List(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var fieldID *uint64
	if fid := c.Query("field_id"); fid != "" {
		id, err := strconv.ParseUint(fid, 10, 64)
		if err == nil {
			fieldID = &id
		}
	}

	results, err := h.svc.List(workspaceID, fieldID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// Get handles GET /conditional-fields/:id
func (h *ConditionalFieldHandler) Get(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	result, err := h.svc.Get(id, workspaceID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update handles PUT /conditional-fields/:id
func (h *ConditionalFieldHandler) Update(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var req request.ConditionalFieldUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, err := h.svc.Update(id, workspaceID, &req)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete handles DELETE /conditional-fields/:id
func (h *ConditionalFieldHandler) Delete(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	if err := h.svc.Delete(id, workspaceID); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conditional field deleted"})
}

// EvaluateFieldVisibility handles POST /conditional-fields/evaluate
func (h *ConditionalFieldHandler) EvaluateFieldVisibility(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req struct {
		IssueTypeID *uint64  `json:"issue_type_id"`
		StateID     uint64   `json:"state_id"`
		Priority    string   `json:"priority"`
		AssigneeIDs []uint64 `json:"assignee_ids"`
		LabelIDs    []uint64 `json:"label_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	results, err := h.svc.EvaluateFieldVisibility(workspaceID, req.IssueTypeID, req.StateID, req.Priority, req.AssigneeIDs, req.LabelIDs)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"visible_field_ids": results})
}
