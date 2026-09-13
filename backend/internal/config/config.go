package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// insecureDefaultSecretKey is the placeholder shipped by older configs and docs.
// Signing tokens with it (or with an empty string) would let anyone forge a valid
// access token, so it is treated as "not configured".
const insecureDefaultSecretKey = "change-me-in-production"

type Config struct {
	DatabaseURL          string
	SecretKey            string
	AccessTokenExpireMin int
	Port                 string
	Debug                bool
	AIAPIKey             string
	AIProvider           string
	AIModel              string
	AIBaseURL            string
	RateLimitRequests    int
	RateLimitWindowSec   int
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
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable"),
		SecretKey:            resolveSecretKey(),
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

// resolveSecretKey returns the JWT signing secret.
//
// A missing or placeholder SECRET_KEY is replaced by a cryptographically random
// per-process key instead of a publicly known default: a predictable signing key
// lets an attacker mint tokens for any user. Callers that need tokens to survive
// a restart (production) must set SECRET_KEY explicitly.
func resolveSecretKey() string {
	if configured := getEnv("SECRET_KEY", ""); configured != "" && configured != insecureDefaultSecretKey {
		return configured
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// Fail closed: never fall back to a predictable signing key.
		panic(fmt.Sprintf("config: cannot generate SECRET_KEY: %v", err))
	}

	fmt.Println("[SECURITY WARNING] SECRET_KEY is not set (or is still the placeholder value). " +
		"A random signing key was generated for this process; all existing tokens are invalid and will " +
		"be invalidated again on restart. Set SECRET_KEY to a long random value (e.g. `openssl rand -hex 32`) " +
		"before deploying.")
	return hex.EncodeToString(key)
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
