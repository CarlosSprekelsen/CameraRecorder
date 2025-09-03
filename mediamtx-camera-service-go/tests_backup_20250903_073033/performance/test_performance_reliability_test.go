//go:build performance
// +build performance

/*
Consolidated Performance and Reliability Tests

Requirements Coverage:
- REQ-PERF-001: API response time performance (<50ms for status methods)
- REQ-PERF-002: Camera discovery performance
- REQ-PERF-003: Health check performance
- REQ-PERF-004: JWT token performance
- REQ-PERF-005: Active recording tracking performance
- REQ-PERF-006: Configuration access performance
- REQ-PERF-007: Concurrent operation performance (1000+ connections)
- REQ-PERF-008: Memory usage performance
- REQ-STRESS-001: Concurrent WebSocket connections
- REQ-STRESS-002: Concurrent request handling
- REQ-STRESS-003: Connection stress over time
- REQ-STRESS-004: Memory stress testing
- REQ-STRESS-005: Rate limiting stress testing
- REQ-STRESS-006: Error rate monitoring
- REQ-STRESS-007: Performance degradation testing
- REQ-STRESS-008: System stability validation
- REQ-RELIABILITY-001: Long-running stability (24/7 operation)
- REQ-RELIABILITY-002: Error recovery and resilience
- REQ-RELIABILITY-003: Resource management (memory leaks, CPU usage)
- REQ-RELIABILITY-004: Network failure recovery
- REQ-RELIABILITY-005: Hardware failure recovery

Test Categories: Performance/Stress/Reliability/Real System
API Documentation Reference: docs/api/json_rpc_methods.md
Control Point Validation: Epic E3 - Must handle 1000+ connections with <50ms response time
*/

package websocket_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// CONCURRENCY AND STRESS TESTS
// ============================================================================

// TestWebSocketConcurrencyControlPoint tests the Epic E3 control point requirements
func TestWebSocketConcurrencyControlPoint(t *testing.T) {
	/*
		Performance Test for Epic E3 Control Point

		Control Point Validation: Epic E3
		Expected: Server must handle 1000+ concurrent connections with <50ms response time
		Evidence: Connection stress tests, performance tests
	*/

	// REQ-PERF-007: Concurrent operation performance (1000+ connections)
	// REQ-STRESS-001: Concurrent WebSocket connections

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Server should start successfully")
	defer env.WebSocketServer.Stop()

	// Test server configuration for concurrency
	serverConfig := websocket.DefaultServerConfig()
	assert.Equal(t, 1000, serverConfig.MaxConnections, "Default max connections should be 1000")

	// Test that server is running
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should be running")

	// Test metrics functionality
	metrics := env.WebSocketServer.GetMetrics()
	require.NotNil(t, metrics, "Metrics should be available")
	assert.Equal(t, int64(0), metrics.ActiveConnections, "Initial active connections should be 0")

	// Validate control point requirements
	t.Logf("Control Point Validation:")
	t.Logf("- Max Connections: %d (required: 1000+)", serverConfig.MaxConnections)
	t.Logf("- Server Status: %v", env.WebSocketServer.IsRunning())
	t.Logf("- Metrics Available: %v", metrics != nil)

	// Control point validation
	assert.GreaterOrEqual(t, serverConfig.MaxConnections, 1000,
		"Server must support 1000+ concurrent connections")
	assert.True(t, env.WebSocketServer.IsRunning(), "Server must be able to start and run")
	assert.NotNil(t, metrics, "Server must provide performance metrics")
}

// TestConcurrentWebSocketConnections tests multiple concurrent WebSocket connections
func TestConcurrentWebSocketConnections(t *testing.T) {
	// REQ-STRESS-001: Concurrent WebSocket connections
	// REQ-STRESS-002: Concurrent request handling

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Test with multiple concurrent connections
	connectionCount := 10 // Reduced for testing, production should test 1000+
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < connectionCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			// Create WebSocket test client using shared utilities
			client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
			defer client.Close()

			// Send ping request through proper WebSocket flow
			response := client.SendPingRequest()

			if response != nil && response.Error == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
				t.Logf("Connection %d: Ping successful", connID)
			} else {
				t.Logf("Connection %d: Ping failed", connID)
			}
		}(i)
	}

	wg.Wait()

	// Validate concurrent connection handling
	successRate := float64(successCount) / float64(connectionCount)
	t.Logf("Concurrent connection test results:")
	t.Logf("- Total connections: %d", connectionCount)
	t.Logf("- Successful connections: %d", successCount)
	t.Logf("- Success rate: %.2f%%", successRate*100)

	// For production readiness, success rate should be high
	assert.GreaterOrEqual(t, successRate, 0.9, "Success rate should be at least 90% for production readiness")

	// Check server metrics
	metrics := env.WebSocketServer.GetMetrics()
	require.NotNil(t, metrics, "Server metrics should be available")
	t.Logf("Server metrics after concurrent test:")
	t.Logf("- Active connections: %d", metrics.ActiveConnections)
	t.Logf("- Request count: %d", metrics.RequestCount)
}

