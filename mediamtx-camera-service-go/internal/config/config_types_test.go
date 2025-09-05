package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-CONFIG-001: Configuration types must support all required fields for API compliance
// REQ-CONFIG-002: Configuration types must have proper mapstructure tags for YAML parsing
// REQ-CONFIG-003: Configuration types must support default values and validation

func TestConfigTypes_LoadFromFixtures(t *testing.T) {
	// Test that all config types can be loaded from fixtures
	// This is much more valuable than testing struct field assignments

	testCases := []struct {
		name    string
		fixture string
	}{
		{
			name:    "Minimal Config",
			fixture: "config_test_minimal.yaml",
		},
		// Note: Other fixtures use production paths that don't exist in tests
		// If we need to test other configs, we should create test-friendly versions
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewTestConfigHelper(t)
			defer helper.CleanupEnvironment()

			// Create test directories
			helper.CreateTestDirectories()

			// Load config from fixture
			configPath := helper.CreateTempConfigFromFixture(tc.fixture)

			cm := CreateConfigManager()
			err := cm.LoadConfig(configPath)
			require.NoError(t, err, "Should load %s successfully", tc.name)

			cfg := cm.GetConfig()
			require.NotNil(t, cfg, "Config should not be nil")

			// Verify all major config sections are loaded
			assert.NotNil(t, cfg.Server, "Server config should be loaded")
			assert.NotNil(t, cfg.MediaMTX, "MediaMTX config should be loaded")
			assert.NotNil(t, cfg.Camera, "Camera config should be loaded")
			assert.NotNil(t, cfg.Logging, "Logging config should be loaded")
			assert.NotNil(t, cfg.Recording, "Recording config should be loaded")
			assert.NotNil(t, cfg.Snapshots, "Snapshots config should be loaded")
			assert.NotNil(t, cfg.FFmpeg, "FFmpeg config should be loaded")
			assert.NotNil(t, cfg.Notifications, "Notifications config should be loaded")
			assert.NotNil(t, cfg.Performance, "Performance config should be loaded")
			assert.NotNil(t, cfg.Security, "Security config should be loaded")
			assert.NotNil(t, cfg.Storage, "Storage config should be loaded")
			assert.NotNil(t, cfg.RetentionPolicy, "RetentionPolicy config should be loaded")
		})
	}
}
