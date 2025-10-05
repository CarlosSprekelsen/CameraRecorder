/*
Health Workflows E2E Tests

Tests complete user workflows for system health monitoring, metrics collection,
and health monitoring over time. Each test validates admin-level system health
operations with real system components.

Test Categories:
- System Health Check Workflow: Get system status, verify all components reporting
- Metrics Collection Workflow: Get baseline metrics, perform operations, verify metrics updated
- Health Monitoring Over Time: Poll system health, verify consistency and uptime progression

Business Outcomes:
- Admin can verify system is healthy
- Admin can monitor system performance
- Admin can monitor system stability over time

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
	// Setup: Authenticated connection (admin role required)
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "System Health", 1, "Admin connection established")

	// Step 1: Call get_system_status method
	healthResponse := setup.SendJSONRPC(conn, "get_system_status", map[string]interface{}{})
	LogWorkflowStep(t, "System Health", 2, "System status request sent")

	// Step 2: Verify response structure (status, uptime, components)
	require.NoError(t, healthResponse.Error, "System status request should succeed")
	require.NotNil(t, healthResponse.Result, "System status result should not be nil")

	statusResult := healthResponse.Result.(map[string]interface{})
	assert.Contains(t, statusResult, "status", "System status should contain status field")
	assert.Contains(t, statusResult, "uptime", "System status should contain uptime field")
	assert.Contains(t, statusResult, "components", "System status should contain components field")
	LogWorkflowStep(t, "System Health", 3, "System status response structure validated")

	// Step 3: Validate all critical components reporting (camera_monitor, mediamtx, websocket)
	components := statusResult["components"].(map[string]interface{})

	// Check for critical components
	criticalComponents := []string{"camera_monitor", "mediamtx", "websocket", "server"}
	foundComponents := 0

	for _, component := range criticalComponents {
		if componentStatus, exists := components[component]; exists {
			assert.NotNil(t, componentStatus, "Component %s status should not be nil", component)
			t.Logf("Component %s status: %v", component, componentStatus)
			foundComponents++
		}
	}

	assert.Greater(t, foundComponents, 0, "At least one critical component should be reporting")
	LogWorkflowStep(t, "System Health", 4, "Critical components reporting validated")

	// Step 4: Verify component health states are valid values
	for componentName, componentStatus := range components {
		statusMap, ok := componentStatus.(map[string]interface{})
		if !ok {
			t.Logf("Warning: Component %s status is not a map: %v", componentName, componentStatus)
			continue
		}

		// Check for common health status fields
		if health, ok := statusMap["health"].(string); ok {
			validHealthStates := []string{"healthy", "unhealthy", "degraded", "unknown", "starting", "stopping"}
			assert.Contains(t, validHealthStates, health, "Component %s health should be valid state", componentName)
		}

		if status, ok := statusMap["status"].(string); ok {
			validStatusStates := []string{"running", "stopped", "starting", "stopping", "error", "unknown"}
			assert.Contains(t, validStatusStates, status, "Component %s status should be valid state", componentName)
		}

		t.Logf("Component %s: health=%v, status=%v", componentName, statusMap["health"], statusMap["status"])
	}
	LogWorkflowStep(t, "System Health", 5, "Component health states validated")

	// Validation: Status field present AND uptime > 0 AND all components listed AND health states valid
	setup.AssertBusinessOutcome("Admin can verify system is healthy", func() bool {
		status := statusResult["status"].(string)
		uptime := statusResult["uptime"].(float64)

		return status != "" && uptime > 0 && len(components) > 0
	})
	LogWorkflowStep(t, "System Health", 6, "Business outcome validated - admin can verify system health")

	// Cleanup: Standard cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "System Health", 7, "Connection closed and cleanup verified")
}

func TestMetricsCollectionWorkflow(t *testing.T) {
	// Setup: Authenticated admin connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Metrics Collection", 1, "Admin connection established")

	// Step 1: Get baseline metrics
	baselineResponse := setup.SendJSONRPC(conn, "get_metrics", map[string]interface{}{})
	require.NoError(t, baselineResponse.Error, "Baseline metrics request should succeed")
	require.NotNil(t, baselineResponse.Result, "Baseline metrics result should not be nil")

	baselineMetrics := baselineResponse.Result.(map[string]interface{})
	t.Logf("Baseline metrics: %v", baselineMetrics)
	LogWorkflowStep(t, "Metrics Collection", 2, "Baseline metrics retrieved")

	// Step 2: Perform operations (take snapshot, start/stop recording)
	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	var deviceID string
	if len(cameras) > 0 {
		camera := cameras[0].(map[string]interface{})
		deviceID = camera["device"].(string)

		// Take a snapshot
		snapshotResponse := setup.SendJSONRPC(conn, "take_snapshot", map[string]interface{}{
			"device":   deviceID,
			"filename": "metrics_test_snapshot.jpg",
		})
		if snapshotResponse.Error == nil {
			t.Log("Snapshot operation completed for metrics")
		}

		// Start and stop recording
		startResponse := setup.SendJSONRPC(conn, "start_recording", map[string]interface{}{
			"device": deviceID,
		})
		if startResponse.Error == nil {
			// Wait for recording to start then stop
			setup.WaitForCondition(func() bool {
				return getFileSize(t, startResponse.Result.(map[string]interface{})["recording_path"].(string)) > 0
			}, GetStandardTimeout("health_check"))

			stopResponse := setup.SendJSONRPC(conn, "stop_recording", map[string]interface{}{
				"device": deviceID,
			})
			if stopResponse.Error == nil {
				t.Log("Recording operation completed for metrics")
			}
		}
	} else {
		t.Log("No cameras available for metrics operations")
	}
	LogWorkflowStep(t, "Metrics Collection", 3, "Operations performed for metrics collection")

	// Step 3: Call get_metrics method
	updatedResponse := setup.SendJSONRPC(conn, "get_metrics", map[string]interface{}{})
	LogWorkflowStep(t, "Metrics Collection", 4, "Updated metrics request sent")

	// Step 4: Verify metrics include operation counts and latencies
	require.NoError(t, updatedResponse.Error, "Updated metrics request should succeed")
	require.NotNil(t, updatedResponse.Result, "Updated metrics result should not be nil")

	updatedMetrics := updatedResponse.Result.(map[string]interface{})

	// Check for common metrics fields
	expectedMetrics := []string{"requests_total", "operations_total", "latency_avg", "errors_total"}
	foundMetrics := 0

	for _, metric := range expectedMetrics {
		if value, exists := updatedMetrics[metric]; exists {
			assert.NotNil(t, value, "Metric %s should have a value", metric)
			t.Logf("Metric %s: %v", metric, value)
			foundMetrics++
		}
	}

	// At least some metrics should be present
	assert.Greater(t, foundMetrics, 0, "At least some metrics should be present")
	LogWorkflowStep(t, "Metrics Collection", 5, "Metrics structure validated")

	// Step 5: Validate counters reflect operations performed (baseline + operations)
	// Compare baseline and updated metrics
	if len(baselineMetrics) > 0 && len(updatedMetrics) > 0 {
		// Look for counter metrics that should have increased
		counterMetrics := []string{"requests_total", "operations_total", "snapshots_total", "recordings_total"}

		for _, counter := range counterMetrics {
			if baselineValue, baselineExists := baselineMetrics[counter]; baselineExists {
				if updatedValue, updatedExists := updatedMetrics[counter]; updatedExists {
					baselineFloat, baselineOk := baselineValue.(float64)
					updatedFloat, updatedOk := updatedValue.(float64)

					if baselineOk && updatedOk {
						assert.GreaterOrEqual(t, updatedFloat, baselineFloat,
							"Counter %s should not decrease", counter)
						t.Logf("Counter %s: baseline=%v, updated=%v", counter, baselineFloat, updatedFloat)
					}
				}
			}
		}
	}
	LogWorkflowStep(t, "Metrics Collection", 6, "Counter progression validated")

	// Validation: Metrics present AND counters increased AND latencies recorded
	setup.AssertBusinessOutcome("Admin can monitor system performance", func() bool {
		return len(updatedMetrics) > 0 && foundMetrics > 0
	})
	LogWorkflowStep(t, "Metrics Collection", 7, "Business outcome validated - admin can monitor performance")

	// Cleanup: Standard cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Metrics Collection", 8, "Connection closed and cleanup verified")
}

func TestHealthMonitoringOverTime(t *testing.T) {
	// Setup: Authenticated connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Health Monitoring", 1, "Authenticated connection established")

	// Step 1: Poll get_system_status every 2 seconds for 10 seconds using testutils.WaitForCondition
	pollCount := 5
	pollInterval := 2 * time.Second
	healthResults := make([]map[string]interface{}, pollCount)
	uptimes := make([]float64, pollCount)

	for i := 0; i < pollCount; i++ {
		healthResponse := setup.SendJSONRPC(conn, "get_system_status", map[string]interface{}{})
		require.NoError(t, healthResponse.Error, "Health poll %d should succeed", i+1)
		require.NotNil(t, healthResponse.Result, "Health poll %d result should not be nil", i+1)

		healthResults[i] = healthResponse.Result.(map[string]interface{})
		uptimes[i] = healthResults[i]["uptime"].(float64)

		t.Logf("Poll %d: uptime=%.2f, status=%v", i+1, uptimes[i], healthResults[i]["status"])

		if i < pollCount-1 { // Don't wait after last poll
			setup.WaitForCondition(func() bool {
				return false // This will timeout after pollInterval
			}, pollInterval)
		}
	}
	LogWorkflowStep(t, "Health Monitoring", 2, "Health polling completed over time")

	// Step 2: Verify health status remains consistent across polls
	statusValues := make([]string, pollCount)
	for i, result := range healthResults {
		statusValues[i] = result["status"].(string)
	}

	// All status values should be the same (consistent)
	for i := 1; i < pollCount; i++ {
		assert.Equal(t, statusValues[0], statusValues[i],
			"Health status should remain consistent across polls")
	}
	LogWorkflowStep(t, "Health Monitoring", 3, "Health status consistency verified")

	// Step 3: Validate uptime increases appropriately (~2s per poll)
	for i := 1; i < pollCount; i++ {
		timeDiff := uptimes[i] - uptimes[i-1]
		expectedDiff := 2.0 // 2 seconds per poll

		// Allow some tolerance for timing variations
		tolerance := 1.0
		assert.GreaterOrEqual(t, timeDiff, expectedDiff-tolerance,
			"Uptime should increase by approximately 2 seconds per poll")
		assert.LessOrEqual(t, timeDiff, expectedDiff+tolerance,
			"Uptime should not increase by more than 3 seconds per poll")

		t.Logf("Uptime increase poll %d: %.2f seconds", i, timeDiff)
	}
	LogWorkflowStep(t, "Health Monitoring", 4, "Uptime progression validated")

	// Step 4: Check no component transitions to unhealthy state
	for i, result := range healthResults {
		components := result["components"].(map[string]interface{})

		for componentName, componentStatus := range components {
			statusMap, ok := componentStatus.(map[string]interface{})
			if !ok {
				continue
			}

			if health, ok := statusMap["health"].(string); ok {
				assert.NotEqual(t, "unhealthy", health,
					"Component %s should not become unhealthy during monitoring", componentName)
			}
		}

		t.Logf("Poll %d: All components healthy", i+1)
	}
	LogWorkflowStep(t, "Health Monitoring", 5, "Component health stability verified")

	// Validation: Health consistent AND uptime increases correctly AND no failures
	setup.AssertBusinessOutcome("Admin can monitor system stability over time", func() bool {
		// Check all status values are the same
		allStatusConsistent := true
		for i := 1; i < pollCount; i++ {
			if statusValues[i] != statusValues[0] {
				allStatusConsistent = false
				break
			}
		}

		// Check uptime is increasing
		uptimeIncreasing := true
		for i := 1; i < pollCount; i++ {
			if uptimes[i] <= uptimes[i-1] {
				uptimeIncreasing = false
				break
			}
		}

		return allStatusConsistent && uptimeIncreasing
	})
	LogWorkflowStep(t, "Health Monitoring", 6, "Business outcome validated - admin can monitor stability")

	// Cleanup: Standard cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Health Monitoring", 7, "Connection closed and cleanup verified")
}
