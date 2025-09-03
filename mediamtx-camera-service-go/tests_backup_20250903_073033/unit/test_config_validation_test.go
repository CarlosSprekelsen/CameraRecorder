//go:build unit
// +build unit

package config

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/assert"
)

// REQ-CONFIG-VALID-001: Configuration validation must catch all invalid configurations
// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages
// REQ-CONFIG-VALID-003: Validation must enforce API compliance requirements

func TestValidationError_Error(t *testing.T) {
	// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages
	err := &ValidationError{
		Field:   "server.port",
		Message: "port must be between 1 and 65535",
	}

	expected := "validation error for field 'server.port': port must be between 1 and 65535"
	assert.Equal(t, expected, err.Error())
}

func TestValidateConfig_ValidConfiguration(t *testing.T) {
	// REQ-CONFIG-VALID-001: Valid configuration must pass validation
	cfg := &Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           8080,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:                                "localhost",
			APIPort:                             9997,
			RTSPPort:                            8554,
			WebRTCPort:                          8889,
			HLSPort:                             8888,
			ConfigPath:                          "/tmp/config.yml",
			RecordingsPath:                      "/tmp/recordings",
			SnapshotsPath:                       "/tmp/snapshots",
			HealthCheckInterval:                 30,
			HealthFailureThreshold:              3,
			HealthCircuitBreakerTimeout:         60,
			HealthMaxBackoffInterval:            300,
			HealthRecoveryConfirmationThreshold: 2,
			BackoffBaseMultiplier:               2.0,
			BackoffJitterRange:                  []float64{0.1, 0.5},
			ProcessTerminationTimeout:           30.0,
			ProcessKillTimeout:                  10.0,
			HealthCheckTimeout:                  5 * time.Second,
			StreamReadiness: config.StreamReadinessConfig{
				Timeout:                     10.0,
				RetryAttempts:               3,
				RetryDelay:                  2.0,
				CheckInterval:               5.0,
				EnableProgressNotifications: true,
				GracefulFallback:            true,
			},
			Codec: config.CodecConfig{
				VideoProfile: "main",
				VideoLevel:   "4.0",
				PixelFormat:  "yuv420p",
				Bitrate:      "1000000",
				Preset:       "medium",
			},
		},
		Camera: config.CameraConfig{
			PollInterval:              5.0,
			DetectionTimeout:          1.0,
			CapabilityTimeout:         3.0,
			CapabilityRetryInterval:   2.0,
			CapabilityMaxRetries:      3,
			EnableCapabilityDetection: true,
			AutoStartStreams:          true,
			DeviceRange:               []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			Format:         "json",
			ConsoleEnabled: true,
			MaxFileSize:    10485760,
			FileEnabled:    true,
			FilePath:       "/var/log/camera-service.log",
			BackupCount:    5,
		},
		Recording: config.RecordingConfig{
			Enabled:              true,
			Format:               "mp4",
			Quality:              "high",
			SegmentDuration:      60,
			MaxSegmentSize:       104857600,
			AutoCleanup:          true,
			CleanupInterval:      3600,
			MaxAge:               86400,
			MaxSize:              1073741824,
			DefaultRotationSize:  104857600,
			DefaultMaxDuration:   24 * time.Hour,
			DefaultRetentionDays: 7,
		},
		Snapshots: config.SnapshotConfig{
			Enabled:         true,
			Format:          "jpeg",
			Quality:         90,
			MaxWidth:        1920,
			MaxHeight:       1080,
			AutoCleanup:     true,
			CleanupInterval: 3600,
			MaxAge:          86400,
			MaxCount:        1000,
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
				ExecutionTimeout:       5.0,
				InternalTimeout:        1000000,
				RetryAttempts:          3,
				RetryDelay:             2.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
				ExecutionTimeout:       10.0,
				InternalTimeout:        1000000, // Valid: positive value
				RetryAttempts:          3,
				RetryDelay:             2.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
				RetryAttempts:   3,
				RetryDelay:      2.0,
				MaxQueueSize:    100,
				CleanupInterval: 300,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval:      30.0,
				RecordingProgressInterval: 1.0,
				ConnectionHealthCheck:     5.0, // Valid: positive value
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
				RecordingStart:  1.5,
				RecordingStop:   1.5,
				FileListing:     0.5,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout:         1.0,
				Tier2RTSPReadyCheckTimeout:    2.0,
				Tier3ActivationTimeout:        3.0,
				Tier3ActivationTriggerTimeout: 1.5,
				TotalOperationTimeout:         10.0,
				ImmediateResponseThreshold:    0.5,
				AcceptableResponseThreshold:   1.0,
				SlowResponseThreshold:         2.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching:           true,
				CacheTTL:                3600,
				MaxConcurrentOperations: 10,
				ConnectionPoolSize:      20,
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
		RetentionPolicy: config.RetentionPolicyConfig{
			Enabled:     true,
			Type:        "age",
			MaxAgeDays:  7,
			MaxSizeGB:   1,
			AutoCleanup: true,
		},
	}

	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_InvalidServerConfig(t *testing.T) {
	// REQ-CONFIG-VALID-001: Invalid server configuration must fail validation
	cfg := &Config{
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
	assert.Contains(t, err.Error(), "mediamtx.webrtc_port")
	assert.Contains(t, err.Error(), "camera.detection_timeout")
	assert.Contains(t, err.Error(), "logging.format")
}

func TestValidateConfig_InvalidMediaMTXConfig(t *testing.T) {
	// REQ-CONFIG-VALID-001: Invalid MediaMTX configuration must fail validation
	cfg := &Config{
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
	assert.Contains(t, err.Error(), "camera.detection_timeout")
	assert.Contains(t, err.Error(), "logging.format")
}

func TestValidateConfig_InvalidWebSocketPath(t *testing.T) {
	// REQ-CONFIG-VALID-003: WebSocket path must start with '/'
	cfg := &Config{
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
	cfg := &Config{
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
	cfg := &Config{
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
	assert.Contains(t, errorMsg, "mediamtx.host")
	assert.Contains(t, errorMsg, "camera.detection_timeout")
	assert.Contains(t, errorMsg, "logging.format")
	assert.Contains(t, errorMsg, "recording.quality")
	assert.Contains(t, errorMsg, "snapshots.quality")
	assert.Contains(t, errorMsg, "ffmpeg.snapshot.execution_timeout")
	assert.Contains(t, errorMsg, "notifications.websocket.max_queue_size")
	assert.Contains(t, errorMsg, "performance.response_time_targets.recording_start")
	assert.Contains(t, errorMsg, "retention_policy.type")
}

func TestValidateConfig_BoundaryValues(t *testing.T) {
	// REQ-CONFIG-VALID-001: Boundary values must be handled correctly
	cfg := &Config{
		Server: config.ServerConfig{
			Host:           "localhost",
			Port:           1, // Valid: minimum port
			WebSocketPath:  "/ws",
			MaxConnections: 1, // Valid: minimum connections
		},
		MediaMTX: config.MediaMTXConfig{
			Host:                                "localhost",
			APIPort:                             65535, // Valid: maximum port
			RTSPPort:                            65535, // Valid: maximum port
			WebRTCPort:                          65535, // Valid: maximum port
			HLSPort:                             65535, // Valid: maximum port
			ConfigPath:                          "/tmp/config.yml",
			RecordingsPath:                      "/tmp/recordings",
			SnapshotsPath:                       "/tmp/snapshots",
			HealthCheckInterval:                 30,                  // Valid: positive value
			HealthFailureThreshold:              3,                   // Valid: positive value
			HealthCircuitBreakerTimeout:         60,                  // Valid: positive value
			HealthMaxBackoffInterval:            300,                 // Valid: positive value
			HealthRecoveryConfirmationThreshold: 2,                   // Valid: positive value
			BackoffBaseMultiplier:               2.0,                 // Valid: positive value
			BackoffJitterRange:                  []float64{0.1, 0.5}, // Valid: positive values
			ProcessTerminationTimeout:           30.0,                // Valid: positive value
			ProcessKillTimeout:                  10.0,                // Valid: positive value
			HealthCheckTimeout:                  5 * time.Second,     // Valid: positive value
			StreamReadiness: config.StreamReadinessConfig{
				Timeout:                     10.0, // Valid: positive value
				RetryAttempts:               3,    // Valid: positive value
				RetryDelay:                  2.0,  // Valid: positive value
				CheckInterval:               5.0,  // Valid: positive value
				EnableProgressNotifications: true,
				GracefulFallback:            true,
			},
			Codec: config.CodecConfig{
				VideoProfile: "main",    // Valid profile
				VideoLevel:   "4.0",     // Valid level
				PixelFormat:  "yuv420p", // Valid format
				Bitrate:      "1000000", // Valid bitrate
				Preset:       "medium",  // Valid preset
			},
		},
		Camera: config.CameraConfig{
			PollInterval:      5.0,
			DetectionTimeout:  1.0, // Valid: positive value
			CapabilityTimeout: 3.0, // Valid: positive value
			DeviceRange:       []int{0, 1},
		},
		Logging: config.LoggingConfig{
			Level:          "info",
			Format:         "json", // Valid format
			ConsoleEnabled: true,
			MaxFileSize:    10485760, // Valid: positive value
		},
		Recording: config.RecordingConfig{
			Enabled:         true,
			Format:          "mp4",
			Quality:         "high",     // Valid quality
			SegmentDuration: 60,         // Valid: positive value
			MaxSegmentSize:  104857600,  // Valid: positive value
			CleanupInterval: 3600,       // Valid: positive value
			MaxAge:          86400,      // Valid: positive value
			MaxSize:         1073741824, // Valid: positive value
		},
		Snapshots: config.SnapshotConfig{
			Enabled:         true,
			Format:          "jpeg",
			Quality:         90,    // Valid quality (1-100)
			MaxWidth:        1920,  // Valid: positive value
			MaxHeight:       1080,  // Valid: positive value
			CleanupInterval: 3600,  // Valid: positive value
			MaxAge:          86400, // Valid: positive value
			MaxCount:        1000,  // Valid: positive value
		},
		FFmpeg: config.FFmpegConfig{
			Snapshot: config.FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
				ExecutionTimeout:       5.0,     // Valid: positive value
				InternalTimeout:        1000000, // Valid: positive value
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
				ExecutionTimeout:       10.0,    // Valid: positive value
				InternalTimeout:        1000000, // Valid: positive value
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
				MaxQueueSize:    100, // Valid: positive value
				CleanupInterval: 300, // Valid: positive value
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval:      30.0,
				RecordingProgressInterval: 1.0, // Valid: positive value
				ConnectionHealthCheck:     5.0, // Valid: positive value
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
				RecordingStart:  1.5, // Valid: positive value
				RecordingStop:   1.5, // Valid: positive value
				FileListing:     0.5, // Valid: positive value
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout:         1.0,
				Tier2RTSPReadyCheckTimeout:    2.0,  // Valid: positive value
				Tier3ActivationTimeout:        3.0,  // Valid: positive value
				Tier3ActivationTriggerTimeout: 1.5,  // Valid: positive value
				TotalOperationTimeout:         10.0, // Valid: positive value
				ImmediateResponseThreshold:    0.5,  // Valid: positive value
				AcceptableResponseThreshold:   1.0,  // Valid: positive value
				SlowResponseThreshold:         2.0,  // Valid: positive value
			},
			Optimization: config.OptimizationConfig{
				EnableCaching:           true,
				CacheTTL:                3600,
				MaxConcurrentOperations: 10,
				ConnectionPoolSize:      20,
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
		RetentionPolicy: config.RetentionPolicyConfig{
			Type: "age", // Valid policy type
		},
	}

	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)
}
