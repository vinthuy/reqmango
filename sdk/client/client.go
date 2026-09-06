package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL points at a local reqmango backend.
const DefaultBaseURL = "http://localhost:8000/api/v1"

// Client is the shared HTTP client for the reqmango REST API.
// It only depends on the standard library.
type Client struct {
	baseURL string // no trailing slash
	token   string
	hc      *http.Client
}

// New creates a client. Empty baseURL uses DefaultBaseURL.
func New(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		hc:      &http.Client{Timeout: 30 * time.Second},
	}
}

// BaseURL returns the configured base URL (for config display).
func (c *Client) BaseURL() string { return c.baseURL }

// do performs one request. On 2xx with JSON body and non-nil out it decodes
// into out and returns response headers. On failure it returns *APIError.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) (http.Header, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: map[string]any{}}
		if len(data) > 0 {
			var m map[string]any
			if json.Unmarshal(data, &m) == nil {
				apiErr.Body = m
				if msg, ok := m["message"].(string); ok {
					apiErr.Message = msg
				}
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(data))
		}
		return nil, apiErr
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return resp.Header, nil
}

// GetJSON performs GET and decodes the JSON body into out.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) (http.Header, error) {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}

// PostJSON performs POST with a JSON body and decodes the response into out.
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error) {
	return c.do(ctx, http.MethodPost, path, query, body, out)
}

// PutJSON performs PUT with a JSON body and decodes the response into out.
func (c *Client) PutJSON(ctx context.Context, path string, query url.Values, body, out any) (http.Header, error) {
	return c.do(ctx, http.MethodPut, path, query, body, out)
}

// DeleteJSON performs DELETE. out may be nil.
func (c *Client) DeleteJSON(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.do(ctx, http.MethodDelete, path, query, nil, out)
	return err
}
