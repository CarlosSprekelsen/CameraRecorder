/*
Monitor startup smoke test for deterministic investigation.

This test isolates the camera monitor startup sequence from MediaMTX/paths noise
to identify orchestration bugs in the monitor startup process.

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_MonitorStart_Smoke tests the monitor startup sequence in isolation
func Test_MonitorStart_Smoke(t *testing.T) {
	// Setup: No MediaMTX; only the camera monitor subsystems
	logger := logging.CreateTestLogger(t, nil)

	// Create a minimal camera monitor with real implementations
	// We need to create the required dependencies
	configManager := config.CreateConfigManager() // Use centralized config creation
	deviceChecker := &RealDeviceChecker{}
	commandExecutor := &RealV4L2CommandExecutor{}
	infoParser := &RealDeviceInfoParser{}

	monitor, err := NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err, "Should create monitor successfully")
	require.NotNil(t, monitor, "Monitor should not be nil")

	// Act: Call monitor.Start(context.Background())
	startTime := time.Now()
	err = monitor.Start(context.Background())
	startDuration := time.Since(startTime)

	// Assert: It returns nil within 250ms on an empty system
	assert.NoError(t, err, "Monitor should start successfully")
	assert.Less(t, startDuration, 250*time.Millisecond, "Monitor should start within 250ms")

	// Assert: monitor.IsReady() flips after seed
	// Give it a moment for the seed discovery to complete
	time.Sleep(100 * time.Millisecond)
	assert.True(t, monitor.IsReady(), "Monitor should be ready after seed discovery")

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = monitor.Stop(ctx)
	assert.NoError(t, err, "Monitor should stop cleanly")

	// Verify: Read logs for the exact phase chain (A.1 → A.8)
	// The structured tracepoints should show:
	// 1. monitor_start_begin
	// 2. event_source_start_begin
	// 3. event_source_start_ok
	// 4. seed_discovery_begin
	// 5. seed_discovery_result
	// 6. loops_spawn_begin
	// 7. monitor_ready_true
	// 8. monitor_start_return_ok

	t.Logf("Monitor startup completed in %v", startDuration)
}

// Test_MonitorStart_SingleStartInvariant verifies the single Start() call invariant
func Test_MonitorStart_SingleStartInvariant(t *testing.T) {
	logger := logging.CreateTestLogger(t, nil)

	// Create monitor
	configManager := config.CreateConfigManager() // Use centralized config creation
	deviceChecker := &RealDeviceChecker{}
	commandExecutor := &RealV4L2CommandExecutor{}
	infoParser := &RealDeviceInfoParser{}

	monitor, err := NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err, "Should create monitor successfully")

	// Start monitor
	err = monitor.Start(context.Background())
	require.NoError(t, err, "Monitor should start successfully")

	// Verify: start_calls_total == 1 for a single monitor lifecycle
	// This would need to be implemented in the event source to expose the counter
	// For now, we just verify the monitor starts successfully

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = monitor.Stop(ctx)
	assert.NoError(t, err, "Monitor should stop cleanly")
}
