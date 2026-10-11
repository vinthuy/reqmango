package service

import (
	"github.com/reqmango/backend/internal/ai/common"
	"github.com/reqmango/backend/internal/model"
)

// AgentInWorkspace loads an agent only if it belongs to the workspace, so a
// route's :id cannot reach another tenant's agent.
func (s *AgentService) AgentInWorkspace(workspaceID, agentID uint64) (*model.Agent, error) {
	agent, err := s.GetByID(agentID)
	if err != nil || agent.WorkspaceID != workspaceID {
		return nil, common.NotFound("Agent not found")
	}
	return agent, nil
}

// IssueLocation returns the workspace and project an issue belongs to.
func (s *AgentService) IssueLocation(issueID uint64) (workspaceID, projectID uint64, err error) {
	var issue model.Issue
	if err := s.db.Select("id", "workspace_id", "project_id").First(&issue, issueID).Error; err != nil {
		return 0, 0, common.NotFound("Issue not found")
	}
	return issue.WorkspaceID, issue.ProjectID, nil
}

// checkDispatchTenant keeps an agent run inside the agent's workspace: the
// issue and project it is pointed at must live there, and agree with each
// other. Callers check the user's own access; this guards against ids taken
// from a request body or another workspace's automation.
func (s *AgentService) checkDispatchTenant(agent *model.Agent, ctx *DispatchContext) error {
	if ctx == nil {
		return nil
	}
	if ctx.WorkspaceID != 0 && ctx.WorkspaceID != agent.WorkspaceID {
		return common.NotFound("Agent not found")
	}
	var issueProject uint64
	if ctx.IssueID != nil {
		var issue model.Issue
		if err := s.db.Select("id", "project_id", "workspace_id").First(&issue, *ctx.IssueID).Error; err != nil ||
			issue.WorkspaceID != agent.WorkspaceID {
			return common.NotFound("Issue not found")
		}
		issueProject = issue.ProjectID
	}
	if ctx.ProjectID != nil {
		var project model.Project
		if err := s.db.Select("id", "workspace_id").First(&project, *ctx.ProjectID).Error; err != nil ||
			project.WorkspaceID != agent.WorkspaceID {
			return common.NotFound("Project not found")
		}
		if issueProject != 0 && issueProject != project.ID {
			return common.Validation("issue_id does not belong to project_id")
		}
	}
	return nil
}
