/*
Hybrid camera discovery monitor implementation.

Provides real-time USB camera detection using polling with capability detection,
following the Python HybridCameraMonitor patterns and project architecture standards.

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration
- REQ-CAM-002: Real-time device status monitoring
- REQ-CAM-003: Device capability probing and format detection

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// HybridCameraMonitor implements camera discovery and monitoring
// Following Python HybridCameraMonitor patterns with Go-specific optimizations
type HybridCameraMonitor struct {
	// Configuration
	deviceRange               []int
	pollInterval              float64
	detectionTimeout          float64
	enableCapabilityDetection bool

	// Dependencies (proper dependency injection)
	configManager   *config.ConfigManager
	logger          *logging.Logger
	deviceChecker   DeviceChecker
	commandExecutor V4L2CommandExecutor
	infoParser      DeviceInfoParser

	// State management
	knownDevices     map[string]*CameraDevice
	capabilityStates map[string]*DeviceCapabilityState
	running          bool
	stopChan         chan struct{}
	stateLock        sync.RWMutex

	// Caching
	capabilityCache map[string]*V4L2Capabilities
	cacheMutex      sync.RWMutex

	// Event handling
	eventHandlers     []CameraEventHandler
	eventCallbacks    []func(CameraEventData)
	eventHandlersLock sync.RWMutex

	// Statistics
	stats *MonitorStats

	// Adaptive polling configuration
	basePollInterval       float64
	currentPollInterval    float64
	minPollInterval        float64
	maxPollInterval        float64
	pollingFailureCount    int
	maxConsecutiveFailures int

	// Capability detection configuration
	capabilityTimeout       float64
	capabilityRetryInterval float64
	capabilityMaxRetries    int
}

// NewHybridCameraMonitor creates a new hybrid camera monitor with proper dependency injection
func NewHybridCameraMonitor(
	configManager *config.ConfigManager,
	logger *logging.Logger,
	deviceChecker DeviceChecker,
	commandExecutor V4L2CommandExecutor,
	infoParser DeviceInfoParser,
) *HybridCameraMonitor {
	if configManager == nil {
		panic("configManager cannot be nil - use existing internal/config/ConfigManager")
	}

	if logger == nil {
		logger = logging.NewLogger("hybrid-camera-monitor")
	}

	cfg := configManager.GetConfig()
	if cfg == nil {
		panic("configuration not available - ensure config is loaded")
	}

	monitor := &HybridCameraMonitor{
		// Configuration from config manager
		deviceRange:               cfg.Camera.DeviceRange,
		pollInterval:              cfg.Camera.PollInterval,
		detectionTimeout:          cfg.Camera.DetectionTimeout,
		enableCapabilityDetection: cfg.Camera.EnableCapabilityDetection,

		// Dependencies
		configManager:   configManager,
		logger:          logger,
		deviceChecker:   deviceChecker,
		commandExecutor: commandExecutor,
		infoParser:      infoParser,

		// State
		knownDevices:     make(map[string]*CameraDevice),
		capabilityStates: make(map[string]*DeviceCapabilityState),
		stopChan:         make(chan struct{}),

		// Caching
		capabilityCache: make(map[string]*V4L2Capabilities),

		// Event handling
		eventHandlers:  make([]CameraEventHandler, 0),
		eventCallbacks: make([]func(CameraEventData), 0),

		// Statistics
		stats: &MonitorStats{
			CurrentPollInterval: cfg.Camera.PollInterval,
		},

		// Adaptive polling
		basePollInterval:       cfg.Camera.PollInterval,
		currentPollInterval:    cfg.Camera.PollInterval,
		minPollInterval:        0.05, // 50ms minimum
		maxPollInterval:        5.0,  // 5s maximum
		maxConsecutiveFailures: 5,

		// Capability detection
		capabilityTimeout:       cfg.Camera.CapabilityTimeout,
		capabilityRetryInterval: cfg.Camera.CapabilityRetryInterval,
		capabilityMaxRetries:    cfg.Camera.CapabilityMaxRetries,
	}

	// Register for configuration hot-reload updates
	configManager.AddUpdateCallback(monitor.handleConfigurationUpdate)

	monitor.logger.WithFields(map[string]interface{}{
		"device_range":                monitor.deviceRange,
		"poll_interval":               monitor.pollInterval,
		"detection_timeout":           monitor.detectionTimeout,
		"enable_capability_detection": monitor.enableCapabilityDetection,
		"capability_timeout":          monitor.capabilityTimeout,
		"capability_max_retries":      monitor.capabilityMaxRetries,
		"action":                      "monitor_created",
	}).Info("Hybrid camera monitor created")

	return monitor
}

// handleConfigurationUpdate handles configuration hot-reload updates
func (m *HybridCameraMonitor) handleConfigurationUpdate(newConfig *config.Config) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	oldPollInterval := m.pollInterval
	oldDeviceRange := m.deviceRange
	oldCapabilityDetection := m.enableCapabilityDetection

	// Update configuration values
	m.deviceRange = newConfig.Camera.DeviceRange
	m.pollInterval = newConfig.Camera.PollInterval
	m.detectionTimeout = newConfig.Camera.DetectionTimeout
	m.enableCapabilityDetection = newConfig.Camera.EnableCapabilityDetection
	m.capabilityTimeout = newConfig.Camera.CapabilityTimeout
	m.capabilityRetryInterval = newConfig.Camera.CapabilityRetryInterval
	m.capabilityMaxRetries = newConfig.Camera.CapabilityMaxRetries

	// Update adaptive polling base values
	m.basePollInterval = newConfig.Camera.PollInterval
	if !m.running {
		// If not running, update current interval immediately
		m.currentPollInterval = newConfig.Camera.PollInterval
	}

	// Log configuration changes
	changes := make(map[string]interface{})
	if oldPollInterval != m.pollInterval {
		changes["poll_interval"] = map[string]float64{
			"old": oldPollInterval,
			"new": m.pollInterval,
		}
	}
	if !reflect.DeepEqual(oldDeviceRange, m.deviceRange) {
		changes["device_range"] = map[string]interface{}{
			"old": oldDeviceRange,
			"new": m.deviceRange,
		}
	}
	if oldCapabilityDetection != m.enableCapabilityDetection {
		changes["enable_capability_detection"] = map[string]bool{
			"old": oldCapabilityDetection,
			"new": m.enableCapabilityDetection,
		}
	}

	if len(changes) > 0 {
		m.logger.WithFields(map[string]interface{}{
			"changes": changes,
			"action":  "configuration_updated",
		}).Info("Camera monitor configuration updated via hot reload")
	}
}

// Start begins camera discovery and monitoring
func (m *HybridCameraMonitor) Start(ctx context.Context) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	m.running = true
	m.stats.Running = true
	m.stats.ActiveTasks++

	m.logger.WithFields(map[string]interface{}{
		"poll_interval": m.pollInterval,
		"device_range":  m.deviceRange,
		"action":        "monitor_started",
	}).Info("Starting hybrid camera monitor")

	// Start monitoring loop
	go m.monitoringLoop(ctx)

	return nil
}

// Stop stops camera discovery and monitoring
func (m *HybridCameraMonitor) Stop() error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	if !m.running {
		return fmt.Errorf("monitor is not running")
	}

	m.running = false
	m.stats.Running = false
	m.stats.ActiveTasks--

	// Only close the channel if it hasn't been closed already
	select {
	case <-m.stopChan:
		// Channel already closed, do nothing
	default:
		close(m.stopChan)
	}

	m.logger.WithFields(map[string]interface{}{
		"action": "monitor_stopped",
	}).Info("Hybrid camera monitor stopped")

	return nil
}

// IsRunning returns whether the monitor is currently running
func (m *HybridCameraMonitor) IsRunning() bool {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	return m.running
}

// GetConnectedCameras returns all currently connected cameras
func (m *HybridCameraMonitor) GetConnectedCameras() map[string]*CameraDevice {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()

	connected := make(map[string]*CameraDevice)
	for path, device := range m.knownDevices {
		if device.Status == DeviceStatusConnected {
			connected[path] = device
		}
	}

	return connected
}

// GetDevice returns a specific device by path
func (m *HybridCameraMonitor) GetDevice(devicePath string) (*CameraDevice, bool) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()

	device, exists := m.knownDevices[devicePath]
	return device, exists
}

// GetMonitorStats returns current monitoring statistics
func (m *HybridCameraMonitor) GetMonitorStats() *MonitorStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := MonitorStats{
		Running:                    m.stats.Running,
		ActiveTasks:                m.stats.ActiveTasks,
		PollingCycles:              m.stats.PollingCycles,
		DeviceStateChanges:         m.stats.DeviceStateChanges,
		CapabilityProbesAttempted:  m.stats.CapabilityProbesAttempted,
		CapabilityProbesSuccessful: m.stats.CapabilityProbesSuccessful,
		CapabilityTimeouts:         m.stats.CapabilityTimeouts,
		CapabilityParseErrors:      m.stats.CapabilityParseErrors,
		PollingFailureCount:        m.stats.PollingFailureCount,
		CurrentPollInterval:        m.stats.CurrentPollInterval,
		KnownDevicesCount:          len(m.knownDevices),
		UdevEventsProcessed:        m.stats.UdevEventsProcessed,
		UdevEventsFiltered:         m.stats.UdevEventsFiltered,
		UdevEventsSkipped:          m.stats.UdevEventsSkipped,
	}
	return &stats
}

// AddEventHandler adds a camera event handler
func (m *HybridCameraMonitor) AddEventHandler(handler CameraEventHandler) {
	m.eventHandlersLock.Lock()
	defer m.eventHandlersLock.Unlock()

	m.eventHandlers = append(m.eventHandlers, handler)
	m.logger.WithFields(map[string]interface{}{
		"handler_type": fmt.Sprintf("%T", handler),
		"action":       "event_handler_added",
	}).Debug("Added camera event handler")
}

// AddEventCallback adds a camera event callback function
func (m *HybridCameraMonitor) AddEventCallback(callback func(CameraEventData)) {
	m.eventHandlersLock.Lock()
	defer m.eventHandlersLock.Unlock()

	m.eventCallbacks = append(m.eventCallbacks, callback)
	m.logger.WithFields(map[string]interface{}{
		"action": "event_callback_added",
	}).Debug("Added camera event callback")
}

// monitoringLoop continuously monitors for device changes
func (m *HybridCameraMonitor) monitoringLoop(ctx context.Context) {
	defer func() {
		m.stats.mu.Lock()
		m.stats.ActiveTasks--
		m.stats.mu.Unlock()
	}()

	ticker := time.NewTicker(time.Duration(m.currentPollInterval * float64(time.Second)))
	defer ticker.Stop()

	m.logger.WithFields(map[string]interface{}{
		"poll_interval": m.currentPollInterval,
		"device_range":  m.deviceRange,
		"action":        "monitoring_loop_started",
	}).Info("Camera monitoring loop started")

	for {
		select {
		case <-ctx.Done():
			m.logger.WithFields(map[string]interface{}{
				"action": "monitoring_loop_stopped",
				"reason": "context_cancelled",
			}).Info("Camera monitoring loop stopped due to context cancellation")
			return
		case <-m.stopChan:
			m.logger.WithFields(map[string]interface{}{
				"action": "monitoring_loop_stopped",
				"reason": "stop_requested",
			}).Info("Camera monitoring loop stopped")
			return
		case <-ticker.C:
			m.discoverCameras()
			m.adjustPollingInterval()
		}
	}
}

// discoverCameras scans for currently connected cameras using parallel processing
func (m *HybridCameraMonitor) discoverCameras() {
	m.stats.mu.Lock()
	m.stats.PollingCycles++
	m.stats.mu.Unlock()

	currentDevices := make(map[string]*CameraDevice)
	var wg sync.WaitGroup
	deviceChan := make(chan *CameraDevice, len(m.deviceRange))

	// Start parallel device checking
	for _, deviceNum := range m.deviceRange {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			devicePath := fmt.Sprintf("/dev/video%d", num)

			device, err := m.createCameraDeviceInfo(devicePath, num)
			if err != nil {
				m.logger.WithFields(map[string]interface{}{
					"device_path": devicePath,
					"error":       err.Error(),
					"action":      "device_check_error",
				}).Debug("Error checking device")
				return
			}

			if device != nil && (device.Status == DeviceStatusConnected || device.Status == DeviceStatusError) {
				deviceChan <- device
			}
		}(deviceNum)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(deviceChan)
	}()

	// Collect results
	for device := range deviceChan {
		currentDevices[device.Path] = device
	}

	m.processDeviceStateChanges(currentDevices)
}

// createCameraDeviceInfo creates device information for a given path with lazy capability detection
func (m *HybridCameraMonitor) createCameraDeviceInfo(devicePath string, deviceNum int) (*CameraDevice, error) {
	// Check if device file exists
	if !m.deviceChecker.Exists(devicePath) {
		return nil, fmt.Errorf("device does not exist")
	}

	device := &CameraDevice{
		Path:      devicePath,
		DeviceNum: deviceNum,
		Status:    DeviceStatusConnected, // Assume connected if file exists
		LastSeen:  time.Now(),
	}

	// Check if we already have cached capabilities for this device
	m.stateLock.RLock()
	existingDevice, exists := m.knownDevices[devicePath]
	m.stateLock.RUnlock()

	if exists && existingDevice != nil {
		// Use cached capabilities if available
		device.Capabilities = existingDevice.Capabilities
		device.Formats = existingDevice.Formats
		device.Name = existingDevice.Name
	} else if m.enableCapabilityDetection {
		// Only probe capabilities on first detection (lazy detection)
		if err := m.probeDeviceCapabilities(device); err != nil {
			device.Status = DeviceStatusError
			device.Error = err.Error()
			m.logger.WithFields(map[string]interface{}{
				"device_path": devicePath,
				"error":       err.Error(),
				"action":      "capability_probe_failed",
			}).Debug("Failed to probe device capabilities")
		}
	} else {
		device.Name = fmt.Sprintf("Video Device %d", deviceNum)
	}

	return device, nil
}

// probeDeviceCapabilities probes device capabilities using v4l2-ctl with caching
func (m *HybridCameraMonitor) probeDeviceCapabilities(device *CameraDevice) error {
	// Check cache first
	m.cacheMutex.RLock()
	if cached, exists := m.capabilityCache[device.Path]; exists {
		device.Capabilities = *cached
		device.Name = cached.CardName
		m.cacheMutex.RUnlock()
		
		m.stats.mu.Lock()
		m.stats.CapabilityProbesSuccessful++
		m.stats.mu.Unlock()
		return nil
	}
	m.cacheMutex.RUnlock()

	m.stats.mu.Lock()
	m.stats.CapabilityProbesAttempted++
	m.stats.mu.Unlock()

	// Execute v4l2-ctl --device /dev/videoX --info
	ctx := context.Background()
	infoOutput, err := m.commandExecutor.ExecuteCommand(ctx, device.Path, "--info")
	if err != nil {
		m.stats.mu.Lock()
		m.stats.CapabilityTimeouts++
		m.stats.mu.Unlock()
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Parse device info
	capabilities, err := m.infoParser.ParseDeviceInfo(infoOutput)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.CapabilityParseErrors++
		m.stats.mu.Unlock()
		return fmt.Errorf("failed to parse device info: %w", err)
	}

	// Cache the capabilities
	m.cacheMutex.Lock()
	m.capabilityCache[device.Path] = &capabilities
	m.cacheMutex.Unlock()

	device.Capabilities = capabilities
	device.Name = device.Capabilities.CardName

	// Execute v4l2-ctl --device /dev/videoX --list-formats-ext
	formatsOutput, err := m.commandExecutor.ExecuteCommand(ctx, device.Path, "--list-formats-ext")
	if err != nil {
		// Log warning but don't fail - device might not support format listing
		m.logger.WithFields(map[string]interface{}{
			"device_path": device.Path,
			"error":       err.Error(),
			"action":      "format_listing_failed",
		}).Warn("Failed to get device formats, using default")

		// Set default formats
		device.Formats = m.getDefaultFormats()
	} else {
		// Parse formats
		formats, err := m.infoParser.ParseDeviceFormats(formatsOutput)
		if err != nil {
			m.logger.WithFields(map[string]interface{}{
				"device_path": device.Path,
				"error":       err.Error(),
				"action":      "format_parsing_failed",
			}).Warn("Failed to parse device formats, using default")
			device.Formats = m.getDefaultFormats()
		} else {
			device.Formats = formats
		}
	}

	m.stats.mu.Lock()
	m.stats.CapabilityProbesSuccessful++
	m.stats.mu.Unlock()

	return nil
}

// getDefaultFormats returns default formats when device probing fails
func (m *HybridCameraMonitor) getDefaultFormats() []V4L2Format {
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
func (m *HybridCameraMonitor) processDeviceStateChanges(currentDevices map[string]*CameraDevice) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	// Find new devices
	for path, device := range currentDevices {
		if existing, exists := m.knownDevices[path]; !exists {
			// New device
			m.knownDevices[path] = device
			m.stats.mu.Lock()
			m.stats.DeviceStateChanges++
			m.stats.mu.Unlock()

			m.logger.WithFields(map[string]interface{}{
				"device_path": path,
				"device_name": device.Name,
				"status":      device.Status,
				"action":      "device_discovered",
			}).Info("New V4L2 device discovered")

			// Generate event
			m.generateCameraEvent(CameraEventConnected, path, device)
		} else {
			// Update existing device
			existing.LastSeen = device.LastSeen
			if existing.Status != device.Status {
				existing.Status = device.Status
				existing.Error = device.Error

				m.stats.mu.Lock()
				m.stats.DeviceStateChanges++
				m.stats.mu.Unlock()

				m.logger.WithFields(map[string]interface{}{
					"device_path": path,
					"old_status":  existing.Status,
					"new_status":  device.Status,
					"action":      "device_status_changed",
				}).Info("V4L2 device status changed")

				// Generate event
				m.generateCameraEvent(CameraEventStatusChanged, path, device)
			}
		}
	}

	// Find removed devices
	for path, device := range m.knownDevices {
		if _, exists := currentDevices[path]; !exists {
			// Device removed
			device.Status = DeviceStatusDisconnected
			m.stats.mu.Lock()
			m.stats.DeviceStateChanges++
			m.stats.mu.Unlock()

			m.logger.WithFields(map[string]interface{}{
				"device_path": path,
				"device_name": device.Name,
				"action":      "device_disconnected",
			}).Info("V4L2 device disconnected")

			// Generate event
			m.generateCameraEvent(CameraEventDisconnected, path, device)
		}
	}
}

// generateCameraEvent generates camera events and notifies handlers
func (m *HybridCameraMonitor) generateCameraEvent(eventType CameraEvent, devicePath string, device *CameraDevice) {
	eventData := CameraEventData{
		DevicePath: devicePath,
		EventType:  eventType,
		Timestamp:  time.Now(),
		DeviceInfo: device,
	}

	// Notify event handlers
	m.eventHandlersLock.RLock()
	for _, handler := range m.eventHandlers {
		go func(h CameraEventHandler) {
			if err := h.HandleCameraEvent(context.Background(), eventData); err != nil {
				m.logger.WithFields(map[string]interface{}{
					"handler_type": fmt.Sprintf("%T", h),
					"error":        err.Error(),
					"action":       "event_handler_error",
				}).Error("Error in camera event handler")
			}
		}(handler)
	}
	m.eventHandlersLock.RUnlock()

	// Notify event callbacks
	m.eventHandlersLock.RLock()
	for _, callback := range m.eventCallbacks {
		go callback(eventData)
	}
	m.eventHandlersLock.RUnlock()
}

// adjustPollingInterval adjusts polling interval based on system responsiveness
func (m *HybridCameraMonitor) adjustPollingInterval() {
	oldInterval := m.currentPollInterval

	// Factor in recent failures - decrease interval (increase frequency) when there are failures
	if m.pollingFailureCount > 0 {
		// Apply failure penalty that increases polling frequency (decreases interval)
		failurePenalty := max(0.5, 1.0-float64(m.pollingFailureCount)*0.1) // Minimum 0.5x interval
		m.currentPollInterval = max(m.minPollInterval, m.currentPollInterval*failurePenalty)
	} else {
		// Gradually return to base interval when no failures
		m.currentPollInterval = min(m.maxPollInterval, m.currentPollInterval*1.1)
	}

	// Update stats if interval changed significantly
	if abs(m.currentPollInterval-oldInterval) > 0.01 {
		m.stats.mu.Lock()
		m.stats.CurrentPollInterval = m.currentPollInterval
		m.stats.mu.Unlock()

		m.logger.WithFields(map[string]interface{}{
			"old_interval":  oldInterval,
			"new_interval":  m.currentPollInterval,
			"failure_count": m.pollingFailureCount,
			"action":        "polling_interval_adjusted",
		}).Debug("Adjusted polling interval")
	}
}

// Helper functions
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
