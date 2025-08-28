//go:build unit
// +build unit

package config_test

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/assert"
)

// REQ-CONFIG-VALID-001: Configuration validation must catch all invalid configurations
// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages
// REQ-CONFIG-VALID-003: Validation must enforce API compliance requirements

func TestValidationError_Error(t *testing.T) {
	// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages
	err := &config.ValidationError{
		Field:   "server.port",
		Message: "port must be between 1 and 65535",
	}

	expected := "validation error for field 'server.port': port must be between 1 and 65535"
	assert.Equal(t, expected, err.Error())
}

func TestValidateConfig_ValidConfiguration(t *testing.T) {
	// REQ-CONFIG-VALID-001: Valid configuration must pass validation
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           8080,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "localhost",
			APIPort:  9997,
			RTSPPort: 8554,
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_InvalidServerConfig(t *testing.T) {
	// REQ-CONFIG-VALID-001: Invalid server configuration must fail validation
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "", // Invalid: empty host
			Port:           0,  // Invalid: port 0
			WebSocketPath:  "", // Invalid: empty path
			MaxConnections: 0,  // Invalid: 0 connections
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "localhost",
			APIPort:  9997,
			RTSPPort: 8554,
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server.host")
	assert.Contains(t, err.Error(), "server.port")
	assert.Contains(t, err.Error(), "server.websocket_path")
	assert.Contains(t, err.Error(), "server.max_connections")
}

func TestValidateConfig_InvalidMediaMTXConfig(t *testing.T) {
	// REQ-CONFIG-VALID-001: Invalid MediaMTX configuration must fail validation
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           8080,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "", // Invalid: empty host
			APIPort:  0,  // Invalid: port 0
			RTSPPort: 0,  // Invalid: port 0
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mediamtx.host")
	assert.Contains(t, err.Error(), "mediamtx.api_port")
	assert.Contains(t, err.Error(), "mediamtx.rtsp_port")
}

func TestValidateConfig_InvalidWebSocketPath(t *testing.T) {
	// REQ-CONFIG-VALID-003: WebSocket path must start with '/'
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           8080,
			WebSocketPath:  "ws", // Invalid: doesn't start with '/'
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "localhost",
			APIPort:  9997,
			RTSPPort: 8554,
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "websocket path must start with '/'")
}

func TestValidateConfig_InvalidPortRange(t *testing.T) {
	// REQ-CONFIG-VALID-001: Port must be in valid range 1-65535
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           70000, // Invalid: port > 65535
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "localhost",
			APIPort:  9997,
			RTSPPort: 8554,
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port must be between 1 and 65535")
}

func TestValidateConfig_MultipleValidationErrors(t *testing.T) {
	// REQ-CONFIG-VALID-002: Multiple validation errors must be reported
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "", // Invalid: empty host
			Port:           0,  // Invalid: port 0
			WebSocketPath:  "", // Invalid: empty path
			MaxConnections: 0,  // Invalid: 0 connections
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "", // Invalid: empty host
			APIPort:  0,  // Invalid: port 0
			RTSPPort: 0,  // Invalid: port 0
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.Error(t, err)

	// Check that all validation errors are included
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "server.host")
	assert.Contains(t, errorMsg, "server.port")
	assert.Contains(t, errorMsg, "server.websocket_path")
	assert.Contains(t, errorMsg, "server.max_connections")
	assert.Contains(t, errorMsg, "mediamtx.host")
	assert.Contains(t, errorMsg, "mediamtx.api_port")
	assert.Contains(t, errorMsg, "mediamtx.rtsp_port")
}

func TestValidateConfig_BoundaryValues(t *testing.T) {
	// REQ-CONFIG-VALID-001: Boundary values must be handled correctly
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           1, // Valid: minimum port
			WebSocketPath:  "/ws",
			MaxConnections: 1, // Valid: minimum connections
		},
		MediaMTX: config.MediaMTXConfig{
			Host:     "localhost",
			APIPort:  65535, // Valid: maximum port
			RTSPPort: 65535, // Valid: maximum port
		},
		Camera: config.CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: config.SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: config.SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: config.StorageConfig{
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)
}
