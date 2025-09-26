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
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	discoveryMode             string
	fallbackPollInterval      float64

	// Enhanced camera sources beyond USB
	cameraSources []CameraSource

	// Dependencies
	configManager     *config.ConfigManager
	logger            *logging.Logger
	deviceChecker     DeviceChecker
	commandExecutor   V4L2CommandExecutor
	infoParser        DeviceInfoParser
	deviceEventSource DeviceEventSource

	// State
	knownDevices     map[string]*CameraDevice
	capabilityStates map[string]*DeviceCapabilityState
	stopChan         chan struct{}
	ready            int32 // Atomic flag for readiness
	// No reference counting needed - monitor owns its device event source instance

	// Concurrency control
	startStopMutex sync.Mutex // Protects start/stop operations from concurrent access

	// Device-level mutex protection for V4L2 operations
	deviceMutexes map[string]*sync.Mutex // Per-device mutexes for V4L2 access
	mutexesLock   sync.RWMutex           // Protects deviceMutexes map

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

	// Event-driven readiness system
	readinessEventChan chan struct{}
	readinessMutex     sync.RWMutex

	// Internal context management (following health monitor pattern)
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Bounded worker pool for event handlers
	eventWorkerPool BoundedWorkerPool
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

	if deviceChecker == nil {
		return nil, fmt.Errorf("deviceChecker cannot be nil")
	}

	if commandExecutor == nil {
		return nil, fmt.Errorf("commandExecutor cannot be nil")
	}

	if infoParser == nil {
		return nil, fmt.Errorf("infoParser cannot be nil")
	}

	if logger == nil {
		logger = logging.GetLogger("camera-monitor")
	}

	// Create fresh device event source instance for this monitor
	deviceEventSource := GetDeviceEventSourceFactory().Create()

	// Get configuration directly from config manager
	cfg := configManager.GetConfig()
	if cfg == nil {
		// Close the device event source if config is not available
		deviceEventSource.Close()
		return nil, fmt.Errorf("no configuration loaded")
	}

	// Set default discovery mode if not configured
	discoveryMode := cfg.Camera.DiscoveryMode
	if discoveryMode == "" {
		discoveryMode = "event-first"
	}

	// Set default fallback poll interval if not configured
	fallbackPollInterval := cfg.Camera.FallbackPollInterval
	if fallbackPollInterval <= 0 {
		fallbackPollInterval = 90.0 // 90 seconds default
	}

	monitor := &HybridCameraMonitor{
		// Configuration from config manager
		deviceRange:               cfg.Camera.DeviceRange,
		pollInterval:              cfg.Camera.PollInterval,
		detectionTimeout:          cfg.Camera.DetectionTimeout,
		enableCapabilityDetection: cfg.Camera.EnableCapabilityDetection,
		discoveryMode:             discoveryMode,
		fallbackPollInterval:      fallbackPollInterval,

		// Enhanced camera sources
		cameraSources: []CameraSource{},

		// Dependencies
		configManager:     configManager,
		logger:            logger,
		deviceChecker:     deviceChecker,
		commandExecutor:   commandExecutor,
		infoParser:        infoParser,
		deviceEventSource: deviceEventSource,

		// State
		knownDevices:       make(map[string]*CameraDevice),
		capabilityStates:   make(map[string]*DeviceCapabilityState),
		stopChan:           make(chan struct{}, 10), // Buffered to prevent deadlock during shutdown
		readinessEventChan: make(chan struct{}, 1),  // Initialize readiness event channel

		// Device-level mutex protection
		deviceMutexes: make(map[string]*sync.Mutex),

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
		minPollInterval:        0.1, // 100ms minimum (same as default)
		maxPollInterval:        5.0, // 5s maximum
		maxConsecutiveFailures: 5,

		// Capability detection
		capabilityTimeout:       cfg.Camera.CapabilityTimeout,
		capabilityRetryInterval: cfg.Camera.CapabilityRetryInterval,
		capabilityMaxRetries:    cfg.Camera.CapabilityMaxRetries,
	}

	// Initialize camera sources from configuration
	monitor.initializeCameraSources()

	// Initialize bounded worker pool for event handlers
	// Use configuration values or defaults
	maxWorkers := 10                // Default
	taskTimeout := 10 * time.Second // Default fallback
	if cfg.Camera.MaxEventHandlerGoroutines > 0 {
		maxWorkers = cfg.Camera.MaxEventHandlerGoroutines
	}
	if cfg.Camera.EventHandlerTimeout > 0 {
		taskTimeout = cfg.Camera.EventHandlerTimeout
	}

	monitor.eventWorkerPool = NewBoundedWorkerPool(maxWorkers, taskTimeout, logger)

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
	// IP camera sources should be added from configuration when needed
	// This function is intentionally empty to avoid hardcoded examples
	// IP cameras can be added via configuration files or environment variables
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
	// Acquire mutex to prevent concurrent start/stop operations
	m.startStopMutex.Lock()
	defer m.startStopMutex.Unlock()

	// Generate unique startup correlation ID
	monStartID := fmt.Sprintf("mon_%d", time.Now().UnixNano())

	// Check if already running (now thread-safe with mutex)
	if atomic.LoadInt32(&m.running) == 1 {
		m.logger.WithFields(logging.Fields{
			"mon_start_id":  monStartID,
			"err_type":      "BUG_DOUBLE_START",
			"current_state": atomic.LoadInt32(&m.running),
		}).Error("monitor_start_return_err - monitor is already running")
		return fmt.Errorf("monitor is already running")
	}

	// Set running flag
	atomic.StoreInt32(&m.running, 1)

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		// Close device event source on cancellation
		if m.deviceEventSource != nil {
			if err := m.deviceEventSource.Close(); err != nil {
				m.logger.WithError(err).Warn("Error closing device event source on cancellation")
			}
		}
		atomic.StoreInt32(&m.running, 0) // Reset flag on cancellation
		m.logger.WithFields(logging.Fields{
			"mon_start_id": monStartID,
			"err_type":     "CTX_CANCELED",
		}).Error("monitor_start_return_err")
		return ctx.Err()
	default:
	}

	// Create internal context for lifecycle management (following health monitor pattern)
	m.ctx, m.cancel = context.WithCancel(ctx)

	m.logger.WithFields(logging.Fields{
		"mon_start_id": monStartID,
		"mode_config":  m.discoveryMode,
		"action":       "monitor_start_begin",
	}).Info("Starting hybrid camera monitor")

	// Start device event source - trust its return value
	m.logger.WithFields(logging.Fields{
		"mon_start_id": monStartID,
		"action":       "event_source_start_begin",
	}).Info("Starting device event source")

	// Release mutex before calling device event source operations to avoid deadlock
	m.startStopMutex.Unlock()

	if err := m.deviceEventSource.Start(m.ctx); err != nil {
		// Re-acquire mutex for cleanup
		m.startStopMutex.Lock()
		atomic.StoreInt32(&m.running, 0) // Reset flag on failure
		// Close the device event source since we failed to start it
		if m.deviceEventSource != nil {
			if err := m.deviceEventSource.Close(); err != nil {
				m.logger.WithError(err).Warn("Error closing device event source after start failure")
			}
		}
		m.logger.WithFields(logging.Fields{
			"mon_start_id": monStartID,
			"err_type":     "ES_START_FATAL",
			"err_msg":      err.Error(),
		}).Error("monitor_start_return_err")
		return fmt.Errorf("failed to start device event source: %w", err)
	}

	// Re-acquire mutex for final setup
	m.startStopMutex.Lock()

	// Device event source is now started and owned by this monitor

	// Set up cleanup function in case of failure after this point
	// Note: This cleanup function is available for future use if needed
	_ = func() {
		atomic.StoreInt32(&m.running, 0)
		if err := m.deviceEventSource.Close(); err != nil {
			m.logger.WithError(err).Warn("Error closing device event source during cleanup")
		}
	}

	m.logger.WithFields(logging.Fields{
		"mon_start_id":     monStartID,
		"events_supported": m.deviceEventSource.EventsSupported(),
		"action":           "event_source_start_ok",
	}).Info("Device event source started successfully")

	// Log the mode we're running in
	if m.deviceEventSource.EventsSupported() {
		m.logger.Info("Running in event-first mode with fsnotify support")
	} else {
		m.logger.Info("Running in poll-only mode (fsnotify not available)")
	}

	// Initialize state while holding lock
	m.stateLock.Lock()
	m.stats.Running = true
	atomic.StoreInt64(&m.stats.ActiveTasks, 1)
	m.stateLock.Unlock()

	// Start the event worker pool
	if err := m.eventWorkerPool.Start(m.ctx); err != nil {
		// Re-acquire mutex for cleanup
		m.startStopMutex.Lock()
		atomic.StoreInt32(&m.running, 0)
		if m.deviceEventSource != nil {
			m.deviceEventSource.Close()
		}
		m.logger.WithError(err).WithField("mon_start_id", monStartID).Error("Failed to start event worker pool")
		return fmt.Errorf("failed to start event worker pool: %w", err)
	}

	// Start monitoring goroutine AFTER releasing lock
	m.logger.WithFields(logging.Fields{
		"mon_start_id":     monStartID,
		"events_supported": m.deviceEventSource.EventsSupported(),
		"action":           "loops_spawn_begin",
	}).Info("Spawning monitor loops")

	m.wg.Add(1)
	go m.monitorLoop(m.ctx, monStartID)

	m.logger.WithFields(logging.Fields{
		"mon_start_id": monStartID,
		"action":       "monitor_start_return_ok",
	}).Info("Monitor start completed successfully")
	return nil
}

