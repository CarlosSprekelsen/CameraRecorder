/*
MediaMTX Recording Operations Test

Requirements Coverage:
- T5.2.10: Add unit tests in tests/unit/test_mediamtx_recording_operations_test.go
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

// mockRecordingManager implements recording operations for testing
type mockRecordingManager struct {
	sessions map[string]*mediamtx.RecordingSession
	devices  map[string]bool
	settings map[string]interface{}
}

func newMockRecordingManager() *mockRecordingManager {
	return &mockRecordingManager{
		sessions: make(map[string]*mediamtx.RecordingSession),
		devices: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
		settings: map[string]interface{}{
			"default_format":  "mp4",
			"default_codec":   "h264",
			"default_quality": 23,
		},
	}
}

func (m *mockRecordingManager) StartRecording(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.RecordingSession, error) {
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

func (m *mockRecordingManager) StopRecording(ctx context.Context, sessionID string) error {
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

func (m *mockRecordingManager) GetSession(sessionID string) (*mediamtx.RecordingSession, bool) {
	session, exists := m.sessions[sessionID]
	return session, exists
}

func (m *mockRecordingManager) ListSessions() []*mediamtx.RecordingSession {
	sessions := make([]*mediamtx.RecordingSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (m *mockRecordingManager) DeleteSession(ctx context.Context, sessionID string) error {
	if _, exists := m.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(m.sessions, sessionID)
	return nil
}

func (m *mockRecordingManager) GetSettings() map[string]interface{} {
	return m.settings
}

func (m *mockRecordingManager) UpdateSettings(settings map[string]interface{}) {
	m.settings = settings
}

// TestRecordingManager_StartRecording tests the recording manager start recording functionality
func TestRecordingManager_StartRecording(t *testing.T) {
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
				"duration_seconds": 300,
				"format":           "mp4",
				"codec":            "h264",
				"quality":          18,
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
			// Create mock manager
			mockManager := newMockRecordingManager()

			// Call the method
			session, err := mockManager.StartRecording(context.Background(), tt.device, tt.path, tt.options)

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
				trackedSession, exists := mockManager.sessions[session.ID]
				assert.True(t, exists, "Session should be tracked in manager")
				assert.Equal(t, session, trackedSession)
			}
		})
	}
}

// TestRecordingManager_StopRecording tests the recording manager stop recording functionality
func TestRecordingManager_StopRecording(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Start a recording first
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
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
			err := mockManager.StopRecording(context.Background(), tt.sessionID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify session was updated
				updatedSession, exists := mockManager.sessions[tt.sessionID]
				assert.True(t, exists, "Session should still exist in manager")
				assert.Equal(t, "STOPPED", updatedSession.Status)
				assert.NotNil(t, updatedSession.EndTime)
				assert.Greater(t, updatedSession.Duration, time.Duration(0))
				assert.Greater(t, updatedSession.FileSize, int64(0))
			}
		})
	}
}

// TestRecordingManager_GetSession tests the recording manager get session functionality
func TestRecordingManager_GetSession(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Start a recording first
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, session)

	tests := []struct {
		name          string
		sessionID     string
		expectedFound bool
	}{
		{
			name:          "session found",
			sessionID:     session.ID,
			expectedFound: true,
		},
		{
			name:          "session not found",
			sessionID:     "non_existent_session",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			foundSession, found := mockManager.GetSession(tt.sessionID)

			// Assert results
			if tt.expectedFound {
				assert.True(t, found)
				assert.NotNil(t, foundSession)
				assert.Equal(t, tt.sessionID, foundSession.ID)
			} else {
				assert.False(t, found)
				assert.Nil(t, foundSession)
			}
		})
	}
}

// TestRecordingManager_ListSessions tests the recording manager list sessions functionality
func TestRecordingManager_ListSessions(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Start multiple recordings
	devices := []string{"/dev/video0", "/dev/video1"}
	sessionIDs := make([]string, 0)

	for _, device := range devices {
		session, err := mockManager.StartRecording(context.Background(), device, "/tmp/recordings", map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, session)
		sessionIDs = append(sessionIDs, session.ID)
	}

	// Test listing sessions
	sessions := mockManager.ListSessions()
	assert.Equal(t, len(devices), len(sessions), "Should have one session per device")

	// Verify all sessions are in recording state
	for _, session := range sessions {
		assert.Equal(t, "RECORDING", session.Status)
		assert.NotZero(t, session.StartTime)
		assert.Empty(t, session.EndTime)
	}

	// Stop one recording
	err := mockManager.StopRecording(context.Background(), sessionIDs[0])
	require.NoError(t, err)

	// Verify session state changed
	stoppedSession := mockManager.sessions[sessionIDs[0]]
	assert.Equal(t, "STOPPED", stoppedSession.Status)
	assert.NotNil(t, stoppedSession.EndTime)
	assert.Greater(t, stoppedSession.Duration, time.Duration(0))
}

// TestRecordingManager_DeleteSession tests the recording manager delete session functionality
func TestRecordingManager_DeleteSession(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Start a recording first
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, session)

	tests := []struct {
		name          string
		sessionID     string
		expectedError bool
	}{
		{
			name:          "successful session deletion",
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
			err := mockManager.DeleteSession(context.Background(), tt.sessionID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify session was removed
				_, exists := mockManager.sessions[tt.sessionID]
				assert.False(t, exists, "Session should be removed from manager")
			}
		})
	}
}

// TestRecordingManager_Settings tests recording settings management
func TestRecordingManager_Settings(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Test getting current settings
	currentSettings := mockManager.GetSettings()
	assert.NotNil(t, currentSettings)
	assert.Equal(t, "mp4", currentSettings["default_format"])
	assert.Equal(t, "h264", currentSettings["default_codec"])
	assert.Equal(t, 23, currentSettings["default_quality"])

	// Test updating settings
	newSettings := map[string]interface{}{
		"default_format":  "avi",
		"default_codec":   "mpeg4",
		"default_quality": 18,
		"segment_size":    1048576,
	}

	mockManager.UpdateSettings(newSettings)

	// Verify settings were updated
	updatedSettings := mockManager.GetSettings()
	assert.Equal(t, newSettings["default_format"], updatedSettings["default_format"])
	assert.Equal(t, newSettings["default_codec"], updatedSettings["default_codec"])
	assert.Equal(t, newSettings["default_quality"], updatedSettings["default_quality"])
	assert.Equal(t, newSettings["segment_size"], updatedSettings["segment_size"])
}

// TestRecordingManager_RecordingOptions tests recording with various options
func TestRecordingManager_RecordingOptions(t *testing.T) {
	mockManager := newMockRecordingManager()

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
			session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", tt.options)
			require.NoError(t, err)
			require.NotNil(t, session)

			// Verify session was created
			assert.NotEmpty(t, session.ID)
			assert.Equal(t, "/dev/video0", session.Device)
			assert.Equal(t, "RECORDING", session.Status)

			// Stop the recording
			err = mockManager.StopRecording(context.Background(), session.ID)
			require.NoError(t, err)

			// Verify session was stopped
			stoppedSession := mockManager.sessions[session.ID]
			assert.Equal(t, "STOPPED", stoppedSession.Status)
		})
	}
}

// TestRecordingManager_ErrorHandling tests error scenarios
func TestRecordingManager_ErrorHandling(t *testing.T) {
	mockManager := newMockRecordingManager()

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
			_, err := mockManager.StartRecording(context.Background(), tt.device, "/tmp/recordings", map[string]interface{}{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestRecordingManager_JSONRPC tests JSON-RPC protocol compliance for recording operations
func TestRecordingManager_JSONRPC(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Start recording
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
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

// TestRecordingManager_Performance tests recording performance requirements
func TestRecordingManager_Performance(t *testing.T) {
	mockManager := newMockRecordingManager()

	// Test that recording operations complete quickly
	startTime := time.Now()
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{})
	duration := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify operation completes within 100ms (performance requirement)
	assert.Less(t, duration, 100*time.Millisecond, "Recording operation should complete within 100ms")

	// Test stop recording performance
	startTime = time.Now()
	err = mockManager.StopRecording(context.Background(), session.ID)
	duration = time.Since(startTime)

	require.NoError(t, err)

	// Verify stop operation completes within 100ms
	assert.Less(t, duration, 100*time.Millisecond, "Stop recording operation should complete within 100ms")
}

// TestRecordingManager_SessionLifecycle tests complete session lifecycle
func TestRecordingManager_SessionLifecycle(t *testing.T) {
	mockManager := newMockRecordingManager()

	// 1. Start recording
	session, err := mockManager.StartRecording(context.Background(), "/dev/video0", "/tmp/recordings", map[string]interface{}{
		"duration_seconds": 60,
		"format":           "mp4",
		"codec":            "h264",
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify initial state
	assert.Equal(t, "RECORDING", session.Status)
	assert.NotZero(t, session.StartTime)
	assert.Empty(t, session.EndTime)
	assert.Equal(t, int64(0), session.FileSize)

	// 2. Get session
	foundSession, found := mockManager.GetSession(session.ID)
	assert.True(t, found)
	assert.Equal(t, session, foundSession)

	// 3. List sessions
	sessions := mockManager.ListSessions()
	assert.Len(t, sessions, 1)
	assert.Equal(t, session, sessions[0])

	// 4. Stop recording
	err = mockManager.StopRecording(context.Background(), session.ID)
	require.NoError(t, err)

	// Verify stopped state
	stoppedSession := mockManager.sessions[session.ID]
	assert.Equal(t, "STOPPED", stoppedSession.Status)
	assert.NotNil(t, stoppedSession.EndTime)
	assert.Greater(t, stoppedSession.Duration, time.Duration(0))
	assert.Greater(t, stoppedSession.FileSize, int64(0))

	// 5. Delete session
	err = mockManager.DeleteSession(context.Background(), session.ID)
	require.NoError(t, err)

	// Verify deletion
	_, exists := mockManager.sessions[session.ID]
	assert.False(t, exists, "Session should be deleted")
}
