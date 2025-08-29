//go:build unit
// +build unit

package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
)

/*
Module: Configuration Private Methods Testing

Requirements Coverage:
- REQ-E1-S1.1-001: Configuration loading from YAML files
- REQ-E1-S1.1-002: Environment variable overrides
- REQ-E1-S1.1-003: Configuration validation
- REQ-CONFIG-001: The system SHALL validate configuration files before loading
- REQ-CONFIG-002: The system SHALL fail fast on configuration errors
- REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting

Test Categories: Unit
API Documentation Reference: N/A (Configuration system)

PURPOSE: Test private methods through their public interfaces to achieve >90% coverage
*/

func TestConfigManager_PrivateMethods_FileWatching(t *testing.T) {
	// Test startFileWatching, stopFileWatching, watchFileChanges through hot reload
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "hot_reload_test.yaml")
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload to exercise file watching methods
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	// Load config - this exercises startFileWatching internally
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Verify initial configuration
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)

	// Test file watching by modifying the file
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.05
  detection_timeout: 1.5
  device_range: [0, 4]
  enable_capability_detection: false
  auto_start_streams: false

logging:
  level: "DEBUG"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 85

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	// Update file - this exercises watchFileChanges and reloadConfiguration
	err = os.WriteFile(configPath, []byte(updatedYAML), 0644)
	require.NoError(t, err)

	// Wait for hot reload to process
	time.Sleep(500 * time.Millisecond)

	// Verify configuration was updated
	cfg = env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)

	// Test stopFileWatching through Stop method
	env.ConfigManager.Stop()

	// Verify configuration is still accessible after stop
	cfg = env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
}

func TestConfigManager_PrivateMethods_Callbacks(t *testing.T) {
	// Test AddUpdateCallback and notifyConfigUpdated through hot reload
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "callback_test.yaml")
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	// Load initial configuration
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Test multiple callbacks - this exercises AddUpdateCallback and notifyConfigUpdated
	callback1Called := make(chan bool, 1)
	callback2Called := make(chan bool, 1)
	var callback1Config, callback2Config *config.Config

	env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
		callback1Config = cfg
		callback1Called <- true
	})

	env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
		callback2Config = cfg
		callback2Called <- true
	})

	// Test callback panic handling
	env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
		panic("test panic in callback")
	})

	// Update configuration to trigger callbacks
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.05
  detection_timeout: 1.5
  device_range: [0, 4]
  enable_capability_detection: false
  auto_start_streams: false

logging:
  level: "DEBUG"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 85

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err = os.WriteFile(configPath, []byte(updatedYAML), 0644)
	require.NoError(t, err)

	// Wait for callbacks to be triggered
	select {
	case <-callback1Called:
		// First callback was called
	case <-time.After(2 * time.Second):
		t.Fatal("First callback was not called within timeout")
	}

	select {
	case <-callback2Called:
		// Second callback was called
	case <-time.After(2 * time.Second):
		t.Fatal("Second callback was not called within timeout")
	}

	// Verify callback configurations
	require.NotNil(t, callback1Config)
	require.NotNil(t, callback2Config)
	assert.Equal(t, "10.0.0.1", callback1Config.Server.Host)
	assert.Equal(t, 8080, callback1Config.Server.Port)
	assert.Equal(t, "10.0.0.2", callback1Config.MediaMTX.Host)
	assert.Equal(t, 9999, callback1Config.MediaMTX.APIPort)

	// Verify both callbacks received the same configuration
	assert.Equal(t, callback1Config.Server.Host, callback2Config.Server.Host)
	assert.Equal(t, callback1Config.Server.Port, callback2Config.Server.Port)

	// Clean up
	env.ConfigManager.Stop()
}

func TestConfigManager_PrivateMethods_SaveConfig(t *testing.T) {
	// Test SaveConfig and setConfigValues through configuration persistence
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "save_test.yaml")
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Load configuration
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Verify initial configuration
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)

	// Test SaveConfig - this exercises setConfigValues internally
	err = env.ConfigManager.SaveConfig()
	require.NoError(t, err)

	// Verify configuration was saved correctly by reloading
	newManager := config.CreateConfigManager()
	err = newManager.LoadConfig(configPath)
	require.NoError(t, err)

	savedCfg := newManager.GetConfig()
	require.NotNil(t, savedCfg)
	assert.Equal(t, cfg.Server.Host, savedCfg.Server.Host)
	assert.Equal(t, cfg.Server.Port, savedCfg.Server.Port)
	assert.Equal(t, cfg.MediaMTX.Host, savedCfg.MediaMTX.Host)
	assert.Equal(t, cfg.MediaMTX.APIPort, savedCfg.MediaMTX.APIPort)
}

func TestConfigManager_PrivateMethods_EnvironmentOverrides(t *testing.T) {
	// Test applyEnvironmentOverrides through environment variable processing
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "env_override_test.yaml")
	baseYAML := `
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err := os.WriteFile(configPath, []byte(baseYAML), 0644)
	require.NoError(t, err)

	// Set environment variables to test overrides - this exercises applyEnvironmentOverrides
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "9000")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_HOST", "192.168.1.200")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_API_PORT", "9998")
	os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG")
	os.Setenv("CAMERA_SERVICE_CAMERA_POLL_INTERVAL", "0.05")
	os.Setenv("CAMERA_SERVICE_RECORDING_ENABLED", "true")
	os.Setenv("CAMERA_SERVICE_SNAPSHOTS_QUALITY", "85")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_HOST")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_API_PORT")
		os.Unsetenv("CAMERA_SERVICE_LOGGING_LEVEL")
		os.Unsetenv("CAMERA_SERVICE_CAMERA_POLL_INTERVAL")
		os.Unsetenv("CAMERA_SERVICE_RECORDING_ENABLED")
		os.Unsetenv("CAMERA_SERVICE_SNAPSHOTS_QUALITY")
	}()

	// Load configuration with environment overrides
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Verify environment overrides were applied
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, 0.05, cfg.Camera.PollInterval)
	assert.True(t, cfg.Recording.Enabled)
	assert.Equal(t, 85, cfg.Snapshots.Quality)
}

func TestConfigManager_PrivateMethods_ValidationIntegration(t *testing.T) {
	// Test validateConfig through configuration validation
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test validation with invalid configuration
	invalidConfigPath := filepath.Join(env.TempDir, "invalid_config.yaml")
	invalidYAML := `
