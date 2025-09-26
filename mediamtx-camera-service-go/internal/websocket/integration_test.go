/*
WebSocket Integration Tests - Real Component Testing

Tests WebSocket functionality using real components and validates against
the OpenRPC API specification. Demonstrates Progressive Readiness pattern
and complete workflow validation.

API Documentation Reference: docs/api/mediamtx_camera_service_openrpc.json
Requirements Coverage:
- REQ-WS-001: WebSocket connection and authentication
- REQ-WS-002: Real-time camera operations
- REQ-WS-003: Error handling and recovery
- REQ-WS-004: Concurrent client support
- REQ-WS-005: Session management

Design Principles:
- Real components only (no mocks)
- Fixture-driven configuration
- Progressive Readiness pattern
- OpenRPC API compliance validation
- Complete workflow testing
*/

package websocket

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestWebSocket_ProgressiveReadiness_Integration validates Progressive Readiness behavior
func TestWebSocket_ProgressiveReadiness_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	t.Log("✅ Progressive Readiness integration test passed")
}

// TestWebSocket_Authentication_Integration validates authentication workflow
func TestWebSocket_Authentication_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test complete authentication workflow
	err := asserter.AssertAuthenticationWorkflow()
	require.NoError(t, err, "Authentication workflow should work")

	t.Log("✅ Authentication integration test passed")
}

// TestWebSocket_CameraManagement_Integration validates camera management operations
func TestWebSocket_CameraManagement_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test camera management workflow
	err := asserter.AssertCameraManagementWorkflow()
	require.NoError(t, err, "Camera management workflow should work")

	t.Log("✅ Camera management integration test passed")
}

// TestWebSocket_Recording_Integration validates recording workflow
func TestWebSocket_Recording_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test complete recording workflow
	err := asserter.AssertRecordingWorkflow()
	require.NoError(t, err, "Recording workflow should work")

	t.Log("✅ Recording integration test passed")
}

// TestWebSocket_Snapshot_Integration validates snapshot workflow
func TestWebSocket_Snapshot_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test snapshot workflow
	err := asserter.AssertSnapshotWorkflow()
	require.NoError(t, err, "Snapshot workflow should work")

	t.Log("✅ Snapshot integration test passed")
}

// TestWebSocket_ErrorRecovery_Integration validates error handling and recovery
func TestWebSocket_ErrorRecovery_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test error recovery workflow
	err := asserter.AssertErrorRecoveryWorkflow()
	require.NoError(t, err, "Error recovery workflow should work")

	t.Log("✅ Error recovery integration test passed")
}

// TestWebSocket_Performance_Integration validates performance requirements
func TestWebSocket_Performance_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test performance requirements
	err := asserter.AssertPerformanceRequirements()
	require.NoError(t, err, "Performance requirements should be met")

	t.Log("✅ Performance integration test passed")
}

// TestWebSocket_CompleteWorkflow_Integration validates complete end-to-end workflow
func TestWebSocket_CompleteWorkflow_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test complete workflow
	err := asserter.AssertCompleteWorkflow()
	require.NoError(t, err, "Complete workflow should work")

	t.Log("✅ Complete workflow integration test passed")
}

// TestWebSocket_ConcurrentClients_Integration validates concurrent client support
func TestWebSocket_ConcurrentClients_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test concurrent clients
	numClients := 3
	results := make(chan error, numClients)

	// Create multiple concurrent clients
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			client := NewWebSocketTestClient(t, asserter.helper.GetServerURL())
			defer client.Close()

			// Connect and authenticate
			err := client.Connect()
			if err != nil {
				results <- err
				return
			}

			authToken, err := asserter.helper.GetJWTToken("operator")
			if err != nil {
				results <- err
				return
			}

			err = client.Authenticate(authToken)
			if err != nil {
				results <- err
				return
			}

			// Test ping
			err = client.Ping()
			if err != nil {
				results <- err
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all clients to complete
	for i := 0; i < numClients; i++ {
		select {
		case err := <-results:
			require.NoError(t, err, "Concurrent client %d should work", i)
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent client test timed out")
		}
	}

	t.Log("✅ Concurrent clients integration test passed")
}