// monitorLoop runs the main monitoring loop
func (m *HybridCameraMonitor) monitorLoop(ctx context.Context, monStartID string) {
	defer m.wg.Done() // Ensure WaitGroup is decremented when goroutine exits

	m.logger.Debug("Monitor loop started")
	defer func() {
		m.logger.Debug("Monitor loop exiting")
		// Don't reset running flag here - let Stop() method handle it
		m.stats.Running = false
		atomic.StoreInt64(&m.stats.ActiveTasks, 0)
	}()

	// Perform initial discovery to seed knownDevices and ensure IsReady() becomes true
	m.logger.WithFields(logging.Fields{
		"mon_start_id": monStartID,
		"action":       "seed_discovery_begin",
	}).Info("Starting seed discovery")

	m.logger.Debug("About to call discoverCameras")

	// Seed discovery is non-fatal - log any errors but continue
	func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithFields(logging.Fields{
					"mon_start_id": monStartID,
					"panic":        r,
					"err_summary":  "panic_in_discovery",
				}).Warn("Seed discovery panic recovered")
			}
		}()
		m.discoverCameras(ctx)
	}()

	m.logger.Debug("discoverCameras completed")

	// Set readiness flag after initial discovery completes (regardless of errors)
	atomic.StoreInt32(&m.ready, 1)

	// Emit readiness event for event-driven systems
	m.emitReadinessEvent()

	m.logger.WithFields(logging.Fields{
		"mon_start_id":  monStartID,
		"found_devices": len(m.knownDevices),
		"action":        "seed_discovery_result",
	}).Info("Seed discovery completed")

	m.logger.WithFields(logging.Fields{
		"mon_start_id": monStartID,
		"action":       "monitor_ready_true",
	}).Info("Monitor is now ready")

	// Emit readiness event when monitor becomes ready
	m.emitReadinessEvent()

	if m.discoveryMode == "event-first" {
		// Start event-first monitoring
		m.startEventFirstMonitoring(ctx)
	} else {
		// Start poll-only monitoring
		m.startPollOnlyMonitoring(ctx)
	}
}

