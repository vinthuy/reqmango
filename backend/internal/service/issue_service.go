package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/rql"
	"gorm.io/gorm"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func sanitizeHTML(html string) string {
	if html == "" {
		return "<p></p>"
	}
	policy := bluemonday.UGCPolicy()
	return policy.Sanitize(html)
}

func stripHTMLTags(html string) string {
	if html == "" {
		return ""
	}
	return htmlTagRegex.ReplaceAllString(html, "")
}

type IssueService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
	webhookSvc      *WebhookService
	automationSvc   *AutomationService
	slackSvc        *SlackService
}

func NewIssueService(db *gorm.DB, notificationSvc *NotificationService, webhookSvc *WebhookService, automationSvc *AutomationService, slackSvc *SlackService) *IssueService {
	return &IssueService{db: db, notificationSvc: notificationSvc, webhookSvc: webhookSvc, automationSvc: automationSvc, slackSvc: slackSvc}
}

// DB returns the database instance for use in handlers (for security checks).
func (s *IssueService) DB() *gorm.DB {
	return s.db
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

	// SequenceID will be assigned inside the transaction below to prevent race conditions

	priority := req.Priority
	if priority == "" {
		priority = common.PriorityNone
	}

	descHTML := sanitizeHTML(req.DescriptionHTML)
	descStripped := stripHTMLTags(descHTML)

	issue := &model.Issue{
		Name:                req.Name,
		DescriptionHTML:     descHTML,
		DescriptionJSON:     req.DescriptionJSON,
		DescriptionStripped: &descStripped,
		Priority:            priority,
		SortOrder:           65535,
		IsDraft:             false,
		ProjectID:           projectID,
		WorkspaceID:         workspaceID,
		StateID:             stateID,
		ParentID:            req.ParentID,
		ExternalID:          req.ExternalID,
		ExternalSource:      req.ExternalSource,
		CoverImageURL:       req.CoverImageURL,
		IssueTypeID:         req.TypeID,
	}

	// Hierarchy validation
	if req.ParentID != nil && *req.ParentID > 0 {
		if *req.ParentID == issue.ID {
			return nil, common.BadRequest("An issue cannot be its own parent")
		}
		if s.checkCircularReference(issue.ID, *req.ParentID) {
			return nil, common.BadRequest("Cannot create circular reference in issue hierarchy")
		}
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
		} else {
			return nil, common.BadRequest("Invalid start_date format, expected RFC3339")
		}
	}
	if req.TargetDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.TargetDate); err == nil {
			issue.TargetDate = &t
		} else {
			return nil, common.BadRequest("Invalid target_date format, expected RFC3339")
		}
	}

	// Validate assignees are project members
	for _, assigneeID := range req.AssigneeIDs {
		var count int64
		s.db.Model(&model.ProjectMember{}).
			Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, assigneeID, true).
			Count(&count)
		if count == 0 {
			return nil, common.BadRequest(fmt.Sprintf("User %d is not a member of this project", assigneeID))
		}
	}

	tx := s.db.Begin()

	// Auto-increment sequence_id (inside transaction to prevent race conditions)
	var maxSeq int
	tx.Model(&model.Issue{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(sequence_id), 0)").Scan(&maxSeq)
	issue.SequenceID = maxSeq + 1

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

	// Save custom field values
	for fieldID, value := range req.CustomFieldValues {
		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		default:
			stringValue = fmt.Sprintf("%v", v)
		}
		tx.Create(&model.IssueCustomFieldValue{
			IssueID: issue.ID,
			FieldID: fieldID,
			Value:   stringValue,
		})
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	// Automation trigger: issue_created (after commit, uses own DB connection)
	s.runAutomations(issue.ID, "issue_created", map[string]interface{}{
		"issue_id": issue.ID, "priority": issue.Priority,
		"state_id": issue.StateID, "project_id": issue.ProjectID,
	})

	s.webhookSvc.Fire(projectID, "issue_created", map[string]interface{}{"issue_id": issue.ID, "name": issue.Name, "priority": issue.Priority})
	return s.buildResponse(issue.ID)
}

// GetByID returns an issue with all relations loaded.
func (s *IssueService) GetByID(issueID uint64) (*response.IssueResponse, error) {
	return s.buildResponse(issueID)
}

// buildSortClause builds ORDER BY clause from sort_by/sort_dir (single, backward compat)
// or sort_config (multi-field sort, preferred).
func (s *IssueService) buildSortClause(filters map[string]interface{}) string {
	colMap := map[string]string{
		"sequence_id": "issues.sequence_id",
		"name":        "issues.name",
		"priority":    "issues.priority",
		"state":       "issues.state_id",
		"issue_type":  "issues.issue_type_id",
		"created_at":  "issues.created_at",
		"updated_at":  "issues.updated_at",
		"start_date":  "issues.start_date",
		"target_date": "issues.target_date",
	}

	// Prefer sort_config (multi-field)
	if sortConfig, ok := filters["sort_config"].([]map[string]string); ok && len(sortConfig) > 0 {
		var parts []string
		for _, entry := range sortConfig {
			field := entry["field"]
			dir := entry["dir"]
			if dir != "asc" && dir != "desc" {
				dir = "desc"
			}
			if col, ok := colMap[field]; ok {
				parts = append(parts, fmt.Sprintf("%s %s", col, dir))
			}
		}
		if len(parts) > 0 {
			return fmt.Sprintf("%s, issues.sort_order ASC, issues.sequence_id DESC", strings.Join(parts, ", "))
		}
		return ""
	}

	// Fallback to sort_by / sort_dir (single, backward compat)
	sortBy, _ := filters["sort_by"].(string)
	sortDir, _ := filters["sort_dir"].(string)
	if sortBy == "" {
		return ""
	}
	if sortDir == "" {
		sortDir = "desc"
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}
	if col, ok := colMap[sortBy]; ok {
		return fmt.Sprintf("%s %s, issues.sort_order ASC, issues.sequence_id DESC", col, sortDir)
	}
	return ""
}

