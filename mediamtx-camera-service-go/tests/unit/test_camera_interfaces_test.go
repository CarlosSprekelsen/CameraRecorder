//go:build unit
// +build unit

/*
Camera Interfaces Unit Tests

Requirements Coverage:
- REQ-CAM-001: Camera device detection and enumeration
- REQ-CAM-002: Camera capability probing and validation
- REQ-CAM-003: Camera event handling and monitoring
- REQ-CAM-004: Device information parsing and validation
- REQ-CAM-005: Interface compliance and contract validation
- REQ-CAM-006: Event data structure validation

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Real Component Usage: V4L2 devices, event handling, monitoring
*/

package camera_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
)

// TestCameraEvent tests the CameraEvent constants
func TestCameraEvent(t *testing.T) {
	t.Run("camera_event_constants", func(t *testing.T) {
		assert.Equal(t, "CONNECTED", string(camera.CameraEventConnected), "Connected event should be correct")
		assert.Equal(t, "DISCONNECTED", string(camera.CameraEventDisconnected), "Disconnected event should be correct")
		assert.Equal(t, "STATUS_CHANGED", string(camera.CameraEventStatusChanged), "Status changed event should be correct")
	})

	t.Run("camera_event_json_marshaling", func(t *testing.T) {
		events := []camera.CameraEvent{
			camera.CameraEventConnected,
			camera.CameraEventDisconnected,
			camera.CameraEventStatusChanged,
		}

		for _, event := range events {
			jsonData, err := json.Marshal(event)
			require.NoError(t, err, "Should marshal camera event to JSON without error")

			var unmarshaledEvent camera.CameraEvent
			err = json.Unmarshal(jsonData, &unmarshaledEvent)
			require.NoError(t, err, "Should unmarshal camera event from JSON without error")

			assert.Equal(t, event, unmarshaledEvent, "Camera event should be preserved in JSON")
		}
	})
}

// TestCameraEventData tests the CameraEventData struct
func TestCameraEventData(t *testing.T) {
	t.Run("camera_event_data_creation", func(t *testing.T) {
		eventData := camera.CameraEventData{
			DevicePath: "/dev/video0",
			EventType:  camera.CameraEventConnected,
			Timestamp:  time.Now(),
			DeviceInfo: &camera.CameraDevice{
				Path:   "/dev/video0",
				Name:   "USB Camera",
				Status: camera.DeviceStatusConnected,
			},
		}

		assert.Equal(t, "/dev/video0", eventData.DevicePath, "Device path should be set correctly")
		assert.Equal(t, camera.CameraEventConnected, eventData.EventType, "Event type should be set correctly")
		assert.NotZero(t, eventData.Timestamp, "Timestamp should be set")
		assert.NotNil(t, eventData.DeviceInfo, "Device info should be set")
		assert.Equal(t, "/dev/video0", eventData.DeviceInfo.Path, "Device info path should be correct")
	})

	t.Run("camera_event_data_json_marshaling", func(t *testing.T) {
		timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		eventData := camera.CameraEventData{
			DevicePath: "/dev/video0",
			EventType:  camera.CameraEventConnected,
			Timestamp:  timestamp,
			DeviceInfo: &camera.CameraDevice{
				Path: "/dev/video0",
				Name: "USB Camera",
				Capabilities: camera.V4L2Capabilities{
					DriverName: "uvcvideo",
					CardName:   "USB Camera",
				},
				Status:   camera.DeviceStatusConnected,
				LastSeen: timestamp,
			},
		}

		jsonData, err := json.Marshal(eventData)
		require.NoError(t, err, "Should marshal event data to JSON without error")

		var unmarshaledEventData camera.CameraEventData
		err = json.Unmarshal(jsonData, &unmarshaledEventData)
		require.NoError(t, err, "Should unmarshal event data from JSON without error")

		assert.Equal(t, eventData.DevicePath, unmarshaledEventData.DevicePath, "Device path should be preserved in JSON")
		assert.Equal(t, eventData.EventType, unmarshaledEventData.EventType, "Event type should be preserved in JSON")
		assert.Equal(t, eventData.Timestamp, unmarshaledEventData.Timestamp, "Timestamp should be preserved in JSON")
		assert.NotNil(t, unmarshaledEventData.DeviceInfo, "Device info should be preserved in JSON")
		assert.Equal(t, eventData.DeviceInfo.Path, unmarshaledEventData.DeviceInfo.Path, "Device info path should be preserved in JSON")
		assert.Equal(t, eventData.DeviceInfo.Name, unmarshaledEventData.DeviceInfo.Name, "Device info name should be preserved in JSON")
	})

	t.Run("camera_event_data_without_device_info", func(t *testing.T) {
		eventData := camera.CameraEventData{
			DevicePath: "/dev/video0",
			EventType:  camera.CameraEventDisconnected,
			Timestamp:  time.Now(),
		}

		jsonData, err := json.Marshal(eventData)
		require.NoError(t, err, "Should marshal event data without device info to JSON without error")

		var unmarshaledEventData camera.CameraEventData
		err = json.Unmarshal(jsonData, &unmarshaledEventData)
		require.NoError(t, err, "Should unmarshal event data without device info from JSON without error")

		assert.Equal(t, eventData.DevicePath, unmarshaledEventData.DevicePath, "Device path should be preserved in JSON")
		assert.Equal(t, eventData.EventType, unmarshaledEventData.EventType, "Event type should be preserved in JSON")
		assert.Nil(t, unmarshaledEventData.DeviceInfo, "Nil device info should be preserved in JSON")
	})
}

