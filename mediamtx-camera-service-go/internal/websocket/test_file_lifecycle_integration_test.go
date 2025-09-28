/*
File Lifecycle Integration Tests

Tests complete file operations using testutils DataValidationHelper
for comprehensive validation of create→list→delete workflows.

Uses configuration-driven paths and leverages existing testutils
for sophisticated file validation.
*/

package websocket

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestWebSocket_FileLifecycle_Complete_Integration tests complete file lifecycle
func TestWebSocket_FileLifecycle_Complete_Integration(t *testing.T) {
	asserter := GetSharedWebSocketAsserter(t)

	// Test complete file lifecycle using testutils validation
	err := asserter.AssertFileLifecycleWorkflow()
	require.NoError(t, err, "File lifecycle workflow should succeed")
}

// TestWebSocket_FileValidation_Comprehensive_Integration tests file validation using testutils
func TestWebSocket_FileValidation_Comprehensive_Integration(t *testing.T) {
	asserter := GetSharedWebSocketAsserter(t)

	// Test snapshot validation using testutils
	err := asserter.AssertSnapshotWorkflow()
	require.NoError(t, err, "Snapshot workflow should succeed")

	// Test recording validation using testutils
	err = asserter.AssertRecordingWorkflow()
	require.NoError(t, err, "Recording workflow should succeed")
}

// TestWebSocket_ConfigurationDrivenPaths_Integration tests configuration-driven path handling
func TestWebSocket_ConfigurationDrivenPaths_Integration(t *testing.T) {
	// Use testutils for comprehensive setup
	asserter := GetSharedWebSocketAsserter(t)
	defer asserter.helper.Cleanup()

	// Test that paths are loaded from testutils, not hardcoded
	err := asserter.AssertSnapshotWorkflow()
	require.NoError(t, err, "Configuration-driven snapshot workflow should succeed")

	// Verify paths are configuration-driven using testutils
	snapshotsPath := testutils.GetTestSnapshotsPath()
	require.Equal(t, "/tmp/snapshots", snapshotsPath, "Snapshots path should be loaded from testutils")

	recordingsPath := testutils.GetTestRecordingsPath()
	require.Equal(t, "/tmp/recordings", recordingsPath, "Recordings path should be loaded from testutils")
}

// TestWebSocket_TestUtilsIntegration_Integration tests integration with testutils
func TestWebSocket_TestUtilsIntegration_Integration(t *testing.T) {
	// Test that testutils DataValidationHelper works correctly
	dvh := testutils.NewDataValidationHelper(t)

	// Test file path building using testutils constants
	snapshotsPath := testutils.GetTestSnapshotsPath()
	cameraID := testutils.GetTestCameraID()
	filename := "test_file.jpg"

	// Test testutils path building
	expectedPath := testutils.BuildSnapshotFilePath(snapshotsPath, cameraID, filename, true, "jpg")
	require.Contains(t, expectedPath, snapshotsPath, "Path should contain snapshots directory")
	require.Contains(t, expectedPath, cameraID, "Path should contain camera ID")
	require.Contains(t, expectedPath, filename, "Path should contain filename")

	// Test testutils file validation
	dvh.AssertFileNotExists(expectedPath, "File should not exist initially")
}
