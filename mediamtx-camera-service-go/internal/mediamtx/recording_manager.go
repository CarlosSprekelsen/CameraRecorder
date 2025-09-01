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
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"syscall"

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

	// Device to session mapping for efficient lookup
	deviceToSession map[string]string // device path -> session ID
	deviceMu        sync.RWMutex

	// File rotation settings
	rotationSettings *RotationSettings

	// Storage monitoring
	storageMonitor *StorageMonitor

	// Segment management
	segmentManager *SegmentManager

	// Storage thresholds (Phase 2 enhancement)
	storageWarnPercent  int
	storageBlockPercent int
}

// RotationSettings defines file rotation behavior
type RotationSettings struct {
	MaxFileSize     int64         `json:"max_file_size"`
	MaxDuration     time.Duration `json:"max_duration"`
	SegmentDuration time.Duration `json:"segment_duration"`
	AutoRotate      bool          `json:"auto_rotate"`
	KeepSegments    int           `json:"keep_segments"`
	OutputFormat    string        `json:"output_format"`

	// Enhanced segment-based rotation (Phase 3 enhancement)
	ContinuityMode  bool   `json:"continuity_mode"`  // Enable recording continuity across segments
	SegmentIndex    int    `json:"segment_index"`    // Current segment index
	SegmentFormat   string `json:"segment_format"`   // Segment filename format
	ResetTimestamps bool   `json:"reset_timestamps"` // Reset timestamps for each segment
	StrftimeEnabled bool   `json:"strftime_enabled"` // Enable strftime in segment names
	SegmentPrefix   string `json:"segment_prefix"`   // Prefix for segment files
	MaxSegments     int    `json:"max_segments"`     // Maximum number of segments to keep
	SegmentRotation bool   `json:"segment_rotation"` // Enable automatic segment rotation
	ContinuityID    string `json:"continuity_id"`    // Continuity identifier for segments
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
// Note: RecordingSession is defined in types.go

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
		ffmpegManager:   ffmpegManager,
		config:          config,
		logger:          logger,
		sessions:        make(map[string]*RecordingSession),
		deviceToSession: make(map[string]string),
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
		storageWarnPercent:  80,
		storageBlockPercent: 90,
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

	// Storage validation check (Phase 2 enhancement)
	if _, err := rm.checkStorageSpace(); err != nil {
		rm.logger.WithError(err).WithField("device", device).Error("Storage validation failed")
		return nil, NewRecordingError(sessionID, device, "start_recording", "storage validation failed: "+err.Error())
	}

	// Create recording session with enhanced use case management (Phase 2 enhancement)
	session := &RecordingSession{
		ID:           sessionID,
		Device:       device,
		Path:         path,
		Status:       "RECORDING",
		StartTime:    time.Now(),
		FilePath:     rm.generateRecordingPath(device, sessionID),
		ContinuityID: generateContinuityID(),
		State:        SessionStateRecording,

		// Enhanced use case management (Phase 2 enhancement)
		UseCase:       UseCaseRecording,  // Default to recording use case
		Priority:      2,                 // Default medium priority
		AutoCleanup:   true,              // Default auto-cleanup
		RetentionDays: 7,                 // Default 7 days retention
		Quality:       "medium",          // Default medium quality
		MaxDuration:   24 * time.Hour,    // Default 24 hour max duration
		AutoRotate:    true,              // Default auto-rotate
		RotationSize:  100 * 1024 * 1024, // Default 100MB rotation size
	}

	// Apply use case specific configuration (Phase 2 enhancement)
	if useCase, ok := options["use_case"].(string); ok {
		switch useCase {
		case "recording":
			session.UseCase = UseCaseRecording
			session.Priority = 1
			session.AutoCleanup = true
			session.RetentionDays = 30
			session.Quality = "high"
			session.MaxDuration = 24 * time.Hour
			session.AutoRotate = true
			session.RotationSize = 100 * 1024 * 1024 // 100MB
		case "viewing":
			session.UseCase = UseCaseViewing
			session.Priority = 2
			session.AutoCleanup = false
			session.RetentionDays = 1
			session.Quality = "medium"
			session.MaxDuration = 2 * time.Hour
			session.AutoRotate = false
			session.RotationSize = 0
		case "snapshot":
			session.UseCase = UseCaseSnapshot
			session.Priority = 3
			session.AutoCleanup = true
			session.RetentionDays = 7
			session.Quality = "low"
			session.MaxDuration = 1 * time.Hour
			session.AutoRotate = false
			session.RotationSize = 0
		}
	}

	// Override with explicit options if provided (Phase 2 enhancement)
	if priority, ok := options["priority"].(int); ok {
		session.Priority = priority
	}
	if autoCleanup, ok := options["auto_cleanup"].(bool); ok {
		session.AutoCleanup = autoCleanup
	}
	if retentionDays, ok := options["retention_days"].(int); ok {
		session.RetentionDays = retentionDays
	}
	if quality, ok := options["quality"].(string); ok {
		session.Quality = quality
	}
	if maxDuration, ok := options["max_duration"].(time.Duration); ok {
		session.MaxDuration = maxDuration
	}
	if autoRotate, ok := options["auto_rotate"].(bool); ok {
		session.AutoRotate = autoRotate
	}
	if rotationSize, ok := options["rotation_size"].(int64); ok {
		session.RotationSize = rotationSize
	}

	// Apply enhanced rotation settings from options (Phase 3 enhancement)
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

	// Enhanced segment-based rotation settings (Phase 3 enhancement)
	if continuityMode, ok := options["continuity_mode"].(bool); ok {
		rm.rotationSettings.ContinuityMode = continuityMode
	}
	if segmentFormat, ok := options["segment_format"].(string); ok && segmentFormat != "" {
		rm.rotationSettings.SegmentFormat = segmentFormat
	}
	if resetTimestamps, ok := options["reset_timestamps"].(bool); ok {
		rm.rotationSettings.ResetTimestamps = resetTimestamps
	}
	if strftimeEnabled, ok := options["strftime_enabled"].(bool); ok {
		rm.rotationSettings.StrftimeEnabled = strftimeEnabled
	}
	if segmentPrefix, ok := options["segment_prefix"].(string); ok && segmentPrefix != "" {
		rm.rotationSettings.SegmentPrefix = segmentPrefix
	}
	if maxSegments, ok := options["max_segments"].(int); ok && maxSegments > 0 {
		rm.rotationSettings.MaxSegments = maxSegments
	}
	if segmentRotation, ok := options["segment_rotation"].(bool); ok {
		rm.rotationSettings.SegmentRotation = segmentRotation
	}

	// Set continuity ID for segment-based recording (Phase 3 enhancement)
	if rm.rotationSettings.ContinuityMode {
		rm.rotationSettings.ContinuityID = session.ContinuityID
	}

	// Enhanced FFmpeg command building with segment support (Phase 3 enhancement)
	var err error
	var pid int

	if rm.rotationSettings.ContinuityMode && rm.rotationSettings.SegmentRotation {
		// Use segmented recording with continuity support
		rm.logger.WithFields(logrus.Fields{
			"session_id":       sessionID,
			"continuity_id":    session.ContinuityID,
			"segment_duration": rm.rotationSettings.SegmentDuration,
		}).Info("Starting segmented recording with continuity")

		err = rm.ffmpegManager.CreateSegmentedRecording(ctx, device, session.FilePath, rm.rotationSettings)
		if err != nil {
			return nil, NewRecordingErrorWithErr(sessionID, device, "start_recording", "failed to start segmented recording", err)
		}
		pid = 0 // Segmented recording doesn't return PID immediately
	} else {
		// Use standard FFmpeg recording
		command := rm.buildAdvancedRecordingCommand(device, session.FilePath, options)
		pid, err = rm.ffmpegManager.StartProcess(ctx, command, session.FilePath)
		if err != nil {
			// Enhanced error categorization and logging (Phase 4 enhancement)
			enhancedErr := CategorizeError(err)
			errorMetadata := GetErrorMetadata(enhancedErr)
			recoveryStrategies := GetRecoveryStrategies(enhancedErr.GetCategory())

			rm.logger.WithFields(logrus.Fields{
				"session_id":          sessionID,
				"device":              device,
				"error_category":      errorMetadata["category"],
				"error_severity":      errorMetadata["severity"],
				"retryable":           errorMetadata["retryable"],
				"recoverable":         errorMetadata["recoverable"],
				"recovery_strategies": recoveryStrategies,
			}).Error("Failed to start FFmpeg process with enhanced error categorization")

			return nil, NewRecordingErrorWithErr(sessionID, device, "start_recording", "failed to start FFmpeg process", err)
		}
	}

	// Update session with PID
	session.Status = "RECORDING"
	session.PID = pid

	// Store session and update device mapping
	rm.sessionsMu.Lock()
	rm.sessions[sessionID] = session
	rm.sessionsMu.Unlock()

	// Add device to session mapping
	rm.addDeviceSessionMapping(device, sessionID)

	// Start monitoring if auto-rotation is enabled
	if rm.rotationSettings.AutoRotate {
		go rm.monitorRecordingForRotation(ctx, sessionID)
	}

	// Schedule auto-stop if max duration is specified (Phase 3 enhancement)
	if session.MaxDuration > 0 {
		rm.scheduleAutoStop(ctx, sessionID, session.MaxDuration)
		rm.logger.WithFields(logrus.Fields{
			"session_id":   sessionID,
			"max_duration": session.MaxDuration,
		}).Debug("Auto-stop scheduled for recording")
	}

	rm.logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"device":       device,
		"path":         path,
		"pid":          pid,
		"max_duration": session.MaxDuration,
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

	// Stop FFmpeg process using stored PID
	if session.PID > 0 {
		if err := rm.ffmpegManager.StopProcess(ctx, session.PID); err != nil {
			rm.logger.WithError(err).WithFields(logrus.Fields{
				"session_id": sessionID,
				"pid":        session.PID,
			}).Warning("Failed to stop FFmpeg process")
		}
	} else {
		rm.logger.WithField("session_id", sessionID).Warning("No PID stored for session, cannot stop FFmpeg process")
	}

	// Remove session and device mapping
	rm.sessionsMu.Lock()
	delete(rm.sessions, sessionID)
	rm.sessionsMu.Unlock()

	// Remove device to session mapping
	rm.removeDeviceSessionMapping(session.Device)

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

	// Stop FFmpeg process using stored PID
	if session.PID > 0 {
		if err := rm.ffmpegManager.StopProcess(ctx, session.PID); err != nil {
			rm.logger.WithError(err).WithFields(logrus.Fields{
				"session_id": sessionID,
				"pid":        session.PID,
			}).Warning("Failed to stop FFmpeg process")
		}
	} else {
		rm.logger.WithField("session_id", sessionID).Warning("No PID stored for session, cannot stop FFmpeg process")
	}

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

	// Enhanced use case specific cleanup (Phase 2 enhancement)
	rm.logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"device":       session.Device,
		"duration":     session.Duration,
		"file_size":    session.FileSize,
		"use_case":     session.UseCase,
		"priority":     session.Priority,
		"auto_cleanup": session.AutoCleanup,
	}).Info("Advanced recording stopped successfully")

	// Perform use case specific cleanup (Phase 2 enhancement)
	go rm.performUseCaseCleanup(ctx, session)

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
	ffmpegCmd := rm.buildSegmentFFmpegCommand(session.Device, segmentPath)

	// Start FFmpeg process
	_, err := rm.ffmpegManager.StartProcess(ctx, ffmpegCmd, segmentPath)
	if err != nil {
		return fmt.Errorf("failed to start FFmpeg process: %w", err)
	}

	// Update session - store segment ID in segments list
	session.Segments = append(session.Segments, segmentID)

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

	if len(session.Segments) == 0 {
		return fmt.Errorf("no segments to rotate")
	}

	// Get current segment ID
	currentSegmentID := session.Segments[len(session.Segments)-1]

	// Get segment from manager
	rm.segmentManager.mu.RLock()
	currentSegment, exists := rm.segmentManager.segments[currentSegmentID]
	rm.segmentManager.mu.RUnlock()

	if !exists {
		return fmt.Errorf("current segment not found: %s", currentSegmentID)
	}

	// End current segment
	currentSegment.EndTime = time.Now()
	currentSegment.Duration = currentSegment.EndTime.Sub(currentSegment.StartTime)

	// Get file size
	if fileInfo, err := os.Stat(currentSegment.FilePath); err == nil {
		currentSegment.Size = fileInfo.Size()
	}

	rm.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"segment_id": currentSegment.ID,
		"duration":   currentSegment.Duration,
		"size":       currentSegment.Size,
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
		segmentID := session.Segments[i]

		// Get segment from manager
		rm.segmentManager.mu.RLock()
		segment, exists := rm.segmentManager.segments[segmentID]
		rm.segmentManager.mu.RUnlock()

		if !exists {
			rm.logger.WithField("segment_id", segmentID).Warning("Segment not found in manager")
			continue
		}

		// Remove file
		if err := os.Remove(segment.FilePath); err != nil {
			rm.logger.WithError(err).WithField("segment_path", segment.FilePath).Warning("Failed to remove old segment file")
		}

		// Remove from segment manager
		rm.segmentManager.mu.Lock()
		delete(rm.segmentManager.segments, segmentID)
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

	for i, segmentID := range session.Segments {
		// Get segment from manager
		rm.segmentManager.mu.RLock()
		segment, exists := rm.segmentManager.segments[segmentID]
		rm.segmentManager.mu.RUnlock()

		if !exists {
			rm.logger.WithField("segment_id", segmentID).Warning("Segment not found in manager")
			continue
		}

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

// getStorageMetrics gets current storage usage metrics using proper filesystem calls
func (sm *StorageMonitor) getStorageMetrics() (*StorageMetrics, error) {
	// Get storage path from recordings path in main config
	// Note: This function needs access to the main config, but we'll use a default path for now
	storagePath := "/tmp" // Default fallback
	if sm.config != nil {
		// Try to get recordings path from environment or use default
		if recordingsPath := os.Getenv("RECORDINGS_PATH"); recordingsPath != "" {
			storagePath = recordingsPath
		}
	}

	// Use syscall.Statfs to get actual filesystem statistics
	var stat syscall.Statfs_t
	err := syscall.Statfs(storagePath, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate storage metrics
	// stat.Blocks is total blocks, stat.Bfree is free blocks, stat.Bavail is available blocks
	blockSize := uint64(stat.Bsize)
	totalBlocks := stat.Blocks
	freeBlocks := stat.Bfree
	availableBlocks := stat.Bavail

	// Convert to bytes
	totalSpace := totalBlocks * blockSize
	availableSpace := availableBlocks * blockSize
	usedSpace := totalSpace - (freeBlocks * blockSize)

	// Calculate usage percentage
	usagePercent := 0
	if totalSpace > 0 {
		usagePercent = int((usedSpace * 100) / totalSpace)
	}

	metrics := &StorageMetrics{
		TotalSpace:     int64(totalSpace),
		UsedSpace:      int64(usedSpace),
		AvailableSpace: int64(availableSpace),
		UsagePercent:   usagePercent,
		LastCheck:      time.Now(),
	}

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

// GetRecordingsList scans the recordings directory and returns a list of recording files with metadata
func (rm *RecordingManager) GetRecordingsList(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	rm.logger.WithFields(logrus.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting recordings list")

	// Get recordings directory path from configuration
	recordingsDir := rm.config.RecordingsPath
	if recordingsDir == "" {
		return nil, fmt.Errorf("recordings path not configured")
	}

	// Check if directory exists and is accessible
	if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
		rm.logger.WithField("directory", recordingsDir).Warn("Recordings directory does not exist")
		return &FileListResponse{
			Files:  []*FileMetadata{},
			Total:  0,
			Limit:  limit,
			Offset: offset,
		}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		rm.logger.WithError(err).WithField("directory", recordingsDir).Error("Error reading recordings directory")
		return nil, fmt.Errorf("failed to read recordings directory: %w", err)
	}

	// Process files and extract metadata
	var files []*FileMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Get file stats
		fileInfo, err := entry.Info()
		if err != nil {
			rm.logger.WithError(err).WithField("filename", filename).Warn("Error accessing file")
			continue
		}

		// Determine if it's a video file
		isVideo := false
		ext := filepath.Ext(filename)
		switch ext {
		case ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv":
			isVideo = true
		}

		// Create file metadata
		fileMetadata := &FileMetadata{
			FileName:    filename,
			FileSize:    fileInfo.Size(),
			CreatedAt:   fileInfo.ModTime(), // Use ModTime as CreatedAt since creation time may not be available
			ModifiedAt:  fileInfo.ModTime(),
			DownloadURL: fmt.Sprintf("/files/recordings/%s", filename),
		}

		// Add duration for video files with comprehensive metadata extraction
		if isVideo {
			duration, err := rm.extractVideoDuration(ctx, filepath.Join(recordingsDir, filename))
			if err != nil {
				rm.logger.WithError(err).WithField("filename", filename).Warn("Failed to extract video duration")
				fileMetadata.Duration = nil
			} else {
				fileMetadata.Duration = &duration
			}
		}

		files = append(files, fileMetadata)
	}

	// Sort files by modified time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})

	// Apply pagination
	totalCount := len(files)
	startIdx := offset
	endIdx := startIdx + limit
	if endIdx > totalCount {
		endIdx = totalCount
	}
	if startIdx > totalCount {
		startIdx = totalCount
	}

	paginatedFiles := files[startIdx:endIdx]

	rm.logger.WithFields(logrus.Fields{
		"total_files": totalCount,
		"returned":    len(paginatedFiles),
	}).Debug("Recordings list retrieved successfully")

	return &FileListResponse{
		Files:  paginatedFiles,
		Total:  totalCount,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// extractVideoDuration extracts duration from video file using FFmpeg
func (rm *RecordingManager) extractVideoDuration(ctx context.Context, filePath string) (int64, error) {
	rm.logger.WithField("file_path", filePath).Debug("Extracting video duration")

	// Use FFmpeg to get video duration
	command := []string{
		"ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filePath,
	}

	// Execute command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to extract video duration: %w", err)
	}

	// Parse duration from output
	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	// Convert to seconds (int64)
	durationSeconds := int64(duration)

	rm.logger.WithFields(logrus.Fields{
		"file_path": filePath,
		"duration":  durationSeconds,
	}).Debug("Video duration extracted successfully")

	return durationSeconds, nil
}

// GetRecordingInfo gets detailed information about a specific recording file
func (rm *RecordingManager) GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	rm.logger.WithField("filename", filename).Debug("Getting recording info")

	// Validate filename
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// Get recordings directory path from configuration
	recordingsDir := rm.config.RecordingsPath
	if recordingsDir == "" {
		return nil, fmt.Errorf("recordings path not configured")
	}

	// Construct full file path
	filePath := filepath.Join(recordingsDir, filename)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("recording file not found: %s", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("error accessing file: %w", err)
	}

	// Check if it's a file (not a directory)
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a file: %s", filename)
	}

	// Determine if it's a video file
	isVideo := false
	ext := filepath.Ext(filename)
	switch ext {
	case ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv":
		isVideo = true
	}

	// Create file metadata
	fileMetadata := &FileMetadata{
		FileName:    filename,
		FileSize:    fileInfo.Size(),
		CreatedAt:   fileInfo.ModTime(), // Use ModTime as CreatedAt since creation time may not be available
		ModifiedAt:  fileInfo.ModTime(),
		DownloadURL: fmt.Sprintf("/files/recordings/%s", filename),
	}

	// Add duration for video files using FFmpeg metadata extraction
	if isVideo {
		duration, err := rm.extractVideoDuration(ctx, filePath)
		if err != nil {
			rm.logger.WithError(err).WithField("filename", filename).Warn("Failed to extract video duration, setting to nil")
			fileMetadata.Duration = nil
		} else {
			fileMetadata.Duration = &duration
		}
	}

	rm.logger.WithField("filename", filename).Debug("Recording info retrieved successfully")
	return fileMetadata, nil
}

// DeleteRecording deletes a recording file
func (rm *RecordingManager) DeleteRecording(ctx context.Context, filename string) error {
	rm.logger.WithField("filename", filename).Debug("Deleting recording")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Get recordings directory path from configuration
	recordingsDir := rm.config.RecordingsPath
	if recordingsDir == "" {
		return fmt.Errorf("recordings path not configured")
	}

	// Construct full file path
	filePath := filepath.Join(recordingsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("recording file not found: %s", filename)
	}

	// Check if it's a file (not a directory)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error accessing file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is not a file: %s", filename)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		rm.logger.WithError(err).WithField("filename", filename).Error("Error deleting recording file")
		return fmt.Errorf("error deleting recording file: %w", err)
	}

	rm.logger.WithField("filename", filename).Info("Recording file deleted successfully")
	return nil
}

