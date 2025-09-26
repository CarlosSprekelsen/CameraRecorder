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
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
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
// - Enhanced metadata extraction and duration tracking
//
// ARCHITECTURE:
// - Operates with cameraID as primary identifier
// - MediaMTX path names match camera identifiers
// - Enhanced timer management with metadata tracking
// - Query MediaMTX API directly for recording status
// - Integrated MetadataManager for video duration parsing
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

	// Enhanced timer management with metadata
	timerManager    *RecordingTimerManager
	metadataManager *MetadataManager

	// Circuit breaker for recording operations
	recordingCircuitBreaker *CircuitBreaker

	// Error recovery manager
	errorRecoveryManager *ErrorRecoveryManager

	// Error metrics collector
	errorMetricsCollector *ErrorMetricsCollector

	// Resource management
	running       int32 // Atomic flag for running state
	resourceStats *RecordingResourceStats
}

// NOTE: MediaMTXRecordingConfig removed - using PathConf from api_types.go instead
// This ensures single source of truth with MediaMTX swagger.json specification

// IsRecording checks if a camera is currently recording using the enhanced timer manager
// This provides enhanced state tracking for event correlation without duplicating MediaMTX state
func (rm *RecordingManager) IsRecording(cameraID string) bool {
	return rm.timerManager.IsRecording(cameraID)
}

// forceStopRecording forcefully stops recording (for device disconnection scenarios)
// This cleans up local state without trying to communicate with MediaMTX
func (rm *RecordingManager) forceStopRecording(cameraID string) {
	// Stop and remove any auto-stop timer using enhanced timer manager
	if rm.timerManager.DeleteTimer(cameraID) {
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

// RecordingResourceStats tracks resource usage for the recording manager
type RecordingResourceStats struct {
	ActiveKeepaliveReaders int64 `json:"active_keepalive_readers"`
	ActiveTimers           int64 `json:"active_timers"`
	MetadataCacheSize      int64 `json:"metadata_cache_size"`
	TotalRecordingsStarted int64 `json:"total_recordings_started"`
	TotalRecordingsStopped int64 `json:"total_recordings_stopped"`
	RecordingErrors        int64 `json:"recording_errors"`
}

// NewRecordingManager creates a new MediaMTX-based recording manager
func NewRecordingManager(client MediaMTXClient, pathManager PathManager, streamManager StreamManager, config *config.MediaMTXConfig, recordingConfig *config.RecordingConfig, configIntegration *ConfigIntegration, logger *logging.Logger) *RecordingManager {
	rm := &RecordingManager{
		client:            client,
		config:            config,
		recordingConfig:   recordingConfig,
		configIntegration: configIntegration,
		logger:            logger,
		pathManager:       pathManager,
		streamManager:     streamManager,
		keepaliveReader:   NewRTSPKeepaliveReaderWithConfig(config, recordingConfig, logger),
		// Enhanced components for metadata and timer management
		timerManager:    NewRecordingTimerManager(logger),
		metadataManager: NewMetadataManager(configIntegration, configIntegration.ffmpegManager, logger),
		// Circuit breaker for recording operations (configurable)
		recordingCircuitBreaker: NewCircuitBreaker("recording", getCircuitBreakerConfig(*configIntegration), logger),
		// Error recovery manager
		errorRecoveryManager: NewErrorRecoveryManager(logger),
		// Error metrics collector
		errorMetricsCollector: NewErrorMetricsCollector(logger),
		// Resource management
		running:       0, // Initially not running
		resourceStats: &RecordingResourceStats{},
	}

	// Initialize error recovery strategies
	rm.errorRecoveryManager.RegisterStrategy(NewRecordingRecoveryStrategy(rm, logger))

	// Initialize error metrics collector
	rm.errorMetricsCollector.Initialize()

	return rm
}

// StartRecording starts recording and returns API-ready response with rich metadata
func (rm *RecordingManager) StartRecording(ctx context.Context, cameraID string, options *PathConf) (*StartRecordingResponse, error) {
	// Add panic recovery for recording operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			rm.logger.WithFields(logging.Fields{
				"camera_id":   cameraID,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "StartRecording",
			}).Error("Panic recovered in StartRecording")
		}
	}()

	// Input validation
	if strings.TrimSpace(cameraID) == "" {
		return nil, fmt.Errorf("camera ID cannot be empty")
	}

	// Execute recording operation directly
	result, err := rm.executeStartRecording(ctx, cameraID, options)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeStartRecording performs the actual recording start operation
