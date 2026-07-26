package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AgentExecutorInterface defines the interface for agent execution.
// Used to avoid circular dependency between internal/service and internal/ai/service.
type AgentExecutorInterface interface {
	DispatchAgent(agentID uint64, userID uint64, task string, ctx *AgentDispatchContext) (*AgentDispatchResult, error)
}

// AgentDispatchContext is a simplified context struct for agent dispatch.
type AgentDispatchContext struct {
	IssueID     *uint64
	ProjectID   *uint64
	WorkspaceID uint64
	TriggeredBy string
}

// AgentDispatchResult is a simplified result struct for agent dispatch.
type AgentDispatchResult struct {
	ResultSummary string
}

type SquadService struct {
	db       *gorm.DB
	agentSvc AgentExecutorInterface
}

func NewSquadService(db *gorm.DB) *SquadService {
	return &SquadService{db: db}
}

// SetAgentExecutor sets the agent executor for squad execution.
func (s *SquadService) SetAgentExecutor(agentSvc AgentExecutorInterface) {
	s.agentSvc = agentSvc
}

func (s *SquadService) Create(wid uint64, req request.SquadCreate) (*response.SquadResponse, error) {
	squad := &model.Squad{
		WorkspaceID:   wid,
		Name:          req.Name,
		Description:   req.Description,
		LeaderAgentID: req.LeaderAgentID,
		Goal:          req.Goal,
		Status:        "active",
	}
	if req.Config != nil {
		squad.Config, _ = json.Marshal(req.Config)
	}

	if err := s.db.Create(squad).Error; err != nil {
		return nil, err
	}

	// Add members
	for _, m := range req.Members {
		member := &model.SquadMember{
			SquadID:       squad.ID,
			AgentID:       m.AgentID,
			Role:          m.Role,
			AgentConfigID: m.AgentConfigID,
			Status:        "active",
			AssignedAt:    time.Now(),
		}
		s.db.Create(member)
	}

	return s.buildResponse(squad), nil
}

func (s *SquadService) Get(id uint64) (*response.SquadResponse, error) {
	var squad model.Squad
	if err := s.db.Preload("Members").First(&squad, id).Error; err != nil {
		return nil, err
	}
	return s.buildResponse(&squad), nil
}

func (s *SquadService) List(wid uint64, projectID *uint64) ([]*response.SquadResponse, error) {
	var squads []model.Squad
	q := s.db.Where("workspace_id = ?", wid)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	}
	if err := q.Preload("Members").Find(&squads).Error; err != nil {
		return nil, err
	}
	var resp []*response.SquadResponse
	for _, sq := range squads {
		resp = append(resp, s.buildResponse(&sq))
	}
	return resp, nil
}

func (s *SquadService) Update(id uint64, req request.SquadUpdate) (*response.SquadResponse, error) {
	var squad model.Squad
	if err := s.db.First(&squad, id).Error; err != nil {
		return nil, err
	}

	if req.Name != nil {
		squad.Name = *req.Name
	}
	if req.Description != nil {
		squad.Description = *req.Description
	}
	if req.LeaderAgentID != nil {
		squad.LeaderAgentID = req.LeaderAgentID
	}
	if req.Goal != nil {
		squad.Goal = *req.Goal
	}
	if req.Config != nil {
		squad.Config, _ = json.Marshal(req.Config)
	}

	if err := s.db.Save(&squad).Error; err != nil {
		return nil, err
	}

	return s.buildResponse(&squad), nil
}

func (s *SquadService) Delete(id uint64) error {
	return s.db.Delete(&model.Squad{}, id).Error
}

func (s *SquadService) AddMember(squadID uint64, req request.SquadMemberAdd) (*response.SquadMemberResponse, error) {
	var squad model.Squad
	if err := s.db.First(&squad, squadID).Error; err != nil {
		return nil, err
	}

	member := &model.SquadMember{
		SquadID:       squadID,
		AgentID:       req.AgentID,
		Role:          req.Role,
		AgentConfigID: req.AgentConfigID,
		Status:        "active",
		AssignedAt:    time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}

	return s.buildMemberResponse(member), nil
}

func (s *SquadService) RemoveMember(squadID, memberID uint64) error {
	return s.db.Model(&model.SquadMember{}).Where("squad_id = ? AND id = ?", squadID, memberID).Update("status", "removed").Error
}

