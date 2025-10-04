package constants

import "time"

// =============================================================================
// JSON-RPC ERROR CODES (Ground Truth from API Documentation)
// =============================================================================
// These constants define the official JSON-RPC error codes as documented
// in docs/api/json_rpc_methods.md and must be used consistently across
// all implementation and test code.

const (
	// Standard JSON-RPC 2.0 Error Codes (RFC 4627)
	JSONRPC_INVALID_REQUEST  = -32600
	JSONRPC_METHOD_NOT_FOUND = -32601
	JSONRPC_INVALID_PARAMS   = -32602
	JSONRPC_INTERNAL_ERROR   = -32603

	// Service-Specific Error Codes (API Documentation)
	API_AUTHENTICATION_REQUIRED = -32001 // Auth Failed (invalid/expired token)
	API_PERMISSION_DENIED       = -32002 // Permission Denied (role lacks permission)
	API_INVALID_STATE           = -32020 // Invalid State (operation not allowed in current state)
	API_UNSUPPORTED             = -32030 // Unsupported (feature/capability not available)
	API_RATE_LIMIT_EXCEEDED     = -32040 // Rate Limited (too many requests)
	API_DEPENDENCY_FAILED       = -32050 // Dependency Failed (MediaMTX/FFmpeg error)
	API_NOT_FOUND               = -32010 // Not Found (recording/file/camera not found)

	// Legacy constants for backward compatibility (deprecated)
	API_INSUFFICIENT_PERMISSIONS = API_PERMISSION_DENIED
	API_CAMERA_NOT_FOUND         = API_NOT_FOUND
	API_RECORDING_IN_PROGRESS    = API_INVALID_STATE
	API_MEDIAMTX_UNAVAILABLE     = API_DEPENDENCY_FAILED
	API_STREAM_NOT_FOUND         = API_NOT_FOUND
	API_FILE_NOT_FOUND           = API_NOT_FOUND

	// Enhanced Recording Management Error Codes
	ERROR_CAMERA_NOT_FOUND         = -1000
	ERROR_CAMERA_NOT_AVAILABLE     = -1001
	ERROR_RECORDING_IN_PROGRESS    = -1002
	ERROR_MEDIAMTX_ERROR           = -1003
	ERROR_CAMERA_ALREADY_RECORDING = -1006
	ERROR_STORAGE_LOW              = -1008
	ERROR_STORAGE_CRITICAL         = -1010
)

// =============================================================================
// WEBSOCKET SERVER CONSTANTS (Shared Implementation/Test Values)
// =============================================================================
// These constants define standard WebSocket server configuration values
// used by both production code and test code for consistency.

const (
	// Connection Timeouts
	WEBSOCKET_READ_TIMEOUT           = 5 * time.Second
	WEBSOCKET_WRITE_TIMEOUT          = 1 * time.Second
	WEBSOCKET_PING_INTERVAL          = 30 * time.Second
	WEBSOCKET_PONG_WAIT              = 60 * time.Second
	WEBSOCKET_SHUTDOWN_TIMEOUT       = 30 * time.Second
	WEBSOCKET_CLIENT_CLEANUP_TIMEOUT = 10 * time.Second

	// Buffer and Message Sizes
	WEBSOCKET_MAX_MESSAGE_SIZE  = 1024 * 1024 // 1MB
	WEBSOCKET_READ_BUFFER_SIZE  = 1024
	WEBSOCKET_WRITE_BUFFER_SIZE = 1024
	WEBSOCKET_TEST_BUFFER_SIZE  = 4096 // Larger buffer for tests

	// Connection Limits
	WEBSOCKET_MAX_CONNECTIONS_PRODUCTION = 1000
	WEBSOCKET_MAX_CONNECTIONS_TEST       = 100

	// Default Server Configuration
	WEBSOCKET_DEFAULT_HOST = "0.0.0.0"
	WEBSOCKET_DEFAULT_PORT = 8002
	WEBSOCKET_DEFAULT_PATH = "/ws"
)

// =============================================================================
// API RESPONSE VALUES (Ground Truth from API Documentation)
// =============================================================================
// These constants define the exact string values that must be returned
// in API responses according to the API documentation.

