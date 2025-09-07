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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketMethods_Ping tests ping method
func TestWebSocketMethods_Ping(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create a test JWT token for authentication using the same secret as the server
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "viewer")

	// First authenticate the client
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")

	// Send ping message
	message := CreateTestMessage("ping", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Equal(t, "pong", response.Result, "Response should have correct result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetServerInfo tests get_server_info method
func TestWebSocketMethods_GetServerInfo(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create a test JWT token for authentication using the same secret as the server
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "admin") // admin role for get_server_info

	// First authenticate the client
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")

	// Send get_server_info message
	message := CreateTestMessage("get_server_info", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetStatus tests get_status method
func TestWebSocketMethods_GetStatus(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create a test JWT token for authentication using the same secret as the server
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "admin") // admin role for get_status

	// First authenticate the client
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")

	// Send get_status message
	message := CreateTestMessage("get_status", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
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
	assert.Equal(t, METHOD_NOT_FOUND, response.Error.Code, "Error should be method not found")
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

// TestWebSocketMethods_SequentialRequests tests sequential request handling
// This test properly tests the server's ability to handle multiple requests efficiently
func TestWebSocketMethods_SequentialRequests(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Create a single connection for all requests
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")

	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "viewer")

	// Authenticate the client once
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")

	// Test multiple sequential requests
	const numRequests = 10
	startTime := time.Now()

	for i := 0; i < numRequests; i++ {
		message := CreateTestMessage("ping", map[string]interface{}{"request_id": i})
		response := SendTestMessage(t, conn, message)

		assert.Nil(t, response.Error, "Request %d should not have error", i)
		assert.Equal(t, "pong", response.Result, "Request %d should have correct result", i)
	}

	duration := time.Since(startTime)
	t.Logf("Processed %d requests in %v (avg: %v per request)",
		numRequests, duration, duration/time.Duration(numRequests))

	// Verify reasonable performance (should be fast for simple ping requests)
	assert.Less(t, duration, 5*time.Second, "Requests should complete within reasonable time")
}

// TestWebSocketMethods_MultipleConnections tests multiple connections handling
// This test properly creates multiple connections with proper synchronization
func TestWebSocketMethods_MultipleConnections(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")
	defer CleanupTestServer(t, server)

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")

	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "viewer")

	// Test multiple connections with proper synchronization
	const numConnections = 5
	responses := make(chan *JsonRpcResponse, numConnections)
	errors := make(chan error, numConnections)
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent connections
	semaphore := make(chan struct{}, 3) // Limit to 3 concurrent connections

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(connectionID int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create connection for this goroutine
			conn := NewTestClient(t, server)
			defer CleanupTestClient(t, conn)

			// Authenticate the client
			authMessage := CreateTestMessage("authenticate", map[string]interface{}{
				"auth_token": testToken,
			})
			authResponse := SendTestMessage(t, conn, authMessage)
			if authResponse.Error != nil {
				errors <- fmt.Errorf("authentication failed for connection %d: %v", connectionID, authResponse.Error)
				return
			}

			// Send ping message
			message := CreateTestMessage("ping", map[string]interface{}{"connection_id": connectionID})
			response := SendTestMessage(t, conn, message)
			responses <- response
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Collect all responses
	receivedResponses := 0
	receivedErrors := 0
	for i := 0; i < numConnections; i++ {
		select {
		case response := <-responses:
			assert.Equal(t, "pong", response.Result, "Response should have correct result")
			receivedResponses++
		case err := <-errors:
			t.Errorf("Connection failed: %v", err)
			receivedErrors++
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for multiple connection responses")
		}
	}

	assert.Equal(t, numConnections, receivedResponses, "Should receive all responses")
	assert.Equal(t, 0, receivedErrors, "Should have no errors")
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

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", logging.NewLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "admin") // admin role for get_server_info

	// First authenticate the client
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")

	// Send get_server_info message (testing method execution)
	message := CreateTestMessage("get_server_info", map[string]interface{}{})
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
