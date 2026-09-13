package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRedactQuery_HidesCredentials(t *testing.T) {
	cases := []struct {
		name      string
		rawQuery  string
		mustHide  []string
		mustKeep  []string
		wantValue string
	}{
		{
			name:      "sse token parameter",
			rawQuery:  "token=eyJhbGciOiJIUzI1NiJ9.secret.signature&project_id=7",
			mustHide:  []string{"eyJhbGciOiJIUzI1NiJ9.secret.signature"},
			mustKeep:  []string{"project_id=7"},
			wantValue: "REDACTED",
		},
		{
			name:      "api key parameter",
			rawQuery:  "api_key=super-secret-value",
			mustHide:  []string{"super-secret-value"},
			wantValue: "REDACTED",
		},
		{
			name:     "harmless query is untouched",
			rawQuery: "page=2&per_page=50",
			mustKeep: []string{"page=2", "per_page=50"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactQuery(tc.rawQuery)

			for _, secret := range tc.mustHide {
				if strings.Contains(got, secret) {
					t.Fatalf("RedactQuery(%q) = %q, still contains %q", tc.rawQuery, got, secret)
				}
			}
			for _, keep := range tc.mustKeep {
				if !strings.Contains(got, keep) {
					t.Fatalf("RedactQuery(%q) = %q, expected to keep %q", tc.rawQuery, got, keep)
				}
			}
			if tc.wantValue != "" && !strings.Contains(got, tc.wantValue) {
				t.Fatalf("RedactQuery(%q) = %q, expected %q marker", tc.rawQuery, got, tc.wantValue)
			}
		})
	}
}

// TestLogger_DoesNotLeakTokenInQuery proves the access log never receives a raw
// credential passed as a query parameter (the SSE/chat authentication path).
func TestLogger_DoesNotLeakTokenInQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(original)

	router := gin.New()
	router.Use(Logger())
	router.GET("/stream", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// Synthetic, non-functional JWT assembled at runtime: it only has to look like a
	// credential to prove the value never reaches the log.
	token := strings.Join([]string{"eyJhbGciOiJIUzI1NiJ9", "leaked-jwt", "signature"}, ".")
	req := httptest.NewRequest(http.MethodGet, "/stream?token="+token, nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	logged := buf.String()
	if strings.Contains(logged, token) {
		t.Fatalf("access log leaked the token: %s", logged)
	}
	if !strings.Contains(logged, "REDACTED") {
		t.Fatalf("access log did not mark the redacted parameter: %s", logged)
	}
}
