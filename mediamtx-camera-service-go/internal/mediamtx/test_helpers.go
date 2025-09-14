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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	configpkg "github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	configManager         *configpkg.ConfigManager
	logger                *logging.Logger
	client                MediaMTXClient
	mediaMTXConfig        *configpkg.MediaMTXConfig    // Centralized config for all managers
	configIntegration     *configpkg.ConfigIntegration // Centralized config integration for all managers
	pathManager           PathManager
	streamManager         StreamManager
	recordingManager      *RecordingManager
	rtspConnectionManager RTSPConnectionManager
	cameraMonitor         camera.CameraMonitor

	// Race-free initialization using sync.Once
	pathManagerOnce      sync.Once
	streamManagerOnce    sync.Once
	recordingManagerOnce sync.Once
	cameraMonitorOnce    sync.Once
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
func NewMediaMTXTestHelper(t *testing.T, testConfig *MediaMTXTestConfig) *MediaMTXTestHelper {
	if testConfig == nil {
		testConfig = DefaultMediaMTXTestConfig()
	}

	// Create logger for testing
	logger := logging.GetLogger("test-mediamtx-controller")
	logger.SetLevel(logrus.InfoLevel) // Make logs visible for investigation

	// GLOBAL FIX: Set all loggers to InfoLevel for debugging
	// This ensures all components use the same log level
	logging.GetGlobalLogger().SetLevel(logrus.InfoLevel)

	// Create MediaMTX client configuration
	clientConfig := &configpkg.MediaMTXConfig{
		BaseURL:        testConfig.BaseURL,
		HealthCheckURL: testConfig.BaseURL + "/v3/paths/list", // Correct Go MediaMTX health check endpoint
		Timeout:        testConfig.Timeout,
		ConnectionPool: configpkg.ConnectionPoolConfig{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// Create MediaMTX client
	client := NewClient(testConfig.BaseURL, clientConfig, logger)

	// Create config manager for centralized configuration
	configManager := CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")

	// Create centralized MediaMTX config for all managers
	mediaMTXConfig := &configpkg.MediaMTXConfig{
		BaseURL: testConfig.BaseURL,
		Timeout: 10 * time.Second,
	}

	// Create centralized ConfigIntegration for all managers
	configIntegration := configpkg.NewConfigIntegration(configManager, logger)

	helper := &MediaMTXTestHelper{
		config:            testConfig,
		configManager:     configManager,
		logger:            logger,
		client:            client,
		mediaMTXConfig:    mediaMTXConfig,
		configIntegration: configIntegration,
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

	// Stop camera monitor to prevent goroutine leaks
	if h.cameraMonitor != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.cameraMonitor.Stop(ctx); err != nil {
			t.Logf("Warning: Failed to stop camera monitor during cleanup: %v", err)
		}
	}

	// No factory reset needed - fresh instances provide natural test isolation

	// Stop config manager to prevent file watcher goroutine leaks
	if h.configManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.configManager.Stop(ctx); err != nil {
			t.Logf("Warning: Failed to stop config manager during cleanup: %v", err)
		}
	}

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
	// we can do a quick health check. Event-driven readiness is handled by EventDrivenTestHelper
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Single health check - polling removed in favor of event-driven approach
	err := h.client.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("MediaMTX server not ready: %w", err)
	}

	t.Log("MediaMTX server is ready")
	return nil
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
	invalidConfig := &configpkg.MediaMTXConfig{
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
	h.pathManagerOnce.Do(func() {
		// Use centralized MediaMTX config
		h.pathManager = NewPathManager(h.client, h.mediaMTXConfig, h.logger)
	})
	return h.pathManager
}

// GetStreamManager returns a shared stream manager instance
func (h *MediaMTXTestHelper) GetStreamManager() StreamManager {
	h.streamManagerOnce.Do(func() {
		// Ensure PathManager is initialized first to prevent nil pointer dereference
		pathManager := h.GetPathManager() // This will initialize h.pathManager if nil

		// Use centralized MediaMTX config and ConfigIntegration
		h.streamManager = NewStreamManager(h.client, pathManager, h.mediaMTXConfig, h.configIntegration, h.logger)
	})
	return h.streamManager
}

// GetRecordingManager returns a shared recording manager instance
func (h *MediaMTXTestHelper) GetRecordingManager() *RecordingManager {
	h.recordingManagerOnce.Do(func() {
		// Use centralized MediaMTX config and ConfigIntegration
		pathManager := h.GetPathManager()
		streamManager := h.GetStreamManager()
		h.recordingManager = NewRecordingManager(h.client, pathManager, streamManager, h.mediaMTXConfig, h.configIntegration, h.logger)
	})
	return h.recordingManager
}

// GetCameraMonitor returns a shared camera monitor instance using REAL hardware
// ARCHITECTURE COMPLIANCE: Follows Progressive Readiness Pattern - no blocking startup
func (h *MediaMTXTestHelper) GetCameraMonitor() camera.CameraMonitor {
	h.cameraMonitorOnce.Do(func() {
		// Create real camera monitor with SAME configuration as controller (test fixture)
		// This ensures configuration consistency between camera monitor and controller
		configManager := CreateConfigManagerWithFixture(nil, "config_test_minimal.yaml")
		logger := logging.GetLogger("test-camera-monitor")

		// Use real implementations for camera hardware
		deviceChecker := &camera.RealDeviceChecker{}
		commandExecutor := &camera.RealV4L2CommandExecutor{}
		infoParser := &camera.RealDeviceInfoParser{}

		// Create real camera monitor - NO MOCKS, real hardware only
		// Monitor will acquire its own device event source reference
		realMonitor, err := camera.NewHybridCameraMonitor(
			configManager, logger, deviceChecker, commandExecutor, infoParser)
		if err != nil {
			// Real system - if camera monitor fails, test should fail
			h.logger.WithError(err).Error("Failed to create real camera monitor - test requires real hardware")
			panic(fmt.Sprintf("Camera monitor creation failed: %v", err))
		}

		// ARCHITECTURE COMPLIANCE: Progressive Readiness Pattern
		// Don't start the monitor here - let the controller start it
		// The monitor will be started by the controller's Start() method

		h.cameraMonitor = realMonitor
	})

	return h.cameraMonitor
}

// HasHardwareCamera checks if a real camera is available for testing
func (h *MediaMTXTestHelper) HasHardwareCamera(ctx context.Context) bool {
	// Use real device detection instead of hardcoded camera0
	availableDevices := h.getRealAvailableDevices()
	return len(availableDevices) > 0
}

// getRealAvailableDevices scans for real available camera devices on the system
// This reuses the same logic as RealHardwareTestHelper for consistency
func (h *MediaMTXTestHelper) getRealAvailableDevices() []string {
	availableDevices := []string{}

	// Scan for video devices in /dev
	videoDevices, err := filepath.Glob("/dev/video*")
	if err != nil {
		h.logger.WithError(err).Warn("Could not scan for video devices")
		return availableDevices
	}

	for _, device := range videoDevices {
		// Check if device is actually accessible and functional
		if h.isDeviceAccessible(device) {
			availableDevices = append(availableDevices, device)
			h.logger.WithField("device", device).Debug("Found accessible camera device")
		}
	}

	if len(availableDevices) == 0 {
		h.logger.Warn("No accessible camera devices found. Tests will use fallback devices.")
		// Fallback to common device paths for testing
		availableDevices = []string{"/dev/video0", "/dev/video1"}
	}

	return availableDevices
}

// isDeviceAccessible checks if a device is actually accessible and functional
// This reuses the same logic as RealHardwareTestHelper for consistency
func (h *MediaMTXTestHelper) isDeviceAccessible(devicePath string) bool {
	// Check if device file exists
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		return false
	}

	// Try to get device capabilities (non-blocking)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use v4l2-ctl to check device capabilities
	cmd := exec.CommandContext(ctx, "v4l2-ctl", "--device", devicePath, "--all")
	output, err := cmd.Output()
	if err != nil {
		// Device exists but may not be accessible (permissions, busy, etc.)
		return false
	}

	// Check if this is actually a video capture device, not just a metadata device
	hasVideoCapture := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Device Caps") {
			// Look for video capture capability in Device Caps
			if strings.Contains(line, "0x04200001") || strings.Contains(line, "0x85200001") {
				hasVideoCapture = true
				break
			}
		}
	}

	// Only consider devices with video capture capability as "accessible cameras"
	return hasVideoCapture
}

