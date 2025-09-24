/*
Camera Format Resolver

This component provides dynamic pixel format detection for camera devices,
leveraging the existing HybridCameraMonitor.selectOptimalPixelFormat method.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion

This resolver addresses the critical issue where FFmpeg commands use hardcoded
pixel formats without considering actual camera capabilities, causing failures
with modern HD cameras that support MJPEG natively.
*/

package mediamtx

import (
	"strings"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// CameraFormatResolver provides dynamic pixel format detection for camera devices
type CameraFormatResolver struct {
	cameraMonitor camera.CameraMonitor
	logger        *logging.Logger
}

// NewCameraFormatResolver creates a new camera format resolver
func NewCameraFormatResolver(cameraMonitor camera.CameraMonitor, logger *logging.Logger) *CameraFormatResolver {
	return &CameraFormatResolver{
		cameraMonitor: cameraMonitor,
		logger:        logger,
	}
}

// GetOptimalFormatForDevice returns the optimal pixel format for a camera device
// This method leverages the existing selectOptimalPixelFormat logic from HybridCameraMonitor
func (r *CameraFormatResolver) GetOptimalFormatForDevice(devicePath string) (string, error) {
	if r.cameraMonitor == nil {
		r.logger.Warn("Camera monitor not available, using fallback format")
		return r.getFallbackFormat(), nil
	}

	// Check if this is a V4L2 device
	if !strings.HasPrefix(devicePath, "/dev/video") {
		r.logger.WithField("device_path", devicePath).Debug("Non-V4L2 device, using fallback format")
		return r.getFallbackFormat(), nil
	}

	// Try to get device from camera monitor
	device, exists := r.cameraMonitor.GetDevice(devicePath)
	if !exists {
		r.logger.WithField("device_path", devicePath).Warn("Device not found in camera monitor, using fallback format")
		return r.getFallbackFormat(), nil
	}

	// Check if device has format information
	if len(device.Formats) == 0 {
		r.logger.WithField("device_path", devicePath).Warn("No format capabilities available, using fallback format")
		return r.getFallbackFormat(), nil
	}

	// Use the existing format selection logic
	// For recording, we prefer compressed formats for efficiency
	optimalFormat, err := r.selectOptimalFormatForRecording(devicePath, device)
	if err != nil {
		r.logger.WithFields(logging.Fields{
			"device_path": devicePath,
			"error":       err,
		}).Warn("Failed to select optimal format, using fallback")
		return r.getFallbackFormat(), nil
	}

	r.logger.WithFields(logging.Fields{
		"device_path":       devicePath,
		"selected_format":   optimalFormat,
		"available_formats": r.getAvailableFormats(device),
	}).Info("Selected optimal pixel format for recording")

	return optimalFormat, nil
}

// selectOptimalFormatForRecording selects the best format for recording operations
// This is adapted from the existing selectOptimalPixelFormat logic
func (r *CameraFormatResolver) selectOptimalFormatForRecording(devicePath string, device *camera.CameraDevice) (string, error) {
	// Define format preferences for recording (efficiency-focused)
	preferredFormats := []string{
		"MJPG", "MJPEG", "JPEG", // Compressed formats - most efficient for recording
		"YUYV", "UYVY", // Uncompressed formats - good fallback
		"RGB24", "BGR24", // RGB formats - universal compatibility
	}

	// Find the best supported format
	for _, preferred := range preferredFormats {
		for _, deviceFormat := range device.Formats {
			if strings.EqualFold(deviceFormat.PixelFormat, preferred) {
				return preferred, nil
			}
		}
	}

	// If no preferred format is found, use the first available format
	if len(device.Formats) > 0 {
		fallback := device.Formats[0].PixelFormat
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallback,
		}).Warn("Using first available format as fallback")
		return fallback, nil
	}

	// Ultimate fallback
	return r.getFallbackFormat(), nil
}

// getAvailableFormats returns a list of available formats for logging
func (r *CameraFormatResolver) getAvailableFormats(device *camera.CameraDevice) []string {
	formats := make([]string, 0, len(device.Formats))
	for _, format := range device.Formats {
		formats = append(formats, format.PixelFormat)
	}
	return formats
}

// getFallbackFormat returns a safe fallback format when capability detection fails
func (r *CameraFormatResolver) getFallbackFormat() string {
	// Use YUYV as fallback - widely supported by most cameras
	return "YUYV"
}

// GetOptimalFormatForDeviceWithFallback returns the optimal format with a specific fallback
// This method allows callers to specify a fallback format from configuration
func (r *CameraFormatResolver) GetOptimalFormatForDeviceWithFallback(devicePath, fallbackFormat string) (string, error) {
	if r.cameraMonitor == nil {
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallbackFormat,
		}).Warn("Camera monitor not available, using configured fallback format")
		return fallbackFormat, nil
	}

	// Check if this is a V4L2 device
	if !strings.HasPrefix(devicePath, "/dev/video") {
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallbackFormat,
		}).Debug("Non-V4L2 device, using configured fallback format")
		return fallbackFormat, nil
	}

	// Try to get device from camera monitor
	device, exists := r.cameraMonitor.GetDevice(devicePath)
	if !exists {
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallbackFormat,
		}).Warn("Device not found in camera monitor, using configured fallback format")
		return fallbackFormat, nil
	}

	// Check if device has format information
	if len(device.Formats) == 0 {
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallbackFormat,
		}).Warn("No format capabilities available, using configured fallback format")
		return fallbackFormat, nil
	}

	// Use the existing format selection logic
	optimalFormat, err := r.selectOptimalFormatForRecording(devicePath, device)
	if err != nil {
		r.logger.WithFields(logging.Fields{
			"device_path":     devicePath,
			"fallback_format": fallbackFormat,
			"error":           err,
		}).Warn("Failed to select optimal format, using configured fallback format")
		return fallbackFormat, nil
	}

	r.logger.WithFields(logging.Fields{
		"device_path":       devicePath,
		"selected_format":   optimalFormat,
		"available_formats": r.getAvailableFormats(device),
	}).Info("Selected optimal pixel format for recording")

	return optimalFormat, nil
}
