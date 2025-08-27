/*
Core JSON-RPC Methods Unit Tests

Requirements Coverage:
- REQ-API-001: JSON-RPC 2.0 protocol compliance
- REQ-API-002: Authentication and authorization
- REQ-API-003: Camera discovery and status reporting
- REQ-API-004: System health monitoring
- REQ-API-005: Server information and capabilities
- REQ-API-006: Stream management and listing
- REQ-API-007: Error handling and validation

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)



// TestPingMethod tests the ping JSON-RPC method
func TestPingMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful ping
	t.Run("successful_ping", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodPing(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.Equal(t, "pong", response.Result)
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodPing(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})
}

// TestAuthenticateMethod tests the authenticate JSON-RPC method
func TestAuthenticateMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	// Test valid JWT authentication
	t.Run("valid_jwt_authentication", func(t *testing.T) {
		// Generate valid JWT token
		token, err := jwtHandler.GenerateToken("test-user", "operator", 24)
		require.NoError(t, err)

		params := map[string]interface{}{
			"auth_token": token,
		}

		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		response, err := server.MethodAuthenticate(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, true, result["authenticated"])
		assert.Equal(t, "operator", result["role"])
		assert.NotEmpty(t, result["session_id"])
	})

	// Test invalid token
	t.Run("invalid_token", func(t *testing.T) {
		params := map[string]interface{}{
			"auth_token": "invalid-token",
		}

		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		response, err := server.MethodAuthenticate(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "invalid token")
	})

	// Test missing token
	t.Run("missing_token", func(t *testing.T) {
		params := map[string]interface{}{}

		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		response, err := server.MethodAuthenticate(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "auth_token parameter is required")
	})
}

// TestGetCameraListMethod tests the get_camera_list JSON-RPC method
func TestGetCameraListMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful camera list retrieval
	t.Run("successful_camera_list_retrieval", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetCameraList(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Contains(t, result, "cameras")
		assert.Contains(t, result, "total")
		assert.Contains(t, result, "connected")

		// Verify data types
		assert.IsType(t, []interface{}{}, result["cameras"])
		assert.IsType(t, float64(0), result["total"])
		assert.IsType(t, float64(0), result["connected"])
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetCameraList(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})
}

// TestGetCameraStatusMethod tests the get_camera_status JSON-RPC method
func TestGetCameraStatusMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful camera status retrieval
	t.Run("successful_camera_status_retrieval", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "/dev/video0",
		}

		response, err := server.MethodGetCameraStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Equal(t, "/dev/video0", result["device"])
		assert.Contains(t, result, "status")
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "resolution")
		assert.Contains(t, result, "fps")
		assert.Contains(t, result, "streams")
	})

	// Test missing device parameter
	t.Run("missing_device_parameter", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetCameraStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "device parameter is required")
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{
			"device": "/dev/video0",
		}

		response, err := server.MethodGetCameraStatus(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})
}

// TestGetStatusMethod tests the get_status JSON-RPC method
func TestGetStatusMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful status retrieval
	t.Run("successful_status_retrieval", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Contains(t, result, "status")
		assert.Contains(t, result, "uptime")
		assert.Contains(t, result, "version")
		assert.Contains(t, result, "components")

		// Verify data types
		assert.IsType(t, "", result["status"])
		assert.IsType(t, float64(0), result["uptime"])
		assert.IsType(t, "", result["version"])
		assert.IsType(t, map[string]interface{}{}, result["components"])
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetStatus(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})

	// Test insufficient role
	t.Run("insufficient_role", func(t *testing.T) {
		viewerClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "viewer",
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetStatus(params, viewerClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "insufficient permissions")
	})
}

// TestGetServerInfoMethod tests the get_server_info JSON-RPC method
func TestGetServerInfoMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful server info retrieval
	t.Run("successful_server_info_retrieval", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetServerInfo(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "version")
		assert.Contains(t, result, "build_date")
		assert.Contains(t, result, "go_version")
		assert.Contains(t, result, "architecture")
		assert.Contains(t, result, "capabilities")
		assert.Contains(t, result, "supported_formats")
		assert.Contains(t, result, "max_cameras")

		// Verify data types
		assert.IsType(t, "", result["name"])
		assert.IsType(t, "", result["version"])
		assert.IsType(t, "", result["build_date"])
		assert.IsType(t, "", result["go_version"])
		assert.IsType(t, "", result["architecture"])
		assert.IsType(t, []interface{}{}, result["capabilities"])
		assert.IsType(t, []interface{}{}, result["supported_formats"])
		assert.IsType(t, float64(0), result["max_cameras"])
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetServerInfo(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})

	// Test insufficient role
	t.Run("insufficient_role", func(t *testing.T) {
		viewerClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "viewer",
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetServerInfo(params, viewerClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "insufficient permissions")
	})
}

// TestGetStreamsMethod tests the get_streams JSON-RPC method
func TestGetStreamsMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful streams retrieval
	t.Run("successful_streams_retrieval", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetStreams(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Contains(t, result, "streams")
		assert.Contains(t, result, "total")
		assert.Contains(t, result, "active")

		// Verify data types
		assert.IsType(t, []interface{}{}, result["streams"])
		assert.IsType(t, float64(0), result["total"])
		assert.IsType(t, float64(0), result["active"])
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetStreams(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})

	// Test insufficient role
	t.Run("insufficient_role", func(t *testing.T) {
		viewerClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "viewer",
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetStreams(params, viewerClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "insufficient permissions")
	})
}

// TestGetCameraCapabilitiesMethod tests the get_camera_capabilities JSON-RPC method
func TestGetCameraCapabilitiesMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-unit-testing-only")
	require.NoError(t, err)

	server := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		nil, // No controller for basic test
	)

	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
	}

	// Test successful capabilities retrieval
	t.Run("successful_capabilities_retrieval", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "/dev/video0",
		}

		response, err := server.MethodGetCameraCapabilities(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Equal(t, "/dev/video0", result["device"])
		assert.Contains(t, result, "formats")
		assert.Contains(t, result, "resolutions")
		assert.Contains(t, result, "fps_options")
		assert.Contains(t, result, "validation_status")

		// Verify data types
		assert.IsType(t, []interface{}{}, result["formats"])
		assert.IsType(t, []interface{}{}, result["resolutions"])
		assert.IsType(t, []interface{}{}, result["fps_options"])
		assert.IsType(t, "", result["validation_status"])
	})

	// Test missing device parameter
	t.Run("missing_device_parameter", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetCameraCapabilities(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "device parameter is required")
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{
			"device": "/dev/video0",
		}

		response, err := server.MethodGetCameraCapabilities(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})
}
