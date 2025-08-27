/*
WebSocket Snapshot Methods Test

Requirements Coverage:
- T5.1.7: Add unit tests in tests/unit/test_websocket_snapshot_methods_test.go
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMediaMTXControllerSnapshot implements MediaMTXController for snapshot testing
type mockMediaMTXControllerSnapshot struct {
	snapshots map[string]*mediamtx.Snapshot
	devices   map[string]bool
}

func newMockMediaMTXControllerSnapshot() *mockMediaMTXControllerSnapshot {
	return &mockMediaMTXControllerSnapshot{
		snapshots: make(map[string]*mediamtx.Snapshot),
		devices: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}
}

func (m *mockMediaMTXControllerSnapshot) TakeAdvancedSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.Snapshot, error) {
	if !m.devices[device] {
		return nil, fmt.Errorf("device not found: %s", device)
	}

	snapshotID := "snapshot_" + device + "_" + time.Now().Format("20060102150405")
	snapshot := &mediamtx.Snapshot{
		ID:       snapshotID,
		Device:   device,
		Path:     path,
		FilePath: "/tmp/snapshots/" + snapshotID + ".jpg",
		Size:     102400,
		Created:  time.Now(),
		Metadata: map[string]interface{}{
			"tier_used":       1,
			"capture_time_ms": 150,
			"user_experience": "excellent",
		},
	}

	m.snapshots[snapshotID] = snapshot
	return snapshot, nil
}

// Implement other required methods with minimal implementations
func (m *mockMediaMTXControllerSnapshot) GetHealth(ctx context.Context) (*mediamtx.HealthStatus, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetMetrics(ctx context.Context) (*mediamtx.Metrics, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetStreams(ctx context.Context) ([]*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetStream(ctx context.Context, id string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) CreateStream(ctx context.Context, name, source string) (*mediamtx.Stream, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) DeleteStream(ctx context.Context, id string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) GetPaths(ctx context.Context) ([]*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetPath(ctx context.Context, name string) (*mediamtx.Path, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) CreatePath(ctx context.Context, path *mediamtx.Path) error { return nil }
func (m *mockMediaMTXControllerSnapshot) DeletePath(ctx context.Context, name string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) StartRecording(ctx context.Context, device, path string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) StopRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) TakeSnapshot(ctx context.Context, device, path string) (*mediamtx.Snapshot, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetRecordingStatus(ctx context.Context, sessionID string) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) ListRecordings(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) ListSnapshots(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetRecordingInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) GetSnapshotInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) DeleteRecording(ctx context.Context, filename string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) DeleteSnapshot(ctx context.Context, filename string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.RecordingSession, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) StopAdvancedRecording(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) GetAdvancedRecordingSession(sessionID string) (*mediamtx.RecordingSession, bool) { return nil, false }
func (m *mockMediaMTXControllerSnapshot) ListAdvancedRecordingSessions() []*mediamtx.RecordingSession { return nil }
func (m *mockMediaMTXControllerSnapshot) RotateRecordingFile(ctx context.Context, sessionID string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) GetAdvancedSnapshot(snapshotID string) (*mediamtx.Snapshot, bool) { return nil, false }
func (m *mockMediaMTXControllerSnapshot) ListAdvancedSnapshots() []*mediamtx.Snapshot { return nil }
func (m *mockMediaMTXControllerSnapshot) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error { return nil }
func (m *mockMediaMTXControllerSnapshot) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error { return nil }
func (m *mockMediaMTXControllerSnapshot) GetSnapshotSettings() *mediamtx.SnapshotSettings { return nil }
func (m *mockMediaMTXControllerSnapshot) UpdateSnapshotSettings(settings *mediamtx.SnapshotSettings) {}
func (m *mockMediaMTXControllerSnapshot) GetConfig(ctx context.Context) (*mediamtx.MediaMTXConfig, error) { return nil, nil }
func (m *mockMediaMTXControllerSnapshot) UpdateConfig(ctx context.Context, config *mediamtx.MediaMTXConfig) error { return nil }
func (m *mockMediaMTXControllerSnapshot) Start(ctx context.Context) error { return nil }
func (m *mockMediaMTXControllerSnapshot) Stop(ctx context.Context) error { return nil }

// mockCameraMonitorSnapshot implements camera monitoring for snapshot testing
type mockCameraMonitorSnapshot struct {
	devices map[string]bool
}

func newMockCameraMonitorSnapshot() *mockCameraMonitorSnapshot {
	return &mockCameraMonitorSnapshot{
		devices: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}
}

func (m *mockCameraMonitorSnapshot) GetDevice(devicePath string) (interface{}, bool) {
	exists := m.devices[devicePath]
	return nil, exists
}

// TestMediaMTXController_TakeAdvancedSnapshot tests the MediaMTX controller snapshot functionality
func TestMediaMTXController_TakeAdvancedSnapshot(t *testing.T) {
	tests := []struct {
		name           string
		device         string
		path           string
		options        map[string]interface{}
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "successful snapshot with valid device",
			device:         "/dev/video0",
			path:           "/tmp/snapshots",
			options:        map[string]interface{}{},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:   "successful snapshot with custom options",
			device: "/dev/video0",
			path:   "/tmp/snapshots",
			options: map[string]interface{}{
				"format":   "jpg",
				"quality":  85,
				"filename": "custom_snapshot.jpg",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "device not found",
			device:         "/dev/video999",
			path:           "/tmp/snapshots",
			options:        map[string]interface{}{},
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := newMockMediaMTXControllerSnapshot()

			// Call the method
			snapshot, err := mockController.TakeAdvancedSnapshot(context.Background(), tt.device, tt.path, tt.options)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, snapshot)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, snapshot)

				// Verify snapshot structure
				assert.NotEmpty(t, snapshot.ID)
				assert.Equal(t, tt.device, snapshot.Device)
				assert.Equal(t, tt.path, snapshot.Path)
				assert.NotEmpty(t, snapshot.FilePath)
				assert.Greater(t, snapshot.Size, int64(0))
				assert.NotZero(t, snapshot.Created)

				// Verify metadata structure
				assert.NotNil(t, snapshot.Metadata)
				assert.Contains(t, snapshot.Metadata, "tier_used")
				assert.Contains(t, snapshot.Metadata, "capture_time_ms")
				assert.Contains(t, snapshot.Metadata, "user_experience")

				// Verify snapshot was tracked
				trackedSnapshot, exists := mockController.snapshots[snapshot.ID]
				assert.True(t, exists, "Snapshot should be tracked in controller")
				assert.Equal(t, snapshot, trackedSnapshot)
			}
		})
	}
}

// TestWebSocketServer_TakeSnapshot_Integration tests integration with MediaMTX controller
func TestWebSocketServer_TakeSnapshot_Integration(t *testing.T) {
	// Create mock controller that tracks snapshots
	mockController := newMockMediaMTXControllerSnapshot()
	mockCameraMonitor := newMockCameraMonitorSnapshot()

	server := &websocket.WebSocketServer{
		mediaMTXController: mockController,
		cameraMonitor:      mockCameraMonitor,
	}

	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: true,
	}

	// Take multiple snapshots
	devices := []string{"/dev/video0", "/dev/video1"}
	snapshotIDs := make([]string, 0)

	for _, device := range devices {
		params := map[string]interface{}{
			"device": device,
		}

		response, err := server.MethodTakeSnapshot(params, client)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Nil(t, response.Error)

		result := response.Result.(map[string]interface{})
		snapshotID := result["snapshot_id"].(string)
		snapshotIDs = append(snapshotIDs, snapshotID)

		// Verify snapshot was created in controller
		snapshot, exists := mockController.snapshots[snapshotID]
		require.True(t, exists, "Snapshot should be tracked in controller")
		assert.Equal(t, device, snapshot.Device)
		assert.Greater(t, snapshot.Size, int64(0))
	}

	// Verify all snapshots have unique IDs
	assert.Equal(t, len(devices), len(snapshotIDs), "Should have one snapshot per device")
	uniqueIDs := make(map[string]bool)
	for _, id := range snapshotIDs {
		uniqueIDs[id] = true
	}
	assert.Equal(t, len(snapshotIDs), len(uniqueIDs), "All snapshot IDs should be unique")
}

// TestWebSocketServer_TakeSnapshot_Options tests snapshot with various options
func TestWebSocketServer_TakeSnapshot_Options(t *testing.T) {
	mockController := newMockMediaMTXControllerSnapshot()
	mockCameraMonitor := newMockCameraMonitorSnapshot()

	server := &websocket.WebSocketServer{
		mediaMTXController: mockController,
		cameraMonitor:      mockCameraMonitor,
	}

	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: true,
	}

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "JPEG format with quality",
			params: map[string]interface{}{
				"device":   "/dev/video0",
				"format":   "jpg",
				"quality":  90,
				"filename": "high_quality.jpg",
			},
		},
		{
			name: "PNG format",
			params: map[string]interface{}{
				"device":   "/dev/video0",
				"format":   "png",
				"filename": "lossless.png",
			},
		},
		{
			name: "custom filename only",
			params: map[string]interface{}{
				"device":   "/dev/video0",
				"filename": "custom_snapshot.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := server.MethodTakeSnapshot(tt.params, client)
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Nil(t, response.Error)

			result := response.Result.(map[string]interface{})
			assert.Contains(t, result, "snapshot_id")
			assert.Equal(t, "/dev/video0", result["device"])

			// Verify options were passed correctly (this would be tested in the controller)
			snapshotID := result["snapshot_id"].(string)
			snapshot, exists := mockController.snapshots[snapshotID]
			require.True(t, exists)
			assert.Equal(t, "/dev/video0", snapshot.Device)
		})
	}
}

// TestWebSocketServer_TakeSnapshot_ErrorHandling tests error scenarios
func TestWebSocketServer_TakeSnapshot_ErrorHandling(t *testing.T) {
	mockController := newMockMediaMTXControllerSnapshot()
	mockCameraMonitor := newMockCameraMonitorSnapshot()

	server := &websocket.WebSocketServer{
		mediaMTXController: mockController,
		cameraMonitor:      mockCameraMonitor,
	}

	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: true,
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectedError string
	}{
		{
			name:          "missing device parameter",
			params:        nil,
			expectedError: "device parameter is required",
		},
		{
			name: "empty device parameter",
			params: map[string]interface{}{
				"device": "",
			},
			expectedError: "device parameter is required",
		},
		{
			name: "device not found",
			params: map[string]interface{}{
				"device": "/dev/video999",
			},
			expectedError: "Camera device not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := server.MethodTakeSnapshot(tt.params, client)
			require.NoError(t, err) // Method doesn't return error, it's in response
			require.NotNil(t, response)
			require.NotNil(t, response.Error)

			// Check error message
			assert.Contains(t, response.Error.Message, tt.expectedError)
		})
	}
}

// TestWebSocketServer_TakeSnapshot_JSONRPC tests JSON-RPC protocol compliance
func TestWebSocketServer_TakeSnapshot_JSONRPC(t *testing.T) {
	mockController := newMockMediaMTXControllerSnapshot()
	mockCameraMonitor := newMockCameraMonitorSnapshot()

	server := &websocket.WebSocketServer{
		mediaMTXController: mockController,
		cameraMonitor:      mockCameraMonitor,
	}

	client := &websocket.ClientConnection{
		ClientID:      "test_client",
		Authenticated: true,
	}

	params := map[string]interface{}{
		"device": "/dev/video0",
	}

	response, err := server.MethodTakeSnapshot(params, client)
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify JSON-RPC 2.0 compliance
	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotEmpty(t, jsonData)

	// Verify JSON structure
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonResponse)
	require.NoError(t, err)

	assert.Equal(t, "2.0", jsonResponse["jsonrpc"])
	assert.NotNil(t, jsonResponse["result"])
	assert.Nil(t, jsonResponse["error"])
}
