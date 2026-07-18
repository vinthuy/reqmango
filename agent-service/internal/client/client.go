package client

import (
	"net/http"
	"time"
)

// AgentClient communicates with the main Reqmango backend.
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAgentClient creates a new client to the main backend.
func NewAgentClient(baseURL string) *AgentClient {
	return &AgentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
