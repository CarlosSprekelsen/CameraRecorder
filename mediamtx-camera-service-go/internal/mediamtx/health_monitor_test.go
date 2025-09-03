/*
MediaMTX Health Monitor Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMediaMTXClient simulates MediaMTX client behavior for testing
type MockMediaMTXClient struct {
	failCount int
	maxFails  int
}

func (m *MockMediaMTXClient) Get(ctx context.Context, path string) ([]byte, error) {
	return nil, nil
}

func (m *MockMediaMTXClient) Post(ctx context.Context, path string, data []byte) ([]byte, error) {
	return nil, nil
}

func (m *MockMediaMTXClient) Put(ctx context.Context, path string, data []byte) ([]byte, error) {
	return nil, nil
}

func (m *MockMediaMTXClient) Delete(ctx context.Context, path string) error {
	return nil
}

func (m *MockMediaMTXClient) HealthCheck(ctx context.Context) error {
	m.failCount++
	if m.failCount <= m.maxFails {
		return assert.AnError
	}
	return nil // Succeed after maxFails
}

func (m *MockMediaMTXClient) Close() error {
	return nil
}

// TestHealthMonitor_Creation_ReqMTX004 tests health monitor creation
func TestHealthMonitor_Creation_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	require.NotNil(t, healthMonitor, "Health monitor should not be nil")
	assert.True(t, healthMonitor.IsHealthy(), "Initial state should be healthy")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Initial circuit should be closed")
}

// TestHealthMonitor_StartStop_ReqMTX004 tests health monitor lifecycle
func TestHealthMonitor_StartStop_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     1 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)
	require.NotNil(t, healthMonitor)

	ctx := context.Background()

	// Test start
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Test stop
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetStatus_ReqMTX004 tests health status retrieval
func TestHealthMonitor_GetStatus_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 2,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test initial status
	status := healthMonitor.GetStatus()
	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, int64(0), status.ErrorCount)
	assert.Equal(t, "healthy", status.CircuitBreakerState)

	// Test status after failures
	healthMonitor.RecordFailure()
	healthMonitor.RecordFailure()

	status = healthMonitor.GetStatus()
	assert.Equal(t, "unhealthy", status.Status)
	assert.Equal(t, int64(2), status.ErrorCount)
	assert.Equal(t, "unhealthy", status.CircuitBreakerState)
}

// TestHealthMonitor_IsHealthy_ReqMTX004 tests health check functionality
func TestHealthMonitor_IsHealthy_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 3,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test initial state
	assert.True(t, healthMonitor.IsHealthy(), "Initial state should be healthy")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Initial circuit should be closed")

	// Test after failures below threshold
	healthMonitor.RecordFailure()
	healthMonitor.RecordFailure()
	assert.True(t, healthMonitor.IsHealthy(), "Should remain healthy after 2 failures")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should remain closed")

	// Test after threshold exceeded
	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after 3 failures")
	assert.True(t, healthMonitor.IsCircuitOpen(), "Circuit should be open after 3 failures")
}

// TestHealthMonitor_GetMetrics_ReqMTX004 tests metrics retrieval
func TestHealthMonitor_GetMetrics_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test initial metrics
	metrics := healthMonitor.GetMetrics()
	assert.Contains(t, metrics, "is_healthy")
	assert.Contains(t, metrics, "failure_count")
	assert.Contains(t, metrics, "status")

	assert.True(t, metrics["is_healthy"].(bool))
	assert.Equal(t, 0, metrics["failure_count"].(int))
	assert.Equal(t, "healthy", metrics["status"].(string))
}

// TestHealthMonitor_RecordSuccess_ReqMTX004 tests success recording
func TestHealthMonitor_RecordSuccess_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 2,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Make unhealthy first
	healthMonitor.RecordFailure()
	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after 2 failures")

	// Test recovery
	healthMonitor.RecordSuccess()
	assert.True(t, healthMonitor.IsHealthy(), "Should recover after success")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should close after success")
}

// TestHealthMonitor_RecordFailure_ReqMTX004 tests failure recording
func TestHealthMonitor_RecordFailure_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 3,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test failure threshold
	healthMonitor.RecordFailure()
	healthMonitor.RecordFailure()
	assert.True(t, healthMonitor.IsHealthy(), "Should remain healthy after 2 failures")

	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after 3 failures")
	assert.True(t, healthMonitor.IsCircuitOpen(), "Circuit should be open after 3 failures")
}

// TestHealthMonitor_Configuration_ReqMTX004 tests configuration handling
func TestHealthMonitor_Configuration_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}

	// Test with custom threshold
	config := &MediaMTXConfig{
		HealthFailureThreshold: 5,
	}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Should remain healthy until 5 failures
	for i := 0; i < 4; i++ {
		healthMonitor.RecordFailure()
		assert.True(t, healthMonitor.IsHealthy(), "Should remain healthy after %d failures", i+1)
	}

	healthMonitor.RecordFailure()
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after 5 failures")
}

// TestHealthMonitor_ConcurrentAccess_ReqMTX004 tests thread safety
func TestHealthMonitor_ConcurrentAccess_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthFailureThreshold: 10,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			healthMonitor.RecordFailure()
			healthMonitor.IsHealthy()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			healthMonitor.RecordSuccess()
			healthMonitor.GetStatus()
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have consistent state
	status := healthMonitor.GetStatus()
	assert.NotNil(t, status, "Status should not be nil after concurrent access")
}

// TestHealthMonitor_ContextHandling_ReqMTX004 tests context handling
func TestHealthMonitor_ContextHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &MediaMTXConfig{
		HealthCheckTimeout: 1 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mockClient := &MockMediaMTXClient{}
	healthMonitor := NewHealthMonitor(mockClient, config, logger)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Should handle cancelled context gracefully")

	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Should stop gracefully even with cancelled context")
}
