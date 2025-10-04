package websocket

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// methodWrapper provides common method execution pattern with logging and error handling.
func (s *WebSocketServer) methodWrapper(methodName string, handler func() (interface{}, error)) func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    methodName,
			"action":    "method_call",
		}).Debug(fmt.Sprintf("%s method called", methodName))

		result, err := handler()
		if err != nil {
			s.logger.WithFields(logging.Fields{
				"client_id": client.ClientID,
				"method":    methodName,
				"action":    "method_error",
				"error":     err.Error(),
			}).Error(fmt.Sprintf("%s method failed", methodName))

			// Enhanced error translation based on error content
			jsonRpcError := s.translateErrorToJsonRpc(err, methodName)
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   jsonRpcError,
			}, nil
		}

		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    methodName,
			"action":    "method_success",
		}).Debug(fmt.Sprintf("%s method completed successfully", methodName))

		s.assertResponseFields(methodName, result)
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Result:  result,
		}, nil
	}
}

// assertResponseFields performs lightweight dev-only checks against API spec
func (s *WebSocketServer) assertResponseFields(method string, result interface{}) {
	if s == nil || s.logger == nil {
		return
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		return
	}
	required := map[string][]string{
		"get_camera_status": {"device", "status", "name", "resolution", "fps", "streams", "metrics"},
		"get_camera_list":   {"cameras", "total", "connected"},
		"get_stream_url":    {"device", "stream_name", "stream_url", "available"},
		"get_stream_status": {"device", "stream_name", "status", "ready"},
		"take_snapshot":     {"device", "filename", "status", "timestamp", "file_size"},
		"start_recording":   {"device", "filename", "status", "start_time", "format"},
		"stop_recording":    {"device", "status"},
		"start_streaming":   {"device", "stream_name", "stream_url", "status", "start_time"},
		"stop_streaming":    {"device", "stream_name", "status", "end_time", "duration"},
		"list_recordings":   {"files", "total", "limit", "offset"},
		"list_snapshots":    {"files", "total", "limit", "offset"},
		"get_storage_info":  {"total_space", "used_space", "available_space", "usage_percentage", "recordings_size", "snapshots_size", "low_space_warning"},
	}
	if fields, exists := required[method]; exists {
		missing := []string{}
		for _, k := range fields {
			if _, ok := m[k]; !ok {
				missing = append(missing, k)
			}
		}
		if len(missing) > 0 {
			s.logger.WithFields(logging.Fields{
				"method":  method,
				"missing": missing,
			}).Warn("Response missing documented fields (dev-only check)")
		}
	}
}

// authenticatedMethodWrapper wraps methods that require authentication using centralized security check
func (s *WebSocketServer) authenticatedMethodWrapper(methodName string, handler func() (interface{}, error)) func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		// Centralized authentication check - replaces 20+ duplicate checks
		if !client.Authenticated {
			s.logger.WithFields(logging.Fields{
				"client_id": client.ClientID,
				"method":    methodName,
				"action":    "auth_required",
				"component": "security_middleware",
			}).Warn("Authentication required for method")

			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "Authentication required", "Authenticate first"),
			}, nil
		}

		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"user_id":   client.UserID,
			"role":      client.Role,
			"method":    methodName,
			"action":    "auth_success",
			"component": "security_middleware",
		}).Debug("Authentication check passed")

		// Call base method wrapper for common logging and error handling
		return s.methodWrapper(methodName, handler)(params, client)
	}
}

// registerBuiltinMethods registers all built-in JSON-RPC methods
func (s *WebSocketServer) registerBuiltinMethods() {
	// Core methods
	s.registerMethod("ping", s.MethodPing, "1.0")
	s.registerMethod("authenticate", s.MethodAuthenticate, "1.0")
	s.registerMethod("get_camera_list", s.MethodGetCameraList, "1.0")
	s.registerMethod("get_camera_status", s.MethodGetCameraStatus, "1.0")

	// System methods
	s.registerMethod("get_metrics", s.MethodGetMetrics, "1.0")
	s.registerMethod("get_camera_capabilities", s.MethodGetCameraCapabilities, "1.0")
	s.registerMethod("get_status", s.MethodGetStatus, "1.0")
	s.registerMethod("get_server_info", s.MethodGetServerInfo, "1.0")
	s.registerMethod("get_streams", s.MethodGetStreams, "1.0")
	s.registerMethod("get_system_status", s.MethodGetSystemStatus, "1.0")

	// File management methods
	s.registerMethod("list_recordings", s.MethodListRecordings, "1.0")
	s.registerMethod("list_snapshots", s.MethodListSnapshots, "1.0")
	s.registerMethod("get_recording_info", s.MethodGetRecordingInfo, "1.0")
	s.registerMethod("get_snapshot_info", s.MethodGetSnapshotInfo, "1.0")
	s.registerMethod("delete_recording", s.MethodDeleteRecording, "1.0")
	s.registerMethod("delete_snapshot", s.MethodDeleteSnapshot, "1.0")
	s.registerMethod("get_storage_info", s.MethodGetStorageInfo, "1.0")
	s.registerMethod("set_retention_policy", s.MethodSetRetentionPolicy, "1.0")
	s.registerMethod("cleanup_old_files", s.MethodCleanupOldFiles, "1.0")

	// Recording and snapshot methods
	s.registerMethod("take_snapshot", s.MethodTakeSnapshot, "1.0")
	s.registerMethod("start_recording", s.MethodStartRecording, "1.0")
	s.registerMethod("stop_recording", s.MethodStopRecording, "1.0")

	// Streaming methods
	s.registerMethod("start_streaming", s.MethodStartStreaming, "1.0")
	s.registerMethod("stop_streaming", s.MethodStopStreaming, "1.0")
	s.registerMethod("get_stream_url", s.MethodGetStreamURL, "1.0")
	s.registerMethod("get_stream_status", s.MethodGetStreamStatus, "1.0")

	// ARCHITECTURE FIX: Notification methods are server-generated only
	// These methods exist for internal server use but are NOT callable by clients
	// camera_status_update and recording_status_update are notifications only
	// They are handled by the WebSocket notification system, not as callable methods

	// Event subscription methods
	s.registerMethod("subscribe_events", s.MethodSubscribeEvents, "1.0")
	s.registerMethod("unsubscribe_events", s.MethodUnsubscribeEvents, "1.0")
	s.registerMethod("get_subscription_stats", s.MethodGetSubscriptionStats, "1.0")

	// External stream discovery methods
	s.registerMethod("discover_external_streams", s.MethodDiscoverExternalStreams, "1.0")
	s.registerMethod("add_external_stream", s.MethodAddExternalStream, "1.0")
	s.registerMethod("remove_external_stream", s.MethodRemoveExternalStream, "1.0")
	s.registerMethod("get_external_streams", s.MethodGetExternalStreams, "1.0")
	s.registerMethod("set_discovery_interval", s.MethodSetDiscoveryInterval, "1.0")

	s.logger.WithField("action", "register_methods").Info("Built-in methods registered")
}

