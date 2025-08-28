/*
Test Setup Utilities

Requirements Coverage:
- REQ-TEST-001: Test environment setup
- REQ-TEST-002: Test data preparation
- REQ-TEST-003: Test configuration management
- REQ-TEST-004: Test authentication setup
- REQ-TEST-005: Test evidence collection

Test Categories: Unit/Integration/Test Infrastructure
API Documentation Reference: docs/api/json_rpc_methods.md
*/

// +build unit

//go:build unit

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/require"
)

// TestEnvironment provides test environment setup and management
type TestEnvironment struct {
	ConfigManager *config.ConfigManager
	Logger        *logging.Logger
	TempDir       string
	ConfigPath    string
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

// TeardownTestEnvironment cleans up test environment
func TeardownTestEnvironment(t *testing.T, env *TestEnvironment) {
	if env != nil && env.TempDir != "" {
		err := os.RemoveAll(env.TempDir)
		require.NoError(t, err, "Failed to cleanup temp directory")
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
  detection_timeout: "2s"
  poll_interval: "100ms"
  device_range: [0, 9]
  capability_detection:
    enabled: true
    max_retries: 3
    timeout: "5s"

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
