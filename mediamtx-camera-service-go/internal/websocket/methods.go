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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// registerBuiltinMethods registers all built-in JSON-RPC methods
// Following Python _register_builtin_methods patterns
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

	// Notification methods
	s.registerMethod("camera_status_update", s.MethodCameraStatusUpdate, "1.0")
	s.registerMethod("recording_status_update", s.MethodRecordingStatusUpdate, "1.0")

	s.logger.WithField("action", "register_methods").Info("Built-in methods registered")
}

// registerMethod registers a JSON-RPC method handler
func (s *WebSocketServer) registerMethod(name string, handler MethodHandler, version string) {
	s.methodsMutex.Lock()
	defer s.methodsMutex.Unlock()

	// Wrap the handler to ensure security and metrics are always applied
	wrappedHandler := func(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
		startTime := time.Now()

		// Apply security checks
		if err := s.checkRateLimit(client); err != nil {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    RATE_LIMIT_EXCEEDED,
					Message: ErrorMessages[RATE_LIMIT_EXCEEDED],
					Data:    err.Error(),
				},
			}, nil
		}

		if err := s.checkMethodPermissions(client, name); err != nil {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INSUFFICIENT_PERMISSIONS,
					Message: ErrorMessages[INSUFFICIENT_PERMISSIONS],
					Data:    err.Error(),
				},
			}, nil
		}

		// Call the original handler
		response, err := handler(params, client)

		// Record metrics
		duration := time.Since(startTime).Seconds()
		s.recordRequest(name, duration)

		// Handle errors
		if err != nil {
			s.metricsMutex.Lock()
			s.metrics.ErrorCount++
			s.metricsMutex.Unlock()
		}

		return response, err
	}

	s.methods[name] = wrappedHandler
	s.methodVersions[name] = version

	s.logger.WithFields(logrus.Fields{
		"method":  name,
		"version": version,
		"action":  "register_method",
	}).Debug("Method registered with security and metrics wrapper")
}

// MethodPing implements the ping method
// Following Python _method_ping implementation
func (s *WebSocketServer) MethodPing(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "ping",
		"action":    "method_call",
	}).Debug("Ping method called")

	// Record performance metrics
	duration := time.Since(startTime).Seconds()
	s.recordRequest("ping", duration)

	// Return "pong" as specified in API documentation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  "pong",
	}, nil
}

// MethodAuthenticate implements the authenticate method
// Following Python _method_authenticate implementation
func (s *WebSocketServer) MethodAuthenticate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	startTime := time.Now()

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

	// Record performance metrics
	duration := time.Since(startTime).Seconds()
	s.recordRequest("authenticate", duration)

	// Return authentication result following Python implementation
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"authenticated": true,
			"role":          claims.Role,
			"permissions":   GetPermissionsForRole(claims.Role),
			"expires_at":    expiresAt.Format(time.RFC3339),
			"session_id":    client.ClientID,
		},
	}, nil
}

// MethodGetCameraList implements the get_camera_list method
// Following Python _method_get_camera_list implementation
func (s *WebSocketServer) MethodGetCameraList(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
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

	// Convert camera list to response format following API documentation exactly
	cameraList := make([]map[string]interface{}, 0, len(cameras))
	connectedCount := 0

	for devicePath, camera := range cameras {
		// Convert device path to camera identifier for API response
		cameraID := s.getCameraIdentifierFromDevicePath(devicePath)

		// Get resolution from first format if available
		resolution := ""
		if len(camera.Formats) > 0 {
			format := camera.Formats[0]
			resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
		}

		// Build streams object following API documentation exactly
		streams := map[string]string{
			"rtsp":   fmt.Sprintf("rtsp://localhost:8554/%s", s.getStreamNameFromDevicePath(devicePath)),
			"webrtc": fmt.Sprintf("http://localhost:8889/%s/webrtc", s.getStreamNameFromDevicePath(devicePath)),
			"hls":    fmt.Sprintf("http://localhost:8888/%s", s.getStreamNameFromDevicePath(devicePath)),
		}

		cameraData := map[string]interface{}{
			"device":     cameraID, // Use camera identifier instead of device path
			"status":     string(camera.Status),
			"name":       camera.Name,
			"resolution": resolution,
			"fps":        30, // Default FPS - can be enhanced later
			"streams":    streams,
		}

		cameraList = append(cameraList, cameraData)

		if camera.Status == "CONNECTED" {
			connectedCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":     client.ClientID,
		"method":        "get_camera_list",
		"total_cameras": len(cameras),
		"connected":     connectedCount,
		"action":        "camera_list_success",
	}).Debug("Camera list retrieved successfully")

	// Return camera list following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"cameras":   cameraList,
			"total":     len(cameras),
			"connected": connectedCount,
		},
	}, nil
}

// MethodGetCameraStatus implements the get_camera_status method
// Following Python _method_get_camera_status implementation
func (s *WebSocketServer) MethodGetCameraStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
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
	cameraID, ok := params["device"].(string)
	if !ok || cameraID == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	// Validate camera identifier format
	if !s.validateCameraIdentifier(cameraID) {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    fmt.Sprintf("Invalid camera identifier format: %s. Expected format: camera[0-9]+", cameraID),
			},
		}, nil
	}

	// Convert camera identifier to device path for internal lookup
	devicePath := s.getDevicePathFromCameraIdentifier(cameraID)

	// Get camera status from camera monitor
	camera, exists := s.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    CAMERA_NOT_FOUND,
				Message: ErrorMessages[CAMERA_NOT_FOUND],
				Data:    fmt.Sprintf("Camera '%s' not found", cameraID),
			},
		}, nil
	}

	// Get resolution from first format if available
	resolution := ""
	if len(camera.Formats) > 0 {
		format := camera.Formats[0]
		resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
	}

	// Build streams object following API documentation exactly
	streams := map[string]string{
		"rtsp":   fmt.Sprintf("rtsp://localhost:8554/%s", s.getStreamNameFromDevicePath(devicePath)),
		"webrtc": fmt.Sprintf("webrtc://localhost:8002/%s", s.getStreamNameFromDevicePath(devicePath)),
		"hls":    fmt.Sprintf("http://localhost:8002/hls/%s.m3u8", s.getStreamNameFromDevicePath(devicePath)),
	}

	// Build capabilities object following API documentation exactly
	capabilities := map[string]interface{}{
		"formats":     []string{}, // Will be populated from camera.Formats
		"resolutions": []string{}, // Will be populated from camera.Formats
	}

	// Populate capabilities from camera data if available
	if len(camera.Formats) > 0 {
		formats := make([]string, 0, len(camera.Formats))
		resolutions := make([]string, 0, len(camera.Formats))

		for _, format := range camera.Formats {
			formats = append(formats, format.PixelFormat)
			resolution := fmt.Sprintf("%dx%d", format.Width, format.Height)
			resolutions = append(resolutions, resolution)
		}

		capabilities["formats"] = formats
		capabilities["resolutions"] = resolutions
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"device":    cameraID,
		"method":    "get_camera_status",
		"status":    string(camera.Status),
		"action":    "camera_status_success",
	}).Debug("Camera status retrieved successfully")

	// Return camera status following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"device":       cameraID,
			"status":       string(camera.Status),
			"name":         camera.Name,
			"resolution":   resolution,
			"fps":          30, // Default FPS - can be enhanced later
			"streams":      streams,
			"capabilities": capabilities,
		},
	}, nil
}

