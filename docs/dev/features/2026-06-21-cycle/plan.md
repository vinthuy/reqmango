# Cycle（周期）管理功能 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement full Cycle management (CRUD + status transitions + progress analysis + burndown chart) for the ReqManPy project, following the patterns established by the existing Issue module.

**Architecture:** Go backend (Gin + GORM) with service/handler/router layers mirroring the Issue pattern. Vue 3 frontend with Pinia store driving CycleList/CycleDetail/CycleCreate views, integrated into the existing Project tab layout.

**Tech Stack:** Go + Gin + GORM + PostgreSQL (backend), Vue 3 + TypeScript + Pinia + Tailwind CSS (frontend)

**Spec reference:** `docs/pages/2026-06-21-cycle-feature-design.md`

---

## File Structure Map

### Go Backend (create/modify)
| File | Responsibility |
|------|---------------|
| `internal/dto/request/cycle.go` | Cycle create/update request structs |
| `internal/dto/response/cycle.go` | Cycle response, progress, statistics, burndown structs |
| `internal/model/cycle.go` | DB model with CancelledAt field |
| `internal/service/cycle_service.go` | All business logic |
| `internal/handler/cycle_handler.go` | HTTP handler methods |
| `internal/router/router.go` | Route registration |

### Vue Frontend (create/modify)
| File | Responsibility |
|------|---------------|
| `stores/cycle.ts` | Pinia store �?all state + actions |
| `api/cycle.ts` | API client �?adjusted to match Go routes |
| `types/cycle.ts` | TypeScript types �?aligned with Go responses |
| `components/CycleCard.vue` | Card �?updated for new data shape |
| `components/CycleList.vue` | List �?store-driven with pagination |
| `components/CycleDetailPanel.vue` | Side panel for Project tab |
| `components/CycleProgressCard.vue` | Progress stats card group |
| `components/CycleBurndownChart.vue` | SVG burndown chart |
| `views/CycleCreate.vue` | Two-step creation wizard |
| `views/CycleDetail.vue` | Full detail page |
| `router/index.ts` | Two new routes |
| `views/Project.vue` | Integrate CycleDetailPanel |

---

### Task 1: Go DTO �?Request/Response Structs

**Files:**
- Create: `backend/internal/dto/request/cycle.go`
- Create: `backend/internal/dto/response/cycle.go`

- [ ] **Step 1: Create request DTO**

```go
// backend/internal/dto/request/cycle.go
package request

// CycleCreateRequest is the request body for creating a cycle.
// ProjectID is set from the URL path by the handler, not from JSON body.
type CycleCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	StartDate   string  `json:"start_date" binding:"required"`   // RFC3339 date
	EndDate     *string `json:"end_date"`                         // RFC3339 date, nullable
	Timezone    string  `json:"timezone"`
	ProjectID   uint64  `json:"-"`                                // Set from URL path by handler
}

// CycleUpdateRequest is the request body for updating a cycle.
type CycleUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`   // RFC3339 date
	EndDate     *string `json:"end_date"`     // RFC3339 date
}
```

- [ ] **Step 2: Create response DTO**

```go
// backend/internal/dto/response/cycle.go
package response

import "time"

// CycleResponse is the full cycle representation returned by the API.
type CycleResponse struct {
	ID              uint64      `json:"id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description"`
	Status          string      `json:"status"`           // computed: upcoming|active|completed|cancelled
	Progress        float64     `json:"progress"`          // 0-100
	TotalIssues     int64       `json:"total_issues"`
	CompletedIssues int64       `json:"completed_issues"`
	StartDate       string      `json:"start_date"`        // "2006-01-02"
	EndDate         *string     `json:"end_date"`          // "2006-01-02", nullable
	ProjectID       uint64      `json:"project_id"`
	WorkspaceID     uint64      `json:"workspace_id"`
	OwnedBy         *UserLite   `json:"owned_by"`
	Project         *ProjectLite `json:"project"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	CreatedByID     *uint64     `json:"created_by_id"`
	UpdatedByID     *uint64     `json:"updated_by_id"`
}

// CycleLite is a compact cycle representation.
type CycleLite struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// CycleProgress represents cycle progress statistics.
type CycleProgress struct {
	CycleID         uint64            `json:"cycle_id"`
	CycleName       string            `json:"cycle_name"`
	TotalIssues     int64             `json:"total_issues"`
	CompletedIssues int64             `json:"completed_issues"`
	Progress        float64           `json:"progress"`
	StateBreakdown  []StateBreakdown  `json:"state_breakdown"`
}

// StateBreakdown shows issue count per state.
type StateBreakdown struct {
	State string `json:"state"`
	Group string `json:"group"`
	Count int64  `json:"count"`
}

// CycleStatistics extends progress with priority breakdown and date info.
type CycleStatistics struct {
	CycleProgress
	PriorityBreakdown map[string]int64 `json:"priority_breakdown"`
	IssueStats        IssueStats       `json:"issue_stats"`
	DateRange         DateRange        `json:"date_range"`
}

type IssueStats struct {
	Total          int64 `json:"total"`
	WithStartDate  int64 `json:"with_start_date"`
	WithTargetDate int64 `json:"with_target_date"`
}

type DateRange struct {
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// BurndownData represents burndown chart data.
type BurndownData struct {
	CycleID         uint64  `json:"cycle_id"`
	CycleName       string  `json:"cycle_name"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	TotalIssues     int64   `json:"total_issues"`
	TotalDays       int     `json:"total_days"`
	DaysElapsed     int     `json:"days_elapsed"`
	IdealDailyBurn  float64 `json:"ideal_daily_burn"`
	IdealRemaining  float64 `json:"ideal_remaining"`
	ActualCompleted int64   `json:"actual_completed"`
	ActualRemaining int64   `json:"actual_remaining"`
	IsOnTrack       bool    `json:"is_on_track"`
}

// UserLite is a compact user representation (reuse from issue response if possible).
// Note: This type already exists in backend/internal/dto/response/auth.go or issue.go.
// If it does not exist, define it here. Check before creating.
```

- [ ] **Step 3: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Should compile (no usages yet, but no syntax errors)

- [ ] **Step 4: Commit**

```bash
git add backend/internal/dto/request/cycle.go backend/internal/dto/response/cycle.go
git commit -m "feat(cycle): add DTO request/response structs"
```

---

### Task 2: Go Model �?Add CancelledAt Field

**Files:**
- Modify: `backend/internal/model/cycle.go`

- [ ] **Step 1: Update Cycle model**

Replace the content of `backend/internal/model/cycle.go`:

```go
package model

import "time"

// Cycle represents an iteration/sprint cycle.
type Cycle struct {
	BaseModel

	Name        string     `gorm:"size:255;not null" json:"name"`
	Description *string    `gorm:"size:1000" json:"description"`
	StartDate   time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate     *time.Time `gorm:"type:date" json:"end_date"`
	CompletedAt *time.Time `json:"completed_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null" json:"workspace_id"`

	// Relationships
	Project    Project      `gorm:"foreignKey:ProjectID" json:"-"`
	IssueLinks []IssueCycle `gorm:"foreignKey:CycleID" json:"-"`
}

