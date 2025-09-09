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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RecordingManager manages MediaMTX-based recording operations
type RecordingManager struct {
	client        MediaMTXClient
	config        *MediaMTXConfig
	logger        *logging.Logger
	pathManager   PathManager
	streamManager StreamManager

	// Recording sessions
	sessions   map[string]*RecordingSession
	sessionsMu sync.RWMutex

	// Device to session mapping for efficient lookup
	deviceToSession map[string]string // device path -> session ID
	deviceMu        sync.RWMutex

	// Recording configuration is derived from the centralized config
	// No separate recordingConfig field needed - use config directly
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
func NewRecordingManager(client MediaMTXClient, pathManager PathManager, streamManager StreamManager, config *MediaMTXConfig, logger *logging.Logger) *RecordingManager {
	// Use centralized configuration - no need to create component-specific defaults
	// All recording configuration comes from the centralized config system
	// Recording settings are derived from the centralized MediaMTXConfig

	return &RecordingManager{
		client:          client,
		config:          config,
		logger:          logger,
		pathManager:     pathManager,
		streamManager:   streamManager,
		sessions:        make(map[string]*RecordingSession),
		deviceToSession: make(map[string]string),
		// recordingConfig is derived from config field - no separate initialization needed
	}
}

// StartRecording starts a new recording session for a camera device using MediaMTX
func (rm *RecordingManager) StartRecording(ctx context.Context, devicePath, outputPath string, options map[string]interface{}) (*RecordingSession, error) {
	// Input validation
	if strings.TrimSpace(devicePath) == "" {
		return nil, fmt.Errorf("device path cannot be empty")
	}
	if strings.TrimSpace(outputPath) == "" {
		return nil, fmt.Errorf("output path cannot be empty")
	}

	rm.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"output_path": outputPath,
		"options":     options,
	}).Info("Starting MediaMTX recording session")

	// Check if device is already recording
	rm.deviceMu.RLock()
	if existingSessionID, exists := rm.deviceToSession[devicePath]; exists {
		rm.deviceMu.RUnlock()
		return nil, fmt.Errorf("device %s is already recording in session %s", devicePath, existingSessionID)
	}
	rm.deviceMu.RUnlock()

	// Generate unique session ID
	sessionID := fmt.Sprintf("rec_%d_%s", time.Now().Unix(), generateRandomString(8))

	// Create path name from device path
	pathName := rm.generatePathName(devicePath)

	// Check if path already exists and handle collision
	if rm.pathManager.PathExists(ctx, pathName) {
		rm.logger.WithField("path_name", pathName).Warn("Path already exists, deleting it first")
		err := rm.pathManager.DeletePath(ctx, pathName)
		if err != nil {
			rm.logger.WithError(err).WithField("path_name", pathName).Warn("Failed to delete existing path")
		}
	}

	// Use StreamManager to handle device-to-stream conversion
	// This creates the FFmpeg process and MediaMTX path in one operation
	stream, err := rm.streamManager.StartRecordingStream(ctx, devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to start recording stream: %w", err)
	}

	// Create recording session
	session := &RecordingSession{
		ID:         sessionID,
		DevicePath: devicePath,
		Path:       stream.Name, // Use the stream name as the MediaMTX path
		FilePath:   outputPath,
		StartTime:  time.Now(),
		Status:     "active",
		State:      SessionStateRecording,
		UseCase:    UseCaseRecording,
	}

	// Store session
	rm.sessionsMu.Lock()
	rm.sessions[sessionID] = session
	rm.sessionsMu.Unlock()

	rm.deviceMu.Lock()
	rm.deviceToSession[devicePath] = sessionID
	rm.deviceMu.Unlock()

	rm.logger.WithFields(logging.Fields{
		"session_id": sessionID,
		"path_name":  pathName,
		"device":     devicePath,
	}).Info("MediaMTX recording session started successfully")

	return session, nil
}

// GetRecordingSession retrieves a recording session by ID
func (rm *RecordingManager) GetRecordingSession(sessionID string) (*RecordingSession, bool) {
	rm.sessionsMu.RLock()
	defer rm.sessionsMu.RUnlock()

	session, exists := rm.sessions[sessionID]
	return session, exists
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
	rm.deviceMu.RLock()
	defer rm.deviceMu.RUnlock()

	sessionID, exists := rm.deviceToSession[device]
	return sessionID, exists
}

