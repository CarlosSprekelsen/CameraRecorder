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
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/fsnotify/fsnotify"
)

// FsnotifyDeviceEventSource implements DeviceEventSource using fsnotify
// This is the default implementation that works in containers without CGO
type FsnotifyDeviceEventSource struct {
	logger          *logging.Logger
	watcher         *fsnotify.Watcher
	events          chan DeviceEvent
	stopChan        chan struct{}
	running         int32          // Using atomic operations instead of mutex
	done            sync.WaitGroup // Wait for event loop to exit
	eventsSupported int32          // Whether fsnotify events are supported
	started         int32          // Whether the event source has started
	startCallsTotal int32          // Debug counter for Start() calls
}

// NewFsnotifyDeviceEventSource creates a new fsnotify-based device event source
// Note: Watcher is created lazily in Start() to prevent resource leaks
func NewFsnotifyDeviceEventSource(logger *logging.Logger) (*FsnotifyDeviceEventSource, error) {
	if logger == nil {
		logger = logging.GetLogger("device-event-source")
	}

	return &FsnotifyDeviceEventSource{
		logger:   logger,
		watcher:  nil,                         // Will be created in Start()
		events:   make(chan DeviceEvent, 100), // Buffered to prevent blocking
		stopChan: make(chan struct{}),
		done:     sync.WaitGroup{},
	}, nil
}

// Start begins monitoring for device events
func (f *FsnotifyDeviceEventSource) Start(ctx context.Context) error {
	// Increment debug counter and log
	callCount := atomic.AddInt32(&f.startCallsTotal, 1)
	esID := fmt.Sprintf("es_%d", time.Now().UnixNano())
	f.logger.WithFields(logging.Fields{
		"es_id":             esID,
		"start_calls_total": callCount,
		"action":            "es_start_called",
	}).Info("Device event source Start() called")
	
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use atomic compare-and-swap to set running state atomically
	if !atomic.CompareAndSwapInt32(&f.running, 0, 1) {
		return fmt.Errorf("device event source is already running")
	}

	// Create a new watcher for each start (fsnotify watchers can't be reused)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// Reset running state if watcher creation fails
		atomic.StoreInt32(&f.running, 0)
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	f.watcher = watcher

	// Recreate channels for each start
	f.events = make(chan DeviceEvent, 100)
	f.stopChan = make(chan struct{})

	// Watch /dev directory for device changes
	err = f.watcher.Add("/dev")
	if err != nil {
		// fsnotify not available (container perms, etc.) - mark as poll-only mode
		f.logger.WithError(err).Warn("fsnotify not available, falling back to poll-only mode")
		atomic.StoreInt32(&f.eventsSupported, 0) // false
		f.watcher.Close()
		f.watcher = nil
	} else {
		// fsnotify is working
		atomic.StoreInt32(&f.eventsSupported, 1) // true
	}

	// Mark as started regardless of fsnotify availability
	atomic.StoreInt32(&f.started, 1)

	f.logger.WithFields(logging.Fields{
		"action": "device_event_source_started",
		"type":   "fsnotify",
	}).Info("Started fsnotify device event source")

	// Start event processing goroutine
	f.done.Add(1)
	go f.eventLoop(ctx)

	return nil
}

// Events returns the channel for device events
func (f *FsnotifyDeviceEventSource) Events() <-chan DeviceEvent {
	return f.events
}

// Close stops the device event source
func (f *FsnotifyDeviceEventSource) Close() error {
	// Use atomic compare-and-swap to set running state atomically
	if !atomic.CompareAndSwapInt32(&f.running, 1, 0) {
		return nil // Already stopped
	}

	// Safely close stop channel to signal event loop to exit
	select {
	case <-f.stopChan:
		// Already closed
	default:
		close(f.stopChan)
	}

	// Wait for event loop to exit (it will close the events channel)
	f.done.Wait()

	if f.watcher != nil {
		err := f.watcher.Close()
		if err != nil {
			f.logger.WithError(err).Warn("Error closing fsnotify watcher")
		}
	}

	f.logger.WithFields(logging.Fields{
		"action": "device_event_source_stopped",
		"type":   "fsnotify",
	}).Info("Stopped fsnotify device event source")

	return nil
}

// EventsSupported returns whether fsnotify events are supported
func (f *FsnotifyDeviceEventSource) EventsSupported() bool {
	return atomic.LoadInt32(&f.eventsSupported) == 1
}