func (Cycle) TableName() string {
	return "cycles"
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/cycle.go
git commit -m "feat(cycle): add CancelledAt field to Cycle model"
```

---

### Task 3: Go CycleService �?Full Business Logic

**Files:**
- Modify: `backend/internal/service/cycle_service.go`

- [ ] **Step 1: Write the complete service**

Replace the entire content of `backend/internal/service/cycle_service.go`:

```go
package service

import (
	"time"

	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/dto/response"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type CycleService struct {
	db *gorm.DB
}

func NewCycleService(db *gorm.DB) *CycleService {
	return &CycleService{db: db}
}

// ==================== Helpers ====================

// computeStatus infers cycle status from timestamps.
func computeStatus(cycle *model.Cycle) string {
	if cycle.CancelledAt != nil {
		return "cancelled"
	}
	if cycle.CompletedAt != nil {
		return "completed"
	}
	today := time.Now().Truncate(24 * time.Hour)
	start := cycle.StartDate.Truncate(24 * time.Hour)
	if today.Before(start) {
		return "upcoming"
	}
	return "active"
}

// dateStr formats a *time.Time as "2006-01-02", returns nil if nil.
func dateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// countIssuesInCycle returns total and completed issue counts for a cycle.
func (s *CycleService) countIssues(cycleID uint64) (total, completed int64, err error) {
	if err = s.db.Model(&model.IssueCycle{}).
		Where("cycle_id = ?", cycleID).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = s.db.Model(&model.IssueCycle{}).
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issue_cycles.cycle_id = ? AND states.group = ?", cycleID, common.StateGroupCompleted).
		Count(&completed).Error; err != nil {
		return total, 0, err
	}
	return total, completed, nil
}

// buildResponse constructs a CycleResponse from a Cycle model.
func (s *CycleService) buildResponse(cycle *model.Cycle) *response.CycleResponse {
	total, completed, _ := s.countIssues(cycle.ID)
	progress := 0.0
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}

	var endDateStr *string
	if cycle.EndDate != nil {
		s := cycle.EndDate.Format("2006-01-02")
		endDateStr = &s
	}

	resp := &response.CycleResponse{
		ID:              cycle.ID,
		Name:            cycle.Name,
		Description:     cycle.Description,
		Status:          computeStatus(cycle),
		Progress:        float64(int(progress*100)) / 100, // 2 decimal places
		TotalIssues:     total,
		CompletedIssues: completed,
		StartDate:       cycle.StartDate.Format("2006-01-02"),
		EndDate:         endDateStr,
		ProjectID:       cycle.ProjectID,
		WorkspaceID:     cycle.WorkspaceID,
		CreatedAt:       cycle.CreatedAt,
		UpdatedAt:       cycle.UpdatedAt,
		CreatedByID:     cycle.CreatedByID,
		UpdatedByID:     cycle.UpdatedByID,
	}

	// Load project
	var project model.Project
	if err := s.db.First(&project, cycle.ProjectID).Error; err == nil {
		resp.Project = &response.ProjectLite{
			ID:         project.ID,
			Name:       project.Name,
			Identifier: project.Identifier,
		}
	}

	// Load owner
	if cycle.CreatedByID != nil {
		var user model.User
		if err := s.db.First(&user, *cycle.CreatedByID).Error; err == nil {
			resp.OwnedBy = &response.UserLite{
				ID:          user.ID,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			}
		}
	}

	return resp
}

// ==================== CRUD ====================

// Create creates a new cycle.
func (s *CycleService) Create(workspaceID, userID uint64, req *request.CycleCreateRequest) (*response.CycleResponse, error) {
	// Validate project exists
	var project model.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	// Parse start date
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, common.BadRequest("Invalid start_date format, use RFC3339")
	}
	startDate = startDate.Truncate(24 * time.Hour)

	// Parse end date (optional)
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return nil, common.BadRequest("Invalid end_date format, use RFC3339")
		}
		t = t.Truncate(24 * time.Hour)
		endDate = &t
	}

	// Validate date range
	if endDate != nil && !startDate.Before(*endDate) && !startDate.Equal(*endDate) {
		return nil, common.BadRequest("Start date must be before or equal to end date")
	}

	cycle := &model.Cycle{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		ProjectID:   req.ProjectID,
		WorkspaceID: workspaceID,
	}
	cycle.CreatedByID = &userID

	if err := s.db.Create(cycle).Error; err != nil {
		return nil, common.Internal("Failed to create cycle")
	}

	return s.buildResponse(cycle), nil
}

// GetByID returns a cycle by ID.
func (s *CycleService) GetByID(cycleID uint64) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}
	return s.buildResponse(&cycle), nil
}

// ListByProject returns cycles for a project with optional status filter and pagination.
func (s *CycleService) ListByProject(projectID uint64, status string, limit, offset int) ([]response.CycleResponse, int64, error) {
	// Validate project exists
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, 0, common.NotFound("Project not found")
	}

	baseQuery := s.db.Model(&model.Cycle{}).Where("project_id = ?", projectID)

	// Get all non-deleted cycles first
	var allCycles []model.Cycle
	if err := baseQuery.Order("start_date DESC").Find(&allCycles).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	// Filter by status (computed, so we must do it in-memory)
	var filtered []model.Cycle
	for _, c := range allCycles {
		if status == "" || computeStatus(&c) == status {
			filtered = append(filtered, c)
		}
	}

	total := int64(len(filtered))

	// Apply pagination
	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	page := filtered[start:end]
	result := make([]response.CycleResponse, len(page))
	for i, c := range page {
		result[i] = *s.buildResponse(&c)
	}

	return result, total, nil
}

// Update updates a cycle.
func (s *CycleService) Update(cycleID, userID uint64, req *request.CycleUpdateRequest) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	if req.Name != nil {
		cycle.Name = *req.Name
	}
	if req.Description != nil {
		cycle.Description = req.Description
	}
	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			return nil, common.BadRequest("Invalid start_date format")
		}
		t = t.Truncate(24 * time.Hour)
		cycle.StartDate = t
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			cycle.EndDate = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.EndDate)
			if err != nil {
				return nil, common.BadRequest("Invalid end_date format")
			}
			t = t.Truncate(24 * time.Hour)
			cycle.EndDate = &t
		}
	}

	// Validate date range
	if cycle.EndDate != nil && !cycle.StartDate.Before(*cycle.EndDate) && !cycle.StartDate.Equal(*cycle.EndDate) {
		return nil, common.BadRequest("Start date must be before or equal to end date")
	}

	cycle.UpdatedByID = &userID
	if err := s.db.Save(&cycle).Error; err != nil {
		return nil, common.Internal("Failed to update cycle")
	}

	return s.buildResponse(&cycle), nil
}

// Delete performs a soft delete on a cycle.
func (s *CycleService) Delete(cycleID uint64) error {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return common.NotFound("Cycle not found")
	}
	// Remove all issue-cycle associations first
	s.db.Where("cycle_id = ?", cycleID).Delete(&model.IssueCycle{})
	return s.db.Delete(&cycle).Error
}

// ==================== Status Transitions ====================

// Start transitions a cycle from upcoming to active.
func (s *CycleService) Start(cycleID, userID uint64) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	st := computeStatus(&cycle)
	if st == "active" {
		return nil, common.BadRequest("Cycle is already active")
	}
	if st == "completed" {
		return nil, common.BadRequest("Cannot start a completed cycle")
	}
	if st == "cancelled" {
		return nil, common.BadRequest("Cannot start a cancelled cycle")
	}

	// If start date is in the future, set it to today
	today := time.Now().Truncate(24 * time.Hour)
	if cycle.StartDate.After(today) {
		cycle.StartDate = today
	}
	cycle.CompletedAt = nil
	cycle.CancelledAt = nil
	cycle.UpdatedByID = &userID

	if err := s.db.Save(&cycle).Error; err != nil {
		return nil, common.Internal("Failed to start cycle")
	}

	return s.buildResponse(&cycle), nil
}

