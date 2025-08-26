package camera

import (
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

// DeviceCapabilityState tracks capability detection state for a device
type DeviceCapabilityState struct {
	LastProbeTime    time.Time                  `json:"last_probe_time"`
	ProbeCount       int                        `json:"probe_count"`
	SuccessCount     int                        `json:"success_count"`
	FailureCount     int                        `json:"failure_count"`
	LastError        string                     `json:"last_error,omitempty"`
	CapabilityResult *CapabilityDetectionResult `json:"capability_result,omitempty"`
}
