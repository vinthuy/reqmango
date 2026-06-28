package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type IntakeHandler struct {
	db *gorm.DB
}

func NewIntakeHandler(db *gorm.DB) *IntakeHandler { return &IntakeHandler{db: db} }

// Submit handles POST /api/v1/intake/:projectId — public, no auth.
func (h *IntakeHandler) Submit(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		TypeID      *uint64 `json:"type_id"`
		Submitter   string `json:"submitter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":err.Error()}); return }

	// Find workspace + default state
	var project model.Project
	if h.db.First(&project, projectID).Error != nil { c.JSON(404, gin.H{"message":"Project not found"}); return }
	var defaultState model.State
	if err := h.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&defaultState).Error; err != nil {
		// Fallback: use first available state
		if h.db.Where("project_id = ?", projectID).First(&defaultState).Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "No default state configured for project"})
			return
		}
	}

	intakeSource := "form"
	intakeStatus := "pending"
	priority := req.Priority
	if priority == "" { priority = "none" }

	issue := &model.Issue{
		Name: req.Name, DescriptionHTML: req.Description,
		Priority: priority, ProjectID: projectID,
		WorkspaceID: project.WorkspaceID, StateID: defaultState.ID,
		IntakeSource: &intakeSource, IntakeStatus: &intakeStatus,
		IssueTypeID: req.TypeID,
	}
	if err := h.db.Create(issue).Error; err != nil { c.JSON(500, gin.H{"message":"Failed to submit"}); return }
	c.JSON(201, gin.H{"id":issue.ID, "name":issue.Name, "status":"pending", "message":"Submitted for review"})
}

// ListPending handles GET /api/v1/projects/:projectId/intake — list pending intake items.
func (h *IntakeHandler) ListPending(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	var issues []model.Issue
	h.db.Where("project_id = ? AND intake_status = ?", projectID, "pending").Order("created_at DESC").Find(&issues)
	c.JSON(200, issues)
}

// Triage handles POST /api/v1/intake/:issueId/triage — accept or reject.
func (h *IntakeHandler) Triage(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req struct {
		Action   string  `json:"action" binding:"required"` // "accept" | "reject"
		Assignee *uint64 `json:"assignee_id"`
		StateID  *uint64 `json:"state_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":err.Error()}); return }

	var issue model.Issue
	if h.db.First(&issue, issueID).Error != nil { c.JSON(404, gin.H{"message":"Issue not found"}); return }

	status := req.Action
	if status == "accept" { status = "accepted" }
	updates := map[string]interface{}{"intake_status": status}
	if req.StateID != nil { updates["state_id"] = *req.StateID }
	if err := h.db.Model(&issue).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update issue"})
		return
	}
	if req.Assignee != nil {
		if err := h.db.Create(&model.IssueAssignee{IssueID: issueID, UserID: *req.Assignee}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to assign user"})
			return
		}
	}
	c.JSON(200, gin.H{"message":"Triage "+req.Action+" completed", "issue_id":issueID})
}
