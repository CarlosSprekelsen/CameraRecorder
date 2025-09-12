/*
WebSocket Methods Unit Tests

Provides comprehensive unit tests for ALL exposed WebSocket methods,
following proper orchestration: WebSocket → MediaMTX Controller.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Complete interface testing

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
Architecture: WebSocket → MediaMTX Controller (proper orchestration)
*/

package websocket

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Global shared MediaMTX instance for parallel test execution
var (
	sharedMediaMTXHelper *mediamtx.MediaMTXTestHelper
	sharedMediaMTXOnce   sync.Once
	sharedMediaMTXMutex  sync.Mutex
)

// getSharedMediaMTXHelper returns a shared MediaMTX instance for parallel test execution
func getSharedMediaMTXHelper(t *testing.T) *mediamtx.MediaMTXTestHelper {
	sharedMediaMTXOnce.Do(func() {
		sharedMediaMTXMutex.Lock()
		defer sharedMediaMTXMutex.Unlock()
		
		// Create shared MediaMTX helper
		sharedMediaMTXHelper = mediamtx.NewMediaMTXTestHelper(t, nil)
		
		// Start MediaMTX controller following proper orchestration
		controller, err := sharedMediaMTXHelper.GetController(t)
		require.NoError(t, err, "Failed to create shared MediaMTX controller")
		
		// Start the controller
		ctx := context.Background()
		if concreteController, ok := controller.(interface{ Start(context.Context) error }); ok {
			err := concreteController.Start(ctx)
			require.NoError(t, err, "Failed to start shared MediaMTX controller")
		}
	})
	
	return sharedMediaMTXHelper
}

// TestWebSocketMethods_Ping tests ping method
func TestWebSocketMethods_Ping(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance for proper orchestration
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	// Set the controller in WebSocket server
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)

	// Start server following Progressive Readiness Pattern
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client
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

// TestWebSocketMethods_Authenticate tests authenticate method
func TestWebSocketMethods_Authenticate(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test authentication
	message := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": "test-jwt-token",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetServerInfo tests get_server_info method
func TestWebSocketMethods_GetServerInfo(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for get_server_info)
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
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for get_status)
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

// TestWebSocketMethods_GetCameraList tests get_camera_list method (WebSocket → MediaMTX Controller)
func TestWebSocketMethods_GetCameraList(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_camera_list message
	message := CreateTestMessage("get_camera_list", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetCameraStatus tests get_camera_status method (WebSocket → MediaMTX Controller)
func TestWebSocketMethods_GetCameraStatus(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_camera_status message
	message := CreateTestMessage("get_camera_status", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	// Note: Result may be nil if camera0 doesn't exist, but error should be properly formatted
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetCameraCapabilities tests get_camera_capabilities method
func TestWebSocketMethods_GetCameraCapabilities(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_camera_capabilities message
	message := CreateTestMessage("get_camera_capabilities", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_TakeSnapshot tests take_snapshot method
func TestWebSocketMethods_TakeSnapshot(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for take_snapshot)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send take_snapshot message
	message := CreateTestMessage("take_snapshot", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_StartRecording tests start_recording method
func TestWebSocketMethods_StartRecording(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for start_recording)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send start_recording message
	message := CreateTestMessage("start_recording", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_StopRecording tests stop_recording method
func TestWebSocketMethods_StopRecording(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for stop_recording)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send stop_recording message
	message := CreateTestMessage("stop_recording", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetMetrics tests get_metrics method
func TestWebSocketMethods_GetMetrics(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for get_metrics)
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Send get_metrics message
	message := CreateTestMessage("get_metrics", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_InvalidJSON tests invalid JSON handling
func TestWebSocketMethods_InvalidJSON(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

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
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

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

// TestWebSocketMethods_UnauthenticatedAccess tests that methods require authentication
func TestWebSocketMethods_UnauthenticatedAccess(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client WITHOUT authentication
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test that unauthenticated access to protected methods fails
	protectedMethods := []string{
		"get_camera_list",
		"get_camera_status", 
		"get_camera_capabilities",
		"start_recording",
		"stop_recording",
		"take_snapshot",
		"get_metrics",
		"get_server_info",
		"get_status",
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

// TestWebSocketMethods_SequentialRequests tests sequential request handling
func TestWebSocketMethods_SequentialRequests(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Create a single connection for all requests
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate the client once
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

	// Verify reasonable performance
	assert.Less(t, duration, 5*time.Second, "Requests should complete within reasonable time")
}

// TestWebSocketMethods_MultipleConnections tests multiple connections handling
func TestWebSocketMethods_MultipleConnections(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use shared MediaMTX instance
	mediaMTXHelper := getSharedMediaMTXHelper(t)
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to get shared MediaMTX controller")
	
	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

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
			conn := helper.NewTestClient(t, server)
			defer helper.CleanupTestClient(t, conn)

			// Authenticate the client
			AuthenticateTestClient(t, conn, "test_user", "viewer")

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