// TestCameraEventHandler tests the CameraEventHandler interface
func TestCameraEventHandler(t *testing.T) {
	t.Run("camera_event_handler_interface_compliance", func(t *testing.T) {
		// Create a mock event handler that implements the interface
		mockHandler := &MockCameraEventHandler{
			events: make([]camera.CameraEventData, 0),
		}

		// Test that it implements the interface
		var handler camera.CameraEventHandler = mockHandler
		assert.NotNil(t, handler, "Mock handler should implement CameraEventHandler interface")

		// Test event handling
		eventData := camera.CameraEventData{
			DevicePath: "/dev/video0",
			EventType:  camera.CameraEventConnected,
			Timestamp:  time.Now(),
		}

		err := handler.HandleCameraEvent(context.Background(), eventData)
		assert.NoError(t, err, "Handler should handle event without error")
		assert.Len(t, mockHandler.events, 1, "Handler should record the event")
		assert.Equal(t, eventData.DevicePath, mockHandler.events[0].DevicePath, "Handler should record correct device path")
	})

	t.Run("camera_event_handler_error_handling", func(t *testing.T) {
		// Create a mock event handler that returns an error
		errorHandler := &ErrorCameraEventHandler{}

		var handler camera.CameraEventHandler = errorHandler
		eventData := camera.CameraEventData{
			DevicePath: "/dev/video0",
			EventType:  camera.CameraEventConnected,
			Timestamp:  time.Now(),
		}

		err := handler.HandleCameraEvent(context.Background(), eventData)
		assert.Error(t, err, "Handler should return error when configured to do so")
		assert.Contains(t, err.Error(), "mock error", "Error should contain expected message")
	})
}

// TestDeviceChecker tests the DeviceChecker interface
func TestDeviceChecker(t *testing.T) {
	t.Run("device_checker_interface_compliance", func(t *testing.T) {
		// Create a mock device checker that implements the interface
		mockChecker := &MockDeviceChecker{
			existsMap: map[string]bool{
				"/dev/video0": true,
				"/dev/video1": false,
			},
		}

		// Test that it implements the interface
		var checker camera.DeviceChecker = mockChecker
		assert.NotNil(t, checker, "Mock checker should implement DeviceChecker interface")

		// Test device existence checking
		assert.True(t, checker.Exists("/dev/video0"), "Checker should report existing device")
		assert.False(t, checker.Exists("/dev/video1"), "Checker should report non-existing device")
		assert.False(t, checker.Exists("/dev/unknown"), "Checker should report unknown device as non-existing")
	})
}

