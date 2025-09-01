//go:build integration
// +build integration

/*
WebSocket Integration Tests

Tests real WebSocket connections and JSON-RPC protocol to exercise the complete system stack
including FFmpeg operations, path management, and core API endpoints.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-API-004: get_camera_status method for camera status
- REQ-API-005: take_snapshot method for photo capture
- REQ-API-006: start_recording method for video recording
- REQ-API-007: stop_recording method for video recording
- REQ-API-008: authenticate method for authentication
- REQ-API-009: Role-based access control with viewer, operator, admin permissions
- REQ-API-011: API methods respond within specified time limits
- REQ-API-014: get_streams method for stream enumeration
- REQ-API-015: list_recordings method for recording file management
- REQ-API-016: list_snapshots method for snapshot file management
- REQ-API-017: get_metrics method for system performance metrics

Test Categories: Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	ID      int                    `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      int           `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TestWebSocketIntegration tests real WebSocket connection and all API methods
func TestWebSocketIntegration(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "WebSocket server should start successfully")
	defer env.WebSocketServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8002/ws", nil)
	require.NoError(t, err, "Should connect to WebSocket server")
	defer conn.Close()

	// Generate authentication token
	token, err := env.JWTHandler.GenerateToken("test-user", "admin", 1)
	require.NoError(t, err, "Should generate JWT token")

	t.Run("AuthenticationFlow", func(t *testing.T) {
		// Test authentication
		authRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": token,
			},
			ID: 1,
		}

		response, err := sendWebSocketRequest(conn, authRequest)
		require.NoError(t, err, "Authentication request should succeed")
		assert.NotNil(t, response.Result, "Authentication should return success result")
	})

	t.Run("PingMethod", func(t *testing.T) {
		// Test ping method
		pingRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      2,
		}

		response, err := sendWebSocketRequest(conn, pingRequest)
		require.NoError(t, err, "Ping request should succeed")
		// API doc shows result should be string "pong"
		result, ok := response.Result.(string)
		require.True(t, ok, "Ping result should be string")
		assert.Equal(t, "pong", result, "Ping should return pong")
	})

	t.Run("CameraListMethod", func(t *testing.T) {
		// Test get_camera_list method
		cameraListRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      3,
		}

		response, err := sendWebSocketRequest(conn, cameraListRequest)
		require.NoError(t, err, "Camera list request should succeed")
		assert.NotNil(t, response.Result, "Camera list should return result")
		assert.Contains(t, response.Result, "cameras", "Camera list should contain cameras field")
	})

	t.Run("CameraStatusMethod", func(t *testing.T) {
		// Test get_camera_status method
		cameraStatusRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_status",
			Params: map[string]interface{}{
				"device": "camera0",
			},
			ID: 4,
		}

		response, err := sendWebSocketRequest(conn, cameraStatusRequest)
		// Camera status may fail if device doesn't exist, which is expected
		if err != nil {
			// Expected error for non-existent device
			assert.Contains(t, err.Error(), "Camera not found or disconnected", "Error should be about camera not found")
		} else {
			assert.NotNil(t, response.Result, "Camera status should return result")
		}
	})

	t.Run("TakeSnapshotMethod", func(t *testing.T) {
		// Test take_snapshot method - this will trigger FFmpeg operations
		snapshotRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "take_snapshot",
			Params: map[string]interface{}{
				"device":   "camera0",
				"filename": "test_snapshot.jpg",
				"quality":  85,
			},
			ID: 5,
		}

		response, err := sendWebSocketRequest(conn, snapshotRequest)
		// Note: This may fail if no camera is available, but it will still exercise the code path
		if err == nil {
			assert.NotNil(t, response, "Snapshot request should return response")
		} else {
			// If it fails, it should be due to camera unavailability, not API issues
			assert.Contains(t, err.Error(), "Camera not found or disconnected", "Error should be about camera availability")
		}
	})

	t.Run("StartRecordingMethod", func(t *testing.T) {
		// Test start_recording method - this will trigger path management and FFmpeg
		startRecordingRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "start_recording",
			Params: map[string]interface{}{
				"device":           "camera0",
				"duration_seconds": 30,
				"format":           "mp4",
				"quality":          23,
			},
			ID: 6,
		}

		response, err := sendWebSocketRequest(conn, startRecordingRequest)
		// Note: This may fail if no camera is available, but it will still exercise the code path
		if err == nil {
			assert.NotNil(t, response, "Start recording request should return response")
		} else {
			// If it fails, it should be due to camera unavailability, not API issues
			assert.Contains(t, err.Error(), "Camera not found or disconnected", "Error should be about camera availability")
		}
	})

	t.Run("StopRecordingMethod", func(t *testing.T) {
		// Test stop_recording method
		stopRecordingRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "stop_recording",
			Params: map[string]interface{}{
				"device": "camera0",
			},
			ID: 7,
		}

		response, err := sendWebSocketRequest(conn, stopRecordingRequest)
		// Note: This may fail if no recording is active, but it will still exercise the code path
		if err == nil {
			assert.NotNil(t, response, "Stop recording request should return response")
		} else {
			// API documentation compliance - exact error message
			assert.Contains(t, err.Error(), "Camera is currently recording", "Error must match API documentation")
		}
	})

	t.Run("ListRecordingsMethod", func(t *testing.T) {
		// Test list_recordings method
		listRecordingsRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "list_recordings",
			Params: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
			ID: 8,
		}

		response, err := sendWebSocketRequest(conn, listRecordingsRequest)
		// List recordings may fail if no recordings exist, which is expected
		if err != nil {
			// API documentation compliance - exact error message
			assert.Contains(t, err.Error(), "Insufficient storage space", "Error must match API documentation")
		} else {
			assert.NotNil(t, response.Result, "List recordings should return result")
			assert.Contains(t, response.Result, "recordings", "List recordings should contain recordings field")
		}
	})

	t.Run("ListSnapshotsMethod", func(t *testing.T) {
		// Test list_snapshots method
		listSnapshotsRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "list_snapshots",
			Params: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
			ID: 9,
		}

		response, err := sendWebSocketRequest(conn, listSnapshotsRequest)
		// List snapshots may fail if no snapshots exist, which is expected
		if err != nil {
			// API documentation compliance - exact error message
			assert.Contains(t, err.Error(), "Insufficient storage space", "Error must match API documentation")
		} else {
			assert.NotNil(t, response.Result, "List snapshots should return result")
			assert.Contains(t, response.Result, "snapshots", "List snapshots should contain snapshots field")
		}
	})

	t.Run("GetStreamsMethod", func(t *testing.T) {
		// Test get_streams method
		getStreamsRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_streams",
			ID:      10,
		}

		response, err := sendWebSocketRequest(conn, getStreamsRequest)
		require.NoError(t, err, "Get streams request should succeed")
		// API doc shows result should be array of stream objects
		streams, ok := response.Result.([]interface{})
		require.True(t, ok, "Get streams result should be array")
		assert.NotNil(t, streams, "Get streams should return array")
	})

	t.Run("GetMetricsMethod", func(t *testing.T) {
		// Test get_metrics method
		getMetricsRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_metrics",
			ID:      11,
		}

		response, err := sendWebSocketRequest(conn, getMetricsRequest)
		require.NoError(t, err, "Get metrics request should succeed")
		assert.NotNil(t, response.Result, "Get metrics should return result")
		assert.Contains(t, response.Result, "cpu_usage", "Get metrics should contain CPU usage")
		assert.Contains(t, response.Result, "memory_usage", "Get metrics should contain memory usage")
	})

	t.Run("GetStatusMethod", func(t *testing.T) {
		// Test get_status method
		getStatusRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_status",
			ID:      12,
		}

		response, err := sendWebSocketRequest(conn, getStatusRequest)
		require.NoError(t, err, "Get status request should succeed")
		assert.NotNil(t, response.Result, "Get status should return result")
		assert.Contains(t, response.Result, "status", "Get status should contain status field")
	})

	t.Run("GetServerInfoMethod", func(t *testing.T) {
		// Test get_server_info method
		getServerInfoRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_server_info",
			ID:      13,
		}

		response, err := sendWebSocketRequest(conn, getServerInfoRequest)
		require.NoError(t, err, "Get server info request should succeed")
		assert.NotNil(t, response.Result, "Get server info should return result")
		assert.Contains(t, response.Result, "version", "Get server info should contain version field")
	})

	t.Run("GetStorageInfoMethod", func(t *testing.T) {
		// Test get_storage_info method
		getStorageInfoRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_storage_info",
			ID:      14,
		}

		// API doc shows get_storage_info method exists but server may not implement it yet
		response, err := sendWebSocketRequest(conn, getStorageInfoRequest)
		if err != nil {
			// API documentation compliance - exact error message
			assert.Contains(t, err.Error(), "Insufficient permissions", "Error must match API documentation")
		} else {
			// Method is implemented - validate response format
			storageInfo, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Get storage info result should be object")
			assert.Contains(t, storageInfo, "total_space", "Get storage info should contain total space")
			assert.Contains(t, storageInfo, "available_space", "Get storage info should contain available space")
		}
	})

	t.Run("PerformanceValidation", func(t *testing.T) {
		// Test performance targets per API documentation
		startTime := time.Now()
		pingRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      15,
		}

		_, err = sendWebSocketRequest(conn, pingRequest)
		responseTime := time.Since(startTime)

		require.NoError(t, err, "Ping request should succeed")
		assert.Less(t, responseTime, 50*time.Millisecond, "Ping response should be <50ms per API documentation")
	})

	t.Run("AuthenticationErrorHandling", func(t *testing.T) {
		// Test authentication error handling
		// Create new connection without authentication
		unauthenticatedConn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8002/ws", nil)
		require.NoError(t, err, "Should connect to WebSocket server")
		defer unauthenticatedConn.Close()

		// Try to access protected method without authentication
		protectedRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      16,
		}

		_, err = sendWebSocketRequest(unauthenticatedConn, protectedRequest)
		assert.Error(t, err, "Should fail with authentication error")
		// API documentation compliance - exact error message
		assert.Contains(t, err.Error(), "Authentication failed", "Error must match API documentation")
	})

	t.Run("InvalidParametersErrorHandling", func(t *testing.T) {
		// Test invalid parameters error handling
		invalidRequest := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_status",
			Params: map[string]interface{}{
				"invalid_param": "invalid_value",
			},
			ID: 17,
		}

		_, err = sendWebSocketRequest(conn, invalidRequest)
		assert.Error(t, err, "Should fail with invalid parameters error")
		// API documentation compliance - exact error message
		assert.Contains(t, err.Error(), "Invalid parameters", "Error must match API documentation")
	})
}

// sendWebSocketRequest sends a JSON-RPC request over WebSocket and returns the response
func sendWebSocketRequest(conn *websocket.Conn, request JSONRPCRequest) (*JSONRPCResponse, error) {
	// Send request
	err := conn.WriteJSON(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var response JSONRPCResponse
	err = conn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	return &response, nil
}
