/*
WebSocket Untested Methods Integration Tests

Tests for API methods that have 0% coverage and need comprehensive testing.
These methods are implemented but not covered by existing tests.

API Documentation Reference: docs/api/json_rpc_methods.md
Requirements Coverage:
- REQ-API-001: Complete JSON-RPC method coverage
- REQ-API-002: API specification compliance
- REQ-API-003: Untested method validation

Design Principles:
- Real components only (no mocks)
- Fixture-driven configuration
- Progressive Readiness pattern validation
- Complete API specification compliance
- Untested method identification and testing

Untested Methods (0% coverage):
- delete_recording
- start_streaming
- stop_streaming
- add_external_stream
- remove_external_stream
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// UNTESTED RECORDING METHODS
// ============================================================================

// TestUntestedAPI_DeleteRecording_Integration tests delete_recording method
func TestUntestedAPI_DeleteRecording_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test delete_recording method with non-existent file (should return error)
	response, err := asserter.client.DeleteRecording("non_existent_recording.mp4")
	require.NoError(t, err, "delete_recording should not fail on client side")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)

	// Should return error for non-existent file - this is expected behavior
	require.NotNil(t, response.Error, "Should return error for non-existent file")
	t.Log("✅ Delete Recording: Properly handles non-existent file")
}

// ============================================================================
// UNTESTED STREAMING METHODS
// ============================================================================

// TestUntestedAPI_StartStreaming_Integration tests start_streaming method
func TestUntestedAPI_StartStreaming_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test start_streaming method
	device := testutils.GetTestCameraID()
	response, err := asserter.client.StartStreaming(device)
	require.NoError(t, err, "start_streaming should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
	t.Log("✅ Start Streaming: Method call successful")
}

// TestUntestedAPI_StopStreaming_Integration tests stop_streaming method
func TestUntestedAPI_StopStreaming_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test stop_streaming method
	device := testutils.GetTestCameraID()
	response, err := asserter.client.StopStreaming(device)
	require.NoError(t, err, "stop_streaming should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
	t.Log("✅ Stop Streaming: Method call successful")
}

// ============================================================================
// UNTESTED EXTERNAL STREAM METHODS
// ============================================================================

// TestUntestedAPI_AddExternalStream_Integration tests add_external_stream method
func TestUntestedAPI_AddExternalStream_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test add_external_stream method (feature now enabled in test config)
	streamURL := "rtsp://example.com/stream"
	streamName := "external_test_stream"
	response, err := asserter.client.AddExternalStream(streamURL, streamName)
	require.NoError(t, err, "add_external_stream should not fail on client side")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)

	// Should succeed for enabled feature - this is expected behavior
	require.Nil(t, response.Error, "Should not return error for enabled external stream feature")
	t.Log("✅ Add External Stream: Successfully adds external stream")
}

// TestUntestedAPI_RemoveExternalStream_Integration tests remove_external_stream method
func TestUntestedAPI_RemoveExternalStream_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test remove_external_stream method (feature now enabled in test config)
	// First add a stream, then remove it to test the full workflow
	streamURL := "rtsp://example.com/stream"
	streamName := "test_external_stream"

	// Add the stream first
	addResponse, err := asserter.client.AddExternalStream(streamURL, streamName)
	require.NoError(t, err, "add_external_stream should not fail on client side")
	require.NotNil(t, addResponse, "Add response should not be nil")
	require.Nil(t, addResponse.Error, "Add should succeed")

	// Now remove the stream
	response, err := asserter.client.RemoveExternalStream(streamURL)
	require.NoError(t, err, "remove_external_stream should not fail on client side")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)

	// Should succeed for enabled feature - this is expected behavior
	require.Nil(t, response.Error, "Should not return error for enabled external stream feature")
	t.Log("✅ Remove External Stream: Successfully removes external stream")
}

// ============================================================================
// COVERAGE IMPROVEMENT PATTERNS
// ============================================================================

// TestCoveragePatterns_AsserterUsage demonstrates asserter patterns for coverage
func TestCoveragePatterns_AsserterUsage(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Pattern 1: Basic method testing
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Pattern 2: Error path testing
	response, err := asserter.client.DeleteRecording("non_existent.mp4")
	require.NoError(t, err, "Client should not fail")
	require.NotNil(t, response, "Response should not be nil")
	require.NotNil(t, response.Error, "Should return error for non-existent file")

	// Pattern 3: Success path testing
	device := testutils.GetTestCameraID()
	response, err = asserter.client.StartStreaming(device)
	require.NoError(t, err, "start_streaming should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Pattern 4: Response validation
	asserter.client.AssertJSONRPCResponse(response, false)

	t.Log("✅ Coverage Patterns: All asserter patterns validated")
}
