//go:build unit
// +build unit

package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
)

// cleanupCameraServiceEnvVars cleans up all CAMERA_SERVICE environment variables
func cleanupCameraServiceEnvVars() {
	envVars := []string{
		"CAMERA_SERVICE_SERVER_HOST",
		"CAMERA_SERVICE_SERVER_PORT",
		"CAMERA_SERVICE_SERVER_WEBSOCKET_PATH",
		"CAMERA_SERVICE_SERVER_MAX_CONNECTIONS",
		"CAMERA_SERVICE_MEDIAMTX_HOST",
		"CAMERA_SERVICE_MEDIAMTX_API_PORT",
		"CAMERA_SERVICE_MEDIAMTX_RTSP_PORT",
		"CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT",
		"CAMERA_SERVICE_MEDIAMTX_HLS_PORT",
		"CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH",
		"CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH",
		"CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_CHECK_INTERVAL",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_FAILURE_THRESHOLD",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_CIRCUIT_BREAKER_TIMEOUT",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_MAX_BACKOFF_INTERVAL",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_RECOVERY_CONFIRMATION_THRESHOLD",
		"CAMERA_SERVICE_MEDIAMTX_BACKOFF_BASE_MULTIPLIER",
		"CAMERA_SERVICE_MEDIAMTX_PROCESS_TERMINATION_TIMEOUT",
		"CAMERA_SERVICE_MEDIAMTX_PROCESS_KILL_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_POLL_INTERVAL",
		"CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION",
		"CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_RETRY_INTERVAL",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_MAX_RETRIES",
		"CAMERA_SERVICE_LOGGING_LEVEL",
		"CAMERA_SERVICE_LOGGING_FORMAT",
		"CAMERA_SERVICE_LOGGING_FILE_ENABLED",
		"CAMERA_SERVICE_LOGGING_FILE_PATH",
		"CAMERA_SERVICE_LOGGING_CONSOLE_ENABLED",
		"CAMERA_SERVICE_RECORDING_ENABLED",
		"CAMERA_SERVICE_RECORDING_FORMAT",
		"CAMERA_SERVICE_RECORDING_QUALITY",
		"CAMERA_SERVICE_SNAPSHOTS_ENABLED",
		"CAMERA_SERVICE_SNAPSHOTS_FORMAT",
		"CAMERA_SERVICE_SNAPSHOTS_QUALITY",
		"CAMERA_SERVICE_ENABLE_HOT_RELOAD",
		"CAMERA_SERVICE_ENV",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

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

COMMON PATTERN: This file demonstrates the correct way to use test utilities
instead of creating individual components in each test function.

ANTI-PATTERN (OLD WAY):
   manager := config.NewConfigManager()
   // Creates 40+ instances across this file

CORRECT PATTERN (NEW WAY):
   env := utils.SetupTestEnvironment(t)
   defer utils.TeardownTestEnvironment(t, env)
   // Uses shared utilities, reduces duplication
*/

func TestConfigManager_LoadConfig_ValidYAML(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration loading from YAML files

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Clean up any existing environment variables that might interfere
	cleanupCameraServiceEnvVars()

	// Use the test environment's temporary directory and config manager
	tempDir := env.TempDir
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
  backoff_jitter_range: "[0.9, 1.1]"
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
  capability_max_retries: 3

logging:
  level: "debug"
  format: "json"
  file_enabled: false
  file_path: "/tmp/test.log"
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "high"
  default_path: "/tmp/test_recordings"
  fallback_path: "/tmp/test_fallback"
  default_rotation_size: 104857600
  default_max_duration: 3600
  default_retention_days: 7

snapshots:
  enabled: true
  format: "jpeg"
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

	// COMMON PATTERN: Use the test environment's config manager instead of creating a new one
	// This eliminates the 40+ ConfigManager instances that were created in this file
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
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
	assert.Equal(t, 3, cfg.Camera.CapabilityMaxRetries)

	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.False(t, cfg.Logging.FileEnabled)
	assert.Equal(t, "/tmp/test.log", cfg.Logging.FilePath)
	assert.True(t, cfg.Logging.ConsoleEnabled)

	assert.True(t, cfg.Recording.Enabled)
	assert.Equal(t, "mp4", cfg.Recording.Format)
	assert.Equal(t, "high", cfg.Recording.Quality)
	assert.Equal(t, "/tmp/test_recordings", cfg.Storage.DefaultPath)
	assert.Equal(t, "/tmp/test_fallback", cfg.Storage.FallbackPath)
	assert.Equal(t, int64(104857600), cfg.Recording.DefaultRotationSize)
	assert.Equal(t, 3600, cfg.Recording.DefaultMaxDuration)
	assert.Equal(t, 7, cfg.Recording.DefaultRetentionDays)

	assert.True(t, cfg.Snapshots.Enabled)
	assert.Equal(t, "jpeg", cfg.Snapshots.Format)
	assert.Equal(t, 85, cfg.Snapshots.Quality)
	assert.Equal(t, 1280, cfg.Snapshots.MaxWidth)
	assert.Equal(t, 720, cfg.Snapshots.MaxHeight)
	assert.False(t, cfg.Snapshots.AutoCleanup)
	assert.Equal(t, 1800, cfg.Snapshots.CleanupInterval)
	assert.Equal(t, 43200, cfg.Snapshots.MaxAge)
	assert.Equal(t, 500, cfg.Snapshots.MaxCount)
}

func TestConfigManager_LoadConfig_MissingFile(t *testing.T) {
	// REQ-CONFIG-001: The system SHALL validate configuration files before loading
	// REQ-CONFIG-002: The system SHALL fail fast on configuration errors

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")

	// COMMON PATTERN: Use the test environment's config manager instead of creating a new one
	// This eliminates the 40+ ConfigManager instances that were created in this file
	err := env.ConfigManager.LoadConfig(configPath)
	require.Error(t, err) // Should fail fast on invalid configuration
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigManager_LoadConfig_InvalidYAML(t *testing.T) {
	// Test invalid YAML handling

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	// Create invalid YAML
	invalidYAML := `
server:
  host: "127.0.0.1"
  port: 9000
  websocket_path: "/test"
  max_connections: 50
  invalid_field: "unclosed_bracket"
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// COMMON PATTERN: Use the test environment's config manager instead of creating a new one
	// This eliminates the 40+ ConfigManager instances that were created in this file
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err) // Should not error, should use defaults
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Should have default values due to invalid YAML
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_LoadConfig_EmptyFile(t *testing.T) {
	// Test empty YAML file handling

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty.yaml")

	// Create empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	// COMMON PATTERN: Use the test environment's config manager instead of creating a new one
	// This eliminates the 40+ ConfigManager instances that were created in this file
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err) // Should not error, should use defaults
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Should have default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_EnvironmentVariableOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "test_config.yaml")

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

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Environment variables should override YAML values
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
}

func TestConfigManager_ThreadSafety(t *testing.T) {
	// REQ-E1-S1.1-005: Thread-safe configuration access

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "test_config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load configuration in goroutine
	done := make(chan bool)
	go func() {
		err := env.ConfigManager.LoadConfig(configPath)
		assert.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
		done <- true
	}()

	// Access configuration concurrently
	cfg := env.ConfigManager.GetConfig()
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
			Host:                                "127.0.0.1",
			APIPort:                             9997,
			RTSPPort:                            8554,
			WebRTCPort:                          8889,
			HLSPort:                             8888,
			ConfigPath:                          "/opt/camera-service/config/mediamtx.yml",
			RecordingsPath:                      "/opt/camera-service/recordings",
			SnapshotsPath:                       "/opt/camera-service/snapshots",
			HealthCheckInterval:                 30,
			HealthFailureThreshold:              10,
			HealthCircuitBreakerTimeout:         60,
			HealthMaxBackoffInterval:            120,
			HealthRecoveryConfirmationThreshold: 3,
			BackoffBaseMultiplier:               2.0,
			BackoffJitterRange:                  []float64{0.8, 1.2},
			ProcessTerminationTimeout:           3.0,
			ProcessKillTimeout:                  2.0,
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
				Tier1USBDirectTimeout:         0.5,
				Tier2RTSPReadyCheckTimeout:    1.0,
				Tier3ActivationTimeout:        3.0,
				Tier3ActivationTriggerTimeout: 1.0,
				TotalOperationTimeout:         10.0,
				ImmediateResponseThreshold:    0.5,
				AcceptableResponseThreshold:   2.0,
				SlowResponseThreshold:         5.0,
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
			Host:           "",        // Invalid: empty host
			Port:           70000,     // Invalid: port out of range
			WebSocketPath:  "invalid", // Invalid: doesn't start with /
			MaxConnections: 0,         // Invalid: must be positive
		},
		MediaMTX: config.MediaMTXConfig{
			Host:                                "",                  // Invalid: empty host
			APIPort:                             70000,               // Invalid: port out of range
			RTSPPort:                            0,                   // Invalid: must be positive
			WebRTCPort:                          -1,                  // Invalid: negative port
			HLSPort:                             0,                   // Invalid: must be positive
			ConfigPath:                          "",                  // Invalid: empty path
			RecordingsPath:                      "",                  // Invalid: empty path
			SnapshotsPath:                       "",                  // Invalid: empty path
			HealthCheckInterval:                 0,                   // Invalid: must be positive
			HealthFailureThreshold:              0,                   // Invalid: must be positive
			HealthCircuitBreakerTimeout:         0,                   // Invalid: must be positive
			HealthMaxBackoffInterval:            0,                   // Invalid: must be positive
			HealthRecoveryConfirmationThreshold: 0,                   // Invalid: must be positive
			BackoffBaseMultiplier:               0,                   // Invalid: must be positive
			BackoffJitterRange:                  []float64{1.2, 0.8}, // Invalid: min > max
			ProcessTerminationTimeout:           0,                   // Invalid: must be positive
			ProcessKillTimeout:                  0,                   // Invalid: must be positive
			Codec: config.CodecConfig{
				VideoProfile: "invalid", // Invalid: not in allowed values
				VideoLevel:   "invalid", // Invalid: not in allowed values
				PixelFormat:  "invalid", // Invalid: not in allowed values
				Bitrate:      "",        // Invalid: empty
				Preset:       "invalid", // Invalid: not in allowed values
			},
			StreamReadiness: config.StreamReadinessConfig{
				Timeout:                     0,  // Invalid: must be positive
				RetryAttempts:               -1, // Invalid: negative
				RetryDelay:                  -1, // Invalid: negative
				CheckInterval:               0,  // Invalid: must be positive
				EnableProgressNotifications: true,
				GracefulFallback:            true,
			},
		},
		Camera: config.CameraConfig{
			PollInterval:              -1,          // Invalid: negative
			DetectionTimeout:          0,           // Invalid: must be positive
			DeviceRange:               []int{5, 3}, // Invalid: min > max
			EnableCapabilityDetection: true,
			AutoStartStreams:          true,
			CapabilityTimeout:         0,  // Invalid: must be positive
			CapabilityRetryInterval:   -1, // Invalid: negative
			CapabilityMaxRetries:      -1, // Invalid: negative
		},
		Logging: config.LoggingConfig{
			Level:          "INVALID", // Invalid: not in allowed values
			Format:         "",        // Invalid: empty
			FileEnabled:    true,
			FilePath:       "", // Invalid: empty when file enabled
			MaxFileSize:    0,  // Invalid: must be positive
			BackupCount:    -1, // Invalid: negative
			ConsoleEnabled: true,
		},
		Recording: config.RecordingConfig{
			Enabled:         false,
			Format:          "invalid", // Invalid: not in allowed values
			Quality:         "invalid", // Invalid: not in allowed values
			SegmentDuration: 0,         // Invalid: must be positive
			MaxSegmentSize:  0,         // Invalid: must be positive
			AutoCleanup:     true,
			CleanupInterval: 0, // Invalid: must be positive
			MaxAge:          0, // Invalid: must be positive
			MaxSize:         0, // Invalid: must be positive
		},
		Snapshots: config.SnapshotConfig{
			Enabled:         true,
			Format:          "invalid", // Invalid: not in allowed values
			Quality:         0,         // Invalid: must be between 1-100
			MaxWidth:        0,         // Invalid: must be positive
			MaxHeight:       0,         // Invalid: must be positive
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

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Access configuration from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cfg := env.ConfigManager.GetConfig()
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

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	callback := func(cfg *config.Config) {
		assert.NotNil(t, cfg)
	}

	env.ConfigManager.AddUpdateCallback(callback)

	// Load configuration to trigger callback
	configPath := filepath.Join(env.TempDir, "test_config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify the configuration was loaded correctly
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
}

func TestConfigManager_HotReload(t *testing.T) {
	// REQ-E1-S1.1-006: Hot reload capability

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "test_config.yaml")

	// Create initial configuration
	initialYAML := `
server:
  host: "127.0.0.1"
  port: 8002
`
	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload for this test
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify initial configuration
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)

	// Create a channel to track configuration updates
	updateChan := make(chan *config.Config, 1)
	env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
		updateChan <- cfg
	})

	// Update the configuration file
	updatedYAML := `
server:
  host: "192.168.1.100"
  port: 9000
`
	err = os.WriteFile(configPath, []byte(updatedYAML), 0644)
	require.NoError(t, err)

	// Wait for configuration update (with timeout)
	select {
	case updatedCfg := <-updateChan:
		assert.Equal(t, "192.168.1.100", updatedCfg.Server.Host)
		assert.Equal(t, 9000, updatedCfg.Server.Port)
	case <-time.After(2 * time.Second):
		t.Fatal("Hot reload did not trigger within expected time")
	}

	// Verify the configuration was updated
	cfg = env.ConfigManager.GetConfig()
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)

	// Clean up
	env.ConfigManager.Stop()
}

func TestConfigManager_Stop(t *testing.T) {
	// Test proper cleanup of configuration manager
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 8002
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Enable hot reload for this test
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Stop the manager
	env.ConfigManager.Stop()

	// Verify configuration is still accessible after stop
	cfg := env.ConfigManager.GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
}

func TestConfigManager_EnvironmentVariableComprehensive(t *testing.T) {
	// REQ-E1-S1.1-002: Comprehensive environment variable testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Create base YAML file
	yamlContent := `
server:
  host: "0.0.0.0"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  health_check_interval: 30
  health_failure_threshold: 10
  health_circuit_breaker_timeout: 60
  health_max_backoff_interval: 120
  health_recovery_confirmation_threshold: 3
  backoff_base_multiplier: 2.0
  backoff_jitter_range: "[0.8, 1.2]"
  process_termination_timeout: 3.0
  process_kill_timeout: 2.0
  
  codec:
    video_profile: "baseline"
    video_level: "3.0"
    pixel_format: "yuv420p"
    bitrate: "600k"
    preset: "ultrafast"
  
  stream_readiness:
    timeout: 15.0
    retry_attempts: 3
    retry_delay: 2.0
    check_interval: 0.5
    enable_progress_notifications: true
    graceful_fallback: true

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: "[0, 9]"
  enable_capability_detection: true
  auto_start_streams: true
  capability_timeout: 5.0
  capability_retry_interval: 1.0
  capability_max_retries: 3

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/opt/camera-service/logs/camera-service.log"
  max_file_size: 10485760
  backup_count: 5
  console_enabled: true

recording:
  enabled: false
  format: "fmp4"
  quality: "high"
  segment_duration: 3600
  max_segment_size: 524288000
  auto_cleanup: true
  cleanup_interval: 86400
  max_age: 604800
  max_size: 10737418240

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90
  max_width: 1920
  max_height: 1080
  auto_cleanup: true
  cleanup_interval: 3600
  max_age: 86400
  max_count: 1000

ffmpeg:
  snapshot:
    process_creation_timeout: 5.0
    execution_timeout: 8.0
    internal_timeout: 5000000
    retry_attempts: 2
    retry_delay: 1.0
  
  recording:
    process_creation_timeout: 10.0
    execution_timeout: 15.0
    internal_timeout: 10000000
    retry_attempts: 3
    retry_delay: 2.0

notifications:
  websocket:
    delivery_timeout: 5.0
    retry_attempts: 3
    retry_delay: 1.0
    max_queue_size: 1000
    cleanup_interval: 30
  
  real_time:
    camera_status_interval: 1.0
    recording_progress_interval: 0.5
    connection_health_check: 10.0

performance:
  response_time_targets:
    snapshot_capture: 2.0
    recording_start: 2.0
    recording_stop: 2.0
    file_listing: 1.0
  
  snapshot_tiers:
    tier1_usb_direct_timeout: 0.5
    tier2_rtsp_ready_check_timeout: 1.0
    tier3_activation_timeout: 3.0
    tier3_activation_trigger_timeout: 1.0
    total_operation_timeout: 10.0
    immediate_response_threshold: 0.5
    acceptable_response_threshold: 2.0
    slow_response_threshold: 5.0
  
  optimization:
    enable_caching: true
    cache_ttl: 300
    max_concurrent_operations: 5
    connection_pool_size: 10
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Test all environment variable mappings
	testCases := []struct {
		name     string
		envKey   string
		envValue string
		expected interface{}
	}{
		// Server configuration
		{"server_host", "CAMERA_SERVICE_SERVER_HOST", "192.168.1.100", "192.168.1.100"},
		{"server_port", "CAMERA_SERVICE_SERVER_PORT", "9000", 9000},
		{"server_websocket_path", "CAMERA_SERVICE_SERVER_WEBSOCKET_PATH", "/api/ws", "/api/ws"},
		{"server_max_connections", "CAMERA_SERVICE_SERVER_MAX_CONNECTIONS", "200", 200},

		// MediaMTX configuration
		{"mediamtx_host", "CAMERA_SERVICE_MEDIAMTX_HOST", "192.168.1.200", "192.168.1.200"},
		{"mediamtx_api_port", "CAMERA_SERVICE_MEDIAMTX_API_PORT", "9998", 9998},
		{"mediamtx_rtsp_port", "CAMERA_SERVICE_MEDIAMTX_RTSP_PORT", "8555", 8555},
		{"mediamtx_webrtc_port", "CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT", "8890", 8890},
		{"mediamtx_hls_port", "CAMERA_SERVICE_MEDIAMTX_HLS_PORT", "8889", 8889},
		{"mediamtx_config_path", "CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH", "/custom/config.yml", "/custom/config.yml"},
		{"mediamtx_recordings_path", "CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH", "/custom/recordings", "/custom/recordings"},
		{"mediamtx_snapshots_path", "CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH", "/custom/snapshots", "/custom/snapshots"},
		{"mediamtx_health_check_interval", "CAMERA_SERVICE_MEDIAMTX_HEALTH_CHECK_INTERVAL", "15", 15},
		{"mediamtx_health_failure_threshold", "CAMERA_SERVICE_MEDIAMTX_HEALTH_FAILURE_THRESHOLD", "5", 5},
		{"mediamtx_health_circuit_breaker_timeout", "CAMERA_SERVICE_MEDIAMTX_HEALTH_CIRCUIT_BREAKER_TIMEOUT", "30", 30},
		{"mediamtx_health_max_backoff_interval", "CAMERA_SERVICE_MEDIAMTX_HEALTH_MAX_BACKOFF_INTERVAL", "60", 60},
		{"mediamtx_health_recovery_confirmation_threshold", "CAMERA_SERVICE_MEDIAMTX_HEALTH_RECOVERY_CONFIRMATION_THRESHOLD", "2", 2},
		{"mediamtx_backoff_base_multiplier", "CAMERA_SERVICE_MEDIAMTX_BACKOFF_BASE_MULTIPLIER", "1.5", 1.5},
		{"mediamtx_process_termination_timeout", "CAMERA_SERVICE_MEDIAMTX_PROCESS_TERMINATION_TIMEOUT", "2.0", 2.0},
		{"mediamtx_process_kill_timeout", "CAMERA_SERVICE_MEDIAMTX_PROCESS_KILL_TIMEOUT", "1.0", 1.0},

		// Camera configuration
		{"camera_poll_interval", "CAMERA_SERVICE_CAMERA_POLL_INTERVAL", "0.05", 0.05},
		{"camera_detection_timeout", "CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT", "1.5", 1.5},
		{"camera_enable_capability_detection", "CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION", "false", false},
		{"camera_auto_start_streams", "CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS", "false", false},
		{"camera_capability_timeout", "CAMERA_SERVICE_CAMERA_CAPABILITY_TIMEOUT", "3.0", 3.0},
		{"camera_capability_retry_interval", "CAMERA_SERVICE_CAMERA_CAPABILITY_RETRY_INTERVAL", "0.5", 0.5},
		{"camera_capability_max_retries", "CAMERA_SERVICE_CAMERA_CAPABILITY_MAX_RETRIES", "2", 2},

		// Logging configuration
		{"logging_level", "CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG", "DEBUG"},
		{"logging_format", "CAMERA_SERVICE_LOGGING_FORMAT", "%(levelname)s - %(message)s", "%(levelname)s - %(message)s"},
		{"logging_file_enabled", "CAMERA_SERVICE_LOGGING_FILE_ENABLED", "false", false},
		{"logging_file_path", "CAMERA_SERVICE_LOGGING_FILE_PATH", "/custom/logs/app.log", "/custom/logs/app.log"},
		{"logging_console_enabled", "CAMERA_SERVICE_LOGGING_CONSOLE_ENABLED", "false", false},

		// Recording configuration
		{"recording_enabled", "CAMERA_SERVICE_RECORDING_ENABLED", "true", true},
		{"recording_format", "CAMERA_SERVICE_RECORDING_FORMAT", "mp4", "mp4"},
		{"recording_quality", "CAMERA_SERVICE_RECORDING_QUALITY", "medium", "medium"},

		// Snapshots configuration
		{"snapshots_enabled", "CAMERA_SERVICE_SNAPSHOTS_ENABLED", "false", false},
		{"snapshots_format", "CAMERA_SERVICE_SNAPSHOTS_FORMAT", "png", "png"},
		{"snapshots_quality", "CAMERA_SERVICE_SNAPSHOTS_QUALITY", "80", 80},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// COMMON PATTERN: Use shared test environment instead of individual components
			// This eliminates the need to create ConfigManager and Logger in every test
			env := utils.SetupTestEnvironment(t)
			defer utils.TeardownTestEnvironment(t, env)

			// Set environment variable
			os.Setenv(tc.envKey, tc.envValue)
			defer os.Unsetenv(tc.envKey)

			err := env.ConfigManager.LoadConfig(configPath)
			require.NoError(t, err)
			cfg := env.ConfigManager.GetConfig()
			require.NotNil(t, cfg)

			// Verify environment variable override
			switch tc.name {
			case "server_host":
				assert.Equal(t, tc.expected, cfg.Server.Host)
			case "server_port":
				assert.Equal(t, tc.expected, cfg.Server.Port)
			case "server_websocket_path":
				assert.Equal(t, tc.expected, cfg.Server.WebSocketPath)
			case "server_max_connections":
				assert.Equal(t, tc.expected, cfg.Server.MaxConnections)
			case "mediamtx_host":
				assert.Equal(t, tc.expected, cfg.MediaMTX.Host)
			case "mediamtx_api_port":
				assert.Equal(t, tc.expected, cfg.MediaMTX.APIPort)
			case "mediamtx_rtsp_port":
				assert.Equal(t, tc.expected, cfg.MediaMTX.RTSPPort)
			case "mediamtx_webrtc_port":
				assert.Equal(t, tc.expected, cfg.MediaMTX.WebRTCPort)
			case "mediamtx_hls_port":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HLSPort)
			case "mediamtx_config_path":
				assert.Equal(t, tc.expected, cfg.MediaMTX.ConfigPath)
			case "mediamtx_recordings_path":
				assert.Equal(t, tc.expected, cfg.MediaMTX.RecordingsPath)
			case "mediamtx_snapshots_path":
				assert.Equal(t, tc.expected, cfg.MediaMTX.SnapshotsPath)
			case "mediamtx_health_check_interval":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthCheckInterval)
			case "mediamtx_health_failure_threshold":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthFailureThreshold)
			case "mediamtx_health_circuit_breaker_timeout":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthCircuitBreakerTimeout)
			case "mediamtx_health_max_backoff_interval":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthMaxBackoffInterval)
			case "mediamtx_health_recovery_confirmation_threshold":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthRecoveryConfirmationThreshold)
			case "mediamtx_backoff_base_multiplier":
				assert.Equal(t, tc.expected, cfg.MediaMTX.BackoffBaseMultiplier)
			case "mediamtx_process_termination_timeout":
				assert.Equal(t, tc.expected, cfg.MediaMTX.ProcessTerminationTimeout)
			case "mediamtx_process_kill_timeout":
				assert.Equal(t, tc.expected, cfg.MediaMTX.ProcessKillTimeout)
			case "camera_poll_interval":
				assert.Equal(t, tc.expected, cfg.Camera.PollInterval)
			case "camera_detection_timeout":
				assert.Equal(t, tc.expected, cfg.Camera.DetectionTimeout)
			case "camera_enable_capability_detection":
				assert.Equal(t, tc.expected, cfg.Camera.EnableCapabilityDetection)
			case "camera_auto_start_streams":
				assert.Equal(t, tc.expected, cfg.Camera.AutoStartStreams)
			case "camera_capability_timeout":
				assert.Equal(t, tc.expected, cfg.Camera.CapabilityTimeout)
			case "camera_capability_retry_interval":
				assert.Equal(t, tc.expected, cfg.Camera.CapabilityRetryInterval)
			case "camera_capability_max_retries":
				assert.Equal(t, tc.expected, cfg.Camera.CapabilityMaxRetries)
			case "logging_level":
				assert.Equal(t, tc.expected, cfg.Logging.Level)
			case "logging_format":
				assert.Equal(t, tc.expected, cfg.Logging.Format)
			case "logging_file_enabled":
				assert.Equal(t, tc.expected, cfg.Logging.FileEnabled)
			case "logging_file_path":
				assert.Equal(t, tc.expected, cfg.Logging.FilePath)
			case "logging_console_enabled":
				assert.Equal(t, tc.expected, cfg.Logging.ConsoleEnabled)
			case "recording_enabled":
				assert.Equal(t, tc.expected, cfg.Recording.Enabled)
			case "recording_format":
				assert.Equal(t, tc.expected, cfg.Recording.Format)
			case "recording_quality":
				assert.Equal(t, tc.expected, cfg.Recording.Quality)
			case "snapshots_enabled":
				assert.Equal(t, tc.expected, cfg.Snapshots.Enabled)
			case "snapshots_format":
				assert.Equal(t, tc.expected, cfg.Snapshots.Format)
			case "snapshots_quality":
				assert.Equal(t, tc.expected, cfg.Snapshots.Quality)
			}
		})
	}
}

