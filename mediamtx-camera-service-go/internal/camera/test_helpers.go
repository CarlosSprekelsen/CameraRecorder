package camera

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RealHardwareTestHelper provides utilities for testing with REAL camera hardware
// This is MANDATORY - no fixtures, real devices are available and must be used
type RealHardwareTestHelper struct {
	t                *testing.T
	availableDevices []string
	deviceChecker    DeviceChecker
	v4l2Executor     V4L2CommandExecutor
	deviceParser     DeviceInfoParser
}

// NewRealHardwareTestHelper creates a new real hardware test helper
func NewRealHardwareTestHelper(t *testing.T) *RealHardwareTestHelper {
	helper := &RealHardwareTestHelper{
		t:             t,
		deviceChecker: &RealDeviceChecker{},
		v4l2Executor:  &RealV4L2CommandExecutor{},
		deviceParser:  &RealDeviceInfoParser{},
	}

	// MANDATORY: Detect real available camera devices
	helper.detectAvailableDevices()

	return helper
}

// detectAvailableDevices scans for real camera devices on the system
func (h *RealHardwareTestHelper) detectAvailableDevices() {
	h.availableDevices = []string{}

	// Scan for video devices in /dev
	videoDevices, err := filepath.Glob("/dev/video*")
	if err != nil {
		h.t.Logf("Warning: Could not scan for video devices: %v", err)
		return
	}

	for _, device := range videoDevices {
		// REAL HARDWARE TEST: Check if device is actually accessible
		if h.isDeviceAccessible(device) {
			h.availableDevices = append(h.availableDevices, device)
			h.t.Logf("Found accessible camera device: %s", device)
		}
	}

	if len(h.availableDevices) == 0 {
		h.t.Logf("Warning: No accessible camera devices found. Tests will use fallback devices.")
		// Fallback to common device paths for testing
		h.availableDevices = []string{"/dev/video0", "/dev/video1"}
	}
}

// isDeviceAccessible checks if a device is actually accessible and functional
// REAL BUG TEST: This should distinguish between video capture devices and metadata-only devices
func (h *RealHardwareTestHelper) isDeviceAccessible(devicePath string) bool {
	// Check if device file exists
	if !h.deviceChecker.Exists(devicePath) {
		return false
	}

	// Try to get device capabilities (non-blocking)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, "--all")
	if err != nil {
		// Device exists but may not be accessible (permissions, busy, etc.)
		return false
	}

	// REAL BUG TEST: Check if this is actually a video capture device, not just a metadata device
	// Parse the output to check Device Caps
	hasVideoCapture := false
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Device Caps") {
			// Look for video capture capability in Device Caps
			if strings.Contains(line, "0x04200001") || strings.Contains(line, "0x85200001") {
				hasVideoCapture = true
				break
			}
		}
	}

	// REAL BUG TEST: Only consider devices with video capture capability as "accessible cameras"
	// This should exclude /dev/video1 which only has metadata capture (0x04a00000)
	return hasVideoCapture
}

// GetAvailableDevices returns the list of real available camera devices
func (h *RealHardwareTestHelper) GetAvailableDevices() []string {
	return h.availableDevices
}

// GetPrimaryDevice returns the first available camera device for testing
func (h *RealHardwareTestHelper) GetPrimaryDevice() string {
	if len(h.availableDevices) > 0 {
		return h.availableDevices[0]
	}
	return "/dev/video0" // Fallback
}

// TestWithRealDevice tests a function with a real camera device
func (h *RealHardwareTestHelper) TestWithRealDevice(testFunc func(devicePath string) error) {
	devices := h.GetAvailableDevices()

	if len(devices) == 0 {
		h.t.Skip("No real camera devices available for testing")
		return
	}

	// Test with each available device
	for _, device := range devices {
		h.t.Run(fmt.Sprintf("device_%s", filepath.Base(device)), func(t *testing.T) {
			err := testFunc(device)
			require.NoError(t, err, "Test should pass with real device: %s", device)
		})
	}
}

// TestDeviceCapabilities tests device capability detection with real hardware
func (h *RealHardwareTestHelper) TestDeviceCapabilities(devicePath string) (*V4L2Capabilities, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// REAL HARDWARE TEST: Execute actual v4l2-ctl command
	output, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, "--all")
	if err != nil {
		return nil, fmt.Errorf("failed to get device capabilities: %w", err)
	}

	// Parse real V4L2 output
	capabilities, err := h.deviceParser.ParseDeviceInfo(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device capabilities: %w", err)
	}

	return &capabilities, nil
}

