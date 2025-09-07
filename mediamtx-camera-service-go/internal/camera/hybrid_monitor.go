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
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// HybridCameraMonitor provides hybrid camera discovery and monitoring
// Enhanced to support USB cameras, IP cameras, RTSP cameras, and other camera types
type HybridCameraMonitor struct {
	// Configuration
	deviceRange               []int
	pollInterval              float64
	detectionTimeout          float64
	enableCapabilityDetection bool

	// Enhanced camera sources beyond USB
	cameraSources []CameraSource

	// Dependencies
	configManager   *config.ConfigManager
	logger          *logging.Logger
	deviceChecker   DeviceChecker
	commandExecutor V4L2CommandExecutor
	infoParser      DeviceInfoParser

	// State
	knownDevices     map[string]*CameraDevice
	capabilityStates map[string]*DeviceCapabilityState
	stopChan         chan struct{}

	// Caching
	capabilityCache map[string]*V4L2Capabilities

	// Event handling
	eventHandlers  []CameraEventHandler
	eventCallbacks []func(CameraEventData)

	// Statistics
	stats *MonitorStats

	// Adaptive polling
	basePollInterval       float64
	currentPollInterval    float64
	minPollInterval        float64
	maxPollInterval        float64
	maxConsecutiveFailures int

	// Capability detection
	capabilityTimeout       float64
	capabilityRetryInterval float64
	capabilityMaxRetries    int

	// Mutex for thread safety
	stateLock sync.RWMutex

	// Running state - using atomic operations
	running int32

	// Event handlers mutex
	eventHandlersLock sync.RWMutex

	// Cache mutex
	cacheMutex sync.RWMutex

	// Polling failure count - using atomic operations
	pollingFailureCount int64

	// Event system integration
	eventNotifier EventNotifier
}

// CameraSource represents a camera source configuration
type CameraSource struct {
	Type        string            `json:"type"`        // "usb", "ip", "rtsp", "http", "network", "file"
	Identifier  string            `json:"identifier"`  // camera0, ip_camera_192_168_1_100, etc.
	Source      string            `json:"source"`      // /dev/video0, rtsp://192.168.1.100:554/stream, etc.
	Enabled     bool              `json:"enabled"`     // Whether this source is enabled
	Options     map[string]string `json:"options"`     // Additional options (port, path, credentials, etc.)
	Description string            `json:"description"` // Human-readable description
}

// NewHybridCameraMonitor creates a new hybrid camera monitor with proper dependency injection
// Enhanced to support multiple camera types beyond USB cameras
func NewHybridCameraMonitor(
	configManager *config.ConfigManager,
	logger *logging.Logger,
	deviceChecker DeviceChecker,
	commandExecutor V4L2CommandExecutor,
	infoParser DeviceInfoParser,
) (*HybridCameraMonitor, error) {
	if configManager == nil {
		return nil, fmt.Errorf("configManager cannot be nil - use existing internal/config/ConfigManager")
	}

	if logger == nil {
		logger = logging.NewLogger("hybrid-camera-monitor")
	}

	cfg := configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not available - ensure config is loaded")
	}

	monitor := &HybridCameraMonitor{
		// Configuration from config manager
		deviceRange:               cfg.Camera.DeviceRange,
		pollInterval:              cfg.Camera.PollInterval,
		detectionTimeout:          cfg.Camera.DetectionTimeout,
		enableCapabilityDetection: cfg.Camera.EnableCapabilityDetection,

		// Enhanced camera sources
		cameraSources: []CameraSource{},

		// Dependencies
		configManager:   configManager,
		logger:          logger,
		deviceChecker:   deviceChecker,
		commandExecutor: commandExecutor,
		infoParser:      infoParser,

		// State
		knownDevices:     make(map[string]*CameraDevice),
		capabilityStates: make(map[string]*DeviceCapabilityState),
		stopChan:         make(chan struct{}, 10), // Buffered to prevent deadlock during shutdown

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

	// Initialize camera sources from configuration
	monitor.initializeCameraSources()

	// Register for configuration hot-reload updates
	monitor.configManager.AddUpdateCallback(monitor.handleConfigurationUpdate)

	return monitor, nil
}