// Stop stops camera discovery and monitoring with context-aware cancellation
func (m *HybridCameraMonitor) Stop(ctx context.Context) error {
	// Acquire mutex to prevent concurrent start/stop operations
	m.startStopMutex.Lock()
	defer m.startStopMutex.Unlock()

	m.logger.Info("Stopping hybrid camera monitor")

	// Check if monitor is actually running - make Stop() idempotent
	if atomic.LoadInt32(&m.running) == 0 {
		m.logger.Debug("Monitor is already stopped - Stop() is idempotent")
		return nil // Idempotent: safe to call Stop() on non-running monitor
	}

	// Don't set running flag to 0 yet - wait until after device event source is closed

	// Cancel internal context first - this interrupts monitorLoop immediately!
	if m.cancel != nil {
		m.cancel()
	}

	// Wait with timeout (following health monitor pattern)
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		// Force shutdown after timeout
		m.logger.Warn("Camera monitor shutdown timeout, forcing stop")
		// Return the context error to indicate timeout
		return ctx.Err()
	}

	// Close the device event source first - ensure it's properly stopped
	if m.deviceEventSource != nil {
		// Release mutex before calling device event source operations to avoid deadlock
		m.startStopMutex.Unlock()

		if err := m.deviceEventSource.Close(); err != nil {
			m.logger.WithError(err).Warn("Error closing device event source")
		}
		// Verify the device event source is actually stopped
		if m.deviceEventSource.Started() {
			m.logger.Warn("Device event source is still started after Close() call")
		}

		// Re-acquire mutex for factory operations
		m.startStopMutex.Lock()
	}

	// Device event source is now properly closed and owned by this monitor

	// Stop the event worker pool
	if m.eventWorkerPool != nil {
		if err := m.eventWorkerPool.Stop(ctx); err != nil {
			m.logger.WithError(err).Warn("Error stopping event worker pool")
		}
	}

	// Set running flag to 0 after device event source is closed
	atomic.StoreInt32(&m.running, 0)
	atomic.StoreInt32(&m.ready, 0)

	m.logger.Info("Hybrid camera monitor stopped")
	return nil
}

// IsRunning returns whether the monitor is currently running
func (m *HybridCameraMonitor) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

// IsReady returns whether the monitor has completed initial discovery and is ready
func (m *HybridCameraMonitor) IsReady() bool {
	return atomic.LoadInt32(&m.ready) == 1
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
		Running:                    atomic.LoadInt32(&m.running) == 1, // Use atomic field consistently
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
		DeviceEventsProcessed:      atomic.LoadInt64(&m.stats.DeviceEventsProcessed),
		DeviceEventsDropped:        atomic.LoadInt64(&m.stats.DeviceEventsDropped),
		DevicesConnected:           atomic.LoadInt64(&m.stats.DevicesConnected),
	}
	return &stats
}

// GetResourceStats returns resource management statistics (implements camera.CleanupManager)
func (m *HybridCameraMonitor) GetResourceStats() map[string]interface{} {
	stats := map[string]interface{}{
		"running":                m.IsRunning(),
		"known_devices_count":    len(m.knownDevices),
		"active_event_handlers":  len(m.eventHandlers),
		"active_event_callbacks": len(m.eventCallbacks),
	}

	// Add worker pool statistics if available
	if m.eventWorkerPool != nil {
		workerStats := m.eventWorkerPool.GetStats()
		stats["worker_pool"] = map[string]interface{}{
			"active_workers":  workerStats.ActiveWorkers,
			"queued_tasks":    workerStats.QueuedTasks,
			"completed_tasks": workerStats.CompletedTasks,
			"failed_tasks":    workerStats.FailedTasks,
			"timeout_tasks":   workerStats.TimeoutTasks,
			"max_workers":     workerStats.MaxWorkers,
		}
	}

	return stats
}

// getDeviceMutex returns a device-specific mutex for V4L2 operations
// This ensures only one concurrent operation per device to prevent V4L2 locking errors
func (m *HybridCameraMonitor) getDeviceMutex(devicePath string) *sync.Mutex {
	m.mutexesLock.RLock()
	deviceMutex, exists := m.deviceMutexes[devicePath]
	if exists {
		m.mutexesLock.RUnlock()
		return deviceMutex
	}
	m.mutexesLock.RUnlock()

	// Create new mutex for this device
	m.mutexesLock.Lock()
	defer m.mutexesLock.Unlock()

	// Double-check in case another goroutine created it
	if deviceMutex, exists := m.deviceMutexes[devicePath]; exists {
		return deviceMutex
	}

	// Create and store new mutex
	deviceMutex = &sync.Mutex{}
	m.deviceMutexes[devicePath] = deviceMutex
	return deviceMutex
}