// registerMethod registers a JSON-RPC method handler
func (s *WebSocketServer) registerMethod(name string, handler MethodHandler, version string) {
	// Wrap the handler to ensure security, readiness, and metrics are always applied
	wrappedHandler := func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		startTime := time.Now()

		// Apply security checks
		if err := s.checkRateLimit(client); err != nil {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(RATE_LIMIT_EXCEEDED, "rate_limit", err.Error(), "Reduce request rate or wait"),
			}, nil
		}

		if err := s.checkMethodPermissions(client, name); err != nil {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(INSUFFICIENT_PERMISSIONS, "insufficient_permissions", err.Error(), "Check user role and permissions"),
			}, nil
		}

		// Progressive Readiness: Check controller readiness for non-authentication methods
		if name != "authenticate" && name != "ping" {
			if !s.isSystemReady() {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					Error:   NewJsonRpcError(MEDIAMTX_UNAVAILABLE, "service_initializing", "Service is still initializing, please retry", "Wait for service to complete startup"),
				}, nil
			}
		}

		// Call the original handler
		response, err := handler(params, client)

		// Record metrics
		duration := time.Since(startTime).Seconds()
		s.recordRequest(name, duration)

		// Handle errors
		if err != nil {
			// Use atomic operation for ErrorCount
			atomic.AddInt64(&s.metrics.ErrorCount, 1)
		}

		return response, err
	}

	// Store method handler in mutex-protected map
	s.logger.WithFields(logging.Fields{
		"method":       name,
		"handler_type": fmt.Sprintf("%T", wrappedHandler),
		"action":       "storing_method",
	}).Info("Storing method handler in map")

	s.methodsMutex.Lock()
	s.methods[name] = wrappedHandler
	s.methodsMutex.Unlock()

	// Verify storage
	s.methodsMutex.RLock()
	if stored, exists := s.methods[name]; exists {
		s.logger.WithFields(logging.Fields{
			"method":      name,
			"stored_type": fmt.Sprintf("%T", stored),
			"action":      "verification_success",
		}).Info("Method handler stored successfully")
	} else {
		s.logger.WithFields(logging.Fields{
			"method": name,
			"action": "verification_failed",
		}).Error("Failed to store method handler")
	}
	s.methodsMutex.RUnlock()

	// Store method version (still needs mutex for map operations)
	s.methodVersionsMutex.Lock()
	s.methodVersions[name] = version
	s.methodVersionsMutex.Unlock()

	s.logger.WithFields(logging.Fields{
		"method":  name,
		"version": version,
		"action":  "register_method",
	}).Debug("Method registered with security and metrics wrapper")
}

// MethodPing implements the ping method
// Authentication: Not required (per API documentation)
// Purpose: Connectivity + envelope sanity check before authenticate
func (s *WebSocketServer) MethodPing(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Record performance metrics
	startTime := time.Now()
	duration := time.Since(startTime).Seconds()
	s.recordRequest("ping", duration)

	// Return "pong" as specified in API documentation
	// No authentication required - this is the only method that works without auth
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  "pong",
		ID:      nil, // Will be set by the server from the original request
	}, nil
}

