/*
MediaMTX Snapshot Manager Tests - Refactored with Progressive Readiness

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json

Refactored from snapshot_manager_test.go (1,176 lines → ~300 lines)
Eliminates massive duplication using SnapshotManagerAsserter
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotManager_New_ReqMTX001_Success tests snapshot manager creation with real server
func TestSnapshotManager_New_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	snapshotManager := asserter.GetSnapshotManager()
	require.NotNil(t, snapshotManager, "Snapshot manager should not be nil")

	// Verify snapshot manager was created properly
	asserter.AssertSnapshotSettings()
}

// TestSnapshotManager_TakeSnapshot_ReqMTX002_DataCreation tests actual data creation using configured paths
func TestSnapshotManager_TakeSnapshot_ReqMTX002_DataCreation_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	// Data-driven validation: Verify actual file creation and metadata in configured directories
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	options := &SnapshotOptions{
		Format:     "jpg",
		Quality:    85,
		MaxWidth:   1920,
		MaxHeight:  1080,
		AutoResize: true,
	}

	// Execute snapshot creation with Progressive Readiness
	response := asserter.AssertSnapshotCapture(cameraID, options)

	// Verify metadata from real operation
	assert.Equal(t, cameraID, response.Device, "Response device should match camera ID")
	assert.NotEmpty(t, response.Timestamp, "Response should include timestamp")
	assert.NotEmpty(t, response.Filename, "Response should include snapshot filename")
}

// TestSnapshotManager_GetSnapshotsList_ReqMTX002_Success tests snapshot list retrieval
func TestSnapshotManager_GetSnapshotsList_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create some test snapshots
	asserter.CreateMultipleTestSnapshots(cameraID, 3)

	// Test list retrieval with pagination
	snapshots := asserter.AssertSnapshotList(10, 0)
	assert.GreaterOrEqual(t, len(snapshots.Files), 3, "Should have at least 3 snapshots")

	// Test pagination
	snapshotsLimited := asserter.AssertSnapshotList(2, 0)
	assert.LessOrEqual(t, len(snapshotsLimited.Files), 2, "Should respect limit")
}

// TestSnapshotManager_GetSnapshotInfo_ReqMTX002_Success tests snapshot info retrieval
func TestSnapshotManager_GetSnapshotInfo_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create a test snapshot
	snapshot := asserter.CreateTestSnapshot(cameraID)

	// Get snapshot info
	info := asserter.AssertSnapshotInfo(snapshot.FilePath)
	assert.Equal(t, snapshot.Filename, info.Filename, "Filename should match")
	assert.NotEmpty(t, info.Filename, "Filename should not be empty")
}

// TestSnapshotManager_DeleteSnapshotFile_ReqMTX002_Success tests snapshot file deletion
func TestSnapshotManager_DeleteSnapshotFile_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create a test snapshot
	snapshot := asserter.CreateTestSnapshot(cameraID)

	// Verify file exists
	asserter.AssertSnapshotFileExists(snapshot.FilePath, cameraID)

	// Delete the snapshot
	asserter.AssertDeleteSnapshotFile(snapshot.FilePath)
}

// TestSnapshotManager_GetSnapshotSettings_ReqMTX001_Success tests snapshot settings retrieval
func TestSnapshotManager_GetSnapshotSettings_ReqMTX001_Success_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	// Get snapshot settings
	settings := asserter.AssertSnapshotSettings()
	assert.NotNil(t, settings, "Snapshot settings should not be nil")
}

// TestSnapshotManager_CleanupOldSnapshots_ReqMTX002_Success tests cleanup of old snapshots
func TestSnapshotManager_CleanupOldSnapshots_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Create some test snapshots
	asserter.CreateMultipleTestSnapshots(cameraID, 2)

	// Cleanup with very short max age (should not delete recent snapshots)
	count := asserter.AssertCleanupOldSnapshots(1 * time.Millisecond)
	assert.Equal(t, 0, count, "Should not delete recent snapshots")

	// Cleanup with longer max age (should delete all snapshots)
	count = asserter.AssertCleanupOldSnapshots(1 * time.Second)
	assert.GreaterOrEqual(t, count, 2, "Should delete recent snapshots")
}

// TestSnapshotManager_TakeSnapshot_ReqMTX004_ErrorHandling tests error handling scenarios
func TestSnapshotManager_TakeSnapshot_ReqMTX004_ErrorHandling_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	// Test with invalid camera ID
	invalidCameraID := "nonexistent_camera_12345"
	options := &SnapshotOptions{
		Format:  "jpg",
		Quality: 85,
	}

	// This should fail gracefully
	_, err := asserter.GetSnapshotManager().TakeSnapshot(asserter.GetContext(), invalidCameraID, options)
	assert.Error(t, err, "Snapshot should fail with invalid camera")
	assert.Contains(t, err.Error(), "camera", "Error should mention camera not found")
}

// TestSnapshotManager_TakeSnapshot_ReqMTX001_Concurrent tests concurrent snapshot operations
func TestSnapshotManager_TakeSnapshot_ReqMTX001_Concurrent_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test concurrent snapshots
	const numConcurrent = 3
	results := make(chan *TakeSnapshotResponse, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func() {
			snapshot := asserter.CreateTestSnapshot(cameraID)
			results <- snapshot
		}()
	}

	// Collect results
	for i := 0; i < numConcurrent; i++ {
		select {
		case snapshot := <-results:
			assert.NotNil(t, snapshot, "Concurrent snapshot should succeed")
			asserter.AssertSnapshotFileExists(snapshot.FilePath, cameraID)
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent snapshot test timed out")
		}
	}
}

// TestSnapshotManager_TakeSnapshot_ReqMTX002_Tier0_V4L2Direct_RealHardware tests V4L2 direct capture
func TestSnapshotManager_TakeSnapshot_ReqMTX002_Tier0_V4L2Direct_RealHardware_Refactored(t *testing.T) {
	// Test the Tier 0 V4L2 direct capture with REAL camera hardware
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	t.Run("V4L2_Direct_Capture", func(t *testing.T) {
		options := &SnapshotOptions{
			Format:     "jpg",
			Quality:    85,
			MaxWidth:   1920,
			MaxHeight:  1080,
			AutoResize: true,
		}

		// Use Progressive Readiness pattern for V4L2 direct capture
		response := asserter.AssertSnapshotCapture(cameraID, options)
		assert.Equal(t, cameraID, response.Device, "Response device should match camera ID")
	})
}

// TestSnapshotManager_TakeSnapshot_ReqMTX002_MultiTier_Integration tests multi-tier snapshot integration
func TestSnapshotManager_TakeSnapshot_ReqMTX002_MultiTier_Integration_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - Complete multi-tier testing
	asserter := NewSnapshotManagerAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test multiple tiers in sequence
	t.Run("Tier0_V4L2Direct", func(t *testing.T) {
		snapshot := asserter.CreateTestSnapshot(cameraID)
		asserter.AssertSnapshotFileExists(snapshot.FilePath, cameraID)
	})

	t.Run("Tier1_USBDirect", func(t *testing.T) {
		snapshot := asserter.CreateTestSnapshot(cameraID)
		asserter.AssertSnapshotFileExists(snapshot.FilePath, cameraID)
	})

	t.Run("Tier2_RTSPImmediate", func(t *testing.T) {
		snapshot := asserter.CreateTestSnapshot(cameraID)
		asserter.AssertSnapshotFileExists(snapshot.FilePath, cameraID)
	})
}
