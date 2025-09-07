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
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	config                *MediaMTXTestConfig
	logger                *logging.Logger
	client                MediaMTXClient
	pathManager           PathManager
	streamManager         StreamManager
	recordingManager      *RecordingManager
	rtspConnectionManager RTSPConnectionManager
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
	// Performance optimization: Since MediaMTX is already running (systemd service),
	// we can skip the polling and just do a quick health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Single health check instead of polling
	err := h.client.HealthCheck(ctx)
	if err != nil {
		// If health check fails, fall back to original polling behavior
		t.Logf("Quick health check failed, falling back to polling: %v", err)
		return h.waitForServerReadyWithPolling(t, timeout)
	}

	t.Log("MediaMTX server is ready")
	return nil
}

// waitForServerReadyWithPolling implements the original polling behavior
func (h *MediaMTXTestHelper) waitForServerReadyWithPolling(t *testing.T, timeout time.Duration) error {
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

// GetPathManager returns a shared path manager instance
func (h *MediaMTXTestHelper) GetPathManager() PathManager {
	if h.pathManager == nil {
		// Convert test config to MediaMTX config
		mediaMTXConfig := &MediaMTXConfig{
			BaseURL: h.config.BaseURL,
			Timeout: 10 * time.Second,
		}
		h.pathManager = NewPathManager(h.client, mediaMTXConfig, h.logger)
	}
	return h.pathManager
}

// GetStreamManager returns a shared stream manager instance
func (h *MediaMTXTestHelper) GetStreamManager() StreamManager {
	if h.streamManager == nil {
		// Convert test config to MediaMTX config
		mediaMTXConfig := &MediaMTXConfig{
			BaseURL: h.config.BaseURL,
			Timeout: 10 * time.Second,
		}
		h.streamManager = NewStreamManager(h.client, mediaMTXConfig, h.logger)
	}
	return h.streamManager
}

// GetRecordingManager returns a shared recording manager instance
func (h *MediaMTXTestHelper) GetRecordingManager() *RecordingManager {
	if h.recordingManager == nil {
		// Convert test config to MediaMTX config
		mediaMTXConfig := &MediaMTXConfig{
			BaseURL: h.config.BaseURL,
			Timeout: 10 * time.Second,
		}
		pathManager := h.GetPathManager()
		streamManager := h.GetStreamManager()
		h.recordingManager = NewRecordingManager(h.client, pathManager, streamManager, mediaMTXConfig, h.logger)
	}
	return h.recordingManager
}

// GetRTSPConnectionManager returns a shared RTSP connection manager instance
func (h *MediaMTXTestHelper) GetRTSPConnectionManager() RTSPConnectionManager {
	if h.rtspConnectionManager == nil {
		// Convert test config to MediaMTX config
		mediaMTXConfig := &MediaMTXConfig{
			BaseURL: h.config.BaseURL,
			Timeout: 10 * time.Second,
		}
		h.rtspConnectionManager = NewRTSPConnectionManager(h.client, mediaMTXConfig, h.logger)
	}
	return h.rtspConnectionManager
}

// CreateConfigManagerWithFixture creates a config manager that loads from test fixtures
func CreateConfigManagerWithFixture(t *testing.T, fixtureName string) *config.ConfigManager {
	configManager := config.CreateConfigManager()

	// Use test fixture instead of creating config manually
	fixturePath := filepath.Join("tests", "fixtures", fixtureName)

	// Check if fixture exists, if not use a fallback path
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		// Try alternative path
		fixturePath = filepath.Join("..", "..", "tests", "fixtures", fixtureName)
	}

	err := configManager.LoadConfig(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load config from fixture %s: %v", fixtureName, err)
	}

	return configManager
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

	// Also check for camera0, camera1, etc. (without underscore)
	if len(pathName) >= 7 && pathName[:6] == "camera" && pathName[6] >= '0' && pathName[6] <= '9' {
		return true
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

// ============================================================================
// INPUT VALIDATION TEST HELPERS
// ============================================================================
// These helpers are designed to catch dangerous bugs through systematic
// input validation testing, not just achieve coverage.

// InputValidationTestScenario represents a test scenario for input validation
type InputValidationTestScenario struct {
	Name         string
	Page         int
	ItemsPerPage int
	ExpectError  bool
	ErrorMsg     string
	Description  string
}

// GetRTSPInputValidationScenarios returns comprehensive input validation scenarios
// for RTSP connection manager that can catch dangerous bugs
func GetRTSPInputValidationScenarios() []InputValidationTestScenario {
	return []InputValidationTestScenario{
		{
			Name:         "negative_page_number",
			Page:         -1,
			ItemsPerPage: 10,
			ExpectError:  true, // Should be rejected with clear error message
			ErrorMsg:     "invalid page number",
			Description:  "Negative page numbers should be rejected",
		},
		{
			Name:         "zero_items_per_page",
			Page:         0,
			ItemsPerPage: 0,
			ExpectError:  true, // Should be rejected with clear error message
			ErrorMsg:     "invalid page number",
			Description:  "Zero page numbers should be rejected",
		},
		{
			Name:         "negative_items_per_page",
			Page:         1,
			ItemsPerPage: -5,
			ExpectError:  true, // Should be rejected with clear error message
			ErrorMsg:     "invalid items per page",
			Description:  "Negative items per page should be rejected",
		},
		{
			Name:         "extremely_large_page",
			Page:         999999999,
			ItemsPerPage: 10,
			ExpectError:  false, // Should handle gracefully, not cause integer overflow
			ErrorMsg:     "",
			Description:  "Extremely large page numbers should be handled gracefully",
		},
		{
			Name:         "extremely_large_items_per_page",
			Page:         1,
			ItemsPerPage: 999999999,
			ExpectError:  true, // Should be rejected - too large
			ErrorMsg:     "invalid items per page",
			Description:  "Extremely large items per page should be rejected",
		},
		{
			Name:         "max_int_page",
			Page:         2147483647, // Max int32
			ItemsPerPage: 10,
			ExpectError:  false, // Should handle gracefully
			ErrorMsg:     "",
			Description:  "Maximum integer page should be handled gracefully",
		},
		{
			Name:         "max_int_items_per_page",
			Page:         1,
			ItemsPerPage: 2147483647, // Max int32
			ExpectError:  true,       // Should be rejected - too large
			ErrorMsg:     "invalid items per page",
			Description:  "Maximum integer items per page should be rejected",
		},
	}
}

// TestRTSPInputValidation tests RTSP connection manager input validation
// This function is designed to catch dangerous bugs, not just achieve coverage
func (h *MediaMTXTestHelper) TestRTSPInputValidation(t *testing.T, rtspManager RTSPConnectionManager) {
	ctx := context.Background()
	scenarios := GetRTSPInputValidationScenarios()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Testing scenario: %s - %s", scenario.Name, scenario.Description)

			// Test the input validation
			_, err := rtspManager.ListConnections(ctx, scenario.Page, scenario.ItemsPerPage)

			if scenario.ExpectError {
				// Should get an error
				require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
				if scenario.ErrorMsg != "" {
					assert.Contains(t, err.Error(), scenario.ErrorMsg,
						"Error message should contain expected text for scenario %s", scenario.Name)
				}
				t.Logf("✅ Scenario %s correctly produced expected error: %v", scenario.Name, err)
			} else {
				// Should NOT get an error (graceful handling)
				if err != nil {
					// This is a BUG - the API should handle these inputs gracefully
					t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
					t.Errorf("🚨 This indicates a dangerous bug - invalid inputs cause API failures instead of graceful handling")
				} else {
					t.Logf("✅ Scenario %s handled gracefully (no error)", scenario.Name)
				}
			}
		})
	}
}

