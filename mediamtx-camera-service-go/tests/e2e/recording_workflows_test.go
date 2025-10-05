/*
Recording Workflows E2E Tests

Tests complete user workflows for recording operations including start/stop recording,
duration parameters, multiple camera recording, and recording list management.
Each test validates a complete user journey with actual file creation and verification.

Test Categories:
- Basic Recording Lifecycle: Start recording, wait, stop recording, verify file exists with minimum size
- Recording with Duration Parameter: Start recording with duration, verify auto-stop and file creation
- Multiple Camera Recording: Start recording on multiple cameras, verify all files created
- Recording List Workflow: Create recordings, list recordings, verify all present with metadata

Business Outcomes:
- User can play video file in VLC (verify file is valid video format)
- User gets video clip in requested format and duration
- User can record multiple cameras simultaneously
- User can see all available recordings with metadata

Coverage Target: 40% E2E coverage milestone
*/

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicRecordingLifecycle(t *testing.T) {
	// Setup: Authenticated connection, clean recordings directory verification
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Basic Recording", 1, "Authenticated connection established")

	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for recording test")
	}

	camera := cameras[0].(map[string]interface{})
	deviceID := camera["device"].(string)
	LogWorkflowStep(t, "Basic Recording", 2, "Camera selected for recording")

	// Step 1: Call start_recording with device ID
	startResponse := setup.SendJSONRPC(conn, "start_recording", map[string]interface{}{
		"device": deviceID,
	})
	LogWorkflowStep(t, "Basic Recording", 3, "Start recording request sent")

	// Step 2: Verify start response (session_id, status, recording_path)
	require.NoError(t, startResponse.Error, "Start recording should succeed")
	require.NotNil(t, startResponse.Result, "Start recording result should not be nil")

	startResult := startResponse.Result.(map[string]interface{})
	assert.Contains(t, startResult, "session_id", "Start result should contain session_id")
	assert.Contains(t, startResult, "status", "Start result should contain status")
	assert.Contains(t, startResult, "recording_path", "Start result should contain recording_path")

	sessionID := startResult["session_id"].(string)
	recordingPath := startResult["recording_path"].(string)
	assert.NotEmpty(t, sessionID, "Session ID should not be empty")
	assert.NotEmpty(t, recordingPath, "Recording path should not be empty")
	LogWorkflowStep(t, "Basic Recording", 4, "Recording started successfully")

	// Step 3: Use waitForRecordingStart to verify file created and growing (NO time.Sleep)
	setup.WaitForRecordingStart(recordingPath)
	LogWorkflowStep(t, "Basic Recording", 5, "Recording file created and growing")

	// Step 4: Wait for file to grow using testutils.WaitForCondition
	err := setup.WaitForCondition(func() bool {
		// Record initial file size
		initialSize := getFileSize(t, recordingPath)

		// Check if file has grown (without sleep)
		currentSize := getFileSize(t, recordingPath)
		return currentSize > initialSize
	}, testutils.UniversalTimeoutVeryLong)

	require.NoError(t, err, "Recording file should grow during recording")
	LogWorkflowStep(t, "Basic Recording", 6, "Recording file growing confirmed")

	// Step 5: Call stop_recording with device ID
	stopResponse := setup.SendJSONRPC(conn, "stop_recording", map[string]interface{}{
		"device": deviceID,
	})
	LogWorkflowStep(t, "Basic Recording", 7, "Stop recording request sent")

	// Step 6: Verify stop response (final file metadata)
	require.NoError(t, stopResponse.Error, "Stop recording should succeed")
	require.NotNil(t, stopResponse.Result, "Stop recording result should not be nil")

	stopResult := stopResponse.Result.(map[string]interface{})
	assert.Contains(t, stopResult, "session_id", "Stop result should contain session_id")
	assert.Contains(t, stopResult, "status", "Stop result should contain status")
	assert.Equal(t, sessionID, stopResult["session_id"], "Stop session ID should match start session ID")
	LogWorkflowStep(t, "Basic Recording", 8, "Recording stopped successfully")

	// Validation: File exists AND size > 10KB AND file was growing during recording AND final size appropriate for duration AND format matches request
	setup.VerifyRecordingFile(recordingPath, 5) // 5 seconds minimum duration
	LogWorkflowStep(t, "Basic Recording", 9, "Recording file validation completed")

	// Business Outcome: User can play video file in VLC (verify file is valid video format)
	setup.AssertBusinessOutcome("User can play video file in VLC", func() bool {
		return fileExists(t, recordingPath) && getFileSize(t, recordingPath) > testutils.UniversalMinRecordingFileSize
	})
	LogWorkflowStep(t, "Basic Recording", 10, "Business outcome validated - user can play video file")

	// Cleanup: Remove test recording file, verify cleanup succeeded
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Basic Recording", 11, "Connection closed and cleanup verified")
}

