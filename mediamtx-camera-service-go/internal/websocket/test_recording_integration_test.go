/*
WebSocket Recording Integration Tests - Stateless Recording Architecture

Tests comprehensive recording workflows including start/stop recording, recording status
queries, file management operations, and recording cleanup. Validates the complete
stateless recording architecture with real components and performance targets.

API Documentation Reference: docs/api/json_rpc_methods.md
Requirements Coverage:
- REQ-WS-002: Real-time camera operations
- REQ-API-002: JSON-RPC 2.0 protocol compliance
- REQ-API-003: Request/response message handling
- REQ-REC-001: Stateless recording architecture
- REQ-REC-002: Recording file management
- REQ-PERF-001: Recording performance targets

Design Principles:
- Real components only (no mocks)
- Stateless recording architecture validation
- Performance target validation
- Complete API specification compliance
- File management and cleanup testing
- Progressive Readiness pattern validation
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// STATELESS RECORDING ARCHITECTURE TESTS
// ============================================================================

// TestRecording_FileLifecycle_Complete_Integration tests complete recording file lifecycle
func TestRecording_FileLifecycle_Complete_Integration(t *testing.T) {
	// Use testutils for comprehensive setup
	dvh := testutils.NewDataValidationHelper(t)

	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test complete recording file lifecycle using testutils validation
	err := asserter.AssertFileLifecycleWorkflow()
	require.NoError(t, err, "Recording file lifecycle workflow should succeed")

	// Validate testutils integration
	require.NotNil(t, dvh, "DataValidationHelper should be created")
}

// TestRecording_FileValidation_Comprehensive_Integration tests recording file validation using testutils
func TestRecording_FileValidation_Comprehensive_Integration(t *testing.T) {
	// Use testutils for comprehensive setup
	dvh := testutils.NewDataValidationHelper(t)

	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test recording validation using testutils
	err := asserter.AssertRecordingWorkflow()
	require.NoError(t, err, "Recording workflow should succeed")

	// Validate testutils integration
	require.NotNil(t, dvh, "DataValidationHelper should be created")
}

// TestRecording_StatelessArchitecture_Integration validates the stateless recording
// architecture where MediaMTX is the source of truth
func TestRecording_StatelessArchitecture_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test stateless recording architecture
	err = asserter.AssertStatelessRecordingArchitecture()
	require.NoError(t, err, "Stateless recording architecture should work")

	t.Log("✅ Stateless recording architecture validated")
}

// TestRecording_MediaMTXSourceOfTruth_Integration validates that MediaMTX is the
// source of truth for recording status
func TestRecording_MediaMTXSourceOfTruth_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test MediaMTX as source of truth
	err = asserter.AssertMediaMTXSourceOfTruth()
	require.NoError(t, err, "MediaMTX source of truth should work")

	t.Log("✅ MediaMTX source of truth validated")
}

// ============================================================================
// RECORDING WORKFLOW TESTS
// ============================================================================

// TestRecording_BasicWorkflow_Integration validates basic recording start/stop workflow
func TestRecording_BasicWorkflow_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test basic recording workflow
	err = asserter.AssertRecordingWorkflow()
	require.NoError(t, err, "Basic recording workflow should work")

	t.Log("✅ Basic recording workflow validated")
}

// TestRecording_StartRecording_Integration validates start recording functionality
func TestRecording_StartRecording_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test start recording
	err = asserter.AssertStartRecording()
	require.NoError(t, err, "Start recording should work")

	t.Log("✅ Start recording validated")
}

// TestRecording_StopRecording_Integration validates stop recording functionality
func TestRecording_StopRecording_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test stop recording
	err = asserter.AssertStopRecording()
	require.NoError(t, err, "Stop recording should work")

	t.Log("✅ Stop recording validated")
}

// TestRecording_RecordingStatus_Integration validates recording status queries
func TestRecording_RecordingStatus_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test recording status queries
	err = asserter.AssertRecordingStatus()
	require.NoError(t, err, "Recording status should work")

	t.Log("✅ Recording status validated")
}

// ============================================================================
// RECORDING DURATION AND FORMAT TESTS
// ============================================================================

// TestRecording_DurationManagement_Integration validates recording duration management
func TestRecording_DurationManagement_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test duration management
	err = asserter.AssertRecordingDurationManagement()
	require.NoError(t, err, "Recording duration management should work")

	t.Log("✅ Recording duration management validated")
}

// TestRecording_FormatSupport_Integration validates recording format support
func TestRecording_FormatSupport_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test format support
	err = asserter.AssertRecordingFormatSupport()
	require.NoError(t, err, "Recording format support should work")

	t.Log("✅ Recording format support validated")
}

// TestRecording_STANAG4609Compliance_Integration validates STANAG 4609 compliance
func TestRecording_STANAG4609Compliance_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test STANAG 4609 compliance
	err = asserter.AssertSTANAG4609Compliance()
	require.NoError(t, err, "STANAG 4609 compliance should work")

	t.Log("✅ STANAG 4609 compliance validated")
}

// ============================================================================
// FILE MANAGEMENT TESTS
// ============================================================================

// TestRecording_FileManagement_Integration validates recording file management operations
func TestRecording_FileManagement_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test recording file management
	err = asserter.AssertRecordingFileManagement()
	require.NoError(t, err, "Recording file management should work")

	t.Log("✅ Recording file management validated")
}

// TestRecording_FileListing_Integration validates recording file listing
func TestRecording_FileListing_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test recording file listing
	err = asserter.AssertRecordingFileListing()
	require.NoError(t, err, "Recording file listing should work")

	t.Log("✅ Recording file listing validated")
}

// TestRecording_FileCleanup_Integration validates recording file cleanup operations
func TestRecording_FileCleanup_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test recording file cleanup
	err = asserter.AssertRecordingFileCleanup()
	require.NoError(t, err, "Recording file cleanup should work")

	t.Log("✅ Recording file cleanup validated")
}

// ============================================================================
// CONCURRENT RECORDING TESTS
// ============================================================================

// TestRecording_ConcurrentRecordings_Integration validates concurrent recording operations
func TestRecording_ConcurrentRecordings_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test concurrent recordings
	err = asserter.AssertConcurrentRecordings()
	require.NoError(t, err, "Concurrent recordings should work")

	t.Log("✅ Concurrent recordings validated")
}

// TestRecording_MultipleCameras_Integration validates recording from multiple cameras
func TestRecording_MultipleCameras_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test multiple cameras recording
	err = asserter.AssertMultipleCamerasRecording()
	require.NoError(t, err, "Multiple cameras recording should work")

	t.Log("✅ Multiple cameras recording validated")
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

// TestRecording_InvalidDevice_Integration validates error handling for invalid devices
func TestRecording_InvalidDevice_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test invalid device error handling
	err = asserter.AssertInvalidDeviceRecording()
	require.NoError(t, err, "Invalid device error handling should work")

	t.Log("✅ Invalid device error handling validated")
}

// TestRecording_UnauthorizedAccess_Integration validates authorization for recording operations
func TestRecording_UnauthorizedAccess_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test unauthorized access error handling
	err = asserter.AssertUnauthorizedRecordingAccess()
	require.NoError(t, err, "Unauthorized access error handling should work")

	t.Log("✅ Unauthorized access error handling validated")
}

// TestRecording_NetworkErrorRecovery_Integration validates network error recovery
func TestRecording_NetworkErrorRecovery_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test network error recovery
	err = asserter.AssertRecordingNetworkErrorRecovery()
	require.NoError(t, err, "Network error recovery should work")

	t.Log("✅ Network error recovery validated")
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

// TestRecording_PerformanceTargets_Integration validates recording performance targets
func TestRecording_PerformanceTargets_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test performance targets
	err = asserter.AssertRecordingPerformanceTargets()
	require.NoError(t, err, "Recording performance targets should work")

	t.Log("✅ Recording performance targets validated")
}

// TestRecording_LoadTesting_Integration validates recording operations under load
func TestRecording_LoadTesting_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := testutils.GetSharedWebSocketAsserter(t)

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test load testing
	err = asserter.AssertRecordingLoadTesting()
	require.NoError(t, err, "Recording load testing should work")

	t.Log("✅ Recording load testing validated")
}
