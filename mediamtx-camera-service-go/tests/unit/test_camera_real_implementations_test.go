//go:build unit
// +build unit

/*
Camera Real Implementations Unit Tests

Requirements Coverage:
- REQ-CAM-001: Camera device detection and enumeration
- REQ-CAM-002: Camera capability probing and validation
- REQ-CAM-003: Real V4L2 device interaction
- REQ-CAM-004: Device information parsing accuracy
- REQ-CAM-005: Error handling for real device operations
- REQ-CAM-006: Format and capability detection

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Real Component Usage: V4L2 devices, file system, command execution
*/

package unit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
)

// TestRealDeviceChecker tests the RealDeviceChecker implementation
func TestRealDeviceChecker(t *testing.T) {
	checker := &camera.RealDeviceChecker{}

	t.Run("exists_real_file", func(t *testing.T) {
		// Test with a file that should exist
		exists := checker.Exists("/dev/null")
		assert.True(t, exists, "Real device checker should detect existing file")
	})

	t.Run("exists_nonexistent_file", func(t *testing.T) {
		// Test with a file that should not exist
		exists := checker.Exists("/dev/nonexistent_device_12345")
		assert.False(t, exists, "Real device checker should not detect non-existent file")
	})

	t.Run("exists_current_directory", func(t *testing.T) {
		// Test with current directory
		currentDir, err := os.Getwd()
		require.NoError(t, err, "Should be able to get current directory")

		exists := checker.Exists(currentDir)
		assert.True(t, exists, "Real device checker should detect existing directory")
	})

	t.Run("exists_temp_file", func(t *testing.T) {
		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "test_device_checker")
		require.NoError(t, err, "Should be able to create temp file")
		defer os.Remove(tempFile.Name())

		exists := checker.Exists(tempFile.Name())
		assert.True(t, exists, "Real device checker should detect temporary file")
	})

	t.Run("exists_empty_path", func(t *testing.T) {
		// Test with empty path
		exists := checker.Exists("")
		assert.False(t, exists, "Real device checker should not detect empty path")
	})

	t.Run("exists_relative_path", func(t *testing.T) {
		// Test with relative path
		exists := checker.Exists(".")
		assert.True(t, exists, "Real device checker should detect current directory")
	})
}

// TestRealV4L2CommandExecutor tests the RealV4L2CommandExecutor implementation
func TestRealV4L2CommandExecutor(t *testing.T) {
	executor := &camera.RealV4L2CommandExecutor{}

	t.Run("execute_command_with_timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with a command that should work (help)
		output, err := executor.ExecuteCommand(ctx, "/dev/video0", "--help")

		// This may fail if v4l2-ctl is not installed, but we test the interface
		if err != nil {
			assert.Contains(t, err.Error(), "executable file not found", "Should fail gracefully if v4l2-ctl not found")
		} else {
			assert.NotEmpty(t, output, "Command output should not be empty")
			assert.Contains(t, output, "v4l2-ctl", "Output should contain v4l2-ctl information")
		}
	})

	t.Run("execute_command_invalid_device", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with invalid device path
		output, err := executor.ExecuteCommand(ctx, "/dev/invalid_device", "--info")

		// This should fail gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "No such file", "Should fail with appropriate error for invalid device")
		} else {
			// If it doesn't fail, output should be empty or contain error info
			t.Logf("Command output: %s", output)
		}
	})

	t.Run("execute_command_cancelled_context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Test with cancelled context
		output, err := executor.ExecuteCommand(ctx, "/dev/video0", "--info")

		assert.Error(t, err, "Should fail with cancelled context")
		assert.Empty(t, output, "Output should be empty with cancelled context")
	})

	t.Run("execute_command_with_args", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with multiple arguments
		output, err := executor.ExecuteCommand(ctx, "/dev/video0", "--list-devices")

		// This may fail if v4l2-ctl is not installed
		if err != nil {
			assert.Contains(t, err.Error(), "executable file not found", "Should fail gracefully if v4l2-ctl not found")
		} else {
			assert.NotEmpty(t, output, "Command output should not be empty")
		}
	})
}

