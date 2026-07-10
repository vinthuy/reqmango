package service

import (
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type CustomFieldService struct {
	db *gorm.DB
}

func NewCustomFieldService(db *gorm.DB) *CustomFieldService {
	return &CustomFieldService{db: db}
}

var validFieldTypes = map[string]bool{
	"text": true, "number": true, "dropdown": true,
	"boolean": true, "date": true, "member": true, "url": true,
}

func (s *CustomFieldService) buildResponse(f model.CustomField) *response.CustomFieldResponse {
	r := &response.CustomFieldResponse{
		ID:           f.ID,
		Name:         f.Name,
		Description:  f.Description,
		FieldType:    f.FieldType,
		IsRequired:   f.IsRequired,
		DefaultValue: f.DefaultValue,
		Placeholder:  f.Placeholder,
		IsActive:     f.IsActive,
		ProjectID:    f.ProjectID,
		WorkspaceID:  f.WorkspaceID,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		Options:      make([]response.CustomFieldOptionResponse, 0),
	}

	if f.FieldType == "dropdown" {
		var opts []model.CustomFieldOption
		s.db.Where("field_id = ?", f.ID).Order("sequence").Find(&opts)
		r.Options = make([]response.CustomFieldOptionResponse, len(opts))
		for i, o := range opts {
			r.Options[i] = response.CustomFieldOptionResponse{
				ID:       o.ID,
				FieldID:  o.FieldID,
				Value:    o.Value,
				Color:    o.Color,
				Sequence: o.Sequence,
			}
		}
	}
	return r
}

func (s *CustomFieldService) validateFieldType(fieldType string) error {
	if !validFieldTypes[fieldType] {
		return common.Validation("Invalid field type. Must be one of: text, number, dropdown, boolean, date, member, url")
	}
	return nil
}

// validateFieldValue validates a value against the field's type definition.
func (s *CustomFieldService) validateFieldValue(field *model.CustomField, value string) error {
	if field.FieldType == "dropdown" && value != "" {
		var count int64
		s.db.Model(&model.CustomFieldOption{}).Where("field_id = ? AND value = ?", field.ID, value).Count(&count)
		if count == 0 {
			return common.Validation(fmt.Sprintf("Value '%s' is not a valid option for field '%s'", value, field.Name))
		}
	}
	if field.FieldType == "boolean" && value != "" {
		if value != "true" && value != "false" {
			return common.Validation("Boolean field must be 'true' or 'false'")
		}
	}
	return nil
}

// ==================== Field CRUD ====================

func (s *CustomFieldService) Create(workspaceID, userID uint64, req request.CustomFieldCreate) (*response.CustomFieldResponse, error) {
	if err := s.validateFieldType(req.FieldType); err != nil {
		return nil, err
	}

	f := model.CustomField{
		Name:         req.Name,
		Description:  req.Description,
		FieldType:    req.FieldType,
		IsRequired:   req.IsRequired,
		DefaultValue: req.DefaultValue,
		Placeholder:  req.Placeholder,
		IsActive:     true,
		ProjectID:    req.ProjectID,
		WorkspaceID:  workspaceID,
	}
	f.CreatedByID = &userID

	if err := s.db.Create(&f).Error; err != nil {
		return nil, common.Internal("Failed to create custom field")
	}

	return s.buildResponse(f), nil
}

func (s *CustomFieldService) List(workspaceID uint64, projectID *uint64, issueTypeID *uint64) ([]response.CustomFieldResponse, error) {
	query := s.db.Model(&model.CustomField{}).Where("workspace_id = ?", workspaceID)

	if projectID != nil {
		query = query.Where("project_id = ? OR (project_id IS NULL AND EXISTS (SELECT 1 FROM project_custom_field_enrollments WHERE project_custom_field_enrollments.field_id = custom_fields.id AND project_custom_field_enrollments.project_id = ? AND project_custom_field_enrollments.is_enabled = true))", *projectID, *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}

	if issueTypeID != nil {
		query = query.Joins("JOIN issue_type_fields ON issue_type_fields.field_id = custom_fields.id").
			Where("issue_type_fields.type_id = ?", *issueTypeID)
	}

	var fields []model.CustomField
	if err := query.Order("created_at").Find(&fields).Error; err != nil {
		return nil, common.Internal("Failed to list custom fields")
	}

	result := make([]response.CustomFieldResponse, len(fields))
	for i, f := range fields {
		result[i] = *s.buildResponse(f)
	}
	if result == nil {
		result = []response.CustomFieldResponse{}
	}
	return result, nil
}

func (s *CustomFieldService) EnrollField(projectID, fieldID uint64) error {
	var field model.CustomField
	if err := s.db.First(&field, fieldID).Error; err != nil {
		return common.NotFound("Custom field not found")
	}

	if field.ProjectID != nil {
		return common.BadRequest("Cannot enroll project-level field")
	}

	var enrollment model.ProjectCustomFieldEnrollment
	if err := s.db.Where("project_id = ? AND field_id = ?", projectID, fieldID).First(&enrollment).Error; err == nil {
		enrollment.IsEnabled = true
		return s.db.Save(&enrollment).Error
	}

	enrollment = model.ProjectCustomFieldEnrollment{
		ProjectID: projectID,
		FieldID:   fieldID,
		IsEnabled: true,
	}
	return s.db.Create(&enrollment).Error
}

func (s *CustomFieldService) UnenrollField(projectID, fieldID uint64) error {
	var field model.CustomField
	if err := s.db.First(&field, fieldID).Error; err != nil {
		return common.NotFound("Custom field not found")
	}

	if field.ProjectID != nil {
		return common.BadRequest("Cannot unenroll project-level field")
	}

	var enrollment model.ProjectCustomFieldEnrollment
	if err := s.db.Where("project_id = ? AND field_id = ?", projectID, fieldID).First(&enrollment).Error; err != nil {
		return common.NotFound("Field not enrolled for this project")
	}

	enrollment.IsEnabled = false
	return s.db.Save(&enrollment).Error
}

func (s *CustomFieldService) Get(fieldID uint64) (*response.CustomFieldResponse, error) {
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}
	return s.buildResponse(f), nil
}

func (s *CustomFieldService) Update(fieldID, userID uint64, req request.CustomFieldUpdate) (*response.CustomFieldResponse, error) {
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Description != nil {
		f.Description = *req.Description
	}
	if req.FieldType != nil {
		if err := s.validateFieldType(*req.FieldType); err != nil {
			return nil, err
		}
		// If changing away from dropdown, warn but allow
		f.FieldType = *req.FieldType
	}
	if req.IsRequired != nil {
		f.IsRequired = *req.IsRequired
	}
	if req.DefaultValue != nil {
		f.DefaultValue = *req.DefaultValue
	}
	if req.Placeholder != nil {
		f.Placeholder = *req.Placeholder
	}
	if req.IsActive != nil {
		f.IsActive = *req.IsActive
	}
	if req.ProjectID != nil {
		f.ProjectID = req.ProjectID
	}

	f.UpdatedByID = &userID
	if err := s.db.Save(&f).Error; err != nil {
		return nil, common.Internal("Failed to update custom field")
	}

	return s.buildResponse(f), nil
}

func (s *CustomFieldService) Delete(fieldID uint64) error {
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return common.NotFound("Custom field not found")
	}

	// Clean up related data
	s.db.Where("field_id = ?", fieldID).Delete(&model.IssueTypeField{})
	s.db.Where("field_id = ?", fieldID).Delete(&model.IssueCustomFieldValue{})
	s.db.Where("field_id = ?", fieldID).Delete(&model.CustomFieldOption{})

	return s.db.Delete(&f).Error
}