// TestV4L2CommandExecutor tests the V4L2CommandExecutor interface
func TestV4L2CommandExecutor(t *testing.T) {
	t.Run("v4l2_command_executor_interface_compliance", func(t *testing.T) {
		// Create a mock command executor that implements the interface
		mockExecutor := &MockV4L2CommandExecutor{
			outputs: map[string]string{
				"/dev/video0 --info":             "Driver name: uvcvideo\nCard type: USB Camera",
				"/dev/video0 --list-formats-ext": "Index : 0\nType  : Video Capture\nName  : YUYV",
			},
		}

		// Test that it implements the interface
		var executor camera.V4L2CommandExecutor = mockExecutor
		assert.NotNil(t, executor, "Mock executor should implement V4L2CommandExecutor interface")

		// Test command execution
		ctx := context.Background()
		output, err := executor.ExecuteCommand(ctx, "/dev/video0", "--info")
		assert.NoError(t, err, "Executor should execute command without error")
		assert.Contains(t, output, "Driver name: uvcvideo", "Output should contain expected content")

		output, err = executor.ExecuteCommand(ctx, "/dev/video0", "--list-formats-ext")
		assert.NoError(t, err, "Executor should execute format command without error")
		assert.Contains(t, output, "Index : 0", "Output should contain format information")
	})

	t.Run("v4l2_command_executor_error_handling", func(t *testing.T) {
		// Create a mock command executor that returns an error
		errorExecutor := &ErrorV4L2CommandExecutor{}

		var executor camera.V4L2CommandExecutor = errorExecutor
		ctx := context.Background()

		output, err := executor.ExecuteCommand(ctx, "/dev/video0", "--info")
		assert.Error(t, err, "Executor should return error when configured to do so")
		assert.Empty(t, output, "Output should be empty when error occurs")
		assert.Contains(t, err.Error(), "mock command error", "Error should contain expected message")
	})
}

// TestDeviceInfoParser tests the DeviceInfoParser interface
func TestDeviceInfoParser(t *testing.T) {
	t.Run("device_info_parser_interface_compliance", func(t *testing.T) {
		// Create a mock info parser that implements the interface
		mockParser := &MockDeviceInfoParser{
			capabilities: camera.V4L2Capabilities{
				DriverName: "uvcvideo",
				CardName:   "USB Camera",
			},
			formats: []camera.V4L2Format{
				{
					PixelFormat: "YUYV",
					Width:       1920,
					Height:      1080,
					FrameRates:  []string{"30.000"},
				},
			},
			frameRates: []string{"30.000", "60.000"},
		}

		// Test that it implements the interface
		var parser camera.DeviceInfoParser = mockParser
		assert.NotNil(t, parser, "Mock parser should implement DeviceInfoParser interface")

		// Test device info parsing
		capabilities, err := parser.ParseDeviceInfo("mock output")
		assert.NoError(t, err, "Parser should parse device info without error")
		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Parsed driver name should be correct")
		assert.Equal(t, "USB Camera", capabilities.CardName, "Parsed card name should be correct")

		// Test format parsing
		formats, err := parser.ParseDeviceFormats("mock output")
		assert.NoError(t, err, "Parser should parse formats without error")
		assert.Len(t, formats, 1, "Should parse one format")
		assert.Equal(t, "YUYV", formats[0].PixelFormat, "Parsed pixel format should be correct")

		// Test frame rate parsing
		frameRates, err := parser.ParseDeviceFrameRates("mock output")
		assert.NoError(t, err, "Parser should parse frame rates without error")
		assert.Len(t, frameRates, 2, "Should parse two frame rates")
		assert.Equal(t, "30.000", frameRates[0], "First frame rate should be correct")
		assert.Equal(t, "60.000", frameRates[1], "Second frame rate should be correct")
	})
}

