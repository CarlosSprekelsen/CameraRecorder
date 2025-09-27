/*
HTTP Health Server Unit Tests

Requirements Coverage:
- REQ-HEALTH-001: Health Monitoring
- REQ-HEALTH-002: HTTP Health Endpoints

Test Categories: Unit
API Documentation Reference: docs/api/health-endpoints.md

Unit tests for HTTP Health Server following existing testing patterns.
Tests thin delegation pattern, endpoint responses, and error handling.
*/

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHealthAPI is a mock implementation of HealthAPI for testing
type mockHealthAPI struct {
	healthResponse      *HealthResponse
	detailedResponse    *DetailedHealthResponse
	readinessResponse   *ReadinessResponse
	livenessResponse    *LivenessResponse
	healthError         error
	detailedError       error
	readinessError      error
	livenessError       error
}

func (m *mockHealthAPI) GetHealth(ctx context.Context) (*HealthResponse, error) {
	return m.healthResponse, m.healthError
}

func (m *mockHealthAPI) GetDetailedHealth(ctx context.Context) (*DetailedHealthResponse, error) {
	return m.detailedResponse, m.detailedError
}

func (m *mockHealthAPI) IsReady(ctx context.Context) (*ReadinessResponse, error) {
	return m.readinessResponse, m.readinessError
}

func (m *mockHealthAPI) IsAlive(ctx context.Context) (*LivenessResponse, error) {
	return m.livenessResponse, m.livenessError
}

func TestNewHTTPHealthServer(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.HTTPHealthConfig
		healthAPI   HealthAPI
		logger      *logging.Logger
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &config.HTTPHealthConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           8003,
				ReadTimeout:    "5s",
				WriteTimeout:   "5s",
				IdleTimeout:    "30s",
				BasicEndpoint:  "/health",
				DetailedEndpoint: "/health/detailed",
				ReadyEndpoint:  "/health/ready",
				LiveEndpoint:   "/health/live",
			},
			healthAPI:   &mockHealthAPI{},
			logger:      logging.GetLogger("test"),
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			healthAPI:   &mockHealthAPI{},
			logger:      logging.GetLogger("test"),
			expectError: true,
		},
		{
			name: "nil health API",
			config: &config.HTTPHealthConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           8003,
				ReadTimeout:    "5s",
				WriteTimeout:   "5s",
				IdleTimeout:    "30s",
				BasicEndpoint:  "/health",
				DetailedEndpoint: "/health/detailed",
				ReadyEndpoint:  "/health/ready",
				LiveEndpoint:   "/health/live",
			},
			healthAPI:   nil,
			logger:      logging.GetLogger("test"),
			expectError: true,
		},
		{
			name: "nil logger",
			config: &config.HTTPHealthConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           8003,
				ReadTimeout:    "5s",
				WriteTimeout:   "5s",
				IdleTimeout:    "30s",
				BasicEndpoint:  "/health",
				DetailedEndpoint: "/health/detailed",
				ReadyEndpoint:  "/health/ready",
				LiveEndpoint:   "/health/live",
			},
			healthAPI:   &mockHealthAPI{},
			logger:      nil,
			expectError: true,
		},
		{
			name: "invalid timeout",
			config: &config.HTTPHealthConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           8003,
				ReadTimeout:    "invalid",
				WriteTimeout:   "5s",
				IdleTimeout:    "30s",
				BasicEndpoint:  "/health",
				DetailedEndpoint: "/health/detailed",
				ReadyEndpoint:  "/health/ready",
				LiveEndpoint:   "/health/live",
			},
			healthAPI:   &mockHealthAPI{},
			logger:      logging.GetLogger("test"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewHTTPHealthServer(tt.config, tt.healthAPI, tt.logger)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Equal(t, tt.config, server.config)
				assert.Equal(t, tt.healthAPI, server.healthAPI)
				assert.Equal(t, tt.logger, server.logger)
			}
		})
	}
}

