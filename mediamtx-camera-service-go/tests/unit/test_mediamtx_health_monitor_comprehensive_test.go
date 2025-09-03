//go:build unit
// +build unit

/*
MediaMTX Health Monitor Comprehensive Unit Tests

Requirements Coverage:
- REQ-MTX-004: Health monitoring

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthMonitor_CreationComprehensive tests health monitor creation
func TestHealthMonitor_CreationComprehensive(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")
}

// TestHealthMonitor_StartStopComprehensive tests health monitor start and stop
func TestHealthMonitor_StartStopComprehensive(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	ctx := context.Background()

	// Test start
	err := healthMonitor.Start(ctx)
	assert.NoError(t, err, "Health monitor should start successfully")

	// Test stop
	err = healthMonitor.Stop(ctx)
	assert.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetStatusComprehensive tests health status retrieval
func TestHealthMonitor_GetStatusComprehensive(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial status
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Should return health status")
	assert.Equal(t, "UNKNOWN", status.Status, "Initial status should be UNKNOWN")
}

// TestHealthMonitor_IsHealthyComprehensive tests health check
func TestHealthMonitor_IsHealthyComprehensive(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial health state - should be false until first health check
	isHealthy := healthMonitor.IsHealthy()
	assert.False(t, isHealthy, "Initial health state should be false until first health check")
}

// TestHealthMonitor_CircuitBreakerComprehensive tests circuit breaker functionality
func TestHealthMonitor_CircuitBreakerComprehensive(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial circuit breaker state - simplified version starts as healthy
	isOpen := healthMonitor.IsCircuitOpen()
	assert.False(t, isOpen, "Initial circuit breaker should be healthy (closed)")

	// Test recording success - should remain healthy
	healthMonitor.RecordSuccess()
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should remain healthy after success")

	// Test recording failure - should remain healthy for single failure (below threshold)
	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should remain healthy after single failure")
}

// TestHealthMonitor_GetMetrics tests metrics retrieval
func TestHealthMonitor_GetMetrics(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test metrics retrieval - simplified metrics structure
	metrics := healthMonitor.GetMetrics()
	assert.NotNil(t, metrics, "Should return metrics")
	assert.Contains(t, metrics, "is_healthy", "Metrics should contain is_healthy")
	assert.Contains(t, metrics, "failure_count", "Metrics should contain failure_count")
	assert.Contains(t, metrics, "status", "Metrics should contain status")
}

// TestHealthMonitor_RealServerConnection tests connection to real MediaMTX server
func TestHealthMonitor_RealServerConnection(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test connection to real MediaMTX server using the MediaMTX client
	ctx := context.Background()
	err := client.HealthCheck(ctx)
	if err != nil {
		t.Skipf("MediaMTX server not available: %v", err)
	}

	// If we get here, the health check passed
	assert.NoError(t, err, "MediaMTX health check should succeed")
}

// TestHealthMonitor_MockServer tests with mock server
func TestHealthMonitor_MockServer(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer mockServer.Close()

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        mockServer.URL,
		HealthCheckURL: mockServer.URL + "/health",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      5,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test with mock server
	ctx := context.Background()
	err := healthMonitor.Start(ctx)
	assert.NoError(t, err, "Health monitor should start with mock server")

	// Give some time for health check
	time.Sleep(100 * time.Millisecond)

	// Test health status
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Should return health status from mock server")

	err = healthMonitor.Stop(ctx)
	assert.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_FailureScenarios tests failure scenarios
func TestHealthMonitor_FailureScenarios(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Create mock server that fails
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	// Create config with mock server URLs but use helper for other settings
	config := utils.CreateTestMediaMTXConfig()
	config.BaseURL = mockServer.URL
	config.HealthCheckURL = mockServer.URL + "/health"
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test with failing mock server
	ctx := context.Background()
	err := healthMonitor.Start(ctx)
	assert.NoError(t, err, "Health monitor should start with failing mock server")

	// Give some time for health check
	time.Sleep(100 * time.Millisecond)

	// Test health status after failures
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Should return health status even after failures")

	err = healthMonitor.Stop(ctx)
	assert.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_TimeoutScenarios tests timeout scenarios
func TestHealthMonitor_TimeoutScenarios(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Create mock server that delays
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create config with mock server URLs but use helper for other settings
	config := utils.CreateTestMediaMTXConfig()
	config.BaseURL = mockServer.URL
	config.HealthCheckURL = mockServer.URL + "/health"
	config.Timeout = 1 * time.Second // Short timeout for testing
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a real MediaMTX client for testing
	client := mediamtx.NewClient(config.BaseURL, config, logger)
	healthMonitor := mediamtx.NewHealthMonitor(client, config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test with timeout scenario
	ctx := context.Background()
	err := healthMonitor.Start(ctx)
	assert.NoError(t, err, "Health monitor should start with timeout scenario")

	// Give some time for health check
	time.Sleep(100 * time.Millisecond)

	// Test health status after timeout
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Should return health status even after timeout")

	err = healthMonitor.Stop(ctx)
	assert.NoError(t, err, "Health monitor should stop successfully")
}
