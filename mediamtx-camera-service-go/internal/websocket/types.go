package websocket

import (
	"encoding/json"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/constants"
	"github.com/gorilla/websocket"
)

// JSON-RPC Error Codes - Using common constants for consistency
const (
	// Standard JSON-RPC 2.0 Error Codes
	INVALID_REQUEST  = constants.JSONRPC_INVALID_REQUEST
	METHOD_NOT_FOUND = constants.JSONRPC_METHOD_NOT_FOUND
	INVALID_PARAMS   = constants.JSONRPC_INVALID_PARAMS
	INTERNAL_ERROR   = constants.JSONRPC_INTERNAL_ERROR

	// Service-Specific Error Codes (aligned with API documentation)
	AUTHENTICATION_REQUIRED = constants.API_AUTHENTICATION_REQUIRED
	PERMISSION_DENIED       = constants.API_PERMISSION_DENIED
	INVALID_STATE           = constants.API_INVALID_STATE
	UNSUPPORTED             = constants.API_UNSUPPORTED
	RATE_LIMIT_EXCEEDED     = constants.API_RATE_LIMIT_EXCEEDED
	DEPENDENCY_FAILED       = constants.API_DEPENDENCY_FAILED
	NOT_FOUND               = constants.API_NOT_FOUND

	// Additional error codes
	INSUFFICIENT_STORAGE     = -32007
	CAPABILITY_NOT_SUPPORTED = -32008

	// Legacy constants for backward compatibility (deprecated)
	INSUFFICIENT_PERMISSIONS = constants.API_INSUFFICIENT_PERMISSIONS
	CAMERA_NOT_FOUND         = constants.API_CAMERA_NOT_FOUND
	RECORDING_IN_PROGRESS    = constants.API_RECORDING_IN_PROGRESS
	MEDIAMTX_UNAVAILABLE     = constants.API_MEDIAMTX_UNAVAILABLE
	FILE_NOT_FOUND           = constants.API_FILE_NOT_FOUND

	// Enhanced Recording Management Error Codes
	ERROR_CAMERA_NOT_FOUND         = constants.ERROR_CAMERA_NOT_FOUND
	ERROR_CAMERA_NOT_AVAILABLE     = constants.ERROR_CAMERA_NOT_AVAILABLE
	ERROR_RECORDING_IN_PROGRESS    = constants.ERROR_RECORDING_IN_PROGRESS
	ERROR_MEDIAMTX_ERROR           = constants.ERROR_MEDIAMTX_ERROR
	ERROR_CAMERA_ALREADY_RECORDING = constants.ERROR_CAMERA_ALREADY_RECORDING
	ERROR_STORAGE_LOW              = constants.ERROR_STORAGE_LOW
	ERROR_STORAGE_CRITICAL         = constants.ERROR_STORAGE_CRITICAL
)

// ErrorMessages maps error codes to their corresponding messages
// Using common constants for consistency with API documentation
var ErrorMessages = constants.APIErrorMessages

// JsonRpcRequest represents a JSON-RPC 2.0 request structure
// Following Python JsonRpcRequest dataclass
type JsonRpcRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	ID      interface{}            `json:"id,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JsonRpcResponse represents a JSON-RPC 2.0 response structure
// Following Python JsonRpcResponse dataclass
type JsonRpcResponse struct {
	JSONRPC  string                 `json:"jsonrpc"`
	ID       interface{}            `json:"id,omitempty"`
	Result   interface{}            `json:"result,omitempty"`
	Error    *JsonRpcError          `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// JsonRpcNotification represents a JSON-RPC 2.0 notification structure
// Following Python JsonRpcNotification dataclass
type JsonRpcNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JsonRpcError represents a JSON-RPC 2.0 error structure
type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JsonRpcError) Error() string {
	return e.Message
}

