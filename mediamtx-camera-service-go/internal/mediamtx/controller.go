// Package mediamtx implements the MediaMTX controller as the single source of truth
// for all video operations and business logic coordination.
//
// The controller serves as Layer 5 (Orchestration) in the component hierarchy,
// coordinating all managers and providing API abstraction between external
// identifiers (camera0, camera1) and internal device paths (/dev/videoN).
//
// Architecture Compliance:
//   - Single Source of Truth: All business logic resides in the controller
//   - No Business Logic in WebSocket: Server delegates all operations to controller
//   - Interface-based Design: Uses dependency injection for all components
//   - Optional Component Pattern: Some components may be nil based on configuration
//   - Stateless Recording: MediaMTX API is the source of truth for recording state
//
// Key Responsibilities:
//   - Camera operations coordination and abstraction layer management
//   - Stream lifecycle management with path reuse optimization
//   - Recording orchestration using stateless MediaMTX API queries
//   - Snapshot capture with multi-tier fallback (V4L2 → FFmpeg → RTSP)
//   - Health monitoring and system readiness coordination
//   - Event notification for real-time client updates
//
// Requirements Coverage:
//   - REQ-MTX-001: MediaMTX service integration with REST API
//   - REQ-MTX-002: Stream management with on-demand FFmpeg processes
//   - REQ-MTX-003: Path creation, deletion, and lifecycle management
//   - REQ-MTX-004: Health monitoring with circuit breaker pattern
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/api/json_rpc_methods.md

package mediamtx

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
)

// controller implements the MediaMTX controller as the central orchestration component.
// It serves as the single source of truth for all video operations, coordinating
// between hardware abstraction (camera monitor), MediaMTX integration (client/managers),
// and API abstraction (camera0 ↔ /dev/video0 mapping).
//
// Thread Safety: This struct is designed for concurrent access with appropriate
// synchronization primitives protecting all shared state.
type controller struct {
	// Layer 2: Core Services - MediaMTX integration and hardware monitoring
	client               MediaMTXClient        // HTTP REST API client for MediaMTX server
	healthMonitor        HealthMonitor         // MediaMTX service health monitoring with circuit breaker
	cameraMonitor        camera.CameraMonitor  // Hardware abstraction with event-driven discovery
	systemMetricsManager *SystemMetricsManager // System-wide resource monitoring and metrics collection

	// Layer 3: Managers - Specialized operation handlers
	pathManager     PathManager           // MediaMTX path lifecycle management
	pathIntegration *PathIntegration      // Optional: Auto-path creation (may be nil)
	streamManager   StreamManager         // Stream lifecycle and FFmpeg coordination
	ffmpegManager   FFmpegManager         // On-demand FFmpeg process management
	rtspManager     RTSPConnectionManager // RTSP connection pooling and keepalive

	// Layer 4: Business Logic - High-level operation orchestration
	recordingManager *RecordingManager // Stateless recording via MediaMTX API
	snapshotManager  *SnapshotManager  // Multi-tier snapshot capture (V4L2→FFmpeg→RTSP)

	// Configuration and Integration
	config            *config.MediaMTXConfig // MediaMTX-specific configuration
	configIntegration *ConfigIntegration     // Centralized configuration access pattern
	logger            *logging.Logger        // Structured logging with context

	// Health and Event Notification
	healthNotificationManager *HealthNotificationManager // Health event coordination
	eventNotifier             MediaMTXEventNotifier      // Real-time client notifications

	// Optional Components (may be nil based on configuration)
	externalDiscovery *ExternalStreamDiscovery // Optional: UAV/UGV stream discovery

	// Concurrency and State Management
	mu        sync.RWMutex    // Protects shared state during configuration changes
	isRunning int32           // Atomic boolean for thread-safe running state (0=false, 1=true)
	startTime time.Time       // Service startup timestamp for metrics
	ctx       context.Context // Root context for graceful shutdown coordination
	cancel    context.CancelFunc

	// Event-Driven Readiness System
	readinessEventChan chan struct{} // Readiness notification channel
	readinessMutex     sync.RWMutex  // Protects readiness event channel access
}

// checkRunningState safely checks if the controller is running using atomic operations.
// This prevents race conditions during concurrent access to the running state.
func (c *controller) checkRunningState() bool {
	return atomic.LoadInt32(&c.isRunning) == 1
}

// hasExternalDiscovery checks if external stream discovery is configured and available.
// Returns false if the component is nil (not configured) following the optional component pattern.
func (c *controller) hasExternalDiscovery() bool {
	return c.externalDiscovery != nil
}

// IsReady returns whether the controller is fully operational and ready to handle requests.
// This implements the progressive readiness pattern where the system becomes ready
// as components complete their initialization.
func (c *controller) IsReady() bool {
	if !c.checkRunningState() {
		c.logger.Debug("Controller not ready: not running")
		return false
	}

	// Camera monitor must complete at least one discovery cycle for hardware readiness
	if c.cameraMonitor != nil && !c.cameraMonitor.IsReady() {
		c.logger.Debug("Controller not ready: camera monitor not ready")
		return false
	}

	// Health monitor is optional - if nil, consider healthy by default following optional component pattern
	if c.healthMonitor != nil && !c.healthMonitor.IsHealthy() {
		c.logger.Debug("Controller not ready: health monitor not healthy")
		return false
	}

	c.logger.Debug("Controller is ready: all components ready")
	return true
}

// emitReadinessEvent emits a readiness event to all subscribers
func (c *controller) emitReadinessEvent() {
	c.readinessMutex.RLock()
	defer c.readinessMutex.RUnlock()

	// Send to the main readiness channel if it exists
	if c.readinessEventChan != nil {
		select {
		case c.readinessEventChan <- struct{}{}:
		default:
			// Channel is full, skip this event
		}
	}
}

