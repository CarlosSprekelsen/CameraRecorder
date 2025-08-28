//go:build unit
// +build unit

/*
MediaMTX Errors Unit Tests

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"errors"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/stretchr/testify/assert"
)

// TestMediaMTXError_ErrorMethod tests MediaMTXError Error method
func TestMediaMTXError_ErrorMethod(t *testing.T) {
	// Test error with details
	originalErr := errors.New("original error")
	mediaMTXErr := &mediamtx.MediaMTXError{
		Code:    500,
		Message: "Internal server error",
		Details: "Database connection failed",
		Op:      "GetStream",
		Err:     originalErr,
	}

	errorString := mediaMTXErr.Error()
	assert.Contains(t, errorString, "MediaMTX error [500]")
	assert.Contains(t, errorString, "Internal server error")
	assert.Contains(t, errorString, "Database connection failed")

	// Test error without details
	mediaMTXErrNoDetails := &mediamtx.MediaMTXError{
		Code:    404,
		Message: "Not found",
		Op:      "GetPath",
		Err:     originalErr,
	}

	errorStringNoDetails := mediaMTXErrNoDetails.Error()
	assert.Contains(t, errorStringNoDetails, "MediaMTX error [404]")
	assert.Contains(t, errorStringNoDetails, "Not found")
	assert.NotContains(t, errorStringNoDetails, "details")
}

// TestMediaMTXError_Unwrap tests MediaMTXError Unwrap method
func TestMediaMTXError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	mediaMTXErr := &mediamtx.MediaMTXError{
		Code:    500,
		Message: "Internal server error",
		Err:     originalErr,
	}

	unwrappedErr := mediaMTXErr.Unwrap()
	assert.Equal(t, originalErr, unwrappedErr)
}

// TestCircuitBreakerError_ErrorMethod tests CircuitBreakerError Error method
func TestCircuitBreakerError_ErrorMethod(t *testing.T) {
	circuitBreakerErr := &mediamtx.CircuitBreakerError{
		State:   "OPEN",
		Message: "Circuit breaker is open",
		Op:      "HealthCheck",
	}

	errorString := circuitBreakerErr.Error()
	assert.Contains(t, errorString, "circuit breaker OPEN")
	assert.Contains(t, errorString, "Circuit breaker is open")
}

// TestStreamError_ErrorMethod tests StreamError Error method
func TestStreamError_ErrorMethod(t *testing.T) {
	originalErr := errors.New("stream not found")
	streamErr := &mediamtx.StreamError{
		StreamID: "test-stream-123",
		Op:       "CreateStream",
		Message:  "Failed to create stream",
		Err:      originalErr,
	}

	errorString := streamErr.Error()
	assert.Contains(t, errorString, "stream test-stream-123")
	assert.Contains(t, errorString, "CreateStream")
	assert.Contains(t, errorString, "Failed to create stream")
}

// TestStreamError_Unwrap tests StreamError Unwrap method
func TestStreamError_Unwrap(t *testing.T) {
	originalErr := errors.New("stream not found")
	streamErr := &mediamtx.StreamError{
		StreamID: "test-stream-123",
		Op:       "CreateStream",
		Message:  "Failed to create stream",
		Err:      originalErr,
	}

	unwrappedErr := streamErr.Unwrap()
	assert.Equal(t, originalErr, unwrappedErr)
}

// TestPathError_ErrorMethod tests PathError Error method
func TestPathError_ErrorMethod(t *testing.T) {
	originalErr := errors.New("path not found")
	pathErr := &mediamtx.PathError{
		PathName: "test-path",
		Op:       "DeletePath",
		Message:  "Failed to delete path",
		Err:      originalErr,
	}

	errorString := pathErr.Error()
	assert.Contains(t, errorString, "path test-path")
	assert.Contains(t, errorString, "DeletePath")
	assert.Contains(t, errorString, "Failed to delete path")
}

// TestPathError_Unwrap tests PathError Unwrap method
func TestPathError_Unwrap(t *testing.T) {
	originalErr := errors.New("path not found")
	pathErr := &mediamtx.PathError{
		PathName: "test-path",
		Op:       "DeletePath",
		Message:  "Failed to delete path",
		Err:      originalErr,
	}

	unwrappedErr := pathErr.Unwrap()
	assert.Equal(t, originalErr, unwrappedErr)
}

// TestRecordingError_ErrorMethod tests RecordingError Error method
func TestRecordingError_ErrorMethod(t *testing.T) {
	originalErr := errors.New("recording failed")
	recordingErr := &mediamtx.RecordingError{
		SessionID: "recording-123",
		Device:    "/dev/video0",
		Op:        "StartRecording",
		Message:   "Failed to start recording",
		Err:       originalErr,
	}

	errorString := recordingErr.Error()
	assert.Contains(t, errorString, "recording recording-123")
	assert.Contains(t, errorString, "device /dev/video0")
	assert.Contains(t, errorString, "StartRecording")
	assert.Contains(t, errorString, "Failed to start recording")
}

// TestRecordingError_Unwrap tests RecordingError Unwrap method
func TestRecordingError_Unwrap(t *testing.T) {
	originalErr := errors.New("recording failed")
	recordingErr := &mediamtx.RecordingError{
		SessionID: "recording-123",
		Device:    "/dev/video0",
		Op:        "StartRecording",
		Message:   "Failed to start recording",
		Err:       originalErr,
	}

	unwrappedErr := recordingErr.Unwrap()
	assert.Equal(t, originalErr, unwrappedErr)
}

// TestErrorWrapping tests error wrapping functionality
func TestErrorWrapping(t *testing.T) {
	// Test wrapping MediaMTXError
	originalErr := errors.New("database error")
	mediaMTXErr := &mediamtx.MediaMTXError{
		Code:    500,
		Message: "Database operation failed",
		Err:     originalErr,
	}

	assert.Equal(t, originalErr, mediaMTXErr.Unwrap())
	assert.Contains(t, mediaMTXErr.Error(), "Database operation failed")
}

// TestErrorUnwrapping tests error unwrapping functionality
func TestErrorUnwrapping(t *testing.T) {
	// Test unwrapping chain
	originalErr := errors.New("original error")
	streamErr := &mediamtx.StreamError{
		StreamID: "test-stream",
		Op:       "CreateStream",
		Message:  "Stream creation failed",
		Err:      originalErr,
	}
	mediaMTXErr := &mediamtx.MediaMTXError{
		Code:    500,
		Message: "Operation failed",
		Err:     streamErr,
	}

	// Unwrap chain
	unwrapped1 := mediaMTXErr.Unwrap()
	assert.Equal(t, streamErr, unwrapped1)

	unwrapped2 := streamErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped2)
}
