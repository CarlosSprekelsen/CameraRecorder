/*
MediaMTX Advanced Recording Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_StartAdvancedRecording_ReqMTX002 tests advanced recording with options
func TestController_StartAdvancedRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Create test output directory
	outputDir := "/tmp/test_advanced_recordings"
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Test advanced recording with options
	device := "camera0"
	outputPath := filepath.Join(outputDir, "advanced_test.mp4")
	options := map[string]interface{}{
		"quality":    "high",
		"resolution": "1920x1080",
		"framerate":  30,
		"bitrate":    "5000k",
		"codec":      "h264",
		"audio":      true,
		"duration":   10, // 10 seconds
	}

	session, err := controller.StartAdvancedRecording(ctx, device, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Verify session properties
	assert.NotEmpty(t, session.ID, "Session should have an ID")
	assert.Equal(t, device, session.DevicePath, "Device path should match")
	assert.Equal(t, outputPath, session.FilePath, "File path should match")
	assert.Equal(t, "active", session.Status, "Session should be active")
	assert.NotZero(t, session.StartTime, "Start time should be set")

	// Stop the recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err, "Stopping advanced recording should succeed")

	// Verify output file was created
	_, err = os.Stat(outputPath)
	assert.NoError(t, err, "Output file should be created")
}

// TestController_StopAdvancedRecording_ReqMTX002 tests stopping advanced recording
func TestController_StopAdvancedRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Create test output directory
	outputDir := "/tmp/test_stop_advanced_recordings"
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording
	device := "camera0"
	options := map[string]interface{}{
		"quality":    "medium",
		"resolution": "1280x720",
		"framerate":  25,
		"duration":   5, // 5 seconds
	}

	session, err := controller.StartAdvancedRecording(ctx, device, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Stop the recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err, "Stopping advanced recording should succeed")

	// Verify session is no longer active
	sessions := controller.ListAdvancedRecordingSessions()
	found := false
	for _, s := range sessions {
		if s.ID == session.ID {
			found = true
			assert.Equal(t, "stopped", s.Status, "Session should be stopped")
			break
		}
	}
	assert.True(t, found, "Session should be found in list")
}

// TestController_GetAdvancedRecordingSession_ReqMTX002 tests getting advanced recording session
func TestController_GetAdvancedRecordingSession_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Create test output directory
	outputDir := "/tmp/test_get_advanced_recordings"
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording
	device := "camera0"
	options := map[string]interface{}{
		"quality":    "low",
		"resolution": "640x480",
		"framerate":  15,
		"duration":   3, // 3 seconds
	}

	session, err := controller.StartAdvancedRecording(ctx, device, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Get the session by ID
	retrievedSession, exists := controller.GetAdvancedRecordingSession(session.ID)
	require.True(t, exists, "Session should exist")
	require.NotNil(t, retrievedSession, "Retrieved session should not be nil")

	// Verify session properties match
	assert.Equal(t, session.ID, retrievedSession.ID, "Session ID should match")
	assert.Equal(t, session.DevicePath, retrievedSession.DevicePath, "Device path should match")
	assert.Equal(t, session.FilePath, retrievedSession.FilePath, "File path should match")
	assert.Equal(t, session.Status, retrievedSession.Status, "Status should match")

	// Test getting non-existent session
	_, exists = controller.GetAdvancedRecordingSession("non-existent-id")
	assert.False(t, exists, "Non-existent session should not exist")
}

// TestController_ListAdvancedRecordingSessions_ReqMTX002 tests listing advanced recording sessions
func TestController_ListAdvancedRecordingSessions_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Create test output directory
	outputDir := "/tmp/test_list_advanced_recordings"
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Get initial session count
	initialSessions := controller.ListAdvancedRecordingSessions()
	initialCount := len(initialSessions)

	// Start multiple advanced recordings
	device := "camera0"
	sessionIDs := make([]string, 3)

	for i := 0; i < 3; i++ {
		options := map[string]interface{}{
			"quality":    "medium",
			"resolution": "1280x720",
			"framerate":  25,
			"duration":   2, // 2 seconds
		}

		session, err := controller.StartAdvancedRecording(ctx, device, options)
		require.NoError(t, err, "Advanced recording should start successfully")
		require.NotNil(t, session, "Recording session should not be nil")
		sessionIDs[i] = session.ID
	}

	// List sessions and verify count increased
	sessions := controller.ListAdvancedRecordingSessions()
	assert.Equal(t, initialCount+3, len(sessions), "Should have 3 more sessions")

	// Verify all our sessions are in the list
	for _, sessionID := range sessionIDs {
		found := false
		for _, session := range sessions {
			if session.ID == sessionID {
				found = true
				assert.Equal(t, "active", session.Status, "Session should be active")
				break
			}
		}
		assert.True(t, found, "Session should be found in list")
	}
}

// TestController_RotateRecordingFile_ReqMTX002 tests rotating recording files
func TestController_RotateRecordingFile_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Create test output directory
	outputDir := "/tmp/test_rotate_advanced_recordings"
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording
	device := "camera0"
	outputPath := filepath.Join(outputDir, "rotate_test.mp4")
	options := map[string]interface{}{
		"quality":    "medium",
		"resolution": "1280x720",
		"framerate":  25,
		"duration":   10, // 10 seconds
	}

	session, err := controller.StartAdvancedRecording(ctx, device, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Rotate the recording file
	err = controller.RotateRecordingFile(ctx, session.ID)
	require.NoError(t, err, "Rotating recording file should succeed")

	// Verify original file exists
	_, err = os.Stat(outputPath)
	assert.NoError(t, err, "Original file should exist")

	// Verify rotated file exists (should have a timestamp or sequence number)
	// The exact naming convention depends on implementation
	files, err := filepath.Glob(filepath.Join(outputDir, "rotate_test*.mp4"))
	require.NoError(t, err, "Globbing files should succeed")
	assert.GreaterOrEqual(t, len(files), 1, "Should have at least one rotated file")

	// Stop the recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err, "Stopping advanced recording should succeed")
}

// TestController_AdvancedRecording_ErrorHandling_ReqMTX004 tests error handling for advanced recording
func TestController_AdvancedRecording_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
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

	// Test starting recording with invalid device
	invalidDevice := "camera999"
	options := map[string]interface{}{
		"quality": "high",
	}

	_, err = controller.StartAdvancedRecording(ctx, invalidDevice, options)
	assert.Error(t, err, "Starting recording with invalid device should fail")

	// Test stopping non-existent recording
	err = controller.StopAdvancedRecording(ctx, "non-existent-session-id")
	assert.Error(t, err, "Stopping non-existent recording should fail")

	// Test rotating non-existent recording
	err = controller.RotateRecordingFile(ctx, "non-existent-session-id")
	assert.Error(t, err, "Rotating non-existent recording should fail")
}