// MethodAuthenticate implements the authenticate method
func (s *WebSocketServer) MethodAuthenticate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Handle authentication errors directly to return proper error codes
	// Extract auth_token parameter
	authToken, ok := params["auth_token"].(string)
	if !ok || authToken == "" {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    "authenticate",
			"action":    "auth_error",
			"error":     "auth_token parameter is required",
		}).Warn("Authentication failed: missing auth_token parameter")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "auth_token parameter is required", "Provide valid auth_token"),
		}, nil
	}

	// Validate JWT token
	claims, err := s.jwtHandler.ValidateToken(authToken)
	if err != nil {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    "authenticate",
			"action":    "auth_error",
			"error":     err.Error(),
		}).Warn("JWT token validation failed")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_failed", "Invalid or expired token", "Provide valid token"),
		}, nil
	}

	// Update client authentication state
	client.Authenticated = true
	client.UserID = claims.UserID
	client.Role = claims.Role
	client.AuthMethod = "jwt"

	// Calculate expiration time
	expiresAt := time.Unix(claims.EXP, 0)

	// Record performance metrics
	startTime := time.Now()
	duration := time.Since(startTime).Seconds()
	s.recordRequest("authenticate", duration)

	s.logger.WithFields(logging.Fields{
		"client_id": client.ClientID,
		"user_id":   client.UserID,
		"role":      client.Role,
		"method":    "authenticate",
		"action":    "auth_success",
	}).Info("Authentication successful")

	// Return authentication result following Python implementation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"authenticated": true,
			"role":          claims.Role,
			"permissions":   s.permissionChecker.GetPermissionsForRole(claims.Role), // Delegate to security module
			"expires_at":    expiresAt.Format(time.RFC3339),
			"session_id":    client.ClientID,
		},
	}, nil
}

// MethodGetCameraList implements the get_camera_list method
func (s *WebSocketServer) MethodGetCameraList(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Delegates to MediaMTX Controller for business logic
	return s.authenticatedMethodWrapper("get_camera_list", func() (interface{}, error) {
		// STRICT PARAMETER VALIDATION: get_camera_list should accept NO parameters
		if len(params) > 0 {
			// Log invalid parameters for debugging
			s.logger.WithFields(logging.Fields{
				"client_id": client.ClientID,
				"method":    "get_camera_list",
				"params":    params,
				"action":    "invalid_params",
			}).Warn("get_camera_list received unexpected parameters")

			return nil, fmt.Errorf("get_camera_list accepts no parameters, received: %v", params)
		}

		// Delegate to MediaMTX controller - returns API-ready APICameraInfo format
		cameraListResponse, err := s.mediaMTXController.GetCameraList(context.Background())
		if err != nil {
			return nil, err
		}

		// MediaMTX Controller handles the API formatting through PathManager abstraction
		// Simply return the API-ready response
		return cameraListResponse, nil
	})(params, client)
}

func (s *WebSocketServer) MethodGetCameraStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// REFACTORED: 56 lines → 15 lines (SRP compliance + centralized error handling)
	return s.authenticatedMethodWrapper("get_camera_status", func() (interface{}, error) {
		// Validate device parameter using centralized validation
		validationResult := s.validationHelper.ValidateDeviceParameter(params)
		if !validationResult.Valid {
			s.validationHelper.LogValidationWarnings(validationResult, "get_camera_status", client.ClientID)
			return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
		}

		// Extract device parameter
		cameraID, ok := params["device"].(string)
		if !ok || cameraID == "" {
			return nil, fmt.Errorf("device parameter is required")
		}

		// Delegate to controller - let wrapper handle error translation
		return s.mediaMTXController.GetCameraStatus(context.Background(), cameraID)
	})(params, client)
}

// MethodGetMetrics implements the get_metrics method
// Thin delegation - Controller returns API-ready GetMetricsResponse
func (s *WebSocketServer) MethodGetMetrics(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Uses wrapper helpers for consistent method execution
	return s.authenticatedMethodWrapper("get_metrics", func() (interface{}, error) {
		// Pure delegation to Controller - returns complete API-ready response
		return s.mediaMTXController.GetMetrics(context.Background())
	})(params, client)
}

func (s *WebSocketServer) MethodGetCameraCapabilities(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Centralized authentication check
	if !client.Authenticated {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    "get_camera_capabilities",
			"action":    "auth_required",
			"component": "security_middleware",
		}).Warn("Authentication required for method")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "Authentication required", "Authenticate first"),
		}, nil
	}

	// Validate device parameter using centralized validation
	validationResult := s.validationHelper.ValidateDeviceParameter(params)
	if !validationResult.Valid {
		// Log validation warnings for debugging
		s.validationHelper.LogValidationWarnings(validationResult, "get_camera_capabilities", client.ClientID)
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INVALID_PARAMS, "invalid_params", fmt.Sprintf("validation failed: %v", validationResult.Errors), "Provide valid device parameter"),
		}, nil
	}

	// Extract validated device parameter
	device := validationResult.Data["device"].(string)

	// Pure delegation to Controller - returns API-ready GetCameraCapabilitiesResponse
	capabilitiesResponse, err := s.mediaMTXController.GetCameraCapabilities(context.Background(), device)
	if err != nil {
		// ✅ FIX 1: Map camera-specific errors to proper API error codes
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not available") {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(CAMERA_NOT_FOUND, "camera_not_found", "Camera not found or disconnected", "Check camera identifier"),
			}, nil
		}
		// For other errors, return internal error
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INTERNAL_ERROR, "internal_error", fmt.Sprintf("camera '%s' capabilities error: %v", device, err), "Retry or contact support if persistent"),
		}, nil
	}

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  capabilitiesResponse,
	}, nil
}

// MethodGetStatus implements the get_status method
func (s *WebSocketServer) MethodGetStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Uses wrapper helpers for consistent method execution
	return s.authenticatedMethodWrapper("get_status", func() (interface{}, error) {

		// Pure delegation to MediaMTX controller - returns API-ready response with comprehensive health status
		if s.mediaMTXController != nil {
			health, err := s.mediaMTXController.GetHealth(context.Background())
			if err != nil {
				return nil, fmt.Errorf("failed to get health status: %w", err)
			}

			// Return the health response directly from the controller
			return map[string]interface{}{
				"status":     health.Status,
				"uptime":     health.Uptime,
				"version":    health.Version,
				"components": health.Components,
			}, nil
		}

		// Fallback if controller is not available
		return map[string]interface{}{
			"status":  "UNHEALTHY",
			"uptime":  float64(0),
			"version": "unknown",
			"components": map[string]interface{}{
				"websocket_server": "error",
				"camera_monitor":   "error",
				"mediamtx":         "error",
			},
		}, nil
	})(params, client)
}

