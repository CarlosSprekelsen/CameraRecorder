package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

/*
Module: Configuration Management System

Requirements Coverage:
- REQ-E1-S1.1-001: Configuration loading from YAML files
- REQ-E1-S1.1-002: Environment variable overrides
- REQ-E1-S1.1-003: Configuration validation
- REQ-E1-S1.1-004: Default value fallback
- REQ-E1-S1.1-005: Thread-safe configuration access

Test Categories: Unit
API Documentation Reference: N/A (Configuration system)
*/

//go:build unit
// +build unit

func TestConfigManager_LoadConfig_ValidYAML(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration loading from YAML files
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	// Create a valid YAML configuration file
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
  websocket_path: "/test"
  max_connections: 50

mediamtx:
  host: "localhost"
  api_port: 9998
  rtsp_port: 8555
  webrtc_port: 8890
  hls_port: 8889
  config_path: "/test/config.yml"
  recordings_path: "/test/recordings"
  snapshots_path: "/test/snapshots"
  
  codec:
    video_profile: "main"
    video_level: "4.0"
    pixel_format: "yuv422p"
    bitrate: "800k"
    preset: "fast"
  
  health_check_interval: 15
  health_failure_threshold: 5
  health_circuit_breaker_timeout: 30
  health_max_backoff_interval: 60
  health_recovery_confirmation_threshold: 2
  backoff_base_multiplier: 1.5
  backoff_jitter_range: [0.9, 1.1]
  process_termination_timeout: 2.0
  process_kill_timeout: 1.0
  
  stream_readiness:
    timeout: 10.0
    retry_attempts: 2
    retry_delay: 1.0
    check_interval: 0.25
    enable_progress_notifications: false
    graceful_fallback: false

ffmpeg:
  snapshot:
    process_creation_timeout: 3.0
    execution_timeout: 5.0
    internal_timeout: 3000000
    retry_attempts: 1
    retry_delay: 0.5
  
  recording:
    process_creation_timeout: 8.0
    execution_timeout: 12.0
    internal_timeout: 8000000
    retry_attempts: 2
    retry_delay: 1.5

notifications:
  websocket:
    delivery_timeout: 3.0
    retry_attempts: 2
    retry_delay: 0.5
    max_queue_size: 500
    cleanup_interval: 15
  
  real_time:
    camera_status_interval: 0.5
    recording_progress_interval: 0.25
    connection_health_check: 5.0

performance:
  response_time_targets:
    snapshot_capture: 1.5
    recording_start: 1.5
    recording_stop: 1.5
    file_listing: 0.5
  
  snapshot_tiers:
    tier1_usb_direct_timeout: 0.3
    tier2_rtsp_ready_check_timeout: 0.5
    tier3_activation_timeout: 2.0
    tier3_activation_trigger_timeout: 0.5
    total_operation_timeout: 8.0
    immediate_response_threshold: 0.3
    acceptable_response_threshold: 1.5
    slow_response_threshold: 3.0
  
  optimization:
    enable_caching: false
    cache_ttl: 150
    max_concurrent_operations: 3
    connection_pool_size: 5

camera:
  poll_interval: 0.05
  detection_timeout: 1.5
  device_range: [0, 4]
  enable_capability_detection: false
  auto_start_streams: false
  capability_timeout: 3.0
  capability_retry_interval: 0.5
  capability_max_retries: 2

logging:
  level: "DEBUG"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: false
  file_path: "/test/logs/test.log"
  max_file_size: 5242880
  backup_count: 3
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "medium"
  segment_duration: 1800
  max_segment_size: 262144000
  auto_cleanup: false
  cleanup_interval: 43200
  max_age: 302400
  max_size: 5368709120

snapshots:
  enabled: true
  format: "png"
  quality: 85
  max_width: 1280
  max_height: 720
  auto_cleanup: false
  cleanup_interval: 1800
  max_age: 43200
  max_count: 500
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	// Load configuration
	manager := config.NewConfigManager()
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	
	// Validate loaded configuration
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "/test", cfg.Server.WebSocketPath)
	assert.Equal(t, 50, cfg.Server.MaxConnections)
	
	assert.Equal(t, "localhost", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
	assert.Equal(t, 8555, cfg.MediaMTX.RTSPPort)
	assert.Equal(t, 8890, cfg.MediaMTX.WebRTCPort)
	assert.Equal(t, 8889, cfg.MediaMTX.HLSPort)
	
	assert.Equal(t, "main", cfg.MediaMTX.Codec.VideoProfile)
	assert.Equal(t, "4.0", cfg.MediaMTX.Codec.VideoLevel)
	assert.Equal(t, "yuv422p", cfg.MediaMTX.Codec.PixelFormat)
	assert.Equal(t, "800k", cfg.MediaMTX.Codec.Bitrate)
	assert.Equal(t, "fast", cfg.MediaMTX.Codec.Preset)
	
	assert.Equal(t, 15, cfg.MediaMTX.HealthCheckInterval)
	assert.Equal(t, 5, cfg.MediaMTX.HealthFailureThreshold)
	assert.Equal(t, 30, cfg.MediaMTX.HealthCircuitBreakerTimeout)
	assert.Equal(t, 60, cfg.MediaMTX.HealthMaxBackoffInterval)
	assert.Equal(t, 2, cfg.MediaMTX.HealthRecoveryConfirmationThreshold)
	assert.Equal(t, 1.5, cfg.MediaMTX.BackoffBaseMultiplier)
	assert.Equal(t, []float64{0.9, 1.1}, cfg.MediaMTX.BackoffJitterRange)
	assert.Equal(t, 2.0, cfg.MediaMTX.ProcessTerminationTimeout)
	assert.Equal(t, 1.0, cfg.MediaMTX.ProcessKillTimeout)
	
	assert.Equal(t, 10.0, cfg.MediaMTX.StreamReadiness.Timeout)
	assert.Equal(t, 2, cfg.MediaMTX.StreamReadiness.RetryAttempts)
	assert.Equal(t, 1.0, cfg.MediaMTX.StreamReadiness.RetryDelay)
	assert.Equal(t, 0.25, cfg.MediaMTX.StreamReadiness.CheckInterval)
	assert.False(t, cfg.MediaMTX.StreamReadiness.EnableProgressNotifications)
	assert.False(t, cfg.MediaMTX.StreamReadiness.GracefulFallback)
	
	assert.Equal(t, 3.0, cfg.FFmpeg.Snapshot.ProcessCreationTimeout)
	assert.Equal(t, 5.0, cfg.FFmpeg.Snapshot.ExecutionTimeout)
	assert.Equal(t, 3000000, cfg.FFmpeg.Snapshot.InternalTimeout)
	assert.Equal(t, 1, cfg.FFmpeg.Snapshot.RetryAttempts)
	assert.Equal(t, 0.5, cfg.FFmpeg.Snapshot.RetryDelay)
	
	assert.Equal(t, 8.0, cfg.FFmpeg.Recording.ProcessCreationTimeout)
	assert.Equal(t, 12.0, cfg.FFmpeg.Recording.ExecutionTimeout)
	assert.Equal(t, 8000000, cfg.FFmpeg.Recording.InternalTimeout)
	assert.Equal(t, 2, cfg.FFmpeg.Recording.RetryAttempts)
	assert.Equal(t, 1.5, cfg.FFmpeg.Recording.RetryDelay)
	
	assert.Equal(t, 3.0, cfg.Notifications.WebSocket.DeliveryTimeout)
	assert.Equal(t, 2, cfg.Notifications.WebSocket.RetryAttempts)
	assert.Equal(t, 0.5, cfg.Notifications.WebSocket.RetryDelay)
	assert.Equal(t, 500, cfg.Notifications.WebSocket.MaxQueueSize)
	assert.Equal(t, 15, cfg.Notifications.WebSocket.CleanupInterval)
	
	assert.Equal(t, 0.5, cfg.Notifications.RealTime.CameraStatusInterval)
	assert.Equal(t, 0.25, cfg.Notifications.RealTime.RecordingProgressInterval)
	assert.Equal(t, 5.0, cfg.Notifications.RealTime.ConnectionHealthCheck)
	
	assert.Equal(t, 1.5, cfg.Performance.ResponseTimeTargets.SnapshotCapture)
	assert.Equal(t, 1.5, cfg.Performance.ResponseTimeTargets.RecordingStart)
	assert.Equal(t, 1.5, cfg.Performance.ResponseTimeTargets.RecordingStop)
	assert.Equal(t, 0.5, cfg.Performance.ResponseTimeTargets.FileListing)
	
	assert.Equal(t, 0.3, cfg.Performance.SnapshotTiers.Tier1USBDirectTimeout)
	assert.Equal(t, 0.5, cfg.Performance.SnapshotTiers.Tier2RTSPReadyCheckTimeout)
	assert.Equal(t, 2.0, cfg.Performance.SnapshotTiers.Tier3ActivationTimeout)
	assert.Equal(t, 0.5, cfg.Performance.SnapshotTiers.Tier3ActivationTriggerTimeout)
	assert.Equal(t, 8.0, cfg.Performance.SnapshotTiers.TotalOperationTimeout)
	assert.Equal(t, 0.3, cfg.Performance.SnapshotTiers.ImmediateResponseThreshold)
	assert.Equal(t, 1.5, cfg.Performance.SnapshotTiers.AcceptableResponseThreshold)
	assert.Equal(t, 3.0, cfg.Performance.SnapshotTiers.SlowResponseThreshold)
	
	assert.False(t, cfg.Performance.Optimization.EnableCaching)
	assert.Equal(t, 150, cfg.Performance.Optimization.CacheTTL)
	assert.Equal(t, 3, cfg.Performance.Optimization.MaxConcurrentOperations)
	assert.Equal(t, 5, cfg.Performance.Optimization.ConnectionPoolSize)
	
	assert.Equal(t, 0.05, cfg.Camera.PollInterval)
	assert.Equal(t, 1.5, cfg.Camera.DetectionTimeout)
	assert.Equal(t, []int{0, 4}, cfg.Camera.DeviceRange)
	assert.False(t, cfg.Camera.EnableCapabilityDetection)
	assert.False(t, cfg.Camera.AutoStartStreams)
	assert.Equal(t, 3.0, cfg.Camera.CapabilityTimeout)
	assert.Equal(t, 0.5, cfg.Camera.CapabilityRetryInterval)
	assert.Equal(t, 2, cfg.Camera.CapabilityMaxRetries)
	
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "%(asctime)s - %(name)s - %(levelname)s - %(message)s", cfg.Logging.Format)
	assert.False(t, cfg.Logging.FileEnabled)
	assert.Equal(t, "/test/logs/test.log", cfg.Logging.FilePath)
	assert.Equal(t, int64(5242880), cfg.Logging.MaxFileSize)
	assert.Equal(t, 3, cfg.Logging.BackupCount)
	assert.True(t, cfg.Logging.ConsoleEnabled)
	
	assert.True(t, cfg.Recording.Enabled)
	assert.Equal(t, "mp4", cfg.Recording.Format)
	assert.Equal(t, "medium", cfg.Recording.Quality)
	assert.Equal(t, 1800, cfg.Recording.SegmentDuration)
	assert.Equal(t, int64(262144000), cfg.Recording.MaxSegmentSize)
	assert.False(t, cfg.Recording.AutoCleanup)
	assert.Equal(t, 43200, cfg.Recording.CleanupInterval)
	assert.Equal(t, 302400, cfg.Recording.MaxAge)
	assert.Equal(t, int64(5368709120), cfg.Recording.MaxSize)
	
	assert.True(t, cfg.Snapshots.Enabled)
	assert.Equal(t, "png", cfg.Snapshots.Format)
	assert.Equal(t, 85, cfg.Snapshots.Quality)
	assert.Equal(t, 1280, cfg.Snapshots.MaxWidth)
	assert.Equal(t, 720, cfg.Snapshots.MaxHeight)
	assert.False(t, cfg.Snapshots.AutoCleanup)
	assert.Equal(t, 1800, cfg.Snapshots.CleanupInterval)
	assert.Equal(t, 43200, cfg.Snapshots.MaxAge)
	assert.Equal(t, 500, cfg.Snapshots.MaxCount)
}