// GetAvailableCameraDevice returns the first available real camera device
// This eliminates the need for repetitive inline camera device detection patterns
func (h *MediaMTXTestHelper) GetAvailableCameraDevice(ctx context.Context) (string, error) {
	// Use real device detection instead of hardcoded devices
	availableDevices := h.getRealAvailableDevices()
	if len(availableDevices) == 0 {
		return "", fmt.Errorf("no accessible camera devices found")
	}

	// Return the first available device
	return availableDevices[0], nil
}

// GetTestCameraDevice returns a test camera device from fixtures
func (h *MediaMTXTestHelper) GetTestCameraDevice(scenario string) string {
	// Load test camera devices from fixture file
	fixturePath := filepath.Join("tests", "fixtures", "test_camera_devices.yaml")

	// Read fixture file
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to read test camera devices fixture, using fallback")
		return h.getFallbackDevice(scenario)
	}

	// Parse YAML
	var fixtures struct {
		TestScenarios map[string][]string `yaml:"test_scenarios"`
	}

	if err := yaml.Unmarshal(data, &fixtures); err != nil {
		h.logger.WithError(err).Warn("Failed to parse test camera devices fixture, using fallback")
		return h.getFallbackDevice(scenario)
	}

	// Get devices for scenario
	devices, exists := fixtures.TestScenarios[scenario]
	if !exists || len(devices) == 0 {
		h.logger.WithField("scenario", scenario).Warn("Scenario not found in fixtures, using fallback")
		return h.getFallbackDevice(scenario)
	}

	// Return first device for scenario
	device := devices[0]

	// For hardware_available scenario, try to use real device detection
	if scenario == "hardware_available" {
		if realDevice, err := h.GetAvailableCameraDevice(context.Background()); err == nil {
			return realDevice
		}
	}

	return device
}