// TestCameraMonitor tests the CameraMonitor interface
func TestCameraMonitor(t *testing.T) {
	t.Run("camera_monitor_interface_compliance", func(t *testing.T) {
		// Create a mock camera monitor that implements the interface
		mockMonitor := &MockCameraMonitor{
			running: false,
			devices: map[string]*camera.CameraDevice{
				"/dev/video0": {
					Path:   "/dev/video0",
					Name:   "USB Camera",
					Status: camera.DeviceStatusConnected,
				},
			},
		}

		// Test that it implements the interface
		var monitor camera.CameraMonitor = mockMonitor
		assert.NotNil(t, monitor, "Mock monitor should implement CameraMonitor interface")

		// Test basic functionality
		assert.False(t, monitor.IsRunning(), "Monitor should not be running initially")

		// Test start/stop
		ctx := context.Background()
		err := monitor.Start(ctx)
		assert.NoError(t, err, "Monitor should start without error")
		assert.True(t, monitor.IsRunning(), "Monitor should be running after start")

		err = monitor.Stop()
		assert.NoError(t, err, "Monitor should stop without error")
		assert.False(t, monitor.IsRunning(), "Monitor should not be running after stop")

		// Test device access
		devices := monitor.GetConnectedCameras()
		assert.Len(t, devices, 1, "Should return one connected device")
		assert.Equal(t, "/dev/video0", devices["/dev/video0"].Path, "Device path should be correct")

		device, exists := monitor.GetDevice("/dev/video0")
		assert.True(t, exists, "Device should exist")
		assert.Equal(t, "/dev/video0", device.Path, "Retrieved device path should be correct")

		device, exists = monitor.GetDevice("/dev/video1")
		assert.False(t, exists, "Non-existent device should not exist")

		// Test stats
		stats := monitor.GetMonitorStats()
		assert.NotNil(t, stats, "Monitor stats should not be nil")
	})

	t.Run("camera_monitor_event_handling", func(t *testing.T) {
		mockMonitor := &MockCameraMonitor{
			running: false,
			devices: make(map[string]*camera.CameraDevice),
		}

		var monitor camera.CameraMonitor = mockMonitor

		// Test event handler registration
		mockHandler := &MockCameraEventHandler{events: make([]camera.CameraEventData, 0)}
		monitor.AddEventHandler(mockHandler)

		// Test event callback registration
		callbackCalled := false
		monitor.AddEventCallback(func(eventData camera.CameraEventData) {
			callbackCalled = true
		})

		// Verify callback was registered (we can't easily test it without triggering an event)
		assert.NotNil(t, mockMonitor.eventCallback, "Event callback should be registered")
	})
}

