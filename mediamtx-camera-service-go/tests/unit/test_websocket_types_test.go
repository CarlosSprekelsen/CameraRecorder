//go:build unit
// +build unit

/*
WebSocket Types Unit Tests

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Error code and message management
- REQ-API-005: Client connection management
- REQ-API-006: Performance metrics tracking
- REQ-API-007: Server configuration management

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocket_ErrorCodes tests error code constants
func TestWebSocket_ErrorCodes(t *testing.T) {
	// REQ-API-004: Error code and message management

	// Test JSON-RPC 2.0 error codes
	assert.Equal(t, -32001, websocket.AUTHENTICATION_REQUIRED)
	assert.Equal(t, -32002, websocket.RATE_LIMIT_EXCEEDED)
	assert.Equal(t, -32003, websocket.INSUFFICIENT_PERMISSIONS)
	assert.Equal(t, -32004, websocket.CAMERA_NOT_FOUND)
	assert.Equal(t, -32005, websocket.RECORDING_IN_PROGRESS)
	assert.Equal(t, -32006, websocket.MEDIAMTX_UNAVAILABLE)
	assert.Equal(t, -32007, websocket.INSUFFICIENT_STORAGE)
	assert.Equal(t, -32008, websocket.CAPABILITY_NOT_SUPPORTED)
	assert.Equal(t, -32601, websocket.METHOD_NOT_FOUND)
	assert.Equal(t, -32602, websocket.INVALID_PARAMS)
	assert.Equal(t, -32603, websocket.INTERNAL_ERROR)

	// Test enhanced recording management error codes
	assert.Equal(t, -1000, websocket.ERROR_CAMERA_NOT_FOUND)
	assert.Equal(t, -1001, websocket.ERROR_CAMERA_NOT_AVAILABLE)
	assert.Equal(t, -1002, websocket.ERROR_RECORDING_IN_PROGRESS)
	assert.Equal(t, -1003, websocket.ERROR_MEDIAMTX_ERROR)
	assert.Equal(t, -1006, websocket.ERROR_CAMERA_ALREADY_RECORDING)
	assert.Equal(t, -1008, websocket.ERROR_STORAGE_LOW)
	assert.Equal(t, -1010, websocket.ERROR_STORAGE_CRITICAL)
}

// TestWebSocket_ErrorMessages tests error message mapping
func TestWebSocket_ErrorMessages(t *testing.T) {
	// REQ-API-004: Error code and message management

	// Test JSON-RPC 2.0 error messages
	assert.Equal(t, "Authentication required", websocket.ErrorMessages[websocket.AUTHENTICATION_REQUIRED])
	assert.Equal(t, "Rate limit exceeded", websocket.ErrorMessages[websocket.RATE_LIMIT_EXCEEDED])
	assert.Equal(t, "Insufficient permissions", websocket.ErrorMessages[websocket.INSUFFICIENT_PERMISSIONS])
	assert.Equal(t, "Camera not found or disconnected", websocket.ErrorMessages[websocket.CAMERA_NOT_FOUND])
	assert.Equal(t, "Recording already in progress", websocket.ErrorMessages[websocket.RECORDING_IN_PROGRESS])
	assert.Equal(t, "MediaMTX service unavailable", websocket.ErrorMessages[websocket.MEDIAMTX_UNAVAILABLE])
	assert.Equal(t, "Insufficient storage space", websocket.ErrorMessages[websocket.INSUFFICIENT_STORAGE])
	assert.Equal(t, "Camera capability not supported", websocket.ErrorMessages[websocket.CAPABILITY_NOT_SUPPORTED])
	assert.Equal(t, "Method not found", websocket.ErrorMessages[websocket.METHOD_NOT_FOUND])
	assert.Equal(t, "Invalid parameters", websocket.ErrorMessages[websocket.INVALID_PARAMS])
	assert.Equal(t, "Internal server error", websocket.ErrorMessages[websocket.INTERNAL_ERROR])

	// Test enhanced recording management error messages
	assert.Equal(t, "Camera not found", websocket.ErrorMessages[websocket.ERROR_CAMERA_NOT_FOUND])
	assert.Equal(t, "Camera not available", websocket.ErrorMessages[websocket.ERROR_CAMERA_NOT_AVAILABLE])
	assert.Equal(t, "Recording in progress", websocket.ErrorMessages[websocket.ERROR_RECORDING_IN_PROGRESS])
	assert.Equal(t, "MediaMTX error", websocket.ErrorMessages[websocket.ERROR_MEDIAMTX_ERROR])
	assert.Equal(t, "Camera is currently recording", websocket.ErrorMessages[websocket.ERROR_CAMERA_ALREADY_RECORDING])
	assert.Equal(t, "Storage space is low", websocket.ErrorMessages[websocket.ERROR_STORAGE_LOW])
	assert.Equal(t, "Storage space is critical", websocket.ErrorMessages[websocket.ERROR_STORAGE_CRITICAL])

	// Test non-existent error code
	_, exists := websocket.ErrorMessages[999999]
	assert.False(t, exists, "Non-existent error code should not have a message")
}

// TestWebSocket_JsonRpcRequest tests JSON-RPC request structure
func TestWebSocket_JsonRpcRequest(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// Test request creation
	request := websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "test_method",
		ID:      1,
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	assert.Equal(t, "2.0", request.JSONRPC)
	assert.Equal(t, "test_method", request.Method)
	assert.Equal(t, 1, request.ID)
	assert.Equal(t, "value1", request.Params["param1"])
	assert.Equal(t, 42, request.Params["param2"])

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledRequest websocket.JsonRpcRequest
	err = json.Unmarshal(jsonData, &unmarshaledRequest)
	require.NoError(t, err)

	assert.Equal(t, request.JSONRPC, unmarshaledRequest.JSONRPC)
	assert.Equal(t, request.Method, unmarshaledRequest.Method)
	// JSON unmarshaling converts numbers to float64, so we need to check the value, not the type
	assert.Equal(t, float64(1), unmarshaledRequest.ID)
	assert.Equal(t, request.Params["param1"], unmarshaledRequest.Params["param1"])
	assert.Equal(t, float64(42), unmarshaledRequest.Params["param2"])
}

// TestWebSocket_JsonRpcRequestWithoutID tests request without ID (notification)
func TestWebSocket_JsonRpcRequestWithoutID(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	request := websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "notification_method",
		Params: map[string]interface{}{
			"notification": true,
		},
	}

	assert.Equal(t, "2.0", request.JSONRPC)
	assert.Equal(t, "notification_method", request.Method)
	assert.Nil(t, request.ID)
	assert.Equal(t, true, request.Params["notification"])

	// Test JSON marshaling without ID
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaledRequest websocket.JsonRpcRequest
	err = json.Unmarshal(jsonData, &unmarshaledRequest)
	require.NoError(t, err)

	assert.Equal(t, request.JSONRPC, unmarshaledRequest.JSONRPC)
	assert.Equal(t, request.Method, unmarshaledRequest.Method)
	assert.Nil(t, unmarshaledRequest.ID)
}

// TestWebSocket_JsonRpcResponse tests JSON-RPC response structure
func TestWebSocket_JsonRpcResponse(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// Test successful response
	response := websocket.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"success": true,
			"data":    "test_data",
		},
	}

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, true, response.Result.(map[string]interface{})["success"])
	assert.Equal(t, "test_data", response.Result.(map[string]interface{})["data"])
	assert.Nil(t, response.Error)

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledResponse websocket.JsonRpcResponse
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	require.NoError(t, err)

	assert.Equal(t, response.JSONRPC, unmarshaledResponse.JSONRPC)
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, float64(1), unmarshaledResponse.ID)
	assert.Equal(t, response.Result.(map[string]interface{})["success"], unmarshaledResponse.Result.(map[string]interface{})["success"])
	assert.Equal(t, response.Result.(map[string]interface{})["data"], unmarshaledResponse.Result.(map[string]interface{})["data"])
}

// TestWebSocket_JsonRpcResponseWithError tests response with error
func TestWebSocket_JsonRpcResponseWithError(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	errorObj := &websocket.JsonRpcError{
		Code:    websocket.INVALID_PARAMS,
		Message: "Invalid parameters",
		Data:    "Additional error data",
	}

	response := websocket.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      1,
		Error:   errorObj,
	}

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID)
	assert.Nil(t, response.Result)
	assert.Equal(t, websocket.INVALID_PARAMS, response.Error.Code)
	assert.Equal(t, "Invalid parameters", response.Error.Message)
	assert.Equal(t, "Additional error data", response.Error.Data)

	// Test JSON marshaling with error
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test JSON unmarshaling with error
	var unmarshaledResponse websocket.JsonRpcResponse
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	require.NoError(t, err)

	assert.Equal(t, response.JSONRPC, unmarshaledResponse.JSONRPC)
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, float64(1), unmarshaledResponse.ID)
	assert.Equal(t, response.Error.Code, unmarshaledResponse.Error.Code)
	assert.Equal(t, response.Error.Message, unmarshaledResponse.Error.Message)
	assert.Equal(t, response.Error.Data, unmarshaledResponse.Error.Data)
}

// TestWebSocket_JsonRpcNotification tests JSON-RPC notification structure
func TestWebSocket_JsonRpcNotification(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	notification := websocket.JsonRpcNotification{
		JSONRPC: "2.0",
		Method:  "status_update",
		Params: map[string]interface{}{
			"status":    "online",
			"timestamp": time.Now().Unix(),
		},
	}

	assert.Equal(t, "2.0", notification.JSONRPC)
	assert.Equal(t, "status_update", notification.Method)
	assert.Equal(t, "online", notification.Params["status"])
	assert.NotNil(t, notification.Params["timestamp"])

	// Test JSON marshaling
	jsonData, err := json.Marshal(notification)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledNotification websocket.JsonRpcNotification
	err = json.Unmarshal(jsonData, &unmarshaledNotification)
	require.NoError(t, err)

	assert.Equal(t, notification.JSONRPC, unmarshaledNotification.JSONRPC)
	assert.Equal(t, notification.Method, unmarshaledNotification.Method)
	assert.Equal(t, notification.Params["status"], unmarshaledNotification.Params["status"])
}

// TestWebSocket_JsonRpcError tests JSON-RPC error structure
func TestWebSocket_JsonRpcError(t *testing.T) {
	// REQ-API-004: Error code and message management

	errorObj := &websocket.JsonRpcError{
		Code:    websocket.CAMERA_NOT_FOUND,
		Message: "Camera not found",
		Data: map[string]interface{}{
			"camera_id": "camera_123",
			"reason":    "disconnected",
		},
	}

	assert.Equal(t, websocket.CAMERA_NOT_FOUND, errorObj.Code)
	assert.Equal(t, "Camera not found", errorObj.Message)
	assert.Equal(t, "camera_123", errorObj.Data.(map[string]interface{})["camera_id"])
	assert.Equal(t, "disconnected", errorObj.Data.(map[string]interface{})["reason"])

	// Test JSON marshaling
	jsonData, err := json.Marshal(errorObj)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledError websocket.JsonRpcError
	err = json.Unmarshal(jsonData, &unmarshaledError)
	require.NoError(t, err)

	assert.Equal(t, errorObj.Code, unmarshaledError.Code)
	assert.Equal(t, errorObj.Message, unmarshaledError.Message)
	assert.Equal(t, errorObj.Data.(map[string]interface{})["camera_id"], unmarshaledError.Data.(map[string]interface{})["camera_id"])
	assert.Equal(t, errorObj.Data.(map[string]interface{})["reason"], unmarshaledError.Data.(map[string]interface{})["reason"])
}

// TestWebSocket_ClientConnection tests client connection structure
func TestWebSocket_ClientConnection(t *testing.T) {
	// REQ-API-005: Client connection management

	now := time.Now()
	client := &websocket.ClientConnection{
		ClientID:      "client_123",
		Authenticated: true,
		UserID:        "user_456",
		Role:          "operator",
		AuthMethod:    "jwt",
		ConnectedAt:   now,
		Subscriptions: map[string]bool{
			"camera_status": true,
			"recording":     false,
		},
	}

	assert.Equal(t, "client_123", client.ClientID)
	assert.True(t, client.Authenticated)
	assert.Equal(t, "user_456", client.UserID)
	assert.Equal(t, "operator", client.Role)
	assert.Equal(t, "jwt", client.AuthMethod)
	assert.Equal(t, now, client.ConnectedAt)
	assert.True(t, client.Subscriptions["camera_status"])
	assert.False(t, client.Subscriptions["recording"])

	// Test subscription management
	client.Subscriptions["new_topic"] = true
	assert.True(t, client.Subscriptions["new_topic"])

	delete(client.Subscriptions, "recording")
	_, exists := client.Subscriptions["recording"]
	assert.False(t, exists)
}

// TestWebSocket_ClientConnectionUnauthenticated tests unauthenticated client
func TestWebSocket_ClientConnectionUnauthenticated(t *testing.T) {
	// REQ-API-005: Client connection management

	now := time.Now()
	client := &websocket.ClientConnection{
		ClientID:      "client_456",
		Authenticated: false,
		UserID:        "",
		Role:          "",
		AuthMethod:    "",
		ConnectedAt:   now,
		Subscriptions: make(map[string]bool),
	}

	assert.Equal(t, "client_456", client.ClientID)
	assert.False(t, client.Authenticated)
	assert.Empty(t, client.UserID)
	assert.Empty(t, client.Role)
	assert.Empty(t, client.AuthMethod)
	assert.Equal(t, now, client.ConnectedAt)
	assert.Empty(t, client.Subscriptions)
}

// TestWebSocket_PerformanceMetrics tests performance metrics structure
func TestWebSocket_PerformanceMetrics(t *testing.T) {
	// REQ-API-006: Performance metrics tracking

	now := time.Now()
	metrics := &websocket.PerformanceMetrics{
		RequestCount: 100,
		ResponseTimes: map[string][]float64{
			"camera_list":  {10.5, 12.3, 8.9},
			"start_record": {25.1, 22.8, 30.2},
		},
		ErrorCount:        5,
		ActiveConnections: 25,
		StartTime:         now,
	}

	assert.Equal(t, int64(100), metrics.RequestCount)
	assert.Len(t, metrics.ResponseTimes["camera_list"], 3)
	assert.Equal(t, 10.5, metrics.ResponseTimes["camera_list"][0])
	assert.Equal(t, 12.3, metrics.ResponseTimes["camera_list"][1])
	assert.Equal(t, 8.9, metrics.ResponseTimes["camera_list"][2])
	assert.Len(t, metrics.ResponseTimes["start_record"], 3)
	assert.Equal(t, int64(5), metrics.ErrorCount)
	assert.Equal(t, int64(25), metrics.ActiveConnections)
	assert.Equal(t, now, metrics.StartTime)

	// Test metrics updates
	metrics.RequestCount++
	assert.Equal(t, int64(101), metrics.RequestCount)

	metrics.ResponseTimes["new_method"] = []float64{15.0, 18.5}
	assert.Len(t, metrics.ResponseTimes["new_method"], 2)

	metrics.ErrorCount++
	assert.Equal(t, int64(6), metrics.ErrorCount)

	metrics.ActiveConnections--
	assert.Equal(t, int64(24), metrics.ActiveConnections)
}

// TestWebSocket_PerformanceMetricsEmpty tests empty performance metrics
func TestWebSocket_PerformanceMetricsEmpty(t *testing.T) {
	// REQ-API-006: Performance metrics tracking

	now := time.Now()
	metrics := &websocket.PerformanceMetrics{
		RequestCount:      0,
		ResponseTimes:     make(map[string][]float64),
		ErrorCount:        0,
		ActiveConnections: 0,
		StartTime:         now,
	}

	assert.Equal(t, int64(0), metrics.RequestCount)
	assert.Empty(t, metrics.ResponseTimes)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, int64(0), metrics.ActiveConnections)
	assert.Equal(t, now, metrics.StartTime)
}

// TestWebSocket_WebSocketMessage tests WebSocket message structure
func TestWebSocket_WebSocketMessage(t *testing.T) {
	// REQ-API-003: Request/response message handling

	now := time.Now()
	messageData := json.RawMessage(`{"key": "value", "number": 42}`)

	message := &websocket.WebSocketMessage{
		Type:      "jsonrpc",
		Data:      messageData,
		Timestamp: now,
		ClientID:  "client_789",
	}

	assert.Equal(t, "jsonrpc", message.Type)
	assert.Equal(t, messageData, message.Data)
	assert.Equal(t, now, message.Timestamp)
	assert.Equal(t, "client_789", message.ClientID)

	// Test JSON marshaling
	jsonData, err := json.Marshal(message)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledMessage websocket.WebSocketMessage
	err = json.Unmarshal(jsonData, &unmarshaledMessage)
	require.NoError(t, err)

	assert.Equal(t, message.Type, unmarshaledMessage.Type)
	// JSON marshaling removes spaces, so we need to compare the parsed content
	var originalData, unmarshaledData map[string]interface{}
	json.Unmarshal(message.Data, &originalData)
	json.Unmarshal(unmarshaledMessage.Data, &unmarshaledData)
	assert.Equal(t, originalData, unmarshaledData)
	assert.Equal(t, message.ClientID, unmarshaledMessage.ClientID)
}

// TestWebSocket_WebSocketMessageWithoutClientID tests message without client ID
func TestWebSocket_WebSocketMessageWithoutClientID(t *testing.T) {
	// REQ-API-003: Request/response message handling

	now := time.Now()
	messageData := json.RawMessage(`{"notification": true}`)

	message := &websocket.WebSocketMessage{
		Type:      "notification",
		Data:      messageData,
		Timestamp: now,
	}

	assert.Equal(t, "notification", message.Type)
	assert.Equal(t, messageData, message.Data)
	assert.Equal(t, now, message.Timestamp)
	assert.Empty(t, message.ClientID)
}

// TestWebSocket_ServerConfig tests server configuration structure
func TestWebSocket_ServerConfig(t *testing.T) {
	// REQ-API-007: Server configuration management

	config := &websocket.ServerConfig{
		Host:           "127.0.0.1",
		Port:           9000,
		WebSocketPath:  "/custom/ws",
		MaxConnections: 500,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   5 * time.Second,
		PingInterval:   45 * time.Second,
		PongWait:       90 * time.Second,
		MaxMessageSize: 2048 * 1024, // 2MB
	}

	assert.Equal(t, "127.0.0.1", config.Host)
	assert.Equal(t, 9000, config.Port)
	assert.Equal(t, "/custom/ws", config.WebSocketPath)
	assert.Equal(t, 500, config.MaxConnections)
	assert.Equal(t, 10*time.Second, config.ReadTimeout)
	assert.Equal(t, 5*time.Second, config.WriteTimeout)
	assert.Equal(t, 45*time.Second, config.PingInterval)
	assert.Equal(t, 90*time.Second, config.PongWait)
	assert.Equal(t, int64(2048*1024), config.MaxMessageSize)
}

// TestWebSocket_DefaultServerConfig tests default server configuration
func TestWebSocket_DefaultServerConfig(t *testing.T) {
	// REQ-API-007: Server configuration management

	config := websocket.DefaultServerConfig()

	assert.Equal(t, "0.0.0.0", config.Host)
	assert.Equal(t, 8002, config.Port)
	assert.Equal(t, "/ws", config.WebSocketPath)
	assert.Equal(t, 1000, config.MaxConnections)
	assert.Equal(t, 5*time.Second, config.ReadTimeout)
	assert.Equal(t, 1*time.Second, config.WriteTimeout)
	assert.Equal(t, 30*time.Second, config.PingInterval)
	assert.Equal(t, 60*time.Second, config.PongWait)
	assert.Equal(t, int64(1024*1024), config.MaxMessageSize) // 1MB
}

// TestWebSocket_MethodHandler tests method handler function type
func TestWebSocket_MethodHandler(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	// Create a test method handler
	var handler websocket.MethodHandler = func(params map[string]interface{}, client *websocket.ClientConnection) (*websocket.JsonRpcResponse, error) {
		return &websocket.JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result: map[string]interface{}{
				"handler_called": true,
				"params":         params,
				"client_id":      client.ClientID,
			},
		}, nil
	}

	// Test the handler
	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: true,
		UserID:        "test_user",
		Role:          "viewer",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	params := map[string]interface{}{
		"test_param": "test_value",
	}

	response, err := handler(params, client)
	require.NoError(t, err)
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, true, response.Result.(map[string]interface{})["handler_called"])
	assert.Equal(t, params, response.Result.(map[string]interface{})["params"])
	assert.Equal(t, "test_client", response.Result.(map[string]interface{})["client_id"])
}

// TestWebSocket_JsonRpcRequestStringID tests request with string ID
func TestWebSocket_JsonRpcRequestStringID(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	request := websocket.JsonRpcRequest{
		JSONRPC: "2.0",
		Method:  "string_id_method",
		ID:      "request_123",
		Params: map[string]interface{}{
			"string_id": true,
		},
	}

	assert.Equal(t, "2.0", request.JSONRPC)
	assert.Equal(t, "string_id_method", request.Method)
	assert.Equal(t, "request_123", request.ID)
	assert.Equal(t, true, request.Params["string_id"])

	// Test JSON marshaling with string ID
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	// Test JSON unmarshaling with string ID
	var unmarshaledRequest websocket.JsonRpcRequest
	err = json.Unmarshal(jsonData, &unmarshaledRequest)
	require.NoError(t, err)

	assert.Equal(t, request.JSONRPC, unmarshaledRequest.JSONRPC)
	assert.Equal(t, request.Method, unmarshaledRequest.Method)
	assert.Equal(t, request.ID, unmarshaledRequest.ID)
}

// TestWebSocket_JsonRpcResponseStringID tests response with string ID
func TestWebSocket_JsonRpcResponseStringID(t *testing.T) {
	// REQ-API-002: JSON-RPC 2.0 protocol implementation

	response := websocket.JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      "response_456",
		Result: map[string]interface{}{
			"string_id_response": true,
		},
	}

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, "response_456", response.ID)
	assert.Equal(t, true, response.Result.(map[string]interface{})["string_id_response"])

	// Test JSON marshaling with string ID
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Test JSON unmarshaling with string ID
	var unmarshaledResponse websocket.JsonRpcResponse
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	require.NoError(t, err)

	assert.Equal(t, response.JSONRPC, unmarshaledResponse.JSONRPC)
	assert.Equal(t, response.ID, unmarshaledResponse.ID)
}

// TestWebSocket_JsonRpcErrorWithoutData tests error without data field
func TestWebSocket_JsonRpcErrorWithoutData(t *testing.T) {
	// REQ-API-004: Error code and message management

	errorObj := &websocket.JsonRpcError{
		Code:    websocket.INTERNAL_ERROR,
		Message: "Internal server error",
	}

	assert.Equal(t, websocket.INTERNAL_ERROR, errorObj.Code)
	assert.Equal(t, "Internal server error", errorObj.Message)
	assert.Nil(t, errorObj.Data)

	// Test JSON marshaling without data
	jsonData, err := json.Marshal(errorObj)
	require.NoError(t, err)

	// Test JSON unmarshaling without data
	var unmarshaledError websocket.JsonRpcError
	err = json.Unmarshal(jsonData, &unmarshaledError)
	require.NoError(t, err)

	assert.Equal(t, errorObj.Code, unmarshaledError.Code)
	assert.Equal(t, errorObj.Message, unmarshaledError.Message)
	assert.Nil(t, unmarshaledError.Data)
}

// TestWebSocket_ClientConnectionSubscriptions tests subscription management
func TestWebSocket_ClientConnectionSubscriptions(t *testing.T) {
	// REQ-API-005: Client connection management

	client := &websocket.ClientConnection{
		ClientID:      "subscription_test",
		Authenticated: true,
		UserID:        "user_789",
		Role:          "admin",
		AuthMethod:    "api_key",
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	// Test adding subscriptions
	client.Subscriptions["camera_events"] = true
	client.Subscriptions["system_alerts"] = true
	client.Subscriptions["performance_metrics"] = false

	assert.True(t, client.Subscriptions["camera_events"])
	assert.True(t, client.Subscriptions["system_alerts"])
	assert.False(t, client.Subscriptions["performance_metrics"])

	// Test subscription count
	assert.Len(t, client.Subscriptions, 3)

	// Test removing subscriptions
	delete(client.Subscriptions, "performance_metrics")
	assert.Len(t, client.Subscriptions, 2)
	_, exists := client.Subscriptions["performance_metrics"]
	assert.False(t, exists)

	// Test subscription toggle
	client.Subscriptions["camera_events"] = false
	assert.False(t, client.Subscriptions["camera_events"])
}

// TestWebSocket_PerformanceMetricsResponseTimes tests response times management
func TestWebSocket_PerformanceMetricsResponseTimes(t *testing.T) {
	// REQ-API-006: Performance metrics tracking

	metrics := &websocket.PerformanceMetrics{
		RequestCount:      0,
		ResponseTimes:     make(map[string][]float64),
		ErrorCount:        0,
		ActiveConnections: 0,
		StartTime:         time.Now(),
	}

	// Test adding response times
	metrics.ResponseTimes["method1"] = append(metrics.ResponseTimes["method1"], 10.5)
	metrics.ResponseTimes["method1"] = append(metrics.ResponseTimes["method1"], 12.3)
	metrics.ResponseTimes["method2"] = append(metrics.ResponseTimes["method2"], 25.1)

	assert.Len(t, metrics.ResponseTimes["method1"], 2)
	assert.Equal(t, 10.5, metrics.ResponseTimes["method1"][0])
	assert.Equal(t, 12.3, metrics.ResponseTimes["method1"][1])
	assert.Len(t, metrics.ResponseTimes["method2"], 1)
	assert.Equal(t, 25.1, metrics.ResponseTimes["method2"][0])

	// Test clearing response times
	metrics.ResponseTimes["method1"] = nil
	assert.Nil(t, metrics.ResponseTimes["method1"])

	// Test non-existent method
	_, exists := metrics.ResponseTimes["nonexistent"]
	assert.False(t, exists)
}