// initializeCameraSources initializes camera sources from configuration
func (m *HybridCameraMonitor) initializeCameraSources() {
	// Add USB camera sources
	for _, deviceNum := range m.deviceRange {
		m.cameraSources = append(m.cameraSources, CameraSource{
			Type:        "usb",
			Identifier:  "camera" + strconv.Itoa(deviceNum),
			Source:      "/dev/video" + strconv.Itoa(deviceNum),
			Enabled:     true,
			Description: "USB Camera " + strconv.Itoa(deviceNum),
		})
	}

	// Add IP camera sources from configuration if available
	// This can be extended to read from config file or environment variables
	m.addIPCameraSources()
}

// addIPCameraSources adds IP camera sources from configuration
func (m *HybridCameraMonitor) addIPCameraSources() {
	// Example IP camera sources - these should come from configuration
	// In a real implementation, this would read from config file or environment variables

	// Example RTSP camera
	m.cameraSources = append(m.cameraSources, CameraSource{
		Type:        "rtsp",
		Identifier:  "ip_camera_192_168_1_100",
		Source:      "rtsp://192.168.1.100:554/stream",
		Enabled:     true,
		Options:     map[string]string{"port": "554", "path": "/stream"},
		Description: "IP Camera 192.168.1.100",
	})

	// Example HTTP camera
	m.cameraSources = append(m.cameraSources, CameraSource{
		Type:        "http",
		Identifier:  "http_camera_192_168_1_101",
		Source:      "http://192.168.1.101:8080/mjpeg",
		Enabled:     true,
		Options:     map[string]string{"port": "8080", "path": "/mjpeg"},
		Description: "HTTP Camera 192.168.1.101",
	})

	// Example network camera
	m.cameraSources = append(m.cameraSources, CameraSource{
		Type:        "network",
		Identifier:  "network_camera_239_0_0_1_1234",
		Source:      "udp://239.0.0.1:1234",
		Enabled:     true,
		Options:     map[string]string{"protocol": "udp", "multicast": "true"},
		Description: "Network Camera 239.0.0.1:1234",
	})
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
	if atomic.LoadInt32(&m.running) == 0 {
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
		m.logger.WithFields(logging.Fields{
			"changes": changes,
			"action":  "configuration_updated",
		}).Info("Camera monitor configuration updated via hot reload")
	}
}

// Start begins camera discovery and monitoring
func (m *HybridCameraMonitor) Start(ctx context.Context) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return fmt.Errorf("monitor is already running")
	}

	m.stats.Running = true
	atomic.AddInt64(&m.stats.ActiveTasks, 1)

	m.logger.WithFields(logging.Fields{
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

	if atomic.LoadInt32(&m.running) == 0 {
		return fmt.Errorf("monitor is not running")
	}

	atomic.StoreInt32(&m.running, 0)
	m.stats.Running = false
	atomic.AddInt64(&m.stats.ActiveTasks, -1)

	// Only close the channel if it hasn't been closed already
	select {
	case <-m.stopChan:
		// Channel already closed, do nothing
	default:
		close(m.stopChan)
	}

	m.logger.WithFields(logging.Fields{
		"action": "monitor_stopped",
	}).Info("Hybrid camera monitor stopped")

	return nil
}

// IsRunning returns whether the monitor is currently running
func (m *HybridCameraMonitor) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
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
	// Create a copy using atomic operations to avoid race conditions
	stats := MonitorStats{
		Running:                    m.stats.Running,
		ActiveTasks:                atomic.LoadInt64(&m.stats.ActiveTasks),
		PollingCycles:              atomic.LoadInt64(&m.stats.PollingCycles),
		DeviceStateChanges:         atomic.LoadInt64(&m.stats.DeviceStateChanges),
		CapabilityProbesAttempted:  atomic.LoadInt64(&m.stats.CapabilityProbesAttempted),
		CapabilityProbesSuccessful: atomic.LoadInt64(&m.stats.CapabilityProbesSuccessful),
		CapabilityTimeouts:         atomic.LoadInt64(&m.stats.CapabilityTimeouts),
		CapabilityParseErrors:      atomic.LoadInt64(&m.stats.CapabilityParseErrors),
		PollingFailureCount:        atomic.LoadInt64(&m.stats.PollingFailureCount),
		CurrentPollInterval:        m.stats.CurrentPollInterval,
		KnownDevicesCount:          int64(len(m.knownDevices)),
		UdevEventsProcessed:        atomic.LoadInt64(&m.stats.UdevEventsProcessed),
		UdevEventsFiltered:         atomic.LoadInt64(&m.stats.UdevEventsFiltered),
		UdevEventsSkipped:          atomic.LoadInt64(&m.stats.UdevEventsSkipped),
	}
	return &stats
}

