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
)

// CameraEventData represents data structure for camera events
type CameraEventData struct {
	DevicePath string        `json:"device_path"`
	EventType  CameraEvent   `json:"event_type"`
	Timestamp  time.Time     `json:"timestamp"`
	DeviceInfo *CameraDevice `json:"device_info,omitempty"`
}

// CameraEventHandler interface for handling camera events
type CameraEventHandler interface {
	HandleCameraEvent(ctx context.Context, eventData CameraEventData) error
}

// DeviceChecker interface for device existence checking
type DeviceChecker interface {
	Exists(path string) bool
}

// V4L2CommandExecutor interface for V4L2 command execution
type V4L2CommandExecutor interface {
	ExecuteCommand(ctx context.Context, devicePath, args string) (string, error)
}

// DeviceInfoParser interface for parsing device information
type DeviceInfoParser interface {
	ParseDeviceInfo(output string) (V4L2Capabilities, error)
	ParseDeviceFormats(output string) ([]V4L2Format, error)
	ParseDeviceFrameRates(output string) ([]string, error)
}

// EventNotifier interface for sending camera events to external systems
type EventNotifier interface {
	NotifyCameraConnected(device *CameraDevice)
	NotifyCameraDisconnected(devicePath string)
	NotifyCameraStatusChange(device *CameraDevice, oldStatus, newStatus DeviceStatus)
	NotifyCapabilityDetected(device *CameraDevice, capabilities V4L2Capabilities)
	NotifyCapabilityError(devicePath string, error string)
}

// DeviceEventSource interface for device event discovery (udev/fsnotify abstraction)
type DeviceEventSource interface {
	Start(ctx context.Context) error
	Events() <-chan DeviceEvent
	Close() error
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

// CameraMonitor interface for camera discovery and monitoring
type CameraMonitor interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	IsReady() bool // indicates first discovery cycle completed
	GetConnectedCameras() map[string]*CameraDevice
	GetDevice(devicePath string) (*CameraDevice, bool)
	GetMonitorStats() *MonitorStats
	AddEventHandler(handler CameraEventHandler)
	AddEventCallback(callback func(CameraEventData))
	SetEventNotifier(notifier EventNotifier)

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
