//go:build integration && real_websocket
// +build integration,real_websocket

/*
WebSocket real connection integration tests.

Tests validate actual WebSocket connections and method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-API-004: get_camera_status method for camera status
- REQ-API-008: authenticate method for authentication
- REQ-API-009: Role-based access control with viewer, operator, admin permissions
- REQ-API-011: API methods respond within specified time limits

Test Categories: Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	ws "github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketRealConnection tests real WebSocket connection and basic communication
func TestWebSocketRealConnection(t *testing.T) {
	/*
		Integration Test for real WebSocket connection

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: WebSocket connection can be established and basic communication works
	*/

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Setup test server
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(env.ConfigManager, env.Logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", 8002)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Test connection is alive
	assert.NotNil(t, conn, "WebSocket connection should be established")
}

// TestPingMethodRealConnection tests ping method with real WebSocket connection
func TestPingMethodRealConnection(t *testing.T) {
	/*
		Integration Test for ping method with real WebSocket connection

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: ping
		Expected Response: {"jsonrpc": "2.0", "result": "pong", "id": 1}
		Performance Target: <50ms response time
	*/

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Setup test server
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(env.ConfigManager, env.Logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", 8002)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Create ping request
	pingRequest := ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
		Params:  map[string]interface{}{},
	}

	// Send request and measure response time
	startTime := time.Now()
	err = conn.WriteJSON(pingRequest)
	require.NoError(t, err)

	// Read response
	var response ws.JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	responseTime := time.Since(startTime)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, 1, response.ID, "Response ID should match request ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result is "pong" per API documentation
	var result string
	resultBytes, err := json.Marshal(response.Result)
	require.NoError(t, err)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)
	assert.Equal(t, "pong", result, "Ping result should be 'pong' per API documentation")

	// Validate performance target
	assert.Less(t, responseTime, 50*time.Millisecond, "Ping response should be <50ms per API documentation")
}

// TestAuthenticateMethodRealConnection tests authenticate method with real WebSocket connection
func TestAuthenticateMethodRealConnection(t *testing.T) {
	/*
		Integration Test for authenticate method with real WebSocket connection

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: authenticate
		Expected Response: {"jsonrpc": "2.0", "result": {"authenticated": true, "role": "operator", "permissions": ["view", "control"], "expires_at": "...", "session_id": "..."}, "id": 2}
		Performance Target: <100ms response time
	*/

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Setup test server
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(env.ConfigManager, env.Logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", 8002)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Generate test token
	testToken, err := jwtHandler.GenerateToken("test_user", "operator", 24)
	require.NoError(t, err)

	// Create authenticate request
	authRequest := ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "authenticate",
		ID:      2,
		Params: map[string]interface{}{
			"auth_token": testToken,
		},
	}

	// Send request and measure response time
	startTime := time.Now()
	err = conn.WriteJSON(authRequest)
	require.NoError(t, err)

	// Read response
	var response ws.JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	responseTime := time.Since(startTime)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, 2, response.ID, "Response ID should match request ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	var result map[string]interface{}
	resultBytes, err := json.Marshal(response.Result)
	require.NoError(t, err)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	// Check required fields per API documentation
	requiredFields := []string{"authenticated", "role", "permissions", "expires_at", "session_id"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Result should contain field '%s' per API documentation", field)
	}

	// Validate specific values
	assert.Equal(t, true, result["authenticated"], "Authenticated should be true")
	assert.Equal(t, "operator", result["role"], "Role should be operator")
	assert.NotEmpty(t, result["session_id"], "Session ID should not be empty")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "Authenticate response should be <100ms per API documentation")
}

// TestGetCameraListMethodRealConnection tests get_camera_list method with real WebSocket connection
func TestGetCameraListMethodRealConnection(t *testing.T) {
	/*
		Integration Test for get_camera_list method with real WebSocket connection

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_camera_list
		Expected Response: {"jsonrpc": "2.0", "result": {"cameras": [...], "total": 1, "connected": 1}, "id": 3}
		Performance Target: <50ms response time
	*/

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Setup test server
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(env.ConfigManager, env.Logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", 8002)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// First authenticate
	testToken, err := jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(t, err)

	authRequest := ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "authenticate",
		ID:      1,
		Params: map[string]interface{}{
			"auth_token": testToken,
		},
	}

	err = conn.WriteJSON(authRequest)
	require.NoError(t, err)

	var authResponse ws.JsonRpcResponse
	err = conn.ReadJSON(&authResponse)
	require.NoError(t, err)
	assert.Nil(t, authResponse.Error, "Authentication should succeed")

	// Create get_camera_list request
	cameraListRequest := ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_list",
		ID:      3,
		Params:  map[string]interface{}{},
	}

	// Send request and measure response time
	startTime := time.Now()
	err = conn.WriteJSON(cameraListRequest)
	require.NoError(t, err)

	// Read response
	var response ws.JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	responseTime := time.Since(startTime)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, 3, response.ID, "Response ID should match request ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	var result map[string]interface{}
	resultBytes, err := json.Marshal(response.Result)
	require.NoError(t, err)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	// Check required fields per API documentation
	requiredFields := []string{"cameras", "total", "connected"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Result should contain field '%s' per API documentation", field)
	}

	// Validate cameras array
	cameras, ok := result["cameras"].([]interface{})
	assert.True(t, ok, "Cameras should be an array")
	assert.NotNil(t, cameras, "Cameras array should not be nil")

	// Validate performance target
	assert.Less(t, responseTime, 50*time.Millisecond, "Get camera list response should be <50ms per API documentation")
}

// TestAuthenticationRequiredError tests authentication error handling with real connection
func TestAuthenticationRequiredError(t *testing.T) {
	/*
		Integration Test for authentication error handling with real WebSocket connection

		API Documentation Reference: docs/api/json_rpc_methods.md
		Expected: Methods requiring authentication should return AUTHENTICATION_REQUIRED error
	*/

	// COMMON PATTERN: Use shared test environment instead of individual components
	// This eliminates the need to create ConfigManager and Logger in every test
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Setup test server
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(env.ConfigManager, env.Logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Start server
	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", 8002)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Try to call get_camera_list without authentication
	cameraListRequest := ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_list",
		ID:      5,
		Params:  map[string]interface{}{},
	}

	err = conn.WriteJSON(cameraListRequest)
	require.NoError(t, err)

	// Read response
	var response ws.JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err)

	// Validate error response per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, 5, response.ID, "Response ID should match request ID")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Nil(t, response.Result, "Response should not have result")

	// Validate error details per API documentation
	assert.Equal(t, -32001, response.Error.Code, "Error code should be AUTHENTICATION_REQUIRED")
	assert.Equal(t, "Authentication required", response.Error.Message, "Error message should match API documentation")
}
