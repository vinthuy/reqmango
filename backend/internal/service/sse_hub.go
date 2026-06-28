package service

import (
	"encoding/json"
	"fmt"
	"sync"
)

// SSEClient represents a connected SSE client.
type SSEClient struct {
	UserID uint64
	Ch     chan string
}

// SSEHub manages SSE connections.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[uint64][]*SSEClient
}

var SSE = &SSEHub{clients: make(map[uint64][]*SSEClient)}

func (h *SSEHub) Register(userID uint64) *SSEClient {
	c := &SSEClient{UserID: userID, Ch: make(chan string, 32)}
	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], c)
	h.mu.Unlock()
	return c
}

func (h *SSEHub) Unregister(c *SSEClient) {
	h.mu.Lock()
	clients := h.clients[c.UserID]
	for i, cl := range clients {
		if cl == c {
			h.clients[c.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	h.mu.Unlock()
	close(c.Ch)
}

func (h *SSEHub) SendToUser(userID uint64, event, data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients[userID] {
		msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
		select {
		case c.Ch <- msg:
		default:
		}
	}
}

// NotifyUser sends a structured notification via SSE.
func (h *SSEHub) NotifyUser(userID uint64, ntype, title, message string) {
	data, _ := json.Marshal(map[string]string{"type": ntype, "title": title, "message": message})
	h.SendToUser(userID, "notification", string(data))
}
