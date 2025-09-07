/*
Test Environment Management

Provides centralized test environment management for WebSocket tests,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-TEST-001: Test environment setup and management
- REQ-TEST-002: Performance optimization for test execution
- REQ-TEST-003: Resource cleanup and isolation

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package testutils

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnvironmentManager manages the shared test environment
type TestEnvironmentManager struct {
	environment *TestEnvironment
	initialized bool
}

var (
	envManager *TestEnvironmentManager
	envMutex   sync.RWMutex
)

// GetTestEnvironment returns the shared test environment
func GetTestEnvironment(t *testing.T) *TestEnvironment {
	// REQ-TEST-001: Test environment setup and management

	envMutex.Lock()
	defer envMutex.Unlock()

	if envManager == nil {
		envManager = &TestEnvironmentManager{}
	}

	if !envManager.initialized {
		envManager.environment = SetupTestEnvironment(t)
		envManager.initialized = true
	}

	return envManager.environment
}

// CleanupTestEnvironment cleans up the shared test environment
func CleanupTestEnvironment(t *testing.T) {
	// REQ-TEST-003: Resource cleanup and isolation

	envMutex.Lock()
	defer envMutex.Unlock()

	if envManager != nil && envManager.initialized {
		TeardownTestEnvironment(t, envManager.environment)
		envManager.initialized = false
		envManager.environment = nil
	}
}

// CreateTestDirectories creates necessary test directories
func CreateTestDirectories(t *testing.T, baseDir string) map[string]string {
	// REQ-TEST-001: Test environment setup and management

	dirs := map[string]string{
		"recordings": filepath.Join(baseDir, "recordings"),
		"snapshots":  filepath.Join(baseDir, "snapshots"),
		"logs":       filepath.Join(baseDir, "logs"),
		"config":     filepath.Join(baseDir, "config"),
	}

	for name, path := range dirs {
		err := os.MkdirAll(path, 0755)
		require.NoError(t, err, "Failed to create test directory: %s", name)
	}

	return dirs
}

// ValidateTestEnvironment validates that the test environment is properly set up
func ValidateTestEnvironment(t *testing.T, env *TestEnvironment) {
	// REQ-TEST-001: Test environment setup and management

	require.NotNil(t, env, "Test environment should not be nil")
	require.NotNil(t, env.Server, "Test server should not be nil")
	require.NotNil(t, env.Config, "Test config should not be nil")
	require.NotEmpty(t, env.TempDir, "Temp directory should not be empty")
	require.NotNil(t, env.Logger, "Logger should not be nil")
	require.NotEmpty(t, env.ConfigPath, "Config path should not be empty")

	// Validate server is running
	require.True(t, env.Server.IsRunning(), "Test server should be running")

	// Validate temp directory exists
	_, err := os.Stat(env.TempDir)
	require.NoError(t, err, "Temp directory should exist")
}

// GetTestServerPort returns the port of the shared test server
func GetTestServerPort() int {
	envMutex.RLock()
	defer envMutex.RUnlock()

	if envManager != nil && envManager.initialized && envManager.environment != nil {
		return envManager.environment.Server.Config.Port
	}
	return 0
}

// IsTestEnvironmentReady checks if the test environment is ready
func IsTestEnvironmentReady() bool {
	envMutex.RLock()
	defer envMutex.RUnlock()

	return envManager != nil && envManager.initialized && envManager.environment != nil && envManager.environment.Server.IsRunning()
}

// ResetTestEnvironment resets the test environment (for cleanup between test suites)
func ResetTestEnvironment(t *testing.T) {
	// REQ-TEST-003: Resource cleanup and isolation

	envMutex.Lock()
	defer envMutex.Unlock()

	if envManager != nil && envManager.initialized {
		// Stop the shared server
		if envManager.environment.Server != nil {
			err := envManager.environment.Server.Stop()
			if err != nil {
				t.Logf("Warning: Failed to stop test server during reset: %v", err)
			}
		}

		// Clean up environment
		TeardownTestEnvironment(t, envManager.environment)

		// Reset state
		envManager.initialized = false
		envManager.environment = nil
	}
}