// TestWebSocket_SessionManagement_Integration validates session management
func TestWebSocket_SessionManagement_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test session management workflow
	client := NewWebSocketTestClient(t, asserter.helper.GetServerURL())
	defer client.Close()

	// Connect and authenticate
	err := client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test multiple operations in same session
	err = client.Ping()
	require.NoError(t, err, "Ping should work in authenticated session")

	_, err = client.GetCameraList()
	require.NoError(t, err, "get_camera_list should work in authenticated session")

	// Test session persistence
	time.Sleep(100 * time.Millisecond)

	err = client.Ping()
	require.NoError(t, err, "Ping should still work after delay")

	t.Log("✅ Session management integration test passed")
}

// TestWebSocket_OpenRPCCompliance_Integration validates OpenRPC API compliance
func TestWebSocket_OpenRPCCompliance_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test OpenRPC API compliance
	client := NewWebSocketTestClient(t, asserter.helper.GetServerURL())
	defer client.Close()

	// Connect and authenticate
	err := client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test all major API methods according to OpenRPC spec
	response, err := client.GetCameraList()
	require.NoError(t, err, "get_camera_list should succeed")
	client.AssertJSONRPCResponse(response, false)
	client.AssertCameraListResult(response.Result)

	cameraID := asserter.helper.GetTestCameraID()
	response, err = client.GetCameraStatus(cameraID)
	require.NoError(t, err, "get_camera_status should succeed")
	client.AssertJSONRPCResponse(response, false)

	response, err = client.ListRecordings(50, 0)
	require.NoError(t, err, "list_recordings should succeed")
	client.AssertJSONRPCResponse(response, false)

	response, err = client.ListSnapshots(50, 0)
	require.NoError(t, err, "list_snapshots should succeed")
	client.AssertJSONRPCResponse(response, false)

	t.Log("✅ OpenRPC compliance integration test passed")
}

// TestWebSocket_ProgressiveReadiness_Performance validates Progressive Readiness performance
func TestWebSocket_ProgressiveReadiness_Performance(t *testing.T) {
	// Use testutils for performance testing
	setup := testutils.SetupTest(t, "config_clean_minimal.yaml")
	defer setup.Cleanup()

	// Create helper and server
	helper := NewWebSocketTestHelper(t)
	err := helper.CreateRealServer()
	require.NoError(t, err, "Failed to create real WebSocket server")

	// Test multiple rapid connections (Progressive Readiness)
	numConnections := 10
	results := make(chan time.Duration, numConnections)

	for i := 0; i < numConnections; i++ {
		go func() {
			start := time.Now()
			client := NewWebSocketTestClient(t, helper.GetServerURL())
			defer client.Close()

			err := client.Connect()
			if err != nil {
				results <- -1 // Error marker
				return
			}

			err = client.Ping()
			if err != nil {
				results <- -1 // Error marker
				return
			}

			results <- time.Since(start)
		}()
	}

	// Collect results
	var totalTime time.Duration
	successCount := 0

	for i := 0; i < numConnections; i++ {
		select {
		case duration := <-results:
			if duration == -1 {
				t.Fatal("Connection failed")
			}
			totalTime += duration
			successCount++
			require.Less(t, duration, 100*time.Millisecond,
				"Connection %d took %v, should be <100ms", i, duration)
		case <-time.After(5 * time.Second):
			t.Fatal("Performance test timed out")
		}
	}

	avgTime := totalTime / time.Duration(successCount)
	require.Less(t, avgTime, 50*time.Millisecond,
		"Average connection time should be <50ms, got %v", avgTime)

	t.Logf("✅ Progressive Readiness performance validated: %d connections, avg %v",
		successCount, avgTime)
}
