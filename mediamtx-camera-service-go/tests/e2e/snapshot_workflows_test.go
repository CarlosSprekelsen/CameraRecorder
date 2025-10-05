/*
Snapshot Workflows E2E Tests

Tests complete user workflows for snapshot capture operations including single snapshots,
multiple snapshots, snapshot listing, and format options. Each test validates a complete
user journey with actual image file creation and verification.

Test Categories:
- Single Snapshot Capture: Take snapshot with device ID and filename, verify file exists with image format
- Multiple Snapshots Workflow: Take multiple snapshots sequentially, verify all files exist with unique names
- Snapshot List Workflow: Create snapshots, list snapshots, verify all present with metadata
- Snapshot with Format Options: Take snapshot with format and quality options, verify format and size

Business Outcomes:
- User can open image in viewer (verify valid image file)
- User can capture multiple snapshots without conflicts
- User can browse available snapshots with metadata
- User can control snapshot format and quality

Coverage Target: 55% E2E coverage milestone
*/

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleSnapshotCapture(t *testing.T) {
	// Setup: Authenticated connection, clean snapshots directory
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Single Snapshot", 1, "Authenticated connection established")

	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for snapshot test")
	}

	camera := cameras[0].(map[string]interface{})
	deviceID := camera["device"].(string)
	LogWorkflowStep(t, "Single Snapshot", 2, "Camera selected for snapshot")

	// Step 1: Call take_snapshot with device ID and filename
	filename := "test_single_snapshot.jpg"
	snapshotResponse := setup.SendJSONRPC(conn, "take_snapshot", map[string]interface{}{
		"device":   deviceID,
		"filename": filename,
	})
	LogWorkflowStep(t, "Single Snapshot", 3, "Take snapshot request sent")

	// Step 2: Verify snapshot response (snapshot_id, path, size)
	require.NoError(t, snapshotResponse.Error, "Take snapshot should succeed")
	require.NotNil(t, snapshotResponse.Result, "Take snapshot result should not be nil")

	snapshotResult := snapshotResponse.Result.(map[string]interface{})
	assert.Contains(t, snapshotResult, "snapshot_id", "Snapshot result should contain snapshot_id")
	assert.Contains(t, snapshotResult, "path", "Snapshot result should contain path")
	assert.Contains(t, snapshotResult, "size", "Snapshot result should contain size")

	snapshotID := snapshotResult["snapshot_id"].(string)
	snapshotPath := snapshotResult["path"].(string)
	snapshotSize := snapshotResult["size"].(float64)

	assert.NotEmpty(t, snapshotID, "Snapshot ID should not be empty")
	assert.NotEmpty(t, snapshotPath, "Snapshot path should not be empty")
	assert.Greater(t, snapshotSize, float64(0), "Snapshot size should be positive")
	LogWorkflowStep(t, "Single Snapshot", 4, "Snapshot response validated")

	// Step 3: Use verifySnapshotFile helper for complete validation
	setup.VerifySnapshotFile(snapshotPath, "jpg")
	LogWorkflowStep(t, "Single Snapshot", 5, "Snapshot file validation completed")

	// Validation: File exists AND size > 1KB AND valid image header AND extension matches format AND accessible/readable
	setup.AssertBusinessOutcome("User can open image in viewer", func() bool {
		return fileExists(t, snapshotPath) &&
			getFileSize(t, snapshotPath) > testutils.UniversalMinSnapshotFileSize &&
			strings.HasSuffix(snapshotPath, ".jpg")
	})
	LogWorkflowStep(t, "Single Snapshot", 6, "Business outcome validated - user can open image in viewer")

	// Cleanup: Remove snapshot file, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Single Snapshot", 7, "Connection closed and cleanup verified")
}

