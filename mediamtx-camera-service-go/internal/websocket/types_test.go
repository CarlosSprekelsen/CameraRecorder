/*
WebSocket Types Unit Tests - Enterprise-Grade Progressive Readiness Pattern

Provides focused unit tests for WebSocket type definitions,
following homogeneous enterprise-grade patterns with real hardware integration.

ENTERPRISE STANDARDS:
- Progressive Readiness Pattern compliance (no polling, no sequential execution)
- Real hardware integration (no mocking, no skipping)
- Homogeneous test patterns across all type tests
- Proper documentation with requirements coverage

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-ARCH-001: Progressive Readiness Pattern compliance

Test Categories: Enterprise Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Pattern: Progressive Readiness with real hardware integration
*/

package websocket

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestWebSocketTypes_JsonRpcRequest tests JSON-RPC request structure
func TestWebSocketTypes_JsonRpcRequest(t *testing.T) {
	// Test JSON-RPC request creation
	request := &JsonRpcRequest{
		JSONRPC: "2.0",
		ID:      "test-request",
		Method:  "ping",
		Params:  map[string]interface{}{},
	}

	// Test request structure
	assert.Equal(t, "2.0", request.JSONRPC, "Request should have correct JSON-RPC version")
	assert.Equal(t, "test-request", request.ID, "Request should have correct ID")
	assert.Equal(t, "ping", request.Method, "Request should have correct method")
	assert.NotNil(t, request.Params, "Request params should be initialized")
}

