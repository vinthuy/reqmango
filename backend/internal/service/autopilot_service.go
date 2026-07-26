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

type AutopilotService struct{ db *gorm.DB }

func NewAutopilotService(db *gorm.DB) *AutopilotService {
	return &AutopilotService{db: db}
}

func (s *AutopilotService) CreateTask(wid uint64, req request.AutopilotTaskCreate) (*response.AutopilotTaskResponse, error) {
	task := &model.AutopilotTask{
		WorkspaceID:       wid,
		Name:              req.Name,
		Description:       req.Description,
		TriggerType:       req.TriggerType,
		CronExpression:    req.CronExpression,
		TaskType:          req.TaskType,
		AgentTemplateID:   req.AgentTemplateID,
		AgentConfigID:     req.AgentConfigID,
		TimeoutSeconds:    req.TimeoutSeconds,
		RetryCount:        req.RetryCount,
		Enabled:           req.Enabled,
	}

	if req.ProjectID != nil {
		task.ProjectID = req.ProjectID
	}
	if req.InputData != nil {
		task.InputData, _ = json.Marshal(req.InputData)
	}
	if req.Config != nil {
		task.Config, _ = json.Marshal(req.Config)
	}
	if req.NotificationConfig != nil {
		task.NotificationConfig, _ = json.Marshal(req.NotificationConfig)
	}

	// Generate trigger URL if webhook type
	if req.TriggerType == "webhook" {
		task.TriggerURL = fmt.Sprintf("/api/v1/autopilot/webhook/%d/%s", wid, generateToken())
	}

	// Calculate next run time for cron tasks
	if req.TriggerType == "cron" && req.CronExpression != "" {
		task.NextRunAt = calculateNextRun(req.CronExpression)
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return s.buildTaskResponse(task), nil
}

func (s *AutopilotService) GetTask(id uint64) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return s.buildTaskResponse(&task), nil
}

func (s *AutopilotService) ListTasks(wid uint64, projectID *uint64, status string) ([]*response.AutopilotTaskResponse, error) {
	var tasks []model.AutopilotTask
	q := s.db.Where("workspace_id = ?", wid)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	var resp []*response.AutopilotTaskResponse
	for _, t := range tasks {
		resp = append(resp, s.buildTaskResponse(&t))
	}
	return resp, nil
}

func (s *AutopilotService) UpdateTask(id uint64, req request.AutopilotTaskUpdate) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.CronExpression != nil {
		task.CronExpression = *req.CronExpression
		task.NextRunAt = calculateNextRun(*req.CronExpression)
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	if req.InputData != nil {
		task.InputData, _ = json.Marshal(req.InputData)
	}
	if req.Config != nil {
		task.Config, _ = json.Marshal(req.Config)
	}

	if err := s.db.Save(&task).Error; err != nil {
		return nil, err
	}

	return s.buildTaskResponse(&task), nil
}

func (s *AutopilotService) DeleteTask(id uint64) error {
	return s.db.Delete(&model.AutopilotTask{}, id).Error
}

func (s *AutopilotService) ToggleTask(id uint64) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	task.Enabled = !task.Enabled
	if err := s.db.Save(&task).Error; err != nil {
		return nil, err
	}
	return s.buildTaskResponse(&task), nil
}

func (s *AutopilotService) ExecuteTask(id uint64, triggerType string) (*response.AutopilotExecutionResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	if !task.Enabled {
		return nil, fmt.Errorf("task is disabled")
	}

	now := time.Now()
	exec := &model.AutopilotExecution{
		TaskID:      id,
		Status:      "running",
		TriggerType: triggerType,
		InputData:   task.InputData,
		StartedAt:   &now,
	}

	if err := s.db.Create(exec).Error; err != nil {
		return nil, err
	}

	// Update task last run time
	task.LastRunAt = &now
	if task.TriggerType == "cron" && task.CronExpression != "" {
		task.NextRunAt = calculateNextRun(task.CronExpression)
	}
	s.db.Save(&task)

	// Execute task logic (simplified)
	go s.executeTaskLogic(exec.ID, task.ID)

	return s.buildExecutionResponse(exec), nil
}

func (s *AutopilotService) executeTaskLogic(executionID, taskID uint64) {
	var task model.AutopilotTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return
	}

	// Simulate task execution
	time.Sleep(2 * time.Second)

	// Update execution status
	completedAt := time.Now()
	var startedAt time.Time
	s.db.Model(&model.AutopilotExecution{}).Select("started_at").Where("id = ?", executionID).Scan(&startedAt)
	durationMs := completedAt.Sub(startedAt).Milliseconds()
	
	s.db.Model(&model.AutopilotExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": completedAt,
		"duration_ms":  durationMs,
	})
}

func (s *AutopilotService) GetExecution(id uint64) (*response.AutopilotExecutionResponse, error) {
	var exec model.AutopilotExecution
	if err := s.db.First(&exec, id).Error; err != nil {
		return nil, err
	}
	return s.buildExecutionResponse(&exec), nil
}

func (s *AutopilotService) ListExecutions(taskID uint64) ([]*response.AutopilotExecutionResponse, error) {
	var execs []model.AutopilotExecution
	if err := s.db.Where("task_id = ?", taskID).Order("created_at DESC").Find(&execs).Error; err != nil {
		return nil, err
	}
	var resp []*response.AutopilotExecutionResponse
	for _, e := range execs {
		resp = append(resp, s.buildExecutionResponse(&e))
	}
	return resp, nil
}

func (s *AutopilotService) buildTaskResponse(task *model.AutopilotTask) *response.AutopilotTaskResponse {
	return &response.AutopilotTaskResponse{
		ID:                task.ID,
		WorkspaceID:       task.WorkspaceID,
		ProjectID:         task.ProjectID,
		Name:              task.Name,
		Description:       task.Description,
		TriggerType:       task.TriggerType,
		CronExpression:    task.CronExpression,
		TriggerURL:        task.TriggerURL,
		TaskType:          task.TaskType,
		AgentTemplateID:   task.AgentTemplateID,
		AgentConfigID:     task.AgentConfigID,
		InputData:         task.InputData,
		Status:            task.Status,
		LastRunAt:         task.LastRunAt,
		NextRunAt:         task.NextRunAt,
		Config:            task.Config,
		NotificationConfig: task.NotificationConfig,
		TimeoutSeconds:    task.TimeoutSeconds,
		RetryCount:        task.RetryCount,
		Enabled:           task.Enabled,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}

func (s *AutopilotService) buildExecutionResponse(exec *model.AutopilotExecution) *response.AutopilotExecutionResponse {
	return &response.AutopilotExecutionResponse{
		ID:          exec.ID,
		TaskID:      exec.TaskID,
		Status:      exec.Status,
		TriggerType: exec.TriggerType,
		InputData:   exec.InputData,
		OutputData:  exec.OutputData,
		ErrorInfo:   exec.ErrorInfo,
		Logs:        exec.Logs,
		StartedAt:   exec.StartedAt,
		CompletedAt: exec.CompletedAt,
		FailedAt:    exec.FailedAt,
		DurationMs:  exec.DurationMs,
		RetryCount:  exec.RetryCount,
		CreatedAt:   exec.CreatedAt,
	}
}

func generateToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())[:12]
}

func calculateNextRun(cronExpr string) *time.Time {
	// Simplified: return next hour
	next := time.Now().Add(1 * time.Hour)
	return &next
}