// TakeDirectSnapshot captures a snapshot directly via V4L2 (Tier 0 - Fastest)
func (m *HybridCameraMonitor) TakeDirectSnapshot(ctx context.Context, devicePath, outputPath string, options map[string]interface{}) (*DirectSnapshot, error) {
	startTime := time.Now()

	m.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"output_path": outputPath,
		"options":     options,
		"tier":        0,
	}).Info("Taking V4L2 direct snapshot")

	// Get device-specific mutex for V4L2 operations
	deviceMutex := m.getDeviceMutex(devicePath)
	deviceMutex.Lock()
	defer deviceMutex.Unlock()

	// Validate device exists using existing infrastructure
	if !m.deviceChecker.Exists(devicePath) {
		return nil, fmt.Errorf("device %s does not exist", devicePath)
	}

	// Get device information for validation
	device, exists := m.GetDevice(devicePath)
	if !exists {
		return nil, fmt.Errorf("device %s not found in monitor", devicePath)
	}

	if device.Status != DeviceStatusConnected {
		return nil, fmt.Errorf("device %s is not connected (status: %s)", devicePath, device.Status)
	}

	// Extract options with defaults
	format := "jpg"
	if f, ok := options["format"].(string); ok && f != "" {
		format = f
	}

	width := 0
	height := 0
	if w, ok := options["width"].(int); ok && w > 0 {
		width = w
	}
	if h, ok := options["height"].(int); ok && h > 0 {
		height = h
	}

	// Select optimal pixel format based on camera capabilities
	optimalPixelFormat, err := m.selectOptimalPixelFormat(devicePath, format)
	if err != nil {
		return nil, fmt.Errorf("failed to select optimal pixel format: %w", err)
	}

	// Build V4L2 command arguments
	args := m.buildV4L2SnapshotArgs(devicePath, outputPath, format, width, height)

	// Add the selected pixel format to the arguments
	args += fmt.Sprintf(" --set-fmt-video pixelformat=%s", optimalPixelFormat)

	// Log the complete command being executed
	fullCommand := fmt.Sprintf("v4l2-ctl --device=%s %s", devicePath, args)
	m.logger.WithFields(logging.Fields{
		"device_path":  devicePath,
		"output_path":  outputPath,
		"format":       format,
		"width":        width,
		"height":       height,
		"pixel_format": optimalPixelFormat,
		"command_args": args,
		"full_command": fullCommand,
	}).Info("Executing V4L2 direct snapshot command")

	// Also log the command directly for debugging
	fmt.Printf("DEBUG: V4L2 Command: %s\n", fullCommand)

	// Execute V4L2 direct capture with fallback format selection
	output, err := m.commandExecutor.ExecuteCommand(ctx, devicePath, args)
	if err != nil {
		// Log the command output and error details
		m.logger.WithFields(logging.Fields{
			"device_path":    devicePath,
			"optimal_format": optimalPixelFormat,
			"error":          err.Error(),
			"command_output": output,
			"full_command":   fmt.Sprintf("v4l2-ctl --device=%s %s", devicePath, args),
		}).Warn("Optimal pixel format failed, trying fallback formats")

		// Try fallback formats
		fallbackFormats := m.getFallbackFormats(devicePath, format)
		success := false
		for _, fallbackFormat := range fallbackFormats {
			if fallbackFormat == optimalPixelFormat {
				continue // Skip the format we already tried
			}

			fallbackArgs := m.buildV4L2SnapshotArgs(devicePath, outputPath, format, width, height)
			fallbackArgs += fmt.Sprintf(" --set-fmt-video pixelformat=%s", fallbackFormat)

			m.logger.WithFields(logging.Fields{
				"device_path":      devicePath,
				"fallback_format":  fallbackFormat,
				"fallback_command": fmt.Sprintf("v4l2-ctl --device=%s %s", devicePath, fallbackArgs),
			}).Info("Trying fallback pixel format")

			fallbackOutput, fallbackErr := m.commandExecutor.ExecuteCommand(ctx, devicePath, fallbackArgs)
			if fallbackErr == nil {
				// Success with fallback format
				m.logger.WithFields(logging.Fields{
					"device_path":     devicePath,
					"fallback_format": fallbackFormat,
					"fallback_output": fallbackOutput,
				}).Info("Fallback format succeeded")
				m.logger.WithFields(logging.Fields{
					"device_path":       devicePath,
					"successful_format": fallbackFormat,
				}).Info("V4L2 direct snapshot succeeded with fallback format")
				success = true
				break
			} else {
				m.logger.WithFields(logging.Fields{
					"device_path":      devicePath,
					"fallback_format":  fallbackFormat,
					"fallback_error":   fallbackErr.Error(),
					"fallback_output":  fallbackOutput,
					"fallback_command": fmt.Sprintf("v4l2-ctl --device=%s %s", devicePath, fallbackArgs),
				}).Warn("Fallback format also failed")
			}
		}

		// If all formats failed, return the original error
		if !success {
			m.logger.WithFields(logging.Fields{
				"device_path": devicePath,
				"output_path": outputPath,
				"error":       err.Error(),
				"tier":        0,
			}).Error("V4L2 direct snapshot failed with all formats")
			return nil, fmt.Errorf("V4L2 direct capture failed: %w", err)
		}
	}

	// Verify file was created and get size
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to verify snapshot file: %w", err)
	}

	captureTime := time.Since(startTime)

	// Create snapshot result
	snapshot := &DirectSnapshot{
		ID:          m.generateSnapshotID(devicePath),
		DevicePath:  devicePath,
		FilePath:    outputPath,
		Size:        fileInfo.Size(),
		Format:      format,
		Width:       width,
		Height:      height,
		CaptureTime: captureTime,
		Created:     time.Now(),
		Metadata: map[string]interface{}{
			"tier_used":      0,
			"capture_method": "v4l2_direct",
			"device_name":    device.Name,
			"device_caps":    device.Capabilities.Capabilities,
		},
	}

	m.logger.WithFields(logging.Fields{
		"device_path":  devicePath,
		"output_path":  outputPath,
		"file_size":    fileInfo.Size(),
		"capture_time": captureTime,
		"format":       format,
		"tier":         0,
	}).Info("V4L2 direct snapshot successful")

	return snapshot, nil
}

