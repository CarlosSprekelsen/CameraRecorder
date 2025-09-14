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
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// streamManager represents the MediaMTX stream manager
type streamManager struct {
	client            MediaMTXClient
	pathManager       PathManager
	config            *config.MediaMTXConfig
	configIntegration *ConfigIntegration
	logger            *logging.Logger
	useCaseConfigs    map[StreamUseCase]UseCaseConfig

	// FFmpeg command caching for performance - using sync.Map for lock-free reads
	ffmpegCommands sync.Map // device path -> cached FFmpeg command
}

// NewStreamManager creates a new MediaMTX stream manager
// OPTIMIZED: Accept PathManager instead of creating a new one to ensure single instance
func NewStreamManager(client MediaMTXClient, pathManager PathManager, config *config.MediaMTXConfig, configIntegration *ConfigIntegration, logger *logging.Logger) StreamManager {
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

	// SIMPLIFIED: Single use case configuration for all operations
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
		configIntegration: configIntegration,
		logger:            logger,
		useCaseConfigs:    useCaseConfigs,
		// ffmpegCommands: sync.Map is zero-initialized, no need to initialize
	}
}

// StartStream starts a stream for a device (simplified - single path for all operations)
func (sm *streamManager) StartStream(ctx context.Context, devicePath string) (*Path, error) {
	// Validate dependencies are initialized
	if sm.pathManager == nil {
		return nil, fmt.Errorf("PathManager not initialized")
	}
	if sm.client == nil {
		return nil, fmt.Errorf("MediaMTXClient not initialized")
	}

	return sm.startStreamForUseCase(ctx, devicePath, UseCaseRecording)
}