server:
  host: ""
  port: 0
  websocket_path: ""
  max_connections: 0

mediamtx:
  host: ""
  api_port: 0
  rtsp_port: 0
  webrtc_port: 0
  hls_port: 0
  config_path: ""
  recordings_path: ""
  snapshots_path: ""

camera:
  poll_interval: -1
  detection_timeout: 0
  device_range: [5, 3]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INVALID"
  format: ""
  file_enabled: true
  file_path: ""
  console_enabled: true

recording:
  enabled: true
  format: "INVALID"
  quality: "INVALID"

snapshots:
  enabled: true
  format: "INVALID"
  quality: 150
`

	err := os.WriteFile(invalidConfigPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// Load invalid configuration - this should exercise validateConfig
	err = env.ConfigManager.LoadConfig(invalidConfigPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")

	// Test validation with valid configuration
	validConfigPath := filepath.Join(env.TempDir, "valid_config.yaml")
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err = os.WriteFile(validConfigPath, []byte(validYAML), 0644)
	require.NoError(t, err)

	// Load valid configuration
	err = env.ConfigManager.LoadConfig(validConfigPath)
	require.NoError(t, err)

	// Verify valid configuration was loaded
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_PrivateMethods_FileWatchingEdgeCases(t *testing.T) {
	// Test file watching edge cases to exercise watchFileChanges and reloadConfiguration
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "edge_case_test.yaml")
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: false
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 90

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err := os.WriteFile(configPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	// Load initial configuration
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Test file removal during hot reload
	err = os.Remove(configPath)
	require.NoError(t, err)

	// Wait for file watcher to detect removal
	time.Sleep(200 * time.Millisecond)

	// Verify configuration is still accessible after file removal
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)

	// Test failed reload scenario
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"
`

	err = os.WriteFile(configPath, []byte(malformedYAML), 0644)
	require.NoError(t, err)

	// Wait for reload attempt
	time.Sleep(200 * time.Millisecond)

	// Configuration should remain unchanged due to reload failure
	cfg = env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)

	// Clean up
	env.ConfigManager.Stop()
}

func TestConfigManager_PrivateMethods_ComprehensiveCoverage(t *testing.T) {
	// Comprehensive test to exercise all private methods and achieve >90% coverage
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "comprehensive_test.yaml")
	comprehensiveYAML := `
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"
  
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
  default_rotation_size: 104857600
  default_max_duration: 3600
  default_retention_days: 7

storage:
  warn_percent: 80
  block_percent: 90
  default_path: "/tmp/test_recordings"
  fallback_path: "/tmp/test_fallback"

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

	err := os.WriteFile(configPath, []byte(comprehensiveYAML), 0644)
	require.NoError(t, err)

	// Enable hot reload for comprehensive testing
	os.Setenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")
	defer os.Unsetenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD")

	// Set environment variables to test overrides
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_LOGGING_LEVEL")
	}()

	// Load configuration - exercises multiple private methods
	err = env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Add callbacks to test notification system
	callbackCalled := make(chan bool, 1)
	env.ConfigManager.AddUpdateCallback(func(cfg *config.Config) {
		callbackCalled <- true
	})

	// Verify initial configuration
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "192.168.1.100", cfg.Server.Host) // From environment override
	assert.Equal(t, 8002, cfg.Server.Port)
	assert.Equal(t, "DEBUG", cfg.Logging.Level) // From environment override

	// Test configuration update to trigger callbacks
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
  config_path: "/tmp/config.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.05
  detection_timeout: 1.5
  device_range: [0, 4]
  enable_capability_detection: false
  auto_start_streams: false

logging:
  level: "DEBUG"
  format: "json"
  file_enabled: false
  console_enabled: true

recording:
  enabled: true
  format: "mp4"
  quality: "high"

snapshots:
  enabled: true
  format: "jpeg"
  quality: 85

storage:
  warn_percent: 80
  block_percent: 90

retention_policy:
  type: "age"
`

	err = os.WriteFile(configPath, []byte(updatedYAML), 0644)
	require.NoError(t, err)

	// Wait for callback
	select {
	case <-callbackCalled:
		// Callback was triggered successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Callback was not triggered within timeout")
	}

	// Verify updated configuration - environment overrides may still apply
	cfg = env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	// Note: Environment overrides may take precedence over file changes
	// This tests the current implementation behavior
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "10.0.0.2", cfg.MediaMTX.Host)
	assert.Equal(t, 9999, cfg.MediaMTX.APIPort)

	// Test SaveConfig
	err = env.ConfigManager.SaveConfig()
	require.NoError(t, err)

	// Clean up
	env.ConfigManager.Stop()

	// Verify configuration is still accessible after stop
	cfg = env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)
	// Configuration should remain accessible after stop
	assert.NotEmpty(t, cfg.Server.Host)
}