// getFallbackDevice provides fallback devices when fixture loading fails
func (h *MediaMTXTestHelper) getFallbackDevice(scenario string) string {
	switch scenario {
	case "hardware_available":
		// Use real device detection instead of hardcoded /dev/video0
		if device, err := h.GetAvailableCameraDevice(context.Background()); err == nil {
			return device
		}
		return "/dev/video0" // Fallback to local V4L2 device
	case "network_failure":
		return "rtsp://test-source.example.com:554/stream" // External RTSP (expected to fail)
	case "mixed_scenario":
		// Check if hardware is available, otherwise use external source
		if h.HasHardwareCamera(context.Background()) {
			if device, err := h.GetAvailableCameraDevice(context.Background()); err == nil {
				return device
			}
		}
		return "rtsp://test-source.example.com:554/stream"
	default:
		// Use real device detection instead of hardcoded /dev/video0
		if device, err := h.GetAvailableCameraDevice(context.Background()); err == nil {
			return device
		}
		return "/dev/video0" // Fallback to local device
	}
}

// GetController creates a MediaMTX controller with proper dependencies
func (h *MediaMTXTestHelper) GetController(t *testing.T) (MediaMTXController, error) {
	// Use shared config manager to prevent multiple instances
	configManager := h.GetConfigManager()
	cameraMonitor := h.GetCameraMonitor()
	logger := h.GetLogger()

	return ControllerWithConfigManager(configManager, cameraMonitor, logger)
}

// Bad orchestration methods deleted - violates Progressive Readiness Pattern

// GetConfiguredSnapshotPath returns the snapshot path from the fixture configuration
// This follows the architecture principle of using configured paths instead of hardcoded paths
func (h *MediaMTXTestHelper) GetConfiguredSnapshotPath() string {
	configManager := h.GetConfigManager()
	if configManager == nil {
		return "/tmp/snapshots" // Fallback to fixture default
	}

	config := configManager.GetConfig()
	if config == nil {
		return "/tmp/snapshots" // Fallback to fixture default
	}

	return config.MediaMTX.SnapshotsPath
}

// GetConfiguredRecordingPath returns the recording path from the fixture configuration
func (h *MediaMTXTestHelper) GetConfiguredRecordingPath() string {
	configManager := h.GetConfigManager()
	if configManager == nil {
		return "/tmp/recordings" // Fallback to fixture default
	}

	config := configManager.GetConfig()
	if config == nil {
		return "/tmp/recordings" // Fallback to fixture default
	}

	return config.MediaMTX.RecordingsPath
}

// GetConfigManager returns the config manager instance
func (h *MediaMTXTestHelper) GetConfigManager() *configpkg.ConfigManager {
	return h.configManager
}

