package service

import (
	"fmt"
	"time"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type IssueService struct {
	db *gorm.DB
}

func NewIssueService(db *gorm.DB) *IssueService {
	return &IssueService{db: db}
}

// Create creates a new issue.
func (s *IssueService) Create(req *request.IssueCreateRequest, projectID, workspaceID, userID uint64) (*response.IssueResponse, error) {
	// Validate project exists
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	// Get default state if not specified
	var stateID uint64
	if req.StateID != nil {
		stateID = *req.StateID
	} else {
		var defaultState model.State
		if err := s.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&defaultState).Error; err != nil {
			return nil, common.NotFound("No default state found for project")
		}
		stateID = defaultState.ID
	}

	// Auto-increment sequence_id
	var maxSeq int
	s.db.Model(&model.Issue{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(sequence_id), 0)").Scan(&maxSeq)

	priority := req.Priority
	if priority == "" {
		priority = common.PriorityNone
	}

	descHTML := req.DescriptionHTML
	if descHTML == "" {
		descHTML = "<p></p>"
	}

	issue := &model.Issue{
		Name:              req.Name,
		DescriptionHTML:   descHTML,
		DescriptionJSON:   req.DescriptionJSON,
		Priority:          priority,
		SequenceID:        maxSeq + 1,
		SortOrder:         65535,
		IsDraft:           false,
		ProjectID:         projectID,
		WorkspaceID:       workspaceID,
		StateID:           stateID,
		ParentID:          req.ParentID,
		ExternalID:        req.ExternalID,
		ExternalSource:    req.ExternalSource,
		IssueTypeID:       req.TypeID,
	}

	// Parse dates
	if req.StartDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartDate); err == nil {
			issue.StartDate = &t
		}
	}
	if req.TargetDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.TargetDate); err == nil {
			issue.TargetDate = &t
		}
	}

	tx := s.db.Begin()

	if err := tx.Create(issue).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to create issue")
	}

	// Create activity: created
	s.createActivity(tx, issue.ID, "created", nil, nil, nil, nil, &userID)

	// Add assignees
	for _, assigneeID := range req.AssigneeIDs {
		tx.Create(&model.IssueAssignee{IssueID: issue.ID, UserID: assigneeID})
	}

	// Add labels
	for _, labelID := range req.LabelIDs {
		tx.Create(&model.IssueLabel{IssueID: issue.ID, LabelID: labelID})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return s.buildResponse(issue.ID)
}

// GetByID returns an issue with all relations loaded.
func (s *IssueService) GetByID(issueID uint64) (*response.IssueResponse, error) {
	return s.buildResponse(issueID)
}

// List returns issues for a project with optional filters and total count.
func (s *IssueService) List(projectID uint64, filters map[string]interface{}, limit, offset int) ([]response.IssueResponse, int64, error) {
	baseQuery := s.db.Model(&model.Issue{}).Where("issues.project_id = ?", projectID)

	// Apply filters
	if stateID, ok := filters["state_id"]; ok && stateID != nil {
		baseQuery = baseQuery.Where("issues.state_id = ?", stateID)
	}
	if priority, ok := filters["priority"]; ok && priority != nil {
		baseQuery = baseQuery.Where("issues.priority = ?", priority)
	}
	if assigneeID, ok := filters["assignee_id"]; ok && assigneeID != nil {
		baseQuery = baseQuery.Joins("JOIN issue_assignees ON issue_assignees.issue_id = issues.id").
			Where("issue_assignees.user_id = ?", assigneeID)
	}
	if parentID, ok := filters["parent_id"]; ok && parentID != nil {
		baseQuery = baseQuery.Where("issues.parent_id = ?", parentID)
	}
	if cycleID, ok := filters["cycle_id"]; ok && cycleID != nil {
		baseQuery = baseQuery.Joins("JOIN issue_cycles ON issue_cycles.issue_id = issues.id").
			Where("issue_cycles.cycle_id = ?", cycleID)
	}
	if isDraft, ok := filters["is_draft"]; ok && isDraft != nil {
		baseQuery = baseQuery.Where("issues.is_draft = ?", isDraft)
	}
	if search, ok := filters["search"]; ok && search != nil {
		searchStr := fmt.Sprintf("%%%s%%", search)
		baseQuery = baseQuery.Where("issues.name ILIKE ? OR COALESCE(issues.description_stripped, '') ILIKE ?",
			searchStr, searchStr)
	}
	if issueTypeID, ok := filters["issue_type_id"]; ok && issueTypeID != nil {
		baseQuery = baseQuery.Where("issues.issue_type_id = ?", issueTypeID)
	}

	// Count total (before limit/offset)
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var issues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("IssueType").
		Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").
		Preload("CycleLink").
		Order("issues.sort_order ASC, issues.sequence_id DESC").
		Limit(limit).Offset(offset).
		Find(&issues).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	result := make([]response.IssueResponse, len(issues))
	for i, issue := range issues {
		resp, err := s.BuildIssueResponse(&issue)
		if err != nil {
			return nil, 0, err
		}
		result[i] = *resp
	}
	return result, total, nil
}

