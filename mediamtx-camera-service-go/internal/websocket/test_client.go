/*
WebSocket Test Client - Real WebSocket Connection Testing

Provides WebSocket client functionality for integration testing with real
JSON-RPC 2.0 protocol implementation and OpenRPC API compliance.

API Documentation Reference: docs/api/mediamtx_camera_service_openrpc.json
Requirements Coverage:
- REQ-WS-001: WebSocket connection and authentication
- REQ-WS-002: Real-time camera operations
- REQ-WS-003: Error handling and recovery

Design Principles:
- Real WebSocket connections
- JSON-RPC 2.0 protocol compliance
- OpenRPC API method validation
- Progressive Readiness testing
*/

package websocket

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
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

// WebSocketTestClient provides real WebSocket client functionality
type WebSocketTestClient struct {
	t      *testing.T
	conn   *websocket.Conn
	url    string
	nextID int
}

// NewWebSocketTestClient creates a new WebSocket test client
func NewWebSocketTestClient(t *testing.T, serverURL string) *WebSocketTestClient {
	return &WebSocketTestClient{
		t:      t,
		url:    serverURL,
		nextID: 1,
	}
}

// Connect establishes a WebSocket connection to the server
func (c *WebSocketTestClient) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	// Create WebSocket connection with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	return nil
}

// SendJSONRPC sends a JSON-RPC 2.0 request and returns the response
func (c *WebSocketTestClient) SendJSONRPC(method string, params interface{}) (*JSONRPCResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to WebSocket")
	}

	// Create JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextID,
	}
	c.nextID++

	// Send request
	err := c.conn.WriteJSON(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send JSON-RPC request: %w", err)
	}

	// Read response with timeout
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

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

	return nil
}

// GetCameraList gets the list of cameras
func (c *WebSocketTestClient) GetCameraList() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_camera_list", nil)
}

// GetCameraStatus gets status for a specific camera
func (c *WebSocketTestClient) GetCameraStatus(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}
	return c.SendJSONRPC("get_camera_status", params)
}

// TakeSnapshot takes a snapshot of a camera
func (c *WebSocketTestClient) TakeSnapshot(device string, filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":   device,
		"filename": filename,
	}
	return c.SendJSONRPC("take_snapshot", params)
}

// StartRecording starts recording a camera
func (c *WebSocketTestClient) StartRecording(device string, duration int, format string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":   device,
		"duration": duration,
		"format":   format,
	}
	return c.SendJSONRPC("start_recording", params)
}

// StopRecording stops recording a camera
func (c *WebSocketTestClient) StopRecording(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}
	return c.SendJSONRPC("stop_recording", params)
}

// ListRecordings lists recording files
func (c *WebSocketTestClient) ListRecordings(limit int, offset int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	return c.SendJSONRPC("list_recordings", params)
}

// ListSnapshots lists snapshot files
func (c *WebSocketTestClient) ListSnapshots(limit int, offset int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	return c.SendJSONRPC("list_snapshots", params)
}

// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// AssertJSONRPCResponse validates JSON-RPC 2.0 response structure
func (c *WebSocketTestClient) AssertJSONRPCResponse(response *JSONRPCResponse, expectError bool) {
	require.Equal(c.t, "2.0", response.JSONRPC, "Response should have correct JSON-RPC version")
	require.NotNil(c.t, response.ID, "Response should have ID")

	if expectError {
		require.NotNil(c.t, response.Error, "Response should have error")
		require.Nil(c.t, response.Result, "Error response should not have result")
	} else {
		require.Nil(c.t, response.Error, "Response should not have error")
	}
}

// AssertCameraListResult validates CameraListResult structure
func (c *WebSocketTestClient) AssertCameraListResult(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields according to OpenRPC spec
	require.Contains(c.t, resultMap, "cameras", "Result should contain 'cameras' field")
	require.Contains(c.t, resultMap, "total", "Result should contain 'total' field")
	require.Contains(c.t, resultMap, "connected", "Result should contain 'connected' field")

	// Validate cameras array
	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(c.t, ok, "Cameras should be an array")

	// Validate each camera structure
	for i, camera := range cameras {
		cameraMap, ok := camera.(map[string]interface{})
		require.True(c.t, ok, "Camera %d should be a map", i)

		require.Contains(c.t, cameraMap, "device", "Camera %d should have 'device' field", i)
		require.Contains(c.t, cameraMap, "status", "Camera %d should have 'status' field", i)

		// Validate device ID pattern: ^camera[0-9]+$
		device, ok := cameraMap["device"].(string)
		require.True(c.t, ok, "Camera %d device should be string", i)
		require.Regexp(c.t, `^camera[0-9]+$`, device, "Camera %d device should match pattern", i)
	}
}
