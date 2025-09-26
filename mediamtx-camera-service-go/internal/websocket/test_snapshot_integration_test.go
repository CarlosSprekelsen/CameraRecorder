/*
WebSocket Snapshot Integration Tests - Multi-Tier Snapshot Architecture

Tests comprehensive snapshot capture workflows including multi-tier performance,
file management operations, and snapshot cleanup. Validates the complete snapshot
architecture with real components and performance targets.

API Documentation Reference: docs/api/json_rpc_methods.md
Requirements Coverage:
- REQ-WS-002: Real-time camera operations
- REQ-API-002: JSON-RPC 2.0 protocol compliance
- REQ-API-003: Request/response message handling
- REQ-PERF-001: Multi-tier snapshot performance targets
- REQ-FILE-001: File management operations

Design Principles:
- Real components only (no mocks)
- Multi-tier snapshot architecture validation
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
// MULTI-TIER SNAPSHOT PERFORMANCE TESTS
// ============================================================================

// TestSnapshot_FileLifecycle_Complete_Integration tests complete snapshot file lifecycle
func TestSnapshot_FileLifecycle_Complete_Integration(t *testing.T) {
	// Use testutils for comprehensive setup
	dvh := testutils.NewDataValidationHelper(t)

	asserter := NewWebSocketIntegrationAsserter(t)

	// Test complete snapshot file lifecycle using testutils validation
	err := asserter.AssertFileLifecycleWorkflow()
	require.NoError(t, err, "Snapshot file lifecycle workflow should succeed")

	// Validate testutils integration
	require.NotNil(t, dvh, "DataValidationHelper should be created")
}

// TestSnapshot_FileValidation_Comprehensive_Integration tests snapshot file validation using testutils
func TestSnapshot_FileValidation_Comprehensive_Integration(t *testing.T) {
	// Use testutils for comprehensive setup
	dvh := testutils.NewDataValidationHelper(t)

	asserter := NewWebSocketIntegrationAsserter(t)

	// Test snapshot validation using testutils
	err := asserter.AssertSnapshotWorkflow()
	require.NoError(t, err, "Snapshot workflow should succeed")

	// Validate testutils integration
	require.NotNil(t, dvh, "DataValidationHelper should be created")
}

// TestSnapshot_MultiTierPerformance_Integration validates the multi-tier snapshot
// architecture performance targets with real components
func TestSnapshot_MultiTierPerformance_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test multi-tier snapshot performance
	err = asserter.AssertMultiTierSnapshotPerformance()
	require.NoError(t, err, "Multi-tier snapshot performance should meet targets")

	t.Log("✅ Multi-tier snapshot performance validated")
}

// TestSnapshot_Tier0DirectV4L2_Integration validates Tier 0 direct V4L2 capture
// performance (<100ms) for USB devices
func TestSnapshot_Tier0DirectV4L2_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test Tier 0 direct V4L2 capture
	err = asserter.AssertTier0DirectV4L2Capture()
	require.NoError(t, err, "Tier 0 direct V4L2 capture should work")

	t.Log("✅ Tier 0 direct V4L2 capture validated")
}

// TestSnapshot_Tier1FFmpegDirect_Integration validates Tier 1 FFmpeg direct capture
// performance (<200ms) when device is accessible
func TestSnapshot_Tier1FFmpegDirect_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test Tier 1 FFmpeg direct capture
	err = asserter.AssertTier1FFmpegDirectCapture()
	require.NoError(t, err, "Tier 1 FFmpeg direct capture should work")

	t.Log("✅ Tier 1 FFmpeg direct capture validated")
}

// TestSnapshot_Tier2RTSPReuse_Integration validates Tier 2 RTSP stream reuse
// performance (<300ms) when stream is already active
func TestSnapshot_Tier2RTSPReuse_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test Tier 2 RTSP stream reuse
	err = asserter.AssertTier2RTSPReuse()
	require.NoError(t, err, "Tier 2 RTSP stream reuse should work")

	t.Log("✅ Tier 2 RTSP stream reuse validated")
}

// TestSnapshot_Tier3StreamActivation_Integration validates Tier 3 stream activation
// performance (<500ms) when creating new MediaMTX path
func TestSnapshot_Tier3StreamActivation_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test Tier 3 stream activation
	err = asserter.AssertTier3StreamActivation()
	require.NoError(t, err, "Tier 3 stream activation should work")

	t.Log("✅ Tier 3 stream activation validated")
}

// ============================================================================
// SNAPSHOT WORKFLOW TESTS
// ============================================================================

// TestSnapshot_BasicWorkflow_Integration validates basic snapshot capture workflow
func TestSnapshot_BasicWorkflow_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test basic snapshot workflow
	err = asserter.AssertSnapshotWorkflow()
	require.NoError(t, err, "Basic snapshot workflow should work")

	t.Log("✅ Basic snapshot workflow validated")
}

// TestSnapshot_CustomFilename_Integration validates snapshot capture with custom filename
func TestSnapshot_CustomFilename_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test custom filename snapshot
	err = asserter.AssertCustomFilenameSnapshot()
	require.NoError(t, err, "Custom filename snapshot should work")

	t.Log("✅ Custom filename snapshot validated")
}

// TestSnapshot_ConcurrentCaptures_Integration validates concurrent snapshot captures
func TestSnapshot_ConcurrentCaptures_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test concurrent snapshot captures
	err = asserter.AssertConcurrentSnapshotCaptures()
	require.NoError(t, err, "Concurrent snapshot captures should work")

	t.Log("✅ Concurrent snapshot captures validated")
}

// ============================================================================
// FILE MANAGEMENT TESTS
// ============================================================================

// TestSnapshot_FileManagement_Integration validates snapshot file management operations
func TestSnapshot_FileManagement_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test snapshot file management
	err = asserter.AssertSnapshotFileManagement()
	require.NoError(t, err, "Snapshot file management should work")

	t.Log("✅ Snapshot file management validated")
}

// TestSnapshot_FileCleanup_Integration validates snapshot file cleanup operations
func TestSnapshot_FileCleanup_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test snapshot file cleanup
	err = asserter.AssertSnapshotFileCleanup()
	require.NoError(t, err, "Snapshot file cleanup should work")

	t.Log("✅ Snapshot file cleanup validated")
}

// TestSnapshot_StorageInfo_Integration validates snapshot storage information
func TestSnapshot_StorageInfo_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test snapshot storage info
	err = asserter.AssertSnapshotStorageInfo()
	require.NoError(t, err, "Snapshot storage info should work")

	t.Log("✅ Snapshot storage info validated")
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

// TestSnapshot_InvalidDevice_Integration validates error handling for invalid devices
func TestSnapshot_InvalidDevice_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test invalid device error handling
	err = asserter.AssertInvalidDeviceSnapshot()
	require.NoError(t, err, "Invalid device error handling should work")

	t.Log("✅ Invalid device error handling validated")
}

// TestSnapshot_UnauthorizedAccess_Integration validates authorization for snapshot operations
func TestSnapshot_UnauthorizedAccess_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test unauthorized access error handling
	err = asserter.AssertUnauthorizedSnapshotAccess()
	require.NoError(t, err, "Unauthorized access error handling should work")

	t.Log("✅ Unauthorized access error handling validated")
}

// TestSnapshot_NetworkErrorRecovery_Integration validates network error recovery
func TestSnapshot_NetworkErrorRecovery_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test network error recovery
	err = asserter.AssertNetworkErrorRecovery()
	require.NoError(t, err, "Network error recovery should work")

	t.Log("✅ Network error recovery validated")
}

// ============================================================================
// PERFORMANCE BENCHMARK TESTS
// ============================================================================

// TestSnapshot_PerformanceBenchmarks_Integration validates snapshot performance benchmarks
func TestSnapshot_PerformanceBenchmarks_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test performance benchmarks
	err = asserter.AssertSnapshotPerformanceBenchmarks()
	require.NoError(t, err, "Snapshot performance benchmarks should work")

	t.Log("✅ Snapshot performance benchmarks validated")
}

// TestSnapshot_LoadTesting_Integration validates snapshot operations under load
func TestSnapshot_LoadTesting_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.Cleanup()

	// Test Progressive Readiness: immediate connection acceptance
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test load testing
	err = asserter.AssertSnapshotLoadTesting()
	require.NoError(t, err, "Snapshot load testing should work")

	t.Log("✅ Snapshot load testing validated")
}
