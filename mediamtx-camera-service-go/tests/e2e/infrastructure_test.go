package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
)

// TestE2EInfrastructure verifies that the E2E test infrastructure is properly set up
// This test validates configuration loading, directory creation, and basic setup
// without requiring external services to be running.
func TestE2EInfrastructure(t *testing.T) {
	// Setup E2E test infrastructure
	setup := NewE2ETestSetup(t)
	require.NotNil(t, setup, "E2E test setup should be created successfully")

	// Verify configuration was loaded
	require.NotNil(t, setup.universal.GetConfigManager(), "Configuration should be loaded")
	
	// Verify logging was set up
	require.NotNil(t, setup.universal.GetLogger(), "Logger should be initialized")
	
	// Verify temporary directories were created
	require.NotEmpty(t, setup.tempDirs, "Temporary directories should be created")
	
	// Get config for validation
	config := setup.universal.GetConfigManager().GetConfig()
	require.NotNil(t, config, "Config should be accessible")
	
	// Verify JWT secret is configured
	require.NotEmpty(t, config.Security.JWTSecretKey, "JWT secret should be configured")
	
	// Verify test-specific paths are configured
	assert.Contains(t, config.MediaMTX.RecordingsPath, "e2e-test", "Recordings path should be test-specific")
	assert.Contains(t, config.MediaMTX.SnapshotsPath, "e2e-test", "Snapshots path should be test-specific")
	
	t.Log("✅ E2E infrastructure setup verified successfully")
}

// TestE2EHelperFunctions verifies that E2E helper functions work correctly
func TestE2EHelperFunctions(t *testing.T) {
	setup := NewE2ETestSetup(t)
	
	// Test JWT token generation
	token := GenerateTestToken(t, "admin", 24)
	require.NotEmpty(t, token, "JWT token should be generated")
	
	// Test timeout functions
	timeout := GetStandardTimeout("test")
	require.NotZero(t, timeout, "Timeout should be non-zero")
	
	t.Log("✅ E2E helper functions verified successfully")
}

// TestE2EConfigurationValidation verifies that the E2E configuration is valid
func TestE2EConfigurationValidation(t *testing.T) {
	setup := NewE2ETestSetup(t)
	
	// Get config for validation
	config := setup.universal.GetConfigManager().GetConfig()
	require.NotNil(t, config, "Config should be accessible")
	
	// MediaMTX configuration
	assert.Equal(t, "localhost", config.MediaMTX.Host, "MediaMTX host should be localhost")
	assert.Equal(t, 9997, config.MediaMTX.APIPort, "MediaMTX API port should be 9997")
	assert.Equal(t, 8002, config.Server.Port, "WebSocket server port should be 8002")
	
	// Security configuration
	assert.NotEmpty(t, config.Security.JWTSecretKey, "JWT secret should be configured")
	assert.Equal(t, 24, config.Security.JWTExpiryHours, "JWT expiry should be 24 hours")
	
	// Recording configuration
	assert.True(t, config.Recording.Enabled, "Recording should be enabled")
	assert.Equal(t, "mp4", config.Recording.RecordFormat, "Recording format should be mp4")
	
	// Snapshot configuration
	assert.True(t, config.Snapshots.Enabled, "Snapshots should be enabled")
	assert.Equal(t, "jpeg", config.Snapshots.Format, "Snapshot format should be jpeg")
	
	t.Log("✅ E2E configuration validation passed")
}

// TestE2ETimeoutConstants verifies that timeout constants are properly defined
func TestE2ETimeoutConstants(t *testing.T) {
	// Test universal timeout constants
	assert.NotZero(t, testutils.UniversalTimeoutShort, "Short timeout should be defined")
	assert.NotZero(t, testutils.UniversalTimeoutMedium, "Medium timeout should be defined")
	assert.NotZero(t, testutils.UniversalTimeoutLong, "Long timeout should be defined")
	assert.NotZero(t, testutils.UniversalTimeoutVeryLong, "Very long timeout should be defined")
	assert.NotZero(t, testutils.UniversalTimeoutExtreme, "Extreme timeout should be defined")
	
	// Test that timeouts are in ascending order
	assert.True(t, testutils.UniversalTimeoutShort < testutils.UniversalTimeoutMedium, "Timeouts should be in ascending order")
	assert.True(t, testutils.UniversalTimeoutMedium < testutils.UniversalTimeoutLong, "Timeouts should be in ascending order")
	assert.True(t, testutils.UniversalTimeoutLong < testutils.UniversalTimeoutVeryLong, "Timeouts should be in ascending order")
	assert.True(t, testutils.UniversalTimeoutVeryLong < testutils.UniversalTimeoutExtreme, "Timeouts should be in ascending order")
	
	t.Log("✅ E2E timeout constants verified")
}

// TestE2EFileSizeConstants verifies that file size constants are properly defined
func TestE2EFileSizeConstants(t *testing.T) {
	// Test universal file size constants
	assert.NotZero(t, testutils.UniversalMinRecordingFileSize, "Minimum recording file size should be defined")
	assert.NotZero(t, testutils.UniversalMinSnapshotFileSize, "Minimum snapshot file size should be defined")
	
	// Test that recording files are larger than snapshot files (reasonable assumption)
	assert.True(t, testutils.UniversalMinRecordingFileSize > testutils.UniversalMinSnapshotFileSize, 
		"Recording files should be larger than snapshot files")
	
	t.Log("✅ E2E file size constants verified")
}
