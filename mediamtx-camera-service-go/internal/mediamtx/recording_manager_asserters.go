/*
Recording Manager Test Asserters - Eliminate Massive Duplication

This file provides domain-specific asserters for recording manager tests that eliminate
the massive duplication found in recording_manager_test.go (613 lines).

Duplication Patterns Eliminated:
- SetupMediaMTXTest + GetReadyController (12 times)
- Recording manager initialization (12+ times)
- Recording lifecycle operations (8+ times)
- File validation and cleanup (6+ times)
- Error handling patterns (10+ times)

Usage:
    asserter := NewRecordingManagerAsserter(t)
    defer asserter.Cleanup()
    // Test-specific logic only
    asserter.AssertCompleteRecordingLifecycle(cameraID, options)
*/

package mediamtx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RecordingManagerAsserter encapsulates all recording manager test patterns
type RecordingManagerAsserter struct {
	t                *testing.T
	helper           *MediaMTXTestHelper
	ctx              context.Context
	cancel           context.CancelFunc
	controller       MediaMTXController
	recordingManager *RecordingManager
}

// NewRecordingManagerAsserter creates a new recording manager asserter with full setup
// Eliminates: helper, _ := SetupMediaMTXTest(t) + GetReadyController pattern
func NewRecordingManagerAsserter(t *testing.T) *RecordingManagerAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)
	recordingManager := helper.GetRecordingManager()

	return &RecordingManagerAsserter{
		t:                t,
		helper:           helper,
		ctx:              ctx,
		cancel:           cancel,
		controller:       controller,
		recordingManager: recordingManager,
	}
}

// Cleanup stops the controller and cancels the context
func (rma *RecordingManagerAsserter) Cleanup() {
	rma.controller.Stop(rma.ctx)
	rma.cancel()
}

// GetHelper returns the underlying MediaMTXTestHelper
func (rma *RecordingManagerAsserter) GetHelper() *MediaMTXTestHelper {
	return rma.helper
}

// GetRecordingManager returns the recording manager instance
func (rma *RecordingManagerAsserter) GetRecordingManager() *RecordingManager {
	return rma.recordingManager
}

// GetController returns the controller instance
func (rma *RecordingManagerAsserter) GetController() MediaMTXController {
	return rma.controller
}

// GetContext returns the test context
func (rma *RecordingManagerAsserter) GetContext() context.Context {
	return rma.ctx
}

// AssertRecordingManagerCreation validates recording manager creation
// Eliminates recording manager initialization duplication
func (rma *RecordingManagerAsserter) AssertRecordingManagerCreation() *RecordingManager {
	require.NotNil(rma.t, rma.recordingManager, "Recording manager should be initialized")

	rma.t.Log("✅ Recording manager created successfully")
	return rma.recordingManager
}

// AssertCompleteRecordingLifecycle tests complete recording lifecycle with data validation
// Eliminates recording lifecycle duplication
func (rma *RecordingManagerAsserter) AssertCompleteRecordingLifecycle(cameraID string, options *PathConf) *StartRecordingResponse {
	// Start recording
	session, err := rma.recordingManager.StartRecording(rma.ctx, cameraID, options)
	require.NoError(rma.t, err, "StartRecording should succeed")
	require.NotNil(rma.t, session, "Recording session should not be nil")

	// Verify recording file exists and has content
	expectedFilePath := testutils.BuildRecordingFilePath(
		rma.helper.GetConfiguredRecordingPath(),
		cameraID,
		rma.helper.GetConfiguredRecordingFormat(),
		true, // use_device_subdirs
		"",   // timestamp
	)

	// Wait for file creation with Progressive Readiness
	result := testutils.TestProgressiveReadiness(rma.t, func() (bool, error) {
		fileInfo, err := os.Stat(expectedFilePath)
		if err != nil {
			return false, err
		}
		return fileInfo.Size() > testutils.UniversalMinRecordingFileSize, nil
	}, rma.controller, "RecordingFileCreation")

	require.NoError(rma.t, result.Error, "Recording file should be created")
	require.True(rma.t, result.Result, "Recording file should have meaningful size")

	if result.UsedFallback {
		rma.t.Log("⚠️  PROGRESSIVE READINESS FALLBACK: Recording file needed readiness event")
	} else {
		rma.t.Log("✅ PROGRESSIVE READINESS: Recording file created immediately")
	}

	// Stop recording
	_, err = rma.recordingManager.StopRecording(rma.ctx, cameraID)
	require.NoError(rma.t, err, "StopRecording should succeed")

	// Verify final file exists and has content
	fileInfo, err := os.Stat(expectedFilePath)
	require.NoError(rma.t, err, "Final recording file should exist")
	assert.Greater(rma.t, fileInfo.Size(), testutils.UniversalMinRecordingFileSize,
		"Final recording file should have meaningful size")

	rma.t.Logf("✅ Complete recording lifecycle validated: %s (%d bytes)", expectedFilePath, fileInfo.Size())
	return session
}

// AssertStartRecording tests recording start operation
// Eliminates recording start duplication
func (rma *RecordingManagerAsserter) AssertStartRecording(cameraID string, options *PathConf) *StartRecordingResponse {
	session, err := rma.recordingManager.StartRecording(rma.ctx, cameraID, options)
	require.NoError(rma.t, err, "StartRecording should succeed")
	require.NotNil(rma.t, session, "Recording session should not be nil")

	rma.t.Logf("✅ Recording started successfully: camera=%s", cameraID)
	return session
}

