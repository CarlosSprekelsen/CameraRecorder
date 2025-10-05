/*
WebSocket Test Client - Universal Testing Utility

Provides universal WebSocket client functionality for all test types:
- Integration tests
- E2E tests  
- Any test needing WebSocket connectivity

This is the single source of truth for WebSocket testing infrastructure.

Design Principles:
- Universal utility available to all test packages
- Real WebSocket connections with JSON-RPC 2.0 protocol
- Progressive Readiness testing support
- Proper timeout management using testutils constants
*/

package testutils

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WebSocketTestClient provides universal WebSocket client for all tests
// Can be used by integration tests, E2E tests, and any test needing WebSocket
type WebSocketTestClient struct {
	t      *testing.T
	url    string
	conn   *websocket.Conn
	nextID int
}

// NewWebSocketTestClient creates a new WebSocket test client
func NewWebSocketTestClient(t *testing.T, url string) *WebSocketTestClient {
	return &WebSocketTestClient{
		t:      t,
		url:    url,
		nextID: 1,
	}
}

// Connect establishes WebSocket connection
func (c *WebSocketTestClient) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: UniversalTimeoutExtreme,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	return nil
}

// SendJSONRPC sends JSON-RPC request and returns response
func (c *WebSocketTestClient) SendJSONRPC(method string, params interface{}) (*JSONRPCResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to WebSocket")
	}

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextID,
	}
	c.nextID++

	err := c.conn.WriteJSON(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send JSON-RPC request: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(UniversalTimeoutExtreme))

	var response JSONRPCResponse
	err = c.conn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON-RPC response: %w", err)
	}

	return &response, nil
}

// Ping tests connectivity (no authentication required)
func (c *WebSocketTestClient) Ping() error {
	response, err := c.SendJSONRPC("ping", nil)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("ping returned error: %d %s", response.Error.Code, response.Error.Message)
	}

	if response.Result != "pong" {
		return fmt.Errorf("expected 'pong', got %v", response.Result)
	}

	return nil
}

// Authenticate authenticates with JWT token
func (c *WebSocketTestClient) Authenticate(authToken string) error {
	params := map[string]interface{}{
		"auth_token": authToken,
	}

	response, err := c.SendJSONRPC("authenticate", params)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("authentication returned error: %d %s", response.Error.Code, response.Error.Message)
	}

	// Validate AuthResult structure according to OpenRPC spec
	authResult, ok := response.Result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid auth result format")
	}

	// Check required fields
	if _, ok := authResult["role"]; !ok {
		return fmt.Errorf("auth result missing 'role' field")
	}

	if _, ok := authResult["authenticated"]; !ok {
		return fmt.Errorf("auth result missing 'authenticated' field")
	}

	// Verify authentication succeeded
	if authenticated, ok := authResult["authenticated"].(bool); !ok || !authenticated {
		return fmt.Errorf("authentication failed: authenticated field is false or invalid")
	}

	return nil
}

// GetCameraList gets camera list
func (c *WebSocketTestClient) GetCameraList() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_camera_list", nil)
}

// GetCameraStatus gets camera status
func (c *WebSocketTestClient) GetCameraStatus(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("get_camera_status", params)
}

// GetCameraCapabilities gets camera capabilities
func (c *WebSocketTestClient) GetCameraCapabilities(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("get_camera_capabilities", params)
}

// StartRecording starts recording
func (c *WebSocketTestClient) StartRecording(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("start_recording", params)
}

// StartRecordingWithDuration starts recording with duration
func (c *WebSocketTestClient) StartRecordingWithDuration(device string, duration int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":   device,
		"duration": duration,
	}
	return c.SendJSONRPC("start_recording", params)
}

// StopRecording stops recording
func (c *WebSocketTestClient) StopRecording(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("stop_recording", params)
}

// ListRecordings lists recordings
func (c *WebSocketTestClient) ListRecordings() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("list_recordings", nil)
}

// TakeSnapshot takes snapshot
func (c *WebSocketTestClient) TakeSnapshot(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("take_snapshot", params)
}

// TakeSnapshotWithFormat takes snapshot with format options
func (c *WebSocketTestClient) TakeSnapshotWithFormat(device string, format string, quality int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":  device,
		"format":  format,
		"quality": quality,
	}
	return c.SendJSONRPC("take_snapshot", params)
}

// ListSnapshots lists snapshots
func (c *WebSocketTestClient) ListSnapshots() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("list_snapshots", nil)
}

// GetSystemHealth gets system health
func (c *WebSocketTestClient) GetSystemHealth() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_system_health", nil)
}

// GetSystemMetrics gets system metrics
func (c *WebSocketTestClient) GetSystemMetrics() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_system_metrics", nil)
}

// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
