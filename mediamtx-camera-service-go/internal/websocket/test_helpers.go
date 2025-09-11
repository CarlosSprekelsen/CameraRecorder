/*
WebSocket Test Helpers

Provides focused test utilities for WebSocket module testing,
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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// setupTestLogging configures logging for all tests
func setupTestLogging() {
	// Configure the global logger factory for tests
	// This ensures all loggers created through the factory use test configuration
	logging.ConfigureGlobalLogging(&logging.LoggingConfig{
		Level:          "error",
		Format:         "json",
		FileEnabled:    false,
		ConsoleEnabled: false,
	})
}

// getTestConfigPathForSetup gets the test config path for TestMain setup
func getTestConfigPathForSetup() string {
	// Start from current directory and walk up to find project root
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

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
			break
		}
		dir = parent
	}

	return ""
}

// NewTestLogger creates a logger for tests that uses the global logging configuration
// This function should only be called after setupTestLogging() has been called
// Uses the logger factory to ensure consistent configuration across all test loggers
func NewTestLogger(name string) *logging.Logger {
	// Use the factory to get a logger with consistent configuration
	return logging.GetLogger(name)
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

	// Port should be released immediately

	return port
}

// createTestConfigManager creates a test configuration manager using existing fixtures
// following the MediaMTX test helper pattern of using fixtures
func createTestConfigManager(t *testing.T) *config.ConfigManager {
	// Create test data directory and required files before loading fixture
	testDataDir := "/tmp/websocket_test_data"
	err := os.MkdirAll(testDataDir, 0755)
	require.NoError(t, err, "Failed to create test data directory")

	// Create required directories and files for configuration validation
	recordingsDir := filepath.Join(testDataDir, "recordings")
	snapshotsDir := filepath.Join(testDataDir, "snapshots")
	mediamtxConfigFile := filepath.Join(testDataDir, "mediamtx.yml")

	err = os.MkdirAll(recordingsDir, 0755)
	require.NoError(t, err, "Failed to create recordings directory")

	err = os.MkdirAll(snapshotsDir, 0755)
	require.NoError(t, err, "Failed to create snapshots directory")

	// Create minimal MediaMTX config file
	err = os.WriteFile(mediamtxConfigFile, []byte("# Test MediaMTX configuration\n"), 0644)
	require.NoError(t, err, "Failed to create MediaMTX config file")

	// Use existing fixture following MediaMTX pattern
	return mediamtx.CreateConfigManagerWithFixture(t, "config_websocket_test.yaml")
}

// NewTestWebSocketServer creates a test WebSocket server using the PRODUCTION constructor
// with proper test dependencies. This ensures tests use the same code paths as production.
func NewTestWebSocketServer(t *testing.T) *WebSocketServer {
	// Create self-contained test configuration (following MediaMTX test helper pattern)
	configManager := createTestConfigManager(t)

	// Get free port automatically (port 0 = OS assigns next available)
	port := GetFreePort()

	// Create server configuration from test config
	cfg := configManager.GetConfig()
	serverConfig := &ServerConfig{
		Host:                 cfg.Server.Host,
		Port:                 port, // Use dynamically assigned port
		WebSocketPath:        cfg.Server.WebSocketPath,
		MaxConnections:       cfg.Server.MaxConnections,
		ReadTimeout:          cfg.Server.ReadTimeout,
		WriteTimeout:         cfg.Server.WriteTimeout,
		PingInterval:         cfg.Server.PingInterval,
		PongWait:             cfg.Server.PongWait,
		MaxMessageSize:       cfg.Server.MaxMessageSize,
		ReadBufferSize:       cfg.Server.ReadBufferSize,
		WriteBufferSize:      cfg.Server.WriteBufferSize,
		ShutdownTimeout:      cfg.Server.ShutdownTimeout,
		ClientCleanupTimeout: cfg.Server.ClientCleanupTimeout,
	}

	// Create logger (logging configuration is set up globally in TestMain)
	logger := NewTestLogger("websocket-test")

	// Create test JWT handler
	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey, logger)
	require.NoError(t, err, "Failed to create test JWT handler")

	// Create REAL test dependencies (not mocks)
	mediaMTXController := createTestMediaMTXController(t, configManager, logger)

	// Use the PRODUCTION constructor with proper dependency injection
	server, err := NewWebSocketServer(
		configManager,
		logger,
		jwtHandler,
		mediaMTXController, // Real MediaMTX controller
	)
	require.NoError(t, err, "Failed to create WebSocket server with production constructor")

	// Override the config with our test-specific port
	server.config = serverConfig

	return server
}

// StartTestServerWithDependencies starts the WebSocket server following the same pattern as main()
// FIXED: Use proper API instead of calling private Start method
func StartTestServerWithDependencies(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// FIXED: MediaMTX controller is already started by MediaMTXTestHelper
	// No need to call private Start method - this violates architecture
	// The MediaMTX controller should be ready to use via its public API

	// Wait for camera monitor to be ready (our new readiness check)
	// This ensures cameras are discovered before tests run
	waitForCameraMonitorReady(t, server.mediaMTXController)

	// Start WebSocket server (following main() pattern)
	err := server.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
}

// REMOVED: createTestCameraMonitor - WebSocket tests should not create camera monitors directly
// Camera monitor creation is MediaMTX controller's responsibility, not WebSocket layer's
// This violates the architecture: WebSocket → MediaMTX Controller → Camera Monitor

// waitForCameraMonitorReady waits for the camera monitor to complete initial discovery
func waitForCameraMonitorReady(t *testing.T, controller mediamtx.MediaMTXControllerAPI) {
	t.Helper()

	const maxWaitTime = 2 * time.Second         // Reasonable timeout
	const checkInterval = 50 * time.Millisecond // Check every 50ms

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Logf("Timeout waiting for camera monitor to be ready (waited %v)", maxWaitTime)
			return // Don't fail the test, just log and continue
		case <-ticker.C:
			// Check if the camera monitor is ready by trying to get camera list
			// If it returns cameras or completes without error, it's ready
			ctx := context.Background()
			cameraList, err := controller.GetCameraList(ctx)
			if err == nil {
				// Camera monitor is ready - it can successfully query cameras
				t.Logf("Camera monitor is ready, found %d cameras", cameraList.Total)
				return
			}
			// If error is "controller not running", keep waiting
			if !strings.Contains(err.Error(), "controller not running") {
				// Other errors might indicate readiness (e.g., no cameras found)
				t.Logf("Camera monitor appears ready (error: %v)", err)
				return
			}
		}
	}
}

// createTestMediaMTXController creates a real MediaMTX controller for testing
// FIXED: Use MediaMTX test helper and start the controller properly
func createTestMediaMTXController(t *testing.T, configManager *config.ConfigManager, logger *logging.Logger) mediamtx.MediaMTXController {
	// Use MediaMTX test helper to properly start MediaMTX server and controller
	// This follows the same pattern as MediaMTX unit tests
	helper := mediamtx.NewMediaMTXTestHelper(t, nil)

	// Get the controller from the helper (this properly starts MediaMTX server)
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Failed to create test MediaMTX controller via helper")

	// Start the controller (same as MediaMTX unit tests do)
	// Cast to concrete type to access Start method (not exposed in MediaMTXControllerAPI interface)
	ctx := context.Background()
	if concreteController, ok := controller.(interface{ Start(context.Context) error }); ok {
		err := concreteController.Start(ctx)
		require.NoError(t, err, "Failed to start MediaMTX controller")
	} else {
		t.Fatal("MediaMTX controller does not implement Start method")
	}

	return controller
}

// NewTestWebSocketServerWithDependencies creates a test server with provided dependencies
func NewTestWebSocketServerWithDependencies(
	t *testing.T,
	mediaMTXController mediamtx.MediaMTXController,
) *WebSocketServer {
	server := NewTestWebSocketServer(t)
	// WebSocket server only depends on MediaMTX Controller (thin protocol layer)
	server.mediaMTXController = mediaMTXController
	return server
}

// NewTestClient creates a test WebSocket client connection
func NewTestClient(t *testing.T, server *WebSocketServer) *websocket.Conn {
	// Start server if not running
	if !server.IsRunning() {
		err := server.Start()
		require.NoError(t, err, "Failed to start test server")

		// Wait for server to be ready with proper verification
		deadline := time.Now().Add(1 * time.Second)
		for time.Now().Before(deadline) {
			if server.IsRunning() {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Connect to server
	url := fmt.Sprintf("ws://localhost:%d/ws", server.config.Port)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "Failed to connect to test server")

	return conn
}

// CreateTestMessage creates a test JSON-RPC message
func CreateTestMessage(method string, params map[string]interface{}) *JsonRpcRequest {
	return &JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      fmt.Sprintf("test-%d", time.Now().UnixNano()),
		Params:  params,
	}
}

// CreateTestNotification creates a test JSON-RPC notification
func CreateTestNotification(method string, params map[string]interface{}) *JsonRpcNotification {
	return &JsonRpcNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// CreateTestResponse creates a test JSON-RPC response
func CreateTestResponse(id interface{}, result interface{}) *JsonRpcResponse {
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   nil,
	}
}

// CreateTestErrorResponse creates a test JSON-RPC error response
func CreateTestErrorResponse(id interface{}, code int, message string) *JsonRpcResponse {
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   NewJsonRpcError(code, "test_error", message, "Check test parameters"),
	}
}

// SendTestMessage sends a test message and waits for response
func SendTestMessage(t *testing.T, conn *websocket.Conn, message *JsonRpcRequest) *JsonRpcResponse {
	// Send message
	err := conn.WriteJSON(message)
	require.NoError(t, err, "Failed to send test message")

	// Read response
	var response JsonRpcResponse
	err = conn.ReadJSON(&response)
	require.NoError(t, err, "Failed to read test response")

	return &response
}

// SendTestNotification sends a test notification (no response expected)
func SendTestNotification(t *testing.T, conn *websocket.Conn, notification *JsonRpcNotification) {
	err := conn.WriteJSON(notification)
	require.NoError(t, err, "Failed to send test notification")
}

// WaitForServerReady waits for the server to be ready
func WaitForServerReady(t *testing.T, server *WebSocketServer, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if server.IsRunning() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Fail(t, "Server failed to become ready within timeout")
}

// CleanupTestServer stops and cleans up a test server
func CleanupTestServer(t *testing.T, server *WebSocketServer) {
	if server != nil && server.IsRunning() {
		err := server.Stop()
		if err != nil {
			t.Logf("Warning: Failed to stop test server: %v", err)
		}
	}
}

// CleanupTestClient closes a test client connection
func CleanupTestClient(t *testing.T, conn *websocket.Conn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			t.Logf("Warning: Failed to close test client: %v", err)
		}
	}
}

// AuthenticateTestClient authenticates a test client using the existing security helpers
// This eliminates duplication of JWT handler creation across tests
func AuthenticateTestClient(t *testing.T, conn *websocket.Conn, userID string, role string) {
	// Use the same secret key as the test configuration to ensure compatibility
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-websocket-tests-only", NewTestLogger("test-jwt"))
	require.NoError(t, err, "Failed to create JWT handler with correct secret")
	testToken := security.GenerateTestToken(t, jwtHandler, userID, role)

	// Authenticate the client
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"auth_token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")
}

// registerDefaultMethods registers default test methods on the server
func (s *WebSocketServer) registerDefaultMethods() {
	// Register ping method
	s.registerMethod("ping", func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		return CreateTestResponse("test-id", "pong"), nil
	}, "1.0")

	// Register echo method for testing
	s.registerMethod("echo", func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		return CreateTestResponse("test-id", params), nil
	}, "1.0")

	// Register error method for testing
	s.registerMethod("error", func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		return CreateTestErrorResponse("test-id", INTERNAL_ERROR, "Test error"), nil
	}, "1.0")
}