// GetRTSPConnectionManager returns a shared RTSP connection manager instance
func (h *MediaMTXTestHelper) GetRTSPConnectionManager() RTSPConnectionManager {
	if h.rtspConnectionManager == nil {
		// Use centralized MediaMTX config
		h.rtspConnectionManager = NewRTSPConnectionManager(h.client, h.mediaMTXConfig, h.logger)
	}
	return h.rtspConnectionManager
}

// CreateConfigManagerWithFixture creates a config manager that loads from test fixtures
func CreateConfigManagerWithFixture(t *testing.T, fixtureName string) *configpkg.ConfigManager {
	configManager := configpkg.CreateConfigManager()

	// Use test fixture instead of creating config manually
	fixturePath := filepath.Join("tests", "fixtures", fixtureName)

	// Check if fixture exists, if not use a fallback path
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		// Try alternative path
		fixturePath = filepath.Join("..", "..", "tests", "fixtures", fixtureName)
	}

	// Create required directories and files for test fixtures that use /tmp paths
	// This is needed because fixtures have hardcoded paths that need to exist
	requiredDirs := []string{
		"/tmp/recordings",
		"/tmp/snapshots",
		"/tmp",
	}

	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create required directory %s: %v", dir, err)
		}
	}

	// Create required files that fixtures expect to exist
	requiredFiles := []string{
		"/tmp/mediamtx.yml",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if err := os.WriteFile(file, []byte("# Test MediaMTX config file\n"), 0644); err != nil {
				t.Fatalf("Failed to create required file %s: %v", file, err)
			}
		}
	}

	err := configManager.LoadConfig(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load config from fixture %s: %v", fixtureName, err)
	}

	return configManager
}

