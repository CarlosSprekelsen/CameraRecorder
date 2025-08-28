//go:build unit
// +build unit

/*
MediaMTX Health Monitor Comprehensive Unit Tests

Requirements Coverage:
- REQ-MTX-004: Health monitoring

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthMonitor_Creation tests health monitor creation
func TestHealthMonitor_Creation(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")
}

// TestHealthMonitor_StartStop tests health monitor start and stop
func TestHealthMonitor_StartStop(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	ctx := context.Background()

	// Test start
	err := healthMonitor.Start(ctx)
	assert.NoError(t, err, "Health monitor should start successfully")

	// Test stop
	err = healthMonitor.Stop(ctx)
	assert.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetStatus tests health status retrieval
func TestHealthMonitor_GetStatus(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial status
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Should return health status")
	assert.Equal(t, "UNKNOWN", status.Status, "Initial status should be UNKNOWN")
}

// TestHealthMonitor_IsHealthy tests health check
func TestHealthMonitor_IsHealthy(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial health state
	isHealthy := healthMonitor.IsHealthy()
	assert.False(t, isHealthy, "Initial health state should be false")
}

// TestHealthMonitor_CircuitBreaker tests circuit breaker functionality
func TestHealthMonitor_CircuitBreaker(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test initial circuit breaker state
	isOpen := healthMonitor.IsCircuitOpen()
	assert.False(t, isOpen, "Initial circuit breaker should be closed")

	// Test recording success
	healthMonitor.RecordSuccess()
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should remain closed after success")

	// Test recording failure
	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should remain closed after single failure")
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test metrics retrieval
	metrics := healthMonitor.GetMetrics()
	assert.NotNil(t, metrics, "Should return metrics")
	assert.Contains(t, metrics, "total_checks", "Metrics should contain total_checks")
	assert.Contains(t, metrics, "successful_checks", "Metrics should contain successful_checks")
	assert.Contains(t, metrics, "failed_checks", "Metrics should contain failed_checks")
}

// TestHealthMonitor_CheckAllComponents tests component health checks
func TestHealthMonitor_CheckAllComponents(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test component health check
	componentStatus := healthMonitor.CheckAllComponents()
	assert.NotNil(t, componentStatus, "Should return component status")
}

// TestHealthMonitor_GetDetailedStatus tests detailed status retrieval
func TestHealthMonitor_GetDetailedStatus(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test detailed status retrieval
	detailedStatus := healthMonitor.GetDetailedStatus()
	assert.NotNil(t, detailedStatus, "Should return detailed status")
	assert.NotEmpty(t, detailedStatus.Status, "Detailed status should have status")
	assert.NotNil(t, detailedStatus.Timestamp, "Detailed status should have timestamp")
}

// TestHealthMonitor_String tests string representation
func TestHealthMonitor_String(t *testing.T) {
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test string representation
	healthString := healthMonitor.String()
	assert.NotEmpty(t, healthString, "String representation should not be empty")
	assert.Contains(t, healthString, "HealthMonitor", "String should contain HealthMonitor")
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should be created successfully")

	// Test connection to real MediaMTX server
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(config.HealthCheckURL)
	if err != nil {
		t.Skipf("MediaMTX server not available: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "MediaMTX health endpoint should respond with 200 OK")
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

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
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

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        mockServer.URL,
		HealthCheckURL: mockServer.URL + "/health",
		Timeout:        5 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      3,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
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

	config := &mediamtx.MediaMTXConfig{
		BaseURL:        mockServer.URL,
		HealthCheckURL: mockServer.URL + "/health",
		Timeout:        1 * time.Second, // Short timeout
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		CircuitBreaker: mediamtx.CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  30 * time.Second,
			MaxFailures:      3,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	healthMonitor := mediamtx.NewHealthMonitor(config, logger)
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