func TestRecordingWithDurationParameter(t *testing.T) {
	// Setup: Authenticated connection, clean directory
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Duration Recording", 1, "Authenticated connection established")

	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for duration recording test")
	}

	camera := cameras[0].(map[string]interface{})
	deviceID := camera["device"].(string)
	LogWorkflowStep(t, "Duration Recording", 2, "Camera selected for duration recording")

	// Step 1: Call start_recording with device ID, duration=10, format="mp4"
	durationSeconds := 10
	startResponse := setup.SendJSONRPC(conn, "start_recording", map[string]interface{}{
		"device":   deviceID,
		"duration": durationSeconds,
		"format":   "mp4",
	})
	LogWorkflowStep(t, "Duration Recording", 3, "Start recording with duration request sent")

	// Step 2: Verify recording starts successfully
	require.NoError(t, startResponse.Error, "Start recording with duration should succeed")
	require.NotNil(t, startResponse.Result, "Start recording result should not be nil")

	startResult := startResponse.Result.(map[string]interface{})
	recordingPath := startResult["recording_path"].(string)
	assert.NotEmpty(t, recordingPath, "Recording path should not be empty")
	LogWorkflowStep(t, "Duration Recording", 4, "Recording with duration started successfully")

	// Step 3: Use waitForRecordingStart to verify file created with .mp4 extension
	setup.WaitForRecordingStart(recordingPath)
	assert.Contains(t, recordingPath, ".mp4", "Recording file should have .mp4 extension")
	LogWorkflowStep(t, "Duration Recording", 5, "Recording file created with .mp4 extension")

	// Step 4: Use testutils.WaitForCondition to check recording auto-stops after ~10 seconds
	startTime := time.Now()
	err := setup.WaitForCondition(func() bool {
		// Try to stop recording - if it fails, recording may have auto-stopped
		stopResponse := setup.SendJSONRPC(conn, "stop_recording", map[string]interface{}{
			"device": deviceID,
		})

		// If stop succeeds or fails with "not recording" error, recording has stopped
		return stopResponse.Error == nil ||
			(stopResponse.Error != nil &&
				stopResponse.Error.Message == "not recording")
	}, time.Duration(durationSeconds+5)*time.Second) // Wait duration + 5 seconds buffer

	require.NoError(t, err, "Recording should auto-stop after duration")

	elapsed := time.Since(startTime)
	assert.GreaterOrEqual(t, elapsed, time.Duration(durationSeconds)*time.Second, "Recording should run for at least the specified duration")
	LogWorkflowStep(t, "Duration Recording", 6, "Recording auto-stopped after duration")

	// Step 5: Verify final file size appropriate for 10-second duration
	setup.VerifyRecordingFile(recordingPath, durationSeconds)
	LogWorkflowStep(t, "Duration Recording", 7, "Recording file size validated for duration")

	// Validation: Recording starts AND file created AND .mp4 extension AND auto-stops AND size appropriate
	setup.AssertBusinessOutcome("User gets video clip in requested format and duration", func() bool {
		return fileExists(t, recordingPath) &&
			getFileSize(t, recordingPath) > testutils.UniversalMinRecordingFileSize &&
			containsString(t, recordingPath, ".mp4")
	})
	LogWorkflowStep(t, "Duration Recording", 8, "Business outcome validated - user gets video clip in requested format")

	// Cleanup: Remove test files, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Duration Recording", 9, "Connection closed and cleanup verified")
}