// MethodGetMetrics implements the get_metrics method
// Following Python _method_get_metrics implementation
func (s *WebSocketServer) MethodGetMetrics(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_metrics",
		"action":    "method_call",
	}).Debug("Get metrics method called")

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

	// Get system metrics from MediaMTX controller
	var systemMetrics *mediamtx.SystemMetrics
	var err error

	if s.mediaMTXController != nil {
		systemMetrics, err = s.mediaMTXController.GetSystemMetrics(context.Background())
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"client_id": client.ClientID,
				"method":    "get_metrics",
				"action":    "controller_error",
			}).Error("Error getting system metrics from controller")
		}
	}

	// Enhanced health metrics are available through systemMetrics (Phase 1 enhancement)

	// Get base performance metrics from existing infrastructure
	baseMetrics := s.GetMetrics()

	// Get active connections count
	s.clientsMutex.RLock()
	activeConnections := len(s.clients)
	s.clientsMutex.RUnlock()

	// Calculate average response time from existing metrics
	var averageResponseTime float64
	var totalResponseTime float64
	var responseCount int

	for _, times := range baseMetrics.ResponseTimes {
		for _, time := range times {
			totalResponseTime += time
			responseCount++
		}
	}

	if responseCount > 0 {
		averageResponseTime = totalResponseTime / float64(responseCount)
	}

	// Calculate error rate
	var errorRate float64
	if baseMetrics.RequestCount > 0 {
		errorRate = float64(baseMetrics.ErrorCount) / float64(baseMetrics.RequestCount) * 100.0
	}

	// Get system resource usage using Go runtime
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := float64(m.Alloc) / 1024 / 1024 // MB

	// Get CPU usage using Linux /proc/stat
	cpuUsage := getCPUUsage()

	// Get goroutines count
	goroutines := runtime.NumGoroutine()

	// Get heap allocation in bytes
	heapAlloc := m.HeapAlloc

	// Enhanced metrics result with sophisticated health monitoring (Phase 1 enhancement)
	result := map[string]interface{}{
		"active_connections":    activeConnections,
		"total_requests":        baseMetrics.RequestCount,
		"average_response_time": averageResponseTime,
		"error_rate":            errorRate,
		"memory_usage":          memoryUsage,
		"cpu_usage":             cpuUsage,
		"goroutines":            goroutines,
		"heap_alloc":            heapAlloc,
	}

	// Use system metrics from controller if available (Phase 1 enhancement)
	if systemMetrics != nil {
		// Use system metrics for response time and error rate, but keep WebSocket connection count
		averageResponseTime = systemMetrics.ResponseTime
		if systemMetrics.RequestCount > 0 {
			errorRate = float64(systemMetrics.ErrorCount) / float64(systemMetrics.RequestCount) * 100.0
		}

		// Add enhanced health monitoring metrics (Phase 1 enhancement)
		result["circuit_breaker_state"] = systemMetrics.CircuitBreakerState
		result["component_status"] = systemMetrics.ComponentStatus
		result["error_counts"] = systemMetrics.ErrorCounts
		result["last_check"] = systemMetrics.LastCheck

		// Update metrics with enhanced values (but preserve WebSocket connection count)
		result["average_response_time"] = averageResponseTime
		result["error_rate"] = errorRate
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":             client.ClientID,
		"method":                "get_metrics",
		"active_connections":    activeConnections,
		"total_requests":        baseMetrics.RequestCount,
		"average_response_time": averageResponseTime,
		"error_rate":            errorRate,
		"memory_usage":          memoryUsage,
		"action":                "metrics_success",
	}).Debug("Metrics retrieved successfully")

	// Return enhanced metrics following API documentation exactly (Phase 1 enhancement)
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  result,
	}, nil
}

// MethodGetCameraCapabilities implements the get_camera_capabilities method
// Following Python _method_get_camera_capabilities implementation
func (s *WebSocketServer) MethodGetCameraCapabilities(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_camera_capabilities",
		"action":    "method_call",
	}).Debug("Get camera capabilities method called")

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

	// Initialize response with architecture defaults following API documentation exactly
	cameraCapabilities := map[string]interface{}{
		"device":            device,
		"formats":           []string{},
		"resolutions":       []string{},
		"fps_options":       []int{},
		"validation_status": "none",
	}

	// Get camera info from camera monitor using existing infrastructure
	camera, exists := s.cameraMonitor.GetDevice(device)
	if !exists {
		cameraCapabilities["validation_status"] = "disconnected"
	} else if camera.Status != "CONNECTED" {
		cameraCapabilities["validation_status"] = "disconnected"
	} else {
		// Camera is connected, get real capability metadata
		// Convert camera formats to string list per API documentation
		formats := make([]string, 0, len(camera.Formats))
		for _, format := range camera.Formats {
			formats = append(formats, format.PixelFormat)
		}
		cameraCapabilities["formats"] = formats

		// Convert resolutions to string list per API documentation
		resolutions := make([]string, 0, len(camera.Formats))
		for _, format := range camera.Formats {
			resolution := fmt.Sprintf("%dx%d", format.Width, format.Height)
			resolutions = append(resolutions, resolution)
		}
		cameraCapabilities["resolutions"] = resolutions

		// Add FPS options as int list per API documentation
		fpsOptions := []int{15, 30, 60}
		cameraCapabilities["fps_options"] = fpsOptions

		// Set validation status to confirmed since we have real data
		cameraCapabilities["validation_status"] = "confirmed"

		s.logger.WithFields(logrus.Fields{
			"client_id":   client.ClientID,
			"device":      device,
			"method":      "get_camera_capabilities",
			"formats":     len(formats),
			"resolutions": len(resolutions),
			"action":      "capabilities_success",
		}).Debug("Camera capabilities retrieved successfully")
	}

	// Return camera capabilities following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  cameraCapabilities,
	}, nil
}

// MethodGetStatus implements the get_status method
// Following Python _method_get_status implementation
func (s *WebSocketServer) MethodGetStatus(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_status",
		"action":    "method_call",
	}).Debug("Get status method called")

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

	// Calculate uptime as positive integer (seconds since start)
	startTime := s.metrics.StartTime
	uptime := int(time.Since(startTime).Seconds())
	if uptime < 0 {
		uptime = 0
	}

	// Determine overall system status
	systemStatus := "healthy"

	// Check component statuses
	websocketServerStatus := "running"
	cameraMonitorStatus := "running"
	mediamtxControllerStatus := "unknown"

	// Check if camera monitor is available and running
	if s.cameraMonitor != nil {
		cameraMonitorStatus = "running"
	} else {
		cameraMonitorStatus = "error"
		systemStatus = "degraded"
	}

	// Check if server is running
	if !s.running {
		websocketServerStatus = "error"
		systemStatus = "degraded"
	}

	// Note: MediaMTX controller status would be checked here when MediaMTX integration is available
	// For now, we'll use "unknown" as per Python implementation pattern

	s.logger.WithFields(logrus.Fields{
		"client_id":     client.ClientID,
		"method":        "get_status",
		"system_status": systemStatus,
		"uptime":        uptime,
		"action":        "status_success",
	}).Debug("System status retrieved successfully")

	// Return status following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"status":  systemStatus,
			"uptime":  uptime,
			"version": "1.0.0",
			"components": map[string]interface{}{
				"websocket_server":    websocketServerStatus,
				"camera_monitor":      cameraMonitorStatus,
				"mediamtx_controller": mediamtxControllerStatus,
			},
		},
	}, nil
}