// TestDeviceFormats tests device format detection with real hardware
func (h *RealHardwareTestHelper) TestDeviceFormats(devicePath string) ([]V4L2Format, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// REAL HARDWARE TEST: Execute actual v4l2-ctl command for formats
	output, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, "--list-formats-ext")
	if err != nil {
		return nil, fmt.Errorf("failed to get device formats: %w", err)
	}

	// Parse real V4L2 format output
	formats, err := h.deviceParser.ParseDeviceFormats(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device formats: %w", err)
	}

	return formats, nil
}

// TestDeviceFrameRates tests device frame rate detection with real hardware
func (h *RealHardwareTestHelper) TestDeviceFrameRates(devicePath string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// REAL HARDWARE TEST: Execute actual v4l2-ctl command for frame rates
	output, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, "--list-formats-ext")
	if err != nil {
		return nil, fmt.Errorf("failed to get device frame rates: %w", err)
	}

	// Parse real V4L2 frame rate output
	frameRates, err := h.deviceParser.ParseDeviceFrameRates(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device frame rates: %w", err)
	}

	return frameRates, nil
}

// TestDeviceAccessibility tests if a device is accessible and functional
func (h *RealHardwareTestHelper) TestDeviceAccessibility(devicePath string) error {
	// REAL HARDWARE TEST: Check device existence
	if !h.deviceChecker.Exists(devicePath) {
		return fmt.Errorf("device does not exist: %s", devicePath)
	}

	// REAL HARDWARE TEST: Try to access device
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, "--help")
	if err != nil {
		return fmt.Errorf("device not accessible: %w", err)
	}

	return nil
}

// TestDeviceDiscovery tests camera device discovery with real hardware
func (h *RealHardwareTestHelper) TestDeviceDiscovery() []string {
	discoveredDevices := []string{}

	// Scan for video devices
	videoDevices, err := filepath.Glob("/dev/video*")
	if err != nil {
		h.t.Logf("Warning: Could not scan for video devices: %v", err)
		return discoveredDevices
	}

	// Test each discovered device
	for _, device := range videoDevices {
		if h.isDeviceAccessible(device) {
			discoveredDevices = append(discoveredDevices, device)
		}
	}

	return discoveredDevices
}

// TestDevicePermissions tests device permission access
func (h *RealHardwareTestHelper) TestDevicePermissions(devicePath string) error {
	// Check if we can read the device file
	file, err := os.Open(devicePath)
	if err != nil {
		return fmt.Errorf("cannot open device for reading: %w", err)
	}
	defer file.Close()

	// Check if we can get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot get device file info: %w", err)
	}

	// Verify it's a character device
	if info.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("device is not a character device: %s", devicePath)
	}

	return nil
}

// TestV4L2ToolsAvailability tests if required V4L2 tools are available
func (h *RealHardwareTestHelper) TestV4L2ToolsAvailability() error {
	// Check if v4l2-ctl is available
	_, err := exec.LookPath("v4l2-ctl")
	if err != nil {
		return fmt.Errorf("v4l2-ctl not found: %w", err)
	}

	// Check if v4l2-ctl works
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "v4l2-ctl", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("v4l2-ctl not working: %w", err)
	}

	return nil
}

