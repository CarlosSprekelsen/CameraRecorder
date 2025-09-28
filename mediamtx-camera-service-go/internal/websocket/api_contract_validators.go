// Package websocket implements API contract validation helpers
//
// This file contains validation helpers that ensure API responses match
// the documented contracts in docs/api/json_rpc_methods.md
//
// These validators enforce strict API compliance and prevent accommodation
// of incorrect responses in tests.

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateStopRecordingResponse validates stop_recording API response structure
func validateStopRecordingResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "StopRecording result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "StopRecording result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "device", "StopRecording must include device field")
	assert.Contains(t, resultMap, "filename", "StopRecording must include filename field")
	assert.Contains(t, resultMap, "status", "StopRecording must include status field")
	assert.Contains(t, resultMap, "start_time", "StopRecording must include start_time field")
	assert.Contains(t, resultMap, "end_time", "StopRecording must include end_time field")
	assert.Contains(t, resultMap, "duration", "StopRecording must include duration field")
	assert.Contains(t, resultMap, "file_size", "StopRecording must include file_size field")
	assert.Contains(t, resultMap, "format", "StopRecording must include format field")

	// Validate field types
	assert.IsType(t, "", resultMap["device"], "device must be string")
	assert.IsType(t, "", resultMap["filename"], "filename must be string")
	assert.IsType(t, "", resultMap["status"], "status must be string")
	assert.IsType(t, "", resultMap["format"], "format must be string")

	// Validate status values (API spec compliance - STOPPED for stop_recording)
	validStatuses := []string{"STOPPED", "STARTING", "STOPPING", "PAUSED", "ERROR", "FAILED"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")
}

// validateStartRecordingResponse validates start_recording API response structure
func validateStartRecordingResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "StartRecording result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "StartRecording result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "device", "StartRecording must include device field")
	assert.Contains(t, resultMap, "filename", "StartRecording must include filename field")
	assert.Contains(t, resultMap, "status", "StartRecording must include status field")
	assert.Contains(t, resultMap, "start_time", "StartRecording must include start_time field")
	assert.Contains(t, resultMap, "format", "StartRecording must include format field")

	// Validate status values (API spec compliance - RECORDING for start_recording)
	validStatuses := []string{"RECORDING", "STARTING", "STOPPING", "PAUSED", "ERROR", "FAILED"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")
}

// validateTakeSnapshotResponse validates take_snapshot API response structure
func validateTakeSnapshotResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "TakeSnapshot result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "TakeSnapshot result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "device", "TakeSnapshot must include device field")
	assert.Contains(t, resultMap, "filename", "TakeSnapshot must include filename field")
	assert.Contains(t, resultMap, "status", "TakeSnapshot must include status field")
	assert.Contains(t, resultMap, "timestamp", "TakeSnapshot must include timestamp field")
	assert.Contains(t, resultMap, "file_size", "TakeSnapshot must include file_size field")
	assert.Contains(t, resultMap, "file_path", "TakeSnapshot must include file_path field")

	// Validate status values (API spec compliance - uppercase as per documentation)
	validStatuses := []string{"SUCCESS", "FAILED"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")
}

// validateRecordingSpecificError validates that recording method errors are recording-related
func validateRecordingSpecificError(t *testing.T, errorCode int, method string) {
	validRecordingErrors := []int{
		CAMERA_NOT_FOUND,
		RECORDING_IN_PROGRESS,
		ERROR_CAMERA_NOT_FOUND,
		ERROR_CAMERA_NOT_AVAILABLE,
		ERROR_RECORDING_IN_PROGRESS,
		ERROR_MEDIAMTX_ERROR,
		INSUFFICIENT_STORAGE,
	}

	assert.Contains(t, validRecordingErrors, errorCode,
		"Method %s should return recording-specific errors, not system errors. Got error code: %d",
		method, errorCode)
}

// validatePingResponse validates ping API response structure using existing test patterns
func validatePingResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "Ping result cannot be nil")
	
	// Ping should return "pong" string
	assert.Equal(t, "pong", result, "Ping should return 'pong'")
}

