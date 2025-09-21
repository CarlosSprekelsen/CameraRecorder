/*
Fixture Loader - Standardized Configuration Loading

Provides standardized fixture loading that eliminates hardcoded paths
and ensures all tests react to fixture changes automatically.
*/

package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/require"
)

// FixtureLoader handles standardized fixture loading
type FixtureLoader struct {
	t *testing.T
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader(t *testing.T) *FixtureLoader {
	return &FixtureLoader{t: t}
}

// LoadConfigFromFixture loads configuration from fixture file
// This eliminates hardcoded config creation across all modules
func (fl *FixtureLoader) LoadConfigFromFixture(fixtureName string) *config.ConfigManager {
	configManager := config.CreateConfigManager()

	// Use standardized fixture path resolution
	fixturePath := fl.resolveFixturePath(fixtureName)
	
	err := configManager.LoadConfig(fixturePath)
	require.NoError(fl.t, err, "Failed to load config from fixture %s", fixtureName)
	
	return configManager
}

// resolveFixturePath resolves fixture path using standard locations
func (fl *FixtureLoader) resolveFixturePath(fixtureName string) string {
	// Try standard fixture locations
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
	
	// Fallback to first path and let it fail with clear error
	return paths[0]
}
