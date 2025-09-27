/*
MediaMTX Error Handling

This package provides comprehensive error handling for the MediaMTX camera service.
It includes structured error types, error wrapping, and context preservation.

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring
*/

package mediamtx

import (
	"encoding/json"
	"fmt"
	"time"
)

// MediaMTXError represents MediaMTX-specific errors with structured information
type MediaMTXError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
	Op      string `json:"op,omitempty"`
	Time    string `json:"time"`
}

func (e *MediaMTXError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("MediaMTX error [%s]: %s (code: %d)", e.Op, e.Message, e.Code)
	}
	return fmt.Sprintf("MediaMTX error: %s (code: %d)", e.Message, e.Code)
}

// Unwrap returns the underlying error if it exists
func (e *MediaMTXError) Unwrap() error {
	return nil
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *MediaMTXError) Is(target error) bool {
	if targetErr, ok := target.(*MediaMTXError); ok {
		return e.Code == targetErr.Code && e.Message == targetErr.Message
	}
	return false
}

// MarshalJSON implements custom JSON marshaling
func (e *MediaMTXError) MarshalJSON() ([]byte, error) {
	type Alias MediaMTXError
	return json.Marshal(&struct {
		*Alias
		Time string `json:"time"`
	}{
		Alias: (*Alias)(e),
		Time:  time.Now().Format(time.RFC3339),
	})
}

// PathError represents path-specific errors
type PathError struct {
	PathName string `json:"path_name"`
	Op       string `json:"op"`
	Message  string `json:"message"`
	Err      error  `json:"-"`
}

func (e *PathError) Error() string {
	return fmt.Sprintf("path %s: %s: %s", e.PathName, e.Op, e.Message)
}

func (e *PathError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *PathError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*PathError); ok {
		// Compare by path name and operation for exact matches
		return e.PathName == targetErr.PathName && e.Op == targetErr.Op
	}

	// Check if this error wraps the target error
	if e.Err != nil {
		return false // We don't implement errors.Is for wrapped errors in this context
	}

	return false
}

// NewMediaMTXError creates a new MediaMTX error
func NewMediaMTXError(code int, message, details string) *MediaMTXError {
	return &MediaMTXError{
		Code:    code,
		Message: message,
		Details: details,
		Time:    time.Now().Format(time.RFC3339),
	}
}

// NewMediaMTXErrorWithOp creates a new MediaMTX error with operation context
func NewMediaMTXErrorWithOp(code int, message, details, op string) *MediaMTXError {
	return &MediaMTXError{
		Code:    code,
		Message: message,
		Details: details,
		Op:      op,
		Time:    time.Now().Format(time.RFC3339),
	}
}

// NewMediaMTXErrorFromHTTP creates a MediaMTX error from HTTP response
func NewMediaMTXErrorFromHTTP(statusCode int, body []byte) *MediaMTXError {
	message := "unknown error"
	details := string(body)

	// Parse structured error from swagger.json Error schema
	var errorResponse struct {
		Error string `json:"error"`
	}
	if len(body) > 0 {
		json.Unmarshal(body, &errorResponse)
		if errorResponse.Error != "" {
			details = errorResponse.Error
		}
	}

	switch statusCode {
	case 400:
		message = "bad request"
	case 401:
		message = "unauthorized access"
	case 403:
		message = "forbidden access"
	case 404:
		message = "resource not found"
	case 409:
		message = "resource conflict"
	case 422:
		message = "validation error"
	case 500:
		message = "internal server error"
	case 502:
		message = "bad gateway"
	case 503:
		message = "service unavailable"
	case 504:
		message = "gateway timeout"
	default:
		message = fmt.Sprintf("unexpected status code: %d", statusCode)
	}

	return &MediaMTXError{
		Code:    statusCode,
		Message: message,
		Details: details,
		Time:    time.Now().Format(time.RFC3339),
	}
}

// NewPathError creates a new path error
func NewPathError(pathName, op, message string) *PathError {
	return &PathError{
		PathName: pathName,
		Op:       op,
		Message:  message,
	}
}

// NewPathErrorWithErr creates a new path error with underlying error
func NewPathErrorWithErr(pathName, op, message string, err error) *PathError {
	return &PathError{
		PathName: pathName,
		Op:       op,
		Message:  message,
		Err:      err,
	}
}

// ExternalDiscoveryDisabledError represents an error when external discovery is disabled
type ExternalDiscoveryDisabledError struct {
	Message string
}

// Error implements the error interface
func (e *ExternalDiscoveryDisabledError) Error() string {
	return e.Message
}