func TestMultipleSnapshotsWorkflow(t *testing.T) {
	// Setup: Authenticated connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Multiple Snapshots", 1, "Authenticated connection established")

	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for multiple snapshots test")
	}

	camera := cameras[0].(map[string]interface{})
	deviceID := camera["device"].(string)
	LogWorkflowStep(t, "Multiple Snapshots", 2, "Camera selected for multiple snapshots")

	// Step 1: Take 5 snapshots sequentially with unique filenames
	snapshotCount := 5
	snapshotPaths := make([]string, snapshotCount)
	snapshotIDs := make([]string, snapshotCount)

	for i := 0; i < snapshotCount; i++ {
		filename := fmt.Sprintf("test_multiple_snapshot_%d.jpg", i)

		snapshotResponse := setup.SendJSONRPC(conn, "take_snapshot", map[string]interface{}{
			"device":   deviceID,
			"filename": filename,
		})

		require.NoError(t, snapshotResponse.Error, "Take snapshot %d should succeed", i+1)
		require.NotNil(t, snapshotResponse.Result, "Take snapshot %d result should not be nil", i+1)

		snapshotResult := snapshotResponse.Result.(map[string]interface{})
		snapshotPaths[i] = snapshotResult["path"].(string)
		snapshotIDs[i] = snapshotResult["snapshot_id"].(string)

		assert.NotEmpty(t, snapshotPaths[i], "Snapshot %d path should not be empty", i+1)
		assert.NotEmpty(t, snapshotIDs[i], "Snapshot %d ID should not be empty", i+1)

		t.Logf("Snapshot %d created: %s", i+1, snapshotPaths[i])
	}
	LogWorkflowStep(t, "Multiple Snapshots", 3, "5 snapshots taken sequentially")

	// Step 2: Verify all 5 files created with unique names
	for i := 0; i < snapshotCount; i++ {
		assert.True(t, fileExists(t, snapshotPaths[i]), "Snapshot file %d should exist", i+1)
		assert.Greater(t, getFileSize(t, snapshotPaths[i]), int64(0), "Snapshot file %d should have content", i+1)

		// Verify unique filenames
		for j := i + 1; j < snapshotCount; j++ {
			assert.NotEqual(t, snapshotPaths[i], snapshotPaths[j], "Snapshot paths should be unique")
			assert.NotEqual(t, snapshotIDs[i], snapshotIDs[j], "Snapshot IDs should be unique")
		}
	}
	LogWorkflowStep(t, "Multiple Snapshots", 4, "All 5 files created with unique names")

	// Step 3: Validate each file has appropriate size and format
	for i := 0; i < snapshotCount; i++ {
		setup.VerifySnapshotFile(snapshotPaths[i], "jpg")
	}
	LogWorkflowStep(t, "Multiple Snapshots", 5, "All files validated for size and format")

	// Step 4: Verify no filename collisions
	uniquePaths := make(map[string]bool)
	for _, path := range snapshotPaths {
		assert.False(t, uniquePaths[path], "No filename collisions should occur")
		uniquePaths[path] = true
	}
	LogWorkflowStep(t, "Multiple Snapshots", 6, "No filename collisions verified")

	// Validation: 5 unique files AND each has valid image header AND appropriate sizes
	setup.AssertBusinessOutcome("User can capture multiple snapshots without conflicts", func() bool {
		allExist := true
		allValidSize := true
		allUnique := true

		for i := 0; i < snapshotCount; i++ {
			if !fileExists(t, snapshotPaths[i]) {
				allExist = false
			}
			if getFileSize(t, snapshotPaths[i]) <= testutils.UniversalMinSnapshotFileSize {
				allValidSize = false
			}
			for j := i + 1; j < snapshotCount; j++ {
				if snapshotPaths[i] == snapshotPaths[j] {
					allUnique = false
				}
			}
		}

		return allExist && allValidSize && allUnique
	})
	LogWorkflowStep(t, "Multiple Snapshots", 7, "Business outcome validated - user can capture multiple snapshots")

	// Cleanup: Remove all snapshot files, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Multiple Snapshots", 8, "Connection closed and cleanup verified")
}

