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

func (s *WorkItemTemplateService) List(projectID uint64, issueTypeID *uint64) ([]response.WorkItemTemplateResponse, error) {
	var templates []model.WorkItemTemplate
	q := s.db.Where("project_id = ?", projectID).Preload("IssueType")
	if issueTypeID != nil {
		q = q.Where("issue_type_id = ?", *issueTypeID)
	}
	if err := q.Order("is_default DESC, created_at ASC").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch work item templates")
	}

	resps := make([]response.WorkItemTemplateResponse, len(templates))
	for i, t := range templates {
		resps[i] = templateToResponse(&t)
	}
	return resps, nil
}

func (s *WorkItemTemplateService) Get(projectID, id uint64) (*response.WorkItemTemplateResponse, error) {
	var t model.WorkItemTemplate
	if err := s.db.Preload("IssueType").Where("id = ? AND project_id = ?", id, projectID).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Work item template not found")
		}
		return nil, common.Internal("Failed to fetch work item template")
	}
	resp := templateToResponse(&t)
	return &resp, nil
}

func (s *WorkItemTemplateService) Create(projectID uint64, req *request.WorkItemTemplateCreate) (*response.WorkItemTemplateResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Failed to fetch project")
	}

	defaultsJSON := normalizeDefaultsJSON(req.Defaults)

	t := &model.WorkItemTemplate{
		Name:         req.Name,
		Description:  req.Description,
		IssueTypeID:  req.IssueTypeID,
		DefaultsJSON: defaultsJSON,
		IsDefault:    req.IsDefault,
		ProjectID:    projectID,
		WorkspaceID:  project.WorkspaceID,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault {
			q := tx.Model(&model.WorkItemTemplate{}).
				Where("project_id = ? AND is_default = ?", projectID, true)
			if req.IssueTypeID != nil {
				q = q.Where("issue_type_id = ?", *req.IssueTypeID)
			} else {
				q = q.Where("issue_type_id IS NULL")
			}
			if err := q.Update("is_default", false).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(t).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to create work item template")
	}

	if err := s.db.Preload("IssueType").First(t, t.ID).Error; err != nil {
		return nil, common.Internal("Failed to fetch created work item template")
	}

	resp := templateToResponse(t)
	return &resp, nil
}

func (s *WorkItemTemplateService) Update(projectID, id uint64, req *request.WorkItemTemplateUpdate) (*response.WorkItemTemplateResponse, error) {
	var t model.WorkItemTemplate
	if err := s.db.Where("id = ? AND project_id = ?", id, projectID).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Work item template not found")
		}
		return nil, common.Internal("Failed to fetch work item template")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IssueTypeID != nil {
		updates["issue_type_id"] = *req.IssueTypeID
	}
	if req.Defaults != nil {
		updates["defaults"] = normalizeDefaultsJSON(*req.Defaults)
	}

	newIssueTypeID := t.IssueTypeID
	if req.IssueTypeID != nil {
		newIssueTypeID = req.IssueTypeID
	}

	setDefault := false
	if req.IsDefault != nil {
		setDefault = *req.IsDefault
		updates["is_default"] = *req.IsDefault
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if setDefault {
			q := tx.Model(&model.WorkItemTemplate{}).
				Where("project_id = ? AND id != ? AND is_default = ?", t.ProjectID, id, true)
			if newIssueTypeID != nil {
				q = q.Where("issue_type_id = ?", *newIssueTypeID)
			} else {
				q = q.Where("issue_type_id IS NULL")
			}
			if err := q.Update("is_default", false).Error; err != nil {
				return err
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(&t).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, common.Internal("Failed to update work item template")
	}

	if err := s.db.Preload("IssueType").Where("id = ? AND project_id = ?", id, projectID).First(&t).Error; err != nil {
		return nil, common.Internal("Failed to fetch updated work item template")
	}

	resp := templateToResponse(&t)
	return &resp, nil
}

func (s *WorkItemTemplateService) Delete(projectID, id uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&model.WorkItemTemplate{})
	if result.Error != nil {
		return common.Internal("Failed to delete work item template")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Work item template not found")
	}
	return nil
}

func normalizeDefaultsJSON(defaults map[string]interface{}) json.RawMessage {
	if defaults == nil || len(defaults) == 0 {
		return json.RawMessage("{}")
	}
	raw, err := json.Marshal(defaults)
	if err != nil {
		return json.RawMessage("{}")
	}
	return raw
}

func templateToResponse(t *model.WorkItemTemplate) response.WorkItemTemplateResponse {
	resp := response.WorkItemTemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		IssueTypeID: t.IssueTypeID,
		Defaults:    t.DefaultsJSON,
		IsDefault:   t.IsDefault,
		ProjectID:   t.ProjectID,
		WorkspaceID: t.WorkspaceID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.IssueType != nil {
		resp.IssueType = &response.IssueTypeLite{
			ID:    t.IssueType.ID,
			Name:  t.IssueType.Name,
			Color: t.IssueType.Color,
			Icon:  t.IssueType.Icon,
		}
	}
	return resp
}