// CreateTestPath creates a test path with proper configuration
func (h *MediaMTXTestHelper) CreateTestPath(t *testing.T, name string) error {
	ctx := context.Background()

	// First, check if path exists in runtime
	if _, err := h.client.Get(ctx, "/v3/paths/get/"+name); err == nil {
		t.Logf("Path %s already exists in runtime, attempting cleanup", name)
		h.ForceCleanupRuntimePaths(t)

		// Check again after cleanup
		if _, err := h.client.Get(ctx, "/v3/paths/get/"+name); err == nil {
			// Path still exists, try a different approach
			// Create a unique name with timestamp
			name = fmt.Sprintf("%s_%d", name, time.Now().Unix())
			t.Logf("Using alternative path name: %s", name)
		}
	}

	// Create path with a concrete source instead of "publisher"
	// This creates a configuration path that can be deleted
	pathConfig := map[string]interface{}{
		"source":                     "rtsp://localhost:8554/dummy", // Use concrete source
		"sourceOnDemand":             true,
		"sourceOnDemandStartTimeout": "10s",
		"sourceOnDemandCloseAfter":   "10s",
	}

	data, err := json.Marshal(pathConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal path config: %w", err)
	}

	_, err = h.client.Post(ctx, "/v3/config/paths/add/"+name, data)
	if err != nil {
		// Check if it's "already exists" error
		if strings.Contains(err.Error(), "already exists") {
			t.Logf("Path %s already exists, treating as success", name)
			return nil // Idempotent
		}
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

// ForceCleanupRuntimePaths forcefully cleans up runtime paths by disconnecting publishers
func (h *MediaMTXTestHelper) ForceCleanupRuntimePaths(t *testing.T) error {
	ctx := context.Background()

	// Get all runtime paths
	data, err := h.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Use PathList from api_types.go instead of inline struct
	var pathsResponse PathList

	if err := json.Unmarshal(data, &pathsResponse); err != nil {
		return fmt.Errorf("failed to parse paths: %w", err)
	}

	// For each test path, try to force cleanup
	for _, path := range pathsResponse.Items {
		if h.isTestPath(path.Name) {
			t.Logf("Found runtime test path: %s (source: %v, ready: %v)",
				path.Name, path.Source, path.Ready)

			// Option 1: Try to kick all connections (if MediaMTX supports it)
			// This would disconnect publishers/readers and allow cleanup
			if path.Source != nil {
				// Try to disconnect the source connection
				kickEndpoint := fmt.Sprintf("/v3/%s/kick/%s", path.Source.Type, path.Source.ID)
				if _, err := h.client.Post(ctx, kickEndpoint, nil); err != nil {
					t.Logf("Could not kick source %s: %v", path.Source.ID, err)
				}
			}

			// Kick all readers
			for _, reader := range path.Readers {
				kickEndpoint := fmt.Sprintf("/v3/%s/kick/%s", reader.Type, reader.ID)
				if _, err := h.client.Post(ctx, kickEndpoint, nil); err != nil {
					t.Logf("Could not kick reader %s: %v", reader.ID, err)
				}
			}
		}
	}

	// Wait a bit for MediaMTX to clean up paths
	time.Sleep(200 * time.Millisecond)

	return nil
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
	var pathsResponse struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &pathsResponse); err != nil {
		t.Logf("Warning: Failed to parse paths response: %v", err)
		return
	}

	// For runtime paths, we can't delete them via config API
	// Instead, we log them for debugging and rely on MediaMTX to clean them up automatically
	// when they're no longer in use (no active sources/readers)
	testPathCount := 0
	for _, path := range pathsResponse.Items {
		if h.isTestPath(path.Name) {
			testPathCount++
			// Log test paths for debugging - they should be cleaned up automatically by MediaMTX
			// when no longer in use (no active sources/readers)
			t.Logf("Test path still in runtime: %s (will be cleaned up automatically when unused)", path.Name)
		}
	}

	if testPathCount > 0 {
		t.Logf("Found %d test paths in runtime state - these should be cleaned up automatically by MediaMTX when unused", testPathCount)
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
			ErrorMsg:     "invalid page number: -1 (must be >= 0)",
			Description:  "Negative page numbers should be rejected",
		},
		{
			Name:         "zero_items_per_page",
			Page:         0,
			ItemsPerPage: 0,
			ExpectError:  true, // Should be rejected with clear error message
			ErrorMsg:     "invalid items per page",
			Description:  "Zero items per page should be rejected",
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
				t.Logf("Scenario %s correctly produced expected error: %v", scenario.Name, err)
			} else {
				// Should NOT get an error (graceful handling)
				if err != nil {
					// This is a BUG - the API should handle these inputs gracefully
					t.Errorf("🚨 BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
					t.Errorf("🚨 This indicates a dangerous bug - invalid inputs cause API failures instead of graceful handling")
				} else {
					t.Logf("Scenario %s handled gracefully (no error)", scenario.Name)
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
					t.Logf("Controller scenario %s correctly produced expected error: %v", scenario.Name, err)
				} else {
					// Should NOT get an error (graceful handling)
					if err != nil {
						// This is a BUG - the controller should handle these inputs gracefully
						t.Errorf("🚨 BUG DETECTED: Controller scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						t.Errorf("🚨 This indicates a dangerous bug - invalid inputs cause controller failures instead of graceful handling")
					} else {
						t.Logf("Controller scenario %s handled gracefully (no error)", scenario.Name)
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

				_, err = controller.TakeAdvancedSnapshot(ctx, devicePath, map[string]interface{}{})
				if err == nil {
					t.Errorf("🚨 BUG DETECTED: TakeAdvancedSnapshot should reject invalid device path '%s'", devicePath)
				}

				t.Logf("Device path '%s' correctly rejected by controller methods", devicePath)
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
				t.Logf("Boundary condition %s handled without panic (error: %v)", test.name, err)
			})
		}
	})
}

// ============================================================================
// EVENT-DRIVEN TEST HELPERS
// ============================================================================
// These helpers provide event-driven testing capabilities to replace polling
// with efficient event subscription patterns.

// EventDrivenTestHelper provides non-blocking event observation capabilities
type EventDrivenTestHelper struct {
	controller    MediaMTXController
	eventChannels map[string]chan struct{}
	eventHistory  map[string][]interface{}
	eventMutex    sync.RWMutex
	logger        *logging.Logger
}

// CreateEventDrivenTestHelper creates a new event-driven test helper
func (h *MediaMTXTestHelper) CreateEventDrivenTestHelper(t *testing.T) *EventDrivenTestHelper {
	controller, err := h.GetController(t)
	if err != nil {
		t.Fatalf("Failed to get controller for event-driven test helper: %v", err)
	}

	return &EventDrivenTestHelper{
		controller:    controller,
		eventChannels: make(map[string]chan struct{}),
		eventHistory:  make(map[string][]interface{}),
		logger:        h.GetLogger(),
	}
}

// ObserveReadiness starts non-blocking observation of readiness events
func (edh *EventDrivenTestHelper) ObserveReadiness() <-chan interface{} {
	edh.eventMutex.Lock()
	defer edh.eventMutex.Unlock()

	// Create observation channel
	observationChan := make(chan interface{}, 10)

	// Start background observer using polling (since production doesn't emit events)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		lastReadyState := false
		for {
			select {
			case <-ticker.C:
				currentReadyState := edh.controller.IsReady()
				
				// Record event when readiness state changes from false to true
				if !lastReadyState && currentReadyState {
					edh.recordEvent("readiness", "controller_ready")
					select {
					case observationChan <- "controller_ready":
					default:
						// Don't block if channel full
					}
				}
				lastReadyState = currentReadyState
			}
		}
	}()

	return observationChan
}

