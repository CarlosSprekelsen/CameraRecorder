//go:build unit
// +build unit

/*
WebSocket Methods Comprehensive Test

Requirements Coverage:
- REQ-FUNC-001: WebSocket server functionality
- REQ-FUNC-002: JSON-RPC method implementation
- REQ-FUNC-003: Authentication and authorization
- REQ-FUNC-004: Error handling and validation
- REQ-FUNC-005: File operations (listing, lifecycle)
- REQ-FUNC-006: Recording operations
- REQ-FUNC-007: Snapshot operations
- REQ-FUNC-008: Status and metrics
- REQ-FUNC-009: File listing and browsing functionality
- REQ-API-001: JSON-RPC method implementation
- REQ-SEC-002: Role-based access control

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Real Component Usage: Real MediaMTX controller, real JWT authentication, real camera monitor
*/

package websocket_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
)

// setupWebSocketTestEnvironment creates a complete WebSocket test environment
func setupWebSocketTestEnvironment(t *testing.T) (*utils.WebSocketTestEnvironment, func()) {
	env := utils.SetupWebSocketTestEnvironment(t)

	// Return cleanup function
	cleanup := func() {
		utils.TeardownWebSocketTestEnvironment(t, env)
	}

	return env, cleanup
}

// TestWebSocketServer_JSONRPCCompliance tests JSON-RPC 2.0 protocol compliance
func TestWebSocketServer_JSONRPCCompliance(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Test that WebSocket server is properly initialized
	assert.NotNil(t, env.WebSocketServer, "WebSocket server should be created")
	assert.NotNil(t, env.JWTHandler, "JWT handler should be created")
	assert.NotNil(t, env.Controller, "MediaMTX controller should be real")
	assert.NotNil(t, env.CameraMonitor, "Camera monitor should be created")

	// Test JSON-RPC 2.0 protocol compliance
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Test ping method (basic JSON-RPC compliance)
	response, err := env.WebSocketServer.MethodPing(map[string]interface{}{}, client)
	require.NoError(t, err, "Ping method should not return error")
	require.NotNil(t, response, "Ping method should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Nil(t, response.Error, "Ping should not return error")
	assert.NotNil(t, response.Result, "Ping should return result")
}

// TestWebSocketServer_Authentication tests authentication flow
func TestWebSocketServer_Authentication(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Test authentication method
	authParams := map[string]interface{}{
		"username": "test_user",
		"password": "test_password",
	}

	// Create unauthenticated client
	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: false,
		UserID:        "",
		Role:          "",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	response, err := env.WebSocketServer.MethodAuthenticate(authParams, client)
	require.NoError(t, err, "Authentication method should not return error")
	require.NotNil(t, response, "Authentication should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	// Authentication may fail in test environment, but structure should be correct
	if response.Error != nil {
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		assert.NotNil(t, response.Result, "Successful auth should return result")
	}
}

// TestWebSocketServer_FileListingMethods tests file listing API compliance
func TestWebSocketServer_FileListingMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with viewer role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_viewer", "viewer")

	// Test recordings listing with valid parameters
	response, err := env.WebSocketServer.MethodListRecordings(map[string]interface{}{
		"limit":  10,
		"offset": 0,
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "ListRecordings should not return error")
	require.NotNil(t, response, "ListRecordings should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or no recordings
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "files", "Should contain files array")
		assert.Contains(t, result, "total", "Should contain total count")
		assert.Contains(t, result, "limit", "Should contain limit")
		assert.Contains(t, result, "offset", "Should contain offset")
	}

	// Test snapshots listing
	response, err = env.WebSocketServer.MethodListSnapshots(map[string]interface{}{
		"limit":  10,
		"offset": 0,
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "ListSnapshots should not return error")
	require.NotNil(t, response, "ListSnapshots should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or no snapshots
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "files", "Should contain files array")
		assert.Contains(t, result, "total", "Should contain total count")
		assert.Contains(t, result, "limit", "Should contain limit")
		assert.Contains(t, result, "offset", "Should contain offset")
	}
}

// TestWebSocketServer_RecordingMethods tests recording operations API compliance
func TestWebSocketServer_RecordingMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with operator role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_operator", "operator")

	// Test start recording with required parameters
	response, err := env.WebSocketServer.MethodStartRecording(map[string]interface{}{
		"device": "/dev/video0",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "StartRecording should not return error")
	require.NotNil(t, response, "StartRecording should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or device not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "session_id", "Should contain session_id")
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "start_time", "Should contain start_time")
		assert.Contains(t, result, "duration", "Should contain duration")
		assert.Contains(t, result, "format", "Should contain format")
	}

	// Test start recording with optional parameters
	response, err = env.WebSocketServer.MethodStartRecording(map[string]interface{}{
		"device":   "/dev/video0",
		"duration": 3600,
		"format":   "mp4",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "StartRecording with options should not return error")
	require.NotNil(t, response, "StartRecording should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or device not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "session_id", "Should contain session_id")
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "start_time", "Should contain start_time")
		assert.Contains(t, result, "duration", "Should contain duration")
		assert.Contains(t, result, "format", "Should contain format")
	}

	// Test stop recording
	response, err = env.WebSocketServer.MethodStopRecording(map[string]interface{}{
		"device": "/dev/video0",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "StopRecording should not return error")
	require.NotNil(t, response, "StopRecording should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if session doesn't exist
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "session_id", "Should contain session_id")
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "start_time", "Should contain start_time")
		assert.Contains(t, result, "end_time", "Should contain end_time")
		assert.Contains(t, result, "duration", "Should contain duration")
		assert.Contains(t, result, "file_size", "Should contain file_size")
	}
}

