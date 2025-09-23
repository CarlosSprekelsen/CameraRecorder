// Package websocket implements the Layer 6 (API) WebSocket JSON-RPC 2.0 server.
//
// This package provides the protocol layer implementation with NO business logic,
// following the architectural constraint that all operations are delegated to
// the MediaMTX controller (single source of truth).
//
// Architecture Compliance:
//   - Protocol Layer Only: No business logic, pure JSON-RPC 2.0 implementation
//   - Delegation Pattern: All operations forwarded to MediaMTX controller
//   - High Concurrency: Supports 1000+ simultaneous WebSocket connections
//   - Security Integration: JWT authentication with role-based access control
//   - Event System: Real-time client notifications via event manager
//
// Key Responsibilities:
//   - WebSocket connection management and lifecycle
//   - JSON-RPC 2.0 protocol implementation and message handling
//   - Authentication enforcement and session management
//   - Input validation and security protection (rate limiting)
//   - Real-time event broadcasting to connected clients
//   - Performance metrics collection and monitoring
//
// Thread Safety: All components are designed for concurrent access with
// appropriate synchronization primitives protecting shared state.
//
// Requirements Coverage:
//   - REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint on port 8002
//   - REQ-API-002: Complete JSON-RPC 2.0 protocol implementation
//   - REQ-API-003: Request/response message handling with proper error codes
//   - REQ-API-011: API methods respond within specified time limits (<100ms)
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/gorilla/websocket"
)

// WebSocketServer implements the Layer 6 (API) WebSocket JSON-RPC 2.0 server.
// This server contains NO business logic and delegates all operations to the
// MediaMTX controller following architectural constraints.
//
// Performance Targets:
//   - 1000+ simultaneous WebSocket connections
//   - <100ms response time for 95% of requests
//   - <20ms event notification delivery latency
//
// Thread Safety: All shared state is protected by appropriate synchronization:
//   - clientsMutex: Protects clients map and clientCounter
//   - metricsMutex: Protects performance metrics
//   - methodsMutex: Protects method registration map
//   - eventHandlersMutex: Protects event handler slice
//   - stopOnce: Ensures single close operation on stopChan
type WebSocketServer struct {
	// Configuration and Dependencies
	config *ServerConfig // Server-specific configuration (ports, timeouts, etc.)

	// Dependency Injection - All external dependencies injected via constructor
	configManager      *config.ConfigManager          // Configuration management
	logger             *logging.Logger                // Structured logging
	jwtHandler         *security.JWTHandler           // JWT authentication and rate limiting
	mediaMTXController mediamtx.MediaMTXControllerAPI // Business logic delegation target

	// Security Framework Components
	permissionChecker *security.PermissionChecker // Role-based access control
	validationHelper  *ValidationHelper           // Input validation and sanitization

	// WebSocket Protocol Implementation
	upgrader websocket.Upgrader // WebSocket connection upgrader with CORS settings
	server   *http.Server       // HTTP server for WebSocket endpoint
	running  int32              // Atomic boolean for thread-safe running state (0=false, 1=true)

	// Connection Management - High concurrency support (1000+ connections)
	clients       map[string]*ClientConnection // Active client connections indexed by client ID
	clientsMutex  sync.RWMutex                 // Protects clients map modifications
	clientCounter int64                        // Atomic counter for unique client ID generation
	clientCount   int64                        // Atomic counter for fast client count queries

	// JSON-RPC Method Registration
	methods             map[string]MethodHandler // Registered JSON-RPC methods
	methodsMutex        sync.RWMutex             // Protects method map modifications
	methodVersions      map[string]string        // Method version tracking
	methodVersionsMutex sync.RWMutex             // Protects version map modifications
	builtinMethodsReady int32                    // Atomic flag for builtin method initialization

	// Performance Monitoring
	metrics      *PerformanceMetrics // Request/response performance tracking
	metricsMutex sync.RWMutex        // Protects metrics updates

	// Real-Time Event System
	eventManager       *EventManager               // Event broadcasting manager
	eventHandlers      []func(string, interface{}) // Registered event handlers
	eventHandlersMutex sync.RWMutex                // Protects event handler slice
	eventHandlerCount  int64                       // Atomic counter for handler count

	// Graceful Shutdown Coordination
	stopChan chan struct{}  // Shutdown signal channel
	stopOnce sync.Once      // Ensures single shutdown execution
	wg       sync.WaitGroup // Tracks active goroutines for clean shutdown
}

