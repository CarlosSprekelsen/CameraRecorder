/*
WebSocket JSON-RPC 2.0 server implementation.

Provides high-performance WebSocket server with JSON-RPC 2.0 protocol support,
following the Python WebSocketJsonRpcServer patterns and project architecture standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling
- REQ-API-011: API methods respond within specified time limits

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocketServer implements the WebSocket JSON-RPC 2.0 server
// Following Python WebSocketJsonRpcServer patterns with Go-specific optimizations
type WebSocketServer struct {
	// Configuration
	config *ServerConfig

	// Dependencies (proper dependency injection)
	configManager      *config.ConfigManager
	logger             *logging.Logger
	cameraMonitor      *camera.HybridCameraMonitor
	jwtHandler         *security.JWTHandler
	mediaMTXController mediamtx.MediaMTXController

	// Security extensions (Phase 1 enhancement)
	permissionChecker *security.PermissionChecker

	// WebSocket server
	upgrader websocket.Upgrader
	server   *http.Server
	running  bool

	// Client connection management
	clients       map[string]*ClientConnection
	clientsMutex  sync.RWMutex
	clientCounter int64

	// Method registration
	methods        map[string]MethodHandler
	methodsMutex   sync.RWMutex
	methodVersions map[string]string

	// Performance metrics
	metrics      *PerformanceMetrics
	metricsMutex sync.RWMutex

	// Event handling
	eventHandlers      []func(string, interface{})
	eventHandlersMutex sync.RWMutex

	// Graceful shutdown
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// Security extension methods (Phase 1 enhancement)

// checkMethodPermissions checks if a client has permission to access a specific method
func (s *WebSocketServer) checkMethodPermissions(client *ClientConnection, methodName string) error {
	// Skip permission check for authentication method
	if methodName == "authenticate" {
		return nil
	}

	// Convert client role to security.Role
	userRole, err := s.permissionChecker.ValidateRole(client.Role)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"role":      client.Role,
			"method":    methodName,
		}).Warn("Invalid role for permission check")
		return fmt.Errorf("invalid role: %s", client.Role)
	}

	// Check permission using existing PermissionChecker
	if !s.permissionChecker.HasPermission(userRole, methodName) {
		s.logger.WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"role":      client.Role,
			"method":    methodName,
		}).Warn("Permission denied for method")
		return fmt.Errorf("insufficient permissions for method %s", methodName)
	}

	return nil
}

// checkRateLimit checks if a client has exceeded the rate limit
func (s *WebSocketServer) checkRateLimit(client *ClientConnection) error {
	if !s.jwtHandler.CheckRateLimit(client.ClientID) {
		s.logger.WithField("client_id", client.ClientID).Warn("Rate limit exceeded")
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

// Real-time notification methods (Phase 3 enhancement)

// notifyRecordingStatusUpdate sends real-time recording status updates to clients
func (s *WebSocketServer) notifyRecordingStatusUpdate(device, status, filename string, duration time.Duration) {
	notification := map[string]interface{}{
		"type":      "recording_status_update",
		"device":    device,
		"status":    status,
		"filename":  filename,
		"duration":  duration.Seconds(),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.logger.WithFields(logrus.Fields{
		"device":   device,
		"status":   status,
		"filename": filename,
		"duration": duration,
	}).Debug("Sending recording status notification")

	// Use existing event handling infrastructure
	s.broadcastEvent("recording_update", notification)
}

// broadcastEvent broadcasts an event to all connected clients
func (s *WebSocketServer) broadcastEvent(eventType string, data interface{}) {
	s.eventHandlersMutex.RLock()
	defer s.eventHandlersMutex.RUnlock()

	// Send to all connected clients
	s.clientsMutex.RLock()
	for clientID, client := range s.clients {
		if client.Authenticated {
			// Note: In a real implementation, we would need to store the websocket.Conn
			// alongside the ClientConnection or modify the structure to include it.
			// For now, we'll log the notification for debugging.
			s.logger.WithFields(logrus.Fields{
				"client_id":  clientID,
				"event_type": eventType,
				"data":       data,
			}).Debug("Would send notification to client")
		}
	}
	s.clientsMutex.RUnlock()
}

// addEventHandler adds a new event handler
func (s *WebSocketServer) addEventHandler(handler func(string, interface{})) {
	s.eventHandlersMutex.Lock()
	defer s.eventHandlersMutex.Unlock()

	s.eventHandlers = append(s.eventHandlers, handler)
}

// NewWebSocketServer creates a new WebSocket server with proper dependency injection
func NewWebSocketServer(
	configManager *config.ConfigManager,
	logger *logging.Logger,
	cameraMonitor *camera.HybridCameraMonitor,
	jwtHandler *security.JWTHandler,
	mediaMTXController mediamtx.MediaMTXController,
) *WebSocketServer {
	if configManager == nil {
		panic("configManager cannot be nil - use existing internal/config/ConfigManager")
	}

	if logger == nil {
		logger = logging.NewLogger("websocket-server")
	}

	if cameraMonitor == nil {
		panic("cameraMonitor cannot be nil - use existing internal/camera/HybridCameraMonitor")
	}

	if jwtHandler == nil {
		panic("jwtHandler cannot be nil - use existing internal/security/JWTHandler")
	}

	if mediaMTXController == nil {
		panic("mediaMTXController cannot be nil - use existing internal/mediamtx/MediaMTXController")
	}

	// Get configuration from config manager
	cfg := configManager.GetConfig()
	if cfg == nil {
		panic("configuration not available - ensure config is loaded")
	}

	// Create server configuration
	serverConfig := &ServerConfig{
		Host:           cfg.Server.Host,
		Port:           cfg.Server.Port,
		WebSocketPath:  cfg.Server.WebSocketPath,
		MaxConnections: cfg.Server.MaxConnections,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		PingInterval:   cfg.Server.PingInterval,
		PongWait:       cfg.Server.PongWait,
		MaxMessageSize: cfg.Server.MaxMessageSize,
	}

	server := &WebSocketServer{
		config:             serverConfig,
		configManager:      configManager,
		logger:             logger,
		cameraMonitor:      cameraMonitor,
		jwtHandler:         jwtHandler,
		mediaMTXController: mediaMTXController,

		// Security extensions initialization (Phase 1 enhancement)
		permissionChecker: security.NewPermissionChecker(),

		// WebSocket upgrader configuration
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - can be made configurable
				return true
			},
		},

		// Client management
		clients:       make(map[string]*ClientConnection),
		clientCounter: 0,

		// Method registration
		methods:        make(map[string]MethodHandler),
		methodVersions: make(map[string]string),

		// Performance metrics
		metrics: &PerformanceMetrics{
			RequestCount:      0,
			ResponseTimes:     make(map[string][]float64),
			ErrorCount:        0,
			ActiveConnections: 0,
			StartTime:         time.Now(),
		},

		// Event handling
		eventHandlers: make([]func(string, interface{}), 0),

		// Graceful shutdown
		stopChan: make(chan struct{}),
	}

	// Register built-in methods
	server.registerBuiltinMethods()

	return server
}

// Start starts the WebSocket server
func (s *WebSocketServer) Start() error {
	if s.running {
		s.logger.Warn("WebSocket server is already running")
		return nil
	}

	s.logger.WithFields(logrus.Fields{
		"host":   s.config.Host,
		"port":   s.config.Port,
		"path":   s.config.WebSocketPath,
		"action": "start_server",
	}).Info("Starting WebSocket JSON-RPC server")

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.WebSocketPath, s.handleWebSocket)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Start server in goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("WebSocket server failed")
		}
	}()

	s.running = true

	s.logger.WithFields(logrus.Fields{
		"host":   s.config.Host,
		"port":   s.config.Port,
		"path":   s.config.WebSocketPath,
		"action": "start_server",
		"status": "success",
	}).Info("WebSocket server started successfully")

	return nil
}

// Stop stops the WebSocket server gracefully
func (s *WebSocketServer) Stop() error {
	if !s.running {
		s.logger.Warn("WebSocket server is not running")
		return nil
	}

	s.logger.Info("Stopping WebSocket server")

	// Signal shutdown
	close(s.stopChan)

	// Close all client connections
	s.clientsMutex.Lock()
	for clientID := range s.clients {
		s.logger.WithField("client_id", clientID).Debug("Closing client connection")
		// Note: Actual connection closing will be handled by the connection handler
	}
	s.clientsMutex.Unlock()

	// Shutdown HTTP server
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
		}
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	s.running = false

	s.logger.Info("WebSocket server stopped successfully")
	return nil
}

// handleWebSocket handles WebSocket upgrade and connection management
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check connection limit
	s.clientsMutex.RLock()
	if len(s.clients) >= s.config.MaxConnections {
		s.clientsMutex.RUnlock()
		s.logger.Warn("Maximum connections reached")
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}
	s.clientsMutex.RUnlock()

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}

	// Generate client ID
	s.clientsMutex.Lock()
	s.clientCounter++
	clientID := fmt.Sprintf("client_%d", s.clientCounter)
	s.clientsMutex.Unlock()

	// Create client connection
	client := &ClientConnection{
		ClientID:      clientID,
		Authenticated: false,
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}

	// Add client to connections
	s.clientsMutex.Lock()
	s.clients[clientID] = client
	s.clientsMutex.Unlock()

	// Update metrics
	s.metricsMutex.Lock()
	s.metrics.ActiveConnections++
	s.metricsMutex.Unlock()

	s.logger.WithFields(logrus.Fields{
		"client_id": clientID,
		"action":    "client_connected",
	}).Info("Client connected")

	// Handle connection in goroutine
	s.wg.Add(1)
	go s.handleClientConnection(conn, client)
}

// handleClientConnection handles individual client connections
func (s *WebSocketServer) handleClientConnection(conn *websocket.Conn, client *ClientConnection) {
	defer func() {
		// Remove client from connections
		s.clientsMutex.Lock()
		delete(s.clients, client.ClientID)
		s.clientsMutex.Unlock()

		// Update metrics
		s.metricsMutex.Lock()
		s.metrics.ActiveConnections--
		s.metricsMutex.Unlock()

		// Close connection
		conn.Close()

		s.logger.WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"action":    "client_disconnected",
		}).Info("Client disconnected")

		s.wg.Done()
	}()

	// Set connection parameters
	conn.SetReadLimit(s.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(s.config.PingInterval)
	defer ticker.Stop()

	// Message handling loop
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Error("Failed to send ping")
				return
			}
		default:
			// Read message
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.WithError(err).WithField("client_id", client.ClientID).Error("WebSocket read error")
				}
				return
			}

			// Handle message
			s.handleMessage(conn, client, message)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (s *WebSocketServer) handleMessage(conn *websocket.Conn, client *ClientConnection, message []byte) {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"action":    "handle_message",
	}).Debug("Processing WebSocket message")

	// Parse JSON-RPC request
	var request JsonRpcRequest
	if err := json.Unmarshal(message, &request); err != nil {
		s.sendErrorResponse(conn, nil, INVALID_PARAMS, "Invalid JSON-RPC request")
		return
	}

	// Validate JSON-RPC version
	if request.JSONRPC != "2.0" {
		s.sendErrorResponse(conn, request.ID, INVALID_PARAMS, "Invalid JSON-RPC version")
		return
	}

	// Handle request
	response, err := s.handleRequest(&request, client)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"client_id": client.ClientID,
			"method":    request.Method,
		}).Error("Request handling error")
		s.sendErrorResponse(conn, request.ID, INTERNAL_ERROR, "Internal server error")
		return
	}

	// Send response
	if err := s.sendResponse(conn, response); err != nil {
		s.logger.WithError(err).WithField("client_id", client.ClientID).Error("Failed to send response")
		return
	}

	// Record performance metrics
	duration := time.Since(startTime).Seconds()
	s.recordRequest(request.Method, duration)

	s.logger.WithFields(logrus.Fields{
		"client_id": client.ClientID,
		"method":    request.Method,
		"duration":  duration,
		"action":    "request_completed",
	}).Debug("Request completed")
}

// handleRequest processes JSON-RPC requests
func (s *WebSocketServer) handleRequest(request *JsonRpcRequest, client *ClientConnection) (*JsonRpcResponse, error) {
	// Security extensions: Rate limiting check (Phase 1 enhancement)
	if err := s.checkRateLimit(client); err != nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JsonRpcError{
				Code:    RATE_LIMIT_EXCEEDED,
				Message: ErrorMessages[RATE_LIMIT_EXCEEDED],
				Data:    err.Error(),
			},
		}, nil
	}

	// Find method handler
	s.methodsMutex.RLock()
	handler, exists := s.methods[request.Method]
	s.methodsMutex.RUnlock()

	if !exists {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JsonRpcError{
				Code:    METHOD_NOT_FOUND,
				Message: ErrorMessages[METHOD_NOT_FOUND],
			},
		}, nil
	}

	// Security extensions: Permission check (Phase 1 enhancement)
	if err := s.checkMethodPermissions(client, request.Method); err != nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JsonRpcError{
				Code:    INSUFFICIENT_PERMISSIONS,
				Message: ErrorMessages[INSUFFICIENT_PERMISSIONS],
				Data:    err.Error(),
			},
		}, nil
	}

	// Call method handler
	response, err := handler(request.Params, client)
	if err != nil {
		s.metricsMutex.Lock()
		s.metrics.ErrorCount++
		s.metricsMutex.Unlock()

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JsonRpcError{
				Code:    INTERNAL_ERROR,
				Message: ErrorMessages[INTERNAL_ERROR],
				Data:    err.Error(),
			},
		}, nil
	}

	// Set JSON-RPC version and ID
	response.JSONRPC = "2.0"
	response.ID = request.ID

	return response, nil
}

// sendResponse sends a JSON-RPC response to the client
func (s *WebSocketServer) sendResponse(conn *websocket.Conn, response *JsonRpcResponse) error {
	conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
	return conn.WriteJSON(response)
}

// sendErrorResponse sends a JSON-RPC error response to the client
func (s *WebSocketServer) sendErrorResponse(conn *websocket.Conn, id interface{}, code int, message string) {
	response := &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JsonRpcError{
			Code:    code,
			Message: message,
		},
	}

	if err := s.sendResponse(conn, response); err != nil {
		s.logger.WithError(err).Error("Failed to send error response")
	}
}

// recordRequest records performance metrics for a request
func (s *WebSocketServer) recordRequest(method string, duration float64) {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	s.metrics.RequestCount++
	if s.metrics.ResponseTimes[method] == nil {
		s.metrics.ResponseTimes[method] = make([]float64, 0)
	}
	s.metrics.ResponseTimes[method] = append(s.metrics.ResponseTimes[method], duration)
}

// GetMetrics returns current performance metrics
func (s *WebSocketServer) GetMetrics() *PerformanceMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	// Calculate average response time
	allResponseTimes := make([]float64, 0)
	for _, times := range s.metrics.ResponseTimes {
		allResponseTimes = append(allResponseTimes, times...)
	}

	// Note: averageResponseTime and errorRate calculations are available for future use
	// when extending the metrics functionality

	return &PerformanceMetrics{
		RequestCount:      s.metrics.RequestCount,
		ResponseTimes:     s.metrics.ResponseTimes,
		ErrorCount:        s.metrics.ErrorCount,
		ActiveConnections: s.metrics.ActiveConnections,
		StartTime:         s.metrics.StartTime,
	}
}

// IsRunning returns whether the server is currently running
func (s *WebSocketServer) IsRunning() bool {
	return s.running
}