// buildV4L2SnapshotArgs builds V4L2 command arguments for direct snapshot capture
func (m *HybridCameraMonitor) buildV4L2SnapshotArgs(devicePath, outputPath, format string, width, height int) string {
	// FIXED: Remove devicePath from args (RealV4L2CommandExecutor adds --device automatically)
	args := []string{
		"--stream-mmap",           // Use memory-mapped streaming
		"--stream-to", outputPath, // Output file path
		"--stream-count", "1", // Capture only 1 frame
	}

	// Add resolution if specified
	if width > 0 && height > 0 {
		args = append(args, "--set-fmt-video", fmt.Sprintf("width=%d,height=%d", width, height))
	}

	// FIXED: Don't add pixel format here - it's added in TakeDirectSnapshot with optimal format
	// This prevents conflicting --set-fmt-video parameters

	return strings.Join(args, " ")
}

// selectOptimalPixelFormat selects the best pixel format for the given output format and camera capabilities
func (m *HybridCameraMonitor) selectOptimalPixelFormat(devicePath, outputFormat string) (string, error) {
	m.logger.WithFields(logging.Fields{
		"device":        devicePath,
		"output_format": outputFormat,
	}).Debug("Selecting optimal pixel format for camera")

	// Get device capabilities
	device, exists := m.GetDevice(devicePath)
	if !exists {
		return "", fmt.Errorf("device %s not found in monitor", devicePath)
	}

	// Check if device has format information
	if len(device.Formats) == 0 {
		m.logger.WithField("device", devicePath).Warn("No format capabilities available, using fallback")
		return m.getFallbackPixelFormat(outputFormat), nil
	}

	// Define format preferences based on output format
	var preferredFormats []string
	switch strings.ToLower(outputFormat) {
	case "jpg", "jpeg":
		// For JPEG output, prefer compressed formats that match
		preferredFormats = []string{"MJPG", "MJPEG", "JPEG", "H264", "YUYV", "RGB24"}
	case "png":
		// For PNG output, prefer uncompressed formats that can be converted
		preferredFormats = []string{"YUYV", "RGB24", "BGR24", "UYVY", "YUV420", "MJPG", "MJPEG"}
	case "":
		// Default: prefer compressed formats for efficiency and resolution
		preferredFormats = []string{"MJPG", "MJPEG", "H264", "YUYV", "RGB24"}
	default:
		// Unknown format, use default preferences
		preferredFormats = []string{"MJPG", "MJPEG", "H264", "YUYV", "RGB24"}
	}

	// Find the best supported format
	for _, preferred := range preferredFormats {
		for _, deviceFormat := range device.Formats {
			if strings.EqualFold(deviceFormat.PixelFormat, preferred) {
				m.logger.WithFields(logging.Fields{
					"device":          devicePath,
					"selected_format": preferred,
					"output_format":   outputFormat,
				}).Info("Selected optimal pixel format")
				return preferred, nil
			}
		}
	}

	// If no preferred format is found, use the first available format
	if len(device.Formats) > 0 {
		fallback := device.Formats[0].PixelFormat
		m.logger.WithFields(logging.Fields{
			"device":          devicePath,
			"fallback_format": fallback,
			"output_format":   outputFormat,
		}).Warn("Using fallback pixel format")
		return fallback, nil
	}

	// Ultimate fallback
	fallback := m.getFallbackPixelFormat(outputFormat)
	m.logger.WithFields(logging.Fields{
		"device":            devicePath,
		"ultimate_fallback": fallback,
	}).Warn("Using ultimate fallback pixel format")
	return fallback, nil
}

// SelectOptimalPixelFormat exposes optimal pixel format selection for external callers.
// It delegates to the internal selectOptimalPixelFormat implementation.
func (m *HybridCameraMonitor) SelectOptimalPixelFormat(devicePath, outputFormat string) (string, error) {
	return m.selectOptimalPixelFormat(devicePath, outputFormat)
}

// getFallbackPixelFormat provides a safe fallback when capability detection fails
func (m *HybridCameraMonitor) getFallbackPixelFormat(outputFormat string) string {
	switch strings.ToLower(outputFormat) {
	case "jpg", "jpeg":
		// Try MJPG first for best quality/performance balance
		return "MJPG"
	case "png":
		// For PNG, use uncompressed format
		return "YUYV"
	default:
		// Default fallback - prefer MJPG for modern cameras, YUYV for compatibility
		return "MJPG"
	}
}

