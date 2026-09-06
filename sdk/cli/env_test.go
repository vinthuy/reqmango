package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// workspaceListServer serves GET /api/v1/workspaces and, when wantAuth is
// non-empty, asserts the Authorization header equals "Bearer <wantAuth>".
func workspaceListServer(t *testing.T, wantAuth string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if wantAuth != "" {
			if got := r.Header.Get("Authorization"); got != "Bearer "+wantAuth {
				t.Errorf("unexpected Authorization header %q, want %q", got, "Bearer "+wantAuth)
			}
		}
		json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "Acme", "slug": "acme"}})
	}))
}

// TestEnvAPIURLFallback: REQMANGO_API_URL env is used when --api-url and the
// config value are both absent.
func TestEnvAPIURLFallback(t *testing.T) {
	srv := workspaceListServer(t, "")
	defer srv.Close()
	t.Setenv("REQMANGO_API_URL", srv.URL+"/api/v1")
	t.Setenv("REQMANGO_PAT", "")

	out, run := newCLI(t, "", &Config{APIURL: "", PAT: "reqmango_pat_cfg", WorkspaceID: 2, ProjectID: 5})
	if err := run("workspace", "list"); err != nil {
		t.Fatalf("workspace list: %v", err)
	}
	if !strings.Contains(out.String(), "Acme") {
		t.Fatalf("unexpected output %q", out.String())
	}
}

// TestEnvPATFallback: REQMANGO_PAT env is used when the config has no PAT.
func TestEnvPATFallback(t *testing.T) {
	srv := workspaceListServer(t, "reqmango_pat_env")
	defer srv.Close()
	t.Setenv("REQMANGO_API_URL", "")
	t.Setenv("REQMANGO_PAT", "reqmango_pat_env")

	out, run := newCLI(t, srv.URL+"/api/v1", &Config{APIURL: srv.URL + "/api/v1", PAT: "", WorkspaceID: 2, ProjectID: 5})
	if err := run("workspace", "list"); err != nil {
		t.Fatalf("workspace list: %v", err)
	}
	if !strings.Contains(out.String(), "Acme") {
		t.Fatalf("unexpected output %q", out.String())
	}
}

// TestEnvPATConfigWins: the config PAT wins over REQMANGO_PAT when present.
func TestEnvPATConfigWins(t *testing.T) {
	srv := workspaceListServer(t, "reqmango_pat_cfg")
	defer srv.Close()
	t.Setenv("REQMANGO_API_URL", "")
	t.Setenv("REQMANGO_PAT", "reqmango_pat_env")

	out, run := newCLI(t, srv.URL+"/api/v1", &Config{APIURL: srv.URL + "/api/v1", PAT: "reqmango_pat_cfg", WorkspaceID: 2, ProjectID: 5})
	if err := run("workspace", "list"); err != nil {
		t.Fatalf("workspace list: %v", err)
	}
	if !strings.Contains(out.String(), "Acme") {
		t.Fatalf("unexpected output %q", out.String())
	}
}
