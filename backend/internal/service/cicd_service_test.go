package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeStringArray(t *testing.T) {
	t.Run("nil becomes empty array", func(t *testing.T) {
		got := string(normalizeStringArray(nil))
		assert.Equal(t, "[]", got)
	})
	t.Run("empty slice becomes empty array", func(t *testing.T) {
		got := string(normalizeStringArray([]string{}))
		assert.Equal(t, "[]", got)
	})
	t.Run("non-empty slice marshals as array", func(t *testing.T) {
		got := string(normalizeStringArray([]string{"push", "pull_request"}))
		assert.Equal(t, `["push","pull_request"]`, got)
	})
}

func TestNormalizeCICDJSON(t *testing.T) {
	t.Run("nil falls back", func(t *testing.T) {
		assert.Equal(t, "{}", string(normalizeCICDJSON(nil, "{}")))
	})
	t.Run("null falls back", func(t *testing.T) {
		assert.Equal(t, "{}", string(normalizeCICDJSON([]byte("null"), "{}")))
	})
	t.Run("valid passes through", func(t *testing.T) {
		in := []byte(`{"a":1}`)
		assert.Equal(t, string(in), string(normalizeCICDJSON(in, "{}")))
	})
	t.Run("array fallback supported", func(t *testing.T) {
		assert.Equal(t, "[]", string(normalizeCICDJSON(nil, "[]")))
	})
}