// List returns issues for a project with optional filters and total count.
func (s *IssueService) List(projectID uint64, filters map[string]interface{}, limit, offset int) ([]response.IssueResponse, int64, error) {
	baseQuery := s.db.Model(&model.Issue{}).Where("issues.project_id = ?", projectID)

	// Build dynamic ORDER BY from sort_by / sort_dir
	sortClause := s.buildSortClause(filters)

	// Handle RQL filter: use RQL service to get matching issue IDs
	if rqlQuery, ok := filters["rql"]; ok && rqlQuery != nil && rqlQuery != "" {
		queryStr := fmt.Sprint(rqlQuery)
		if queryStr != "" {
			rqlSvc := rql.NewRQLService()
			page := 1
			if limit > 0 {
				page = (offset / limit) + 1
			}
			// Extract current user ID for template variable resolution ($CURRENT_USER)
			var currentUserID uint64
			if uid, ok := filters["current_user_id"]; ok {
				switch v := uid.(type) {
				case uint64:
					currentUserID = v
				case uint:
					currentUserID = uint64(v)
				case float64:
					currentUserID = uint64(v)
				}
			}
			rqlIssues, rqlTotal, rqlErr := rqlSvc.SearchIssuesWithUser(s.db, projectID, queryStr, page, limit, currentUserID)
			if rqlErr != nil {
				return nil, 0, common.BadRequest("RQL parse error: " + rqlErr.Error())
			}
			if len(rqlIssues) == 0 {
				return []response.IssueResponse{}, rqlTotal, nil
			}
			ids := make([]uint64, len(rqlIssues))
			for i, issue := range rqlIssues {
				ids[i] = issue.ID
			}
			var issues []model.Issue
			q := s.db.Where("id IN ?", ids).
				Preload("State").
				Preload("IssueType").
				Preload("AssigneeLinks.User").
				Preload("LabelLinks.Label").
				Preload("CycleLink")
			if sortClause != "" {
				q = q.Order(sortClause)
			} else {
				q = q.Order("sort_order ASC, sequence_id DESC")
			}
			if err := q.Find(&issues).Error; err != nil {
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
			return result, rqlTotal, nil
		}
	}

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
	defaultOrder := "issues.sort_order ASC, issues.sequence_id DESC"
	if sortClause != "" {
		defaultOrder = sortClause
	}

	if needDedup {
		// Use subquery to get distinct issue IDs, then fetch full rows (with sort applied in subquery)
		var ids []uint64
		subBase := baseQuery
		if sortClause != "" {
			subBase = subBase.Order(sortClause)
		}
		if err := subBase.Distinct("issues.id").Limit(limit).Offset(offset).Pluck("issues.id", &ids).Error; err != nil {
			return nil, 0, common.Internal("Database error")
		}
		if len(ids) > 0 {
			idOrder := "CASE issues.id "
			for i, id := range ids {
				idOrder += fmt.Sprintf("WHEN %d THEN %d ", id, i)
			}
			idOrder += "END"
			if err := s.db.Where("id IN ?", ids).
				Preload("State").
				Preload("IssueType").
				Preload("AssigneeLinks.User").
				Preload("LabelLinks.Label").
				Preload("CycleLink").
				Order(idOrder).
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
			Order(defaultOrder).
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

// ListByWorkspace returns issues across all projects in a workspace.
func (s *IssueService) ListByWorkspace(workspaceID uint64, filters map[string]interface{}, limit, offset int) ([]response.IssueResponse, int64, error) {
	var projectIDs []uint64
	if err := s.db.Model(&model.Project{}).Where("workspace_id = ?", workspaceID).Pluck("id", &projectIDs).Error; err != nil {
		return nil, 0, err
	}
	if len(projectIDs) == 0 {
		return []response.IssueResponse{}, 0, nil
	}
	return s.listByProjects(projectIDs, filters, limit, offset)
}

func (s *IssueService) listByProjects(projectIDs []uint64, filters map[string]interface{}, limit, offset int) ([]response.IssueResponse, int64, error) {
	baseQuery := s.db.Model(&model.Issue{}).Where("issues.project_id IN ?", projectIDs)

	sortClause := s.buildSortClause(filters)

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
	if cycleID, ok := filters["cycle_id"]; ok && cycleID != nil {
		baseQuery = baseQuery.Joins("JOIN issue_cycles ON issue_cycles.issue_id = issues.id").
			Where("issue_cycles.cycle_id = ?", cycleID)
	}
	if search, ok := filters["search"]; ok && search != nil {
		searchStr := fmt.Sprintf("%%%s%%", search)
		baseQuery = baseQuery.Where("issues.name ILIKE ? OR COALESCE(issues.description_stripped, '') ILIKE ?",
			searchStr, searchStr)
	}
	if issueTypeID, ok := filters["issue_type_id"]; ok && issueTypeID != nil {
		baseQuery = baseQuery.Where("issues.issue_type_id = ?", issueTypeID)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	defaultOrder := "issues.sort_order ASC, issues.sequence_id DESC"
	if sortClause != "" {
		defaultOrder = sortClause
	}

	var issues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("IssueType").
		Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").
		Preload("CycleLink").
		Preload("Parent").
		Preload("Project").
		Order(defaultOrder).
		Offset(offset).
		Limit(limit).
		Find(&issues).Error; err != nil {
		return nil, 0, err
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
	var oldStateID uint64 // Store old state ID for automation context

	if req.Name != nil && *req.Name != issue.Name {
		oldVal := issue.Name
		s.createActivity(tx, issueID, "updated", strPtr("name"), strPtr(oldVal), req.Name, nil, &userID)
		issue.Name = *req.Name
		hasChanges = true
	}
	if req.DescriptionHTML != nil && *req.DescriptionHTML != issue.DescriptionHTML {
		s.createActivity(tx, issueID, "updated", strPtr("description"), nil, nil, nil, &userID)
		issue.DescriptionHTML = sanitizeHTML(*req.DescriptionHTML)
		descStripped := stripHTMLTags(issue.DescriptionHTML)
		issue.DescriptionStripped = &descStripped
		hasChanges = true
	}
	if req.Priority != nil && *req.Priority != issue.Priority {
		oldVal := issue.Priority
		s.createActivity(tx, issueID, "updated", strPtr("priority"), strPtr(oldVal), req.Priority, nil, &userID)
		issue.Priority = *req.Priority
		hasChanges = true
	}
	if req.StateID != nil && *req.StateID != issue.StateID {
		oldStateID = issue.StateID
		newStateID := *req.StateID
		if err := s.validateStateTransition(tx, issue.ProjectID, issueID, oldStateID, newStateID, userID); err != nil {
			tx.Rollback()
			return nil, err
		}
		issue.StateID = newStateID

		var oldState, newState model.State
		s.db.First(&oldState, oldStateID)
		if err := s.db.First(&newState, newStateID).Error; err == nil {
			if newState.Group == common.StateGroupCompleted {
				now := time.Now()
				issue.CompletedAt = &now
			} else {
				issue.CompletedAt = nil
			}
		}

		s.createActivity(tx, issueID, "updated", strPtr("state"),
			strPtr(oldState.Name), strPtr(newState.Name), nil, &userID)
		hasChanges = true

		if s.notificationSvc != nil {

			recipientIDs := make(map[uint64]bool)

			var assignees []model.IssueAssignee
			s.db.Where("issue_id = ?", issueID).Find(&assignees)
			for _, a := range assignees {
				if a.UserID != userID {
					recipientIDs[a.UserID] = true
				}
			}

			var watchers []model.IssueWatcher
			s.db.Where("issue_id = ?", issueID).Find(&watchers)
			for _, w := range watchers {
				if w.UserID != userID {
					recipientIDs[w.UserID] = true
				}
			}

			if len(recipientIDs) > 0 {
				ids := make([]uint64, 0, len(recipientIDs))
				for id := range recipientIDs {
					ids = append(ids, id)
				}
				title := fmt.Sprintf("状态变更: %s", issue.Name)
				message := fmt.Sprintf("工作项 #%d 状态从 %s 变为 %s", issue.SequenceID, oldState.Name, newState.Name)
				issueIDPtr := issueID
				projectIDPtr := issue.ProjectID
				s.notificationSvc.TriggerNotificationsBulk(tx, "issue_state_changed", title, message, ids, &userID, &projectIDPtr, &issueIDPtr)
			}
		}
	}
	if req.SortOrder != nil && *req.SortOrder != issue.SortOrder {
		issue.SortOrder = *req.SortOrder
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
		s.createActivity(tx, issueID, "updated", strPtr("cycle"), nil, nil, nil, &userID)
		hasChanges = true
	}

	// Handle modules
	if req.ModuleIDs != nil {
		tx.Where("issue_id = ?", issueID).Delete(&model.ModuleIssue{})
		for _, mid := range req.ModuleIDs {
			tx.Create(&model.ModuleIssue{IssueID: issueID, ModuleID: mid})
		}
		s.createActivity(tx, issueID, "updated", strPtr("module"), nil, nil, nil, &userID)
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
		s.createActivity(tx, issueID, "updated", strPtr("issue_type"), nil, nil, nil, &userID)
		hasChanges = true
	}

	// Parse dates
	if req.StartDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartDate); err == nil {
			s.createActivity(tx, issueID, "updated", strPtr("start_date"), nil, nil, nil, &userID)
			issue.StartDate = &t
			tx.Save(&issue)
			hasChanges = true
		}
	}
	if req.TargetDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.TargetDate); err == nil {
			s.createActivity(tx, issueID, "updated", strPtr("target_date"), nil, nil, nil, &userID)
			issue.TargetDate = &t
			tx.Save(&issue)
			hasChanges = true
		}
	}

	if req.CoverImageURL != nil {
		issue.CoverImageURL = req.CoverImageURL
		tx.Save(&issue)
		hasChanges = true
	}

	// Handle parent
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			issue.ParentID = nil
		} else {
			if *req.ParentID == issueID {
				tx.Rollback()
				return nil, common.BadRequest("An issue cannot be its own parent")
			}
			if s.checkCircularReference(issueID, *req.ParentID) {
				tx.Rollback()
				return nil, common.BadRequest("Cannot create circular reference in issue hierarchy")
			}
			var parent model.Issue
			if err := s.db.First(&parent, *req.ParentID).Error; err != nil {
				tx.Rollback()
				return nil, common.NotFound("Parent issue not found")
			}
			issue.ParentID = req.ParentID
		}
		tx.Save(&issue)
		s.createActivity(tx, issueID, "updated", strPtr("parent"), nil, nil, nil, &userID)
		hasChanges = true
	}

	if !hasChanges {
		tx.Rollback()
		return s.buildResponse(issueID)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to commit update")
	}

	event := "issue_updated"
	if req.StateID != nil {
		event = "state_changed"
	}

	// Automation trigger: fire after commit for state changes
	if req.StateID != nil {
		// Fetch new state details for context
		var newState model.State
		s.db.First(&newState, *req.StateID)

		s.runAutomations(issueID, "state_changed", map[string]interface{}{
			"issue_id":    issueID,
			"old_state":   fmt.Sprintf("%d", oldStateID), // Use the saved old value
			"new_state":   fmt.Sprintf("%d", *req.StateID),
			"state_group": newState.Group,
			"project_id":  issue.ProjectID,
			"priority":    issue.Priority,
		})
	}

	// Automation trigger: fire after commit for issue updates (non-state changes)
	if hasChanges && req.StateID == nil {
		s.runAutomations(issueID, "issue_updated", map[string]interface{}{
			"issue_id":   issueID,
			"project_id": issue.ProjectID,
			"priority":   issue.Priority,
			"state_id":   issue.StateID,
		})
	}

	s.webhookSvc.Fire(issue.ProjectID, event, map[string]interface{}{"issue_id": issueID, "name": issue.Name, "priority": issue.Priority, "state_id": issue.StateID})

	var assignees []model.IssueAssignee
	s.db.Where("issue_id = ?", issueID).Find(&assignees)
	for _, a := range assignees {
		if a.UserID != userID {
			SSE.NotifyUser(a.UserID, "issue_updated", "工作项已更新", issue.Name)
		}
	}

	return s.buildResponse(issueID)
}

