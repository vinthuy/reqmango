package handler

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	aiservice "github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type IntakeHandler struct {
	db  *gorm.DB
	svc *service.IntakeService
}

func NewIntakeHandler(db *gorm.DB, notifier service.IntakeNotifier) *IntakeHandler {
	return &IntakeHandler{db: db, svc: service.NewIntakeService(db, notifier)}
}

// requireProjectRole returns the project ID when the current user is an active
// project member with at least minRole (superusers always pass).
func (h *IntakeHandler) requireProjectRole(c *gin.Context, minRole int) (uint64, bool) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project id"})
		return 0, false
	}
	user := middleware.GetCurrentUser(c)
	if user.IsSuperuser {
		return projectID, true
	}
	var member model.ProjectMember
	if err := h.db.Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, user.ID, true).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied: you are not a member of this project"})
		return 0, false
	}
	if member.Role < minRole {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied: insufficient project role"})
		return 0, false
	}
	return projectID, true
}

func submitted(c *gin.Context, issue *model.Issue) {
	c.JSON(http.StatusCreated, gin.H{"id": issue.ID, "sequence_id": issue.SequenceID, "name": issue.Name, "status": service.IntakePending, "message": "Submitted for review"})
}

// Submit handles POST /api/v1/intake/:projectId — public web form.
func (h *IntakeHandler) Submit(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
		return
	}
	var req service.IntakeSubmission
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid email"})
			return
		}
	}
	issue, err := h.svc.Submit(projectID, "form", req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	submitted(c, issue)
}

// Info handles GET /api/v1/intake/:projectId/info — public project name for the form.
func (h *IntakeHandler) Info(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	info, err := h.svc.PublicInfo(projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, info)
}

// Webhook handles POST /api/v1/intake-channels/:token/webhook — generic JSON payload.
func (h *IntakeHandler) Webhook(c *gin.Context) {
	st, err := h.svc.SettingsByToken(c.Param("token"))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	var req struct {
		Title       string  `json:"title"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Body        string  `json:"body"`
		Priority    string  `json:"priority"`
		TypeID      *uint64 `json:"type_id"`
		Submitter   string  `json:"submitter"`
		Email       string  `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	in := service.IntakeSubmission{
		Name: firstNonEmpty(req.Title, req.Name), Description: firstNonEmpty(req.Description, req.Body),
		Priority: req.Priority, TypeID: req.TypeID, Submitter: req.Submitter, Email: req.Email,
	}
	issue, err := h.svc.Submit(st.ProjectID, "webhook", in)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	submitted(c, issue)
}

// Email handles POST /api/v1/intake-channels/:token/email — inbound-parse payloads
// (JSON or form fields: from, subject, text / body-plain / stripped-text).
func (h *IntakeHandler) Email(c *gin.Context) {
	st, err := h.svc.SettingsByToken(c.Param("token"))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	fields := map[string]string{}
	if strings.HasPrefix(c.ContentType(), "application/json") {
		var raw map[string]interface{}
		if err := c.ShouldBindJSON(&raw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		for k, v := range raw {
			if s, ok := v.(string); ok {
				fields[strings.ToLower(k)] = s
			}
		}
	} else {
		_ = c.Request.ParseMultipartForm(10 << 20)
		for _, k := range []string{"from", "subject", "text", "body-plain", "stripped-text", "sender"} {
			fields[k] = c.PostForm(k)
		}
	}
	in := service.IntakeSubmission{
		Name:        fields["subject"],
		Description: firstNonEmpty(fields["stripped-text"], fields["text"], fields["body-plain"]),
	}
	if in.Name == "" {
		in.Name = "(no subject)"
	}
	if addr, err := mail.ParseAddress(firstNonEmpty(fields["from"], fields["sender"])); err == nil {
		in.Submitter, in.Email = addr.Name, addr.Address
		if in.Submitter == "" {
			in.Submitter = addr.Address
		}
	}
	issue, err := h.svc.Submit(st.ProjectID, "email", in)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	submitted(c, issue)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// List handles GET /api/v1/projects/:projectId/intake?status=&source=&q=&limit=&offset=
func (h *IntakeHandler) List(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleGuest)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	res, err := h.svc.List(projectID, service.IntakeListQuery{
		Status: c.Query("status"), Source: c.Query("source"), Search: c.Query("q"), Limit: limit, Offset: offset,
	})
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Triage handles POST /api/v1/projects/:projectId/intake/:issueId/triage
func (h *IntakeHandler) Triage(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleMember)
	if !ok {
		return
	}
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid issue id"})
		return
	}
	var req service.IntakeTriageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	issue, err := h.svc.Triage(projectID, issueID, middleware.GetUserID(c), req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"issue_id": issue.ID, "status": issue.IntakeStatus, "state_id": issue.StateID})
}

// SetSpecAI wires the model used to draft specs from intake items.
func (h *IntakeHandler) SetSpecAI(ai *aiservice.AIService) { h.svc.SetSpecAI(ai) }

// GetSpec handles GET /api/v1/projects/:projectId/intake/:issueId/spec
func (h *IntakeHandler) GetSpec(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleGuest)
	if !ok {
		return
	}
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid issue id"})
		return
	}
	res, err := h.svc.Spec(projectID, issueID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// DraftSpec handles POST /api/v1/projects/:projectId/intake/:issueId/spec
func (h *IntakeHandler) DraftSpec(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleMember)
	if !ok {
		return
	}
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid issue id"})
		return
	}
	res, err := h.svc.DraftSpec(c.Request.Context(), projectID, issueID, middleware.GetUserID(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetSettings handles GET /api/v1/projects/:projectId/intake/settings
func (h *IntakeHandler) GetSettings(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleMember)
	if !ok {
		return
	}
	st, err := h.svc.Settings(projectID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

// UpdateSettings handles PUT /api/v1/projects/:projectId/intake/settings (project admins).
func (h *IntakeHandler) UpdateSettings(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleAdmin)
	if !ok {
		return
	}
	var req service.IntakeSettingsUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	st, err := h.svc.UpdateSettings(projectID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

// Metrics handles GET /api/v1/projects/:projectId/intake/metrics?days=
func (h *IntakeHandler) Metrics(c *gin.Context) {
	projectID, ok := h.requireProjectRole(c, common.RoleGuest)
	if !ok {
		return
	}
	days := 30
	if v := c.Query("days"); v != "" {
		d, err := strconv.Atoi(v)
		if err != nil || (d != 7 && d != 14 && d != 30 && d != 90) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "days must be one of 7, 14, 30, 90"})
			return
		}
		days = d
	}
	m, err := h.svc.Metrics(projectID, days)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}
