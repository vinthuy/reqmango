// backend/internal/agent/loop/statemachine_test.go
package loop

import (
	"testing"
)

func TestStateMachineHappyPath(t *testing.T) {
	sm := NewStateMachine()
	if sm.Current() != StateIdle {
		t.Fatalf("expected idle, got %s", sm.Current())
	}

	steps := []struct {
		to  LoopState
		err bool
	}{
		{StatePlanning, false},
		{StateActing, false},
		{StateObserving, false},
		{StateReasoning, false},
		{StateActing, false},  // loop back
		{StateObserving, false},
		{StateReasoning, false},
		{StateCompleted, false},
	}

	for i, step := range steps {
		err := sm.Transition(step.to)
		if step.err && err == nil {
			t.Fatalf("step %d: expected error for transition to %s", i, step.to)
		}
		if !step.err && err != nil {
			t.Fatalf("step %d: unexpected error: %v", i, err)
		}
	}

	if !sm.IsTerminal() {
		t.Fatal("expected terminal state")
	}
}

func TestStateMachineInvalidTransition(t *testing.T) {
	sm := NewStateMachine()
	err := sm.Transition(StateCompleted)
	if err == nil {
		t.Fatal("expected error for idle->completed")
	}
}

func TestStateMachineTerminalNoTransition(t *testing.T) {
	sm := NewStateMachine()
	sm.Transition(StatePlanning)
	sm.Transition(StateActing)
	sm.Transition(StateObserving)
	sm.Transition(StateReasoning)
	sm.Transition(StateFailed)

	if !sm.IsTerminal() {
		t.Fatal("expected terminal after failed")
	}

	err := sm.Transition(StateActing)
	if err == nil {
		t.Fatal("expected error transitioning from terminal state")
	}
}
