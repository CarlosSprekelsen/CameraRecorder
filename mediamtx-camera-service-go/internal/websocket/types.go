/*
WebSocket JSON-RPC 2.0 types and structures.

Provides JSON-RPC 2.0 request, response, and notification structures
following the Python WebSocketJsonRpcServer patterns and project architecture standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// JSON-RPC Error Codes (RFC 32700) - Following Python implementation
const (
	AUTHENTICATION_REQUIRED  = -32001
	RATE_LIMIT_EXCEEDED      = -32002
	INSUFFICIENT_PERMISSIONS = -32003
	CAMERA_NOT_FOUND         = -32004
	RECORDING_IN_PROGRESS    = -32005
	MEDIAMTX_UNAVAILABLE     = -32006
	INSUFFICIENT_STORAGE     = -32007
	CAPABILITY_NOT_SUPPORTED = -32008
	METHOD_NOT_FOUND         = -32601
	INVALID_PARAMS           = -32602
	INTERNAL_ERROR           = -32603

	// Enhanced Recording Management Error Codes
	ERROR_CAMERA_NOT_FOUND         = -1000
	ERROR_CAMERA_NOT_AVAILABLE     = -1001
	ERROR_RECORDING_IN_PROGRESS    = -1002
	ERROR_MEDIAMTX_ERROR           = -1003
	ERROR_CAMERA_ALREADY_RECORDING = -1006
	ERROR_STORAGE_LOW              = -1008
	ERROR_STORAGE_CRITICAL         = -1010
)

// ErrorMessages maps error codes to their corresponding messages
// Following Python ERROR_MESSAGES dictionary
var ErrorMessages = map[int]string{
	AUTHENTICATION_REQUIRED:        "Authentication required",
	RATE_LIMIT_EXCEEDED:            "Rate limit exceeded",
	INSUFFICIENT_PERMISSIONS:       "Insufficient permissions",
	CAMERA_NOT_FOUND:               "Camera not found or disconnected",
	RECORDING_IN_PROGRESS:          "Recording already in progress",
	MEDIAMTX_UNAVAILABLE:           "MediaMTX service unavailable",
	INSUFFICIENT_STORAGE:           "Insufficient storage space",
	CAPABILITY_NOT_SUPPORTED:       "Camera capability not supported",
	METHOD_NOT_FOUND:               "Method not found",
	INVALID_PARAMS:                 "Invalid parameters",
	INTERNAL_ERROR:                 "Internal server error",
	ERROR_CAMERA_NOT_FOUND:         "Camera not found",
	ERROR_CAMERA_NOT_AVAILABLE:     "Camera not available",
	ERROR_RECORDING_IN_PROGRESS:    "Recording in progress",
	ERROR_MEDIAMTX_ERROR:           "MediaMTX error",
	ERROR_CAMERA_ALREADY_RECORDING: "Camera is currently recording",
	ERROR_STORAGE_LOW:              "Storage space is low",
	ERROR_STORAGE_CRITICAL:         "Storage space is critical",
}

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
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
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
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	WebSocketPath  string        `mapstructure:"websocket_path"`
	MaxConnections int           `mapstructure:"max_connections"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	PingInterval   time.Duration `mapstructure:"ping_interval"`
	PongWait       time.Duration `mapstructure:"pong_wait"`
	MaxMessageSize int64         `mapstructure:"max_message_size"`
}

// DefaultServerConfig returns default WebSocket server configuration
// Optimized for Epic E3 performance requirements: <50ms response time, 1000+ connections
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:           "0.0.0.0",
		Port:           8002,
		WebSocketPath:  "/ws",
		MaxConnections: 1000,
		ReadTimeout:    5 * time.Second,  // Reduced for faster response detection
		WriteTimeout:   1 * time.Second,  // Reduced for faster message delivery
		PingInterval:   30 * time.Second, // Keep reasonable for connection health
		PongWait:       60 * time.Second, // Keep reasonable for connection stability
		MaxMessageSize: 1024 * 1024,      // 1MB
	}
}
