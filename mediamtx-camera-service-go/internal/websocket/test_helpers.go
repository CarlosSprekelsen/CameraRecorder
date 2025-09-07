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
	"fmt"
	"net"
	"net/http"
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

// NewTestWebSocketServer creates a test WebSocket server using fixtures
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
	config := &ServerConfig{
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

	// Create minimal test dependencies
	logger := logging.NewLogger("websocket-test")

	// Create test JWT handler
	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey, logger)
	require.NoError(t, err, "Failed to create test JWT handler")

	// Create test permission checker
	permissionChecker := security.NewPermissionChecker()

	// Create test validation helper
	inputValidator := security.NewInputValidator(logger, nil)
	validationHelper := NewValidationHelper(inputValidator, logging.NewLogger("test-validation-helper"))

	// Create test event manager
	eventManager := NewEventManager(logging.NewLogger("test-event-manager"))

	// Create server with test dependencies
	server := &WebSocketServer{
		config:            config,
		logger:            logger,
		jwtHandler:        jwtHandler,
		permissionChecker: permissionChecker,
		validationHelper:  validationHelper,
		eventManager:      eventManager,
		clients:           make(map[string]*ClientConnection),
		methods:           make(map[string]MethodHandler),
		methodVersions:    make(map[string]string),
		metrics: &PerformanceMetrics{
			RequestCount:      0,
			ResponseTimes:     make(map[string][]float64),
			ErrorCount:        0,
			ActiveConnections: 0,
			StartTime:         time.Now(),
		},
		eventHandlers: make([]func(string, interface{}), 0),
		stopChan:      make(chan struct{}, 10), // Buffered to prevent deadlock during shutdown
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in tests
			},
		},
	}

	// Register built-in methods (same as production server)
	server.registerBuiltinMethods()

	return server
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
	if !server.running {
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
		if server.running {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Fail(t, "Server failed to become ready within timeout")
}

// CleanupTestServer stops and cleans up a test server
func CleanupTestServer(t *testing.T, server *WebSocketServer) {
	if server != nil && server.running {
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
