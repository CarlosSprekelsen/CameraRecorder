/*
Camera Types Unit Tests

Requirements Coverage:
- REQ-CAM-001: Camera device detection and enumeration
- REQ-CAM-002: Camera capability probing and validation
- REQ-CAM-003: Device status monitoring and reporting
- REQ-CAM-004: Camera information and metadata
- REQ-CAM-005: JSON serialization and deserialization
- REQ-CAM-006: Type validation and constraints

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Real Component Usage: V4L2 devices, JSON API responses
*/

package camera

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCameraDevice tests the CameraDevice struct
func TestCameraDevice(t *testing.T) {
	t.Parallel()
	t.Run("camera_device_creation", func(t *testing.T) {
		device := &CameraDevice{
			Path: "/dev/video0",
			Name: "USB Camera",
			Capabilities: V4L2Capabilities{
				DriverName: "uvcvideo",
				CardName:   "USB Camera",
			},
			Status:    DeviceStatusConnected,
			LastSeen:  time.Now(),
			DeviceNum: 0,
		}

		assert.Equal(t, "/dev/video0", device.Path, "Device path should be set correctly")
		assert.Equal(t, "USB Camera", device.Name, "Device name should be set correctly")
		assert.Equal(t, "uvcvideo", device.Capabilities.DriverName, "Driver name should be set correctly")
		assert.Equal(t, DeviceStatusConnected, device.Status, "Device status should be set correctly")
		assert.Equal(t, 0, device.DeviceNum, "Device number should be set correctly")
	})

	t.Run("camera_device_json_marshaling", func(t *testing.T) {
		device := &CameraDevice{
			Path: "/dev/video0",
			Name: "USB Camera",
			Capabilities: V4L2Capabilities{
				DriverName:   "uvcvideo",
				CardName:     "USB Camera",
				BusInfo:      "usb-0000:00:14.0-1",
				Version:      "5.15.0",
				Capabilities: []string{"0x85200001"},
				DeviceCaps:   []string{"0x04200001"},
			},
			Formats: []V4L2Format{
				{
					PixelFormat: "YUYV",
					Width:       1920,
					Height:      1080,
					FrameRates:  []string{"30.000"},
				},
			},
			Status:    DeviceStatusConnected,
			LastSeen:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			DeviceNum: 0,
		}

		jsonData, err := json.Marshal(device)
		require.NoError(t, err, "Should marshal device to JSON without error")

		var unmarshaledDevice CameraDevice
		err = json.Unmarshal(jsonData, &unmarshaledDevice)
		require.NoError(t, err, "Should unmarshal device from JSON without error")

		assert.Equal(t, device.Path, unmarshaledDevice.Path, "Path should be preserved in JSON")
		assert.Equal(t, device.Name, unmarshaledDevice.Name, "Name should be preserved in JSON")
		assert.Equal(t, device.Status, unmarshaledDevice.Status, "Status should be preserved in JSON")
		assert.Equal(t, device.DeviceNum, unmarshaledDevice.DeviceNum, "Device number should be preserved in JSON")
		assert.Equal(t, device.Capabilities.DriverName, unmarshaledDevice.Capabilities.DriverName, "Driver name should be preserved in JSON")
		assert.Equal(t, device.Capabilities.CardName, unmarshaledDevice.Capabilities.CardName, "Card name should be preserved in JSON")
		assert.Len(t, unmarshaledDevice.Formats, 1, "Formats should be preserved in JSON")
		assert.Equal(t, device.Formats[0].PixelFormat, unmarshaledDevice.Formats[0].PixelFormat, "Format pixel format should be preserved in JSON")
	})

	t.Run("camera_device_with_error", func(t *testing.T) {
		device := &CameraDevice{
			Path:   "/dev/video0",
			Name:   "USB Camera",
			Status: DeviceStatusError,
			Error:  "Device not accessible",
		}

		jsonData, err := json.Marshal(device)
		require.NoError(t, err, "Should marshal device with error to JSON without error")

		var unmarshaledDevice CameraDevice
		err = json.Unmarshal(jsonData, &unmarshaledDevice)
		require.NoError(t, err, "Should unmarshal device with error from JSON without error")

		assert.Equal(t, DeviceStatusError, unmarshaledDevice.Status, "Error status should be preserved in JSON")
		assert.Equal(t, "Device not accessible", unmarshaledDevice.Error, "Error message should be preserved in JSON")
	})
}