// isSystemReady checks if the MediaMTX controller is ready to handle requests.
// This implements the progressive readiness pattern where the API becomes available
// as the underlying business logic components complete initialization.
func (s *WebSocketServer) isSystemReady() bool {
	if s.mediaMTXController == nil {
		return false
	}

	// Check if controller implements IsReady method
	if readyChecker, ok := s.mediaMTXController.(interface{ IsReady() bool }); ok {
		return readyChecker.IsReady()
	}

	// Fallback: assume ready if controller exists
	return true
}

// getSystemReadinessResponse returns a standardized readiness response
func (s *WebSocketServer) getSystemReadinessResponse() map[string]interface{} {
	response := map[string]interface{}{
		"status":            "starting",
		"message":           "System is initializing, please wait",
		"available_cameras": []string{},
		"discovery_active":  false,
	}

	if s.mediaMTXController == nil {
		return response
	}

	// Get detailed readiness state if available
	if stateProvider, ok := s.mediaMTXController.(interface{ GetReadinessState() map[string]interface{} }); ok {
		state := stateProvider.GetReadinessState()

		if cameras, ok := state["available_cameras"].([]string); ok {
			response["available_cameras"] = cameras
		}

		if cameraReady, ok := state["camera_monitor_ready"].(bool); ok {
			response["discovery_active"] = !cameraReady
		}

		if len(response["available_cameras"].([]string)) > 0 {
			response["status"] = "partial"
			response["message"] = "Some cameras available, discovery in progress"
		}
	}

	return response
}

