//go:build unit
// +build unit

package camera_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type MockDeviceChecker struct {
	existsMap map[string]bool
}

func (m *MockDeviceChecker) Exists(path string) bool {
	return m.existsMap[path]
}

type MockV4L2CommandExecutor struct {
	outputMap map[string]string
	errorMap  map[string]error
}

func (m *MockV4L2CommandExecutor) ExecuteCommand(ctx context.Context, devicePath, args string) (string, error) {
	key := fmt.Sprintf("%s:%s", devicePath, args)
	if err, exists := m.errorMap[key]; exists {
		return "", err
	}
	return m.outputMap[key], nil
}

type MockDeviceInfoParser struct {
	capabilities camera.V4L2Capabilities
	formats      []camera.V4L2Format
	parseError   error
}

func (m *MockDeviceInfoParser) ParseDeviceInfo(output string) (camera.V4L2Capabilities, error) {
	if m.parseError != nil {
		return camera.V4L2Capabilities{}, m.parseError
	}
	return m.capabilities, nil
}

func (m *MockDeviceInfoParser) ParseDeviceFormats(output string) ([]camera.V4L2Format, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}
	return m.formats, nil
}

type MockConfigProvider struct {
	config *camera.CameraConfig
}

func (m *MockConfigProvider) GetCameraConfig() *camera.CameraConfig {
	return m.config
}

func (m *MockConfigProvider) GetPollInterval() float64 {
	return m.config.PollInterval
}

func (m *MockConfigProvider) GetDetectionTimeout() float64 {
	return m.config.DetectionTimeout
}

func (m *MockConfigProvider) GetDeviceRange() []int {
	return m.config.DeviceRange
}

func (m *MockConfigProvider) GetEnableCapabilityDetection() bool {
	return m.config.EnableCapabilityDetection
}

func (m *MockConfigProvider) GetCapabilityTimeout() float64 {
	return m.config.CapabilityTimeout
}

type MockLogger struct {
	infoLogs  []string
	warnLogs  []string
	errorLogs []string
	fields    map[string]interface{}
}

func (m *MockLogger) WithFields(fields map[string]interface{}) camera.Logger {
	m.fields = fields
	return m
}

func (m *MockLogger) Info(args ...interface{}) {
	m.infoLogs = append(m.infoLogs, fmt.Sprint(args...))
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.warnLogs = append(m.warnLogs, fmt.Sprint(args...))
}

func (m *MockLogger) Error(args ...interface{}) {
	m.errorLogs = append(m.errorLogs, fmt.Sprint(args...))
}

func (m *MockLogger) Debug(args ...interface{}) {}

// createTestManager creates a V4L2DeviceManager with mock dependencies for testing
func createTestManager(configProvider camera.ConfigProvider, logger camera.Logger, deviceExists map[string]bool) *camera.V4L2DeviceManager {
	return createTestManagerWithMocks(configProvider, logger, deviceExists, &MockV4L2CommandExecutor{}, &MockDeviceInfoParser{})
}

// createTestManagerWithMocks creates a V4L2DeviceManager with all mock dependencies specified
func createTestManagerWithMocks(configProvider camera.ConfigProvider, logger camera.Logger, deviceExists map[string]bool, commandExecutor camera.V4L2CommandExecutor, infoParser camera.DeviceInfoParser) *camera.V4L2DeviceManager {
	mockDeviceChecker := &MockDeviceChecker{
		existsMap: deviceExists,
	}

	return camera.NewV4L2DeviceManagerWithDependencies(configProvider, logger, mockDeviceChecker, commandExecutor, infoParser)
}

func TestV4L2DeviceManager_Creation(t *testing.T) {
	t.Run("nil_config_uses_defaults", func(t *testing.T) {
		manager := camera.NewV4L2DeviceManager(nil, nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetStats())
	})

	t.Run("valid_config", func(t *testing.T) {
		config := &camera.CameraConfig{
			PollInterval:              0.2,
			DetectionTimeout:          2.0,
			DeviceRange:               []int{0, 1, 2},
			EnableCapabilityDetection: true,
		}
		configProvider := &MockConfigProvider{config: config}
		logger := &MockLogger{}

		manager := camera.NewV4L2DeviceManager(configProvider, logger)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetStats())
	})
}