// ErrorData standardizes JSON-RPC error data payloads
type ErrorData struct {
	Reason     string `json:"reason,omitempty"`
	Details    string `json:"details,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// NewJsonRpcError creates a standardized JSON-RPC error
func NewJsonRpcError(code int, reason, details, suggestion string) *JsonRpcError {
	return &JsonRpcError{
		Code:    code,
		Message: ErrorMessages[code],
		Data: &ErrorData{
			Reason:     reason,
			Details:    details,
			Suggestion: suggestion,
		},
	}
}

// ClientConnection represents a connected WebSocket client
// Following Python ClientConnection class
type ClientConnection struct {
	ClientID      string
	Authenticated bool
	UserID        string
	Role          string
	AuthMethod    string
	ConnectedAt   time.Time
	Subscriptions map[string]bool
	Conn          *websocket.Conn `json:"-"` // WebSocket connection for sending messages
}

// PerformanceMetrics tracks WebSocket server performance
// Following Python PerformanceMetrics class
// Note: RequestCount, ErrorCount, and ActiveConnections use atomic operations for thread safety
type PerformanceMetrics struct {
	RequestCount      int64
	ResponseTimes     map[string][]float64
	ErrorCount        int64
	ActiveConnections int64
	StartTime         time.Time
}

// MethodHandler defines the signature for JSON-RPC method handlers
type MethodHandler func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error)

// WebSocketMessage represents a WebSocket message with metadata
type WebSocketMessage struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	ClientID  string          `json:"client_id,omitempty"`
}

// ServerConfig contains WebSocket server configuration
// Following Python server configuration patterns
type ServerConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 int           `mapstructure:"port"`
	WebSocketPath        string        `mapstructure:"websocket_path"`
	MaxConnections       int           `mapstructure:"max_connections"`
	ReadTimeout          time.Duration `mapstructure:"read_timeout"`
	WriteTimeout         time.Duration `mapstructure:"write_timeout"`
	PingInterval         time.Duration `mapstructure:"ping_interval"`
	PongWait             time.Duration `mapstructure:"pong_wait"`
	MaxMessageSize       int64         `mapstructure:"max_message_size"`
	ReadBufferSize       int           `mapstructure:"read_buffer_size"`
	WriteBufferSize      int           `mapstructure:"write_buffer_size"`
	ShutdownTimeout      time.Duration `mapstructure:"shutdown_timeout"`       // Default: 30 seconds
	ClientCleanupTimeout time.Duration `mapstructure:"client_cleanup_timeout"` // Default: 10 seconds
	AutoCloseAfter       time.Duration `mapstructure:"auto_close_after"`       // Default: 0 (never auto-close)
}

// DefaultServerConfig returns default WebSocket server configuration
// Optimized for Epic E3 performance requirements: <50ms response time, 1000+ connections
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:                 constants.WEBSOCKET_DEFAULT_HOST,
		Port:                 constants.WEBSOCKET_DEFAULT_PORT,
		WebSocketPath:        constants.WEBSOCKET_DEFAULT_PATH,
		MaxConnections:       constants.WEBSOCKET_MAX_CONNECTIONS_PRODUCTION,
		ReadTimeout:          constants.WEBSOCKET_READ_TIMEOUT,
		WriteTimeout:         constants.WEBSOCKET_WRITE_TIMEOUT,
		PingInterval:         constants.WEBSOCKET_PING_INTERVAL,
		PongWait:             constants.WEBSOCKET_PONG_WAIT,
		MaxMessageSize:       constants.WEBSOCKET_MAX_MESSAGE_SIZE,
		ReadBufferSize:       constants.WEBSOCKET_READ_BUFFER_SIZE,
		WriteBufferSize:      constants.WEBSOCKET_WRITE_BUFFER_SIZE,
		ShutdownTimeout:      constants.WEBSOCKET_SHUTDOWN_TIMEOUT,
		ClientCleanupTimeout: constants.WEBSOCKET_CLIENT_CLEANUP_TIMEOUT,
		AutoCloseAfter:       0, // Default: never auto-close
	}
}