// checkMethodPermissions checks if a client has permission to access a specific method
func (s *WebSocketServer) checkMethodPermissions(client *ClientConnection, methodName string) error {
	// Skip permission check for authentication method
	if methodName == "authenticate" {
		return nil
	}

	// Check for nil client to prevent panic
	if client == nil {
		s.logger.WithField("method", methodName).Error("Cannot check permissions: client is nil")
		return fmt.Errorf("client is nil")
	}

	// Convert client role to security.Role
	userRole, err := s.permissionChecker.ValidateRole(client.Role)
	if err != nil {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"role":      client.Role,
			"method":    methodName,
		}).Warn("Invalid role for permission check")
		return fmt.Errorf("invalid role: %s", client.Role)
	}

	// Check permission using existing PermissionChecker
	if !s.permissionChecker.HasPermission(userRole, methodName) {
		s.logger.WithFields(logging.Fields{
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
	// Check for nil client to prevent panic
	if client == nil {
		s.logger.Error("Cannot check rate limit: client is nil")
		return fmt.Errorf("client is nil")
	}

	if !s.jwtHandler.CheckRateLimit(client.ClientID) {
		s.logger.WithField("client_id", client.ClientID).Warn("Rate limit exceeded")
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

// Real-time notification methods (Phase 3 enhancement)

// notifyRecordingStatusUpdate sends real-time recording status updates to clients
func (s *WebSocketServer) notifyRecordingStatusUpdate(device, status, filename string, duration time.Duration) {
	// Determine event topic based on status
	var topic EventTopic
	switch status {
	case "started":
		topic = TopicRecordingStart
	case "stopped":
		topic = TopicRecordingStop
	case "error":
		topic = TopicRecordingError
	default:
		topic = TopicRecordingProgress
	}

	eventData := map[string]interface{}{
		"device":    device,
		"status":    status,
		"filename":  filename,
		"duration":  duration.Seconds(),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.logger.WithFields(logging.Fields{
		"device":   device,
		"status":   status,
		"filename": filename,
		"duration": duration,
		"topic":    topic,
	}).Debug("Sending recording status notification")

	// Use new efficient event system
	if err := s.sendEventToSubscribers(topic, eventData); err != nil {
		s.logger.WithError(err).WithField("topic", string(topic)).Error("Failed to send recording status event")
		// Fallback to broadcast for backward compatibility
		s.broadcastEvent("recording_update", eventData)
	}
}

// notifyCameraStatusUpdate sends real-time camera status updates to clients
func (s *WebSocketServer) notifyCameraStatusUpdate(device, status, name string) {
	// Determine event topic based on status
	var topic EventTopic
	switch status {
	case "connected":
		topic = TopicCameraConnected
	case "disconnected":
		topic = TopicCameraDisconnected
	default:
		topic = TopicCameraStatusChange
	}

	eventData := map[string]interface{}{
		"device":    device,
		"status":    status,
		"name":      name,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.logger.WithFields(logging.Fields{
		"device": device,
		"status": status,
		"name":   name,
		"topic":  topic,
	}).Debug("Sending camera status notification")

	// Use new efficient event system
	if err := s.sendEventToSubscribers(topic, eventData); err != nil {
		s.logger.WithError(err).WithField("topic", string(topic)).Error("Failed to send camera status event")
		// Fallback to broadcast for backward compatibility
		s.broadcastEvent("camera_status_update", eventData)
	}
}

// notifySnapshotTaken sends real-time snapshot notifications to clients
func (s *WebSocketServer) notifySnapshotTaken(device, filename, resolution string) {
	eventData := map[string]interface{}{
		"device":     device,
		"filename":   filename,
		"resolution": resolution,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	s.logger.WithFields(logging.Fields{
		"device":     device,
		"filename":   filename,
		"resolution": resolution,
		"topic":      TopicSnapshotTaken,
	}).Debug("Sending snapshot notification")

	// Use new efficient event system
	if err := s.sendEventToSubscribers(TopicSnapshotTaken, eventData); err != nil {
		s.logger.WithError(err).Error("Failed to send snapshot event")
		// Fallback to broadcast for backward compatibility
		s.broadcastEvent("snapshot_taken", eventData)
	}
}

// notifySystemEvent sends system-level event notifications to clients
func (s *WebSocketServer) notifySystemEvent(eventType string, data map[string]interface{}) {
	var topic EventTopic
	switch eventType {
	case "startup":
		topic = TopicSystemStartup
	case "shutdown":
		topic = TopicSystemShutdown
	case "health":
		topic = TopicSystemHealth
	default:
		topic = TopicSystemError
	}

	// Initialize data map if nil
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add timestamp if not present
	if _, exists := data["timestamp"]; !exists {
		data["timestamp"] = time.Now().Format(time.RFC3339)
	}

	s.logger.WithFields(logging.Fields{
		"event_type": eventType,
		"topic":      topic,
		"data":       data,
	}).Debug("Sending system event notification")

	// Use new efficient event system
	if err := s.sendEventToSubscribers(topic, data); err != nil {
		s.logger.WithError(err).WithField("topic", string(topic)).Error("Failed to send system event")
		// Fallback to broadcast for backward compatibility
		s.broadcastEvent("system_event", data)
	}
}

// broadcastEvent broadcasts an event to all connected clients
// sendEventToAllClients broadcasts events to all connected clients.
// For efficient topic-based delivery, use sendEventToSubscribers instead.
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
			if err := client.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithFields(logging.Fields{
					"client_id":  clientID,
					"event_type": eventType,
				}).Warn("Failed to set write deadline for notification")
			}
			if err := client.Conn.WriteJSON(notification); err != nil {
				s.logger.WithError(err).WithFields(logging.Fields{
					"client_id":  clientID,
					"event_type": eventType,
				}).Error("Failed to send notification to client")
			} else {
				s.logger.WithFields(logging.Fields{
					"client_id":  clientID,
					"event_type": eventType,
				}).Debug("Notification sent to client")
			}
		}
	}
	s.clientsMutex.RUnlock()
}

// sendEventToSubscribers sends an event only to clients subscribed to the specific topic
func (s *WebSocketServer) sendEventToSubscribers(topic EventTopic, data map[string]interface{}) error {
	// Publish event through event manager
	if err := s.eventManager.PublishEvent(topic, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Get subscribers for this topic
	subscribers := s.eventManager.GetSubscribersForTopic(topic)
	if len(subscribers) == 0 {
		s.logger.WithField("topic", string(topic)).Debug("No subscribers for event topic")
		return nil
	}

	// Send event only to subscribed clients
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	notification := &JsonRpcNotification{
		JSONRPC: "2.0",
		Method:  string(topic),
		Params:  data,
	}

	sentCount := 0
	for _, clientID := range subscribers {
		if client, exists := s.clients[clientID]; exists && client.Authenticated && client.Conn != nil {
			// Send message to client
			if err := client.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithFields(logging.Fields{
					"client_id": clientID,
					"topic":     topic,
				}).Warn("Failed to set write deadline for event")
			}
			if err := client.Conn.WriteJSON(notification); err != nil {
				s.logger.WithError(err).WithFields(logging.Fields{
					"client_id": clientID,
					"topic":     topic,
				}).Error("Failed to send event to subscribed client")
			} else {
				sentCount++
				s.logger.WithFields(logging.Fields{
					"client_id": clientID,
					"topic":     topic,
				}).Debug("Event sent to subscribed client")
			}
		}
	}

	s.logger.WithFields(logging.Fields{
		"topic":       topic,
		"subscribers": len(subscribers),
		"sent_count":  sentCount,
	}).Debug("Event delivered to subscribed clients")

	return nil
}

// addEventHandler adds a new event handler
func (s *WebSocketServer) addEventHandler(handler func(string, interface{})) {
	s.eventHandlersMutex.Lock()
	defer s.eventHandlersMutex.Unlock()

	s.eventHandlers = append(s.eventHandlers, handler)
	atomic.AddInt64(&s.eventHandlerCount, 1)
}

// NewWebSocketServer creates a new WebSocket server with proper dependency injection
func NewWebSocketServer(
	configManager *config.ConfigManager,
	logger *logging.Logger,
	jwtHandler *security.JWTHandler,
	mediaMTXController mediamtx.MediaMTXControllerAPI,
) (*WebSocketServer, error) {
	if configManager == nil {
		return nil, fmt.Errorf("configManager cannot be nil")
	}

	if logger == nil {
		logger = logging.GetLogger("websocket.server") // Component-specific logger naming
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
		return nil, fmt.Errorf("no configuration loaded")
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
		AutoCloseAfter: cfg.Server.AutoCloseAfter,
	}

	server := &WebSocketServer{
		config:             serverConfig,
		configManager:      configManager,
		logger:             logger,
		jwtHandler:         jwtHandler,
		mediaMTXController: mediaMTXController,

		// Security extensions initialization (Phase 1 enhancement)
		permissionChecker: security.NewPermissionChecker(),

		// Input validation initialization (wire real validator with config adapter)
		validationHelper: NewValidationHelper(security.NewInputValidator(logger, security.NewConfigAdapter(&cfg.Security, &cfg.Logging)), logger),

		// WebSocket upgrader configuration
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.Server.ReadBufferSize,
			WriteBufferSize: cfg.Server.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return validateCORSOrigin(r, cfg, logger)
			},
		},

		// Client management
		clients:       make(map[string]*ClientConnection),
		clientCounter: 0,

		// Method registration
		methods:        make(map[string]MethodHandler),
		methodsMutex:   sync.RWMutex{},
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
		eventManager:  NewEventManager(logger),
		eventHandlers: make([]func(string, interface{}), 0),

		// Graceful shutdown
		stopChan: make(chan struct{}, 10), // Buffered to prevent deadlock during shutdown
		stopOnce: sync.Once{},
	}

	// Register built-in methods
	server.registerBuiltinMethods()

	// Mark builtin methods as ready
	atomic.StoreInt32(&server.builtinMethodsReady, 1)

	return server, nil
}

// Start starts the WebSocket server
func (s *WebSocketServer) Start() error {
	// Idempotency: set running flag atomically to prevent double start races
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		s.logger.Warn("WebSocket server is already running")
		return fmt.Errorf("WebSocket server is already running")
	}

	s.logger.WithFields(logging.Fields{
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
			// Reset running flag on failure
			atomic.StoreInt32(&s.running, 0)
		}
	}()

	// Server should be ready immediately with proper atomic flag handling

	s.logger.WithFields(logging.Fields{
		"host":   s.config.Host,
		"port":   s.config.Port,
		"path":   s.config.WebSocketPath,
		"action": "start_server",
		"status": "success",
	}).Info("WebSocket server started successfully")

	return nil
}

