/*
WebSocket File Listing Methods Test

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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mediamtx-camera-service-go/internal/mediamtx"
	"mediamtx-camera-service-go/internal/websocket"
	"mediamtx-camera-service-go/internal/logging"
)

// Mock MediaMTX controller for testing
type mockMediaMTXController struct {
	recordingsResponse *mediamtx.FileListResponse
	snapshotsResponse  *mediamtx.FileListResponse
	recordingsError    error
	snapshotsError     error
}

func (m *mockMediaMTXController) ListRecordings(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) {
	return m.recordingsResponse, m.recordingsError
}

func (m *mockMediaMTXController) ListSnapshots(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) {
	return m.snapshotsResponse, m.snapshotsError
}

// Implement other required methods with empty implementations
func (m *mockMediaMTXController) GetHealth(ctx context.Context) (*mediamtx.HealthStatus, error) { return nil, nil }
func (m *mockMediaMTXController) GetMetrics(ctx context.Context) (*mediamtx.Metrics, error) { return nil, nil }
func (m *mockMediaMTXController) GetStreams(ctx context.Context) ([]*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXController) GetStream(ctx context.Context, id string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXController) CreateStream(ctx context.Context, name, source string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXController) DeleteStream(ctx context.Context, id string) error { return nil }
func (m *mockMediaMTXController) GetPaths(ctx context.Context) ([]*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXController) GetPath(ctx context.Context, name string) (*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXController) CreatePath(ctx context.Context, path *mediamtx.Path) error { return nil }
func (m *mockMediaMTXController) DeletePath(ctx context.Context, name string) error { return nil }
func (m *mockMediaMTXController) StartRecording(ctx context.Context, device, path string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXController) StopRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXController) TakeSnapshot(ctx context.Context, device, path string) (*mediamtx.Snapshot, error) { return nil, nil }
func (m *mockMediaMTXController) GetRecordingStatus(ctx context.Context, sessionID string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXController) GetRecordingInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) { return nil, nil }
func (m *mockMediaMTXController) GetSnapshotInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) { return nil, nil }
func (m *mockMediaMTXController) DeleteRecording(ctx context.Context, filename string) error { return nil }
func (m *mockMediaMTXController) DeleteSnapshot(ctx context.Context, filename string) error { return nil }
func (m *mockMediaMTXController) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXController) StopAdvancedRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXController) GetAdvancedRecordingSession(sessionID string) (*mediamtx.RecordingSession, bool) { return nil, false }
func (m *mockMediaMTXController) ListAdvancedRecordingSessions() []*mediamtx.RecordingSession { return nil }
func (m *mockMediaMTXController) RotateRecordingFile(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXController) TakeAdvancedSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.Snapshot, error) { return nil, nil }
func (m *mockMediaMTXController) GetAdvancedSnapshot(snapshotID string) (*mediamtx.Snapshot, bool) { return nil, false }
func (m *mockMediaMTXController) ListAdvancedSnapshots() []*mediamtx.Snapshot { return nil }
func (m *mockMediaMTXController) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error { return nil }
func (m *mockMediaMTXController) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error { return nil }
func (m *mockMediaMTXController) GetSnapshotSettings() *mediamtx.SnapshotSettings { return nil }
func (m *mockMediaMTXController) UpdateSnapshotSettings(settings *mediamtx.SnapshotSettings) {}
func (m *mockMediaMTXController) GetConfig(ctx context.Context) (*mediamtx.MediaMTXConfig, error) { return nil, nil }
func (m *mockMediaMTXController) UpdateConfig(ctx context.Context, config *mediamtx.MediaMTXConfig) error { return nil }
func (m *mockMediaMTXController) Start(ctx context.Context) error { return nil }
func (m *mockMediaMTXController) Stop(ctx context.Context) error { return nil }

func TestWebSocketServer_MethodListRecordings(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-file-listing-test")

	tests := []struct {
		name           string
		params         map[string]interface{}
		mockResponse   *mediamtx.FileListResponse
		mockError      error
		expectedResult map[string]interface{}
		expectedError  bool
	}{
		{
			name: "successful recordings list with default parameters",
			params: map[string]interface{}{},
			mockResponse: &mediamtx.FileListResponse{
				Files: []*mediamtx.FileMetadata{
					{
						FileName:    "camera0_2025-01-15_14-30-00.mp4",
						FileSize:    1073741824,
						CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
						ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
						Duration:    nil,
						DownloadURL: "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
					},
				},
				Total:  1,
				Limit:  100,
				Offset: 0,
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"files": []map[string]interface{}{
					{
						"filename":     "camera0_2025-01-15_14-30-00.mp4",
						"file_size":    int64(1073741824),
						"created_at":   "2025-01-15T14:30:00Z",
						"modified_at":  "2025-01-15T14:30:00Z",
						"download_url": "/files/recordings/camera0_2025-01-15_14-30-00.mp4",
					},
				},
				"total":  1,
				"limit":  100,
				"offset": 0,
			},
			expectedError: false,
		},
		{
			name: "successful recordings list with custom parameters",
			params: map[string]interface{}{
				"limit":  10,
				"offset": 5,
			},
			mockResponse: &mediamtx.FileListResponse{
				Files:  []*mediamtx.FileMetadata{},
				Total:  0,
				Limit:  10,
				Offset: 5,
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"files":  []map[string]interface{}{},
				"total":  0,
				"limit":  10,
				"offset": 5,
			},
			expectedError: false,
		},
		{
			name:           "controller error",
			params:         map[string]interface{}{},
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXController{
				recordingsResponse: tt.mockResponse,
				recordingsError:    tt.mockError,
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
			response, err := server.MethodListRecordings(tt.params, client)

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

				// Validate files array
				files, ok := result["files"].([]map[string]interface{})
				require.True(t, ok, "Files should be an array")

				expectedFiles := tt.expectedResult["files"].([]map[string]interface{})
				assert.Len(t, files, len(expectedFiles))

				// Validate pagination fields
				assert.Equal(t, tt.expectedResult["total"], result["total"])
				assert.Equal(t, tt.expectedResult["limit"], result["limit"])
				assert.Equal(t, tt.expectedResult["offset"], result["offset"])
			}
		})
	}
}

func TestWebSocketServer_MethodListSnapshots(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-snapshots-listing-test")

	tests := []struct {
		name           string
		params         map[string]interface{}
		mockResponse   *mediamtx.FileListResponse
		mockError      error
		expectedResult map[string]interface{}
		expectedError  bool
	}{
		{
			name: "successful snapshots list with default parameters",
			params: map[string]interface{}{},
			mockResponse: &mediamtx.FileListResponse{
				Files: []*mediamtx.FileMetadata{
					{
						FileName:    "snapshot_2025-01-15_14-30-00.jpg",
						FileSize:    204800,
						CreatedAt:   time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
						ModifiedAt:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
						Duration:    nil,
						DownloadURL: "/files/snapshots/snapshot_2025-01-15_14-30-00.jpg",
					},
				},
				Total:  1,
				Limit:  100,
				Offset: 0,
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"files": []map[string]interface{}{
					{
						"filename":     "snapshot_2025-01-15_14-30-00.jpg",
						"file_size":    int64(204800),
						"created_at":   "2025-01-15T14:30:00Z",
						"modified_at":  "2025-01-15T14:30:00Z",
						"download_url": "/files/snapshots/snapshot_2025-01-15_14-30-00.jpg",
					},
				},
				"total":  1,
				"limit":  100,
				"offset": 0,
			},
			expectedError: false,
		},
		{
			name: "successful snapshots list with custom parameters",
			params: map[string]interface{}{
				"limit":  5,
				"offset": 10,
			},
			mockResponse: &mediamtx.FileListResponse{
				Files:  []*mediamtx.FileMetadata{},
				Total:  0,
				Limit:  5,
				Offset: 10,
			},
			mockError: nil,
			expectedResult: map[string]interface{}{
				"files":  []map[string]interface{}{},
				"total":  0,
				"limit":  5,
				"offset": 10,
			},
			expectedError: false,
		},
		{
			name:           "controller error",
			params:         map[string]interface{}{},
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := &mockMediaMTXController{
				snapshotsResponse: tt.mockResponse,
				snapshotsError:    tt.mockError,
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
			response, err := server.MethodListSnapshots(tt.params, client)

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

				// Validate files array
				files, ok := result["files"].([]map[string]interface{})
				require.True(t, ok, "Files should be an array")

				expectedFiles := tt.expectedResult["files"].([]map[string]interface{})
				assert.Len(t, files, len(expectedFiles))

				// Validate pagination fields
				assert.Equal(t, tt.expectedResult["total"], result["total"])
				assert.Equal(t, tt.expectedResult["limit"], result["limit"])
				assert.Equal(t, tt.expectedResult["offset"], result["offset"])
			}
		})
	}
}

func TestWebSocketServer_FileListingParameterValidation(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-parameter-validation-test")

	// Create mock controller
	mockController := &mockMediaMTXController{
		recordingsResponse: &mediamtx.FileListResponse{
			Files:  []*mediamtx.FileMetadata{},
			Total:  0,
			Limit:  100,
			Offset: 0,
		},
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

	tests := []struct {
		name          string
		params        map[string]interface{}
		method        string
		expectedError bool
	}{
		{
			name: "invalid limit type",
			params: map[string]interface{}{
				"limit": "invalid",
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "invalid offset type",
			params: map[string]interface{}{
				"offset": "invalid",
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "negative limit",
			params: map[string]interface{}{
				"limit": -1,
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "negative offset",
			params: map[string]interface{}{
				"offset": -1,
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "limit too large",
			params: map[string]interface{}{
				"limit": 1001,
			},
			method:        "list_recordings",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response *websocket.JsonRpcResponse
			var err error

			switch tt.method {
			case "list_recordings":
				response, err = server.MethodListRecordings(tt.params, client)
			case "list_snapshots":
				response, err = server.MethodListSnapshots(tt.params, client)
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

func TestWebSocketServer_FileListingAuthentication(t *testing.T) {
	// Setup test logger
	logger := logging.NewLogger("websocket-authentication-test")

	// Create mock controller
	mockController := &mockMediaMTXController{
		recordingsResponse: &mediamtx.FileListResponse{
			Files:  []*mediamtx.FileMetadata{},
			Total:  0,
			Limit:  100,
			Offset: 0,
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
			name: "unauthenticated user",
			client: &websocket.ClientConnection{
				User: "",
				Role: "",
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "user without role",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "",
			},
			method:        "list_recordings",
			expectedError: true,
		},
		{
			name: "viewer role - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "viewer",
			},
			method:        "list_recordings",
			expectedError: false,
		},
		{
			name: "operator role - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "operator",
			},
			method:        "list_recordings",
			expectedError: false,
		},
		{
			name: "admin role - should succeed",
			client: &websocket.ClientConnection{
				User: "test_user",
				Role: "admin",
			},
			method:        "list_recordings",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{}

			var response *websocket.JsonRpcResponse
			var err error

			switch tt.method {
			case "list_recordings":
				response, err = server.MethodListRecordings(params, tt.client)
			case "list_snapshots":
				response, err = server.MethodListSnapshots(params, tt.client)
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
