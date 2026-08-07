package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

// ======== Cron field parsing ========

func TestParseCronField(t *testing.T) {
	t.Run("wildcard expands to full range", func(t *testing.T) {
		got, err := parseCronField("*", 0, 5)
		assert.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, got)
	})
	t.Run("single value", func(t *testing.T) {
		got, err := parseCronField("3", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{3}, got)
	})
	t.Run("comma list", func(t *testing.T) {
		got, err := parseCronField("1,5,9", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 5, 9}, got)
	})
	t.Run("range", func(t *testing.T) {
		got, err := parseCronField("2-4", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{2, 3, 4}, got)
	})
	t.Run("step over wildcard", func(t *testing.T) {
		got, err := parseCronField("*/15", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{0, 15, 30, 45}, got)
	})
	t.Run("step over range", func(t *testing.T) {
		got, err := parseCronField("0-59/30", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{0, 30}, got)
	})
	t.Run("duplicates deduped", func(t *testing.T) {
		got, err := parseCronField("5,5,5", 0, 59)
		assert.NoError(t, err)
		assert.Equal(t, []int{5}, got)
	})
	t.Run("out of range rejected", func(t *testing.T) {
		_, err := parseCronField("60", 0, 59)
		assert.Error(t, err)
	})
	t.Run("invalid step rejected", func(t *testing.T) {
		_, err := parseCronField("*/0", 0, 59)
		assert.Error(t, err)
	})
	t.Run("empty field rejected", func(t *testing.T) {
		_, err := parseCronField("", 0, 59)
		assert.Error(t, err)
	})
	t.Run("non-numeric rejected", func(t *testing.T) {
		_, err := parseCronField("abc", 0, 59)
		assert.Error(t, err)
	})
}

// ======== Cron expression parsing ========

func TestParseCron(t *testing.T) {
	t.Run("valid 5-field expression", func(t *testing.T) {
		s, err := parseCron("*/5 * * * *")
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Contains(t, s.minute, 0)
		assert.Contains(t, s.minute, 5)
	})
	t.Run("too few fields", func(t *testing.T) {
		_, err := parseCron("* * * *")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "5 fields")
	})
	t.Run("too many fields", func(t *testing.T) {
		_, err := parseCron("* * * * * *")
		assert.Error(t, err)
	})
	t.Run("whitespace trimmed", func(t *testing.T) {
		s, err := parseCron("  0   0   *   *   *  ")
		assert.NoError(t, err)
		assert.Equal(t, []int{0}, s.minute)
		assert.Equal(t, []int{0}, s.hour)
	})
	t.Run("invalid field surfaces context", func(t *testing.T) {
		_, err := parseCron("60 * * * *")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minute")
	})
}

// ======== calculateNextRun ========

func TestCalculateNextRun(t *testing.T) {
	// Use a fixed reference time so assertions are deterministic.
	ref := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) // Friday 12:00 UTC

	t.Run("every minute fires next minute", func(t *testing.T) {
		next, err := calculateNextRun("* * * * *", ref)
		assert.NoError(t, err)
		assert.Equal(t, ref.Add(time.Minute), next)
	})
	t.Run("hourly at :30", func(t *testing.T) {
		next, err := calculateNextRun("30 * * * *", ref)
		assert.NoError(t, err)
		assert.Equal(t, 30, next.Minute())
		assert.Equal(t, 12, next.Hour())
	})
	t.Run("daily at 00:00", func(t *testing.T) {
		next, err := calculateNextRun("0 0 * * *", ref)
		assert.NoError(t, err)
		assert.Equal(t, 0, next.Minute())
		assert.Equal(t, 0, next.Hour())
		assert.True(t, next.After(ref))
		// Should be the next day's midnight (2026-08-08 00:00:00 UTC).
		assert.Equal(t, ref.AddDate(0, 0, 1).Truncate(24*time.Hour), next)
	})
	t.Run("every 15 minutes", func(t *testing.T) {
		next, err := calculateNextRun("*/15 * * * *", ref)
		assert.NoError(t, err)
		assert.Equal(t, 15, next.Minute())
	})
	t.Run("specific day of week", func(t *testing.T) {
		// ref is Friday (5). Sunday=0. Next Sunday after 2026-08-07.
		next, err := calculateNextRun("0 0 * * 0", ref)
		assert.NoError(t, err)
		assert.Equal(t, time.Sunday, next.Weekday())
		assert.True(t, next.After(ref))
	})
	t.Run("invalid expression errors", func(t *testing.T) {
		_, err := calculateNextRun("not-a-cron", ref)
		assert.Error(t, err)
	})
	t.Run("impossible date errors within 5y deadline", func(t *testing.T) {
		// Feb 30 never exists.
		_, err := calculateNextRun("0 0 30 2 *", ref)
		assert.Error(t, err)
	})
}