// ==================== Options ====================

func (s *CustomFieldService) CreateOption(fieldID uint64, req request.CustomFieldOptionCreate) (*response.CustomFieldOptionResponse, error) {
	var f model.CustomField
	if err := s.db.First(&f, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}
	if f.FieldType != "dropdown" {
		return nil, common.BadRequest("Options can only be added to dropdown fields")
	}

	opt := model.CustomFieldOption{
		FieldID:  fieldID,
		Value:    req.Value,
		Color:    req.Color,
		Sequence: req.Sequence,
	}

	// Check for duplicate value
	var count int64
	s.db.Model(&model.CustomFieldOption{}).Where("field_id = ? AND value = ?", fieldID, req.Value).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Option value already exists for this field")
	}

	if err := s.db.Create(&opt).Error; err != nil {
		return nil, common.Internal("Failed to create option")
	}

	return &response.CustomFieldOptionResponse{
		ID:       opt.ID,
		FieldID:  opt.FieldID,
		Value:    opt.Value,
		Color:    opt.Color,
		Sequence: opt.Sequence,
	}, nil
}

func (s *CustomFieldService) UpdateOption(fieldID, optionID uint64, req request.CustomFieldOptionUpdate) (*response.CustomFieldOptionResponse, error) {
	var opt model.CustomFieldOption
	if err := s.db.Where("id = ? AND field_id = ?", optionID, fieldID).First(&opt).Error; err != nil {
		return nil, common.NotFound("Option not found")
	}

	if req.Value != nil {
		opt.Value = *req.Value
	}
	if req.Color != nil {
		opt.Color = *req.Color
	}
	if req.Sequence != nil {
		opt.Sequence = *req.Sequence
	}

	if err := s.db.Save(&opt).Error; err != nil {
		return nil, common.Internal("Failed to update option")
	}

	return &response.CustomFieldOptionResponse{
		ID:       opt.ID,
		FieldID:  opt.FieldID,
		Value:    opt.Value,
		Color:    opt.Color,
		Sequence: opt.Sequence,
	}, nil
}

