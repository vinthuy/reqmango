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

type CustomFieldHandler struct {
	svc *service.CustomFieldService
}

func NewCustomFieldHandler(svc *service.CustomFieldService) *CustomFieldHandler {
	return &CustomFieldHandler{svc: svc}
}

func (h *CustomFieldHandler) parseFieldID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("fieldId"), 10, 64)
}

func (h *CustomFieldHandler) parseOptionID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("optionId"), 10, 64)
}

func (h *CustomFieldHandler) parseIssueID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("issueId"), 10, 64)
}

func (h *CustomFieldHandler) respondError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(appErr.Code, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}

// ==================== Field CRUD ====================

func (h *CustomFieldHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.CustomFieldCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.Create(workspaceID, user.ID, req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *CustomFieldHandler) List(c *gin.Context) {
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

	var issueTypeID *uint64
	if v := c.Query("issue_type_id"); v != "" {
		if id, e := strconv.ParseUint(v, 10, 64); e == nil {
			issueTypeID = &id
		}
	}

	resp, svcErr := h.svc.List(workspaceID, projectID, issueTypeID)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) Get(c *gin.Context) {
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	resp, svcErr := h.svc.Get(fieldID)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	var req request.CustomFieldUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.Update(fieldID, user.ID, req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) Delete(c *gin.Context) {
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	if svcErr := h.svc.Delete(fieldID); svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom field deleted"})
}

// ==================== Options ====================

func (h *CustomFieldHandler) CreateOption(c *gin.Context) {
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	var req request.CustomFieldOptionCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.CreateOption(fieldID, req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *CustomFieldHandler) UpdateOption(c *gin.Context) {
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}
	optionID, err := h.parseOptionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid option ID"})
		return
	}

	var req request.CustomFieldOptionUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.UpdateOption(fieldID, optionID, req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) DeleteOption(c *gin.Context) {
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}
	optionID, err := h.parseOptionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid option ID"})
		return
	}

	if svcErr := h.svc.DeleteOption(fieldID, optionID); svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Option deleted"})
}

// ==================== Issue Values ====================

func (h *CustomFieldHandler) SetIssueValue(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	var req request.IssueCustomFieldValueCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.SetIssueValue(issueID, req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *CustomFieldHandler) BulkSetIssueValues(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	var req request.BulkCustomFieldValueUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	req.IssueID = issueID

	resp, svcErr := h.svc.BulkSetIssueValues(req)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) ListIssueValues(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	resp, svcErr := h.svc.ListIssueValues(issueID)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) UpdateIssueValue(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	var req request.IssueCustomFieldValueUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	resp, svcErr := h.svc.UpdateIssueValue(issueID, fieldID, req.Value)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *CustomFieldHandler) DeleteIssueValue(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}
	fieldID, err := h.parseFieldID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid field ID"})
		return
	}

	if svcErr := h.svc.DeleteIssueValue(issueID, fieldID); svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom field value deleted"})
}

func (h *CustomFieldHandler) GetIssueFieldsWithValues(c *gin.Context) {
	issueID, err := h.parseIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	resp, svcErr := h.svc.GetIssueFieldsWithValues(issueID)
	if svcErr != nil {
		h.respondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}
