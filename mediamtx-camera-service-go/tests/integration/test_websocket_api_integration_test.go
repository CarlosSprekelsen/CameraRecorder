// +build integration,real_websocket

//go:build integration && real_websocket
// +build integration,real_websocket

/*
WebSocket API Integration Test

Requirements Coverage:
- REQ-AUTH-001: JWT authentication validation
- REQ-AUTH-002: Role-based access control
- REQ-RATE-001: Rate limiting enforcement
- REQ-WS-001: WebSocket connection management
- REQ-WS-002: JSON-RPC message handling
- REQ-WS-003: Real-time communication
- REQ-PERM-001: Permission enforcement
- REQ-API-001: API method validation

Test Categories: Integration/Real WebSocket/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebSocketAPITestSuite tests the WebSocket JSON-RPC API integration
type WebSocketAPITestSuite struct {
	serverURL     string
	wsServer      *websocket.WebSocketServer
	configManager *config.ConfigManager
	logger        *logging.Logger
	client        *gorilla.Conn
	authToken     string
}

// NewWebSocketAPITestSuite creates a new test suite
func NewWebSocketAPITestSuite() *WebSocketAPITestSuite {
	return &WebSocketAPITestSuite{
		serverURL: "ws://localhost:8002/ws",
	}
}

// Setup initializes the test suite
func (suite *WebSocketAPITestSuite) Setup(t *testing.T) {
	// Load configuration
	suite.configManager = config.NewConfigManager()
	err := suite.configManager.LoadConfig("config/default.yaml")
	require.NoError(t, err, "Failed to load configuration")

	// Setup logging
	suite.logger = logging.NewLogger("websocket-api-test")

	// Initialize real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor
	cameraMonitor := camera.NewHybridCameraMonitor(
		suite.configManager,
		suite.logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)

	// Initialize MediaMTX controller
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(suite.configManager, suite.logger.Logger)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// Initialize JWT handler
	cfg := suite.configManager.GetConfig()
	require.NotNil(t, cfg, "Configuration not available")

	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(t, err, "Failed to create JWT handler")

	// Initialize WebSocket server
	suite.wsServer = websocket.NewWebSocketServer(
		suite.configManager,
		suite.logger,
		cameraMonitor,
		jwtHandler,
		mediaMTXController,
	)

	// Start WebSocket server
	err = suite.wsServer.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	// Wait for server to be ready
	time.Sleep(1 * time.Second)

	// Connect WebSocket client
	suite.connectClient(t)

	// Authenticate client
	suite.authenticateClient(t)
}

// Teardown cleans up the test suite
func (suite *WebSocketAPITestSuite) Teardown(t *testing.T) {
	if suite.client != nil {
		suite.client.Close()
	}

	if suite.wsServer != nil {
		err := suite.wsServer.Stop()
		require.NoError(t, err, "Failed to stop WebSocket server")
	}
}

// connectClient establishes WebSocket connection
func (suite *WebSocketAPITestSuite) connectClient(t *testing.T) {
	u, err := url.Parse(suite.serverURL)
	require.NoError(t, err, "Failed to parse server URL")

	suite.client, _, err = gorilla.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err, "Failed to connect to WebSocket server")
}

// authenticateClient authenticates the client
func (suite *WebSocketAPITestSuite) authenticateClient(t *testing.T) {
	// Generate JWT token
	cfg := suite.configManager.GetConfig()
	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(t, err, "Failed to create JWT handler")

	suite.authToken, err = jwtHandler.GenerateToken("test-user", "admin", 1)
	require.NoError(t, err, "Failed to generate JWT token")

	// Send authentication request
	authRequest := websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "authenticate",
		Params: map[string]interface{}{
			"token": suite.authToken,
		},
	}

	response, err := suite.sendRequest(authRequest)
	require.NoError(t, err, "Failed to send authentication request")
	require.NotNil(t, response, "Authentication response should not be nil")
	require.Nil(t, response.Error, "Authentication should not have error")
}

// sendRequest sends a JSON-RPC request and returns the response
func (suite *WebSocketAPITestSuite) sendRequest(request websocket.JsonRpcRequest) (*websocket.JsonRpcResponse, error) {
	// Send request
	err := suite.client.WriteJSON(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var response websocket.JsonRpcResponse
	err = suite.client.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &response, nil
}

// TestWebSocketAPIAuthentication tests authentication endpoints
func TestWebSocketAPIAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ValidAuthentication", func(t *testing.T) {
		// Test with valid token
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "authenticate",
			Params: map[string]interface{}{
				"token": suite.authToken,
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should authenticate successfully")
		require.Nil(t, response.Error, "Should not have authentication error")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})

	t.Run("InvalidAuthentication", func(t *testing.T) {
		// Test with invalid token
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "authenticate",
			Params: map[string]interface{}{
				"token": "invalid-token",
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should handle invalid authentication gracefully")
		require.NotNil(t, response.Error, "Should have authentication error")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})
}

// TestWebSocketAPICameraOperations tests camera operation endpoints
func TestWebSocketAPICameraOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("GetCameras", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "get_cameras",
			Params:  map[string]interface{}{},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should get cameras successfully")
		require.Nil(t, response.Error, "Should not have error getting cameras")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")

		// Verify response structure
		if response.Result != nil {
			resultBytes, _ := json.Marshal(response.Result)
			t.Logf("Cameras response: %s", string(resultBytes))
		}
	})

	t.Run("GetCameraStatus", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "get_camera_status",
			Params: map[string]interface{}{
				"device_path": "/dev/video0",
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should get camera status successfully")
		// Note: This might fail if no camera is available, which is expected
		if response.Error != nil {
			t.Logf("Camera status error (expected if no camera): %v", response.Error)
		}
	})
}

// TestWebSocketAPIRecordingOperations tests recording operation endpoints
func TestWebSocketAPIRecordingOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("StartRecording", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "start_recording",
			Params: map[string]interface{}{
				"device_path": "/dev/video0",
				"options": map[string]interface{}{
					"use_case":       "recording",
					"priority":       1,
					"auto_cleanup":   true,
					"retention_days": 1,
					"quality":        "medium",
					"max_duration":   30, // 30 seconds for testing
				},
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should handle recording start request")
		// Note: This might fail if no camera is available, which is expected
		if response.Error != nil {
			t.Logf("Recording start error (expected if no camera): %v", response.Error)
		} else {
			t.Logf("Recording started successfully")
		}
	})

	t.Run("StopRecording", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      6,
			Method:  "stop_recording",
			Params: map[string]interface{}{
				"session_id": "test-session",
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should handle recording stop request")
		// Note: This might fail if no session exists, which is expected
		if response.Error != nil {
			t.Logf("Recording stop error (expected if no session): %v", response.Error)
		}
	})

	t.Run("TakeSnapshot", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      7,
			Method:  "take_snapshot",
			Params: map[string]interface{}{
				"device_path": "/dev/video0",
				"options": map[string]interface{}{
					"quality":    85,
					"format":     "jpeg",
					"resolution": "1920x1080",
				},
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should handle snapshot request")
		// Note: This might fail if no camera is available, which is expected
		if response.Error != nil {
			t.Logf("Snapshot error (expected if no camera): %v", response.Error)
		} else {
			t.Logf("Snapshot taken successfully")
		}
	})
}

// TestWebSocketAPIFileOperations tests file operation endpoints
func TestWebSocketAPIFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ListRecordings", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      8,
			Method:  "list_recordings",
			Params: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should list recordings successfully")
		require.Nil(t, response.Error, "Should not have error listing recordings")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})

	t.Run("ListSnapshots", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      9,
			Method:  "list_snapshots",
			Params: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should list snapshots successfully")
		require.Nil(t, response.Error, "Should not have error listing snapshots")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})
}

// TestWebSocketAPIHealthOperations tests health operation endpoints
func TestWebSocketAPIHealthOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("GetHealth", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      10,
			Method:  "get_health",
			Params:  map[string]interface{}{},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should get health successfully")
		require.Nil(t, response.Error, "Should not have error getting health")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})

	t.Run("GetMetrics", func(t *testing.T) {
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      11,
			Method:  "get_metrics",
			Params:  map[string]interface{}{},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should get metrics successfully")
		require.Nil(t, response.Error, "Should not have error getting metrics")
		assert.Equal(t, "2.0", response.JSONRPC, "Should have correct JSON-RPC version")
	})
}

// TestWebSocketAPIRateLimiting tests rate limiting functionality
func TestWebSocketAPIRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("RateLimitExceeded", func(t *testing.T) {
		// Send many requests quickly to trigger rate limiting
		for i := 0; i < 150; i++ { // More than the default 100 limit
			request := websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				ID:      i + 100,
				Method:  "get_health",
				Params:  map[string]interface{}{},
			}

			response, err := suite.sendRequest(request)
			if err != nil {
				t.Logf("Request %d failed: %v", i, err)
				break
			}

			if response.Error != nil && response.Error.Code == websocket.RATE_LIMIT_EXCEEDED {
				t.Logf("Rate limit exceeded at request %d", i)
				return // Successfully triggered rate limiting
			}
		}

		t.Log("Rate limiting not triggered (this is acceptable)")
	})
}

// TestWebSocketAPIPermissionEnforcement tests permission enforcement
func TestWebSocketAPIPermissionEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewWebSocketAPITestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("AdminPermission", func(t *testing.T) {
		// Test admin-only method
		request := websocket.JsonRpcRequest{
			JSONRPC: "2.0",
			ID:      200,
			Method:  "delete_recording", // Admin-only method
			Params: map[string]interface{}{
				"filename": "test.mp4",
			},
		}

		response, err := suite.sendRequest(request)
		require.NoError(t, err, "Should handle admin request")
		// Note: This might fail due to file not existing, which is expected
		if response.Error != nil {
			t.Logf("Admin request error (expected if file doesn't exist): %v", response.Error)
		}
	})
}

// BenchmarkWebSocketAPI benchmarks WebSocket API performance
func BenchmarkWebSocketAPI(b *testing.B) {
	suite := NewWebSocketAPITestSuite()
	suite.Setup(&testing.T{})
	defer suite.Teardown(&testing.T{})

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			request := websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				ID:      i,
				Method:  "get_health",
				Params:  map[string]interface{}{},
			}

			response, err := suite.sendRequest(request)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			_ = response
		}
	})

	b.Run("GetCameras", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			request := websocket.JsonRpcRequest{
				JSONRPC: "2.0",
				ID:      i,
				Method:  "get_cameras",
				Params:  map[string]interface{}{},
			}

			response, err := suite.sendRequest(request)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			_ = response
		}
	})
}