// Delete performs a soft delete on an issue.
func (s *IssueService) Delete(issueID, userID uint64) error {
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if err := s.checkProjectMembership(issue.ProjectID, userID); err != nil {
		return err
	}

	return s.db.Delete(&issue).Error
}

// Archive archives an issue.
func (s *IssueService) Archive(issueID, userID uint64) error {
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if err := s.checkProjectMembership(issue.ProjectID, userID); err != nil {
		return err
	}

	now := time.Now()
	issue.ArchivedAt = &now
	return s.db.Save(&issue).Error
}

// Restore restores an archived or deleted issue.
func (s *IssueService) Restore(issueID, userID uint64) (*response.IssueResponse, error) {
	var issue model.Issue
	if err := s.db.Unscoped().First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	if err := s.checkProjectMembership(issue.ProjectID, userID); err != nil {
		return nil, err
	}

	issue.ArchivedAt = nil
	issue.DeletedAt = gorm.DeletedAt{}
	s.db.Unscoped().Save(&issue)

	return s.buildResponse(issueID)
}

func (s *IssueService) checkProjectMembership(projectID, userID uint64) error {
	var count int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Count(&count)
	if count == 0 {
		return common.Forbidden("You must be a member of the project to perform this action")
	}
	return nil
}

func (s *IssueService) checkCircularReference(issueID, parentID uint64) bool {
	visited := make(map[uint64]bool)
	currentID := parentID
	for {
		if currentID == 0 {
			return false
		}
		if currentID == issueID {
			return true
		}
		if visited[currentID] {
			return false
		}
		visited[currentID] = true
		var issue model.Issue
		if err := s.db.Select("parent_id").First(&issue, currentID).Error; err != nil {
			return false
		}
		if issue.ParentID == nil {
			return false
		}
		currentID = *issue.ParentID
	}
}