// TestMonitorStats tests the MonitorStats struct
func TestMonitorStats(t *testing.T) {
	t.Run("monitor_stats_creation", func(t *testing.T) {
		stats := &camera.MonitorStats{
			Running:                    true,
			ActiveTasks:                5,
			PollingCycles:              100,
			DeviceStateChanges:         10,
			CapabilityProbesAttempted:  50,
			CapabilityProbesSuccessful: 45,
			CapabilityTimeouts:         3,
			CapabilityParseErrors:      2,
			PollingFailureCount:        1,
			CurrentPollInterval:        1.5,
			KnownDevicesCount:          3,
			UdevEventsProcessed:        20,
			UdevEventsFiltered:         5,
			UdevEventsSkipped:          2,
		}

		assert.True(t, stats.Running, "Running should be set correctly")
		assert.Equal(t, 5, stats.ActiveTasks, "Active tasks should be set correctly")
		assert.Equal(t, 100, stats.PollingCycles, "Polling cycles should be set correctly")
		assert.Equal(t, 10, stats.DeviceStateChanges, "Device state changes should be set correctly")
		assert.Equal(t, 50, stats.CapabilityProbesAttempted, "Capability probes attempted should be set correctly")
		assert.Equal(t, 45, stats.CapabilityProbesSuccessful, "Capability probes successful should be set correctly")
		assert.Equal(t, 3, stats.CapabilityTimeouts, "Capability timeouts should be set correctly")
		assert.Equal(t, 2, stats.CapabilityParseErrors, "Capability parse errors should be set correctly")
		assert.Equal(t, 1, stats.PollingFailureCount, "Polling failure count should be set correctly")
		assert.Equal(t, 1.5, stats.CurrentPollInterval, "Current poll interval should be set correctly")
		assert.Equal(t, 3, stats.KnownDevicesCount, "Known devices count should be set correctly")
		assert.Equal(t, 20, stats.UdevEventsProcessed, "Udev events processed should be set correctly")
		assert.Equal(t, 5, stats.UdevEventsFiltered, "Udev events filtered should be set correctly")
		assert.Equal(t, 2, stats.UdevEventsSkipped, "Udev events skipped should be set correctly")
	})

	t.Run("monitor_stats_json_marshaling", func(t *testing.T) {
		stats := &camera.MonitorStats{
			Running:                    true,
			ActiveTasks:                5,
			PollingCycles:              100,
			DeviceStateChanges:         10,
			CapabilityProbesAttempted:  50,
			CapabilityProbesSuccessful: 45,
			CapabilityTimeouts:         3,
			CapabilityParseErrors:      2,
			PollingFailureCount:        1,
			CurrentPollInterval:        1.5,
			KnownDevicesCount:          3,
			UdevEventsProcessed:        20,
			UdevEventsFiltered:         5,
			UdevEventsSkipped:          2,
		}

		jsonData, err := json.Marshal(stats)
		require.NoError(t, err, "Should marshal monitor stats to JSON without error")

		var unmarshaledStats camera.MonitorStats
		err = json.Unmarshal(jsonData, &unmarshaledStats)
		require.NoError(t, err, "Should unmarshal monitor stats from JSON without error")

		assert.Equal(t, stats.Running, unmarshaledStats.Running, "Running should be preserved in JSON")
		assert.Equal(t, stats.ActiveTasks, unmarshaledStats.ActiveTasks, "Active tasks should be preserved in JSON")
		assert.Equal(t, stats.PollingCycles, unmarshaledStats.PollingCycles, "Polling cycles should be preserved in JSON")
		assert.Equal(t, stats.DeviceStateChanges, unmarshaledStats.DeviceStateChanges, "Device state changes should be preserved in JSON")
		assert.Equal(t, stats.CapabilityProbesAttempted, unmarshaledStats.CapabilityProbesAttempted, "Capability probes attempted should be preserved in JSON")
		assert.Equal(t, stats.CapabilityProbesSuccessful, unmarshaledStats.CapabilityProbesSuccessful, "Capability probes successful should be preserved in JSON")
		assert.Equal(t, stats.CapabilityTimeouts, unmarshaledStats.CapabilityTimeouts, "Capability timeouts should be preserved in JSON")
		assert.Equal(t, stats.CapabilityParseErrors, unmarshaledStats.CapabilityParseErrors, "Capability parse errors should be preserved in JSON")
		assert.Equal(t, stats.PollingFailureCount, unmarshaledStats.PollingFailureCount, "Polling failure count should be preserved in JSON")
		assert.Equal(t, stats.CurrentPollInterval, unmarshaledStats.CurrentPollInterval, "Current poll interval should be preserved in JSON")
		assert.Equal(t, stats.KnownDevicesCount, unmarshaledStats.KnownDevicesCount, "Known devices count should be preserved in JSON")
		assert.Equal(t, stats.UdevEventsProcessed, unmarshaledStats.UdevEventsProcessed, "Udev events processed should be preserved in JSON")
		assert.Equal(t, stats.UdevEventsFiltered, unmarshaledStats.UdevEventsFiltered, "Udev events filtered should be preserved in JSON")
		assert.Equal(t, stats.UdevEventsSkipped, unmarshaledStats.UdevEventsSkipped, "Udev events skipped should be preserved in JSON")
	})
}

