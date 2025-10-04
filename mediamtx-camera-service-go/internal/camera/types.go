package camera

import (
	"time"
)

// CameraDevice represents a discovered camera with detected capabilities.
// All fields are populated during capability probing except Error, which
// is set only when probing fails.
type CameraDevice struct {
	Path         string
	Name         string
	Capabilities V4L2Capabilities
	Formats      []V4L2Format
	Status       DeviceStatus
	LastSeen     time.Time
	DeviceNum    int
	Error        string `json:"error,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	Product      string `json:"product,omitempty"`
	Serial       string `json:"serial,omitempty"`
}

// DeviceStatus indicates the current operational state of a camera device.
type DeviceStatus string

const (
	DeviceStatusConnected    DeviceStatus = "CONNECTED"
	DeviceStatusDisconnected DeviceStatus = "DISCONNECTED"
	DeviceStatusError        DeviceStatus = "ERROR"
	DeviceStatusProbing      DeviceStatus = "PROBING"
)

// V4L2Capabilities contains device driver information and supported features.
type V4L2Capabilities struct {
	DriverName   string
	CardName     string
	BusInfo      string
	Version      string
	Capabilities []string
	DeviceCaps   []string
}

// V4L2Format describes a video format with resolution and frame rate support.
type V4L2Format struct {
	PixelFormat string
	Width       int
	Height      int
	FrameRates  []string
}

// DeviceCapabilityState tracks capability detection history and statistics.
type DeviceCapabilityState struct {
	LastProbeTime    time.Time
	ProbeCount       int
	SuccessCount     int
	FailureCount     int
	LastError        string                     `json:"last_error,omitempty"`
	CapabilityResult *CapabilityDetectionResult `json:"capability_result,omitempty"`
}

// DirectSnapshot contains metadata for a V4L2 direct capture operation.
type DirectSnapshot struct {
	ID          string
	DevicePath  string
	FilePath    string
	Size        int64
	Format      string
	Width       int `json:"width,omitempty"`
	Height      int `json:"height,omitempty"`
	CaptureTime time.Duration
	Created     time.Time
	Metadata    map[string]interface{}
}