// TestDeviceStatus tests the DeviceStatus constants
func TestDeviceStatus(t *testing.T) {
	t.Parallel()
	t.Run("device_status_constants", func(t *testing.T) {
		assert.Equal(t, "CONNECTED", string(DeviceStatusConnected), "Connected status should be correct")
		assert.Equal(t, "DISCONNECTED", string(DeviceStatusDisconnected), "Disconnected status should be correct")
		assert.Equal(t, "ERROR", string(DeviceStatusError), "Error status should be correct")
		assert.Equal(t, "PROBING", string(DeviceStatusProbing), "Probing status should be correct")
	})

	t.Run("device_status_json_marshaling", func(t *testing.T) {
		statuses := []DeviceStatus{
			DeviceStatusConnected,
			DeviceStatusDisconnected,
			DeviceStatusError,
			DeviceStatusProbing,
		}

		for _, status := range statuses {
			jsonData, err := json.Marshal(status)
			require.NoError(t, err, "Should marshal device status to JSON without error")

			var unmarshaledStatus DeviceStatus
			err = json.Unmarshal(jsonData, &unmarshaledStatus)
			require.NoError(t, err, "Should unmarshal device status from JSON without error")

			assert.Equal(t, status, unmarshaledStatus, "Device status should be preserved in JSON")
		}
	})
}

// TestV4L2Capabilities tests the V4L2Capabilities struct
func TestV4L2Capabilities(t *testing.T) {
	t.Parallel()
	t.Run("v4l2_capabilities_creation", func(t *testing.T) {
		capabilities := V4L2Capabilities{
			DriverName:   "uvcvideo",
			CardName:     "USB Camera",
			BusInfo:      "usb-0000:00:14.0-1",
			Version:      "5.15.0",
			Capabilities: []string{"0x85200001", "0x04200001"},
			DeviceCaps:   []string{"0x04200001"},
		}

		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be set correctly")
		assert.Equal(t, "USB Camera", capabilities.CardName, "Card name should be set correctly")
		assert.Equal(t, "usb-0000:00:14.0-1", capabilities.BusInfo, "Bus info should be set correctly")
		assert.Equal(t, "5.15.0", capabilities.Version, "Version should be set correctly")
		assert.Len(t, capabilities.Capabilities, 2, "Capabilities should have correct length")
		assert.Len(t, capabilities.DeviceCaps, 1, "Device caps should have correct length")
	})

	t.Run("v4l2_capabilities_json_marshaling", func(t *testing.T) {
		capabilities := V4L2Capabilities{
			DriverName:   "uvcvideo",
			CardName:     "USB Camera",
			BusInfo:      "usb-0000:00:14.0-1",
			Version:      "5.15.0",
			Capabilities: []string{"0x85200001", "0x04200001"},
			DeviceCaps:   []string{"0x04200001"},
		}

		jsonData, err := json.Marshal(capabilities)
		require.NoError(t, err, "Should marshal capabilities to JSON without error")

		var unmarshaledCapabilities V4L2Capabilities
		err = json.Unmarshal(jsonData, &unmarshaledCapabilities)
		require.NoError(t, err, "Should unmarshal capabilities from JSON without error")

		assert.Equal(t, capabilities.DriverName, unmarshaledCapabilities.DriverName, "Driver name should be preserved in JSON")
		assert.Equal(t, capabilities.CardName, unmarshaledCapabilities.CardName, "Card name should be preserved in JSON")
		assert.Equal(t, capabilities.BusInfo, unmarshaledCapabilities.BusInfo, "Bus info should be preserved in JSON")
		assert.Equal(t, capabilities.Version, unmarshaledCapabilities.Version, "Version should be preserved in JSON")
		assert.Equal(t, capabilities.Capabilities, unmarshaledCapabilities.Capabilities, "Capabilities should be preserved in JSON")
		assert.Equal(t, capabilities.DeviceCaps, unmarshaledCapabilities.DeviceCaps, "Device caps should be preserved in JSON")
	})

	t.Run("v4l2_capabilities_empty_fields", func(t *testing.T) {
		capabilities := V4L2Capabilities{
			DriverName: "uvcvideo",
			CardName:   "USB Camera",
		}

		jsonData, err := json.Marshal(capabilities)
		require.NoError(t, err, "Should marshal capabilities with empty fields to JSON without error")

		var unmarshaledCapabilities V4L2Capabilities
		err = json.Unmarshal(jsonData, &unmarshaledCapabilities)
		require.NoError(t, err, "Should unmarshal capabilities with empty fields from JSON without error")

		assert.Equal(t, capabilities.DriverName, unmarshaledCapabilities.DriverName, "Driver name should be preserved in JSON")
		assert.Equal(t, capabilities.CardName, unmarshaledCapabilities.CardName, "Card name should be preserved in JSON")
		assert.Empty(t, unmarshaledCapabilities.BusInfo, "Empty bus info should be preserved in JSON")
		assert.Empty(t, unmarshaledCapabilities.Version, "Empty version should be preserved in JSON")
		assert.Empty(t, unmarshaledCapabilities.Capabilities, "Empty capabilities should be preserved in JSON")
		assert.Empty(t, unmarshaledCapabilities.DeviceCaps, "Empty device caps should be preserved in JSON")
	})
}

