/*
MediaMTX Recording Recovery Strategies Implementation

This package provides specific recovery strategies for recording operations.
It implements the RecoveryStrategy interface for various recording error scenarios.

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit/Integration
*/

package mediamtx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RecordingRecoveryStrategy handles recording-specific error recovery
type RecordingRecoveryStrategy struct {
	recordingManager *RecordingManager
	logger           *logging.Logger
}

// NewRecordingRecoveryStrategy creates a new recording recovery strategy
func NewRecordingRecoveryStrategy(recordingManager *RecordingManager, logger *logging.Logger) *RecordingRecoveryStrategy {
	return &RecordingRecoveryStrategy{
		recordingManager: recordingManager,
		logger:           logger,
	}
}

// CanRecover determines if this strategy can handle the error
func (rrs *RecordingRecoveryStrategy) CanRecover(ctx *ErrorContext, err error) bool {
	// Only handle recording-related errors
	if ctx.Component != "RecordingManager" {
		return false
	}

	errorMsg := strings.ToLower(err.Error())

	// Handle MediaMTX API errors
	if strings.Contains(errorMsg, "mediamtx") ||
		strings.Contains(errorMsg, "path not found") ||
		strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "404") ||
		strings.Contains(errorMsg, "409") {
		return true
	}

	// Handle keepalive errors
	if strings.Contains(errorMsg, "keepalive") ||
		strings.Contains(errorMsg, "rtsp") {
		return true
	}

	// Handle path creation errors
	if strings.Contains(errorMsg, "path") &&
		(strings.Contains(errorMsg, "create") || strings.Contains(errorMsg, "failed")) {
		return true
	}

	return false
}

// Recover attempts to recover from recording errors
func (rrs *RecordingRecoveryStrategy) Recover(ctx context.Context, errorCtx *ErrorContext, err error) error {
	errorMsg := strings.ToLower(err.Error())

	// Strategy 1: Path not found - try to recreate the path
	if strings.Contains(errorMsg, "path not found") || strings.Contains(errorMsg, "404") {
		return rrs.recoverPathNotFound(ctx, errorCtx, err)
	}

	// Strategy 2: Path already exists - this is usually not an error, but if it is, try to reset
	if strings.Contains(errorMsg, "already exists") || strings.Contains(errorMsg, "409") {
		return rrs.recoverPathConflict(ctx, errorCtx, err)
	}

	// Strategy 3: Keepalive/RTSP errors - restart keepalive
	if strings.Contains(errorMsg, "keepalive") || strings.Contains(errorMsg, "rtsp") {
		return rrs.recoverKeepaliveError(ctx, errorCtx, err)
	}

	// Strategy 4: Generic path errors - try to reset the path
	if strings.Contains(errorMsg, "path") && strings.Contains(errorMsg, "failed") {
		return rrs.recoverPathError(ctx, errorCtx, err)
	}

	// Default: log and return original error
	rrs.logger.WithFields(logging.Fields{
		"component": errorCtx.Component,
		"operation": errorCtx.Operation,
		"camera_id": errorCtx.CameraID,
		"error":     err.Error(),
	}).Debug("No specific recovery strategy for error type")

	return err
}

// GetRecoveryDelay returns the delay before attempting recovery
func (rrs *RecordingRecoveryStrategy) GetRecoveryDelay() time.Duration {
	return 2 * time.Second // Wait 2 seconds before attempting recovery
}

// GetStrategyName returns the name of this recovery strategy
func (rrs *RecordingRecoveryStrategy) GetStrategyName() string {
	return "RecordingRecovery"
}

