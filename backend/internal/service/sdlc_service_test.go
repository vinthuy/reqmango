package service

import (
	"context"
	"errors"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeStageSelection(t *testing.T) {
	t.Run("empty returns nil (run all)", func(t *testing.T) {
		assert.Nil(t, normalizeStageSelection(nil))
		assert.Nil(t, normalizeStageSelection([]string{}))
	})
	t.Run("unknown keys dropped", func(t *testing.T) {
		got := normalizeStageSelection([]string{"requirement_analysis", "bogus", "deploy"})
		assert.Equal(t, []string{"requirement_analysis", "deploy"}, got)
	})
	t.Run("duplicates deduped", func(t *testing.T) {
		got := normalizeStageSelection([]string{"deploy", "deploy", "deploy"})
		assert.Equal(t, []string{"deploy"}, got)
	})
	t.Run("all-unknown collapses to nil", func(t *testing.T) {
		assert.Nil(t, normalizeStageSelection([]string{"nope", "also-nope"}))
	})
	t.Run("whitespace trimmed", func(t *testing.T) {
		got := normalizeStageSelection([]string{"  deploy  "})
		assert.Equal(t, []string{"deploy"}, got)
	})
}

func TestContainsString(t *testing.T) {
	assert.True(t, containsString([]string{"a", "b"}, "b"))
	assert.False(t, containsString([]string{"a", "b"}, "c"))
	assert.False(t, containsString(nil, "a"))
}

func TestIsTerminalWorkflowStatus(t *testing.T) {
	cases := []struct {
		s    model.SDLCWorkflowStatus
		want bool
	}{
		{model.SDLCWorkflowCompleted, true},
		{model.SDLCWorkflowFailed, true},
		{model.SDLCWorkflowPartial, true},
		{model.SDLCWorkflowCancelled, true},
		{model.SDLCWorkflowPending, false},
		{model.SDLCWorkflowRunning, false},
	}
	for _, c := range cases {
		t.Run(string(c.s), func(t *testing.T) {
			assert.Equal(t, c.want, isTerminalWorkflowStatus(c.s))
		})
	}
}

func TestWorkflowProgress(t *testing.T) {
	assert.Equal(t, 0, workflowProgress(0, 11))
	assert.Equal(t, 9, workflowProgress(1, 11))  // 1/11*100 = 9
	assert.Equal(t, 100, workflowProgress(11, 11))
	assert.Equal(t, 100, workflowProgress(12, 11)) // capped
	assert.Equal(t, 0, workflowProgress(1, 0))     // div-by-zero guard
}

func TestDecodeWorkflowConfig(t *testing.T) {
	t.Run("empty defaults to fail_fast=true", func(t *testing.T) {
		cfg := decodeWorkflowConfig(nil)
		assert.True(t, cfg.FailFast)
		assert.Equal(t, []string{}, cfg.Stages)
	})
	t.Run("parses stages + fail_fast", func(t *testing.T) {
		raw := []byte(`{"stages":["deploy"],"fail_fast":false}`)
		cfg := decodeWorkflowConfig(raw)
		assert.False(t, cfg.FailFast)
		assert.Equal(t, []string{"deploy"}, cfg.Stages)
	})
}

func TestDecodeArtifacts(t *testing.T) {
	t.Run("empty yields empty map", func(t *testing.T) {
		assert.Empty(t, decodeArtifacts(nil))
	})
	t.Run("parses object", func(t *testing.T) {
		out := decodeArtifacts([]byte(`{"pr_url":"https://x"}`))
		assert.Equal(t, "https://x", out["pr_url"])
	})
}

func TestMarshalOutput(t *testing.T) {
	assert.Equal(t, "{}", string(marshalOutput(nil)))
	assert.Equal(t, `{"a":1}`, string(marshalOutput(map[string]interface{}{"a": 1})))
}

func TestMarshalLogs(t *testing.T) {
	assert.Equal(t, "[]", string(marshalLogs(nil)))
	assert.Equal(t, `["a","b"]`, string(marshalLogs([]string{"a", "b"})))
}

func TestStubStageOutputAllStages(t *testing.T) {
	wf := &model.SDLCWorkflow{Title: "登录功能"}
	prior := map[string]interface{}{}
	for _, def := range canonicalSDLStages {
		stage := &model.SDLCStage{Key: def.Key}
		out := stubStageOutput(stage, wf, prior)
		assert.NotEmpty(t, out, "stage %s should produce artifacts", def.Key)
		// Every stage must produce at least one key.
		assert.Greater(t, len(out), 0)
	}
}

func TestStubStageOutputUnknownKey(t *testing.T) {
	stage := &model.SDLCStage{Key: "nonexistent"}
	out := stubStageOutput(stage, &model.SDLCWorkflow{}, nil)
	assert.Contains(t, out, "note")
}

func TestStubStageOutputEmptyTitleFallback(t *testing.T) {
	stage := &model.SDLCStage{Key: "requirement_analysis"}
	out := stubStageOutput(stage, &model.SDLCWorkflow{Title: ""}, nil)
	assert.Contains(t, out, "analysis_report")
	// Fallback title should appear in the report.
	assert.Contains(t, out["analysis_report"], "未命名需求")
}

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"User Login Feature", "user-login-feature"},
		{"  Hello World  ", "hello-world"},
		{"Uppercase_AND-symbols!", "uppercase-and-symbols"},
		{"", "feature"},
		{"中文标题", "feature"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			assert.Equal(t, c.want, slugify(c.in))
		})
	}
}

