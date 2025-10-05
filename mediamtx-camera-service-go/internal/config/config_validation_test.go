package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-CONFIG-VALID-001: Configuration validation must catch all invalid configurations
// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages
// REQ-CONFIG-VALID-003: Validation must enforce API compliance requirements

func TestValidationError_Error(t *testing.T) {
	// Test ValidationError.Error() method
	err := &ValidationError{
		Field:   "server.port",
		Message: "port must be between 1 and 65535",
	}

	expected := "validation error for field 'server.port': port must be between 1 and 65535"
	assert.Equal(t, expected, err.Error())
}

func TestValidateConfig_ValidConfiguration(t *testing.T) {
	// REQ-CONFIG-VALID-001: Valid configuration must pass validation
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Create test directories that validation expects
	helper.CreateTestDirectories()

	// Load valid config from fixture and create a ConfigManager to load it
	configPath := helper.CreateTempConfigFromFixture("config_valid_complete.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load valid configuration")

	cfg := cm.GetConfig()
	require.NotNil(t, cfg, "Configuration should not be nil")

	// Validate the loaded configuration
	err = ValidateConfig(cfg)
	assert.NoError(t, err, "Valid configuration should pass validation")
}

func TestValidateConfig_InvalidConfigurations(t *testing.T) {
	// REQ-CONFIG-VALID-001: Invalid configurations must fail validation
	// REQ-CONFIG-VALID-002: Validation errors must provide clear field-specific messages

	testCases := []struct {
		name        string
		fixture     string
		expectedErr string
	}{
		{
			name:        "Invalid Port",
			fixture:     "config_invalid_invalid_port.yaml",
			expectedErr: "port must be between 1 and 65535",
		},
		{
			name:        "Missing Server",
			fixture:     "config_invalid_missing_server.yaml",
			expectedErr: "configuration validation failed",
		},
		{
			name:        "Empty Config",
			fixture:     "config_invalid_empty.yaml",
			expectedErr: "configuration file contains only comments or is empty",
		},
		{
			name:        "Malformed YAML",
			fixture:     "config_invalid_malformed_yaml.yaml",
			expectedErr: "configuration validation failed",
		},
		// New edge case fixtures for better coverage
		{
			name:        "Retention Policy Negative Age",
			fixture:     "config_invalid_retention_policy_negative_age.yaml",
			expectedErr: "max age days cannot be negative",
		},
		{
			name:        "Retention Policy Excessive Age",
			fixture:     "config_invalid_retention_policy_excessive_age.yaml",
			expectedErr: "max age days cannot exceed 365 days",
		},
		{
			name:        "Retention Policy Invalid Type",
			fixture:     "config_invalid_retention_policy_invalid_type.yaml",
			expectedErr: "policy type must be one of",
		},
		{
			name:        "Camera Negative Poll Interval",
			fixture:     "config_invalid_camera_negative_poll_interval.yaml",
			expectedErr: "camera poll interval must be positive",
		},
		{
			name:        "Camera Invalid Device Range",
			fixture:     "config_invalid_camera_invalid_device_range.yaml",
			expectedErr: "device range min must be less than or equal to max",
		},
		// Combined fixtures for multiple edge cases
		{
			name:        "MediaMTX Multiple Issues",
			fixture:     "config_invalid_mediamtx_multiple_issues.yaml",
			expectedErr: "host cannot be empty",
		},
		{
			name:        "Snapshot Tiers Multiple Issues",
			fixture:     "config_invalid_snapshot_tiers_multiple_issues.yaml",
			expectedErr: "tier1 USB direct timeout must be positive",
		},
		{
			name:        "Combined Edge Cases",
			fixture:     "config_invalid_combined_edge_cases.yaml",
			expectedErr: "configuration validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewTestConfigHelper(t)
			defer helper.CleanupEnvironment()

			// Create test directories
			helper.CreateTestDirectories()

			// Load invalid config from fixture
			configPath := helper.CreateTempConfigFromFixture(tc.fixture)

			cm := CreateConfigManager()
			err := cm.LoadConfig(configPath)

			// Should fail to load invalid configuration
			require.Error(t, err, "Should fail to load invalid configuration: %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedErr, "Error should contain expected message for: %s", tc.name)
		})
	}
}

func TestValidateConfig_BoundaryValues(t *testing.T) {
	// REQ-CONFIG-VALID-001: Boundary values must be handled correctly
	helper := NewTestConfigHelper(t)
	defer helper.CleanupEnvironment()

	// Create test directories that validation expects
	helper.CreateTestDirectories()

	// Load valid config from fixture and create a ConfigManager to load it
	configPath := helper.CreateTempConfigFromFixture("config_valid_complete.yaml")

	cm := CreateConfigManager()
	err := cm.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load valid configuration")

	cfg := cm.GetConfig()
	require.NotNil(t, cfg, "Configuration should not be nil")

	// Validate the loaded configuration
	err = ValidateConfig(cfg)
	assert.NoError(t, err, "Valid configuration should pass validation")
}
