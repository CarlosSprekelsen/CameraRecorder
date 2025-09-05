/*
WebSocket Validation Helper Unit Tests

Provides focused unit tests for WebSocket validation functionality,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestValidationHelper_Creation tests validation helper creation
func TestValidationHelper_Creation(t *testing.T) {
	logger := logging.NewLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logrus.New())

	assert.NotNil(t, helper, "Validation helper should be created")
}

// TestValidationHelper_ValidatePaginationParams tests pagination parameter validation
func TestValidationHelper_ValidatePaginationParams(t *testing.T) {
	logger := logging.NewLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logrus.New())

	// Test valid pagination parameters
	validParams := map[string]interface{}{
		"limit":  50,
		"offset": 10,
	}

	result := helper.ValidatePaginationParams(validParams)
	assert.True(t, result.Valid, "Valid pagination params should pass validation")
	assert.Empty(t, result.Errors, "Valid pagination params should have no errors")

	// Test nil parameters (should use defaults)
	result = helper.ValidatePaginationParams(nil)
	assert.True(t, result.Valid, "Nil params should use defaults")
	assert.Empty(t, result.Errors, "Nil params should have no errors")

	// Test empty parameters (should use defaults)
	emptyParams := map[string]interface{}{}
	result = helper.ValidatePaginationParams(emptyParams)
	assert.True(t, result.Valid, "Empty params should use defaults")
	assert.Empty(t, result.Errors, "Empty params should have no errors")
}

// TestValidationHelper_ValidateResult tests validation result structure
func TestValidationHelper_ValidateResult(t *testing.T) {
	// Test validation result creation
	result := NewValidationResult()
	assert.True(t, result.Valid, "New validation result should be valid")
	assert.Empty(t, result.Errors, "New validation result should have no errors")
	assert.Empty(t, result.Warnings, "New validation result should have no warnings")
	assert.NotNil(t, result.Data, "New validation result should have data map")

	// Test adding error
	result.AddError("Test error")
	assert.False(t, result.Valid, "Validation result should be invalid after adding error")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Equal(t, "Test error", result.Errors[0], "Error message should match")

	// Test adding warning
	result.AddWarning("Test warning")
	assert.Len(t, result.Warnings, 1, "Should have one warning")
	assert.Equal(t, "Test warning", result.Warnings[0], "Warning message should match")

	// Test adding data
	result.AddData("test_key", "test_value")
	assert.Equal(t, "test_value", result.Data["test_key"], "Data should be added correctly")

	// Test getting first error
	firstError := result.GetFirstError()
	assert.Equal(t, "Test error", firstError, "First error should be correct")
}
