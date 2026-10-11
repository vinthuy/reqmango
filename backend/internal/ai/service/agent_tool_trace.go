package service

import (
	"encoding/json"
	"time"

	"github.com/reqmango/backend/internal/model"
)

const (
	maxTracedToolCalls  = 40
	maxTracedInputBytes = 1000
)

// toolTrace records the tool calls of one agent run for the activity audit
// log, failures included: a call the scope check rejected is exactly what an
// auditor needs to see.
type toolTrace struct {
	calls []model.AgentToolCall
}

func (t *toolTrace) add(name string, input json.RawMessage, err error, d time.Duration) {
	if len(t.calls) >= maxTracedToolCalls {
		return
	}
	call := model.AgentToolCall{Name: name, Input: boundedInput(input), DurationMs: d.Milliseconds()}
	if err != nil {
		call.Error = err.Error()
	}
	t.calls = append(t.calls, call)
}

// JSON returns the trace for storage, or nil when no tool was called.
func (t *toolTrace) JSON() json.RawMessage {
	if len(t.calls) == 0 {
		return nil
	}
	b, err := json.Marshal(t.calls)
	if err != nil {
		return nil
	}
	return b
}

// boundedInput keeps arguments readable without letting one long prompt-stuffed
// argument bloat the audit table; oversized input is stored as a JSON string.
func boundedInput(input json.RawMessage) json.RawMessage {
	if len(input) == 0 || !json.Valid(input) {
		return json.RawMessage(`{}`)
	}
	if len(input) <= maxTracedInputBytes {
		return input
	}
	r := []rune(string(input))
	for len(string(r)) > maxTracedInputBytes {
		r = r[:len(r)*9/10]
	}
	b, _ := json.Marshal(string(r) + "…")
	return b
}
