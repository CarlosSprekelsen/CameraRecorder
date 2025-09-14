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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_StartAdvancedRecording_ReqMTX002 tests advanced recording with options
func TestController_StartAdvancedRecording_ReqMTX002(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Just get controller - no orchestration
	controller, err := helper.GetController(t)
	require.NoError(t, err)
	require.NotNil(t, controller)

	// Just start it - no waiting
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Setup test - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
	options := map[string]interface{}{
		"quality":    "high",
		"resolution": "1920x1080",
		"framerate":  30,
		"bitrate":    "5000k",
		"codec":      "h264",
		"audio":      true,
		"duration":   10,
	}

	// Try to record with simple retry
	var session *RecordingSession
	for i := 0; i < 3; i++ {
		session, err = controller.StartAdvancedRecording(ctx, device, options)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "not ready") {
			time.Sleep(time.Second)
			continue
		}
		require.NoError(t, err)
	}

	require.NotNil(t, session)

	// Verify it works
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, device, session.DevicePath)
	assert.Equal(t, "active", session.Status)

	// Stop recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err)
}

// TestController_StopAdvancedRecording_ReqMTX002 tests stopping advanced recording
func TestController_StopAdvancedRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create test output directory using configured path
	outputDir := filepath.Join(helper.GetConfiguredRecordingPath(), "test_stop_advanced_recordings")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
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
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create test output directory using configured path
	outputDir := filepath.Join(helper.GetConfiguredRecordingPath(), "test_get_advanced_recordings")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
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
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create test output directory using configured path
	outputDir := filepath.Join(helper.GetConfiguredRecordingPath(), "test_list_advanced_recordings")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Get initial session count
	initialSessions := controller.ListAdvancedRecordingSessions()
	initialCount := len(initialSessions)

	// Start multiple advanced recordings - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
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
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create test output directory using configured path
	outputDir := filepath.Join(helper.GetConfiguredRecordingPath(), "test_rotate_advanced_recordings")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err, "Creating output directory should succeed")
	defer os.RemoveAll(outputDir)

	// Start advanced recording - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
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
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
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

// TestController_EventDrivenAdvancedRecording_ReqMTX002 tests event-driven advanced recording
func TestController_EventDrivenAdvancedRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities with event-driven approach
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create event-driven test helper
	eventHelper := helper.CreateEventDrivenTestHelper(t)
	defer eventHelper.Cleanup()

	// Test event-driven recording with optimized timeouts - get available device instead of hardcoded camera0
	cameraList, err := controller.GetCameraList(ctx)
	require.NoError(t, err, "Should be able to get camera list")
	require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")

	device := cameraList.Cameras[0].Device // Use first available camera
	options := map[string]interface{}{
		"quality":    "high",
		"resolution": "1920x1080",
		"framerate":  30,
		"bitrate":    "5000k",
		"codec":      "h264",
		"audio":      true,
		"duration":   5, // 5 seconds
	}

	// Start recording with event-driven readiness check
	recordingCtx, recordingCancel := context.WithTimeout(ctx, 15*time.Second)
	defer recordingCancel()

	session, err := controller.StartAdvancedRecording(recordingCtx, device, options)
	require.NoError(t, err, "Advanced recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Use event-driven approach to wait for recording readiness
	// Controller started, no need to wait for readiness

	// Verify session properties
	assert.NotEmpty(t, session.ID, "Session should have an ID")
	assert.Equal(t, device, session.DevicePath, "Device path should match")
	assert.NotEmpty(t, session.FilePath, "File path should be generated")
	assert.Equal(t, "active", session.Status, "Session should be active")
	assert.NotZero(t, session.StartTime, "Start time should be set")

	// Verify file creation with optimized timeout (TODO: Replace with event-driven file creation notifications)
	require.Eventually(t, func() bool {
		_, err := os.Stat(session.FilePath)
		return err == nil
	}, 3*time.Second, 50*time.Millisecond, "FFmpeg should create recording file within 3 seconds (optimized polling)")

	// Test non-blocking event observation for verification
	eventHelper.ObserveHealthChanges()
	eventHelper.ObserveReadiness()

	// No waiting - just verify events occurred after work is done
	// This follows the Progressive Readiness Pattern

	// Stop the recording
	err = controller.StopAdvancedRecording(ctx, session.ID)
	require.NoError(t, err, "Stopping advanced recording should succeed")

	// Verify output file was created
	_, err = os.Stat(session.FilePath)
	assert.NoError(t, err, "Output file should be created")
}
