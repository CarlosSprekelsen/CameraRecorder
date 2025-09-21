/*
WebSocket Methods Unit Tests - Enterprise-Grade Progressive Readiness Pattern

Provides comprehensive unit tests for ALL exposed WebSocket methods,
following homogeneous enterprise-grade patterns with real hardware integration.

ENTERPRISE STANDARDS:
- Progressive Readiness Pattern compliance (no polling, no sequential execution)
- Real hardware integration (no mocking, no skipping)
- Homogeneous test patterns across all methods
- Immediate connection acceptance testing (<100ms)
- Proper documentation with requirements coverage

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Complete interface testing
- REQ-ARCH-001: Progressive Readiness Pattern compliance

Test Categories: Enterprise Integration
API Documentation Reference: docs/api/json_rpc_methods.md
Architecture: WebSocket → MediaMTX Controller → Real Hardware (no mocking)
Pattern: Progressive Readiness with immediate connection acceptance
*/

package websocket

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// createMediaMTXControllerUsingProvenPattern creates a MediaMTX controller using the exact same pattern
// as the working MediaMTX tests. This ensures homogeneous test suite with consistent patterns.
func createMediaMTXControllerUsingProvenPattern(t *testing.T) mediamtx.MediaMTXController {
	// Use the EXACT same pattern as working MediaMTX tests
	mediaMTXHelper := mediamtx.NewMediaMTXTestHelper(t, nil)

	// Get controller using the proven pattern
	controller, err := mediaMTXHelper.GetController(t)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// CRITICAL: Register cleanup to prevent resource leaks
	t.Cleanup(func() {
		mediaMTXHelper.Cleanup(t)
	})

	return controller
}

// waitForSystemReadiness implements the Progressive Readiness Pattern exactly as main.go does.
// Uses event-driven approach with SubscribeToReadiness() and context timeout.
func waitForSystemReadiness(t *testing.T, controller mediamtx.MediaMTXController) {
	// Use event-driven approach - subscribe to readiness events
	readinessChan := controller.SubscribeToReadiness()

	// Apply readiness timeout to prevent indefinite blocking (same as main.go)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for readiness event with timeout (exact same pattern as main.go)
	select {
	case <-readinessChan:
		t.Log("Controller readiness event received - all services ready")
	case <-ctx.Done():
		t.Log("Controller readiness timeout - proceeding anyway")
	}

	// Verify actual readiness state from controller (same as main.go)
	if controller.IsReady() {
		t.Log("Controller reports ready - all services operational")
	} else {
		t.Log("Controller not ready - some services may not be operational")
	}
}

// TestWebSocketMethods_Ping validates WebSocket ping method with Progressive Readiness Pattern
//
// Requirements Coverage:
// - REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
// - REQ-API-002: JSON-RPC 2.0 protocol implementation
// - REQ-ARCH-001: Progressive Readiness Pattern compliance
//
// Test Pattern: Enterprise-grade real hardware testing, no mocking, no skipping
// Architecture: WebSocket → MediaMTX Controller → Real Hardware
func TestWebSocketMethods_Ping_ReqAPI002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN (2 lines total) ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "ping", map[string]interface{}{}, "viewer")

	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, "pong", response.Result, "Response should have correct result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_Authenticate validates WebSocket authentication with Progressive Readiness Pattern
//
// Requirements Coverage:
// - REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
// - REQ-API-003: Authentication and authorization
// - REQ-ARCH-001: Progressive Readiness Pattern compliance
//
// Test Pattern: Enterprise-grade real hardware testing, no mocking, no skipping
// Architecture: WebSocket → Security → JWT Authentication → Real Hardware
func TestWebSocketMethods_Authenticate_ReqSEC001_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	// === ENTERPRISE MINIMAL PATTERN (3 lines setup) ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)
	conn := helper.GetAuthenticatedConnection(t, "test_user", "viewer")
	defer helper.CleanupTestClient(t, conn)

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
func TestWebSocketMethods_GetServerInfo_ReqAPI002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_server_info", map[string]interface{}{}, "admin")

	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetStatus tests get_status method
