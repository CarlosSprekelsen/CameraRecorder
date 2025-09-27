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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
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
		HandshakeTimeout: testutils.UniversalTimeoutExtreme,
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
	c.conn.SetReadDeadline(time.Now().Add(testutils.UniversalTimeoutExtreme))

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

	response, err := c.SendJSONRPC("start_recording", params)
	if err != nil {
		return nil, err
	}

	// Check for error response first
	if response.Error != nil {
		// Enhanced error logging for investigation
		fmt.Printf("DEBUG: StartRecording error details:\n")
		fmt.Printf("  Code: %d\n", response.Error.Code)
		fmt.Printf("  Message: %s\n", response.Error.Message)
		if response.Error.Data != nil {
			fmt.Printf("  Data: %+v\n", response.Error.Data)
		}
		return nil, fmt.Errorf("start_recording failed: %s", response.Error.Message)
	}

	// Add response validation to prevent nil result
	if response.Result == nil {
		return nil, fmt.Errorf("start_recording response missing result field")
	}

	// Validate response structure
	resultMap, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("start_recording result is not a map: %T", response.Result)
	}

	// Validate required fields per API documentation
	requiredFields := []string{"device", "filename", "status", "start_time", "format"}
	for _, field := range requiredFields {
		if _, exists := resultMap[field]; !exists {
			return nil, fmt.Errorf("start_recording result missing required field: %s", field)
		}
	}

	return response, nil
}

// StopRecording stops recording a camera
func (c *WebSocketTestClient) StopRecording(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}

	response, err := c.SendJSONRPC("stop_recording", params)
	if err != nil {
		return nil, err
	}

	// Add response validation to prevent nil result
	if response.Result == nil {
		return nil, fmt.Errorf("stop_recording response missing result field")
	}

	// Validate response structure
	resultMap, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stop_recording result is not a map: %T", response.Result)
	}

	// Validate required fields per API documentation
	requiredFields := []string{"device", "status"}
	for _, field := range requiredFields {
		if _, exists := resultMap[field]; !exists {
			return nil, fmt.Errorf("stop_recording result missing required field: %s", field)
		}
	}

	return response, nil
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

// GetCameraCapabilities gets camera capabilities
func (c *WebSocketTestClient) GetCameraCapabilities(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}
	return c.SendJSONRPC("get_camera_capabilities", params)
}

// DeleteRecording deletes a recording file by filename
func (c *WebSocketTestClient) DeleteRecording(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"filename": filename,
	}
	return c.SendJSONRPC("delete_recording", params)
}

// DeleteSnapshot deletes a snapshot file by filename
func (c *WebSocketTestClient) DeleteSnapshot(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"filename": filename,
	}
	return c.SendJSONRPC("delete_snapshot", params)
}

// ============================================================================
// MISSING STREAMING METHODS
// ============================================================================

// GetStreamUrl gets the stream URL for a camera
// CRITICAL: JSON-RPC API uses "device" parameter, not "camera_id"
func (c *WebSocketTestClient) GetStreamUrl(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}
	return c.SendJSONRPC("get_stream_url", params)
}

// GetStreamStatus gets the stream status for a camera
// CRITICAL: JSON-RPC API uses "device" parameter, not "camera_id"
func (c *WebSocketTestClient) GetStreamStatus(device string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"device": device,
	}
	return c.SendJSONRPC("get_stream_status", params)
}

// ============================================================================
// MISSING SYSTEM MONITORING METHODS
// ============================================================================

// GetMetrics gets system metrics
func (c *WebSocketTestClient) GetMetrics() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_metrics", nil)
}

// GetStreams gets all active streams
func (c *WebSocketTestClient) GetStreams() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_streams", nil)
}

// ============================================================================
// MISSING SYSTEM STATUS METHODS
// ============================================================================

// GetStatus gets system status
func (c *WebSocketTestClient) GetStatus() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_status", nil)
}

// GetSystemStatus gets system readiness status (viewer accessible)
func (c *WebSocketTestClient) GetSystemStatus() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_system_status", nil)
}

// GetServerInfo gets server information
func (c *WebSocketTestClient) GetServerInfo() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_server_info", nil)
}

// ============================================================================
// MISSING STORAGE MANAGEMENT METHODS
// ============================================================================

// GetStorageInfo gets storage information
func (c *WebSocketTestClient) GetStorageInfo() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_storage_info", nil)
}