// MethodGetSystemStatus implements the get_system_status method
// Returns detailed system readiness information
func (s *WebSocketServer) MethodGetSystemStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Uses wrapper helpers for consistent method execution
	return s.authenticatedMethodWrapper("get_system_status", func() (interface{}, error) {
		// Return system readiness response - no additional processing needed
		return s.getSystemReadinessResponse(), nil
	})(params, client)
}

// MethodGetServerInfo implements the get_server_info method
func (s *WebSocketServer) MethodGetServerInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Uses wrapper helpers for consistent method execution
	return s.authenticatedMethodWrapper("get_server_info", func() (interface{}, error) {
		// Pure delegation to Controller - returns API-ready GetServerInfoResponse
		return s.mediaMTXController.GetServerInfo(context.Background())
	})(params, client)
}

// MethodGetStreams implements the get_streams method
func (s *WebSocketServer) MethodGetStreams(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Uses wrapper helpers for consistent method execution
	return s.authenticatedMethodWrapper("get_streams", func() (interface{}, error) {
		// Delegate to MediaMTX controller - investigate what it returns
		streams, err := s.mediaMTXController.GetStreams(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get streams from MediaMTX service: %v", err)
		}

		// Return Controller's API-ready response directly - thin delegation
		return streams, nil
	})(params, client)
}

// MethodListRecordings implements the list_recordings method
func (s *WebSocketServer) MethodListRecordings(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("list_recordings", func() (interface{}, error) {
		// Extract parameters WITHOUT defaults
		limit := 0  // No default!
		offset := 0 // No default!

		if limitParam, exists := params["limit"]; exists {
			if limitVal, ok := limitParam.(float64); ok {
				limit = int(limitVal)
			}
		}

		if offsetParam, exists := params["offset"]; exists {
			if offsetVal, ok := offsetParam.(float64); ok {
				offset = int(offsetVal)
			}
		}

		// Pure delegation - RecordingManager handles defaults
		fileList, err := s.mediaMTXController.ListRecordings(context.Background(), limit, offset)
		if err != nil {
			return nil, fmt.Errorf("error getting recordings list: %v", err)
		}

		// Return Controller's API-ready response directly - thin delegation
		return fileList, nil
	})(params, client)
}

// MethodDeleteRecording implements the delete_recording method
func (s *WebSocketServer) MethodDeleteRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// REFACTORED: 81 lines → 20 lines → 14 lines (SRP compliance + centralized error handling)
	return s.authenticatedMethodWrapper("delete_recording", func() (interface{}, error) {

		// Validate parameters
		if params == nil {
			return nil, fmt.Errorf("filename parameter is required")
		}

		filename, ok := params["filename"].(string)
		if !ok || filename == "" {
			return nil, fmt.Errorf("filename must be a non-empty string")
		}

		// Delegate to controller - let wrapper handle error translation
		err := s.mediaMTXController.DeleteRecording(context.Background(), filename)
		if err != nil {
			return nil, fmt.Errorf("error deleting recording: %v", err)
		}

		// Return success response
		return map[string]interface{}{
			"filename": filename,
			"deleted":  true,
			"message":  "Recording file deleted successfully",
		}, nil
	})(params, client)
}

// MethodDeleteSnapshot implements the delete_snapshot method
func (s *WebSocketServer) MethodDeleteSnapshot(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Centralized authentication check
	if !client.Authenticated {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    "delete_snapshot",
			"action":    "auth_required",
			"component": "security_middleware",
		}).Warn("Authentication required for method")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "Authentication required", "Authenticate first"),
		}, nil
	}

	// Validate filename parameter
	validationResult := s.validationHelper.ValidateFilenameParameter(params)
	if !validationResult.Valid {
		s.validationHelper.LogValidationWarnings(validationResult, "delete_snapshot", client.ClientID)
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INVALID_PARAMS, "invalid_params", "validation failed", "Provide valid filename parameter"),
		}, nil
	}

	// Extract validated filename
	filename := validationResult.Data["filename"].(string)

	// Use MediaMTX controller to delete snapshot - thin delegation
	err := s.mediaMTXController.DeleteSnapshot(context.Background(), filename)
	if err != nil {
		// Map specific errors to JSON-RPC error codes
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(FILE_NOT_FOUND, "file_not_found", "File not found or inaccessible", "Verify filename"),
			}, nil
		}
		// For other errors, return internal error
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INTERNAL_ERROR, "internal_error", fmt.Sprintf("error deleting snapshot: %v", err), "Retry or contact support if persistent"),
		}, nil
	}

	// Return success response
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"filename": filename,
			"deleted":  true,
			"message":  "Snapshot file deleted successfully",
		},
	}, nil
}

func (s *WebSocketServer) MethodGetStorageInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("get_storage_info", func() (interface{}, error) {

		// Get storage info from controller (thin delegation)
		info, err := s.mediaMTXController.GetStorageInfo(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error getting storage information: %v", err)
		}

		// Return Controller's API-ready response directly - thin delegation
		return info, nil
	})(params, client)
}