// TestRealDeviceInfoParser tests the RealDeviceInfoParser implementation
func TestRealDeviceInfoParser(t *testing.T) {
	parser := &camera.RealDeviceInfoParser{}

	t.Run("parse_device_info_valid_output", func(t *testing.T) {
		validOutput := `
Driver name       : uvcvideo
Card type         : USB Camera
Bus info          : usb-0000:00:14.0-1
Driver version    : 5.15.0
Capabilities      : 0x85200001
Device Caps       : 0x04200001
`

		capabilities, err := parser.ParseDeviceInfo(validOutput)
		require.NoError(t, err, "Should parse valid device info without error")

		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed correctly")
		assert.Equal(t, "USB Camera", capabilities.CardName, "Card name should be parsed correctly")
		assert.Equal(t, "usb-0000:00:14.0-1", capabilities.BusInfo, "Bus info should be parsed correctly")
		assert.Equal(t, "5.15.0", capabilities.Version, "Version should be parsed correctly")
		assert.Len(t, capabilities.Capabilities, 1, "Should parse capabilities")
		assert.Len(t, capabilities.DeviceCaps, 1, "Should parse device caps")
	})

	t.Run("parse_device_info_missing_fields", func(t *testing.T) {
		incompleteOutput := `
Driver name       : uvcvideo
Bus info          : usb-0000:00:14.0-1
`

		capabilities, err := parser.ParseDeviceInfo(incompleteOutput)
		require.NoError(t, err, "Should parse incomplete device info without error")

		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed correctly")
		assert.Equal(t, "Unknown Video Device", capabilities.CardName, "Should use default card name")
		assert.Equal(t, "usb-0000:00:14.0-1", capabilities.BusInfo, "Bus info should be parsed correctly")
		assert.Equal(t, "", capabilities.Version, "Version should be empty")
	})

	t.Run("parse_device_info_empty_output", func(t *testing.T) {
		capabilities, err := parser.ParseDeviceInfo("")
		require.NoError(t, err, "Should handle empty output without error")

		assert.Equal(t, "Unknown Video Device", capabilities.CardName, "Should use default card name")
		assert.Equal(t, "unknown", capabilities.DriverName, "Should use default driver name")
		assert.Empty(t, capabilities.Capabilities, "Capabilities should be empty")
		assert.Empty(t, capabilities.DeviceCaps, "Device caps should be empty")
	})

	t.Run("parse_device_formats_valid_output", func(t *testing.T) {
		validOutput := `
Index : 0
Type  : Video Capture
Name  : YUYV
Size  : 1920x1080
fps   : 30.000

Index : 1
Type  : Video Capture
Name  : MJPEG
Size  : 1280x720
fps   : 60.000
`

		formats, err := parser.ParseDeviceFormats(validOutput)
		require.NoError(t, err, "Should parse valid format output without error")

		assert.Len(t, formats, 2, "Should parse two formats")

		// Check first format
		assert.Equal(t, "YUYV", formats[0].PixelFormat, "First format pixel format should be correct")
		assert.Equal(t, 1920, formats[0].Width, "First format width should be correct")
		assert.Equal(t, 1080, formats[0].Height, "First format height should be correct")
		assert.Len(t, formats[0].FrameRates, 1, "First format should have one frame rate")
		assert.Equal(t, "30.000", formats[0].FrameRates[0], "First format frame rate should be correct")

		// Check second format
		assert.Equal(t, "MJPEG", formats[1].PixelFormat, "Second format pixel format should be correct")
		assert.Equal(t, 1280, formats[1].Width, "Second format width should be correct")
		assert.Equal(t, 720, formats[1].Height, "Second format height should be correct")
		assert.Len(t, formats[1].FrameRates, 1, "Second format should have one frame rate")
		assert.Equal(t, "60.000", formats[1].FrameRates[0], "Second format frame rate should be correct")
	})

	t.Run("parse_device_formats_empty_output", func(t *testing.T) {
		formats, err := parser.ParseDeviceFormats("")
		require.NoError(t, err, "Should handle empty format output without error")

		assert.Empty(t, formats, "Should return empty formats list")
	})

	t.Run("parse_device_formats_invalid_size", func(t *testing.T) {
		invalidOutput := `
Index : 0
Type  : Video Capture
Name  : YUYV
Size  : invalid_size
fps   : 30.000
`

		formats, err := parser.ParseDeviceFormats(invalidOutput)
		require.NoError(t, err, "Should handle invalid size without error")

		assert.Len(t, formats, 1, "Should parse one format")
		assert.Equal(t, 0, formats[0].Width, "Width should be 0 for invalid size")
		assert.Equal(t, 0, formats[0].Height, "Height should be 0 for invalid size")
	})

	t.Run("parse_device_frame_rates_valid_output", func(t *testing.T) {
		validOutput := `
30.000 fps
60.000 FPS
Frame rate: 25.0
30 Hz
1920x1080@60
Interval: [1/30]
[1/25]
1/30 s
30 frames per second
rate: 24
fps: 29.97
`

		frameRates, err := parser.ParseDeviceFrameRates(validOutput)
		require.NoError(t, err, "Should parse valid frame rate output without error")

		expectedRates := []string{"30.000", "60.000", "25.0", "30", "60", "30", "25", "30", "30", "24", "29.97"}
		assert.ElementsMatch(t, expectedRates, frameRates, "Should parse all frame rate patterns")
	})

	t.Run("parse_device_frame_rates_empty_output", func(t *testing.T) {
		frameRates, err := parser.ParseDeviceFrameRates("")
		require.NoError(t, err, "Should handle empty frame rate output without error")

		assert.Empty(t, frameRates, "Should return empty frame rates list")
	})

	t.Run("parse_device_frame_rates_no_matches", func(t *testing.T) {
		noRatesOutput := `
This is some text without any frame rates
Just random content
No fps information here
`

		frameRates, err := parser.ParseDeviceFrameRates(noRatesOutput)
		require.NoError(t, err, "Should handle output without frame rates without error")

		assert.Empty(t, frameRates, "Should return empty frame rates list")
	})

	t.Run("parse_device_frame_rates_duplicate_rates", func(t *testing.T) {
		duplicateOutput := `
30.000 fps
30.000 FPS
30 Hz
30 frames per second
`

		frameRates, err := parser.ParseDeviceFrameRates(duplicateOutput)
		require.NoError(t, err, "Should handle duplicate frame rates without error")

		assert.Len(t, frameRates, 1, "Should deduplicate frame rates")
		assert.Equal(t, "30.000", frameRates[0], "Should keep one instance of duplicate rate")
	})
}

