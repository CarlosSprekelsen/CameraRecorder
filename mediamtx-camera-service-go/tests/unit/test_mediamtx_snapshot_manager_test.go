/*
MediaMTX Snapshot Manager Unit Tests

Requirements Coverage:
- REQ-MTX-005: Multi-tier snapshot functionality
- REQ-MTX-006: Configuration integration
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFFmpegManager implements FFmpegManager interface for testing
type mockFFmpegManager struct{}

func (m *mockFFmpegManager) GetFileInfo(ctx context.Context, filePath string) (int64, string, error) {
	return 0, "", nil
}

func (m *mockFFmpegManager) BuildCommand(args ...string) []string {
	return []string{"ffmpeg"}
}

// TestSnapshotManager_Creation tests snapshot manager creation
func TestSnapshotManager_Creation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)
	require.NotNil(t, snapshotManager, "Snapshot manager should not be nil")

	// Verify snapshot manager has default settings
	settings := snapshotManager.GetSnapshotSettings()
	assert.NotNil(t, settings, "Snapshot settings should not be nil")
	assert.Equal(t, "jpg", settings.Format, "Default format should be jpg")
	assert.Equal(t, 85, settings.Quality, "Default quality should be 85")
}

// TestSnapshotManager_TakeSnapshot tests basic snapshot functionality
func TestSnapshotManager_TakeSnapshot(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	ctx := context.Background()

	// Test snapshot with options
	options := map[string]interface{}{
		"format":  "png",
		"quality": 90,
	}

	// Note: This test requires actual camera hardware or mock setup
	// For unit testing, we test the method signature and error handling
	_, err := snapshotManager.TakeSnapshot(ctx, "/dev/video0", "/tmp/test_snapshot", options)
	// We expect an error since we don't have actual camera hardware in unit tests
	// This validates that the method exists and handles errors appropriately
	assert.Error(t, err, "Should return error when camera not available")
}

// TestSnapshotManager_GetSnapshot tests snapshot retrieval
func TestSnapshotManager_GetSnapshot(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Test getting non-existent snapshot
	snapshot, exists := snapshotManager.GetSnapshot("non-existent-id")
	assert.False(t, exists, "Non-existent snapshot should not exist")
	assert.Nil(t, snapshot, "Non-existent snapshot should be nil")
}

// TestSnapshotManager_ListSnapshots tests snapshot listing
func TestSnapshotManager_ListSnapshots(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Test listing snapshots (should be empty initially)
	snapshots := snapshotManager.ListSnapshots()
	assert.NotNil(t, snapshots, "Snapshots list should not be nil")
	assert.Len(t, snapshots, 0, "Initial snapshots list should be empty")
}

// TestSnapshotManager_DeleteSnapshot tests snapshot deletion
func TestSnapshotManager_DeleteSnapshot(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	ctx := context.Background()

	// Test deleting non-existent snapshot
	err := snapshotManager.DeleteSnapshot(ctx, "non-existent-id")
	assert.Error(t, err, "Should return error when deleting non-existent snapshot")
}

// TestSnapshotManager_CleanupOldSnapshots tests snapshot cleanup
func TestSnapshotManager_CleanupOldSnapshots(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	ctx := context.Background()

	// Test cleanup with no snapshots (should not error)
	err := snapshotManager.CleanupOldSnapshots(ctx, 24*time.Hour, 100)
	assert.NoError(t, err, "Cleanup should not error when no snapshots exist")
}

// TestSnapshotManager_GetSnapshotSettings tests snapshot settings retrieval
func TestSnapshotManager_GetSnapshotSettings(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Test getting snapshot settings
	settings := snapshotManager.GetSnapshotSettings()
	assert.NotNil(t, settings, "Snapshot settings should not be nil")
	assert.Equal(t, "jpg", settings.Format, "Default format should be jpg")
	assert.Equal(t, 85, settings.Quality, "Default quality should be 85")
}

// TestSnapshotManager_UpdateSnapshotSettings tests snapshot settings update
func TestSnapshotManager_UpdateSnapshotSettings(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Create new settings
	newSettings := &mediamtx.SnapshotSettings{
		Format:      "png",
		Quality:     90,
		MaxWidth:    1920,
		MaxHeight:   1080,
		AutoResize:  true,
		Compression: 8,
	}

	// Test updating snapshot settings
	snapshotManager.UpdateSnapshotSettings(newSettings)

	// Verify settings were updated
	settings := snapshotManager.GetSnapshotSettings()
	assert.Equal(t, "png", settings.Format, "Format should be updated to png")
	assert.Equal(t, 90, settings.Quality, "Quality should be updated to 90")
	assert.Equal(t, 1920, settings.MaxWidth, "MaxWidth should be updated to 1920")
	assert.Equal(t, 1080, settings.MaxHeight, "MaxHeight should be updated to 1080")
	assert.True(t, settings.AutoResize, "AutoResize should be updated to true")
	assert.Equal(t, 8, settings.Compression, "Compression should be updated to 8")
}

// TestSnapshotManager_MultiTierConfiguration tests multi-tier configuration
func TestSnapshotManager_MultiTierConfiguration(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Test getting tier configuration (should return defaults when no config manager)
	tierConfig := snapshotManager.GetTierConfiguration()
	assert.NotNil(t, tierConfig, "Tier configuration should not be nil")
	assert.Equal(t, 0.5, tierConfig.Tier1USBDirectTimeout, "Tier1 timeout should be 0.5")
	assert.Equal(t, 1.0, tierConfig.Tier2RTSPReadyCheckTimeout, "Tier2 timeout should be 1.0")
	assert.Equal(t, 3.0, tierConfig.Tier3ActivationTimeout, "Tier3 timeout should be 3.0")
}

// TestSnapshotManager_ErrorHandling tests error handling scenarios
func TestSnapshotManager_ErrorHandling(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	ctx := context.Background()

	// Test snapshot with invalid device path
	_, err := snapshotManager.TakeSnapshot(ctx, "", "/tmp/test", nil)
	assert.Error(t, err, "Should return error with empty device path")

	// Test snapshot with invalid output path
	_, err = snapshotManager.TakeSnapshot(ctx, "/dev/video0", "", nil)
	assert.Error(t, err, "Should return error with empty output path")
}

// TestSnapshotManager_ConcurrentAccess tests concurrent access scenarios
func TestSnapshotManager_ConcurrentAccess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	// Test concurrent snapshot settings access
	done := make(chan bool, 2)

	go func() {
		settings := snapshotManager.GetSnapshotSettings()
		assert.NotNil(t, settings, "Settings should not be nil")
		done <- true
	}()

	go func() {
		newSettings := &mediamtx.SnapshotSettings{
			Format:  "png",
			Quality: 90,
		}
		snapshotManager.UpdateSnapshotSettings(newSettings)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestSnapshotManager_ConfigurationIntegration tests configuration integration
func TestSnapshotManager_ConfigurationIntegration(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration manager
	configManager := config.NewConfigManager()

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock FFmpeg manager
	ffmpegManager := &mediamtx.MockFFmpegManager{}

	// Create snapshot manager with configuration integration
	snapshotManager := mediamtx.NewSnapshotManagerWithConfig(ffmpegManager, testConfig, configManager, logger)
	require.NotNil(t, snapshotManager, "Snapshot manager should not be nil")

	// Test that configuration integration works
	tierConfig := snapshotManager.GetTierConfiguration()
	assert.NotNil(t, tierConfig, "Tier configuration should not be nil")
}
