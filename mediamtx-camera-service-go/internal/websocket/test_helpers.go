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
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// GetFreePort returns a free port for testing
func GetFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 8002 // fallback
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 8002 // fallback
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

// NewTestWebSocketServer creates a test WebSocket server with minimal dependencies
func NewTestWebSocketServer(t *testing.T) *WebSocketServer {
	// Create test configuration
	port := GetFreePort()
	config := &ServerConfig{
		Port:         port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PongWait:     60 * time.Second,
	}

	// Create minimal test dependencies
	logger := logging.NewLogger("websocket-test")

	// Create test JWT handler
	jwtHandler, err := security.NewJWTHandler("test-secret-key")
	require.NoError(t, err, "Failed to create test JWT handler")

	// Create test permission checker
	permissionChecker := security.NewPermissionChecker()

	// Create test validation helper
	inputValidator := security.NewInputValidator(logger, nil)
	validationHelper := NewValidationHelper(inputValidator, logrus.New())

	// Create test event manager
	eventManager := NewEventManager(logrus.New())

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
		metrics:           &PerformanceMetrics{},
		eventHandlers:     make([]func(string, interface{}), 0),
		stopChan:          make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in tests
			},
		},
	}

	// Register default test methods
	server.registerDefaultMethods()

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
