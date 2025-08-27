/*
Epic E3 JSON-RPC Methods Implementation Tests

Tests for the newly implemented JSON-RPC methods following Python patterns:
- get_metrics
- get_camera_capabilities  
- get_status

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Core method implementations

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEpicE3MethodGetMetrics tests the get_metrics method implementation
func TestEpicE3MethodGetMetrics(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}
	
	// Test get_metrics method
	response, err := server.MethodGetMetrics(map[string]interface{}{}, client)
	
	// Validate response
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.Error)
	
	// Validate result structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	
	// Check required fields from Python implementation
	assert.Contains(t, result, "active_connections")
	assert.Contains(t, result, "total_requests")
	assert.Contains(t, result, "average_response_time")
	assert.Contains(t, result, "error_rate")
	assert.Contains(t, result, "memory_usage")
	assert.Contains(t, result, "cpu_usage")
	
	// Validate data types
	assert.IsType(t, float64(0), result["active_connections"])
	assert.IsType(t, int64(0), result["total_requests"])
	assert.IsType(t, float64(0), result["average_response_time"])
	assert.IsType(t, float64(0), result["error_rate"])
	assert.IsType(t, float64(0), result["memory_usage"])
	assert.IsType(t, float64(0), result["cpu_usage"])
}

// TestEpicE3MethodGetCameraCapabilities tests the get_camera_capabilities method implementation
func TestEpicE3MethodGetCameraCapabilities(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}
	
	// Test get_camera_capabilities method with valid device
	response, err := server.MethodGetCameraCapabilities(map[string]interface{}{
		"device": "/dev/video0",
	}, client)
	
	// Validate response
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.Error)
	
	// Validate result structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	
	// Check required fields from Python implementation
	assert.Contains(t, result, "device")
	assert.Contains(t, result, "formats")
	assert.Contains(t, result, "resolutions")
	assert.Contains(t, result, "fps_options")
	assert.Contains(t, result, "validation_status")
	
	// Validate data types
	assert.IsType(t, "", result["device"])
	assert.IsType(t, []interface{}{}, result["formats"])
	assert.IsType(t, []interface{}{}, result["resolutions"])
	assert.IsType(t, []interface{}{}, result["fps_options"])
	assert.IsType(t, "", result["validation_status"])
	
	// Test error case - missing device parameter
	response, err = server.MethodGetCameraCapabilities(map[string]interface{}{}, client)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.NotNil(t, response.Error)
	assert.Equal(t, websocket.INVALID_PARAMS, response.Error.Code)
}

// TestEpicE3MethodGetStatus tests the get_status method implementation
func TestEpicE3MethodGetStatus(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}
	
	// Test get_status method
	response, err := server.MethodGetStatus(map[string]interface{}{}, client)
	
	// Validate response
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.Error)
	
	// Validate result structure
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	
	// Check required fields from Python implementation
	assert.Contains(t, result, "status")
	assert.Contains(t, result, "uptime")
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "components")
	
	// Validate data types
	assert.IsType(t, "", result["status"])
	assert.IsType(t, float64(0), result["uptime"])
	assert.IsType(t, "", result["version"])
	
	// Validate components structure
	components, ok := result["components"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, components, "websocket_server")
	assert.Contains(t, components, "camera_monitor")
	assert.Contains(t, components, "mediamtx_controller")
	
	// Validate component statuses
	assert.IsType(t, "", components["websocket_server"])
	assert.IsType(t, "", components["camera_monitor"])
	assert.IsType(t, "", components["mediamtx_controller"])
}

// TestEpicE3MethodAuthentication tests authentication requirements for Epic E3 methods
func TestEpicE3MethodAuthentication(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create unauthenticated test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: false,
		ConnectedAt:   time.Now(),
	}
	
	// Test all Epic E3 methods with unauthenticated client
	methods := []struct {
		name   string
		params map[string]interface{}
		handler func(map[string]interface{}, *websocket.ClientConnection) (*websocket.JsonRpcResponse, error)
	}{
		{"get_metrics", map[string]interface{}{}, server.MethodGetMetrics},
		{"get_camera_capabilities", map[string]interface{}{"device": "/dev/video0"}, server.MethodGetCameraCapabilities},
		{"get_status", map[string]interface{}{}, server.MethodGetStatus},
	}
	
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			response, err := method.handler(method.params, client)
			
			require.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
			assert.NotNil(t, response.Error)
			assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		})
	}
}

// TestEpicE3MethodRegistration tests that Epic E3 methods are properly registered
func TestEpicE3MethodRegistration(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Test that Epic E3 methods are registered
	expectedMethods := []string{
		"get_metrics",
		"get_camera_capabilities", 
		"get_status",
	}
	
	for _, methodName := range expectedMethods {
		t.Run(methodName, func(t *testing.T) {
			// This would require exposing the methods map for testing
			// For now, we'll test that the methods can be called without panic
			client := &websocket.ClientConnection{
				ClientID:      "test-client",
				Authenticated: true,
				UserID:        "test-user",
				Role:          "admin",
				ConnectedAt:   time.Now(),
			}
			
			var response *websocket.JsonRpcResponse
			var err error
			
			switch methodName {
			case "get_metrics":
				response, err = server.MethodGetMetrics(map[string]interface{}{}, client)
			case "get_camera_capabilities":
				response, err = server.MethodGetCameraCapabilities(map[string]interface{}{"device": "/dev/video0"}, client)
			case "get_status":
				response, err = server.MethodGetStatus(map[string]interface{}{}, client)
			}
			
			require.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
		})
	}
}

// TestEpicE3MethodPerformance tests that Epic E3 methods meet performance requirements
func TestEpicE3MethodPerformance(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}
	
	// Test performance for each Epic E3 method
	methods := []struct {
		name   string
		params map[string]interface{}
		handler func(map[string]interface{}, *websocket.ClientConnection) (*websocket.JsonRpcResponse, error)
	}{
		{"get_metrics", map[string]interface{}{}, server.MethodGetMetrics},
		{"get_camera_capabilities", map[string]interface{}{"device": "/dev/video0"}, server.MethodGetCameraCapabilities},
		{"get_status", map[string]interface{}{}, server.MethodGetStatus},
	}
	
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			start := time.Now()
			
			response, err := method.handler(method.params, client)
			
			duration := time.Since(start)
			
			require.NoError(t, err)
			assert.NotNil(t, response)
			
			// Performance requirement: <50ms response time
			assert.Less(t, duration, 50*time.Millisecond, 
				"Method %s took %v, expected <50ms", method.name, duration)
		})
	}
}

// TestEpicE3MethodJSONCompatibility tests JSON serialization compatibility
func TestEpicE3MethodJSONCompatibility(t *testing.T) {
	// Setup test infrastructure using existing components
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger)
	jwtHandler := security.NewJWTHandler(configManager, logger)
	
	server := websocket.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	
	// Create test client
	client := &websocket.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}
	
	// Test JSON serialization for each Epic E3 method
	methods := []struct {
		name   string
		params map[string]interface{}
		handler func(map[string]interface{}, *websocket.ClientConnection) (*websocket.JsonRpcResponse, error)
	}{
		{"get_metrics", map[string]interface{}{}, server.MethodGetMetrics},
		{"get_camera_capabilities", map[string]interface{}{"device": "/dev/video0"}, server.MethodGetCameraCapabilities},
		{"get_status", map[string]interface{}{}, server.MethodGetStatus},
	}
	
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			response, err := method.handler(method.params, client)
			
			require.NoError(t, err)
			assert.NotNil(t, response)
			
			// Test JSON serialization
			jsonData, err := json.Marshal(response)
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)
			
			// Test JSON deserialization
			var decodedResponse websocket.JsonRpcResponse
			err = json.Unmarshal(jsonData, &decodedResponse)
			require.NoError(t, err)
			
			// Validate decoded response
			assert.Equal(t, response.JSONRPC, decodedResponse.JSONRPC)
			assert.Equal(t, response.ID, decodedResponse.ID)
			assert.Equal(t, response.Result, decodedResponse.Result)
		})
	}
}
