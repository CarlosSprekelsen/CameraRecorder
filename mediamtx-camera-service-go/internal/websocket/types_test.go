/*
WebSocket Types Unit Tests

Provides focused unit tests for WebSocket type definitions,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"testing"
	"time"

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
	assert.Equal(t, -32600, error.Code, "Error should have correct code")
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
		ReadTimeout:          5 * time.Second,
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
	assert.Equal(t, 5*time.Second, config.ReadTimeout, "Read timeout should be set")
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