// TestControllerInputValidation tests controller input validation
// This function is designed to catch dangerous bugs in controller methods
func (h *MediaMTXTestHelper) TestControllerInputValidation(t *testing.T, controller MediaMTXController) {
	ctx := context.Background()

	// Test RTSP connection input validation through controller
	t.Run("RTSP_Connections_Input_Validation", func(t *testing.T) {
		scenarios := GetRTSPInputValidationScenarios()

		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				t.Logf("Testing controller RTSP scenario: %s - %s", scenario.Name, scenario.Description)

				// Test through controller
				_, err := controller.ListRTSPConnections(ctx, scenario.Page, scenario.ItemsPerPage)

				if scenario.ExpectError {
					require.Error(t, err, "Controller scenario %s should produce an error", scenario.Name)
					if scenario.ErrorMsg != "" {
						assert.Contains(t, err.Error(), scenario.ErrorMsg,
							"Controller error message should contain expected text for scenario %s", scenario.Name)
					}
					t.Logf("✅ Controller scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the controller should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Controller scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - invalid inputs cause controller failures instead of graceful handling")
					} else {
						t.Logf("✅ Controller scenario %s handled gracefully (no error)", scenario.Name)
					}
				}
			})
		}
	})

	// Test device path validation
	t.Run("Device_Path_Validation", func(t *testing.T) {
		invalidDevicePaths := []string{
			"",                              // Empty device path
			"invalid_device",                // Invalid device path
			"/dev/video999",                 // Non-existent device
			"camera999",                     // Non-existent camera
			"../../etc/passwd",              // Path traversal attempt
			"<script>alert('xss')</script>", // XSS attempt
		}

		for _, devicePath := range invalidDevicePaths {
			t.Run(fmt.Sprintf("device_%s", devicePath), func(t *testing.T) {
				t.Logf("Testing device path validation: %s", devicePath)

				// Test various controller methods with invalid device paths
				_, err := controller.GetStreamStatus(ctx, devicePath)
				if err == nil {
					t.Errorf("🚨 BUG DETECTED: GetStreamStatus should reject invalid device path '%s'", devicePath)
				}

				_, err = controller.StartStreaming(ctx, devicePath)
				if err == nil {
					t.Errorf("🚨 BUG DETECTED: StartStreaming should reject invalid device path '%s'", devicePath)
				}

				_, err = controller.TakeAdvancedSnapshot(ctx, devicePath, "/tmp/test.jpg", map[string]interface{}{})
				if err == nil {
					t.Errorf("🚨 BUG DETECTED: TakeAdvancedSnapshot should reject invalid device path '%s'", devicePath)
				}

				t.Logf("✅ Device path '%s' correctly rejected by controller methods", devicePath)
			})
		}
	})
}

