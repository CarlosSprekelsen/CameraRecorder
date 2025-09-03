/*
MediaMTX Integration Error Types

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MediaMTXError represents MediaMTX-specific errors
type MediaMTXError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Op      string `json:"op,omitempty"`
	Err     error  `json:"-"`
}

func (e *MediaMTXError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("MediaMTX error [%d]: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("MediaMTX error [%d]: %s", e.Code, e.Message)
}

func (e *MediaMTXError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *MediaMTXError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*MediaMTXError); ok {
		// Compare by code and message for exact matches
		return e.Code == targetErr.Code && e.Message == targetErr.Message
	}
	
	// Check if this error wraps the target error
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}
	
	// Check if target is a predefined error constant (errors.errorString)
	// This allows errors.Is to work with predefined error constants
	targetStr := target.Error()
	eStr := e.Error()
	eMsg := e.Message
	
	// Compare error message content
	// Check if the target message is contained in our error message
	// This handles cases where predefined constants have different wording
	result := strings.Contains(eStr, targetStr) || strings.Contains(eMsg, targetStr) || 
		   strings.Contains(targetStr, eMsg) || strings.Contains(targetStr, eStr)
	
	// If direct string matching fails, try semantic matching
	// This handles cases where the predefined constants use different wording
	if !result {
		// Check if both contain "MediaMTX" and similar concepts
		if strings.Contains(targetStr, "MediaMTX") && strings.Contains(eStr, "MediaMTX") {
			// Check for semantic similarity in the error type
			if strings.Contains(targetStr, "not found") && strings.Contains(eStr, "Not Found") {
				result = true
			} else if strings.Contains(targetStr, "timeout") && strings.Contains(eStr, "timeout") {
				result = true
			} else if strings.Contains(targetStr, "unauthorized") && strings.Contains(eStr, "unauthorized") {
				result = true
			} else if strings.Contains(targetStr, "forbidden") && strings.Contains(eStr, "forbidden") {
				result = true
			} else if strings.Contains(targetStr, "conflict") && strings.Contains(eStr, "conflict") {
				result = true
			} else if strings.Contains(targetStr, "internal") && strings.Contains(eStr, "internal") {
				result = true
			} else if strings.Contains(targetStr, "server error") && strings.Contains(eStr, "Server Error") {
				result = true
			}
		}
	}
	
	return result
}

// CircuitBreakerError represents circuit breaker errors
type CircuitBreakerError struct {
	State   string `json:"state"`
	Message string `json:"message"`
	Op      string `json:"op,omitempty"`
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker %s: %s", e.State, e.Message)
}

// StreamError represents stream-specific errors
type StreamError struct {
	StreamID string `json:"stream_id"`
	Op       string `json:"op"`
	Message  string `json:"message"`
	Err      error  `json:"-"`
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("stream %s: %s: %s", e.StreamID, e.Op, e.Message)
}

func (e *StreamError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *StreamError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*StreamError); ok {
		// Compare by stream ID and operation for exact matches
		return e.StreamID == targetErr.StreamID && e.Op == targetErr.Op
	}

	// Check if this error wraps the target error
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}

	return false
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
		return errors.Is(e.Err, target)
	}

	return false
}

// RecordingError represents recording-specific errors
type RecordingError struct {
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
	Op        string `json:"op"`
	Message   string `json:"message"`
	Err       error  `json:"-"`
}

func (e *RecordingError) Error() string {
	return fmt.Sprintf("recording %s (device %s): %s: %s", e.SessionID, e.Device, e.Op, e.Message)
}

func (e *RecordingError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *RecordingError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*RecordingError); ok {
		// Compare by session ID and operation for exact matches
		return e.SessionID == targetErr.SessionID && e.Op == targetErr.Op
	}
	
	// Check if this error wraps the target error
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}
	
	return false
}

// FFmpegError represents FFmpeg process errors
type FFmpegError struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	Op      string `json:"op"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Enhanced error categorization (Phase 4 enhancement)
type ErrorCategory string

const (
	ErrorCategorySystem     ErrorCategory = "SYSTEM"
	ErrorCategoryNetwork    ErrorCategory = "NETWORK"
	ErrorCategoryResource   ErrorCategory = "RESOURCE"
	ErrorCategoryValidation ErrorCategory = "VALIDATION"
	ErrorCategorySecurity   ErrorCategory = "SECURITY"
	ErrorCategoryTimeout    ErrorCategory = "TIMEOUT"
	ErrorCategoryRecovery   ErrorCategory = "RECOVERY"
)

// ErrorSeverity represents error severity levels
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorContext represents additional error context
type ErrorContext struct {
	Category    ErrorCategory          `json:"category"`
	Severity    ErrorSeverity          `json:"severity"`
	Retryable   bool                   `json:"retryable"`
	Recoverable bool                   `json:"recoverable"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   string                 `json:"timestamp"`
	TraceID     string                 `json:"trace_id,omitempty"`
}

// EnhancedError represents an enhanced error with categorization and context
type EnhancedError struct {
	BaseError   error        `json:"base_error"`
	Context     ErrorContext `json:"context"`
	RecoveryOps []string     `json:"recovery_ops,omitempty"`
}

func (e *EnhancedError) Error() string {
	return fmt.Sprintf("enhanced error [%s/%s]: %v", e.Context.Category, e.Context.Severity, e.BaseError)
}

func (e *EnhancedError) Unwrap() error {
	return e.BaseError
}

func (e *EnhancedError) IsRetryable() bool {
	return e.Context.Retryable
}

func (e *EnhancedError) IsRecoverable() bool {
	return e.Context.Recoverable
}

func (e *EnhancedError) GetCategory() ErrorCategory {
	return e.Context.Category
}

func (e *EnhancedError) GetSeverity() ErrorSeverity {
	return e.Context.Severity
}

func (e *FFmpegError) Error() string {
	return fmt.Sprintf("FFmpeg process %d (%s): %s: %s", e.PID, e.Command, e.Op, e.Message)
}

func (e *FFmpegError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *FFmpegError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*FFmpegError); ok {
		// Compare by PID and operation for exact matches
		return e.PID == targetErr.PID && e.Op == targetErr.Op
	}
	
	// Check if this error wraps the target error
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}
	
	return false
}

// ConfigurationError represents configuration errors
type ConfigurationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error for %s=%s: %s", e.Field, e.Value, e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for comparing with predefined error constants
func (e *ConfigurationError) Is(target error) bool {
	// Check if target is the same type
	if targetErr, ok := target.(*ConfigurationError); ok {
		// Compare by field and value for exact matches
		return e.Field == targetErr.Field && e.Value == targetErr.Value
	}
	
	// Check if this error wraps the target error
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}
	
	return false
}

// Predefined error constants
var (
	// MediaMTX service errors
	ErrMediaMTXUnavailable     = errors.New("MediaMTX service unavailable")
	ErrMediaMTXTimeout         = errors.New("MediaMTX service timeout")
	ErrMediaMTXInvalidResponse = errors.New("MediaMTX invalid response")
	ErrMediaMTXUnauthorized    = errors.New("MediaMTX unauthorized access")
	ErrMediaMTXForbidden       = errors.New("MediaMTX forbidden access")
	ErrMediaMTXNotFound        = errors.New("MediaMTX resource not found")
	ErrMediaMTXConflict        = errors.New("MediaMTX resource conflict")
	ErrMediaMTXInternal        = errors.New("MediaMTX internal server error")

	// Circuit breaker errors
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open")
	ErrCircuitTimeout  = errors.New("circuit breaker timeout")

	// Stream errors
	ErrStreamNotFound    = errors.New("stream not found")
	ErrStreamExists      = errors.New("stream already exists")
	ErrStreamInvalid     = errors.New("invalid stream configuration")
	ErrStreamUnavailable = errors.New("stream unavailable")
	ErrStreamBusy        = errors.New("stream is busy")

	// Path errors
	ErrPathNotFound    = errors.New("path not found")
	ErrPathExists      = errors.New("path already exists")
	ErrPathInvalid     = errors.New("invalid path configuration")
	ErrPathUnavailable = errors.New("path unavailable")
	ErrPathBusy        = errors.New("path is busy")

	// Recording errors
	ErrRecordingNotFound    = errors.New("recording session not found")
	ErrRecordingExists      = errors.New("recording session already exists")
	ErrRecordingInvalid     = errors.New("invalid recording configuration")
	ErrRecordingUnavailable = errors.New("recording unavailable")
	ErrRecordingBusy        = errors.New("recording is busy")
	ErrRecordingFailed      = errors.New("recording failed")

	// FFmpeg errors
	ErrFFmpegNotFound       = errors.New("FFmpeg not found")
	ErrFFmpegProcessFailed  = errors.New("FFmpeg process failed")
	ErrFFmpegTimeout        = errors.New("FFmpeg process timeout")
	ErrFFmpegInvalidCommand = errors.New("FFmpeg invalid command")
	ErrFFmpegOutputError    = errors.New("FFmpeg output error")

	// Configuration errors
	ErrConfigInvalid     = errors.New("invalid configuration")
	ErrConfigMissing     = errors.New("missing configuration")
	ErrConfigUnsupported = errors.New("unsupported configuration")
)

// NewMediaMTXError creates a new MediaMTX error
func NewMediaMTXError(code int, message, details string) *MediaMTXError {
	return &MediaMTXError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewMediaMTXErrorWithOp creates a new MediaMTX error with operation context
func NewMediaMTXErrorWithOp(code int, message, details, op string) *MediaMTXError {
	return &MediaMTXError{
		Code:    code,
		Message: message,
		Details: details,
		Op:      op,
	}
}

// NewMediaMTXErrorFromHTTP creates a MediaMTX error from HTTP response
func NewMediaMTXErrorFromHTTP(statusCode int, body []byte) *MediaMTXError {
	message := "unknown error"
	details := string(body)

	switch statusCode {
	case http.StatusUnauthorized:
		message = "unauthorized access"
	case http.StatusForbidden:
		message = "forbidden access"
	case http.StatusNotFound:
		message = "resource not found"
	case http.StatusConflict:
		message = "resource conflict"
	case http.StatusInternalServerError:
		message = "internal server error"
	case http.StatusBadGateway:
		message = "bad gateway"
	case http.StatusServiceUnavailable:
		message = "service unavailable"
	case http.StatusGatewayTimeout:
		message = "gateway timeout"
	}

	return &MediaMTXError{
		Code:    statusCode,
		Message: message,
		Details: details,
	}
}

// NewStreamError creates a new stream error
func NewStreamError(streamID, op, message string) *StreamError {
	return &StreamError{
		StreamID: streamID,
		Op:       op,
		Message:  message,
	}
}

// NewStreamErrorWithErr creates a new stream error with underlying error
func NewStreamErrorWithErr(streamID, op, message string, err error) *StreamError {
	return &StreamError{
		StreamID: streamID,
		Op:       op,
		Message:  message,
		Err:      err,
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

// NewRecordingError creates a new recording error
func NewRecordingError(sessionID, device, op, message string) *RecordingError {
	return &RecordingError{
		SessionID: sessionID,
		Device:    device,
		Op:        op,
		Message:   message,
	}
}

// NewRecordingErrorWithErr creates a new recording error with underlying error
func NewRecordingErrorWithErr(sessionID, device, op, message string, err error) *RecordingError {
	return &RecordingError{
		SessionID: sessionID,
		Device:    device,
		Op:        op,
		Message:   message,
		Err:       err,
	}
}

// NewFFmpegError creates a new FFmpeg error
func NewFFmpegError(pid int, command, op, message string) *FFmpegError {
	return &FFmpegError{
		PID:     pid,
		Command: command,
		Op:      op,
		Message: message,
	}
}

// NewFFmpegErrorWithErr creates a new FFmpeg error with underlying error
func NewFFmpegErrorWithErr(pid int, command, op, message string, err error) *FFmpegError {
	return &FFmpegError{
		PID:     pid,
		Command: command,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// Enhanced error handling functions (Phase 4 enhancement)

// NewEnhancedError creates a new enhanced error with categorization
func NewEnhancedError(baseError error, category ErrorCategory, severity ErrorSeverity, retryable, recoverable bool) *EnhancedError {
	return &EnhancedError{
		BaseError: baseError,
		Context: ErrorContext{
			Category:    category,
			Severity:    severity,
			Retryable:   retryable,
			Recoverable: recoverable,
			Timestamp:   time.Now().Format(time.RFC3339),
		},
	}
}

// CategorizeError automatically categorizes errors based on their type and content
func CategorizeError(err error) *EnhancedError {
	if err == nil {
		return nil
	}

	// Check for specific error types
	switch {
	case errors.Is(err, ErrMediaMTXUnavailable):
		return NewEnhancedError(err, ErrorCategoryNetwork, ErrorSeverityHigh, true, true)
	case errors.Is(err, ErrMediaMTXTimeout):
		return NewEnhancedError(err, ErrorCategoryTimeout, ErrorSeverityMedium, true, true)
	case errors.Is(err, ErrMediaMTXUnauthorized):
		return NewEnhancedError(err, ErrorCategorySecurity, ErrorSeverityHigh, false, true)
	case errors.Is(err, ErrMediaMTXForbidden):
		return NewEnhancedError(err, ErrorCategorySecurity, ErrorSeverityHigh, false, true)
	case errors.Is(err, ErrStreamNotFound):
		return NewEnhancedError(err, ErrorCategoryResource, ErrorSeverityMedium, false, true)
	case errors.Is(err, ErrRecordingFailed):
		return NewEnhancedError(err, ErrorCategorySystem, ErrorSeverityHigh, true, true)
	case errors.Is(err, ErrFFmpegProcessFailed):
		return NewEnhancedError(err, ErrorCategorySystem, ErrorSeverityHigh, true, true)
	case errors.Is(err, ErrCircuitOpen):
		return NewEnhancedError(err, ErrorCategoryRecovery, ErrorSeverityMedium, true, true)
	default:
		// Default categorization based on error message
		errStr := err.Error()
		switch {
		case contains(errStr, "timeout"):
			return NewEnhancedError(err, ErrorCategoryTimeout, ErrorSeverityMedium, true, true)
		case contains(errStr, "not found"):
			return NewEnhancedError(err, ErrorCategoryResource, ErrorSeverityMedium, false, true)
		case contains(errStr, "permission"):
			return NewEnhancedError(err, ErrorCategorySecurity, ErrorSeverityHigh, false, true)
		case contains(errStr, "network"):
			return NewEnhancedError(err, ErrorCategoryNetwork, ErrorSeverityHigh, true, true)
		case contains(errStr, "invalid"):
			return NewEnhancedError(err, ErrorCategoryValidation, ErrorSeverityMedium, false, true)
		default:
			return NewEnhancedError(err, ErrorCategorySystem, ErrorSeverityMedium, false, true)
		}
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// GetRecoveryStrategies returns recovery strategies for different error categories
func GetRecoveryStrategies(category ErrorCategory) []string {
	switch category {
	case ErrorCategoryNetwork:
		return []string{"retry_with_backoff", "check_connectivity", "restart_service"}
	case ErrorCategoryTimeout:
		return []string{"increase_timeout", "retry_with_backoff", "check_system_load"}
	case ErrorCategoryResource:
		return []string{"cleanup_resources", "restart_service", "check_disk_space"}
	case ErrorCategorySystem:
		return []string{"restart_service", "check_logs", "restart_ffmpeg"}
	case ErrorCategorySecurity:
		return []string{"check_credentials", "verify_permissions", "contact_admin"}
	case ErrorCategoryValidation:
		return []string{"validate_input", "check_configuration", "update_parameters"}
	case ErrorCategoryRecovery:
		return []string{"wait_for_recovery", "reset_circuit_breaker", "restart_service"}
	default:
		return []string{"check_logs", "restart_service"}
	}
}

// ShouldRetry determines if an error should be retried
func ShouldRetry(err error) bool {
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr.IsRetryable()
	}

	// Check for specific retryable errors
	return errors.Is(err, ErrMediaMTXTimeout) ||
		errors.Is(err, ErrMediaMTXUnavailable) ||
		errors.Is(err, ErrRecordingFailed) ||
		errors.Is(err, ErrFFmpegProcessFailed) ||
		errors.Is(err, ErrCircuitOpen)
}

// GetErrorMetadata extracts metadata from an error for logging and monitoring
func GetErrorMetadata(err error) map[string]interface{} {
	metadata := make(map[string]interface{})

	if enhancedErr, ok := err.(*EnhancedError); ok {
		metadata["category"] = enhancedErr.GetCategory()
		metadata["severity"] = enhancedErr.GetSeverity()
		metadata["retryable"] = enhancedErr.IsRetryable()
		metadata["recoverable"] = enhancedErr.IsRecoverable()
		metadata["recovery_ops"] = enhancedErr.RecoveryOps
		metadata["timestamp"] = enhancedErr.Context.Timestamp
		if enhancedErr.Context.TraceID != "" {
			metadata["trace_id"] = enhancedErr.Context.TraceID
		}
	}

	// Add error type information
	metadata["error_type"] = fmt.Sprintf("%T", err)
	metadata["error_message"] = err.Error()

	return metadata
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(field, value, message string) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewConfigurationErrorWithErr creates a new configuration error with underlying error
func NewConfigurationErrorWithErr(field, value, message string, err error) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// IsMediaMTXError checks if an error is a MediaMTX error
func IsMediaMTXError(err error) bool {
	var mediaMTXErr *MediaMTXError
	return errors.As(err, &mediaMTXErr)
}

// IsStreamError checks if an error is a stream error
func IsStreamError(err error) bool {
	var streamErr *StreamError
	return errors.As(err, &streamErr)
}

// IsPathError checks if an error is a path error
func IsPathError(err error) bool {
	var pathErr *PathError
	return errors.As(err, &pathErr)
}

// IsRecordingError checks if an error is a recording error
func IsRecordingError(err error) bool {
	var recordingErr *RecordingError
	return errors.As(err, &recordingErr)
}

// IsFFmpegError checks if an error is an FFmpeg error
func IsFFmpegError(err error) bool {
	var ffmpegErr *FFmpegError
	return errors.As(err, &ffmpegErr)
}

// IsConfigurationError checks if an error is a configuration error
func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

// IsCircuitBreakerError checks if an error is a circuit breaker error
func IsCircuitBreakerError(err error) bool {
	var circuitErr *CircuitBreakerError
	return errors.As(err, &circuitErr)
}
