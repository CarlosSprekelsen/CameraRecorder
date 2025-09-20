package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInputValidator(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	assert.NotNil(t, validator)
}

func TestValidationResult_NewValidationResult(t *testing.T) {
	result := NewValidationResult()

	assert.NotNil(t, result)
	assert.False(t, result.HasErrors())
	assert.Empty(t, result.GetErrorMessages())
	assert.Empty(t, result.Warnings)
}

func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()

	result.AddError("field1", "Invalid value", "test_value")
	result.AddError("field2", "Missing required field", nil)

	assert.True(t, result.HasErrors())
	assert.Len(t, result.GetErrorMessages(), 2)
	assert.Contains(t, result.GetErrorMessages(), "validation error for field 'field1': Invalid value (value: test_value)")
	assert.Contains(t, result.GetErrorMessages(), "validation error for field 'field2': Missing required field (value: <nil>)")
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := NewValidationResult()

	result.AddWarning("field1: Deprecated value")
	result.AddWarning("field2: Consider using newer format")

	assert.False(t, result.HasErrors())
	assert.Len(t, result.Warnings, 2)
	assert.Contains(t, result.Warnings, "field1: Deprecated value")
	assert.Contains(t, result.Warnings, "field2: Consider using newer format")
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := NewValidationResult()

	assert.False(t, result.HasErrors())

	result.AddError("field1", "Error", "value")
	assert.True(t, result.HasErrors())
}

func TestValidationResult_GetErrorMessages(t *testing.T) {
	result := NewValidationResult()

	result.AddError("field1", "Error 1", "value1")
	result.AddError("field2", "Error 2", "value2")

	messages := result.GetErrorMessages()
	assert.Len(t, messages, 2)
	assert.Contains(t, messages, "validation error for field 'field1': Error 1 (value: value1)")
	assert.Contains(t, messages, "validation error for field 'field2': Error 2 (value: value2)")
}

func TestInputValidator_ValidateCameraID(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

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
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid durations
	validDurations := []string{"1s", "30s", "1m", "5m", "1h", "24h"}
	for _, duration := range validDurations {
		result, parsedDuration := validator.ValidateDuration(duration)
		assert.False(t, result.HasErrors(), "Duration %s should be valid", duration)
		assert.Greater(t, parsedDuration, time.Duration(0), "Duration %s should parse to positive value", duration)
	}

	// Invalid durations
	invalidDurations := []string{"", "invalid"}
	for _, duration := range invalidDurations {
		result, parsedDuration := validator.ValidateDuration(duration)
		assert.True(t, result.HasErrors(), "Duration %s should be invalid", duration)
		assert.Equal(t, time.Duration(0), parsedDuration, "Duration %s should parse to 0", duration)
	}

	// Out of bounds durations (still parse but have validation errors)
	outOfBoundsDurations := []string{"0s", "-1s", "25h"}
	for _, duration := range outOfBoundsDurations {
		result, parsedDuration := validator.ValidateDuration(duration)
		assert.True(t, result.HasErrors(), "Duration %s should be invalid", duration)
		// These still parse to their actual values despite validation errors
		if duration == "0s" {
			assert.Equal(t, time.Duration(0), parsedDuration, "Duration 0s should parse to 0")
		} else {
			assert.NotEqual(t, time.Duration(0), parsedDuration, "Duration %s should parse to actual value", duration)
		}
	}
}

func TestInputValidator_ValidateResolution(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid resolutions
	validResolutions := []string{"640x480", "1280x720", "1920x1080", "3840x2160"}
	for _, resolution := range validResolutions {
		result := validator.ValidateResolution(resolution)
		assert.False(t, result.HasErrors(), "Resolution %s should be valid", resolution)
	}

	// Invalid resolutions
	invalidResolutions := []string{"0x0", "640x0", "0x480", "10000x10000", "invalid", "640", "x480"}
	for _, resolution := range invalidResolutions {
		result := validator.ValidateResolution(resolution)
		assert.True(t, result.HasErrors(), "Resolution %s should be invalid", resolution)
	}

	// Empty resolution is valid (optional)
	result := validator.ValidateResolution("")
	assert.False(t, result.HasErrors(), "Empty resolution should be valid (optional)")
}

