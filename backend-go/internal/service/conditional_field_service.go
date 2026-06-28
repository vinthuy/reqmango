package service

import (
	"encoding/json"
	"strings"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

// ConditionalFieldService handles conditional field business logic.
type ConditionalFieldService struct {
	db *gorm.DB
}

// NewConditionalFieldService creates a new ConditionalFieldService.
func NewConditionalFieldService(db *gorm.DB) *ConditionalFieldService {
	return &ConditionalFieldService{db: db}
}

// Create creates a new conditional field rule.
func (s *ConditionalFieldService) Create(req *request.ConditionalFieldCreateRequest) (*request.ConditionalFieldResponse, error) {
	// Validate field exists
	var field model.CustomField
	if err := s.db.First(&field, req.FieldID).Error; err != nil {
		return nil, common.NotFound("Custom field not found")
	}

	// Validate condition type
	validTypes := []string{"type", "state", "priority", "assignee", "label", "cycle", "module"}
	if !contains(validTypes, req.ConditionType) {
		return nil, common.BadRequest("Invalid condition type")
	}

	// Validate operator
	validOperators := []string{"equals", "not_equals", "contains", "not_contains", "is_empty", "is_not_empty"}
	if !contains(validOperators, req.Operator) {
		return nil, common.BadRequest("Invalid operator")
	}

	// Serialize condition values
	conditionValuesJSON := ""
	if len(req.ConditionValues) > 0 {
		data, _ := json.Marshal(req.ConditionValues)
		conditionValuesJSON = string(data)
	}

	cf := &model.ConditionalField{
		WorkspaceID:     req.WorkspaceID,
		FieldID:         req.FieldID,
		ConditionType:   req.ConditionType,
		Operator:        req.Operator,
		ConditionValues: conditionValuesJSON,
		IsEnabled:       req.IsEnabled,
		Priority:        req.Priority,
		Description:     req.Description,
	}

	if err := s.db.Create(cf).Error; err != nil {
		return nil, common.Internal("Failed to create conditional field")
	}

	return s.toResponse(cf), nil
}

// List returns conditional fields for a workspace or field.
func (s *ConditionalFieldService) List(workspaceID uint64, fieldID *uint64) ([]request.ConditionalFieldResponse, error) {
	var rules []model.ConditionalField
	q := s.db.Where("workspace_id = ?", workspaceID)

	if fieldID != nil {
		q = q.Where("field_id = ?", *fieldID)
	}

	if err := q.Order("priority ASC, created_at DESC").Find(&rules).Error; err != nil {
		return nil, common.Internal("Failed to fetch conditional fields")
	}

	responses := make([]request.ConditionalFieldResponse, len(rules))
	for i, r := range rules {
		responses[i] = *s.toResponse(&r)
	}

	return responses, nil
}

// Get returns a single conditional field rule.
func (s *ConditionalFieldService) Get(id, workspaceID uint64) (*request.ConditionalFieldResponse, error) {
	var cf model.ConditionalField
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&cf).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Conditional field not found")
		}
		return nil, common.Internal("Failed to fetch conditional field")
	}
	return s.toResponse(&cf), nil
}

// Update updates a conditional field rule.
func (s *ConditionalFieldService) Update(id, workspaceID uint64, req *request.ConditionalFieldUpdateRequest) (*request.ConditionalFieldResponse, error) {
	var cf model.ConditionalField
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&cf).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Conditional field not found")
		}
		return nil, common.Internal("Failed to fetch conditional field")
	}

	updates := make(map[string]interface{})

	if req.ConditionType != nil {
		validTypes := []string{"type", "state", "priority", "assignee", "label", "cycle", "module"}
		if !contains(validTypes, *req.ConditionType) {
			return nil, common.BadRequest("Invalid condition type")
		}
		updates["condition_type"] = *req.ConditionType
	}

	if req.Operator != nil {
		validOperators := []string{"equals", "not_equals", "contains", "not_contains", "is_empty", "is_not_empty"}
		if !contains(validOperators, *req.Operator) {
			return nil, common.BadRequest("Invalid operator")
		}
		updates["operator"] = *req.Operator
	}

	if req.ConditionValues != nil {
		data, _ := json.Marshal(req.ConditionValues)
		updates["condition_values"] = string(data)
	}

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if err := s.db.Model(&cf).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update conditional field")
	}

	s.db.First(&cf, id)
	return s.toResponse(&cf), nil
}

