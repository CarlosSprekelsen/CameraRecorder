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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// streamManager represents the MediaMTX stream manager
type streamManager struct {
	client         MediaMTXClient
	config         *MediaMTXConfig
	logger         *logging.Logger
	useCaseConfigs map[StreamUseCase]UseCaseConfig
}

// NewStreamManager creates a new MediaMTX stream manager
func NewStreamManager(client MediaMTXClient, config *MediaMTXConfig, logger *logging.Logger) StreamManager {
	// Initialize use case configurations based on Python implementation
	useCaseConfigs := map[StreamUseCase]UseCaseConfig{
		UseCaseRecording: {
			RunOnDemandCloseAfter:   "0s", // Never auto-close for recording
			RunOnDemandRestart:      true,
			RunOnDemandStartTimeout: "10s",
			Suffix:                  "",
		},
		UseCaseViewing: {
			RunOnDemandCloseAfter:   "300s", // 5 minutes after last viewer
			RunOnDemandRestart:      true,
			RunOnDemandStartTimeout: "10s",
			Suffix:                  "_viewing",
		},
		UseCaseSnapshot: {
			RunOnDemandCloseAfter:   "60s", // 1 minute after capture
			RunOnDemandRestart:      false,
			RunOnDemandStartTimeout: "5s",
			Suffix:                  "_snapshot",
		},
	}

	return &streamManager{
		client:         client,
		config:         config,
		logger:         logger,
		useCaseConfigs: useCaseConfigs,
	}
}

// StartRecordingStream starts a stream optimized for recording with file rotation
func (sm *streamManager) StartRecordingStream(ctx context.Context, devicePath string) (*Stream, error) {
	return sm.startStreamForUseCase(ctx, devicePath, UseCaseRecording)
}

// StartViewingStream starts a stream optimized for live viewing
func (sm *streamManager) StartViewingStream(ctx context.Context, devicePath string) (*Stream, error) {
	return sm.startStreamForUseCase(ctx, devicePath, UseCaseViewing)
}

// StartSnapshotStream starts a stream optimized for quick snapshot capture
func (sm *streamManager) StartSnapshotStream(ctx context.Context, devicePath string) (*Stream, error) {
	return sm.startStreamForUseCase(ctx, devicePath, UseCaseSnapshot)
}

