/*
MediaMTX Client Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient_ReqMTX001 tests client creation
func TestNewClient_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client, "Client should not be nil")
}

// TestClient_Get_ReqMTX001 tests GET request functionality
func TestClient_Get_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	data, err := client.Get(ctx, "/test")
	require.NoError(t, err, "GET request should succeed")
	assert.Equal(t, `{"status":"ok"}`, string(data))
}

// TestClient_Post_ReqMTX001 tests POST request functionality
func TestClient_Post_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	data, err := client.Post(ctx, "/test", []byte(`{"name":"test"}`))
	require.NoError(t, err, "POST request should succeed")
	assert.Equal(t, `{"id":"123"}`, string(data))
}

// TestClient_Put_ReqMTX001 tests PUT request functionality
func TestClient_Put_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated":true}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	data, err := client.Put(ctx, "/test", []byte(`{"name":"updated"}`))
	require.NoError(t, err, "PUT request should succeed")
	assert.Equal(t, `{"updated":true}`, string(data))
}

// TestClient_Delete_ReqMTX001 tests DELETE request functionality
func TestClient_Delete_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	err := client.Delete(ctx, "/test")
	require.NoError(t, err, "DELETE request should succeed")
}

// TestClient_HealthCheck_ReqMTX004 tests health check functionality
func TestClient_HealthCheck_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL:        mockServer.URL,
		HealthCheckURL: mockServer.URL + "/health",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	err := client.HealthCheck(ctx)
	require.NoError(t, err, "Health check should succeed")
}

// TestClient_HealthCheck_Failure_ReqMTX004 tests health check failure handling
func TestClient_HealthCheck_Failure_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Create mock server that fails
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL:        mockServer.URL,
		HealthCheckURL: mockServer.URL + "/health",
		Timeout:        5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	err := client.HealthCheck(ctx)
	require.Error(t, err, "Health check should fail with server error")
}

// TestClient_Timeout_ReqMTX007 tests timeout handling
func TestClient_Timeout_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Create mock server that delays response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 50 * time.Millisecond, // Short timeout
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()
	_, err := client.Get(ctx, "/test")
	require.Error(t, err, "Request should timeout")
}

// TestClient_Close_ReqMTX001 tests client cleanup
func TestClient_Close_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client)

	err := client.Close()
	require.NoError(t, err, "Client should close successfully")
}

// TestClient_ErrorHandling_ReqMTX007 tests various error scenarios
func TestClient_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	// Create mock server that returns different error codes
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/400":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"bad request"}`))
		case "/404":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		case "/500":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	ctx := context.Background()

	// Test 400 error
	_, err := client.Get(ctx, "/400")
	require.Error(t, err, "Should get error for 400 status")

	// Test 404 error
	_, err = client.Get(ctx, "/404")
	require.Error(t, err, "Should get error for 404 status")

	// Test 500 error
	_, err = client.Get(ctx, "/500")
	require.Error(t, err, "Should get error for 500 status")
}

// TestClient_ConcurrentRequests_ReqMTX001 tests concurrent request handling
func TestClient_ConcurrentRequests_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer mockServer.Close()

	config := &MediaMTXConfig{
		BaseURL: mockServer.URL,
		Timeout: 5 * time.Second,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	client := NewClient(mockServer.URL, config, logger)
	require.NotNil(t, client)

	// Test concurrent requests
	done := make(chan bool, 5)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		go func() {
			_, err := client.Get(ctx, "/test")
			assert.NoError(t, err, "Concurrent request should succeed")
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}
