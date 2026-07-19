package loop

import (
	"testing"
)

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine()
	if sm.Current() != StateIdle {
		t.Errorf("initial state = %q, want %q", sm.Current(), StateIdle)
	}
	if sm.IsTerminal() {
		t.Error("initial state should not be terminal")
	}
}

func TestStateMachine_ValidTransitions(t *testing.T) {
	// Test the complete happy path: idle -> planning -> acting -> observing -> reasoning -> completed
	path := []LoopState{
		StateIdle,
		StatePlanning,
		StateActing,
		StateObserving,
		StateReasoning,
		StateCompleted,
	}
	sm := NewStateMachine()
	for i := 1; i < len(path); i++ {
		t.Run(string(path[i-1])+" -> "+string(path[i]), func(t *testing.T) {
			err := sm.Transition(path[i])
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if sm.Current() != path[i] {
				t.Errorf("state = %q, want %q", sm.Current(), path[i])
			}
		})
	}
}

func TestStateMachine_InvalidTransitions(t *testing.T) {
	invalidPairs := []struct {
		from LoopState
		to   LoopState
	}{
		{StateIdle, StateCompleted},   // cannot go straight to completed
		{StateIdle, StateActing},      // must go through planning
		{StatePlanning, StateIdle},    // no backward
		{StateCompleted, StateIdle},   // terminal cannot transition
		{StateFailed, StatePlanning},  // terminal cannot transition
		{StateActing, StatePlanning},  // no backward
	}
	for _, pair := range invalidPairs {
		t.Run(string(pair.from)+" -> "+string(pair.to), func(t *testing.T) {
			sm := NewStateMachine()
			// manually set to 'from' state
			sm.current = pair.from
			err := sm.Transition(pair.to)
			if err == nil {
				t.Errorf("expected error for invalid transition %s -> %s", pair.from, pair.to)
			}
		})
	}
}

func TestStateMachine_WaitingPath(t *testing.T) {
	// idle -> planning -> acting -> observing -> reasoning -> waiting -> acting
	sm := NewStateMachine()
	transitions := []LoopState{StatePlanning, StateActing, StateObserving, StateReasoning, StateWaiting, StateActing}
	for _, to := range transitions {
		if err := sm.Transition(to); err != nil {
			t.Fatalf("unexpected error transitioning to %s: %v", to, err)
		}
	}
	if sm.Current() != StateActing {
		t.Errorf("state = %q, want %q", sm.Current(), StateActing)
	}
}

func TestStateMachine_FailedPath(t *testing.T) {
	// idle -> planning -> acting -> failed
	sm := NewStateMachine()
	for _, to := range []LoopState{StatePlanning, StateActing, StateFailed} {
		if err := sm.Transition(to); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if !sm.IsTerminal() {
		t.Error("failed state should be terminal")
	}
	if err := sm.Transition(StatePlanning); err == nil {
		t.Error("should not transition from terminal state")
	}
}

func TestStateMachine_EscalatePath(t *testing.T) {
	// idle -> planning -> acting -> observing -> reasoning -> escalate
	// escalate is not in the state machine, so it should be manual (via reasoning -> completed/failed)
	sm := NewStateMachine()
	for _, to := range []LoopState{StatePlanning, StateActing, StateObserving, StateReasoning, StateCompleted} {
		if err := sm.Transition(to); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if !sm.IsTerminal() {
		t.Error("completed state should be terminal")
	}
}

func TestStateMachine_MultipleLoops(t *testing.T) {
	// simulate: planning -> acting -> observing -> reasoning -> acting -> observing -> reasoning -> completed
	sm := NewStateMachine()
	loop1 := []LoopState{StatePlanning, StateActing, StateObserving, StateReasoning}
	loop2 := []LoopState{StateActing, StateObserving, StateReasoning}
	finish := []LoopState{StateCompleted}

	for _, to := range loop1 {
		sm.Transition(to)
	}
	for _, to := range loop2 {
		sm.Transition(to)
	}
	for _, to := range finish {
		sm.Transition(to)
	}

	if sm.Current() != StateCompleted {
		t.Errorf("state = %q, want %q", sm.Current(), StateCompleted)
	}
}
