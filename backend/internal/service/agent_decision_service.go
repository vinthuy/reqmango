package service

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AgentDecisionService manages agent decision records.
type AgentDecisionService struct {
	db *gorm.DB
}

// NewAgentDecisionService creates a new AgentDecisionService.
func NewAgentDecisionService(db *gorm.DB) *AgentDecisionService {
	return &AgentDecisionService{db: db}
}

// AgentDecisionRecord represents a decision record to be saved.
type AgentDecisionRecord struct {
	AgentID       uint64   `json:"agent_id"`
	IssueID       *uint64  `json:"issue_id"`
	AgentTaskID   *uint64  `json:"agent_task_id"`
	WorkflowRunID *uint64  `json:"workflow_run_id"`
	NodeType      string   `json:"node_type"`
	Thinking      string   `json:"thinking"`
	Decision      string   `json:"decision"`
	Reasoning     string   `json:"reasoning"`
	Alternatives  []string `json:"alternatives"`
	Confidence    float64  `json:"confidence"`
}

// AgentDecisionResponse represents a decision record in API response.
type AgentDecisionResponse struct {
	ID            uint64   `json:"id"`
	AgentID       uint64   `json:"agent_id"`
	AgentName     string   `json:"agent_name"`
	IssueID       *uint64  `json:"issue_id"`
	AgentTaskID   *uint64  `json:"agent_task_id"`
	WorkflowRunID *uint64  `json:"workflow_run_id"`
	NodeType      string   `json:"node_type"`
	Thinking      string   `json:"thinking"`
	Decision      string   `json:"decision"`
	Reasoning     string   `json:"reasoning"`
	Alternatives  []string `json:"alternatives"`
	Confidence    float64  `json:"confidence"`
	CreatedAt     time.Time `json:"created_at"`
}

// Record saves a decision record.
func (s *AgentDecisionService) Record(record *AgentDecisionRecord) error {
	alternativesJSON, _ := json.Marshal(record.Alternatives)

	decision := map[string]interface{}{
		"agent_id":        record.AgentID,
		"issue_id":        record.IssueID,
		"agent_task_id":   record.AgentTaskID,
		"workflow_run_id": record.WorkflowRunID,
		"node_type":       record.NodeType,
		"thinking":        record.Thinking,
		"decision":        record.Decision,
		"reasoning":       record.Reasoning,
		"alternatives":    alternativesJSON,
		"confidence":      record.Confidence,
	}

	return s.db.Table("agent_decisions").Create(decision).Error
}

// ListByIssue returns all decision records for an issue.
func (s *AgentDecisionService) ListByIssue(issueID uint64) ([]AgentDecisionResponse, error) {
	var records []struct {
		ID            uint64    `json:"id"`
		AgentID       uint64    `json:"agent_id"`
		IssueID       *uint64   `json:"issue_id"`
		AgentTaskID   *uint64   `json:"agent_task_id"`
		WorkflowRunID *uint64   `json:"workflow_run_id"`
		NodeType      string    `json:"node_type"`
		Thinking      string    `json:"thinking"`
		Decision      string    `json:"decision"`
		Reasoning     string    `json:"reasoning"`
		Alternatives  []byte    `json:"alternatives"`
		Confidence    float64   `json:"confidence"`
		CreatedAt     time.Time `json:"created_at"`
	}

	err := s.db.Table("agent_decisions").
		Where("issue_id = ?", issueID).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	var result []AgentDecisionResponse
	for _, r := range records {
		resp := AgentDecisionResponse{
			ID:            r.ID,
			AgentID:       r.AgentID,
			IssueID:       r.IssueID,
			AgentTaskID:   r.AgentTaskID,
			WorkflowRunID: r.WorkflowRunID,
			NodeType:      r.NodeType,
			Thinking:      r.Thinking,
			Decision:      r.Decision,
			Reasoning:     r.Reasoning,
			Confidence:    r.Confidence,
			CreatedAt:     r.CreatedAt,
		}

		// Get agent name
		var agentName string
		s.db.Raw("SELECT name FROM agents WHERE id = ?", r.AgentID).Scan(&agentName)
		resp.AgentName = agentName

		// Parse alternatives
		if r.Alternatives != nil {
			json.Unmarshal(r.Alternatives, &resp.Alternatives)
		}

		result = append(result, resp)
	}

	if result == nil {
		result = []AgentDecisionResponse{}
	}

	return result, nil
}

