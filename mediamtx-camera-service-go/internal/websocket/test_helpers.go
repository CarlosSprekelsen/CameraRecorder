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
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// setupTestLogging configures logging for all tests
func setupTestLogging() {
	// Apply the test logging configuration directly without loading config
	// This prevents the config manager from creating loggers with default settings
	logging.SetupLogging(&logging.LoggingConfig{
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
// For now, we return the global logger to ensure it respects the configuration
func NewTestLogger(name string) *logging.Logger {
	// Use the global logger which respects the SetupLogging configuration
	// This is a workaround for the logging system design issue
	return logging.GetLogger()
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

// NewTestWebSocketServer creates a test WebSocket server using the PRODUCTION constructor
// with proper test dependencies. This ensures tests use the same code paths as production.
func NewTestWebSocketServer(t *testing.T) *WebSocketServer {
	// Load test configuration from fixtures
	configPath := getTestConfigPath(t)
	configManager := config.CreateConfigManager()
	err := configManager.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load test configuration from fixtures")

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
	cameraMonitor := createTestCameraMonitor(t, configManager, logger)
	mediaMTXController := createTestMediaMTXController(t, logger)

	// Use the PRODUCTION constructor with proper dependency injection
	server, err := NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor, // Real camera monitor
		jwtHandler,
		mediaMTXController, // Real MediaMTX controller
	)
	require.NoError(t, err, "Failed to create WebSocket server with production constructor")

	// Override the config with our test-specific port
	server.config = serverConfig

	return server
}

// StartTestServerWithDependencies starts the WebSocket server following the same pattern as main()
// This ensures proper startup sequence: camera monitor first, then WebSocket server
func StartTestServerWithDependencies(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Follow main() startup pattern: start camera monitor first, then WebSocket server
	// This is the responsibility of the test helper to set up unit tests properly

	// Start camera monitor first (following main() pattern)
	ctx := context.Background()
	cameraMonitor := server.GetCameraMonitor()
	if cameraMonitor != nil {
		err := cameraMonitor.Start(ctx)
		require.NoError(t, err, "Failed to start camera monitor")
	}

	// Start WebSocket server (following main() pattern)
	err := server.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
}

// createTestCameraMonitor creates a real camera monitor for testing
// NOTE: Monitor is created but NOT started - tests should start it explicitly following main() pattern
func createTestCameraMonitor(t *testing.T, configManager *config.ConfigManager, logger *logging.Logger) camera.CameraMonitor {
	// Create real camera monitor using the same pattern as camera tests
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	monitor, err := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(t, err, "Failed to create test camera monitor")

	return monitor
}

// createTestMediaMTXController creates a real MediaMTX controller for testing
func createTestMediaMTXController(t *testing.T, logger *logging.Logger) mediamtx.MediaMTXController {
	// Create MediaMTX test helper
	helper := mediamtx.NewMediaMTXTestHelper(t, nil)

	// Create config manager with test fixture
	configManager := mediamtx.CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")

	// Create controller using the same pattern as MediaMTX tests
	controller, err := mediamtx.ControllerWithConfigManager(configManager, helper.GetLogger())
	require.NoError(t, err, "Failed to create test MediaMTX controller")

	return controller
}

// NewTestWebSocketServerWithDependencies creates a test server with provided dependencies
func NewTestWebSocketServerWithDependencies(
	t *testing.T,
	cameraMonitor camera.CameraMonitor,
	mediaMTXController mediamtx.MediaMTXController,
) *WebSocketServer {
	server := NewTestWebSocketServer(t)
	server.cameraMonitor = cameraMonitor
	server.mediaMTXController = mediaMTXController
	return server
}

// NewTestClient creates a test WebSocket client connection
func NewTestClient(t *testing.T, server *WebSocketServer) *websocket.Conn {
	// Start server if not running
	if !server.IsRunning() {
		err := server.Start()
		require.NoError(t, err, "Failed to start test server")

		// Wait for server to be ready
		time.Sleep(100 * time.Millisecond)
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
		Error: &JsonRpcError{
			Code:    code,
			Message: message,
		},
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
