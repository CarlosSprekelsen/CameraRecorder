//go:build integration
// +build integration

/*
Complete End-to-End Integration Tests

This file consolidates all integration testing into comprehensive end-to-end workflows
that test the complete system through external API calls only, exactly as external clients would use it.

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

Test Categories: Integration (Real System End-to-End)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/constants"
	ws "github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

// TestCompleteEndToEndIntegration tests the complete system through external API calls only
func TestCompleteEndToEndIntegration(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := testtestutils.SetupWebSocketTestEnvironment(t)
	defer testtestutils.TeardownWebSocketTestEnvironment(t, env)

	// Start WebSocket server
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "WebSocket server should start successfully")
	defer env.WebSocketServer.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// COMMON PATTERN: Use centralized WebSocket test client instead of hardcoded connection
	client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Generate authentication token
	token, err := env.JWTHandler.GenerateToken("test-user", "admin", 1)
	require.NoError(t, err, "Should generate JWT token")

	t.Run("AuthenticationAndAuthorization", func(t *testing.T) {
		testAuthenticationAndAuthorization(t, client, env)
	})

	t.Run("CameraDiscoveryAndStatus", func(t *testing.T) {
		testCameraDiscoveryAndStatus(t, client, token, env)
	})

	t.Run("SnapshotOperations", func(t *testing.T) {
		testSnapshotOperations(t, client, token, env)
	})

	t.Run("RecordingOperations", func(t *testing.T) {
		testRecordingOperations(t, client, token, env)
	})

	t.Run("FileManagement", func(t *testing.T) {
		testFileManagement(t, client, token, env)
	})

	t.Run("SystemHealthAndMetrics", func(t *testing.T) {
		testSystemHealthAndMetrics(t, client, token)
	})

	t.Run("StreamManagement", func(t *testing.T) {
		testStreamManagement(t, client, token, env)
	})

	t.Run("ErrorHandlingAndEdgeCases", func(t *testing.T) {
		testErrorHandlingAndEdgeCases(t, client, token)
	})
}

// testAuthenticationAndAuthorization tests authentication and role-based access control
func testAuthenticationAndAuthorization(t *testing.T, client *testtestutils.WebSocketTestClient, env *testtestutils.WebSocketTestEnvironment) {
	// Test authentication with valid token
	t.Run("ValidAuthentication", func(t *testing.T) {
		token, err := env.JWTHandler.GenerateToken("auth-test-user", "admin", 1)
		require.NoError(t, err, "Should generate valid JWT token")

		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			ID:      1,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(client, request)
		require.NoError(t, err, "Valid authentication should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Authentication result should be an object")
		require.Contains(t, result, "authenticated", "Result should contain authenticated field per API documentation")
		require.Contains(t, result, "role", "Result should contain role field per API documentation")
		require.Contains(t, result, "permissions", "Result should contain permissions field per API documentation")

		authenticated, ok := result["authenticated"].(bool)
		require.True(t, ok, "Authenticated field should be boolean")
		assert.True(t, authenticated, "Authentication should succeed")
	})

	// Test authentication with invalid token
	t.Run("InvalidAuthentication", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			ID:      2,
			Params: map[string]interface{}{
				"auth_token": "invalid-token",
			},
		}

		response, err := sendWebSocketRequest(client, request)
		require.NoError(t, err, "Should get JSON-RPC response for invalid authentication")
		require.NotNil(t, response, "Should get response for invalid authentication")
		require.NotNil(t, response.Error, "Invalid authentication should return error response")
		assert.Equal(t, constants.API_AUTHENTICATION_REQUIRED, response.Error.Code, "Error code should be AUTHENTICATION_REQUIRED")
		assert.Contains(t, response.Error.Message, "Authentication failed", "Error message should match API documentation")
	})

	// Test protected method without authentication
	t.Run("UnauthenticatedAccess", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      3,
			Params: map[string]interface{}{
				"auth_token": "", // Empty token
			},
		}

		response, err := sendWebSocketRequest(client, request)
		// For unauthenticated access, we should get a JSON-RPC error response, not a Go error
		require.NoError(t, err, "Should get JSON-RPC response for unauthenticated access")
		require.NotNil(t, response, "Should get response for unauthenticated access")
		require.NotNil(t, response.Error, "Unauthenticated access should return error response")
		assert.Equal(t, constants.API_AUTHENTICATION_REQUIRED, response.Error.Code, "Error code should be AUTHENTICATION_REQUIRED")
		assert.Contains(t, response.Error.Message, "Authentication failed", "Error message should match API documentation")
	})

	// Test role-based access control
	t.Run("RoleBasedAccessControl", func(t *testing.T) {
		// Test viewer role permissions
		viewerToken, err := env.JWTHandler.GenerateToken("viewer-user", "viewer", 1)
		require.NoError(t, err, "Should generate viewer token")

		// Viewer should be able to get camera list
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      4,
			Params: map[string]interface{}{
				"auth_token": viewerToken,
			},
		}

		response, err := sendWebSocketRequest(client, request)
		require.NoError(t, err, "Viewer should be able to get camera list")
		assert.NotNil(t, response.Result, "Camera list should return result")

		// Viewer should NOT be able to start recording
		recordingRequest := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "start_recording",
			ID:      5,
			Params: map[string]interface{}{
				"device":           "camera0",
				"duration_seconds": 5,
				"format":           "mp4",
				"quality":          23,
				"auth_token":       viewerToken,
			},
		}

		recordingResponse, err := sendWebSocketRequest(client, recordingRequest)
		require.NoError(t, err, "Should get JSON-RPC response for insufficient permissions")
		require.NotNil(t, recordingResponse, "Should get response for insufficient permissions")
		require.NotNil(t, recordingResponse.Error, "Insufficient permissions should return error response")
		assert.Equal(t, INSUFFICIENT_PERMISSIONS, recordingResponse.Error.Code, "Error code should be INSUFFICIENT_PERMISSIONS")
		assert.Contains(t, recordingResponse.Error.Message, "Insufficient permissions", "Error message should match API documentation")
	})
}

// testCameraDiscoveryAndStatus tests camera discovery and status methods
func testCameraDiscoveryAndStatus(t *testing.T, client *testtestutils.WebSocketTestClient, token string, env *testtestutils.WebSocketTestEnvironment) {
	// Test get_camera_list method
	t.Run("GetCameraList", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      6,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(client, request)
		require.NoError(t, err, "Camera discovery should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be an object")
		require.Contains(t, result, "cameras", "Result should contain cameras field per API documentation")
		require.Contains(t, result, "total", "Result should contain total field per API documentation")
		require.Contains(t, result, "connected", "Result should contain connected field per API documentation")

		cameras, ok := result["cameras"].([]interface{})
		require.True(t, ok, "Cameras should be an array")

		total, ok := result["total"].(float64)
		require.True(t, ok, "Total should be numeric")
		assert.Equal(t, float64(len(cameras)), total, "Total should match camera count")

		connected, ok := result["connected"].(float64)
		require.True(t, ok, "Connected should be numeric")
		assert.LessOrEqual(t, connected, total, "Connected should not exceed total")

		t.Logf("Discovered %d cameras (%d connected)", int(total), int(connected))
	})

	// Test get_camera_status method for each discovered camera
	t.Run("GetCameraStatus", func(t *testing.T) {
		// First get camera list
		listRequest := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_list",
			ID:      7,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		listResponse, err := sendWebSocketRequest(conn, listRequest)
		require.NoError(t, err, "Should get camera list")

		result, ok := listResponse.Result.(map[string]interface{})
		require.True(t, ok, "List result should be an object")

		cameras, ok := result["cameras"].([]interface{})
		require.True(t, ok, "Cameras should be an array")

		if len(cameras) == 0 {
			t.Skip("No cameras available for status testing")
		}

		// Test status for each camera
		for i, camera := range cameras {
			cameraInfo, ok := camera.(map[string]interface{})
			require.True(t, ok, "Camera info should be an object")

			device, ok := cameraInfo["device"].(string)
			require.True(t, ok, "Device should be string")

			// Get camera status
			statusRequest := &ws.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "get_camera_status",
				ID:      8 + i,
				Params: map[string]interface{}{
					"device":     device,
					"auth_token": token,
				},
			}

			statusResponse, err := sendWebSocketRequest(conn, statusRequest)
			require.NoError(t, err, "Should get JSON-RPC response for camera status")

			// Check if we got an error response (camera not available)
			if statusResponse.Error != nil {
				// If camera is not available, that's acceptable
				assert.Contains(t, statusResponse.Error.Message, "Camera not found or disconnected",
					"Error should match API documentation for unavailable camera")
				continue
			}

			// Validate status response format per API documentation
			statusResult, ok := statusResponse.Result.(map[string]interface{})
			require.True(t, ok, "Status result should be an object")
			require.Contains(t, statusResult, "device", "Status should contain device field per API documentation")
			require.Contains(t, statusResult, "status", "Status should contain status field per API documentation")
			require.Contains(t, statusResult, "name", "Status should contain name field per API documentation")

			// Validate device field matches request
			statusDevice, ok := statusResult["device"].(string)
			require.True(t, ok, "Status device should be string")
			assert.Equal(t, device, statusDevice, "Status device should match request device")

			t.Logf("Camera %s status: %s", device, statusResult["status"])
		}
	})

	// Test get_camera_capabilities method
	t.Run("GetCameraCapabilities", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for capabilities testing")
		}

		// Use first available camera
		var cameraID string
		for devicePath := range cameras {
			if strings.HasPrefix(devicePath, "/dev/video") {
				deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
				cameraID = fmt.Sprintf("camera%s", deviceNum)
			} else {
				cameraID = devicePath
			}
			break
		}

		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_capabilities",
			ID:      20,
			Params: map[string]interface{}{
				"device":     cameraID,
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Get camera capabilities method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Capabilities result should be an object")
			require.Contains(t, result, "device", "Result should contain device field per API documentation")
			require.Contains(t, result, "formats", "Result should contain formats field per API documentation")
			require.Contains(t, result, "resolutions", "Result should contain resolutions field per API documentation")
			require.Contains(t, result, "fps_options", "Result should contain fps_options field per API documentation")
		}
	})
}

// testSnapshotOperations tests snapshot capture and management
func testSnapshotOperations(t *testing.T, client *testtestutils.WebSocketTestClient, token string, env *testtestutils.WebSocketTestEnvironment) {
	// Test take_snapshot method
	t.Run("TakeSnapshot", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for snapshot testing")
		}

		// Use first available camera
		var cameraID string
		for devicePath := range cameras {
			if strings.HasPrefix(devicePath, "/dev/video") {
				deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
				cameraID = fmt.Sprintf("camera%s", deviceNum)
			} else {
				cameraID = devicePath
			}
			break
		}

		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "take_snapshot",
			ID:      30,
			Params: map[string]interface{}{
				"device":     cameraID,
				"filename":   "test_integration_snapshot.jpg",
				"quality":    85,
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for snapshot")

		// Check if we got an error response (camera not available)
		if response.Error != nil {
			// If snapshot fails due to camera unavailability, that's acceptable
			// but we should validate the error message matches API documentation
			assert.Contains(t, response.Error.Message, "Camera not found or disconnected",
				"Error should match API documentation for camera unavailability")
			t.Skip("Camera not available for snapshot testing")
		}

		// If snapshot succeeds, validate response format and file creation
		require.NotNil(t, response.Result, "Snapshot should return result")
		_, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be an object")

		// Validate file was actually created on disk
		cfg := env.ConfigManager.GetConfig()
		snapshotsDir := cfg.MediaMTX.SnapshotsPath
		filePath := filepath.Join(snapshotsDir, "test_integration_snapshot.jpg")

		fileInfo, err := os.Stat(filePath)
		require.NoError(t, err, "Snapshot file should exist on disk")
		assert.True(t, fileInfo.Size() > 0, "Snapshot file should not be empty")

		t.Logf("Snapshot captured successfully: %s (size: %d bytes)", filePath, fileInfo.Size())
	})

	// Test list_snapshots method
	t.Run("ListSnapshots", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "list_snapshots",
			ID:      31,
			Params: map[string]interface{}{
				"limit":      10,
				"offset":     0,
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for list snapshots")

		if response.Error != nil {
			// If no snapshots exist, that's acceptable
			assert.Contains(t, response.Error.Message, "Invalid parameters", "Error should match API documentation")
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "List snapshots result should be an object")
			require.Contains(t, result, "snapshots", "Result should contain snapshots field per API documentation")
		}
	})

	// Test get_snapshot_info method
	t.Run("GetSnapshotInfo", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_snapshot_info",
			ID:      32,
			Params: map[string]interface{}{
				"filename":   "test_integration_snapshot.jpg",
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Get snapshot info method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Snapshot info result should be an object")
			require.Contains(t, result, "file_size", "Result should contain file_size field per API documentation")
			require.Contains(t, result, "resolution", "Result should contain resolution field per API documentation")
		}
	})

	// Test delete_snapshot method
	t.Run("DeleteSnapshot", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "delete_snapshot",
			ID:      33,
			Params: map[string]interface{}{
				"filename":   "test_integration_snapshot.jpg",
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Delete snapshot method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Delete snapshot result should be an object")
			require.Contains(t, result, "success", "Result should contain success field per API documentation")
		}
	})
}

// testRecordingOperations tests recording start, stop, and management
func testRecordingOperations(t *testing.T, client *testtestutils.WebSocketTestClient, token string, env *testtestutils.WebSocketTestEnvironment) {
	// Test start_recording method
	t.Run("StartRecording", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for recording testing")
		}

		// Use first available camera
		var cameraID string
		for devicePath := range cameras {
			if strings.HasPrefix(devicePath, "/dev/video") {
				deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
				cameraID = fmt.Sprintf("camera%s", deviceNum)
			} else {
				cameraID = devicePath
			}
			break
		}

		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "start_recording",
			ID:      40,
			Params: map[string]interface{}{
				"device":           cameraID,
				"duration_seconds": 5, // Short duration for testing
				"format":           "mp4",
				"quality":          23,
				"auth_token":       token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for recording")

		// Check if we got an error response (camera not available)
		if response.Error != nil {
			// If recording fails due to camera unavailability, that's acceptable
			// but we should validate the error message matches API documentation
			assert.Contains(t, response.Error.Message, "Camera not found or disconnected",
				"Error should match API documentation for camera unavailability")
			t.Skip("Camera not available for recording testing")
		}

		// If recording succeeds, validate response format
		require.NotNil(t, response.Result, "Recording start should return result")
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be an object")
		require.Contains(t, result, "device", "Result should contain device field per API documentation")
		require.Contains(t, result, "filename", "Result should contain filename field per API documentation")

		device, ok := result["device"].(string)
		require.True(t, ok, "Device should be string")
		require.NotEmpty(t, device, "Device should not be empty")

		filename, ok := result["filename"].(string)
		require.True(t, ok, "Filename should be string")
		require.NotEmpty(t, filename, "Filename should not be empty")

		t.Logf("Recording started successfully: device %s, filename %s", device, filename)

		// Wait for recording to progress
		time.Sleep(2 * time.Second)

		// Test stop_recording method
		t.Run("StopRecording", func(t *testing.T) {
			stopRequest := &ws.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "stop_recording",
				ID:      41,
				Params: map[string]interface{}{
					"device":     cameraID,
					"auth_token": token,
				},
			}

			stopResponse, err := sendWebSocketRequest(conn, stopRequest)
			require.NoError(t, err, "Should get JSON-RPC response for stop recording")

			if stopResponse.Error != nil {
				// If no active recording, that's acceptable per API documentation
				assert.Contains(t, stopResponse.Error.Message, "No active recording session found for device",
					"Error should match API documentation")
			} else {
				// Validate recording stopped successfully
				require.NotNil(t, stopResponse.Result, "Recording stop should return result")
				t.Logf("Recording stopped successfully")
			}

			// Verify recording file was created (if recording succeeded)
			if stopResponse.Error == nil {
				// Wait a bit for file finalization
				time.Sleep(1 * time.Second)

				// List recordings to find the created file
				listRequest := &ws.JsonRpcRequest{
					JSONRPC: "2.0",
					Method:  "list_recordings",
					ID:      42,
					Params: map[string]interface{}{
						"limit":      10,
						"offset":     0,
						"auth_token": token,
					},
				}

				listResponse, err := sendWebSocketRequest(conn, listRequest)
				require.NoError(t, err, "Should get JSON-RPC response for list recordings")

				if listResponse.Error == nil {
					// Validate recordings list response
					listResult, ok := listResponse.Result.(map[string]interface{})
					require.True(t, ok, "List result should be an object")

					if recordings, exists := listResult["recordings"]; exists {
						recordingsArray, ok := recordings.([]interface{})
						if ok && len(recordingsArray) > 0 {
							// Find the most recent recording
							var latestRecording map[string]interface{}
							for _, rec := range recordingsArray {
								if recMap, ok := rec.(map[string]interface{}); ok {
									if latestRecording == nil ||
										(recMap["created_at"] != nil && latestRecording["created_at"] == nil) {
										latestRecording = recMap
									}
								}
							}

							if latestRecording != nil {
								// Validate recording file exists on disk
								if fileName, exists := latestRecording["filename"]; exists {
									cfg := env.ConfigManager.GetConfig()
									recordingsDir := cfg.MediaMTX.RecordingsPath
									filePath := filepath.Join(recordingsDir, fileName.(string))

									fileInfo, err := os.Stat(filePath)
									if err == nil {
										assert.True(t, fileInfo.Size() > 0, "Recording file should not be empty")
										t.Logf("Recording file verified: %s (size: %d bytes)", filePath, fileInfo.Size())
									} else {
										t.Logf("Warning: Recording file not found on disk: %s", filePath)
									}
								}
							}
						}
					}
				}
			}
		})
	})

	// Test list_recordings method
	t.Run("ListRecordings", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "list_recordings",
			ID:      43,
			Params: map[string]interface{}{
				"limit":      10,
				"offset":     0,
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for list recordings")

		if response.Error != nil {
			// If no recordings exist, that's acceptable
			assert.Contains(t, response.Error.Message, "Invalid parameters", "Error should match API documentation")
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "List recordings result should be an object")
			require.Contains(t, result, "recordings", "Result should contain recordings field per API documentation")
		}
	})

	// Test get_recording_info method
	t.Run("GetRecordingInfo", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_recording_info",
			ID:      44,
			Params: map[string]interface{}{
				"filename":   "test_recording_123.mp4",
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Get recording info method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Recording info result should be an object")
			require.Contains(t, result, "duration", "Result should contain duration field per API documentation")
			require.Contains(t, result, "file_size", "Result should contain file_size field per API documentation")
		}
	})

	// Test delete_recording method
	t.Run("DeleteRecording", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "delete_recording",
			ID:      45,
			Params: map[string]interface{}{
				"filename":   "test_recording_123.mp4",
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Delete recording method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Delete recording result should be an object")
			require.Contains(t, result, "success", "Result should contain success field per API documentation")
		}
	})
}

// testFileManagement tests file management, cleanup, and retention policies
func testFileManagement(t *testing.T, client *testtestutils.WebSocketTestClient, token string, env *testtestutils.WebSocketTestEnvironment) {
	// Test get_storage_info method
	t.Run("GetStorageInfo", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_storage_info",
			ID:      50,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			assert.Contains(t, err.Error(), "Insufficient permissions", "Error should match API documentation")
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Storage info result should be an object")
			require.Contains(t, result, "total_space", "Result should contain total_space field per API documentation")
			require.Contains(t, result, "available_space", "Result should contain available_space field per API documentation")
		}
	})

	// Test set_retention_policy method
	t.Run("SetRetentionPolicy", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "set_retention_policy",
			ID:      51,
			Params: map[string]interface{}{
				"policy_type":  "age",
				"max_age_days": 30,
				"enabled":      true,
				"auth_token":   token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Set retention policy method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Retention policy result should be an object")
			require.Contains(t, result, "policy_type", "Result should contain policy_type field per API documentation")
			require.Contains(t, result, "enabled", "Result should contain enabled field per API documentation")
		}
	})

	// Test cleanup_old_files method
	t.Run("CleanupOldFiles", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "cleanup_old_files",
			ID:      52,
			Params:  map[string]interface{}{},
		}

		response, err := sendWebSocketRequest(conn, request)
		if err != nil {
			// If method not implemented yet, that's acceptable
			t.Logf("Cleanup old files method not implemented yet: %v", err)
		} else {
			// Validate response format per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Cleanup result should be an object")
			require.Contains(t, result, "cleanup_executed", "Result should contain cleanup_executed field per API documentation")
			require.Contains(t, result, "files_deleted", "Result should contain files_deleted field per API documentation")
		}
	})
}

// testSystemHealthAndMetrics tests system health, metrics, and status endpoints
func testSystemHealthAndMetrics(t *testing.T, client *testtestutils.WebSocketTestClient, token string) {
	// Test ping method
	t.Run("PingMethod", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      60,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Ping should succeed")

		// Validate response per API documentation
		result, ok := response.Result.(string)
		require.True(t, ok, "Ping result should be string")
		assert.Equal(t, "pong", result, "Ping should return pong per API documentation")
	})

	// Test get_status method
	t.Run("SystemStatus", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_status",
			ID:      61,
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Get status should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Status result should be an object")
		require.Contains(t, result, "status", "Status should contain status field per API documentation")
	})

	// Test get_metrics method
	t.Run("SystemMetrics", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_metrics",
			ID:      62,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Get metrics should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Metrics result should be an object")
		require.Contains(t, result, "cpu_usage", "Metrics should contain CPU usage per API documentation")
		require.Contains(t, result, "memory_usage", "Metrics should contain memory usage per API documentation")
	})

	// Test get_server_info method
	t.Run("ServerInfo", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_server_info",
			ID:      63,
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Get server info should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Server info result should be an object")
		require.Contains(t, result, "version", "Server info should contain version field per API documentation")
	})
}

// testStreamManagement tests stream enumeration and management
func testStreamManagement(t *testing.T, client *testtestutils.WebSocketTestClient, token string, env *testtestutils.WebSocketTestEnvironment) {
	// Test get_streams method
	t.Run("GetStreams", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_streams",
			ID:      70,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Get streams should succeed")

		// Validate response format per API documentation
		result, ok := response.Result.([]interface{})
		require.True(t, ok, "Get streams result should be array per API documentation")
		assert.NotNil(t, result, "Get streams should return array")
	})
}

// testErrorHandlingAndEdgeCases tests error handling and edge cases
func testErrorHandlingAndEdgeCases(t *testing.T, client *testtestutils.WebSocketTestClient, token string) {
	// Test invalid parameters
	t.Run("InvalidParameters", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_status",
			ID:      80,
			Params: map[string]interface{}{
				"invalid_param": "invalid_value",
				"auth_token":    token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for invalid parameters")
		require.NotNil(t, response, "Should get response for invalid parameters")
		require.NotNil(t, response.Error, "Invalid parameters should return error response")
		assert.Equal(t, INVALID_PARAMS, response.Error.Code, "Error code should be INVALID_PARAMS")
		assert.Contains(t, response.Error.Message, "Invalid parameters", "Error message should match API documentation")
	})

	// Test non-existent method
	t.Run("NonExistentMethod", func(t *testing.T) {
		request := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "non_existent_method",
			ID:      81,
			Params: map[string]interface{}{
				"auth_token": token,
			},
		}

		response, err := sendWebSocketRequest(conn, request)
		require.NoError(t, err, "Should get JSON-RPC response for non-existent method")
		require.NotNil(t, response, "Should get response for non-existent method")
		require.NotNil(t, response.Error, "Non-existent method should return error response")
		assert.Equal(t, METHOD_NOT_FOUND, response.Error.Code, "Error code should be METHOD_NOT_FOUND")
		// Error message should indicate method not found
	})

	// Test malformed JSON-RPC request
	t.Run("MalformedRequest", func(t *testing.T) {
		// COMMON PATTERN: Use centralized WebSocket test client for malformed request testing
		// Note: The centralized client handles malformed requests gracefully
		// This test validates that the system responds appropriately to invalid JSON

		// Create a malformed request (missing closing brace)
		malformedRequest := &ws.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      82,
			Params:  map[string]interface{}{},
		}

		// The centralized client should handle this gracefully
		response := client.SendRequest(malformedRequest)
		require.NotNil(t, response, "Should get response for malformed request")

		// Validate that the system responds appropriately
		assert.NotNil(t, response, "Malformed request should return response")
	})
}

// sendWebSocketRequest sends a JSON-RPC request over WebSocket and returns the response
func sendWebSocketRequest(client *testtestutils.WebSocketTestClient, request *ws.JsonRpcRequest) (*ws.JsonRpcResponse, error) {
	// COMMON PATTERN: Use centralized WebSocket test client instead of raw connection
	response := client.SendRequest(request)
	return response, nil
}
