/*
Snapshot Manager Test Asserters - Eliminate Massive Duplication

This file provides domain-specific asserters for snapshot manager tests that eliminate
the massive duplication found in snapshot_manager_test.go (1,176 lines).

Duplication Patterns Eliminated:
- SetupMediaMTXTest + GetReadyController (15 times)
- Camera ID retrieval (15+ times)
- Progressive Readiness pattern (15+ times)
- File validation (15+ times)
- Snapshot settings validation (8+ times)

Usage:
    asserter := NewSnapshotManagerAsserter(t)
    defer asserter.Cleanup()
    // Test-specific logic only
    asserter.AssertSnapshotCapture(cameraID, options)
*/

package mediamtx

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SnapshotManagerAsserter encapsulates all snapshot manager test patterns
type SnapshotManagerAsserter struct {
	t               *testing.T
	helper          *MediaMTXTestHelper
	ctx             context.Context
	cancel          context.CancelFunc
	controller      MediaMTXController
	snapshotManager *SnapshotManager
}

// NewSnapshotManagerAsserter creates a new snapshot manager asserter with full setup
// Eliminates: helper, _ := SetupMediaMTXTest(t) + GetReadyController pattern
func NewSnapshotManagerAsserter(t *testing.T) *SnapshotManagerAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)
	snapshotManager := helper.GetSnapshotManager()

	return &SnapshotManagerAsserter{
		t:               t,
		helper:          helper,
		ctx:             ctx,
		cancel:          cancel,
		controller:      controller,
		snapshotManager: snapshotManager,
	}
}

// Cleanup stops the controller and cancels the context
func (sma *SnapshotManagerAsserter) Cleanup() {
	sma.controller.Stop(sma.ctx)
	sma.cancel()
}

// GetHelper returns the underlying MediaMTXTestHelper
func (sma *SnapshotManagerAsserter) GetHelper() *MediaMTXTestHelper {
	return sma.helper
}

// GetSnapshotManager returns the snapshot manager instance
func (sma *SnapshotManagerAsserter) GetSnapshotManager() *SnapshotManager {
	return sma.snapshotManager
}

// GetController returns the controller instance
func (sma *SnapshotManagerAsserter) GetController() MediaMTXController {
	return sma.controller
}

// GetContext returns the test context
func (sma *SnapshotManagerAsserter) GetContext() context.Context {
	return sma.ctx
}

// MustGetCameraID gets an available camera ID with validation
// Eliminates: camera ID retrieval duplication
func (sma *SnapshotManagerAsserter) MustGetCameraID() string {
	cameraID, err := sma.helper.GetAvailableCameraIdentifierFromController(sma.ctx, sma.controller)
	require.NoError(sma.t, err, "Must have available camera for snapshot testing")
	return cameraID
}

// AssertSnapshotCapture performs complete snapshot capture with Progressive Readiness
// Eliminates 30+ lines of snapshot test duplication
func (sma *SnapshotManagerAsserter) AssertSnapshotCapture(cameraID string, options *SnapshotOptions) *TakeSnapshotResponse {
	// Use Progressive Readiness pattern
	result := testutils.TestProgressiveReadiness(sma.t, func() (*TakeSnapshotResponse, error) {
		return sma.snapshotManager.TakeSnapshot(sma.ctx, cameraID, options)
	}, sma.controller, "TakeSnapshot")

	require.NoError(sma.t, result.Error, "Snapshot must succeed")
	require.NotNil(sma.t, result.Result, "Snapshot response must not be nil")

	if result.UsedFallback {
		sma.t.Log("⚠️  PROGRESSIVE READINESS FALLBACK: Snapshot needed readiness event")
	} else {
		sma.t.Log("✅ PROGRESSIVE READINESS: Snapshot succeeded immediately")
	}

	response := result.Result

	// Validate response structure
	assert.Equal(sma.t, cameraID, response.Device, "Response device should match camera ID")
	assert.NotEmpty(sma.t, response.Timestamp, "Response should include timestamp")
	assert.NotEmpty(sma.t, response.Filename, "Response should include snapshot filename")
	assert.NotEmpty(sma.t, response.FilePath, "Response should include file path")

	// Validate actual file creation
	sma.AssertSnapshotFileExists(response.FilePath, response.Device)

	return response
}

