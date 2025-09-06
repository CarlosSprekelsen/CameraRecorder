/*
Event System Integration Layer

Connects camera monitor and other components to the WebSocket event system,
implementing the EventNotifier interface for seamless event propagation.

Requirements Coverage:
- REQ-API-001: Efficient event delivery
- REQ-API-002: Component integration
- REQ-API-003: Event routing

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// EventIntegration connects camera monitor and other components to the event system
type EventIntegration struct {
	eventManager *EventManager
	logger       *logging.Logger
}

// NewEventIntegration creates a new event integration layer
func NewEventIntegration(eventManager *EventManager, logger *logging.Logger) *EventIntegration {
	return &EventIntegration{
		eventManager: eventManager,
		logger:       logger,
	}
}

// CameraEventNotifier implements the camera.EventNotifier interface
type CameraEventNotifier struct {
	eventManager *EventManager
	logger       *logging.Logger
}

// NewCameraEventNotifier creates a new camera event notifier
func NewCameraEventNotifier(eventManager *EventManager, logger *logging.Logger) *CameraEventNotifier {
	return &CameraEventNotifier{
		eventManager: eventManager,
		logger:       logger,
	}
}

// NotifyCameraConnected notifies when a camera is connected
func (n *CameraEventNotifier) NotifyCameraConnected(device *camera.CameraDevice) {
	eventData := map[string]interface{}{
		"device":    device.Path,
		"name":      device.Name,
		"status":    string(device.Status),
		"driver":    device.Capabilities.DriverName,
		"card_name": device.Capabilities.CardName,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicCameraConnected, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device.Path).Error("Failed to publish camera connected event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device": device.Path,
			"name":   device.Name,
			"topic":  TopicCameraConnected,
		}).Debug("Published camera connected event")
	}
}

// NotifyCameraDisconnected notifies when a camera is disconnected
func (n *CameraEventNotifier) NotifyCameraDisconnected(devicePath string) {
	eventData := map[string]interface{}{
		"device":    devicePath,
		"status":    "disconnected",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicCameraDisconnected, eventData); err != nil {
		n.logger.WithError(err).WithField("device", devicePath).Error("Failed to publish camera disconnected event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device": devicePath,
			"topic":  TopicCameraDisconnected,
		}).Debug("Published camera disconnected event")
	}
}

// NotifyCameraStatusChange notifies when camera status changes
func (n *CameraEventNotifier) NotifyCameraStatusChange(device *camera.CameraDevice, oldStatus, newStatus camera.DeviceStatus) {
	eventData := map[string]interface{}{
		"device":     device.Path,
		"name":       device.Name,
		"old_status": string(oldStatus),
		"new_status": string(newStatus),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicCameraStatusChange, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device.Path).Error("Failed to publish camera status change event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device":     device.Path,
			"old_status": oldStatus,
			"new_status": newStatus,
			"topic":      TopicCameraStatusChange,
		}).Debug("Published camera status change event")
	}
}

// NotifyCapabilityDetected notifies when camera capabilities are detected
func (n *CameraEventNotifier) NotifyCapabilityDetected(device *camera.CameraDevice, capabilities camera.V4L2Capabilities) {
	eventData := map[string]interface{}{
		"device":       device.Path,
		"name":         device.Name,
		"driver":       capabilities.DriverName,
		"card_name":    capabilities.CardName,
		"bus_info":     capabilities.BusInfo,
		"capabilities": capabilities.Capabilities,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicCameraCapabilityDetected, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device.Path).Error("Failed to publish capability detected event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device": device.Path,
			"driver": capabilities.DriverName,
			"topic":  TopicCameraCapabilityDetected,
		}).Debug("Published capability detected event")
	}
}

// NotifyCapabilityError notifies when camera capability detection fails
func (n *CameraEventNotifier) NotifyCapabilityError(devicePath string, errorMsg string) {
	eventData := map[string]interface{}{
		"device":    devicePath,
		"error":     errorMsg,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicCameraCapabilityError, eventData); err != nil {
		n.logger.WithError(err).WithField("device", devicePath).Error("Failed to publish capability error event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device": devicePath,
			"error":  errorMsg,
			"topic":  TopicCameraCapabilityError,
		}).Debug("Published capability error event")
	}
}

// MediaMTXEventNotifier implements MediaMTX event notifications
type MediaMTXEventNotifier struct {
	eventManager *EventManager
	logger       *logging.Logger
}

// NewMediaMTXEventNotifier creates a new MediaMTX event notifier
func NewMediaMTXEventNotifier(eventManager *EventManager, logger *logging.Logger) *MediaMTXEventNotifier {
	return &MediaMTXEventNotifier{
		eventManager: eventManager,
		logger:       logger,
	}
}

// NotifyRecordingStarted notifies when MediaMTX recording starts
func (n *MediaMTXEventNotifier) NotifyRecordingStarted(device, sessionID, filename string) {
	eventData := map[string]interface{}{
		"device":     device,
		"session_id": sessionID,
		"filename":   filename,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicMediaMTXRecordingStarted, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device).Error("Failed to publish recording started event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device":     device,
			"session_id": sessionID,
			"filename":   filename,
			"topic":      TopicMediaMTXRecordingStarted,
		}).Debug("Published recording started event")
	}
}

// NotifyRecordingStopped notifies when MediaMTX recording stops
func (n *MediaMTXEventNotifier) NotifyRecordingStopped(device, sessionID, filename string, duration time.Duration) {
	eventData := map[string]interface{}{
		"device":     device,
		"session_id": sessionID,
		"filename":   filename,
		"duration":   duration.Seconds(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicMediaMTXRecordingStopped, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device).Error("Failed to publish recording stopped event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device":     device,
			"session_id": sessionID,
			"filename":   filename,
			"duration":   duration,
			"topic":      TopicMediaMTXRecordingStopped,
		}).Debug("Published recording stopped event")
	}
}

// NotifyStreamStarted notifies when MediaMTX stream starts
func (n *MediaMTXEventNotifier) NotifyStreamStarted(device, streamID, streamType string) {
	eventData := map[string]interface{}{
		"device":      device,
		"stream_id":   streamID,
		"stream_type": streamType,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicMediaMTXStreamStarted, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device).Error("Failed to publish stream started event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device":      device,
			"stream_id":   streamID,
			"stream_type": streamType,
			"topic":       TopicMediaMTXStreamStarted,
		}).Debug("Published stream started event")
	}
}

// NotifyStreamStopped notifies when MediaMTX stream stops
func (n *MediaMTXEventNotifier) NotifyStreamStopped(device, streamID, streamType string) {
	eventData := map[string]interface{}{
		"device":      device,
		"stream_id":   streamID,
		"stream_type": streamType,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicMediaMTXStreamStopped, eventData); err != nil {
		n.logger.WithError(err).WithField("device", device).Error("Failed to publish stream stopped event")
	} else {
		n.logger.WithFields(logging.Fields{
			"device":      device,
			"stream_id":   streamID,
			"stream_type": streamType,
			"topic":       TopicMediaMTXStreamStopped,
		}).Debug("Published stream stopped event")
	}
}

// SystemEventNotifier implements system-level event notifications
type SystemEventNotifier struct {
	eventManager *EventManager
	logger       *logging.Logger
}

// NewSystemEventNotifier creates a new system event notifier
func NewSystemEventNotifier(eventManager *EventManager, logger *logging.Logger) *SystemEventNotifier {
	return &SystemEventNotifier{
		eventManager: eventManager,
		logger:       logger,
	}
}

// NotifySystemStartup notifies when the system starts up
func (n *SystemEventNotifier) NotifySystemStartup(version, buildInfo string) {
	eventData := map[string]interface{}{
		"version":    version,
		"build_info": buildInfo,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicSystemStartup, eventData); err != nil {
		n.logger.WithError(err).Error("Failed to publish system startup event")
	} else {
		n.logger.WithFields(logging.Fields{
			"version": version,
			"topic":   TopicSystemStartup,
		}).Info("Published system startup event")
	}
}

// NotifySystemShutdown notifies when the system shuts down
func (n *SystemEventNotifier) NotifySystemShutdown(reason string) {
	eventData := map[string]interface{}{
		"reason":    reason,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicSystemShutdown, eventData); err != nil {
		n.logger.WithError(err).Error("Failed to publish system shutdown event")
	} else {
		n.logger.WithFields(logging.Fields{
			"reason": reason,
			"topic":  TopicSystemShutdown,
		}).Info("Published system shutdown event")
	}
}

// NotifySystemHealth notifies about system health status
func (n *SystemEventNotifier) NotifySystemHealth(status string, metrics map[string]interface{}) {
	eventData := map[string]interface{}{
		"status":    status,
		"metrics":   metrics,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := n.eventManager.PublishEvent(TopicSystemHealth, eventData); err != nil {
		n.logger.WithError(err).Error("Failed to publish system health event")
	} else {
		n.logger.WithFields(logging.Fields{
			"status": status,
			"topic":  TopicSystemHealth,
		}).Debug("Published system health event")
	}
}