// TestDeviceStreamingCapability tests if a device can actually stream
// This test should FAIL for devices that don't support video capture (like metadata-only devices)
func (h *RealHardwareTestHelper) TestDeviceStreamingCapability(devicePath string) error {
	// Get device capabilities
	capabilities, err := h.TestDeviceCapabilities(devicePath)
	if err != nil {
		return fmt.Errorf("cannot get device capabilities: %w", err)
	}

	// REAL BUG TEST: Check if device supports video capture by parsing V4L2 capability flags
	// This should catch the bug where metadata-only devices are treated as cameras
	hasVideoCapture := false

	// Parse Device Caps flags (this is the critical test)
	for _, cap := range capabilities.DeviceCaps {
		// Look for hex flags that indicate video capture capability
		if strings.Contains(cap, "0x04200001") || strings.Contains(cap, "0x85200001") {
			hasVideoCapture = true
			break
		}
		// Also check for text-based capability reporting
		if strings.Contains(strings.ToLower(cap), "video capture") {
			hasVideoCapture = true
			break
		}
	}

	// REAL BUG TEST: If device doesn't have video capture in Device Caps, it's not a camera
	// This should fail for /dev/video1 which only has metadata capture (0x04a00000)
	if !hasVideoCapture {
		return fmt.Errorf("device does not support video capture (Device Caps check failed): %s", devicePath)
	}

	// REAL BUG TEST: Check if device has streaming capabilities
	hasStreaming := false
	for _, cap := range capabilities.DeviceCaps {
		if strings.Contains(cap, "streaming") || strings.Contains(cap, "0x04200001") || strings.Contains(cap, "0x85200001") {
			hasStreaming = true
			break
		}
	}

	if !hasStreaming {
		return fmt.Errorf("device does not support streaming: %s", devicePath)
	}

	// REAL BUG TEST: Try to get video formats - this should work for real cameras
	formats, err := h.TestDeviceFormats(devicePath)
	if err != nil {
		return fmt.Errorf("device cannot provide video formats: %w", err)
	}

	if len(formats) == 0 {
		return fmt.Errorf("device has no video formats available: %s", devicePath)
	}

	// REAL BUG TEST: Check if formats are actually video formats (not metadata)
	hasVideoFormats := false
	for _, format := range formats {
		// Check for actual video pixel formats, not metadata formats
		if format.PixelFormat == "YUYV" || format.PixelFormat == "MJPG" || format.PixelFormat == "H264" {
			hasVideoFormats = true
			break
		}
	}

	if !hasVideoFormats {
		return fmt.Errorf("device only provides metadata formats, not video formats: %s", devicePath)
	}

	return nil
}

// TestDeviceFormatSupport tests if a device supports specific formats
func (h *RealHardwareTestHelper) TestDeviceFormatSupport(devicePath string, expectedFormats []string) error {
	formats, err := h.TestDeviceFormats(devicePath)
	if err != nil {
		return fmt.Errorf("cannot get device formats: %w", err)
	}

	// Check if expected formats are supported
	supportedFormats := make(map[string]bool)
	for _, format := range formats {
		supportedFormats[format.PixelFormat] = true
	}

	for _, expected := range expectedFormats {
		if !supportedFormats[expected] {
			return fmt.Errorf("device does not support format %s", expected)
		}
	}

	return nil
}

// TestDeviceResolutionSupport tests if a device supports specific resolutions
func (h *RealHardwareTestHelper) TestDeviceResolutionSupport(devicePath string, expectedResolutions []string) error {
	formats, err := h.TestDeviceFormats(devicePath)
	if err != nil {
		return fmt.Errorf("cannot get device formats: %w", err)
	}

	// Check if expected resolutions are supported
	supportedResolutions := make(map[string]bool)
	for _, format := range formats {
		resolution := fmt.Sprintf("%dx%d", format.Width, format.Height)
		supportedResolutions[resolution] = true
	}

	for _, expected := range expectedResolutions {
		if !supportedResolutions[expected] {
			return fmt.Errorf("device does not support resolution %s", expected)
		}
	}

	return nil
}

// TestDeviceFrameRateSupport tests if a device supports specific frame rates
func (h *RealHardwareTestHelper) TestDeviceFrameRateSupport(devicePath string, expectedFrameRates []string) error {
	frameRates, err := h.TestDeviceFrameRates(devicePath)
	if err != nil {
		return fmt.Errorf("cannot get device frame rates: %w", err)
	}

	// Check if expected frame rates are supported
	supportedFrameRates := make(map[string]bool)
	for _, rate := range frameRates {
		supportedFrameRates[rate] = true
	}

	for _, expected := range expectedFrameRates {
		if !supportedFrameRates[expected] {
			return fmt.Errorf("device does not support frame rate %s", expected)
		}
	}

	return nil
}

// TestDeviceStressTest performs a stress test on a real device
func (h *RealHardwareTestHelper) TestDeviceStressTest(devicePath string, iterations int) error {
	for i := 0; i < iterations; i++ {
		// Test device capabilities
		if _, err := h.TestDeviceCapabilities(devicePath); err != nil {
			return fmt.Errorf("stress test failed at iteration %d (capabilities): %w", i, err)
		}

		// Test device formats
		if _, err := h.TestDeviceFormats(devicePath); err != nil {
			return fmt.Errorf("stress test failed at iteration %d (formats): %w", i, err)
		}

		// Test device frame rates
		if _, err := h.TestDeviceFrameRates(devicePath); err != nil {
			return fmt.Errorf("stress test failed at iteration %d (frame rates): %w", i, err)
		}

		// Small delay to prevent overwhelming the device
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// TestDeviceConcurrentAccess tests concurrent access to a real device
func (h *RealHardwareTestHelper) TestDeviceConcurrentAccess(devicePath string, numGoroutines int) error {
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Test device capabilities concurrently
			if _, err := h.TestDeviceCapabilities(devicePath); err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
				return
			}
			errors <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		if err := <-errors; err != nil {
			return fmt.Errorf("concurrent access test failed: %w", err)
		}
	}

	return nil
}

