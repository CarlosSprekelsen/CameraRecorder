package camera

import (
	"context"
	"fmt"
	"sync"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// V4L2IntegrationManager manages the integration between V4L2 camera interface and configuration system
type V4L2IntegrationManager struct {
	deviceManager *V4L2DeviceManager
	configManager *config.ConfigManager
	logger        *logging.Logger
	mu            sync.RWMutex
	running       bool
	stopChan      chan struct{}
}

// NewV4L2IntegrationManager creates a new V4L2 integration manager
func NewV4L2IntegrationManager(configManager *config.ConfigManager, logger *logging.Logger) *V4L2IntegrationManager {
	if logger == nil {
		logger = logging.NewLogger("v4l2-integration")
	}

	return &V4L2IntegrationManager{
		configManager: configManager,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the V4L2 integration with configuration management
func (im *V4L2IntegrationManager) Start() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.running {
		return fmt.Errorf("integration manager is already running")
	}

	// Get camera configuration from config manager
	cfg := im.configManager.GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not available")
	}

	// Create device manager with real configuration and logging
	im.deviceManager = NewV4L2DeviceManager(im.configManager, im.logger)

	// Start device manager
	if err := im.deviceManager.Start(); err != nil {
		return fmt.Errorf("failed to start device manager: %w", err)
	}

	// Add configuration update callback
	im.configManager.AddUpdateCallback(im.handleConfigUpdate)

	im.running = true
	im.logger.Info("V4L2 integration manager started")

	return nil
}

// Stop stops the V4L2 integration
func (im *V4L2IntegrationManager) Stop() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if !im.running {
		return fmt.Errorf("integration manager is not running")
	}

	// Stop device manager
	if im.deviceManager != nil {
		if err := im.deviceManager.Stop(); err != nil {
			im.logger.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("Error stopping device manager")
		}
	}

	im.running = false
	close(im.stopChan)
	im.logger.Info("V4L2 integration manager stopped")

	return nil
}

// GetDeviceManager returns the device manager instance
func (im *V4L2IntegrationManager) GetDeviceManager() *V4L2DeviceManager {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.deviceManager
}

// GetStats returns device manager statistics
func (im *V4L2IntegrationManager) GetStats() *DeviceManagerStats {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager != nil {
		return im.deviceManager.GetStats()
	}
	return &DeviceManagerStats{}
}

// IsRunning returns whether the integration manager is running
func (im *V4L2IntegrationManager) IsRunning() bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.running
}

// handleConfigUpdate handles configuration updates
func (im *V4L2IntegrationManager) handleConfigUpdate(newConfig *config.Config) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if !im.running {
		return
	}

	im.logger.WithFields(map[string]interface{}{
		"action": "config_update",
	}).Info("Configuration updated, restarting device manager")

	// Stop current device manager
	if im.deviceManager != nil {
		if err := im.deviceManager.Stop(); err != nil {
			im.logger.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("Error stopping device manager during config update")
		}
	}

	// Create new device manager with updated configuration
	im.deviceManager = NewV4L2DeviceManager(im.configManager, im.logger)

	// Start new device manager
	if err := im.deviceManager.Start(); err != nil {
		im.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Error starting device manager after config update")
	} else {
		im.logger.Info("Device manager restarted successfully with new configuration")
	}
}

// EnumerateDevices enumerates all V4L2 devices
func (im *V4L2IntegrationManager) EnumerateDevices(ctx context.Context) ([]*V4L2Device, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return nil, fmt.Errorf("device manager not available")
	}

	return im.deviceManager.EnumerateDevices(ctx)
}

// ProbeCapabilities probes capabilities for a specific device
func (im *V4L2IntegrationManager) ProbeCapabilities(ctx context.Context, devicePath string) (*V4L2Capabilities, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return nil, fmt.Errorf("device manager not available")
	}

	return im.deviceManager.ProbeCapabilities(ctx, devicePath)
}

// StartMonitoring starts device monitoring
func (im *V4L2IntegrationManager) StartMonitoring(ctx context.Context) error {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return fmt.Errorf("device manager not available")
	}

	return im.deviceManager.StartMonitoring(ctx)
}