// startStreamForUseCase starts a stream for the specified use case
func (sm *streamManager) startStreamForUseCase(ctx context.Context, devicePath string, useCase StreamUseCase) (*Stream, error) {
	// Validate device path
	if err := sm.validateDevicePath(devicePath); err != nil {
		return nil, fmt.Errorf("failed to validate device path %s: %w", devicePath, err)
	}

	// Generate stream name with use case suffix
	streamName := sm.GenerateStreamName(devicePath, useCase)

	// Get use case configuration
	useCaseConfig, exists := sm.useCaseConfigs[useCase]
	if !exists {
		return nil, fmt.Errorf("unsupported use case: %s", useCase)
	}

	// Build FFmpeg command for device-to-stream conversion
	ffmpegCommand := sm.buildFFmpegCommand(devicePath, streamName)

	// Create path configuration with use case specific settings
	pathConfig := map[string]interface{}{
		"runOnDemand":             ffmpegCommand,
		"runOnDemandRestart":      useCaseConfig.RunOnDemandRestart,
		"runOnDemandStartTimeout": useCaseConfig.RunOnDemandStartTimeout,
		"runOnDemandCloseAfter":   useCaseConfig.RunOnDemandCloseAfter,
		"runOnUnDemand":           "",
	}

	// Marshal request
	data, err := json.Marshal(pathConfig)
	if err != nil {
		return nil, NewStreamErrorWithErr(streamName, "create_stream", "failed to marshal request", err)
	}

	// Send request to MediaMTX path API
	responseData, err := sm.client.Post(ctx, fmt.Sprintf("/v3/config/paths/add/%s", streamName), data)
	if err != nil {
		// Check if this is a "path already exists" error (idempotent success)
		if strings.Contains(err.Error(), "path already exists") || strings.Contains(err.Error(), "already exists") {
			sm.logger.WithField("stream_name", streamName).Info("MediaMTX path already exists, treating as success")
			// Return a mock stream response for idempotent success
			stream := &Stream{
				Name:     streamName,
				URL:      sm.GenerateStreamURL(streamName),
				ConfName: streamName,
				Ready:    false,
				Tracks:   []string{},
				Readers:  []PathReader{},
			}
			return stream, nil
		}
		return nil, NewStreamErrorWithErr(streamName, "create_stream", "failed to create stream", err)
	}

	// Handle MediaMTX API response: successful path creation returns empty body
	if len(responseData) == 0 {
		// Empty response means successful path creation (validated against actual MediaMTX API)
		stream := &Stream{
			Name:     streamName,
			URL:      sm.GenerateStreamURL(streamName),
			ConfName: streamName,
			Ready:    false,
			Tracks:   []string{},
			Readers:  []PathReader{},
		}
		sm.logger.WithFields(map[string]interface{}{
			"stream_name": streamName,
			"use_case":    useCase,
			"device_path": devicePath,
		}).Info("MediaMTX stream created successfully with use case configuration")
		return stream, nil
	}

	// If response has content, try to parse it (for error cases or future API changes)
	stream, err := parseStreamResponse(responseData)
	if err != nil {
		return nil, NewStreamErrorWithErr(streamName, "create_stream", "failed to parse stream response", err)
	}

	sm.logger.WithFields(map[string]interface{}{
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

	// Validate device path format (matches Python implementation)
	if !strings.HasPrefix(devicePath, "/dev/video") && !strings.HasPrefix(devicePath, "/dev/custom") {
		return fmt.Errorf("invalid device path format: %s. Must be /dev/video<N> or /dev/custom<name>", devicePath)
	}

	return nil
}

// GenerateStreamName generates stream name for the given device and use case
func (sm *streamManager) GenerateStreamName(devicePath string, useCase StreamUseCase) string {
	// Extract device number from path (matches Python implementation)
	parts := strings.Split(devicePath, "/")
	deviceName := parts[len(parts)-1]

	var baseName string
	if strings.HasPrefix(deviceName, "video") {
		deviceNum := deviceName[5:] // Remove "video" prefix
		baseName = fmt.Sprintf("camera%s", deviceNum)
	} else {
		baseName = deviceName
	}

	// Add use case suffix
	useCaseConfig := sm.useCaseConfigs[useCase]
	streamName := baseName + useCaseConfig.Suffix

	return streamName
}

// buildFFmpegCommand builds FFmpeg command for camera stream
func (sm *streamManager) buildFFmpegCommand(devicePath, streamName string) string {
	return fmt.Sprintf(
		"ffmpeg -f v4l2 -i %s -c:v libx264 -preset ultrafast -tune zerolatency "+
			"-f rtsp rtsp://%s:%d/%s",
		devicePath, sm.config.Host, sm.config.RTSPPort, streamName)
}

// CreateStream creates a new stream (legacy method for backward compatibility)
func (sm *streamManager) CreateStream(ctx context.Context, name, source string) (*Stream, error) {
	sm.logger.WithFields(map[string]interface{}{
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

		// Use the new function that matches Python implementation
		data, err := marshalCreateUSBPathRequest(name, ffmpegCommand)
		if err != nil {
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to marshal request", err)
		}

		// Send request to MediaMTX path API
		responseData, err := sm.client.Post(ctx, fmt.Sprintf("/v3/config/paths/add/%s", name), data)
		if err != nil {
			// Check if this is a "path already exists" error (idempotent success)
			if strings.Contains(err.Error(), "path already exists") || strings.Contains(err.Error(), "already exists") {
				sm.logger.WithField("stream_name", name).Info("MediaMTX path already exists, treating as success")
				// Return a mock stream response for idempotent success
				stream := &Stream{
					Name:     name,
					URL:      sm.GenerateStreamURL(name),
					ConfName: name,
					Ready:    false,
					Tracks:   []string{},
					Readers:  []PathReader{},
				}
				return stream, nil
			}
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to create stream", err)
		}

		// Handle MediaMTX API response: successful path creation returns empty body
		if len(responseData) == 0 {
			// Empty response means successful path creation (validated against actual MediaMTX API)
			stream := &Stream{
				Name:     name,
				URL:      sm.GenerateStreamURL(name),
				ConfName: name,
				Ready:    false,
				Tracks:   []string{},
				Readers:  []PathReader{},
			}
			sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully with FFmpeg publishing")
			return stream, nil
		}

		// If response has content, try to parse it (for error cases or future API changes)
		stream, err := parseStreamResponse(responseData)
		if err != nil {
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to parse stream response", err)
		}

		sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully with FFmpeg publishing")
		return stream, nil
	} else {
		// For non-USB sources (RTSP URLs, etc.), use direct source
		data, err := marshalCreateStreamRequest(name, source)
		if err != nil {
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to marshal request", err)
		}

		// Send request - MediaMTX uses paths, not streams
		responseData, err := sm.client.Post(ctx, fmt.Sprintf("/v3/config/paths/add/%s", name), data)
		if err != nil {
			// Check if this is a "path already exists" error (idempotent success)
			if strings.Contains(err.Error(), "path already exists") || strings.Contains(err.Error(), "already exists") {
				sm.logger.WithField("stream_name", name).Info("MediaMTX path already exists, treating as success")
				// Return a mock stream response for idempotent success
				stream := &Stream{
					Name:     name,
					URL:      sm.GenerateStreamURL(name),
					ConfName: name,
					Ready:    false,
					Tracks:   []string{},
					Readers:  []PathReader{},
				}
				return stream, nil
			}
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to create stream", err)
		}

		// Handle MediaMTX API response: successful path creation returns empty body
		if len(responseData) == 0 {
			// Empty response means successful path creation (validated against actual MediaMTX API)
			stream := &Stream{
				Name:     name,
				URL:      sm.GenerateStreamURL(name),
				ConfName: name,
				Ready:    false,
				Tracks:   []string{},
				Readers:  []PathReader{},
			}
			sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully")
			return stream, nil
		}

		// If response has content, try to parse it (for error cases or future API changes)
		stream, err := parseStreamResponse(responseData)
		if err != nil {
			return nil, NewStreamErrorWithErr(name, "create_stream", "failed to parse stream response", err)
		}

		sm.logger.WithField("stream_name", stream.Name).Info("MediaMTX stream created successfully")
		return stream, nil
	}
}

// CreateStreamWithUseCase creates a new stream with use case specific configuration
func (sm *streamManager) CreateStreamWithUseCase(ctx context.Context, name, source string, useCase StreamUseCase) (*Stream, error) {
	sm.logger.WithFields(map[string]interface{}{
		"name":     name,
		"source":   source,
		"use_case": useCase,
	}).Debug("Creating MediaMTX stream with use case configuration")

	// Get use case configuration
	useCaseConfig, exists := sm.useCaseConfigs[useCase]
	if !exists {
		return nil, fmt.Errorf("unsupported use case: %s", useCase)
	}

	// Add use case suffix to stream name if specified
	streamName := name
	if useCaseConfig.Suffix != "" {
		streamName = name + useCaseConfig.Suffix
	}

	// Create path configuration with use case specific settings
	// This would be used to configure MediaMTX paths with specific lifecycle policies
	_ = map[string]interface{}{
		"runOnDemandCloseAfter":   useCaseConfig.RunOnDemandCloseAfter,
		"runOnDemandRestart":      useCaseConfig.RunOnDemandRestart,
		"runOnDemandStartTimeout": useCaseConfig.RunOnDemandStartTimeout,
	}

	// Create the stream with use case specific configuration
	// This would typically involve creating a MediaMTX path with the specific configuration
	// For now, we'll use the basic CreateStream method but log the use case configuration
	sm.logger.WithFields(map[string]interface{}{
		"stream_name": streamName,
		"use_case":    useCase,
		"config":      useCaseConfig,
	}).Info("Creating stream with use case specific configuration")

	// Use the existing CreateStream method for now
	// In a full implementation, this would create a MediaMTX path with the specific configuration
	return sm.CreateStream(ctx, streamName, source)
}

// DeleteStream deletes a stream
func (sm *streamManager) DeleteStream(ctx context.Context, id string) error {
	sm.logger.WithField("stream_id", id).Debug("Deleting MediaMTX stream")

	err := sm.client.Delete(ctx, fmt.Sprintf("/v3/config/paths/delete/%s", id))
	if err != nil {
		return NewStreamErrorWithErr(id, "delete_stream", "failed to delete stream", err)
	}

	sm.logger.WithField("stream_id", id).Info("MediaMTX stream deleted successfully")
	return nil
}

// GetStream gets a specific stream
func (sm *streamManager) GetStream(ctx context.Context, id string) (*Stream, error) {
	sm.logger.WithField("stream_id", id).Debug("Getting MediaMTX stream")

	data, err := sm.client.Get(ctx, fmt.Sprintf("/v3/paths/get/%s", id))
	if err != nil {
		return nil, NewStreamErrorWithErr(id, "get_stream", "failed to get stream", err)
	}

	stream, err := parseStreamResponse(data)
	if err != nil {
		return nil, NewStreamErrorWithErr(id, "get_stream", "failed to parse stream response", err)
	}

	return stream, nil
}

// ListStreams lists all streams
func (sm *streamManager) ListStreams(ctx context.Context) ([]*Stream, error) {
	sm.logger.Debug("Listing MediaMTX streams")

	data, err := sm.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	streams, err := parseStreamsResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse streams response: %w", err)
	}

	sm.logger.WithField("count", fmt.Sprintf("%d", len(streams))).Debug("MediaMTX streams listed successfully")
	return streams, nil
}

// MonitorStream monitors a stream
func (sm *streamManager) MonitorStream(ctx context.Context, id string) error {
	sm.logger.WithField("stream_id", id).Debug("Monitoring MediaMTX stream")

	// Get stream status
	status, err := sm.GetStreamStatus(ctx, id)
	if err != nil {
		return NewStreamErrorWithErr(id, "monitor_stream", "failed to get stream status", err)
	}

	sm.logger.WithFields(map[string]interface{}{
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
		return "", NewStreamErrorWithErr(id, "get_stream_status", "failed to get stream", err)
	}

	// Convert MediaMTX ready status to our status format
	if stream.Ready {
		return "READY", nil
	}
	return "NOT_READY", nil
}

// CheckStreamReadiness checks if a stream is ready for operations (enhanced existing stream manager)
func (sm *streamManager) CheckStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error) {
	sm.logger.WithFields(map[string]interface{}{
		"stream_name": streamName,
		"timeout":     timeout,
	}).Debug("Checking stream readiness")

	// Get current stream status from MediaMTX
	data, err := sm.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return false, fmt.Errorf("failed to get MediaMTX active paths: %w", err)
	}

	// Parse paths response
	var pathsResponse struct {
		Items []struct {
			Name  string `json:"name"`
			Ready bool   `json:"ready"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &pathsResponse); err != nil {
		return false, fmt.Errorf("failed to parse paths response: %w", err)
	}

	// Find the specific stream
	for _, path := range pathsResponse.Items {
		if path.Name == streamName {
			if path.Ready {
				sm.logger.WithField("stream_name", streamName).Debug("Stream is ready")
				return true, nil
			}
			sm.logger.WithField("stream_name", streamName).Debug("Stream is not ready")
			return false, nil
		}
	}

	return false, fmt.Errorf("stream %s not found", streamName)
}

// WaitForStreamReadiness waits for a stream to become ready (enhanced existing stream manager)
func (sm *streamManager) WaitForStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error) {
	sm.logger.WithFields(map[string]interface{}{
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

// StopViewingStream stops a viewing stream for the specified device
func (sm *streamManager) StopViewingStream(ctx context.Context, device string) error {
	sm.logger.WithFields(map[string]interface{}{
		"device": device,
		"action": "stop_viewing_stream",
	}).Info("Stopping viewing stream")

	// Generate stream name for viewing use case
	streamName := sm.GenerateStreamName(device, UseCaseViewing)

	// Delete the stream from MediaMTX
	err := sm.DeleteStream(ctx, streamName)
	if err != nil {
		sm.logger.WithFields(map[string]interface{}{
			"device":      device,
			"stream_name": streamName,
			"error":       err.Error(),
		}).Error("Failed to stop viewing stream")
		return fmt.Errorf("failed to stop viewing stream: %w", err)
	}

	sm.logger.WithFields(map[string]interface{}{
		"device":      device,
		"stream_name": streamName,
	}).Info("Viewing stream stopped successfully")

	return nil
}

// StopStreaming stops any active stream for the specified device
func (sm *streamManager) StopStreaming(ctx context.Context, device string) error {
	sm.logger.WithFields(map[string]interface{}{
		"device": device,
		"action": "stop_streaming",
	}).Info("Stopping any active stream")

	// Try to stop viewing stream first
	if err := sm.StopViewingStream(ctx, device); err == nil {
		return nil
	}

	// If viewing stream doesn't exist, try to stop recording stream
	streamName := sm.GenerateStreamName(device, UseCaseRecording)
	err := sm.DeleteStream(ctx, streamName)
	if err != nil {
		sm.logger.WithFields(map[string]interface{}{
			"device":      device,
			"stream_name": streamName,
			"error":       err.Error(),
		}).Error("Failed to stop recording stream")
		return fmt.Errorf("failed to stop recording stream: %w", err)
	}

	sm.logger.WithFields(map[string]interface{}{
		"device":      device,
		"stream_name": streamName,
	}).Info("Recording stream stopped successfully")

	return nil
}

// GenerateStreamURL generates the RTSP URL for a stream
func (sm *streamManager) GenerateStreamURL(streamName string) string {
	return fmt.Sprintf("rtsp://%s:%d/%s", sm.config.Host, sm.config.RTSPPort, streamName)
}
