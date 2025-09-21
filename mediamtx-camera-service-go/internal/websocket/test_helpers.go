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
	"github.com/stretchr/testify/assert"
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
	t             *testing.T // Test instance for consistent error handling
	config        *WebSocketTestConfig
	configManager *config.ConfigManager
	logger        *logging.Logger
	server        *WebSocketServer
	// listener removed - using improved port allocation instead
	mediaMTXController mediamtx.MediaMTXController
	mediaMTXHelper     *mediamtx.MediaMTXTestHelper // Store helper for cleanup
	jwtHandler         *security.JWTHandler

	// Race-free initialization using sync.Once
	serverOnce     sync.Once
	listenerOnce   sync.Once
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

	// Create logger using standardized enterprise pattern
	logger := NewTestLogger("enterprise-websocket-test")

	// Create test data directory
	err := os.MkdirAll(config.TestDataDir, 0755)
	require.NoError(t, err, "Failed to create test data directory")

	// Create config manager using standardized enterprise pattern
	configManager := createStandardTestConfig(t)

	// Get free port automatically (will be replaced with listener-based approach)
	if config.Port == 0 {
		config.Port = 0 // Use 0 to indicate dynamic port allocation needed
	}

	return &WebSocketTestHelper{
		t:             t,
		config:        config,
		configManager: configManager,
		logger:        logger,
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

	// CRITICAL: Cleanup MediaMTX controller to prevent fsnotify file descriptor leaks
	if h.mediaMTXHelper != nil {
		h.mediaMTXHelper.Cleanup(t)
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
		cfg := h.configManager.GetConfig()
		if cfg == nil {
			t.Fatalf("No configuration loaded")
		}
		// Get bound port to eliminate race conditions
		_ = h.GetFreePortReliably() // Ensures port is allocated and stored in h.config.Port

		serverConfig := &ServerConfig{
			Host:                 h.config.Host,
			Port:                 h.config.Port, // Port set by GetListener()
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
		// Create JWT handler using standardized enterprise pattern
		jwtHandler, err := h.createStandardJWTHandler()
		require.NoError(t, err, "Failed to create standardized JWT handler")

		// Create MediaMTX controller
		mediaMTXController := h.GetMediaMTXController(t)

		// Create WebSocket server using production constructor
		server, err := NewWebSocketServer(
			h.configManager,
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
		h.mediaMTXHelper = helper // Store helper for cleanup
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
// Old GetFreePort function completely removed - use helper.GetFreePortReliably() instead

// GetFreePortReliably returns a free port using a more reliable method
func (h *WebSocketTestHelper) GetFreePortReliably() int {
	h.listenerOnce.Do(func() {
		// Bind temporarily to get port, then close immediately before server starts
		tempListener, err := net.Listen("tcp", ":0")
		require.NoError(h.t, err, "Failed to bind temporary listener")

		h.config.Port = tempListener.Addr().(*net.TCPAddr).Port
		tempListener.Close() // Close immediately to release for server

		// Small delay to ensure port is fully released
		time.Sleep(1 * time.Millisecond)
	})
	return h.config.Port
}

// Enterprise Test Infrastructure Standards
const ENTERPRISE_TEST_JWT_SECRET = "enterprise-websocket-test-jwt-secret-key"

// createStandardJWTHandler creates JWT handler using standardized enterprise pattern
func (h *WebSocketTestHelper) createStandardJWTHandler() (*security.JWTHandler, error) {
	return security.NewJWTHandler(ENTERPRISE_TEST_JWT_SECRET, h.logger)
}

// DEPRECATED: createStandardTestConfig - Use consolidated config helper approach
// TODO: Migrate all WebSocket tests to use config.NewTestConfigHelper(t).CreateTestDirectories()
// This function will be removed once all tests are migrated to the unified approach
func createStandardTestConfig(t *testing.T) *config.ConfigManager {
	// Use consolidated config helper for directory creation (same as MediaMTX tests)
	configHelper := config.NewTestConfigHelper(t)
	configHelper.CreateTestDirectories() // Creates /tmp/recordings, /tmp/snapshots with 0777

	// Use canonical config fixture (same as MediaMTX tests)
	return mediamtx.CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")
}

// REMOVED: NewTestWebSocketServer - use helper.GetServer() for standardized pattern

// REMOVED: StartTestServerWithDependencies - use helper.StartServer() for standardized pattern

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
	// FIXED: Use get_status (API-compliant) with admin role (API requirement)
	h.authenticateConnection(t, conn, "test_user", "admin")

	message := CreateTestMessage("get_status", nil)
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
		// Progressive Readiness Pattern: Server accepts connections immediately after Start()
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
// REMOVED: NewTestWebSocketServerWithDependencies - use helper.GetServer() for standardized pattern

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
	// Progressive Readiness Pattern: Server accepts connections immediately after Start()
	// No waiting required - server is ready when Start() returns successfully

	if server.IsRunning() {
		t.Log("WebSocket server is ready for connections")
		return
	}

	// If not running, this indicates a bug in the test setup
	require.Fail(t, "WebSocket server not running - check test setup, server.Start() should have been called")
}

// CleanupTestServer stops and cleans up a test server
func DELETED_CleanupTestServer_UNUSED(t *testing.T, server *WebSocketServer) {
	if server != nil && server.IsRunning() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Stop(ctx)
		if err != nil {
			t.Logf("Warning: Failed to stop test server: %v", err)
		}
	}
}

// ValidateProgressiveReadinessCompliance performs enterprise-grade Progressive Readiness validation
func (h *WebSocketTestHelper) ValidateProgressiveReadinessCompliance(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Enterprise Test 1: Connection acceptance timing validation
	connectionTimes := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		startTime := time.Now()
		conn := h.NewTestClient(t, server)
		connectionTimes[i] = time.Since(startTime)
		h.CleanupTestClient(t, conn)

		assert.Less(t, connectionTimes[i], 100*time.Millisecond,
			"Connection %d took %v - should be <100ms (Progressive Readiness)", i, connectionTimes[i])
	}

	// Enterprise Test 2: Concurrent connection acceptance
	const concurrentConnections = 20
	var wg sync.WaitGroup
	results := make([]bool, concurrentConnections)

	wg.Add(concurrentConnections)
	for i := 0; i < concurrentConnections; i++ {
		go func(index int) {
			defer wg.Done()
			conn := h.NewTestClient(t, server)
			results[index] = (conn != nil)
			if conn != nil {
				h.CleanupTestClient(t, conn)
			}
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(t, concurrentConnections, successCount,
		"All concurrent connections should succeed (Progressive Readiness)")
}

// TestEnterpriseGradeOperations tests enterprise-grade operation patterns
func (h *WebSocketTestHelper) TestEnterpriseGradeOperations(t *testing.T, server *WebSocketServer) {
	t.Helper()

	conn := h.NewTestClient(t, server)
	defer h.CleanupTestClient(t, conn)

	// Enterprise Test 1: System status should always be available
	statusMessage := CreateTestMessage("get_system_status", nil)
	statusResponse := SendTestMessage(t, conn, statusMessage)

	require.NotNil(t, statusResponse, "System status should always respond")

	// Enterprise Test 2: Component operations should gracefully handle initialization
	operationTests := []struct {
		method string
		params map[string]interface{}
		name   string
	}{
		{"get_camera_list", nil, "Camera List"},
		{"take_snapshot", map[string]interface{}{"device": "camera0"}, "Snapshot"},
		{"start_recording", map[string]interface{}{"device": "camera0", "filename": "test.mp4"}, "Recording"},
	}

	for _, test := range operationTests {
		message := CreateTestMessage(test.method, test.params)
		response := SendTestMessage(t, conn, message)

		require.NotNil(t, response, "%s operation should always respond", test.name)

		if response.Error != nil {
			// Should get meaningful error, not "system not ready"
			assert.NotEqual(t, -32002, response.Error.Code,
				"%s should not get generic 'not ready' error", test.name)
		}
	}
}

// TestArchitecturalCompliance_ProgressiveReadiness validates architectural compliance
func TestArchitecturalCompliance_ProgressiveReadiness(t *testing.T, server *WebSocketServer) {
	t.Helper()

	// Architectural Test 1: Server starts accepting operations immediately
	startTime := time.Now()
	helper := &WebSocketTestHelper{} // Create helper instance for consistent pattern
	conn := helper.NewTestClient(t, server)
	defer helper.CleanupTestClient(t, conn)

	startDuration := time.Since(startTime)
	assert.Less(t, startDuration, 100*time.Millisecond,
		"Server connection should be immediate (Progressive Readiness)")

	// Architectural Test 2: Operations are accepted immediately (may use fallback)
	operationStart := time.Now()
	statusMessage := CreateTestMessage("get_system_status", nil)
	response := SendTestMessage(t, conn, statusMessage)
	operationDuration := time.Since(operationStart)

	assert.Less(t, operationDuration, 200*time.Millisecond,
		"Operations should respond quickly via fallback if needed")

	require.NotNil(t, response, "System status should always respond")

	// Architectural Test 3: No blocking "system not ready" errors
	if response.Error != nil {
		assert.NotEqual(t, -32002, response.Error.Code,
			"Should not get 'system not ready' blocking error")
	}

	// Architectural Test 4: Graceful degradation for component-dependent operations
	cameraMessage := CreateTestMessage("get_camera_list", nil)
	cameraResponse := SendTestMessage(t, conn, cameraMessage)

	require.NotNil(t, cameraResponse, "Camera list should always respond")

	// May return empty list or error, but should not block
	if cameraResponse.Error != nil {
		assert.NotEqual(t, -32002, cameraResponse.Error.Code,
			"Camera operations should gracefully degrade, not block")
	}
}

// GetAuthenticatedConnection returns a ready-to-use authenticated WebSocket connection
// This eliminates the 12-line setup pattern across all tests
func (h *WebSocketTestHelper) GetAuthenticatedConnection(t *testing.T, userID string, role string) *websocket.Conn {
	// Standard setup pattern - all in one method
	controller := h.GetMediaMTXController(t)
	server := h.GetServer(t)
	server.SetMediaMTXController(controller)
	server = h.StartServer(t)
	conn := h.NewTestClient(t, server)

	// Authenticate the connection
	h.authenticateConnection(t, conn, userID, role)

	return conn
}

// TestMethod provides the most minimal pattern for simple method testing
// Reduces test setup to just 2 lines for basic method tests
func (h *WebSocketTestHelper) TestMethod(t *testing.T, method string, params map[string]interface{}, userRole string) *JsonRpcResponse {
	conn := h.GetAuthenticatedConnection(t, "test_user", userRole)
	defer h.CleanupTestClient(t, conn)

	message := CreateTestMessage(method, params)
	return SendTestMessage(t, conn, message)
}

// AuthenticateTestClient authenticates a WebSocket connection (standalone function for compatibility)
func AuthenticateTestClient(t *testing.T, conn *websocket.Conn, userID string, role string) {
	// Use standardized enterprise JWT handler creation
	jwtHandler, err := security.NewJWTHandler(ENTERPRISE_TEST_JWT_SECRET, NewTestLogger("test-jwt"))
	require.NoError(t, err, "Failed to create standardized JWT handler")
	// Generate test token directly since security.GenerateTestToken requires build tags
	testToken, err := jwtHandler.GenerateToken(userID, role, 24)
	require.NoError(t, err, "Failed to generate test token")

	// Send authentication message
	authMessage := CreateTestMessage("authenticate", map[string]interface{}{
		"token": testToken,
	})
	authResponse := SendTestMessage(t, conn, authMessage)
	require.Nil(t, authResponse.Error, "Authentication should succeed")
	require.Equal(t, "authenticated", authResponse.Result, "Authentication should return success")
}

// authenticateConnection authenticates a connection (internal helper)
func (h *WebSocketTestHelper) authenticateConnection(t *testing.T, conn *websocket.Conn, userID string, role string) {
	// Use standardized enterprise JWT handler creation
	jwtHandler, err := h.createStandardJWTHandler()
	require.NoError(t, err, "Failed to create standardized JWT handler")
	// Generate test token directly since security.GenerateTestToken requires build tags
	testToken, err := jwtHandler.GenerateToken(userID, role, 24)
	require.NoError(t, err, "Failed to generate test token")

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