// TestInputValidationBoundaryConditions tests boundary conditions that can cause dangerous bugs
func (h *MediaMTXTestHelper) TestInputValidationBoundaryConditions(t *testing.T, controller MediaMTXController) {
	ctx := context.Background()

	t.Run("Boundary_Conditions", func(t *testing.T) {
		// Test boundary conditions that could cause integer overflow or underflow
		boundaryTests := []struct {
			name         string
			page         int
			itemsPerPage int
			description  string
		}{
			{
				name:         "min_int_page",
				page:         -2147483648, // Min int32
				itemsPerPage: 10,
				description:  "Minimum integer page should be handled gracefully",
			},
			{
				name:         "min_int_items_per_page",
				page:         0,
				itemsPerPage: -2147483648, // Min int32
				description:  "Minimum integer items per page should be handled gracefully",
			},
		}

		for _, test := range boundaryTests {
			t.Run(test.name, func(t *testing.T) {
				t.Logf("Testing boundary condition: %s - %s", test.name, test.description)

				// Test that boundary conditions don't cause panics or crashes
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("🚨 BUG DETECTED: Boundary condition %s caused panic: %v", test.name, r)
					}
				}()

				_, err := controller.ListRTSPConnections(ctx, test.page, test.itemsPerPage)
				// We don't care about the error here, just that it doesn't panic
				t.Logf("✅ Boundary condition %s handled without panic (error: %v)", test.name, err)
			})
		}
	})
}

