/*
Mock CameraMonitor implementation for testing.

Provides a fully controllable mock implementation of the CameraMonitor interface
for unit testing WebSocket server and other components that depend on camera monitoring.

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration (mock)
- REQ-CAM-002: Real-time device status monitoring (mock)
- REQ-CAM-003: Device capability probing and format detection (mock)

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package utils

import (
	"context"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
)

// MockCameraMonitor provides a mock implementation of the CameraMonitor interface
// for testing purposes with full control over behavior and state
type MockCameraMonitor struct {
	// Mock state
	running          bool
	connectedCameras map[string]*camera.CameraDevice
	deviceStates     map[string]*camera.CameraDevice
	monitorStats     *camera.MonitorStats
	eventHandlers    []camera.CameraEventHandler
	eventCallbacks   []func(camera.CameraEventData)

	// Mock behavior control
	startError error
	stopError  error
	startDelay time.Duration
	stopDelay  time.Duration

	// Mock data
	mockDevices map[string]*camera.CameraDevice
	mockStats   *camera.MonitorStats

	// Thread safety
	mu sync.RWMutex
}

// NewMockCameraMonitor creates a new mock camera monitor for testing
func NewMockCameraMonitor() *MockCameraMonitor {
	return &MockCameraMonitor{
		connectedCameras: make(map[string]*camera.CameraDevice),
		deviceStates:     make(map[string]*camera.CameraDevice),
		eventHandlers:    make([]camera.CameraEventHandler, 0),
		eventCallbacks:   make([]func(camera.CameraEventData), 0),
		mockDevices:      make(map[string]*camera.CameraDevice),
		mockStats: &camera.MonitorStats{
			Running:                    false,
			ActiveTasks:                0,
			PollingCycles:              0,
			DeviceStateChanges:         0,
			CapabilityProbesAttempted:  0,
			CapabilityProbesSuccessful: 0,
			CapabilityTimeouts:         0,
			CapabilityParseErrors:      0,
			PollingFailureCount:        0,
			CurrentPollInterval:        1.0,
			KnownDevicesCount:          0,
			UdevEventsProcessed:        0,
			UdevEventsFiltered:         0,
			UdevEventsSkipped:          0,
		},
	}
}

// SetMockDevices sets mock camera devices for testing
func (m *MockCameraMonitor) SetMockDevices(devices map[string]*camera.CameraDevice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockDevices = devices
	m.connectedCameras = devices
	m.mockStats.KnownDevicesCount = len(devices)
}

// SetMockStats sets mock monitoring statistics for testing
func (m *MockCameraMonitor) SetMockStats(stats *camera.MonitorStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockStats = stats
}

// SetStartError sets an error to be returned when Start is called
func (m *MockCameraMonitor) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startError = err
}

// SetStopError sets an error to be returned when Stop is called
func (m *MockCameraMonitor) SetStopError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopError = err
}

// SetStartDelay sets a delay when Start is called
func (m *MockCameraMonitor) SetStartDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startDelay = delay
}

// SetStopDelay sets a delay when Stop is called
func (m *MockCameraMonitor) SetStopDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopDelay = delay
}

// AddMockDevice adds a mock camera device for testing
func (m *MockCameraMonitor) AddMockDevice(devicePath string, device *camera.CameraDevice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockDevices[devicePath] = device
	m.connectedCameras[devicePath] = device
	m.mockStats.KnownDevicesCount = len(m.mockDevices)
}

// RemoveMockDevice removes a mock camera device for testing
func (m *MockCameraMonitor) RemoveMockDevice(devicePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.mockDevices, devicePath)
	delete(m.connectedCameras, devicePath)
	m.mockStats.KnownDevicesCount = len(m.mockDevices)
}

// TriggerMockEvent triggers a mock camera event for testing
func (m *MockCameraMonitor) TriggerMockEvent(eventData camera.CameraEventData) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Notify event handlers
	for _, handler := range m.eventHandlers {
		handler.HandleCameraEvent(context.Background(), eventData)
	}

	// Notify event callbacks
	for _, callback := range m.eventCallbacks {
		callback(eventData)
	}
}

// Interface implementation methods

// Start starts the mock camera monitor
func (m *MockCameraMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}

	if m.startError != nil {
		return m.startError
	}

	m.running = true
	m.mockStats.Running = true
	return nil
}

// Stop stops the mock camera monitor
func (m *MockCameraMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopDelay > 0 {
		time.Sleep(m.stopDelay)
	}

	if m.stopError != nil {
		return m.stopError
	}

	m.running = false
	m.mockStats.Running = false
	return nil
}

// IsRunning returns whether the mock camera monitor is running
func (m *MockCameraMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetConnectedCameras returns the mock connected cameras
func (m *MockCameraMonitor) GetConnectedCameras() map[string]*camera.CameraDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*camera.CameraDevice)
	for k, v := range m.connectedCameras {
		result[k] = v
	}
	return result
}

// GetDevice returns a specific mock camera device
func (m *MockCameraMonitor) GetDevice(devicePath string) (*camera.CameraDevice, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, exists := m.connectedCameras[devicePath]
	return device, exists
}

// GetMonitorStats returns the mock monitoring statistics
func (m *MockCameraMonitor) GetMonitorStats() *camera.MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	stats := *m.mockStats
	return &stats
}

// AddEventHandler adds a camera event handler
func (m *MockCameraMonitor) AddEventHandler(handler camera.CameraEventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)
}

// AddEventCallback adds a camera event callback function
func (m *MockCameraMonitor) AddEventCallback(callback func(camera.CameraEventData)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventCallbacks = append(m.eventCallbacks, callback)
}

// Verify interface compliance
var _ camera.CameraMonitor = (*MockCameraMonitor)(nil)
