/*
WebSocket Shared Infrastructure - Singleton Pattern

Creates SHARED WebSocket test infrastructure to eliminate the 131 instances
of duplicate test setup. Uses singleton pattern to reuse expensive components
across all tests.

Design Principles:
- SHARED SHARED SHARED: One infrastructure for all tests
- OPTIMIZE OPTIMIZE OPTIMIZE: Massive resource reduction
- GOOD PATTERNS: Reuses existing asserter logic
- MINIMAL CHANGE: Simple find-and-replace migration
*/

package websocket

import (
	"sync"
	"testing"
)

// SHARED WebSocket Infrastructure - Singleton Pattern
// This eliminates the 131 instances of duplicate infrastructure creation

var (
	sharedWebSocketAsserter *WebSocketIntegrationAsserter
	sharedAsserterOnce      sync.Once
)

// GetSharedWebSocketAsserter returns the SHARED WebSocket integration asserter
// This creates the expensive infrastructure (MediaMTX, Camera, JWT, Server) ONCE
// and reuses it across all tests, eliminating the 131 duplicate instances
func GetSharedWebSocketAsserter(t *testing.T) *WebSocketIntegrationAsserter {
	sharedAsserterOnce.Do(func() {
		// Create ONE shared asserter with full infrastructure
		// This is the ONLY place where expensive components are created
		sharedWebSocketAsserter = NewWebSocketIntegrationAsserter(t)
	})
	return sharedWebSocketAsserter
}

// CleanupSharedWebSocketInfrastructure cleans up the shared infrastructure
// This should be called at the end of the test suite
func CleanupSharedWebSocketInfrastructure() {
	if sharedWebSocketAsserter != nil {
		sharedWebSocketAsserter.Cleanup()
		sharedWebSocketAsserter = nil
		// Reset the once to allow re-initialization if needed
		sharedAsserterOnce = sync.Once{}
	}
}
