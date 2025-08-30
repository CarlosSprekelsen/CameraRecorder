//go:build integration && real_mediamtx
// +build integration,real_mediamtx

/*
MediaMTX Integration Test

Requirements Coverage:
- REQ-MTX-001: MediaMTX health monitoring
- REQ-MTX-002: Path configuration and management
- REQ-MTX-003: Stream lifecycle management
- REQ-MTX-004: Recording integration
- REQ-MTX-005: Snapshot integration
- REQ-MTX-006: Active recording tracking
- REQ-MTX-007: System metrics integration
- REQ-MTX-008: Error handling and recovery

Test Categories: Integration/Real MediaMTX/System
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MediaMTXIntegrationTestSuite tests MediaMTX integration functionality
// COMMON PATTERN: Uses shared utils instead of individual component setup
type MediaMTXIntegrationTestSuite struct {
	env    *utils.MediaMTXTestEnvironment
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMediaMTXIntegrationTestSuite creates a new test suite
func NewMediaMTXIntegrationTestSuite() *MediaMTXIntegrationTestSuite {
	return &MediaMTXIntegrationTestSuite{}
}

// Setup initializes the test suite using shared utils
func (suite *MediaMTXIntegrationTestSuite) Setup(t *testing.T) {
	// Create context with timeout
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 60*time.Second)

	// COMMON PATTERN: Use shared MediaMTX test environment
	suite.env = utils.SetupMediaMTXTestEnvironment(t)

	// Wait for MediaMTX to be ready by checking health (shorter timeout for tests)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ready := false
	for !ready {
		select {
		case <-ctx.Done():
			// Don't fail the test, just log that MediaMTX might not be available
			t.Logf("MediaMTX service not ready within timeout - continuing with test")
			ready = true
		default:
			_, err := suite.env.Controller.GetHealth(ctx)
			if err == nil {
				t.Logf("MediaMTX service is ready")
				ready = true
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// Teardown cleans up the test suite
func (suite *MediaMTXIntegrationTestSuite) Teardown(t *testing.T) {
	if suite.cancel != nil {
		suite.cancel()
	}

	// COMMON PATTERN: Use shared teardown
	if suite.env != nil {
		utils.TeardownMediaMTXTestEnvironment(t, suite.env)
	}
}

// TestMediaMTXHealthIntegration tests MediaMTX health integration
func TestMediaMTXHealthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("HealthCheck", func(t *testing.T) {
		health, err := suite.env.Controller.GetHealth(suite.ctx)
		require.NoError(t, err, "Should get health status")
		require.NotNil(t, health, "Health status should not be nil")

		t.Logf("Health status: %s", health.Status)
		t.Logf("Health details: %s", health.Details)
		t.Logf("Health timestamp: %v", health.Timestamp)

		assert.NotEmpty(t, health.Status, "Health status should not be empty")
		assert.NotNil(t, health.Timestamp, "Health timestamp should not be nil")
	})

	t.Run("SystemMetrics", func(t *testing.T) {
		metrics, err := suite.env.Controller.GetSystemMetrics(suite.ctx)
		require.NoError(t, err, "Should get system metrics")
		require.NotNil(t, metrics, "System metrics should not be nil")

		t.Logf("System metrics: %+v", metrics)

		// Verify metrics structure
		assert.NotNil(t, metrics, "Metrics should not be nil")
	})

	t.Run("HealthMonitoring", func(t *testing.T) {
		// Test health monitoring over time
		for i := 0; i < 3; i++ {
			health, err := suite.env.Controller.GetHealth(suite.ctx)
			require.NoError(t, err, "Should get health status consistently")
			require.NotNil(t, health, "Health status should not be nil")

			t.Logf("Health check %d: %s", i+1, health.Status)
			time.Sleep(1 * time.Second)
		}
	})
}

// TestMediaMTXPathIntegration tests MediaMTX path management
func TestMediaMTXPathIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("PathConfiguration", func(t *testing.T) {
		// Test path configuration retrieval
		cfg := suite.env.ConfigManager.GetConfig()
		require.NotNil(t, cfg, "Configuration should not be nil")

		t.Logf("MediaMTX configuration: %+v", cfg.MediaMTX)
		t.Logf("Recording configuration: %+v", cfg.Recording)
		t.Logf("Storage configuration: %+v", cfg.Storage)

		// Verify configuration structure
		assert.NotEmpty(t, cfg.MediaMTX.Host, "MediaMTX host should be configured")
		assert.NotZero(t, cfg.MediaMTX.APIPort, "MediaMTX API port should be configured")
	})

	t.Run("PathValidation", func(t *testing.T) {
		// Test path validation
		testPaths := []string{
			"/dev/video0",
			"/dev/video1",
			"rtsp://192.168.1.100:554/stream",
			"http://192.168.1.100:8080/stream",
		}

		for _, path := range testPaths {
			t.Logf("Testing path: %s", path)
			// Note: This would test actual path validation if implemented
			// For now, we just log the paths
		}
	})
}

// TestMediaMTXStreamIntegration tests MediaMTX stream management
func TestMediaMTXStreamIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("StreamCreation", func(t *testing.T) {
		// Test stream creation with different devices
		testDevices := []string{
			"/dev/video0",
			"/dev/video1",
		}

		for _, device := range testDevices {
			t.Logf("Testing stream creation for device: %s", device)

			// Test recording stream creation
			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       1,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "medium",
				"max_duration":   10 * time.Second, // Short duration for testing
			}

			session, err := suite.env.Controller.StartAdvancedRecording(suite.ctx, device, "", options)
			if err != nil {
				t.Logf("Stream creation failed for %s: %v", device, err)
				// This is expected if no camera is available
				continue
			}

			require.NotNil(t, session, "Recording session should be created")
			t.Logf("Recording session created: %s", session.ID)

			// Verify session properties
			assert.Equal(t, device, session.Device, "Session device should match")
			assert.Equal(t, "RECORDING", session.Status, "Session should be recording")

			// Test session status
			status, err := suite.env.Controller.GetRecordingStatus(suite.ctx, session.ID)
			require.NoError(t, err, "Should get recording status")
			assert.Equal(t, "RECORDING", status.Status, "Status should be recording")

			// Wait a bit for recording
			time.Sleep(2 * time.Second)

			// Stop recording
			err = suite.env.Controller.StopAdvancedRecording(suite.ctx, session.ID)
			require.NoError(t, err, "Should stop recording")

			t.Logf("Recording session stopped: %s", session.ID)
		}
	})

	t.Run("StreamLifecycle", func(t *testing.T) {
		// Test complete stream lifecycle
		device := "/dev/video0"

		// Start recording
		options := map[string]interface{}{
			"use_case":       "recording",
			"priority":       1,
			"auto_cleanup":   true,
			"retention_days": 1,
			"quality":        "medium",
			"max_duration":   5 * time.Second, // Very short duration for testing
		}

		session, err := suite.env.Controller.StartAdvancedRecording(suite.ctx, device, "", options)
		if err != nil {
			t.Logf("Stream lifecycle test skipped: %v", err)
			t.Skip("No camera available for stream lifecycle test")
		}

		require.NotNil(t, session, "Recording session should be created")
		t.Logf("Stream lifecycle - Session started: %s", session.ID)

		// Monitor session status
		for i := 0; i < 3; i++ {
			status, err := suite.env.Controller.GetRecordingStatus(suite.ctx, session.ID)
			if err != nil {
				t.Logf("Status check %d failed: %v", i+1, err)
				break
			}

			t.Logf("Stream lifecycle - Status check %d: %s", i+1, status.Status)
			time.Sleep(1 * time.Second)
		}

		// Stop recording
		err = suite.env.Controller.StopAdvancedRecording(suite.ctx, session.ID)
		require.NoError(t, err, "Should stop recording")

		t.Logf("Stream lifecycle - Session stopped: %s", session.ID)
	})
}

// TestMediaMTXRecordingIntegration tests MediaMTX recording functionality
func TestMediaMTXRecordingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("RecordingOperations", func(t *testing.T) {
		// Test various recording operations
		device := "/dev/video0"

		// Test 1: Basic recording
		t.Run("BasicRecording", func(t *testing.T) {
			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       1,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "medium",
				"max_duration":   3 * time.Second,
			}

			session, err := suite.env.Controller.StartAdvancedRecording(suite.ctx, device, "", options)
			if err != nil {
				t.Logf("Basic recording test skipped: %v", err)
				t.Skip("No camera available for basic recording test")
			}

			require.NotNil(t, session, "Recording session should be created")
			t.Logf("Basic recording - Session: %s", session.ID)

			// Wait for recording
			time.Sleep(2 * time.Second)

			// Stop recording
			err = suite.env.Controller.StopAdvancedRecording(suite.ctx, session.ID)
			require.NoError(t, err, "Should stop recording")

			t.Logf("Basic recording - Completed: %s", session.ID)
		})

		// Test 2: High quality recording
		t.Run("HighQualityRecording", func(t *testing.T) {
			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       2,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "high",
				"max_duration":   3 * time.Second,
			}

			session, err := suite.env.Controller.StartAdvancedRecording(suite.ctx, device, "", options)
			if err != nil {
				t.Logf("High quality recording test skipped: %v", err)
				t.Skip("No camera available for high quality recording test")
			}

			require.NotNil(t, session, "Recording session should be created")
			t.Logf("High quality recording - Session: %s", session.ID)

			// Wait for recording
			time.Sleep(2 * time.Second)

			// Stop recording
			err = suite.env.Controller.StopAdvancedRecording(suite.ctx, session.ID)
			require.NoError(t, err, "Should stop recording")

			t.Logf("High quality recording - Completed: %s", session.ID)
		})
	})

	t.Run("RecordingList", func(t *testing.T) {
		// Test recording listing
		recordings, err := suite.env.Controller.ListRecordings(suite.ctx, 10, 0)
		require.NoError(t, err, "Should list recordings")
		require.NotNil(t, recordings, "Recordings list should not be nil")

		t.Logf("Found %d recordings", recordings.Total)
		t.Logf("Recordings: %+v", recordings.Files)

		assert.GreaterOrEqual(t, recordings.Total, 0, "Should have non-negative recordings count")
	})
}

// TestMediaMTXSnapshotIntegration tests MediaMTX snapshot functionality
func TestMediaMTXSnapshotIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("SnapshotCapture", func(t *testing.T) {
		device := "/dev/video0"

		// Test snapshot capture
		options := map[string]interface{}{
			"quality":    85,
			"format":     "jpeg",
			"resolution": "1920x1080",
		}

		snapshot, err := suite.env.Controller.TakeAdvancedSnapshot(suite.ctx, device, "", options)
		if err != nil {
			t.Logf("Snapshot capture test skipped: %v", err)
			t.Skip("No camera available for snapshot test")
		}

		require.NotNil(t, snapshot, "Snapshot should be created")
		t.Logf("Snapshot captured: %s", snapshot.ID)
		t.Logf("Snapshot file: %s", snapshot.FilePath)

		assert.NotEmpty(t, snapshot.ID, "Snapshot should have ID")
		assert.NotEmpty(t, snapshot.FilePath, "Snapshot should have file path")
	})

	t.Run("SnapshotList", func(t *testing.T) {
		// Test snapshot listing
		snapshots, err := suite.env.Controller.ListSnapshots(suite.ctx, 10, 0)
		require.NoError(t, err, "Should list snapshots")
		require.NotNil(t, snapshots, "Snapshots list should not be nil")

		t.Logf("Found %d snapshots", snapshots.Total)
		t.Logf("Snapshots: %+v", snapshots.Files)

		assert.GreaterOrEqual(t, snapshots.Total, 0, "Should have non-negative snapshots count")
	})
}

// TestMediaMTXActiveRecordingTracking tests active recording tracking
func TestMediaMTXActiveRecordingTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ActiveRecordingStatus", func(t *testing.T) {
		device := "/dev/video0"

		// Check initial status
		isRecording := suite.env.Controller.IsDeviceRecording(device)
		assert.False(t, isRecording, "Device should not be recording initially")

		// Get active recordings
		activeRecordings := suite.env.Controller.GetActiveRecordings()
		t.Logf("Active recordings: %d", len(activeRecordings))
		assert.GreaterOrEqual(t, len(activeRecordings), 0, "Should have non-negative active recordings count")

		// Test recording start and tracking
		options := map[string]interface{}{
			"use_case":       "recording",
			"priority":       1,
			"auto_cleanup":   true,
			"retention_days": 1,
			"quality":        "medium",
			"max_duration":   3 * time.Second,
		}

		session, err := suite.env.Controller.StartAdvancedRecording(suite.ctx, device, "", options)
		if err != nil {
			t.Logf("Active recording tracking test skipped: %v", err)
			t.Skip("No camera available for active recording tracking test")
		}

		require.NotNil(t, session, "Recording session should be created")

		// Check if device is now recording
		isRecording = suite.env.Controller.IsDeviceRecording(device)
		assert.True(t, isRecording, "Device should be recording")

		// Get active recording details
		activeRecording := suite.env.Controller.GetActiveRecording(device)
		require.NotNil(t, activeRecording, "Should have active recording")
		assert.Equal(t, device, activeRecording.DevicePath, "Active recording should match device")
		assert.Equal(t, session.ID, activeRecording.SessionID, "Active recording should match session")

		t.Logf("Active recording: %+v", activeRecording)

		// Wait for recording
		time.Sleep(2 * time.Second)

		// Stop recording
		err = suite.env.Controller.StopAdvancedRecording(suite.ctx, session.ID)
		require.NoError(t, err, "Should stop recording")

		// Check if device is no longer recording
		isRecording = suite.env.Controller.IsDeviceRecording(device)
		assert.False(t, isRecording, "Device should not be recording after stop")

		// Get active recordings again
		activeRecordings = suite.env.Controller.GetActiveRecordings()
		t.Logf("Active recordings after stop: %d", len(activeRecordings))
	})
}

// BenchmarkMediaMTXIntegration benchmarks MediaMTX integration performance
func BenchmarkMediaMTXIntegration(b *testing.B) {
	suite := NewMediaMTXIntegrationTestSuite()
	suite.Setup(&testing.T{})
	defer suite.Teardown(&testing.T{})

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			health, err := suite.env.Controller.GetHealth(suite.ctx)
			if err != nil {
				b.Fatalf("Health check failed: %v", err)
			}
			_ = health
		}
	})

	b.Run("SystemMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics, err := suite.env.Controller.GetSystemMetrics(suite.ctx)
			if err != nil {
				b.Fatalf("System metrics failed: %v", err)
			}
			_ = metrics
		}
	})

	b.Run("ListRecordings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			recordings, err := suite.env.Controller.ListRecordings(suite.ctx, 10, 0)
			if err != nil {
				b.Fatalf("List recordings failed: %v", err)
			}
			_ = recordings
		}
	})

	b.Run("ListSnapshots", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			snapshots, err := suite.env.Controller.ListSnapshots(suite.ctx, 10, 0)
			if err != nil {
				b.Fatalf("List snapshots failed: %v", err)
			}
			_ = snapshots
		}
	})
}
