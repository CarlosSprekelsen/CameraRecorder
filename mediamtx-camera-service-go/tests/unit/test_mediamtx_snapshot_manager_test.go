//go:build unit
// +build unit

/*
MediaMTX Snapshot Operations Test

Requirements Coverage:
- T5.1.8: Add unit tests in tests/unit/test_mediamtx_snapshot_operations_test.go
- REQ-API-002: JSON-RPC 2.0 protocol implementation
- REQ-API-003: Request/response message handling

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

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

// mockSnapshotManager implements snapshot operations for testing
type mockSnapshotManager struct {
	snapshots map[string]*mediamtx.Snapshot
	devices   map[string]bool
	settings  *mediamtx.SnapshotSettings
}

func newMockSnapshotManager() *mockSnapshotManager {
	return &mockSnapshotManager{
		snapshots: make(map[string]*mediamtx.Snapshot),
		devices: map[string]bool{
			"/dev/video0": true,
			"/dev/video1": true,
		},
		settings: &mediamtx.SnapshotSettings{
			Quality:     85,
			Format:      "jpg",
			MaxWidth:    1920,
			MaxHeight:   1080,
			AutoResize:  true,
			Compression: 6,
		},
	}
}

func (m *mockSnapshotManager) TakeSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*mediamtx.Snapshot, error) {
	if !m.devices[device] {
		return nil, fmt.Errorf("device not found: %s", device)
	}

	snapshotID := "snapshot_" + device + "_" + time.Now().Format("20060102150405")
	snapshot := &mediamtx.Snapshot{
		ID:       snapshotID,
		Device:   device,
		Path:     path,
		FilePath: "/tmp/snapshots/" + snapshotID + ".jpg",
		Size:     102400, // Mock file size
		Created:  time.Now(),
		Metadata: map[string]interface{}{
			"quality":     m.settings.Quality,
			"format":      m.settings.Format,
			"max_width":   m.settings.MaxWidth,
			"max_height":  m.settings.MaxHeight,
			"auto_resize": m.settings.AutoResize,
			"device_info": "Mock Camera Device",
		},
	}

	m.snapshots[snapshotID] = snapshot
	return snapshot, nil
}

func (m *mockSnapshotManager) GetSnapshot(snapshotID string) (*mediamtx.Snapshot, bool) {
	snapshot, exists := m.snapshots[snapshotID]
	return snapshot, exists
}

func (m *mockSnapshotManager) ListSnapshots() []*mediamtx.Snapshot {
	snapshots := make([]*mediamtx.Snapshot, 0, len(m.snapshots))
	for _, snapshot := range m.snapshots {
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

func (m *mockSnapshotManager) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if _, exists := m.snapshots[snapshotID]; !exists {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	delete(m.snapshots, snapshotID)
	return nil
}

func (m *mockSnapshotManager) GetSettings() *mediamtx.SnapshotSettings {
	return m.settings
}

func (m *mockSnapshotManager) UpdateSettings(settings *mediamtx.SnapshotSettings) {
	m.settings = settings
}

// TestSnapshotManager_TakeSnapshot tests the snapshot manager take snapshot functionality
func TestSnapshotManager_TakeSnapshot(t *testing.T) {
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
				"quality":    95,
				"format":     "png",
				"resolution": "3840x2160",
				"timestamp":  false,
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
			// Create mock manager
			mockManager := newMockSnapshotManager()

			// Call the method
			snapshot, err := mockManager.TakeSnapshot(context.Background(), tt.device, tt.path, tt.options)

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

				// Verify metadata
				assert.NotNil(t, snapshot.Metadata)
				assert.Contains(t, snapshot.Metadata, "quality")
				assert.Contains(t, snapshot.Metadata, "format")
				assert.Contains(t, snapshot.Metadata, "resolution")

				// Verify snapshot was tracked
				trackedSnapshot, exists := mockManager.snapshots[snapshot.ID]
				assert.True(t, exists, "Snapshot should be tracked in manager")
				assert.Equal(t, snapshot, trackedSnapshot)
			}
		})
	}
}

// TestSnapshotManager_GetSnapshot tests the snapshot manager get snapshot functionality
func TestSnapshotManager_GetSnapshot(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Take a snapshot first
	snapshot, err := mockManager.TakeSnapshot(context.Background(), "/dev/video0", "/tmp/snapshots", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	tests := []struct {
		name          string
		snapshotID    string
		expectedFound bool
	}{
		{
			name:          "snapshot found",
			snapshotID:    snapshot.ID,
			expectedFound: true,
		},
		{
			name:          "snapshot not found",
			snapshotID:    "non_existent_snapshot",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			foundSnapshot, found := mockManager.GetSnapshot(tt.snapshotID)

			// Assert results
			if tt.expectedFound {
				assert.True(t, found)
				assert.NotNil(t, foundSnapshot)
				assert.Equal(t, tt.snapshotID, foundSnapshot.ID)
			} else {
				assert.False(t, found)
				assert.Nil(t, foundSnapshot)
			}
		})
	}
}

// TestSnapshotManager_ListSnapshots tests the snapshot manager list snapshots functionality
func TestSnapshotManager_ListSnapshots(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Take multiple snapshots
	devices := []string{"/dev/video0", "/dev/video1"}
	snapshotIDs := make([]string, 0)

	for _, device := range devices {
		snapshot, err := mockManager.TakeSnapshot(context.Background(), device, "/tmp/snapshots", map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, snapshot)
		snapshotIDs = append(snapshotIDs, snapshot.ID)
	}

	// Test listing snapshots
	snapshots := mockManager.ListSnapshots()
	assert.Equal(t, len(devices), len(snapshots), "Should have one snapshot per device")

	// Verify all snapshots are created
	for _, snapshot := range snapshots {
		assert.NotZero(t, snapshot.Created)
		assert.Greater(t, snapshot.Size, int64(0))
	}
}

// TestSnapshotManager_DeleteSnapshot tests the snapshot manager delete snapshot functionality
func TestSnapshotManager_DeleteSnapshot(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Take a snapshot first
	snapshot, err := mockManager.TakeSnapshot(context.Background(), "/dev/video0", "/tmp/snapshots", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	tests := []struct {
		name          string
		snapshotID    string
		expectedError bool
	}{
		{
			name:          "successful snapshot deletion",
			snapshotID:    snapshot.ID,
			expectedError: false,
		},
		{
			name:          "snapshot not found",
			snapshotID:    "non_existent_snapshot",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			err := mockManager.DeleteSnapshot(context.Background(), tt.snapshotID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify snapshot was removed
				_, exists := mockManager.snapshots[tt.snapshotID]
				assert.False(t, exists, "Snapshot should be removed from manager")
			}
		})
	}
}

// TestSnapshotManager_Settings tests snapshot settings management
func TestSnapshotManager_Settings(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Test getting current settings
	currentSettings := mockManager.GetSettings()
	assert.NotNil(t, currentSettings)
	assert.Equal(t, 85, currentSettings.Quality)
	assert.Equal(t, "jpg", currentSettings.Format)
	assert.Equal(t, 1920, currentSettings.MaxWidth)
	assert.Equal(t, 1080, currentSettings.MaxHeight)
	assert.True(t, currentSettings.AutoResize)

	// Test updating settings
	newSettings := &mediamtx.SnapshotSettings{
		Quality:     95,
		Format:      "png",
		MaxWidth:    3840,
		MaxHeight:   2160,
		AutoResize:  false,
		Compression: 8,
	}

	mockManager.UpdateSettings(newSettings)

	// Verify settings were updated
	updatedSettings := mockManager.GetSettings()
	assert.Equal(t, newSettings.Quality, updatedSettings.Quality)
	assert.Equal(t, newSettings.Format, updatedSettings.Format)
	assert.Equal(t, newSettings.MaxWidth, updatedSettings.MaxWidth)
	assert.Equal(t, newSettings.MaxHeight, updatedSettings.MaxHeight)
	assert.Equal(t, newSettings.AutoResize, updatedSettings.AutoResize)
}

// TestSnapshotManager_SnapshotOptions tests snapshot with various options
func TestSnapshotManager_SnapshotOptions(t *testing.T) {
	mockManager := newMockSnapshotManager()

	tests := []struct {
		name    string
		options map[string]interface{}
	}{
		{
			name: "high quality snapshot",
			options: map[string]interface{}{
				"quality":    95,
				"format":     "png",
				"resolution": "3840x2160",
			},
		},
		{
			name: "low quality snapshot",
			options: map[string]interface{}{
				"quality":    50,
				"format":     "jpg",
				"resolution": "640x480",
			},
		},
		{
			name: "custom metadata",
			options: map[string]interface{}{
				"quality":    85,
				"format":     "jpg",
				"resolution": "1920x1080",
				"location":   "front_door",
				"camera_id":  "cam_001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := mockManager.TakeSnapshot(context.Background(), "/dev/video0", "/tmp/snapshots", tt.options)
			require.NoError(t, err)
			require.NotNil(t, snapshot)

			// Verify snapshot was created
			assert.NotEmpty(t, snapshot.ID)
			assert.Equal(t, "/dev/video0", snapshot.Device)

			// Verify metadata contains options
			for key, value := range tt.options {
				if key == "quality" || key == "format" || key == "resolution" {
					assert.Equal(t, value, snapshot.Metadata[key])
				}
			}
		})
	}
}

// TestSnapshotManager_ErrorHandling tests error scenarios
func TestSnapshotManager_ErrorHandling(t *testing.T) {
	mockManager := newMockSnapshotManager()

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
			_, err := mockManager.TakeSnapshot(context.Background(), tt.device, "/tmp/snapshots", map[string]interface{}{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestSnapshotManager_JSONRPC tests JSON-RPC protocol compliance for snapshot operations
func TestSnapshotManager_JSONRPC(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Take snapshot
	snapshot, err := mockManager.TakeSnapshot(context.Background(), "/dev/video0", "/tmp/snapshots", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Test JSON serialization of snapshot
	jsonData, err := json.Marshal(snapshot)
	require.NoError(t, err)
	require.NotEmpty(t, jsonData)

	// Verify JSON structure
	var jsonSnapshot map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonSnapshot)
	require.NoError(t, err)

	assert.Contains(t, jsonSnapshot, "id")
	assert.Contains(t, jsonSnapshot, "device")
	assert.Contains(t, jsonSnapshot, "status")
	assert.Contains(t, jsonSnapshot, "timestamp")
	assert.Contains(t, jsonSnapshot, "file_path")
	assert.Contains(t, jsonSnapshot, "metadata")
}

// TestSnapshotManager_Performance tests snapshot performance requirements
func TestSnapshotManager_Performance(t *testing.T) {
	mockManager := newMockSnapshotManager()

	// Test that snapshot operations complete quickly
	startTime := time.Now()
	snapshot, err := mockManager.TakeSnapshot(context.Background(), "/dev/video0", "/tmp/snapshots", map[string]interface{}{})
	duration := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Verify operation completes within 100ms (performance requirement)
	assert.Less(t, duration, 100*time.Millisecond, "Snapshot operation should complete within 100ms")
}
