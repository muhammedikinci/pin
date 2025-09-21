package sse


//go:generate mockgen -source $GOFILE -destination ../mocks/mock_event_broadcaster.go -package mocks

// Event represents a server-sent event that can be broadcasted to clients

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
	
	// GetClientCount returns the number of connected clients
	GetClientCount() int
}