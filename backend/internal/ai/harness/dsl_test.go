package harness

import (
	"testing"
)

func TestParsePipelineDSL_Valid(t *testing.T) {
	yamlContent := []byte(`
name: test-pipeline
description: A test pipeline
version: "1.0"
pipeline:
  mode: sequential
  stages:
    - name: plan
      agent: planner-agent
      type: planner
    - name: execute
      agent: executor-agent
      type: executor
    - name: review
      agent: reviewer-agent
      type: reviewer
      mode: adversarial
retry:
  max_attempts: 3
  backoff: exponential
budget:
  max_tokens: 50000
  on_exhausted: escalate_to_human
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	if dsl.Name != "test-pipeline" {
		t.Errorf("Name = %q, want 'test-pipeline'", dsl.Name)
	}
	if dsl.Version != "1.0" {
		t.Errorf("Version = %q, want '1.0'", dsl.Version)
	}
	if dsl.Pipeline.Mode != "sequential" {
		t.Errorf("Mode = %q, want 'sequential'", dsl.Pipeline.Mode)
	}
	if len(dsl.Pipeline.Stages) != 3 {
		t.Errorf("got %d stages, want 3", len(dsl.Pipeline.Stages))
	}
	if dsl.Retry.MaxAttempts != 3 {
		t.Errorf("Retry.MaxAttempts = %d, want 3", dsl.Retry.MaxAttempts)
	}
}

func TestParsePipelineDSL_DefaultMode(t *testing.T) {
	yamlContent := []byte(`
name: simple
pipeline:
  stages:
    - name: do-it
      agent: worker
      type: executor
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	config, err := dsl.ToPipelineConfig()
	if err != nil {
		t.Fatalf("ToPipelineConfig() error = %v", err)
	}
	if config.Mode != ModeSequential {
		t.Errorf("default mode = %q, want 'sequential'", config.Mode)
	}
}

func TestValidate_MissingName(t *testing.T) {
	dsl := &PipelineDSL{
		Name: "",
		Pipeline: PipelineStagesDSL{
			Stages: []StageDSL{{Name: "s1", Agent: "a1", StageType: "executor"}},
		},
	}
	err := dsl.Validate()
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestValidate_NoStages(t *testing.T) {
	dsl := &PipelineDSL{
		Name:     "empty",
		Pipeline: PipelineStagesDSL{Stages: nil},
	}
	err := dsl.Validate()
	if err == nil {
		t.Error("expected error for empty stages")
	}
}

func TestValidate_InvalidStageType(t *testing.T) {
	dsl := &PipelineDSL{
		Name: "bad-type",
		Pipeline: PipelineStagesDSL{
			Stages: []StageDSL{
				{Name: "s1", Agent: "a1", StageType: "invalid_type"},
			},
		},
	}
	err := dsl.Validate()
	if err == nil {
		t.Error("expected error for invalid stage type")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	dsl := &PipelineDSL{
		Name: "bad-mode",
		Pipeline: PipelineStagesDSL{
			Mode:   "parallel", // not a valid mode
			Stages: []StageDSL{{Name: "s1", Agent: "a1", StageType: "executor"}},
		},
	}
	err := dsl.Validate()
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestParsePipelineDSL_Tournament(t *testing.T) {
	yamlContent := []byte(`
name: tournament-pipe
pipeline:
  mode: tournament
  stages:
    - name: contestant1
      agent: agent-a
      type: executor
    - name: contestant2
      agent: agent-b
      type: executor
    - name: contestant3
      agent: agent-c
      type: executor
    - name: judge
      agent: judge-agent
      type: judge
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	if dsl.Pipeline.Mode != "tournament" {
		t.Errorf("Mode = %q, want 'tournament'", dsl.Pipeline.Mode)
	}
	if len(dsl.Pipeline.Stages) != 4 {
		t.Errorf("got %d stages, want 4", len(dsl.Pipeline.Stages))
	}
}

func TestParsePipelineDSL_FanOut(t *testing.T) {
	yamlContent := []byte(`
name: fan-out-pipe
pipeline:
  mode: fan_out
  stages:
    - name: plan
      agent: planner
      type: planner
    - name: worker-a
      agent: agent-a
      type: executor
      input_from: plan
    - name: worker-b
      agent: agent-b
      type: executor
      input_from: plan
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	if dsl.Pipeline.Mode != "fan_out" {
		t.Errorf("Mode = %q, want 'fan_out'", dsl.Pipeline.Mode)
	}
}

func TestParsePipelineDSL_Classify(t *testing.T) {
	yamlContent := []byte(`
name: classify-pipe
pipeline:
  mode: classify
  stages:
    - name: classifier
      agent: classifier-agent
      type: executor
    - name: high-priority
      agent: fast-agent
      type: executor
    - name: low-priority
      agent: slow-agent
      type: executor
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	if dsl.Pipeline.Mode != "classify" {
		t.Errorf("Mode = %q, want 'classify'", dsl.Pipeline.Mode)
	}
}

func TestParsePipelineDSL_InvalidYAML(t *testing.T) {
	_, err := ParsePipelineDSL([]byte(`{{invalid yaml`))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParsePipelineDSL_TriggerConfig(t *testing.T) {
	yamlContent := []byte(`
name: triggered-pipe
trigger:
  type: cron
  schedule: "0 9 * * 1-5"
pipeline:
  stages:
    - name: daily
      agent: worker
      type: executor
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	if dsl.Trigger.Type != "cron" {
		t.Errorf("Trigger.Type = %q, want 'cron'", dsl.Trigger.Type)
	}
	if dsl.Trigger.Schedule != "0 9 * * 1-5" {
		t.Errorf("Trigger.Schedule = %q, want '0 9 * * 1-5'", dsl.Trigger.Schedule)
	}
}

func TestParsePipelineDSL_TournamentMode_ConvertsConfig(t *testing.T) {
	yamlContent := []byte(`
name: tourney
pipeline:
  mode: tournament
  stages:
    - name: a
      agent: agent-a
      type: executor
    - name: b
      agent: agent-b
      type: executor
    - name: judge
      agent: judge
      type: judge
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	config, err := dsl.ToPipelineConfig()
	if err != nil {
		t.Fatalf("ToPipelineConfig() error = %v", err)
	}
	if config.Mode != ModeTournament {
		t.Errorf("config.Mode = %q, want 'tournament'", config.Mode)
	}
	if len(config.Stages) != 3 {
		t.Errorf("got %d stages, want 3", len(config.Stages))
	}
}

func TestParsePipelineDSL_AdversarialReviewer(t *testing.T) {
	yamlContent := []byte(`
name: adversarial-pipe
pipeline:
  stages:
    - name: code
      agent: coder
      type: executor
    - name: review
      agent: reviewer
      type: reviewer
      mode: adversarial
`)
	dsl, err := ParsePipelineDSL(yamlContent)
	if err != nil {
		t.Fatalf("ParsePipelineDSL() error = %v", err)
	}
	config, err := dsl.ToPipelineConfig()
	if err != nil {
		t.Fatalf("ToPipelineConfig() error = %v", err)
	}
	if len(config.Stages) != 2 {
		t.Fatalf("got %d stages, want 2", len(config.Stages))
	}
	if !config.Stages[1].Adversarial {
		t.Error("reviewer stage should be adversarial")
	}
}

func TestParsePipelineDSL_MissingStageName(t *testing.T) {
	yamlContent := []byte(`
name: bad-stage
pipeline:
  stages:
    - agent: worker
      type: executor
`)
	_, err := ParsePipelineDSL(yamlContent)
	if err == nil {
		t.Error("expected error for missing stage name")
	}
}