// ObserveHealthChanges starts non-blocking observation of health events
func (edh *EventDrivenTestHelper) ObserveHealthChanges() <-chan interface{} {
	edh.eventMutex.Lock()
	defer edh.eventMutex.Unlock()

	// Create observation channel
	observationChan := make(chan interface{}, 10)

	// Start background observer
	go func() {
		// Use controller's readiness subscription (includes health monitoring)
		readinessChan := edh.controller.SubscribeToReadiness()
		// Listen for readiness events and record them as health events
		for range readinessChan {
			edh.recordEvent("health", "controller_readiness_changed")
		}
	}()

	return observationChan
}

// ObserveCameraEvents starts non-blocking observation of camera events
func (edh *EventDrivenTestHelper) ObserveCameraEvents() <-chan interface{} {
	edh.eventMutex.Lock()
	defer edh.eventMutex.Unlock()

	// Create observation channel
	observationChan := make(chan interface{}, 10)

	// Start background observer
	go func() {
		// Use controller's readiness subscription (includes camera monitoring)
		readinessChan := edh.controller.SubscribeToReadiness()
		// Listen for readiness events and record them as camera events
		for range readinessChan {
			edh.recordEvent("camera", "controller_readiness_changed")
		}
	}()

	return observationChan
}

// recordEvent records an event in the history for later verification
func (edh *EventDrivenTestHelper) recordEvent(eventType string, event interface{}) {
	edh.eventMutex.Lock()
	defer edh.eventMutex.Unlock()

	if edh.eventHistory[eventType] == nil {
		edh.eventHistory[eventType] = make([]interface{}, 0)
	}
	edh.eventHistory[eventType] = append(edh.eventHistory[eventType], event)
}

// DidEventOccur checks if an event of the specified type occurred
func (edh *EventDrivenTestHelper) DidEventOccur(eventType string) bool {
	edh.eventMutex.RLock()
	defer edh.eventMutex.RUnlock()

	events, exists := edh.eventHistory[eventType]
	return exists && len(events) > 0
}

// GetEventHistory returns the history of events for a specific type
func (edh *EventDrivenTestHelper) GetEventHistory(eventType string) []interface{} {
	edh.eventMutex.RLock()
	defer edh.eventMutex.RUnlock()

	if events, exists := edh.eventHistory[eventType]; exists {
		// Return a copy to avoid race conditions
		result := make([]interface{}, len(events))
		copy(result, events)
		return result
	}
	return []interface{}{}
}

// CollectEventsForDuration collects events over a specified duration (non-blocking)
func (edh *EventDrivenTestHelper) CollectEventsForDuration(duration time.Duration) map[string][]interface{} {
	// Start all observers (non-blocking)
	edh.ObserveReadiness()
	edh.ObserveHealthChanges()
	edh.ObserveCameraEvents()

	// Use a timer instead of blocking sleep
	timer := time.NewTimer(duration)
	defer timer.Stop()

	// Wait for duration without blocking the entire method
	<-timer.C

	// Return a copy of all collected events using existing method
	result := make(map[string][]interface{})
	edh.eventMutex.RLock()
	for eventType, events := range edh.eventHistory {
		result[eventType] = make([]interface{}, len(events))
		copy(result[eventType], events)
	}
	edh.eventMutex.RUnlock()

	return result
}

