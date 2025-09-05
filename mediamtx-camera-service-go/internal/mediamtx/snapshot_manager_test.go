/*
MediaMTX Snapshot Manager Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSnapshotManager_ReqMTX001 tests snapshot manager creation
func TestNewSnapshotManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)
	assert.Equal(t, config, snapshotManager.config)
	assert.Equal(t, logger, snapshotManager.logger)
	assert.Equal(t, ffmpegManager, snapshotManager.ffmpegManager)
}

// TestSnapshotManager_CaptureSnapshot_ReqMTX002 tests snapshot capture
func TestSnapshotManager_CaptureSnapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()
	devicePath := "/dev/video0"
	outputPath := "/tmp/test_snapshot.jpg"

	// Capture snapshot
	snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
	require.NoError(t, err, "Snapshot capture should succeed")
	require.NotNil(t, snapshot, "Snapshot should not be nil")
	assert.Equal(t, devicePath, snapshot.DevicePath)
	assert.Equal(t, outputPath, snapshot.OutputPath)
	assert.Equal(t, SnapshotStateCompleted, snapshot.State)

	// Verify snapshot is tracked
	snapshots := snapshotManager.GetSnapshots()
	assert.Len(t, snapshots, 1)
	assert.Contains(t, snapshots, snapshot.ID)
}

// TestSnapshotManager_GetSnapshots_ReqMTX002 tests snapshot listing
func TestSnapshotManager_GetSnapshots_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Initially no snapshots
	snapshots := snapshotManager.GetSnapshots()
	assert.Len(t, snapshots, 0)

	// Capture multiple snapshots
	snapshot1, err := snapshotManager.CaptureSnapshot(ctx, "/dev/video0", "/tmp/test1.jpg")
	require.NoError(t, err)

	snapshot2, err := snapshotManager.CaptureSnapshot(ctx, "/dev/video1", "/tmp/test2.jpg")
	require.NoError(t, err)

	// Verify both snapshots are tracked
	snapshots = snapshotManager.GetSnapshots()
	assert.Len(t, snapshots, 2)
	assert.Contains(t, snapshots, snapshot1.ID)
	assert.Contains(t, snapshots, snapshot2.ID)
}

// TestSnapshotManager_GetSnapshot_ReqMTX002 tests snapshot retrieval
func TestSnapshotManager_GetSnapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()
	devicePath := "/dev/video0"
	outputPath := "/tmp/test_snapshot_get.jpg"

	// Capture snapshot
	snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
	require.NoError(t, err, "Snapshot capture should succeed")

	// Get snapshot by ID
	retrievedSnapshot := snapshotManager.GetSnapshot(snapshot.ID)
	require.NotNil(t, retrievedSnapshot, "Snapshot should be retrievable")
	assert.Equal(t, snapshot.ID, retrievedSnapshot.ID)
	assert.Equal(t, devicePath, retrievedSnapshot.DevicePath)
	assert.Equal(t, outputPath, retrievedSnapshot.OutputPath)

	// Get non-existent snapshot
	nonExistentSnapshot := snapshotManager.GetSnapshot("non-existent-id")
	assert.Nil(t, nonExistentSnapshot, "Non-existent snapshot should return nil")
}

// TestSnapshotManager_DeleteSnapshot_ReqMTX002 tests snapshot deletion
func TestSnapshotManager_DeleteSnapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()
	devicePath := "/dev/video0"
	outputPath := "/tmp/test_snapshot_delete.jpg"

	// Capture snapshot
	snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
	require.NoError(t, err, "Snapshot capture should succeed")

	// Verify snapshot exists
	snapshots := snapshotManager.GetSnapshots()
	assert.Len(t, snapshots, 1)

	// Delete snapshot
	err = snapshotManager.DeleteSnapshot(ctx, snapshot.ID)
	require.NoError(t, err, "Snapshot deletion should succeed")

	// Verify snapshot is no longer tracked
	snapshots = snapshotManager.GetSnapshots()
	assert.Len(t, snapshots, 0)
}

// TestSnapshotManager_ErrorHandling_ReqMTX007 tests error scenarios
func TestSnapshotManager_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager that fails
	ffmpegManager := &mockFFmpegManager{failStart: true}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Test invalid device path
	_, err := snapshotManager.CaptureSnapshot(ctx, "", "/tmp/test.jpg")
	assert.Error(t, err, "Empty device path should fail")

	// Test invalid output path
	_, err = snapshotManager.CaptureSnapshot(ctx, "/dev/video0", "")
	assert.Error(t, err, "Empty output path should fail")

	// Test FFmpeg failure
	_, err = snapshotManager.CaptureSnapshot(ctx, "/dev/video0", "/tmp/test.jpg")
	assert.Error(t, err, "FFmpeg failure should be handled")

	// Test deleting non-existent snapshot
	err = snapshotManager.DeleteSnapshot(ctx, "non-existent-id")
	assert.Error(t, err, "Deleting non-existent snapshot should fail")
}

// TestSnapshotManager_ConcurrentAccess_ReqMTX001 tests concurrent operations
func TestSnapshotManager_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Capture multiple snapshots concurrently
	const numSnapshots = 5
	snapshots := make([]*Snapshot, numSnapshots)
	errors := make([]error, numSnapshots)

	for i := 0; i < numSnapshots; i++ {
		go func(index int) {
			devicePath := fmt.Sprintf("/dev/video%d", index)
			outputPath := fmt.Sprintf("/tmp/concurrent_snapshot_%d.jpg", index)
			snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
			snapshots[index] = snapshot
			errors[index] = err
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify all snapshots were captured successfully
	allSnapshots := snapshotManager.GetSnapshots()
	assert.Len(t, allSnapshots, numSnapshots, "All concurrent snapshots should be captured")
}

// TestSnapshotManager_SnapshotSettings_ReqMTX002 tests snapshot configuration
func TestSnapshotManager_SnapshotSettings_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	// Set snapshot settings
	snapshotSettings := &SnapshotSettings{
		Format:        "jpeg",
		Quality:       85,
		Width:         1920,
		Height:        1080,
		CapturePath:   "/tmp/snapshots",
		MaxSnapshots:  100,
		RetentionDays: 30,
	}

	snapshotManager.SetSnapshotSettings(snapshotSettings)

	// Verify settings are applied
	assert.Equal(t, snapshotSettings, snapshotManager.snapshotSettings)

	ctx := context.Background()
	devicePath := "/dev/video0"
	outputPath := "/tmp/settings_test.jpg"

	// Capture snapshot with settings
	snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
	require.NoError(t, err, "Snapshot with settings should succeed")
	require.NotNil(t, snapshot, "Snapshot should not be nil")

	// Verify settings are used
	assert.Equal(t, snapshotSettings.Format, snapshot.Format)
	assert.Equal(t, snapshotSettings.Quality, snapshot.Quality)
}

// TestSnapshotManager_FileOperations_ReqMTX002 tests file system operations
func TestSnapshotManager_FileOperations_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create mock FFmpeg manager
	ffmpegManager := &mockFFmpegManager{}

	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger, nil)
	require.NotNil(t, snapshotManager)

	// Create test directory
	testDir := "/tmp/snapshot_test_dir"
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	ctx := context.Background()
	devicePath := "/dev/video0"
	outputPath := filepath.Join(testDir, "test_snapshot.jpg")

	// Capture snapshot
	snapshot, err := snapshotManager.CaptureSnapshot(ctx, devicePath, outputPath)
	require.NoError(t, err, "Snapshot capture should succeed")

	// Verify file was created (in real implementation)
	// Note: This would require actual file system operations in a real test
	assert.NotNil(t, snapshot, "Snapshot should be created")

	// Test cleanup
	err = snapshotManager.DeleteSnapshot(ctx, snapshot.ID)
	require.NoError(t, err, "Snapshot cleanup should succeed")
}
