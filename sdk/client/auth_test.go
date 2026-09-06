package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["email"] != "a@b.c" || body["password"] != "pw" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt", "token_type": "Bearer", "expires_at": "2026-09-13T00:00:00Z"})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "")
	tok, err := c.Login(context.Background(), "a@b.c", "pw")
	if err != nil || tok.AccessToken != "jwt" {
		t.Fatalf("Login: %v %+v", err, tok)
	}
}

func TestCreatePAT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/tokens" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer jwt" {
			t.Errorf("expected JWT auth, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"token": "reqmango_pat_x", "id": 3, "name": "cli",
			"token_prefix": "reqmango_pat_ab3d", "created_at": "2026-09-06T00:00:00Z",
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "jwt")
	resp, err := c.CreatePAT(context.Background(), CreatePATRequest{Name: "cli"})
	if err != nil || resp.Token != "reqmango_pat_x" || resp.ID != 3 {
		t.Fatalf("CreatePAT: %v %+v", err, resp)
	}
}
