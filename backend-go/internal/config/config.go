package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL           string
	SecretKey             string
	AccessTokenExpireMin  int
	Port                  string
	Debug                 bool
	AIAPIKey              string
	AIProvider            string
	AIModel               string
	AIBaseURL             string
	RateLimitRequests     int
	RateLimitWindowSec    int
}

func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")

	viper.AutomaticEnv()

	// Try to read .env file, ignore error if not found
	_ = viper.ReadInConfig()

	cfg := &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/reqmanpy?sslmode=disable"),
		SecretKey:            getEnv("SECRET_KEY", "change-me-in-production"),
		AccessTokenExpireMin: getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 10080),
		Port:                 getEnv("PORT", "8000"),
		Debug:                getEnvBool("DEBUG", true),
		AIAPIKey:             getEnv("AI_API_KEY", getEnv("DEEPSEEK_API_KEY", "")),
		AIProvider:           getEnv("AI_PROVIDER", "deepseek"),
		AIModel:              getEnv("AI_MODEL", "deepseek-chat"),
		AIBaseURL:            getEnv("AI_BASE_URL", "https://api.deepseek.com/v1"),
		RateLimitRequests:    getEnvInt("RATE_LIMIT_REQUESTS", 500),
		RateLimitWindowSec:   getEnvInt("RATE_LIMIT_WINDOW_SEC", 60),
	}

	fmt.Printf("Config loaded: port=%s, db_url=%s\n", cfg.Port, maskDSN(cfg.DatabaseURL))
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := viper.GetString(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		viper.Set(key, value)
	}
	viper.SetDefault(key, fallback)
	return viper.GetInt(key)
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		viper.Set(key, value)
	}
	viper.SetDefault(key, fallback)
	return viper.GetBool(key)
}

func maskDSN(dsn string) string {
	// Simple mask: just show the beginning
	if len(dsn) > 30 {
		return dsn[:30] + "..."
	}
	return dsn
}
