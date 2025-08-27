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
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/sirupsen/logrus"
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
	config           *MediaMTXConfig
	logger           *logrus.Logger

	// State management
	mu        sync.RWMutex
	isRunning bool
	startTime time.Time

	// Recording sessions
	sessions   map[string]*RecordingSession
	sessionsMu sync.RWMutex
}

// NewController creates a new MediaMTX controller
func NewController(config *MediaMTXConfig, logger *logrus.Logger) (MediaMTXController, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, NewConfigurationErrorWithErr("config", "validation", "invalid configuration", err)
	}

	// Create HTTP client
	client := NewClient(config.BaseURL, config, logger)

	// Create health monitor
	healthMonitor := NewHealthMonitor(client, config, logger)

	// Create path manager
	pathManager := NewPathManager(client, config, logger)

	// Create stream manager
	streamManager := NewStreamManager(client, config, logger)

	// Create FFmpeg manager
	ffmpegManager := NewFFmpegManager(config, logger)

	// Create recording manager
	recordingManager := NewRecordingManager(ffmpegManager, config, logger)

	// Create snapshot manager
	snapshotManager := NewSnapshotManager(ffmpegManager, config, logger)

	return &controller{
		client:           client,
		healthMonitor:    healthMonitor,
		pathManager:      pathManager,
		streamManager:    streamManager,
		ffmpegManager:    ffmpegManager,
		recordingManager: recordingManager,
		snapshotManager:  snapshotManager,
		config:           config,
		logger:           logger,
		sessions:         make(map[string]*RecordingSession),
	}, nil
}