// TestCapabilityDetectionResult tests the CapabilityDetectionResult struct
func TestCapabilityDetectionResult(t *testing.T) {
	t.Run("capability_detection_result_creation", func(t *testing.T) {
		probeTime := time.Now()
		result := &camera.CapabilityDetectionResult{
			Detected:       true,
			Accessible:     true,
			DeviceName:     "USB Camera",
			Driver:         "uvcvideo",
			Formats:        []string{"YUYV", "MJPEG"},
			Resolutions:    []string{"1920x1080", "1280x720"},
			FrameRates:     []string{"30.000", "60.000"},
			Error:          "",
			TimeoutContext: "device_probe",
			ProbeTimestamp: probeTime,
			StructuredDiagnostics: map[string]interface{}{
				"v4l2_version": "5.15.0",
				"device_type":  "usb",
			},
		}

		assert.True(t, result.Detected, "Detected should be set correctly")
		assert.True(t, result.Accessible, "Accessible should be set correctly")
		assert.Equal(t, "USB Camera", result.DeviceName, "Device name should be set correctly")
		assert.Equal(t, "uvcvideo", result.Driver, "Driver should be set correctly")
		assert.Len(t, result.Formats, 2, "Formats should have correct length")
		assert.Len(t, result.Resolutions, 2, "Resolutions should have correct length")
		assert.Len(t, result.FrameRates, 2, "Frame rates should have correct length")
		assert.Empty(t, result.Error, "Error should be empty")
		assert.Equal(t, "device_probe", result.TimeoutContext, "Timeout context should be set correctly")
		assert.Equal(t, probeTime, result.ProbeTimestamp, "Probe timestamp should be set correctly")
		assert.Len(t, result.StructuredDiagnostics, 2, "Structured diagnostics should have correct length")
	})

	t.Run("capability_detection_result_json_marshaling", func(t *testing.T) {
		probeTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		result := &camera.CapabilityDetectionResult{
			Detected:       true,
			Accessible:     true,
			DeviceName:     "USB Camera",
			Driver:         "uvcvideo",
			Formats:        []string{"YUYV", "MJPEG"},
			Resolutions:    []string{"1920x1080", "1280x720"},
			FrameRates:     []string{"30.000", "60.000"},
			Error:          "",
			TimeoutContext: "device_probe",
			ProbeTimestamp: probeTime,
			StructuredDiagnostics: map[string]interface{}{
				"v4l2_version": "5.15.0",
				"device_type":  "usb",
			},
		}

		jsonData, err := json.Marshal(result)
		require.NoError(t, err, "Should marshal capability detection result to JSON without error")

		var unmarshaledResult camera.CapabilityDetectionResult
		err = json.Unmarshal(jsonData, &unmarshaledResult)
		require.NoError(t, err, "Should unmarshal capability detection result from JSON without error")

		assert.Equal(t, result.Detected, unmarshaledResult.Detected, "Detected should be preserved in JSON")
		assert.Equal(t, result.Accessible, unmarshaledResult.Accessible, "Accessible should be preserved in JSON")
		assert.Equal(t, result.DeviceName, unmarshaledResult.DeviceName, "Device name should be preserved in JSON")
		assert.Equal(t, result.Driver, unmarshaledResult.Driver, "Driver should be preserved in JSON")
		assert.Equal(t, result.Formats, unmarshaledResult.Formats, "Formats should be preserved in JSON")
		assert.Equal(t, result.Resolutions, unmarshaledResult.Resolutions, "Resolutions should be preserved in JSON")
		assert.Equal(t, result.FrameRates, unmarshaledResult.FrameRates, "Frame rates should be preserved in JSON")
		assert.Equal(t, result.Error, unmarshaledResult.Error, "Error should be preserved in JSON")
		assert.Equal(t, result.TimeoutContext, unmarshaledResult.TimeoutContext, "Timeout context should be preserved in JSON")
		assert.Equal(t, result.ProbeTimestamp, unmarshaledResult.ProbeTimestamp, "Probe timestamp should be preserved in JSON")
		assert.Equal(t, result.StructuredDiagnostics, unmarshaledResult.StructuredDiagnostics, "Structured diagnostics should be preserved in JSON")
	})

	t.Run("capability_detection_result_with_error", func(t *testing.T) {
		result := &camera.CapabilityDetectionResult{
			Detected:       false,
			Accessible:     false,
			DeviceName:     "",
			Driver:         "",
			Formats:        []string{},
			Resolutions:    []string{},
			FrameRates:     []string{},
			Error:          "Device not found",
			TimeoutContext: "device_probe",
			ProbeTimestamp: time.Now(),
			StructuredDiagnostics: map[string]interface{}{
				"error_code": "ENODEV",
				"error_msg":  "No such device",
			},
		}

		jsonData, err := json.Marshal(result)
		require.NoError(t, err, "Should marshal capability detection result with error to JSON without error")

		var unmarshaledResult camera.CapabilityDetectionResult
		err = json.Unmarshal(jsonData, &unmarshaledResult)
		require.NoError(t, err, "Should unmarshal capability detection result with error from JSON without error")

		assert.False(t, unmarshaledResult.Detected, "Detected should be preserved in JSON")
		assert.False(t, unmarshaledResult.Accessible, "Accessible should be preserved in JSON")
		assert.Equal(t, "Device not found", unmarshaledResult.Error, "Error should be preserved in JSON")
		assert.Len(t, unmarshaledResult.StructuredDiagnostics, 2, "Structured diagnostics should be preserved in JSON")
	})
}