// SubscribeToReadiness subscribes to controller readiness events
func (c *controller) SubscribeToReadiness() <-chan struct{} {
	c.readinessMutex.RLock()
	defer c.readinessMutex.RUnlock()
	return c.readinessEventChan
}

// GetReadinessState returns detailed readiness information
func (c *controller) GetReadinessState() map[string]interface{} {
	state := map[string]interface{}{
		"controller_running":     c.checkRunningState(),
		"camera_monitor_ready":   false,
		"health_monitor_healthy": false,
		"available_cameras":      []string{},
	}

	if c.cameraMonitor != nil {
		state["camera_monitor_ready"] = c.cameraMonitor.IsReady()
		state["camera_monitor_running"] = c.cameraMonitor.IsRunning()

		if c.cameraMonitor.IsReady() {
			cameras := c.cameraMonitor.GetConnectedCameras()
			cameraIDs := make([]string, 0, len(cameras))
			for devicePath := range cameras {
				// Convert device path to camera ID (camera0, camera1, etc.)
				if cameraID, exists := c.pathManager.GetCameraForDevicePath(devicePath); exists {
					cameraIDs = append(cameraIDs, cameraID)
				}
			}
			state["available_cameras"] = cameraIDs
		}
	}

	if c.healthMonitor != nil {
		state["health_monitor_healthy"] = c.healthMonitor.IsHealthy()
	}

	return state
}

// ControllerWithConfigManager creates a new MediaMTX controller with configuration integration
//
// Optional Components Pattern:
// Some components are optional based on configuration and may be nil:
// - externalDiscovery: Only if external stream sources are configured
// - pathIntegration: Only if auto-path creation is enabled
//
// All methods MUST check for nil before using optional components.
func ControllerWithConfigManager(configManager *config.ConfigManager, cameraMonitor camera.CameraMonitor, logger *logging.Logger) (MediaMTXController, error) {
	// Create configuration integration
	configIntegration := NewConfigIntegration(configManager, logger)

	// Get MediaMTX configuration
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaMTX configuration: %w", err)
	}

	// Create HTTP client
	client := NewClient(mediaMTXConfig.BaseURL, mediaMTXConfig, logger)

	// Configure MediaMTX paths if override is enabled
	logger.Info("Checking MediaMTX override configuration", "override_enabled", mediaMTXConfig.OverrideMediaMTXPaths, "recordings_path", mediaMTXConfig.RecordingsPath, "snapshots_path", mediaMTXConfig.SnapshotsPath, "config_source", "ControllerWithConfigManager")
	if mediaMTXConfig.OverrideMediaMTXPaths {
		// Get configuration for path generation
		cfg := configManager.GetConfig()
		if cfg != nil {
			// Generate MediaMTX record path pattern for recordings
			recordPath := GenerateRecordingPath(mediaMTXConfig, &cfg.Recording)

			// Generate snapshot path pattern for snapshots (for logging/debugging)
			snapshotPath := GenerateSnapshotPath(mediaMTXConfig, &cfg.Snapshots, "camera0") // Use generic device name for global config

			// Configure MediaMTX global settings via API
			globalConfig := map[string]interface{}{
				"recordPath":   recordPath,
				"recordFormat": cfg.Recording.RecordFormat,
				// MediaMTX handles snapshots per-path or via direct FFmpeg capture
				// Global snapshot configuration is not supported by MediaMTX API
			}

			// Convert to JSON for API call
			configData, err := json.Marshal(globalConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal MediaMTX global config: %w", err)
			}

			// Apply configuration to MediaMTX
			ctx := context.Background()
			err = client.Patch(ctx, MediaMTXConfigGlobalPatch, configData)
			if err != nil {
				logger.WithError(err).Warn("Failed to override MediaMTX paths - continuing with defaults")
				// Don't fail completely - MediaMTX will use its defaults
			} else {
				logger.WithFields(logging.Fields{
					"recordPath":   recordPath,
					"recordFormat": cfg.Recording.RecordFormat,
					"snapshotPath": snapshotPath,
				}).Info("MediaMTX paths overridden successfully")
			}
		}
	}

	// Create health monitor with configuration integration
	healthMonitor := NewHealthMonitor(client, mediaMTXConfig, configIntegration, logger)

	// Create path manager with camera monitor (consolidated camera operations)
	pathManager := NewPathManagerWithCamera(client, mediaMTXConfig, cameraMonitor, logger)

	// Create FFmpeg manager
	ffmpegManager := NewFFmpegManager(mediaMTXConfig, logger)

	// Get recording configuration
	cfg := configManager.GetConfig()
	recordingConfig := &cfg.Recording

	// Create stream manager with shared PathManager
	streamManager := NewStreamManager(client, pathManager, mediaMTXConfig, recordingConfig, configIntegration, logger)

	// Create recording manager (using existing client and pathManager)
	recordingManager := NewRecordingManager(client, pathManager, streamManager, ffmpegManager, mediaMTXConfig, recordingConfig, configIntegration, logger)

	// Create snapshot manager with configuration integration
	snapshotManager := NewSnapshotManagerWithConfig(ffmpegManager, streamManager, cameraMonitor, pathManager, mediaMTXConfig, configManager, logger)

	// Create RTSP connection manager
	rtspManager := NewRTSPConnectionManager(client, mediaMTXConfig, logger)

	// Create path integration (the missing link!)
	pathIntegration := NewPathIntegration(pathManager, cameraMonitor, configIntegration, logger)

	// Get full config for health notification manager
	fullConfig := configManager.GetConfig()
	if fullConfig == nil {
		return nil, fmt.Errorf("failed to get full configuration for health notification manager")
	}

	// Create health notification manager (will be connected to SystemEventNotifier later)
	healthNotificationManager := NewHealthNotificationManager(fullConfig, logger, nil)

	// Create system metrics manager
	systemMetricsManager := NewSystemMetricsManager(fullConfig, recordingConfig, configIntegration, logger)

	return &controller{
		client:                    client,
		healthMonitor:             healthMonitor,
		pathManager:               pathManager,
		pathIntegration:           pathIntegration,
		streamManager:             streamManager,
		ffmpegManager:             ffmpegManager,
		recordingManager:          recordingManager,
		snapshotManager:           snapshotManager,
		rtspManager:               rtspManager,
		cameraMonitor:             cameraMonitor,
		config:                    mediaMTXConfig,
		configIntegration:         configIntegration,
		logger:                    logger,
		healthNotificationManager: healthNotificationManager,
		systemMetricsManager:      systemMetricsManager,
		// externalDiscovery: nil - intentionally not initialized (optional component)
		// No local recording state - query MediaMTX directly
		readinessEventChan: make(chan struct{}, 10), // Buffered channel for readiness events
	}, nil
}

