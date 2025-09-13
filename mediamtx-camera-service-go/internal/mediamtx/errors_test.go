/*
MediaMTX Errors Unit Tests

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMediaMTXError_ReqMTX007 tests MediaMTX error creation
func TestNewMediaMTXError_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	code := 404
	message := "Resource not found"
	details := "The requested path does not exist"

	err := NewMediaMTXError(code, message, details)
	require.NotNil(t, err, "Error should not be nil")

	assert.Equal(t, code, err.Code, "Error code should match")
	assert.Equal(t, message, err.Message, "Error message should match")
	assert.Equal(t, details, err.Details, "Error details should match")
}

// TestNewMediaMTXErrorWithOp_ReqMTX007 tests MediaMTX error creation with operation
func TestNewMediaMTXErrorWithOp_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	code := 500
	message := "Internal server error"
	details := "Database connection failed"
	op := "CreatePath"

	err := NewMediaMTXErrorWithOp(code, message, details, op)
	require.NotNil(t, err, "Error should not be nil")

	assert.Equal(t, code, err.Code, "Error code should match")
	assert.Equal(t, message, err.Message, "Error message should match")
	assert.Equal(t, details, err.Details, "Error details should match")
	assert.Equal(t, op, err.Op, "Operation should match")
}

// TestNewMediaMTXErrorFromHTTP_ReqMTX007 tests HTTP error creation
func TestNewMediaMTXErrorFromHTTP_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	testCases := []struct {
		statusCode      int
		body            []byte
		expectedMessage string
	}{
		{http.StatusUnauthorized, []byte("Unauthorized"), "unauthorized access"},
		{http.StatusForbidden, []byte("Forbidden"), "forbidden access"},
		{http.StatusNotFound, []byte("Not Found"), "resource not found"},
		{http.StatusConflict, []byte("Conflict"), "resource conflict"},
		{http.StatusInternalServerError, []byte("Internal Error"), "internal server error"},
		{http.StatusBadGateway, []byte("Bad Gateway"), "bad gateway"},
		{http.StatusServiceUnavailable, []byte("Service Unavailable"), "service unavailable"},
		{http.StatusGatewayTimeout, []byte("Gateway Timeout"), "gateway timeout"},
		{999, []byte("Unknown"), "unknown error"},
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.statusCode), func(t *testing.T) {
			err := NewMediaMTXErrorFromHTTP(tc.statusCode, tc.body)
			require.NotNil(t, err, "Error should not be nil")

			assert.Equal(t, tc.statusCode, err.Code, "Status code should match")
			assert.Equal(t, tc.expectedMessage, err.Message, "Message should match")
			assert.Equal(t, string(tc.body), err.Details, "Body should be in details")
		})
	}
}

// TestNewStreamError_ReqMTX007 - REMOVED: StreamError type was removed during dead code sweep

// TestNewPathError_ReqMTX007 tests path error creation
func TestNewPathError_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	pathName := "test_path"
	op := "CreatePath"
	message := "Failed to create path"

	err := NewPathError(pathName, op, message)
	require.NotNil(t, err, "Error should not be nil")

	assert.Equal(t, pathName, err.PathName, "Path name should match")
	assert.Equal(t, op, err.Op, "Operation should match")
	assert.Equal(t, message, err.Message, "Message should match")
}

// TestNewRecordingError_ReqMTX007 - REMOVED: RecordingError type was removed during dead code sweep

// TestNewFFmpegError_ReqMTX007 - REMOVED: FFmpegError type was removed during dead code sweep

// TestError_Error_ReqMTX007 tests error message formatting
func TestError_Error_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	testCases := []struct {
		err      error
		expected string
	}{
		{
			NewMediaMTXError(404, "Not Found", "Resource missing"),
			"MediaMTX error: Not Found (code: 404)",
		},
		{
			NewMediaMTXErrorWithOp(500, "Server Error", "Database failed", "CreatePath"),
			"MediaMTX error [CreatePath]: Server Error (code: 500)",
		},
		// StreamError test case removed - type was deleted during dead code sweep
		{
			NewPathError("path123", "Create", "Failed to create"),
			"path path123: Create: Failed to create",
		},
		// RecordingError and FFmpegError test cases removed - types were deleted during dead code sweep
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.err.Error(), "Error message should match")
		})
	}
}

// TestError_Unwrap_ReqMTX007 tests error unwrapping
func TestError_Unwrap_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Test that errors can be unwrapped properly
	mediaMTXErr := &MediaMTXError{
		Code:    500,
		Message: "Server Error",
		Details: "Database failed",
		Op:      "CreatePath",
	}

	// MediaMTXError no longer wraps errors - it's a simple error type
	unwrapped := errors.Unwrap(mediaMTXErr)
	assert.Nil(t, unwrapped, "MediaMTXError should not wrap other errors")
}

// TestError_Is_ReqMTX007 tests error type checking
func TestError_Is_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Test error type checking with reflection since errors.Is requires Is method implementation
	mediaMTXErr := NewMediaMTXError(404, "Not Found", "Resource missing")

	// Test type using reflection
	assert.Equal(t, "*mediamtx.MediaMTXError", reflect.TypeOf(mediaMTXErr).String(), "Should be MediaMTXError type")

	// Test that it's not a different error type
	assert.NotEqual(t, "*mediamtx.StreamError", reflect.TypeOf(mediaMTXErr).String(), "Should not be StreamError type")
}

// TestError_Is_WithErrorsIs_ReqMTX007 - REMOVED: Error constants were removed during dead code sweep

// TestError_As_ReqMTX007 tests error type assertion
func TestError_As_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Test error type assertion with errors.As
	mediaMTXErr := NewMediaMTXError(500, "Server Error", "Database failed")

	var target *MediaMTXError
	assert.True(t, errors.As(mediaMTXErr, &target), "Should be able to assert as MediaMTXError")
	assert.Equal(t, mediaMTXErr, target, "Target should be set to the error")
}

// TestError_Validation_ReqMTX007 tests error validation
func TestError_Validation_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Test that errors are properly validated
	testCases := []struct {
		name        string
		err         error
		shouldBeNil bool
	}{
		{"Valid MediaMTX Error", NewMediaMTXError(404, "Not Found", "Resource missing"), false},
		// StreamError test case removed - type was deleted during dead code sweep
		{"Valid Path Error", NewPathError("path123", "Create", "Failed to create"), false},
		// RecordingError and FFmpegError test cases removed - types were deleted during dead code sweep
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldBeNil {
				assert.Nil(t, tc.err, "Error should be nil")
			} else {
				assert.NotNil(t, tc.err, "Error should not be nil")
				assert.NotEmpty(t, tc.err.Error(), "Error message should not be empty")
			}
		})
	}
}
