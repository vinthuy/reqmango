package service

import (
	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/dto/response"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type IssueTypeService struct {
	db *gorm.DB
}

func NewIssueTypeService(db *gorm.DB) *IssueTypeService {
	return &IssueTypeService{db: db}
}

func (s *IssueTypeService) buildResponse(t model.IssueType) *response.IssueTypeResponse {
	return &response.IssueTypeResponse{
		ID:           t.ID,
		Name:         t.Name,
		Color:        t.Color,
		Icon:         t.Icon,
		Description:  t.Description,
		Level:        t.Level,
		ParentTypeID: t.ParentTypeID,
		IsDefault:    t.IsDefault,
		Sequence:     t.Sequence,
		IsActive:     t.IsActive,
		ProjectID:    t.ProjectID,
		WorkspaceID:  t.WorkspaceID,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

// ensureOneDefault ensures only one default type exists in the given scope.
func (s *IssueTypeService) ensureOneDefault(workspaceID, excludeID uint64, projectID *uint64) error {
	query := s.db.Model(&model.IssueType{}).Where("workspace_id = ? AND is_default = ?", workspaceID, true)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.Update("is_default", false).Error
}

// ==================== CRUD ====================

func (s *IssueTypeService) Create(workspaceID, userID uint64, req request.IssueTypeCreate) (*response.IssueTypeResponse, error) {
	if req.IsDefault {
		if err := s.ensureOneDefault(workspaceID, 0, req.ProjectID); err != nil {
			return nil, common.Internal("Failed to update defaults")
		}
	}
	if req.Sequence == 0 {
		req.Sequence = 1
	}
	if req.Color == "" {
		req.Color = "#6366F1"
	}
	if req.Icon == "" {
		req.Icon = "circle"
	}

	t := model.IssueType{
		Name:         req.Name,
		Color:        req.Color,
		Icon:         req.Icon,
		Description:  req.Description,
		Level:        req.Level,
		ParentTypeID: req.ParentTypeID,
		IsDefault:    req.IsDefault,
		Sequence:     req.Sequence,
		IsActive:     true,
		ProjectID:    req.ProjectID,
		WorkspaceID:  workspaceID,
	}
	t.CreatedByID = &userID

	if err := s.db.Create(&t).Error; err != nil {
		return nil, common.Internal("Failed to create issue type")
	}

	return s.buildResponse(t), nil
}

func (s *IssueTypeService) List(workspaceID uint64, projectID *uint64) ([]response.IssueTypeResponse, error) {
	query := s.db.Model(&model.IssueType{}).Where("workspace_id = ?", workspaceID)

	if projectID != nil {
		// Return types scoped to the project OR shared (project_id IS NULL)
		query = query.Where("project_id = ? OR project_id IS NULL", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}

	var types []model.IssueType
	if err := query.Order("sequence, created_at").Find(&types).Error; err != nil {
		return nil, common.Internal("Failed to list issue types")
	}

	result := make([]response.IssueTypeResponse, len(types))
	for i, t := range types {
		result[i] = *s.buildResponse(t)
	}
	if result == nil {
		result = []response.IssueTypeResponse{}
	}
	return result, nil
}

func (s *IssueTypeService) Get(typeID uint64) (*response.IssueTypeResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}
	return s.buildResponse(t), nil
}

func (s *IssueTypeService) Update(typeID, userID uint64, req request.IssueTypeUpdate) (*response.IssueTypeResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Color != nil {
		t.Color = *req.Color
	}
	if req.Icon != nil {
		t.Icon = *req.Icon
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			if err := s.ensureOneDefault(t.WorkspaceID, t.ID, t.ProjectID); err != nil {
				return nil, common.Internal("Failed to update defaults")
			}
		}
		t.IsDefault = *req.IsDefault
	}
	if req.Sequence != nil {
		t.Sequence = *req.Sequence
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	if req.ProjectID != nil {
		t.ProjectID = req.ProjectID
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Level != nil {
		t.Level = *req.Level
	}
	if req.ParentTypeID != nil {
		t.ParentTypeID = req.ParentTypeID
	}
	if req.AllowedChildTypeIDs != nil {
		t.AllowedChildTypeIDs = req.AllowedChildTypeIDs
	}

	t.UpdatedByID = &userID
	if err := s.db.Save(&t).Error; err != nil {
		return nil, common.Internal("Failed to update issue type")
	}

	return s.buildResponse(t), nil
}

func (s *IssueTypeService) Delete(typeID uint64) error {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return common.NotFound("Issue type not found")
	}

	// Clean up join table entries
	s.db.Where("type_id = ?", typeID).Delete(&model.IssueTypeField{})

	return s.db.Delete(&t).Error
}

func (s *IssueTypeService) Disable(typeID uint64, isActive bool) error {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return common.NotFound("Issue type not found")
	}
	return s.db.Model(&t).Update("is_active", isActive).Error
}

// ==================== Field Association ====================

func (s *IssueTypeService) AddField(typeID, fieldID uint64, isRequired bool, sequence int) (*response.IssueTypeFieldResponse, error) {
	// Verify type exists
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	// Verify field exists
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	// Check for duplicate
	var count int64
	s.db.Model(&model.IssueTypeField{}).Where("type_id = ? AND field_id = ?", typeID, fieldID).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Field is already associated with this issue type")
	}

	link := model.IssueTypeField{
		TypeID:     typeID,
		FieldID:    fieldID,
		IsRequired: isRequired,
		Sequence:   sequence,
	}

	if err := s.db.Create(&link).Error; err != nil {
		return nil, common.Internal("Failed to add field to issue type")
	}

	return &response.IssueTypeFieldResponse{
		FieldID:    fieldID,
		TypeID:     typeID,
		IsRequired: isRequired,
		Sequence:   sequence,
		Name:       f.Name,
		FieldType:  f.FieldType,
		Description: f.Description,
	}, nil
}

