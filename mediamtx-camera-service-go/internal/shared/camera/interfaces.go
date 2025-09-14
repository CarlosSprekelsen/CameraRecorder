package camera

import (
	"context"
	"time"
)

// CameraDevice represents a V4L2 video device
type CameraDevice struct {
	Path         string           `json:"path"`
	Name         string           `json:"name"`
	Capabilities V4L2Capabilities `json:"capabilities"`
	Formats      []V4L2Format     `json:"formats"`
	Status       DeviceStatus     `json:"status"`
	LastSeen     time.Time        `json:"last_seen"`
	DeviceNum    int              `json:"device_num"`
	Error        string           `json:"error,omitempty"`
	Vendor       string           `json:"vendor,omitempty"`
	Product      string           `json:"product,omitempty"`
	Serial       string           `json:"serial,omitempty"`
}

// DeviceStatus represents the current status of a V4L2 device
type DeviceStatus string

const (
	DeviceStatusConnected    DeviceStatus = "CONNECTED"
	DeviceStatusDisconnected DeviceStatus = "DISCONNECTED"
	DeviceStatusError        DeviceStatus = "ERROR"
	DeviceStatusProbing      DeviceStatus = "PROBING"
)

// V4L2Capabilities represents the capabilities of a V4L2 device
type V4L2Capabilities struct {
	DriverName   string   `json:"driver_name"`
	CardName     string   `json:"card_name"`
	BusInfo      string   `json:"bus_info"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	DeviceCaps   []string `json:"device_caps"`
}

// V4L2Format represents a video format supported by a V4L2 device
type V4L2Format struct {
	PixelFormat string   `json:"pixel_format"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	FrameRates  []string `json:"frame_rates"`
}

// CameraEventData represents camera device events
type CameraEventData struct {
	EventType string        `json:"event_type"`
	Device    *CameraDevice `json:"device"`
	Timestamp time.Time     `json:"timestamp"`
}

// CameraEventHandler handles camera device events
type CameraEventHandler interface {
	OnCameraAdded(device *CameraDevice)
	OnCameraRemoved(device *CameraDevice)
	OnCameraChanged(device *CameraDevice)
}

// EventNotifier provides a way to notify about camera events
type EventNotifier interface {
	NotifyCameraEvent(event CameraEventData)
}

// MonitorStats provides statistics about the camera monitor
type MonitorStats struct {
	TotalDevices     int           `json:"total_devices"`
	ConnectedDevices int           `json:"connected_devices"`
	LastScanTime     time.Time     `json:"last_scan_time"`
	ScanDuration     time.Duration `json:"scan_duration"`
	EventCount       int64         `json:"event_count"`
}

// CameraMonitor interface defines the contract for camera monitoring
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
}
