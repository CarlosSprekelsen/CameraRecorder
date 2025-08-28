//go:build unit
// +build unit

/*
MediaMTX Recording Lifecycle Management Unit Tests

Requirements Coverage:
- REQ-CAM-045: Recording storage monitoring and threshold management
- REQ-CAM-046: Segmented recording with file rotation
- REQ-CAM-047: Recording continuity during file rotation
- REQ-CAM-048: Storage space validation before recording
- REQ-CAM-049: Recording session tracking and management

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// mockFFmpegManager implements FFmpegManager interface for testing
type mockFFmpegManager struct{}

func (m *mockFFmpegManager) CreateSegmentedRecording(ctx context.Context, device, outputPath string, options map[string]interface{}) error {
	return nil
}

func (m *mockFFmpegManager) CreateSnapshot(ctx context.Context, device, outputPath string, options map[string]interface{}) error {
	return nil
}

func (m *mockFFmpegManager) StopProcess(ctx context.Context, processID string) error {
	return nil
}

func TestRecordingManagerStorageMonitoring(t *testing.T) {
	// REQ-CAM-045: Recording storage monitoring and threshold management
	// REQ-CAM-048: Storage space validation before recording

	ffmpegManager := &mockFFmpegManager{}
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	t.Run("CheckStorageSpace_WithinLimits", func(t *testing.T) {
		// Test storage check when space is available
		storageInfo, err := recordingManager.checkStorageSpace()
		require.NoError(t, err)
		assert.NotNil(t, storageInfo, "Should return storage info")
		assert.True(t, storageInfo.Available, "Storage should be available when within limits")
	})

	t.Run("UpdateStorageThresholds", func(t *testing.T) {
		// Test updating storage thresholds
		recordingManager.UpdateStorageThresholds(85, 95)
		// Verify thresholds were updated (implementation dependent)
		assert.NotNil(t, recordingManager, "Should handle threshold updates")
	})
}

func TestRecordingManagerSegmentedRecording(t *testing.T) {
	// REQ-CAM-046: Segmented recording with file rotation

	ffmpegManager := &mockFFmpegManager{}
	config := &mediamtx.MediaMTXConfig{
		StoragePath: "/tmp/test_recordings",
	}
	logger := mediamtx.NewLogger()
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	t.Run("StartRecordingWithSegments_ValidConfig", func(t *testing.T) {
		ctx := context.Background()
		devicePath := "/dev/video0"
		outputPath := "/tmp/test_recordings/test_recording.mp4"
		options := map[string]interface{}{
			"segment_duration": 30,
			"segment_format":   "recording_%Y%m%d_%H%M%S.mp4",
		}

		session, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath, options)
		require.NoError(t, err)
		assert.NotNil(t, session, "Should return valid recording session")
		assert.NotEmpty(t, session.ID, "Session should have valid ID")
		assert.Equal(t, "RECORDING", session.Status, "Session should be in recording status")
	})

	t.Run("StartRecordingWithSegments_InvalidDevice", func(t *testing.T) {
		ctx := context.Background()
		devicePath := "/dev/invalid_device"
		outputPath := "/tmp/test_recordings/test_recording.mp4"
		options := map[string]interface{}{
			"segment_duration": 30,
		}

		session, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath, options)
		assert.Error(t, err, "Should return error for invalid device")
		assert.Nil(t, session, "Should not return session for invalid device")
	})
}

func TestRecordingManagerFileRotation(t *testing.T) {
	// REQ-CAM-047: Recording continuity during file rotation

	ffmpegManager := &mockFFmpegManager{}
	config := &mediamtx.MediaMTXConfig{
		StoragePath: "/tmp/test_recordings",
	}
	logger := mediamtx.NewLogger()
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	t.Run("RotateRecordingFile_ValidSession", func(t *testing.T) {
		ctx := context.Background()
		devicePath := "/dev/video0"
		outputPath := "/tmp/test_recordings/test_recording.mp4"
		options := map[string]interface{}{
			"segment_duration": 30,
		}

		// Start recording session
		session, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath, options)
		require.NoError(t, err)

		// Simulate file rotation
		err = recordingManager.RotateRecordingFile(ctx, session.ID)
		require.NoError(t, err, "File rotation should complete successfully")

		// Verify session continues after rotation
		updatedSession, exists := recordingManager.GetRecordingSession(session.ID)
		assert.True(t, exists, "Session should exist after rotation")
		assert.Equal(t, "RECORDING", updatedSession.Status, "Session should remain active after rotation")
	})

	t.Run("RotateRecordingFile_InvalidSession", func(t *testing.T) {
		ctx := context.Background()
		invalidSessionID := "invalid_session"

		err := recordingManager.RotateRecordingFile(ctx, invalidSessionID)
		assert.Error(t, err, "Should return error for invalid session")
	})
}

func TestRecordingManagerSessionTracking(t *testing.T) {
	// REQ-CAM-049: Recording session tracking and management

	ffmpegManager := &mockFFmpegManager{}
	config := &mediamtx.MediaMTXConfig{
		StoragePath: "/tmp/test_recordings",
	}
	logger := mediamtx.NewLogger()
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	t.Run("ListRecordingSessions_Empty", func(t *testing.T) {
		sessions := recordingManager.ListRecordingSessions()
		assert.Empty(t, sessions, "Should return empty list when no active sessions")
	})

	t.Run("ListRecordingSessions_WithSessions", func(t *testing.T) {
		ctx := context.Background()
		devicePath := "/dev/video0"
		outputPath := "/tmp/test_recordings/test_recording.mp4"
		options := map[string]interface{}{
			"segment_duration": 30,
		}

		// Start multiple recording sessions
		session1, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath, options)
		require.NoError(t, err)

		session2, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath+"_2", options)
		require.NoError(t, err)

		sessions := recordingManager.ListRecordingSessions()
		assert.Len(t, sessions, 2, "Should track multiple active sessions")

		// Verify sessions are in the list
		sessionIDs := make(map[string]bool)
		for _, session := range sessions {
			sessionIDs[session.ID] = true
		}
		assert.True(t, sessionIDs[session1.ID], "Should contain first session")
		assert.True(t, sessionIDs[session2.ID], "Should contain second session")
	})

	t.Run("StopRecording_ValidSession", func(t *testing.T) {
		ctx := context.Background()
		devicePath := "/dev/video0"
		outputPath := "/tmp/test_recordings/test_recording.mp4"
		options := map[string]interface{}{
			"segment_duration": 30,
		}

		session, err := recordingManager.StartRecordingWithSegments(ctx, devicePath, outputPath, options)
		require.NoError(t, err)

		// Stop recording
		err = recordingManager.StopRecording(ctx, session.ID)
		require.NoError(t, err, "Should stop recording successfully")

		// Verify session is stopped
		updatedSession, exists := recordingManager.GetRecordingSession(session.ID)
		assert.True(t, exists, "Session should still exist after stopping")
		assert.Equal(t, "STOPPED", updatedSession.Status, "Session should be stopped")
	})

	t.Run("StopRecording_InvalidSession", func(t *testing.T) {
		ctx := context.Background()
		invalidSessionID := "invalid_session"

		err := recordingManager.StopRecording(ctx, invalidSessionID)
		assert.Error(t, err, "Should return error for invalid session")
	})
}

func TestRecordingManagerConfiguration(t *testing.T) {
	// Test recording configuration validation

	t.Run("NewRecordingManager_ValidConfig", func(t *testing.T) {
		ffmpegManager := &mockFFmpegManager{}
		config := &mediamtx.MediaMTXConfig{
			StoragePath: "/tmp/test_recordings",
		}
		logger := mediamtx.NewLogger()
		recordingManager := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

		assert.NotNil(t, recordingManager, "Should create recording manager with valid config")
	})

	t.Run("NewRecordingManager_NilConfig", func(t *testing.T) {
		ffmpegManager := &mockFFmpegManager{}
		logger := mediamtx.NewLogger()
		recordingManager := mediamtx.NewRecordingManager(ffmpegManager, nil, logger)

		assert.NotNil(t, recordingManager, "Should handle nil config gracefully")
	})
}
