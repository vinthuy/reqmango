package mcp

import (
	"crypto/subtle"
	"net/http"
)

// BearerAuth wraps next and rejects requests whose Authorization header is
// not "Bearer <pat>" (spec §5.1: HTTP transport is protected by the PAT).
// Constant-time comparison avoids leaking prefix matches.
func BearerAuth(pat string, next http.Handler) http.Handler {
	want := "Bearer " + pat
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(want)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"unauthorized: missing or invalid Bearer token"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
