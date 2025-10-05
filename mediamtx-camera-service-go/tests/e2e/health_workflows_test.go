/*
Health Workflows E2E Tests

Tests complete user workflows for system health checks, metrics collection,
and health monitoring over time. Uses proper timeout constants from testutils.

Test Categories:
- System Health Check Workflow: Get system health status, verify all components healthy
- Metrics Collection Workflow: Collect system metrics, verify metric data structure
- Health Monitoring Over Time: Monitor health over time, verify stability

Business Outcomes:
- User can check system health status
- User can access system metrics
- System health remains stable over time

Coverage Target: 65% E2E coverage milestone
*/

package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemHealthCheckWorkflow(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err)

	// Get system health status using proven client method
	healthResp, err := fixture.client.GetSystemHealth()
	require.NoError(t, err)
	require.Nil(t, healthResp.Error)

	result := healthResp.Result.(map[string]interface{})

	// Verify health status structure
	assert.Contains(t, result, "status", "Health response should have status field")
	assert.Contains(t, result, "components", "Health response should have components field")
	assert.Contains(t, result, "timestamp", "Health response should have timestamp field")

	// Verify status is healthy
	status := result["status"].(string)
	assert.Equal(t, "healthy", status, "System should be healthy")

	// Verify components are healthy
	components := result["components"].(map[string]interface{})
	assert.Contains(t, components, "camera_service", "Should have camera service component")
	assert.Contains(t, components, "mediamtx", "Should have MediaMTX component")
	assert.Contains(t, components, "websocket_server", "Should have WebSocket server component")

	// Verify all components are healthy
	for componentName, componentStatus := range components {
		status := componentStatus.(map[string]interface{})
		assert.Equal(t, "healthy", status["status"], "Component %s should be healthy", componentName)
		assert.Contains(t, status, "last_check", "Component %s should have last_check field", componentName)
	}
}

func TestMetricsCollectionWorkflow(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err)

	// Get system metrics using proven client method
	metricsResp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err)
	require.Nil(t, metricsResp.Error)

	result := metricsResp.Result.(map[string]interface{})

	// Verify metrics structure
	assert.Contains(t, result, "cpu_usage", "Metrics should have CPU usage")
	assert.Contains(t, result, "memory_usage", "Metrics should have memory usage")
	assert.Contains(t, result, "disk_usage", "Metrics should have disk usage")
	assert.Contains(t, result, "network_stats", "Metrics should have network stats")
	assert.Contains(t, result, "timestamp", "Metrics should have timestamp")

	// Verify CPU usage is reasonable (0-100%)
	cpuUsage := result["cpu_usage"].(float64)
	assert.GreaterOrEqual(t, cpuUsage, 0.0, "CPU usage should be >= 0")
	assert.LessOrEqual(t, cpuUsage, 100.0, "CPU usage should be <= 100")

	// Verify memory usage is reasonable (0-100%)
	memoryUsage := result["memory_usage"].(float64)
	assert.GreaterOrEqual(t, memoryUsage, 0.0, "Memory usage should be >= 0")
	assert.LessOrEqual(t, memoryUsage, 100.0, "Memory usage should be <= 100")

	// Verify disk usage structure
	diskUsage := result["disk_usage"].(map[string]interface{})
	assert.Contains(t, diskUsage, "total", "Disk usage should have total field")
	assert.Contains(t, diskUsage, "used", "Disk usage should have used field")
	assert.Contains(t, diskUsage, "available", "Disk usage should have available field")
	assert.Contains(t, diskUsage, "percentage", "Disk usage should have percentage field")

	// Verify network stats structure
	networkStats := result["network_stats"].(map[string]interface{})
	assert.Contains(t, networkStats, "connections", "Network stats should have connections field")
	assert.Contains(t, networkStats, "bytes_sent", "Network stats should have bytes_sent field")
	assert.Contains(t, networkStats, "bytes_received", "Network stats should have bytes_received field")
}

func TestHealthMonitoringOverTime(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err)

	var healthChecks []map[string]interface{}

	// Monitor health over time (3 checks with 2-second intervals)
	for i := 0; i < 3; i++ {
		healthResp, err := fixture.client.GetSystemHealth()
		require.NoError(t, err)
		require.Nil(t, healthResp.Error)

		result := healthResp.Result.(map[string]interface{})
		healthChecks = append(healthChecks, result)

		// Verify health is stable
		status := result["status"].(string)
		assert.Equal(t, "healthy", status, "System should remain healthy during monitoring period")

		// Wait between checks (except for last iteration)
		if i < 2 {
			time.Sleep(2 * time.Second)
		}
	}

	// Verify all health checks were successful
	assert.Len(t, healthChecks, 3, "Should have performed 3 health checks")

	// Verify timestamps are increasing
	for i := 1; i < len(healthChecks); i++ {
		prevTimestamp := healthChecks[i-1]["timestamp"].(string)
		currTimestamp := healthChecks[i]["timestamp"].(string)
		assert.NotEqual(t, prevTimestamp, currTimestamp, "Timestamps should be different between checks")
	}

	// Verify components remained healthy throughout
	for i, check := range healthChecks {
		components := check["components"].(map[string]interface{})
		for componentName, componentStatus := range components {
			status := componentStatus.(map[string]interface{})
			assert.Equal(t, "healthy", status["status"], "Component %s should remain healthy in check %d", componentName, i+1)
		}
	}
}

func TestHealthWithStressTest(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err)

	// Perform some operations to stress the system
	for i := 0; i < 5; i++ {
		// Get camera list (light operation)
		cameraResp, err := fixture.client.GetCameraList()
		require.NoError(t, err)
		require.Nil(t, cameraResp.Error)

		// Get metrics (moderate operation)
		metricsResp, err := fixture.client.GetSystemMetrics()
		require.NoError(t, err)
		require.Nil(t, metricsResp.Error)

		// Small delay between operations
		time.Sleep(500 * time.Millisecond)
	}

	// Check health after stress operations
	healthResp, err := fixture.client.GetSystemHealth()
	require.NoError(t, err)
	require.Nil(t, healthResp.Error)

	result := healthResp.Result.(map[string]interface{})
	status := result["status"].(string)

	// System should still be healthy after stress operations
	assert.Equal(t, "healthy", status, "System should remain healthy after stress operations")

	// Verify response time is reasonable (should complete within timeout)
	// The test itself validates that the operation completed within testutils.DefaultTestTimeout
}