// Update updates an issue and records activity for changes.
func (s *IssueService) Update(issueID uint64, req *request.IssueUpdateRequest, userID uint64) (*response.IssueResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	tx := s.db.Begin()

	// Track changes for activity
	hasChanges := false

	if req.Name != nil && *req.Name != issue.Name {
		oldVal := issue.Name
		s.createActivity(tx, issueID, "updated", strPtr("name"), strPtr(oldVal), req.Name, nil, &userID)
		issue.Name = *req.Name
		hasChanges = true
	}
	if req.DescriptionHTML != nil && *req.DescriptionHTML != issue.DescriptionHTML {
		issue.DescriptionHTML = *req.DescriptionHTML
		hasChanges = true
	}
	if req.Priority != nil && *req.Priority != issue.Priority {
		oldVal := issue.Priority
		s.createActivity(tx, issueID, "updated", strPtr("priority"), strPtr(oldVal), req.Priority, nil, &userID)
		issue.Priority = *req.Priority
		hasChanges = true
	}
	if req.StateID != nil && *req.StateID != issue.StateID {
		oldStateID := issue.StateID
		issue.StateID = *req.StateID

		// Check if completed
		var newState model.State
		if err := s.db.First(&newState, *req.StateID).Error; err == nil {
			if newState.Group == common.StateGroupCompleted {
				now := time.Now()
				issue.CompletedAt = &now
			} else {
				issue.CompletedAt = nil
			}
		}

		s.createActivity(tx, issueID, "updated", strPtr("state_id"),
			strPtr(fmt.Sprintf("%d", oldStateID)), strPtr(fmt.Sprintf("%d", *req.StateID)), nil, &userID)
		hasChanges = true
	}

	// Save basic fields
	if hasChanges {
		if err := tx.Save(&issue).Error; err != nil {
			tx.Rollback()
			return nil, common.Internal("Failed to update issue")
		}
	}

	// Handle assignees
	if req.AssigneeIDs != nil {
		// Remove existing
		tx.Where("issue_id = ?", issueID).Delete(&model.IssueAssignee{})
		// Add new
		for _, aid := range req.AssigneeIDs {
			tx.Create(&model.IssueAssignee{IssueID: issueID, UserID: aid})
		}
		s.createActivity(tx, issueID, "updated", strPtr("assignees"), nil, nil, nil, &userID)
		hasChanges = true
	}

	// Handle labels
	if req.LabelIDs != nil {
		tx.Where("issue_id = ?", issueID).Delete(&model.IssueLabel{})
		for _, lid := range req.LabelIDs {
			tx.Create(&model.IssueLabel{IssueID: issueID, LabelID: lid})
		}
		s.createActivity(tx, issueID, "updated", strPtr("labels"), nil, nil, nil, &userID)
		hasChanges = true
	}

	// Handle cycle
	if req.CycleID != nil {
		tx.Where("issue_id = ?", issueID).Delete(&model.IssueCycle{})
		tx.Create(&model.IssueCycle{IssueID: issueID, CycleID: *req.CycleID})
		hasChanges = true
	}

	// Handle issue type
	if req.TypeID != nil {
		issue.IssueTypeID = req.TypeID
		tx.Save(&issue)
		hasChanges = true
	}

	// Parse dates
	if req.StartDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartDate); err == nil {
			issue.StartDate = &t
			tx.Save(&issue)
			hasChanges = true
		}
	}
	if req.TargetDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.TargetDate); err == nil {
			issue.TargetDate = &t
			tx.Save(&issue)
			hasChanges = true
		}
	}

	if !hasChanges {
		tx.Rollback()
		return s.buildResponse(issueID)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to commit update")
	}

	return s.buildResponse(issueID)
}

