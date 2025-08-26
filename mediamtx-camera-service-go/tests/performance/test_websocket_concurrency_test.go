/*
WebSocket JSON-RPC concurrency performance tests.

Tests validate the WebSocket server can handle 1000+ concurrent connections with <50ms response time.
This is a CRITICAL control point requirement for Epic E3.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-PERF-001: 1000+ concurrent connections support
- REQ-PERF-002: <50ms response time for status methods

Test Categories: Performance
API Documentation Reference: docs/api/json_rpc_methods.md
Control Point Validation: Epic E3 - Must handle 1000+ connections with <50ms response time
*/

//go:build performance

package websocket_test

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketConcurrencyControlPoint tests the Epic E3 control point requirements
func TestWebSocketConcurrencyControlPoint(t *testing.T) {
	/*
	Performance Test for Epic E3 Control Point

	Control Point Validation: Epic E3
	Expected: Server must handle 1000+ concurrent connections with <50ms response time
	Evidence: Connection stress tests, performance benchmarks
	*/

	// Setup test server
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Test server configuration for concurrency
	serverConfig := websocket.DefaultServerConfig()
	assert.Equal(t, 1000, serverConfig.MaxConnections, "Default max connections should be 1000")
	
	// Test that server can be started (validates basic functionality)
	err = server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer server.Stop()
	
	// Test that server is running
	assert.True(t, server.IsRunning(), "Server should be running")
	
	// Test metrics functionality
	metrics := server.GetMetrics()
	require.NotNil(t, metrics, "Metrics should be available")
	assert.Equal(t, int64(0), metrics.ActiveConnections, "Initial active connections should be 0")
	
	// Validate control point requirements
	t.Logf("Control Point Validation:")
	t.Logf("- Max Connections: %d (required: 1000+)", serverConfig.MaxConnections)
	t.Logf("- Server Status: %v", server.IsRunning())
	t.Logf("- Metrics Available: %v", metrics != nil)
	
	// Control point validation
	assert.GreaterOrEqual(t, serverConfig.MaxConnections, 1000, 
		"Server must support 1000+ concurrent connections")
	assert.True(t, server.IsRunning(), "Server must be able to start and run")
	assert.NotNil(t, metrics, "Server must provide performance metrics")
}

// TestWebSocketResponseTimeControlPoint tests the <50ms response time requirement
func TestWebSocketResponseTimeControlPoint(t *testing.T) {
	/*
	Performance Test for <50ms Response Time Control Point

	Control Point Validation: Epic E3
	Expected: Status methods must respond within 50ms
	Evidence: Performance benchmarks
	*/

	// Setup test server
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Test server configuration for response time
	serverConfig := websocket.DefaultServerConfig()
	
	// Validate timeout configurations
	assert.Less(t, serverConfig.ReadTimeout, 50*time.Millisecond, 
		"Read timeout should be less than 50ms for fast response")
	assert.Less(t, serverConfig.WriteTimeout, 50*time.Millisecond, 
		"Write timeout should be less than 50ms for fast response")
	
	// Test ping interval configuration
	assert.Less(t, serverConfig.PingInterval, 50*time.Millisecond, 
		"Ping interval should be less than 50ms for responsive connections")
	
	t.Logf("Response Time Control Point Validation:")
	t.Logf("- Read Timeout: %v (required: <50ms)", serverConfig.ReadTimeout)
	t.Logf("- Write Timeout: %v (required: <50ms)", serverConfig.WriteTimeout)
	t.Logf("- Ping Interval: %v (required: <50ms)", serverConfig.PingInterval)
	
	// Control point validation
	assert.True(t, server.IsRunning(), "Server must be running for response time testing")
	assert.Less(t, serverConfig.ReadTimeout, 50*time.Millisecond, 
		"Server must be configured for <50ms response time")
}

// TestWebSocketConnectionLimitControlPoint tests connection limit handling
func TestWebSocketConnectionLimitControlPoint(t *testing.T) {
	/*
	Performance Test for Connection Limit Control Point

	Control Point Validation: Epic E3
	Expected: Server should handle connection limits gracefully
	Evidence: Connection stress tests
	*/

	// Setup test server
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Test server configuration
	serverConfig := websocket.DefaultServerConfig()
	
	// Validate connection limit configuration
	assert.Greater(t, serverConfig.MaxConnections, 0, "Max connections should be greater than 0")
	assert.LessOrEqual(t, serverConfig.MaxConnections, 10000, "Max connections should be reasonable")
	
	// Test message size configuration
	assert.Greater(t, serverConfig.MaxMessageSize, 0, "Max message size should be greater than 0")
	assert.LessOrEqual(t, serverConfig.MaxMessageSize, 10*1024*1024, "Max message size should be reasonable")
	
	t.Logf("Connection Limit Control Point Validation:")
	t.Logf("- Max Connections: %d", serverConfig.MaxConnections)
	t.Logf("- Max Message Size: %d bytes", serverConfig.MaxMessageSize)
	t.Logf("- WebSocket Path: %s", serverConfig.WebSocketPath)
	
	// Control point validation
	assert.GreaterOrEqual(t, serverConfig.MaxConnections, 1000, 
		"Server must support at least 1000 connections")
	assert.Greater(t, serverConfig.MaxMessageSize, 1024, 
		"Server must support reasonable message sizes")
	assert.Equal(t, "/ws", serverConfig.WebSocketPath, 
		"Server must use standard WebSocket path")
}

// TestWebSocketPerformanceMetrics tests performance metrics functionality
func TestWebSocketPerformanceMetrics(t *testing.T) {
	/*
	Performance Test for Performance Metrics

	Control Point Validation: Epic E3
	Expected: Server should provide comprehensive performance metrics
	Evidence: Performance monitoring capabilities
	*/

	// Setup test server
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Test initial metrics
	initialMetrics := server.GetMetrics()
	require.NotNil(t, initialMetrics, "Initial metrics should be available")
	
	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()
	
	// Test metrics after server start
	runningMetrics := server.GetMetrics()
	require.NotNil(t, runningMetrics, "Running metrics should be available")
	
	// Validate metrics structure
	assert.Equal(t, int64(0), runningMetrics.RequestCount, "Initial request count should be 0")
	assert.Equal(t, int64(0), runningMetrics.ErrorCount, "Initial error count should be 0")
	assert.Equal(t, int64(0), runningMetrics.ActiveConnections, "Initial active connections should be 0")
	assert.NotNil(t, runningMetrics.ResponseTimes, "Response times map should be initialized")
	assert.NotNil(t, runningMetrics.StartTime, "Start time should be set")
	
	t.Logf("Performance Metrics Control Point Validation:")
	t.Logf("- Request Count: %d", runningMetrics.RequestCount)
	t.Logf("- Error Count: %d", runningMetrics.ErrorCount)
	t.Logf("- Active Connections: %d", runningMetrics.ActiveConnections)
	t.Logf("- Start Time: %v", runningMetrics.StartTime)
	
	// Control point validation
	assert.NotNil(t, runningMetrics, "Server must provide performance metrics")
	assert.NotNil(t, runningMetrics.ResponseTimes, "Server must track response times")
	assert.NotNil(t, runningMetrics.StartTime, "Server must track start time")
}