// AddEventHandler adds a camera event handler
func (m *HybridCameraMonitor) AddEventHandler(handler CameraEventHandler) {
	m.eventHandlersLock.Lock()
	defer m.eventHandlersLock.Unlock()

	m.eventHandlers = append(m.eventHandlers, handler)
	m.logger.WithFields(logging.Fields{
		"handler_type": fmt.Sprintf("%T", handler), // Keep fmt.Sprintf for type reflection
		"action":       "event_handler_added",
	}).Debug("Added camera event handler")
}

// AddEventCallback adds a camera event callback function
func (m *HybridCameraMonitor) AddEventCallback(callback func(CameraEventData)) {
	m.eventHandlersLock.Lock()
	defer m.eventHandlersLock.Unlock()

	m.eventCallbacks = append(m.eventCallbacks, callback)
	m.logger.WithFields(logging.Fields{
		"action": "event_callback_added",
	}).Debug("Added camera event callback")
}

// SetEventNotifier sets the event notifier for external event system integration
func (m *HybridCameraMonitor) SetEventNotifier(notifier EventNotifier) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	m.eventNotifier = notifier
}

// monitoringLoop continuously monitors for device changes
func (m *HybridCameraMonitor) monitoringLoop(ctx context.Context) {
	defer func() {
		atomic.AddInt64(&m.stats.ActiveTasks, -1)
	}()

	ticker := time.NewTicker(time.Duration(m.currentPollInterval * float64(time.Second)))
	defer ticker.Stop()

	m.logger.WithFields(logging.Fields{
		"poll_interval": m.currentPollInterval,
		"device_range":  m.deviceRange,
		"action":        "monitoring_loop_started",
	}).Info("Camera monitoring loop started")

	for {
		select {
		case <-ctx.Done():
			m.logger.WithFields(logging.Fields{
				"action": "monitoring_loop_stopped",
				"reason": "context_cancelled",
			}).Info("Camera monitoring loop stopped due to context cancellation")
			return
		case <-m.stopChan:
			m.logger.WithFields(logging.Fields{
				"action": "monitoring_loop_stopped",
				"reason": "stop_requested",
			}).Info("Camera monitoring loop stopped")
			return
		case <-ticker.C:
			m.discoverCameras(ctx)
			m.adjustPollingInterval()
		}
	}
}