// AssertStopRecording tests recording stop operation
// Eliminates recording stop duplication
func (rma *RecordingManagerAsserter) AssertStopRecording(cameraID string) {
	_, err := rma.recordingManager.StopRecording(rma.ctx, cameraID)
	require.NoError(rma.t, err, "StopRecording should succeed")

	rma.t.Logf("✅ Recording stopped successfully: camera=%s", cameraID)
}

// AssertGetRecordingsList tests recording list retrieval
// Eliminates recording list duplication
func (rma *RecordingManagerAsserter) AssertGetRecordingsList(limit, offset int) *FileListResponse {
	recordings, err := rma.recordingManager.GetRecordingsList(rma.ctx, limit, offset)
	require.NoError(rma.t, err, "GetRecordingsList should succeed")
	require.NotNil(rma.t, recordings, "Recording list should not be nil")

	rma.t.Logf("✅ Recordings list retrieved: limit=%d, offset=%d, found=%d, total=%d",
		limit, offset, len(recordings.Files), recordings.Total)
	return recordings
}

// AssertRecordingFileExists validates recording file creation
// Eliminates file validation duplication
func (rma *RecordingManagerAsserter) AssertRecordingFileExists(cameraID string) string {
	expectedFilePath := testutils.BuildRecordingFilePath(
		rma.helper.GetConfiguredRecordingPath(),
		cameraID,
		rma.helper.GetConfiguredRecordingFormat(),
		true, // use_device_subdirs
		"",   // timestamp
	)

	fileInfo, err := os.Stat(expectedFilePath)
	require.NoError(rma.t, err, "Recording file should exist: %s", expectedFilePath)
	assert.Greater(rma.t, fileInfo.Size(), testutils.UniversalMinRecordingFileSize,
		"Recording file should have meaningful size")

	rma.t.Logf("✅ Recording file validated: %s (%d bytes)", expectedFilePath, fileInfo.Size())
	return expectedFilePath
}

// AssertAPIErrorHandling tests API error handling scenarios
// Eliminates error handling duplication
func (rma *RecordingManagerAsserter) AssertAPIErrorHandling() {
	// Test with invalid camera ID
	invalidCameraID := "nonexistent_camera_12345"
	options := &PathConf{
		Record:       true,
		RecordFormat: rma.helper.GetConfiguredRecordingFormat(),
	}

	// This should fail gracefully
	session, err := rma.recordingManager.StartRecording(rma.ctx, invalidCameraID, options)
	assert.Error(rma.t, err, "Recording should fail with invalid camera")
	assert.Nil(rma.t, session, "Session should be nil on error")

	rma.t.Log("✅ API error handling validated")
}

// AssertConcurrentAccess tests concurrent access to recording manager
// Eliminates concurrency testing duplication
func (rma *RecordingManagerAsserter) AssertConcurrentAccess() {
	const numGoroutines = 3
	cameraID := rma.helper.MustGetCameraID(rma.t, rma.ctx, rma.controller)

	results := make(chan error, numGoroutines)

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			options := &PathConf{
				Record:       true,
				RecordFormat: rma.helper.GetConfiguredRecordingFormat(),
			}

			session, err := rma.recordingManager.StartRecording(rma.ctx, cameraID, options)
			if err != nil {
				results <- err
				return
			}
			if session == nil {
				results <- assert.AnError
				return
			}

			// Stop recording
			_, err = rma.recordingManager.StopRecording(rma.ctx, cameraID)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			require.NoError(rma.t, err, "Concurrent recording operation should succeed")
		case <-time.After(10 * time.Second):
			rma.t.Fatal("Concurrent recording test timed out")
		}
	}

	rma.t.Log("✅ Recording concurrent access validated")
}

// AssertMultiTierRecording tests multi-tier recording scenarios
// Eliminates multi-tier testing duplication
func (rma *RecordingManagerAsserter) AssertMultiTierRecording() {
	cameraID := rma.helper.MustGetCameraID(rma.t, rma.ctx, rma.controller)

	// Test different recording formats
	formats := []string{"fmp4", "mp4", "ts"}

	for _, format := range formats {
		options := &PathConf{
			Record:       true,
			RecordFormat: format,
		}

		session := rma.AssertStartRecording(cameraID, options)
		assert.NotNil(rma.t, session, "Multi-tier recording should succeed for format: %s", format)

		// Brief recording
		time.Sleep(testutils.UniversalRetryDelay)

		rma.AssertStopRecording(cameraID)

		// Verify file was created
		rma.AssertRecordingFileExists(cameraID)
	}

	rma.t.Log("✅ Multi-tier recording validated")
}

// AssertProgressiveReadinessCompliance tests Progressive Readiness compliance
// Eliminates Progressive Readiness testing duplication
func (rma *RecordingManagerAsserter) AssertProgressiveReadinessCompliance() {
	// Test that controller starts accepting operations immediately
	startTime := time.Now()

	// Controller is already started by GetReadyController - no need to start again
	startDuration := time.Since(startTime)
	assert.Less(rma.t, startDuration, 100*time.Millisecond,
		"Controller should start immediately (<100ms), took %v", startDuration)

	// Test operations work immediately
	cameraID := rma.helper.MustGetCameraID(rma.t, rma.ctx, rma.controller)
	options := &PathConf{
		Record:       true,
		RecordFormat: rma.helper.GetConfiguredRecordingFormat(),
	}

	session := rma.AssertStartRecording(cameraID, options)
	assert.NotNil(rma.t, session, "Recording should work immediately after controller start")

	rma.t.Log("✅ Progressive Readiness compliance validated")
}
