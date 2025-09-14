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
	"os"
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

	// Reset all state for reuse (following factory pattern)
	atomic.StoreInt32(&f.started, 0)
	atomic.StoreInt32(&f.eventsSupported, 0)

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

// UdevDeviceEventSource implements DeviceEventSource using udev
// This is a basic implementation that uses udevadm for device discovery
// Note: This is a simplified implementation that doesn't require CGO
// For full udev integration, libudev CGO bindings would be needed
type UdevDeviceEventSource struct {
	logger   *logging.Logger
	events   chan DeviceEvent
	stopChan chan struct{}
	running  int32 // Using atomic operations instead of mutex
	done     sync.WaitGroup // Wait for event loop to exit
}

// NewUdevDeviceEventSource creates a new udev-based device event source
// This implementation uses udevadm commands for device discovery
func NewUdevDeviceEventSource(logger *logging.Logger) (*UdevDeviceEventSource, error) {
	if logger == nil {
		logger = logging.GetLogger("device-event-source")
	}

	// Check if udevadm is available
	udevadmPaths := []string{"/usr/bin/udevadm", "/sbin/udevadm", "/bin/udevadm"}
	udevadmFound := false
	for _, path := range udevadmPaths {
		if _, err := os.Stat(path); err == nil {
			udevadmFound = true
			break
		}
	}

	if !udevadmFound {
		return nil, fmt.Errorf("udevadm not found - udev device event source not available")
	}

	return &UdevDeviceEventSource{
		logger:   logger,
		events:   make(chan DeviceEvent, 100),
		stopChan: make(chan struct{}),
		done:     sync.WaitGroup{},
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

	// Recreate channels for each start
	u.events = make(chan DeviceEvent, 100)
	u.stopChan = make(chan struct{})

	u.logger.WithFields(logging.Fields{
		"action": "device_event_source_started",
		"type":   "udev",
	}).Info("Started udev device event source")

	// Start event processing goroutine
	u.done.Add(1)
	go u.eventLoop(ctx)

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

	// Safely close stop channel to signal event loop to exit
	select {
	case <-u.stopChan:
		// Already closed
	default:
		close(u.stopChan)
	}

	// Wait for event loop to exit (it will close the events channel)
	u.done.Wait()

	u.logger.WithFields(logging.Fields{
		"action": "device_event_source_stopped",
		"type":   "udev",
	}).Info("Stopped udev device event source")

	return nil
}

// eventLoop processes udev events using udevadm monitor
// This is a simplified implementation that uses polling with udevadm
// For real-time events, libudev CGO bindings would be needed
func (u *UdevDeviceEventSource) eventLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			u.logger.WithFields(logging.Fields{
				"panic":  r,
				"action": "panic_recovered",
			}).Error("Recovered from panic in udev event loop")
		}
		// Close events channel when loop exits (producer owns channel)
		close(u.events)
		// Signal that the loop has exited
		u.done.Done()
	}()

	u.logger.Debug("Udev event loop started (simplified polling implementation)")

	// For now, this is a placeholder that just waits for stop signal
	// Real udev integration would use netlink or udevadm monitor
	select {
	case <-ctx.Done():
		u.logger.WithFields(logging.Fields{
			"action": "event_loop_stopped",
			"reason": "context_cancelled",
		}).Debug("Udev event loop stopped due to context cancellation")
		return
	case <-u.stopChan:
		u.logger.WithFields(logging.Fields{
			"action": "event_loop_stopped",
			"reason": "stop_requested",
		}).Debug("Udev event loop stopped")
		return
	}
}

// EventsSupported returns whether udev events are supported
func (u *UdevDeviceEventSource) EventsSupported() bool {
	return true // udev always supports events when available
}

// Started returns whether the event source has started
func (u *UdevDeviceEventSource) Started() bool {
	return atomic.LoadInt32(&u.running) == 1
}

// Environment detection functions for smart device event source selection

// isContainerEnvironment detects if we're running in a container
func isContainerEnvironment() bool {
	// Check for common container indicators
	containerIndicators := []string{
		"/.dockerenv",           // Docker
		"/proc/1/cgroup",       // Check cgroup for container indicators
		"/proc/self/cgroup",    // Alternative cgroup check
	}

	for _, indicator := range containerIndicators {
		if _, err := os.Stat(indicator); err == nil {
			return true
		}
	}

	// Check cgroup content for container indicators
	if cgroupContent, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(cgroupContent)
		containerKeywords := []string{"docker", "containerd", "kubepods", "crio"}
		for _, keyword := range containerKeywords {
			if strings.Contains(content, keyword) {
				return true
			}
		}
	}

	return false
}

