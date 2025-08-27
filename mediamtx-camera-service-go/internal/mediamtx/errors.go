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

// FFmpegError represents FFmpeg process errors
type FFmpegError struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	Op      string `json:"op"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *FFmpegError) Error() string {
	return fmt.Sprintf("FFmpeg process %d (%s): %s: %s", e.PID, e.Command, e.Op, e.Message)
}

func (e *FFmpegError) Unwrap() error {
	return e.Err
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

// Predefined error constants
var (
	// MediaMTX service errors
	ErrMediaMTXUnavailable = errors.New("MediaMTX service unavailable")
	ErrMediaMTXTimeout     = errors.New("MediaMTX service timeout")
	ErrMediaMTXInvalidResponse = errors.New("MediaMTX invalid response")
	ErrMediaMTXUnauthorized = errors.New("MediaMTX unauthorized access")
	ErrMediaMTXForbidden    = errors.New("MediaMTX forbidden access")
	ErrMediaMTXNotFound     = errors.New("MediaMTX resource not found")
	ErrMediaMTXConflict     = errors.New("MediaMTX resource conflict")
	ErrMediaMTXInternal     = errors.New("MediaMTX internal server error")

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
	ErrFFmpegNotFound    = errors.New("FFmpeg not found")
	ErrFFmpegProcessFailed = errors.New("FFmpeg process failed")
	ErrFFmpegTimeout      = errors.New("FFmpeg process timeout")
	ErrFFmpegInvalidCommand = errors.New("FFmpeg invalid command")
	ErrFFmpegOutputError  = errors.New("FFmpeg output error")

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
