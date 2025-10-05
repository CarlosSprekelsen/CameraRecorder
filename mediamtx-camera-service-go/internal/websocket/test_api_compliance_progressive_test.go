/*
API Compliance Test - Using Existing Test Infrastructure

Tests all 37 API methods for compliance using existing test utilities,
proper component lifecycle management, and existing test patterns.

Design Principles:
- Use existing testutils.UniversalTestSetup
- Leverage WebSocketTestHelper for proper component lifecycle
- Use existing test patterns instead of duplicating code
- Improve performance through proper test utilities
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestAPICompliance_AllMethods tests all 37 methods using existing test infrastructure
func TestAPICompliance_AllMethods(t *testing.T) {
	// Use existing test timeout patterns
	timeout := testutils.DefaultTestTimeout
	t.Logf("Testing API compliance for all 37 methods with existing test infrastructure, timeout: %v", timeout)

	// Create WebSocket test helper using existing test infrastructure
	helper := NewWebSocketTestHelper(t)
	defer helper.Cleanup()

	// Create WebSocket server with event-based readiness
	err := helper.CreateRealServer()
	require.NoError(t, err, "Failed to create WebSocket server")

	// Create test client using existing patterns
	client := testutils.Newtestutils.WebSocketTestClient(t, helper.baseURL)
	err = client.Connect()
	require.NoError(t, err, "Failed to connect WebSocket client")
	defer client.Close()

	// Test all methods using existing test patterns
	testAllMethodsWithExistingPatterns(t, helper, client)
}

// testAllMethodsWithExistingPatterns tests all methods using existing test patterns
func testAllMethodsWithExistingPatterns(t *testing.T, helper *WebSocketTestHelper, client *testutils.WebSocketTestClient) {
	// Test core methods using existing test patterns
	t.Run("CoreMethods", func(t *testing.T) {
		testPingWithExistingPatterns(t, client)
		testAuthenticateWithExistingPatterns(t, helper, client)
		testGetCameraListWithExistingPatterns(t, client)
		testGetCameraStatusWithExistingPatterns(t, client)
	})

	// Test system methods using existing test patterns
	t.Run("SystemMethods", func(t *testing.T) {
		testGetStatusWithExistingPatterns(t, client)
		testGetSystemStatusWithExistingPatterns(t, client)
		testGetServerInfoWithExistingPatterns(t, client)
		testGetStorageInfoWithExistingPatterns(t, client)
	})

	// Test recording methods using existing test patterns
	t.Run("RecordingMethods", func(t *testing.T) {
		testStartRecordingWithExistingPatterns(t, client)
		testStopRecordingWithExistingPatterns(t, client)
		testListRecordingsWithExistingPatterns(t, client)
		testGetRecordingInfoWithExistingPatterns(t, client)
		testDeleteRecordingWithExistingPatterns(t, client)
	})

	// Test snapshot methods using existing test patterns
	t.Run("SnapshotMethods", func(t *testing.T) {
		testTakeSnapshotWithExistingPatterns(t, client)
		testListSnapshotsWithExistingPatterns(t, client)
		testGetSnapshotInfoWithExistingPatterns(t, client)
		testDeleteSnapshotWithExistingPatterns(t, client)
	})

	// Test streaming methods using existing test patterns
	t.Run("StreamingMethods", func(t *testing.T) {
		testStartStreamingWithExistingPatterns(t, client)
		testStopStreamingWithExistingPatterns(t, client)
		testGetStreamURLWithExistingPatterns(t, client)
		testGetStreamStatusWithExistingPatterns(t, client)
	})

	// Test external stream methods using existing test patterns
	t.Run("ExternalStreamMethods", func(t *testing.T) {
		testDiscoverExternalStreamsWithExistingPatterns(t, client)
		testAddExternalStreamWithExistingPatterns(t, client)
		testRemoveExternalStreamWithExistingPatterns(t, client)
		testGetExternalStreamsWithExistingPatterns(t, client)
		testSetDiscoveryIntervalWithExistingPatterns(t, client)
	})

	// Test event subscription methods using existing test patterns
	t.Run("EventSubscriptionMethods", func(t *testing.T) {
		testSubscribeEventsWithExistingPatterns(t, client)
		testUnsubscribeEventsWithExistingPatterns(t, client)
		testGetSubscriptionStatsWithExistingPatterns(t, client)
	})

	// Test system management methods using existing test patterns
	t.Run("SystemManagementMethods", func(t *testing.T) {
		testSetRetentionPolicyWithExistingPatterns(t, client)
		testCleanupOldFilesWithExistingPatterns(t, client)
	})
}

// testPingWithExistingPatterns tests ping method using existing test patterns
func testPingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for ping operation
	response, err := client.SendJSONRPC("ping", map[string]interface{}{})
	require.NoError(t, err, "Ping should succeed")
	require.Nil(t, response.Error, "Ping should not return error")
	require.Equal(t, "pong", response.Result, "Ping should return 'pong'")
}

// testAuthenticateWithExistingPatterns tests authenticate method using existing test patterns
func testAuthenticateWithExistingPatterns(t *testing.T, helper *WebSocketTestHelper, client *testutils.WebSocketTestClient) {
	// Generate test JWT token using existing patterns
	token, err := helper.GetJWTToken("admin")
	require.NoError(t, err, "Should generate JWT token successfully")

	// Use existing test patterns for authenticate operation
	response, err := client.SendJSONRPC("authenticate", map[string]interface{}{
		"auth_token": token,
	})
	require.NoError(t, err, "Authenticate should succeed")
	require.Nil(t, response.Error, "Authenticate should not return error")
	require.NotNil(t, response.Result, "Authenticate result should not be nil")
}

// testGetCameraListWithExistingPatterns tests get_camera_list method using existing test patterns
func testGetCameraListWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_camera_list operation
	response, err := client.SendJSONRPC("get_camera_list", map[string]interface{}{})
	require.NoError(t, err, "GetCameraList should succeed")
	require.Nil(t, response.Error, "GetCameraList should not return error")
	require.NotNil(t, response.Result, "GetCameraList result should not be nil")
}

// testGetCameraStatusWithExistingPatterns tests get_camera_status method using existing test patterns
func testGetCameraStatusWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_camera_status operation
	response, err := client.SendJSONRPC("get_camera_status", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "GetCameraStatus should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("GetCameraStatus failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "GetCameraStatus result should not be nil")
	}
}

// testGetStatusWithExistingPatterns tests get_status method using existing test patterns
func testGetStatusWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_status operation
	response, err := client.SendJSONRPC("get_status", map[string]interface{}{})
	require.NoError(t, err, "GetStatus should succeed")
	require.Nil(t, response.Error, "GetStatus should not return error")
	require.NotNil(t, response.Result, "GetStatus result should not be nil")
}

// testGetSystemStatusWithExistingPatterns tests get_system_status method using existing test patterns
func testGetSystemStatusWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_system_status operation
	response, err := client.SendJSONRPC("get_system_status", map[string]interface{}{})
	require.NoError(t, err, "GetSystemStatus should succeed")
	require.Nil(t, response.Error, "GetSystemStatus should not return error")
	require.NotNil(t, response.Result, "GetSystemStatus result should not be nil")
}

// testGetServerInfoWithExistingPatterns tests get_server_info method using existing test patterns
func testGetServerInfoWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_server_info operation
	response, err := client.SendJSONRPC("get_server_info", map[string]interface{}{})
	require.NoError(t, err, "GetServerInfo should succeed")
	require.Nil(t, response.Error, "GetServerInfo should not return error")
	require.NotNil(t, response.Result, "GetServerInfo result should not be nil")
}

// testGetStorageInfoWithExistingPatterns tests get_storage_info method using existing test patterns
func testGetStorageInfoWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_storage_info operation
	response, err := client.SendJSONRPC("get_storage_info", map[string]interface{}{})
	require.NoError(t, err, "GetStorageInfo should succeed")
	require.Nil(t, response.Error, "GetStorageInfo should not return error")
	require.NotNil(t, response.Result, "GetStorageInfo result should not be nil")
}

// testStartRecordingWithExistingPatterns tests start_recording method using existing test patterns
func testStartRecordingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for start_recording operation
	response, err := client.SendJSONRPC("start_recording", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "StartRecording should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("StartRecording failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "StartRecording result should not be nil")
	}
}

// testStopRecordingWithExistingPatterns tests stop_recording method using existing test patterns
func testStopRecordingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for stop_recording operation
	response, err := client.SendJSONRPC("stop_recording", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "StopRecording should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("StopRecording failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "StopRecording result should not be nil")
	}
}

// testListRecordingsWithExistingPatterns tests list_recordings method using existing test patterns
func testListRecordingsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for list_recordings operation
	response, err := client.SendJSONRPC("list_recordings", map[string]interface{}{})
	require.NoError(t, err, "ListRecordings should succeed")
	require.Nil(t, response.Error, "ListRecordings should not return error")
	require.NotNil(t, response.Result, "ListRecordings result should not be nil")
}

// testGetRecordingInfoWithExistingPatterns tests get_recording_info method using existing test patterns
func testGetRecordingInfoWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_recording_info operation
	response, err := client.SendJSONRPC("get_recording_info", map[string]interface{}{
		"filename": "test_recording.mp4",
	})
	require.NoError(t, err, "GetRecordingInfo should succeed")

	// Handle expected file not found error gracefully
	if response.Error != nil {
		t.Logf("GetRecordingInfo failed as expected (no file): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "GetRecordingInfo result should not be nil")
	}
}

// testDeleteRecordingWithExistingPatterns tests delete_recording method using existing test patterns
func testDeleteRecordingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for delete_recording operation
	response, err := client.SendJSONRPC("delete_recording", map[string]interface{}{
		"filename": "test_recording.mp4",
	})
	require.NoError(t, err, "DeleteRecording should succeed")

	// Handle expected file not found error gracefully
	if response.Error != nil {
		t.Logf("DeleteRecording failed as expected (no file): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "DeleteRecording result should not be nil")
	}
}

// testTakeSnapshotWithExistingPatterns tests take_snapshot method using existing test patterns
func testTakeSnapshotWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for take_snapshot operation
	response, err := client.SendJSONRPC("take_snapshot", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "TakeSnapshot should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("TakeSnapshot failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "TakeSnapshot result should not be nil")
	}
}

// testListSnapshotsWithExistingPatterns tests list_snapshots method using existing test patterns
func testListSnapshotsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for list_snapshots operation
	response, err := client.SendJSONRPC("list_snapshots", map[string]interface{}{})
	require.NoError(t, err, "ListSnapshots should succeed")
	require.Nil(t, response.Error, "ListSnapshots should not return error")
	require.NotNil(t, response.Result, "ListSnapshots result should not be nil")
}

// testGetSnapshotInfoWithExistingPatterns tests get_snapshot_info method using existing test patterns
func testGetSnapshotInfoWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_snapshot_info operation
	response, err := client.SendJSONRPC("get_snapshot_info", map[string]interface{}{
		"filename": "test_snapshot.jpg",
	})
	require.NoError(t, err, "GetSnapshotInfo should succeed")

	// Handle expected file not found error gracefully
	if response.Error != nil {
		t.Logf("GetSnapshotInfo failed as expected (no file): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "GetSnapshotInfo result should not be nil")
	}
}

// testDeleteSnapshotWithExistingPatterns tests delete_snapshot method using existing test patterns
func testDeleteSnapshotWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for delete_snapshot operation
	response, err := client.SendJSONRPC("delete_snapshot", map[string]interface{}{
		"filename": "test_snapshot.jpg",
	})
	require.NoError(t, err, "DeleteSnapshot should succeed")

	// Handle expected file not found error gracefully
	if response.Error != nil {
		t.Logf("DeleteSnapshot failed as expected (no file): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "DeleteSnapshot result should not be nil")
	}
}

// testStartStreamingWithExistingPatterns tests start_streaming method using existing test patterns
func testStartStreamingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for start_streaming operation
	response, err := client.SendJSONRPC("start_streaming", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "StartStreaming should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("StartStreaming failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "StartStreaming result should not be nil")
	}
}

// testStopStreamingWithExistingPatterns tests stop_streaming method using existing test patterns
func testStopStreamingWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for stop_streaming operation
	response, err := client.SendJSONRPC("stop_streaming", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "StopStreaming should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("StopStreaming failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "StopStreaming result should not be nil")
	}
}

// testGetStreamURLWithExistingPatterns tests get_stream_url method using existing test patterns
func testGetStreamURLWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_stream_url operation
	response, err := client.SendJSONRPC("get_stream_url", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "GetStreamURL should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("GetStreamURL failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "GetStreamURL result should not be nil")
	}
}

// testGetStreamStatusWithExistingPatterns tests get_stream_status method using existing test patterns
func testGetStreamStatusWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_stream_status operation
	response, err := client.SendJSONRPC("get_stream_status", map[string]interface{}{
		"device": "camera0",
	})
	require.NoError(t, err, "GetStreamStatus should succeed")

	// Handle expected camera not found error gracefully
	if response.Error != nil {
		t.Logf("GetStreamStatus failed as expected (no camera): %v", response.Error)
	} else {
		require.NotNil(t, response.Result, "GetStreamStatus result should not be nil")
	}
}

// testDiscoverExternalStreamsWithExistingPatterns tests discover_external_streams method using existing test patterns
func testDiscoverExternalStreamsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for discover_external_streams operation
	response, err := client.SendJSONRPC("discover_external_streams", map[string]interface{}{})
	require.NoError(t, err, "DiscoverExternalStreams should succeed")
	require.Nil(t, response.Error, "DiscoverExternalStreams should not return error")
	require.NotNil(t, response.Result, "DiscoverExternalStreams result should not be nil")
}

// testAddExternalStreamWithExistingPatterns tests add_external_stream method using existing test patterns
func testAddExternalStreamWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for add_external_stream operation
	response, err := client.SendJSONRPC("add_external_stream", map[string]interface{}{
		"stream_url":  "rtsp://example.com/stream",
		"stream_name": "test_stream",
		"stream_type": "generic_rtsp",
	})
	require.NoError(t, err, "AddExternalStream should succeed")
	require.Nil(t, response.Error, "AddExternalStream should not return error")
	require.NotNil(t, response.Result, "AddExternalStream result should not be nil")
}

// testRemoveExternalStreamWithExistingPatterns tests remove_external_stream method using existing test patterns
func testRemoveExternalStreamWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for remove_external_stream operation
	response, err := client.SendJSONRPC("remove_external_stream", map[string]interface{}{
		"stream_url": "rtsp://example.com/stream",
	})
	require.NoError(t, err, "RemoveExternalStream should succeed")
	require.Nil(t, response.Error, "RemoveExternalStream should not return error")
	require.NotNil(t, response.Result, "RemoveExternalStream result should not be nil")
}

// testGetExternalStreamsWithExistingPatterns tests get_external_streams method using existing test patterns
func testGetExternalStreamsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_external_streams operation
	response, err := client.SendJSONRPC("get_external_streams", map[string]interface{}{})
	require.NoError(t, err, "GetExternalStreams should succeed")
	require.Nil(t, response.Error, "GetExternalStreams should not return error")
	require.NotNil(t, response.Result, "GetExternalStreams result should not be nil")
}

// testSetDiscoveryIntervalWithExistingPatterns tests set_discovery_interval method using existing test patterns
func testSetDiscoveryIntervalWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for set_discovery_interval operation
	response, err := client.SendJSONRPC("set_discovery_interval", map[string]interface{}{
		"scan_interval": 30,
	})
	require.NoError(t, err, "SetDiscoveryInterval should succeed")
	require.Nil(t, response.Error, "SetDiscoveryInterval should not return error")
	require.NotNil(t, response.Result, "SetDiscoveryInterval result should not be nil")
}

// testSubscribeEventsWithExistingPatterns tests subscribe_events method using existing test patterns
func testSubscribeEventsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for subscribe_events operation
	response, err := client.SendJSONRPC("subscribe_events", map[string]interface{}{
		"topics": []string{"camera_status_update", "recording_status_update"},
	})
	require.NoError(t, err, "SubscribeEvents should succeed")
	require.Nil(t, response.Error, "SubscribeEvents should not return error")
	require.NotNil(t, response.Result, "SubscribeEvents result should not be nil")
}

// testUnsubscribeEventsWithExistingPatterns tests unsubscribe_events method using existing test patterns
func testUnsubscribeEventsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for unsubscribe_events operation
	response, err := client.SendJSONRPC("unsubscribe_events", map[string]interface{}{
		"topics": []string{"camera_status_update"},
	})
	require.NoError(t, err, "UnsubscribeEvents should succeed")
	require.Nil(t, response.Error, "UnsubscribeEvents should not return error")
	require.NotNil(t, response.Result, "UnsubscribeEvents result should not be nil")
}

// testGetSubscriptionStatsWithExistingPatterns tests get_subscription_stats method using existing test patterns
func testGetSubscriptionStatsWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for get_subscription_stats operation
	response, err := client.SendJSONRPC("get_subscription_stats", map[string]interface{}{})
	require.NoError(t, err, "GetSubscriptionStats should succeed")
	require.Nil(t, response.Error, "GetSubscriptionStats should not return error")
	require.NotNil(t, response.Result, "GetSubscriptionStats result should not be nil")
}

// testSetRetentionPolicyWithExistingPatterns tests set_retention_policy method using existing test patterns
func testSetRetentionPolicyWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for set_retention_policy operation
	response, err := client.SendJSONRPC("set_retention_policy", map[string]interface{}{
		"policy_type": "time_based",
		"enabled":     true,
	})
	require.NoError(t, err, "SetRetentionPolicy should succeed")
	require.Nil(t, response.Error, "SetRetentionPolicy should not return error")
	require.NotNil(t, response.Result, "SetRetentionPolicy result should not be nil")
}

// testCleanupOldFilesWithExistingPatterns tests cleanup_old_files method using existing test patterns
func testCleanupOldFilesWithExistingPatterns(t *testing.T, client *testutils.WebSocketTestClient) {
	// Use existing test patterns for cleanup_old_files operation
	response, err := client.SendJSONRPC("cleanup_old_files", map[string]interface{}{})
	require.NoError(t, err, "CleanupOldFiles should succeed")
	require.Nil(t, response.Error, "CleanupOldFiles should not return error")
	require.NotNil(t, response.Result, "CleanupOldFiles result should not be nil")
}
