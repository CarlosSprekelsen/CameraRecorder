/*
WebSocket Test Asserters - Workflow Validation Testing

Provides workflow asserters for WebSocket integration testing that validate
complete user workflows against the OpenRPC API specification.

API Documentation Reference: docs/api/mediamtx_camera_service_openrpc.json
Requirements Coverage:
- REQ-WS-001: WebSocket connection and authentication
- REQ-WS-002: Real-time camera operations
- REQ-WS-003: Error handling and recovery
- REQ-WS-004: Concurrent client support
- REQ-WS-005: Session management

Design Principles:
- Complete workflow validation
- OpenRPC API compliance
- Progressive Readiness testing
- Real component integration
*/

package websocket

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// WebSocketIntegrationAsserter handles complete WebSocket integration workflows
type WebSocketIntegrationAsserter struct {
	t      *testing.T
	helper *WebSocketTestHelper
	client *WebSocketTestClient
}

// NewWebSocketIntegrationAsserter creates a new WebSocket integration asserter
func NewWebSocketIntegrationAsserter(t *testing.T) *WebSocketIntegrationAsserter {
	helper := NewWebSocketTestHelper(t)

	// Create real WebSocket server
	err := helper.CreateRealServer()
	require.NoError(t, err, "Failed to create real WebSocket server")

	// Create WebSocket client
	client := NewWebSocketTestClient(t, helper.GetServerURL())

	asserter := &WebSocketIntegrationAsserter{
		t:      t,
		helper: helper,
		client: client,
	}

	// Register cleanup
	t.Cleanup(func() {
		asserter.Cleanup()
	})

	return asserter
}

// Cleanup performs cleanup of all resources
func (a *WebSocketIntegrationAsserter) Cleanup() {
	if a.client != nil {
		a.client.Close()
	}
	// Helper cleanup is handled by its own cleanup
}

// AssertProgressiveReadiness validates Progressive Readiness behavior
func (a *WebSocketIntegrationAsserter) AssertProgressiveReadiness() error {
	// Test immediate connection acceptance (Progressive Readiness pattern)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	// Test that ping works immediately (no authentication required)
	err = a.client.Ping()
	require.NoError(a.t, err, "Ping should work immediately without authentication")

	a.t.Log("✅ Progressive Readiness validated: immediate connection and ping")
	return nil
}

// AssertAuthenticationWorkflow validates complete authentication workflow
func (a *WebSocketIntegrationAsserter) AssertAuthenticationWorkflow() error {
	// Connect to WebSocket (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	// Test ping before authentication (should work)
	err = a.client.Ping()
	require.NoError(a.t, err, "Ping should work before authentication")

	// Get JWT token for testing
	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	// Authenticate with JWT token
	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Authentication should succeed")

	a.t.Log("✅ Authentication workflow validated")
	return nil
}