// Started returns whether the event source has started
func (f *FsnotifyDeviceEventSource) Started() bool {
	return atomic.LoadInt32(&f.started) == 1
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
		// Close events channel when loop exits (producer owns channel)
		close(f.events)
		// Signal that the loop has exited
		f.done.Done()
	}()

	// If fsnotify is not supported, just wait for stop signal
	if f.watcher == nil {
		f.logger.Debug("Running in poll-only mode (no fsnotify events)")
		select {
		case <-ctx.Done():
			return
		case <-f.stopChan:
			return
		}
	}

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
	running  int32 // Using atomic operations instead of mutex
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
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use atomic compare-and-swap to set running state atomically
	if !atomic.CompareAndSwapInt32(&u.running, 0, 1) {
		return fmt.Errorf("device event source is already running")
	}

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
	// Use atomic compare-and-swap to set running state atomically
	if !atomic.CompareAndSwapInt32(&u.running, 1, 0) {
		return nil // Already stopped
	}

	// Safely close channels
	select {
	case <-u.stopChan:
		// Already closed
	default:
		close(u.stopChan)
	}

	select {
	case <-u.events:
		// Already closed
	default:
		close(u.events)
	}

	u.logger.WithFields(logging.Fields{
		"action": "device_event_source_stopped",
		"type":   "udev",
	}).Info("Stopped udev device event source")

	return nil
}

// EventsSupported returns whether udev events are supported
func (u *UdevDeviceEventSource) EventsSupported() bool {
	return true // udev always supports events when available
}

// Started returns whether the event source has started
func (u *UdevDeviceEventSource) Started() bool {
	return atomic.LoadInt32(&u.running) == 1
}

// DeviceEventSourceFactory manages singleton device event sources with ref counting
// This prevents resource leaks by ensuring only one fsnotify watcher per process
type DeviceEventSourceFactory struct {
	mu       sync.RWMutex
	instance *FsnotifyDeviceEventSource
	refCount int
	logger   *logging.Logger
}

var (
	globalFactory *DeviceEventSourceFactory
	factoryOnce   sync.Once
)

// GetDeviceEventSourceFactory returns the global singleton factory
func GetDeviceEventSourceFactory() *DeviceEventSourceFactory {
	factoryOnce.Do(func() {
		globalFactory = &DeviceEventSourceFactory{
			logger: logging.GetLogger("device-event-source-factory"),
		}
	})
	return globalFactory
}

// Acquire returns a device event source instance, creating it if needed
// Increments the reference count
func (f *DeviceEventSourceFactory) Acquire() DeviceEventSource {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.instance == nil {
		// Create new instance without allocating watcher yet
		f.instance = &FsnotifyDeviceEventSource{
			logger:          f.logger,
			watcher:         nil, // Will be created in Start()
			events:          make(chan DeviceEvent, 100),
			stopChan:        make(chan struct{}),
			running:         0,
			done:            sync.WaitGroup{},
			eventsSupported: 0, // Will be set in Start()
			started:         0, // Will be set in Start()
		}
		f.logger.Info("Created new device event source instance")
	}

	f.refCount++
	f.logger.WithField("ref_count", fmt.Sprintf("%d", f.refCount)).Debug("Acquired device event source")
	return f.instance
}

// Release decrements the reference count and closes the instance when count reaches zero
func (f *DeviceEventSourceFactory) Release() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.refCount <= 0 {
		f.logger.Error("Release called with zero or negative ref count - this indicates a bug")
		return fmt.Errorf("release underflow: refCount=%d", f.refCount)
	}

	f.refCount--
	f.logger.WithField("ref_count", fmt.Sprintf("%d", f.refCount)).Debug("Released device event source")

	if f.refCount == 0 && f.instance != nil {
		f.logger.Info("Closing device event source - final reference released")
		err := f.instance.Close()
		f.instance = nil
		return err
	}

	return nil
}

// ResetForTests forces cleanup of the singleton for test isolation
// This should only be called in test cleanup to ensure no resource leaks
func (f *DeviceEventSourceFactory) ResetForTests() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.instance != nil {
		f.logger.Info("Force closing device event source for test cleanup")
		f.instance.Close()
		f.instance = nil
	}
	f.refCount = 0
}
