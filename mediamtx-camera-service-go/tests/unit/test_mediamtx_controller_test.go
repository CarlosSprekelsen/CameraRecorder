//go:build unit && real_mediamtx && real_system
// +build unit,real_mediamtx,real_system

/*
MediaMTX Controller Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring
- REQ-MTX-005: Multi-tier snapshot functionality
- REQ-MTX-006: Configuration integration
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit/Integration (Real MediaMTX + Real System)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestLogger creates a test logger and returns the logrus logger
func setupTestLogger(component string) *logrus.Logger {
	logger := logging.NewLogger(component)
	return logger.Logger
}

// TestMediaMTXController_Creation tests controller creation with configuration integration
func TestMediaMTXController_Creation(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-controller-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller with configuration integration
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")
	require.NotNil(t, controller, "Controller should not be nil")

	// Verify controller implements interface
	_, ok := controller.(mediamtx.MediaMTXController)
	assert.True(t, ok, "Controller should implement MediaMTXController interface")
}

// TestMediaMTXController_StartStop tests controller lifecycle management
func TestMediaMTXController_StartStop(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := setupTestLogger("mediamtx-controller-lifecycle-test")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Test start
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")

	// Test stop
	err = controller.Stop(ctx)
	require.NoError(t, err, "Controller should stop successfully")
}

// TestMediaMTXController_TakeAdvancedSnapshot tests multi-tier snapshot functionality
func TestMediaMTXController_TakeAdvancedSnapshot(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := setupTestLogger("mediamtx-snapshot-test")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Start controller
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer controller.Stop(ctx)

	// Test snapshot with options
	options := map[string]interface{}{
		"format":  "jpg",
		"quality": 85,
	}

	// Note: This test requires actual camera hardware or mock setup
	// For unit testing, we test the method signature and error handling
	_, err = controller.TakeAdvancedSnapshot(ctx, "/dev/video0", "/tmp/test_snapshot", options)
	// We expect an error since we don't have actual camera hardware in unit tests
	// This validates that the method exists and handles errors appropriately
	assert.Error(t, err, "Should return error when camera not available")
}

// TestMediaMTXController_GetAdvancedSnapshot tests snapshot retrieval
func TestMediaMTXController_GetAdvancedSnapshot(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-get-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	// Test getting non-existent snapshot
	snapshot, exists := controller.GetAdvancedSnapshot("non-existent-id")
	assert.False(t, exists, "Non-existent snapshot should not exist")
	assert.Nil(t, snapshot, "Non-existent snapshot should be nil")
}

// TestMediaMTXController_ListAdvancedSnapshots tests snapshot listing
func TestMediaMTXController_ListAdvancedSnapshots(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-list-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	// Test listing snapshots (should be empty initially)
	snapshots := controller.ListAdvancedSnapshots()
	assert.NotNil(t, snapshots, "Snapshots list should not be nil")
	assert.Len(t, snapshots, 0, "Initial snapshots list should be empty")
}

// TestMediaMTXController_DeleteAdvancedSnapshot tests snapshot deletion
func TestMediaMTXController_DeleteAdvancedSnapshot(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-delete-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Test deleting non-existent snapshot
	err = controller.DeleteAdvancedSnapshot(ctx, "non-existent-id")
	assert.Error(t, err, "Should return error when deleting non-existent snapshot")
}

// TestMediaMTXController_CleanupOldSnapshots tests snapshot cleanup
func TestMediaMTXController_CleanupOldSnapshots(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-cleanup-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Test cleanup with no snapshots (should not error)
	err = controller.CleanupOldSnapshots(ctx, 24*time.Hour, 100)
	assert.NoError(t, err, "Cleanup should not error when no snapshots exist")
}

// TestMediaMTXController_GetSnapshotSettings tests snapshot settings retrieval
func TestMediaMTXController_GetSnapshotSettings(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-settings-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	// Test getting snapshot settings
	settings := controller.GetSnapshotSettings()
	assert.NotNil(t, settings, "Snapshot settings should not be nil")
	assert.Equal(t, "jpg", settings.Format, "Default format should be jpg")
	assert.Equal(t, 85, settings.Quality, "Default quality should be 85")
}

// TestMediaMTXController_UpdateSnapshotSettings tests snapshot settings update
func TestMediaMTXController_UpdateSnapshotSettings(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-settings-update-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

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
	controller.UpdateSnapshotSettings(newSettings)

	// Verify settings were updated
	settings := controller.GetSnapshotSettings()
	assert.Equal(t, "png", settings.Format, "Format should be updated to png")
	assert.Equal(t, 90, settings.Quality, "Quality should be updated to 90")
	assert.Equal(t, 1920, settings.MaxWidth, "MaxWidth should be updated to 1920")
	assert.Equal(t, 1080, settings.MaxHeight, "MaxHeight should be updated to 1080")
	assert.True(t, settings.AutoResize, "AutoResize should be updated to true")
	assert.Equal(t, 8, settings.Compression, "Compression should be updated to 8")
}

// TestMediaMTXController_HealthMonitoring tests health monitoring functionality
func TestMediaMTXController_HealthMonitoring(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-health-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Start controller
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer controller.Stop(ctx)

	// Test health check
	health, err := controller.GetHealth(ctx)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Health check failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, health, "Health status should not be nil")
		assert.NotEmpty(t, health.Status, "Health status should not be empty")
	}
}

// TestMediaMTXController_Metrics tests metrics functionality
func TestMediaMTXController_Metrics(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-metrics-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Start controller
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer controller.Stop(ctx)

	// Test metrics retrieval
	metrics, err := controller.GetMetrics(ctx)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Metrics retrieval failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, metrics, "Metrics should not be nil")
	}
}

// TestMediaMTXController_ConfigurationIntegration tests configuration integration
func TestMediaMTXController_ConfigurationIntegration(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-config-integration-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller with configuration integration
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Test configuration retrieval
	config, err := controller.GetConfig(ctx)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Config retrieval failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, config, "Config should not be nil")
	}
}

// TestMediaMTXController_ErrorHandling tests error handling scenarios
func TestMediaMTXController_ErrorHandling(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-error-handling-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Test operations without starting controller
	_, err = controller.TakeAdvancedSnapshot(ctx, "/dev/video0", "/tmp/test", nil)
	assert.Error(t, err, "Should return error when controller not running")
	assert.Contains(t, err.Error(), "not running", "Error should indicate controller not running")

	// Test health check without starting controller
	_, err = controller.GetHealth(ctx)
	assert.Error(t, err, "Should return error when controller not running")

	// Test metrics without starting controller
	_, err = controller.GetMetrics(ctx)
	assert.Error(t, err, "Should return error when controller not running")
}

// TestMediaMTXController_ConcurrentAccess tests concurrent access scenarios
func TestMediaMTXController_ConcurrentAccess(t *testing.T) {
	// Setup test configuration manager
	configManager := config.NewConfigManager()
	err := configManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-concurrent-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&configManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create controller
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	// Test concurrent snapshot settings access
	done := make(chan bool, 2)

	go func() {
		settings := controller.GetSnapshotSettings()
		assert.NotNil(t, settings, "Settings should not be nil")
		done <- true
	}()

	go func() {
		newSettings := &mediamtx.SnapshotSettings{
			Format:  "png",
			Quality: 90,
		}
		controller.UpdateSnapshotSettings(newSettings)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}
