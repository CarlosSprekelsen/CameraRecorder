package camera

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DeviceStatus represents the current status of a V4L2 device
type DeviceStatus string

const (
	DeviceStatusConnected    DeviceStatus = "CONNECTED"
	DeviceStatusDisconnected DeviceStatus = "DISCONNECTED"
	DeviceStatusError        DeviceStatus = "ERROR"
	DeviceStatusProbing      DeviceStatus = "PROBING"
)

// V4L2Capabilities represents the capabilities of a V4L2 device
type V4L2Capabilities struct {
	DriverName   string   `json:"driver_name"`
	CardName     string   `json:"card_name"`
	BusInfo      string   `json:"bus_info"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	DeviceCaps   []string `json:"device_caps"`
}

// V4L2Format represents a video format supported by a V4L2 device
type V4L2Format struct {
	PixelFormat string   `json:"pixel_format"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	FrameRates  []string `json:"frame_rates"`
}

// V4L2Device represents a V4L2 video device
type V4L2Device struct {
	Path         string           `json:"path"`
	Name         string           `json:"name"`
	Capabilities V4L2Capabilities `json:"capabilities"`
	Formats      []V4L2Format     `json:"formats"`
	Status       DeviceStatus     `json:"status"`
	LastSeen     time.Time        `json:"last_seen"`
	DeviceNum    int              `json:"device_num"`
	Error        string           `json:"error,omitempty"`
}

// V4L2DeviceManager manages V4L2 device discovery and monitoring
type V4L2DeviceManager struct {
	configProvider    ConfigProvider
	logger            Logger
	deviceChecker     DeviceChecker
	commandExecutor   V4L2CommandExecutor
	infoParser        DeviceInfoParser
	devices           map[string]*V4L2Device
	mu                sync.RWMutex
	stopChan          chan struct{}
	running           bool
	stats             *DeviceManagerStats
}

// CameraConfig represents camera discovery configuration
type CameraConfig struct {
	PollInterval              float64 `json:"poll_interval"`
	DetectionTimeout          float64 `json:"detection_timeout"`
	DeviceRange               []int   `json:"device_range"`
	EnableCapabilityDetection bool    `json:"enable_capability_detection"`
	CapabilityTimeout         float64 `json:"capability_timeout"`
	CapabilityRetryInterval   float64 `json:"capability_retry_interval"`
	CapabilityMaxRetries      int     `json:"capability_max_retries"`
}

// DeviceManagerStats tracks statistics for the device manager
type DeviceManagerStats struct {
	DevicesDiscovered          int     `json:"devices_discovered"`
	EventsProcessed            int     `json:"events_processed"`
	CapabilityProbes           int     `json:"capability_probes"`
	CapabilityProbesAttempted  int     `json:"capability_probes_attempted"`
	CapabilityProbesSuccessful int     `json:"capability_probes_successful"`
	CapabilityTimeouts         int     `json:"capability_timeouts"`
	CapabilityParseErrors      int     `json:"capability_parse_errors"`
	PollingCycles              int     `json:"polling_cycles"`
	CurrentPollInterval        float64 `json:"current_poll_interval"`
	Running                    bool    `json:"running"`
	ActiveTasks                int     `json:"active_tasks"`
	mu                         sync.RWMutex
}

// NewV4L2DeviceManager creates a new V4L2 device manager with dependency injection
func NewV4L2DeviceManager(configProvider ConfigProvider, logger Logger) *V4L2DeviceManager {
	if configProvider == nil {
		// Create default config provider
		configProvider = &DefaultConfigProvider{
			config: &CameraConfig{
				PollInterval:              0.1,
				DetectionTimeout:          1.0,
				DeviceRange:               []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				EnableCapabilityDetection: true,
				CapabilityTimeout:         5.0,
				CapabilityRetryInterval:   1.0,
				CapabilityMaxRetries:      3,
			},
		}
	}

	if logger == nil {
		logger = &DefaultLogger{}
	}

	return &V4L2DeviceManager{
		configProvider:  configProvider,
		logger:          logger,
		deviceChecker:   &RealDeviceChecker{},
		commandExecutor: &RealV4L2CommandExecutor{},
		infoParser:      &RealDeviceInfoParser{},
		devices:         make(map[string]*V4L2Device),
		stopChan:        make(chan struct{}),
		stats: &DeviceManagerStats{
			CurrentPollInterval: configProvider.GetPollInterval(),
		},
	}
}

// DefaultConfigProvider provides default configuration
type DefaultConfigProvider struct {
	config *CameraConfig
}

func (d *DefaultConfigProvider) GetCameraConfig() *CameraConfig {
	return d.config
}