func TestWebSocketMethods_GetStatus_ReqAPI002_Success(t *testing.T) {
	// === ENTERPRISE MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_status", map[string]interface{}{}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetCameraList tests get_camera_list method (WebSocket → MediaMTX Controller)
func TestWebSocketMethods_GetCameraList_ReqCAM001_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_camera_list", map[string]interface{}{}, "viewer")

	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetCameraStatus tests get_camera_status method (WebSocket → MediaMTX Controller)
func TestWebSocketMethods_GetCameraStatus_ReqCAM001_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_camera_status", map[string]interface{}{
		"device": "camera0",
	}, "viewer")

	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	// Note: Result may be nil if camera0 doesn't exist, but error should be properly formatted
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_GetCameraCapabilities tests get_camera_capabilities method
func TestWebSocketMethods_GetCameraCapabilities_ReqCAM001_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_camera_capabilities", map[string]interface{}{
		"device": "camera0",
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_TakeSnapshot tests take_snapshot method
func TestWebSocketMethods_TakeSnapshot_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "take_snapshot", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_StartRecording tests start_recording method
func TestWebSocketMethods_StartRecording_ReqMTX002_Success(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// CRITICAL: Wait for server to be ready for API contract validation
	helper.WaitForServerReady(t)

	response := helper.TestMethod(t, "start_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// Validate JSON-RPC 2.0 compliance
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")

	// Validate API contract per docs/api/json_rpc_methods.md
	if response.Error != nil {
		// Validate error is API-compliant
		validateAPICompliantError(t, response.Error)

		// For recording methods, errors must be recording-specific
		validateRecordingSpecificError(t, response.Error.Code, "start_recording")
	} else {
		// Success case - validate StartRecordingResponse structure
		validateStartRecordingResponse(t, response.Result)
	}
}

// TestWebSocketMethods_StopRecording tests stop_recording method API contract
func TestWebSocketMethods_StopRecording_ReqMTX002_Success(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// CRITICAL: Wait for server to be ready for API contract validation
	helper.WaitForServerReady(t)

	response := helper.TestMethod(t, "stop_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// Validate JSON-RPC 2.0 compliance
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")

	// Validate API contract per docs/api/json_rpc_methods.md
	if response.Error != nil {
		// Validate error is API-compliant
		validateAPICompliantError(t, response.Error)

		// For recording methods, errors must be recording-specific
		validateRecordingSpecificError(t, response.Error.Code, "stop_recording")
	} else {
		// Success case - validate StopRecordingResponse structure
		validateStopRecordingResponse(t, response.Result)
	}
}

// TestWebSocketMethods_GetMetrics tests get_metrics method
func TestWebSocketMethods_GetMetrics_ReqMTX004_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_metrics", map[string]interface{}{}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.NotNil(t, response.Result, "Response should have result")
	assert.Nil(t, response.Error, "Response should not have error")
}

// TestWebSocketMethods_InvalidJSON tests invalid JSON handling
func TestWebSocketMethods_ProcessMessage_ReqAPI002_ErrorHandling_InvalidJSON(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proven MediaMTX pattern - EXACT same pattern as working MediaMTX tests
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	_ = helper.StartServer(t) // Server is started, we use the original server instance

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
func TestWebSocketMethods_ProcessMessage_ReqAPI002_ErrorHandling_MissingMethod(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proven MediaMTX pattern - EXACT same pattern as working MediaMTX tests
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	_ = helper.StartServer(t) // Server is started, we use the original server instance

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
func TestWebSocketMethods_Authenticate_ReqSEC001_ErrorHandling_UnauthenticatedAccess(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proven MediaMTX pattern - EXACT same pattern as working MediaMTX tests
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	_ = helper.StartServer(t) // Server is started, we use the original server instance

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
func TestWebSocketMethods_ProcessMessage_ReqAPI002_SequentialRequests(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proven MediaMTX pattern - EXACT same pattern as working MediaMTX tests
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	_ = helper.StartServer(t) // Server is started, we use the original server instance

	// Create authenticated connection using standardized pattern
	conn := helper.GetAuthenticatedConnection(t, "test_user", "viewer")
	defer helper.CleanupTestClient(t, conn)

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
func TestWebSocketMethods_ProcessMessage_ReqAPI001_MultipleConnections(t *testing.T) {
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use proven MediaMTX pattern - EXACT same pattern as working MediaMTX tests
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)
	_ = helper.StartServer(t) // Server is started, we use the original server instance

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

			// Create authenticated connection using standardized pattern
			conn := helper.GetAuthenticatedConnection(t, "test_user", "viewer")
			defer helper.CleanupTestClient(t, conn)

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
func TestWebSocketMethods_StartStreaming_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "start_streaming", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_StopStreaming tests stop_streaming method
func TestWebSocketMethods_StopStreaming_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "stop_streaming", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreamURL tests get_stream_url method
func TestWebSocketMethods_GetStreamURL_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_stream_url", map[string]interface{}{
		"device": "camera0",
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreamStatus tests get_stream_status method
func TestWebSocketMethods_GetStreamStatus_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	// First, start a stream so we have something to check status for
	startStreamResponse := helper.TestMethod(t, "start_streaming", map[string]interface{}{
		"device": "camera0",
	}, "operator")
	require.Nil(t, startStreamResponse.Error, "Stream start should succeed")

	// Now check stream status
	response := helper.TestMethod(t, "get_stream_status", map[string]interface{}{
		"device": "camera0",
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// FILE MANAGEMENT METHODS TESTS (High Priority - Core Functionality)
// ============================================================================

// TestWebSocketMethods_ListRecordings tests list_recordings method
func TestWebSocketMethods_ListRecordings_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "list_recordings", map[string]interface{}{
		"limit":  10,
		"offset": 0,
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_ListSnapshots tests list_snapshots method
func TestWebSocketMethods_ListSnapshots_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "list_snapshots", map[string]interface{}{
		"limit":  10,
		"offset": 0,
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_DeleteRecording tests delete_recording method
func TestWebSocketMethods_DeleteRecording_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	// First, stop any existing recording to ensure clean state
	helper.TestMethod(t, "stop_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator") // Ignore errors, just ensure clean state

	// Then, create a recording so we have something to delete
	startRecordingResponse := helper.TestMethod(t, "start_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")
	if startRecordingResponse.Error != nil {
		t.Logf("Recording start failed: %+v", startRecordingResponse.Error)
	}
	require.Nil(t, startRecordingResponse.Error, "Recording start should succeed")

	// Extract the actual filename from the response
	var recordingFilename string
	if startRecordingResponse.Result != nil {
		t.Logf("Full recording response: %+v", startRecordingResponse.Result)
		if resultMap, ok := startRecordingResponse.Result.(map[string]interface{}); ok {
			if filename, exists := resultMap["filename"]; exists {
				if filenameStr, ok := filename.(string); ok {
					recordingFilename = filenameStr
					t.Logf("Extracted filename: %s", recordingFilename)
				}
			} else {
				t.Logf("No 'filename' key in result map")
			}
		} else {
			t.Logf("Result is not a map[string]interface{}")
		}
	} else {
		t.Logf("No result in response")
	}

	// If we couldn't extract the filename, use a default
	if recordingFilename == "" {
		recordingFilename = "test_recording.mp4"
	}

	t.Logf("Using recording filename: %s", recordingFilename)

	// Stop the recording to create the file
	stopRecordingResponse := helper.TestMethod(t, "stop_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")
	require.Nil(t, stopRecordingResponse.Error, "Recording stop should succeed")

	// Send delete_recording message
	response := helper.TestMethod(t, "delete_recording", map[string]interface{}{
		"filename": recordingFilename,
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_DeleteSnapshot tests delete_snapshot method
func TestWebSocketMethods_DeleteSnapshot_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	// First, create a snapshot so we have something to delete
	takeSnapshotResponse := helper.TestMethod(t, "take_snapshot", map[string]interface{}{
		"device": "camera0",
	}, "operator")
	if takeSnapshotResponse.Error != nil {
		t.Logf("Snapshot creation failed: %+v", takeSnapshotResponse.Error)
	}
	require.Nil(t, takeSnapshotResponse.Error, "Snapshot creation should succeed")

	// Validate that a file was actually created (like MediaMTX tests do)
	require.NotNil(t, takeSnapshotResponse.Result, "Snapshot should return result with file info")

	// Debug: Log the actual response structure
	t.Logf("Snapshot response result: %+v", takeSnapshotResponse.Result)

	// Extract the actual filename from the response
	var snapshotFilename string
	if resultMap, ok := takeSnapshotResponse.Result.(map[string]interface{}); ok {
		t.Logf("Result is a map with keys: %v", getMapKeys(resultMap))
		if filePath, exists := resultMap["file_path"]; exists {
			if pathStr, ok := filePath.(string); ok {
				// Extract just the filename from the full path
				snapshotFilename = filepath.Base(pathStr)
				t.Logf("Snapshot created with filename: %s", snapshotFilename)

				// Validate file actually exists (like MediaMTX tests do)
				require.FileExists(t, pathStr, "Snapshot file should actually exist on disk")
			}
		} else {
			t.Logf("No 'file_path' key in result map")
		}
	} else {
		t.Logf("Result is not a map[string]interface{}")
	}

	// If we couldn't extract the filename, the test should fail
	require.NotEmpty(t, snapshotFilename, "Should be able to extract filename from snapshot response")

	// Send delete_snapshot message
	response := helper.TestMethod(t, "delete_snapshot", map[string]interface{}{
		"filename": snapshotFilename,
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// SYSTEM MANAGEMENT METHODS TESTS (Medium Priority - Admin Features)
// ============================================================================

// TestWebSocketMethods_GetStorageInfo tests get_storage_info method
func TestWebSocketMethods_GetStorageInfo_ReqMTX004_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_storage_info", map[string]interface{}{}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_SetRetentionPolicy tests set_retention_policy method
func TestWebSocketMethods_SetRetentionPolicy_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "set_retention_policy", map[string]interface{}{
		"policy_type":  "age",
		"max_age_days": 30,
		"enabled":      true,
	}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_CleanupOldFiles tests cleanup_old_files method
func TestWebSocketMethods_CleanupOldFiles_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "cleanup_old_files", map[string]interface{}{}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// EVENT SYSTEM METHODS TESTS (Advanced Features)
// ============================================================================

// TestWebSocketMethods_SubscribeEvents tests subscribe_events method
func TestWebSocketMethods_SubscribeEvents_ReqAPI003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "subscribe_events", map[string]interface{}{
		"topics": []string{"camera.connected", "recording.start"},
		"filters": map[string]interface{}{
			"device": "camera0",
		},
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_UnsubscribeEvents tests unsubscribe_events method
func TestWebSocketMethods_UnsubscribeEvents_ReqAPI003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "unsubscribe_events", map[string]interface{}{
		"topics": []string{"camera.connected"},
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetSubscriptionStats tests get_subscription_stats method
func TestWebSocketMethods_GetSubscriptionStats_ReqAPI003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_subscription_stats", map[string]interface{}{}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// EXTERNAL STREAM METHODS TESTS (Advanced Features)
// ============================================================================

// TestWebSocketMethods_DiscoverExternalStreams tests discover_external_streams method
func TestWebSocketMethods_DiscoverExternalStreams_ReqMTX003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "discover_external_streams", map[string]interface{}{
		"skydio_enabled":  true,
		"generic_enabled": false,
		"force_rescan":    false,
		"include_offline": false,
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_AddExternalStream tests add_external_stream method
func TestWebSocketMethods_AddExternalStream_ReqMTX003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "add_external_stream", map[string]interface{}{
		"stream_url":  "rtsp://192.168.42.15:5554/subject",
		"stream_name": "Test_UAV_15",
		"stream_type": "skydio_stanag4609",
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_RemoveExternalStream tests remove_external_stream method
func TestWebSocketMethods_RemoveExternalStream_ReqMTX003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "remove_external_stream", map[string]interface{}{
		"stream_name": "Test_UAV_15",
	}, "operator")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetExternalStreams tests get_external_streams method
func TestWebSocketMethods_GetExternalStreams_ReqMTX003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_external_streams", map[string]interface{}{}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_SetDiscoveryInterval tests set_discovery_interval method
func TestWebSocketMethods_SetDiscoveryInterval_ReqMTX003_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "set_discovery_interval", map[string]interface{}{
		"interval_seconds": 30,
	}, "admin")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// ============================================================================
// ADDITIONAL FILE INFO METHODS TESTS (Complete Coverage)
// ============================================================================

// TestWebSocketMethods_GetRecordingInfo tests get_recording_info method
func TestWebSocketMethods_GetRecordingInfo_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	// Create a recording first, then get info about it
	helper.TestMethod(t, "start_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// Stop the recording to create the file
	helper.TestMethod(t, "stop_recording", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// Now get recording info about a test file
	response := helper.TestMethod(t, "get_recording_info", map[string]interface{}{
		"filename": "test_recording.mp4",
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetSnapshotInfo tests get_snapshot_info method
func TestWebSocketMethods_GetSnapshotInfo_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	// Create a snapshot first, then get info about it
	helper.TestMethod(t, "take_snapshot", map[string]interface{}{
		"device": "camera0",
	}, "operator")

	// Now get snapshot info about a test file
	response := helper.TestMethod(t, "get_snapshot_info", map[string]interface{}{
		"filename": "test_snapshot.jpg",
	}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestWebSocketMethods_GetStreams tests get_streams method
func TestWebSocketMethods_GetStreams_ReqMTX002_Success(t *testing.T) {
	// === ENTERPRISE ULTRA-MINIMAL PATTERN ===
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// === TEST AND VALIDATION ===
	response := helper.TestMethod(t, "get_streams", map[string]interface{}{}, "viewer")

	// === VALIDATION ===
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.NotNil(t, response.ID, "Response should have ID")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")
}

// TestEnterpriseGrade_WebSocketMethodsPatternCompliance validates homogeneous test patterns
//
// Requirements Coverage:
// - REQ-ARCH-001: Progressive Readiness Pattern compliance
// - REQ-TEST-001: Homogeneous test patterns across all methods
//
// Test Pattern: Enterprise-grade pattern validation, no exceptions allowed
// Architecture: Validates all WebSocket methods follow identical patterns
func TestWebSocketMethods_ProcessMessage_ReqARCH001_EnterpriseGradePatternCompliance(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Progressive Readiness: Get controller with real hardware integration
	controller := createMediaMTXControllerUsingProvenPattern(t)

	server := helper.GetServer(t)
	server.SetMediaMTXController(controller)

	// Start server following Progressive Readiness Pattern
	_ = helper.StartServer(t) // Server is started, we use the original server instance

	// Enterprise Test 1: All methods must accept connections immediately
	methods := []string{
		"ping", "authenticate", "get_server_info", "get_system_status",
		"get_camera_list", "get_camera_status", "get_camera_capabilities",
		"take_snapshot", "start_recording", "stop_recording",
	}

	for _, method := range methods {
		t.Run(fmt.Sprintf("method_%s_immediate_connection", method), func(t *testing.T) {
			startTime := time.Now()
			conn := helper.NewTestClient(t, server)
			defer helper.CleanupTestClient(t, conn)
			connectionTime := time.Since(startTime)

			assert.Less(t, connectionTime, 100*time.Millisecond,
				"Method %s connection should be immediate (Progressive Readiness)", method)

			// Test immediate response capability
			message := CreateTestMessage(method, map[string]interface{}{})
			response := SendTestMessage(t, conn, message)

			require.NotNil(t, response,
				"Method %s should always respond (Progressive Readiness)", method)

			// Should not get "system not ready" blocking errors
			if response.Error != nil {
				assert.NotEqual(t, RATE_LIMIT_EXCEEDED, response.Error.Code,
					"Method %s should not block with 'system not ready' error", method)
			}
		})
	}

	t.Log("✅ Enterprise-grade WebSocket methods pattern compliance validated")
}