// TestRealDeviceInfoParserIntegration tests integration of parsing methods
func TestRealDeviceInfoParserIntegration(t *testing.T) {
	parser := &camera.RealDeviceInfoParser{}

	t.Run("parse_complete_device_info", func(t *testing.T) {
		completeOutput := `
Driver name       : uvcvideo
Card type         : USB Camera
Bus info          : usb-0000:00:14.0-1
Driver version    : 5.15.0
Capabilities      : 0x85200001 0x04200001
Device Caps       : 0x04200001
`

		capabilities, err := parser.ParseDeviceInfo(completeOutput)
		require.NoError(t, err, "Should parse complete device info without error")

		assert.Equal(t, "uvcvideo", capabilities.DriverName, "Driver name should be parsed correctly")
		assert.Equal(t, "USB Camera", capabilities.CardName, "Card name should be parsed correctly")
		assert.Equal(t, "usb-0000:00:14.0-1", capabilities.BusInfo, "Bus info should be parsed correctly")
		assert.Equal(t, "5.15.0", capabilities.Version, "Version should be parsed correctly")
		assert.Len(t, capabilities.Capabilities, 2, "Should parse multiple capabilities")
		assert.Len(t, capabilities.DeviceCaps, 1, "Should parse device caps")
	})

	t.Run("parse_complete_formats_with_sizes", func(t *testing.T) {
		completeFormatsOutput := `
Index : 0
Type  : Video Capture
Name  : YUYV
Size  : 1920x1080
fps   : 30.000

Index : 1
Type  : Video Capture
Name  : MJPEG
Size  : 1280x720
fps   : 60.000

Index : 2
Type  : Video Capture
Name  : RGB24
Size  : 640x480
fps   : 25.000
`

		formats, err := parser.ParseDeviceFormats(completeFormatsOutput)
		require.NoError(t, err, "Should parse complete formats without error")

		assert.Len(t, formats, 3, "Should parse three formats")

		// Verify size parsing works correctly
		assert.Equal(t, 1920, formats[0].Width, "First format width should be parsed correctly")
		assert.Equal(t, 1080, formats[0].Height, "First format height should be parsed correctly")
		assert.Equal(t, 1280, formats[1].Width, "Second format width should be parsed correctly")
		assert.Equal(t, 720, formats[1].Height, "Second format height should be parsed correctly")
		assert.Equal(t, 640, formats[2].Width, "Third format width should be parsed correctly")
		assert.Equal(t, 480, formats[2].Height, "Third format height should be parsed correctly")
	})
}