func (d *DefaultConfigProvider) GetPollInterval() float64 {
	return d.config.PollInterval
}

func (d *DefaultConfigProvider) GetDetectionTimeout() float64 {
	return d.config.DetectionTimeout
}

func (d *DefaultConfigProvider) GetDeviceRange() []int {
	return d.config.DeviceRange
}

func (d *DefaultConfigProvider) GetEnableCapabilityDetection() bool {
	return d.config.EnableCapabilityDetection
}

func (d *DefaultConfigProvider) GetCapabilityTimeout() float64 {
	return d.config.CapabilityTimeout
}

// DefaultLogger provides default logging
type DefaultLogger struct{}

func (d *DefaultLogger) WithFields(fields map[string]interface{}) Logger {
	return d
}

func (d *DefaultLogger) Info(args ...interface{}) {}
func (d *DefaultLogger) Warn(args ...interface{}) {}
func (d *DefaultLogger) Error(args ...interface{}) {}
func (d *DefaultLogger) Debug(args ...interface{}) {}

// Start begins device discovery and monitoring
func (dm *V4L2DeviceManager) Start() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.running {
		return fmt.Errorf("device manager is already running")
	}

	dm.running = true
	dm.stats.Running = true
	dm.logger.Info("V4L2 device manager started")

	// Start polling loop
	go dm.pollingLoop()

	return nil
}

// Stop stops device discovery and monitoring
func (dm *V4L2DeviceManager) Stop() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if !dm.running {
		return fmt.Errorf("device manager is not running")
	}

	dm.running = false
	dm.stats.Running = false
	close(dm.stopChan)
	dm.logger.Info("V4L2 device manager stopped")

	return nil
}

// GetConnectedDevices returns all currently connected devices
func (dm *V4L2DeviceManager) GetConnectedDevices() map[string]*V4L2Device {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	connected := make(map[string]*V4L2Device)
	for path, device := range dm.devices {
		if device.Status == DeviceStatusConnected {
			connected[path] = device
		}
	}

	return connected
}

// GetDevice returns a specific device by path
func (dm *V4L2DeviceManager) GetDevice(path string) (*V4L2Device, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	device, exists := dm.devices[path]
	return device, exists
}

// GetStats returns current statistics
func (dm *V4L2DeviceManager) GetStats() *DeviceManagerStats {
	dm.stats.mu.RLock()
	defer dm.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *dm.stats
	return &stats
}

// pollingLoop continuously polls for device changes
func (dm *V4L2DeviceManager) pollingLoop() {
	ticker := time.NewTicker(time.Duration(dm.configProvider.GetPollInterval() * float64(time.Second)))
	defer ticker.Stop()

	dm.logger.WithFields(map[string]interface{}{
		"poll_interval": dm.configProvider.GetPollInterval(),
		"device_range":  dm.configProvider.GetDeviceRange(),
	}).Info("Starting V4L2 device polling loop")

	for {
		select {
		case <-dm.stopChan:
			dm.logger.Debug("Polling loop stopped")
			return
		case <-ticker.C:
			dm.discoverDevices()
		}
	}
}

// discoverDevices scans for currently connected devices
func (dm *V4L2DeviceManager) discoverDevices() {
	dm.stats.mu.Lock()
	dm.stats.PollingCycles++
	dm.stats.mu.Unlock()

	currentDevices := make(map[string]*V4L2Device)

	for _, deviceNum := range dm.configProvider.GetDeviceRange() {
		devicePath := fmt.Sprintf("/dev/video%d", deviceNum)

		device, err := dm.createDeviceInfo(devicePath, deviceNum)
		if err != nil {
			dm.logger.WithFields(logrus.Fields{
				"device_path": devicePath,
				"error":       err.Error(),
			}).Debug("Error checking device")
			continue
		}

		if device != nil && (device.Status == DeviceStatusConnected || device.Status == DeviceStatusError) {
			currentDevices[devicePath] = device
		}
	}

	dm.processDeviceStateChanges(currentDevices)
}

// createDeviceInfo creates device information for a given path
func (dm *V4L2DeviceManager) createDeviceInfo(devicePath string, deviceNum int) (*V4L2Device, error) {
	// Check if device file exists
	if !dm.deviceChecker.Exists(devicePath) {
		return nil, fmt.Errorf("device does not exist")
	}

	device := &V4L2Device{
		Path:      devicePath,
		DeviceNum: deviceNum,
		Status:    DeviceStatusProbing,
		LastSeen:  time.Now(),
	}

	// Probe device capabilities if enabled
	if dm.configProvider.GetEnableCapabilityDetection() {
		if err := dm.probeDeviceCapabilities(device); err != nil {
			device.Status = DeviceStatusError
			device.Error = err.Error()
			dm.logger.WithFields(logrus.Fields{
				"device_path": devicePath,
				"error":       err.Error(),
			}).Debug("Failed to probe device capabilities")
		} else {
			device.Status = DeviceStatusConnected
		}
	} else {
		device.Status = DeviceStatusConnected
		device.Name = fmt.Sprintf("Video Device %d", deviceNum)
	}

	return device, nil
}



