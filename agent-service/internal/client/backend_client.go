// agent-service/internal/client/backend_client.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BackendClient calls the main Reqmango backend for project/issue/user data.
type BackendClient struct {
	baseURL    string
	httpClient *http.Client
}

// IssueInfo is the minimal issue data returned by the main backend.
type IssueInfo struct {
	ID                uint64 `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	Priority          string `json:"priority"`
	ProjectID         uint64 `json:"project_id"`
	WorkspaceID       uint64 `json:"workspace_id"`
	SequenceID        int    `json:"sequence_id"`
	ProjectIdentifier string `json:"project_identifier"`
}

// ProjectInfo is the minimal project data.
type ProjectInfo struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	WorkspaceID uint64 `json:"workspace_id"`
}

// UserInfo is the minimal user data.
type UserInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *BackendClient) get(token, path string, result interface{}) error {
	req, _ := http.NewRequest("GET", c.baseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *BackendClient) post(token, path string, reqBody, result interface{}) error {
	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// GetIssue fetches a single issue by ID.
func (c *BackendClient) GetIssue(issueID uint64, token string) (*IssueInfo, error) {
	var issue IssueInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/issues/%d", issueID), &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetProject fetches a single project by ID.
func (c *BackendClient) GetProject(projectID uint64, token string) (*ProjectInfo, error) {
	var project ProjectInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/projects/%d", projectID), &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetUser fetches a single user by ID.
func (c *BackendClient) GetUser(userID uint64, token string) (*UserInfo, error) {
	var user UserInfo
	if err := c.get(token, fmt.Sprintf("/api/internal/users/%d", userID), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SearchIssues searches issues via RQL.
func (c *BackendClient) SearchIssues(token, query string) ([]IssueInfo, error) {
	var issues []IssueInfo
	if err := c.post(token, "/api/internal/issues/search", map[string]string{"query": query}, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}
