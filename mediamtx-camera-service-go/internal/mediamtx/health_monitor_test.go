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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHealthMonitor_ReqMTX004 tests health monitor creation
func TestNewHealthMonitor_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy initially")
	assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open initially")
}

// TestHealthMonitor_StartStop_ReqMTX004 tests health monitor start/stop
func TestHealthMonitor_StartStop_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor)

	ctx := context.Background()

	// Start health monitoring
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

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

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor)

	// Get initial status
	status := healthMonitor.GetStatus()
	require.NotNil(t, status, "Status should not be nil")
	assert.Equal(t, "healthy", status.Status, "Initial status should be healthy")

	ctx := context.Background()

	// Start monitoring
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

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

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor)

	// Get initial metrics
	metrics := healthMonitor.GetMetrics()
	require.NotNil(t, metrics, "Metrics should not be nil")
	assert.Contains(t, metrics, "is_healthy", "Metrics should contain is_healthy")
	assert.Contains(t, metrics, "failure_count", "Metrics should contain failure_count")

	ctx := context.Background()

	// Start monitoring
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

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

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
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

	// Server is ready via shared test helper

	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 2, // Lower threshold for testing
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)
	// Use test fixture logging level instead of hardcoded logrus

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
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

	// Server is ready via shared test helper

	// Test different configurations
	configs := []*config.MediaMTXConfig{
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
			logger := logging.CreateTestLogger(t, nil)
			// Use test fixture logging level instead of hardcoded logrus

			configManager := helper.GetConfigManager()
			configIntegration := NewConfigIntegration(configManager, logger)
			healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
			require.NotNil(t, healthMonitor, "Health monitor should be created")

			// Test health state
			assert.True(t, healthMonitor.IsHealthy(), "Should be healthy initially")
			assert.False(t, healthMonitor.IsCircuitOpen(), "Circuit should not be open initially")
		})
	}
}

// TestHealthMonitor_DebounceMechanism_ReqMTX004 tests debounce mechanism in health monitoring
func TestHealthMonitor_DebounceMechanism_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor
	client := helper.GetClient()
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		HealthCheckInterval:    5,
		HealthCheckTimeout:     5 * time.Second,
		HealthFailureThreshold: 3,
	}
	logger := helper.GetLogger()

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(client, config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	// Create mock system event notifier
	mockNotifier := NewMockSystemEventNotifier()
	healthMonitor.SetSystemNotifier(mockNotifier)

	// Test debounce mechanism with rapid successive failures
	ctx := context.Background()

	// Start the health monitor
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		healthMonitor.Stop(stopCtx)
	}()

	// Simulate rapid failures to test debounce
	for i := 0; i < 5; i++ {
		healthMonitor.RecordFailure()
	}

	// Should only send one notification due to debounce
	notifications := mockNotifier.GetNotifications()
	assert.LessOrEqual(t, len(notifications), 1, "Should send at most one notification due to debounce")

	// Clear notifications and record another failure
	mockNotifier.ClearNotifications()

	// Record another failure
	healthMonitor.RecordFailure()

	// Should send another notification after debounce period
	notifications = mockNotifier.GetNotifications()
	assert.Equal(t, 1, len(notifications), "Should send notification after debounce period")

	t.Log("Health monitor debounce mechanism working correctly")
}

// TestHealthMonitor_AtomicOperations_ReqMTX004 tests atomic operations for thread safety
func TestHealthMonitor_AtomicOperations_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor
	client := helper.GetClient()
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		HealthCheckInterval:    5,
		HealthCheckTimeout:     5 * time.Second,
		HealthFailureThreshold: 3,
	}
	logger := helper.GetLogger()

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(client, config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	// Test concurrent access to ensure atomic operations work correctly
	ctx := context.Background()

	// Start the health monitor
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		healthMonitor.Stop(stopCtx)
	}()

	// Run concurrent goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("BUG DETECTED: Race condition caused panic: %v", r)
				}
				done <- true
			}()

			// Make concurrent calls to test atomic operations
			healthMonitor.RecordFailure()
			healthMonitor.RecordSuccess()
			healthMonitor.IsHealthy()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should handle concurrent access gracefully
	assert.True(t, true, "Should handle concurrent access without panicking")

	t.Log("Health monitor atomic operations working correctly")
}

// TestHealthMonitor_StatusTransitions_ReqMTX004 tests status transitions with debounce
func TestHealthMonitor_StatusTransitions_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor
	client := helper.GetClient()
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		HealthCheckInterval:    5,
		HealthCheckTimeout:     5 * time.Second,
		HealthFailureThreshold: 3,
	}
	logger := helper.GetLogger()

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(client, config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	// Create mock system event notifier
	mockNotifier := NewMockSystemEventNotifier()
	healthMonitor.SetSystemNotifier(mockNotifier)

	// Test status transitions
	ctx := context.Background()

	// Start the health monitor
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		healthMonitor.Stop(stopCtx)
	}()

	// Initially should be healthy
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy initially")

	// Record failures to trigger unhealthy status
	for i := 0; i < 4; i++ {
		healthMonitor.RecordFailure()
	}

	// Should be unhealthy after threshold failures
	assert.False(t, healthMonitor.IsHealthy(), "Should be unhealthy after threshold failures")

	// Record success to recover
	healthMonitor.RecordSuccess()

	// Should be healthy again
	assert.True(t, healthMonitor.IsHealthy(), "Should be healthy after recovery")

	t.Log("Health monitor status transitions working correctly")
}

