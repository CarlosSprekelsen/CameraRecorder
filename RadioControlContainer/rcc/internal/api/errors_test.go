package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
)

func TestToAPIError_NilError(t *testing.T) {
	status, body := ToAPIError(nil)

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	if body != nil {
		t.Errorf("Expected nil body, got %v", body)
	}
}

func TestToAPIError_AdapterErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "INVALID_RANGE",
			err:            adapter.ErrInvalidRange,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RANGE",
		},
		{
			name:           "BUSY",
			err:            adapter.ErrBusy,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "BUSY",
		},
		{
			name:           "UNAVAILABLE",
			err:            adapter.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "UNAVAILABLE",
		},
		{
			name:           "INTERNAL",
			err:            adapter.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := ToAPIError(tt.err)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}

			// Parse response body
			var response Response
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}

			if response.CorrelationID == "" {
				t.Error("Expected correlation ID to be set")
			}
		})
	}
}

func TestToAPIError_VendorError(t *testing.T) {
	originalErr := errors.New("vendor specific error")
	vendorErr := &adapter.VendorError{
		Code:     adapter.ErrInvalidRange,
		Original: originalErr,
		Details:  map[string]string{"vendor": "test"},
	}

	status, body := ToAPIError(vendorErr)

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Parse response body
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}

	if response.Code != "INVALID_RANGE" {
		t.Errorf("Expected code 'INVALID_RANGE', got '%s'", response.Code)
	}

	if response.Details == nil {
		t.Error("Expected details to be preserved")
	}

	if response.CorrelationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}

func TestToAPIError_APIErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "UNAUTHORIZED",
			err:            ErrUnauthorizedError,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "FORBIDDEN",
			err:            ErrForbiddenError,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "NOT_FOUND",
			err:            ErrNotFoundError,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := ToAPIError(tt.err)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}

			// Parse response body
			var response Response
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != "error" {
				t.Errorf("Expected result 'error', got '%s'", response.Result)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}

			if response.CorrelationID == "" {
				t.Error("Expected correlation ID to be set")
			}
		})
	}
}

func TestToAPIError_UnknownError(t *testing.T) {
	unknownErr := errors.New("unknown error")

	status, body := ToAPIError(unknownErr)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, status)
	}

	// Parse response body
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}

	if response.Code != "INTERNAL" {
		t.Errorf("Expected code 'INTERNAL', got '%s'", response.Code)
	}

	if response.CorrelationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}

func TestToAPIError_APILayerError(t *testing.T) {
	apiErr := &APIError{
		Code:       "CUSTOM_ERROR",
		Message:    "Custom error message",
		Details:    map[string]string{"key": "value"},
		StatusCode: http.StatusBadRequest,
	}

	status, body := ToAPIError(apiErr)

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Parse response body
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}

	if response.Code != "CUSTOM_ERROR" {
		t.Errorf("Expected code 'CUSTOM_ERROR', got '%s'", response.Code)
	}

	if response.Message != "Custom error message" {
		t.Errorf("Expected message 'Custom error message', got '%s'", response.Message)
	}

	if response.CorrelationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}

func TestMapAdapterError(t *testing.T) {
	tests := []struct {
		name           string
		adapterErr     error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "INVALID_RANGE",
			adapterErr:     adapter.ErrInvalidRange,
			expectedCode:   "INVALID_RANGE",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "BUSY",
			adapterErr:     adapter.ErrBusy,
			expectedCode:   "BUSY",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "UNAVAILABLE",
			adapterErr:     adapter.ErrUnavailable,
			expectedCode:   "UNAVAILABLE",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "INTERNAL",
			adapterErr:     adapter.ErrInternal,
			expectedCode:   "INTERNAL",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, status := mapAdapterError(tt.adapterErr)

			if code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, code)
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     error
		original error
		expected string
	}{
		{
			name:     "INVALID_RANGE",
			code:     adapter.ErrInvalidRange,
			original: nil,
			expected: "Parameter value is outside the allowed range",
		},
		{
			name:     "BUSY",
			code:     adapter.ErrBusy,
			original: nil,
			expected: "Service is busy, please retry with backoff",
		},
		{
			name:     "UNAVAILABLE",
			code:     adapter.ErrUnavailable,
			original: nil,
			expected: "Service is temporarily unavailable",
		},
		{
			name:     "INTERNAL",
			code:     adapter.ErrInternal,
			original: nil,
			expected: "Internal server error",
		},
		{
			name:     "UNAUTHORIZED",
			code:     ErrUnauthorizedError,
			original: nil,
			expected: "Authentication required",
		},
		{
			name:     "FORBIDDEN",
			code:     ErrForbiddenError,
			original: nil,
			expected: "Insufficient permissions",
		},
		{
			name:     "NOT_FOUND",
			code:     ErrNotFoundError,
			original: nil,
			expected: "Resource not found",
		},
		{
			name:     "Unknown with original",
			code:     errors.New("UNKNOWN"),
			original: errors.New("original error"),
			expected: "original error",
		},
		{
			name:     "Unknown without original",
			code:     errors.New("UNKNOWN"),
			original: nil,
			expected: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := getErrorMessage(tt.code, tt.original)

			if message != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, message)
			}
		})
	}
}

func TestMarshalErrorResponse(t *testing.T) {
	code := "TEST_ERROR"
	message := "Test error message"
	details := map[string]string{"key": "value"}

	body := marshalErrorResponse(code, message, details)

	// Parse response body
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}

	if response.Code != code {
		t.Errorf("Expected code '%s', got '%s'", code, response.Code)
	}

	if response.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, response.Message)
	}

	if response.CorrelationID == "" {
		t.Error("Expected correlation ID to be set")
	}
}

func TestNewAPIError(t *testing.T) {
	apiErr := NewAPIError("TEST_ERROR", "Test message", http.StatusBadRequest, map[string]string{"key": "value"})

	if apiErr.Code != "TEST_ERROR" {
		t.Errorf("Expected code 'TEST_ERROR', got '%s'", apiErr.Code)
	}

	if apiErr.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", apiErr.Message)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	// Test Error() method
	errorString := apiErr.Error()
	expected := "TEST_ERROR: Test message"
	if errorString != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errorString)
	}
}
