package service

import (
	"fmt"

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

func (s *ModuleService) DB() *gorm.DB {
	return s.db
}

func (s *ModuleService) buildResponse(m model.Module, override *model.ModuleInheritanceOverride) *response.ModuleResponse {
	id := m.ID
	name := m.Name
	description := m.Description

	if override != nil && !override.IsExcluded {
		if override.OverrideName != nil {
			name = *override.OverrideName
		}
		if override.OverrideDescription != nil {
			description = *override.OverrideDescription
		}
	}

	return &response.ModuleResponse{
		ID:          &id,
		Name:        name,
		Description: description,
		ProjectID:   m.ProjectID,
		WorkspaceID: m.WorkspaceID,
		ParentID:    m.ParentID,
		Order:       m.Order,
		IsArchived:  m.IsArchived,
		ArchivedAt:  m.ArchivedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		IsInherited: m.ProjectID == nil,
		HasOverride: override != nil && !override.IsExcluded && (override.OverrideName != nil || override.OverrideDescription != nil),
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

func (s *ModuleService) getOverrides(projectID uint64) (map[uint64]*model.ModuleInheritanceOverride, error) {
	var overrides []model.ModuleInheritanceOverride
	if err := s.db.Where("project_id = ?", projectID).Find(&overrides).Error; err != nil {
		return nil, err
	}

	result := make(map[uint64]*model.ModuleInheritanceOverride)
	for i := range overrides {
		result[overrides[i].WorkspaceModuleID] = &overrides[i]
	}
	return result, nil
}

// ==================== CRUD ====================

func (s *ModuleService) Create(workspaceID, userID uint64, req request.ModuleCreate) (*response.ModuleResponse, error) {
	var project model.Project
	if err := s.db.First(&project, req.ProjectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	projectIDPtr := req.ProjectID
	module := model.Module{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   &projectIDPtr,
		WorkspaceID: workspaceID,
		ParentID:    req.ParentID,
	}
	module.CreatedByID = &userID

	if err := s.db.Create(&module).Error; err != nil {
		return nil, common.Internal("Failed to create module")
	}

	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) List(projectID, workspaceID uint64, includeArchived bool) ([]response.ModuleResponse, int64, error) {
	overrides, err := s.getOverrides(projectID)
	if err != nil {
		fmt.Printf("[ModuleService.List ERROR] getOverrides failed: projectID=%d err=%v\n", projectID, err)
		return nil, 0, common.Internal("Failed to load overrides: " + err.Error())
	}

	query := s.db.Model(&model.Module{}).Where("(project_id = ? OR (project_id IS NULL AND workspace_id = ?))", projectID, workspaceID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		fmt.Printf("[ModuleService.List ERROR] Count failed: projectID=%d workspaceID=%d err=%v\n", projectID, workspaceID, err)
		return nil, 0, common.Internal("Failed to count modules: " + err.Error())
	}

	var modules []model.Module
	if err := query.Order("\"order\", created_at").Find(&modules).Error; err != nil {
		fmt.Printf("[ModuleService.List ERROR] Find failed: projectID=%d workspaceID=%d err=%v\n", projectID, workspaceID, err)
		return nil, 0, common.Internal("Failed to list modules: " + err.Error())
	}

	result := make([]response.ModuleResponse, 0, len(modules))
	for _, m := range modules {
		if m.ProjectID == nil {
			if override, exists := overrides[m.ID]; exists {
				if override.IsExcluded {
					continue
				}
				result = append(result, *s.buildResponse(m, override))
			} else {
				result = append(result, *s.buildResponse(m, nil))
			}
		} else {
			result = append(result, *s.buildResponse(m, nil))
		}
	}

	return result, int64(len(result)), nil
}

func (s *ModuleService) Search(projectID, workspaceID uint64, query string) ([]response.ModuleResponse, error) {
	overrides, err := s.getOverrides(projectID)
	if err != nil {
		return nil, common.Internal("Failed to load overrides")
	}

	var modules []model.Module
	if err := s.db.Where("(project_id = ? OR (project_id IS NULL AND workspace_id = ?)) AND name ILIKE ?", projectID, workspaceID, "%"+query+"%").Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to search modules")
	}

	result := make([]response.ModuleResponse, 0, len(modules))
	for _, m := range modules {
		if m.ProjectID == nil {
			if override, exists := overrides[m.ID]; exists {
				if override.IsExcluded {
					continue
				}
				result = append(result, *s.buildResponse(m, override))
			} else {
				result = append(result, *s.buildResponse(m, nil))
			}
		} else {
			result = append(result, *s.buildResponse(m, nil))
		}
	}
	return result, nil
}

func (s *ModuleService) ListWorkspaceModules(workspaceID uint64) ([]response.ModuleResponse, error) {
	var modules []model.Module
	if err := s.db.Where("workspace_id = ? AND project_id IS NULL AND is_archived = ?", workspaceID, false).Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to list workspace modules")
	}
	result := make([]response.ModuleResponse, len(modules))
	for i, m := range modules {
		result[i] = *s.buildResponse(m, nil)
	}
	return result, nil
}

func (s *ModuleService) CreateWorkspaceModule(workspaceID, userID uint64, req request.ModuleCreate) (*response.ModuleResponse, error) {
	module := model.Module{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   nil,
		WorkspaceID: workspaceID,
		ParentID:    req.ParentID,
	}
	module.CreatedByID = &userID

	if err := s.db.Create(&module).Error; err != nil {
		return nil, common.Internal("Failed to create workspace module")
	}

	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) GetWorkspaceModule(workspaceID, moduleID uint64) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.Where("id = ? AND workspace_id = ? AND project_id IS NULL", moduleID, workspaceID).First(&module).Error; err != nil {
		return nil, common.NotFound("Workspace module not found")
	}
	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) UpdateWorkspaceModule(moduleID, userID uint64, req request.ModuleUpdate) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.Where("id = ? AND project_id IS NULL", moduleID).First(&module).Error; err != nil {
		return nil, common.NotFound("Workspace module not found")
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
		return nil, common.Internal("Failed to update workspace module")
	}

	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) DeleteWorkspaceModule(workspaceID, moduleID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Where("parent_id = ?", moduleID).Update("parent_id", nil).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to update child modules")
	}
	if err := tx.Where("module_id = ?", moduleID).Delete(&model.ModuleIssue{}).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to delete module issues")
	}
	if err := tx.Where("workspace_module_id = ?", moduleID).Delete(&model.ModuleInheritanceOverride{}).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to delete inheritance overrides")
	}
	result := tx.Where("id = ? AND workspace_id = ? AND project_id IS NULL", moduleID, workspaceID).Delete(&model.Module{})
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.NotFound("Workspace module not found")
	}
	return tx.Commit().Error
}