// Start starts the MediaMTX controller
func (c *controller) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.isRunning) == 1 {
		return fmt.Errorf("controller is already running")
	}

	// Create cancellable context for controller lifecycle management
	c.ctx, c.cancel = context.WithCancel(ctx)

	c.logger.Info("Starting MediaMTX controller")

	// Start health monitor
	if err := c.healthMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health monitor: %w", err)
	}

	// Wire SystemMetricsManager dependencies
	if c.systemMetricsManager != nil {
		c.systemMetricsManager.SetDependencies(c.recordingManager, c.cameraMonitor, c.streamManager)
	}

	// Start camera monitor with startup coordination
	if c.cameraMonitor != nil {
		// Check camera monitor running state to avoid duplicate starts
		if cameraMonitor, ok := c.cameraMonitor.(interface{ IsRunning() bool }); ok && cameraMonitor.IsRunning() {
			c.logger.Info("Camera monitor already running, skipping start")
		} else {
			if err := c.cameraMonitor.Start(ctx); err != nil {
				c.logger.WithError(err).Error("Failed to start camera monitor")
				return fmt.Errorf("failed to start camera monitor: %w", err)
			}
			c.logger.Info("Camera monitor started successfully")
		}
	}

	// Start path integration (connects cameras to MediaMTX paths)
	if c.pathIntegration != nil {
		if err := c.pathIntegration.Start(ctx); err != nil {
			c.logger.WithError(err).Error("Failed to start path integration")
			return fmt.Errorf("failed to start path integration: %w", err)
		}
		c.logger.Info("Path integration started successfully")
	}

	// Register Controller as camera event handler (CRITICAL MISSING WIRING)
	if c.cameraMonitor != nil {
		c.cameraMonitor.AddEventHandler(c) // Controller implements CameraEventHandler
		c.logger.Info("Controller registered as camera event handler")
	}

	atomic.StoreInt32(&c.isRunning, 1)
	c.startTime = time.Now()

	// Start readiness monitoring goroutine for Progressive Readiness pattern
	go c.monitorReadiness()

	c.logger.Info("MediaMTX controller started successfully")
	return nil
}

// monitorReadiness monitors controller readiness and emits events when ready
// This implements the Progressive Readiness pattern - components become ready as they initialize
func (c *controller) monitorReadiness() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastReadyState := false
	readyEventEmitted := false // Prevent duplicate events

	for {
		select {
		case <-ticker.C:
			if !c.checkRunningState() {
				// Controller stopped, exit monitoring
				return
			}

			currentReadyState := c.IsReady()

			// Emit readiness event only once when controller becomes ready
			if !lastReadyState && currentReadyState && !readyEventEmitted {
				c.logger.Info("Controller became ready - emitting readiness event")
				c.emitReadinessEvent()
				readyEventEmitted = true
			}

			// Reset if controller becomes unready (for recovery scenarios)
			if lastReadyState && !currentReadyState {
				readyEventEmitted = false
			}

			lastReadyState = currentReadyState

		case <-c.ctx.Done():
			// Context cancelled, exit gracefully
			return
		}
	}
}

// Stop stops the MediaMTX controller
func (c *controller) Stop(ctx context.Context) error {
	// Use atomic operation for thread-safe state check
	if !atomic.CompareAndSwapInt32(&c.isRunning, 1, 0) {
		return fmt.Errorf("controller is not running")
	}

	c.logger.Info("Stopping MediaMTX controller")

	// No need to track active recordings - MediaMTX manages its own state

	// Stop path integration first
	if c.pathIntegration != nil {
		if err := c.pathIntegration.Stop(ctx); err != nil {
			c.logger.WithError(err).Error("Failed to stop path integration")
		} else {
			c.logger.Info("Path integration stopped successfully")
		}
	}

	// Stop camera monitor
	if c.cameraMonitor != nil {
		if err := c.cameraMonitor.Stop(ctx); err != nil {
			c.logger.WithError(err).Error("Failed to stop camera monitor")
		} else {
			c.logger.Info("Camera monitor stopped successfully")
		}
	}

	// Stop health monitor
	if err := c.healthMonitor.Stop(ctx); err != nil {
		c.logger.WithError(err).Error("Failed to stop health monitor")
	}

	// Stop external discovery
	if c.hasExternalDiscovery() {
		if err := c.externalDiscovery.Stop(ctx); err != nil {
			c.logger.WithError(err).Error("Failed to stop external discovery")
		} else {
			c.logger.Info("External discovery stopped successfully")
		}
	}

	// Close HTTP client
	if err := c.client.Close(); err != nil {
		c.logger.WithError(err).Error("Failed to close HTTP client")
	}

	atomic.StoreInt32(&c.isRunning, 0)

	// Cancel controller context to stop readiness monitoring
	if c.cancel != nil {
		c.cancel()
	}

	c.logger.Info("MediaMTX controller stopped successfully")
	return nil
}

// GetHealth returns the current health status
func (c *controller) GetHealth(ctx context.Context) (*GetHealthResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to HealthMonitor - returns API-ready response with health status
	return c.healthMonitor.GetHealthAPI(ctx, c.startTime)
}

