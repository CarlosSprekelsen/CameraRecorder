package security

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// InputValidator provides centralized input validation and sanitization
type InputValidator struct {
	logger *logrus.Logger
	config interface{} // Will be typed based on existing config structure
}

// NewInputValidator creates a new input validator
func NewInputValidator(logger *logrus.Logger, config interface{}) *InputValidator {
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
	Valid   bool
	Errors  []*ValidationError
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
		regexp.MustCompile(`^camera[0-9]+$`),                                    // USB cameras
		regexp.MustCompile(`^ip_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`),         // IP cameras
		regexp.MustCompile(`^http_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`),       // HTTP cameras
		regexp.MustCompile(`^network_camera_[0-9]+_[0-9]+_[0-9]+_[0-9]+_[0-9]+$`), // Network cameras
		regexp.MustCompile(`^file_camera_[a-zA-Z0-9_]+$`),                       // File sources
		regexp.MustCompile(`^camera_[0-9]+$`),                                   // Hash-based fallback
	}
)

// ValidateCameraID validates camera identifier format
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
		iv.logger.WithFields(logrus.Fields{
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
		iv.logger.WithFields(logrus.Fields{
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

// ValidateRecordingOptions validates recording options map
func (iv *InputValidator) ValidateRecordingOptions(options map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if options == nil {
		return result // Options are optional
	}

	// Validate individual options
	if duration, exists := options["duration"]; exists {
		if durationResult, _ := iv.ValidateDuration(fmt.Sprintf("%v", duration)); durationResult.HasErrors() {
			result.Errors = append(result.Errors, durationResult.Errors...)
		}
	}

	if quality, exists := options["quality"]; exists {
		if qualityResult := iv.ValidateQuality(fmt.Sprintf("%v", quality)); qualityResult.HasErrors() {
			result.Errors = append(result.Errors, qualityResult.Errors...)
		}
	}

	if priority, exists := options["priority"]; exists {
		if priorityResult := iv.ValidatePriority(priority); priorityResult.HasErrors() {
			result.Errors = append(result.Errors, priorityResult.Errors...)
		}
	}

	if retentionDays, exists := options["retention_days"]; exists {
		if retentionResult := iv.ValidateRetentionDays(retentionDays); retentionResult.HasErrors() {
			result.Errors = append(result.Errors, retentionResult.Errors...)
		}
	}

	if useCase, exists := options["use_case"]; exists {
		if useCaseResult := iv.ValidateUseCase(fmt.Sprintf("%v", useCase)); useCaseResult.HasErrors() {
			result.Errors = append(result.Errors, useCaseResult.Errors...)
		}
	}

	if autoCleanup, exists := options["auto_cleanup"]; exists {
		if autoCleanupResult := iv.ValidateAutoCleanup(autoCleanup); autoCleanupResult.HasErrors() {
			result.Errors = append(result.Errors, autoCleanupResult.Errors...)
		}
	}

	// Update overall validation result
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// SanitizeString removes potentially dangerous characters from strings
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