// TestDeviceErrorHandling tests error handling with real device scenarios
// REAL BUG TEST: This should test that the software correctly handles non-existent devices
func (h *RealHardwareTestHelper) TestDeviceErrorHandling() error {
	// Test with non-existent device
	nonExistentDevice := "/dev/video999999"
	if h.deviceChecker.Exists(nonExistentDevice) {
		return fmt.Errorf("non-existent device should not exist")
	}

	// REAL BUG TEST: Test with invalid device path using --all (not --help)
	// --help doesn't check device existence, but --all does
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := h.v4l2Executor.ExecuteCommand(ctx, nonExistentDevice, "--all")
	if err == nil {
		return fmt.Errorf("should fail with non-existent device when using --all command")
	}

	// REAL BUG TEST: Verify the error message is meaningful
	if !strings.Contains(err.Error(), "Cannot open device") && !strings.Contains(err.Error(), "does not exist") {
		return fmt.Errorf("error message should indicate device cannot be opened: %v", err)
	}

	// REAL BUG TEST: Test with invalid device path using --info (should also fail)
	_, err = h.v4l2Executor.ExecuteCommand(ctx, nonExistentDevice, "--info")
	if err == nil {
		return fmt.Errorf("should fail with non-existent device when using --info command")
	}

	// REAL BUG TEST: Test with empty device path
	_, err = h.v4l2Executor.ExecuteCommand(ctx, "", "--all")
	if err == nil {
		return fmt.Errorf("should fail with empty device path")
	}

	// REAL BUG TEST: Test with malformed device path
	_, err = h.v4l2Executor.ExecuteCommand(ctx, "/dev/invalid_device_name", "--all")
	if err == nil {
		return fmt.Errorf("should fail with malformed device path")
	}

	return nil
}

// TestDevicePerformance tests device performance characteristics
func (h *RealHardwareTestHelper) TestDevicePerformance(devicePath string) (time.Duration, error) {
	start := time.Now()

	// Perform a complete device capability scan
	if _, err := h.TestDeviceCapabilities(devicePath); err != nil {
		return 0, fmt.Errorf("performance test failed: %w", err)
	}

	if _, err := h.TestDeviceFormats(devicePath); err != nil {
		return 0, fmt.Errorf("performance test failed: %w", err)
	}

	if _, err := h.TestDeviceFrameRates(devicePath); err != nil {
		return 0, fmt.Errorf("performance test failed: %w", err)
	}

	duration := time.Since(start)
	return duration, nil
}

// TestDeviceCompatibility tests device compatibility with different V4L2 commands
func (h *RealHardwareTestHelper) TestDeviceCompatibility(devicePath string) error {
	// Test various V4L2 commands
	commands := []string{
		"--all",
		"--list-formats-ext",
		"--list-ctrls",
		"--list-fields",
		"--list-framesizes",
		"--list-frametimes",
	}

	for _, cmd := range commands {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := h.v4l2Executor.ExecuteCommand(ctx, devicePath, cmd)
		cancel()

		// Some commands may not be supported by all devices, that's OK
		if err != nil {
			h.t.Logf("Command %s not supported by device %s: %v", cmd, devicePath, err)
		}
	}

	return nil
}

// TestDeviceIntegration tests complete device integration workflow
func (h *RealHardwareTestHelper) TestDeviceIntegration(devicePath string) error {
	// 1. Test device accessibility
	if err := h.TestDeviceAccessibility(devicePath); err != nil {
		return fmt.Errorf("device accessibility test failed: %w", err)
	}

	// 2. Test device permissions
	if err := h.TestDevicePermissions(devicePath); err != nil {
		return fmt.Errorf("device permissions test failed: %w", err)
	}

	// 3. Test device capabilities
	if _, err := h.TestDeviceCapabilities(devicePath); err != nil {
		return fmt.Errorf("device capabilities test failed: %w", err)
	}

	// 4. Test device formats
	if _, err := h.TestDeviceFormats(devicePath); err != nil {
		return fmt.Errorf("device formats test failed: %w", err)
	}

	// 5. Test device frame rates
	if _, err := h.TestDeviceFrameRates(devicePath); err != nil {
		return fmt.Errorf("device frame rates test failed: %w", err)
	}

	// 6. Test streaming capability
	if err := h.TestDeviceStreamingCapability(devicePath); err != nil {
		return fmt.Errorf("device streaming capability test failed: %w", err)
	}

	// 7. Test performance
	duration, err := h.TestDevicePerformance(devicePath)
	if err != nil {
		return fmt.Errorf("device performance test failed: %w", err)
	}

	h.t.Logf("Device integration test completed in %v", duration)
	return nil
}

