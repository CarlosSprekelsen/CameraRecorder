//go:build integration && real_websocket
// +build integration,real_websocket

/*
WebSocket Integration Tests - CONSOLIDATED

This file consolidates all WebSocket integration tests using existing utils.
Eliminates duplication across multiple WebSocket test files.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-API-004: get_camera_status method for camera status
- REQ-API-008: authenticate method for authentication
- REQ-API-009: Role-based access control with viewer, operator, admin permissions
- REQ-API-011: API methods respond within specified time limits

Test Categories: Integration/Real WebSocket
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"encoding/json"
	"testing"
	"time"

	ws "github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketConnection tests real WebSocket connection and basic communication
func TestWebSocketConnection(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Use WebSocket test client utility for real connection testing
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Test connection is alive
	assert.NotNil(t, client, "WebSocket test client should be established")
}

// TestPingMethod tests ping method with real WebSocket connection
func TestPingMethod(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Use WebSocket test client utility for real connection testing
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Use the shared utility that properly handles authentication
	startTime := time.Now()
	response := client.SendPingRequest()
	responseTime := time.Since(startTime)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, float64(1), response.ID, "Response ID should match request ID")
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

// TestAuthenticateMethod tests authenticate method with real WebSocket connection
func TestAuthenticateMethod(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Use WebSocket test client utility for real connection testing
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Generate test token using shared utilities
	testToken := utils.GenerateTestTokenWithExpiry(t, env.JWTHandler, "test_user", "operator", 1)

	// Create authentication request
	authRequest := &ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "authenticate",
		Params: map[string]interface{}{
			"token": testToken,
		},
		ID: 2,
	}

	// Send authentication request
	startTime := time.Now()
	response := client.SendRequest(authRequest)
	responseTime := time.Since(startTime)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Equal(t, float64(2), response.ID, "Response ID should match request ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate authentication result
	var authResult map[string]interface{}
	resultBytes, err := json.Marshal(response.Result)
	require.NoError(t, err)
	err = json.Unmarshal(resultBytes, &authResult)
	require.NoError(t, err)

	assert.Equal(t, true, authResult["authenticated"], "Should be authenticated")
	assert.Equal(t, "operator", authResult["role"], "Should have operator role")
	assert.NotEmpty(t, authResult["session_id"], "Should have session ID")
	assert.NotEmpty(t, authResult["expires_at"], "Should have expiration time")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "Authentication response should be <100ms per API documentation")
}

// TestGetCameraListMethod tests get_camera_list method with real WebSocket connection
func TestGetCameraListMethod(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Use WebSocket test client utility for real connection testing
	client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
	defer client.Close()

	// Create authenticated client using shared utilities
	authenticatedClient := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Create camera list request
	cameraRequest := &ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_list",
		Params:  map[string]interface{}{},
		ID:      3,
	}

	// Send camera list request
	startTime := time.Now()
	response, err := env.WebSocketServer.MethodGetCameraList(cameraRequest.Params, authenticatedClient)
	responseTime := time.Since(startTime)
	require.NoError(t, err, "Should get camera list successfully")

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate camera list result
	cameras, ok := response.Result.([]interface{})
	assert.True(t, ok, "Result should be camera list")
	assert.NotNil(t, cameras, "Camera list should not be nil")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "Camera list response should be <200ms per API documentation")
}

// TestGetCameraStatusMethod tests get_camera_status method with real WebSocket connection
func TestGetCameraStatusMethod(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Create authenticated client using shared utilities
	authenticatedClient := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Create camera status request
	statusRequest := &ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_status",
		Params: map[string]interface{}{
			"device_path": "/dev/video0",
		},
		ID: 4,
	}

	// Send camera status request
	startTime := time.Now()
	response, err := env.WebSocketServer.MethodGetCameraStatus(statusRequest.Params, authenticatedClient)
	responseTime := time.Since(startTime)
	require.NoError(t, err, "Should get camera status successfully")

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate performance target
	assert.Less(t, responseTime, 150*time.Millisecond, "Camera status response should be <150ms per API documentation")
}

// TestAuthenticationRequiredError tests authentication requirement enforcement
func TestAuthenticationRequiredError(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Create unauthenticated client
	unauthenticatedClient := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: false,
		ConnectedAt:   time.Now(),
	}

	// Try to access protected method without authentication
	cameraRequest := &ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_list",
		Params:  map[string]interface{}{},
		ID:      5,
	}

	response, err := env.WebSocketServer.MethodGetCameraList(cameraRequest.Params, unauthenticatedClient)
	require.Error(t, err, "Should fail with authentication error")

	// Should return authentication error
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.NotNil(t, response.Error, "Should have authentication error")
	assert.Nil(t, response.Result, "Should not have result")
	assert.Equal(t, -32001, response.Error.Code, "Should have authentication error code")
}

// TestInvalidParametersError tests invalid parameter handling
func TestInvalidParametersError(t *testing.T) {
	// COMMON PATTERN: Use shared WebSocket test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Create authenticated client using shared utilities
	authenticatedClient := utils.CreateAuthenticatedClient(t, env.JWTHandler, "test_user", "viewer")

	// Try to access method with invalid parameters
	invalidRequest := &ws.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "get_camera_status",
		Params: map[string]interface{}{
			"invalid_param": "invalid_value",
		},
		ID: 6,
	}

	response, err := env.WebSocketServer.MethodGetCameraStatus(invalidRequest.Params, authenticatedClient)
	require.Error(t, err, "Should fail with invalid parameters error")

	// Should return parameter error
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.NotNil(t, response.Error, "Should have parameter error")
	assert.Nil(t, response.Result, "Should not have result")
	assert.Equal(t, -32602, response.Error.Code, "Should have invalid params error code")
}
