package service

import (
	"context"
	"testing"
)

func TestParseIntakeSpec(t *testing.T) {
	fenced := "```json\n{\"summary\":\"S\",\"acceptance_criteria\":[\"a\"],\"suggested_priority\":\"high\"}\n```"
	d := parseIntakeSpec(fenced)
	if d == nil || d.Summary != "S" || d.SuggestedPriority != "high" {
		t.Fatalf("fenced JSON not parsed: %+v", d)
	}

	d = parseIntakeSpec(`{"summary":"S","acceptance_criteria":["a"],"suggested_priority":"p0"}`)
	if d == nil || d.SuggestedPriority != "" {
		t.Fatalf("unknown priority should be dropped: %+v", d)
	}

	for _, bad := range []string{"", "not json", `{"summary":"S"}`, `{"acceptance_criteria":["a"]}`} {
		if parseIntakeSpec(bad) != nil {
			t.Fatalf("expected nil for %q", bad)
		}
	}
}

func TestDraftIntakeSpecFallsBackToTemplate(t *testing.T) {
	d := (&AIService{}).DraftIntakeSpec(context.Background(), "导出 Excel 失败", "点击导出后报错", "")
	if d.Source != "template" {
		t.Fatalf("expected template source, got %q", d.Source)
	}
	if d.SuggestedPriority != "high" || len(d.AcceptanceCriteria) == 0 || d.Summary == "" {
		t.Fatalf("template draft incomplete: %+v", d)
	}
}
