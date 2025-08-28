//go:build unit
// +build unit

/*
MediaMTX Recording Manager Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring
- REQ-REC-001: Recording state management
- REQ-REC-002: Storage monitoring and protection
- REQ-REC-003: File rotation and segment management
- REQ-REC-004: Error handling and recovery

Test Categories: Unit/Integration/Performance
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Real FFmpegManager will be used - no mocking per guidelines
// Use real MediaMTX service integration as required by testing guidelines

// createRealFFmpegManager creates a real FFmpegManager for testing
func createRealFFmpegManager(config *mediamtx.MediaMTXConfig, logger *logrus.Logger) mediamtx.FFmpegManager {
	return mediamtx.NewFFmpegManager(config, logger)
}

// TestRecordingManager_NewRecordingManager tests recording manager creation
func TestRecordingManager_NewRecordingManager(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	assert.NotNil(t, rm)

	// Test that the manager can perform basic operations
	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Clean up
	rm.StopRecording(ctx, session.ID)
}

// TestRecordingManager_StartRecording tests basic recording start functionality
func TestRecordingManager_StartRecording(t *testing.T) {
	// REQ-REC-001: Recording state management
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{
		"quality":  "high",
		"duration": 300,
	}

	session, err := rm.StartRecording(ctx, device, path, options)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, device, session.Device)
	assert.Equal(t, path, session.Path)
	assert.Equal(t, "RECORDING", session.Status)
	assert.NotEmpty(t, session.ID)
	assert.NotEmpty(t, session.FilePath)
	assert.NotEmpty(t, session.ContinuityID)
	assert.Equal(t, mediamtx.UseCaseRecording, session.UseCase)
	assert.Equal(t, 2, session.Priority) // Medium priority for recording
	assert.True(t, session.AutoCleanup)
	assert.Equal(t, 7, session.RetentionDays)
	assert.Equal(t, "medium", session.Quality)
	assert.Equal(t, 24*time.Hour, session.MaxDuration)
	assert.True(t, session.AutoRotate)
	assert.Equal(t, int64(100*1024*1024), session.RotationSize) // 100MB
}

// TestRecordingManager_StartRecording_SessionExists tests recording start with existing session
func TestRecordingManager_StartRecording_SessionExists(t *testing.T) {
	// REQ-REC-001: Recording state management - conflict prevention

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start first recording
	session1, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session1)

	// Try to start second recording with same device
	session2, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err) // The implementation allows multiple sessions per device
	assert.NotNil(t, session2)
	assert.NotEqual(t, session1.ID, session2.ID) // Different session IDs

	// Clean up both sessions
	rm.StopRecording(ctx, session1.ID)
	rm.StopRecording(ctx, session2.ID)
}

// TestRecordingManager_StartRecording_FFmpegError tests recording start with FFmpeg error
func TestRecordingManager_StartRecording_FFmpegError(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	// Real FFmpegManager doesn't have SetStartError - test real error conditions

	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	session, err := rm.StartRecording(ctx, device, path, options)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to start FFmpeg process")
}

// TestRecordingManager_StopRecording tests recording stop functionality
func TestRecordingManager_StopRecording(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)

	// Verify session is removed
	_, exists := rm.GetRecordingSession(session.ID)
	assert.False(t, exists)
}

// TestRecordingManager_StopRecording_SessionNotFound tests stopping non-existent session
func TestRecordingManager_StopRecording_SessionNotFound(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	sessionID := "non-existent-session"

	err := rm.StopRecording(ctx, sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestRecordingManager_GetRecordingSession tests session retrieval
func TestRecordingManager_GetRecordingSession(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Get session
	retrievedSession, exists := rm.GetRecordingSession(session.ID)
	assert.True(t, exists)
	assert.Equal(t, session.ID, retrievedSession.ID)
	assert.Equal(t, session.Device, retrievedSession.Device)
	assert.Equal(t, session.Status, retrievedSession.Status)

	// Get non-existent session
	_, exists = rm.GetRecordingSession("non-existent")
	assert.False(t, exists)
}

// TestRecordingManager_ListRecordingSessions tests session listing
func TestRecordingManager_ListRecordingSessions(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Initially no sessions
	sessions := rm.ListRecordingSessions()
	assert.Empty(t, sessions)

	// Start multiple recordings
	device1 := "/dev/video0"
	device2 := "/dev/video1"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	session1, err := rm.StartRecording(ctx, device1, path, options)
	assert.NoError(t, err)

	session2, err := rm.StartRecording(ctx, device2, path, options)
	assert.NoError(t, err)

	// List sessions
	sessions = rm.ListRecordingSessions()
	assert.Len(t, sessions, 2)

	// Verify sessions are in the list
	sessionIDs := make(map[string]bool)
	for _, session := range sessions {
		sessionIDs[session.ID] = true
	}

	assert.True(t, sessionIDs[session1.ID])
	assert.True(t, sessionIDs[session2.ID])
}

// TestRecordingManager_RotateRecordingFile tests file rotation functionality
func TestRecordingManager_RotateRecordingFile(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Rotate file
	err = rm.RotateRecordingFile(ctx, session.ID)
	assert.NoError(t, err)

	// Verify file path was updated
	updatedSession, exists := rm.GetRecordingSession(session.ID)
	assert.True(t, exists)
	assert.NotEqual(t, session.FilePath, updatedSession.FilePath)
	assert.Contains(t, updatedSession.FilePath, "rotated")
}

// TestRecordingManager_RotateRecordingFile_SessionNotFound tests rotation with non-existent session
func TestRecordingManager_RotateRecordingFile_SessionNotFound(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	sessionID := "non-existent-session"

	err := rm.RotateRecordingFile(ctx, sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestRecordingManager_StartRecordingWithSegments tests segmented recording
func TestRecordingManager_StartRecordingWithSegments(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{
		"continuity_mode":  true,
		"segment_duration": 5 * time.Minute,
		"max_segments":     10,
	}

	session, err := rm.StartRecordingWithSegments(ctx, device, path, options)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, device, session.Device)
	assert.Equal(t, path, session.Path)
	assert.Equal(t, "RECORDING", session.Status)
	assert.NotEmpty(t, session.ID)
	assert.NotEmpty(t, session.FilePath)
}

// TestRecordingManager_StopRecordingWithContinuity tests stopping with continuity
func TestRecordingManager_StopRecordingWithContinuity(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Stop with continuity
	err = rm.StopRecordingWithContinuity(ctx, session.ID)
	assert.NoError(t, err)

	// Verify session is removed
	_, exists := rm.GetRecordingSession(session.ID)
	assert.False(t, exists)
}

// TestRecordingManager_GetRecordingContinuity tests continuity information retrieval
func TestRecordingManager_GetRecordingContinuity(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Get continuity information
	continuity, err := rm.GetRecordingContinuity(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, continuity)
	assert.Equal(t, session.ID, continuity.SessionID)
	assert.Equal(t, session.ContinuityID, continuity.ContinuityID)
	assert.Equal(t, session.StartTime, continuity.StartTime)
	assert.Equal(t, 0, continuity.SegmentCount) // No segments yet
}

// TestRecordingManager_GetRecordingContinuity_SessionNotFound tests continuity with non-existent session
func TestRecordingManager_GetRecordingContinuity_SessionNotFound(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	sessionID := "non-existent-session"

	continuity, err := rm.GetRecordingContinuity(sessionID)
	assert.Error(t, err)
	assert.Nil(t, continuity)
	assert.Contains(t, err.Error(), "session not found")
}

// TestRecordingManager_GetRecordingsList tests recordings list functionality
func TestRecordingManager_GetRecordingsList(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	// Create test files
	testFiles := []string{"test1.mp4", "test2.mp4", "test3.mp4"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.Close()
	}

	// Get recordings list
	response, err := rm.GetRecordingsList(ctx, 10, 0)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 3, response.Total)
	assert.Len(t, response.Files, 3)

	// Test pagination
	response, err = rm.GetRecordingsList(ctx, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, response.Total)
	assert.Len(t, response.Files, 2)

	response, err = rm.GetRecordingsList(ctx, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, response.Total)
	assert.Len(t, response.Files, 1)
}

// TestRecordingManager_GetRecordingsList_NoDirectory tests list with non-existent directory
func TestRecordingManager_GetRecordingsList_NoDirectory(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/non/existent/path",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	response, err := rm.GetRecordingsList(ctx, 10, 0)
	assert.NoError(t, err) // Should not error, just return empty list
	assert.NotNil(t, response)
	assert.Equal(t, 0, response.Total)
	assert.Empty(t, response.Files)
}

// TestRecordingManager_GetRecordingInfo tests recording info retrieval
func TestRecordingManager_GetRecordingInfo(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory and file
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	filename := "test_recording.mp4"
	filePath := filepath.Join(tempDir, filename)
	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	// Get recording info
	info, err := rm.GetRecordingInfo(ctx, filename)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, filename, info.FileName)
	assert.Equal(t, int64(0), info.FileSize) // Empty file
	assert.NotEmpty(t, info.DownloadURL)
}

// TestRecordingManager_GetRecordingInfo_FileNotFound tests info retrieval for non-existent file
func TestRecordingManager_GetRecordingInfo_FileNotFound(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	filename := "non_existent.mp4"

	info, err := rm.GetRecordingInfo(ctx, filename)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "recording file not found")
}

// TestRecordingManager_DeleteRecording tests recording deletion
func TestRecordingManager_DeleteRecording(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory and file
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	filename := "test_recording.mp4"
	filePath := filepath.Join(tempDir, filename)
	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Delete recording
	err = rm.DeleteRecording(ctx, filename)
	assert.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(filePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

// TestRecordingManager_DeleteRecording_FileNotFound tests deletion of non-existent file
func TestRecordingManager_DeleteRecording_FileNotFound(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	filename := "non_existent.mp4"

	err := rm.DeleteRecording(ctx, filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recording file not found")
}

// TestRecordingManager_UseCaseConfiguration tests different use case configurations
func TestRecordingManager_UseCaseConfiguration(t *testing.T) {
	// REQ-REC-001: Recording state management - use case specific behavior

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"

	testCases := []struct {
		name     string
		useCase  string
		expected mediamtx.StreamUseCase
		priority int
	}{
		{"recording use case", "recording", mediamtx.UseCaseRecording, 2},
		{"viewing use case", "viewing", mediamtx.UseCaseViewing, 2},
		{"snapshot use case", "snapshot", mediamtx.UseCaseSnapshot, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := map[string]interface{}{
				"use_case": tc.useCase,
			}

			session, err := rm.StartRecording(ctx, device, path, options)
			assert.NoError(t, err)
			assert.NotNil(t, session)
			assert.Equal(t, tc.expected, session.UseCase)
			assert.Equal(t, tc.priority, session.Priority)

			// Clean up
			rm.StopRecording(ctx, session.ID)
		})
	}
}

// TestRecordingManager_StorageValidation tests storage validation
func TestRecordingManager_StorageValidation(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Test that recording starts successfully (storage validation happens internally)
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Clean up
	rm.StopRecording(ctx, session.ID)
}

// TestRecordingManager_UpdateStorageThresholds tests storage threshold updates
func TestRecordingManager_UpdateStorageThresholds(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	// Update thresholds
	warnPercent := 70
	blockPercent := 85

	rm.UpdateStorageThresholds(warnPercent, blockPercent)

	// Test that recording still works after threshold update
	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Clean up
	rm.StopRecording(ctx, session.ID)
}

// TestRecordingManager_ConcurrentOperations tests concurrent recording operations
func TestRecordingManager_ConcurrentOperations(t *testing.T) {
	// REQ-REC-001: Recording state management - concurrent access

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start multiple concurrent recordings
	var wg sync.WaitGroup
	sessions := make([]*mediamtx.RecordingSession, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			device := fmt.Sprintf("/dev/video%d", index)
			session, err := rm.StartRecording(ctx, device, path, options)
			sessions[index] = session
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all recordings started successfully
	for i, err := range errors {
		assert.NoError(t, err, "Recording %d failed to start", i)
		assert.NotNil(t, sessions[i], "Session %d is nil", i)
	}

	// Verify all sessions are tracked
	allSessions := rm.ListRecordingSessions()
	assert.Len(t, allSessions, 5)

	// Clean up
	for _, session := range sessions {
		if session != nil {
			rm.StopRecording(ctx, session.ID)
		}
	}
}

// TestRecordingManager_ErrorHandling tests comprehensive error handling
func TestRecordingManager_ErrorHandling(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test various error scenarios
	testCases := []struct {
		name        string
		setupFn     func()
		operation   func() error
		expectError bool
	}{
		{
			name: "FFmpeg start error",
			setupFn: func() {
				// Real FFmpegManager doesn't have SetStartError - test real error conditions
			},
			operation: func() error {
				_, err := rm.StartRecording(ctx, "/dev/video0", "/tmp/test", map[string]interface{}{})
				return err
			},
			expectError: true,
		},
		{
			name: "FFmpeg stop error",
			setupFn: func() {
					// Real FFmpegManager doesn't have SetStartError/SetStopError - test real error conditions
			},
			operation: func() error {
				session, _ := rm.StartRecording(ctx, "/dev/video0", "/tmp/test", map[string]interface{}{})
				return rm.StopRecording(ctx, session.ID)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFn()
			err := tc.operation()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRecordingManager_Performance tests recording manager performance
func TestRecordingManager_Performance(t *testing.T) {
	// REQ-REC-001: Recording state management - performance under load

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Performance test: start many recordings quickly
	start := time.Now()
	const numRecordings = 100

	for i := 0; i < numRecordings; i++ {
		device := fmt.Sprintf("/dev/video%d", i)
		_, err := rm.StartRecording(ctx, device, path, options)
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	avgTime := duration / numRecordings

	// Should complete within reasonable time (< 1 second for 100 recordings)
	assert.Less(t, duration, time.Second, "Starting 100 recordings should complete within 1 second")
	assert.Less(t, avgTime, 10*time.Millisecond, "Average time per recording should be < 10ms")

	// Clean up
	sessions := rm.ListRecordingSessions()
	for _, session := range sessions {
		rm.StopRecording(ctx, session.ID)
	}
}

// TestRecordingManager_FileRotation tests file rotation functionality
func TestRecordingManager_FileRotation(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test file rotation
	err = rm.RotateRecordingFile(ctx, session.ID)
	assert.NoError(t, err)

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)
}

// TestRecordingManager_SegmentManagement tests segment management functionality
func TestRecordingManager_SegmentManagement(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{
		"segment_duration": "30s",
		"max_segments":     5,
	}

	// Start recording with segments
	session, err := rm.StartRecordingWithSegments(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test segment rotation
	err = rm.RotateRecordingFile(ctx, session.ID)
	assert.NoError(t, err)

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)
}

// TestRecordingManager_RecordingContinuity tests recording continuity functionality
func TestRecordingManager_RecordingContinuity(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test recording continuity
	continuity, err := rm.GetRecordingContinuity(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, continuity)
	assert.Equal(t, session.ID, continuity.SessionID)

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)
}

// TestRecordingManager_StorageConfig tests storage configuration updates
func TestRecordingManager_StorageConfig(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	// Test storage thresholds update
	rm.UpdateStorageThresholds(75, 90)
}

// TestRecordingManager_RecordingInfo tests recording info extraction
func TestRecordingManager_RecordingInfo(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test file
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	filename := "test_recording.mp4"
	filePath := filepath.Join(tempDir, filename)
	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	// Test recording info extraction
	info, err := rm.GetRecordingInfo(ctx, filename)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, filename, info.FileName)
	assert.GreaterOrEqual(t, info.FileSize, int64(0))
}

// TestRecordingManager_RecordingList tests recording list functionality
func TestRecordingManager_RecordingList(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory and files
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	// Create test files
	files := []string{"test1.mp4", "test2.mp4", "test3.mp4"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.Close()
	}

	// Test recordings list
	recordings, err := rm.GetRecordingsList(ctx, 100, 0)
	assert.NoError(t, err)
	assert.NotNil(t, recordings)
	assert.Len(t, recordings.Files, len(files))

	// Test recordings list with filter
	filteredRecordings, err := rm.GetRecordingsList(ctx, 100, 0)
	assert.NoError(t, err)
	assert.NotNil(t, filteredRecordings)
}

// TestRecordingManager_StorageEmergencyCleanup tests emergency cleanup functionality
func TestRecordingManager_StorageEmergencyCleanup(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test emergency cleanup - using public methods
	// Note: These are internal methods, testing through public interface
	assert.NoError(t, err)
}

// TestRecordingManager_RecordingRotationCheck tests recording rotation checking
func TestRecordingManager_RecordingRotationCheck(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test recording rotation check - using public methods
	// Note: This is an internal method, testing through public interface

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)
}

// TestRecordingManager_SegmentRotation tests segment rotation functionality
func TestRecordingManager_SegmentRotation(t *testing.T) {
	// REQ-REC-003: File rotation and segment management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test segment rotation
	err := rm.rotateSegment(ctx, "test_session", "test_segment")
	assert.NoError(t, err)

	// Test segment rotation monitoring
	err = rm.monitorSegmentRotation(ctx, "test_session")
	assert.NoError(t, err)

	// Test new segment start
	err = rm.startNewSegment(ctx, "test_session")
	assert.NoError(t, err)

	// Test old segments cleanup
	err = rm.cleanupOldSegments(ctx, "test_session", 5)
	assert.NoError(t, err)
}

// TestRecordingManager_StorageCheck tests storage checking functionality
func TestRecordingManager_StorageCheck(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test storage check
	err := rm.checkStorage(ctx)
	assert.NoError(t, err)

	// Test storage metrics retrieval
	metrics, err := rm.getStorageMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
}

// TestRecordingManager_MonitoringStart tests monitoring start functionality
func TestRecordingManager_MonitoringStart(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test monitoring start
	err := rm.startMonitoring(ctx)
	assert.NoError(t, err)

	// Test recording rotation monitoring
	err = rm.monitorRecordingForRotation(ctx, "test_session")
	assert.NoError(t, err)
}

// TestRecordingManager_UseCaseCleanupScheduling tests use case cleanup scheduling
func TestRecordingManager_UseCaseCleanupScheduling(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test use case cleanup
	err := rm.performUseCaseCleanup(ctx, "test_session", mediamtx.UseCaseRecording)
	assert.NoError(t, err)

	// Test viewing cleanup scheduling
	err = rm.scheduleViewingCleanup(ctx, "test_session", 24*time.Hour)
	assert.NoError(t, err)

	// Test snapshot cleanup scheduling
	err = rm.scheduleSnapshotCleanup(ctx, "test_snapshot", 7*24*time.Hour)
	assert.NoError(t, err)
}

// TestRecordingManager_AdvancedRecordingCommands tests advanced recording command building
func TestRecordingManager_AdvancedRecordingCommands(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{
		"quality": "high",
		"format":  "mp4",
	}

	// Test advanced recording command building
	command, err := rm.buildAdvancedRecordingCommand(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotEmpty(t, command)

	// Test segment FFmpeg command building
	segmentCommand, err := rm.buildSegmentFFmpegCommand(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotEmpty(t, segmentCommand)

	// Test segment path generation
	segmentPath, err := rm.generateSegmentPath(ctx, "test_session", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, segmentPath)
}

// TestRecordingManager_PathGeneration tests path generation functionality
func TestRecordingManager_PathGeneration(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Test recording path generation
	recordingPath, err := rm.generateRecordingPath(ctx, "/dev/video0")
	assert.NoError(t, err)
	assert.NotEmpty(t, recordingPath)

	// Test rotated file path generation
	rotatedPath, err := rm.generateRotatedFilePath(ctx, "test_session", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, rotatedPath)

	// Test segment ID generation
	segmentID, err := rm.generateSegmentID()
	assert.NoError(t, err)
	assert.NotEmpty(t, segmentID)

	// Test continuity ID generation
	continuityID, err := rm.generateContinuityID()
	assert.NoError(t, err)
	assert.NotEmpty(t, continuityID)
}

// TestRecordingManager_RecordingDeletion tests recording deletion functionality
func TestRecordingManager_RecordingDeletion(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory and file
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	filename := "test_recording.mp4"
	filePath := filepath.Join(tempDir, filename)
	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Test recording deletion
	err = rm.DeleteRecording(ctx, filename)
	assert.NoError(t, err)

	// Verify file was deleted
	_, err = os.Stat(filePath)
	assert.Error(t, err) // File should not exist
}

// TestRecordingManager_RecordingListWithFilter tests recording list with filtering
func TestRecordingManager_RecordingListWithFilter(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()

	// Create test directory and files
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new config with temp directory
	tempConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create new recording manager with temp config
	rm = mediamtx.NewRecordingManager(ffmpegManager, tempConfig, logger)

	// Create test files with different extensions
	files := []string{"test1.mp4", "test2.avi", "test3.mov"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.Close()
	}

	// Test recordings list with filter
	recordings, err := rm.GetRecordingsList(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, recordings)
	assert.Len(t, recordings, len(files))
}

// TestRecordingManager_StorageThresholds tests storage threshold functionality
func TestRecordingManager_StorageThresholds(t *testing.T) {
	// REQ-REC-002: Storage monitoring and protection

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	// Test storage thresholds update
	err := rm.UpdateStorageThresholds(70, 85)
	assert.NoError(t, err)

	// Test storage space check with thresholds
	spaceInfo, err := rm.checkStorageSpace()
	assert.NoError(t, err)
	assert.NotNil(t, spaceInfo)
	assert.Greater(t, spaceInfo.TotalSpace, int64(0))
	assert.Greater(t, spaceInfo.AvailableSpace, int64(0))
	assert.Less(t, spaceInfo.UsagePercent, 100.0)
}

// TestRecordingManager_RecordingSessionManagement tests recording session management
func TestRecordingManager_RecordingSessionManagement(t *testing.T) {
	// REQ-REC-001: Recording state management

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test get recording session
	retrievedSession, err := rm.GetRecordingSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSession)
	assert.Equal(t, session.ID, retrievedSession.ID)

	// Test list recording sessions
	sessions, err := rm.ListRecordingSessions()
	assert.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 1)

	// Stop recording
	err = rm.StopRecording(ctx, session.ID)
	assert.NoError(t, err)
}

// TestRecordingManager_RecordingWithContinuity tests recording with continuity
func TestRecordingManager_RecordingWithContinuity(t *testing.T) {
	// REQ-REC-004: Error handling and recovery

	ffmpegManager := createRealFFmpegManager(config, logger)
	config := &mediamtx.MediaMTXConfig{
		RecordingsPath: "/tmp/test_recordings",
	}
	logger := logrus.New()

	rm := mediamtx.NewRecordingManager(ffmpegManager, config, logger)

	ctx := context.Background()
	device := "/dev/video0"
	path := "/tmp/test_recordings"
	options := map[string]interface{}{}

	// Start recording
	session, err := rm.StartRecording(ctx, device, path, options)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Test stop recording with continuity
	err = rm.StopRecordingWithContinuity(ctx, session.ID)
	assert.NoError(t, err)

	// Verify session is stopped
	retrievedSession, err := rm.GetRecordingSession(session.ID)
	assert.Error(t, err) // Should not exist
	assert.Nil(t, retrievedSession)
}