// recoverPathNotFound handles path not found errors
func (rrs *RecordingRecoveryStrategy) recoverPathNotFound(ctx context.Context, errorCtx *ErrorContext, err error) error {
	rrs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
		"path_name": errorCtx.PathName,
	}).Info("Attempting to recover from path not found error")

	// Try to recreate the path
	if errorCtx.CameraID != "" {
		// Get device path
		devicePath, exists := rrs.recordingManager.pathManager.GetDevicePathForCamera(errorCtx.CameraID)
		if !exists {
			return fmt.Errorf("device path not found for camera %s: %w", errorCtx.CameraID, err)
		}

		// Build path configuration
		// Use resolver-based path configuration for recovery
		pathOptions, buildErr := rrs.recordingManager.configIntegration.BuildRecordingPathConf(devicePath, errorCtx.CameraID)
		if buildErr != nil {
			return fmt.Errorf("failed to build path configuration: %w", buildErr)
		}

		// Create the path
		createErr := rrs.recordingManager.pathManager.CreatePath(ctx, errorCtx.CameraID, devicePath, pathOptions)
		if createErr != nil {
			return fmt.Errorf("failed to recreate path: %w", createErr)
		}

		rrs.logger.WithFields(logging.Fields{
			"camera_id": errorCtx.CameraID,
			"path_name": errorCtx.PathName,
		}).Info("Successfully recreated path after not found error")
	} else {
		return fmt.Errorf("device path not found: no camera ID provided")
	}

	return nil
}

// recoverPathConflict handles path conflict errors
func (rrs *RecordingRecoveryStrategy) recoverPathConflict(ctx context.Context, errorCtx *ErrorContext, err error) error {
	rrs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
		"path_name": errorCtx.PathName,
	}).Info("Attempting to recover from path conflict error")

	// For path conflicts, we usually just need to continue with the existing path
	// This is often not a real error in MediaMTX
	rrs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
		"path_name": errorCtx.PathName,
	}).Debug("Path conflict resolved - using existing path")

	return nil
}

// recoverKeepaliveError handles keepalive/RTSP errors
func (rrs *RecordingRecoveryStrategy) recoverKeepaliveError(ctx context.Context, errorCtx *ErrorContext, err error) error {
	rrs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
	}).Info("Attempting to recover from keepalive/RTSP error")

	// Stop existing keepalive
	if errorCtx.CameraID != "" {
		rrs.recordingManager.stopRTSPKeepalive(errorCtx.CameraID)

		// Wait a moment
		time.Sleep(1 * time.Second)

		// Restart keepalive
		restartErr := rrs.recordingManager.startRTSPKeepalive(ctx, errorCtx.CameraID)
		if restartErr != nil {
			return fmt.Errorf("failed to restart keepalive: %w", restartErr)
		}

		rrs.logger.WithFields(logging.Fields{
			"camera_id": errorCtx.CameraID,
		}).Info("Successfully restarted keepalive after error")
	}

	return nil
}

// recoverPathError handles generic path errors
func (rrs *RecordingRecoveryStrategy) recoverPathError(ctx context.Context, errorCtx *ErrorContext, err error) error {
	rrs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
		"path_name": errorCtx.PathName,
	}).Info("Attempting to recover from generic path error")

	// Try to reset the path by deleting and recreating it
	if errorCtx.CameraID != "" {
		// Delete the path first
		deleteErr := rrs.recordingManager.pathManager.DeletePath(ctx, errorCtx.CameraID)
		if deleteErr != nil {
			rrs.logger.WithError(deleteErr).Debug("Failed to delete path during recovery (this may be expected)")
		}

		// Wait a moment
		time.Sleep(1 * time.Second)

		// Recreate the path
		return rrs.recoverPathNotFound(ctx, errorCtx, err)
	}

	return err
}

// StreamRecoveryStrategy handles stream-specific error recovery
type StreamRecoveryStrategy struct {
	streamManager *streamManager
	logger        *logging.Logger
}

// NewStreamRecoveryStrategy creates a new stream recovery strategy
func NewStreamRecoveryStrategy(streamManager *streamManager, logger *logging.Logger) *StreamRecoveryStrategy {
	return &StreamRecoveryStrategy{
		streamManager: streamManager,
		logger:        logger,
	}
}

