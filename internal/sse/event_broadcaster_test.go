package sse

import (
	"testing"
	"time"

	"github.com/muhammedikinci/pin/internal/interfaces"
)

func TestEventBroadcaster_NewEventBroadcaster(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	if broadcaster == nil {
		t.Fatal("Expected broadcaster to be created, got nil")
	}

	if broadcaster.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if len(broadcaster.clients) != 0 {
		t.Error("Expected clients map to be empty initially")
	}
}

func TestEventBroadcaster_AddClient(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	clientChan := make(chan interfaces.Event, 10)

	clientID := broadcaster.AddClient(clientChan)

	if clientID == "" {
		t.Error("Expected non-empty client ID")
	}

	if len(broadcaster.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(broadcaster.clients))
	}

	// Should receive welcome event
	select {
	case event := <-clientChan:
		if event.Type != "connection" {
			t.Errorf("Expected connection event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive welcome event")
	}
}

func TestEventBroadcaster_RemoveClient(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	clientChan := make(chan interfaces.Event, 10)

	clientID := broadcaster.AddClient(clientChan)

	// Drain welcome event
	<-clientChan

	broadcaster.RemoveClient(clientID)

	if len(broadcaster.clients) != 0 {
		t.Errorf("Expected 0 clients after removal, got %d", len(broadcaster.clients))
	}

	// Channel should be closed
	select {
	case _, ok := <-clientChan:
		if ok {
			t.Error("Expected client channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected client channel to be closed immediately")
	}
}

func TestEventBroadcaster_Broadcast(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	// Add multiple clients
	client1Chan := make(chan interfaces.Event, 10)
	client2Chan := make(chan interfaces.Event, 10)

	client1ID := broadcaster.AddClient(client1Chan)
	client2ID := broadcaster.AddClient(client2Chan)

	// Drain welcome events
	<-client1Chan
	<-client2Chan

	// Broadcast event
	testEvent := interfaces.Event{
		Type: "test",
		Data: map[string]interface{}{"message": "test message"},
	}

	broadcaster.Broadcast(testEvent)

	// Both clients should receive the event
	select {
	case event := <-client1Chan:
		if event.Type != "test" {
			t.Errorf("Expected test event type, got %s", event.Type)
		}
		if event.ID == "" {
			t.Error("Expected event ID to be generated")
		}
		if event.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client 1 should have received the event")
	}

	select {
	case event := <-client2Chan:
		if event.Type != "test" {
			t.Errorf("Expected test event type, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client 2 should have received the event")
	}

	// Clean up
	broadcaster.RemoveClient(client1ID)
	broadcaster.RemoveClient(client2ID)
}

func TestEventBroadcaster_BroadcastWithClosedBroadcaster(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	clientChan := make(chan interfaces.Event, 10)

	broadcaster.AddClient(clientChan)
	broadcaster.Close()

	// Broadcasting after close should not panic
	testEvent := interfaces.Event{
		Type: "test",
		Data: map[string]interface{}{"message": "test message"},
	}

	broadcaster.Broadcast(testEvent) // Should not panic
}

func TestEventBroadcaster_Close(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	client1Chan := make(chan interfaces.Event, 10)
	client2Chan := make(chan interfaces.Event, 10)

	broadcaster.AddClient(client1Chan)
	broadcaster.AddClient(client2Chan)

	// Drain welcome events
	<-client1Chan
	<-client2Chan

	broadcaster.Close()

	if len(broadcaster.clients) != 0 {
		t.Errorf("Expected 0 clients after close, got %d", len(broadcaster.clients))
	}

	if !broadcaster.closed {
		t.Error("Expected broadcaster to be marked as closed")
	}

	// All client channels should be closed
	select {
	case _, ok := <-client1Chan:
		if ok {
			t.Error("Expected client1 channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected client1 channel to be closed immediately")
	}

	select {
	case _, ok := <-client2Chan:
		if ok {
			t.Error("Expected client2 channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected client2 channel to be closed immediately")
	}
}

func TestEventBroadcaster_GetClientCount(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	if count := broadcaster.GetClientCount(); count != 0 {
		t.Errorf("Expected 0 clients initially, got %d", count)
	}

	client1Chan := make(chan interfaces.Event, 10)
	client2Chan := make(chan interfaces.Event, 10)

	broadcaster.AddClient(client1Chan)
	if count := broadcaster.GetClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}

	broadcaster.AddClient(client2Chan)
	if count := broadcaster.GetClientCount(); count != 2 {
		t.Errorf("Expected 2 clients, got %d", count)
	}
}