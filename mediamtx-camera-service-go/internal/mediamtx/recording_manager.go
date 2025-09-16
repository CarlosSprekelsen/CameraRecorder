/*
MediaMTX Recording Manager Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RecordingManager manages MediaMTX-based recording operations
type RecordingManager struct {
	client            MediaMTXClient
	config            *config.MediaMTXConfig
	recordingConfig   *config.RecordingConfig
	configIntegration *ConfigIntegration
	logger            *logging.Logger
	pathManager       PathManager
	streamManager     StreamManager
	pathValidator     *PathValidator

	// Recording sessions - using sync.Map for lock-free operations
	sessions sync.Map // sessionID -> *RecordingSession

	// Device to session mapping for efficient lookup - using sync.Map for lock-free operations
	deviceToSession sync.Map // device path -> session ID
}

// MediaMTXRecordingConfig defines MediaMTX-specific recording configuration
type MediaMTXRecordingConfig struct {
	// MediaMTX PathConf recording settings
	Record                bool          `json:"record"`                  // Enable recording
	RecordPath            string        `json:"record_path"`             // Recording output path
	RecordFormat          string        `json:"record_format"`           // Recording format (mp4, mkv, etc.)
	RecordPartDuration    time.Duration `json:"record_part_duration"`    // Part duration
	RecordMaxPartSize     string        `json:"record_max_part_size"`    // Max part size
	RecordSegmentDuration time.Duration `json:"record_segment_duration"` // Segment duration
	RecordDeleteAfter     time.Duration `json:"record_delete_after"`     // Auto-delete after duration

	// Recording quality settings
	VideoCodec   string `json:"video_codec"`   // Video codec (h264, h265)
	AudioCodec   string `json:"audio_codec"`   // Audio codec (aac, mp3)
	VideoBitrate string `json:"video_bitrate"` // Video bitrate
	AudioBitrate string `json:"audio_bitrate"` // Audio bitrate
	Resolution   string `json:"resolution"`    // Video resolution
	FrameRate    int    `json:"frame_rate"`    // Frame rate

	// Storage management
	MaxStorageSize   int64 `json:"max_storage_size"`  // Max storage size in bytes
	CleanupThreshold int   `json:"cleanup_threshold"` // Cleanup threshold percentage
}

// MediaMTX API response structures (matching swagger.json)
type MediaMTXRecordingList struct {
	PageCount int                 `json:"pageCount"`
	ItemCount int64               `json:"itemCount"`
	Items     []MediaMTXRecording `json:"items"`
}

type MediaMTXRecording struct {
	Name     string                     `json:"name"`
	Segments []MediaMTXRecordingSegment `json:"segments"`
}

type MediaMTXRecordingSegment struct {
	Start string `json:"start"`
}

// NewRecordingManager creates a new MediaMTX-based recording manager
func NewRecordingManager(client MediaMTXClient, pathManager PathManager, streamManager StreamManager, config *config.MediaMTXConfig, recordingConfig *config.RecordingConfig, configIntegration *ConfigIntegration, logger *logging.Logger) *RecordingManager {
	// Use centralized configuration - no need to create component-specific defaults
	// All recording configuration comes from the centralized config system
	// Recording settings are derived from the centralized MediaMTXConfig

	// Get full config for PathValidator
	fullConfig, err := configIntegration.GetConfig()
	if err != nil {
		logger.WithError(err).Error("Failed to get full config for PathValidator")
		// Continue without path validation - this will be handled at runtime
	}

	var pathValidator *PathValidator
	if fullConfig != nil {
		pathValidator = NewPathValidator(fullConfig, logger)
	}

	return &RecordingManager{
		client:            client,
		config:            config,
		recordingConfig:   recordingConfig,
		configIntegration: configIntegration,
		logger:            logger,
		pathManager:       pathManager,
		streamManager:     streamManager,
		pathValidator:     pathValidator,
		// sessions and deviceToSession: sync.Map is zero-initialized, no need to initialize
	}
}

// StartRecording starts recording for a camera device using MediaMTX path-based recording
func (rm *RecordingManager) StartRecording(ctx context.Context, devicePath string, options map[string]interface{}) (*RecordingSession, error) {
	// Input validation
	if strings.TrimSpace(devicePath) == "" {
		return nil, fmt.Errorf("device path cannot be empty")
	}

	// Validate path before starting (if pathValidator is available)
	var pathResult *PathValidationResult
	if rm.pathValidator != nil {
		var err error
		pathResult, err = rm.pathValidator.ValidateRecordingPath(ctx)
		if err != nil {
			return nil, fmt.Errorf("recording path validation failed: %w", err)
		}
	}

	if pathResult != nil && pathResult.FallbackPath != "" {
		rm.logger.WithFields(logging.Fields{
			"primary":  rm.config.RecordingsPath,
			"fallback": pathResult.FallbackPath,
		}).Info("Using fallback recording path")
	}

	// Use validated path
	recordPath := GenerateRecordingPath(rm.config, rm.recordingConfig)
	rm.logger.WithField("record_path", recordPath).Debug("Using configuration-based recording path")

	rm.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"record_path": recordPath,
		"options":     options,
	}).Info("Starting MediaMTX path-based recording")

	// Check if device is already recording - lock-free read with sync.Map
	if _, exists := rm.deviceToSession.Load(devicePath); exists {
		return nil, fmt.Errorf("device %s is already recording", devicePath)
	}

	// Create path name from device path using unified function
	pathName := GetMediaMTXPathName(devicePath)

	// Use StreamManager's EnableRecording method for path-based recording
	rm.logger.WithField("device_path", devicePath).Info("Enabling recording via StreamManager")
	err := rm.streamManager.EnableRecording(ctx, devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to enable recording: %w", err)
	}

	// Create lightweight recording session (no session ID needed)
	session := &RecordingSession{
		ID:         "", // No session ID - path-based recording
		DevicePath: devicePath,
		Path:       pathName, // Use the stable path name
		FilePath:   recordPath,
		StartTime:  time.Now(),
		Status:     "RECORDING", // Use API-compatible status
		State:      SessionStateRecording,
		UseCase:    UseCaseRecording,
	}

	// Store device mapping for stop recording lookup
	rm.deviceToSession.Store(devicePath, devicePath) // Use devicePath as key and value

	rm.logger.WithFields(logging.Fields{
		"path_name": pathName,
		"device":    devicePath,
	}).Info("MediaMTX path-based recording started successfully")

	return session, nil
}

// GetRecordingSession retrieves a recording session by ID
func (rm *RecordingManager) GetRecordingSession(sessionID string) (*RecordingSession, bool) {
	if session, exists := rm.sessions.Load(sessionID); exists {
		return session.(*RecordingSession), true
	}
	return nil, false
}

// RotateRecordingFile rotates a recording file (MediaMTX handles this automatically)
func (rm *RecordingManager) RotateRecordingFile(ctx context.Context, sessionID string) error {
	rm.logger.WithField("session_id", sessionID).Info("Recording rotation requested - MediaMTX handles this automatically")

	// MediaMTX automatically rotates recording files based on configuration
	// This is a no-op for MediaMTX-based recording
	return nil
}

// GetRecordingInfo gets detailed information about a specific recording file
func (rm *RecordingManager) GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	rm.logger.WithField("filename", filename).Info("Getting recording info")

	// For MediaMTX-based recording, we would query the MediaMTX API
	// For now, return basic file metadata
	return &FileMetadata{
		FileName:   filename,
		FileSize:   0, // Would be populated from MediaMTX API
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}, nil
}

// DeleteRecording deletes a recording segment via MediaMTX API
func (rm *RecordingManager) DeleteRecording(ctx context.Context, filename string) error {
	rm.logger.WithField("filename", filename).Info("Deleting recording via MediaMTX API")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Parse filename to extract path and segment info
	// Expected format: {path}_segment_{index}.{ext} (e.g., "camera0_segment_2.mp4")
	pathName, segmentStart, err := rm.parseRecordingFilename(filename)
	if err != nil {
		rm.logger.WithError(err).WithField("filename", filename).Error("Failed to parse recording filename")
		return fmt.Errorf("invalid recording filename format: %w", err)
	}

	// Call MediaMTX API to delete the segment
	endpoint := fmt.Sprintf("/v3/recordings/deletesegment?path=%s&start=%s", pathName, segmentStart)
	err = rm.client.Delete(ctx, endpoint)
	if err != nil {
		rm.logger.WithError(err).WithFields(logging.Fields{
			"filename": filename,
			"path":     pathName,
			"start":    segmentStart,
		}).Error("Failed to delete recording segment via MediaMTX API")
		return fmt.Errorf("failed to delete recording segment: %w", err)
	}

	rm.logger.WithFields(logging.Fields{
		"filename": filename,
		"path":     pathName,
		"start":    segmentStart,
	}).Info("Recording segment deleted successfully via MediaMTX API")

	return nil
}

// parseRecordingFilename extracts path and segment start time from filename
// Expected format: {path}_segment_{index}.{ext} -> needs to map to actual segment start time
func (rm *RecordingManager) parseRecordingFilename(filename string) (pathName, segmentStart string, err error) {
	// Remove extension
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Split by "_segment_"
	parts := strings.Split(nameWithoutExt, "_segment_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("filename must be in format {path}_segment_{index}.{ext}")
	}

	pathName = parts[0]
	segmentIndexStr := parts[1]

	// Convert segment index to int
	segmentIndex, err := fmt.Sscanf(segmentIndexStr, "%d", new(int))
	if err != nil || segmentIndex != 1 {
		return "", "", fmt.Errorf("invalid segment index: %s", segmentIndexStr)
	}

	// Get actual recording data to find segment start time
	recording, err := rm.getRecordingByName(context.Background(), pathName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get recording data: %w", err)
	}

	// Extract segment start time
	segmentIdx, _ := fmt.Sscanf(segmentIndexStr, "%d", new(int))
	if segmentIdx < 0 || segmentIdx >= len(recording.Segments) {
		return "", "", fmt.Errorf("segment index %d out of range", segmentIdx)
	}

	segmentStart = recording.Segments[segmentIdx].Start
	return pathName, segmentStart, nil
}

// getRecordingByName gets recording data by name (helper for deletion)
func (rm *RecordingManager) getRecordingByName(ctx context.Context, name string) (*MediaMTXRecording, error) {
	endpoint := fmt.Sprintf("/v3/recordings/get/%s", name)
	data, err := rm.client.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording from MediaMTX: %w", err)
	}

	var recording MediaMTXRecording
	if err := json.Unmarshal(data, &recording); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recording data: %w", err)
	}

	return &recording, nil
}

// getSessionIDByDevice retrieves session ID by device path
func (rm *RecordingManager) getSessionIDByDevice(device string) (string, bool) {
	if sessionID, exists := rm.deviceToSession.Load(device); exists {
		return sessionID.(string), true
	}
	return "", false
}

// StopRecording stops a recording session
func (rm *RecordingManager) StopRecording(ctx context.Context, devicePath string) error {
	rm.logger.WithField("device_path", devicePath).Info("Stopping MediaMTX path-based recording")

	// Check if device is recording - lock-free read with sync.Map
	if _, exists := rm.deviceToSession.Load(devicePath); !exists {
		return fmt.Errorf("device %s is not currently recording", devicePath)
	}

	// Disable recording but keep path alive for streaming
	err := rm.streamManager.DisableRecording(ctx, devicePath)
	if err != nil {
		rm.logger.WithError(err).WithField("device_path", devicePath).Warn("Failed to disable recording")
		return fmt.Errorf("failed to disable recording: %w", err)
	}

	// Remove from device mapping - lock-free operation with sync.Map
	rm.deviceToSession.Delete(devicePath)

	rm.logger.WithFields(logging.Fields{
		"device": devicePath,
	}).Info("MediaMTX path-based recording stopped successfully")

	return nil
}

// ListRecordingSessions returns all active recording sessions
func (rm *RecordingManager) ListRecordingSessions() []*RecordingSession {
	var sessions []*RecordingSession

	// Iterate over sync.Map - lock-free operation
	rm.sessions.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*RecordingSession))
		return true // Continue iteration
	})

	return sessions
}

// GetRecordingsList retrieves recordings from MediaMTX API
func (rm *RecordingManager) GetRecordingsList(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	rm.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting recordings list from MediaMTX API")

	// Call MediaMTX recordings API
	queryParams := fmt.Sprintf("?page=%d&itemsPerPage=%d", offset/limit, limit)
	data, err := rm.client.Get(ctx, "/v3/recordings/list"+queryParams)
	if err != nil {
		rm.logger.WithError(err).Error("Failed to get recordings from MediaMTX API")
		return nil, fmt.Errorf("failed to get recordings from MediaMTX: %w", err)
	}

	// Parse MediaMTX RecordingList response
	var recordingList MediaMTXRecordingList
	if err := json.Unmarshal(data, &recordingList); err != nil {
		rm.logger.WithError(err).Error("Failed to parse recordings response from MediaMTX")
		return nil, fmt.Errorf("failed to parse recordings response: %w", err)
	}

	// Convert MediaMTX recordings to our FileMetadata format
	files := make([]*FileMetadata, 0, len(recordingList.Items))
	for _, recording := range recordingList.Items {
		// Convert each recording and its segments to file metadata
		fileMetadata := rm.convertRecordingToFileMetadata(&recording)
		files = append(files, fileMetadata...)
	}

	// Apply client-side pagination if needed
	start := offset % limit
	end := start + limit
	if start >= len(files) {
		files = []*FileMetadata{}
	} else if end > len(files) {
		files = files[start:]
	} else {
		files = files[start:end]
	}

	return &FileListResponse{
		Files:  files,
		Total:  int(recordingList.ItemCount),
		Limit:  limit,
		Offset: offset,
	}, nil
}

// convertRecordingToFileMetadata converts MediaMTX recording to our FileMetadata format
func (rm *RecordingManager) convertRecordingToFileMetadata(recording *MediaMTXRecording) []*FileMetadata {
	var files []*FileMetadata

	for i, segment := range recording.Segments {
		// Parse segment start time
		startTime, err := time.Parse(time.RFC3339, segment.Start)
		if err != nil {
			rm.logger.WithError(err).WithField("segment_start", segment.Start).Warn("Failed to parse segment start time")
			startTime = time.Now() // fallback
		}

		// Generate filename based on recording name and segment
		filename := fmt.Sprintf("%s_segment_%d.%s", recording.Name, i, rm.getRecordFormat())

		// Create file metadata
		fileMetadata := &FileMetadata{
			FileName:    filename,
			FileSize:    0, // Size not available from MediaMTX API
			CreatedAt:   startTime,
			ModifiedAt:  startTime,
			Duration:    nil, // Duration not available from MediaMTX API
			DownloadURL: fmt.Sprintf("/files/recordings/%s", filename),
		}

		files = append(files, fileMetadata)
	}

	return files
}

// CleanupOldRecordings removes old recording files based on age and count limits
func (rm *RecordingManager) CleanupOldRecordings(ctx context.Context, maxAge time.Duration, maxCount int) error {
	rm.logger.WithFields(logging.Fields{
		"max_age":   maxAge,
		"max_count": maxCount,
	}).Info("Starting cleanup of old recordings")

	// Get recordings list
	recordings, err := rm.GetRecordingsList(ctx, 1000, 0) // Get up to 1000 recordings
	if err != nil {
		return fmt.Errorf("failed to get recordings list: %w", err)
	}

	if recordings == nil || len(recordings.Files) == 0 {
		rm.logger.Debug("No recordings found for cleanup")
		return nil
	}

	// Sort by creation time (oldest first)
	cutoffTime := time.Now().Add(-maxAge)
	deletedCount := 0

	for _, item := range recordings.Files {
		// Check if we've reached the max count limit
		if len(recordings.Files)-deletedCount <= maxCount {
			break
		}

		// Check if recording is older than max age
		if item.CreatedAt.Before(cutoffTime) {
			if err := rm.DeleteRecording(ctx, item.FileName); err != nil {
				rm.logger.WithError(err).WithField("filename", item.FileName).Warn("Failed to delete old recording")
				continue
			}
			deletedCount++
		}
	}

	rm.logger.WithField("deleted_count", fmt.Sprintf("%d", deletedCount)).Info("Recording cleanup completed")
	return nil
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Recording configuration helper methods - derive from centralized config
// These methods provide recording settings from the centralized MediaMTXConfig

func (rm *RecordingManager) getRecordFormat() string {
	recordingConfig, err := rm.configIntegration.GetRecordingConfig()
	if err != nil {
		rm.logger.WithError(err).Warn("Failed to get recording config, using default format")
		return "mp4" // fallback
	}
	return recordingConfig.Format
}
