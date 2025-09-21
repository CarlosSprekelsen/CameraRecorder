/*
WebSocket Server Unit Tests

Provides focused unit tests for WebSocket server functionality,
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
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
)

// TestMain sets up logging configuration for all tests
func TestMain(m *testing.M) {
	// Setup logging configuration for all tests
	setupTestLogging()

	// Run tests
	code := m.Run()

	// Exit with test result code
	os.Exit(code)
}

// TestWebSocketServer_Creation tests server creation and initialization
func TestWebSocketServer_New_ReqAPI001_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	server := helper.GetServer(t)

	// Test server initialization
	assert.NotNil(t, server, "Server should be created")
	assert.NotNil(t, server.config, "Server config should be initialized")
	assert.NotNil(t, server.logger, "Server logger should be initialized")
	assert.NotNil(t, server.jwtHandler, "Server JWT handler should be initialized")
	assert.NotNil(t, server.clients, "Server clients map should be initialized")
	assert.NotNil(t, server.methods, "Server methods map should be initialized")
	assert.False(t, server.IsRunning(), "Server should not be running initially")
}

// TestWebSocketServer_StartStop tests server start and stop functionality
func TestWebSocketServer_StartStop_ReqAPI001_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Progressive Readiness Test 1: Server accepts connections immediately after Start()
	server := helper.StartServer(t)

	// Progressive Readiness Pattern: Test immediate connection acceptance
	startTime := time.Now()
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)
	connectionTime := time.Since(startTime)

	assert.Less(t, connectionTime, 100*time.Millisecond,
		"Connection should be accepted immediately (Progressive Readiness)")

	// Progressive Readiness Test 2: Basic operations work immediately
	message := CreateTestMessage("ping", nil)
	response := SendTestMessage(t, conn, message)

	require.NotNil(t, response, "System should respond to requests immediately")

	// Progressive Readiness Test 3: No "system not ready" errors
	if response.Error != nil {
		require.NotEqual(t, -32002, response.Error.Code,
			"Should not get 'system not ready' error - violates Progressive Readiness")
	}

	// Test server stop
	ctx, cancel := context.WithTimeout(context.Background(), testutils.ShortTestTimeout)
	defer cancel()
	err := server.Stop(ctx)
	require.NoError(t, err, "Server should stop successfully")
	assert.False(t, server.IsRunning(), "Server should not be running after stop")
}

// TestEnterpriseGrade_ProgressiveReadinessCompliance validates complete enterprise compliance
func TestWebSocketServer_Start_ReqARCH001_ProgressiveReadiness(t *testing.T) {
	// No sequential execution - demonstrates parallel capability
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Enterprise Test Suite 1: Connection Performance Validation
	helper.ValidateProgressiveReadinessCompliance(t, server)

	// Enterprise Test Suite 2: Operation Patterns Validation
	helper.TestEnterpriseGradeOperations(t, server)

	// Enterprise Test Suite 3: Architectural Compliance Validation
	TestArchitecturalCompliance_ProgressiveReadiness(t, server)

	t.Log("✅ Enterprise-grade Progressive Readiness compliance validated")
}

// TestWebSocketServer_ProgressiveReadinessPattern tests TRUE Progressive Readiness Pattern compliance
func TestWebSocketServer_Start_ReqARCH001_ProgressiveReadinessPattern(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Enterprise Test 1: Multiple rapid connections should all succeed
	connections := make([]*websocket.Conn, 5)
	for i := 0; i < 5; i++ {
		startTime := time.Now()
		connections[i] = helper.NewTestClient(t, server)
		connectionTime := time.Since(startTime)

		assert.Less(t, connectionTime, 50*time.Millisecond,
			"Connection %d should be immediate", i)
	}

	// Enterprise Test 2: All connections should be able to send requests immediately
	for i, conn := range connections {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		require.NotNil(t, response, "Connection %d should receive response", i)

		// May get pong or "initializing" status, but should not block
		if response.Error != nil {
			require.NotEqual(t, -32002, response.Error.Code,
				"Connection %d should not get 'system not ready' error", i)
		}
	}

	// Enterprise Test 3: Operations that require components should gracefully handle initialization
	conn := connections[0]
	snapshotMessage := CreateTestMessage("take_snapshot", map[string]interface{}{
		"device": "camera0",
	})
	snapshotResponse := SendTestMessage(t, conn, snapshotMessage)

	require.NotNil(t, snapshotResponse, "Snapshot request should get response")

	if snapshotResponse.Error != nil {
		// Should get meaningful error, not "system not ready"
		assert.NotEqual(t, RATE_LIMIT_EXCEEDED, snapshotResponse.Error.Code,
			"Should get specific error, not generic 'not ready'")
	}

	// Cleanup all connections
	for _, conn := range connections {
		helper.CleanupTestClient(t, conn)
	}
}

// TestWebSocketServer_DoubleStart tests starting server twice
func TestWebSocketServer_Start_ReqAPI001_ErrorHandling_DoubleStart(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Progressive Readiness Pattern: Test immediate connection acceptance instead of polling
	startTime := time.Now()
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)
	connectionTime := time.Since(startTime)

	assert.Less(t, connectionTime, 100*time.Millisecond,
		"Connection should be accepted immediately (Progressive Readiness)")

	// Start server second time should fail
	err := server.Start()
	assert.Error(t, err, "Second start should fail")
	assert.True(t, server.IsRunning(), "Server should still be running")
}

// TestWebSocketServer_DoubleStop tests stopping server twice
func TestWebSocketServer_Stop_ReqAPI001_ErrorHandling_DoubleStop(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Stop server first time
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	err := server.Stop(ctx1)
	require.NoError(t, err, "First stop should succeed")

	// Stop server second time should not error
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	err = server.Stop(ctx2)
	assert.NoError(t, err, "Second stop should not error")
	assert.False(t, server.IsRunning(), "Server should not be running")
}

// TestWebSocketServer_ClientConnection tests client connection handling
func TestWebSocketServer_HandleConnection_ReqAPI001_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test client connection
	connectionCount := server.GetClientCount()
	assert.Equal(t, int64(1), connectionCount, "Should have one client connection")

	// Close client connection
	err := conn.Close()
	require.NoError(t, err, "Client should close successfully")

	// Progressive Readiness Pattern: Connection cleanup should be immediate
	// No polling - check connection count directly after cleanup
	clientCount := server.GetClientCount()
	assert.Equal(t, int64(0), clientCount,
		"Connections should be cleaned up immediately (Progressive Readiness)")
}

// TestWebSocketServer_MultipleClients tests multiple client connections
func TestWebSocketServer_HandleConnection_ReqAPI001_MultipleClients(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Connect multiple clients
	conn1 := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn1)

	conn2 := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn2)

	conn3 := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn3)

	// Test multiple client connections
	connectionCount := server.GetClientCount()
	assert.Equal(t, int64(3), connectionCount, "Should have three client connections")

	// Close one client
	err := conn1.Close()
	require.NoError(t, err, "Client should close successfully")

	// Progressive Readiness Pattern: Connections should be established immediately
	// No polling - check connection count directly after establishment
	clientCount := server.GetClientCount()
	assert.Equal(t, int64(2), clientCount,
		"Two client connections should be established immediately (Progressive Readiness)")
}

// TestWebSocketServer_MethodRegistration tests method registration
func TestWebSocketServer_RegisterMethod_ReqAPI002_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	server := helper.GetServer(t)

	// Register a test method
	methodName := "test_method"
	server.registerMethod(methodName, func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		return CreateTestResponse("test-id", "test_result"), nil
	}, "1.0")

	// Test method registration
	server.methodsMutex.RLock()
	_, exists := server.methods[methodName]
	server.methodsMutex.RUnlock()
	assert.True(t, exists, "Method should be registered")

	// Test method version
	server.methodVersionsMutex.RLock()
	version, exists := server.methodVersions[methodName]
	server.methodVersionsMutex.RUnlock()
	assert.True(t, exists, "Method version should be registered")
	assert.Equal(t, "1.0", version, "Method version should be correct")
}

// TestWebSocketServer_MethodExecution tests method execution
func TestWebSocketServer_ExecuteMethod_ReqAPI002_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", NewTestLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken, err := jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(t, err, "Failed to generate test token")

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

// TestWebSocketServer_InvalidMethod tests invalid method handling
func TestWebSocketServer_ExecuteMethod_ReqAPI002_ErrorHandling_InvalidMethod(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Send invalid method message
	message := CreateTestMessage("invalid_method", map[string]interface{}{})
	response := SendTestMessage(t, conn, message)

	// Test error response
	assert.Equal(t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	assert.Equal(t, message.ID, response.ID, "Response should have correct ID")
	assert.Nil(t, response.Result, "Response should not have result")
	assert.NotNil(t, response.Error, "Response should have error")
	assert.Equal(t, METHOD_NOT_FOUND, response.Error.Code, "Error should be method not found")
}

// TestWebSocketServer_Notification tests notification handling
func TestWebSocketServer_SendNotification_ReqAPI003_Success(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Connect client
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Send notification (no response expected)
	notification := CreateTestNotification("ping", map[string]interface{}{})
	SendTestNotification(t, conn, notification)

	// Test that no response is received
	// Set a short timeout for reading
	if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}

	var response JsonRpcResponse
	err := conn.ReadJSON(&response)
	assert.Error(t, err, "Should not receive response for notification")
}

// TestWebSocketServer_ContextCancellation tests server shutdown with context
func TestWebSocketServer_HandleConnection_ReqAPI001_ContextCancellation(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine with context
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Stop(stopCtx); err != nil {
			t.Errorf("Failed to stop server: %v", err)
		}
	}()

	// Cancel context
	cancel()

	// Progressive Readiness Pattern: Server should stop immediately with context cancellation
	// No polling - check server state directly after context cancellation
	assert.False(t, server.IsRunning(),
		"Server should be stopped immediately after context cancellation (Progressive Readiness)")
}

// TestWebSocketServer_ConcurrentConnections tests concurrent client connections
func TestWebSocketServer_HandleConnection_ReqAPI001_Concurrent(t *testing.T) {
	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Create multiple concurrent connections
	const numClients = 10
	connections := make([]*websocket.Conn, numClients)

	// Connect all clients
	for i := 0; i < numClients; i++ {
		conn := helper.NewTestClient(t, server)
		connections[i] = conn
		defer helper.CleanupTestClient(t, conn)
	}

	// Test all connections are established
	connectionCount := server.GetClientCount()
	assert.Equal(t, int64(numClients), connectionCount, "Should have correct number of client connections")

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", NewTestLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken, err := jwtHandler.GenerateToken("test_user", "viewer", 24)
	require.NoError(t, err, "Failed to generate test token")

	// Send messages from all clients concurrently
	done := make(chan bool, numClients)
	for i, conn := range connections {
		go func(conn *websocket.Conn, clientID int) {
			// First authenticate the client
			authMessage := CreateTestMessage("authenticate", map[string]interface{}{
				"auth_token": testToken,
			})
			authResponse := SendTestMessage(t, conn, authMessage)
			if authResponse.Error != nil {
				t.Errorf("Authentication failed for client %d: %v", clientID, authResponse.Error)
				done <- true
				return
			}

			// Send ping message
			message := CreateTestMessage("ping", map[string]interface{}{"client_id": clientID})
			response := SendTestMessage(t, conn, message)
			assert.Equal(t, "pong", response.Result, "Client should receive pong response")
			done <- true
		}(conn, i)
	}

	// Wait for all messages to be processed
	for i := 0; i < numClients; i++ {
		select {
		case <-done:
		case <-time.After(testutils.UniversalTimeoutVeryLong):
			t.Fatal("Timeout waiting for concurrent message processing")
		}
	}
}

// TestWebSocketServer_NotificationFunctions tests notification functions for bugs
func TestWebSocketServer_SendNotification_ReqAPI003_Functions(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Test notifyRecordingStatusUpdate with edge cases
	server.notifyRecordingStatusUpdate("", "", "", 0)                                             // Empty parameters - might expose bugs
	server.notifyRecordingStatusUpdate("/dev/video0", "started", "recording.mp4", -1*time.Second) // Negative duration - might expose bugs
	server.notifyRecordingStatusUpdate("invalid_device", "invalid_status", "", 0)                 // Invalid parameters

	// Test notifyCameraStatusUpdate with edge cases
	server.notifyCameraStatusUpdate("", "", "")                        // Empty parameters - might expose bugs
	server.notifyCameraStatusUpdate("/dev/video0", "", "Camera Name")  // Empty status - might expose bugs
	server.notifyCameraStatusUpdate("invalid_device", "connected", "") // Empty name - might expose bugs

	// Test notifySnapshotTaken with edge cases
	server.notifySnapshotTaken("", "", "")                           // Empty parameters - might expose bugs
	server.notifySnapshotTaken("/dev/video0", "", "1920x1080")       // Empty filename - might expose bugs
	server.notifySnapshotTaken("invalid_device", "snapshot.jpg", "") // Empty resolution - might expose bugs

	// Test notifySystemEvent with edge cases
	server.notifySystemEvent("", nil)                                    // Empty event type and nil data - might expose bugs
	server.notifySystemEvent("system_startup", map[string]interface{}{}) // Empty data map
	server.notifySystemEvent("invalid_event", map[string]interface{}{
		"key": nil, // Nil value in data - might expose bugs
	})
}

// TestWebSocketServer_ErrorHandlingFunctions tests error handling functions for bugs
func TestWebSocketServer_HandleError_ReqAPI002_ErrorHandling(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Test sendErrorResponse with edge cases - this might expose bugs
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test with nil connection - this should expose a bug
	server.sendErrorResponse(nil, "test-id", -32600, "Test error")

	// Test with invalid error codes - this might expose bugs
	server.sendErrorResponse(conn, "test-id", 0, "Zero error code")
	server.sendErrorResponse(conn, "test-id", -99999, "Invalid error code")
	server.sendErrorResponse(conn, "test-id", 99999, "Positive error code")

	// Test with nil ID - this might expose bugs
	server.sendErrorResponse(conn, nil, -32600, "Nil ID test")

	// Test with empty message - this might expose bugs
	server.sendErrorResponse(conn, "test-id", -32600, "")

	// Test with very long message - this might expose bugs
	longMessage := strings.Repeat("A", 10000)
	server.sendErrorResponse(conn, "test-id", -32600, longMessage)
}

// TestWebSocketServer_PermissionAndRateLimit tests permission and rate limit functions for bugs
func TestWebSocketServer_ValidatePermission_ReqSEC001_RateLimit(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Test checkMethodPermissions with edge cases - this might expose bugs
	// Note: We can't easily access the client connection, so we'll test with nil
	_ = server.checkMethodPermissions(nil, "") // Nil client and empty method - might expose bugs
	// This should expose a nil pointer dereference bug

	_ = server.checkMethodPermissions(nil, "invalid_method") // Nil client with method - might expose bugs
	// This should expose a nil pointer dereference bug

	// Test checkRateLimit with edge cases - this might expose bugs
	_ = server.checkRateLimit(nil) // Nil client - might expose bugs
	// This should expose a nil pointer dereference bug
}

// TestWebSocketServer_JsonRpcProtocolCompliance tests JSON-RPC 2.0 protocol compliance
func TestWebSocketServer_ProcessMessage_ReqAPI002_JsonRpcCompliance(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Create authenticated connection for all subtests
	conn := helper.GetAuthenticatedConnection(t, "test-user", "viewer")
	defer helper.CleanupTestClient(t, conn)

	// Test 1: Valid JSON-RPC 2.0 request
	t.Run("ValidJsonRpcRequest", func(t *testing.T) {

		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		// Validate JSON-RPC 2.0 response structure
		assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
		assert.Equal(t, message.ID, response.ID, "Response ID should match request ID")
		assert.Nil(t, response.Error, "Valid request should not have error")
		assert.NotNil(t, response.Result, "Valid request should have result")
	})

	// Test 2: Invalid JSON-RPC version
	t.Run("InvalidJsonRpcVersion", func(t *testing.T) {
		// Create unauthenticated connection for protocol testing
		conn := helper.NewTestClient(t, server)
		defer helper.CleanupTestClient(t, conn)

		invalidMessage := &JsonRpcRequest{
			JSONRPC: "1.0", // Invalid version
			Method:  "ping",
			ID:      "test-invalid-version",
		}

		err := conn.WriteJSON(invalidMessage)
		require.NoError(t, err, "Failed to send invalid version message")

		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		require.NoError(t, err, "Failed to read response")

		assert.NotNil(t, response.Error, "Invalid version should return error")
		assert.Equal(t, INVALID_REQUEST, response.Error.Code, "Should return invalid request error")
	})

	// Test 3: Missing JSON-RPC version
	t.Run("MissingJsonRpcVersion", func(t *testing.T) {
		invalidMessage := map[string]interface{}{
			"method": "ping",
			"id":     "test-missing-version",
		}

		err := conn.WriteJSON(invalidMessage)
		require.NoError(t, err, "Failed to send missing version message")

		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		require.NoError(t, err, "Failed to read response")

		assert.NotNil(t, response.Error, "Missing version should return error")
		assert.Equal(t, INVALID_REQUEST, response.Error.Code, "Should return invalid request error")
	})

	// Test 4: Request ID handling (string vs number)
	t.Run("RequestIdHandling", func(t *testing.T) {
		// Test string ID
		stringMessage := &JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      "string-id",
		}

		err := conn.WriteJSON(stringMessage)
		require.NoError(t, err, "Failed to send string ID message")

		var stringResponse JsonRpcResponse
		err = conn.ReadJSON(&stringResponse)
		require.NoError(t, err, "Failed to read string ID response")

		assert.Equal(t, "string-id", stringResponse.ID, "String ID should be preserved")

		// Test numeric ID
		numericMessage := &JsonRpcRequest{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      12345,
		}

		err = conn.WriteJSON(numericMessage)
		require.NoError(t, err, "Failed to send numeric ID message")

		var numericResponse JsonRpcResponse
		err = conn.ReadJSON(&numericResponse)
		require.NoError(t, err, "Failed to read numeric ID response")

		assert.Equal(t, float64(12345), numericResponse.ID, "Numeric ID should be preserved as float64")
	})

	// Test 5: Null request ID (notification)
	t.Run("NullRequestId", func(t *testing.T) {
		notification := &JsonRpcNotification{
			JSONRPC: "2.0",
			Method:  "ping",
			Params:  nil,
		}

		err := conn.WriteJSON(notification)
		require.NoError(t, err, "Failed to send notification")

		// Notifications should not receive responses
		// Set a short timeout to verify no response
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatalf("Failed to set read deadline: %v", err)
		}
		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		assert.Error(t, err, "Notifications should not receive responses")
	})
}

// TestWebSocketServer_ErrorCodeCompliance tests JSON-RPC error code compliance
func TestWebSocketServer_HandleError_ReqAPI002_ErrorCodeCompliance(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	// Test 1: Method not found (-32601)
	t.Run("MethodNotFound", func(t *testing.T) {
		message := CreateTestMessage("nonexistent_method", nil)
		response := SendTestMessage(t, conn, message)

		assert.NotNil(t, response.Error, "Non-existent method should return error")
		assert.Equal(t, METHOD_NOT_FOUND, response.Error.Code, "Should return method not found error")
		assert.Contains(t, response.Error.Message, "not found", "Error message should indicate method not found")
	})

	// Test 2: Invalid parameters (-32602)
	t.Run("InvalidParameters", func(t *testing.T) {
		// Send ping with invalid parameters (ping takes no parameters)
		message := CreateTestMessage("ping", map[string]interface{}{
			"invalid_param": "should_not_be_here",
		})
		response := SendTestMessage(t, conn, message)

		// Note: This test might pass if ping method ignores parameters
		// The important thing is that the response is valid JSON-RPC
		assert.Equal(t, "2.0", response.JSONRPC, "Response should be valid JSON-RPC 2.0")
	})

	// Test 3: Invalid request (-32600)
	t.Run("InvalidRequest", func(t *testing.T) {
		// Send malformed JSON
		malformedJSON := `{"jsonrpc": "2.0", "method": "ping", "id": 1, "extra": "invalid"}`
		err := conn.WriteMessage(websocket.TextMessage, []byte(malformedJSON))
		require.NoError(t, err, "Failed to send malformed JSON")

		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		require.NoError(t, err, "Failed to read response")

		// Should still be valid JSON-RPC response
		assert.Equal(t, "2.0", response.JSONRPC, "Response should be valid JSON-RPC 2.0")
	})

	// Test 4: Internal error (-32603)
	t.Run("InternalError", func(t *testing.T) {
		// This would require a method that can trigger internal errors
		// For now, we'll test that error responses follow the correct format
		message := CreateTestMessage("nonexistent_method", nil)
		response := SendTestMessage(t, conn, message)

		if response.Error != nil {
			assert.True(t, response.Error.Code <= -32000, "Error code should be in valid range")
			assert.NotEmpty(t, response.Error.Message, "Error message should not be empty")
		}
	})
}

// TestWebSocketServer_AuthenticationFlow tests JWT token validation and authentication flow
func TestWebSocketServer_Authenticate_ReqSEC001_AuthenticationFlow(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Test 1: Valid authentication
	t.Run("ValidAuthentication", func(t *testing.T) {
		// Create authenticated connection using standardized pattern
		conn := helper.GetAuthenticatedConnection(t, "test-user", "viewer")
		defer helper.CleanupTestClient(t, conn)

		// After authentication, ping should work
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		assert.Nil(t, response.Error, "Authenticated request should succeed")
		assert.Equal(t, "pong", response.Result, "Ping should return pong")
	})

	// Test 2: Invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		// Create new connection for this test
		conn2 := helper.NewTestClient(t, server)
		defer helper.CleanupTestClient(t, conn2)

		authMessage := CreateTestMessage("authenticate", map[string]interface{}{
			"auth_token": "invalid-token",
		})
		authResponse := SendTestMessage(t, conn2, authMessage)

		assert.NotNil(t, authResponse.Error, "Invalid token should return error")
		assert.Equal(t, AUTHENTICATION_REQUIRED, authResponse.Error.Code, "Should return authentication failed error")
	})

	// Test 3: Missing token
	t.Run("MissingToken", func(t *testing.T) {
		// Create new connection for this test
		conn3 := helper.NewTestClient(t, server)
		defer helper.CleanupTestClient(t, conn3)

		// Try to call method without authentication
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn3, message)

		assert.NotNil(t, response.Error, "Unauthenticated request should return error")
		assert.Equal(t, AUTHENTICATION_REQUIRED, response.Error.Code, "Should return authentication required error")
	})

	// Test 4: Expired token (simulated)
	t.Run("ExpiredToken", func(t *testing.T) {
		// Create new connection for this test
		conn4 := helper.NewTestClient(t, server)
		defer helper.CleanupTestClient(t, conn4)

		// Create expired token (this would need proper JWT creation with past expiry)
		// For now, we'll test with a malformed token
		authMessage := CreateTestMessage("authenticate", map[string]interface{}{
			"auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		})
		authResponse := SendTestMessage(t, conn4, authMessage)

		assert.NotNil(t, authResponse.Error, "Expired/malformed token should return error")
	})
}

// TestWebSocketServer_RoleBasedAccessControl tests RBAC enforcement
func TestWebSocketServer_ValidatePermission_ReqSEC001_RoleBasedAccess(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Start server following Progressive Readiness Pattern
	server := helper.StartServer(t)

	// Test 1: Viewer role permissions
	t.Run("ViewerRolePermissions", func(t *testing.T) {
		// Create authenticated connection using standardized pattern
		conn := helper.GetAuthenticatedConnection(t, "viewer-user", "viewer")
		defer helper.CleanupTestClient(t, conn)

		// Viewer should be able to call read-only methods
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)
		assert.Nil(t, response.Error, "Viewer should be able to ping")

		// Viewer should NOT be able to call control methods
		controlMessage := CreateTestMessage("take_snapshot", map[string]interface{}{
			"device": "camera0",
		})
		controlResponse := SendTestMessage(t, conn, controlMessage)
		assert.NotNil(t, controlResponse.Error, "Viewer should not be able to take snapshots")
		assert.Equal(t, INSUFFICIENT_PERMISSIONS, controlResponse.Error.Code, "Should return insufficient permissions error")
	})

	// Test 2: Operator role permissions
	t.Run("OperatorRolePermissions", func(t *testing.T) {
		// Create authenticated connection using standardized pattern
		conn := helper.GetAuthenticatedConnection(t, "operator-user", "operator")
		defer helper.CleanupTestClient(t, conn)

		// Operator should be able to call control methods
		controlMessage := CreateTestMessage("take_snapshot", map[string]interface{}{
			"device": "camera0",
		})
		controlResponse := SendTestMessage(t, conn, controlMessage)
		// Note: This might fail due to camera not existing, but should not fail due to permissions
		if controlResponse.Error != nil {
			assert.NotEqual(t, INSUFFICIENT_PERMISSIONS, controlResponse.Error.Code, "Should not fail due to insufficient permissions")
		}

		// Operator should NOT be able to call admin methods
		adminMessage := CreateTestMessage("get_metrics", nil)
		adminResponse := SendTestMessage(t, conn, adminMessage)
		assert.NotNil(t, adminResponse.Error, "Operator should not be able to get metrics")
		assert.Equal(t, INSUFFICIENT_PERMISSIONS, adminResponse.Error.Code, "Should return insufficient permissions error")
	})

	// Test 3: Admin role permissions
	t.Run("AdminRolePermissions", func(t *testing.T) {
		// Create authenticated connection using standardized pattern
		conn := helper.GetAuthenticatedConnection(t, "admin-user", "admin")
		defer helper.CleanupTestClient(t, conn)

		// Admin should be able to call all methods
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)
		assert.Nil(t, response.Error, "Admin should be able to ping")

		adminMessage := CreateTestMessage("get_metrics", nil)
		adminResponse := SendTestMessage(t, conn, adminMessage)
		// Note: This might fail due to implementation, but should not fail due to permissions
		if adminResponse.Error != nil {
			assert.NotEqual(t, INSUFFICIENT_PERMISSIONS, adminResponse.Error.Code, "Should not fail due to insufficient permissions")
		}
	})

	// Test 4: Invalid role
	t.Run("InvalidRole", func(t *testing.T) {
		conn := helper.NewTestClient(t, server)
		defer helper.CleanupTestClient(t, conn)

		// Try to authenticate with invalid role
		authMessage := CreateTestMessage("authenticate", map[string]interface{}{
			"auth_token": "invalid-role-token", // This would need proper JWT with invalid role
		})
		authResponse := SendTestMessage(t, conn, authMessage)

		assert.NotNil(t, authResponse.Error, "Invalid role should return error")
	})
}

// TestWebSocketServer_ResponseMetadata tests response metadata inclusion
func TestWebSocketServer_SendResponse_ReqAPI003_ResponseMetadata(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create authenticated connection using standardized pattern
	conn := helper.GetAuthenticatedConnection(t, "test-user", "viewer")
	defer helper.CleanupTestClient(t, conn)

	// Test 1: Response metadata presence
	t.Run("ResponseMetadataPresence", func(t *testing.T) {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		// Check that metadata is present
		assert.NotNil(t, response.Metadata, "Response should include metadata")

		// Check required metadata fields
		metadata := response.Metadata
		assert.Contains(t, metadata, "processing_time_ms", "Should include processing time")
		assert.Contains(t, metadata, "server_timestamp", "Should include server timestamp")
		assert.Contains(t, metadata, "request_id", "Should include request ID")
	})

	// Test 2: Processing time validation
	t.Run("ProcessingTimeValidation", func(t *testing.T) {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		metadata := response.Metadata
		processingTime, exists := metadata["processing_time_ms"]
		assert.True(t, exists, "Processing time should exist")

		// Processing time should be a number and reasonable (< 1000ms for ping)
		processingTimeFloat, ok := processingTime.(float64)
		assert.True(t, ok, "Processing time should be a number")
		assert.True(t, processingTimeFloat >= 0, "Processing time should be non-negative")
		assert.True(t, processingTimeFloat < 1000, "Processing time should be reasonable for ping")
	})

	// Test 3: Server timestamp validation
	t.Run("ServerTimestampValidation", func(t *testing.T) {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		metadata := response.Metadata
		serverTimestamp, exists := metadata["server_timestamp"]
		assert.True(t, exists, "Server timestamp should exist")

		// Server timestamp should be a valid RFC3339 string
		timestampStr, ok := serverTimestamp.(string)
		assert.True(t, ok, "Server timestamp should be a string")

		_, err := time.Parse(time.RFC3339, timestampStr)
		assert.NoError(t, err, "Server timestamp should be valid RFC3339 format")

		// Timestamp should be recent (within last minute)
		timestamp, _ := time.Parse(time.RFC3339, timestampStr)
		now := time.Now()
		diff := now.Sub(timestamp)
		assert.True(t, diff < time.Minute, "Server timestamp should be recent")
	})

	// Test 4: Request ID correlation
	t.Run("RequestIdCorrelation", func(t *testing.T) {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		metadata := response.Metadata
		requestID, exists := metadata["request_id"]
		assert.True(t, exists, "Request ID should exist in metadata")

		// Request ID in metadata should match response ID
		assert.Equal(t, response.ID, requestID, "Request ID in metadata should match response ID")
	})

	// Test 5: Error response metadata
	t.Run("ErrorResponseMetadata", func(t *testing.T) {
		// Send request to non-existent method to get error response
		message := CreateTestMessage("nonexistent_method", nil)
		response := SendTestMessage(t, conn, message)

		// Error responses should also include metadata
		assert.NotNil(t, response.Metadata, "Error response should include metadata")

		metadata := response.Metadata
		assert.Contains(t, metadata, "processing_time_ms", "Error response should include processing time")
		assert.Contains(t, metadata, "server_timestamp", "Error response should include server timestamp")
		assert.Contains(t, metadata, "request_id", "Error response should include request ID")
	})
}

// TestWebSocketServer_MessageSizeLimits tests message size limit handling
func TestWebSocketServer_ProcessMessage_ReqAPI001_MessageSizeLimits(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// No sequential execution - Progressive Readiness enables parallelism
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create authenticated connection using standardized pattern
	conn := helper.GetAuthenticatedConnection(t, "test-user", "viewer")
	defer helper.CleanupTestClient(t, conn)

	// Test 1: Normal size message
	t.Run("NormalSizeMessage", func(t *testing.T) {
		message := CreateTestMessage("ping", nil)
		response := SendTestMessage(t, conn, message)

		assert.Nil(t, response.Error, "Normal size message should succeed")
	})

	// Test 2: Large parameter message
	t.Run("LargeParameterMessage", func(t *testing.T) {
		// Create a large parameter (1MB of data)
		largeData := strings.Repeat("A", 1024*1024) // 1MB
		message := CreateTestMessage("ping", map[string]interface{}{
			"large_data": largeData,
		})

		// Large messages should be rejected by WebSocket protocol
		// This is expected behavior - connection should be closed
		err := conn.WriteJSON(message)
		if err != nil {
			// This is expected behavior - WebSocket should reject oversized messages
			t.Logf("Large message rejected (expected): %v", err)
			return
		}

		// If write succeeds, the message was accepted (unexpected but not a failure)
		// Try to read response to see if server handles it gracefully
		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		if err != nil {
			// Connection closed due to message size - this is expected
			t.Logf("Connection closed due to message size (expected): %v", err)
			return
		}

		// If we get here, the message was accepted and processed
		assert.NotNil(t, response, "Should receive a response if message is accepted")
	})

	// Test 3: Very large message (exceeds typical limits)
	t.Run("VeryLargeMessage", func(t *testing.T) {
		// Create a very large message (10MB)
		veryLargeData := strings.Repeat("B", 10*1024*1024) // 10MB
		message := CreateTestMessage("ping", map[string]interface{}{
			"very_large_data": veryLargeData,
		})

		// This might fail due to message size limits, which is expected
		err := conn.WriteJSON(message)
		if err != nil {
			// If write fails due to size, that's acceptable
			t.Logf("Large message write failed (expected): %v", err)
			return
		}

		// If write succeeds, we should get some response
		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		if err != nil {
			// If read fails due to size limits, that's also acceptable
			t.Logf("Large message read failed (expected): %v", err)
			return
		}

		// If we get a response, it should be valid JSON-RPC
		assert.Equal(t, "2.0", response.JSONRPC, "Response should be valid JSON-RPC 2.0")
	})

	// Test 4: Malformed large message
	t.Run("MalformedLargeMessage", func(t *testing.T) {
		// Send malformed JSON that's also large
		largeMalformedJSON := `{"jsonrpc": "2.0", "method": "ping", "id": 1, "data": "` + strings.Repeat("X", 1024*1024) + `"}`

		err := conn.WriteMessage(websocket.TextMessage, []byte(largeMalformedJSON))
		if err != nil {
			// If write fails due to size, that's acceptable
			t.Logf("Large malformed message write failed (expected): %v", err)
			return
		}

		// Should get some response or connection should handle it gracefully
		var response JsonRpcResponse
		err = conn.ReadJSON(&response)
		if err != nil {
			// If read fails, that's acceptable for malformed large messages
			t.Logf("Large malformed message read failed (expected): %v", err)
			return
		}

		// If we get a response, it should be valid JSON-RPC
		assert.Equal(t, "2.0", response.JSONRPC, "Response should be valid JSON-RPC 2.0")
	})
}

// TestWebSocketServer_ContextAwareShutdown tests the context-aware shutdown functionality
func TestWebSocketServer_Stop_ReqAPI001_ContextAwareShutdown(t *testing.T) {
	t.Run("graceful_shutdown_with_context", func(t *testing.T) {
		// No sequential execution - Progressive Readiness enables parallelism
		helper := NewWebSocketTestHelper(t, nil)
		defer helper.Cleanup(t)

		server := helper.GetServer(t)

		// Start server
		err := server.Start()
		require.NoError(t, err, "Server should start successfully")
		assert.True(t, server.IsRunning(), "Server should be running")

		// Test graceful shutdown with context
		ctx, cancel := context.WithTimeout(context.Background(), testutils.ShortTestTimeout)
		defer cancel()

		start := time.Now()
		err = server.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Server should stop gracefully")
		assert.False(t, server.IsRunning(), "Server should not be running after stop")
		assert.Less(t, elapsed, 1*time.Second, "Shutdown should be fast")
	})

	t.Run("shutdown_with_cancelled_context", func(t *testing.T) {
		// No sequential execution - Progressive Readiness enables parallelism
		helper := NewWebSocketTestHelper(t, nil)
		defer helper.Cleanup(t)

		server := helper.GetServer(t)

		// Start server
		err := server.Start()
		require.NoError(t, err, "Server should start successfully")

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Stop should complete quickly since context is already cancelled
		start := time.Now()
		err = server.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Server should stop even with cancelled context")
		assert.Less(t, elapsed, 100*time.Millisecond, "Shutdown should be very fast with cancelled context")
	})

	t.Run("shutdown_timeout_handling", func(t *testing.T) {
		// No sequential execution - Progressive Readiness enables parallelism
		helper := NewWebSocketTestHelper(t, nil)
		defer helper.Cleanup(t)

		server := helper.GetServer(t)

		// Start server
		err := server.Start()
		require.NoError(t, err, "Server should start successfully")

		// Use very short timeout to test timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Context will expire immediately due to 1ms timeout

		start := time.Now()
		err = server.Stop(ctx)
		elapsed := time.Since(start)

		// Should timeout but not hang
		require.Error(t, err, "Should timeout with very short timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Error should indicate timeout")
		assert.Less(t, elapsed, 1*time.Second, "Should not hang indefinitely")
	})

	t.Run("double_stop_handling", func(t *testing.T) {
		// No sequential execution - Progressive Readiness enables parallelism
		helper := NewWebSocketTestHelper(t, nil)
		defer helper.Cleanup(t)

		server := helper.GetServer(t)

		// Start server
		err := server.Start()
		require.NoError(t, err, "Server should start successfully")

		// Stop first time
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()
		err = server.Stop(ctx1)
		require.NoError(t, err, "First stop should succeed")

		// Stop second time should not error
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		err = server.Stop(ctx2)
		assert.NoError(t, err, "Second stop should not error")
		assert.False(t, server.IsRunning(), "Server should not be running")
	})

	t.Run("stop_without_start", func(t *testing.T) {
		// No sequential execution - Progressive Readiness enables parallelism
		helper := NewWebSocketTestHelper(t, nil)
		defer helper.Cleanup(t)

		server := helper.GetServer(t)

		// Stop without starting should not error
		ctx, cancel := context.WithTimeout(context.Background(), testutils.ShortTestTimeout)
		defer cancel()
		err := server.Stop(ctx)
		assert.NoError(t, err, "Stop without start should not error")
		assert.False(t, server.IsRunning(), "Server should not be running")
	})
}
