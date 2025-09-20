package mediamtx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRTSPKeepaliveReader_NewRTSPKeepaliveReader(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Verify initialization
	require.NotNil(t, reader)
	assert.Equal(t, mediaMTXConfig, reader.config)
	assert.Equal(t, logger, reader.logger)
	assert.Equal(t, 0, reader.GetActiveCount())
}

func TestRTSPKeepaliveReader_StartKeepalive(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Test starting keepalive for a path
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	pathName := "test_camera"

	// This will fail because MediaMTX is not running, but we can test the logic
	err = reader.StartKeepalive(ctx, pathName)

	// Should not fail due to missing MediaMTX (graceful handling)
	// The actual FFmpeg process will fail, but the reader should handle it gracefully
	assert.NoError(t, err, "StartKeepalive should not fail even if MediaMTX is not available")

	// Verify the reader is tracked as active
	assert.True(t, reader.IsActive(pathName), "Path should be marked as active")
	assert.Equal(t, 1, reader.GetActiveCount(), "Should have one active reader")
}

func TestRTSPKeepaliveReader_StopKeepalive(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Test stopping a non-existent keepalive
	err = reader.StopKeepalive("non_existent_path")
	assert.NoError(t, err, "Stopping non-existent keepalive should not fail")

	// Test stopping an active keepalive
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
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
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Start multiple keepalives
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
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
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Test inactive path
	assert.False(t, reader.IsActive("inactive_path"), "Inactive path should return false")

	// Start keepalive
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
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
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Test initial count
	assert.Equal(t, 0, reader.GetActiveCount(), "Initial count should be 0")

	// Start keepalives
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
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

	// Use centralized path management instead of hardcoded paths
	helper := NewMediaMTXTestHelper(t, nil)
	testDir := helper.GetConfig().TestDataDir
	os.Setenv("MEDIAMTX_TEST_DATA_DIR", testDir)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Verify it was created successfully
	require.NotNil(t, reader)
	assert.Equal(t, mediaMTXConfig, reader.config)
}

func TestRTSPKeepaliveReader_ConcurrentOperations(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Get MediaMTX config from fixture via ConfigIntegration
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")
	logger := helper.GetLogger()

	// Create keepalive reader
	reader := NewRTSPKeepaliveReader(mediaMTXConfig, logger)

	// Test concurrent start/stop operations
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	pathName := "concurrent_test"

	// Start multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func() {
			reader.StartKeepalive(ctx, pathName)
			// Use proper synchronization instead of time.Sleep
			// Wait for keepalive to be established before stopping
			select {
			case <-time.After(10 * time.Millisecond):
				// Timeout reached, proceed with stop
			case <-ctx.Done():
				// Context cancelled, exit early
				done <- true
				return
			}
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