// End transitions a cycle from active to completed.
func (s *CycleService) End(cycleID, userID uint64) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	st := computeStatus(&cycle)
	if st == "completed" {
		return nil, common.BadRequest("Cycle is already completed")
	}
	if st == "cancelled" {
		return nil, common.BadRequest("Cannot end a cancelled cycle")
	}

	now := time.Now()
	cycle.CompletedAt = &now
	cycle.CancelledAt = nil
	cycle.UpdatedByID = &userID

	if err := s.db.Save(&cycle).Error; err != nil {
		return nil, common.Internal("Failed to end cycle")
	}

	return s.buildResponse(&cycle), nil
}

// Cancel cancels a cycle.
func (s *CycleService) Cancel(cycleID, userID uint64) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	if computeStatus(&cycle) == "completed" {
		return nil, common.BadRequest("Cannot cancel a completed cycle")
	}

	now := time.Now()
	cycle.CancelledAt = &now
	cycle.CompletedAt = nil
	cycle.UpdatedByID = &userID

	if err := s.db.Save(&cycle).Error; err != nil {
		return nil, common.Internal("Failed to cancel cycle")
	}

	return s.buildResponse(&cycle), nil
}

// ==================== Issue Association ====================

// AddIssue adds an issue to a cycle.
func (s *CycleService) AddIssue(cycleID, issueID uint64) error {
	// Validate cycle exists
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return common.NotFound("Cycle not found")
	}

	if computeStatus(&cycle) == "completed" || computeStatus(&cycle) == "cancelled" {
		return common.BadRequest("Cannot add issues to a completed or cancelled cycle")
	}

	// Validate issue exists
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	// Validate same project
	if issue.ProjectID != cycle.ProjectID {
		return common.BadRequest("Issue does not belong to this cycle's project")
	}

	// Check if already in this cycle
	var count int64
	s.db.Model(&model.IssueCycle{}).Where("issue_id = ? AND cycle_id = ?", issueID, cycleID).Count(&count)
	if count > 0 {
		return common.BadRequest("Issue is already in this cycle")
	}

	if err := s.db.Create(&model.IssueCycle{IssueID: issueID, CycleID: cycleID}).Error; err != nil {
		return common.Internal("Failed to add issue to cycle")
	}

	return nil
}

// RemoveIssue removes an issue from a cycle.
func (s *CycleService) RemoveIssue(cycleID, issueID uint64) error {
	result := s.db.Where("issue_id = ? AND cycle_id = ?", issueID, cycleID).Delete(&model.IssueCycle{})
	if result.RowsAffected == 0 {
		return common.NotFound("Issue is not in this cycle")
	}
	return nil
}

// ListIssues returns issues in a cycle with optional filters.
func (s *CycleService) ListIssues(cycleID uint64, stateID *uint64, priority string, limit, offset int) ([]response.IssueResponse, int64, error) {
	// Validate cycle exists
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, 0, common.NotFound("Cycle not found")
	}

	baseQuery := s.db.Model(&model.Issue{}).
		Joins("JOIN issue_cycles ON issue_cycles.issue_id = issues.id").
		Where("issue_cycles.cycle_id = ?", cycleID)

	if stateID != nil {
		baseQuery = baseQuery.Where("issues.state_id = ?", *stateID)
	}
	if priority != "" {
		baseQuery = baseQuery.Where("issues.priority = ?", priority)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var issues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("Project").
		Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").
		Preload("CycleLink").
		Order("issues.sort_order ASC, issues.sequence_id DESC").
		Limit(limit).Offset(offset).
		Find(&issues).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	issueSvc := &IssueService{db: s.db}
	result := make([]response.IssueResponse, len(issues))
	for i, issue := range issues {
		resp, err := issueSvc.buildIssueResponse(&issue)
		if err != nil {
			return nil, 0, err
		}
		result[i] = *resp
	}

	return result, total, nil
}

// ==================== Analysis ====================

// GetProgress returns cycle progress statistics.
func (s *CycleService) GetProgress(cycleID uint64) (*response.CycleProgress, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	total, completed, _ := s.countIssues(cycleID)
	progress := 0.0
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}

	// State breakdown
	type stateRow struct {
		Name  string
		Group string
		Count int64
	}
	var rows []stateRow
	s.db.Table("issue_cycles").
		Select("states.name, states.group, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issue_cycles.cycle_id = ?", cycleID).
		Group("states.name, states.group").
		Scan(&rows)

	breakdown := make([]response.StateBreakdown, len(rows))
	for i, r := range rows {
		breakdown[i] = response.StateBreakdown{
			State: r.Name,
			Group: r.Group,
			Count: r.Count,
		}
	}

	return &response.CycleProgress{
		CycleID:         cycleID,
		CycleName:       cycle.Name,
		TotalIssues:     total,
		CompletedIssues: completed,
		Progress:        float64(int(progress*100)) / 100,
		StateBreakdown:  breakdown,
	}, nil
}

// GetStatistics returns detailed cycle statistics.
func (s *CycleService) GetStatistics(cycleID uint64) (*response.CycleStatistics, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	progress, err := s.GetProgress(cycleID)
	if err != nil {
		return nil, err
	}

	// Priority breakdown
	type priorityRow struct {
		Priority string
		Count    int64
	}
	var pRows []priorityRow
	s.db.Table("issue_cycles").
		Select("issues.priority, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ?", cycleID).
		Group("issues.priority").
		Scan(&pRows)

	priorityBreakdown := make(map[string]int64)
	for _, r := range pRows {
		priorityBreakdown[r.Priority] = r.Count
	}

	// Issue date stats
	var withStart, withTarget int64
	s.db.Table("issue_cycles").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ? AND issues.start_date IS NOT NULL", cycleID).Count(&withStart)
	s.db.Table("issue_cycles").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ? AND issues.target_date IS NOT NULL", cycleID).Count(&withTarget)

	return &response.CycleStatistics{
		CycleProgress:     *progress,
		PriorityBreakdown: priorityBreakdown,
		IssueStats: response.IssueStats{
			Total:          progress.TotalIssues,
			WithStartDate:  withStart,
			WithTargetDate: withTarget,
		},
		DateRange: response.DateRange{
			StartDate: dateStr(&cycle.StartDate),
			EndDate:   dateStr(cycle.EndDate),
		},
	}, nil
}

// GetBurndown returns burndown chart data.
func (s *CycleService) GetBurndown(cycleID uint64) (*response.BurndownData, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	if cycle.EndDate == nil {
		return nil, common.BadRequest("Cycle does not have end date")
	}

	total, completed, _ := s.countIssues(cycleID)

	totalDays := int(cycle.EndDate.Sub(cycle.StartDate).Hours() / 24)
	if totalDays <= 0 {
		totalDays = 1
	}

	today := time.Now().Truncate(24 * time.Hour)
	daysElapsed := 0
	if today.After(cycle.StartDate) {
		daysElapsed = int(today.Sub(cycle.StartDate).Hours() / 24)
	}
	if daysElapsed > totalDays {
		daysElapsed = totalDays
	}

	idealDailyBurn := float64(total) / float64(totalDays)
	idealRemaining := float64(total) - idealDailyBurn*float64(daysElapsed)
	if idealRemaining < 0 {
		idealRemaining = 0
	}

	actualRemaining := total - completed

	return &response.BurndownData{
		CycleID:         cycleID,
		CycleName:       cycle.Name,
		StartDate:       cycle.StartDate.Format("2006-01-02"),
		EndDate:         cycle.EndDate.Format("2006-01-02"),
		TotalIssues:     total,
		TotalDays:       totalDays,
		DaysElapsed:     daysElapsed,
		IdealDailyBurn:  float64(int(idealDailyBurn*100)) / 100,
		IdealRemaining:  float64(int(idealRemaining*100)) / 100,
		ActualCompleted: completed,
		ActualRemaining: actualRemaining,
		IsOnTrack:       float64(actualRemaining) <= idealRemaining,
	}, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly. Check that `IssueService.buildIssueResponse` is exported (uppercase `B` in `BuildIssueResponse` if needed).

Note: If `buildIssueResponse` is lowercase (unexported), the CycleService cannot call it. In that case, change it to `BuildIssueResponse` in `issue_service.go` or duplicate the issue-building logic in CycleService.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/cycle_service.go
git commit -m "feat(cycle): implement full CycleService with CRUD, status, issues, analysis"
```

---

### Task 4: Go CycleHandler �?HTTP Handlers

**Files:**
- Modify: `backend/internal/handler/cycle_handler.go`

- [ ] **Step 1: Write the complete handler**

Replace the entire content of `backend/internal/handler/cycle_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/dto/response"
	"github.com/reqmanpy/backend/internal/middleware"
	"github.com/reqmanpy/backend/internal/service"
)

