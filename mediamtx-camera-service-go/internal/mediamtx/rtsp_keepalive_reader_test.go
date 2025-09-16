package mediamtx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRTSPKeepaliveReader_NewRTSPKeepaliveReader(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Verify initialization
	require.NotNil(t, reader)
	assert.Equal(t, cfg, reader.config)
	assert.Equal(t, logger, reader.logger)
	assert.Equal(t, 0, reader.GetActiveCount())
}

func TestRTSPKeepaliveReader_StartKeepalive(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Test starting keepalive for a path
	ctx := context.Background()
	pathName := "test_camera"

	// This will fail because MediaMTX is not running, but we can test the logic
	err := reader.StartKeepalive(ctx, pathName)

	// Should not fail due to missing MediaMTX (graceful handling)
	// The actual FFmpeg process will fail, but the reader should handle it gracefully
	assert.NoError(t, err, "StartKeepalive should not fail even if MediaMTX is not available")

	// Verify the reader is tracked as active
	assert.True(t, reader.IsActive(pathName), "Path should be marked as active")
	assert.Equal(t, 1, reader.GetActiveCount(), "Should have one active reader")
}

func TestRTSPKeepaliveReader_StopKeepalive(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Test stopping a non-existent keepalive
	err := reader.StopKeepalive("non_existent_path")
	assert.NoError(t, err, "Stopping non-existent keepalive should not fail")

	// Test stopping an active keepalive
	ctx := context.Background()
	pathName := "test_camera"

	// Start keepalive (will fail gracefully if MediaMTX not available)
	reader.StartKeepalive(ctx, pathName)

	// Stop keepalive
	err = reader.StopKeepalive(pathName)
	assert.NoError(t, err, "Stopping keepalive should not fail")

	// Verify the reader is no longer tracked as active
	assert.False(t, reader.IsActive(pathName), "Path should no longer be marked as active")
	assert.Equal(t, 0, reader.GetActiveCount(), "Should have no active readers")
}

func TestRTSPKeepaliveReader_StopAll(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Start multiple keepalives
	ctx := context.Background()
	reader.StartKeepalive(ctx, "camera1")
	reader.StartKeepalive(ctx, "camera2")
	reader.StartKeepalive(ctx, "camera3")

	// Verify all are active
	assert.Equal(t, 3, reader.GetActiveCount(), "Should have three active readers")

	// Stop all
	reader.StopAll()

	// Verify all are stopped
	assert.Equal(t, 0, reader.GetActiveCount(), "Should have no active readers")
	assert.False(t, reader.IsActive("camera1"), "Camera1 should not be active")
	assert.False(t, reader.IsActive("camera2"), "Camera2 should not be active")
	assert.False(t, reader.IsActive("camera3"), "Camera3 should not be active")
}

func TestRTSPKeepaliveReader_IsActive(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Test inactive path
	assert.False(t, reader.IsActive("inactive_path"), "Inactive path should return false")

	// Start keepalive
	ctx := context.Background()
	pathName := "active_path"
	reader.StartKeepalive(ctx, pathName)

	// Test active path
	assert.True(t, reader.IsActive(pathName), "Active path should return true")

	// Stop keepalive
	reader.StopKeepalive(pathName)

	// Test inactive path again
	assert.False(t, reader.IsActive(pathName), "Stopped path should return false")
}

func TestRTSPKeepaliveReader_GetActiveCount(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Test initial count
	assert.Equal(t, 0, reader.GetActiveCount(), "Initial count should be 0")

	// Start keepalives
	ctx := context.Background()
	reader.StartKeepalive(ctx, "camera1")
	assert.Equal(t, 1, reader.GetActiveCount(), "Count should be 1 after starting one")

	reader.StartKeepalive(ctx, "camera2")
	assert.Equal(t, 2, reader.GetActiveCount(), "Count should be 2 after starting two")

	reader.StartKeepalive(ctx, "camera3")
	assert.Equal(t, 3, reader.GetActiveCount(), "Count should be 3 after starting three")

	// Stop one
	reader.StopKeepalive("camera2")
	assert.Equal(t, 2, reader.GetActiveCount(), "Count should be 2 after stopping one")

	// Stop all
	reader.StopAll()
	assert.Equal(t, 0, reader.GetActiveCount(), "Count should be 0 after stopping all")
}

func TestRTSPKeepaliveReader_EnvironmentVariables(t *testing.T) {
	// Test with environment variables for paths
	originalEnv := os.Getenv("MEDIAMTX_TEST_DATA_DIR")
	defer func() {
		if originalEnv != "" {
			os.Setenv("MEDIAMTX_TEST_DATA_DIR", originalEnv)
		} else {
			os.Unsetenv("MEDIAMTX_TEST_DATA_DIR")
		}
	}()

	// Set test environment variable
	testDir := "/tmp/test_mediamtx_env"
	os.Setenv("MEDIAMTX_TEST_DATA_DIR", testDir)

	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Verify it was created successfully
	require.NotNil(t, reader)
	assert.Equal(t, cfg, reader.config)
}

func TestRTSPKeepaliveReader_ConcurrentOperations(t *testing.T) {
	// Create test config
	cfg := &config.MediaMTXConfig{
		Host:     "localhost",
		RTSPPort: 8554,
	}

	// Create logger
	logger := logging.CreateTestLogger(t, nil)

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(cfg, logger)

	// Test concurrent start/stop operations
	ctx := context.Background()
	pathName := "concurrent_test"

	// Start multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func() {
			reader.StartKeepalive(ctx, pathName)
			time.Sleep(10 * time.Millisecond)
			reader.StopKeepalive(pathName)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify final state
	assert.Equal(t, 0, reader.GetActiveCount(), "Final count should be 0 after concurrent operations")
	assert.False(t, reader.IsActive(pathName), "Path should not be active after concurrent operations")
}
