package harness

import (
	"testing"
)

const samplePipelineYAML = `
name: sprint-review
description: "Sprint end review pipeline"
trigger:
  type: cron
  schedule: "0 9 * * 5"
pipeline:
  mode: sequential
  stages:
    - name: analyze
      type: planner
      agent: sprint-analyzer
      model: claude-opus-4-8
      effort: high
    - name: review
      type: reviewer
      agent: sprint-reviewer
      model: claude-sonnet-4-6
      effort: high
      mode: adversarial
    - name: report
      type: executor
      agent: report-generator
      model: claude-sonnet-4-6
      effort: medium
retry:
  max_attempts: 3
  backoff: exponential
budget:
  max_tokens: 100000
  on_exhausted: escalate_to_human
`

func TestParsePipelineDSL(t *testing.T) {
	dsl, err := ParsePipelineDSL([]byte(samplePipelineYAML))
	if err != nil {
		t.Fatalf("ParsePipelineDSL failed: %v", err)
	}
	if dsl.Name != "sprint-review" {
		t.Errorf("expected name 'sprint-review', got '%s'", dsl.Name)
	}
	if len(dsl.Pipeline.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(dsl.Pipeline.Stages))
	}
	if dsl.Pipeline.Mode != "sequential" {
		t.Errorf("expected mode 'sequential', got '%s'", dsl.Pipeline.Mode)
	}
}

func TestParsePipelineDSLValidation(t *testing.T) {
	_, err := ParsePipelineDSL([]byte(`name: ""`))
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}

	_, err = ParsePipelineDSL([]byte(`name: test
pipeline:
  stages: []`))
	if err == nil {
		t.Fatal("expected validation error for empty stages")
	}
}

func TestToPipelineConfig(t *testing.T) {
	dsl, _ := ParsePipelineDSL([]byte(samplePipelineYAML))
	config, err := dsl.ToPipelineConfig()
	if err != nil {
		t.Fatalf("ToPipelineConfig failed: %v", err)
	}
	if config.Name != "sprint-review" {
		t.Errorf("config name mismatch")
	}
	if config.Mode != ModeSequential {
		t.Errorf("expected ModeSequential, got %s", config.Mode)
	}
	if len(config.Stages) != 3 {
		t.Errorf("expected 3 stages in config, got %d", len(config.Stages))
	}
	if !config.Stages[1].Adversarial {
		t.Error("expected second stage to have adversarial=true")
	}
}