func (s *WebSocketServer) MethodCleanupOldFiles(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {

	return s.authenticatedMethodWrapper("cleanup_old_files", func() (interface{}, error) {
		// Delegate to Controller for cleanup logic (single source of truth)
		result, err := s.mediaMTXController.CleanupOldFiles(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to cleanup old files: %v", err)
		}
		return result, nil
	})(params, client)
}

// MethodSetRetentionPolicy implements the set_retention_policy method
func (s *WebSocketServer) MethodSetRetentionPolicy(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("set_retention_policy", func() (interface{}, error) {
		// Validate retention policy parameters
		validationResult := s.validationHelper.ValidateRetentionPolicyParameters(params)
		if !validationResult.Valid {
			s.validationHelper.LogValidationWarnings(validationResult, "set_retention_policy", client.ClientID)
			return nil, fmt.Errorf("validation failed: %s", "validation failed")
		}

		// Extract validated parameters
		policyType := validationResult.Data["policy_type"].(string)
		enabled := validationResult.Data["enabled"].(bool)

		// Delegate to Controller for retention policy logic (single source of truth)
		result, err := s.mediaMTXController.SetRetentionPolicy(context.Background(), enabled, policyType, params)
		if err != nil {
			return nil, fmt.Errorf("failed to set retention policy: %v", err)
		}
		return result, nil
	})(params, client)
}

func (s *WebSocketServer) MethodListSnapshots(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("list_snapshots", func() (interface{}, error) {

		// Validate pagination parameters
		validationResult := s.validationHelper.ValidatePaginationParams(params)
		if !validationResult.Valid {
			s.validationHelper.LogValidationWarnings(validationResult, "list_snapshots", client.ClientID)
			return nil, fmt.Errorf("validation failed: %s", "validation failed")
		}

		// Extract validated parameters
		limit := validationResult.Data["limit"].(int)
		offset := validationResult.Data["offset"].(int)

		// Use MediaMTX controller to get snapshots list - thin delegation
		fileList, err := s.mediaMTXController.ListSnapshots(context.Background(), limit, offset)
		if err != nil {
			return nil, fmt.Errorf("error getting snapshots list: %v", err)
		}

		// Check if no snapshots found
		if fileList.Total == 0 {
			// API doc: return success with empty result object, not an error
			return map[string]interface{}{
				"files":  []map[string]interface{}{},
				"total":  0,
				"limit":  limit,
				"offset": offset,
			}, nil
		}

		// Convert SnapshotFileInfo to map for JSON response
		files := make([]map[string]interface{}, len(fileList.Snapshots))
		for i, file := range fileList.Snapshots {
			fileData := map[string]interface{}{
				"filename":      file.Filename,
				"file_size":     file.FileSize,
				"modified_time": file.ModifiedTime, // API compliant field name
				"download_url":  file.DownloadURL,
			}

			files[i] = fileData
		}

		// Return snapshots list
		return map[string]interface{}{
			"files":  files,
			"total":  fileList.Total,
			"limit":  fileList.Limit,
			"offset": fileList.Offset,
		}, nil
	})(params, client)
}

// MethodTakeSnapshot implements the take_snapshot method
func (s *WebSocketServer) MethodTakeSnapshot(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("take_snapshot", func() (interface{}, error) {
		// 1. Input validation only (API contract validation)
		validationResult := s.validationHelper.ValidateSnapshotParameters(params)
		if !validationResult.Valid {
			s.validationHelper.LogValidationWarnings(validationResult, "take_snapshot", client.ClientID)
			return nil, fmt.Errorf("validation failed")
		}

		// 2. Extract parameters and convert to strongly-typed options
		devicePath := validationResult.Data["device"].(string)
		optionsMap := validationResult.Data["options"].(map[string]interface{})

		// Convert map to strongly-typed SnapshotOptions for type safety
		options := mediamtx.SnapshotOptionsFromMap(optionsMap)

		// 3. Pure delegation - Controller and SnapshotManager handle all business logic
		// No duplicate validation, no response formatting, no business logic
		snapshot, err := s.mediaMTXController.TakeAdvancedSnapshot(context.Background(), devicePath, options)
		if err != nil {
			return nil, fmt.Errorf("failed to take snapshot: %v", err)
		}

		// 4. Return snapshot as-is - SnapshotManager should provide API-ready response
		return snapshot, nil
	})(params, client)
}

// MethodStartRecording implements the start_recording method
func (s *WebSocketServer) MethodStartRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("start_recording", func() (interface{}, error) {
		// Pure delegation - pass raw params to controller
		// Controller handles all parameter extraction and business logic
		return s.mediaMTXController.StartRecording(context.Background(), params)
	})(params, client)
}

// MethodStopRecording implements the stop_recording method
func (s *WebSocketServer) MethodStopRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("stop_recording", func() (interface{}, error) {
		// Validate parameters
		if params == nil {
			return nil, fmt.Errorf("device parameter is required")
		}

		cameraID, ok := params["device"].(string)
		if !ok || cameraID == "" {
			return nil, fmt.Errorf("device parameter is required")
		}

		// Validate device parameter using centralized validation
		val := s.validationHelper.ValidateDeviceParameter(map[string]interface{}{"device": cameraID})
		if !val.Valid {
			return nil, fmt.Errorf("invalid device parameter: %v", val.Errors)
		}

		// Pure delegation to Controller - returns API-ready StopRecordingResponse
		return s.mediaMTXController.StopRecording(context.Background(), cameraID)
	})(params, client)
}

