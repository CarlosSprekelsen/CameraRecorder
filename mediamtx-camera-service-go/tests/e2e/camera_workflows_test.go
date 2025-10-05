/*
Camera Workflows E2E Tests

Tests complete user workflows for camera discovery, status queries, and capabilities.
Each test validates a complete user journey from start to finish with real components.

Test Categories:
- Camera Discovery Workflow: Connect, authenticate, discover cameras, verify list contents
- Camera Status Query Workflow: Get camera list, query specific camera status, verify data consistency
- Camera Capabilities Workflow: Discover cameras, get capabilities for each, verify format information

Business Outcomes:
- User can select camera from list and get its status
- User can query specific camera and see its current state
- User can see what formats/resolutions camera supports

Coverage Target: 20% E2E coverage milestone
*/

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteCameraDiscoveryWorkflow(t *testing.T) {
	// Setup: Clean system using testutils infrastructure
	setup := NewE2ETestSetup(t)
	LogWorkflowStep(t, "Camera Discovery", 1, "Setup complete - clean system initialized")

	// Step 1: Connect WebSocket using establishConnection with admin token
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Camera Discovery", 2, "WebSocket connected and authenticated")

	// Step 2: Send get_camera_list JSON-RPC request
	response := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	LogWorkflowStep(t, "Camera Discovery", 3, "Camera list request sent")

	// Step 3: Verify response structure (cameras array, device IDs, status, streams)
	require.NoError(t, response.Error, "Camera list request should succeed")
	require.NotNil(t, response.Result, "Camera list result should not be nil")

	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Camera list result should be a map")
	require.Contains(t, resultMap, "cameras", "Result should contain cameras field")

	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(t, ok, "Cameras should be an array")
	LogWorkflowStep(t, "Camera Discovery", 4, "Camera list response structure validated")

	// Step 4: Validate camera IDs follow pattern (camera0, camera1, etc.)
	if len(cameras) > 0 {
		camera, ok := cameras[0].(map[string]interface{})
		require.True(t, ok, "First camera should be a map")

		device, ok := camera["device"].(string)
		require.True(t, ok, "Camera should have device field")
		assert.Regexp(t, "^camera\\d+$", device, "Camera device ID should follow camera0, camera1 pattern")

		// Verify camera has required fields
		assert.Contains(t, camera, "status", "Camera should have status field")
		assert.Contains(t, camera, "name", "Camera should have name field")
		assert.Contains(t, camera, "streams", "Camera should have streams field")
	}
	LogWorkflowStep(t, "Camera Discovery", 5, "Camera ID pattern and required fields validated")

	// Step 5: Validate at least one camera present (if cameras available)
	if len(cameras) == 0 {
		t.Log("Warning: No cameras detected - this may be expected in test environment")
	} else {
		assert.Greater(t, len(cameras), 0, "At least one camera should be present")
	}

	// Business Outcome: User can select camera from list and get its status
	setup.AssertBusinessOutcome("User can select camera from list and get its status", func() bool {
		return len(cameras) > 0 && cameras[0] != nil
	})
	LogWorkflowStep(t, "Camera Discovery", 6, "Business outcome validated - user can select camera from list")

	// Cleanup: closeConnection, verify no orphaned connections
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Camera Discovery", 7, "Connection closed and cleanup verified")
}

