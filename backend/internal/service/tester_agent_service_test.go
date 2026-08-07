package service

import (
	"context"
	"errors"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeTesterJSON(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{"nil becomes {}", nil, "{}"},
		{"empty becomes {}", []byte{}, "{}"},
		{"null becomes {}", []byte("null"), "{}"},
		{"valid json passes through", []byte(`{"foo":"bar"}`), `{"foo":"bar"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(normalizeTesterJSON(tt.in))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateTitle(t *testing.T) {
	t.Run("short string unchanged", func(t *testing.T) {
		assert.Equal(t, "abc", truncateTitle("abc", 10))
	})
	t.Run("long string truncated with ellipsis", func(t *testing.T) {
		got := truncateTitle("abcdefghijklmnopqrstuvwxyz", 5)
		assert.Equal(t, "abcde...", got)
	})
	t.Run("exact length unchanged", func(t *testing.T) {
		assert.Equal(t, "abcde", truncateTitle("abcde", 5))
	})
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"amp escaped", "a & b", "a &amp; b"},
		{"lt gt escaped", "<script>", "&lt;script&gt;"},
		{"quote escaped", `say "hi"`, "say &quot;hi&quot;"},
		{"plain passes through", "hello world", "hello world"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeHTML(tt.in))
		})
	}
}

func TestTesterStrPtr(t *testing.T) {
	p := testerStrPtr("tester_agent")
	if assert.NotNil(t, p) {
		assert.Equal(t, "tester_agent", *p)
	}
}

func TestIsEdgeEmptyCase(t *testing.T) {
	t.Run("tc-2 id matches", func(t *testing.T) {
		assert.True(t, isEdgeEmptyCase(TestCase{ID: "tc-2", Name: "Anything"}))
	})
	t.Run("name with empty matches", func(t *testing.T) {
		assert.True(t, isEdgeEmptyCase(TestCase{ID: "tc-9", Name: "Edge case - empty input"}))
	})
	t.Run("description with empty matches", func(t *testing.T) {
		assert.True(t, isEdgeEmptyCase(TestCase{ID: "tc-9", Name: "X", Description: "verify empty"}))
	})
	t.Run("non-empty case does not match", func(t *testing.T) {
		assert.False(t, isEdgeEmptyCase(TestCase{ID: "tc-1", Name: "Happy path", Description: "primary scenario"}))
	})
}

func TestStubGenerateCases(t *testing.T) {
	t.Run("produces deterministic cases", func(t *testing.T) {
		req := TestCaseGenerationRequest{
			WorkspaceID:        7,
			Title:              "Login page",
			RequirementText:    "As a user I want to log in",
			AcceptanceCriteria: "Given valid creds, when I submit, then I am authenticated.",
			TestScope:          "unit",
		}
		cases := stubGenerateCases(req)
		if assert.Len(t, cases, 3) {
			assert.Equal(t, "tc-1", cases[0].ID)
			assert.Equal(t, "tc-2", cases[1].ID)
			assert.Equal(t, "tc-3", cases[2].ID)
			// Every case needs a stable id and expected outcome.
			for _, c := range cases {
				assert.NotEmpty(t, c.ID)
				assert.NotEmpty(t, c.Expected)
			}
		}
	})

	t.Run("uses acceptance criteria as expected outcome", func(t *testing.T) {
		req := TestCaseGenerationRequest{
			Title:              "X",
			AcceptanceCriteria: "AC-1: returns 200",
		}
		cases := stubGenerateCases(req)
		assert.Equal(t, "AC-1: returns 200", cases[0].Expected)
	})

	t.Run("falls back when acceptance criteria empty", func(t *testing.T) {
		req := TestCaseGenerationRequest{Title: "X"}
		cases := stubGenerateCases(req)
		assert.Contains(t, cases[0].Expected, "Requirement is implemented")
	})

	t.Run("truncates long title in happy-path name", func(t *testing.T) {
		longTitle := "This is a very long requirement title that should be truncated"
		req := TestCaseGenerationRequest{Title: longTitle, AcceptanceCriteria: "AC"}
		cases := stubGenerateCases(req)
		assert.Contains(t, cases[0].Name, "Happy path - ")
		// Name should not contain the full long title.
		assert.NotContains(t, cases[0].Name, "should be truncated")
	})
}

func TestLLMTestCaseGenerator_NilLLM(t *testing.T) {
	g := &llmTestCaseGenerator{llm: nil}
	cases, err := g.Generate(context.Background(), TestCaseGenerationRequest{
		Title:              "Test",
		RequirementText:    "Requirement",
		AcceptanceCriteria: "AC",
	})
	assert.NoError(t, err)
	assert.Len(t, cases, 3)
}

func TestStubTestExecutor(t *testing.T) {
	exec := &stubTestExecutor{}
	cases := []TestCase{
		{ID: "tc-1", Name: "Happy path"},
		{ID: "tc-2", Name: "Edge case - empty input"},
		{ID: "tc-3", Name: "Negative case"},
	}

	t.Run("marks empty-input case as failed", func(t *testing.T) {
		results, err := exec.Execute(context.Background(), TestExecutionRequest{Cases: cases})
		assert.NoError(t, err)
		if assert.Len(t, results, 3) {
			assert.Equal(t, "passed", results[0].Status)
			assert.Equal(t, "failed", results[1].Status)
			assert.NotEmpty(t, results[1].Error)
			assert.Equal(t, "passed", results[2].Status)
		}
	})

	t.Run("populates duration and case ids", func(t *testing.T) {
		results, err := exec.Execute(context.Background(), TestExecutionRequest{Cases: cases})
		assert.NoError(t, err)
		for i, r := range results {
			assert.Equal(t, cases[i].ID, r.CaseID)
			assert.Equal(t, cases[i].Name, r.Name)
			assert.Greater(t, r.DurationMs, int64(0))
		}
	})

	t.Run("empty cases returns empty results", func(t *testing.T) {
		results, err := exec.Execute(context.Background(), TestExecutionRequest{})
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

// fakeCaseGenerator is a test TestCaseGenerator that returns canned cases.
type fakeCaseGenerator struct {
	cases []TestCase
	err   error
}

func (f *fakeCaseGenerator) Generate(ctx context.Context, req TestCaseGenerationRequest) ([]TestCase, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cases, nil
}

// fakeExecutor is a test TestExecutor that returns canned results.
type fakeExecutor struct {
	results []TestResult
	err     error
}

func (f *fakeExecutor) Execute(ctx context.Context, req TestExecutionRequest) ([]TestResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}

func TestSetTestCaseGenerator(t *testing.T) {
	t.Run("overrides default generator", func(t *testing.T) {
		svc := &TesterAgentService{}
		original := svc.generator
		fake := &fakeCaseGenerator{cases: []TestCase{{ID: "fake-1", Name: "Fake"}}}
		svc.SetTestCaseGenerator(fake)
		assert.NotEqual(t, original, svc.generator)
		assert.Equal(t, fake, svc.generator)

		got, err := svc.generator.Generate(context.Background(), TestCaseGenerationRequest{})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "fake-1", got[0].ID)
	})

	t.Run("nil generator is ignored", func(t *testing.T) {
		svc := &TesterAgentService{}
		original := svc.generator
		svc.SetTestCaseGenerator(nil)
		assert.Equal(t, original, svc.generator)
	})
}

func TestSetTestExecutor(t *testing.T) {
	t.Run("overrides default executor", func(t *testing.T) {
		svc := &TesterAgentService{}
		original := svc.executor
		fake := &fakeExecutor{results: []TestResult{{CaseID: "x", Status: "passed"}}}
		svc.SetTestExecutor(fake)
		assert.NotEqual(t, original, svc.executor)
		assert.Equal(t, fake, svc.executor)

		got, err := svc.executor.Execute(context.Background(), TestExecutionRequest{})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("nil executor is ignored", func(t *testing.T) {
		svc := &TesterAgentService{}
		original := svc.executor
		svc.SetTestExecutor(nil)
		assert.Equal(t, original, svc.executor)
	})
}

func TestFakeGeneratorAndExecutor(t *testing.T) {
	t.Run("fake generator propagates error", func(t *testing.T) {
		f := &fakeCaseGenerator{err: errors.New("gen boom")}
		_, err := f.Generate(context.Background(), TestCaseGenerationRequest{})
		assert.EqualError(t, err, "gen boom")
	})

	t.Run("fake executor propagates error", func(t *testing.T) {
		f := &fakeExecutor{err: errors.New("exec boom")}
		_, err := f.Execute(context.Background(), TestExecutionRequest{})
		assert.EqualError(t, err, "exec boom")
	})
}

func TestBuildBugTitle(t *testing.T) {
	svc := &TesterAgentService{}
	t.Run("uses case name", func(t *testing.T) {
		got := svc.buildBugTitle(nil, TestResult{Name: "Login fails on empty"})
		assert.Equal(t, "[Tester Agent] Login fails on empty", got)
	})
	t.Run("falls back to case id when name empty", func(t *testing.T) {
		got := svc.buildBugTitle(nil, TestResult{CaseID: "tc-2", Name: ""})
		assert.Equal(t, "[Tester Agent] tc-2", got)
	})
}

func TestBuildBugDescription(t *testing.T) {
	svc := &TesterAgentService{}
	tJob := &model.TesterJob{BaseModel: model.BaseModel{ID: 42}, Title: "Login feature"}
	r := TestResult{Name: "Empty input", Error: "500 <error>"}

	desc := svc.buildBugDescription(tJob, r)
	assert.Contains(t, desc, "Tester Agent")
	assert.Contains(t, desc, "Empty input")
	assert.Contains(t, desc, "500 &lt;error&gt;") // error is HTML-escaped
	assert.Contains(t, desc, "#42")
	assert.Contains(t, desc, "Login feature")
}
