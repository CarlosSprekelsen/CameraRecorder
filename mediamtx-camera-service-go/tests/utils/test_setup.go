// +build unit

//go:build unit
// +build unit

/*
Test Setup Utilities

Requirements Coverage:
- REQ-TEST-001: Test environment setup
- REQ-TEST-002: Test data preparation
- REQ-TEST-003: Test configuration management
- REQ-TEST-004: Test authentication setup
- REQ-TEST-005: Test evidence collection
- REQ-TEST-006: Real MediaMTX controller setup
- REQ-TEST-007: Test-specific MediaMTX configuration

Test Categories: Unit/Integration/Test Infrastructure
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package utils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/require"
)

// TestEnvironment provides test environment setup and management
type TestEnvironment struct {
	ConfigManager *config.ConfigManager
	Logger        *logging.Logger
	TempDir       string
	ConfigPath    string
}

// MediaMTXTestEnvironment provides MediaMTX-specific test environment
type MediaMTXTestEnvironment struct {
	*TestEnvironment
	Controller mediamtx.MediaMTXController
}

// SetupTestEnvironment creates a proper test environment with configuration
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "camera-service-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Copy test configuration to temp directory
	configPath := filepath.Join(tempDir, "test_config.yaml")
	err = copyTestConfig(configPath)
	require.NoError(t, err, "Failed to copy test configuration")

	// Initialize configuration manager
	configManager := config.NewConfigManager()
	err = configManager.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load test configuration")

	// Initialize logger
	logger := logging.NewLogger("test-environment")

	// Create test directories
	err = createTestDirectories(tempDir)
	require.NoError(t, err, "Failed to create test directories")

	return &TestEnvironment{
		ConfigManager: configManager,
		Logger:        logger,
		TempDir:       tempDir,
		ConfigPath:    configPath,
	}
}

// SetupMediaMTXTestEnvironment creates a test environment with real MediaMTX controller
func SetupMediaMTXTestEnvironment(t *testing.T) *MediaMTXTestEnvironment {
	// Setup base test environment
	baseEnv := SetupTestEnvironment(t)

	// Create real MediaMTX controller
	controller, err := mediamtx.NewControllerWithConfigManager(baseEnv.ConfigManager, baseEnv.Logger.Logger)
	require.NoError(t, err, "Failed to create real MediaMTX controller")

	return &MediaMTXTestEnvironment{
		TestEnvironment: baseEnv,
		Controller:      controller,
	}
}

// WebSocketTestEnvironment provides complete WebSocket testing environment
type WebSocketTestEnvironment struct {
	*MediaMTXTestEnvironment
	JWTHandler      *security.JWTHandler
	CameraMonitor   *camera.HybridCameraMonitor
	WebSocketServer *websocket.WebSocketServer
}

// SetupWebSocketTestEnvironment creates a complete WebSocket test environment
func SetupWebSocketTestEnvironment(t *testing.T) *WebSocketTestEnvironment {
	// Setup MediaMTX test environment
	mediaEnv := SetupMediaMTXTestEnvironment(t)

	// Create JWT handler
	jwtHandler := SetupTestJWTHandler(t, mediaEnv.ConfigManager)

	// Create camera monitor with real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(
		mediaEnv.ConfigManager,
		mediaEnv.Logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)

	// Create WebSocket server with all dependencies
	webSocketServer := websocket.NewWebSocketServer(
		mediaEnv.ConfigManager,
		mediaEnv.Logger,
		cameraMonitor,
		jwtHandler,
		mediaEnv.Controller,
	)

	return &WebSocketTestEnvironment{
		MediaMTXTestEnvironment: mediaEnv,
		JWTHandler:              jwtHandler,
		CameraMonitor:           cameraMonitor,
		WebSocketServer:         webSocketServer,
	}
}

// TeardownWebSocketTestEnvironment cleans up WebSocket test environment
func TeardownWebSocketTestEnvironment(t *testing.T, env *WebSocketTestEnvironment) {
	if env != nil {
		// Stop camera monitor
		if env.CameraMonitor != nil {
			if err := env.CameraMonitor.Stop(); err != nil {
				t.Logf("Warning: Failed to stop camera monitor: %v", err)
			}
		}

		// Teardown MediaMTX environment
		if env.MediaMTXTestEnvironment != nil {
			TeardownMediaMTXTestEnvironment(t, env.MediaMTXTestEnvironment)
		}
	}
}

// TeardownMediaMTXTestEnvironment cleans up MediaMTX test environment
func TeardownMediaMTXTestEnvironment(t *testing.T, env *MediaMTXTestEnvironment) {
	if env != nil && env.Controller != nil {
		// Stop controller with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := env.Controller.Stop(ctx); err != nil {
			t.Logf("Warning: Failed to stop MediaMTX controller: %v", err)
		}
	}

	// Teardown base environment
	if env != nil && env.TestEnvironment != nil {
		TeardownTestEnvironment(t, env.TestEnvironment)
	}
}

// TeardownTestEnvironment cleans up test environment
func TeardownTestEnvironment(t *testing.T, env *TestEnvironment) {
	if env != nil && env.TempDir != "" {
		err := os.RemoveAll(env.TempDir)
		require.NoError(t, err, "Failed to cleanup temp directory")
	}
}

// SetupRealMediaMTXController creates a real MediaMTX controller for testing
func SetupRealMediaMTXController(t *testing.T, configManager *config.ConfigManager, logger *logging.Logger) mediamtx.MediaMTXController {
	controller, err := mediamtx.NewControllerWithConfigManager(configManager, logger.Logger)
	require.NoError(t, err, "Failed to create real MediaMTX controller")
	return controller
}

// CreateTestMediaMTXConfig creates test-specific MediaMTX configuration
func CreateTestMediaMTXConfig(tempDir string) *config.MediaMTXConfig {
	return &config.MediaMTXConfig{
		Host:               "localhost",
		APIPort:            9997,
		HealthCheckTimeout: 5 * time.Second,
		RecordingsPath:     filepath.Join(tempDir, "recordings"),
		SnapshotsPath:      filepath.Join(tempDir, "snapshots"),
	}
}

// copyTestConfig copies the test configuration file to the specified path
func copyTestConfig(destPath string) error {
	// Read the test configuration from fixtures
	sourcePath := "tests/fixtures/test_config.yaml"

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// Create a minimal test config if fixture doesn't exist
		return createMinimalTestConfig(destPath)
	}

	// Copy the file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, sourceData, 0644)
}

// createMinimalTestConfig creates a minimal test configuration
func createMinimalTestConfig(configPath string) error {
	minimalConfig := `# Minimal Test Configuration
mediamtx:
  host: "localhost"
  api_port: 9997
  health_check_timeout: "5s"
  recording:
    default_path: "/tmp/test_recordings"
    fallback_path: "/tmp/test_fallback"
    default_rotation_size: 104857600
    default_max_duration: 3600
    default_retention_days: 7
  snapshots:
    default_path: "/tmp/test_snapshots"
    fallback_path: "/tmp/test_snapshot_fallback"
    format: "jpeg"
    quality: 85
    max_width: 1920
    max_height: 1080
    default_retention_days: 30

security:
  jwt_secret_key: "test-secret-key-for-unit-testing-only-not-for-production"
  jwt_expiry_hours: 24
  rate_limit_requests: 100
  rate_limit_window: "1m"

storage:
  warn_percent: 80
  block_percent: 90
  default_path: "/tmp/test_storage"
  fallback_path: "/tmp/test_storage_fallback"

camera:
  detection_timeout: 2.0
  poll_interval: 0.1
  device_range: [0, 9]
  capability_detection:
    enabled: true
    max_retries: 3
    timeout: 5.0

websocket:
  host: "localhost"
  port: 8002
  read_timeout: "30s"
  write_timeout: "30s"
  max_message_size: 1048576

logging:
  level: "debug"
  format: "json"
  output: "stdout"
`

	return os.WriteFile(configPath, []byte(minimalConfig), 0644)
}

// createTestDirectories creates necessary test directories
func createTestDirectories(tempDir string) error {
	dirs := []string{
		filepath.Join(tempDir, "recordings"),
		filepath.Join(tempDir, "snapshots"),
		filepath.Join(tempDir, "storage"),
		filepath.Join(tempDir, "fallback"),
		filepath.Join(tempDir, "fallback_recordings"),
		filepath.Join(tempDir, "fallback_snapshots"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetTestConfigPath returns the path to the test configuration file
func GetTestConfigPath() string {
	// Try to find test config in fixtures first
	fixturePath := "tests/fixtures/test_config.yaml"
	if _, err := os.Stat(fixturePath); err == nil {
		return fixturePath
	}

	// Fallback to default config
	return "config/default.yaml"
}

// ValidateTestConfiguration validates that test configuration is properly set up
func ValidateTestConfiguration(t *testing.T, configManager *config.ConfigManager) {
	cfg := configManager.GetConfig()
	require.NotNil(t, cfg, "Configuration should not be nil")

	// Validate critical configuration sections
	require.NotEmpty(t, cfg.Security.JWTSecretKey, "JWT secret key should be configured")
	require.Greater(t, cfg.Security.RateLimitRequests, 0, "Rate limit requests should be configured")
	require.NotEmpty(t, cfg.MediaMTX.Host, "MediaMTX host should be configured")
	require.NotZero(t, cfg.MediaMTX.APIPort, "MediaMTX API port should be configured")
	require.Greater(t, cfg.Storage.WarnPercent, 0, "Storage warn percent should be configured")
	require.Greater(t, cfg.Storage.BlockPercent, 0, "Storage block percent should be configured")
}

// ValidateMediaMTXController validates that MediaMTX controller is properly set up
func ValidateMediaMTXController(t *testing.T, controller mediamtx.MediaMTXController) {
	require.NotNil(t, controller, "MediaMTX controller should not be nil")

	// Test basic controller functionality
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get health status (may fail if MediaMTX not running, but controller should be valid)
	_, err := controller.GetHealth(ctx)
	// We don't require success here as MediaMTX may not be running in test environment
	// But we do require that the controller is properly initialized
	if err != nil {
		t.Logf("MediaMTX health check failed (expected if MediaMTX not running): %v", err)
	}
}

// SetupTestJWTHandler creates a JWT handler for testing
func SetupTestJWTHandler(t *testing.T, configManager *config.ConfigManager) *security.JWTHandler {
	cfg := configManager.GetConfig()
	require.NotNil(t, cfg, "Configuration should be available")

	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(t, err, "Failed to create JWT handler")

	return jwtHandler
}

// GenerateTestToken creates a test JWT token for authentication
func GenerateTestToken(t *testing.T, jwtHandler *security.JWTHandler, userID string, role string) string {
	token, err := jwtHandler.GenerateToken(userID, role, 24) // 24 hours expiry
	require.NoError(t, err, "Failed to generate test token")
	return token
}

// CreateAuthenticatedClient creates an authenticated client connection for testing
func CreateAuthenticatedClient(t *testing.T, jwtHandler *security.JWTHandler, userID string, role string) *websocket.ClientConnection {
	_ = GenerateTestToken(t, jwtHandler, userID, role) // Generate token for authentication

	return &websocket.ClientConnection{
		ClientID:      "test_client_" + userID,
		Authenticated: true,
		UserID:        userID,
		Role:          role,
		ConnectedAt:   time.Now(),
		Subscriptions: make(map[string]bool),
	}
}
