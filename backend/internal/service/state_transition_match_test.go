package service

import (
	"testing"

	"github.com/reqmango/backend/internal/model"
)

func TestStatesEquivalent(t *testing.T) {
	a := &model.State{Name: "Todo", Group: "unstarted"}
	b := &model.State{Name: "todo", Group: "unstarted"}
	c := &model.State{Name: "Done", Group: "completed"}
	if !statesEquivalent(a, b) {
		t.Fatal("expected name-equivalent states to match")
	}
	if statesEquivalent(a, c) {
		t.Fatal("expected different names not to match")
	}
	sameID := &model.State{Name: "X"}
	sameID.ID = 7
	other := &model.State{Name: "Y"}
	other.ID = 7
	if !statesEquivalent(sameID, other) {
		t.Fatal("expected same ID to match regardless of name")
	}
	if statesEquivalent(nil, a) || statesEquivalent(a, nil) {
		t.Fatal("nil should not match")
	}
}
