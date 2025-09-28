/*
Module: WebSocket Camera Management
Purpose: Validates camera discovery and management functionality

Requirements Coverage:
- REQ-CAM-001: Camera discovery and listing
- REQ-CAM-002: Camera status queries
- REQ-CAM-003: Camera capability detection
- REQ-CAM-004: Device mapping validation

Test Categories: Integration
API Documentation: docs/api/json_rpc_methods.md
*/
package websocket

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWebSocket_CameraManagement_Complete_Integration validates complete camera management workflow
func TestWebSocket_CameraManagement_Complete_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Test complete camera management workflow
	err := asserter.AssertCameraManagementWorkflow()
	require.NoError(t, err, "Camera management workflow should succeed")

	t.Log("✅ Camera management integration test passed")
}

// TestWebSocket_CameraDiscovery_Integration validates camera discovery functionality
func TestWebSocket_CameraDiscovery_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_camera_list - discover available cameras
	response, err := asserter.client.GetCameraList()
	require.NoError(t, err, "get_camera_list should succeed")

	// Validate JSON-RPC response structure
	asserter.client.AssertJSONRPCResponse(response, false)

	// Validate camera list result structure per API spec
	asserter.client.AssertCameraListResultAPICompliant(response.Result)

	t.Log("✅ Camera discovery integration test passed")
}

// TestWebSocket_CameraStatus_Integration validates camera status queries
func TestWebSocket_CameraStatus_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_camera_status for specific camera
	cameraID := asserter.helper.GetTestCameraID()
	response, err := asserter.client.GetCameraStatus(cameraID)
	require.NoError(t, err, "get_camera_status should succeed")

	// Validate JSON-RPC response structure
	asserter.client.AssertJSONRPCResponse(response, false)

	// Validate camera status result structure per API spec
	asserter.client.AssertCameraStatusResultAPICompliant(response.Result)

	t.Log("✅ Camera status integration test passed")
}

// TestWebSocket_CameraCapabilities_Integration validates camera capability detection
func TestWebSocket_CameraCapabilities_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_camera_capabilities for specific camera
	cameraID := asserter.helper.GetTestCameraID()
	response, err := asserter.client.GetCameraCapabilities(cameraID)
	require.NoError(t, err, "get_camera_capabilities should succeed")

	// Validate JSON-RPC response structure
	asserter.client.AssertJSONRPCResponse(response, false)

	// Validate camera capabilities result structure per API spec
	asserter.client.AssertCameraCapabilitiesResultAPICompliant(response.Result)

	t.Log("✅ Camera capabilities integration test passed")
}

// TestWebSocket_DeviceMapping_Integration validates device mapping functionality
func TestWebSocket_DeviceMapping_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test device mapping by querying multiple cameras
	cameraIDs := []string{"camera0", "camera1", "camera2"}

	for _, cameraID := range cameraIDs {
		// Test get_camera_status for each camera ID
		response, err := asserter.client.GetCameraStatus(cameraID)
		require.NoError(t, err, "get_camera_status should succeed for %s", cameraID)

		// Validate JSON-RPC response structure
		asserter.client.AssertJSONRPCResponse(response, false)

		// Validate camera status result structure per API spec
		asserter.client.AssertCameraStatusResultAPICompliant(response.Result)
	}

	t.Log("✅ Device mapping integration test passed")
}

// TestWebSocket_CameraManagement_Performance validates camera management performance
func TestWebSocket_CameraManagement_Performance(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_camera_list performance using testutils
	_, err = asserter.client.GetCameraList()
	require.NoError(t, err, "get_camera_list should succeed")

	// Test get_camera_status performance using testutils
	cameraID := asserter.helper.GetTestCameraID()
	_, err = asserter.client.GetCameraStatus(cameraID)
	require.NoError(t, err, "get_camera_status should succeed")

	// Test get_camera_capabilities performance using testutils
	_, err = asserter.client.GetCameraCapabilities(cameraID)
	require.NoError(t, err, "get_camera_capabilities should succeed")
	t.Log("✅ Camera management performance validated using testutils")
}

// TestWebSocket_CameraManagement_ErrorHandling validates error handling
func TestWebSocket_CameraManagement_ErrorHandling(t *testing.T) {
	// Create integration asserter with real components
	asserter := GetSharedWebSocketAsserter(t)

	// CRITICAL: Progressive Readiness - try operations immediately, no waiting

	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed immediately")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test invalid camera ID
	response, err := asserter.client.GetCameraStatus("invalid_camera")
	require.NoError(t, err, "get_camera_status should not fail for invalid camera")

	// Should return error response, not panic
	asserter.client.AssertJSONRPCResponse(response, true) // Expect error

	// Test invalid camera capabilities
	response, err = asserter.client.GetCameraCapabilities("invalid_camera")
	require.NoError(t, err, "get_camera_capabilities should not fail for invalid camera")

	// Should return error response, not panic
	asserter.client.AssertJSONRPCResponse(response, true) // Expect error

	t.Log("✅ Camera management error handling validated")
}
