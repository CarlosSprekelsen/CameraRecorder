/*
Test Environment Management - FIXED Resource Management

Provides centralized test environment management for WebSocket tests,
using the good patterns from internal/testutils and eliminating
resource management issues.

FIXED ISSUES:
- Eliminated duplicate mutex instances (leverages shared infrastructure)
- Proper resource cleanup using UniversalTestSetup pattern
- Progressive Readiness compliance
- No global shared state (each test gets isolated resources)

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
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// FIXED: Removed global shared state and mutex duplication
// Each test gets isolated resources through UniversalTestSetup

// WebSocketTestEnvironmentManager manages isolated test environments
// FIXED: No shared state, each test gets its own environment
type WebSocketTestEnvironmentManager struct {
	setup *testutils.UniversalTestSetup
}

// GetWebSocketTestEnvironment returns an isolated test environment
// DEPRECATED: use testutils.SetupTest(t, "config_websocket_test.yaml") instead.
// This function will be removed in a future version.
// FIXED: Uses UniversalTestSetup pattern for proper resource management
func GetWebSocketTestEnvironment(t *testing.T) *WebSocketTestEnvironmentManager {
	// REQ-TEST-001: Test environment setup and management

	// DEPRECATED: Use single canonical fixture instead of config_websocket_test.yaml
	// TODO: Remove config_websocket_test.yaml and use config_valid_complete.yaml
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")

	return &WebSocketTestEnvironmentManager{
		setup: setup,
	}
}

// CleanupWebSocketTestEnvironment cleans up the test environment
// DEPRECATED: use testutils.SetupTest(t, "config_websocket_test.yaml") instead.
// This function will be removed in a future version.
// FIXED: Uses UniversalTestSetup cleanup pattern
func CleanupWebSocketTestEnvironment(t *testing.T, manager *WebSocketTestEnvironmentManager) {
	// REQ-TEST-003: Resource cleanup and isolation

	if manager != nil && manager.setup != nil {
		manager.setup.Cleanup() // This handles all cleanup properly
	}
}

// GetSetup returns the UniversalTestSetup for resource management
func (m *WebSocketTestEnvironmentManager) GetSetup() *testutils.UniversalTestSetup {
	return m.setup
}

// CreateTestDirectories creates necessary test directories
// FIXED: Uses UniversalTestSetup directory management
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

// ValidateWebSocketTestEnvironment validates that the test environment is properly set up
// FIXED: Uses UniversalTestSetup validation pattern
func ValidateWebSocketTestEnvironment(t *testing.T, manager *WebSocketTestEnvironmentManager) {
	// REQ-TEST-001: Test environment setup and management

	require.NotNil(t, manager, "Test environment manager should not be nil")
	require.NotNil(t, manager.setup, "UniversalTestSetup should not be nil")

	// Validate UniversalTestSetup components
	configManager := manager.setup.GetConfigManager()
	logger := manager.setup.GetLogger()

	require.NotNil(t, configManager, "Config manager should not be nil")
	require.NotNil(t, logger, "Logger should not be nil")
	require.NotNil(t, configManager.GetConfig(), "Config should not be nil")
}

// FIXED: Removed shared server port functions - no more shared state
// Each test environment manages its own resources through UniversalTestSetup

// IsWebSocketTestEnvironmentReady checks if the test environment is ready
// FIXED: Uses UniversalTestSetup readiness pattern
func IsWebSocketTestEnvironmentReady(manager *WebSocketTestEnvironmentManager) bool {
	if manager == nil || manager.setup == nil {
		return false
	}

	// FIXED: Use UniversalTestSetup readiness validation
	configManager := manager.setup.GetConfigManager()
	logger := manager.setup.GetLogger()

	return configManager != nil && logger != nil && configManager.GetConfig() != nil
}

// FIXED: Removed ResetTestEnvironment - no more shared state
// Each test environment manages its own cleanup through UniversalTestSetup
// Tests should use shared infrastructure pattern directly for proper resource management