func TestHTTPHealthServer_handleBasicHealth(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	
	tests := []struct {
		name           string
		mockResponse   *HealthResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful health check",
			mockResponse: &HealthResponse{
				Status:    HealthStatusHealthy,
				Timestamp: time.Now(),
				Version:   "1.0.0",
				Uptime:    "1h30m",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "healthy",
				"version": "1.0.0",
				"uptime": "1h30m",
			},
		},
		{
			name:           "health API error",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock health API
			mockAPI := &mockHealthAPI{
				healthResponse: tt.mockResponse,
				healthError:    tt.mockError,
			}
			
			// Create server
			server, err := NewHTTPHealthServer(config, mockAPI, logger)
			require.NoError(t, err)
			
			// Create request
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			
			// Call handler
			server.handleBasicHealth(w, req)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			// Check expected fields
			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestHTTPHealthServer_handleDetailedHealth(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	
	tests := []struct {
		name           string
		mockResponse   *DetailedHealthResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful detailed health check",
			mockResponse: &DetailedHealthResponse{
				Status:    HealthStatusHealthy,
				Timestamp: time.Now(),
				Version:   "1.0.0",
				Uptime:    "1h30m",
				Components: []ComponentStatus{
					{
						Name:        "mediamtx",
						Status:      HealthStatusHealthy,
						LastChecked: time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "healthy",
				"version": "1.0.0",
				"uptime": "1h30m",
			},
		},
		{
			name:           "detailed health API error",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock health API
			mockAPI := &mockHealthAPI{
				detailedResponse: tt.mockResponse,
				detailedError:    tt.mockError,
			}
			
			// Create server
			server, err := NewHTTPHealthServer(config, mockAPI, logger)
			require.NoError(t, err)
			
			// Create request
			req := httptest.NewRequest("GET", "/health/detailed", nil)
			w := httptest.NewRecorder()
			
			// Call handler
			server.handleDetailedHealth(w, req)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			// Check expected fields
			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestHTTPHealthServer_handleReadiness(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	
	tests := []struct {
		name           string
		mockResponse   *ReadinessResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "system ready",
			mockResponse: &ReadinessResponse{
				Ready:     true,
				Timestamp: time.Now(),
				Message:   "System is ready",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"ready": true,
				"message": "System is ready",
			},
		},
		{
			name: "system not ready",
			mockResponse: &ReadinessResponse{
				Ready:     false,
				Timestamp: time.Now(),
				Message:   "System not ready",
			},
			mockError:      nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"ready": false,
				"message": "System not ready",
			},
		},
		{
			name:           "readiness API error",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock health API
			mockAPI := &mockHealthAPI{
				readinessResponse: tt.mockResponse,
				readinessError:    tt.mockError,
			}
			
			// Create server
			server, err := NewHTTPHealthServer(config, mockAPI, logger)
			require.NoError(t, err)
			
			// Create request
			req := httptest.NewRequest("GET", "/health/ready", nil)
			w := httptest.NewRecorder()
			
			// Call handler
			server.handleReadiness(w, req)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			// Check expected fields
			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestHTTPHealthServer_handleLiveness(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	
	tests := []struct {
		name           string
		mockResponse   *LivenessResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "system alive",
			mockResponse: &LivenessResponse{
				Alive:     true,
				Timestamp: time.Now(),
				Message:   "System is alive",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"alive": true,
				"message": "System is alive",
			},
		},
		{
			name: "system not alive",
			mockResponse: &LivenessResponse{
				Alive:     false,
				Timestamp: time.Now(),
				Message:   "System is not alive",
			},
			mockError:      nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"alive": false,
				"message": "System is not alive",
			},
		},
		{
			name:           "liveness API error",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock health API
			mockAPI := &mockHealthAPI{
				livenessResponse: tt.mockResponse,
				livenessError:    tt.mockError,
			}
			
			// Create server
			server, err := NewHTTPHealthServer(config, mockAPI, logger)
			require.NoError(t, err)
			
			// Create request
			req := httptest.NewRequest("GET", "/health/live", nil)
			w := httptest.NewRecorder()
			
			// Call handler
			server.handleLiveness(w, req)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			// Check expected fields
			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestHTTPHealthServer_GetServerInfo(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	mockAPI := &mockHealthAPI{}
	
	// Create server
	server, err := NewHTTPHealthServer(config, mockAPI, logger)
	require.NoError(t, err)
	
	// Get server info
	info := server.GetServerInfo()
	
	// Assert server info
	assert.Equal(t, true, info["enabled"])
	assert.Equal(t, "localhost", info["host"])
	assert.Equal(t, 8003, info["port"])
	assert.NotNil(t, info["start_time"])
	assert.NotNil(t, info["uptime"])
	
	endpoints, ok := info["endpoints"].([]string)
	require.True(t, ok)
	assert.Contains(t, endpoints, "/health")
	assert.Contains(t, endpoints, "/health/detailed")
	assert.Contains(t, endpoints, "/health/ready")
	assert.Contains(t, endpoints, "/health/live")
}

func TestHTTPHealthServer_Stop(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        true,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	mockAPI := &mockHealthAPI{}
	
	// Create server
	server, err := NewHTTPHealthServer(config, mockAPI, logger)
	require.NoError(t, err)
	
	// Stop server
	err = server.Stop()
	assert.NoError(t, err)
}

func TestHTTPHealthServer_Stop_Disabled(t *testing.T) {
	// Setup
	config := &config.HTTPHealthConfig{
		Enabled:        false,
		Host:           "localhost",
		Port:           8003,
		ReadTimeout:    "5s",
		WriteTimeout:   "5s",
		IdleTimeout:    "30s",
		BasicEndpoint:  "/health",
		DetailedEndpoint: "/health/detailed",
		ReadyEndpoint:  "/health/ready",
		LiveEndpoint:   "/health/live",
	}
	
	logger := logging.GetLogger("test")
	mockAPI := &mockHealthAPI{}
	
	// Create server
	server, err := NewHTTPHealthServer(config, mockAPI, logger)
	require.NoError(t, err)
	
	// Stop server
	err = server.Stop()
	assert.NoError(t, err)
}