// performUseCaseCleanup performs use case specific cleanup operations (Phase 2 enhancement)
func (rm *RecordingManager) performUseCaseCleanup(ctx context.Context, session *RecordingSession) {
	rm.logger.WithFields(logrus.Fields{
		"session_id":   session.ID,
		"use_case":     session.UseCase,
		"priority":     session.Priority,
		"auto_cleanup": session.AutoCleanup,
	}).Debug("Performing use case specific cleanup")

	// Skip cleanup if auto-cleanup is disabled
	if !session.AutoCleanup {
		rm.logger.WithField("session_id", session.ID).Debug("Auto-cleanup disabled, skipping cleanup")
		return
	}

	// Perform use case specific cleanup
	switch session.UseCase {
	case UseCaseRecording:
		// Recording use case: Keep files for longer, high priority
		rm.logger.WithField("session_id", session.ID).Debug("Performing recording use case cleanup")
		// Recording files are kept for retention period, no immediate cleanup

	case UseCaseViewing:
		// Viewing use case: Clean up quickly, medium priority
		rm.logger.WithField("session_id", session.ID).Debug("Performing viewing use case cleanup")
		// Viewing files are cleaned up after short retention period
		go rm.scheduleViewingCleanup(ctx, session)

	case UseCaseSnapshot:
		// Snapshot use case: Clean up after medium retention, low priority
		rm.logger.WithField("session_id", session.ID).Debug("Performing snapshot use case cleanup")
		// Snapshot files are cleaned up after medium retention period
		go rm.scheduleSnapshotCleanup(ctx, session)

	default:
		rm.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
			"use_case":   session.UseCase,
		}).Warn("Unknown use case, skipping cleanup")
	}
}

