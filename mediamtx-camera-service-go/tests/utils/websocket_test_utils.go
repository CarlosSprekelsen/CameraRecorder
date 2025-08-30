//go:build unit || integration || performance || security

/*
WebSocket Testing Utilities - REUSABLE ACROSS TEST CATEGORIES

This file provides reusable WebSocket testing utilities for unit, integration, performance,
and security tests that need to test real WebSocket connections.

COMMON PATTERN USAGE:
Instead of duplicating WebSocket connection code in every test:
   conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8002/ws", nil)

Use the shared utilities:
   client := utils.NewWebSocketTestClient(t, server, jwtHandler)
   response := client.SendRequest(request)

ANTI-PATTERNS TO AVOID:
- DON'T duplicate WebSocket connection code across test files
- DON'T hardcode WebSocket URLs in tests
- DON'T create connection setup in every integration test
- DON'T bypass the WebSocket layer in integration tests

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Authentication and authorization with JWT tokens
- REQ-ERROR-001: WebSocket server shall handle MediaMTX connection failures gracefully
- REQ-ERROR-002: WebSocket server shall handle authentication failures gracefully
- REQ-ERROR-003: WebSocket server shall handle invalid JSON-RPC requests gracefully

Test Categories: Unit/Integration/Performance/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package utils

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// WebSocketTestClient provides a reusable WebSocket client for testing
// This should be used in integration, performance, and security tests
type WebSocketTestClient struct {
	t          *testing.T
	conn       *gorilla.Conn
	server     *websocket.WebSocketServer
	jwtHandler *security.JWTHandler
	url        string
}

// NewWebSocketTestClient creates a new WebSocket test client
// This is the PRIMARY utility for WebSocket integration testing
func NewWebSocketTestClient(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) *WebSocketTestClient {
	// Start server if not running
	if !server.IsRunning() {
		err := server.Start()
		require.NoError(t, err, "Failed to start WebSocket server for testing")

		// Wait for server to be ready
		time.Sleep(100 * time.Millisecond)
	}

	// Connect to WebSocket server
	url := "ws://localhost:8002/ws"
	conn, _, err := gorilla.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "Failed to connect to WebSocket server")

	return &WebSocketTestClient{
		t:          t,
		conn:       conn,
		server:     server,
		jwtHandler: jwtHandler,
		url:        url,
	}
}

// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendRequest sends a JSON-RPC request over WebSocket and returns the response
// This exercises the ACTUAL WebSocket connection flow that real clients use
func (c *WebSocketTestClient) SendRequest(request *websocket.JsonRpcRequest) *websocket.JsonRpcResponse {
	// Send request over WebSocket
	err := c.conn.WriteJSON(request)
	require.NoError(c.t, err, "Failed to send JSON-RPC request over WebSocket")

	// Minimal delay for server processing (reduced from 50ms for performance)
	time.Sleep(5 * time.Millisecond)

	// Read response from WebSocket
	var response websocket.JsonRpcResponse
	err = c.conn.ReadJSON(&response)
	require.NoError(c.t, err, "Failed to read JSON-RPC response from WebSocket")

	return &response
}

// SendPingRequest sends a ping request and validates the response
// This tests the basic WebSocket connection and JSON-RPC flow
func (c *WebSocketTestClient) SendPingRequest() *websocket.JsonRpcResponse {
	// First authenticate the client through the WebSocket connection
	// Ping requires authentication per API documentation
	token, err := c.jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(c.t, err, "Failed to generate test token")

	authRequest := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "authenticate",
		ID:      0,
		Params: map[string]interface{}{
			"auth_token": token,
		},
	}

	// Send authentication request
	authResponse := c.SendRequest(authRequest)
	require.Nil(c.t, authResponse.Error, "Authentication should succeed")

	// Now send ping request
	request := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
		Params:  map[string]interface{}{},
	}

	response := c.SendRequest(request)

	// Validate ping response
	require.Equal(c.t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.t, float64(1), response.ID, "Response should have correct ID")
	require.NotNil(c.t, response.Result, "Ping response should have result")
	require.Nil(c.t, response.Error, "Ping response should not have error")

	return response
}

// SendAuthenticationRequest sends an authentication request
// This tests the authentication flow over WebSocket
func (c *WebSocketTestClient) SendAuthenticationRequest(authToken string) *websocket.JsonRpcResponse {
	request := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "authenticate",
		ID:      2,
		Params: map[string]interface{}{
			"auth_token": authToken,
		},
	}

	response := c.SendRequest(request)

	// Validate authentication response structure
	require.Equal(c.t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.t, float64(2), response.ID, "Response should have correct ID")

	return response
}

// SendInvalidRequest sends an invalid JSON-RPC request to test error handling
// This tests the error handling flow over WebSocket
func (c *WebSocketTestClient) SendInvalidRequest() *websocket.JsonRpcResponse {
	request := &websocket.JsonRpcRequest{
		JSONRPC: "1.0", // Invalid version
		Method:  "ping",
		ID:      3,
		Params:  map[string]interface{}{},
	}

	response := c.SendRequest(request)

	// Validate error response structure
	require.Equal(c.t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.t, float64(3), response.ID, "Response should have correct ID")
	require.NotNil(c.t, response.Error, "Invalid request should return error")
	require.Equal(c.t, websocket.INVALID_PARAMS, response.Error.Code, "Should return INVALID_PARAMS error")

	return response
}

// GetServerMetrics returns the server metrics after WebSocket operations
// This validates that WebSocket operations are properly tracked
func (c *WebSocketTestClient) GetServerMetrics() *websocket.PerformanceMetrics {
	return c.server.GetMetrics()
}

// TestWebSocketConnectionFlow tests the complete WebSocket connection flow
// This is a reusable test function that can be used across different test categories
func TestWebSocketConnectionFlow(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) {
	// Create WebSocket test client
	client := NewWebSocketTestClient(t, server, jwtHandler)
	defer client.Close()

	// Test basic ping functionality
	response := client.SendPingRequest()
	require.NotNil(t, response.Result, "Ping should return result")

	// Test authentication with invalid token
	authResponse := client.SendAuthenticationRequest("invalid_token_for_testing")
	require.NotNil(t, authResponse.Error, "Invalid token should return error")
	require.Equal(t, websocket.AUTHENTICATION_REQUIRED, authResponse.Error.Code, "Should return AUTHENTICATION_REQUIRED error")

	// Test invalid JSON-RPC request
	errorResponse := client.SendInvalidRequest()
	require.NotNil(t, errorResponse.Error, "Invalid request should return error")
	require.Equal(t, websocket.INVALID_PARAMS, errorResponse.Error.Code, "Should return INVALID_PARAMS error")

	// Verify that WebSocket operations are tracked in metrics
	metrics := client.GetServerMetrics()
	require.Greater(t, metrics.RequestCount, int64(0), "Request count should be incremented by WebSocket requests")
}

// TestWebSocketAuthenticationFlow tests the complete authentication flow
// This is a reusable test function for authentication testing
func TestWebSocketAuthenticationFlow(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) {
	// Create WebSocket test client
	client := NewWebSocketTestClient(t, server, jwtHandler)
	defer client.Close()

	// Test authentication with valid token
	token, err := jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(t, err, "Failed to generate test token")
	authResponse := client.SendAuthenticationRequest(token)
	require.Nil(t, authResponse.Error, "Valid token should authenticate successfully")
	require.NotNil(t, authResponse.Result, "Authentication should return result")

	// Test ping after authentication
	pingResponse := client.SendPingRequest()
	require.NotNil(t, pingResponse.Result, "Ping should work after authentication")
}

// TestWebSocketErrorHandling tests various error scenarios
// This is a reusable test function for error handling testing
func TestWebSocketErrorHandling(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) {
	// Create WebSocket test client
	client := NewWebSocketTestClient(t, server, jwtHandler)
	defer client.Close()

	// Test non-existent method
	nonExistentRequest := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "non_existent_method",
		ID:      4,
		Params:  map[string]interface{}{},
	}
	nonExistentResponse := client.SendRequest(nonExistentRequest)
	require.NotNil(t, nonExistentResponse.Error, "Non-existent method should return error")
	require.Equal(t, websocket.METHOD_NOT_FOUND, nonExistentResponse.Error.Code, "Should return METHOD_NOT_FOUND error")

	// Test missing method
	missingMethodRequest := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "", // Missing method
		ID:      5,
		Params:  map[string]interface{}{},
	}
	missingMethodResponse := client.SendRequest(missingMethodRequest)
	require.NotNil(t, missingMethodResponse.Error, "Missing method should return error")
}
