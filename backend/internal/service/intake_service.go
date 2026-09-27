package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

const (
	IntakePending   = "pending"
	IntakeSnoozed   = "snoozed"
	IntakeAccepted  = "accepted"
	IntakeRejected  = "rejected"
	IntakeDuplicate = "duplicate"
)

// intakeVisibleClause hides untriaged, rejected and duplicate intake requests from work-item lists.
const intakeVisibleClause = "(issues.intake_status IS NULL OR issues.intake_status = 'accepted')"

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

var intakeSources = map[string]bool{"form": true, "webhook": true, "email": true}

type IntakeNotifier interface {
	NotifyIssueCreated(issue *model.Issue)
}

type IntakeService struct {
	db       *gorm.DB
	notifier IntakeNotifier
}

func NewIntakeService(db *gorm.DB, notifier IntakeNotifier) *IntakeService {
	return &IntakeService{db: db, notifier: notifier}
}

func newIntakeToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------- settings ----------

func (s *IntakeService) Settings(projectID uint64) (*model.ProjectIntakeSetting, error) {
	var st model.ProjectIntakeSetting
	err := s.db.Where("project_id = ?", projectID).First(&st).Error
	if err == nil {
		return &st, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, common.Internal("Failed to load intake settings")
	}
	st = model.ProjectIntakeSetting{
		ProjectID: projectID, FormEnabled: true, WebhookEnabled: true, EmailEnabled: true,
		Token: newIntakeToken(), SLAHours: 48,
	}
	if err := s.db.Create(&st).Error; err != nil {
		// Lost a creation race: read the row the other request created.
		if s.db.Where("project_id = ?", projectID).First(&st).Error == nil {
			return &st, nil
		}
		return nil, common.Internal("Failed to create intake settings")
	}
	return &st, nil
}

type IntakeSettingsUpdate struct {
	FormEnabled    *bool `json:"form_enabled"`
	WebhookEnabled *bool `json:"webhook_enabled"`
	EmailEnabled   *bool `json:"email_enabled"`
	SLAHours       *int  `json:"sla_hours"`
	RotateToken    bool  `json:"rotate_token"`
}

func (s *IntakeService) UpdateSettings(projectID uint64, in IntakeSettingsUpdate) (*model.ProjectIntakeSetting, error) {
	st, err := s.Settings(projectID)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if in.FormEnabled != nil {
		updates["form_enabled"] = *in.FormEnabled
	}
	if in.WebhookEnabled != nil {
		updates["webhook_enabled"] = *in.WebhookEnabled
	}
	if in.EmailEnabled != nil {
		updates["email_enabled"] = *in.EmailEnabled
	}
	if in.SLAHours != nil {
		if *in.SLAHours < 1 || *in.SLAHours > 720 {
			return nil, common.Validation("sla_hours must be between 1 and 720")
		}
		updates["sla_hours"] = *in.SLAHours
	}
	if in.RotateToken {
		updates["token"] = newIntakeToken()
	}
	if len(updates) > 0 {
		if err := s.db.Model(st).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update intake settings")
		}
	}
	return s.Settings(projectID)
}

// ---------- submission ----------

type IntakeSubmission struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	TypeID      *uint64 `json:"type_id"`
	Submitter   string  `json:"submitter"`
	Email       string  `json:"email"`
}

type IntakeProjectInfo struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	FormEnabled bool   `json:"form_enabled"`
}

func (s *IntakeService) PublicInfo(projectID uint64) (*IntakeProjectInfo, error) {
	var p model.Project
	if err := s.db.First(&p, projectID).Error; err != nil || p.ArchivedAt != nil {
		return nil, common.ProjectNotFound()
	}
	st, err := s.Settings(projectID)
	if err != nil {
		return nil, err
	}
	return &IntakeProjectInfo{ID: p.ID, Name: p.Name, Identifier: p.Identifier, FormEnabled: st.FormEnabled}, nil
}

func (s *IntakeService) SettingsByToken(token string) (*model.ProjectIntakeSetting, error) {
	if len(token) < 16 {
		return nil, common.NotFound("Intake channel not found")
	}
	var st model.ProjectIntakeSetting
	if err := s.db.Where("token = ?", token).First(&st).Error; err != nil {
		return nil, common.NotFound("Intake channel not found")
	}
	return &st, nil
}

var intakePriorities = map[string]bool{"none": true, "low": true, "medium": true, "high": true, "urgent": true}

