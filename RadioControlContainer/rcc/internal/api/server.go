// Package api implements ApiGateway from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Expose northbound HTTP/JSON commands and SSE endpoint; translate HTTP requests into orchestrator calls; throttle per client."
//
// Source: OpenAPI v1
// Quote: "Base URL: http://<edge-hub>/api/v1"
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/radio-control/rcc/internal/telemetry"
)

// Server represents the HTTP API server.
type Server struct {
	httpServer   *http.Server
	telemetryHub *telemetry.Hub
	// TODO: Add command orchestrator, radio manager, etc.
}

// NewServer creates a new API server.
func NewServer(telemetryHub *telemetry.Hub) *Server {
	return &Server{
		telemetryHub: telemetryHub,
	}
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// Register all routes
	s.RegisterRoutes(mux)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

// GetServer returns the underlying HTTP server for testing.
func (s *Server) GetServer() *http.Server {
	return s.httpServer
}