// validateGetRecordingInfoResponse validates get_recording_info API response structure
func validateGetRecordingInfoResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetRecordingInfo result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetRecordingInfo result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "filename", "GetRecordingInfo must include filename field")
	assert.Contains(t, resultMap, "file_size", "GetRecordingInfo must include file_size field")
	assert.Contains(t, resultMap, "duration", "GetRecordingInfo must include duration field")
	assert.Contains(t, resultMap, "created_time", "GetRecordingInfo must include created_time field")
	assert.Contains(t, resultMap, "download_url", "GetRecordingInfo must include download_url field")

	// Validate field types
	assert.IsType(t, "", resultMap["filename"], "filename must be string")
	assert.IsType(t, float64(0), resultMap["file_size"], "file_size must be number")
	assert.IsType(t, float64(0), resultMap["duration"], "duration must be number")
	assert.IsType(t, "", resultMap["created_time"], "created_time must be string")
	assert.IsType(t, "", resultMap["download_url"], "download_url must be string")
}

// validateGetSnapshotInfoResponse validates get_snapshot_info API response structure
func validateGetSnapshotInfoResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetSnapshotInfo result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetSnapshotInfo result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "filename", "GetSnapshotInfo must include filename field")
	assert.Contains(t, resultMap, "file_size", "GetSnapshotInfo must include file_size field")
	assert.Contains(t, resultMap, "created_time", "GetSnapshotInfo must include created_time field")
	assert.Contains(t, resultMap, "download_url", "GetSnapshotInfo must include download_url field")

	// Validate field types
	assert.IsType(t, "", resultMap["filename"], "filename must be string")
	assert.IsType(t, float64(0), resultMap["file_size"], "file_size must be number")
	assert.IsType(t, "", resultMap["created_time"], "created_time must be string")
	assert.IsType(t, "", resultMap["download_url"], "download_url must be string")
}

// validateGetStatusResponse validates get_status API response structure
func validateGetStatusResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetStatus result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetStatus result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "status", "GetStatus must include status field")
	assert.Contains(t, resultMap, "uptime", "GetStatus must include uptime field")
	assert.Contains(t, resultMap, "version", "GetStatus must include version field")
	assert.Contains(t, resultMap, "components", "GetStatus must include components field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"HEALTHY", "DEGRADED", "UNHEALTHY"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")

	// Validate components
	if components, ok := resultMap["components"].(map[string]interface{}); ok {
		validComponentStates := []string{"RUNNING", "STOPPED", "ERROR", "STARTING", "STOPPING"}
		for component, state := range components {
			assert.Contains(t, validComponentStates, state,
				"Component %s state must be valid", component)
		}
	}
}

// validateGetSystemStatusResponse validates get_system_status API response structure
func validateGetSystemStatusResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetSystemStatus result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetSystemStatus result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "status", "GetSystemStatus must include status field")
	assert.Contains(t, resultMap, "message", "GetSystemStatus must include message field")
	assert.Contains(t, resultMap, "available_cameras", "GetSystemStatus must include available_cameras field")
	assert.Contains(t, resultMap, "discovery_active", "GetSystemStatus must include discovery_active field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"starting", "partial", "ready"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, "", resultMap["status"], "status must be string")
	assert.IsType(t, "", resultMap["message"], "message must be string")
	assert.IsType(t, []interface{}{}, resultMap["available_cameras"], "available_cameras must be array")
	assert.IsType(t, true, resultMap["discovery_active"], "discovery_active must be boolean")
}

// validateGetServerInfoResponse validates get_server_info API response structure
func validateGetServerInfoResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetServerInfo result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetServerInfo result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "name", "GetServerInfo must include name field")
	assert.Contains(t, resultMap, "version", "GetServerInfo must include version field")
	assert.Contains(t, resultMap, "build_date", "GetServerInfo must include build_date field")
	assert.Contains(t, resultMap, "go_version", "GetServerInfo must include go_version field")
	assert.Contains(t, resultMap, "architecture", "GetServerInfo must include architecture field")
	assert.Contains(t, resultMap, "capabilities", "GetServerInfo must include capabilities field")
	assert.Contains(t, resultMap, "supported_formats", "GetServerInfo must include supported_formats field")
	assert.Contains(t, resultMap, "max_cameras", "GetServerInfo must include max_cameras field")

	// Validate field types
	assert.IsType(t, "", resultMap["name"], "name must be string")
	assert.IsType(t, "", resultMap["version"], "version must be string")
	assert.IsType(t, "", resultMap["build_date"], "build_date must be string")
	assert.IsType(t, "", resultMap["go_version"], "go_version must be string")
	assert.IsType(t, "", resultMap["architecture"], "architecture must be string")
	assert.IsType(t, []interface{}{}, resultMap["capabilities"], "capabilities must be array")
	assert.IsType(t, []interface{}{}, resultMap["supported_formats"], "supported_formats must be array")
	assert.IsType(t, float64(0), resultMap["max_cameras"], "max_cameras must be number")
}

