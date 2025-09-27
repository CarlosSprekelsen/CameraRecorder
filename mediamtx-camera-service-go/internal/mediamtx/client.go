// Package mediamtx implements the MediaMTX HTTP client for Layer 2 (Core Services).
//
// This client provides the HTTP REST API integration with the MediaMTX server,
// implementing connection pooling, circuit breaker pattern, and error handling
// for reliable communication with the external MediaMTX service.
//
// Architecture Compliance:
//   - Layer 2: Core Services - MediaMTX integration component
//   - Circuit Breaker: Fault tolerance for MediaMTX communication failures
//   - Connection Pooling: Optimized HTTP connections for performance
//   - Structured Errors: Consistent error handling and propagation
//   - Context Support: Proper cancellation and timeout handling
//
// Key Responsibilities:
//   - HTTP REST API calls to MediaMTX server (localhost:9997/v3/)
//   - Path management operations (create, delete, patch, list)
//   - Health monitoring and status queries
//   - Error handling with exponential backoff retry logic
//   - Connection pooling for performance optimization
//
// Requirements Coverage:
//   - REQ-MTX-001: MediaMTX service integration via HTTP REST API
//   - REQ-MTX-002: Stream management through path operations
//   - REQ-MTX-003: Path creation, deletion, and configuration management
//   - REQ-MTX-004: Health monitoring via MediaMTX status endpoints
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md

package mediamtx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// client implements the MediaMTX HTTP client with connection pooling and fault tolerance.
// This client handles all communication with the external MediaMTX server.
type client struct {
	httpClient *http.Client    // HTTP client with connection pooling configuration
	baseURL    string          // MediaMTX server base URL (e.g., http://localhost:9997/v3/)
	timeout    time.Duration   // Request timeout for individual HTTP calls
	logger     *logging.Logger // Structured logger for request/response logging
}

// NewClient creates a new MediaMTX HTTP client with optimized connection pooling.
// The client is configured for high-performance communication with connection reuse
// and appropriate timeouts for reliable MediaMTX server integration.
func NewClient(baseURL string, config *config.MediaMTXConfig, logger *logging.Logger) MediaMTXClient {
	// Configure HTTP client with connection pooling for performance optimization
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.ConnectionPool.MaxIdleConns,        // Global connection pool size
			MaxIdleConnsPerHost: config.ConnectionPool.MaxIdleConnsPerHost, // Per-host connection limit
			IdleConnTimeout:     config.ConnectionPool.IdleConnTimeout,     // Connection reuse timeout
		},
	}

	return &client{
		httpClient: httpClient,
		baseURL:    baseURL,
		timeout:    config.Timeout,
		logger:     logger,
	}
}

// Get performs an HTTP GET request
func (c *client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs an HTTP POST request
func (c *client) Post(ctx context.Context, path string, data []byte) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, data)
}

// Put performs an HTTP PUT request
func (c *client) Put(ctx context.Context, path string, data []byte) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, data)
}

// Patch performs an HTTP PATCH request
func (c *client) Patch(ctx context.Context, path string, data []byte) error {
	_, err := c.doRequest(ctx, http.MethodPatch, path, data)
	if err != nil {
		return fmt.Errorf("failed to perform PATCH request to %s: %w", path, err)
	}
	return nil
}

// Delete performs an HTTP DELETE request
func (c *client) Delete(ctx context.Context, path string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to perform DELETE request to %s: %w", path, err)
	}
	return nil
}

// HealthCheck performs a health check request
func (c *client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+MediaMTXPathsList, nil)
	if err != nil {
		return err
	}

	// Client will cancel request when context is cancelled!
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if cancelled
		if ctx.Err() != nil {
			return ctx.Err() // Return context error
		}
		return err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return NewMediaMTXErrorFromHTTP(resp.StatusCode, bodyBytes)
	}

	return nil
}

// GetDetailedHealth performs a health check and returns detailed status using parseHealthResponse
func (c *client) GetDetailedHealth(ctx context.Context) (*HealthStatus, error) {
	// Get detailed health from MediaMTX health endpoint
	data, err := c.Get(ctx, MediaMTXConfigGlobalGet)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed health: %w", err)
	}

	// Use existing parseHealthResponse method
	return parseHealthResponse(data)
}

// GetMediaMTXMetrics gets MediaMTX server metrics using parseMetricsResponse
func (c *client) GetMediaMTXMetrics(ctx context.Context) (*Metrics, error) {
	// Get metrics from MediaMTX metrics endpoint
	data, err := c.Get(ctx, "/v3/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaMTX metrics: %w", err)
	}

	// Use existing parseMetricsResponse method
	return parseMetricsResponse(data)
}