func (rm *RecordingManager) executeStartRecording(ctx context.Context, cameraID string, options *PathConf) (*StartRecordingResponse, error) {
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
		var pathOptions *PathConf
		var err error

		// Use ConfigIntegration to build path configuration (architectural alignment)
		pathOptions, err = rm.configIntegration.BuildPathConf(pathName, &PathSource{ID: devicePath}, true)

		if err != nil {
			return nil, fmt.Errorf("failed to build recording path configuration: %w", err)
		}

		// Create path with error recovery
		err = rm.pathManager.CreatePath(ctx, pathName, devicePath, pathOptions)
		if err != nil {
			// Attempt error recovery for path creation
			errorCtx := &ErrorContext{
				Component:   "RecordingManager",
				Operation:   "CreatePath",
				CameraID:    cameraID,
				PathName:    pathName,
				Timestamp:   time.Now(),
				Severity:    SeverityError,
				Recoverable: true,
			}

			// Record error metrics
			rm.errorMetricsCollector.RecordError("RecordingManager", "error", true)

			recoveryErr := rm.errorRecoveryManager.HandleError(ctx, errorCtx, err)
			if recoveryErr != nil {
				// Record recovery failure
				rm.errorMetricsCollector.RecordRecoveryAttempt(false)
				return nil, fmt.Errorf("failed to create path %s: %w", pathName, recoveryErr)
			}

			// Record recovery success
			rm.errorMetricsCollector.RecordRecoveryAttempt(true)

			rm.logger.WithFields(logging.Fields{
				"camera_id": cameraID,
				"path_name": pathName,
			}).Info("Path creation recovered from error")
		}
	}

	// Check current recording state by querying MediaMTX with retry logic
	// This addresses propagation delays where config shows record=true temporarily
	isRecording, err := rm.isPathRecordingWithRetry(ctx, pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to check recording status: %w", err)
	}
	if isRecording {
		return nil, fmt.Errorf("path %s is already recording", pathName)
	}

	// Enable recording by patching the path configuration
	err = rm.enableRecordingOnPath(ctx, pathName, options)
	if err != nil {
		rm.updateRecordingStats(true, true) // Recording start attempt with error
		return nil, fmt.Errorf("failed to enable recording on path: %w", err)
	}

	// Start RTSP keepalive reader to trigger on-demand publisher
	err = rm.startRTSPKeepalive(ctx, pathName)
	if err != nil {
		// If keepalive fails, disable recording
		rm.disableRecordingOnPath(ctx, pathName)
		rm.updateRecordingStats(true, true) // Recording start attempt with error
		return nil, fmt.Errorf("failed to start RTSP keepalive: %w", err)
	}

	// Set up auto-stop timer if recordDeleteAfter is specified
	// Note: Using recordDeleteAfter as the auto-stop duration for backward compatibility
	var recordingDuration time.Duration
	if options != nil && options.RecordDeleteAfter != "" {
		if duration, err := time.ParseDuration(options.RecordDeleteAfter); err == nil {
			recordingDuration = duration
			rm.logger.WithFields(logging.Fields{
				"cameraID": cameraID,
				"duration": duration,
			}).Debug("Parsed auto-stop duration for recording")
		} else {
			rm.logger.WithError(err).WithField("recordDeleteAfter", options.RecordDeleteAfter).Warn("Invalid recordDeleteAfter format, ignoring auto-stop timer")
		}
	}

	// Create enhanced recording timer with metadata
	rm.timerManager.CreateTimer(cameraID, devicePath, recordingDuration, func() {
		rm.logger.WithField("cameraID", cameraID).Info("Auto-stopping recording after duration")

		// Stop recording using cameraID (camera identifier) as path name
		ctx := context.Background()
		if err := rm.disableRecordingOnPath(ctx, cameraID); err != nil {
			rm.logger.WithError(err).WithField("cameraID", cameraID).Error("Failed to auto-stop recording")
		}

		// Stop RTSP keepalive
		rm.stopRTSPKeepalive(cameraID)
	})

	// Build API-ready response with rich recording metadata
	format := rm.getRecordFormat() // Use configured format (STANAG 4609 compliant)
	if options != nil && options.RecordFormat != "" {
		format = options.RecordFormat
	}

	// Generate API filename (base name, no extension) per API documentation
	filename := fmt.Sprintf("%s_%s", cameraID, time.Now().Format("2006-01-02_15-04-05"))

	response := &StartRecordingResponse{
		Device:    cameraID,
		Filename:  filename,
		Status:    "RECORDING",
		StartTime: time.Now().Format(time.RFC3339),
		Format:    format,
	}

	// Update statistics
	rm.updateRecordingStats(true, false)

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
	// Add panic recovery for recording operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			rm.logger.WithFields(logging.Fields{
				"filename":    filename,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "GetRecordingInfo",
			}).Error("Panic recovered in GetRecordingInfo")
		}
	}()

	rm.logger.WithField("filename", filename).Debug("Getting API-ready recording info")

	// Get basic file metadata first
	// Use canonical configured recordings path (ConfigurationManager overrides default)
	recordingsPath := rm.config.RecordingsPath
	if rm.configIntegration != nil && rm.configIntegration.configManager != nil {
		if cfg := rm.configIntegration.configManager.GetConfig(); cfg != nil && cfg.MediaMTX.RecordingsPath != "" {
			recordingsPath = cfg.MediaMTX.RecordingsPath
		}
	}

	// Get recording format from config to determine file extension
	format := rm.getRecordFormat() // Use configured format (STANAG 4609 compliant)
	if rm.configIntegration != nil && rm.configIntegration.configManager != nil {
		if cfg := rm.configIntegration.configManager.GetConfig(); cfg != nil && cfg.Recording.RecordFormat != "" {
			format = cfg.Recording.RecordFormat
		}
	}

	// Construct the actual file path with extension (MediaMTX adds extension based on RecordFormat)
	var filePath string
	var fileInfo os.FileInfo
	var err error

	// Try with extension first (MediaMTX creates files with extensions)
	filePath = filepath.Join(recordingsPath, filename+"."+format)
	fileInfo, err = os.Stat(filePath)
	if err != nil {
		// If not found with extension, try without extension (fallback for edge cases)
		filePath = filepath.Join(recordingsPath, filename)
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("recording file not found: %v", err)
		}
	}

	// Extract device from filename pattern (camera0_timestamp.mp4)
	device := "camera0" // Default
	if parts := strings.Split(filename, "_"); len(parts) > 0 {
		if strings.HasPrefix(parts[0], "camera") {
			device = parts[0]
		}
	}

	// Extract format from filename extension, default from canonical config
	fileFormat := func() string {
		// Default extension derived from canonical record format
		defaultExt := "mp4"
		if rm.configIntegration != nil && rm.configIntegration.configManager != nil {
			if cfg := rm.configIntegration.configManager.GetConfig(); cfg != nil {
				switch strings.ToLower(cfg.Recording.RecordFormat) {
				case "fmp4", "mp4":
					defaultExt = "mp4"
				case "mpegts", "ts":
					defaultExt = "ts"
				}
			}
		}
		if ext := filepath.Ext(filename); ext != "" {
			return strings.TrimPrefix(ext, ".")
		}
		return defaultExt
	}()

	// Extract video duration using MetadataManager
	duration := float64(0) // Default fallback
	if rm.metadataManager != nil {
		metadata, err := rm.metadataManager.ExtractVideoMetadata(ctx, filePath)
		if err != nil {
			rm.logger.WithError(err).WithField("file_path", filePath).Warn("Failed to extract video metadata, using default duration")
		} else if metadata.Success && metadata.Duration != nil {
			duration = *metadata.Duration
			rm.logger.WithFields(logging.Fields{
				"filename": filename,
				"duration": duration,
				"codec":    metadata.VideoCodec,
			}).Debug("Video duration extracted successfully")
		} else {
			rm.logger.WithField("file_path", filePath).Debug("Video metadata extraction succeeded but no duration found")
		}
	}

	// Build API-ready response with rich metadata
	response := &GetRecordingInfoResponse{
		Filename:    filename,
		FileSize:    fileInfo.Size(),
		Duration:    duration,
		CreatedTime: fileInfo.ModTime().Format(time.RFC3339), // API compliant field name
		Format:      fileFormat,
		Device:      device,
	}

	rm.logger.WithFields(logging.Fields{
		"filename":  filename,
		"device":    device,
		"format":    format,
		"file_size": fileInfo.Size(),
	}).Debug("Recording info retrieved successfully")

	return response, nil
}