// probeDeviceCapabilities probes device capabilities using v4l2-ctl
func (dm *V4L2DeviceManager) probeDeviceCapabilities(device *V4L2Device) error {
	dm.stats.mu.Lock()
	dm.stats.CapabilityProbesAttempted++
	dm.stats.mu.Unlock()

	// Execute v4l2-ctl --device /dev/videoX --info
	ctx := context.Background()
	infoOutput, err := dm.commandExecutor.ExecuteCommand(ctx, device.Path, "--info")
	if err != nil {
		dm.stats.mu.Lock()
		dm.stats.CapabilityTimeouts++
		dm.stats.mu.Unlock()
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Parse device info
	capabilities, err := dm.infoParser.ParseDeviceInfo(infoOutput)
	if err != nil {
		dm.stats.mu.Lock()
		dm.stats.CapabilityParseErrors++
		dm.stats.mu.Unlock()
		return fmt.Errorf("failed to parse device info: %w", err)
	}

	device.Capabilities = capabilities
	device.Name = device.Capabilities.CardName

	// Execute v4l2-ctl --device /dev/videoX --list-formats-ext
	formatsOutput, err := dm.commandExecutor.ExecuteCommand(ctx, device.Path, "--list-formats-ext")
	if err != nil {
		// Log warning but don't fail - device might not support format listing
		dm.logger.WithFields(map[string]interface{}{
			"device_path": device.Path,
			"error":       err.Error(),
		}).Warn("Failed to get device formats, using default")

		// Set default formats
		device.Formats = dm.getDefaultFormats()
	} else {
		// Parse formats
		formats, err := dm.infoParser.ParseDeviceFormats(formatsOutput)
		if err != nil {
			dm.logger.WithFields(map[string]interface{}{
				"device_path": device.Path,
				"error":       err.Error(),
			}).Warn("Failed to parse device formats, using default")
			device.Formats = dm.getDefaultFormats()
		} else {
			device.Formats = formats
		}
	}

	dm.stats.mu.Lock()
	dm.stats.CapabilityProbesSuccessful++
	dm.stats.mu.Unlock()

	return nil
}

// getDefaultFormats returns default formats when device probing fails
func (dm *V4L2DeviceManager) getDefaultFormats() []V4L2Format {
	return []V4L2Format{
		{
			PixelFormat: "YUYV",
			Width:       640,
			Height:      480,
			FrameRates:  []string{"30.000 fps", "25.000 fps"},
		},
		{
			PixelFormat: "MJPG",
			Width:       1280,
			Height:      720,
			FrameRates:  []string{"30.000 fps", "25.000 fps", "15.000 fps"},
		},
	}
}

// processDeviceStateChanges processes changes in device state
func (dm *V4L2DeviceManager) processDeviceStateChanges(currentDevices map[string]*V4L2Device) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Find new devices
	for path, device := range currentDevices {
		if existing, exists := dm.devices[path]; !exists {
			// New device
			dm.devices[path] = device
			dm.stats.mu.Lock()
			dm.stats.DevicesDiscovered++
			dm.stats.EventsProcessed++
			dm.stats.mu.Unlock()

			dm.logger.WithFields(map[string]interface{}{
				"device_path": path,
				"device_name": device.Name,
				"status":      device.Status,
			}).Info("New V4L2 device discovered")
		} else {
			// Update existing device
			existing.LastSeen = device.LastSeen
			if existing.Status != device.Status {
				existing.Status = device.Status
				existing.Error = device.Error

				dm.stats.mu.Lock()
				dm.stats.EventsProcessed++
				dm.stats.mu.Unlock()

				dm.logger.WithFields(map[string]interface{}{
					"device_path": path,
					"old_status":  existing.Status,
					"new_status":  device.Status,
				}).Info("V4L2 device status changed")
			}
		}
	}

	// Find removed devices
	for path, device := range dm.devices {
		if _, exists := currentDevices[path]; !exists {
			// Device removed
			device.Status = DeviceStatusDisconnected
			dm.stats.mu.Lock()
			dm.stats.EventsProcessed++
			dm.stats.mu.Unlock()

			dm.logger.WithFields(map[string]interface{}{
				"device_path": path,
				"device_name": device.Name,
			}).Info("V4L2 device disconnected")
		}
	}
}