// ============================================================================
// JSON MALFORMATION TEST HELPERS
// ============================================================================
// These helpers are designed to catch dangerous bugs through systematic
// JSON malformation testing, not just achieve coverage.

// JSONMalformationTestScenario represents a test scenario for JSON malformation
type JSONMalformationTestScenario struct {
	Name        string
	JSONData    []byte
	ExpectError bool
	ErrorMsg    string
	Description string
}

// GetJSONMalformationScenarios returns comprehensive JSON malformation scenarios
// that can catch dangerous bugs in JSON parsing functions
func GetJSONMalformationScenarios() []JSONMalformationTestScenario {
	return []JSONMalformationTestScenario{
		{
			Name:        "empty_json",
			JSONData:    []byte(""),
			ExpectError: true,
			ErrorMsg:    "empty response body",
			Description: "Empty JSON should be handled gracefully",
		},
		{
			Name:        "null_json",
			JSONData:    []byte("null"),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "Null JSON should be handled gracefully",
		},
		{
			Name:        "malformed_json",
			JSONData:    []byte(`{"invalid": json}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "Malformed JSON should be handled gracefully",
		},
		{
			Name:        "incomplete_json",
			JSONData:    []byte(`{"incomplete":`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "Incomplete JSON should be handled gracefully",
		},
		{
			Name:        "unexpected_json_structure",
			JSONData:    []byte(`{"unexpected": "structure", "not": "what we expect"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "Unexpected JSON structure should be handled gracefully",
		},
		{
			Name:        "json_with_invalid_types",
			JSONData:    []byte(`{"items": "not_an_array", "count": "not_a_number"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with invalid types should be handled gracefully",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"items": []}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with missing required fields should be handled gracefully",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "extra_field": "should_be_ignored", "another_extra": 123}`),
			ExpectError: false, // Should handle gracefully by ignoring extra fields
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
		{
			Name:        "json_with_unicode_issues",
			JSONData:    []byte(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "unicode": "test\u0000null\u0000byte"}`),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with Unicode issues should be handled gracefully",
		},
		{
			Name:        "json_with_very_large_strings",
			JSONData:    []byte(fmt.Sprintf(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "large_string": "%s"}`, strings.Repeat("x", 1000000))),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with very large strings should be handled gracefully",
		},
		{
			Name:        "json_with_deeply_nested_objects",
			JSONData:    []byte(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "nested": {"level1": {"level2": {"level3": {"level4": {"level5": "deep"}}}}}}`),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with deeply nested objects should be handled gracefully",
		},
		{
			Name:        "json_with_special_characters",
			JSONData:    []byte(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "special": "test\"quotes\"and'single'quotes\nand\tnewlines"}`),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with special characters should be handled gracefully",
		},
	}
}

// TestJSONParsingErrors tests JSON parsing functions with malformed data
// This function is designed to catch dangerous bugs, not just achieve coverage
func (h *MediaMTXTestHelper) TestJSONParsingErrors(t *testing.T) {
	scenarios := GetJSONMalformationScenarios()

	// Test parseStreamsResponse function
	t.Run("parseStreamsResponse_JSON_Errors", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				t.Logf("Testing parseStreamsResponse with scenario: %s - %s", scenario.Name, scenario.Description)

				// Test the JSON parsing function
				_, err := parseStreamsResponse(scenario.JSONData)

				if scenario.ExpectError {
					// Should get an error
					require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
					if scenario.ErrorMsg != "" {
						assert.Contains(t, err.Error(), scenario.ErrorMsg,
							"Error message should contain expected text for scenario %s", scenario.Name)
					}
					t.Logf("✅ Scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the JSON parsing should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - malformed JSON causes parsing failures instead of graceful handling")
					} else {
						t.Logf("✅ Scenario %s handled gracefully (no error)", scenario.Name)
					}
				}
			})
		}
	})

	// Test parseStreamResponse function
	t.Run("parseStreamResponse_JSON_Errors", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				t.Logf("Testing parseStreamResponse with scenario: %s - %s", scenario.Name, scenario.Description)

				// Test the JSON parsing function
				_, err := parseStreamResponse(scenario.JSONData)

				if scenario.ExpectError {
					// Should get an error
					require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
					if scenario.ErrorMsg != "" {
						assert.Contains(t, err.Error(), scenario.ErrorMsg,
							"Error message should contain expected text for scenario %s", scenario.Name)
					}
					t.Logf("✅ Scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the JSON parsing should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - malformed JSON causes parsing failures instead of graceful handling")
					} else {
						t.Logf("✅ Scenario %s handled gracefully (no error)", scenario.Name)
					}
				}
			})
		}
	})

	// Test parseHealthResponse function
	t.Run("parseHealthResponse_JSON_Errors", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				t.Logf("Testing parseHealthResponse with scenario: %s - %s", scenario.Name, scenario.Description)

				// Test the JSON parsing function
				_, err := parseHealthResponse(scenario.JSONData)

				if scenario.ExpectError {
					// Should get an error
					require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
					if scenario.ErrorMsg != "" {
						assert.Contains(t, err.Error(), scenario.ErrorMsg,
							"Error message should contain expected text for scenario %s", scenario.Name)
					}
					t.Logf("✅ Scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the JSON parsing should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - malformed JSON causes parsing failures instead of graceful handling")
					} else {
						t.Logf("✅ Scenario %s handled gracefully (no error)", scenario.Name)
					}
				}
			})
		}
	})

	// Test parsePathsResponse function
	t.Run("parsePathsResponse_JSON_Errors", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				t.Logf("Testing parsePathsResponse with scenario: %s - %s", scenario.Name, scenario.Description)

				// Test the JSON parsing function
				_, err := parsePathsResponse(scenario.JSONData)

				if scenario.ExpectError {
					// Should get an error
					require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
					if scenario.ErrorMsg != "" {
						assert.Contains(t, err.Error(), scenario.ErrorMsg,
							"Error message should contain expected text for scenario %s", scenario.Name)
					}
					t.Logf("✅ Scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the JSON parsing should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - malformed JSON causes parsing failures instead of graceful handling")
					} else {
						t.Logf("✅ Scenario %s handled gracefully (no error)", scenario.Name)
					}
				}
			})
		}
	})
}

