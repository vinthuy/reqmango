package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthLogin_WritesConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt", "token_type": "Bearer", "expires_at": "2026-09-13T00:00:00Z"})
		case "/api/v1/auth/tokens":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"token": "reqmango_pat_secret", "id": 1, "name": "cli"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	var out, errBuf bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{
		"auth", "login", "--config", cfgPath, "--api-url", srv.URL + "/api/v1",
		"--email", "a@b.c", "--password", "pw",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\nstderr: %s", err, errBuf.String())
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.PAT != "reqmango_pat_secret" || cfg.APIURL != srv.URL+"/api/v1" {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestWorkspaceSwitch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "Acme", "slug": "acme"}})
	}))
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := SaveConfig(cfgPath, &Config{APIURL: srv.URL + "/api/v1", PAT: "reqmango_pat_x"}); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"workspace", "switch", "2", "--config", cfgPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\nstderr: %s", err, errBuf.String())
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil || cfg.WorkspaceID != 2 {
		t.Fatalf("workspace not persisted: %+v err=%v", cfg, err)
	}
	if !strings.Contains(out.String(), "Acme") {
		t.Fatalf("expected confirmation with workspace name, got %q", out.String())
	}
}
