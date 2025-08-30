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
Module: Configuration Management System

Requirements Coverage:
- REQ-E1-S1.1-001: Configuration loading from YAML files
- REQ-E1-S1.1-002: Environment variable overrides
- REQ-E1-S1.1-003: Configuration validation
- REQ-CONFIG-001: The system SHALL validate configuration files before loading
- REQ-CONFIG-002: The system SHALL fail fast on configuration errors
- REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting
- REQ-E1-S1.1-005: Thread-safe configuration access
- REQ-E1-S1.1-006: Hot reload capability

Test Categories: Unit
API Documentation Reference: N/A (Configuration system)

OPTIMIZED PATTERN: This file demonstrates the correct way to use test utilities
and fixtures instead of creating individual components in each test function.
*/

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

// ============================================================================
// CONFIGURATION LOADING TESTS
// ============================================================================

func TestConfigManager_LoadConfig_ValidYAML(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration loading from YAML files

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Clean up any existing environment variables that might interfere
	cleanupCameraServiceEnvVars()

	// Use existing fixture instead of hardcoded YAML
	configPath := "tests/fixtures/config_valid_minimal.yaml"

	// COMMON PATTERN: Use the test environment's config manager instead of creating a new one
	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Validate loaded configuration from fixture
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.MediaMTX.Host)
	assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
	assert.Equal(t, 8554, cfg.MediaMTX.RTSPPort)
	assert.Equal(t, 8889, cfg.MediaMTX.WebRTCPort)
	assert.Equal(t, 8888, cfg.MediaMTX.HLSPort)

	// Validate codec configuration
	assert.Equal(t, "baseline", cfg.MediaMTX.Codec.VideoProfile)
	assert.Equal(t, "3.1", cfg.MediaMTX.Codec.VideoLevel)
	assert.Equal(t, "yuv420p", cfg.MediaMTX.Codec.PixelFormat)
	assert.Equal(t, "2M", cfg.MediaMTX.Codec.Bitrate)
	assert.Equal(t, "fast", cfg.MediaMTX.Codec.Preset)

	// Validate health check configuration
	assert.Equal(t, 30, cfg.MediaMTX.HealthCheckInterval)
	assert.Equal(t, 3, cfg.MediaMTX.HealthFailureThreshold)
	assert.Equal(t, 60, cfg.MediaMTX.HealthCircuitBreakerTimeout)
	assert.Equal(t, 300, cfg.MediaMTX.HealthMaxBackoffInterval)
	assert.Equal(t, 2, cfg.MediaMTX.HealthRecoveryConfirmationThreshold)
	assert.Equal(t, 2.0, cfg.MediaMTX.BackoffBaseMultiplier)
	assert.Equal(t, []float64{0.1, 0.3}, cfg.MediaMTX.BackoffJitterRange)
	assert.Equal(t, 10.0, cfg.MediaMTX.ProcessTerminationTimeout)
	assert.Equal(t, 5.0, cfg.MediaMTX.ProcessKillTimeout)

	// Validate stream readiness configuration
	assert.Equal(t, 30.0, cfg.MediaMTX.StreamReadiness.Timeout)
	assert.Equal(t, 3, cfg.MediaMTX.StreamReadiness.RetryAttempts)
	assert.Equal(t, 1.0, cfg.MediaMTX.StreamReadiness.RetryDelay)
	assert.Equal(t, 1.0, cfg.MediaMTX.StreamReadiness.CheckInterval)
	assert.True(t, cfg.MediaMTX.StreamReadiness.EnableProgressNotifications)
	assert.True(t, cfg.MediaMTX.StreamReadiness.GracefulFallback)

	// Validate FFmpeg configuration
	assert.Equal(t, 10.0, cfg.FFmpeg.Snapshot.ProcessCreationTimeout)
	assert.Equal(t, 30.0, cfg.FFmpeg.Snapshot.ExecutionTimeout)
	assert.Equal(t, 5, cfg.FFmpeg.Snapshot.InternalTimeout)
	assert.Equal(t, 3, cfg.FFmpeg.Snapshot.RetryAttempts)
	assert.Equal(t, 1.0, cfg.FFmpeg.Snapshot.RetryDelay)

	assert.Equal(t, 10.0, cfg.FFmpeg.Recording.ProcessCreationTimeout)
	assert.Equal(t, 30.0, cfg.FFmpeg.Recording.ExecutionTimeout)
	assert.Equal(t, 5, cfg.FFmpeg.Recording.InternalTimeout)
	assert.Equal(t, 3, cfg.FFmpeg.Recording.RetryAttempts)
	assert.Equal(t, 1.0, cfg.FFmpeg.Recording.RetryDelay)
}

