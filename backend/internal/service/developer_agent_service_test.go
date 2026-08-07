package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeDeveloperJSON(t *testing.T) {
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
			got := string(normalizeDeveloperJSON(tt.in))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeBranchSlug(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"Add Login Page", "add-login-page"},
		{"Fix CVE-2024-1234!", "fix-cve-2024-1234"},
		{"UPPERCASE Title", "uppercase-title"},
		{"under_score_to_dash", "under-score-to-dash"},
		{"trim-trailing-dashes---", "trim-trailing-dashes"},
		{"---leading-trim", "leading-trim"},
		{"emoji 🚀 stripped", "emoji-stripped"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeBranchSlug(tt.in))
		})
	}

	t.Run("long slug is truncated to 40 chars", func(t *testing.T) {
		long := strings.Repeat("a", 100)
		got := sanitizeBranchSlug(long)
		assert.Equal(t, 40, len(got))
	})
}

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"no fence passes through", `{"files":[]}`, `{"files":[]}`},
		{"json fence stripped", "```json\n{\"files\":[]}\n```", `{"files":[]}`},
		{"plain fence stripped", "```\n{\"files\":[]}\n```", `{"files":[]}`},
		{"whitespace trimmed", `  {"files":[]}  `, `{"files":[]}`},
		{"only opening fence handled", "```\n{\"files\":[]}", `{"files":[]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripCodeFences(tt.in))
		})
	}
}

func TestIsBranchExistsErr(t *testing.T) {
	t.Run("nil returns false", func(t *testing.T) {
		assert.False(t, isBranchExistsErr(nil))
	})

	t.Run("plain error with already exists", func(t *testing.T) {
		assert.True(t, isBranchExistsErr(errors.New("Branch 'foo' already exists")))
	})

	t.Run("plain error without marker", func(t *testing.T) {
		assert.False(t, isBranchExistsErr(errors.New("network timeout")))
	})

	t.Run("AppError BadRequest with already exists", func(t *testing.T) {
		err := common.BadRequest("Branch 'dev-agent/1-foo' already exists or SHA is invalid")
		assert.True(t, isBranchExistsErr(err))
	})

	t.Run("AppError NotFound without marker", func(t *testing.T) {
		err := common.NotFound("Branch not found")
		assert.False(t, isBranchExistsErr(err))
	})
}

func TestStubGenerate(t *testing.T) {
	req := CodeGenerationRequest{
		WorkspaceID:     7,
		Title:           "Add login page",
		RequirementText: "As a user I want to log in",
		Language:        "typescript",
	}
	files := stubGenerate(req)
	if assert.Len(t, files, 1) {
		assert.Equal(t, "DEVELOPER_AGENT_OUTPUT.md", files[0].Path)
		assert.Contains(t, files[0].Content, "Add login page")
		assert.Contains(t, files[0].Content, "As a user I want to log in")
		assert.Contains(t, files[0].Content, "typescript")
	}
}

func TestLLMCodeGenerator_NilLLM(t *testing.T) {
	g := &llmCodeGenerator{llm: nil}
	files, err := g.Generate(context.Background(), CodeGenerationRequest{
		Title:           "Test",
		RequirementText: "Requirement",
	})
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "DEVELOPER_AGENT_OUTPUT.md", files[0].Path)
}

func TestGenerateBranchName(t *testing.T) {
	svc := &DeveloperAgentService{}

	t.Run("uses slug from title", func(t *testing.T) {
		job := &model.DeveloperJob{BaseModel: model.BaseModel{ID: 42}, Title: "Add Login"}
		assert.Equal(t, "dev-agent/42-add-login", svc.generateBranchName(job))
	})

	t.Run("falls back to feature slug when title is empty", func(t *testing.T) {
		job := &model.DeveloperJob{BaseModel: model.BaseModel{ID: 1}, Title: ""}
		assert.Equal(t, "dev-agent/1-feature", svc.generateBranchName(job))
	})

	t.Run("strips non-alphanumeric characters", func(t *testing.T) {
		job := &model.DeveloperJob{BaseModel: model.BaseModel{ID: 99}, Title: "Fix CVE-2024!!! 🚀"}
		assert.Equal(t, "dev-agent/99-fix-cve-2024", svc.generateBranchName(job))
	})
}

func TestBuildDefaultPRBody(t *testing.T) {
	svc := &DeveloperAgentService{}
	job := &model.DeveloperJob{
		BaseModel:  model.BaseModel{ID: 5},
		Title:      "Add Login Page",
		BranchName: "dev-agent/5-add-login-page",
	}
	files := []GeneratedFile{
		{Path: "src/login.tsx"},
		{Path: "src/login.css"},
	}

	t.Run("includes title, requirement, files, and committed count", func(t *testing.T) {
		body := svc.buildDefaultPRBody(job, "As a user I want to log in", files, 2)
		assert.Contains(t, body, "Add Login Page")
		assert.Contains(t, body, "As a user I want to log in")
		assert.Contains(t, body, "src/login.tsx")
		assert.Contains(t, body, "src/login.css")
		assert.Contains(t, body, "2 file(s) committed")
		assert.Contains(t, body, "dev-agent/5-add-login-page")
	})

	t.Run("falls back when requirement is empty", func(t *testing.T) {
		body := svc.buildDefaultPRBody(job, "", files, 1)
		assert.Contains(t, body, "No requirement text provided")
	})
}

// fakeGenerator is a test CodeGenerator that returns canned files.
type fakeGenerator struct {
	files []GeneratedFile
	err   error
}

func (f *fakeGenerator) Generate(ctx context.Context, req CodeGenerationRequest) ([]GeneratedFile, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.files, nil
}

func TestSetCodeGenerator(t *testing.T) {
	t.Run("overrides default generator", func(t *testing.T) {
		svc := &DeveloperAgentService{}
		original := svc.generator
		fake := &fakeGenerator{files: []GeneratedFile{{Path: "fake.txt", Content: "hi"}}}
		svc.SetCodeGenerator(fake)
		assert.NotEqual(t, original, svc.generator)
		assert.Equal(t, fake, svc.generator)

		got, err := svc.generator.Generate(context.Background(), CodeGenerationRequest{})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "fake.txt", got[0].Path)
	})

	t.Run("nil generator is ignored", func(t *testing.T) {
		svc := &DeveloperAgentService{}
		original := svc.generator
		svc.SetCodeGenerator(nil)
		assert.Equal(t, original, svc.generator)
	})
}

func TestFakeGenerator(t *testing.T) {
	t.Run("returns error when configured", func(t *testing.T) {
		f := &fakeGenerator{err: errors.New("boom")}
		_, err := f.Generate(context.Background(), CodeGenerationRequest{})
		assert.EqualError(t, err, "boom")
	})

	t.Run("returns files when no error", func(t *testing.T) {
		f := &fakeGenerator{files: []GeneratedFile{{Path: "a.go", Content: "package main"}}}
		got, err := f.Generate(context.Background(), CodeGenerationRequest{})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})
}