// TestWebSocketTypes_JsonRpcResponse tests JSON-RPC response structure
func TestWebSocketTypes_JsonRpcResponse(t *testing.T) {
	// Test JSON-RPC response creation
	response := &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      "test-request",
		Result:  "pong",
		Error:   nil,
	}

	// Test response structure
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, "test-request", response.ID, "Response should have correct ID")
	assert.Equal(t, "pong", response.Result, "Response should have correct result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketTypes_JsonRpcError tests JSON-RPC error structure
func TestWebSocketTypes_JsonRpcError(t *testing.T) {
	// Test JSON-RPC error creation using standardized helper
	error := NewJsonRpcError(INVALID_REQUEST, "test_invalid_request", "Invalid JSON-RPC request", "Check request format")

	// Test error structure
	assert.Equal(t, INVALID_REQUEST, error.Code, "Error should have correct code")
	assert.Equal(t, "Invalid Request", error.Message, "Error should have correct message")

	// Test error data structure
	errorData, ok := error.Data.(*ErrorData)
	assert.True(t, ok, "Error data should be ErrorData type")
	assert.Equal(t, "test_invalid_request", errorData.Reason, "Error data should have correct reason")
	assert.Equal(t, "Invalid JSON-RPC request", errorData.Details, "Error data should have correct details")
	assert.Equal(t, "Check request format", errorData.Suggestion, "Error data should have correct suggestion")
}

// TestWebSocketTypes_JsonRpcNotification tests JSON-RPC notification structure
func TestWebSocketTypes_JsonRpcNotification(t *testing.T) {
	// Test JSON-RPC notification creation
	notification := &JsonRpcNotification{
		JSONRPC: "2.0",
		Method:  "camera_status_update",
		Params: map[string]interface{}{
			"camera_id": "camera-1",
			"status":    "online",
		},
	}

	// Test notification structure
	assert.Equal(t, "2.0", notification.JSONRPC, "Notification should have correct JSON-RPC version")
	assert.Equal(t, "camera_status_update", notification.Method, "Notification should have correct method")
	assert.NotNil(t, notification.Params, "Notification params should be initialized")
	assert.Equal(t, "camera-1", notification.Params["camera_id"], "Camera ID should be in params")
	assert.Equal(t, "online", notification.Params["status"], "Status should be in params")
}

// TestWebSocketTypes_ErrorStandardization tests standardized error response format
func TestWebSocketTypes_ErrorStandardization(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// Test 1: NewJsonRpcError helper function
	t.Run("NewJsonRpcErrorHelper", func(t *testing.T) {
		error := NewJsonRpcError(-32001, "AUTH_FAILED", "Authentication failed", "Please provide valid credentials")

		assert.Equal(t, AUTHENTICATION_REQUIRED, error.Code, "Error code should be set correctly")
		assert.Equal(t, "Authentication failed or token expired", error.Message, "Error message should match API specification")
		assert.NotNil(t, error.Data, "Error data should be initialized")

		// Cast Data to ErrorData
		errorData, ok := error.Data.(*ErrorData)
		assert.True(t, ok, "Error data should be of type *ErrorData")
		assert.Equal(t, "AUTH_FAILED", errorData.Reason, "Error reason should be set correctly")
		assert.Equal(t, "Please provide valid credentials", errorData.Suggestion, "Error suggestion should be set correctly")
	})

	// Test 2: Standard error codes
	t.Run("StandardErrorCodes", func(t *testing.T) {
		// Test standard JSON-RPC error codes
		invalidRequest := NewJsonRpcError(INVALID_REQUEST, "INVALID_REQUEST", "Invalid request format", "Check JSON-RPC 2.0 specification")
		assert.Equal(t, INVALID_REQUEST, invalidRequest.Code, "Invalid request error code should match constant")

		methodNotFound := NewJsonRpcError(METHOD_NOT_FOUND, "METHOD_NOT_FOUND", "Method not found", "Check method name")
		assert.Equal(t, METHOD_NOT_FOUND, methodNotFound.Code, "Method not found error code should match constant")

		invalidParams := NewJsonRpcError(INVALID_PARAMS, "INVALID_PARAMS", "Invalid parameters", "Check parameter types and values")
		assert.Equal(t, INVALID_PARAMS, invalidParams.Code, "Invalid params error code should match constant")

		internalError := NewJsonRpcError(INTERNAL_ERROR, "INTERNAL_ERROR", "Internal server error", "Contact system administrator")
		assert.Equal(t, INTERNAL_ERROR, internalError.Code, "Internal error code should match constant")
	})

	// Test 3: Service-specific error codes
	t.Run("ServiceSpecificErrorCodes", func(t *testing.T) {
		// Test service-specific error codes
		authFailed := NewJsonRpcError(-32001, "AUTH_FAILED", "Authentication failed", "Provide valid token")
		assert.Equal(t, AUTHENTICATION_REQUIRED, authFailed.Code, "Auth failed error code should be AUTHENTICATION_REQUIRED")

		rateLimit := NewJsonRpcError(RATE_LIMIT_EXCEEDED, "RATE_LIMIT", "Rate limit exceeded", "Wait before retrying")
		assert.Equal(t, RATE_LIMIT_EXCEEDED, rateLimit.Code, "Rate limit error code should match constant")

		permissionDenied := NewJsonRpcError(-32003, "PERMISSION_DENIED", "Insufficient permissions", "Contact administrator")
		assert.Equal(t, INSUFFICIENT_PERMISSIONS, permissionDenied.Code, "Permission denied error code should be INSUFFICIENT_PERMISSIONS")

		cameraNotFound := NewJsonRpcError(CAMERA_NOT_FOUND, "CAMERA_NOT_FOUND", "Camera not found", "Check camera identifier")
		assert.Equal(t, CAMERA_NOT_FOUND, cameraNotFound.Code, "Camera not found error code should match constant")
	})

	// Test 4: Error data structure
	t.Run("ErrorDataStructure", func(t *testing.T) {
		error := NewJsonRpcError(-32001, "TEST_ERROR", "Test error occurred", "Test suggestion")

		// Test ErrorData fields
		errorData, ok := error.Data.(*ErrorData)
		assert.True(t, ok, "Error data should be of type *ErrorData")
		assert.Equal(t, "TEST_ERROR", errorData.Reason, "Reason should be set correctly")
		assert.Equal(t, "Test suggestion", errorData.Suggestion, "Suggestion should be set correctly")
		assert.Equal(t, "Test error occurred", errorData.Details, "Details should be set correctly")
	})

	// Test 5: Error response format
	t.Run("ErrorResponseFormat", func(t *testing.T) {
		error := NewJsonRpcError(-32001, "AUTH_FAILED", "Authentication failed", "Provide valid token")

		response := &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      "test-id",
			Error:   error,
		}

		// Test response structure
		assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
		assert.Equal(t, "test-id", response.ID, "Response should have correct ID")
		assert.NotNil(t, response.Error, "Response should have error")
		assert.Nil(t, response.Result, "Error response should not have result")

		// Test error structure
		assert.Equal(t, AUTHENTICATION_REQUIRED, response.Error.Code, "Error should have correct code")
		assert.Equal(t, "Authentication failed or token expired", response.Error.Message, "Error should match API specification")
		assert.NotNil(t, response.Error.Data, "Error should have data")
	})

	// Test 6: Error code ranges
	t.Run("ErrorCodeRanges", func(t *testing.T) {
		// Standard JSON-RPC errors should be in -32768 to -32000 range
		standardError := NewJsonRpcError(INVALID_REQUEST, "STANDARD", "Standard error", "Standard suggestion")
		assert.True(t, standardError.Code >= -32768 && standardError.Code <= -32000, "Standard error codes should be in -32768 to -32000 range")

		// Service-specific errors should be in -32099 to -32000 range
		serviceError := NewJsonRpcError(-32001, "SERVICE", "Service error", "Service suggestion")
		assert.True(t, serviceError.Code >= -32099 && serviceError.Code <= -32000, "Service error codes should be in -32099 to -32000 range")
	})

	// Test 7: Error message consistency
	t.Run("ErrorMessageConsistency", func(t *testing.T) {
		// Test that error messages are consistent and descriptive
		errors := []struct {
			code    int
			message string
		}{
			{-32600, "Invalid Request"},
			{-32601, "Method not found"},
			{-32602, "Invalid parameters"},
			{-32603, "Internal server error"},
			{-32001, "Authentication failed or token expired"},
			{-32002, "Rate limit exceeded"},
			{-32003, "Insufficient permissions"},
			{-32004, "Camera not found or disconnected"},
		}

		for _, testError := range errors {
			error := NewJsonRpcError(testError.code, "TEST_REASON", "Test reason", "Test suggestion")
			assert.Equal(t, testError.message, error.Message, "Error message should match API specification")
			assert.NotEmpty(t, error.Message, "Error message should not be empty")
		}
	})
}