// MethodGetServerInfo implements the get_server_info method
// Following Python _method_get_server_info implementation
func (s *WebSocketServer) MethodGetServerInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_server_info",
		"action":    "method_call",
	}).Debug("Get server info method called")

	// Record performance metrics first
	duration := time.Since(startTime).Seconds()
	s.recordRequest("get_server_info", duration)

	// Check authentication
	if !client.Authenticated {
		// Increment error count for authentication failure
		s.metricsMutex.Lock()
		s.metrics.ErrorCount++
		s.metricsMutex.Unlock()

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    AUTHENTICATION_REQUIRED,
				Message: ErrorMessages[AUTHENTICATION_REQUIRED],
			},
		}, nil
	}

	// Check permissions
	if err := s.checkMethodPermissions(client, "get_server_info"); err != nil {
		// Increment error count for permission failure
		s.metricsMutex.Lock()
		s.metrics.ErrorCount++
		s.metricsMutex.Unlock()

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INSUFFICIENT_PERMISSIONS,
				Message: ErrorMessages[INSUFFICIENT_PERMISSIONS],
				Data:    err.Error(),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_server_info",
		"action":    "server_info_success",
	}).Debug("Server info retrieved successfully")

	// Return server info following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"name":              "MediaMTX Camera Service",
			"version":           "1.0.0",
			"build_date":        time.Now().Format("2006-01-02"),
			"go_version":        runtime.Version(),
			"architecture":      runtime.GOARCH,
			"capabilities":      []string{"snapshots", "recordings", "streaming"},
			"supported_formats": []string{"mp4", "mkv", "jpg"},
			"max_cameras":       10,
		},
	}, nil
}

// MethodGetStreams implements the get_streams method
// Following Python _method_get_streams implementation
func (s *WebSocketServer) MethodGetStreams(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_streams",
		"action":    "method_call",
	}).Debug("Get streams method called")

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

	// Get streams from MediaMTX controller
	streams, err := s.mediaMTXController.GetStreams(context.Background())
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "get_streams",
			"action":    "get_streams_error",
		}).Error("Failed to get streams from MediaMTX controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    MEDIAMTX_UNAVAILABLE,
				Message: "Failed to get streams from MediaMTX service",
				Data: map[string]interface{}{
					"reason": err.Error(),
				},
			},
		}, nil
	}

	// Convert streams to response format
	streamList := make([]map[string]interface{}, 0, len(streams))
	for _, stream := range streams {
		// Convert MediaMTX API structure to our response format
		sourceStr := ""
		if stream.Source != nil {
			sourceStr = stream.Source.Type
		}

		status := "NOT_READY"
		if stream.Ready {
			status = "READY"
		}

		streamList = append(streamList, map[string]interface{}{
			"id":     stream.Name, // Use Name as ID for backward compatibility
			"name":   stream.Name,
			"source": sourceStr,
			"status": status,
		})
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":    client.ClientID,
		"method":       "get_streams",
		"stream_count": len(streamList),
		"action":       "get_streams_success",
	}).Debug("Successfully retrieved streams from MediaMTX controller")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  streamList,
	}, nil
}

// MethodListRecordings implements the list_recordings method
// Following Python _method_list_recordings implementation
func (s *WebSocketServer) MethodListRecordings(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "list_recordings",
		"action":    "method_call",
	}).Debug("List recordings method called")

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

	// Parse parameters with defaults
	limit := 100
	offset := 0

	if params != nil {
		if limitVal, ok := params["limit"]; ok {
			if limitInt, ok := limitVal.(int); ok && limitInt >= 1 && limitInt <= 1000 {
				limit = limitInt
			} else {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					Error: &JsonRpcError{
						Code:    INVALID_PARAMS,
						Message: ErrorMessages[INVALID_PARAMS],
						Data:    "Invalid limit parameter: must be integer between 1 and 1000",
					},
				}, nil
			}
		}

		if offsetVal, ok := params["offset"]; ok {
			if offsetInt, ok := offsetVal.(int); ok && offsetInt >= 0 {
				offset = offsetInt
			} else {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					Error: &JsonRpcError{
						Code:    INVALID_PARAMS,
						Message: ErrorMessages[INVALID_PARAMS],
						Data:    "Invalid offset parameter: must be non-negative integer",
					},
				}, nil
			}
		}
	}

	// Use MediaMTX controller to get recordings list
	fileList, err := s.mediaMTXController.ListRecordings(context.Background(), limit, offset)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_recordings",
			"action":    "controller_error",
		}).Error("Error getting recordings list from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error getting recordings list: %v", err),
			},
		}, nil
	}

	// Check if no recordings found
	if fileList.Total == 0 {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "No recordings found",
				Data:    "No recording files found in storage",
			},
		}, nil
	}

	// Convert FileMetadata to map for JSON response
	files := make([]map[string]interface{}, len(fileList.Files))
	for i, file := range fileList.Files {
		fileData := map[string]interface{}{
			"filename":      file.FileName,
			"file_size":     file.FileSize,
			"modified_time": file.ModifiedAt.Format(time.RFC3339),
			"download_url":  file.DownloadURL,
		}

		// Add duration if available
		if file.Duration != nil {
			fileData["duration"] = *file.Duration
		}

		files[i] = fileData
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "list_recordings",
		"total_files": fileList.Total,
		"returned":    len(files),
		"action":      "recordings_listed",
	}).Debug("Recordings listed successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"files":  files,
			"total":  fileList.Total,
			"limit":  fileList.Limit,
			"offset": fileList.Offset,
		},
	}, nil
}

// MethodDeleteRecording implements the delete_recording method
// Following Python _method_delete_recording implementation
func (s *WebSocketServer) MethodDeleteRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "delete_recording",
		"action":    "method_call",
	}).Debug("Delete recording method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename parameter is required",
			},
		}, nil
	}

	filename, ok := params["filename"].(string)
	if !ok || filename == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename must be a non-empty string",
			},
		}, nil
	}

	// Use MediaMTX controller to delete recording
	err := s.mediaMTXController.DeleteRecording(context.Background(), filename)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "delete_recording",
			"filename":  filename,
			"action":    "controller_error",
		}).Error("Error deleting recording from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error deleting recording: %v", err),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "delete_recording",
		"filename":  filename,
		"action":    "delete_success",
	}).Info("Recording file deleted successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"filename": filename,
			"deleted":  true,
			"message":  "Recording file deleted successfully",
		},
	}, nil
}