// Delete deletes a conditional field rule.
func (s *ConditionalFieldService) Delete(id, workspaceID uint64) error {
	result := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&model.ConditionalField{})
	if result.Error != nil {
		return common.Internal("Failed to delete conditional field")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Conditional field not found")
	}
	return nil
}

// EvaluateFieldVisibility evaluates which fields should be visible for an issue.
func (s *ConditionalFieldService) EvaluateFieldVisibility(workspaceID uint64, issueTypeID *uint64, stateID uint64, priority string, assigneeIDs []uint64, labelIDs []uint64) ([]uint64, error) {
	var rules []model.ConditionalField
	if err := s.db.Where("workspace_id = ? AND is_enabled = ?", workspaceID, true).
		Order("priority ASC").Find(&rules).Error; err != nil {
		return nil, common.Internal("Failed to fetch conditional fields")
	}

	// Collect field IDs that should be hidden
	hiddenFieldIDs := make(map[uint64]bool)

	for _, rule := range rules {
		values := s.parseConditionValues(rule.ConditionValues)
		matches := s.evaluateCondition(rule.ConditionType, rule.Operator, values,
			issueTypeID, stateID, priority, assigneeIDs, labelIDs)

		if matches {
			// Field should be hidden when condition matches
			hiddenFieldIDs[rule.FieldID] = true
		}
	}

	// Return field IDs that are NOT hidden
	visibleFields := make([]uint64, 0)
	var allFields []model.CustomField
	s.db.Where("workspace_id = ?", workspaceID).Find(&allFields)

	for _, field := range allFields {
		if !hiddenFieldIDs[field.ID] {
			visibleFields = append(visibleFields, field.ID)
		}
	}

	return visibleFields, nil
}

func (s *ConditionalFieldService) evaluateCondition(conditionType, operator string, conditionValues []string,
	issueTypeID *uint64, stateID uint64, priority string, assigneeIDs []uint64, labelIDs []uint64) bool {

	var fieldValue string
	var listValues []string

	switch conditionType {
	case "type":
		if issueTypeID != nil {
			fieldValue = strings.TrimSpace(strings.ToLower(string(rune(*issueTypeID))))
		}
	case "state":
		fieldValue = strings.TrimSpace(strings.ToLower(string(rune(stateID))))
	case "priority":
		fieldValue = strings.TrimSpace(strings.ToLower(priority))
	case "assignee":
		for _, id := range assigneeIDs {
			listValues = append(listValues, strings.TrimSpace(strings.ToLower(string(rune(id)))))
		}
	case "label":
		for _, id := range labelIDs {
			listValues = append(listValues, strings.TrimSpace(strings.ToLower(string(rune(id)))))
		}
	}

	switch operator {
	case "equals":
		return strings.TrimSpace(strings.ToLower(fieldValue)) == strings.TrimSpace(strings.ToLower(conditionValues[0]))
	case "not_equals":
		return strings.TrimSpace(strings.ToLower(fieldValue)) != strings.TrimSpace(strings.ToLower(conditionValues[0]))
	case "contains":
		for _, v := range conditionValues {
			for _, lv := range listValues {
				if lv == strings.TrimSpace(strings.ToLower(v)) {
					return true
				}
			}
			if strings.Contains(strings.ToLower(fieldValue), strings.ToLower(v)) {
				return true
			}
		}
		return false
	case "not_contains":
		for _, v := range conditionValues {
			for _, lv := range listValues {
				if lv == strings.TrimSpace(strings.ToLower(v)) {
					return false
				}
			}
			if strings.Contains(strings.ToLower(fieldValue), strings.ToLower(v)) {
				return false
			}
		}
		return true
	case "is_empty":
		return fieldValue == "" && len(listValues) == 0
	case "is_not_empty":
		return fieldValue != "" || len(listValues) > 0
	}

	return false
}

func (s *ConditionalFieldService) parseConditionValues(jsonStr string) []string {
	if jsonStr == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(jsonStr), &values); err != nil {
		return []string{jsonStr}
	}
	return values
}

func (s *ConditionalFieldService) toResponse(cf *model.ConditionalField) *request.ConditionalFieldResponse {
	return &request.ConditionalFieldResponse{
		ID:              cf.ID,
		WorkspaceID:     cf.WorkspaceID,
		FieldID:         cf.FieldID,
		ConditionType:   cf.ConditionType,
		Operator:        cf.Operator,
		ConditionValues: s.parseConditionValues(cf.ConditionValues),
		IsEnabled:       cf.IsEnabled,
		Priority:        cf.Priority,
		Description:     cf.Description,
		CreatedAt:       cf.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       cf.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