// GetMetrics returns the current metrics
func (c *controller) GetMetrics(ctx context.Context) (*GetMetricsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SystemMetricsManager - returns API-ready response with comprehensive metrics aggregation
	return c.systemMetricsManager.GetMetricsAPI(ctx)
}

// GetSystemMetrics returns comprehensive system performance metrics
func (c *controller) GetSystemMetrics(ctx context.Context) (*GetSystemMetricsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SystemMetricsManager - returns API-ready response with system calculations
	return c.systemMetricsManager.GetSystemMetricsAPI(ctx)
}

// GetServerInfo returns server information and capabilities
func (c *controller) GetServerInfo(ctx context.Context) (*GetServerInfoResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Check MediaMTX connection status
	mediaMTXStatus := "connected"
	if _, err := c.pathManager.ListPaths(ctx); err != nil {
		mediaMTXStatus = "disconnected"
	}

	// Get version from centralized configuration
	versionInfo := c.configIntegration.GetVersionInfo()

	// Build API-ready response
	response := &GetServerInfoResponse{
		ServiceName:   "MediaMTX Camera Service",
		Version:       versionInfo.Version,
		Status:        "running",
		StartTime:     c.startTime.Format(time.RFC3339),
		Uptime:        time.Since(c.startTime).String(),
		MediaMTXReady: mediaMTXStatus == "connected",
	}

	return response, nil
}

// CleanupOldFiles performs cleanup of old files based on retention policy using centralized configuration.
// Implements file lifecycle management for recording and snapshot storage with support for
// age-based, count-based, and size-based cleanup strategies.
//
// Pure delegation pattern: Controller orchestrates cleanup by delegating to manager methods
// and aggregating results for API response. All business logic resides in managers.
//
// Supports both retention policy types:
// - Age-based: Removes files older than MaxAgeDays and excess files beyond MaxCount
// - Size-based: Removes oldest files until total storage is under MaxSizeGB limit
//
// Returns accurate counts of removed recordings, snapshots, and space freed using
// centralized configuration from configIntegration.GetCleanupLimits().
func (c *controller) CleanupOldFiles(ctx context.Context) (*CleanupOldFilesResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Get current configuration
	cfg, err := c.configIntegration.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	// Check if retention policy is enabled
	if !cfg.RetentionPolicy.Enabled {
		return nil, fmt.Errorf("retention policy is not enabled")
	}

	// Perform cleanup based on retention policy
	var deletedCount int
	var totalSize int64

	// Get cleanup limits from centralized configuration
	maxAge, maxCount, maxSize, err := c.configIntegration.GetCleanupLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to get cleanup limits: %v", err)
	}

	// Pure delegation to manager methods - pass size parameter for both age and size-based cleanup
	var recordingsRemoved, snapshotsRemoved int
	var recordingsSpaceFreed, snapshotsSpaceFreed int64

	// Clean up recordings using extended manager method
	recordingsRemoved, recordingsSpaceFreed, err = c.recordingManager.CleanupOldRecordings(ctx, maxAge, maxCount, maxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup old recordings: %v", err)
	}

	// Clean up snapshots using extended manager method
	snapshotsRemoved, snapshotsSpaceFreed, err = c.snapshotManager.CleanupOldSnapshots(ctx, maxAge, maxCount, maxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup old snapshots: %v", err)
	}

	// Aggregate results from manager methods
	deletedCount = recordingsRemoved + snapshotsRemoved
	totalSize = recordingsSpaceFreed + snapshotsSpaceFreed

	// Build API-ready response using CleanupOldFilesResponse from rpc_types.go
	response := &CleanupOldFilesResponse{
		RecordingsRemoved: recordingsRemoved,
		SnapshotsRemoved:  snapshotsRemoved,
		SpaceFreed:        totalSize,
		Status:            "completed",
		Message:           fmt.Sprintf("Cleaned up %d files (%d recordings, %d snapshots), freed %d bytes", deletedCount, recordingsRemoved, snapshotsRemoved, totalSize),
	}
	return response, nil
}

// SetRetentionPolicy updates the retention policy configuration
func (c *controller) SetRetentionPolicy(ctx context.Context, enabled bool, policyType string, params map[string]interface{}) (*SetRetentionPolicyResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Get current configuration
	cfg, err := c.configIntegration.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	// Update retention policy configuration
	cfg.RetentionPolicy.Enabled = enabled
	cfg.RetentionPolicy.Type = policyType

	// Initialize response values
	maxAge := ""
	maxSize := ""

	// Update policy-specific parameters
	if policyType == "age" {
		if maxAgeDays, ok := params["max_age_days"].(float64); ok {
			cfg.RetentionPolicy.MaxAgeDays = int(maxAgeDays)
			maxAge = fmt.Sprintf("%d days", int(maxAgeDays))
		} else if maxAgeDays, ok := params["max_age_days"].(int); ok {
			cfg.RetentionPolicy.MaxAgeDays = maxAgeDays
			maxAge = fmt.Sprintf("%d days", maxAgeDays)
		}
	} else if policyType == "size" {
		if maxSizeGB, ok := params["max_size_gb"].(float64); ok {
			cfg.RetentionPolicy.MaxSizeGB = int(maxSizeGB)
			maxSize = fmt.Sprintf("%d GB", int(maxSizeGB))
		} else if maxSizeGB, ok := params["max_size_gb"].(int); ok {
			cfg.RetentionPolicy.MaxSizeGB = maxSizeGB
			maxSize = fmt.Sprintf("%d GB", maxSizeGB)
		}
	}

	// Build API-ready response
	response := &SetRetentionPolicyResponse{
		Success:    true,
		PolicyType: policyType,
		MaxAge:     maxAge,
		MaxSize:    maxSize,
		Message:    "Retention policy configuration updated successfully",
	}

	return response, nil
}

