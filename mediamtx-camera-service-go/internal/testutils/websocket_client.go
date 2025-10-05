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

// StartRecordingWithOptions starts recording with duration and format
func (c *WebSocketTestClient) StartRecordingWithOptions(device string, duration int, format string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":   device,
		"duration": duration,
		"format":   format,
	}
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

// ListRecordingsWithPagination lists recordings with pagination
func (c *WebSocketTestClient) ListRecordingsWithPagination(limit, offset int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	return c.SendJSONRPC("list_recordings", params)
}

// TakeSnapshot takes snapshot
func (c *WebSocketTestClient) TakeSnapshot(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("take_snapshot", params)
}

// TakeSnapshotWithFilename takes snapshot with custom filename
func (c *WebSocketTestClient) TakeSnapshotWithFilename(device string, filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device":   device,
		"filename": filename,
	}
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

// ListSnapshotsWithPagination lists snapshots with pagination
func (c *WebSocketTestClient) ListSnapshotsWithPagination(limit, offset int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	return c.SendJSONRPC("list_snapshots", params)
}

// GetSystemHealth gets system health
func (c *WebSocketTestClient) GetSystemHealth() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_system_health", nil)
}

// GetSystemMetrics gets system metrics
func (c *WebSocketTestClient) GetSystemMetrics() (*JSONRPCResponse, error) {
    return c.SendJSONRPC("get_metrics", nil)
}

// GetStatus gets full system status (admin)
func (c *WebSocketTestClient) GetStatus() (*JSONRPCResponse, error) {
    return c.SendJSONRPC("get_status", nil)
}

// GetSystemStatus gets viewer-accessible system readiness
func (c *WebSocketTestClient) GetSystemStatus() (*JSONRPCResponse, error) {
    return c.SendJSONRPC("get_system_status", nil)
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

// AssertCameraListResult validates camera list result structure
func (c *WebSocketTestClient) AssertCameraListResult(result interface{}) {
	require.NotNil(c.t, result, "Camera list result should not be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Camera list result should be a map")
	require.Contains(c.t, resultMap, "cameras", "Camera list should contain 'cameras' field")

	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(c.t, ok, "Cameras field should be an array")
	require.NotEmpty(c.t, cameras, "Camera list should not be empty")
}

// DeleteSnapshot deletes a snapshot
func (c *WebSocketTestClient) DeleteSnapshot(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"filename": filename}
	return c.SendJSONRPC("delete_snapshot", params)
}

// DeleteRecording deletes a recording
func (c *WebSocketTestClient) DeleteRecording(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"filename": filename}
	return c.SendJSONRPC("delete_recording", params)
}

// AssertCameraListResultAPICompliant validates camera list result structure for API compliance
func (c *WebSocketTestClient) AssertCameraListResultAPICompliant(result interface{}) {
	require.NotNil(c.t, result, "Camera list result should not be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Camera list result should be a map")
	require.Contains(c.t, resultMap, "cameras", "Camera list should contain 'cameras' field")

	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(c.t, ok, "Cameras field should be an array")
	require.NotEmpty(c.t, cameras, "Camera list should not be empty")
}

// AssertCameraStatusResultAPICompliant validates camera status result structure for API compliance
func (c *WebSocketTestClient) AssertCameraStatusResultAPICompliant(result interface{}) {
	require.NotNil(c.t, result, "Camera status result should not be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Camera status result should be a map")
	require.Contains(c.t, resultMap, "device", "Camera status should contain 'device' field")
	require.Contains(c.t, resultMap, "status", "Camera status should contain 'status' field")
}

// AssertCameraCapabilitiesResultAPICompliant validates camera capabilities result structure for API compliance
func (c *WebSocketTestClient) AssertCameraCapabilitiesResultAPICompliant(result interface{}) {
	require.NotNil(c.t, result, "Camera capabilities result should not be nil")

	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Camera capabilities result should be a map")
	require.Contains(c.t, resultMap, "formats", "Camera capabilities should contain 'formats' field")
	require.Contains(c.t, resultMap, "resolutions", "Camera capabilities should contain 'resolutions' field")
}

// SubscribeEvents subscribes to event notifications
func (c *WebSocketTestClient) SubscribeEvents(events []string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"events": events}
	return c.SendJSONRPC("subscribe_events", params)
}

// StartStreaming starts streaming
func (c *WebSocketTestClient) StartStreaming(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("start_streaming", params)
}

// StopStreaming stops streaming
func (c *WebSocketTestClient) StopStreaming(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"device": device}
	return c.SendJSONRPC("stop_streaming", params)
}

// UnsubscribeEvents unsubscribes from event notifications
func (c *WebSocketTestClient) UnsubscribeEvents(events []string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"events": events}
	return c.SendJSONRPC("unsubscribe_events", params)
}

// GetSubscriptionStats gets subscription statistics
func (c *WebSocketTestClient) GetSubscriptionStats() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_subscription_stats", nil)
}

// AddExternalStream adds an external stream
func (c *WebSocketTestClient) AddExternalStream(url string, name string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"url":  url,
		"name": name,
	}
	return c.SendJSONRPC("add_external_stream", params)
}

// RemoveExternalStream removes an external stream
func (c *WebSocketTestClient) RemoveExternalStream(url string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{"url": url}
	return c.SendJSONRPC("remove_external_stream", params)
}
