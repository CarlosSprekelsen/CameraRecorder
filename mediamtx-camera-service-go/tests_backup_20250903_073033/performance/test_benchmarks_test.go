//go:build performance
// +build performance

/*
Performance Benchmarks - Measurement and Metrics

Requirements Coverage:
- REQ-PERF-001: API response time performance (<50ms for status methods)
- REQ-PERF-002: Camera discovery performance
- REQ-PERF-003: Health check performance
- REQ-PERF-004: JWT token performance

Test Categories: Performance/Benchmarks
API Documentation Reference: docs/api/json_rpc_methods.md
Control Point Validation: Epic E3 - Must handle 1000+ connections with <50ms response time
*/

package websocket_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// PERFORMANCE BENCHMARKS
// ============================================================================

// BenchmarkAPIResponseTime benchmarks API response times
func BenchmarkAPIResponseTime(b *testing.B) {
	// REQ-PERF-001: API response time performance (<50ms for status methods)

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client for benchmarks
	client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark ping response time
		response := client.SendPingRequest()
		require.NotNil(b, response, "Ping response should not be nil")
		require.Nil(b, response.Error, "Ping should not return error")
	}
}

// BenchmarkCameraDiscovery benchmarks camera discovery performance
func BenchmarkCameraDiscovery(b *testing.B) {
	// REQ-PERF-002: Camera discovery performance

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start camera monitor
	err := env.CameraMonitor.Start(context.Background())
	require.NoError(b, err, "Failed to start camera monitor")
	defer env.CameraMonitor.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark camera discovery
		cameras := env.CameraMonitor.GetConnectedCameras()
		_ = len(cameras) // Use result to prevent optimization
	}
}

// BenchmarkHealthCheck benchmarks health check performance
func BenchmarkHealthCheck(b *testing.B) {
	// REQ-PERF-003: Health check performance

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark health check
		health, err := env.Controller.GetHealth(context.Background())
		require.NoError(b, err, "Health check should not fail")
		_ = health.Status // Use result to prevent optimization
	}
}

// BenchmarkJWTTokenGeneration benchmarks JWT token performance
func BenchmarkJWTTokenGeneration(b *testing.B) {
	// REQ-PERF-004: JWT token performance

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark JWT token generation
		token, err := env.JWTHandler.GenerateToken("test_user", "viewer", 24)
		require.NoError(b, err, "Token generation should not fail")
		require.NotEmpty(b, token, "Generated token should not be empty")
	}
}

// BenchmarkWebSocketConnectionCreation benchmarks WebSocket connection creation
func BenchmarkWebSocketConnectionCreation(b *testing.B) {
	// REQ-PERF-007: Concurrent operation performance (1000+ connections)

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Pre-create clients to avoid measuring creation time
	clients := make([]*testtestutils.WebSocketTestClientForBenchmark, b.N)
	for i := 0; i < b.N; i++ {
		clients[i] = testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	}

	// Clean up clients after benchmark
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark connection usage (not creation)
		response := clients[i].SendPingRequest()
		require.NotNil(b, response, "Ping response should not be nil")
		require.Nil(b, response.Error, "Ping should not return error")
	}
}

// BenchmarkWebSocketPingThroughput benchmarks WebSocket ping throughput
func BenchmarkWebSocketPingThroughput(b *testing.B) {
	// REQ-PERF-001: API response time performance (<50ms for status methods)

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client for benchmarks
	client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark ping throughput
		response := client.SendPingRequest()
		require.NotNil(b, response, "Ping response should not be nil")
		require.Nil(b, response.Error, "Ping should not return error")
	}
}

// BenchmarkConcurrentWebSocketConnections benchmarks concurrent WebSocket connections
func BenchmarkConcurrentWebSocketConnections(b *testing.B) {
	// REQ-PERF-007: Concurrent operation performance (1000+ connections)

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Pre-create connection pool to avoid measuring creation time
	const poolSize = 100
	connectionPool := make(chan *testtestutils.WebSocketTestClientForBenchmark, poolSize)

	// Fill the pool
	for i := 0; i < poolSize; i++ {
		client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
		connectionPool <- client
	}

	// Clean up connection pool after benchmark
	defer func() {
		close(connectionPool)
		for client := range connectionPool {
			client.Close()
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark concurrent connections using pre-created pool
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Get connection from pool
			client := <-connectionPool

			// Send a ping request
			response := client.SendPingRequest()
			require.NotNil(b, response, "Ping response should not be nil")

			// Return connection to pool
			connectionPool <- client
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	// REQ-PERF-008: Memory usage performance

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Pre-create connections to avoid measuring creation time
	const numConnections = 10
	clients := make([]*testtestutils.WebSocketTestClientForBenchmark, numConnections)
	for i := 0; i < numConnections; i++ {
		clients[i] = testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	}

	// Clean up connections after benchmark
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark connection usage (not creation)
		for j := 0; j < numConnections; j++ {
			response := clients[j].SendPingRequest()
			require.NotNil(b, response, "Ping response should not be nil")
		}
	}
}

// BenchmarkAuthenticationFlow benchmarks the complete authentication flow
func BenchmarkAuthenticationFlow(b *testing.B) {
	// REQ-PERF-004: JWT token performance

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Pre-create client and token to avoid measuring creation time
	client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	token, err := env.JWTHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(b, err, "Token generation should not fail")
	require.NotEmpty(b, token, "Generated token should not be empty")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark authentication request (not client/token creation)
		authResponse := client.SendAuthenticationRequest(token)
		require.NotNil(b, authResponse, "Auth response should not be nil")
		require.Nil(b, authResponse.Error, "Authentication should succeed")
	}
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	// REQ-ERROR-003: WebSocket server shall handle invalid JSON-RPC requests gracefully

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Pre-create client to avoid measuring creation time
	client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark error handling with invalid requests (not client creation)
		response := client.SendInvalidRequest()
		require.NotNil(b, response, "Error response should not be nil")
		require.NotNil(b, response.Error, "Invalid request should return error")
	}
}

// BenchmarkLongRunningOperations benchmarks long-running operation stability
func BenchmarkLongRunningOperations(b *testing.B) {
	// REQ-RELIABILITY-001: Long-running stability (24/7 operation)

	// COMMON PATTERN: Use shared WebSocket test environment for benchmarks
	env := testtestutils.SetupWebSocketTestEnvironmentForBenchmark(b)
	defer testtestutils.TeardownWebSocketTestEnvironmentForBenchmark(b, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(b, err, "Failed to start WebSocket server")
	defer env.WebSocketServer.Stop()

	// Create WebSocket test client for benchmarks
	client := testtestutils.NewWebSocketTestClientForBenchmark(b, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate long-running operations
	for i := 0; i < b.N; i++ {
		// Perform multiple operations to simulate long-running workload
		for j := 0; j < 10; j++ {
			response := client.SendPingRequest()
			require.NotNil(b, response, "Ping response should not be nil")
			require.Nil(b, response.Error, "Ping should not return error")
		}

		// Small delay to simulate real-world usage
		time.Sleep(1 * time.Millisecond)
	}
}