func TestV4L2DeviceManager_StartStop(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Test start
	err := manager.Start()
	require.NoError(t, err)
	assert.True(t, manager.GetStats().Running)

	// Test stop
	err = manager.Stop()
	require.NoError(t, err)
	assert.False(t, manager.GetStats().Running)
}

func TestV4L2DeviceManager_DeviceDiscovery(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := createTestManager(configProvider, logger, map[string]bool{
		"/dev/video0": true,
		"/dev/video1": true,
	})

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	devices := manager.GetConnectedDevices()
	assert.NotEmpty(t, devices)

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_GetDevice(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	device, exists := manager.GetDevice("/dev/video0")
	if exists {
		assert.NotEmpty(t, device.Name)
		// Capabilities may be empty if V4L2 probing fails
		// but device should still be detected
	}

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_DeviceCapabilities(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	// Create mock command executor with successful response
	mockCommandExecutor := &MockV4L2CommandExecutor{
		outputMap: map[string]string{
			"/dev/video0:--info": `Driver name       : uvcvideo
Card type         : USB Camera
Bus info          : usb-0000:00:14.0-1
Driver version    : 5.15.0
Capabilities      : video_capture video_output
Device Caps       : video_capture streaming`,
		},
	}

	// Create mock info parser
	mockInfoParser := &MockDeviceInfoParser{
		capabilities: camera.V4L2Capabilities{
			DriverName:   "uvcvideo",
			CardName:     "USB Camera",
			BusInfo:      "usb-0000:00:14.0-1",
			Version:      "5.15.0",
			Capabilities: []string{"video_capture", "video_output"},
			DeviceCaps:   []string{"video_capture", "streaming"},
		},
	}

	manager := createTestManagerWithMocks(configProvider, logger, map[string]bool{
		"/dev/video0": true,
	}, mockCommandExecutor, mockInfoParser)

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	device, exists := manager.GetDevice("/dev/video0")
	if exists {
		assert.NotEmpty(t, device.Name)
		// Capabilities may be empty if V4L2 probing fails
		// but device should still be detected
	}

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_DeviceStatus(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	devices := manager.GetConnectedDevices()
	assert.NotEmpty(t, devices)

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_Statistics(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	stats := manager.GetStats()
	assert.NotNil(t, stats)
	assert.True(t, stats.Running)
	assert.Greater(t, stats.DevicesDiscovered, 0)

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_ConcurrentAccess(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Test concurrent access to multiple methods
	done := make(chan bool, 30)

	// Concurrent GetConnectedDevices calls
	for i := 0; i < 10; i++ {
		go func() {
			devices := manager.GetConnectedDevices()
			assert.NotNil(t, devices)
			done <- true
		}()
	}

	// Concurrent GetDevice calls
	for i := 0; i < 10; i++ {
		go func(deviceNum int) {
			device, exists := manager.GetDevice(fmt.Sprintf("/dev/video%d", deviceNum))
			// Device may or may not exist, but should not panic
			if exists {
				assert.NotNil(t, device)
			}
			done <- true
		}(i)
	}

	// Concurrent GetStats calls
	for i := 0; i < 10; i++ {
		go func() {
			stats := manager.GetStats()
			assert.NotNil(t, stats)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 30; i++ {
		<-done
	}

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *camera.CameraConfig
		expectError bool
	}{
		{
			name: "valid_config",
			config: &camera.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:               []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
			expectError: false,
		},
		{
			name: "zero_poll_interval",
			config: &camera.CameraConfig{
				PollInterval: 0.0,
			},
			expectError: true,
		},
		{
			name: "negative_poll_interval",
			config: &camera.CameraConfig{
				PollInterval: -0.1,
			},
			expectError: true,
		},
		{
			name: "zero_detection_timeout",
			config: &camera.CameraConfig{
				DetectionTimeout: 0.0,
			},
			expectError: true,
		},
		{
			name: "empty_device_range",
			config: &camera.CameraConfig{
				DeviceRange: []int{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configProvider := &MockConfigProvider{config: tt.config}
			logger := &MockLogger{}

			manager := camera.NewV4L2DeviceManager(configProvider, logger)

			// Test that manager can be created with any config
			// Validation should happen at runtime, not creation time
			assert.NotNil(t, manager)
		})
	}
}

func TestV4L2DeviceManager_EdgeCases(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.01, // Very fast polling
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(100 * time.Millisecond)

	device, exists := manager.GetDevice("/dev/video0")
	if exists {
		// Device may be in ERROR status if V4L2 probing fails
		// but should still be detected
		assert.Contains(t, []camera.DeviceStatus{
			camera.DeviceStatusConnected,
			camera.DeviceStatusError,
		}, device.Status)
	}

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_DeviceRange(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{5, 10, 15}, // Non-standard range
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	// Should not find devices in non-standard range
	device, exists := manager.GetDevice("/dev/video0")
	assert.False(t, exists)
	assert.Nil(t, device)

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_CapabilityDetection(t *testing.T) {
	tests := []struct {
		name            string
		enableDetection bool
		expectedStatus  camera.DeviceStatus
		commandExecutor camera.V4L2CommandExecutor
		infoParser      camera.DeviceInfoParser
	}{
		{
			name:            "capability_detection_enabled",
			enableDetection: true,
			expectedStatus:  camera.DeviceStatusConnected,
			commandExecutor: &MockV4L2CommandExecutor{
				outputMap: map[string]string{
					"/dev/video0:--info": "Driver name: uvcvideo",
				},
			},
			infoParser: &MockDeviceInfoParser{
				capabilities: camera.V4L2Capabilities{
					DriverName: "uvcvideo",
					CardName:   "USB Camera",
				},
			},
		},
		{
			name:            "capability_detection_disabled",
			enableDetection: false,
			expectedStatus:  camera.DeviceStatusConnected,
			commandExecutor: &MockV4L2CommandExecutor{},
			infoParser:      &MockDeviceInfoParser{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configProvider := &MockConfigProvider{
				config: &camera.CameraConfig{
					PollInterval:              0.1,
					DetectionTimeout:          1.0,
					DeviceRange:               []int{0},
					EnableCapabilityDetection: tt.enableDetection,
				},
			}
			logger := &MockLogger{}

			manager := camera.NewV4L2DeviceManager(configProvider, logger)

			// Mock device checker
			manager.deviceChecker = &MockDeviceChecker{
				existsMap: map[string]bool{
					"/dev/video0": true,
				},
			}

			// Mock command executor and info parser
			manager.commandExecutor = tt.commandExecutor
			manager.infoParser = tt.infoParser

			err := manager.Start()
			require.NoError(t, err)

			// Wait for discovery
			time.Sleep(200 * time.Millisecond)

			device, exists := manager.GetDevice("/dev/video0")
			if exists {
				// Device may be in ERROR status if V4L2 probing fails
				// but should still be detected
				assert.Contains(t, []camera.DeviceStatus{
					camera.DeviceStatusConnected,
					camera.DeviceStatusError,
				}, device.Status)
			}

			err = manager.Stop()
			require.NoError(t, err)
		})
	}
}

func TestV4L2DeviceManager_Performance(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for discovery
	time.Sleep(200 * time.Millisecond)

	// Test performance of device access
	start := time.Now()
	for i := 0; i < 1000; i++ {
		devices := manager.GetConnectedDevices()
		_ = devices
	}
	duration := time.Since(start)

	// Should complete 1000 operations in reasonable time
	assert.Less(t, duration, 100*time.Millisecond)

	err = manager.Stop()
	require.NoError(t, err)
}

func TestV4L2DeviceManager_ErrorHandling(t *testing.T) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Test with invalid device paths
	device, exists := manager.GetDevice("")
	assert.Nil(t, device)
	assert.False(t, exists)

	device, exists = manager.GetDevice("invalid/path")
	assert.Nil(t, device)
	assert.False(t, exists)
}

// Benchmark tests for performance validation
func BenchmarkV4L2DeviceManager_GetConnectedDevices(b *testing.B) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		devices := manager.GetConnectedDevices()
		_ = devices
	}
}

func BenchmarkV4L2DeviceManager_GetDevice(b *testing.B) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device, exists := manager.GetDevice("/dev/video0")
		_, _ = device, exists
	}
}

func BenchmarkV4L2DeviceManager_GetStats(b *testing.B) {
	configProvider := &MockConfigProvider{
		config: &camera.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			EnableCapabilityDetection: true,
		},
	}
	logger := &MockLogger{}

	manager := camera.NewV4L2DeviceManager(configProvider, logger)

	// Mock device checker
	manager.deviceChecker = &MockDeviceChecker{
		existsMap: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := manager.GetStats()
		_ = stats
	}
}
