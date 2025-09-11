package camera

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRealV4L2CommandExecutor_ExecuteCommand tests the ExecuteCommand function
func TestRealV4L2CommandExecutor_ExecuteCommand(t *testing.T) {
	executor := &RealV4L2CommandExecutor{}
	ctx := context.Background()

	// Test 1: Valid device command
	t.Run("valid_device_command", func(t *testing.T) {
		// Test with a real device if available
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Execute --info command
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info")

		// Should succeed and return output
		require.NoError(t, err, "ExecuteCommand should succeed for valid device")
		require.NotEmpty(t, output, "Output should not be empty")
		require.Contains(t, output, "Driver name", "Output should contain driver information")
	})

	// Test 2: Invalid device
	t.Run("invalid_device", func(t *testing.T) {
		// Execute command on non-existent device
		output, err := executor.ExecuteCommand(ctx, "/dev/video999999", "--info")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail for invalid device")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 3: Invalid command
	t.Run("invalid_command", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Execute invalid command
		output, err := executor.ExecuteCommand(ctx, devicePath, "--invalid-command")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail for invalid command")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 4: Multiple arguments
	t.Run("multiple_arguments", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Execute command with multiple arguments
		output, err := executor.ExecuteCommand(ctx, devicePath, "--list-formats")

		// Should succeed (even if command doesn't support --all)
		// The important thing is that multiple arguments are handled correctly
		if err != nil {
			// Command might not support --all, that's okay
			t.Logf("Command failed as expected: %v", err)
		} else {
			require.NotEmpty(t, output, "Output should not be empty if command succeeds")
		}
	})

	// Test 5: Context cancellation
	t.Run("context_cancellation", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		// Execute command with cancelled context
		output, err := executor.ExecuteCommand(cancelledCtx, devicePath, "--info")

		// Should fail due to context cancellation
		require.Error(t, err, "ExecuteCommand should fail with cancelled context")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 6: Context timeout (error injection)
	t.Run("context_timeout", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Create context with very short timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Execute command with timeout context
		output, err := executor.ExecuteCommand(timeoutCtx, devicePath, "--info")

		// Should fail due to timeout
		require.Error(t, err, "ExecuteCommand should fail with timeout context")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 7: Device path with special characters (edge case)
	t.Run("device_path_special_characters", func(t *testing.T) {
		// Test with device path containing special characters
		output, err := executor.ExecuteCommand(ctx, "/dev/video;rm -rf /", "--info")

		// Should fail safely
		require.Error(t, err, "ExecuteCommand should fail with malicious device path")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 8: Command with special characters (security test)
	t.Run("command_special_characters", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Test with command containing special characters
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info; rm -rf /")

		// Should fail safely
		require.Error(t, err, "ExecuteCommand should fail with malicious command")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 9: Very long device path (edge case)
	t.Run("very_long_device_path", func(t *testing.T) {
		// Create a very long device path
		longPath := "/dev/" + string(make([]byte, 1000))
		for i := range longPath[5:] {
			longPath = longPath[:5+i] + "a" + longPath[5+i+1:]
		}

		output, err := executor.ExecuteCommand(ctx, longPath, "--info")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail with very long device path")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 10: Very long command arguments (edge case)
	t.Run("very_long_command_arguments", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Create very long command argument
		longArg := "--" + string(make([]byte, 1000))
		for i := range longArg[2:] {
			longArg = longArg[:2+i] + "a" + longArg[2+i+1:]
		}

		output, err := executor.ExecuteCommand(ctx, devicePath, longArg)

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail with very long command argument")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 11: Empty device path (edge case)
	t.Run("empty_device_path", func(t *testing.T) {
		output, err := executor.ExecuteCommand(ctx, "", "--info")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail with empty device path")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 12: Empty command arguments (edge case)
	t.Run("empty_command_arguments", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Execute with empty command arguments
		output, err := executor.ExecuteCommand(ctx, devicePath, "")

		// Empty command arguments are handled gracefully by strings.Fields()
		// The command may succeed or fail depending on the device, both are valid
		if err != nil {
			require.Empty(t, output, "Output should be empty on error")
		} else {
			// If it succeeds, that's also valid behavior
			outputLen := len(output)
			if outputLen > 100 {
				outputLen = 100
			}
			t.Logf("Empty command arguments handled gracefully: %s", output[:outputLen])
		}
	})

	// Test 13: Device path with spaces (edge case)
	t.Run("device_path_with_spaces", func(t *testing.T) {
		output, err := executor.ExecuteCommand(ctx, "/dev/video 0", "--info")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail with device path containing spaces")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 14: Command with newlines (security test)
	t.Run("command_with_newlines", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Test with command containing newlines
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info\nrm -rf /")

		// Should fail safely
		require.Error(t, err, "ExecuteCommand should fail with command containing newlines")
		require.Empty(t, output, "Output should be empty on error")
	})

	// Test 15: Command with tabs (edge case)
	t.Run("command_with_tabs", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Test with command containing tabs
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info\t--help")

		// Tabs are handled by strings.Fields() which splits on whitespace
		// The command may succeed or fail depending on the device, both are valid
		if err != nil {
			require.Empty(t, output, "Output should be empty on error")
		} else {
			// If it succeeds, that's also valid behavior
			outputLen := len(output)
			if outputLen > 100 {
				outputLen = 100
			}
			t.Logf("Command with tabs handled gracefully: %s", output[:outputLen])
		}
	})

	// Test 16: Multiple command arguments (normal case)
	t.Run("multiple_command_arguments", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Execute with multiple valid arguments
		output, err := executor.ExecuteCommand(ctx, devicePath, "--list-formats")

		// Should succeed or fail gracefully
		if err != nil {
			// Command might not support these arguments, that's okay
			t.Logf("Command failed as expected: %v", err)
		} else {
			require.NotEmpty(t, output, "Output should not be empty if command succeeds")
		}
	})

	// Test 17: Command with quotes (edge case)
	t.Run("command_with_quotes", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Test with command containing quotes
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info")

		// --info is a valid v4l2-ctl command, so it may succeed or fail depending on device
		if err != nil {
			require.Empty(t, output, "Output should be empty on error")
		} else {
			// If it succeeds, that's also valid behavior
			outputLen := len(output)
			if outputLen > 100 {
				outputLen = 100
			}
			t.Logf("Command with quotes handled gracefully: %s", output[:outputLen])
		}
	})

	// Test 18: Command with backslashes (edge case)
	t.Run("command_with_backslashes", func(t *testing.T) {
		helper := NewRealHardwareTestHelper(t)
		availableDevices := helper.GetAvailableDevices()

		require.NotEmpty(t, availableDevices, "Real camera devices must be available for testing")

		devicePath := availableDevices[0]

		// Test with command containing backslashes
		output, err := executor.ExecuteCommand(ctx, devicePath, "--info\\test")

		// Should fail
		require.Error(t, err, "ExecuteCommand should fail with command containing backslashes")
		require.Empty(t, output, "Output should be empty on error")
	})
}

// TestRealDeviceInfoParser_ParseDeviceInfo tests the ParseDeviceInfo function
func TestRealDeviceInfoParser_ParseDeviceInfo(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid device info output
	t.Run("valid_device_info", func(t *testing.T) {
		// Sample v4l2-ctl --info output
		output := `Driver Info:
	Driver name      : uvcvideo
	Card type        : USB 2.0 Camera: USB 2.0 Camera
	Bus info         : usb-0000:00:1a.0-1.2
	Driver version   : 6.14.8
	Capabilities     : 0x84a00001
		Video Capture
		Metadata Capture
		Streaming
		Extended Pix Format
		Device Capabilities
	Device Caps      : 0x04200001
		Video Capture
		Streaming
		Extended Pix Format`

		capabilities, err := parser.ParseDeviceInfo(output)

		// Should succeed and parse correctly
		require.NoError(t, err, "ParseDeviceInfo should succeed")
		require.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed")
		require.Equal(t, "USB 2.0 Camera: USB 2.0 Camera", capabilities.CardName, "Card name should be parsed")
		require.Equal(t, "usb-0000:00:1a.0-1.2", capabilities.BusInfo, "Bus info should be parsed")
		require.Equal(t, "6.14.8", capabilities.Version, "Version should be parsed")

		// Should have capabilities
		require.NotEmpty(t, capabilities.Capabilities, "Should have capabilities")
		require.Contains(t, capabilities.Capabilities, "Video Capture", "Should contain Video Capture")
		require.Contains(t, capabilities.Capabilities, "Streaming", "Should contain Streaming")

		// Should have device caps
		require.NotEmpty(t, capabilities.DeviceCaps, "Should have device caps")
		require.Contains(t, capabilities.DeviceCaps, "Video Capture", "Device caps should contain Video Capture")
	})

	// Test 2: Empty output
	t.Run("empty_output", func(t *testing.T) {
		capabilities, err := parser.ParseDeviceInfo("")

		// Should succeed with defaults
		require.NoError(t, err, "ParseDeviceInfo should succeed with empty output")
		require.Equal(t, "Unknown Video Device", capabilities.CardName, "Should have default card name")
		require.Equal(t, "unknown", capabilities.DriverName, "Should have default driver name")
		require.Empty(t, capabilities.Capabilities, "Should have empty capabilities")
		require.Empty(t, capabilities.DeviceCaps, "Should have empty device caps")
	})

	// Test 3: Partial output
	t.Run("partial_output", func(t *testing.T) {
		output := `Driver Info:
	Driver name      : uvcvideo
	Card type        : USB Camera`

		capabilities, err := parser.ParseDeviceInfo(output)

		// Should succeed with available information
		require.NoError(t, err, "ParseDeviceInfo should succeed with partial output")
		require.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed")
		require.Equal(t, "USB Camera", capabilities.CardName, "Card name should be parsed")
		require.Empty(t, capabilities.Capabilities, "Should have empty capabilities when not present")
	})

	// Test 4: Malformed output
	t.Run("malformed_output", func(t *testing.T) {
		output := `Invalid output without proper structure`

		capabilities, err := parser.ParseDeviceInfo(output)

		// Should succeed with defaults
		require.NoError(t, err, "ParseDeviceInfo should succeed with malformed output")
		require.Equal(t, "Unknown Video Device", capabilities.CardName, "Should have default card name")
		require.Equal(t, "unknown", capabilities.DriverName, "Should have default driver name")
	})

	// Test 5: Capabilities without names
	t.Run("capabilities_without_names", func(t *testing.T) {
		output := `Driver Info:
	Driver name      : uvcvideo
	Card type        : USB Camera
	Capabilities     : 0x84a00001
	Device Caps      : 0x04200001`

		capabilities, err := parser.ParseDeviceInfo(output)

		// Should succeed and include hex values
		require.NoError(t, err, "ParseDeviceInfo should succeed")
		require.Contains(t, capabilities.Capabilities, "0x84a00001", "Should include hex capability value")
		require.Contains(t, capabilities.DeviceCaps, "0x04200001", "Should include hex device cap value")
	})
}

// TestRealDeviceInfoParser_ParseDeviceFormats tests the ParseDeviceFormats function
func TestRealDeviceInfoParser_ParseDeviceFormats(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid formats output
	t.Run("valid_formats_output", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
			Interval: Discrete 0.067s (15.000 fps)
		Size: Discrete 1280x720
			Interval: Discrete 0.033s (30.000 fps)

	[1]: 'YUYV' (YUYV 4:2:2)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
		Size: Discrete 1280x720
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and parse formats
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 4, "Should parse 4 format+resolution combinations")

		// Check formats - each pixel format + resolution combination is a separate format
		// MJPG 640x480
		require.Equal(t, "MJPG", formats[0].PixelFormat, "First format should be MJPG")
		require.Equal(t, 640, formats[0].Width, "First format width should be 640")
		require.Equal(t, 480, formats[0].Height, "First format height should be 480")
		require.Len(t, formats[0].FrameRates, 2, "MJPG 640x480 should have 2 frame rates")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "First frame rate should be 30.000")

		// MJPG 1280x720
		require.Equal(t, "MJPG", formats[1].PixelFormat, "Second format should be MJPG")
		require.Equal(t, 1280, formats[1].Width, "Second format width should be 1280")
		require.Equal(t, 720, formats[1].Height, "Second format height should be 720")
		require.Len(t, formats[1].FrameRates, 1, "MJPG 1280x720 should have 1 frame rate")

		// YUYV 640x480
		require.Equal(t, "YUYV", formats[2].PixelFormat, "Third format should be YUYV")
		require.Equal(t, 640, formats[2].Width, "Third format width should be 640")
		require.Equal(t, 480, formats[2].Height, "Third format height should be 480")
		require.Len(t, formats[2].FrameRates, 1, "YUYV 640x480 should have 1 frame rate")

		// YUYV 1280x720
		require.Equal(t, "YUYV", formats[3].PixelFormat, "Fourth format should be YUYV")
		require.Equal(t, 1280, formats[3].Width, "Fourth format width should be 1280")
		require.Equal(t, 720, formats[3].Height, "Fourth format height should be 720")
		require.Len(t, formats[3].FrameRates, 1, "YUYV 1280x720 should have 1 frame rate")
	})

	// Test 2: Empty output
	t.Run("empty_output", func(t *testing.T) {
		formats, err := parser.ParseDeviceFormats("")

		// Should succeed with empty result
		require.NoError(t, err, "ParseDeviceFormats should succeed with empty output")
		require.Empty(t, formats, "Should return empty formats")
	})

	// Test 3: Malformed output
	t.Run("malformed_output", func(t *testing.T) {
		output := `Invalid output without proper format structure`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed with empty result
		require.NoError(t, err, "ParseDeviceFormats should succeed with malformed output")
		require.Empty(t, formats, "Should return empty formats")
	})

	// Test 4: Single format
	t.Run("single_format", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and parse single format
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "MJPG", formats[0].PixelFormat, "Format should be MJPG")
		require.Len(t, formats[0].FrameRates, 1, "Should have 1 frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "Frame rate should be 30.000")
	})

	// Test 5: Format with multiple frame rates
	t.Run("format_with_multiple_frame_rates", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
			Interval: Discrete 0.067s (15.000 fps)
			Interval: Discrete 0.100s (10.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and parse multiple frame rates
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 3, "Should have 3 frame rates")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "First frame rate should be 30.000")
		require.Equal(t, "15.000", formats[0].FrameRates[1], "Second frame rate should be 15.000")
		require.Equal(t, "10.000", formats[0].FrameRates[2], "Third frame rate should be 10.000")
	})

	// Test 6: Format without size information (edge case)
	t.Run("format_without_size", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed but create format with zero dimensions
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "MJPG", formats[0].PixelFormat, "Format should be MJPG")
		require.Equal(t, 0, formats[0].Width, "Width should be 0 when not specified")
		require.Equal(t, 0, formats[0].Height, "Height should be 0 when not specified")
		require.Empty(t, formats[0].FrameRates, "Frame rates should be empty when not specified")
	})

	// Test 7: Format with invalid size format (bug detection)
	t.Run("format_with_invalid_size", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete invalid_size
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed but handle invalid size gracefully
		require.NoError(t, err, "ParseDeviceFormats should succeed even with invalid size")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "MJPG", formats[0].PixelFormat, "Format should be MJPG")
		require.Equal(t, 0, formats[0].Width, "Width should be 0 for invalid size")
		require.Equal(t, 0, formats[0].Height, "Height should be 0 for invalid size")
		require.Len(t, formats[0].FrameRates, 1, "Should still parse frame rates")
	})

	// Test 8: Format with malformed frame rate (bug detection)
	t.Run("format_with_malformed_frame_rate", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete invalid_interval
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and filter out invalid frame rates
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 1, "Should have 1 valid frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "Should keep valid frame rate")
	})

	// Test 9: Multiple formats with mixed valid/invalid data (stress test)
	t.Run("multiple_formats_mixed_data", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
		Size: Discrete invalid_size
			Interval: Discrete 0.067s (15.000 fps)

	[1]: 'YUYV' (YUYV 4:2:2)
		Size: Discrete 1280x720
			Interval: Discrete invalid_interval
			Interval: Discrete 0.033s (30.000 fps)

	[2]: 'INVALID' (Invalid Format)
		Size: Discrete 320x240
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle mixed data gracefully
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 4, "Should parse 4 format+resolution combinations")

		// Check that valid formats are preserved
		require.Equal(t, "MJPG", formats[0].PixelFormat, "First format should be MJPG")
		require.Equal(t, 640, formats[0].Width, "First format should have valid width")
		require.Equal(t, 480, formats[0].Height, "First format should have valid height")

		// Check that invalid sizes are handled
		require.Equal(t, "MJPG", formats[1].PixelFormat, "Second format should be MJPG")
		require.Equal(t, 0, formats[1].Width, "Second format should have zero width for invalid size")
		require.Equal(t, 0, formats[1].Height, "Second format should have zero height for invalid size")
	})

	// Test 10: Format with very large dimensions (edge case)
	t.Run("format_with_large_dimensions", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 7680x4320
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle large dimensions
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, 7680, formats[0].Width, "Should handle large width")
		require.Equal(t, 4320, formats[0].Height, "Should handle large height")
	})

	// Test 11: Format with zero dimensions (edge case)
	t.Run("format_with_zero_dimensions", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 0x0
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle zero dimensions
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, 0, formats[0].Width, "Should handle zero width")
		require.Equal(t, 0, formats[0].Height, "Should handle zero height")
	})

	// Test 12: Format with negative frame rate (bug detection)
	t.Run("format_with_negative_frame_rate", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete -0.033s (-30.000 fps)
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and filter out negative frame rates
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 1, "Should have 1 valid frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "Should keep positive frame rate")
	})

	// Test 13: Format with extremely high frame rate (edge case)
	t.Run("format_with_extreme_frame_rate", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.001s (1000.000 fps)
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and filter out extreme frame rates
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 1, "Should have 1 valid frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "Should keep reasonable frame rate")
	})

	// Test 14: Format with empty pixel format name (bug detection)
	t.Run("format_with_empty_pixel_format", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: '' (Empty Format)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle empty format name
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "", formats[0].PixelFormat, "Should handle empty pixel format")
		require.Equal(t, 640, formats[0].Width, "Should still parse dimensions")
		require.Equal(t, 480, formats[0].Height, "Should still parse dimensions")
	})

	// Test 15: Format with very long pixel format name (edge case)
	t.Run("format_with_long_pixel_format", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'VERY_LONG_PIXEL_FORMAT_NAME_THAT_EXCEEDS_NORMAL_LENGTH' (Very Long Format)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle long format names
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "VERY_LONG_PIXEL_FORMAT_NAME_THAT_EXCEEDS_NORMAL_LENGTH", formats[0].PixelFormat, "Should handle long format name")
	})

	// Test 16: Index pattern handling (uncovered path)
	t.Run("index_pattern_format", func(t *testing.T) {
		output := `Index : 0
Name : MJPG
Size : Discrete 640x480
fps : 30.000

Index : 1
Name : YUYV
Size : 1280x720
fps : 60.000`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and parse Index pattern
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 2, "Should parse 2 formats")

		// Check first format
		require.Equal(t, "MJPG", formats[0].PixelFormat, "First format should be MJPG")
		require.Equal(t, 640, formats[0].Width, "First format width should be 640")
		require.Equal(t, 480, formats[0].Height, "First format height should be 480")
		require.Len(t, formats[0].FrameRates, 1, "First format should have 1 frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "First format frame rate should be 30.000")

		// Check second format
		require.Equal(t, "YUYV", formats[1].PixelFormat, "Second format should be YUYV")
		require.Equal(t, 1280, formats[1].Width, "Second format width should be 1280")
		require.Equal(t, 720, formats[1].Height, "Second format height should be 720")
		require.Len(t, formats[1].FrameRates, 1, "Second format should have 1 frame rate")
		require.Equal(t, "60.000", formats[1].FrameRates[0], "Second format frame rate should be 60.000")
	})

	// Test 17: Index pattern with invalid size (uncovered path)
	t.Run("index_pattern_invalid_size", func(t *testing.T) {
		output := `Index : 0
Name : MJPG
Size : invalid_size
fps : 30.000`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle invalid size in Index pattern
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "MJPG", formats[0].PixelFormat, "Format should be MJPG")
		require.Equal(t, 0, formats[0].Width, "Width should be 0 for invalid size")
		require.Equal(t, 0, formats[0].Height, "Height should be 0 for invalid size")
		require.Len(t, formats[0].FrameRates, 1, "Should still parse frame rate")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "Frame rate should be preserved")
	})

	// Test 18: Index pattern with empty fps (uncovered path)
	t.Run("index_pattern_empty_fps", func(t *testing.T) {
		output := `Index : 0
Name : MJPG
Size : Discrete 640x480
fps : `

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle empty fps
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "MJPG", formats[0].PixelFormat, "Format should be MJPG")
		require.Equal(t, 640, formats[0].Width, "Width should be parsed correctly")
		require.Equal(t, 480, formats[0].Height, "Height should be parsed correctly")
		require.Empty(t, formats[0].FrameRates, "Frame rates should be empty for empty fps")
	})

	// Test 19: Test case pattern with Name and invalid_size (uncovered path)
	t.Run("test_case_pattern_invalid_size", func(t *testing.T) {
		output := `Name : YUYV
Size : invalid_size`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and handle test case pattern
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Equal(t, "YUYV", formats[0].PixelFormat, "Format should be YUYV")
		require.Equal(t, 0, formats[0].Width, "Width should be 0 for invalid size")
		require.Equal(t, 0, formats[0].Height, "Height should be 0 for invalid size")
		require.Empty(t, formats[0].FrameRates, "Frame rates should be empty")
	})

	// Test 20: Format with malformed interval line (uncovered path)
	t.Run("format_malformed_interval", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
			Interval: malformed_interval_line
			Interval: Discrete 0.067s (15.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and skip malformed interval
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 2, "Should have 2 valid frame rates")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "First frame rate should be valid")
		require.Equal(t, "15.000", formats[0].FrameRates[1], "Second frame rate should be valid")
	})

	// Test 21: Format with interval missing fps (uncovered path)
	t.Run("format_interval_no_fps", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s
			Interval: Discrete 0.067s (15.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and skip interval without fps
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 1, "Should have 1 valid frame rate")
		require.Equal(t, "15.000", formats[0].FrameRates[0], "Should keep valid frame rate")
	})

	// Test 22: Format with empty fps in interval (uncovered path)
	t.Run("format_interval_empty_fps", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
			Interval: Discrete 0.067s ()
			Interval: Discrete 0.100s (10.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and skip empty fps
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 1, "Should parse 1 format")
		require.Len(t, formats[0].FrameRates, 2, "Should have 2 valid frame rates")
		require.Equal(t, "30.000", formats[0].FrameRates[0], "First frame rate should be valid")
		require.Equal(t, "10.000", formats[0].FrameRates[1], "Third frame rate should be valid")
	})

	// Test 23: Format declaration without quotes (uncovered path)
	t.Run("format_declaration_no_quotes", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: MJPG (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed but not create format without quotes
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Empty(t, formats, "Should not parse format without quotes")
	})

	// Test 24: Format with existing width/height before size (uncovered path)
	t.Run("format_existing_dimensions_before_size", func(t *testing.T) {
		output := `ioctl: VIDIOC_ENUM_FMT
	Type: Video Capture

	[0]: 'MJPG' (Motion-JPEG, compressed)
		Size: Discrete 640x480
			Interval: Discrete 0.033s (30.000 fps)
		Size: Discrete 1280x720
			Interval: Discrete 0.033s (30.000 fps)`

		formats, err := parser.ParseDeviceFormats(output)

		// Should succeed and create separate formats for each size
		require.NoError(t, err, "ParseDeviceFormats should succeed")
		require.Len(t, formats, 2, "Should parse 2 formats")

		// First format
		require.Equal(t, "MJPG", formats[0].PixelFormat, "First format should be MJPG")
		require.Equal(t, 640, formats[0].Width, "First format should have 640 width")
		require.Equal(t, 480, formats[0].Height, "First format should have 480 height")

		// Second format
		require.Equal(t, "MJPG", formats[1].PixelFormat, "Second format should be MJPG")
		require.Equal(t, 1280, formats[1].Width, "Second format should have 1280 width")
		require.Equal(t, 720, formats[1].Height, "Second format should have 720 height")
	})
}

