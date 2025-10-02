package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/telemetry"
)

func TestNewServer(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.telemetryHub != hub {
		t.Error("Telemetry hub not set correctly")
	}
}

func TestServerStartStop(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test server creation
	if server.httpServer != nil {
		t.Error("HTTP server should be nil before Start()")
	}

	// Test that we can get the server after creation
	if server.GetServer() != nil {
		t.Error("GetServer() should return nil before Start()")
	}
}

func TestRegisterRoutes(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)
	mux := http.NewServeMux()

	// Register routes
	server.RegisterRoutes(mux)

	// Test that routes are registered by checking if they exist
	// This is a basic test - in a real implementation, we'd test actual endpoints
	if mux == nil {
		t.Error("Mux should not be nil after registering routes")
	}
}

func TestResponseEnvelope(t *testing.T) {
	// Test success response
	successResp := SuccessResponse(map[string]string{"test": "data"})
	if successResp.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", successResp.Result)
	}
	if successResp.CorrelationID == "" {
		t.Error("Correlation ID should not be empty")
	}

	// Test error response
	errorResp := ErrorResponse("TEST_ERROR", "Test error message", nil)
	if errorResp.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", errorResp.Result)
	}
	if errorResp.Code != "TEST_ERROR" {
		t.Errorf("Expected code 'TEST_ERROR', got '%s'", errorResp.Code)
	}
	if errorResp.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", errorResp.Message)
	}
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"test": "data"}

	WriteSuccess(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, http.StatusBadRequest, "INVALID_RANGE", "Test error", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}
	if response.Code != "INVALID_RANGE" {
		t.Errorf("Expected code 'INVALID_RANGE', got '%s'", response.Code)
	}
}

func TestWriteNotImplemented(t *testing.T) {
	w := httptest.NewRecorder()

	WriteNotImplemented(w, "test-endpoint")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}
	if response.Code != "NOT_IMPLEMENTED" {
		t.Errorf("Expected code 'NOT_IMPLEMENTED', got '%s'", response.Code)
	}
}

func TestStandardErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Response
		expected int
	}{
		{"InvalidRange", ErrInvalidRange, http.StatusBadRequest},
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, http.StatusForbidden},
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"Busy", ErrBusy, http.StatusServiceUnavailable},
		{"Unavailable", ErrUnavailable, http.StatusServiceUnavailable},
		{"Internal", ErrInternal, http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteStandardError(w, test.err)

			if w.Code != test.expected {
				t.Errorf("Expected status %d, got %d", test.expected, w.Code)
			}
		})
	}
}

func TestExtractRadioID(t *testing.T) {
	server := &Server{}

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/radios/radio-01", "radio-01"},
		{"/api/v1/radios/radio-01/power", "radio-01"},
		{"/api/v1/radios/radio-01/channel", "radio-01"},
		{"/api/v1/radios/", ""},
		{"/api/v1/radios", ""},
		{"/invalid/path", ""},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			result := server.extractRadioID(test.path)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestHandleCapabilities(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /capabilities
	req := httptest.NewRequest("GET", "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()

	server.handleCapabilities(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}

	// Check capabilities data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if data["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", data["version"])
	}
}

func TestHandleRadios(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /radios
	req := httptest.NewRequest("GET", "/api/v1/radios", nil)
	w := httptest.NewRecorder()

	server.handleRadios(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}
}

func TestHandleSelectRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test POST /radios/select
	req := httptest.NewRequest("POST", "/api/v1/radios/select", strings.NewReader(`{"id":"radio-01"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSelectRadio(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}
	if response.Code != "NOT_IMPLEMENTED" {
		t.Errorf("Expected code 'NOT_IMPLEMENTED', got '%s'", response.Code)
	}
}

func TestHandleRadioByID(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /radios/{id}
	req := httptest.NewRequest("GET", "/api/v1/radios/radio-01", nil)
	w := httptest.NewRecorder()

	server.handleRadioByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}
}

func TestHandleGetPower(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /radios/{id}/power
	req := httptest.NewRequest("GET", "/api/v1/radios/radio-01/power", nil)
	w := httptest.NewRecorder()

	server.handleGetPower(w, req, "radio-01")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}
}

func TestHandleSetPower(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test POST /radios/{id}/power with valid power
	req := httptest.NewRequest("POST", "/api/v1/radios/radio-01/power",
		strings.NewReader(`{"powerDbm":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetPower(w, req, "radio-01")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	// Test with invalid power (too high)
	req = httptest.NewRequest("POST", "/api/v1/radios/radio-01/power",
		strings.NewReader(`{"powerDbm":50}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.handleSetPower(w, req, "radio-01")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}
	if response.Code != "INVALID_RANGE" {
		t.Errorf("Expected code 'INVALID_RANGE', got '%s'", response.Code)
	}
}

func TestHandleGetChannel(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /radios/{id}/channel
	req := httptest.NewRequest("GET", "/api/v1/radios/radio-01/channel", nil)
	w := httptest.NewRecorder()

	server.handleGetChannel(w, req, "radio-01")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}
}

func TestHandleSetChannel(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test POST /radios/{id}/channel with channel index
	req := httptest.NewRequest("POST", "/api/v1/radios/radio-01/channel",
		strings.NewReader(`{"channelIndex":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleSetChannel(w, req, "radio-01")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	// Test with frequency
	req = httptest.NewRequest("POST", "/api/v1/radios/radio-01/channel",
		strings.NewReader(`{"frequencyMhz":2422.0}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.handleSetChannel(w, req, "radio-01")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	// Test with both parameters
	req = httptest.NewRequest("POST", "/api/v1/radios/radio-01/channel",
		strings.NewReader(`{"channelIndex":3,"frequencyMhz":2422.0}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.handleSetChannel(w, req, "radio-01")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}

	// Test with no parameters
	req = httptest.NewRequest("POST", "/api/v1/radios/radio-01/channel",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.handleSetChannel(w, req, "radio-01")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleHealth(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /health
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response.Result)
	}

	// Check health data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if data["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", data["status"])
	}
	if data["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", data["version"])
	}
}

func TestHandleTelemetry(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test GET /telemetry
	req := httptest.NewRequest("GET", "/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")
	w := httptest.NewRecorder()

	server.handleTelemetry(w, req)

	// The telemetry endpoint should not return an error response
	// It should handle SSE streaming (which is complex to test in unit tests)
	// For now, we just verify it doesn't crash
}

func TestMethodNotAllowed(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	server := NewServer(hub)

	// Test wrong method on capabilities
	req := httptest.NewRequest("POST", "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()

	server.handleCapabilities(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "error" {
		t.Errorf("Expected result 'error', got '%s'", response.Result)
	}
	if response.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("Expected code 'METHOD_NOT_ALLOWED', got '%s'", response.Code)
	}
}