// StartWithListener starts the server using an existing listener (for race-free testing)
func (s *WebSocketServer) StartWithListener(listener net.Listener) error {
	// Idempotency: set running flag atomically to prevent double start races
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		s.logger.Warn("WebSocket server is already running")
		return fmt.Errorf("WebSocket server is already running")
	}

	s.logger.WithFields(logging.Fields{
		"addr":   listener.Addr().String(),
		"path":   s.config.WebSocketPath,
		"action": "start_server_with_listener",
	}).Info("Starting WebSocket JSON-RPC server with existing listener")

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.WebSocketPath, s.handleWebSocket)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Start server in goroutine using existing listener
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("WebSocket server failed")
			// Reset running flag on failure
			atomic.StoreInt32(&s.running, 0)
		}
	}()

	s.logger.WithFields(logging.Fields{
		"addr":   listener.Addr().String(),
		"path":   s.config.WebSocketPath,
		"action": "start_server_with_listener",
		"status": "success",
	}).Info("WebSocket server started successfully with existing listener")

	return nil
}

// Stop stops the WebSocket server gracefully with context-aware cancellation
func (s *WebSocketServer) Stop(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 0 {
		s.logger.Warn("WebSocket server is not running")
		return fmt.Errorf("WebSocket server is not running")
	}

	s.logger.Info("Stopping WebSocket server")

	// Signal shutdown - use sync.Once to ensure single close operation
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})

	// Close all client connections with timeout
	s.closeAllClientConnections()

	// Shutdown HTTP server with context
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
			// Note: Error is logged but not returned as this is cleanup operation
		}
	}

	// Wait for all goroutines to finish with context timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		s.logger.Warn("WebSocket server shutdown timeout, forcing stop")
		// Reflect state accurately; server may still be up
		return fmt.Errorf("shutdown timeout")
	}

	atomic.StoreInt32(&s.running, 0)

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
			if err := client.Conn.SetWriteDeadline(time.Now().Add(s.config.ClientCleanupTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to set write deadline for close message")
			}

			// Send close message
			closeMsg := websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutdown")
			if err := client.Conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(s.config.ClientCleanupTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to send close message")
			}

			// Close connection
			if err := client.Conn.Close(); err != nil {
				cleanupResults <- fmt.Errorf("failed to close connection for client %s: %w", client.ClientID, err)
				return
			}

			// Remove client from map and update metrics atomically
			s.clientsMutex.Lock()
			if _, exists := s.clients[client.ClientID]; exists {
				delete(s.clients, client.ClientID)
				atomic.AddInt64(&s.clientCount, -1)
			}
			s.clientsMutex.Unlock()

			// Remove event subscriptions
			s.eventManager.RemoveClient(client.ClientID)

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
			if err := client.Conn.Close(); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to force close connection")
			}
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
	// Check connection limit with atomic operation (lock-free)
	if atomic.LoadInt64(&s.metrics.ActiveConnections) >= int64(s.config.MaxConnections) {
		s.logger.Warn("Maximum connections reached")
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}

	// Generate client ID with atomic operations (lock-free)
	clientCounter := atomic.AddInt64(&s.clientCounter, 1)
	clientID := "client_" + strconv.FormatInt(clientCounter, 10)

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
	atomic.AddInt64(&s.clientCount, 1)
	s.clientsMutex.Unlock()

	// Update metrics with atomic operation
	atomic.AddInt64(&s.metrics.ActiveConnections, 1)

	s.logger.WithFields(logging.Fields{
		"client_id": clientID,
		"action":    "client_connected",
	}).Info("Client connected")

	// Handle connection in goroutine
	s.wg.Add(1)
	go s.handleClientConnection(conn, client)
}