// SetSystemEventNotifier sets the system event notifier for health notifications
func (c *controller) SetSystemEventNotifier(notifier SystemEventNotifier) {
	if c.healthNotificationManager != nil {
		c.healthNotificationManager.systemNotifier = notifier
	}

	// Also set it on the health monitor
	if c.healthMonitor != nil {
		c.healthMonitor.SetSystemNotifier(notifier)
	}
}

// GetStorageInfo returns information about the storage space used by recordings and snapshots.
func (c *controller) GetStorageInfo(ctx context.Context) (*GetStorageInfoResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SystemMetricsManager - returns API-ready response with storage calculations
	return c.systemMetricsManager.GetStorageInfoAPI(ctx)
}

// GetStreams returns all streams using cameraID-first architecture
func (c *controller) GetStreams(ctx context.Context) (*GetStreamsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to StreamManager - no business logic in Controller!
	return c.streamManager.ListStreams(ctx)
}

// GetStream returns a specific stream
func (c *controller) GetStream(ctx context.Context, id string) (*Path, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.streamManager.GetStream(ctx, id)
}

// CreateStream creates a new stream
func (c *controller) CreateStream(ctx context.Context, name, source string) (*Path, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.streamManager.CreateStream(ctx, name, source)
}

// DeleteStream deletes a stream
func (c *controller) DeleteStream(ctx context.Context, id string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	return c.streamManager.DeleteStream(ctx, id)
}

// GetPaths returns all runtime paths
func (c *controller) GetPaths(ctx context.Context) ([]*Path, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Get runtime paths from path manager
	paths, err := c.pathManager.GetRuntimePaths(ctx)
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// GetPath returns a specific path
func (c *controller) GetPath(ctx context.Context, name string) (*Path, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.pathManager.GetPath(ctx, name)
}

// CreatePath creates a new path
func (c *controller) CreatePath(ctx context.Context, path *Path) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	// Build comprehensive path configuration using centralized config
	options, err := c.configIntegration.BuildPathConf(path.Name, path.Source, false)
	if err != nil {
		return fmt.Errorf("failed to build path configuration: %w", err)
	}

	// Build source URL using centralized configuration
	source, err := c.configIntegration.BuildSourceURL(path.Name, path.Source)
	if err != nil {
		return fmt.Errorf("failed to build source URL: %w", err)
	}

	return c.pathManager.CreatePath(ctx, path.Name, source, options)
}

// DeletePath deletes a path
func (c *controller) DeletePath(ctx context.Context, name string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	return c.pathManager.DeletePath(ctx, name)
}

// DiscoverExternalStreams discovers external streams
func (c *controller) DiscoverExternalStreams(ctx context.Context, options DiscoveryOptions) (*DiscoverExternalStreamsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	if !c.hasExternalDiscovery() {
		return nil, fmt.Errorf("external stream discovery is not configured")
	}

	// Pure delegation to ExternalStreamDiscovery - returns API-ready response
	return c.externalDiscovery.DiscoverExternalStreamsAPI(ctx, options)
}

// AddExternalStream adds an external stream to the system
func (c *controller) AddExternalStream(ctx context.Context, stream *ExternalStream) (*AddExternalStreamResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	if !c.hasExternalDiscovery() {
		return nil, fmt.Errorf("external stream discovery is not configured")
	}

	// Pure delegation to ExternalStreamDiscovery - returns API-ready response
	return c.externalDiscovery.AddExternalStreamAPI(ctx, stream)
}

// RemoveExternalStream removes an external stream from the system
func (c *controller) RemoveExternalStream(ctx context.Context, streamURL string) (*RemoveExternalStreamResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	if !c.hasExternalDiscovery() {
		return nil, fmt.Errorf("external stream discovery is not configured")
	}

	// Pure delegation to ExternalStreamDiscovery - returns API-ready response
	return c.externalDiscovery.RemoveExternalStreamAPI(ctx, streamURL)
}

// GetExternalStreams returns all discovered external streams
func (c *controller) GetExternalStreams(ctx context.Context) (*GetExternalStreamsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	if !c.hasExternalDiscovery() {
		// Return empty API-ready response for unconfigured external discovery
		return &GetExternalStreamsResponse{
			ExternalStreams: []ExternalStreamInfo{},
			SkydioStreams:   []ExternalStreamInfo{},
			GenericStreams:  []ExternalStreamInfo{},
			TotalCount:      0,
			Timestamp:       time.Now().Unix(),
		}, nil
	}

	// Pure delegation to ExternalStreamDiscovery - returns API-ready response
	return c.externalDiscovery.GetExternalStreamsAPI(ctx)
}

// StartRecording starts recording for a camera device
func (c *controller) StartRecording(ctx context.Context, cameraID string, options *PathConf) (*StartRecordingResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to RecordingManager - returns API-ready response with rich metadata
	return c.recordingManager.StartRecording(ctx, cameraID, options)
}

// StopRecording stops recording for a camera device
func (c *controller) StopRecording(ctx context.Context, cameraID string) (*StopRecordingResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to RecordingManager - returns API-ready response with actual metadata
	return c.recordingManager.StopRecording(ctx, cameraID)
}

// GetConfig returns the current configuration
func (c *controller) GetConfig(ctx context.Context) (*config.MediaMTXConfig, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.config, nil
}

// UpdateConfig updates the configuration
func (c *controller) UpdateConfig(ctx context.Context, config *config.MediaMTXConfig) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	// Validate new configuration
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Update configuration
	c.mu.Lock()
	c.config = config
	c.mu.Unlock()

	c.logger.Info("Configuration updated successfully")
	return nil
}

// GetRecordingManager returns the recording manager for cleanup operations
func (c *controller) GetRecordingManager() *RecordingManager {
	return c.recordingManager
}