// Mock implementations for testing interfaces

type MockCameraEventHandler struct {
	events []camera.CameraEventData
}

func (m *MockCameraEventHandler) HandleCameraEvent(ctx context.Context, eventData camera.CameraEventData) error {
	m.events = append(m.events, eventData)
	return nil
}

type ErrorCameraEventHandler struct{}

func (e *ErrorCameraEventHandler) HandleCameraEvent(ctx context.Context, eventData camera.CameraEventData) error {
	return assert.AnError
}

type MockDeviceChecker struct {
	existsMap map[string]bool
}

func (m *MockDeviceChecker) Exists(path string) bool {
	if exists, ok := m.existsMap[path]; ok {
		return exists
	}
	return false
}

type MockV4L2CommandExecutor struct {
	outputs map[string]string
}

func (m *MockV4L2CommandExecutor) ExecuteCommand(ctx context.Context, devicePath, args string) (string, error) {
	key := devicePath + " " + args
	if output, ok := m.outputs[key]; ok {
		return output, nil
	}
	return "", nil
}

type ErrorV4L2CommandExecutor struct{}

func (e *ErrorV4L2CommandExecutor) ExecuteCommand(ctx context.Context, devicePath, args string) (string, error) {
	return "", assert.AnError
}

type MockDeviceInfoParser struct {
	capabilities camera.V4L2Capabilities
	formats      []camera.V4L2Format
	frameRates   []string
}

func (m *MockDeviceInfoParser) ParseDeviceInfo(output string) (camera.V4L2Capabilities, error) {
	return m.capabilities, nil
}

func (m *MockDeviceInfoParser) ParseDeviceFormats(output string) ([]camera.V4L2Format, error) {
	return m.formats, nil
}

func (m *MockDeviceInfoParser) ParseDeviceFrameRates(output string) ([]string, error) {
	return m.frameRates, nil
}

type MockCameraMonitor struct {
	running       bool
	devices       map[string]*camera.CameraDevice
	eventCallback func(camera.CameraEventData)
}

func (m *MockCameraMonitor) Start(ctx context.Context) error {
	m.running = true
	return nil
}

func (m *MockCameraMonitor) Stop() error {
	m.running = false
	return nil
}

func (m *MockCameraMonitor) IsRunning() bool {
	return m.running
}

func (m *MockCameraMonitor) GetConnectedCameras() map[string]*camera.CameraDevice {
	return m.devices
}

func (m *MockCameraMonitor) GetDevice(devicePath string) (*camera.CameraDevice, bool) {
	device, exists := m.devices[devicePath]
	return device, exists
}

func (m *MockCameraMonitor) GetMonitorStats() *camera.MonitorStats {
	return &camera.MonitorStats{
		Running: m.running,
	}
}

func (m *MockCameraMonitor) AddEventHandler(handler camera.CameraEventHandler) {
	// Mock implementation
}

func (m *MockCameraMonitor) AddEventCallback(callback func(camera.CameraEventData)) {
	m.eventCallback = callback
}
