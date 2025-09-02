package security_test

import (
	"testing"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewInputValidator(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	assert.NotNil(t, validator)
}

func TestValidationResult_NewValidationResult(t *testing.T) {
	result := security.NewValidationResult()
	
	assert.NotNil(t, result)
	assert.False(t, result.HasErrors())
	assert.Empty(t, result.GetErrorMessages())
	assert.Empty(t, result.Warnings)
}

func TestValidationResult_AddError(t *testing.T) {
	result := security.NewValidationResult()
	
	result.AddError("field1", "Invalid value", "test_value")
	result.AddError("field2", "Missing required field", nil)
	
	assert.True(t, result.HasErrors())
	assert.Len(t, result.GetErrorMessages(), 2)
	assert.Contains(t, result.GetErrorMessages(), "validation error for field 'field1': Invalid value")
	assert.Contains(t, result.GetErrorMessages(), "validation error for field 'field2': Missing required field")
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := security.NewValidationResult()
	
	result.AddWarning("field1: Deprecated value")
	result.AddWarning("field2: Consider using newer format")
	
	assert.False(t, result.HasErrors())
	assert.Len(t, result.Warnings, 2)
	assert.Contains(t, result.Warnings, "field1: Deprecated value")
	assert.Contains(t, result.Warnings, "field2: Consider using newer format")
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := security.NewValidationResult()
	
	assert.False(t, result.HasErrors())
	
	result.AddError("field1", "Error", "value")
	assert.True(t, result.HasErrors())
}

func TestValidationResult_GetErrorMessages(t *testing.T) {
	result := security.NewValidationResult()
	
	result.AddError("field1", "Error 1", "value1")
	result.AddError("field2", "Error 2", "value2")
	
	messages := result.GetErrorMessages()
	assert.Len(t, messages, 2)
	assert.Contains(t, messages, "validation error for field 'field1': Error 1")
	assert.Contains(t, messages, "validation error for field 'field2': Error 2")
}

func TestInputValidator_ValidateCameraID(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid camera IDs
	validIDs := []string{"camera001", "camera123", "ip_camera_192_168_1_100"}
	for _, id := range validIDs {
		result := validator.ValidateCameraID(id)
		assert.False(t, result.HasErrors(), "Camera ID '%s' should be valid", id)
	}
	
	// Invalid camera IDs
	invalidIDs := []string{"", "camera", "CAM_", "camera@123"}
	for _, id := range invalidIDs {
		result := validator.ValidateCameraID(id)
		assert.True(t, result.HasErrors(), "Camera ID '%s' should be invalid", id)
	}
}

func TestInputValidator_ValidateDuration(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid durations
	validDurations := []string{"1s", "30s", "1m", "5m", "1h", "24h"}
	for _, duration := range validDurations {
		result, parsedDuration := validator.ValidateDuration(duration)
		assert.False(t, result.HasErrors(), "Duration %s should be valid", duration)
		assert.Greater(t, parsedDuration, 0, "Duration %s should parse to positive value", duration)
	}
	
	// Invalid durations
	invalidDurations := []string{"", "0s", "-1s", "25h", "invalid"}
	for _, duration := range invalidDurations {
		result, parsedDuration := validator.ValidateDuration(duration)
		assert.True(t, result.HasErrors(), "Duration %s should be invalid", duration)
		assert.Equal(t, 0, parsedDuration, "Duration %s should parse to 0", duration)
	}
}

func TestInputValidator_ValidateResolution(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid resolutions
	validResolutions := []string{"640x480", "1280x720", "1920x1080", "3840x2160"}
	for _, resolution := range validResolutions {
		result := validator.ValidateResolution(resolution)
		assert.False(t, result.HasErrors(), "Resolution %s should be valid", resolution)
	}
	
	// Invalid resolutions
	invalidResolutions := []string{"", "0x0", "640x0", "0x480", "10000x10000", "invalid", "640", "x480"}
	for _, resolution := range invalidResolutions {
		result := validator.ValidateResolution(resolution)
		assert.True(t, result.HasErrors(), "Resolution %s should be invalid", resolution)
	}
}

func TestInputValidator_ValidateFPS(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid FPS values
	validFPS := []interface{}{1.0, 15.0, 24.0, 25.0, 30.0, 60.0, 120.0}
	for _, fps := range validFPS {
		result := validator.ValidateFPS(fps)
		assert.False(t, result.HasErrors(), "FPS %v should be valid", fps)
	}
	
	// Invalid FPS values
	invalidFPS := []interface{}{0.0, -1.0, -30.0, 121.0, 1000.0, "invalid", nil}
	for _, fps := range invalidFPS {
		result := validator.ValidateFPS(fps)
		assert.True(t, result.HasErrors(), "FPS %v should be invalid", fps)
	}
}

func TestInputValidator_ValidateQuality(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid quality values
	validQualities := []string{"1", "25", "50", "75", "100"}
	for _, quality := range validQualities {
		result := validator.ValidateQuality(quality)
		assert.False(t, result.HasErrors(), "Quality %s should be valid", quality)
	}
	
	// Invalid quality values
	invalidQualities := []string{"0", "-1", "-50", "101", "150", "invalid", ""}
	for _, quality := range invalidQualities {
		result := validator.ValidateQuality(quality)
		assert.True(t, result.HasErrors(), "Quality %s should be invalid", quality)
	}
}

func TestInputValidator_ValidatePriority(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid priority values
	validPriorities := []string{"1", "2", "3", "4", "5"}
	for _, priority := range validPriorities {
		result := validator.ValidatePriority(priority)
		assert.False(t, result.HasErrors(), "Priority %s should be valid", priority)
	}
	
	// Invalid priority values
	invalidPriorities := []string{"0", "-1", "-5", "6", "10", "invalid", ""}
	for _, priority := range invalidPriorities {
		result := validator.ValidatePriority(priority)
		assert.True(t, result.HasErrors(), "Priority %s should be invalid", priority)
	}
}

func TestInputValidator_ValidateRetentionDays(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid retention days
	validDays := []string{"1", "7", "30", "90", "365"}
	for _, days := range validDays {
		result := validator.ValidateRetentionDays(days)
		assert.False(t, result.HasErrors(), "Retention days %s should be valid", days)
	}
	
	// Invalid retention days
	invalidDays := []string{"0", "-1", "-7", "366", "1000", "invalid", ""}
	for _, days := range invalidDays {
		result := validator.ValidateRetentionDays(days)
		assert.True(t, result.HasErrors(), "Retention days %s should be invalid", days)
	}
}

func TestInputValidator_ValidateUseCase(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid use cases
	validUseCases := []string{"surveillance", "monitoring", "recording", "snapshot", "streaming"}
	for _, useCase := range validUseCases {
		result := validator.ValidateUseCase(useCase)
		assert.False(t, result.HasErrors(), "Use case '%s' should be valid", useCase)
	}
	
	// Invalid use cases
	invalidUseCases := []string{"", "invalid", "test", "random", "unknown"}
	for _, useCase := range invalidUseCases {
		result := validator.ValidateUseCase(useCase)
		assert.True(t, result.HasErrors(), "Use case '%s' should be invalid", useCase)
	}
}

func TestInputValidator_ValidateAutoCleanup(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Valid auto-cleanup values
	validCleanup := []bool{true, false}
	for _, cleanup := range validCleanup {
		result := validator.ValidateAutoCleanup(cleanup)
		assert.False(t, result.HasErrors(), "Auto-cleanup %v should be valid", cleanup)
	}
}

func TestInputValidator_SanitizeString(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Test string sanitization
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world"},
		{"<script>alert('xss')</script>", "<script>alert('xss')</script>"},
		{"user@example.com", "user@example.com"},
		{"file/path/with\\backslashes", "file/path/with\\backslashes"},
		{"normal text 123", "normal text 123"},
		{"", ""},
	}
	
	for _, tc := range testCases {
		result := validator.SanitizeString(tc.input)
		assert.Equal(t, tc.expected, result, "Input: '%s'", tc.input)
	}
}

func TestInputValidator_SanitizeMap(t *testing.T) {
	logger := logging.NewLogger("test")
	validator := security.NewInputValidator(logger, nil)
	
	// Test map sanitization
	input := map[string]interface{}{
		"name":        "John Doe",
		"email":       "john@example.com",
		"script":      "<script>alert('xss')</script>",
		"path":        "/usr/local/bin",
		"number":      42,
		"boolean":     true,
		"nested": map[string]interface{}{
			"key":   "value",
			"script": "<script>alert('nested')</script>",
		},
	}
	
	result := validator.SanitizeMap(input)
	
	// Check that strings are sanitized
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
	assert.Equal(t, "<script>alert('xss')</script>", result["script"])
	assert.Equal(t, "/usr/local/bin", result["path"])
	
	// Check that non-strings are preserved
	assert.Equal(t, 42, result["number"])
	assert.Equal(t, true, result["boolean"])
	
	// Check nested map
	nested, ok := result["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", nested["key"])
	assert.Equal(t, "<script>alert('nested')</script>", nested["script"])
}
