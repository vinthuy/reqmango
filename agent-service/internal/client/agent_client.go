package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AgentClient calls the main Reqmango backend's Agent API.
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
}

// AgentInfo is the minimal agent info returned by the main backend.
type AgentInfo struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// DispatchResult is the result of dispatching an agent.
type DispatchResult struct {
	ID            uint64 `json:"id"`
	ResultSummary string `json:"result_summary"`
	Action        string `json:"action"`
}

// DispatchRequest is the request body for agent dispatch.
type DispatchRequest struct {
	Task       string  `json:"task"`
	IssueID    *uint64 `json:"issue_id,omitempty"`
	ProjectID  *uint64 `json:"project_id,omitempty"`
	TriggeredBy string `json:"triggered_by"`
}

func NewAgentClient(baseURL string) *AgentClient {
	return &AgentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListByWorkspace fetches agents for a workspace from the main backend.
func (c *AgentClient) ListByWorkspace(workspaceID uint64, token string) ([]AgentInfo, error) {
	url := fmt.Sprintf("%s/api/v1/workspaces/%d/agents", c.baseURL, workspaceID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent list failed (status %d): %s", resp.StatusCode, string(body))
	}

	var agents []AgentInfo
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("failed to decode agents: %w", err)
	}
	return agents, nil
}

// DispatchAgent dispatches an agent via the main backend's API.
func (c *AgentClient) DispatchAgent(workspaceID, agentID, userID uint64, task string, issueID, projectID *uint64, token string) (*DispatchResult, error) {
	url := fmt.Sprintf("%s/api/v1/workspaces/%d/agents/%d/dispatch", c.baseURL, workspaceID, agentID)

	reqBody := DispatchRequest{
		Task:        task,
		IssueID:     issueID,
		ProjectID:   projectID,
		TriggeredBy: "loop",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent dispatch failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result DispatchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode dispatch result: %w", err)
	}
	return &result, nil
}
