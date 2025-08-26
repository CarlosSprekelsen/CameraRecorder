package camera

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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
}

// ConfigProvider interface for configuration access
type ConfigProvider interface {
	GetCameraConfig() *CameraConfig
	GetPollInterval() float64
	GetDetectionTimeout() float64
	GetDeviceRange() []int
	GetEnableCapabilityDetection() bool
	GetCapabilityTimeout() float64
}

// Logger interface for structured logging
type Logger interface {
	WithFields(fields map[string]interface{}) Logger
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
}

// RealDeviceChecker implements DeviceChecker for real file system
type RealDeviceChecker struct{}

func (r *RealDeviceChecker) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// RealV4L2CommandExecutor implements V4L2CommandExecutor for real V4L2 commands
type RealV4L2CommandExecutor struct{}

func (r *RealV4L2CommandExecutor) ExecuteCommand(ctx context.Context, devicePath, args string) (string, error) {
	cmd := exec.CommandContext(ctx, "v4l2-ctl", "--device", devicePath)
	cmd.Args = append(cmd.Args, args)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// RealDeviceInfoParser implements DeviceInfoParser for real V4L2 output parsing
type RealDeviceInfoParser struct{}

func (r *RealDeviceInfoParser) ParseDeviceInfo(output string) (V4L2Capabilities, error) {
	capabilities := V4L2Capabilities{
		Capabilities: []string{},
		DeviceCaps:   []string{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "Driver name") {
			capabilities.DriverName = r.extractValue(line)
		} else if strings.HasPrefix(line, "Card type") || strings.HasPrefix(line, "Device name") {
			capabilities.CardName = r.extractValue(line)
		} else if strings.HasPrefix(line, "Bus info") {
			capabilities.BusInfo = r.extractValue(line)
		} else if strings.HasPrefix(line, "Driver version") {
			capabilities.Version = r.extractValue(line)
		} else if strings.Contains(line, "Capabilities") {
			caps := r.parseCapabilities(line)
			capabilities.Capabilities = append(capabilities.Capabilities, caps...)
		} else if strings.Contains(line, "Device Caps") {
			caps := r.parseCapabilities(line)
			capabilities.DeviceCaps = append(capabilities.DeviceCaps, caps...)
		}
	}

	// Set defaults if not found
	if capabilities.CardName == "" {
		capabilities.CardName = "Unknown Video Device"
	}
	if capabilities.DriverName == "" {
		capabilities.DriverName = "unknown"
	}

	return capabilities, nil
}

func (r *RealDeviceInfoParser) ParseDeviceFormats(output string) ([]V4L2Format, error) {
	var formats []V4L2Format
	
	lines := strings.Split(output, "\n")
	var currentFormat *V4L2Format
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "Index") && strings.Contains(line, "Type") {
			// New format entry
			if currentFormat != nil {
				formats = append(formats, *currentFormat)
			}
			currentFormat = &V4L2Format{
				FrameRates: []string{},
			}
		} else if currentFormat != nil {
			if strings.Contains(line, "Name") {
				currentFormat.PixelFormat = r.extractValue(line)
			} else if strings.Contains(line, "Size") {
				size := r.extractValue(line)
				width, height := r.parseSize(size)
				currentFormat.Width = width
				currentFormat.Height = height
			} else if strings.Contains(line, "fps") {
				fps := r.extractValue(line)
				if fps != "" {
					currentFormat.FrameRates = append(currentFormat.FrameRates, fps)
				}
			}
		}
	}
	
	// Add the last format
	if currentFormat != nil {
		formats = append(formats, *currentFormat)
	}

	return formats, nil
}

// Helper methods for RealDeviceInfoParser
func (r *RealDeviceInfoParser) extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (r *RealDeviceInfoParser) parseCapabilities(line string) []string {
	var capabilities []string
	
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return capabilities
	}
	
	caps := strings.Fields(parts[1])
	for _, cap := range caps {
		cap = strings.TrimSpace(cap)
		if cap != "" {
			capabilities = append(capabilities, cap)
		}
	}
	
	return capabilities
}

func (r *RealDeviceInfoParser) parseSize(size string) (int, int) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	
	width, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	
	return width, height
}