// ListByTask returns all decision records for a task.
func (s *AgentDecisionService) ListByTask(taskID uint64) ([]AgentDecisionResponse, error) {
	var records []struct {
		ID            uint64    `json:"id"`
		AgentID       uint64    `json:"agent_id"`
		IssueID       *uint64   `json:"issue_id"`
		AgentTaskID   *uint64   `json:"agent_task_id"`
		WorkflowRunID *uint64   `json:"workflow_run_id"`
		NodeType      string    `json:"node_type"`
		Thinking      string    `json:"thinking"`
		Decision      string    `json:"decision"`
		Reasoning     string    `json:"reasoning"`
		Alternatives  []byte    `json:"alternatives"`
		Confidence    float64   `json:"confidence"`
		CreatedAt     time.Time `json:"created_at"`
	}

	err := s.db.Table("agent_decisions").
		Where("agent_task_id = ?", taskID).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	var result []AgentDecisionResponse
	for _, r := range records {
		resp := AgentDecisionResponse{
			ID:            r.ID,
			AgentID:       r.AgentID,
			IssueID:       r.IssueID,
			AgentTaskID:   r.AgentTaskID,
			WorkflowRunID: r.WorkflowRunID,
			NodeType:      r.NodeType,
			Thinking:      r.Thinking,
			Decision:      r.Decision,
			Reasoning:     r.Reasoning,
			Confidence:    r.Confidence,
			CreatedAt:     r.CreatedAt,
		}

		// Get agent name
		var agentName string
		s.db.Raw("SELECT name FROM agents WHERE id = ?", r.AgentID).Scan(&agentName)
		resp.AgentName = agentName

		// Parse alternatives
		if r.Alternatives != nil {
			json.Unmarshal(r.Alternatives, &resp.Alternatives)
		}

		result = append(result, resp)
	}

	if result == nil {
		result = []AgentDecisionResponse{}
	}

	return result, nil
}

// ListByProject returns all decision records for a project (via issue or workflow run association).
func (s *AgentDecisionService) ListByProject(projectID uint64, limit int) ([]AgentDecisionResponse, error) {
	var records []struct {
		ID            uint64    `json:"id"`
		AgentID       uint64    `json:"agent_id"`
		IssueID       *uint64   `json:"issue_id"`
		AgentTaskID   *uint64   `json:"agent_task_id"`
		WorkflowRunID *uint64   `json:"workflow_run_id"`
		NodeType      string    `json:"node_type"`
		Thinking      string    `json:"thinking"`
		Decision      string    `json:"decision"`
		Reasoning     string    `json:"reasoning"`
		Alternatives  []byte    `json:"alternatives"`
		Confidence    float64   `json:"confidence"`
		CreatedAt     time.Time `json:"created_at"`
	}

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	err := s.db.Table("agent_decisions").
		Where(`issue_id IN (SELECT id FROM issues WHERE project_id = ?)
			OR workflow_run_id IN (SELECT id FROM workflow_runs WHERE workflow_id IN (SELECT id FROM agent_workflows WHERE project_id = ?))`,
			projectID, projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	var result []AgentDecisionResponse
	for _, r := range records {
		resp := AgentDecisionResponse{
			ID:            r.ID,
			AgentID:       r.AgentID,
			IssueID:       r.IssueID,
			AgentTaskID:   r.AgentTaskID,
			WorkflowRunID: r.WorkflowRunID,
			NodeType:      r.NodeType,
			Thinking:      r.Thinking,
			Decision:      r.Decision,
			Reasoning:     r.Reasoning,
			Confidence:    r.Confidence,
			CreatedAt:     r.CreatedAt,
		}

		var agentName string
		s.db.Raw("SELECT name FROM agents WHERE id = ?", r.AgentID).Scan(&agentName)
		resp.AgentName = agentName

		if r.Alternatives != nil {
			json.Unmarshal(r.Alternatives, &resp.Alternatives)
		}

		result = append(result, resp)
	}

	if result == nil {
		result = []AgentDecisionResponse{}
	}

	return result, nil
}
