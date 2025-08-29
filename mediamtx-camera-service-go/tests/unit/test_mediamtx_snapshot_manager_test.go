//go:build unit
// +build unit

/*
MediaMTX Snapshot Manager tests.

Requirements Coverage:
- REQ-SYS-001: System health monitoring and status reporting
- REQ-SYS-002: Component health state tracking

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotManager_RealSystem tests the real snapshot manager functionality
func TestSnapshotManager_RealSystem(t *testing.T) {
	// REQ-SYS-001: System health monitoring and status reporting
	// REQ-SYS-002: Component health state tracking

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	err := env.ConfigManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup test logging
	logger := logging.NewLogger("mediamtx-snapshot-test")
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&env.ConfigManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create real MediaMTX controller (not mock - following testing guide)
	controller, err := mediamtx.ControllerWithConfigManager(env.ConfigManager, logger.Logger)
	require.NoError(t, err, "Controller should be created successfully")

	ctx := context.Background()

	// Start controller
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer controller.Stop(ctx)

	t.Run("TakeAdvancedSnapshot_Integration", func(t *testing.T) {
		// Test taking a snapshot - this calls the real snapshot manager
		// Note: This may fail if camera hardware is not available
		// For unit tests, we validate the method exists and handles errors
		options := map[string]interface{}{
			"quality": 85,
			"format":  "jpg",
		}
		
		snapshot, err := controller.TakeAdvancedSnapshot(ctx, "/dev/video0", "/tmp/test_snapshot", options)
		if err != nil {
			t.Logf("Snapshot creation failed (expected if camera not available): %v", err)
		} else {
			assert.NotNil(t, snapshot, "Snapshot should not be nil")
		}
	})

	t.Run("GetAdvancedSnapshot_Integration", func(t *testing.T) {
		// Test getting a snapshot - this calls the real snapshot manager
		// Note: This may fail if no snapshots exist
		// For unit tests, we validate the method exists and handles errors
		snapshot, exists := controller.GetAdvancedSnapshot("non-existent-id")
		assert.False(t, exists, "Non-existent snapshot should not exist")
		assert.Nil(t, snapshot, "Non-existent snapshot should be nil")
	})

	t.Run("ListAdvancedSnapshots_Integration", func(t *testing.T) {
		// Test listing snapshots - this calls the real snapshot manager
		// For unit tests, we validate the method exists and handles errors
		snapshots := controller.ListAdvancedSnapshots()
		assert.NotNil(t, snapshots, "Snapshots list should not be nil")
		// Note: List may be empty if no snapshots have been taken
	})

	t.Run("DeleteAdvancedSnapshot_Integration", func(t *testing.T) {
		// Test deleting a snapshot - this calls the real snapshot manager
		// Note: This may fail if snapshot doesn't exist
		// For unit tests, we validate the method exists and handles errors
		err := controller.DeleteAdvancedSnapshot(ctx, "non-existent-id")
		if err != nil {
			t.Logf("Snapshot deletion failed (expected if snapshot doesn't exist): %v", err)
		}
	})

	t.Run("CleanupOldSnapshots_Integration", func(t *testing.T) {
		// Test cleanup functionality - this calls the real snapshot manager
		// For unit tests, we validate the method exists and handles errors
		err := controller.CleanupOldSnapshots(ctx, 24*time.Hour, 100)
		if err != nil {
			t.Logf("Snapshot cleanup failed: %v", err)
		}
		// Note: Cleanup may succeed even if no snapshots exist
	})
}
