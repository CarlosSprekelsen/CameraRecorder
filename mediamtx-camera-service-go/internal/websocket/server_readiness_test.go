// Package websocket implements tests for server readiness behavior
//
// This file contains the SINGLE test that validates server-not-ready behavior.
// All other tests must wait for server readiness before validating API contracts.
//
// Test Categories: Server Readiness
// Requirements Coverage: Progressive Readiness Pattern validation

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketServer_NotReady_APIContract validates server-not-ready behavior
// This is the ONLY test that should validate MEDIAMTX_UNAVAILABLE responses
func TestWebSocketServer_NotReady_APIContract(t *testing.T) {
	// Create helper but do NOT wait for readiness
	helper := NewWebSocketTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Test various API methods when server is not ready
	testCases := []struct {
		method string
		params map[string]interface{}
		role   string
	}{
		{"stop_recording", map[string]interface{}{"device": "camera0"}, "operator"},
		{"start_recording", map[string]interface{}{"device": "camera0"}, "operator"},
		{"take_snapshot", map[string]interface{}{"device": "camera0"}, "operator"},
		{"get_camera_list", nil, "viewer"},
		{"get_camera_status", map[string]interface{}{"device": "camera0"}, "viewer"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			response := helper.TestMethod(t, tc.method, tc.params, tc.role)

			// When server is not ready, should return MEDIAMTX_UNAVAILABLE
			require.NotNil(t, response.Error, "Server not ready should return error")
			assert.Equal(t, MEDIAMTX_UNAVAILABLE, response.Error.Code,
				"Server not ready should return MEDIAMTX_UNAVAILABLE")
			assert.Equal(t, "MediaMTX service unavailable", response.Error.Message,
				"Error message should match API specification")

			// Validate error data structure
			require.NotNil(t, response.Error.Data, "Error should have data")
			errorData, ok := response.Error.Data.(map[string]interface{})
			require.True(t, ok, "Error data should be map")
			assert.Equal(t, "service_initializing", errorData["reason"])
			assert.Contains(t, errorData["details"], "initializing")
		})
	}
}