func (s *WebSocketServer) MethodGetRecordingInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// Centralized authentication check
	if !client.Authenticated {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    "get_recording_info",
			"action":    "auth_required",
			"component": "security_middleware",
		}).Warn("Authentication required for method")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "Authentication required", "Authenticate first"),
		}, nil
	}

	// Validate filename parameter
	validationResult := s.validationHelper.ValidateFilenameParameter(params)
	if !validationResult.Valid {
		s.validationHelper.LogValidationWarnings(validationResult, "get_recording_info", client.ClientID)
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INVALID_PARAMS, "invalid_params", "validation failed", "Provide valid filename parameter"),
		}, nil
	}

	// Extract validated filename parameter
	filename := validationResult.Data["filename"].(string)

	// Pure delegation to Controller - returns API-ready GetRecordingInfoResponse
	recordingInfo, err := s.mediaMTXController.GetRecordingInfo(context.Background(), filename)
	if err != nil {
		// Map specific errors to JSON-RPC error codes
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   NewJsonRpcError(FILE_NOT_FOUND, "file_not_found", "File not found or inaccessible", "Verify filename"),
			}, nil
		}
		// For other errors, return internal error
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error:   NewJsonRpcError(INTERNAL_ERROR, "internal_error", fmt.Sprintf("error getting recording info: %v", err), "Retry or contact support if persistent"),
		}, nil
	}

	// Return success response
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  recordingInfo,
	}, nil
}

// MethodGetSnapshotInfo implements the get_snapshot_info method
func (s *WebSocketServer) MethodGetSnapshotInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("get_snapshot_info", func() (interface{}, error) {

		// Validate filename parameter
		validationResult := s.validationHelper.ValidateFilenameParameter(params)
		if !validationResult.Valid {
			s.validationHelper.LogValidationWarnings(validationResult, "get_snapshot_info", client.ClientID)
			return nil, fmt.Errorf("validation failed: %s", "validation failed")
		}

		// Extract validated filename parameter
		filename := validationResult.Data["filename"].(string)

		// Pure delegation to Controller - returns API-ready GetSnapshotInfoResponse
		return s.mediaMTXController.GetSnapshotInfo(context.Background(), filename)
	})(params, client)
}

// ARCHITECTURE FIX: These are notification handlers, not callable methods
// They are used internally by the server notification system
// They should NEVER be called by clients - they are server-generated notifications only

// handleCameraStatusUpdateNotification processes camera status update notifications
// This is called internally by the server notification system
func (s *WebSocketServer) handleCameraStatusUpdateNotification(params map[string]interface{}) {
	// REQ-API-020: WebSocket server shall support camera_status_update notifications
	// REQ-API-021: Notifications shall include device, status, name, resolution, fps, and streams

	s.logger.WithFields(logging.Fields{
		"method": "camera_status_update",
		"params": params,
	}).Info("Processing camera status update notification")

	// This is handled by the WebSocket notification system
	// The actual notification sending is done by notifyCameraStatusUpdate()
}

// handleRecordingStatusUpdateNotification processes recording status update notifications
// This is called internally by the server notification system
func (s *WebSocketServer) handleRecordingStatusUpdateNotification(params map[string]interface{}) {
	// REQ-API-022: WebSocket server shall support recording_status_update notifications
	// REQ-API-023: Notifications shall include device, status, filename, and duration

	s.logger.WithFields(logging.Fields{
		"method": "recording_status_update",
		"params": params,
	}).Info("Processing recording status update notification")

	// This is handled by the WebSocket notification system
	// The actual notification sending is done by notifyRecordingStatusUpdate()
}

// MethodSubscribeEvents handles client subscription to event topics
func (s *WebSocketServer) MethodSubscribeEvents(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("subscribe_events", func() (interface{}, error) {
		// Validate required parameters
		topicsParam, exists := params["topics"]
		if !exists {
			return nil, fmt.Errorf("topics parameter is required")
		}

		// Parse topics parameter
		var topics []EventTopic
		switch v := topicsParam.(type) {
		case []interface{}:
			for _, topic := range v {
				if topicStr, ok := topic.(string); ok {
					topics = append(topics, EventTopic(topicStr))
				}
			}
		case []string:
			for _, topic := range v {
				topics = append(topics, EventTopic(topic))
			}
		default:
			return nil, fmt.Errorf("topics must be an array of strings")
		}

		if len(topics) == 0 {
			return nil, fmt.Errorf("at least one topic must be specified")
		}

		// Parse optional filters
		var filters map[string]interface{}
		if filtersParam, exists := params["filters"]; exists {
			if filtersMap, ok := filtersParam.(map[string]interface{}); ok {
				filters = filtersMap
			}
		}

		// Subscribe client to events
		err := s.eventManager.Subscribe(client.ClientID, topics, filters)
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to events: %v", err)
		}

		// Update client last seen
		s.eventManager.UpdateClientLastSeen(client.ClientID)

		// Return subscription result
		return map[string]interface{}{
			"subscribed": true,
			"topics":     topics,
			"filters":    filters,
		}, nil
	})(params, client)
}