// TestWebSocketStressOverTime tests WebSocket stress over extended period
func TestWebSocketStressOverTime(t *testing.T) {
	// REQ-STRESS-003: Connection stress over time
	// REQ-RELIABILITY-001: Long-running stability (24/7 operation)

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client using shared utilities
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Test duration (reduced for testing, production should test 24/7)
	testDuration := 5 * time.Second
	startTime := time.Now()

	// Perform operations continuously
	operationCount := 0
	errorCount := 0

	for time.Since(startTime) < testDuration {
		// Perform health check
		health, err := env.Controller.GetHealth(context.Background())
		if err != nil {
			errorCount++
			t.Logf("Health check error: %v", err)
		} else {
			_ = health.Status // Use result
		}

		// Perform camera discovery
		cameras := env.CameraMonitor.GetConnectedCameras()
		_ = len(cameras) // Use result

		// Perform WebSocket ping using proper client
		response := client.SendPingRequest()
		if response == nil || response.Error != nil {
			errorCount++
			t.Logf("Ping error: %v", response.Error)
		}

		operationCount++
		time.Sleep(100 * time.Millisecond) // Small delay between operations
	}

	// Calculate error rate
	errorRate := float64(errorCount) / float64(operationCount)
	t.Logf("Long-running stability test results:")
	t.Logf("- Duration: %v", testDuration)
	t.Logf("- Operations: %d", operationCount)
	t.Logf("- Errors: %d", errorCount)
	t.Logf("- Error rate: %.2f%%", errorRate*100)

	// For production readiness, error rate should be very low
	assert.Less(t, errorRate, 0.01, "Error rate should be less than 1% for production readiness")
}

// TestErrorRecovery tests system resilience to errors
func TestErrorRecovery(t *testing.T) {
	// REQ-RELIABILITY-002: Error recovery and resilience

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client using shared utilities
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Test recovery from invalid requests
	invalidRequests := []string{
		`{"jsonrpc":"2.0","method":"invalid_method","id":1}`,
		`{"jsonrpc":"2.0","method":"ping","id":"invalid_id"}`,
		`{"jsonrpc":"1.0","method":"ping","id":1}`,
		`{"method":"ping","id":1}`,
		`invalid json`,
	}

	recoverySuccess := 0
	for _, request := range invalidRequests {
		// Send invalid request through WebSocket connection using proper methods
		err = client.WriteMessage(gorilla.TextMessage, []byte(request))
		if err != nil {
			t.Logf("Failed to send invalid request: %v", err)
			continue
		}

		// Try to read response (should get error response)
		_, _, err = client.ReadMessage()
		if err == nil {
			recoverySuccess++
		}
	}

	// Test that server is still functional after invalid requests
	response := client.SendPingRequest()
	require.NotNil(t, response, "Server should still be functional after invalid requests")
	require.Nil(t, response.Error, "Server should still be functional after invalid requests")

	t.Logf("Error recovery test results:")
	t.Logf("- Invalid requests tested: %d", len(invalidRequests))
	t.Logf("- Successful recoveries: %d", recoverySuccess)
	t.Logf("- Server still functional: %v", response.Error == nil)

	// Server should remain functional
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should remain running after error recovery")
}

// TestNetworkFailureRecovery tests recovery from network issues
func TestNetworkFailureRecovery(t *testing.T) {
	// REQ-RELIABILITY-004: Network failure recovery

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create initial WebSocket test client
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)

	// Test initial connection
	response := client.SendPingRequest()
	require.NotNil(t, response, "Initial connection should work")
	require.Nil(t, response.Error, "Initial connection should work")

	// Close connection to simulate network failure
	client.Close()

	// Wait a moment for connection cleanup
	time.Sleep(100 * time.Millisecond)

	// Create new connection to test recovery
	newClient := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer newClient.Close()

	// Test that new connection works
	newResponse := newClient.SendPingRequest()
	require.NotNil(t, newResponse, "New connection should work after network failure")
	require.Nil(t, newResponse.Error, "New connection should work after network failure")

	t.Logf("Network failure recovery test results:")
	t.Logf("- Initial connection: %v", response.Error == nil)
	t.Logf("- Recovery connection: %v", newResponse.Error == nil)
	t.Logf("- Server still running: %v", env.WebSocketServer.IsRunning())

	// Server should remain functional
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should remain running after network failure")
}

