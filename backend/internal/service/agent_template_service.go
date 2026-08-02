package service

import (
	"encoding/json"
	"errors"

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

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards mutations against privilege escalation.
func (s *AgentTemplateService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage agent templates")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage agent templates")
	}
	return nil
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

func (s *AgentTemplateService) Create(wid uint64, callerID uint64, req request.AgentTemplateCreate) (*response.AgentTemplateResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
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

func (s *AgentTemplateService) Update(id uint64, callerID uint64, req request.AgentTemplateUpdate) (*response.AgentTemplateResponse, error) {
	var template model.AgentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return nil, common.NotFound("Agent template not found")
	}
	if template.WorkspaceID != nil {
		if err := s.checkWorkspaceAdmin(*template.WorkspaceID, callerID); err != nil {
			return nil, err
		}
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

func (s *AgentTemplateService) Delete(id uint64, callerID uint64) error {
	var template model.AgentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return common.NotFound("Agent template not found")
	}

	if template.IsPreset {
		return common.BadRequest("Cannot delete preset agent template")
	}

	if err := s.checkWorkspaceAdmin(*template.WorkspaceID, callerID); err != nil {
		return err
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

// ======== Preset Templates ========

// PresetAgentTemplates defines all built-in preset agent templates.
var PresetAgentTemplates = []model.AgentTemplate{
	{
		Name:        "需求分析师",
		Description: strPtr("专注于分析用户需求，提取关键信息，生成结构化的需求文档"),
		IsPreset:    true,
		Icon:        "📋",
		SystemPrompt: `你是一位专业的需求分析师。你的任务是：
1. 分析用户提供的需求描述
2. 提取关键信息和功能点
3. 识别需求中的约束和非功能需求
4. 生成清晰、结构化的需求文档

请使用专业的需求分析方法，确保需求完整、准确、可测试。`,
		SkillMode: "auto",
		Version:   "1.0",
		Status:    "active",
	},
	{
		Name:        "开发者",
		Description: strPtr("专注于代码审查、代码优化和技术实现"),
		IsPreset:    true,
		Icon:        "💻",
		SystemPrompt: `你是一位经验丰富的软件开发者。你的任务是：
1. 审查代码质量和潜在问题
2. 分析代码性能瓶颈并提供优化建议
3. 确保代码符合最佳实践和编码规范
4. 提供具体的改进方案和重构建议

请使用专业的代码分析方法，确保代码质量和可维护性。`,
		SkillMode: "auto",
		Version:   "1.0",
		Status:    "active",
	},
	{
		Name:        "测试员",
		Description: strPtr("专注于问题分类、测试用例设计和缺陷分析"),
		IsPreset:    true,
		Icon:        "🧪",
		SystemPrompt: `你是一位专业的软件测试工程师。你的任务是：
1. 分析问题描述并进行分类
2. 识别问题类型（Bug、功能请求、改进建议等）
3. 建议问题优先级和处理策略
4. 设计测试用例验证修复效果

请使用专业的测试方法，确保问题被准确识别和处理。`,
		SkillMode: "auto",
		Version:   "1.0",
		Status:    "active",
	},
	{
		Name:        "文档编写者",
		Description: strPtr("专注于文档生成、技术文档编写和内容整理"),
		IsPreset:    true,
		Icon:        "📝",
		SystemPrompt: `你是一位专业的技术文档编写者。你的任务是：
1. 分析输入内容并生成文档大纲
2. 编写结构清晰、内容完整的技术文档
3. 确保文档易于理解和使用
4. 支持多种文档格式输出

请使用专业的文档编写方法，确保文档质量和可读性。`,
		SkillMode: "auto",
		Version:   "1.0",
		Status:    "active",
	},
	{
		Name:        "敏捷教练",
		Description: strPtr("专注于会议纪要、团队协作和敏捷流程优化"),
		IsPreset:    true,
		Icon:        "🤝",
		SystemPrompt: `你是一位专业的敏捷教练。你的任务是：
1. 提取会议关键信息和讨论要点
2. 整理会议决策事项和行动项
3. 生成结构化的会议纪要
4. 跟踪行动项完成情况

请使用专业的敏捷方法，确保团队高效协作和持续改进。`,
		SkillMode: "auto",
		Version:   "1.0",
		Status:    "active",
	},
}

// InitializePresetTemplates initializes preset agent templates if they don't exist.
func (s *AgentTemplateService) InitializePresetTemplates() error {
	for _, preset := range PresetAgentTemplates {
		var existing model.AgentTemplate
		if err := s.db.Where("name = ? AND is_preset = ?", preset.Name, true).
			First(&existing).Error; err == nil {
			// Template already exists, skip
			continue
		}

		// Create new preset template
		if err := s.db.Create(&preset).Error; err != nil {
			return common.Internal("Failed to create preset agent template: " + preset.Name)
		}
	}

	return nil
}
