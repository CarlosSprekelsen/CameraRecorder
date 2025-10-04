package camera

import (
	"context"
	"time"
)

// CameraEvent represents camera connection events
type CameraEvent string

const (
	CameraEventConnected     CameraEvent = "CONNECTED"
	CameraEventDisconnected  CameraEvent = "DISCONNECTED"
	CameraEventStatusChanged CameraEvent = "STATUS_CHANGED"
	CameraEventReady         CameraEvent = "READY"
)

// CameraEventData represents data structure for camera events
type CameraEventData struct {
	DevicePath string        `json:"device_path"`
	EventType  CameraEvent   `json:"event_type"`
	Timestamp  time.Time     `json:"timestamp"`
	DeviceInfo *CameraDevice `json:"device_info,omitempty"`
}

// CameraEventHandler handles camera events asynchronously.
//
// Implementations must handle context cancellation and return errors for
// failed event processing. Event handlers are called concurrently via
// a bounded worker pool to prevent blocking the main monitoring loop.
type CameraEventHandler interface {
	HandleCameraEvent(ctx context.Context, eventData CameraEventData) error
}

// DeviceChecker validates device file existence on the file system.
//
// Implementations must return true if the device path exists and is accessible,
// false otherwise. Used for USB camera device validation before capability probing.
type DeviceChecker interface {
	Exists(path string) bool
}

// V4L2CommandExecutor executes V4L2 commands on camera devices.
//
// Implementations must handle command execution with proper context cancellation
// and timeout handling. Commands are executed via v4l2-ctl or equivalent tools.
// Returns the command output or an error if execution fails, device is unavailable,
// or context is canceled.
type V4L2CommandExecutor interface {
	ExecuteCommand(ctx context.Context, devicePath, args string) (string, error)
}

// DeviceInfoParser parses V4L2 command output into structured device information.
//
// Implementations must handle malformed input gracefully and return meaningful
// errors for parsing failures. Parses device capabilities, supported formats,
// and frame rates from v4l2-ctl command output.
type DeviceInfoParser interface {
	ParseDeviceInfo(output string) (V4L2Capabilities, error)
	ParseDeviceFormats(output string) ([]V4L2Format, error)
	ParseDeviceFrameRates(output string) ([]string, error)
}

// EventNotifier sends camera events to external event systems.
//
// Implementations must handle notification failures gracefully and not block
// the main monitoring loop. Used for integrating with external event systems
// like WebSocket APIs or message queues.
type EventNotifier interface {
	NotifyCameraConnected(device *CameraDevice)
	NotifyCameraDisconnected(devicePath string)
	NotifyCameraStatusChange(device *CameraDevice, oldStatus, newStatus DeviceStatus)
	NotifyCapabilityDetected(device *CameraDevice, capabilities V4L2Capabilities)
	NotifyCapabilityError(devicePath string, error string)
}

// DeviceEventSource provides device connection/disconnection events.
//
// Implementations must support graceful startup/shutdown and return a channel
// of device events. Used for event-driven device discovery via udev or fsnotify.
// EventsSupported() returns false if falling back to polling-only mode.
type DeviceEventSource interface {
	Start(ctx context.Context) error
	Events() <-chan DeviceEvent
	Close() error
	EventsSupported() bool
	Started() bool
}

// DeviceEvent represents a device event from udev/fsnotify
type DeviceEvent struct {
	Type       DeviceEventType `json:"type"`
	DevicePath string          `json:"device_path"`
	Vendor     string          `json:"vendor,omitempty"`
	Product    string          `json:"product,omitempty"`
	Serial     string          `json:"serial,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

// DeviceEventType represents the type of device event
type DeviceEventType string

const (
	DeviceEventAdd    DeviceEventType = "add"
	DeviceEventRemove DeviceEventType = "remove"
	DeviceEventChange DeviceEventType = "change"
)

// CameraMonitor manages camera discovery, monitoring, and lifecycle.
//
// Implementations must support concurrent access, graceful shutdown, and
// event-driven device discovery. Provides unified interface for camera
// management across USB, IP, and RTSP camera types.
type CameraMonitor interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
	IsReady() bool // indicates first discovery cycle completed
	GetConnectedCameras() map[string]*CameraDevice
	GetDevice(devicePath string) (*CameraDevice, bool)
	GetMonitorStats() *MonitorStats
	AddEventHandler(handler CameraEventHandler)
	AddEventCallback(callback func(CameraEventData))
	SetEventNotifier(notifier EventNotifier)

	// Event-driven readiness system
	SubscribeToReadiness() <-chan struct{}

	// V4L2 Direct Snapshot Capture (Tier 0 - Fastest)
	TakeDirectSnapshot(ctx context.Context, devicePath, outputPath string, options map[string]interface{}) (*DirectSnapshot, error)
}

// MonitorStats tracks monitoring statistics
type MonitorStats struct {
	Running                    bool    `json:"running"`
	ActiveTasks                int64   `json:"active_tasks"`
	PollingCycles              int64   `json:"polling_cycles"`
	DeviceStateChanges         int64   `json:"device_state_changes"`
	CapabilityProbesAttempted  int64   `json:"capability_probes_attempted"`
	CapabilityProbesSuccessful int64   `json:"capability_probes_successful"`
	CapabilityTimeouts         int64   `json:"capability_timeouts"`
	CapabilityParseErrors      int64   `json:"capability_parse_errors"`
	PollingFailureCount        int64   `json:"polling_failure_count"`
	CurrentPollInterval        float64 `json:"current_poll_interval"`
	KnownDevicesCount          int64   `json:"known_devices_count"`
	UdevEventsProcessed        int64   `json:"udev_events_processed"`
	UdevEventsFiltered         int64   `json:"udev_events_filtered"`
	UdevEventsSkipped          int64   `json:"udev_events_skipped"`
	DeviceEventsProcessed      int64   `json:"device_events_processed"`
	DeviceEventsDropped        int64   `json:"device_events_dropped"`
	DevicesConnected           int64   `json:"devices_connected"`
	// Removed mu sync.RWMutex - using atomic operations instead
}

// ResourceManager interface for components that need lifecycle management
type ResourceManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
}

// CleanupManager interface for components that need resource cleanup
type CleanupManager interface {
	Cleanup(ctx context.Context) error
	GetResourceStats() map[string]interface{}
}

// BoundedWorkerPool interface for managing goroutine pools with resource limits
type BoundedWorkerPool interface {
	ResourceManager // Includes Start, Stop, IsRunning methods
	Submit(ctx context.Context, task func(context.Context)) error
	GetStats() WorkerPoolStats
}

// WorkerPoolStats represents statistics for a worker pool
type WorkerPoolStats struct {
	ActiveWorkers  int   `json:"active_workers"`
	QueuedTasks    int   `json:"queued_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	TimeoutTasks   int64 `json:"timeout_tasks"`
	MaxWorkers     int   `json:"max_workers"`
}

// CapabilityDetectionResult represents the result of device capability detection
type CapabilityDetectionResult struct {
	Detected              bool                   `json:"detected"`
	Accessible            bool                   `json:"accessible"`
	DeviceName            string                 `json:"device_name"`
	Driver                string                 `json:"driver"`
	Formats               []string               `json:"formats"`
	Resolutions           []string               `json:"resolutions"`
	FrameRates            []string               `json:"frame_rates"`
	Error                 string                 `json:"error,omitempty"`
	TimeoutContext        string                 `json:"timeout_context,omitempty"`
	ProbeTimestamp        time.Time              `json:"probe_timestamp"`
	StructuredDiagnostics map[string]interface{} `json:"structured_diagnostics,omitempty"`
}
