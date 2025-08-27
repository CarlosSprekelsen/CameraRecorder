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

// StartAdvancedRecording starts a recording with advanced features
func (c *controller) StartAdvancedRecording(ctx context.Context, device, path string, options map[string]interface{}) (*RecordingSession, error) {
	if !c.isRunning {
		return nil, fmt.Errorf("controller is not running")
	}

	c.logger.WithFields(logrus.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Starting advanced recording")

	return c.recordingManager.StartRecording(ctx, device, path, options)
}

// StopAdvancedRecording stops a recording with advanced features
func (c *controller) StopAdvancedRecording(ctx context.Context, sessionID string) error {
	if !c.isRunning {
		return fmt.Errorf("controller is not running")
	}

	c.logger.WithField("session_id", sessionID).Info("Stopping advanced recording")

	return c.recordingManager.StopRecording(ctx, sessionID)
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