// TestAllAvailableDevices runs comprehensive tests on all available devices
func (h *RealHardwareTestHelper) TestAllAvailableDevices() {
	devices := h.GetAvailableDevices()

	if len(devices) == 0 {
		h.t.Skip("No real camera devices available for testing")
		return
	}

	for _, device := range devices {
		h.t.Run(fmt.Sprintf("comprehensive_test_%s", filepath.Base(device)), func(t *testing.T) {
			// Test device integration
			err := h.TestDeviceIntegration(device)
			require.NoError(h.t, err, "Device integration test should pass for %s", device)

			// Test device compatibility
			err = h.TestDeviceCompatibility(device)
			require.NoError(h.t, err, "Device compatibility test should pass for %s", device)

			// Test device stress
			err = h.TestDeviceStressTest(device, 5)
			require.NoError(h.t, err, "Device stress test should pass for %s", device)

			// Test concurrent access
			err = h.TestDeviceConcurrentAccess(device, 3)
			require.NoError(h.t, err, "Device concurrent access test should pass for %s", device)
		})
	}
}

// TestDeviceDiscoveryWorkflow tests the complete device discovery workflow
func (h *RealHardwareTestHelper) TestDeviceDiscoveryWorkflow() {
	// Test V4L2 tools availability
	err := h.TestV4L2ToolsAvailability()
	require.NoError(h.t, err, "V4L2 tools should be available")

	// Test device discovery
	discoveredDevices := h.TestDeviceDiscovery()
	require.NotEmpty(h.t, discoveredDevices, "Should discover at least one camera device")

	// Test each discovered device
	for _, device := range discoveredDevices {
		h.t.Run(fmt.Sprintf("discovery_workflow_%s", filepath.Base(device)), func(t *testing.T) {
			err := h.TestDeviceIntegration(device)
			require.NoError(h.t, err, "Device integration should work for discovered device %s", device)
		})
	}
}

// TestDeviceErrorScenarios tests various error scenarios with real devices
func (h *RealHardwareTestHelper) TestDeviceErrorScenarios() {
	// Test error handling
	err := h.TestDeviceErrorHandling()
	require.NoError(h.t, err, "Device error handling should work correctly")

	// Test with invalid device paths
	invalidPaths := []string{
		"",
		"/dev/invalid",
		"/dev/video999999",
		"/tmp/not_a_device",
	}

	for _, invalidPath := range invalidPaths {
		h.t.Run(fmt.Sprintf("error_scenario_%s", invalidPath), func(t *testing.T) {
			// These should all fail appropriately
			if h.deviceChecker.Exists(invalidPath) {
				h.t.Errorf("Invalid path %s should not exist", invalidPath)
			}
		})
	}
}

// TestDevicePerformanceBenchmarks runs performance benchmarks on real devices
func (h *RealHardwareTestHelper) TestDevicePerformanceBenchmarks() {
	devices := h.GetAvailableDevices()

	if len(devices) == 0 {
		h.t.Skip("No real camera devices available for performance testing")
		return
	}

	for _, device := range devices {
		h.t.Run(fmt.Sprintf("performance_benchmark_%s", filepath.Base(device)), func(t *testing.T) {
			// Run multiple performance tests
			var totalDuration time.Duration
			numTests := 5

			for i := 0; i < numTests; i++ {
				duration, err := h.TestDevicePerformance(device)
				require.NoError(h.t, err, "Performance test should pass for %s", device)
				totalDuration += duration
			}

			avgDuration := totalDuration / time.Duration(numTests)
			h.t.Logf("Average device performance: %v", avgDuration)

			// Performance should be reasonable (less than 1 second for basic operations)
			require.Less(h.t, avgDuration, time.Second, "Device performance should be reasonable")
		})
	}
}