// AssertSnapshotFileExists validates that a snapshot file exists and has meaningful content
// Eliminates file validation duplication
func (sma *SnapshotManagerAsserter) AssertSnapshotFileExists(filePath, deviceID string) {
	// Verify file exists
	require.FileExists(sma.t, filePath, "Snapshot file must exist: %s", filePath)

	// Verify file is in configured directory
	configuredPath := sma.helper.GetConfiguredSnapshotPath()
	assert.True(sma.t, strings.HasPrefix(filePath, configuredPath),
		"Snapshot file should be in configured directory: %s", configuredPath)

	// Verify file has meaningful content
	fileInfo, err := os.Stat(filePath)
	require.NoError(sma.t, err, "Should be able to stat snapshot file")

	minSize := int64(testutils.UniversalMinSnapshotFileSize) // 1KB from universal_constants.go
	assert.Greater(sma.t, fileInfo.Size(), minSize,
		"Snapshot file must have meaningful content (>%d bytes), got %d bytes", minSize, fileInfo.Size())

	sma.t.Logf("✅ Snapshot file validated: %s (%d bytes)", filePath, fileInfo.Size())
}

// AssertSnapshotList gets snapshot list with validation
// Eliminates snapshot list testing duplication
func (sma *SnapshotManagerAsserter) AssertSnapshotList(limit, offset int) *FileListResponse {
	snapshots, err := sma.snapshotManager.GetSnapshotsList(sma.ctx, limit, offset)
	require.NoError(sma.t, err, "GetSnapshotsList should succeed")
	require.NotNil(sma.t, snapshots, "Snapshot list should not be nil")

	sma.t.Logf("✅ Snapshot list retrieved: %d items (limit=%d, offset=%d)", len(snapshots.Files), limit, offset)
	return snapshots
}

// AssertSnapshotInfo gets snapshot info with validation
// Eliminates snapshot info testing duplication
func (sma *SnapshotManagerAsserter) AssertSnapshotInfo(filePath string) *GetSnapshotInfoResponse {
	info, err := sma.snapshotManager.GetSnapshotInfo(sma.ctx, filePath)
	require.NoError(sma.t, err, "GetSnapshotInfo should succeed")
	require.NotNil(sma.t, info, "Snapshot info should not be nil")

	sma.t.Logf("✅ Snapshot info retrieved for: %s", filePath)
	return info
}

// AssertSnapshotSettings gets snapshot settings with validation
// Eliminates snapshot settings testing duplication
func (sma *SnapshotManagerAsserter) AssertSnapshotSettings() *SnapshotSettings {
	settings := sma.snapshotManager.GetSnapshotSettings()
	require.NotNil(sma.t, settings, "Snapshot settings should not be nil")

	sma.t.Log("✅ Snapshot settings retrieved successfully")
	return settings
}

// AssertDeleteSnapshotFile deletes a snapshot file with validation
// Eliminates snapshot deletion testing duplication
func (sma *SnapshotManagerAsserter) AssertDeleteSnapshotFile(filePath string) {
	err := sma.snapshotManager.DeleteSnapshotFile(sma.ctx, filePath)
	require.NoError(sma.t, err, "DeleteSnapshotFile should succeed")

	// Verify file was actually deleted
	assert.NoFileExists(sma.t, filePath, "Snapshot file should be deleted: %s", filePath)
	sma.t.Logf("✅ Snapshot file deleted: %s", filePath)
}

// AssertCleanupOldSnapshots cleans up old snapshots with validation
// Eliminates cleanup testing duplication
func (sma *SnapshotManagerAsserter) AssertCleanupOldSnapshots(maxAge time.Duration) int {
	count, totalSize, err := sma.snapshotManager.CleanupOldSnapshots(sma.ctx, maxAge, 0, 0)
	require.NoError(sma.t, err, "CleanupOldSnapshots should succeed")

	sma.t.Logf("✅ Cleanup completed: %d old snapshots removed (%d bytes, maxAge=%v)", count, totalSize, maxAge)
	return count
}

// CreateTestSnapshot creates a test snapshot for testing purposes
// Eliminates test setup duplication
func (sma *SnapshotManagerAsserter) CreateTestSnapshot(cameraID string) *TakeSnapshotResponse {
	options := &SnapshotOptions{
		Format:     "jpg",
		Quality:    85,
		MaxWidth:   1920,
		MaxHeight:  1080,
		AutoResize: true,
	}

	return sma.AssertSnapshotCapture(cameraID, options)
}

// CreateMultipleTestSnapshots creates multiple test snapshots for list testing
// Eliminates multi-snapshot test setup duplication
func (sma *SnapshotManagerAsserter) CreateMultipleTestSnapshots(cameraID string, count int) []*TakeSnapshotResponse {
	var snapshots []*TakeSnapshotResponse

	for i := 0; i < count; i++ {
		snapshot := sma.CreateTestSnapshot(cameraID)
		snapshots = append(snapshots, snapshot)

		// Small delay between snapshots to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	sma.t.Logf("✅ Created %d test snapshots", count)
	return snapshots
}
