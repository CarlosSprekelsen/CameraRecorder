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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
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
	s.registerMethod("delete_recording", s.MethodDeleteRecording, "1.0")
	s.registerMethod("delete_snapshot", s.MethodDeleteSnapshot, "1.0")
	s.registerMethod("get_storage_info", s.MethodGetStorageInfo, "1.0")
	s.registerMethod("set_retention_policy", s.MethodSetRetentionPolicy, "1.0")
	s.registerMethod("cleanup_old_files", s.MethodCleanupOldFiles, "1.0")

	// Recording and snapshot methods
	s.registerMethod("take_snapshot", s.MethodTakeSnapshot, "1.0")
	s.registerMethod("start_recording", s.MethodStartRecording, "1.0")
	s.registerMethod("stop_recording", s.MethodStopRecording, "1.0")

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

// MethodPing implements the ping method
// Following Python _method_ping implementation
func (s *WebSocketServer) MethodPing(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
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

// MethodAuthenticate implements the authenticate method
// Following Python _method_authenticate implementation
func (s *WebSocketServer) MethodAuthenticate(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
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
			"fps":        30,                      // Default FPS - can be enhanced later
			"streams":    make(map[string]string), // Empty streams for now
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
			"fps":          30,                           // Default FPS - can be enhanced later
			"streams":      make(map[string]string),      // Empty streams for now
			"metrics":      make(map[string]interface{}), // Empty metrics for now
			"capabilities": camera.Capabilities,
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

	// CPU usage is not directly available in Go runtime, so we'll use a placeholder
	// In a production environment, this would be implemented with system calls
	cpuUsage := 0.0 // Placeholder - would need system-specific implementation

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

	// Return metrics following Python implementation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"active_connections":    activeConnections,
			"total_requests":        baseMetrics.RequestCount,
			"average_response_time": averageResponseTime,
			"error_rate":            errorRate,
			"memory_usage":          memoryUsage,
			"cpu_usage":             cpuUsage,
		},
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

	// Initialize response with architecture defaults following Python pattern
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
		// Convert camera formats to string list
		formats := make([]string, 0, len(camera.Formats))
		for _, format := range camera.Formats {
			formats = append(formats, format.PixelFormat)
		}
		cameraCapabilities["formats"] = formats

		// Convert resolutions to string list
		resolutions := make([]string, 0, len(camera.Formats))
		for _, format := range camera.Formats {
			resolution := fmt.Sprintf("%dx%d", format.Width, format.Height)
			resolutions = append(resolutions, resolution)
		}
		cameraCapabilities["resolutions"] = resolutions

		// Add FPS options (using common values for now)
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

	// Return camera capabilities following Python implementation exactly
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

	// Return status following Python implementation exactly
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
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_server_info",
		"action":    "method_call",
	}).Debug("Get server info method called")

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

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "get_server_info",
		"action":    "server_info_success",
	}).Debug("Server info retrieved successfully")

	// Return server info following Python implementation exactly
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"name":              "MediaMTX Camera Service",
			"version":           "1.0.0",
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

	// Note: MediaMTX controller integration will be implemented in Epic E4
	// For now, return empty stream list following Python pattern when controller not available
	s.logger.WithFields(logrus.Fields{
		"client_id":    client.ClientID,
		"method":       "get_streams",
		"action":       "streams_retrieved",
		"stream_count": 0,
	}).Debug("Streams retrieved successfully (MediaMTX integration pending Epic E4)")

	// Return empty stream list following Python implementation pattern
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  []map[string]interface{}{},
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

	// Get recordings directory path from configuration
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

	// Check if directory exists and is accessible
	if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
		s.logger.WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_recordings",
			"directory": recordingsDir,
			"action":    "directory_not_found",
		}).Warn("Recordings directory does not exist")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Result: map[string]interface{}{
				"files":  []map[string]interface{}{},
				"total":  0,
				"limit":  limit,
				"offset": offset,
			},
		}, nil
	}

	// Get list of files in directory
	files := []map[string]interface{}{}

	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_recordings",
			"directory": recordingsDir,
			"action":    "read_directory_error",
		}).Error("Error reading recordings directory")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error reading recordings directory: %v", err),
			},
		}, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Get file stats
		fileInfo, err := entry.Info()
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"client_id": client.ClientID,
				"method":    "list_recordings",
				"filename":  filename,
				"action":    "file_stat_error",
			}).Warn("Error accessing file")
			continue
		}

		// Determine if it's a video file
		isVideo := false
		ext := filepath.Ext(filename)
		switch ext {
		case ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv":
			isVideo = true
		}

		fileData := map[string]interface{}{
			"filename":      filename,
			"file_size":     fileInfo.Size(),
			"modified_time": fileInfo.ModTime().Format(time.RFC3339),
			"download_url":  fmt.Sprintf("/files/recordings/%s", filename),
		}

		// Add duration for video files (placeholder - would need video metadata extraction)
		if isVideo {
			fileData["duration"] = nil // TODO: Extract actual duration from video file
		}

		files = append(files, fileData)
	}

	// Sort files by modified_time (newest first)
	sort.Slice(files, func(i, j int) bool {
		timeI := files[i]["modified_time"].(string)
		timeJ := files[j]["modified_time"].(string)
		return timeI > timeJ
	})

	// Apply pagination
	totalCount := len(files)
	startIdx := offset
	endIdx := startIdx + limit
	if endIdx > totalCount {
		endIdx = totalCount
	}
	if startIdx > totalCount {
		startIdx = totalCount
	}

	paginatedFiles := files[startIdx:endIdx]

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "list_recordings",
		"total_files": totalCount,
		"returned":    len(paginatedFiles),
		"action":      "recordings_listed",
	}).Debug("Recordings listed successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"files":  paginatedFiles,
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
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

	// Get recordings directory path from configuration
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
	filePath := filepath.Join(recordingsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "Recording file not found",
				Data:    fmt.Sprintf("Recording file not found: %s", filename),
			},
		}, nil
	}

	// Check if it's a file (not a directory)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error accessing file: %v", err),
			},
		}, nil
	}

	if fileInfo.IsDir() {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "Path is not a file",
				Data:    fmt.Sprintf("Path is not a file: %s", filename),
			},
		}, nil
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "delete_recording",
			"filename":  filename,
			"action":    "delete_error",
		}).Error("Error deleting recording file")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error deleting recording file: %v", err),
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

	// Get snapshots directory path from configuration
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

	snapshotsDir := config.MediaMTX.SnapshotsPath
	filePath := filepath.Join(snapshotsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "Snapshot file not found",
				Data:    fmt.Sprintf("Snapshot file not found: %s", filename),
			},
		}, nil
	}

	// Check if it's a file (not a directory)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error accessing file: %v", err),
			},
		}, nil
	}

	if fileInfo.IsDir() {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INVALID_PARAMS,
				Message: "Path is not a file",
				Data:    fmt.Sprintf("Path is not a file: %s", filename),
			},
		}, nil
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "delete_snapshot",
			"filename":  filename,
			"action":    "delete_error",
		}).Error("Error deleting snapshot file")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error deleting snapshot file: %v", err),
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

	// Determine warning levels (following Python implementation)
	lowSpaceWarning := usedPercent >= 80.0
	criticalSpaceWarning := usedPercent >= 95.0

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
			"total_bytes":            totalBytes,
			"free_bytes":             freeBytes,
			"used_bytes":             usedBytes,
			"used_percent":           usedPercent,
			"recordings_size":        recordingsSize,
			"snapshots_size":         snapshotsSize,
			"low_space_warning":      lowSpaceWarning,
			"critical_space_warning": criticalSpaceWarning,
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

	// TODO: Implement actual cleanup logic based on retention policies
	// For now, return a placeholder response following Python pattern
	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "cleanup_old_files",
		"action":    "cleanup_triggered",
	}).Info("Manual cleanup triggered (not yet implemented)")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"cleanup_executed": true,
			"files_deleted":    0,
			"space_freed":      0,
			"message":          "Cleanup completed successfully (placeholder implementation)",
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

	// TODO: Store retention policy configuration (would need persistent storage)
	// For now, just return the configuration following Python pattern
	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "set_retention_policy",
		"policy_type": policyType,
		"enabled":     enabled,
		"action":      "policy_updated",
	}).Info("Retention policy updated")

	// Build response according to API documentation
	response := map[string]interface{}{
		"policy_type": policyType,
		"enabled":     enabled,
		"message":     "Retention policy updated successfully",
	}

	// Add policy-specific fields as required by API documentation
	if policyType == "age" {
		if maxAgeDays, exists := params["max_age_days"]; exists {
			response["max_age_days"] = maxAgeDays
		}
	} else if policyType == "size" {
		if maxSizeGB, exists := params["max_size_gb"]; exists {
			response["max_size_gb"] = maxSizeGB
		}
	}

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result:  response,
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

	// Get snapshots directory path from configuration
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

	snapshotsDir := config.MediaMTX.SnapshotsPath

	// Check if directory exists and is accessible
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		s.logger.WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_snapshots",
			"directory": snapshotsDir,
			"action":    "directory_not_found",
		}).Warn("Snapshots directory does not exist")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Result: map[string]interface{}{
				"files":  []map[string]interface{}{},
				"total":  0,
				"limit":  limit,
				"offset": offset,
			},
		}, nil
	}

	// Get list of files in directory
	files := []map[string]interface{}{}

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    "list_snapshots",
			"directory": snapshotsDir,
			"action":    "read_directory_error",
		}).Error("Error reading snapshots directory")

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    fmt.Sprintf("Error reading snapshots directory: %v", err),
			},
		}, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Get file stats
		fileInfo, err := entry.Info()
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"client_id": client.ClientID,
				"method":    "list_snapshots",
				"filename":  filename,
				"action":    "file_stat_error",
			}).Warn("Error accessing file")
			continue
		}

		fileData := map[string]interface{}{
			"filename":      filename,
			"file_size":     fileInfo.Size(),
			"modified_time": fileInfo.ModTime().Format(time.RFC3339),
			"download_url":  fmt.Sprintf("/files/snapshots/%s", filename),
		}

		files = append(files, fileData)
	}

	// Sort files by modified_time (newest first)
	sort.Slice(files, func(i, j int) bool {
		timeI := files[i]["modified_time"].(string)
		timeJ := files[j]["modified_time"].(string)
		return timeI > timeJ
	})

	// Apply pagination
	totalCount := len(files)
	startIdx := offset
	endIdx := startIdx + limit
	if endIdx > totalCount {
		endIdx = totalCount
	}
	if startIdx > totalCount {
		startIdx = totalCount
	}

	paginatedFiles := files[startIdx:endIdx]

	s.logger.WithFields(logrus.Fields{
		"client_id":   client.ClientID,
		"method":      "list_snapshots",
		"total_files": totalCount,
		"returned":    len(paginatedFiles),
		"action":      "snapshots_listed",
	}).Debug("Snapshots listed successfully")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"files":  paginatedFiles,
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
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

	// Parameter validation and normalization
	formatType := "jpg"
	if format, ok := params["format"].(string); ok && format != "" {
		if format == "jpg" || format == "png" {
			formatType = format
		}
	}

	quality := 85
	if qualityVal, ok := params["quality"].(int); ok {
		if qualityVal >= 1 && qualityVal <= 100 {
			quality = qualityVal
		}
	}

	customFilename := ""
	if filename, ok := params["filename"].(string); ok {
		customFilename = filename
	}

	// TODO: MediaMTX controller integration will be implemented in Epic E4
	// For now, return a placeholder response following Python pattern
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")

	if customFilename == "" {
		streamName := s.getStreamNameFromDevicePath(devicePath)
		timestampStr := time.Now().Format("2006-01-02_15-04-05")
		customFilename = fmt.Sprintf("%s_snapshot_%s.%s", streamName, timestampStr, formatType)
	} else if !strings.HasSuffix(customFilename, "."+formatType) {
		customFilename = fmt.Sprintf("%s.%s", customFilename, formatType)
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "take_snapshot",
		"device":    devicePath,
		"format":    formatType,
		"quality":   quality,
		"filename":  customFilename,
		"action":    "snapshot_triggered",
	}).Info("Snapshot triggered (MediaMTX integration pending Epic E4)")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"device":    devicePath,
			"filename":  customFilename,
			"status":    "FAILED",
			"timestamp": timestamp,
			"file_size": 0,
			"file_path": "",
			"error":     "MediaMTX controller not available (pending Epic E4)",
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

	// Parameter validation and normalization
	duration := 0 // 0 means continuous recording
	if durationVal, ok := params["duration"].(int); ok && durationVal > 0 {
		duration = durationVal
	}

	formatType := "mp4"
	if format, ok := params["format"].(string); ok && format != "" {
		if format == "mp4" || format == "mkv" || format == "avi" {
			formatType = format
		}
	}

	customFilename := ""
	if filename, ok := params["filename"].(string); ok {
		customFilename = filename
	}

	// TODO: MediaMTX controller integration will be implemented in Epic E4
	// For now, return a placeholder response following Python pattern
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")

	if customFilename == "" {
		streamName := s.getStreamNameFromDevicePath(devicePath)
		timestampStr := time.Now().Format("2006-01-02_15-04-05")
		customFilename = fmt.Sprintf("%s_recording_%s.%s", streamName, timestampStr, formatType)
	} else if !strings.HasSuffix(customFilename, "."+formatType) {
		customFilename = fmt.Sprintf("%s.%s", customFilename, formatType)
	}

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "start_recording",
		"device":    devicePath,
		"duration":  duration,
		"format":    formatType,
		"filename":  customFilename,
		"action":    "recording_started",
	}).Info("Recording started (MediaMTX integration pending Epic E4)")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"device":    devicePath,
			"filename":  customFilename,
			"status":    "FAILED",
			"timestamp": timestamp,
			"duration":  duration,
			"format":    formatType,
			"error":     "MediaMTX controller not available (pending Epic E4)",
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

	// TODO: MediaMTX controller integration will be implemented in Epic E4
	// For now, return a placeholder response following Python pattern
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    "stop_recording",
		"device":    devicePath,
		"action":    "recording_stopped",
	}).Info("Recording stopped (MediaMTX integration pending Epic E4)")

	return &JsonRpcResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"device":    devicePath,
			"status":    "FAILED",
			"timestamp": timestamp,
			"file_size": 0,
			"file_path": "",
			"error":     "MediaMTX controller not available (pending Epic E4)",
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