// MethodUnsubscribeEvents handles client unsubscription from event topics
func (s *WebSocketServer) MethodUnsubscribeEvents(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("unsubscribe_events", func() (interface{}, error) {
		// Parse optional topics parameter (if not provided, unsubscribe from all)
		var topics []EventTopic
		if topicsParam, exists := params["topics"]; exists {
			switch v := topicsParam.(type) {
			case []interface{}:
				for _, topic := range v {
					if topicStr, ok := topic.(string); ok {
						topics = append(topics, EventTopic(topicStr))
					}
				}
			case []string:
				for _, topic := range v {
					topics = append(topics, EventTopic(topic))
				}
			}
		}

		// Unsubscribe client from events
		err := s.eventManager.Unsubscribe(client.ClientID, topics)
		if err != nil {
			return nil, fmt.Errorf("failed to unsubscribe: %v", err)
		}

		// Update client last seen
		s.eventManager.UpdateClientLastSeen(client.ClientID)

		// Return unsubscription result
		return map[string]interface{}{
			"unsubscribed": true,
			"topics":       topics,
		}, nil
	})(params, client)
}

// MethodGetSubscriptionStats returns statistics about event subscriptions
func (s *WebSocketServer) MethodGetSubscriptionStats(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.methodWrapper("get_subscription_stats", func() (interface{}, error) {
		// Get subscription statistics
		stats := s.eventManager.GetSubscriptionStats()

		// Get client's own subscriptions
		clientTopics := s.eventManager.GetClientSubscriptions(client.ClientID)

		return map[string]interface{}{
			"global_stats":  stats,
			"client_topics": clientTopics,
			"client_id":     client.ClientID,
		}, nil
	})(params, client)
}

// MethodStartStreaming starts a live streaming session for the specified camera device
func (s *WebSocketServer) MethodStartStreaming(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("start_streaming", func() (interface{}, error) {
		// Validate device parameter using centralized validation
		validationResult := s.validationHelper.ValidateDeviceParameter(params)
		if !validationResult.Valid {
			// Log validation warnings for debugging
			s.validationHelper.LogValidationWarnings(validationResult, "start_streaming", client.ClientID)
			return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
		}

		// Extract validated device parameter
		device := validationResult.Data["device"].(string)

		// Pure delegation to Controller - returns API-ready GetStreamURLResponse
		return s.mediaMTXController.StartStreaming(context.Background(), device)
	})(params, client)
}

// MethodStopStreaming stops the active streaming session for the specified camera device
func (s *WebSocketServer) MethodStopStreaming(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("stop_streaming", func() (interface{}, error) {
		// Validate device parameter using centralized validation
		validationResult := s.validationHelper.ValidateDeviceParameter(params)
		if !validationResult.Valid {
			// Log validation warnings for debugging
			s.validationHelper.LogValidationWarnings(validationResult, "stop_streaming", client.ClientID)
			return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
		}

		// Extract validated device parameter
		device := validationResult.Data["device"].(string)

		// Stop streaming using controller (maps internally)
		err := s.mediaMTXController.StopStreaming(context.Background(), device)
		if err != nil {
			return nil, fmt.Errorf("failed to stop streaming: %v", err)
		}

		// Return stop result
		return map[string]interface{}{
			"device":           device,
			"stream_name":      fmt.Sprintf("%s_viewing", device),
			"status":           "STOPPED",
			"start_time":       time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"end_time":         time.Now().Format(time.RFC3339),
			"duration":         300,
			"stream_continues": false,
		}, nil
	})(params, client)
}

// MethodGetStreamURL gets the stream URL for a specific camera device
func (s *WebSocketServer) MethodGetStreamURL(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("get_stream_url", func() (interface{}, error) {
		// Validate device parameter using centralized validation
		validationResult := s.validationHelper.ValidateDeviceParameter(params)
		if !validationResult.Valid {
			// Log validation warnings for debugging
			s.validationHelper.LogValidationWarnings(validationResult, "get_stream_url", client.ClientID)
			return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
		}

		// Extract validated device parameter
		device := validationResult.Data["device"].(string)

		// ULTRA THIN: Delegate to Controller - returns complete API-ready response
		streamURLResp, err := s.mediaMTXController.GetStreamURL(context.Background(), device)
		if err != nil {
			return nil, fmt.Errorf("failed to get stream URL: %v", err)
		}

		// Return Controller's API-ready response directly - no business logic duplication
		return streamURLResp, nil
	})(params, client)
}

// MethodGetStreamStatus gets detailed status information for a specific camera stream
func (s *WebSocketServer) MethodGetStreamStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("get_stream_status", func() (interface{}, error) {
		// Validate device parameter using centralized validation
		validationResult := s.validationHelper.ValidateDeviceParameter(params)
		if !validationResult.Valid {
			// Log validation warnings for debugging
			s.validationHelper.LogValidationWarnings(validationResult, "get_stream_status", client.ClientID)
			return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
		}

		// Extract validated device parameter
		device := validationResult.Data["device"].(string)

		// Get stream status from controller using camera ID
		stream, err := s.mediaMTXController.GetStreamStatus(context.Background(), device)
		if err != nil {
			return nil, fmt.Errorf("stream not found or not active: %v", err)
		}

		// ULTRA THIN: Return Controller's API-ready response directly - no business logic duplication
		return stream, nil
	})(params, client)
}

// MethodDiscoverExternalStreams discovers external streams (Skydio UAVs, etc.)
func (s *WebSocketServer) MethodDiscoverExternalStreams(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("discover_external_streams", func() (interface{}, error) {
		// Get discovery options from params
		options := mediamtx.DiscoveryOptions{
			SkydioEnabled:  true,  // Default to Skydio discovery
			GenericEnabled: false, // Default to disabled
		}

		if skydioEnabled, ok := params["skydio_enabled"].(bool); ok {
			options.SkydioEnabled = skydioEnabled
		}
		if genericEnabled, ok := params["generic_enabled"].(bool); ok {
			options.GenericEnabled = genericEnabled
		}
		if forceRescan, ok := params["force_rescan"].(bool); ok {
			options.ForceRescan = forceRescan
		}
		if includeOffline, ok := params["include_offline"].(bool); ok {
			options.IncludeOffline = includeOffline
		}

		// Trigger discovery with options
		result, err := s.mediaMTXController.DiscoverExternalStreams(context.Background(), options)
		if err != nil {
			return nil, err
		}

		return result, nil
	})(params, client)
}