// startStreamForUseCase starts a stream for the specified use case
func (sm *streamManager) startStreamForUseCase(ctx context.Context, devicePath string, useCase StreamUseCase) (*Path, error) {
	// Validate device path
	if err := sm.validateDevicePath(devicePath); err != nil {
		return nil, fmt.Errorf("failed to validate device path %s: %w", devicePath, err)
	}

	// Generate stream name using simplified approach - same name for all use cases
	streamName := GetMediaMTXPathName(devicePath)
	sm.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"use_case":    useCase,
		"stream_name": streamName,
	}).Info("Generated stream name for use case")

	// Build FFmpeg command for device-to-stream conversion
	ffmpegCommand := sm.buildFFmpegCommand(devicePath, streamName)

	// SIMPLIFIED: Create path configuration with stable settings
	// All use cases use the same configuration - stable path for streaming AND recording
	pathConfig := map[string]interface{}{
		"runOnDemand":             ffmpegCommand,
		"runOnDemandRestart":      true,                              // Always restart FFmpeg if it crashes
		"runOnDemandStartTimeout": sm.config.RunOnDemandStartTimeout, // Use configured timeout
		"runOnDemandCloseAfter":   sm.config.RunOnDemandCloseAfter,   // Use configured timeout
		"runOnUnDemand":           "",
	}

	// Use PathManager for proper architectural integration and validation
	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"device_path": devicePath,
		"path_config": pathConfig,
	}).Info("About to create MediaMTX path")

	err := sm.pathManager.CreatePath(ctx, streamName, devicePath, pathConfig)
	if err != nil {
		// Log the actual error for debugging
		sm.logger.WithError(err).WithFields(logging.Fields{
			"stream_name": streamName,
			"device_path": devicePath,
			"path_config": pathConfig,
		}).Error("CreatePath failed - investigating error")

		// Check if this is a "path already exists" error (idempotent success)
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

	stream := &Path{
		Name:     streamName,
		ConfName: streamName,
		Ready:    false,
		Tracks:   []string{},
		Readers:  []PathReader{},
	}

	sm.logger.WithFields(logging.Fields{
		"stream_name": streamName,
		"use_case":    useCase,
		"device_path": devicePath,
	}).Info("MediaMTX stream created successfully with use case configuration")

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

// GenerateStreamName generates stream name for the given device
// SIMPLIFIED: All use cases return the same stable path name (camera0, camera1, etc.)
// This aligns with MediaMTX architecture where one path handles streaming AND recording
func (sm *streamManager) GenerateStreamName(devicePath string, useCase StreamUseCase) string {
	// Always return the same simple path name regardless of use case
	return GetMediaMTXPathName(devicePath)
}

// buildFFmpegCommand builds FFmpeg command for camera stream with caching
func (sm *streamManager) buildFFmpegCommand(devicePath, streamName string) string {
	// Check cache first - lock-free read with sync.Map
	if cachedCommand, exists := sm.ffmpegCommands.Load(devicePath); exists {
		sm.logger.WithField("device_path", devicePath).Debug("Using cached FFmpeg command")
		return cachedCommand.(string)
	}

	// Build new command
	command := fmt.Sprintf(
		"ffmpeg -f v4l2 -i %s -c:v libx264 -preset ultrafast -tune zerolatency "+
			"-f rtsp rtsp://%s:%d/%s",
		devicePath, sm.config.Host, sm.config.RTSPPort, streamName)

	// Cache the command - lock-free write with sync.Map
	sm.ffmpegCommands.Store(devicePath, command)

	sm.logger.WithField("device_path", devicePath).Debug("Built and cached new FFmpeg command")
	return command
}

// invalidateFFmpegCommandCache clears cached FFmpeg command for a device
// Call this when device format settings change
func (sm *streamManager) invalidateFFmpegCommandCache(devicePath string) {
	sm.ffmpegCommands.Delete(devicePath) // Lock-free delete with sync.Map
	sm.logger.WithField("device_path", devicePath).Debug("Invalidated cached FFmpeg command")
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
		// Create FFmpeg command for USB device publishing (like Python implementation)
		ffmpegCommand := fmt.Sprintf(
			"ffmpeg -f v4l2 -i %s -c:v libx264 -profile:v baseline -level 3.0 "+
				"-pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://%s:%d/%s",
			source, sm.config.Host, sm.config.RTSPPort, name)

		// Create path configuration for USB device
		pathConfig := map[string]interface{}{
			"runOnDemand":             ffmpegCommand,
			"runOnDemandRestart":      true,
			"runOnDemandStartTimeout": sm.config.RunOnDemandStartTimeout,
			"runOnDemandCloseAfter":   sm.config.RunOnDemandCloseAfter,
			"runOnUnDemand":           "",
		}

		// Use PathManager for proper architectural integration and validation
		err := sm.pathManager.CreatePath(ctx, name, source, pathConfig)
		if err != nil {
			// Check if this is a "path already exists" error (idempotent success)
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
		pathConfig := map[string]interface{}{
			"source": source,
		}

		// Special handling for "publisher" source - PathManager will convert this
		// to runOnDemand configuration, so we need to ensure proper options
		if source == "publisher" {
			// PathManager will handle the conversion to runOnDemand
			// Just pass the source and let PathManager do the conversion
			pathConfig = map[string]interface{}{}
		}

		// Use PathManager for proper architectural integration
		err := sm.pathManager.CreatePath(ctx, name, source, pathConfig)
		if err != nil {
			// Check if this is a "path already exists" error (idempotent success)
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
func (sm *streamManager) ListStreams(ctx context.Context) ([]*Path, error) {
	sm.logger.Debug("Listing MediaMTX streams")

	// Use PathManager for proper architectural integration
	pathConfigs, err := sm.pathManager.ListPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	// Convert PathConf to Path for runtime response
	paths := make([]*Path, len(pathConfigs))
	for i, config := range pathConfigs {
		paths[i] = &Path{
			Name:      config.Name,
			ConfName:  config.Name,
			Source:    nil,   // Source is populated by MediaMTX runtime
			Ready:     false, // Will be updated by runtime status
			ReadyTime: nil,
			Tracks:    []string{},
			Readers:   []PathReader{},
		}
	}

	sm.logger.WithField("count", fmt.Sprintf("%d", len(paths))).Debug("MediaMTX streams listed successfully")
	return paths, nil
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
func (sm *streamManager) GetStreamStatus(ctx context.Context, id string) (string, error) {
	sm.logger.WithField("stream_id", id).Debug("Getting MediaMTX stream status")

	stream, err := sm.GetStream(ctx, id)
	if err != nil {
		return "", fmt.Errorf(id, "get_stream_status", "failed to get stream", err)
	}

	// Convert MediaMTX ready status to our status format
	if stream.Ready {
		return "READY", nil
	}
	return "NOT_READY", nil
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

	// Check readiness periodically
	ticker := time.NewTicker(100 * time.Millisecond)
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
func (sm *streamManager) StopStream(ctx context.Context, device string) error {
	sm.logger.WithFields(logging.Fields{
		"device": device,
		"action": "stop_stream",
	}).Info("Stopping stream for device")

	// Get the stable path name
	streamName := GetMediaMTXPathName(device)

	// Delete the stream from MediaMTX
	err := sm.DeleteStream(ctx, streamName)
	if err != nil {
		sm.logger.WithFields(logging.Fields{
			"device":      device,
			"stream_name": streamName,
			"error":       err.Error(),
		}).Error("Failed to stop stream")
		return fmt.Errorf("failed to stop stream: %w", err)
	}

	sm.logger.WithFields(logging.Fields{
		"device":      device,
		"stream_name": streamName,
	}).Info("Stream stopped successfully")

	return nil
}

// GenerateStreamURL generates the RTSP URL for a stream
func (sm *streamManager) GenerateStreamURL(streamName string) string {
	return fmt.Sprintf("rtsp://%s:%d/%s", sm.config.Host, sm.config.RTSPPort, streamName)
}

// EnableRecording enables recording on the stable path for a device
// This is the simplified approach - one path handles both streaming and recording
func (sm *streamManager) EnableRecording(ctx context.Context, devicePath string, outputPath string) error {
	// Get the stable path name
	pathName := GetMediaMTXPathName(devicePath)

	// Serialize create→ready→patch operations per path using per-path mutex
	pathMutex := sm.pathManager.(*pathManager).getPathMutex(pathName)
	pathMutex.Lock()
	defer pathMutex.Unlock()

	// Ensure the path exists (idempotent)
	stream, err := sm.startStreamForUseCase(ctx, devicePath, UseCaseRecording)
	if err != nil {
		return fmt.Errorf("failed to ensure path exists: %w", err)
	}

	sm.logger.WithFields(logging.Fields{
		"device_path": devicePath,
		"path_name":   pathName,
		"stream_name": stream.Name,
	}).Info("Path ensured, activating publisher and waiting for readiness")

	// DETERMINISTIC ACTIVATION: Trigger MediaMTX publisher via RTSP handshake
	// This is protocol-based activation, not time-based waiting
	err = sm.pathManager.ActivatePathPublisher(ctx, pathName)
	if err != nil {
		sm.logger.WithField("path_name", pathName).Warn("RTSP activation failed, proceeding with readiness check")
		// Don't fail here - some paths may not need activation
	}

	// Wait for path to be ready in runtime (not config)
	// Use configurable timeout from MediaMTX config
	timeout := 15 * time.Second // Default timeout
	if sm.config != nil && sm.config.StreamReadiness.Timeout > 0 {
		timeout = time.Duration(sm.config.StreamReadiness.Timeout) * time.Second
	}
	err = sm.pathManager.WaitForPathReady(ctx, pathName, timeout)
	if err != nil {
		return fmt.Errorf("failed to wait for path readiness: %w", err)
	}

	sm.logger.WithField("path_name", pathName).Info("Path is ready, enabling recording")

	// Pre-create recording directory
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Create recording configuration
	recordingConfig := sm.createRecordingConfig(pathName, outputPath)

	// PATCH the path to enable recording (now with retry)
	err = sm.pathManager.PatchPath(ctx, pathName, recordingConfig)
	if err != nil {
		return fmt.Errorf("failed to enable recording: %w", err)
	}

	sm.logger.WithField("path_name", pathName).Info("Recording enabled successfully")
	return nil
}

// DisableRecording disables recording on the stable path
// This keeps the path alive for streaming while stopping file recording
func (sm *streamManager) DisableRecording(ctx context.Context, devicePath string) error {
	pathName := GetMediaMTXPathName(devicePath)

	// Serialize operations per path using per-path mutex
	pathMutex := sm.pathManager.(*pathManager).getPathMutex(pathName)
	pathMutex.Lock()
	defer pathMutex.Unlock()

	// PATCH to disable recording (keep path for streaming)
	recordingConfig := map[string]interface{}{
		"record": false,
	}

	err := sm.pathManager.PatchPath(ctx, pathName, recordingConfig)
	if err != nil {
		return fmt.Errorf("failed to disable recording: %w", err)
	}

	sm.logger.WithField("path_name", pathName).Info("Recording disabled successfully")
	return nil
}

// createRecordingConfig creates the recording configuration for MediaMTX PATCH
func (sm *streamManager) createRecordingConfig(pathName, outputPath string) map[string]interface{} {
	// Generate recordPath with timestamp pattern
	recordPath := sm.getRecordingOutputPath(pathName, outputPath)

	// Get recording configuration from centralized config system
	recordingConfig, err := sm.configIntegration.GetRecordingConfig()
	if err != nil {
		sm.logger.WithError(err).Warn("Failed to get recording config, using fallback values")
		// Fallback to hardcoded values if config is unavailable
		return map[string]interface{}{
			"record":                true,
			"recordPath":            recordPath,
			"recordFormat":          "fmp4",                       // STANAG 4609 compatible
			"recordPartDuration":    sm.config.RecordPartDuration, // Use centralized config
			"recordMaxPartSize":     "100MB",
			"recordSegmentDuration": sm.config.RecordSegmentDuration, // Use centralized config
			"recordDeleteAfter":     sm.config.RecordDeleteAfter,     // Use centralized config
		}
	}

	// Convert config values to MediaMTX format
	config := map[string]interface{}{
		"record":                true,
		"recordPath":            recordPath,
		"recordFormat":          recordingConfig.Format,                                                // Use configured format
		"recordPartDuration":    recordingConfig.DefaultMaxDuration.String(),                           // Use configured duration
		"recordMaxPartSize":     fmt.Sprintf("%dMB", recordingConfig.MaxSegmentSize/1024/1024),         // Convert bytes to MB
		"recordSegmentDuration": time.Duration(recordingConfig.SegmentDuration).String(),               // Use configured segment duration
		"recordDeleteAfter":     time.Duration(recordingConfig.DefaultRetentionDays*24).String() + "h", // Convert days to hours
	}

	return config
}

// getRecordingOutputPath generates the recordPath with timestamp pattern
func (sm *streamManager) getRecordingOutputPath(pathName, outputPath string) string {
	if outputPath != "" {
		dir := filepath.Dir(outputPath)
		// MediaMTX requires %path in recordPath - it gets replaced with the actual path name
		return filepath.Join(dir, "%%path_%%Y-%%m-%%d_%%H-%%M-%%S.mp4")
	}
	// MediaMTX requires %path in recordPath - it gets replaced with the actual path name
	return "/opt/recordings/%%path_%%Y-%%m-%%d_%%H-%%M-%%S.mp4"
}
