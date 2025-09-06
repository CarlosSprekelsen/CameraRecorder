/*
MediaMTX Controller Implementation

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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// controller represents the main MediaMTX controller
type controller struct {
	client           MediaMTXClient
	healthMonitor    HealthMonitor
	pathManager      PathManager
	streamManager    StreamManager
	ffmpegManager    FFmpegManager
	recordingManager *RecordingManager
	snapshotManager  *SnapshotManager
	rtspManager      RTSPConnectionManager
	config           *MediaMTXConfig
	logger           *logging.Logger

	// State management
	mu        sync.RWMutex
	isRunning bool
	startTime time.Time

	// Recording sessions
	sessions   map[string]*RecordingSession
	sessionsMu sync.RWMutex

	// Active recording tracking (Phase 2 enhancement)
	activeRecordings map[string]*ActiveRecording
	recordingMutex   sync.RWMutex
}

// Abstraction layer mapping functions
// These functions handle the conversion between camera identifiers (camera0, camera1)
// and device paths (/dev/video0, /dev/video1) to maintain proper API abstraction

// getCameraIdentifierFromDevicePath converts a device path to a camera identifier
// Example: /dev/video0 -> camera0
func (c *controller) getCameraIdentifierFromDevicePath(devicePath string) string {
	// Extract the number from /dev/video{N}
	if strings.HasPrefix(devicePath, "/dev/video") {
		number := strings.TrimPrefix(devicePath, "/dev/video")
		return "camera" + number
	}
	// If it's already a camera identifier, return as is
	if strings.HasPrefix(devicePath, "camera") {
		return devicePath
	}
	// Fallback: return the original path
	return devicePath
}

// getDevicePathFromCameraIdentifier converts a camera identifier to a device path
// Example: camera0 -> /dev/video0
func (c *controller) getDevicePathFromCameraIdentifier(cameraID string) string {
	// Extract the number from camera{N}
	if strings.HasPrefix(cameraID, "camera") {
		number := strings.TrimPrefix(cameraID, "camera")
		return "/dev/video" + number
	}
	// If it's already a device path, return as is
	if strings.HasPrefix(cameraID, "/dev/video") {
		return cameraID
	}
	// Fallback: return the original identifier
	return cameraID
}

// validateCameraIdentifier validates that a camera identifier follows the correct pattern
func (c *controller) validateCameraIdentifier(cameraID string) bool {
	// Must match pattern camera[0-9]+
	matched, _ := regexp.MatchString(`^camera[0-9]+$`, cameraID)
	return matched
}

// Active recording management methods (Phase 2 enhancement)

// IsDeviceRecording checks if a device is currently recording
func (c *controller) IsDeviceRecording(devicePath string) bool {
	c.recordingMutex.RLock()
	defer c.recordingMutex.RUnlock()

	// Abstraction layer: Convert camera identifier to device path if needed
	var actualDevicePath string
	if c.validateCameraIdentifier(devicePath) {
		actualDevicePath = c.getDevicePathFromCameraIdentifier(devicePath)
	} else {
		actualDevicePath = devicePath
	}

	_, exists := c.activeRecordings[actualDevicePath]
	return exists
}

// StartActiveRecording starts tracking an active recording session
func (c *controller) StartActiveRecording(devicePath, sessionID, streamName string) error {
	c.recordingMutex.Lock()
	defer c.recordingMutex.Unlock()

	// Abstraction layer: Convert camera identifier to device path if needed
	var actualDevicePath string
	if c.validateCameraIdentifier(devicePath) {
		actualDevicePath = c.getDevicePathFromCameraIdentifier(devicePath)
	} else {
		actualDevicePath = devicePath
	}

	// Check for existing recording
	if _, exists := c.activeRecordings[actualDevicePath]; exists {
		return fmt.Errorf("device %s is already recording", devicePath)
	}

	// Create active recording entry
	c.activeRecordings[actualDevicePath] = &ActiveRecording{
		SessionID:  sessionID,
		DevicePath: devicePath, // Store camera identifier for API consistency
		StartTime:  time.Now(),
		StreamName: streamName,
		Status:     "RECORDING",
	}

	c.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"session_id":  sessionID,
		"stream_name": streamName,
	}).Info("Active recording started")

	return nil
}

// StopActiveRecording stops tracking an active recording session
func (c *controller) StopActiveRecording(devicePath string) error {
	c.recordingMutex.Lock()
	defer c.recordingMutex.Unlock()

	// Abstraction layer: Convert camera identifier to device path if needed
	var actualDevicePath string
	if c.validateCameraIdentifier(devicePath) {
		actualDevicePath = c.getDevicePathFromCameraIdentifier(devicePath)
	} else {
		actualDevicePath = devicePath
	}

	recording, exists := c.activeRecordings[actualDevicePath]
	if !exists {
		return fmt.Errorf("no active recording found for device %s", devicePath)
	}

	// Update status and remove from active recordings
	recording.Status = "STOPPED"
	delete(c.activeRecordings, actualDevicePath)

	c.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"session_id":  recording.SessionID,
		"duration":    time.Since(recording.StartTime),
	}).Info("Active recording stopped")

	return nil
}

// GetActiveRecordings returns all active recording sessions
func (c *controller) GetActiveRecordings() map[string]*ActiveRecording {
	c.recordingMutex.RLock()
	defer c.recordingMutex.RUnlock()

	// Return a copy to avoid race conditions
	activeRecordings := make(map[string]*ActiveRecording)
	for devicePath, recording := range c.activeRecordings {
		// Convert device path back to camera identifier for API consistency
		cameraID := c.getCameraIdentifierFromDevicePath(devicePath)
		activeRecordings[cameraID] = &ActiveRecording{
			SessionID:  recording.SessionID,
			DevicePath: cameraID, // Return camera identifier for API consistency
			StartTime:  recording.StartTime,
			StreamName: recording.StreamName,
			Status:     recording.Status,
		}
	}

	return activeRecordings
}

// GetActiveRecording gets active recording details for a device
func (c *controller) GetActiveRecording(devicePath string) *ActiveRecording {
	c.recordingMutex.RLock()
	defer c.recordingMutex.RUnlock()

	// Abstraction layer: Convert camera identifier to device path if needed
	var actualDevicePath string
	if c.validateCameraIdentifier(devicePath) {
		actualDevicePath = c.getDevicePathFromCameraIdentifier(devicePath)
	} else {
		actualDevicePath = devicePath
	}

	return c.activeRecordings[actualDevicePath]
}

// ControllerWithConfigManager creates a new MediaMTX controller with configuration integration
func ControllerWithConfigManager(configManager *config.ConfigManager, logger *logging.Logger) (MediaMTXController, error) {
	// Create configuration integration
	configIntegration := NewConfigIntegration(configManager, logger)

	// Get MediaMTX configuration
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaMTX configuration: %w", err)
	}

	// Create HTTP client
	client := NewClient(mediaMTXConfig.BaseURL, mediaMTXConfig, logger)

	// Create health monitor
	healthMonitor := NewHealthMonitor(client, mediaMTXConfig, logger)

	// Create path manager
	pathManager := NewPathManager(client, mediaMTXConfig, logger)

	// Create FFmpeg manager
	ffmpegManager := NewFFmpegManager(mediaMTXConfig, logger)

	// Create stream manager
	streamManager := NewStreamManager(client, mediaMTXConfig, logger)

	// Create recording manager (using existing client and pathManager)
	recordingManager := NewRecordingManager(client, pathManager, streamManager, mediaMTXConfig, logger)

	// Create snapshot manager with configuration integration
	snapshotManager := NewSnapshotManagerWithConfig(ffmpegManager, streamManager, mediaMTXConfig, configManager, logger)

	// Create RTSP connection manager
	rtspManager := NewRTSPConnectionManager(client, mediaMTXConfig, logger)

	return &controller{
		client:           client,
		healthMonitor:    healthMonitor,
		pathManager:      pathManager,
		streamManager:    streamManager,
		ffmpegManager:    ffmpegManager,
		recordingManager: recordingManager,
		snapshotManager:  snapshotManager,
		rtspManager:      rtspManager,
		config:           mediaMTXConfig,
		logger:           logger,
		sessions:         make(map[string]*RecordingSession),
		activeRecordings: make(map[string]*ActiveRecording),
	}, nil
}

// Start starts the MediaMTX controller
func (c *controller) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("controller is already running")
	}

	c.logger.Info("Starting MediaMTX controller")

	// Start health monitor
	if err := c.healthMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health monitor: %w", err)
	}

	c.isRunning = true
	c.startTime = time.Now()

	c.logger.Info("MediaMTX controller started successfully")
	return nil
}

// Stop stops the MediaMTX controller
func (c *controller) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.Info("Stopping MediaMTX controller")

	// Stop all recording sessions
	c.sessionsMu.Lock()
	for sessionID, session := range c.sessions {
		if session.Status == "RECORDING" {
			c.logger.WithField("session_id", sessionID).Info("Stopping recording session")
			if err := c.stopRecordingInternal(ctx, sessionID); err != nil {
				c.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to stop recording session")
			}
		}
	}
	c.sessionsMu.Unlock()

	// Stop health monitor
	if err := c.healthMonitor.Stop(ctx); err != nil {
		c.logger.WithError(err).Error("Failed to stop health monitor")
	}

	// Close HTTP client
	if err := c.client.Close(); err != nil {
		c.logger.WithError(err).Error("Failed to close HTTP client")
	}

	c.isRunning = false

	c.logger.Info("MediaMTX controller stopped successfully")
	return nil
}

// GetHealth returns the current health status
func (c *controller) GetHealth(ctx context.Context) (*HealthStatus, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	status := c.healthMonitor.GetStatus()
	return &status, nil
}

// GetMetrics returns the current metrics
func (c *controller) GetMetrics(ctx context.Context) (*Metrics, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Get streams for metrics
	streams, err := c.GetStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams for metrics: %w", err)
	}

	// Calculate metrics
	activeStreams := 0
	for _, stream := range streams {
		if stream.Ready {
			activeStreams++
		}
	}

	// Get health status for additional metrics
	healthStatus := c.healthMonitor.GetStatus()

	metrics := &Metrics{
		ActiveStreams: activeStreams,
		TotalStreams:  len(streams),
		CPUUsage:      0.0, // Would need system metrics
		MemoryUsage:   0.0, // Would need system metrics
		Uptime:        int64(time.Since(c.startTime).Seconds()),
	}

	// Add health metrics if available
	if healthStatus.Metrics.ActiveStreams > 0 {
		metrics.ActiveStreams = healthStatus.Metrics.ActiveStreams
		metrics.TotalStreams = healthStatus.Metrics.TotalStreams
		metrics.CPUUsage = healthStatus.Metrics.CPUUsage
		metrics.MemoryUsage = healthStatus.Metrics.MemoryUsage
		metrics.Uptime = healthStatus.Metrics.Uptime
	}

	return metrics, nil
}

// GetSystemMetrics returns comprehensive system performance metrics
// Following Python PerformanceMetrics.get_metrics() implementation
func (c *controller) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Get enhanced health monitor metrics (Phase 1 enhancement)
	healthMetrics := c.healthMonitor.GetMetrics()

	// Get health status for component information
	healthStatus := c.healthMonitor.GetStatus()

	// Calculate component status
	componentStatus := make(map[string]string)
	componentStatus["mediamtx_controller"] = "running"
	componentStatus["health_monitor"] = healthStatus.Status
	componentStatus["path_manager"] = "running"
	componentStatus["stream_manager"] = "running"
	componentStatus["recording_manager"] = "running"
	componentStatus["snapshot_manager"] = "running"

	// Get RTSP connection health
	rtspHealth, err := c.rtspManager.GetConnectionHealth(ctx)
	if err != nil {
		componentStatus["rtsp_connection_manager"] = "error"
		c.logger.WithError(err).Error("Failed to get RTSP connection health")
	} else {
		componentStatus["rtsp_connection_manager"] = rtspHealth.Status
	}

	// Simplified error counts - only track basic failure count
	errorCounts := make(map[string]int64)
	if failureCount, ok := healthMetrics["failure_count"].(int); ok {
		errorCounts["health_check"] = int64(failureCount)
	}

	// Get circuit breaker state - simplified version
	circuitBreakerState := "CLOSED"
	if isHealthy, ok := healthMetrics["is_healthy"].(bool); ok && !isHealthy {
		circuitBreakerState = "OPEN"
	}

	// Calculate response time (average from health metrics) - simplified version
	responseTime := 0.0
	if lastCheck, ok := healthMetrics["last_check"].(time.Time); ok {
		responseTime = float64(time.Since(lastCheck).Milliseconds())
	}

	// Get RTSP connection metrics
	rtspMetrics := c.rtspManager.GetConnectionMetrics(ctx)

	// Add RTSP connection count to active connections
	activeConnections := 0
	if rtspConnections, ok := rtspMetrics["total_connections"].(int); ok {
		activeConnections = rtspConnections
	}

	// Add RTSP-specific error counts
	if rtspConnections, ok := rtspMetrics["total_connections"].(int); ok && rtspConnections > c.config.RTSPMonitoring.MaxConnections {
		errorCounts["rtsp_connection_limit"] = int64(rtspConnections - c.config.RTSPMonitoring.MaxConnections)
	}

	systemMetrics := &SystemMetrics{
		RequestCount:        0, // Will be populated by WebSocket server
		ResponseTime:        responseTime,
		ErrorCount:          int64(healthMetrics["failure_count"].(int)),
		ActiveConnections:   int64(activeConnections),
		ComponentStatus:     componentStatus,
		ErrorCounts:         errorCounts,
		LastCheck:           healthStatus.LastCheck,
		CircuitBreakerState: circuitBreakerState,
	}

	return systemMetrics, nil
}

// GetStreams returns all streams
func (c *controller) GetStreams(ctx context.Context) ([]*Stream, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.streamManager.ListStreams(ctx)
}

// GetStream returns a specific stream
func (c *controller) GetStream(ctx context.Context, id string) (*Stream, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.streamManager.GetStream(ctx, id)
}

// CreateStream creates a new stream
func (c *controller) CreateStream(ctx context.Context, name, source string) (*Stream, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.streamManager.CreateStream(ctx, name, source)
}

// DeleteStream deletes a stream
func (c *controller) DeleteStream(ctx context.Context, id string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	return c.streamManager.DeleteStream(ctx, id)
}

// GetPaths returns all paths
func (c *controller) GetPaths(ctx context.Context) ([]*Path, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.pathManager.ListPaths(ctx)
}

// GetPath returns a specific path
func (c *controller) GetPath(ctx context.Context, name string) (*Path, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.pathManager.GetPath(ctx, name)
}

// CreatePath creates a new path
func (c *controller) CreatePath(ctx context.Context, path *Path) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	options := make(map[string]interface{})

	// Convert path fields to options
	if path.SourceOnDemand {
		options["sourceOnDemand"] = path.SourceOnDemand
	}
	if path.SourceOnDemandStartTimeout > 0 {
		options["sourceOnDemandStartTimeout"] = path.SourceOnDemandStartTimeout.String()
	}
	if path.SourceOnDemandCloseAfter > 0 {
		options["sourceOnDemandCloseAfter"] = path.SourceOnDemandCloseAfter.String()
	}
	if path.PublishUser != "" {
		options["publishUser"] = path.PublishUser
	}
	if path.PublishPass != "" {
		options["publishPass"] = path.PublishPass
	}
	if path.ReadUser != "" {
		options["readUser"] = path.ReadUser
	}
	if path.ReadPass != "" {
		options["readPass"] = path.ReadPass
	}
	if path.RunOnDemand != "" {
		options["runOnDemand"] = path.RunOnDemand
	}
	if path.RunOnDemandRestart {
		options["runOnDemandRestart"] = path.RunOnDemandRestart
	}
	if path.RunOnDemandCloseAfter > 0 {
		options["runOnDemandCloseAfter"] = path.RunOnDemandCloseAfter.String()
	}
	if path.RunOnDemandStartTimeout > 0 {
		options["runOnDemandStartTimeout"] = path.RunOnDemandStartTimeout.String()
	}

	return c.pathManager.CreatePath(ctx, path.Name, path.Source, options)
}

// DeletePath deletes a path
func (c *controller) DeletePath(ctx context.Context, name string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	return c.pathManager.DeletePath(ctx, name)
}

// StartRecording starts a recording session
func (c *controller) StartRecording(ctx context.Context, device, path string) (*RecordingSession, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Validate input parameters for security
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}
	if path == "" {
		return nil, fmt.Errorf("recording path is required")
	}

	// Abstraction layer: Convert camera identifier to device path if needed
	var devicePath string
	var cameraID string

	if c.validateCameraIdentifier(device) {
		// Device is a camera identifier (e.g., "camera0")
		cameraID = device
		devicePath = c.getDevicePathFromCameraIdentifier(device)
		c.logger.WithFields(logging.Fields{
			"camera_id":   cameraID,
			"device_path": devicePath,
		}).Debug("Converted camera identifier to device path")
	} else {
		// Device is already a device path (e.g., "/dev/video0")
		devicePath = device
		cameraID = c.getCameraIdentifierFromDevicePath(device)
	}

	// Generate session ID
	sessionID := generateSessionID(devicePath)

	// Check if session already exists
	c.sessionsMu.Lock()
	if _, exists := c.sessions[sessionID]; exists {
		c.sessionsMu.Unlock()
		return nil, NewRecordingError(sessionID, devicePath, "start_recording", "session already exists")
	}
	c.sessionsMu.Unlock()

	// Create recording session
	session := &RecordingSession{
		ID:        sessionID,
		Device:    cameraID, // Store camera identifier for API consistency
		Path:      path,
		Status:    "STARTING",
		StartTime: time.Now(),
		FilePath:  generateRecordingPath(devicePath, sessionID),
	}

	// Use MediaMTX RecordingManager for recording (no FFmpeg)
	options := map[string]interface{}{
		"format":  "mp4",
		"codec":   "h264",
		"quality": "medium",
	}

	recordingSession, err := c.recordingManager.StartRecording(ctx, devicePath, session.FilePath, options)
	if err != nil {
		return nil, NewRecordingErrorWithErr(sessionID, devicePath, "start_recording", "failed to start MediaMTX recording", err)
	}

	// Update session with MediaMTX recording info
	session.Status = "RECORDING"
	session.PID = recordingSession.PID // MediaMTX session ID
	session.Path = recordingSession.Path

	// Store session
	c.sessionsMu.Lock()
	c.sessions[sessionID] = session
	c.sessionsMu.Unlock()

	// Start tracking active recording for API consistency
	if err := c.StartActiveRecording(cameraID, sessionID, ""); err != nil {
		c.logger.WithError(err).WithField("session_id", sessionID).Warning("Failed to start active recording tracking")
	}

	c.logger.WithFields(logging.Fields{
		"session_id":       sessionID,
		"device":           cameraID,
		"device_path":      devicePath,
		"path":             path,
		"mediamtx_session": recordingSession.ID,
	}).Info("MediaMTX recording session started")

	return session, nil
}

// StopRecording stops a recording session
func (c *controller) StopRecording(ctx context.Context, sessionID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	// Validate input parameters for security
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	return c.stopRecordingInternal(ctx, sessionID)
}

// stopRecordingInternal stops a recording session (internal method)
func (c *controller) stopRecordingInternal(ctx context.Context, sessionID string) error {
	c.sessionsMu.Lock()
	session, exists := c.sessions[sessionID]
	if !exists {
		c.sessionsMu.Unlock()
		return NewRecordingError(sessionID, "", "stop_recording", "session not found")
	}

	if session.Status != "RECORDING" {
		c.sessionsMu.Unlock()
		return NewRecordingError(sessionID, session.Device, "stop_recording", "session is not recording")
	}

	// Update session status
	session.Status = "STOPPING"
	c.sessionsMu.Unlock()

	// Stop MediaMTX recording using RecordingManager
	if err := c.recordingManager.StopRecording(ctx, sessionID); err != nil {
		c.logger.WithError(err).WithFields(logging.Fields{
			"session_id": sessionID,
		}).Warning("Failed to stop MediaMTX recording")
	}

	// Update session status
	c.sessionsMu.Lock()
	session.Status = "STOPPED"
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)

	// Get file size (MediaMTX handles file management)
	if fileSize, _, err := c.ffmpegManager.GetFileInfo(ctx, session.FilePath); err == nil {
		session.FileSize = fileSize
	}

	c.sessionsMu.Unlock()

	// Stop tracking active recording for API consistency
	if err := c.StopActiveRecording(session.Device); err != nil {
		c.logger.WithError(err).WithField("session_id", sessionID).Warning("Failed to stop active recording tracking")
	}

	c.logger.WithFields(logging.Fields{
		"session_id": sessionID,
		"device":     session.Device,
		"duration":   session.Duration,
		"file_size":  session.FileSize,
	}).Info("Recording session stopped")

	return nil
}

// TakeSnapshot takes a snapshot
func (c *controller) TakeSnapshot(ctx context.Context, device, path string) (*Snapshot, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Validate input parameters for security
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	// Abstraction layer: Convert camera identifier to device path if needed
	var devicePath string
	var cameraID string

	if c.validateCameraIdentifier(device) {
		// Device is a camera identifier (e.g., "camera0")
		cameraID = device
		devicePath = c.getDevicePathFromCameraIdentifier(device)
		c.logger.WithFields(logging.Fields{
			"camera_id":   cameraID,
			"device_path": devicePath,
		}).Debug("Converted camera identifier to device path")
	} else {
		// Device is already a device path (e.g., "/dev/video0")
		devicePath = device
		cameraID = c.getCameraIdentifierFromDevicePath(device)
	}

	// Generate snapshot ID and path
	snapshotID := generateSnapshotID(devicePath)
	snapshotPath := generateSnapshotPath(devicePath, snapshotID)

	// Take snapshot using FFmpeg
	err := c.ffmpegManager.TakeSnapshot(ctx, devicePath, snapshotPath)
	if err != nil {
		return nil, NewFFmpegErrorWithErr(0, "snapshot", "take_snapshot", "failed to take snapshot", err)
	}

	// Get file info
	fileSize, _, err := c.ffmpegManager.GetFileInfo(ctx, snapshotPath)
	if err != nil {
		return nil, NewFFmpegErrorWithErr(0, "snapshot", "get_file_info", "failed to get file info", err)
	}

	snapshot := &Snapshot{
		ID:       snapshotID,
		Device:   cameraID, // Store camera identifier for API consistency
		Path:     path,
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
	}

	c.logger.WithFields(logging.Fields{
		"snapshot_id": snapshotID,
		"device":      cameraID,
		"device_path": devicePath,
		"path":        path,
		"file_size":   fileSize,
	}).Info("Snapshot taken")

	return snapshot, nil
}

// GetConfig returns the current configuration
func (c *controller) GetConfig(ctx context.Context) (*MediaMTXConfig, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.config, nil
}

// UpdateConfig updates the configuration
func (c *controller) UpdateConfig(ctx context.Context, config *MediaMTXConfig) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	// Validate new configuration
	if err := validateConfig(config); err != nil {
		return NewConfigurationErrorWithErr("config", "validation", "invalid configuration", err)
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
func validateConfig(config *MediaMTXConfig) error {
	if config == nil {
		return NewConfigurationError("config", "nil", "configuration cannot be nil")
	}

	if config.BaseURL == "" {
		return NewConfigurationError("base_url", "", "base URL is required")
	}

	if config.Timeout <= 0 {
		return NewConfigurationError("timeout", config.Timeout.String(), "timeout must be positive")
	}

	if config.RetryAttempts < 0 {
		return NewConfigurationError("retry_attempts", strconv.Itoa(config.RetryAttempts), "retry attempts cannot be negative")
	}

	if config.RetryDelay <= 0 {
		return NewConfigurationError("retry_delay", config.RetryDelay.String(), "retry delay must be positive")
	}

	// Validate circuit breaker configuration
	if config.CircuitBreaker.FailureThreshold <= 0 {
		return NewConfigurationError("circuit_breaker.failure_threshold", strconv.Itoa(config.CircuitBreaker.FailureThreshold), "failure threshold must be positive")
	}

	if config.CircuitBreaker.RecoveryTimeout <= 0 {
		return NewConfigurationError("circuit_breaker.recovery_timeout", config.CircuitBreaker.RecoveryTimeout.String(), "recovery timeout must be positive")
	}

	if config.CircuitBreaker.MaxFailures <= 0 {
		return NewConfigurationError("circuit_breaker.max_failures", strconv.Itoa(config.CircuitBreaker.MaxFailures), "max failures must be positive")
	}

	// Validate connection pool configuration
	if config.ConnectionPool.MaxIdleConns <= 0 {
		return NewConfigurationError("connection_pool.max_idle_conns", strconv.Itoa(config.ConnectionPool.MaxIdleConns), "max idle connections must be positive")
	}

	if config.ConnectionPool.MaxIdleConnsPerHost <= 0 {
		return NewConfigurationError("connection_pool.max_idle_conns_per_host", strconv.Itoa(config.ConnectionPool.MaxIdleConnsPerHost), "max idle connections per host must be positive")
	}

	if config.ConnectionPool.IdleConnTimeout <= 0 {
		return NewConfigurationError("connection_pool.idle_conn_timeout", config.ConnectionPool.IdleConnTimeout.String(), "idle connection timeout must be positive")
	}

	return nil
}

// generateSessionID generates a unique session ID
func generateSessionID(device string) string {
	return "rec_" + device + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// generateSnapshotID generates a unique snapshot ID
func generateSnapshotID(device string) string {
	return "snap_" + device + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// generateRecordingPath generates a recording file path
func generateRecordingPath(device, sessionID string) string {
	// Handle camera identifiers in file naming
	if strings.HasPrefix(device, "camera") {
		// Convert camera0 to camera0 for consistent naming
		return fmt.Sprintf("/opt/camera-service/recordings/%s_%s.mp4", device, sessionID)
	}
	// Handle device paths by extracting the device name
	if strings.HasPrefix(device, "/dev/video") {
		deviceName := strings.TrimPrefix(device, "/dev/")
		return fmt.Sprintf("/opt/camera-service/recordings/%s_%s.mp4", deviceName, sessionID)
	}
	return fmt.Sprintf("/opt/camera-service/recordings/%s_%s.mp4", device, sessionID)
}

// generateSnapshotPath generates a snapshot file path
func generateSnapshotPath(device, snapshotID string) string {
	// Handle camera identifiers in file naming
	if strings.HasPrefix(device, "camera") {
		// Convert camera0 to camera0 for consistent naming
		return fmt.Sprintf("/opt/camera-service/snapshots/%s_%s.jpg", device, snapshotID)
	}
	// Handle device paths by extracting the device name
	if strings.HasPrefix(device, "/dev/video") {
		deviceName := strings.TrimPrefix(device, "/dev/")
		return fmt.Sprintf("/opt/camera-service/snapshots/%s_%s.jpg", deviceName, snapshotID)
	}
	return fmt.Sprintf("/opt/camera-service/snapshots/%s_%s.jpg", device, snapshotID)
}

// StartAdvancedRecording starts a recording with advanced features and full state management
func (c *controller) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*RecordingSession, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Starting advanced recording with full state management")

	// Validate device exists
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	// Abstraction layer: Convert camera identifier to device path if needed
	var devicePath string
	var cameraID string

	if c.validateCameraIdentifier(device) {
		// Device is a camera identifier (e.g., "camera0")
		cameraID = device
		devicePath = c.getDevicePathFromCameraIdentifier(device)
		c.logger.WithFields(logging.Fields{
			"camera_id":   cameraID,
			"device_path": devicePath,
		}).Debug("Converted camera identifier to device path")
	} else {
		// Device is already a device path (e.g., "/dev/video0")
		devicePath = device
		cameraID = c.getCameraIdentifierFromDevicePath(device)
	}

	// Create advanced recording session with full state management
	session, err := c.recordingManager.StartRecording(ctx, devicePath, path, options)
	if err != nil {
		return nil, fmt.Errorf("failed to start advanced recording: %w", err)
	}

	// Store session in controller for state tracking
	c.sessionsMu.Lock()
	c.sessions[session.ID] = session
	c.sessionsMu.Unlock()

	// Initialize session state tracking for Python equivalence
	session.State = SessionStateRecording
	session.ContinuityID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	session.Segments = make([]string, 0)

	// Store the camera identifier in the session for API consistency
	session.Device = cameraID

	// Start tracking active recording for API consistency
	if err := c.StartActiveRecording(cameraID, session.ID, ""); err != nil {
		c.logger.WithError(err).WithField("session_id", session.ID).Warning("Failed to start active recording tracking")
	}

	c.logger.WithFields(logging.Fields{
		"session_id":    session.ID,
		"device":        device,
		"status":        session.Status,
		"state":         session.State,
		"continuity_id": session.ContinuityID,
	}).Info("Advanced recording session started successfully with full state tracking")

	return session, nil
}

// StopAdvancedRecording stops a recording with advanced features and state persistence
func (c *controller) StopAdvancedRecording(ctx context.Context, sessionID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("session_id", sessionID).Info("Stopping advanced recording with state persistence")

	// Get session for state tracking
	c.sessionsMu.RLock()
	session, exists := c.sessions[sessionID]
	c.sessionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("recording session not found: %s", sessionID)
	}

	// Update session state for Python equivalence
	session.State = SessionStateStopped
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)

	// Stop recording using manager
	err := c.recordingManager.StopRecording(ctx, sessionID)
	if err != nil {
		// Update state even if stop fails
		session.State = SessionStateError
		return fmt.Errorf("failed to stop advanced recording: %w", err)
	}

	// Stop tracking active recording for API consistency
	if err := c.StopActiveRecording(session.Device); err != nil {
		c.logger.WithError(err).WithField("session_id", sessionID).Warning("Failed to stop active recording tracking")
	}

	// Persist session state for Python equivalence
	c.persistSessionState(session)

	c.logger.WithFields(logging.Fields{
		"session_id":    sessionID,
		"state":         session.State,
		"duration":      session.Duration,
		"continuity_id": session.ContinuityID,
	}).Info("Advanced recording stopped successfully with state persistence")

	return nil
}

// persistSessionState persists session state for Python equivalence
func (c *controller) persistSessionState(session *RecordingSession) {
	c.logger.WithFields(logging.Fields{
		"session_id":    session.ID,
		"state":         session.State,
		"continuity_id": session.ContinuityID,
	}).Debug("Persisting session state for Python equivalence")

	// Store session in controller's session map for persistence
	c.sessionsMu.Lock()
	c.sessions[session.ID] = session
	c.sessionsMu.Unlock()

	// Log session state for monitoring and debugging
	c.logger.WithFields(logging.Fields{
		"session_id":    session.ID,
		"device":        session.Device,
		"state":         session.State,
		"status":        session.Status,
		"continuity_id": session.ContinuityID,
		"duration":      session.Duration,
		"file_size":     session.FileSize,
		"segments":      len(session.Segments),
	}).Info("Session state persisted successfully")
}

// GetAdvancedRecordingSession gets a recording session
func (c *controller) GetAdvancedRecordingSession(sessionID string) (*RecordingSession, bool) {
	return c.recordingManager.GetRecordingSession(sessionID)
}

// ListAdvancedRecordingSessions lists all recording sessions
func (c *controller) ListAdvancedRecordingSessions() []*RecordingSession {
	return c.recordingManager.ListRecordingSessions()
}

// RotateRecordingFile rotates a recording file
func (c *controller) RotateRecordingFile(ctx context.Context, sessionID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	return c.recordingManager.RotateRecordingFile(ctx, sessionID)
}

// TakeAdvancedSnapshot takes a snapshot with multi-tier approach (enhanced existing method)
func (c *controller) TakeAdvancedSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*Snapshot, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Taking multi-tier advanced snapshot")

	// Abstraction layer: Convert camera identifier to device path if needed
	var devicePath string
	var cameraID string

	if c.validateCameraIdentifier(device) {
		// Device is a camera identifier (e.g., "camera0")
		cameraID = device
		devicePath = c.getDevicePathFromCameraIdentifier(device)
		c.logger.WithFields(logging.Fields{
			"camera_id":   cameraID,
			"device_path": devicePath,
		}).Debug("Converted camera identifier to device path")
	} else {
		// Device is already a device path (e.g., "/dev/video0")
		devicePath = device
		cameraID = c.getCameraIdentifierFromDevicePath(device)
	}

	// Use enhanced snapshot manager with multi-tier capability
	snapshot, err := c.snapshotManager.TakeSnapshot(ctx, devicePath, path, options)
	if err != nil {
		c.logger.WithError(err).WithFields(logging.Fields{
			"device": device,
			"path":   path,
		}).Error("Multi-tier snapshot failed")
		return nil, fmt.Errorf("failed to take multi-tier snapshot for device %s: %w", device, err)
	}

	// Store the camera identifier in the snapshot for API consistency
	snapshot.Device = cameraID

	// Log tier information for monitoring
	if snapshot.Metadata != nil {
		if tierUsed, ok := snapshot.Metadata["tier_used"]; ok {
			c.logger.WithFields(logging.Fields{
				"device":    cameraID,
				"tier_used": tierUsed,
				"file_size": snapshot.Size,
			}).Info("Multi-tier snapshot completed successfully")
		}
	}

	return snapshot, nil
}

// GetAdvancedSnapshot gets a snapshot by ID
func (c *controller) GetAdvancedSnapshot(snapshotID string) (*Snapshot, bool) {
	return c.snapshotManager.GetSnapshot(snapshotID)
}

// ListAdvancedSnapshots lists all snapshots
func (c *controller) ListAdvancedSnapshots() []*Snapshot {
	return c.snapshotManager.ListSnapshots()
}

// DeleteAdvancedSnapshot deletes a snapshot
func (c *controller) DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	return c.snapshotManager.DeleteSnapshot(ctx, snapshotID)
}

// GetSnapshotSettings gets current snapshot settings
func (c *controller) GetSnapshotSettings() *SnapshotSettings {
	return c.snapshotManager.GetSnapshotSettings()
}

// UpdateSnapshotSettings updates snapshot settings
func (c *controller) UpdateSnapshotSettings(settings *SnapshotSettings) {
	c.snapshotManager.UpdateSnapshotSettings(settings)
}

// GetRecordingStatus gets the status of a recording session
func (c *controller) GetRecordingStatus(ctx context.Context, sessionID string) (*RecordingSession, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Validate input parameters for security
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	c.logger.WithField("session_id", sessionID).Debug("Getting recording status")

	// Check if session exists
	c.sessionsMu.RLock()
	session, exists := c.sessions[sessionID]
	c.sessionsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("recording session not found: %s", sessionID)
	}

	return session, nil
}

// ListRecordings lists recording files with metadata and pagination
func (c *controller) ListRecordings(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Listing recordings")

	return c.recordingManager.GetRecordingsList(ctx, limit, offset)
}

// ListSnapshots lists snapshot files with metadata and pagination
func (c *controller) ListSnapshots(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Listing snapshots")

	return c.snapshotManager.GetSnapshotsList(ctx, limit, offset)
}

// GetRecordingInfo gets detailed information about a specific recording file
func (c *controller) GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Getting recording info")

	return c.recordingManager.GetRecordingInfo(ctx, filename)
}

// GetSnapshotInfo gets detailed information about a specific snapshot file
func (c *controller) GetSnapshotInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Getting snapshot info")

	return c.snapshotManager.GetSnapshotInfo(ctx, filename)
}

// DeleteRecording deletes a recording file
func (c *controller) DeleteRecording(ctx context.Context, filename string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Deleting recording")

	return c.recordingManager.DeleteRecording(ctx, filename)
}

// DeleteSnapshot deletes a snapshot file
func (c *controller) DeleteSnapshot(ctx context.Context, filename string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("filename", filename).Debug("Deleting snapshot")

	return c.snapshotManager.DeleteSnapshotFile(ctx, filename)
}

// GetSessionIDByDevice gets session ID by device path using optimized lookup
func (c *controller) GetSessionIDByDevice(device string) (string, bool) {
	// Abstraction layer: Convert camera identifier to device path if needed
	var devicePath string
	if c.validateCameraIdentifier(device) {
		devicePath = c.getDevicePathFromCameraIdentifier(device)
	} else {
		devicePath = device
	}

	return c.recordingManager.getSessionIDByDevice(devicePath)
}

// RTSP Connection Management Methods

// ListRTSPConnections lists all RTSP connections
func (c *controller) ListRTSPConnections(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionList, error) {
	if !c.isRunning {
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
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Debug("Getting RTSP connection")

	return c.rtspManager.GetConnection(ctx, id)
}

// ListRTSPSessions lists all RTSP sessions
func (c *controller) ListRTSPSessions(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionSessionList, error) {
	if !c.isRunning {
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
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Debug("Getting RTSP session")

	return c.rtspManager.GetSession(ctx, id)
}

// KickRTSPSession kicks out an RTSP session from the server
func (c *controller) KickRTSPSession(ctx context.Context, id string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("id", id).Info("Kicking RTSP session")

	return c.rtspManager.KickSession(ctx, id)
}

// GetRTSPConnectionHealth returns the health status of RTSP connections
func (c *controller) GetRTSPConnectionHealth(ctx context.Context) (*HealthStatus, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	return c.rtspManager.GetConnectionHealth(ctx)
}

// GetRTSPConnectionMetrics returns metrics about RTSP connections
func (c *controller) GetRTSPConnectionMetrics(ctx context.Context) map[string]interface{} {
	if !c.isRunning {
		return map[string]interface{}{
			"error": "controller is not running",
		}
	}

	return c.rtspManager.GetConnectionMetrics(ctx)
}

// StartStreaming starts a live streaming session for the specified device
func (c *controller) StartStreaming(ctx context.Context, device string) (*Stream, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"device": device,
		"action": "start_streaming",
	}).Info("Starting streaming session")

	// Use StreamManager to start viewing stream
	stream, err := c.streamManager.StartViewingStream(ctx, device)
	if err != nil {
		c.logger.WithFields(logging.Fields{
			"device": device,
			"error":  err.Error(),
		}).Error("Failed to start streaming")
		return nil, fmt.Errorf("failed to start streaming: %w", err)
	}

	c.logger.WithFields(logging.Fields{
		"device":      device,
		"stream_name": stream.Name,
		"stream_url":  stream.URL,
	}).Info("Streaming session started successfully")

	return stream, nil
}

// StopStreaming stops the streaming session for the specified device
func (c *controller) StopStreaming(ctx context.Context, device string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logging.Fields{
		"device": device,
		"action": "stop_streaming",
	}).Info("Stopping streaming session")

	// Use StreamManager to stop viewing stream
	err := c.streamManager.StopViewingStream(ctx, device)
	if err != nil {
		c.logger.WithFields(logging.Fields{
			"device": device,
			"error":  err.Error(),
		}).Error("Failed to stop streaming")
		return fmt.Errorf("failed to stop streaming: %w", err)
	}

	c.logger.WithFields(logging.Fields{
		"device": device,
	}).Info("Streaming session stopped successfully")

	return nil
}

// GetStreamURL returns the stream URL for the specified device
func (c *controller) GetStreamURL(ctx context.Context, device string) (string, error) {
	if !c.isRunning {
		return "", fmt.Errorf("controller is not running")
	}

	// Generate stream name and URL using existing StreamManager method
	streamName := c.streamManager.GenerateStreamName(device, UseCaseViewing)
	streamURL := c.streamManager.GenerateStreamURL(streamName)

	c.logger.WithFields(logging.Fields{
		"device":      device,
		"stream_name": streamName,
		"stream_url":  streamURL,
	}).Debug("Generated stream URL")

	return streamURL, nil
}

// GetStreamStatus returns the status of the streaming session for the specified device
func (c *controller) GetStreamStatus(ctx context.Context, device string) (*Stream, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	// Generate stream name for viewing use case
	streamName := c.streamManager.GenerateStreamName(device, UseCaseViewing)

	// Try to get the stream from MediaMTX
	stream, err := c.streamManager.GetStream(ctx, streamName)
	if err != nil {
		c.logger.WithFields(logging.Fields{
			"device":      device,
			"stream_name": streamName,
			"error":       err.Error(),
		}).Debug("Stream not found or not active")
		return nil, fmt.Errorf("stream not found or not active: %w", err)
	}

	c.logger.WithFields(logging.Fields{
		"device":      device,
		"stream_name": stream.Name,
		"ready":       stream.Ready,
	}).Debug("Retrieved stream status")

	return stream, nil
}