// getFallbackFormats returns a list of fallback formats to try if the optimal format fails
func (m *HybridCameraMonitor) getFallbackFormats(devicePath, outputFormat string) []string {
	// Get device capabilities
	device, exists := m.GetDevice(devicePath)
	if !exists {
		// Return generic fallbacks if device not found
		return m.getGenericFallbackFormats(outputFormat)
	}

	// If device has format information, use it
	if len(device.Formats) > 0 {
		var fallbacks []string
		for _, deviceFormat := range device.Formats {
			fallbacks = append(fallbacks, deviceFormat.PixelFormat)
		}
		return fallbacks
	}

	// Return generic fallbacks
	return m.getGenericFallbackFormats(outputFormat)
}

// getGenericFallbackFormats returns common fallback formats when device capabilities are unknown
func (m *HybridCameraMonitor) getGenericFallbackFormats(outputFormat string) []string {
	switch strings.ToLower(outputFormat) {
	case "jpg", "jpeg":
		// Try common JPEG-compatible formats in priority order
		return []string{"MJPG", "MJPEG", "H264", "YUYV", "RGB24", "BGR24"}
	case "png":
		// Try common uncompressed formats for PNG
		return []string{"YUYV", "RGB24", "BGR24", "UYVY", "YUV420"}
	default:
		// Generic fallbacks in priority order
		return []string{"MJPG", "MJPEG", "H264", "YUYV", "RGB24", "BGR24"}
	}
}

// generateSnapshotID generates a unique snapshot ID
func (m *HybridCameraMonitor) generateSnapshotID(devicePath string) string {
	timestamp := time.Now().UnixNano()
	deviceName := filepath.Base(devicePath)
	return fmt.Sprintf("v4l2_direct_%s_%d", deviceName, timestamp)
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

// SubscribeToReadiness subscribes to camera monitor readiness events
// FIXED: Creates per-subscriber channel instead of returning shared channel
func (m *HybridCameraMonitor) SubscribeToReadiness() <-chan struct{} {
	m.readinessMutex.Lock()
	defer m.readinessMutex.Unlock()

	// Create unique channel for this subscriber
	subscriberChan := make(chan struct{}, 1)

	// If already ready, send immediate notification
	if m.IsReady() {
		select {
		case subscriberChan <- struct{}{}:
		default:
			// Channel is full, skip notification
		}
	}

	return subscriberChan
}

// emitReadinessEvent emits a readiness event to all subscribers
func (m *HybridCameraMonitor) emitReadinessEvent() {
	m.readinessMutex.RLock()
	defer m.readinessMutex.RUnlock()

	// Send to the main readiness channel if it exists
	if m.readinessEventChan != nil {
		select {
		case m.readinessEventChan <- struct{}{}:
		default:
			// Channel is full, skip this event
		}
	}
}

// SetEventNotifier sets the event notifier for external event system integration
func (m *HybridCameraMonitor) SetEventNotifier(notifier EventNotifier) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	m.eventNotifier = notifier
}

// startEventFirstMonitoring starts event-first monitoring with slow reconcile fallback
// INVARIANT: Event source lifecycle is owned by HybridCameraMonitor.Start/Stop.
// No other method starts it - the source is already started when this is called.
func (m *HybridCameraMonitor) startEventFirstMonitoring(ctx context.Context) {
	m.logger.WithFields(logging.Fields{
		"action": "event_first_monitoring_started",
	}).Info("Starting event-first camera monitoring")

	// Device event source is already started in Start() method
	// Just verify it's running and has event support
	if !m.deviceEventSource.EventsSupported() {
		m.logger.Warn("Device event source doesn't support events, falling back to poll-only")
		m.startPollOnlyMonitoring(ctx)
		return
	}

	// Start event processing loop
	go m.eventLoop(ctx)

	// Start slow reconcile loop for drift correction
	go m.reconcileLoop(ctx)
}

// startPollOnlyMonitoring starts poll-only monitoring
func (m *HybridCameraMonitor) startPollOnlyMonitoring(ctx context.Context) {
	m.logger.WithFields(logging.Fields{
		"poll_interval": m.pollInterval,
		"device_range":  m.deviceRange,
		"action":        "poll_only_monitoring_started",
	}).Info("Starting poll-only camera monitoring")

	// Use the configured poll interval for poll-only mode
	go m.reconcileLoop(ctx)
}

// eventLoop processes device events from the event source
func (m *HybridCameraMonitor) eventLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.WithFields(logging.Fields{
				"panic":  r,
				"action": "panic_recovered",
			}).Error("Recovered from panic in event loop")
		}
	}()

	m.logger.WithFields(logging.Fields{
		"action": "event_loop_started",
	}).Debug("Device event loop started")

	for {
		select {
		case <-ctx.Done():
			m.logger.WithFields(logging.Fields{
				"action": "event_loop_stopped",
				"reason": "context_cancelled",
			}).Debug("Device event loop stopped due to context cancellation")
			return
		case <-m.stopChan:
			m.logger.WithFields(logging.Fields{
				"action": "event_loop_stopped",
				"reason": "stop_requested",
			}).Debug("Device event loop stopped")
			return
		case event, ok := <-m.deviceEventSource.Events():
			if !ok {
				m.logger.Debug("Device event source channel closed")
				return
			}
			m.processDeviceEvent(ctx, event)
		}
	}
}

