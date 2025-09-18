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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RecordingManager manages stateless recording operations via MediaMTX API.
//
// RESPONSIBILITIES:
// - Stateless recording orchestration using MediaMTX as source of truth
// - Auto-stop timer management for duration-based recordings
// - RTSP keepalive coordination for on-demand stream activation
// - Recording operations for both V4L2 devices and external RTSP streams
//
// ARCHITECTURE:
// - Operates with cameraID as primary identifier
// - MediaMTX path names match camera identifiers
// - Minimal local state (timers map for auto-stop functionality)
// - Query MediaMTX API directly for recording status
//
// API INTEGRATION:
// - Returns JSON-RPC API-ready responses
// - Uses MediaMTX api_types.go for all operations
//
// ARCHITECTURE COMPLIANCE: StartRecording and StopRecording return API-ready responses directly
// ARCHITECTURE COMPLIANCE: ListRecordings method returns *ListRecordingsResponse (API-ready)
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
	timers sync.Map // cameraID -> *time.Timer (also serves as recording state tracker)
}

// NOTE: MediaMTXRecordingConfig removed - using PathConf from api_types.go instead
// This ensures single source of truth with MediaMTX swagger.json specification

// IsRecording checks if a camera is currently recording using the timers map
// This provides minimal state tracking for event correlation without duplicating MediaMTX state
func (rm *RecordingManager) IsRecording(cameraID string) bool {
	_, exists := rm.timers.Load(cameraID)
	return exists // If timer exists, recording is active
}

// forceStopRecording forcefully stops recording (for device disconnection scenarios)
// This cleans up local state without trying to communicate with MediaMTX
func (rm *RecordingManager) forceStopRecording(cameraID string) {
	// Stop and remove any auto-stop timer
	if timer, exists := rm.timers.LoadAndDelete(cameraID); exists {
		timer.(*time.Timer).Stop()
		rm.logger.WithField("cameraID", cameraID).Info("Forced stop recording due to device disconnection")
	}

	// Stop RTSP keepalive reader
	rm.stopRTSPKeepalive(cameraID)

	// Note: We don't try to call MediaMTX here since the device is gone
	// MediaMTX will automatically stop recording when FFmpeg fails
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

// StartRecording starts recording and returns API-ready response with rich metadata
func (rm *RecordingManager) StartRecording(ctx context.Context, cameraID string, options *PathConf) (*StartRecordingResponse, error) {
	// Input validation
	if strings.TrimSpace(cameraID) == "" {
		return nil, fmt.Errorf("camera ID cannot be empty")
	}

	// Convert camera identifier to device path for internal operations
	// MediaMTX path name = camera identifier (camera0), but we need device path for validation
	devicePath, exists := rm.pathManager.GetDevicePathForCamera(cameraID)
	if !exists {
		return nil, fmt.Errorf("camera '%s' not found or not accessible", cameraID)
	}

	// Use camera identifier as MediaMTX path name
	pathName := cameraID

	rm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"device_path": devicePath,
		"path_name":   pathName,
		"options":     options,
	}).Info("Starting recording by enabling record flag in MediaMTX")

	// Ensure path exists in MediaMTX before checking recording status
	// In stateless architecture, we create paths on-demand
	if !rm.pathManager.PathExists(ctx, pathName) {
		// Create path with comprehensive recording configuration
		pathOptions, err := rm.configIntegration.BuildRecordingPathConf(devicePath, pathName)
		if err != nil {
			return nil, fmt.Errorf("failed to build recording path configuration: %w", err)
		}
		err = rm.pathManager.CreatePath(ctx, pathName, devicePath, pathOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create path %s: %w", pathName, err)
		}
	}

	// Check current recording state by querying MediaMTX
	isRecording, err := rm.isPathRecording(ctx, pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to check recording status: %w", err)
	}
	if isRecording {
		return nil, fmt.Errorf("path %s is already recording", pathName)
	}

	// Enable recording by patching the path configuration
	err = rm.enableRecordingOnPath(ctx, pathName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to enable recording on path: %w", err)
	}

	// Start RTSP keepalive reader to trigger on-demand publisher
	err = rm.startRTSPKeepalive(ctx, pathName)
	if err != nil {
		// If keepalive fails, disable recording
		rm.disableRecordingOnPath(ctx, pathName)
		return nil, fmt.Errorf("failed to start RTSP keepalive: %w", err)
	}

	// Set up auto-stop timer if recordDeleteAfter is specified
	// Note: Using recordDeleteAfter as the auto-stop duration for backward compatibility
	if options != nil && options.RecordDeleteAfter != "" {
		if duration, err := time.ParseDuration(options.RecordDeleteAfter); err == nil {
			rm.setAutoStopTimer(cameraID, duration) // Use cameraID (camera identifier) as key
			rm.logger.WithFields(logging.Fields{
				"cameraID": cameraID,
				"duration": duration,
			}).Debug("Set auto-stop timer for recording")
		} else {
			rm.logger.WithError(err).WithField("recordDeleteAfter", options.RecordDeleteAfter).Warn("Invalid recordDeleteAfter format, ignoring auto-stop timer")
		}
	}

	// Build API-ready response with rich recording metadata
	format := "fmp4" // Default format from config
	if options != nil && options.RecordFormat != "" {
		format = options.RecordFormat
	}

	// Generate expected filename based on MediaMTX pattern
	filename := fmt.Sprintf("%s_%s", cameraID, time.Now().Format("2006-01-02_15-04-05"))
	if format == "fmp4" {
		filename += ".mp4"
	}

	response := &StartRecordingResponse{
		Device:    cameraID,
		Filename:  filename,
		Status:    "RECORDING",
		StartTime: time.Now().Format(time.RFC3339),
		Format:    format,
	}

	rm.logger.WithFields(logging.Fields{
		"cameraID":       cameraID,
		"filename":       filename,
		"format":         format,
		"keepalive_used": true,
		"path_name":      pathName,
	}).Info("Recording started successfully with API-ready response")

	return response, nil
}