func (s *IntakeService) Submit(projectID uint64, source string, in IntakeSubmission) (*model.Issue, error) {
	if !intakeSources[source] {
		return nil, common.BadRequest("unknown intake source")
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, common.Validation("name is required")
	}
	if len([]rune(in.Name)) > 255 {
		in.Name = string([]rune(in.Name)[:255])
	}
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil || project.ArchivedAt != nil {
		return nil, common.ProjectNotFound()
	}
	st, err := s.Settings(projectID)
	if err != nil {
		return nil, err
	}
	enabled := map[string]bool{"form": st.FormEnabled, "webhook": st.WebhookEnabled, "email": st.EmailEnabled}[source]
	if !enabled {
		return nil, common.Forbidden("This intake channel is disabled")
	}

	var state model.State
	if err := s.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&state).Error; err != nil {
		if err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").First(&state).Error; err != nil {
			return nil, common.Internal("No default state configured for project")
		}
	}
	priority := strings.ToLower(strings.TrimSpace(in.Priority))
	if !intakePriorities[priority] {
		priority = "none"
	}
	if in.TypeID != nil {
		var cnt int64
		s.db.Model(&model.IssueType{}).Where("id = ? AND workspace_id = ?", *in.TypeID, project.WorkspaceID).Count(&cnt)
		if cnt == 0 {
			in.TypeID = nil
		}
	}

	// Public channels are unauthenticated: never store submitter-supplied markup.
	desc := strings.TrimSpace(in.Description)
	if desc == "" {
		desc = "<p></p>"
	} else {
		desc = "<p>" + strings.ReplaceAll(html.EscapeString(desc), "\n", "<br>") + "</p>"
	}
	src, status := source, IntakePending
	issue := &model.Issue{
		Name: in.Name, DescriptionHTML: desc, Priority: priority,
		ProjectID: projectID, WorkspaceID: project.WorkspaceID, StateID: state.ID,
		IntakeSource: &src, IntakeStatus: &status, IssueTypeID: in.TypeID,
		IntakeSubmitter: optStr(in.Submitter, 255), IntakeEmail: optStr(in.Email, 255),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var maxSeq int
		tx.Model(&model.Issue{}).Where("project_id = ?", projectID).Select("COALESCE(MAX(sequence_id), 0)").Scan(&maxSeq)
		issue.SequenceID = maxSeq + 1
		if err := tx.Create(issue).Error; err != nil {
			return err
		}
		verb, field, nv := "created", "intake_source", source
		return tx.Create(&model.IssueActivity{IssueID: &issue.ID, Verb: verb, Field: &field, NewValue: &nv}).Error
	})
	if err != nil {
		return nil, common.Internal("Failed to submit")
	}
	if s.notifier != nil {
		s.notifier.NotifyIssueCreated(issue)
	}
	return issue, nil
}

func optStr(v string, max int) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	if r := []rune(v); len(r) > max {
		v = string(r[:max])
	}
	return &v
}

// ---------- listing ----------