// scheduleViewingCleanup schedules cleanup for viewing use case files (Phase 2 enhancement)
func (rm *RecordingManager) scheduleViewingCleanup(ctx context.Context, session *RecordingSession) {
	// Viewing files are cleaned up after 1 day
	cleanupDelay := 24 * time.Hour

	rm.logger.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"cleanup_delay": cleanupDelay,
	}).Debug("Scheduling viewing cleanup")

	time.Sleep(cleanupDelay)

	// Check if file still exists and delete if retention period exceeded
	if err := rm.cleanupExpiredFile(ctx, session); err != nil {
		rm.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to cleanup viewing file")
	}
}

// scheduleSnapshotCleanup schedules cleanup for snapshot use case files (Phase 2 enhancement)
func (rm *RecordingManager) scheduleSnapshotCleanup(ctx context.Context, session *RecordingSession) {
	// Snapshot files are cleaned up after retention period
	cleanupDelay := time.Duration(session.RetentionDays) * 24 * time.Hour

	rm.logger.WithFields(logrus.Fields{
		"session_id":     session.ID,
		"cleanup_delay":  cleanupDelay,
		"retention_days": session.RetentionDays,
	}).Debug("Scheduling snapshot cleanup")

	time.Sleep(cleanupDelay)

	// Check if file still exists and delete if retention period exceeded
	if err := rm.cleanupExpiredFile(ctx, session); err != nil {
		rm.logger.WithError(err).WithField("session_id", session.ID).Error("Failed to cleanup snapshot file")
	}
}