// SetRetentionPolicy sets retention policy
// CRITICAL: JSON-RPC API uses specific parameters, not nested "policy" object
func (c *WebSocketTestClient) SetRetentionPolicy(policyType string, maxAgeDays int, maxSizeGb int, enabled bool) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"policy_type":  policyType,
		"max_age_days": maxAgeDays,
		"max_size_gb":  maxSizeGb,
		"enabled":      enabled,
	}
	return c.SendJSONRPC("set_retention_policy", params)
}

// CleanupOldFiles cleans up old files
func (c *WebSocketTestClient) CleanupOldFiles() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("cleanup_old_files", nil)
}

// ============================================================================
// MISSING FILE INFO METHODS
// ============================================================================

// GetRecordingInfo gets recording file information
func (c *WebSocketTestClient) GetRecordingInfo(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"filename": filename,
	}
	return c.SendJSONRPC("get_recording_info", params)
}

// GetSnapshotInfo gets snapshot file information
func (c *WebSocketTestClient) GetSnapshotInfo(filename string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"filename": filename,
	}
	return c.SendJSONRPC("get_snapshot_info", params)
}

// ============================================================================
// MISSING EVENT SUBSCRIPTION METHODS
// ============================================================================

// SubscribeEvents subscribes to events
func (c *WebSocketTestClient) SubscribeEvents(topics []string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"topics": topics,
	}
	return c.SendJSONRPC("subscribe_events", params)
}

// UnsubscribeEvents unsubscribes from events
func (c *WebSocketTestClient) UnsubscribeEvents() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("unsubscribe_events", nil)
}

// GetSubscriptionStats gets subscription statistics
func (c *WebSocketTestClient) GetSubscriptionStats() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_subscription_stats", nil)
}

// ============================================================================
// MISSING EXTERNAL STREAM METHODS
// ============================================================================

// DiscoverExternalStreams discovers external streams
func (c *WebSocketTestClient) DiscoverExternalStreams() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("discover_external_streams", nil)
}

// GetExternalStreams gets external streams
func (c *WebSocketTestClient) GetExternalStreams() (*JSONRPCResponse, error) {
	return c.SendJSONRPC("get_external_streams", nil)
}

// SetDiscoveryInterval sets discovery interval
// CRITICAL: JSON-RPC API uses "scan_interval" parameter, not "interval"
func (c *WebSocketTestClient) SetDiscoveryInterval(scanInterval int) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"scan_interval": scanInterval,
	}
	return c.SendJSONRPC("set_discovery_interval", params)
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

// AssertCameraStatusResult validates CameraStatusResult structure
func (c *WebSocketTestClient) AssertCameraStatusResult(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields according to OpenRPC spec
	require.Contains(c.t, resultMap, "device", "Result should contain 'device' field")
	require.Contains(c.t, resultMap, "status", "Result should contain 'status' field")
	require.Contains(c.t, resultMap, "capabilities", "Result should contain 'capabilities' field")

	// Validate device field
	device, ok := resultMap["device"].(string)
	require.True(c.t, ok, "Device should be string")
	require.Regexp(c.t, `^camera[0-9]+$`, device, "Device should match pattern")

	// Validate status field
	status, ok := resultMap["status"].(string)
	require.True(c.t, ok, "Status should be string")
	require.Contains(c.t, []string{"connected", "disconnected", "error"}, status, "Status should be valid")

	// Validate capabilities field
	capabilities, ok := resultMap["capabilities"].(map[string]interface{})
	require.True(c.t, ok, "Capabilities should be a map")
	require.Contains(c.t, capabilities, "formats", "Capabilities should contain 'formats' field")
	require.Contains(c.t, capabilities, "resolutions", "Capabilities should contain 'resolutions' field")
}

// AssertCameraCapabilitiesResult validates CameraCapabilitiesResult structure
func (c *WebSocketTestClient) AssertCameraCapabilitiesResult(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields according to OpenRPC spec
	require.Contains(c.t, resultMap, "device", "Result should contain 'device' field")
	require.Contains(c.t, resultMap, "formats", "Result should contain 'formats' field")
	require.Contains(c.t, resultMap, "resolutions", "Result should contain 'resolutions' field")
	require.Contains(c.t, resultMap, "frame_rates", "Result should contain 'frame_rates' field")

	// Validate device field
	device, ok := resultMap["device"].(string)
	require.True(c.t, ok, "Device should be string")
	require.Regexp(c.t, `^camera[0-9]+$`, device, "Device should match pattern")

	// Validate formats array
	formats, ok := resultMap["formats"].([]interface{})
	require.True(c.t, ok, "Formats should be an array")

	// Validate resolutions array
	_, ok = resultMap["resolutions"].([]interface{})
	require.True(c.t, ok, "Resolutions should be an array")

	// Validate frame_rates array
	_, ok = resultMap["frame_rates"].([]interface{})
	require.True(c.t, ok, "Frame rates should be an array")

	// At least one format should be available
	require.GreaterOrEqual(c.t, len(formats), 1, "At least one format should be available")
}

