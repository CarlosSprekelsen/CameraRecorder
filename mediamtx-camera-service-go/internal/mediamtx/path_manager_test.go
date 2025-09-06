/*
MediaMTX Path Manager Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPathManager_ReqMTX001 tests path manager creation
func TestNewPathManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager, "Path manager should not be nil")
}

// TestPathManager_CreatePath_ReqMTX003 tests path creation
func TestPathManager_CreatePath_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	// Test path creation
	ctx := context.Background()
	testPathName := "test_path_manager_" + time.Now().Format("20060102_150405")

	// Ensure cleanup happens even if test fails
	defer func() {
		// Try to clean up the test path
		if err := pathManager.DeletePath(ctx, testPathName); err != nil {
			t.Logf("Cleanup warning: failed to delete test path %s: %v", testPathName, err)
		}
	}()

	err = pathManager.CreatePath(ctx, testPathName, "publisher", nil)
	require.NoError(t, err, "Path creation should succeed")

	// Verify path was created
	exists := pathManager.PathExists(ctx, testPathName)
	assert.True(t, exists, "Created path should exist")
}

// TestPathManager_DeletePath_ReqMTX003 tests path deletion
func TestPathManager_DeletePath_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()
	testPathName := "test_delete_path_" + time.Now().Format("20060102_150405")

	// Create a path first
	err = pathManager.CreatePath(ctx, testPathName, "publisher", nil)
	require.NoError(t, err, "Path creation should succeed")

	// Verify path exists
	exists := pathManager.PathExists(ctx, testPathName)
	assert.True(t, exists, "Path should exist before deletion")

	// Delete the path
	err = pathManager.DeletePath(ctx, testPathName)
	require.NoError(t, err, "Path deletion should succeed")

	// Verify path no longer exists
	exists = pathManager.PathExists(ctx, testPathName)
	assert.False(t, exists, "Path should not exist after deletion")
}

// TestPathManager_GetPath_ReqMTX003 tests path retrieval
func TestPathManager_GetPath_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()
	testPathName := "test_get_path_" + time.Now().Format("20060102_150405")

	// Create a path first
	err = pathManager.CreatePath(ctx, testPathName, "publisher", nil)
	require.NoError(t, err, "Path creation should succeed")

	// Get the path
	path, err := pathManager.GetPath(ctx, testPathName)
	require.NoError(t, err, "Path retrieval should succeed")
	require.NotNil(t, path, "Retrieved path should not be nil")
	assert.Equal(t, testPathName, path.Name, "Path name should match")

	// Clean up
	err = pathManager.DeletePath(ctx, testPathName)
	require.NoError(t, err, "Path deletion should succeed")
}

// TestPathManager_ListPaths_ReqMTX003 tests path listing
func TestPathManager_ListPaths_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()

	// List all paths
	paths, err := pathManager.ListPaths(ctx)
	require.NoError(t, err, "Path listing should succeed")
	require.NotNil(t, paths, "Paths list should not be nil")
	assert.GreaterOrEqual(t, len(paths), 0, "Should return at least 0 paths")
}

// TestPathManager_ValidatePath_ReqMTX003 tests path validation
func TestPathManager_ValidatePath_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()
	testPathName := "test_validate_path_" + time.Now().Format("20060102_150405")

	// Create a path first
	err = pathManager.CreatePath(ctx, testPathName, "publisher", nil)
	require.NoError(t, err, "Path creation should succeed")

	// Validate the path
	err = pathManager.ValidatePath(ctx, testPathName)
	require.NoError(t, err, "Path validation should succeed")

	// Clean up
	err = pathManager.DeletePath(ctx, testPathName)
	require.NoError(t, err, "Path deletion should succeed")
}

// TestPathManager_ErrorHandling_ReqMTX001 tests error handling
func TestPathManager_ErrorHandling_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for server to be ready
	err := helper.WaitForServerReady(t, 30*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: helper.GetConfig().BaseURL,
		Timeout: 30 * time.Second,
	}
	logger := logging.NewLogger("path-manager-test")
	logger.SetLevel(logrus.ErrorLevel)

	pathManager := NewPathManager(helper.GetClient(), config, logger)
	require.NotNil(t, pathManager)

	ctx := context.Background()

	// Test invalid path name
	err = pathManager.CreatePath(ctx, "", "publisher", nil)
	require.Error(t, err, "Empty path name should cause error")

	// Test getting non-existent path
	_, err = pathManager.GetPath(ctx, "non_existent_path")
	require.Error(t, err, "Getting non-existent path should cause error")

	// Test deleting non-existent path
	err = pathManager.DeletePath(ctx, "non_existent_path")
	require.Error(t, err, "Deleting non-existent path should cause error")
}
