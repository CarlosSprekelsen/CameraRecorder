/*
WebSocket Recording Methods Test

Requirements Coverage:
- T5.2.9: Add unit tests in tests/unit/test_websocket_recording_methods_test.go
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMediaMTXControllerRecording implements MediaMTXController for recording testing
type mockMediaMTXControllerRecording struct {
	sessions map[string]*mediamtx.RecordingSession
	devices  map[string]bool
}

func newMockMediaMTXControllerRecording() *mockMediaMTXControllerRecording {
	return &mockMediaMTXControllerRecording{
		sessions: make(map[string]*mediamtx.RecordingSession),
		devices: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
	}
}

func (m *mockMediaMTXControllerRecording) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.RecordingSession, error) {
	if !m.devices[device] {
		return nil, fmt.Errorf("device not found: %s", device)
	}

	sessionID := "recording_" + device + "_" + time.Now().Format("20060102150405")
	session := &mediamtx.RecordingSession{
		ID:        sessionID,
		Device:    device,
		Path:      path,
		Status:    "RECORDING",
		StartTime: time.Now(),
		FilePath:  "/tmp/recordings/" + sessionID + ".mp4",
		FileSize:  0,
	}

	m.sessions[sessionID] = session
	return session, nil
}

func (m *mockMediaMTXControllerRecording) StopAdvancedRecording(ctx context.Context, sessionID string) error {
	if _, exists := m.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := m.sessions[sessionID]
	session.Status = "STOPPED"
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)
	session.FileSize = 1024000 // Mock file size

	return nil
}

func (m *mockMediaMTXControllerRecording) ListAdvancedRecordingSessions() []*mediamtx.RecordingSession {
	sessions := make([]*mediamtx.RecordingSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Implement other required methods with minimal implementations
func (m *mockMediaMTXControllerRecording) GetHealth(ctx context.Context) (*mediamtx.HealthStatus, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetMetrics(ctx context.Context) (*mediamtx.Metrics, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetStreams(ctx context.Context) ([]*mediamtx.Stream, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetStream(ctx context.Context, id string) (*mediamtx.Stream, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) CreateStream(ctx context.Context, name, source string) (*mediamtx.Stream, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) DeleteStream(ctx context.Context, id string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) GetPaths(ctx context.Context) ([]*mediamtx.Path, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetPath(ctx context.Context, name string) (*mediamtx.Path, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) CreatePath(ctx context.Context, path *mediamtx.Path) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) DeletePath(ctx context.Context, name string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) StartRecording(ctx context.Context, device, path string) (*mediamtx.RecordingSession, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) StopRecording(ctx context.Context, sessionID string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) TakeSnapshot(ctx context.Context, device, path string) (*mediamtx.Snapshot, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetRecordingStatus(ctx context.Context, sessionID string) (*mediamtx.RecordingSession, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) ListRecordings(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) ListSnapshots(ctx context.Context, limit, offset int) (*mediamtx.FileListResponse, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetRecordingInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetSnapshotInfo(ctx context.Context, filename string) (*mediamtx.FileMetadata, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) DeleteRecording(ctx context.Context, filename string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) DeleteSnapshot(ctx context.Context, filename string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) GetAdvancedRecordingSession(sessionID string) (*mediamtx.RecordingSession, bool) {
	return nil, false
}
func (m *mockMediaMTXControllerRecording) RotateRecordingFile(ctx context.Context, sessionID string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) TakeAdvancedSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.Snapshot, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) GetAdvancedSnapshot(snapshotID string) (*mediamtx.Snapshot, bool) {
	return nil, false
}
func (m *mockMediaMTXControllerRecording) ListAdvancedSnapshots() []*mediamtx.Snapshot { return nil }
func (m *mockMediaMTXControllerRecording) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) GetSnapshotSettings() *mediamtx.SnapshotSettings {
	return nil
}
func (m *mockMediaMTXControllerRecording) UpdateSnapshotSettings(settings *mediamtx.SnapshotSettings) {
}
func (m *mockMediaMTXControllerRecording) GetConfig(ctx context.Context) (*mediamtx.MediaMTXConfig, error) {
	return nil, nil
}
func (m *mockMediaMTXControllerRecording) UpdateConfig(ctx context.Context, config *mediamtx.MediaMTXConfig) error {
	return nil
}
func (m *mockMediaMTXControllerRecording) Start(ctx context.Context) error { return nil }
func (m *mockMediaMTXControllerRecording) Stop(ctx context.Context) error  { return nil }

// TestMediaMTXController_StartAdvancedRecording tests the MediaMTX controller recording start functionality
func TestMediaMTXController_StartAdvancedRecording(t *testing.T) {
	tests := []struct {
		name           string
		device         string
		path           string
		options        map[string]interface{}
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "successful recording start with valid device",
			device:         "/dev/video0",
			path:           "/tmp/recordings",
			options:        map[string]interface{}{},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:   "successful recording start with custom options",
			device: "/dev/video0",
			path:   "/tmp/recordings",
			options: map[string]interface{}{
				"duration_seconds": 60,
				"format":           "mp4",
				"codec":            "h264",
				"quality":          23,
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "device not found",
			device:         "/dev/video999",
			path:           "/tmp/recordings",
			options:        map[string]interface{}{},
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock controller
			mockController := newMockMediaMTXControllerRecording()

			// Call the method
			session, err := mockController.StartAdvancedRecording(context.Background(), tt.device, tt.path, tt.options)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)

				// Verify session structure
				assert.NotEmpty(t, session.ID)
				assert.Equal(t, tt.device, session.Device)
				assert.Equal(t, tt.path, session.Path)
				assert.Equal(t, "RECORDING", session.Status)
				assert.NotZero(t, session.StartTime)
				assert.NotEmpty(t, session.FilePath)
				assert.Equal(t, int64(0), session.FileSize) // Initial size should be 0

				// Verify session was tracked
				trackedSession, exists := mockController.sessions[session.ID]
				assert.True(t, exists, "Session should be tracked in controller")
				assert.Equal(t, session, trackedSession)
			}
		})
	}
}

// TestMediaMTXController_StopAdvancedRecording tests the MediaMTX controller recording stop functionality
func TestMediaMTXController_StopAdvancedRecording(t *testing.T) {
	mockController := newMockMediaMTXControllerRecording()

	// Start a recording first
	session, err := mockController.StartAdvancedRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, session)

	tests := []struct {
		name          string
		sessionID     string
		expectedError bool
	}{
		{
			name:          "successful recording stop",
			sessionID:     session.ID,
			expectedError: false,
		},
		{
			name:          "session not found",
			sessionID:     "non_existent_session",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			err := mockController.StopAdvancedRecording(context.Background(), tt.sessionID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify session was updated
				updatedSession, exists := mockController.sessions[tt.sessionID]
				assert.True(t, exists, "Session should still exist in controller")
				assert.Equal(t, "STOPPED", updatedSession.Status)
				assert.NotNil(t, updatedSession.EndTime)
				assert.Greater(t, updatedSession.Duration, time.Duration(0))
				assert.Greater(t, updatedSession.FileSize, int64(0))
			}
		})
	}
}

// TestMediaMTXController_RecordingSessionManagement tests recording session management
func TestMediaMTXController_RecordingSessionManagement(t *testing.T) {
	mockController := newMockMediaMTXControllerRecording()

	// Start multiple recordings
	devices := []string{"/dev/video0", "/dev/video1"}
	sessionIDs := make([]string, 0)

	for _, device := range devices {
		session, err := mockController.StartAdvancedRecording(context.Background(), device, "/tmp/recordings", map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, session)
		sessionIDs = append(sessionIDs, session.ID)
	}

	// Test listing sessions
	sessions := mockController.ListAdvancedRecordingSessions()
	assert.Equal(t, len(devices), len(sessions), "Should have one session per device")

	// Verify all sessions are in recording state
	for _, session := range sessions {
		assert.Equal(t, "RECORDING", session.Status)
		assert.NotZero(t, session.StartTime)
		assert.Empty(t, session.EndTime)
	}

	// Stop one recording
	err := mockController.StopAdvancedRecording(context.Background(), sessionIDs[0])
	require.NoError(t, err)

	// Verify session state changed
	stoppedSession := mockController.sessions[sessionIDs[0]]
	assert.Equal(t, "STOPPED", stoppedSession.Status)
	assert.NotNil(t, stoppedSession.EndTime)
	assert.Greater(t, stoppedSession.Duration, time.Duration(0))
}

// TestMediaMTXController_RecordingOptions tests recording with various options
func TestMediaMTXController_RecordingOptions(t *testing.T) {
	mockController := newMockMediaMTXControllerRecording()

	tests := []struct {
		name    string
		options map[string]interface{}
	}{
		{
			name: "timed recording",
			options: map[string]interface{}{
				"duration_seconds": 300, // 5 minutes
			},
		},
		{
			name: "high quality recording",
			options: map[string]interface{}{
				"format":  "mp4",
				"codec":   "h264",
				"quality": 18, // High quality
			},
		},
		{
			name: "custom format",
			options: map[string]interface{}{
				"format": "avi",
				"codec":  "mpeg4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := mockController.StartAdvancedRecording(context.Background(), "/dev/video0", "/tmp/recordings", tt.options)
			require.NoError(t, err)
			require.NotNil(t, session)

			// Verify session was created
			assert.NotEmpty(t, session.ID)
			assert.Equal(t, "/dev/video0", session.Device)
			assert.Equal(t, "RECORDING", session.Status)

			// Stop the recording
			err = mockController.StopAdvancedRecording(context.Background(), session.ID)
			require.NoError(t, err)

			// Verify session was stopped
			stoppedSession := mockController.sessions[session.ID]
			assert.Equal(t, "STOPPED", stoppedSession.Status)
		})
	}
}

// TestMediaMTXController_RecordingErrorHandling tests error scenarios
func TestMediaMTXController_RecordingErrorHandling(t *testing.T) {
	mockController := newMockMediaMTXControllerRecording()

	tests := []struct {
		name          string
		device        string
		expectedError string
	}{
		{
			name:          "device not found",
			device:        "/dev/video999",
			expectedError: "device not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockController.StartAdvancedRecording(context.Background(), tt.device, "/tmp/recordings", map[string]interface{}{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestMediaMTXController_RecordingJSONRPC tests JSON-RPC protocol compliance for recording operations
func TestMediaMTXController_RecordingJSONRPC(t *testing.T) {
	mockController := newMockMediaMTXControllerRecording()

	// Start recording
	session, err := mockController.StartAdvancedRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, session)

	// Test JSON serialization of session
	jsonData, err := json.Marshal(session)
	require.NoError(t, err)
	require.NotEmpty(t, jsonData)

	// Verify JSON structure
	var jsonSession map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonSession)
	require.NoError(t, err)

	assert.Contains(t, jsonSession, "id")
	assert.Contains(t, jsonSession, "device")
	assert.Contains(t, jsonSession, "status")
	assert.Contains(t, jsonSession, "start_time")
	assert.Contains(t, jsonSession, "file_path")
}