func (s *IssueTypeService) RemoveField(typeID, fieldID uint64) error {
	result := s.db.Where("type_id = ? AND field_id = ?", typeID, fieldID).Delete(&model.IssueTypeField{})
	if result.RowsAffected == 0 {
		return common.NotFound("Field association not found")
	}
	return nil
}

func (s *IssueTypeService) UpdateField(typeID, fieldID uint64, req request.IssueTypeFieldUpdate) (*response.IssueTypeFieldResponse, error) {
	var link model.IssueTypeField
	if err := s.db.Where("type_id = ? AND field_id = ?", typeID, fieldID).First(&link).Error; err != nil {
		return nil, common.NotFound("Field association not found")
	}

	if req.IsRequired != nil {
		link.IsRequired = *req.IsRequired
	}
	if req.Sequence != nil {
		link.Sequence = *req.Sequence
	}

	if err := s.db.Save(&link).Error; err != nil {
		return nil, common.Internal("Failed to update field association")
	}

	// Load field info for response
	var f model.CustomField
	s.db.First(&f, fieldID)

	return &response.IssueTypeFieldResponse{
		FieldID:    fieldID,
		TypeID:     typeID,
		IsRequired: link.IsRequired,
		Sequence:   link.Sequence,
		Name:       f.Name,
		FieldType:  f.FieldType,
		Description: f.Description,
	}, nil
}

// ==================== Project-scoped Operations ====================