// validateGetStorageInfoResponse validates get_storage_info API response structure
func validateGetStorageInfoResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetStorageInfo result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetStorageInfo result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "total_space", "GetStorageInfo must include total_space field")
	assert.Contains(t, resultMap, "used_space", "GetStorageInfo must include used_space field")
	assert.Contains(t, resultMap, "available_space", "GetStorageInfo must include available_space field")
	assert.Contains(t, resultMap, "usage_percentage", "GetStorageInfo must include usage_percentage field")
	assert.Contains(t, resultMap, "recordings_size", "GetStorageInfo must include recordings_size field")
	assert.Contains(t, resultMap, "snapshots_size", "GetStorageInfo must include snapshots_size field")
	assert.Contains(t, resultMap, "low_space_warning", "GetStorageInfo must include low_space_warning field")

	// Validate field types
	assert.IsType(t, float64(0), resultMap["total_space"], "total_space must be number")
	assert.IsType(t, float64(0), resultMap["used_space"], "used_space must be number")
	assert.IsType(t, float64(0), resultMap["available_space"], "available_space must be number")
	assert.IsType(t, float64(0), resultMap["usage_percentage"], "usage_percentage must be number")
	assert.IsType(t, float64(0), resultMap["recordings_size"], "recordings_size must be number")
	assert.IsType(t, float64(0), resultMap["snapshots_size"], "snapshots_size must be number")
	assert.IsType(t, true, resultMap["low_space_warning"], "low_space_warning must be boolean")
}

// validateSetRetentionPolicyResponse validates set_retention_policy API response structure
func validateSetRetentionPolicyResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "SetRetentionPolicy result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "SetRetentionPolicy result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "policy_type", "SetRetentionPolicy must include policy_type field")
	assert.Contains(t, resultMap, "enabled", "SetRetentionPolicy must include enabled field")
	assert.Contains(t, resultMap, "message", "SetRetentionPolicy must include message field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"UPDATED", "ERROR"}
	if status, exists := resultMap["status"]; exists {
		assert.Contains(t, validStatuses, status, "status must be valid value")
	}

	// Validate field types
	assert.IsType(t, "", resultMap["policy_type"], "policy_type must be string")
	assert.IsType(t, true, resultMap["enabled"], "enabled must be boolean")
	assert.IsType(t, "", resultMap["message"], "message must be string")
}

// validateCleanupOldFilesResponse validates cleanup_old_files API response structure
func validateCleanupOldFilesResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "CleanupOldFiles result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "CleanupOldFiles result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "cleanup_executed", "CleanupOldFiles must include cleanup_executed field")
	assert.Contains(t, resultMap, "files_deleted", "CleanupOldFiles must include files_deleted field")
	assert.Contains(t, resultMap, "space_freed", "CleanupOldFiles must include space_freed field")
	assert.Contains(t, resultMap, "message", "CleanupOldFiles must include message field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"SUCCESS", "FAILED"}
	if status, exists := resultMap["status"]; exists {
		assert.Contains(t, validStatuses, status, "status must be valid value")
	}

	// Validate field types
	assert.IsType(t, true, resultMap["cleanup_executed"], "cleanup_executed must be boolean")
	assert.IsType(t, float64(0), resultMap["files_deleted"], "files_deleted must be number")
	assert.IsType(t, float64(0), resultMap["space_freed"], "space_freed must be number")
	assert.IsType(t, "", resultMap["message"], "message must be string")
}

