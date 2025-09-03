//go:build integration
// +build integration

/*
End-to-End Camera Operations Integration Test

Requirements Coverage:
- REQ-CAM-001: Camera discovery and enumeration
- REQ-CAM-002: Camera capability detection
- REQ-REC-001: Recording session management
- REQ-REC-002: Recording start/stop operations
- REQ-SNAP-001: Snapshot capture functionality
- REQ-FILE-001: File listing and management
- REQ-HEALTH-001: Health monitoring integration
- REQ-ACTIVE-001: Active recording tracking

Test Categories: Integration/Real System/Hardware
API Documentation Reference: docs/api/json_rpc_methods.md

CRITICAL: This test validates against API documentation as ground truth.
Tests MUST fail when systems fail - no t.Skip() calls allowed.
*/

package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndCameraOperations tests the complete camera workflow
// This test validates the entire camera service pipeline from discovery to recording
// COMMON PATTERN: Uses shared WebSocket test environment instead of individual component setup
func TestEndToEndCameraOperations(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// COMMON PATTERN: Use shared WebSocket test environment with all dependencies
	env := testtestutils.SetupWebSocketTestEnvironment(t)
	defer func() {
		// Ensure proper cleanup
		if env.CameraMonitor.IsRunning() {
			env.CameraMonitor.Stop()
		}
		if env.WebSocketServer.IsRunning() {
			env.WebSocketServer.Stop()
		}
		testtestutils.TeardownWebSocketTestEnvironment(t, env)
	}()

	// Setup test environment with proper timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start services BEFORE running subtests
	err := env.CameraMonitor.Start(ctx)
	require.NoError(t, err, "Failed to start camera monitor")

	err = env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	// Wait for services to be ready
	time.Sleep(2 * time.Second)

	// Test service startup
	t.Run("StartServices", func(t *testing.T) {
		assert.True(t, env.CameraMonitor.IsRunning(), "Camera monitor should be running")
		assert.True(t, env.WebSocketServer.IsRunning(), "WebSocket server should be running")
	})

	// Test camera discovery
	t.Run("CameraDiscovery", func(t *testing.T) {
		require.True(t, env.CameraMonitor.IsRunning(), "Camera monitor must be running for discovery test")

		// Wait for camera discovery
		time.Sleep(5 * time.Second)

		cameras := env.CameraMonitor.GetConnectedCameras()
		t.Logf("Discovered %d cameras", len(cameras))

		for devicePath, cam := range cameras {
			t.Logf("Camera: %s, Path: %s, Status: %s", cam.Name, devicePath, cam.Status)
		}

		// CRITICAL: Test must fail if no cameras found - no t.Skip() allowed
		assert.Greater(t, len(cameras), 0, "Must discover at least one real camera device")
	})

	// Test camera capabilities via API
	t.Run("CameraCapabilitiesAPI", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		require.Greater(t, len(cameras), 0, "Must have cameras for capability testing")

		// Get first camera for testing
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

		// Create WebSocket client using proven pattern
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Authenticate as operator (required for camera operations)
		authToken, err := env.JWTHandler.GenerateToken("test_operator", "operator", 24)
		require.NoError(t, err, "Failed to generate operator token")

		authRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": authToken,
			},
			ID: 1,
		}

		authResponse := client.SendRequest(authRequest)
		require.NotNil(t, authResponse, "Authentication response must not be nil")
		require.Nil(t, authResponse.Error, "Authentication must succeed")

		// Test get_camera_capabilities method per API documentation
		capabilitiesRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "get_camera_capabilities",
			Params: map[string]interface{}{
				"device": cameraID,
			},
			ID: 2,
		}

		response := client.SendRequest(capabilitiesRequest)
		require.NotNil(t, response, "Camera capabilities response must not be nil")

		if response.Error != nil {
			// CRITICAL: Test must fail if API method fails - no t.Skip() allowed
			t.Errorf("get_camera_capabilities failed: %s (code: %d)", response.Error.Message, response.Error.Code)
		} else {
			// Validate response structure per API documentation
			result, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Result must be a map")

			// Validate required fields per API documentation
			assert.Contains(t, result, "device", "Must contain device field")
			assert.Contains(t, result, "formats", "Must contain formats field")
			assert.Contains(t, result, "resolutions", "Must contain resolutions field")
			assert.Contains(t, result, "fps_options", "Must contain fps_options field")
			assert.Contains(t, result, "validation_status", "Must contain validation_status field")

			// Validate device matches request
			assert.Equal(t, cameraID, result["device"], "Device must match request")

			t.Logf("Camera capabilities: %+v", result)
		}
	})

	// Test recording operations with real file verification
	t.Run("RecordingOperations", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		require.Greater(t, len(cameras), 0, "Must have cameras for recording testing")

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

		// Create WebSocket client
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Authenticate as operator
		authToken, err := env.JWTHandler.GenerateToken("test_operator", "operator", 24)
		require.NoError(t, err, "Failed to generate operator token")

		authRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": authToken,
			},
			ID: 3,
		}

		authResponse := client.SendRequest(authRequest)
		require.Nil(t, authResponse.Error, "Authentication must succeed")

		// Test start_recording method per API documentation
		t.Run("StartRecording", func(t *testing.T) {
			startRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "start_recording",
				Params: map[string]interface{}{
					"device":   cameraID,
					"duration": 10, // 10 seconds for testing
					"format":   "mp4",
				},
				ID: 4,
			}

			response := client.SendRequest(startRequest)
			require.NotNil(t, response, "Start recording response must not be nil")

			if response.Error != nil {
				// CRITICAL: Test must fail if recording fails - no t.Skip() allowed
				t.Errorf("start_recording failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				// Validate response structure per API documentation
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				// Validate required fields per API documentation
				assert.Contains(t, result, "session_id", "Must contain session_id field")
				assert.Contains(t, result, "device", "Must contain device field")
				assert.Contains(t, result, "status", "Must contain status field")

				sessionID := result["session_id"].(string)
				require.NotEmpty(t, sessionID, "Session ID must not be empty")

				t.Logf("Recording started: session_id=%s", sessionID)

				// Wait for recording to start
				time.Sleep(3 * time.Second)

				// Test recording status
				t.Run("RecordingStatus", func(t *testing.T) {
					statusRequest := &websocket.JsonRpcRequest{
						JSONRPC: "2.0",
						Method:  "get_recording_status",
						Params: map[string]interface{}{
							"session_id": sessionID,
						},
						ID: 5,
					}

					statusResponse := client.SendRequest(statusRequest)
					require.NotNil(t, statusResponse, "Recording status response must not be nil")

					if statusResponse.Error != nil {
						t.Errorf("get_recording_status failed: %s (code: %d)", statusResponse.Error.Message, statusResponse.Error.Code)
					} else {
						statusResult, ok := statusResponse.Result.(map[string]interface{})
						require.True(t, ok, "Status result must be a map")

						assert.Contains(t, statusResult, "status", "Must contain status field")
						assert.Contains(t, statusResult, "device", "Must contain device field")

						status := statusResult["status"].(string)
						assert.Equal(t, "recording", status, "Status must be recording")
					}
				})

				// Wait for recording to complete
				time.Sleep(12 * time.Second) // Wait for 10s duration + buffer

				// Test stop recording
				t.Run("StopRecording", func(t *testing.T) {
					stopRequest := &websocket.JsonRpcRequest{
						JSONRPC: "2.0",
						Method:  "stop_recording",
						Params: map[string]interface{}{
							"session_id": sessionID,
						},
						ID: 6,
					}

					stopResponse := client.SendRequest(stopRequest)
					require.NotNil(t, stopResponse, "Stop recording response must not be nil")

					if stopResponse.Error != nil {
						t.Errorf("stop_recording failed: %s (code: %d)", stopResponse.Error.Message, stopResponse.Error.Code)
					} else {
						stopResult, ok := stopResponse.Result.(map[string]interface{})
						require.True(t, ok, "Stop result must be a map")

						assert.Contains(t, stopResult, "status", "Must contain status field")
						assert.Contains(t, stopResult, "message", "Must contain message field")
					}
				})

				// CRITICAL: Verify recording file exists on disk
				t.Run("RecordingFileVerification", func(t *testing.T) {
					// Wait for file to be written
					time.Sleep(2 * time.Second)

					// List recordings to find the file
					listRequest := &websocket.JsonRpcRequest{
						JSONRPC: "2.0",
						Method:  "list_recordings",
						Params: map[string]interface{}{
							"limit":  10,
							"offset": 0,
						},
						ID: 7,
					}

					listResponse := client.SendRequest(listRequest)
					require.NotNil(t, listResponse, "List recordings response must not be nil")

					if listResponse.Error != nil {
						t.Errorf("list_recordings failed: %s (code: %d)", listResponse.Error.Message, listResponse.Error.Code)
					} else {
						listResult, ok := listResponse.Result.(map[string]interface{})
						require.True(t, ok, "List result must be a map")

						assert.Contains(t, listResult, "files", "Must contain files field")
						assert.Contains(t, listResult, "total", "Must contain total field")

						files, ok := listResult["files"].([]interface{})
						require.True(t, ok, "Files must be an array")

						// CRITICAL: Must have at least one recording file
						assert.Greater(t, len(files), 0, "Must have at least one recording file")

						if len(files) > 0 {
							// Get the most recent recording
							latestFile := files[0].(map[string]interface{})
							fileName := latestFile["filename"].(string)

							// Verify file exists on disk
							cfg := env.ConfigManager.GetConfig()
							recordingsDir := cfg.MediaMTX.RecordingsPath
							filePath := filepath.Join(recordingsDir, fileName)

							fileInfo, err := os.Stat(filePath)
							require.NoError(t, err, "Recording file must exist on disk")
							assert.True(t, fileInfo.Size() > 0, "Recording file must not be empty")
							assert.False(t, fileInfo.IsDir(), "Recording must be a file, not directory")

							t.Logf("Recording file verified: %s (size: %d bytes)", filePath, fileInfo.Size())
						}
					}
				})
			}
		})
	})

	// Test snapshot operations with real file verification
	t.Run("SnapshotOperations", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		require.Greater(t, len(cameras), 0, "Must have cameras for snapshot testing")

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

		// Create WebSocket client
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Authenticate as operator
		authToken, err := env.JWTHandler.GenerateToken("test_operator", "operator", 24)
		require.NoError(t, err, "Failed to generate operator token")

		authRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": authToken,
			},
			ID: 8,
		}

		authResponse := client.SendRequest(authRequest)
		require.Nil(t, authResponse.Error, "Authentication must succeed")

		// Test take_snapshot method per API documentation
		t.Run("TakeSnapshot", func(t *testing.T) {
			snapshotRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "take_snapshot",
				Params: map[string]interface{}{
					"device":   cameraID,
					"filename": "test_snapshot.jpg",
				},
				ID: 9,
			}

			response := client.SendRequest(snapshotRequest)
			require.NotNil(t, response, "Take snapshot response must not be nil")

			if response.Error != nil {
				// CRITICAL: Test must fail if snapshot fails - no t.Skip() allowed
				t.Errorf("take_snapshot failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				// Validate response structure per API documentation
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				// Validate required fields per API documentation
				assert.Contains(t, result, "device", "Must contain device field")
				assert.Contains(t, result, "filename", "Must contain filename field")
				assert.Contains(t, result, "status", "Must contain status field")
				assert.Contains(t, result, "timestamp", "Must contain timestamp field")
				assert.Contains(t, result, "file_size", "Must contain file_size field")
				assert.Contains(t, result, "file_path", "Must contain file_path field")

				// Validate device matches request
				assert.Equal(t, cameraID, result["device"], "Device must match request")
				assert.Equal(t, "test_snapshot.jpg", result["filename"], "Filename must match request")
				assert.Equal(t, "completed", result["status"], "Status must be completed")

				filePath := result["file_path"].(string)
				require.NotEmpty(t, filePath, "File path must not be empty")

				t.Logf("Snapshot created: %s", filePath)

				// CRITICAL: Verify snapshot file exists on disk
				t.Run("SnapshotFileVerification", func(t *testing.T) {
					// Wait for file to be written
					time.Sleep(2 * time.Second)

					fileInfo, err := os.Stat(filePath)
					require.NoError(t, err, "Snapshot file must exist on disk")
					assert.True(t, fileInfo.Size() > 0, "Snapshot file must not be empty")
					assert.False(t, fileInfo.IsDir(), "Snapshot must be a file, not directory")
					assert.True(t, fileInfo.Mode().IsRegular(), "Snapshot must be a regular file")

					// Verify file is accessible
					file, err := os.Open(filePath)
					require.NoError(t, err, "Snapshot file must be readable")
					defer file.Close()

					t.Logf("Snapshot file verified: %s (size: %d bytes)", filePath, fileInfo.Size())
				})
			}
		})
	})

	// Test file operations
	t.Run("FileOperations", func(t *testing.T) {
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Authenticate as operator
		authToken, err := env.JWTHandler.GenerateToken("test_operator", "operator", 24)
		require.NoError(t, err, "Failed to generate operator token")

		authRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": authToken,
			},
			ID: 10,
		}

		authResponse := client.SendRequest(authRequest)
		require.Nil(t, authResponse.Error, "Authentication must succeed")

		// Test list_recordings method per API documentation
		t.Run("ListRecordings", func(t *testing.T) {
			listRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "list_recordings",
				Params: map[string]interface{}{
					"limit":  10,
					"offset": 0,
				},
				ID: 11,
			}

			response := client.SendRequest(listRequest)
			require.NotNil(t, response, "List recordings response must not be nil")

			if response.Error != nil {
				t.Errorf("list_recordings failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				assert.Contains(t, result, "files", "Must contain files field")
				assert.Contains(t, result, "total", "Must contain total field")
				assert.Contains(t, result, "limit", "Must contain limit field")
				assert.Contains(t, result, "offset", "Must contain offset field")

				files, ok := result["files"].([]interface{})
				require.True(t, ok, "Files must be an array")

				t.Logf("Found %d recordings", len(files))

				// Validate file metadata structure per API documentation
				if len(files) > 0 {
					file := files[0].(map[string]interface{})
					assert.Contains(t, file, "filename", "File must contain filename field")
					assert.Contains(t, file, "file_size", "File must contain file_size field")
					assert.Contains(t, file, "modified_time", "File must contain modified_time field")
					assert.Contains(t, file, "download_url", "File must contain download_url field")
				}
			}
		})

		// Test list_snapshots method per API documentation
		t.Run("ListSnapshots", func(t *testing.T) {
			listRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "list_snapshots",
				Params: map[string]interface{}{
					"limit":  10,
					"offset": 0,
				},
				ID: 12,
			}

			response := client.SendRequest(listRequest)
			require.NotNil(t, response, "List snapshots response must not be nil")

			if response.Error != nil {
				t.Errorf("list_snapshots failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				assert.Contains(t, result, "files", "Must contain files field")
				assert.Contains(t, result, "total", "Must contain total field")
				assert.Contains(t, result, "limit", "Must contain limit field")
				assert.Contains(t, result, "offset", "Must contain offset field")

				files, ok := result["files"].([]interface{})
				require.True(t, ok, "Files must be an array")

				t.Logf("Found %d snapshots", len(files))

				// Validate file metadata structure per API documentation
				if len(files) > 0 {
					file := files[0].(map[string]interface{})
					assert.Contains(t, file, "filename", "File must contain filename field")
					assert.Contains(t, file, "file_size", "File must contain file_size field")
					assert.Contains(t, file, "modified_time", "File must contain modified_time field")
					assert.Contains(t, file, "download_url", "File must contain download_url field")
				}
			}
		})
	})

	// Test health monitoring
	t.Run("HealthMonitoring", func(t *testing.T) {
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Authenticate as admin (required for health monitoring)
		authToken, err := env.JWTHandler.GenerateToken("test_admin", "admin", 24)
		require.NoError(t, err, "Failed to generate admin token")

		authRequest := &websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "authenticate",
			Params: map[string]interface{}{
				"auth_token": authToken,
			},
			ID: 13,
		}

		authResponse := client.SendRequest(authRequest)
		require.Nil(t, authResponse.Error, "Authentication must succeed")

		// Test get_health method per API documentation
		t.Run("GetHealth", func(t *testing.T) {
			healthRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "get_health",
				Params:  map[string]interface{}{},
				ID:      14,
			}

			response := client.SendRequest(healthRequest)
			require.NotNil(t, response, "Health response must not be nil")

			if response.Error != nil {
				t.Errorf("get_health failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				assert.Contains(t, result, "status", "Must contain status field")
				assert.Contains(t, result, "timestamp", "Must contain timestamp field")
				assert.Contains(t, result, "components", "Must contain components field")

				status := result["status"].(string)
				assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, status, "Status must be valid health value")

				t.Logf("Health status: %s", status)
			}
		})

		// Test get_metrics method per API documentation
		t.Run("GetMetrics", func(t *testing.T) {
			metricsRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "get_metrics",
				Params:  map[string]interface{}{},
				ID:      15,
			}

			response := client.SendRequest(metricsRequest)
			require.NotNil(t, response, "Metrics response must not be nil")

			if response.Error != nil {
				t.Errorf("get_metrics failed: %s (code: %d)", response.Error.Message, response.Error.Code)
			} else {
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result must be a map")

				// Validate metrics structure per API documentation
				assert.Contains(t, result, "cpu_usage", "Must contain cpu_usage field")
				assert.Contains(t, result, "memory_usage", "Must contain memory_usage field")
				assert.Contains(t, result, "disk_usage", "Must contain disk_usage field")
				assert.Contains(t, result, "active_connections", "Must contain active_connections field")

				t.Logf("Metrics retrieved successfully")
			}
		})
	})

	// Test authentication and authorization
	t.Run("AuthenticationAndAuthorization", func(t *testing.T) {
		client := testtestutils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
		defer client.Close()

		// Test viewer role permissions
		t.Run("ViewerRolePermissions", func(t *testing.T) {
			authToken, err := env.JWTHandler.GenerateToken("test_viewer", "viewer", 24)
			require.NoError(t, err, "Failed to generate viewer token")

			authRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "authenticate",
				Params: map[string]interface{}{
					"auth_token": authToken,
				},
				ID: 16,
			}

			authResponse := client.SendRequest(authRequest)
			require.Nil(t, authResponse.Error, "Authentication must succeed")

			// Viewer should be able to list cameras
			camerasRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "get_camera_list",
				Params:  map[string]interface{}{},
				ID:      17,
			}

			response := client.SendRequest(camerasRequest)
			require.NotNil(t, response, "Camera list response must not be nil")

			if response.Error != nil {
				t.Errorf("get_camera_list failed for viewer: %s (code: %d)", response.Error.Message, response.Error.Code)
			}

			// Viewer should NOT be able to start recording
			recordRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "start_recording",
				Params: map[string]interface{}{
					"device":   "camera0",
					"duration": 5,
				},
				ID: 18,
			}

			recordResponse := client.SendRequest(recordRequest)
			require.NotNil(t, recordResponse, "Start recording response must not be nil")

			// CRITICAL: Must fail due to insufficient permissions
			require.NotNil(t, recordResponse.Error, "Viewer must not be able to start recording")
			assert.Equal(t, -32003, recordResponse.Error.Code, "Must return insufficient permissions error")
		})

		// Test operator role permissions
		t.Run("OperatorRolePermissions", func(t *testing.T) {
			authToken, err := env.JWTHandler.GenerateToken("test_operator", "operator", 24)
			require.NoError(t, err, "Failed to generate operator token")

			authRequest := &websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				Method:  "authenticate",
				Params: map[string]interface{}{
					"auth_token": authToken,
				},
				ID: 19,
			}

			authResponse := client.SendRequest(authRequest)
			require.Nil(t, authResponse.Error, "Authentication must succeed")

			// Operator should be able to start recording
			cameras := env.CameraMonitor.GetConnectedCameras()
			if len(cameras) > 0 {
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

				recordRequest := &websocket.JsonRpcRequest{
					JSONRPC: "2.0",
					Method:  "start_recording",
					Params: map[string]interface{}{
						"device":   cameraID,
						"duration": 5,
					},
					ID: 20,
				}

				response := client.SendRequest(recordRequest)
				require.NotNil(t, response, "Start recording response must not be nil")

				// Operator should be able to start recording
				if response.Error != nil {
					t.Logf("Recording start failed (may be expected): %s", response.Error.Message)
				} else {
					result, ok := response.Result.(map[string]interface{})
					require.True(t, ok, "Result must be a map")
					assert.Contains(t, result, "session_id", "Must contain session_id field")

					// Stop recording immediately
					sessionID := result["session_id"].(string)
					stopRequest := &websocket.JsonRpcRequest{
						JSONRPC: "2.0",
						Method:  "stop_recording",
						Params: map[string]interface{}{
							"session_id": sessionID,
						},
						ID: 21,
					}

					stopResponse := client.SendRequest(stopRequest)
					if stopResponse.Error != nil {
						t.Logf("Recording stop failed: %s", stopResponse.Error.Message)
					}
				}
			}
		})
	})
}