func (s *SquadService) StartExecution(squadID uint64, req request.SquadExecutionStart) (*response.SquadExecutionResponse, error) {
	var squad model.Squad
	if err := s.db.Preload("Members").First(&squad, squadID).Error; err != nil {
		return nil, err
	}

	exec := &model.SquadExecution{
		SquadID:   squadID,
		Status:    "running",
		Goal:      req.Goal,
		StartedAt: &time.Time{},
	}
	*exec.StartedAt = time.Now()

	if req.InputData != nil {
		exec.InputData, _ = json.Marshal(req.InputData)
	}

	if err := s.db.Create(exec).Error; err != nil {
		return nil, err
	}

	// Track execution logs
	var logs []string
	var outputs []string
	failedCount := 0

	// Execute tasks for each member
	for _, member := range squad.Members {
		if member.Role == "observer" {
			continue
		}

		// Create task record
		taskStartedAt := time.Now()
		task := &model.SquadTask{
			SquadID:   squadID,
			MemberID:  member.ID,
			Status:    "running",
			Priority:  "medium",
			Progress:  0,
			StartedAt: &taskStartedAt,
		}
		if err := s.db.Create(task).Error; err != nil {
			return nil, err
		}

		logs = append(logs, fmt.Sprintf("[%s] Agent %d (%s) started task", time.Now().Format("15:04:05"), member.AgentID, member.Role))

		// Call agent service to execute task
		var result string
		if s.agentSvc != nil {
			taskDescription := fmt.Sprintf("作为团队成员（角色：%s），请执行以下目标：%s", member.Role, req.Goal)
			dispatchResult, err := s.agentSvc.DispatchAgent(member.AgentID, req.UserID, taskDescription, &AgentDispatchContext{
				WorkspaceID: squad.WorkspaceID,
				TriggeredBy: "squad",
			})
			if err != nil {
				logs = append(logs, fmt.Sprintf("[%s] Agent %d (%s) failed: %v", time.Now().Format("15:04:05"), member.AgentID, member.Role, err))
				task.Status = "failed"
				task.Feedback = err.Error()
				failedCount++
			} else {
				result = dispatchResult.ResultSummary
				logs = append(logs, fmt.Sprintf("[%s] Agent %d (%s) completed", time.Now().Format("15:04:05"), member.AgentID, member.Role))
				task.Status = "completed"
				task.Progress = 100
				task.Feedback = result
				outputs = append(outputs, result)
			}
		} else {
			// Fallback when no agent service is available
			result = fmt.Sprintf("Agent %d (%s) would execute: %s", member.AgentID, member.Role, req.Goal)
			logs = append(logs, fmt.Sprintf("[%s] Agent %d (%s) executed (stub)", time.Now().Format("15:04:05"), member.AgentID, member.Role))
			task.Status = "completed"
			task.Progress = 100
			task.Feedback = result
			outputs = append(outputs, result)
		}

		// Set completed time for task
		taskCompletedAt := time.Now()
		task.CompletedAt = &taskCompletedAt

		// Update task status
		if err := s.db.Save(task).Error; err != nil {
			return nil, err
		}
	}

	// Update execution with logs and output
	logsJSON, _ := json.Marshal(logs)
	outputsJSON, _ := json.Marshal(outputs)
	exec.Logs = logsJSON
	exec.OutputData = outputsJSON
	
	// Set execution status based on failure count
	if failedCount > 0 {
		exec.Status = "partial_failed"
	} else {
		exec.Status = "completed"
	}
	
	completedAt := time.Now()
	exec.CompletedAt = &completedAt

	if err := s.db.Save(exec).Error; err != nil {
		return nil, err
	}

	return s.buildExecutionResponse(exec), nil
}

func (s *SquadService) GetExecution(id uint64) (*response.SquadExecutionResponse, error) {
	var exec model.SquadExecution
	if err := s.db.First(&exec, id).Error; err != nil {
		return nil, err
	}
	return s.buildExecutionResponse(&exec), nil
}

func (s *SquadService) ListExecutions(squadID uint64) ([]*response.SquadExecutionResponse, error) {
	var execs []model.SquadExecution
	if err := s.db.Where("squad_id = ?", squadID).Order("created_at DESC").Find(&execs).Error; err != nil {
		return nil, err
	}
	var resp []*response.SquadExecutionResponse
	for _, e := range execs {
		resp = append(resp, s.buildExecutionResponse(&e))
	}
	return resp, nil
}

func (s *SquadService) buildResponse(squad *model.Squad) *response.SquadResponse {
	resp := &response.SquadResponse{
		ID:             squad.ID,
		WorkspaceID:    squad.WorkspaceID,
		Name:           squad.Name,
		Description:    squad.Description,
		LeaderAgentID:  squad.LeaderAgentID,
		Status:         squad.Status,
		Goal:           squad.Goal,
		CreatedAt:      squad.CreatedAt,
		UpdatedAt:      squad.UpdatedAt,
	}
	// Include members
	var members []response.SquadMemberResponse
	for _, m := range squad.Members {
		members = append(members, *s.buildMemberResponse(&m))
	}
	resp.Members = members
	return resp
}

func (s *SquadService) buildMemberResponse(member *model.SquadMember) *response.SquadMemberResponse {
	return &response.SquadMemberResponse{
		ID:             member.ID,
		SquadID:        member.SquadID,
		AgentID:        member.AgentID,
		Role:           member.Role,
		AgentConfigID:  member.AgentConfigID,
		Status:         member.Status,
		AssignedAt:     member.AssignedAt,
	}
}

func (s *SquadService) buildExecutionResponse(exec *model.SquadExecution) *response.SquadExecutionResponse {
	return &response.SquadExecutionResponse{
		ID:           exec.ID,
		SquadID:      exec.SquadID,
		Status:       exec.Status,
		Goal:         exec.Goal,
		InputData:    exec.InputData,
		OutputData:   exec.OutputData,
		Logs:         exec.Logs,
		StartedAt:    exec.StartedAt,
		CompletedAt:  exec.CompletedAt,
		FailedAt:     exec.FailedAt,
		ErrorInfo:    exec.ErrorInfo,
		CreatedAt:    exec.CreatedAt,
	}
}