// Delete performs a soft delete on an issue.
func (s *IssueService) Delete(issueID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}
	return s.db.Delete(&issue).Error
}

// Archive archives an issue.
func (s *IssueService) Archive(issueID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}
	now := time.Now()
	issue.ArchivedAt = &now
	return s.db.Save(&issue).Error
}

// Restore restores an archived or deleted issue.
func (s *IssueService) Restore(issueID uint64) (*response.IssueResponse, error) {
	var issue model.Issue
	if err := s.db.Unscoped().First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	issue.ArchivedAt = nil
	issue.DeletedAt = gorm.DeletedAt{}
	s.db.Unscoped().Save(&issue)

	return s.buildResponse(issueID)
}

// AddAssignee adds an assignee to an issue.
func (s *IssueService) AddAssignee(issueID, userID, actorID uint64) error {
	// Verify issue exists
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	// Check if already assigned
	var count int64
	s.db.Model(&model.IssueAssignee{}).
		Where("issue_id = ? AND user_id = ?", issueID, userID).Count(&count)
	if count > 0 {
		return common.Conflict("User is already assigned to this issue")
	}

	if err := s.db.Create(&model.IssueAssignee{IssueID: issueID, UserID: userID}).Error; err != nil {
		return common.Internal("Failed to add assignee")
	}

	s.createActivity(s.db, issueID, "updated", strPtr("assignees"), nil, nil, nil, &actorID)
	return nil
}

// RemoveAssignee removes an assignee from an issue.
func (s *IssueService) RemoveAssignee(issueID, userID, actorID uint64) error {
	result := s.db.Where("issue_id = ? AND user_id = ?", issueID, userID).Delete(&model.IssueAssignee{})
	if result.RowsAffected == 0 {
		return common.NotFound("Assignee not found on this issue")
	}

	s.createActivity(s.db, issueID, "updated", strPtr("assignees"), nil, nil, nil, &actorID)
	return nil
}

// AddLabel adds a label to an issue.
func (s *IssueService) AddLabel(issueID, labelID, actorID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	var count int64
	s.db.Model(&model.IssueLabel{}).
		Where("issue_id = ? AND label_id = ?", issueID, labelID).Count(&count)
	if count > 0 {
		return common.Conflict("Label already added to this issue")
	}

	if err := s.db.Create(&model.IssueLabel{IssueID: issueID, LabelID: labelID}).Error; err != nil {
		return common.Internal("Failed to add label")
	}

	s.createActivity(s.db, issueID, "updated", strPtr("labels"), nil, nil, nil, &actorID)
	return nil
}

// RemoveLabel removes a label from an issue.
func (s *IssueService) RemoveLabel(issueID, labelID, actorID uint64) error {
	result := s.db.Where("issue_id = ? AND label_id = ?", issueID, labelID).Delete(&model.IssueLabel{})
	if result.RowsAffected == 0 {
		return common.NotFound("Label not found on this issue")
	}

	s.createActivity(s.db, issueID, "updated", strPtr("labels"), nil, nil, nil, &actorID)
	return nil
}

// SetCycle sets the cycle for an issue.
func (s *IssueService) SetCycle(issueID, cycleID, actorID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	// Remove existing cycle
	s.db.Where("issue_id = ?", issueID).Delete(&model.IssueCycle{})
	// Set new cycle
	if err := s.db.Create(&model.IssueCycle{IssueID: issueID, CycleID: cycleID}).Error; err != nil {
		return common.Internal("Failed to set cycle")
	}

	s.createActivity(s.db, issueID, "updated", strPtr("cycle"), nil, nil, nil, &actorID)
	return nil
}

