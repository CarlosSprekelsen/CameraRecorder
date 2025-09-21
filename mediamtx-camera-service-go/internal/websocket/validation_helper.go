package websocket

import (
	"strconv"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
)

// ValidationHelper provides centralized validation for JSON-RPC method parameters
type ValidationHelper struct {
	inputValidator *security.InputValidator
	logger         *logging.Logger
}

// NewValidationHelper creates a new validation helper
func NewValidationHelper(inputValidator *security.InputValidator, logger *logging.Logger) *ValidationHelper {
	return &ValidationHelper{
		inputValidator: inputValidator,
		logger:         logger,
	}
}

// ValidationResult contains validation results with error details
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Data     map[string]interface{}
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
		Data:     make(map[string]interface{}),
	}
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(message string) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, message)
}

// AddWarning adds a validation warning
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// AddData adds data to the validation result
func (vr *ValidationResult) AddData(key string, value interface{}) {
	vr.Data[key] = value
}

// GetFirstError returns the first error message or empty string
func (vr *ValidationResult) GetFirstError() string {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return ""
}

// ValidatePaginationParams validates limit and offset parameters
func (vh *ValidationHelper) ValidatePaginationParams(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		// Use defaults for nil params
		result.AddData("limit", 100)
		result.AddData("offset", 0)
		return result
	}

	// Validate limit
	limit := 100 // Default
	if limitVal, exists := params["limit"]; exists {
		if limitResult := vh.inputValidator.ValidateLimit(limitVal); limitResult.HasErrors() {
			result.AddError(limitResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated limit
		switch v := limitVal.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		case string:
			if limitInt, err := strconv.Atoi(v); err == nil {
				limit = limitInt
			}
		}
	}

	// Validate offset
	offset := 0 // Default
	if offsetVal, exists := params["offset"]; exists {
		if offsetResult := vh.inputValidator.ValidateOffset(offsetVal); offsetResult.HasErrors() {
			result.AddError(offsetResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated offset
		switch v := offsetVal.(type) {
		case int:
			offset = v
		case float64:
			offset = int(v)
		case string:
			if offsetInt, err := strconv.Atoi(v); err == nil {
				offset = offsetInt
			}
		}
	}

	result.AddData("limit", limit)
	result.AddData("offset", offset)
	return result
}

// ValidateDeviceParameter validates the device parameter
func (vh *ValidationHelper) ValidateDeviceParameter(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("device parameter is required")
		return result
	}

	deviceVal, exists := params["device"]
	if !exists {
		result.AddError("device parameter is required")
		return result
	}

	if deviceResult := vh.inputValidator.ValidateDevicePath(deviceVal); deviceResult.HasErrors() {
		result.AddError(deviceResult.GetErrorMessages()[0])
		return result
	}

	// Extract validated device path
	devicePath, ok := deviceVal.(string)
	if !ok {
		result.AddError("device parameter must be a string")
		return result
	}

	result.AddData("device", devicePath)
	return result
}

// ValidateFilenameParameter validates the filename parameter
func (vh *ValidationHelper) ValidateFilenameParameter(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("filename parameter is required")
		return result
	}

	filenameVal, exists := params["filename"]
	if !exists {
		result.AddError("filename parameter is required")
		return result
	}

	if filenameResult := vh.inputValidator.ValidateFilename(filenameVal); filenameResult.HasErrors() {
		result.AddError(filenameResult.GetErrorMessages()[0])
		return result
	}

	// Extract validated filename
	filename, ok := filenameVal.(string)
	if !ok {
		result.AddError("filename parameter must be a string")
		return result
	}

	result.AddData("filename", filename)
	return result
}