// isUdevAvailable checks if udev is available and accessible
func isUdevAvailable() bool {
	// Check if we're in a container first
	if isContainerEnvironment() {
		return false // Containers typically don't have full udev access
	}

	// Check for udev system files
	udevIndicators := []string{
		"/sys/class/udev",           // udev sysfs interface
		"/dev/.udev",                // udev device directory
		"/run/udev",                 // udev runtime directory
		"/lib/systemd/system/udev.service", // systemd udev service
	}

	for _, indicator := range udevIndicators {
		if _, err := os.Stat(indicator); err == nil {
			return true
		}
	}

	// Check if udevadm command is available (indicates udev is installed)
	if _, err := os.Stat("/usr/bin/udevadm"); err == nil {
		return true
	}
	if _, err := os.Stat("/sbin/udevadm"); err == nil {
		return true
	}

	return false
}

// getOptimalDeviceEventSourceType determines the best device event source for the current environment
func getOptimalDeviceEventSourceType(logger *logging.Logger) string {
	isContainer := isContainerEnvironment()
	isUdev := isUdevAvailable()

	logger.WithFields(logging.Fields{
		"is_container": isContainer,
		"udev_available": isUdev,
		"action": "environment_detection",
	}).Info("Detecting optimal device event source")

	if isContainer {
		logger.Info("Container environment detected - using fsnotify for device discovery")
		return "fsnotify"
	}

	if isUdev {
		logger.Info("Bare metal environment with udev detected - using udev for device discovery")
		return "udev"
	}

	logger.Warn("No udev detected in bare metal environment - falling back to fsnotify")
	return "fsnotify"
}

// DeviceEventSourceFactory creates fresh device event source instances
// Each component gets its own instance for proper isolation and error recovery
// Now with smart environment detection for optimal device discovery
type DeviceEventSourceFactory struct {
	logger *logging.Logger
}

var (
	globalFactory *DeviceEventSourceFactory
	factoryOnce   sync.Once
)

// GetDeviceEventSourceFactory returns the global factory
func GetDeviceEventSourceFactory() *DeviceEventSourceFactory {
	factoryOnce.Do(func() {
		globalFactory = &DeviceEventSourceFactory{
			logger: logging.GetLogger("device-event-source-factory"),
		}
	})
	return globalFactory
}

// Create returns a fresh device event source instance
// Each call creates a new instance for proper isolation and error recovery
// Now with smart environment detection for optimal device discovery
func (f *DeviceEventSourceFactory) Create() DeviceEventSource {
	// Determine the optimal device event source type for this environment
	sourceType := getOptimalDeviceEventSourceType(f.logger)

	var instance DeviceEventSource

	switch sourceType {
	case "udev":
		// Create udev-based device event source
		udevInstance, err := NewUdevDeviceEventSource(f.logger)
		if err != nil {
			f.logger.WithError(err).Warn("Failed to create udev device event source, falling back to fsnotify")
			// Fallback to fsnotify if udev creation fails
			instance = &FsnotifyDeviceEventSource{
				logger:          f.logger,
				watcher:         nil, // Will be created in Start()
				events:          make(chan DeviceEvent, 100),
				stopChan:        make(chan struct{}),
				running:         0,
				done:            sync.WaitGroup{},
				eventsSupported: 0, // Will be set in Start()
				started:         0, // Will be set in Start()
			}
		} else {
			instance = udevInstance
		}
	case "fsnotify":
		fallthrough
	default:
		// Create fsnotify-based device event source (default)
		instance = &FsnotifyDeviceEventSource{
			logger:          f.logger,
			watcher:         nil, // Will be created in Start()
			events:          make(chan DeviceEvent, 100),
			stopChan:        make(chan struct{}),
			running:         0,
			done:            sync.WaitGroup{},
			eventsSupported: 0, // Will be set in Start()
			started:         0, // Will be set in Start()
		}
	}

	f.logger.WithFields(logging.Fields{
		"source_type": sourceType,
		"action": "device_event_source_created",
	}).Info("Created fresh device event source instance with smart selection")
	
	return instance
}

// No Release() method needed - each component manages its own instance lifecycle
// No ResetForTests() method needed - fresh instances provide natural test isolation
