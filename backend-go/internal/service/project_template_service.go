package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type ProjectTemplateService struct {
	db *gorm.DB
}

func NewProjectTemplateService(db *gorm.DB) *ProjectTemplateService {
	return &ProjectTemplateService{db: db}
}

func (s *ProjectTemplateService) buildResponse(t model.ProjectTemplate) *response.ProjectTemplateResponse {
	r := &response.ProjectTemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		WorkspaceID: t.WorkspaceID,
		IsDefault:   t.IsDefault,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Types:       make([]response.ProjectTemplateTypeResponse, 0),
	}
	for _, link := range t.TypeLinks {
		r.Types = append(r.Types, response.ProjectTemplateTypeResponse{
			TemplateID:     link.TemplateID,
			IssueTypeID:    link.IssueTypeID,
			IsRequired:     link.IsRequired,
			DefaultStateID: link.DefaultStateID,
			Sequence:       link.Sequence,
			TypeName:       link.IssueType.Name,
			TypeColor:      link.IssueType.Color,
			TypeIcon:       link.IssueType.Icon,
		})
	}
	return r
}

func (s *ProjectTemplateService) Create(workspaceID, userID uint64, req request.ProjectTemplateCreate) (*response.ProjectTemplateResponse, error) {
	if req.IsDefault {
		s.db.Model(&model.ProjectTemplate{}).Where("workspace_id = ? AND is_default = ?", workspaceID, true).
			Update("is_default", false)
	}
	t := model.ProjectTemplate{
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: workspaceID,
		IsDefault:   req.IsDefault,
	}
	t.CreatedByID = &userID
	if err := s.db.Create(&t).Error; err != nil {
		return nil, common.Internal("Failed to create template")
	}
	return s.buildResponse(t), nil
}

func (s *ProjectTemplateService) List(workspaceID uint64) ([]response.ProjectTemplateResponse, error) {
	var templates []model.ProjectTemplate
	if err := s.db.Preload("TypeLinks.IssueType").Where("workspace_id = ?", workspaceID).
		Order("created_at").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to list templates")
	}
	result := make([]response.ProjectTemplateResponse, len(templates))
	for i, t := range templates {
		result[i] = *s.buildResponse(t)
	}
	if result == nil {
		result = []response.ProjectTemplateResponse{}
	}
	return result, nil
}

func (s *ProjectTemplateService) Get(templateID uint64) (*response.ProjectTemplateResponse, error) {
	var t model.ProjectTemplate
	if err := s.db.Preload("TypeLinks.IssueType").First(&t, templateID).Error; err != nil {
		return nil, common.NotFound("Template not found")
	}
	return s.buildResponse(t), nil
}

func (s *ProjectTemplateService) Update(templateID, userID uint64, req request.ProjectTemplateUpdate) (*response.ProjectTemplateResponse, error) {
	var t model.ProjectTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		return nil, common.NotFound("Template not found")
	}
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.IsDefault != nil && *req.IsDefault {
		s.db.Model(&model.ProjectTemplate{}).Where("workspace_id = ? AND is_default = ?", t.WorkspaceID, true).
			Update("is_default", false)
		t.IsDefault = true
	}
	t.UpdatedByID = &userID
	if err := s.db.Save(&t).Error; err != nil {
		return nil, common.Internal("Failed to update template")
	}
	return s.Get(templateID)
}

func (s *ProjectTemplateService) Delete(templateID uint64) error {
	var t model.ProjectTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		return common.NotFound("Template not found")
	}
	s.db.Where("template_id = ?", templateID).Delete(&model.ProjectTemplateType{})
	return s.db.Delete(&t).Error
}

func (s *ProjectTemplateService) AddType(templateID, issueTypeID uint64, isRequired bool, defaultStateID *uint64, sequence int) (*response.ProjectTemplateTypeResponse, error) {
	// Verify template exists
	var t model.ProjectTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		return nil, common.NotFound("Template not found")
	}
	// Verify issue type exists
	var it model.IssueType
	if err := s.db.First(&it, issueTypeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}
	// Check for duplicate
	var count int64
	s.db.Model(&model.ProjectTemplateType{}).Where("template_id = ? AND issue_type_id = ?", templateID, issueTypeID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Issue type already in template")
	}
	link := model.ProjectTemplateType{
		TemplateID:     templateID,
		IssueTypeID:    issueTypeID,
		IsRequired:     isRequired,
		DefaultStateID: defaultStateID,
		Sequence:       sequence,
	}
	if err := s.db.Create(&link).Error; err != nil {
		return nil, common.Internal("Failed to add type to template")
	}
	return &response.ProjectTemplateTypeResponse{
		TemplateID:     templateID,
		IssueTypeID:    issueTypeID,
		IsRequired:     isRequired,
		DefaultStateID: defaultStateID,
		Sequence:       sequence,
		TypeName:       it.Name,
		TypeColor:      it.Color,
		TypeIcon:       it.Icon,
	}, nil
}

func (s *ProjectTemplateService) RemoveType(templateID, issueTypeID uint64) error {
	result := s.db.Where("template_id = ? AND issue_type_id = ?", templateID, issueTypeID).Delete(&model.ProjectTemplateType{})
	if result.RowsAffected == 0 {
		return common.NotFound("Type not found in template")
	}
	return nil
}
