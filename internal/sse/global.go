package sse

import (
	"sync"
)

var (
	globalBroadcaster EventBroadcaster
	broadcasterMutex  sync.RWMutex
)

// SetGlobalBroadcaster sets the global event broadcaster
func SetGlobalBroadcaster(broadcaster EventBroadcaster) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()
	globalBroadcaster = broadcaster
}

// GetGlobalBroadcaster returns the global event broadcaster
func GetGlobalBroadcaster() EventBroadcaster {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()
	return globalBroadcaster
}

// BroadcastGlobalEvent broadcasts an event using the global broadcaster
func BroadcastGlobalEvent(event Event) {
	broadcaster := GetGlobalBroadcaster()
	if broadcaster != nil {
		broadcaster.Broadcast(event)
	}
}