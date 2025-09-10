/*
WebSocket Test Utilities

Provides shared test utilities for WebSocket module testing,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-TEST-001: Test environment setup and management
- REQ-TEST-002: Performance optimization for test execution

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package testutils

import (
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// Global test server instance for reuse across tests
var (
	sharedTestServer *websocket.WebSocketServer
	serverMutex      sync.RWMutex
	serverPort       int
)

// TestEnvironment represents the shared test environment
type TestEnvironment struct {
	Server     *websocket.WebSocketServer
	Config     *config.Config
	TempDir    string
	Logger     *logging.Logger
	ConfigPath string
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
	time.Sleep(10 * time.Millisecond)

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

// SetupTestEnvironment creates a shared test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// REQ-TEST-001: Test environment setup and management

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "websocket_test_*")
	require.NoError(t, err, "Failed to create temporary test directory")

	// Load test configuration
	configPath := getTestConfigPath(t)
	configManager := config.CreateConfigManager()
	err = configManager.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load test configuration from fixtures")

	// Create logger
	logger := logging.GetLogger("websocket-test-utils")

	// Get shared test server
	server := GetSharedTestServer(t, configManager.GetConfig(), tempDir)

	return &TestEnvironment{
		Server:     server,
		Config:     configManager.GetConfig(),
		TempDir:    tempDir,
		Logger:     logger,
		ConfigPath: configPath,
	}
}

// TeardownTestEnvironment cleans up the test environment
func TeardownTestEnvironment(t *testing.T, env *TestEnvironment) {
	// REQ-TEST-001: Test environment cleanup

	if env != nil {
		// Clean up temporary directory
		if env.TempDir != "" {
			err := os.RemoveAll(env.TempDir)
			if err != nil {
				t.Logf("Warning: Failed to clean up temp directory: %v", err)
			}
		}
	}
}

// GetSharedTestServer returns a shared test server instance
func GetSharedTestServer(t *testing.T, cfg *config.Config, tempDir string) *websocket.WebSocketServer {
	// REQ-TEST-002: Performance optimization for test execution

	serverMutex.Lock()
	defer serverMutex.Unlock()

	// Return existing server if available and running
	if sharedTestServer != nil && sharedTestServer.IsRunning() {
		return sharedTestServer
	}

	// Use the existing test helper to create server
	server := websocket.NewTestWebSocketServer(t)

	// Start the server
	err := server.Start()
	require.NoError(t, err, "Failed to start shared test server")

	// Wait for server to be ready
	WaitForServerReady(t, server, 5*time.Second)

	sharedTestServer = server
	return server
}

// NewTestClient creates a test WebSocket client connection
func NewTestClient(t *testing.T, server *websocket.WebSocketServer) *gorilla.Conn {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint

	// Use the existing test helper
	return websocket.NewTestClient(t, server)
}

// CreateTestMessage creates a test JSON-RPC message
func CreateTestMessage(method string, params map[string]interface{}) *websocket.JsonRpcRequest {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	return websocket.CreateTestMessage(method, params)
}

// CreateTestNotification creates a test JSON-RPC notification
func CreateTestNotification(method string, params map[string]interface{}) *websocket.JsonRpcNotification {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	return websocket.CreateTestNotification(method, params)
}

// SendTestMessage sends a test message and waits for response
func SendTestMessage(t *testing.T, conn *gorilla.Conn, message *websocket.JsonRpcRequest) *websocket.JsonRpcResponse {
	// REQ-API-003: Request/response message handling

	// Use the existing test helper
	return websocket.SendTestMessage(t, conn, message)
}

// SendTestNotification sends a test notification (no response expected)
func SendTestNotification(t *testing.T, conn *gorilla.Conn, notification *websocket.JsonRpcNotification) {
	// REQ-API-003: Request/response message handling

	// Use the existing test helper
	websocket.SendTestNotification(t, conn, notification)
}

// WaitForServerReady waits for the server to be ready
func WaitForServerReady(t *testing.T, server *websocket.WebSocketServer, timeout time.Duration) {
	// REQ-TEST-002: Performance optimization for test execution

	// Use the existing test helper
	websocket.WaitForServerReady(t, server, timeout)
}

// CleanupTestClient closes a test client connection
func CleanupTestClient(t *testing.T, conn *gorilla.Conn) {
	// Use the existing test helper
	websocket.CleanupTestClient(t, conn)
}

// CleanupSharedTestServer stops the shared test server
func CleanupSharedTestServer(t *testing.T) {
	// REQ-TEST-001: Test environment cleanup

	serverMutex.Lock()
	defer serverMutex.Unlock()

	if sharedTestServer != nil {
		err := sharedTestServer.Stop()
		if err != nil {
			t.Logf("Warning: Failed to stop shared test server: %v", err)
		}
		sharedTestServer = nil
	}
}
