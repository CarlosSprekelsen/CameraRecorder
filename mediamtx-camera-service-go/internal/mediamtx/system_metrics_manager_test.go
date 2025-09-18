/*
MediaMTX System Metrics Manager Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-004: Health monitoring and system metrics
- REQ-API-001: JSON-RPC API compliance for metrics endpoints

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSystemMetricsManager_ReqMTX004 tests system metrics manager creation
func TestNewSystemMetricsManager_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and system metrics
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	// Create SystemMetricsManager with proper dependencies
	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager, "SystemMetricsManager should be created successfully")
}

// TestSystemMetricsManager_GetStorageInfoAPI_ReqMTX004 tests storage info API method
func TestSystemMetricsManager_GetStorageInfoAPI_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and system metrics
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	ctx := context.Background()

	// Test GetStorageInfoAPI method - new API-ready response
	response, err := systemMetricsManager.GetStorageInfoAPI(ctx)
	require.NoError(t, err, "GetStorageInfoAPI should succeed")
	require.NotNil(t, response, "GetStorageInfoAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.GreaterOrEqual(t, response.TotalSpace, int64(0), "Total space should be non-negative")
	assert.GreaterOrEqual(t, response.UsedSpace, int64(0), "Used space should be non-negative")
	assert.GreaterOrEqual(t, response.AvailableSpace, int64(0), "Available space should be non-negative")
	assert.GreaterOrEqual(t, response.UsagePercentage, 0.0, "Usage percentage should be non-negative")
	assert.LessOrEqual(t, response.UsagePercentage, 100.0, "Usage percentage should not exceed 100%")
	assert.GreaterOrEqual(t, response.RecordingsSize, int64(0), "Recordings size should be non-negative")
	assert.GreaterOrEqual(t, response.SnapshotsSize, int64(0), "Snapshots size should be non-negative")

	// Validate logical consistency
	assert.LessOrEqual(t, response.UsedSpace, response.TotalSpace, "Used space should not exceed total space")
	assert.Equal(t, response.TotalSpace-response.UsedSpace, response.AvailableSpace, "Available space should equal total minus used")
}

// TestSystemMetricsManager_GetSystemMetricsAPI_ReqMTX004 tests system metrics API method
func TestSystemMetricsManager_GetSystemMetricsAPI_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and system metrics
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	ctx := context.Background()

	// Test GetSystemMetricsAPI method - new API-ready response
	response, err := systemMetricsManager.GetSystemMetricsAPI(ctx)
	require.NoError(t, err, "GetSystemMetricsAPI should succeed")
	require.NotNil(t, response, "GetSystemMetricsAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.GreaterOrEqual(t, response.CPUUsage, 0.0, "CPU usage should be non-negative")
	assert.LessOrEqual(t, response.CPUUsage, 100.0, "CPU usage should not exceed 100%")
	assert.GreaterOrEqual(t, response.MemoryUsage, 0.0, "Memory usage should be non-negative")
	assert.LessOrEqual(t, response.MemoryUsage, 100.0, "Memory usage should not exceed 100%")
	assert.GreaterOrEqual(t, response.DiskUsage, 0.0, "Disk usage should be non-negative")
	assert.LessOrEqual(t, response.DiskUsage, 100.0, "Disk usage should not exceed 100%")
	assert.Greater(t, response.Goroutines, 0, "Should have active goroutines")
	assert.Greater(t, response.HeapAlloc, int64(0), "Heap allocation should be positive")
}

// TestSystemMetricsManager_GetMetricsAPI_ReqMTX004 tests comprehensive metrics API method
func TestSystemMetricsManager_GetMetricsAPI_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and system metrics
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	ctx := context.Background()

	// Test GetMetricsAPI method - new comprehensive API-ready response
	response, err := systemMetricsManager.GetMetricsAPI(ctx)
	require.NoError(t, err, "GetMetricsAPI should succeed")
	require.NotNil(t, response, "GetMetricsAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.NotEmpty(t, response.Timestamp, "Response should include timestamp")
	assert.NotNil(t, response.SystemMetrics, "Response should include system metrics")
	assert.NotNil(t, response.ComponentMetrics, "Response should include component metrics")

	// Validate system metrics structure
	systemMetrics := response.SystemMetrics
	if cpuUsage, exists := systemMetrics["cpu_usage"]; exists {
		if cpuFloat, ok := cpuUsage.(float64); ok {
			assert.GreaterOrEqual(t, cpuFloat, 0.0, "CPU usage should be non-negative")
			assert.LessOrEqual(t, cpuFloat, 100.0, "CPU usage should not exceed 100%")
		}
	}

	if memUsage, exists := systemMetrics["memory_usage"]; exists {
		if memFloat, ok := memUsage.(float64); ok {
			assert.GreaterOrEqual(t, memFloat, 0.0, "Memory usage should be non-negative")
			assert.LessOrEqual(t, memFloat, 100.0, "Memory usage should not exceed 100%")
		}
	}
}

// TestSystemMetricsManager_APICompliance_ReqAPI001 tests API compliance for all methods
func TestSystemMetricsManager_APICompliance_ReqAPI001(t *testing.T) {
	// REQ-API-001: JSON-RPC API compliance for metrics endpoints
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	ctx := context.Background()

	// Test all API methods for compliance with JSON-RPC documentation
	t.Run("GetStorageInfoAPI_Compliance", func(t *testing.T) {
		response, err := systemMetricsManager.GetStorageInfoAPI(ctx)
		require.NoError(t, err, "API method should not return error")
		require.NotNil(t, response, "API method should return structured response")

		// Validate all documented fields are present
		assert.IsType(t, int64(0), response.TotalSpace, "TotalSpace should be int64")
		assert.IsType(t, int64(0), response.UsedSpace, "UsedSpace should be int64")
		assert.IsType(t, int64(0), response.AvailableSpace, "AvailableSpace should be int64")
		assert.IsType(t, float64(0), response.UsagePercentage, "UsagePercentage should be float64")
		assert.IsType(t, int64(0), response.RecordingsSize, "RecordingsSize should be int64")
		assert.IsType(t, int64(0), response.SnapshotsSize, "SnapshotsSize should be int64")
		assert.IsType(t, false, response.LowSpaceWarning, "LowSpaceWarning should be bool")
	})

	t.Run("GetSystemMetricsAPI_Compliance", func(t *testing.T) {
		response, err := systemMetricsManager.GetSystemMetricsAPI(ctx)
		require.NoError(t, err, "API method should not return error")
		require.NotNil(t, response, "API method should return structured response")

		// Validate all documented fields are present
		assert.IsType(t, float64(0), response.CPUUsage, "CPUUsage should be float64")
		assert.IsType(t, float64(0), response.MemoryUsage, "MemoryUsage should be float64")
		assert.IsType(t, float64(0), response.DiskUsage, "DiskUsage should be float64")
		assert.IsType(t, 0, response.Goroutines, "Goroutines should be int")
		assert.IsType(t, int64(0), response.HeapAlloc, "HeapAlloc should be int64")
	})

	t.Run("GetMetricsAPI_Compliance", func(t *testing.T) {
		response, err := systemMetricsManager.GetMetricsAPI(ctx)
		require.NoError(t, err, "API method should not return error")
		require.NotNil(t, response, "API method should return structured response")

		// Validate all documented fields are present
		assert.IsType(t, "", response.Timestamp, "Timestamp should be string")
		assert.NotNil(t, response.SystemMetrics, "SystemMetrics should be present")
		assert.NotNil(t, response.ComponentMetrics, "ComponentMetrics should be present")
		assert.IsType(t, map[string]interface{}{}, response.SystemMetrics, "SystemMetrics should be map")
		assert.IsType(t, map[string]interface{}{}, response.ComponentMetrics, "ComponentMetrics should be map")
	})
}

// TestSystemMetricsManager_ErrorHandling_ReqMTX004 tests error handling scenarios
func TestSystemMetricsManager_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and system metrics - error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Methods should handle cancelled context gracefully
	_, err := systemMetricsManager.GetStorageInfoAPI(cancelledCtx)
	// Error handling should be graceful - either succeed quickly or return context error
	if err != nil {
		assert.Contains(t, err.Error(), "context", "Context cancellation should be handled properly")
	}
}

// TestSystemMetricsManager_ProgressiveReadiness_ReqMTX004 tests progressive readiness pattern
func TestSystemMetricsManager_ProgressiveReadiness_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring - Progressive Readiness Pattern
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create SystemMetricsManager using existing test infrastructure
	config := helper.GetConfig()
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, helper.GetLogger())
	logger := helper.GetLogger()

	systemMetricsManager := NewSystemMetricsManager(
		config,
		configIntegration,
		logger,
	)
	require.NotNil(t, systemMetricsManager)

	ctx := context.Background()

	// SystemMetricsManager should work immediately after creation (Progressive Readiness)
	// No initialization or warm-up period required
	response, err := systemMetricsManager.GetStorageInfoAPI(ctx)
	require.NoError(t, err, "SystemMetricsManager should work immediately (Progressive Readiness)")
	require.NotNil(t, response, "Should return valid response immediately")

	// All methods should be available immediately
	sysResponse, err := systemMetricsManager.GetSystemMetricsAPI(ctx)
	require.NoError(t, err, "GetSystemMetricsAPI should work immediately")
	require.NotNil(t, sysResponse, "Should return valid system metrics immediately")

	metricsResponse, err := systemMetricsManager.GetMetricsAPI(ctx)
	require.NoError(t, err, "GetMetricsAPI should work immediately")
	require.NotNil(t, metricsResponse, "Should return valid comprehensive metrics immediately")
}
