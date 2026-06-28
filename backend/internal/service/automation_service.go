package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AutomationService manages automation rules.
type AutomationService struct {
	db *gorm.DB
}

func NewAutomationService(db *gorm.DB) *AutomationService {
	return &AutomationService{db: db}
}

// ======== Request/Response types ========

type AutomationCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	TriggerType string  `json:"trigger_type" binding:"required"`
	Conditions  string  `json:"conditions"`
	Actions     string  `json:"actions" binding:"required"`
	IsEnabled   *bool   `json:"is_enabled"`
	Sequence    int     `json:"sequence"`
}

type AutomationUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TriggerType *string `json:"trigger_type"`
	Conditions  *string `json:"conditions"`
	Actions     *string `json:"actions"`
	IsEnabled   *bool   `json:"is_enabled"`
	Sequence    *int    `json:"sequence"`
}

type AutomationResponse struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ProjectID      uint64 `json:"project_id"`
	TriggerType    string `json:"trigger_type"`
	Conditions     string `json:"conditions"`
	Actions        string `json:"actions"`
	IsEnabled      bool   `json:"is_enabled"`
	Sequence       int    `json:"sequence"`
	ExecutionCount int    `json:"execution_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ======== CRUD ========

func (s *AutomationService) List(projectID uint64) ([]AutomationResponse, error) {
	var rules []model.AutomationRule
	if err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&rules).Error; err != nil {
		return nil, common.Internal("Failed to list automation rules")
	}
	res := make([]AutomationResponse, len(rules))
	for i, r := range rules {
		res[i] = s.toResponse(&r)
	}
	if res == nil {
		res = []AutomationResponse{}
	}
	return res, nil
}

func (s *AutomationService) Get(id uint64) (*AutomationResponse, error) {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Automation rule not found")
		}
		return nil, common.Internal("Failed to get automation rule")
	}
	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Create(projectID uint64, req *AutomationCreateRequest) (*AutomationResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}

	rule := model.AutomationRule{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectID,
		TriggerType: req.TriggerType,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		IsEnabled:   enabled,
		Sequence:    req.Sequence,
	}
	if rule.Sequence == 0 {
		rule.Sequence = 1
	}

	// Validate JSON
	if err := s.validateJSON(req.Conditions, req.Actions); err != nil {
		return nil, common.Validation(err.Error())
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, common.Internal("Failed to create automation rule")
	}
	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Update(id uint64, req *AutomationUpdateRequest) (*AutomationResponse, error) {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Automation rule not found")
		}
		return nil, common.Internal("Failed to get automation rule")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.TriggerType != nil {
		updates["trigger_type"] = *req.TriggerType
	}
	if req.Conditions != nil {
		if err := s.validateJSON(*req.Conditions, rule.Actions); err != nil {
			return nil, common.Validation(err.Error())
		}
		updates["conditions"] = *req.Conditions
	}
	if req.Actions != nil {
		if err := s.validateJSON(rule.Conditions, *req.Actions); err != nil {
			return nil, common.Validation(err.Error())
		}
		updates["actions"] = *req.Actions
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.Sequence != nil {
		updates["sequence"] = *req.Sequence
	}

	if err := s.db.Model(&rule).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update automation rule")
	}

	s.db.First(&rule, id)
	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Delete(id uint64) error {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Automation rule not found")
		}
		return common.Internal("Failed to get automation rule")
	}
	if err := s.db.Delete(&rule).Error; err != nil {
		return common.Internal("Failed to delete automation rule")
	}
	return nil
}

// ======== Execution ========

// ExecuteTrigger runs all matching automation rules for a given trigger event.
func (s *AutomationService) ExecuteTrigger(projectID uint64, triggerType string, issueID uint64, context map[string]interface{}) []string {
	var rules []model.AutomationRule
	if err := s.db.Where("project_id = ? AND trigger_type = ? AND is_enabled = ?", projectID, triggerType, true).
		Order("sequence ASC").Find(&rules).Error; err != nil {
		return nil
	}

	var results []string
	for _, rule := range rules {
		if s.evaluateConditions(rule.Conditions, context) {
			actionResults := s.executeActions(rule.Actions, issueID, context)
			results = append(results, actionResults...)

			// Increment execution count
			s.db.Model(&rule).Update("execution_count", rule.ExecutionCount+1)
		}
	}
	return results
}

// ======== Helpers ========

func (s *AutomationService) toResponse(rule *model.AutomationRule) AutomationResponse {
	return AutomationResponse{
		ID:             rule.ID,
		Name:           rule.Name,
		Description:    rule.Description,
		ProjectID:      rule.ProjectID,
		TriggerType:    rule.TriggerType,
		Conditions:     rule.Conditions,
		Actions:        rule.Actions,
		IsEnabled:      rule.IsEnabled,
		Sequence:       rule.Sequence,
		ExecutionCount: rule.ExecutionCount,
		CreatedAt:      rule.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      rule.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (s *AutomationService) validateJSON(conditions, actions string) error {
	if conditions != "" && conditions != "[]" {
		var conds []interface{}
		if err := json.Unmarshal([]byte(conditions), &conds); err != nil {
			return fmt.Errorf("invalid conditions JSON: %w", err)
		}
	}
	if actions != "" && actions != "[]" {
		var acts []interface{}
		if err := json.Unmarshal([]byte(actions), &acts); err != nil {
			return fmt.Errorf("invalid actions JSON: %w", err)
		}
	}
	return nil
}

func (s *AutomationService) evaluateConditions(conditionsJSON string, context map[string]interface{}) bool {
	if conditionsJSON == "" || conditionsJSON == "[]" {
		return true // No conditions = always trigger
	}

	var conditions []map[string]interface{}
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		return false
	}

	for _, cond := range conditions {
		field, _ := cond["field"].(string)
		op, _ := cond["operator"].(string)
		value := cond["value"]

		contextVal, ok := context[field]
		if !ok {
			return false
		}

		switch op {
		case "equals":
			if fmt.Sprintf("%v", contextVal) != fmt.Sprintf("%v", value) {
				return false
			}
		case "not_equals":
			if fmt.Sprintf("%v", contextVal) == fmt.Sprintf("%v", value) {
				return false
			}
		case "contains":
			ctxStr := fmt.Sprintf("%v", contextVal)
			valStr := fmt.Sprintf("%v", value)
			if !strings.Contains(ctxStr, valStr) {
				return false
			}
		case "in":
			// value should be an array
			if valArr, ok := value.([]interface{}); ok {
				found := false
				ctxStr := fmt.Sprintf("%v", contextVal)
				for _, v := range valArr {
					if ctxStr == fmt.Sprintf("%v", v) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}

func (s *AutomationService) executeActions(actionsJSON string, issueID uint64, context map[string]interface{}) []string {
	if actionsJSON == "" || actionsJSON == "[]" {
		return nil
	}

	var actions []map[string]interface{}
	if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
		return nil
	}

	var results []string

	for _, action := range actions {
		actionType, _ := action["type"].(string)
		field, _ := action["field"].(string)
		value := action["value"]

		switch actionType {
		case "set_field":
			if field == "state_id" {
				if stateID, ok := toUint64(value); ok {
					s.db.Model(&model.Issue{}).Where("id = ?", issueID).Update(field, stateID)
					results = append(results, fmt.Sprintf("Set state to %d", stateID))
				}
			} else if field == "priority" {
				s.db.Model(&model.Issue{}).Where("id = ?", issueID).Update(field, value)
				results = append(results, fmt.Sprintf("Set %s to %v", field, value))
			}
		case "add_label":
			if labelID, ok := toUint64(value); ok {
				label := model.IssueLabel{IssueID: issueID, LabelID: labelID}
				s.db.Create(&label)
				results = append(results, fmt.Sprintf("Added label %d", labelID))
			}
		case "add_comment":
			if comment, ok := value.(string); ok {
				cmt := model.Comment{IssueID: issueID, Body: comment}
				s.db.Create(&cmt)
				results = append(results, "Added automated comment")
			}
		case "assign_to":
			if assigneeID, ok := toUint64(value); ok {
				assignee := model.IssueAssignee{IssueID: issueID, UserID: assigneeID}
				s.db.Create(&assignee)
				results = append(results, fmt.Sprintf("Assigned to user %d", assigneeID))
			}
		}
	}

	return results
}

func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		return uint64(val), true
	case int:
		return uint64(val), true
	case int64:
		return uint64(val), true
	case uint64:
		return val, true
	case string:
		var n uint64
		if _, err := fmt.Sscanf(val, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}

