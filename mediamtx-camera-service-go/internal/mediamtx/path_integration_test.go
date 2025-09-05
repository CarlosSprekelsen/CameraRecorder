/*
MediaMTX Path Integration Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPathIntegration_ReqMTX001 tests path integration creation
func TestNewPathIntegration_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)
	assert.Equal(t, pathManager, pathIntegration.pathManager)
	assert.Equal(t, cameraMonitor, pathIntegration.cameraMonitor)
	assert.Equal(t, configManager, pathIntegration.configManager)
	assert.Equal(t, logger, pathIntegration.logger)
}

// TestPathIntegration_Start_ReqMTX001 tests integration startup
func TestPathIntegration_Start_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err, "Path integration should start successfully")

	// Verify integration is running (check if it can list paths)
	activePaths := pathIntegration.ListActivePaths()
	assert.NotNil(t, activePaths, "Integration should be able to list paths")

	// Stop integration
	err = pathIntegration.Stop(ctx)
	require.NoError(t, err, "Path integration should stop successfully")
}

// TestPathIntegration_Stop_ReqMTX001 tests integration shutdown
func TestPathIntegration_Stop_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err, "Path integration should start successfully")

	// Stop integration
	err = pathIntegration.Stop(ctx)
	require.NoError(t, err, "Path integration should stop successfully")

	// Verify integration is stopped (check if it can still list paths)
	activePaths := pathIntegration.ListActivePaths()
	assert.NotNil(t, activePaths, "Integration should still be able to list paths after stop")
}

// TestPathIntegration_CreatePathForCamera_ReqMTX003 tests path creation for camera
func TestPathIntegration_CreatePathForCamera_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err)
	defer pathIntegration.Stop(ctx)

	// Create path for camera
	devicePath := "/dev/video0"
	err = pathIntegration.CreatePathForCamera(ctx, devicePath)
	require.NoError(t, err, "Path creation for camera should succeed")

	// Verify path was created
	_, exists := pathIntegration.GetPathForCamera(devicePath)
	assert.True(t, exists, "Path should exist for camera")
}

// TestPathIntegration_DeletePathForCamera_ReqMTX003 tests path deletion for camera
func TestPathIntegration_DeletePathForCamera_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err)
	defer pathIntegration.Stop(ctx)

	// First, create a path for camera
	devicePath := "/dev/video0"
	err = pathIntegration.CreatePathForCamera(ctx, devicePath)
	require.NoError(t, err)

	// Verify path was created
	pathName, exists := pathIntegration.GetPathForCamera(devicePath)
	assert.True(t, exists, "Path should exist for camera")

	// Now delete the path for camera
	err = pathIntegration.DeletePathForCamera(ctx, devicePath)
	require.NoError(t, err, "Path deletion for camera should succeed")

	// Verify path was removed
	_, exists = pathIntegration.GetPathForCamera(devicePath)
	assert.False(t, exists, "Path should not exist after deletion")
}

// TestPathIntegration_ListActivePaths_ReqMTX002 tests active path listing
func TestPathIntegration_ListActivePaths_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err)
	defer pathIntegration.Stop(ctx)

	// Initially no active paths
	activePaths := pathIntegration.ListActivePaths()
	assert.Len(t, activePaths, 0, "Initially no active paths")

	// Create paths for multiple cameras
	err = pathIntegration.CreatePathForCamera(ctx, "/dev/video0")
	require.NoError(t, err)

	err = pathIntegration.CreatePathForCamera(ctx, "/dev/video1")
	require.NoError(t, err)

	// Verify both paths are active
	activePaths = pathIntegration.ListActivePaths()
	assert.Len(t, activePaths, 2, "Two paths should be active")
}

// TestPathIntegration_ErrorHandling_ReqMTX007 tests error scenarios
func TestPathIntegration_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies that fail
	pathManager := &mockPathManager{failCreate: true}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &mockConfigManager{}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err)
	defer pathIntegration.Stop(ctx)

	// Test path creation with failure
	err = pathIntegration.CreatePathForCamera(ctx, "/dev/video0")
	assert.Error(t, err, "Path creation should fail when path manager fails")

	// Test path deletion of non-existent camera
	err = pathIntegration.DeletePathForCamera(ctx, "/dev/nonexistent")
	assert.NoError(t, err, "Deleting path for non-existent camera should not fail")
}

// TestPathIntegration_ConcurrentAccess_ReqMTX001 tests concurrent operations
func TestPathIntegration_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create mock dependencies
	pathManager := &mockPathManager{}
	cameraMonitor := &mockCameraMonitor{}
	configManager := &config.ConfigManager{} // Use real ConfigManager
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configManager, logger)
	require.NotNil(t, pathIntegration)

	ctx := context.Background()

	// Start integration
	err := pathIntegration.Start(ctx)
	require.NoError(t, err)
	defer pathIntegration.Stop(ctx)

	// Create paths for multiple cameras concurrently
	const numCameras = 5
	errors := make([]error, numCameras)

	for i := 0; i < numCameras; i++ {
		go func(index int) {
			devicePath := fmt.Sprintf("/dev/video%d", index)
			err := pathIntegration.CreatePathForCamera(ctx, devicePath)
			errors[index] = err
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify all paths were created successfully
	activePaths := pathIntegration.ListActivePaths()
	assert.Len(t, activePaths, numCameras, "All concurrent path creations should succeed")
}

// Mock implementations for testing

type mockPathManager struct {
	failCreate bool
}

func (m *mockPathManager) CreatePath(ctx context.Context, name, source string, options map[string]interface{}) error {
	if m.failCreate {
		return fmt.Errorf("mock path creation failure")
	}
	return nil
}

func (m *mockPathManager) DeletePath(ctx context.Context, name string) error {
	return nil
}

func (m *mockPathManager) GetPath(ctx context.Context, name string) (*Path, error) {
	return &Path{Name: name}, nil
}

func (m *mockPathManager) ListPaths(ctx context.Context) ([]*Path, error) {
	return []*Path{}, nil
}

func (m *mockPathManager) ValidatePath(ctx context.Context, name string) error {
	return nil
}

func (m *mockPathManager) PathExists(ctx context.Context, name string) bool {
	return true
}

type mockCameraMonitor struct{}

func (m *mockCameraMonitor) Start(ctx context.Context) error {
	return nil
}

func (m *mockCameraMonitor) Stop() error {
	return nil
}

func (m *mockCameraMonitor) IsRunning() bool {
	return true
}

func (m *mockCameraMonitor) GetConnectedCameras() map[string]*camera.CameraDevice {
	return map[string]*camera.CameraDevice{}
}

func (m *mockCameraMonitor) GetDevice(devicePath string) (*camera.CameraDevice, bool) {
	return nil, false
}

func (m *mockCameraMonitor) GetMonitorStats() *camera.MonitorStats {
	return &camera.MonitorStats{}
}

func (m *mockCameraMonitor) AddEventHandler(handler camera.CameraEventHandler) {
}

func (m *mockCameraMonitor) AddEventCallback(callback func(camera.CameraEventData)) {
}

func (m *mockCameraMonitor) SetEventNotifier(notifier camera.EventNotifier) {
}