// discoverCameras scans for currently connected cameras using parallel processing
// Enhanced to support multiple camera types beyond USB cameras
func (m *HybridCameraMonitor) discoverCameras(ctx context.Context) {
	atomic.AddInt64(&m.stats.PollingCycles, 1)

	currentDevices := make(map[string]*CameraDevice)
	var wg sync.WaitGroup
	deviceChan := make(chan *CameraDevice, len(m.cameraSources))
	errorChan := make(chan error, len(m.cameraSources))

	// Start parallel device checking for all camera sources
	for _, source := range m.cameraSources {
		if !source.Enabled {
			continue
		}

		wg.Add(1)
		go func(src CameraSource) {
			defer func() {
				// Recover from panics in goroutine and propagate as errors
				if r := recover(); r != nil {
					panicErr := fmt.Errorf("panic in device check goroutine for source %s: %v", src.Identifier, r)
					m.logger.WithFields(logging.Fields{
						"source": src.Identifier,
						"panic":  r,
						"action": "panic_recovered",
					}).Error("Recovered from panic in device check goroutine")

					// Propagate panic as error instead of swallowing it
					select {
					case errorChan <- panicErr:
					default:
						// If error channel is full, log the overflow
						m.logger.WithError(panicErr).Warn("Error channel overflow, panic error dropped")
					}
				}
				wg.Done()
			}()

			device, err := m.createCameraDeviceInfoFromSource(ctx, src)
			if err != nil {
				m.logger.WithFields(logging.Fields{
					"source": src.Identifier,
					"error":  err.Error(),
					"action": "device_check_error",
				}).Debug("Error checking device")

				// Propagate device check errors
				select {
				case errorChan <- fmt.Errorf("device check failed for source %s: %w", src.Identifier, err):
				default:
					m.logger.WithError(err).Warn("Error channel overflow, device check error dropped")
				}
				return
			}

			if device != nil && (device.Status == DeviceStatusConnected || device.Status == DeviceStatusError) {
				deviceChan <- device
			}
		}(source)
	}

	// Wait for all goroutines to complete
	go func() {
		defer func() {
			// Recover from panics in goroutine and propagate as errors
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("panic in device collection goroutine: %v", r)
				m.logger.WithFields(logging.Fields{
					"panic":  r,
					"action": "panic_recovered",
				}).Error("Recovered from panic in device collection goroutine")

				// Propagate panic as error
				select {
				case errorChan <- panicErr:
				default:
					m.logger.WithError(panicErr).Warn("Error channel overflow, panic error dropped")
				}
			}
		}()
		wg.Wait()
		close(deviceChan)
		close(errorChan)
	}()

	// Collect results and errors
	for device := range deviceChan {
		currentDevices[device.Path] = device
	}

	// Process any errors that occurred during device checking
	for err := range errorChan {
		m.logger.WithError(err).Warn("Device check error occurred")
		// Optionally increment error counters or trigger recovery mechanisms
		atomic.AddInt64(&m.pollingFailureCount, 1)
	}

	m.processDeviceStateChanges(ctx, currentDevices)
}

// createCameraDeviceInfoFromSource creates device information for a given camera source
func (m *HybridCameraMonitor) createCameraDeviceInfoFromSource(ctx context.Context, source CameraSource) (*CameraDevice, error) {
	switch source.Type {
	case "usb":
		return m.createUSBCameraDeviceInfo(ctx, source)
	case "rtsp", "http", "network":
		return m.createNetworkCameraDeviceInfo(ctx, source)
	case "file":
		return m.createFileCameraDeviceInfo(ctx, source)
	default:
		return m.createGenericCameraDeviceInfo(ctx, source)
	}
}

// createUSBCameraDeviceInfo creates device information for USB cameras
func (m *HybridCameraMonitor) createUSBCameraDeviceInfo(ctx context.Context, source CameraSource) (*CameraDevice, error) {
	// Extract device number from source path
	var deviceNum int
	_, err := fmt.Sscanf(source.Source, "/dev/video%d", &deviceNum)
	if err != nil {
		return nil, fmt.Errorf("invalid USB device path: %s", source.Source)
	}

	return m.createCameraDeviceInfo(ctx, source.Source, deviceNum)
}

// createCameraDeviceInfo creates device information for a given path with lazy capability detection
func (m *HybridCameraMonitor) createCameraDeviceInfo(ctx context.Context, devicePath string, deviceNum int) (*CameraDevice, error) {
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
		if err := m.probeDeviceCapabilities(ctx, device); err != nil {
			device.Status = DeviceStatusError
			device.Error = err.Error()
			m.logger.WithFields(logging.Fields{
				"device_path": devicePath,
				"error":       err.Error(),
				"action":      "capability_probe_failed",
			}).Debug("Failed to probe device capabilities")
		}
	} else {
		device.Name = "Video Device " + strconv.Itoa(deviceNum)
	}

	return device, nil
}