func TestCronScheduleMatch(t *testing.T) {
	s, err := parseCron("30 14 * * *")
	assert.NoError(t, err)
	t.Run("matching time", func(t *testing.T) {
		tt := time.Date(2026, 8, 7, 14, 30, 0, 0, time.UTC)
		assert.True(t, s.match(tt))
	})
	t.Run("wrong minute", func(t *testing.T) {
		tt := time.Date(2026, 8, 7, 14, 31, 0, 0, time.UTC)
		assert.False(t, s.match(tt))
	})
	t.Run("wrong hour", func(t *testing.T) {
		tt := time.Date(2026, 8, 7, 15, 30, 0, 0, time.UTC)
		assert.False(t, s.match(tt))
	})
}

func TestContainsInt(t *testing.T) {
	assert.True(t, containsInt([]int{1, 2, 3}, 2))
	assert.False(t, containsInt([]int{1, 2, 3}, 4))
	assert.False(t, containsInt(nil, 1))
}

// ======== Trigger type validation ========

func TestAutopilotTriggerTypes(t *testing.T) {
	assert.True(t, autopilotTriggerTypes["cron"])
	assert.True(t, autopilotTriggerTypes["webhook"])
	assert.True(t, autopilotTriggerTypes["manual"])
	assert.False(t, autopilotTriggerTypes["bogus"])
	assert.False(t, autopilotTriggerTypes[""])
}

// ======== Stub executor output ========

func TestStubAutopilotTaskOutput(t *testing.T) {
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 7}, Name: "每日报告", TaskType: "report"}
	out := stubAutopilotTaskOutput(task)
	assert.Contains(t, out, "report_url")
	assert.Contains(t, out["report_url"], "autopilot-7")
	assert.Contains(t, out["summary"], "每日报告")
}

func TestStubAutopilotTaskOutputAllTypes(t *testing.T) {
	cases := []struct{ taskType string }{
		{"report"}, {"build_report"}, {"scan"}, {"security_scan"},
		{"sync"}, {"sync_issues"}, {"backup"}, {"notify"}, {"notification"},
		{"custom_unknown_type"},
	}
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 1}, Name: "T"}
	for _, c := range cases {
		task.TaskType = c.taskType
		out := stubAutopilotTaskOutput(task)
		assert.NotEmpty(t, out, "task_type %s should produce output", c.taskType)
	}
}

func TestStubAutopilotTaskOutputEmptyNameFallback(t *testing.T) {
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 1}, Name: "", TaskType: "report"}
	out := stubAutopilotTaskOutput(task)
	assert.Contains(t, out["summary"], "未命名任务")
}

func TestStubAutopilotTaskOutputDefaultBranch(t *testing.T) {
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 42}, Name: "自定义", TaskType: "weird_type"}
	out := stubAutopilotTaskOutput(task)
	assert.Equal(t, "weird_type", out["task_type"])
	assert.Equal(t, uint64(42), out["task_id"])
	assert.Contains(t, out["message"], "自定义")
}

// ======== Stub executor invocation ========

// recordingAutopilotExecutor is a test AutopilotExecutor that records the task it
// received and returns a canned output / optional error.
type recordingAutopilotExecutor struct {
	receivedTaskID uint64
	failErr        error
}

func (e *recordingAutopilotExecutor) Execute(task *model.AutopilotTask, exec *model.AutopilotExecution) (map[string]interface{}, []string, error) {
	e.receivedTaskID = task.ID
	if e.failErr != nil {
		return nil, nil, e.failErr
	}
	return map[string]interface{}{"ran": true}, []string{"executed"}, nil
}

