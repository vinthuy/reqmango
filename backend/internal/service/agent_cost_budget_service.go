package service

import (
	"time"

	"gorm.io/gorm"
)

// AgentCostBudgetService manages project-level AI cost budget.
type AgentCostBudgetService struct {
	db *gorm.DB
}

// NewAgentCostBudgetService creates a new AgentCostBudgetService.
func NewAgentCostBudgetService(db *gorm.DB) *AgentCostBudgetService {
	return &AgentCostBudgetService{db: db}
}

// BudgetResponse represents the budget configuration in API response.
type BudgetResponse struct {
	ID             uint64   `json:"id"`
	ProjectID      uint64   `json:"project_id"`
	MonthlyBudget  float64  `json:"monthly_budget"`
	CurrentCost    float64  `json:"current_cost"`
	AlertThreshold float64  `json:"alert_threshold"`
	AutoBlock      bool     `json:"auto_block"`
	LastResetAt    *string  `json:"last_reset_at"`
	BudgetUsage    float64  `json:"budget_usage"` // percentage
}

// UpdateBudgetRequest represents the request to update budget configuration.
type UpdateBudgetRequest struct {
	MonthlyBudget  *float64 `json:"monthly_budget"`
	AlertThreshold *float64 `json:"alert_threshold"`
	AutoBlock      *bool    `json:"auto_block"`
}

// Get returns the budget configuration for a project.
func (s *AgentCostBudgetService) Get(projectID uint64) (*BudgetResponse, error) {
	var budget struct {
		ID             uint64     `json:"id"`
		ProjectID      uint64     `json:"project_id"`
		MonthlyBudget  float64    `json:"monthly_budget"`
		CurrentCost    float64    `json:"current_cost"`
		AlertThreshold float64    `json:"alert_threshold"`
		AutoBlock      bool       `json:"auto_block"`
		LastResetAt    *time.Time `json:"last_reset_at"`
	}

	err := s.db.Table("agent_cost_budgets").
		Where("project_id = ?", projectID).
		First(&budget).Error

	if err == gorm.ErrRecordNotFound {
		// Create default budget
		budget = struct {
			ID             uint64     `json:"id"`
			ProjectID      uint64     `json:"project_id"`
			MonthlyBudget  float64    `json:"monthly_budget"`
			CurrentCost    float64    `json:"current_cost"`
			AlertThreshold float64    `json:"alert_threshold"`
			AutoBlock      bool       `json:"auto_block"`
			LastResetAt    *time.Time `json:"last_reset_at"`
		}{
			ProjectID:      projectID,
			MonthlyBudget:  100.0,   // $100 default
			CurrentCost:    0,
			AlertThreshold: 80.0,    // 80% default
			AutoBlock:      false,
		}
		err = s.db.Table("agent_cost_budgets").Create(&budget).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	resp := &BudgetResponse{
		ID:             budget.ID,
		ProjectID:      budget.ProjectID,
		MonthlyBudget:  budget.MonthlyBudget,
		CurrentCost:    budget.CurrentCost,
		AlertThreshold: budget.AlertThreshold,
		AutoBlock:      budget.AutoBlock,
		BudgetUsage:    0,
	}

	if budget.LastResetAt != nil {
		lastReset := budget.LastResetAt.Format(time.RFC3339)
		resp.LastResetAt = &lastReset
	}

	if budget.MonthlyBudget > 0 {
		resp.BudgetUsage = (budget.CurrentCost / budget.MonthlyBudget) * 100
	}

	return resp, nil
}

// Update updates the budget configuration for a project.
func (s *AgentCostBudgetService) Update(projectID uint64, req UpdateBudgetRequest) error {
	updates := map[string]interface{}{}

	if req.MonthlyBudget != nil {
		updates["monthly_budget"] = *req.MonthlyBudget
	}
	if req.AlertThreshold != nil {
		updates["alert_threshold"] = *req.AlertThreshold
	}
	if req.AutoBlock != nil {
		updates["auto_block"] = *req.AutoBlock
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.Table("agent_cost_budgets").
		Where("project_id = ?", projectID).
		Updates(updates)

	if result.RowsAffected == 0 {
		// Create if not exists
		budget := map[string]interface{}{
			"project_id":      projectID,
			"monthly_budget":  100.0,
			"alert_threshold": 80.0,
		}
		for k, v := range updates {
			budget[k] = v
		}
		return s.db.Table("agent_cost_budgets").Create(budget).Error
	}

	return result.Error
}

// CheckBudget checks if the project has budget for an operation.
func (s *AgentCostBudgetService) CheckBudget(projectID uint64, estimatedCost float64) (bool, string, error) {
	var budget struct {
		MonthlyBudget  float64 `json:"monthly_budget"`
		CurrentCost    float64 `json:"current_cost"`
		AlertThreshold float64 `json:"alert_threshold"`
		AutoBlock      bool    `json:"auto_block"`
	}

	err := s.db.Table("agent_cost_budgets").
		Where("project_id = ?", projectID).
		First(&budget).Error

	if err == gorm.ErrRecordNotFound {
		// No budget configured, allow
		return true, "", nil
	}
	if err != nil {
		return false, "", err
	}

	// Check if budget is set
	if budget.MonthlyBudget <= 0 {
		return true, "", nil
	}

	// Check if over budget
	newCost := budget.CurrentCost + estimatedCost
	if newCost > budget.MonthlyBudget && budget.AutoBlock {
		return false, "AI budget exceeded. Please increase budget or wait for next month.", nil
	}

	// Check alert threshold
	usage := (newCost / budget.MonthlyBudget) * 100
	if usage >= budget.AlertThreshold {
		return true, "AI budget usage is above " + string(rune(int('0'+int(budget.AlertThreshold)/10))) + "% threshold.", nil
	}

	return true, "", nil
}

// RecordCost records the cost of an AI operation.
func (s *AgentCostBudgetService) RecordCost(projectID uint64, cost float64) error {
	return s.db.Table("agent_cost_budgets").
		Where("project_id = ?", projectID).
		Update("current_cost", gorm.Expr("current_cost + ?", cost)).Error
}

// ResetMonthly resets the monthly cost for all projects.
func (s *AgentCostBudgetService) ResetMonthly() error {
	return s.db.Table("agent_cost_budgets").
		Updates(map[string]interface{}{
			"current_cost": 0,
			"last_reset_at": time.Now(),
		}).Error
}