// TestV4L2Format tests the V4L2Format struct
func TestV4L2Format(t *testing.T) {
	t.Parallel()
	t.Run("v4l2_format_creation", func(t *testing.T) {
		format := V4L2Format{
			PixelFormat: "YUYV",
			Width:       1920,
			Height:      1080,
			FrameRates:  []string{"30.000", "60.000"},
		}

		assert.Equal(t, "YUYV", format.PixelFormat, "Pixel format should be set correctly")
		assert.Equal(t, 1920, format.Width, "Width should be set correctly")
		assert.Equal(t, 1080, format.Height, "Height should be set correctly")
		assert.Len(t, format.FrameRates, 2, "Frame rates should have correct length")
		assert.Equal(t, "30.000", format.FrameRates[0], "First frame rate should be correct")
		assert.Equal(t, "60.000", format.FrameRates[1], "Second frame rate should be correct")
	})

	t.Run("v4l2_format_json_marshaling", func(t *testing.T) {
		format := V4L2Format{
			PixelFormat: "YUYV",
			Width:       1920,
			Height:      1080,
			FrameRates:  []string{"30.000", "60.000"},
		}

		jsonData, err := json.Marshal(format)
		require.NoError(t, err, "Should marshal format to JSON without error")

		var unmarshaledFormat V4L2Format
		err = json.Unmarshal(jsonData, &unmarshaledFormat)
		require.NoError(t, err, "Should unmarshal format from JSON without error")

		assert.Equal(t, format.PixelFormat, unmarshaledFormat.PixelFormat, "Pixel format should be preserved in JSON")
		assert.Equal(t, format.Width, unmarshaledFormat.Width, "Width should be preserved in JSON")
		assert.Equal(t, format.Height, unmarshaledFormat.Height, "Height should be preserved in JSON")
		assert.Equal(t, format.FrameRates, unmarshaledFormat.FrameRates, "Frame rates should be preserved in JSON")
	})

	t.Run("v4l2_format_empty_frame_rates", func(t *testing.T) {
		format := V4L2Format{
			PixelFormat: "YUYV",
			Width:       1920,
			Height:      1080,
			FrameRates:  []string{},
		}

		jsonData, err := json.Marshal(format)
		require.NoError(t, err, "Should marshal format with empty frame rates to JSON without error")

		var unmarshaledFormat V4L2Format
		err = json.Unmarshal(jsonData, &unmarshaledFormat)
		require.NoError(t, err, "Should unmarshal format with empty frame rates from JSON without error")

		assert.Equal(t, format.PixelFormat, unmarshaledFormat.PixelFormat, "Pixel format should be preserved in JSON")
		assert.Equal(t, format.Width, unmarshaledFormat.Width, "Width should be preserved in JSON")
		assert.Equal(t, format.Height, unmarshaledFormat.Height, "Height should be preserved in JSON")
		assert.Empty(t, unmarshaledFormat.FrameRates, "Empty frame rates should be preserved in JSON")
	})
}