type IntakeItem struct {
	ID              uint64     `json:"id"`
	SequenceID      int        `json:"sequence_id"`
	Name            string     `json:"name"`
	DescriptionHTML string     `json:"description_html"`
	Priority        string     `json:"priority"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	Submitter       *string    `json:"submitter"`
	Email           *string    `json:"email"`
	CreatedAt       time.Time  `json:"created_at"`
	TriagedAt       *time.Time `json:"triaged_at"`
	TriagedByName   *string    `json:"triaged_by_name"`
	SnoozedUntil    *time.Time `json:"snoozed_until"`
	DuplicateOf     *uint64    `json:"duplicate_of"`
	DuplicateSeq    *int       `json:"duplicate_sequence_id"`
	DuplicateName   *string    `json:"duplicate_name"`
	Note            *string    `json:"note"`
	StateID         uint64     `json:"state_id"`
	AgeHours        float64    `json:"age_hours"`
	SLAOverdue      bool       `json:"sla_overdue"`
}

type IntakeListResult struct {
	Items    []IntakeItem     `json:"items"`
	Total    int64            `json:"total"`
	Counts   map[string]int64 `json:"counts"`
	SLAHours int              `json:"sla_hours"`
}

type IntakeListQuery struct {
	Status string
	Source string
	Search string
	Limit  int
	Offset int
}

func (s *IntakeService) scopeStatus(q *gorm.DB, status string, now time.Time) *gorm.DB {
	switch status {
	case IntakePending:
		// Snoozes expire back into the pending queue.
		return q.Where("(issues.intake_status = ? OR (issues.intake_status = ? AND (issues.intake_snoozed_until IS NULL OR issues.intake_snoozed_until <= ?)))", IntakePending, IntakeSnoozed, now)
	case IntakeSnoozed:
		return q.Where("issues.intake_status = ? AND issues.intake_snoozed_until > ?", IntakeSnoozed, now)
	case IntakeAccepted, IntakeRejected, IntakeDuplicate:
		return q.Where("issues.intake_status = ?", status)
	}
	return q
}

func (s *IntakeService) List(projectID uint64, lq IntakeListQuery) (*IntakeListResult, error) {
	st, err := s.Settings(projectID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if lq.Status == "" {
		lq.Status = IntakePending
	}
	if lq.Limit <= 0 || lq.Limit > 100 {
		lq.Limit = 50
	}
	base := func() *gorm.DB {
		q := s.db.Model(&model.Issue{}).Where("issues.project_id = ? AND issues.intake_source IS NOT NULL AND issues.deleted_at IS NULL", projectID)
		if lq.Source != "" {
			q = q.Where("issues.intake_source = ?", lq.Source)
		}
		if t := strings.TrimSpace(lq.Search); t != "" {
			like := "%" + strings.ToLower(t) + "%"
			q = q.Where("(LOWER(issues.name) LIKE ? OR LOWER(COALESCE(issues.intake_submitter,'')) LIKE ? OR LOWER(COALESCE(issues.intake_email,'')) LIKE ?)", like, like, like)
		}
		return q
	}

	counts := map[string]int64{}
	for _, k := range []string{IntakePending, IntakeSnoozed, IntakeAccepted, IntakeRejected, IntakeDuplicate} {
		var n int64
		s.scopeStatus(base(), k, now).Count(&n)
		counts[k] = n
	}

	var issues []model.Issue
	q := s.scopeStatus(base(), lq.Status, now)
	var total int64
	q.Count(&total)
	order := "issues.created_at ASC"
	if lq.Status != IntakePending {
		order = "COALESCE(issues.intake_triaged_at, issues.created_at) DESC"
	}
	if err := s.scopeStatus(base(), lq.Status, now).Order(order).Limit(lq.Limit).Offset(lq.Offset).Find(&issues).Error; err != nil {
		return nil, common.Internal("Failed to list intake")
	}

	userIDs, dupIDs := []uint64{}, []uint64{}
	for _, is := range issues {
		if is.IntakeTriagedBy != nil {
			userIDs = append(userIDs, *is.IntakeTriagedBy)
		}
		if is.IntakeDuplicateOf != nil {
			dupIDs = append(dupIDs, *is.IntakeDuplicateOf)
		}
	}
	names := map[uint64]string{}
	if len(userIDs) > 0 {
		var users []model.User
		s.db.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			n := u.DisplayName
			if n == "" {
				n = u.Email
			}
			names[u.ID] = n
		}
	}
	dups := map[uint64]model.Issue{}
	if len(dupIDs) > 0 {
		var ds []model.Issue
		s.db.Select("id, sequence_id, name").Where("id IN ?", dupIDs).Find(&ds)
		for _, d := range ds {
			dups[d.ID] = d
		}
	}

	items := make([]IntakeItem, 0, len(issues))
	for _, is := range issues {
		it := IntakeItem{
			ID: is.ID, SequenceID: is.SequenceID, Name: is.Name, DescriptionHTML: is.DescriptionHTML,
			Priority: is.Priority, Submitter: is.IntakeSubmitter, Email: is.IntakeEmail,
			CreatedAt: is.CreatedAt, TriagedAt: is.IntakeTriagedAt, SnoozedUntil: is.IntakeSnoozedUntil,
			DuplicateOf: is.IntakeDuplicateOf, Note: is.IntakeNote, StateID: is.StateID,
		}
		if is.IntakeSource != nil {
			it.Source = *is.IntakeSource
		}
		if is.IntakeStatus != nil {
			it.Status = *is.IntakeStatus
		}
		if it.Status == IntakeSnoozed && (is.IntakeSnoozedUntil == nil || !is.IntakeSnoozedUntil.After(now)) {
			it.Status = IntakePending
		}
		if is.IntakeTriagedBy != nil {
			if n, ok := names[*is.IntakeTriagedBy]; ok {
				it.TriagedByName = &n
			}
		}
		if is.IntakeDuplicateOf != nil {
			if d, ok := dups[*is.IntakeDuplicateOf]; ok {
				seq, nm := d.SequenceID, d.Name
				it.DuplicateSeq, it.DuplicateName = &seq, &nm
			}
		}
		end := now
		if is.IntakeTriagedAt != nil && it.Status != IntakePending && it.Status != IntakeSnoozed {
			end = *is.IntakeTriagedAt
		}
		it.AgeHours = math.Round(end.Sub(is.CreatedAt).Hours()*10) / 10
		it.SLAOverdue = it.Status == IntakePending && it.AgeHours > float64(st.SLAHours)
		items = append(items, it)
	}
	return &IntakeListResult{Items: items, Total: total, Counts: counts, SLAHours: st.SLAHours}, nil
}

// ---------- triage ----------

type IntakeTriageInput struct {
	Action      string  `json:"action"` // accept | reject | snooze | duplicate | reopen
	StateID     *uint64 `json:"state_id"`
	AssigneeID  *uint64 `json:"assignee_id"`
	Reason      string  `json:"reason"`
	SnoozeHours int     `json:"snooze_hours"`
	DuplicateOf *uint64 `json:"duplicate_of"`
}

func (s *IntakeService) Triage(projectID, issueID, actorID uint64, in IntakeTriageInput) (*model.Issue, error) {
	var issue model.Issue
	if err := s.db.Where("id = ? AND project_id = ? AND intake_source IS NOT NULL", issueID, projectID).First(&issue).Error; err != nil {
		return nil, common.NotFound("Intake item not found")
	}
	old := ""
	if issue.IntakeStatus != nil {
		old = *issue.IntakeStatus
	}
	now := time.Now()
	updates := map[string]interface{}{"intake_triaged_at": now, "intake_triaged_by": actorID}
	var newStatus string
	var note *string
	var mergeTarget *model.Issue

	switch in.Action {
	case "accept":
		newStatus = IntakeAccepted
		if in.StateID != nil {
			var cnt int64
			s.db.Model(&model.State{}).Where("id = ? AND project_id = ?", *in.StateID, projectID).Count(&cnt)
			if cnt == 0 {
				return nil, common.Validation("state does not belong to this project")
			}
			updates["state_id"] = *in.StateID
		}
		if in.AssigneeID != nil {
			var cnt int64
			s.db.Model(&model.ProjectMember{}).Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, *in.AssigneeID, true).Count(&cnt)
			if cnt == 0 {
				return nil, common.Validation("assignee is not a member of this project")
			}
		}
		updates["intake_snoozed_until"] = nil
	case "reject":
		newStatus = IntakeRejected
		if r := strings.TrimSpace(in.Reason); r != "" {
			note = &r
		}
		updates["intake_note"] = note
		updates["intake_snoozed_until"] = nil
		if id := s.cancelledStateID(projectID); id != 0 {
			updates["state_id"] = id
		}
	case "snooze":
		if in.SnoozeHours < 1 || in.SnoozeHours > 24*90 {
			return nil, common.Validation("snooze_hours must be between 1 and 2160")
		}
		if old != IntakePending && old != IntakeSnoozed {
			return nil, common.Validation("only pending items can be snoozed")
		}
		newStatus = IntakeSnoozed
		updates["intake_snoozed_until"] = now.Add(time.Duration(in.SnoozeHours) * time.Hour)
		// Snoozing isn't a triage decision; keep time-to-triage honest.
		delete(updates, "intake_triaged_at")
		delete(updates, "intake_triaged_by")
	case "duplicate":
		if in.DuplicateOf == nil || *in.DuplicateOf == issueID {
			return nil, common.Validation("duplicate_of must reference another work item")
		}
		var target model.Issue
		if err := s.db.Where("id = ? AND project_id = ?", *in.DuplicateOf, projectID).First(&target).Error; err != nil {
			return nil, common.Validation("duplicate target not found in this project")
		}
		if target.IntakeStatus != nil && *target.IntakeStatus != IntakeAccepted {
			return nil, common.Validation("duplicate target must be an accepted work item")
		}
		newStatus = IntakeDuplicate
		mergeTarget = &target
		updates["intake_duplicate_of"] = target.ID
		updates["intake_snoozed_until"] = nil
		if id := s.cancelledStateID(projectID); id != 0 {
			updates["state_id"] = id
		}
	case "reopen":
		if old == IntakePending {
			return nil, common.Validation("item is already pending")
		}
		if old == IntakeAccepted {
			return nil, common.Validation("accepted items cannot be reopened")
		}
		newStatus = IntakePending
		updates = map[string]interface{}{
			"intake_triaged_at": nil, "intake_triaged_by": nil, "intake_snoozed_until": nil,
			"intake_duplicate_of": nil, "intake_note": nil,
		}
		if id := s.defaultStateID(projectID); id != 0 {
			updates["state_id"] = id
		}
	default:
		return nil, common.BadRequest("unknown triage action")
	}
	if old == IntakeAccepted && in.Action != "accept" {
		return nil, common.Validation("item has already been accepted")
	}
	updates["intake_status"] = newStatus

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Issue{}).Where("id = ?", issue.ID).Updates(updates).Error; err != nil {
			return err
		}
		if in.Action == "accept" && in.AssigneeID != nil {
			var cnt int64
			tx.Model(&model.IssueAssignee{}).Where("issue_id = ? AND user_id = ?", issue.ID, *in.AssigneeID).Count(&cnt)
			if cnt == 0 {
				if err := tx.Create(&model.IssueAssignee{IssueID: issue.ID, UserID: *in.AssigneeID}).Error; err != nil {
					return err
				}
			}
		}
		field := "intake_status"
		ov, nv := old, newStatus
		act := &model.IssueActivity{IssueID: &issue.ID, Verb: "updated", Field: &field, OldValue: &ov, NewValue: &nv, ActorID: &actorID}
		act.CreatedByID = &actorID
		if note != nil {
			act.Comment = note
		}
		if err := tx.Create(act).Error; err != nil {
			return err
		}
		if mergeTarget != nil {
			// Comment bodies are plain text; the client escapes them on render.
			who := "—"
			if issue.IntakeSubmitter != nil {
				who = *issue.IntakeSubmitter
			}
			if issue.IntakeEmail != nil {
				who += " <" + *issue.IntakeEmail + ">"
			}
			text := strings.TrimSpace(html.UnescapeString(htmlTagRe.ReplaceAllString(
				strings.ReplaceAll(issue.DescriptionHTML, "<br>", "\n"), "")))
			body := fmt.Sprintf("Merged intake request #%d: %s\nFrom: %s", issue.SequenceID, issue.Name, who)
			if text != "" {
				body += "\n\n" + text
			}
			if err := tx.Create(&model.Comment{IssueID: mergeTarget.ID, AuthorID: &actorID, Body: body}).Error; err != nil {
				return err
			}
			mf := "intake_merged"
			mv := fmt.Sprintf("%d", issue.ID)
			if err := tx.Create(&model.IssueActivity{IssueID: &mergeTarget.ID, Verb: "updated", Field: &mf, NewValue: &mv, ActorID: &actorID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to triage intake item")
	}
	s.db.First(&issue, issue.ID)
	return &issue, nil
}

func (s *IntakeService) cancelledStateID(projectID uint64) uint64 {
	var st model.State
	if s.db.Where("project_id = ? AND \"group\" = ?", projectID, "cancelled").Order("sequence ASC").First(&st).Error == nil {
		return st.ID
	}
	return 0
}

func (s *IntakeService) defaultStateID(projectID uint64) uint64 {
	var st model.State
	if s.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&st).Error == nil {
		return st.ID
	}
	return 0
}

// ---------- metrics ----------

type IntakeSourceStat struct {
	Source   string `json:"source"`
	Received int64  `json:"received"`
	Accepted int64  `json:"accepted"`
}

type IntakeTrendPoint struct {
	Date     string `json:"date"`
	Received int64  `json:"received"`
	Triaged  int64  `json:"triaged"`
}

type IntakeMetrics struct {
	Days            int                `json:"days"`
	SLAHours        int                `json:"sla_hours"`
	Received        int64              `json:"received"`
	Accepted        int64              `json:"accepted"`
	Rejected        int64              `json:"rejected"`
	Duplicate       int64              `json:"duplicate"`
	PendingNow      int64              `json:"pending_now"`
	SnoozedNow      int64              `json:"snoozed_now"`
	OverdueNow      int64              `json:"overdue_now"`
	AcceptanceRate  *float64           `json:"acceptance_rate"`
	AvgTriageHours  *float64           `json:"avg_triage_hours"`
	MedianTriageHrs *float64           `json:"median_triage_hours"`
	SLAMetRate      *float64           `json:"sla_met_rate"`
	BySource        []IntakeSourceStat `json:"by_source"`
	Trend           []IntakeTrendPoint `json:"trend"`
}

func (s *IntakeService) Metrics(projectID uint64, days int) (*IntakeMetrics, error) {
	st, err := s.Settings(projectID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	loc := now.Location()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(days - 1))
	m := &IntakeMetrics{Days: days, SLAHours: st.SLAHours, BySource: []IntakeSourceStat{}, Trend: []IntakeTrendPoint{}}

	base := func() *gorm.DB {
		return s.db.Model(&model.Issue{}).Where("issues.project_id = ? AND issues.intake_source IS NOT NULL AND issues.deleted_at IS NULL", projectID)
	}

	var received []model.Issue
	base().Select("id, created_at, intake_source, intake_status, intake_triaged_at").Where("issues.created_at >= ?", startDay).Find(&received)
	m.Received = int64(len(received))
	srcIdx := map[string]*IntakeSourceStat{}
	for _, is := range received {
		src := "form"
		if is.IntakeSource != nil {
			src = *is.IntakeSource
		}
		ss, ok := srcIdx[src]
		if !ok {
			ss = &IntakeSourceStat{Source: src}
			srcIdx[src] = ss
		}
		ss.Received++
		if is.IntakeStatus != nil && *is.IntakeStatus == IntakeAccepted {
			ss.Accepted++
		}
	}
	for _, k := range []string{"form", "webhook", "email"} {
		if ss, ok := srcIdx[k]; ok {
			m.BySource = append(m.BySource, *ss)
		}
	}

	// Decisions made inside the window, regardless of when the request arrived.
	var triaged []model.Issue
	base().Select("id, created_at, intake_status, intake_triaged_at").
		Where("issues.intake_triaged_at >= ? AND issues.intake_status IN ?", startDay, []string{IntakeAccepted, IntakeRejected, IntakeDuplicate}).
		Find(&triaged)
	hours := make([]float64, 0, len(triaged))
	var metSLA int
	for _, is := range triaged {
		switch *is.IntakeStatus {
		case IntakeAccepted:
			m.Accepted++
		case IntakeRejected:
			m.Rejected++
		case IntakeDuplicate:
			m.Duplicate++
		}
		h := is.IntakeTriagedAt.Sub(is.CreatedAt).Hours()
		if h < 0 {
			h = 0
		}
		hours = append(hours, h)
		if h <= float64(st.SLAHours) {
			metSLA++
		}
	}
	if n := len(hours); n > 0 {
		sum := 0.0
		for _, h := range hours {
			sum += h
		}
		avg := round1(sum / float64(n))
		sort.Float64s(hours)
		med := hours[n/2]
		if n%2 == 0 {
			med = (hours[n/2-1] + hours[n/2]) / 2
		}
		med = round1(med)
		rate := round1(float64(m.Accepted) * 100 / float64(n))
		met := round1(float64(metSLA) * 100 / float64(n))
		m.AvgTriageHours, m.MedianTriageHrs, m.AcceptanceRate, m.SLAMetRate = &avg, &med, &rate, &met
	}

	s.scopeStatus(base(), IntakePending, now).Count(&m.PendingNow)
	s.scopeStatus(base(), IntakeSnoozed, now).Count(&m.SnoozedNow)
	s.scopeStatus(base(), IntakePending, now).Where("issues.created_at < ?", now.Add(-time.Duration(st.SLAHours)*time.Hour)).Count(&m.OverdueNow)

	buckets := map[string]*IntakeTrendPoint{}
	for d := 0; d < days; d++ {
		key := startDay.AddDate(0, 0, d).Format("2006-01-02")
		p := IntakeTrendPoint{Date: key}
		m.Trend = append(m.Trend, p)
	}
	for i := range m.Trend {
		buckets[m.Trend[i].Date] = &m.Trend[i]
	}
	for _, is := range received {
		if p, ok := buckets[is.CreatedAt.In(loc).Format("2006-01-02")]; ok {
			p.Received++
		}
	}
	for _, is := range triaged {
		if p, ok := buckets[is.IntakeTriagedAt.In(loc).Format("2006-01-02")]; ok {
			p.Triaged++
		}
	}
	return m, nil
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