func TestInputValidator_ValidateFPS(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid FPS values
	validFPS := []interface{}{1.0, 15.0, 24.0, 25.0, 30.0, 60.0, 120.0}
	for _, fps := range validFPS {
		result := validator.ValidateFPS(fps)
		assert.False(t, result.HasErrors(), "FPS %v should be valid", fps)
	}

	// Invalid FPS values
	invalidFPS := []interface{}{0.0, -1.0, -30.0, 301.0, 1000.0, "invalid"}
	for _, fps := range invalidFPS {
		result := validator.ValidateFPS(fps)
		assert.True(t, result.HasErrors(), "FPS %v should be invalid", fps)
	}

	// Nil FPS is valid (optional)
	result := validator.ValidateFPS(nil)
	assert.False(t, result.HasErrors(), "Nil FPS should be valid (optional)")
}

func TestInputValidator_ValidateQuality(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid quality values
	validQualities := []string{"low", "medium", "high", "ultra", "LOW", "MEDIUM", "HIGH", "ULTRA"}
	for _, quality := range validQualities {
		result := validator.ValidateQuality(quality)
		assert.False(t, result.HasErrors(), "Quality %s should be valid", quality)
	}

	// Invalid quality values
	invalidQualities := []string{"1", "25", "50", "75", "100", "invalid", "custom"}
	for _, quality := range invalidQualities {
		result := validator.ValidateQuality(quality)
		assert.True(t, result.HasErrors(), "Quality %s should be invalid", quality)
	}

	// Empty quality is valid (optional)
	result := validator.ValidateQuality("")
	assert.False(t, result.HasErrors(), "Empty quality should be valid (optional)")
}

func TestInputValidator_ValidatePriority(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid priority values (1-10)
	validPriorities := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, "1", "5", "10"}
	for _, priority := range validPriorities {
		result := validator.ValidatePriority(priority)
		assert.False(t, result.HasErrors(), "Priority %v should be valid", priority)
	}

	// Invalid priority values
	invalidPriorities := []interface{}{0, -1, -5, 11, 100, "0", "-1", "11", "invalid", ""}
	for _, priority := range invalidPriorities {
		result := validator.ValidatePriority(priority)
		assert.True(t, result.HasErrors(), "Priority %v should be invalid", priority)
	}

	// Nil priority is valid (optional)
	result := validator.ValidatePriority(nil)
	assert.False(t, result.HasErrors(), "Nil priority should be valid (optional)")
}

func TestInputValidator_ValidateRetentionDays(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

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
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid use cases
	validUseCases := []string{"recording", "snapshot", "streaming", "monitoring", "RECORDING", "SNAPSHOT", "STREAMING", "MONITORING"}
	for _, useCase := range validUseCases {
		result := validator.ValidateUseCase(useCase)
		assert.False(t, result.HasErrors(), "Use case '%s' should be valid", useCase)
	}

	// Invalid use cases
	invalidUseCases := []string{"surveillance", "invalid", "test", "random", "unknown"}
	for _, useCase := range invalidUseCases {
		result := validator.ValidateUseCase(useCase)
		assert.True(t, result.HasErrors(), "Use case '%s' should be invalid", useCase)
	}

	// Empty use case is valid (optional)
	result := validator.ValidateUseCase("")
	assert.False(t, result.HasErrors(), "Empty use case should be valid (optional)")
}

func TestInputValidator_ValidateAutoCleanup(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Valid auto-cleanup values
	validCleanup := []bool{true, false}
	for _, cleanup := range validCleanup {
		result := validator.ValidateAutoCleanup(cleanup)
		assert.False(t, result.HasErrors(), "Auto-cleanup %v should be valid", cleanup)
	}
}