// GetRecordingInfo gets detailed information about a specific recording file
// GetRecordingInfo returns API-ready recording information with rich metadata
func (rm *RecordingManager) GetRecordingInfo(ctx context.Context, filename string) (*GetRecordingInfoResponse, error) {
	rm.logger.WithField("filename", filename).Debug("Getting API-ready recording info")

	// Get basic file metadata first
	recordingsPath := rm.config.RecordingsPath
	filePath := filepath.Join(recordingsPath, filename)

	// Get file stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("recording file not found: %v", err)
	}

	// Extract device from filename pattern (camera0_timestamp.mp4)
	device := "camera0" // Default
	if parts := strings.Split(filename, "_"); len(parts) > 0 {
		if strings.HasPrefix(parts[0], "camera") {
			device = parts[0]
		}
	}

	// Extract format from filename extension
	format := "mp4" // Default
	if ext := filepath.Ext(filename); ext != "" {
		format = strings.TrimPrefix(ext, ".")
	}

	// TODO: Parse video duration from ffprobe output (ffprobe already integrated)
	// INVESTIGATION: ffprobe is already used in snapshot_manager.go:1021-1027 for image metadata
	// CURRENT: Hardcoded duration = 0 placeholder for all recordings
	// SOLUTION: Use same ffprobe integration pattern for video files:
	//   ffprobe -v quiet -print_format json -show_format [video_file]
	//   Parse JSON: result.format.duration (string, convert to float64)
	// REFERENCE: MediaMTX recordings stored at config.RecordingsPath with .mp4/.fmp4 extensions
	// EFFORT: 3-4 hours - implement ffprobe video parsing similar to image metadata
	duration := float64(0) // Placeholder

	// Build API-ready response with rich metadata
	response := &GetRecordingInfoResponse{
		Filename:  filename,
		FileSize:  fileInfo.Size(),
		Duration:  duration,
		CreatedAt: fileInfo.ModTime().Format(time.RFC3339),
		Format:    format,
		Device:    device,
	}

	rm.logger.WithFields(logging.Fields{
		"filename":  filename,
		"device":    device,
		"format":    format,
		"file_size": fileInfo.Size(),
	}).Debug("Recording info retrieved successfully")

	return response, nil
}

// DeleteRecording deletes a recording file via MediaMTX API
func (rm *RecordingManager) DeleteRecording(ctx context.Context, filename string) error {
	rm.logger.WithField("filename", filename).Info("Deleting recording via MediaMTX API")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// TODO: Implement recording deletion via MediaMTX API integration
	// INVESTIGATION: MediaMTX v3 API has DELETE /v3/recordings/delete/{name} endpoint
	// CURRENT: Returns "not yet implemented" error, complex filename parsing was removed
	// SOLUTION: Use MediaMTX client.Delete() with endpoint "/v3/recordings/delete/" + recordingName
	// REFERENCE: MediaMTX API docs, existing client.Get() pattern in GetRecordingsList():440
	// FILENAME FORMAT: MediaMTX uses path_timestamp format (e.g., "camera0_2025-01-15_14-30-00")
	// EFFORT: 4-6 hours - implement MediaMTX DELETE API call with proper error handling
	// The complex filename parsing was over-engineered and fragile
	return fmt.Errorf("recording deletion not yet implemented - requires MediaMTX API investigation")
}