// CanRecover determines if this strategy can handle the error
func (srs *StreamRecoveryStrategy) CanRecover(ctx *ErrorContext, err error) bool {
	// Only handle stream-related errors
	if ctx.Component != "StreamManager" {
		return false
	}

	errorMsg := strings.ToLower(err.Error())

	// Handle FFmpeg errors
	if strings.Contains(errorMsg, "ffmpeg") ||
		strings.Contains(errorMsg, "process") ||
		strings.Contains(errorMsg, "command") {
		return true
	}

	// Handle stream creation errors
	if strings.Contains(errorMsg, "stream") &&
		(strings.Contains(errorMsg, "create") || strings.Contains(errorMsg, "failed")) {
		return true
	}

	return false
}

// Recover attempts to recover from stream errors
func (srs *StreamRecoveryStrategy) Recover(ctx context.Context, errorCtx *ErrorContext, err error) error {
	errorMsg := strings.ToLower(err.Error())

	// Strategy 1: FFmpeg errors - try to restart the stream
	if strings.Contains(errorMsg, "ffmpeg") || strings.Contains(errorMsg, "process") {
		return srs.recoverFFmpegError(ctx, errorCtx, err)
	}

	// Strategy 2: Stream creation errors - try to reset and recreate
	if strings.Contains(errorMsg, "stream") && strings.Contains(errorMsg, "create") {
		return srs.recoverStreamCreationError(ctx, errorCtx, err)
	}

	return err
}

// GetRecoveryDelay returns the delay before attempting recovery
func (srs *StreamRecoveryStrategy) GetRecoveryDelay() time.Duration {
	return 3 * time.Second // Wait 3 seconds before attempting recovery
}

// GetStrategyName returns the name of this recovery strategy
func (srs *StreamRecoveryStrategy) GetStrategyName() string {
	return "StreamRecovery"
}

// recoverFFmpegError handles FFmpeg process errors
func (srs *StreamRecoveryStrategy) recoverFFmpegError(ctx context.Context, errorCtx *ErrorContext, err error) error {
	srs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
	}).Info("Attempting to recover from FFmpeg error")

	// For FFmpeg errors, we typically need to restart the stream
	// This will trigger a new FFmpeg process
	if errorCtx.CameraID != "" {
		// Stop the stream first
		stopErr := srs.streamManager.StopStream(ctx, errorCtx.CameraID)
		if stopErr != nil {
			srs.logger.WithError(stopErr).Debug("Failed to stop stream during recovery (this may be expected)")
		}

		// Wait for cleanup
		time.Sleep(2 * time.Second)

		// Restart the stream
		_, restartErr := srs.streamManager.StartStream(ctx, errorCtx.CameraID)
		if restartErr != nil {
			return fmt.Errorf("failed to restart stream: %w", restartErr)
		}

		srs.logger.WithFields(logging.Fields{
			"camera_id": errorCtx.CameraID,
		}).Info("Successfully restarted stream after FFmpeg error")
	}

	return nil
}

// recoverStreamCreationError handles stream creation errors
func (srs *StreamRecoveryStrategy) recoverStreamCreationError(ctx context.Context, errorCtx *ErrorContext, err error) error {
	srs.logger.WithFields(logging.Fields{
		"camera_id": errorCtx.CameraID,
	}).Info("Attempting to recover from stream creation error")

	// For stream creation errors, try to clean up and retry
	if errorCtx.CameraID != "" {
		// Clean up any existing stream
		srs.streamManager.StopStream(ctx, errorCtx.CameraID)

		// Wait for cleanup
		time.Sleep(2 * time.Second)

		// Retry stream creation
		_, retryErr := srs.streamManager.StartStream(ctx, errorCtx.CameraID)
		if retryErr != nil {
			return fmt.Errorf("failed to retry stream creation: %w", retryErr)
		}

		srs.logger.WithFields(logging.Fields{
			"camera_id": errorCtx.CameraID,
		}).Info("Successfully recovered from stream creation error")
	}

	return nil
}
