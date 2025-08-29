//go:build unit
// +build unit

/*
MediaMTX File Lifecycle Tests

Requirements Coverage:
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaMTXController_GetRecordingInfoLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "recording_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create controller using proper constructor
	controller, err := mediamtx.NewController(testConfig, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "get existing recording info",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "get non-existent recording info",
			filename:      "non_existent.mp4",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := controller.GetRecordingInfo(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, fileInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fileInfo)
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/recordings/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		})
	}
}

func TestMediaMTXController_GetSnapshotInfoLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create controller using proper constructor
	controller, err := mediamtx.NewController(testConfig, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "get existing snapshot info",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "get non-existent snapshot info",
			filename:      "non_existent.jpg",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := controller.GetSnapshotInfo(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, fileInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fileInfo)
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/snapshots/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		})
	}
}

func TestMediaMTXController_DeleteRecordingLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_delete_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for delete lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create controller using proper constructor
	controller, err := mediamtx.NewController(testConfig, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "delete existing recording",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "delete non-existent recording",
			filename:      "non_existent.mp4",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.DeleteRecording(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify file was deleted
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestMediaMTXController_DeleteSnapshotLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_delete_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for delete lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		SnapshotsPath: tempDir,
	}

	// Create controller using proper constructor
	controller, err := mediamtx.NewController(testConfig, logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "delete existing snapshot",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "delete non-existent snapshot",
			filename:      "non_existent.jpg",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.DeleteSnapshot(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify file was deleted
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestRecordingManager_GetRecordingInfoLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_manager_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for manager lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Create recording manager using proper constructor
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, testConfig, logger)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "get existing recording info",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "get non-existent recording info",
			filename:      "non_existent.mp4",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := recordingManager.GetRecordingInfo(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, fileInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fileInfo)
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/recordings/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		})
	}
}

func TestSnapshotManager_GetSnapshotInfo(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_manager_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for manager lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		SnapshotsPath: tempDir,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Create snapshot manager using proper constructor
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "get existing snapshot info",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "get non-existent snapshot info",
			filename:      "non_existent.jpg",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := snapshotManager.GetSnapshotInfo(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, fileInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fileInfo)
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/snapshots/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		})
	}
}

func TestRecordingManager_DeleteRecordingLifecycle(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_manager_delete_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for manager delete lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: tempDir,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Create recording manager using proper constructor
	recordingManager := mediamtx.NewRecordingManager(ffmpegManager, testConfig, logger)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "delete existing recording",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "delete non-existent recording",
			filename:      "non_existent.mp4",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := recordingManager.DeleteRecording(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify file was deleted
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestSnapshotManager_DeleteSnapshotFile(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_manager_delete_lifecycle")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for manager delete lifecycle test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		SnapshotsPath: tempDir,
	}

	// Create FFmpeg manager
	ffmpegManager := mediamtx.NewFFmpegManager(testConfig, logger)

	// Create snapshot manager using proper constructor
	snapshotManager := mediamtx.NewSnapshotManager(ffmpegManager, testConfig, logger)

	tests := []struct {
		name          string
		filename      string
		expectedError bool
	}{
		{
			name:          "delete existing snapshot",
			filename:      fileName,
			expectedError: false,
		},
		{
			name:          "delete non-existent snapshot",
			filename:      "non_existent.jpg",
			expectedError: true,
		},
		{
			name:          "empty filename",
			filename:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := snapshotManager.DeleteSnapshotFile(context.Background(), tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify file was deleted
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestFileLifecycle_CompleteWorkflow(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_file_lifecycle_workflow")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	recordingFileName := "camera0_2025-01-15_14-30-00.mp4"
	recordingFilePath := filepath.Join(tempDir, "recordings", recordingFileName)
	snapshotFileName := "snapshot_2025-01-15_14-30-00.jpg"
	snapshotFilePath := filepath.Join(tempDir, "snapshots", snapshotFileName)

	// Create directories
	err = os.MkdirAll(filepath.Dir(recordingFilePath), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Dir(snapshotFilePath), 0755)
	require.NoError(t, err)

	// Create test files
	recordingContent := "test recording content for complete workflow"
	err = os.WriteFile(recordingFilePath, []byte(recordingContent), 0644)
	require.NoError(t, err)

	snapshotContent := "test snapshot content for complete workflow"
	err = os.WriteFile(snapshotFilePath, []byte(snapshotContent), 0644)
	require.NoError(t, err)

	// Set file modification times
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(recordingFilePath, modTime, modTime)
	require.NoError(t, err)
	err = os.Chtimes(snapshotFilePath, modTime, modTime)
	require.NoError(t, err)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		RecordingsPath: filepath.Join(tempDir, "recordings"),
		SnapshotsPath:  filepath.Join(tempDir, "snapshots"),
	}

	// Create controller using proper constructor
	controller, err := mediamtx.NewController(testConfig, logger)
	require.NoError(t, err)

	t.Run("complete file lifecycle workflow", func(t *testing.T) {
		// Step 1: Get recording info
		recordingInfo, err := controller.GetRecordingInfo(context.Background(), recordingFileName)
		assert.NoError(t, err)
		assert.NotNil(t, recordingInfo)
		assert.Equal(t, recordingFileName, recordingInfo.FileName)
		assert.True(t, recordingInfo.FileSize > 0)

		// Step 2: Get snapshot info
		snapshotInfo, err := controller.GetSnapshotInfo(context.Background(), snapshotFileName)
		assert.NoError(t, err)
		assert.NotNil(t, snapshotInfo)
		assert.Equal(t, snapshotFileName, snapshotInfo.FileName)
		assert.True(t, snapshotInfo.FileSize > 0)

		// Step 3: Delete recording
		err = controller.DeleteRecording(context.Background(), recordingFileName)
		assert.NoError(t, err)

		// Verify recording was deleted
		_, err = os.Stat(recordingFilePath)
		assert.True(t, os.IsNotExist(err))

		// Step 4: Delete snapshot
		err = controller.DeleteSnapshot(context.Background(), snapshotFileName)
		assert.NoError(t, err)

		// Verify snapshot was deleted
		_, err = os.Stat(snapshotFilePath)
		assert.True(t, os.IsNotExist(err))

		// Step 5: Verify files are gone
		_, err = controller.GetRecordingInfo(context.Background(), recordingFileName)
		assert.Error(t, err)

		_, err = controller.GetSnapshotInfo(context.Background(), snapshotFileName)
		assert.Error(t, err)
	})
}
