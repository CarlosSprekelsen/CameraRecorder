// Package api implements ApiGateway from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Expose northbound HTTP/JSON commands and SSE endpoint; translate HTTP requests into orchestrator calls; throttle per client."
//
// Source: OpenAPI v1
// Quote: "Minimal, stable contract for selecting a radio, setting channel and power, and receiving telemetry."
package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterRoutes registers all OpenAPI v1 endpoints.
// Source: OpenAPI v1 §3
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// API v1 base path
	apiV1 := "/api/v1"

	// Capabilities endpoint
	mux.HandleFunc(apiV1+"/capabilities", s.handleCapabilities)

	// Radios endpoints
	mux.HandleFunc(apiV1+"/radios", s.handleRadios)
	mux.HandleFunc(apiV1+"/radios/select", s.handleSelectRadio)

	// Radio-specific endpoints (power, channel, individual radio)
	mux.HandleFunc(apiV1+"/radios/", s.handleRadioEndpoints)

	// Telemetry endpoint
	mux.HandleFunc(apiV1+"/telemetry", s.handleTelemetry)

	// Health endpoint
	mux.HandleFunc(apiV1+"/health", s.handleHealth)
}

// handleCapabilities handles GET /capabilities
// Source: OpenAPI v1 §3.1
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Return capabilities
	capabilities := map[string]interface{}{
		"telemetry": []string{"sse"},
		"commands":  []string{"http-json"},
		"version":   "1.0.0",
	}

	WriteSuccess(w, capabilities)
}

// handleRadios handles GET /radios
// Source: OpenAPI v1 §3.2
func (s *Server) handleRadios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// TODO: Get actual radios from radio manager
	// For now, return stub data
	radios := map[string]interface{}{
		"activeRadioId": "",
		"items":         []interface{}{},
	}

	WriteSuccess(w, radios)
}

// handleSelectRadio handles POST /radios/select
// Source: OpenAPI v1 §3.3
func (s *Server) handleSelectRadio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only POST method is allowed", nil)
		return
	}

	// TODO: Implement radio selection logic
	WriteNotImplemented(w, "POST /radios/select")
}

// handleRadioEndpoints handles all radio-specific endpoints.
// Routes to appropriate handler based on path.
func (s *Server) handleRadioEndpoints(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract radio ID and determine endpoint type
	radioID := s.extractRadioID(path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	// Route based on path suffix
	if strings.HasSuffix(path, "/power") {
		s.handleRadioPower(w, r)
	} else if strings.HasSuffix(path, "/channel") {
		s.handleRadioChannel(w, r)
	} else {
		// Default to individual radio endpoint
		s.handleRadioByID(w, r)
	}
}

// handleRadioByID handles GET /radios/{id}
// Source: OpenAPI v1 §3.4
func (s *Server) handleRadioByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	// TODO: Get actual radio data from radio manager
	// For now, return stub data
	radio := map[string]interface{}{
		"id":           radioID,
		"model":        "Unknown",
		"status":       "offline",
		"capabilities": map[string]interface{}{},
		"state":        map[string]interface{}{},
	}

	WriteSuccess(w, radio)
}

// handleRadioPower handles GET/POST /radios/{id}/power
// Source: OpenAPI v1 §3.5 & §3.6
func (s *Server) handleRadioPower(w http.ResponseWriter, r *http.Request) {
	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetPower(w, r, radioID)
	case http.MethodPost:
		s.handleSetPower(w, r, radioID)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET and POST methods are allowed", nil)
	}
}

// handleGetPower handles GET /radios/{id}/power
// Source: OpenAPI v1 §3.5
func (s *Server) handleGetPower(w http.ResponseWriter, r *http.Request, radioID string) {
	// TODO: Get actual power from radio
	// For now, return stub data
	power := map[string]interface{}{
		"powerDbm": 30,
	}

	WriteSuccess(w, power)
}