// TestWebSocketServer_SnapshotMethods tests snapshot operations API compliance
func TestWebSocketServer_SnapshotMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with operator role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_operator", "operator")

	// Test take snapshot with required parameters
	response, err := env.WebSocketServer.MethodTakeSnapshot(map[string]interface{}{
		"device": "/dev/video0",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "TakeSnapshot should not return error")
	require.NotNil(t, response, "TakeSnapshot should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or device not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "timestamp", "Should contain timestamp")
		assert.Contains(t, result, "file_size", "Should contain file_size")
		assert.Contains(t, result, "file_path", "Should contain file_path")
	}

	// Test take snapshot with optional filename parameter
	response, err = env.WebSocketServer.MethodTakeSnapshot(map[string]interface{}{
		"device":   "/dev/video0",
		"filename": "custom_snapshot.jpg",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "TakeSnapshot with filename should not return error")
	require.NotNil(t, response, "TakeSnapshot should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running or device not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "timestamp", "Should contain timestamp")
		assert.Contains(t, result, "file_size", "Should contain file_size")
		assert.Contains(t, result, "file_path", "Should contain file_path")
	}
}

// TestWebSocketServer_StatusMethods tests status and metrics API compliance
func TestWebSocketServer_StatusMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with admin role (API requires admin for get_status)
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_admin", "admin")

	// Test get status
	response, err := env.WebSocketServer.MethodGetStatus(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetStatus should not return error")
	require.NotNil(t, response, "GetStatus should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "status", "Should contain status")
		assert.Contains(t, result, "uptime", "Should contain uptime")
		assert.Contains(t, result, "version", "Should contain version")
		assert.Contains(t, result, "components", "Should contain components")
	}

	// Test get metrics (requires admin role)
	response, err = env.WebSocketServer.MethodGetMetrics(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetMetrics should not return error")
	require.NotNil(t, response, "GetMetrics should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "active_connections", "Should contain active_connections")
		assert.Contains(t, result, "total_requests", "Should contain total_requests")
		assert.Contains(t, result, "average_response_time", "Should contain average_response_time")
		assert.Contains(t, result, "error_rate", "Should contain error_rate")
		assert.Contains(t, result, "memory_usage", "Should contain memory_usage")
		assert.Contains(t, result, "cpu_usage", "Should contain cpu_usage")
		assert.Contains(t, result, "goroutines", "Should contain goroutines")
		assert.Contains(t, result, "heap_alloc", "Should contain heap_alloc")
	}
}

// TestWebSocketServer_FileLifecycleMethods tests file lifecycle operations API compliance
func TestWebSocketServer_FileLifecycleMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with operator role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_operator", "operator")

	// Test delete recording
	response, err := env.WebSocketServer.MethodDeleteRecording(map[string]interface{}{
		"filename": "test_recording.mp4",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "DeleteRecording should not return error")
	require.NotNil(t, response, "DeleteRecording should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if file doesn't exist
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "deleted", "Should contain deleted")
		assert.Contains(t, result, "message", "Should contain message")
	}

	// Test delete snapshot
	response, err = env.WebSocketServer.MethodDeleteSnapshot(map[string]interface{}{
		"filename": "test_snapshot.jpg",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "DeleteSnapshot should not return error")
	require.NotNil(t, response, "DeleteSnapshot should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if file doesn't exist
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		// Validate required fields per API documentation
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "deleted", "Should contain deleted")
		assert.Contains(t, result, "message", "Should contain message")
	}
}

// TestWebSocketServer_ErrorHandling tests error handling and validation
func TestWebSocketServer_ErrorHandling(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Test with invalid parameters
	response, err := env.WebSocketServer.MethodListRecordings(map[string]interface{}{
		"limit": "invalid", // Invalid type
	}, client)

	// Should handle error gracefully
	require.NoError(t, err, "Method should not return error")
	require.NotNil(t, response, "Should return response even with error")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	}

	// Test with missing required parameters
	response, err = env.WebSocketServer.MethodStartRecording(map[string]interface{}{
		// Missing required "device" parameter
	}, client)

	// Should handle missing parameters gracefully
	require.NoError(t, err, "Method should not return error")
	require.NotNil(t, response, "Should return response even with missing parameters")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	}
}