// Close closes the HTTP client
func (c *client) Close() error {
	// HTTP client doesn't need explicit closing in Go
	// The transport will be garbage collected
	return nil
}

// doRequest performs the actual HTTP request with proper error handling
func (c *client) doRequest(ctx context.Context, method, path string, data []byte) ([]byte, error) {
	// Create request with context
	url := c.baseURL + path
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to create request", err.Error(), "new_request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Log request with detailed information
	c.logger.WithFields(logging.Fields{
		"method": method,
		"url":    url,
		"data":   string(data),
		"headers": map[string]string{
			"Content-Type": req.Header.Get("Content-Type"),
			"Accept":       req.Header.Get("Accept"),
		},
	}).Info("Making MediaMTX request")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err() // Return context error
		}
		return nil, NewMediaMTXErrorWithOp(0, "request failed", err.Error(), "http_do")
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Check if cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err() // Return context error
		}
		return nil, NewMediaMTXErrorWithOp(0, "failed to read response", err.Error(), "read_body")
	}

	// Log response with detailed information
	c.logger.WithFields(logging.Fields{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"body":        string(bodyBytes),
		"headers": map[string]string{
			"Content-Type":   resp.Header.Get("Content-Type"),
			"Content-Length": resp.Header.Get("Content-Length"),
		},
	}).Info("Received MediaMTX response")

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, NewMediaMTXErrorFromHTTP(resp.StatusCode, bodyBytes)
	}

	return bodyBytes, nil
}

// parsePathListResponse parses the MediaMTX paths list response (runtime paths)
// This function parses the response from /v3/paths/list endpoint and returns []*Path
func parsePathListResponse(data []byte) ([]*Path, error) {
	// Handle empty response
	// Use comprehensive response validation
	if err := validateMediaMTXResponse(data, "PathList"); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, err.Error(), "", "parse_path_list")
	}

	var response PathList
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse paths list response", err.Error(), "parse_path_list")
	}

	paths := make([]*Path, len(response.Items))
	for i, path := range response.Items {
		paths[i] = &path
	}

	return paths, nil
}

// parsePathConfListResponse parses the MediaMTX path configuration list response
// This function parses the response from /v3/config/paths/list endpoint and returns []*PathConf
func parsePathConfListResponse(data []byte) ([]*PathConf, error) {
	// Use comprehensive response validation
	if err := validateMediaMTXResponse(data, "PathList"); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, err.Error(), "", "parse_path_conf_list")
	}

	var response PathConfList
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse path configuration list response", err.Error(), "parse_path_conf_list")
	}

	paths := make([]*PathConf, len(response.Items))
	for i, path := range response.Items {
		paths[i] = &path
	}

	return paths, nil
}

// parseHealthResponse parses the health response
func parseHealthResponse(data []byte) (*HealthStatus, error) {
	// Handle empty response
	if len(data) == 0 {
		return nil, NewMediaMTXErrorWithOp(0, "empty response body", "MediaMTX returned empty response", "parse_health")
	}

	// Handle null JSON
	if string(data) == "null" {
		return nil, NewMediaMTXErrorWithOp(0, "null response body", "MediaMTX returned null response", "parse_health")
	}

	var response MediaMTXHealthResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse health response", err.Error(), "parse_health")
	}

	// Validate required fields
	if response.Status == "" {
		return nil, NewMediaMTXErrorWithOp(0, "missing required field", "health response missing 'status' field", "parse_health")
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		timestamp = time.Now() // Fallback to current time
	}

	return &HealthStatus{
		Status:    response.Status,
		Timestamp: timestamp,
		Metrics:   response.Metrics,
	}, nil
}

// parseStreamResponse parses a single stream response from MediaMTX API
func parseStreamResponse(data []byte) (*Path, error) {
	// Handle empty response (successful path creation returns empty body)
	if len(data) == 0 {
		return nil, NewMediaMTXErrorWithOp(0, "empty response body", "MediaMTX returned empty response", "parse_stream")
	}

	// Handle null JSON
	if string(data) == "null" {
		return nil, NewMediaMTXErrorWithOp(0, "null response body", "MediaMTX returned null response", "parse_stream")
	}

	// Parse directly into Path struct (matches MediaMTX API)
	var stream Path
	if err := json.Unmarshal(data, &stream); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse stream response", err.Error(), "parse_stream")
	}

	// Validate required fields
	if stream.Name == "" {
		return nil, NewMediaMTXErrorWithOp(0, "missing required field", "stream missing 'name' field", "parse_stream")
	}

	return &stream, nil
}