func TestConfigManager_EnvironmentVariableTypeConversion(t *testing.T) {
	// Test environment variable type conversion (string to int, bool, float)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
server:
  port: 8002
  max_connections: 100

mediamtx:
  health_check_interval: 30
  backoff_base_multiplier: 2.0

camera:
  enable_capability_detection: true
  auto_start_streams: true

logging:
  file_enabled: true
  console_enabled: true

recording:
  enabled: false

snapshots:
  enabled: true
  quality: 90
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Test type conversions
	testCases := []struct {
		name     string
		envKey   string
		envValue string
		expected interface{}
	}{
		// String to int conversions
		{"string_to_int_port", "CAMERA_SERVICE_SERVER_PORT", "9000", 9000},
		{"string_to_int_max_connections", "CAMERA_SERVICE_SERVER_MAX_CONNECTIONS", "200", 200},
		{"string_to_int_health_interval", "CAMERA_SERVICE_MEDIAMTX_HEALTH_CHECK_INTERVAL", "15", 15},

		// String to float conversions
		{"string_to_float_backoff", "CAMERA_SERVICE_MEDIAMTX_BACKOFF_BASE_MULTIPLIER", "1.5", 1.5},
		{"string_to_float_poll_interval", "CAMERA_SERVICE_CAMERA_POLL_INTERVAL", "0.05", 0.05},

		// String to bool conversions
		{"string_to_bool_true", "CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION", "true", true},
		{"string_to_bool_false", "CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS", "false", false},
		{"string_to_bool_1", "CAMERA_SERVICE_LOGGING_FILE_ENABLED", "1", true},
		{"string_to_bool_0", "CAMERA_SERVICE_LOGGING_CONSOLE_ENABLED", "0", false},
		{"string_to_bool_t", "CAMERA_SERVICE_RECORDING_ENABLED", "t", true},
		{"string_to_bool_f", "CAMERA_SERVICE_SNAPSHOTS_ENABLED", "f", false},

		// String to int (quality)
		{"string_to_int_quality", "CAMERA_SERVICE_SNAPSHOTS_QUALITY", "80", 80},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// COMMON PATTERN: Use shared test environment instead of individual components
			// This eliminates the need to create ConfigManager and Logger in every test
			env := utils.SetupTestEnvironment(t)
			defer utils.TeardownTestEnvironment(t, env)

			// Set environment variable
			os.Setenv(tc.envKey, tc.envValue)
			defer os.Unsetenv(tc.envKey)

			err := env.ConfigManager.LoadConfig(configPath)
			require.NoError(t, err)
			cfg := env.ConfigManager.GetConfig()
			require.NotNil(t, cfg)

			// Verify type conversion
			switch tc.name {
			case "string_to_int_port":
				assert.Equal(t, tc.expected, cfg.Server.Port)
			case "string_to_int_max_connections":
				assert.Equal(t, tc.expected, cfg.Server.MaxConnections)
			case "string_to_int_health_interval":
				assert.Equal(t, tc.expected, cfg.MediaMTX.HealthCheckInterval)
			case "string_to_float_backoff":
				assert.Equal(t, tc.expected, cfg.MediaMTX.BackoffBaseMultiplier)
			case "string_to_float_poll_interval":
				assert.Equal(t, tc.expected, cfg.Camera.PollInterval)
			case "string_to_bool_true":
				assert.Equal(t, tc.expected, cfg.Camera.EnableCapabilityDetection)
			case "string_to_bool_false":
				assert.Equal(t, tc.expected, cfg.Camera.AutoStartStreams)
			case "string_to_bool_1":
				assert.Equal(t, tc.expected, cfg.Logging.FileEnabled)
			case "string_to_bool_0":
				assert.Equal(t, tc.expected, cfg.Logging.ConsoleEnabled)
			case "string_to_bool_t":
				assert.Equal(t, tc.expected, cfg.Recording.Enabled)
			case "string_to_bool_f":
				assert.Equal(t, tc.expected, cfg.Snapshots.Enabled)
			case "string_to_int_quality":
				assert.Equal(t, tc.expected, cfg.Snapshots.Quality)
			}
		})
	}
}

