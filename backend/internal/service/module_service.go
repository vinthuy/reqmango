package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ModuleService struct {
	db *gorm.DB
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{db: db}
}

func (s *ModuleService) buildResponse(m model.Module) *response.ModuleResponse {
	return &response.ModuleResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		ProjectID:   m.ProjectID,
		WorkspaceID: m.WorkspaceID,
		ParentID:    m.ParentID,
		Order:       m.Order,
		IsArchived:  m.IsArchived,
		ArchivedAt:  m.ArchivedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (s *ModuleService) countIssues(moduleID uint64) (total, completed int64, err error) {
	if err = s.db.Model(&model.ModuleIssue{}).Where("module_id = ?", moduleID).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group = ?", moduleID, common.StateGroupCompleted).
		Count(&completed).Error; err != nil {
		return total, 0, err
	}
	return total, completed, nil
}

// ==================== CRUD ====================

func (s *ModuleService) Create(workspaceID, userID uint64, req request.ModuleCreate) (*response.ModuleResponse, error) {
	var project model.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	module := model.Module{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		WorkspaceID: workspaceID,
		ParentID:    req.ParentID,
	}
	module.CreatedByID = &userID

	if err := s.db.Create(&module).Error; err != nil {
		return nil, common.Internal("Failed to create module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) List(projectID, workspaceID uint64, includeArchived bool) ([]response.ModuleResponse, int64, error) {
	query := s.db.Model(&model.Module{}).Where("project_id = ? AND workspace_id = ?", projectID, workspaceID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Failed to count modules")
	}

	var modules []model.Module
	if err := query.Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, 0, common.Internal("Failed to list modules")
	}

	result := make([]response.ModuleResponse, len(modules))
	for i, m := range modules {
		result[i] = *s.buildResponse(m)
	}

	return result, total, nil
}

func (s *ModuleService) Search(projectID, workspaceID uint64, query string) ([]response.ModuleResponse, error) {
	var modules []model.Module
	if err := s.db.Where("project_id = ? AND workspace_id = ? AND name ILIKE ?", projectID, workspaceID, "%"+query+"%").Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to search modules")
	}
	result := make([]response.ModuleResponse, len(modules))
	for i, m := range modules {
		result[i] = *s.buildResponse(m)
	}
	return result, nil
}

func (s *ModuleService) Get(moduleID uint64) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}
	return s.buildResponse(module), nil
}

func (s *ModuleService) Update(moduleID, userID uint64, req request.ModuleUpdate) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	if req.Name != nil {
		module.Name = *req.Name
	}
	if req.Description != nil {
		module.Description = *req.Description
	}
	if req.ParentID != nil {
		module.ParentID = req.ParentID
	}

	module.UpdatedByID = &userID
	if err := s.db.Save(&module).Error; err != nil {
		return nil, common.Internal("Failed to update module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) Delete(moduleID uint64) error {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return common.NotFound("Module not found")
	}
	s.db.Where("module_id = ?", moduleID).Delete(&model.ModuleIssue{})
	return s.db.Delete(&module).Error
}

// ==================== Issue Association ====================

func (s *ModuleService) AddIssue(moduleID, issueID uint64) error {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return common.NotFound("Module not found")
	}

	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return common.NotFound("Issue not found")
	}

	if issue.ProjectID != module.ProjectID {
		return common.BadRequest("Issue does not belong to this module's project")
	}

	var count int64
	s.db.Model(&model.ModuleIssue{}).Where("module_id = ? AND issue_id = ?", moduleID, issueID).Count(&count)
	if count > 0 {
		return common.BadRequest("Issue is already in this module")
	}

	if err := s.db.Create(&model.ModuleIssue{ModuleID: moduleID, IssueID: issueID}).Error; err != nil {
		return common.Internal("Failed to add issue to module")
	}

	return nil
}

func (s *ModuleService) RemoveIssue(moduleID, issueID uint64) error {
	result := s.db.Where("module_id = ? AND issue_id = ?", moduleID, issueID).Delete(&model.ModuleIssue{})
	if result.RowsAffected == 0 {
		return common.NotFound("Issue is not in this module")
	}
	return nil
}