// GetSnapshotManager returns the snapshot manager for cleanup operations
func (c *controller) GetSnapshotManager() *SnapshotManager {
	return c.snapshotManager
}

// validateConfig validates the MediaMTX configuration
func validateConfig(config *config.MediaMTXConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if config.RetryDelay <= 0 {
		return fmt.Errorf("retry delay must be positive")
	}

	// Validate circuit breaker configuration
	if config.CircuitBreaker.FailureThreshold <= 0 {
		return fmt.Errorf("circuit breaker failure threshold must be positive")
	}

	if config.CircuitBreaker.RecoveryTimeout <= 0 {
		return fmt.Errorf("circuit breaker recovery timeout must be positive")
	}

	if config.CircuitBreaker.MaxFailures <= 0 {
		return fmt.Errorf("circuit breaker max failures must be positive")
	}

	// Validate connection pool configuration
	if config.ConnectionPool.MaxIdleConns <= 0 {
		return fmt.Errorf("connection pool max idle connections must be positive")
	}

	if config.ConnectionPool.MaxIdleConnsPerHost <= 0 {
		return fmt.Errorf("connection pool max idle connections per host must be positive")
	}

	if config.ConnectionPool.IdleConnTimeout <= 0 {
		return fmt.Errorf("connection pool idle connection timeout must be positive")
	}

	return nil
}

// generateSnapshotID generates a unique snapshot ID
func generateSnapshotID(device string) string {
	return "snap_" + device + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// TakeAdvancedSnapshot takes a snapshot with advanced options
func (c *controller) TakeAdvancedSnapshot(ctx context.Context, cameraID string, options *SnapshotOptions) (*TakeSnapshotResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SnapshotManager - returns API-ready response with rich metadata
	return c.snapshotManager.TakeSnapshot(ctx, cameraID, options)
}

// GetAdvancedSnapshot gets a snapshot by ID
func (c *controller) GetAdvancedSnapshot(snapshotID string) (*Snapshot, bool) {
	if !c.checkRunningState() {
		return nil, false
	}
	return c.snapshotManager.GetSnapshot(snapshotID)
}

// All public API methods now consistently check running state
// Exceptions: internal helpers and event handlers (documented with NOTE comments)
// ListAdvancedSnapshots lists all snapshots
func (c *controller) ListAdvancedSnapshots() []*Snapshot {
	if !c.checkRunningState() {
		return []*Snapshot{}
	}
	return c.snapshotManager.ListSnapshotsInternal()
}

// DeleteAdvancedSnapshot deletes a snapshot
func (c *controller) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	return c.snapshotManager.DeleteSnapshot(ctx, snapshotID)
}

// GetSnapshotSettings gets current snapshot settings
func (c *controller) GetSnapshotSettings() *SnapshotSettings {
	if !c.checkRunningState() {
		return nil
	}
	return c.snapshotManager.GetSnapshotSettings()
}

// UpdateSnapshotSettings updates snapshot settings
func (c *controller) UpdateSnapshotSettings(settings *SnapshotSettings) {
	if !c.checkRunningState() {
		return
	}
	c.snapshotManager.UpdateSnapshotSettings(settings)
}

// ListRecordings lists recording files with metadata and pagination
func (c *controller) ListRecordings(ctx context.Context, limit, offset int) (*ListRecordingsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to RecordingManager - returns API-ready response with metadata extraction
	return c.recordingManager.ListRecordings(ctx, limit, offset)
}

// ListSnapshots lists snapshot files with metadata and pagination
func (c *controller) ListSnapshots(ctx context.Context, limit, offset int) (*ListSnapshotsResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SnapshotManager - returns API-ready response with metadata extraction
	return c.snapshotManager.ListSnapshots(ctx, limit, offset)
}

// GetRecordingInfo gets detailed information about a specific recording file
func (c *controller) GetRecordingInfo(ctx context.Context, filename string) (*GetRecordingInfoResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to RecordingManager - returns API-ready GetRecordingInfoResponse
	return c.recordingManager.GetRecordingInfo(ctx, filename)
}

// GetSnapshotInfo gets detailed information about a specific snapshot file
func (c *controller) GetSnapshotInfo(ctx context.Context, filename string) (*GetSnapshotInfoResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to SnapshotManager - returns API-ready GetSnapshotInfoResponse
	return c.snapshotManager.GetSnapshotInfo(ctx, filename)
}

// DeleteRecording deletes a recording file
func (c *controller) DeleteRecording(ctx context.Context, filename string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Deleting recording")

	return c.recordingManager.DeleteRecording(ctx, filename)
}

// DeleteSnapshot deletes a snapshot file
func (c *controller) DeleteSnapshot(ctx context.Context, filename string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Deleting snapshot")

	return c.snapshotManager.DeleteSnapshotFile(ctx, filename)
}

// RTSP Connection Management Methods

// ListRTSPConnections lists all RTSP connections
func (c *controller) ListRTSPConnections(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionList, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"page":         strconv.Itoa(page),
		"itemsPerPage": strconv.Itoa(itemsPerPage),
	}).Debug("Listing RTSP connections")

	return c.rtspManager.ListConnections(ctx, page, itemsPerPage)
}

// GetRTSPConnection gets a specific RTSP connection by ID
func (c *controller) GetRTSPConnection(ctx context.Context, id string) (*RTSPConnection, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Debug("Getting RTSP connection")

	return c.rtspManager.GetConnection(ctx, id)
}

// ListRTSPSessions lists all RTSP sessions
func (c *controller) ListRTSPSessions(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionSessionList, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"page":         strconv.Itoa(page),
		"itemsPerPage": strconv.Itoa(itemsPerPage),
	}).Debug("Listing RTSP sessions")

	return c.rtspManager.ListSessions(ctx, page, itemsPerPage)
}

