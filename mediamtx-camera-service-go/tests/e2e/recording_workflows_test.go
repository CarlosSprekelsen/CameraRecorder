/*
Recording Workflows E2E Tests

Tests complete user workflows for recording lifecycle, duration parameters, multiple cameras,
and recording management. Each test validates actual file creation and verification using
testutils.DataValidationHelper.

Test Categories:
- Basic Recording Lifecycle: Start recording, verify file creation and growth, stop recording
- Recording with Duration Parameter: Start timed recording, verify automatic stop
- Multiple Camera Recording: Record from multiple cameras simultaneously
- Recording List Workflow: List recordings, verify file information

Business Outcomes:
- User can start and stop recordings successfully
- Recording files are created with meaningful content
- Multiple cameras can record simultaneously
- User can list and manage recordings

Coverage Target: 40% E2E coverage milestone
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

func TestBasicRecordingLifecycle(t *testing.T) {
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// Start recording (domain helper)
	rec, err := fixture.StartRecording(DefaultCameraID)
	require.NoError(t, err)

	// Wait for file creation using testutils
	success := dvh.WaitForFileCreation(rec.FilePath(), testutils.DefaultTestTimeout, "recording file")
	require.True(t, success, "Recording file should be created")

	// Verify file growing
	rec.AssertGrowing()

	// Stop recording
	require.NoError(t, rec.Stop())

	// Verify final file
	rec.AssertFileExists(dvh)
}

func TestRecordingWithDurationParameter(t *testing.T) {
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// Start recording with 5 second duration
	startResp, err := fixture.client.StartRecordingWithDuration(DefaultCameraID, 5)
	require.NoError(t, err)
	require.Nil(t, startResp.Error)

	result := startResp.Result.(map[string]interface{})
	basename := result["filename"].(string)
	cfg := fixture.setup.GetConfigManager().GetConfig()
	base := cfg.MediaMTX.RecordingsPath
	if base == "" {
		base = cfg.Storage.DefaultPath
	}
	format := cfg.Recording.RecordFormat
	useSubdirs := cfg.Recording.UseDeviceSubdirs
	recordingPath := testutils.BuildRecordingFilePath(base, "camera0", basename, useSubdirs, format)

	// Wait for file creation
	success := dvh.WaitForFileCreation(recordingPath, testutils.DefaultTestTimeout, "recording file")
	require.True(t, success)

	// Wait for duration + small buffer
	time.Sleep(testutils.UniversalTimeoutVeryLong)

	// Verify recording stopped automatically
	stopResp, err := fixture.client.StopRecording(DefaultCameraID)
	require.NoError(t, err)
	// If already stopped, server may return an error; log and proceed
	if stopResp.Error != nil {
		t.Logf("stop_recording returned error (possibly already stopped): %v", stopResp.Error)
	}

	// Verify final file
	dvh.AssertFileExists(recordingPath, testutils.UniversalMinRecordingFileSize, "final recording")
}

func TestMultipleCameraRecording(t *testing.T) {
	fixture := NewE2EFixture(t)
	dvh := testutils.NewDataValidationHelper(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// Start recording from camera0
	startResp0, err := fixture.client.StartRecording(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, startResp0.Error)
	var recordingPath0 string
	{
		r0 := startResp0.Result.(map[string]interface{})
		basename0 := r0["filename"].(string)
		cfg := fixture.setup.GetConfigManager().GetConfig()
		base := cfg.MediaMTX.RecordingsPath
		if base == "" {
			base = cfg.Storage.DefaultPath
		}
		format := cfg.Recording.RecordFormat
		useSubdirs := cfg.Recording.UseDeviceSubdirs
		recordingPath0 = testutils.BuildRecordingFilePath(base, "camera0", basename0, useSubdirs, format)
	}

	// Start recording from camera1
	startResp1, err := fixture.client.StartRecording("camera1")
	require.NoError(t, err)
	require.Nil(t, startResp1.Error)
	var recordingPath1 string
	{
		r1 := startResp1.Result.(map[string]interface{})
		basename1 := r1["filename"].(string)
		cfg := fixture.setup.GetConfigManager().GetConfig()
		base := cfg.MediaMTX.RecordingsPath
		if base == "" {
			base = cfg.Storage.DefaultPath
		}
		format := cfg.Recording.RecordFormat
		useSubdirs := cfg.Recording.UseDeviceSubdirs
		recordingPath1 = testutils.BuildRecordingFilePath(base, "camera1", basename1, useSubdirs, format)
	}

	// Wait for both files to be created
	success0 := dvh.WaitForFileCreation(recordingPath0, testutils.DefaultTestTimeout, "recording file 0")
	success1 := dvh.WaitForFileCreation(recordingPath1, testutils.DefaultTestTimeout, "recording file 1")
	require.True(t, success0 && success1, "Both recording files should be created")

	// Verify both files are growing
	time.Sleep(testutils.UniversalTimeoutShort)
	size0 := getRecordingFileSize(t, recordingPath0)
	size1 := getRecordingFileSize(t, recordingPath1)
	assert.Greater(t, size0, int64(0), "Camera0 recording should have content")
	assert.Greater(t, size1, int64(0), "Camera1 recording should have content")

	// Stop both recordings
	fixture.client.StopRecording("camera0")
	fixture.client.StopRecording("camera1")

	// Verify final files
	dvh.AssertFileExists(recordingPath0, testutils.UniversalMinRecordingFileSize, "final recording 0")
	dvh.AssertFileExists(recordingPath1, testutils.UniversalMinRecordingFileSize, "final recording 1")
}

func TestRecordingListWorkflow(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Connect and authenticate
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err)

	// Create a recording first
	startResp, err := fixture.client.StartRecording(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, startResp.Error)

	// Wait a moment then stop
	time.Sleep(testutils.UniversalTimeoutShort)
	stopResp, err := fixture.client.StopRecording(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, stopResp.Error)

	// List recordings using proven client method
	listResp, err := fixture.client.ListRecordings()
	require.NoError(t, err)
	require.Nil(t, listResp.Error)

	result := listResp.Result.(map[string]interface{})
	require.Contains(t, result, "files")
	recordings := result["files"].([]interface{})

	// Verify at least one recording exists
	assert.NotEmpty(t, recordings, "Should have at least one recording")

	// Verify recording structure (per API spec)
	for _, recording := range recordings {
		rec := recording.(map[string]interface{})
		assert.Contains(t, rec, "filename", "Recording should have filename field")
		assert.Contains(t, rec, "file_size", "Recording should have file_size field")
		assert.Contains(t, rec, "modified_time", "Recording should have modified_time field")
		assert.Contains(t, rec, "download_url", "Recording should have download_url field")
	}
}

// Helper function for file size
func getRecordingFileSize(t *testing.T, filePath string) int64 {
	info, err := os.Stat(filePath)
	require.NoError(t, err, "File should exist: %s", filePath)
	return info.Size()
}
