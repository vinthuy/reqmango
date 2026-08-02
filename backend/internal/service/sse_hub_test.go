package service

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSSEHub_RegisterChatAndBroadcast(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}

	c1 := h.RegisterChat(42, 1)
	c2 := h.RegisterChat(42, 2)

	got := make([]string, 0, 2)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); s := <-c1.Ch; mu.Lock(); got = append(got, s); mu.Unlock() }()
	go func() { defer wg.Done(); s := <-c2.Ch; mu.Lock(); got = append(got, s); mu.Unlock() }()

	h.BroadcastToChat(42, "message_new", map[string]string{"content": "hi"})
	wg.Wait()

	if len(got) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", len(got))
	}
	for _, msg := range got {
		if !strings.Contains(msg, "event: message_new") || !strings.Contains(msg, "hi") {
			t.Errorf("unexpected message: %q", msg)
		}
	}
}

func TestSSEHub_UnregisterChatRemovesClient(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}
	c := h.RegisterChat(7, 1)
	h.UnregisterChat(7, c)

	h.mu.RLock()
	remaining := len(h.chatClients[7])
	h.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("expected 0 remaining chat clients, got %d", remaining)
	}
}

func TestSSEHub_BroadcastToChatDoesNotBlockWhenChannelFull(t *testing.T) {
	h := &SSEHub{
		clients:     make(map[uint64][]*SSEClient),
		chatClients: make(map[uint64][]*SSEClient),
	}
	c := h.RegisterChat(1, 1) // channel buffer 32
	// Fill the channel
	for i := 0; i < 32; i++ {
		c.Ch <- "filler"
	}
	done := make(chan struct{})
	go func() {
		h.BroadcastToChat(1, "message_new", "overflow") // should not block
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("BroadcastToChat blocked on full channel")
	}
}
