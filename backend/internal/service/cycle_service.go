package service

import (
	"log"
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
	total, completed, err := s.countIssues(cycle.ID)
	if err != nil {
		log.Printf("cycle: error counting issues for cycle %d: %v", cycle.ID, err)
	}
	progress := 0.0
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}

	var endDateStr *string
	if cycle.EndDate != nil {
		val := cycle.EndDate.Format("2006-01-02")
		endDateStr = &val
	}

	resp := &response.CycleResponse{
		ID:              cycle.ID,
		Name:            cycle.Name,
		Description:     cycle.Description,
		Status:          computeStatus(cycle),
		Progress:        float64(int(progress*100)) / 100,
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

	var project model.Project
	if err := s.db.First(&project, cycle.ProjectID).Error; err == nil {
		resp.Project = &response.ProjectLite{
			ID:         project.ID,
			Name:       project.Name,
			Identifier: project.Identifier,
		}
	}

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

func (s *CycleService) Create(workspaceID, userID uint64, req *request.CycleCreateRequest) (*response.CycleResponse, error) {
	var project model.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, common.BadRequest("Invalid start_date format, use RFC3339")
	}
	startDate = startDate.Truncate(24 * time.Hour)

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return nil, common.BadRequest("Invalid end_date format, use RFC3339")
		}
		t = t.Truncate(24 * time.Hour)
		endDate = &t
	}

	if endDate != nil && startDate.After(*endDate) {
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

func (s *CycleService) GetByID(cycleID uint64) (*response.CycleResponse, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}
	return s.buildResponse(&cycle), nil
}

func (s *CycleService) ListByProject(projectID uint64, status string, limit, offset int) ([]response.CycleResponse, int64, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, 0, common.NotFound("Project not found")
	}

	var allCycles []model.Cycle
	if err := s.db.Where("project_id = ?", projectID).Order("start_date DESC").Find(&allCycles).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var filtered []model.Cycle
	for _, c := range allCycles {
		if status == "" || computeStatus(&c) == status {
			filtered = append(filtered, c)
		}
	}

	total := int64(len(filtered))

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

	if cycle.EndDate != nil && cycle.StartDate.After(*cycle.EndDate) {
		return nil, common.BadRequest("Start date must be before or equal to end date")
	}

	cycle.UpdatedByID = &userID
	if err := s.db.Save(&cycle).Error; err != nil {
		return nil, common.Internal("Failed to update cycle")
	}

	return s.buildResponse(&cycle), nil
}

func (s *CycleService) Delete(cycleID uint64) error {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return common.NotFound("Cycle not found")
	}
	if err := s.db.Where("cycle_id = ?", cycleID).Delete(&model.IssueCycle{}).Error; err != nil {
		return common.Internal("Failed to cleanup cycle issues")
	}
	return s.db.Delete(&cycle).Error
}

// ==================== Status Transitions ====================

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

func (s *CycleService) AddIssue(cycleID, issueID uint64) error {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return common.NotFound("Cycle not found")
	}

	st := computeStatus(&cycle)
	if st == "completed" || st == "cancelled" {
		return common.BadRequest("Cannot add issues to a completed or cancelled cycle")
	}

	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if issue.ProjectID != cycle.ProjectID {
		return common.BadRequest("Issue does not belong to this cycle's project")
	}

	var count int64
	s.db.Model(&model.IssueCycle{}).Where("issue_id = ? AND cycle_id = ?", issueID, cycleID).Count(&count)
	if count > 0 {
		return common.BadRequest("Issue is already in this cycle")
	}

	// Remove issue from any existing cycle first (one-cycle-per-issue constraint)
	s.db.Where("issue_id = ?", issueID).Delete(&model.IssueCycle{})

	if err := s.db.Create(&model.IssueCycle{IssueID: issueID, CycleID: cycleID}).Error; err != nil {
		return common.Internal("Failed to add issue to cycle")
	}

	return nil
}

func (s *CycleService) RemoveIssue(cycleID, issueID uint64) error {
	result := s.db.Where("issue_id = ? AND cycle_id = ?", issueID, cycleID).Delete(&model.IssueCycle{})
	if result.RowsAffected == 0 {
		return common.NotFound("Issue is not in this cycle")
	}
	return nil
}

func (s *CycleService) ListIssues(cycleID uint64, stateID *uint64, priority string, limit, offset int) ([]response.IssueResponse, int64, error) {
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
		resp, err := issueSvc.BuildIssueResponse(&issue)
		if err != nil {
			return nil, 0, err
		}
		result[i] = *resp
	}

	return result, total, nil
}

// ==================== Analysis ====================

func (s *CycleService) GetProgress(cycleID uint64) (*response.CycleProgress, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	total, completed, err := s.countIssues(cycleID)
	if err != nil {
		log.Printf("cycle: error counting issues for cycle %d: %v", cycleID, err)
	}
	progress := 0.0
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}

	type stateRow struct {
		Name  string
		Group string
		Count int64
	}
	var rows []stateRow
	if err := s.db.Table("issue_cycles").
		Select("states.name, states.group, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issue_cycles.cycle_id = ?", cycleID).
		Group("states.name, states.group").
		Scan(&rows).Error; err != nil {
		log.Printf("cycle: error scanning state breakdown for cycle %d: %v", cycleID, err)
	}

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

func (s *CycleService) GetStatistics(cycleID uint64) (*response.CycleStatistics, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	prog, err := s.GetProgress(cycleID)
	if err != nil {
		return nil, err
	}

	type priorityRow struct {
		Priority string
		Count    int64
	}
	var pRows []priorityRow
	if err := s.db.Table("issue_cycles").
		Select("issues.priority, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ?", cycleID).
		Group("issues.priority").
		Scan(&pRows).Error; err != nil {
		log.Printf("cycle: error scanning priority breakdown for cycle %d: %v", cycleID, err)
	}

	priorityBreakdown := make(map[string]int64)
	for _, r := range pRows {
		priorityBreakdown[r.Priority] = r.Count
	}

	var withStart, withTarget int64
	s.db.Table("issue_cycles").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ? AND issues.start_date IS NOT NULL", cycleID).Count(&withStart)
	s.db.Table("issue_cycles").
		Joins("JOIN issues ON issues.id = issue_cycles.issue_id").
		Where("issue_cycles.cycle_id = ? AND issues.target_date IS NOT NULL", cycleID).Count(&withTarget)

	return &response.CycleStatistics{
		CycleProgress:     *prog,
		PriorityBreakdown: priorityBreakdown,
		IssueStats: response.IssueStats{
			Total:          prog.TotalIssues,
			WithStartDate:  withStart,
			WithTargetDate: withTarget,
		},
		DateRange: response.DateRange{
			StartDate: cycle.StartDate.Format("2006-01-02"),
			EndDate:   dateStr(cycle.EndDate),
		},
	}, nil
}

func (s *CycleService) GetBurndown(cycleID uint64) (*response.BurndownData, error) {
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return nil, common.NotFound("Cycle not found")
	}

	if cycle.EndDate == nil {
		return nil, common.BadRequest("Cycle does not have end date")
	}

	total, completed, err := s.countIssues(cycleID)
	if err != nil {
		log.Printf("cycle: error counting issues for cycle %d: %v", cycleID, err)
	}

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
