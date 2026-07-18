package config

import "os"

type Config struct {
	Port           string
	DatabaseURL    string
	MainBackendURL string // URL of the main Reqmango backend for agent API calls
	SecretKey      string // JWT secret (shared with main backend)
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
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
