package camera

import (
	"context"
	"sync"
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

// CameraMonitor interface for camera discovery and monitoring
type CameraMonitor interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	GetConnectedCameras() map[string]*CameraDevice
	GetDevice(devicePath string) (*CameraDevice, bool)
	GetMonitorStats() *MonitorStats
	AddEventHandler(handler CameraEventHandler)
	AddEventCallback(callback func(CameraEventData))
	SetEventNotifier(notifier EventNotifier)
}

// MonitorStats tracks monitoring statistics
type MonitorStats struct {
	Running                    bool    `json:"running"`
	ActiveTasks                int     `json:"active_tasks"`
	PollingCycles              int     `json:"polling_cycles"`
	DeviceStateChanges         int     `json:"device_state_changes"`
	CapabilityProbesAttempted  int     `json:"capability_probes_attempted"`
	CapabilityProbesSuccessful int     `json:"capability_probes_successful"`
	CapabilityTimeouts         int     `json:"capability_timeouts"`
	CapabilityParseErrors      int     `json:"capability_parse_errors"`
	PollingFailureCount        int     `json:"polling_failure_count"`
	CurrentPollInterval        float64 `json:"current_poll_interval"`
	KnownDevicesCount          int     `json:"known_devices_count"`
	UdevEventsProcessed        int     `json:"udev_events_processed"`
	UdevEventsFiltered         int     `json:"udev_events_filtered"`
	UdevEventsSkipped          int     `json:"udev_events_skipped"`
	mu                         sync.RWMutex
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