func (s *ModuleService) ListIssues(moduleID uint64, stateID *uint64, priority string, limit, offset int) ([]response.IssueResponse, int64, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, 0, common.NotFound("Module not found")
	}

	baseQuery := s.db.Model(&model.Issue{}).
		Joins("JOIN module_issues ON module_issues.issue_id = issues.id").
		Where("module_issues.module_id = ?", moduleID)

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

// ==================== Progress & Statistics ====================

func (s *ModuleService) GetProgress(moduleID uint64) (*response.ModuleProgress, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	total, completed, _ := s.countIssues(moduleID)
	progress := 0
	if total > 0 {
		progress = int(float64(completed) / float64(total) * 100)
	}

	return &response.ModuleProgress{
		ModuleID:    moduleID,
		ModuleName:  module.Name,
		TotalIssues: total,
		Completed:   completed,
		Progress:    progress,
	}, nil
}

func (s *ModuleService) GetStatistics(moduleID uint64) (*response.ModuleStatistics, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	total, completed, _ := s.countIssues(moduleID)

	var active int64
	s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group IN ?", moduleID, []string{common.StateGroupBacklog, common.StateGroupUnstarted, common.StateGroupStarted}).
		Count(&active)

	var cancelled int64
	s.db.Model(&model.ModuleIssue{}).
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ? AND states.group = ?", moduleID, common.StateGroupCancelled).
		Count(&cancelled)

	type priorityRow struct {
		Priority string
		Count    int64
	}
	var pRows []priorityRow
	s.db.Table("module_issues").
		Select("issues.priority, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Where("module_issues.module_id = ?", moduleID).
		Group("issues.priority").
		Scan(&pRows)
	byPriority := make(map[string]int64)
	for _, r := range pRows {
		byPriority[r.Priority] = r.Count
	}

	type stateRow struct {
		Name  string
		Count int64
	}
	var sRows []stateRow
	s.db.Table("module_issues").
		Select("states.name, COUNT(*) as count").
		Joins("JOIN issues ON issues.id = module_issues.issue_id").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("module_issues.module_id = ?", moduleID).
		Group("states.name").
		Scan(&sRows)
	byState := make(map[string]int64)
	for _, r := range sRows {
		byState[r.Name] = r.Count
	}

	return &response.ModuleStatistics{
		ModuleID:     moduleID,
		ModuleName:   module.Name,
		TotalIssues:  total,
		ActiveIssues: active,
		Completed:    completed,
		Cancelled:    cancelled,
		ByPriority:   byPriority,
		ByState:      byState,
	}, nil
}

// ==================== Tree ====================

func (s *ModuleService) BuildTree(projectID uint64) ([]*response.ModuleTreeNode, error) {
	var modules []model.Module
	if err := s.db.Where("project_id = ? AND is_archived = ?", projectID, false).
		Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to load modules")
	}

	childrenMap := make(map[uint64][]*response.ModuleTreeNode)
	var roots []*response.ModuleTreeNode

	for _, m := range modules {
		total, completed, _ := s.countIssues(m.ID)
		progress := 0
		if total > 0 {
			progress = int(float64(completed) / float64(total) * 100)
		}

		node := &response.ModuleTreeNode{
			ModuleResponse:  *s.buildResponse(m),
			Children:        nil,
			TotalIssues:     total,
			CompletedIssues: completed,
			Progress:        progress,
		}

		if m.ParentID == nil {
			roots = append(roots, node)
		} else {
			childrenMap[*m.ParentID] = append(childrenMap[*m.ParentID], node)
		}
	}

	var attachChildren func(nodes []*response.ModuleTreeNode)
	attachChildren = func(nodes []*response.ModuleTreeNode) {
		for _, n := range nodes {
			n.Children = childrenMap[n.ID]
			if n.Children == nil {
				n.Children = []*response.ModuleTreeNode{}
			}
			attachChildren(n.Children)
		}
	}
	attachChildren(roots)

	if roots == nil {
		roots = []*response.ModuleTreeNode{}
	}
	return roots, nil
}
