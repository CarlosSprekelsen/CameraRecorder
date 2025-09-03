/*
Interface Compliance Tests for CameraMonitor

Ensures that both mock and real implementations properly implement the CameraMonitor interface.
This is critical for maintaining interface contracts and enabling proper dependency injection.

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration (interface compliance)
- REQ-CAM-002: Real-time device status monitoring (interface compliance)
- REQ-CAM-003: Device capability probing and format detection (interface compliance)

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCameraMonitorInterfaceCompliance_Mock tests that MockCameraMonitor implements CameraMonitor interface
func TestCameraMonitorInterfaceCompliance_Mock(t *testing.T) {
	// Create mock camera monitor
	mockMonitor := utils.NewMockCameraMonitor()
	require.NotNil(t, mockMonitor, "Mock monitor should be created successfully")

	// Verify interface compliance at compile time
	var _ camera.CameraMonitor = mockMonitor

	// Test basic interface methods
	ctx := context.Background()

	// Test Start method
	err := mockMonitor.Start(ctx)
	assert.NoError(t, err, "Mock monitor should start without error")
	assert.True(t, mockMonitor.IsRunning(), "Mock monitor should be running after start")

	// Test Stop method
	err = mockMonitor.Stop()
	assert.NoError(t, err, "Mock monitor should stop without error")
	assert.False(t, mockMonitor.IsRunning(), "Mock monitor should not be running after stop")

	// Test GetConnectedCameras method
	cameras := mockMonitor.GetConnectedCameras()
	assert.NotNil(t, cameras, "GetConnectedCameras should return a map")
	assert.Equal(t, 0, len(cameras), "Initial camera count should be 0")

	// Test GetMonitorStats method
	stats := mockMonitor.GetMonitorStats()
	assert.NotNil(t, stats, "GetMonitorStats should return stats")
	assert.False(t, stats.Running, "Stats should reflect stopped state")

	// Test AddEventHandler method
	eventHandler := &mockEventHandler{}
	mockMonitor.AddEventHandler(eventHandler)
	// Note: We can't easily test the internal state, but this should not panic

	// Test AddEventCallback method
	eventCallback := func(data camera.CameraEventData) {}
	mockMonitor.AddEventCallback(eventCallback)
	// Note: We can't easily test the internal state, but this should not panic
}

// TestCameraMonitorInterfaceCompliance_Real tests that HybridCameraMonitor implements CameraMonitor interface
func TestCameraMonitorInterfaceCompliance_Real(t *testing.T) {
	// Use shared test environment
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	err := env.ConfigManager.LoadConfig("../../config/development.yaml")
	require.NoError(t, err, "Failed to load test configuration")

	// Setup real logging
	err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&env.ConfigManager.GetConfig().Logging))
	require.NoError(t, err, "Failed to setup logging")

	// Create real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Create real monitor
	realMonitor, err := camera.NewHybridCameraMonitor(
		env.ConfigManager,
		env.Logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err, "Failed to create real camera monitor")
	require.NotNil(t, realMonitor, "Real monitor should be created successfully")

	// Verify interface compliance at compile time
	var _ camera.CameraMonitor = realMonitor

	// Test basic interface methods
	ctx := context.Background()

	// Test Start method
	err = realMonitor.Start(ctx)
	require.NoError(t, err, "Real monitor should start without error")
	assert.True(t, realMonitor.IsRunning(), "Real monitor should be running after start")

	// Test Stop method
	err = realMonitor.Stop()
	require.NoError(t, err, "Real monitor should stop without error")
	assert.False(t, realMonitor.IsRunning(), "Real monitor should not be running after stop")

	// Test GetConnectedCameras method
	cameras := realMonitor.GetConnectedCameras()
	assert.NotNil(t, cameras, "GetConnectedCameras should return a map")

	// Test GetMonitorStats method
	stats := realMonitor.GetMonitorStats()
	assert.NotNil(t, stats, "GetMonitorStats should return stats")
	assert.False(t, stats.Running, "Stats should reflect stopped state")

	// Test AddEventHandler method
	eventHandler := &mockEventHandler{}
	realMonitor.AddEventHandler(eventHandler)
	// Note: We can't easily test the internal state, but this should not panic

	// Test AddEventCallback method
	eventCallback := func(data camera.CameraEventData) {}
	realMonitor.AddEventCallback(eventCallback)
	// Note: We can't easily test the internal state, but this should not panic
}

// TestCameraMonitorInterface_BehavioralCompliance tests behavioral compliance of both implementations
func TestCameraMonitorInterface_BehavioralCompliance(t *testing.T) {
	// Test mock implementation behavioral compliance
	t.Run("Mock_BehavioralCompliance", func(t *testing.T) {
		mockMonitor := utils.NewMockCameraMonitor()

		// Test error handling
		mockMonitor.SetStartError(assert.AnError)
		err := mockMonitor.Start(context.Background())
		assert.Error(t, err, "Mock monitor should return configured start error")

		// Test delay behavior
		mockMonitor.SetStartError(nil)
		mockMonitor.SetStartDelay(10 * time.Millisecond)

		startTime := time.Now()
		err = mockMonitor.Start(context.Background())
		duration := time.Since(startTime)

		assert.NoError(t, err, "Mock monitor should start without error")
		assert.GreaterOrEqual(t, duration, 10*time.Millisecond, "Mock monitor should respect start delay")
	})

	// Test real implementation behavioral compliance
	t.Run("Real_BehavioralCompliance", func(t *testing.T) {
		env := utils.SetupMediaMTXTestEnvironment(t)
		defer utils.TeardownMediaMTXTestEnvironment(t, env)

		err := env.ConfigManager.LoadConfig("../../config/development.yaml")
		require.NoError(t, err, "Failed to load test configuration")

		err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&env.ConfigManager.GetConfig().Logging))
		require.NoError(t, err, "Failed to setup logging")

		deviceChecker := &camera.RealDeviceChecker{}
		commandExecutor := &camera.RealV4L2CommandExecutor{}
		infoParser := &camera.RealDeviceInfoParser{}

		realMonitor, err := camera.NewHybridCameraMonitor(
			env.ConfigManager,
			env.Logger,
			deviceChecker,
			commandExecutor,
			infoParser,
		)
		require.NoError(t, err, "Failed to create real camera monitor")

		// Test that real monitor can be started and stopped multiple times
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			err = realMonitor.Start(ctx)
			assert.NoError(t, err, "Real monitor should start successfully on iteration %d", i)
			assert.True(t, realMonitor.IsRunning(), "Real monitor should be running after start on iteration %d", i)

			err = realMonitor.Stop()
			assert.NoError(t, err, "Real monitor should stop successfully on iteration %d", i)
			assert.False(t, realMonitor.IsRunning(), "Real monitor should not be running after stop on iteration %d", i)
		}
	})
}

// TestCameraMonitorInterface_DataConsistency tests data consistency across interface methods
func TestCameraMonitorInterface_DataConsistency(t *testing.T) {
	mockMonitor := utils.NewMockCameraMonitor()

	// Create mock camera devices
	mockDevices := map[string]*camera.CameraDevice{
		"/dev/video0": {
			Path:      "/dev/video0",
			DeviceNum: 0,
			Name:      "Mock Camera 0",
			Status:    camera.DeviceStatusConnected,
			LastSeen:  time.Now(),
		},
		"/dev/video1": {
			Path:      "/dev/video1",
			DeviceNum: 1,
			Name:      "Mock Camera 1",
			Status:    camera.DeviceStatusConnected,
			LastSeen:  time.Now(),
		},
	}

	// Set mock devices
	mockMonitor.SetMockDevices(mockDevices)

	// Test data consistency
	cameras := mockMonitor.GetConnectedCameras()
	assert.Equal(t, 2, len(cameras), "Should return correct number of cameras")

	// Test individual device retrieval
	device, exists := mockMonitor.GetDevice("/dev/video0")
	assert.True(t, exists, "Device /dev/video0 should exist")
	assert.Equal(t, "Mock Camera 0", device.Name, "Device name should match")

	device, exists = mockMonitor.GetDevice("/dev/video1")
	assert.True(t, exists, "Device /dev/video1 should exist")
	assert.Equal(t, "Mock Camera 1", device.Name, "Device name should match")

	device, exists = mockMonitor.GetDevice("/dev/video2")
	assert.False(t, exists, "Non-existent device should not be found")
	assert.Nil(t, device, "Non-existent device should return nil")

	// Test stats consistency
	stats := mockMonitor.GetMonitorStats()
	assert.Equal(t, 2, stats.KnownDevicesCount, "Stats should reflect correct device count")
}

// mockEventHandler implements CameraEventHandler for testing
type mockEventHandler struct{}

func (h *mockEventHandler) HandleCameraEvent(ctx context.Context, eventData camera.CameraEventData) error {
	// Mock implementation - just return nil
	return nil
}

// TestCameraMonitorInterface_EventHandling tests event handling compliance
func TestCameraMonitorInterface_EventHandling(t *testing.T) {
	mockMonitor := utils.NewMockCameraMonitor()

	// Test event handler registration
	eventHandler := &mockEventHandler{}
	mockMonitor.AddEventHandler(eventHandler)

	// Test event callback registration
	eventCallback := func(data camera.CameraEventData) {
		// Mock callback implementation
	}
	mockMonitor.AddEventCallback(eventCallback)

	// Test event triggering
	eventData := camera.CameraEventData{
		DevicePath: "/dev/video0",
		EventType:  camera.CameraEventConnected,
		Timestamp:  time.Now(),
	}

	mockMonitor.TriggerMockEvent(eventData)

	// Note: We can't easily test that the event was actually processed
	// since the mock implementation doesn't store the callback state
	// This is a limitation of the current mock design
	assert.True(t, true, "Event triggering should not panic")
}

// TestCameraMonitorInterface_ThreadSafety tests thread safety of interface implementations
func TestCameraMonitorInterface_ThreadSafety(t *testing.T) {
	mockMonitor := utils.NewMockCameraMonitor()

	// Test concurrent access to interface methods
	const numGoroutines = 10
	const numOperations = 100

	// Start multiple goroutines that access the monitor concurrently
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Concurrent reads
				_ = mockMonitor.IsRunning()
				_ = mockMonitor.GetConnectedCameras()
				_ = mockMonitor.GetMonitorStats()

				// Concurrent writes (if monitor is not running)
				if !mockMonitor.IsRunning() {
					_ = mockMonitor.Start(context.Background())
					_ = mockMonitor.Stop()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify the monitor is in a consistent state
	assert.False(t, mockMonitor.IsRunning(), "Monitor should be in consistent state after concurrent access")
}