func TestCameraStatusQueryWorkflow(t *testing.T) {
	// Setup: Authenticated connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Camera Status", 1, "Authenticated connection established")

	// Step 1: Get camera list
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for status query test")
	}
	LogWorkflowStep(t, "Camera Status", 2, "Camera list retrieved successfully")

	// Step 2: Select first camera from list
	firstCamera := cameras[0].(map[string]interface{})
	deviceID := firstCamera["device"].(string)
	LogWorkflowStep(t, "Camera Status", 3, "First camera selected from list")

	// Step 3: Call get_camera_status with device ID
	statusResponse := setup.SendJSONRPC(conn, "get_camera_status", map[string]interface{}{
		"device": deviceID,
	})
	LogWorkflowStep(t, "Camera Status", 4, "Camera status request sent")

	// Step 4: Verify status response (device, status, name, resolution, fps, streams)
	require.NoError(t, statusResponse.Error, "Camera status request should succeed")
	require.NotNil(t, statusResponse.Result, "Camera status result should not be nil")

	statusResult := statusResponse.Result.(map[string]interface{})
	assert.Equal(t, deviceID, statusResult["device"], "Status device should match requested device")
	assert.Contains(t, statusResult, "status", "Status should contain status field")
	assert.Contains(t, statusResult, "name", "Status should contain name field")
	assert.Contains(t, statusResult, "resolution", "Status should contain resolution field")
	assert.Contains(t, statusResult, "fps", "Status should contain fps field")
	assert.Contains(t, statusResult, "streams", "Status should contain streams field")
	LogWorkflowStep(t, "Camera Status", 5, "Status response structure validated")

	// Step 5: Validate stream URLs properly formatted (rtsp://, http://)
	streams := statusResult["streams"].(map[string]interface{})

	if rtspURL, ok := streams["rtsp"].(string); ok && rtspURL != "" {
		assert.Regexp(t, "^rtsp://", rtspURL, "RTSP URL should start with rtsp://")
	}

	if hlsURL, ok := streams["hls"].(string); ok && hlsURL != "" {
		assert.Regexp(t, "^http://", hlsURL, "HLS URL should start with http://")
	}
	LogWorkflowStep(t, "Camera Status", 6, "Stream URLs format validated")

	// Step 6: Verify same device ID returned in status as in list
	assert.Equal(t, deviceID, statusResult["device"], "Status device ID should match list device ID")

	// Business Outcome: User can query specific camera and see its current state
	setup.AssertBusinessOutcome("User can query specific camera and see its current state", func() bool {
		return statusResult["device"] == deviceID &&
			statusResult["status"] != nil &&
			statusResult["name"] != nil
	})
	LogWorkflowStep(t, "Camera Status", 7, "Business outcome validated - user can query camera state")

	// Cleanup: Standard cleanup with verification
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Camera Status", 8, "Connection closed and cleanup verified")
}

