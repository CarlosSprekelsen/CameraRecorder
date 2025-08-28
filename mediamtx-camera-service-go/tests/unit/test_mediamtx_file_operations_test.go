/*
MediaMTX File Operations Test

Requirements Coverage:
- REQ-FUNC-009: File listing and browsing functionality
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package mediamtx_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

func TestMediaMTXController_ListRecordings(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-file-operations-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recordings")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording files
	testFiles := []struct {
		name     string
		size     int64
		content  string
		expected bool
	}{
		{"camera0_2025-01-15_14-30-00.mp4", 1073741824, "test recording content", true},
		{"camera0_2025-01-15_15-00-00.mp4", 2147483648, "test recording content 2", true},
		{"invalid_file.txt", 1024, "not a recording", false},
		{"camera1_2025-01-15_16-00-00.mp4", 536870912, "test recording content 3", true},
	}

	for _, tf := range testFiles {
		if tf.expected {
			filePath := filepath.Join(tempDir, tf.name)
			err := os.WriteFile(filePath, []byte(tf.content), 0644)
			require.NoError(t, err)

			// Set file modification time
			modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
			err = os.Chtimes(filePath, modTime, modTime)
			require.NoError(t, err)
		}
	}

	// Create recording manager with test directory
	recordingManager := &mediamtx.RecordingManager{
		Config: &mediamtx.MediaMTXConfig{
			RecordingsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		RecordingManager: recordingManager,
		Logger:           logger,
	}

	tests := []struct {
		name           string
		limit          int
		offset         int
		expectedCount  int
		expectedTotal  int
		expectedError  bool
	}{
		{
			name:          "list all recordings",
			limit:         100,
			offset:        0,
			expectedCount: 3,
			expectedTotal: 3,
			expectedError: false,
		},
		{
			name:          "list with limit",
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedTotal: 3,
			expectedError: false,
		},
		{
			name:          "list with offset",
			limit:         100,
			offset:        1,
			expectedCount: 2,
			expectedTotal: 3,
			expectedError: false,
		},
		{
			name:          "list with limit and offset",
			limit:         1,
			offset:        1,
			expectedCount: 1,
			expectedTotal: 3,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := controller.ListRecordings(context.Background(), tt.limit, tt.offset)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Files, tt.expectedCount)
				assert.Equal(t, tt.expectedTotal, response.Total)
				assert.Equal(t, tt.limit, response.Limit)
				assert.Equal(t, tt.offset, response.Offset)

				// Validate file metadata
				for _, file := range response.Files {
					assert.NotEmpty(t, file.FileName)
					assert.True(t, file.FileSize > 0)
					assert.NotZero(t, file.CreatedAt)
					assert.NotZero(t, file.ModifiedAt)
					assert.Contains(t, file.DownloadURL, "/files/recordings/")
					assert.Contains(t, file.FileName, ".mp4")
				}
			}
		})
	}
}

func TestMediaMTXController_ListSnapshots(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-snapshots-operations-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshots")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot files
	testFiles := []struct {
		name     string
		size     int64
		content  string
		expected bool
	}{
		{"snapshot_2025-01-15_14-30-00.jpg", 204800, "test snapshot content", true},
		{"snapshot_2025-01-15_15-00-00.jpg", 245760, "test snapshot content 2", true},
		{"invalid_file.txt", 1024, "not a snapshot", false},
		{"snapshot_2025-01-15_16-00-00.png", 153600, "test snapshot content 3", true},
	}

	for _, tf := range testFiles {
		if tf.expected {
			filePath := filepath.Join(tempDir, tf.name)
			err := os.WriteFile(filePath, []byte(tf.content), 0644)
			require.NoError(t, err)

			// Set file modification time
			modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
			err = os.Chtimes(filePath, modTime, modTime)
			require.NoError(t, err)
		}
	}

	// Create snapshot manager with test directory
	snapshotManager := &mediamtx.SnapshotManager{
		Config: &mediamtx.MediaMTXConfig{
			SnapshotsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		SnapshotManager: snapshotManager,
		Logger:          logger,
	}

	tests := []struct {
		name           string
		limit          int
		offset         int
		expectedCount  int
		expectedTotal  int
		expectedError  bool
	}{
		{
			name:          "list all snapshots",
			limit:         100,
			offset:        0,
			expectedCount: 3,
			expectedTotal: 3,
			expectedError: false,
		},
		{
			name:          "list with limit",
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedTotal: 3,
			expectedError: false,
		},
		{
			name:          "list with offset",
			limit:         100,
			offset:        1,
			expectedCount: 2,
			expectedTotal: 3,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := controller.ListSnapshots(context.Background(), tt.limit, tt.offset)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Files, tt.expectedCount)
				assert.Equal(t, tt.expectedTotal, response.Total)
				assert.Equal(t, tt.limit, response.Limit)
				assert.Equal(t, tt.offset, response.Offset)

				// Validate file metadata
				for _, file := range response.Files {
					assert.NotEmpty(t, file.FileName)
					assert.True(t, file.FileSize > 0)
					assert.NotZero(t, file.CreatedAt)
					assert.NotZero(t, file.ModifiedAt)
					assert.Contains(t, file.DownloadURL, "/files/snapshots/")
					assert.True(t, filepath.Ext(file.FileName) == ".jpg" || filepath.Ext(file.FileName) == ".png")
				}
			}
		})
	}
}

func TestMediaMTXController_GetRecordingInfo(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-recording-info-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_info")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for info test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create recording manager with test directory
	recordingManager := &mediamtx.RecordingManager{
		Config: &mediamtx.MediaMTXConfig{
			RecordingsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		RecordingManager: recordingManager,
		Logger:           logger,
	}

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

func TestMediaMTXController_GetSnapshotInfo(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-snapshot-info-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_info")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for info test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Create snapshot manager with test directory
	snapshotManager := &mediamtx.SnapshotManager{
		Config: &mediamtx.MediaMTXConfig{
			SnapshotsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		SnapshotManager: snapshotManager,
		Logger:          logger,
	}

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

func TestMediaMTXController_DeleteRecording(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-recording-delete-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_recording_delete")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test recording file
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	filePath := filepath.Join(tempDir, fileName)
	content := "test recording content for delete test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create recording manager with test directory
	recordingManager := &mediamtx.RecordingManager{
		Config: &mediamtx.MediaMTXConfig{
			RecordingsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		RecordingManager: recordingManager,
		Logger:           logger,
	}

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

func TestMediaMTXController_DeleteSnapshot(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("mediamtx-snapshot-delete-test")

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "test_snapshot_delete")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test snapshot file
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	filePath := filepath.Join(tempDir, fileName)
	content := "test snapshot content for delete test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Create snapshot manager with test directory
	snapshotManager := &mediamtx.SnapshotManager{
		Config: &mediamtx.MediaMTXConfig{
			SnapshotsPath: tempDir,
		},
		Logger: logger,
	}

	// Create controller
	controller := &mediamtx.Controller{
		SnapshotManager: snapshotManager,
		Logger:          logger,
	}

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

func TestFileMetadata_JSONSerialization(t *testing.T) {
	// Test FileMetadata JSON serialization
	metadata := &mediamtx.FileMetadata{
		FileName:    "test_file.mp4",
		FileSize:    1073741824,
		CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		Duration:    nil,
		DownloadURL: "/files/recordings/test_file.mp4",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metadata)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserialized mediamtx.FileMetadata
	err = json.Unmarshal(jsonData, &deserialized)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, metadata.FileName, deserialized.FileName)
	assert.Equal(t, metadata.FileSize, deserialized.FileSize)
	assert.Equal(t, metadata.CreatedAt, deserialized.CreatedAt)
	assert.Equal(t, metadata.ModifiedAt, deserialized.ModifiedAt)
	assert.Equal(t, metadata.Duration, deserialized.Duration)
	assert.Equal(t, metadata.DownloadURL, deserialized.DownloadURL)
}

func TestFileListResponse_JSONSerialization(t *testing.T) {
	// Test FileListResponse JSON serialization
	response := &mediamtx.FileListResponse{
		Files: []*mediamtx.FileMetadata{
			{
				FileName:    "test_file.mp4",
				FileSize:    1073741824,
				CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				Duration:    nil,
				DownloadURL: "/files/recordings/test_file.mp4",
			},
		},
		Total:  1,
		Limit:  100,
		Offset: 0,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserialized mediamtx.FileListResponse
	err = json.Unmarshal(jsonData, &deserialized)
	assert.NoError(t, err)

	// Verify fields
	assert.Len(t, deserialized.Files, 1)
	assert.Equal(t, response.Total, deserialized.Total)
	assert.Equal(t, response.Limit, deserialized.Limit)
	assert.Equal(t, response.Offset, deserialized.Offset)
}
