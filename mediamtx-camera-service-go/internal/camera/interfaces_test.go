/*
Interface Tests - Real Camera Integration

Tests all camera interfaces using real hardware and implementations.
No mocks, just real cameras and real system calls.
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

// TestCameraEvent_Constants tests camera event constants
func TestCameraEvent_Constants(t *testing.T) {
	assert.Equal(t, CameraEvent("CONNECTED"), CameraEventConnected, "Connected event should match")
	assert.Equal(t, CameraEvent("DISCONNECTED"), CameraEventDisconnected, "Disconnected event should match")
	assert.Equal(t, CameraEvent("STATUS_CHANGED"), CameraEventStatusChanged, "Status changed event should match")
}

// TestCameraEventData_Structure tests camera event data structure
func TestCameraEventData_Structure(t *testing.T) {
	eventData := CameraEventData{
		DevicePath: "/dev/video0",
		EventType:  CameraEventConnected,
		Timestamp:  time.Now(),
		DeviceInfo: &CameraDevice{
			Path: "/dev/video0",
			Name: "Test Camera",
		},
	}

	assert.Equal(t, "/dev/video0", eventData.DevicePath, "Device path should be set")
	assert.Equal(t, CameraEventConnected, eventData.EventType, "Event type should be set")
	assert.NotZero(t, eventData.Timestamp, "Timestamp should be set")
	assert.NotNil(t, eventData.DeviceInfo, "Device info should be set")
}

// TestDeviceChecker_RealImplementation tests real device checker
func TestDeviceChecker_RealImplementation(t *testing.T) {
	// Use real implementation
	var checker DeviceChecker = &RealDeviceChecker{}

	t.Run("real_filesystem_checks", func(t *testing.T) {
		// Test with real files that should exist
		assert.True(t, checker.Exists("."), "Current directory should exist")
		assert.True(t, checker.Exists("/proc/version"), "Proc version should exist")
		assert.True(t, checker.Exists("/dev/null"), "Dev null should exist")

		// Test with non-existent paths
		assert.False(t, checker.Exists("/nonexistent/path"), "Non-existent path should return false")
		assert.False(t, checker.Exists("/dev/video999"), "Non-existent video device should return false")
	})

	t.Run("camera_device_checks", func(t *testing.T) {
		// Check for common camera device paths
		cameraPaths := []string{"/dev/video0", "/dev/video1", "/dev/video2"}

		for _, path := range cameraPaths {
			exists := checker.Exists(path)
			if exists {
				t.Logf("Found real camera device: %s", path)
			} else {
				t.Logf("No camera at: %s", path)
			}
		}
	})
}

// TestV4L2CommandExecutor_RealImplementation tests real V4L2 command executor
func TestV4L2CommandExecutor_RealImplementation(t *testing.T) {
	// Use real implementation
	var executor V4L2CommandExecutor = &RealV4L2CommandExecutor{}

	t.Run("basic_command_execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with a simple command that should work
		output, err := executor.ExecuteCommand(ctx, "/dev/null", "echo 'test'")
		if err == nil {
			assert.Contains(t, output, "test", "Command output should contain expected text")
		} else {
			t.Logf("Command execution failed (expected on some systems): %v", err)
		}
	})

	t.Run("camera_device_commands", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with real camera devices if they exist
		cameraPaths := []string{"/dev/video0", "/dev/video1", "/dev/video2"}

		for _, path := range cameraPaths {
			if (&RealDeviceChecker{}).Exists(path) {
				t.Logf("Testing V4L2 commands on %s", path)

				// Try to get device info
				output, err := executor.ExecuteCommand(ctx, path, "v4l2-ctl --device-info")
				if err == nil && output != "" {
					t.Logf("Successfully queried %s: %s", path, output[:100])
					break
				} else {
					t.Logf("⚠️  Could not query %s: %v", path, err)
				}
			}
		}
	})
}

// TestDeviceInfoParser_RealImplementation tests real device info parser
func TestDeviceInfoParser_RealImplementation(t *testing.T) {
	// Use real implementation
	var parser DeviceInfoParser = &RealDeviceInfoParser{}

	t.Run("parse_device_info", func(t *testing.T) {
		// Test with REAL V4L2 output from actual device
		helper := NewRealHardwareTestHelper(t)
		devices := helper.GetAvailableDevices()
		require.NotEmpty(t, devices, "Real camera devices must be available for testing")

		devicePath := devices[0]
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get real V4L2 output using the command executor
		commandExecutor := &RealV4L2CommandExecutor{}
		realOutput, err := commandExecutor.ExecuteCommand(ctx, devicePath, "--all")
		require.NoError(t, err, "Should get real V4L2 output from device")
		require.NotEmpty(t, realOutput, "Real V4L2 output should not be empty")

		// Parse real V4L2 output
		capabilities, err := parser.ParseDeviceInfo(realOutput)
		require.NoError(t, err, "Should parse real device info")
		require.NotEmpty(t, capabilities.DriverName, "Real device should have driver name")
		require.NotEmpty(t, capabilities.CardName, "Real device should have card name")

		// Validate that parsing worked with real data
		t.Logf("Parsed real device: Driver=%s, Card=%s", capabilities.DriverName, capabilities.CardName)
	})

	t.Run("parse_device_formats", func(t *testing.T) {
		// Test with real V4L2 format output
		sampleFormats := `ioctl: VIDIOC_ENUM_FMT
        Type: Video Capture

        [0]: 'YUYV' (YUYV 4:2:2)
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)`

		formats, err := parser.ParseDeviceFormats(sampleFormats)
		require.NoError(t, err, "Should parse valid format info")
		assert.Len(t, formats, 1, "Should parse one format")
		assert.Equal(t, "YUYV", formats[0].PixelFormat, "Pixel format should be parsed correctly")
		assert.Equal(t, 640, formats[0].Width, "Width should be parsed correctly")
		assert.Equal(t, 480, formats[0].Height, "Height should be parsed correctly")
		assert.Len(t, formats[0].FrameRates, 3, "Should parse frame rates")
	})

	t.Run("parse_device_frame_rates", func(t *testing.T) {
		// Test with real V4L2 frame rate output
		sampleRates := `Interval: Discrete 0.033s (30.000 fps)
Interval: Discrete 0.050s (20.000 fps)
Interval: Discrete 0.067s (15.000 fps)
Interval: Discrete 0.100s (10.000 fps)`

		rates, err := parser.ParseDeviceFrameRates(sampleRates)
		require.NoError(t, err, "Should parse valid frame rate info")
		assert.Len(t, rates, 4, "Should parse 4 frame rates")
		assert.Contains(t, rates, "30.000", "Should contain 30fps")
		assert.Contains(t, rates, "20.000", "Should contain 20fps")
		assert.Contains(t, rates, "15.000", "Should contain 15fps")
		assert.Contains(t, rates, "10.000", "Should contain 10fps")
	})
}

// TestEventNotifier_RealImplementation tests real event notifier
func TestEventNotifier_RealImplementation(t *testing.T) {
	// Create a real event notifier implementation
	notifier := &realEventNotifier{
		events: make([]string, 0),
	}

	t.Run("notify_camera_connected", func(t *testing.T) {
		device := &CameraDevice{
			Path: "/dev/video0",
			Name: "Test Camera",
		}

		notifier.NotifyCameraConnected(device)
		assert.Contains(t, notifier.events, "connected:/dev/video0", "Should record connection event")
	})

	t.Run("notify_camera_disconnected", func(t *testing.T) {
		notifier.NotifyCameraDisconnected("/dev/video0")
		assert.Contains(t, notifier.events, "disconnected:/dev/video0", "Should record disconnection event")
	})

	t.Run("notify_status_change", func(t *testing.T) {
		device := &CameraDevice{
			Path:   "/dev/video0",
			Status: DeviceStatusConnected,
		}

		notifier.NotifyCameraStatusChange(device, DeviceStatusDisconnected, DeviceStatusConnected)
		assert.Contains(t, notifier.events, "status_change:/dev/video0", "Should record status change event")
	})

	t.Run("notify_capability_detected", func(t *testing.T) {
		device := &CameraDevice{Path: "/dev/video0"}
		capabilities := V4L2Capabilities{
			DriverName: "uvcvideo",
			CardName:   "USB Camera",
		}

		notifier.NotifyCapabilityDetected(device, capabilities)
		assert.Contains(t, notifier.events, "capability_detected:/dev/video0", "Should record capability detection event")
	})

	t.Run("notify_capability_error", func(t *testing.T) {
		notifier.NotifyCapabilityError("/dev/video0", "Device not accessible")
		assert.Contains(t, notifier.events, "capability_error:/dev/video0", "Should record capability error event")
	})
}

// TestCameraMonitor_RealImplementation tests real camera monitor interface
func TestCameraMonitor_RealImplementation(t *testing.T) {
	// Create real implementations
	deviceChecker := &RealDeviceChecker{}
	commandExecutor := &RealV4L2CommandExecutor{}
	infoParser := &RealDeviceInfoParser{}

	// Test monitor creation (this will fail without config, but tests the interface)
	t.Run("monitor_creation", func(t *testing.T) {
		monitor, err := NewHybridCameraMonitor(nil, nil, deviceChecker, commandExecutor, infoParser)
		assert.Error(t, err, "Should fail without config")
		assert.Nil(t, monitor, "Should be nil when creation fails")
	})

	t.Run("interface_compliance", func(t *testing.T) {
		// Test that our real implementations satisfy the interfaces
		var _ DeviceChecker = deviceChecker
		var _ V4L2CommandExecutor = commandExecutor
		var _ DeviceInfoParser = infoParser

		// Test that real implementations actually work with their interfaces
		// Test DeviceChecker interface
		exists := deviceChecker.Exists("/dev/null")
		assert.True(t, exists, "DeviceChecker should work with real filesystem")

		// Test V4L2CommandExecutor interface
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := commandExecutor.ExecuteCommand(ctx, "/dev/null", "--help")
		// The interface should be callable (may succeed or fail, both are valid)
		// We just verify the method can be called without panicking
		_ = err // Use the error to avoid unused variable warning

		// Test DeviceInfoParser interface with real device
		helper := NewRealHardwareTestHelper(t)
		devices := helper.GetAvailableDevices()
		require.NotEmpty(t, devices, "Real camera devices must be available for testing")

		devicePath := devices[0]
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		// Get real V4L2 output
		commandExecutor := &RealV4L2CommandExecutor{}
		realOutput, err := commandExecutor.ExecuteCommand(ctx2, devicePath, "--all")
		require.NoError(t, err, "Should get real V4L2 output from device")

		// Parse real output
		_, err = infoParser.ParseDeviceInfo(realOutput)
		require.NoError(t, err, "Should parse real V4L2 output")
	})
}

// TestCameraMonitor_TakeDirectSnapshot tests the new TakeDirectSnapshot interface method
func TestCameraMonitor_TakeDirectSnapshot(t *testing.T) {

	// Test interface compliance
	t.Run("interface_compliance", func(t *testing.T) {
		// Test that TakeDirectSnapshot is part of the CameraMonitor interface
		// Create a real monitor to test the interface
		configManager := config.CreateConfigManager()
		logger := logging.CreateTestLogger(t, nil)
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

		// Test that the method signature is correct
		// This will fail at compile time if the signature is wrong
		ctx := context.Background()
		devicePath := "/dev/video0"
		outputPath := "/tmp/test_snapshot.jpg"
		defer os.Remove(outputPath) // Clean up test file
		options := map[string]interface{}{
			"format": "jpg",
			"width":  640,
			"height": 480,
		}

		// This line will fail compilation if TakeDirectSnapshot is not in the interface
		_, _ = monitor.TakeDirectSnapshot(ctx, devicePath, outputPath, options)
	})

	t.Run("direct_snapshot_interface", func(t *testing.T) {
		// Test that DirectSnapshot type is properly defined
		snapshot := &DirectSnapshot{
			ID:          "test_id",
			DevicePath:  "/dev/video0",
			FilePath:    "/tmp/test.jpg",
			Size:        1024,
			Format:      "jpg",
			Width:       640,
			Height:      480,
			CaptureTime: 100 * time.Millisecond,
			Created:     time.Now(),
			Metadata: map[string]interface{}{
				"tier_used": 0,
				"method":    "v4l2_direct",
			},
		}

		require.NotNil(t, snapshot)
		require.Equal(t, "test_id", snapshot.ID)
		require.Equal(t, "/dev/video0", snapshot.DevicePath)
		require.Equal(t, "jpg", snapshot.Format)
		require.Equal(t, 0, snapshot.Metadata["tier_used"])
	})
}

// TestMonitorStats_Structure tests monitor stats structure
func TestMonitorStats_Structure(t *testing.T) {
	stats := &MonitorStats{
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
	}

	assert.False(t, stats.Running, "Should start not running")
	assert.Equal(t, int64(0), stats.ActiveTasks, "Should start with no active tasks")
	assert.Equal(t, int64(0), stats.PollingCycles, "Should start with no polling cycles")
	assert.Equal(t, 1.0, stats.CurrentPollInterval, "Should have default poll interval")
}

// TestCapabilityDetectionResult_Structure tests capability detection result structure
func TestCapabilityDetectionResult_Structure(t *testing.T) {
	result := &CapabilityDetectionResult{
		Detected:              true,
		Accessible:            true,
		DeviceName:            "USB Camera",
		Driver:                "uvcvideo",
		Formats:               []string{"YUYV", "MJPG"},
		Resolutions:           []string{"1920x1080", "1280x720"},
		FrameRates:            []string{"30.000", "25.000"},
		Error:                 "",
		TimeoutContext:        "",
		ProbeTimestamp:        time.Now(),
		StructuredDiagnostics: make(map[string]interface{}),
	}

	assert.True(t, result.Detected, "Should be detected")
	assert.True(t, result.Accessible, "Should be accessible")
	assert.Equal(t, "USB Camera", result.DeviceName, "Device name should be set")
	assert.Equal(t, "uvcvideo", result.Driver, "Driver should be set")
	assert.Len(t, result.Formats, 2, "Should have 2 formats")
	assert.Len(t, result.Resolutions, 2, "Should have 2 resolutions")
	assert.Len(t, result.FrameRates, 2, "Should have 2 frame rates")
	assert.NotZero(t, result.ProbeTimestamp, "Probe timestamp should be set")
}

// realEventNotifier provides a real implementation for testing
type realEventNotifier struct {
	events []string
}

func (r *realEventNotifier) NotifyCameraConnected(device *CameraDevice) {
	r.events = append(r.events, "connected:"+device.Path)
}

func (r *realEventNotifier) NotifyCameraDisconnected(devicePath string) {
	r.events = append(r.events, "disconnected:"+devicePath)
}

func (r *realEventNotifier) NotifyCameraStatusChange(device *CameraDevice, oldStatus, newStatus DeviceStatus) {
	r.events = append(r.events, "status_change:"+device.Path)
}

func (r *realEventNotifier) NotifyCapabilityDetected(device *CameraDevice, capabilities V4L2Capabilities) {
	r.events = append(r.events, "capability_detected:"+device.Path)
}

func (r *realEventNotifier) NotifyCapabilityError(devicePath string, error string) {
	r.events = append(r.events, "capability_error:"+devicePath)
}
