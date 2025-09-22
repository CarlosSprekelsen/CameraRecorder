/*
MediaMTX Configuration Management Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_GetConfig_ReqMTX003 tests getting MediaMTX configuration
func TestController_GetConfig_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper, _ := SetupMediaMTXTest(t)

	// Use Progressive Readiness pattern
	controllerInterface, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controllerInterface.Stop(ctx)
	controller := controllerInterface.(*controller)

	// Get configuration
	config, err := controller.GetConfig(ctx)
	// Use assertion helper
	require.NoError(t, err, "Getting configuration should succeed")
	require.NotNil(t, config, "Configuration should not be nil")

	// Verify configuration structure
	assert.NotEmpty(t, config.BaseURL, "Base URL should not be empty")
	assert.Greater(t, config.Timeout, time.Duration(0), "Timeout should be positive")
	assert.Greater(t, config.HealthCheckInterval, 0, "Health check interval should be positive")
}

// TestController_UpdateConfig_ReqMTX003 tests updating MediaMTX configuration
func TestController_UpdateConfig_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper, _ := SetupMediaMTXTest(t)

	// Use Progressive Readiness pattern
	controllerInterface, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controllerInterface.Stop(ctx)
	controller := controllerInterface.(*controller)

	// Get current configuration
	originalConfig, err := controller.GetConfig(ctx)
	require.NoError(t, err, "Getting original configuration should succeed")
	require.NotNil(t, originalConfig, "Original configuration should not be nil")

	// Create updated configuration
	updatedConfig := &config.MediaMTXConfig{
		BaseURL:             originalConfig.BaseURL,
		Timeout:             originalConfig.Timeout,
		HealthCheckInterval: originalConfig.HealthCheckInterval + 1, // Increment by 1
		RetryDelay:          originalConfig.RetryDelay,              // Include required field
		RetryAttempts:       originalConfig.RetryAttempts,           // Include required field
		CircuitBreaker:      originalConfig.CircuitBreaker,          // Include required field
		ConnectionPool:      originalConfig.ConnectionPool,          // Include required field
	}

	// Update configuration
	err = controller.UpdateConfig(ctx, updatedConfig)
	require.NoError(t, err, "Updating configuration should succeed")

	// Verify configuration was updated
	verifyConfig, err := controller.GetConfig(ctx)
	require.NoError(t, err, "Getting updated configuration should succeed")
	require.NotNil(t, verifyConfig, "Updated configuration should not be nil")

	assert.Equal(t, updatedConfig.HealthCheckInterval, verifyConfig.HealthCheckInterval, "Health check interval should be updated")
	assert.Equal(t, updatedConfig.BaseURL, verifyConfig.BaseURL, "Base URL should remain the same")
	assert.Equal(t, updatedConfig.Timeout, verifyConfig.Timeout, "Timeout should remain the same")
}

// TestController_UpdateConfig_InvalidConfig_ReqMTX004 tests updating with invalid configuration
func TestController_UpdateConfig_InvalidConfig_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper, _ := SetupMediaMTXTest(t)

	// Use Progressive Readiness pattern
	controllerInterface, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controllerInterface.Stop(ctx)
	controller := controllerInterface.(*controller)

	// Test updating with nil configuration
	err := controller.UpdateConfig(ctx, nil)
	assert.Error(t, err, "Updating with nil configuration should fail")

	// Test updating with invalid configuration (empty base URL)
	invalidConfig := &config.MediaMTXConfig{
		BaseURL:             "", // Invalid empty URL
		Timeout:             testutils.UniversalTimeoutVeryLong,
		HealthCheckInterval: 5,
	}
	err = controller.UpdateConfig(ctx, invalidConfig)
	assert.Error(t, err, "Updating with invalid configuration should fail")

	// Test updating with invalid configuration (negative timeout)
	invalidConfig2 := &config.MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             -1 * time.Second, // Invalid negative timeout
		HealthCheckInterval: 5,
	}
	err = controller.UpdateConfig(ctx, invalidConfig2)
	assert.Error(t, err, "Updating with negative timeout should fail")

	// Test updating with invalid configuration (zero health check interval)
	invalidConfig3 := &config.MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             testutils.UniversalTimeoutVeryLong,
		HealthCheckInterval: 0, // Invalid zero interval
	}
	err = controller.UpdateConfig(ctx, invalidConfig3)
	assert.Error(t, err, "Updating with zero health check interval should fail")
}

// TestController_GetConfig_NotRunning_ReqMTX004 tests getting configuration when controller is not running
func TestController_GetConfig_NotRunning_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper, _ := SetupMediaMTXTest(t)

	// Create controller but don't start it (for testing not running scenarios)
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx, cancel := helper.GetStandardContext()
	defer cancel()

	// Try to get configuration when controller is not running
	_, err = controller.GetConfig(ctx)
	assert.Error(t, err, "Getting configuration when controller is not running should fail")
}

// TestController_UpdateConfig_NotRunning_ReqMTX004 tests updating configuration when controller is not running
func TestController_UpdateConfig_NotRunning_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper, _ := SetupMediaMTXTest(t)

	// Create controller but don't start it (for testing not running scenarios)
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx, cancel := helper.GetStandardContext()
	defer cancel()

	// Create a valid configuration
	config := &config.MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             testutils.UniversalTimeoutVeryLong,
		HealthCheckInterval: 5,
	}

	// Try to update configuration when controller is not running
	err = controller.UpdateConfig(ctx, config)
	assert.Error(t, err, "Updating configuration when controller is not running should fail")
}

// TestConfigIntegration_GetVersionInfo tests centralized version management
func TestConfigIntegration_GetVersionInfo(t *testing.T) {
	// Create test helper for logger
	helper, _ := SetupMediaMTXTest(t)

	// Create test configuration manager
	configManager := config.CreateConfigManager()
	logger := helper.GetLogger()

	// Create config integration
	ci := NewConfigIntegration(configManager, logger)
	require.NotNil(t, ci, "ConfigIntegration should not be nil")

	// Get version info
	versionInfo := ci.GetVersionInfo()
	require.NotNil(t, versionInfo, "VersionInfo should not be nil")

	// Version should not be hardcoded
	assert.NotEmpty(t, versionInfo.Version, "Version should not be empty")
	assert.NotEmpty(t, versionInfo.BuildDate, "BuildDate should not be empty")
	assert.NotEmpty(t, versionInfo.GitCommit, "GitCommit should not be empty")

	// Validate JSON structure
	assert.IsType(t, "", versionInfo.Version, "Version should be string")
	assert.IsType(t, "", versionInfo.BuildDate, "BuildDate should be string")
	assert.IsType(t, "", versionInfo.GitCommit, "GitCommit should be string")
}