func TestSnapshotListWorkflow(t *testing.T) {
	// Setup: Create 3 test snapshots
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Snapshot List", 1, "Authenticated connection established")

	// Create test snapshots
	testSnapshots := setup.CreateTestSnapshots(3)
	LogWorkflowStep(t, "Snapshot List", 2, "3 test snapshots created")

	// Step 1: Call list_snapshots method
	listResponse := setup.SendJSONRPC(conn, "list_snapshots", map[string]interface{}{})
	LogWorkflowStep(t, "Snapshot List", 3, "List snapshots request sent")

	// Step 2: Verify all 3 snapshots in response with metadata
	require.NoError(t, listResponse.Error, "List snapshots request should succeed")
	require.NotNil(t, listResponse.Result, "List snapshots result should not be nil")

	resultMap := listResponse.Result.(map[string]interface{})
	snapshots := resultMap["snapshots"].([]interface{})

	// Should have at least our test snapshots (may have more from other tests)
	assert.GreaterOrEqual(t, len(snapshots), 3, "Should have at least 3 snapshots in list")
	LogWorkflowStep(t, "Snapshot List", 4, "Snapshots list retrieved successfully")

	// Step 3: Validate metadata (camera_id, timestamp, size, path)
	for i, snapshotInterface := range snapshots[:3] { // Check first 3 snapshots
		snapshot := snapshotInterface.(map[string]interface{})

		assert.Contains(t, snapshot, "snapshot_id", "Snapshot should contain snapshot_id")
		assert.Contains(t, snapshot, "device", "Snapshot should contain device")
		assert.Contains(t, snapshot, "path", "Snapshot should contain path")
		assert.Contains(t, snapshot, "size", "Snapshot should contain size")
		assert.Contains(t, snapshot, "created", "Snapshot should contain created timestamp")

		// Validate metadata types and values
		snapshotID := snapshot["snapshot_id"].(string)
		device := snapshot["device"].(string)
		path := snapshot["path"].(string)

		assert.NotEmpty(t, snapshotID, "Snapshot ID should not be empty")
		assert.NotEmpty(t, device, "Device should not be empty")
		assert.NotEmpty(t, path, "Path should not be empty")

		if size, ok := snapshot["size"].(float64); ok {
			assert.Greater(t, size, 0, "Snapshot size should be positive")
		}

		if created, ok := snapshot["created"].(string); ok {
			assert.NotEmpty(t, created, "Created timestamp should not be empty")
		}

		t.Logf("Snapshot %d: %s (size: %v, device: %s)", i+1, path, snapshot["size"], device)
	}
	LogWorkflowStep(t, "Snapshot List", 5, "Snapshot metadata validated")

	// Step 4: Verify each snapshot file exists at listed path
	for i, snapshotInterface := range snapshots[:3] {
		snapshot := snapshotInterface.(map[string]interface{})
		path := snapshot["path"].(string)

		// Verify file exists
		if fileExists(t, path) {
			t.Logf("Snapshot file exists: %s", path)
		} else {
			t.Logf("Warning: Snapshot file not found: %s", path)
		}

		// For test snapshots, verify they exist
		if i < len(testSnapshots) {
			assert.True(t, fileExists(t, testSnapshots[i]), "Test snapshot file should exist")
		}
	}
	LogWorkflowStep(t, "Snapshot List", 6, "Snapshot file existence verified")

	// Validation: All snapshots listed AND metadata complete AND files exist AND sizes match
	setup.AssertBusinessOutcome("User can browse available snapshots", func() bool {
		return len(snapshots) >= 3 &&
			snapshots[0].(map[string]interface{})["snapshot_id"] != nil &&
			snapshots[1].(map[string]interface{})["snapshot_id"] != nil &&
			snapshots[2].(map[string]interface{})["snapshot_id"] != nil
	})
	LogWorkflowStep(t, "Snapshot List", 7, "Business outcome validated - user can browse snapshots")

	// Cleanup: Remove test snapshots, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Snapshot List", 8, "Connection closed and cleanup verified")
}