// MethodAddExternalStream adds an external stream to the system
func (s *WebSocketServer) MethodAddExternalStream(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("add_external_stream", func() (interface{}, error) {
		// Extract stream parameters
		streamURL, ok := params["stream_url"].(string)
		if !ok || streamURL == "" {
			return nil, fmt.Errorf("stream_url parameter is required")
		}

		streamName, ok := params["stream_name"].(string)
		if !ok || streamName == "" {
			return nil, fmt.Errorf("stream_name parameter is required")
		}

		streamType, ok := params["stream_type"].(string)
		if !ok {
			streamType = "generic_rtsp" // Default type
		}

		// Create external stream
		stream := &mediamtx.ExternalStream{
			URL:          streamURL,
			Name:         streamName,
			Type:         streamType,
			Status:       "discovered",
			DiscoveredAt: time.Now(),
			LastSeen:     time.Now(),
			Capabilities: map[string]interface{}{
				"protocol": "rtsp",
				"source":   "external",
			},
		}

		// Pure delegation to Controller - returns API-ready AddExternalStreamResponse
		return s.mediaMTXController.AddExternalStream(context.Background(), stream)
	})(params, client)
}

// MethodRemoveExternalStream removes an external stream from the system
func (s *WebSocketServer) MethodRemoveExternalStream(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("remove_external_stream", func() (interface{}, error) {
		// Extract stream URL
		streamURL, ok := params["stream_url"].(string)
		if !ok || streamURL == "" {
			return nil, fmt.Errorf("stream_url parameter is required")
		}

		// Pure delegation to Controller - returns API-ready RemoveExternalStreamResponse
		return s.mediaMTXController.RemoveExternalStream(context.Background(), streamURL)
	})(params, client)
}

// MethodGetExternalStreams returns all discovered external streams
func (s *WebSocketServer) MethodGetExternalStreams(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("get_external_streams", func() (interface{}, error) {
		// Pure delegation to Controller - returns API-ready GetExternalStreamsResponse
		return s.mediaMTXController.GetExternalStreams(context.Background())
	})(params, client)
}

// MethodSetDiscoveryInterval sets the discovery scan interval
func (s *WebSocketServer) MethodSetDiscoveryInterval(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	return s.authenticatedMethodWrapper("set_discovery_interval", func() (interface{}, error) {
		// Extract interval parameter
		interval, ok := params["scan_interval"].(float64)
		if !ok {
			return nil, fmt.Errorf("scan_interval parameter is required and must be a number")
		}

		// Validate interval (0 = on-demand only, >0 = periodic scanning)
		if interval < 0 {
			return nil, fmt.Errorf("scan_interval must be >= 0")
		}

		// Pure delegation to Controller - returns API-ready SetDiscoveryIntervalResponse
		return s.mediaMTXController.SetDiscoveryInterval(int(interval))
	})(params, client)
}

// translateErrorToJsonRpc converts business logic errors to appropriate JSON-RPC errors
func (s *WebSocketServer) translateErrorToJsonRpc(err error, methodName string) *JsonRpcError {
	errMsg := err.Error()

	// External discovery disabled error - check for both variations
	if (strings.Contains(strings.ToLower(errMsg), "external stream discovery") ||
		strings.Contains(strings.ToLower(errMsg), "external discovery")) &&
		(strings.Contains(errMsg, "disabled") || strings.Contains(errMsg, "not configured")) {
		return NewJsonRpcError(UNSUPPORTED, "feature_disabled",
			"External stream discovery is disabled in configuration", "Enable external discovery in configuration")
	}

	// Check for specific error patterns from Phase 1 enhanced error messages
	if strings.Contains(errMsg, "status 404") {
		// MediaMTX path not found
		return NewJsonRpcError(CAMERA_NOT_FOUND, "RECORDINGS_NOT_FOUND",
			"No recordings found", "Check if recordings exist")
	}
	if strings.Contains(errMsg, "status 503") || strings.Contains(errMsg, "failed to connect") {
		// MediaMTX service unavailable
		return NewJsonRpcError(MEDIAMTX_UNAVAILABLE, "SERVICE_UNAVAILABLE",
			"MediaMTX service is not responding", "Try again later")
	}
	if strings.Contains(errMsg, "status 403") {
		// Permission denied
		return NewJsonRpcError(INSUFFICIENT_PERMISSIONS, "ACCESS_DENIED",
			"Access to recordings denied", "Check permissions")
	}

	// Camera-specific errors
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "camera device not found") {
		return NewJsonRpcError(CAMERA_NOT_FOUND, "camera_not_found",
			"Camera not found or disconnected", "Check camera identifier")
	}

	// File operation errors
	if strings.Contains(errMsg, "file not found") {
		return NewJsonRpcError(FILE_NOT_FOUND, "file_not_found",
			"File not found or inaccessible", "Verify filename and path")
	}

	// Generic error fallback
	return NewJsonRpcError(INTERNAL_ERROR, "METHOD_ERROR",
		fmt.Sprintf("%s: %v", methodName, err), "Contact support if issue persists")
}
