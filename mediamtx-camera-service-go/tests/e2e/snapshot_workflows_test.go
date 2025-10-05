/*
Snapshot Workflows E2E Tests

Tests complete user workflows for snapshot capture, multiple snapshots, listing,
and format options. Each test validates actual image file creation and verification
using testutils.DataValidationHelper.

Test Categories:
- Single Snapshot Capture: Take snapshot, verify image file creation
- Multiple Snapshots Workflow: Take multiple snapshots, verify all files created
- Snapshot List Workflow: List snapshots, verify file information
- Snapshot with Format Options: Test different image formats and quality settings

Business Outcomes:
- User can capture snapshots successfully
- Snapshot files are created with meaningful content
- Multiple snapshots can be taken
- User can list and manage snapshots

Coverage Target: 55% E2E coverage milestone
*/

package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleSnapshotCapture(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	dvh := testutils.NewDataValidationHelper(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err)
	
	// Take snapshot using proven client method
	snapshotResp, err := asserter.TakeSnapshot("camera0")
	require.NoError(t, err)
	require.Nil(t, snapshotResp.Error)
	
	result := snapshotResp.Result.(map[string]interface{})
	snapshotPath := result["filename"].(string)
	
	// Wait for file creation using testutils
	success := dvh.WaitForFileCreation(snapshotPath, testutils.DefaultTestTimeout, "snapshot file")
	require.True(t, success, "Snapshot file should be created")
	
	// Verify file has meaningful content
	dvh.AssertFileExists(snapshotPath, testutils.UniversalMinSnapshotFileSize, "snapshot file")
	
	// Verify file is a valid image (has proper extension)
	assert.Contains(t, snapshotPath, ".jpg", "Snapshot should be JPEG format")
}

func TestMultipleSnapshotsWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	dvh := testutils.NewDataValidationHelper(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err)
	
	var snapshotPaths []string
	
	// Take 3 snapshots using proven client method
	for i := 0; i < 3; i++ {
		snapshotResp, err := asserter.TakeSnapshot("camera0")
		require.NoError(t, err)
		require.Nil(t, snapshotResp.Error)
		
		result := snapshotResp.Result.(map[string]interface{})
		snapshotPath := result["filename"].(string)
		snapshotPaths = append(snapshotPaths, snapshotPath)
		
		// Wait for file creation
		success := dvh.WaitForFileCreation(snapshotPath, testutils.DefaultTestTimeout, "snapshot file")
		require.True(t, success, "Snapshot file %d should be created", i+1)
		
		// Small delay between snapshots
		time.Sleep(500 * time.Millisecond)
	}
	
	// Verify all files exist and have content
	for _, snapshotPath := range snapshotPaths {
		dvh.AssertFileExists(snapshotPath, testutils.UniversalMinSnapshotFileSize, "snapshot file")
		assert.Contains(t, snapshotPath, ".jpg", "Snapshot should be JPEG format")
	}
}

func TestSnapshotListWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err)
	
	// Create a snapshot first
	snapshotResp, err := asserter.TakeSnapshot("camera0")
	require.NoError(t, err)
	require.Nil(t, snapshotResp.Error)
	
	// Wait for snapshot to be processed
	time.Sleep(1 * time.Second)
	
	// List snapshots using proven client method
	listResp, err := asserter.ListSnapshots()
	require.NoError(t, err)
	require.Nil(t, listResp.Error)
	
	result := listResp.Result.(map[string]interface{})
	require.Contains(t, result, "files")
	snapshots := result["files"].([]interface{})
	
	// Verify at least one snapshot exists
	assert.NotEmpty(t, snapshots, "Should have at least one snapshot")
	
	// Verify snapshot structure
	for _, snapshot := range snapshots {
		snap := snapshot.(map[string]interface{})
		assert.Contains(t, snap, "device", "Snapshot should have device field")
		assert.Contains(t, snap, "file_path", "Snapshot should have file_path field")
		assert.Contains(t, snap, "timestamp", "Snapshot should have timestamp field")
		assert.Contains(t, snap, "file_size", "Snapshot should have file_size field")
		assert.Contains(t, snap, "format", "Snapshot should have format field")
	}
}

func TestSnapshotWithFormatOptions(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)
	dvh := testutils.NewDataValidationHelper(t)
	
	// Connect and authenticate
	err := asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err)
	
	// Test JPEG format with quality setting using proven client method
	jpegResp, err := asserter.TakeSnapshotWithFormat("camera0", "jpeg", 85)
	require.NoError(t, err)
	require.Nil(t, jpegResp.Error)
	
	jpegResult := jpegResp.Result.(map[string]interface{})
	jpegPath := jpegResult["filename"].(string)
	
	// Wait for file creation
	success := dvh.WaitForFileCreation(jpegPath, testutils.DefaultTestTimeout, "JPEG snapshot")
	require.True(t, success)
	
	// Verify JPEG file
	dvh.AssertFileExists(jpegPath, testutils.UniversalMinSnapshotFileSize, "JPEG snapshot")
	assert.Contains(t, jpegPath, ".jpg", "Should be JPEG format")
	
	// Test PNG format using proven client method
	pngResp, err := asserter.TakeSnapshotWithFormat("camera0", "png", 0)
	require.NoError(t, err)
	require.Nil(t, pngResp.Error)
	
	pngResult := pngResp.Result.(map[string]interface{})
	pngPath := pngResult["filename"].(string)
	
	// Wait for file creation
	success = dvh.WaitForFileCreation(pngPath, testutils.DefaultTestTimeout, "PNG snapshot")
	require.True(t, success)
	
	// Verify PNG file
	dvh.AssertFileExists(pngPath, testutils.UniversalMinSnapshotFileSize, "PNG snapshot")
	assert.Contains(t, pngPath, ".png", "Should be PNG format")
	
	// Verify both files exist and are different
	assert.NotEqual(t, jpegPath, pngPath, "JPEG and PNG snapshots should be different files")
	
	jpegSize := getSnapshotFileSize(t, jpegPath)
	pngSize := getSnapshotFileSize(t, pngPath)
	assert.Greater(t, jpegSize, int64(0), "JPEG file should have content")
	assert.Greater(t, pngSize, int64(0), "PNG file should have content")
}

// Helper function for file size
func getSnapshotFileSize(t *testing.T, filePath string) int64 {
	info, err := os.Stat(filePath)
	require.NoError(t, err, "File should exist: %s", filePath)
	return info.Size()
}