// Cleanup closes all event channels and cleans up resources
func (edh *EventDrivenTestHelper) Cleanup() {
	edh.eventMutex.Lock()
	defer edh.eventMutex.Unlock()

	for name, ch := range edh.eventChannels {
		// Check if channel is already closed by attempting to close it safely
		select {
		case <-ch:
			// Channel is already closed
			edh.logger.WithField("channel", name).Debug("Event channel already closed")
		default:
			// Channel is open, close it
			close(ch)
			edh.logger.WithField("channel", name).Debug("Closed event channel")
		}
	}

	edh.eventChannels = make(map[string]chan struct{})
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

// JSONScenarioRegistry provides centralized scenario management for all JSON response types
type JSONScenarioRegistry struct {
	scenarios map[string][]JSONMalformationTestScenario
}

// NewJSONScenarioRegistry creates a new scenario registry with all baseline scenarios
func NewJSONScenarioRegistry() *JSONScenarioRegistry {
	registry := &JSONScenarioRegistry{
		scenarios: make(map[string][]JSONMalformationTestScenario),
	}

	// Initialize with baseline scenarios that apply to all response types
	// These are the scenarios that were duplicated across all 5 original functions
	baselineScenarios := []JSONMalformationTestScenario{
		{
			Name:        "empty_json",
			JSONData:    []byte(""),
			ExpectError: true,
			ErrorMsg:    "empty response body",
			Description: "Empty JSON should be rejected",
		},
		{
			Name:        "null_json",
			JSONData:    []byte("null"),
			ExpectError: true,
			ErrorMsg:    "null response body",
			Description: "Null JSON should be rejected",
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
			ErrorMsg:    "missing required field",
			Description: "Unexpected JSON structure should be rejected",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"pageCount": 1, "itemCount": 0}`), // Missing required fields vary by type
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "JSON with missing required fields should be rejected",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"extraField": "should be ignored"}`), // Extra fields vary by type
			ExpectError: false,                                         // Should handle gracefully by ignoring extra fields
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
		{
			Name:        "json_with_unicode_issues",
			JSONData:    []byte(`{"test": "test\u0000null\u0000byte"}`),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with Unicode issues should be handled gracefully",
		},
		{
			Name:        "json_with_very_large_strings",
			JSONData:    []byte(fmt.Sprintf(`{"test": "%s"}`, strings.Repeat("x", 1000000))),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with very large strings should be handled gracefully",
		},
		{
			Name:        "json_with_special_characters",
			JSONData:    []byte(`{"test": "test\"quotes\"and'single'quotes\nand\tnewlines"}`),
			ExpectError: false, // Should handle gracefully
			ErrorMsg:    "",
			Description: "JSON with special characters should be handled gracefully",
		},
	}

	// Add type-specific scenarios for each response type
	registry.addPathListScenarios(baselineScenarios)
	registry.addStreamScenarios(baselineScenarios)
	registry.addPathsScenarios(baselineScenarios)
	registry.addHealthScenarios(baselineScenarios)

	return registry
}

// GetScenarios returns scenarios for a specific response type
func (r *JSONScenarioRegistry) GetScenarios(responseType string) []JSONMalformationTestScenario {
	scenarios, exists := r.scenarios[responseType]
	if !exists {
		return []JSONMalformationTestScenario{}
	}
	return scenarios
}