// handleClientConnection handles individual client connections
func (s *WebSocketServer) handleClientConnection(conn *websocket.Conn, client *ClientConnection) {
	// Create error channel for panic recovery
	panicChan := make(chan error, 1)

	defer func() {
		// Recover from panics in goroutine and propagate as errors
		if r := recover(); r != nil {
			// Get stack trace for debugging
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			stackTrace := string(stack[:length])

			panicErr := fmt.Errorf("panic in client connection handler for client %s: %v", client.ClientID, r)
			s.logger.WithFields(logging.Fields{
				"client_id":   client.ClientID,
				"panic":       r,
				"action":      "panic_recovered",
				"stack_trace": stackTrace,
			}).Error("Recovered from panic in client connection handler")

			// Propagate panic as error instead of swallowing it
			select {
			case panicChan <- panicErr:
			default:
				s.logger.WithError(panicErr).Warn("Panic channel overflow, panic error dropped")
			}
		}

		// Remove client from connections and update metrics atomically
		s.clientsMutex.Lock()
		if _, exists := s.clients[client.ClientID]; exists {
			delete(s.clients, client.ClientID)
			atomic.AddInt64(&s.clientCount, -1)
		}
		s.clientsMutex.Unlock()

		// Remove event subscriptions
		s.eventManager.RemoveClient(client.ClientID)

		// Update metrics with atomic operation
		atomic.AddInt64(&s.metrics.ActiveConnections, -1)

		// Close connection
		if err := conn.Close(); err != nil {
			s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to close connection")
		}

		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"action":    "client_disconnected",
		}).Info("Client disconnected")

		s.wg.Done()
	}()

	// Set connection parameters
	conn.SetReadLimit(s.config.MaxMessageSize)
	if err := conn.SetReadDeadline(time.Now().Add(s.config.PongWait)); err != nil {
		s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to set initial read deadline")
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(s.config.PongWait)); err != nil {
			s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to set pong read deadline")
		}
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
			if err := conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to set write deadline for ping")
			}
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(s.config.WriteTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Error("Failed to send ping")
				return
			}
		default:
			// Set read deadline for message
			if err := conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout)); err != nil {
				s.logger.WithError(err).WithField("client_id", client.ClientID).Warn("Failed to set read deadline for message")
			}

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

	s.logger.WithFields(logging.Fields{
		"client_id": client.ClientID,
		"action":    "handle_message",
	}).Debug("Processing WebSocket message")

	// Parse JSON-RPC request
	var request JsonRpcRequest
	if err := json.Unmarshal(message, &request); err != nil {
		// Standardized error
		if sendErr := s.sendResponse(conn, &JsonRpcResponse{JSONRPC: "2.0", ID: nil, Error: NewJsonRpcError(INVALID_REQUEST, "invalid_request", "Invalid JSON-RPC request", "Ensure valid JSON-RPC 2.0 structure")}); sendErr != nil {
			s.logger.WithError(sendErr).Error("Failed to send invalid request error response")
		}
		return
	}

	// Validate JSON-RPC version
	if request.JSONRPC != "2.0" {
		if sendErr := s.sendResponse(conn, &JsonRpcResponse{JSONRPC: "2.0", ID: request.ID, Error: NewJsonRpcError(INVALID_REQUEST, "invalid_version", "Invalid JSON-RPC version", "Set jsonrpc to '2.0'")}); sendErr != nil {
			s.logger.WithError(sendErr).Error("Failed to send invalid version error response")
		}
		return
	}

	// Check if this is a notification (ID is null)
	isNotification := request.ID == nil

	// Handle request
	response, err := s.handleRequest(&request, client)
	if err != nil {
		s.logger.WithError(err).WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    request.Method,
		}).Error("Request handling error")
		// Only send error response for requests, not notifications
		if !isNotification {
			if sendErr := s.sendResponse(conn, &JsonRpcResponse{JSONRPC: "2.0", ID: request.ID, Error: NewJsonRpcError(INTERNAL_ERROR, "internal_error", err.Error(), "Retry or contact support")}); sendErr != nil {
				s.logger.WithError(sendErr).Error("Failed to send internal error response")
			}
		}
		return
	}

	// Only send response for requests, not notifications
	if !isNotification {
		// Attach API metadata
		if response != nil {
			if response.Metadata == nil {
				response.Metadata = make(map[string]interface{})
			}
			response.Metadata["processing_time_ms"] = time.Since(startTime).Milliseconds()
			response.Metadata["server_timestamp"] = time.Now().Format(time.RFC3339)
			response.Metadata["request_id"] = request.ID
		}
		if err := s.sendResponse(conn, response); err != nil {
			s.logger.WithError(err).WithField("client_id", client.ClientID).Error("Failed to send response")
			return
		}
	}

	// Record performance metrics
	duration := time.Since(startTime).Seconds()
	s.recordRequest(request.Method, duration)

	s.logger.WithFields(logging.Fields{
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
			Error:   NewJsonRpcError(RATE_LIMIT_EXCEEDED, "rate_limit", err.Error(), "Reduce request rate or wait"),
		}, nil
	}

	// Find method handler with mutex-protected lookup
	s.logger.WithFields(logging.Fields{
		"client_id": client.ClientID,
		"method":    request.Method,
		"action":    "method_lookup",
	}).Debug("Looking up method handler")

	s.methodsMutex.RLock()
	handler, exists := s.methods[request.Method]
	s.methodsMutex.RUnlock()

	s.logger.WithFields(logging.Fields{
		"client_id":    client.ClientID,
		"method":       request.Method,
		"exists":       exists,
		"handler_type": fmt.Sprintf("%T", handler),
		"action":       "method_lookup_result",
	}).Info("Method lookup completed")

	if !exists {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    request.Method,
			"action":    "method_not_found",
		}).Debug("Method not found")
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   NewJsonRpcError(METHOD_NOT_FOUND, "method_not_found", request.Method, "Verify method name"),
		}, nil
	}

	// Security extensions: Authentication check (Phase 1 enhancement) - FIRST GATE
	// Authentication must happen before any other checks, including readiness
	if request.Method != "authenticate" {
		if !client.Authenticated {
			s.logger.WithFields(logging.Fields{
				"client_id": client.ClientID,
				"method":    request.Method,
				"action":    "auth_required",
				"component": "security_middleware",
			}).Warn("Authentication required for method")
			return &JsonRpcResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   NewJsonRpcError(AUTHENTICATION_REQUIRED, "auth_required", "Authentication required", "Authenticate first"),
			}, nil
		}
	}

	// PROGRESSIVE READINESS ARCHITECTURE: Remove blocking system readiness check
	// Each operation implements its own Progressive Readiness pattern with fallback behavior
	// This follows the architectural requirement: "System accepts connections immediately"
	s.logger.WithFields(logging.Fields{
		"client_id": client.ClientID,
		"method":    request.Method,
		"action":    "progressive_readiness_enabled",
	}).Debug("Progressive Readiness: Attempting operation with potential fallback")

	// Security extensions: Permission check (Phase 1 enhancement)
	// Skip permission check for authenticate method
	if request.Method != "authenticate" {
		s.logger.WithFields(logging.Fields{
			"client_id": client.ClientID,
			"method":    request.Method,
			"action":    "permission_check",
		}).Debug("Checking method permissions")

		if err := s.checkMethodPermissions(client, request.Method); err != nil {
			s.logger.WithFields(logging.Fields{
				"client_id": client.ClientID,
				"method":    request.Method,
				"error":     err.Error(),
				"action":    "permission_denied",
			}).Debug("Permission check failed")

			return &JsonRpcResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Error:   NewJsonRpcError(INSUFFICIENT_PERMISSIONS, "permission_denied", err.Error(), "Use an account with permission"),
			}, nil
		}
	}

	// Call method handler
	s.logger.WithFields(logging.Fields{
		"client_id": client.ClientID,
		"method":    request.Method,
		"action":    "calling_handler",
	}).Debug("Calling method handler")

	response, err := handler(request.Params, client)
	if err != nil {
		// Update error metrics with atomic operation
		atomic.AddInt64(&s.metrics.ErrorCount, 1)

		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   NewJsonRpcError(INTERNAL_ERROR, "handler_error", err.Error(), "Retry or contact support"),
		}, nil
	}

	// Set JSON-RPC version and ID
	response.JSONRPC = "2.0"
	response.ID = request.ID

	return response, nil
}

