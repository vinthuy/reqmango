package model

import "time"

// AgentCostBudget represents project-level AI cost budget configuration.
//
// ProjectID uses `unique`, not `uniqueIndex`: the SQL migration declares a
// single-column UNIQUE constraint (…_project_id_key), and GORM's migrator
// compares that constraint against the `Unique` tag setting. Declaring
// `uniqueIndex` instead makes GORM try to drop a constraint named
// uni_<table>_<column>, which never exists, and AutoMigrate aborts — a fresh
// install would fail to start.
type AgentCostBudget struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ProjectID      uint64     `gorm:"not null;unique" json:"project_id"`
	MonthlyBudget  float64    `json:"monthly_budget"`                  // monthly budget limit
	CurrentCost    float64    `json:"current_cost"`                    // current used cost
	AlertThreshold float64    `json:"alert_threshold"`                 // alert threshold (percentage)
	AutoBlock      bool       `gorm:"default:false" json:"auto_block"` // block when over budget
	LastResetAt    *time.Time `json:"last_reset_at"`                   // last reset time
}

func (AgentCostBudget) TableName() string {
	return "agent_cost_budgets"
}

// AgentSLA represents agent execution SLA configuration.
// ProjectID uses `unique` for the same reason as AgentCostBudget above.
type AgentSLA struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ProjectID      uint64    `gorm:"not null;unique" json:"project_id"`
	NormalTaskMax  int       `gorm:"default:1800" json:"normal_task_max"`  // normal task max time (seconds)
	ComplexTaskMax int       `gorm:"default:7200" json:"complex_task_max"` // complex task max time (seconds)
	AutoEscalation bool      `gorm:"default:true" json:"auto_escalation"`  // auto escalate on timeout
	Enabled        bool      `gorm:"default:true" json:"enabled"`
}

func (AgentSLA) TableName() string {
	return "agent_slas"
}
