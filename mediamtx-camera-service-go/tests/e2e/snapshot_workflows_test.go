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
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	cap, err := fixture.TakeSnapshot(DefaultCameraID)
	require.NoError(t, err)
	cap.AssertImageValid(dvh)
}

func TestMultipleSnapshotsWorkflow(t *testing.T) {
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	var paths []string
	for i := 0; i < 3; i++ {
		cap, err := fixture.TakeSnapshot(DefaultCameraID)
		require.NoError(t, err)
		paths = append(paths, cap.FilePath())
		dvh.WaitForFileCreation(cap.FilePath(), testutils.DefaultTestTimeout, "snapshot file")
		time.Sleep(testutils.UniversalTimeoutShort)
	}
	for _, p := range paths {
		dvh.AssertFileExists(p, testutils.UniversalMinSnapshotFileSize, "snapshot file")
	}
}

func TestSnapshotListWorkflow(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// Create a snapshot first
	cap, err := fixture.TakeSnapshot(DefaultCameraID)
	require.NoError(t, err)
	_ = cap

	// Wait for snapshot to be processed
	time.Sleep(testutils.UniversalTimeoutShort)

	// List snapshots using client
	listResp, err := fixture.client.ListSnapshots()
	require.NoError(t, err)
	require.Nil(t, listResp.Error)

	result := listResp.Result.(map[string]interface{})
	require.Contains(t, result, "files")
	snapshots := result["files"].([]interface{})

	// Verify at least one snapshot exists
	assert.NotEmpty(t, snapshots, "Should have at least one snapshot")

	// Verify snapshot structure (spec-compliant minimal checks)
	for _, snapshot := range snapshots {
		snap := snapshot.(map[string]interface{})
		assert.Contains(t, snap, "filename")
		assert.Contains(t, snap, "file_size")
		assert.Contains(t, snap, "modified_time")
		assert.Contains(t, snap, "download_url")
	}
}

func TestSnapshotWithFormatOptions(t *testing.T) {
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// JPEG
	jpegResp, err := fixture.client.TakeSnapshotWithFormat(DefaultCameraID, "jpeg", 85)
	require.NoError(t, err)
	require.Nil(t, jpegResp.Error)
	jpegPath := fixture.SnapshotPath(DefaultCameraID, jpegResp.Result.(map[string]interface{})["filename"].(string))
	dvh.WaitForFileCreation(jpegPath, testutils.DefaultTestTimeout, "JPEG snapshot")
	dvh.AssertFileExists(jpegPath, testutils.UniversalMinSnapshotFileSize, "JPEG snapshot")

	// PNG
	pngResp, err := fixture.client.TakeSnapshotWithFormat(DefaultCameraID, "png", 0)
	require.NoError(t, err)
	require.Nil(t, pngResp.Error)
	pngPath := fixture.SnapshotPath(DefaultCameraID, pngResp.Result.(map[string]interface{})["filename"].(string))
	dvh.WaitForFileCreation(pngPath, testutils.DefaultTestTimeout, "PNG snapshot")
	dvh.AssertFileExists(pngPath, testutils.UniversalMinSnapshotFileSize, "PNG snapshot")

	// Verify both files exist and are different
	assert.NotEqual(t, jpegPath, pngPath, "JPEG and PNG snapshots should be different files")

	jpegSize := getSnapshotFileSize(t, jpegPath)
	pngSize := getSnapshotFileSize(t, pngPath)
	assert.Greater(t, jpegSize, int64(0))
	assert.Greater(t, pngSize, int64(0))
}

// Helper function for file size
func getSnapshotFileSize(t *testing.T, filePath string) int64 {
	info, err := os.Stat(filePath)
	require.NoError(t, err, "File should exist: %s", filePath)
	return info.Size()
}
