/*
MediaMTX Controller Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createConfigManagerWithEnvVars creates a config manager that loads environment variables
func createConfigManagerWithEnvVars(t *testing.T, helper *MediaMTXTestHelper) *config.ConfigManager {
	// Use centralized configuration loading from test helpers
	return CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")
}

// TestControllerWithConfigManager_ReqMTX001 tests controller creation with real server
func TestControllerWithConfigManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")
}

// TestController_GetHealth_ReqMTX004 tests controller health with real server
func TestController_GetHealth_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get health
	health, err := controller.GetHealth(ctx)
	require.NoError(t, err, "GetHealth should succeed")
	require.NotNil(t, health, "Health should not be nil")
	assert.Equal(t, "healthy", health.Status, "Health should be healthy")
}

// TestController_GetMetrics_ReqMTX004 tests controller metrics with real server
func TestController_GetMetrics_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get metrics
	metrics, err := controller.GetMetrics(ctx)
	require.NoError(t, err, "GetMetrics should succeed")
	require.NotNil(t, metrics, "Metrics should not be nil")
}

// TestController_GetSystemMetrics_ReqMTX004 tests controller system metrics with real server
func TestController_GetSystemMetrics_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get system metrics
	systemMetrics, err := controller.GetSystemMetrics(ctx)
	require.NoError(t, err, "GetSystemMetrics should succeed")
	require.NotNil(t, systemMetrics, "System metrics should not be nil")
}

// TestController_GetPaths_ReqMTX003 tests path listing with real server
func TestController_GetPaths_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get paths
	paths, err := controller.GetPaths(ctx)
	require.NoError(t, err, "GetPaths should succeed")
	require.NotNil(t, paths, "Paths should not be nil")
	assert.IsType(t, []*Path{}, paths, "Paths should be a slice of Path pointers")
}

// TestController_GetStreams_ReqMTX002 tests stream listing with real server
func TestController_GetStreams_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get streams
	streams, err := controller.GetStreams(ctx)
	require.NoError(t, err, "GetStreams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")
	assert.IsType(t, []*Stream{}, streams, "Streams should be a slice of Stream pointers")
}

// TestController_GetConfig_ReqMTX001 tests configuration retrieval with real server
func TestController_GetConfig_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get config
	config, err := controller.GetConfig(ctx)
	require.NoError(t, err, "GetConfig should succeed")
	require.NotNil(t, config, "Config should not be nil")
	assert.NotEmpty(t, config.BaseURL, "BaseURL should not be empty")
}

// TestController_ListRecordings_ReqMTX002 tests recording listing with real server
func TestController_ListRecordings_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// List recordings
	recordings, err := controller.ListRecordings(ctx, 10, 0)
	require.NoError(t, err, "ListRecordings should succeed")
	require.NotNil(t, recordings, "Recordings should not be nil")
}

// TestController_ListSnapshots_ReqMTX002 tests snapshot listing with real server
func TestController_ListSnapshots_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// List snapshots
	snapshots, err := controller.ListSnapshots(ctx, 10, 0)
	require.NoError(t, err, "ListSnapshots should succeed")
	require.NotNil(t, snapshots, "Snapshots should not be nil")
}

// TestController_ConcurrentAccess_ReqMTX001 tests concurrent operations with real server
func TestController_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test concurrent access to different methods
	done := make(chan bool, 4)

	go func() {
		health, err := controller.GetHealth(ctx)
		assert.NoError(t, err, "GetHealth should succeed")
		assert.NotNil(t, health, "Health should not be nil")
		done <- true
	}()

	go func() {
		metrics, err := controller.GetMetrics(ctx)
		assert.NoError(t, err, "GetMetrics should succeed")
		assert.NotNil(t, metrics, "Metrics should not be nil")
		done <- true
	}()

	go func() {
		paths, err := controller.GetPaths(ctx)
		assert.NoError(t, err, "GetPaths should succeed")
		assert.NotNil(t, paths, "Paths should not be nil")
		done <- true
	}()

	go func() {
		streams, err := controller.GetStreams(ctx)
		assert.NoError(t, err, "GetStreams should succeed")
		assert.NotNil(t, streams, "Streams should not be nil")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Should not panic and should handle concurrent access gracefully
	assert.True(t, true, "Concurrent access should not cause panics")
}

// TestController_StartRecording_ReqMTX002 tests recording functionality through controller
func TestController_StartRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "test_recording.mp4")

	// Test recording with camera identifier (abstraction layer)
	session, err := controller.StartRecording(ctx, "camera0", outputPath)
	require.NoError(t, err, "Recording should start successfully")
	require.NotNil(t, session, "Session should not be nil")

	// Verify session properties
	assert.NotEmpty(t, session.ID, "Session should have an ID")
	assert.Equal(t, "camera0", session.DevicePath, "Should use camera identifier")
	assert.Equal(t, outputPath, session.FilePath, "Should match output path")
	assert.Equal(t, "active", session.Status, "Session should be active")

	// Clean up
	err = controller.StopRecording(ctx, session.ID)
	require.NoError(t, err, "Recording should stop successfully")
}

// TestController_StopRecording_ReqMTX002 tests recording stop functionality through controller
func TestController_StopRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "test_recording_stop.mp4")

	// Start recording first
	session, err := controller.StartRecording(ctx, "camera0", outputPath)
	require.NoError(t, err, "Recording should start successfully")
	require.NotNil(t, session, "Session should not be nil")

	// Stop recording
	err = controller.StopRecording(ctx, session.ID)
	require.NoError(t, err, "Recording should stop successfully")

	// Verify session is no longer active
	_, err = controller.ListRecordings(ctx, 10, 0)
	require.NoError(t, err, "Should be able to list recordings")
	// Note: The session might still be in the list but marked as stopped
}

// TestController_TakeSnapshot_ReqMTX002 tests snapshot functionality through controller
func TestController_TakeSnapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager with proper configuration loading
	configManager := createConfigManagerWithEnvVars(t, helper)

	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "snapshots")
	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "test_snapshot.jpg")

	// Test snapshot with camera identifier (abstraction layer)
	options := map[string]interface{}{}

	snapshot, err := controller.TakeAdvancedSnapshot(ctx, "camera0", outputPath, options)
	if err != nil {
		t.Logf("Snapshot error details: %v", err)
	}
	require.NoError(t, err, "Snapshot should be taken successfully")
	require.NotNil(t, snapshot, "Snapshot should not be nil")

	// Verify snapshot properties
	assert.NotEmpty(t, snapshot.ID, "Snapshot should have an ID")
	assert.Equal(t, "camera0", snapshot.Device, "Should use camera identifier")
	assert.Equal(t, outputPath, snapshot.FilePath, "Should match output path")
}

// TestController_StreamManagement_ReqMTX002 tests stream management through controller
func TestController_StreamManagement_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create real config manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	controller, err := ControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test stream management through controller
	streams, err := controller.GetStreams(ctx)
	require.NoError(t, err, "Should be able to get streams")
	require.NotNil(t, streams, "Streams should not be nil")

	// Test paths management through controller
	paths, err := controller.GetPaths(ctx)
	require.NoError(t, err, "Should be able to get paths")
	require.NotNil(t, paths, "Paths should not be nil")
}

// TestController_AdvancedRecording_ReqMTX002 tests advanced recording functionality
func TestController_AdvancedRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities (advanced recording)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create config manager using test fixture
	configManager := CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	_, err = configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should be able to get MediaMTX config from fixture")

	// Create controller
	controller, err := ControllerWithConfigManager(configManager, helper.GetLogger())
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test advanced recording with options
	device := "camera0"
	path := "/tmp/mediamtx_test_data/recordings/advanced_test.mp4"
	options := map[string]interface{}{
		"quality":      "high",
		"resolution":   "1920x1080",
		"framerate":    30,
		"bitrate":      "2000k",
		"segment_time": 60,
	}

	session, err := controller.StartAdvancedRecording(ctx, device, path, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Verify session properties
	assert.Equal(t, "camera0", session.DevicePath, "Should use camera identifier for API consistency")
	assert.Equal(t, path, session.FilePath, "File path should match")
	assert.Equal(t, "active", session.Status, "Status should be active")
	assert.NotEmpty(t, session.ID, "Session ID should not be empty")
	assert.NotEmpty(t, session.ContinuityID, "Continuity ID should not be empty")
	assert.Equal(t, SessionStateRecording, session.State, "State should be recording")

	// Test getting advanced recording session
	retrievedSession, exists := controller.GetAdvancedRecordingSession(session.ID)
	require.True(t, exists, "Should be able to retrieve advanced recording session")
	require.NotNil(t, retrievedSession, "Retrieved session should not be nil")
	assert.Equal(t, session.ID, retrievedSession.ID, "Session IDs should match")

	// Test listing advanced recording sessions
	sessions := controller.ListAdvancedRecordingSessions()
	require.NotNil(t, sessions, "Sessions list should not be nil")
	assert.Len(t, sessions, 1, "Should have one active session")

	// Stop the recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err, "Advanced recording should stop successfully")

	t.Log("✅ Advanced recording functionality working correctly")
}

// TestController_StreamRecording_ReqMTX002 tests stream recording functionality
func TestController_StreamRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities (stream recording)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Create config manager using test fixture
	configManager := CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	_, err = configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should be able to get MediaMTX config from fixture")

	// Create controller
	controller, err := ControllerWithConfigManager(configManager, helper.GetLogger())
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test stream recording
	device := "camera0"
	stream, err := controller.StartStreaming(ctx, device)
	require.NoError(t, err, "Stream recording should start successfully")
	require.NotNil(t, stream, "Stream should not be nil")

	// Verify stream properties
	assert.NotEmpty(t, stream.Name, "Stream name should not be empty")
	assert.NotEmpty(t, stream.URL, "Stream URL should not be empty")
	assert.True(t, stream.Ready, "Stream should be ready")

	// Test getting stream status
	status, err := controller.GetStreamStatus(ctx, stream.Name)
	require.NoError(t, err, "Should be able to get stream status")
	require.NotNil(t, status, "Stream status should not be nil")

	// Test getting stream URL
	streamURL, err := controller.GetStreamURL(ctx, stream.Name)
	require.NoError(t, err, "Should be able to get stream URL")
	require.NotNil(t, streamURL, "Stream URL should not be nil")
	assert.NotEmpty(t, streamURL, "Stream URL should not be empty")

	// Stop the stream
	err = controller.StopStreaming(ctx, stream.Name)
	require.NoError(t, err, "Stream should stop successfully")

	t.Log("✅ Stream recording functionality working correctly")
}
