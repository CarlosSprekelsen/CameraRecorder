/*
WebSocket server basic functionality tests.

Provides basic tests to verify WebSocket server instantiation and core functionality
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWebSocketServer tests WebSocket server instantiation
func TestNewWebSocketServer(t *testing.T) {
	// Create test dependencies
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create mock camera monitor
	cameraMonitor := &camera.HybridCameraMonitor{}

	// Create JWT handler with test secret
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	// Test successful instantiation
	server := NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)

	assert.NotNil(t, server)
	assert.NotNil(t, server.config)
	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.cameraMonitor)
	assert.NotNil(t, server.jwtHandler)
	assert.NotNil(t, server.metrics)
	assert.False(t, server.running)
}

// TestWebSocketServerMethods tests method registration
func TestWebSocketServerMethods(t *testing.T) {
	// Create test dependencies
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)

	// Test that core methods are registered
	assert.Contains(t, server.methods, "ping")
	assert.Contains(t, server.methods, "authenticate")
	assert.Contains(t, server.methods, "get_camera_list")
	assert.Contains(t, server.methods, "get_camera_status")
}

// TestJsonRpcRequestValidation tests JSON-RPC request validation
func TestJsonRpcRequestValidation(t *testing.T) {
	// Test valid request
	validRequest := &JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      1,
	}

	assert.Equal(t, "2.0", validRequest.JSONRPC)
	assert.Equal(t, "ping", validRequest.Method)
	assert.Equal(t, 1, validRequest.ID)

	// Test invalid JSON-RPC version
	invalidRequest := &JsonRpcRequest{
		JSONRPC: "1.0",
		Method:  "ping",
		ID:      1,
	}

	assert.Equal(t, "1.0", invalidRequest.JSONRPC)
}

// TestErrorMessages tests error message mapping
func TestErrorMessages(t *testing.T) {
	// Test that error messages are defined
	assert.NotEmpty(t, ErrorMessages[AUTHENTICATION_REQUIRED])
	assert.NotEmpty(t, ErrorMessages[METHOD_NOT_FOUND])
	assert.NotEmpty(t, ErrorMessages[INVALID_PARAMS])
	assert.NotEmpty(t, ErrorMessages[INTERNAL_ERROR])

	// Test specific error messages
	assert.Equal(t, "Authentication required", ErrorMessages[AUTHENTICATION_REQUIRED])
	assert.Equal(t, "Method not found", ErrorMessages[METHOD_NOT_FOUND])
	assert.Equal(t, "Invalid parameters", ErrorMessages[INVALID_PARAMS])
	assert.Equal(t, "Internal server error", ErrorMessages[INTERNAL_ERROR])
}

// TestDefaultServerConfig tests default server configuration
func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "0.0.0.0", config.Host)
	assert.Equal(t, 8002, config.Port)
	assert.Equal(t, "/ws", config.WebSocketPath)
	assert.Equal(t, 1000, config.MaxConnections)
	assert.Equal(t, int64(1024*1024), config.MaxMessageSize)
}
