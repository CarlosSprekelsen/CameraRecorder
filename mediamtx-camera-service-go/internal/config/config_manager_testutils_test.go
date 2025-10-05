package config_test

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigManager_LoadConfig_WithTestUtils tests config loading using testutils.SetupTest
// This avoids the import cycle issue by using a separate test package
func TestConfigManager_LoadConfig_WithTestUtils(t *testing.T) {
	testCases := []struct {
		name        string
		fixture     string
		expectError bool
		description string
	}{
		{
			name:        "Valid YAML",
			fixture:     "config_valid_complete.yaml",
			expectError: false,
			description: "Should load valid configuration successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use testutils.SetupTest which handles directory creation and environment loading
			setup := testutils.SetupTest(t, tc.fixture)

			// Get the config manager from the setup
			configManager := setup.GetConfigManager()
			require.NotNil(t, configManager, "Config manager should be created")

			cfg := configManager.GetConfig()
			if tc.expectError {
				require.Nil(t, cfg, tc.description)
			} else {
				require.NotNil(t, cfg, tc.description)
				// Access config fields through the public interface
				assert.Equal(t, "0.0.0.0", cfg.Server.Host)
				assert.Equal(t, 8002, cfg.Server.Port)
				assert.Equal(t, "/ws", cfg.Server.WebSocketPath)
				assert.NotEmpty(t, cfg.Security.JWTSecretKey, "JWT secret key should be populated")
				assert.NotEmpty(t, cfg.MediaMTX.Codec.VideoProfile, "Video profile should be populated")
			}
		})
	}
}

// TestConfigManager_StructMapping_WithTestUtils tests that struct mapping works correctly
func TestConfigManager_StructMapping_WithTestUtils(t *testing.T) {
	// Use the minimal debug fixture
	setup := testutils.SetupTest(t, "config_debug_minimal.yaml")

	configManager := setup.GetConfigManager()
	require.NotNil(t, configManager, "Config manager should be created")

	cfg := configManager.GetConfig()
	require.NotNil(t, cfg, "Config should be loaded")

	// Test that struct fields are properly mapped
	assert.Equal(t, "/ws", cfg.Server.WebSocketPath, "WebSocketPath should be mapped correctly")
	assert.NotEmpty(t, cfg.Security.JWTSecretKey, "JWTSecretKey should be mapped correctly")
	assert.Equal(t, "main", cfg.MediaMTX.Codec.VideoProfile, "VideoProfile should be mapped correctly")
	assert.Equal(t, "0.0.0.0", cfg.Server.Host, "Host should be mapped correctly")
	assert.Equal(t, 8002, cfg.Server.Port, "Port should be mapped correctly")
}
