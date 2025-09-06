/*
MediaMTX Test Helpers - Unit Testing with Real MediaMTX Server

This file provides utilities for unit testing against the REAL MediaMTX server
using the correct Go API endpoints as documented in the API documentation.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Global mutex to prevent parallel test execution
// MediaMTX tests must run sequentially because they share the same server resources
var testMutex sync.Mutex

// MediaMTXTestConfig provides configuration for MediaMTX server testing
type MediaMTXTestConfig struct {
	BaseURL      string // MediaMTX API base URL (http://localhost:9997)
	Timeout      time.Duration
	TestDataDir  string
	CleanupAfter bool
}

// DefaultMediaMTXTestConfig returns default configuration for MediaMTX server testing
func DefaultMediaMTXTestConfig() *MediaMTXTestConfig {
	return &MediaMTXTestConfig{
		BaseURL:      "http://localhost:9997", // MediaMTX API port (standard)
		Timeout:      30 * time.Second,
		TestDataDir:  "/tmp/mediamtx_test_data",
		CleanupAfter: true,
	}
}

// MediaMTXTestHelper provides utilities for MediaMTX server testing
type MediaMTXTestHelper struct {
	config *MediaMTXTestConfig
	logger *logging.Logger
	client MediaMTXClient
}

// EnsureSequentialExecution ensures tests run sequentially to avoid MediaMTX server conflicts
// Call this at the beginning of each test that uses MediaMTX server
func EnsureSequentialExecution(t *testing.T) {
	testMutex.Lock()
	t.Cleanup(func() {
		testMutex.Unlock()
	})
}

// NewMediaMTXTestHelper creates a new test helper for MediaMTX server testing
func NewMediaMTXTestHelper(t *testing.T, config *MediaMTXTestConfig) *MediaMTXTestHelper {
	if config == nil {
		config = DefaultMediaMTXTestConfig()
	}

	// Create logger for testing
	logger := logging.NewLogger("test-helper")
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise during tests

	// Create MediaMTX client configuration
	clientConfig := &MediaMTXConfig{
		BaseURL:        config.BaseURL,
		HealthCheckURL: config.BaseURL + "/v3/paths/list", // Correct Go MediaMTX health check endpoint
		Timeout:        config.Timeout,
		ConnectionPool: ConnectionPoolConfig{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Create MediaMTX client
	client := NewClient(config.BaseURL, clientConfig, logger)

	helper := &MediaMTXTestHelper{
		config: config,
		logger: logger,
		client: client,
	}

	// Ensure test data directory exists
	err := helper.ensureTestDataDir()
	require.NoError(t, err, "Failed to create test data directory")

	return helper
}

// ensureTestDataDir creates the test data directory if it doesn't exist
func (h *MediaMTXTestHelper) ensureTestDataDir() error {
	return os.MkdirAll(h.config.TestDataDir, 0755)
}

// Cleanup performs comprehensive cleanup of test resources
func (h *MediaMTXTestHelper) Cleanup(t *testing.T) {
	if h.config == nil || !h.config.CleanupAfter {
		return
	}

	t.Log("Starting MediaMTX test cleanup...")

	// Clean up MediaMTX paths created during tests
	h.cleanupMediaMTXPaths(t)

	// Clean up local test data
	h.cleanupLocalTestData(t)

	// Close client connection
	if h.client != nil {
		h.client.Close()
	}

	t.Log("MediaMTX test cleanup completed")
}

// WaitForServerReady waits for the MediaMTX server to be ready using health check
func (h *MediaMTXTestHelper) WaitForServerReady(t *testing.T, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for MediaMTX server to be ready")
		case <-ticker.C:
			// Use MediaMTX health check via /v3/paths/list endpoint
			err := h.client.HealthCheck(ctx)
			if err != nil {
				continue
			}
			t.Log("MediaMTX server is ready")
			return nil
		}
	}
}

// TestMediaMTXHealth tests the MediaMTX health check
func (h *MediaMTXTestHelper) TestMediaMTXHealth(t *testing.T) error {
	ctx := context.Background()
	err := h.client.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("MediaMTX health check failed: %w", err)
	}
	t.Log("MediaMTX health check passed")
	return nil
}

// TestMediaMTXPaths tests the MediaMTX paths endpoint
func (h *MediaMTXTestHelper) TestMediaMTXPaths(t *testing.T) error {
	ctx := context.Background()
	data, err := h.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Verify we got a valid response (should be JSON array)
	if len(data) == 0 {
		return fmt.Errorf("empty response from paths endpoint")
	}

	t.Logf("MediaMTX paths endpoint returned %d bytes", len(data))
	return nil
}

// TestMediaMTXConfigPaths tests the MediaMTX config paths endpoint
func (h *MediaMTXTestHelper) TestMediaMTXConfigPaths(t *testing.T) error {
	ctx := context.Background()
	data, err := h.client.Get(ctx, "/v3/config/paths/list")
	if err != nil {
		return fmt.Errorf("failed to get config paths: %w", err)
	}

	// Verify we got a valid response (should be JSON)
	if len(data) == 0 {
		return fmt.Errorf("empty response from config paths endpoint")
	}

	t.Logf("MediaMTX config paths endpoint returned %d bytes", len(data))
	return nil
}

// TestMediaMTXFailure tests MediaMTX server failure scenarios
func (h *MediaMTXTestHelper) TestMediaMTXFailure(t *testing.T) error {
	ctx := context.Background()

	// Test invalid endpoint
	_, err := h.client.Get(ctx, "/v3/invalid/endpoint")
	if err == nil {
		return fmt.Errorf("expected error for invalid endpoint")
	}
	t.Logf("Expected failure for invalid endpoint: %v", err)

	// Test invalid path creation (using correct endpoint)
	_, err = h.client.Post(ctx, "/v3/config/paths/add", []byte(`{"invalid": "data"}`))
	if err == nil {
		return fmt.Errorf("expected error for invalid path creation")
	}
	t.Logf("Expected failure for invalid path creation: %v", err)

	return nil
}

// SimulateMediaMTXFailure simulates MediaMTX server failure for testing error handling
func (h *MediaMTXTestHelper) SimulateMediaMTXFailure(t *testing.T) error {
	// Create a client with invalid URL to simulate server failure
	invalidConfig := &MediaMTXConfig{
		BaseURL: "http://localhost:9999", // Invalid port
		Timeout: 1 * time.Second,
	}

	invalidClient := NewClient("http://localhost:9999", invalidConfig, h.logger)
	defer invalidClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := invalidClient.HealthCheck(ctx)
	if err != nil {
		t.Logf("Simulated MediaMTX failure: %v", err)
		return nil // Expected failure
	}

	return fmt.Errorf("expected MediaMTX failure but connection succeeded")
}

// GetConfig returns the test configuration
func (h *MediaMTXTestHelper) GetConfig() *MediaMTXTestConfig {
	return h.config
}

// GetLogger returns the test logger
func (h *MediaMTXTestHelper) GetLogger() *logging.Logger {
	return h.logger
}

// GetClient returns the MediaMTX client for testing
func (h *MediaMTXTestHelper) GetClient() MediaMTXClient {
	return h.client
}

// CreateTestPath creates a test path for testing purposes
func (h *MediaMTXTestHelper) CreateTestPath(t *testing.T, name string) error {
	ctx := context.Background()
	pathData := fmt.Sprintf(`{"name":"%s","source":"publisher"}`, name)
	_, err := h.client.Post(ctx, "/v3/config/paths/add", []byte(pathData))
	if err != nil {
		return fmt.Errorf("failed to create test path %s: %w", name, err)
	}
	t.Logf("Created test path: %s", name)
	return nil
}

// DeleteTestPath deletes a test path
func (h *MediaMTXTestHelper) DeleteTestPath(t *testing.T, name string) error {
	ctx := context.Background()
	err := h.client.Delete(ctx, "/v3/config/paths/delete/"+name)
	if err != nil {
		return fmt.Errorf("failed to delete test path %s: %w", name, err)
	}
	t.Logf("Deleted test path: %s", name)
	return nil
}

// GetPathInfo gets information about a specific path
func (h *MediaMTXTestHelper) GetPathInfo(t *testing.T, name string) ([]byte, error) {
	ctx := context.Background()
	data, err := h.client.Get(ctx, "/v3/paths/get/"+name)
	if err != nil {
		return nil, fmt.Errorf("failed to get path info for %s: %w", name, err)
	}
	return data, nil
}

// cleanupMediaMTXPaths cleans up all MediaMTX paths created during tests
func (h *MediaMTXTestHelper) cleanupMediaMTXPaths(t *testing.T) {
	if h.client == nil {
		return
	}

	ctx := context.Background()

	// Get all paths from MediaMTX
	data, err := h.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		t.Logf("Warning: Failed to get paths for cleanup: %v", err)
		return
	}

	// Parse paths response to find test paths
	// Note: This is a simplified cleanup - in production you'd want more sophisticated path tracking
	var pathsResponse struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &pathsResponse); err != nil {
		t.Logf("Warning: Failed to parse paths response: %v", err)
		return
	}

	// Delete test paths (paths that start with "test_" or "camera_")
	// Only try to delete paths that actually exist in the current response
	for _, path := range pathsResponse.Items {
		if h.isTestPath(path.Name) {
			err := h.client.Delete(ctx, "/v3/config/paths/delete/"+path.Name)
			if err != nil {
				// Only log as warning if it's not a 404 (path not found) error
				if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
					t.Logf("Warning: Failed to delete test path %s: %v", path.Name, err)
				}
			} else {
				t.Logf("Cleaned up test path: %s", path.Name)
			}
		}
	}
}

// isTestPath determines if a path was created during testing
func (h *MediaMTXTestHelper) isTestPath(pathName string) bool {
	// Check for common test path patterns
	testPrefixes := []string{"test_", "camera_", "rec_"}
	for _, prefix := range testPrefixes {
		if len(pathName) > len(prefix) && pathName[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// cleanupLocalTestData cleans up local test files and directories
func (h *MediaMTXTestHelper) cleanupLocalTestData(t *testing.T) {
	if h.config == nil || h.config.TestDataDir == "" {
		return
	}

	// Remove test data directory
	if err := os.RemoveAll(h.config.TestDataDir); err != nil {
		t.Logf("Warning: Failed to remove test data directory %s: %v", h.config.TestDataDir, err)
	} else {
		t.Logf("Cleaned up test data directory: %s", h.config.TestDataDir)
	}
}
