package client

import (
	"fmt"

	aiservice "github.com/reqmango/backend/internal/ai/service"
)

// AgentClient calls the AgentService directly (merged, no separate microservice).
type AgentClient struct {
	agentSvc *aiservice.AgentService
}

// NewAgentClient creates an AgentClient backed by the local AgentService.
func NewAgentClient(agentSvc *aiservice.AgentService) *AgentClient {
	return &AgentClient{agentSvc: agentSvc}
}

// DispatchAgent dispatches a task to an agent (direct call to AgentService).
func (c *AgentClient) DispatchAgent(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, triggeredBy string) error {
	_, err := c.agentSvc.DispatchAgent(agentID, userID, task, &aiservice.DispatchContext{
		IssueID:     issueID,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		TriggeredBy: triggeredBy,
	})
	return err
}

// HandleMention triggers an agent mention response (direct call to AgentService).
func (c *AgentClient) HandleMention(workspaceID, agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) error {
	if issueID == nil {
		return fmt.Errorf("agent mention requires an issue context")
	}
	_, err := c.agentSvc.HandleMention(agentID, commentID, userID, commentBody, issueName, issueID)
	return err
}

// DispatchAgentWithResult dispatches a task to an agent and returns the
// human-readable result summary. Used by ChatService to capture agent replies
// as chat messages.
func (c *AgentClient) DispatchAgentWithResult(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, triggeredBy string) (string, error) {
	act, err := c.agentSvc.DispatchAgent(agentID, userID, task, &aiservice.DispatchContext{
		IssueID:     issueID,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		TriggeredBy: triggeredBy,
	})
	if err != nil {
		return "", err
	}
	if act == nil {
		return "", nil
	}
	return act.ResultSummary, nil
}
