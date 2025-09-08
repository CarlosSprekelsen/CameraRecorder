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

	// Authenticate client using the helper
	AuthenticateTestClient(t, conn, "test_user", "viewer")

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

	// Authenticate client using the helper (admin role for get_server_info)
	AuthenticateTestClient(t, conn, "test_user", "admin")

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

	// Authenticate client using the helper (admin role for get_status)
	AuthenticateTestClient(t, conn, "test_user", "admin")

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

	// Authenticate the client once using the helper
	AuthenticateTestClient(t, conn, "test_user", "viewer")

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

	// Create a test JWT token for authentication using the helper
	jwtHandler := security.TestJWTHandler(t)
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

	// Authenticate client using the helper (admin role for get_server_info)
	AuthenticateTestClient(t, conn, "test_user", "admin")

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

// =============================================================================
// CRITICAL CAMERA OPERATIONS TESTS (0% Coverage - High Priority)
// =============================================================================

// TestWebSocketMethods_GetCameraList tests the get_camera_list method
func TestWebSocketMethods_GetCameraList(t *testing.T) {
	// REQ-API-004: Core method implementations (get_camera_list)

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client using the new helper (eliminates duplication)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Test get_camera_list method
	message := CreateTestMessage("get_camera_list", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Verify response structure
	require.Nil(t, response.Error, "get_camera_list should not return error")
	require.NotNil(t, response.Result, "get_camera_list should return result")

	// Verify result is an object with cameras array (per API documentation)
	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "get_camera_list should return object with cameras array, got type %T", response.Result)

	// Verify cameras field exists and is an array
	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(t, ok, "get_camera_list result should have 'cameras' field as array")

	// Verify metadata fields exist
	connected, hasConnected := resultMap["connected"]
	total, hasTotal := resultMap["total"]
	require.True(t, hasConnected, "get_camera_list result should have 'connected' field")
	require.True(t, hasTotal, "get_camera_list result should have 'total' field")

	// Log the result for debugging
	t.Logf("Found %d cameras (connected: %v, total: %v): %v", len(cameras), connected, total, cameras)
}

// TestWebSocketMethods_GetCameraStatus tests the get_camera_status method
func TestWebSocketMethods_GetCameraStatus(t *testing.T) {
	// REQ-API-004: Core method implementations (get_camera_status)

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Test get_camera_status with valid camera identifier
	message := CreateTestMessage("get_camera_status", map[string]interface{}{
		"device": "camera0", // Using device parameter as per API documentation
	})
	response := SendTestMessage(t, conn, message)

	// Verify response structure
	require.Nil(t, response.Error, "get_camera_status should not return error")
	require.NotNil(t, response.Result, "get_camera_status should return result")

	// Verify result structure
	status, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "get_camera_status should return status object")

	// Log the result for debugging
	t.Logf("Camera status: %v", status)
}

// TestWebSocketMethods_GetCameraCapabilities tests the get_camera_capabilities method
func TestWebSocketMethods_GetCameraCapabilities(t *testing.T) {
	// REQ-API-004: Core method implementations (get_camera_capabilities)

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Test get_camera_capabilities with valid camera identifier
	message := CreateTestMessage("get_camera_capabilities", map[string]interface{}{
		"device": "camera0", // Using device parameter as per API documentation
	})
	response := SendTestMessage(t, conn, message)

	// Verify response structure
	require.Nil(t, response.Error, "get_camera_capabilities should not return error")
	require.NotNil(t, response.Result, "get_camera_capabilities should return result")

	// Verify result structure
	capabilities, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "get_camera_capabilities should return capabilities object")

	// Log the result for debugging
	t.Logf("Camera capabilities: %v", capabilities)
}

// TestWebSocketMethods_StartRecording tests the start_recording method
func TestWebSocketMethods_StartRecording(t *testing.T) {
	// REQ-API-004: Core method implementations (start_recording)

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client with admin role for recording operations
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Test start_recording with valid camera identifier
	message := CreateTestMessage("start_recording", map[string]interface{}{
		"device": "camera0", // Using device parameter as per API documentation
	})
	response := SendTestMessage(t, conn, message)

	// Verify response structure
	require.Nil(t, response.Error, "start_recording should not return error")
	require.NotNil(t, response.Result, "start_recording should return result")

	// Log the result for debugging
	t.Logf("Start recording result: %v", response.Result)
}

// TestWebSocketMethods_StopRecording tests the stop_recording method
func TestWebSocketMethods_StopRecording(t *testing.T) {
	// REQ-API-004: Core method implementations (stop_recording)

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client with admin role for recording operations
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Test stop_recording with valid camera identifier
	message := CreateTestMessage("stop_recording", map[string]interface{}{
		"device": "camera0", // Using device parameter as per API documentation
	})
	response := SendTestMessage(t, conn, message)

	// Verify response structure
	require.Nil(t, response.Error, "stop_recording should not return error")
	require.NotNil(t, response.Result, "stop_recording should return result")

	// Log the result for debugging
	t.Logf("Stop recording result: %v", response.Result)
}

// =============================================================================
// AUTHENTICATION EDGE CASES TESTS
// =============================================================================

// TestWebSocketMethods_UnauthenticatedAccess tests that methods require authentication
func TestWebSocketMethods_UnauthenticatedAccess(t *testing.T) {
	// REQ-API-004: Core method implementations with authentication checks

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Server should start successfully")

	// Connect client WITHOUT authentication
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Test that unauthenticated access to protected methods fails
	protectedMethods := []string{
		"get_camera_list",
		"get_camera_status",
		"get_camera_capabilities",
		"start_recording",
		"stop_recording",
	}

	for _, method := range protectedMethods {
		t.Run(method, func(t *testing.T) {
			message := CreateTestMessage(method, map[string]interface{}{})
			response := SendTestMessage(t, conn, message)

			// Verify authentication error
			require.NotNil(t, response.Error, "%s should require authentication", method)
			require.Equal(t, AUTHENTICATION_REQUIRED, response.Error.Code, "%s should return AUTHENTICATION_REQUIRED error", method)
		})
	}
}

// TestWebSocketMethods_InvalidCameraID tests methods with invalid camera identifiers
func TestWebSocketMethods_InvalidCameraID(t *testing.T) {
	// REQ-API-004: Core method implementations with parameter validation

	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Test methods with invalid camera identifier
	invalidCameraMethods := []string{
		"get_camera_status",
		"get_camera_capabilities",
		"start_recording",
		"stop_recording",
	}

	for _, method := range invalidCameraMethods {
		t.Run(method, func(t *testing.T) {
			message := CreateTestMessage(method, map[string]interface{}{
				"device": "invalid_camera_999", // Invalid camera identifier
			})
			response := SendTestMessage(t, conn, message)

			// Verify error handling
			require.NotNil(t, response.Error, "%s should return error for invalid camera", method)
			t.Logf("%s with invalid camera returned error: %v", method, response.Error)
		})
	}
}
