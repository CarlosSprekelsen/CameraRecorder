//go:build test

/*
WebSocket Shared Test Infrastructure - Test-Only Singleton Pattern

Creates SHARED WebSocket test infrastructure to eliminate the 144 instances
of duplicate test setup. This file is test-only and uses Go's build tags
to ensure it's only compiled during tests.

Design Principles:
- SHARED SHARED SHARED: One infrastructure for all tests
- OPTIMIZE OPTIMIZE OPTIMIZE: Massive resource reduction
- GOOD PATTERNS: Reuses existing asserter logic
- MINIMAL CHANGE: Simple find-and-replace migration
- TEST-ONLY: Respects Go's compilation model with build tags
*/

package websocket

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// SHARED WebSocket Infrastructure - Singleton Pattern (Test-Only)
// This eliminates the 144 instances of duplicate infrastructure creation

var (
	sharedWebSocketAsserter *WebSocketIntegrationAsserter
	sharedAsserterOnce      sync.Once
)

// GetSharedWebSocketAsserter returns the SHARED WebSocket integration asserter
// This creates the expensive infrastructure (MediaMTX, Camera, JWT, Server) ONCE
// and reuses it across all tests, eliminating the 144 duplicate instances
//
// This function is test-only and will only be available during test compilation
func GetSharedWebSocketAsserter(t *testing.T) *WebSocketIntegrationAsserter {
	sharedAsserterOnce.Do(func() {
		// Create ONE shared asserter with full infrastructure
		// This is the ONLY place where expensive components are created
		// FIXED: Use the actual implementation instead of recursive call
		sharedWebSocketAsserter = createSharedWebSocketAsserter(t)
	})
	return sharedWebSocketAsserter
}

// createSharedWebSocketAsserter creates the actual shared WebSocket integration asserter
// This implements the logic from test_asserters_test.go but as a reusable function
func createSharedWebSocketAsserter(t *testing.T) *WebSocketIntegrationAsserter {
	helper := NewWebSocketTestHelper(t)

	// Create real WebSocket server
	err := helper.CreateRealServer()
	require.NoError(t, err, "Failed to create real WebSocket server")

	// Create WebSocket client
	client := NewWebSocketTestClient(t, helper.GetServerURL())

	asserter := &WebSocketIntegrationAsserter{
		t:      t,
		helper: helper,
		client: client,
	}

	return asserter
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
