/*
MediaMTX Recording Timer Implementation

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
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RecordingTimer represents an enhanced timer with metadata for accurate duration tracking.
//
// RESPONSIBILITIES:
// - Enhanced timer structure with start time and metadata tracking
// - Accurate duration calculation for recording operations
// - Device and camera mapping for proper abstraction
// - Metadata storage for analytics and reporting
//
// ARCHITECTURE:
// - Replaces simple *time.Timer in sync.Map with structured timer
// - Maintains compatibility with existing timer operations
// - Provides accurate start time tracking for duration calculations
// - Supports additional metadata for enhanced analytics
type RecordingTimer struct {
	Timer     *time.Timer            `json:"-"`          // Actual timer (not serialized)
	StartTime time.Time              `json:"start_time"` // Recording start time
	Duration  time.Duration          `json:"duration"`   // Configured duration (0 for indefinite)
	CameraID  string                 `json:"camera_id"`  // Camera identifier (camera0, camera1, etc.)
	Device    string                 `json:"device"`     // Device path (/dev/video0, etc.)
	AutoStop  bool                   `json:"auto_stop"`  // Whether timer will auto-stop recording
	Metadata  map[string]interface{} `json:"metadata"`   // Additional metadata
	CreatedAt time.Time              `json:"created_at"` // When timer was created
}

// RecordingTimerManager manages recording timers with enhanced metadata tracking.
//
// RESPONSIBILITIES:
// - Thread-safe timer management using sync.Map
// - Timer creation with metadata tracking
// - Duration calculation and timer cleanup
// - Integration with existing RecordingManager patterns
type RecordingTimerManager struct {
	timers sync.Map // cameraID -> *RecordingTimer
	logger *logging.Logger
}

// NewRecordingTimerManager creates a new recording timer manager
func NewRecordingTimerManager(logger *logging.Logger) *RecordingTimerManager {
	return &RecordingTimerManager{
		logger: logger,
	}
}

// CreateTimer creates a new recording timer with metadata
func (rtm *RecordingTimerManager) CreateTimer(cameraID, device string, duration time.Duration, callback func()) *RecordingTimer {
	now := time.Now()

	recordingTimer := &RecordingTimer{
		StartTime: now,
		Duration:  duration,
		CameraID:  cameraID,
		Device:    device,
		AutoStop:  duration > 0, // Auto-stop if duration is specified
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
	}

	// Create actual timer if duration is specified
	if duration > 0 {
		recordingTimer.Timer = time.AfterFunc(duration, func() {
			rtm.logger.WithFields(logging.Fields{
				"camera_id": cameraID,
				"duration":  duration,
			}).Info("Recording timer expired, executing callback")

			// Execute callback
			callback()

			// Clean up timer
			rtm.DeleteTimer(cameraID)
		})
	}

	// Store timer
	rtm.timers.Store(cameraID, recordingTimer)

	rtm.logger.WithFields(logging.Fields{
		"camera_id": cameraID,
		"device":    device,
		"duration":  duration,
		"auto_stop": recordingTimer.AutoStop,
	}).Debug("Recording timer created")

	return recordingTimer
}

// GetTimer retrieves a recording timer by camera ID
func (rtm *RecordingTimerManager) GetTimer(cameraID string) (*RecordingTimer, bool) {
	if timer, exists := rtm.timers.Load(cameraID); exists {
		return timer.(*RecordingTimer), true
	}
	return nil, false
}

// DeleteTimer removes and stops a recording timer
func (rtm *RecordingTimerManager) DeleteTimer(cameraID string) bool {
	if timerInterface, exists := rtm.timers.LoadAndDelete(cameraID); exists {
		timer := timerInterface.(*RecordingTimer)

		// Stop the actual timer if it exists
		if timer.Timer != nil {
			timer.Timer.Stop()
		}

		rtm.logger.WithField("camera_id", cameraID).Debug("Recording timer deleted")
		return true
	}
	return false
}

// IsRecording checks if a camera is currently recording
func (rtm *RecordingTimerManager) IsRecording(cameraID string) bool {
	_, exists := rtm.timers.Load(cameraID)
	return exists
}

// GetRecordingDuration calculates the current recording duration
func (rtm *RecordingTimerManager) GetRecordingDuration(cameraID string) (time.Duration, bool) {
	if timer, exists := rtm.GetTimer(cameraID); exists {
		return time.Since(timer.StartTime), true
	}
	return 0, false
}

// GetRecordingInfo returns recording information for a camera
func (rtm *RecordingTimerManager) GetRecordingInfo(cameraID string) (*RecordingInfo, bool) {
	if timer, exists := rtm.GetTimer(cameraID); exists {
		duration := time.Since(timer.StartTime)

		return &RecordingInfo{
			CameraID:           cameraID,
			Device:             timer.Device,
			StartTime:          timer.StartTime,
			Duration:           duration,
			ConfiguredDuration: timer.Duration,
			AutoStop:           timer.AutoStop,
			IsActive:           true,
			Metadata:           timer.Metadata,
		}, true
	}
	return nil, false
}

// ListActiveRecordings returns all active recording timers
func (rtm *RecordingTimerManager) ListActiveRecordings() []*RecordingInfo {
	var recordings []*RecordingInfo

	rtm.timers.Range(func(key, value interface{}) bool {
		cameraID := key.(string)
		timer := value.(*RecordingTimer)

		duration := time.Since(timer.StartTime)

		recordings = append(recordings, &RecordingInfo{
			CameraID:           cameraID,
			Device:             timer.Device,
			StartTime:          timer.StartTime,
			Duration:           duration,
			ConfiguredDuration: timer.Duration,
			AutoStop:           timer.AutoStop,
			IsActive:           true,
			Metadata:           timer.Metadata,
		})

		return true // Continue iteration
	})

	return recordings
}

// UpdateTimerMetadata updates metadata for an existing timer
func (rtm *RecordingTimerManager) UpdateTimerMetadata(cameraID string, metadata map[string]interface{}) bool {
	if timer, exists := rtm.GetTimer(cameraID); exists {
		if timer.Metadata == nil {
			timer.Metadata = make(map[string]interface{})
		}

		// Update metadata
		for key, value := range metadata {
			timer.Metadata[key] = value
		}

		rtm.logger.WithFields(logging.Fields{
			"camera_id": cameraID,
			"metadata":  metadata,
		}).Debug("Recording timer metadata updated")

		return true
	}
	return false
}

// RecordingInfo represents information about an active recording
type RecordingInfo struct {
	CameraID           string                 `json:"camera_id"`
	Device             string                 `json:"device"`
	StartTime          time.Time              `json:"start_time"`
	Duration           time.Duration          `json:"duration"`
	ConfiguredDuration time.Duration          `json:"configured_duration"`
	AutoStop           bool                   `json:"auto_stop"`
	IsActive           bool                   `json:"is_active"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// GetDurationSeconds returns duration in seconds as float64
func (ri *RecordingInfo) GetDurationSeconds() float64 {
	return ri.Duration.Seconds()
}

// GetStartTimeISO returns start time in ISO 8601 format
func (ri *RecordingInfo) GetStartTimeISO() string {
	return ri.StartTime.Format(time.RFC3339)
}

// GetEndTime calculates the expected end time for auto-stop recordings
func (ri *RecordingInfo) GetEndTime() *time.Time {
	if ri.AutoStop && ri.ConfiguredDuration > 0 {
		endTime := ri.StartTime.Add(ri.ConfiguredDuration)
		return &endTime
	}
	return nil
}

// GetEndTimeISO returns expected end time in ISO 8601 format
func (ri *RecordingInfo) GetEndTimeISO() string {
	if endTime := ri.GetEndTime(); endTime != nil {
		return endTime.Format(time.RFC3339)
	}
	return ""
}

// StopAll stops all active recording timers with context timeout
func (rtm *RecordingTimerManager) StopAll(ctx context.Context) error {
	var stoppedTimers []string

	// Collect all timer keys
	rtm.timers.Range(func(key, value interface{}) bool {
		cameraID := key.(string)
		stoppedTimers = append(stoppedTimers, cameraID)
		return true
	})

	// Stop each timer
	for _, cameraID := range stoppedTimers {
		rtm.DeleteTimer(cameraID)
	}

	rtm.logger.WithField("stopped_count", fmt.Sprintf("%d", len(stoppedTimers))).Info("All recording timers stopped")
	return nil
}
