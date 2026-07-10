package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type WorkItemTemplateService struct {
	db *gorm.DB
}

func NewWorkItemTemplateService(db *gorm.DB) *WorkItemTemplateService {
	return &WorkItemTemplateService{db: db}
}

func (s *WorkItemTemplateService) List(projectID uint64) ([]response.WorkItemTemplateResponse, error) {
	var templates []model.WorkItemTemplate
	if err := s.db.Where("project_id = ?", projectID).Order("is_default DESC, name").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch work item templates")
	}

	return s.convertToResponses(templates), nil
}

func (s *WorkItemTemplateService) Get(id, projectID uint64) (*response.WorkItemTemplateResponse, error) {
	var template model.WorkItemTemplate
	if err := s.db.Where("id = ? AND project_id = ?", id, projectID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Work item template not found")
		}
		return nil, common.Internal("Failed to fetch work item template")
	}

	return s.convertToResponse(&template), nil
}

func (s *WorkItemTemplateService) Create(projectID, workspaceID uint64, req *request.WorkItemTemplateCreate) (*response.WorkItemTemplateResponse, error) {
	if req.IsDefault {
		if err := s.db.Model(&model.WorkItemTemplate{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false).Error; err != nil {
			return nil, common.Internal("Failed to update default template")
		}
	}

	defaultsJSON, err := json.Marshal(req.Defaults)
	if err != nil {
		return nil, common.BadRequest("Invalid defaults JSON")
	}

	template := &model.WorkItemTemplate{
		Name:        req.Name,
		IssueTypeID: req.IssueTypeID,
		Defaults:    defaultsJSON,
		IsDefault:   req.IsDefault,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}

	if err := s.db.Create(template).Error; err != nil {
		return nil, common.Internal("Failed to create work item template")
	}

	return s.convertToResponse(template), nil
}

func (s *WorkItemTemplateService) Update(id, projectID uint64, req *request.WorkItemTemplateUpdate) (*response.WorkItemTemplateResponse, error) {
	var template model.WorkItemTemplate
	if err := s.db.Where("id = ? AND project_id = ?", id, projectID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Work item template not found")
		}
		return nil, common.Internal("Failed to fetch work item template")
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := s.db.Model(&model.WorkItemTemplate{}).Where("project_id = ? AND is_default = ?", projectID, true).Update("is_default", false).Error; err != nil {
			return nil, common.Internal("Failed to update default template")
		}
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IssueTypeID != nil {
		updates["issue_type_id"] = *req.IssueTypeID
	}
	if req.Defaults != nil {
		defaultsJSON, err := json.Marshal(*req.Defaults)
		if err != nil {
			return nil, common.BadRequest("Invalid defaults JSON")
		}
		updates["defaults"] = defaultsJSON
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	if err := s.db.Model(&template).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update work item template")
	}

	return s.convertToResponse(&template), nil
}

func (s *WorkItemTemplateService) Delete(id, projectID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&model.WorkItemTemplate{})
	if result.Error != nil {
		return common.Internal("Failed to delete work item template")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Work item template not found")
	}
	return nil
}

func (s *WorkItemTemplateService) GetDefault(projectID uint64) (*response.WorkItemTemplateResponse, error) {
	var template model.WorkItemTemplate
	if err := s.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, common.Internal("Failed to fetch default template")
	}
	return s.convertToResponse(&template), nil
}

func (s *WorkItemTemplateService) convertToResponses(templates []model.WorkItemTemplate) []response.WorkItemTemplateResponse {
	responses := make([]response.WorkItemTemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = *s.convertToResponse(&t)
	}
	return responses
}

func (s *WorkItemTemplateService) convertToResponse(template *model.WorkItemTemplate) *response.WorkItemTemplateResponse {
	return &response.WorkItemTemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		IssueTypeID: template.IssueTypeID,
		Defaults:    template.Defaults,
		IsDefault:   template.IsDefault,
		ProjectID:   template.ProjectID,
		WorkspaceID: template.WorkspaceID,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}
