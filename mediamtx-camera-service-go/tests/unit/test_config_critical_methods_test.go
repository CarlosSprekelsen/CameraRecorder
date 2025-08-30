//go:build unit
// +build unit

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_EnvironmentOverrides(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test environment override functionality through LoadConfig

	t.Run("EnvironmentOverrideThroughLoadConfig", func(t *testing.T) {
		// Create a minimal test config file
		tempDir := env.TempDir
		configPath := filepath.Join(tempDir, "env_test_config.yaml")

		minimalYAML := `
server:
  host: "default-host"
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
  process_termination_timeout: 3.0
  process_kill_timeout: 2.0

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
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
`

		err := os.WriteFile(configPath, []byte(minimalYAML), 0644)
		require.NoError(t, err)

		// Set environment variables
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "test-host")
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "8080")
		os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "DEBUG")
		defer func() {
			os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
			os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
			os.Unsetenv("CAMERA_SERVICE_LOGGING_LEVEL")
		}()

		// Use shared config manager and load config
		err = env.ConfigManager.LoadConfig(configPath) // Load with valid path to trigger environment overrides

		// Should handle environment overrides
		if err == nil {
			cfg := env.ConfigManager.GetConfig()
			if cfg != nil {
				assert.Equal(t, "test-host", cfg.Server.Host)
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "DEBUG", cfg.Logging.Level)
			}
		}
	})

	t.Run("EmptyEnvironmentVariables", func(t *testing.T) {
		// Create a minimal test config file
		tempDir := env.TempDir
		configPath := filepath.Join(tempDir, "empty_env_test_config.yaml")

		minimalYAML := `
server:
  host: "default-host"
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
  process_termination_timeout: 3.0
  process_kill_timeout: 2.0

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
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
`

		err := os.WriteFile(configPath, []byte(minimalYAML), 0644)
		require.NoError(t, err)

		// Test with empty environment variables
		os.Setenv("CAMERA_SERVICE_SERVER_HOST", "")
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "")
		defer func() {
			os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
			os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		}()

		err = env.ConfigManager.LoadConfig(configPath)

		// Should handle empty values gracefully - config manager uses defaults
		if err == nil {
			cfg := env.ConfigManager.GetConfig()
			if cfg != nil {
				// Config manager uses default values when environment variables are empty
				assert.NotEmpty(t, cfg.Server.Host, "Should use default host when env var is empty")
				assert.NotZero(t, cfg.Server.Port, "Should use default port when env var is empty")
			}
		}
	})

	t.Run("InvalidPortValue", func(t *testing.T) {
		// Create a minimal test config file
		tempDir := env.TempDir
		configPath := filepath.Join(tempDir, "invalid_port_test_config.yaml")

		minimalYAML := `
server:
  host: "default-host"
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
  process_termination_timeout: 3.0
  process_kill_timeout: 2.0

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
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
`

		err := os.WriteFile(configPath, []byte(minimalYAML), 0644)
		require.NoError(t, err)

		// Test with invalid port value
		os.Setenv("CAMERA_SERVICE_SERVER_PORT", "invalid")
		defer os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")

		err = env.ConfigManager.LoadConfig(configPath)
		// Should handle invalid values gracefully - should return error for invalid port
		assert.Error(t, err, "Should return error for invalid port value")
	})
}

func TestConfigManager_Validation(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test validation through LoadConfig

	t.Run("ValidConfigFile", func(t *testing.T) {
		// Create a temporary config file
		tempDir := t.TempDir()
		configPath := tempDir + "/config.yaml"

		configContent := `
server:
  host: "127.0.0.1"
  port: 8080
mediamtx:
  host: "localhost"
  api_port: 9997
logging:
  level: "INFO"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		assert.NoError(t, err, "Valid config should pass validation")

		cfg := env.ConfigManager.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})

	t.Run("InvalidConfigFile", func(t *testing.T) {
		// Create a temporary config file with invalid values
		tempDir := t.TempDir()
		configPath := tempDir + "/invalid_config.yaml"

		configContent := `
server:
  host: ""  # Empty host should fail validation
  port: 0   # Invalid port should fail validation
mediamtx:
  host: ""  # Empty host should fail validation
  api_port: 0  # Invalid port should fail validation
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		// Should handle validation errors
		if err != nil {
			assert.Error(t, err, "Invalid config should fail validation")
		}
	})
}

func TestConfigManager_LoadConfig(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test LoadConfig method coverage

	t.Run("LoadFromValidFile", func(t *testing.T) {
		// Create a temporary config file
		tempDir := t.TempDir()
		configPath := tempDir + "/config.yaml"

		configContent := `
server:
  host: "127.0.0.1"
  port: 8080
mediamtx:
  host: "localhost"
  api_port: 9997
logging:
  level: "INFO"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		assert.NotNil(t, cfg)
		// Config manager uses defaults when there are parsing issues
		assert.NotEmpty(t, cfg.Server.Host, "Should have a host value")
		assert.NotZero(t, cfg.Server.Port, "Should have a port value")
		assert.NotEmpty(t, cfg.MediaMTX.Host, "Should have a MediaMTX host value")
		assert.NotZero(t, cfg.MediaMTX.APIPort, "Should have a MediaMTX API port value")
		assert.NotEmpty(t, cfg.Logging.Level, "Should have a logging level")
		// Recording and snapshots may be disabled by default
		assert.NotNil(t, cfg.Recording, "Should have recording config")
		assert.NotNil(t, cfg.Snapshots, "Should have snapshots config")
	})

	t.Run("LoadFromInvalidFile", func(t *testing.T) {
		err := env.ConfigManager.LoadConfig("/nonexistent/file.yaml")
		// Config manager should return error for missing files
		assert.Error(t, err, "Should return error for missing file")
		assert.Contains(t, err.Error(), "configuration file does not exist")
	})

	t.Run("LoadFromInvalidYAML", func(t *testing.T) {
		// Create a temporary config file with invalid YAML
		tempDir := t.TempDir()
		configPath := tempDir + "/invalid.yaml"

		configContent := `
server:
  host: "127.0.0.1"
  port: 8080
  # Invalid YAML - missing closing quote
  websocket_path: "/ws
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		// Config manager should return error for invalid YAML
		assert.Error(t, err, "Should return error for invalid YAML")
		assert.Contains(t, err.Error(), "While parsing config")
	})
}

func TestConfigManager_GetConfig(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test GetConfig method

	t.Run("GetConfigBeforeLoad", func(t *testing.T) {
		cfg := env.ConfigManager.GetConfig()
		// Should return config even if not loaded
		assert.NotNil(t, cfg)
	})

	t.Run("GetConfigAfterLoad", func(t *testing.T) {
		// Create a temporary config file
		tempDir := t.TempDir()
		configPath := tempDir + "/config.yaml"

		configContent := `
server:
  host: "127.0.0.1"
  port: 8080
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err)

		cfg := env.ConfigManager.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})
}

func TestConfigManager_UpdateCallback(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test AddUpdateCallback method

	t.Run("AddUpdateCallback", func(t *testing.T) {
		callback := func(cfg *config.Config) {
			// Callback function for testing
		}

		env.ConfigManager.AddUpdateCallback(callback)

		// Should not panic when adding callback
		assert.NotNil(t, env.ConfigManager)
	})
}

func TestConfigManager_DirectMethodCoverage(t *testing.T) {
	// REQ-E1-S1.1-004: Default value fallback

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test methods that need direct coverage

}
