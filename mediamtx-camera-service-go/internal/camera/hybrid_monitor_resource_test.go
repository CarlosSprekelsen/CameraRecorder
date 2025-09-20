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
	logger := logging.CreateTestLogger(t, nil)

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
	logger := logging.CreateTestLogger(t, nil)

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

	// Add test event handler
	var handlerCalled int32
	handler := &TestEventHandler{
		callback: func(ctx context.Context, eventData CameraEventData) error {
			atomic.AddInt32(&handlerCalled, 1)
			time.Sleep(DefaultPollInterval) // Brief work
			return nil
		},
	}

	monitor.AddEventHandler(handler)

	// Verify handler is tracked
	stats := monitor.GetResourceStats()
	assert.Equal(t, 1, stats["active_event_handlers"].(int))

	// Generate test event using existing patterns
	testDevice := &CameraDevice{
		Path:   DefaultDevicePath,
		Name:   "Test Camera",
		Status: DeviceStatusConnected,
	}

	monitor.generateCameraEvent(ctx, CameraEventConnected, DefaultDevicePath, testDevice)

	// Wait for handler execution
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&handlerCalled) == 1
	}, DefaultTestTimeout, DefaultPollInterval)

	// Verify worker pool processed the event
	workerStats := monitor.eventWorkerPool.GetStats()
	assert.Equal(t, int64(1), workerStats.CompletedTasks)
}

func TestHybridMonitor_GracefulShutdownWithWorkerPool(t *testing.T) {
	// REQ-RESOURCE-001: Graceful shutdown with active worker pool

	// Use existing test pattern
	configManager := config.CreateConfigManager()
	logger := logging.CreateTestLogger(t, nil)

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

	// Add slow event handler
	var handlerStarted, handlerCompleted int32
	handler := &TestEventHandler{
		callback: func(ctx context.Context, eventData CameraEventData) error {
			atomic.StoreInt32(&handlerStarted, 1)
			time.Sleep(QuickTestTimeout)
			atomic.StoreInt32(&handlerCompleted, 1)
			return nil
		},
	}

	monitor.AddEventHandler(handler)

	// Generate event
	testDevice := &CameraDevice{
		Path:   DefaultDevicePath,
		Name:   "Test Camera",
		Status: DeviceStatusConnected,
	}

	monitor.generateCameraEvent(ctx, CameraEventConnected, DefaultDevicePath, testDevice)

	// Wait for handler to start
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&handlerStarted) == 1
	}, DefaultTestTimeout, DefaultPollInterval)

	// Stop monitor - should wait for handlers to complete
	stopCtx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout*2)
	defer cancel()

	err = monitor.Stop(stopCtx)
	require.NoError(t, err)

	// Verify handler completed
	assert.Equal(t, int32(1), atomic.LoadInt32(&handlerCompleted))
	assert.False(t, monitor.IsRunning())
	assert.False(t, monitor.eventWorkerPool.IsRunning())
}

func TestHybridMonitor_WorkerPoolFailureHandling(t *testing.T) {
	// REQ-RESOURCE-001: Worker pool failure handling and recovery

	// Use existing test pattern
	configManager := config.CreateConfigManager()
	logger := logging.CreateTestLogger(t, nil)

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

	// Add handler that panics
	handler := &TestEventHandler{
		callback: func(ctx context.Context, eventData CameraEventData) error {
			panic("test panic - should be handled by worker pool")
		},
	}

	// Get baseline statistics before adding handler (account for startup events)
	time.Sleep(QuickTestTimeout) // Wait for startup events to complete
	baselineStats := monitor.eventWorkerPool.GetStats()
	baselineFailures := baselineStats.FailedTasks

	monitor.AddEventHandler(handler)

	// Generate event
	testDevice := &CameraDevice{
		Path:   DefaultDevicePath,
		Name:   "Test Camera",
		Status: DeviceStatusConnected,
	}

	monitor.generateCameraEvent(ctx, CameraEventConnected, DefaultDevicePath, testDevice)

	// Wait for panic handling
	time.Sleep(QuickTestTimeout)

	// Verify system handled panic gracefully
	assert.True(t, monitor.IsRunning())
	assert.True(t, monitor.eventWorkerPool.IsRunning())

	// Verify worker pool recorded exactly 1 additional failure
	finalStats := monitor.eventWorkerPool.GetStats()
	expectedFailures := baselineFailures + 1
	assert.Equal(t, expectedFailures, finalStats.FailedTasks,
		"Should have exactly 1 additional failed task (baseline: %d, expected: %d)",
		baselineFailures, expectedFailures)
}

// TestEventHandler is a simple test implementation following existing patterns
type TestEventHandler struct {
	callback func(context.Context, CameraEventData) error
}

func (h *TestEventHandler) HandleCameraEvent(ctx context.Context, eventData CameraEventData) error {
	if h.callback != nil {
		return h.callback(ctx, eventData)
	}
	return nil
}