func (s *CustomFieldService) DeleteOption(fieldID, optionID uint64) error {
	result := s.db.Where("id = ? AND field_id = ?", optionID, fieldID).Delete(&model.CustomFieldOption{})
	if result.RowsAffected == 0 {
		return common.NotFound("Option not found")
	}
	return nil
}

// ==================== Issue Values ====================

func (s *CustomFieldService) SetIssueValue(issueID uint64, req request.IssueCustomFieldValueCreate) (*response.IssueCustomFieldValueResponse, error) {
	// Verify issue exists
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	// Verify field exists
	var field model.CustomField
	if err := s.db.First(&field, req.FieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	// Validate value against field type
	if err := s.validateFieldValue(&field, req.Value); err != nil {
		return nil, err
	}

	// Read old value before upsert
	var old model.IssueCustomFieldValue
	oldExists := s.db.Where("issue_id = ? AND field_id = ?", issueID, req.FieldID).First(&old).Error == nil

	// Upsert: delete existing, then create
	s.db.Where("issue_id = ? AND field_id = ?", issueID, req.FieldID).Delete(&model.IssueCustomFieldValue{})

	newVal := strings.TrimSpace(req.Value)
	v := model.IssueCustomFieldValue{
		IssueID: issueID,
		FieldID: req.FieldID,
		Value:   newVal,
	}

	if err := s.db.Create(&v).Error; err != nil {
		return nil, common.Internal("Failed to set custom field value")
	}

	// Record activity: "changed custom field X from A to B"
	oldStr := ""
	if oldExists { oldStr = old.Value }
	fieldLabel := field.Name
	s.db.Create(&model.IssueActivity{
		IssueID:  &issueID,
		Verb:     "updated",
		Field:    strPtr("custom_field"),
		OldValue: &oldStr,
		NewValue: &newVal,
		Comment:  &fieldLabel,
	})

	return &response.IssueCustomFieldValueResponse{
		IssueID:   issueID,
		FieldID:   req.FieldID,
		Value:     v.Value,
		FieldName: field.Name,
		FieldType: field.FieldType,
	}, nil
}

func (s *CustomFieldService) BulkSetIssueValues(req request.BulkCustomFieldValueUpdate) ([]response.IssueCustomFieldValueResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, req.IssueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	result := make([]response.IssueCustomFieldValueResponse, 0, len(req.Values))
	for _, item := range req.Values {
		// Validate field exists
		var field model.CustomField
		if err := s.db.First(&field, item.FieldID).Error; err != nil {
			continue // skip invalid field IDs
		}

		// Upsert
		s.db.Where("issue_id = ? AND field_id = ?", req.IssueID, item.FieldID).Delete(&model.IssueCustomFieldValue{})
		v := model.IssueCustomFieldValue{
			IssueID: req.IssueID,
			FieldID: item.FieldID,
			Value:   strings.TrimSpace(item.Value),
		}
		if err := s.db.Create(&v).Error; err != nil {
			continue
		}

		result = append(result, response.IssueCustomFieldValueResponse{
			IssueID:   req.IssueID,
			FieldID:   item.FieldID,
			Value:     v.Value,
			FieldName: field.Name,
			FieldType: field.FieldType,
		})
	}

	return result, nil
}

func (s *CustomFieldService) ListIssueValues(issueID uint64) ([]response.IssueCustomFieldValueResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	var values []model.IssueCustomFieldValue
	if err := s.db.Preload("Field").Where("issue_id = ?", issueID).Find(&values).Error; err != nil {
		return nil, common.Internal("Failed to list custom field values")
	}

	result := make([]response.IssueCustomFieldValueResponse, len(values))
	for i, v := range values {
		result[i] = response.IssueCustomFieldValueResponse{
			IssueID:   v.IssueID,
			FieldID:   v.FieldID,
			Value:     v.Value,
			FieldName: v.Field.Name,
			FieldType: v.Field.FieldType,
		}
	}
	if result == nil {
		result = []response.IssueCustomFieldValueResponse{}
	}
	return result, nil
}

func (s *CustomFieldService) UpdateIssueValue(issueID, fieldID uint64, val string) (*response.IssueCustomFieldValueResponse, error) {
	var field model.CustomField
	if err := s.db.First(&field, fieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	if err := s.validateFieldValue(&field, val); err != nil {
		return nil, err
	}

	var v model.IssueCustomFieldValue
	result := s.db.Where("issue_id = ? AND field_id = ?", issueID, fieldID).First(&v)
	if result.Error != nil {
		// Create if not exists
		v = model.IssueCustomFieldValue{
			IssueID: issueID,
			FieldID: fieldID,
			Value:   strings.TrimSpace(val),
		}
		if err := s.db.Create(&v).Error; err != nil {
			return nil, common.Internal("Failed to create custom field value")
		}
	} else {
		v.Value = strings.TrimSpace(val)
		if err := s.db.Save(&v).Error; err != nil {
			return nil, common.Internal("Failed to update custom field value")
		}
	}

	return &response.IssueCustomFieldValueResponse{
		IssueID:   issueID,
		FieldID:   fieldID,
		Value:     v.Value,
		FieldName: field.Name,
		FieldType: field.FieldType,
	}, nil
}

func (s *CustomFieldService) DeleteIssueValue(issueID, fieldID uint64) error {
	result := s.db.Where("issue_id = ? AND field_id = ?", issueID, fieldID).Delete(&model.IssueCustomFieldValue{})
	if result.RowsAffected == 0 {
		return common.NotFound("Custom field value not found")
	}
	return nil
}

func (s *CustomFieldService) GetIssueFieldsWithValues(issueID uint64) (*response.IssueCustomFieldsResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, common.NotFound("Issue not found")
	}

	// Get all custom fields available in this workspace (shared + project-scoped)
	var fields []model.CustomField
	s.db.Where("workspace_id = ? AND is_active = ?", issue.WorkspaceID, true).
		Where("project_id = ? OR project_id IS NULL", issue.ProjectID).
		Find(&fields)

	// Get existing values
	valueMap := make(map[uint64]string)
	var values []model.IssueCustomFieldValue
	s.db.Where("issue_id = ?", issueID).Find(&values)
	for _, v := range values {
		valueMap[v.FieldID] = v.Value
	}

	type fieldWithVal struct {
		response.CustomFieldResponse
		Value string `json:"value"`
	}

	resp := &response.IssueCustomFieldsResponse{
		IssueID: issueID,
		Fields:  make([]response.FieldWithValue, 0),
	}

	for _, f := range fields {
		cr := s.buildResponse(f)
		fwv := response.FieldWithValue{
			CustomFieldResponse: *cr,
			Value:               valueMap[f.ID],
		}
		resp.Fields = append(resp.Fields, fwv)
	}

	return resp, nil
}
