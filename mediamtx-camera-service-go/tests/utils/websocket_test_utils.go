/*
WebSocket Test Utilities - Message Utilities Only

Provides message utilities for WebSocket testing.
For WebSocket server creation, use tests/testutils/websocket_server.go
This file contains only message utilities.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit/Integration/E2E
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package testutils

import (
	"fmt"
	"net"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// Message utilities for WebSocket testing - no server management here

// GetFreePort returns a free port for testing using port 0 for automatic OS assignment
func GetFreePort() int {
	// Use port 0 to let OS assign next available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 8002 // fallback
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Small delay to ensure port is released
	// REMOVED: time.Sleep() violation of Progressive Readiness Pattern
	// Server accepts connections immediately after Start() per architectural requirements

	return port
}

// Server management functions removed - use tests/testutils/websocket_server.go instead

// NewTestClient creates a test WebSocket client connection
func NewTestClient(t *testing.T, server *websocket.WebSocketServer) *gorilla.Conn {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// Start server if needed and connect directly
	if !server.IsRunning() {
		err := server.Start()
		require.NoError(t, err, "Failed to start test server")
	}
	url := fmt.Sprintf("ws://localhost:%d/ws", server.GetConfig().Port)
	conn, _, err := gorilla.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "Failed to connect to test server")
	return conn
}

// CreateTestMessage creates a test JSON-RPC message
// FIXED: Implement locally to avoid undefined references
func CreateTestMessage(method string, params map[string]interface{}) *websocket.JsonRpcRequest {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation
	return &websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

// CreateTestNotification creates a test JSON-RPC notification
// FIXED: Implement locally to avoid undefined references
func CreateTestNotification(method string, params map[string]interface{}) *websocket.JsonRpcNotification {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation
	return &websocket.JsonRpcNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// SendTestMessage sends a test message and waits for response
// FIXED: Implement locally using gorilla WebSocket
func SendTestMessage(t *testing.T, conn *gorilla.Conn, message *websocket.JsonRpcRequest) *websocket.JsonRpcResponse {
	// REQ-API-003: Request/response message handling

	// Send message
	err := conn.WriteJSON(message)
	require.NoError(t, err, "Failed to send test message")

	// Read response
	var response websocket.JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err, "Failed to read test response")

	return &response
}

// SendTestNotification sends a test notification (no response expected)
// FIXED: Implement locally using gorilla WebSocket
func SendTestNotification(t *testing.T, conn *gorilla.Conn, notification *websocket.JsonRpcNotification) {
	// REQ-API-003: Request/response message handling

	// Send notification
	err := conn.WriteJSON(notification)
	require.NoError(t, err, "Failed to send test notification")
}

// AssertServerReady validates that the server is ready using Progressive Readiness
// FIXED: Uses Progressive Readiness pattern - no waiting, immediate validation
func AssertServerReady(t *testing.T, server *websocket.WebSocketServer) {
	// REQ-TEST-002: Performance optimization for test execution
	// Progressive Readiness Pattern: Server should be ready immediately after Start()

	if !server.IsRunning() {
		require.Fail(t, "WebSocket server not running - Progressive Readiness violation in test setup")
	}

	t.Log("✅ WebSocket server ready for immediate connections (Progressive Readiness)")
}

// CleanupTestClient closes a test client connection
func CleanupTestClient(t *testing.T, conn *gorilla.Conn) {
	if conn != nil {
		_ = conn.Close()
	}
}

// FIXED: Removed CleanupSharedTestServer - no more shared state
// Each test environment manages its own cleanup through UniversalTestSetup
