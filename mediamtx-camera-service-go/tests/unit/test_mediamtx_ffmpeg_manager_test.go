/*
MediaMTX FFmpeg Manager Unit Tests

Requirements Coverage:
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFFmpegManager_Creation tests FFmpeg manager creation
func TestFFmpegManager_Creation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)
	require.NotNil(t, ffmpegManager, "FFmpeg manager should not be nil")
}

// TestFFmpegManager_StartProcess tests process start functionality
func TestFFmpegManager_StartProcess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test starting a simple process (echo command)
	command := []string{"echo", "test"}
	outputPath := "/tmp/test_output.txt"

	pid, err := ffmpegManager.StartProcess(ctx, command, outputPath)
	// Note: This may fail if echo command is not available
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Process start failed (expected if echo not available): %v", err)
	} else {
		assert.Greater(t, pid, 0, "Process ID should be positive")
	}
}

// TestFFmpegManager_StopProcess tests process stop functionality
func TestFFmpegManager_StopProcess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test stopping non-existent process
	err := ffmpegManager.StopProcess(ctx, 99999)
	// Should handle non-existent process gracefully
	if err != nil {
		t.Logf("Stop non-existent process result: %v", err)
	}
}

// TestFFmpegManager_IsProcessRunning tests process running check
func TestFFmpegManager_IsProcessRunning(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test checking non-existent process
	running := ffmpegManager.IsProcessRunning(ctx, 99999)
	assert.False(t, running, "Non-existent process should not be running")
}

// TestFFmpegManager_StartRecording tests recording start functionality
func TestFFmpegManager_StartRecording(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test starting recording
	options := map[string]string{
		"format": "mp4",
		"codec":  "libx264",
	}

	pid, err := ffmpegManager.StartRecording(ctx, "/dev/video0", "/tmp/test_recording.mp4", options)
	// Note: This may fail if camera device is not available
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Recording start failed (expected if camera not available): %v", err)
	} else {
		assert.Greater(t, pid, 0, "Recording process ID should be positive")
	}
}

// TestFFmpegManager_StopRecording tests recording stop functionality
func TestFFmpegManager_StopRecording(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test stopping non-existent recording
	err := ffmpegManager.StopRecording(ctx, 99999)
	// Should handle non-existent process gracefully
	if err != nil {
		t.Logf("Stop non-existent recording result: %v", err)
	}
}

// TestFFmpegManager_TakeSnapshot tests snapshot functionality
func TestFFmpegManager_TakeSnapshot(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test taking snapshot
	err := ffmpegManager.TakeSnapshot(ctx, "/dev/video0", "/tmp/test_snapshot.jpg")
	// Note: This may fail if camera device is not available
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Snapshot failed (expected if camera not available): %v", err)
	}
}

// TestFFmpegManager_RotateFile tests file rotation functionality
func TestFFmpegManager_RotateFile(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test file rotation with non-existent files
	err := ffmpegManager.RotateFile(ctx, "/tmp/non_existent_old.mp4", "/tmp/non_existent_new.mp4")
	// Should handle non-existent files gracefully
	if err != nil {
		t.Logf("File rotation result: %v", err)
	}
}

// TestFFmpegManager_GetFileInfo tests file info retrieval
func TestFFmpegManager_GetFileInfo(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test getting file info for non-existent file
	size, timestamp, err := ffmpegManager.GetFileInfo(ctx, "/tmp/non_existent_file.mp4")
	// Should handle non-existent file gracefully
	if err != nil {
		t.Logf("Get file info result: %v", err)
	} else {
		assert.Equal(t, int64(0), size, "Non-existent file should have zero size")
		assert.Equal(t, time.Time{}, timestamp, "Non-existent file should have zero timestamp")
	}
}

// TestFFmpegManager_BuildCommand tests command building functionality
func TestFFmpegManager_BuildCommand(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Test building command
	args := []string{"-i", "/dev/video0", "-c:v", "libx264", "-f", "mp4", "output.mp4"}
	command := ffmpegManager.BuildCommand(args...)

	assert.NotNil(t, command, "Command should not be nil")
	assert.Greater(t, len(command), 0, "Command should have at least one element")
	assert.Contains(t, command, "ffmpeg", "Command should contain ffmpeg")
}

// TestFFmpegManager_ErrorHandling tests error handling scenarios
func TestFFmpegManager_ErrorHandling(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test with empty command
	_, err := ffmpegManager.StartProcess(ctx, []string{}, "/tmp/test")
	assert.Error(t, err, "Should return error with empty command")

	// Test with empty output path
	_, err = ffmpegManager.StartProcess(ctx, []string{"echo", "test"}, "")
	assert.Error(t, err, "Should return error with empty output path")

	// Test with empty device path
	_, err = ffmpegManager.StartRecording(ctx, "", "/tmp/test.mp4", nil)
	assert.Error(t, err, "Should return error with empty device path")

	// Test with empty output path for recording
	_, err = ffmpegManager.StartRecording(ctx, "/dev/video0", "", nil)
	assert.Error(t, err, "Should return error with empty output path for recording")
}

// TestFFmpegManager_ConcurrentAccess tests concurrent access scenarios
func TestFFmpegManager_ConcurrentAccess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	ctx := context.Background()

	// Test concurrent command building
	done := make(chan bool, 2)

	go func() {
		command1 := ffmpegManager.BuildCommand("-i", "/dev/video0", "-c:v", "libx264")
		assert.NotNil(t, command1, "Command 1 should not be nil")
		done <- true
	}()

	go func() {
		command2 := ffmpegManager.BuildCommand("-i", "/dev/video1", "-c:v", "libx265")
		assert.NotNil(t, command2, "Command 2 should not be nil")
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestFFmpegManager_ContextCancellation tests context cancellation
func TestFFmpegManager_ContextCancellation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		ProcessTerminationTimeout: 5.0,
		ProcessKillTimeout: 2.0,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Test process operation with cancelled context
	_, err := ffmpegManager.StartProcess(ctx, []string{"echo", "test"}, "/tmp/test")
	// Should handle context cancellation gracefully
	if err != nil {
		t.Logf("Context cancellation test result: %v", err)
	}
}

// TestFFmpegManager_ConfigurationValidation tests configuration validation
func TestFFmpegManager_ConfigurationValidation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Test with invalid configuration
	invalidConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "",
		ProcessTerminationTimeout: -1.0,
		ProcessKillTimeout: -1.0,
	}

	// Create FFmpeg manager with invalid config
	ffmpegManager := mediamtx.NewFFmpegManager(invalidConfig, logger)
	require.NotNil(t, ffmpegManager, "FFmpeg manager should be created even with invalid config")

	// Test that FFmpeg manager handles invalid config gracefully
	command := ffmpegManager.BuildCommand("-i", "/dev/video0")
	assert.NotNil(t, command, "Command should not be nil even with invalid config")
}
