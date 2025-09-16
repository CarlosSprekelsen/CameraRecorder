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
	keepaliveReader   *RTSPKeepaliveReader

	// Optional timers for auto-stop functionality
	timers sync.Map // device -> *time.Timer
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

	return &RecordingManager{
		client:            client,
		config:            config,
		recordingConfig:   recordingConfig,
		configIntegration: configIntegration,
		logger:            logger,
		pathManager:       pathManager,
		streamManager:     streamManager,
		keepaliveReader:   NewRTSPKeepaliveReader(config, logger),
		// No local state needed - query MediaMTX directly for recording status
	}
}

// StartRecording starts recording by enabling the record flag in MediaMTX
func (rm *RecordingManager) StartRecording(ctx context.Context, device string, options map[string]interface{}) error {
	// Input validation
	if strings.TrimSpace(device) == "" {
		return fmt.Errorf("device cannot be empty")
	}

	// Convert camera identifier to device path for internal operations
	// MediaMTX path name = camera identifier (camera0), but we need device path for validation
	devicePath, exists := rm.pathManager.GetDevicePathForCamera(device)
	if !exists {
		return fmt.Errorf("camera '%s' not found or not accessible", device)
	}

	// Use camera identifier as MediaMTX path name
	pathName := device

	rm.logger.WithFields(logging.Fields{
		"device":      device,
		"device_path": devicePath,
		"path_name":   pathName,
		"options":     options,
	}).Info("Starting recording by enabling record flag in MediaMTX")

	// Check if already recording by querying MediaMTX
	isRecording, err := rm.isPathRecording(ctx, pathName)
	if err != nil {
		return fmt.Errorf("failed to check recording status: %w", err)
	}
	if isRecording {
		return fmt.Errorf("path %s is already recording", pathName)
	}

	// Enable recording by patching the path configuration
	err = rm.enableRecordingOnPath(ctx, pathName, options)
	if err != nil {
		return fmt.Errorf("failed to enable recording on path: %w", err)
	}

	// Start RTSP keepalive reader to trigger on-demand publisher
	err = rm.startRTSPKeepalive(ctx, pathName)
	if err != nil {
		// If keepalive fails, disable recording
		rm.disableRecordingOnPath(ctx, pathName)
		return fmt.Errorf("failed to start RTSP keepalive: %w", err)
	}

	// Set up auto-stop timer if duration_s is specified
	if durationStr, ok := options["duration_s"].(string); ok && durationStr != "" {
		if duration, err := time.ParseDuration(durationStr + "s"); err == nil {
			rm.setAutoStopTimer(device, duration) // Use device (camera identifier) as key
			rm.logger.WithFields(logging.Fields{
				"device":   device,
				"duration": duration,
			}).Debug("Set auto-stop timer for recording")
		} else {
			rm.logger.WithError(err).WithField("duration_s", durationStr).Warn("Invalid duration_s format, ignoring auto-stop timer")
		}
	}

	rm.logger.WithFields(logging.Fields{
		"path_name": pathName,
		"device":    device,
	}).Info("Recording started successfully with RTSP keepalive")

	return nil
}

// GetRecordingInfo gets detailed information about a specific recording file
func (rm *RecordingManager) GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	rm.logger.WithField("filename", filename).Info("Getting recording info")

	// TODO:For MediaMTX-based recording, we would query the MediaMTX API
	// For now, return basic file metadata
	return &FileMetadata{
		FileName:   filename,
		FileSize:   0, // Would be populated from MediaMTX API
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}, nil
}

// DeleteRecording deletes a recording file via MediaMTX API
func (rm *RecordingManager) DeleteRecording(ctx context.Context, filename string) error {
	rm.logger.WithField("filename", filename).Info("Deleting recording via MediaMTX API")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// For now, return an error indicating this feature needs MediaMTX API investigation
	// The complex filename parsing was over-engineered and fragile
	return fmt.Errorf("recording deletion not yet implemented - requires MediaMTX API investigation")
}

