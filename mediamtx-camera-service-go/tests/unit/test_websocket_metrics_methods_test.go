/*
WebSocket Metrics Methods Test

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Core method implementations (get_metrics)

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

// Use type aliases to avoid import conflicts
type HealthStatus = mediamtx.HealthStatus
type Metrics = mediamtx.Metrics
type SystemMetrics = mediamtx.SystemMetrics
type Stream = mediamtx.Stream
type Path = mediamtx.Path
type RecordingSession = mediamtx.RecordingSession
type Snapshot = mediamtx.Snapshot
type FileListResponse = mediamtx.FileListResponse
type FileMetadata = mediamtx.FileMetadata
type SnapshotSettings = mediamtx.SnapshotSettings
type MediaMTXConfig = mediamtx.MediaMTXConfig

// TestGetMetricsMethod tests the get_metrics JSON-RPC method
func TestGetMetricsMethod(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")
	cameraMonitor := &camera.HybridCameraMonitor{}
	jwtHandler := security.NewJWTHandler(configManager, logger)

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

	// Test successful metrics retrieval
	t.Run("successful_metrics_retrieval", func(t *testing.T) {
		params := map[string]interface{}{}

		response, err := server.MethodGetMetrics(params, client)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)

		// Verify required fields from API documentation
		assert.Contains(t, result, "active_connections")
		assert.Contains(t, result, "total_requests")
		assert.Contains(t, result, "average_response_time")
		assert.Contains(t, result, "error_rate")
		assert.Contains(t, result, "memory_usage")
		assert.Contains(t, result, "cpu_usage")
		assert.Contains(t, result, "goroutines")
		assert.Contains(t, result, "heap_alloc")

		// Verify data types
		assert.IsType(t, float64(0), result["active_connections"])
		assert.IsType(t, float64(0), result["total_requests"])
		assert.IsType(t, float64(0), result["average_response_time"])
		assert.IsType(t, float64(0), result["error_rate"])
		assert.IsType(t, float64(0), result["memory_usage"])
		assert.IsType(t, float64(0), result["cpu_usage"])
		assert.IsType(t, float64(0), result["goroutines"])
		assert.IsType(t, float64(0), result["heap_alloc"])
	})

	// Test authentication required
	t.Run("authentication_required", func(t *testing.T) {
		unauthenticatedClient := &websocket.ClientConnection{
			ClientID:      "test-client",
			Authenticated: false,
		}

		params := map[string]interface{}{}

		response, err := server.MethodGetMetrics(params, unauthenticatedClient)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, websocket.AUTHENTICATION_REQUIRED, response.Error.Code)
		assert.Equal(t, "Authentication required", response.Error.Message)
	})
}
