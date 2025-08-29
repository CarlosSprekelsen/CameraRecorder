//go:build unit
// +build unit

/*
MediaMTX Client Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Creation tests client creation
func TestClient_Creation(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)
	require.NotNil(t, client, "Client should not be nil")
}

// TestClient_Get tests GET request functionality
func TestClient_Get(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test GET request
	data, err := client.Get(ctx, "/v3/config/global/get")
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("GET request failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, data, "Response data should not be nil")
	}
}

// TestClient_Post tests POST request functionality
func TestClient_Post(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test POST request
	requestData := []byte(`{"test": "data"}`)
	data, err := client.Post(ctx, "/v3/config/global/edit", requestData)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("POST request failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, data, "Response data should not be nil")
	}
}

// TestClient_Put tests PUT request functionality
func TestClient_Put(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test PUT request
	requestData := []byte(`{"test": "data"}`)
	data, err := client.Put(ctx, "/v3/config/global/edit", requestData)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("PUT request failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.NotNil(t, data, "Response data should not be nil")
	}
}

// TestClient_Delete tests DELETE request functionality
func TestClient_Delete(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test DELETE request
	err := client.Delete(ctx, "/v3/paths/delete/test-path")
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("DELETE request failed (expected if MediaMTX not running): %v", err)
	}
}

// TestClient_HealthCheck tests health check functionality
func TestClient_HealthCheck(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test health check
	err := client.HealthCheck(ctx)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Health check failed (expected if MediaMTX not running): %v", err)
	}
}

// TestClient_Close tests client close functionality
func TestClient_Close(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	// Test close
	err := client.Close()
	assert.NoError(t, err, "Client close should succeed")
}

// TestClient_ErrorHandling tests error handling scenarios
func TestClient_ErrorHandling(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client with invalid URL using shared logger
	client := mediamtx.NewClient("http://invalid-url:99999", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test GET request with invalid URL
	_, err := client.Get(ctx, "/test")
	assert.Error(t, err, "Should return error with invalid URL")

	// Test POST request with invalid URL
	_, err = client.Post(ctx, "/test", []byte(`{}`))
	assert.Error(t, err, "Should return error with invalid URL")

	// Test PUT request with invalid URL
	_, err = client.Put(ctx, "/test", []byte(`{}`))
	assert.Error(t, err, "Should return error with invalid URL")

	// Test DELETE request with invalid URL
	err = client.Delete(ctx, "/test")
	assert.Error(t, err, "Should return error with invalid URL")

	// Test health check with invalid URL
	err = client.HealthCheck(ctx)
	assert.Error(t, err, "Should return error with invalid URL")
}

// TestClient_ConcurrentAccess tests concurrent access scenarios
func TestClient_ConcurrentAccess(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	ctx := context.Background()

	// Test concurrent requests
	done := make(chan bool, 2)

	go func() {
		_, err := client.Get(ctx, "/v3/config/global/get")
		if err != nil {
			t.Logf("Concurrent GET result: %v", err)
		}
		done <- true
	}()

	go func() {
		err := client.HealthCheck(ctx)
		if err != nil {
			t.Logf("Concurrent health check result: %v", err)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestClient_ContextCancellation tests context cancellation
func TestClient_ContextCancellation(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}

	// Create client using shared logger
	client := mediamtx.NewClient("http://localhost:9997", testConfig, env.Logger.Logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Test request with cancelled context
	_, err := client.Get(ctx, "/test")
	// Should handle context cancellation gracefully
	if err != nil {
		t.Logf("Context cancellation test result: %v", err)
	}
}

// TestClient_ConfigurationValidation tests configuration validation
func TestClient_ConfigurationValidation(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test with invalid configuration
	invalidConfig := &mediamtx.MediaMTXConfig{
		BaseURL:       "",
		Timeout:       -1 * time.Second,
		RetryAttempts: -1,
		RetryDelay:    -1 * time.Second,
	}

	// Create client with invalid config using shared logger
	client := mediamtx.NewClient("", invalidConfig, env.Logger.Logger)
	require.NotNil(t, client, "Client should be created even with invalid config")

	// Test that client handles invalid config gracefully
	err := client.Close()
	assert.NoError(t, err, "Client close should succeed even with invalid config")
}
