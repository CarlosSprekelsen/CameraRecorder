package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// HTTPHealthServer implements HTTP health endpoints with thin delegation pattern.
type HTTPHealthServer struct {
	config    *config.HTTPHealthConfig
	logger    *logging.Logger
	healthAPI HealthAPI
	server    *http.Server
	startTime time.Time
}

// NewHTTPHealthServer creates a new HTTP health server instance
func NewHTTPHealthServer(config *config.HTTPHealthConfig, healthAPI HealthAPI, logger *logging.Logger) (*HTTPHealthServer, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	if healthAPI == nil {
		return nil, fmt.Errorf("health API cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	server := &HTTPHealthServer{
		config:    config,
		logger:    logger,
		healthAPI: healthAPI,
		startTime: time.Now(),
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Register health endpoints
	mux.HandleFunc(config.BasicEndpoint, server.handleBasicHealth)
	mux.HandleFunc(config.DetailedEndpoint, server.handleDetailedHealth)
	mux.HandleFunc(config.ReadyEndpoint, server.handleReadiness)
	mux.HandleFunc(config.LiveEndpoint, server.handleLiveness)

	// Parse timeouts
	readTimeout, err := time.ParseDuration(config.ReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid read timeout: %w", err)
	}

	writeTimeout, err := time.ParseDuration(config.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid write timeout: %w", err)
	}

	idleTimeout, err := time.ParseDuration(config.IdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid idle timeout: %w", err)
	}

	server.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	server.logger.WithFields(logging.Fields{
		"host":    config.Host,
		"port":    config.Port,
		"enabled": config.Enabled,
	}).Info("HTTP Health Server initialized")

	return server, nil
}

// Start starts the HTTP health server
func (hs *HTTPHealthServer) Start(ctx context.Context) error {
	if !hs.config.Enabled {
		hs.logger.Info("HTTP Health Server is disabled")
		return nil
	}

	hs.logger.WithFields(logging.Fields{
		"address": hs.server.Addr,
		"endpoints": []string{
			hs.config.BasicEndpoint,
			hs.config.DetailedEndpoint,
			hs.config.ReadyEndpoint,
			hs.config.LiveEndpoint,
		},
	}).Info("Starting HTTP Health Server")

	// Start server in goroutine
	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.logger.WithError(err).Error("HTTP Health Server failed to start")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hs.server.Shutdown(shutdownCtx); err != nil {
		hs.logger.WithError(err).Error("HTTP Health Server shutdown failed")
		return err
	}

	hs.logger.Info("HTTP Health Server stopped")
	return nil
}

// Stop stops the HTTP health server
func (hs *HTTPHealthServer) Stop() error {
	if hs.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return hs.server.Shutdown(ctx)
}

// handleBasicHealth handles the basic health endpoint
func (hs *HTTPHealthServer) handleBasicHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Delegate to HealthAPI - NO business logic in HTTP server
	response, err := hs.healthAPI.GetHealth(r.Context())
	if err != nil {
		hs.logger.WithError(err).Error("Failed to get health status")
		hs.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set response headers
	hs.setResponseHeaders(w)

	// Write response
	hs.writeJSONResponse(w, http.StatusOK, response)

	// Log request
	hs.logRequest(r, "basic_health", time.Since(start), http.StatusOK)
}

// handleDetailedHealth handles the detailed health endpoint
func (hs *HTTPHealthServer) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Delegate to HealthAPI - NO business logic in HTTP server
	response, err := hs.healthAPI.GetDetailedHealth(r.Context())
	if err != nil {
		hs.logger.WithError(err).Error("Failed to get detailed health status")
		hs.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set response headers
	hs.setResponseHeaders(w)

	// Write response
	hs.writeJSONResponse(w, http.StatusOK, response)

	// Log request
	hs.logRequest(r, "detailed_health", time.Since(start), http.StatusOK)
}

// handleReadiness handles the readiness probe endpoint
func (hs *HTTPHealthServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Delegate to HealthAPI - NO business logic in HTTP server
	response, err := hs.healthAPI.IsReady(r.Context())
	if err != nil {
		hs.logger.WithError(err).Error("Failed to check readiness")
		hs.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set response headers
	hs.setResponseHeaders(w)

	// Determine HTTP status code
	statusCode := http.StatusOK
	if !response.Ready {
		statusCode = http.StatusServiceUnavailable
	}

	// Write response
	hs.writeJSONResponse(w, statusCode, response)

	// Log request
	hs.logRequest(r, "readiness", time.Since(start), statusCode)
}

// handleLiveness handles the liveness probe endpoint
func (hs *HTTPHealthServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Delegate to HealthAPI - NO business logic in HTTP server
	response, err := hs.healthAPI.IsAlive(r.Context())
	if err != nil {
		hs.logger.WithError(err).Error("Failed to check liveness")
		hs.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set response headers
	hs.setResponseHeaders(w)

	// Determine HTTP status code
	statusCode := http.StatusOK
	if !response.Alive {
		statusCode = http.StatusServiceUnavailable
	}

	// Write response
	hs.writeJSONResponse(w, statusCode, response)

	// Log request
	hs.logRequest(r, "liveness", time.Since(start), statusCode)
}

// setResponseHeaders sets common response headers
func (hs *HTTPHealthServer) setResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Add CORS headers if needed
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// writeJSONResponse writes a JSON response
func (hs *HTTPHealthServer) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		hs.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes an error response
func (hs *HTTPHealthServer) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	// Set response headers first
	hs.setResponseHeaders(w)

	errorResponse := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    statusCode,
	}

	hs.writeJSONResponse(w, statusCode, errorResponse)
}

// logRequest logs HTTP request details
func (hs *HTTPHealthServer) logRequest(r *http.Request, endpoint string, duration time.Duration, statusCode int) {
	hs.logger.WithFields(logging.Fields{
		"method":      r.Method,
		"endpoint":    endpoint,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"duration":    duration.String(),
		"status_code": statusCode,
	}).Debug("HTTP health request processed")
}

// GetServerInfo returns information about the HTTP health server
func (hs *HTTPHealthServer) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    hs.config.Enabled,
		"host":       hs.config.Host,
		"port":       hs.config.Port,
		"start_time": hs.startTime,
		"uptime":     time.Since(hs.startTime).String(),
		"endpoints": []string{
			hs.config.BasicEndpoint,
			hs.config.DetailedEndpoint,
			hs.config.ReadyEndpoint,
			hs.config.LiveEndpoint,
		},
	}
}
