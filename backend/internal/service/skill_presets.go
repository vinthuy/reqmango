package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/model"
)

// PresetSkill defines a preset skill template.
type PresetSkill struct {
	Name        string
	Description string
	SkillType   string
	SkillMD     string
	Parameters  []map[string]interface{}
	Tags        []string
}

// PresetSkills contains all built-in preset skills.
var PresetSkills = []PresetSkill{
	{
		Name:        "代码审查",
		Description: "分析代码质量，识别潜在问题和改进建议",
		SkillType:   "builtin",
		Tags:        []string{"code", "review", "quality"},
		SkillMD: `## Step 1: 分析代码结构
**Tool:** analyze_code
**Input:** {"code": "{{code}}", "language": "{{language}}"}

## Step 2: 识别问题
分析代码中的潜在问题，包括：
- 代码复杂度
- 潜在的 bug
- 安全漏洞
- 性能问题

## Step 3: 生成改进建议
根据分析结果，提供具体的改进建议和优化方案。`,
		Parameters: []map[string]interface{}{
			{"name": "code", "type": "string", "required": true, "description": "要审查的代码"},
			{"name": "language", "type": "string", "required": true, "description": "代码语言"},
		},
	},
	{
		Name:        "需求分析",
		Description: "分析用户需求，提取关键信息，生成需求文档",
		SkillType:   "builtin",
		Tags:        []string{"requirement", "analysis", "document"},
		SkillMD: `## Step 1: 提取关键信息
**Tool:** extract_requirements
**Input:** {"input": "{{input}}"}

## Step 2: 分类整理
将提取的需求按照以下类别进行整理：
- 功能需求
- 非功能需求
- 约束条件

## Step 3: 生成需求文档
根据整理结果，生成结构化的需求文档，包含：
- 需求概述
- 详细需求列表
- 优先级标记`,
		Parameters: []map[string]interface{}{
			{"name": "input", "type": "string", "required": true, "description": "原始需求文本"},
		},
	},
	{
		Name:        "文档生成",
		Description: "根据内容自动生成文档，支持多种格式",
		SkillType:   "builtin",
		Tags:        []string{"documentation", "generation"},
		SkillMD: `## Step 1: 分析输入内容
**Tool:** analyze_content
**Input:** {"content": "{{content}}", "format": "{{format}}"}

## Step 2: 生成文档大纲
根据内容分析结果，生成文档大纲结构。

## Step 3: 生成文档内容
按照大纲生成完整的文档内容。`,
		Parameters: []map[string]interface{}{
			{"name": "content", "type": "string", "required": true, "description": "原始内容"},
			{"name": "format", "type": "string", "required": false, "description": "文档格式", "default": "markdown"},
		},
	},
	{
		Name:        "问题分类",
		Description: "自动分类问题，识别问题类型和优先级",
		SkillType:   "builtin",
		Tags:        []string{"issue", "classification", "triage"},
		SkillMD: `## Step 1: 分析问题描述
**Tool:** analyze_issue
**Input:** {"description": "{{description}}", "title": "{{title}}"}

## Step 2: 识别问题类型
根据分析结果，识别问题类型：
- Bug
- 功能请求
- 改进建议
- 文档问题

## Step 3: 建议优先级
根据问题的影响范围和紧急程度，建议合适的优先级。`,
		Parameters: []map[string]interface{}{
			{"name": "description", "type": "string", "required": true, "description": "问题描述"},
			{"name": "title", "type": "string", "required": true, "description": "问题标题"},
		},
	},
	{
		Name:        "代码优化",
		Description: "分析代码性能，提供优化建议和重构方案",
		SkillType:   "builtin",
		Tags:        []string{"code", "optimization", "performance"},
		SkillMD: `## Step 1: 性能分析
**Tool:** analyze_performance
**Input:** {"code": "{{code}}", "language": "{{language}}"}

## Step 2: 识别瓶颈
分析代码中的性能瓶颈和优化机会。

## Step 3: 生成优化方案
提供具体的代码优化建议和重构方案。`,
		Parameters: []map[string]interface{}{
			{"name": "code", "type": "string", "required": true, "description": "要优化的代码"},
			{"name": "language", "type": "string", "required": true, "description": "代码语言"},
		},
	},
	{
		Name:        "会议纪要",
		Description: "根据会议内容生成结构化的会议纪要",
		SkillType:   "builtin",
		Tags:        []string{"meeting", "minutes", "summary"},
		SkillMD: `## Step 1: 提取关键信息
**Tool:** extract_meeting_info
**Input:** {"transcript": "{{transcript}}"}

## Step 2: 整理要点
将会议内容整理为：
- 会议主题
- 讨论要点
- 决策事项
- 行动项

## Step 3: 生成会议纪要
根据整理结果，生成结构化的会议纪要文档。`,
		Parameters: []map[string]interface{}{
			{"name": "transcript", "type": "string", "required": true, "description": "会议记录或转录文本"},
		},
	},
}

// ToSkill converts a PresetSkill to a model.Skill.
func (p *PresetSkill) ToSkill(workspaceID uint64) model.Skill {
	paramsJSON, _ := json.Marshal(p.Parameters)
	tagsJSON, _ := json.Marshal(p.Tags)

	return model.Skill{
		Name:        p.Name,
		Description: &p.Description,
		SkillType:   p.SkillType,
		Version:     "1.0",
		Status:      "active",
		SkillMD:     p.SkillMD,
		Parameters:  paramsJSON,
		Tags:        tagsJSON,
		UsageCount:  0,
		IsShared:    true,
		IsPreset:    true,
		WorkspaceID: workspaceID,
	}
}