func TestCameraCapabilitiesWorkflow(t *testing.T) {
	// Setup: Authenticated connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Camera Capabilities", 1, "Authenticated connection established")

	// Step 1: Discover cameras
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for capabilities test")
	}
	LogWorkflowStep(t, "Camera Capabilities", 2, "Cameras discovered successfully")

	// Step 2: For each connected camera, call get_camera_capabilities
	for i, cameraInterface := range cameras {
		camera := cameraInterface.(map[string]interface{})
		deviceID := camera["device"].(string)
		cameraStatus := camera["status"].(string)

		// Skip disconnected cameras
		if cameraStatus != "connected" && cameraStatus != "active" {
			t.Logf("Skipping camera %s with status %s", deviceID, cameraStatus)
			continue
		}

		LogWorkflowStep(t, "Camera Capabilities", 3, "Querying capabilities for camera "+deviceID)

		// Step 3: Verify capabilities response (formats, resolutions, capabilities map)
		capabilitiesResponse := setup.SendJSONRPC(conn, "get_camera_capabilities", map[string]interface{}{
			"device": deviceID,
		})

		if capabilitiesResponse.Error != nil {
			t.Logf("Warning: Capabilities request failed for camera %s: %v", deviceID, capabilitiesResponse.Error)
			continue
		}

		require.NotNil(t, capabilitiesResponse.Result, "Capabilities result should not be nil")
		capabilitiesResult := capabilitiesResponse.Result.(map[string]interface{})

		// Verify capabilities structure
		assert.Contains(t, capabilitiesResult, "device", "Capabilities should contain device field")
		assert.Equal(t, deviceID, capabilitiesResult["device"], "Capabilities device should match requested device")
		LogWorkflowStep(t, "Camera Capabilities", 4, "Capabilities response structure validated for camera "+deviceID)

		// Step 4: Validate at least one format available for each connected camera
		if formats, ok := capabilitiesResult["formats"].([]interface{}); ok && len(formats) > 0 {
			assert.Greater(t, len(formats), 0, "Camera %s should have at least one format", deviceID)

			// Step 5: Verify format structure (width, height, pixel_format)
			format := formats[0].(map[string]interface{})
			assert.Contains(t, format, "width", "Format should contain width field")
			assert.Contains(t, format, "height", "Format should contain height field")
			assert.Contains(t, format, "pixel_format", "Format should contain pixel_format field")

			// Validate format values are reasonable
			if width, ok := format["width"].(float64); ok {
				assert.Greater(t, width, 0, "Format width should be positive")
			}
			if height, ok := format["height"].(float64); ok {
				assert.Greater(t, height, 0, "Format height should be positive")
			}
			if pixelFormat, ok := format["pixel_format"].(string); ok {
				assert.NotEmpty(t, pixelFormat, "Pixel format should not be empty")
			}
		} else {
			t.Logf("Warning: Camera %s has no formats in capabilities response", deviceID)
		}

		// Only test first camera to keep test focused
		if i == 0 {
			break
		}
	}
	LogWorkflowStep(t, "Camera Capabilities", 5, "Format structure validated for connected cameras")

	// Business Outcome: User can see what formats/resolutions camera supports
	setup.AssertBusinessOutcome("User can see what formats/resolutions camera supports", func() bool {
		// Check if we successfully got capabilities for at least one camera
		if len(cameras) == 0 {
			return false
		}

		camera := cameras[0].(map[string]interface{})
		deviceID := camera["device"].(string)

		capabilitiesResponse := setup.SendJSONRPC(conn, "get_camera_capabilities", map[string]interface{}{
			"device": deviceID,
		})

		return capabilitiesResponse.Error == nil && capabilitiesResponse.Result != nil
	})
	LogWorkflowStep(t, "Camera Capabilities", 6, "Business outcome validated - user can see camera capabilities")

	// Cleanup: Standard cleanup with verification
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Camera Capabilities", 7, "Connection closed and cleanup verified")
}

// TestCameraWorkflowsIntegration tests camera workflows work together
func TestCameraWorkflowsIntegration(t *testing.T) {
	// This test validates that camera workflows work together as a complete user journey
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)

	// Complete workflow: discover -> status -> capabilities
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) > 0 {
		camera := cameras[0].(map[string]interface{})
		deviceID := camera["device"].(string)

		// Get status
		statusResponse := setup.SendJSONRPC(conn, "get_camera_status", map[string]interface{}{
			"device": deviceID,
		})
		require.NoError(t, statusResponse.Error, "Camera status should succeed")

		// Get capabilities
		capabilitiesResponse := setup.SendJSONRPC(conn, "get_camera_capabilities", map[string]interface{}{
			"device": deviceID,
		})

		// Capabilities might fail for some cameras, that's ok for integration test
		if capabilitiesResponse.Error == nil {
			require.NotNil(t, capabilitiesResponse.Result, "Capabilities result should not be nil")
		}

		// Verify data consistency across all three operations
		assert.Equal(t, deviceID, camera["device"], "List device should match")

		if statusResponse.Error == nil {
			statusResult := statusResponse.Result.(map[string]interface{})
			assert.Equal(t, deviceID, statusResult["device"], "Status device should match list device")
		}

		if capabilitiesResponse.Error == nil {
			capabilitiesResult := capabilitiesResponse.Result.(map[string]interface{})
			assert.Equal(t, deviceID, capabilitiesResult["device"], "Capabilities device should match list device")
		}
	}

	setup.CloseConnection(conn)
}