func TestRecordingAutopilotExecutorSuccess(t *testing.T) {
	ex := &recordingAutopilotExecutor{}
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 9}, TaskType: "report"}
	exec := &model.AutopilotExecution{BaseModel: model.BaseModel{ID: 1}, TriggerType: "manual"}
	out, logs, err := ex.Execute(task, exec)
	assert.NoError(t, err)
	assert.Equal(t, uint64(9), ex.receivedTaskID)
	assert.True(t, out["ran"].(bool))
	assert.Equal(t, []string{"executed"}, logs)
}

func TestRecordingAutopilotExecutorFailure(t *testing.T) {
	sentinel := errors.New("boom")
	ex := &recordingAutopilotExecutor{failErr: sentinel}
	task := &model.AutopilotTask{BaseModel: model.BaseModel{ID: 1}, TaskType: "scan"}
	_, _, err := ex.Execute(task, &model.AutopilotExecution{})
	assert.Same(t, sentinel, err)
}

// Compile-time interface check.
var _ AutopilotExecutor = (*recordingAutopilotExecutor)(nil)

// ======== Helpers ========

func TestNormalizeAutopilotJSON(t *testing.T) {
	t.Run("nil yields fallback", func(t *testing.T) {
		assert.Equal(t, "{}", string(normalizeAutopilotJSON(nil, "{}")))
	})
	t.Run("empty map yields fallback", func(t *testing.T) {
		assert.Equal(t, "{}", string(normalizeAutopilotJSON(map[string]interface{}{}, "{}")))
	})
	t.Run("marshals content", func(t *testing.T) {
		got := normalizeAutopilotJSON(map[string]interface{}{"a": 1}, "{}")
		assert.Equal(t, `{"a":1}`, string(got))
	})
}

func TestNormalizeAutopilotJSONRaw(t *testing.T) {
	assert.Equal(t, "{}", string(normalizeAutopilotJSONRaw(nil, "{}")))
	assert.Equal(t, "[]", string(normalizeAutopilotJSONRaw(nil, "[]")))
	assert.Equal(t, "{}", string(normalizeAutopilotJSONRaw(json.RawMessage("null"), "{}")))
	assert.Equal(t, `{"a":1}`, string(normalizeAutopilotJSONRaw(json.RawMessage(`{"a":1}`), "{}")))
}

func TestMarshalAutopilotOutput(t *testing.T) {
	assert.Equal(t, "{}", string(marshalAutopilotOutput(nil)))
	got := marshalAutopilotOutput(map[string]interface{}{"k": "v"})
	assert.Equal(t, `{"k":"v"}`, string(got))
}

func TestMarshalAutopilotLogs(t *testing.T) {
	assert.Equal(t, "[]", string(marshalAutopilotLogs(nil)))
	assert.Equal(t, `["a","b"]`, string(marshalAutopilotLogs([]string{"a", "b"})))
}

func TestSummarizeAutopilotOutput(t *testing.T) {
	assert.Equal(t, "(empty)", summarizeAutopilotOutput(nil))
	out := map[string]interface{}{"a": 1, "b": 2}
	summary := summarizeAutopilotOutput(out)
	assert.Contains(t, summary, "a")
	assert.Contains(t, summary, "b")
}

func TestGenerateWebhookToken(t *testing.T) {
	a := generateWebhookToken()
	b := generateWebhookToken()
	assert.NotEmpty(t, a)
	// crypto/rand token is 16 bytes hex-encoded → 32 chars.
	assert.Equal(t, 32, len(a), "token should be 32 hex chars")
	assert.Equal(t, 32, len(b), "token should be 32 hex chars")
	// Two random 128-bit tokens should effectively never collide.
	assert.NotEqual(t, a, b)
}

// ======== Status constants ========

func TestAutopilotStatusConstants(t *testing.T) {
	// Task statuses are stable strings the API contract depends on.
	assert.Equal(t, "active", autopilotTaskActive)
	assert.Equal(t, "paused", autopilotTaskPaused)
	// Execution statuses form the pending -> running -> completed/failed flow.
	assert.Equal(t, "pending", autopilotExecPending)
	assert.Equal(t, "running", autopilotExecRunning)
	assert.Equal(t, "completed", autopilotExecComplete)
	assert.Equal(t, "failed", autopilotExecFailed)
	// Terminal statuses are distinct from the in-flight ones.
	assert.NotEqual(t, autopilotExecRunning, autopilotExecComplete)
	assert.NotEqual(t, autopilotExecRunning, autopilotExecFailed)
	assert.NotEqual(t, autopilotExecComplete, autopilotExecFailed)
}
