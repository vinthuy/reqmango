package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type AgentTemplateService struct{ db *gorm.DB }

func NewAgentTemplateService(db *gorm.DB) *AgentTemplateService {
	return &AgentTemplateService{db: db}
}

// validateSkillIDs validates that all referenced skill IDs exist in the workspace.
func (s *AgentTemplateService) validateSkillIDs(wid uint64, skillsJSON json.RawMessage) error {
	if skillsJSON == nil || len(skillsJSON) == 0 {
		return nil
	}

	var skillIDs []uint64
	if err := json.Unmarshal(skillsJSON, &skillIDs); err != nil {
		return common.BadRequest("Invalid available_skills format: must be an array of skill IDs")
	}

	if len(skillIDs) == 0 {
		return nil
	}

	// Check if all skill IDs exist
	var count int64
	s.db.Model(&model.Skill{}).
		Where("id IN ? AND workspace_id = ?", skillIDs, wid).
		Count(&count)

	if int(count) != len(skillIDs) {
		return common.BadRequest("Some referenced skill IDs do not exist in the workspace")
	}

	return nil
}

func (s *AgentTemplateService) Create(wid uint64, req request.AgentTemplateCreate) (*response.AgentTemplateResponse, error) {
	// Validate skill IDs
	if err := s.validateSkillIDs(wid, req.AvailableSkills); err != nil {
		return nil, err
	}

	template := model.AgentTemplate{
		Name:            req.Name,
		Description:     req.Description,
		IsPreset:        false,
		Icon:            req.Icon,
		SystemPrompt:    req.SystemPrompt,
		AvailableSkills: req.AvailableSkills,
		AvailableTools:  req.AvailableTools,
		DefaultConfig:   req.DefaultConfig,
		Version:         req.Version,
		Status:          "active",
		WorkspaceID:     &wid,
	}

	if err := s.db.Create(&template).Error; err != nil {
		return nil, common.Internal("Failed to create agent template")
	}

	return s.Get(template.ID)
}

func (s *AgentTemplateService) Get(id uint64) (*response.AgentTemplateResponse, error) {
	var template model.AgentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return nil, common.NotFound("Agent template not found")
	}

	return s.toResponse(&template), nil
}

func (s *AgentTemplateService) List(wid uint64) ([]response.AgentTemplateResponse, error) {
	var templates []model.AgentTemplate
	s.db.Where("workspace_id = ? OR is_preset = ?", wid, true).Find(&templates)

	res := make([]response.AgentTemplateResponse, 0, len(templates))
	for _, t := range templates {
		res = append(res, *s.toResponse(&t))
	}

	return res, nil
}

func (s *AgentTemplateService) Update(id uint64, req request.AgentTemplateUpdate) (*response.AgentTemplateResponse, error) {
	var template model.AgentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return nil, common.NotFound("Agent template not found")
	}

	// Validate skill IDs if provided
	if req.AvailableSkills != nil && template.WorkspaceID != nil {
		if err := s.validateSkillIDs(*template.WorkspaceID, *req.AvailableSkills); err != nil {
			return nil, err
		}
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = req.Description
	}
	if req.Icon != nil {
		template.Icon = *req.Icon
	}
	if req.SystemPrompt != nil {
		template.SystemPrompt = *req.SystemPrompt
	}
	if req.AvailableSkills != nil {
		template.AvailableSkills = *req.AvailableSkills
	}
	if req.AvailableTools != nil {
		template.AvailableTools = *req.AvailableTools
	}
	if req.DefaultConfig != nil {
		template.DefaultConfig = *req.DefaultConfig
	}
	if req.Version != nil {
		template.Version = *req.Version
	}
	if req.Status != nil {
		template.Status = *req.Status
	}

	if err := s.db.Save(&template).Error; err != nil {
		return nil, common.Internal("Failed to update agent template")
	}

	return s.Get(id)
}

func (s *AgentTemplateService) Delete(id uint64) error {
	var template model.AgentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return common.NotFound("Agent template not found")
	}

	if template.IsPreset {
		return common.BadRequest("Cannot delete preset agent template")
	}

	return s.db.Delete(&template).Error
}

func (s *AgentTemplateService) toResponse(t *model.AgentTemplate) *response.AgentTemplateResponse {
	var workspaceID *uint64
	if t.WorkspaceID != nil {
		workspaceID = t.WorkspaceID
	}

	return &response.AgentTemplateResponse{
		ID:               t.ID,
		Name:             t.Name,
		Description:      t.Description,
		IsPreset:         t.IsPreset,
		Icon:             t.Icon,
		SystemPrompt:     t.SystemPrompt,
		AvailableSkills:  t.AvailableSkills,
		AvailableTools:   t.AvailableTools,
		DefaultConfig:    t.DefaultConfig,
		Version:          t.Version,
		Status:           t.Status,
		WorkspaceID:      workspaceID,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
