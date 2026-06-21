package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
)

type ModuleService struct {
	db *gorm.DB
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{db: db}
}

func (s *ModuleService) Create(req request.ModuleCreate) (*response.ModuleResponse, error) {
	module := model.Module{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		WorkspaceID: req.WorkspaceID,
		ParentID:    req.ParentID,
	}

	if err := s.db.Create(&module).Error; err != nil {
		return nil, common.Internal("Failed to create module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) List(projectID, workspaceID uint64, includeArchived bool) ([]response.ModuleResponse, error) {
	query := s.db.Where("project_id = ? AND workspace_id = ?", projectID, workspaceID)
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}

	var modules []model.Module
	if err := query.Order("\"order\", created_at").Find(&modules).Error; err != nil {
		return nil, common.Internal("Failed to list modules")
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Module not found")
		}
		return nil, common.Internal("Failed to get module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) Update(moduleID uint64, req request.ModuleUpdate) (*response.ModuleResponse, error) {
	var module model.Module
	if err := s.db.First(&module, moduleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Module not found")
		}
		return nil, common.Internal("Failed to get module")
	}

	if req.Name != "" {
		module.Name = req.Name
	}
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.ParentID != nil {
		module.ParentID = req.ParentID
	}

	if err := s.db.Save(&module).Error; err != nil {
		return nil, common.Internal("Failed to update module")
	}

	return s.buildResponse(module), nil
}

func (s *ModuleService) Delete(moduleID uint64) error {
	if err := s.db.Delete(&model.Module{}, moduleID).Error; err != nil {
		return common.Internal("Failed to delete module")
	}
	return nil
}

func (s *ModuleService) buildResponse(module model.Module) *response.ModuleResponse {
	return &response.ModuleResponse{
		ID:          module.ID,
		Name:        module.Name,
		Description: module.Description,
		ProjectID:   module.ProjectID,
		WorkspaceID: module.WorkspaceID,
		ParentID:    module.ParentID,
		Order:       module.Order,
		IsArchived:  module.IsArchived,
		ArchivedAt:  module.ArchivedAt,
		CreatedAt:   module.CreatedAt,
		UpdatedAt:   module.UpdatedAt,
	}
}
