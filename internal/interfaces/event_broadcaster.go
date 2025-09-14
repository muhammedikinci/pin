package interfaces

import "time"

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks

// Event represents a server-sent event that can be broadcasted to clients
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventBroadcaster defines the interface for broadcasting events to SSE clients
type EventBroadcaster interface {
	// Broadcast sends an event to all connected SSE clients
	Broadcast(event Event)
	
	// AddClient adds a new SSE client connection
	AddClient(clientChan chan Event) string
	
	// RemoveClient removes an SSE client connection
	RemoveClient(clientID string)
	
	// Close shuts down the event broadcaster
	Close()
}