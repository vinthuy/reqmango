package config

import "os"

type Config struct {
	Port           string
	DatabaseURL    string
	MainBackendURL string // URL of the main Reqmango backend for agent API calls
	SecretKey      string // JWT secret (shared with main backend)
	AIAPIKey       string
	AIModel        string
	AIBaseURL      string
	AIProvider     string
}

func Load() *Config {
	port := os.Getenv("AGENT_SERVICE_PORT")
	if port == "" {
		port = "8001"
	}
	return &Config{
		Port:           port,
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		MainBackendURL: getEnv("MAIN_BACKEND_URL", "http://localhost:8000"),
		SecretKey:      getEnv("SECRET_KEY", "change-me-in-production"),
		AIAPIKey:       os.Getenv("AI_API_KEY"),
		AIModel:        getEnv("AI_MODEL", "deepseek-chat"),
		AIBaseURL:      getEnv("AI_BASE_URL", "https://api.deepseek.com/v1"),
		AIProvider:     getEnv("AI_PROVIDER", "deepseek"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
