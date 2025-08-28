package camera

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

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
	cmd.Args = append(cmd.Args, strings.Fields(args)...)

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
	var currentPixelFormat string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for new format entry: [0]: 'YUYV' (YUYV 4:2:2)
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]:") && strings.Contains(line, "'") {
			// New format entry - save the pixel format for subsequent sizes
			if strings.Contains(line, "'") {
				start := strings.Index(line, "'") + 1
				end := strings.LastIndex(line, "'")
				if start > 0 && end > start {
					currentPixelFormat = line[start:end]
				}
			}
		} else if currentPixelFormat != "" {
			if strings.Contains(line, "Size:") {
				// Create a new format entry for each size
				if currentFormat != nil {
					formats = append(formats, *currentFormat)
				}
				
				currentFormat = &V4L2Format{
					PixelFormat: currentPixelFormat,
					FrameRates:  []string{},
				}
				
				// Extract size from "Size: Discrete 640x480"
				size := r.extractValue(line)
				// Remove "Discrete " prefix if present
				if strings.HasPrefix(size, "Discrete ") {
					size = strings.TrimPrefix(size, "Discrete ")
				}
				width, height := r.parseSize(size)
				currentFormat.Width = width
				currentFormat.Height = height
			} else if currentFormat != nil && strings.Contains(line, "Interval:") && strings.Contains(line, "fps") {
				// Extract fps from interval line: Interval: Discrete 0.033s (30.000 fps)
				if strings.Contains(line, "(") && strings.Contains(line, "fps") {
					start := strings.Index(line, "(") + 1
					end := strings.Index(line, "fps")
					if start > 0 && end > start {
						fps := strings.TrimSpace(line[start:end])
						if fps != "" {
							currentFormat.FrameRates = append(currentFormat.FrameRates, fps)
						}
					}
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

func (r *RealDeviceInfoParser) ParseDeviceFrameRates(output string) ([]string, error) {
	var frameRates []string

	// Enhanced frame rate patterns for robust parsing (following Python patterns)
	frameRatePatterns := []string{
		`(?m)^\s*(\d+(?:\.\d+)?)\s*fps\b`,            // 30.000 fps
		`(?m)^\s*(\d+(?:\.\d+)?)\s*FPS\b`,            // 30.000 FPS
		`Frame\s*rate[:\s]+(\d+(?:\.\d+)?)`,          // Frame rate: 30.0
		`(?m)^\s*(\d+(?:\.\d+)?)\s*Hz\b`,             // 30 Hz
		`@(\d+(?:\.\d+)?)\b`,                         // 1920x1080@60
		`Interval:\s*\[1/(\d+(?:\.\d+)?)\]`,          // Interval: [1/30]
		`\[1/(\d+(?:\.\d+)?)\]`,                      // [1/30]
		`1/(\d+(?:\.\d+)?)\s*s`,                      // 1/30 s
		`(\d+(?:\.\d+)?)\s*frame[s]?\s*per\s*second`, // 30 frames per second
		`rate:\s*(\d+(?:\.\d+)?)`,                    // rate: 30
		`fps:\s*(\d+(?:\.\d+)?)`,                     // fps: 30
	}

	seenRates := make(map[string]bool)

	for _, pattern := range frameRatePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(output, -1)

		for _, match := range matches {
			if len(match) > 1 {
				rate := strings.TrimSpace(match[1])
				if rate != "" && !seenRates[rate] {
					frameRates = append(frameRates, rate)
					seenRates[rate] = true
				}
			}
		}
	}

	return frameRates, nil
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
