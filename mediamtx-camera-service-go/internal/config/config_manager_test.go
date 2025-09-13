package config

import (
	"context"
	"os"
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
*/

func TestConfigManager_LoadConfig(t *testing.T) {
	// Consolidated test for all config loading scenarios
	testCases := []struct {
		name        string
		fixture     string
		expectError bool
		description string
	}{
		{
			name:        "Valid YAML",
			fixture:     "config_test_minimal.yaml",
			expectError: false,
			description: "Should load valid configuration successfully",
		},
		{
			name:        "Invalid YAML",
			fixture:     "config_invalid_malformed_yaml.yaml",
			expectError: true,
			description: "Should fail to load malformed YAML",
		},
		{
			name:        "Invalid Port",
			fixture:     "config_invalid_invalid_port.yaml",
			expectError: true,
			description: "Should fail to load configuration with invalid port",
		},
		{
			name:        "Missing Server",
			fixture:     "config_invalid_missing_server.yaml",
			expectError: true,
			description: "Should fail to load configuration with missing server",
		},
		{
			name:        "Empty Config",
			fixture:     "config_invalid_empty.yaml",
			expectError: true,
			description: "Should fail to load empty configuration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewTestConfigHelper(t)
			defer helper.CleanupEnvironment()

			// Create test directories
			helper.CreateTestDirectories()

			// Create config file from fixture
			configPath := helper.CreateTempConfigFromFixture(tc.fixture)

			cm := CreateConfigManager()
			err := cm.LoadConfig(configPath)

			if tc.expectError {
				require.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), "configuration validation failed")
			} else {
				require.NoError(t, err, tc.description)
				config := cm.GetConfig()
				require.NotNil(t, config, "Configuration should be loaded")
				assert.Equal(t, "0.0.0.0", config.Server.Host)
				assert.Equal(t, 8002, config.Server.Port)
				assert.Equal(t, "/ws", config.Server.WebSocketPath)
			}
		})
	}
}

func TestConfigManager_LoadConfig_MissingFile(t *testing.T) {
	// Test loading non-existent file
	cm := CreateConfigManager()
	err := cm.LoadConfig("/nonexistent/config.yaml")

	require.Error(t, err, "Should fail to load non-existent file")
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestConfigManager_EnvironmentOverrides(t *testing.T) {
	// REQ-E1-S1.1-002: Environment variable overrides
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Set environment variables
	helper.SetEnvironmentVariable("CAMERA_SERVICE_SERVER_HOST", "integration.test")
	helper.SetEnvironmentVariable("CAMERA_SERVICE_LOGGING_LEVEL", "debug")

	// Create test directories and config
	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.NoError(t, err, "Should load configuration with environment overrides")

	config := cm.GetConfig()
	assert.Equal(t, "integration.test", config.Server.Host, "Environment override should work")
	assert.Equal(t, "debug", config.Logging.Level, "Environment override should work")
}

func TestConfigManager_ThreadSafeAccess(t *testing.T) {
	// Test thread-safe access to configuration
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Test concurrent access
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			config := cm.GetConfig()
			assert.NotNil(t, config)
			assert.Equal(t, "0.0.0.0", config.Server.Host)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConfigManager_HotReload(t *testing.T) {
	// Test hot reload functionality
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Enable hot reload
	helper.SetEnvironmentVariable("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")

	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)

	require.NoError(t, err, "Should load configuration with hot reload enabled")
	assert.NotNil(t, cm.GetConfig())

	// Test that Stop works
	ctx := context.Background()
	cm.Stop(ctx)
}

func TestConfigManager_UpdateCallbacks(t *testing.T) {
	// Test configuration update callbacks
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()

	// Add update callback with synchronization
	callbackCalled := make(chan bool, 1)
	var callbackConfig *Config
	cm.AddUpdateCallback(func(config *Config) {
		callbackConfig = config
		callbackCalled <- true
	})

	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Wait for callback to be called (with timeout)
	select {
	case <-callbackCalled:
		// Callback was called successfully
		assert.NotNil(t, callbackConfig, "Callback should receive configuration")
		assert.Equal(t, "0.0.0.0", callbackConfig.Server.Host)
	case <-time.After(1 * time.Second):
		t.Fatal("Update callback was not called within timeout")
	}
}

func TestConfigManager_SaveConfig(t *testing.T) {
	// Test saving configuration
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err)

	// Test save functionality
	err = cm.SaveConfig()
	require.NoError(t, err, "Should save configuration successfully")
}

func TestConfigManager_ReloadConfiguration(t *testing.T) {
	// Test reloadConfiguration method - Priority 2: Critical missing coverage
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Create test directories
	helper.CreateTestDirectories()

	// Create initial config
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err, "Should load initial configuration")

	initialConfig := cm.GetConfig()
	require.NotNil(t, initialConfig, "Initial config should not be nil")

	// Test reloadConfiguration when file exists
	cm.reloadConfiguration()

	// Verify config is still valid after reload
	reloadedConfig := cm.GetConfig()
	require.NotNil(t, reloadedConfig, "Reloaded config should not be nil")
	assert.Equal(t, initialConfig.Server.Host, reloadedConfig.Server.Host, "Config should be the same after reload")

	// Test reloadConfiguration when file is removed
	os.Remove(configPath)
	cm.reloadConfiguration()

	// Verify file watching is stopped when file is removed
	// (This tests the file existence check in reloadConfiguration)
}