// createNetworkCameraDeviceInfo creates device information for network cameras
func (m *HybridCameraMonitor) createNetworkCameraDeviceInfo(ctx context.Context, source CameraSource) (*CameraDevice, error) {
	// For network cameras, we assume they're connected if we can reach them
	// In a real implementation, you might want to test connectivity

	device := &CameraDevice{
		Path:   source.Source,
		Name:   source.Description,
		Status: DeviceStatusConnected, // Assume connected for network cameras
		Capabilities: V4L2Capabilities{
			DriverName: "network_camera",
			CardName:   source.Description,
			BusInfo:    source.Source,
		},
		Formats: []V4L2Format{
			{
				Width:       1920,
				Height:      1080,
				PixelFormat: "YUYV",
				FrameRates:  []string{"30", "25", "15"},
			},
		},
		LastSeen: time.Now(),
	}

	return device, nil
}

// createFileCameraDeviceInfo creates device information for file-based cameras
func (m *HybridCameraMonitor) createFileCameraDeviceInfo(ctx context.Context, source CameraSource) (*CameraDevice, error) {
	// Check if file exists
	if !m.deviceChecker.Exists(source.Source) {
		return &CameraDevice{
			Path:   source.Source,
			Name:   source.Description,
			Status: DeviceStatusDisconnected,
		}, nil
	}

	device := &CameraDevice{
		Path:   source.Source,
		Name:   source.Description,
		Status: DeviceStatusConnected,
		Capabilities: V4L2Capabilities{
			DriverName: "file_source",
			CardName:   source.Description,
			BusInfo:    source.Source,
		},
		Formats: []V4L2Format{
			{
				Width:       1920,
				Height:      1080,
				PixelFormat: "H264",
				FrameRates:  []string{"30", "25", "15"},
			},
		},
		LastSeen: time.Now(),
	}

	return device, nil
}

// createGenericCameraDeviceInfo creates device information for generic camera types
func (m *HybridCameraMonitor) createGenericCameraDeviceInfo(ctx context.Context, source CameraSource) (*CameraDevice, error) {
	device := &CameraDevice{
		Path:   source.Source,
		Name:   source.Description,
		Status: DeviceStatusConnected, // Assume connected for generic types
		Capabilities: V4L2Capabilities{
			DriverName: "generic_camera",
			CardName:   source.Description,
			BusInfo:    source.Source,
		},
		Formats: []V4L2Format{
			{
				Width:       1920,
				Height:      1080,
				PixelFormat: "YUYV",
				FrameRates:  []string{"30", "25", "15"},
			},
		},
		LastSeen: time.Now(),
	}

	return device, nil
}

