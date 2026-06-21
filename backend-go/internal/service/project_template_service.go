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
			TypeTemplateID: link.TypeTemplateID,
			IsRequired:     link.IsRequired,
			DefaultStateID: link.DefaultStateID,
			Sequence:       link.Sequence,
			TypeName:       link.TypeTemplate.Name,
			TypeColor:      link.TypeTemplate.Color,
			TypeIcon:       link.TypeTemplate.Icon,
			TypeLevel:      link.TypeTemplate.Level,
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
	if err := s.db.Preload("TypeLinks.TypeTemplate").Where("workspace_id = ?", workspaceID).
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
	if err := s.db.Preload("TypeLinks.TypeTemplate").First(&t, templateID).Error; err != nil {
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

func (s *ProjectTemplateService) AddType(templateID, typeTemplateID uint64, isRequired bool, defaultStateID *uint64, sequence int) (*response.ProjectTemplateTypeResponse, error) {
	var t model.ProjectTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		return nil, common.NotFound("Template not found")
	}
	var tt model.IssueTypeTemplate
	if err := s.db.First(&tt, typeTemplateID).Error; err != nil {
		return nil, common.NotFound("Type template not found")
	}
	var count int64
	s.db.Model(&model.ProjectTemplateType{}).Where("template_id = ? AND type_template_id = ?", templateID, typeTemplateID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Type template already in project template")
	}
	link := model.ProjectTemplateType{
		TemplateID: templateID, TypeTemplateID: typeTemplateID,
		IsRequired: isRequired, DefaultStateID: defaultStateID, Sequence: sequence,
	}
	if err := s.db.Create(&link).Error; err != nil {
		return nil, common.Internal("Failed to add type template")
	}
	return &response.ProjectTemplateTypeResponse{
		TemplateID: templateID, TypeTemplateID: typeTemplateID,
		IsRequired: isRequired, DefaultStateID: defaultStateID, Sequence: sequence,
		TypeName: tt.Name, TypeColor: tt.Color, TypeIcon: tt.Icon,
	}, nil
}

func (s *ProjectTemplateService) RemoveType(templateID, typeTemplateID uint64) error {
	result := s.db.Where("template_id = ? AND type_template_id = ?", templateID, typeTemplateID).Delete(&model.ProjectTemplateType{})
	if result.RowsAffected == 0 {
		return common.NotFound("Type template not in project template")
	}
	return nil
}
