package service

import (
	"testing"
	"time"
)

func TestAgentReplyDebouncer_AllowFirstBlocksSecond(t *testing.T) {
	d := NewAgentReplyDebouncer(30 * time.Second)
	if !d.Allow(5, 99) {
		t.Fatal("first call should be allowed")
	}
	if d.Allow(5, 99) {
		t.Fatal("second call within window should be blocked")
	}
}

func TestAgentReplyDebouncer_DifferentKeysIndependent(t *testing.T) {
	d := NewAgentReplyDebouncer(30 * time.Second)
	if !d.Allow(5, 1) {
		t.Fatal("agent 5 issue 1 should be allowed")
	}
	if !d.Allow(6, 1) {
		t.Fatal("agent 6 issue 1 should be allowed (different agent)")
	}
	if !d.Allow(5, 2) {
		t.Fatal("agent 5 issue 2 should be allowed (different issue)")
	}
}

func TestAgentReplyDebouncer_AllowsAfterWindowExpires(t *testing.T) {
	d := NewAgentReplyDebouncer(50 * time.Millisecond)
	if !d.Allow(1, 1) {
		t.Fatal("first call should be allowed")
	}
	time.Sleep(60 * time.Millisecond)
	if !d.Allow(1, 1) {
		t.Fatal("call after window expiry should be allowed")
	}
}

func TestAgentReplyDebouncer_CleanupRemovesExpiredKeys(t *testing.T) {
	d := NewAgentReplyDebouncer(20 * time.Millisecond)
	d.Allow(1, 1)
	time.Sleep(30 * time.Millisecond)
	d.Cleanup()
	d.mu.Lock()
	n := len(d.cache)
	d.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected 0 keys after cleanup, got %d", n)
	}
}
