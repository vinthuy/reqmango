package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reqmango", "config.json")

	cfg := &Config{APIURL: "http://localhost:8000/api/v1", PAT: "reqmango_pat_abc", WorkspaceID: 2, ProjectID: 5}
	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.APIURL != cfg.APIURL || got.PAT != cfg.PAT || got.WorkspaceID != 2 || got.ProjectID != 5 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}

	// Windows always reports 0666 for newly created files regardless of the
	// requested mode, so the 0600 assertion only runs on non-Windows OSes.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("config should be 0600, got %o", info.Mode().Perm())
		}
	}
}

func TestLoadConfig_Missing(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nope", "config.json"))
	if err != nil || cfg.APIURL != "" {
		t.Fatalf("expected empty config, got %+v err=%v", cfg, err)
	}
}