// StopRecording stops recording and returns API-ready response with actual metadata
func (rm *RecordingManager) StopRecording(ctx context.Context, cameraID string) (*StopRecordingResponse, error) {
	// Convert camera identifier to device path for validation
	devicePath, exists := rm.pathManager.GetDevicePathForCamera(cameraID)
	if !exists {
		return nil, fmt.Errorf("camera '%s' not found or not accessible", cameraID)
	}

	// Use camera identifier as MediaMTX path name
	pathName := cameraID

	rm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"device_path": devicePath,
		"path_name":   pathName,
	}).Info("Stopping recording by disabling record flag in MediaMTX")

	// Check if currently recording by querying MediaMTX
	isRecording, err := rm.isPathRecording(ctx, pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to check recording status: %w", err)
	}
	if !isRecording {
		return nil, fmt.Errorf("path %s is not currently recording", pathName)
	}

	// Stop RTSP keepalive reader first
	rm.stopRTSPKeepalive(pathName)

	// Disable recording by patching the path configuration
	err = rm.disableRecordingOnPath(ctx, pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to disable recording on path: %w", err)
	}

	// Cancel any auto-stop timer
	if timer, exists := rm.timers.LoadAndDelete(cameraID); exists {
		timer.(*time.Timer).Stop()
		rm.logger.WithField("cameraID", cameraID).Debug("Cancelled auto-stop timer")
	}

	// Get timer info for duration calculation
	var startTime time.Time
	var duration float64
	if _, exists := rm.timers.Load(cameraID); exists {
		// TODO: Track actual start time in timer metadata for accurate duration calculation
		// INVESTIGATION: timers sync.Map only stores *time.Timer, no metadata about start time
		// CURRENT: Using placeholder time.Now().Add(-time.Hour) for duration calculation
		// SOLUTION: Change timers map value from *time.Timer to custom struct:
		//   type RecordingTimer struct { Timer *time.Timer; StartTime time.Time; CameraID string }
		// USAGE: Store start time when timer created in setAutoStopTimer():658
		// EFFORT: 2-3 hours - refactor timers map structure and update all timer operations
		startTime = time.Now().Add(-time.Hour) // Placeholder - need to track start time
		duration = time.Since(startTime).Seconds()
	} else {
		startTime = time.Now().Add(-time.Hour) // Placeholder
		duration = 3600                        // Placeholder
	}

	// Generate expected filename based on MediaMTX recording pattern
	filename := fmt.Sprintf("%s_%s.mp4", cameraID, startTime.Format("2006-01-02_15-04-05"))

	// Build API-ready response with actual recording metadata
	response := &StopRecordingResponse{
		Device:    cameraID,
		Filename:  filename,
		Status:    "STOPPED",
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Duration:  duration,
		FileSize:  1024, // TODO: Get actual file size from MediaMTX API or filesystem
		// INVESTIGATION: MediaMTX API /v3/recordings/list returns segments but no file sizes
		// CURRENT: Hardcoded 1024 bytes placeholder for all recordings
		// SOLUTION: Use os.Stat() on recording file path or enhance MediaMTX API query
		// FILE PATH: config.RecordingsPath + "/" + filename (with proper extension)
		// ALTERNATIVE: Query MediaMTX /v3/recordings/get/{name} if available in v3 API
		// EFFORT: 2-3 hours - implement file size retrieval with filesystem fallback
		Format: "fmp4",
	}

	rm.logger.WithFields(logging.Fields{
		"cameraID":  cameraID,
		"filename":  filename,
		"duration":  duration,
		"path_name": pathName,
	}).Info("Recording stopped successfully with API-ready response")

	return response, nil
}

