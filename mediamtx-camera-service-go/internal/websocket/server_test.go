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
func TestWebSocketServer_Creation(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

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
func TestWebSocketServer_StartStop(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)
	assert.True(t, server.IsRunning(), "Server should be running after start")

	// Wait for server to be ready
	WaitForServerReady(t, server, 1*time.Second)

	// Test server stop
	err := server.Stop()
	require.NoError(t, err, "Server should stop successfully")
	assert.False(t, server.IsRunning(), "Server should not be running after stop")
}

// TestWebSocketServer_DoubleStart tests starting server twice
func TestWebSocketServer_DoubleStart(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)
	assert.True(t, server.IsRunning(), "Server should be running after first start")

	// Start server second time should fail
	err := server.Start()
	assert.Error(t, err, "Second start should fail")
	assert.True(t, server.IsRunning(), "Server should still be running")
}

// TestWebSocketServer_DoubleStop tests stopping server twice
func TestWebSocketServer_DoubleStop(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Stop server first time
	err := server.Stop()
	require.NoError(t, err, "First stop should succeed")

	// Stop server second time should not error
	err = server.Stop()
	assert.NoError(t, err, "Second stop should not error")
	assert.False(t, server.IsRunning(), "Server should not be running")
}

// TestWebSocketServer_ClientConnection tests client connection handling
func TestWebSocketServer_ClientConnection(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Test client connection
	server.clientsMutex.RLock()
	connectionCount := len(server.clients)
	server.clientsMutex.RUnlock()
	assert.Equal(t, 1, connectionCount, "Should have one client connection")

	// Close client connection
	err := conn.Close()
	require.NoError(t, err, "Client should close successfully")

	// Wait for connection cleanup
	time.Sleep(100 * time.Millisecond)
	server.clientsMutex.RLock()
	connectionCount = len(server.clients)
	server.clientsMutex.RUnlock()
	assert.Equal(t, 0, connectionCount, "Should have no client connections")
}

// TestWebSocketServer_MultipleClients tests multiple client connections
func TestWebSocketServer_MultipleClients(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect multiple clients
	conn1 := NewTestClient(t, server)
	defer CleanupTestClient(t, conn1)

	conn2 := NewTestClient(t, server)
	defer CleanupTestClient(t, conn2)

	conn3 := NewTestClient(t, server)
	defer CleanupTestClient(t, conn3)

	// Test multiple client connections
	server.clientsMutex.RLock()
	connectionCount := len(server.clients)
	server.clientsMutex.RUnlock()
	assert.Equal(t, 3, connectionCount, "Should have three client connections")

	// Close one client
	err := conn1.Close()
	require.NoError(t, err, "Client should close successfully")

	// Wait for connection cleanup
	time.Sleep(100 * time.Millisecond)
	server.clientsMutex.RLock()
	connectionCount = len(server.clients)
	server.clientsMutex.RUnlock()
	assert.Equal(t, 2, connectionCount, "Should have two client connections")
}

// TestWebSocketServer_MethodRegistration tests method registration
func TestWebSocketServer_MethodRegistration(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

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
func TestWebSocketServer_MethodExecution(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", NewTestLogger("test-jwt"))
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

// TestWebSocketServer_InvalidMethod tests invalid method handling
func TestWebSocketServer_InvalidMethod(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

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
func TestWebSocketServer_Notification(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Connect client
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

	// Send notification (no response expected)
	notification := CreateTestNotification("ping", map[string]interface{}{})
	SendTestNotification(t, conn, notification)

	// Test that no response is received
	// Set a short timeout for reading
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	var response JsonRpcResponse
	err := conn.ReadJSON(&response)
	assert.Error(t, err, "Should not receive response for notification")
}

// TestWebSocketServer_ContextCancellation tests server shutdown with context
func TestWebSocketServer_ContextCancellation(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine with context
	go func() {
		<-ctx.Done()
		server.Stop()
	}()

	// Cancel context
	cancel()

	// Wait for server to stop
	time.Sleep(100 * time.Millisecond)
	assert.False(t, server.IsRunning(), "Server should be stopped after context cancellation")
}

// TestWebSocketServer_ConcurrentConnections tests concurrent client connections
func TestWebSocketServer_ConcurrentConnections(t *testing.T) {
	server := NewTestWebSocketServer(t)
	defer CleanupTestServer(t, server)

	// Start server with proper dependencies (following main() pattern)
	StartTestServerWithDependencies(t, server)

	// Create multiple concurrent connections
	const numClients = 10
	connections := make([]*websocket.Conn, numClients)

	// Connect all clients
	for i := 0; i < numClients; i++ {
		conn := NewTestClient(t, server)
		connections[i] = conn
		defer CleanupTestClient(t, conn)
	}

	// Test all connections are established
	server.clientsMutex.RLock()
	connectionCount := len(server.clients)
	server.clientsMutex.RUnlock()
	assert.Equal(t, numClients, connectionCount, "Should have correct number of client connections")

	// Create a test JWT token for authentication
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", NewTestLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler")
	testToken := security.GenerateTestToken(t, jwtHandler, "test_user", "viewer")

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
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent message processing")
		}
	}
}

// TestWebSocketServer_NotificationFunctions tests notification functions for bugs
func TestWebSocketServer_NotificationFunctions(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	server := NewTestWebSocketServer(t)
	StartTestServerWithDependencies(t, server)
	defer CleanupTestServer(t, server)

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
func TestWebSocketServer_ErrorHandlingFunctions(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	server := NewTestWebSocketServer(t)
	StartTestServerWithDependencies(t, server)
	defer CleanupTestServer(t, server)

	// Test sendErrorResponse with edge cases - this might expose bugs
	conn := NewTestClient(t, server)
	defer CleanupTestClient(t, conn)

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
func TestWebSocketServer_PermissionAndRateLimit(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	server := NewTestWebSocketServer(t)
	StartTestServerWithDependencies(t, server)
	defer CleanupTestServer(t, server)

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