// MethodDeleteSnapshot implements the delete_snapshot method
// Following Python _method_delete_snapshot implementation
func (s *WebSocketServer) MethodDeleteSnapshot(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "delete_snapshot",
		"action":    "method_call",
	}).Debug("Delete snapshot method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename parameter is required",
			},
		}, nil
	}

	filename, ok := params["filename"].(string)
	if !ok || filename == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename must be a non-empty string",
			},
		}, nil
	}

	// Use MediaMTX controller to delete snapshot
	err := s.mediaMTXController.DeleteSnapshot(context.Background(), filename)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "delete_snapshot",
			"filename":  filename,
			"action":    "controller_error",
		}).Error("Error deleting snapshot from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error deleting snapshot: %v", err),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "delete_snapshot",
		"filename":  filename,
		"action":    "delete_success",
	}).Info("Snapshot file deleted successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"filename": filename,
			"deleted":  true,
			"message":  "Snapshot file deleted successfully",
		},
	}, nil
}

// MethodGetStorageInfo implements the get_storage_info method
// Following Python _method_get_storage_info implementation
func (s *WebSocketServer) MethodGetStorageInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_storage_info",
		"action":    "method_call",
	}).Debug("Get storage info method called")

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

	// Get configuration for directory paths
	config := s.configManager.GetConfig()
	if config == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    "Configuration not available",
			},
		}, nil
	}

	recordingsDir := config.MediaMTX.RecordingsPath
	snapshotsDir := config.MediaMTX.SnapshotsPath

	// Get storage space information using Go's syscall package
	var stat unix.Statfs_t
	err := unix.Statfs(recordingsDir, &stat)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "get_storage_info",
			"directory": recordingsDir,
			"action":    "statfs_error",
		}).Error("Error getting storage information")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error getting storage information: %v", err),
			},
		}, nil
	}

	// Calculate storage space information
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	freeBytes := stat.Bfree * uint64(stat.Bsize)
	usedBytes := totalBytes - freeBytes
	usedPercent := float64(usedBytes) / float64(totalBytes) * 100.0

	// Calculate directory sizes
	recordingsSize := int64(0)
	if _, err := os.Stat(recordingsDir); err == nil {
		recordingsSize = s.calculateDirectorySize(recordingsDir)
	}

	snapshotsSize := int64(0)
	if _, err := os.Stat(snapshotsDir); err == nil {
		snapshotsSize = s.calculateDirectorySize(snapshotsDir)
	}

	// Determine warning levels (following API documentation)
	lowSpaceWarning := usedPercent >= 80.0

	s.logger.WithFields(logrus.Fields{
		"client_id":    client.ClientID,
		"method":       "get_storage_info",
		"total_gb":     totalBytes / 1024 / 1024 / 1024,
		"used_gb":      usedBytes / 1024 / 1024 / 1024,
		"used_percent": usedPercent,
		"action":       "storage_info_success",
	}).Debug("Storage information retrieved successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"total_space":       totalBytes,
			"used_space":        usedBytes,
			"available_space":   freeBytes,
			"usage_percentage":  usedPercent,
			"recordings_size":   recordingsSize,
			"snapshots_size":    snapshotsSize,
			"low_space_warning": lowSpaceWarning,
		},
	}, nil
}

// calculateDirectorySize calculates the total size of a directory recursively
func (s *WebSocketServer) calculateDirectorySize(dirPath string) int64 {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"directory": dirPath,
			"action":    "calculate_size_error",
		}).Warn("Error calculating directory size")
	}

	return totalSize
}

// MethodCleanupOldFiles implements the cleanup_old_files method
// Following Python _method_cleanup_old_files implementation
func (s *WebSocketServer) MethodCleanupOldFiles(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "cleanup_old_files",
		"action":    "method_call",
	}).Debug("Cleanup old files method called")

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

	// Get current configuration
	cfg := s.configManager.GetConfig()
	if cfg == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    "Configuration not available",
			},
		}, nil
	}

	// Check if retention policy is enabled
	if !cfg.RetentionPolicy.Enabled {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "Retention policy is not enabled",
			},
		}, nil
	}

	// Perform cleanup based on retention policy
	var deletedCount int
	var totalSize int64
	var err error

	if cfg.RetentionPolicy.Type == "age" {
		// Age-based cleanup
		deletedCount, totalSize, err = s.performAgeBasedCleanup(cfg.RetentionPolicy.MaxAgeDays)
	} else if cfg.RetentionPolicy.Type == "size" {
		// Size-based cleanup
		deletedCount, totalSize, err = s.performSizeBasedCleanup(cfg.RetentionPolicy.MaxSizeGB)
	} else {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "Unsupported retention policy type for cleanup",
			},
		}, nil
	}

	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "cleanup_old_files",
			"action":    "cleanup_error",
		}).Error("Error during file cleanup")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Cleanup failed: %v", err),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":     client.ClientID,
		"method":        "cleanup_old_files",
		"deleted_count": deletedCount,
		"total_size":    totalSize,
		"action":        "cleanup_completed",
	}).Info("File cleanup completed successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"cleanup_executed": true,
			"files_deleted":    deletedCount,
			"space_freed":      totalSize,
			"message":          "File cleanup completed successfully",
		},
	}, nil
}

// MethodSetRetentionPolicy implements the set_retention_policy method
// Following Python _method_set_retention_policy implementation
func (s *WebSocketServer) MethodSetRetentionPolicy(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "set_retention_policy",
		"action":    "method_call",
	}).Debug("Set retention policy method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "Parameters are required",
			},
		}, nil
	}

	// Extract and validate policy_type
	policyType, ok := params["policy_type"].(string)
	if !ok || policyType == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "policy_type must be one of: age, size, manual",
			},
		}, nil
	}

	// Validate policy_type values
	validPolicyTypes := map[string]bool{"age": true, "size": true, "manual": true}
	if !validPolicyTypes[policyType] {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "policy_type must be one of: age, size, manual",
			},
		}, nil
	}

	// Extract and validate enabled
	enabled, ok := params["enabled"].(bool)
	if !ok {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "enabled must be a boolean value",
			},
		}, nil
	}

	// Validate age-based policy parameters
	if policyType == "age" {
		maxAgeDays, exists := params["max_age_days"]
		if !exists {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_age_days is required for age-based policy",
				},
			}, nil
		}

		// Convert to float64 for validation (handles both int and float)
		var maxAgeFloat float64
		switch v := maxAgeDays.(type) {
		case int:
			maxAgeFloat = float64(v)
		case float64:
			maxAgeFloat = v
		default:
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_age_days must be a positive number for age-based policy",
				},
			}, nil
		}

		if maxAgeFloat <= 0 {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_age_days must be a positive number for age-based policy",
				},
			}, nil
		}
	}

	// Validate size-based policy parameters
	if policyType == "size" {
		maxSizeGB, exists := params["max_size_gb"]
		if !exists {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_size_gb is required for size-based policy",
				},
			}, nil
		}

		// Convert to float64 for validation (handles both int and float)
		var maxSizeFloat float64
		switch v := maxSizeGB.(type) {
		case int:
			maxSizeFloat = float64(v)
		case float64:
			maxSizeFloat = v
		default:
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_size_gb must be a positive number for size-based policy",
				},
			}, nil
		}

		if maxSizeFloat <= 0 {
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				Error: &JsonRpcError{
					Code:    INVALID_PARAMS,
					Message: ErrorMessages[INVALID_PARAMS],
					Data:    "max_size_gb must be a positive number for size-based policy",
				},
			}, nil
		}
	}

	// Get current configuration
	cfg := s.configManager.GetConfig()
	if cfg == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    "Configuration not available",
			},
		}, nil
	}

	// Update retention policy configuration
	cfg.RetentionPolicy.Enabled = enabled
	cfg.RetentionPolicy.Type = policyType

	// Update policy-specific parameters
	if policyType == "age" {
		if maxAgeDays, ok := params["max_age_days"].(float64); ok {
			cfg.RetentionPolicy.MaxAgeDays = int(maxAgeDays)
		} else if maxAgeDays, ok := params["max_age_days"].(int); ok {
			cfg.RetentionPolicy.MaxAgeDays = maxAgeDays
		}
	} else if policyType == "size" {
		if maxSizeGB, ok := params["max_size_gb"].(float64); ok {
			cfg.RetentionPolicy.MaxSizeGB = int(maxSizeGB)
		} else if maxSizeGB, ok := params["max_size_gb"].(int); ok {
			cfg.RetentionPolicy.MaxSizeGB = maxSizeGB
		}
	}

	// Note: Configuration changes are applied immediately in memory
	// For persistent changes, the configuration file should be updated

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "set_retention_policy",
		"policy_type": policyType,
		"enabled":     enabled,
		"action":      "policy_updated",
	}).Info("Retention policy configuration updated")

	// Build response result based on policy type
	result := map[string]interface{}{
		"policy_type": policyType,
		"enabled":     enabled,
		"message":     "Retention policy configuration updated successfully",
	}

	// Include policy-specific parameters in response
	if policyType == "age" {
		if maxAgeDays, ok := params["max_age_days"].(float64); ok {
			result["max_age_days"] = int(maxAgeDays)
		} else if maxAgeDays, ok := params["max_age_days"].(int); ok {
			result["max_age_days"] = maxAgeDays
		}
	} else if policyType == "size" {
		if maxSizeGB, ok := params["max_size_gb"].(float64); ok {
			result["max_size_gb"] = int(maxSizeGB)
		} else if maxSizeGB, ok := params["max_size_gb"].(int); ok {
			result["max_size_gb"] = maxSizeGB
		}
	}

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  result,
	}, nil
}

