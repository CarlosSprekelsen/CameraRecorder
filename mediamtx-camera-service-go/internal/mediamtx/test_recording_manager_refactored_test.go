/*
MediaMTX Recording Manager Tests - Refactored with Progressive Readiness

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json

Refactored from recording_manager_test.go (613 lines → ~200 lines)
Eliminates massive duplication using RecordingManagerAsserter
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestNewRecordingManager_ReqMTX001_Refactored tests recording manager creation with real hardware
func TestNewRecordingManager_ReqMTX001_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	// Validate recording manager creation
	asserter.AssertRecordingManagerCreation()
}

// TestRecordingManager_CompleteLifecycle_ReqMTX002_Refactored tests complete recording lifecycle with data validation
func TestRecordingManager_CompleteLifecycle_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	// Complete lifecycle validation: Start → Verify Recording → Stop → Verify File
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.GetHelper().MustGetCameraID(t, asserter.GetContext(), asserter.GetController())
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}

	asserter.AssertCompleteRecordingLifecycle(cameraID, options)
}

// TestRecordingManager_StopRecording_ReqMTX002_Refactored tests recording stop operation
func TestRecordingManager_StopRecording_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.GetHelper().MustGetCameraID(t, asserter.GetContext(), asserter.GetController())
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}

	// Start recording first
	asserter.AssertStartRecording(cameraID, options)

	// Brief recording
	time.Sleep(testutils.UniversalRetryDelay)

	// Stop recording
	asserter.AssertStopRecording(cameraID)
}

// TestRecordingManager_GetRecordingsListAPI_ReqMTX002_Refactored tests recording list API
func TestRecordingManager_GetRecordingsListAPI_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - API compliance validation
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	// Test different pagination scenarios
	testCases := []struct {
		name   string
		limit  int
		offset int
	}{
		{"first_page", 10, 0},
		{"second_page", 5, 5},
		{"large_page", 100, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asserter.AssertGetRecordingsList(tc.limit, tc.offset)
		})
	}
}

// TestRecordingManager_StartRecordingCreatesPath_ReqMTX003_Refactored tests path creation
func TestRecordingManager_StartRecordingCreatesPath_ReqMTX003_Refactored(t *testing.T) {
	// REQ-MTX-003: Path creation and persistence - Validate MediaMTX API integration
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.GetHelper().MustGetCameraID(t, asserter.GetContext(), asserter.GetController())
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}

	// Start recording and verify path creation
	session := asserter.AssertStartRecording(cameraID, options)
	assert.NotNil(t, session, "Recording session should be created")

	// Brief recording
	time.Sleep(testutils.UniversalRetryDelay)

	// Verify file was created (path creation validation)
	asserter.AssertRecordingFileExists(cameraID)

	// Stop recording
	asserter.AssertStopRecording(cameraID)
}

// TestRecordingManager_APISchemaCompliance_ReqMTX001_Refactored tests API schema compliance
func TestRecordingManager_APISchemaCompliance_ReqMTX001_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration - Schema validation per swagger.json
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	// Test API schema compliance through basic operations
	recordings := asserter.AssertGetRecordingsList(10, 0)
	assert.NotNil(t, recordings, "Recording list response should match schema")

	// Validate response structure
	assert.NotNil(t, recordings.Files, "Files field should exist")
	assert.GreaterOrEqual(t, recordings.Total, int64(0), "Total should be non-negative")
}

// TestRecordingManager_APIErrorHandling_ReqMTX004_Refactored tests API error handling
func TestRecordingManager_APIErrorHandling_ReqMTX004_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring and circuit breaker - Error handling validation
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertAPIErrorHandling()
}

// TestRecordingManager_ErrorHandling_ReqMTX007_Refactored tests error handling
func TestRecordingManager_ErrorHandling_ReqMTX007_Refactored(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	// Test error scenarios
	asserter.AssertAPIErrorHandling()
}

// TestRecordingManager_ConcurrentAccess_ReqMTX001_Refactored tests concurrent access
func TestRecordingManager_ConcurrentAccess_ReqMTX001_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertConcurrentAccess()
}

// TestRecordingManager_StartRecordingWithSegments_ReqMTX002_Refactored tests segmented recording
func TestRecordingManager_StartRecordingWithSegments_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.GetHelper().MustGetCameraID(t, asserter.GetContext(), asserter.GetController())

	// Test segmented recording with different formats
	formats := []string{"fmp4", "mp4"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			options := &PathConf{
				Record:       true,
				RecordFormat: format,
			}

			session := asserter.AssertStartRecording(cameraID, options)
			assert.NotNil(t, session, "Segmented recording should succeed for format: %s", format)

			// Brief recording
			time.Sleep(testutils.UniversalRetryDelay)

			asserter.AssertStopRecording(cameraID)
		})
	}
}

// TestRecordingManager_MultiTierRecording_ReqMTX002_Refactored tests multi-tier recording
func TestRecordingManager_MultiTierRecording_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertMultiTierRecording()
}

// TestRecordingManager_ProgressiveReadinessCompliance_ReqMTX001_Refactored tests Progressive Readiness compliance
func TestRecordingManager_ProgressiveReadinessCompliance_ReqMTX001_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration - Progressive Readiness validation
	asserter := NewRecordingManagerAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertProgressiveReadinessCompliance()
}