// validateDiscoverExternalStreamsResponse validates discover_external_streams API response structure
func validateDiscoverExternalStreamsResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "DiscoverExternalStreams result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "DiscoverExternalStreams result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "discovered_streams", "DiscoverExternalStreams must include discovered_streams field")
	assert.Contains(t, resultMap, "skydio_streams", "DiscoverExternalStreams must include skydio_streams field")
	assert.Contains(t, resultMap, "generic_streams", "DiscoverExternalStreams must include generic_streams field")
	assert.Contains(t, resultMap, "scan_timestamp", "DiscoverExternalStreams must include scan_timestamp field")
	assert.Contains(t, resultMap, "total_found", "DiscoverExternalStreams must include total_found field")

	// Validate field types
	assert.IsType(t, []interface{}{}, resultMap["discovered_streams"], "discovered_streams must be array")
	assert.IsType(t, []interface{}{}, resultMap["skydio_streams"], "skydio_streams must be array")
	assert.IsType(t, []interface{}{}, resultMap["generic_streams"], "generic_streams must be array")
	assert.IsType(t, "", resultMap["scan_timestamp"], "scan_timestamp must be string")
	assert.IsType(t, float64(0), resultMap["total_found"], "total_found must be number")
}

// validateAddExternalStreamResponse validates add_external_stream API response structure
func validateAddExternalStreamResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "AddExternalStream result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "AddExternalStream result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "stream_url", "AddExternalStream must include stream_url field")
	assert.Contains(t, resultMap, "stream_name", "AddExternalStream must include stream_name field")
	assert.Contains(t, resultMap, "stream_type", "AddExternalStream must include stream_type field")
	assert.Contains(t, resultMap, "status", "AddExternalStream must include status field")
	assert.Contains(t, resultMap, "timestamp", "AddExternalStream must include timestamp field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"ADDED", "ERROR"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, "", resultMap["stream_url"], "stream_url must be string")
	assert.IsType(t, "", resultMap["stream_name"], "stream_name must be string")
	assert.IsType(t, "", resultMap["stream_type"], "stream_type must be string")
	assert.IsType(t, "", resultMap["status"], "status must be string")
	assert.IsType(t, "", resultMap["timestamp"], "timestamp must be string")
}

// validateRemoveExternalStreamResponse validates remove_external_stream API response structure
func validateRemoveExternalStreamResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "RemoveExternalStream result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "RemoveExternalStream result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "stream_url", "RemoveExternalStream must include stream_url field")
	assert.Contains(t, resultMap, "status", "RemoveExternalStream must include status field")
	assert.Contains(t, resultMap, "timestamp", "RemoveExternalStream must include timestamp field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"REMOVED", "ERROR"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, "", resultMap["stream_url"], "stream_url must be string")
	assert.IsType(t, "", resultMap["status"], "status must be string")
	assert.IsType(t, "", resultMap["timestamp"], "timestamp must be string")
}

// validateGetExternalStreamsResponse validates get_external_streams API response structure
func validateGetExternalStreamsResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "GetExternalStreams result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "GetExternalStreams result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "external_streams", "GetExternalStreams must include external_streams field")
	assert.Contains(t, resultMap, "skydio_streams", "GetExternalStreams must include skydio_streams field")
	assert.Contains(t, resultMap, "generic_streams", "GetExternalStreams must include generic_streams field")
	assert.Contains(t, resultMap, "total_count", "GetExternalStreams must include total_count field")
	assert.Contains(t, resultMap, "timestamp", "GetExternalStreams must include timestamp field")

	// Validate field types
	assert.IsType(t, []interface{}{}, resultMap["external_streams"], "external_streams must be array")
	assert.IsType(t, []interface{}{}, resultMap["skydio_streams"], "skydio_streams must be array")
	assert.IsType(t, []interface{}{}, resultMap["generic_streams"], "generic_streams must be array")
	assert.IsType(t, float64(0), resultMap["total_count"], "total_count must be number")
	assert.IsType(t, "", resultMap["timestamp"], "timestamp must be string")
}

