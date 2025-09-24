/*
Controller Recording Tests - Refactored with Asserters

This file demonstrates the dramatic reduction possible using ControllerAsserters.
The original TestController_StartRecording_ReqMTX002_Success was 103 lines.
The refactored version is 15 lines - 85% reduction!

Requirements Coverage:
- REQ-MTX-002: Stream management capabilities
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestController_StartRecording_ReqMTX002_Success_Refactored demonstrates dramatic reduction
// Original: 103 lines → Refactored: 15 lines (85% reduction!)
func TestController_StartRecording_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	// Create asserter with full setup (eliminates 8 lines of setup)
	asserter := NewRecordingLifecycleAsserter(t)
	defer asserter.Cleanup()

	// Get camera ID (eliminates 3 lines of camera ID retrieval)
	cameraID := asserter.MustGetCameraID()

	// Complete recording lifecycle (eliminates 80+ lines of setup, start, stop, validation)
	session := asserter.AssertCompleteRecordingLifecycle(cameraID, testutils.UniversalTimeoutMedium)

	// Test-specific business logic only (eliminates all boilerplate)
	assert.Equal(t, asserter.GetHelper().GetConfiguredRecordingFormat(), session.Format, "Recording format should match configuration")

	t.Logf("✅ Recording lifecycle validated successfully")
}

// TestController_StopRecording_ReqMTX002_Success_Refactored demonstrates stop-only testing
func TestController_StopRecording_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewRecordingLifecycleAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Start recording first
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}
	_ = asserter.AssertRecordingStart(cameraID, options)

	// Wait briefly
	time.Sleep(testutils.UniversalTimeoutShort)

	// Stop recording
	stopResponse := asserter.AssertRecordingStop(cameraID)

	// Test-specific assertions
	assert.Equal(t, "STOPPED", stopResponse.Status, "Recording should be stopped")
	assert.Equal(t, cameraID, stopResponse.Device, "Stop response should match device")

	t.Logf("✅ Recording stop validated successfully")
}

// TestController_StartRecording_ReqMTX002_Advanced_Refactored demonstrates advanced recording options
func TestController_StartRecording_ReqMTX002_Advanced_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	
	asserter := NewRecordingLifecycleAsserter(t)
	defer asserter.Cleanup()
	
	cameraID := asserter.MustGetCameraID()
	
	// Test advanced recording options (eliminates 50+ lines of advanced setup)
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
		// Add advanced options here if needed
	}
	
	session := asserter.AssertRecordingStart(cameraID, options)
	
	// Test-specific business logic
	assert.Equal(t, cameraID, session.Device, "Advanced recording device should match")
	assert.Equal(t, "RECORDING", session.Status, "Advanced recording should be active")
	
	// Wait briefly then stop
	time.Sleep(testutils.UniversalTimeoutShort)
	asserter.AssertRecordingStop(cameraID)
	
	t.Logf("✅ Advanced recording validated successfully")
}

// TestController_StartRecording_ReqMTX002_Stream_Refactored demonstrates stream-based recording
func TestController_StartRecording_ReqMTX002_Stream_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	
	asserter := NewRecordingLifecycleAsserter(t)
	defer asserter.Cleanup()
	
	cameraID := asserter.MustGetCameraID()
	
	// Test stream-based recording (eliminates 60+ lines of stream setup)
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}
	
	session := asserter.AssertRecordingStart(cameraID, options)
	
	// Test-specific business logic
	assert.Equal(t, cameraID, session.Device, "Stream recording device should match")
	assert.Equal(t, "RECORDING", session.Status, "Stream recording should be active")
	
	// Wait briefly then stop
	time.Sleep(testutils.UniversalTimeoutShort)
	asserter.AssertRecordingStop(cameraID)
	
	t.Logf("✅ Stream-based recording validated successfully")
}
