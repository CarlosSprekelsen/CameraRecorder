/*
WebSocket Test Utilities - FIXED Resource Management

Provides shared test utilities for WebSocket module testing,
using the good patterns from internal/testutils and eliminating
resource management issues.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-TEST-001: Test environment setup and management
- REQ-TEST-002: Performance optimization for test execution

FIXED ISSUES:
- Eliminated duplicate mutex instances (leverages shared infrastructure)
- Proper resource cleanup using UniversalTestSetup pattern
- Progressive Readiness compliance
- No global shared state (each test gets isolated resources)

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package testutils

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// FIXED: Removed global shared state - each test gets isolated resources
// This eliminates mutex duplication and resource leaks

// WebSocketTestEnvironment represents an isolated test environment
// FIXED: Each test gets its own environment - no shared state
type WebSocketTestEnvironment struct {
	Server     *websocket.WebSocketServer
	Config     *config.Config
	TempDir    string
	Logger     *logging.Logger
	ConfigPath string
	Setup      *testutils.UniversalTestSetup // FIXED: Use good pattern from testutils
}

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

// getTestConfigPath finds the WebSocket test configuration file in fixtures
func getTestConfigPath(t *testing.T) string {
	// Start from current directory and walk up to find project root
	dir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Found project root, look for WebSocket test config
			configPath := filepath.Join(dir, "tests", "fixtures", "config_websocket_test.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	require.Fail(t, "Could not find WebSocket test configuration file in fixtures")
	return ""
}

// SetupWebSocketTestEnvironment creates an isolated test environment
// SetupWebSocketTestEnvironment creates a WebSocket test environment
// DEPRECATED: use testutils.SetupTest(t, "config_websocket_test.yaml") instead.
// This function will be removed in a future version.
// FIXED: Uses UniversalTestSetup pattern for proper resource management
func SetupWebSocketTestEnvironment(t *testing.T) *WebSocketTestEnvironment {
	// REQ-TEST-001: Test environment setup and management

	// DEPRECATED: Use single canonical fixture instead of config_websocket_test.yaml
	// TODO: Remove config_websocket_test.yaml and use config_valid_complete.yaml
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")
	configManager := setup.GetConfigManager()
	logger := setup.GetLogger()

	// FIXED: Use UniversalTestSetup for proper resource management
	// Tests should use shared infrastructure pattern directly

	return &WebSocketTestEnvironment{
		Server:     nil, // FIXED: Tests should use shared infrastructure directly
		Config:     configManager.GetConfig(),
		TempDir:    "", // FIXED: Use UniversalTestSetup for temp dir management
		Logger:     logger,
		ConfigPath: "",    // FIXED: Use UniversalTestSetup for config management
		Setup:      setup, // FIXED: Leverage good pattern from testutils
	}
}

// TeardownWebSocketTestEnvironment cleans up the test environment
// DEPRECATED: use testutils.SetupTest(t, "config_websocket_test.yaml") instead.
// This function will be removed in a future version.
// FIXED: Uses UniversalTestSetup cleanup pattern
func TeardownWebSocketTestEnvironment(t *testing.T, env *WebSocketTestEnvironment) {
	// REQ-TEST-001: Test environment cleanup

	if env != nil {
		// FIXED: Use UniversalTestSetup cleanup (proper resource management)
		if env.Setup != nil {
			env.Setup.Cleanup() // This handles all cleanup properly
		}

		// Clean up WebSocket server
		if env.Server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := env.Server.Stop(ctx); err != nil {
				t.Logf("Warning: Failed to stop WebSocket server: %v", err)
			}
		}
	}
}

// FIXED: Removed createIsolatedWebSocketServer - use shared infrastructure instead
// This eliminates duplicate server creation and leverages existing good patterns

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