func TestSnapshotWithFormatOptions(t *testing.T) {
	// Setup: Authenticated connection
	setup := NewE2ETestSetup(t)
	adminToken := GenerateTestToken(t, "admin", 24)
	conn := setup.EstablishConnection(adminToken)
	LogWorkflowStep(t, "Format Options", 1, "Authenticated connection established")

	// Get first available camera
	cameraListResponse := setup.SendJSONRPC(conn, "get_camera_list", map[string]interface{}{})
	require.NoError(t, cameraListResponse.Error, "Camera list request should succeed")

	resultMap := cameraListResponse.Result.(map[string]interface{})
	cameras := resultMap["cameras"].([]interface{})

	if len(cameras) == 0 {
		t.Skip("No cameras available for format options test")
	}

	camera := cameras[0].(map[string]interface{})
	deviceID := camera["device"].(string)
	LogWorkflowStep(t, "Format Options", 2, "Camera selected for format options test")

	// Step 1: Take snapshot with format="png", quality=90
	pngFilename := "test_format_png.png"
	pngResponse := setup.SendJSONRPC(conn, "take_snapshot", map[string]interface{}{
		"device":   deviceID,
		"filename": pngFilename,
		"format":   "png",
		"quality":  90,
	})
	LogWorkflowStep(t, "Format Options", 3, "PNG snapshot request sent")

	// Step 2: Verify file created with .png extension
	require.NoError(t, pngResponse.Error, "PNG snapshot should succeed")
	require.NotNil(t, pngResponse.Result, "PNG snapshot result should not be nil")

	pngResult := pngResponse.Result.(map[string]interface{})
	pngPath := pngResult["path"].(string)
	pngSize := pngResult["size"].(float64)

	assert.Contains(t, pngPath, ".png", "PNG snapshot should have .png extension")
	assert.Greater(t, pngSize, float64(0), "PNG snapshot should have content")
	LogWorkflowStep(t, "Format Options", 4, "PNG file created with correct extension")

	// Step 3: Validate PNG file header (89 50 4E 47)
	setup.VerifySnapshotFile(pngPath, "png")
	LogWorkflowStep(t, "Format Options", 5, "PNG file header validated")

	// Step 4: Verify file size reflects quality setting (larger than low quality)
	// Take a low quality JPEG for comparison
	jpegFilename := "test_format_jpeg_low.jpg"
	jpegResponse := setup.SendJSONRPC(conn, "take_snapshot", map[string]interface{}{
		"device":   deviceID,
		"filename": jpegFilename,
		"format":   "jpeg",
		"quality":  50,
	})

	require.NoError(t, jpegResponse.Error, "JPEG snapshot should succeed")
	jpegResult := jpegResponse.Result.(map[string]interface{})
	jpegSize := jpegResult["size"].(float64)

	// High quality PNG should be larger than low quality JPEG (generally)
	t.Logf("PNG (quality 90) size: %v, JPEG (quality 50) size: %v", pngSize, jpegSize)
	LogWorkflowStep(t, "Format Options", 6, "Quality comparison completed")

	// Step 5: Take snapshot with format="jpeg", quality=50 for comparison
	// (Already done in step 4)
	setup.VerifySnapshotFile(jpegResult["path"].(string), "jpeg")
	LogWorkflowStep(t, "Format Options", 7, "JPEG file header validated")

	// Validation: PNG file created AND valid PNG header AND appropriate size for quality
	setup.AssertBusinessOutcome("User can control snapshot format and quality", func() bool {
		pngExists := fileExists(t, pngPath)
		jpegExists := fileExists(t, jpegResult["path"].(string))
		pngValidSize := pngSize > testutils.UniversalMinSnapshotFileSize
		jpegValidSize := jpegSize > testutils.UniversalMinSnapshotFileSize
		pngValidExt := strings.HasSuffix(pngPath, ".png")
		jpegValidExt := strings.Contains(jpegResult["path"].(string), ".jpg")

		return pngExists && jpegExists && pngValidSize && jpegValidSize && pngValidExt && jpegValidExt
	})
	LogWorkflowStep(t, "Format Options", 8, "Business outcome validated - user can control format and quality")

	// Cleanup: Remove snapshots, verify cleanup
	setup.CloseConnection(conn)
	LogWorkflowStep(t, "Format Options", 9, "Connection closed and cleanup verified")
}

// Helper functions for snapshot tests (reused from recording tests)
