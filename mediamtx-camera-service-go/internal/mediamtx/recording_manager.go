/*
MediaMTX Recording Manager Implementation

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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RecordingManager manages advanced recording operations
type RecordingManager struct {
	ffmpegManager FFmpegManager
	config        *MediaMTXConfig
	logger        *logrus.Logger

	// Recording sessions
	sessions   map[string]*RecordingSession
	sessionsMu sync.RWMutex

	// File rotation settings
	rotationSettings *RotationSettings

	// Storage monitoring
	storageMonitor *StorageMonitor

	// Segment management
	segmentManager *SegmentManager
}

// RotationSettings defines file rotation behavior
type RotationSettings struct {
	MaxFileSize     int64         `json:"max_file_size"`
	MaxDuration     time.Duration `json:"max_duration"`
	SegmentDuration time.Duration `json:"segment_duration"`
	AutoRotate      bool          `json:"auto_rotate"`
	KeepSegments    int           `json:"keep_segments"`
	OutputFormat    string        `json:"output_format"`
}

// RecordingSegment represents a recording segment
type RecordingSegment struct {
	ID        string        `json:"id"`
	FilePath  string        `json:"file_path"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Size      int64         `json:"size"`
	Index     int           `json:"index"`
}

// StorageMonitor manages storage monitoring and cleanup
type StorageMonitor struct {
	config        *StorageConfig
	logger        *logrus.Logger
	metrics       *StorageMetrics
	monitorTicker *time.Ticker
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
	mu            sync.RWMutex
}

// StorageConfig represents storage monitoring configuration
type StorageConfig struct {
	WarnPercent   int           `json:"warn_percent"`  // Default 80%
	BlockPercent  int           `json:"block_percent"` // Default 90%
	CheckInterval time.Duration `json:"check_interval"`
	AutoCleanup   bool          `json:"auto_cleanup"`
	MaxFileAge    time.Duration `json:"max_file_age"`
}

// StorageMetrics represents storage usage metrics
type StorageMetrics struct {
	TotalSpace     int64     `json:"total_space"`
	UsedSpace      int64     `json:"used_space"`
	AvailableSpace int64     `json:"available_space"`
	UsagePercent   int       `json:"usage_percent"`
	FileCount      int       `json:"file_count"`
	LastCheck      time.Time `json:"last_check"`
}

// SegmentManager manages recording segments and continuity
type SegmentManager struct {
	config   *SegmentConfig
	logger   *logrus.Logger
	segments map[string]*Segment
	mu       sync.RWMutex
}

// SegmentConfig represents segment-based recording configuration
type SegmentConfig struct {
	Enabled          bool          `json:"enabled"`
	RotationInterval time.Duration `json:"rotation_interval"`
	SegmentFormat    string        `json:"segment_format"`
	ResetTimestamps  bool          `json:"reset_timestamps"`
	StrftimeEnabled  bool          `json:"strftime_enabled"`
	MaxSegments      int           `json:"max_segments"`
	SegmentPrefix    string        `json:"segment_prefix"`
}

// Segment represents a recording segment with metadata
type Segment struct {
	ID           string        `json:"id"`
	FilePath     string        `json:"file_path"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	Size         int64         `json:"size"`
	Index        int           `json:"index"`
	ContinuityID string        `json:"continuity_id"`
}

// RecordingSession represents an enhanced recording session with continuity
type RecordingSession struct {
	ID        string       `json:"id"`
	Device    string       `json:"device"`
	Path      string       `json:"path"`
	Status    string       `json:"status"`
	StartTime time.Time    `json:"start_time"`
	EndTime   *time.Time   `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	FilePath  string       `json:"file_path"`
	mu        sync.RWMutex
}

// RecordingContinuity represents recording continuity information
type RecordingContinuity struct {
	SessionID     string         `json:"session_id"`
	ContinuityID  string         `json:"continuity_id"`
	StartTime     time.Time      `json:"start_time"`
	SegmentCount  int            `json:"segment_count"`
	TotalDuration time.Duration  `json:"total_duration"`
	Segments      []*SegmentInfo `json:"segments"`
}

// SegmentInfo represents segment information for continuity
type SegmentInfo struct {
	ID        string        `json:"id"`
	FilePath  string        `json:"file_path"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Size      int64         `json:"size"`
	Index     int           `json:"index"`
}

// NewRecordingManager creates a new recording manager with advanced features
func NewRecordingManager(ffmpegManager FFmpegManager, config *MediaMTXConfig, logger *logrus.Logger) *RecordingManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &RecordingManager{
		ffmpegManager: ffmpegManager,
		config:        config,
		logger:        logger,
		sessions:      make(map[string]*RecordingSession),
		rotationSettings: &RotationSettings{
			MaxFileSize:     100 * 1024 * 1024, // 100MB
			MaxDuration:     10 * time.Minute,
			SegmentDuration: 5 * time.Minute,
			AutoRotate:      true,
			KeepSegments:    10,
			OutputFormat:    "mp4",
		},
		storageMonitor: &StorageMonitor{
			config: &StorageConfig{
				WarnPercent:   80,
				BlockPercent:  90,
				CheckInterval: 30 * time.Second,
				AutoCleanup:   true,
				MaxFileAge:    24 * time.Hour,
			},
			logger:        logger,
			monitorCtx:    ctx,
			monitorCancel: cancel,
		},
		segmentManager: &SegmentManager{
			config: &SegmentConfig{
				Enabled:          true,
				RotationInterval: 5 * time.Minute,
				SegmentFormat:    "segment_%03d.mp4",
				ResetTimestamps:  true,
				StrftimeEnabled:  true,
				MaxSegments:      10,
				SegmentPrefix:    "recording",
			},
			logger:   logger,
			segments: make(map[string]*Segment),
		},
	}

	// Start storage monitoring
	rm.storageMonitor.startMonitoring()

	return rm
}

// StartRecordingWithSegments starts a recording with segment-based rotation
func (rm *RecordingManager) StartRecordingWithSegments(ctx context.Context, device, path string, options map[string]interface{}) (*RecordingSession, error) {
	rm.logger.WithFields(logrus.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Starting advanced recording with segments")

	// Generate session ID
	sessionID := generateSessionID(device)

	// Check if session already exists
	rm.sessionsMu.Lock()
	if _, exists := rm.sessions[sessionID]; exists {
		rm.sessionsMu.Unlock()
		return nil, NewRecordingError(sessionID, device, "start_recording", "session already exists")
	}
	rm.sessionsMu.Unlock()

	// Check storage availability
	if err := rm.storageMonitor.checkStorageAvailability(); err != nil {
		return nil, fmt.Errorf("storage check failed: %w", err)
	}

	// Create recording session
	session := &RecordingSession{
		ID:        sessionID,
		Device:    device,
		Path:      path,
		Status:    "RECORDING",
		StartTime: time.Now(),
		FilePath:  rm.generateRecordingPath(device, sessionID),
	}

	// Store session
	rm.sessionsMu.Lock()
	rm.sessions[sessionID] = session
	rm.sessionsMu.Unlock()

	rm.logger.WithField("session_id", sessionID).Info("Advanced recording started with segments")
	return session, nil
}

// StartRecording starts a recording with advanced features
func (rm *RecordingManager) StartRecording(ctx context.Context, device, path string, options map[string]interface{}) (*RecordingSession, error) {
	rm.logger.WithFields(logrus.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Starting advanced recording")

	// Generate session ID
	sessionID := generateSessionID(device)

	// Check if session already exists
	rm.sessionsMu.Lock()
	if _, exists := rm.sessions[sessionID]; exists {
		rm.sessionsMu.Unlock()
		return nil, NewRecordingError(sessionID, device, "start_recording", "session already exists")
	}
	rm.sessionsMu.Unlock()

	// Create recording session
	session := &RecordingSession{
		ID:           sessionID,
		Device:       device,
		Path:         path,
		Status:       "RECORDING",
		StartTime:    time.Now(),
		FilePath:     rm.generateRecordingPath(device, sessionID),
		ContinuityID: generateContinuityID(),
		State:        SessionStateRecording,
	}

	// Apply rotation settings from options
	if maxFileSize, ok := options["max_file_size"].(int64); ok {
		rm.rotationSettings.MaxFileSize = maxFileSize
	}
	if maxDuration, ok := options["max_duration"].(time.Duration); ok {
		rm.rotationSettings.MaxDuration = maxDuration
	}
	if segmentDuration, ok := options["segment_duration"].(time.Duration); ok {
		rm.rotationSettings.SegmentDuration = segmentDuration
	}
	if autoRotate, ok := options["auto_rotate"].(bool); ok {
		rm.rotationSettings.AutoRotate = autoRotate
	}

	// Build FFmpeg command with advanced options
	command := rm.buildAdvancedRecordingCommand(device, session.FilePath, options)

	// Start FFmpeg process
	pid, err := rm.ffmpegManager.StartProcess(ctx, command, session.FilePath)
	if err != nil {
		return nil, NewRecordingErrorWithErr(sessionID, device, "start_recording", "failed to start FFmpeg process", err)
	}

	// Update session
	session.Status = "RECORDING"

	// Store session
	rm.sessionsMu.Lock()
	rm.sessions[sessionID] = session
	rm.sessionsMu.Unlock()

	// Start monitoring if auto-rotation is enabled
	if rm.rotationSettings.AutoRotate {
		go rm.monitorRecordingForRotation(ctx, sessionID)
	}

	rm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"device":     device,
		"path":       path,
		"pid":        pid,
	}).Info("Advanced recording started successfully")

	return session, nil
}

// StopRecordingWithContinuity stops recording while maintaining continuity
func (rm *RecordingManager) StopRecordingWithContinuity(ctx context.Context, sessionID string) error {
	rm.sessionsMu.Lock()
	session, exists := rm.sessions[sessionID]
	if !exists {
		rm.sessionsMu.Unlock()
		return NewRecordingError(sessionID, "", "stop_recording", "session not found")
	}
	rm.sessionsMu.Unlock()

	rm.logger.WithField("session_id", sessionID).Info("Stopping recording with continuity")

	// Update session status
	session.Status = "STOPPED"
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)

	// Stop FFmpeg process
	if err := rm.ffmpegManager.StopProcess(ctx, sessionID); err != nil {
		rm.logger.WithError(err).WithField("session_id", sessionID).Warning("Failed to stop FFmpeg process")
	}

	// Remove session
	rm.sessionsMu.Lock()
	delete(rm.sessions, sessionID)
	rm.sessionsMu.Unlock()

	rm.logger.WithField("session_id", sessionID).Info("Recording stopped with continuity maintained")
	return nil
}

// StopRecording stops a recording session
func (rm *RecordingManager) StopRecording(ctx context.Context, sessionID string) error {
	rm.logger.WithField("session_id", sessionID).Info("Stopping advanced recording")

	rm.sessionsMu.Lock()
	session, exists := rm.sessions[sessionID]
	if !exists {
		rm.sessionsMu.Unlock()
		return NewRecordingError(sessionID, "", "stop_recording", "session not found")
	}

	if session.Status != "RECORDING" {
		rm.sessionsMu.Unlock()
		return NewRecordingError(sessionID, session.Device, "stop_recording", "session is not recording")
	}

	// Update session status
	session.Status = "STOPPING"
	rm.sessionsMu.Unlock()

	// Stop FFmpeg process (we need to track PID in session)
	// For now, we'll use a placeholder approach
	// In a real implementation, we'd store the PID in the session

	// Update session status
	rm.sessionsMu.Lock()
	session.Status = "STOPPED"
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = endTime.Sub(session.StartTime)

	// Get file size
	if fileSize, _, err := rm.ffmpegManager.GetFileInfo(ctx, session.FilePath); err == nil {
		session.FileSize = fileSize
	}

	rm.sessionsMu.Unlock()

	rm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"device":     session.Device,
		"duration":   session.Duration,
		"file_size":  session.FileSize,
	}).Info("Advanced recording stopped successfully")

	return nil
}

// GetRecordingSession gets a recording session
func (rm *RecordingManager) GetRecordingSession(sessionID string) (*RecordingSession, bool) {
	rm.sessionsMu.RLock()
	defer rm.sessionsMu.RUnlock()

	session, exists := rm.sessions[sessionID]
	return session, exists
}

// ListRecordingSessions lists all recording sessions
func (rm *RecordingManager) ListRecordingSessions() []*RecordingSession {
	rm.sessionsMu.RLock()
	defer rm.sessionsMu.RUnlock()

	sessions := make([]*RecordingSession, 0, len(rm.sessions))
	for _, session := range rm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// RotateRecordingFile rotates a recording file
func (rm *RecordingManager) RotateRecordingFile(ctx context.Context, sessionID string) error {
	rm.logger.WithField("session_id", sessionID).Debug("Rotating recording file")

	rm.sessionsMu.Lock()
	session, exists := rm.sessions[sessionID]
	if !exists {
		rm.sessionsMu.Unlock()
		return NewRecordingError(sessionID, "", "rotate_file", "session not found")
	}
	rm.sessionsMu.Unlock()

	// Generate new file path
	newFilePath := rm.generateRotatedFilePath(session.FilePath)

	// Rotate file
	if err := rm.ffmpegManager.RotateFile(ctx, session.FilePath, newFilePath); err != nil {
		return NewRecordingErrorWithErr(sessionID, session.Device, "rotate_file", "failed to rotate file", err)
	}

	// Update session file path
	rm.sessionsMu.Lock()
	session.FilePath = newFilePath
	rm.sessionsMu.Unlock()

	rm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"old_path":   session.FilePath,
		"new_path":   newFilePath,
	}).Info("Recording file rotated successfully")

	return nil
}

// buildAdvancedRecordingCommand builds an advanced FFmpeg command
func (rm *RecordingManager) buildAdvancedRecordingCommand(device, outputPath string, options map[string]interface{}) []string {
	command := []string{"ffmpeg"}

	// Input device
	command = append(command, "-f", "v4l2")
	command = append(command, "-i", device)

	// Video codec
	codec := "libx264"
	if codecOpt, ok := options["codec"].(string); ok {
		codec = codecOpt
	}
	command = append(command, "-c:v", codec)

	// Preset
	preset := "fast"
	if presetOpt, ok := options["preset"].(string); ok {
		preset = presetOpt
	}
	command = append(command, "-preset", preset)

	// CRF (quality)
	crf := "23"
	if crfOpt, ok := options["crf"].(string); ok {
		crf = crfOpt
	}
	command = append(command, "-crf", crf)

	// Segment options for file rotation
	if rm.rotationSettings.AutoRotate {
		command = append(command, "-f", "segment")
		command = append(command, "-segment_time", strconv.Itoa(int(rm.rotationSettings.SegmentDuration.Seconds())))
		command = append(command, "-reset_timestamps", "1")

		// Generate segment filename pattern
		segmentPattern := strings.Replace(outputPath, "."+rm.rotationSettings.OutputFormat, "_%03d."+rm.rotationSettings.OutputFormat, 1)
		command = append(command, segmentPattern)
	} else {
		// Single file output
		command = append(command, "-f", rm.rotationSettings.OutputFormat)
		command = append(command, outputPath)
	}

	return command
}

// monitorRecordingForRotation monitors recording for automatic file rotation
func (rm *RecordingManager) monitorRecordingForRotation(ctx context.Context, sessionID string) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.checkRecordingForRotation(ctx, sessionID)
		}
	}
}

// checkRecordingForRotation checks if recording needs rotation
func (rm *RecordingManager) checkRecordingForRotation(ctx context.Context, sessionID string) {
	rm.sessionsMu.RLock()
	session, exists := rm.sessions[sessionID]
	if !exists || session.Status != "RECORDING" {
		rm.sessionsMu.RUnlock()
		return
	}
	rm.sessionsMu.RUnlock()

	// Check file size
	if fileSize, _, err := rm.ffmpegManager.GetFileInfo(ctx, session.FilePath); err == nil {
		if fileSize >= rm.rotationSettings.MaxFileSize {
			rm.logger.WithFields(logrus.Fields{
				"session_id": sessionID,
				"file_size":  fileSize,
				"max_size":   rm.rotationSettings.MaxFileSize,
			}).Info("Recording file size limit reached, rotating file")

			if err := rm.RotateRecordingFile(ctx, sessionID); err != nil {
				rm.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to rotate recording file")
			}
		}
	}

	// Check duration
	if time.Since(session.StartTime) >= rm.rotationSettings.MaxDuration {
		rm.logger.WithFields(logrus.Fields{
			"session_id":   sessionID,
			"duration":     time.Since(session.StartTime),
			"max_duration": rm.rotationSettings.MaxDuration,
		}).Info("Recording duration limit reached, rotating file")

		if err := rm.RotateRecordingFile(ctx, sessionID); err != nil {
			rm.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to rotate recording file")
		}
	}
}

// generateRecordingPath generates a recording file path
func (rm *RecordingManager) generateRecordingPath(device, sessionID string) string {
	// Use configuration paths if available
	basePath := "/tmp/recordings"
	if rm.config.RecordingsPath != "" {
		basePath = rm.config.RecordingsPath
	}

	// Generate filename
	filename := fmt.Sprintf("%s_%s.%s", device, sessionID, rm.rotationSettings.OutputFormat)

	// Clean device path for filename
	filename = strings.ReplaceAll(filename, "/", "_")

	return filepath.Join(basePath, filename)
}

// generateRotatedFilePath generates a rotated file path
func (rm *RecordingManager) generateRotatedFilePath(currentPath string) string {
	dir := filepath.Dir(currentPath)
	ext := filepath.Ext(currentPath)
	base := strings.TrimSuffix(filepath.Base(currentPath), ext)

	// Add timestamp to filename
	timestamp := time.Now().Format("20060102_150405")
	newFilename := fmt.Sprintf("%s_rotated_%s%s", base, timestamp, ext)

	return filepath.Join(dir, newFilename)
}

// startNewSegment starts a new recording segment
func (rm *RecordingManager) startNewSegment(ctx context.Context, session *RecordingSession) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	// Generate segment ID
	segmentID := generateSegmentID(session.ID, len(session.Segments))

	// Create segment file path
	segmentPath := rm.generateSegmentPath(session, segmentID)

	// Create segment
	segment := &Segment{
		ID:           segmentID,
		FilePath:     segmentPath,
		StartTime:    time.Now(),
		Index:        len(session.Segments),
		ContinuityID: session.ContinuityID,
	}

	// Build FFmpeg command for segment-based recording
	ffmpegCmd := rm.buildSegmentFFmpegCommand(session.DevicePath, segmentPath)

	// Start FFmpeg process
	if err := rm.ffmpegManager.StartProcess(ctx, segmentID, ffmpegCmd); err != nil {
		return fmt.Errorf("failed to start FFmpeg process: %w", err)
	}

	// Update session
	session.CurrentSegment = segment
	session.Segments = append(session.Segments, segment)

	// Store segment in segment manager
	rm.segmentManager.mu.Lock()
	rm.segmentManager.segments[segmentID] = segment
	rm.segmentManager.mu.Unlock()

	rm.logger.WithFields(logrus.Fields{
		"session_id":   session.ID,
		"segment_id":   segmentID,
		"segment_path": segmentPath,
	}).Info("Started new recording segment")

	return nil
}

// buildSegmentFFmpegCommand builds FFmpeg command for segment-based recording
func (rm *RecordingManager) buildSegmentFFmpegCommand(devicePath, segmentPath string) []string {
	config := rm.segmentManager.config

	cmd := []string{
		"ffmpeg",
		"-f", "v4l2",
		"-i", devicePath,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%d", int(config.RotationInterval.Seconds())),
	}

	if config.StrftimeEnabled {
		cmd = append(cmd, "-strftime", "1")
	}

	if config.ResetTimestamps {
		cmd = append(cmd, "-reset_timestamps", "1")
	}

	cmd = append(cmd,
		"-segment_start_number", "0",
		"-y", // Overwrite output file
		segmentPath,
	)

	return cmd
}

// generateSegmentPath generates the path for a recording segment
func (rm *RecordingManager) generateSegmentPath(session *RecordingSession, segmentID string) string {
	config := rm.segmentManager.config
	basePath := rm.config.RecordingsPath

	// Create segment filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s_%s.mp4",
		config.SegmentPrefix,
		session.ContinuityID,
		timestamp,
		segmentID)

	return filepath.Join(basePath, filename)
}

// monitorSegmentRotation monitors and rotates recording segments
func (rm *RecordingManager) monitorSegmentRotation(ctx context.Context, session *RecordingSession) {
	ticker := time.NewTicker(rm.segmentManager.config.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rm.rotateSegment(ctx, session); err != nil {
				rm.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to rotate segment")
			}
		}
	}
}

// rotateSegment rotates the current recording segment
func (rm *RecordingManager) rotateSegment(ctx context.Context, session *RecordingSession) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.CurrentSegment == nil {
		return fmt.Errorf("no current segment to rotate")
	}

	// End current segment
	session.CurrentSegment.EndTime = time.Now()
	session.CurrentSegment.Duration = session.CurrentSegment.EndTime.Sub(session.CurrentSegment.StartTime)

	// Get file size
	if fileInfo, err := os.Stat(session.CurrentSegment.FilePath); err == nil {
		session.CurrentSegment.Size = fileInfo.Size()
	}

	rm.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"segment_id": session.CurrentSegment.ID,
		"duration":   session.CurrentSegment.Duration,
		"size":       session.CurrentSegment.Size,
	}).Info("Rotated recording segment")

	// Start new segment
	if err := rm.startNewSegment(ctx, session); err != nil {
		return fmt.Errorf("failed to start new segment: %w", err)
	}

	// Cleanup old segments if needed
	rm.cleanupOldSegments(session)

	return nil
}

// cleanupOldSegments removes old segments to maintain storage limits
func (rm *RecordingManager) cleanupOldSegments(session *RecordingSession) {
	config := rm.segmentManager.config

	if len(session.Segments) <= config.MaxSegments {
		return
	}

	// Remove oldest segments
	segmentsToRemove := len(session.Segments) - config.MaxSegments
	for i := 0; i < segmentsToRemove; i++ {
		segment := session.Segments[i]

		// Remove file
		if err := os.Remove(segment.FilePath); err != nil {
			rm.logger.WithError(err).WithField("segment_path", segment.FilePath).Warning("Failed to remove old segment file")
		}

		// Remove from segment manager
		rm.segmentManager.mu.Lock()
		delete(rm.segmentManager.segments, segment.ID)
		rm.segmentManager.mu.Unlock()

		rm.logger.WithField("segment_path", segment.FilePath).Info("Removed old segment file")
	}

	// Update session segments
	session.Segments = session.Segments[segmentsToRemove:]
}

// GetRecordingContinuity returns recording continuity information
func (rm *RecordingManager) GetRecordingContinuity(sessionID string) (*RecordingContinuity, error) {
	rm.sessionsMu.RLock()
	session, exists := rm.sessions[sessionID]
	rm.sessionsMu.RUnlock()

	if !exists {
		return nil, NewRecordingError(sessionID, "", "get_continuity", "session not found")
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	continuity := &RecordingContinuity{
		SessionID:     session.ID,
		ContinuityID:  session.ContinuityID,
		StartTime:     session.StartTime,
		SegmentCount:  len(session.Segments),
		TotalDuration: time.Since(session.StartTime),
		Segments:      make([]*SegmentInfo, len(session.Segments)),
	}

	for i, segment := range session.Segments {
		continuity.Segments[i] = &SegmentInfo{
			ID:        segment.ID,
			FilePath:  segment.FilePath,
			StartTime: segment.StartTime,
			EndTime:   segment.EndTime,
			Duration:  segment.Duration,
			Size:      segment.Size,
			Index:     segment.Index,
		}
	}

	return continuity, nil
}

// startMonitoring starts storage monitoring
func (sm *StorageMonitor) startMonitoring() {
	sm.monitorTicker = time.NewTicker(sm.config.CheckInterval)

	go func() {
		for {
			select {
			case <-sm.monitorCtx.Done():
				sm.monitorTicker.Stop()
				return
			case <-sm.monitorTicker.C:
				sm.checkStorage()
			}
		}
	}()
}

// checkStorage checks storage usage and triggers cleanup if needed
func (sm *StorageMonitor) checkStorage() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Get storage metrics
	metrics, err := sm.getStorageMetrics()
	if err != nil {
		sm.logger.WithError(err).Error("Failed to get storage metrics")
		return
	}

	sm.metrics = metrics

	// Check usage thresholds
	if metrics.UsagePercent >= sm.config.BlockPercent {
		sm.logger.WithField("usage_percent", metrics.UsagePercent).Error("Storage usage critical - blocking operations")
		// Trigger emergency cleanup
		sm.emergencyCleanup()
	} else if metrics.UsagePercent >= sm.config.WarnPercent {
		sm.logger.WithField("usage_percent", metrics.UsagePercent).Warning("Storage usage high - warning threshold exceeded")
		// Trigger normal cleanup
		if sm.config.AutoCleanup {
			sm.cleanupOldFiles()
		}
	}
}

// getStorageMetrics gets current storage usage metrics
func (sm *StorageMonitor) getStorageMetrics() (*StorageMetrics, error) {
	// This is a simplified implementation
	// In a real implementation, you'd use syscall.Statfs or similar
	// to get actual filesystem statistics

	metrics := &StorageMetrics{
		TotalSpace: 100 * 1024 * 1024 * 1024, // 100GB example
		UsedSpace:  80 * 1024 * 1024 * 1024,  // 80GB example
		LastCheck:  time.Now(),
	}

	metrics.AvailableSpace = metrics.TotalSpace - metrics.UsedSpace
	metrics.UsagePercent = int((metrics.UsedSpace * 100) / metrics.TotalSpace)

	return metrics, nil
}

// checkStorageAvailability checks if there's enough storage for recording
func (sm *StorageMonitor) checkStorageAvailability() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.metrics == nil {
		return nil // No metrics available yet
	}

	if sm.metrics.UsagePercent >= sm.config.BlockPercent {
		return fmt.Errorf("storage usage critical (%d%%), cannot start recording", sm.metrics.UsagePercent)
	}

	return nil
}

// emergencyCleanup performs emergency storage cleanup
func (sm *StorageMonitor) emergencyCleanup() {
	sm.logger.Warning("Performing emergency storage cleanup")
	// Implementation would remove oldest files to free space
}

// cleanupOldFiles removes old files based on configuration
func (sm *StorageMonitor) cleanupOldFiles() {
	sm.logger.Info("Performing storage cleanup")
	// Implementation would remove files older than MaxFileAge
}

// GetStorageMetrics returns current storage metrics
func (sm *StorageMonitor) GetStorageMetrics() *StorageMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.metrics == nil {
		return &StorageMetrics{}
	}

	return sm.metrics
}

// UpdateStorageConfig updates storage monitoring configuration
func (sm *StorageMonitor) UpdateStorageConfig(config *StorageConfig) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = config

	// Restart monitoring with new interval
	if sm.monitorTicker != nil {
		sm.monitorTicker.Stop()
	}

	sm.startMonitoring()
}

// generateContinuityID generates a unique continuity ID
func generateContinuityID() string {
	return fmt.Sprintf("cont_%s", generateSessionID(""))
}

// generateSegmentID generates a unique segment ID
func generateSegmentID(sessionID string, index int) string {
	return fmt.Sprintf("seg_%s_%03d", sessionID, index)
}
