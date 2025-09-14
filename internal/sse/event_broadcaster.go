package sse

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/muhammedikinci/pin/internal/interfaces"
)

// EventBroadcaster implements the interfaces.EventBroadcaster interface
// It manages SSE client connections and broadcasts events to all connected clients
type EventBroadcaster struct {
	clients map[string]chan interfaces.Event
	mutex   sync.RWMutex
	closed  bool
}

// NewEventBroadcaster creates a new event broadcaster instance
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		clients: make(map[string]chan interfaces.Event),
	}
}

// Broadcast sends an event to all connected SSE clients
func (eb *EventBroadcaster) Broadcast(event interfaces.Event) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	if eb.closed {
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Send to all connected clients
	for clientID, clientChan := range eb.clients {
		select {
		case clientChan <- event:
			// Event sent successfully
		default:
			// Client channel is full or closed, remove it
			go eb.RemoveClient(clientID)
		}
	}
}

// AddClient adds a new SSE client connection and returns the client ID
func (eb *EventBroadcaster) AddClient(clientChan chan interfaces.Event) string {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.closed {
		return ""
	}

	clientID := uuid.New().String()
	eb.clients[clientID] = clientChan

	// Send welcome event
	welcomeEvent := interfaces.Event{
		ID:        uuid.New().String(),
		Type:      "connection",
		Data:      map[string]interface{}{"message": "Connected to PIN SSE server", "clientId": clientID},
		Timestamp: time.Now(),
	}

	select {
	case clientChan <- welcomeEvent:
	default:
		// If we can't send welcome event, remove the client immediately
		delete(eb.clients, clientID)
		return ""
	}

	return clientID
}

// RemoveClient removes an SSE client connection
func (eb *EventBroadcaster) RemoveClient(clientID string) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if clientChan, exists := eb.clients[clientID]; exists {
		close(clientChan)
		delete(eb.clients, clientID)
	}
}

// Close shuts down the event broadcaster and closes all client connections
func (eb *EventBroadcaster) Close() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	eb.closed = true

	// Close all client channels
	for clientID, clientChan := range eb.clients {
		close(clientChan)
		delete(eb.clients, clientID)
	}
}

// GetClientCount returns the number of connected clients
func (eb *EventBroadcaster) GetClientCount() int {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return len(eb.clients)
}