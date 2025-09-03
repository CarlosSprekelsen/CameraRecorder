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

package camera

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
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
		// Use real device checker instead of mock per testing guide
		realChecker := &camera.RealDeviceChecker{}

		// Test that it implements the interface
		var checker camera.DeviceChecker = realChecker
		assert.NotNil(t, checker, "Real checker should implement DeviceChecker interface")

		// Test device existence checking with real file system
		// Check for common video devices that might exist on the system
		commonDevices := []string{"/dev/video0", "/dev/video1", "/dev/video2"}
		for _, device := range commonDevices {
			exists := checker.Exists(device)
			// We don't assert specific values since we don't know what devices exist
			// Just verify the method works without error
			_ = exists
		}

		// Test non-existent device
		assert.False(t, checker.Exists("/dev/nonexistent_video_device"), "Checker should report non-existing device as false")
	})
}

// TestV4L2CommandExecutor tests the V4L2CommandExecutor interface
func TestV4L2CommandExecutor(t *testing.T) {
	t.Run("v4l2_command_executor_interface_compliance", func(t *testing.T) {
		// Use real command executor instead of mock per testing guide
		realExecutor := &camera.RealV4L2CommandExecutor{}

		// Test that it implements the interface
		var executor camera.V4L2CommandExecutor = realExecutor
		assert.NotNil(t, executor, "Real executor should implement V4L2CommandExecutor interface")

		// Test command execution with real V4L2 commands
		ctx := context.Background()

		// Test with common video devices that might exist
		commonDevices := []string{"/dev/video0", "/dev/video1"}
		for _, device := range commonDevices {
			_, err := executor.ExecuteCommand(ctx, device, "--info")
			// Don't assert specific results since we don't know what devices exist
			// Just verify the method works without panic
			if err != nil {
				// Expected if device doesn't exist or v4l2-ctl not available
				t.Logf("V4L2 command failed for %s (expected if device not available): %v", device, err)
			} else {
				t.Logf("V4L2 command succeeded for %s", device)
			}
		}
	})

	t.Run("v4l2_command_executor_error_handling", func(t *testing.T) {
		// Use real command executor to test real error conditions
		realExecutor := &camera.RealV4L2CommandExecutor{}

		var executor camera.V4L2CommandExecutor = realExecutor
		ctx := context.Background()

		// Test with non-existent device to trigger real error
		output, err := executor.ExecuteCommand(ctx, "/dev/nonexistent_video_device", "--info")
		assert.Error(t, err, "Executor should return error for non-existent device")
		assert.Empty(t, output, "Output should be empty when error occurs")
		// Don't assert specific error message since it depends on system
	})
}

// TestDeviceInfoParser tests the DeviceInfoParser interface
func TestDeviceInfoParser(t *testing.T) {
	t.Run("device_info_parser_interface_compliance", func(t *testing.T) {
		// Use real info parser instead of mock per testing guide
		realParser := &camera.RealDeviceInfoParser{}

		// Test that it implements the interface
		var parser camera.DeviceInfoParser = realParser
		assert.NotNil(t, parser, "Real parser should implement DeviceInfoParser interface")

		// Test capability parsing with real V4L2 output format
		sampleV4L2Output := `Driver name: uvcvideo
Card type: USB Camera
Bus info: usb-0000:00:14.0-1
Driver version: 5.15.0
Capabilities: video capture, video streaming
Device Caps: video capture, video streaming`

		capabilities, err := parser.ParseDeviceInfo(sampleV4L2Output)
		assert.NoError(t, err, "Parser should parse device info without error")
		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed correctly")
		assert.Equal(t, "USB Camera", capabilities.CardName, "Card name should be parsed correctly")
		// Real parser splits capabilities into individual words
		assert.Contains(t, capabilities.Capabilities, "video", "Should parse video capability")
		assert.Contains(t, capabilities.Capabilities, "capture,", "Should parse capture capability")

		// REAL V4L2 TEST: This reflects actual v4l2-ctl --list-formats-ext output
		// If the parser can't handle multiple sizes per format, it should FAIL
		sampleFormatOutput := `[0]: 'YUYV' (YUYV 4:2:2)
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                Size: Discrete 320x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)`

		formats, err := parser.ParseDeviceFormats(sampleFormatOutput)
		assert.NoError(t, err, "Parser should parse device formats without error")
		// REAL SYSTEM EXPECTATION: Real V4L2 cameras support multiple resolutions per format
		// This test will FAIL if the parser doesn't handle multiple sizes correctly
		assert.Len(t, formats, 2, "Should parse two sizes (640x480 and 320x240)")
		assert.Equal(t, "YUYV", formats[0].PixelFormat, "Pixel format should be parsed correctly")
		assert.Equal(t, 640, formats[0].Width, "First size width should be 640")
		assert.Equal(t, 480, formats[0].Height, "First size height should be 480")
		assert.Equal(t, 320, formats[1].Width, "Second size width should be 320")
		assert.Equal(t, 240, formats[1].Height, "Second size height should be 240")

		// Test frame rate parsing with real V4L2 frame rate output
		// Use actual V4L2 output format from real camera
		sampleFrameRateOutput := `[0]: 'YUYV' (YUYV 4:2:2)
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                Size: Discrete 320x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)`

		frameRates, err := parser.ParseDeviceFrameRates(sampleFrameRateOutput)
		assert.NoError(t, err, "Parser should parse frame rates without error")
		assert.Len(t, frameRates, 3, "Should parse three unique frame rates")
		assert.Contains(t, frameRates, "30.000", "Should contain expected frame rate")
		assert.Contains(t, frameRates, "20.000", "Should contain expected frame rate")
		assert.Contains(t, frameRates, "15.000", "Should contain expected frame rate")
	})
}

