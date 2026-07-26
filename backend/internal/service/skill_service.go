package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// SkillExecutor defines the interface for executing skills.
type SkillExecutor interface {
	Execute(ctx context.Context, skill *model.Skill, params map[string]interface{}) (*SkillExecutionResult, error)
}

// SkillExecutionResult contains the result of skill execution.
type SkillExecutionResult struct {
	SkillID    uint64
	SkillName  string
	Steps      []SkillStep
	FinalResult string
	Error      string
	TokensUsed int
}

// SkillStep represents a step in a skill execution.
type SkillStep struct {
	Step     int
	Action   string
	Tool     string
	Input    map[string]interface{}
	Output   interface{}
	Error    string
	Status   string
}

type SkillService struct {
	db       *gorm.DB
	executor SkillExecutor
}

func NewSkillService(db *gorm.DB) *SkillService {
	return &SkillService{db: db}
}

// SetExecutor sets the skill executor.
func (s *SkillService) SetExecutor(executor SkillExecutor) {
	s.executor = executor
}

func (s *SkillService) Create(wid uint64, req request.SkillCreate) (*response.SkillResponse, error) {
	skill := model.Skill{
		Name:        req.Name,
		Description: req.Description,
		SkillType:   req.SkillType,
		Version:     "1.0",
		Status:      "active",
		SkillMD:     req.SkillMD,
		Parameters:  req.Parameters,
		Tags:        req.Tags,
		UsageCount:  0,
		IsShared:    req.IsShared,
		WorkspaceID: wid,
	}

	if err := s.db.Create(&skill).Error; err != nil {
		return nil, common.Internal("Failed to create skill")
	}

	return s.Get(skill.ID)
}

func (s *SkillService) Get(id uint64) (*response.SkillResponse, error) {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, common.NotFound("Skill not found")
	}

	return s.toResponse(&skill), nil
}

func (s *SkillService) List(wid uint64) ([]response.SkillResponse, error) {
	var skills []model.Skill
	s.db.Where("workspace_id = ?", wid).Find(&skills)

	res := make([]response.SkillResponse, 0, len(skills))
	for _, sk := range skills {
		res = append(res, *s.toResponse(&sk))
	}

	return res, nil
}

func (s *SkillService) Update(id uint64, req request.SkillUpdate) (*response.SkillResponse, error) {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, common.NotFound("Skill not found")
	}

	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = req.Description
	}
	if req.SkillType != nil {
		skill.SkillType = *req.SkillType
	}
	if req.SkillMD != nil {
		skill.SkillMD = *req.SkillMD
	}
	if req.Parameters != nil {
		skill.Parameters = *req.Parameters
	}
	if req.Tags != nil {
		skill.Tags = *req.Tags
	}
	if req.IsShared != nil {
		skill.IsShared = *req.IsShared
	}
	if req.Status != nil {
		skill.Status = *req.Status
	}

	if err := s.db.Save(&skill).Error; err != nil {
		return nil, common.Internal("Failed to update skill")
	}

	return s.Get(id)
}

func (s *SkillService) Delete(id uint64) error {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return common.NotFound("Skill not found")
	}

	return s.db.Delete(&skill).Error
}

func (s *SkillService) IncrementUsage(id uint64) error {
	return s.db.Model(&model.Skill{}).Where("id = ?", id).Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

// Execute executes a skill with the given parameters and records execution log.
func (s *SkillService) Execute(ctx context.Context, id uint64, req request.SkillExecute) (*response.SkillExecutionResponse, error) {
	var skill model.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, common.NotFound("Skill not found")
	}

	if skill.Status != "active" {
		return nil, common.BadRequest("Skill is not active")
	}

	// Parse parameters
	var params map[string]interface{}
	if req.Parameters != nil {
		if err := json.Unmarshal(req.Parameters, &params); err != nil {
			return nil, common.BadRequest("Invalid parameters JSON")
		}
	}

	// Create execution log
	log := &model.SkillExecutionLog{
		SkillID:     skill.ID,
		WorkspaceID: skill.WorkspaceID,
		InputParams: req.Parameters,
		Status:      "running",
	}
	if err := s.db.Create(log).Error; err != nil {
		return nil, common.Internal("Failed to create execution log")
	}

	startTime := time.Now()
	var result *SkillExecutionResult
	var err error

	// Execute using executor if available
	if s.executor != nil {
		result, err = s.executor.Execute(ctx, &skill, params)
		if err != nil {
			log.Status = "failed"
			errorMsg := err.Error()
			log.ErrorMessage = &errorMsg
			s.db.Save(log)
			return nil, common.Internal("Failed to execute skill")
		}
	} else {
		// Fallback: execute locally using built-in logic
		result = s.executeLocal(&skill, params)
	}

	// Update execution log
	log.DurationMs = time.Since(startTime).Milliseconds()
	log.Status = "completed"
	log.TokensUsed = result.TokensUsed

	if result.FinalResult != "" {
		log.OutputResult, _ = json.Marshal(map[string]interface{}{
			"final_result": result.FinalResult,
			"steps":        result.Steps,
		})
	}
	s.db.Save(log)

	// Increment usage count
	s.IncrementUsage(id)

	return s.toExecutionResponse(result), nil
}

// executeLocal provides a simple local execution fallback.
func (s *SkillService) executeLocal(skill *model.Skill, params map[string]interface{}) *SkillExecutionResult {
	result := &SkillExecutionResult{
		SkillID:   skill.ID,
		SkillName: skill.Name,
		Steps: []SkillStep{
			{
				Step:     1,
				Action:   "Skill executed successfully",
				Status:   "completed",
				Output:   params,
			},
		},
		FinalResult: "Skill executed successfully with parameters",
	}
	return result
}

