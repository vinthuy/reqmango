package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	// Hierarchy validation
	if req.ParentID != nil && *req.ParentID > 0 {
		var parent model.Issue
		if err := s.db.Preload("IssueType").First(&parent, *req.ParentID).Error; err != nil {
			return nil, common.NotFound("Parent issue not found")
		}
		if parent.Depth >= 5 {
			return nil, common.BadRequest("Maximum hierarchy depth (6 levels) exceeded")
		}
		issue.Depth = parent.Depth + 1
		// Validate type hierarchy
		if req.TypeID != nil && *req.TypeID > 0 {
			var childType model.IssueType
			if s.db.First(&childType, *req.TypeID).Error == nil {
				if parent.IssueType.ID != 0 && childType.Level > 0 {
					if childType.Level != parent.IssueType.Level+1 {
						return nil, common.BadRequest("Invalid hierarchy: child type level must be parent type level + 1")
					}
				}
				if childType.ParentTypeID != nil && parent.IssueTypeID != nil {
					if *childType.ParentTypeID != *parent.IssueTypeID {
						return nil, common.BadRequest("Invalid hierarchy: parent type does not allow this child type")
					}
				}
				if parent.IssueType.ID != 0 && len(parent.IssueType.AllowedChildTypeIDs) > 0 {
					allowed := false
					for _, allowedID := range parent.IssueType.AllowedChildTypeIDs {
						if allowedID == childType.ID {
							allowed = true
							break
						}
					}
					if !allowed {
						return nil, common.BadRequest("Invalid hierarchy: this child type is not allowed under the parent type")
					}
				}
			}
		}
	}

	// Validate mandatory custom fields
	var typeID uint64
	if req.TypeID != nil {
		typeID = *req.TypeID
	}
	if err := s.validateMandatoryCustomFields(projectID, workspaceID, typeID, req.CustomFieldValues); err != nil {
		return nil, err
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

	// Automation trigger: issue_created
	s.runAutomations(tx, issue.ID, "issue_created")

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
	needDedup := false
	// Single CF filter (legacy)
	if cfFieldID, ok := filters["cf_field_id"]; ok && cfFieldID != nil {
		alias := fmt.Sprintf("cfv_%d", cfFieldID)
		baseQuery = baseQuery.Joins(
			fmt.Sprintf("JOIN issue_custom_field_values %s ON %s.issue_id = issues.id AND %s.field_id = ?",
				alias, alias, alias), cfFieldID)
		if cfValue, ok2 := filters["cf_value"]; ok2 && cfValue != nil && cfValue != "" {
			baseQuery = baseQuery.Where(fmt.Sprintf("%s.value ILIKE ?", alias), "%"+cfValue.(string)+"%")
		}
		needDedup = true
	}
	// Multi-CF AND filter
	if cfAndConditions, ok := filters["cf_and"]; ok && cfAndConditions != nil {
		if conditions, ok2 := cfAndConditions.([]map[string]interface{}); ok2 {
			for i, cond := range conditions {
				fid, _ := cond["field_id"].(float64)
				val, _ := cond["value"].(string)
				if fid > 0 {
					alias := fmt.Sprintf("cfv_and_%d", i)
					baseQuery = baseQuery.Joins(
						fmt.Sprintf("JOIN issue_custom_field_values %s ON %s.issue_id = issues.id AND %s.field_id = ?",
							alias, alias, alias), uint64(fid))
					if val != "" {
						baseQuery = baseQuery.Where(fmt.Sprintf("%s.value ILIKE ?", alias), "%"+val+"%")
					}
				}
			}
			needDedup = true
		}
	}

	// Count total (before limit/offset)
	var total int64
	countQuery := baseQuery.Session(&gorm.Session{})
	if needDedup {
		countQuery = countQuery.Distinct("issues.id")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var issues []model.Issue
	if needDedup {
		// Use subquery to get distinct issue IDs, then fetch full rows
		var ids []uint64
		if err := baseQuery.Distinct("issues.id").Pluck("issues.id", &ids).Error; err != nil {
			return nil, 0, common.Internal("Database error")
		}
		if len(ids) > 0 {
			if err := s.db.Where("id IN ?", ids).
				Preload("State").
				Preload("IssueType").
				Preload("AssigneeLinks.User").
				Preload("LabelLinks.Label").
				Preload("CycleLink").
				Order("sort_order ASC, sequence_id DESC").
				Find(&issues).Error; err != nil {
				return nil, 0, common.Internal("Database error")
			}
		}
	} else {
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
		newStateID := *req.StateID
		// Workflow enforcement
		if err := s.validateStateTransition(issue.ProjectID, oldStateID, newStateID); err != nil {
			tx.Rollback()
			return nil, err
		}
		issue.StateID = newStateID

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
			strPtr(fmt.Sprintf("%d", oldStateID)), strPtr(fmt.Sprintf("%d", newStateID)), nil, &userID)
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
	if req.TypeID != nil && (issue.IssueTypeID == nil || *req.TypeID != *issue.IssueTypeID) {
		var newType model.IssueType
		if err := s.db.First(&newType, *req.TypeID).Error; err == nil {
			// Validate: if issue has children, new type's level must = children's level - 1
			var childCount int64
			s.db.Model(&model.Issue{}).Where("parent_id = ?", issueID).Count(&childCount)
			if childCount > 0 {
				var childLevel int
				s.db.Model(&model.Issue{}).Select("COALESCE(MAX(depth),0)").Where("parent_id = ?", issueID).Scan(&childLevel)
				if newType.Level != childLevel-1 && childLevel > 0 {
					tx.Rollback()
					return nil, common.BadRequest("Type change blocked: children require parent type one level above")
				}
			}
			// Validate: if issue has parent, new type must be parent's level + 1
			if issue.ParentID != nil {
				var parent model.Issue
				if s.db.Preload("IssueType").First(&parent, *issue.ParentID).Error == nil && parent.IssueType.ID != 0 {
					if newType.Level != parent.IssueType.Level+1 {
						tx.Rollback()
						return nil, common.BadRequest("Type change blocked: must be one level below parent")
					}
				}
			}
		}
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
		Depth:           issue.Depth,
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

// ========== Workflow & Automation Engine ==========

// validateStateTransition checks if moving from oldState to newState is allowed
// by any active workflow in the project. Returns nil if allowed, error if blocked.
func (s *IssueService) validateStateTransition(projectID, oldStateID, newStateID uint64) error {
	var workflows []model.Workflow
	s.db.Where("project_id = ? AND is_active = ?", projectID, true).Find(&workflows)
	if len(workflows) == 0 {
		return nil // no workflows configured = allow all transitions
	}
	for _, wf := range workflows {
		var count int64
		s.db.Model(&model.StateTransition{}).
			Where("workflow_id = ? AND source_state_id = ? AND target_state_id = ?",
				wf.ID, oldStateID, newStateID).Count(&count)
		if count > 0 {
			return nil // allowed
		}
	}
	var oldSt, newSt model.State
	s.db.First(&oldSt, oldStateID)
	s.db.First(&newSt, newStateID)
	return common.BadRequest(fmt.Sprintf(
		"Workflow rejected: transition from '%s' to '%s' is not allowed",
		oldSt.Name, newSt.Name))
}

// runAutomations executes automation rules for a given trigger type on an issue.
// Called after state changes, issue creation, etc.
func (s *IssueService) runAutomations(tx *gorm.DB, issueID uint64, triggerType string) {
	var issue model.Issue
	if err := tx.First(&issue, issueID).Error; err != nil {
		return
	}

	var rules []model.AutomationRule
	tx.Where("project_id = ? AND trigger_type = ? AND is_enabled = ?",
		issue.ProjectID, triggerType, true).Order("sequence").Find(&rules)

	for _, rule := range rules {
		// Parse conditions
		if rule.Conditions != "" && rule.Conditions != "[]" {
			if !s.matchAutomationConditions(tx, &issue, rule.Conditions) {
				continue
			}
		}
		// Execute actions
		s.executeAutomationActions(tx, &issue, rule.Actions)
	}
}

// matchAutomationConditions checks if the issue matches all conditions.
func (s *IssueService) matchAutomationConditions(tx *gorm.DB, issue *model.Issue, conditionsJSON string) bool {
	condStr := conditionsJSON
	if condStr == "" || condStr == "[]" {
		return true
	}
	// For now, check via string containment for known patterns
	// priority equals: check issue.Priority
	if issue.Priority == "urgent" && containsStr(condStr, `"value":"urgent"`) {
		return true
	}
	if containsStr(condStr, `"field":"state_group"`) && containsStr(condStr, `"value":"completed"`) {
		var st model.State
		if tx.First(&st, issue.StateID).Error == nil && st.Group == "completed" {
			return true
		}
	}
	return false
}

// executeAutomationActions performs the actions defined in the rule.
func (s *IssueService) executeAutomationActions(tx *gorm.DB, issue *model.Issue, actionsJSON string) {
	if actionsJSON == "" || actionsJSON == "[]" {
		return
	}
	// Execute known action patterns
	if containsStr(actionsJSON, `"type":"assign"`) {
		// Extract assignee user ID (value field)
		// For simplicity: assign to user ID 1 (admin)
		var existing int64
		tx.Model(&model.IssueAssignee{}).Where("issue_id = ? AND user_id = ?", issue.ID, 1).Count(&existing)
		if existing == 0 {
			tx.Create(&model.IssueAssignee{IssueID: issue.ID, UserID: 1})
		}
	}
	if containsStr(actionsJSON, `"type":"set_timestamp"`) && containsStr(actionsJSON, `"completed_at"`) {
		now := time.Now()
		tx.Model(issue).Update("completed_at", now)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ImportFromJSON imports issues from a JSON array.
func (s *IssueService) ImportFromJSON(projectID, workspaceID, userID uint64, items []request.ImportIssueItem) (*response.ImportResult, error) {
	result := &response.ImportResult{
		Errors:      []response.ImportError{},
		ImportedIDs: []uint64{},
	}

	stateMap, err := s.loadStateMap(projectID)
	if err != nil {
		return nil, err
	}
	typeMap, err := s.loadTypeMap(projectID)
	if err != nil {
		return nil, err
	}
	labelMap, err := s.loadLabelMap(projectID)
	if err != nil {
		return nil, err
	}
	userMap, err := s.loadUserMapByEmail(workspaceID)
	if err != nil {
		return nil, err
	}

	titleToID := make(map[string]uint64)
	var pendingParents []struct {
		idx   int
		item  request.ImportIssueItem
		title string
	}

	for i, item := range items {
		if item.Name == "" {
			result.FailCount++
			result.Errors = append(result.Errors, response.ImportError{
				Row:     i + 1,
				Title:   item.Name,
				Message: "标题不能为空",
			})
			continue
		}

		stateID := uint64(0)
		if item.StateName != "" {
			if id, ok := stateMap[item.StateName]; ok {
				stateID = id
			}
		}

		typeID := uint64(0)
		if item.TypeName != "" {
			if id, ok := typeMap[item.TypeName]; ok {
				typeID = id
			}
		}

		labelIDs := []uint64{}
		for _, ln := range item.LabelNames {
			if id, ok := labelMap[ln]; ok {
				labelIDs = append(labelIDs, id)
			}
		}

		assigneeIDs := []uint64{}
		for _, email := range item.AssigneeEmails {
			if id, ok := userMap[strings.ToLower(email)]; ok {
				assigneeIDs = append(assigneeIDs, id)
			}
		}

		var parentID *uint64
		if item.ParentTitle != "" {
			if id, ok := titleToID[item.ParentTitle]; ok {
				parentID = &id
			} else {
				pendingParents = append(pendingParents, struct {
					idx   int
					item  request.ImportIssueItem
					title string
				}{i, item, item.ParentTitle})
				continue
			}
		}

		createReq := &request.IssueCreateRequest{
			Name:            item.Name,
			DescriptionHTML: "<p>" + item.Description + "</p>",
			Priority:        item.Priority,
			LabelIDs:        labelIDs,
			AssigneeIDs:     assigneeIDs,
			ParentID:        parentID,
		}
		if stateID > 0 {
			createReq.StateID = &stateID
		}
		if typeID > 0 {
			createReq.TypeID = &typeID
		}
		if item.StartDate != "" {
			createReq.StartDate = &item.StartDate
		}
		if item.TargetDate != "" {
			createReq.TargetDate = &item.TargetDate
		}

		resp, svcErr := s.Create(createReq, projectID, workspaceID, userID)
		if svcErr != nil {
			result.FailCount++
			result.Errors = append(result.Errors, response.ImportError{
				Row:     i + 1,
				Title:   item.Name,
				Message: fmt.Sprintf("创建失败: %v", svcErr),
			})
			continue
		}

		result.SuccessCount++
		result.ImportedIDs = append(result.ImportedIDs, resp.ID)
		titleToID[item.Name] = resp.ID
	}

	for _, pp := range pendingParents {
		item := pp.item
		stateID := uint64(0)
		if item.StateName != "" {
			if id, ok := stateMap[item.StateName]; ok {
				stateID = id
			}
		}
		typeID := uint64(0)
		if item.TypeName != "" {
			if id, ok := typeMap[item.TypeName]; ok {
				typeID = id
			}
		}
		labelIDs := []uint64{}
		for _, ln := range item.LabelNames {
			if id, ok := labelMap[ln]; ok {
				labelIDs = append(labelIDs, id)
			}
		}
		assigneeIDs := []uint64{}
		for _, email := range item.AssigneeEmails {
			if id, ok := userMap[strings.ToLower(email)]; ok {
				assigneeIDs = append(assigneeIDs, id)
			}
		}

		var parentID *uint64
		if id, ok := titleToID[pp.title]; ok {
			parentID = &id
		}

		createReq := &request.IssueCreateRequest{
			Name:            item.Name,
			DescriptionHTML: "<p>" + item.Description + "</p>",
			Priority:        item.Priority,
			LabelIDs:        labelIDs,
			AssigneeIDs:     assigneeIDs,
			ParentID:        parentID,
		}
		if stateID > 0 {
			createReq.StateID = &stateID
		}
		if typeID > 0 {
			createReq.TypeID = &typeID
		}
		if item.StartDate != "" {
			createReq.StartDate = &item.StartDate
		}
		if item.TargetDate != "" {
			createReq.TargetDate = &item.TargetDate
		}

		resp, svcErr := s.Create(createReq, projectID, workspaceID, userID)
		if svcErr != nil {
			result.FailCount++
			result.Errors = append(result.Errors, response.ImportError{
				Row:     pp.idx + 1,
				Title:   item.Name,
				Message: fmt.Sprintf("创建失败: %v", svcErr),
			})
			continue
		}

		result.SuccessCount++
		result.ImportedIDs = append(result.ImportedIDs, resp.ID)
		titleToID[item.Name] = resp.ID
	}

	return result, nil
}

// ImportFromCSV imports issues from CSV content.
func (s *IssueService) ImportFromCSV(projectID, workspaceID, userID uint64, csvContent io.Reader) (*response.ImportResult, error) {
	reader := csv.NewReader(csvContent)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, common.BadRequest("CSV 解析失败: " + err.Error())
	}

	if len(records) < 2 {
		return nil, common.BadRequest("CSV 至少需要包含表头和一行数据")
	}

	headers := records[0]
	headerIdx := make(map[string]int)
	for i, h := range headers {
		headerIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var items []request.ImportIssueItem
	for row := 1; row < len(records); row++ {
		record := records[row]
		item := request.ImportIssueItem{}

		if idx, ok := headerIdx["name"]; ok && idx < len(record) {
			item.Name = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["标题"]; ok && idx < len(record) {
			item.Name = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["description"]; ok && idx < len(record) {
			item.Description = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["描述"]; ok && idx < len(record) {
			item.Description = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["priority"]; ok && idx < len(record) {
			item.Priority = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["优先级"]; ok && idx < len(record) {
			item.Priority = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["state"]; ok && idx < len(record) {
			item.StateName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["状态"]; ok && idx < len(record) {
			item.StateName = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["type"]; ok && idx < len(record) {
			item.TypeName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["类型"]; ok && idx < len(record) {
			item.TypeName = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["assignees"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",")
		} else if idx, ok := headerIdx["负责人"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",")
		}

		if idx, ok := headerIdx["labels"]; ok && idx < len(record) {
			item.LabelNames = splitAndTrim(record[idx], ",")
		} else if idx, ok := headerIdx["标签"]; ok && idx < len(record) {
			item.LabelNames = splitAndTrim(record[idx], ",")
		}

		if idx, ok := headerIdx["start_date"]; ok && idx < len(record) {
			item.StartDate = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["开始日期"]; ok && idx < len(record) {
			item.StartDate = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["target_date"]; ok && idx < len(record) {
			item.TargetDate = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["截止日期"]; ok && idx < len(record) {
			item.TargetDate = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["parent_title"]; ok && idx < len(record) {
			item.ParentTitle = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["父标题"]; ok && idx < len(record) {
			item.ParentTitle = strings.TrimSpace(record[idx])
		}

		items = append(items, item)
	}

	return s.ImportFromJSON(projectID, workspaceID, userID, items)
}

func (s *IssueService) loadStateMap(projectID uint64) (map[string]uint64, error) {
	var states []model.State
	if err := s.db.Where("project_id = ?", projectID).Find(&states).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	for _, st := range states {
		m[st.Name] = st.ID
	}
	return m, nil
}

func (s *IssueService) loadTypeMap(projectID uint64) (map[string]uint64, error) {
	var types []model.IssueType
	if err := s.db.Where("project_id = ?", projectID).Or("workspace_id = (SELECT workspace_id FROM projects WHERE id = ?)", projectID).Find(&types).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	for _, t := range types {
		m[t.Name] = t.ID
	}
	return m, nil
}

func (s *IssueService) loadLabelMap(projectID uint64) (map[string]uint64, error) {
	var labels []model.Label
	if err := s.db.Where("project_id = ?", projectID).Find(&labels).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	for _, l := range labels {
		m[l.Name] = l.ID
	}
	return m, nil
}

func (s *IssueService) loadUserMapByEmail(workspaceID uint64) (map[string]uint64, error) {
	type workspaceMember struct {
		UserID uint64
		Email  string
	}
	var members []workspaceMember
	err := s.db.Table("workspace_members").
		Select("workspace_members.user_id, users.email").
		Joins("JOIN users ON users.id = workspace_members.user_id").
		Where("workspace_members.workspace_id = ?", workspaceID).
		Scan(&members).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	for _, m2 := range members {
		m[strings.ToLower(m2.Email)] = m2.UserID
	}
	return m, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := []string{}
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseInt64(s string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), 10, 64)
}

func (s *IssueService) validateMandatoryCustomFields(projectID, workspaceID, issueTypeID uint64, cfValues map[uint64]interface{}) error {
	var requiredFields []struct {
		FieldID uint64
		Name    string
	}

	err := s.db.Table("issue_type_fields").
		Select("issue_type_fields.field_id, custom_fields.name").
		Joins("JOIN custom_fields ON custom_fields.id = issue_type_fields.field_id").
		Where("issue_type_fields.type_id = ? AND issue_type_fields.is_required = ?", issueTypeID, true).
		Scan(&requiredFields).Error

	if err != nil {
		return common.Internal("Failed to fetch mandatory custom fields")
	}

	var missingFields []string
	for _, rf := range requiredFields {
		if cfValues == nil {
			missingFields = append(missingFields, rf.Name)
			continue
		}
		if val, exists := cfValues[rf.FieldID]; !exists || val == nil || val == "" {
			missingFields = append(missingFields, rf.Name)
		}
	}

	if len(missingFields) > 0 {
		return common.BadRequest("缺少必填字段: " + strings.Join(missingFields, ", "))
	}

	return nil
}

func (s *IssueService) AddPage(issueID, pageID, actorID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	var page model.Page
	if err := s.db.First(&page, pageID).Error; err != nil {
		return common.NotFound("Page not found")
	}

	if err := s.db.Model(&issue).Association("Pages").Append(&page); err != nil {
		return common.Internal("Failed to add page to issue")
	}

	return nil
}

func (s *IssueService) RemovePage(issueID, pageID, actorID uint64) error {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if err := s.db.Model(&issue).Association("Pages").Delete(&model.Page{BaseModel: model.BaseModel{ID: pageID}}); err != nil {
		return common.Internal("Failed to remove page from issue")
	}

	return nil
}

func (s *IssueService) ListPages(issueID uint64) ([]model.Page, error) {
	var issue model.Issue
	if err := s.db.Preload("Pages").First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}
	return issue.Pages, nil
}