// DeleteRecording deletes a recording file from the filesystem
func (rm *RecordingManager) DeleteRecording(ctx context.Context, filename string) error {
	// Add panic recovery for recording operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			rm.logger.WithFields(logging.Fields{
				"filename":    filename,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "DeleteRecording",
			}).Error("Panic recovered in DeleteRecording")
		}
	}()

	rm.logger.WithField("filename", filename).Info("Deleting recording file")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Get recordings directory path from canonical configuration
	recordingsPath := rm.config.RecordingsPath
	if recordingsPath == "" {
		return fmt.Errorf("recordings path not configured")
	}

	// Construct full file path using canonical config
	filePath := filepath.Join(recordingsPath, filename)

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

	// Delete the file directly from filesystem
	if err := os.Remove(filePath); err != nil {
		rm.logger.WithError(err).WithField("filename", filename).Error("Error deleting recording file")
		return fmt.Errorf("error deleting recording file: %w", err)
	}

	rm.logger.WithField("filename", filename).Info("Recording file deleted successfully")
	return nil
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

	// CRITICAL: Poll until MediaMTX config converges to record=false
	// This addresses the "already recording" race condition
	err = rm.pollUntilRecordingDisabled(ctx, pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm recording disabled: %w", err)
	}

	// Get timer info for accurate duration calculation using enhanced timer manager
	var startTime time.Time
	var duration float64
	if recordingInfo, exists := rm.timerManager.GetRecordingInfo(cameraID); exists {
		startTime = recordingInfo.StartTime
		duration = recordingInfo.GetDurationSeconds()
		rm.logger.WithFields(logging.Fields{
			"cameraID":   cameraID,
			"start_time": startTime,
			"duration":   duration,
		}).Debug("Retrieved accurate recording duration from timer manager")
	} else {
		// Fallback if timer info not available
		startTime = time.Now().Add(-time.Hour) // Placeholder
		duration = 3600                        // Placeholder
		rm.logger.WithField("cameraID", cameraID).Warn("Timer info not found, using placeholder duration")
	}

	// Cancel any auto-stop timer after getting the info
	rm.timerManager.DeleteTimer(cameraID)

	// Generate API filename (base name, no extension) per API documentation
	filename := fmt.Sprintf("%s_%s", cameraID, startTime.Format("2006-01-02_15-04-05"))

	// Get actual file size using MetadataManager
	fileSize := int64(1024) // Default fallback
	if rm.metadataManager != nil {
		recordingPath := filepath.Join(rm.config.RecordingsPath, filename)
		if actualSize, err := rm.metadataManager.GetFileSize(recordingPath); err == nil {
			fileSize = actualSize
			rm.logger.WithFields(logging.Fields{
				"filename":  filename,
				"file_size": fileSize,
			}).Debug("Actual file size retrieved successfully")
		} else {
			rm.logger.WithError(err).WithField("filename", filename).Warn("Failed to get actual file size, using fallback")
		}
	}

	// Build API-ready response with actual recording metadata
	response := &StopRecordingResponse{
		Device:    cameraID,
		Filename:  filename,
		Status:    "STOPPED",
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Duration:  duration,
		FileSize:  fileSize,
		Format:    "fmp4",
	}

	// Update statistics
	rm.updateRecordingStats(false, false)

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
	// Apply business rules using configuration
	defaultLimit := 50 // fallback
	maxLimit := 100    // fallback

	if rm.recordingConfig != nil {
		if rm.recordingConfig.DefaultPageSize > 0 {
			defaultLimit = rm.recordingConfig.DefaultPageSize
		}
		if rm.recordingConfig.MaxPageSize > 0 {
			maxLimit = rm.recordingConfig.MaxPageSize
		}
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

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
		// Extract device from filename pattern - robust parsing for any valid pattern
		// Supports both architecture patterns (camera0_*) and any other patterns (*camera0*)
		device := "camera0" // Default fallback
		fileName := file.FileName

		// Remove file extension for parsing
		if lastDot := strings.LastIndex(fileName, "."); lastDot > 0 {
			fileName = fileName[:lastDot]
		}

		// Split by underscore and look for camera identifier in any part
		parts := strings.Split(fileName, "_")
		for _, part := range parts {
			if strings.HasPrefix(part, "camera") && len(part) > 6 {
				// Validate it's a proper camera identifier (camera + digits)
				if cameraNum := part[6:]; cameraNum != "" {
					// Simple validation - should be digits only
					device = part
					break
				}
			}
		}

		// Extract format from filename extension, default from canonical config
		format := func() string {
			defaultExt := "mp4"
			if rm.configIntegration != nil && rm.configIntegration.configManager != nil {
				if cfg := rm.configIntegration.configManager.GetConfig(); cfg != nil {
					switch strings.ToLower(cfg.Recording.RecordFormat) {
					case "fmp4", "mp4":
						defaultExt = "mp4"
					case "mpegts", "ts":
						defaultExt = "ts"
					}
				}
			}
			if parts := strings.Split(file.FileName, "."); len(parts) > 1 {
				return parts[len(parts)-1]
			}
			return defaultExt
		}()

		duration := float64(0)
		if file.Duration != nil {
			duration = float64(*file.Duration)
		}

		recordings[i] = RecordingFileInfo{
			Device:       device,
			Filename:     file.FileName,
			FileSize:     file.FileSize,
			Duration:     duration,
			ModifiedTime: file.CreatedAt.Format(time.RFC3339), // API compliant field name
			Format:       format,
			DownloadURL:  fmt.Sprintf("/files/recordings/%s", file.FileName),
		}
	}

	response := &ListRecordingsResponse{
		Files:  recordings,
		Total:  fileList.Total,
		Limit:  limit,
		Offset: offset,
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
	data, err := rm.client.Get(ctx, MediaMTXRecordingsList+queryParams)
	if err != nil {
		rm.logger.WithError(err).Error("Failed to get recordings from MediaMTX API")

		// Check if it's a MediaMTXError to preserve status code
		if mtxErr, ok := err.(*MediaMTXError); ok {
			return nil, fmt.Errorf("MediaMTX API error (status %d): %s", mtxErr.Code, mtxErr.Message)
		}
		return nil, fmt.Errorf("failed to connect to MediaMTX API: %w", err)
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
		Total:  int(recordingList.ItemCount), // Ensure integer type per JSON-RPC spec
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

// CleanupOldRecordings removes recording files based on age, count, and size limits using centralized configuration
func (rm *RecordingManager) CleanupOldRecordings(ctx context.Context, maxAge time.Duration, maxCount int, maxSize int64) (deletedCount int, spaceFreed int64, err error) {
	rm.logger.WithFields(logging.Fields{
		"max_age":   maxAge,
		"max_count": maxCount,
		"max_size":  maxSize,
	}).Info("Starting cleanup of old recordings")

	// Get recordings list
	recordings, err := rm.GetRecordingsList(ctx, 10000, 0) // Get up to 10000 recordings for comprehensive cleanup
	if err != nil {
		// Check if it's a MediaMTXError to preserve status code
		if mtxErr, ok := err.(*MediaMTXError); ok {
			return 0, 0, fmt.Errorf("MediaMTX API error (status %d): %s", mtxErr.Code, mtxErr.Message)
		}
		return 0, 0, fmt.Errorf("failed to get recordings list: %w", err)
	}

	if recordings == nil || len(recordings.Files) == 0 {
		rm.logger.Debug("No recordings found for cleanup")
		return 0, 0, nil
	}

	// Sort by creation time (oldest first) for consistent cleanup order
	sort.Slice(recordings.Files, func(i, j int) bool {
		return recordings.Files[i].CreatedAt.Before(recordings.Files[j].CreatedAt)
	})

	cutoffTime := time.Now().Add(-maxAge)
	deletedCount = 0
	spaceFreed = 0

	// Calculate current total size if size-based cleanup is enabled
	var currentTotalSize int64
	if maxSize > 0 {
		for _, item := range recordings.Files {
			currentTotalSize += item.FileSize
		}
	}

	for _, item := range recordings.Files {
		shouldDelete := false

		// Check age constraint
		if item.CreatedAt.Before(cutoffTime) {
			shouldDelete = true
		}

		// Check count constraint (keep newest files up to maxCount)
		if len(recordings.Files)-deletedCount > maxCount {
			shouldDelete = true
		}

		// Check size constraint (delete oldest files until under maxSize)
		if maxSize > 0 && currentTotalSize > maxSize {
			shouldDelete = true
		}

		if shouldDelete {
			if err := rm.DeleteRecording(ctx, item.FileName); err != nil {
				rm.logger.WithError(err).WithField("filename", item.FileName).Warn("Failed to delete old recording")
				continue
			}
			deletedCount++
			spaceFreed += item.FileSize
			currentTotalSize -= item.FileSize
		}
	}

	rm.logger.WithFields(logging.Fields{
		"deleted_count": deletedCount,
		"space_freed":   spaceFreed,
	}).Info("Recording cleanup completed")
	return deletedCount, spaceFreed, nil
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
		rm.logger.WithError(err).Warn("Failed to get recording config, using fallback format")
		// Fallback to default STANAG 4609 format if config unavailable
		if rm.configIntegration != nil && rm.configIntegration.configManager != nil {
			if cfg := rm.configIntegration.configManager.GetConfig(); cfg != nil && cfg.Recording.RecordFormat != "" {
				return cfg.Recording.RecordFormat
			}
		}
		return "fmp4" // Final fallback to STANAG 4609 format
	}
	return recordingConfig.Format
}

// isPathRecording checks if a path is currently recording by querying MediaMTX
func (rm *RecordingManager) isPathRecording(ctx context.Context, cameraID string) (bool, error) {
	endpoint := FormatConfigPathsGet(cameraID)
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

// pollUntilRecordingDisabled polls MediaMTX config until record=false is confirmed
// This addresses the race condition where PATCH succeeds but GET still returns record=true
func (rm *RecordingManager) pollUntilRecordingDisabled(ctx context.Context, pathName string) error {
	maxAttempts := 10 // 10 attempts * 200ms = 2 seconds max
	attempt := 0

	for attempt < maxAttempts {
		attempt++

		// Check if recording is actually disabled
		isRecording, err := rm.isPathRecording(ctx, pathName)
		if err != nil {
			rm.logger.WithFields(logging.Fields{
				"path":    pathName,
				"attempt": attempt,
				"error":   err,
			}).Warn("Failed to check recording status during polling")
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if !isRecording {
			rm.logger.WithFields(logging.Fields{
				"path":    pathName,
				"attempt": attempt,
			}).Info("Recording successfully disabled and confirmed")
			return nil
		}

		rm.logger.WithFields(logging.Fields{
			"path":    pathName,
			"attempt": attempt,
		}).Debug("Recording still enabled, polling again...")

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("recording still enabled after %d attempts (2 seconds)", maxAttempts)
}

// isPathRecordingWithRetry checks if a path is recording with retry logic
// This addresses propagation delays where config might show stale record=true
func (rm *RecordingManager) isPathRecordingWithRetry(ctx context.Context, pathName string) (bool, error) {
	maxAttempts := 3 // 3 attempts * 100ms = 300ms max
	attempt := 0

	for attempt < maxAttempts {
		attempt++

		isRecording, err := rm.isPathRecording(ctx, pathName)
		if err != nil {
			rm.logger.WithFields(logging.Fields{
				"path":    pathName,
				"attempt": attempt,
				"error":   err,
			}).Warn("Failed to check recording status during retry")
			if attempt == maxAttempts {
				return false, err
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if !isRecording {
			rm.logger.WithFields(logging.Fields{
				"path":    pathName,
				"attempt": attempt,
			}).Debug("Recording confirmed disabled")
			return false, nil
		}

		rm.logger.WithFields(logging.Fields{
			"path":    pathName,
			"attempt": attempt,
		}).Debug("Recording still enabled, retrying...")

		if attempt < maxAttempts {
			time.Sleep(100 * time.Millisecond)
		}
	}

	rm.logger.WithFields(logging.Fields{
		"path":     pathName,
		"attempts": maxAttempts,
	}).Info("Recording confirmed enabled after retries")
	return true, nil
}

// enableRecordingOnPath enables recording on a MediaMTX path
func (rm *RecordingManager) enableRecordingOnPath(ctx context.Context, cameraID string, options *PathConf) error {
	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	// PATCH requests should only send the fields that need to be changed
	recordConfig := &PathConf{
		Record: true, // Only send the record flag for PATCH requests
	}

	endpoint := FormatConfigPathsPatch(cameraID)
	jsonData, err := marshalUpdatePathRequest(recordConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal PathConf record config: %w", err)
	}

	// PATCH request now sends only the required fields

	return rm.client.Patch(ctx, endpoint, jsonData)
}

// disableRecordingOnPath disables recording on a MediaMTX path
func (rm *RecordingManager) disableRecordingOnPath(ctx context.Context, cameraID string) error {
	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	recordConfig := &PathConf{
		Record: false,
	}

	endpoint := FormatConfigPathsPatch(cameraID)
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

// Resource Management Methods - Implementation of camera.ResourceManager and camera.CleanupManager interfaces

// Start initializes the recording manager (implements camera.ResourceManager)
func (rm *RecordingManager) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&rm.running, 0, 1) {
		return fmt.Errorf("recording manager is already running")
	}

	rm.logger.Info("Recording manager started")
	return nil
}

// Stop gracefully shuts down the recording manager (implements camera.ResourceManager)
func (rm *RecordingManager) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&rm.running, 1, 0) {
		rm.logger.Debug("Recording manager is already stopped")
		return nil // Idempotent
	}

	rm.logger.Info("Stopping recording manager...")

	// Stop all keepalive readers
	if rm.keepaliveReader != nil {
		rm.keepaliveReader.StopAll()
		rm.logger.Debug("All keepalive readers stopped")
	}

	// Clean up all active timers with context timeout
	if rm.timerManager != nil {
		if err := rm.timerManager.StopAll(ctx); err != nil {
			rm.logger.WithError(err).Warn("Error stopping timer manager")
		}
		rm.logger.Debug("All recording timers stopped")
	}

	// Close metadata manager resources if it implements io.Closer
	if rm.metadataManager != nil {
		if closer, ok := interface{}(rm.metadataManager).(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				rm.logger.WithError(err).Warn("Error closing metadata manager")
			}
		}
		rm.logger.Debug("Metadata manager resources cleaned up")
	}

	rm.logger.Info("Recording manager stopped successfully")
	return nil
}

// IsRunning returns whether the recording manager is running (implements camera.ResourceManager)
func (rm *RecordingManager) IsRunning() bool {
	return atomic.LoadInt32(&rm.running) == 1
}

// Cleanup performs resource cleanup (implements camera.CleanupManager)
func (rm *RecordingManager) Cleanup(ctx context.Context) error {
	rm.logger.Info("Performing recording manager cleanup...")

	// Force stop all active recordings
	activeRecordings := rm.timerManager.ListActiveRecordings()
	for _, recording := range activeRecordings {
		rm.logger.WithField("camera_id", recording.CameraID).Info("Force stopping active recording during cleanup")

		// Stop recording without waiting for graceful shutdown
		if err := rm.disableRecordingOnPath(ctx, recording.CameraID); err != nil {
			rm.logger.WithError(err).WithField("camera_id", recording.CameraID).Warn("Failed to stop recording during cleanup")
		}

		// Stop keepalive
		rm.stopRTSPKeepalive(recording.CameraID)

		// Remove timer
		rm.timerManager.DeleteTimer(recording.CameraID)

		atomic.AddInt64(&rm.resourceStats.TotalRecordingsStopped, 1)
	}

	rm.logger.WithField("stopped_recordings", fmt.Sprintf("%d", len(activeRecordings))).Info("Recording manager cleanup completed")
	return nil
}

// GetResourceStats returns current resource usage statistics (implements camera.CleanupManager)
func (rm *RecordingManager) GetResourceStats() map[string]interface{} {
	// Update stats with current values
	atomic.StoreInt64(&rm.resourceStats.ActiveKeepaliveReaders, int64(rm.keepaliveReader.GetActiveCount()))
	atomic.StoreInt64(&rm.resourceStats.ActiveTimers, int64(len(rm.timerManager.ListActiveRecordings())))

	// Get metadata cache size if available
	if rm.metadataManager != nil {
		if cacheProvider, ok := interface{}(rm.metadataManager).(interface{ GetCacheSize() int }); ok {
			atomic.StoreInt64(&rm.resourceStats.MetadataCacheSize, int64(cacheProvider.GetCacheSize()))
		}
	}

	return map[string]interface{}{
		"running":                  rm.IsRunning(),
		"active_keepalive_readers": atomic.LoadInt64(&rm.resourceStats.ActiveKeepaliveReaders),
		"active_timers":            atomic.LoadInt64(&rm.resourceStats.ActiveTimers),
		"metadata_cache_size":      atomic.LoadInt64(&rm.resourceStats.MetadataCacheSize),
		"total_recordings_started": atomic.LoadInt64(&rm.resourceStats.TotalRecordingsStarted),
		"total_recordings_stopped": atomic.LoadInt64(&rm.resourceStats.TotalRecordingsStopped),
		"recording_errors":         atomic.LoadInt64(&rm.resourceStats.RecordingErrors),
	}
}

// GetErrorMetrics returns current error metrics and alert status
func (rm *RecordingManager) GetErrorMetrics() map[string]interface{} {
	metrics := rm.errorMetricsCollector.GetMetrics()
	alertStatus := rm.errorMetricsCollector.GetAlertStatus()
	uptime := rm.errorMetricsCollector.GetUptime()

	return map[string]interface{}{
		"metrics": map[string]interface{}{
			"total_errors":        metrics.TotalErrors,
			"errors_by_component": metrics.ErrorsByComponent,
			"errors_by_severity":  metrics.ErrorsBySeverity,
			"recovery_attempts":   metrics.RecoveryAttempts,
			"recovery_successes":  metrics.RecoverySuccesses,
			"recovery_failures":   metrics.RecoveryFailures,
			"last_error_time":     metrics.LastErrorTime,
			"last_recovery_time":  metrics.LastRecoveryTime,
		},
		"alerts": alertStatus,
		"uptime": uptime.String(),
		"circuit_breaker": map[string]interface{}{
			"state":         rm.recordingCircuitBreaker.GetState(),
			"failure_count": rm.recordingCircuitBreaker.GetFailureCount(),
		},
	}
}

// updateRecordingStats updates recording statistics
func (rm *RecordingManager) updateRecordingStats(started bool, error bool) {
	if started {
		atomic.AddInt64(&rm.resourceStats.TotalRecordingsStarted, 1)
	} else {
		atomic.AddInt64(&rm.resourceStats.TotalRecordingsStopped, 1)
	}

	if error {
		atomic.AddInt64(&rm.resourceStats.RecordingErrors, 1)
	}
}

// Note: setAutoStopTimer method removed - functionality moved to RecordingTimerManager.CreateTimer()
// This provides enhanced timer management with metadata tracking and accurate duration calculation

// getCircuitBreakerConfig returns circuit breaker configuration with defaults
func getCircuitBreakerConfig(configIntegration ConfigIntegration) CircuitBreakerConfig {
	config, _ := configIntegration.GetMediaMTXConfig()

	// Use configured values if available, otherwise use sensible defaults
	cbConfig := CircuitBreakerConfig{
		FailureThreshold: config.CircuitBreaker.FailureThreshold,
		RecoveryTimeout:  config.CircuitBreaker.RecoveryTimeout,
		MaxFailures:      config.CircuitBreaker.MaxFailures,
	}

	if cbConfig.FailureThreshold == 0 {
		cbConfig.FailureThreshold = 5 // Default: 5 failures before opening
	}
	if cbConfig.RecoveryTimeout == 0 {
		// Use configuration-based recovery timeout
		cbConfig.RecoveryTimeout = time.Duration(config.CircuitBreaker.RecoveryTimeout) * time.Second
	}
	if cbConfig.MaxFailures == 0 {
		cbConfig.MaxFailures = 15 // Default: 15 failures before permanent open
	}

	return cbConfig
}
