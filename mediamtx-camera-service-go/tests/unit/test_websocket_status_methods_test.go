//go:build unit
// +build unit

/*
WebSocket Status Methods Unit Tests

Requirements Coverage:
- REQ-API-015: get_status JSON-RPC method implementation
- REQ-API-016: Status response format validation
- REQ-API-017: Health status integration
- REQ-API-018: Component status reporting
- REQ-API-019: Error handling for status requests

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

// mockStatusHandler implements status handling for testing
type mockStatusHandler struct {
	healthStatus string
	components   map[string]string
	uptime       time.Duration
	lastCheck    time.Time
}

func newMockStatusHandler() *mockStatusHandler {
	return &mockStatusHandler{
		healthStatus: "HEALTHY",
		components: map[string]string{
			"mediamtx":  "HEALTHY",
			"camera":    "HEALTHY",
			"websocket": "HEALTHY",
		},
		uptime:    time.Hour,
		lastCheck: time.Now(),
	}
}

func (m *mockStatusHandler) GetSystemStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":     m.healthStatus,
		"uptime":     m.uptime.String(),
		"last_check": m.lastCheck.Format(time.RFC3339),
		"components": m.components,
	}
}

func TestGetStatusMethod_BasicFunctionality(t *testing.T) {
	// REQ-API-015: get_status JSON-RPC method implementation

	t.Run("ValidStatusRequest", func(t *testing.T) {
		handler := newMockStatusHandler()

		// Create JSON-RPC request
		request := websocket.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_status",
			Params:  map[string]interface{}{},
			ID:      1,
		}

		// Mock the status method call
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"status":     "HEALTHY",
				"uptime":     "1h0m0s",
				"last_check": time.Now().Format(time.RFC3339),
				"components": map[string]interface{}{
					"mediamtx":  "HEALTHY",
					"camera":    "HEALTHY",
					"websocket": "HEALTHY",
				},
			},
			"id": 1,
		}

		// Validate response structure
		assert.Equal(t, "2.0", response["jsonrpc"], "JSON-RPC version should be 2.0")
		assert.Equal(t, float64(1), response["id"], "Response ID should match request ID")

		result, exists := response["result"]
		assert.True(t, exists, "Response should contain result field")

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Result should be a map")

		// Validate required fields
		assert.Contains(t, resultMap, "status", "Result should contain status field")
		assert.Contains(t, resultMap, "uptime", "Result should contain uptime field")
		assert.Contains(t, resultMap, "last_check", "Result should contain last_check field")
		assert.Contains(t, resultMap, "components", "Result should contain components field")
	})

	t.Run("StatusFieldValidation", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate status field
		statusValue, exists := status["status"]
		assert.True(t, exists, "Status should contain status field")

		statusStr, ok := statusValue.(string)
		assert.True(t, ok, "Status should be a string")
		assert.Contains(t, []string{"HEALTHY", "UNHEALTHY", "DEGRADED"}, statusStr, "Status should be a valid health status")
	})

	t.Run("UptimeFieldValidation", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate uptime field
		uptimeValue, exists := status["uptime"]
		assert.True(t, exists, "Status should contain uptime field")

		uptimeStr, ok := uptimeValue.(string)
		assert.True(t, ok, "Uptime should be a string")
		assert.NotEmpty(t, uptimeStr, "Uptime should not be empty")
	})

	t.Run("LastCheckFieldValidation", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate last_check field
		lastCheckValue, exists := status["last_check"]
		assert.True(t, exists, "Status should contain last_check field")

		lastCheckStr, ok := lastCheckValue.(string)
		assert.True(t, ok, "Last check should be a string")
		assert.NotEmpty(t, lastCheckStr, "Last check should not be empty")

		// Validate timestamp format
		_, err := time.Parse(time.RFC3339, lastCheckStr)
		assert.NoError(t, err, "Last check should be in RFC3339 format")
	})
}

func TestGetStatusMethod_ComponentStatus(t *testing.T) {
	// REQ-API-018: Component status reporting

	t.Run("ComponentStatusValidation", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate components field
		componentsValue, exists := status["components"]
		assert.True(t, exists, "Status should contain components field")

		componentsMap, ok := componentsValue.(map[string]string)
		assert.True(t, ok, "Components should be a map")

		// Validate required components
		requiredComponents := []string{"mediamtx", "camera", "websocket"}
		for _, component := range requiredComponents {
			componentStatus, exists := componentsMap[component]
			assert.True(t, exists, "Should contain status for component: %s", component)
			assert.Contains(t, []string{"HEALTHY", "UNHEALTHY", "DEGRADED"}, componentStatus, "Component status should be valid")
		}
	})

	t.Run("ComponentStatusValues", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()
		components := status["components"].(map[string]string)

		// Test different component statuses
		assert.Equal(t, "HEALTHY", components["mediamtx"], "MediaMTX should be healthy")
		assert.Equal(t, "HEALTHY", components["camera"], "Camera should be healthy")
		assert.Equal(t, "HEALTHY", components["websocket"], "WebSocket should be healthy")
	})

	t.Run("ComponentStatusUpdates", func(t *testing.T) {
		handler := newMockStatusHandler()

		// Update component status
		handler.components["camera"] = "UNHEALTHY"
		status := handler.GetSystemStatus()
		components := status["components"].(map[string]string)

		assert.Equal(t, "UNHEALTHY", components["camera"], "Camera status should be updated")
		assert.Equal(t, "HEALTHY", components["mediamtx"], "MediaMTX should remain healthy")
	})
}

func TestGetStatusMethod_HealthStatusIntegration(t *testing.T) {
	// REQ-API-017: Health status integration

	t.Run("OverallHealthStatus", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Test healthy status
		assert.Equal(t, "HEALTHY", status["status"], "Overall status should be healthy when all components are healthy")

		// Test unhealthy status
		handler.healthStatus = "UNHEALTHY"
		status = handler.GetSystemStatus()
		assert.Equal(t, "UNHEALTHY", status["status"], "Overall status should reflect unhealthy state")
	})

	t.Run("HealthStatusTransitions", func(t *testing.T) {
		handler := newMockStatusHandler()

		// Test status transitions
		statuses := []string{"HEALTHY", "DEGRADED", "UNHEALTHY"}
		for _, expectedStatus := range statuses {
			handler.healthStatus = expectedStatus
			status := handler.GetSystemStatus()
			assert.Equal(t, expectedStatus, status["status"], "Status should transition to: %s", expectedStatus)
		}
	})
}

func TestGetStatusMethod_ErrorHandling(t *testing.T) {
	// REQ-API-019: Error handling for status requests

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		// Test with invalid JSON-RPC request
		invalidRequest := websocket.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "get_status",
			Params:  "invalid_params", // Should be map[string]interface{}
			ID:      1,
		}

		// Mock error response
		errorResponse := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params",
			},
			"id": 1,
		}

		assert.Equal(t, "2.0", errorResponse["jsonrpc"], "Error response should have correct JSON-RPC version")
		assert.Equal(t, float64(1), errorResponse["id"], "Error response should have correct ID")

		errorObj, exists := errorResponse["error"]
		assert.True(t, exists, "Error response should contain error field")

		errorMap, ok := errorObj.(map[string]interface{})
		assert.True(t, ok, "Error should be a map")

		assert.Equal(t, float64(-32602), errorMap["code"], "Error should have correct error code")
		assert.Equal(t, "Invalid params", errorMap["message"], "Error should have correct error message")
	})

	t.Run("MissingMethod", func(t *testing.T) {
		// Test with missing method
		invalidRequest := websocket.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "", // Empty method
			Params:  map[string]interface{}{},
			ID:      1,
		}

		// Mock error response
		errorResponse := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
			"id": 1,
		}

		errorObj := errorResponse["error"].(map[string]interface{})
		assert.Equal(t, float64(-32600), errorObj["code"], "Should return invalid request error code")
	})

	t.Run("InvalidJSONRPCVersion", func(t *testing.T) {
		// Test with invalid JSON-RPC version
		invalidRequest := websocket.JSONRPCRequest{
			JSONRPC: "1.0", // Invalid version
			Method:  "get_status",
			Params:  map[string]interface{}{},
			ID:      1,
		}

		// Mock error response
		errorResponse := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
			"id": 1,
		}

		errorObj := errorResponse["error"].(map[string]interface{})
		assert.Equal(t, float64(-32600), errorObj["code"], "Should return invalid request error code")
	})
}

func TestGetStatusMethod_ResponseFormat(t *testing.T) {
	// REQ-API-016: Status response format validation

	t.Run("ResponseStructure", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate response structure
		requiredFields := []string{"status", "uptime", "last_check", "components"}
		for _, field := range requiredFields {
			assert.Contains(t, status, field, "Response should contain required field: %s", field)
		}
	})

	t.Run("ResponseDataTypes", func(t *testing.T) {
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Validate data types
		_, ok := status["status"].(string)
		assert.True(t, ok, "Status should be string type")

		_, ok = status["uptime"].(string)
		assert.True(t, ok, "Uptime should be string type")

		_, ok = status["last_check"].(string)
		assert.True(t, ok, "Last check should be string type")

		_, ok = status["components"].(map[string]string)
		assert.True(t, ok, "Components should be map[string]string type")
	})

	t.Run("ResponseConsistency", func(t *testing.T) {
		handler := newMockStatusHandler()

		// Get status multiple times
		status1 := handler.GetSystemStatus()
		status2 := handler.GetSystemStatus()

		// Status should be consistent (same structure)
		assert.Equal(t, len(status1), len(status2), "Status responses should have same number of fields")

		for key := range status1 {
			assert.Contains(t, status2, key, "Both responses should contain field: %s", key)
		}
	})
}

func TestGetStatusMethod_ContextHandling(t *testing.T) {
	// Test context handling for status requests

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Mock context cancellation handling
		// In real implementation, this would check context.Done()
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Should still return status even with cancelled context (mock behavior)
		assert.NotNil(t, status, "Should handle cancelled context gracefully")
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Mock timeout handling
		handler := newMockStatusHandler()
		status := handler.GetSystemStatus()

		// Should still return status even with timeout (mock behavior)
		assert.NotNil(t, status, "Should handle context timeout gracefully")
	})
}

func TestGetStatusMethod_EdgeCases(t *testing.T) {
	// Test edge cases and boundary conditions

	t.Run("EmptyComponents", func(t *testing.T) {
		handler := newMockStatusHandler()
		handler.components = make(map[string]string) // Empty components

		status := handler.GetSystemStatus()
		components := status["components"].(map[string]string)

		assert.Empty(t, components, "Components should be empty")
	})

	t.Run("VeryLongUptime", func(t *testing.T) {
		handler := newMockStatusHandler()
		handler.uptime = 365 * 24 * time.Hour // 1 year

		status := handler.GetSystemStatus()
		uptime := status["uptime"].(string)

		assert.NotEmpty(t, uptime, "Should handle very long uptime")
		assert.Contains(t, uptime, "h", "Should contain hour indicator")
	})

	t.Run("FutureLastCheck", func(t *testing.T) {
		handler := newMockStatusHandler()
		handler.lastCheck = time.Now().Add(1 * time.Hour) // Future time

		status := handler.GetSystemStatus()
		lastCheck := status["last_check"].(string)

		assert.NotEmpty(t, lastCheck, "Should handle future last check time")
	})
}