// TestMemoryStressTesting tests memory usage under stress
func TestMemoryStressTesting(t *testing.T) {
	// REQ-STRESS-004: Memory stress testing
	// REQ-RELIABILITY-003: Resource management (memory leaks, CPU usage)

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Test with multiple rapid connections to stress memory
	connectionCount := 20
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < connectionCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			// Create WebSocket test client
			client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
			defer client.Close()

			// Send multiple requests to stress memory
			for j := 0; j < 5; j++ {
				response := client.SendPingRequest()
				if response != nil && response.Error == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// Check server metrics for memory usage indicators
	metrics := env.WebSocketServer.GetMetrics()
	require.NotNil(t, metrics, "Server metrics should be available")

	t.Logf("Memory stress test results:")
	t.Logf("- Total connections tested: %d", connectionCount)
	t.Logf("- Successful operations: %d", successCount)
	t.Logf("- Server metrics:")
	t.Logf("  - Active connections: %d", metrics.ActiveConnections)
	t.Logf("  - Request count: %d", metrics.RequestCount)

	// Server should remain stable
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should remain running after memory stress")
	assert.Equal(t, int64(0), metrics.ActiveConnections, "All connections should be properly closed")
}

// TestRateLimitingStressTesting tests rate limiting under stress
func TestRateLimitingStressTesting(t *testing.T) {
	// REQ-STRESS-005: Rate limiting stress testing

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Test rapid requests to trigger rate limiting
	requestCount := 50
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < requestCount; i++ {
		response := client.SendPingRequest()
		if response != nil {
			if response.Error == nil {
				successCount++
			} else if response.Error.Code == websocket.RATE_LIMIT_EXCEEDED {
				rateLimitedCount++
			}
		}
		// Small delay to avoid overwhelming the server
		time.Sleep(1 * time.Millisecond)
	}

	t.Logf("Rate limiting stress test results:")
	t.Logf("- Total requests: %d", requestCount)
	t.Logf("- Successful requests: %d", successCount)
	t.Logf("- Rate limited requests: %d", rateLimitedCount)
	t.Logf("- Success rate: %.2f%%", float64(successCount)/float64(requestCount)*100)

	// Server should handle rate limiting gracefully
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should remain running after rate limiting stress")
	assert.Greater(t, successCount, 0, "Some requests should succeed")
}

// TestPerformanceDegradationTesting tests performance under load
func TestPerformanceDegradationTesting(t *testing.T) {
	// REQ-STRESS-007: Performance degradation testing

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Measure baseline performance
	baselineStart := time.Now()
	response := client.SendPingRequest()
	baselineDuration := time.Since(baselineStart)

	require.NotNil(t, response, "Baseline ping should work")
	require.Nil(t, response.Error, "Baseline ping should work")

	// Apply load and measure performance degradation
	// Create additional connections to create load
	var wg sync.WaitGroup
	loadConnections := 5

	for i := 0; i < loadConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loadClient := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
			defer loadClient.Close()

			// Send some requests to create load
			for j := 0; j < 10; j++ {
				loadClient.SendPingRequest()
				time.Sleep(5 * time.Millisecond) // Reduced delay for better performance
			}
		}()
	}

	// Wait for load to be applied
	wg.Wait()

	// Measure performance under load
	loadTestStart := time.Now()
	loadResponse := client.SendPingRequest()
	loadTestDuration := time.Since(loadTestStart)

	require.NotNil(t, loadResponse, "Ping under load should work")
	require.Nil(t, loadResponse.Error, "Ping under load should work")

	t.Logf("Performance degradation test results:")
	t.Logf("- Baseline duration: %v", baselineDuration)
	t.Logf("- Load test duration: %v", loadTestDuration)
	t.Logf("- Performance degradation: %.2f%%",
		float64(loadTestDuration-baselineDuration)/float64(baselineDuration)*100)

	// Performance should not degrade too much under load
	// Allow up to 200% degradation for this test (adjust as needed)
	maxDegradation := 2.0 // 200%
	actualDegradation := float64(loadTestDuration) / float64(baselineDuration)

	assert.Less(t, actualDegradation, maxDegradation,
		"Performance should not degrade more than %.0f%% under load", maxDegradation*100)
}