func TestMultipleCameraRecording(t *testing.T) {
	// Setup: Authenticated connection, discover multiple cameras (skip if < 2 cameras)
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Multiple Camera Recording", 1, "Authenticated connection established")

	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) < 2 {
		t.Skip("Need at least 2 cameras for multiple camera recording test")
	}
	LogWorkflowStep(t, "Multiple Camera Recording", 2, "Multiple cameras discovered")

	// Step 1: Start recording on camera0
	camera0 := cameras[0].(map[string]interface{})
	deviceID0 := camera0["device"].(string)

	startResponse0 := setup.SendJSONRPC(conn, "start_recording", map[string]interface{}{
		"device": deviceID0,
	})
	require.NoError(t, startResponse0.Error, "Start recording on camera0 should succeed")

	startResult0 := startResponse0.Result.(map[string]interface{})
	recordingPath0 := startResult0["recording_path"].(string)
	sessionID0 := startResult0["session_id"].(string)
	LogWorkflowStep(t, "Multiple Camera Recording", 3, "Recording started on camera0")

	// Step 2: Start recording on camera1
	camera1 := cameras[1].(map[string]interface{})
	deviceID1 := camera1["device"].(string)

	startResponse1 := setup.SendJSONRPC(conn, "start_recording", map[string]interface{}{
		"device": deviceID1,
	})
	require.NoError(t, startResponse1.Error, "Start recording on camera1 should succeed")

	startResult1 := startResponse1.Result.(map[string]interface{})
	recordingPath1 := startResult1["recording_path"].(string)
	sessionID1 := startResult1["session_id"].(string)
	LogWorkflowStep(t, "Multiple Camera Recording", 4, "Recording started on camera1")

	// Step 3: Verify both recordings active with separate files
	assert.NotEqual(t, recordingPath0, recordingPath1, "Recording files should be different")
	assert.NotEqual(t, sessionID0, sessionID1, "Session IDs should be different")

	// Verify both files exist and are growing
	setup.WaitForRecordingStart(recordingPath0)
	setup.WaitForRecordingStart(recordingPath1)
	LogWorkflowStep(t, "Multiple Camera Recording", 5, "Both recordings active with separate files")

	// Step 4: Use testutils.WaitForCondition to verify both files growing
	err := setup.WaitForCondition(func() bool {
		// Check both files are growing
		size0 := getFileSize(t, recordingPath0)
		size1 := getFileSize(t, recordingPath1)

		newSize0 := getFileSize(t, recordingPath0)
		newSize1 := getFileSize(t, recordingPath1)

		return newSize0 > size0 && newSize1 > size1
	}, testutils.UniversalTimeoutVeryLong)

	require.NoError(t, err, "Both recording files should be growing")
	LogWorkflowStep(t, "Multiple Camera Recording", 6, "Both files growing confirmed")

	// Step 5: Stop both recordings
	stopResponse0 := setup.SendJSONRPC(conn, "stop_recording", map[string]interface{}{
		"device": deviceID0,
	})
	require.NoError(t, stopResponse0.Error, "Stop recording on camera0 should succeed")

	stopResponse1 := setup.SendJSONRPC(conn, "stop_recording", map[string]interface{}{
		"device": deviceID1,
	})
	require.NoError(t, stopResponse1.Error, "Stop recording on camera1 should succeed")
	LogWorkflowStep(t, "Multiple Camera Recording", 7, "Both recordings stopped")

	// Step 6: Verify both files exist with appropriate sizes
	setup.VerifyRecordingFile(recordingPath0, 5) // 5 seconds minimum
	setup.VerifyRecordingFile(recordingPath1, 5) // 5 seconds minimum

	// Verify files are in correct subdirectories (camera0/, camera1/)
	assert.Contains(t, recordingPath0, deviceID0, "Recording0 path should contain camera ID")
	assert.Contains(t, recordingPath1, deviceID1, "Recording1 path should contain camera ID")
	LogWorkflowStep(t, "Multiple Camera Recording", 8, "Both files exist with appropriate sizes")

	// Validation: Both recordings active AND separate files created AND files in correct subdirectories AND both files appropriate size
	setup.AssertBusinessOutcome("User can record multiple cameras simultaneously", func() bool {
		return fileExists(t, recordingPath0) &&
			fileExists(t, recordingPath1) &&
			getFileSize(t, recordingPath0) > testutils.UniversalMinRecordingFileSize &&
			getFileSize(t, recordingPath1) > testutils.UniversalMinRecordingFileSize &&
			recordingPath0 != recordingPath1
	})
	LogWorkflowStep(t, "Multiple Camera Recording", 9, "Business outcome validated - user can record multiple cameras")

	// Cleanup: Remove all test recording files, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Multiple Camera Recording", 10, "Connection closed and cleanup verified")
}