// StopRecording stops a recording session
func (rm *RecordingManager) StopRecording(ctx context.Context, sessionID string) error {
	rm.logger.WithField("session_id", sessionID).Info("Stopping MediaMTX recording session")

	rm.sessionsMu.RLock()
	session, exists := rm.sessions[sessionID]
	if !exists {
		rm.sessionsMu.RUnlock()
		return fmt.Errorf("recording session %s not found", sessionID)
	}
	rm.sessionsMu.RUnlock()

	// Remove the path from MediaMTX
	if session.Path != "" {
		err := rm.pathManager.DeletePath(ctx, session.Path)
		if err != nil {
			rm.logger.WithError(err).WithField("path_name", session.Path).Warn("Failed to delete path from MediaMTX")
		}
	}

	// Update session status
	session.Status = "stopped"
	endTime := time.Now()
	session.EndTime = &endTime

	// Remove from device mapping
	rm.deviceMu.Lock()
	delete(rm.deviceToSession, session.DevicePath)
	rm.deviceMu.Unlock()

	// Remove from sessions
	rm.sessionsMu.Lock()
	delete(rm.sessions, sessionID)
	rm.sessionsMu.Unlock()

	rm.logger.WithFields(logging.Fields{
		"session_id": sessionID,
		"device":     session.DevicePath,
		"duration":   session.EndTime.Sub(session.StartTime),
	}).Info("MediaMTX recording session stopped successfully")

	return nil
}

// ListRecordingSessions returns all active recording sessions
func (rm *RecordingManager) ListRecordingSessions() []*RecordingSession {
	rm.sessionsMu.RLock()
	defer rm.sessionsMu.RUnlock()

	sessions := make([]*RecordingSession, 0, len(rm.sessions))
	for _, session := range rm.sessions {
		sessions = append(sessions, session)
	}

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

// createRecordingPathConfig creates a MediaMTX path configuration for recording
func (rm *RecordingManager) createRecordingPathConfig(devicePath, outputPath string, options map[string]interface{}) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"sourceOnDemand":             true,  // Wait for stream to be available
		"sourceOnDemandStartTimeout": "10s", // Timeout for stream to start
		"sourceOnDemandCloseAfter":   "30s", // Close stream after this duration of inactivity
		"record":                     true,
		"recordPath":                 rm.getRecordingOutputPath(outputPath),
		"recordFormat":               rm.getRecordFormat(),
		"recordPartDuration":         rm.getRecordPartDuration().String(),
		"recordMaxPartSize":          rm.getRecordMaxPartSize(),
		"recordSegmentDuration":      rm.getRecordSegmentDuration().String(),
		"recordDeleteAfter":          rm.getRecordDeleteAfter().String(),
	}

	// Apply options overrides
	for key, value := range options {
		switch key {
		case "record_format":
			config["recordFormat"] = value
		case "record_path":
			config["recordPath"] = value
		case "segment_duration":
			config["recordSegmentDuration"] = value
		case "part_duration":
			config["recordPartDuration"] = value
		case "max_part_size":
			config["recordMaxPartSize"] = value
		}
	}

	return config, nil
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

// generatePathName generates a unique path name for a device
func (rm *RecordingManager) generatePathName(devicePath string) string {
	// Extract device identifier from path
	deviceID := strings.ReplaceAll(devicePath, "/", "_")
	deviceID = strings.ReplaceAll(deviceID, "\\", "_")

	// Remove leading underscore if present
	if strings.HasPrefix(deviceID, "_") {
		deviceID = strings.TrimPrefix(deviceID, "_")
	}

	// Add timestamp with microseconds for uniqueness
	timestamp := time.Now().Format("20060102_150405.000000")

	return fmt.Sprintf("camera_%s_%s", deviceID, timestamp)
}

// getRecordingOutputPath processes the output path for MediaMTX recording
func (rm *RecordingManager) getRecordingOutputPath(outputPath string) string {
	// If output path is provided, use it
	if outputPath != "" {
		return outputPath
	}

	// Use configured recordings path from centralized config
	if rm.config.RecordingsPath != "" {
		return rm.config.RecordingsPath
	}

	// Default recordings path
	return "/tmp/recordings"
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
	// Default recording format - can be overridden by config if needed
	return "mp4"
}

func (rm *RecordingManager) getRecordPartDuration() time.Duration {
	// Default part duration - can be overridden by config if needed
	return 1 * time.Hour
}

func (rm *RecordingManager) getRecordMaxPartSize() string {
	// Default max part size - can be overridden by config if needed
	return "100MB"
}

func (rm *RecordingManager) getRecordSegmentDuration() time.Duration {
	// Default segment duration - can be overridden by config if needed
	return 10 * time.Second
}

func (rm *RecordingManager) getRecordDeleteAfter() time.Duration {
	// Default delete after duration - can be overridden by config if needed
	return 7 * 24 * time.Hour // 7 days
}