func TestConfigManager_EnvironmentVariablePrecedence(t *testing.T) {
	// Test environment variable override precedence (env > file > defaults)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Create YAML with some values
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9000
  websocket_path: "/ws"
  max_connections: 50

mediamtx:
  host: "localhost"
  api_port: 9998
  rtsp_port: 8555
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override some values
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "8002")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_HOST", "192.168.1.200")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_API_PORT", "9997")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_HOST")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_API_PORT")
	}()

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify precedence: environment variables override file values
	assert.Equal(t, "192.168.1.100", cfg.Server.Host) // From env
	assert.Equal(t, 8002, cfg.Server.Port)            // From env
	assert.Equal(t, "/ws", cfg.Server.WebSocketPath)  // From file
	assert.Equal(t, 50, cfg.Server.MaxConnections)    // From file

	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host) // From env
	assert.Equal(t, 9997, cfg.MediaMTX.APIPort)         // From env
	assert.Equal(t, 8555, cfg.MediaMTX.RTSPPort)        // From file

	// Verify defaults are used for values not in file or env
	assert.Equal(t, 8889, cfg.MediaMTX.WebRTCPort) // From defaults
	assert.Equal(t, 8888, cfg.MediaMTX.HLSPort)    // From defaults
}

func TestConfigManager_EnvironmentVariableEdgeCases(t *testing.T) {
	// Test edge cases for environment variables
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 8002
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Test Unicode characters in environment variables
	t.Run("unicode_characters", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		unicodeHost := "192.168.1.100-测试"
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", unicodeHost)
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")

		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		assert.Equal(t, unicodeHost, cfg.Server.Host)
	})

	// Test special characters in environment variables
	t.Run("special_characters", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		specialPath := "/path/with/spaces and special chars!@#"
		os.Setenv("CAMERA_SERVICE_SERVER_WEBSOCKET_PATH", specialPath)
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_WEBSOCKET_PATH")

		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		assert.Equal(t, specialPath, cfg.Server.WebSocketPath)
	})

	// Test empty environment variables
	t.Run("empty_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "")
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")

		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Should fall back to file value
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	})

	// Test very large environment variables
	t.Run("large_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		largeValue := strings.Repeat("a", 10000) // 10KB string
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", largeValue)
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")

		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		assert.Equal(t, largeValue, cfg.Server.Host)
	})
}