func TestIsValidProvider(t *testing.T) {
	cases := []struct {
		p    model.CICDProvider
		want bool
	}{
		{model.CICDProviderGitHubActions, true},
		{model.CICDProviderGitLabCI, true},
		{model.CICDProviderJenkins, true},
		{model.CICDProviderGeneric, true},
		{"unknown", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(string(c.p), func(t *testing.T) {
			assert.Equal(t, c.want, isValidProvider(c.p))
		})
	}
}

func TestIsValidTrigger(t *testing.T) {
	cases := []struct {
		v    model.BuildTrigger
		want bool
	}{
		{model.BuildTriggerManual, true},
		{model.BuildTriggerPush, true},
		{model.BuildTriggerPull, true},
		{model.BuildTriggerSchedule, true},
		{model.BuildTriggerAgent, true},
		{model.BuildTriggerWebhook, true},
		{"unknown", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(string(c.v), func(t *testing.T) {
			assert.Equal(t, c.want, isValidTrigger(c.v))
		})
	}
}

// fakeCICDProvider is a test CICDProvider that returns canned responses and
// records calls for assertion.
type fakeCICDProvider struct {
	triggerCalls int
	statusCalls  int
	cancelCalls  int
	externalID   string
	buildURL     string
	triggerErr   error
	statusSeq    []CICDProviderStatus
	cancelErr    error
}

func (f *fakeCICDProvider) Trigger(ctx context.Context, cfg *model.CICDConfig, req CICDTriggerRequest) (string, string, error) {
	f.triggerCalls++
	if f.triggerErr != nil {
		return "", "", f.triggerErr
	}
	if f.externalID == "" {
		return "fake-build-1", "https://example.invalid/builds/fake-build-1", nil
	}
	return f.externalID, f.buildURL, nil
}

func (f *fakeCICDProvider) GetStatus(ctx context.Context, cfg *model.CICDConfig, externalID string) (CICDProviderStatus, error) {
	idx := f.statusCalls
	f.statusCalls++
	if idx >= len(f.statusSeq) {
		// Default terminal state.
		return CICDProviderStatus{Status: model.BuildSuccess, Progress: 100, Stage: "completed"}, nil
	}
	return f.statusSeq[idx], nil
}

func (f *fakeCICDProvider) Cancel(ctx context.Context, cfg *model.CICDConfig, externalID string) error {
	f.cancelCalls++
	return f.cancelErr
}

var _ CICDProvider = (*fakeCICDProvider)(nil)

func TestSetCICDProvider(t *testing.T) {
	t.Run("overrides default provider", func(t *testing.T) {
		svc := &CICDService{}
		original := svc.provider
		fake := &fakeCICDProvider{externalID: "x", buildURL: "u"}
		svc.SetCICDProvider(fake)
		assert.NotEqual(t, original, svc.provider)
		assert.Equal(t, fake, svc.provider)
	})
	t.Run("nil provider is ignored", func(t *testing.T) {
		svc := &CICDService{}
		original := svc.provider
		svc.SetCICDProvider(nil)
		assert.Equal(t, original, svc.provider)
	})
}

func TestFakeCICDProvider(t *testing.T) {
	t.Run("trigger returns canned ids", func(t *testing.T) {
		f := &fakeCICDProvider{externalID: "ext-1", buildURL: "https://u.invalid/1"}
		gotID, gotURL, err := f.Trigger(context.Background(), &model.CICDConfig{}, CICDTriggerRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "ext-1", gotID)
		assert.Equal(t, "https://u.invalid/1", gotURL)
		assert.Equal(t, 1, f.triggerCalls)
	})
	t.Run("trigger propagates error", func(t *testing.T) {
		f := &fakeCICDProvider{triggerErr: errors.New("boom")}
		_, _, err := f.Trigger(context.Background(), &model.CICDConfig{}, CICDTriggerRequest{})
		assert.EqualError(t, err, "boom")
	})
	t.Run("getstatus walks sequence then returns success", func(t *testing.T) {
		f := &fakeCICDProvider{
			statusSeq: []CICDProviderStatus{
				{Status: model.BuildRunning, Progress: 30, Stage: "build"},
				{Status: model.BuildRunning, Progress: 60, Stage: "test"},
			},
		}
		s1, err := f.GetStatus(context.Background(), nil, "x")
		assert.NoError(t, err)
		assert.Equal(t, model.BuildRunning, s1.Status)
		assert.Equal(t, "build", s1.Stage)

		s2, err := f.GetStatus(context.Background(), nil, "x")
		assert.NoError(t, err)
		assert.Equal(t, "test", s2.Stage)

		s3, err := f.GetStatus(context.Background(), nil, "x")
		assert.NoError(t, err)
		assert.Equal(t, model.BuildSuccess, s3.Status)
	})
	t.Run("cancel records call", func(t *testing.T) {
		f := &fakeCICDProvider{}
		err := f.Cancel(context.Background(), nil, "x")
		assert.NoError(t, err)
		assert.Equal(t, 1, f.cancelCalls)
	})
}

func TestConfigToResponse(t *testing.T) {
	svc := &CICDService{}
	t.Run("nil trigger events become empty array", func(t *testing.T) {
		cfg := &model.CICDConfig{TriggerEvents: nil}
		resp := svc.configToResponse(cfg)
		assert.Equal(t, []string{}, resp.TriggerEvents)
	})
	t.Run("null trigger events become empty array", func(t *testing.T) {
		cfg := &model.CICDConfig{TriggerEvents: []byte("null")}
		resp := svc.configToResponse(cfg)
		assert.Equal(t, []string{}, resp.TriggerEvents)
	})
	t.Run("valid trigger events parsed", func(t *testing.T) {
		cfg := &model.CICDConfig{TriggerEvents: []byte(`["push","pull_request"]`)}
		resp := svc.configToResponse(cfg)
		assert.Equal(t, []string{"push", "pull_request"}, resp.TriggerEvents)
	})
	t.Run("nil extra config becomes object", func(t *testing.T) {
		cfg := &model.CICDConfig{ExtraConfig: nil}
		resp := svc.configToResponse(cfg)
		assert.Equal(t, "{}", string(resp.ExtraConfig))
	})
	t.Run("provider string preserved", func(t *testing.T) {
		cfg := &model.CICDConfig{Provider: model.CICDProviderGitHubActions}
		resp := svc.configToResponse(cfg)
		assert.Equal(t, "github_actions", resp.Provider)
	})
}

func TestBuildToResponse(t *testing.T) {
	svc := &CICDService{}
	t.Run("nil stages become empty array", func(t *testing.T) {
		b := &model.BuildRecord{Stages: nil}
		resp := svc.buildToResponse(b, nil)
		assert.Equal(t, "[]", string(resp.Stages))
	})
	t.Run("config name propagated when supplied", func(t *testing.T) {
		b := &model.BuildRecord{}
		cfg := &model.CICDConfig{Name: "prod-pipeline"}
		resp := svc.buildToResponse(b, cfg)
		assert.Equal(t, "prod-pipeline", resp.CICDConfigName)
	})
	t.Run("status and trigger serialized as strings", func(t *testing.T) {
		b := &model.BuildRecord{Status: model.BuildRunning, Trigger: model.BuildTriggerAgent}
		resp := svc.buildToResponse(b, nil)
		assert.Equal(t, "running", resp.Status)
		assert.Equal(t, "agent", resp.Trigger)
	})
	t.Run("stages payload round-trips", func(t *testing.T) {
		b := &model.BuildRecord{Stages: json.RawMessage(`[{"name":"build","status":"success"}]`)}
		resp := svc.buildToResponse(b, nil)
		assert.Contains(t, string(resp.Stages), `"build"`)
	})
}

// ======== Stub provider ========

func TestStubCICDProvider_Trigger(t *testing.T) {
	p := &stubCICDProvider{}
	t.Run("returns non-empty external id", func(t *testing.T) {
		id, _, err := p.Trigger(context.Background(), &model.CICDConfig{BaseModel: model.BaseModel{ID: 7}}, CICDTriggerRequest{})
		assert.NoError(t, err)
		assert.Contains(t, id, "stub-7-")
	})
	t.Run("build url empty when no endpoint", func(t *testing.T) {
		id, url, err := p.Trigger(context.Background(), &model.CICDConfig{BaseModel: model.BaseModel{ID: 1}}, CICDTriggerRequest{})
		assert.NoError(t, err)
		assert.Empty(t, url)
		_ = id
	})
	t.Run("build url composed from endpoint", func(t *testing.T) {
		_, url, err := p.Trigger(context.Background(), &model.CICDConfig{BaseModel: model.BaseModel{ID: 2}, APIEndpoint: "https://ci.example.invalid/"}, CICDTriggerRequest{})
		assert.NoError(t, err)
		assert.Contains(t, url, "https://ci.example.invalid/builds/")
	})
	t.Run("nil config errors", func(t *testing.T) {
		_, _, err := p.Trigger(context.Background(), nil, CICDTriggerRequest{})
		assert.Error(t, err)
	})
}

func TestStubCICDProvider_GetStatus(t *testing.T) {
	p := &stubCICDProvider{}
	t.Run("first poll returns running with first stage", func(t *testing.T) {
		s, err := p.GetStatus(context.Background(), &model.CICDConfig{}, "stub-1")
		assert.NoError(t, err)
		assert.Equal(t, model.BuildRunning, s.Status)
		assert.Equal(t, "build", s.Stage)
		assert.Equal(t, 25, s.Progress)
		if assert.Len(t, s.Stages, 3) {
			assert.Equal(t, "running", s.Stages[0].Status)
			assert.Equal(t, "pending", s.Stages[1].Status)
		}
	})
	t.Run("final poll returns success with all stages done", func(t *testing.T) {
		var last CICDProviderStatus
		for i := 0; i < 6; i++ {
			last, _ = p.GetStatus(context.Background(), &model.CICDConfig{}, "stub-final")
		}
		assert.Equal(t, model.BuildSuccess, last.Status)
		assert.Equal(t, 100, last.Progress)
		for _, st := range last.Stages {
			assert.Equal(t, "success", st.Status)
		}
	})
	t.Run("progress monotonically non-decreasing across polls", func(t *testing.T) {
		// Use a fresh stub to avoid shared callCount with the previous subtests.
		fresh := &stubCICDProvider{}
		prev := 0
		for i := 0; i < 5; i++ {
			s, _ := fresh.GetStatus(context.Background(), &model.CICDConfig{}, "stub-mono")
			assert.GreaterOrEqual(t, s.Progress, prev)
			prev = s.Progress
		}
	})
}

func TestStubCICDProvider_Cancel(t *testing.T) {
	p := &stubCICDProvider{}
	t.Run("cancel is no-op and never errors", func(t *testing.T) {
		err := p.Cancel(context.Background(), nil, "x")
		assert.NoError(t, err)
	})
}

// Verify the stub provider covers the canonical stage list.
func TestStubCICDStagesList(t *testing.T) {
	assert.Equal(t, []string{"build", "test", "deploy"}, stubCICDStages)
}

// Ensure the marshalled stage type round-trips.
func TestCICDStage_JSONRoundTrip(t *testing.T) {
	now := time.Now()
	s := CICDStage{Name: "build", Status: "success", DurationMs: 1234, StartedAt: &now, CompletedAt: &now, LogURL: "https://log.invalid"}
	b, err := json.Marshal(s)
	assert.NoError(t, err)
	var got CICDStage
	assert.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, s.Name, got.Name)
	assert.Equal(t, s.Status, got.Status)
	assert.Equal(t, s.DurationMs, got.DurationMs)
	assert.Equal(t, s.LogURL, got.LogURL)
}
