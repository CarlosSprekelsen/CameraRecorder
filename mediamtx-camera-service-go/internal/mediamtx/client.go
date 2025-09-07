/*
MediaMTX HTTP Client Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// client represents the MediaMTX HTTP client
type client struct {
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
	logger     *logging.Logger
}

// NewClient creates a new MediaMTX HTTP client
func NewClient(baseURL string, config *MediaMTXConfig, logger *logging.Logger) MediaMTXClient {
	// Create HTTP client with connection pooling
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.ConnectionPool.MaxIdleConns,
			MaxIdleConnsPerHost: config.ConnectionPool.MaxIdleConnsPerHost,
			IdleConnTimeout:     config.ConnectionPool.IdleConnTimeout,
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
	_, err := c.Get(ctx, "/v3/paths/list")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
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

	// Log request
	c.logger.WithFields(logging.Fields{
		"method": method,
		"url":    url,
		"data":   string(data),
	}).Debug("Making MediaMTX request")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "request failed", err.Error(), "http_do")
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to read response", err.Error(), "read_body")
	}

	// Log response
	c.logger.WithFields(logging.Fields{
		"status_code": resp.StatusCode,
		"body":        string(bodyBytes),
	}).Debug("Received MediaMTX response")

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, NewMediaMTXErrorFromHTTP(resp.StatusCode, bodyBytes)
	}

	return bodyBytes, nil
}

// getStreamsResponse represents the MediaMTX streams response
type getStreamsResponse struct {
	ItemCount int      `json:"itemCount"`
	PageCount int      `json:"pageCount"`
	Items     []Stream `json:"items"`
}

// getPathsResponse represents the MediaMTX paths response
type getPathsResponse struct {
	ItemCount int                    `json:"itemCount"`
	PageCount int                    `json:"pageCount"`
	Items     []MediaMTXPathResponse `json:"items"`
}

// healthResponse represents the MediaMTX health response
type healthResponse struct {
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Metrics   Metrics `json:"metrics"`
}

// parseStreamsResponse parses the streams response
func parseStreamsResponse(data []byte) ([]*Stream, error) {
	// Handle empty response
	if len(data) == 0 {
		return nil, NewMediaMTXErrorWithOp(0, "empty response body", "MediaMTX returned empty response", "parse_streams")
	}

	// Handle null JSON
	if string(data) == "null" {
		return nil, NewMediaMTXErrorWithOp(0, "null response body", "MediaMTX returned null response", "parse_streams")
	}

	var response getStreamsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse streams response", err.Error(), "parse_streams")
	}

	// Validate required fields
	if response.Items == nil {
		return nil, NewMediaMTXErrorWithOp(0, "missing required field", "response missing 'items' field", "parse_streams")
	}

	streams := make([]*Stream, len(response.Items))
	for i, stream := range response.Items {
		streams[i] = &stream
	}

	return streams, nil
}

// parsePathsResponse parses the paths response
func parsePathsResponse(data []byte) ([]*Path, error) {
	// Handle empty response
	if len(data) == 0 {
		return nil, NewMediaMTXErrorWithOp(0, "empty response body", "MediaMTX returned empty response", "parse_paths")
	}

	// Handle null JSON
	if string(data) == "null" {
		return nil, NewMediaMTXErrorWithOp(0, "null response body", "MediaMTX returned null response", "parse_paths")
	}

	var response getPathsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse paths response", err.Error(), "parse_paths")
	}

	// Validate required fields
	if response.Items == nil {
		return nil, NewMediaMTXErrorWithOp(0, "missing required field", "response missing 'items' field", "parse_paths")
	}

	paths := make([]*Path, len(response.Items))
	for i, item := range response.Items {
		// Convert MediaMTXPathResponse to Path
		path := &Path{
			ID:   item.Name, // Use name as ID
			Name: item.Name,
			// Extract source string from complex object
			Source: extractSourceString(item.Source),
		}
		paths[i] = path
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

	var response healthResponse
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
func parseStreamResponse(data []byte) (*Stream, error) {
	// Handle empty response (successful path creation returns empty body)
	if len(data) == 0 {
		return nil, NewMediaMTXErrorWithOp(0, "empty response body", "MediaMTX returned empty response", "parse_stream")
	}

	// Handle null JSON
	if string(data) == "null" {
		return nil, NewMediaMTXErrorWithOp(0, "null response body", "MediaMTX returned null response", "parse_stream")
	}

	// Parse directly into Stream struct (matches MediaMTX API)
	var stream Stream
	if err := json.Unmarshal(data, &stream); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse stream response", err.Error(), "parse_stream")
	}

	// Validate required fields
	if stream.Name == "" {
		return nil, NewMediaMTXErrorWithOp(0, "missing required field", "stream missing 'name' field", "parse_stream")
	}

	return &stream, nil
}

// extractSourceString extracts source information from MediaMTX API response
func extractSourceString(source interface{}) string {
	if source == nil {
		return ""
	}

	// Handle different source formats from MediaMTX API
	switch v := source.(type) {
	case string:
		return v
	case map[string]interface{}:
		if sourceType, ok := v["type"].(string); ok {
			return sourceType
		}
	}
	return ""
}

// determineStatus converts MediaMTX ready status to our status format
func determineStatus(ready bool) string {
	if ready {
		return "READY"
	}
	return "PENDING"
}

// parsePathResponse parses a single path response
func parsePathResponse(data []byte) (*Path, error) {
	var path Path
	if err := json.Unmarshal(data, &path); err != nil {
		return nil, NewMediaMTXErrorWithOp(0, "failed to parse path response", err.Error(), "parse_path")
	}
	return &path, nil
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

// createPathRequest represents a path creation request
type createPathRequest struct {
	Name   string                 `json:"name"`
	Source string                 `json:"source"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// updatePathRequest represents a path update request
type updatePathRequest struct {
	Config map[string]interface{} `json:"config"`
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

// marshalUpdatePathRequest marshals a path update request
func marshalUpdatePathRequest(config map[string]interface{}) ([]byte, error) {
	request := updatePathRequest{
		Config: config,
	}
	return json.Marshal(request)
}

// marshalCreateUSBPathRequest marshals a USB device path creation request (matches Python implementation)
func marshalCreateUSBPathRequest(name, ffmpegCommand string) ([]byte, error) {
	request := map[string]interface{}{
		"runOnDemand":        ffmpegCommand,
		"runOnDemandRestart": true,
	}
	return json.Marshal(request)
}
