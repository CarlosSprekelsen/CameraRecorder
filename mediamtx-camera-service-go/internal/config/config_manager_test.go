package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
*/

// ============================================================================
// CONFIGURATION LOADING TESTS
// ============================================================================

func TestConfigManager_LoadConfig_ValidYAML(t *testing.T) {
	// REQ-E1-S1.1-001: Configuration loading from YAML files
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Create config from valid fixture
	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.NoError(t, err, "Should load valid configuration without error")
	assert.NotNil(t, cm.GetConfig(), "Configuration should be loaded")

	config := cm.GetConfig()
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8002, config.Server.Port)
	assert.Equal(t, "/ws", config.Server.WebSocketPath)
}

func TestConfigManager_LoadConfig_InvalidYAML(t *testing.T) {
	// REQ-CONFIG-002: Fail fast on configuration errors
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Create config from invalid fixture
	configPath := helper.CreateTempConfigFromFixture("config_invalid_malformed_yaml.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.Error(t, err, "Should fail to load invalid YAML")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigManager_LoadConfig_MissingFile(t *testing.T) {
	// REQ-CONFIG-002: Fail fast on configuration errors
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	cm := CreateConfigManager()
	err := cm.LoadConfig("/nonexistent/config.yaml")

	require.Error(t, err, "Should fail to load non-existent file")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigManager_LoadConfig_EmptyFile(t *testing.T) {
	// REQ-CONFIG-002: Fail fast on configuration errors
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Create config from empty fixture
	configPath := helper.CreateTempConfigFromFixture("config_invalid_empty.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.Error(t, err, "Should fail to load empty configuration")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

// ============================================================================
// ENVIRONMENT VARIABLE OVERRIDE TESTS
// ============================================================================

func TestConfigManager_LoadConfig_EnvironmentOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Set environment variables
	helper.SetEnvironmentVariable("CAMERA_SERVICE_SERVER_HOST", "192.168.1.100")
	helper.SetEnvironmentVariable("CAMERA_SERVICE_SERVER_PORT", "9090")
	helper.SetEnvironmentVariable("CAMERA_SERVICE_MEDIAMTX_HOST", "mediamtx.example.com")

	// Create config from valid fixture
	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.NoError(t, err, "Should load configuration with environment overrides")

	config := cm.GetConfig()
	assert.Equal(t, "192.168.1.100", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "mediamtx.example.com", config.MediaMTX.Host)
}

// ============================================================================
// CONFIGURATION VALIDATION TESTS
// ============================================================================

func TestConfigManager_LoadConfig_InvalidPort(t *testing.T) {
	// REQ-CONFIG-003: Early detection and clear error reporting
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Create config from invalid port fixture
	configPath := helper.CreateTempConfigFromFixture("config_invalid_invalid_port.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.Error(t, err, "Should fail to load configuration with invalid port")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigManager_LoadConfig_MissingRequiredFields(t *testing.T) {
	// REQ-CONFIG-003: Early detection and clear error reporting
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Create config from missing server fixture
	configPath := helper.CreateTempConfigFromFixture("config_invalid_missing_server.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.Error(t, err, "Should fail to load configuration with missing required fields")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

// ============================================================================
// THREAD SAFETY TESTS
// ============================================================================

func TestConfigManager_ThreadSafeAccess(t *testing.T) {
	// REQ-E1-S1.1-005: Thread-safe configuration access
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")
	cm := CreateConfigManager()

	// Load initial config
	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Concurrent access test
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Read config multiple times
			for j := 0; j < 100; j++ {
				config := cm.GetConfig()
				assert.NotNil(t, config)
				assert.Equal(t, "0.0.0.0", config.Server.Host)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// ============================================================================
// HOT RELOAD TESTS
// ============================================================================

func TestConfigManager_HotReload_Enabled(t *testing.T) {
	// REQ-E1-S1.1-006: Hot reload capability
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Enable hot reload
	helper.SetEnvironmentVariable("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")

	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")
	cm := CreateConfigManager()

	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Verify hot reload is enabled
	assert.True(t, cm.IsHotReloadEnabled(), "Hot reload should be enabled")
}

func TestConfigManager_HotReload_Disabled(t *testing.T) {
	// REQ-E1-S1.1-006: Hot reload capability (disabled by default)
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")
	cm := CreateConfigManager()

	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Verify hot reload is disabled by default
	assert.False(t, cm.IsHotReloadEnabled(), "Hot reload should be disabled by default")
}

// ============================================================================
// CONFIGURATION UPDATE CALLBACK TESTS
// ============================================================================

func TestConfigManager_UpdateCallbacks(t *testing.T) {
	// Test configuration update callbacks
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")
	cm := CreateConfigManager()

	// Track callback calls
	callbackCalled := false
	var callbackConfig *Config

	cm.AddUpdateCallback(func(config *Config) {
		callbackCalled = true
		callbackConfig = config
	})

	// Load config
	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Wait a bit for the async callback to execute
	time.Sleep(100 * time.Millisecond)

	// Verify callback was called
	assert.True(t, callbackCalled, "Update callback should be called")
	assert.NotNil(t, callbackConfig, "Callback should receive configuration")
	assert.Equal(t, "0.0.0.0", callbackConfig.Server.Host)
}

// ============================================================================
// DEFAULT CONFIGURATION TESTS
// ============================================================================

func TestConfigManager_DefaultConfiguration(t *testing.T) {
	// Test default configuration values
	cm := CreateConfigManager()

	defaultConfig := cm.GetDefaultConfig()
	require.NotNil(t, defaultConfig, "Default configuration should be available")

	// Verify some default values
	assert.Equal(t, "0.0.0.0", defaultConfig.Server.Host)
	assert.Equal(t, 8002, defaultConfig.Server.Port)
	assert.Equal(t, "/ws", defaultConfig.Server.WebSocketPath)
}

// ============================================================================
// CONFIGURATION PERSISTENCE TESTS
// ============================================================================

func TestConfigManager_GetConfigPath(t *testing.T) {
	// Test configuration path retrieval
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	configPath := helper.CreateTempConfigFromFixture("config_valid_minimal.yaml")
	cm := CreateConfigManager()

	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	retrievedPath := cm.GetConfigPath()
	assert.Equal(t, configPath, retrievedPath, "Should return correct configuration path")
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestConfigManager_LoadConfig_ValidationErrors(t *testing.T) {
	// REQ-CONFIG-002: Fail fast on configuration errors
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Test with various invalid configurations
	testCases := []struct {
		name        string
		fixtureName string
		expectedErr string
	}{
		{
			name:        "Malformed YAML",
			fixtureName: "config_invalid_malformed_yaml.yaml",
			expectedErr: "configuration validation failed",
		},
		{
			name:        "Invalid Port",
			fixtureName: "config_invalid_invalid_port.yaml",
			expectedErr: "configuration validation failed",
		},
		{
			name:        "Missing Server",
			fixtureName: "config_invalid_missing_server.yaml",
			expectedErr: "configuration validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := helper.CreateTempConfigFromFixture(tc.fixtureName)
			cm := CreateConfigManager()

			err := cm.LoadConfig(configPath)
			require.Error(t, err, "Should fail to load invalid configuration")
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

func TestConfigManager_LoadConfig_Performance(t *testing.T) {
	// Test configuration loading performance
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	configPath := helper.CreateTempConfigFromFixture("config_valid_ultra_efficient.yaml")
	cm := CreateConfigManager()

	// Measure loading time
	start := time.Now()
	err := cm.LoadConfig(configPath)
	loadTime := time.Since(start)

	require.NoError(t, err, "Should load configuration successfully")

	// Performance assertion (should load within reasonable time)
	assert.Less(t, loadTime, 100*time.Millisecond, "Configuration loading should be fast")
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestConfigManager_Integration_CompleteWorkflow(t *testing.T) {
	// Test complete configuration workflow
	helper := NewTestConfigHelper(t)
	helper.CleanupEnvironment()
	defer helper.CleanupEnvironment()

	// Set environment overrides
	helper.SetEnvironmentVariable("CAMERA_SERVICE_SERVER_HOST", "integration.test")
	helper.SetEnvironmentVariable("CAMERA_SERVICE_LOGGING_LEVEL", "debug")

	// Load configuration
	configPath := helper.CreateTempConfigFromFixture("config_valid_complete.yaml")
	cm := CreateConfigManager()

	// Register callback
	callbackCalled := false
	cm.AddUpdateCallback(func(config *Config) {
		callbackCalled = true
	})

	// Load config
	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Wait a bit for the async callback to execute
	time.Sleep(100 * time.Millisecond)

	// Verify complete workflow
	assert.True(t, callbackCalled, "Callback should be called")

	config := cm.GetConfig()
	assert.Equal(t, "integration.test", config.Server.Host, "Environment override should work")
	assert.Equal(t, "debug", config.Logging.Level, "Environment override should work")
	assert.NotNil(t, config.MediaMTX, "MediaMTX config should be loaded")
	assert.NotNil(t, config.Camera, "Camera config should be loaded")
	assert.NotNil(t, config.Logging, "Logging config should be loaded")
}
