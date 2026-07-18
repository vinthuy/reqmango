package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AgentClient calls the agent-service for agent dispatch and mentions.
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAgentClient(baseURL string) *AgentClient {
	return &AgentClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AgentClient) do(method, path string, body interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.baseURL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agent-service request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent-service %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DispatchAgent dispatches a task to an agent.
func (c *AgentClient) DispatchAgent(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, triggeredBy string) error {
	return c.do("POST", fmt.Sprintf("/api/v1/workspaces/%d/agents/%d/dispatch", workspaceID, agentID), map[string]interface{}{
		"task":         task,
		"issue_id":     issueID,
		"project_id":   projectID,
		"triggered_by": triggeredBy,
	})
}

// HandleMention triggers an agent mention response.
func (c *AgentClient) HandleMention(workspaceID, agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) error {
	return c.do("POST", fmt.Sprintf("/api/v1/workspaces/%d/agents/%d/mention", workspaceID, agentID), map[string]interface{}{
		"comment_id":   commentID,
		"issue_id":     issueID,
		"comment_body": commentBody,
		"issue_name":   issueName,
	})
}
