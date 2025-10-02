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

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// Server represents the HTTP API server.
type Server struct {
	httpServer   *http.Server
	telemetryHub *telemetry.Hub
	orchestrator *command.Orchestrator
	radioManager *radio.Manager
	startTime    time.Time
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

// NewServer creates a new API server.
func NewServer(telemetryHub *telemetry.Hub, orchestrator *command.Orchestrator, radioManager *radio.Manager, readTimeout, writeTimeout, idleTimeout time.Duration) *Server {
	return &Server{
		telemetryHub: telemetryHub,
		orchestrator: orchestrator,
		radioManager: radioManager,
		startTime:    time.Now(),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		idleTimeout:  idleTimeout,
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
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
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
