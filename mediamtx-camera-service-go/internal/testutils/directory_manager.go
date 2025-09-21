/*
Directory Manager - Configuration-Driven Directory Creation

Provides configuration-driven directory creation that eliminates
hardcoded paths and ensures all directory management comes from fixtures.

GOAL: Edit fixture → All tests react automatically
*/

package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// DirectoryManager handles configuration-driven directory creation
type DirectoryManager struct {
	t            *testing.T
	createdDirs  []string
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager(t *testing.T) *DirectoryManager {
	return &DirectoryManager{
		t:           t,
		createdDirs: make([]string, 0),
	}
}

// CreateDirectoriesFromFixture creates directories based on fixture configuration
// This eliminates hardcoded directory paths across all modules
func (dm *DirectoryManager) CreateDirectoriesFromFixture(fixtureName string) {
	// Load fixture to get configured paths
	fixtureConfig := dm.loadFixtureConfig(fixtureName)
	
	// Extract directory paths from configuration
	directories := dm.extractDirectoryPaths(fixtureConfig)
	
	// Create directories with proper permissions
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0777)
		require.NoError(dm.t, err, "Failed to create directory: %s", dir)
		dm.createdDirs = append(dm.createdDirs, dir)
	}
}

// GetCreatedDirectories returns list of created directories for cleanup
func (dm *DirectoryManager) GetCreatedDirectories() []string {
	return dm.createdDirs
}

// loadFixtureConfig loads fixture configuration
func (dm *DirectoryManager) loadFixtureConfig(fixtureName string) map[string]interface{} {
	// Use same fixture resolution as FixtureLoader
	fixturePath := dm.resolveFixturePath(fixtureName)
	
	data, err := os.ReadFile(fixturePath)
	require.NoError(dm.t, err, "Failed to read fixture %s", fixtureName)
	
	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	require.NoError(dm.t, err, "Failed to parse fixture %s", fixtureName)
	
	return config
}

// extractDirectoryPaths extracts all directory paths from configuration
func (dm *DirectoryManager) extractDirectoryPaths(config map[string]interface{}) []string {
	var directories []string
	
	// Extract MediaMTX paths
	if mediamtx, ok := config["mediamtx"].(map[string]interface{}); ok {
		if recordingsPath, ok := mediamtx["recordings_path"].(string); ok {
			directories = append(directories, recordingsPath)
		}
		if snapshotsPath, ok := mediamtx["snapshots_path"].(string); ok {
			directories = append(directories, snapshotsPath)
		}
		if configPath, ok := mediamtx["config_path"].(string); ok {
			// For config files, create parent directory
			directories = append(directories, filepath.Dir(configPath))
		}
	}
	
	// Extract storage paths
	if storage, ok := config["storage"].(map[string]interface{}); ok {
		if defaultPath, ok := storage["default_path"].(string); ok {
			directories = append(directories, defaultPath)
		}
		if fallbackPath, ok := storage["fallback_path"].(string); ok {
			directories = append(directories, fallbackPath)
		}
	}
	
	// Extract logging paths
	if logging, ok := config["logging"].(map[string]interface{}); ok {
		if filePath, ok := logging["file_path"].(string); ok {
			// For log files, create parent directory
			directories = append(directories, filepath.Dir(filePath))
		}
	}
	
	return directories
}

// resolveFixturePath resolves fixture path using standard locations
func (dm *DirectoryManager) resolveFixturePath(fixtureName string) string {
	// Same logic as FixtureLoader for consistency
	paths := []string{
		filepath.Join("tests", "fixtures", fixtureName),
		filepath.Join("..", "..", "tests", "fixtures", fixtureName),
		filepath.Join("..", "..", "..", "tests", "fixtures", fixtureName),
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return paths[0]
}