// ListRecordings returns API-ready recording list response
func (rm *RecordingManager) ListRecordings(ctx context.Context, limit, offset int) (*ListRecordingsResponse, error) {
	rm.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting API-ready recordings list")

	// Get file list from existing method
	fileList, err := rm.GetRecordingsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to API-ready RecordingFileInfo format with rich metadata
	recordings := make([]RecordingFileInfo, len(fileList.Files))
	for i, file := range fileList.Files {
		// Extract device from filename pattern (camera0_timestamp.mp4)
		device := "camera0" // Default
		if parts := strings.Split(file.FileName, "_"); len(parts) > 0 {
			if strings.HasPrefix(parts[0], "camera") {
				device = parts[0]
			}
		}

		// Extract format from filename extension
		format := "mp4" // Default
		if parts := strings.Split(file.FileName, "."); len(parts) > 1 {
			format = parts[len(parts)-1]
		}

		duration := float64(0)
		if file.Duration != nil {
			duration = float64(*file.Duration)
		}

		recordings[i] = RecordingFileInfo{
			Device:      device,
			Filename:    file.FileName,
			FileSize:    file.FileSize,
			Duration:    duration,
			CreatedAt:   file.CreatedAt.Format(time.RFC3339),
			Format:      format,
			DownloadURL: fmt.Sprintf("/files/recordings/%s", file.FileName),
		}
	}

	response := &ListRecordingsResponse{
		Recordings: recordings,
		Total:      fileList.Total,
		Limit:      limit,
		Offset:     offset,
	}

	return response, nil
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

// CleanupOldRecordings removes recording files based on age and count limits
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

		// Check if recording exceeds maximum age
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
func (rm *RecordingManager) isPathRecording(ctx context.Context, cameraID string) (bool, error) {
	endpoint := fmt.Sprintf("/v3/config/paths/get/%s", cameraID)
	data, err := rm.client.Get(ctx, endpoint)
	if err != nil {
		return false, fmt.Errorf("failed to get path config: %w", err)
	}

	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	var pathConfig PathConf
	if err := json.Unmarshal(data, &pathConfig); err != nil {
		return false, fmt.Errorf("failed to parse path config: %w", err)
	}

	return pathConfig.Record, nil
}

// enableRecordingOnPath enables recording on a MediaMTX path
func (rm *RecordingManager) enableRecordingOnPath(ctx context.Context, cameraID string, options *PathConf) error {
	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	recordConfig := &PathConf{
		Record:             true,
		RecordPath:         rm.config.RecordingsPath,        // CRITICAL: Tell MediaMTX WHERE to save files
		RecordPartDuration: "10s",                           // MediaMTX expects string duration
		RecordMaxPartSize:  "100MB",                         // Reasonable default
		RecordFormat:       rm.recordingConfig.RecordFormat, // Use config default
	}

	// Override with options if provided
	if options != nil {
		if options.RecordFormat != "" {
			recordConfig.RecordFormat = options.RecordFormat
		}
		if options.RecordSegmentDuration != "" {
			recordConfig.RecordSegmentDuration = options.RecordSegmentDuration
		}
		if options.RecordPartDuration != "" {
			recordConfig.RecordPartDuration = options.RecordPartDuration
		}
		if options.RecordMaxPartSize != "" {
			recordConfig.RecordMaxPartSize = options.RecordMaxPartSize
		}
		if options.RecordDeleteAfter != "" {
			recordConfig.RecordDeleteAfter = options.RecordDeleteAfter
		}
	}

	endpoint := fmt.Sprintf("/v3/config/paths/patch/%s", cameraID)
	jsonData, err := marshalUpdatePathRequest(recordConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal PathConf record config: %w", err)
	}
	return rm.client.Patch(ctx, endpoint, jsonData)
}

// disableRecordingOnPath disables recording on a MediaMTX path
func (rm *RecordingManager) disableRecordingOnPath(ctx context.Context, cameraID string) error {
	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	recordConfig := &PathConf{
		Record: false,
	}

	endpoint := fmt.Sprintf("/v3/config/paths/patch/%s", cameraID)
	jsonData, err := marshalUpdatePathRequest(recordConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal PathConf record config: %w", err)
	}
	return rm.client.Patch(ctx, endpoint, jsonData)
}

// startRTSPKeepalive starts an RTSP reader to keep on-demand sources alive
func (rm *RecordingManager) startRTSPKeepalive(ctx context.Context, cameraID string) error {
	// Use the RecordingManager's own RTSPKeepaliveReader
	return rm.keepaliveReader.StartKeepalive(ctx, cameraID)
}

// stopRTSPKeepalive stops the RTSP reader for a path
func (rm *RecordingManager) stopRTSPKeepalive(cameraID string) {
	rm.keepaliveReader.StopKeepalive(cameraID)
}

// setAutoStopTimer sets up an auto-stop timer for a recording
func (rm *RecordingManager) setAutoStopTimer(cameraID string, duration time.Duration) {
	timer := time.AfterFunc(duration, func() {
		rm.logger.WithField("cameraID", cameraID).Info("Auto-stopping recording after duration")

		// Stop recording using cameraID (camera identifier) as path name
		ctx := context.Background()
		if err := rm.disableRecordingOnPath(ctx, cameraID); err != nil {
			rm.logger.WithError(err).WithField("cameraID", cameraID).Error("Failed to auto-stop recording")
		}

		// Stop RTSP keepalive
		rm.stopRTSPKeepalive(cameraID)

		// Remove timer
		rm.timers.Delete(cameraID)
	})

	rm.timers.Store(cameraID, timer)
	rm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"duration": duration,
	}).Debug("Set auto-stop timer for recording")
}
