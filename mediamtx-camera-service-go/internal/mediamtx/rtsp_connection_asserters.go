/*
RTSP Connection Manager Test Asserters - Eliminate Massive Duplication

This file provides domain-specific asserters for RTSP connection manager tests that eliminate
the massive duplication found in rtsp_connection_manager_test.go (745 lines).

Duplication Patterns Eliminated:
- SetupMediaMTXTest + GetReadyController (16 times)
- RTSP manager initialization (16+ times)
- Connection listing with pagination (8+ times)
- Health and metrics checking (6+ times)
- Error handling patterns (10+ times)

Usage:
    asserter := NewRTSPConnectionAsserter(t)
    defer asserter.Cleanup()
    // Test-specific logic only
    asserter.AssertListConnections(page, itemsPerPage)
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RTSPConnectionAsserter encapsulates all RTSP connection manager test patterns
type RTSPConnectionAsserter struct {
	t           *testing.T
	helper      *MediaMTXTestHelper
	ctx         context.Context
	cancel      context.CancelFunc
	controller  MediaMTXController
	rtspManager RTSPConnectionManager
}

// NewRTSPConnectionAsserter creates a new RTSP connection asserter with full setup
// Eliminates: helper, _ := SetupMediaMTXTest(t) + GetReadyController pattern
func NewRTSPConnectionAsserter(t *testing.T) *RTSPConnectionAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)
	rtspManager := helper.GetRTSPConnectionManager()

	return &RTSPConnectionAsserter{
		t:           t,
		helper:      helper,
		ctx:         ctx,
		cancel:      cancel,
		controller:  controller,
		rtspManager: rtspManager,
	}
}

// Cleanup stops the controller and cancels the context
func (rca *RTSPConnectionAsserter) Cleanup() {
	rca.controller.Stop(rca.ctx)
	rca.cancel()
}

// GetHelper returns the underlying MediaMTXTestHelper
func (rca *RTSPConnectionAsserter) GetHelper() *MediaMTXTestHelper {
	return rca.helper
}

// GetRTSPManager returns the RTSP connection manager instance
func (rca *RTSPConnectionAsserter) GetRTSPManager() RTSPConnectionManager {
	return rca.rtspManager
}

// GetController returns the controller instance
func (rca *RTSPConnectionAsserter) GetController() MediaMTXController {
	return rca.controller
}

// GetContext returns the test context
func (rca *RTSPConnectionAsserter) GetContext() context.Context {
	return rca.ctx
}

// AssertRTSPManagerCreation validates RTSP manager creation
// Eliminates RTSP manager initialization duplication
func (rca *RTSPConnectionAsserter) AssertRTSPManagerCreation() RTSPConnectionManager {
	require.NotNil(rca.t, rca.rtspManager, "RTSP connection manager should not be nil")

	rca.t.Log("✅ RTSP connection manager created successfully")
	return rca.rtspManager
}

// AssertListConnections tests RTSP connection listing with pagination
// Eliminates connection listing duplication
func (rca *RTSPConnectionAsserter) AssertListConnections(page, itemsPerPage int) *RTSPConnectionList {
	connections, err := rca.rtspManager.ListConnections(rca.ctx, page, itemsPerPage)
	rca.helper.AssertStandardResponse(rca.t, connections, err, "ListConnections")
	require.NotNil(rca.t, connections.Items, "Connections items should not be nil")

	rca.t.Logf("✅ RTSP connections listed: page=%d, items_per_page=%d, found=%d, total_pages=%d, total_items=%d",
		page, itemsPerPage, len(connections.Items), connections.PageCount, connections.ItemCount)
	return connections
}

// AssertListSessions tests RTSP session listing with pagination
// Eliminates session listing duplication
func (rca *RTSPConnectionAsserter) AssertListSessions(page, itemsPerPage int) *RTSPConnectionSessionList {
	sessions, err := rca.rtspManager.ListSessions(rca.ctx, page, itemsPerPage)
	rca.helper.AssertStandardResponse(rca.t, sessions, err, "ListSessions")
	require.NotNil(rca.t, sessions.Items, "Sessions items should not be nil")

	rca.t.Logf("✅ RTSP sessions listed: page=%d, items_per_page=%d, found=%d, total_pages=%d, total_items=%d",
		page, itemsPerPage, len(sessions.Items), sessions.PageCount, sessions.ItemCount)
	return sessions
}

// AssertGetConnectionHealth tests RTSP connection health checking
// Eliminates health checking duplication
func (rca *RTSPConnectionAsserter) AssertGetConnectionHealth() *HealthStatus {
	health, err := rca.rtspManager.GetConnectionHealth(rca.ctx)
	rca.helper.AssertStandardResponse(rca.t, health, err, "GetConnectionHealth")
	require.NotNil(rca.t, health, "Connection health should not be nil")

	rca.t.Logf("✅ RTSP connection health retrieved: status=%s", health.Status)
	return health
}

// AssertGetConnectionMetrics tests RTSP connection metrics retrieval
// Eliminates metrics checking duplication
func (rca *RTSPConnectionAsserter) AssertGetConnectionMetrics() map[string]interface{} {
	metrics := rca.rtspManager.GetConnectionMetrics(rca.ctx)
	require.NotNil(rca.t, metrics, "Connection metrics should not be nil")

	rca.t.Logf("✅ RTSP connection metrics retrieved: %d metrics", len(metrics))
	return metrics
}

// AssertConfigurationManagement tests RTSP configuration operations
// Eliminates configuration testing duplication
func (rca *RTSPConnectionAsserter) AssertConfigurationManagement() {
	// Test that we can get connections and sessions (basic configuration validation)
	connections := rca.AssertListConnections(0, 1)
	assert.NotNil(rca.t, connections, "Should be able to list connections")

	sessions := rca.AssertListSessions(0, 1)
	assert.NotNil(rca.t, sessions, "Should be able to list sessions")

	rca.t.Log("✅ RTSP configuration management validated")
}

// AssertErrorHandling tests RTSP error handling scenarios
// Eliminates error handling duplication
func (rca *RTSPConnectionAsserter) AssertErrorHandling() {
	// Test with invalid pagination parameters
	connections, err := rca.rtspManager.ListConnections(rca.ctx, -1, 10)
	assert.Error(rca.t, err, "Should get error for negative page")
	assert.Nil(rca.t, connections, "Connections should be nil on error")

	connections, err = rca.rtspManager.ListConnections(rca.ctx, 0, -1)
	assert.Error(rca.t, err, "Should get error for negative items per page")
	assert.Nil(rca.t, connections, "Connections should be nil on error")

	rca.t.Log("✅ RTSP error handling validated")
}

// AssertPerformanceTest tests RTSP performance scenarios
// Eliminates performance testing duplication
func (rca *RTSPConnectionAsserter) AssertPerformanceTest() {
	start := time.Now()

	// Test multiple rapid operations
	for i := 0; i < 5; i++ {
		connections := rca.AssertListConnections(0, 10)
		assert.NotNil(rca.t, connections, "Connections should not be nil in performance test")

		sessions := rca.AssertListSessions(0, 10)
		assert.NotNil(rca.t, sessions, "Sessions should not be nil in performance test")
	}

	duration := time.Since(start)
	rca.t.Logf("✅ RTSP performance test completed in %v", duration)

	// Performance should be reasonable (less than 5 seconds for 10 operations)
	assert.Less(rca.t, duration, 5*time.Second, "Performance test should complete within 5 seconds")
}

// AssertConcurrentAccess tests concurrent access to RTSP manager
// Eliminates concurrency testing duplication
func (rca *RTSPConnectionAsserter) AssertConcurrentAccess() {
	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			connections, err := rca.rtspManager.ListConnections(rca.ctx, 0, 10)
			if err != nil {
				results <- err
				return
			}
			if connections == nil {
				results <- assert.AnError
				return
			}
			results <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			require.NoError(rca.t, err, "Concurrent RTSP operation should succeed")
		case <-time.After(10 * time.Second):
			rca.t.Fatal("Concurrent RTSP test timed out")
		}
	}

	rca.t.Log("✅ RTSP concurrent access validated")
}

// AssertStressTest tests RTSP stress scenarios
// Eliminates stress testing duplication
func (rca *RTSPConnectionAsserter) AssertStressTest() {
	const numOperations = 20

	start := time.Now()

	for i := 0; i < numOperations; i++ {
		connections := rca.AssertListConnections(0, 5)
		assert.NotNil(rca.t, connections, "Connections should not be nil in stress test")

		// Small delay to prevent overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	duration := time.Since(start)
	rca.t.Logf("✅ RTSP stress test completed: %d operations in %v", numOperations, duration)

	// Stress test should complete within reasonable time
	assert.Less(rca.t, duration, 10*time.Second, "Stress test should complete within 10 seconds")
}

// AssertIntegrationWithController tests RTSP integration with controller
// Eliminates integration testing duplication
func (rca *RTSPConnectionAsserter) AssertIntegrationWithController() {
	// Test that RTSP manager works with the controller
	connections := rca.AssertListConnections(0, 10)
	assert.NotNil(rca.t, connections, "Connections should be available through controller integration")

	// Test controller health
	health, err := rca.controller.GetHealth(rca.ctx)
	rca.helper.AssertStandardResponse(rca.t, health, err, "Controller health check")
	require.NotNil(rca.t, health, "Controller health should not be nil")

	rca.t.Log("✅ RTSP integration with controller validated")
}

// AssertInputValidation tests RTSP input validation
// Eliminates input validation duplication
func (rca *RTSPConnectionAsserter) AssertInputValidation() {
	// Test with invalid pagination parameters
	connections, err := rca.rtspManager.ListConnections(rca.ctx, -1, 10)
	assert.Error(rca.t, err, "Should get error for negative page")
	assert.Nil(rca.t, connections, "Connections should be nil on error")

	connections, err = rca.rtspManager.ListConnections(rca.ctx, 0, -1)
	assert.Error(rca.t, err, "Should get error for negative items per page")
	assert.Nil(rca.t, connections, "Connections should be nil on error")

	rca.t.Log("✅ RTSP input validation validated")
}