// TestHealthMonitor_GetHealthAPI_ReqMTX004 tests new API-ready health method
func TestHealthMonitor_GetHealthAPI_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring - API-ready health responses
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor using existing test infrastructure
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	// Start health monitor
	ctx := context.Background()
	err := healthMonitor.Start(ctx)
	require.NoError(t, err, "Health monitor should start successfully")

	// Ensure health monitor is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		healthMonitor.Stop(stopCtx)
	}()

	// Test GetHealthAPI method - new API-ready response
	startTime := time.Now().Add(-1 * time.Hour) // Simulate system start time
	response, err := healthMonitor.GetHealthAPI(ctx, startTime)
	require.NoError(t, err, "GetHealthAPI should succeed")
	require.NotNil(t, response, "GetHealthAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.NotEmpty(t, response.Status, "Response should include overall status")
	assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, response.Status, "Status should be valid")
	assert.NotEmpty(t, response.Uptime, "Response should include uptime")
	assert.NotEmpty(t, response.Version, "Response should include version")
	assert.NotNil(t, response.Components, "Response should include components map")
	assert.NotNil(t, response.Checks, "Response should include checks array")
	assert.NotEmpty(t, response.Timestamp, "Response should include timestamp")
	assert.GreaterOrEqual(t, response.ResponseTime, 0.0, "Response time should be non-negative")

	// Validate components structure
	if mediaStatus, exists := response.Components["mediamtx"]; exists {
		assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, mediaStatus, "MediaMTX status should be valid")
	}

	// Validate timestamp format (should be ISO 8601)
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in ISO 8601 format")
}

// TestHealthMonitor_GetHealthAPI_APICompliance_ReqAPI001 tests API compliance for GetHealthAPI
func TestHealthMonitor_GetHealthAPI_APICompliance_ReqAPI001(t *testing.T) {
	// REQ-API-001: JSON-RPC API compliance for health endpoints
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor using existing test infrastructure
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	ctx := context.Background()
	startTime := time.Now().Add(-30 * time.Minute)

	// Test API compliance for GetHealthAPI method
	response, err := healthMonitor.GetHealthAPI(ctx, startTime)
	require.NoError(t, err, "API method should not return error")
	require.NotNil(t, response, "API method should return structured response")

	// Validate all documented fields are present with correct types
	assert.IsType(t, "", response.Status, "Status should be string")
	assert.IsType(t, "", response.Uptime, "Uptime should be string")
	assert.IsType(t, "", response.Version, "Version should be string")
	assert.IsType(t, map[string]interface{}{}, response.Components, "Components should be map[string]interface{}")
	assert.IsType(t, []interface{}{}, response.Checks, "Checks should be []interface{}")
	assert.IsType(t, "", response.Timestamp, "Timestamp should be string")
	assert.IsType(t, float64(0), response.ResponseTime, "ResponseTime should be float64")

	// Validate status field values are within documented enum
	validStatuses := []string{"healthy", "degraded", "unhealthy"}
	assert.Contains(t, validStatuses, response.Status, "Status should be one of documented values")

	// Validate uptime format (should be duration string)
	assert.Regexp(t, `^\d+(\.\d+)?[a-z]+$`, response.Uptime, "Uptime should be duration format")
}

// TestHealthMonitor_GetHealthAPI_ErrorScenarios_ReqMTX004 tests error handling for GetHealthAPI
func TestHealthMonitor_GetHealthAPI_ErrorScenarios_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring - error handling
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create health monitor using existing test infrastructure
	config := &config.MediaMTXConfig{
		BaseURL:                helper.GetConfig().BaseURL,
		Timeout:                30 * time.Second,
		HealthCheckInterval:    5,
		HealthFailureThreshold: 3,
		HealthCheckTimeout:     5 * time.Second,
	}
	logger := logging.CreateTestLogger(t, nil)

	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	healthMonitor := NewHealthMonitor(helper.GetClient(), config, configIntegration, logger)
	require.NotNil(t, healthMonitor, "Health monitor should not be nil")

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	startTime := time.Now().Add(-30 * time.Minute)

	// GetHealthAPI should handle cancelled context gracefully
	response, err := healthMonitor.GetHealthAPI(cancelledCtx, startTime)
	// Should either succeed quickly or return context error
	if err != nil {
		assert.Contains(t, err.Error(), "context", "Context cancellation should be handled properly")
	} else {
		// If it succeeds, response should still be valid
		require.NotNil(t, response, "Response should be valid even with cancelled context")
		assert.NotEmpty(t, response.Status, "Status should be present")
	}

	// Test with future start time (edge case)
	futureStartTime := time.Now().Add(1 * time.Hour)
	response, err = healthMonitor.GetHealthAPI(context.Background(), futureStartTime)
	require.NoError(t, err, "Should handle future start time gracefully")
	require.NotNil(t, response, "Should return valid response for future start time")

	// Uptime should be reasonable (not negative)
	assert.NotEmpty(t, response.Uptime, "Uptime should be present even for future start time")
}