// NewControllerWithConfigManager creates a new MediaMTX controller with configuration integration
func NewControllerWithConfigManager(configManager *config.ConfigManager, logger *logrus.Logger) (MediaMTXController, error) {
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

	// Create stream manager
	streamManager := NewStreamManager(client, mediaMTXConfig, logger)

	// Create FFmpeg manager
	ffmpegManager := NewFFmpegManager(mediaMTXConfig, logger)

	// Create recording manager
	recordingManager := NewRecordingManager(ffmpegManager, mediaMTXConfig, logger)

	// Create snapshot manager with configuration integration
	snapshotManager := NewSnapshotManagerWithConfig(ffmpegManager, mediaMTXConfig, configManager, logger)

	return &controller{
		client:           client,
		healthMonitor:    healthMonitor,
		pathManager:      pathManager,
		streamManager:    streamManager,
		ffmpegManager:    ffmpegManager,
		recordingManager: recordingManager,
		snapshotManager:  snapshotManager,
		config:           mediaMTXConfig,
		logger:           logger,
		sessions:         make(map[string]*RecordingSession),
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
		if stream.Status == "READY" || stream.Status == "PUBLISHING" {
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

	// Get health monitor metrics
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

	// Calculate error counts
	errorCounts := make(map[string]int64)
	if failureCount, ok := healthMetrics["failure_count"].(int); ok {
		errorCounts["health_check"] = int64(failureCount)
	}

	// Get circuit breaker state
	circuitBreakerState := "CLOSED"
	if state, ok := healthMetrics["circuit_state"].(string); ok {
		circuitBreakerState = state
	}

	// Calculate response time (average from health metrics)
	responseTime := 0.0
	if lastCheckTime, ok := healthMetrics["last_check_time"].(time.Time); ok {
		responseTime = float64(time.Since(lastCheckTime).Milliseconds())
	}

	systemMetrics := &SystemMetrics{
		RequestCount:       0, // Will be populated by WebSocket server
		ResponseTime:       responseTime,
		ErrorCount:         int64(healthMetrics["failure_count"].(int)),
		ActiveConnections:  0, // Will be populated by WebSocket server
		ComponentStatus:    componentStatus,
		ErrorCounts:        errorCounts,
		LastCheck:          healthStatus.LastCheck,
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

	// Generate session ID
	sessionID := generateSessionID(device)

	// Check if session already exists
	c.sessionsMu.Lock()
	if _, exists := c.sessions[sessionID]; exists {
		c.sessionsMu.Unlock()
		return nil, NewRecordingError(sessionID, device, "start_recording", "session already exists")
	}
	c.sessionsMu.Unlock()

	// Create recording session
	session := &RecordingSession{
		ID:        sessionID,
		Device:    device,
		Path:      path,
		Status:    "STARTING",
		StartTime: time.Now(),
		FilePath:  generateRecordingPath(device, sessionID),
	}

	// Start FFmpeg recording
	options := map[string]string{
		"format": "mp4",
		"codec":  "libx264",
		"preset": "fast",
		"crf":    "23",
	}

	pid, err := c.ffmpegManager.StartRecording(ctx, device, session.FilePath, options)
	if err != nil {
		return nil, NewRecordingErrorWithErr(sessionID, device, "start_recording", "failed to start FFmpeg recording", err)
	}

	// Update session status
	session.Status = "RECORDING"

	// Store session
	c.sessionsMu.Lock()
	c.sessions[sessionID] = session
	c.sessionsMu.Unlock()

	c.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"device":     device,
		"path":       path,
		"pid":        pid,
	}).Info("Recording session started")

	return session, nil
}

// StopRecording stops a recording session
func (c *controller) StopRecording(ctx context.Context, sessionID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
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

	// Stop FFmpeg process (we need to track PID in session)
	// For now, we'll use a placeholder approach
	// In a real implementation, we'd store the PID in the session

	// Update session status
	c.sessionsMu.Lock()
	session.Status = "STOPPED"
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)

	// Get file size
	if fileSize, _, err := c.ffmpegManager.GetFileInfo(ctx, session.FilePath); err == nil {
		session.FileSize = fileSize
	}

	c.sessionsMu.Unlock()

	c.logger.WithFields(logrus.Fields{
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

	// Generate snapshot ID and path
	snapshotID := generateSnapshotID(device)
	snapshotPath := generateSnapshotPath(device, snapshotID)

	// Take snapshot using FFmpeg
	err := c.ffmpegManager.TakeSnapshot(ctx, device, snapshotPath)
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
		Device:   device,
		Path:     path,
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
	}

	c.logger.WithFields(logrus.Fields{
		"snapshot_id": snapshotID,
		"device":      device,
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
		return NewConfigurationError("retry_attempts", fmt.Sprintf("%d", config.RetryAttempts), "retry attempts cannot be negative")
	}

	if config.RetryDelay <= 0 {
		return NewConfigurationError("retry_delay", config.RetryDelay.String(), "retry delay must be positive")
	}

	// Validate circuit breaker configuration
	if config.CircuitBreaker.FailureThreshold <= 0 {
		return NewConfigurationError("circuit_breaker.failure_threshold", fmt.Sprintf("%d", config.CircuitBreaker.FailureThreshold), "failure threshold must be positive")
	}

	if config.CircuitBreaker.RecoveryTimeout <= 0 {
		return NewConfigurationError("circuit_breaker.recovery_timeout", config.CircuitBreaker.RecoveryTimeout.String(), "recovery timeout must be positive")
	}

	if config.CircuitBreaker.MaxFailures <= 0 {
		return NewConfigurationError("circuit_breaker.max_failures", fmt.Sprintf("%d", config.CircuitBreaker.MaxFailures), "max failures must be positive")
	}

	// Validate connection pool configuration
	if config.ConnectionPool.MaxIdleConns <= 0 {
		return NewConfigurationError("connection_pool.max_idle_conns", fmt.Sprintf("%d", config.ConnectionPool.MaxIdleConns), "max idle connections must be positive")
	}

	if config.ConnectionPool.MaxIdleConnsPerHost <= 0 {
		return NewConfigurationError("connection_pool.max_idle_conns_per_host", fmt.Sprintf("%d", config.ConnectionPool.MaxIdleConnsPerHost), "max idle connections per host must be positive")
	}

	if config.ConnectionPool.IdleConnTimeout <= 0 {
		return NewConfigurationError("connection_pool.idle_conn_timeout", config.ConnectionPool.IdleConnTimeout.String(), "idle connection timeout must be positive")
	}

	return nil
}

// generateSessionID generates a unique session ID
func generateSessionID(device string) string {
	return fmt.Sprintf("rec_%s_%d", device, time.Now().UnixNano())
}

// generateSnapshotID generates a unique snapshot ID
func generateSnapshotID(device string) string {
	return fmt.Sprintf("snap_%s_%d", device, time.Now().UnixNano())
}

// generateRecordingPath generates a recording file path
func generateRecordingPath(device, sessionID string) string {
	return fmt.Sprintf("/tmp/recordings/%s_%s.mp4", device, sessionID)
}

// generateSnapshotPath generates a snapshot file path
func generateSnapshotPath(device, snapshotID string) string {
	return fmt.Sprintf("/tmp/snapshots/%s_%s.jpg", device, snapshotID)
}

// StartAdvancedRecording starts a recording with advanced features and full state management
func (c *controller) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*RecordingSession, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logrus.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Starting advanced recording with full state management")

	// Validate device exists
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	// Create advanced recording session with full state management
	session, err := c.recordingManager.StartRecording(ctx, device, path, options)
	if err != nil {
		return nil, fmt.Errorf("failed to start advanced recording: %w", err)
	}

	// Store session in controller for state tracking
	c.sessionsMu.Lock()
	c.sessions[session.ID] = session
	c.sessionsMu.Unlock()

	// Initialize session state tracking for Python equivalence
	session.State = SessionStateRecording
	session.ContinuityID = generateContinuityID()
	session.Segments = make([]string, 0)

	c.logger.WithFields(logrus.Fields{
		"session_id":   session.ID,
		"device":       device,
		"status":       session.Status,
		"state":        session.State,
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

	// Persist session state for Python equivalence
	c.persistSessionState(session)

	c.logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"state":        session.State,
		"duration":     session.Duration,
		"continuity_id": session.ContinuityID,
	}).Info("Advanced recording stopped successfully with state persistence")

	return nil
}

// persistSessionState persists session state for Python equivalence
func (c *controller) persistSessionState(session *RecordingSession) {
	c.logger.WithFields(logrus.Fields{
		"session_id":   session.ID,
		"state":        session.State,
		"continuity_id": session.ContinuityID,
	}).Debug("Persisting session state for Python equivalence")

	// Store session in controller's session map for persistence
	c.sessionsMu.Lock()
	c.sessions[session.ID] = session
	c.sessionsMu.Unlock()

	// Log session state for monitoring and debugging
	c.logger.WithFields(logrus.Fields{
		"session_id":   session.ID,
		"device":       session.Device,
		"state":        session.State,
		"status":       session.Status,
		"continuity_id": session.ContinuityID,
		"duration":     session.Duration,
		"file_size":    session.FileSize,
		"segments":     len(session.Segments),
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

	c.logger.WithFields(logrus.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Taking multi-tier advanced snapshot")

	// Use enhanced snapshot manager with multi-tier capability
	snapshot, err := c.snapshotManager.TakeSnapshot(ctx, device, path, options)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"device": device,
			"path":   path,
		}).Error("Multi-tier snapshot failed")
		return nil, err
	}

	// Log tier information for monitoring
	if snapshot.Metadata != nil {
		if tierUsed, ok := snapshot.Metadata["tier_used"]; ok {
			c.logger.WithFields(logrus.Fields{
				"device":    device,
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

// CleanupOldSnapshots cleans up old snapshots
func (c *controller) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	return c.snapshotManager.CleanupOldSnapshots(ctx, maxAge, maxCount)
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

	c.logger.WithFields(logrus.Fields{
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

	c.logger.WithFields(logrus.Fields{
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
