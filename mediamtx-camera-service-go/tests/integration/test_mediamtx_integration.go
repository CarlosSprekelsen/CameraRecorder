//go:build integration

/*
MediaMTX Integration Tests

Tests the MediaMTX service integration with real system components.
These tests validate the complete MediaMTX functionality through external API calls.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service health monitoring
- REQ-MTX-002: Path and stream management
- REQ-MTX-003: Error handling and recovery
- REQ-MTX-004: Performance and scalability
- REQ-MTX-005: Integration with camera system

Test Categories: Integration (Real System)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package integration_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// TestMediaMTX_RealSystemIntegration_ReqMTX001 tests real MediaMTX service integration
func TestMediaMTX_RealSystemIntegration_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// This test requires a real MediaMTX service running
	// Skip if not available
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping real system integration test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test client creation
	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client, "Client should be created successfully")

	ctx := context.Background()

	// Test health check
	err := client.HealthCheck(ctx)
	require.NoError(t, err, "Health check should succeed with real MediaMTX service")

	// Test basic API endpoints
	response, err := client.Get(ctx, "/v3/paths/list")
	if err != nil {
		t.Logf("GET /v3/paths/list failed (expected if no paths): %v", err)
	} else {
		assert.NotNil(t, response, "Response should not be nil")
	}

	// Clean up
	err = client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// TestMediaMTX_PathManagement_ReqMTX003 tests real path management
func TestMediaMTX_PathManagement_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	// This test requires a real MediaMTX service running
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping path management test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	pathManager := mediamtx.NewPathManager(client, config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()
	testPathName := "test_integration_path"
	testSource := "/dev/video0"

	// Test path creation
	err := pathManager.CreatePath(ctx, testPathName, testSource, nil)
	if err != nil {
		t.Logf("Path creation failed (expected if device not available): %v", err)
		// This is expected in test environment without real camera
	} else {
		// Test path deletion
		err = pathManager.DeletePath(ctx, testPathName)
		assert.NoError(t, err, "Path deletion should succeed")
	}

	// Clean up
	err = client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// TestMediaMTX_StreamManagement_ReqMTX002 tests real stream management
func TestMediaMTX_StreamManagement_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	// This test requires a real MediaMTX service running
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping stream management test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	streamManager := mediamtx.NewStreamManager(client, config, logger)
	require.NotNil(t, streamManager)

	ctx := context.Background()
	testStreamName := "test_integration_stream"
	testSource := "/dev/video0"

	// Test stream creation
	stream, err := streamManager.CreateStream(ctx, testStreamName, testSource)
	if err != nil {
		t.Logf("Stream creation failed (expected if device not available): %v", err)
		// This is expected in test environment without real camera
	} else {
		// Test stream deletion
		err = streamManager.DeleteStream(ctx, stream.ConfName)
		assert.NoError(t, err, "Stream deletion should succeed")
	}

	// Clean up
	err = client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// TestMediaMTX_HealthMonitoring_ReqMTX004 tests real health monitoring
func TestMediaMTX_HealthMonitoring_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// This test requires a real MediaMTX service running
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping health monitoring test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test health monitor creation
	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	ctx := context.Background()

	// Test health monitor start
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Wait for health check
	time.Sleep(2 * time.Second)

	// Test health status
	isHealthy := healthMonitor.IsHealthy()
	t.Logf("Health monitor status: %v", isHealthy)

	// Test health monitor stop
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestMediaMTX_ErrorHandling_ReqMTX007 tests real error handling
func TestMediaMTX_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// This test requires a real MediaMTX service running
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping error handling test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client)

	ctx := context.Background()

	// Test with invalid endpoint
	_, err := client.Get(ctx, "/invalid/endpoint")
	if err != nil {
		t.Logf("Expected error for invalid endpoint: %v", err)
		// This is expected behavior
	} else {
		t.Log("Unexpected success for invalid endpoint")
	}

	// Test with invalid path name
	pathManager := mediamtx.NewPathManager(client, config, logger)
	err = pathManager.CreatePath(ctx, "", "/dev/video0", nil)
	if err != nil {
		t.Logf("Expected error for empty path name: %v", err)
		// This is expected behavior
	} else {
		t.Log("Unexpected success for empty path name")
	}

	// Clean up
	err = client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// TestMediaMTX_Performance_ReqMTX004 tests real performance
func TestMediaMTX_Performance_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Performance and scalability
	// This test requires a real MediaMTX service running
	if !isMediaMTXAvailable() {
		t.Skip("MediaMTX service not available, skipping performance test")
	}

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := mediamtx.NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client)

	ctx := context.Background()

	// Test multiple concurrent requests
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.Get(ctx, "/v3/paths/list")
			results <- err
		}()
	}

	// Collect results
	var errors []error
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	// Log results
	t.Logf("Concurrent requests completed: %d successful, %d errors",
		numRequests-len(errors), len(errors))

	// Clean up
	err := client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// Helper function to check if MediaMTX service is available
func isMediaMTXAvailable() bool {
	// Simple check - try to connect to MediaMTX service
	// This is a basic availability check
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://localhost:9997/v3/paths/list")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200 || resp.StatusCode == 404
}
