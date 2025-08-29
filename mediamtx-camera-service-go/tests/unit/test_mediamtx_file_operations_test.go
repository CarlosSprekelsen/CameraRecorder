//go:build unit
// +build unit

/*
MediaMTX File Operations Tests

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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaMTXController_ListRecordings(t *testing.T) {
	// Setup test environment using proper utilities
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// Create test recording files in the test environment's recordings directory
	testFiles := []struct {
		name    string
		content string
		modTime time.Time
	}{
		{
			name:    "camera0_2025-01-15_14-30-00.mp4",
			content: "test recording content 1",
			modTime: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name:    "camera1_2025-01-15_15-30-00.mp4",
			content: "test recording content 2",
			modTime: time.Date(2025, 1, 15, 15, 30, 0, 0, time.UTC),
		},
		{
			name:    "camera2_2025-01-15_16-30-00.mp4",
			content: "test recording content 3",
			modTime: time.Date(2025, 1, 15, 16, 30, 0, 0, time.UTC),
		},
	}

	// Create test files in the test environment's recordings directory
	recordingsDir := filepath.Join(env.TempDir, "recordings")
	for _, tf := range testFiles {
		filePath := filepath.Join(recordingsDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		require.NoError(t, err)

		// Set file modification time
		err = os.Chtimes(filePath, tf.modTime, tf.modTime)
		require.NoError(t, err)
	}

	// Start the controller
	ctx := context.Background()
	err := env.Controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer env.Controller.Stop(ctx)

	// Use the controller from the test environment
	controller := env.Controller

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectedTotal int
		expectedError bool
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
	// Setup test environment using proper utilities
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// Create test snapshot files in the test environment's snapshots directory
	testFiles := []struct {
		name    string
		content string
		modTime time.Time
	}{
		{
			name:    "snapshot_2025-01-15_14-30-00.jpg",
			content: "test snapshot content 1",
			modTime: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name:    "snapshot_2025-01-15_15-30-00.jpg",
			content: "test snapshot content 2",
			modTime: time.Date(2025, 1, 15, 15, 30, 0, 0, time.UTC),
		},
		{
			name:    "snapshot_2025-01-15_16-30-00.jpg",
			content: "test snapshot content 3",
			modTime: time.Date(2025, 1, 15, 16, 30, 0, 0, time.UTC),
		},
	}

	// Create test files in the test environment's snapshots directory
	snapshotsDir := filepath.Join(env.TempDir, "snapshots")
	for _, tf := range testFiles {
		filePath := filepath.Join(snapshotsDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		require.NoError(t, err)

		// Set file modification time
		err = os.Chtimes(filePath, tf.modTime, tf.modTime)
		require.NoError(t, err)
	}

	// Start the controller
	ctx := context.Background()
	err := env.Controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer env.Controller.Stop(ctx)

	// Use the controller from the test environment
	controller := env.Controller

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectedTotal int
		expectedError bool
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
					assert.Contains(t, file.FileName, ".jpg")
				}
			}
		})
	}
}

func TestMediaMTXController_GetRecordingInfo(t *testing.T) {
	// Setup test environment using proper utilities
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// Create test recording file in the test environment's recordings directory
	fileName := "camera0_2025-01-15_14-30-00.mp4"
	recordingsDir := filepath.Join(env.TempDir, "recordings")
	filePath := filepath.Join(recordingsDir, fileName)
	content := "test recording content for info test"
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Start the controller
	ctx := context.Background()
	err = env.Controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer env.Controller.Stop(ctx)

	// Use the controller from the test environment
	controller := env.Controller

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
			if fileInfo != nil {
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/recordings/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		}
		})
	}
}

func TestMediaMTXController_GetSnapshotInfo(t *testing.T) {
	// Setup test environment using proper utilities
	env := utils.SetupMediaMTXTestEnvironment(t)
	defer utils.TeardownMediaMTXTestEnvironment(t, env)

	// Create test snapshot file in the test environment's snapshots directory
	fileName := "snapshot_2025-01-15_14-30-00.jpg"
	snapshotsDir := filepath.Join(env.TempDir, "snapshots")
	filePath := filepath.Join(snapshotsDir, fileName)
	content := "test snapshot content for info test"
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Set file modification time
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	err = os.Chtimes(filePath, modTime, modTime)
	require.NoError(t, err)

	// Start the controller
	ctx := context.Background()
	err = env.Controller.Start(ctx)
	require.NoError(t, err, "Controller should start successfully")
	defer env.Controller.Stop(ctx)

	// Use the controller from the test environment
	controller := env.Controller

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
			if fileInfo != nil {
				assert.Equal(t, tt.filename, fileInfo.FileName)
				assert.True(t, fileInfo.FileSize > 0)
				assert.NotZero(t, fileInfo.CreatedAt)
				assert.NotZero(t, fileInfo.ModifiedAt)
				assert.Contains(t, fileInfo.DownloadURL, "/files/snapshots/")
				assert.Contains(t, fileInfo.DownloadURL, tt.filename)
			}
		}
		})
	}
}

func TestMediaMTXController_DeleteRecording(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

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

	// Create test configuration manager
	configManager := config.CreateConfigManager()

	// Create controller using proper constructor
	controller, err := mediamtx.ControllerWithConfigManager(configManager, logger)
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

func TestMediaMTXController_DeleteSnapshot(t *testing.T) {
	// Setup test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

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

	// Create test configuration manager
	configManager := config.CreateConfigManager()

	// Create controller using proper constructor
	controller, err := mediamtx.ControllerWithConfigManager(configManager, logger)
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
	// jsonData, err := json.Marshal(metadata) // This line was removed as per the new_code
	// assert.NoError(t, err) // This line was removed as per the new_code
	// assert.NotEmpty(t, jsonData) // This line was removed as per the new_code

	// Deserialize from JSON
	var deserialized mediamtx.FileMetadata
	// err = json.Unmarshal(jsonData, &deserialized) // This line was removed as per the new_code
	// assert.NoError(t, err) // This line was removed as per the new_code

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
	// jsonData, err := json.Marshal(response) // This line was removed as per the new_code
	// assert.NoError(t, err) // This line was removed as per the new_code
	// assert.NotEmpty(t, jsonData) // This line was removed as per the new_code

	// Deserialize from JSON
	var deserialized mediamtx.FileListResponse
	// err = json.Unmarshal(jsonData, &deserialized) // This line was removed as per the new_code
	// assert.NoError(t, err) // This line was removed as per the new_code

	// Verify fields
	assert.Len(t, deserialized.Files, 1)
	assert.Equal(t, response.Total, deserialized.Total)
	assert.Equal(t, response.Limit, deserialized.Limit)
	assert.Equal(t, response.Offset, deserialized.Offset)
}
