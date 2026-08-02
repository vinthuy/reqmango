package service

import (
	"errors"
	"fmt"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// IssueAgentService manages issue assignment to agents.
type IssueAgentService struct {
	db           *gorm.DB
	agentTaskSvc *AgentTaskService
	budgetSvc    *AgentCostBudgetService
	slaSvc       *AgentSLAService
	decisionSvc  *AgentDecisionService
}

// NewIssueAgentService creates a new IssueAgentService.
func NewIssueAgentService(
	db *gorm.DB,
	agentTaskSvc *AgentTaskService,
	budgetSvc *AgentCostBudgetService,
	slaSvc *AgentSLAService,
	decisionSvc *AgentDecisionService,
) *IssueAgentService {
	return &IssueAgentService{
		db:           db,
		agentTaskSvc: agentTaskSvc,
		budgetSvc:    budgetSvc,
		slaSvc:       slaSvc,
		decisionSvc:  decisionSvc,
	}
}

// AssignRequest represents the request to assign an issue to an agent.
type AssignRequest struct {
	AgentID  uint64  `json:"agent_id"`
	Priority string  `json:"priority"`
	Deadline *string `json:"deadline"`
	Notes    string  `json:"notes"`
}

// AgentStatusResponse represents the agent execution status for an issue.
type AgentStatusResponse struct {
	IssueID      uint64  `json:"issue_id"`
	AgentID      *uint64 `json:"agent_id"`
	AgentName    string  `json:"agent_name"`
	TaskID       *uint64 `json:"task_id"`
	TaskStatus   string  `json:"task_status"`
	TaskProgress int     `json:"task_progress"`
	StartedAt    *string `json:"started_at"`
	EstimatedEnd *string `json:"estimated_end"`
	SLABreach    bool    `json:"sla_breach"`
}

// ExecutionPreview represents a preview of what the agent will do.
type ExecutionPreview struct {
	AgentID       uint64   `json:"agent_id"`
	AgentName     string   `json:"agent_name"`
	Steps         []string `json:"steps"`
	EstimatedTime int      `json:"estimated_time"`
	EstimatedCost float64  `json:"estimated_cost"`
	Confidence    float64  `json:"confidence"`
}

// Assign assigns an issue to an agent and creates an AgentTask.
func (s *IssueAgentService) Assign(issueID uint64, req AssignRequest) (*model.AgentTask, error) {
	// 0. Look up issue's project/workspace for budget check and task attribution
	var issue struct {
		ProjectID   uint64
		WorkspaceID uint64
	}
	if err := s.db.Model(&model.Issue{}).Select("project_id, workspace_id").Where("id = ?", issueID).First(&issue).Error; err != nil {
		return nil, fmt.Errorf("issue not found: %w", err)
	}

	// 1. Check budget
	ok, msg, err := s.budgetSvc.CheckBudget(issue.ProjectID, 0.05) // estimated cost for assignment
	if err != nil {
		return nil, fmt.Errorf("budget check failed: %w", err)
	}
	if !ok {
		return nil, errors.New(msg)
	}

	// 2. Update issue
	result := s.db.Model(&model.Issue{}).
		Where("id = ?", issueID).
		Updates(map[string]interface{}{
			"agent_assignee_id": req.AgentID,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("issue not found")
	}

	// 3. Create AgentTask (workspace_id is NOT NULL)
	task := &model.AgentTask{
		Title:      fmt.Sprintf("Issue #%d", issueID),
		Status:     "enqueue",
		Priority:   req.Priority,
		IssueID:    &issueID,
		WorkspaceID: issue.WorkspaceID,
		ProjectID:  &issue.ProjectID,
	}
	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	// 4. Update issue with task ID
	s.db.Model(&model.Issue{}).Where("id = ?", issueID).Update("agent_task_id", task.ID)

	// 5. Record decision
	s.decisionSvc.Record(&AgentDecisionRecord{
		AgentID:     req.AgentID,
		IssueID:     &issueID,
		AgentTaskID: &task.ID,
		NodeType:    "issue_assignment",
		Reasoning:   req.Notes,
		Confidence:  0.9,
	})

	// 6. Record cost
	s.budgetSvc.RecordCost(issue.ProjectID, 0.001)

	return task, nil
}

// Unassign removes agent assignment from an issue.
func (s *IssueAgentService) Unassign(issueID uint64) error {
	result := s.db.Model(&model.Issue{}).
		Where("id = ?", issueID).
		Updates(map[string]interface{}{
			"agent_assignee_id": nil,
			"agent_task_id":     nil,
		})

	if result.RowsAffected == 0 {
		return errors.New("issue not found")
	}

	return result.Error
}

// GetStatus returns the agent execution status for an issue.
func (s *IssueAgentService) GetStatus(issueID uint64) (*AgentStatusResponse, error) {
	var issue struct {
		ID              uint64  `json:"id"`
		AgentAssigneeID *uint64 `json:"agent_assignee_id"`
		AgentTaskID     *uint64 `json:"agent_task_id"`
	}

	if err := s.db.Raw("SELECT id, agent_assignee_id, agent_task_id FROM issues WHERE id = ?", issueID).Scan(&issue).Error; err != nil {
		return nil, err
	}

	resp := &AgentStatusResponse{
		IssueID: issue.ID,
		AgentID: issue.AgentAssigneeID,
		TaskID:  issue.AgentTaskID,
	}

	if issue.AgentAssigneeID != nil {
		var agentName string
		s.db.Raw("SELECT name FROM agents WHERE id = ?", *issue.AgentAssigneeID).Scan(&agentName)
		resp.AgentName = agentName
	}

	if issue.AgentTaskID != nil {
		var task struct {
			Status   string `json:"status"`
			Progress int    `json:"progress"`
		}
		s.db.Raw("SELECT status, progress FROM agent_tasks WHERE id = ?", *issue.AgentTaskID).Scan(&task)
		resp.TaskStatus = task.Status
		resp.TaskProgress = task.Progress
	}

	return resp, nil
}

// PreviewExecution returns a preview of what the agent will do.
func (s *IssueAgentService) PreviewExecution(issueID, agentID uint64) (*ExecutionPreview, error) {
	var agent struct {
		Name string `json:"name"`
	}
	if err := s.db.Raw("SELECT name FROM agents WHERE id = ?", agentID).Scan(&agent).Error; err != nil {
		return nil, err
	}

	preview := &ExecutionPreview{
		AgentID:       agentID,
		AgentName:     agent.Name,
		Steps:         []string{"Analyze issue", "Plan approach", "Execute task", "Report results"},
		EstimatedTime: 1800,
		EstimatedCost: 0.05,
		Confidence:    0.85,
	}

	return preview, nil
}

// BulkAssign assigns multiple issues to an agent.
func (s *IssueAgentService) BulkAssign(issueIDs []uint64, agentID uint64) ([]*model.AgentTask, error) {
	var tasks []*model.AgentTask

	for _, issueID := range issueIDs {
		task, err := s.Assign(issueID, AssignRequest{
			AgentID:  agentID,
			Priority: "medium",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to assign issue %d: %w", issueID, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
