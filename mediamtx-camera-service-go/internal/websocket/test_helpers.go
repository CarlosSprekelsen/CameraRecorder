/*
WebSocket Test Helpers - Progressive Readiness Pattern Support

Provides focused test utilities for WebSocket module testing,
following the Progressive Readiness Pattern from the architecture.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-ARCH-001: Progressive Readiness Pattern compliance

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
Architecture Reference: docs/architecture/go-architecture-guide.md
*/

package websocket

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// Global mutex to prevent parallel test execution
// WebSocket tests must run sequentially because they share the same server resources
var testMutex sync.Mutex

// WebSocketTestConfig provides configuration for WebSocket server testing
type WebSocketTestConfig struct {
	Host                 string
	Port                 int
	WebSocketPath        string
	MaxConnections       int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	PingInterval         time.Duration
	PongWait             time.Duration
	MaxMessageSize       int64
	ReadBufferSize       int
	WriteBufferSize      int
	ShutdownTimeout      time.Duration
	ClientCleanupTimeout time.Duration
	TestDataDir          string
	CleanupAfter         bool
}

// DefaultWebSocketTestConfig returns default configuration for WebSocket server testing
func DefaultWebSocketTestConfig() *WebSocketTestConfig {
	return &WebSocketTestConfig{
		Host:                 "localhost",
		Port:                 0, // Will be assigned dynamically
		WebSocketPath:        "/ws",
		MaxConnections:       100,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		PingInterval:         30 * time.Second,
		PongWait:             60 * time.Second,
		MaxMessageSize:       1024 * 1024, // 1MB
		ReadBufferSize:       4096,
		WriteBufferSize:      4096,
		ShutdownTimeout:      30 * time.Second,
		ClientCleanupTimeout: 5 * time.Second,
		TestDataDir:          "/tmp/websocket_test_data",
		CleanupAfter:         true,
	}
}

// WebSocketTestHelper provides utilities for WebSocket server testing
type WebSocketTestHelper struct {
	config             *WebSocketTestConfig
	configIntegration  *mediamtx.ConfigIntegration
	logger             *logging.Logger
	server             *WebSocketServer
	mediaMTXController mediamtx.MediaMTXController
	jwtHandler         *security.JWTHandler

	// Race-free initialization using sync.Once
	serverOnce     sync.Once
	controllerOnce sync.Once
	jwtHandlerOnce sync.Once
}

// EnsureSequentialExecution ensures tests run sequentially to avoid WebSocket server conflicts
// Call this at the beginning of each test that uses WebSocket server
func EnsureSequentialExecution(t *testing.T) {
	testMutex.Lock()
	t.Cleanup(func() {
		testMutex.Unlock()
	})
}

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

// NewWebSocketTestHelper creates a new test helper for WebSocket server testing
func NewWebSocketTestHelper(t *testing.T, config *WebSocketTestConfig) *WebSocketTestHelper {
	if config == nil {
		config = DefaultWebSocketTestConfig()
	}

	// Create logger for testing
	logger := logging.GetLogger("test-websocket-server")

	// Create test data directory
	err := os.MkdirAll(config.TestDataDir, 0755)
	require.NoError(t, err, "Failed to create test data directory")

	// Create config manager using existing fixtures
	configManager := createTestConfigManager(t)

	// Get free port automatically
	if config.Port == 0 {
		config.Port = GetFreePort()
	}

	// Create config integration
	configIntegration := mediamtx.NewConfigIntegration(configManager, logger)

	return &WebSocketTestHelper{
		config:            config,
		configIntegration: configIntegration,
		logger:            logger,
	}
}

// Cleanup cleans up the test helper resources
func (h *WebSocketTestHelper) Cleanup(t *testing.T) {
	if h.server != nil && h.server.IsRunning() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := h.server.Stop(ctx)
		if err != nil {
			t.Logf("Warning: Failed to stop WebSocket server: %v", err)
		}
	}

	if h.config.CleanupAfter && h.config.TestDataDir != "" {
		err := os.RemoveAll(h.config.TestDataDir)
		if err != nil {
			t.Logf("Warning: Failed to clean up test data directory: %v", err)
		}
	}
}

// GetServer returns the WebSocket server instance (lazy initialization)
func (h *WebSocketTestHelper) GetServer(t *testing.T) *WebSocketServer {
	h.serverOnce.Do(func() {
		// Create server configuration
		cfg, err := h.configIntegration.GetConfig()
		if err != nil {
			t.Fatalf("Failed to get configuration: %v", err)
		}
		serverConfig := &ServerConfig{
			Host:                 h.config.Host,
			Port:                 h.config.Port,
			WebSocketPath:        h.config.WebSocketPath,
			MaxConnections:       h.config.MaxConnections,
			ReadTimeout:          h.config.ReadTimeout,
			WriteTimeout:         h.config.WriteTimeout,
			PingInterval:         h.config.PingInterval,
			PongWait:             h.config.PongWait,
			MaxMessageSize:       h.config.MaxMessageSize,
			ReadBufferSize:       h.config.ReadBufferSize,
			WriteBufferSize:      h.config.WriteBufferSize,
			ShutdownTimeout:      h.config.ShutdownTimeout,
			ClientCleanupTimeout: h.config.ClientCleanupTimeout,
		}

		// Create JWT handler
		jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey, h.logger)
		require.NoError(t, err, "Failed to create JWT handler")

		// Create MediaMTX controller
		mediaMTXController := h.GetMediaMTXController(t)

		// Create WebSocket server using production constructor
		server, err := NewWebSocketServer(
			h.configIntegration,
			h.logger,
			jwtHandler,
			mediaMTXController,
		)
		require.NoError(t, err, "Failed to create WebSocket server")

		// Override config with test-specific settings
		server.config = serverConfig
		h.server = server
	})

	return h.server
}