// MethodListSnapshots implements the list_snapshots method
// Following Python _method_list_snapshots implementation
func (s *WebSocketServer) MethodListSnapshots(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "list_snapshots",
		"action":    "method_call",
	}).Debug("List snapshots method called")

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

	// Parse parameters with defaults
	limit := 100
	offset := 0

	if params != nil {
		if limitVal, ok := params["limit"]; ok {
			if limitInt, ok := limitVal.(int); ok && limitInt >= 1 && limitInt <= 1000 {
				limit = limitInt
			} else {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					Error: &JsonRpcError{
						Code:    INVALID_PARAMS,
						Message: ErrorMessages[INVALID_PARAMS],
						Data:    "Invalid limit parameter: must be integer between 1 and 1000",
					},
				}, nil
			}
		}

		if offsetVal, ok := params["offset"]; ok {
			if offsetInt, ok := offsetVal.(int); ok && offsetInt >= 0 {
				offset = offsetInt
			} else {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					Error: &JsonRpcError{
						Code:    INVALID_PARAMS,
						Message: ErrorMessages[INVALID_PARAMS],
						Data:    "Invalid offset parameter: must be non-negative integer",
					},
				}, nil
			}
		}
	}

	// Use MediaMTX controller to get snapshots list
	fileList, err := s.mediaMTXController.ListSnapshots(context.Background(), limit, offset)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_snapshots",
			"action":    "controller_error",
		}).Error("Error getting snapshots list from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error getting snapshots list: %v", err),
			},
		}, nil
	}

	// Check if no snapshots found
	if fileList.Total == 0 {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "No snapshots found",
				Data:    "No snapshot files found in storage",
			},
		}, nil
	}

	// Convert FileMetadata to map for JSON response
	files := make([]map[string]interface{}, len(fileList.Files))
	for i, file := range fileList.Files {
		fileData := map[string]interface{}{
			"filename":      file.FileName,
			"file_size":     file.FileSize,
			"modified_time": file.ModifiedAt.Format(time.RFC3339),
			"download_url":  file.DownloadURL,
		}

		files[i] = fileData
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "list_snapshots",
		"total_files": fileList.Total,
		"returned":    len(files),
		"action":      "snapshots_listed",
	}).Debug("Snapshots listed successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"files":  files,
			"total":  fileList.Total,
			"limit":  fileList.Limit,
			"offset": fileList.Offset,
		},
	}, nil
}

// MethodTakeSnapshot implements the take_snapshot method
// Following Python _method_take_snapshot implementation
func (s *WebSocketServer) MethodTakeSnapshot(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "take_snapshot",
		"action":    "method_call",
	}).Debug("Take snapshot method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	devicePath, ok := params["device"].(string)
	if !ok || devicePath == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	// Extract optional parameters
	options := make(map[string]interface{})
	if filename, ok := params["filename"].(string); ok && filename != "" {
		options["filename"] = filename
	}
	if format, ok := params["format"].(string); ok && format != "" {
		options["format"] = format
	}
	if quality, ok := params["quality"].(int); ok && quality > 0 {
		options["quality"] = quality
	}

	// Validate camera device exists
	_, exists := s.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    CAMERA_NOT_FOUND,
				Message: ErrorMessages[CAMERA_NOT_FOUND],
				Data:    fmt.Sprintf("Camera device %s not found", devicePath),
			},
		}, nil
	}

	// Take snapshot using MediaMTX controller
	snapshot, err := s.mediaMTXController.TakeAdvancedSnapshot(context.Background(), devicePath, "", options)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "take_snapshot",
			"device":    devicePath,
			"action":    "take_snapshot_error",
		}).Error("Failed to take snapshot using MediaMTX controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    MEDIAMTX_UNAVAILABLE,
				Message: "Failed to take snapshot",
				Data: map[string]interface{}{
					"reason": err.Error(),
				},
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "take_snapshot",
		"device":      devicePath,
		"snapshot_id": snapshot.ID,
		"action":      "take_snapshot_success",
	}).Info("Successfully took snapshot using MediaMTX controller")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"snapshot_id": snapshot.ID,
			"device":      snapshot.Device,
			"file_path":   snapshot.FilePath,
			"size":        snapshot.Size,
			"created":     snapshot.Created,
		},
	}, nil
}