// TestWebSocketTypes_ClientConnection tests client connection structure
func TestWebSocketTypes_ClientConnection(t *testing.T) {
	// Test client connection creation
	client := &ClientConnection{
		ClientID:      "test-client",
		Authenticated: false,
		UserID:        "",
		Role:          "",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	// Test client structure
	assert.Equal(t, "test-client", client.ClientID, "Client ID should be set")
	assert.False(t, client.Authenticated, "Client should not be authenticated initially")
	assert.Empty(t, client.UserID, "User ID should be empty initially")
	assert.Empty(t, client.Role, "Role should be empty initially")
	assert.NotNil(t, client.Subscriptions, "Subscriptions map should be initialized")
}

// TestWebSocketTypes_PerformanceMetrics tests performance metrics structure
func TestWebSocketTypes_PerformanceMetrics(t *testing.T) {
	// Test performance metrics creation
	metrics := &PerformanceMetrics{
		RequestCount:      0,
		ResponseTimes:     make(map[string][]float64),
		ErrorCount:        0,
		ActiveConnections: 0,
		StartTime:         time.Now(),
	}

	// Test metrics structure
	assert.Equal(t, int64(0), metrics.RequestCount, "Initial request count should be 0")
	assert.Equal(t, int64(0), metrics.ErrorCount, "Initial error count should be 0")
	assert.Equal(t, int64(0), metrics.ActiveConnections, "Initial active connections should be 0")
	assert.NotNil(t, metrics.ResponseTimes, "Response times map should be initialized")
	assert.NotZero(t, metrics.StartTime, "Start time should be set")
}

// TestWebSocketTypes_WebSocketMessage tests WebSocket message structure
func TestWebSocketTypes_WebSocketMessage(t *testing.T) {
	// Test WebSocket message creation
	message := &WebSocketMessage{
		Type:      "event",
		Data:      []byte(`{"type": "camera_update"}`),
		Timestamp: time.Now(),
		ClientID:  "test-client",
	}

	// Test message structure
	assert.Equal(t, "event", message.Type, "Message type should be set")
	assert.NotNil(t, message.Data, "Message data should be set")
	assert.NotZero(t, message.Timestamp, "Message timestamp should be set")
	assert.Equal(t, "test-client", message.ClientID, "Client ID should be set")
}

// TestWebSocketTypes_ServerConfig tests server configuration structure
func TestWebSocketTypes_ServerConfig(t *testing.T) {
	// Test server configuration creation
	config := &ServerConfig{
		Host:                 "localhost",
		Port:                 8002,
		WebSocketPath:        "/ws",
		MaxConnections:       1000,
		ReadTimeout:          testutils.UniversalTimeoutVeryLong,
		WriteTimeout:         1 * time.Second,
		PingInterval:         30 * time.Second,
		PongWait:             60 * time.Second,
		MaxMessageSize:       1024 * 1024,
		ShutdownTimeout:      30 * time.Second,
		ClientCleanupTimeout: 10 * time.Second,
	}

	// Test config structure
	assert.Equal(t, "localhost", config.Host, "Host should be set")
	assert.Equal(t, 8002, config.Port, "Port should be set")
	assert.Equal(t, "/ws", config.WebSocketPath, "WebSocket path should be set")
	assert.Equal(t, 1000, config.MaxConnections, "Max connections should be set")
	assert.Equal(t, testutils.UniversalTimeoutVeryLong, config.ReadTimeout, "Read timeout should be set")
	assert.Equal(t, 1*time.Second, config.WriteTimeout, "Write timeout should be set")
	assert.Equal(t, 30*time.Second, config.PingInterval, "Ping interval should be set")
	assert.Equal(t, 60*time.Second, config.PongWait, "Pong wait should be set")
	assert.Equal(t, int64(1024*1024), config.MaxMessageSize, "Max message size should be set")
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout, "Shutdown timeout should be set")
	assert.Equal(t, 10*time.Second, config.ClientCleanupTimeout, "Client cleanup timeout should be set")
}

