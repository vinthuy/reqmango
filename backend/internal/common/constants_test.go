package common

import (
	"testing"
)

func TestRoleConstants(t *testing.T) {
	if RoleGuest > RoleMember {
		t.Error("RoleGuest should be less than RoleMember")
	}
	if RoleMember > RoleAdmin {
		t.Error("RoleMember should be less than RoleAdmin")
	}
	if RoleGuest < 0 {
		t.Error("RoleGuest should be non-negative")
	}
	if RoleAdmin < 0 {
		t.Error("RoleAdmin should be non-negative")
	}
}

func TestPriorityConstants(t *testing.T) {
	validPriorities := map[string]bool{
		PriorityUrgent: true,
		PriorityHigh:   true,
		PriorityMedium: true,
		PriorityLow:    true,
		PriorityNone:   true,
	}

	// Verify they're all distinct
	if PriorityUrgent == PriorityHigh {
		t.Error("Priorities should be distinct")
	}
	if PriorityMedium == PriorityLow {
		t.Error("Priorities should be distinct")
	}

	// Verify known values
	if PriorityUrgent != "urgent" {
		t.Errorf("PriorityUrgent = %q, want 'urgent'", PriorityUrgent)
	}
	if PriorityHigh != "high" {
		t.Errorf("PriorityHigh = %q, want 'high'", PriorityHigh)
	}
	if PriorityMedium != "medium" {
		t.Errorf("PriorityMedium = %q, want 'medium'", PriorityMedium)
	}
	if PriorityLow != "low" {
		t.Errorf("PriorityLow = %q, want 'low'", PriorityLow)
	}
	if PriorityNone != "none" {
		t.Errorf("PriorityNone = %q, want 'none'", PriorityNone)
	}

	for _, p := range []string{PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone} {
		if !validPriorities[p] {
			t.Errorf("Unexpected priority value: %q", p)
		}
	}
}

func TestStateGroupConstants(t *testing.T) {
	// Verify state groups are defined and distinct
	groups := []string{
		StateGroupBacklog,
		StateGroupUnstarted,
		StateGroupStarted,
		StateGroupCompleted,
		StateGroupCancelled,
	}

	seen := make(map[string]bool)
	for _, g := range groups {
		if g == "" {
			t.Error("State group should not be empty")
		}
		if seen[g] {
			t.Errorf("Duplicate state group: %q", g)
		}
		seen[g] = true
	}

	if StateGroupBacklog != "backlog" {
		t.Errorf("StateGroupBacklog = %q, want 'backlog'", StateGroupBacklog)
	}
	if StateGroupUnstarted != "unstarted" {
		t.Errorf("StateGroupUnstarted = %q, want 'unstarted'", StateGroupUnstarted)
	}
	if StateGroupCompleted != "completed" {
		t.Errorf("StateGroupCompleted = %q, want 'completed'", StateGroupCompleted)
	}
}

func TestDefaultStates(t *testing.T) {
	expectedCount := 6
	if len(DefaultStates) != expectedCount {
		t.Errorf("DefaultStates count = %d, want %d", len(DefaultStates), expectedCount)
	}

	expectedGroups := map[string]bool{
		StateGroupBacklog:   false,
		StateGroupUnstarted: false,
		StateGroupStarted:   false,
		StateGroupCompleted: false,
		StateGroupCancelled: false,
	}

	hasDefault := false
	for i, s := range DefaultStates {
		if s.Name == "" {
			t.Errorf("DefaultStates[%d]: Name should not be empty", i)
		}
		if s.Color == "" {
			t.Errorf("DefaultStates[%d]: Color should not be empty", i)
		}
		if s.Sequence != i+1 {
			t.Errorf("DefaultStates[%d]: Sequence = %d, want %d", i, s.Sequence, i+1)
		}

		if _, ok := expectedGroups[s.Group]; ok {
			expectedGroups[s.Group] = true
		} else {
			t.Errorf("DefaultStates[%d]: unexpected group %q", i, s.Group)
		}

		if s.IsDefault {
			hasDefault = true
		}
	}

	if !hasDefault {
		t.Error("At least one DefaultState should have IsDefault=true")
	}

	// Verify state group coverage
	for group, found := range expectedGroups {
		if !found {
			t.Errorf("State group %q not covered in DefaultStates", group)
		}
	}

	// Verify sequences are continuous and unique
	seqs := make(map[int]bool)
	for _, s := range DefaultStates {
		if seqs[s.Sequence] {
			t.Errorf("Duplicate DefaultState sequence: %d", s.Sequence)
		}
		seqs[s.Sequence] = true
	}
}

func TestDefaultStates_Colors(t *testing.T) {
	// Colors should be valid hex colors
	for i, s := range DefaultStates {
		if s.Color[0] != '#' {
			t.Errorf("DefaultStates[%d]: color %q should start with '#'", i, s.Color)
		}
		if len(s.Color) != 7 {
			t.Errorf("DefaultStates[%d]: color %q should be 7 chars (#RRGGBB)", i, s.Color)
		}
	}
}