// MethodStartRecording implements the start_recording method
// Following Python _method_start_recording implementation
func (s *WebSocketServer) MethodStartRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "start_recording",
		"action":    "method_call",
	}).Debug("Start recording method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	devicePath, ok := params["device"].(string)
	if !ok || devicePath == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	// Extract optional parameters with enhanced use case support (Phase 2 enhancement)
	options := make(map[string]interface{})
	if duration, ok := params["duration_seconds"].(int); ok && duration > 0 {
		options["max_duration"] = time.Duration(duration) * time.Second
	}
	if format, ok := params["format"].(string); ok && format != "" {
		options["output_format"] = format
	}
	if codec, ok := params["codec"].(string); ok && codec != "" {
		options["codec"] = codec
	}
	if quality, ok := params["quality"].(int); ok && quality > 0 {
		options["crf"] = quality
	}

	// Enhanced use case management parameters (Phase 2 enhancement)
	if useCase, ok := params["use_case"].(string); ok && useCase != "" {
		options["use_case"] = useCase
	}
	if priority, ok := params["priority"].(int); ok && priority > 0 {
		options["priority"] = priority
	}
	if autoCleanup, ok := params["auto_cleanup"].(bool); ok {
		options["auto_cleanup"] = autoCleanup
	}
	if retentionDays, ok := params["retention_days"].(int); ok && retentionDays > 0 {
		options["retention_days"] = retentionDays
	}
	if qualityStr, ok := params["quality_level"].(string); ok && qualityStr != "" {
		options["quality"] = qualityStr
	}
	if autoRotate, ok := params["auto_rotate"].(bool); ok {
		options["auto_rotate"] = autoRotate
	}
	if rotationSize, ok := params["rotation_size"].(int64); ok && rotationSize > 0 {
		options["rotation_size"] = rotationSize
	}

	// Enhanced segment-based rotation parameters (Phase 3 enhancement)
	if continuityMode, ok := params["continuity_mode"].(bool); ok {
		options["continuity_mode"] = continuityMode
	}
	if segmentFormat, ok := params["segment_format"].(string); ok && segmentFormat != "" {
		options["segment_format"] = segmentFormat
	}
	if resetTimestamps, ok := params["reset_timestamps"].(bool); ok {
		options["reset_timestamps"] = resetTimestamps
	}
	if strftimeEnabled, ok := params["strftime_enabled"].(bool); ok {
		options["strftime_enabled"] = strftimeEnabled
	}
	if segmentPrefix, ok := params["segment_prefix"].(string); ok && segmentPrefix != "" {
		options["segment_prefix"] = segmentPrefix
	}
	if maxSegments, ok := params["max_segments"].(int); ok && maxSegments > 0 {
		options["max_segments"] = maxSegments
	}
	if segmentRotation, ok := params["segment_rotation"].(bool); ok {
		options["segment_rotation"] = segmentRotation
	}

	// Validate camera device exists
	_, exists := s.cameraMonitor.GetDevice(devicePath)
	if !exists {
		// Enhanced error categorization and logging (Phase 4 enhancement)
		enhancedErr := mediamtx.CategorizeError(fmt.Errorf("camera device not found: %s", devicePath))
		errorMetadata := mediamtx.GetErrorMetadata(enhancedErr)
		recoveryStrategies := mediamtx.GetRecoveryStrategies(enhancedErr.GetCategory())

		s.logger.WithFields(logrus.Fields{
			"client_id":           client.ClientID,
			"method":              "start_recording",
			"device":              devicePath,
			"error_category":      errorMetadata["category"],
			"error_severity":      errorMetadata["severity"],
			"retryable":           errorMetadata["retryable"],
			"recoverable":         errorMetadata["recoverable"],
			"recovery_strategies": recoveryStrategies,
		}).Warn("Camera device not found with enhanced error categorization")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    CAMERA_NOT_FOUND,
				Message: ErrorMessages[CAMERA_NOT_FOUND],
				Data: map[string]interface{}{
					"device":              devicePath,
					"error_category":      errorMetadata["category"],
					"error_severity":      errorMetadata["severity"],
					"recovery_strategies": recoveryStrategies,
				},
			},
		}, nil
	}

	// Start recording using MediaMTX controller
	session, err := s.mediaMTXController.StartAdvancedRecording(context.Background(), devicePath, "", options)
	if err != nil {
		// Enhanced error categorization and logging (Phase 4 enhancement)
		enhancedErr := mediamtx.CategorizeError(err)
		errorMetadata := mediamtx.GetErrorMetadata(enhancedErr)
		recoveryStrategies := mediamtx.GetRecoveryStrategies(enhancedErr.GetCategory())

		s.logger.WithFields(logrus.Fields{
			"client_id":           client.ClientID,
			"method":              "start_recording",
			"device":              devicePath,
			"action":              "start_recording_error",
			"error_category":      errorMetadata["category"],
			"error_severity":      errorMetadata["severity"],
			"retryable":           errorMetadata["retryable"],
			"recoverable":         errorMetadata["recoverable"],
			"recovery_strategies": recoveryStrategies,
		}).Error("Failed to start recording using MediaMTX controller with enhanced error categorization")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    MEDIAMTX_UNAVAILABLE,
				Message: "Failed to start recording",
				Data: map[string]interface{}{
					"reason": err.Error(),
				},
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":  client.ClientID,
		"method":     "start_recording",
		"device":     devicePath,
		"session_id": session.ID,
		"action":     "start_recording_success",
	}).Info("Successfully started recording using MediaMTX controller")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"session_id": session.ID,
			"device":     session.Device,
			"status":     session.Status,
			"start_time": session.StartTime,
			// Enhanced use case information (Phase 2 enhancement)
			"use_case":       session.UseCase,
			"priority":       session.Priority,
			"auto_cleanup":   session.AutoCleanup,
			"retention_days": session.RetentionDays,
			"quality":        session.Quality,
			"max_duration":   session.MaxDuration.String(),
			"auto_rotate":    session.AutoRotate,
			"rotation_size":  session.RotationSize,
			// Enhanced segment-based rotation information (Phase 3 enhancement)
			"continuity_id": session.ContinuityID,
		},
	}, nil
}

// MethodStopRecording implements the stop_recording method
// Following Python _method_stop_recording implementation
func (s *WebSocketServer) MethodStopRecording(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "stop_recording",
		"action":    "method_call",
	}).Debug("Stop recording method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	cameraID, ok := params["device"].(string)
	if !ok || cameraID == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required",
			},
		}, nil
	}

	// Validate camera identifier format
	if !s.validateCameraIdentifier(cameraID) {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    fmt.Sprintf("Invalid camera identifier format: %s. Expected format: camera[0-9]+", cameraID),
			},
		}, nil
	}

	// Convert camera identifier to device path for internal operations
	devicePath := s.getDevicePathFromCameraIdentifier(cameraID)

	// Validate camera device exists
	_, exists := s.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    CAMERA_NOT_FOUND,
				Message: ErrorMessages[CAMERA_NOT_FOUND],
				Data:    fmt.Sprintf("Camera device %s not found", devicePath),
			},
		}, nil
	}

	// Get session ID from device using optimized lookup
	sessionID, exists := s.mediaMTXController.GetSessionIDByDevice(devicePath)
	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "No active recording session found for device",
				Data: map[string]interface{}{
					"device": devicePath,
				},
			},
		}, nil
	}

	if sessionID == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "No active recording session found for device",
				Data: map[string]interface{}{
					"device": devicePath,
				},
			},
		}, nil
	}

	// Stop recording using MediaMTX controller
	err := s.mediaMTXController.StopAdvancedRecording(context.Background(), sessionID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id":  client.ClientID,
			"method":     "stop_recording",
			"device":     devicePath,
			"session_id": sessionID,
			"action":     "stop_recording_error",
		}).Error("Failed to stop recording using MediaMTX controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    MEDIAMTX_UNAVAILABLE,
				Message: "Failed to stop recording",
				Data: map[string]interface{}{
					"reason": err.Error(),
				},
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id":  client.ClientID,
		"method":     "stop_recording",
		"device":     cameraID,
		"session_id": sessionID,
		"action":     "stop_recording_success",
	}).Info("Successfully stopped recording using MediaMTX controller")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"session_id": sessionID,
			"device":     cameraID,
			"status":     "STOPPED",
		},
	}, nil
}

