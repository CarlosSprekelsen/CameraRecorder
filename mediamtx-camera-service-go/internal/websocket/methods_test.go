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

// createMediaMTXControllerForTest creates an individual MediaMTX controller for each test
// This ensures true test isolation and enables parallel test execution
func createMediaMTXControllerForTest(t *testing.T) mediamtx.MediaMTXController {
	// Create individual MediaMTX helper per test for isolation
	mediaMTXHelper := mediamtx.NewMediaMTXTestHelper(t, nil)

	// Start MediaMTX controller following proper orchestration
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// Start the controller
	ctx := context.Background()
	if concreteController, ok := controller.(interface{ Start(context.Context) error }); ok {
		err := concreteController.Start(ctx)
		require.NoError(t, err, "Failed to start MediaMTX controller")
	}

	return controller
}

// TestWebSocketMethods_Ping tests ping method
func TestWebSocketMethods_Ping(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test authentication using proper test infrastructure
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Verify authentication worked by testing a protected method
	message := CreateTestMessage("ping", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response - ping should work after authentication
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Equal(t, "pong", response.Result, "Response should have correct result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetServerInfo tests get_server_info method
func TestWebSocketMethods_GetServerInfo(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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

	// Test response - should have error since no recording is active
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.NotNil(t, response.Error, "Response should have error when stopping non-existent recording")
	assert.Contains(t, response.Error.Message, "not currently recording", "Error should indicate no active recording")
}

// TestWebSocketMethods_GetMetrics tests get_metrics method
func TestWebSocketMethods_GetMetrics(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Send invalid JSON
	err := conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

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

// ============================================================================
// STREAMING METHODS TESTS (High Priority - Core Functionality)
// ============================================================================

// TestWebSocketMethods_StartStreaming tests start_streaming method
func TestWebSocketMethods_StartStreaming(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for start_streaming)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send start_streaming message
	message := CreateTestMessage("start_streaming", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_StopStreaming tests stop_streaming method
func TestWebSocketMethods_StopStreaming(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for stop_streaming)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send stop_streaming message
	message := CreateTestMessage("stop_streaming", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreamURL tests get_stream_url method
func TestWebSocketMethods_GetStreamURL(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_stream_url)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_stream_url message
	message := CreateTestMessage("get_stream_url", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreamStatus tests get_stream_status method
func TestWebSocketMethods_GetStreamStatus(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_stream_status)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_stream_status message
	message := CreateTestMessage("get_stream_status", map[string]interface{}{
		"device": "camera0",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// FILE MANAGEMENT METHODS TESTS (High Priority - Core Functionality)
// ============================================================================

// TestWebSocketMethods_ListRecordings tests list_recordings method
func TestWebSocketMethods_ListRecordings(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for list_recordings)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send list_recordings message
	message := CreateTestMessage("list_recordings", map[string]interface{}{
		"limit":  10,
		"offset": 0,
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_ListSnapshots tests list_snapshots method
func TestWebSocketMethods_ListSnapshots(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for list_snapshots)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send list_snapshots message
	message := CreateTestMessage("list_snapshots", map[string]interface{}{
		"limit":  10,
		"offset": 0,
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_DeleteRecording tests delete_recording method
func TestWebSocketMethods_DeleteRecording(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for delete_recording)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send delete_recording message
	message := CreateTestMessage("delete_recording", map[string]interface{}{
		"filename": "test_recording.mp4",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_DeleteSnapshot tests delete_snapshot method
func TestWebSocketMethods_DeleteSnapshot(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for delete_snapshot)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send delete_snapshot message
	message := CreateTestMessage("delete_snapshot", map[string]interface{}{
		"filename": "test_snapshot.jpg",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// SYSTEM MANAGEMENT METHODS TESTS (Medium Priority - Admin Features)
// ============================================================================

// TestWebSocketMethods_GetStorageInfo tests get_storage_info method
func TestWebSocketMethods_GetStorageInfo(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for get_storage_info)
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Send get_storage_info message
	message := CreateTestMessage("get_storage_info", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_SetRetentionPolicy tests set_retention_policy method
func TestWebSocketMethods_SetRetentionPolicy(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for set_retention_policy)
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Send set_retention_policy message
	message := CreateTestMessage("set_retention_policy", map[string]interface{}{
		"policy_type":  "age",
		"max_age_days": 30,
		"enabled":      true,
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_CleanupOldFiles tests cleanup_old_files method
func TestWebSocketMethods_CleanupOldFiles(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for cleanup_old_files)
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Send cleanup_old_files message
	message := CreateTestMessage("cleanup_old_files", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// EVENT SYSTEM METHODS TESTS (Advanced Features)
// ============================================================================

// TestWebSocketMethods_SubscribeEvents tests subscribe_events method
func TestWebSocketMethods_SubscribeEvents(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for subscribe_events)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send subscribe_events message
	message := CreateTestMessage("subscribe_events", map[string]interface{}{
		"topics": []string{"camera.connected", "recording.start"},
		"filters": map[string]interface{}{
			"device": "camera0",
		},
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_UnsubscribeEvents tests unsubscribe_events method
func TestWebSocketMethods_UnsubscribeEvents(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for unsubscribe_events)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send unsubscribe_events message
	message := CreateTestMessage("unsubscribe_events", map[string]interface{}{
		"topics": []string{"camera.connected"},
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetSubscriptionStats tests get_subscription_stats method
func TestWebSocketMethods_GetSubscriptionStats(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_subscription_stats)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_subscription_stats message
	message := CreateTestMessage("get_subscription_stats", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// EXTERNAL STREAM METHODS TESTS (Advanced Features)
// ============================================================================

// TestWebSocketMethods_DiscoverExternalStreams tests discover_external_streams method
func TestWebSocketMethods_DiscoverExternalStreams(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for discover_external_streams)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send discover_external_streams message
	message := CreateTestMessage("discover_external_streams", map[string]interface{}{
		"skydio_enabled":  true,
		"generic_enabled": false,
		"force_rescan":    false,
		"include_offline": false,
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_AddExternalStream tests add_external_stream method
func TestWebSocketMethods_AddExternalStream(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for add_external_stream)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send add_external_stream message
	message := CreateTestMessage("add_external_stream", map[string]interface{}{
		"stream_url":  "rtsp://192.168.42.15:5554/subject",
		"stream_name": "Test_UAV_15",
		"stream_type": "skydio_stanag4609",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_RemoveExternalStream tests remove_external_stream method
func TestWebSocketMethods_RemoveExternalStream(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (operator role for remove_external_stream)
	AuthenticateTestClient(t, conn, "test_user", "operator")

	// Send remove_external_stream message
	message := CreateTestMessage("remove_external_stream", map[string]interface{}{
		"stream_url": "rtsp://192.168.42.15:5554/subject",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetExternalStreams tests get_external_streams method
func TestWebSocketMethods_GetExternalStreams(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_external_streams)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_external_streams message
	message := CreateTestMessage("get_external_streams", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_SetDiscoveryInterval tests set_discovery_interval method
func TestWebSocketMethods_SetDiscoveryInterval(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (admin role for set_discovery_interval)
	AuthenticateTestClient(t, conn, "test_user", "admin")

	// Send set_discovery_interval message
	message := CreateTestMessage("set_discovery_interval", map[string]interface{}{
		"scan_interval": 300,
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// ADDITIONAL FILE INFO METHODS TESTS (Complete Coverage)
// ============================================================================

// TestWebSocketMethods_GetRecordingInfo tests get_recording_info method
func TestWebSocketMethods_GetRecordingInfo(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_recording_info)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_recording_info message
	message := CreateTestMessage("get_recording_info", map[string]interface{}{
		"filename": "test_recording.mp4",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetSnapshotInfo tests get_snapshot_info method
func TestWebSocketMethods_GetSnapshotInfo(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_snapshot_info)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_snapshot_info message
	message := CreateTestMessage("get_snapshot_info", map[string]interface{}{
		"filename": "test_snapshot.jpg",
	})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreams tests get_streams method
func TestWebSocketMethods_GetStreams(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create individual MediaMTX controller for test isolation
	controller := createMediaMTXControllerForTest(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	server = helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Authenticate client (viewer role for get_streams)
	AuthenticateTestClient(t, conn, "test_user", "viewer")

	// Send get_streams message
	message := CreateTestMessage("get_streams", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}