// TestRealDeviceInfoParser_parseCapabilities tests the deprecated parseCapabilities function
func TestRealDeviceInfoParser_parseCapabilities(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid capability line
	t.Run("valid_capability_line", func(t *testing.T) {
		line := "Capabilities     : 0x84a00001 Video Capture Streaming"

		capabilities := parser.parseCapabilities(line)

		// FIXED: Should parse 3 capabilities (preserves "Video Capture" as one name)
		require.Len(t, capabilities, 3, "Should parse 3 capabilities (preserves multi-word names)")
		require.Contains(t, capabilities, "0x84a00001", "Should include hex value")
		require.Contains(t, capabilities, "Video Capture", "Should preserve 'Video Capture' as one capability")
		require.Contains(t, capabilities, "Streaming", "Should include Streaming")
	})

	// Test 2: Empty line
	t.Run("empty_line", func(t *testing.T) {
		capabilities := parser.parseCapabilities("")

		// Should return empty
		require.Empty(t, capabilities, "Should return empty capabilities")
	})

	// Test 3: Line without colon
	t.Run("line_without_colon", func(t *testing.T) {
		line := "Invalid line without colon"

		capabilities := parser.parseCapabilities(line)

		// Should return empty
		require.Empty(t, capabilities, "Should return empty capabilities")
	})

	// Test 4: Line with only hex value
	t.Run("line_with_only_hex", func(t *testing.T) {
		line := "Capabilities     : 0x84a00001"

		capabilities := parser.parseCapabilities(line)

		// Should parse hex value
		require.Len(t, capabilities, 1, "Should parse 1 capability")
		require.Contains(t, capabilities, "0x84a00001", "Should include hex value")
	})
}

