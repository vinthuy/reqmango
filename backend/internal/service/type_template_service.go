package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type TypeTemplateService struct {
	db *gorm.DB
}

func NewTypeTemplateService(db *gorm.DB) *TypeTemplateService {
	return &TypeTemplateService{db: db}
}

func (s *TypeTemplateService) buildResponse(t model.IssueTypeTemplate) *response.TypeTemplateResponse {
	r := &response.TypeTemplateResponse{
		ID:           t.ID,
		Name:         t.Name,
		Color:        t.Color,
		Icon:         t.Icon,
		Description:  t.Description,
		Level:        t.Level,
		ParentTypeID: t.ParentTypeID,
		WorkspaceID:  t.WorkspaceID,
		Fields:       make([]response.TypeTemplateFieldResponse, 0),
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	for _, link := range t.FieldLinks {
		r.Fields = append(r.Fields, response.TypeTemplateFieldResponse{
			TemplateTypeID: link.TemplateTypeID,
			FieldID:        link.FieldID,
			IsRequired:     link.IsRequired,
			Sequence:       link.Sequence,
			FieldName:      link.Field.Name,
			FieldType:      link.Field.FieldType,
		})
	}
	return r
}

func (s *TypeTemplateService) Create(workspaceID, userID uint64, req request.TypeTemplateCreate) (*response.TypeTemplateResponse, error) {
	t := model.IssueTypeTemplate{
		Name:         req.Name,
		Color:        req.Color,
		Icon:         req.Icon,
		Description:  req.Description,
		Level:        req.Level,
		ParentTypeID: req.ParentTypeID,
		WorkspaceID:  workspaceID,
	}
	if t.Color == "" { t.Color = "#6366F1" }
	if t.Icon == ""  { t.Icon = "circle" }
	t.CreatedByID = &userID
	if err := s.db.Create(&t).Error; err != nil {
		return nil, common.Internal("Failed to create type template")
	}
	return s.buildResponse(t), nil
}

func (s *TypeTemplateService) List(workspaceID uint64) ([]response.TypeTemplateResponse, error) {
	var templates []model.IssueTypeTemplate
	if err := s.db.Preload("FieldLinks.Field").Where("workspace_id = ?", workspaceID).
		Order("level, created_at").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to list type templates")
	}
	result := make([]response.TypeTemplateResponse, len(templates))
	for i, t := range templates {
		result[i] = *s.buildResponse(t)
	}
	if result == nil { result = []response.TypeTemplateResponse{} }
	return result, nil
}

func (s *TypeTemplateService) Get(id uint64) (*response.TypeTemplateResponse, error) {
	var t model.IssueTypeTemplate
	if err := s.db.Preload("FieldLinks.Field").First(&t, id).Error; err != nil {
		return nil, common.NotFound("Type template not found")
	}
	return s.buildResponse(t), nil
}

func (s *TypeTemplateService) Update(id, userID uint64, req request.TypeTemplateUpdate) (*response.TypeTemplateResponse, error) {
	var t model.IssueTypeTemplate
	if err := s.db.First(&t, id).Error; err != nil {
		return nil, common.NotFound("Type template not found")
	}
	if req.Name != nil { t.Name = *req.Name }
	if req.Color != nil { t.Color = *req.Color }
	if req.Icon != nil { t.Icon = *req.Icon }
	if req.Description != nil { t.Description = *req.Description }
	if req.Level != nil { t.Level = *req.Level }
	if req.ParentTypeID != nil { t.ParentTypeID = req.ParentTypeID }
	t.UpdatedByID = &userID
	if err := s.db.Save(&t).Error; err != nil {
		return nil, common.Internal("Failed to update type template")
	}
	return s.Get(id)
}

func (s *TypeTemplateService) Delete(id uint64) error {
	var t model.IssueTypeTemplate
	if err := s.db.First(&t, id).Error; err != nil {
		return common.NotFound("Type template not found")
	}
	s.db.Where("template_type_id = ?", id).Delete(&model.IssueTypeTemplateField{})
	s.db.Where("type_template_id = ?", id).Delete(&model.ProjectTemplateType{})
	return s.db.Delete(&t).Error
}

func (s *TypeTemplateService) BindField(templateTypeID, fieldID uint64, isRequired bool, seq int) (*response.TypeTemplateFieldResponse, error) {
	var t model.IssueTypeTemplate
	if err := s.db.First(&t, templateTypeID).Error; err != nil {
		return nil, common.NotFound("Type template not found")
	}
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}
	var count int64
	s.db.Model(&model.IssueTypeTemplateField{}).Where("template_type_id = ? AND field_id = ?", templateTypeID, fieldID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Field already bound to this type template")
	}
	link := model.IssueTypeTemplateField{TemplateTypeID: templateTypeID, FieldID: fieldID, IsRequired: isRequired, Sequence: seq}
	if err := s.db.Create(&link).Error; err != nil {
		return nil, common.Internal("Failed to bind field")
	}
	return &response.TypeTemplateFieldResponse{
		TemplateTypeID: templateTypeID, FieldID: fieldID, IsRequired: isRequired, Sequence: seq,
		FieldName: f.Name, FieldType: f.FieldType,
	}, nil
}

func (s *TypeTemplateService) UnbindField(templateTypeID, fieldID uint64) error {
	result := s.db.Where("template_type_id = ? AND field_id = ?", templateTypeID, fieldID).Delete(&model.IssueTypeTemplateField{})
	if result.RowsAffected == 0 {
		return common.NotFound("Field binding not found")
	}
	return nil
}
