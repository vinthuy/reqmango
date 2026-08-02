package model

import "time"

// AgentCostBudget represents project-level AI cost budget configuration.
type AgentCostBudget struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ProjectID      uint64     `gorm:"not null;uniqueIndex" json:"project_id"`
	MonthlyBudget  float64    `json:"monthly_budget"` // monthly budget limit
	CurrentCost    float64    `json:"current_cost"`   // current used cost
	AlertThreshold float64    `json:"alert_threshold"` // alert threshold (percentage)
	AutoBlock      bool       `gorm:"default:false" json:"auto_block"` // block when over budget
	LastResetAt    *time.Time `json:"last_reset_at"` // last reset time
}

func (AgentCostBudget) TableName() string {
	return "agent_cost_budgets"
}

// AgentSLA represents agent execution SLA configuration.
type AgentSLA struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ProjectID       uint64    `gorm:"not null;uniqueIndex" json:"project_id"`
	NormalTaskMax   int       `gorm:"default:1800" json:"normal_task_max"`  // normal task max time (seconds)
	ComplexTaskMax  int       `gorm:"default:7200" json:"complex_task_max"` // complex task max time (seconds)
	AutoEscalation  bool      `gorm:"default:true" json:"auto_escalation"`  // auto escalate on timeout
	Enabled         bool      `gorm:"default:true" json:"enabled"`
}

func (AgentSLA) TableName() string {
	return "agent_slas"
}