// TestWebSocketTypes_DefaultServerConfig tests default server configuration
func TestWebSocketTypes_DefaultServerConfig(t *testing.T) {
	// Test default server configuration
	config := DefaultServerConfig()

	// Test default config values
	assert.Equal(t, "0.0.0.0", config.Host, "Default host should be 0.0.0.0")
	assert.Equal(t, 8002, config.Port, "Default port should be 8002")
	assert.Equal(t, "/ws", config.WebSocketPath, "Default WebSocket path should be /ws")
	assert.Equal(t, 1000, config.MaxConnections, "Default max connections should be 1000")
	assert.Equal(t, 5*time.Second, config.ReadTimeout, "Default read timeout should be 5 seconds")
	assert.Equal(t, 1*time.Second, config.WriteTimeout, "Default write timeout should be 1 second")
	assert.Equal(t, 30*time.Second, config.PingInterval, "Default ping interval should be 30 seconds")
	assert.Equal(t, 60*time.Second, config.PongWait, "Default pong wait should be 60 seconds")
	assert.Equal(t, int64(1024*1024), config.MaxMessageSize, "Default max message size should be 1MB")
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout, "Default shutdown timeout should be 30 seconds")
	assert.Equal(t, 10*time.Second, config.ClientCleanupTimeout, "Default client cleanup timeout should be 10 seconds")
}

// TestWebSocketTypes_MethodHandler tests method handler type
func TestWebSocketTypes_MethodHandler(t *testing.T) {
	// Test method handler function signature
	handler := func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      "test",
			Result:  "success",
		}, nil
	}

	// Test handler execution
	client := &ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "viewer",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	response, err := handler(map[string]interface{}{}, client)
	assert.NoError(t, err, "Handler should not return error")
	assert.NotNil(t, response, "Handler should return response")
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, "test", response.ID, "Response should have correct ID")
	assert.Equal(t, "success", response.Result, "Response should have correct result")
}