type CycleHandler struct {
	svc *service.CycleService
}

func NewCycleHandler(svc *service.CycleService) *CycleHandler {
	return &CycleHandler{svc: svc}
}

func (h *CycleHandler) parseCycleID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("cycleId"), 10, 64)
}

func (h *CycleHandler) parseProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

// appError sends an AppError response if err is one, otherwise returns false.
func appError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	return true
}

// ==================== CRUD ====================

// Create handles POST /projects/:projectId/cycles?workspace_id=uint
func (h *CycleHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	var req request.CycleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Override project_id from URL param
	projectID, err := h.parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	req.ProjectID = projectID

	resp, svcErr := h.svc.Create(workspaceID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List handles GET /projects/:projectId/cycles?status=str&limit=int&offset=int
func (h *CycleHandler) List(c *gin.Context) {
	projectID, err := h.parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)
	status := c.Query("status")

	cycles, total, svcErr := h.svc.ListByProject(projectID, status, p.Limit, p.Offset)
	if appError(c, svcErr) {
		return
	}

	if cycles == nil {
		cycles = []response.CycleResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  cycles,
		"total":  total,
		"limit":  p.Limit,
		"offset": p.Offset,
	})
}

// Get handles GET /cycles/:cycleId
func (h *CycleHandler) Get(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetByID(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /cycles/:cycleId
func (h *CycleHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	var req request.CycleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(cycleID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /cycles/:cycleId
func (h *CycleHandler) Delete(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	if appError(c, h.svc.Delete(cycleID)) {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ==================== Status Transitions ====================

// Start handles POST /cycles/:cycleId/start
func (h *CycleHandler) Start(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.Start(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// End handles POST /cycles/:cycleId/end
func (h *CycleHandler) End(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.End(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Cancel handles POST /cycles/:cycleId/cancel
func (h *CycleHandler) Cancel(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.Cancel(cycleID, user.ID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== Issue Association ====================

// AddIssue handles POST /cycles/:cycleId/issues?issue_id=uint
func (h *CycleHandler) AddIssue(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	issueID, err := strconv.ParseUint(c.Query("issue_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue_id"})
		return
	}

	if appError(c, h.svc.AddIssue(cycleID, issueID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cycle_id": cycleID,
		"issue_id": issueID,
		"action":   "added",
	})
}

// RemoveIssue handles DELETE /cycles/:cycleId/issues/:issueId
func (h *CycleHandler) RemoveIssue(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	if appError(c, h.svc.RemoveIssue(cycleID, issueID)) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cycle_id": cycleID,
		"issue_id": issueID,
		"action":   "removed",
	})
}

// ListIssues handles GET /cycles/:cycleId/issues?state_id=&priority=&limit=&offset=
func (h *CycleHandler) ListIssues(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)

	var stateID *uint64
	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			stateID = &id
		}
	}

	priority := c.Query("priority")

	issues, _, svcErr := h.svc.ListIssues(cycleID, stateID, priority, p.Limit, p.Offset)
	if appError(c, svcErr) {
		return
	}

	if issues == nil {
		issues = []response.IssueResponse{}
	}

	c.JSON(http.StatusOK, issues)
}

// ==================== Analysis ====================

// GetProgress handles GET /cycles/:cycleId/progress
func (h *CycleHandler) GetProgress(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetProgress(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStatistics handles GET /cycles/:cycleId/statistics
func (h *CycleHandler) GetStatistics(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetStatistics(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetBurndown handles GET /cycles/:cycleId/burndown
func (h *CycleHandler) GetBurndown(c *gin.Context) {
	cycleID, err := h.parseCycleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cycle ID"})
		return
	}

	resp, svcErr := h.svc.GetBurndown(cycleID)
	if appError(c, svcErr) {
		return
	}

	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/cycle_handler.go
git commit -m "feat(cycle): implement full CycleHandler with all endpoints"
```

---

### Task 5: Go Router �?Register Cycle Routes

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Step 1: Add cycle routes to router**

In `backend/internal/router/router.go`, add the cycle route group AFTER the existing `projects.GET("/:projectId/cycles", cycleH.List)` line and BEFORE the `// ---- Issues` block.

Add this block after the existing cycleH.List route and the settings group:

```go
			// ---- Cycles (protected) ----
			cycles := v1.Group("/cycles", authMiddleware)
			{
				cycles.GET("/:cycleId", cycleH.Get)
				cycles.PUT("/:cycleId", cycleH.Update)
				cycles.DELETE("/:cycleId", cycleH.Delete)
				cycles.POST("/:cycleId/start", cycleH.Start)
				cycles.POST("/:cycleId/end", cycleH.End)
				cycles.POST("/:cycleId/cancel", cycleH.Cancel)
				cycles.POST("/:cycleId/issues", cycleH.AddIssue)
				cycles.DELETE("/:cycleId/issues/:issueId", cycleH.RemoveIssue)
				cycles.GET("/:cycleId/issues", cycleH.ListIssues)
				cycles.GET("/:cycleId/progress", cycleH.GetProgress)
				cycles.GET("/:cycleId/statistics", cycleH.GetStatistics)
				cycles.GET("/:cycleId/burndown", cycleH.GetBurndown)
			}
```

Also add the route for creating cycles �?modify the existing project routes to include:

```go
				projects.POST("/:projectId/cycles", cycleH.Create)     // ?workspace_id=
```

- [ ] **Step 2: Verify compilation**

Run: `cd backend && go build ./internal/router/`
Expected: Compiles cleanly

- [ ] **Step 3: Verify full build**

Run: `cd backend && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/router/router.go
git commit -m "feat(cycle): register all cycle API routes"
```

---

### Task 6: Go �?Fix buildIssueResponse Visibility

**Files:**
- Modify: `backend/internal/service/issue_service.go`

- [ ] **Step 1: Export buildIssueResponse**

If the build fails because `buildIssueResponse` is unexported, renamet it to `BuildIssueResponse` in `issue_service.go`:

In `issue_service.go`, change:
- `func (s *IssueService) buildIssueResponse` �?`func (s *IssueService) BuildIssueResponse`
- All internal call sites: `s.buildIssueResponse(` �?`s.BuildIssueResponse(`

Update `cycle_service.go` accordingly:
- `issueSvc.buildIssueResponse(&issue)` �?`issueSvc.BuildIssueResponse(&issue)`

- [ ] **Step 2: Verify full build**

Run: `cd backend && go build ./...`
Expected: Compiles cleanly

- [ ] **Step 3: Start server and test one endpoint**

Run: `cd backend && ./server.exe`
Test: `curl http://localhost:8080/api/v1/projects/1/cycles -H "Authorization: Bearer <valid_token>"`
Expected: Empty array `{"items":[],"total":0,"limit":50,"offset":0}`

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/issue_service.go backend/internal/service/cycle_service.go
git commit -m "fix(cycle): export BuildIssueResponse for CycleService use"
```

---

### Task 7: Frontend �?Update Types & API Client

**Files:**
- Modify: `frontend/src/types/cycle.ts`
- Modify: `frontend/src/api/cycle.ts`

- [ ] **Step 1: Update TypeScript types to align with Go responses**

Modify `frontend/src/types/cycle.ts` �?adjust `CycleResponse`:

```typescript
// frontend/src/types/cycle.ts �?adjust the CycleResponse interface
export interface CycleResponse extends CycleBase {
  id: number
  status: CycleStatus           // 'upcoming' | 'active' | 'completed' | 'cancelled'
  progress: number              // 0-100, float
  total_issues: number
  completed_issues: number
  project_id: number
  workspace_id: number
  owned_by?: UserLite
  project?: ProjectLite
  created_at: string
  updated_at: string
  created_by_id?: number
  updated_by_id?: number
}

export interface ProjectLite {
  id: number
  name: string
  identifier: string
}
```

- [ ] **Step 2: Update API client to match Go routes**

Modify `frontend/src/api/cycle.ts`:

```typescript
// frontend/src/api/cycle.ts �?adjusted URL paths
import api from './index'
import type { CycleCreate, CycleUpdate, CycleResponse, CycleProgress, CycleStatistics, BurndownData } from '@/types/cycle'

export async function createCycle(projectId: number, workspaceId: number, data: CycleCreate): Promise<CycleResponse> {
  const response = await api.post(`/projects/${projectId}/cycles?workspace_id=${workspaceId}`, data)
  return response.data
}

export async function listCycles(projectId: number, options?: { status?: string; limit?: number; offset?: number }): Promise<{ items: CycleResponse[]; total: number; limit: number; offset: number }> {
  const params = new URLSearchParams()
  if (options?.status) params.append('status', options.status)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  const response = await api.get(`/projects/${projectId}/cycles?${params.toString()}`)
  return response.data
}

export async function getCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.get(`/cycles/${cycleId}`)
  return response.data
}

export async function updateCycle(cycleId: number, data: CycleUpdate): Promise<CycleResponse> {
  const response = await api.put(`/cycles/${cycleId}`, data)
  return response.data
}

export async function deleteCycle(cycleId: number): Promise<void> {
  await api.delete(`/cycles/${cycleId}`)
}

export async function startCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/start`)
  return response.data
}

export async function endCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/end`)
  return response.data
}

export async function cancelCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/cancel`)
  return response.data
}

export async function addIssueToCycle(cycleId: number, issueId: number): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.post(`/cycles/${cycleId}/issues?issue_id=${issueId}`)
  return response.data
}

export async function removeIssueFromCycle(cycleId: number, issueId: number): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.delete(`/cycles/${cycleId}/issues/${issueId}`)
  return response.data
}

export async function getCycleIssues(cycleId: number, options?: { state_id?: number; priority?: string; limit?: number; offset?: number }): Promise<any[]> {
  const params = new URLSearchParams()
  if (options?.state_id) params.append('state_id', options.state_id.toString())
  if (options?.priority) params.append('priority', options.priority)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  const response = await api.get(`/cycles/${cycleId}/issues?${params.toString()}`)
  return response.data
}

export async function getCycleProgress(cycleId: number): Promise<CycleProgress> {
  const response = await api.get(`/cycles/${cycleId}/progress`)
  return response.data
}

export async function getCycleStatistics(cycleId: number): Promise<CycleStatistics> {
  const response = await api.get(`/cycles/${cycleId}/statistics`)
  return response.data
}

export async function getBurndownData(cycleId: number): Promise<BurndownData> {
  const response = await api.get(`/cycles/${cycleId}/burndown`)
  return response.data
}

export const cycleApi = {
  createCycle, listCycles, getCycle, updateCycle, deleteCycle,
  startCycle, endCycle, cancelCycle,
  addIssueToCycle, removeIssueFromCycle, getCycleIssues,
  getCycleProgress, getCycleStatistics, getBurndownData
}
export default cycleApi
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/cycle.ts frontend/src/api/cycle.ts
git commit -m "feat(cycle): update TypeScript types and API client for Go backend"
```

---

### Task 8: Frontend �?Pinia Store

**Files:**
- Create: `frontend/src/stores/cycle.ts`

- [ ] **Step 1: Create Cycle Pinia store**

```typescript
// frontend/src/stores/cycle.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { cycleApi } from '@/api/cycle'
import type { CycleResponse, CycleCreate, CycleUpdate, CycleProgress, CycleStatistics, BurndownData } from '@/types/cycle'

export const useCycleStore = defineStore('cycle', () => {
  const cycles = ref<CycleResponse[]>([])
  const currentCycle = ref<CycleResponse | null>(null)
  const progress = ref<CycleProgress | null>(null)
  const statistics = ref<CycleStatistics | null>(null)
  const burndown = ref<BurndownData | null>(null)
  const cycleIssues = ref<any[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // ---- List ----
  async function fetchCycles(projectId: number, status?: string) {
    isLoading.value = true
    error.value = null
    try {
      const result = await cycleApi.listCycles(projectId, { status })
      cycles.value = result.items
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ---- Get one ----
  async function fetchCycle(cycleId: number) {
    isLoading.value = true
    error.value = null
    try {
      currentCycle.value = await cycleApi.getCycle(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ---- CRUD ----
  async function createCycle(projectId: number, workspaceId: number, data: CycleCreate) {
    isLoading.value = true
    error.value = null
    try {
      const created = await cycleApi.createCycle(projectId, workspaceId, data)
      cycles.value.unshift(created)
      return created
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function updateCycleAction(cycleId: number, data: CycleUpdate) {
    error.value = null
    try {
      const updated = await cycleApi.updateCycle(cycleId, data)
      const idx = cycles.value.findIndex(c => c.id === cycleId)
      if (idx !== -1) cycles.value[idx] = updated
      if (currentCycle.value?.id === cycleId) currentCycle.value = updated
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function deleteCycleAction(cycleId: number) {
    error.value = null
    try {
      await cycleApi.deleteCycle(cycleId)
      cycles.value = cycles.value.filter(c => c.id !== cycleId)
      if (currentCycle.value?.id === cycleId) currentCycle.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ---- Status ----
  async function startCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.startCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function endCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.endCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function cancelCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.cancelCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  // ---- Issues ----
  async function addIssueToCycle(cycleId: number, issueId: number) {
    error.value = null
    try {
      const result = await cycleApi.addIssueToCycle(cycleId, issueId)
      // Refetch to update counts
      await fetchCycleIssues(cycleId)
      await fetchProgress(cycleId)
      return result
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function removeIssueFromCycle(cycleId: number, issueId: number) {
    error.value = null
    try {
      const result = await cycleApi.removeIssueFromCycle(cycleId, issueId)
      cycleIssues.value = cycleIssues.value.filter((i: any) => i.id !== issueId)
      await fetchProgress(cycleId)
      return result
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function fetchCycleIssues(cycleId: number, filters?: { state_id?: number; priority?: string }) {
    isLoading.value = true
    error.value = null
    try {
      cycleIssues.value = await cycleApi.getCycleIssues(cycleId, filters)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ---- Analysis ----
  async function fetchProgress(cycleId: number) {
    try {
      progress.value = await cycleApi.getCycleProgress(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchStatistics(cycleId: number) {
    try {
      statistics.value = await cycleApi.getCycleStatistics(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchBurndown(cycleId: number) {
    try {
      burndown.value = await cycleApi.getBurndownData(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ---- Helper ----
  function updateCycleInList(updated: CycleResponse) {
    const idx = cycles.value.findIndex(c => c.id === updated.id)
    if (idx !== -1) cycles.value[idx] = updated
    if (currentCycle.value?.id === updated.id) currentCycle.value = updated
  }

  return {
    cycles, currentCycle, progress, statistics, burndown, cycleIssues, isLoading, error,
    fetchCycles, fetchCycle,
    createCycle, updateCycleAction, deleteCycleAction,
    startCycle, endCycle, cancelCycle,
    addIssueToCycle, removeIssueFromCycle, fetchCycleIssues,
    fetchProgress, fetchStatistics, fetchBurndown,
  }
})
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/stores/cycle.ts
git commit -m "feat(cycle): add Pinia store for cycle management"
```

---

### Task 9: Frontend �?Update Existing Components (CycleCard, CycleList)

**Files:**
- Modify: `frontend/src/components/CycleCard.vue`
- Modify: `frontend/src/components/CycleList.vue`

- [ ] **Step 1: Update CycleCard to use store and new data shape**

In `CycleCard.vue`, the component already handles status display and progress. The main changes are:
- Props should accept `CycleResponse` from the updated types
- Emit events remain the same

No major structural changes needed �?the existing Card is well-designed. Just review that `props.cycle.status` is read correctly (it's now a string from the API, not computed).

- [ ] **Step 2: Update CycleList to use store**

Replace the data-fetching logic in `CycleList.vue`:

```typescript
// In CycleList.vue <script setup>, replace manual API calls with store:
import { useCycleStore } from '@/stores/cycle'

const cycleStore = useCycleStore()

const cycles = computed(() => cycleStore.cycles)
const loading = computed(() => cycleStore.isLoading)

onMounted(async () => {
  await cycleStore.fetchCycles(props.projectId)
})

watch(() => filters.value.status, async (newStatus) => {
  await cycleStore.fetchCycles(props.projectId, newStatus || undefined)
})
```

Remove the direct `cycleApi.listCycles()` call and the local `cycles` / `loading` refs.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/CycleCard.vue frontend/src/components/CycleList.vue
git commit -m "feat(cycle): update CycleCard and CycleList to use Pinia store"
```

---

### Task 10: Frontend �?New Components (CycleDetailPanel, CycleProgressCard, CycleBurndownChart)

**Files:**
- Create: `frontend/src/components/CycleDetailPanel.vue`
- Create: `frontend/src/components/CycleProgressCard.vue`
- Create: `frontend/src/components/CycleBurndownChart.vue`

- [ ] **Step 1: Create CycleProgressCard.vue**

```vue
<!-- frontend/src/components/CycleProgressCard.vue -->
<template>
  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <!-- Total Issues -->
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-gray-900">{{ progress?.total_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">总工作项</div>
    </div>

    <!-- Completed -->
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-green-600">{{ progress?.completed_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">已完�?/div>
    </div>

    <!-- Progress % -->
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-indigo-600">{{ formatProgress(progress?.progress ?? 0) }}</div>
      <div class="text-xs text-gray-500 mt-1">进度</div>
    </div>

    <!-- Remaining -->
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-orange-600">{{ (progress?.total_issues ?? 0) - (progress?.completed_issues ?? 0) }}</div>
      <div class="text-xs text-gray-500 mt-1">待完�?/div>
    </div>
  </div>

  <!-- State Breakdown -->
  <div v-if="progress?.state_breakdown?.length" class="mt-4 bg-white rounded-lg border border-gray-200 p-4">
    <h4 class="text-sm font-medium text-gray-700 mb-3">状态分�?/h4>
    <div class="space-y-2">
      <div v-for="sb in progress.state_breakdown" :key="sb.state" class="flex items-center justify-between">
        <span class="text-sm text-gray-600">{{ sb.state }}</span>
        <span class="text-sm font-medium text-gray-900">{{ sb.count }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CycleProgress } from '@/types/cycle'

defineProps<{
  progress: CycleProgress | null
}>()

function formatProgress(progress: number): string {
  return `${Math.round(progress)}%`
}
</script>
```

- [ ] **Step 2: Create CycleBurndownChart.vue**

```vue
<!-- frontend/src/components/CycleBurndownChart.vue -->
<template>
  <div class="bg-white rounded-lg border border-gray-200 p-4">
    <h4 class="text-sm font-medium text-gray-700 mb-3">燃尽�?/h4>

    <div v-if="!data" class="text-center py-8 text-gray-400 text-sm">
      暂无燃尽图数据（需要设置起止日期）
    </div>

    <div v-else class="relative">
      <!-- SVG Chart -->
      <svg viewBox="0 0 400 200" class="w-full h-48">
        <!-- Grid lines -->
        <line v-for="i in 5" :key="'h'+i" x1="0" :y1="i*40" x2="400" :y2="i*40" stroke="#f3f4f6" stroke-width="1" />

        <!-- Ideal line -->
        <polyline
          :points="idealLinePoints"
          fill="none"
          stroke="#9CA3AF"
          stroke-width="2"
          stroke-dasharray="6,3"
        />

        <!-- Actual line -->
        <polyline
          :points="actualLinePoints"
          fill="none"
          stroke="#3B82F6"
          stroke-width="2"
        />
      </svg>

      <!-- Legend -->
      <div class="flex items-center justify-center space-x-6 mt-3 text-xs text-gray-500">
        <div class="flex items-center">
          <div class="w-4 h-0.5 bg-gray-400 mr-1" style="border-top: 2px dashed #9CA3AF"></div>
          理想�?        </div>
        <div class="flex items-center">
          <div class="w-4 h-0.5 bg-blue-500 mr-1"></div>
          实际完成
        </div>
        <div :class="data.is_on_track ? 'text-green-600' : 'text-red-600'">
          {{ data.is_on_track ? '�?进度正常' : '�?进度落后' }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BurndownData } from '@/types/cycle'

const props = defineProps<{
  data: BurndownData | null
}>()

const idealLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)
  let points = ''
  for (let d = 0; d <= total_days; d++) {
    const x = d * xStep
    const y = Math.max(0, (total_issues - (total_issues / total_days) * d)) * yScale
    points += `${x},${y} `
  }
  return points.trim()
})

const actualLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days, days_elapsed, actual_completed } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)
  const startX = 0
  const startY = total_issues * yScale
  const endX = Math.min(days_elapsed, total_days) * xStep
  const endY = Math.max(0, (total_issues - actual_completed)) * yScale
  return `${startX},${startY} ${endX},${endY}`
})
</script>
```

- [ ] **Step 3: Create CycleDetailPanel.vue**

```vue
<!-- frontend/src/components/CycleDetailPanel.vue -->
<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      <!-- Header -->
      <div class="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900">{{ cycle?.name }}</h3>
        <button @click="$emit('close')" class="p-1 text-gray-400 hover:text-gray-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div v-if="loading" class="flex justify-center py-12">
        <svg class="animate-spin h-6 w-6 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <div v-else-if="cycle" class="p-4 space-y-4">
        <!-- Status & Progress -->
        <CycleProgressCard :progress="cycleStore.progress" />

        <!-- Burndown -->
        <CycleBurndownChart :data="cycleStore.burndown" />

        <!-- Issues in cycle -->
        <div>
          <h4 class="text-sm font-medium text-gray-700 mb-2">周期工作�?({{ cycleStore.cycleIssues.length }})</h4>
          <div v-if="cycleStore.cycleIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">
            暂无工作�?          </div>
          <div v-else class="space-y-2">
            <div
              v-for="issue in cycleStore.cycleIssues"
              :key="issue.id"
              class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm"
            >
              <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
              <button
                @click="handleRemoveIssue(issue.id)"
                class="ml-2 text-gray-400 hover:text-red-500"
                title="移除"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { watch, computed } from 'vue'
import { useCycleStore } from '@/stores/cycle'
import CycleProgressCard from './CycleProgressCard.vue'
import CycleBurndownChart from './CycleBurndownChart.vue'
import type { CycleResponse } from '@/types/cycle'

const props = defineProps<{
  cycle: CycleResponse | null
  visible: boolean
}>()

defineEmits<{
  (e: 'close'): void
}>()

const cycleStore = useCycleStore()

const loading = computed(() => cycleStore.isLoading)

watch(() => props.visible, async (v) => {
  if (v && props.cycle) {
    await Promise.all([
      cycleStore.fetchProgress(props.cycle.id),
      cycleStore.fetchBurndown(props.cycle.id),
      cycleStore.fetchCycleIssues(props.cycle.id),
    ])
  }
})

async function handleRemoveIssue(issueId: number) {
  if (props.cycle) {
    await cycleStore.removeIssueFromCycle(props.cycle.id, issueId)
  }
}
</script>

<style scoped>
.slide-enter-active, .slide-leave-active {
  transition: transform 0.3s ease;
}
.slide-enter-from, .slide-leave-to {
  transform: translateX(100%);
}
</style>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/CycleProgressCard.vue frontend/src/components/CycleBurndownChart.vue frontend/src/components/CycleDetailPanel.vue
git commit -m "feat(cycle): add CycleDetailPanel, ProgressCard, BurndownChart components"
```

---

### Task 11: Frontend �?Views (CycleCreate, CycleDetail) + Router Integration

**Files:**
- Create: `frontend/src/views/CycleCreate.vue`
- Create: `frontend/src/views/CycleDetail.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/views/Project.vue`

- [ ] **Step 1: Create CycleCreate.vue �?two-step wizard**

```vue
<!-- frontend/src/views/CycleCreate.vue -->
<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center space-x-4">
        <button @click="goBack" class="text-gray-500 hover:text-gray-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 class="text-lg font-semibold text-gray-900">创建新周�?/h1>
      </div>
    </div>

    <!-- Steps indicator -->
    <div class="max-w-3xl mx-auto px-6 py-6">
      <div class="flex items-center justify-center mb-8">
        <div v-for="(step, i) in steps" :key="i" class="flex items-center">
          <div
            class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium"
            :class="currentStep >= i ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-500'"
          >{{ i + 1 }}</div>
          <span class="ml-2 text-sm" :class="currentStep >= i ? 'text-indigo-600 font-medium' : 'text-gray-400'">{{ step }}</span>
          <div v-if="i < steps.length - 1" class="w-16 h-0.5 mx-3" :class="currentStep > i ? 'bg-indigo-600' : 'bg-gray-200'"></div>
        </div>
      </div>

      <!-- Step 1: Basic Info -->
      <div v-if="currentStep === 0" class="bg-white rounded-lg shadow-sm p-6 space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700">名称 <span class="text-red-500">*</span></label>
          <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="�? Sprint 1" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">描述</label>
          <textarea v-model="form.description" rows="3" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="描述此周期的目标..."></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">开始日�?<span class="text-red-500">*</span></label>
            <input v-model="form.start_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">结束日期</label>
            <input v-model="form.end_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
        </div>
      </div>

      <!-- Step 2: Select Issues -->
      <div v-if="currentStep === 1" class="bg-white rounded-lg shadow-sm p-6">
        <p class="text-sm text-gray-500 mb-4">�?Backlog 中选择要加入此周期的工作项（可跳过，稍后添加）</p>
        <div class="max-h-96 overflow-y-auto space-y-2">
          <label v-for="issue in backlogIssues" :key="issue.id" class="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer">
            <input type="checkbox" :value="issue.id" v-model="selectedIssueIds" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
            <span class="ml-3 text-sm text-gray-900">{{ issue.name }}</span>
            <span class="ml-auto text-xs text-gray-400">#{{ issue.sequence_id }}</span>
          </label>
          <p v-if="backlogIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无 Backlog 工作�?/p>
        </div>
      </div>

      <!-- Step 3: Confirm -->
      <div v-if="currentStep === 2" class="bg-white rounded-lg shadow-sm p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-4">确认创建</h3>
        <div class="space-y-2 text-sm">
          <div class="flex"><span class="text-gray-500 w-20">名称:</span><span class="text-gray-900">{{ form.name }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">描述:</span><span class="text-gray-900">{{ form.description || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">开�?</span><span class="text-gray-900">{{ form.start_date }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">结束:</span><span class="text-gray-900">{{ form.end_date || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">工作�?</span><span class="text-gray-900">{{ selectedIssueIds.length }} �?/span></div>
        </div>
      </div>

      <!-- Navigation -->
      <div class="flex justify-between mt-6">
        <button v-if="currentStep > 0" @click="currentStep--" class="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">上一�?/button>
        <div v-else></div>

        <button v-if="currentStep < 2" @click="currentStep++" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700">下一�?/button>

        <button
          v-if="currentStep === 2"
          @click="submitCycle"
          :disabled="submitting"
          class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700 disabled:opacity-50"
        >
          {{ submitting ? '创建�?..' : '创建周期' }}
        </button>
      </div>

      <!-- Error -->
      <div v-if="cycleStore.error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-600">
        {{ cycleStore.error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCycleStore } from '@/stores/cycle'
import { issueApi } from '@/api/issue'

const route = useRoute()
const router = useRouter()
const cycleStore = useCycleStore()

const workspaceId = Number(route.params.workspaceId)
const projectId = Number(route.params.projectId)

const steps = ['基本信息', '选择工作�?, '确认']

const currentStep = ref(0)
const submitting = ref(false)

const form = ref({
  name: '',
  description: '',
  start_date: new Date().toISOString().slice(0, 10),
  end_date: '',
})

const selectedIssueIds = ref<number[]>([])
const backlogIssues = ref<any[]>([])

onMounted(async () => {
  try {
    const result = await issueApi.listIssues(projectId, { limit: 100 })
    // Filter for backlog/uncompleted issues without a cycle
    backlogIssues.value = (Array.isArray(result) ? result : result.items || [])
      .filter((i: any) => !i.cycle_id && i.state_group !== 'completed' && i.state_group !== 'cancelled')
  } catch {}
})

function goBack() {
  router.push(`/workspace/${route.params.slug || workspaceId}/project/${projectId}`)
}

async function submitCycle() {
  submitting.value = true
  const created = await cycleStore.createCycle(projectId, workspaceId, {
    name: form.value.name,
    description: form.value.description || undefined,
    start_date: new Date(form.value.start_date).toISOString(),
    end_date: form.value.end_date ? new Date(form.value.end_date).toISOString() : undefined,
    project_id: projectId,
  })

  if (!created) {
    submitting.value = false
    return
  }

  // Add selected issues to the cycle
  for (const issueId of selectedIssueIds.value) {
    await cycleStore.addIssueToCycle(created.id, issueId)
  }

  submitting.value = false
  router.push(`/workspace/${route.params.slug || workspaceId}/project/${projectId}`)
}
</script>
```

- [ ] **Step 2: Create CycleDetail.vue �?full detail page**

```vue
<!-- frontend/src/views/CycleDetail.vue -->
<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 class="text-lg font-semibold text-gray-900">{{ cycle?.name }}</h1>
          <span :class="statusClass">{{ cycle?.status }}</span>
        </div>
        <div class="flex items-center space-x-2">
          <button v-if="cycle?.status === 'upcoming'" @click="handleStart" class="px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700">开�?/button>
          <button v-if="cycle?.status === 'active'" @click="handleEnd" class="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">结束</button>
          <button v-if="cycle?.status !== 'completed' && cycle?.status !== 'cancelled'" @click="handleCancel" class="px-3 py-1.5 border border-gray-300 text-sm text-gray-600 rounded hover:bg-gray-50">取消</button>
          <button @click="handleDelete" class="px-3 py-1.5 border border-red-300 text-sm text-red-600 rounded hover:bg-red-50">删除</button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div v-if="loading" class="flex justify-center py-20">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
    </div>

    <div v-else-if="cycle" class="max-w-5xl mx-auto px-6 py-6 space-y-6">
      <!-- Progress Cards -->
      <CycleProgressCard :progress="cycleStore.progress" />

      <!-- Burndown -->
      <CycleBurndownChart :data="cycleStore.burndown" />

      <!-- Issues -->
      <div class="bg-white rounded-lg border border-gray-200 p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-3">周期工作�?/h3>
        <div v-if="cycleStore.cycleIssues.length === 0" class="text-sm text-gray-400 text-center py-8">
          暂无工作�?        </div>
        <div v-else class="space-y-2">
          <div v-for="issue in cycleStore.cycleIssues" :key="issue.id" class="flex items-center p-2 bg-gray-50 rounded text-sm">
            <span class="text-gray-900">{{ issue.name }}</span>
            <span class="ml-auto text-xs text-gray-400">#{{ issue.sequence_id }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCycleStore } from '@/stores/cycle'
import CycleProgressCard from '@/components/CycleProgressCard.vue'
import CycleBurndownChart from '@/components/CycleBurndownChart.vue'

const route = useRoute()
const router = useRouter()
const cycleStore = useCycleStore()

const cycleId = Number(route.params.cycleId)
const cycle = computed(() => cycleStore.currentCycle)
const loading = computed(() => cycleStore.isLoading)

const statusClass = computed(() => {
  const map: Record<string, string> = {
    upcoming: 'px-2 py-0.5 text-xs rounded bg-blue-100 text-blue-700',
    active: 'px-2 py-0.5 text-xs rounded bg-green-100 text-green-700',
    completed: 'px-2 py-0.5 text-xs rounded bg-gray-100 text-gray-600',
    cancelled: 'px-2 py-0.5 text-xs rounded bg-red-100 text-red-700',
  }
  return map[cycle.value?.status ?? ''] || ''
})

onMounted(async () => {
  await Promise.all([
    cycleStore.fetchCycle(cycleId),
  ])
  if (cycle.value) {
    await Promise.all([
      cycleStore.fetchProgress(cycleId),
      cycleStore.fetchBurndown(cycleId),
      cycleStore.fetchCycleIssues(cycleId),
    ])
  }
})

function goBack() {
  router.back()
}

async function handleStart() {
  await cycleStore.startCycle(cycleId)
  await cycleStore.fetchCycle(cycleId)
}

async function handleEnd() {
  if (confirm('确定要结束这个周期吗�?)) {
    await cycleStore.endCycle(cycleId)
    await cycleStore.fetchCycle(cycleId)
  }
}

async function handleCancel() {
  if (confirm('确定要取消这个周期吗�?)) {
    await cycleStore.cancelCycle(cycleId)
    await cycleStore.fetchCycle(cycleId)
  }
}

async function handleDelete() {
  if (confirm('确定要删除这个周期吗？此操作不可撤销�?)) {
    await cycleStore.deleteCycleAction(cycleId)
    router.back()
  }
}
</script>
```

- [ ] **Step 3: Add routes to router**

In `frontend/src/router/index.ts`, add two new routes:

```typescript
    {
      path: '/workspaces/:workspaceId/projects/:projectId/cycles/new',
      name: 'CycleCreate',
      component: () => import('@/views/CycleCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/cycles/:cycleId',
      name: 'CycleDetail',
      component: () => import('@/views/CycleDetail.vue'),
      meta: { requiresAuth: true }
    },
```

- [ ] **Step 4: Integrate CycleDetailPanel into Project.vue**

In `frontend/src/views/Project.vue`, add CycleDetailPanel:

In the template, add after the existing `IssueDetailPanel`:

```vue
      <!-- 侧滑周期详情面板 -->
      <CycleDetailPanel
        :cycle="selectedCycle"
        :visible="cyclePanelVisible"
        @close="cyclePanelVisible = false"
      />
```

In the script, add imports and state:

```typescript
import CycleDetailPanel from '@/components/CycleDetailPanel.vue'
import type { CycleResponse } from '@/types/cycle'

const selectedCycle = ref<CycleResponse | null>(null)
const cyclePanelVisible = ref(false)
```

Update the CycleList usage in the template to handle the `@select` event:

```vue
        <CycleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openCyclePanel"
          @create="goToCycleCreate"
        />
```

Add the handler functions in script:

```typescript
function openCyclePanel(cycle: CycleResponse) {
  selectedCycle.value = cycle
  cyclePanelVisible.value = true
}

function goToCycleCreate() {
  router.push(`/workspaces/${workspaceId.value}/projects/${projectId.value}/cycles/new`)
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/CycleCreate.vue frontend/src/views/CycleDetail.vue frontend/src/router/index.ts frontend/src/views/Project.vue
git commit -m "feat(cycle): add CycleCreate, CycleDetail views and router integration"
```

---

### Task 12: End-to-End Smoke Test

- [ ] **Step 1: Start Go backend**

Run: `cd backend && ./server.exe`
Expected: Server starts on port 8080

- [ ] **Step 2: Start Vue frontend**

Run: `cd frontend && npm run dev`
Expected: Dev server starts, accessible at localhost

- [ ] **Step 3: Test cycle creation flow**

Manual test:
1. Login to the app
2. Navigate to a project
3. Click "周期" tab �?should see empty state or existing cycles
4. Click "新建周期" �?should navigate to creation wizard
5. Fill in name, dates �?proceed �?confirm �?submit
6. Should redirect back to project, cycle list should show new cycle

- [ ] **Step 4: Test status transitions**

Manual test:
1. Click on a cycle card �?side panel opens
2. Verify progress/stats display
3. Click "开�? �?status changes to active
4. Click "结束" �?status changes to completed

- [ ] **Step 5: Test cycle detail page**

Manual test:
1. Navigate to `/workspaces/:wsId/projects/:pId/cycles/:cycleId`
2. Verify burndown chart renders
3. Verify issue list displays

- [ ] **Step 6: Commit any fixes**

```bash
git add -A
git commit -m "fix(cycle): end-to-end smoke test fixes"
```

---

## Plan Summary

| Task | Area | Files | Steps |
|------|------|-------|-------|
| 1 | Go DTO | 2 new | 4 |
| 2 | Go Model | 1 modify | 3 |
| 3 | Go Service | 1 modify | 3 |
| 4 | Go Handler | 1 modify | 3 |
| 5 | Go Router | 1 modify | 4 |
| 6 | Go Fix visibility | 2 modify | 4 |
| 7 | Frontend Types/API | 2 modify | 3 |
| 8 | Frontend Store | 1 new | 2 |
| 9 | Frontend Components (update) | 2 modify | 3 |
| 10 | Frontend Components (new) | 3 new | 4 |
| 11 | Frontend Views + Router | 2 new, 2 modify | 5 |
| 12 | Smoke Test | all | 6 |

**Total: 12 tasks, ~44 steps**