// getStreamNameFromDevicePath extracts stream name from device path
func (s *WebSocketServer) getStreamNameFromDevicePath(devicePath string) string {
	// Extract device name from path (e.g., "/dev/video0" -> "video0")
	parts := strings.Split(devicePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// MethodGetRecordingInfo implements the get_recording_info method
// Following API documentation exactly
func (s *WebSocketServer) MethodGetRecordingInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_recording_info",
		"action":    "method_call",
	}).Debug("Get recording info method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename parameter is required",
			},
		}, nil
	}

	filename, ok := params["filename"].(string)
	if !ok || filename == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename must be a non-empty string",
			},
		}, nil
	}

	// Use MediaMTX controller to get recording info
	fileMetadata, err := s.mediaMTXController.GetRecordingInfo(context.Background(), filename)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "get_recording_info",
			"filename":  filename,
			"action":    "controller_error",
		}).Error("Error getting recording info from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error getting recording info: %v", err),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_recording_info",
		"filename":  filename,
		"action":    "recording_info_success",
	}).Debug("Recording info retrieved successfully")

	// Return recording info following API documentation exactly
	result := map[string]interface{}{
		"filename":     fileMetadata.FileName,
		"file_size":    fileMetadata.FileSize,
		"created_time": fileMetadata.CreatedAt.Format(time.RFC3339),
		"download_url": fileMetadata.DownloadURL,
	}

	// Add duration if available (video metadata extraction is already implemented)
	if fileMetadata.Duration != nil {
		result["duration"] = *fileMetadata.Duration
	} else {
		// Duration is nil for non-video files or when extraction fails
		result["duration"] = nil
	}

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  result,
	}, nil
}

// MethodGetSnapshotInfo implements the get_snapshot_info method
// Following API documentation exactly
func (s *WebSocketServer) MethodGetSnapshotInfo(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_snapshot_info",
		"action":    "method_call",
	}).Debug("Get snapshot info method called")

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

	// Validate parameters
	if params == nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename parameter is required",
			},
		}, nil
	}

	filename, ok := params["filename"].(string)
	if !ok || filename == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "filename must be a non-empty string",
			},
		}, nil
	}

	// Use MediaMTX controller to get snapshot info
	fileMetadata, err := s.mediaMTXController.GetSnapshotInfo(context.Background(), filename)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "get_snapshot_info",
			"filename":  filename,
			"action":    "controller_error",
		}).Error("Error getting snapshot info from controller")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error getting snapshot info: %v", err),
			},
		}, nil
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_snapshot_info",
		"filename":  filename,
		"action":    "snapshot_info_success",
	}).Debug("Snapshot info retrieved successfully")

	// Return snapshot info following API documentation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"filename":     fileMetadata.FileName,
			"file_size":    fileMetadata.FileSize,
			"created_time": fileMetadata.CreatedAt.Format(time.RFC3339),
			"download_url": fileMetadata.DownloadURL,
		},
	}, nil
}

// GetPermissionsForRole returns permissions for a given role
// Following Python role-based access control patterns
func GetPermissionsForRole(role string) []string {
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

// getCPUUsage returns CPU usage percentage using Linux /proc/stat
func getCPUUsage() float64 {
	// Read /proc/stat to get CPU statistics
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0.0 // Return 0 if we can't read CPU stats
	}

	// Parse the first line (total CPU stats)
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0
	}

	// Parse CPU line: cpu  user nice system idle iowait irq softirq steal guest guest_nice
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0.0
	}

	// Convert fields to integers
	var stats []int64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseInt(fields[i], 10, 64)
		if err != nil {
			return 0.0
		}
		stats = append(stats, val)
	}

	// Calculate total and idle time
	total := int64(0)
	for _, stat := range stats {
		total += stat
	}

	idle := int64(0)
	if len(stats) >= 4 {
		idle = stats[3] // idle time is the 4th field
	}

	// Calculate CPU usage percentage
	if total == 0 {
		return 0.0
	}

	usage := float64(total-idle) / float64(total) * 100.0
	return usage
}

// performAgeBasedCleanup performs age-based file cleanup
func (s *WebSocketServer) performAgeBasedCleanup(maxAgeDays int) (int, int64, error) {
	// Get recordings and snapshots directories from config
	cfg := s.configManager.GetConfig()
	if cfg == nil {
		return 0, 0, fmt.Errorf("configuration not available")
	}

	recordingsPath := cfg.MediaMTX.RecordingsPath
	snapshotsPath := cfg.MediaMTX.SnapshotsPath

	// Calculate cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)

	var totalDeleted int
	var totalSize int64

	// Clean recordings
	deleted, size, err := s.cleanupDirectory(recordingsPath, cutoffTime)
	if err != nil {
		s.logger.WithError(err).WithField("path", recordingsPath).Error("Error cleaning recordings directory")
	} else {
		totalDeleted += deleted
		totalSize += size
	}

	// Clean snapshots
	deleted, size, err = s.cleanupDirectory(snapshotsPath, cutoffTime)
	if err != nil {
		s.logger.WithError(err).WithField("path", snapshotsPath).Error("Error cleaning snapshots directory")
	} else {
		totalDeleted += deleted
		totalSize += size
	}

	return totalDeleted, totalSize, nil
}

// performSizeBasedCleanup performs size-based file cleanup
func (s *WebSocketServer) performSizeBasedCleanup(maxSizeGB int) (int, int64, error) {
	// Get recordings and snapshots directories from config
	cfg := s.configManager.GetConfig()
	if cfg == nil {
		return 0, 0, fmt.Errorf("configuration not available")
	}

	recordingsPath := cfg.MediaMTX.RecordingsPath
	snapshotsPath := cfg.MediaMTX.SnapshotsPath

	// Calculate max size in bytes
	maxSizeBytes := int64(maxSizeGB) * 1024 * 1024 * 1024

	var totalDeleted int
	var totalSize int64

	// Clean recordings by size
	deleted, size, err := s.cleanupDirectoryBySize(recordingsPath, maxSizeBytes)
	if err != nil {
		s.logger.WithError(err).WithField("path", recordingsPath).Error("Error cleaning recordings directory by size")
	} else {
		totalDeleted += deleted
		totalSize += size
	}

	// Clean snapshots by size
	deleted, size, err = s.cleanupDirectoryBySize(snapshotsPath, maxSizeBytes)
	if err != nil {
		s.logger.WithError(err).WithField("path", snapshotsPath).Error("Error cleaning snapshots directory by size")
	} else {
		totalDeleted += deleted
		totalSize += size
	}

	return totalDeleted, totalSize, nil
}

// cleanupDirectory removes files older than cutoff time
func (s *WebSocketServer) cleanupDirectory(dirPath string, cutoffTime time.Time) (int, int64, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var deletedCount int
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Failed to get file info")
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				s.logger.WithError(err).WithField("file", filePath).Warn("Failed to delete old file")
				continue
			}

			deletedCount++
			totalSize += info.Size()

			s.logger.WithFields(logrus.Fields{
				"file":     filePath,
				"size":     info.Size(),
				"modified": info.ModTime(),
				"action":   "file_deleted",
			}).Debug("Deleted old file")
		}
	}

	return deletedCount, totalSize, nil
}

