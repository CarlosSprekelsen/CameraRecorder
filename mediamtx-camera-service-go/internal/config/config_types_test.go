package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		DefaultPath:  "/opt/recordings",
		FallbackPath: "/tmp/recordings",
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 80, cfg.WarnPercent)
	assert.Equal(t, 90, cfg.BlockPercent)
	assert.Equal(t, "/opt/recordings", cfg.DefaultPath)
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
		RecordingsPath:                      "/opt/recordings",
		SnapshotsPath:                       "/opt/snapshots",
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
		StreamReadiness: StreamReadinessConfig{
			Timeout:                     10.0,
			RetryAttempts:               3,
			RetryDelay:                  2.0,
			CheckInterval:               5.0,
			EnableProgressNotifications: true,
			GracefulFallback:            true,
		},
		Codec: CodecConfig{
			VideoProfile: "main",
			VideoLevel:   "4.0",
			PixelFormat:  "yuv420p",
			Bitrate:      "1000000",
			Preset:       "medium",
		},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 9997, cfg.APIPort)
	assert.Equal(t, 8554, cfg.RTSPPort)
	assert.Equal(t, 8889, cfg.WebRTCPort)
	assert.Equal(t, 8888, cfg.HLSPort)
	assert.Equal(t, "/etc/mediamtx/mediamtx.yml", cfg.ConfigPath)
	assert.Equal(t, "/opt/recordings", cfg.RecordingsPath)
	assert.Equal(t, "/opt/snapshots", cfg.SnapshotsPath)
	assert.Equal(t, 30, cfg.HealthCheckInterval)
	assert.Equal(t, 3, cfg.HealthFailureThreshold)
	assert.Equal(t, 60, cfg.HealthCircuitBreakerTimeout)
	assert.Equal(t, 300, cfg.HealthMaxBackoffInterval)
	assert.Equal(t, 2, cfg.HealthRecoveryConfirmationThreshold)
	assert.Equal(t, 2.0, cfg.BackoffBaseMultiplier)
	assert.Equal(t, []float64{0.1, 0.5}, cfg.BackoffJitterRange)
	assert.Equal(t, 30.0, cfg.ProcessTerminationTimeout)
	assert.Equal(t, 10.0, cfg.ProcessKillTimeout)
	assert.Equal(t, 5*time.Second, cfg.HealthCheckTimeout)
	assert.Equal(t, 10.0, cfg.StreamReadiness.Timeout)
	assert.Equal(t, 3, cfg.StreamReadiness.RetryAttempts)
	assert.Equal(t, 2.0, cfg.StreamReadiness.RetryDelay)
	assert.Equal(t, 5.0, cfg.StreamReadiness.CheckInterval)
	assert.True(t, cfg.StreamReadiness.EnableProgressNotifications)
	assert.True(t, cfg.StreamReadiness.GracefulFallback)
	assert.Equal(t, "main", cfg.Codec.VideoProfile)
	assert.Equal(t, "4.0", cfg.Codec.VideoLevel)
	assert.Equal(t, "yuv420p", cfg.Codec.PixelFormat)
	assert.Equal(t, "1000000", cfg.Codec.Bitrate)
	assert.Equal(t, "medium", cfg.Codec.Preset)
}

func TestFFmpegConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: FFmpeg configuration must support snapshot and recording settings
	cfg := &FFmpegConfig{
		Snapshot: FFmpegSnapshotConfig{
			ProcessCreationTimeout: 10.0,
			ExecutionTimeout:       30.0,
			InternalTimeout:        5,
			RetryAttempts:          3,
			RetryDelay:             2.0,
		},
		Recording: FFmpegRecordingConfig{
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
		WebSocket: WebSocketNotificationConfig{
			DeliveryTimeout: 5.0,
			RetryAttempts:   3,
			RetryDelay:      1.0,
			MaxQueueSize:    1000,
			CleanupInterval: 60,
		},
		RealTime: RealTimeNotificationConfig{
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
		ResponseTimeTargets: ResponseTimeTargetsConfig{
			SnapshotCapture: 2.0,
			RecordingStart:  5.0,
			RecordingStop:   3.0,
			FileListing:     1.0,
		},
		SnapshotTiers: SnapshotTiersConfig{
			Tier1USBDirectTimeout:         1.0,
			Tier2RTSPReadyCheckTimeout:    3.0,
			Tier3ActivationTimeout:        5.0,
			Tier3ActivationTriggerTimeout: 2.0,
			TotalOperationTimeout:         10.0,
			ImmediateResponseThreshold:    0.5,
			AcceptableResponseThreshold:   2.0,
			SlowResponseThreshold:         5.0,
		},
		Optimization: OptimizationConfig{
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
		CapabilityTimeout:         10.0,
		CapabilityRetryInterval:   5.0,
		CapabilityMaxRetries:      3,
		EnableCapabilityDetection: true,
		AutoStartStreams:          true,
		DeviceRange:               []int{0, 1, 2},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, 5.0, cfg.PollInterval)
	assert.Equal(t, 30.0, cfg.DetectionTimeout)
	assert.Equal(t, 10.0, cfg.CapabilityTimeout)
	assert.Equal(t, 5.0, cfg.CapabilityRetryInterval)
	assert.Equal(t, 3, cfg.CapabilityMaxRetries)
	assert.True(t, cfg.EnableCapabilityDetection)
	assert.True(t, cfg.AutoStartStreams)
	assert.Equal(t, []int{0, 1, 2}, cfg.DeviceRange)
}

func TestLoggingConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Logging configuration must support file and console output
	cfg := &LoggingConfig{
		Level:          "info",
		Format:         "json",
		ConsoleEnabled: true,
		MaxFileSize:    10485760,
		FileEnabled:    true,
		FilePath:       "/var/log/camera-service.log",
		BackupCount:    5,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.True(t, cfg.ConsoleEnabled)
	assert.Equal(t, int64(10485760), cfg.MaxFileSize)
	assert.True(t, cfg.FileEnabled)
	assert.Equal(t, "/var/log/camera-service.log", cfg.FilePath)
	assert.Equal(t, 5, cfg.BackupCount)
}

func TestRecordingConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-003: Recording configuration must support format, quality, and cleanup settings
	cfg := &RecordingConfig{
		Enabled:              true,
		Format:               "mp4",
		Quality:              "high",
		SegmentDuration:      60,
		MaxSegmentSize:       104857600,
		AutoCleanup:          true,
		CleanupInterval:      3600,
		MaxAge:               86400,
		MaxSize:              1073741824,
		DefaultRotationSize:  104857600, // 100MB
		DefaultMaxDuration:   24 * time.Hour,
		DefaultRetentionDays: 7,
	}

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "mp4", cfg.Format)
	assert.Equal(t, "high", cfg.Quality)
	assert.Equal(t, 60, cfg.SegmentDuration)
	assert.Equal(t, int64(104857600), cfg.MaxSegmentSize)
	assert.True(t, cfg.AutoCleanup)
	assert.Equal(t, 3600, cfg.CleanupInterval)
	assert.Equal(t, 86400, cfg.MaxAge)
	assert.Equal(t, int64(1073741824), cfg.MaxSize)
	assert.Equal(t, int64(104857600), cfg.DefaultRotationSize)
	assert.Equal(t, 24*time.Hour, cfg.DefaultMaxDuration)
	assert.Equal(t, 7, cfg.DefaultRetentionDays)
}

func TestSnapshotConfig_Structure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-001: Snapshot configuration must support format, quality, and cleanup settings
	cfg := &SnapshotConfig{
		Enabled:         true,
		Format:          "jpeg",
		Quality:         90,
		MaxAge:          24 * 3600,
		MaxCount:        1000,
		MaxWidth:        1920,
		MaxHeight:       1080,
		AutoCleanup:     true,
		CleanupInterval: 3600,
	}

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "jpeg", cfg.Format)
	assert.Equal(t, 90, cfg.Quality)
	assert.Equal(t, 24*3600, cfg.MaxAge)
	assert.Equal(t, 1000, cfg.MaxCount)
	assert.Equal(t, 1920, cfg.MaxWidth)
	assert.Equal(t, 1080, cfg.MaxHeight)
	assert.True(t, cfg.AutoCleanup)
	assert.Equal(t, 3600, cfg.CleanupInterval)
}

func TestConfig_CompleteStructure(t *testing.T) {
	t.Parallel()
	// REQ-CONFIG-002: Complete configuration must support all sections
	cfg := &Config{
		Server: ServerConfig{
			Host:           "localhost",
			Port:           8080,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: MediaMTXConfig{
			Host:     "localhost",
			APIPort:  9997,
			RTSPPort: 8554,
		},
		Camera: CameraConfig{
			PollInterval: 5.0,
			DeviceRange:  []int{0, 1},
		},
		Logging: LoggingConfig{
			Level:          "info",
			ConsoleEnabled: true,
		},
		Recording: RecordingConfig{
			Enabled: true,
			Format:  "mp4",
		},
		Snapshots: SnapshotConfig{
			Enabled: true,
			Format:  "jpeg",
		},
		FFmpeg: FFmpegConfig{
			Snapshot: FFmpegSnapshotConfig{
				ProcessCreationTimeout: 10.0,
			},
			Recording: FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
			},
		},
		Notifications: NotificationsConfig{
			WebSocket: WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
			},
			RealTime: RealTimeNotificationConfig{
				CameraStatusInterval: 30.0,
			},
		},
		Performance: PerformanceConfig{
			ResponseTimeTargets: ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
			},
			SnapshotTiers: SnapshotTiersConfig{
				Tier1USBDirectTimeout: 1.0,
			},
			Optimization: OptimizationConfig{
				EnableCaching: true,
			},
		},
		Security: SecurityConfig{
			JWTSecretKey:   "test-secret",
			JWTExpiryHours: 24,
		},
		Storage: StorageConfig{
			WarnPercent:  80,
			BlockPercent: 90,
			DefaultPath:  "/opt/recordings",
			FallbackPath: "/tmp/recordings",
		},
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
}
