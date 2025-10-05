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
	asserter := NewE2EWorkflowAsserter(t)
	
	// Connect and authenticate using proven flow
	err := asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err, "Authentication should succeed")
	
	// Get camera list using proven client method
	response, err := asserter.GetCameraList()
	require.NoError(t, err, "Get camera list should succeed")
	require.Nil(t, response.Error, "Should not have error")
	
	// Verify response structure
	result := response.Result.(map[string]interface{})
	require.Contains(t, result, "cameras")
	cameras := result["cameras"].([]interface{})
	assert.NotEmpty(t, cameras, "Should have at least one camera")
	
	// Verify camera structure
	for _, camera := range cameras {
		cam := camera.(map[string]interface{})
		assert.Contains(t, cam, "device", "Camera should have device field")
		assert.Contains(t, cam, "status", "Camera should have status field")
		assert.Contains(t, cam, "streams", "Camera should have streams field")
	}
}

func TestCameraStatusQueryWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err)
	
	// Get camera list first
	listResp, err := asserter.GetCameraList()
	require.NoError(t, err)
	require.Nil(t, listResp.Error)
	
	result := listResp.Result.(map[string]interface{})
	cameras := result["cameras"].([]interface{})
	require.NotEmpty(t, cameras)
	
	// Get first camera ID
	firstCamera := cameras[0].(map[string]interface{})
	deviceID := firstCamera["device"].(string)
	
	// Query camera status using proven client method
	statusResp, err := asserter.GetCameraStatus(deviceID)
	require.NoError(t, err)
	require.Nil(t, statusResp.Error)
	
	// Verify status response
	status := statusResp.Result.(map[string]interface{})
	assert.Equal(t, deviceID, status["device"])
	assert.Contains(t, status, "status")
	assert.Contains(t, status, "streams")
}

func TestCameraCapabilitiesWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err)
	
	// Get cameras and query capabilities
	listResp, err := asserter.GetCameraList()
	require.NoError(t, err)
	require.Nil(t, listResp.Error)
	
	result := listResp.Result.(map[string]interface{})
	cameras := result["cameras"].([]interface{})
	
	for _, camera := range cameras {
		cam := camera.(map[string]interface{})
		deviceID := cam["device"].(string)
		
		// Get capabilities using proven client method
		capsResp, err := asserter.GetCameraCapabilities(deviceID)
		require.NoError(t, err)
		require.Nil(t, capsResp.Error)
		
		caps := capsResp.Result.(map[string]interface{})
		assert.Contains(t, caps, "formats")
		assert.Contains(t, caps, "resolutions")
		assert.Contains(t, caps, "framerates")
	}
}