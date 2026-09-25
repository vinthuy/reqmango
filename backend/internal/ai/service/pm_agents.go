package service

import (
	"errors"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// PMAgentDef describes an out-of-box project-management agent seeded per workspace.
type PMAgentDef struct {
	Name         string
	Avatar       string
	Capabilities []string
	SystemPrompt string
}

// PMAgentDefs returns the four out-of-box PM agent definitions (Spec §6).
func PMAgentDefs() []PMAgentDef {
	return []PMAgentDef{
		{
			Name:         "请求分诊",
			Avatar:       "🏥",
			Capabilities: []string{"triage", "search_issues", "suggest_labels"},
			SystemPrompt: `你是请求分诊 Agent。对新建/Intake 工作项：建议类型与优先级、可能重复项、标签与路由。只输出结构化建议；不擅自改库。`,
		},
		{
			Name:         "交付风险",
			Avatar:       "⚠️",
			Capabilities: []string{"search_issues", "get_cycle_progress", "add_comment"},
			SystemPrompt: `你是交付风险 Agent。分析 Cycle/Issue 的阻塞、逾期与依赖风险，用评论给出可执行建议。`,
		},
		{
			Name:         "Sprint 总结",
			Avatar:       "📊",
			Capabilities: []string{"get_cycle_progress", "search_issues", "list_cycles"},
			SystemPrompt: `你是 Sprint 总结 Agent。根据 Cycle 进度与完成情况写中文总结，可建议存为 Page。`,
		},
		{
			Name:         "Spec 草稿",
			Avatar:       "📝",
			Capabilities: []string{"get_issue", "add_comment"},
			SystemPrompt: `你是 Spec 草稿 Agent。根据 Issue 写需求提纲（背景/目标/范围/验收），输出评论或 Page 草稿。`,
		},
	}
}

// EnsurePMAgents idempotently creates the four out-of-box PM agents for a workspace.
// Matching is by stable Name within the workspace (non-deleted rows).
func (s *AgentService) EnsurePMAgents(workspaceID, userID uint64) ([]*model.Agent, error) {
	var out []*model.Agent
	for _, def := range PMAgentDefs() {
		var existing model.Agent
		err := s.db.Where("workspace_id = ? AND name = ? AND deleted_at IS NULL", workspaceID, def.Name).First(&existing).Error
		if err == nil {
			out = append(out, &existing)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		prompt := def.SystemPrompt
		perm := string(model.PermissionModePublicTo)
		vis := string(model.VisibilityWorkspace)
		ag, err := s.Create(workspaceID, userID, &AgentCreateRequest{
			Name:         def.Name,
			Avatar:       def.Avatar,
			AgentType:    "builtin",
			Capabilities: def.Capabilities,
			Status:       "active",
			SystemPrompt: &prompt,
			// model.AgentPermissionMode has no bare "public"; public_to + workspace target
			// makes the agent invocable by all workspace members.
			PermissionMode: &perm,
			Visibility:     &vis,
			InvocationTargets: []model.AgentInvocationTarget{
				{TargetType: "workspace"},
			},
		})
		if err != nil {
			return nil, err
		}
		out = append(out, ag)
	}
	return out, nil
}