// handleSetPower handles POST /radios/{id}/power
// Source: OpenAPI v1 §3.6
func (s *Server) handleSetPower(w http.ResponseWriter, r *http.Request, radioID string) {
	// Parse request body
	var request struct {
		PowerDbm int `json:"powerDbm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Invalid JSON in request body", nil)
		return
	}

	// Validate power range
	if request.PowerDbm < 0 || request.PowerDbm > 39 {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Power must be between 0 and 39 dBm", nil)
		return
	}

	// TODO: Implement power setting logic
	WriteNotImplemented(w, "POST /radios/{id}/power")
}

// handleRadioChannel handles GET/POST /radios/{id}/channel
// Source: OpenAPI v1 §3.7 & §3.8
func (s *Server) handleRadioChannel(w http.ResponseWriter, r *http.Request) {
	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetChannel(w, r, radioID)
	case http.MethodPost:
		s.handleSetChannel(w, r, radioID)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET and POST methods are allowed", nil)
	}
}

// handleGetChannel handles GET /radios/{id}/channel
// Source: OpenAPI v1 §3.7
func (s *Server) handleGetChannel(w http.ResponseWriter, r *http.Request, radioID string) {
	// TODO: Get actual channel from radio
	// For now, return stub data
	channel := map[string]interface{}{
		"frequencyMhz": 2412.0,
		"channelIndex": nil, // May be null if frequency not in derived channel set
	}

	WriteSuccess(w, channel)
}

// handleSetChannel handles POST /radios/{id}/channel
// Source: OpenAPI v1 §3.8
func (s *Server) handleSetChannel(w http.ResponseWriter, r *http.Request, radioID string) {
	// Parse request body
	var request struct {
		ChannelIndex *int     `json:"channelIndex,omitempty"`
		FrequencyMhz *float64 `json:"frequencyMhz,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Invalid JSON in request body", nil)
		return
	}

	// Validate that at least one parameter is provided
	if request.ChannelIndex == nil && request.FrequencyMhz == nil {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Either channelIndex or frequencyMhz must be provided", nil)
		return
	}

	// TODO: Implement channel setting logic
	WriteNotImplemented(w, "POST /radios/{id}/channel")
}

// handleTelemetry handles GET /telemetry (SSE)
// Source: OpenAPI v1 §3.9
func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Wire to Telemetry Hub Subscribe
	if s.telemetryHub == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
			"Telemetry service not available", nil)
		return
	}

	// Subscribe to telemetry stream
	ctx := r.Context()
	if err := s.telemetryHub.Subscribe(ctx, w, r); err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL",
			"Failed to subscribe to telemetry stream", nil)
		return
	}
}

// handleHealth handles GET /health
// Source: OpenAPI v1 §3.10
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// TODO: Implement actual health checks
	// For now, return basic health status
	health := map[string]interface{}{
		"status":    "ok",
		"uptimeSec": 0, // TODO: Calculate actual uptime
		"version":   "1.0.0",
	}

	WriteSuccess(w, health)
}

// extractRadioID extracts the radio ID from a URL path.
// Handles paths like /api/v1/radios/{id}/power, /api/v1/radios/{id}/channel, etc.
func (s *Server) extractRadioID(path string) string {
	// Remove /api/v1/radios/ prefix
	prefix := "/api/v1/radios/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	// Get the part after the prefix
	remaining := path[len(prefix):]

	// Split by '/' to get the radio ID (first part)
	parts := strings.Split(remaining, "/")
	if len(parts) == 0 {
		return ""
	}

	radioID := parts[0]
	if radioID == "" {
		return ""
	}

	return radioID
}

// parseRadioIDFromPath is a helper to parse radio ID from various path patterns.
func parseRadioIDFromPath(path string) string {
	// Handle different path patterns:
	// /api/v1/radios/{id}
	// /api/v1/radios/{id}/power
	// /api/v1/radios/{id}/channel

	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "api" || parts[2] != "v1" || parts[3] != "radios" {
		return ""
	}

	if len(parts) < 5 {
		return ""
	}

	return parts[4]
}