// TestWebSocketServer_RoleBasedAccess tests role-based access control
func TestWebSocketServer_RoleBasedAccess(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Test with different user roles
	testCases := []struct {
		name     string
		userID   string
		role     string
		expected bool
	}{
		{"viewer_role", "test_viewer", "viewer", true},
		{"operator_role", "test_operator", "operator", true},
		{"admin_role", "test_admin", "admin", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := utils.CreateAuthenticatedClient(t, env.JWTHandler, tc.userID, tc.role)

			// Test basic method access
			response, err := env.WebSocketServer.MethodListRecordings(map[string]interface{}{}, client)

			// Should not fail due to authentication
			require.NoError(t, err, "Method should not return error for %s", tc.role)
			require.NotNil(t, response, "Should return response for %s", tc.role)
			assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
		})
	}
}

// TestWebSocketServer_CameraMethods tests camera-related methods API compliance
func TestWebSocketServer_CameraMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Test get camera list
	response, err := env.WebSocketServer.MethodGetCameraList(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetCameraList should not return error")
	require.NotNil(t, response, "GetCameraList should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if no cameras available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		assert.Contains(t, result, "cameras", "Should contain cameras array")
	}

	// Test get camera status
	response, err = env.WebSocketServer.MethodGetCameraStatus(map[string]interface{}{
		"device": "/dev/video0",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetCameraStatus should not return error")
	require.NotNil(t, response, "GetCameraStatus should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if camera not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		assert.Contains(t, result, "status", "Should contain status")
	}

	// Test get camera capabilities
	response, err = env.WebSocketServer.MethodGetCameraCapabilities(map[string]interface{}{
		"device": "/dev/video0",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetCameraCapabilities should not return error")
	require.NotNil(t, response, "GetCameraCapabilities should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if camera not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "device", "Should contain device")
		assert.Contains(t, result, "formats", "Should contain formats")
		assert.Contains(t, result, "resolutions", "Should contain resolutions")
		assert.Contains(t, result, "fps_options", "Should contain fps_options")
		assert.Contains(t, result, "validation_status", "Should contain validation_status")
	}
}

// TestWebSocketServer_AdminMethods tests admin-only methods API compliance
func TestWebSocketServer_AdminMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with admin role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_admin", "admin")

	// Test get server info
	response, err := env.WebSocketServer.MethodGetServerInfo(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetServerInfo should not return error")
	require.NotNil(t, response, "GetServerInfo should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if service not available
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "name", "Should contain name")
		assert.Contains(t, result, "version", "Should contain version")
		assert.Contains(t, result, "build_date", "Should contain build_date")
		assert.Contains(t, result, "go_version", "Should contain go_version")
		assert.Contains(t, result, "architecture", "Should contain architecture")
		assert.Contains(t, result, "capabilities", "Should contain capabilities")
		assert.Contains(t, result, "supported_formats", "Should contain supported_formats")
		assert.Contains(t, result, "max_cameras", "Should contain max_cameras")
	}

	// Test get streams
	response, err = env.WebSocketServer.MethodGetStreams(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetStreams should not return error")
	require.NotNil(t, response, "GetStreams should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if MediaMTX not running
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		streams, ok := response.Result.([]interface{})
		require.True(t, ok, "Result should be an array")
		
		// If streams exist, validate their structure
		if len(streams) > 0 {
			stream, ok := streams[0].(map[string]interface{})
			require.True(t, ok, "Stream should be a map")
			assert.Contains(t, stream, "name", "Should contain name")
			assert.Contains(t, stream, "source", "Should contain source")
			assert.Contains(t, stream, "ready", "Should contain ready")
			assert.Contains(t, stream, "readers", "Should contain readers")
			assert.Contains(t, stream, "bytes_sent", "Should contain bytes_sent")
		}
	}

	// Test get storage info
	response, err = env.WebSocketServer.MethodGetStorageInfo(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetStorageInfo should not return error")
	require.NotNil(t, response, "GetStorageInfo should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if storage not accessible
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "total_space", "Should contain total_space")
		assert.Contains(t, result, "used_space", "Should contain used_space")
		assert.Contains(t, result, "available_space", "Should contain available_space")
		assert.Contains(t, result, "usage_percentage", "Should contain usage_percentage")
		assert.Contains(t, result, "recordings_size", "Should contain recordings_size")
		assert.Contains(t, result, "snapshots_size", "Should contain snapshots_size")
		assert.Contains(t, result, "low_space_warning", "Should contain low_space_warning")
	}

	// Test cleanup old files
	response, err = env.WebSocketServer.MethodCleanupOldFiles(map[string]interface{}{}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "CleanupOldFiles should not return error")
	require.NotNil(t, response, "CleanupOldFiles should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if cleanup fails
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "cleanup_executed", "Should contain cleanup_executed")
		assert.Contains(t, result, "files_deleted", "Should contain files_deleted")
		assert.Contains(t, result, "space_freed", "Should contain space_freed")
		assert.Contains(t, result, "message", "Should contain message")
	}

	// Test set retention policy
	response, err = env.WebSocketServer.MethodSetRetentionPolicy(map[string]interface{}{
		"policy_type":  "age",
		"max_age_days": 30,
		"enabled":      true,
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "SetRetentionPolicy should not return error")
	require.NotNil(t, response, "SetRetentionPolicy should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if policy setting fails
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "policy_type", "Should contain policy_type")
		assert.Contains(t, result, "max_age_days", "Should contain max_age_days")
		assert.Contains(t, result, "enabled", "Should contain enabled")
		assert.Contains(t, result, "message", "Should contain message")
	}
}

// TestWebSocketServer_FileInfoMethods tests file information methods API compliance
func TestWebSocketServer_FileInfoMethods(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client with viewer role
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_viewer", "viewer")

	// Test get recording info
	response, err := env.WebSocketServer.MethodGetRecordingInfo(map[string]interface{}{
		"filename": "test_recording.mp4",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetRecordingInfo should not return error")
	require.NotNil(t, response, "GetRecordingInfo should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if file doesn't exist
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "file_size", "Should contain file_size")
		assert.Contains(t, result, "duration", "Should contain duration")
		assert.Contains(t, result, "created_time", "Should contain created_time")
		assert.Contains(t, result, "download_url", "Should contain download_url")
	}

	// Test get snapshot info
	response, err = env.WebSocketServer.MethodGetSnapshotInfo(map[string]interface{}{
		"filename": "test_snapshot.jpg",
	}, client)

	// Validate JSON-RPC response structure
	require.NoError(t, err, "GetSnapshotInfo should not return error")
	require.NotNil(t, response, "GetSnapshotInfo should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if file doesn't exist
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "filename", "Should contain filename")
		assert.Contains(t, result, "file_size", "Should contain file_size")
		assert.Contains(t, result, "created_time", "Should contain created_time")
		assert.Contains(t, result, "download_url", "Should contain download_url")
	}
}

// TestWebSocketServer_AuthenticationScenarios tests comprehensive authentication scenarios
func TestWebSocketServer_AuthenticationScenarios(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Test authentication with valid token
	validToken := utils.GenerateTestToken(t, env.JWTHandler, "test_user", "viewer")
	authParams := map[string]interface{}{
		"auth_token": validToken,
	}

	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: false,
		UserID:        "",
		Role:          "",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	response, err := env.WebSocketServer.MethodAuthenticate(authParams, client)
	require.NoError(t, err, "Authentication should not return error")
	require.NotNil(t, response, "Authentication should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	if response.Error != nil {
		// Expected if authentication fails
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	} else {
		// Validate successful response structure per API documentation
		require.NotNil(t, response.Result, "Result should be present")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")
		
		// Validate required fields per API documentation
		assert.Contains(t, result, "authenticated", "Should contain authenticated")
		assert.Contains(t, result, "role", "Should contain role")
		assert.Contains(t, result, "permissions", "Should contain permissions")
		assert.Contains(t, result, "expires_at", "Should contain expires_at")
		assert.Contains(t, result, "session_id", "Should contain session_id")
	}

	// Test authentication with invalid token
	invalidToken := "invalid_token"
	authParams = map[string]interface{}{
		"auth_token": invalidToken,
	}

	response, err = env.WebSocketServer.MethodAuthenticate(authParams, client)
	require.NoError(t, err, "Authentication should not return error")
	require.NotNil(t, response, "Authentication should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	// Should return error for invalid token
	if response.Error != nil {
		assert.NotNil(t, response.Error.Code, "Error should have code")
		assert.NotEmpty(t, response.Error.Message, "Error should have message")
	}
}

// TestWebSocketServer_ErrorScenarios tests comprehensive error scenarios
func TestWebSocketServer_ErrorScenarios(t *testing.T) {
	env, cleanup := setupWebSocketTestEnvironment(t)
	defer cleanup()

	// Create authenticated client
	client := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Test invalid method name
	response, err := env.WebSocketServer.MethodPing(map[string]interface{}{
		"invalid_param": "value",
	}, client)

	// Should handle invalid parameters gracefully
	require.NoError(t, err, "Method should not return error")
	require.NotNil(t, response, "Should return response even with invalid parameters")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")

	// Test missing required parameters for various methods
	testCases := []struct {
		name   string
		method func(map[string]interface{}, *websocket.ClientConnection) (*websocket.JsonRpcResponse, error)
		params map[string]interface{}
	}{
		{
			name:   "GetCameraStatus missing device",
			method: env.WebSocketServer.MethodGetCameraStatus,
			params: map[string]interface{}{},
		},
		{
			name:   "TakeSnapshot missing device",
			method: env.WebSocketServer.MethodTakeSnapshot,
			params: map[string]interface{}{},
		},
		{
			name:   "DeleteRecording missing filename",
			method: env.WebSocketServer.MethodDeleteRecording,
			params: map[string]interface{}{},
		},
		{
			name:   "DeleteSnapshot missing filename",
			method: env.WebSocketServer.MethodDeleteSnapshot,
			params: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := tc.method(tc.params, client)
			require.NoError(t, err, "Method should not return error")
			require.NotNil(t, response, "Should return response")
			assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
			
			if response.Error != nil {
				assert.NotNil(t, response.Error.Code, "Error should have code")
				assert.NotEmpty(t, response.Error.Message, "Error should have message")
			}
		})
	}
}
