/*
WebSocket JSON-RPC 2.0 method registration and core method implementations.

Provides method registration system and core JSON-RPC method implementations
following the Python WebSocketJsonRpcServer patterns and project architecture standards.

Requirements Coverage:
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-004: Core method implementations (ping, authenticate, get_camera_list, get_camera_status)

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// registerBuiltinMethods registers all built-in JSON-RPC methods
// Following Python _register_builtin_methods patterns
func (s *WebSocketServer) registerBuiltinMethods() {
	// Core methods
	s.registerMethod("ping", s.methodPing, "1.0")
	s.registerMethod("authenticate", s.methodAuthenticate, "1.0")
	s.registerMethod("get_camera_list", s.methodGetCameraList, "1.0")
	s.registerMethod("get_camera_status", s.methodGetCameraStatus, "1.0")

	s.logger.WithField("action", "register_methods").Info("Built-in methods registered")
}

// registerMethod registers a JSON-RPC method handler
func (s *WebSocketServer) registerMethod(name string, handler MethodHandler, version string) {
	s.methodsMutex.Lock()
	defer s.methodsMutex.Unlock()

	s.methods[name] = handler
	s.methodVersions[name] = version

	s.logger.WithFields(logrus.Fields{
		"method":  name,
		"version": version,
		"action":  "register_method",
	}).Debug("Method registered")
}

// methodPing implements the ping method
// Following Python _method_ping implementation
func (s *WebSocketServer) methodPing(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "ping",
		"action":    "method_call",
	}).Debug("Ping method called")

	// Return "pong" as specified in API documentation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  "pong",
	}, nil
}

// methodAuthenticate implements the authenticate method
// Following Python _method_authenticate implementation
func (s *WebSocketServer) methodAuthenticate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "authenticate",
		"action":    "method_call",
	}).Debug("Authenticate method called")

	// Extract auth_token parameter
	authToken, ok := params["auth_token"].(string)
	if !ok || authToken == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "auth_token parameter is required",
			},
		}, nil
	}

	// Validate JWT token
	claims, err := s.jwtHandler.ValidateToken(authToken)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "authenticate",
			"action":    "authentication_failed",
		}).Warn("Authentication failed")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    AUTHENTICATION_REQUIRED,
				Message: "Authentication failed",
				Data: map[string]interface{}{
					"reason": "Invalid or expired token",
				},
			},
		}, nil
	}

	// Update client authentication state
	client.Authenticated = true
	client.UserID = claims.UserID
	client.Role = claims.Role
	client.AuthMethod = "jwt"

	// Calculate expiration time
	expiresAt := time.Unix(claims.EXP, 0)

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"user_id":   claims.UserID,
		"role":      claims.Role,
		"method":    "authenticate",
		"action":    "authentication_success",
	}).Info("Authentication successful")

	// Return authentication result following Python implementation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"authenticated": true,
			"role":          claims.Role,
			"permissions":   getPermissionsForRole(claims.Role),
			"expires_at":    expiresAt.Format(time.RFC3339),
			"session_id":    client.ClientID,
		},
	}, nil
}

// methodGetCameraList implements the get_camera_list method
// Following Python _method_get_camera_list implementation
func (s *WebSocketServer) methodGetCameraList(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_camera_list",
		"action":    "method_call",
	}).Debug("Get camera list method called")

	// Check authentication
	if !client.Authenticated {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    AUTHENTICATION_REQUIRED,
				Message: ErrorMessages[AUTHENTICATION_REQUIRED],
			},
		}, nil
	}

	// Get camera list from camera monitor
	cameras := s.cameraMonitor.GetConnectedCameras()
	
	// Convert camera list to response format
	cameraList := make([]map[string]interface{}, 0, len(cameras))
	connectedCount := 0

	for devicePath, camera := range cameras {
		// Get resolution from first format if available
		resolution := ""
		if len(camera.Formats) > 0 {
			format := camera.Formats[0]
			resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
		}

		cameraData := map[string]interface{}{
			"device":     devicePath,
			"status":     string(camera.Status),
			"name":       camera.Name,
			"resolution": resolution,
			"fps":        30, // Default FPS - can be enhanced later
			"streams":    make(map[string]string), // Empty streams for now
		}

		cameraList = append(cameraList, cameraData)

		if camera.Status == "CONNECTED" {
			connectedCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":      client.ClientID,
		"method":         "get_camera_list",
		"total_cameras":  len(cameras),
		"connected":      connectedCount,
		"action":         "camera_list_success",
	}).Debug("Camera list retrieved successfully")

	// Return camera list following Python implementation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"cameras":   cameraList,
			"total":     len(cameras),
			"connected": connectedCount,
		},
	}, nil
}

// methodGetCameraStatus implements the get_camera_status method
// Following Python _method_get_camera_status implementation
func (s *WebSocketServer) methodGetCameraStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_camera_status",
		"action":    "method_call",
	}).Debug("Get camera status method called")

	// Check authentication
	if !client.Authenticated {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    AUTHENTICATION_REQUIRED,
				Message: ErrorMessages[AUTHENTICATION_REQUIRED],
			},
		}, nil
	}

	// Extract device parameter
	device, ok := params["device"].(string)
	if !ok || device == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	// Get camera status from camera monitor
	camera, exists := s.cameraMonitor.GetDevice(device)
	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "Camera not found",
				Data:    fmt.Sprintf("Camera device %s not found", device),
			},
		}, nil
	}

	// Get resolution from first format if available
	resolution := ""
	if len(camera.Formats) > 0 {
		format := camera.Formats[0]
		resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"device":    device,
		"method":    "get_camera_status",
		"status":    string(camera.Status),
		"action":    "camera_status_success",
	}).Debug("Camera status retrieved successfully")

	// Return camera status following Python implementation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"device":       device,
			"status":       string(camera.Status),
			"name":         camera.Name,
			"resolution":   resolution,
			"fps":          30, // Default FPS - can be enhanced later
			"streams":      make(map[string]string), // Empty streams for now
			"metrics":      make(map[string]interface{}), // Empty metrics for now
			"capabilities": camera.Capabilities,
		},
	}, nil
}

// getPermissionsForRole returns permissions for a given role
// Following Python role-based access control patterns
func getPermissionsForRole(role string) []string {
	switch role {
	case "admin":
		return []string{"view", "control", "admin"}
	case "operator":
		return []string{"view", "control"}
	case "viewer":
		return []string{"view"}
	default:
		return []string{}
	}
}
