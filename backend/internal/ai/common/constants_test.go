package common

import "testing"

func TestStateGroupConstants(t *testing.T) {
	// Verify all expected state groups are defined and distinct.
	groups := map[string]string{
		"backlog":   StateGroupBacklog,
		"unstarted": StateGroupUnstarted,
		"started":   StateGroupStarted,
		"completed": StateGroupCompleted,
		"cancelled": StateGroupCancelled,
	}
	for name, val := range groups {
		if val != name {
			t.Errorf("StateGroup constant mismatch: %s = %q, want %q", name, val, name)
		}
	}

	// Verify uniqueness
	seen := make(map[string]bool)
	all := []string{
		StateGroupBacklog, StateGroupUnstarted, StateGroupStarted,
		StateGroupCompleted, StateGroupCancelled,
	}
	for _, s := range all {
		if seen[s] {
			t.Errorf("duplicate state group: %q", s)
		}
		seen[s] = true
	}
	if len(seen) != 5 {
		t.Errorf("expected 5 unique state groups, got %d", len(seen))
	}
}