// probeDeviceCapabilities probes device capabilities using v4l2-ctl with caching
func (m *HybridCameraMonitor) probeDeviceCapabilities(ctx context.Context, device *CameraDevice) error {
	// Check cache first
	m.cacheMutex.RLock()
	if cached, exists := m.capabilityCache[device.Path]; exists {
		device.Capabilities = *cached
		device.Name = cached.CardName
		m.cacheMutex.RUnlock()

		atomic.AddInt64(&m.stats.CapabilityProbesSuccessful, 1)
		return nil
	}
	m.cacheMutex.RUnlock()

	atomic.AddInt64(&m.stats.CapabilityProbesAttempted, 1)

	// Execute v4l2-ctl --device /dev/videoX --info
	infoOutput, err := m.commandExecutor.ExecuteCommand(ctx, device.Path, "--info")
	if err != nil {
		atomic.AddInt64(&m.stats.CapabilityTimeouts, 1)
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Parse device info
	capabilities, err := m.infoParser.ParseDeviceInfo(infoOutput)
	if err != nil {
		atomic.AddInt64(&m.stats.CapabilityParseErrors, 1)
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
		m.logger.WithFields(logging.Fields{
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
			m.logger.WithFields(logging.Fields{
				"device_path": device.Path,
				"error":       err.Error(),
				"action":      "format_parsing_failed",
			}).Warn("Failed to parse device formats, using default")
			device.Formats = m.getDefaultFormats()
		} else {
			device.Formats = formats
		}
	}

	atomic.AddInt64(&m.stats.CapabilityProbesSuccessful, 1)

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
func (m *HybridCameraMonitor) processDeviceStateChanges(ctx context.Context, currentDevices map[string]*CameraDevice) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	// Find new devices
	for path, device := range currentDevices {
		if existing, exists := m.knownDevices[path]; !exists {
			// New device
			m.knownDevices[path] = device
			atomic.AddInt64(&m.stats.DeviceStateChanges, 1)

			m.logger.WithFields(logging.Fields{
				"device_path": path,
				"device_name": device.Name,
				"status":      device.Status,
				"action":      "device_discovered",
			}).Info("New V4L2 device discovered")

			// Generate event
			m.generateCameraEvent(ctx, CameraEventConnected, path, device)
		} else {
			// Update existing device
			existing.LastSeen = device.LastSeen
			if existing.Status != device.Status {
				existing.Status = device.Status
				existing.Error = device.Error

				atomic.AddInt64(&m.stats.DeviceStateChanges, 1)

				m.logger.WithFields(logging.Fields{
					"device_path": path,
					"old_status":  existing.Status,
					"new_status":  device.Status,
					"action":      "device_status_changed",
				}).Info("V4L2 device status changed")

				// Generate event
				m.generateCameraEvent(ctx, CameraEventStatusChanged, path, device)
			}
		}
	}

	// Find removed devices
	for path, device := range m.knownDevices {
		if _, exists := currentDevices[path]; !exists {
			// Device removed
			device.Status = DeviceStatusDisconnected
			atomic.AddInt64(&m.stats.DeviceStateChanges, 1)

			m.logger.WithFields(logging.Fields{
				"device_path": path,
				"device_name": device.Name,
				"action":      "device_disconnected",
			}).Info("V4L2 device disconnected")

			// Generate event
			m.generateCameraEvent(ctx, CameraEventDisconnected, path, device)
		}
	}
}

// generateCameraEvent generates camera events and notifies handlers
func (m *HybridCameraMonitor) generateCameraEvent(ctx context.Context, eventType CameraEvent, devicePath string, device *CameraDevice) {
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
			defer func() {
				// Recover from panics in goroutine
				if r := recover(); r != nil {
					m.logger.WithFields(logging.Fields{
						"handler_type": fmt.Sprintf("%T", h), // Keep fmt.Sprintf for type reflection
						"panic":        r,
						"action":       "panic_recovered",
					}).Error("Recovered from panic in camera event handler")
				}
			}()

			if err := h.HandleCameraEvent(ctx, eventData); err != nil {
				m.logger.WithFields(logging.Fields{
					"handler_type": fmt.Sprintf("%T", h), // Keep fmt.Sprintf for type reflection
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
	failureCount := atomic.LoadInt64(&m.pollingFailureCount)
	if failureCount > 0 {
		// Apply failure penalty that increases polling frequency (decreases interval)
		failurePenalty := max(0.5, 1.0-float64(failureCount)*0.1) // Minimum 0.5x interval
		m.currentPollInterval = max(m.minPollInterval, m.currentPollInterval*failurePenalty)
	} else {
		// Gradually return to base interval when no failures
		m.currentPollInterval = min(m.maxPollInterval, m.currentPollInterval*1.1)
	}

	// Update stats if interval changed significantly
	if abs(m.currentPollInterval-oldInterval) > 0.01 {
		m.stats.CurrentPollInterval = m.currentPollInterval

		m.logger.WithFields(logging.Fields{
			"old_interval":  oldInterval,
			"new_interval":  m.currentPollInterval,
			"failure_count": atomic.LoadInt64(&m.pollingFailureCount),
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