// GetRTSPSession gets a specific RTSP session by ID
func (c *controller) GetRTSPSession(ctx context.Context, id string) (*RTSPConnectionSession, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Debug("Getting RTSP session")

	return c.rtspManager.GetSession(ctx, id)
}

// KickRTSPSession kicks out an RTSP session from the server
func (c *controller) KickRTSPSession(ctx context.Context, id string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Info("Kicking RTSP session")

	return c.rtspManager.KickSession(ctx, id)
}

// GetRTSPConnectionHealth returns the health status of RTSP connections
func (c *controller) GetRTSPConnectionHealth(ctx context.Context) (*HealthStatus, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.rtspManager.GetConnectionHealth(ctx)
}

// GetRTSPConnectionMetrics returns metrics about RTSP connections
func (c *controller) GetRTSPConnectionMetrics(ctx context.Context) map[string]interface{} {
	if !c.checkRunningState() {
		return map[string]interface{}{
			"error": "controller is not running",
		}
	}

	return c.rtspManager.GetConnectionMetrics(ctx)
}

// StartStreaming starts a live streaming session
func (c *controller) StartStreaming(ctx context.Context, cameraID string) (*GetStreamURLResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to StreamManager - returns API-ready response
	return c.streamManager.StartStream(ctx, cameraID)
}

// StopStreaming stops the streaming session for the specified device
func (c *controller) StopStreaming(ctx context.Context, cameraID string) error {
	if !c.checkRunningState() {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"action":   "stop_streaming",
	}).Info("Stopping streaming session")

	// Use StreamManager to stop viewing stream
	err := c.streamManager.StopStream(ctx, cameraID)
	if err != nil {
		c.logger.WithFields(logging.Fields{
			"cameraID": cameraID,
			"error":    err.Error(),
		}).Error("Failed to stop streaming")
		return fmt.Errorf("failed to stop streaming: %w", err)
	}

	c.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
	}).Info("Streaming session stopped successfully")

	return nil
}

// GetStreamURL returns the stream URL for the specified device
func (c *controller) GetStreamURL(ctx context.Context, device string) (*GetStreamURLResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to StreamManager - consolidates URL generation and status checking
	return c.streamManager.GetStreamURL(ctx, device)
}

// GetStreamStatus returns the status of the streaming session for the specified device
func (c *controller) GetStreamStatus(ctx context.Context, device string) (*GetStreamStatusResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Pure delegation to StreamManager - already returns complete GetStreamStatusResponse
	return c.streamManager.GetStreamStatus(ctx, device)
}

// GetCameraList returns a list of all discovered cameras with their current status
func (c *controller) GetCameraList(ctx context.Context) (*CameraListResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller not running")
	}

	// Delegate to PathManager (returns API-ready format)
	response, err := c.pathManager.GetCameraList(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get camera list from path manager")
		return nil, fmt.Errorf("failed to get camera list: %w", err)
	}

	c.logger.WithFields(logging.Fields{
		"total":     response.Total,
		"connected": response.Connected,
	}).Info("Retrieved API-ready camera list through PathManager")

	return response, nil
}

// GetCameraStatus returns the status for a specific camera device
func (c *controller) GetCameraStatus(ctx context.Context, device string) (*GetCameraStatusResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller not running")
	}

	// Validate device parameter
	if device == "" {
		return nil, fmt.Errorf("device parameter is required")
	}

	// Delegate to PathManager (consolidates camera operations and abstraction layer)
	response, err := c.pathManager.GetCameraStatus(ctx, device)
	if err != nil {
		c.logger.WithFields(logging.Fields{"device": device}).WithError(err).Error("Failed to get camera status from path manager")
		return nil, fmt.Errorf("camera device not found: %s", device)
	}

	c.logger.WithFields(logging.Fields{
		"device": device,
		"status": response.Status,
		"name":   response.Name,
	}).Info("Retrieved camera status through PathManager")

	return response, nil
}

// GetCameraCapabilities returns detailed capabilities for a specific camera device
func (c *controller) GetCameraCapabilities(ctx context.Context, device string) (*GetCameraCapabilitiesResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller not running")
	}

	// Validate device parameter
	if device == "" {
		return nil, fmt.Errorf("device parameter is required")
	}

	// Delegate to PathManager (consolidates camera operations and abstraction layer)
	response, err := c.pathManager.GetCameraCapabilities(ctx, device)
	if err != nil {
		c.logger.WithFields(logging.Fields{"device": device}).WithError(err).Error("Failed to get camera capabilities from path manager")
		return nil, fmt.Errorf("camera device capabilities not available: %s", device)
	}

	c.logger.WithFields(logging.Fields{
		"device":            device,
		"formats_count":     len(response.Formats),
		"fps_options_count": len(response.FrameRates),
	}).Info("Retrieved camera capabilities through PathManager")

	return response, nil
}

// GetCameraForDevicePath converts device path to camera identifier (for event abstraction)
func (c *controller) GetCameraForDevicePath(devicePath string) (string, bool) {
	if !c.checkRunningState() {
		return "", false
	}
	return c.pathManager.GetCameraForDevicePath(devicePath)
}

// GetDevicePathForCamera converts camera identifier to device path (for event abstraction)
func (c *controller) GetDevicePathForCamera(cameraID string) (string, bool) {
	if !c.checkRunningState() {
		return "", false
	}
	return c.pathManager.GetDevicePathForCamera(cameraID)
}

// ValidateCameraDevice validates that a camera device exists and is accessible
func (c *controller) ValidateCameraDevice(ctx context.Context, device string) (bool, error) {
	if !c.checkRunningState() {
		return false, fmt.Errorf("controller not running")
	}

	// Validate device parameter
	if device == "" {
		return false, fmt.Errorf("device parameter is required")
	}

	// Delegate to PathManager (consolidates camera operations and abstraction layer)
	exists, err := c.pathManager.ValidateCameraDevice(ctx, device)
	if err != nil {
		c.logger.WithFields(logging.Fields{"device": device}).WithError(err).Error("Failed to validate camera device through path manager")
		return false, err
	}

	c.logger.WithFields(logging.Fields{
		"device": device,
		"exists": exists,
	}).Info("Device validation through PathManager")

	return exists, nil
}

