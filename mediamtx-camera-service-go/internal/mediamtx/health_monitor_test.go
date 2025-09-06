/*
MediaMTX Health Monitor Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHealthMonitor_ReqMTX004 tests health monitor creation
func TestNewHealthMonitor_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy initially")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open initially")
}

// TestHealthMonitor_StartStop_ReqMTX004 tests health monitor start/stop
func TestHealthMonitor_StartStop_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor)

	ctx := context.Background()

	// Start health monitoring
	err = healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Wait for initial health check
	time.Sleep(2 * time.Second)

	// Verify health state
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open")

	// Stop health monitoring
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetStatus_ReqMTX004 tests health status retrieval
func TestHealthMonitor_GetStatus_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor)

	// Get initial status
	status := healthMonitor.GetStatus()
	require.NotNil(t, status, "Status should not be nil")
	assert.Equal(t, "healthy", status.Status, "Initial status should be healthy")

	ctx := context.Background()

	// Start monitoring
	err = healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Wait for health check
	time.Sleep(2 * time.Second)

	// Get status after monitoring
	status = healthMonitor.GetStatus()
	require.NotNil(t, status, "Status should not be nil")
	assert.Equal(t, "healthy", status.Status, "Status should be healthy after monitoring")

	// Stop monitoring
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_GetMetrics_ReqMTX004 tests health metrics retrieval
func TestHealthMonitor_GetMetrics_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor)

	// Get initial metrics
	metrics := healthMonitor.GetMetrics()
	require.NotNil(t, metrics, "Metrics should not be nil")
	assert.Contains(t, metrics, "is_healthy", "Metrics should contain is_healthy")
	assert.Contains(t, metrics, "failure_count", "Metrics should contain failure_count")

	ctx := context.Background()

	// Start monitoring
	err = healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Wait for health check
	time.Sleep(2 * time.Second)

	// Get metrics after monitoring
	metrics = healthMonitor.GetMetrics()
	require.NotNil(t, metrics, "Metrics should not be nil")
	assert.True(t, metrics["is_healthy"].(bool), "Should be healthy after monitoring")

	// Stop monitoring
	err = healthMonitor.Stop(ctx)
	require.NoError(t, err, "Health monitor should stop successfully")
}

// TestHealthMonitor_RecordSuccess_ReqMTX004 tests success recording
func TestHealthMonitor_RecordSuccess_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor)

	// Record success
	healthMonitor.RecordSuccess()

	// Verify health state
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy after success")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open after success")

	// Get status
	status := healthMonitor.GetStatus()
	require.NotNil(t, status, "Status should not be nil")
	assert.Equal(t, "healthy", status.Status, "Status should be healthy after success")
}

// TestHealthMonitor_RecordFailure_ReqMTX004 tests failure recording
func TestHealthMonitor_RecordFailure_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 2, // Lower threshold for testing
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.NewLogger("health-monitor-test")
	logger.SetLevel(logrus.ErrorLevel)

	healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
	require.NotNil(t, healthMonitor)

	// Record failure
	healthMonitor.RecordFailure()

	// Verify health state (should still be healthy with one failure)
	assert.True(t, healthMonitor.IsHealthy(), "Should still be healthy with one failure")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open with one failure")

	// Record another failure to trigger threshold
	healthMonitor.RecordFailure()

	// Now should be unhealthy
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after threshold failures")
	assert.True(t, healthMonitor.IsCircuitOpen(), "Circuit should be open after threshold failures")

	// Get status
	status := healthMonitor.GetStatus()
	require.NotNil(t, status, "Status should not be nil")
	assert.Equal(t, "unhealthy", status.Status, "Status should be unhealthy after failures")
}

// TestHealthMonitor_Configuration_ReqMTX004 tests different configurations
func TestHealthMonitor_Configuration_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	// Test different configurations
	configs := []*MediaMTXConfig{
		{
			BaseURL:                helper.GetConfig().BaseURL,
			Timeout:                30 * time.Second,
			HealthFailureThreshold: 1,
			HealthCheckTimeout:     1 * time.Second,
		},
		{
			BaseURL:                helper.GetConfig().BaseURL,
			Timeout:                30 * time.Second,
			HealthFailureThreshold: 5,
			HealthCheckTimeout:     10 * time.Second,
		},
		{
			BaseURL:                helper.GetConfig().BaseURL,
			Timeout:                30 * time.Second,
			HealthFailureThreshold: 3,
			HealthCheckTimeout:     5 * time.Second,
		},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("config_%d", i), func(t *testing.T) {
			logger := logging.NewLogger("health-monitor-test")
			logger.SetLevel(logrus.ErrorLevel)

			healthMonitor := NewHealthMonitor(helper.GetClient(), config, logger)
			require.NotNil(t, healthMonitor, "Health monitor should be created")

			// Test health state
			assert.True(t, healthMonitor.IsHealthy(), "Should be healthy initially")
			assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open initially")
		})
	}
}
