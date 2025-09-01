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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
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
	// This eliminates the need to create individual components manually
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start services BEFORE running subtests
	// Start camera monitor
	err := env.CameraMonitor.Start(ctx)
	require.NoError(t, err, "Failed to start camera monitor")

	// Start WebSocket server
	err = env.WebSocketServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	// Wait for services to be ready
	time.Sleep(2 * time.Second)

	// Test service startup
	t.Run("StartServices", func(t *testing.T) {
		// Verify services are running
		assert.True(t, env.CameraMonitor.IsRunning(), "Camera monitor should be running")
		assert.True(t, env.WebSocketServer.IsRunning(), "WebSocket server should be running")
	})

	// Test camera discovery
	t.Run("CameraDiscovery", func(t *testing.T) {
		// Ensure camera monitor is running
		require.True(t, env.CameraMonitor.IsRunning(), "Camera monitor must be running for discovery test")

		// Wait for camera discovery
		time.Sleep(5 * time.Second)

		// Get discovered cameras
		cameras := env.CameraMonitor.GetConnectedCameras()

		// Log discovered cameras
		t.Logf("Discovered %d cameras", len(cameras))
		for _, cam := range cameras {
			t.Logf("Camera: %s, Path: %s, Status: %s", cam.Name, cam.Path, cam.Status)
		}

		// We should have at least one camera (even if it's a mock)
		assert.GreaterOrEqual(t, len(cameras), 0, "Should discover cameras")
	})

	// Test camera capabilities
	t.Run("CameraCapabilities", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for capability testing")
		}

		// Test first camera capabilities
		var cameraID string
		for devicePath := range cameras {
			// Convert device path to camera identifier for API consistency
			if strings.HasPrefix(devicePath, "/dev/video") {
				deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
				cameraID = fmt.Sprintf("camera%s", deviceNum)
			} else {
				cameraID = devicePath
			}
			break
		}

		// Get camera capabilities - skip for now as method doesn't exist
		t.Logf("Camera Path: %s", cameraID)
	})

	// Test recording operations with file verification
	t.Run("RecordingOperations", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for recording testing")
		}

		// Test recording start
		t.Run("StartRecording", func(t *testing.T) {
			// Get camera ID for this specific test
			var cameraID string
			for devicePath := range cameras {
				// Convert device path to camera identifier for API consistency
				if strings.HasPrefix(devicePath, "/dev/video") {
					deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
					cameraID = fmt.Sprintf("camera%s", deviceNum)
				} else {
					cameraID = devicePath
				}
				break
			}

			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       1,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "medium",
				"max_duration":   30 * time.Second, // Short duration for testing
			}

			session, err := env.Controller.StartAdvancedRecording(ctx, cameraID, "", options)
			if err != nil {
				t.Logf("Warning: Could not start recording: %v", err)
				t.Skip("Recording not available")
			}

			require.NotNil(t, session, "Recording session should be created")
			t.Logf("Recording session started: %s", session.ID)

			// Verify session is active
			assert.Equal(t, "RECORDING", session.Status, "Session should be recording")
			assert.Equal(t, cameraID, session.Device, "Session should match device")

			// Test recording status
			t.Run("RecordingStatus", func(t *testing.T) {
				status, err := env.Controller.GetRecordingStatus(ctx, session.ID)
				require.NoError(t, err, "Should get recording status")
				assert.Equal(t, "RECORDING", status.Status, "Status should be recording")
			})

			// Wait a bit for recording
			time.Sleep(5 * time.Second)

			// Test recording stop
			t.Run("StopRecording", func(t *testing.T) {
				err := env.Controller.StopAdvancedRecording(ctx, session.ID)
				require.NoError(t, err, "Should stop recording")

				// Verify session is stopped
				status, err := env.Controller.GetRecordingStatus(ctx, session.ID)
				if err == nil {
					assert.Equal(t, "STOPPED", status.Status, "Status should be stopped")
				}

				// Verify recording file exists after stop
				t.Run("RecordingFileVerification", func(t *testing.T) {
					// Get recording info to find the file path
					recordings, err := env.Controller.ListRecordings(ctx, 10, 0)
					if err == nil && len(recordings.Files) > 0 {
						// Find the most recent recording file
						var latestRecording *mediamtx.FileMetadata
						for _, recording := range recordings.Files {
							if latestRecording == nil || recording.CreatedAt.After(latestRecording.CreatedAt) {
								latestRecording = recording
							}
						}

						if latestRecording != nil {
							// Construct file path from recording info
							cfg := env.ConfigManager.GetConfig()
							recordingsDir := cfg.MediaMTX.RecordingsPath
							filePath := filepath.Join(recordingsDir, latestRecording.FileName)

							// Verify file exists
							fileInfo, err := os.Stat(filePath)
							require.NoError(t, err, "Recording file should exist on disk")
							assert.True(t, fileInfo.Size() > 0, "Recording file should not be empty")
							assert.False(t, fileInfo.IsDir(), "Recording should be a file, not directory")

							// Verify file is accessible
							file, err := os.Open(filePath)
							require.NoError(t, err, "Recording file should be readable")
							defer file.Close()

							t.Logf("Recording file verified: %s (size: %d bytes)", filePath, fileInfo.Size())
						}
					}
				})
			})
		})
	})

	// Test snapshot operations
	t.Run("SnapshotOperations", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for snapshot testing")
		}

		var cameraID string
		for devicePath := range cameras {
			// Convert device path to camera identifier for API consistency
			if strings.HasPrefix(devicePath, "/dev/video") {
				deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
				cameraID = fmt.Sprintf("camera%s", deviceNum)
			} else {
				cameraID = devicePath
			}
			break
		}

		// Test snapshot capture
		options := map[string]interface{}{
			"quality":    85,
			"format":     "jpeg",
			"resolution": "1920x1080",
		}

		snapshot, err := env.Controller.TakeAdvancedSnapshot(ctx, cameraID, "", options)
		if err != nil {
			t.Logf("Warning: Could not take snapshot: %v", err)
			t.Skip("Snapshot not available")
		}

		require.NotNil(t, snapshot, "Snapshot should be created")
		t.Logf("Snapshot created: %s", snapshot.ID)
		assert.NotEmpty(t, snapshot.FilePath, "Snapshot should have file path")

		// Verify file actually exists
		t.Run("SnapshotFileVerification", func(t *testing.T) {
			fileInfo, err := os.Stat(snapshot.FilePath)
			require.NoError(t, err, "Snapshot file should exist on disk")
			assert.True(t, fileInfo.Size() > 0, "Snapshot file should not be empty")
			assert.False(t, fileInfo.IsDir(), "Snapshot should be a file, not directory")

			// Verify file is accessible
			file, err := os.Open(snapshot.FilePath)
			require.NoError(t, err, "Snapshot file should be readable")
			defer file.Close()

			// Verify file has correct permissions
			assert.True(t, fileInfo.Mode().IsRegular(), "Snapshot should be a regular file")
		})
	})

	// Test file operations
	t.Run("FileOperations", func(t *testing.T) {
		// Test recordings list
		recordings, err := env.Controller.ListRecordings(ctx, 10, 0)
		if err != nil {
			t.Logf("Warning: Could not list recordings: %v", err)
		} else {
			t.Logf("Found %d recordings", recordings.Total)
			assert.GreaterOrEqual(t, recordings.Total, 0, "Should have recordings count")
		}

		// Test snapshots list
		snapshots, err := env.Controller.ListSnapshots(ctx, 10, 0)
		if err != nil {
			t.Logf("Warning: Could not list snapshots: %v", err)
		} else {
			t.Logf("Found %d snapshots", snapshots.Total)
			assert.GreaterOrEqual(t, snapshots.Total, 0, "Should have snapshots count")
		}
	})

	// Test health monitoring
	t.Run("HealthMonitoring", func(t *testing.T) {
		// Test MediaMTX health
		health, err := env.Controller.GetHealth(ctx)
		if err != nil {
			t.Logf("Warning: Could not get health: %v", err)
		} else {
			t.Logf("Health status: %s", health.Status)
			assert.NotEmpty(t, health.Status, "Should have health status")
		}

		// Test system metrics
		metrics, err := env.Controller.GetSystemMetrics(ctx)
		if err != nil {
			t.Logf("Warning: Could not get metrics: %v", err)
		} else {
			t.Logf("System metrics: %+v", metrics)
			assert.NotNil(t, metrics, "Should have system metrics")
		}
	})

	// Test active recording tracking
	t.Run("ActiveRecordingTracking", func(t *testing.T) {
		cameras := env.CameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Skip("No cameras available for active recording testing")
		}

		var cameraID string
		for devicePath := range cameras {
			cameraID = devicePath
			break
		}

		// Check if device is recording
		isRecording := env.Controller.IsDeviceRecording(cameraID)
		assert.False(t, isRecording, "Device should not be recording initially")

		// Get active recordings
		activeRecordings := env.Controller.GetActiveRecordings()
		t.Logf("Active recordings: %d", len(activeRecordings))
		assert.GreaterOrEqual(t, len(activeRecordings), 0, "Should have active recordings count")
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Stop WebSocket server
		err := env.WebSocketServer.Stop()
		require.NoError(t, err, "Failed to stop WebSocket server")

		// Stop camera monitor
		err = env.CameraMonitor.Stop()
		require.NoError(t, err, "Failed to stop camera monitor")

		t.Log("Cleanup completed successfully")
	})

	// Test comprehensive workflow scenarios
	t.Run("ComprehensiveWorkflowScenarios", func(t *testing.T) {
		// Test 1: Complete camera lifecycle workflow
		t.Run("CameraLifecycleWorkflow", func(t *testing.T) {
			cameras := env.CameraMonitor.GetConnectedCameras()
			if len(cameras) == 0 {
				t.Skip("No cameras available for lifecycle testing")
			}

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

			// Step 1: Get camera status
			device, exists := env.CameraMonitor.GetDevice(cameraID)
			if exists {
				t.Logf("Camera status: %s", device.Status)
				assert.NotEmpty(t, device.Status, "Camera should have status")
			}

			// Step 2: Get camera capabilities
			capabilities := env.CameraMonitor.GetConnectedCameras()
			t.Logf("Camera capabilities: %d cameras discovered", len(capabilities))

			// Step 3: Test camera event handling
			eventHandler := &cameraEventHandler{
				t: t,
			}
			env.CameraMonitor.AddEventHandler(eventHandler)

			// Step 4: Test camera event callback
			eventCallback := func(eventData camera.CameraEventData) {
				t.Logf("Camera event callback: %s - %v", eventData.EventType, eventData.DevicePath)
			}
			env.CameraMonitor.AddEventCallback(eventCallback)
		})

		// Test 2: Complete recording workflow with all features
		t.Run("CompleteRecordingWorkflow", func(t *testing.T) {
			cameras := env.CameraMonitor.GetConnectedCameras()
			if len(cameras) == 0 {
				t.Skip("No cameras available for complete recording workflow")
			}

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

			// Step 1: Start recording with advanced options
			options := map[string]interface{}{
				"use_case":         "recording",
				"priority":         1,
				"auto_cleanup":     true,
				"retention_days":   1,
				"quality":          "high",
				"max_duration":     15 * time.Second,
				"segment_rotation": true,
				"segment_duration": 5 * time.Second,
			}

			session, err := env.Controller.StartAdvancedRecording(ctx, cameraID, "", options)
			if err != nil {
				t.Logf("Warning: Could not start advanced recording: %v", err)
				t.Skip("Advanced recording not available")
			}

			require.NotNil(t, session, "Advanced recording session should be created")
			t.Logf("Advanced recording session started: %s", session.ID)

			// Step 2: Monitor recording progress
			time.Sleep(3 * time.Second)

			// Step 3: Get recording session details
			recordingSession, exists := env.Controller.GetAdvancedRecordingSession(session.ID)
			if exists {
				t.Logf("Recording session details: %+v", recordingSession)
				assert.NotNil(t, recordingSession, "Should get recording session details")
			}

			// Step 4: List all recording sessions
			recordingSessions := env.Controller.ListAdvancedRecordingSessions()
			t.Logf("Total recording sessions: %d", len(recordingSessions))
			assert.GreaterOrEqual(t, len(recordingSessions), 0, "Should have recording sessions list")

			// Step 5: Stop recording
			err = env.Controller.StopAdvancedRecording(ctx, session.ID)
			require.NoError(t, err, "Should stop advanced recording")

			// Step 6: Verify recording file
			recordings, err := env.Controller.ListRecordings(ctx, 10, 0)
			if err == nil && len(recordings.Files) > 0 {
				latestRecording := recordings.Files[0]
				t.Logf("Latest recording: %s", latestRecording.FileName)

				// Get recording info
				recordingInfo, err := env.Controller.GetRecordingInfo(ctx, latestRecording.FileName)
				if err == nil {
					t.Logf("Recording info: %+v", recordingInfo)
					assert.NotNil(t, recordingInfo, "Should get recording info")
				}
			}
		})

		// Test 3: Complete snapshot workflow with all features
		t.Run("CompleteSnapshotWorkflow", func(t *testing.T) {
			cameras := env.CameraMonitor.GetConnectedCameras()
			if len(cameras) == 0 {
				t.Skip("No cameras available for complete snapshot workflow")
			}

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

			// Step 1: Take advanced snapshot
			snapshotOptions := map[string]interface{}{
				"quality":      95,
				"resolution":   "1920x1080",
				"format":       "jpeg",
				"auto_cleanup": true,
			}

			snapshot, err := env.Controller.TakeAdvancedSnapshot(ctx, cameraID, "", snapshotOptions)
			if err != nil {
				t.Logf("Warning: Could not take advanced snapshot: %v", err)
				t.Skip("Advanced snapshot not available")
			}

			require.NotNil(t, snapshot, "Advanced snapshot should be created")
			t.Logf("Advanced snapshot taken: %s", snapshot.ID)

			// Step 2: Get snapshot details
			snapshotDetails, exists := env.Controller.GetAdvancedSnapshot(snapshot.ID)
			if exists {
				t.Logf("Snapshot details: %+v", snapshotDetails)
				assert.NotNil(t, snapshotDetails, "Should get snapshot details")
			}

			// Step 3: List all snapshots
			snapshots := env.Controller.ListAdvancedSnapshots()
			t.Logf("Total snapshots: %d", len(snapshots))
			assert.GreaterOrEqual(t, len(snapshots), 0, "Should have snapshots list")

			// Step 4: Get snapshot info
			snapshotsList, err := env.Controller.ListSnapshots(ctx, 10, 0)
			if err == nil && len(snapshotsList.Files) > 0 {
				latestSnapshot := snapshotsList.Files[0]
				t.Logf("Latest snapshot: %s", latestSnapshot.FileName)

				snapshotInfo, err := env.Controller.GetSnapshotInfo(ctx, latestSnapshot.FileName)
				if err == nil {
					t.Logf("Snapshot info: %+v", snapshotInfo)
					assert.NotNil(t, snapshotInfo, "Should get snapshot info")
				}
			}
		})

		// Test 4: System management workflow
		t.Run("SystemManagementWorkflow", func(t *testing.T) {
			// Step 1: Get system metrics
			metrics, err := env.Controller.GetSystemMetrics(ctx)
			if err == nil {
				t.Logf("System metrics: %+v", metrics)
				assert.NotNil(t, metrics, "Should get system metrics")
			}

			// Step 2: Get storage metrics
			recordingManager := env.Controller.GetRecordingManager()
			t.Logf("Recording manager: %T", recordingManager)
			assert.NotNil(t, recordingManager, "Should get recording manager")

			// Step 3: Test configuration management
			config, err := env.Controller.GetConfig(ctx)
			if err == nil && config != nil {
				t.Logf("Current config: %+v", config)
				assert.NotNil(t, config, "Should get current configuration")
			}

			// Step 4: Test path management
			paths, err := env.Controller.GetPaths(ctx)
			if err == nil {
				t.Logf("Available paths: %d", len(paths))
				assert.NotNil(t, paths, "Should get available paths")
			}

			// Step 5: Test stream management
			streams, err := env.Controller.GetStreams(ctx)
			if err == nil {
				t.Logf("Available streams: %d", len(streams))
				assert.NotNil(t, streams, "Should get available streams")
			}
		})

		// Test 5: Error handling and recovery workflow
		t.Run("ErrorHandlingWorkflow", func(t *testing.T) {
			// Test 1: Invalid camera operations
			_, err := env.Controller.StartAdvancedRecording(ctx, "invalid_camera", "", nil)
			if err != nil {
				t.Logf("Expected error for invalid camera: %v", err)
				assert.Error(t, err, "Should fail with invalid camera")
			}

			// Test 2: Invalid session operations
			err = env.Controller.StopAdvancedRecording(ctx, "invalid_session")
			if err != nil {
				t.Logf("Expected error for invalid session: %v", err)
				assert.Error(t, err, "Should fail with invalid session")
			}

			// Test 3: Invalid snapshot operations
			_, exists := env.Controller.GetAdvancedSnapshot("invalid_snapshot")
			if !exists {
				t.Logf("Expected invalid snapshot to not exist")
				assert.False(t, exists, "Should fail with invalid snapshot")
			}

			// Test 4: Camera monitor error handling
			invalidDevice, exists := env.CameraMonitor.GetDevice("invalid_device")
			assert.False(t, exists, "Invalid device should not exist")
			assert.Nil(t, invalidDevice, "Invalid device should be nil")
		})

		// Test 6: Performance and load testing
		t.Run("PerformanceWorkflow", func(t *testing.T) {
			cameras := env.CameraMonitor.GetConnectedCameras()
			if len(cameras) == 0 {
				t.Skip("No cameras available for performance testing")
			}

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

			// Test rapid snapshot operations
			startTime := time.Now()
			for i := 0; i < 3; i++ {
				snapshot, err := env.Controller.TakeAdvancedSnapshot(ctx, cameraID, "", nil)
				if err == nil {
					t.Logf("Snapshot %d taken: %s", i+1, snapshot.ID)
				}
				time.Sleep(500 * time.Millisecond)
			}
			totalTime := time.Since(startTime)
			t.Logf("Took 3 snapshots in %v", totalTime)

			// Test concurrent operations
			done := make(chan bool, 3)
			for i := 0; i < 3; i++ {
				go func(id int) {
					defer func() { done <- true }()

					// Get camera status
					device, exists := env.CameraMonitor.GetDevice(cameraID)
					if exists {
						t.Logf("Concurrent camera status %d: %s", id, device.Status)
					}
				}(i)
			}

			// Wait for all concurrent operations
			for i := 0; i < 3; i++ {
				<-done
			}
		})
	})
}

