package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// WSMetricMessage defines the standard JSON payload streamed to WebSocket clients.
type WSMetricMessage struct {
	TestID             string        `json:"test_id"`
	Status             string        `json:"status"`
	Timestamp          string        `json:"timestamp"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ErrorRate          float64       `json:"error_rate"`
	RequestsPerSecond  float64       `json:"requests_per_second"`
	AverageLatencyMs   float64       `json:"average_latency_ms"`
	P50LatencyMs       float64       `json:"p50_latency_ms"`
	P95LatencyMs       float64       `json:"p95_latency_ms"`
	P99LatencyMs       float64       `json:"p99_latency_ms"`
	ActiveUsers        int           `json:"active_users"`
	StatusCodes        map[int]int64 `json:"status_codes,omitempty"`
}

// Hub manages active WebSocket clients, subscriptions, and broadcasts.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	stopChan   chan struct{}
	stopped    bool
}

// NewHub creates a new WebSocket Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		stopChan:   make(chan struct{}),
	}
}

// Run starts the event loop processing client connections and broadcast events.
func (h *Hub) Run() {
	for {
		select {
		case <-h.stopChan:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Drop slow client to prevent blocking the hub
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToTest sends a metric payload to clients subscribed to a specific test or all tests.
func (h *Hub) BroadcastToTest(testID string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Deliver if client has no specific test filter, matches "*", or matches testID
		if client.testID == "" || client.testID == "*" || client.testID == testID {
			select {
			case client.send <- data:
			default:
				log.Printf("[WebSocket] Warning: buffer full for client; dropping message")
			}
		}
	}

	return nil
}

// ConnectedClients returns the current number of active WebSocket connections.
func (h *Hub) ConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Close gracefully terminates the hub and all connected client channels.
func (h *Hub) Close() {
	h.mu.Lock()
	if !h.stopped {
		h.stopped = true
		close(h.stopChan)
	}
	h.mu.Unlock()
}
