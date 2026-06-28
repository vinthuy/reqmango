package main

import "os"

// Config holds the MCP server configuration loaded from environment variables.
type Config struct {
	// APIBaseURL is the base URL of the ReqManPy API (e.g. "http://localhost:8000/api/v1").
	APIBaseURL string
	// APIToken is the JWT Bearer token used to authenticate against the ReqManPy API.
	APIToken string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	return &Config{
		APIBaseURL: getEnv("REQMANPY_API_URL", "http://localhost:8000/api/v1"),
		APIToken:   getEnv("REQMANPY_API_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
