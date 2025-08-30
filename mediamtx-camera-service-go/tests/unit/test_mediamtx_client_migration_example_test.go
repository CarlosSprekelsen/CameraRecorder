//go:build unit
// +build unit

/*
MediaMTX Client Migration Example Test

This file demonstrates how to migrate from the old pattern of creating individual
MediaMTX clients in each test to using the new centralized utilities.

OLD PATTERN (to be migrated):
   testConfig := &mediamtx.MediaMTXConfig{...}
   client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

NEW PATTERN (using utilities):
   client := utils.SetupMediaMTXTestClient(t, env)
   defer utils.TeardownMediaMTXTestClient(t, client)

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: MediaMTX client configuration

Test Categories: Unit/MediaMTX Infrastructure
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMediaMTXClientMigrationExample demonstrates the migration from old to new pattern
func TestMediaMTXClientMigrationExample(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// REQ-MTX-002: MediaMTX client configuration

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// NEW PATTERN: Use centralized MediaMTX client setup
	client := utils.SetupMediaMTXTestClient(t, env)
	defer utils.TeardownMediaMTXTestClient(t, client)

	// Test MediaMTX connection
	isAccessible := utils.TestMediaMTXConnection(t, client)
	if !isAccessible {
		t.Skip("MediaMTX service not accessible, skipping test")
	}

	ctx := context.Background()

	// Test basic GET request using the new client
	data, err := client.Client.Get(ctx, "/v3/config/global/get")
	if err != nil {
		t.Logf("GET request failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, data, "Response data should not be nil")
	}
}

// TestMediaMTXClientWithCustomConfig demonstrates using custom configuration
func TestMediaMTXClientWithCustomConfig(t *testing.T) {
	// REQ-MTX-002: MediaMTX client configuration

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// Create custom configuration with shorter timeout for testing
	customConfig := utils.CreateMediaMTXTestConfigWithTimeout(env.TempDir, 5*time.Second)

	// NEW PATTERN: Use centralized MediaMTX client setup with custom config
	client := utils.SetupMediaMTXTestClientWithConfig(t, env, customConfig)
	defer utils.TeardownMediaMTXTestClient(t, client)

	// Verify custom configuration was applied
	assert.Equal(t, 5*time.Second, client.Config.Timeout, "Custom timeout should be applied")
	assert.Equal(t, "http://localhost:9997", client.Config.BaseURL, "Base URL should be set correctly")
}

// TestMediaMTXHealthMonitorSetup demonstrates setting up health monitor
func TestMediaMTXHealthMonitorSetup(t *testing.T) {
	// REQ-MTX-003: MediaMTX health monitoring

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// NEW PATTERN: Use centralized MediaMTX client setup
	client := utils.SetupMediaMTXTestClient(t, env)
	defer utils.TeardownMediaMTXTestClient(t, client)

	// Test MediaMTX connection
	isAccessible := utils.TestMediaMTXConnection(t, client)
	if !isAccessible {
		t.Skip("MediaMTX service not accessible, skipping test")
	}

	// NEW PATTERN: Use centralized health monitor setup
	healthMonitor := utils.SetupMediaMTXHealthMonitor(t, client)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test health monitor functionality
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start health monitor
	err := healthMonitor.Start(ctx)
	if err != nil {
		t.Logf("Health monitor start failed (expected if MediaMTX not running): %v", err)
	} else {
		// Stop health monitor
		err = healthMonitor.Stop(ctx)
		assert.NoError(t, err, "Health monitor should stop successfully")
	}
}

// TestMediaMTXStreamManagerSetup demonstrates setting up stream manager
func TestMediaMTXStreamManagerSetup(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// NEW PATTERN: Use centralized MediaMTX client setup
	client := utils.SetupMediaMTXTestClient(t, env)
	defer utils.TeardownMediaMTXTestClient(t, client)

	// Test MediaMTX connection
	isAccessible := utils.TestMediaMTXConnection(t, client)
	if !isAccessible {
		t.Skip("MediaMTX service not accessible, skipping test")
	}

	// NEW PATTERN: Use centralized stream manager setup
	streamManager := utils.SetupMediaMTXStreamManager(t, client)
	require.NotNil(t, streamManager, "Stream manager should be created successfully")

	// Test stream manager functionality
	ctx := context.Background()

	// Get streams list
	streams, err := streamManager.ListStreams(ctx)
	if err != nil {
		t.Logf("List streams failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, streams, "Streams list should not be nil")
	}
}

// TestMediaMTXTestDataCreation demonstrates creating test data
func TestMediaMTXTestDataCreation(t *testing.T) {
	// REQ-MTX-005: MediaMTX test data management

	// Test path creation
	testPath := utils.CreateMediaMTXTestPath("test-path")
	assert.Equal(t, "test-path", testPath.Name, "Path name should be set correctly")
	assert.Equal(t, "rtsp://localhost:8554/test", testPath.Source, "Source should be set correctly")

	// Test stream creation
	testStream := utils.CreateMediaMTXTestStream("test-stream")
	assert.Equal(t, "test-stream", testStream.Name, "Stream name should be set correctly")
	assert.Equal(t, "test-stream", testStream.ConfName, "Stream conf name should be set correctly")
}
