package camera

import (
	"fmt"
	"sync"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/sirupsen/logrus"
)

// V4L2IntegrationManager manages the integration between V4L2 camera interface and configuration system
type V4L2IntegrationManager struct {
	deviceManager *V4L2DeviceManager
	configManager *config.ConfigManager
	logger        *logrus.Logger
	mu            sync.RWMutex
	running       bool
	stopChan      chan struct{}
}

// NewV4L2IntegrationManager creates a new V4L2 integration manager
func NewV4L2IntegrationManager(configManager *config.ConfigManager, logger *logrus.Logger) *V4L2IntegrationManager {
	if logger == nil {
		logger = logrus.New()
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

	// Create camera configuration
	cameraConfig := &CameraConfig{
		PollInterval:              cfg.Camera.PollInterval,
		DetectionTimeout:          cfg.Camera.DetectionTimeout,
		DeviceRange:               cfg.Camera.DeviceRange,
		EnableCapabilityDetection: cfg.Camera.EnableCapabilityDetection,
		CapabilityTimeout:         cfg.Camera.CapabilityTimeout,
		CapabilityRetryInterval:   cfg.Camera.CapabilityRetryInterval,
		CapabilityMaxRetries:      cfg.Camera.CapabilityMaxRetries,
	}

	// Create adapters for dependency injection
	configProvider := &IntegrationConfigProvider{config: cameraConfig}
	loggerAdapter := &LogrusLoggerAdapter{logger: im.logger}
	
	// Create device manager with configuration
	im.deviceManager = NewV4L2DeviceManager(configProvider, loggerAdapter)

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
			im.logger.WithError(err).Error("Error stopping device manager")
		}
	}

	im.running = false
	close(im.stopChan)
	im.logger.Info("V4L2 integration manager stopped")

	return nil
}

// GetConnectedDevices returns all currently connected devices
func (im *V4L2IntegrationManager) GetConnectedDevices() map[string]*V4L2Device {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return make(map[string]*V4L2Device)
	}

	return im.deviceManager.GetConnectedDevices()
}

// GetDevice returns a specific device by path
func (im *V4L2IntegrationManager) GetDevice(path string) (*V4L2Device, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return nil, false
	}

	return im.deviceManager.GetDevice(path)
}

// GetStats returns current statistics
func (im *V4L2IntegrationManager) GetStats() *DeviceManagerStats {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if im.deviceManager == nil {
		return &DeviceManagerStats{}
	}

	return im.deviceManager.GetStats()
}

// handleConfigUpdate handles configuration updates
func (im *V4L2IntegrationManager) handleConfigUpdate(newConfig *config.Config) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if !im.running || im.deviceManager == nil {
		return
	}

	im.logger.WithFields(logrus.Fields{
		"component": "v4l2_integration",
		"action":    "config_update",
	}).Info("Handling configuration update")

	// Note: In a real implementation, we would update the device manager configuration
	// For now, we'll just log the update
	im.logger.WithFields(logrus.Fields{
		"poll_interval":               newConfig.Camera.PollInterval,
		"detection_timeout":           newConfig.Camera.DetectionTimeout,
		"device_range":                newConfig.Camera.DeviceRange,
		"enable_capability_detection": newConfig.Camera.EnableCapabilityDetection,
	}).Debug("Configuration updated")
}

// ValidateConfiguration validates that the configuration is compatible with V4L2 operations
func (im *V4L2IntegrationManager) ValidateConfiguration(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	if cfg.Camera.PollInterval <= 0 {
		return fmt.Errorf("poll_interval must be greater than 0")
	}

	if cfg.Camera.DetectionTimeout <= 0 {
		return fmt.Errorf("detection_timeout must be greater than 0")
	}

	if len(cfg.Camera.DeviceRange) == 0 {
		return fmt.Errorf("device_range cannot be empty")
	}

	// Validate device range values
	for _, deviceNum := range cfg.Camera.DeviceRange {
		if deviceNum < 0 {
			return fmt.Errorf("device numbers must be non-negative")
		}
	}

	return nil
}

// IntegrationConfigProvider adapts CameraConfig to ConfigProvider interface
type IntegrationConfigProvider struct {
	config *CameraConfig
}

func (i *IntegrationConfigProvider) GetCameraConfig() *CameraConfig {
	return i.config
}

func (i *IntegrationConfigProvider) GetPollInterval() float64 {
	return i.config.PollInterval
}

func (i *IntegrationConfigProvider) GetDetectionTimeout() float64 {
	return i.config.DetectionTimeout
}

func (i *IntegrationConfigProvider) GetDeviceRange() []int {
	return i.config.DeviceRange
}

func (i *IntegrationConfigProvider) GetEnableCapabilityDetection() bool {
	return i.config.EnableCapabilityDetection
}

func (i *IntegrationConfigProvider) GetCapabilityTimeout() float64 {
	return i.config.CapabilityTimeout
}

// LogrusLoggerAdapter adapts logrus.Logger to Logger interface
type LogrusLoggerAdapter struct {
	logger *logrus.Logger
}

func (l *LogrusLoggerAdapter) WithFields(fields map[string]interface{}) Logger {
	logrusFields := logrus.Fields{}
	for k, v := range fields {
		logrusFields[k] = v
	}
	return &LogrusLoggerAdapter{logger: l.logger.WithFields(logrusFields).Logger}
}

func (l *LogrusLoggerAdapter) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLoggerAdapter) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLoggerAdapter) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLoggerAdapter) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}
