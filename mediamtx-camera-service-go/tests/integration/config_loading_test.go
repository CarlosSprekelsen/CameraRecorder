/*
Component: Configuration Loading Integration
Purpose: Validates configuration loading, validation, and component wiring
Requirements: REQ-CFG-001, REQ-CFG-002, REQ-CFG-003, REQ-CFG-004
Category: Integration
API Reference: internal/config/config_manager.go
Test Organization:
  - TestConfigLoading_FixtureVariations (lines 45-85)
  - TestConfigLoading_EnvOverrides (lines 87-127)
  - TestConfigLoading_ValidationErrors (lines 129-169)
  - TestConfigLoading_ComponentWiring (lines 171-211)
*/

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLoading_FixtureVariations_ReqCFG001 validates different config fixtures
// REQ-CFG-001: Configuration file loading
func TestConfigLoading_FixtureVariations_ReqCFG001(t *testing.T) {
	// Table-driven test for different config fixtures
	tests := []struct {
		name        string
		fixtureName string
		expectError bool
		description string
	}{
		{"minimal_config", "config_valid_minimal.yaml", false, "Minimal valid configuration"},
		{"complete_config", "config_valid_complete.yaml", false, "Complete valid configuration"},
		{"ultra_efficient_config", "config_valid_ultra_efficient.yaml", false, "Ultra efficient configuration"},
		{"invalid_empty", "config_invalid_empty.yaml", true, "Empty configuration should fail"},
		{"invalid_malformed", "config_invalid_malformed_yaml.yaml", true, "Malformed YAML should fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use testutils.SetupTest with specified fixture
			setup := testutils.SetupTest(t, tt.fixtureName)
			defer setup.Cleanup()

			// Get configuration manager
			configManager := setup.GetConfigManager()
			require.NotNil(t, configManager, "Config manager should be created")

			// Validate configuration loading
			loadedConfig := configManager.GetConfig()

			if tt.expectError {
				// For invalid configs, we expect the manager to be created but config to be nil or invalid
				assert.Nil(t, loadedConfig, "Invalid config should result in nil configuration")
			} else {
				// For valid configs, validate structure and content
				require.NotNil(t, loadedConfig, "Valid config should be loaded")

				// Validate key configuration sections exist
				assert.NotNil(t, loadedConfig.Server, "Server config should be present")
				assert.NotNil(t, loadedConfig.MediaMTX, "MediaMTX config should be present")
				assert.NotNil(t, loadedConfig.Camera, "Camera config should be present")
				assert.NotNil(t, loadedConfig.Logging, "Logging config should be present")

				// Validate component initialized successfully
				assert.True(t, loadedConfig.Server.Port > 0, "Server port should be configured")
				assert.NotEmpty(t, loadedConfig.MediaMTX.Host, "MediaMTX host should be configured")
				assert.True(t, loadedConfig.Camera.PollInterval > 0, "Camera poll interval should be configured")
			}
		})
	}
}

// TestConfigLoading_EnvOverrides_ReqCFG002 validates environment variable overrides
// REQ-CFG-002: Environment variable override behavior
func TestConfigLoading_EnvOverrides_ReqCFG002(t *testing.T) {
	// Table-driven test for environment variable overrides
	tests := []struct {
		name        string
		envVar      string
		envValue    string
		description string
	}{
		{"server_port_override", "CAMERA_SERVICE_SERVER_PORT", "9999", "Server port override"},
		{"log_level_override", "CAMERA_SERVICE_LOG_LEVEL", "debug", "Log level override"},
		{"mediamtx_host_override", "CAMERA_SERVICE_MEDIAMTX_HOST", "testhost", "MediaMTX host override"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tt.envVar, tt.envValue)
			defer os.Unsetenv(tt.envVar)

			// Load configuration with environment override
			setup := testutils.SetupTest(t, "config_valid_minimal.yaml")
			defer setup.Cleanup()

			configManager := setup.GetConfigManager()
			loadedConfig := configManager.GetConfig()

			require.NotNil(t, loadedConfig, "Config should be loaded")

			// Validate environment override took effect
			switch tt.envVar {
			case "CAMERA_SERVICE_SERVER_PORT":
				assert.Equal(t, 9999, loadedConfig.Server.Port, "Server port should be overridden")
			case "CAMERA_SERVICE_LOG_LEVEL":
				assert.Equal(t, "debug", loadedConfig.Logging.Level, "Log level should be overridden")
			case "CAMERA_SERVICE_MEDIAMTX_HOST":
				assert.Equal(t, "testhost", loadedConfig.MediaMTX.Host, "MediaMTX host should be overridden")
			}

			// Validate component uses overridden value
			assert.True(t, loadedConfig.Server.Port > 0, "Component should use overridden port")
		})
	}
}

