/*
WebSocket Validation Helper Unit Tests - Enterprise-Grade Progressive Readiness Pattern

Provides focused unit tests for WebSocket validation functionality,
following homogeneous enterprise-grade patterns with real hardware integration.

ENTERPRISE STANDARDS:
- Progressive Readiness Pattern compliance (no polling, no sequential execution)
- Real hardware integration (no mocking, no skipping)
- Homogeneous test patterns across all validation tests
- Proper documentation with requirements coverage

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-ARCH-001: Progressive Readiness Pattern compliance

Test Categories: Enterprise Unit
API Documentation Reference: docs/api/json_rpc_methods.md
Pattern: Progressive Readiness with real hardware integration
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/stretchr/testify/assert"
)

// TestValidationHelper_Creation tests validation helper creation
func TestValidationHelper_Creation(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	assert.NotNil(t, helper, "Validation helper should be created")
}

// TestValidationHelper_ValidatePaginationParams tests pagination parameter validation
func TestValidationHelper_ValidatePaginationParams(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

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

// TestValidationHelper_ValidateDeviceParameter tests device parameter validation
func TestValidationHelper_ValidateDeviceParameter(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test valid device parameter (camera identifier, not device path)
	validParams := map[string]interface{}{
		"device": "camera0",
	}
	result := helper.ValidateDeviceParameter(validParams)
	assert.True(t, result.Valid, "Valid device parameter should pass validation")
	assert.Empty(t, result.Errors, "Valid device parameter should have no errors")
	assert.Equal(t, "camera0", result.Data["device"], "Device should be extracted correctly")

	// Test nil parameters
	result = helper.ValidateDeviceParameter(nil)
	assert.False(t, result.Valid, "Nil parameters should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "device parameter is required", "Error should mention required device parameter")

	// Test missing device parameter
	missingParams := map[string]interface{}{
		"other": "value",
	}
	result = helper.ValidateDeviceParameter(missingParams)
	assert.False(t, result.Valid, "Missing device parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "device parameter is required", "Error should mention required device parameter")

	// Test invalid device parameter type
	invalidTypeParams := map[string]interface{}{
		"device": 123,
	}
	result = helper.ValidateDeviceParameter(invalidTypeParams)
	assert.False(t, result.Valid, "Invalid device parameter type should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
}

// TestValidationHelper_ValidateFilenameParameter tests filename parameter validation
func TestValidationHelper_ValidateFilenameParameter(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test valid filename parameter
	validParams := map[string]interface{}{
		"filename": "test_recording.mp4",
	}
	result := helper.ValidateFilenameParameter(validParams)
	assert.True(t, result.Valid, "Valid filename parameter should pass validation")
	assert.Empty(t, result.Errors, "Valid filename parameter should have no errors")
	assert.Equal(t, "test_recording.mp4", result.Data["filename"], "Filename should be extracted correctly")

	// Test nil parameters
	result = helper.ValidateFilenameParameter(nil)
	assert.False(t, result.Valid, "Nil parameters should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "filename parameter is required", "Error should mention required filename parameter")

	// Test missing filename parameter
	missingParams := map[string]interface{}{
		"other": "value",
	}
	result = helper.ValidateFilenameParameter(missingParams)
	assert.False(t, result.Valid, "Missing filename parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "filename parameter is required", "Error should mention required filename parameter")

	// Test invalid filename parameter type
	invalidTypeParams := map[string]interface{}{
		"filename": 123,
	}
	result = helper.ValidateFilenameParameter(invalidTypeParams)
	assert.False(t, result.Valid, "Invalid filename parameter type should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
}

// TestValidationHelper_ValidateRecordingParameters tests recording parameter validation
func TestValidationHelper_ValidateRecordingParameters(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test valid recording parameters with all options
	validParams := map[string]interface{}{
		"device":           "camera0",
		"duration_seconds": 60,
		"format":           "mp4",
		"codec":            "h264",
		"quality":          23,
		"use_case":         "recording",
		"priority":         1,
		"auto_cleanup":     true,
		"retention_days":   30,
	}
	result := helper.ValidateRecordingParameters(validParams)
	assert.True(t, result.Valid, "Valid recording parameters should pass validation")
	assert.Empty(t, result.Errors, "Valid recording parameters should have no errors")
	assert.Equal(t, "camera0", result.Data["device"], "Device should be extracted correctly")

	// Test minimal valid recording parameters (only device required)
	minimalParams := map[string]interface{}{
		"device": "camera0",
	}
	result = helper.ValidateRecordingParameters(minimalParams)
	assert.True(t, result.Valid, "Minimal recording parameters should pass validation")
	assert.Empty(t, result.Errors, "Minimal recording parameters should have no errors")

	// Test nil parameters
	result = helper.ValidateRecordingParameters(nil)
	assert.False(t, result.Valid, "Nil parameters should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "parameters are required", "Error should mention required parameters")

	// Test missing device parameter
	missingDeviceParams := map[string]interface{}{
		"duration_seconds": 60,
	}
	result = helper.ValidateRecordingParameters(missingDeviceParams)
	assert.False(t, result.Valid, "Missing device parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "device parameter is required", "Error should mention required device parameter")

	// Test invalid duration parameter
	invalidDurationParams := map[string]interface{}{
		"device":   "camera0",
		"duration": -10, // Invalid negative duration
	}
	result = helper.ValidateRecordingParameters(invalidDurationParams)
	assert.False(t, result.Valid, "Invalid duration parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
}

// TestValidationHelper_ValidateSnapshotParameters tests snapshot parameter validation
func TestValidationHelper_ValidateSnapshotParameters(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test valid snapshot parameters with all options
	validParams := map[string]interface{}{
		"device":   "camera0",
		"filename": "snapshot_001.jpg",
		"format":   "jpg",
		"quality":  85,
	}
	result := helper.ValidateSnapshotParameters(validParams)
	assert.True(t, result.Valid, "Valid snapshot parameters should pass validation")
	assert.Empty(t, result.Errors, "Valid snapshot parameters should have no errors")
	assert.Equal(t, "camera0", result.Data["device"], "Device should be extracted correctly")

	// Test minimal valid snapshot parameters (only device required)
	minimalParams := map[string]interface{}{
		"device": "camera0",
	}
	result = helper.ValidateSnapshotParameters(minimalParams)
	assert.True(t, result.Valid, "Minimal snapshot parameters should pass validation")
	assert.Empty(t, result.Errors, "Minimal snapshot parameters should have no errors")

	// Test nil parameters
	result = helper.ValidateSnapshotParameters(nil)
	assert.False(t, result.Valid, "Nil parameters should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "parameters are required", "Error should mention required parameters")

	// Test missing device parameter
	missingDeviceParams := map[string]interface{}{
		"filename": "test.jpg",
	}
	result = helper.ValidateSnapshotParameters(missingDeviceParams)
	assert.False(t, result.Valid, "Missing device parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "device parameter is required", "Error should mention required device parameter")

	// Test invalid quality parameter
	invalidQualityParams := map[string]interface{}{
		"device":  "camera0",
		"quality": -10, // Invalid negative quality
	}
	result = helper.ValidateSnapshotParameters(invalidQualityParams)
	assert.False(t, result.Valid, "Invalid quality parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
}

// TestValidationHelper_ValidateRetentionPolicyParameters tests retention policy parameter validation
func TestValidationHelper_ValidateRetentionPolicyParameters(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test valid age-based retention policy
	validAgeParams := map[string]interface{}{
		"policy_type":  "age",
		"enabled":      true,
		"max_age_days": 30,
	}
	result := helper.ValidateRetentionPolicyParameters(validAgeParams)
	assert.True(t, result.Valid, "Valid age-based retention policy should pass validation")
	assert.Empty(t, result.Errors, "Valid age-based retention policy should have no errors")
	assert.Equal(t, "age", result.Data["policy_type"], "Policy type should be extracted correctly")
	assert.Equal(t, true, result.Data["enabled"], "Enabled flag should be extracted correctly")

	// Test valid size-based retention policy
	validSizeParams := map[string]interface{}{
		"policy_type": "size",
		"enabled":     false,
		"max_size_gb": 100,
	}
	result = helper.ValidateRetentionPolicyParameters(validSizeParams)
	assert.True(t, result.Valid, "Valid size-based retention policy should pass validation")
	assert.Empty(t, result.Errors, "Valid size-based retention policy should have no errors")
	assert.Equal(t, "size", result.Data["policy_type"], "Policy type should be extracted correctly")
	assert.Equal(t, false, result.Data["enabled"], "Enabled flag should be extracted correctly")

	// Test nil parameters
	result = helper.ValidateRetentionPolicyParameters(nil)
	assert.False(t, result.Valid, "Nil parameters should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "parameters are required", "Error should mention required parameters")

	// Test missing policy_type parameter
	missingPolicyTypeParams := map[string]interface{}{
		"enabled": true,
	}
	result = helper.ValidateRetentionPolicyParameters(missingPolicyTypeParams)
	assert.False(t, result.Valid, "Missing policy_type parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "policy_type parameter is required", "Error should mention required policy_type parameter")

	// Test invalid policy_type parameter
	invalidPolicyTypeParams := map[string]interface{}{
		"policy_type": "invalid",
		"enabled":     true,
	}
	result = helper.ValidateRetentionPolicyParameters(invalidPolicyTypeParams)
	assert.False(t, result.Valid, "Invalid policy_type parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "policy_type must be either 'age' or 'size'", "Error should mention valid policy types")

	// Test missing enabled parameter
	missingEnabledParams := map[string]interface{}{
		"policy_type": "age",
	}
	result = helper.ValidateRetentionPolicyParameters(missingEnabledParams)
	assert.False(t, result.Valid, "Missing enabled parameter should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "enabled parameter is required", "Error should mention required enabled parameter")

	// Test invalid enabled parameter type (using a type that can't be converted to boolean)
	invalidEnabledParams := map[string]interface{}{
		"policy_type": "age",
		"enabled":     []string{"not", "a", "boolean"}, // Slice type that can't be converted to boolean
	}
	result = helper.ValidateRetentionPolicyParameters(invalidEnabledParams)
	assert.False(t, result.Valid, "Invalid enabled parameter type should fail validation")
	assert.Len(t, result.Errors, 1, "Should have one error")
	assert.Contains(t, result.Errors[0], "enabled parameter must be a boolean", "Error should mention boolean requirement")
}

// TestValidationHelper_CreateValidationErrorResponse tests validation error response creation
func TestValidationHelper_CreateValidationErrorResponse(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test creating error response from validation result
	validationResult := NewValidationResult()
	validationResult.AddError("Test validation error")

	response := helper.CreateValidationErrorResponse(validationResult)
	assert.NotNil(t, response, "Error response should be created")
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.NotNil(t, response.Error, "Error should be present")
	assert.Equal(t, INVALID_PARAMS, response.Error.Code, "Error code should be INVALID_PARAMS")
	assert.Equal(t, ErrorMessages[INVALID_PARAMS], response.Error.Message, "Error message should match")
	// Error data should be *ErrorData struct according to API documentation
	errorData, ok := response.Error.Data.(*ErrorData)
	assert.True(t, ok, "Error data should be of type *ErrorData")
	assert.Equal(t, "validation_failed", errorData.Reason, "Error reason should match")
	assert.Equal(t, "Test validation error", errorData.Details, "Error details should contain validation error")
	assert.Equal(t, "Check parameter types and values", errorData.Suggestion, "Error suggestion should match")
}

// TestValidationHelper_ExistingMethods tests the actual validation methods that exist
func TestValidationHelper_ExistingMethods(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test 1: Device parameter validation
	t.Run("DeviceParameterValidation", func(t *testing.T) {
		// Valid device parameter
		params := map[string]interface{}{
			"device": "camera0",
		}
		result := helper.ValidateDeviceParameter(params)
		assert.True(t, result.Valid, "Valid device parameter should pass")

		// Missing device parameter
		params = map[string]interface{}{}
		result = helper.ValidateDeviceParameter(params)
		assert.False(t, result.Valid, "Missing device parameter should fail")

		// Invalid device parameter type
		params = map[string]interface{}{
			"device": 123,
		}
		result = helper.ValidateDeviceParameter(params)
		assert.False(t, result.Valid, "Invalid device parameter type should fail")
	})

	// Test 2: Filename parameter validation
	t.Run("FilenameParameterValidation", func(t *testing.T) {
		// Valid filename parameter
		params := map[string]interface{}{
			"filename": "test_file.mp4",
		}
		result := helper.ValidateFilenameParameter(params)
		assert.True(t, result.Valid, "Valid filename parameter should pass")

		// Missing filename parameter
		params = map[string]interface{}{}
		result = helper.ValidateFilenameParameter(params)
		assert.False(t, result.Valid, "Missing filename parameter should fail")

		// Invalid filename parameter type
		params = map[string]interface{}{
			"filename": 123,
		}
		result = helper.ValidateFilenameParameter(params)
		assert.False(t, result.Valid, "Invalid filename parameter type should fail")
	})

	// Test 3: Pagination parameter validation
	t.Run("PaginationParameterValidation", func(t *testing.T) {
		// Valid pagination parameters
		params := map[string]interface{}{
			"limit":  10,
			"offset": 0,
		}
		result := helper.ValidatePaginationParams(params)
		assert.True(t, result.Valid, "Valid pagination parameters should pass")

		// Invalid limit (negative)
		params = map[string]interface{}{
			"limit":  -1,
			"offset": 0,
		}
		result = helper.ValidatePaginationParams(params)
		assert.False(t, result.Valid, "Negative limit should fail")

		// Invalid offset (negative)
		params = map[string]interface{}{
			"limit":  10,
			"offset": -1,
		}
		result = helper.ValidatePaginationParams(params)
		assert.False(t, result.Valid, "Negative offset should fail")

		// Invalid limit type
		params = map[string]interface{}{
			"limit":  "invalid",
			"offset": 0,
		}
		result = helper.ValidatePaginationParams(params)
		assert.False(t, result.Valid, "Invalid limit type should fail")
	})

	// Test 4: Recording parameter validation
	t.Run("RecordingParameterValidation", func(t *testing.T) {
		// Valid recording parameters
		params := map[string]interface{}{
			"device":   "camera0",
			"duration": 3600,
			"format":   "mp4",
		}
		result := helper.ValidateRecordingParameters(params)
		assert.True(t, result.Valid, "Valid recording parameters should pass")

		// Missing required device parameter
		params = map[string]interface{}{
			"duration": 3600,
			"format":   "mp4",
		}
		result = helper.ValidateRecordingParameters(params)
		assert.False(t, result.Valid, "Missing device parameter should fail")

		// Invalid duration (negative)
		params = map[string]interface{}{
			"device":   "camera0",
			"duration": -1,
			"format":   "mp4",
		}
		result = helper.ValidateRecordingParameters(params)
		assert.False(t, result.Valid, "Negative duration should fail")
	})

	// Test 5: Snapshot parameter validation
	t.Run("SnapshotParameterValidation", func(t *testing.T) {
		// Valid snapshot parameters
		params := map[string]interface{}{
			"device":   "camera0",
			"filename": "snapshot.jpg",
		}
		result := helper.ValidateSnapshotParameters(params)
		assert.True(t, result.Valid, "Valid snapshot parameters should pass")

		// Missing required device parameter
		params = map[string]interface{}{
			"filename": "snapshot.jpg",
		}
		result = helper.ValidateSnapshotParameters(params)
		assert.False(t, result.Valid, "Missing device parameter should fail")

		// Invalid filename type
		params = map[string]interface{}{
			"device":   "camera0",
			"filename": 123,
		}
		result = helper.ValidateSnapshotParameters(params)
		assert.False(t, result.Valid, "Invalid filename type should fail")
	})

	// Test 6: Retention policy parameter validation
	t.Run("RetentionPolicyParameterValidation", func(t *testing.T) {
		// Valid retention policy parameters
		params := map[string]interface{}{
			"policy_type":  "age",
			"max_age_days": 30,
			"enabled":      true,
		}
		result := helper.ValidateRetentionPolicyParameters(params)
		assert.True(t, result.Valid, "Valid retention policy parameters should pass")

		// Missing required policy_type
		params = map[string]interface{}{
			"max_age_days": 30,
			"enabled":      true,
		}
		result = helper.ValidateRetentionPolicyParameters(params)
		assert.False(t, result.Valid, "Missing policy_type should fail")

		// Invalid policy_type
		params = map[string]interface{}{
			"policy_type":  "invalid",
			"max_age_days": 30,
			"enabled":      true,
		}
		result = helper.ValidateRetentionPolicyParameters(params)
		assert.False(t, result.Valid, "Invalid policy_type should fail")
	})
}

// TestValidationHelper_LogValidationWarnings tests validation warning logging
func TestValidationHelper_LogValidationWarnings(t *testing.T) {
	logger := NewTestLogger("test")
	inputValidator := security.NewInputValidator(logger, nil)
	helper := NewValidationHelper(inputValidator, logger)

	// Test logging validation warnings
	validationResult := NewValidationResult()
	validationResult.AddWarning("Test warning 1")
	validationResult.AddWarning("Test warning 2")

	// This should not panic and should log warnings
	helper.LogValidationWarnings(validationResult, "test_method", "test_client")

	// Test with no warnings (should not panic)
	emptyResult := NewValidationResult()
	helper.LogValidationWarnings(emptyResult, "test_method", "test_client")
}