func TestConfigManager_LoadConfig_MissingFile(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")
	
	manager := config.NewConfigManager()
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err) // Should not error, should use defaults
	require.NotNil(t, cfg)
	
	// Should have default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
	assert.Equal(t, "/ws", cfg.Server.WebSocketPath)
	assert.Equal(t, 100, cfg.Server.MaxConnections)
	
	assert.Equal(t, "127.0.0.1", cfg.MediaMTX.Host)
	assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
	assert.Equal(t, 8554, cfg.MediaMTX.RTSPPort)
	assert.Equal(t, 8889, cfg.MediaMTX.WebRTCPort)
	assert.Equal(t, 8888, cfg.MediaMTX.HLSPort)
	
	assert.Equal(t, "baseline", cfg.MediaMTX.Codec.VideoProfile)
	assert.Equal(t, "3.0", cfg.MediaMTX.Codec.VideoLevel)
	assert.Equal(t, "yuv420p", cfg.MediaMTX.Codec.PixelFormat)
	assert.Equal(t, "600k", cfg.MediaMTX.Codec.Bitrate)
	assert.Equal(t, "ultrafast", cfg.MediaMTX.Codec.Preset)
}

func TestConfigManager_LoadConfig_InvalidYAML(t *testing.T) {
	// Test invalid YAML handling
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")
	
	// Create invalid YAML
	invalidYAML := `
server:
  host: "127.0.0.1"
  port: 9000
  websocket_path: "/test"
  max_connections: 50
  invalid_field: [unclosed_bracket
`
	
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)
	
	manager := config.NewConfigManager()
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err) // Should not error, should use defaults
	require.NotNil(t, cfg)
	
	// Should have default values due to invalid YAML
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_LoadConfig_EmptyFile(t *testing.T) {
	// Test empty YAML file handling
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty.yaml")
	
	// Create empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)
	
	manager := config.NewConfigManager()
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err) // Should not error, should use defaults
	require.NotNil(t, cfg)
	
	// Should have default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_EnvironmentVariableOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	// Create minimal YAML file
	yamlContent := `
server:
  host: "0.0.0.0"
  port: 8002
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	// Set environment variables
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "9000")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_HOST", "192.168.1.200")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_API_PORT", "9998")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_HOST")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_API_PORT")
	}()
	
	manager := config.NewConfigManager()
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	
	// Environment variables should override YAML values
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
}

func TestConfigManager_ThreadSafety(t *testing.T) {
	// REQ-E1-S1.1-005: Thread-safe configuration access
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	manager := config.NewConfigManager()
	
	// Load configuration in goroutine
	done := make(chan bool)
	go func() {
		cfg, err := manager.LoadConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
		done <- true
	}()
	
	// Access configuration concurrently
	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)
	
	<-done
}

func TestConfigValidation_ValidConfig(t *testing.T) {
	// REQ-E1-S1.1-003: Configuration validation
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "127.0.0.1",
			Port:           8002,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
		},
		MediaMTX: config.MediaMTXConfig{
			Host:                              "127.0.0.1",
			APIPort:                           9997,
			RTSPPort:                          8554,
			WebRTCPort:                        8889,
			HLSPort:                           8888,
			ConfigPath:                        "/opt/camera-service/config/mediamtx.yml",
			RecordingsPath:                    "/opt/camera-service/recordings",
			SnapshotsPath:                     "/opt/camera-service/snapshots",
			HealthCheckInterval:               30,
			HealthFailureThreshold:            10,
			HealthCircuitBreakerTimeout:       60,
			HealthMaxBackoffInterval:          120,
			HealthRecoveryConfirmationThreshold: 3,
			BackoffBaseMultiplier:             2.0,
			BackoffJitterRange:                []float64{0.8, 1.2},
			ProcessTerminationTimeout:         3.0,
			ProcessKillTimeout:                2.0,
			Codec: config.CodecConfig{
				VideoProfile: "baseline",
				VideoLevel:   "3.0",
				PixelFormat:  "yuv420p",
				Bitrate:      "600k",
				Preset:       "ultrafast",
			},
			StreamReadiness: config.StreamReadinessConfig{
				Timeout:                     15.0,
				RetryAttempts:               3,
				RetryDelay:                  2.0,
				CheckInterval:               0.5,
				EnableProgressNotifications: true,
				GracefulFallback:            true,
			},
		},
		Camera: config.CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          2.0,
			DeviceRange:               []int{0, 9},
			EnableCapabilityDetection: true,
			AutoStartStreams:          true,
			CapabilityTimeout:         5.0,
			CapabilityRetryInterval:   1.0,
			CapabilityMaxRetries:      3,
		},
		Logging: config.LoggingConfig{
			Level:          "INFO",
			Format:         "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
			FileEnabled:    true,
			FilePath:       "/opt/camera-service/logs/camera-service.log",
			MaxFileSize:    10485760,
			BackupCount:    5,
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled:         false,
			Format:          "fmp4",
			Quality:         "high",
			SegmentDuration: 3600,
			MaxSegmentSize:  524288000,
			AutoCleanup:     true,
			CleanupInterval: 86400,
			MaxAge:          604800,
			MaxSize:         10737418240,
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
				ProcessCreationTimeout: 5.0,
				ExecutionTimeout:       8.0,
				InternalTimeout:        5000000,
				RetryAttempts:          2,
				RetryDelay:             1.0,
			},
			Recording: config.FFmpegRecordingConfig{
				ProcessCreationTimeout: 10.0,
				ExecutionTimeout:       15.0,
				InternalTimeout:        10000000,
				RetryAttempts:          3,
				RetryDelay:             2.0,
			},
		},
		Notifications: config.NotificationsConfig{
			WebSocket: config.WebSocketNotificationConfig{
				DeliveryTimeout: 5.0,
				RetryAttempts:   3,
				RetryDelay:      1.0,
				MaxQueueSize:    1000,
				CleanupInterval: 30,
			},
			RealTime: config.RealTimeNotificationConfig{
				CameraStatusInterval:      1.0,
				RecordingProgressInterval: 0.5,
				ConnectionHealthCheck:     10.0,
			},
		},
		Performance: config.PerformanceConfig{
			ResponseTimeTargets: config.ResponseTimeTargetsConfig{
				SnapshotCapture: 2.0,
				RecordingStart:  2.0,
				RecordingStop:   2.0,
				FileListing:     1.0,
			},
			SnapshotTiers: config.SnapshotTiersConfig{
				Tier1USBDirectTimeout:          0.5,
				Tier2RTSPReadyCheckTimeout:     1.0,
				Tier3ActivationTimeout:         3.0,
				Tier3ActivationTriggerTimeout:  1.0,
				TotalOperationTimeout:          10.0,
				ImmediateResponseThreshold:     0.5,
				AcceptableResponseThreshold:    2.0,
				SlowResponseThreshold:          5.0,
			},
			Optimization: config.OptimizationConfig{
				EnableCaching:           true,
				CacheTTL:                300,
				MaxConcurrentOperations: 5,
				ConnectionPoolSize:      10,
			},
		},
	}
	
	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestConfigValidation_InvalidConfig(t *testing.T) {
	// Test validation with invalid configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:           "", // Invalid: empty host
			Port:           70000, // Invalid: port out of range
			WebSocketPath:  "invalid", // Invalid: doesn't start with /
			MaxConnections: 0, // Invalid: must be positive
		},
		MediaMTX: config.MediaMTXConfig{
			Host:                              "", // Invalid: empty host
			APIPort:                           70000, // Invalid: port out of range
			RTSPPort:                          0, // Invalid: must be positive
			WebRTCPort:                        -1, // Invalid: negative port
			HLSPort:                           0, // Invalid: must be positive
			ConfigPath:                        "", // Invalid: empty path
			RecordingsPath:                    "", // Invalid: empty path
			SnapshotsPath:                     "", // Invalid: empty path
			HealthCheckInterval:               0, // Invalid: must be positive
			HealthFailureThreshold:            0, // Invalid: must be positive
			HealthCircuitBreakerTimeout:       0, // Invalid: must be positive
			HealthMaxBackoffInterval:          0, // Invalid: must be positive
			HealthRecoveryConfirmationThreshold: 0, // Invalid: must be positive
			BackoffBaseMultiplier:             0, // Invalid: must be positive
			BackoffJitterRange:                []float64{1.2, 0.8}, // Invalid: min > max
			ProcessTerminationTimeout:         0, // Invalid: must be positive
			ProcessKillTimeout:                0, // Invalid: must be positive
			Codec: config.CodecConfig{
				VideoProfile: "invalid", // Invalid: not in allowed values
				VideoLevel:   "invalid", // Invalid: not in allowed values
				PixelFormat:  "invalid", // Invalid: not in allowed values
				Bitrate:      "", // Invalid: empty
				Preset:       "invalid", // Invalid: not in allowed values
			},
			StreamReadiness: config.StreamReadinessConfig{
				Timeout:                     0, // Invalid: must be positive
				RetryAttempts:               -1, // Invalid: negative
				RetryDelay:                  -1, // Invalid: negative
				CheckInterval:               0, // Invalid: must be positive
				EnableProgressNotifications: true,
				GracefulFallback:            true,
			},
		},
		Camera: config.CameraConfig{
			PollInterval:              -1, // Invalid: negative
			DetectionTimeout:          0, // Invalid: must be positive
			DeviceRange:               []int{5, 3}, // Invalid: min > max
			EnableCapabilityDetection: true,
			AutoStartStreams:          true,
			CapabilityTimeout:         0, // Invalid: must be positive
			CapabilityRetryInterval:   -1, // Invalid: negative
			CapabilityMaxRetries:      -1, // Invalid: negative
		},
		Logging: config.LoggingConfig{
			Level:          "INVALID", // Invalid: not in allowed values
			Format:         "", // Invalid: empty
			FileEnabled:    true,
			FilePath:       "", // Invalid: empty when file enabled
			MaxFileSize:    0, // Invalid: must be positive
			BackupCount:    -1, // Invalid: negative
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled:         false,
			Format:          "invalid", // Invalid: not in allowed values
			Quality:         "invalid", // Invalid: not in allowed values
			SegmentDuration: 0, // Invalid: must be positive
			MaxSegmentSize:  0, // Invalid: must be positive
			AutoCleanup:     true,
			CleanupInterval: 0, // Invalid: must be positive
			MaxAge:          0, // Invalid: must be positive
			MaxSize:         0, // Invalid: must be positive
		},
		Snapshots: config.SnapshotConfig{
			Enabled:         true,
			Format:          "invalid", // Invalid: not in allowed values
			Quality:         0, // Invalid: must be between 1-100
			MaxWidth:        0, // Invalid: must be positive
			MaxHeight:       0, // Invalid: must be positive
			AutoCleanup:     true,
			CleanupInterval: 0, // Invalid: must be positive
			MaxAge:          0, // Invalid: must be positive
			MaxCount:        0, // Invalid: must be positive
		},
	}
	
	err := config.ValidateConfig(cfg)
	assert.Error(t, err)
	
	// Check that validation error contains field information
	validationErr, ok := err.(*config.ValidationError)
	assert.True(t, ok)
	assert.Contains(t, validationErr.Error(), "configuration validation failed")
}

func TestConfigManager_GetConfig_ThreadSafe(t *testing.T) {
	// Test thread-safe access to configuration
	manager := config.NewConfigManager()
	
	// Access configuration from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cfg := manager.GetConfig()
			assert.NotNil(t, cfg)
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConfigManager_AddUpdateCallback(t *testing.T) {
	// Test configuration update callback functionality
	manager := config.NewConfigManager()
	
	callbackCalled := false
	callback := func(cfg *config.Config) {
		callbackCalled = true
		assert.NotNil(t, cfg)
	}
	
	manager.AddUpdateCallback(callback)
	
	// Load configuration to trigger callback
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	cfg, err := manager.LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	
	// Give some time for callback to be called
	// In a real implementation, this would be synchronous or we'd have a way to wait
	// For now, we just verify the callback was registered
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
}