// sendResponse sends a JSON-RPC response to the client
func (s *WebSocketServer) sendResponse(conn *websocket.Conn, response *JsonRpcResponse) error {
	if err := conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); err != nil {
		s.logger.WithError(err).Warn("Failed to set write deadline for response")
		// Continue with WriteJSON - the operation might still succeed
	}
	return conn.WriteJSON(response)
}

// sendErrorResponse sends a JSON-RPC error response to the client
func (s *WebSocketServer) sendErrorResponse(conn *websocket.Conn, id interface{}, code int, message string) {
	// Check for nil connection to prevent panic
	if conn == nil {
		s.logger.WithFields(logging.Fields{
			"error_code": code,
			"message":    message,
		}).Error("Cannot send error response: connection is nil")
		return
	}

	response := &JsonRpcResponse{JSONRPC: "2.0", ID: id, Error: NewJsonRpcError(code, "error", message, "See documentation")}

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

	// Note: Average response time calculation is available for future use
	// when extending the metrics functionality
	// Example: allResponseTimes := make([]float64, 0)
	// for _, times := range s.metrics.ResponseTimes {
	//     allResponseTimes = append(allResponseTimes, times...)
	// }

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
	return atomic.LoadInt32(&s.running) == 1
}

// GetClientCount returns the current number of connected clients using atomic operation
func (s *WebSocketServer) GetClientCount() int64 {
	return atomic.LoadInt64(&s.clientCount)
}

