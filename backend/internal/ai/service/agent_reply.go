package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/model"
)

const agentSummaryMaxRunes = 500

// agentFailureReply is posted instead of the raw LLM error, which may contain endpoint details.
const agentFailureReply = "⚠️ 这次处理失败了，请稍后重试；详细原因可在 Agent 活动记录中查看。"

// agentActivitySummary is the human-readable result stored on an AgentActivity.
func agentActivitySummary(content string, executedTools []string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		if len(executedTools) > 0 {
			return fmt.Sprintf("执行了 %d 个工具调用", len(executedTools))
		}
		return "未产生输出"
	}
	r := []rune(content)
	if len(r) > agentSummaryMaxRunes {
		return string(r[:agentSummaryMaxRunes]) + "..."
	}
	return content
}

// toolErrorJSON encodes a tool failure so the model always receives valid JSON.
func toolErrorJSON(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}

func agentIssueContextBlock(issueID uint64, ic IssueAnalysisContext) string {
	return fmt.Sprintf("\n\n你正在处理的工作项（issue_id=%d），无需再调用工具查询其基本信息：\n%s\n", issueID, issueFactsBlock(ic))
}

// loadDispatchIssue loads the issue an agent is working on, if any.
func (s *AgentService) loadDispatchIssue(issueID *uint64) (*model.Issue, *IssueAnalysisContext) {
	if issueID == nil || s.aiSvc == nil {
		return nil, nil
	}
	var issue model.Issue
	if err := s.db.First(&issue, *issueID).Error; err != nil {
		return nil, nil
	}
	var project model.Project
	identifier := ""
	if s.db.Select("identifier").First(&project, issue.ProjectID).Error == nil {
		identifier = project.Identifier
	}
	ic := s.aiSvc.loadIssueAnalysisContext(&issue, identifier)
	return &issue, &ic
}

// threadRootID resolves the top-level comment of a thread; comments are only one level deep.
func (s *AgentService) threadRootID(commentID *uint64) *uint64 {
	if commentID == nil {
		return nil
	}
	var c model.Comment
	if err := s.db.Select("id", "parent_id").First(&c, *commentID).Error; err != nil {
		return nil
	}
	if c.ParentID != nil {
		return c.ParentID
	}
	return &c.ID
}

// postAgentComment writes the agent's reply into the issue's comment thread.
// It bypasses CommentService on purpose so agent replies never re-trigger mentions.
// suggestions, when present, ride along on the comment so the UI can render
// one-click apply actions for the changes the agent proposed.
func (s *AgentService) postAgentComment(agent *model.Agent, issueID uint64, parentID *uint64, body string, suggestions []model.IssueSuggestion) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}
	c := model.Comment{
		IssueID:     issueID,
		AgentID:     &agent.ID,
		Body:        body,
		ParentID:    parentID,
		Suggestions: encodeSuggestions(suggestions),
	}
	s.db.Create(&c)
}
