package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestConfigHelper provides utilities for testing configuration management
type TestConfigHelper struct {
	t            *testing.T
	fixturesPath string
	tempDir      string
}

// NewTestConfigHelper creates a new test configuration helper
func NewTestConfigHelper(t *testing.T) *TestConfigHelper {
	// Get the fixtures path relative to the project root
	projectRoot := getProjectRoot()
	fixturesPath := filepath.Join(projectRoot, "tests", "fixtures")

	// Create temporary directory for test files
	tempDir := t.TempDir()

	return &TestConfigHelper{
		t:            t,
		fixturesPath: fixturesPath,
		tempDir:      tempDir,
	}
}

// LoadFixtureConfig loads a configuration file from the fixtures directory
func (h *TestConfigHelper) LoadFixtureConfig(fixtureName string) string {
	fixturePath := filepath.Join(h.fixturesPath, fixtureName)
	content, err := os.ReadFile(fixturePath)
	require.NoError(h.t, err, "Failed to read fixture: %s", fixtureName)
	return string(content)
}

// CreateTempConfigFile creates a temporary configuration file with the given content
func (h *TestConfigHelper) CreateTempConfigFile(content string) string {
	tempFile := filepath.Join(h.tempDir, "test_config.yaml")
	err := os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(h.t, err, "Failed to create temporary config file")
	return tempFile
}

// CreateTestDirectories creates the directories needed for test configuration validation
func (h *TestConfigHelper) CreateTestDirectories() {
	// Create directories that the config validation expects to exist
	dirs := []string{
		"/tmp/mediamtx.yml",
		"/tmp/recordings",
		"/tmp/snapshots",
		"/tmp/camera-service.log",
	}

	for _, dir := range dirs {
		// For files, create the parent directory
		if strings.HasSuffix(dir, ".yml") || strings.HasSuffix(dir, ".log") {
			parentDir := filepath.Dir(dir)
			err := os.MkdirAll(parentDir, 0755)
			require.NoError(h.t, err, "Failed to create directory: %s", parentDir)

			// Create empty file
			_, err = os.Create(dir)
			require.NoError(h.t, err, "Failed to create file: %s", dir)
		} else {
			// For directories, create them with 0777 permissions for test access
			err := os.MkdirAll(dir, 0777)
			require.NoError(h.t, err, "Failed to create directory: %s", dir)
		}
	}
}

// CreateTempConfigFromFixture creates a temporary config file from a fixture
func (h *TestConfigHelper) CreateTempConfigFromFixture(fixtureName string) string {
	content := h.LoadFixtureConfig(fixtureName)
	return h.CreateTempConfigFile(content)
}

// CleanupEnvironment cleans up environment variables that might interfere with tests
func (h *TestConfigHelper) CleanupEnvironment() {
	envVars := []string{
		"CAMERA_SERVICE_SERVER_HOST",
		"CAMERA_SERVICE_SERVER_PORT",
		"CAMERA_SERVICE_SERVER_WEBSOCKET_PATH",
		"CAMERA_SERVICE_SERVER_MAX_CONNECTIONS",
		"CAMERA_SERVICE_MEDIAMTX_HOST",
		"CAMERA_SERVICE_MEDIAMTX_API_PORT",
		"CAMERA_SERVICE_MEDIAMTX_RTSP_PORT",
		"CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT",
		"CAMERA_SERVICE_MEDIAMTX_HLS_PORT",
		"CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH",
		"CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH",
		"CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_CHECK_INTERVAL",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_FAILURE_THRESHOLD",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_CIRCUIT_BREAKER_TIMEOUT",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_MAX_BACKOFF_INTERVAL",
		"CAMERA_SERVICE_MEDIAMTX_HEALTH_RECOVERY_CONFIRMATION_THRESHOLD",
		"CAMERA_SERVICE_MEDIAMTX_BACKOFF_BASE_MULTIPLIER",
		"CAMERA_SERVICE_MEDIAMTX_PROCESS_TERMINATION_TIMEOUT",
		"CAMERA_SERVICE_MEDIAMTX_PROCESS_KILL_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_POLL_INTERVAL",
		"CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION",
		"CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_TIMEOUT",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_RETRY_INTERVAL",
		"CAMERA_SERVICE_CAMERA_CAPABILITY_MAX_RETRIES",
		"CAMERA_SERVICE_LOGGING_LEVEL",
		"CAMERA_SERVICE_LOGGING_FORMAT",
		"CAMERA_SERVICE_LOGGING_FILE_ENABLED",
		"CAMERA_SERVICE_LOGGING_FILE_PATH",
		"CAMERA_SERVICE_LOGGING_CONSOLE_ENABLED",
		"CAMERA_SERVICE_RECORDING_ENABLED",
		"CAMERA_SERVICE_RECORDING_FORMAT",
		"CAMERA_SERVICE_RECORDING_QUALITY",
		"CAMERA_SERVICE_SNAPSHOTS_ENABLED",
		"CAMERA_SERVICE_SNAPSHOTS_FORMAT",
		"CAMERA_SERVICE_SNAPSHOTS_QUALITY",
		"CAMERA_SERVICE_ENABLE_HOT_RELOAD",
		"CAMERA_SERVICE_ENV",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

// SetEnvironmentVariable sets a specific environment variable for testing
func (h *TestConfigHelper) SetEnvironmentVariable(key, value string) {
	os.Setenv(key, value)
}

// getProjectRoot finds the project root directory by looking for go.mod
func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("Could not find go.mod file")
		}
		dir = parent
	}
}
