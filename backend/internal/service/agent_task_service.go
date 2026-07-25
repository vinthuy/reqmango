package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type AgentTaskService struct {
	db        *gorm.DB
	toolSvc   *ToolService
}

func NewAgentTaskService(db *gorm.DB) *AgentTaskService {
	return &AgentTaskService{
		db:      db,
		toolSvc: NewToolService(db),
	}
}

func (s *AgentTaskService) Create(wid uint64, req request.AgentTaskCreate) (*response.AgentTaskResponse, error) {
	task := model.AgentTask{
		Title:            req.Title,
		Description:      req.Description,
		Status:           "enqueue",
		Priority:         req.Priority,
		Progress:         0,
		TaskType:         req.TaskType,
		InputData:        req.InputData,
		AgentTemplateID:  req.AgentTemplateID,
		AgentConfigID:    req.AgentConfigID,
		WorkspaceID:      wid,
		ProjectID:        req.ProjectID,
		IssueID:          req.IssueID,
		EnqueuedAt:       time.Now(),
		EstimatedTime:    req.EstimatedTime,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, common.Internal("Failed to create agent task")
	}

	resp, err := s.Get(task.ID)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.created", resp)
	}
	return resp, err
}

func (s *AgentTaskService) Get(id uint64) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	return s.toResponse(&task), nil
}

func (s *AgentTaskService) List(wid uint64, status string) ([]response.AgentTaskResponse, error) {
	var tasks []model.AgentTask
	query := s.db.Where("workspace_id = ?", wid)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Order("enqueued_at DESC").Find(&tasks)

	res := make([]response.AgentTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		res = append(res, *s.toResponse(&t))
	}

	return res, nil
}

func (s *AgentTaskService) Update(id uint64, req request.AgentTaskUpdate) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Progress != nil {
		task.Progress = *req.Progress
	}
	if req.OutputData != nil {
		task.OutputData = *req.OutputData
	}
	if req.ErrorInfo != nil {
		task.ErrorInfo = req.ErrorInfo
	}

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to update agent task")
	}

	resp, err := s.Get(id)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.updated", resp)
	}
	return resp, err
}

func (s *AgentTaskService) Delete(id uint64) error {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return common.NotFound("Agent task not found")
	}

	return s.db.Delete(&task).Error
}

func (s *AgentTaskService) Claim(id uint64, runtimeID uint64) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	if task.Status != "enqueue" {
		return nil, common.BadRequest("Task is not available for claiming")
	}

	task.Status = "claimed"
	if runtimeID > 0 {
		task.RuntimeID = &runtimeID
		s.db.Model(&model.Runtime{}).Where("id = ?", runtimeID).Update("current_load", gorm.Expr("current_load + 1"))
	}
	now := time.Now()
	task.ClaimedAt = &now

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to claim task")
	}

	return s.Get(id)
}

func (s *AgentTaskService) Start(id uint64) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	if task.Status != "claimed" {
		return nil, common.BadRequest("Task must be claimed before starting")
	}

	task.Status = "running"
	now := time.Now()
	task.StartedAt = &now

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to start task")
	}

	resp, err := s.Get(id)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.started", resp)
	}

	// Execute tool if specified in input data
	go s.executeTaskTool(id, task.WorkspaceID, task.InputData)

	return resp, err
}

// executeTaskTool executes the tool specified in task input data
func (s *AgentTaskService) executeTaskTool(taskID, wid uint64, inputData json.RawMessage) {
	var input map[string]interface{}
	if err := json.Unmarshal(inputData, &input); err != nil {
		return
	}

	// Check if tool_id is specified in input data
	toolIDVal, ok := input["tool_id"]
	if !ok {
		return
	}

	toolID, ok := toolIDVal.(float64)
	if !ok {
		return
	}

	// Extract tool parameters
	toolParams, _ := input["tool_params"].(map[string]interface{})
	paramsBytes, _ := json.Marshal(toolParams)

	// Execute the tool
	callReq := request.CallToolRequest{
		ToolID:      uint64(toolID),
		InputParams: paramsBytes,
	}

	result, err := s.toolSvc.Call(wid, callReq)
	if err != nil {
		s.AddLog(taskID, "error", fmt.Sprintf("Tool execution failed: %v", err), nil)
		// Mark task as failed if tool execution fails
		s.Fail(taskID, request.AgentTaskFail{ErrorInfo: fmt.Sprintf("Tool execution failed: %v", err)})
		return
	}

	// Log tool execution result
	s.AddLog(taskID, "info", fmt.Sprintf("Tool executed: %s", result.ToolName), nil)

	// Update task with output
	outputBytes, _ := json.Marshal(result.OutputResult)
	s.db.Model(&model.AgentTask{}).Where("id = ?", taskID).Update("output_data", outputBytes)

	// Mark task as completed
	completeReq := request.AgentTaskComplete{
		OutputData: outputBytes,
		ActualTime: int(result.DurationMs),
	}
	s.Complete(taskID, completeReq)
}

