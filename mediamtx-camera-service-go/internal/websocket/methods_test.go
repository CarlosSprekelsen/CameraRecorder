/*
WebSocket Methods Unit Tests

Provides focused unit tests for WebSocket method handling,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketMethods_Ping tests ping method
func TestWebSocketMethods_Ping(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send ping message
	message := CreateTestMessage("ping", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Equal(t, "pong", response.Result, "Response should have correct result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_Echo tests echo method
func TestWebSocketMethods_Echo(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send echo message with parameters
	testParams := map[string]interface{}{
		"message": "hello world",
		"number":  42,
		"boolean": true,
	}
	message := CreateTestMessage("echo", testParams)
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Equal(t, testParams, response.Result, "Response should echo parameters")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_Error tests error method
func TestWebSocketMethods_Error(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send error message
	message := CreateTestMessage("error", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test error response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Result, "Response should not have result")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Equal(t, INTERNAL_ERROR, response.Error.Code, "Error should be internal error")
	assert.Equal(t, "Test error", response.Error.Message, "Error should have correct message")
}

// TestWebSocketMethods_InvalidJSON tests invalid JSON handling
func TestWebSocketMethods_InvalidJSON(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send invalid JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	require.NoError(t, err, "Should send invalid JSON")

	// Read response
	var response JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err, "Should read error response")

	// Test error response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Nil(t, response.Result, "Response should not have result")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Equal(t, INVALID_REQUEST, response.Error.Code, "Error should be invalid request")
}

// TestWebSocketMethods_MissingMethod tests missing method handling
func TestWebSocketMethods_MissingMethod(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send message without method
	message := &JsonRpcRequest{
		JSONRPC: "2.0",
		ID:      "test-request",
		// Method is missing
		Params: map[string]interface{}{},
	}
	response := SendTestMessage(t, conn, message)

	// Test error response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Result, "Response should not have result")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Equal(t, INVALID_REQUEST, response.Error.Code, "Error should be invalid request")
}

// TestWebSocketMethods_MissingJSONRPC tests missing JSON-RPC version
func TestWebSocketMethods_MissingJSONRPC(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send message without JSON-RPC version
	message := &JsonRpcRequest{
		// JSONRPC is missing
		Method: "ping",
		ID:     "test-request",
		Params: map[string]interface{}{},
	}
	response := SendTestMessage(t, conn, message)

	// Test error response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Result, "Response should not have result")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Equal(t, INVALID_REQUEST, response.Error.Code, "Error should be invalid request")
}

// TestWebSocketMethods_ConcurrentRequests tests concurrent request handling
func TestWebSocketMethods_ConcurrentRequests(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send multiple concurrent requests
	const numRequests = 10
	responses := make(chan *JsonRpcResponse, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			message := CreateTestMessage("ping", map[string]interface{}{"request_id": requestID})
			response := SendTestMessage(t, conn, message)
			responses <- response
		}(i)
	}

	// Collect all responses
	receivedResponses := 0
	for i := 0; i < numRequests; i++ {
		select {
		case response := <-responses:
			assert.Equal(t, "pong", response.Result, "Response should have correct result")
			receivedResponses++
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent responses")
		}
	}

	assert.Equal(t, numRequests, receivedResponses, "Should receive all responses")
}

// TestWebSocketMethods_LargePayload tests large payload handling
func TestWebSocketMethods_LargePayload(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create large payload
	largeData := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = "This is a large string to test payload handling"
	}

	// Send echo message with large payload
	message := CreateTestMessage("echo", map[string]interface{}{"large_data": largeData})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_Timeout tests request timeout handling
func TestWebSocketMethods_Timeout(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Register a method that takes time
	server.registerMethod("slow_method", func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		time.Sleep(2 * time.Second)
		return CreateTestResponse("test-id", "slow_result"), nil
	}, "1.0")

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	// Send slow method message
	CreateTestMessage("slow_method", map[string]interface{}{})

	// This should timeout
	var response JsonRpcResponse
	err = conn.ReadJSON(&response)
	assert.Error(t, err, "Should timeout on slow method")
}