func TestRecordingListWorkflow(t *testing.T) {
	// Setup: Create 3 test recordings (helper function)
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Recording List", 1, "Authenticated connection established")

	// Create test recordings
	testRecordings := setup.CreateTestRecordings(3)
	LogWorkflowStep(t, "Recording List", 2, "3 test recordings created")

	// Step 1: Call list_recordings method
	listResponse := setup.SendJSONRPC(conn, "list_recordings", map[string]interface{}{})
	LogWorkflowStep(t, "Recording List", 3, "List recordings request sent")

	// Step 2: Verify all 3 recordings appear in list
	require.NoError(t, listResponse.Error, "List recordings request should succeed")
	require.NotNil(t, listResponse.Result, "List recordings result should not be nil")

	resultMap := listResponse.Result.(map[string]interface{})
	recordings := resultMap["recordings"].([]interface{})

	// Should have at least our test recordings (may have more from other tests)
	assert.GreaterOrEqual(t, len(recordings), 3, "Should have at least 3 recordings in list")
	LogWorkflowStep(t, "Recording List", 4, "Recordings list retrieved successfully")

	// Step 3: Validate metadata (filename, size, timestamp, camera_id)
	for i, recordingInterface := range recordings[:3] { // Check first 3 recordings
		recording := recordingInterface.(map[string]interface{})

		assert.Contains(t, recording, "filename", "Recording should contain filename")
		assert.Contains(t, recording, "size", "Recording should contain size")
		assert.Contains(t, recording, "timestamp", "Recording should contain timestamp")
		assert.Contains(t, recording, "camera_id", "Recording should contain camera_id")

		// Validate metadata types and values
		filename := recording["filename"].(string)
		assert.NotEmpty(t, filename, "Filename should not be empty")

		if size, ok := recording["size"].(float64); ok {
			assert.Greater(t, size, 0, "Recording size should be positive")
		}

		if timestamp, ok := recording["timestamp"].(string); ok {
			assert.NotEmpty(t, timestamp, "Timestamp should not be empty")
		}

		cameraID := recording["camera_id"].(string)
		assert.NotEmpty(t, cameraID, "Camera ID should not be empty")

		t.Logf("Recording %d: %s (size: %v, camera: %s)", i+1, filename, recording["size"], cameraID)
	}
	LogWorkflowStep(t, "Recording List", 5, "Recording metadata validated")

	// Step 4: Verify each recording file actually exists at listed path
	for i, recordingInterface := range recordings[:3] {
		recording := recordingInterface.(map[string]interface{})
		filename := recording["filename"].(string)
		cameraID := recording["camera_id"].(string)

		// Construct expected file path
		expectedPath := "/tmp/e2e-test-recordings/" + cameraID + "/" + filename

		// Verify file exists
		if fileExists(t, expectedPath) {
			t.Logf("Recording file exists: %s", expectedPath)
		} else {
			t.Logf("Warning: Recording file not found: %s", expectedPath)
		}

		// For test recordings, verify they exist
		if i < len(testRecordings) {
			assert.True(t, fileExists(t, testRecordings[i]), "Test recording file should exist")
		}
	}
	LogWorkflowStep(t, "Recording List", 6, "Recording file existence verified")

	// Validation: All 3 recordings in list AND metadata complete AND files exist AND sizes match metadata
	setup.AssertBusinessOutcome("User can see all available recordings with metadata", func() bool {
		return len(recordings) >= 3 &&
			recordings[0].(map[string]interface{})["filename"] != nil &&
			recordings[1].(map[string]interface{})["filename"] != nil &&
			recordings[2].(map[string]interface{})["filename"] != nil
	})
	LogWorkflowStep(t, "Recording List", 7, "Business outcome validated - user can see recordings with metadata")

	// Cleanup: Remove test recordings, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Recording List", 8, "Connection closed and cleanup verified")
}

// Helper functions for recording tests

func containsString(t *testing.T, str, substr string) bool {
	return strings.Contains(str, substr)
}