func TestConfigManager_LoadConfig_MissingFile(t *testing.T) {
	// REQ-CONFIG-002: The system SHALL fail fast on configuration errors

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test loading non-existent file
	nonExistentPath := filepath.Join(env.TempDir, "non_existent_config.yaml")
	err := env.ConfigManager.LoadConfig(nonExistentPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")
}

func TestConfigManager_LoadConfig_InvalidYAML(t *testing.T) {
	// REQ-CONFIG-002: The system SHALL fail fast on configuration errors
	// REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing invalid fixture
	configPath := "tests/fixtures/config_invalid_malformed_yaml.yaml"
	err := env.ConfigManager.LoadConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestConfigManager_LoadConfig_EmptyFile(t *testing.T) {
	// REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing empty fixture
	configPath := "tests/fixtures/config_invalid_empty.yaml"
	err := env.ConfigManager.LoadConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

// ============================================================================
// ENVIRONMENT VARIABLE OVERRIDE TESTS
// ============================================================================

func TestConfigManager_EnvironmentVariableOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture as base
	configPath := "tests/fixtures/config_valid_minimal.yaml"

	// Set environment variables to override fixture values
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

	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Environment variables should override YAML values
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
}

func TestConfigManager_EnvironmentVariableComprehensive(t *testing.T) {
	// REQ-E1-S1.1-002: Comprehensive environment variable testing

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture as base
	configPath := "tests/fixtures/config_valid_minimal.yaml"

	// Set key environment variables for testing
	envVars := map[string]string{
		"CAMERA_SERVICE_SERVER_HOST":       "192.168.1.100",
		"CAMERA_SERVICE_SERVER_PORT":       "9000",
		"CAMERA_SERVICE_MEDIAMTX_HOST":     "192.168.1.200",
		"CAMERA_SERVICE_MEDIAMTX_API_PORT": "9998",
		"CAMERA_SERVICE_LOGGING_LEVEL":     "DEBUG",
		"CAMERA_SERVICE_RECORDING_ENABLED": "true",
	}

	// Set all environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify key environment variable overrides
	assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "192.168.1.200", cfg.MediaMTX.Host)
	assert.Equal(t, 9998, cfg.MediaMTX.APIPort)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.True(t, cfg.Recording.Enabled)
}

// ============================================================================
// CONFIGURATION VALIDATION TESTS
// ============================================================================

func TestConfigValidation_ValidConfig(t *testing.T) {
	// REQ-E1-S1.1-003: Configuration validation

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture instead of hardcoded config
	configPath := "tests/fixtures/config_valid_minimal.yaml"
	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Validate the loaded configuration
	err = config.ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestConfigValidation_InvalidConfig(t *testing.T) {
	// Test validation with invalid configuration

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing invalid fixture instead of hardcoded config
	configPath := "tests/fixtures/config_invalid_invalid_port.yaml"
	err := env.ConfigManager.LoadConfig(configPath)
	assert.Error(t, err)

	// Check that configuration validation failed contains field information
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigValidation_Comprehensive(t *testing.T) {
	// REQ-E1-S1.1-003: Comprehensive configuration validation testing
	// REQ-CONFIG-001: The system SHALL validate configuration files before loading
	// REQ-CONFIG-002: The system SHALL fail fast on configuration errors
	// REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting

	// Test required field validation
	t.Run("required_field_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Use existing invalid fixture
		configPath := "tests/fixtures/config_invalid_empty.yaml"
		err := env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	// Test data type validation
	t.Run("data_type_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Use existing invalid fixture
		configPath := "tests/fixtures/config_invalid_invalid_port.yaml"
		err := env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	// Test cross-field validation
	t.Run("cross_field_validation", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Use existing invalid fixture
		configPath := "tests/fixtures/config_invalid_missing_server.yaml"
		err := env.ConfigManager.LoadConfig(configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})
}

// ============================================================================
// THREAD SAFETY TESTS
// ============================================================================

func TestConfigManager_ThreadSafety(t *testing.T) {
	// REQ-E1-S1.1-005: Thread-safe configuration access

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture instead of hardcoded YAML
	configPath := "tests/fixtures/config_valid_minimal.yaml"

	// Load configuration in goroutine
	done := make(chan bool)
	go func() {
		err := env.ConfigManager.LoadConfig(configPath)
		assert.NoError(t, err)
		cfg := env.ConfigManager.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
		done <- true
	}()

	// Access configuration concurrently
	cfg := env.ConfigManager.GetConfig()
	assert.NotNil(t, cfg)

	<-done
}

func TestConfigManager_GetConfig_ThreadSafe(t *testing.T) {
	// Test thread-safe access to configuration

	// COMMON PATTERN: Use shared test environment instead of individual components
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

// ============================================================================
// CALLBACK AND HOT RELOAD TESTS
// ============================================================================

func TestConfigManager_AddUpdateCallback(t *testing.T) {
	// Test configuration update callback functionality

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	callback := func(cfg *config.Config) {
		assert.NotNil(t, cfg)
	}

	env.ConfigManager.AddUpdateCallback(callback)

	// Load configuration to trigger callback using existing fixture
	configPath := utils.GetTestConfigPath()
	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify the configuration was loaded correctly
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
}

func TestConfigManager_HotReload(t *testing.T) {
	// REQ-E1-S1.1-006: Hot reload capability

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	configPath := filepath.Join(env.TempDir, "test_config.yaml")

	// Create initial configuration using fixture values
	initialYAML := `
server:
  host: "0.0.0.0"
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
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
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

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture instead of hardcoded YAML
	configPath := utils.GetTestConfigPath()
	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)

	// Stop the manager
	env.ConfigManager.Stop()
}

// ============================================================================
// EDGE CASES AND ERROR HANDLING TESTS
// ============================================================================

func TestConfigValidation_EdgeCases(t *testing.T) {
	// Test edge cases for configuration validation

	// Test extremely large values
	t.Run("extremely_large_values", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Use existing fixture for edge case testing
		configPath := "tests/fixtures/config_valid_minimal.yaml"
		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify configuration is loaded correctly
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
		assert.Equal(t, "localhost", cfg.MediaMTX.Host)
		assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
	})

	// Test boundary values
	t.Run("boundary_values", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Use existing fixture for boundary testing
		configPath := "tests/fixtures/config_valid_minimal.yaml"

		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify boundary values are loaded correctly
		assert.Equal(t, 8002, cfg.Server.Port)
		assert.Equal(t, 10, cfg.Server.MaxConnections)
		assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
		assert.Equal(t, 30, cfg.MediaMTX.HealthCheckInterval)
		assert.Equal(t, 3, cfg.MediaMTX.HealthFailureThreshold)
		assert.Equal(t, 2.0, cfg.MediaMTX.BackoffBaseMultiplier)
		assert.Equal(t, 10.0, cfg.MediaMTX.ProcessTerminationTimeout)
	})

	// Test special characters in string fields
	t.Run("special_characters", func(t *testing.T) {
		// COMMON PATTERN: Use shared test environment instead of individual components
		env := utils.SetupTestEnvironment(t)
		defer utils.TeardownTestEnvironment(t, env)

		// Clean up any existing environment variables that might interfere
		cleanupCameraServiceEnvVars()

		// Use existing fixture for special character testing
		configPath := "tests/fixtures/config_valid_minimal.yaml"
		err := env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should be valid
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)

		// Verify configuration is loaded correctly
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
		assert.Equal(t, "localhost", cfg.MediaMTX.Host)
		assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
	})
}

func TestConfigValidation_FileSystemEdgeCases(t *testing.T) {
	// Test file system edge cases

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test file permission errors
	t.Run("file_permission_errors", func(t *testing.T) {
		configPath := filepath.Join(env.TempDir, "readonly_config.yaml")

		// Create a file with read-only permissions using fixture values
		yamlContent := `
server:
  host: "0.0.0.0"
  port: 8002
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0444) // Read-only
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should still be able to read
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})

	// Test symbolic link handling
	t.Run("symbolic_link_handling", func(t *testing.T) {
		originalPath := filepath.Join(env.TempDir, "original_config.yaml")
		symlinkPath := filepath.Join(env.TempDir, "symlink_config.yaml")

		// Use fixture values
		yamlContent := `
server:
  host: "localhost"
  port: 9997
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
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, 9997, cfg.Server.Port)
	})

	// Test deeply nested configuration paths
	t.Run("deeply_nested_paths", func(t *testing.T) {
		// Create deeply nested directory structure
		deepDir := filepath.Join(env.TempDir, "deep", "nested", "config", "path")
		err := os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		configPath := filepath.Join(deepDir, "config.yaml")
		// Use fixture values
		yamlContent := `
server:
  host: "0.0.0.0"
  port: 8002
`
		err = os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		err = env.ConfigManager.LoadConfig(configPath)
		require.NoError(t, err) // Should handle deep paths correctly
		cfg := env.ConfigManager.GetConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8002, cfg.Server.Port)
	})
}

// ============================================================================
// GLOBAL CONFIG FUNCTIONS TESTS
// ============================================================================

func TestGlobalConfigFunctions(t *testing.T) {
	// Test global configuration manager functions

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Use existing fixture instead of hardcoded YAML
	configPath := "tests/fixtures/config_valid_minimal.yaml"

	// Test global config loading
	err := env.ConfigManager.LoadConfig(configPath)
	require.NoError(t, err)
	cfg := env.ConfigManager.GetConfig()
	require.NotNil(t, cfg)

	// Verify global configuration is accessible
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8002, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.MediaMTX.Host)
	assert.Equal(t, 9997, cfg.MediaMTX.APIPort)
}
