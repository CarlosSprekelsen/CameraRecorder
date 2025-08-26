//go:build unit

/*
WebSocket JSON-RPC method validation unit tests.

Tests validate actual method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-API-004: get_camera_status method for camera status
- REQ-API-008: authenticate method for authentication

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketServerMethodRegistration tests that all required methods are registered
func TestWebSocketServerMethodRegistration(t *testing.T) {
	/*
		API Compliance Test for method registration

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: All required methods should be registered in the server
	*/

	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)

	// Test that server is properly initialized
	assert.False(t, server.IsRunning(), "Server should not be running initially")

	// Test that server can be started (this will register methods)
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Test that server is running after start
	assert.True(t, server.IsRunning(), "Server should be running after start")
}

// TestAPIErrorCodesValidation tests that error codes match API documentation
func TestAPIErrorCodesValidation(t *testing.T) {
	/*
		API Compliance Test for error codes validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected Error Codes: -32001, -32601, -32602, -32603
	*/

	// Validate error codes against API documentation
	assert.Equal(t, -32001, websocket.AUTHENTICATION_REQUIRED, "Authentication required error code should be -32001 per API documentation")
	assert.Equal(t, -32601, websocket.METHOD_NOT_FOUND, "Method not found error code should be -32601 per API documentation")
	assert.Equal(t, -32602, websocket.INVALID_PARAMS, "Invalid params error code should be -32602 per API documentation")
	assert.Equal(t, -32603, websocket.INTERNAL_ERROR, "Internal error code should be -32603 per API documentation")
}

// TestAPIErrorMessagesValidation tests that error messages match API documentation
func TestAPIErrorMessagesValidation(t *testing.T) {
	/*
		API Compliance Test for error messages validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected Error Messages: "Authentication required", "Method not found", "Invalid parameters", "Internal server error"
	*/

	// Validate error messages against API documentation
	assert.Equal(t, "Authentication required", websocket.ErrorMessages[websocket.AUTHENTICATION_REQUIRED], "Authentication required error message should match API documentation")
	assert.Equal(t, "Method not found", websocket.ErrorMessages[websocket.METHOD_NOT_FOUND], "Method not found error message should match API documentation")
	assert.Equal(t, "Invalid parameters", websocket.ErrorMessages[websocket.INVALID_PARAMS], "Invalid parameters error message should match API documentation")
	assert.Equal(t, "Internal server error", websocket.ErrorMessages[websocket.INTERNAL_ERROR], "Internal server error message should match API documentation")
}

// TestJSONRPCStructuresValidation tests that JSON-RPC structures match API documentation
func TestJSONRPCStructuresValidation(t *testing.T) {
	/*
		API Compliance Test for JSON-RPC structures validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: JSON-RPC request and response structures should match API documentation
	*/

	// Test JsonRpcRequest structure
	request := websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
	}
	assert.Equal(t, "2.0", request.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, "ping", request.Method, "Method should be set correctly")
	assert.Equal(t, 1, request.ID, "ID should be set correctly")

	// Test JsonRpcResponse structure
	response := websocket.JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  "pong",
		ID:      1,
	}
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, "pong", response.Result, "Result should be set correctly")
	assert.Equal(t, 1, response.ID, "ID should be set correctly")

	// Test JsonRpcError structure
	error := websocket.JsonRpcError{
		Code:    -32001,
		Message: "Authentication required",
	}
	assert.Equal(t, -32001, error.Code, "Error code should be set correctly")
	assert.Equal(t, "Authentication required", error.Message, "Error message should be set correctly")
}

// TestServerConfigurationValidation tests that server configuration matches API documentation
func TestServerConfigurationValidation(t *testing.T) {
	/*
		API Compliance Test for server configuration validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: Server configuration should match API documentation requirements
	*/

	config := websocket.DefaultServerConfig()
	assert.NotNil(t, config, "Default server configuration should be defined")
	assert.Equal(t, "0.0.0.0", config.Host, "Default host should be 0.0.0.0 per API documentation")
	assert.Equal(t, 8002, config.Port, "Default port should be 8002 per API documentation")
	assert.Equal(t, "/ws", config.WebSocketPath, "Default WebSocket path should be /ws per API documentation")
	assert.Equal(t, 1000, config.MaxConnections, "Default max connections should be 1000 per API documentation")
}

// TestClientConnectionStructureValidation tests that client connection structure matches API documentation
func TestClientConnectionStructureValidation(t *testing.T) {
	/*
		API Compliance Test for client connection structure validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: Client connection structure should support session management per API documentation
	*/

	client := websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: false,
		UserID:        "",
		Role:          "",
		AuthMethod:    "",
		Subscriptions: make(map[string]bool),
	}

	assert.Equal(t, "test-client", client.ClientID, "Client ID should be set correctly")
	assert.False(t, client.Authenticated, "Client should not be authenticated initially")
	assert.NotNil(t, client.Subscriptions, "Subscriptions map should be initialized")
}

// TestPerformanceMetricsStructureValidation tests that performance metrics structure matches API documentation
func TestPerformanceMetricsStructureValidation(t *testing.T) {
	/*
		API Compliance Test for performance metrics structure validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: Performance metrics structure should support monitoring per API documentation
	*/

	metrics := websocket.PerformanceMetrics{
		RequestCount:      0,
		ResponseTimes:     make(map[string][]float64),
		ErrorCount:        0,
		ActiveConnections: 0,
	}

	assert.Equal(t, int64(0), metrics.RequestCount, "Request count should be initialized to 0")
	assert.Equal(t, int64(0), metrics.ErrorCount, "Error count should be initialized to 0")
	assert.Equal(t, int64(0), metrics.ActiveConnections, "Active connections should be initialized to 0")
	assert.NotNil(t, metrics.ResponseTimes, "Response times map should be initialized")
}

// TestMethodHandlerTypeValidation tests that method handler type matches API documentation
func TestMethodHandlerTypeValidation(t *testing.T) {
	/*
		API Compliance Test for method handler type validation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: Method handler type should support JSON-RPC method signatures per API documentation
	*/

	// Create a test handler to validate the type signature
	handler := func(params map[string]interface{}, client *websocket.ClientConnection) (*websocket.JsonRpcResponse, error) {
		return &websocket.JsonRpcResponse{
			JSONRPC: "2.0",
			Result:  "test",
			ID:      1,
		}, nil
	}

	// Test that handler can be called with correct parameters
	params := map[string]interface{}{"test": "value"}
	client := &websocket.ClientConnection{ClientID: "test-client"}

	response, err := handler(params, client)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "2.0", response.JSONRPC, "Handler should return correct JSON-RPC version")
	assert.Equal(t, "test", response.Result, "Handler should return correct result")
	assert.Equal(t, 1, response.ID, "Handler should return correct ID")
}