// TestJSONParsingPanicProtection tests that JSON parsing functions don't panic
// This function is designed to catch dangerous bugs that could cause crashes
func (h *MediaMTXTestHelper) TestJSONParsingPanicProtection(t *testing.T) {
	// Test data that could cause panics
	panicTestData := [][]byte{
		[]byte(`{"items": [{"name": null}]}`),             // Null values in arrays
		[]byte(`{"items": [{"name": {"nested": null}}]}`), // Nested null values
		[]byte(`{"items": [{"name": []}]}`),               // Arrays instead of strings
		[]byte(`{"items": [{"name": {}}]}`),               // Objects instead of strings
		[]byte(`{"items": [{"name": 123}]}`),              // Numbers instead of strings
		[]byte(`{"items": [{"name": true}]}`),             // Booleans instead of strings
	}

	t.Run("Panic_Protection_Tests", func(t *testing.T) {
		for i, data := range panicTestData {
			t.Run(fmt.Sprintf("panic_test_%d", i), func(t *testing.T) {
				t.Logf("Testing panic protection with data: %s", string(data))

				// Test that functions don't panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("🚨 BUG DETECTED: JSON parsing caused panic with data %s: %v", string(data), r)
					}
				}()

				// Test all parsing functions
				_, err1 := parseStreamsResponse(data)
				_, err2 := parseStreamResponse(data)
				_, err3 := parseHealthResponse(data)
				_, err4 := parsePathsResponse(data)

				// We don't care about errors here, just that no panic occurred
				t.Logf("✅ No panic occurred (errors: %v, %v, %v, %v)", err1, err2, err3, err4)
			})
		}
	})
}