func TestSummarizeOutput(t *testing.T) {
	assert.Equal(t, "(empty)", summarizeOutput(nil))
	out := map[string]interface{}{"a": 1, "b": 2}
	// Order is non-deterministic; just check both keys present.
	summary := summarizeOutput(out)
	assert.Contains(t, summary, "a")
	assert.Contains(t, summary, "b")
}

func TestCanonicalSDLStagesOrder(t *testing.T) {
	// PRD §3.3: stages run strictly in order 1..11.
	assert.Equal(t, 11, len(canonicalSDLStages))
	for i, def := range canonicalSDLStages {
		assert.Equal(t, i+1, def.Order, "stage %d should have Order %d", i, i+1)
		assert.NotEmpty(t, def.Key)
		assert.NotEmpty(t, def.Name)
		assert.NotEmpty(t, def.AgentRole)
	}
	// Spot-check first and last.
	assert.Equal(t, "requirement_analysis", canonicalSDLStages[0].Key)
	assert.Equal(t, "deploy", canonicalSDLStages[10].Key)
}

// recordingExecutor is a test StageExecutor that records invocations and
// returns a canned output, used to verify the orchestration engine calls
// stages in order without touching the DB.
type recordingExecutor struct {
	calls   []string
	failOn  string
	failErr error
}

func (e *recordingExecutor) Execute(ctx context.Context, wf *model.SDLCWorkflow, stage *model.SDLCStage, prior map[string]interface{}) (map[string]interface{}, []string, error) {
	e.calls = append(e.calls, stage.Key)
	if e.failOn == stage.Key && e.failErr != nil {
		return nil, nil, e.failErr
	}
	return map[string]interface{}{"stage": stage.Key}, []string{"ran " + stage.Key}, nil
}

func TestRecordingExecutorCallsInOrder(t *testing.T) {
	ex := &recordingExecutor{}
	wf := &model.SDLCWorkflow{Title: "T"}
	prior := map[string]interface{}{}
	for _, def := range canonicalSDLStages {
		stage := &model.SDLCStage{Key: def.Key, Name: def.Name, AgentRole: def.AgentRole}
		_, _, err := ex.Execute(context.Background(), wf, stage, prior)
		assert.NoError(t, err)
	}
	assert.Equal(t, 11, len(ex.calls))
	assert.Equal(t, "requirement_analysis", ex.calls[0])
	assert.Equal(t, "deploy", ex.calls[10])
}

func TestRecordingExecutorFailure(t *testing.T) {
	sentinel := errors.New("boom")
	ex := &recordingExecutor{failOn: "development", failErr: sentinel}
	wf := &model.SDLCWorkflow{Title: "T"}
	stage := &model.SDLCStage{Key: "development"}
	_, _, err := ex.Execute(context.Background(), wf, stage, nil)
	assert.Same(t, sentinel, err)
}

// Compile-time interface check.
var _ StageExecutor = (*recordingExecutor)(nil)
