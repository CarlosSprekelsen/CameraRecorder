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
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// PathIntegration provides integration between MediaMTX path management and camera discovery
type PathIntegration struct {
	pathManager   PathManager
	cameraMonitor camera.CameraMonitor
	configManager *config.ConfigManager
	logger        *logging.Logger

	// Path tracking
	activePaths   map[string]*Path
	activePathsMu sync.RWMutex

	// Camera-path mapping removed - delegated to PathManager
	// PathManager is the single source of truth for device->path mapping

	// Goroutine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewPathIntegration creates a new path integration
func NewPathIntegration(pathManager PathManager, cameraMonitor camera.CameraMonitor, configManager *config.ConfigManager, logger *logging.Logger) *PathIntegration {
	return &PathIntegration{
		pathManager:   pathManager,
		cameraMonitor: cameraMonitor,
		configManager: configManager,
		logger:        logger,
		activePaths:   make(map[string]*Path),
		// cameraPaths removed - delegated to PathManager
	}
}

// Start starts the path integration
func (pi *PathIntegration) Start(ctx context.Context) error {
	pi.logger.Info("Starting MediaMTX path integration")

	// Create cancellable context for background goroutine
	pi.ctx, pi.cancel = context.WithCancel(ctx)

	// Start monitoring camera changes with proper goroutine management
	pi.wg.Add(1)
	go pi.monitorCameraChanges(pi.ctx)

	// Skip initial path creation - paths will be created on-demand
	// This eliminates the race condition with camera discovery
	pi.logger.Info("Path integration configured for on-demand path creation")

	pi.logger.Info("MediaMTX path integration started successfully")
	return nil
}

// Stop stops the path integration
func (pi *PathIntegration) Stop(ctx context.Context) error {
	pi.logger.Info("Stopping MediaMTX path integration")

	// Cancel first
	if pi.cancel != nil {
		pi.cancel()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		pi.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		// Timeout - force cleanup
		pi.logger.Warn("Path integration shutdown timeout")
	}

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

	// Check if path already exists via PathManager (single source of truth)
	if existingPath, exists := pi.pathManager.GetCameraForDevicePath(device); exists {
		pi.logger.WithFields(logging.Fields{
			"device": device,
			"path":   existingPath,
		}).Debug("Path already exists for camera")
		return nil // Idempotent success - path already exists
	}

	// Get configuration
	cfg := pi.configManager.GetConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get configuration from config manager")
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

	// Cache update removed - PathManager handles mapping internally

	pi.logger.WithFields(logging.Fields{
		"device": device,
		"path":   pathName,
		"status": cameraDevice.Status,
	}).Info("MediaMTX path created for camera")

	return nil
}

// DeletePathForCamera deletes a MediaMTX path for a specific camera
func (pi *PathIntegration) DeletePathForCamera(ctx context.Context, device string) error {
	pi.logger.WithField("device", device).Debug("Deleting MediaMTX path for camera")

	// Get path name via PathManager (single source of truth)
	pathName, exists := pi.pathManager.GetCameraForDevicePath(device)
	if !exists {
		pi.logger.WithField("device", device).Debug("No path found for camera")
		return nil // Idempotent operation - no error for non-existent paths
	}

	// Delete the path
	if err := pi.pathManager.DeletePath(ctx, pathName); err != nil {
		return fmt.Errorf("failed to delete path %s for device %s: %w", pathName, device, err)
	}

	// Remove from tracking
	pi.activePathsMu.Lock()
	delete(pi.activePaths, pathName)
	pi.activePathsMu.Unlock()

	// Cache deletion removed - PathManager handles mapping internally

	pi.logger.WithFields(logging.Fields{
		"device": device,
		"path":   pathName,
	}).Info("MediaMTX path deleted for camera")

	return nil
}

// GetPathForCamera gets the MediaMTX path for a specific camera
func (pi *PathIntegration) GetPathForCamera(device string) (string, bool) {
	// DELEGATE TO PATHMANAGER - eliminate duplicate cache
	// PathManager is the single source of truth for device->path mapping
	return pi.pathManager.GetCameraForDevicePath(device)
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
	defer pi.wg.Done() // Ensure WaitGroup is decremented when goroutine exits

	// Use configurable ticker interval (default 5 seconds)
	tickerInterval := 5 * time.Second // Default fallback
	if pi.configManager != nil {
		cfg := pi.configManager.GetConfig()
		if cfg != nil && cfg.MediaMTX.HealthMonitorDefaults.CheckInterval > 0 {
			tickerInterval = time.Duration(cfg.MediaMTX.HealthMonitorDefaults.CheckInterval * float64(time.Second))
		}
	}
	ticker := time.NewTicker(tickerInterval)
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
	// Check context at start
	select {
	case <-ctx.Done():
		return
	default:
	}

	cameras := pi.cameraMonitor.GetConnectedCameras()

	for devicePath, camera := range cameras {
		// Check context in loop!
		select {
		case <-ctx.Done():
			pi.logger.Info("Camera changes handling cancelled")
			return
		default:
		}

		if camera.Status == "CONNECTED" {
			_, hasPath := pi.pathManager.GetCameraForDevicePath(devicePath)
			if !hasPath {
				// Use context-aware creation
				if err := pi.CreatePathForCamera(ctx, devicePath); err != nil {
					// Check if error is due to cancellation
					if ctx.Err() != nil {
						return
					}
					pi.logger.WithError(err).Error("Failed to create path")
				}
			}
		}
	}
}

// createPathsForExistingCameras creates paths for cameras that already exist
func (pi *PathIntegration) createPathsForExistingCameras(ctx context.Context) error {
	// Check context at start
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cameras := pi.cameraMonitor.GetConnectedCameras()

	for devicePath, camera := range cameras {
		// Check context in loop!
		select {
		case <-ctx.Done():
			pi.logger.Info("Creating paths for existing cameras cancelled")
			return ctx.Err()
		default:
		}

		if camera.Status == "CONNECTED" {
			if err := pi.CreatePathForCamera(ctx, devicePath); err != nil {
				// Check if error is due to cancellation
				if ctx.Err() != nil {
					return ctx.Err()
				}
				pi.logger.WithError(err).WithField("device", devicePath).Error("Failed to create path for existing camera")
			}
		}
	}

	return nil
}

// generatePathName generates a unique path name for a camera device
func (pi *PathIntegration) generatePathName(device string) string {
	// Convert device path to MediaMTX path name
	// e.g., /dev/video0 -> camera0, /dev/video1 -> camera1

	// Extract number from /dev/video{N}
	if strings.HasPrefix(device, "/dev/video") {
		number := strings.TrimPrefix(device, "/dev/video")
		return fmt.Sprintf("camera%s", number)
	}

	// For custom devices like /dev/custom_cam1
	if strings.HasPrefix(device, "/dev/") {
		deviceName := strings.TrimPrefix(device, "/dev/")
		// Replace non-alphanumeric with underscores for valid path names
		deviceName = strings.ReplaceAll(deviceName, "/", "_")
		return fmt.Sprintf("camera_%s", deviceName)
	}

	// Fallback: sanitize the device string
	sanitized := strings.ReplaceAll(device, "/", "_")
	return fmt.Sprintf("camera_%s", sanitized)
}

// GetPathStatus gets the status of a specific path
func (pi *PathIntegration) GetPathStatus(ctx context.Context, pathName string) (*Path, error) {
	return pi.pathManager.GetPath(ctx, pathName)
}

// ValidatePath validates a specific path
func (pi *PathIntegration) ValidatePath(ctx context.Context, pathName string) error {
	return pi.pathManager.ValidatePath(ctx, pathName)
}
