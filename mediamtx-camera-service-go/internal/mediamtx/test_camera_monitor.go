/*
MediaMTX Test Camera Monitor Implementation

This file provides a test implementation of the camera.CameraMonitor interface
for MediaMTX unit tests. It follows the existing test patterns and maintains
compatibility with the real camera monitor interface.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
)

// TestCameraMonitor provides a test implementation of camera.CameraMonitor
// for MediaMTX unit tests. It simulates camera discovery and monitoring
// without requiring real camera hardware.
type TestCameraMonitor struct {
	// State management
	mu              sync.RWMutex
	isRunning       bool
	connectedCameras map[string]*camera.CameraDevice
	
	// Event handling
	eventHandlers  []camera.CameraEventHandler
	eventCallbacks []func(camera.CameraEventData)
	eventNotifier  camera.EventNotifier
	
	// Statistics
	stats *camera.MonitorStats
}

// NewTestCameraMonitor creates a new test camera monitor instance
func NewTestCameraMonitor() *TestCameraMonitor {
	return &TestCameraMonitor{
		connectedCameras: make(map[string]*camera.CameraDevice),
		eventHandlers:    make([]camera.CameraEventHandler, 0),
		eventCallbacks:   make([]func(camera.CameraEventData), 0),
		stats: &camera.MonitorStats{
			Running:                    false,
			ActiveTasks:                0,
			PollingCycles:              0,
			DeviceStateChanges:         0,
			CapabilityProbesAttempted:  0,
			CapabilityProbesSuccessful: 0,
			CapabilityTimeouts:         0,
			CapabilityParseErrors:      0,
			PollingFailureCount:        0,
			CurrentPollInterval:        5.0,
			KnownDevicesCount:          0,
			UdevEventsProcessed:        0,
			UdevEventsFiltered:         0,
			UdevEventsSkipped:          0,
		},
	}
}

// Start starts the test camera monitor
func (m *TestCameraMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.isRunning {
		return nil // Already running
	}
	
	m.isRunning = true
	m.stats.Running = true
	
	// Simulate camera discovery for testing
	m.simulateCameraDiscovery()
	
	return nil
}

// Stop stops the test camera monitor
func (m *TestCameraMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.isRunning {
		return nil // Already stopped
	}
	
	m.isRunning = false
	m.stats.Running = false
	
	return nil
}

// IsRunning returns true if the monitor is running
func (m *TestCameraMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// GetConnectedCameras returns a map of connected cameras
func (m *TestCameraMonitor) GetConnectedCameras() map[string]*camera.CameraDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to prevent external modification
	result := make(map[string]*camera.CameraDevice)
	for k, v := range m.connectedCameras {
		result[k] = v
	}
	return result
}

// GetDevice returns a specific camera device by path
func (m *TestCameraMonitor) GetDevice(devicePath string) (*camera.CameraDevice, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	device, exists := m.connectedCameras[devicePath]
	return device, exists
}

// GetMonitorStats returns monitoring statistics
func (m *TestCameraMonitor) GetMonitorStats() *camera.MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy to prevent external modification
	stats := *m.stats
	return &stats
}

// AddEventHandler adds a camera event handler
func (m *TestCameraMonitor) AddEventHandler(handler camera.CameraEventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.eventHandlers = append(m.eventHandlers, handler)
}

// AddEventCallback adds a camera event callback
func (m *TestCameraMonitor) AddEventCallback(callback func(camera.CameraEventData)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.eventCallbacks = append(m.eventCallbacks, callback)
}

// SetEventNotifier sets the event notifier
func (m *TestCameraMonitor) SetEventNotifier(notifier camera.EventNotifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.eventNotifier = notifier
}

// simulateCameraDiscovery simulates camera discovery for testing
func (m *TestCameraMonitor) simulateCameraDiscovery() {
	// Create test cameras for MediaMTX testing
	testCameras := map[string]*camera.CameraDevice{
		"camera0": {
			Path:   "/dev/video0",
			Name:   "Test Camera 0",
			Status: camera.DeviceStatusConnected,
			Capabilities: camera.V4L2Capabilities{
				CardName:     "Test Camera 0",
				DriverName:   "test_driver",
				BusInfo:      "usb-0000:00:14.0-1",
				Version:      "5.15.0",
				Capabilities: []string{"VIDEO_CAPTURE", "STREAMING"},
				DeviceCaps:   []string{"VIDEO_CAPTURE", "STREAMING"},
			},
			LastSeen: time.Now(),
		},
		"camera1": {
			Path:   "/dev/video1",
			Name:   "Test Camera 1",
			Status: camera.DeviceStatusConnected,
			Capabilities: camera.V4L2Capabilities{
				CardName:     "Test Camera 1",
				DriverName:   "test_driver",
				BusInfo:      "usb-0000:00:14.0-2",
				Version:      "5.15.0",
				Capabilities: []string{"VIDEO_CAPTURE", "STREAMING"},
				DeviceCaps:   []string{"VIDEO_CAPTURE", "STREAMING"},
			},
			LastSeen: time.Now(),
		},
	}
	
	// Add test cameras to connected cameras
	for path, device := range testCameras {
		m.connectedCameras[path] = device
	}
	
	// Update statistics
	m.stats.KnownDevicesCount = int64(len(m.connectedCameras))
	m.stats.DeviceStateChanges += int64(len(testCameras))
}

// AddTestCamera adds a test camera for specific test scenarios
func (m *TestCameraMonitor) AddTestCamera(devicePath string, device *camera.CameraDevice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.connectedCameras[devicePath] = device
	m.stats.KnownDevicesCount = int64(len(m.connectedCameras))
	m.stats.DeviceStateChanges++
}

// RemoveTestCamera removes a test camera for specific test scenarios
func (m *TestCameraMonitor) RemoveTestCamera(devicePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.connectedCameras, devicePath)
	m.stats.KnownDevicesCount = int64(len(m.connectedCameras))
	m.stats.DeviceStateChanges++
}

// ClearTestCameras clears all test cameras
func (m *TestCameraMonitor) ClearTestCameras() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.connectedCameras = make(map[string]*camera.CameraDevice)
	m.stats.KnownDevicesCount = 0
}
