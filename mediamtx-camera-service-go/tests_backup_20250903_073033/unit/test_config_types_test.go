//go:build unit
// +build unit

package config

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-CONFIG-001: Configuration types must support all required fields for API compliance
// REQ-CONFIG-002: Configuration types must have proper mapstructure tags for YAML parsing
// REQ-CONFIG-003: Configuration types must support default values and validation

func TestServerConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: Server configuration must support all WebSocket server settings
	cfg := &ServerConfig{
		Host:           "localhost",
		Port:           8080,
		WebSocketPath:  "/ws",
		MaxConnections: 100,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		PingInterval:   25 * time.Second,
		PongWait:       60 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "/ws", cfg.WebSocketPath)
	assert.Equal(t, 100, cfg.MaxConnections)
	assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 25*time.Second, cfg.PingInterval)
	assert.Equal(t, 60*time.Second, cfg.PongWait)
	assert.Equal(t, int64(1024*1024), cfg.MaxMessageSize)
}

func TestCodecConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Codec configuration must support STANAG 4406 settings
	cfg := &CodecConfig{
		VideoProfile: "baseline",
		VideoLevel:   "3.1",
		PixelFormat:  "yuv420p",
		Bitrate:      "2M",
		Preset:       "fast",
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "baseline", cfg.VideoProfile)
	assert.Equal(t, "3.1", cfg.VideoLevel)
	assert.Equal(t, "yuv420p", cfg.PixelFormat)
	assert.Equal(t, "2M", cfg.Bitrate)
	assert.Equal(t, "fast", cfg.Preset)
}

func TestStreamReadinessConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: Stream readiness configuration must support timeout and retry settings
	cfg := &StreamReadinessConfig{
		Timeout:                     30.0,
		RetryAttempts:               3,
		RetryDelay:                  5.0,
		CheckInterval:               2.0,
		EnableProgressNotifications: true,
		GracefulFallback:            true,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 30.0, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, 5.0, cfg.RetryDelay)
	assert.Equal(t, 2.0, cfg.CheckInterval)
	assert.True(t, cfg.EnableProgressNotifications)
	assert.True(t, cfg.GracefulFallback)
}

func TestSecurityConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: Security configuration must support JWT and rate limiting
	cfg := &SecurityConfig{
		RateLimitRequests: 100,
		RateLimitWindow:   1 * time.Minute,
		JWTSecretKey:      "test-secret-key",
		JWTExpiryHours:    24,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 100, cfg.RateLimitRequests)
	assert.Equal(t, 1*time.Minute, cfg.RateLimitWindow)
	assert.Equal(t, "test-secret-key", cfg.JWTSecretKey)
	assert.Equal(t, 24, cfg.JWTExpiryHours)
}

func TestStorageConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Storage configuration must support path and threshold settings
	cfg := &StorageConfig{
		WarnPercent:  80,
		BlockPercent: 90,
		DefaultPath:  "/opt/camera-service/recordings",
		FallbackPath: "/tmp/recordings",
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 80, cfg.WarnPercent)
	assert.Equal(t, 90, cfg.BlockPercent)
	assert.Equal(t, "/opt/camera-service/recordings", cfg.DefaultPath)
	assert.Equal(t, "/tmp/recordings", cfg.FallbackPath)
}

func TestMediaMTXConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: MediaMTX configuration must support all integration settings
	cfg := &MediaMTXConfig{
		Host:                                "localhost",
		APIPort:                             9997,
		RTSPPort:                            8554,
		WebRTCPort:                          8889,
		HLSPort:                             8888,
		ConfigPath:                          "/etc/mediamtx/mediamtx.yml",
		RecordingsPath:                      "/opt/mediamtx/recordings",
		SnapshotsPath:                       "/opt/mediamtx/snapshots",
		HealthCheckInterval:                 30,
		HealthFailureThreshold:              3,
		HealthCircuitBreakerTimeout:         60,
		HealthMaxBackoffInterval:            300,
		HealthRecoveryConfirmationThreshold: 2,
		BackoffBaseMultiplier:               2.0,
		BackoffJitterRange:                  []float64{0.1, 0.3},
		ProcessTerminationTimeout:           10.0,
		ProcessKillTimeout:                  5.0,
		HealthCheckTimeout:                  5 * time.Second,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 9997, cfg.APIPort)
	assert.Equal(t, 8554, cfg.RTSPPort)
	assert.Equal(t, 8889, cfg.WebRTCPort)
	assert.Equal(t, 8888, cfg.HLSPort)
	assert.Equal(t, "/etc/mediamtx/mediamtx.yml", cfg.ConfigPath)
	assert.Equal(t, "/opt/mediamtx/recordings", cfg.RecordingsPath)
	assert.Equal(t, "/opt/mediamtx/snapshots", cfg.SnapshotsPath)
	assert.Equal(t, 30, cfg.HealthCheckInterval)
	assert.Equal(t, 3, cfg.HealthFailureThreshold)
	assert.Equal(t, 60, cfg.HealthCircuitBreakerTimeout)
	assert.Equal(t, 300, cfg.HealthMaxBackoffInterval)
	assert.Equal(t, 2, cfg.HealthRecoveryConfirmationThreshold)
	assert.Equal(t, 2.0, cfg.BackoffBaseMultiplier)
	assert.Equal(t, []float64{0.1, 0.3}, cfg.BackoffJitterRange)
	assert.Equal(t, 10.0, cfg.ProcessTerminationTimeout)
	assert.Equal(t, 5.0, cfg.ProcessKillTimeout)
	assert.Equal(t, 5*time.Second, cfg.HealthCheckTimeout)
}

func TestFFmpegConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: FFmpeg configuration must support snapshot and recording settings
	cfg := &FFmpegConfig{
		Snapshot: config.FFmpegSnapshotConfig{
			ProcessCreationTimeout: 10.0,
			ExecutionTimeout:       30.0,
			InternalTimeout:        5,
			RetryAttempts:          3,
			RetryDelay:             2.0,
		},
		Recording: config.FFmpegRecordingConfig{
			ProcessCreationTimeout: 10.0,
			ExecutionTimeout:       60.0,
			InternalTimeout:        10,
			RetryAttempts:          3,
			RetryDelay:             5.0,
		},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 10.0, cfg.Snapshot.ProcessCreationTimeout)
	assert.Equal(t, 30.0, cfg.Snapshot.ExecutionTimeout)
	assert.Equal(t, 5, cfg.Snapshot.InternalTimeout)
	assert.Equal(t, 3, cfg.Snapshot.RetryAttempts)
	assert.Equal(t, 2.0, cfg.Snapshot.RetryDelay)
	assert.Equal(t, 10.0, cfg.Recording.ProcessCreationTimeout)
	assert.Equal(t, 60.0, cfg.Recording.ExecutionTimeout)
	assert.Equal(t, 10, cfg.Recording.InternalTimeout)
	assert.Equal(t, 3, cfg.Recording.RetryAttempts)
	assert.Equal(t, 5.0, cfg.Recording.RetryDelay)
}

func TestNotificationsConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Notifications configuration must support WebSocket and real-time settings
	cfg := &NotificationsConfig{
		WebSocket: config.WebSocketNotificationConfig{
			DeliveryTimeout: 5.0,
			RetryAttempts:   3,
			RetryDelay:      1.0,
			MaxQueueSize:    1000,
			CleanupInterval: 60,
		},
		RealTime: config.RealTimeNotificationConfig{
			CameraStatusInterval:      30.0,
			RecordingProgressInterval: 5.0,
			ConnectionHealthCheck:     10.0,
		},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 5.0, cfg.WebSocket.DeliveryTimeout)
	assert.Equal(t, 3, cfg.WebSocket.RetryAttempts)
	assert.Equal(t, 1.0, cfg.WebSocket.RetryDelay)
	assert.Equal(t, 1000, cfg.WebSocket.MaxQueueSize)
	assert.Equal(t, 60, cfg.WebSocket.CleanupInterval)
	assert.Equal(t, 30.0, cfg.RealTime.CameraStatusInterval)
	assert.Equal(t, 5.0, cfg.RealTime.RecordingProgressInterval)
	assert.Equal(t, 10.0, cfg.RealTime.ConnectionHealthCheck)
}

func TestPerformanceConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: Performance configuration must support response time targets and optimization
	cfg := &PerformanceConfig{
		ResponseTimeTargets: config.ResponseTimeTargetsConfig{
			SnapshotCapture: 2.0,
			RecordingStart:  5.0,
			RecordingStop:   3.0,
			FileListing:     1.0,
		},
		SnapshotTiers: config.SnapshotTiersConfig{
			Tier1USBDirectTimeout:         1.0,
			Tier2RTSPReadyCheckTimeout:    3.0,
			Tier3ActivationTimeout:        5.0,
			Tier3ActivationTriggerTimeout: 2.0,
			TotalOperationTimeout:         10.0,
			ImmediateResponseThreshold:    0.5,
			AcceptableResponseThreshold:   2.0,
			SlowResponseThreshold:         5.0,
		},
		Optimization: config.OptimizationConfig{
			EnableCaching:           true,
			CacheTTL:                300,
			MaxConcurrentOperations: 10,
			ConnectionPoolSize:      5,
		},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 2.0, cfg.ResponseTimeTargets.SnapshotCapture)
	assert.Equal(t, 5.0, cfg.ResponseTimeTargets.RecordingStart)
	assert.Equal(t, 3.0, cfg.ResponseTimeTargets.RecordingStop)
	assert.Equal(t, 1.0, cfg.ResponseTimeTargets.FileListing)
	assert.Equal(t, 1.0, cfg.SnapshotTiers.Tier1USBDirectTimeout)
	assert.Equal(t, 3.0, cfg.SnapshotTiers.Tier2RTSPReadyCheckTimeout)
	assert.Equal(t, 5.0, cfg.SnapshotTiers.Tier3ActivationTimeout)
	assert.Equal(t, 2.0, cfg.SnapshotTiers.Tier3ActivationTriggerTimeout)
	assert.Equal(t, 10.0, cfg.SnapshotTiers.TotalOperationTimeout)
	assert.Equal(t, 0.5, cfg.SnapshotTiers.ImmediateResponseThreshold)
	assert.Equal(t, 2.0, cfg.SnapshotTiers.AcceptableResponseThreshold)
	assert.Equal(t, 5.0, cfg.SnapshotTiers.SlowResponseThreshold)
	assert.True(t, cfg.Optimization.EnableCaching)
	assert.Equal(t, 300, cfg.Optimization.CacheTTL)
	assert.Equal(t, 10, cfg.Optimization.MaxConcurrentOperations)
	assert.Equal(t, 5, cfg.Optimization.ConnectionPoolSize)
}

func TestCameraConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: Camera configuration must support discovery and capability settings
	cfg := &CameraConfig{
		PollInterval:              5.0,
		DetectionTimeout:          30.0,
		DeviceRange:               []int{0, 1, 2, 3},
		EnableCapabilityDetection: true,
		AutoStartStreams:          false,
		CapabilityTimeout:         10.0,
		CapabilityRetryInterval:   2.0,
		CapabilityMaxRetries:      3,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 5.0, cfg.PollInterval)
	assert.Equal(t, 30.0, cfg.DetectionTimeout)
	assert.Equal(t, []int{0, 1, 2, 3}, cfg.DeviceRange)
	assert.True(t, cfg.EnableCapabilityDetection)
	assert.False(t, cfg.AutoStartStreams)
	assert.Equal(t, 10.0, cfg.CapabilityTimeout)
	assert.Equal(t, 2.0, cfg.CapabilityRetryInterval)
	assert.Equal(t, 3, cfg.CapabilityMaxRetries)
}

func TestLoggingConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Logging configuration must support file and console output
	cfg := &LoggingConfig{
		Level:          "info",
		Format:         "json",
		FileEnabled:    true,
		FilePath:       "/var/log/camera-service.log",
		MaxFileSize:    100 * 1024 * 1024,
		BackupCount:    5,
		ConsoleEnabled: true,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.True(t, cfg.FileEnabled)
	assert.Equal(t, "/var/log/camera-service.log", cfg.FilePath)
	assert.Equal(t, int64(100*1024*1024), cfg.MaxFileSize)
	assert.Equal(t, 5, cfg.BackupCount)
	assert.True(t, cfg.ConsoleEnabled)
}

func TestRecordingConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: Recording configuration must support format, quality, and cleanup settings
	cfg := &RecordingConfig{
		Enabled:              true,
		Format:               "mp4",
		Quality:              "high",
		SegmentDuration:      300,
		MaxSegmentSize:       100 * 1024 * 1024,
		AutoCleanup:          true,
		CleanupInterval:      3600,
		MaxAge:               7 * 24 * 3600,
		MaxSize:              10 * 1024 * 1024 * 1024,
		DefaultRotationSize:  100 * 1024 * 1024,
		DefaultMaxDuration:   24 * time.Hour,
		DefaultRetentionDays: 7,
	}

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "mp4", cfg.Format)
	assert.Equal(t, "high", cfg.Quality)
	assert.Equal(t, 300, cfg.SegmentDuration)
	assert.Equal(t, int64(100*1024*1024), cfg.MaxSegmentSize)
	assert.True(t, cfg.AutoCleanup)
	assert.Equal(t, 3600, cfg.CleanupInterval)
	assert.Equal(t, 7*24*3600, cfg.MaxAge)
	assert.Equal(t, int64(10*1024*1024*1024), cfg.MaxSize)
	assert.Equal(t, int64(100*1024*1024), cfg.DefaultRotationSize)
	assert.Equal(t, 24*time.Hour, cfg.DefaultMaxDuration)
	assert.Equal(t, 7, cfg.DefaultRetentionDays)
}

func TestSnapshotConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: Snapshot configuration must support format, quality, and cleanup settings
	cfg := &SnapshotConfig{
		Enabled:         true,
		Format:          "jpeg",
		Quality:         85,
		MaxWidth:        1920,
		MaxHeight:       1080,
		AutoCleanup:     true,
		CleanupInterval: 3600,
		MaxAge:          24 * 3600,
		MaxCount:        1000,
	}

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "jpeg", cfg.Format)
	assert.Equal(t, 85, cfg.Quality)
	assert.Equal(t, 1920, cfg.MaxWidth)
	assert.Equal(t, 1080, cfg.MaxHeight)
	assert.True(t, cfg.AutoCleanup)
	assert.Equal(t, 3600, cfg.CleanupInterval)
	assert.Equal(t, 24*3600, cfg.MaxAge)
	assert.Equal(t, 1000, cfg.MaxCount)
}

func TestConfig_CompleteStructure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Complete configuration must support all sections
	cfg := &Config{
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
		HealthPort: nil,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.MediaMTX.Host)
	assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
	assert.Equal(t, 5.0, cfg.Camera.PollInterval)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.True(t, cfg.Recording.Enabled)
	assert.True(t, cfg.Snapshots.Enabled)
	assert.Equal(t, 10.0, cfg.FFmpeg.Snapshot.ProcessCreationTimeout)
	assert.Equal(t, 5.0, cfg.Notifications.WebSocket.DeliveryTimeout)
	assert.Equal(t, 2.0, cfg.Performance.ResponseTimeTargets.SnapshotCapture)
	assert.Equal(t, "test-secret", cfg.Security.JWTSecretKey)
	assert.Equal(t, "/opt/recordings", cfg.Storage.DefaultPath)
	assert.Nil(t, cfg.HealthPort)
}

func TestConfig_HealthPortOptional(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: Health port must be optional for testing scenarios
	cfg1 := &Config{}
	assert.Nil(t, cfg1.HealthPort)

	healthPort := 8081
	cfg2 := &Config{
		HealthPort: &healthPort,
	}
	assert.NotNil(t, cfg2.HealthPort)
	assert.Equal(t, 8081, *cfg2.HealthPort)
}