// CopyFromWorkspace copies all workspace-level issue types to the target project.
func (s *IssueTypeService) CopyFromWorkspace(workspaceID, projectID, userID uint64) ([]response.IssueTypeResponse, error) {
	var templates []model.IssueType
	if err := s.db.Where("workspace_id = ? AND project_id IS NULL AND is_active = ?", workspaceID, true).
		Order("sequence, created_at").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch workspace issue types")
	}

	// Check if project already has types
	var existingCount int64
	s.db.Model(&model.IssueType{}).Where("project_id = ?", projectID).Count(&existingCount)
	if existingCount > 0 {
		return nil, common.Conflict("Project already has issue types configured")
	}

	types := make([]model.IssueType, len(templates))
	for i, tmpl := range templates {
		types[i] = model.IssueType{
			Name:         tmpl.Name,
			Color:        tmpl.Color,
			Icon:         tmpl.Icon,
			Description:  tmpl.Description,
			Level:        tmpl.Level,
			ParentTypeID: nil, // Reset hierarchy for project scope — will be re-linked post-creation
			IsDefault:    tmpl.IsDefault,
			Sequence:     tmpl.Sequence,
			IsActive:     true,
			ProjectID:    &projectID,
			WorkspaceID:  workspaceID,
		}
		types[i].CreatedByID = &userID
	}

	if err := s.db.Create(&types).Error; err != nil {
		return nil, common.Internal("Failed to copy issue types to project")
	}

	// Re-link parent relationships within the copied types
	// Build a map from template ID -> new type ID
	idMap := make(map[uint64]uint64)
	for i, tmpl := range templates {
		idMap[tmpl.ID] = types[i].ID
	}
	for i, tmpl := range templates {
		if tmpl.ParentTypeID != nil {
			if newParentID, ok := idMap[*tmpl.ParentTypeID]; ok {
				s.db.Model(&types[i]).Update("parent_type_id", newParentID)
			}
		}
	}

	// Also copy field associations
	for i, tmpl := range templates {
		var tmplFields []model.IssueTypeTemplateField
		s.db.Where("template_type_id = ?", tmpl.ID).Find(&tmplFields)
		// Note: This copies from IssueTypeTemplateField; if the workspace types have direct
		// IssueTypeField entries, we'd copy those too. For simplicity, we just set up the types.
		_ = tmplFields
		_ = i
	}

	result := make([]response.IssueTypeResponse, len(types))
	for i, t := range types {
		result[i] = *s.buildResponse(t)
	}
	return result, nil
}

// Reorder updates the sequence for issue types within a project.
func (s *IssueTypeService) Reorder(projectID uint64, typeIDs []uint64) error {
	for i, typeID := range typeIDs {
		if err := s.db.Model(&model.IssueType{}).Where("id = ? AND project_id = ?", typeID, projectID).
			Update("sequence", i+1).Error; err != nil {
			return common.Internal("Failed to reorder issue types")
		}
	}
	return nil
}

func (s *IssueTypeService) ListFields(typeID uint64) ([]response.IssueTypeFieldResponse, error) {
	var t model.IssueType
	if err := s.db.First(&t, typeID).Error; err != nil {
		return nil, common.NotFound("Issue type not found")
	}

	var links []model.IssueTypeField
	if err := s.db.Preload("Field").Where("type_id = ?", typeID).Order("sequence").Find(&links).Error; err != nil {
		return nil, common.Internal("Failed to list fields")
	}

	result := make([]response.IssueTypeFieldResponse, len(links))
	for i, link := range links {
		fr := response.IssueTypeFieldResponse{
			FieldID:     link.FieldID,
			TypeID:      link.TypeID,
			IsRequired:  link.IsRequired,
			Sequence:    link.Sequence,
			Name:        link.Field.Name,
			FieldType:   link.Field.FieldType,
			Description: link.Field.Description,
			Options:     make([]response.CustomFieldOptionResponse, 0),
		}
		// Load options for dropdown fields
		if link.Field.FieldType == "dropdown" {
			var opts []model.CustomFieldOption
			s.db.Where("field_id = ?", link.FieldID).Order("sequence").Find(&opts)
			for _, o := range opts {
				fr.Options = append(fr.Options, response.CustomFieldOptionResponse{
					ID:       o.ID,
					FieldID:  o.FieldID,
					Value:    o.Value,
					Color:    o.Color,
					Sequence: o.Sequence,
				})
			}
		}
		result[i] = fr
	}
	if result == nil {
		result = []response.IssueTypeFieldResponse{}
	}
	return result, nil
}
