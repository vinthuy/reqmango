package harness

import (
	"testing"
)

func TestBuildRefutePrompt(t *testing.T) {
	prompt := buildRefutePrompt("Build a login page", "<html>login form</html>")
	if len(prompt) == 0 {
		t.Fatal("expected non-empty prompt")
	}
	if !contains(prompt, "SPECIFICATION") {
		t.Error("prompt should contain SPECIFICATION section")
	}
	if !contains(prompt, "OUTPUT TO REVIEW") {
		t.Error("prompt should contain OUTPUT TO REVIEW section")
	}
	if !contains(prompt, "REFUTED") {
		t.Error("prompt should instruct to default to REFUTED")
	}
}

func TestParseVerdictPass(t *testing.T) {
	response := `{"pass": true, "issues": [], "confidence": 0.9}`
	v := parseVerdict(response, 1)
	if !v.Pass {
		t.Error("expected pass=true from response")
	}
}

func TestParseVerdictFail(t *testing.T) {
	response := `{"pass": false, "issues": [{"severity": "critical", "description": "missing error handling"}], "confidence": 0.8}`
	v := parseVerdict(response, 2)
	if v.Pass {
		t.Error("expected pass=false from response")
	}
}

func TestParseVerdictDefaultToRefuted(t *testing.T) {
	response := "The output looks generally good but could use some minor improvements."
	v := parseVerdict(response, 3)
	if v.Pass {
		t.Error("should default to refuted when no clear pass/fail indicator")
	}
}