// cleanupExpiredFile cleans up expired files based on retention policy (Phase 2 enhancement)
func (rm *RecordingManager) cleanupExpiredFile(ctx context.Context, session *RecordingSession) error {
	// Check if file exists
	if _, err := os.Stat(session.FilePath); os.IsNotExist(err) {
		rm.logger.WithField("session_id", session.ID).Debug("File already deleted, skipping cleanup")
		return nil
	}

	// Check if file is older than retention period
	fileInfo, err := os.Stat(session.FilePath)
	if err != nil {
		return fmt.Errorf("error accessing file: %w", err)
	}

	retentionPeriod := time.Duration(session.RetentionDays) * 24 * time.Hour
	fileAge := time.Since(fileInfo.ModTime())

	if fileAge > retentionPeriod {
		rm.logger.WithFields(logrus.Fields{
			"session_id":       session.ID,
			"file_age":         fileAge,
			"retention_period": retentionPeriod,
		}).Info("File exceeds retention period, deleting")

		if err := os.Remove(session.FilePath); err != nil {
			return fmt.Errorf("error deleting expired file: %w", err)
		}

		rm.logger.WithField("session_id", session.ID).Info("Expired file deleted successfully")
	} else {
		rm.logger.WithFields(logrus.Fields{
			"session_id":       session.ID,
			"file_age":         fileAge,
			"retention_period": retentionPeriod,
		}).Debug("File within retention period, keeping")
	}

	return nil
}