// ValidateRecordingParameters validates recording-specific parameters
func (vh *ValidationHelper) ValidateRecordingParameters(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("parameters are required")
		return result
	}

	// Validate device parameter first
	deviceResult := vh.ValidateDeviceParameter(params)
	if !deviceResult.Valid {
		result.AddError(deviceResult.GetFirstError())
		return result
	}
	result.AddData("device", deviceResult.Data["device"])

	// Validate optional parameters
	options := make(map[string]interface{})

	// Duration validation
	if duration, exists := params["duration"]; exists {
		if durationResult := vh.inputValidator.ValidatePositiveInteger(duration, "duration"); durationResult.HasErrors() {
			result.AddError(durationResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated duration
		switch v := duration.(type) {
		case int:
			options["max_duration"] = v
		case float64:
			options["max_duration"] = int(v)
		case string:
			if durationInt, err := strconv.Atoi(v); err == nil {
				options["max_duration"] = durationInt
			}
		}
	}

	// Format validation
	if format, exists := params["format"]; exists {
		if formatResult := vh.inputValidator.ValidateOptionalString(format, "format"); formatResult.HasErrors() {
			result.AddError(formatResult.GetErrorMessages()[0])
			return result
		}
		if formatStr, ok := format.(string); ok && formatStr != "" {
			options["output_format"] = formatStr
		}
	}

	// Codec validation
	if codec, exists := params["codec"]; exists {
		if codecResult := vh.inputValidator.ValidateOptionalString(codec, "codec"); codecResult.HasErrors() {
			result.AddError(codecResult.GetErrorMessages()[0])
			return result
		}
		if codecStr, ok := codec.(string); ok && codecStr != "" {
			options["codec"] = codecStr
		}
	}

	// Quality validation
	if quality, exists := params["quality"]; exists {
		if qualityResult := vh.inputValidator.ValidatePositiveInteger(quality, "quality"); qualityResult.HasErrors() {
			result.AddError(qualityResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated quality
		switch v := quality.(type) {
		case int:
			options["crf"] = v
		case float64:
			options["crf"] = int(v)
		case string:
			if qualityInt, err := strconv.Atoi(v); err == nil {
				options["crf"] = qualityInt
			}
		}
	}

	// Use case validation
	if useCase, exists := params["use_case"]; exists {
		if useCaseResult := vh.inputValidator.ValidateOptionalString(useCase, "use_case"); useCaseResult.HasErrors() {
			result.AddError(useCaseResult.GetErrorMessages()[0])
			return result
		}
		if useCaseStr, ok := useCase.(string); ok && useCaseStr != "" {
			options["use_case"] = useCaseStr
		}
	}

	// Priority validation
	if priority, exists := params["priority"]; exists {
		if priorityResult := vh.inputValidator.ValidatePositiveInteger(priority, "priority"); priorityResult.HasErrors() {
			result.AddError(priorityResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated priority
		switch v := priority.(type) {
		case int:
			options["priority"] = v
		case float64:
			options["priority"] = int(v)
		case string:
			if priorityInt, err := strconv.Atoi(v); err == nil {
				options["priority"] = priorityInt
			}
		}
	}

	// Auto cleanup validation
	if autoCleanup, exists := params["auto_cleanup"]; exists {
		if autoCleanupResult := vh.inputValidator.ValidateBooleanParameter(autoCleanup, "auto_cleanup"); autoCleanupResult.HasErrors() {
			result.AddError(autoCleanupResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated auto cleanup
		switch v := autoCleanup.(type) {
		case bool:
			options["auto_cleanup"] = v
		case int:
			options["auto_cleanup"] = v == 1
		case float64:
			options["auto_cleanup"] = v == 1.0
		case string:
			options["auto_cleanup"] = v == "true"
		}
	}

	// Retention days validation
	if retentionDays, exists := params["retention_days"]; exists {
		if retentionResult := vh.inputValidator.ValidatePositiveInteger(retentionDays, "retention_days"); retentionResult.HasErrors() {
			result.AddError(retentionResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated retention days
		switch v := retentionDays.(type) {
		case int:
			options["retention_days"] = v
		case float64:
			options["retention_days"] = int(v)
		case string:
			if retentionInt, err := strconv.Atoi(v); err == nil {
				options["retention_days"] = retentionInt
			}
		}
	}

	result.AddData("options", options)
	return result
}

// ValidateSnapshotParameters validates snapshot-specific parameters
func (vh *ValidationHelper) ValidateSnapshotParameters(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("parameters are required")
		return result
	}

	// Validate device parameter first
	deviceResult := vh.ValidateDeviceParameter(params)
	if !deviceResult.Valid {
		result.AddError(deviceResult.GetFirstError())
		return result
	}
	result.AddData("device", deviceResult.Data["device"])

	// Validate optional parameters
	options := make(map[string]interface{})

	// Filename validation
	if filename, exists := params["filename"]; exists {
		if filenameResult := vh.inputValidator.ValidateOptionalString(filename, "filename"); filenameResult.HasErrors() {
			result.AddError(filenameResult.GetErrorMessages()[0])
			return result
		}
		if filenameStr, ok := filename.(string); ok && filenameStr != "" {
			options["filename"] = filenameStr
		}
	}

	// Format validation
	if format, exists := params["format"]; exists {
		if formatResult := vh.inputValidator.ValidateOptionalString(format, "format"); formatResult.HasErrors() {
			result.AddError(formatResult.GetErrorMessages()[0])
			return result
		}
		if formatStr, ok := format.(string); ok && formatStr != "" {
			options["format"] = formatStr
		}
	}

	// Quality validation
	if quality, exists := params["quality"]; exists {
		if qualityResult := vh.inputValidator.ValidatePositiveInteger(quality, "quality"); qualityResult.HasErrors() {
			result.AddError(qualityResult.GetErrorMessages()[0])
			return result
		}
		// Extract validated quality
		switch v := quality.(type) {
		case int:
			options["quality"] = v
		case float64:
			options["quality"] = int(v)
		case string:
			if qualityInt, err := strconv.Atoi(v); err == nil {
				options["quality"] = qualityInt
			}
		}
	}

	result.AddData("options", options)
	return result
}

// ValidateRetentionPolicyParameters validates retention policy parameters
func (vh *ValidationHelper) ValidateRetentionPolicyParameters(params map[string]interface{}) *ValidationResult {
	result := NewValidationResult()

	if params == nil {
		result.AddError("parameters are required")
		return result
	}

	// Validate policy type
	policyTypeVal, exists := params["policy_type"]
	if !exists {
		result.AddError("policy_type parameter is required")
		return result
	}

	policyType, ok := policyTypeVal.(string)
	if !ok {
		result.AddError("policy_type parameter must be a string")
		return result
	}

	if policyType != "age" && policyType != "size" {
		result.AddError("policy_type must be either 'age' or 'size'")
		return result
	}

	// Validate enabled flag
	enabledVal, exists := params["enabled"]
	if !exists {
		result.AddError("enabled parameter is required")
		return result
	}

	var enabled bool
	switch v := enabledVal.(type) {
	case bool:
		enabled = v
	case int:
		enabled = v == 1
	case float64:
		enabled = v == 1.0
	case string:
		enabled = v == "true"
	default:
		result.AddError("enabled parameter must be a boolean")
		return result
	}

	// Validate policy-specific parameters
	switch policyType {
	case "age":
		if maxAgeDays, exists := params["max_age_days"]; exists {
			if maxAgeResult := vh.inputValidator.ValidatePositiveInteger(maxAgeDays, "max_age_days"); maxAgeResult.HasErrors() {
				result.AddError(maxAgeResult.GetErrorMessages()[0])
				return result
			}
		}
	case "size":
		if maxSizeGB, exists := params["max_size_gb"]; exists {
			if maxSizeResult := vh.inputValidator.ValidatePositiveInteger(maxSizeGB, "max_size_gb"); maxSizeResult.HasErrors() {
				result.AddError(maxSizeResult.GetErrorMessages()[0])
				return result
			}
		}
	}

	result.AddData("policy_type", policyType)
	result.AddData("enabled", enabled)
	return result
}

// CreateValidationErrorResponse creates a JSON-RPC error response from validation results
func (vh *ValidationHelper) CreateValidationErrorResponse(validationResult *ValidationResult) *JsonRpcResponse {
	reason := "validation_failed"
	details := ""
	if len(validationResult.Errors) > 0 {
		details = validationResult.Errors[0]
	}

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Error:   NewJsonRpcError(INVALID_PARAMS, reason, details, "Check parameter types and values"),
	}
}

// LogValidationWarnings logs validation warnings for debugging
func (vh *ValidationHelper) LogValidationWarnings(validationResult *ValidationResult, method string, clientID string) {
	if len(validationResult.Warnings) > 0 {
		vh.logger.WithFields(logging.Fields{
			"client_id": clientID,
			"method":    method,
			"warnings":  validationResult.Warnings,
		}).Warn("Validation warnings during parameter processing")
	}
}