// AssertCameraManagementWorkflow validates camera management operations
func (a *WebSocketIntegrationAsserter) AssertCameraManagementWorkflow() error {
	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Authentication should succeed")

	// Test get_camera_list
	response, err := a.client.GetCameraList()
	require.NoError(a.t, err, "get_camera_list should succeed")

	a.client.AssertJSONRPCResponse(response, false)
	a.client.AssertCameraListResult(response.Result)

	// Test get_camera_status for specific camera
	cameraID := a.helper.GetTestCameraID()
	response, err = a.client.GetCameraStatus(cameraID)
	require.NoError(a.t, err, "get_camera_status should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	a.t.Log("✅ Camera management workflow validated")
	return nil
}

// AssertRecordingWorkflow validates complete recording workflow
func (a *WebSocketIntegrationAsserter) AssertRecordingWorkflow() error {
	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Authentication should succeed")

	cameraID := a.helper.GetTestCameraID()

	// Test start_recording
	response, err := a.client.StartRecording(cameraID, 10, "mp4")
	require.NoError(a.t, err, "start_recording should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	// Wait a bit for recording to start
	time.Sleep(testutils.UniversalTimeoutShort)

	// Test stop_recording
	response, err = a.client.StopRecording(cameraID)
	require.NoError(a.t, err, "stop_recording should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	// Test list_recordings
	response, err = a.client.ListRecordings(50, 0)
	require.NoError(a.t, err, "list_recordings should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	a.t.Log("✅ Recording workflow validated")
	return nil
}

// AssertSnapshotWorkflow validates snapshot workflow
func (a *WebSocketIntegrationAsserter) AssertSnapshotWorkflow() error {
	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Authentication should succeed")

	cameraID := a.helper.GetTestCameraID()
	filename := "test_snapshot_" + time.Now().Format("2006-01-02_15-04-05") + ".jpg"

	// Test take_snapshot
	response, err := a.client.TakeSnapshot(cameraID, filename)
	require.NoError(a.t, err, "take_snapshot should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	// Test list_snapshots
	response, err = a.client.ListSnapshots(50, 0)
	require.NoError(a.t, err, "list_snapshots should succeed")

	a.client.AssertJSONRPCResponse(response, false)

	a.t.Log("✅ Snapshot workflow validated")
	return nil
}

// AssertErrorRecoveryWorkflow validates error handling and recovery
func (a *WebSocketIntegrationAsserter) AssertErrorRecoveryWorkflow() error {
	// Test invalid authentication (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	// Try authentication with invalid token
	err = a.client.Authenticate("invalid_token")
	require.Error(a.t, err, "Authentication with invalid token should fail")

	// Test ping still works after authentication failure
	err = a.client.Ping()
	require.NoError(a.t, err, "Ping should still work after authentication failure")

	// Test valid authentication after failure
	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Valid authentication should succeed after failure")

	a.t.Log("✅ Error recovery workflow validated")
	return nil
}

// AssertPerformanceRequirements validates performance guarantees
func (a *WebSocketIntegrationAsserter) AssertPerformanceRequirements() error {
	// Connect and authenticate (Progressive Readiness - immediate acceptance)
	err := a.client.Connect()
	require.NoError(a.t, err, "WebSocket connection should succeed immediately")

	authToken, err := a.helper.GetJWTToken("operator")
	require.NoError(a.t, err, "Should be able to create JWT token")

	err = a.client.Authenticate(authToken)
	require.NoError(a.t, err, "Authentication should succeed")

	// Test status method performance (<50ms)
	start := time.Now()
	_, err = a.client.GetCameraList()
	require.NoError(a.t, err, "get_camera_list should succeed")
	statusTime := time.Since(start)
	require.Less(a.t, statusTime, 50*time.Millisecond,
		"Status method should be <50ms, got %v", statusTime)

	// Test control method performance (<100ms)
	start = time.Now()
	cameraID := a.helper.GetTestCameraID()
	_, err = a.client.TakeSnapshot(cameraID, "perf_test.jpg")
	require.NoError(a.t, err, "take_snapshot should succeed")
	controlTime := time.Since(start)
	require.Less(a.t, controlTime, 100*time.Millisecond,
		"Control method should be <100ms, got %v", controlTime)

	a.t.Log("✅ Performance requirements validated")
	return nil
}

// AssertCompleteWorkflow validates complete end-to-end workflow
func (a *WebSocketIntegrationAsserter) AssertCompleteWorkflow() error {
	// Progressive Readiness
	err := a.AssertProgressiveReadiness()
	require.NoError(a.t, err, "Progressive Readiness should work")

	// Authentication
	err = a.AssertAuthenticationWorkflow()
	require.NoError(a.t, err, "Authentication workflow should work")

	// Camera Management
	err = a.AssertCameraManagementWorkflow()
	require.NoError(a.t, err, "Camera management workflow should work")

	// Recording
	err = a.AssertRecordingWorkflow()
	require.NoError(a.t, err, "Recording workflow should work")

	// Snapshots
	err = a.AssertSnapshotWorkflow()
	require.NoError(a.t, err, "Snapshot workflow should work")

	// Error Recovery
	err = a.AssertErrorRecoveryWorkflow()
	require.NoError(a.t, err, "Error recovery workflow should work")

	// Performance
	err = a.AssertPerformanceRequirements()
	require.NoError(a.t, err, "Performance requirements should be met")

	a.t.Log("✅ Complete workflow validated")
	return nil
}