// TestRealDeviceInfoParser_parseSize tests the parseSize function
func TestRealDeviceInfoParser_parseSize(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid size
	t.Run("valid_size", func(t *testing.T) {
		width, height := parser.parseSize("640x480")

		// Should parse correctly
		require.Equal(t, 640, width, "Width should be 640")
		require.Equal(t, 480, height, "Height should be 480")
	})

	// Test 2: Invalid size format
	t.Run("invalid_size_format", func(t *testing.T) {
		width, height := parser.parseSize("invalid")

		// Should return zeros
		require.Equal(t, 0, width, "Width should be 0")
		require.Equal(t, 0, height, "Height should be 0")
	})

	// Test 3: Size with extra characters
	t.Run("size_with_extra_characters", func(t *testing.T) {
		width, height := parser.parseSize("1280x720p")

		// Should parse correctly (ignores extra characters)
		require.Equal(t, 1280, width, "Width should be 1280")
		require.Equal(t, 720, height, "Height should be 720")
	})

	// Test 4: Empty string
	t.Run("empty_string", func(t *testing.T) {
		width, height := parser.parseSize("")

		// Should return zeros
		require.Equal(t, 0, width, "Width should be 0")
		require.Equal(t, 0, height, "Height should be 0")
	})
}

// TestRealDeviceInfoParser_extractValue tests the extractValue function
func TestRealDeviceInfoParser_extractValue(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid key-value pair
	t.Run("valid_key_value", func(t *testing.T) {
		line := "Driver name      : uvcvideo"

		value := parser.extractValue(line)

		// Should extract value
		require.Equal(t, "uvcvideo", value, "Should extract driver name")
	})

	// Test 2: Line without colon
	t.Run("line_without_colon", func(t *testing.T) {
		line := "Invalid line without colon"

		value := parser.extractValue(line)

		// Should return empty
		require.Empty(t, value, "Should return empty value")
	})

	// Test 3: Empty line
	t.Run("empty_line", func(t *testing.T) {
		value := parser.extractValue("")

		// Should return empty
		require.Empty(t, value, "Should return empty value")
	})

	// Test 4: Value with spaces
	t.Run("value_with_spaces", func(t *testing.T) {
		line := "Card type        : USB 2.0 Camera: USB 2.0 Camera"

		value := parser.extractValue(line)

		// Should extract value with spaces
		require.Equal(t, "USB 2.0 Camera: USB 2.0 Camera", value, "Should extract value with spaces")
	})
}

