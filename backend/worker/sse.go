package worker

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// SSEProgress represents a progress update event.
type SSEProgress struct {
	Current   int    `json:"current"`
	Total     int    `json:"total"`
	Progress  int    `json:"progress"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
}

// SSEComplete represents a completion event.
type SSEComplete struct {
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
}

// SSEError represents an error event.
type SSEError struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// SSEManager manages SSE connections for clients.
type SSEManager struct {
	clients map[uint]chan SSEEvent
	mu      sync.RWMutex
}

// NewSSEManager creates a new SSE manager.
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[uint]chan SSEEvent),
	}
}

// Subscribe adds a client channel for a given user/task ID.
func (m *SSEManager) Subscribe(id uint) chan SSEEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan SSEEvent, 100) // Buffer for async writes
	m.clients[id] = ch
	return ch
}

// Unsubscribe removes a client channel.
func (m *SSEManager) Unsubscribe(id uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, ok := m.clients[id]; ok {
		close(ch)
		delete(m.clients, id)
	}
}

// Publish sends an event to a specific client.
func (m *SSEManager) Publish(id uint, event SSEEvent) {
	m.mu.RLock()
	ch, ok := m.clients[id]
	m.mu.RUnlock()

	if ok {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// SendSSEProgress sends a progress update to a client.
func (m *SSEManager) SendSSEProgress(id uint, current, total int, status string) {
	progress := 0
	if total > 0 {
		progress = (current * 100) / total
	}

	event := SSEEvent{
		Event: "progress",
		Data: SSEProgress{
			Current:  current,
			Total:    total,
			Progress: progress,
			Status:   status,
		},
	}
	m.Publish(id, event)
}

// SendSSEComplete sends a completion event to a client.
func (m *SSEManager) SendSSEComplete(id uint, result interface{}) {
	event := SSEEvent{
		Event: "complete",
		Data: SSEComplete{
			Status: "done",
			Result: result,
		},
	}
	m.Publish(id, event)
}

// SendSSEError sends an error event to a client.
func (m *SSEManager) SendSSEError(id uint, errMsg string) {
	event := SSEEvent{
		Event: "error",
		Data: SSEError{
			Status: "failed",
			Error:  errMsg,
		},
	}
	m.Publish(id, event)
}

// SSEHandler returns a Gin handler for SSE connections.
func SSEHandler(sseManager *SSEManager, taskID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("X-Accel-Buffering", "no")

		// Subscribe to SSE events
		ch := sseManager.Subscribe(taskID)
		defer sseManager.Unsubscribe(taskID)

		// Send initial connection event
		c.SSEvent("connected", gin.H{"status": "connected", "task_id": taskID})
		c.Writer.Flush()

		// Stream events to client
		clientGone := c.Request.Context().Done()
		for {
			select {
			case <-clientGone:
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(event.Data)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Event, string(data))
				c.Writer.Flush()
			}
		}
	}
}