// AssertCameraListResultAPICompliant validates CameraListResult structure per API specification
func (c *WebSocketTestClient) AssertCameraListResultAPICompliant(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields per API spec: cameras, total, connected
	require.Contains(c.t, resultMap, "cameras", "Result should contain 'cameras' field per API spec")
	require.Contains(c.t, resultMap, "total", "Result should contain 'total' field per API spec")
	require.Contains(c.t, resultMap, "connected", "Result should contain 'connected' field per API spec")

	// Validate cameras array
	cameras, ok := resultMap["cameras"].([]interface{})
	require.True(c.t, ok, "Cameras should be an array per API spec")

	// Validate each camera structure per API spec
	for i, camera := range cameras {
		cameraMap, ok := camera.(map[string]interface{})
		require.True(c.t, ok, "Camera %d should be a map", i)

		// Required fields per API spec: device, status, name, resolution, fps, streams
		require.Contains(c.t, cameraMap, "device", "Camera %d should have 'device' field per API spec", i)
		require.Contains(c.t, cameraMap, "status", "Camera %d should have 'status' field per API spec", i)
		require.Contains(c.t, cameraMap, "name", "Camera %d should have 'name' field per API spec", i)
		require.Contains(c.t, cameraMap, "resolution", "Camera %d should have 'resolution' field per API spec", i)
		require.Contains(c.t, cameraMap, "fps", "Camera %d should have 'fps' field per API spec", i)
		require.Contains(c.t, cameraMap, "streams", "Camera %d should have 'streams' field per API spec", i)

		// Validate device ID pattern per API spec: camera0, camera1, etc.
		device, ok := cameraMap["device"].(string)
		require.True(c.t, ok, "Camera %d device should be string", i)
		require.Regexp(c.t, `^camera[0-9]+$`, device, "Camera %d device should match pattern per API spec", i)

		// Validate status values per API spec
		status, ok := cameraMap["status"].(string)
		require.True(c.t, ok, "Camera %d status should be string", i)
		require.Contains(c.t, []string{"CONNECTED", "DISCONNECTED", "ERROR"}, status,
			"Camera %d status should be valid per API spec", i)

		// Validate streams object per API spec
		streams, ok := cameraMap["streams"].(map[string]interface{})
		require.True(c.t, ok, "Camera %d streams should be a map per API spec", i)
		require.Contains(c.t, streams, "rtsp", "Camera %d should have 'rtsp' stream per API spec", i)
		require.Contains(c.t, streams, "hls", "Camera %d should have 'hls' stream per API spec", i)
	}

	// Validate total and connected are integers per API spec
	total, ok := resultMap["total"].(float64) // JSON numbers are float64
	require.True(c.t, ok, "Total should be a number per API spec")
	require.GreaterOrEqual(c.t, int(total), 0, "Total should be non-negative per API spec")

	connected, ok := resultMap["connected"].(float64) // JSON numbers are float64
	require.True(c.t, ok, "Connected should be a number per API spec")
	require.GreaterOrEqual(c.t, int(connected), 0, "Connected should be non-negative per API spec")
	require.LessOrEqual(c.t, int(connected), int(total), "Connected should not exceed total per API spec")
}