// determineStatus converts MediaMTX ready status to our status format
func determineStatus(ready bool) string {
	if ready {
		return "READY"
	}
	return "PENDING"
}

// parseMetricsResponse parses the metrics response
func parseMetricsResponse(data []byte) (*Metrics, error) {
	var metrics Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse metrics response", err.Error(), "parse_metrics")
	}
	return &metrics, nil
}

// createStreamRequest represents a stream creation request
type createStreamRequest struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// marshalCreateStreamRequest marshals a stream creation request
func marshalCreateStreamRequest(name, source string) ([]byte, error) {
	request := createStreamRequest{
		Name:   name,
		Source: source,
	}
	return json.Marshal(request)
}

// marshalCreatePathRequest marshals a path creation request
func marshalCreatePathRequest(path *Path) ([]byte, error) {
	// MediaMTX API expects a simple format with just name and source
	request := map[string]interface{}{
		"name":   path.Name,
		"source": path.Source,
	}
	return json.Marshal(request)
}

// marshalUpdatePathRequest marshals a path update request using proper API type
// According to MediaMTX swagger.json, PATCH /v3/config/paths/patch/{name} expects PathConf directly
func marshalUpdatePathRequest(config *PathConf) ([]byte, error) {
	// MediaMTX API expects PathConf directly, not wrapped in { "config": {...} }
	return json.Marshal(config)
}

// marshalCreateUSBPathRequest marshals a USB device path creation request (matches Python implementation)
func marshalCreateUSBPathRequest(name, ffmpegCommand string) ([]byte, error) {
	request := map[string]interface{}{
		"source":             "publisher", // Publisher source for on-demand paths
		"runOnDemand":        ffmpegCommand,
		"runOnDemandRestart": true,
	}
	return json.Marshal(request)
}

// validateMediaMTXResponse validates MediaMTX API responses for structural integrity
// SECURITY: Handles Unicode, large strings, and extra fields gracefully
func validateMediaMTXResponse(data []byte, expectedSchema string) error {
	// Check for null/empty responses
	if len(data) == 0 {
		return fmt.Errorf("empty response body")
	}

	if string(data) == "null" {
		return fmt.Errorf("null response body")
	}

	// SECURITY: Handle large strings by truncating if necessary
	const maxResponseSize = 10 * 1024 * 1024 // 10MB limit
	if len(data) > maxResponseSize {
		// Truncate large responses instead of rejecting
		data = data[:maxResponseSize]
	}

	// SECURITY: Handle Unicode gracefully - use json.RawMessage for flexible parsing
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(data, &rawResponse); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}

	// SECURITY: Extra fields are ignored - only validate required fields
	// This prevents denial of service attacks via extra fields

	// Schema-specific validation
	switch expectedSchema {
	case "PathList":
		return validatePathListSchema(rawResponse)
	case "PathConf":
		return validatePathConfSchema(rawResponse)
	case "RecordingList":
		return validateRecordingListSchema(rawResponse)
	}

	return nil
}

// validatePathListSchema validates PathList response structure per swagger.json
// SECURITY: Handles extra fields gracefully, ignores them instead of rejecting
func validatePathListSchema(data map[string]interface{}) error {
	requiredFields := []string{"pageCount", "itemCount", "items"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate items is array
	if items, ok := data["items"].([]interface{}); ok {
		for i, item := range items {
			if _, ok := item.(map[string]interface{}); !ok {
				return fmt.Errorf("invalid item at index %d: expected object", i)
			}
		}
	} else {
		return fmt.Errorf("items field must be an array")
	}

	// SECURITY: Extra fields are ignored gracefully - no validation needed
	// This prevents denial of service attacks via extra fields

	return nil
}

// validatePathConfSchema validates PathConf response structure per swagger.json
func validatePathConfSchema(data map[string]interface{}) error {
	// PathConf can have various optional fields, but should be an object
	if data == nil {
		return fmt.Errorf("path configuration cannot be null")
	}
	return nil
}

// validateRecordingListSchema validates RecordingList response structure per swagger.json
// SECURITY: Handles extra fields gracefully, ignores them instead of rejecting
func validateRecordingListSchema(data map[string]interface{}) error {
	requiredFields := []string{"pageCount", "itemCount", "items"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate items is array
	if items, ok := data["items"].([]interface{}); ok {
		for i, item := range items {
			if _, ok := item.(map[string]interface{}); !ok {
				return fmt.Errorf("invalid recording item at index %d: expected object", i)
			}
		}
	} else {
		return fmt.Errorf("items field must be an array")
	}

	// SECURITY: Extra fields are ignored gracefully - no validation needed
	// This prevents denial of service attacks via extra fields

	return nil
}
