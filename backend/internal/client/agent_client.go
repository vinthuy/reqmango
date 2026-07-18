package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AgentClient calls the agent-service for agent dispatch and mentions.
type AgentClient struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

func NewAgentClient(baseURL, secretKey string) *AgentClient {
	return &AgentClient{
		baseURL:    baseURL,
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// mintToken signs a short-lived JWT for the acting user so agent-service's
// auth middleware (shared secret, "sub" claim) accepts service-to-service calls.
func (c *AgentClient) mintToken(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(userID, 10),
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(c.secretKey))
}

func (c *AgentClient) do(method, path string, userID uint64, body interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.baseURL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	token, err := c.mintToken(userID)
	if err != nil {
		return fmt.Errorf("failed to sign service token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
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
	return c.do("POST", fmt.Sprintf("/api/v1/workspaces/%d/agents/%d/dispatch", workspaceID, agentID), userID, map[string]interface{}{
		"task":         task,
		"issue_id":     issueID,
		"project_id":   projectID,
		"triggered_by": triggeredBy,
	})
}

// HandleMention triggers an agent mention response.
func (c *AgentClient) HandleMention(workspaceID, agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) error {
	if issueID == nil {
		return fmt.Errorf("agent mention requires an issue context")
	}
	return c.do("POST", fmt.Sprintf("/api/v1/issues/%d/agents/%d/mention", *issueID, agentID), userID, map[string]interface{}{
		"comment_id":   commentID,
		"comment_body": commentBody,
		"issue_name":   issueName,
	})
}