// Storage validation methods (Phase 2 enhancement)

// checkStorageSpace checks available storage space and returns status
func (rm *RecordingManager) checkStorageSpace() (*StorageInfo, error) {
	// Get storage path from config
	storagePath := "/opt/camera-service/recordings"
	if rm.config != nil && rm.config.RecordingsPath != "" {
		storagePath = rm.config.RecordingsPath
	}

	// Get file system statistics
	var stat syscall.Statfs_t
	if err := syscall.Statfs(storagePath, &stat); err != nil {
		return nil, fmt.Errorf("failed to get storage info: %w", err)
	}

	// Calculate storage metrics
	totalSpace := int64(stat.Blocks) * int64(stat.Bsize)
	availableSpace := int64(stat.Bavail) * int64(stat.Bsize)
	usedSpace := totalSpace - availableSpace
	usagePercent := int((usedSpace * 100) / totalSpace)

	storageInfo := &StorageInfo{
		TotalSpace:     totalSpace,
		AvailableSpace: availableSpace,
		UsedSpace:      usedSpace,
		UsagePercent:   usagePercent,
		WarnThreshold:  rm.storageWarnPercent,
		BlockThreshold: rm.storageBlockPercent,
	}

	// Check against thresholds
	if usagePercent >= rm.storageBlockPercent {
		return storageInfo, fmt.Errorf("storage space critical (%d%% used)", usagePercent)
	}

	if usagePercent >= rm.storageWarnPercent {
		rm.logger.WithFields(logrus.Fields{
			"usage_percent":  usagePercent,
			"warn_threshold": rm.storageWarnPercent,
			"storage_path":   storagePath,
		}).Warn("Storage usage high")
	}

	return storageInfo, nil
}

