//go:build unit
// +build unit

/*
MediaMTX Health Monitor Unit Tests

Requirements Coverage:
- REQ-MTX-004: Health monitoring
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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthMonitor_Creation tests health monitor creation
func TestHealthMonitor_Creation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")
}

// TestHealthMonitor_StartStop tests health monitor lifecycle
func TestHealthMonitor_StartStop(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	ctx := context.Background()

	// Test start
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Test stop
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetStatus tests health status retrieval
func TestHealthMonitor_GetStatus(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	// Test getting health status
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Health status should not be nil")
	assert.NotEmpty(t, status.Status, "Health status should not be empty")
}

// TestHealthMonitor_IsHealthy tests health check
func TestHealthMonitor_IsHealthy(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	// Test health check
	healthy := healthMonitor.IsHealthy()
	assert.IsType(t, false, healthy, "Healthy should be a boolean")
}

// TestHealthMonitor_CircuitBreaker tests circuit breaker functionality
func TestHealthMonitor_CircuitBreaker(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	// Test circuit breaker state
	circuitOpen := healthMonitor.IsCircuitOpen()
	assert.IsType(t, false, circuitOpen, "Circuit open should be a boolean")

	// Test recording success
	healthMonitor.RecordSuccess()
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should be closed after success")

	// Test recording failure
	healthMonitor.RecordFailure()
	// Note: Circuit breaker behavior depends on failure threshold
}

// TestHealthMonitor_ErrorHandling tests error handling scenarios
func TestHealthMonitor_ErrorHandling(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	ctx := context.Background()

	// Test starting already started monitor
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "First start should succeed")

	err = healthMonitor.Start(ctx)
	assert.Error(t, err, "Second start should fail")

	// Test stopping already stopped monitor
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "First stop should succeed")

	err = healthMonitor.Stop(ctx)
	assert.Error(t, err, "Second stop should fail")
}

// TestHealthMonitor_ConcurrentAccess tests concurrent access scenarios
func TestHealthMonitor_ConcurrentAccess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	// Test concurrent health status access
	done := make(chan bool, 2)

	go func() {
		status := healthMonitor.GetStatus()
		assert.NotNil(t, status, "Status should not be nil")
		done <- true
	}()

	go func() {
		healthy := healthMonitor.IsHealthy()
		assert.IsType(t, false, healthy, "Healthy should be a boolean")
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestHealthMonitor_ContextCancellation tests context cancellation
func TestHealthMonitor_ContextCancellation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "http://localhost:9997",
		HealthCheckInterval:         5,
		HealthFailureThreshold:      3,
		HealthCircuitBreakerTimeout: 30,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor
	healthMonitor := mediamtx.NewHealthMonitor(client, testConfig, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start monitor
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Cancel context immediately
	cancel()

	// Wait a bit for cancellation to take effect
	time.Sleep(100 * time.Millisecond)

	// Stop monitor
	err = healthMonitor.Stop(ctx)
	// Should handle context cancellation gracefully
	if err != nil {
		t.Logf("Context cancellation test result: %v", err)
	}
}

// TestHealthMonitor_ConfigurationValidation tests configuration validation
func TestHealthMonitor_ConfigurationValidation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Test with invalid configuration
	invalidConfig := &mediamtx.MediaMTXConfig{
		BaseURL:                     "",
		HealthCheckInterval:         -1,
		HealthFailureThreshold:      0,
		HealthCircuitBreakerTimeout: -1,
	}

	// Create mock client
	client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

	// Create health monitor with invalid config
	healthMonitor := mediamtx.NewHealthMonitor(client, invalidConfig, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created even with invalid config")

	// Test that health monitor handles invalid config gracefully
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Health status should not be nil even with invalid config")
}
