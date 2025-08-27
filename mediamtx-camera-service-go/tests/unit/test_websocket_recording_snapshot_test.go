//go:build unit
// +build unit

/*
WebSocket JSON-RPC recording and snapshot methods unit tests.

Tests validate recording and snapshot method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-REC-001: take_snapshot method for camera snapshots
- REQ-REC-002: start_recording method for video recording
- REQ-REC-003: stop_recording method for recording control
- REQ-REC-004: get_streams method for stream enumeration
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

// TestTakeSnapshotMethodImplementation tests take_snapshot method implementation
// REQ-REC-001: take_snapshot method for camera snapshots
func TestTakeSnapshotMethodImplementation(t *testing.T) {
	/*
		Unit Test for take_snapshot method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: take_snapshot
		Expected Response: {"jsonrpc": "2.0", "result": {"snapshot_taken": true, "filename": "snapshot_20250115_120000.jpg", "size": 51200, "camera": "/dev/video0", "timestamp": "2025-01-15T12:00:00Z"}, "id": 1}
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

	// Test take_snapshot method with valid camera
	params := map[string]interface{}{
		"camera": "/dev/video0",
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
	requiredFields := []string{"snapshot_taken", "filename", "size", "camera", "timestamp"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["snapshot_taken"], "snapshot_taken should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, float64(0), result["size"], "size should be float64")
	assert.IsType(t, "", result["camera"], "camera should be string")
	assert.IsType(t, "", result["timestamp"], "timestamp should be string")

	// Validate performance target
	assert.Less(t, responseTime, 500*time.Millisecond, "take_snapshot response should be <500ms per API documentation")

	// Test error case - missing camera parameter
	params = map[string]interface{}{}
	response, err = server.MethodTakeSnapshot(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing camera parameter
	assert.NotNil(t, response.Error, "Should return error for missing camera parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestStartRecordingMethodImplementation tests start_recording method implementation
// REQ-REC-002: start_recording method for video recording
func TestStartRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for start_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: start_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"recording_started": true, "recording_id": "rec_20250115_120000", "filename": "recording_20250115_120000.mp4", "camera": "/dev/video0", "start_time": "2025-01-15T12:00:00Z"}, "id": 1}
		Performance Target: <1000ms response time
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

	// Test start_recording method with valid camera
	params := map[string]interface{}{
		"camera": "/dev/video0",
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
	requiredFields := []string{"recording_started", "recording_id", "filename", "camera", "start_time"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["recording_started"], "recording_started should be bool")
	assert.IsType(t, "", result["recording_id"], "recording_id should be string")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, "", result["camera"], "camera should be string")
	assert.IsType(t, "", result["start_time"], "start_time should be string")

	// Validate performance target
	assert.Less(t, responseTime, 1000*time.Millisecond, "start_recording response should be <1000ms per API documentation")

	// Test error case - missing camera parameter
	params = map[string]interface{}{}
	response, err = server.MethodStartRecording(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing camera parameter
	assert.NotNil(t, response.Error, "Should return error for missing camera parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestStopRecordingMethodImplementation tests stop_recording method implementation
// REQ-REC-003: stop_recording method for recording control
func TestStopRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for stop_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: stop_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"recording_stopped": true, "recording_id": "rec_20250115_120000", "filename": "recording_20250115_120000.mp4", "duration": 300, "size": 1024000, "stop_time": "2025-01-15T12:05:00Z"}, "id": 1}
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

	// Test stop_recording method with valid recording_id
	params := map[string]interface{}{
		"recording_id": "rec_20250115_120000",
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
	requiredFields := []string{"recording_stopped", "recording_id", "filename", "duration", "size", "stop_time"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["recording_stopped"], "recording_stopped should be bool")
	assert.IsType(t, "", result["recording_id"], "recording_id should be string")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, float64(0), result["duration"], "duration should be float64")
	assert.IsType(t, float64(0), result["size"], "size should be float64")
	assert.IsType(t, "", result["stop_time"], "stop_time should be string")

	// Validate performance target
	assert.Less(t, responseTime, 500*time.Millisecond, "stop_recording response should be <500ms per API documentation")

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
		Expected Response: {"jsonrpc": "2.0", "result": {"streams": [{"stream_id": "stream_1", "camera": "/dev/video0", "url": "rtsp://localhost:8554/stream_1", "status": "active", "viewers": 2}], "total_streams": 1, "active_streams": 1}, "id": 1}
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
	requiredFields := []string{"streams", "total_streams", "active_streams"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["streams"], "streams should be array")
	assert.IsType(t, float64(0), result["total_streams"], "total_streams should be float64")
	assert.IsType(t, float64(0), result["active_streams"], "active_streams should be float64")

	// Validate stream structure if streams exist
	streams, ok := result["streams"].([]interface{})
	if ok && len(streams) > 0 {
		stream, ok := streams[0].(map[string]interface{})
		require.True(t, ok, "stream should be a map")

		streamFields := []string{"stream_id", "camera", "url", "status", "viewers"}
		for _, field := range streamFields {
			assert.Contains(t, stream, field, "Missing stream field '%s' per API documentation", field)
		}

		assert.IsType(t, "", stream["stream_id"], "stream_id should be string")
		assert.IsType(t, "", stream["camera"], "camera should be string")
		assert.IsType(t, "", stream["url"], "url should be string")
		assert.IsType(t, "", stream["status"], "status should be string")
		assert.IsType(t, float64(0), stream["viewers"], "viewers should be float64")
	}

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "get_streams response should be <200ms per API documentation")
}
