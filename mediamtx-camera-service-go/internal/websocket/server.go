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
	"strconv"
	"sync"
	"sync/atomic"
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
//
// Thread Safety: This struct is designed to be thread-safe for concurrent operations.
// All shared state is protected by appropriate mutexes:
// - clientsMutex: Protects clients map and clientCounter
// - metricsMutex: Protects metrics struct
// - methodsMutex: Protects methods map
// - eventHandlersMutex: Protects eventHandlers slice
// - stopOnce: Ensures single close operation on stopChan
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
	clientCounter int64 // Protected by clientsMutex

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
	stopOnce sync.Once
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
		if client.Authenticated && client.Conn != nil {
			// Create notification message
			notification := &JsonRpcNotification{
				JSONRPC: "2.0",
				Method:  eventType,
				Params:  data.(map[string]interface{}),
			}

			// Send message to client
			client.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
			if err := client.Conn.WriteJSON(notification); err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"client_id":  clientID,
					"event_type": eventType,
				}).Error("Failed to send notification to client")
			} else {
				s.logger.WithFields(logrus.Fields{
					"client_id":  clientID,
					"event_type": eventType,
				}).Debug("Notification sent to client")
			}
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
) (*WebSocketServer, error) {
	if configManager == nil {
		return nil, fmt.Errorf("configManager cannot be nil - use existing internal/config/ConfigManager")
	}

	if logger == nil {
		logger = logging.NewLogger("websocket-server")
	}

	if cameraMonitor == nil {
		return nil, fmt.Errorf("cameraMonitor cannot be nil - use existing internal/camera/HybridCameraMonitor")
	}

	if jwtHandler == nil {
		return nil, fmt.Errorf("jwtHandler cannot be nil - use existing internal/security/JWTHandler")
	}

	if mediaMTXController == nil {
		return nil, fmt.Errorf("mediaMTXController cannot be nil - use existing internal/mediamtx/MediaMTXController")
	}

	// Get configuration from config manager
	cfg := configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not available - ensure config is loaded")
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
		stopChan: make(chan struct{}, 10), // Buffered to prevent deadlock during shutdown
		stopOnce: sync.Once{},
	}

	// Register built-in methods
	server.registerBuiltinMethods()

	return server, nil
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
			// Note: Error is logged but not returned as this is in a goroutine
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

	// Signal shutdown - use sync.Once to ensure single close operation
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})

	// Close all client connections with timeout
	s.closeAllClientConnections()

	// Shutdown HTTP server
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
			// Note: Error is logged but not returned as this is cleanup operation
		}
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	s.running = false

	s.logger.Info("WebSocket server stopped successfully")
	return nil
}