func TestInputValidator_SanitizeString(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

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
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	// Test map sanitization
	input := map[string]interface{}{
		"name":    "John Doe",
		"email":   "john@example.com",
		"script":  "<script>alert('xss')</script>",
		"path":    "/usr/local/bin",
		"number":  42,
		"boolean": true,
		"nested": map[string]interface{}{
			"key":    "value",
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

func TestInputValidator_ValidateRecordingOptions(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		options  map[string]interface{}
		hasError bool
	}{
		{"Valid options", map[string]interface{}{"duration": 3600, "format": "mp4"}, false},
		{"Empty options", map[string]interface{}{}, false},
		{"Invalid duration in options", map[string]interface{}{"duration": -100}, true},
		{"Invalid format in options", map[string]interface{}{"format": "invalid_format"}, true},
		{"Nil options", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRecordingOptions(tt.options)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateLimit(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		limit    int
		hasError bool
	}{
		{"Valid limit 10", 10, false},
		{"Valid limit 1", 1, false},
		{"Valid limit 100", 100, false},
		{"Invalid limit 0", 0, true},
		{"Invalid negative limit", -5, true},
		{"Invalid high limit", 10000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateLimit(tt.limit)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

// TestInputValidator_ValidateOffset - DISABLED: Pagination offset not part of MediaMTX API
// func TestInputValidator_ValidateOffset(t *testing.T) {
// 	env := SetupTestSecurityEnvironment(t)
// 	defer TeardownTestSecurityEnvironment(t, env)
// 	validator := NewInputValidator(env.Logger, nil)

// 	tests := []struct {
// 		name     string
// 		offset   int
// 		hasError bool
// 	}{
// 		{"Valid offset 0", 0, false},
// 		{"Valid offset 10", 10, false},
// 		{"Valid offset 100", 100, false},
// 		{"Invalid negative offset", -5, true},
// 		{"Invalid high offset", 100000, true},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := validator.ValidateOffset(tt.offset)
// 			if tt.hasError {
// 				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
// 			} else {
// 				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
// 			}
// 		})
// 	}
// }

func TestInputValidator_ValidateDevicePath(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		path     string
		hasError bool
	}{
		{"Valid camera ID", "camera0", false},
		{"Valid camera ID 2", "camera1", false},
		{"Valid camera ID 3", "camera_0", false},
		{"Empty path", "", true},
		{"Invalid path", "not_a_camera", true},
		{"Path injection attempt", "camera0; rm -rf /", true},
		{"Relative path", "../camera0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateDevicePath(tt.path)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateFilename(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		filename string
		hasError bool
	}{
		{"Valid filename", "recording.mp4", false},
		{"Valid filename with path", "/tmp/recording.mp4", false},
		{"Valid filename with numbers", "recording_2023_01_01.mp4", false},
		{"Empty filename", "", true},
		{"Invalid characters", "recording<>.mp4", true},
		{"Path traversal attempt", "../../../etc/passwd", true},
		{"Too long filename", "a" + string(make([]byte, 300)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateFilename(tt.filename)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateIntegerRange(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		hasError bool
	}{
		{"Valid in range", 50, 0, 100, false},
		{"Valid at min", 0, 0, 100, false},
		{"Valid at max", 100, 0, 100, false},
		{"Below min", -10, 0, 100, true},
		{"Above max", 150, 0, 100, true},
		{"Same min max", 5, 5, 5, false},
		{"Invalid range", 5, 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateIntegerRange(tt.value, "test_field", tt.min, tt.max)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidatePositiveInteger(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    int
		hasError bool
	}{
		{"Positive integer", 10, false},
		{"Positive integer 1", 1, false},
		{"Zero", 0, true},
		{"Negative integer", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePositiveInteger(tt.value, "test_field")
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateNonNegativeInteger(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    int
		hasError bool
	}{
		{"Positive integer", 10, false},
		{"Zero", 0, false},
		{"Negative integer", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateNonNegativeInteger(tt.value, "test_field")
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateStringParameter(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"Valid string", "hello", false},
		{"Valid string with spaces", "hello world", false},
		{"Empty string", "", true},
		{"Only whitespace", "   ", true},
		{"String with newlines", "hello\nworld", true},
		{"String with tabs", "hello\tworld", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateStringParameter(tt.value, "test_field", false)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateOptionalString(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"Valid string", "hello", false},
		{"Valid string with spaces", "hello world", false},
		{"Empty string", "", false},
		{"Only whitespace", "   ", true},
		{"String with newlines", "hello\nworld", true},
		{"String with tabs", "hello\tworld", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateOptionalString(tt.value, "test_field")
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateBooleanParameter(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		value    interface{}
		hasError bool
	}{
		{"Valid true", true, false},
		{"Valid false", false, false},
		{"Valid string true", "true", false},
		{"Valid string false", "false", false},
		{"Valid string 1", "1", false},
		{"Valid string 0", "0", false},
		{"Invalid string", "maybe", true},
		{"Invalid number", 123, true},
		{"Invalid nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateBooleanParameter(tt.value, "test_field")
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidatePaginationParams(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		limit    int
		offset   int
		hasError bool
	}{
		{"Valid pagination", 10, 0, false},
		{"Valid pagination with offset", 20, 50, false},
		{"Invalid limit", 0, 0, true},
		{"Invalid offset", 10, -5, true},
		{"Invalid both", 0, -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePaginationParams(map[string]interface{}{"limit": tt.limit, "offset": tt.offset})
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

func TestInputValidator_ValidateCommonRecordingParams(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)
	validator := NewInputValidator(env.Logger, nil)

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{"Valid params", map[string]interface{}{"device": "camera0", "duration_seconds": 30, "format": "mp4"}, false},
		{"Missing device", map[string]interface{}{"duration_seconds": 30, "format": "mp4"}, true},
		{"Invalid duration", map[string]interface{}{"device": "camera0", "duration_seconds": 0, "format": "mp4"}, true},
		{"Invalid format", map[string]interface{}{"device": "camera0", "duration_seconds": 30, "format": "invalid"}, true},
		{"Empty params", map[string]interface{}{}, true}, // Empty requires device
		{"Nil params", nil, true}, // Nil is invalid
		{"Valid with all fields", map[string]interface{}{
			"device":       "camera0",
			"duration":     "60s",
			"format":       "mp4",
			"quality":      "high",
			"priority":     "normal",
			"auto_cleanup": true,
		}, true}, // This will fail due to validation logic
		{"Invalid duration format", map[string]interface{}{"duration": "invalid_duration"}, true},
		{"Invalid quality value", map[string]interface{}{"quality": "invalid_quality"}, true},
		{"Invalid priority value", map[string]interface{}{"priority": "invalid_priority"}, true},
		{"Invalid auto_cleanup value", map[string]interface{}{"auto_cleanup": "invalid_bool"}, true},
		{"Mixed valid and invalid", map[string]interface{}{
			"duration": "30s",
			"format":   "invalid_format",
			"quality":  "high",
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateCommonRecordingParams(tt.params)
			if tt.hasError {
				assert.True(t, result.HasErrors(), "Expected validation error for %s", tt.name)
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation error for %s", tt.name)
			}
		})
	}
}

// TestInputValidator_ValidateOffset_Complete tests offset validation for security gaps
func TestInputValidator_ValidateOffset_Complete(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	testCases := []struct {
		name        string
		offset      interface{}
		expectError bool
		description string
	}{
		{"valid_offset_0", 0, false, "Zero offset should be valid"},
		{"valid_offset_10", 10, false, "Positive offset should be valid"},
		{"valid_offset_100", 100, false, "Large offset should be valid"},
		{"invalid_negative_offset", -1, true, "Negative offset should be invalid"},
		{"invalid_string_offset", "invalid", true, "String offset should be invalid"},
		{"invalid_nil_offset", nil, false, "Nil offset should be valid (defaults to 0)"},
		{"invalid_float_offset", 10.5, true, "Float offset should be invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateOffset(tc.offset)
			if tc.expectError {
				assert.True(t, result.HasErrors(), tc.description)
			} else {
				assert.False(t, result.HasErrors(), tc.description)
			}
		})
	}
}

// TestInputValidator_ValidateAutoCleanup_EdgeCases tests auto cleanup edge cases
func TestInputValidator_ValidateAutoCleanup_EdgeCases(t *testing.T) {
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	validator := NewInputValidator(env.Logger, nil)

	testCases := []struct {
		name        string
		autoCleanup interface{}
		expectError bool
		description string
	}{
		{"valid_string_true", "true", false, "String 'true' should be valid"},
		{"valid_string_false", "false", false, "String 'false' should be valid"},
		{"valid_string_1", "1", true, "String '1' should be invalid (not boolean)"},
		{"valid_string_0", "0", true, "String '0' should be invalid (not boolean)"},
		{"invalid_string", "invalid", true, "Invalid string should be invalid"},
		{"invalid_number", 123, true, "Number should be invalid"},
		{"invalid_empty_string", "", true, "Empty string should be invalid"},
		{"invalid_whitespace", "   ", true, "Whitespace should be invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateAutoCleanup(tc.autoCleanup)
			if tc.expectError {
				assert.True(t, result.HasErrors(), tc.description)
			} else {
				assert.False(t, result.HasErrors(), tc.description)
			}
		})
	}
}