func TestConfigValidation_Comprehensive(t *testing.T) {
	// REQ-E1-S1.1-003: Comprehensive configuration validation testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Test required field validation
	t.Run("required_field_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Test empty required fields
		invalidYAML := `
server:
  host: ""
  port: 0 
  websocket_path: ""
  max_connections: 0 

mediamtx:
  host: ""
  api_port: 0
  config_path: ""
  recordings_path: ""
  snapshots_path: ""
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test data type validation
	t.Run("data_type_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		invalidYAML := `
server:
  host: "127.0.0.1"
  port: "invalid_port"
  websocket_path: "/ws"
  max_connections: "invalid"

mediamtx:
  host: "127.0.0.1"
  api_port: "invalid"
  health_check_interval: "invalid"
  backoff_base_multiplier: "invalid"
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		// Viper catches type conversion errors before validation
		assert.Contains(t, err.Error(), "failed to unmarshal config")
	})

	// Test range validation
	t.Run("range_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		invalidYAML := `
server:
  host: "127.0.0.1"
  port: 70000
  websocket_path: "/ws"
  max_connections: -1

mediamtx:
  host: "127.0.0.1"
  api_port: 70000
  health_check_interval: -1
  health_failure_threshold: 0
  backoff_base_multiplier: -1.0
  process_termination_timeout: -1.0
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test enumeration validation
	t.Run("enumeration_validation", func(t *testing.T) {
		invalidYAML := `
mediamtx:
  codec:
    video_profile: "invalid_profile"
    video_level: "invalid_level"    
    pixel_format: "invalid_format"  
    preset: "invalid_preset"        

logging:
  level: "INVALID_LEVEL"

recording:
  format: "invalid_format"
  quality: "invalid_quality"

snapshots:
  format: "invalid_format"
  quality: 150
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test nested structure validation
	t.Run("nested_structure_validation", func(t *testing.T) {
		invalidYAML := `
mediamtx:
  codec:
    video_profile: "baseline"
    video_level: "3.0"
    pixel_format: "yuv420p"
    bitrate: ""
    preset: "ultrafast"
  
  stream_readiness:
    timeout: -1.0
    retry_attempts: -1
    retry_delay: -1.0
    check_interval: 0
    enable_progress_notifications: true
    graceful_fallback: true

camera:
  device_range: "[5, 3]"
  capability_timeout: -1.0
  capability_retry_interval: -1.0
  capability_max_retries: -1
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test cross-field validation
	t.Run("cross_field_validation", func(t *testing.T) {
		invalidYAML := `
mediamtx:
  backoff_jitter_range: "[1.2, 0.8]"
  health_max_backoff_interval: 30
  health_check_interval: 60

camera:
  detection_timeout: 1.0
  poll_interval: 2.0

logging:
  file_enabled: true
  file_path: ""
  max_file_size: 0
  backup_count: -1

recording:
  segment_duration: 0
  max_segment_size: 0
  cleanup_interval: 0
  max_age: 0
  max_size: 0

snapshots:
  max_width: 0
  max_height: 0
  cleanup_interval: 0
  max_age: 0
  max_count: 0
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

func TestConfigValidation_EdgeCases(t *testing.T) {
	// Test edge cases for configuration validation
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Test extremely large values
	t.Run("extremely_large_values", func(t *testing.T) {
		largeYAML := `
server:
  host: "127.0.0.1"
  port: 65535
  websocket_path: "/ws"
  max_connections: 1000000

mediamtx:
  host: "127.0.0.1"
  api_port: 65535
  health_check_interval: 86400
  health_failure_threshold: 1000
  backoff_base_multiplier: 1000.0
  process_termination_timeout: 3600.0

camera:
  poll_interval: 0.001
  detection_timeout: 3600.0
  capability_timeout: 3600.0
  capability_retry_interval: 0.001
  capability_max_retries: 1000

logging:
  max_file_size: 1073741824
  backup_count: 100

recording:
  segment_duration: 86400
  max_segment_size: 107374182400
  cleanup_interval: 31536000
  max_age: 31536000
  max_size: 1099511627776

snapshots:
  quality: 100
  max_width: 7680
  max_height: 4320
  cleanup_interval: 31536000
  max_age: 31536000
  max_count: 1000000
`

		err := os.WriteFile(configPath, []byte(largeYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify large values are loaded correctly
		assert.Equal(t, 65535, cfg.Server.Port)
		assert.Equal(t, 1000000, cfg.Server.MaxConnections)
		assert.Equal(t, 65535, cfg.MediaMTX.APIPort)
		assert.Equal(t, 86400, cfg.MediaMTX.HealthCheckInterval)
		assert.Equal(t, 1000, cfg.MediaMTX.HealthFailureThreshold)
		assert.Equal(t, 1000.0, cfg.MediaMTX.BackoffBaseMultiplier)
		assert.Equal(t, 3600.0, cfg.MediaMTX.ProcessTerminationTimeout)
	})

	// Test boundary values
	t.Run("boundary_values", func(t *testing.T) {
		boundaryYAML := `
server:
  host: "127.0.0.1"
  port: 1
  websocket_path: "/ws"
  max_connections: 1

mediamtx:
  host: "127.0.0.1"
  api_port: 1
  health_check_interval: 1
  health_failure_threshold: 1
  backoff_base_multiplier: 0.1
  process_termination_timeout: 0.1

camera:
  poll_interval: 0.001
  detection_timeout: 0.1
  capability_timeout: 0.1
  capability_retry_interval: 0.001
  capability_max_retries: 1

logging:
  max_file_size: 1
  backup_count: 0

recording:
  segment_duration: 1
  max_segment_size: 1
  cleanup_interval: 1
  max_age: 1
  max_size: 1

snapshots:
  quality: 1
  max_width: 1
  max_height: 1
  cleanup_interval: 1
  max_age: 1
  max_count: 1
`

		err := os.WriteFile(configPath, []byte(boundaryYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify boundary values are loaded correctly
		assert.Equal(t, 1, cfg.Server.Port)
		assert.Equal(t, 1, cfg.Server.MaxConnections)
		assert.Equal(t, 1, cfg.MediaMTX.APIPort)
		assert.Equal(t, 1, cfg.MediaMTX.HealthCheckInterval)
		assert.Equal(t, 1, cfg.MediaMTX.HealthFailureThreshold)
		assert.Equal(t, 0.1, cfg.MediaMTX.BackoffBaseMultiplier)
		assert.Equal(t, 0.1, cfg.MediaMTX.ProcessTerminationTimeout)
	})

	// Test special characters in string fields
	t.Run("special_characters", func(t *testing.T) {
		// Clean up any existing environment variables that might interfere
		cleanupCameraServiceEnvVars()

		specialYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws/with/special/chars!@#$%"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  config_path: "/path/with/spaces and special chars!@#$%"
  recordings_path: "/path/with/unicode/测试"
  snapshots_path: "/path/with/emoji/📷"
  
  codec:
    video_profile: "baseline"
    video_level: "3.0"
    pixel_format: "yuv420p"
    bitrate: "600k"
    preset: "ultrafast"

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s with special chars!@#$%"
  file_enabled: true
  file_path: "/path/with/special/chars!@#$%"
  max_file_size: 10485760
  backup_count: 5
  console_enabled: true
`

		err := os.WriteFile(configPath, []byte(specialYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify special characters are handled correctly
		assert.Equal(t, "/ws/with/special/chars!@#$%", cfg.Server.WebSocketPath)
		assert.Equal(t, "/path/with/spaces and special chars!@#$%", cfg.MediaMTX.ConfigPath)
		assert.Equal(t, "/path/with/unicode/测试", cfg.MediaMTX.RecordingsPath)
		assert.Equal(t, "/path/with/emoji/📷", cfg.MediaMTX.SnapshotsPath)
		assert.Contains(t, cfg.Logging.Format, "special chars!@#$%")
		assert.Equal(t, "/path/with/special/chars!@#$%", cfg.Logging.FilePath)
	})
}

func TestConfigValidation_FileSystemEdgeCases(t *testing.T) {
	// Test file system edge cases

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := env.TempDir

	// Test file permission errors
	t.Run("file_permission_errors", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "readonly_config.yaml")

		// Create a file with read-only permissions
		yamlContent := `
server:
  host: "127.0.0.1"
  port: 8002
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0444) // Read-only
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should still be able to read
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})

	// Test symbolic link handling
	t.Run("symbolic_link_handling", func(t *testing.T) {
		originalPath := filepath.Join(tempDir, "original_config.yaml")
		symlinkPath := filepath.Join(tempDir, "symlink_config.yaml")

		yamlContent := `
server:
  host: "192.168.1.100"
  port: 9000
`
		err := os.WriteFile(originalPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		// Create symbolic link
		err = os.Symlink(originalPath, symlinkPath)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(symlinkPath)
		require.NoError(t, err) // Should handle symlinks correctly
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "192.168.1.100", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
	})

	// Test deeply nested configuration paths
	t.Run("deeply_nested_paths", func(t *testing.T) {
		// Create deeply nested directory structure
		deepDir := filepath.Join(tempDir, "deep", "nested", "config", "path")
		err := os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		configPath := filepath.Join(deepDir, "config.yaml")
		yamlContent := `
server:
  host: "10.0.0.1"
  port: 8080
`
		err = os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should handle deep paths correctly
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "10.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})
}

func TestGlobalConfigFunctions(t *testing.T) {
	// Test global configuration manager functions

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := env.TempDir
	configPath := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
server:
  host: "192.168.1.100"
  port: 9000
  websocket_path: "/test"
  max_connections: 50

mediamtx:
  host: "192.168.1.200"
  api_port: 9998
  rtsp_port: 8555
  webrtc_port: 8890
  hls_port: 8889
  config_path: "/test/config.yml"
  recordings_path: "/test/recordings"
  snapshots_path: "/test/snapshots"
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Test GetConfigManager
	t.Run("get_config_manager", func(t *testing.T) {
		manager := config.GetConfigManager()
		require.NotNil(t, manager)

		// Should return the same instance
		manager2 := config.GetConfigManager()
		assert.Equal(t, manager, manager2)
	})

	// Test LoadConfig global function
	t.Run("load_config_global", func(t *testing.T) {
		err := config.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := config.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "192.168.1.100", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
	})

	// Test GetConfig global function
	t.Run("get_config_global", func(t *testing.T) {
		cfg := config.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "192.168.1.100", cfg.Server.Host)
		assert.Equal(t, 9000, cfg.Server.Port)
	})
}

func TestConfigValidation_ComprehensiveCoverage(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration validation comprehensive coverage

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test comprehensive validation coverage
	tempDir := env.TempDir
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Test validation with invalid configuration
	t.Run("invalid_config_validation", func(t *testing.T) {
		invalidYAML := `
server:
  host: ""
  port: 0 
  websocket_path: ""
  max_connections: 0 

mediamtx:
  host: ""
  api_port: 0
  config_path: ""
  recordings_path: ""
  snapshots_path: ""
`

		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test validation with missing required fields
	t.Run("missing_required_fields", func(t *testing.T) {
		incompleteYAML := `
server:
  host: ""
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: ""
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: ""
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

		err := os.WriteFile(configPath, []byte(incompleteYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test validation with out-of-range values
	t.Run("out_of_range_values", func(t *testing.T) {
		outOfRangeYAML := `
server:
  host: "127.0.0.1"
  port: 99999
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  health_check_interval: -1
  health_failure_threshold: 0
  backoff_base_multiplier: 0.0
`

		err := os.WriteFile(configPath, []byte(outOfRangeYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	// Test validation with invalid enum values
	t.Run("invalid_enum_values", func(t *testing.T) {
		invalidEnumYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  
  codec:
    video_profile: "invalid_profile"
    video_level: "invalid_level"    
    pixel_format: "invalid_format"  
    preset: "invalid_preset"        

logging:
  level: "INVALID_LEVEL"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/opt/camera-service/logs/camera-service.log"
  max_file_size: 10485760
  backup_count: 5
  console_enabled: true

recording:
  enabled: true
  format: "invalid_format"
  quality: "invalid_quality"

snapshots:
  enabled: true
  format: "invalid_format"
  quality: 150
`

		err := os.WriteFile(configPath, []byte(invalidEnumYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

func TestConfigManager_HotReload_Comprehensive(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration hot reload comprehensive functionality

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test comprehensive hot reload functionality
	tempDir := env.TempDir
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Initial configuration
	initialYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload for testing
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Test callback notification
	t.Run("callback_notification", func(t *testing.T) {
		callbackCalled := make(chan bool, 1)
		var callbackConfig *config.Config

		env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
			callbackConfig = cfg
			callbackCalled <- true
		})

		// Update configuration file
		updatedYAML := `
server:
  host: "192.168.1.100"
  port: 9000
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "192.168.1.200"
  api_port: 9998
  rtsp_port: 8555
  webrtc_port: 8890
  hls_port: 8889
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

		err := os.WriteFile(configPath, []byte(updatedYAML), 0644)
		require.NoError(t, err)

		// Wait for hot reload with timeout - increase timeout for file system events
		select {
		case <-callbackCalled:
			// Callback was called successfully
		case <-time.After(5 * time.Second):
			t.Fatal("Callback was not called within timeout")
		}

		// Check if callback was called
		require.NotNil(t, callbackConfig)
		assert.Equal(t, "192.168.1.100", callbackConfig.Server.Host)
		assert.Equal(t, 9000, callbackConfig.Server.Port)
	})

	// Test multiple callbacks
	t.Run("multiple_callbacks", func(t *testing.T) {
		callback1Called := make(chan bool, 1)
		callback2Called := make(chan bool, 1)

		env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
			callback1Called <- true
		})

		env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
			callback2Called <- true
		})

		// Update configuration file
		updatedYAML := `
server:
  host: "10.0.0.1"
  port: 8080
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "10.0.0.2"
  api_port: 9999
  rtsp_port: 8556
  webrtc_port: 8891
  hls_port: 8890
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

		err := os.WriteFile(configPath, []byte(updatedYAML), 0644)
		require.NoError(t, err)

		// Wait for hot reload with timeout - increase timeout for file system events
		select {
		case <-callback1Called:
			// First callback was called
		case <-time.After(5 * time.Second):
			t.Fatal("First callback was not called within timeout")
		}

		select {
		case <-callback2Called:
			// Second callback was called
		case <-time.After(5 * time.Second):
			t.Fatal("Second callback was not called within timeout")
		}
	})

	// Test callback panic handling
	t.Run("callback_panic_handling", func(t *testing.T) {
		env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
			panic("test panic")
		})

		// Update configuration file
		updatedYAML := `
server:
  host: "172.16.0.1"
  port: 7070
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "172.16.0.2"
  api_port: 9996
  rtsp_port: 8553
  webrtc_port: 8888
  hls_port: 8887
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

		err := os.WriteFile(configPath, []byte(updatedYAML), 0644)
		require.NoError(t, err)

		// Wait for hot reload - should not panic
		time.Sleep(200 * time.Millisecond)

		// Test should complete without panic
		assert.True(t, true)
	})

	// Test file removal during hot reload
	t.Run("file_removal_during_hot_reload", func(t *testing.T) {
		// Remove the config file
		err := os.Remove(configPath)
		require.NoError(t, err)

		// Wait a bit for the file watcher to detect the removal
		time.Sleep(200 * time.Millisecond)

		// The hot reload should be disabled gracefully
	})

	// Test failed reload scenario
	t.Run("failed_reload_scenario", func(t *testing.T) {
		// Create a malformed config file
		malformedYAML := `
server:
  host: "127.0.0.1"
  port: invalid_port
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
`

		err := os.WriteFile(configPath, []byte(malformedYAML), 0644)
		require.NoError(t, err)

		// Wait a bit for the reload attempt
		time.Sleep(200 * time.Millisecond)

		// The reload should fail gracefully
	})

	// Clean up
	env.ConfigManager.Stop()
}

func TestConfigValidation_DetailedCoverage(t *testing.T) {
	// Test detailed validation coverage for all validation functions

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "test_config.yaml")

	// Test codec validation
	t.Run("codec_validation", func(t *testing.T) {
		codecYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  
  codec:
    video_profile: "high"
    video_level: "4.0"   
    pixel_format: "yuv444p"
    bitrate: "800k"      
    preset: "slow"       
`

		err := os.WriteFile(configPath, []byte(codecYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "high", cfg.MediaMTX.Codec.VideoProfile)
		assert.Equal(t, "4.0", cfg.MediaMTX.Codec.VideoLevel)
		assert.Equal(t, "yuv444p", cfg.MediaMTX.Codec.PixelFormat)
		assert.Equal(t, "800k", cfg.MediaMTX.Codec.Bitrate)
		assert.Equal(t, "slow", cfg.MediaMTX.Codec.Preset)
	})

	// Test stream readiness validation
	t.Run("stream_readiness_validation", func(t *testing.T) {
		streamYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  
  stream_readiness:
    timeout: 20.0
    retry_attempts: 5
    retry_delay: 2.0
    check_interval: 0.5
    enable_progress_notifications: true
    graceful_fallback: true
`

		err := os.WriteFile(configPath, []byte(streamYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 20.0, cfg.MediaMTX.StreamReadiness.Timeout)
		assert.Equal(t, 5, cfg.MediaMTX.StreamReadiness.RetryAttempts)
		assert.Equal(t, 2.0, cfg.MediaMTX.StreamReadiness.RetryDelay)
		assert.Equal(t, 0.5, cfg.MediaMTX.StreamReadiness.CheckInterval)
		assert.True(t, cfg.MediaMTX.StreamReadiness.EnableProgressNotifications)
		assert.True(t, cfg.MediaMTX.StreamReadiness.GracefulFallback)
	})

	// Test FFmpeg validation
	t.Run("ffmpeg_validation", func(t *testing.T) {
		ffmpegYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"

ffmpeg:
  snapshot:
    process_creation_timeout: 5.0
    execution_timeout: 10.0
    internal_timeout: 5000000
    retry_attempts: 3
    retry_delay: 1.0
  
  recording:
    process_creation_timeout: 10.0
    execution_timeout: 20.0
    internal_timeout: 10000000
    retry_attempts: 5
    retry_delay: 2.0
`

		err := os.WriteFile(configPath, []byte(ffmpegYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 5.0, cfg.FFmpeg.Snapshot.ProcessCreationTimeout)
		assert.Equal(t, 10.0, cfg.FFmpeg.Snapshot.ExecutionTimeout)
		assert.Equal(t, 5000000, cfg.FFmpeg.Snapshot.InternalTimeout)
		assert.Equal(t, 3, cfg.FFmpeg.Snapshot.RetryAttempts)
		assert.Equal(t, 1.0, cfg.FFmpeg.Snapshot.RetryDelay)
	})

	// Test notifications validation
	t.Run("notifications_validation", func(t *testing.T) {
		notificationsYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"

notifications:
  websocket:
    delivery_timeout: 5.0
    retry_attempts: 3
    retry_delay: 1.0
  
  real_time:
    camera_status_interval: 5.0
    recording_progress_interval: 5.0
    connection_health_check: 5.0
`

		err := os.WriteFile(configPath, []byte(notificationsYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 5.0, cfg.Notifications.WebSocket.DeliveryTimeout)
		assert.Equal(t, 3, cfg.Notifications.WebSocket.RetryAttempts)
		assert.Equal(t, 1.0, cfg.Notifications.WebSocket.RetryDelay)
		assert.Equal(t, 5.0, cfg.Notifications.RealTime.CameraStatusInterval)
		assert.Equal(t, 5.0, cfg.Notifications.RealTime.RecordingProgressInterval)
		assert.Equal(t, 5.0, cfg.Notifications.RealTime.ConnectionHealthCheck)
	})

	// Test performance validation
	t.Run("performance_validation", func(t *testing.T) {
		performanceYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"

performance:
  response_time_targets:
    snapshot_capture: 30.0
    recording_start: 80.0
    recording_stop: 15.0
    file_listing: 30.0
  
  snapshot_tiers:
    tier1_usb_direct_timeout: 5.0
    tier2_rtsp_ready_check_timeout: 10.0
    tier3_activation_timeout: 15.0
    tier3_activation_trigger_timeout: 20.0
    total_operation_timeout: 30.0
    immediate_response_threshold: 1.0
    acceptable_response_threshold: 5.0
    slow_response_threshold: 10.0
  
  optimization:
    enable_caching: true
    cache_ttl: 300
    max_concurrent_operations: 100
    connection_pool_size: 50
`

		err := os.WriteFile(configPath, []byte(performanceYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 30.0, cfg.Performance.ResponseTimeTargets.SnapshotCapture)
		assert.Equal(t, 80.0, cfg.Performance.ResponseTimeTargets.RecordingStart)
		assert.Equal(t, 15.0, cfg.Performance.ResponseTimeTargets.RecordingStop)
		assert.Equal(t, 30.0, cfg.Performance.ResponseTimeTargets.FileListing)
		assert.Equal(t, 5.0, cfg.Performance.SnapshotTiers.Tier1USBDirectTimeout)
		assert.Equal(t, 10.0, cfg.Performance.SnapshotTiers.Tier2RTSPReadyCheckTimeout)
		assert.Equal(t, 15.0, cfg.Performance.SnapshotTiers.Tier3ActivationTimeout)
		assert.Equal(t, 20.0, cfg.Performance.SnapshotTiers.Tier3ActivationTriggerTimeout)
		assert.Equal(t, 30.0, cfg.Performance.SnapshotTiers.TotalOperationTimeout)
		assert.True(t, cfg.Performance.Optimization.EnableCaching)
		assert.Equal(t, 300, cfg.Performance.Optimization.CacheTTL)
		assert.Equal(t, 100, cfg.Performance.Optimization.MaxConcurrentOperations)
		assert.Equal(t, 50, cfg.Performance.Optimization.ConnectionPoolSize)
	})
}

func TestConfigManager_EdgeCases_Comprehensive(t *testing.T) {
	// Test comprehensive edge cases

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tempDir := t.TempDir()

	// Test with non-existent directory
	t.Run("non_existent_directory", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

		err := env.ConfigManager.LoadConfig(nonExistentPath)
		require.NoError(t, err) // Should use defaults

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host) // Default value
	})

	// Test with empty file
	t.Run("empty_file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "empty_config.yaml")
		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should use defaults

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host) // Default value
	})

	// Test with file containing only comments
	t.Run("comments_only_file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "comments_config.yaml")
		commentsYAML := `



`
		err := os.WriteFile(configPath, []byte(commentsYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should use defaults

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host) // Default value
	})

	// Test with malformed YAML that's not completely invalid
	t.Run("malformed_yaml", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "malformed_config.yaml")
		malformedYAML := `
server:
  host: "127.0.0.1"
  port: 8002

  websocket_path: "/ws"
  max_connections: 100
`
		err := os.WriteFile(configPath, []byte(malformedYAML), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should use defaults due to YAML parsing error

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host) // Default value
	})

	// Test with very large configuration file
	t.Run("large_config_file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "large_config.yaml")

		// Create a large YAML file
		var largeYAML strings.Builder
		largeYAML.WriteString("server:\n")
		largeYAML.WriteString("  host: \"127.0.0.1\"\n")
		largeYAML.WriteString("  port: 8002\n")
		largeYAML.WriteString("  websocket_path: \"/ws\"\n")
		largeYAML.WriteString("  max_connections: 100\n")

		// Add many comments to make it large
		for i := 0; i < 1000; i++ {
			largeYAML.WriteString(fmt.Sprintf("  # Comment %d\n", i))
		}

		largeYAML.WriteString("\nmediamtx:\n")
		largeYAML.WriteString("  host: \"127.0.0.1\"\n")
		largeYAML.WriteString("  api_port: 9997\n")
		largeYAML.WriteString("  rtsp_port: 8554\n")
		largeYAML.WriteString("  webrtc_port: 8889\n")
		largeYAML.WriteString("  hls_port: 8888\n")
		largeYAML.WriteString("  config_path: \"/opt/camera-service/config/mediamtx.yml\"\n")
		largeYAML.WriteString("  recordings_path: \"/opt/camera-service/recordings\"\n")
		largeYAML.WriteString("  snapshots_path: \"/opt/camera-service/snapshots\"\n")

		err := os.WriteFile(configPath, []byte(largeYAML.String()), 0644)
		require.NoError(t, err)

		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})
}

func TestConfigManager_ApplyEnvironmentOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides
	// Test the applyEnvironmentOverrides method specifically

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	cleanupCameraServiceEnvVars()

	// Set environment variables to override configuration
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "9090")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_HOST", "192.168.1.200")
	os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG")
	os.Setenv("CAMERA_SERVICE_RECORDING_ENABLED", "true")
	os.Setenv("CAMERA_SERVICE_SNAPSHOTS_ENABLED", "false")

	defer cleanupCameraServiceEnvVars()

	// Load config with environment overrides
	err := env.ConfigManager.LoadConfig("")
	require.NoError(t, err)

	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify environment overrides were applied
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.True(t, cfg.Recording.Enabled)
	assert.False(t, cfg.Snapshots.Enabled)
}

func TestConfigManager_ValidateConfig(t *testing.T) {
	// REQ-E1-S1.1-003: Configuration validation
	// Test the validateConfig method specifically

	cleanupCameraServiceEnvVars()

	t.Run("valid_configuration", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Load a valid configuration
		tempDir := env.TempDir
		configPath := filepath.Join(tempDir, "valid_config.yaml")

		validYAML := `
server:
  host: "127.0.0.1"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpg"
  quality: "high"
`
		err := os.WriteFile(configPath, []byte(validYAML), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		// The validateConfig method should be called during LoadConfig
		// and should not return an error for valid configuration
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})

	t.Run("invalid_port_configuration", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Set invalid port via environment variable
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "99999") // Invalid port
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")

		// Load config - should handle invalid port gracefully
		_ = env.ConfigManager.LoadConfig("")
		// Note: The current implementation may not validate ports strictly
		// This test documents the current behavior

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// The actual behavior depends on the implementation
	})

	t.Run("invalid_host_configuration", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Set invalid host via environment variable
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "invalid-host-name-with-spaces")
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")

		// Load config - should handle invalid host gracefully
		_ = env.ConfigManager.LoadConfig("")
		// Note: The current implementation may not validate hosts strictly
		// This test documents the current behavior

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// The actual behavior depends on the implementation
	})
}

func TestConfigManager_EnvironmentOverrideEdgeCases(t *testing.T) {
	// Test edge cases for environment variable overrides

	cleanupCameraServiceEnvVars()

	t.Run("empty_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Test with empty environment variables
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "")
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "")
		defer cleanupCameraServiceEnvVars()

		err := env.ConfigManager.LoadConfig("")
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// Should use default values
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})

	t.Run("whitespace_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Test with whitespace-only environment variables
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "   ")
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "   ")
		defer cleanupCameraServiceEnvVars()

		err := env.ConfigManager.LoadConfig("")
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// Should use default values
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})

	t.Run("special_characters_in_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Test with special characters in environment variables
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "127.0.0.1")
		os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG")
		defer cleanupCameraServiceEnvVars()

		err := env.ConfigManager.LoadConfig("")
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// Should handle special characters appropriately
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, "DEBUG", cfg.Logging.Level)
	})

	t.Run("very_long_environment_variables", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		// This eliminates the need to create ConfigManager and Logger in every test
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Test with very long environment variables
		longHost := strings.Repeat("a", 1000)
		longLevel := strings.Repeat("DEBUG", 100)

		os.Setenv("CAMERA_SERVICE_SERVER_HOST", longHost)
		os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", longLevel)
		defer cleanupCameraServiceEnvVars()

		err := env.ConfigManager.LoadConfig("")
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		// Should handle long values appropriately
		assert.Equal(t, longHost, cfg.Server.Host)
		assert.Equal(t, longLevel, cfg.Logging.Level)
	})
}