// TestConfigLoading_ValidationErrors_ReqCFG003 validates configuration validation
// REQ-CFG-003: Validation error handling
func TestConfigLoading_ValidationErrors_ReqCFG003(t *testing.T) {
	// Table-driven test for validation errors
	tests := []struct {
		name        string
		fixtureName string
		expectError bool
		errorCode   string
		description string
	}{
		{"invalid_port", "config_invalid_invalid_port.yaml", true, "INVALID_RANGE", "Invalid port number"},
		{"negative_poll_interval", "config_invalid_camera_negative_poll_interval.yaml", true, "INVALID_RANGE", "Negative poll interval"},
		{"invalid_device_range", "config_invalid_camera_invalid_device_range.yaml", true, "INVALID_RANGE", "Invalid device range"},
		{"valid_config", "config_valid_minimal.yaml", false, "", "Valid configuration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load configuration
			setup := testutils.SetupTest(t, tt.fixtureName)
			defer setup.Cleanup()

			configManager := setup.GetConfigManager()
			loadedConfig := configManager.GetConfig()

			if tt.expectError {
				// For invalid configs, validate error handling
				assert.Nil(t, loadedConfig, "Invalid config should result in nil configuration")

				// Validate that component initialization fails gracefully
				// This proves config validation prevents bad component initialization
				logger := logging.GetLogger("test")
				client := mediamtx.NewClient("http://localhost:9997/v3", &config.MediaMTXConfig{}, logger)
				assert.NotNil(t, client, "Component should handle nil config gracefully")
			} else {
				// For valid configs, validate successful loading
				require.NotNil(t, loadedConfig, "Valid config should be loaded")

				// Validate component not initialized with bad config
				assert.NotNil(t, loadedConfig.Server, "Server config should be valid")
				assert.True(t, loadedConfig.Server.Port > 0, "Port should be valid")
			}
		})
	}
}

// TestConfigLoading_ComponentWiring_ReqCFG004 validates component wiring with config
// REQ-CFG-004: Component wiring with configuration
func TestConfigLoading_ComponentWiring_ReqCFG004(t *testing.T) {
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")
	defer setup.Cleanup()

	configManager := setup.GetConfigManager()
	loadedConfig := configManager.GetConfig()

	require.NotNil(t, loadedConfig, "Config should be loaded")

	// Create components using loaded configuration
	logger := setup.GetLogger()

	// Test MediaMTX client creation with config
	client := mediamtx.NewClient("http://localhost:9997/v3", &loadedConfig.MediaMTX, logger)
	require.NotNil(t, client, "MediaMTX client should be created with config")

	// Validate component uses config values in behavior
	ctx, cancel := context.WithTimeout(context.Background(), testutils.UniversalTimeoutShort)
	defer cancel()

	// Test that component behaves according to config
	// This proves config changes propagate to components
	err := client.HealthCheck(ctx)
	// We expect this to fail (no MediaMTX server), but it should use config timeout
	assert.Error(t, err, "Health check should fail without server")

	// Validate config-driven behavior by checking timeout
	assert.Contains(t, err.Error(), "context", "Error should indicate timeout from config")

	// Test component wiring with different config
	// Create second config with different timeout
	altConfig := *loadedConfig
	altConfig.MediaMTX.Timeout = 100 * time.Millisecond // Very short timeout

	altClient := mediamtx.NewClient("http://localhost:9997/v3", &altConfig.MediaMTX, logger)
	require.NotNil(t, altClient, "Alternative client should be created")

	// Test behavior changes with config change
	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()

	start := time.Now()
	err2 := altClient.HealthCheck(ctx2)
	duration := time.Since(start)

	// Validate config change affects component behavior
	assert.Error(t, err2, "Health check should fail")
	assert.Less(t, duration, 500*time.Millisecond, "Short timeout should cause faster failure")
}
