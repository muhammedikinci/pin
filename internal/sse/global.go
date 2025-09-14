package sse

import (
	"sync"

	"github.com/muhammedikinci/pin/internal/interfaces"
)

var (
	globalBroadcaster interfaces.EventBroadcaster
	broadcasterMutex  sync.RWMutex
)

// SetGlobalBroadcaster sets the global event broadcaster
func SetGlobalBroadcaster(broadcaster interfaces.EventBroadcaster) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()
	globalBroadcaster = broadcaster
}

// GetGlobalBroadcaster returns the global event broadcaster
func GetGlobalBroadcaster() interfaces.EventBroadcaster {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()
	return globalBroadcaster
}

// BroadcastGlobalEvent broadcasts an event using the global broadcaster
func BroadcastGlobalEvent(event interfaces.Event) {
	broadcaster := GetGlobalBroadcaster()
	if broadcaster != nil {
		broadcaster.Broadcast(event)
	}
}