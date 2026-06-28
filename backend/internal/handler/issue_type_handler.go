package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/middleware"
	"github.com/reqmanpy/backend/internal/service"
)

type IssueTypeHandler struct {
	svc *service.IssueTypeService
}

func NewIssueTypeHandler(svc *service.IssueTypeService) *IssueTypeHandler {
	return &IssueTypeHandler{svc: svc}
}

func (h *IssueTypeHandler) parseTypeID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("typeId"), 10, 64)
}

func (h *IssueTypeHandler) parseFieldID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("fieldId"), 10, 64)
}

// ==================== CRUD ====================

func (h *IssueTypeHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.IssueTypeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.Create(workspaceID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *IssueTypeHandler) List(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var projectID *uint64
	if v := c.Query("project_id"); v != "" {
		if id, e := strconv.ParseUint(v, 10, 64); e == nil {
			projectID = &id
		}
	}

	resp, svcErr := h.svc.List(workspaceID, projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *IssueTypeHandler) Get(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	resp, svcErr := h.svc.Get(typeID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *IssueTypeHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	var req request.IssueTypeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.Update(typeID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *IssueTypeHandler) Delete(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	if svcErr := h.svc.Delete(typeID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue type deleted"})
}

func (h *IssueTypeHandler) Disable(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	var req request.IssueTypeDisable
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if svcErr := h.svc.Disable(typeID, req.IsActive); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue type status updated"})
}

// ==================== Field Association ====================

func (h *IssueTypeHandler) ListFields(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	resp, svcErr := h.svc.ListFields(typeID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *IssueTypeHandler) AddField(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	var req request.IssueTypeFieldCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.AddField(typeID, req.FieldID, req.IsRequired, req.Sequence)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *IssueTypeHandler) RemoveField(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	if svcErr := h.svc.RemoveField(typeID, fieldID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Field removed from issue type"})
}

func (h *IssueTypeHandler) UpdateField(c *gin.Context) {
	typeID, err := h.parseTypeID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	var req request.IssueTypeFieldUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.UpdateField(typeID, fieldID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