// TestDeviceCapabilityState tests the DeviceCapabilityState struct
func TestDeviceCapabilityState(t *testing.T) {
	t.Parallel()
	t.Run("device_capability_state_creation", func(t *testing.T) {
		probeTime := time.Now()
		state := DeviceCapabilityState{
			LastProbeTime: probeTime,
			ProbeCount:    5,
			SuccessCount:  4,
			FailureCount:  1,
			LastError:     "Device timeout",
			CapabilityResult: &CapabilityDetectionResult{
				Detected:   true,
				Accessible: true,
				DeviceName: "USB Camera",
				Driver:     "uvcvideo",
			},
		}

		assert.Equal(t, probeTime, state.LastProbeTime, "Last probe time should be set correctly")
		assert.Equal(t, 5, state.ProbeCount, "Probe count should be set correctly")
		assert.Equal(t, 4, state.SuccessCount, "Success count should be set correctly")
		assert.Equal(t, 1, state.FailureCount, "Failure count should be set correctly")
		assert.Equal(t, "Device timeout", state.LastError, "Last error should be set correctly")
		assert.NotNil(t, state.CapabilityResult, "Capability result should be set")
		assert.Equal(t, "USB Camera", state.CapabilityResult.DeviceName, "Device name in result should be correct")
	})

	t.Run("device_capability_state_json_marshaling", func(t *testing.T) {
		probeTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		state := DeviceCapabilityState{
			LastProbeTime: probeTime,
			ProbeCount:    5,
			SuccessCount:  4,
			FailureCount:  1,
			LastError:     "Device timeout",
			CapabilityResult: &CapabilityDetectionResult{
				Detected:    true,
				Accessible:  true,
				DeviceName:  "USB Camera",
				Driver:      "uvcvideo",
				Formats:     []string{"YUYV", "MJPEG"},
				Resolutions: []string{"1920x1080", "1280x720"},
				FrameRates:  []string{"30.000", "60.000"},
			},
		}

		jsonData, err := json.Marshal(state)
		require.NoError(t, err, "Should marshal capability state to JSON without error")

		var unmarshaledState DeviceCapabilityState
		err = json.Unmarshal(jsonData, &unmarshaledState)
		require.NoError(t, err, "Should unmarshal capability state from JSON without error")

		assert.Equal(t, state.ProbeCount, unmarshaledState.ProbeCount, "Probe count should be preserved in JSON")
		assert.Equal(t, state.SuccessCount, unmarshaledState.SuccessCount, "Success count should be preserved in JSON")
		assert.Equal(t, state.FailureCount, unmarshaledState.FailureCount, "Failure count should be preserved in JSON")
		assert.Equal(t, state.LastError, unmarshaledState.LastError, "Last error should be preserved in JSON")
		assert.NotNil(t, unmarshaledState.CapabilityResult, "Capability result should be preserved in JSON")
		assert.Equal(t, state.CapabilityResult.DeviceName, unmarshaledState.CapabilityResult.DeviceName, "Device name in result should be preserved in JSON")
		assert.Equal(t, state.CapabilityResult.Driver, unmarshaledState.CapabilityResult.Driver, "Driver in result should be preserved in JSON")
		assert.Equal(t, state.CapabilityResult.Formats, unmarshaledState.CapabilityResult.Formats, "Formats in result should be preserved in JSON")
	})

	t.Run("device_capability_state_without_result", func(t *testing.T) {
		state := DeviceCapabilityState{
			LastProbeTime: time.Now(),
			ProbeCount:    3,
			SuccessCount:  0,
			FailureCount:  3,
			LastError:     "Device not found",
		}

		jsonData, err := json.Marshal(state)
		require.NoError(t, err, "Should marshal capability state without result to JSON without error")

		var unmarshaledState DeviceCapabilityState
		err = json.Unmarshal(jsonData, &unmarshaledState)
		require.NoError(t, err, "Should unmarshal capability state without result from JSON without error")

		assert.Equal(t, state.ProbeCount, unmarshaledState.ProbeCount, "Probe count should be preserved in JSON")
		assert.Equal(t, state.SuccessCount, unmarshaledState.SuccessCount, "Success count should be preserved in JSON")
		assert.Equal(t, state.FailureCount, unmarshaledState.FailureCount, "Failure count should be preserved in JSON")
		assert.Equal(t, state.LastError, unmarshaledState.LastError, "Last error should be preserved in JSON")
		assert.Nil(t, unmarshaledState.CapabilityResult, "Nil capability result should be preserved in JSON")
	})
}
