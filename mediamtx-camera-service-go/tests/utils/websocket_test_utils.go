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
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// GetFreePort returns a free port by letting the OS assign one
// This is the same approach used in the Python test suite
func GetFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 8002 // fallback
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 8002 // fallback
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

// WebSocketTestClient provides a reusable WebSocket client for testing
// This should be used in integration, performance, and security tests
type WebSocketTestClient struct {
	t          *testing.T
	conn       *gorilla.Conn
	server     *websocket.WebSocketServer
	jwtHandler *security.JWTHandler
	url        string
	port       int
}

// NewWebSocketTestClient creates a new WebSocket test client with free port
// This is the PRIMARY utility for WebSocket integration testing
func NewWebSocketTestClient(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) *WebSocketTestClient {
	// Get a free port (same approach as Python test suite)
	port := GetFreePort()
	t.Logf("Using free port: %d", port)

	// Create a new server configuration with the free port
	serverConfig := server.GetConfig()
	if serverConfig == nil {
		// Fallback to default config if server config is nil
		serverConfig = websocket.DefaultServerConfig()
	}

	// Create a new config with the free port
	newConfig := *serverConfig
	newConfig.Port = port

	// Create a new server instance with the free port
	// We need to get the dependencies from the original server
	// This is a bit of a hack, but it's the cleanest way to handle this
	// The server doesn't expose its dependencies, so we'll create a new one
	// with the same dependencies but different port

	// For now, let's try a different approach - create a new server with the config
	// but we need the dependencies from the original server
	// Let me check if we can access the server's dependencies

	// Since we can't easily access the server's dependencies, let's try to modify the original server's config
	// and restart it if needed

	if server.IsRunning() {
		// Stop the server first
		server.Stop()
		time.Sleep(100 * time.Millisecond) // Give it time to stop
	}

	// Set the new config
	server.SetConfig(&newConfig)

	// Start the server with the new config
	err := server.Start()
	require.NoError(t, err, "Failed to start WebSocket server for testing")

	// Give server a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// Wait for server to be ready - use retry logic to handle race condition
	maxRetries := 30
	retryDelay := 100 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		// First check if the port is listening (TCP level check)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			// Port is listening, now try WebSocket connection
			url := fmt.Sprintf("ws://localhost:%d/ws", port)
			wsConn, _, err := gorilla.DefaultDialer.Dial(url, nil)
			if err == nil {
				// Server is ready, close test connection and create real one
				wsConn.Close()
				t.Logf("WebSocket server ready after %d attempts on port %d", i+1, port)
				break
			}
			t.Logf("Port listening but WebSocket not ready, attempt %d/%d: %v", i+1, maxRetries, err)
		} else {
			t.Logf("Port not listening, attempt %d/%d: %v", i+1, maxRetries, err)
		}

		// Wait a bit before retrying
		time.Sleep(retryDelay)

		// If this is the last retry, fail the test
		if i == maxRetries-1 {
			require.NoError(t, err, "Failed to connect to WebSocket server after %d retries on port %d", maxRetries, port)
		}
	}

	// Connect to WebSocket server
	url := fmt.Sprintf("ws://localhost:%d/ws", port)
	conn, _, err := gorilla.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "Failed to connect to WebSocket server on port %d", port)

	return &WebSocketTestClient{
		t:          t,
		conn:       conn,
		server:     server,
		jwtHandler: jwtHandler,
		url:        url,
		port:       port,
	}
}

// NewWebSocketTestClientForExistingServer creates a WebSocket test client that connects to an already-running server
// Use this when you want to create multiple clients for the same server
func NewWebSocketTestClientForExistingServer(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler, port int) *WebSocketTestClient {
	// Connect to the already-running WebSocket server
	url := fmt.Sprintf("ws://localhost:%d/ws", port)
	conn, _, err := gorilla.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "Failed to connect to WebSocket server on port %d", port)

	return &WebSocketTestClient{
		t:          t,
		conn:       conn,
		server:     server,
		jwtHandler: jwtHandler,
		url:        url,
		port:       port,
	}
}

// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetPort returns the port number used by this client
func (c *WebSocketTestClient) GetPort() int {
	return c.port
}