const (
	// JSON-RPC Protocol
	JSONRPC_VERSION = "2.0"

	// Camera Status Values (from API documentation)
	CAMERA_STATUS_CONNECTED    = "CONNECTED"
	CAMERA_STATUS_DISCONNECTED = "DISCONNECTED"
	CAMERA_STATUS_ERROR        = "ERROR"

	// Recording Status Values (from API documentation)
	RECORDING_STATUS_RECORDING = "RECORDING"
	RECORDING_STATUS_STOPPED   = "STOPPED"
	RECORDING_STATUS_FAILED    = "FAILED"

	// Streaming Status Values (standardized to UPPERCASE)
	STREAMING_STATUS_STARTED  = "STARTED"
	STREAMING_STATUS_STOPPED  = "STOPPED"
	STREAMING_STATUS_FAILED   = "FAILED"
	STREAMING_STATUS_ACTIVE   = "ACTIVE"
	STREAMING_STATUS_INACTIVE = "INACTIVE"
	STREAMING_STATUS_STARTING = "STARTING"
	STREAMING_STATUS_STOPPING = "STOPPING"

	// System Status Values (standardized to UPPERCASE)
	SYSTEM_STATUS_HEALTHY   = "HEALTHY"
	SYSTEM_STATUS_DEGRADED  = "DEGRADED"
	SYSTEM_STATUS_UNHEALTHY = "UNHEALTHY"

	// Component Status Values (standardized to UPPERCASE)
	COMPONENT_STATUS_RUNNING  = "RUNNING"
	COMPONENT_STATUS_STOPPED  = "STOPPED"
	COMPONENT_STATUS_ERROR    = "ERROR"
	COMPONENT_STATUS_STARTING = "STARTING"
	COMPONENT_STATUS_STOPPING = "STOPPING"

	// Snapshot Status Values (standardized to UPPERCASE)
	SNAPSHOT_STATUS_COMPLETED = "COMPLETED"
	SNAPSHOT_STATUS_SUCCESS   = "SUCCESS"
	SNAPSHOT_STATUS_FAILED    = "FAILED"

	// Validation Status Values (standardized to UPPERCASE)
	VALIDATION_STATUS_NONE         = "NONE"
	VALIDATION_STATUS_DISCONNECTED = "DISCONNECTED"
	VALIDATION_STATUS_CONFIRMED    = "CONFIRMED"
)

// =============================================================================
// API FORMAT CONSTANTS (Ground Truth from API Documentation)
// =============================================================================
// These constants define valid format values for API parameters.

const (
	// Recording Formats (from API documentation)
	RECORDING_FORMAT_FMP4 = "fmp4"
	RECORDING_FORMAT_MP4  = "mp4"
	RECORDING_FORMAT_MKV  = "mkv"

	// Snapshot Formats (from API documentation)
	SNAPSHOT_FORMAT_JPEG = "jpeg"
	SNAPSHOT_FORMAT_JPG  = "jpg"

	// Stream Protocols (from API documentation)
	STREAM_PROTOCOL_RTSP   = "rtsp"
	STREAM_PROTOCOL_WEBRTC = "webrtc"
	STREAM_PROTOCOL_HLS    = "hls"
)

// =============================================================================
// ERROR MESSAGES (Ground Truth from API Documentation)
// =============================================================================
// Standard error messages that match the API documentation exactly.

var APIErrorMessages = map[int]string{
	JSONRPC_INVALID_REQUEST:     "Invalid Request",
	JSONRPC_METHOD_NOT_FOUND:    "Method not found",
	JSONRPC_INVALID_PARAMS:      "Invalid parameters",
	JSONRPC_INTERNAL_ERROR:      "Internal server error",
	API_AUTHENTICATION_REQUIRED: "Authentication failed or token expired",
	API_PERMISSION_DENIED:       "Permission denied",
	API_INVALID_STATE:           "Invalid state",
	API_UNSUPPORTED:             "Unsupported",
	API_RATE_LIMIT_EXCEEDED:     "Rate limited",
	API_DEPENDENCY_FAILED:       "Dependency failed",
	API_NOT_FOUND:               "Not found",

	// Enhanced Recording Management Error Codes
	ERROR_CAMERA_NOT_FOUND:         "Camera not found",
	ERROR_CAMERA_NOT_AVAILABLE:     "Camera not available",
	ERROR_RECORDING_IN_PROGRESS:    "Recording in progress",
	ERROR_MEDIAMTX_ERROR:           "MediaMTX error",
	ERROR_CAMERA_ALREADY_RECORDING: "Camera is currently recording",
	ERROR_STORAGE_LOW:              "Storage space is low",
	ERROR_STORAGE_CRITICAL:         "Storage space is critical",
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// GetAPIErrorMessage returns the standard error message for an error code
func GetAPIErrorMessage(code int) string {
	if message, exists := APIErrorMessages[code]; exists {
		return message
	}
	return "Unknown error"
}

// IsValidCameraStatus checks if a camera status value is valid per API documentation
func IsValidCameraStatus(status string) bool {
	validStatuses := []string{CAMERA_STATUS_CONNECTED, CAMERA_STATUS_DISCONNECTED, CAMERA_STATUS_ERROR}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidRecordingFormat checks if a recording format is valid per API documentation
func IsValidRecordingFormat(format string) bool {
	validFormats := []string{RECORDING_FORMAT_FMP4, RECORDING_FORMAT_MP4, RECORDING_FORMAT_MKV}
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}
