package loop

import (
	"fmt"
)

// LoopState represents the current phase of a Loop execution.
type LoopState string

const (
	StateIdle      LoopState = "idle"
	StatePlanning  LoopState = "planning"
	StateActing    LoopState = "acting"
	StateObserving LoopState = "observing"
	StateReasoning LoopState = "reasoning"
	StateWaiting   LoopState = "waiting"
	StateCompleted LoopState = "completed"
	StateFailed    LoopState = "failed"
)

// Decision is the outcome of the reasoning phase.
type Decision string

const (
	DecideContinue Decision = "continue"
	DecideStop     Decision = "stop"
	DecideEscalate Decision = "escalate"
	DecideWait     Decision = "wait"
)

// AllowedTransitions is the complete state transition map.
var AllowedTransitions = map[LoopState][]LoopState{
	StateIdle:      {StatePlanning},
	StatePlanning:  {StateActing, StateFailed},
	StateActing:    {StateObserving, StateFailed},
	StateObserving: {StateReasoning, StateFailed},
	StateReasoning: {StateActing, StateWaiting, StateCompleted, StateFailed},
	StateWaiting:   {StateActing, StateCompleted, StateFailed},
	StateCompleted: {}, // terminal
	StateFailed:    {}, // terminal
}

// StateMachine manages deterministic Loop state transitions.
type StateMachine struct {
	current LoopState
}

func NewStateMachine() *StateMachine {
	return &StateMachine{current: StateIdle}
}

func (sm *StateMachine) Current() LoopState {
	return sm.current
}

func (sm *StateMachine) IsTerminal() bool {
	return sm.current == StateCompleted || sm.current == StateFailed
}

// Transition attempts to move to the target state. Returns error if invalid.
func (sm *StateMachine) Transition(to LoopState) error {
	allowed, ok := AllowedTransitions[sm.current]
	if !ok {
		return fmt.Errorf("unknown current state: %s", sm.current)
	}
	for _, s := range allowed {
		if s == to {
			sm.current = to
			return nil
		}
	}
	return fmt.Errorf("invalid transition: %s -> %s", sm.current, to)
}
