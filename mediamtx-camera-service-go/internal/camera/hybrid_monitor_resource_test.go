/*
HybridCameraMonitor Resource Management Tests

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration
- REQ-CAM-002: Real-time device status monitoring
- REQ-RESOURCE-001: Bounded worker pool integration

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHybridMonitor_WorkerPoolIntegration(t *testing.T) {
	// REQ-RESOURCE-001: Worker pool integration with camera monitor

	// Use existing test pattern from hybrid_monitor_test.go
	configManager := config.CreateConfigManager()
	logger := logging.GetLoggerFactory().CreateLogger("test")

	// Create real implementations (following existing pattern)
	deviceChecker := &RealDeviceChecker{}
	commandExecutor := &RealV4L2CommandExecutor{}
	infoParser := &RealDeviceInfoParser{}

	// Create monitor with test config
	monitor, err := NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	// Test worker pool is initialized
	assert.NotNil(t, monitor.eventWorkerPool)

	// Start monitor
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop(ctx)

	// Verify worker pool is running
	assert.True(t, monitor.eventWorkerPool.IsRunning())

	// Test resource stats include worker pool information
	stats := monitor.GetResourceStats()
	assert.Contains(t, stats, "worker_pool")

	workerPoolStats := stats["worker_pool"].(map[string]interface{})
	assert.Contains(t, workerPoolStats, "max_workers")
	assert.Contains(t, workerPoolStats, "active_workers")
	assert.Contains(t, workerPoolStats, "completed_tasks")
}

func TestHybridMonitor_EventHandlerResourceManagement(t *testing.T) {
	// REQ-CAM-002: Event handler resource management

	// Use existing test pattern
	configManager := config.CreateConfigManager()
	logger := logging.GetLoggerFactory().CreateLogger("test")

	// Create real implementations
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
	require.NoError(t, err)

	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop(ctx)

	// Wait for monitor to be ready
	require.Eventually(t, func() bool {
		return monitor.IsReady()
	}, DefaultTestTimeout, DefaultPollInterval, "Monitor should be ready")

	// Add test event handler
	var handlerCalled int32
	handler := &WorkingTestEventHandler{
		started:      &handlerCalled,        // Reuse started as called counter
		completed:    &handlerCalled,        // Same counter
		workDuration: 50 * time.Millisecond, // Quick work
	}

	monitor.AddEventHandler(handler)

	// Give a moment for handler registration to complete
	time.Sleep(100 * time.Millisecond)

	// Verify handler is tracked in resource stats
	stats := monitor.GetResourceStats()
	activeHandlers, ok := stats["active_event_handlers"].(int)
	require.True(t, ok, "Should have active_event_handlers in stats")
	assert.Equal(t, 1, activeHandlers, "Should track active event handlers")

	// Test that graceful shutdown works with registered handlers
	stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = monitor.Stop(stopCtx)
	require.NoError(t, err, "Stop should complete without timeout")

	// Verify monitor is properly stopped
	assert.False(t, monitor.IsRunning(), "Monitor should not be running after stop")
	assert.False(t, monitor.eventWorkerPool.IsRunning(), "Worker pool should be stopped")

	t.Log("✅ Event handler resource management validated")
}

func TestHybridMonitor_GracefulShutdownWithWorkerPool(t *testing.T) {
	// REQ-RESOURCE-001: Graceful shutdown with active worker pool

	// Use existing test pattern
	configManager := config.CreateConfigManager()
	logger := logging.GetLoggerFactory().CreateLogger("test")

	// Create real implementations
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
	require.NoError(t, err)

	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// Wait for startup events to complete to ensure test isolation
	time.Sleep(QuickTestTimeout)

	// Add a working event handler that actually does work
	var handlerStarted, handlerCompleted int32
	handler := &WorkingTestEventHandler{
		started:      &handlerStarted,
		completed:    &handlerCompleted,
		workDuration: 200 * time.Millisecond, // Simulate some work
	}

	monitor.AddEventHandler(handler)

	// Generate event to trigger handler
	testDevice := &CameraDevice{
		Path:   DefaultDevicePath,
		Name:   "Test Camera",
		Status: DeviceStatusConnected,
	}

	monitor.generateCameraEvent(ctx, CameraEventConnected, DefaultDevicePath, testDevice)

	// Wait for handler to start processing
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&handlerStarted) == 1
	}, DefaultTestTimeout, DefaultPollInterval, "Handler should start processing")

	// Stop monitor with a reasonable timeout - should wait for handlers to complete
	stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Measure shutdown time to verify graceful behavior
	startTime := time.Now()
	err = monitor.Stop(stopCtx)
	shutdownDuration := time.Since(startTime)

	// Verify graceful shutdown behavior
	require.NoError(t, err, "Stop should complete without timeout")

	// Verify handler had time to complete its work (should take at least 200ms)
	// Allow some tolerance for timing variations
	assert.GreaterOrEqual(t, shutdownDuration, 150*time.Millisecond,
		"Shutdown should wait for handler to complete work (tolerance: 150ms)")

	// Verify handler completed its work
	assert.Equal(t, int32(1), atomic.LoadInt32(&handlerCompleted),
		"Handler should have completed its work during graceful shutdown")

	// Verify monitor is properly stopped
	assert.False(t, monitor.IsRunning(), "Monitor should not be running after stop")
	assert.False(t, monitor.eventWorkerPool.IsRunning(), "Worker pool should be stopped")

	t.Logf("✅ Graceful shutdown completed in %v (handler work: %v)",
		shutdownDuration, handler.workDuration)
}

func TestHybridMonitor_WorkerPoolFailureHandling(t *testing.T) {
	// REQ-RESOURCE-001: Worker pool failure handling and recovery

	// Use existing test pattern
	configManager := config.CreateConfigManager()
	logger := logging.GetLoggerFactory().CreateLogger("test")

	// Create real implementations
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
	require.NoError(t, err)

	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)
	defer monitor.Stop(ctx)

	// Wait for monitor to be ready
	require.Eventually(t, func() bool {
		return monitor.IsReady()
	}, DefaultTestTimeout, DefaultPollInterval, "Monitor should be ready")

	// Add handler that returns an error to test failure handling
	var handlerCalled int32
	handler := &FailingTestEventHandler{
		started:   &handlerCalled,
		completed: &handlerCalled,
	}

	monitor.AddEventHandler(handler)

	// Generate event
	testDevice := &CameraDevice{
		Path:   DefaultDevicePath,
		Name:   "Test Camera",
		Status: DeviceStatusConnected,
	}

	monitor.generateCameraEvent(ctx, CameraEventConnected, DefaultDevicePath, testDevice)

	// Wait for handler to be called (even if it fails)
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&handlerCalled) == 1
	}, DefaultTestTimeout, DefaultPollInterval, "Handler should be called")

	// Verify system handled failure gracefully - the key test
	assert.True(t, monitor.IsRunning(), "Monitor should still be running after handler failure")
	assert.True(t, monitor.eventWorkerPool.IsRunning(), "Worker pool should still be running")

	// Test graceful shutdown with failed handler
	stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = monitor.Stop(stopCtx)
	require.NoError(t, err, "Stop should complete without timeout even with failed handler")

	// Verify monitor is properly stopped
	assert.False(t, monitor.IsRunning(), "Monitor should not be running after stop")
	assert.False(t, monitor.eventWorkerPool.IsRunning(), "Worker pool should be stopped")

	t.Log("✅ Worker pool failure handling validated - system remains stable")
}

// WorkingTestEventHandler is a test handler that actually does work to test graceful shutdown
type WorkingTestEventHandler struct {
	started      *int32
	completed    *int32
	workDuration time.Duration
}

func (h *WorkingTestEventHandler) HandleCameraEvent(ctx context.Context, eventData CameraEventData) error {
	// Mark handler as started
	atomic.StoreInt32(h.started, 1)

	// Simulate some work that takes time
	select {
	case <-time.After(h.workDuration):
		// Work completed successfully
		atomic.StoreInt32(h.completed, 1)
		return nil
	case <-ctx.Done():
		// Context was cancelled - this is expected during shutdown
		return ctx.Err()
	}
}

// FailingTestEventHandler is a test handler that always returns an error to test failure handling
type FailingTestEventHandler struct {
	started   *int32
	completed *int32
}

func (h *FailingTestEventHandler) HandleCameraEvent(ctx context.Context, eventData CameraEventData) error {
	// Mark handler as started
	atomic.StoreInt32(h.started, 1)

	// Always return an error to test failure handling
	atomic.StoreInt32(h.completed, 1)
	return fmt.Errorf("test handler failure")
}