// cleanupDirectoryBySize removes oldest files until directory size is under limit
func (s *WebSocketServer) cleanupDirectoryBySize(dirPath string, maxSizeBytes int64) (int, int64, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Collect file information
	type fileInfo struct {
		path    string
		size    int64
		modTime time.Time
	}

	var files []fileInfo
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Failed to get file info")
			continue
		}

		files = append(files, fileInfo{
			path:    filePath,
			size:    info.Size(),
			modTime: info.ModTime(),
		})
		totalSize += info.Size()
	}

	// Sort files by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove oldest files until we're under the size limit
	var deletedCount int
	var deletedSize int64

	for _, file := range files {
		if totalSize <= maxSizeBytes {
			break
		}

		if err := os.Remove(file.path); err != nil {
			s.logger.WithError(err).WithField("file", file.path).Warn("Failed to delete file")
			continue
		}

		deletedCount++
		deletedSize += file.size
		totalSize -= file.size

		s.logger.WithFields(logrus.Fields{
			"file":     file.path,
			"size":     file.size,
			"modified": file.modTime,
			"action":   "file_deleted",
		}).Debug("Deleted file for size management")
	}

	return deletedCount, deletedSize, nil
}

// Abstraction layer mapping functions
// These functions handle the conversion between camera identifiers (camera0, camera1)
// and device paths (/dev/video0, /dev/video1) to maintain proper API abstraction

// getCameraIdentifierFromDevicePath converts a device path to a camera identifier
// Example: /dev/video0 -> camera0
func (s *WebSocketServer) getCameraIdentifierFromDevicePath(devicePath string) string {
	// Extract the number from /dev/video{N}
	if strings.HasPrefix(devicePath, "/dev/video") {
		number := strings.TrimPrefix(devicePath, "/dev/video")
		return fmt.Sprintf("camera%s", number)
	}
	// If it's already a camera identifier, return as is
	if strings.HasPrefix(devicePath, "camera") {
		return devicePath
	}
	// Fallback: return the original path
	return devicePath
}

// getDevicePathFromCameraIdentifier converts a camera identifier to a device path
// Example: camera0 -> /dev/video0
func (s *WebSocketServer) getDevicePathFromCameraIdentifier(cameraID string) string {
	// Extract the number from camera{N}
	if strings.HasPrefix(cameraID, "camera") {
		number := strings.TrimPrefix(cameraID, "camera")
		return fmt.Sprintf("/dev/video%s", number)
	}
	// If it's already a device path, return as is
	if strings.HasPrefix(cameraID, "/dev/video") {
		return cameraID
	}
	// Fallback: return the original identifier
	return cameraID
}

// validateCameraIdentifier validates that a camera identifier follows the correct pattern
func (s *WebSocketServer) validateCameraIdentifier(cameraID string) bool {
	// Must match pattern camera[0-9]+
	matched, _ := regexp.MatchString(`^camera[0-9]+$`, cameraID)
	return matched
}

// MethodCameraStatusUpdate handles camera status update notifications
// Following Python implementation patterns and API documentation specification
func (s *WebSocketServer) MethodCameraStatusUpdate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// REQ-API-020: WebSocket server shall support camera_status_update notifications
	// REQ-API-021: Notifications shall include device, status, name, resolution, fps, and streams

	s.logger.WithFields(logrus.Fields{
		"action": "camera_status_update",
		"client": client.ClientID,
		"role":   client.Role,
	}).Debug("Processing camera status update notification")

	// Validate required parameters per API documentation
	device, ok := params["device"].(string)
	if !ok || device == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required and must be a string",
			},
		}, nil
	}

	status, ok := params["status"].(string)
	if !ok || status == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "status parameter is required and must be a string",
			},
		}, nil
	}

	// Validate status values per API documentation
	validStatuses := []string{"CONNECTED", "DISCONNECTED", "ERROR", "RECORDING", "IDLE"}
	statusValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			statusValid = true
			break
		}
	}

	if !statusValid {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    fmt.Sprintf("status must be one of: %v", validStatuses),
			},
		}, nil
	}

	// Extract optional parameters
	name, _ := params["name"].(string)
	resolution, _ := params["resolution"].(string)
	fps, _ := params["fps"].(float64)
	streams, _ := params["streams"].(map[string]interface{})

	// Create notification structure per API documentation
	notification := map[string]interface{}{
		"device":     device,
		"status":     status,
		"name":       name,
		"resolution": resolution,
		"fps":        int(fps),
		"streams":    streams,
	}

	// Broadcast notification to all connected clients
	s.broadcastEvent("camera_status_update", notification)

	s.logger.WithFields(logrus.Fields{
		"action":    "camera_status_update",
		"device":    device,
		"status":    status,
		"client":    client.ClientID,
		"broadcast": true,
	}).Info("Camera status update notification broadcasted")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"success":      true,
			"notification": "camera_status_update",
			"device":       device,
			"status":       status,
			"broadcast":    true,
		},
	}, nil
}

// MethodRecordingStatusUpdate handles recording status update notifications
// Following Python implementation patterns and API documentation specification
func (s *WebSocketServer) MethodRecordingStatusUpdate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
	// REQ-API-022: WebSocket server shall support recording_status_update notifications
	// REQ-API-023: Notifications shall include device, status, filename, and duration

	s.logger.WithFields(logrus.Fields{
		"action": "recording_status_update",
		"client": client.ClientID,
		"role":   client.Role,
	}).Debug("Processing recording status update notification")

	// Validate required parameters per API documentation
	device, ok := params["device"].(string)
	if !ok || device == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "device parameter is required and must be a string",
			},
		}, nil
	}

	status, ok := params["status"].(string)
	if !ok || status == "" {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    "status parameter is required and must be a string",
			},
		}, nil
	}

	// Validate status values per API documentation
	validStatuses := []string{"STARTED", "STOPPED", "ERROR", "PAUSED", "RESUMED"}
	statusValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			statusValid = true
			break
		}
	}

	if !statusValid {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: ErrorMessages[INVALID_PARAMS],
				Data:    fmt.Sprintf("status must be one of: %v", validStatuses),
			},
		}, nil
	}

	// Extract optional parameters
	filename, _ := params["filename"].(string)
	duration, _ := params["duration"].(float64)

	// Create notification structure per API documentation
	notification := map[string]interface{}{
		"device":   device,
		"status":   status,
		"filename": filename,
		"duration": int64(duration),
	}

	// Broadcast notification to all connected clients
	s.broadcastEvent("recording_status_update", notification)

	s.logger.WithFields(logrus.Fields{
		"action":    "recording_status_update",
		"device":    device,
		"status":    status,
		"filename":  filename,
		"client":    client.ClientID,
		"broadcast": true,
	}).Info("Recording status update notification broadcasted")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"success":      true,
			"notification": "recording_status_update",
			"device":       device,
			"status":       status,
			"broadcast":    true,
		},
	}, nil
}
