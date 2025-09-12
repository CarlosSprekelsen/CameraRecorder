/*
Device event source implementations for udev/fsnotify integration.

Provides event-driven device discovery using fsnotify (default) and optional udev
as the primary discovery mechanism with polling as fallback.

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration
- REQ-CAM-002: Real-time device status monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/fsnotify/fsnotify"
)

// FsnotifyDeviceEventSource implements DeviceEventSource using fsnotify
// This is the default implementation that works in containers without CGO
type FsnotifyDeviceEventSource struct {
	logger    *logging.Logger
	watcher   *fsnotify.Watcher
	events    chan DeviceEvent
	stopChan  chan struct{}
	running   bool
	mu        sync.RWMutex
	eventChan chan DeviceEvent
}

// NewFsnotifyDeviceEventSource creates a new fsnotify-based device event source
func NewFsnotifyDeviceEventSource(logger *logging.Logger) (*FsnotifyDeviceEventSource, error) {
	if logger == nil {
		logger = logging.GetLogger("device-event-source")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &FsnotifyDeviceEventSource{
		logger:    logger,
		watcher:   watcher,
		events:    make(chan DeviceEvent, 100), // Buffered to prevent blocking
		stopChan:  make(chan struct{}),
		eventChan: make(chan DeviceEvent, 100),
	}, nil
}

// Start begins monitoring for device events
func (f *FsnotifyDeviceEventSource) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return fmt.Errorf("device event source is already running")
	}

	// Watch /dev directory for device changes
	err := f.watcher.Add("/dev")
	if err != nil {
		return fmt.Errorf("failed to watch /dev directory: %w", err)
	}

	f.running = true
	f.logger.WithFields(logging.Fields{
		"action": "device_event_source_started",
		"type":   "fsnotify",
	}).Info("Started fsnotify device event source")

	// Start event processing goroutine
	go f.eventLoop(ctx)

	return nil
}

// Events returns the channel for device events
func (f *FsnotifyDeviceEventSource) Events() <-chan DeviceEvent {
	return f.events
}

// Close stops the device event source
func (f *FsnotifyDeviceEventSource) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return nil
	}

	f.running = false
	close(f.stopChan)

	if f.watcher != nil {
		err := f.watcher.Close()
		if err != nil {
			f.logger.WithError(err).Warn("Error closing fsnotify watcher")
		}
	}

	close(f.events)
	close(f.eventChan)

	f.logger.WithFields(logging.Fields{
		"action": "device_event_source_stopped",
		"type":   "fsnotify",
	}).Info("Stopped fsnotify device event source")

	return nil
}

// eventLoop processes fsnotify events and filters for video devices
func (f *FsnotifyDeviceEventSource) eventLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			f.logger.WithFields(logging.Fields{
				"panic":  r,
				"action": "panic_recovered",
			}).Error("Recovered from panic in fsnotify event loop")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			f.logger.WithFields(logging.Fields{
				"action": "event_loop_stopped",
				"reason": "context_cancelled",
			}).Debug("Fsnotify event loop stopped due to context cancellation")
			return
		case <-f.stopChan:
			f.logger.WithFields(logging.Fields{
				"action": "event_loop_stopped",
				"reason": "stop_requested",
			}).Debug("Fsnotify event loop stopped")
			return
		case event, ok := <-f.watcher.Events:
			if !ok {
				f.logger.Debug("Fsnotify watcher events channel closed")
				return
			}
			f.processEvent(event)
		case err, ok := <-f.watcher.Errors:
			if !ok {
				f.logger.Debug("Fsnotify watcher errors channel closed")
				return
			}
			f.logger.WithError(err).Warn("Fsnotify watcher error")
		}
	}
}

// processEvent processes a single fsnotify event and filters for video devices
func (f *FsnotifyDeviceEventSource) processEvent(event fsnotify.Event) {
	// Only process video device events
	if !strings.HasPrefix(filepath.Base(event.Name), "video") {
		return
	}

	// Determine event type based on fsnotify operation
	var eventType DeviceEventType
	switch {
	case event.Op&fsnotify.Create != 0:
		eventType = DeviceEventAdd
	case event.Op&fsnotify.Remove != 0:
		eventType = DeviceEventRemove
	case event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Chmod != 0:
		eventType = DeviceEventChange
	default:
		// Skip other operations
		return
	}

	deviceEvent := DeviceEvent{
		Type:       eventType,
		DevicePath: event.Name,
		Timestamp:  time.Now(),
	}

	f.logger.WithFields(logging.Fields{
		"action":      "device_event_processed",
		"device_path": deviceEvent.DevicePath,
		"event_type":  deviceEvent.Type,
		"fsnotify_op": event.Op.String(),
	}).Debug("Processed device event")

	// Send event to channel (non-blocking)
	select {
	case f.events <- deviceEvent:
		// Event sent successfully
	default:
		f.logger.WithFields(logging.Fields{
			"device_path": deviceEvent.DevicePath,
			"event_type":  deviceEvent.Type,
			"action":      "event_dropped",
		}).Warn("Device event dropped - channel full")
	}
}

// UdevDeviceEventSource implements DeviceEventSource using libudev
// This is an optional implementation that requires CGO and libudev
// It's provided as a placeholder for future implementation
type UdevDeviceEventSource struct {
	logger   *logging.Logger
	events   chan DeviceEvent
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex
}

// NewUdevDeviceEventSource creates a new udev-based device event source
// This is a placeholder implementation - real udev integration would go here
func NewUdevDeviceEventSource(logger *logging.Logger) (*UdevDeviceEventSource, error) {
	if logger == nil {
		logger = logging.GetLogger("device-event-source")
	}

	return &UdevDeviceEventSource{
		logger:   logger,
		events:   make(chan DeviceEvent, 100),
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins monitoring for device events using udev
func (u *UdevDeviceEventSource) Start(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.running {
		return fmt.Errorf("device event source is already running")
	}

	u.running = true
	u.logger.WithFields(logging.Fields{
		"action": "device_event_source_started",
		"type":   "udev",
	}).Info("Started udev device event source (placeholder)")

	// TODO: Implement real udev integration
	// This would involve:
	// 1. Creating a udev monitor
	// 2. Filtering for video devices
	// 3. Processing netlink events
	// 4. Converting to DeviceEvent structs

	return nil
}

// Events returns the channel for device events
func (u *UdevDeviceEventSource) Events() <-chan DeviceEvent {
	return u.events
}

// Close stops the device event source
func (u *UdevDeviceEventSource) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.running {
		return nil
	}

	u.running = false
	close(u.stopChan)
	close(u.events)

	u.logger.WithFields(logging.Fields{
		"action": "device_event_source_stopped",
		"type":   "udev",
	}).Info("Stopped udev device event source")

	return nil
}