// UpdateStorageThresholds updates storage thresholds from configuration (Phase 4 enhancement)
func (rm *RecordingManager) UpdateStorageThresholds(warnPercent, blockPercent int) {
	rm.storageWarnPercent = warnPercent
	rm.storageBlockPercent = blockPercent

	rm.logger.WithFields(logrus.Fields{
		"warn_percent":  warnPercent,
		"block_percent": blockPercent,
	}).Info("Storage thresholds updated from configuration")
}

// Auto-stop functionality methods (Phase 3 enhancement)

// scheduleAutoStop schedules automatic stopping of a recording after a specified duration
func (rm *RecordingManager) scheduleAutoStop(ctx context.Context, sessionID string, duration time.Duration) {
	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Auto-stop the recording
			if err := rm.StopRecording(ctx, sessionID); err != nil {
				rm.logger.WithError(err).WithField("session_id", sessionID).Error("Auto-stop failed")
			} else {
				rm.logger.WithField("session_id", sessionID).Info("Recording auto-stopped after duration")
			}
		case <-ctx.Done():
			// Context cancelled, don't auto-stop
			rm.logger.WithField("session_id", sessionID).Debug("Auto-stop cancelled due to context cancellation")
			return
		}
	}()
}

// StorageInfo represents storage space information
type StorageInfo struct {
	TotalSpace     int64 `json:"total_space"`
	AvailableSpace int64 `json:"available_space"`
	UsedSpace      int64 `json:"used_space"`
	UsagePercent   int   `json:"usage_percent"`
	WarnThreshold  int   `json:"warn_threshold"`
	BlockThreshold int   `json:"block_threshold"`
}

