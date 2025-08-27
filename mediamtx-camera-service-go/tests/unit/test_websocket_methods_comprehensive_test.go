//go:build unit
// +build unit

/*
WebSocket JSON-RPC comprehensive method implementation unit tests.

Tests validate all method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-API-004: get_camera_status method for camera status
- REQ-API-005: get_metrics method for system monitoring
- REQ-API-006: get_camera_capabilities method for device capabilities
- REQ-API-007: get_status method for system health
- REQ-API-008: authenticate method for authentication
- REQ-API-009: Role-based access control with viewer, operator, admin permissions
- REQ-API-010: get_server_info method for server information
- REQ-API-011: API methods respond within specified time limits
- REQ-FILE-001: list_recordings method for recording enumeration
- REQ-FILE-002: list_snapshots method for snapshot enumeration
- REQ-FILE-003: delete_recording method for recording deletion
- REQ-FILE-004: delete_snapshot method for snapshot deletion
- REQ-FILE-005: get_storage_info method for storage monitoring
- REQ-FILE-006: set_retention_policy method for policy management
- REQ-FILE-007: cleanup_old_files method for file cleanup
- REQ-REC-001: take_snapshot method for camera snapshots
- REQ-REC-002: start_recording method for video recording
- REQ-REC-003: stop_recording method for recording control
- REQ-REC-004: get_streams method for stream enumeration

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

// =============================================================================
// CORE METHODS (ping, authenticate, get_camera_list, get_camera_status)
// =============================================================================

// TestPingMethodImplementation tests ping method implementation
// REQ-API-002: ping method for health checks
func TestPingMethodImplementation(t *testing.T) {
	/*
		Unit Test for ping method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: ping
		Expected Response: {"jsonrpc": "2.0", "result": "pong", "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test ping method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodPing(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result is "pong" per API documentation
	assert.Equal(t, "pong", response.Result, "Ping result should be 'pong' per API documentation")

	// Validate performance target
	assert.Less(t, responseTime, 50*time.Millisecond, "Ping response should be <50ms per API documentation")
}

// TestAuthenticateMethodImplementation tests authenticate method implementation
// REQ-API-008: authenticate method for authentication
func TestAuthenticateMethodImplementation(t *testing.T) {
	/*
		Unit Test for authenticate method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: authenticate
		Expected Response: {"jsonrpc": "2.0", "result": {"authenticated": true, "role": "operator", "permissions": ["view", "control"], "expires_at": "...", "session_id": "..."}, "id": 2}
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
		Authenticated: false,
		ConnectedAt:   time.Now(),
	}

	// Test authenticate method with valid credentials
	params := map[string]interface{}{
		"username": "test_user",
		"password": "test_password",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodAuthenticate(params, client)
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
	requiredFields := []string{"authenticated", "role", "permissions", "expires_at", "session_id"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["authenticated"], "authenticated should be bool")
	assert.IsType(t, "", result["role"], "role should be string")
	assert.IsType(t, []interface{}{}, result["permissions"], "permissions should be array")
	assert.IsType(t, "", result["expires_at"], "expires_at should be string")
	assert.IsType(t, "", result["session_id"], "session_id should be string")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "authenticate response should be <100ms per API documentation")
}

// TestGetCameraListMethodImplementation tests get_camera_list method implementation
// REQ-API-003: get_camera_list method for camera enumeration
func TestGetCameraListMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_camera_list method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_camera_list
		Expected Response: {"jsonrpc": "2.0", "result": {"cameras": [{"device": "/dev/video0", "name": "USB Camera", "status": "connected", "capabilities": ["snapshot", "recording"]}], "total_count": 1}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test get_camera_list method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetCameraList(params, client)
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
	requiredFields := []string{"cameras", "total_count"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["cameras"], "cameras should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "get_camera_list response should be <200ms per API documentation")
}

// TestGetCameraStatusMethodImplementation tests get_camera_status method implementation
// REQ-API-004: get_camera_status method for camera status
func TestGetCameraStatusMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_camera_status method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_camera_status
		Expected Response: {"jsonrpc": "2.0", "result": {"device": "/dev/video0", "status": "connected", "recording": false, "streaming": true, "last_seen": "2025-01-15T12:00:00Z"}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test get_camera_status method with valid device
	params := map[string]interface{}{
		"device": "/dev/video0",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetCameraStatus(params, client)
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
	requiredFields := []string{"device", "status", "recording", "streaming", "last_seen"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, "", result["device"], "device should be string")
	assert.IsType(t, "", result["status"], "status should be string")
	assert.IsType(t, true, result["recording"], "recording should be bool")
	assert.IsType(t, true, result["streaming"], "streaming should be bool")
	assert.IsType(t, "", result["last_seen"], "last_seen should be string")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_camera_status response should be <100ms per API documentation")

	// Test error case - missing device parameter
	params = map[string]interface{}{}
	response, err = server.MethodGetCameraStatus(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing device parameter
	assert.NotNil(t, response.Error, "Should return error for missing device parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// =============================================================================
// SYSTEM METHODS (get_metrics, get_camera_capabilities, get_status, get_server_info)
// =============================================================================

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

// =============================================================================
// FILE MANAGEMENT METHODS (list_recordings, list_snapshots, delete_recording, delete_snapshot, get_storage_info, set_retention_policy, cleanup_old_files)
// =============================================================================

// TestListRecordingsMethodImplementation tests list_recordings method implementation
// REQ-FILE-001: list_recordings method for recording enumeration
func TestListRecordingsMethodImplementation(t *testing.T) {
	/*
		Unit Test for list_recordings method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: list_recordings
		Expected Response: {"jsonrpc": "2.0", "result": {"recordings": [{"filename": "recording_001.mp4", "size": 1024000, "created": "2025-01-15T12:00:00Z", "duration": 300}], "total_count": 1}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test list_recordings method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodListRecordings(params, client)
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
	requiredFields := []string{"recordings", "total_count"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["recordings"], "recordings should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "list_recordings response should be <200ms per API documentation")
}

// TestListSnapshotsMethodImplementation tests list_snapshots method implementation
// REQ-FILE-002: list_snapshots method for snapshot enumeration
func TestListSnapshotsMethodImplementation(t *testing.T) {
	/*
		Unit Test for list_snapshots method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: list_snapshots
		Expected Response: {"jsonrpc": "2.0", "result": {"snapshots": [{"filename": "snapshot_001.jpg", "size": 51200, "created": "2025-01-15T12:00:00Z", "camera": "/dev/video0"}], "total_count": 1}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test list_snapshots method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodListSnapshots(params, client)
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
	requiredFields := []string{"snapshots", "total_count"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["snapshots"], "snapshots should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "list_snapshots response should be <200ms per API documentation")
}

// TestDeleteRecordingMethodImplementation tests delete_recording method implementation
// REQ-FILE-003: delete_recording method for recording deletion
func TestDeleteRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for delete_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: delete_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"deleted": true, "filename": "recording_001.mp4"}, "id": 1}
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
		Role:          "operator",
		ConnectedAt:   time.Now(),
	}

	// Test delete_recording method with valid filename
	params := map[string]interface{}{
		"filename": "test_recording.mp4",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodDeleteRecording(params, client)
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
	requiredFields := []string{"deleted", "filename"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["deleted"], "deleted should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "delete_recording response should be <100ms per API documentation")

	// Test error case - missing filename parameter
	params = map[string]interface{}{}
	response, err = server.MethodDeleteRecording(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing filename parameter
	assert.NotNil(t, response.Error, "Should return error for missing filename parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestDeleteSnapshotMethodImplementation tests delete_snapshot method implementation
// REQ-FILE-004: delete_snapshot method for snapshot deletion
func TestDeleteSnapshotMethodImplementation(t *testing.T) {
	/*
		Unit Test for delete_snapshot method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: delete_snapshot
		Expected Response: {"jsonrpc": "2.0", "result": {"deleted": true, "filename": "snapshot_001.jpg"}, "id": 1}
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
		Role:          "operator",
		ConnectedAt:   time.Now(),
	}

	// Test delete_snapshot method with valid filename
	params := map[string]interface{}{
		"filename": "test_snapshot.jpg",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodDeleteSnapshot(params, client)
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
	requiredFields := []string{"deleted", "filename"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["deleted"], "deleted should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "delete_snapshot response should be <100ms per API documentation")

	// Test error case - missing filename parameter
	params = map[string]interface{}{}
	response, err = server.MethodDeleteSnapshot(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing filename parameter
	assert.NotNil(t, response.Error, "Should return error for missing filename parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestGetStorageInfoMethodImplementation tests get_storage_info method implementation
// REQ-FILE-005: get_storage_info method for storage monitoring
func TestGetStorageInfoMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_storage_info method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_storage_info
		Expected Response: {"jsonrpc": "2.0", "result": {"total_space": 107374182400, "used_space": 53687091200, "available_space": 53687091200, "recordings_size": 26843545600, "snapshots_size": 13421772800}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test get_storage_info method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetStorageInfo(params, client)
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
	requiredFields := []string{"total_space", "used_space", "available_space", "recordings_size", "snapshots_size"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, int64(0), result["total_space"], "total_space should be int64")
	assert.IsType(t, int64(0), result["used_space"], "used_space should be int64")
	assert.IsType(t, int64(0), result["available_space"], "available_space should be int64")
	assert.IsType(t, int64(0), result["recordings_size"], "recordings_size should be int64")
	assert.IsType(t, int64(0), result["snapshots_size"], "snapshots_size should be int64")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_storage_info response should be <100ms per API documentation")
}

// TestSetRetentionPolicyMethodImplementation tests set_retention_policy method implementation
// REQ-FILE-006: set_retention_policy method for policy management
func TestSetRetentionPolicyMethodImplementation(t *testing.T) {
	/*
		Unit Test for set_retention_policy method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: set_retention_policy
		Expected Response: {"jsonrpc": "2.0", "result": {"updated": true, "policy": {"recordings_days": 30, "snapshots_days": 7, "max_size_gb": 100}}, "id": 1}
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

	// Test set_retention_policy method with valid policy
	params := map[string]interface{}{
		"recordings_days": 30,
		"snapshots_days":  7,
		"max_size_gb":     100,
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodSetRetentionPolicy(params, client)
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
	requiredFields := []string{"updated", "policy"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["updated"], "updated should be bool")
	assert.IsType(t, map[string]interface{}{}, result["policy"], "policy should be map")

	// Validate policy structure
	policy, ok := result["policy"].(map[string]interface{})
	require.True(t, ok, "policy should be a map")

	policyFields := []string{"recordings_days", "snapshots_days", "max_size_gb"}
	for _, field := range policyFields {
		assert.Contains(t, policy, field, "Missing policy field '%s' per API documentation", field)
		assert.IsType(t, float64(0), policy[field], "policy field should be float64")
	}

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "set_retention_policy response should be <100ms per API documentation")

	// Test error case - missing required parameters
	params = map[string]interface{}{}
	response, err = server.MethodSetRetentionPolicy(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing parameters
	assert.NotNil(t, response.Error, "Should return error for missing parameters")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestCleanupOldFilesMethodImplementation tests cleanup_old_files method implementation
// REQ-FILE-007: cleanup_old_files method for file cleanup
func TestCleanupOldFilesMethodImplementation(t *testing.T) {
	/*
		Unit Test for cleanup_old_files method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: cleanup_old_files
		Expected Response: {"jsonrpc": "2.0", "result": {"cleaned": true, "deleted_recordings": 5, "deleted_snapshots": 10, "freed_space": 1073741824}, "id": 1}
		Performance Target: <500ms response time
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

	// Test cleanup_old_files method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodCleanupOldFiles(params, client)
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
	requiredFields := []string{"cleaned", "deleted_recordings", "deleted_snapshots", "freed_space"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["cleaned"], "cleaned should be bool")
	assert.IsType(t, float64(0), result["deleted_recordings"], "deleted_recordings should be float64")
	assert.IsType(t, float64(0), result["deleted_snapshots"], "deleted_snapshots should be float64")
	assert.IsType(t, int64(0), result["freed_space"], "freed_space should be int64")

	// Validate performance target
	assert.Less(t, responseTime, 500*time.Millisecond, "cleanup_old_files response should be <500ms per API documentation")
}

// =============================================================================
// RECORDING AND SNAPSHOT METHODS (take_snapshot, start_recording, stop_recording, get_streams)
// =============================================================================

// TestTakeSnapshotMethodImplementation tests take_snapshot method implementation
// REQ-REC-001: take_snapshot method for camera snapshots
func TestTakeSnapshotMethodImplementation(t *testing.T) {
	/*
		Unit Test for take_snapshot method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: take_snapshot
		Expected Response: {"jsonrpc": "2.0", "result": {"success": true, "filename": "snapshot_20250115_120000.jpg", "size": 51200, "camera": "/dev/video0"}, "id": 1}
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
		Role:          "operator",
		ConnectedAt:   time.Now(),
	}

	// Test take_snapshot method with valid device
	params := map[string]interface{}{
		"device": "/dev/video0",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodTakeSnapshot(params, client)
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
	requiredFields := []string{"success", "filename", "size", "camera"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["success"], "success should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, int64(0), result["size"], "size should be int64")
	assert.IsType(t, "", result["camera"], "camera should be string")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "take_snapshot response should be <200ms per API documentation")

	// Test error case - missing device parameter
	params = map[string]interface{}{}
	response, err = server.MethodTakeSnapshot(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing device parameter
	assert.NotNil(t, response.Error, "Should return error for missing device parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestStartRecordingMethodImplementation tests start_recording method implementation
// REQ-REC-002: start_recording method for video recording
func TestStartRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for start_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: start_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"started": true, "recording_id": "rec_001", "filename": "recording_20250115_120000.mp4", "camera": "/dev/video0"}, "id": 1}
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
		Role:          "operator",
		ConnectedAt:   time.Now(),
	}

	// Test start_recording method with valid device
	params := map[string]interface{}{
		"device": "/dev/video0",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodStartRecording(params, client)
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
	requiredFields := []string{"started", "recording_id", "filename", "camera"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["started"], "started should be bool")
	assert.IsType(t, "", result["recording_id"], "recording_id should be string")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, "", result["camera"], "camera should be string")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "start_recording response should be <100ms per API documentation")

	// Test error case - missing device parameter
	params = map[string]interface{}{}
	response, err = server.MethodStartRecording(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing device parameter
	assert.NotNil(t, response.Error, "Should return error for missing device parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestStopRecordingMethodImplementation tests stop_recording method implementation
// REQ-REC-003: stop_recording method for recording control
func TestStopRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for stop_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: stop_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"stopped": true, "recording_id": "rec_001", "duration": 300, "final_size": 1024000}, "id": 1}
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
		Role:          "operator",
		ConnectedAt:   time.Now(),
	}

	// Test stop_recording method with valid recording_id
	params := map[string]interface{}{
		"recording_id": "test_recording_001",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodStopRecording(params, client)
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
	requiredFields := []string{"stopped", "recording_id", "duration", "final_size"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["stopped"], "stopped should be bool")
	assert.IsType(t, "", result["recording_id"], "recording_id should be string")
	assert.IsType(t, float64(0), result["duration"], "duration should be float64")
	assert.IsType(t, int64(0), result["final_size"], "final_size should be int64")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "stop_recording response should be <100ms per API documentation")

	// Test error case - missing recording_id parameter
	params = map[string]interface{}{}
	response, err = server.MethodStopRecording(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing recording_id parameter
	assert.NotNil(t, response.Error, "Should return error for missing recording_id parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestGetStreamsMethodImplementation tests get_streams method implementation
// REQ-REC-004: get_streams method for stream enumeration
func TestGetStreamsMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_streams method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_streams
		Expected Response: {"jsonrpc": "2.0", "result": {"streams": [{"name": "camera0", "url": "rtsp://localhost:8554/camera0", "status": "active", "viewers": 2}], "total_count": 1}, "id": 1}
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
		Role:          "viewer",
		ConnectedAt:   time.Now(),
	}

	// Test get_streams method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetStreams(params, client)
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
	requiredFields := []string{"streams", "total_count"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["streams"], "streams should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_streams response should be <100ms per API documentation")
}