// RemoveCycle removes the cycle from an issue.
func (s *IssueService) RemoveCycle(issueID, actorID uint64) error {
	s.db.Where("issue_id = ?", issueID).Delete(&model.IssueCycle{})
	s.createActivity(s.db, issueID, "updated", strPtr("cycle"), nil, nil, nil, &actorID)
	return nil
}

// GetActivities returns issue activity history.
func (s *IssueService) GetActivities(issueID uint64, limit, offset int) ([]response.IssueActivityResponse, error) {
	var activities []model.IssueActivity
	if err := s.db.Where("issue_id = ?", issueID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&activities).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.IssueActivityResponse, len(activities))
	for i, a := range activities {
		result[i] = response.IssueActivityResponse{
			ID:        a.ID,
			IssueID:   a.IssueID,
			Verb:      a.Verb,
			Field:     a.Field,
			OldValue:  a.OldValue,
			NewValue:  a.NewValue,
			Comment:   a.Comment,
			ActorID:   a.ActorID,
			CreatedAt: a.CreatedAt,
		}
	}
	return result, nil
}

// GetStatistics returns issue statistics for a project.
func (s *IssueService) GetStatistics(projectID uint64) (*response.IssueStatistics, error) {
	stats := &response.IssueStatistics{
		ByState:    make(map[string]int64),
		ByPriority: make(map[string]int64),
	}

	// Total count
	s.db.Model(&model.Issue{}).Where("project_id = ?", projectID).Count(&stats.Total)

	// Completed
	s.db.Model(&model.Issue{}).
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issues.project_id = ? AND states.group = ?", projectID, common.StateGroupCompleted).
		Count(&stats.CompletedCount)

	// Drafts
	s.db.Model(&model.Issue{}).Where("project_id = ? AND is_draft = ?", projectID, true).
		Count(&stats.DraftCount)

	// By state
	var states []model.State
	s.db.Where("project_id = ?", projectID).Find(&states)
	for _, st := range states {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND state_id = ?", projectID, st.ID).Count(&count)
		stats.ByState[st.Name] = count
	}

	// By priority
	for _, p := range []string{common.PriorityUrgent, common.PriorityHigh, common.PriorityMedium, common.PriorityLow, common.PriorityNone} {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND priority = ?", projectID, p).Count(&count)
		stats.ByPriority[p] = count
	}

	return stats, nil
}

// Search searches issues by query across a workspace.
func (s *IssueService) Search(workspaceID uint64, query string, projectID *uint64, limit int) ([]response.IssueSearchResult, error) {
	q := s.db.Model(&model.Issue{}).
		Joins("JOIN projects ON projects.id = issues.project_id").
		Where("issues.workspace_id = ?", workspaceID).
		Where("issues.name ILIKE ? OR COALESCE(issues.description_stripped, '') ILIKE ?",
			"%"+query+"%", "%"+query+"%")

	if projectID != nil {
		q = q.Where("issues.project_id = ?", *projectID)
	}

	var issues []model.Issue
	if err := q.Preload("Project").Limit(limit).Find(&issues).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.IssueSearchResult, len(issues))
	for i, issue := range issues {
		result[i] = response.IssueSearchResult{
			ID:                issue.ID,
			Name:              issue.Name,
			SequenceID:        issue.SequenceID,
			ProjectIdentifier: issue.Project.Identifier,
			ProjectID:         issue.ProjectID,
		}
	}
	return result, nil
}

// BulkUpdate updates multiple issues with the same changes.
func (s *IssueService) BulkUpdate(projectID uint64, req *request.BulkUpdateRequest, userID uint64) ([]response.IssueResponse, error) {
	var results []response.IssueResponse
	for _, issueID := range req.IssueIDs {
		updateReq := &request.IssueUpdateRequest{
			Priority:     req.Priority,
			StateID:      req.StateID,
			AssigneeIDs:  req.AssigneeIDs,
			LabelIDs:     req.LabelIDs,
			StartDate:    req.StartDate,
			TargetDate:   req.TargetDate,
		}
		resp, err := s.Update(issueID, updateReq, userID)
		if err != nil {
			continue // Skip failed, continue with others
		}
		results = append(results, *resp)
	}
	return results, nil
}

