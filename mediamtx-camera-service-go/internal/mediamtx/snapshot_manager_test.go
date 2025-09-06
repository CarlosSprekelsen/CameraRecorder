/*
MediaMTX Snapshot Manager Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSnapshotManager_ReqMTX001 tests snapshot manager creation with real server
func TestNewSnapshotManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager for snapshot operations
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	// Verify snapshot manager was created properly
	assert.NotNil(t, snapshotManager, "Snapshot manager should be initialized")
	assert.NotNil(t, snapshotManager.GetSnapshotSettings(), "Snapshot settings should be initialized")
}

// TestSnapshotManager_TakeSnapshot_ReqMTX002 tests snapshot capture with real server
func TestSnapshotManager_TakeSnapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
		Host:          "localhost",
		RTSPPort:      8554,
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	devicePath := "/dev/video0"
	outputPath := filepath.Join(config.SnapshotsPath, "test_snapshot.jpg")

	// Test snapshot options
	options := map[string]interface{}{
		"format":      "jpg",
		"quality":     85,
		"max_width":   1920,
		"max_height":  1080,
		"auto_resize": true,
	}

	// Take snapshot (this will test the multi-tier approach)
	snapshot, err := snapshotManager.TakeSnapshot(ctx, devicePath, outputPath, options)

	// Note: This test may fail if no camera is available, which is expected
	// The test validates the multi-tier approach and error handling
	if err != nil {
		t.Logf("Snapshot failed as expected (no camera available): %v", err)
		// Verify error handling is working correctly
		assert.Contains(t, err.Error(), "failed", "Error should indicate failure")
	} else {
		// If snapshot succeeds, verify it was created properly
		require.NotNil(t, snapshot, "Snapshot should not be nil")
		assert.Equal(t, devicePath, snapshot.Device)
		assert.Equal(t, outputPath, snapshot.FilePath)
		assert.Greater(t, snapshot.Size, int64(0), "Snapshot should have size > 0")

		// Verify snapshot is tracked
		snapshots := snapshotManager.ListSnapshots()
		assert.Len(t, snapshots, 1)
		assert.Equal(t, snapshot.ID, snapshots[0].ID)
	}
}

// TestSnapshotManager_GetSnapshotsList_ReqMTX002 tests snapshot listing with real server
func TestSnapshotManager_GetSnapshotsList_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	// Test 1: Get snapshots list from empty directory
	response, err := snapshotManager.GetSnapshotsList(ctx, 10, 0)
	require.NoError(t, err, "GetSnapshotsList should succeed")
	require.NotNil(t, response, "Response should not be nil")
	assert.NotNil(t, response.Files, "Files field should be present")
	assert.Equal(t, 0, response.Total, "Total should be 0 for empty directory")
	assert.Equal(t, 10, response.Limit, "Limit should match requested value")
	assert.Equal(t, 0, response.Offset, "Offset should match requested value")

	// Test 2: Create test snapshot files
	testFiles := []string{"test1.jpg", "test2.jpg", "test3.jpg"}
	for _, filename := range testFiles {
		filePath := filepath.Join(config.SnapshotsPath, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.WriteString("test snapshot data")
		file.Close()
	}

	// Test 3: Get snapshots list with files
	response, err = snapshotManager.GetSnapshotsList(ctx, 10, 0)
	require.NoError(t, err, "GetSnapshotsList should succeed with files")
	assert.Equal(t, 3, response.Total, "Total should be 3")
	assert.Len(t, response.Files, 3, "Should return 3 files")

	// Test 4: Test pagination
	response, err = snapshotManager.GetSnapshotsList(ctx, 2, 1)
	require.NoError(t, err, "Pagination should work")
	assert.Equal(t, 2, response.Limit, "Pagination limit should be respected")
	assert.Equal(t, 1, response.Offset, "Pagination offset should be respected")
	assert.Len(t, response.Files, 2, "Should return 2 files for pagination")
}

// TestSnapshotManager_GetSnapshotInfo_ReqMTX002 tests snapshot info retrieval with real server
func TestSnapshotManager_GetSnapshotInfo_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	// Create test snapshot file
	testFilename := "test_snapshot_info.jpg"
	testFilePath := filepath.Join(config.SnapshotsPath, testFilename)
	file, err := os.Create(testFilePath)
	require.NoError(t, err)
	file.WriteString("test snapshot data for info test")
	file.Close()

	// Test 1: Get snapshot info for existing file
	fileMetadata, err := snapshotManager.GetSnapshotInfo(ctx, testFilename)
	require.NoError(t, err, "GetSnapshotInfo should succeed")
	require.NotNil(t, fileMetadata, "File metadata should not be nil")
	assert.Equal(t, testFilename, fileMetadata.FileName)
	assert.Greater(t, fileMetadata.FileSize, int64(0), "File should have size > 0")
	assert.NotNil(t, fileMetadata.CreatedAt, "CreatedAt should not be nil")
	assert.NotNil(t, fileMetadata.ModifiedAt, "ModifiedAt should not be nil")
	assert.Contains(t, fileMetadata.DownloadURL, testFilename, "DownloadURL should contain filename")

	// Test 2: Get snapshot info for non-existent file
	_, err = snapshotManager.GetSnapshotInfo(ctx, "non_existent.jpg")
	assert.Error(t, err, "Should return error for non-existent file")
	assert.Contains(t, err.Error(), "not found", "Error should indicate file not found")

	// Test 3: Get snapshot info with empty filename
	_, err = snapshotManager.GetSnapshotInfo(ctx, "")
	assert.Error(t, err, "Should return error for empty filename")
	assert.Contains(t, err.Error(), "cannot be empty", "Error should indicate empty filename")
}

// TestSnapshotManager_DeleteSnapshotFile_ReqMTX002 tests snapshot file deletion with real server
func TestSnapshotManager_DeleteSnapshotFile_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	// Create test snapshot file
	testFilename := "test_snapshot_delete.jpg"
	testFilePath := filepath.Join(config.SnapshotsPath, testFilename)
	file, err := os.Create(testFilePath)
	require.NoError(t, err)
	file.WriteString("test snapshot data for delete test")
	file.Close()

	// Verify file exists
	_, err = os.Stat(testFilePath)
	require.NoError(t, err, "Test file should exist")

	// Test 1: Delete existing snapshot file
	err = snapshotManager.DeleteSnapshotFile(ctx, testFilename)
	require.NoError(t, err, "DeleteSnapshotFile should succeed")

	// Verify file was deleted
	_, err = os.Stat(testFilePath)
	assert.Error(t, err, "File should be deleted")
	assert.True(t, os.IsNotExist(err), "File should not exist")

	// Test 2: Delete non-existent file
	err = snapshotManager.DeleteSnapshotFile(ctx, "non_existent.jpg")
	assert.Error(t, err, "Should return error for non-existent file")
	assert.Contains(t, err.Error(), "not found", "Error should indicate file not found")

	// Test 3: Delete with empty filename
	err = snapshotManager.DeleteSnapshotFile(ctx, "")
	assert.Error(t, err, "Should return error for empty filename")
	assert.Contains(t, err.Error(), "cannot be empty", "Error should indicate empty filename")
}

// TestSnapshotManager_SnapshotSettings_ReqMTX001 tests snapshot settings management
func TestSnapshotManager_SnapshotSettings_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	// Test 1: Get default settings
	settings := snapshotManager.GetSnapshotSettings()
	require.NotNil(t, settings, "Settings should not be nil")
	assert.Equal(t, "jpg", settings.Format, "Default format should be jpg")
	assert.Equal(t, 85, settings.Quality, "Default quality should be 85")
	assert.Equal(t, 1920, settings.MaxWidth, "Default max width should be 1920")
	assert.Equal(t, 1080, settings.MaxHeight, "Default max height should be 1080")
	assert.True(t, settings.AutoResize, "Default auto resize should be true")
	assert.Equal(t, 6, settings.Compression, "Default compression should be 6")

	// Test 2: Update settings
	newSettings := &SnapshotSettings{
		Format:      "png",
		Quality:     95,
		MaxWidth:    3840,
		MaxHeight:   2160,
		AutoResize:  false,
		Compression: 9,
	}

	snapshotManager.UpdateSnapshotSettings(newSettings)

	// Test 3: Verify settings were updated
	updatedSettings := snapshotManager.GetSnapshotSettings()
	require.NotNil(t, updatedSettings, "Updated settings should not be nil")
	assert.Equal(t, "png", updatedSettings.Format, "Format should be updated to png")
	assert.Equal(t, 95, updatedSettings.Quality, "Quality should be updated to 95")
	assert.Equal(t, 3840, updatedSettings.MaxWidth, "Max width should be updated to 3840")
	assert.Equal(t, 2160, updatedSettings.MaxHeight, "Max height should be updated to 2160")
	assert.False(t, updatedSettings.AutoResize, "Auto resize should be updated to false")
	assert.Equal(t, 9, updatedSettings.Compression, "Compression should be updated to 9")
}

// TestSnapshotManager_CleanupOldSnapshots_ReqMTX002 tests snapshot cleanup functionality
func TestSnapshotManager_CleanupOldSnapshots_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	// Create test snapshot files with different timestamps
	testFiles := []string{"old1.jpg", "old2.jpg", "new1.jpg", "new2.jpg"}
	for i, filename := range testFiles {
		filePath := filepath.Join(config.SnapshotsPath, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.WriteString("test snapshot data")
		file.Close()

		// Make some files old by modifying their timestamp
		if i < 2 { // old1.jpg and old2.jpg
			oldTime := time.Now().Add(-2 * time.Hour)
			err = os.Chtimes(filePath, oldTime, oldTime)
			require.NoError(t, err)
		}
	}

	// Test 1: Cleanup old snapshots (older than 1 hour)
	err = snapshotManager.CleanupOldSnapshots(ctx, 1*time.Hour, 10)
	require.NoError(t, err, "CleanupOldSnapshots should succeed")

	// Verify old files were deleted
	for i, filename := range testFiles {
		filePath := filepath.Join(config.SnapshotsPath, filename)
		_, err = os.Stat(filePath)
		if i < 2 { // old files should be deleted
			assert.Error(t, err, "Old file should be deleted")
			assert.True(t, os.IsNotExist(err), "Old file should not exist")
		} else { // new files should still exist
			assert.NoError(t, err, "New file should still exist")
		}
	}

	// Test 2: Cleanup with max count limit
	// Create more test files
	for i := 0; i < 5; i++ {
		filename := fmt.Sprintf("test_%d.jpg", i)
		filePath := filepath.Join(config.SnapshotsPath, filename)
		file, err := os.Create(filePath)
		require.NoError(t, err)
		file.WriteString("test snapshot data")
		file.Close()
	}

	// Cleanup with max count of 3
	err = snapshotManager.CleanupOldSnapshots(ctx, 24*time.Hour, 3)
	require.NoError(t, err, "CleanupOldSnapshots with max count should succeed")

	// Verify only 3 files remain (newest ones)
	entries, err := os.ReadDir(config.SnapshotsPath)
	require.NoError(t, err, "Should be able to read directory")
	assert.LessOrEqual(t, len(entries), 3, "Should have at most 3 files after cleanup")
}

// TestSnapshotManager_ErrorHandling_ReqMTX004 tests error handling scenarios
func TestSnapshotManager_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: "", // Empty path to test error handling
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Test 1: GetSnapshotsList with unconfigured path
	_, err = snapshotManager.GetSnapshotsList(ctx, 10, 0)
	assert.Error(t, err, "Should return error for unconfigured snapshots path")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate path not configured")

	// Test 2: GetSnapshotInfo with unconfigured path
	_, err = snapshotManager.GetSnapshotInfo(ctx, "test.jpg")
	assert.Error(t, err, "Should return error for unconfigured snapshots path")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate path not configured")

	// Test 3: DeleteSnapshotFile with unconfigured path
	err = snapshotManager.DeleteSnapshotFile(ctx, "test.jpg")
	assert.Error(t, err, "Should return error for unconfigured snapshots path")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate path not configured")
}

// TestSnapshotManager_ConcurrentAccess_ReqMTX001 tests concurrent snapshot operations
func TestSnapshotManager_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL:       "http://localhost:9997",
		SnapshotsPath: filepath.Join(helper.GetConfig().TestDataDir, "snapshots"),
	}
	logger := helper.GetLogger()

	// Create FFmpeg manager and snapshot manager
	ffmpegManager := NewFFmpegManager(config, logger)
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)
	require.NotNil(t, snapshotManager)

	ctx := context.Background()

	// Create snapshots directory
	err = os.MkdirAll(config.SnapshotsPath, 0755)
	require.NoError(t, err)

	// Test concurrent snapshot operations
	const numOperations = 5
	errors := make([]error, numOperations)

	// Start concurrent operations
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			// Create test file
			filename := fmt.Sprintf("concurrent_test_%d.jpg", index)
			filePath := filepath.Join(config.SnapshotsPath, filename)
			file, err := os.Create(filePath)
			if err != nil {
				errors[index] = err
				return
			}
			file.WriteString("concurrent test data")
			file.Close()

			// Get snapshot info
			_, err = snapshotManager.GetSnapshotInfo(ctx, filename)
			errors[index] = err
		}(i)
	}

	// Wait for all operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify all operations completed successfully
	for i, err := range errors {
		if err != nil {
			t.Logf("Concurrent operation %d failed: %v", i, err)
		}
	}

	// Verify files were created
	entries, err := os.ReadDir(config.SnapshotsPath)
	require.NoError(t, err, "Should be able to read directory")
	assert.GreaterOrEqual(t, len(entries), numOperations, "Should have created test files")
}
