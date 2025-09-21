/*
Universal Test Utilities - Domain-Agnostic Test Infrastructure

Provides common test utilities that can be used across all modules,
inspired by the best patterns from MediaMTX helpers but without
module-specific dependencies.

Requirements Coverage:
- REQ-TEST-001: Universal test setup and teardown
- REQ-TEST-002: Configuration-driven directory management
- REQ-TEST-003: Standardized fixture loading
- REQ-TEST-004: Domain-agnostic assertion utilities

Design Principles:
- No module-specific dependencies
- Configuration-driven (no hardcoded paths)
- Fixture-based (edit fixture → all tests react)
- Progressive migration support
*/

package testutils

import (
	"os"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// UniversalTestSetup provides common test setup pattern
type UniversalTestSetup struct {
	t             *testing.T
	configManager *config.ConfigManager
	logger        *logging.Logger
	tempDirs      []string
}

// SetupTest creates universal test setup (inspired by MediaMTX helper)
func SetupTest(t *testing.T, fixtureName string) *UniversalTestSetup {
	// Create logger using standardized pattern
	logger := logging.GetLogger("test-universal")
	
	// Create directories from fixture configuration
	directoryManager := NewDirectoryManager(t)
	directoryManager.CreateDirectoriesFromFixture(fixtureName)
	
	// Load configuration using fixture
	fixtureLoader := NewFixtureLoader(t)
	configManager := fixtureLoader.LoadConfigFromFixture(fixtureName)
	
	setup := &UniversalTestSetup{
		t:             t,
		configManager: configManager,
		logger:        logger,
		tempDirs:      directoryManager.GetCreatedDirectories(),
	}
	
	// Register cleanup
	t.Cleanup(func() {
		setup.Cleanup()
	})
	
	return setup
}

// Cleanup performs universal cleanup
func (s *UniversalTestSetup) Cleanup() {
	// Clean up temporary directories
	for _, dir := range s.tempDirs {
		os.RemoveAll(dir)
	}
}

// GetConfigManager returns the config manager
func (s *UniversalTestSetup) GetConfigManager() *config.ConfigManager {
	return s.configManager
}

// GetLogger returns the logger
func (s *UniversalTestSetup) GetLogger() *logging.Logger {
	return s.logger
}
