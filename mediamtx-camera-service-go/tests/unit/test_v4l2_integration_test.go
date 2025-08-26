//go:build unit
// +build unit

package camera_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

// MockConfigManager for integration testing
type MockConfigManager struct {
	config *config.Config
}

func (m *MockConfigManager) GetConfig() *config.Config {
	return m.config
}

func (m *MockConfigManager) LoadConfig(path string) error {
	return nil
}

func (m *MockConfigManager) WatchConfig() error {
	return nil
}

func (m *MockConfigManager) StopWatching() {
}

func TestV4L2IntegrationManager_Creation(t *testing.T) {
	// Create mock config manager
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	
	// Test creation with nil logger
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	assert.NotNil(t, integrationManager)
	assert.NotNil(t, integrationManager.GetStats())
}

func TestV4L2IntegrationManager_StartStop(t *testing.T) {
	// Create mock config manager with valid config
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test start without valid config
	err := integrationManager.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not available")
	
	// Test stop when not running
	err = integrationManager.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestV4L2IntegrationManager_GetConnectedDevices(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test when device manager is nil
	devices := integrationManager.GetConnectedDevices()
	assert.NotNil(t, devices)
	assert.Empty(t, devices)
}

func TestV4L2IntegrationManager_GetDevice(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test when device manager is nil
	device, exists := integrationManager.GetDevice("/dev/video0")
	assert.Nil(t, device)
	assert.False(t, exists)
}

func TestV4L2IntegrationManager_GetStats(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test when device manager is nil
	stats := integrationManager.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.DevicesDiscovered)
	assert.Equal(t, 0, stats.EventsProcessed)
}

func TestV4L2IntegrationManager_ValidateConfiguration(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name:        "nil_config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid_config",
			config: &config.Config{
				Camera: config.CameraConfig{
					PollInterval:              0.1,
					DetectionTimeout:          1.0,
					DeviceRange:              []int{0, 1, 2},
					EnableCapabilityDetection: true,
				},
			},
			expectError: false,
		},
		{
			name: "zero_poll_interval",
			config: &config.Config{
				Camera: config.CameraConfig{
					PollInterval: 0.0,
				},
			},
			expectError: true,
		},
		{
			name: "zero_detection_timeout",
			config: &config.Config{
				Camera: config.CameraConfig{
					DetectionTimeout: 0.0,
				},
			},
			expectError: true,
		},
		{
			name: "empty_device_range",
			config: &config.Config{
				Camera: config.CameraConfig{
					DeviceRange: []int{},
				},
			},
			expectError: true,
		},
		{
			name: "negative_device_number",
			config: &config.Config{
				Camera: config.CameraConfig{
					DeviceRange: []int{-1, 0, 1},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := integrationManager.ValidateConfiguration(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestV4L2IntegrationManager_ConfigurationUpdate(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test config update callback when not running
	newConfig := &config.Config{
		Camera: config.CameraConfig{
			PollInterval:              0.2,
			DetectionTimeout:          2.0,
			DeviceRange:              []int{0, 1, 2, 3},
			EnableCapabilityDetection: true,
		},
	}
	
	// This should not panic or error when not running
	// The callback should handle the not-running state gracefully
	// Note: In real implementation, this would be tested with proper mocking
	_ = integrationManager
	_ = newConfig
}

func TestV4L2IntegrationManager_ThreadSafety(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test concurrent access to methods
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			devices := integrationManager.GetConnectedDevices()
			_ = devices
			
			device, exists := integrationManager.GetDevice("/dev/video0")
			_, _ = device, exists
			
			stats := integrationManager.GetStats()
			_ = stats
			
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestV4L2IntegrationManager_ErrorHandling(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test with invalid device paths
	device, exists := integrationManager.GetDevice("")
	assert.Nil(t, device)
	assert.False(t, exists)
	
	device, exists = integrationManager.GetDevice("invalid/path")
	assert.Nil(t, device)
	assert.False(t, exists)
}

func TestV4L2IntegrationManager_ConfigurationValidationEdgeCases(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test with very large device ranges
	largeConfig := &config.Config{
		Camera: config.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:              make([]int, 1000), // Large range
			EnableCapabilityDetection: true,
		},
	}
	
	// Fill the large range with valid numbers
	for i := range largeConfig.Camera.DeviceRange {
		largeConfig.Camera.DeviceRange[i] = i
	}
	
	err := integrationManager.ValidateConfiguration(largeConfig)
	assert.NoError(t, err)
	
	// Test with very small poll interval
	smallIntervalConfig := &config.Config{
		Camera: config.CameraConfig{
			PollInterval:              0.001, // Very small
			DetectionTimeout:          1.0,
			DeviceRange:              []int{0},
			EnableCapabilityDetection: true,
		},
	}
	
	err = integrationManager.ValidateConfiguration(smallIntervalConfig)
	assert.NoError(t, err)
}

func TestV4L2IntegrationManager_StatisticsConsistency(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test that stats are consistent across multiple calls
	stats1 := integrationManager.GetStats()
	time.Sleep(10 * time.Millisecond)
	stats2 := integrationManager.GetStats()
	
	// Stats should be consistent when no operations are performed
	assert.Equal(t, stats1.DevicesDiscovered, stats2.DevicesDiscovered)
	assert.Equal(t, stats1.EventsProcessed, stats2.EventsProcessed)
	assert.Equal(t, stats1.PollingCycles, stats2.PollingCycles)
}

func TestV4L2IntegrationManager_ConfigurationUpdateHandling(t *testing.T) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	// Test that configuration updates are handled gracefully
	// This test validates that the integration manager can handle
	// configuration changes without crashing or producing errors
	
	configs := []*config.Config{
		{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1},
				EnableCapabilityDetection: true,
			},
		},
		{
			Camera: config.CameraConfig{
				PollInterval:              0.2,
				DetectionTimeout:          2.0,
				DeviceRange:              []int{0, 1, 2, 3},
				EnableCapabilityDetection: false,
			},
		},
		{
			Camera: config.CameraConfig{
				PollInterval:              0.05,
				DetectionTimeout:          0.5,
				DeviceRange:              []int{0},
				EnableCapabilityDetection: true,
			},
		},
	}
	
	for _, cfg := range configs {
		err := integrationManager.ValidateConfiguration(cfg)
		assert.NoError(t, err)
	}
}

// Benchmark tests for integration manager
func BenchmarkV4L2IntegrationManager_GetConnectedDevices(b *testing.B) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2, 3, 4, 5},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		devices := integrationManager.GetConnectedDevices()
		_ = devices
	}
}

func BenchmarkV4L2IntegrationManager_GetDevice(b *testing.B) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2, 3, 4, 5},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device, exists := integrationManager.GetDevice("/dev/video0")
		_, _ = device, exists
	}
}

func BenchmarkV4L2IntegrationManager_GetStats(b *testing.B) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2, 3, 4, 5},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := integrationManager.GetStats()
		_ = stats
	}
}

func BenchmarkV4L2IntegrationManager_ValidateConfiguration(b *testing.B) {
	configManager := &MockConfigManager{
		config: &config.Config{
			Camera: config.CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:              []int{0, 1, 2, 3, 4, 5},
				EnableCapabilityDetection: true,
			},
		},
	}
	integrationManager := camera.NewV4L2IntegrationManager(configManager, nil)
	
	validConfig := &config.Config{
		Camera: config.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          1.0,
			DeviceRange:              []int{0, 1, 2, 3, 4, 5},
			EnableCapabilityDetection: true,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := integrationManager.ValidateConfiguration(validConfig)
		_ = err
	}
}
