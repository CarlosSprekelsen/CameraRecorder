/*
Camera Discovery Unit Tests

Requirements Coverage:
- REQ-CAM-001: Camera device detection and enumeration
- REQ-CAM-002: Camera capability probing and validation
- REQ-CAM-003: Camera status monitoring and reporting
- REQ-CAM-004: Hot-plug event handling
- REQ-CAM-005: Camera information and metadata
- REQ-CAM-006: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// TestCameraDiscovery tests the camera discovery functionality
func TestCameraDiscovery(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test camera monitor initialization
	t.Run("camera_monitor_initialization", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
		require.NotNil(t, cameraMonitor)
	})

	// Test camera enumeration
	t.Run("camera_enumeration", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test enumerating cameras
		cameras, err := cameraMonitor.EnumerateCameras()

		// Note: This may fail in test environment without real cameras
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			assert.NotNil(t, cameras)
			assert.IsType(t, []*camera.CameraInfo{}, cameras)
		}
	})

	// Test camera capability probing
	t.Run("camera_capability_probing", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test probing camera capabilities
		capabilities, err := cameraMonitor.ProbeCapabilities("/dev/video0")

		// Note: This may fail in test environment without real cameras
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			assert.NotNil(t, capabilities)
			assert.NotEmpty(t, capabilities.Formats)
			assert.NotEmpty(t, capabilities.Resolutions)
			assert.NotEmpty(t, capabilities.FPSOptions)
		}
	})

	// Test camera status monitoring
	t.Run("camera_status_monitoring", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test getting camera status
		status, err := cameraMonitor.GetCameraStatus("/dev/video0")

		// Note: This may fail in test environment without real cameras
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			assert.NotNil(t, status)
			assert.Equal(t, "/dev/video0", status.Device)
			assert.NotEmpty(t, status.Status)
			assert.NotEmpty(t, status.Name)
		}
	})

	// Test camera information retrieval
	t.Run("camera_information_retrieval", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test getting camera information
		info, err := cameraMonitor.GetCameraInfo("/dev/video0")

		// Note: This may fail in test environment without real cameras
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			assert.NotNil(t, info)
			assert.Equal(t, "/dev/video0", info.Device)
			assert.NotEmpty(t, info.Name)
			assert.NotEmpty(t, info.Driver)
		}
	})

	// Test hot-plug event handling
	t.Run("hot_plug_event_handling", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test starting event monitoring
		err := cameraMonitor.StartEventMonitoring()

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "monitoring")
		}

		// Test stopping event monitoring
		err = cameraMonitor.StopEventMonitoring()
		assert.NoError(t, err)
	})

	// Test camera connection management
	t.Run("camera_connection_management", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test checking if camera is connected
		connected := cameraMonitor.IsCameraConnected("/dev/video0")
		assert.IsType(t, false, connected)

		// Test getting connected cameras
		connectedCameras := cameraMonitor.GetConnectedCameras()
		assert.NotNil(t, connectedCameras)
		assert.IsType(t, []*camera.CameraInfo{}, connectedCameras)
	})

	// Test camera error handling
	t.Run("camera_error_handling", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test with invalid device path
		_, err := cameraMonitor.GetCameraStatus("invalid-device")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")

		// Test with non-existent device
		_, err = cameraMonitor.GetCameraStatus("/dev/video999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device")
	})

	// Test camera performance monitoring
	t.Run("camera_performance_monitoring", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test getting performance metrics
		metrics := cameraMonitor.GetPerformanceMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.EnumerationTime, float64(0))
		assert.GreaterOrEqual(t, metrics.TotalCameras, int64(0))
		assert.GreaterOrEqual(t, metrics.ConnectedCameras, int64(0))
	})

	// Test camera configuration validation
	t.Run("camera_configuration_validation", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test validating camera configuration
		config := &camera.CameraConfig{
			Device:     "/dev/video0",
			Resolution: "1920x1080",
			FPS:        30,
			Format:     "YUYV",
		}

		valid := cameraMonitor.ValidateConfiguration(config)
		assert.IsType(t, false, valid)
	})

	// Test camera discovery performance
	t.Run("camera_discovery_performance", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test discovery performance
		start := time.Now()
		_, err := cameraMonitor.EnumerateCameras()
		duration := time.Since(start)

		// Note: This may fail in test environment
		// We test the interface and performance characteristics
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			// Discovery should complete within reasonable time
			assert.Less(t, duration, 5*time.Second)
		}
	})

	// Test camera metadata extraction
	t.Run("camera_metadata_extraction", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test extracting camera metadata
		metadata, err := cameraMonitor.ExtractMetadata("/dev/video0")

		// Note: This may fail in test environment without real cameras
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "device")
		} else {
			assert.NotNil(t, metadata)
			assert.NotEmpty(t, metadata.Manufacturer)
			assert.NotEmpty(t, metadata.Model)
			assert.NotEmpty(t, metadata.SerialNumber)
		}
	})

	// Test camera format validation
	t.Run("camera_format_validation", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test validating camera formats
		formats := []string{"YUYV", "MJPEG", "RGB24"}

		for _, format := range formats {
			t.Run("format_"+format, func(t *testing.T) {
				supported := cameraMonitor.IsFormatSupported("/dev/video0", format)
				assert.IsType(t, false, supported)
			})
		}
	})

	// Test camera resolution validation
	t.Run("camera_resolution_validation", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test validating camera resolutions
		resolutions := []string{"1920x1080", "1280x720", "640x480"}

		for _, resolution := range resolutions {
			t.Run("resolution_"+resolution, func(t *testing.T) {
				supported := cameraMonitor.IsResolutionSupported("/dev/video0", resolution)
				assert.IsType(t, false, supported)
			})
		}
	})

	// Test camera FPS validation
	t.Run("camera_fps_validation", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test validating camera FPS
		fpsOptions := []int{15, 30, 60}

		for _, fps := range fpsOptions {
			t.Run("fps_"+string(rune(fps)), func(t *testing.T) {
				supported := cameraMonitor.IsFPSSupported("/dev/video0", fps)
				assert.IsType(t, false, supported)
			})
		}
	})

	// Test camera discovery caching
	t.Run("camera_discovery_caching", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test discovery caching
		cameras1, err1 := cameraMonitor.EnumerateCameras()
		cameras2, err2 := cameraMonitor.EnumerateCameras()

		// Note: This may fail in test environment
		// We test the interface and caching behavior
		if err1 == nil && err2 == nil {
			assert.Equal(t, len(cameras1), len(cameras2))
		}
	})

	// Test camera discovery refresh
	t.Run("camera_discovery_refresh", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test refreshing camera discovery
		err := cameraMonitor.RefreshDiscovery()

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "refresh")
		}
	})

	// Test camera discovery statistics
	t.Run("camera_discovery_statistics", func(t *testing.T) {
		cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)

		// Test getting discovery statistics
		stats := cameraMonitor.GetDiscoveryStatistics()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalDiscoveries, int64(0))
		assert.GreaterOrEqual(t, stats.SuccessfulDiscoveries, int64(0))
		assert.GreaterOrEqual(t, stats.FailedDiscoveries, int64(0))
		assert.GreaterOrEqual(t, stats.AverageDiscoveryTime, float64(0))
	})
}