// AssertCameraStatusResultAPICompliant validates CameraStatusResult structure per API specification
func (c *WebSocketTestClient) AssertCameraStatusResultAPICompliant(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields per API spec: device, status, name, resolution, fps, streams
	require.Contains(c.t, resultMap, "device", "Result should contain 'device' field per API spec")
	require.Contains(c.t, resultMap, "status", "Result should contain 'status' field per API spec")
	require.Contains(c.t, resultMap, "name", "Result should contain 'name' field per API spec")
	require.Contains(c.t, resultMap, "resolution", "Result should contain 'resolution' field per API spec")
	require.Contains(c.t, resultMap, "fps", "Result should contain 'fps' field per API spec")
	require.Contains(c.t, resultMap, "streams", "Result should contain 'streams' field per API spec")

	// Validate device field per API spec
	device, ok := resultMap["device"].(string)
	require.True(c.t, ok, "Device should be string per API spec")
	require.Regexp(c.t, `^camera[0-9]+$`, device, "Device should match pattern per API spec")

	// Validate status field per API spec
	status, ok := resultMap["status"].(string)
	require.True(c.t, ok, "Status should be string per API spec")
	require.Contains(c.t, []string{"CONNECTED", "DISCONNECTED", "ERROR"}, status,
		"Status should be valid per API spec")

	// Validate streams object per API spec
	streams, ok := resultMap["streams"].(map[string]interface{})
	require.True(c.t, ok, "Streams should be a map per API spec")
	require.Contains(c.t, streams, "rtsp", "Should have 'rtsp' stream per API spec")
	require.Contains(c.t, streams, "hls", "Should have 'hls' stream per API spec")

	// Validate optional metrics field if present per API spec
	if metrics, exists := resultMap["metrics"]; exists {
		metricsMap, ok := metrics.(map[string]interface{})
		require.True(c.t, ok, "Metrics should be a map per API spec")
		require.Contains(c.t, metricsMap, "bytes_sent", "Metrics should contain 'bytes_sent' per API spec")
		require.Contains(c.t, metricsMap, "readers", "Metrics should contain 'readers' per API spec")
		require.Contains(c.t, metricsMap, "uptime", "Metrics should contain 'uptime' per API spec")
	}

	// Validate optional capabilities field if present per API spec
	if capabilities, exists := resultMap["capabilities"]; exists {
		capabilitiesMap, ok := capabilities.(map[string]interface{})
		require.True(c.t, ok, "Capabilities should be a map per API spec")
		require.Contains(c.t, capabilitiesMap, "formats", "Capabilities should contain 'formats' per API spec")
		require.Contains(c.t, capabilitiesMap, "resolutions", "Capabilities should contain 'resolutions' per API spec")
	}
}

// AssertCameraCapabilitiesResultAPICompliant validates CameraCapabilitiesResult structure per API specification
func (c *WebSocketTestClient) AssertCameraCapabilitiesResultAPICompliant(result interface{}) {
	resultMap, ok := result.(map[string]interface{})
	require.True(c.t, ok, "Result should be a map")

	// Check required fields per API spec: device, formats, resolutions, fps_options, validation_status
	require.Contains(c.t, resultMap, "device", "Result should contain 'device' field per API spec")
	require.Contains(c.t, resultMap, "formats", "Result should contain 'formats' field per API spec")
	require.Contains(c.t, resultMap, "resolutions", "Result should contain 'resolutions' field per API spec")
	require.Contains(c.t, resultMap, "fps_options", "Result should contain 'fps_options' field per API spec")
	require.Contains(c.t, resultMap, "validation_status", "Result should contain 'validation_status' field per API spec")

	// Validate device field per API spec
	device, ok := resultMap["device"].(string)
	require.True(c.t, ok, "Device should be string per API spec")
	require.Regexp(c.t, `^camera[0-9]+$`, device, "Device should match pattern per API spec")

	// Validate formats array per API spec
	formats, ok := resultMap["formats"].([]interface{})
	require.True(c.t, ok, "Formats should be an array per API spec")
	require.GreaterOrEqual(c.t, len(formats), 1, "At least one format should be available per API spec")

	// Validate resolutions array per API spec
	resolutions, ok := resultMap["resolutions"].([]interface{})
	require.True(c.t, ok, "Resolutions should be an array per API spec")
	require.GreaterOrEqual(c.t, len(resolutions), 1, "At least one resolution should be available per API spec")

	// Validate fps_options array per API spec
	fpsOptions, ok := resultMap["fps_options"].([]interface{})
	require.True(c.t, ok, "FPS options should be an array per API spec")
	require.GreaterOrEqual(c.t, len(fpsOptions), 1, "At least one FPS option should be available per API spec")

	// Validate validation_status per API spec
	validationStatus, ok := resultMap["validation_status"].(string)
	require.True(c.t, ok, "Validation status should be string per API spec")
	require.Contains(c.t, []string{"NONE", "DISCONNECTED", "CONFIRMED"}, validationStatus,
		"Validation status should be valid per API spec")
}