// BulkDelete deletes multiple issues.
func (s *IssueService) BulkDelete(issueIDs []uint64) error {
	for _, id := range issueIDs {
		s.Delete(id)
	}
	return nil
}

// ==================== Private helpers ====================

func (s *IssueService) buildResponse(issueID uint64) (*response.IssueResponse, error) {
	var issue model.Issue
	if err := s.db.
		Preload("State").
		Preload("Project").
		Preload("IssueType").
		Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").
		Preload("CycleLink").
		Preload("SubIssues").
		First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	return s.BuildIssueResponse(&issue)
}

func (s *IssueService) BuildIssueResponse(issue *model.Issue) (*response.IssueResponse, error) {
	resp := &response.IssueResponse{
		ID:              issue.ID,
		Name:            issue.Name,
		DescriptionHTML: issue.DescriptionHTML,
		DescriptionJSON: issue.DescriptionJSON,
		Priority:        issue.Priority,
		SequenceID:      issue.SequenceID,
		SortOrder:       issue.SortOrder,
		StartDate:       issue.StartDate,
		TargetDate:      issue.TargetDate,
		CompletedAt:     issue.CompletedAt,
		IsDraft:         issue.IsDraft,
		ArchivedAt:      issue.ArchivedAt,
		ProjectID:       issue.ProjectID,
		WorkspaceID:     issue.WorkspaceID,
		StateID:         issue.StateID,
		ParentID:        issue.ParentID,
		ExternalID:      issue.ExternalID,
		ExternalSource:  issue.ExternalSource,
		LinkCount:       0,
		AttachmentCount: 0,
		CreatedAt:       issue.CreatedAt,
		UpdatedAt:       issue.UpdatedAt,
		CreatedByID:     issue.CreatedByID,
		UpdatedByID:     issue.UpdatedByID,
	}

	// State info
	if issue.State.ID != 0 {
		resp.StateName = issue.State.Name
		resp.StateGroup = issue.State.Group
	}

	// Project info
	if issue.Project.ID != 0 {
		resp.Project = &response.ProjectLite{
			ID:         issue.Project.ID,
			Name:       issue.Project.Name,
			Identifier: issue.Project.Identifier,
		}
	}

	// Issue Type info
	if issue.IssueType.ID != 0 {
		resp.IssueType = &response.IssueTypeLite{
			ID:    issue.IssueType.ID,
			Name:  issue.IssueType.Name,
			Color: issue.IssueType.Color,
			Icon:  issue.IssueType.Icon,
		}
	}

	// Assignees
	resp.Assignees = make([]response.UserLite, 0)
	resp.Assignees = []response.UserLite{} // ensure not null
	for _, link := range issue.AssigneeLinks {
		if link.User.ID != 0 {
			resp.Assignees = append(resp.Assignees, response.UserLite{
				ID:          link.User.ID,
				DisplayName: link.User.DisplayName,
				Email:       link.User.Email,
			})
		}
	}

	// Labels
	resp.Labels = make([]uint64, 0)
	resp.LabelDetails = make([]response.LabelLite, 0)
	for _, link := range issue.LabelLinks {
		if link.Label.ID != 0 {
			resp.Labels = append(resp.Labels, link.LabelID)
			// Note: GORM preload puts Label data in Label field
			if link.Label.Name != "" {
				resp.LabelDetails = append(resp.LabelDetails, response.LabelLite{
					ID:    link.Label.ID,
					Name:  link.Label.Name,
					Color: link.Label.Color,
				})
			}
		}
	}

	// Cycle
	if issue.CycleLink != nil {
		resp.CycleID = &issue.CycleLink.CycleID
	}

	// Sub-issues count
	resp.SubIssuesCount = int64(len(issue.SubIssues))

	// Soft delete
	if issue.DeletedAt.Valid {
		resp.DeletedAt = &issue.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp, nil
}

func (s *IssueService) createActivity(db *gorm.DB, issueID uint64, verb string, field, oldValue, newValue, comment *string, actorID *uint64) {
	activity := &model.IssueActivity{
		IssueID:  &issueID,
		Verb:     verb,
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
		Comment:  comment,
		ActorID:  actorID,
	}
	db.Create(activity)
}

func strPtr(s string) *string {
	return &s
}
