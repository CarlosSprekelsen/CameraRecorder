package testutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// Universal Test Constants - Eliminates timeout duplication across all modules
const (
	DefaultTestTimeout = 30 * time.Second // Standard test timeout
	ShortTestTimeout   = 5 * time.Second  // Quick operations
	LongTestTimeout    = 60 * time.Second // Extended operations
)

// UniversalTestSetup provides common test setup pattern
type UniversalTestSetup struct {
	t             *testing.T
	configManager *config.ConfigManager
	logger        *logging.Logger
	tempDirs      []string
	fixtureName   string // Store fixture name for cleanup
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
		fixtureName:   fixtureName, // Store for cleanup
	}

	// Register cleanup
	t.Cleanup(func() {
		setup.Cleanup()
	})

	return setup
}

// Cleanup performs universal cleanup - fixture-driven and content-only
func (s *UniversalTestSetup) Cleanup() {
	// Clean only CONTENT of configured directories (from fixture), keep directories
	s.cleanFixtureDirectoryContents()

	// Clean up test-specific temporary directories (these can be fully removed)
	for _, dir := range s.tempDirs {
		os.RemoveAll(dir)
	}
}

// cleanFixtureDirectoryContents cleans content of directories specified in fixture
func (s *UniversalTestSetup) cleanFixtureDirectoryContents() {
	if s.configManager == nil {
		return
	}

	config := s.configManager.GetConfig()
	if config == nil {
		return
	}

	// Get paths from fixture configuration (not hardcoded!)
	pathsToClean := []string{}

	// Add MediaMTX paths if they exist in config
	if config.MediaMTX.RecordingsPath != "" {
		pathsToClean = append(pathsToClean, config.MediaMTX.RecordingsPath)
	}
	if config.MediaMTX.SnapshotsPath != "" {
		pathsToClean = append(pathsToClean, config.MediaMTX.SnapshotsPath)
	}

	// Add storage paths if they exist in config
	if config.Storage.DefaultPath != "" {
		pathsToClean = append(pathsToClean, config.Storage.DefaultPath)
	}
	if config.Storage.FallbackPath != "" && config.Storage.FallbackPath != config.Storage.DefaultPath {
		pathsToClean = append(pathsToClean, config.Storage.FallbackPath)
	}

	// Clean CONTENT of each directory, keep directory itself
	for _, dirPath := range pathsToClean {
		s.cleanDirectoryContentsOnly(dirPath)
	}
}

// cleanDirectoryContentsOnly removes files inside directory but keeps directory
func (s *UniversalTestSetup) cleanDirectoryContentsOnly(dirPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Directory doesn't exist or can't read - that's fine
		return
	}

	// Remove each file/subdirectory, but keep the parent directory
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		os.RemoveAll(fullPath) // Remove file or subdirectory
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

// GetStandardContext returns standard test context with timeout
// This eliminates timeout duplication across all modules
func (s *UniversalTestSetup) GetStandardContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultTestTimeout)
}

// GetStandardContextWithTimeout returns test context with custom timeout
// This provides flexibility while maintaining consistency
func (s *UniversalTestSetup) GetStandardContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// GetConfiguredPath returns configured path from fixture with fallback
// This eliminates path resolution duplication across all modules
func (s *UniversalTestSetup) GetConfiguredPath(pathKey, envVar, fallback string) string {
	// Try configuration first
	config := s.configManager.GetConfig()
	if config != nil {
		// Use reflection or type assertion to get path from config
		// This approach works across all module configurations
		if pathValue := s.extractPathFromConfig(config, pathKey); pathValue != "" {
			return pathValue
		}
	}

	// Try environment variable
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	// Use fallback
	return fallback
}

// extractPathFromConfig extracts path value from configuration using interface approach
// This works across different module configuration structures
func (s *UniversalTestSetup) extractPathFromConfig(config interface{}, pathKey string) string {
	// This is a simplified approach - can be enhanced with reflection if needed
	// For now, return empty string to use fallback approach
	return ""
}