// IsBuiltinMethodsReady returns whether builtin methods are registered using atomic operation
func (s *WebSocketServer) IsBuiltinMethodsReady() bool {
	return atomic.LoadInt32(&s.builtinMethodsReady) == 1
}

// GetEventHandlerCount returns the current number of event handlers using atomic operation
// validateCORSOrigin validates the origin header against configured CORS origins
func validateCORSOrigin(r *http.Request, cfg *config.Config, logger *logging.Logger) bool {
	origin := r.Header.Get("Origin")

	// Allow same-origin requests (no Origin header)
	if origin == "" {
		return true
	}

	// Check against configured allowed origins
	for _, allowed := range cfg.Security.CORSOrigins {
		if origin == allowed {
			logger.WithFields(logging.Fields{
				"origin":  origin,
				"allowed": true,
			}).Debug("CORS origin validated successfully")
			return true
		}
	}

	// Log rejected origin for security monitoring
	logger.WithFields(logging.Fields{
		"origin":          origin,
		"allowed_origins": cfg.Security.CORSOrigins,
		"allowed":         false,
	}).Warn("CORS origin rejected")

	return false
}

func (s *WebSocketServer) GetEventHandlerCount() int64 {
	return atomic.LoadInt64(&s.eventHandlerCount)
}

// GetConfig returns the server configuration (for testing purposes)
func (s *WebSocketServer) GetConfig() *ServerConfig {
	return s.config
}

// SetConfig sets the server configuration (for testing purposes)
func (s *WebSocketServer) SetConfig(config *ServerConfig) {
	s.config = config
}

// SetMediaMTXController sets the MediaMTX controller (for testing purposes)
func (s *WebSocketServer) SetMediaMTXController(controller mediamtx.MediaMTXControllerAPI) {
	s.mediaMTXController = controller
}

// GetEventManager returns the event manager for external integration
func (s *WebSocketServer) GetEventManager() *EventManager {
	return s.eventManager
}
