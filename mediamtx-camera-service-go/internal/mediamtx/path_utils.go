/*
MediaMTX Path Utilities - Unified Path Name Generation

This module provides centralized path name generation for MediaMTX paths,
ensuring consistency across all components that interact with MediaMTX.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion

API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"fmt"
	"strings"
)

// GetMediaMTXPathName generates consistent path names for MediaMTX
// This is the single source of truth for path naming across all components
//
// Examples:
//
//	GetMediaMTXPathName("/dev/video0") -> "camera0"
//	GetMediaMTXPathName("/dev/video1") -> "camera1"
//	GetMediaMTXPathName("/dev/video2") -> "camera2"
//
// This function ensures that both StreamManager and RecordingManager
// use the same path names, enabling path reuse as documented in the
// MediaMTX Swagger API specification.
func GetMediaMTXPathName(devicePath string) string {
	if devicePath == "" {
		return ""
	}

	// Extract device name from path
	parts := strings.Split(devicePath, "/")
	deviceName := parts[len(parts)-1]

	// Convert V4L2 device names to camera identifiers
	if strings.HasPrefix(deviceName, "video") {
		deviceNum := deviceName[5:] // Remove "video" prefix
		return fmt.Sprintf("camera%s", deviceNum)
	}

	// For non-V4L2 devices, use the device name as-is
	return deviceName
}

// ValidatePathName validates that a path name conforms to MediaMTX requirements
// Based on MediaMTX Swagger API specification
func ValidatePathName(pathName string) error {
	if pathName == "" {
		return fmt.Errorf("path name cannot be empty")
	}

	// MediaMTX path names should be simple identifiers
	// Avoid complex characters that might cause issues
	if strings.Contains(pathName, " ") {
		return fmt.Errorf("path name cannot contain spaces: %s", pathName)
	}

	if strings.Contains(pathName, "/") {
		return fmt.Errorf("path name cannot contain forward slashes: %s", pathName)
	}

	return nil
}

// GetDevicePathFromCameraIdentifier converts camera identifier back to device path
// This is the inverse operation of GetMediaMTXPathName
//
// Examples:
//
//	GetDevicePathFromCameraIdentifier("camera0") -> "/dev/video0"
//	GetDevicePathFromCameraIdentifier("camera1") -> "/dev/video1"
func GetDevicePathFromCameraIdentifier(cameraID string) string {
	if cameraID == "" {
		return ""
	}

	// Handle camera identifiers (camera0, camera1, etc.)
	if strings.HasPrefix(cameraID, "camera") {
		deviceNum := cameraID[6:] // Remove "camera" prefix
		return fmt.Sprintf("/dev/video%s", deviceNum)
	}

	// For non-camera identifiers, assume it's already a device path
	return cameraID
}