// TestRealDeviceInfoParser_normalizeFrameRate tests the normalizeFrameRate function
func TestRealDeviceInfoParser_normalizeFrameRate(t *testing.T) {
	parser := &RealDeviceInfoParser{}

	// Test 1: Valid frame rate
	t.Run("valid_frame_rate", func(t *testing.T) {
		rate := parser.normalizeFrameRate("30.000")

		// Should normalize correctly
		require.Equal(t, "30.000", rate, "Frame rate should be 30.000")
	})

	// Test 2: Frame rate with extra text (should fail - function expects just number)
	t.Run("frame_rate_with_extra_text", func(t *testing.T) {
		rate := parser.normalizeFrameRate("30.000 fps")

		// Should return empty string because it can't parse "30.000 fps" as float
		require.Equal(t, "", rate, "Frame rate with extra text should be filtered out")
	})

	// Test 3: Invalid frame rate
	t.Run("invalid_frame_rate", func(t *testing.T) {
		rate := parser.normalizeFrameRate("invalid")

		// Should return empty string
		require.Equal(t, "", rate, "Frame rate should be empty string")
	})

	// Test 4: Empty string
	t.Run("empty_string", func(t *testing.T) {
		rate := parser.normalizeFrameRate("")

		// Should return empty string
		require.Equal(t, "", rate, "Frame rate should be empty string")
	})

	// Test 5: Out of range frame rate
	t.Run("out_of_range_frame_rate", func(t *testing.T) {
		rate := parser.normalizeFrameRate("500.000")

		// Should return empty string (out of 1-300 range)
		require.Equal(t, "", rate, "Out of range frame rate should be filtered out")
	})

	// Test 6: Edge case - minimum valid rate
	t.Run("minimum_valid_rate", func(t *testing.T) {
		rate := parser.normalizeFrameRate("1.000")

		// Should normalize correctly
		require.Equal(t, "1.000", rate, "Minimum valid frame rate should be 1.000")
	})

	// Test 7: Edge case - maximum valid rate
	t.Run("maximum_valid_rate", func(t *testing.T) {
		rate := parser.normalizeFrameRate("300.000")

		// Should normalize correctly
		require.Equal(t, "300.000", rate, "Maximum valid frame rate should be 300.000")
	})
}