// SendRequest sends a JSON-RPC request over WebSocket and returns the response
// This exercises the ACTUAL WebSocket connection flow that real clients use
func (c *WebSocketTestClient) SendRequest(request *websocket.JsonRpcRequest) *websocket.JsonRpcResponse {
	// Send request over WebSocket
	err := c.conn.WriteJSON(request)
	require.NoError(c.t, err, "Failed to send JSON-RPC request over WebSocket")

	// Minimal delay for server processing (reduced for performance)
	time.Sleep(1 * time.Millisecond)

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

// WriteMessage sends a raw message over the WebSocket connection
// This is used for testing invalid requests and error scenarios
func (c *WebSocketTestClient) WriteMessage(messageType int, data []byte) error {
	return c.conn.WriteMessage(messageType, data)
}

// ReadMessage reads a raw message from the WebSocket connection
// This is used for testing invalid requests and error scenarios
func (c *WebSocketTestClient) ReadMessage() (messageType int, p []byte, err error) {
	return c.conn.ReadMessage()
}

// TestWebSocketConnectionFlow tests the complete WebSocket connection flow
// This is a reusable test function that can be used across different test categories
func TestWebSocketConnectionFlow(t *testing.T, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) {
	// Create WebSocket test client
	client := NewWebSocketTestClient(t, server, jwtHandler)
	defer client.Close()

	// Test authentication flow
	authResponse := client.SendAuthenticationRequest("test_token")
	require.NotNil(t, authResponse, "Authentication response should not be nil")

	// Test ping flow
	pingResponse := client.SendPingRequest()
	require.NotNil(t, pingResponse, "Ping response should not be nil")
	require.Nil(t, pingResponse.Error, "Ping should not return error")

	// Test error handling flow
	errorResponse := client.SendInvalidRequest()
	require.NotNil(t, errorResponse, "Error response should not be nil")
	require.NotNil(t, errorResponse.Error, "Invalid request should return error")
}

// ============================================================================
// BENCHMARK-SPECIFIC WEBSOCKET TEST CLIENT FUNCTIONS
// ============================================================================

// WebSocketTestClientForBenchmark provides a reusable WebSocket client for benchmarks
// This is the benchmark version of WebSocketTestClient
type WebSocketTestClientForBenchmark struct {
	b          *testing.B
	conn       *gorilla.Conn
	server     *websocket.WebSocketServer
	jwtHandler *security.JWTHandler
	url        string
}

// NewWebSocketTestClientForBenchmark creates a new WebSocket test client for benchmarks
// This is the benchmark version of NewWebSocketTestClient
func NewWebSocketTestClientForBenchmark(b *testing.B, server *websocket.WebSocketServer, jwtHandler *security.JWTHandler) *WebSocketTestClientForBenchmark {
	// Start server if not running
	if !server.IsRunning() {
		err := server.Start()
		require.NoError(b, err, "Failed to start WebSocket server for testing")

		// Wait for server to be ready
		time.Sleep(100 * time.Millisecond)
	}

	// Connect to WebSocket server
	url := "ws://localhost:8002/ws"
	conn, _, err := gorilla.DefaultDialer.Dial(url, nil)
	require.NoError(b, err, "Failed to connect to WebSocket server")

	return &WebSocketTestClientForBenchmark{
		b:          b,
		conn:       conn,
		server:     server,
		jwtHandler: jwtHandler,
		url:        url,
	}
}

// Close closes the WebSocket connection for benchmarks
func (c *WebSocketTestClientForBenchmark) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendRequest sends a JSON-RPC request over WebSocket and returns the response for benchmarks
// This is the benchmark version of SendRequest
func (c *WebSocketTestClientForBenchmark) SendRequest(request *websocket.JsonRpcRequest) *websocket.JsonRpcResponse {
	// Send request over WebSocket
	err := c.conn.WriteJSON(request)
	require.NoError(c.b, err, "Failed to send JSON-RPC request over WebSocket")

	// Minimal delay for server processing (reduced for performance)
	time.Sleep(1 * time.Millisecond)

	// Read response from WebSocket
	var response websocket.JsonRpcResponse
	err = c.conn.ReadJSON(&response)
	require.NoError(c.b, err, "Failed to read JSON-RPC response from WebSocket")

	return &response
}

// SendPingRequest sends a ping request and validates the response for benchmarks
// This is the benchmark version of SendPingRequest
func (c *WebSocketTestClientForBenchmark) SendPingRequest() *websocket.JsonRpcResponse {
	// First authenticate the client through the WebSocket connection
	// Ping requires authentication per API documentation
	token, err := c.jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(c.b, err, "Failed to generate test token")

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
	require.Nil(c.b, authResponse.Error, "Authentication should succeed")

	// Now send ping request
	request := &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
		Params:  map[string]interface{}{},
	}

	response := c.SendRequest(request)

	// Validate ping response
	require.Equal(c.b, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.b, float64(1), response.ID, "Response should have correct ID")
	require.NotNil(c.b, response.Result, "Ping response should have result")
	require.Nil(c.b, response.Error, "Ping response should not have error")

	return response
}

// SendAuthenticationRequest sends an authentication request for benchmarks
// This is the benchmark version of SendAuthenticationRequest
func (c *WebSocketTestClientForBenchmark) SendAuthenticationRequest(authToken string) *websocket.JsonRpcResponse {
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
	require.Equal(c.b, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.b, float64(2), response.ID, "Response should have correct ID")

	return response
}

// SendInvalidRequest sends an invalid JSON-RPC request to test error handling for benchmarks
// This is the benchmark version of SendInvalidRequest
func (c *WebSocketTestClientForBenchmark) SendInvalidRequest() *websocket.JsonRpcResponse {
	request := &websocket.JsonRpcRequest{
		JSONRPC: "1.0", // Invalid version
		Method:  "ping",
		ID:      3,
		Params:  map[string]interface{}{},
	}

	response := c.SendRequest(request)

	// Validate error response structure
	require.Equal(c.b, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.Equal(c.b, float64(3), response.ID, "Response should have correct ID")
	require.NotNil(c.b, response.Error, "Invalid request should return error")
	require.Equal(c.b, websocket.INVALID_PARAMS, response.Error.Code, "Should return INVALID_PARAMS error")

	return response
}

// GetServerMetrics returns the server metrics after WebSocket operations for benchmarks
// This is the benchmark version of GetServerMetrics
func (c *WebSocketTestClientForBenchmark) GetServerMetrics() *websocket.PerformanceMetrics {
	return c.server.GetMetrics()
}

// DialWebSocket creates a direct WebSocket connection for testing
// This is used for testing invalid requests and error scenarios
func DialWebSocket(url string) (*gorilla.Conn, *http.Response, error) {
	return gorilla.DefaultDialer.Dial(url, nil)
}
