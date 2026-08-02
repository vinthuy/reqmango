package service

import (
	"time"

	"gorm.io/gorm"
)

// AgentSLAService manages agent execution SLA configuration.
type AgentSLAService struct {
	db *gorm.DB
}

// NewAgentSLAService creates a new AgentSLAService.
func NewAgentSLAService(db *gorm.DB) *AgentSLAService {
	return &AgentSLAService{db: db}
}

// SLAResponse represents the SLA configuration in API response.
type SLAResponse struct {
	ID             uint64 `json:"id"`
	ProjectID      uint64 `json:"project_id"`
	NormalTaskMax  int    `json:"normal_task_max"`  // seconds
	ComplexTaskMax int    `json:"complex_task_max"` // seconds
	AutoEscalation bool   `json:"auto_escalation"`
	Enabled        bool   `json:"enabled"`
}

// UpdateSLARequest represents the request to update SLA configuration.
type UpdateSLARequest struct {
	NormalTaskMax  *int  `json:"normal_task_max"`
	ComplexTaskMax *int  `json:"complex_task_max"`
	AutoEscalation *bool `json:"auto_escalation"`
	Enabled        *bool `json:"enabled"`
}

// Get returns the SLA configuration for a project.
func (s *AgentSLAService) Get(projectID uint64) (*SLAResponse, error) {
	var sla struct {
		ID             uint64 `json:"id"`
		ProjectID      uint64 `json:"project_id"`
		NormalTaskMax  int    `json:"normal_task_max"`
		ComplexTaskMax int    `json:"complex_task_max"`
		AutoEscalation bool   `json:"auto_escalation"`
		Enabled        bool   `json:"enabled"`
	}

	err := s.db.Table("agent_slas").
		Where("project_id = ?", projectID).
		First(&sla).Error

	if err == gorm.ErrRecordNotFound {
		// Create default SLA
		sla = struct {
			ID             uint64 `json:"id"`
			ProjectID      uint64 `json:"project_id"`
			NormalTaskMax  int    `json:"normal_task_max"`
			ComplexTaskMax int    `json:"complex_task_max"`
			AutoEscalation bool   `json:"auto_escalation"`
			Enabled        bool   `json:"enabled"`
		}{
			ProjectID:      projectID,
			NormalTaskMax:  1800,  // 30 minutes
			ComplexTaskMax: 7200,  // 2 hours
			AutoEscalation: true,
			Enabled:        true,
		}
		err = s.db.Table("agent_slas").Create(&sla).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &SLAResponse{
		ID:             sla.ID,
		ProjectID:      sla.ProjectID,
		NormalTaskMax:  sla.NormalTaskMax,
		ComplexTaskMax: sla.ComplexTaskMax,
		AutoEscalation: sla.AutoEscalation,
		Enabled:        sla.Enabled,
	}, nil
}

// Update updates the SLA configuration for a project.
func (s *AgentSLAService) Update(projectID uint64, req UpdateSLARequest) error {
	updates := map[string]interface{}{}

	if req.NormalTaskMax != nil {
		updates["normal_task_max"] = *req.NormalTaskMax
	}
	if req.ComplexTaskMax != nil {
		updates["complex_task_max"] = *req.ComplexTaskMax
	}
	if req.AutoEscalation != nil {
		updates["auto_escalation"] = *req.AutoEscalation
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.Table("agent_slas").
		Where("project_id = ?", projectID).
		Updates(updates)

	if result.RowsAffected == 0 {
		// Create if not exists
		sla := map[string]interface{}{
			"project_id":      projectID,
			"normal_task_max":  1800,
			"complex_task_max": 7200,
			"auto_escalation": true,
			"enabled":         true,
		}
		for k, v := range updates {
			sla[k] = v
		}
		return s.db.Table("agent_slas").Create(sla).Error
	}

	return result.Error
}

// CheckSLA checks if a task has breached its SLA.
func (s *AgentSLAService) CheckSLA(projectID uint64, taskID uint64, taskType string) (bool, error) {
	var sla struct {
		NormalTaskMax  int  `json:"normal_task_max"`
		ComplexTaskMax int  `json:"complex_task_max"`
		AutoEscalation bool `json:"auto_escalation"`
		Enabled        bool `json:"enabled"`
	}

	err := s.db.Table("agent_slas").
		Where("project_id = ?", projectID).
		First(&sla).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // no SLA configured
		}
		return false, err
	}

	if !sla.Enabled {
		return false, nil
	}

	// Get task start time and duration
	var task struct {
		StartedAt *time.Time `json:"started_at"`
	}

	err = s.db.Raw("SELECT started_at FROM agent_tasks WHERE id = ?", taskID).Scan(&task).Error
	if err != nil || task.StartedAt == nil {
		return false, nil
	}

	elapsed := time.Since(*task.StartedAt).Seconds()

	// Check based on task type
	maxTime := sla.NormalTaskMax
	if taskType == "complex" {
		maxTime = sla.ComplexTaskMax
	}

	return float64(maxTime) < elapsed, nil
}

// StartMonitoring starts the SLA monitoring goroutine.
func (s *AgentSLAService) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.checkRunningTasks()
		}
	}()
}

// checkRunningTasks checks all running tasks for SLA breach.
func (s *AgentSLAService) checkRunningTasks() {
	var tasks []struct {
		ID        uint64 `json:"id"`
		ProjectID uint64 `json:"project_id"`
		StartedAt *time.Time `json:"started_at"`
	}

	s.db.Raw(`
		SELECT id, project_id, started_at 
		FROM agent_tasks 
		WHERE status = 'running' AND started_at IS NOT NULL
	`).Scan(&tasks)

	for _, task := range tasks {
		if task.StartedAt == nil {
			continue
		}

		elapsed := time.Since(*task.StartedAt).Seconds()

		// Get SLA for project
		var sla struct {
			NormalTaskMax  int  `json:"normal_task_max"`
			AutoEscalation bool `json:"auto_escalation"`
		}

		err := s.db.Table("agent_slas").
			Where("project_id = ?", task.ProjectID).
			First(&sla).Error

		if err != nil {
			continue
		}

		// Check if SLA breached
		if float64(sla.NormalTaskMax) < elapsed {
			// Mark task as SLA breached
			s.db.Table("agent_tasks").
				Where("id = ?", task.ID).
				Update("sla_breach", true)
		}
	}
}
