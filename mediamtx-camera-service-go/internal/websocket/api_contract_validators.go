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

	// Validate status values
	validStatuses := []string{"STOPPED", "FAILED"}
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

	// Validate status values
	validStatuses := []string{"RECORDING", "FAILED"}
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

	// Validate status values
	validStatuses := []string{"success", "failed"}
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
