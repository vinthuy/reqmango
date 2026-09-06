package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type pingResp struct {
	Pong string `json:"pong"`
}

func TestDo_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer reqmango_pat_abc" {
			t.Errorf("missing auth header, got %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/ping" || r.URL.Query().Get("a") != "1" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("X-Total-Count", "42")
		json.NewEncoder(w).Encode(pingResp{Pong: "ok"})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "reqmango_pat_abc")
	var out pingResp
	hdr, err := c.GetJSON(context.Background(), "/ping", url.Values{"a": {"1"}}, &out)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if out.Pong != "ok" || hdr.Get("X-Total-Count") != "42" {
		t.Fatalf("unexpected out=%+v hdr=%v", out, hdr)
	}
}

func TestDo_APIError_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "token expired"})
	}))
	defer srv.Close()

	c := New(srv.URL, "reqmango_pat_abc")
	var out pingResp
	_, err := c.GetJSON(context.Background(), "/ping", nil, &out)
	apiErr := AsAPIError(err)
	if apiErr == nil {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Message != "token expired" {
		t.Fatalf("unexpected APIError: %+v", apiErr)
	}
}

func TestDo_APIError_409_ExtraFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"message":       "approval_required",
			"transition_id": float64(9),
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "reqmango_pat_abc")
	_, err := c.GetJSON(context.Background(), "/ping", nil, &pingResp{})
	apiErr := AsAPIError(err)
	if apiErr == nil || apiErr.Body["transition_id"] != float64(9) {
		t.Fatalf("expected Body to carry extra fields, got %+v", apiErr)
	}
}

func TestNew_DefaultBaseURL(t *testing.T) {
	c := New("", "t")
	if c.baseURL != DefaultBaseURL {
		t.Fatalf("expected default baseURL, got %q", c.baseURL)
	}
}
