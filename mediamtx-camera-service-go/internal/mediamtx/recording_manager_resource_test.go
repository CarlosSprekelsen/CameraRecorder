/*
RecordingManager Resource Management Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-RESOURCE-001: Resource lifecycle management

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordingManager_ResourceLifecycle(t *testing.T) {
	// REQ-RESOURCE-001: Resource lifecycle management

	// Use existing test helper pattern
	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()

	// Test initial state
	assert.False(t, rm.IsRunning())

	// Test start
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest
	err := rm.Start(ctx)
	require.NoError(t, err)
	assert.True(t, rm.IsRunning())

	// Test resource stats are available
	stats := rm.GetResourceStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "running")
	assert.Contains(t, stats, "active_keepalive_readers")
	assert.Contains(t, stats, "active_timers")
	assert.True(t, stats["running"].(bool))

	// Test stop
	err = rm.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, rm.IsRunning())
}

func TestRecordingManager_ResourceCleanup(t *testing.T) {
	// REQ-RESOURCE-001: Resource cleanup with active recordings

	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest

	err := rm.Start(ctx)
	require.NoError(t, err)
	defer rm.Stop(ctx)

	// Create active timers to simulate recordings
	rm.timerManager.CreateTimer("camera0", "/dev/video0", 0, func() {})
	rm.timerManager.CreateTimer("camera1", "/dev/video1", 0, func() {})

	// Verify timers exist
	assert.True(t, rm.timerManager.IsRecording("camera0"))
	assert.True(t, rm.timerManager.IsRecording("camera1"))

	// Perform cleanup
	err = rm.Cleanup(ctx)
	require.NoError(t, err)

	// Verify all resources cleaned up
	assert.False(t, rm.timerManager.IsRecording("camera0"))
	assert.False(t, rm.timerManager.IsRecording("camera1"))

	// Verify stats updated
	stats := rm.GetResourceStats()
	assert.Equal(t, int64(2), stats["total_recordings_stopped"].(int64))
}

func TestRecordingManager_StatsTracking(t *testing.T) {
	// REQ-RESOURCE-001: Resource statistics tracking

	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest

	err := rm.Start(ctx)
	require.NoError(t, err)
	defer rm.Stop(ctx)

	// Simulate recording operations
	rm.updateRecordingStats(true, false)  // Successful start
	rm.updateRecordingStats(false, false) // Successful stop
	rm.updateRecordingStats(true, true)   // Error case

	// Verify stats
	stats := rm.GetResourceStats()
	assert.Equal(t, int64(2), stats["total_recordings_started"].(int64))
	assert.Equal(t, int64(1), stats["total_recordings_stopped"].(int64))
	assert.Equal(t, int64(1), stats["recording_errors"].(int64))
}

func TestRecordingManager_KeepaliveIntegration(t *testing.T) {
	// REQ-MTX-002: RTSP keepalive integration

	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest

	err := rm.Start(ctx)
	require.NoError(t, err)
	defer rm.Stop(ctx)

	// Test keepalive reader is properly initialized and configured
	assert.NotNil(t, rm.keepaliveReader)

	// Test resource stats include keepalive information
	stats := rm.GetResourceStats()
	assert.Contains(t, stats, "active_keepalive_readers")
	assert.Equal(t, int64(0), stats["active_keepalive_readers"].(int64))
}

func TestRecordingManager_TimerIntegration(t *testing.T) {
	// REQ-RESOURCE-001: Timer management integration

	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest

	err := rm.Start(ctx)
	require.NoError(t, err)
	defer rm.Stop(ctx)

	// Test timer creation and management
	rm.timerManager.CreateTimer("test_camera", "/dev/video0", TestTimeoutShort, func() {})

	// Verify timer exists and is tracked in stats
	assert.True(t, rm.timerManager.IsRecording("test_camera"))

	stats := rm.GetResourceStats()
	assert.Equal(t, int64(1), stats["active_timers"].(int64))

	// Wait for timer to expire
	time.Sleep(TestTimeoutShort + TestTimeoutShort/2)

	// Verify timer auto-cleanup
	assert.False(t, rm.timerManager.IsRecording("test_camera"))
}

func TestRecordingManager_GracefulShutdown(t *testing.T) {
	// REQ-RESOURCE-001: Graceful shutdown with active resources

	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper, ctx := SetupMediaMTXTest(t)

	rm := helper.GetRecordingManager()
	// MINIMAL: Helper provides standard context
	// Context already provided by SetupMediaMTXTest

	err := rm.Start(ctx)
	require.NoError(t, err)

	// Create active resources
	rm.timerManager.CreateTimer("active_camera", "/dev/video0", 0, func() {})

	// Verify resource is active
	assert.True(t, rm.timerManager.IsRecording("active_camera"))

	// Stop should clean up all resources
	stopCtx, cancel := context.WithTimeout(context.Background(), TestTimeoutLong)
	defer cancel()

	err = rm.Stop(stopCtx)
	require.NoError(t, err)

	// Verify all resources cleaned up
	assert.False(t, rm.timerManager.IsRecording("active_camera"))
	assert.False(t, rm.IsRunning())
}