// reconcileLoop performs slow periodic reconciliation to correct drift
func (m *HybridCameraMonitor) reconcileLoop(ctx context.Context) {
	// Use fallback poll interval for event-first mode, configured interval for poll-only
	pollInterval := m.fallbackPollInterval
	if m.discoveryMode == "poll-only" {
		pollInterval = m.pollInterval
	}

	ticker := time.NewTicker(time.Duration(pollInterval * float64(time.Second)))
	defer ticker.Stop()

	m.logger.WithFields(logging.Fields{
		"poll_interval": pollInterval,
		"mode":          m.discoveryMode,
		"action":        "reconcile_loop_started",
	}).Info("Camera reconcile loop started")

	for {
		select {
		case <-ctx.Done():
			m.logger.WithFields(logging.Fields{
				"action": "reconcile_loop_stopped",
				"reason": "context_cancelled",
			}).Info("Camera reconcile loop stopped due to context cancellation")
			return
		case <-m.stopChan:
			m.logger.WithFields(logging.Fields{
				"action": "reconcile_loop_stopped",
				"reason": "stop_requested",
			}).Info("Camera reconcile loop stopped")
			return
		case <-ticker.C:
			m.reconcileDevices(ctx)
		}
	}
}

// processDeviceEvent processes a single device event from the event source
func (m *HybridCameraMonitor) processDeviceEvent(ctx context.Context, event DeviceEvent) {
	atomic.AddInt64(&m.stats.DeviceEventsProcessed, 1)

	m.logger.WithFields(logging.Fields{
		"action":      "device_event_processing",
		"device_path": event.DevicePath,
		"event_type":  event.Type,
		"vendor":      event.Vendor,
		"product":     event.Product,
		"serial":      event.Serial,
	}).Debug("Processing device event")

	switch event.Type {
	case DeviceEventAdd:
		m.handleDeviceAdd(ctx, event)
	case DeviceEventRemove:
		m.handleDeviceRemove(ctx, event)
	case DeviceEventChange:
		m.handleDeviceChange(ctx, event)
	default:
		m.logger.WithFields(logging.Fields{
			"device_path": event.DevicePath,
			"event_type":  event.Type,
			"action":      "unknown_event_type",
		}).Warn("Unknown device event type")
	}
}

// handleDeviceAdd handles device add events
func (m *HybridCameraMonitor) handleDeviceAdd(ctx context.Context, event DeviceEvent) {
	// Check if device already exists in our map
	m.stateLock.RLock()
	_, exists := m.knownDevices[event.DevicePath]
	m.stateLock.RUnlock()

	if exists {
		// Device already known, skip
		return
	}

	// Create device info for the new device
	device, err := m.createDeviceFromEvent(ctx, event)
	if err != nil {
		m.logger.WithFields(logging.Fields{
			"device_path": event.DevicePath,
			"error":       err.Error(),
			"action":      "device_add_failed",
		}).Warn("Failed to create device info for new device")
		return
	}

	// Add to known devices
	m.stateLock.Lock()
	m.knownDevices[event.DevicePath] = device
	atomic.AddInt64(&m.stats.DeviceStateChanges, 1)
	atomic.AddInt64(&m.stats.DevicesConnected, 1)
	m.stateLock.Unlock()

	m.logger.WithFields(logging.Fields{
		"device_path": event.DevicePath,
		"device_name": device.Name,
		"status":      device.Status,
		"action":      "device_discovered",
	}).Info("New V4L2 device discovered via event")

	// Generate event
	m.generateCameraEvent(ctx, CameraEventConnected, event.DevicePath, device)
}

// handleDeviceRemove handles device remove events
func (m *HybridCameraMonitor) handleDeviceRemove(ctx context.Context, event DeviceEvent) {
	m.stateLock.Lock()
	device, exists := m.knownDevices[event.DevicePath]
	if exists {
		device.Status = DeviceStatusDisconnected
		atomic.AddInt64(&m.stats.DeviceStateChanges, 1)
		atomic.AddInt64(&m.stats.DevicesConnected, -1)
	}
	m.stateLock.Unlock()

	if exists {
		m.logger.WithFields(logging.Fields{
			"device_path": event.DevicePath,
			"device_name": device.Name,
			"action":      "device_disconnected",
		}).Info("V4L2 device disconnected via event")

		// Generate event
		m.generateCameraEvent(ctx, CameraEventDisconnected, event.DevicePath, device)
	}
}

// handleDeviceChange handles device change events
func (m *HybridCameraMonitor) handleDeviceChange(ctx context.Context, event DeviceEvent) {
	m.stateLock.RLock()
	device, exists := m.knownDevices[event.DevicePath]
	m.stateLock.RUnlock()

	if !exists {
		// Device not known, treat as add
		m.handleDeviceAdd(ctx, event)
		return
	}

	// Update device info
	device.LastSeen = time.Now()
	// Could probe capabilities again here if needed

	m.logger.WithFields(logging.Fields{
		"device_path": event.DevicePath,
		"device_name": device.Name,
		"action":      "device_changed",
	}).Debug("V4L2 device changed via event")

	// Generate event
	m.generateCameraEvent(ctx, CameraEventStatusChanged, event.DevicePath, device)
}