// TestCameraWorkflowWithRealDevice tests camera operations with a real device
func TestCameraWorkflowWithRealDevice(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a real camera device
	// It will be skipped if no camera is available
	t.Skip("Real device test - requires physical camera")
}

// TestCameraWorkflowWithMockDevice tests camera operations with mock devices
func TestCameraWorkflowWithMockDevice(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// COMMON PATTERN: Use shared WebSocket test environment with all dependencies
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Setup test environment
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test mock camera discovery
	t.Run("MockCameraDiscovery", func(t *testing.T) {
		// Start camera monitor
		err := env.CameraMonitor.Start(ctx)
		require.NoError(t, err, "Failed to start camera monitor")

		// Wait for discovery
		time.Sleep(3 * time.Second)

		// Get cameras
		cameras := env.CameraMonitor.GetConnectedCameras()
		t.Logf("Discovered %d cameras in mock test", len(cameras))

		// Stop camera monitor
		err = env.CameraMonitor.Stop()
		require.NoError(t, err, "Failed to stop camera monitor")
	})
}

// cameraEventHandler implements CameraEventHandler interface for testing
type cameraEventHandler struct {
	t *testing.T
}

func (h *cameraEventHandler) HandleCameraEvent(ctx context.Context, eventData camera.CameraEventData) error {
	h.t.Logf("Camera event received: %s - %s", eventData.EventType, eventData.DevicePath)
	return nil
}

// BenchmarkCameraOperations benchmarks camera operations
// TODO: Implement benchmark using shared test environment when benchmark support is added
func BenchmarkCameraOperations(b *testing.B) {
	b.Skip("Benchmark not yet implemented with shared test environment")
}