func (s *AgentTaskService) Complete(id uint64, req request.AgentTaskComplete) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	if task.Status != "running" {
		return nil, common.BadRequest("Task must be running to complete")
	}

	task.Status = "completed"
	task.OutputData = req.OutputData
	task.Progress = 100
	now := time.Now()
	task.CompletedAt = &now
	task.ActualTime = &req.ActualTime

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to complete task")
	}

	if task.RuntimeID != nil {
		s.db.Model(&model.Runtime{}).Where("id = ?", *task.RuntimeID).Update("current_load", gorm.Expr("current_load - 1"))
	}

	resp, err := s.Get(id)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.completed", resp)
	}
	return resp, err
}

func (s *AgentTaskService) Fail(id uint64, req request.AgentTaskFail) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	task.Status = "failed"
	task.ErrorInfo = &req.ErrorInfo
	now := time.Now()
	task.CompletedAt = &now

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to mark task as failed")
	}

	if task.RuntimeID != nil {
		s.db.Model(&model.Runtime{}).Where("id = ?", *task.RuntimeID).Update("current_load", gorm.Expr("current_load - 1"))
	}

	resp, err := s.Get(id)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.failed", resp)
	}
	return resp, err
}

func (s *AgentTaskService) Cancel(id uint64) (*response.AgentTaskResponse, error) {
	var task model.AgentTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, common.NotFound("Agent task not found")
	}

	if task.Status == "completed" || task.Status == "failed" {
		return nil, common.BadRequest("Cannot cancel completed or failed task")
	}

	task.Status = "cancelled"
	now := time.Now()
	task.CancelledAt = &now

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to cancel task")
	}

	if task.RuntimeID != nil {
		s.db.Model(&model.Runtime{}).Where("id = ?", *task.RuntimeID).Update("current_load", gorm.Expr("current_load - 1"))
	}

	resp, err := s.Get(id)
	if err == nil && resp != nil {
		s.pushTaskEvent("agent_task.cancelled", resp)
	}
	return resp, err
}

func (s *AgentTaskService) AddLog(taskID uint64, level, message string, metadata []byte) error {
	log := model.TaskLog{
		TaskID:   taskID,
		Level:    level,
		Message:  message,
		Metadata: metadata,
	}

	return s.db.Create(&log).Error
}

func (s *AgentTaskService) GetLogs(taskID uint64) ([]response.TaskLogResponse, error) {
	var logs []model.TaskLog
	s.db.Where("task_id = ?", taskID).Order("created_at ASC").Find(&logs)

	res := make([]response.TaskLogResponse, 0, len(logs))
	for _, l := range logs {
		res = append(res, response.TaskLogResponse{
			ID:        l.ID,
			TaskID:    l.TaskID,
			Level:     l.Level,
			Message:   l.Message,
			Metadata:  l.Metadata,
			CreatedAt: l.CreatedAt,
		})
	}

	return res, nil
}

func (s *AgentTaskService) toResponse(t *model.AgentTask) *response.AgentTaskResponse {
	return &response.AgentTaskResponse{
		ID:              t.ID,
		Title:           t.Title,
		Description:     t.Description,
		Status:          t.Status,
		Priority:        t.Priority,
		Progress:        t.Progress,
		TaskType:        t.TaskType,
		InputData:       t.InputData,
		OutputData:      t.OutputData,
		ErrorInfo:       t.ErrorInfo,
		Logs:            t.Logs,
		AgentTemplateID: t.AgentTemplateID,
		AgentConfigID:   t.AgentConfigID,
		RuntimeID:       t.RuntimeID,
		WorkspaceID:     t.WorkspaceID,
		ProjectID:       t.ProjectID,
		IssueID:         t.IssueID,
		EnqueuedAt:      t.EnqueuedAt,
		ClaimedAt:       t.ClaimedAt,
		StartedAt:       t.StartedAt,
		CompletedAt:     t.CompletedAt,
		CancelledAt:     t.CancelledAt,
		EstimatedTime:   t.EstimatedTime,
		ActualTime:      t.ActualTime,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func (s *AgentTaskService) pushTaskEvent(event string, task *response.AgentTaskResponse) {
	data, _ := json.Marshal(task)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}
