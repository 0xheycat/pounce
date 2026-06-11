package api

import "sync"

// Hub is a tiny fan-out broadcaster for Server-Sent Events. Each connected
// client gets a buffered channel; slow clients simply drop frames rather than
// blocking the engine.
type Hub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[chan []byte]struct{})}
}

// Add registers a new client and returns its channel.
func (h *Hub) Add() chan []byte {
	ch := make(chan []byte, 32)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Remove unregisters a client and closes its channel.
func (h *Hub) Remove(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// Broadcast sends a frame to every client, dropping it for any client whose
// buffer is full.
func (h *Hub) Broadcast(b []byte) {
	h.mu.Lock()
	for ch := range h.clients {
		select {
		case ch <- b:
		default:
		}
	}
	h.mu.Unlock()
}