// GetMediaMTXController returns the MediaMTX controller instance (lazy initialization)
func (h *WebSocketTestHelper) GetMediaMTXController(t *testing.T) mediamtx.MediaMTXController {
	h.controllerOnce.Do(func() {
		// Use MediaMTX test helper to create controller
		helper := mediamtx.NewMediaMTXTestHelper(t, nil)
		controller, err := helper.GetController(t)
		require.NoError(t, err, "Failed to create MediaMTX controller")

		// Start the controller (following MediaMTX test pattern)
		ctx := context.Background()
		if concreteController, ok := controller.(interface{ Start(context.Context) error }); ok {
			err := concreteController.Start(ctx)
			require.NoError(t, err, "Failed to start MediaMTX controller")
		}

		h.mediaMTXController = controller
	})

	return h.mediaMTXController
}

// StartServer starts the WebSocket server following Progressive Readiness Pattern
// FIXED: No longer waits for camera monitor readiness - follows architecture pattern
func (h *WebSocketTestHelper) StartServer(t *testing.T) *WebSocketServer {
	server := h.GetServer(t)

	// Start WebSocket server immediately - Progressive Readiness Pattern
	// System accepts connections immediately, features become available as components initialize
	err := server.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	return server
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

	// Create config integration and server configuration
	configIntegration := mediamtx.NewConfigIntegration(configManager, logger)
	cfg, err := configIntegration.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get configuration: %v", err)
	}
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
		configIntegration,
		logger,
		jwtHandler,
		mediaMTXController, // Real MediaMTX controller
	)
	require.NoError(t, err, "Failed to create WebSocket server with production constructor")

	// Override the config with our test-specific port
	server.config = serverConfig

	return server
}

// StartTestServerWithDependencies starts the WebSocket server following Progressive Readiness Pattern
// FIXED: No longer waits for camera monitor readiness - follows architecture pattern
func StartTestServerWithDependencies(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Start WebSocket server immediately - Progressive Readiness Pattern
	// System accepts connections immediately, features become available as components initialize
	err := server.Start()
	require.NoError(t, err, "Failed to start WebSocket server")
}

// REMOVED: waitForCameraMonitorReady - This function violated the Progressive Readiness Pattern
// The architecture states: "System accepts connections immediately"
// "Features become available as components initialize"
// "No blocking startup dependencies"

// ValidateProgressiveReadiness validates that the system follows the Progressive Readiness Pattern
func (h *WebSocketTestHelper) ValidateProgressiveReadiness(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Test 1: System should accept connections immediately
	conn := h.NewTestClient(t, server)
	defer h.CleanupTestClient(t, conn)

	// Test 2: System should return readiness status instead of blocking
	message := CreateTestMessage("get_system_status", nil)
	response := SendTestMessage(t, conn, message)

	require.NotNil(t, response, "System should respond to status requests immediately")
	require.Nil(t, response.Error, "System should not error on status requests")
	require.NotNil(t, response.Result, "System should return readiness status")

	// Test 3: System should handle requests even when not fully ready
	// This validates the Progressive Readiness Pattern behavior
	pingMessage := CreateTestMessage("ping", nil)
	pingResponse := SendTestMessage(t, conn, pingMessage)

	// System should either respond with pong or readiness status
	require.NotNil(t, pingResponse, "System should respond to ping requests")
	if pingResponse.Error == nil {
		// If no error, should get pong response
		require.Equal(t, "pong", pingResponse.Result, "Should get pong response when system is ready")
	} else {
		// If error, should be authentication error, not system not ready error
		require.NotEqual(t, -32002, pingResponse.Error.Code, "Should not get system not ready error")
	}

	t.Log("Progressive Readiness Pattern validation passed")
}

// TestProgressiveReadinessBehavior tests the Progressive Readiness Pattern behavior
func (h *WebSocketTestHelper) TestProgressiveReadinessBehavior(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Test immediate connection acceptance
	conn := h.NewTestClient(t, server)
	defer h.CleanupTestClient(t, conn)

	// Test that system returns proper readiness status
	statusMessage := CreateTestMessage("get_system_status", nil)
	statusResponse := SendTestMessage(t, conn, statusMessage)

	require.NotNil(t, statusResponse, "System should respond to status requests immediately")
	require.Nil(t, statusResponse.Error, "Status request should not error")

	// Validate response structure
	statusResult, ok := statusResponse.Result.(map[string]interface{})
	require.True(t, ok, "Status response should be a map")
	require.Contains(t, statusResult, "status", "Status response should contain status field")
	require.Contains(t, statusResult, "message", "Status response should contain message field")
	require.Contains(t, statusResult, "available_cameras", "Status response should contain available_cameras field")
	require.Contains(t, statusResult, "discovery_active", "Status response should contain discovery_active field")

	t.Log("Progressive Readiness behavior test passed")
}

// NewTestClient creates a test WebSocket client connection
func (h *WebSocketTestHelper) NewTestClient(t *testing.T, server *WebSocketServer) *websocket.Conn {
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

// CleanupTestClient closes a test client connection
func (h *WebSocketTestHelper) CleanupTestClient(t *testing.T, conn *websocket.Conn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			t.Logf("Warning: Failed to close test client: %v", err)
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Stop(ctx)
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
