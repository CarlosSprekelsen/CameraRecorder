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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_GetConfig_ReqMTX003 tests getting MediaMTX configuration
func TestController_GetConfig_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get configuration
	config, err := controller.GetConfig(ctx)
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
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get current configuration
	originalConfig, err := controller.GetConfig(ctx)
	require.NoError(t, err, "Getting original configuration should succeed")
	require.NotNil(t, originalConfig, "Original configuration should not be nil")

	// Create updated configuration
	updatedConfig := &MediaMTXConfig{
		BaseURL:             originalConfig.BaseURL,
		Timeout:             originalConfig.Timeout,
		HealthCheckInterval: originalConfig.HealthCheckInterval + 1, // Increment by 1
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
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test updating with nil configuration
	err = controller.UpdateConfig(ctx, nil)
	assert.Error(t, err, "Updating with nil configuration should fail")

	// Test updating with invalid configuration (empty base URL)
	invalidConfig := &MediaMTXConfig{
		BaseURL:             "", // Invalid empty URL
		Timeout:             5 * time.Second,
		HealthCheckInterval: 5,
	}
	err = controller.UpdateConfig(ctx, invalidConfig)
	assert.Error(t, err, "Updating with invalid configuration should fail")

	// Test updating with invalid configuration (negative timeout)
	invalidConfig2 := &MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             -1 * time.Second, // Invalid negative timeout
		HealthCheckInterval: 5,
	}
	err = controller.UpdateConfig(ctx, invalidConfig2)
	assert.Error(t, err, "Updating with negative timeout should fail")

	// Test updating with invalid configuration (zero health check interval)
	invalidConfig3 := &MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             5 * time.Second,
		HealthCheckInterval: 0, // Invalid zero interval
	}
	err = controller.UpdateConfig(ctx, invalidConfig3)
	assert.Error(t, err, "Updating with zero health check interval should fail")
}

// TestController_GetConfig_NotRunning_ReqMTX004 tests getting configuration when controller is not running
func TestController_GetConfig_NotRunning_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller but don't start it
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()

	// Try to get configuration when controller is not running
	_, err = controller.GetConfig(ctx)
	assert.Error(t, err, "Getting configuration when controller is not running should fail")
}

// TestController_UpdateConfig_NotRunning_ReqMTX004 tests updating configuration when controller is not running
func TestController_UpdateConfig_NotRunning_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller but don't start it
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()

	// Create a valid configuration
	config := &MediaMTXConfig{
		BaseURL:             "http://localhost:9997",
		Timeout:             5 * time.Second,
		HealthCheckInterval: 5,
	}

	// Try to update configuration when controller is not running
	err = controller.UpdateConfig(ctx, config)
	assert.Error(t, err, "Updating configuration when controller is not running should fail")
}