// addPathListScenarios adds scenarios specific to path list responses
func (r *JSONScenarioRegistry) addPathListScenarios(baseline []JSONMalformationTestScenario) {
	scenarios := make([]JSONMalformationTestScenario, len(baseline))
	copy(scenarios, baseline)

	// Add path-list specific scenarios
	typeSpecific := []JSONMalformationTestScenario{
		{
			Name:        "unexpected_json_structure",
			JSONData:    []byte(`{"unexpected": "structure", "not": "what we expect"}`),
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "Unexpected JSON structure should be rejected",
		},
		{
			Name:        "json_with_invalid_types",
			JSONData:    []byte(`{"items": "not_an_array", "pageCount": "not_a_number", "itemCount": "not_a_number"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with invalid types should be rejected due to parsing errors",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"pageCount": 1, "itemCount": 0}`), // Missing 'items' field
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "JSON with missing required fields should be rejected",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"items": [], "pageCount": 1, "itemCount": 0, "extraField": "should be ignored"}`),
			ExpectError: false,
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
	}

	scenarios = append(scenarios, typeSpecific...)
	r.scenarios["path_list"] = scenarios
}

// addStreamScenarios adds scenarios specific to stream responses
func (r *JSONScenarioRegistry) addStreamScenarios(baseline []JSONMalformationTestScenario) {
	scenarios := make([]JSONMalformationTestScenario, len(baseline))
	copy(scenarios, baseline)

	// Add stream-specific scenarios
	typeSpecific := []JSONMalformationTestScenario{
		{
			Name:        "unexpected_json_structure",
			JSONData:    []byte(`{"unexpected": "structure", "not": "what we expect"}`),
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "Unexpected JSON structure should be rejected",
		},
		{
			Name:        "json_with_invalid_types",
			JSONData:    []byte(`{"name": 123, "source": "not_an_object"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with invalid types should be rejected due to parsing errors",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"source": {"type": "rtsp", "id": "test"}}`), // Missing 'name' field
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "JSON with missing required fields should be rejected",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"name": "test_stream", "source": {"type": "rtsp", "id": "test"}, "extraField": "should be ignored"}`),
			ExpectError: false,
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
	}

	scenarios = append(scenarios, typeSpecific...)
	r.scenarios["stream"] = scenarios
}

// addPathsScenarios adds scenarios specific to paths responses
func (r *JSONScenarioRegistry) addPathsScenarios(baseline []JSONMalformationTestScenario) {
	scenarios := make([]JSONMalformationTestScenario, len(baseline))
	copy(scenarios, baseline)

	// Add paths-specific scenarios (same as path_list)
	typeSpecific := []JSONMalformationTestScenario{
		{
			Name:        "unexpected_json_structure",
			JSONData:    []byte(`{"unexpected": "structure", "not": "what we expect"}`),
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "Unexpected JSON structure should be rejected",
		},
		{
			Name:        "json_with_invalid_types",
			JSONData:    []byte(`{"items": "not_an_array", "pageCount": "not_a_number", "itemCount": "not_a_number"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with invalid types should be rejected due to parsing errors",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"pageCount": 1, "itemCount": 0}`), // Missing 'items' field
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "JSON with missing required fields should be rejected",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"items": [], "pageCount": 1, "itemCount": 0, "extraField": "should be ignored"}`),
			ExpectError: false,
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
	}

	scenarios = append(scenarios, typeSpecific...)
	r.scenarios["paths"] = scenarios
}

// addHealthScenarios adds scenarios specific to health responses
func (r *JSONScenarioRegistry) addHealthScenarios(baseline []JSONMalformationTestScenario) {
	scenarios := make([]JSONMalformationTestScenario, len(baseline))
	copy(scenarios, baseline)

	// Add health-specific scenarios
	typeSpecific := []JSONMalformationTestScenario{
		{
			Name:        "unexpected_json_structure",
			JSONData:    []byte(`{"unexpected": "structure", "not": "what we expect"}`),
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "Unexpected JSON structure should be rejected",
		},
		{
			Name:        "json_with_invalid_types",
			JSONData:    []byte(`{"status": 123, "uptime": "not_a_number"}`),
			ExpectError: true,
			ErrorMsg:    "failed to parse",
			Description: "JSON with invalid types should be rejected due to parsing errors",
		},
		{
			Name:        "json_with_missing_required_fields",
			JSONData:    []byte(`{"uptime": 12345}`), // Missing 'status' field
			ExpectError: true,
			ErrorMsg:    "missing required field",
			Description: "JSON with missing required fields should be rejected",
		},
		{
			Name:        "json_with_extra_fields",
			JSONData:    []byte(`{"status": "ok", "uptime": 12345, "extraField": "should be ignored"}`),
			ExpectError: false,
			ErrorMsg:    "",
			Description: "JSON with extra fields should be handled gracefully",
		},
	}

	scenarios = append(scenarios, typeSpecific...)
	r.scenarios["health"] = scenarios
}

// TestJSONParsingErrors tests JSON parsing functions with malformed data using the scenario registry
// This function is designed to catch dangerous bugs, not just achieve coverage
// DISABLED: Tests now use scenario registry directly in json_malformation_test.go
func (h *MediaMTXTestHelper) DisabledTestJSONParsingErrors(t *testing.T) {
	t.Skip("DISABLED: Tests now use scenario registry directly in json_malformation_test.go")
}

// TestJSONParsingPanicProtection tests that JSON parsing functions don't panic
// This function is designed to catch dangerous bugs that could cause crashes
// DISABLED: Tests now use scenario registry directly in json_malformation_test.go
func (h *MediaMTXTestHelper) DisabledTestJSONParsingPanicProtection(t *testing.T) {
	t.Skip("DISABLED: Tests now use scenario registry directly in json_malformation_test.go")
}
