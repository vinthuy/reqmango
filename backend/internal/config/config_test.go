package config

import (
	"testing"

	"gorm.io/gorm/logger"
)

func TestDBLogLevel_ExplicitSettingWins(t *testing.T) {
	cases := map[string]logger.LogLevel{
		"silent":  logger.Silent,
		"SILENT":  logger.Silent,
		" error ": logger.Error,
		"warn":    logger.Warn,
		"warning": logger.Warn,
		"info":    logger.Info,
	}
	for setting, want := range cases {
		cfg := &Config{Debug: true, DBLogLevelSetting: setting}
		if got := cfg.DBLogLevel(); got != want {
			t.Errorf("DB_LOG_LEVEL=%q: got %v, want %v", setting, got, want)
		}
	}
}

func TestDBLogLevel_FallsBackToDebugFlag(t *testing.T) {
	// A long-running / production process must not log every SQL statement, so the
	// absence of DB_LOG_LEVEL falls back to DEBUG and is quiet when DEBUG is off.
	if got := (&Config{Debug: true}).DBLogLevel(); got != logger.Info {
		t.Errorf("debug on: got %v, want Info", got)
	}
	if got := (&Config{Debug: false}).DBLogLevel(); got != logger.Warn {
		t.Errorf("debug off: got %v, want Warn", got)
	}
}

func TestDBLogLevel_UnknownSettingFallsBackToDebugFlag(t *testing.T) {
	cfg := &Config{Debug: false, DBLogLevelSetting: "verbose"}
	if got := cfg.DBLogLevel(); got != logger.Warn {
		t.Errorf("unknown DB_LOG_LEVEL: got %v, want Warn", got)
	}
}

func TestResolveSecretKey_RejectsPlaceholderAndEmpty(t *testing.T) {
	t.Setenv("SECRET_KEY", insecureDefaultSecretKey)
	placeholderKey := resolveSecretKey()
	if placeholderKey == insecureDefaultSecretKey {
		t.Fatal("placeholder SECRET_KEY must not be used for signing")
	}
	if len(placeholderKey) != 64 {
		t.Fatalf("generated key length = %d, want 64 hex chars", len(placeholderKey))
	}

	t.Setenv("SECRET_KEY", "")
	generated := resolveSecretKey()
	if generated == "" || generated == placeholderKey {
		t.Fatal("a fresh random key is expected when SECRET_KEY is unset")
	}

	t.Setenv("SECRET_KEY", "a-real-configured-secret")
	if got := resolveSecretKey(); got != "a-real-configured-secret" {
		t.Fatalf("configured SECRET_KEY = %q, want it to be used verbatim", got)
	}
}
