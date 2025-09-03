//go:build unit
// +build unit

/*
Validation Helper Comprehensive Test

Tests the centralized validation helper for JSON-RPC method parameters.
This ensures that all input validation is consistent and secure across all methods.

Requirements Coverage:
- REQ-FUNC-004: Error handling and validation
- REQ-API-001: JSON-RPC method implementation
- REQ-SEC-001: Input validation and sanitization

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationHelper_ValidatePaginationParams(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidPaginationParams", func(t *testing.T) {
		params := map[string]interface{}{
			"limit":  50,
			"offset": 10,
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, 50, result.Data["limit"])
		assert.Equal(t, 10, result.Data["offset"])
	})

	t.Run("ValidPaginationParamsWithDefaults", func(t *testing.T) {
		params := map[string]interface{}{}

		result := validationHelper.ValidatePaginationParams(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, 100, result.Data["limit"])
		assert.Equal(t, 0, result.Data["offset"])
	})

	t.Run("ValidPaginationParamsNil", func(t *testing.T) {
		result := validationHelper.ValidatePaginationParams(nil)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, 100, result.Data["limit"])
		assert.Equal(t, 0, result.Data["offset"])
	})

	t.Run("InvalidLimitTooSmall", func(t *testing.T) {
		params := map[string]interface{}{
			"limit": 0,
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 1000")
	})

	t.Run("InvalidLimitTooLarge", func(t *testing.T) {
		params := map[string]interface{}{
			"limit": 1001,
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 1000")
	})

	t.Run("InvalidOffsetNegative", func(t *testing.T) {
		params := map[string]interface{}{
			"offset": -1,
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be non-negative integer")
	})

	t.Run("ValidPaginationParamsStringConversion", func(t *testing.T) {
		params := map[string]interface{}{
			"limit":  "25",
			"offset": "5",
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, 25, result.Data["limit"])
		assert.Equal(t, 5, result.Data["offset"])
	})

	t.Run("ValidPaginationParamsFloatConversion", func(t *testing.T) {
		params := map[string]interface{}{
			"limit":  30.0,
			"offset": 15.0,
		}

		result := validationHelper.ValidatePaginationParams(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, 30, result.Data["limit"])
		assert.Equal(t, 15, result.Data["offset"])
	})
}

func TestValidationHelper_ValidateDeviceParameter(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidDeviceParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "camera0",
		}

		result := validationHelper.ValidateDeviceParameter(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "camera0", result.Data["device"])
	})

	t.Run("MissingDeviceParameter", func(t *testing.T) {
		params := map[string]interface{}{}

		result := validationHelper.ValidateDeviceParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter is required")
	})

	t.Run("EmptyDeviceParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "",
		}

		result := validationHelper.ValidateDeviceParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter cannot be empty")
	})

	t.Run("InvalidDeviceType", func(t *testing.T) {
		params := map[string]interface{}{
			"device": 123,
		}

		result := validationHelper.ValidateDeviceParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter must be a string")
	})

	t.Run("DeviceParameterWithPathTraversal", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "../../../etc/passwd",
		}

		result := validationHelper.ValidateDeviceParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "contains invalid path characters")
	})

	t.Run("NilParams", func(t *testing.T) {
		result := validationHelper.ValidateDeviceParameter(nil)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter is required")
	})
}

func TestValidationHelper_ValidateFilenameParameter(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidFilenameParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"filename": "recording_2025-01-15.mp4",
		}

		result := validationHelper.ValidateFilenameParameter(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "recording_2025-01-15.mp4", result.Data["filename"])
	})

	t.Run("MissingFilenameParameter", func(t *testing.T) {
		params := map[string]interface{}{}

		result := validationHelper.ValidateFilenameParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "filename parameter is required")
	})

	t.Run("EmptyFilenameParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"filename": "",
		}

		result := validationHelper.ValidateFilenameParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "filename parameter cannot be empty")
	})

	t.Run("InvalidFilenameType", func(t *testing.T) {
		params := map[string]interface{}{
			"filename": 123,
		}

		result := validationHelper.ValidateFilenameParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "filename parameter must be a string")
	})

	t.Run("FilenameWithPathTraversal", func(t *testing.T) {
		params := map[string]interface{}{
			"filename": "../../../etc/passwd",
		}

		result := validationHelper.ValidateFilenameParameter(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "contains invalid path characters")
	})

	t.Run("NilParams", func(t *testing.T) {
		result := validationHelper.ValidateFilenameParameter(nil)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "filename parameter is required")
	})
}

func TestValidationHelper_ValidateRecordingParameters(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidRecordingParameters", func(t *testing.T) {
		params := map[string]interface{}{
			"device":           "camera0",
			"duration_seconds": 60,
			"format":           "mp4",
			"codec":            "h264",
			"quality":          23,
			"use_case":         "surveillance",
			"priority":         1,
			"auto_cleanup":     true,
			"retention_days":   30,
		}

		result := validationHelper.ValidateRecordingParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "camera0", result.Data["device"])

		options := result.Data["options"].(map[string]interface{})
		assert.Equal(t, 60, options["max_duration"])
		assert.Equal(t, "mp4", options["output_format"])
		assert.Equal(t, "h264", options["codec"])
		assert.Equal(t, 23, options["crf"])
		assert.Equal(t, "surveillance", options["use_case"])
		assert.Equal(t, 1, options["priority"])
		assert.Equal(t, true, options["auto_cleanup"])
		assert.Equal(t, 30, options["retention_days"])
	})

	t.Run("ValidRecordingParametersMinimal", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "camera0",
		}

		result := validationHelper.ValidateRecordingParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "camera0", result.Data["device"])
	})

	t.Run("MissingDeviceParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"duration_seconds": 60,
		}

		result := validationHelper.ValidateRecordingParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter is required")
	})

	t.Run("InvalidDurationParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"device":           "camera0",
			"duration_seconds": -1,
		}

		result := validationHelper.ValidateRecordingParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 2147483647")
	})

	t.Run("InvalidQualityParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"device":  "camera0",
			"quality": 0,
		}

		result := validationHelper.ValidateRecordingParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 2147483647")
	})

	t.Run("NilParams", func(t *testing.T) {
		result := validationHelper.ValidateRecordingParameters(nil)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "parameters are required")
	})
}

func TestValidationHelper_ValidateSnapshotParameters(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidSnapshotParameters", func(t *testing.T) {
		params := map[string]interface{}{
			"device":   "camera0",
			"filename": "snapshot_2025-01-15.jpg",
			"format":   "jpeg",
			"quality":  85,
		}

		result := validationHelper.ValidateSnapshotParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "camera0", result.Data["device"])

		options := result.Data["options"].(map[string]interface{})
		assert.Equal(t, "snapshot_2025-01-15.jpg", options["filename"])
		assert.Equal(t, "jpeg", options["format"])
		assert.Equal(t, 85, options["quality"])
	})

	t.Run("ValidSnapshotParametersMinimal", func(t *testing.T) {
		params := map[string]interface{}{
			"device": "camera0",
		}

		result := validationHelper.ValidateSnapshotParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "camera0", result.Data["device"])
	})

	t.Run("MissingDeviceParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"filename": "snapshot.jpg",
		}

		result := validationHelper.ValidateSnapshotParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "device parameter is required")
	})

	t.Run("InvalidQualityParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"device":  "camera0",
			"quality": 0,
		}

		result := validationHelper.ValidateSnapshotParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 2147483647")
	})

	t.Run("NilParams", func(t *testing.T) {
		result := validationHelper.ValidateSnapshotParameters(nil)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "parameters are required")
	})
}

func TestValidationHelper_ValidateRetentionPolicyParameters(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("ValidRetentionPolicyParametersAge", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type":  "age",
			"enabled":      true,
			"max_age_days": 30,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "age", result.Data["policy_type"])
		assert.Equal(t, true, result.Data["enabled"])
	})

	t.Run("ValidRetentionPolicyParametersSize", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "size",
			"enabled":     true,
			"max_size_gb": 10,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "size", result.Data["policy_type"])
		assert.Equal(t, true, result.Data["enabled"])
	})

	t.Run("ValidRetentionPolicyParametersMinimal", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "age",
			"enabled":     false,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.True(t, result.Valid, "Validation should pass")
		assert.Equal(t, "age", result.Data["policy_type"])
		assert.Equal(t, false, result.Data["enabled"])
	})

	t.Run("MissingPolicyTypeParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"enabled": true,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "policy_type parameter is required")
	})

	t.Run("MissingEnabledParameter", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "age",
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "enabled parameter is required")
	})

	t.Run("InvalidPolicyType", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "invalid",
			"enabled":     true,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be either 'age' or 'size'")
	})

	t.Run("InvalidEnabledType", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "age",
			"enabled":     "invalid",
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be a boolean")
	})

	t.Run("InvalidMaxAgeDays", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type":  "age",
			"enabled":      true,
			"max_age_days": -1,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 2147483647")
	})

	t.Run("InvalidMaxSizeGB", func(t *testing.T) {
		params := map[string]interface{}{
			"policy_type": "size",
			"enabled":     true,
			"max_size_gb": 0,
		}

		result := validationHelper.ValidateRetentionPolicyParameters(params)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "must be integer between 1 and 2147483647")
	})

	t.Run("NilParams", func(t *testing.T) {
		result := validationHelper.ValidateRetentionPolicyParameters(nil)
		require.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.Errors[0], "parameters are required")
	})
}

func TestValidationHelper_CreateValidationErrorResponse(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("CreateErrorResponse", func(t *testing.T) {
		validationResult := NewValidationResult()
		validationResult.AddError("Test error message")

		response := validationHelper.CreateValidationErrorResponse(validationResult)
		require.NotNil(t, response, "Response should not be nil")
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.NotNil(t, response.Error)
		assert.Equal(t, INVALID_PARAMS, response.Error.Code)
		assert.Equal(t, ErrorMessages[INVALID_PARAMS], response.Error.Message)
		assert.Equal(t, "Test error message", response.Error.Data)
	})
}

func TestValidationHelper_LogValidationWarnings(t *testing.T) {
	// Create validation helper
	inputValidator := security.NewInputValidator(nil, nil)
	logger := logrus.New()
	validationHelper := NewValidationHelper(inputValidator, logger)

	t.Run("LogWarnings", func(t *testing.T) {
		validationResult := NewValidationResult()
		validationResult.AddWarning("Test warning message")

		// This should not panic and should log the warning
		validationHelper.LogValidationWarnings(validationResult, "test_method", "test_client")
		// Note: In a real test, you might want to capture log output to verify
	})

	t.Run("NoWarnings", func(t *testing.T) {
		validationResult := NewValidationResult()

		// This should not panic and should not log anything
		validationHelper.LogValidationWarnings(validationResult, "test_method", "test_client")
	})
}
