package main

import "testing"

func TestExitCodeGate(t *testing.T) {
	cases := []struct {
		name    string
		results []result
		want    int
	}{
		{"clean run", []result{{Verdict: "denied"}, {Verdict: "ok-empty"}, {Verdict: "WRITE-ok"}}, 0},
		{"cross tenant", []result{{Verdict: "XTENANT"}}, 1},
		{"escalation", []result{{Verdict: "ESCALATE"}}, 1},
		{"leak", []result{{Verdict: "LEAK"}}, 1},
		{"unauthorised write", []result{{Verdict: "WRITE"}}, 1},
		{"unresolved route", []result{{Verdict: "unresolved:attachmentId"}}, 1},
		{"handler panic surfaced as 500", []result{{Verdict: "status500"}}, 1},
		{"upstream and AI failures are allowed", []result{{Verdict: "status502"}, {Verdict: "status503"}}, 0},
		{
			"public catalogs are allowed",
			[]result{
				{Method: "GET", Path: "/api/v1/permissions", Verdict: "READ"},
				{Method: "GET", Path: "/api/v1/issues/import/template", Verdict: "READ"},
				{Method: "POST", Path: "/api/v1/notifications/read-all", Verdict: "WRITE"},
			},
			0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := exitCode(tc.results); got != tc.want {
				t.Fatalf("exitCode() = %d, want %d", got, tc.want)
			}
		})
	}
}
