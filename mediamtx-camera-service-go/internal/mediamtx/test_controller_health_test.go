/*
Controller Health Tests - Refactored with Asserters

This file demonstrates the dramatic reduction possible using HealthAsserters.
Original tests had massive duplication of setup, Progressive Readiness, and validation.
Refactored tests focus on business logic only.

Requirements Coverage:
- REQ-MTX-004: Health monitoring
- REQ-MTX-004: System metrics
- REQ-MTX-004: Monitoring capabilities
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_GetHealth_ReqMTX004_Success_Refactored demonstrates health testing with asserters
func TestController_GetHealth_ReqMTX004_Success_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	// Create health asserter with full setup (eliminates 8 lines of setup)
	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Health check with Progressive Readiness built-in (eliminates 20+ lines of readiness handling)
	health := asserter.AssertHealthCheck()

	// Test-specific business logic only
	assert.Equal(t, "HEALTHY", health.Status, "Health should be healthy")

	// Verify component statuses are healthy
	if len(health.Components) > 0 {
		if cameraStatus, exists := health.Components["camera_monitor"]; exists {
			if statusMap, ok := cameraStatus.(map[string]interface{}); ok {
				if status, ok := statusMap["status"].(string); ok {
					assert.Equal(t, "HEALTHY", status, "Camera monitor should be healthy when controller is ready")
				}
			}
		}
	}

	t.Log("Health check completed successfully with Progressive Readiness Pattern")
}

// TestController_GetMetrics_ReqMTX004_Success_Refactored demonstrates metrics testing
func TestController_GetMetrics_ReqMTX004_Success_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Metrics check with built-in validation (eliminates 15+ lines of setup and validation)
	metrics := asserter.AssertMetricsCheck()

	// Test-specific business logic only
	assert.NotNil(t, metrics, "Metrics response should not be nil")
	assert.NotEmpty(t, metrics.Timestamp, "Metrics should have timestamp")

	t.Log("Metrics check completed successfully")
}

// TestController_GetSystemMetrics_ReqMTX004_Success_Refactored demonstrates system metrics
func TestController_GetSystemMetrics_ReqMTX004_Success_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// System metrics check
	systemMetrics, err := asserter.GetReadyController().GetSystemMetrics(asserter.GetContext())
	require.NoError(t, err, "System metrics should be available")
	require.NotNil(t, systemMetrics, "System metrics response should not be nil")

	// Test-specific business logic only
	assert.NotEmpty(t, systemMetrics.Timestamp, "System metrics should have timestamp")
	assert.GreaterOrEqual(t, systemMetrics.CPUUsage, 0.0, "CPU usage should be non-negative")
	assert.GreaterOrEqual(t, systemMetrics.MemoryUsage, 0.0, "Memory usage should be non-negative")

	t.Log("System metrics check completed successfully")
}

// TestController_GetHealth_ReqMTX001_Concurrent_Refactored demonstrates concurrent health testing
func TestController_GetHealth_ReqMTX001_Concurrent_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Test concurrent health checks (eliminates 30+ lines of goroutine setup)
	done := make(chan bool, testutils.UniversalConcurrencyGoroutines)

	for i := 0; i < testutils.UniversalConcurrencyGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine performs health check
			health := asserter.AssertHealthCheck()
			assert.Equal(t, "HEALTHY", health.Status, "Health should be healthy in concurrent test %d", id)

			t.Logf("Concurrent health check %d completed successfully", id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < testutils.UniversalConcurrencyGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(testutils.UniversalTimeoutExtreme):
			t.Fatal("Timeout waiting for concurrent health checks")
		}
	}

	t.Log("All concurrent health checks completed successfully")
}

// TestController_GetHealth_ReqMTX004_Monitoring_Refactored demonstrates monitoring capabilities
func TestController_GetHealth_ReqMTX004_Monitoring_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Perform multiple health checks over time (eliminates 25+ lines of timing logic)
	healthChecks := make([]*GetHealthResponse, 0, 5)

	for i := 0; i < 5; i++ {
		health := asserter.AssertHealthCheck()
		healthChecks = append(healthChecks, health)

		// Small delay between checks
		time.Sleep(testutils.UniversalTimeoutShort)
	}

	// Test-specific business logic: verify consistency
	for i, health := range healthChecks {
		assert.Equal(t, "HEALTHY", health.Status, "Health should be consistent across checks, check %d", i)
		assert.NotEmpty(t, health.Timestamp, "Health should have timestamp, check %d", i)
	}

	t.Log("Monitoring health checks completed successfully")
}

// TestController_GetHealth_ReqMTX004_Monitoring_Extended demonstrates extended monitoring
func TestController_GetHealth_ReqMTX004_Monitoring_Extended_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Test multiple health endpoints (eliminates 20+ lines of endpoint testing setup)
	health := asserter.AssertHealthCheck()
	assert.Equal(t, "HEALTHY", health.Status, "Health should be healthy")

	metrics := asserter.AssertMetricsCheck()
	assert.NotNil(t, metrics, "Metrics should be available")

	systemMetrics, err := asserter.GetReadyController().GetSystemMetrics(asserter.GetContext())
	require.NoError(t, err, "System metrics should be available")
	assert.NotNil(t, systemMetrics, "System metrics should not be nil")

	t.Log("Extended monitoring test completed successfully")
}

// ============================================================================
// HEALTH ERROR TESTS - REQ-MTX-004
// ============================================================================

// TestController_GetHealth_ReqMTX004_ServerDown_Error tests health check when MediaMTX server is unreachable
func TestController_GetHealth_ReqMTX004_ServerDown_Error(t *testing.T) {
	// REQ-MTX-004: Health monitoring with server down error handling

	// Create helper but don't start MediaMTX server (simulate server down)
	helper, _ := SetupMediaMTXTest(t)

	// Get controller without starting MediaMTX (server down scenario)
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed even without server")

	// Create context for health check
	ctx, cancel := context.WithTimeout(context.Background(), testutils.UniversalTimeoutShort)
	defer cancel()

	// Try to get health when server is down
	health, err := controller.GetHealth(ctx)

	// Should get an error about server being unreachable
	assert.Error(t, err, "Health check should fail when server is down")
	assert.Nil(t, health, "Health response should be nil on error")

	// Verify error indicates server issue (connection failure or controller not running)
	if err != nil {
		errorMsg := err.Error()
		assert.True(t,
			containsAny(errorMsg, []string{"connection", "not running", "unreachable"}),
			"Error should indicate server issue: %s", errorMsg)
	}

	t.Log("✅ Server down scenario handled correctly")
}

// TestController_GetHealth_ReqMTX004_Timeout_Error tests health check with context timeout
func TestController_GetHealth_ReqMTX004_Timeout_Error(t *testing.T) {
	// REQ-MTX-004: Health monitoring with timeout error handling

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Create a very short timeout context to simulate timeout
	shortCtx, shortCancel := context.WithTimeout(asserter.GetContext(), 1*time.Millisecond)
	defer shortCancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	// Try health check with expired context
	health, err := asserter.GetReadyController().GetHealth(shortCtx)

	// Should get context timeout error
	assert.Error(t, err, "Health check should fail with timeout")
	assert.Nil(t, health, "Health response should be nil on timeout")

	// Verify it's a context error
	assert.Equal(t, context.DeadlineExceeded, err, "Should be deadline exceeded error")

	t.Log("✅ Timeout scenario handled correctly")
}

// TestController_GetMetrics_ReqMTX004_ServerDown_Error tests metrics check when server is down
func TestController_GetMetrics_ReqMTX004_ServerDown_Error(t *testing.T) {
	// REQ-MTX-004: Health monitoring with metrics server down handling

	// Create helper but don't start MediaMTX server
	helper, _ := SetupMediaMTXTest(t)

	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")

	ctx, cancel := context.WithTimeout(context.Background(), testutils.UniversalTimeoutShort)
	defer cancel()

	// Try to get metrics when server is down
	metrics, err := controller.GetMetrics(ctx)

	// Should get an error about server being unreachable
	assert.Error(t, err, "Metrics check should fail when server is down")
	assert.Nil(t, metrics, "Metrics response should be nil on error")

	// Verify error indicates server issue (connection failure or controller not running)
	if err != nil {
		errorMsg := err.Error()
		assert.True(t,
			containsAny(errorMsg, []string{"connection", "not running", "unreachable"}),
			"Error should indicate server issue: %s", errorMsg)
	}

	t.Log("✅ Metrics server down scenario handled correctly")
}

// TestController_GetSystemMetrics_ReqMTX004_Timeout_Error tests system metrics with timeout
func TestController_GetSystemMetrics_ReqMTX004_Timeout_Error(t *testing.T) {
	// REQ-MTX-004: Health monitoring with system metrics timeout handling

	asserter := NewHealthAsserter(t)
	defer asserter.Cleanup()

	// Create a very short timeout context
	shortCtx, shortCancel := context.WithTimeout(asserter.GetContext(), 1*time.Millisecond)
	defer shortCancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	// Try system metrics with expired context
	metrics, err := asserter.GetReadyController().GetSystemMetrics(shortCtx)

	// Should get context timeout error
	assert.Error(t, err, "System metrics should fail with timeout")
	assert.Nil(t, metrics, "System metrics should be nil on timeout")

	// Verify it's a context error
	assert.Equal(t, context.DeadlineExceeded, err, "Should be deadline exceeded error")

	t.Log("✅ System metrics timeout scenario handled correctly")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) &&
			(s == substr ||
				len(s) > len(substr) &&
					(s[:len(substr)] == substr ||
						s[len(s)-len(substr):] == substr ||
						indexOf(s, substr) >= 0)) {
			return true
		}
	}
	return false
}

// indexOf finds the index of a substring in a string
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