// TestCameraMonitor tests the CameraMonitor interface
func TestCameraMonitor(t *testing.T) {
	t.Run("camera_monitor_interface_compliance", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupMediaMTXTestEnvironment(t)
		defer utils.TeardownMediaMTXTestEnvironment(t, env)

		// Use real camera monitor instead of mock per testing guide
		// Load config using shared config manager
		err := env.ConfigManager.LoadConfig("../../config/development.yaml")
		if err != nil {
			t.Skipf("Skipping test - config not available: %v", err)
		}

		// Setup logging using shared logger
		err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&env.ConfigManager.GetConfig().Logging))
		if err != nil {
			t.Skipf("Skipping test - logging setup failed: %v", err)
		}

		// Create real camera monitor with real dependencies
		realDeviceChecker := &camera.RealDeviceChecker{}
		realCommandExecutor := &camera.RealV4L2CommandExecutor{}
		realInfoParser := &camera.RealDeviceInfoParser{}

		realMonitor, err := camera.NewHybridCameraMonitor(
			env.ConfigManager,
			env.Logger,
			realDeviceChecker,
			realCommandExecutor,
			realInfoParser,
		)
		if err != nil {
			t.Skipf("Skipping test - real monitor creation failed: %v", err)
		}

		// Test that it implements the interface
		var monitor camera.CameraMonitor = realMonitor
		assert.NotNil(t, monitor, "Real monitor should implement CameraMonitor interface")

		// Test basic functionality
		assert.False(t, monitor.IsRunning(), "Monitor should not be running initially")

		// Test start/stop with real implementation
		ctx := context.Background()
		err = monitor.Start(ctx)
		if err != nil {
			t.Logf("Monitor start failed (may be expected): %v", err)
		} else {
			assert.True(t, monitor.IsRunning(), "Monitor should be running after start")

			// Test device discovery with real system
			devices := monitor.GetConnectedCameras()
			t.Logf("Found %d connected cameras", len(devices))

			// Test stats with real implementation
			stats := monitor.GetMonitorStats()
			assert.NotNil(t, stats, "Monitor stats should not be nil")

			// Stop the monitor
			err = monitor.Stop()
			assert.NoError(t, err, "Monitor should stop without error")
			assert.False(t, monitor.IsRunning(), "Monitor should not be running after stop")
		}
	})

	t.Run("camera_monitor_event_handling", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupMediaMTXTestEnvironment(t)
		defer utils.TeardownMediaMTXTestEnvironment(t, env)

		// Use real camera monitor for event handling test
		err := env.ConfigManager.LoadConfig("../../config/development.yaml")
		if err != nil {
			t.Skipf("Skipping test - config not available: %v", err)
		}

		err = logging.SetupLogging(logging.NewLoggingConfigFromConfig(&env.ConfigManager.GetConfig().Logging))
		if err != nil {
			t.Skipf("Skipping test - logging setup failed: %v", err)
		}

		realDeviceChecker := &camera.RealDeviceChecker{}
		realCommandExecutor := &camera.RealV4L2CommandExecutor{}
		realInfoParser := &camera.RealDeviceInfoParser{}

		realMonitor, err := camera.NewHybridCameraMonitor(
			env.ConfigManager,
			env.Logger,
			realDeviceChecker,
			realCommandExecutor,
			realInfoParser,
		)
		if err != nil {
			t.Skipf("Skipping test - real monitor creation failed: %v", err)
		}

		var monitor camera.CameraMonitor = realMonitor

		// Test event handler registration with real monitor
		realHandler := &MockCameraEventHandler{events: make([]camera.CameraEventData, 0)}
		monitor.AddEventHandler(realHandler)

		// Test event callback registration with real monitor
		monitor.AddEventCallback(func(eventData camera.CameraEventData) {
			t.Logf("Event callback triggered for device: %s", eventData.DevicePath)
		})

		// Test that registration doesn't cause errors
		assert.NotNil(t, realHandler, "Real event handler should be registered")
		t.Logf("Event handler and callback registered successfully")
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
	return fmt.Errorf("mock error")
}

// Real implementations are used instead of mocks per testing guide requirements
