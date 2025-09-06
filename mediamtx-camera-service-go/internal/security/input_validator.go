package security

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// InputValidator provides centralized input validation and sanitization.
// It uses SecurityConfigProvider for type-safe configuration access and eliminates
// the need for interface{} usage, improving type safety and maintainability.
type InputValidator struct {
	logger *logging.Logger
	config SecurityConfigProvider // Type-safe configuration provider
}

// NewInputValidator creates a new input validator with type-safe configuration.
// It accepts a SecurityConfigProvider interface to ensure type safety and eliminate
// the need for interface{} usage and type assertions.
func NewInputValidator(logger *logging.Logger, config SecurityConfigProvider) *InputValidator {
	return &InputValidator{
		logger: logger,
		config: config,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", ve.Field, ve.Message, ve.Value)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool
	Errors   []*ValidationError
	Warnings []string
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field, message string, value interface{}) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// AddWarning adds a validation warning
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// GetErrorMessages returns all error messages as a slice
func (vr *ValidationResult) GetErrorMessages() []string {
	messages := make([]string, len(vr.Errors))
	for i, err := range vr.Errors {
		messages[i] = err.Error()
	}
	return messages
}

// Camera ID validation patterns
var (
	cameraIDPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^camera[0-9]+$`),                                      // USB cameras
		regexp.MustCompile(`^ip_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`),             // IP cameras
		regexp.MustCompile(`^http_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`),           // HTTP cameras
		regexp.MustCompile(`^network_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`), // Network cameras
		regexp.MustCompile(`^file_camera_[a-zA-Z0-9_]+$`),                         // File sources
		regexp.MustCompile(`^camera_[0-9]+$`),                                     // Hash-based fallback
	}
)

// ValidateCameraID validates camera identifier format against known patterns.
// This method prevents injection attacks by ensuring camera IDs match expected
// formats for USB cameras, IP cameras, HTTP cameras, network cameras, and file sources.
// Returns a ValidationResult with detailed error information if validation fails.
func (iv *InputValidator) ValidateCameraID(cameraID string) *ValidationResult {
	result := NewValidationResult()

	if cameraID == "" {
		result.AddError("camera_id", "cannot be empty", cameraID)
		return result
	}

	// Check if camera ID matches any valid pattern
	valid := false
	for _, pattern := range cameraIDPatterns {
		if pattern.MatchString(cameraID) {
			valid = true
			break
		}
	}

	if !valid {
		result.AddError("camera_id", "invalid format", cameraID)
		iv.logger.WithFields(logging.Fields{
			"camera_id": cameraID,
			"action":    "validation_failed",
		}).Warn("Invalid camera ID format detected")
	}

	return result
}

// ValidateDuration validates and parses duration strings
func (iv *InputValidator) ValidateDuration(duration string) (*ValidationResult, time.Duration) {
	result := NewValidationResult()

	if duration == "" {
		result.AddError("duration", "cannot be empty", duration)
		return result, 0
	}

	// Parse duration
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		result.AddError("duration", "invalid format", duration)
		iv.logger.WithFields(logging.Fields{
			"duration": duration,
			"error":    err.Error(),
			"action":   "validation_failed",
		}).Warn("Invalid duration format detected")
		return result, 0
	}

	// Check reasonable bounds (1 second to 24 hours)
	if parsedDuration < time.Second {
		result.AddError("duration", "must be at least 1 second", duration)
	}
	if parsedDuration > 24*time.Hour {
		result.AddError("duration", "cannot exceed 24 hours", duration)
	}

	return result, parsedDuration
}

// ValidateResolution validates resolution strings
func (iv *InputValidator) ValidateResolution(resolution string) *ValidationResult {
	result := NewValidationResult()

	if resolution == "" {
		return result // Empty resolution is optional
	}

	// Expected format: "1920x1080"
	resolutionPattern := regexp.MustCompile(`^(\d+)x(\d+)$`)
	matches := resolutionPattern.FindStringSubmatch(resolution)

	if len(matches) != 3 {
		result.AddError("resolution", "invalid format, expected WIDTHxHEIGHT", resolution)
		return result
	}

	width, err := strconv.Atoi(matches[1])
	if err != nil {
		result.AddError("resolution", "invalid width value", matches[1])
		return result
	}

	height, err := strconv.Atoi(matches[2])
	if err != nil {
		result.AddError("resolution", "invalid height value", matches[2])
		return result
	}

	// Check reasonable bounds
	if width < 1 || width > 7680 {
		result.AddError("resolution", "width must be between 1 and 7680", width)
	}
	if height < 1 || height > 4320 {
		result.AddError("resolution", "height must be between 1 and 4320", height)
	}

	return result
}

// ValidateFPS validates FPS values
func (iv *InputValidator) ValidateFPS(fps interface{}) *ValidationResult {
	result := NewValidationResult()

	if fps == nil {
		return result // FPS is optional
	}

	var fpsValue float64
	switch v := fps.(type) {
	case float64:
		fpsValue = v
	case int:
		fpsValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			fpsValue = parsed
		} else {
			result.AddError("fps", "invalid numeric format", v)
			return result
		}
	default:
		result.AddError("fps", "unsupported type", fps)
		return result
	}

	// Check reasonable bounds (1 to 300 FPS)
	if fpsValue < 1 || fpsValue > 300 {
		result.AddError("fps", "must be between 1 and 300", fpsValue)
	}

	return result
}

// ValidateQuality validates quality strings
func (iv *InputValidator) ValidateQuality(quality string) *ValidationResult {
	result := NewValidationResult()

	if quality == "" {
		return result // Quality is optional
	}

	validQualities := []string{"low", "medium", "high", "ultra"}
	valid := false
	for _, q := range validQualities {
		if strings.ToLower(quality) == q {
			valid = true
			break
		}
	}

	if !valid {
		result.AddError("quality", "must be one of: low, medium, high, ultra", quality)
	}

	return result
}

// ValidatePriority validates priority values
func (iv *InputValidator) ValidatePriority(priority interface{}) *ValidationResult {
	result := NewValidationResult()

	if priority == nil {
		return result // Priority is optional
	}

	var priorityValue int
	switch v := priority.(type) {
	case int:
		priorityValue = v
	case float64:
		priorityValue = int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			priorityValue = parsed
		} else {
			result.AddError("priority", "invalid numeric format", v)
			return result
		}
	default:
		result.AddError("priority", "unsupported type", priority)
		return result
	}

	// Check reasonable bounds (1 to 10)
	if priorityValue < 1 || priorityValue > 10 {
		result.AddError("priority", "must be between 1 and 10", priorityValue)
	}

	return result
}

// ValidateRetentionDays validates retention days
func (iv *InputValidator) ValidateRetentionDays(retentionDays interface{}) *ValidationResult {
	result := NewValidationResult()

	if retentionDays == nil {
		return result // Retention days is optional
	}

	var retentionValue int
	switch v := retentionDays.(type) {
	case int:
		retentionValue = v
	case float64:
		retentionValue = int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			retentionValue = parsed
		} else {
			result.AddError("retention_days", "invalid numeric format", v)
			return result
		}
	default:
		result.AddError("retention_days", "unsupported type", retentionDays)
		return result
	}

	// Check reasonable bounds (1 to 365 days)
	if retentionValue < 1 || retentionValue > 365 {
		result.AddError("retention_days", "must be between 1 and 10", retentionValue)
	}

	return result
}

// ValidateUseCase validates use case strings
func (iv *InputValidator) ValidateUseCase(useCase string) *ValidationResult {
	result := NewValidationResult()

	if useCase == "" {
		return result // Use case is optional
	}

	validUseCases := []string{"recording", "snapshot", "streaming", "monitoring"}
	valid := false
	for _, uc := range validUseCases {
		if strings.ToLower(useCase) == uc {
			valid = true
			break
		}
	}

	if !valid {
		result.AddError("use_case", "must be one of: recording, snapshot, streaming, monitoring", useCase)
	}

	return result
}

// ValidateAutoCleanup validates auto cleanup boolean values
func (iv *InputValidator) ValidateAutoCleanup(autoCleanup interface{}) *ValidationResult {
	result := NewValidationResult()

	if autoCleanup == nil {
		return result // Auto cleanup is optional
	}

	switch v := autoCleanup.(type) {
	case bool:
		// Boolean is always valid
	case string:
		if strings.ToLower(v) != "true" && strings.ToLower(v) != "false" {
			result.AddError("auto_cleanup", "must be true or false", v)
		}
	case int:
		if v != 0 && v != 1 {
			result.AddError("auto_cleanup", "must be 0 (false) or 1 (true)", v)
		}
	default:
		result.AddError("auto_cleanup", "unsupported type", autoCleanup)
	}

	return result
}

// ValidateRecordingOptions validates recording options map according to MediaMTX API specification
func (iv *InputValidator) ValidateRecordingOptions(options map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if options == nil {
		return result // Options are optional
	}

	// Validate MediaMTX API recording parameters
	// According to API docs: start_recording accepts device (required), duration (optional), format (optional)

	// Validate duration parameter (optional, number)
	if duration, exists := options["duration"]; exists {
		if durationResult := iv.ValidatePositiveInteger(duration, "duration"); durationResult.HasErrors() {
			result.Errors = append(result.Errors, durationResult.Errors...)
		}
	}

	// Validate format parameter (optional, string) - must be "mp4" or "mkv" according to MediaMTX API
	if format, exists := options["format"]; exists {
		if formatResult := iv.ValidateOptionalString(format, "format"); formatResult.HasErrors() {
			result.Errors = append(result.Errors, formatResult.Errors...)
		} else {
			// Additional format validation for MediaMTX API
			if str, ok := format.(string); ok && str != "" {
				if str != "mp4" && str != "mkv" {
					result.AddError("format", "must be 'mp4' or 'mkv'", str)
				}
			}
		}
	}

	return result
}

// SanitizeString removes potentially dangerous characters from strings.
// This method removes null bytes and control characters (except tab, newline, carriage return)
// to prevent injection attacks and ensure safe string handling.
// Returns the sanitized string with trimmed whitespace.
func (iv *InputValidator) SanitizeString(input string) string {
	// Remove null bytes and control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Keep tab, newline, carriage return
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// SanitizeMap sanitizes all string values in a map
func (iv *InputValidator) SanitizeMap(input map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range input {
		switch v := value.(type) {
		case string:
			sanitized[key] = iv.SanitizeString(v)
		case map[string]interface{}:
			sanitized[key] = iv.SanitizeMap(v)
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// ValidateLimit validates limit parameter for pagination
func (iv *InputValidator) ValidateLimit(limit interface{}) *ValidationResult {
	result := NewValidationResult()

	if limit == nil {
		return result // Limit is optional
	}

	switch v := limit.(type) {
	case int:
		if v < 1 || v > 1000 {
			result.AddError("limit", "must be integer between 1 and 1000", v)
		}
	case float64:
		if v < 1 || v > 1000 || v != float64(int(v)) {
			result.AddError("limit", "must be integer between 1 and 1000", v)
		}
	case string:
		if limitInt, err := strconv.Atoi(v); err != nil || limitInt < 1 || limitInt > 1000 {
			result.AddError("limit", "must be integer between 1 and 1000", v)
		}
	default:
		result.AddError("limit", "unsupported type, must be integer between 1 and 1000", limit)
	}

	return result
}

// ValidateOffset validates offset parameter for pagination
func (iv *InputValidator) ValidateOffset(offset interface{}) *ValidationResult {
	result := NewValidationResult()

	if offset == nil {
		return result // Offset is optional
	}

	switch v := offset.(type) {
	case int:
		if v < 0 {
			result.AddError("offset", "must be non-negative integer", v)
		}
	case float64:
		if v < 0 || v != float64(int(v)) {
			result.AddError("offset", "must be non-negative integer", v)
		}
	case string:
		if offsetInt, err := strconv.Atoi(v); err != nil || offsetInt < 0 {
			result.AddError("offset", "must be non-negative integer", v)
		}
	default:
		result.AddError("offset", "unsupported type, must be non-negative integer", offset)
	}

	return result
}

// ValidateDevicePath validates device path format and security.
// This method prevents path traversal attacks by checking for dangerous path characters
// and validates camera identifier format if the device appears to be a camera.
// Returns a ValidationResult with detailed error information if validation fails.
func (iv *InputValidator) ValidateDevicePath(devicePath interface{}) *ValidationResult {
	result := NewValidationResult()

	if devicePath == nil {
		result.AddError("device", "device parameter is required", nil)
		return result
	}

	deviceStr, ok := devicePath.(string)
	if !ok {
		result.AddError("device", "device parameter must be a string", devicePath)
		return result
	}

	if deviceStr == "" {
		result.AddError("device", "device parameter cannot be empty", deviceStr)
		return result
	}

	// Sanitize the device path
	deviceStr = iv.SanitizeString(deviceStr)

	// Check for path traversal attempts
	if strings.Contains(deviceStr, "..") || strings.Contains(deviceStr, "/") || strings.Contains(deviceStr, "\\") {
		result.AddError("device", "device parameter contains invalid path characters", deviceStr)
		return result
	}

	// Validate camera identifier format - all device parameters should be valid camera IDs
	if cameraResult := iv.ValidateCameraID(deviceStr); cameraResult.HasErrors() {
		result.Errors = append(result.Errors, cameraResult.Errors...)
		result.Valid = false
	}

	return result
}

// ValidateFilename validates filename format and security
func (iv *InputValidator) ValidateFilename(filename interface{}) *ValidationResult {
	result := NewValidationResult()

	if filename == nil {
		result.AddError("filename", "filename parameter is required", nil)
		return result
	}

	filenameStr, ok := filename.(string)
	if !ok {
		result.AddError("filename", "filename parameter must be a string", filename)
		return result
	}

	if filenameStr == "" {
		result.AddError("filename", "filename parameter cannot be empty", filenameStr)
		return result
	}

	// Check filename length (prevent extremely long filenames)
	if len(filenameStr) > 255 {
		result.AddError("filename", "filename too long (max 255 characters)", filenameStr)
		return result
	}

	// Sanitize the filename
	filenameStr = iv.SanitizeString(filenameStr)

	// Check for path traversal attempts (but allow legitimate paths)
	if strings.Contains(filenameStr, "..") {
		result.AddError("filename", "filename parameter contains path traversal attempt", filenameStr)
		return result
	}

	// Check for dangerous characters that could be used in injection attacks
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(filenameStr, char) {
			result.AddError("filename", fmt.Sprintf("filename parameter contains invalid character: %s", char), filenameStr)
			return result
		}
	}

	// Check for potentially dangerous extensions (optional security measure)
	dangerousExtensions := []string{".exe", ".bat", ".sh", ".py", ".js", ".php"}
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(strings.ToLower(filenameStr), ext) {
			result.AddWarning(fmt.Sprintf("filename has potentially dangerous extension: %s", ext))
		}
	}

	return result
}

// ValidateIntegerRange validates integer parameters within a specified range
func (iv *InputValidator) ValidateIntegerRange(value interface{}, fieldName string, min, max int) *ValidationResult {
	result := NewValidationResult()

	if value == nil {
		return result // Value is optional
	}

	switch v := value.(type) {
	case int:
		if v < min || v > max {
			result.AddError(fieldName, fmt.Sprintf("must be integer between %d and %d", min, max), v)
		}
	case float64:
		if v < float64(min) || v > float64(max) || v != float64(int(v)) {
			result.AddError(fieldName, fmt.Sprintf("must be integer between %d and %d", min, max), v)
		}
	case string:
		if intVal, err := strconv.Atoi(v); err != nil || intVal < min || intVal > max {
			result.AddError(fieldName, fmt.Sprintf("must be integer between %d and %d", min, max), v)
		}
	default:
		result.AddError(fieldName, fmt.Sprintf("unsupported type, must be integer between %d and %d", min, max), value)
	}

	return result
}

// ValidatePositiveInteger validates that a parameter is a positive integer
func (iv *InputValidator) ValidatePositiveInteger(value interface{}, fieldName string) *ValidationResult {
	return iv.ValidateIntegerRange(value, fieldName, 1, 2147483647) // Max int32
}

// ValidateNonNegativeInteger validates that a parameter is a non-negative integer
func (iv *InputValidator) ValidateNonNegativeInteger(value interface{}, fieldName string) *ValidationResult {
	return iv.ValidateIntegerRange(value, fieldName, 0, 2147483647) // Max int32
}

// ValidateStringParameter validates a required string parameter
func (iv *InputValidator) ValidateStringParameter(value interface{}, fieldName string, allowEmpty bool) *ValidationResult {
	result := NewValidationResult()

	if value == nil {
		result.AddError(fieldName, fmt.Sprintf("%s parameter is required", fieldName), nil)
		return result
	}

	str, ok := value.(string)
	if !ok {
		result.AddError(fieldName, fmt.Sprintf("%s parameter must be a string", fieldName), value)
		return result
	}

	if !allowEmpty && str == "" {
		result.AddError(fieldName, fmt.Sprintf("%s parameter cannot be empty", fieldName), str)
		return result
	}

	// Check for whitespace-only strings
	if strings.TrimSpace(str) == "" && str != "" {
		result.AddError(fieldName, fmt.Sprintf("%s parameter cannot be only whitespace", fieldName), str)
		return result
	}

	// Check for control characters (newlines, tabs, etc.)
	for _, char := range str {
		if char < 32 { // Reject all control characters including tab, newline, carriage return
			result.AddError(fieldName, fmt.Sprintf("%s parameter contains invalid control character", fieldName), str)
			return result
		}
	}

	// Sanitize the string
	sanitized := iv.SanitizeString(str)
	if sanitized != str {
		result.AddWarning(fmt.Sprintf("%s parameter was sanitized", fieldName))
	}

	return result
}

// ValidateOptionalString validates an optional string parameter
func (iv *InputValidator) ValidateOptionalString(value interface{}, fieldName string) *ValidationResult {
	result := NewValidationResult()

	if value == nil {
		return result // Value is optional
	}

	str, ok := value.(string)
	if !ok {
		result.AddError(fieldName, fmt.Sprintf("%s parameter must be a string", fieldName), value)
		return result
	}

	// Check for whitespace-only strings (but allow empty strings)
	if str != "" && strings.TrimSpace(str) == "" {
		result.AddError(fieldName, fmt.Sprintf("%s parameter cannot be only whitespace", fieldName), str)
		return result
	}

	// Check for control characters (newlines, tabs, etc.)
	for _, char := range str {
		if char < 32 { // Reject all control characters including tab, newline, carriage return
			result.AddError(fieldName, fmt.Sprintf("%s parameter contains invalid control character", fieldName), str)
			return result
		}
	}

	// Sanitize the string if not empty
	if str != "" {
		sanitized := iv.SanitizeString(str)
		if sanitized != str {
			result.AddWarning(fmt.Sprintf("%s parameter was sanitized", fieldName))
		}
	}

	return result
}

// ValidateBooleanParameter validates a boolean parameter
func (iv *InputValidator) ValidateBooleanParameter(value interface{}, fieldName string) *ValidationResult {
	result := NewValidationResult()

	if value == nil {
		result.AddError(fieldName, "boolean parameter is required", nil)
		return result
	}

	switch v := value.(type) {
	case bool:
		// Boolean is always valid
	case string:
		lowerStr := strings.ToLower(v)
		if lowerStr != "true" && lowerStr != "false" && lowerStr != "1" && lowerStr != "0" {
			result.AddError(fieldName, "must be true, false, 1, or 0", v)
		}
	case int:
		if v != 0 && v != 1 {
			result.AddError(fieldName, "must be 0 (false) or 1 (true)", v)
		}
	case float64:
		if v != 0.0 && v != 1.0 {
			result.AddError(fieldName, "must be 0.0 (false) or 1.0 (true)", v)
		}
	default:
		result.AddError(fieldName, "unsupported type, must be boolean", value)
	}

	return result
}

// ValidatePaginationParams validates limit and offset parameters together
func (iv *InputValidator) ValidatePaginationParams(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		return result // No parameters to validate
	}

	// Validate limit
	if limit, exists := params["limit"]; exists {
		if limitResult := iv.ValidateLimit(limit); limitResult.HasErrors() {
			result.Errors = append(result.Errors, limitResult.Errors...)
		}
	}

	// Validate offset
	if offset, exists := params["offset"]; exists {
		if offsetResult := iv.ValidateOffset(offset); offsetResult.HasErrors() {
			result.Errors = append(result.Errors, offsetResult.Errors...)
		}
	}

	// Update overall validation result
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// ValidateCommonRecordingParams validates common recording parameters
func (iv *InputValidator) ValidateCommonRecordingParams(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("params", "parameters are required", nil)
		return result
	}

	// Validate device parameter (required)
	if device, exists := params["device"]; exists {
		if deviceResult := iv.ValidateDevicePath(device); deviceResult.HasErrors() {
			result.Errors = append(result.Errors, deviceResult.Errors...)
		}
	} else {
		result.AddError("device", "device parameter is required", nil)
	}

	// Validate optional parameters
	if duration, exists := params["duration_seconds"]; exists {
		if durationResult := iv.ValidatePositiveInteger(duration, "duration_seconds"); durationResult.HasErrors() {
			result.Errors = append(result.Errors, durationResult.Errors...)
		}
	}

	if format, exists := params["format"]; exists {
		if formatResult := iv.ValidateOptionalString(format, "format"); formatResult.HasErrors() {
			result.Errors = append(result.Errors, formatResult.Errors...)
		} else {
			// Additional format validation for MediaMTX API
			if str, ok := format.(string); ok && str != "" {
				if str != "mp4" && str != "mkv" {
					result.AddError("format", "must be 'mp4' or 'mkv'", str)
				}
			}
		}
	}

	if codec, exists := params["codec"]; exists {
		if codecResult := iv.ValidateOptionalString(codec, "codec"); codecResult.HasErrors() {
			result.Errors = append(result.Errors, codecResult.Errors...)
		}
	}

	if quality, exists := params["quality"]; exists {
		if qualityResult := iv.ValidatePositiveInteger(quality, "quality"); qualityResult.HasErrors() {
			result.Errors = append(result.Errors, qualityResult.Errors...)
		}
	}

	if useCase, exists := params["use_case"]; exists {
		if useCaseResult := iv.ValidateOptionalString(useCase, "use_case"); useCaseResult.HasErrors() {
			result.Errors = append(result.Errors, useCaseResult.Errors...)
		}
	}

	if priority, exists := params["priority"]; exists {
		if priorityResult := iv.ValidatePositiveInteger(priority, "priority"); priorityResult.HasErrors() {
			result.Errors = append(result.Errors, priorityResult.Errors...)
		}
	}

	if autoCleanup, exists := params["auto_cleanup"]; exists {
		if autoCleanupResult := iv.ValidateBooleanParameter(autoCleanup, "auto_cleanup"); autoCleanupResult.HasErrors() {
			result.Errors = append(result.Errors, autoCleanupResult.Errors...)
		}
	}

	if retentionDays, exists := params["retention_days"]; exists {
		if retentionResult := iv.ValidatePositiveInteger(retentionDays, "retention_days"); retentionResult.HasErrors() {
			result.Errors = append(result.Errors, retentionResult.Errors...)
		}
	}

	// Update overall validation result
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}
