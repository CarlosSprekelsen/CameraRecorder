/*
Controller Config Tests - Refactored with Asserters

This file demonstrates the dramatic reduction possible using ConfigAsserters.
Original tests had massive duplication of setup, Progressive Readiness, and validation.
Refactored tests focus on business logic only.

Requirements Coverage:
- REQ-MTX-001: Configuration management
*/

package mediamtx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigIntegration_GetRecordingConfig_ReqMTX001_Success_Refactored demonstrates config integration
func TestConfigIntegration_GetRecordingConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	// Create config asserter with full setup (eliminates 6 lines of setup)
	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get recording config (eliminates 10+ lines of setup and validation)
	config := asserter.AssertGetRecordingConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "Recording config should not be nil")
	assert.NotEmpty(t, config.Format, "Recording format should be configured")

	t.Logf("✅ Recording config validated: %s", config.Format)
}

// TestConfigIntegration_GetSnapshotConfig_ReqMTX001_Success_Refactored demonstrates snapshot config
func TestConfigIntegration_GetSnapshotConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get snapshot config (eliminates 10+ lines of setup and validation)
	config := asserter.AssertGetSnapshotConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "Snapshot config should not be nil")
	assert.NotEmpty(t, config.Format, "Snapshot format should be configured")

	t.Logf("✅ Snapshot config validated: %s", config.Format)
}

// TestConfigIntegration_GetFFmpegConfig_ReqMTX001_Success_Refactored demonstrates FFmpeg config
func TestConfigIntegration_GetFFmpegConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get FFmpeg config (eliminates 10+ lines of setup and validation)
	config := asserter.AssertGetFFmpegConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "FFmpeg config should not be nil")

	t.Logf("✅ FFmpeg config validated successfully")
}

// TestConfigIntegration_GetCameraConfig_ReqMTX001_Success_Refactored demonstrates camera config
func TestConfigIntegration_GetCameraConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get camera config (eliminates 10+ lines of setup and validation)
	config := asserter.AssertGetCameraConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "Camera config should not be nil")
	assert.Greater(t, config.PollInterval, 0.0, "Poll interval should be positive")

	t.Logf("✅ Camera config validated: %.2f seconds", config.PollInterval)
}

// TestConfigIntegration_GetPerformanceConfig_ReqMTX001_Success_Refactored demonstrates performance config
func TestConfigIntegration_GetPerformanceConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get performance config (eliminates 10+ lines of setup and validation)
	config := asserter.AssertGetPerformanceConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "Performance config should not be nil")

	t.Logf("✅ Performance config validated successfully")
}

// TestController_GetConfig_ReqMTX001_Success_Refactored demonstrates main config retrieval
func TestController_GetConfig_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: Configuration management

	asserter := NewConfigAsserter(t)
	defer asserter.Cleanup()

	// Get main config (eliminates 15+ lines of setup and validation)
	config := asserter.AssertGetMainConfig()

	// Test-specific business logic only
	assert.NotNil(t, config, "Main config should not be nil")

	t.Logf("✅ Main config validated successfully")
}