// createDeviceFromEvent creates a CameraDevice from a DeviceEvent
func (m *HybridCameraMonitor) createDeviceFromEvent(ctx context.Context, event DeviceEvent) (*CameraDevice, error) {
	// Extract device number from path
	var deviceNum int
	_, err := fmt.Sscanf(event.DevicePath, "/dev/video%d", &deviceNum)
	if err != nil {
		return nil, fmt.Errorf("invalid device path: %s", event.DevicePath)
	}

	device := &CameraDevice{
		Path:      event.DevicePath,
		DeviceNum: deviceNum,
		Status:    DeviceStatusConnected,
		LastSeen:  time.Now(),
		Vendor:    event.Vendor,
		Product:   event.Product,
		Serial:    event.Serial,
	}

	// Set default name
	device.Name = fmt.Sprintf("Video Device %d", deviceNum)

	// Probe capabilities if enabled
	if m.enableCapabilityDetection {
		if err := m.probeDeviceCapabilities(ctx, device); err != nil {
			device.Status = DeviceStatusError
			device.Error = err.Error()
			m.logger.WithFields(logging.Fields{
				"device_path": event.DevicePath,
				"error":       err.Error(),
				"action":      "capability_probe_failed",
			}).Debug("Failed to probe device capabilities")
		}
	}

	return device, nil
}

// reconcileDevices performs slow periodic reconciliation to correct drift
func (m *HybridCameraMonitor) reconcileDevices(ctx context.Context) {
	atomic.AddInt64(&m.stats.PollingCycles, 1)

	m.logger.WithFields(logging.Fields{
		"action": "reconcile_devices",
	}).Debug("Performing device reconciliation")

	// Get current devices from filesystem (only real devices, no index guessing)
	currentDevices := make(map[string]*CameraDevice)

	// Check each device in our known devices to see if it still exists
	m.stateLock.RLock()
	knownPaths := make([]string, 0, len(m.knownDevices))
	for path := range m.knownDevices {
		knownPaths = append(knownPaths, path)
	}
	m.stateLock.RUnlock()

	for _, path := range knownPaths {
		if m.deviceChecker.Exists(path) {
			// Device still exists, keep it
			m.stateLock.RLock()
			device := m.knownDevices[path]
			m.stateLock.RUnlock()

			if device != nil {
				device.LastSeen = time.Now()
				currentDevices[path] = device
			}
		}
		// If device doesn't exist, it will be removed in processDeviceStateChanges
	}

	// Process any state changes (removals)
	m.processDeviceStateChanges(ctx, currentDevices)
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
					"source":      src.Identifier,
					"source_type": src.Type,
					"source_path": src.Source,
					"error":       err.Error(),
					"action":      "device_check_error",
				}).Debug("Error checking device")

				// Propagate device check errors with more context
				select {
				case errorChan <- fmt.Errorf("device check failed for source %s (type: %s, path: %s): %w", src.Identifier, src.Type, src.Source, err):
				default:
					m.logger.WithError(err).WithFields(logging.Fields{
						"source":      src.Identifier,
						"source_type": src.Type,
						"source_path": src.Source,
					}).Warn("Error channel overflow, device check error dropped")
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
		m.logger.WithError(err).WithField("error_type", "device_check").Warn("Device check error occurred")
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
		return m.createNetworkCameraDeviceInfo(source)
	case "file":
		return m.createFileCameraDeviceInfo(source)
	default:
		return m.createGenericCameraDeviceInfo(source)
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
		return nil, fmt.Errorf("device does not exist: %s", devicePath)
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
func (m *HybridCameraMonitor) createNetworkCameraDeviceInfo(source CameraSource) (*CameraDevice, error) {
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
func (m *HybridCameraMonitor) createFileCameraDeviceInfo(source CameraSource) (*CameraDevice, error) {
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
func (m *HybridCameraMonitor) createGenericCameraDeviceInfo(source CameraSource) (*CameraDevice, error) {
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

	// Notify event handlers using bounded worker pool
	m.eventHandlersLock.RLock()
	handlers := make([]CameraEventHandler, len(m.eventHandlers))
	copy(handlers, m.eventHandlers)
	m.eventHandlersLock.RUnlock()

	for _, handler := range handlers {
		h := handler // Capture for closure
		if err := m.eventWorkerPool.Submit(ctx, func(taskCtx context.Context) {
			if err := h.HandleCameraEvent(taskCtx, eventData); err != nil {
				m.logger.WithFields(logging.Fields{
					"handler_type": fmt.Sprintf("%T", h), // Keep fmt.Sprintf for type reflection
					"error":        err.Error(),
					"action":       "event_handler_error",
				}).Error("Error in camera event handler")
			}
		}); err != nil {
			m.logger.WithFields(logging.Fields{
				"handler_type": fmt.Sprintf("%T", h), // Keep fmt.Sprintf for type reflection
				"error":        err.Error(),
				"action":       "event_handler_submit_failed",
			}).Warn("Failed to submit event handler to worker pool")
		}
	}

	// Notify event callbacks using bounded worker pool
	m.eventHandlersLock.RLock()
	callbacks := make([]func(CameraEventData), len(m.eventCallbacks))
	copy(callbacks, m.eventCallbacks)
	m.eventHandlersLock.RUnlock()

	for _, callback := range callbacks {
		cb := callback // Capture for closure
		if err := m.eventWorkerPool.Submit(ctx, func(taskCtx context.Context) {
			cb(eventData)
		}); err != nil {
			m.logger.WithFields(logging.Fields{
				"error":  err.Error(),
				"action": "event_callback_submit_failed",
			}).Warn("Failed to submit event callback to worker pool")
		}
	}
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
