package constants

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAPIConstants verifies that API constants are properly defined
func TestAPIConstants(t *testing.T) {
	// Test JSON-RPC error codes
	assert.Equal(t, -32600, JSONRPC_INVALID_REQUEST, "Invalid request error code should match JSON-RPC standard")
	assert.Equal(t, -32001, API_AUTHENTICATION_REQUIRED, "Authentication error code should match API documentation")
	assert.Equal(t, -32004, API_CAMERA_NOT_FOUND, "Camera not found error code should match API documentation")

	// Test WebSocket timeout constants
	assert.Equal(t, 5*time.Second, WEBSOCKET_READ_TIMEOUT, "Read timeout should be 5 seconds")
	assert.Equal(t, 30*time.Second, WEBSOCKET_PING_INTERVAL, "Ping interval should be 30 seconds")
	assert.Equal(t, 8002, WEBSOCKET_DEFAULT_PORT, "Default port should be 8002")

	// Test status value constants
	assert.Equal(t, "CONNECTED", CAMERA_STATUS_CONNECTED, "Camera connected status should match API documentation")
	assert.Equal(t, "RECORDING", RECORDING_STATUS_RECORDING, "Recording status should match API documentation")
	assert.Equal(t, "2.0", JSONRPC_VERSION, "JSON-RPC version should be 2.0")
}

// TestAPIErrorMessages verifies error message mapping
func TestAPIErrorMessages(t *testing.T) {
	// Test that error messages exist for all error codes
	assert.NotEmpty(t, GetAPIErrorMessage(API_CAMERA_NOT_FOUND), "Should have error message for camera not found")
	assert.NotEmpty(t, GetAPIErrorMessage(JSONRPC_INVALID_REQUEST), "Should have error message for invalid request")
	assert.Equal(t, "Unknown error", GetAPIErrorMessage(999999), "Should return unknown error for invalid code")
}

// TestValidationHelpers verifies validation helper functions
func TestValidationHelpers(t *testing.T) {
	// Test camera status validation
	assert.True(t, IsValidCameraStatus(CAMERA_STATUS_CONNECTED), "CONNECTED should be valid camera status")
	assert.True(t, IsValidCameraStatus(CAMERA_STATUS_DISCONNECTED), "DISCONNECTED should be valid camera status")
	assert.False(t, IsValidCameraStatus("INVALID"), "INVALID should not be valid camera status")

	// Test recording format validation
	assert.True(t, IsValidRecordingFormat(RECORDING_FORMAT_FMP4), "fmp4 should be valid recording format")
	assert.True(t, IsValidRecordingFormat(RECORDING_FORMAT_MP4), "mp4 should be valid recording format")
	assert.False(t, IsValidRecordingFormat("invalid"), "invalid should not be valid recording format")
}
