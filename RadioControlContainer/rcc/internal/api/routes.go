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
	"time"
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

    // Fetch radios from RadioManager
    if s.radioManager == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
            "Radio manager not available", nil)
        return
    }

    list := s.radioManager.List()
    WriteSuccess(w, list)
}

// handleSelectRadio handles POST /radios/select
// Source: OpenAPI v1 §3.3
func (s *Server) handleSelectRadio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only POST method is allowed", nil)
		return
	}

    // Parse request
    var req struct {
        ID string `json:"id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        WriteError(w, http.StatusBadRequest, "INVALID_RANGE", "Missing or invalid id", nil)
        return
    }

    // Validate radio exists and select via RadioManager
    if s.radioManager == nil || s.orchestrator == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
        return
    }

    if err := s.radioManager.SetActive(req.ID); err != nil {
        WriteError(w, http.StatusNotFound, "NOT_FOUND", "Radio not found", nil)
        return
    }

    // Call orchestrator to confirm selection (ping adapter/state)
    if err := s.orchestrator.SelectRadio(r.Context(), req.ID); err != nil {
        status, body := ToAPIError(err)
        w.WriteHeader(status)
        w.Write(body)
        return
    }

    WriteSuccess(w, map[string]string{"activeRadioId": req.ID})
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

    if s.radioManager == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
            "Radio manager not available", nil)
        return
    }

    radio, ok := s.radioManager.GetRadio(radioID)
    if !ok {
        WriteError(w, http.StatusNotFound, "NOT_FOUND", "Radio not found", nil)
        return
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
    if s.orchestrator == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
        return
    }
    state, err := s.orchestrator.GetState(r.Context(), radioID)
    if err != nil {
        status, body := ToAPIError(err)
        w.WriteHeader(status)
        w.Write(body)
        return
    }
    WriteSuccess(w, map[string]interface{}{"powerDbm": state.PowerDbm})
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

    if s.orchestrator == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
        return
    }
    if err := s.orchestrator.SetPower(r.Context(), radioID, request.PowerDbm); err != nil {
        status, body := ToAPIError(err)
        w.WriteHeader(status)
        w.Write(body)
        return
    }
    WriteSuccess(w, map[string]interface{}{"powerDbm": request.PowerDbm})
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
    if s.orchestrator == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
        return
    }
    state, err := s.orchestrator.GetState(r.Context(), radioID)
    if err != nil {
        status, body := ToAPIError(err)
        w.WriteHeader(status)
        w.Write(body)
        return
    }
    // channelIndex may be null if not in derived set; we return frequency
    WriteSuccess(w, map[string]interface{}{"frequencyMhz": state.FrequencyMhz, "channelIndex": nil})
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

    if s.orchestrator == nil {
        WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
        return
    }

    // Frequency wins if both provided
    if request.FrequencyMhz != nil {
        if err := s.orchestrator.SetChannel(r.Context(), radioID, *request.FrequencyMhz); err != nil {
            status, body := ToAPIError(err)
            w.WriteHeader(status)
            w.Write(body)
            return
        }
        WriteSuccess(w, map[string]interface{}{"frequencyMhz": *request.FrequencyMhz, "channelIndex": request.ChannelIndex})
        return
    }

    // If only index provided, translate via radioManager channels (if available)
    if request.ChannelIndex != nil {
        if s.radioManager == nil {
            WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Radio manager not available", nil)
            return
        }
        radio, ok := s.radioManager.GetRadio(radioID)
        if !ok {
            WriteError(w, http.StatusNotFound, "NOT_FOUND", "Radio not found", nil)
            return
        }
        // Find frequency for index
        var freq float64
        found := false
        for _, ch := range radio.Capabilities.Channels {
            if ch.Index == *request.ChannelIndex {
                freq = ch.FrequencyMhz
                found = true
                break
            }
        }
        if !found {
            WriteError(w, http.StatusBadRequest, "INVALID_RANGE", "Invalid channelIndex", nil)
            return
        }
        if err := s.orchestrator.SetChannel(r.Context(), radioID, freq); err != nil {
            status, body := ToAPIError(err)
            w.WriteHeader(status)
            w.Write(body)
            return
        }
        WriteSuccess(w, map[string]interface{}{"frequencyMhz": freq, "channelIndex": *request.ChannelIndex})
        return
    }
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
    uptime := 0.0
    if !s.startTime.IsZero() {
        uptime = time.Since(s.startTime).Seconds()
    }
    health := map[string]interface{}{
        "status":    "ok",
        "uptimeSec": uptime,
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
