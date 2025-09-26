/*
WebSocket Missing API Methods Integration Tests

Tests for API methods that are documented in the JSON-RPC specification
but are missing from the current test suite. This ensures complete API coverage.

API Documentation Reference: docs/api/json_rpc_methods.md
Requirements Coverage:
- REQ-API-001: Complete JSON-RPC method coverage
- REQ-API-002: API specification compliance
- REQ-API-003: Missing method validation

Design Principles:
- Real components only (no mocks)
- Fixture-driven configuration
- Progressive Readiness pattern validation
- Complete API specification compliance
- Missing method identification and testing
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MISSING STREAMING METHODS
// ============================================================================

// TestMissingAPI_GetStreamUrl_Integration tests get_stream_url method
func TestMissingAPI_GetStreamUrl_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_stream_url method
	// CRITICAL: JSON-RPC API uses "device" parameter, not "camera_id"
	device := testutils.GetTestCameraID()
	response, err := asserter.client.GetStreamUrl(device)
	require.NoError(t, err, "get_stream_url should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetStreamStatus_Integration tests get_stream_status method
func TestMissingAPI_GetStreamStatus_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_stream_status method
	// CRITICAL: JSON-RPC API uses "device" parameter, not "camera_id"
	device := testutils.GetTestCameraID()
	response, err := asserter.client.GetStreamStatus(device)
	require.NoError(t, err, "get_stream_status should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING SYSTEM MONITORING METHODS
// ============================================================================

// TestMissingAPI_GetMetrics_Integration tests get_metrics method
func TestMissingAPI_GetMetrics_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: get_metrics requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_metrics method
	response, err := asserter.client.GetMetrics()
	require.NoError(t, err, "get_metrics should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetStreams_Integration tests get_streams method
func TestMissingAPI_GetStreams_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_streams method
	response, err := asserter.client.GetStreams()
	require.NoError(t, err, "get_streams should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING SYSTEM STATUS METHODS
// ============================================================================

// TestMissingAPI_GetStatus_Integration tests get_status method
func TestMissingAPI_GetStatus_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: get_status requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_status method
	response, err := asserter.client.GetStatus()
	require.NoError(t, err, "get_status should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetServerInfo_Integration tests get_server_info method
func TestMissingAPI_GetServerInfo_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_server_info method
	response, err := asserter.client.GetServerInfo()
	require.NoError(t, err, "get_server_info should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING STORAGE MANAGEMENT METHODS
// ============================================================================

// TestMissingAPI_GetStorageInfo_Integration tests get_storage_info method
func TestMissingAPI_GetStorageInfo_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: get_storage_info requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_storage_info method
	response, err := asserter.client.GetStorageInfo()
	require.NoError(t, err, "get_storage_info should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_SetRetentionPolicy_Integration tests set_retention_policy method
func TestMissingAPI_SetRetentionPolicy_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: set_retention_policy requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test set_retention_policy method
	// CRITICAL: JSON-RPC API uses specific parameters, not nested "policy" object
	response, err := asserter.client.SetRetentionPolicy("age", 30, 10, true)
	require.NoError(t, err, "set_retention_policy should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_CleanupOldFiles_Integration tests cleanup_old_files method
func TestMissingAPI_CleanupOldFiles_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: cleanup_old_files requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test cleanup_old_files method
	response, err := asserter.client.CleanupOldFiles()
	require.NoError(t, err, "cleanup_old_files should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING FILE INFO METHODS
// ============================================================================

// TestMissingAPI_GetRecordingInfo_Integration tests get_recording_info method
func TestMissingAPI_GetRecordingInfo_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_recording_info method
	filename := "test_recording.mp4"
	response, err := asserter.client.GetRecordingInfo(filename)
	require.NoError(t, err, "get_recording_info should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetSnapshotInfo_Integration tests get_snapshot_info method
func TestMissingAPI_GetSnapshotInfo_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_snapshot_info method
	filename := "test_snapshot.jpg"
	response, err := asserter.client.GetSnapshotInfo(filename)
	require.NoError(t, err, "get_snapshot_info should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING EVENT SUBSCRIPTION METHODS
// ============================================================================

// TestMissingAPI_SubscribeEvents_Integration tests subscribe_events method
func TestMissingAPI_SubscribeEvents_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test subscribe_events method
	eventTypes := []string{"camera_status", "recording_status"}
	response, err := asserter.client.SubscribeEvents(eventTypes)
	require.NoError(t, err, "subscribe_events should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_UnsubscribeEvents_Integration tests unsubscribe_events method
func TestMissingAPI_UnsubscribeEvents_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test unsubscribe_events method
	response, err := asserter.client.UnsubscribeEvents()
	require.NoError(t, err, "unsubscribe_events should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetSubscriptionStats_Integration tests get_subscription_stats method
func TestMissingAPI_GetSubscriptionStats_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_subscription_stats method
	response, err := asserter.client.GetSubscriptionStats()
	require.NoError(t, err, "get_subscription_stats should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// ============================================================================
// MISSING EXTERNAL STREAM METHODS
// ============================================================================

// TestMissingAPI_DiscoverExternalStreams_Integration tests discover_external_streams method
func TestMissingAPI_DiscoverExternalStreams_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test discover_external_streams method
	response, err := asserter.client.DiscoverExternalStreams()
	require.NoError(t, err, "discover_external_streams should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_GetExternalStreams_Integration tests get_external_streams method
func TestMissingAPI_GetExternalStreams_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	authToken, err := asserter.helper.GetJWTToken("operator")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test get_external_streams method
	response, err := asserter.client.GetExternalStreams()
	require.NoError(t, err, "get_external_streams should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}

// TestMissingAPI_SetDiscoveryInterval_Integration tests set_discovery_interval method
func TestMissingAPI_SetDiscoveryInterval_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)

	// Connect and authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "WebSocket connection should succeed")

	// CRITICAL: set_discovery_interval requires admin role per permissions matrix
	authToken, err := asserter.helper.GetJWTToken("admin")
	require.NoError(t, err, "Should be able to create JWT token")

	err = asserter.client.Authenticate(authToken)
	require.NoError(t, err, "Authentication should succeed")

	// Test set_discovery_interval method
	// CRITICAL: JSON-RPC API uses "scan_interval" parameter, not "interval"
	scanInterval := 30 // 30 seconds
	response, err := asserter.client.SetDiscoveryInterval(scanInterval)
	require.NoError(t, err, "set_discovery_interval should succeed")
	require.NotNil(t, response, "Response should not be nil")

	// Validate response structure per API documentation
	asserter.client.AssertJSONRPCResponse(response, false)
}
