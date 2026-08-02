package service

import (
	"fmt"
	"sync"
	"time"
)

// AgentReplyDebouncer prevents the same agent from being triggered for the
// same issue more than once within a configurable window. v1 is in-memory;
// restart clears the cache (acceptable).
type AgentReplyDebouncer struct {
	window time.Duration
	mu     sync.Mutex
	cache  map[string]time.Time // key: "agentID:issueID"
}

// NewAgentReplyDebouncer creates a debouncer with the given window.
func NewAgentReplyDebouncer(window time.Duration) *AgentReplyDebouncer {
	return &AgentReplyDebouncer{window: window, cache: make(map[string]time.Time)}
}

// Allow returns true if the agent+issue pair has not been triggered within the
// window, and records the trigger time. Returns false if within the window.
func (d *AgentReplyDebouncer) Allow(agentID, issueID uint64) bool {
	key := fmt.Sprintf("%d:%d", agentID, issueID)
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.cache[key]; ok && time.Since(t) < d.window {
		return false
	}
	d.cache[key] = time.Now()
	return true
}

// Cleanup removes expired entries. Call periodically from a background goroutine.
func (d *AgentReplyDebouncer) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for k, t := range d.cache {
		if now.Sub(t) >= d.window {
			delete(d.cache, k)
		}
	}
}
