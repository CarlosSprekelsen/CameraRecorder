//go:build unit
// +build unit

/*
WebSocket JSON-RPC system methods unit tests.

Tests validate system method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-API-005: get_metrics method for system monitoring
- REQ-API-006: get_camera_capabilities method for device capabilities
- REQ-API-007: get_status method for system health
- REQ-API-010: get_server_info method for server information
- REQ-API-011: API methods respond within specified time limits

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	ws "github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMetricsMethodImplementation tests get_metrics method implementation
// REQ-API-005: get_metrics method for system monitoring
func TestGetMetricsMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_metrics method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_metrics
		Expected Response: {"jsonrpc": "2.0", "result": {"active_connections": 5, "total_requests": 1000, "average_response_time": 45.2, "error_rate": 0.02, "memory_usage": 85.5, "cpu_usage": 23.1}, "id": 1}
		Performance Target: <100ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	
	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}
	
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test get_metrics method
	params := map[string]interface{}{}
	
	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetMetrics(params, client)
	responseTime := time.Since(startTime)
	
	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"active_connections", "total_requests", "average_response_time", "error_rate", "memory_usage", "cpu_usage"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, float64(0), result["active_connections"], "active_connections should be float64")
	assert.IsType(t, int64(0), result["total_requests"], "total_requests should be int64")
	assert.IsType(t, float64(0), result["average_response_time"], "average_response_time should be float64")
	assert.IsType(t, float64(0), result["error_rate"], "error_rate should be float64")
	assert.IsType(t, float64(0), result["memory_usage"], "memory_usage should be float64")
	assert.IsType(t, float64(0), result["cpu_usage"], "cpu_usage should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_metrics response should be <100ms per API documentation")
}

// TestGetCameraCapabilitiesMethodImplementation tests get_camera_capabilities method implementation
// REQ-API-006: get_camera_capabilities method for device capabilities
func TestGetCameraCapabilitiesMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_camera_capabilities method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_camera_capabilities
		Expected Response: {"jsonrpc": "2.0", "result": {"device": "/dev/video0", "formats": ["YUYV", "MJPG"], "resolutions": ["640x480", "1280x720"], "fps_options": ["30", "60"], "validation_status": "valid"}, "id": 1}
		Performance Target: <200ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	
	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}
	
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test get_camera_capabilities method with valid device
	params := map[string]interface{}{
		"device": "/dev/video0",
	}
	
	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetCameraCapabilities(params, client)
	responseTime := time.Since(startTime)
	
	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"device", "formats", "resolutions", "fps_options", "validation_status"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, "", result["device"], "device should be string")
	assert.IsType(t, []interface{}{}, result["formats"], "formats should be array")
	assert.IsType(t, []interface{}{}, result["resolutions"], "resolutions should be array")
	assert.IsType(t, []interface{}{}, result["fps_options"], "fps_options should be array")
	assert.IsType(t, "", result["validation_status"], "validation_status should be string")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "get_camera_capabilities response should be <200ms per API documentation")

	// Test error case - missing device parameter
	params = map[string]interface{}{}
	response, err = server.MethodGetCameraCapabilities(params, client)
	
	require.NoError(t, err)
	require.NotNil(t, response)
	
	// Should return error for missing device parameter
	assert.NotNil(t, response.Error, "Should return error for missing device parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestGetStatusMethodImplementation tests get_status method implementation
// REQ-API-007: get_status method for system health
func TestGetStatusMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_status method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_status
		Expected Response: {"jsonrpc": "2.0", "result": {"status": "healthy", "uptime": 3600, "version": "1.0.0", "components": {"camera_monitor": "running", "websocket_server": "running", "mediamtx": "running"}}, "id": 1}
		Performance Target: <100ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	
	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}
	
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test get_status method
	params := map[string]interface{}{}
	
	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetStatus(params, client)
	responseTime := time.Since(startTime)
	
	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"status", "uptime", "version", "components"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, "", result["status"], "status should be string")
	assert.IsType(t, float64(0), result["uptime"], "uptime should be float64")
	assert.IsType(t, "", result["version"], "version should be string")
	assert.IsType(t, map[string]interface{}{}, result["components"], "components should be map")

	// Validate components structure
	components, ok := result["components"].(map[string]interface{})
	require.True(t, ok, "components should be a map")
	
	componentFields := []string{"camera_monitor", "websocket_server", "mediamtx"}
	for _, field := range componentFields {
		assert.Contains(t, components, field, "Missing component field '%s' per API documentation", field)
		assert.IsType(t, "", components[field], "component status should be string")
	}

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_status response should be <100ms per API documentation")
}

// TestGetServerInfoMethodImplementation tests get_server_info method implementation
// REQ-API-010: get_server_info method for server information
func TestGetServerInfoMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_server_info method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_server_info
		Expected Response: {"jsonrpc": "2.0", "result": {"name": "MediaMTX Camera Service", "version": "1.0.0", "build_date": "2025-01-15", "go_version": "1.21.0", "architecture": "linux/amd64"}, "id": 1}
		Performance Target: <50ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	
	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}
	
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test get_server_info method
	params := map[string]interface{}{}
	
	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetServerInfo(params, client)
	responseTime := time.Since(startTime)
	
	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"name", "version", "build_date", "go_version", "architecture"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, "", result["name"], "name should be string")
	assert.IsType(t, "", result["version"], "version should be string")
	assert.IsType(t, "", result["build_date"], "build_date should be string")
	assert.IsType(t, "", result["go_version"], "go_version should be string")
	assert.IsType(t, "", result["architecture"], "architecture should be string")

	// Validate performance target
	assert.Less(t, responseTime, 50*time.Millisecond, "get_server_info response should be <50ms per API documentation")
}
