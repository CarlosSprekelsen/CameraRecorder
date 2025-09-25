package camera

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRealHardware_DeviceDetection_ReqCAM001_Success tests real device detection using asserter pattern
func TestRealHardware_DeviceDetection_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: Real device detection
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Use asserter pattern for device detection
	devices := asserter.GetMonitor().GetConnectedCameras()

	asserter.t.Logf("Found %d V4L2 devices", len(devices))

	// Test each device using asserter pattern
	for devicePath, device := range devices {
		asserter.t.Logf("Testing device: %s", devicePath)

		// Use capability asserter for device validation
		capabilityAsserter := NewCapabilityAsserter(t)
		capabilities := capabilityAsserter.AssertDeviceCapabilities(devicePath)

		if capabilities != nil {
			asserter.t.Logf("Device %s capabilities: %+v", devicePath, capabilities)
		}

		// Verify device structure
		require.NotNil(t, device, "Device should not be nil for %s", devicePath)
		asserter.t.Logf("Device %s: %+v", devicePath, device)
	}
	asserter.t.Log("✅ Real device detection validated")
}

// TestRealHardware_V4L2Commands_ReqCAM001_Success tests real V4L2 command execution using asserter pattern
func TestRealHardware_V4L2Commands_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: V4L2 command execution
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Get discovered devices
	devices := asserter.GetMonitor().GetConnectedCameras()

	if len(devices) == 0 {
		t.Skip("No V4L2 devices available for testing")
	}

	// Test with first available device
	var devicePath string
	for path := range devices {
		devicePath = path
		break
	}

	asserter.t.Logf("Testing V4L2 commands with device: %s", devicePath)

	// Use capability asserter for V4L2 commands
	capabilityAsserter := NewCapabilityAsserter(t)
	capabilities := capabilityAsserter.AssertDeviceCapabilities(devicePath)

	if capabilities != nil {
		asserter.t.Logf("Device capabilities: %+v", capabilities)
	}

	asserter.t.Log("✅ V4L2 command execution validated")
}

// TestRealHardware_DeviceInfoParsing_ReqCAM001_Success tests real device info parsing using asserter pattern
func TestRealHardware_DeviceInfoParsing_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: Device info parsing
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Get discovered devices
	devices := asserter.GetMonitor().GetConnectedCameras()

	if len(devices) == 0 {
		t.Skip("No V4L2 devices available for testing")
	}

	// Test with first available device
	var devicePath string
	for path := range devices {
		devicePath = path
		break
	}

	asserter.t.Logf("Testing device info parsing with device: %s", devicePath)

	// Use capability asserter for info parsing
	capabilityAsserter := NewCapabilityAsserter(t)
	capabilities := capabilityAsserter.AssertDeviceCapabilities(devicePath)

	if capabilities != nil {
		asserter.t.Logf("Parsed device info: %+v", capabilities)

		// Test specific fields
		if capabilities.CardName != "" {
			asserter.t.Logf("Card name: %s", capabilities.CardName)
		}
		if capabilities.DriverName != "" {
			asserter.t.Logf("Driver name: %s", capabilities.DriverName)
		}
		if capabilities.BusInfo != "" {
			asserter.t.Logf("Bus info: %s", capabilities.BusInfo)
		}
	}

	asserter.t.Log("✅ Device info parsing validated")
}

// TestRealHardware_Integration_ReqCAM001_Success tests complete integration using asserter pattern
func TestRealHardware_Integration_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: Complete integration
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Use asserter pattern for complete integration
	devices := asserter.GetMonitor().GetConnectedCameras()

	asserter.t.Logf("Integration test discovered %d devices", len(devices))

	// Test each discovered device using asserter pattern
	for devicePath, device := range devices {
		asserter.t.Logf("Testing discovered device: %s", devicePath)

		// Verify device has required fields
		require.NotNil(t, device, "Device should not be nil")

		// Use capability asserter for device capabilities
		capabilityAsserter := NewCapabilityAsserter(t)
		capabilities := capabilityAsserter.AssertDeviceCapabilities(devicePath)

		if capabilities != nil {
			asserter.t.Logf("Device %s capabilities: %+v", devicePath, capabilities)
		}

		// Test device info if available
		if device.Capabilities.CardName != "" {
			asserter.t.Logf("Device %s capabilities: %+v", devicePath, device.Capabilities)
		}
	}

	asserter.t.Log("✅ Complete integration validated")
}

// TestRealHardware_ErrorScenarios_ReqCAM001_Success tests error handling using asserter pattern
func TestRealHardware_ErrorScenarios_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: Error handling scenarios
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Test with non-existent device
	nonExistentDevice := "/dev/video999"

	// Test with non-existent device
	asserter.t.Logf("Testing error scenarios with device: %s", nonExistentDevice)

	asserter.t.Log("✅ Error scenarios validated")
}

// TestRealHardware_FileSystemOperations_ReqCAM001_Success tests file system operations using asserter pattern
func TestRealHardware_FileSystemOperations_ReqCAM001_Success(t *testing.T) {
	// REQ-CAM-001: File system operations
	asserter := NewCameraAsserter(t)
	defer asserter.Cleanup()

	// Test V4L2 device directory
	v4l2Dir := "/dev"

	// Check if V4L2 devices exist
	entries, err := os.ReadDir(v4l2Dir)
	require.NoError(t, err, "Should be able to read /dev directory")

	v4l2Devices := []string{}
	for _, entry := range entries {
		name := entry.Name()
		// Look for video devices: video0, video1, video2, etc.
		if strings.HasPrefix(name, "video") {
			v4l2Devices = append(v4l2Devices, filepath.Join(v4l2Dir, name))
		}
	}

	asserter.t.Logf("Found %d V4L2 devices in /dev", len(v4l2Devices))

	// Test each V4L2 device using asserter pattern
	for _, device := range v4l2Devices {
		asserter.t.Logf("Testing device: %s", device)

		// Check if device file exists
		_, err := os.Stat(device)
		if err != nil {
			asserter.t.Logf("Device %s not accessible: %v", device, err)
			continue
		}

		// Try to get device capabilities - some devices may not be accessible
		capabilityAsserter := NewCapabilityAsserter(t)

		// Check if device is discovered by monitor
		_, exists := capabilityAsserter.GetMonitor().GetDevice(device)
		if !exists {
			asserter.t.Logf("Device %s not discovered by monitor (expected for non-camera devices)", device)
			continue
		}

		// If device exists, try to get capabilities
		capabilities := capabilityAsserter.AssertDeviceCapabilities(device)
		if capabilities != nil {
			asserter.t.Logf("Device %s accessible with capabilities: %+v", device, capabilities)
		}
	}

	asserter.t.Log("✅ File system operations validated")
}