// GetHealthMonitor returns the health monitor instance for threshold notifications
// NOTE: No running state check - used internally by health system during startup/shutdown
func (c *controller) GetHealthMonitor() HealthMonitor {
	return c.healthMonitor
}

// SetDiscoveryInterval sets the external discovery scan interval dynamically
func (c *controller) SetDiscoveryInterval(interval int) (*SetDiscoveryIntervalResponse, error) {
	if !c.checkRunningState() {
		return nil, fmt.Errorf("controller is not running")
	}

	// Validate interval (0 = on-demand only, >0 = periodic scanning)
	if interval < 0 {
		return nil, fmt.Errorf("scan_interval must be >= 0")
	}

	// Check if external discovery is available (optional component)
	if !c.hasExternalDiscovery() {
		return nil, fmt.Errorf("external discovery not configured")
	}

	// Update the discovery service scan interval dynamically
	// This updates the running ticker without requiring restart
	err := c.externalDiscovery.UpdateScanInterval(interval)
	if err != nil {
		return nil, fmt.Errorf("failed to update discovery interval: %w", err)
	}

	// Build API-ready response
	response := &SetDiscoveryIntervalResponse{
		ScanInterval: interval,
		Status:       "updated",
		Message:      "Discovery interval updated successfully",
		Timestamp:    time.Now().Unix(),
	}

	c.logger.WithFields(logging.Fields{
		"old_interval": "unknown", // TODO: Get previous interval from discovery service
		"new_interval": interval,
	}).Info("Discovery scan interval updated successfully")

	return response, nil
}

// calculateCPUUsage calculates current CPU usage percentage
// NOTE: No running state check - internal helper function
func (c *controller) calculateCPUUsage() float64 {
	// Use gopsutil for accurate CPU usage calculation
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to get CPU usage, falling back to placeholder")
		return 0.0 // Return 0 instead of GC-based calculation
	}

	if len(percentages) == 0 {
		return 0.0
	}

	return percentages[0]
}

// calculateDiskUsage calculates current disk usage percentage
// NOTE: No running state check - internal helper function
func (c *controller) calculateDiskUsage() float64 {
	// Use gopsutil for accurate disk usage calculation
	// Get usage for the root filesystem where recordings are typically stored
	usage, err := disk.Usage("/")
	if err != nil {
		// Try alternative paths if root fails
		usage, err = disk.Usage(".")
		if err != nil {
			c.logger.WithError(err).Warn("Failed to get disk usage, falling back to placeholder")
			return 0.0 // TODO: Implement proper CPU usage calculation when stats unavailable
		}
	}

	// Calculate percentage: (used / total) * 100
	if usage.Total == 0 {
		return 0.0
	}

	percentUsed := float64(usage.Used) / float64(usage.Total) * 100.0
	return percentUsed
}

// OnCameraDisconnected handles camera disconnection events
// This is called by the camera monitor when a USB camera is unplugged
// NOTE: No running state check - event handler called during system lifecycle
func (c *controller) OnCameraDisconnected(devicePath string) {
	// Convert device path to camera identifier using existing utilities
	cameraID, exists := c.pathManager.GetCameraForDevicePath(devicePath)
	if !exists {
		return // Not a managed camera, ignore
	}

	// Check if this camera was recording using existing RecordingManager state
	if c.recordingManager.IsRecording(cameraID) {
		c.logger.WithFields(logging.Fields{
			"cameraID":    cameraID,
			"device_path": devicePath,
		}).Warn("Camera disconnected during recording - notifying failure")

		// Emit recording failure event for immediate UX feedback
		if c.eventNotifier != nil {
			c.eventNotifier.NotifyRecordingFailed(cameraID, "device_disconnected")
		}

		// Clean up recording state (stop timer, cleanup)
		c.recordingManager.forceStopRecording(cameraID)
	} // Close the if c.recordingManager.IsRecording(cameraID) block

	// Update all statuses due to camera disconnection
	// 1. Camera list status → disconnected (handled by camera monitor)
	// 2. Recording status → inactive (handled above)
	// 3. Stream status → inactive (handled by PathManager.DeletePath() on camera disconnect)
	// 4. WebSocket clients → real-time notification (handled by NotifyRecordingFailed)

	// ARCHITECTURE NOTE: Stream cleanup and camera notifications are handled separately
	// Stream cleanup: Handled by PathManager.DeletePath() when camera disconnects
	// Camera notifications: Require event system architecture review (not simple TODO)
	// See ARCHITECTURE.md for component responsibilities and event flow
}

// HandleCameraEvent implements camera.CameraEventHandler interface
// This is called by the camera monitor for all camera events
// NOTE: No running state check - event handler called during system lifecycle
func (c *controller) HandleCameraEvent(ctx context.Context, eventData camera.CameraEventData) error {
	switch eventData.EventType {
	case camera.CameraEventDisconnected:
		c.OnCameraDisconnected(eventData.DevicePath)
	case camera.CameraEventConnected:
		// ARCHITECTURE DECISION: Camera connected events logged only (no action needed)
		// RATIONALE: Camera discovery handled by CameraMonitor, paths created on-demand
		c.logger.WithField("device_path", eventData.DevicePath).Info("Camera connected")
	case camera.CameraEventStatusChanged:
		// ARCHITECTURE DECISION: Camera status changes logged only (no action needed)
		// RATIONALE: Status changes handled by individual operations (recording, snapshot)
		c.logger.WithField("device_path", eventData.DevicePath).Info("Camera status changed")
	}
	return nil
}

// File management methods are implemented and wired to RecordingManager and SnapshotManager
