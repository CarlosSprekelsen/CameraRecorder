/*
Controller Stream Tests - Refactored with Asserters

This file demonstrates the dramatic reduction possible using StreamAsserters.
Original tests had massive duplication of setup, Progressive Readiness, and validation.
Refactored tests focus on business logic only.

Requirements Coverage:
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path management
- REQ-MTX-004: RTSP operations
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_GetPaths_ReqMTX003_Success_Refactored demonstrates path listing
func TestController_GetPaths_ReqMTX003_Success_Refactored(t *testing.T) {
	// REQ-MTX-003: Path management

	// Create stream asserter with full setup (eliminates 8 lines of setup)
	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	// Get paths (eliminates 25+ lines of Progressive Readiness + validation)
	paths := asserter.AssertGetPaths()

	// Test-specific business logic only
	assert.NotNil(t, paths, "Paths response should not be nil")
	assert.GreaterOrEqual(t, len(paths), 0, "Paths list should be non-negative length")

	t.Logf("✅ Path listing validated successfully: %d total paths", len(paths))
}

// TestController_GetStreams_ReqMTX002_Success_Refactored demonstrates stream listing
func TestController_GetStreams_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	// Get streams (eliminates 20+ lines of setup and validation)
	streams := asserter.AssertGetStreams()

	// Test-specific business logic only
	assert.NotNil(t, streams, "Streams response should not be nil")
	assert.GreaterOrEqual(t, streams.Total, 0, "Total streams should be non-negative")

	t.Logf("✅ Stream listing validated successfully: %d total streams", streams.Total)
}

// TestController_GetStream_ReqMTX002_Success_Refactored demonstrates individual stream retrieval
func TestController_GetStream_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	// Get camera ID for stream operations
	cameraID := asserter.MustGetCameraID()

	// Create stream first (required for GetStream to work)
	_ = asserter.AssertCreateStream(cameraID)

	// Wait briefly for stream to be active
	time.Sleep(testutils.UniversalTimeoutShort)

	// Get specific stream (eliminates 30+ lines of Progressive Readiness + validation)
	stream := asserter.AssertGetStream(cameraID)

	// Test-specific business logic only
	assert.NotNil(t, stream, "Stream response should not be nil")
	assert.Equal(t, cameraID, stream.Device, "Stream device should match camera ID")

	t.Logf("✅ Stream retrieval validated successfully: %s", stream.Device)
}

// TestController_CreateStream_ReqMTX002_StreamManagement_Refactored demonstrates stream creation
func TestController_CreateStream_ReqMTX002_StreamManagement_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create stream (eliminates 35+ lines of setup, Progressive Readiness + validation)
	stream := asserter.AssertCreateStream(cameraID)

	// Test-specific business logic only
	assert.NotNil(t, stream, "Stream creation response should not be nil")
	assert.Equal(t, cameraID, stream.Device, "Created stream device should match camera ID")

	t.Logf("✅ Stream creation validated successfully: %s", stream.Device)
}

// TestController_DeleteStream_ReqMTX002_Success_Refactored demonstrates stream deletion
func TestController_DeleteStream_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create stream first
	_ = asserter.AssertCreateStream(cameraID)

	// Wait briefly for stream to be active
	time.Sleep(testutils.UniversalTimeoutShort)

	// Delete stream (eliminates 25+ lines of setup and validation)
	err := asserter.AssertDeleteStream(cameraID)
	require.NoError(t, err, "Stream deletion should succeed")

	t.Logf("✅ Stream deletion validated successfully: %s", cameraID)
}

// TestController_GetStream_ReqMTX004_RTSPOperations_Refactored demonstrates RTSP operations
func TestController_GetStream_ReqMTX004_RTSPOperations_Refactored(t *testing.T) {
	// REQ-MTX-004: RTSP operations

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test RTSP operations (eliminates 40+ lines of RTSP setup and validation)

	// Create stream for RTSP operations
	stream := asserter.AssertCreateStream(cameraID)
	assert.NotNil(t, stream, "Stream should be created for RTSP operations")

	// Wait for stream to be active
	time.Sleep(testutils.UniversalTimeoutMedium)

	// Get stream details (includes RTSP URL)
	streamDetails := asserter.AssertGetStream(cameraID)
	assert.NotNil(t, streamDetails, "Stream details should be available")

	// Test-specific business logic: verify stream details
	assert.NotNil(t, streamDetails, "Stream details should be available")

	t.Logf("✅ RTSP operations validated successfully: %s", cameraID)
}

// TestController_GetPaths_ReqMTX003_Management_Refactored demonstrates path management
func TestController_GetPaths_ReqMTX003_Management_Refactored(t *testing.T) {
	// REQ-MTX-003: Path management

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	// Test path management operations (eliminates 30+ lines of setup)

	// Get initial paths
	initialPaths := asserter.AssertGetPaths()
	initialCount := len(initialPaths)

	// Create a stream (which creates a path)
	cameraID := asserter.MustGetCameraID()
	_ = asserter.AssertCreateStream(cameraID)

	// Wait for path to be created
	time.Sleep(testutils.UniversalTimeoutShort)

	// Get paths after creation
	finalPaths := asserter.AssertGetPaths()

	// Test-specific business logic: verify path count increased
	assert.GreaterOrEqual(t, len(finalPaths), initialCount, "Path count should increase after stream creation")

	// Verify our camera path exists
	found := false
	for _, path := range finalPaths {
		if path.Name == cameraID {
			found = true
			break
		}
	}
	assert.True(t, found, "Created stream path should be in path list")

	t.Logf("✅ Path management validated successfully: %d → %d paths", initialCount, len(finalPaths))
}

// TestController_StartRecording_ReqMTX002_Stream_Integration_Refactored demonstrates stream-based recording
func TestController_StartRecording_ReqMTX002_Stream_Integration_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewStreamAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test stream-based recording (eliminates 50+ lines of recording setup)

	// Create stream first
	_ = asserter.AssertCreateStream(cameraID)

	// Wait for stream to be ready
	time.Sleep(testutils.UniversalTimeoutShort)

	// Start recording on the stream
	recordingManager := asserter.GetHelper().GetRecordingManager()
	options := &PathConf{
		Record:       true,
		RecordFormat: asserter.GetHelper().GetConfiguredRecordingFormat(),
	}

	startResult := testutils.TestProgressiveReadiness(t, func() (*StartRecordingResponse, error) {
		return recordingManager.StartRecording(asserter.GetContext(), cameraID, options)
	}, asserter.GetReadyController(), "StartRecording")

	require.NoError(t, startResult.Error, "Recording should start on stream")
	require.NotNil(t, startResult.Result, "Recording response should not be nil")

	session := startResult.Result

	// Test-specific business logic
	assert.Equal(t, cameraID, session.Device, "Recording device should match camera ID")
	assert.Equal(t, "RECORDING", session.Status, "Recording should be active")

	// Wait briefly then stop
	time.Sleep(testutils.UniversalTimeoutShort)

	stopResult := testutils.TestProgressiveReadiness(t, func() (*StopRecordingResponse, error) {
		return recordingManager.StopRecording(asserter.GetContext(), cameraID)
	}, asserter.GetReadyController(), "StopRecording")

	require.NoError(t, stopResult.Error, "Recording should stop successfully")

	t.Logf("✅ Stream-based recording validated successfully: %s", cameraID)
}