func (s *ModuleService) Get(moduleID uint64) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) GetWithProjectContext(moduleID, projectID uint64) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	if module.ProjectID != nil {
		return s.buildResponse(module, nil), nil
	}

	var override model.ModuleInheritanceOverride
	err := s.db.Where("project_id = ? AND workspace_module_id = ?", projectID, moduleID).First(&override).Error
	if err == nil && !override.IsExcluded {
		return s.buildResponse(module, &override), nil
	}
	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) Update(moduleID, userID uint64, req request.ModuleUpdate) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, common.NotFound("Module not found")
	}

	if module.ProjectID == nil {
		return nil, common.BadRequest("Cannot update workspace module through this endpoint")
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

	return s.buildResponse(module, nil), nil
}

func (s *ModuleService) Delete(moduleID uint64) error {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return common.NotFound("Module not found")
	}
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Where("parent_id = ?", moduleID).Update("parent_id", nil).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to update child modules")
	}
	if err := tx.Where("module_id = ?", moduleID).Delete(&model.ModuleIssue{}).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to delete module issues")
	}
	if err := tx.Delete(&module).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to delete module")
	}
	return tx.Commit().Error
}

// ==================== Inheritance Override ====================

func (s *ModuleService) CreateOrUpdateOverride(projectID, workspaceModuleID uint64, req request.ModuleOverrideRequest) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.Where("id = ? AND project_id IS NULL", workspaceModuleID).First(&module).Error; err != nil {
		return nil, common.NotFound("Workspace module not found")
	}

	var override model.ModuleInheritanceOverride
	err := s.db.Where("project_id = ? AND workspace_module_id = ?", projectID, workspaceModuleID).First(&override).Error

	if err == gorm.ErrRecordNotFound {
		override = model.ModuleInheritanceOverride{
			ProjectID:         projectID,
			WorkspaceModuleID: workspaceModuleID,
		}
	}

	override.IsExcluded = req.IsExcluded
	override.OverrideName = req.OverrideName
	override.OverrideDescription = req.OverrideDescription

	if err := s.db.Save(&override).Error; err != nil {
		return nil, common.Internal("Failed to save override")
	}

	return s.buildResponse(module, &override), nil
}

func (s *ModuleService) DeleteOverride(projectID, workspaceModuleID uint64) error {
	result := s.db.Where("project_id = ? AND workspace_module_id = ?", projectID, workspaceModuleID).Delete(&model.ModuleInheritanceOverride{})
	if result.RowsAffected == 0 {
		return common.NotFound("Override not found")
	}
	return nil
}

func (s *ModuleService) GetOverride(projectID, workspaceModuleID uint64) (*model.ModuleInheritanceOverride, error) {
	var override model.ModuleInheritanceOverride
	err := s.db.Where("project_id = ? AND workspace_module_id = ?", projectID, workspaceModuleID).First(&override).Error
	if err != nil {
		return nil, err
	}
	return &override, nil
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

	if module.ProjectID != nil && issue.ProjectID != *module.ProjectID {
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

func (s *ModuleService) BuildTree(projectID, workspaceID uint64) ([]*response.ModuleTreeNode, error) {
	overrides, err := s.getOverrides(projectID)
	if err != nil {
		return nil, common.Internal("Failed to load overrides")
	}

	var modules []model.Module
	if err := s.db.Where("(project_id = ? OR (project_id IS NULL AND workspace_id = ?)) AND is_archived = ?", projectID, workspaceID, false).
		Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to load modules")
	}

	childrenMap := make(map[uint64][]*response.ModuleTreeNode)
	var roots []*response.ModuleTreeNode

	for _, m := range modules {
		if m.ProjectID == nil {
			if override, exists := overrides[m.ID]; exists && override.IsExcluded {
				continue
			}
		}

		total, completed, _ := s.countIssues(m.ID)
		progress := 0
		if total > 0 {
			progress = int(float64(completed) / float64(total) * 100)
		}

		var override *model.ModuleInheritanceOverride
		if m.ProjectID == nil {
			if o, exists := overrides[m.ID]; exists {
				override = o
			}
		}

		node := &response.ModuleTreeNode{
			ModuleResponse:  *s.buildResponse(m, override),
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
			if n.ID != nil {
				n.Children = childrenMap[*n.ID]
			}
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