// validateSetDiscoveryIntervalResponse validates set_discovery_interval API response structure
func validateSetDiscoveryIntervalResponse(t *testing.T, result interface{}) {
	require.NotNil(t, result, "SetDiscoveryInterval result cannot be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "SetDiscoveryInterval result must be object")

	// Required fields per API documentation
	assert.Contains(t, resultMap, "scan_interval", "SetDiscoveryInterval must include scan_interval field")
	assert.Contains(t, resultMap, "status", "SetDiscoveryInterval must include status field")
	assert.Contains(t, resultMap, "message", "SetDiscoveryInterval must include message field")
	assert.Contains(t, resultMap, "timestamp", "SetDiscoveryInterval must include timestamp field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"UPDATED", "ERROR"}
	assert.Contains(t, validStatuses, resultMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, float64(0), resultMap["scan_interval"], "scan_interval must be number")
	assert.IsType(t, "", resultMap["status"], "status must be string")
	assert.IsType(t, "", resultMap["message"], "message must be string")
	assert.IsType(t, "", resultMap["timestamp"], "timestamp must be string")
}

// validateCameraStatusUpdateNotification validates camera_status_update notification structure
func validateCameraStatusUpdateNotification(t *testing.T, notification interface{}) {
	require.NotNil(t, notification, "CameraStatusUpdate notification cannot be nil")

	notifMap, ok := notification.(map[string]interface{})
	require.True(t, ok, "CameraStatusUpdate notification must be object")

	// Required fields per API documentation
	assert.Contains(t, notifMap, "device", "CameraStatusUpdate must include device field")
	assert.Contains(t, notifMap, "status", "CameraStatusUpdate must include status field")
	assert.Contains(t, notifMap, "name", "CameraStatusUpdate must include name field")
	assert.Contains(t, notifMap, "resolution", "CameraStatusUpdate must include resolution field")
	assert.Contains(t, notifMap, "fps", "CameraStatusUpdate must include fps field")
	assert.Contains(t, notifMap, "streams", "CameraStatusUpdate must include streams field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"CONNECTED", "DISCONNECTED", "ERROR"}
	assert.Contains(t, validStatuses, notifMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, "", notifMap["device"], "device must be string")
	assert.IsType(t, "", notifMap["status"], "status must be string")
	assert.IsType(t, "", notifMap["name"], "name must be string")
	assert.IsType(t, "", notifMap["resolution"], "resolution must be string")
	assert.IsType(t, float64(0), notifMap["fps"], "fps must be number")
	assert.IsType(t, map[string]interface{}{}, notifMap["streams"], "streams must be object")
}

// validateRecordingStatusUpdateNotification validates recording_status_update notification structure
func validateRecordingStatusUpdateNotification(t *testing.T, notification interface{}) {
	require.NotNil(t, notification, "RecordingStatusUpdate notification cannot be nil")

	notifMap, ok := notification.(map[string]interface{})
	require.True(t, ok, "RecordingStatusUpdate notification must be object")

	// Required fields per API documentation
	assert.Contains(t, notifMap, "device", "RecordingStatusUpdate must include device field")
	assert.Contains(t, notifMap, "status", "RecordingStatusUpdate must include status field")
	assert.Contains(t, notifMap, "filename", "RecordingStatusUpdate must include filename field")
	assert.Contains(t, notifMap, "duration", "RecordingStatusUpdate must include duration field")

	// Validate status values (API spec compliance)
	validStatuses := []string{"STARTED", "STOPPED", "ERROR"}
	assert.Contains(t, validStatuses, notifMap["status"], "status must be valid value")

	// Validate field types
	assert.IsType(t, "", notifMap["device"], "device must be string")
	assert.IsType(t, "", notifMap["status"], "status must be string")
	assert.IsType(t, "", notifMap["filename"], "filename must be string")
	assert.IsType(t, float64(0), notifMap["duration"], "duration must be number")
}

// validateAPICompliantError validates error follows JSON-RPC 2.0 and API specification
func validateAPICompliantError(t *testing.T, err *JsonRpcError) {
	require.NotNil(t, err, "Error cannot be nil")

	// Validate error code is defined in API specification
	validErrorCodes := []int{
		// Standard JSON-RPC 2.0 errors
		-32600, -32601, -32602, -32603,
		// Service-specific errors
		AUTHENTICATION_REQUIRED, RATE_LIMIT_EXCEEDED, INSUFFICIENT_PERMISSIONS,
		CAMERA_NOT_FOUND, RECORDING_IN_PROGRESS, MEDIAMTX_UNAVAILABLE,
		INSUFFICIENT_STORAGE, CAPABILITY_NOT_SUPPORTED,
		// Enhanced recording errors
		ERROR_CAMERA_NOT_FOUND, ERROR_CAMERA_NOT_AVAILABLE,
		ERROR_RECORDING_IN_PROGRESS, ERROR_MEDIAMTX_ERROR,
	}

	assert.Contains(t, validErrorCodes, err.Code,
		"Error code %d is not defined in API specification", err.Code)

	// Validate error message is not empty
	assert.NotEmpty(t, err.Message, "Error message cannot be empty")

	// Validate error data exists
	assert.NotNil(t, err.Data, "Error data should exist")
}