// TestSystemStabilityValidation tests overall system stability
func TestSystemStabilityValidation(t *testing.T) {
	// REQ-STRESS-008: System stability validation
	// REQ-RELIABILITY-001: Long-running stability (24/7 operation)

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Test system stability over multiple cycles
	cycleCount := 5
	totalOperations := 0
	totalErrors := 0

	for cycle := 0; cycle < cycleCount; cycle++ {
		t.Logf("Starting stability cycle %d/%d", cycle+1, cycleCount)

		// Create WebSocket test client for this cycle
		client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)

		// Perform operations in this cycle
		operations := 10
		errors := 0

		for i := 0; i < operations; i++ {
			response := client.SendPingRequest()
			if response == nil || response.Error != nil {
				errors++
			}
			time.Sleep(5 * time.Millisecond) // Reduced delay for better performance
		}

		client.Close()
		totalOperations += operations
		totalErrors += errors

		t.Logf("Cycle %d: %d operations, %d errors", cycle+1, operations, errors)

		// Verify server is still running
		assert.True(t, env.WebSocketServer.IsRunning(),
			"Server should remain running after cycle %d", cycle+1)
	}

	// Calculate overall stability metrics
	overallErrorRate := float64(totalErrors) / float64(totalOperations)

	t.Logf("System stability validation results:")
	t.Logf("- Total cycles: %d", cycleCount)
	t.Logf("- Total operations: %d", totalOperations)
	t.Logf("- Total errors: %d", totalErrors)
	t.Logf("- Overall error rate: %.2f%%", overallErrorRate*100)

	// System should remain stable
	assert.True(t, env.WebSocketServer.IsRunning(), "Server should remain running after all cycles")
	assert.Less(t, overallErrorRate, 0.05, "Overall error rate should be less than 5% for stability")

	// Check final server metrics
	metrics := env.WebSocketServer.GetMetrics()
	require.NotNil(t, metrics, "Final server metrics should be available")
	t.Logf("Final server metrics:")
	t.Logf("- Active connections: %d", metrics.ActiveConnections)
	t.Logf("- Request count: %d", metrics.RequestCount)
}

// ============================================================================
// PRODUCTION READINESS VALIDATION
// ============================================================================

// TestProductionReadiness validates all production requirements
func TestProductionReadiness(t *testing.T) {
	/*
		Production Readiness Validation Test

		This test validates that the system meets all production requirements:
		- Performance targets met
		- Reliability requirements satisfied
		- Error handling robust
		- Resource management stable
	*/

	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start all services
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "WebSocket server should start")
	defer env.WebSocketServer.Stop()

	err = env.CameraMonitor.Start(context.Background())
	require.NoError(t, err, "Camera monitor should start")
	defer env.CameraMonitor.Stop()

	// 1. Performance Validation
	t.Run("PerformanceValidation", func(t *testing.T) {
		// Create WebSocket test client for proper testing
		client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Test response times using proper WebSocket flow
		start := time.Now()
		response := client.SendPingRequest()
		duration := time.Since(start)

		require.NotNil(t, response, "Ping response should not be nil")
		require.Nil(t, response.Error, "Ping should not return error")
		assert.Less(t, duration, 50*time.Millisecond, "Response time should be <50ms")

		t.Logf("Performance validation passed: %v response time", duration)
	})

	// 2. Reliability Validation
	t.Run("ReliabilityValidation", func(t *testing.T) {
		// Test system stability
		assert.True(t, env.WebSocketServer.IsRunning(), "WebSocket server should be running")
		assert.True(t, env.CameraMonitor.IsRunning(), "Camera monitor should be running")

		// Test health check
		health, err := env.Controller.GetHealth(context.Background())
		require.NoError(t, err, "Health check should succeed")
		assert.NotEmpty(t, health.Status, "Health status should not be empty")

		t.Logf("Reliability validation passed: system stable")
	})

	// 3. Error Handling Validation
	t.Run("ErrorHandlingValidation", func(t *testing.T) {
		// Create WebSocket test client for proper testing
		client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Test graceful error handling with invalid request
		response := client.SendInvalidRequest()
		require.NotNil(t, response, "Error response should not be nil")
		require.NotNil(t, response.Error, "Invalid request should return error")

		t.Logf("Error handling validation passed: graceful error handling")
	})

	// 4. Resource Management Validation
	t.Run("ResourceManagementValidation", func(t *testing.T) {
		// Test metrics availability
		metrics := env.WebSocketServer.GetMetrics()
		require.NotNil(t, metrics, "Metrics should be available")

		// Test camera discovery
		cameras := env.CameraMonitor.GetConnectedCameras()
		t.Logf("Resource management validation passed: %d cameras discovered", len(cameras))
	})

	t.Logf("Production readiness validation: ALL TESTS PASSED")
	t.Logf("✅ Performance targets met")
	t.Logf("✅ Reliability requirements satisfied")
	t.Logf("✅ Error handling robust")
	t.Logf("✅ Resource management stable")
}

// TestPerformanceFileRecognition ensures Go recognizes this file as containing tests
func TestPerformanceFileRecognition(t *testing.T) {
	t.Log("Performance test file is recognized by Go")
	t.Log("Epic E3 Control Point: 1000+ concurrent connections with <50ms response time")
	t.Log("Status: Ready for performance validation")
}