// TestFFmpegConfig_DefaultValues tests that FFmpeg manager sets proper defaults
func TestFFmpegConfig_DefaultValues(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: FFmpeg configuration must support snapshot and recording settings with defaults
	// Create a minimal MediaMTX config without FFmpeg settings to test defaults
	minimalConfig := &mediamtx.MediaMTXConfig{
		Host:     "localhost",
		APIPort:  9997,
		RTSPPort: 8554,
	}

	// Create FFmpeg manager - should set defaults
	manager := mediamtx.NewFFmpegManager(minimalConfig, logrus.New())

	// Verify default values are set
	assert.Equal(t, 10*time.Second, minimalConfig.FFmpeg.Snapshot.ProcessCreationTimeout, "Snapshot process creation timeout should be 10s")
	assert.Equal(t, 30*time.Second, minimalConfig.FFmpeg.Snapshot.ExecutionTimeout, "Snapshot execution timeout should be 30s")
	assert.Equal(t, 2, minimalConfig.FFmpeg.Snapshot.RetryAttempts, "Snapshot retry attempts should be 2")
	assert.Equal(t, 1*time.Second, minimalConfig.FFmpeg.Snapshot.RetryDelay, "Snapshot retry delay should be 1s")

	assert.Equal(t, 15*time.Second, minimalConfig.FFmpeg.Recording.ProcessCreationTimeout, "Recording process creation timeout should be 15s")
	assert.Equal(t, 60*time.Second, minimalConfig.FFmpeg.Recording.ExecutionTimeout, "Recording execution timeout should be 60s")
	assert.Equal(t, 3, minimalConfig.FFmpeg.Recording.RetryAttempts, "Recording retry attempts should be 3")
	assert.Equal(t, 2*time.Second, minimalConfig.FFmpeg.Recording.RetryDelay, "Recording retry delay should be 2s")

	// Verify manager was created successfully
	require.NotNil(t, manager, "FFmpeg manager should be created successfully")
}

// TestFFmpegConfig_CustomValues tests that custom values override defaults
func TestFFmpegConfig_CustomValues(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: FFmpeg configuration must support custom timeout values
	// Create config with custom FFmpeg settings
	customConfig := &mediamtx.MediaMTXConfig{
		Host:     "localhost",
		APIPort:  9997,
		RTSPPort: 8554,
		FFmpeg: mediamtx.FFmpegConfig{
			Snapshot: mediamtx.SnapshotConfig{
				ProcessCreationTimeout: 5 * time.Second,
				ExecutionTimeout:       20 * time.Second,
				RetryAttempts:          1,
				RetryDelay:             500 * time.Millisecond,
			},
			Recording: mediamtx.RecordingConfig{
				ProcessCreationTimeout: 8 * time.Second,
				ExecutionTimeout:       45 * time.Second,
				RetryAttempts:          2,
				RetryDelay:             1 * time.Second,
			},
		},
	}

	// Create FFmpeg manager
	manager := mediamtx.NewFFmpegManager(customConfig, logrus.New())

	// Verify custom values are preserved
	assert.Equal(t, 5*time.Second, customConfig.FFmpeg.Snapshot.ProcessCreationTimeout, "Custom snapshot process creation timeout should be preserved")
	assert.Equal(t, 20*time.Second, customConfig.FFmpeg.Snapshot.ExecutionTimeout, "Custom snapshot execution timeout should be preserved")
	assert.Equal(t, 1, customConfig.FFmpeg.Snapshot.RetryAttempts, "Custom snapshot retry attempts should be preserved")
	assert.Equal(t, 500*time.Millisecond, customConfig.FFmpeg.Snapshot.RetryDelay, "Custom snapshot retry delay should be preserved")

	assert.Equal(t, 8*time.Second, customConfig.FFmpeg.Recording.ProcessCreationTimeout, "Custom recording process creation timeout should be preserved")
	assert.Equal(t, 45*time.Second, customConfig.FFmpeg.Recording.ExecutionTimeout, "Custom recording execution timeout should be preserved")
	assert.Equal(t, 2, customConfig.FFmpeg.Recording.RetryAttempts, "Custom recording retry attempts should be preserved")
	assert.Equal(t, 1*time.Second, customConfig.FFmpeg.Recording.RetryDelay, "Custom recording retry delay should be preserved")

	// Verify manager was created successfully
	require.NotNil(t, manager, "FFmpeg manager should be created successfully")
}