// Device to session mapping helper methods

// addDeviceSessionMapping adds a device to session mapping
func (rm *RecordingManager) addDeviceSessionMapping(device, sessionID string) {
	rm.deviceMu.Lock()
	defer rm.deviceMu.Unlock()
	rm.deviceToSession[device] = sessionID
}

// removeDeviceSessionMapping removes a device to session mapping
func (rm *RecordingManager) removeDeviceSessionMapping(device string) {
	rm.deviceMu.Lock()
	defer rm.deviceMu.Unlock()
	delete(rm.deviceToSession, device)
}

// getSessionIDByDevice gets session ID by device path
func (rm *RecordingManager) getSessionIDByDevice(device string) (string, bool) {
	rm.deviceMu.RLock()
	defer rm.deviceMu.RUnlock()
	sessionID, exists := rm.deviceToSession[device]
	return sessionID, exists
}

// GetSessionByDevice gets a recording session by device path
func (rm *RecordingManager) GetSessionByDevice(device string) (*RecordingSession, bool) {
	sessionID, exists := rm.getSessionIDByDevice(device)
	if !exists {
		return nil, false
	}

	rm.sessionsMu.RLock()
	defer rm.sessionsMu.RUnlock()
	session, exists := rm.sessions[sessionID]
	return session, exists
}