// closeAllClientConnections closes all client connections with timeout
func (s *WebSocketServer) closeAllClientConnections() {
	s.logger.Info("Starting client connection cleanup")

	// Create cleanup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ClientCleanupTimeout)
	defer cancel()

	// Get list of clients to close
	s.clientsMutex.Lock()
	clientsToClose := make([]*ClientConnection, 0, len(s.clients))
	for clientID, client := range s.clients {
		clientsToClose = append(clientsToClose, client)
		s.logger.WithField("client_id", clientID).Debug("Queuing client connection for cleanup")
	}
	s.clientsMutex.Unlock()

	if len(clientsToClose) == 0 {
		s.logger.Debug("No client connections to clean up")
		return
	}

	// Close connections concurrently with timeout
	var wg sync.WaitGroup
	cleanupResults := make(chan error, len(clientsToClose))

	for _, client := range clientsToClose {
		wg.Add(1)
		go func(client *ClientConnection) {
			defer wg.Done()

			// Set close deadline
			client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

			// Send close message
			closeMsg := websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutdown")
			if err := client.Conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(5*time.Second)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to send close message")
			}

			// Close connection
			if err := client.Conn.Close(); err != nil {
				cleanupResults <- fmt.Errorf("failed to close connection for client %s: %w", client.ClientID, err)
				return
			}

			// Remove client from map and update metrics atomically
			s.clientsMutex.Lock()
			delete(s.clients, client.ClientID)
			s.clientsMutex.Unlock()

			// Use atomic operation for metrics update
			atomic.AddInt64(&s.metrics.ActiveConnections, -1)

			cleanupResults <- nil
		}(client)
	}

	// Wait for cleanup with timeout
	cleanupDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(cleanupDone)
	}()

	select {
	case <-ctx.Done():
		s.logger.Warn("Client cleanup timeout reached, forcing connection closure")
		// Force close remaining connections
		for _, client := range clientsToClose {
			client.Conn.Close()
		}
	case <-cleanupDone:
		s.logger.Debug("All client connections cleaned up successfully")
	}

	// Check cleanup results
	close(cleanupResults)
	errorCount := 0
	for err := range cleanupResults {
		if err != nil {
			errorCount++
			s.logger.WithError(err).Warn("Client cleanup error")
		}
	}

	if errorCount > 0 {
		s.logger.WithField("error_count", fmt.Sprintf("%d", errorCount)).Warn("Some client connections had cleanup errors")
	} else {
		s.logger.Info("All client connections cleaned up successfully")
	}
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

	// Generate client ID with proper synchronization
	s.clientsMutex.Lock()
	s.clientCounter++
	clientID := "client_" + strconv.FormatInt(s.clientCounter, 10)
	s.clientsMutex.Unlock()

	// Create client connection
	client := &ClientConnection{
		ClientID:      clientID,
		Authenticated: false,
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
		Conn:          conn,
	}

	// Add client to connections and update metrics atomically
	s.clientsMutex.Lock()
	s.clients[clientID] = client
	s.clientsMutex.Unlock()

	// Update metrics with atomic operation
	atomic.AddInt64(&s.metrics.ActiveConnections, 1)

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
		// Recover from panics in goroutine
		if r := recover(); r != nil {
			s.logger.WithFields(logrus.Fields{
				"client_id": client.ClientID,
				"panic":     r,
				"action":    "panic_recovered",
			}).Error("Recovered from panic in client connection handler")
		}

		// Remove client from connections and update metrics atomically
		s.clientsMutex.Lock()
		delete(s.clients, client.ClientID)
		s.clientsMutex.Unlock()

		// Update metrics with atomic operation
		atomic.AddInt64(&s.metrics.ActiveConnections, -1)

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

	// Create message handling context with timeout
	msgCtx, msgCancel := context.WithCancel(context.Background())
	defer msgCancel()

	// Message handling loop
	for {
		select {
		case <-s.stopChan:
			s.logger.WithField("client_id", client.ClientID).Debug("Server shutdown signal received, closing client connection")
			return
		case <-msgCtx.Done():
			s.logger.WithField("client_id", client.ClientID).Debug("Message context cancelled, closing client connection")
			return
		case <-ticker.C:
			// Set write deadline for ping
			conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Error("Failed to send ping")
				return
			}
		default:
			// Set read deadline for message
			conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

			// Read message with timeout
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.WithError(err).WithField("client_id", client.ClientID).Error("WebSocket read error")
				} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					s.logger.WithField("client_id", client.ClientID).Debug("Client connection closed normally")
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
	// Skip permission check for authenticate method
	if request.Method != "authenticate" {
		if err := s.checkMethodPermissions(client, request.Method); err != nil {
			// Check if client is not authenticated
			if !client.Authenticated {
				return &JsonRpcResponse{
					JSONRPC: "2.0",
					ID:      request.ID,
					Error: &JsonRpcError{
						Code:    AUTHENTICATION_REQUIRED,
						Message: ErrorMessages[AUTHENTICATION_REQUIRED],
					},
				}, nil
			}

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
	}

	// Call method handler
	response, err := handler(request.Params, client)
	if err != nil {
		// Update error metrics with atomic operation
		atomic.AddInt64(&s.metrics.ErrorCount, 1)

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
	// Use atomic operation for RequestCount
	atomic.AddInt64(&s.metrics.RequestCount, 1)

	// ResponseTimes still needs mutex protection due to map operations
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	if s.metrics.ResponseTimes[method] == nil {
		s.metrics.ResponseTimes[method] = make([]float64, 0)
	}
	s.metrics.ResponseTimes[method] = append(s.metrics.ResponseTimes[method], duration)
}

// GetMetrics returns current performance metrics
func (s *WebSocketServer) GetMetrics() *PerformanceMetrics {
	// Use atomic operations for reading counters
	requestCount := atomic.LoadInt64(&s.metrics.RequestCount)
	errorCount := atomic.LoadInt64(&s.metrics.ErrorCount)
	activeConnections := atomic.LoadInt64(&s.metrics.ActiveConnections)

	// ResponseTimes still needs mutex protection due to map operations
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	// Calculate average response time
	allResponseTimes := make([]float64, 0)
	for _, times := range s.metrics.ResponseTimes {
		allResponseTimes = append(allResponseTimes, times...)
	}

	// Note: averageResponseTime and errorRate calculations are available for future use
	// when extending the metrics functionality

	// Create a deep copy to prevent race conditions
	responseTimesCopy := make(map[string][]float64)
	for method, times := range s.metrics.ResponseTimes {
		timesCopy := make([]float64, len(times))
		copy(timesCopy, times)
		responseTimesCopy[method] = timesCopy
	}

	return &PerformanceMetrics{
		RequestCount:      requestCount,
		ResponseTimes:     responseTimesCopy,
		ErrorCount:        errorCount,
		ActiveConnections: activeConnections,
		StartTime:         s.metrics.StartTime,
	}
}

// IsRunning returns whether the server is currently running
func (s *WebSocketServer) IsRunning() bool {
	return s.running
}

// GetConfig returns the server configuration (for testing purposes)
func (s *WebSocketServer) GetConfig() *ServerConfig {
	return s.config
}

// SetConfig sets the server configuration (for testing purposes)
func (s *WebSocketServer) SetConfig(config *ServerConfig) {
	s.config = config
}
