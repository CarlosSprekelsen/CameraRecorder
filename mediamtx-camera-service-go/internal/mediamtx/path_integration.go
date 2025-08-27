/*
MediaMTX Path Integration with Camera Discovery

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/sirupsen/logrus"
)

// PathIntegration provides integration between MediaMTX path management and camera discovery
type PathIntegration struct {
	pathManager   PathManager
	cameraMonitor camera.CameraMonitor
	configManager *config.ConfigManager
	logger        *logrus.Logger

	// Path tracking
	activePaths   map[string]*Path
	activePathsMu sync.RWMutex

	// Camera-path mapping
	cameraPaths   map[string]string // device -> path name
	cameraPathsMu sync.RWMutex
}

// NewPathIntegration creates a new path integration
func NewPathIntegration(pathManager PathManager, cameraMonitor camera.CameraMonitor, configManager *config.ConfigManager, logger *logrus.Logger) *PathIntegration {
	return &PathIntegration{
		pathManager:   pathManager,
		cameraMonitor: cameraMonitor,
		configManager: configManager,
		logger:        logger,
		activePaths:   make(map[string]*Path),
		cameraPaths:   make(map[string]string),
	}
}

// Start starts the path integration
func (pi *PathIntegration) Start(ctx context.Context) error {
	pi.logger.Info("Starting MediaMTX path integration")

	// Start monitoring camera changes
	go pi.monitorCameraChanges(ctx)

	// Create paths for existing cameras
	if err := pi.createPathsForExistingCameras(ctx); err != nil {
		pi.logger.WithError(err).Error("Failed to create paths for existing cameras")
	}

	pi.logger.Info("MediaMTX path integration started successfully")
	return nil
}

// Stop stops the path integration
func (pi *PathIntegration) Stop(ctx context.Context) error {
	pi.logger.Info("Stopping MediaMTX path integration")

	// Clean up all active paths
	pi.activePathsMu.Lock()
	for pathName := range pi.activePaths {
		if err := pi.pathManager.DeletePath(ctx, pathName); err != nil {
			pi.logger.WithError(err).WithField("path_name", pathName).Error("Failed to delete path during shutdown")
		}
	}
	pi.activePathsMu.Unlock()

	pi.logger.Info("MediaMTX path integration stopped successfully")
	return nil
}

// CreatePathForCamera creates a MediaMTX path for a specific camera
func (pi *PathIntegration) CreatePathForCamera(ctx context.Context, device string) error {
	pi.logger.WithField("device", device).Debug("Creating MediaMTX path for camera")

	// Get camera device
	cameraDevice, exists := pi.cameraMonitor.GetDevice(device)
	if !exists {
		return fmt.Errorf("camera device %s not found", device)
	}

	// Generate path name
	pathName := pi.generatePathName(device)

	// Check if path already exists
	pi.cameraPathsMu.Lock()
	if existingPath, exists := pi.cameraPaths[device]; exists {
		pi.cameraPathsMu.Unlock()
		pi.logger.WithFields(logrus.Fields{
			"device": device,
			"path":   existingPath,
		}).Debug("Path already exists for camera")
		return nil
	}
	pi.cameraPathsMu.Unlock()

	// Get configuration
	cfg := pi.configManager.GetConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get configuration")
	}

	// Create path options
	options := map[string]interface{}{
		"sourceOnDemand":             true,
		"sourceOnDemandStartTimeout": "10s",
		"sourceOnDemandCloseAfter":   "30s",
	}

	// Add camera-specific options
	if len(cameraDevice.Formats) > 0 {
		format := cameraDevice.Formats[0]
		options["resolution"] = fmt.Sprintf("%dx%d", format.Width, format.Height)
		if len(format.FrameRates) > 0 {
			options["fps"] = format.FrameRates[0]
		}
	}

	// Create the path
	if err := pi.pathManager.CreatePath(ctx, pathName, device, options); err != nil {
		return fmt.Errorf("failed to create path %s for device %s: %w", pathName, device, err)
	}

	// Track the path
	pi.activePathsMu.Lock()
	pi.activePaths[pathName] = &Path{
		Name:   pathName,
		Source: device,
	}
	pi.activePathsMu.Unlock()

	pi.cameraPathsMu.Lock()
	pi.cameraPaths[device] = pathName
	pi.cameraPathsMu.Unlock()

	pi.logger.WithFields(logrus.Fields{
		"device": device,
		"path":   pathName,
		"status": cameraDevice.Status,
	}).Info("MediaMTX path created for camera")

	return nil
}

// DeletePathForCamera deletes a MediaMTX path for a specific camera
func (pi *PathIntegration) DeletePathForCamera(ctx context.Context, device string) error {
	pi.logger.WithField("device", device).Debug("Deleting MediaMTX path for camera")

	// Get path name
	pi.cameraPathsMu.Lock()
	pathName, exists := pi.cameraPaths[device]
	if !exists {
		pi.cameraPathsMu.Unlock()
		pi.logger.WithField("device", device).Debug("No path found for camera")
		return nil
	}
	pi.cameraPathsMu.Unlock()

	// Delete the path
	if err := pi.pathManager.DeletePath(ctx, pathName); err != nil {
		return fmt.Errorf("failed to delete path %s for device %s: %w", pathName, device, err)
	}

	// Remove from tracking
	pi.activePathsMu.Lock()
	delete(pi.activePaths, pathName)
	pi.activePathsMu.Unlock()

	pi.cameraPathsMu.Lock()
	delete(pi.cameraPaths, device)
	pi.cameraPathsMu.Unlock()

	pi.logger.WithFields(logrus.Fields{
		"device": device,
		"path":   pathName,
	}).Info("MediaMTX path deleted for camera")

	return nil
}

// GetPathForCamera gets the MediaMTX path for a specific camera
func (pi *PathIntegration) GetPathForCamera(device string) (string, bool) {
	pi.cameraPathsMu.RLock()
	defer pi.cameraPathsMu.RUnlock()

	pathName, exists := pi.cameraPaths[device]
	return pathName, exists
}

// ListActivePaths lists all active MediaMTX paths
func (pi *PathIntegration) ListActivePaths() []*Path {
	pi.activePathsMu.RLock()
	defer pi.activePathsMu.RUnlock()

	paths := make([]*Path, 0, len(pi.activePaths))
	for _, path := range pi.activePaths {
		paths = append(paths, path)
	}

	return paths
}

// monitorCameraChanges monitors camera changes and updates paths accordingly
func (pi *PathIntegration) monitorCameraChanges(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pi.handleCameraChanges(ctx)
		}
	}
}

// handleCameraChanges handles camera status changes
func (pi *PathIntegration) handleCameraChanges(ctx context.Context) {
	// Get all cameras
	cameras := pi.cameraMonitor.GetConnectedCameras()

	// Track current cameras
	currentDevices := make(map[string]bool)
	for devicePath, camera := range cameras {
		currentDevices[devicePath] = true

		// Check if camera is connected and has no path
		if camera.Status == "CONNECTED" {
			pi.cameraPathsMu.RLock()
			_, hasPath := pi.cameraPaths[devicePath]
			pi.cameraPathsMu.RUnlock()

			if !hasPath {
				// Create path for new camera
				if err := pi.CreatePathForCamera(ctx, devicePath); err != nil {
					pi.logger.WithError(err).WithField("device", devicePath).Error("Failed to create path for new camera")
				}
			}
		}
	}

	// Check for disconnected cameras
	pi.cameraPathsMu.RLock()
	for device := range pi.cameraPaths {
		if !currentDevices[device] {
			// Camera disconnected, delete path
			if err := pi.DeletePathForCamera(ctx, device); err != nil {
				pi.logger.WithError(err).WithField("device", device).Error("Failed to delete path for disconnected camera")
			}
		}
	}
	pi.cameraPathsMu.RUnlock()
}

// createPathsForExistingCameras creates paths for cameras that already exist
func (pi *PathIntegration) createPathsForExistingCameras(ctx context.Context) error {
	cameras := pi.cameraMonitor.GetConnectedCameras()

	for devicePath, camera := range cameras {
		if camera.Status == "CONNECTED" {
			if err := pi.CreatePathForCamera(ctx, devicePath); err != nil {
				pi.logger.WithError(err).WithField("device", devicePath).Error("Failed to create path for existing camera")
			}
		}
	}

	return nil
}

// generatePathName generates a unique path name for a camera device
func (pi *PathIntegration) generatePathName(device string) string {
	// Convert device path to path name
	// e.g., /dev/video0 -> camera_video0
	pathName := fmt.Sprintf("camera_%s", device)

	// Remove leading slash and replace slashes with underscores
	if len(pathName) > 0 && pathName[0] == '/' {
		pathName = pathName[1:]
	}

	// Replace slashes with underscores
	for i := 0; i < len(pathName); i++ {
		if pathName[i] == '/' {
			pathName = pathName[:i] + "_" + pathName[i+1:]
		}
	}

	return pathName
}

// GetPathStatus gets the status of a specific path
func (pi *PathIntegration) GetPathStatus(ctx context.Context, pathName string) (*Path, error) {
	return pi.pathManager.GetPath(ctx, pathName)
}

// ValidatePath validates a specific path
func (pi *PathIntegration) ValidatePath(ctx context.Context, pathName string) error {
	return pi.pathManager.ValidatePath(ctx, pathName)
}