// StopRecording stops recording by disabling the record flag in MediaMTX
func (rm *RecordingManager) StopRecording(ctx context.Context, device string) error {
	// Convert camera identifier to device path for validation
	devicePath, exists := rm.pathManager.GetDevicePathForCamera(device)
	if !exists {
		return fmt.Errorf("camera '%s' not found or not accessible", device)
	}

	// Use camera identifier as MediaMTX path name
	pathName := device

	rm.logger.WithFields(logging.Fields{
		"device":      device,
		"device_path": devicePath,
		"path_name":   pathName,
	}).Info("Stopping recording by disabling record flag in MediaMTX")

	// Check if currently recording by querying MediaMTX
	isRecording, err := rm.isPathRecording(ctx, pathName)
	if err != nil {
		return fmt.Errorf("failed to check recording status: %w", err)
	}
	if !isRecording {
		return fmt.Errorf("path %s is not currently recording", pathName)
	}

	// Stop RTSP keepalive reader first
	rm.stopRTSPKeepalive(pathName)

	// Disable recording by patching the path configuration
	err = rm.disableRecordingOnPath(ctx, pathName)
	if err != nil {
		return fmt.Errorf("failed to disable recording on path: %w", err)
	}

	// Cancel any auto-stop timer
	if timer, exists := rm.timers.LoadAndDelete(device); exists {
		timer.(*time.Timer).Stop()
		rm.logger.WithField("device", device).Debug("Cancelled auto-stop timer")
	}

	rm.logger.WithFields(logging.Fields{
		"path_name": pathName,
		"device":    devicePath,
	}).Info("Recording stopped successfully")

	return nil
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

	for _, segment := range recording.Segments {
		// Parse segment start time
		startTime, err := time.Parse(time.RFC3339, segment.Start)
		if err != nil {
			rm.logger.WithError(err).WithField("segment_start", segment.Start).Warn("Failed to parse segment start time")
			startTime = time.Now() // fallback
		}

		// Generate filename based on recording name and segment start time
		filename := fmt.Sprintf("%s_%s", recording.Name, segment.Start)

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
		return "fmp4" // fallback to fmp4 as per STANAG decision
	}
	return recordingConfig.Format
}

// isPathRecording checks if a path is currently recording by querying MediaMTX
func (rm *RecordingManager) isPathRecording(ctx context.Context, pathName string) (bool, error) {
	endpoint := fmt.Sprintf("/v3/config/paths/get/%s", pathName)
	data, err := rm.client.Get(ctx, endpoint)
	if err != nil {
		return false, fmt.Errorf("failed to get path config: %w", err)
	}

	var pathConfig map[string]interface{}
	if err := json.Unmarshal(data, &pathConfig); err != nil {
		return false, fmt.Errorf("failed to parse path config: %w", err)
	}

	record, ok := pathConfig["record"].(bool)
	return ok && record, nil
}

// enableRecordingOnPath enables recording on a MediaMTX path
func (rm *RecordingManager) enableRecordingOnPath(ctx context.Context, pathName string, options map[string]interface{}) error {
	recordConfig := map[string]interface{}{
		"record": true,
	}

	// Add optional recording settings from options
	if format, ok := options["format"].(string); ok && format != "" {
		recordConfig["recordFormat"] = format
	} else {
		recordConfig["recordFormat"] = "fmp4" // Default to fmp4
	}

	if duration, ok := options["duration"].(string); ok && duration != "" {
		recordConfig["recordSegmentDuration"] = duration
	}

	endpoint := fmt.Sprintf("/v3/config/paths/patch/%s", pathName)
	jsonData, err := json.Marshal(recordConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal record config: %w", err)
	}
	return rm.client.Patch(ctx, endpoint, jsonData)
}

// disableRecordingOnPath disables recording on a MediaMTX path
func (rm *RecordingManager) disableRecordingOnPath(ctx context.Context, pathName string) error {
	recordConfig := map[string]interface{}{
		"record": false,
	}

	endpoint := fmt.Sprintf("/v3/config/paths/patch/%s", pathName)
	jsonData, err := json.Marshal(recordConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal record config: %w", err)
	}
	return rm.client.Patch(ctx, endpoint, jsonData)
}

// startRTSPKeepalive starts an RTSP reader to keep on-demand sources alive
func (rm *RecordingManager) startRTSPKeepalive(ctx context.Context, pathName string) error {
	// Use the RecordingManager's own RTSPKeepaliveReader
	return rm.keepaliveReader.StartKeepalive(ctx, pathName)
}

// stopRTSPKeepalive stops the RTSP reader for a path
func (rm *RecordingManager) stopRTSPKeepalive(pathName string) {
	rm.keepaliveReader.StopKeepalive(pathName)
}

// setAutoStopTimer sets up an auto-stop timer for a recording
func (rm *RecordingManager) setAutoStopTimer(device string, duration time.Duration) {
	timer := time.AfterFunc(duration, func() {
		rm.logger.WithField("device", device).Info("Auto-stopping recording after duration")

		// Stop recording using device (camera identifier) as path name
		ctx := context.Background()
		if err := rm.disableRecordingOnPath(ctx, device); err != nil {
			rm.logger.WithError(err).WithField("device", device).Error("Failed to auto-stop recording")
		}

		// Stop RTSP keepalive
		rm.stopRTSPKeepalive(device)

		// Remove timer
		rm.timers.Delete(device)
	})

	rm.timers.Store(device, timer)
	rm.logger.WithFields(logging.Fields{
		"device":   device,
		"duration": duration,
	}).Debug("Set auto-stop timer for recording")
}
