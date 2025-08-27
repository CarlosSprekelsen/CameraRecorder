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
	"time"

	"github.com/sirupsen/logrus"
)

// streamManager represents the MediaMTX stream manager
type streamManager struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logrus.Logger
}

// NewStreamManager creates a new MediaMTX stream manager
func NewStreamManager(client MediaMTXClient, config *MediaMTXConfig, logger *logrus.Logger) StreamManager {
	return &streamManager{
		client: client,
		config: config,
		logger: logger,
	}
}

// CreateStream creates a new stream
func (sm *streamManager) CreateStream(ctx context.Context, name, source string) (*Stream, error) {
	sm.logger.WithFields(logrus.Fields{
		"name":   name,
		"source": source,
	}).Debug("Creating MediaMTX stream")

	// Marshal request
	data, err := marshalCreateStreamRequest(name, source)
	if err != nil {
		return nil, NewStreamErrorWithErr(name, "create_stream", "failed to marshal request", err)
	}

	// Send request
	responseData, err := sm.client.Post(ctx, "/v3/streams/add", data)
	if err != nil {
		return nil, NewStreamErrorWithErr(name, "create_stream", "failed to create stream", err)
	}

	// Parse response
	stream, err := parseStreamResponse(responseData)
	if err != nil {
		return nil, NewStreamErrorWithErr(name, "create_stream", "failed to parse stream response", err)
	}

	sm.logger.WithField("stream_id", stream.ID).Info("MediaMTX stream created successfully")
	return stream, nil
}

// DeleteStream deletes a stream
func (sm *streamManager) DeleteStream(ctx context.Context, id string) error {
	sm.logger.WithField("stream_id", id).Debug("Deleting MediaMTX stream")

	err := sm.client.Delete(ctx, fmt.Sprintf("/v3/streams/delete/%s", id))
	if err != nil {
		return NewStreamErrorWithErr(id, "delete_stream", "failed to delete stream", err)
	}

	sm.logger.WithField("stream_id", id).Info("MediaMTX stream deleted successfully")
	return nil
}

// GetStream gets a specific stream
func (sm *streamManager) GetStream(ctx context.Context, id string) (*Stream, error) {
	sm.logger.WithField("stream_id", id).Debug("Getting MediaMTX stream")

	data, err := sm.client.Get(ctx, fmt.Sprintf("/v3/streams/get/%s", id))
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

	data, err := sm.client.Get(ctx, "/v3/streams/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	streams, err := parseStreamsResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse streams response: %w", err)
	}

	sm.logger.WithField("count", len(streams)).Debug("MediaMTX streams listed successfully")
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

	sm.logger.WithFields(logrus.Fields{
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

	return stream.Status, nil
}

// CheckStreamReadiness checks if a stream is ready for operations (enhanced existing stream manager)
func (sm *streamManager) CheckStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error) {
	sm.logger.WithFields(logrus.Fields{
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
	sm.logger.WithFields(logrus.Fields{
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