// CleanupOldRecordings cleans up old recording files based on age and count limits
func (rm *RecordingManager) CleanupOldRecordings(ctx context.Context, maxAge time.Duration, maxCount int) error {
	rm.logger.WithFields(logrus.Fields{
		"max_age":   maxAge,
		"max_count": maxCount,
	}).Debug("Starting cleanup of old recordings")

	// Get recordings list
	recordings, err := rm.GetRecordingsList(ctx, 1000, 0) // Get all recordings for cleanup
	if err != nil {
		return fmt.Errorf("failed to get recordings list: %w", err)
	}

	if len(recordings.Files) == 0 {
		rm.logger.Debug("No recordings found for cleanup")
		return nil
	}

	// Sort recordings by modification time (oldest first)
	sort.Slice(recordings.Files, func(i, j int) bool {
		return recordings.Files[i].ModifiedAt.Before(recordings.Files[j].ModifiedAt)
	})

	cutoffTime := time.Now().Add(-maxAge)
	var deletedCount int
	var deletedSize int64

	// Delete old recordings based on age
	for _, file := range recordings.Files {
		if file.ModifiedAt.Before(cutoffTime) {
			if err := rm.DeleteRecording(ctx, file.FileName); err != nil {
				rm.logger.WithError(err).WithField("filename", file.FileName).Warn("Failed to delete old recording")
				continue
			}
			deletedCount++
			deletedSize += file.FileSize
			rm.logger.WithField("filename", file.FileName).Debug("Deleted old recording file")
		}
	}

	// If we still have too many files, delete oldest ones
	if len(recordings.Files)-deletedCount > maxCount {
		remainingFiles := recordings.Files[deletedCount:]
		excessCount := len(remainingFiles) - maxCount
		
		for i := 0; i < excessCount; i++ {
			file := remainingFiles[i]
			if err := rm.DeleteRecording(ctx, file.FileName); err != nil {
				rm.logger.WithError(err).WithField("filename", file.FileName).Warn("Failed to delete excess recording")
				continue
			}
			deletedCount++
			deletedSize += file.FileSize
			rm.logger.WithField("filename", file.FileName).Debug("Deleted excess recording file")
		}
	}

	rm.logger.WithFields(logrus.Fields{
		"deleted_count": deletedCount,
		"deleted_size":  deletedSize,
		"max_age":       maxAge,
		"max_count":     maxCount,
	}).Info("Cleanup of old recordings completed")

	return nil
}