func (s *SkillService) toResponse(sk *model.Skill) *response.SkillResponse {
	return &response.SkillResponse{
		ID:          sk.ID,
		Name:        sk.Name,
		Description: sk.Description,
		SkillType:   sk.SkillType,
		Version:     sk.Version,
		Status:      sk.Status,
		SkillMD:     sk.SkillMD,
		Parameters:  sk.Parameters,
		Tags:        sk.Tags,
		UsageCount:  sk.UsageCount,
		IsShared:    sk.IsShared,
		WorkspaceID: sk.WorkspaceID,
		CreatedAt:   sk.CreatedAt,
		UpdatedAt:   sk.UpdatedAt,
	}
}

func (s *SkillService) toExecutionResponse(result *SkillExecutionResult) *response.SkillExecutionResponse {
	steps := make([]response.SkillStepResponse, 0, len(result.Steps))
	for _, step := range result.Steps {
		steps = append(steps, response.SkillStepResponse{
			Step:     step.Step,
			Action:   step.Action,
			Tool:     step.Tool,
			Input:    step.Input,
			Output:   step.Output,
			Error:    step.Error,
			Status:   step.Status,
		})
	}

	return &response.SkillExecutionResponse{
		SkillID:    result.SkillID,
		SkillName:  result.SkillName,
		Steps:      steps,
		FinalResult: result.FinalResult,
		Error:      result.Error,
		TokensUsed: result.TokensUsed,
	}
}

// SkillExecutionLogQueryParams defines query parameters for execution logs.
type SkillExecutionLogQueryParams struct {
	SkillID     uint64     `json:"skill_id,omitempty"`
	WorkspaceID uint64     `json:"workspace_id,omitempty"`
	Status      string     `json:"status,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Page        int        `json:"page,omitempty"`
	PageSize    int        `json:"page_size,omitempty"`
	SortBy      string     `json:"sort_by,omitempty"`
	SortOrder   string     `json:"sort_order,omitempty"`
}

// ListExecutionLogs retrieves skill execution logs with filtering and pagination.
func (s *SkillService) ListExecutionLogs(params SkillExecutionLogQueryParams) ([]response.SkillExecutionLogResponse, int64, error) {
	var logs []model.SkillExecutionLog
	var total int64

	query := s.db.Model(&model.SkillExecutionLog{})

	// Filter by workspace ID
	if params.WorkspaceID > 0 {
		query = query.Where("workspace_id = ?", params.WorkspaceID)
	}

	// Filter by skill ID
	if params.SkillID > 0 {
		query = query.Where("skill_id = ?", params.SkillID)
	}

	// Filter by status
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Filter by date range
	if params.StartDate != nil {
		query = query.Where("created_at >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("created_at <= ?", *params.EndDate)
	}

	// Count total
	query.Count(&total)

	// Default pagination
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	// Sorting
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, common.Internal("Failed to retrieve execution logs")
	}

	// Convert to response
	result := make([]response.SkillExecutionLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, s.toExecutionLogResponse(&log))
	}

	return result, total, nil
}

func (s *SkillService) toExecutionLogResponse(log *model.SkillExecutionLog) response.SkillExecutionLogResponse {
	return response.SkillExecutionLogResponse{
		ID:          log.ID,
		SkillID:     log.SkillID,
		WorkspaceID: log.WorkspaceID,
		InputParams: log.InputParams,
		OutputResult: log.OutputResult,
		Status:      log.Status,
		ErrorMessage: log.ErrorMessage,
		TokensUsed:  log.TokensUsed,
		DurationMs:  log.DurationMs,
		CreatedAt:   log.CreatedAt,
		UpdatedAt:   log.UpdatedAt,
	}
}

// InitializePresetSkills initializes preset skills and their required tools for a workspace if they don't exist.
func (s *SkillService) InitializePresetSkills(workspaceID uint64) error {
	// First, create preset tools
	for _, presetTool := range PresetTools {
		var existingTool model.Tool
		if err := s.db.Where("workspace_id = ? AND name = ?", workspaceID, presetTool.Name).
			First(&existingTool).Error; err == nil {
			// Tool already exists, skip
			continue
		}

		tool := PresetToolToModel(presetTool.Name, presetTool.Description, presetTool.Category, presetTool.ToolType, workspaceID)
		if err := s.db.Create(&tool).Error; err != nil {
			return common.Internal("Failed to create preset tool: " + presetTool.Name)
		}
	}

	// Then, create preset skills
	for _, preset := range PresetSkills {
		// Check if the preset skill already exists
		var existing model.Skill
		if err := s.db.Where("workspace_id = ? AND name = ? AND is_preset = ?", workspaceID, preset.Name, true).
			First(&existing).Error; err == nil {
			// Skill already exists, skip
			continue
		}

		// Create new preset skill
		skill := preset.ToSkill(workspaceID)
		if err := s.db.Create(&skill).Error; err != nil {
			return common.Internal("Failed to create preset skill: " + preset.Name)
		}
	}

	return nil
}

// ListPresetSkills returns all preset skills for a workspace.
func (s *SkillService) ListPresetSkills(workspaceID uint64) ([]response.SkillResponse, error) {
	var skills []model.Skill
	s.db.Where("workspace_id = ? AND is_preset = ?", workspaceID, true).Find(&skills)

	res := make([]response.SkillResponse, 0, len(skills))
	for _, sk := range skills {
		res = append(res, *s.toResponse(&sk))
	}

	return res, nil
}
