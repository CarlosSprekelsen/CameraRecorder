/*
WebSocket File Lifecycle Methods Test

Requirements Coverage:
- REQ-FUNC-009: File listing and browsing functionality
- REQ-API-001: JSON-RPC method implementation
- REQ-SEC-002: Role-based access control

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mediamtx-camera-service-go/internal/mediamtx"
	"mediamtx-camera-service-go/internal/websocket"
	"mediamtx-camera-service-go/internal/logging"
)

// Mock MediaMTX controller for testing
type mockMediaMTXControllerLifecycle struct {
	recordingInfoResponse *mediamtx.FileMetadata
	snapshotInfoResponse  *mediamtx.FileMetadata
	recordingInfoError    error
	snapshotInfoError     error
	deleteRecordingError  error
	deleteSnapshotError   error
}

func (m *mockMediaMTXControllerLifecycle) GetRecordingInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) {
	return m.recordingInfoResponse, m.recordingInfoError
}

func (m *mockMediaMTXControllerLifecycle) GetSnapshotInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) {
	return m.snapshotInfoResponse, m.snapshotInfoError
}

func (m *mockMediaMTXControllerLifecycle) DeleteRecording(ctx context.Context, filename string) error {
	return m.deleteRecordingError
}

func (m *mockMediaMTXControllerLifecycle) DeleteSnapshot(ctx context.Context, filename string) error {
	return m.deleteSnapshotError
}

// Implement other required methods with empty implementations
func (m *mockMediaMTXControllerLifecycle) GetHealth(ctx context.Context) (*mediamtx.HealthStatus, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetMetrics(ctx context.Context) (*mediamtx.Metrics, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetStreams(ctx context.Context) ([]*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetStream(ctx context.Context, id string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) CreateStream(ctx context.Context, name, source string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) DeleteStream(ctx context.Context, id string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) GetPaths(ctx context.Context) ([]*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetPath(ctx context.Context, name string) (*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) CreatePath(ctx context.Context, path *mediamtx.Path) error { return nil }
func (m *mockMediaMTXControllerLifecycle) DeletePath(ctx context.Context, name string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) StartRecording(ctx context.Context, device, path string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) StopRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) TakeSnapshot(ctx context.Context, device, path string) (*mediamtx.Snapshot, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetRecordingStatus(ctx context.Context, sessionID string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) ListRecordings(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) ListSnapshots(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) StopAdvancedRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) GetAdvancedRecordingSession(sessionID string) (*mediamtx.RecordingSession, bool) { return nil, false }
func (m *mockMediaMTXControllerLifecycle) ListAdvancedRecordingSessions() []*mediamtx.RecordingSession { return nil }
func (m *mockMediaMTXControllerLifecycle) RotateRecordingFile(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) TakeAdvancedSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.Snapshot, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) GetAdvancedSnapshot(snapshotID string) (*mediamtx.Snapshot, bool) { return nil, false }
func (m *mockMediaMTXControllerLifecycle) ListAdvancedSnapshots() []*mediamtx.Snapshot { return nil }
func (m *mockMediaMTXControllerLifecycle) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error { return nil }
func (m *mockMediaMTXControllerLifecycle) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error { return nil }
func (m *mockMediaMTXControllerLifecycle) GetSnapshotSettings() *mediamtx.SnapshotSettings { return nil }
func (m *mockMediaMTXControllerLifecycle) UpdateSnapshotSettings(settings *mediamtx.SnapshotSettings) {}
func (m *mockMediaMTXControllerLifecycle) GetConfig(ctx context.Context) (*mediamtx.MediaMTXConfig, error) { return nil, nil }
func (m *mockMediaMTXControllerLifecycle) UpdateConfig(ctx context.Context, config *mediamtx.MediaMTXConfig) error { return nil }
func (m *mockMediaMTXControllerLifecycle) Start(ctx context.Context) error { return nil }
func (m *mockMediaMTXControllerLifecycle) Stop(ctx context.Context) error { return nil }

func TestWebSocketServer_MethodGetRecordingInfo(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-recording-info-test")

	tests := []struct {
		name           string
		params         map[string]interface{}
		mockResponse   *mediamtx.FileMetadata
		mockError      error
		expectedResult map[string]interface{}
		expectedError  bool
	}{
		{
			name: "successful recording info retrieval",
			params: map[string]interface{}{
				"filename": "camera0_2025-01-15_14-30-00.mp4",
			},
			mockResponse: &mediamtx.FileMetadata{
				FileName:    "camera0_2025-01-15_14-30-00.mp4",
				FileSize:    1073741824,
				CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				Duration:    nil,
				DownloadURL: "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"filename":     "camera0_2025-01-15_14-30-00.mp4",
				"file_size":    int64(1073741824),
				"created_at":   "2025-01-15T14:30:00Z",
				"modified_at":  "2025-01-15T14:30:00Z",
				"download_url": "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
			},
			expectedError: false,
		},
		{
			name: "recording info with duration",
			params: map[string]interface{}{
				"filename": "camera0_2025-01-15_14-30-00.mp4",
			},
			mockResponse: &mediamtx.FileMetadata{
				FileName:    "camera0_2025-01-15_14-30-00.mp4",
				FileSize:    1073741824,
				CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				Duration:    func() *int64 { d := int64(3600); return &d }(),
				DownloadURL: "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"filename":     "camera0_2025-01-15_14-30-00.mp4",
				"file_size":    int64(1073741824),
				"created_at":   "2025-01-15T14:30:00Z",
				"modified_at":  "2025-01-15T14:30:00Z",
				"duration":     int64(3600),
				"download_url": "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
			},
			expectedError: false,
		},
		{
			name: "controller error",
			params: map[string]interface{}{
				"filename": "non_existent.mp4",
			},
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXControllerLifecycle{
				recordingInfoResponse: tt.mockResponse,
				recordingInfoError:    tt.mockError,
			}

			// Create WebSocket server
			server := &websocket.WebSocketServer{
				MediaMTXController: mockController,
				Logger:             logger,
			}

			// Create mock client connection
			client := &websocket.ClientConnection{
				User: "test_user",
				Role: "viewer",
			}

			// Call method
			response, err := server.MethodGetRecordingInfo(tt.params, client)

			// Validate results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error)

				// Validate result structure
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result should be a map")

				// Validate required fields
				assert.Equal(t, tt.expectedResult["filename"], result["filename"])
				assert.Equal(t, tt.expectedResult["file_size"], result["file_size"])
				assert.Equal(t, tt.expectedResult["created_at"], result["created_at"])
				assert.Equal(t, tt.expectedResult["modified_at"], result["modified_at"])
				assert.Equal(t, tt.expectedResult["download_url"], result["download_url"])

				// Validate optional duration field
				if duration, exists := tt.expectedResult["duration"]; exists {
					assert.Equal(t, duration, result["duration"])
				} else {
					assert.NotContains(t, result, "duration")
				}
			}
		})
	}
}

func TestWebSocketServer_MethodGetSnapshotInfo(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-snapshot-info-test")

	tests := []struct {
		name           string
		params         map[string]interface{}
		mockResponse   *mediamtx.FileMetadata
		mockError      error
		expectedResult map[string]interface{}
		expectedError  bool
	}{
		{
			name: "successful snapshot info retrieval",
			params: map[string]interface{}{
				"filename": "snapshot_2025-01-15_14-30-00.jpg",
			},
			mockResponse: &mediamtx.FileMetadata{
				FileName:    "snapshot_2025-01-15_14-30-00.jpg",
				FileSize:    204800,
				CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				Duration:    nil,
				DownloadURL: "/files/snapshots/snapshot_2025-01-15_14-30-00.jpg",
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"filename":     "snapshot_2025-01-15_14-30-00.jpg",
				"file_size":    int64(204800),
				"created_at":   "2025-01-15T14:30:00Z",
				"modified_at":  "2025-01-15T14:30:00Z",
				"download_url": "/files/snapshots/snapshot_2025-01-15_14-30-00.jpg",
			},
			expectedError: false,
		},
		{
			name: "controller error",
			params: map[string]interface{}{
				"filename": "non_existent.jpg",
			},
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXControllerLifecycle{
				snapshotInfoResponse: tt.mockResponse,
				snapshotInfoError:    tt.mockError,
			}

			// Create WebSocket server
			server := &websocket.WebSocketServer{
				MediaMTXController: mockController,
				Logger:             logger,
			}

			// Create mock client connection
			client := &websocket.ClientConnection{
				User: "test_user",
				Role: "viewer",
			}

			// Call method
			response, err := server.MethodGetSnapshotInfo(tt.params, client)

			// Validate results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error)

				// Validate result structure
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result should be a map")

				// Validate required fields
				assert.Equal(t, tt.expectedResult["filename"], result["filename"])
				assert.Equal(t, tt.expectedResult["file_size"], result["file_size"])
				assert.Equal(t, tt.expectedResult["created_at"], result["created_at"])
				assert.Equal(t, tt.expectedResult["modified_at"], result["modified_at"])
				assert.Equal(t, tt.expectedResult["download_url"], result["download_url"])
			}
		})
	}
}

func TestWebSocketServer_MethodDeleteRecording(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-recording-delete-test")

	tests := []struct {
		name          string
		params        map[string]interface{}
		mockError     error
		expectedError bool
	}{
		{
			name: "successful recording deletion",
			params: map[string]interface{}{
				"filename": "camera0_2025-01-15_14-30-00.mp4",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "controller error",
			params: map[string]interface{}{
				"filename": "non_existent.mp4",
			},
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXControllerLifecycle{
				deleteRecordingError: tt.mockError,
			}

			// Create WebSocket server
			server := &websocket.WebSocketServer{
				MediaMTXController: mockController,
				Logger:             logger,
			}

			// Create mock client connection
			client := &websocket.ClientConnection{
				User: "test_user",
				Role: "operator",
			}

			// Call method
			response, err := server.MethodDeleteRecording(tt.params, client)

			// Validate results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error)

				// Validate result structure
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result should be a map")
				assert.Equal(t, "success", result["status"])
			}
		})
	}
}

func TestWebSocketServer_MethodDeleteSnapshot(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-snapshot-delete-test")

	tests := []struct {
		name          string
		params        map[string]interface{}
		mockError     error
		expectedError bool
	}{
		{
			name: "successful snapshot deletion",
			params: map[string]interface{}{
				"filename": "snapshot_2025-01-15_14-30-00.jpg",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "controller error",
			params: map[string]interface{}{
				"filename": "non_existent.jpg",
			},
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXControllerLifecycle{
				deleteSnapshotError: tt.mockError,
			}

			// Create WebSocket server
			server := &websocket.WebSocketServer{
				MediaMTXController: mockController,
				Logger:             logger,
			}

			// Create mock client connection
			client := &websocket.ClientConnection{
				User: "test_user",
				Role: "operator",
			}

			// Call method
			response, err := server.MethodDeleteSnapshot(tt.params, client)

			// Validate results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error)

				// Validate result structure
				result, ok := response.Result.(map[string]interface{})
				require.True(t, ok, "Result should be a map")
				assert.Equal(t, "success", result["status"])
			}
		})
	}
}

func TestWebSocketServer_FileLifecycleParameterValidation(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-lifecycle-parameter-validation-test")

	// Create mock controller
	mockController := &mockMediaMTXControllerLifecycle{}

	// Create WebSocket server
	server := &websocket.WebSocketServer{
		MediaMTXController: mockController,
		Logger:             logger,
	}

	// Create mock client connection
	client := &websocket.ClientConnection{
		User: "test_user",
		Role: "viewer",
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		method        string
		expectedError bool
	}{
		{
			name: "missing filename parameter",
			params: map[string]interface{}{},
			method:        "get_recording_info",
			expectedError: true,
		},
		{
			name: "empty filename parameter",
			params: map[string]interface{}{
				"filename": "",
			},
			method:        "get_recording_info",
			expectedError: true,
		},
		{
			name: "invalid filename type",
			params: map[string]interface{}{
				"filename": 123,
			},
			method:        "get_recording_info",
			expectedError: true,
		},
		{
			name: "missing filename for deletion",
			params: map[string]interface{}{},
			method:        "delete_recording",
			expectedError: true,
		},
		{
			name: "empty filename for deletion",
			params: map[string]interface{}{
				"filename": "",
			},
			method:        "delete_recording",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response *websocket.JsonRpcResponse
			var err error

			switch tt.method {
			case "get_recording_info":
				response, err = server.MethodGetRecordingInfo(tt.params, client)
			case "get_snapshot_info":
				response, err = server.MethodGetSnapshotInfo(tt.params, client)
			case "delete_recording":
				response, err = server.MethodDeleteRecording(tt.params, client)
			case "delete_snapshot":
				response, err = server.MethodDeleteSnapshot(tt.params, client)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func TestWebSocketServer_FileLifecycleAuthentication(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-lifecycle-authentication-test")

	// Create mock controller
	mockController := &mockMediaMTXControllerLifecycle{
		recordingInfoResponse: &mediamtx.FileMetadata{
			FileName:    "test.mp4",
			FileSize:    1024,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			DownloadURL: "/files/recordings/test.mp4",
		},
	}

	// Create WebSocket server
	server := &websocket.WebSocketServer{
		MediaMTXController: mockController,
		Logger:             logger,
	}

	tests := []struct {
		name          string
		client        *websocket.ClientConnection
		method        string
		expectedError bool
	}{
		{
			name: "unauthenticated user for info",
			client: &websocket.ClientConnection{
				User: "",
				Role: "",
			},
			method:        "get_recording_info",
			expectedError: true,
		},
		{
			name: "viewer role for info - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "viewer",
			},
			method:        "get_recording_info",
			expectedError: false,
		},
		{
			name: "operator role for info - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "operator",
			},
			method:        "get_recording_info",
			expectedError: false,
		},
		{
			name: "viewer role for deletion - should fail",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "viewer",
			},
			method:        "delete_recording",
			expectedError: true,
		},
		{
			name: "operator role for deletion - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "operator",
			},
			method:        "delete_recording",
			expectedError: false,
		},
		{
			name: "admin role for deletion - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "admin",
			},
			method:        "delete_recording",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"filename": "test.mp4",
			}

			var response *websocket.JsonRpcResponse
			var err error

			switch tt.method {
			case "get_recording_info":
				response, err = server.MethodGetRecordingInfo(params, tt.client)
			case "get_snapshot_info":
				response, err = server.MethodGetSnapshotInfo(params, tt.client)
			case "delete_recording":
				response, err = server.MethodDeleteRecording(params, tt.client)
			case "delete_snapshot":
				response, err = server.MethodDeleteSnapshot(params, tt.client)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}
