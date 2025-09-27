/*
MediaMTX Stream Manager Implementation

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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// StreamMetadata represents metadata for stream tracking and analytics
type StreamMetadata struct {
	CameraID     string                 `json:"camera_id"`
	DevicePath   string                 `json:"device_path"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	ActivityLog  []StreamActivity       `json:"activity_log"`
	IsActive     bool                   `json:"is_active"`
	UseCase      StreamUseCase          `json:"use_case"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// StreamActivity represents an activity event in stream lifecycle
type StreamActivity struct {
	Timestamp   time.Time              `json:"timestamp"`
	Event       string                 `json:"event"` // "created", "started", "stopped", "error"
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// AddActivity adds an activity event to the stream metadata
func (sm *StreamMetadata) AddActivity(event, description string, data map[string]interface{}) {
	activity := StreamActivity{
		Timestamp:   time.Now(),
		Event:       event,
		Description: description,
		Data:        data,
	}

	sm.ActivityLog = append(sm.ActivityLog, activity)
	sm.LastActivity = activity.Timestamp

	// Keep only last 50 activities to prevent memory bloat
	if len(sm.ActivityLog) > 50 {
		sm.ActivityLog = sm.ActivityLog[len(sm.ActivityLog)-50:]
	}
}

// GetDuration returns the duration since stream start
func (sm *StreamMetadata) GetDuration() time.Duration {
	if sm.StartTime.IsZero() {
		return 0
	}
	return time.Since(sm.StartTime)
}

// GetLastActivityISO returns last activity time in ISO 8601 format
func (sm *StreamMetadata) GetLastActivityISO() string {
	if sm.LastActivity.IsZero() {
		return ""
	}
	return sm.LastActivity.Format(time.RFC3339)
}

// GetStartTimeISO returns start time in ISO 8601 format
func (sm *StreamMetadata) GetStartTimeISO() string {
	if sm.StartTime.IsZero() {
		return ""
	}
	return sm.StartTime.Format(time.RFC3339)
}

// streamManager manages MediaMTX stream lifecycle and FFmpeg process coordination.
//
// RESPONSIBILITIES:
// - Stream creation and lifecycle management via MediaMTX API
// - FFmpeg command generation and process coordination
// - Stream status monitoring and health checks
// - API-ready response formatting for stream operations
// - Enhanced start time and activity tracking
//
// ARCHITECTURE:
// - Operates with cameraID as primary identifier
// - Converts to devicePath only when generating FFmpeg commands
// - Uses MediaMTX api_types.go for all operations
// - Enhanced metadata tracking for stream analytics
//
// API INTEGRATION:
// - Returns JSON-RPC API-ready responses
// - Handles both V4L2 devices and external RTSP streams
type streamManager struct {
	client            MediaMTXClient
	pathManager       PathManager
	config            *config.MediaMTXConfig
	recordingConfig   *config.RecordingConfig
	configIntegration *ConfigIntegration
	ffmpegManager     FFmpegManager
	logger            *logging.Logger
	useCaseConfigs    map[StreamUseCase]UseCaseConfig
	keepaliveReader   *RTSPKeepaliveReader

	// Enhanced stream tracking with metadata
	streamMetadata sync.Map // cameraID -> *StreamMetadata
}

// NewStreamManager creates a new MediaMTX stream manager
// OPTIMIZED: Accept PathManager instead of creating a new one to ensure single instance
func NewStreamManager(client MediaMTXClient, pathManager PathManager, config *config.MediaMTXConfig, recordingConfig *config.RecordingConfig, configIntegration *ConfigIntegration, ffmpegManager FFmpegManager, logger *logging.Logger) StreamManager {
	// Fail fast if required dependencies are nil
	if client == nil {
		panic("MediaMTXClient cannot be nil")
	}
	if pathManager == nil {
		panic("PathManager cannot be nil")
	}
	if config == nil {
		panic("MediaMTXConfig cannot be nil")
	}
	if logger == nil {
		panic("Logger cannot be nil")
	}

	// All operations use the same stable path with consistent settings
	useCaseConfigs := map[StreamUseCase]UseCaseConfig{
		UseCaseRecording: {
			RunOnDemandCloseAfter:   config.RunOnDemandCloseAfter,   // Use centralized config
			RunOnDemandRestart:      true,                           // Always restart FFmpeg if it crashes
			RunOnDemandStartTimeout: config.RunOnDemandStartTimeout, // Use centralized config
			Suffix:                  "",                             // No suffix - simple path names
		},
	}

	return &streamManager{
		client:            client,
		pathManager:       pathManager,
		config:            config,
		recordingConfig:   recordingConfig,
		configIntegration: configIntegration,
		ffmpegManager:     ffmpegManager,
		logger:            logger,
		useCaseConfigs:    useCaseConfigs,
		keepaliveReader:   NewRTSPKeepaliveReader(config, logger), // ADD THIS
	}
}

// StartStream starts a stream for a camera using cameraID-first architecture
func (sm *streamManager) StartStream(ctx context.Context, cameraID string) (*StartStreamingResponse, error) {
	// Add panic recovery for stream operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			sm.logger.WithFields(logging.Fields{
				"camera_id":   cameraID,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "StartStream",
			}).Error("Panic recovered in StartStream")
		}
	}()

	// Validate dependencies are initialized
	if sm.pathManager == nil {
		return nil, fmt.Errorf("PathManager not initialized")
	}
	if sm.client == nil {
		return nil, fmt.Errorf("MediaMTXClient not initialized")
	}

	sm.logger.WithField("cameraID", cameraID).Info("Starting stream with cameraID-first approach")

	// Pure delegation to startStreamForUseCase - no conversion ping-pong!
	_, err := sm.startStreamForUseCase(ctx, cameraID, UseCaseRecording)
	if err != nil {
		return nil, err
	}

	// Build API-ready response
	streamName := fmt.Sprintf("camera_%s_viewing", cameraID)
	streamURL := sm.GenerateStreamURL(cameraID)
	response := &StartStreamingResponse{
		Device:         cameraID,
		StreamName:     streamName,
		StreamURL:      streamURL,
		Status:         "STARTED",
		StartTime:      time.Now().Format(time.RFC3339),
		AutoCloseAfter: "300s",
		FfmpegCommand:  fmt.Sprintf("ffmpeg -f v4l2 -i /dev/%s -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp %s", cameraID, streamURL),
	}

	return response, nil
}

// startStreamForUseCase starts a stream for the specified use case
func (sm *streamManager) startStreamForUseCase(ctx context.Context, cameraID string, useCase StreamUseCase) (*Path, error) {
	// Add panic recovery for stream operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			sm.logger.WithFields(logging.Fields{
				"camera_id":   cameraID,
				"use_case":    useCase,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "startStreamForUseCase",
			}).Error("Panic recovered in startStreamForUseCase")
		}
	}()

	// Get devicePath only when needed for FFmpeg command generation
	devicePath, exists := sm.pathManager.GetDevicePathForCamera(cameraID)
	if !exists {
		// For external streams, cameraID is the source
		devicePath = cameraID
	}

	// Validate device path
	if err := sm.validateDevicePath(devicePath); err != nil {
		return nil, fmt.Errorf("failed to validate device path %s: %w", devicePath, err)
	}

	// Use cameraID directly as MediaMTX path name
	streamName := cameraID
	sm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"device_path": devicePath,
		"use_case":    useCase,
		"stream_name": streamName,
	}).Info("Starting stream with cameraID as path name")

	// Create or update stream metadata for tracking
	now := time.Now()
	metadata := &StreamMetadata{
		CameraID:     cameraID,
		DevicePath:   devicePath,
		StartTime:    now,
		LastActivity: now,
		ActivityLog:  []StreamActivity{},
		IsActive:     false, // Will be set to true when stream is successfully created
		UseCase:      useCase,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    now,
	}

	metadata.AddActivity("creating", "Stream creation initiated", map[string]interface{}{
		"use_case":    useCase,
		"device_path": devicePath,
		"stream_name": streamName,
	})

	// Store metadata for tracking
	sm.streamMetadata.Store(cameraID, metadata)

	// Build FFmpeg command using injected FFmpegManager
	ffmpegCommand, err := sm.ffmpegManager.BuildRunOnDemandCommand(devicePath, streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to build FFmpeg command: %w", err)
	}

	// All use cases use the same configuration - stable path for streaming AND recording
	pathConfig := &PathConf{
		RunOnDemand:             ffmpegCommand,
		RunOnDemandRestart:      true,                              // Always restart FFmpeg if it crashes
		RunOnDemandStartTimeout: sm.config.RunOnDemandStartTimeout, // Use configured timeout
		RunOnDemandCloseAfter:   sm.config.RunOnDemandCloseAfter,   // Use configured timeout
		RunOnUnDemand:           "",
	}

	// Use PathManager for proper architectural integration and validation
	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"device_path": devicePath,
		"path_config": pathConfig,
	}).Info("About to create MediaMTX path")

	err = sm.pathManager.CreatePath(ctx, streamName, devicePath, pathConfig)
	if err != nil {
		// Log the actual error for debugging
		sm.logger.WithError(err).WithFields(logging.Fields{
			"stream_name": streamName,
			"device_path": devicePath,
			"path_config": pathConfig,
		}).Error("CreatePath failed - investigating error")

		// Check if this is a "path exists" error (idempotent success)
		errorMsg := err.Error()
		sm.logger.WithField("error_message", errorMsg).Error("CreatePath error message")

		// Check both the error message and details for the specific error text
		isAlreadyExists := strings.Contains(errorMsg, "path already exists") ||
			strings.Contains(errorMsg, "already exists")

		// Also check the details field for MediaMTXError
		if mediaMTXErr, ok := err.(*MediaMTXError); ok {
			isAlreadyExists = isAlreadyExists ||
				strings.Contains(mediaMTXErr.Details, "path already exists") ||
				strings.Contains(mediaMTXErr.Details, "already exists")
		}

		if isAlreadyExists {
			sm.logger.WithField("stream_name", streamName).Info("MediaMTX path already exists, treating as success")

			// Update metadata for existing stream
			if metadata, exists := sm.streamMetadata.Load(cameraID); exists {
				streamMeta := metadata.(*StreamMetadata)
				streamMeta.AddActivity("reused", "Existing MediaMTX path reused", map[string]interface{}{
					"stream_name": streamName,
				})
				streamMeta.IsActive = true
			}

			// Return a mock stream response for idempotent success
			stream := &Path{
				Name:     streamName,
				ConfName: streamName,
				Ready:    false,
				Tracks:   []string{},
				Readers:  []PathReader{},
			}
			return stream, nil
		}
		return nil, fmt.Errorf(streamName, "create_stream", "failed to create stream", err)
	}

	// PathManager.CreatePath succeeded - create stream response
	sm.logger.WithField("stream_name", streamName).Info("MediaMTX path created successfully")

	// Update metadata for successful creation
	if metadata, exists := sm.streamMetadata.Load(cameraID); exists {
		streamMeta := metadata.(*StreamMetadata)
		streamMeta.AddActivity("created", "MediaMTX path created successfully", map[string]interface{}{
			"stream_name": streamName,
			"on_demand":   true,
		})
		streamMeta.IsActive = true
	}

	stream := &Path{
		Name:     streamName,
		ConfName: streamName,
		Ready:    false, // On-demand paths are never ready until accessed
		Tracks:   []string{},
		Readers:  []PathReader{},
	}

	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"use_case":    useCase,
		"device_path": devicePath,
		"on_demand":   true,
	}).Info("MediaMTX on-demand stream created successfully - will activate on first access")

	return stream, nil
}

// validateDevicePath validates device path format and accessibility
func (sm *streamManager) validateDevicePath(devicePath string) error {
	if devicePath == "" {
		return fmt.Errorf("device path cannot be empty")
	}

	// Validate device path format - accept both local devices, external RTSP sources, and abstract camera identifiers
	// According to architecture: camera identifiers (camera0, camera1) are valid at API abstraction layer
	if !strings.HasPrefix(devicePath, "/dev/video") &&
		!strings.HasPrefix(devicePath, "/dev/custom") &&
		!strings.HasPrefix(devicePath, "rtsp://") &&
		!strings.HasPrefix(devicePath, "rtmp://") &&
		!strings.HasPrefix(devicePath, "camera") {
		return fmt.Errorf("invalid device path format: %s. Must be /dev/video<N>, /dev/custom<name>, rtsp://, rtmp://, or camera<N>", devicePath)
	}

	return nil
}

// GenerateStreamName generates stream name using cameraID-first approach
// SIMPLIFIED: All use cases return the same stable path name (camera0, camera1, etc.)
// This aligns with MediaMTX architecture where one path handles streaming AND recording
func (sm *streamManager) GenerateStreamName(cameraID string, useCase StreamUseCase) string {
	// Centralized naming: passthrough cameraID
	return cameraID
}

// CreateStream creates a new stream with automatic USB device handling
func (sm *streamManager) CreateStream(ctx context.Context, name, source string) (*Path, error) {
	sm.logger.WithFields(logging.Fields{
		"name":   name,
		"source": source,
	}).Debug("Creating MediaMTX stream")

	// Validate stream name
	if name == "" {
		return nil, fmt.Errorf("stream name cannot be empty")
	}

	// Validate source
	if source == "" {
		return nil, fmt.Errorf("source cannot be empty")
	}

	// Check if source is a USB device path (starts with /dev/video)
	if strings.HasPrefix(source, "/dev/video") {
		// Create FFmpeg command for USB device publishing using injected FFmpegManager
		ffmpegCommand, err := sm.ffmpegManager.BuildRunOnDemandCommand(source, name)
		if err != nil {
			return nil, fmt.Errorf("failed to build FFmpeg command: %w", err)
		}

		// Create path configuration for USB device
		pathConfig := &PathConf{
			RunOnDemand:             ffmpegCommand,
			RunOnDemandRestart:      true,
			RunOnDemandStartTimeout: sm.config.RunOnDemandStartTimeout,
			RunOnDemandCloseAfter:   sm.config.RunOnDemandCloseAfter,
			RunOnUnDemand:           "",
		}

		// Use PathManager for proper architectural integration and validation
		err = sm.pathManager.CreatePath(ctx, name, source, pathConfig)
		if err != nil {
			// Check if this is a "path exists" error (idempotent success)
			errorMsg := err.Error()
			isAlreadyExists := strings.Contains(errorMsg, "path already exists") ||
				strings.Contains(errorMsg, "already exists")

			// Also check the details field for MediaMTXError
			if mediaMTXErr, ok := err.(*MediaMTXError); ok {
				isAlreadyExists = isAlreadyExists ||
					strings.Contains(mediaMTXErr.Details, "path already exists") ||
					strings.Contains(mediaMTXErr.Details, "already exists")
			}

			if isAlreadyExists {
				sm.logger.WithField("stream_name", name).Info("MediaMTX path already exists, treating as success")
				// Return a mock stream response for idempotent success
				stream := &Path{
					Name:     name,
					ConfName: name,
					Ready:    false,
					Tracks:   []string{},
					Readers:  []PathReader{},
				}
				return stream, nil
			}
			return nil, fmt.Errorf(name, "create_stream", "failed to create stream", err)
		}

		// PathManager.CreatePath succeeded - create stream response
		stream := &Path{
			Name:     name,
			ConfName: name,
			Ready:    false,
			Tracks:   []string{},
			Readers:  []PathReader{},
		}
		sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully with FFmpeg publishing")
		return stream, nil
	} else {
		// For non-USB sources (RTSP URLs, etc.), use direct source
		// Create path configuration for direct source
		pathConfig := &PathConf{
			Source: source,
		}

		// Special handling for "publisher" source - use ConfigIntegration to build proper configuration
		if source == "publisher" {
			// Use ConfigIntegration to build proper PathConf with RunOnDemand configuration
			// This follows the established architecture pattern
			builtPathConf, err := sm.configIntegration.BuildPathConf(name, nil, false)
			if err != nil {
				return nil, fmt.Errorf("failed to build path configuration for publisher source: %w", err)
			}
			pathConfig = builtPathConf
		}

		// Use PathManager for proper architectural integration
		err := sm.pathManager.CreatePath(ctx, name, source, pathConfig)
		if err != nil {
			// Check if this is a "path exists" error (idempotent success)
			errorMsg := err.Error()
			isAlreadyExists := strings.Contains(errorMsg, "path already exists") ||
				strings.Contains(errorMsg, "already exists")

			// Also check the details field for MediaMTXError
			if mediaMTXErr, ok := err.(*MediaMTXError); ok {
				isAlreadyExists = isAlreadyExists ||
					strings.Contains(mediaMTXErr.Details, "path already exists") ||
					strings.Contains(mediaMTXErr.Details, "already exists")
			}

			if isAlreadyExists {
				sm.logger.WithField("stream_name", name).Info("MediaMTX path already exists, treating as success")
				// Return a mock stream response for idempotent success
				stream := &Path{
					Name:     name,
					ConfName: name,
					Ready:    false,
					Tracks:   []string{},
					Readers:  []PathReader{},
				}
				return stream, nil
			}
			return nil, fmt.Errorf(name, "create_stream", "failed to create stream", err)
		}

		// PathManager.CreatePath succeeded - create stream response
		stream := &Path{
			Name:     name,
			ConfName: name,
			Ready:    false,
			Tracks:   []string{},
			Readers:  []PathReader{},
		}
		sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully")
		return stream, nil
	}
}

// DeleteStream deletes a stream
func (sm *streamManager) DeleteStream(ctx context.Context, id string) error {
	sm.logger.WithField("stream_id", id).Debug("Deleting MediaMTX stream")

	// Use PathManager for proper architectural integration
	err := sm.pathManager.DeletePath(ctx, id)
	if err != nil {
		return fmt.Errorf(id, "delete_stream", "failed to delete stream", err)
	}

	sm.logger.WithField("stream_id", id).Info("MediaMTX stream deleted successfully")
	return nil
}

// GetStream gets a specific stream
func (sm *streamManager) GetStream(ctx context.Context, id string) (*Path, error) {
	sm.logger.WithField("stream_id", id).Debug("Getting MediaMTX stream")

	// Use PathManager for proper architectural integration
	path, err := sm.pathManager.GetPath(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(id, "get_stream", "failed to get stream", err)
	}

	// Convert Path to Stream
	stream := &Path{
		Name:     path.Name,
		ConfName: path.Name,      // Use name as confName since Path doesn't have ConfName
		Ready:    false,          // Path doesn't have Ready field, default to false
		Tracks:   []string{},     // Path doesn't have Tracks field, default to empty
		Readers:  []PathReader{}, // Path doesn't have Readers field, default to empty
	}

	return stream, nil
}

// ListStreams lists all streams
func (sm *streamManager) ListStreams(ctx context.Context) (*GetStreamsResponse, error) {
	sm.logger.Debug("Listing MediaMTX streams with cameraID-first approach")

	// Get runtime paths for actual stream status
	runtimePaths, err := sm.pathManager.GetRuntimePaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	// Convert to API-ready StreamInfo format
	streamInfos := make([]StreamInfo, len(runtimePaths))
	activeCount := 0
	inactiveCount := 0

	for i, path := range runtimePaths {
		status := "inactive"
		if path.Ready {
			status = "active"
			activeCount++
		} else {
			inactiveCount++
		}

		startTime := ""
		lastActivity := ""
		if path.ReadyTime != nil {
			startTime = *path.ReadyTime
			lastActivity = startTime
		}

		// Convert PathSource to string
		sourceStr := ""
		if path.Source != nil {
			sourceStr = path.Source.Type
		}

		streamInfos[i] = StreamInfo{
			Name:         path.Name, // This is already cameraID (camera0, camera1)
			Status:       status,
			Source:       sourceStr,
			Viewers:      len(path.Readers),
			StartTime:    startTime,
			LastActivity: lastActivity,
			BytesSent:    path.BytesSent,
		}
	}

	// Build API-ready response
	response := &GetStreamsResponse{
		Streams:   streamInfos,
		Total:     len(runtimePaths),
		Active:    activeCount,
		Inactive:  inactiveCount,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	sm.logger.WithField("count", fmt.Sprintf("%d", len(runtimePaths))).Debug("MediaMTX streams listed successfully")
	return response, nil
}

// MonitorStream monitors a stream
func (sm *streamManager) MonitorStream(ctx context.Context, id string) error {
	sm.logger.WithField("stream_id", id).Debug("Monitoring MediaMTX stream")

	// Get stream status
	status, err := sm.GetStreamStatus(ctx, id)
	if err != nil {
		return fmt.Errorf(id, "monitor_stream", "failed to get stream status", err)
	}

	sm.logger.WithFields(logging.Fields{
		"stream_id": id,
		"status":    status,
	}).Debug("MediaMTX stream status")

	return nil
}

// GetStreamStatus gets the status of a stream
func (sm *streamManager) GetStreamStatus(ctx context.Context, cameraID string) (*GetStreamStatusResponse, error) {
	sm.logger.WithField("cameraID", cameraID).Debug("Getting stream status with cameraID-first approach")

	// Use cameraID directly as MediaMTX path name
	stream, err := sm.GetStream(ctx, cameraID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream status for %s: %w", cameraID, err)
	}

	// Build API-ready response using existing determineStatus method
	status := determineStatus(stream.Ready)

	response := &GetStreamStatusResponse{
		Device:       cameraID,
		StreamName:   cameraID + "_video_stream", // Generate stream name
		Status:       status,
		Ready:        stream.Ready, // Use actual readiness status
		StreamURL:    sm.GenerateStreamURL(cameraID),
		Viewers:      len(stream.Readers),
		BytesSent:    stream.BytesSent,
		StartTime:    "",
		LastActivity: "",
	}

	// Enhanced start time and activity tracking using stream metadata
	if metadata, exists := sm.streamMetadata.Load(cameraID); exists {
		streamMeta := metadata.(*StreamMetadata)
		response.StartTime = streamMeta.GetStartTimeISO()
		response.LastActivity = streamMeta.GetLastActivityISO()

		sm.logger.WithFields(logging.Fields{
			"camera_id":     cameraID,
			"start_time":    response.StartTime,
			"last_activity": response.LastActivity,
			"duration":      streamMeta.GetDuration(),
		}).Debug("Enhanced stream status with metadata tracking")
	} else {
		// Fallback to stream ReadyTime if metadata not available
		if stream.ReadyTime != nil {
			response.StartTime = *stream.ReadyTime
			response.LastActivity = *stream.ReadyTime
		}
	}

	return response, nil
}

// CheckStreamReadiness checks if a stream is ready for operations (enhanced existing stream manager)
func (sm *streamManager) CheckStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error) {
	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"timeout":     timeout,
	}).Debug("Checking stream readiness")

	// Get current stream status from MediaMTX using PathManager
	paths, err := sm.pathManager.ListPaths(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get MediaMTX active paths: %w", err)
	}

	// Find the specific stream
	for _, path := range paths {
		if path.Name == streamName {
			// Path struct doesn't have Ready field, so we'll assume it's ready if it exists
			// In a real implementation, we might need to check the actual MediaMTX API for readiness
			sm.logger.WithField("stream_name", streamName).Debug("Stream found, assuming ready")
			return true, nil
		}
	}

	return false, fmt.Errorf("stream %s not found", streamName)
}

// WaitForStreamReadiness waits for a stream to become ready (enhanced existing stream manager)
func (sm *streamManager) WaitForStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error) {
	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"timeout":     timeout,
	}).Info("Waiting for stream readiness")

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use dedicated stream manager ticker interval for optimal performance (configurable)
	tickerInterval := time.Duration(sm.config.StreamReadiness.StreamManagerTickerInterval * float64(time.Second))
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return false, fmt.Errorf("timeout waiting for stream %s to become ready", streamName)
		case <-ticker.C:
			ready, err := sm.CheckStreamReadiness(ctx, streamName, 1*time.Second)
			if err != nil {
				sm.logger.WithError(err).WithField("stream_name", streamName).Debug("Stream readiness check failed, continuing")
				continue
			}
			if ready {
				sm.logger.WithField("stream_name", streamName).Info("Stream became ready")
				return true, nil
			}
		}
	}
}

// StopStream stops the stream for a device (simplified - single path)
func (sm *streamManager) StopStream(ctx context.Context, cameraID string) error {
	sm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"action":   "stop_stream",
	}).Info("Stopping stream using cameraID-first approach")

	// Use cameraID directly as MediaMTX path name (no conversion!)
	streamName := cameraID

	// Delete the stream from MediaMTX
	err := sm.DeleteStream(ctx, streamName)
	if err != nil {
		// Update metadata for error
		if metadata, exists := sm.streamMetadata.Load(cameraID); exists {
			streamMeta := metadata.(*StreamMetadata)
			streamMeta.AddActivity("error", "Failed to stop stream", map[string]interface{}{
				"error": err.Error(),
			})
		}

		sm.logger.WithFields(logging.Fields{
			"cameraID":    cameraID,
			"stream_name": streamName,
			"error":       err.Error(),
		}).Error("Failed to stop stream")
		return fmt.Errorf("failed to stop stream: %w", err)
	}

	// Update metadata for successful stop and clean up
	if metadata, exists := sm.streamMetadata.Load(cameraID); exists {
		streamMeta := metadata.(*StreamMetadata)
		streamMeta.AddActivity("stopped", "Stream stopped successfully", map[string]interface{}{
			"stream_name": streamName,
			"duration":    streamMeta.GetDuration().String(),
		})
		streamMeta.IsActive = false

		// Clean up metadata after successful stop (optional)
		// sm.streamMetadata.Delete(cameraID)
	}

	sm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"stream_name": streamName,
	}).Info("Stream stopped successfully")

	return nil
}

// GenerateStreamURL generates the RTSP URL for a stream
func (sm *streamManager) GenerateStreamURL(streamName string) string {
	return fmt.Sprintf("rtsp://%s:%d/%s", sm.config.Host, sm.config.RTSPPort, streamName)
}

// GetStreamURL returns stream URL with status checking - consolidates Controller business logic
func (sm *streamManager) GetStreamURL(ctx context.Context, cameraID string) (*GetStreamURLResponse, error) {
	sm.logger.WithField("cameraID", cameraID).Debug("Getting stream URL with status")

	// Generate stream URL using cameraID directly
	streamURL := sm.GenerateStreamURL(cameraID)

	// Get actual stream status from MediaMTX
	streamStatus, err := sm.GetStreamStatus(ctx, cameraID)
	streamName := fmt.Sprintf("camera_%s_viewing", cameraID)

	if err != nil {
		// Stream doesn't exist or error - return available URL anyway
		return &GetStreamURLResponse{
			Device:          cameraID,
			StreamName:      streamName,
			StreamURL:       streamURL,
			Available:       true,
			ActiveConsumers: 0,
			StreamStatus:    "NOT_READY",
		}, nil
	}

	// Use actual status from StreamManager
	response := &GetStreamURLResponse{
		Device:          cameraID,
		StreamName:      streamName,
		StreamURL:       streamURL,
		Available:       streamStatus.Status == "active",
		ActiveConsumers: streamStatus.Viewers,
		StreamStatus: func() string {
			if streamStatus.Status == "active" {
				return "READY"
			} else {
				return "NOT_READY"
			}
		}(),
	}

	return response, nil
}

// EnableRecording enables recording on the stable path for a camera
func (sm *streamManager) EnableRecording(ctx context.Context, cameraID string) error {
	// Add panic recovery for stream operations
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			sm.logger.WithFields(logging.Fields{
				"camera_id":   cameraID,
				"panic":       r,
				"stack_trace": string(stack[:length]),
				"operation":   "EnableRecording",
			}).Error("Panic recovered in EnableRecording")
		}
	}()

	// Serialize create→ready→patch operations per path using per-path mutex
	pathMutex := sm.pathManager.(*pathManager).getPathMutex(cameraID)
	pathMutex.Lock()
	defer pathMutex.Unlock()

	// Ensure the path exists (idempotent)
	stream, err := sm.startStreamForUseCase(ctx, cameraID, UseCaseRecording)
	if err != nil {
		return fmt.Errorf("failed to ensure path exists: %w", err)
	}

	sm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"stream_name": stream.Name,
	}).Info("Path ensured, starting keepalive reader for recording")

	// START KEEPALIVE READER - This is the KEY CHANGE
	// This creates an RTSP connection that triggers runOnDemand
	err = sm.keepaliveReader.StartKeepalive(ctx, cameraID)
	if err != nil {
		return fmt.Errorf("failed to start keepalive reader: %w", err)
	}

	// Wait for the FFmpeg publisher to start using context-aware timeout
	select {
	case <-time.After(time.Duration(sm.config.StreamReadiness.CheckInterval * float64(time.Second))):
		// FFmpeg publisher should be started now
	case <-ctx.Done():
		// Context cancelled, return early
		return fmt.Errorf("context cancelled while waiting for FFmpeg publisher to start")
	}

	sm.logger.WithField("cameraID", cameraID).Info("Keepalive reader active, enabling recording")

	// Generate recording path from configuration
	recordPath := GenerateRecordingPath(sm.config, sm.recordingConfig)

	// Pre-create recording directory
	outputDir := sm.config.RecordingsPath
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		// Stop keepalive on error
		sm.keepaliveReader.StopKeepalive(cameraID)
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Create recording configuration
	recordingConfig := sm.createRecordingConfig(cameraID, recordPath)

	// PATCH the path to enable recording using PathConf directly
	err = sm.pathManager.PatchPath(ctx, cameraID, recordingConfig)
	if err != nil {
		// Stop keepalive on error
		sm.keepaliveReader.StopKeepalive(cameraID)
		return fmt.Errorf("failed to enable recording: %w", err)
	}

	sm.logger.WithFields(logging.Fields{
		"cameraID":      cameraID,
		"record_path":   recordPath,
		"has_keepalive": true,
	}).Info("Recording enabled successfully with keepalive reader")

	return nil
}

// DisableRecording disables recording on the stable path for a camera
func (sm *streamManager) DisableRecording(ctx context.Context, cameraID string) error {

	// Serialize operations per path using per-path mutex
	pathMutex := sm.pathManager.(*pathManager).getPathMutex(cameraID)
	pathMutex.Lock()
	defer pathMutex.Unlock()

	sm.logger.WithField("cameraID", cameraID).Info("Disabling recording")

	// Create disable configuration
	disableConfig := &PathConf{
		Record: false,
	}

	// PATCH the path to disable recording
	err := sm.pathManager.PatchPath(ctx, cameraID, disableConfig)
	if err != nil {
		sm.logger.WithError(err).Error("Failed to disable recording config")
		// Continue to stop keepalive even if patch fails
	}

	// STOP KEEPALIVE READER - This is the KEY CHANGE
	// This closes the RTSP connection, allowing runOnDemand to stop FFmpeg
	err = sm.keepaliveReader.StopKeepalive(cameraID)
	if err != nil {
		sm.logger.WithError(err).Warn("Failed to stop keepalive reader")
	}

	sm.logger.WithFields(logging.Fields{
		"cameraID":          cameraID,
		"keepalive_stopped": true,
	}).Info("Recording disabled and keepalive reader stopped")

	return nil
}

// createRecordingConfig creates the recording configuration for MediaMTX PATCH
func (sm *streamManager) createRecordingConfig(cameraID, outputPath string) *PathConf {
	// Generate recordPath with timestamp pattern
	recordPath := sm.getRecordingOutputPath(cameraID, outputPath)

	// Use PathConf from api_types.go - single source of truth with MediaMTX swagger.json
	config := &PathConf{
		Record:                true,
		RecordPath:            recordPath,
		RecordFormat:          "fmp4", // STANAG 4609 compatible
		RecordPartDuration:    sm.config.RecordPartDuration,
		RecordMaxPartSize:     "100MB",
		RecordSegmentDuration: sm.config.RecordSegmentDuration,
		RecordDeleteAfter:     sm.config.RecordDeleteAfter,
	}

	sm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"record_path": recordPath,
		"config":      config,
	}).Debug("Created recording configuration for PATCH")

	return config
}

// getRecordingOutputPath generates the recordPath with timestamp pattern
func (sm *streamManager) getRecordingOutputPath(cameraID, outputPath string) string {
	if outputPath != "" {
		dir := filepath.Dir(outputPath)
		// MediaMTX requires %path in recordPath - single % not double %%
		// No extension - MediaMTX adds it based on recordFormat
		return filepath.Join(dir, "%path_%Y-%m-%d_%H-%M-%S")
	}
	// Use centralized configuration for default recordings path
	cfg := sm.configIntegration.configManager.GetConfig()
	if cfg != nil && cfg.MediaMTX.RecordingsPath != "" {
		return filepath.Join(cfg.MediaMTX.RecordingsPath, "%path_%Y-%m-%d_%H-%M-%S")
	}
	return "%path_%Y-%m-%d_%H-%M-%S"
}

// Helper method to check device type and determine if keepalive is needed
// shouldUseKeepalive is UNUSED by design - keepalive decisions moved to recording-specific context
// ARCHITECTURE: Keepalive only needed for MediaMTX on-demand recording activation
// SCOPE: Not needed for general streaming, only for recording operations
func (sm *streamManager) shouldUseKeepalive(devicePath string) bool {
	// For V4L2 devices using runOnDemand, always need keepalive
	if strings.HasPrefix(devicePath, "/dev/video") {
		return true
	}

	// For external RTSP sources, check if they're configured with sourceOnDemand
	if strings.HasPrefix(devicePath, "rtsp://") {
		// Could check path configuration here to determine if keepalive needed
		// ARCHITECTURE INVESTIGATION REQUIRED: External RTSP sources are ALSO on-demand in MediaMTX
		// CURRENT ASSUMPTION: External sources don't need keepalive (MAY BE INCORRECT)
		// REALITY: MediaMTX treats ALL sources as on-demand, including external RTSP
		// QUESTION: Do external RTSP sources need keepalive for immediate recording start?
		// INVESTIGATION NEEDED: Test recording start time with/without keepalive for external sources
		return false
	}

	// Default to using keepalive for safety
	return true
}
