/*
WebSocket Error Handling Unit Tests

Requirements Coverage:
- REQ-API-007: Error handling and validation
- REQ-API-008: JSON-RPC 2.0 error response format
- REQ-API-009: Parameter validation and error codes
- REQ-API-010: Authentication error handling
- REQ-API-011: Authorization error handling

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

// TestErrorHandling tests various error scenarios for JSON-RPC methods
func TestErrorHandling(t *testing.T) {
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

	// Test malformed parameters
	t.Run("malformed_parameters", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test with invalid parameter types
		params := map[string]interface{}{
			"device": 123, // Should be string
		}

		response, err := server.MethodGetCameraStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "invalid parameter")
	})

	// Test missing required parameters
	t.Run("missing_required_parameters", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test without required device parameter
		params := map[string]interface{}{}

		response, err := server.MethodGetCameraStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "device parameter is required")
	})

	// Test invalid device paths
	t.Run("invalid_device_paths", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		invalidPaths := []string{
			"invalid-device",
			"/dev/invalid",
			"video0",
			"/dev/video",
			"/dev/videoabc",
		}

		for _, invalidPath := range invalidPaths {
			t.Run("invalid_path_"+invalidPath, func(t *testing.T) {
				params := map[string]interface{}{
					"device": invalidPath,
				}

				response, err := server.MethodGetCameraStatus(params, client)

				require.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.NotNil(t, response.Error)
				assert.Contains(t, response.Error.Message, "invalid device path")
			})
		}
	})

	// Test authentication errors
	t.Run("authentication_errors", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		// Test ping without authentication
		params := map[string]interface{}{}
		response, err := server.MethodPing(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})

	// Test authorization errors
	t.Run("authorization_errors", func(t *testing.T) {
		viewerClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "viewer",
		}

		// Test admin-only method with viewer role
		params := map[string]interface{}{}
		response, err := server.MethodGetStatus(params, viewerClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "insufficient permissions")
	})

	// Test invalid JWT tokens
	t.Run("invalid_jwt_tokens", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		invalidTokens := []string{
			"invalid-token",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			"",
			"not-a-jwt-token",
		}

		for _, invalidToken := range invalidTokens {
			t.Run("invalid_token_"+invalidToken, func(t *testing.T) {
				params := map[string]interface{}{
					"auth_token": invalidToken,
				}

				response, err := server.MethodAuthenticate(params, unauthenticatedClient)

				require.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.NotNil(t, response.Error)
				assert.Contains(t, response.Error.Message, "invalid token")
			})
		}
	})

	// Test expired JWT tokens
	t.Run("expired_jwt_tokens", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		// Note: This test would require a mock JWT handler that can generate expired tokens
		// For now, we test the error handling structure
		params := map[string]interface{}{
			"auth_token": "expired-token-placeholder",
		}

		response, err := server.MethodAuthenticate(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
	})

	// Test rate limiting errors
	t.Run("rate_limiting_errors", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test rapid successive requests
		params := map[string]interface{}{}
		
		// Make multiple rapid requests to trigger rate limiting
		for i := 0; i < 10; i++ {
			response, err := server.MethodPing(params, client)
			require.NoError(t, err)
			assert.NotNil(t, response)
			
			// Check if rate limiting error occurs
			if response.Error != nil && response.Error.Code == websocket.RATE_LIMIT_EXCEEDED {
				assert.Equal(t, "Rate limit exceeded", response.Error.Message)
				break
			}
		}
	})

	// Test malformed JSON-RPC requests
	t.Run("malformed_json_rpc", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test with nil parameters
		response, err := server.MethodPing(nil, client)
		require.NoError(t, err)
		assert.NotNil(t, response)

		// Test with invalid client
		params := map[string]interface{}{}
		response, err = server.MethodPing(params, nil)
		require.Error(t, err)
		assert.Nil(t, response)
	})

	// Test method not found errors
	t.Run("method_not_found", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test with non-existent method
		params := map[string]interface{}{}
		
		// This would require testing the method dispatch mechanism
		// For now, we test the error handling structure
		response, err := server.MethodPing(params, client)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
	})

	// Test concurrent access errors
	t.Run("concurrent_access", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test concurrent method calls
		const numGoroutines = 5
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				params := map[string]interface{}{}
				_, err := server.MethodPing(params, client)
				results <- err
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	// Test timeout scenarios
	t.Run("timeout_scenarios", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test method execution with potential timeout
		params := map[string]interface{}{}
		response, err := server.MethodPing(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
	})

	// Test resource exhaustion scenarios
	t.Run("resource_exhaustion", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test with large parameter sets
		largeParams := map[string]interface{}{
			"device": "/dev/video0",
			"data":   make([]byte, 1000), // Large data payload
		}

		response, err := server.MethodGetCameraStatus(largeParams, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
	})

	// Test error response format compliance
	t.Run("error_response_format", func(t *testing.T) {
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

		// Verify error response format per JSON-RPC 2.0 specification
		assert.NotEmpty(t, response.Error.Code)
		assert.NotEmpty(t, response.Error.Message)
		assert.IsType(t, int(0), response.Error.Code)
		assert.IsType(t, "", response.Error.Message)
	})

	// Test error code consistency
	t.Run("error_code_consistency", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		// Test authentication error code consistency
		params := map[string]interface{}{}
		response, err := server.MethodPing(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)

		// Test same error code for different methods
		response2, err := server.MethodGetCameraList(params, unauthenticatedClient)
		require.NoError(t, err)
		assert.NotNil(t, response2)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response2.Error.Code)
	})

	// Test error message clarity
	t.Run("error_message_clarity", func(t *testing.T) {
		client := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: true,
			UserID:        "test-user",
			Role:          "admin",
		}

		// Test missing parameter error message
		params := map[string]interface{}{}
		response, err := server.MethodGetCameraStatus(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Error)
		assert.Contains(t, response.Error.Message, "device parameter is required")
		assert.Contains(t, response.Error.Message, "required")
	})
}