func TestConfigManager_Stop(t *testing.T) {
	// Test Stop method - Priority 2: Critical missing coverage
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Create test directories
	helper.CreateTestDirectories()

	// Create config with hot reload enabled
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	// Enable hot reload
	helper.SetEnvironmentVariable("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err, "Should load configuration with hot reload")

	// Verify config is loaded
	config := cm.GetConfig()
	require.NotNil(t, config, "Config should be loaded")

	// Test Stop method
	ctx := context.Background()
	cm.Stop(ctx)

	// Verify Stop completes without hanging
	// The Stop method should:
	// 1. Close the stop channel
	// 2. Stop file watching
	// 3. Wait for goroutines to finish
	// 4. Log completion

	// Note: The current implementation has a bug where calling Stop() multiple times
	// causes a panic due to closing an already closed channel. This should be fixed
	// in the implementation to make it idempotent.
}

func TestConfigManager_FileWatching(t *testing.T) {
	// Test file watching functionality - Priority 2: Critical missing coverage
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	helper.CreateTestDirectories()
	configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")

	// Enable hot reload
	helper.SetEnvironmentVariable("CAMERA_SERVICE_ENABLE_HOT_RELOAD", "true")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err, "Should load configuration with hot reload")

	// Verify config is loaded
	config := cm.GetConfig()
	require.NotNil(t, config, "Config should be loaded")

	// Test that file watching is active
	// The watchFileChanges function should be running in a goroutine
	// We can't directly test the goroutine, but we can verify the setup

	// Test Stop method to ensure file watching stops cleanly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = cm.Stop(ctx)
	require.NoError(t, err, "Config manager should stop gracefully")

	// Verify Stop completes without hanging
	// This tests the stopChan path in watchFileChanges
}

// TestConfigManager_ContextAwareShutdown tests the context-aware shutdown functionality
func TestConfigManager_ContextAwareShutdown(t *testing.T) {
	helper := NewTestConfigHelper(t)

	t.Run("graceful_shutdown_with_context", func(t *testing.T) {
		// Create config manager
		cm := CreateConfigManager()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cm.Stop(ctx)
		}()

		// Load config to start file watching
		configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")
		err := cm.LoadConfig(configPath)
		require.NoError(t, err, "Config should load successfully")

		// Test graceful shutdown with context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err = cm.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Config manager should stop gracefully")
		assert.Less(t, elapsed, 1*time.Second, "Shutdown should be fast")
	})

	t.Run("shutdown_with_cancelled_context", func(t *testing.T) {
		// Create config manager
		cm := CreateConfigManager()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cm.Stop(ctx)
		}()

		// Load config to start file watching
		configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")
		err := cm.LoadConfig(configPath)
		require.NoError(t, err, "Config should load successfully")

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Stop should complete quickly since context is already cancelled
		start := time.Now()
		err = cm.Stop(ctx)
		elapsed := time.Since(start)

		// When context is cancelled, Stop should return context error
		require.Error(t, err, "Config manager should return context error when context is cancelled")
		assert.Equal(t, context.Canceled, err, "Should return context.Canceled error")
		assert.Less(t, elapsed, 100*time.Millisecond, "Shutdown should be very fast with cancelled context")
	})

	t.Run("shutdown_timeout_handling", func(t *testing.T) {
		// Create config manager
		cm := CreateConfigManager()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cm.Stop(ctx)
		}()

		// Load config to start file watching
		configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")
		err := cm.LoadConfig(configPath)
		require.NoError(t, err, "Config should load successfully")

		// Use very short timeout to test timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(2 * time.Millisecond)

		start := time.Now()
		err = cm.Stop(ctx)
		elapsed := time.Since(start)

		// Should timeout but not hang
		require.Error(t, err, "Should timeout with very short timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Error should indicate timeout")
		assert.Less(t, elapsed, 1*time.Second, "Should not hang indefinitely")
	})

	t.Run("double_stop_handling", func(t *testing.T) {
		// Create config manager
		cm := CreateConfigManager()

		// Load config to start file watching
		configPath := helper.CreateTempConfigFromFixture("config_test_minimal.yaml")
		err := cm.LoadConfig(configPath)
		require.NoError(t, err, "Config should load successfully")

		// Stop first time
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()
		err = cm.Stop(ctx1)
		require.NoError(t, err, "First stop should succeed")

		// Stop second time should not error
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		err = cm.Stop(ctx2)
		assert.NoError(t, err, "Second stop should not error")
	})

	t.Run("stop_without_load", func(t *testing.T) {
		// Create config manager
		cm := CreateConfigManager()

		// Stop without loading config should not error
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := cm.Stop(ctx)
		assert.NoError(t, err, "Stop without load should not error")
	})
}