// AddAssignee adds an assignee to an issue.
func (s *IssueService) AddAssignee(issueID, userID, actorID uint64) error {
	// Verify issue exists
	var issue model.Issue
	if err := s.db.Preload("Project").First(&issue, issueID).Error; err != nil {
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

	// Trigger notification for assignment
	if s.notificationSvc != nil && userID != actorID {
		title := fmt.Sprintf("工作项分配: %s", issue.Name)
		message := fmt.Sprintf("您被分配到工作项 #%d: %s", issue.SequenceID, issue.Name)
		issueIDPtr := issueID
		projectIDPtr := issue.ProjectID
		s.notificationSvc.TriggerNotification(s.db, "issue_assigned", title, message, userID, &actorID, &projectIDPtr, &issueIDPtr)
	}

	// Automation trigger: assignee_changed
	s.runAutomations(issueID, "assignee_changed", map[string]interface{}{
		"issue_id": issueID, "project_id": issue.ProjectID,
		"new_assignee": userID, "actor_id": actorID,
	})

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

	// Validate cycle exists
	var cycle model.Cycle
	if err := s.db.First(&cycle, cycleID).Error; err != nil {
		return common.NotFound("Cycle not found")
	}

	tx := s.db.Begin()
	// Remove existing cycle
	tx.Where("issue_id = ?", issueID).Delete(&model.IssueCycle{})
	// Set new cycle
	if err := tx.Create(&model.IssueCycle{IssueID: issueID, CycleID: cycleID}).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to set cycle")
	}

	s.createActivity(tx, issueID, "updated", strPtr("cycle"), nil, nil, nil, &actorID)
	if err := tx.Commit().Error; err != nil {
		return common.Internal("Failed to commit transaction")
	}
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
	if err := s.db.Preload("Actor").
		Where("issue_id = ?", issueID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&activities).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.IssueActivityResponse, len(activities))
	for i, a := range activities {
		actorName := ""
		actorAvatar := ""
		if a.Actor != nil {
			actorName = a.Actor.DisplayName
			if a.Actor.Avatar != nil {
				actorAvatar = *a.Actor.Avatar
			}
		}
		result[i] = response.IssueActivityResponse{
			ID:               a.ID,
			IssueID:          a.IssueID,
			Verb:             a.Verb,
			Field:            a.Field,
			OldValue:         a.OldValue,
			NewValue:         a.NewValue,
			Comment:          a.Comment,
			ActorID:          a.ActorID,
			ActorDisplayName: actorName,
			ActorAvatar:      actorAvatar,
			CreatedAt:        a.CreatedAt,
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

// Suggest returns search suggestions for issues based on query prefix.
func (s *IssueService) Suggest(projectID uint64, query string, limit int) ([]response.IssueSearchResult, error) {
	if query == "" {
		return []response.IssueSearchResult{}, nil
	}

	q := s.db.Model(&model.Issue{}).
		Preload("Project").
		Where("issues.project_id = ? AND issues.archived_at IS NULL", projectID).
		Where("issues.name ILIKE ? OR issues.sequence_id::text LIKE ?",
			"%"+query+"%", "%"+query+"%").
		Order("issues.updated_at DESC").
		Limit(limit)

	var issues []model.Issue
	if err := q.Find(&issues).Error; err != nil {
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
func (s *IssueService) BulkUpdate(projectID uint64, req *request.BulkUpdateRequest, userID uint64) (*response.BulkUpdateResultResponse, error) {
	if len(req.IssueIDs) == 0 {
		return &response.BulkUpdateResultResponse{}, nil
	}

	tx := s.db.Begin()

	var failedItems []response.BulkFailedItem
	var successIDs []uint64

	hasSimpleUpdates := req.Priority != nil || req.StartDate != nil || req.TargetDate != nil || req.SortOrder != nil

	if hasSimpleUpdates {
		updateMap := make(map[string]interface{})
		if req.Priority != nil {
			updateMap["priority"] = *req.Priority
		}
		if req.StartDate != nil {
			if t, err := time.Parse(time.RFC3339, *req.StartDate); err == nil {
				updateMap["start_date"] = t
			}
		}
		if req.TargetDate != nil {
			if t, err := time.Parse(time.RFC3339, *req.TargetDate); err == nil {
				updateMap["target_date"] = t
			}
		}
		if req.SortOrder != nil {
			updateMap["sort_order"] = *req.SortOrder
		}
		if len(updateMap) > 0 {
			if err := tx.Model(&model.Issue{}).Where("id IN ?", req.IssueIDs).Updates(updateMap).Error; err != nil {
				tx.Rollback()
				return nil, common.Internal("Failed to bulk update issues")
			}
		}
	}

	if req.StateID != nil {
		for _, issueID := range req.IssueIDs {
			var issue model.Issue
			if err := tx.First(&issue, issueID).Error; err != nil {
				failedItems = append(failedItems, response.BulkFailedItem{IssueID: issueID, Reason: "Issue not found"})
				continue
			}
			oldStateID := issue.StateID
			newStateID := *req.StateID
			if oldStateID == newStateID {
				successIDs = append(successIDs, issueID)
				continue
			}
			if err := s.validateStateTransition(tx, projectID, issueID, oldStateID, newStateID, userID); err != nil {
				failedItems = append(failedItems, response.BulkFailedItem{IssueID: issueID, Reason: err.Error()})
				continue
			}
			issue.StateID = newStateID
			var newState model.State
			if err := tx.First(&newState, newStateID).Error; err == nil {
				if newState.Group == common.StateGroupCompleted {
					now := time.Now()
					issue.CompletedAt = &now
				} else {
					issue.CompletedAt = nil
				}
			}
			tx.Save(&issue)
			successIDs = append(successIDs, issueID)
		}
	} else {
		successIDs = req.IssueIDs
	}

	if req.AssigneeIDs != nil {
		tx.Where("issue_id IN ?", req.IssueIDs).Delete(&model.IssueAssignee{})
		for _, issueID := range req.IssueIDs {
			for _, aid := range req.AssigneeIDs {
				tx.Create(&model.IssueAssignee{IssueID: issueID, UserID: aid})
			}
		}
	}

	if req.LabelIDs != nil {
		tx.Where("issue_id IN ?", req.IssueIDs).Delete(&model.IssueLabel{})
		for _, issueID := range req.IssueIDs {
			for _, lid := range req.LabelIDs {
				tx.Create(&model.IssueLabel{IssueID: issueID, LabelID: lid})
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit bulk update")
	}

	if req.StateID != nil {
		for _, issueID := range successIDs {
			s.runAutomations(issueID, "state_changed", map[string]interface{}{
				"issue_id":   issueID,
				"new_state":  fmt.Sprintf("%d", *req.StateID),
				"project_id": projectID,
			})
		}
	} else if req.Priority != nil || req.AssigneeIDs != nil || req.LabelIDs != nil {
		for _, issueID := range successIDs {
			s.runAutomations(issueID, "issue_updated", map[string]interface{}{
				"issue_id":   issueID,
				"project_id": projectID,
			})
		}
	}

	var issues []model.Issue
	if err := s.db.Preload("State").Preload("IssueType").Preload("AssigneeLinks.User").Preload("LabelLinks.Label").Preload("CycleLink").Where("id IN ?", successIDs).Find(&issues).Error; err != nil {
		return nil, common.Internal("Failed to fetch updated issues")
	}

	updatedItems := make([]response.IssueResponse, 0, len(issues))
	for _, issue := range issues {
		resp, err := s.BuildIssueResponse(&issue)
		if err != nil {
			continue
		}
		updatedItems = append(updatedItems, *resp)
	}

	return &response.BulkUpdateResultResponse{
		SuccessCount: len(successIDs),
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
		UpdatedItems: updatedItems,
	}, nil
}

// BulkDelete deletes multiple issues.
func (s *IssueService) BulkDelete(issueIDs []uint64, userID uint64) (*response.BulkDeleteResultResponse, error) {
	if len(issueIDs) == 0 {
		return &response.BulkDeleteResultResponse{}, nil
	}

	var projectIDs []uint64
	s.db.Model(&model.Issue{}).Where("id IN ?", issueIDs).Pluck("DISTINCT project_id", &projectIDs)
	for _, pid := range projectIDs {
		if err := s.checkProjectMembership(pid, userID); err != nil {
			return nil, err
		}
	}

	var failedItems []response.BulkFailedItem
	var successIDs []uint64
	tx := s.db.Begin()

	for _, issueID := range issueIDs {
		var issue model.Issue
		if err := tx.First(&issue, issueID).Error; err != nil {
			failedItems = append(failedItems, response.BulkFailedItem{IssueID: issueID, Reason: "Issue not found"})
			continue
		}
		if err := tx.Delete(&issue).Error; err != nil {
			failedItems = append(failedItems, response.BulkFailedItem{IssueID: issueID, Reason: "Failed to delete issue"})
			continue
		}
		successIDs = append(successIDs, issueID)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit bulk delete")
	}

	return &response.BulkDeleteResultResponse{
		SuccessCount: len(successIDs),
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
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
		Preload("ModuleLinks").
		Preload("Parent.State").
		Preload("Parent.IssueType").
		Preload("Parent.AssigneeLinks.User").
		Preload("SubIssues.State").
		Preload("SubIssues.IssueType").
		Preload("SubIssues.AssigneeLinks.User").
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
		CoverImageURL:   issue.CoverImageURL,
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

	// Parent issue — use preloaded data, or load manually if preload failed
	if issue.ParentID != nil {
		parentIssue := issue.Parent
		if parentIssue == nil || parentIssue.ID == 0 {
			var p model.Issue
			if err := s.db.Preload("State").Preload("IssueType").Preload("AssigneeLinks.User").First(&p, *issue.ParentID).Error; err == nil {
				parentIssue = &p
			}
		}
		if parentIssue != nil && parentIssue.ID != 0 {
			parent := &response.RelatedIssueLite{
				ID:         parentIssue.ID,
				SequenceID: parentIssue.SequenceID,
				Name:       parentIssue.Name,
				StateID:    parentIssue.StateID,
				Priority:   parentIssue.Priority,
				TargetDate: parentIssue.TargetDate,
				Assignees:  []response.UserLite{},
			}
			if parentIssue.State.ID != 0 {
				parent.StateName = parentIssue.State.Name
				parent.StateGroup = parentIssue.State.Group
			}
			if parentIssue.IssueType.ID != 0 {
				parent.IssueType = &response.IssueTypeLite{
					ID:    parentIssue.IssueType.ID,
					Name:  parentIssue.IssueType.Name,
					Color: parentIssue.IssueType.Color,
					Icon:  parentIssue.IssueType.Icon,
				}
			}
			for _, link := range parentIssue.AssigneeLinks {
				if link.User.ID != 0 {
					parent.Assignees = append(parent.Assignees, response.UserLite{
						ID:          link.User.ID,
						DisplayName: link.User.DisplayName,
						Email:       link.User.Email,
					})
				}
			}
			resp.Parent = parent
		}
	}

	// Sub-issues — use preloaded data, or load manually if preload failed
	subIssues := issue.SubIssues
	if len(subIssues) == 0 {
		s.db.Where("parent_id = ?", issue.ID).
			Preload("State").Preload("IssueType").Preload("AssigneeLinks.User").
			Find(&subIssues)
	}
	resp.SubIssues = make([]response.SubIssueLite, 0)
	for _, sub := range subIssues {
		si := response.SubIssueLite{
			ID:         sub.ID,
			SequenceID: sub.SequenceID,
			Name:       sub.Name,
			StateID:    sub.StateID,
			Priority:   sub.Priority,
			TargetDate: sub.TargetDate,
			Assignees:  []response.UserLite{},
		}
		if sub.State.ID != 0 {
			si.StateName = sub.State.Name
			si.StateGroup = sub.State.Group
		}
		if sub.IssueType.ID != 0 {
			si.IssueType = &response.IssueTypeLite{
				ID:    sub.IssueType.ID,
				Name:  sub.IssueType.Name,
				Color: sub.IssueType.Color,
				Icon:  sub.IssueType.Icon,
			}
		}
		for _, link := range sub.AssigneeLinks {
			if link.User.ID != 0 {
				si.Assignees = append(si.Assignees, response.UserLite{
					ID:          link.User.ID,
					DisplayName: link.User.DisplayName,
					Email:       link.User.Email,
				})
			}
		}
		resp.SubIssues = append(resp.SubIssues, si)
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

	// Modules
	resp.ModuleIDs = make([]uint64, 0)
	for _, link := range issue.ModuleLinks {
		if link.ModuleID != 0 {
			resp.ModuleIDs = append(resp.ModuleIDs, link.ModuleID)
		}
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
// Uses the provided db (should be a transaction when called within one) for consistency.
func (s *IssueService) validateStateTransition(db *gorm.DB, projectID, issueID, oldStateID, newStateID, userID uint64) error {
	// Get issue type ID if available for workflow filtering
	var issueTypeID *uint64
	var issue model.Issue
	if err := db.Select("issue_type_id").First(&issue, issueID).Error; err == nil && issue.IssueTypeID != nil {
		issueTypeID = issue.IssueTypeID
	}

	// Get workspace ID from project for workspace-level workflow lookup
	var workspaceID uint64
	db.Raw("SELECT workspace_id FROM projects WHERE id = ?", projectID).Scan(&workspaceID)

	// Build the workflow query condition
	var workflows []model.Workflow
	query := db.Where("is_active = ?", true)

	// Include both project-level workflows (project_id = ?) AND workspace-level workflows (workspace_id = ? AND project_id IS NULL)
	if issueTypeID != nil {
		// Include workflows bound to this issue type OR workflows with no issue type binding
		query = query.Where("(project_id = ? AND (issue_type_id = ? OR issue_type_id IS NULL)) OR (workspace_id = ? AND project_id IS NULL AND (issue_type_id = ? OR issue_type_id IS NULL))",
			projectID, *issueTypeID, workspaceID, *issueTypeID)
	} else {
		// Only include workflows with no issue type binding
		query = query.Where("(project_id = ? AND issue_type_id IS NULL) OR (workspace_id = ? AND project_id IS NULL AND issue_type_id IS NULL)",
			projectID, workspaceID)
	}
	query.Find(&workflows)

	if len(workflows) == 0 {
		return nil // no workflows configured = allow all transitions
	}
	for _, wf := range workflows {
		var transition model.StateTransition
		err := db.Where("workflow_id = ? AND source_state_id = ? AND target_state_id = ?",
			wf.ID, oldStateID, newStateID).First(&transition).Error
		if err != nil {
			continue // not found in this workflow, try next
		}
		// Transition found — check rule_type
		if transition.RuleType == "allow" {
			return nil // simple allow, no restriction
		}
		if transition.RuleType == "approval" {
			// Check if the user is an authorized approver
			if transition.ApproverIDs != nil && *transition.ApproverIDs != "" {
				allowedIDs := strings.Split(*transition.ApproverIDs, ",")
				uidStr := fmt.Sprintf("%d", userID)
				for _, id := range allowedIDs {
					if strings.TrimSpace(id) == uidStr {
						return nil // user is an approved approver
					}
				}
			}
			// Check role-based approval using actual RBAC roles
			if transition.RoleAllowed != "" {
				var userRoleLevel int
				// Check workspace member role
				var member struct {
					Role int
				}
				if err := db.Raw("SELECT role FROM workspace_members WHERE user_id = ? AND workspace_id = (SELECT workspace_id FROM projects WHERE id = ?) LIMIT 1",
					userID, projectID).Scan(&member).Error; err == nil {
					userRoleLevel = member.Role
				}
				// Check project member role (may override workspace role)
				var prjMember struct {
					Role int
				}
				if err := db.Raw("SELECT role FROM project_members WHERE user_id = ? AND project_id = ? LIMIT 1",
					userID, projectID).Scan(&prjMember).Error; err == nil && prjMember.Role > userRoleLevel {
					userRoleLevel = prjMember.Role
				}
				// Get the role level for the allowed role name
				var allowedRole struct {
					Level int
				}
				if err := db.Raw("SELECT level FROM roles WHERE name = ? AND workspace_id IS NULL LIMIT 1",
					transition.RoleAllowed).Scan(&allowedRole).Error; err == nil {
					if userRoleLevel >= allowedRole.Level {
						return nil
					}
				}
			}
			return common.BadRequest("Approval required: you are not authorized to approve this transition")
		}
		return nil // unknown rule_type, allow
	}
	var oldSt, newSt model.State
	db.First(&oldSt, oldStateID)
	db.First(&newSt, newStateID)
	return common.BadRequest(fmt.Sprintf(
		"Workflow rejected: transition from '%s' to '%s' is not allowed",
		oldSt.Name, newSt.Name))
}

// runAutomations executes automation rules for a given trigger type on an issue.
// Called after transaction commit — uses the automation service's event bus.
func (s *IssueService) runAutomations(issueID uint64, triggerType string, context map[string]interface{}) {
	if s.automationSvc == nil {
		return
	}

	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		log.Printf("[IssueService] Failed to fetch issue %d for automation: %v", issueID, err)
		return
	}

	// 创建事件并发布到事件总线（异步执行）
	event := Event{
		Type:      triggerType,
		ProjectID: issue.ProjectID,
		IssueID:   issueID,
		Context:   context,
		Timestamp: time.Now(),
	}

	if err := s.automationSvc.PublishEvent(event); err != nil {
		log.Printf("[IssueService] Failed to publish automation event: %v", err)
	}
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

func stripBOM(r io.Reader) io.Reader {
	return &bomStripper{Reader: r, done: false}
}

type bomStripper struct {
	io.Reader
	done bool
}

func (bs *bomStripper) Read(p []byte) (int, error) {
	if !bs.done {
		bs.done = true
		buf := make([]byte, len(p)+3)
		n, err := bs.Reader.Read(buf)
		if err != nil {
			return n, err
		}
		if n >= 3 && buf[0] == 0xef && buf[1] == 0xbb && buf[2] == 0xbf {
			copy(p, buf[3:n])
			return n - 3, nil
		}
		copy(p, buf[:n])
		return n, nil
	}
	return bs.Reader.Read(p)
}

// ImportFromCSV imports issues from CSV content.
func (s *IssueService) ImportFromCSV(projectID, workspaceID, userID uint64, csvContent io.Reader) (*response.ImportResult, error) {
	reader := csv.NewReader(stripBOM(csvContent))
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
		} else if idx, ok := headerIdx["summary"]; ok && idx < len(record) {
			item.Name = strings.TrimSpace(record[idx]) // Jira
		} else if idx, ok := headerIdx["title"]; ok && idx < len(record) {
			item.Name = strings.TrimSpace(record[idx]) // Linear
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
		} else if idx, ok := headerIdx["status"]; ok && idx < len(record) {
			item.StateName = strings.TrimSpace(record[idx]) // Jira/Linear
		}

		if idx, ok := headerIdx["type"]; ok && idx < len(record) {
			item.TypeName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["类型"]; ok && idx < len(record) {
			item.TypeName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["issue type"]; ok && idx < len(record) {
			item.TypeName = strings.TrimSpace(record[idx]) // Jira
		}

		if idx, ok := headerIdx["assignees"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",")
		} else if idx, ok := headerIdx["负责人"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",")
		} else if idx, ok := headerIdx["assignee"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",") // Jira/Linear
		} else if idx, ok := headerIdx["reporter"]; ok && idx < len(record) {
			item.AssigneeEmails = splitAndTrim(record[idx], ",") // Jira fallback
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
		} else if idx, ok := headerIdx["created"]; ok && idx < len(record) {
			item.StartDate = strings.TrimSpace(record[idx]) // Jira
		} else if idx, ok := headerIdx["created at"]; ok && idx < len(record) {
			item.StartDate = strings.TrimSpace(record[idx]) // Linear
		}

		if idx, ok := headerIdx["target_date"]; ok && idx < len(record) {
			item.TargetDate = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["截止日期"]; ok && idx < len(record) {
			item.TargetDate = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["due date"]; ok && idx < len(record) {
			item.TargetDate = strings.TrimSpace(record[idx]) // Jira/Linear
		}

		if idx, ok := headerIdx["parent_title"]; ok && idx < len(record) {
			item.ParentTitle = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["父标题"]; ok && idx < len(record) {
			item.ParentTitle = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["parent"]; ok && idx < len(record) {
			item.ParentTitle = strings.TrimSpace(record[idx]) // Jira/Linear
		}

		if idx, ok := headerIdx["module"]; ok && idx < len(record) {
			item.ModuleName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["模块"]; ok && idx < len(record) {
			item.ModuleName = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["cycle"]; ok && idx < len(record) {
			item.CycleName = strings.TrimSpace(record[idx])
		} else if idx, ok := headerIdx["周期"]; ok && idx < len(record) {
			item.CycleName = strings.TrimSpace(record[idx])
		}

		if idx, ok := headerIdx["estimate"]; ok && idx < len(record) {
			if v, err := strconv.ParseFloat(strings.TrimSpace(record[idx]), 64); err == nil {
				item.EstimatePoints = &v
			}
		} else if idx, ok := headerIdx["估点"]; ok && idx < len(record) {
			if v, err := strconv.ParseFloat(strings.TrimSpace(record[idx]), 64); err == nil {
				item.EstimatePoints = &v
			}
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

// ==================== Flow Metrics & Export ====================

// FlowMetrics contains aggregated issue counts by state.
type FlowMetrics struct {
	StateCounts    map[string]int `json:"state_counts"`
	PriorityCounts map[string]int `json:"priority_counts"`
	TypeCounts     map[string]int `json:"type_counts"`
	Total          int            `json:"total"`
}

// GetFlowMetrics returns flow metrics for a project.
func (s *IssueService) GetFlowMetrics(projectID uint64) (*FlowMetrics, error) {
	metrics := &FlowMetrics{
		StateCounts:    make(map[string]int),
		PriorityCounts: make(map[string]int),
		TypeCounts:     make(map[string]int),
	}

	var issues []model.Issue
	if err := s.db.Where("project_id = ?", projectID).Preload("State").Preload("IssueType").Find(&issues).Error; err != nil {
		return nil, err
	}

	metrics.Total = len(issues)
	for _, issue := range issues {
		stateName := "未知"
		if issue.State.Name != "" {
			stateName = issue.State.Name
		}
		metrics.StateCounts[stateName]++

		priority := "none"
		if issue.Priority != "" {
			priority = issue.Priority
		}
		metrics.PriorityCounts[priority]++

		typeName := "未知"
		if issue.IssueType.Name != "" {
			typeName = issue.IssueType.Name
		}
		metrics.TypeCounts[typeName]++
	}

	return metrics, nil
}

// ExportIssueItem is a simplified issue struct for CSV/JSON export.
type ExportIssueItem struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Priority      string   `json:"priority"`
	StateName     string   `json:"state_name"`
	IssueTypeName string   `json:"issue_type_name"`
	AssigneeNames []string `json:"assignee_names"`
	LabelNames    []string `json:"label_names"`
	StartDate     string   `json:"start_date"`
	TargetDate    string   `json:"target_date"`
	ParentName    string   `json:"parent_name"`
}

// ExportIssues returns all issues for a project in export-friendly format.
func (s *IssueService) ExportIssues(projectID uint64) ([]ExportIssueItem, error) {
	var issues []model.Issue
	if err := s.db.Where("project_id = ?", projectID).
		Preload("State").Preload("IssueType").Preload("AssigneeLinks.User").
		Preload("LabelLinks.Label").Preload("Parent").
		Order("issues.sequence_id ASC").
		Find(&issues).Error; err != nil {
		return nil, err
	}

	result := make([]ExportIssueItem, 0, len(issues))
	for _, issue := range issues {
		item := ExportIssueItem{
			Name:       issue.Name,
			Priority:   issue.Priority,
			StartDate:  "",
			TargetDate: "",
		}
		if issue.DescriptionStripped != nil {
			item.Description = *issue.DescriptionStripped
		}
		if issue.StartDate != nil {
			item.StartDate = issue.StartDate.Format("2006-01-02")
		}
		if issue.TargetDate != nil {
			item.TargetDate = issue.TargetDate.Format("2006-01-02")
		}
		if issue.State.Name != "" {
			item.StateName = issue.State.Name
		}
		if issue.IssueType.Name != "" {
			item.IssueTypeName = issue.IssueType.Name
		}
		for _, link := range issue.AssigneeLinks {
			item.AssigneeNames = append(item.AssigneeNames, link.User.DisplayName)
		}
		for _, link := range issue.LabelLinks {
			item.LabelNames = append(item.LabelNames, link.Label.Name)
		}
		if issue.Parent != nil {
			item.ParentName = issue.Parent.Name
		}
		result = append(result, item)
	}
	return result, nil
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

func (s *IssueService) ReorderSubIssues(parentID uint64, issueIDs []uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return common.Internal("Failed to begin transaction")
	}

	for i, issueID := range issueIDs {
		sortOrder := float64(i)
		if err := tx.Model(&model.Issue{}).
			Where("id = ? AND parent_id = ?", issueID, parentID).
			Update("sort_order", sortOrder).Error; err != nil {
			tx.Rollback()
			return common.Internal("Failed to update sort order")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return common.Internal("Failed to commit transaction")
	}

	return nil
}

// BulkCopy copies issues to another project, creating new copies while preserving the originals.
func (s *IssueService) BulkCopy(req *request.BulkCopyRequest, userID uint64) (*response.BulkCopyMoveResultResponse, error) {
	var targetProject model.Project
	if err := s.db.First(&targetProject, req.TargetProjectID).Error; err != nil {
		return nil, common.NotFound("Target project not found")
	}

	var issues []model.Issue
	if err := s.db.Preload("AssigneeLinks").Preload("LabelLinks").Preload("IssueType").
		Where("id IN ?", req.IssueIDs).Find(&issues).Error; err != nil {
		return nil, common.Internal("Failed to fetch issues")
	}

	if len(issues) == 0 {
		return nil, common.BadRequest("No issues found")
	}

	var failedItems []response.BulkFailedItem
	var results []response.IssueResponse
	tx := s.db.Begin()

	for _, issue := range issues {
		copiedIssue := &model.Issue{
			Name:            issue.Name,
			DescriptionHTML: issue.DescriptionHTML,
			DescriptionJSON: issue.DescriptionJSON,
			Priority:        issue.Priority,
			SequenceID:      issue.SequenceID,
			SortOrder:       issue.SortOrder,
			IsDraft:         issue.IsDraft,
			ProjectID:       req.TargetProjectID,
			WorkspaceID:     targetProject.WorkspaceID,
			StateID:         issue.StateID,
			Depth:           issue.Depth,
			ExternalID:      issue.ExternalID,
			ExternalSource:  issue.ExternalSource,
			IssueTypeID:     issue.IssueTypeID,
		}

		if issue.StartDate != nil {
			copiedIssue.StartDate = issue.StartDate
		}
		if issue.TargetDate != nil {
			copiedIssue.TargetDate = issue.TargetDate
		}

		if err := tx.Create(copiedIssue).Error; err != nil {
			failedItems = append(failedItems, response.BulkFailedItem{IssueID: issue.ID, Reason: "Failed to copy issue"})
			continue
		}

		for _, assignee := range issue.AssigneeLinks {
			tx.Create(&model.IssueAssignee{IssueID: copiedIssue.ID, UserID: assignee.UserID})
		}

		for _, label := range issue.LabelLinks {
			tx.Create(&model.IssueLabel{IssueID: copiedIssue.ID, LabelID: label.LabelID})
		}

		if req.IncludeSubtasks {
			s.copySubtasks(tx, issue.ID, copiedIssue.ID, req.TargetProjectID, targetProject.WorkspaceID)
		}

		resp, _ := s.buildResponse(copiedIssue.ID)
		if resp != nil {
			results = append(results, *resp)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return &response.BulkCopyMoveResultResponse{
		SuccessCount: len(results),
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
		Results:      results,
	}, nil
}

// copySubtasks recursively copies subtasks to the target issue.
func (s *IssueService) copySubtasks(tx *gorm.DB, sourceParentID, targetParentID, targetProjectID, workspaceID uint64) {
	var subtasks []model.Issue
	s.db.Preload("AssigneeLinks").Preload("LabelLinks").
		Where("parent_id = ?", sourceParentID).Find(&subtasks)

	for _, subtask := range subtasks {
		copiedSubtask := &model.Issue{
			Name:            subtask.Name,
			DescriptionHTML: subtask.DescriptionHTML,
			DescriptionJSON: subtask.DescriptionJSON,
			Priority:        subtask.Priority,
			SequenceID:      subtask.SequenceID,
			SortOrder:       subtask.SortOrder,
			IsDraft:         subtask.IsDraft,
			ProjectID:       targetProjectID,
			WorkspaceID:     workspaceID,
			StateID:         subtask.StateID,
			ParentID:        &targetParentID,
			Depth:           subtask.Depth,
			ExternalID:      subtask.ExternalID,
			ExternalSource:  subtask.ExternalSource,
			IssueTypeID:     subtask.IssueTypeID,
		}

		if subtask.StartDate != nil {
			copiedSubtask.StartDate = subtask.StartDate
		}
		if subtask.TargetDate != nil {
			copiedSubtask.TargetDate = subtask.TargetDate
		}

		tx.Create(copiedSubtask)

		for _, assignee := range subtask.AssigneeLinks {
			tx.Create(&model.IssueAssignee{IssueID: copiedSubtask.ID, UserID: assignee.UserID})
		}

		for _, label := range subtask.LabelLinks {
			tx.Create(&model.IssueLabel{IssueID: copiedSubtask.ID, LabelID: label.LabelID})
		}

		s.copySubtasks(tx, subtask.ID, copiedSubtask.ID, targetProjectID, workspaceID)
	}
}

// BulkMove moves issues to another project, removing them from the source project.
func (s *IssueService) BulkMove(req *request.BulkMoveRequest, userID uint64) (*response.BulkCopyMoveResultResponse, error) {
	var targetProject model.Project
	if err := s.db.First(&targetProject, req.TargetProjectID).Error; err != nil {
		return nil, common.NotFound("Target project not found")
	}

	var issues []model.Issue
	if err := s.db.Preload("AssigneeLinks").Preload("LabelLinks").
		Where("id IN ?", req.IssueIDs).Find(&issues).Error; err != nil {
		return nil, common.Internal("Failed to fetch issues")
	}

	if len(issues) == 0 {
		return nil, common.BadRequest("No issues found")
	}

	var failedItems []response.BulkFailedItem
	var results []response.IssueResponse
	tx := s.db.Begin()

	for _, issue := range issues {
		sourceProjectID := issue.ProjectID

		if err := tx.Model(&issue).Updates(map[string]interface{}{
			"project_id":    req.TargetProjectID,
			"workspace_id":  targetProject.WorkspaceID,
			"updated_by_id": userID,
		}).Error; err != nil {
			failedItems = append(failedItems, response.BulkFailedItem{IssueID: issue.ID, Reason: "Failed to move issue"})
			continue
		}

		if req.IncludeSubtasks {
			s.moveSubtasks(tx, issue.ID, req.TargetProjectID, targetProject.WorkspaceID)
		}

		resp, _ := s.buildResponse(issue.ID)
		if resp != nil {
			results = append(results, *resp)
		}

		sourceProjectIDStr := strconv.FormatUint(sourceProjectID, 10)
		targetProjectIDStr := strconv.FormatUint(req.TargetProjectID, 10)
		s.createActivity(tx, issue.ID, "moved", nil,
			&sourceProjectIDStr, &targetProjectIDStr, nil, &userID)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return &response.BulkCopyMoveResultResponse{
		SuccessCount: len(results),
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
		Results:      results,
	}, nil
}

// moveSubtasks recursively moves subtasks to the target project.
func (s *IssueService) moveSubtasks(tx *gorm.DB, parentID, targetProjectID, workspaceID uint64) {
	var subtasks []model.Issue
	s.db.Where("parent_id = ?", parentID).Find(&subtasks)

	for _, subtask := range subtasks {
		tx.Model(&subtask).Updates(map[string]interface{}{
			"project_id":   targetProjectID,
			"workspace_id": workspaceID,
		})

		s.moveSubtasks(tx, subtask.ID, targetProjectID, workspaceID)
	}
}

// ConvertType converts an issue's type to the target type.
func (s *IssueService) ConvertType(issueID uint64, req *request.ConvertTypeRequest, userID uint64) (*response.IssueResponse, error) {
	var issue model.Issue
	if err := s.db.Preload("IssueType").First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	var targetType model.IssueType
	if err := s.db.First(&targetType, req.TargetTypeID).Error; err != nil {
		return nil, common.NotFound("Target issue type not found")
	}

	if issue.IssueTypeID != nil && *issue.IssueTypeID == req.TargetTypeID {
		return nil, common.BadRequest("Issue already has this type")
	}

	oldTypeName := ""
	if issue.IssueType.ID != 0 {
		oldTypeName = issue.IssueType.Name
	}

	tx := s.db.Begin()

	if err := tx.Model(&issue).Updates(map[string]interface{}{
		"issue_type_id": req.TargetTypeID,
		"updated_by_id": userID,
	}).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to update issue type")
	}

	s.createActivity(tx, issue.ID, "converted", nil,
		&oldTypeName, &targetType.Name, nil, &userID)

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return s.buildResponse(issueID)
}

// ==================== Tree View ====================

// Tree returns paginated root-level issues for tree view, optionally with search.
func (s *IssueService) Tree(projectID uint64, filters map[string]interface{}, limit, offset int) ([]response.TreeIssueResponse, int64, error) {
	baseQuery := s.db.Model(&model.Issue{}).Where("issues.project_id = ?", projectID).Where("issues.parent_id IS NULL")

	// Apply simple filters (state, priority, search on root level)
	if stateID, ok := filters["state_id"]; ok && stateID != nil {
		baseQuery = baseQuery.Where("issues.state_id = ?", stateID)
	}
	if priority, ok := filters["priority"]; ok && priority != nil {
		baseQuery = baseQuery.Where("issues.priority = ?", priority)
	}
	if issueTypeID, ok := filters["issue_type_id"]; ok && issueTypeID != nil {
		baseQuery = baseQuery.Where("issues.issue_type_id = ?", issueTypeID)
	}

	searchTerm, hasSearch := filters["search"]
	if hasSearch && searchTerm != nil && searchTerm != "" {
		searchStr := fmt.Sprintf("%%%s%%", searchTerm)
		baseQuery = baseQuery.Where("issues.name ILIKE ? OR COALESCE(issues.description_stripped, '') ILIKE ?",
			searchStr, searchStr)
	}

	// Count total
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	// Paginate
	var issues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("IssueType").
		Order("issues.sort_order ASC, issues.sequence_id DESC").
		Limit(limit).Offset(offset).
		Find(&issues).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	// Collect issue IDs for batch sub-issues count
	issueIDs := make([]uint64, len(issues))
	for i, issue := range issues {
		issueIDs[i] = issue.ID
	}

	// Batch count sub-issues
	countMap := make(map[uint64]int64)
	if len(issueIDs) > 0 {
		type countResult struct {
			ParentID uint64
			Count    int64
		}
		var counts []countResult
		s.db.Model(&model.Issue{}).
			Select("parent_id, COUNT(*) as count").
			Where("parent_id IN ?", issueIDs).
			Group("parent_id").
			Scan(&counts)
		for _, c := range counts {
			countMap[c.ParentID] = c.Count
		}
	}

	result := make([]response.TreeIssueResponse, len(issues))
	for i, issue := range issues {
		resp := buildTreeIssueFromModel(&issue, countMap)
		result[i] = *resp
	}
	return result, total, nil
}

// TreeSearch finds issues matching the search term anywhere in the tree,
// and returns results with ancestor chains for tree view display.
func (s *IssueService) TreeSearch(projectID uint64, search string, limit, offset int) ([]response.TreeSearchResult, int64, error) {
	searchStr := fmt.Sprintf("%%%s%%", search)

	// Find all matching issues (not archived)
	baseQuery := s.db.Model(&model.Issue{}).
		Where("issues.project_id = ?", projectID).
		Where("issues.archived_at IS NULL").
		Where("issues.name ILIKE ? OR COALESCE(issues.description_stripped, '') ILIKE ?",
			searchStr, searchStr)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	var matchedIssues []model.Issue
	if err := baseQuery.
		Preload("State").
		Preload("IssueType").
		Order("issues.depth ASC, issues.sort_order ASC, issues.sequence_id DESC").
		Limit(limit).Offset(offset).
		Find(&matchedIssues).Error; err != nil {
		return nil, 0, common.Internal("Database error")
	}

	// For each matched issue, find its root ancestor and build ancestor chain
	results := make([]response.TreeSearchResult, 0, len(matchedIssues))

	for _, matched := range matchedIssues {
		// If already root, no ancestor chain needed
		if matched.ParentID == nil || *matched.ParentID == 0 {
			rootResp := buildTreeIssueFromModel(&matched, nil)
			rootResp.IsSearchMatch = true
			results = append(results, response.TreeSearchResult{
				RootIssue:     *rootResp,
				MatchedIssue:  *rootResp,
				AncestorChain: []response.AncestorInfo{},
			})
			continue
		}

		// Build ancestor chain by walking up parent_id
		ancestorChain := make([]response.AncestorInfo, 0)
		currentParentID := *matched.ParentID
		var rootIssue *model.Issue

		for {
			var parent model.Issue
			if err := s.db.Select("id, name, sequence_id, parent_id, depth, priority, state_id, sort_order").
				First(&parent, currentParentID).Error; err != nil {
				break
			}

			if parent.ParentID == nil || *parent.ParentID == 0 {
				// This is the root
				rootIssue = &parent
				break
			}

			// Insert at beginning to maintain root->child order
			ancestorChain = append([]response.AncestorInfo{{
				ID:         parent.ID,
				Name:       parent.Name,
				SequenceID: parent.SequenceID,
			}}, ancestorChain...)
			currentParentID = *parent.ParentID
		}

		if rootIssue == nil {
			// Fallback: use matched issue itself as root
			rootResp := buildTreeIssueFromModel(&matched, nil)
			rootResp.IsSearchMatch = true
			results = append(results, response.TreeSearchResult{
				RootIssue:     *rootResp,
				MatchedIssue:  *rootResp,
				AncestorChain: []response.AncestorInfo{},
			})
			continue
		}

		// Load full root issue details
		var fullRoot model.Issue
		if err := s.db.Preload("State").Preload("IssueType").First(&fullRoot, rootIssue.ID).Error; err != nil {
			continue
		}
		rootResp := buildTreeIssueFromModel(&fullRoot, nil)

		matchedResp := buildTreeIssueFromModel(&matched, nil)
		matchedResp.IsSearchMatch = true

		results = append(results, response.TreeSearchResult{
			RootIssue:     *rootResp,
			MatchedIssue:  *matchedResp,
			AncestorChain: ancestorChain,
		})
	}

	return results, total, nil
}

// Children returns direct children of a parent issue (lazy loading for tree expansion).
func (s *IssueService) Children(parentID uint64) ([]response.TreeIssueResponse, error) {
	var children []model.Issue
	if err := s.db.
		Where("parent_id = ?", parentID).
		Preload("State").
		Preload("IssueType").
		Order("sort_order ASC, sequence_id DESC").
		Find(&children).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	// Batch count sub-issues for each child
	childIDs := make([]uint64, len(children))
	for i, c := range children {
		childIDs[i] = c.ID
	}
	countMap := make(map[uint64]int64)
	if len(childIDs) > 0 {
		type countResult struct {
			ParentID uint64
			Count    int64
		}
		var counts []countResult
		s.db.Model(&model.Issue{}).
			Select("parent_id, COUNT(*) as count").
			Where("parent_id IN ?", childIDs).
			Group("parent_id").
			Scan(&counts)
		for _, c := range counts {
			countMap[c.ParentID] = c.Count
		}
	}

	result := make([]response.TreeIssueResponse, len(children))
	for i, child := range children {
		resp := buildTreeIssueFromModel(&child, countMap)
		result[i] = *resp
	}
	return result, nil
}

// buildTreeIssueFromModel converts a model.Issue to TreeIssueResponse with optional count map.
func buildTreeIssueFromModel(issue *model.Issue, countMap map[uint64]int64) *response.TreeIssueResponse {
	resp := &response.TreeIssueResponse{
		ID:          issue.ID,
		Name:        issue.Name,
		SequenceID:  issue.SequenceID,
		Priority:    issue.Priority,
		StateID:     issue.StateID,
		ParentID:    issue.ParentID,
		Depth:       issue.Depth,
		IssueTypeID: issue.IssueTypeID,
		StartDate:   issue.StartDate,
		TargetDate:  issue.TargetDate,
	}

	if issue.State.ID != 0 {
		resp.StateName = issue.State.Name
		resp.StateGroup = issue.State.Group
	}
	if issue.IssueType.ID != 0 {
		resp.IssueType = &response.IssueTypeLite{
			ID:    issue.IssueType.ID,
			Name:  issue.IssueType.Name,
			Color: issue.IssueType.Color,
			Icon:  issue.IssueType.Icon,
		}
	}

	if countMap != nil {
		if count, ok := countMap[issue.ID]; ok {
			resp.SubIssuesCount = count
			resp.HasChildren = count > 0
		}
	} else {
		resp.SubIssuesCount = int64(len(issue.SubIssues))
		resp.HasChildren = len(issue.SubIssues) > 0
	}

	return resp
}

// MergeDuplicates merges multiple duplicate issues into a target issue.
// Source issues will be deleted after merging.
func (s *IssueService) MergeDuplicates(req *request.MergeDuplicatesRequest, userID uint64) (*response.IssueResponse, error) {
	var targetIssue model.Issue
	if err := s.db.Preload("LabelLinks").Preload("AssigneeLinks").First(&targetIssue, req.TargetIssueID).Error; err != nil {
		return nil, common.NotFound("Target issue not found")
	}

	if len(req.SourceIssueIDs) == 0 {
		return nil, common.BadRequest("No source issues provided")
	}

	var sourceIssues []model.Issue
	if err := s.db.Preload("LabelLinks").Preload("AssigneeLinks").
		Where("id IN ?", req.SourceIssueIDs).Find(&sourceIssues).Error; err != nil {
		return nil, common.Internal("Failed to fetch source issues")
	}

	if len(sourceIssues) == 0 {
		return nil, common.BadRequest("No source issues found")
	}

	tx := s.db.Begin()

	if req.KeepSourceLabels {
		for _, source := range sourceIssues {
			for _, labelLink := range source.LabelLinks {
				var exists bool
				tx.Model(&targetIssue).Where("issue_id = ? AND label_id = ?", targetIssue.ID, labelLink.LabelID).
					First(&model.IssueLabel{}).Scan(&exists)
				if !exists {
					tx.Create(&model.IssueLabel{IssueID: targetIssue.ID, LabelID: labelLink.LabelID})
				}
			}
		}
	}

	if req.KeepSourceAssignees {
		for _, source := range sourceIssues {
			for _, assigneeLink := range source.AssigneeLinks {
				var exists bool
				tx.Model(&targetIssue).Where("issue_id = ? AND user_id = ?", targetIssue.ID, assigneeLink.UserID).
					First(&model.IssueAssignee{}).Scan(&exists)
				if !exists {
					tx.Create(&model.IssueAssignee{IssueID: targetIssue.ID, UserID: assigneeLink.UserID})
				}
			}
		}
	}

	for _, source := range sourceIssues {
		tx.Delete(&source)
	}

	sourceIDsStr := fmt.Sprintf("%v", req.SourceIssueIDs)
	s.createActivity(tx, targetIssue.ID, "merged", nil,
		&sourceIDsStr, nil, nil, &userID)

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return s.buildResponse(targetIssue.ID)
}

// ==================== Watcher ====================

// AddWatcher adds a user as a watcher for an issue.
func (s *IssueService) AddWatcher(issueID, userID uint64) error {
	var watcher model.IssueWatcher
	if err := s.db.Where("issue_id = ? AND user_id = ?", issueID, userID).First(&watcher).Error; err == nil {
		return common.BadRequest("User is already watching this issue")
	}

	return s.db.Create(&model.IssueWatcher{IssueID: issueID, UserID: userID}).Error
}

// RemoveWatcher removes a user from watching an issue.
func (s *IssueService) RemoveWatcher(issueID, userID uint64) error {
	result := s.db.Where("issue_id = ? AND user_id = ?", issueID, userID).Delete(&model.IssueWatcher{})
	if result.RowsAffected == 0 {
		return common.BadRequest("User is not watching this issue")
	}
	return result.Error
}

// ListWatchers returns watchers for an issue.
func (s *IssueService) ListWatchers(issueID uint64) ([]uint64, error) {
	var watchers []model.IssueWatcher
	if err := s.db.Where("issue_id = ?", issueID).Find(&watchers).Error; err != nil {
		return nil, common.Internal("Failed to fetch watchers")
	}
	userIDs := make([]uint64, len(watchers))
	for i, w := range watchers {
		userIDs[i] = w.UserID
	}
	return userIDs, nil
}
