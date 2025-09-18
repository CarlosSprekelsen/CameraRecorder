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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
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

// GenerateRecordingPath generates the MediaMTX recordPath pattern
// This returns the PATTERN for MediaMTX to use, not an actual file path
// MediaMTX will replace %path, %Y, %m, %d, etc. and add the extension
func GenerateRecordingPath(cfg *config.MediaMTXConfig, recordingCfg *config.RecordingConfig) string {
	basePath := cfg.RecordingsPath
	pattern := recordingCfg.FileNamePattern

	if recordingCfg.UseDeviceSubdirs {
		// MediaMTX will replace %path with the actual path name (e.g., camera0)
		return filepath.Join(basePath, "%path", pattern)
	}

	return filepath.Join(basePath, pattern)
}

// GenerateSnapshotPath generates an actual file path for snapshots
// Unlike recordings, snapshots are created directly by FFmpeg, not MediaMTX
func GenerateSnapshotPath(cfg *config.MediaMTXConfig, snapshotCfg *config.SnapshotConfig, devicePath string) string {
	basePath := cfg.SnapshotsPath
	deviceName := GetMediaMTXPathName(devicePath) // e.g., "camera0"

	// Create device subdirectory if configured
	if snapshotCfg.UseDeviceSubdirs {
		basePath = filepath.Join(basePath, deviceName)
	}

	// Generate filename from pattern
	filename := expandSnapshotPattern(snapshotCfg.FileNamePattern, deviceName)

	return filepath.Join(basePath, filename)
}

// expandSnapshotPattern expands pattern variables for snapshots
func expandSnapshotPattern(pattern string, deviceName string) string {
	now := time.Now()
	result := pattern

	// Replace device name
	result = strings.ReplaceAll(result, "%device", deviceName)

	// Replace timestamp
	result = strings.ReplaceAll(result, "%timestamp", fmt.Sprintf("%d", now.Unix()))

	// Replace date/time components if needed
	result = strings.ReplaceAll(result, "%Y", fmt.Sprintf("%04d", now.Year()))
	result = strings.ReplaceAll(result, "%m", fmt.Sprintf("%02d", now.Month()))
	result = strings.ReplaceAll(result, "%d", fmt.Sprintf("%02d", now.Day()))
	result = strings.ReplaceAll(result, "%H", fmt.Sprintf("%02d", now.Hour()))
	result = strings.ReplaceAll(result, "%M", fmt.Sprintf("%02d", now.Minute()))
	result = strings.ReplaceAll(result, "%S", fmt.Sprintf("%02d", now.Second()))

	return result
}

// GetRecordingFilePath constructs the expected file path for a recording
// This is used to verify where MediaMTX actually wrote the file
func GetRecordingFilePath(cfg *config.MediaMTXConfig, recordingCfg *config.RecordingConfig, pathName string, startTime time.Time) string {
	basePath := cfg.RecordingsPath

	if recordingCfg.UseDeviceSubdirs {
		basePath = filepath.Join(basePath, pathName)
	}

	// Construct filename as MediaMTX would create it
	filename := fmt.Sprintf("%s_%04d-%02d-%02d_%02d-%02d-%02d",
		pathName,
		startTime.Year(), startTime.Month(), startTime.Day(),
		startTime.Hour(), startTime.Minute(), startTime.Second())

	// Add extension based on format
	if recordingCfg.RecordFormat == "fmp4" {
		filename += ".mp4"
	} else if recordingCfg.RecordFormat == "mpegts" {
		filename += ".ts"
	}

	return filepath.Join(basePath, filename)
}

// BuildFFmpegCommand builds a proper FFmpeg command based on device type and configuration
// This centralizes FFmpeg command generation to prevent hardcoded echo commands
func BuildFFmpegCommand(devicePath, streamName string, cfg *config.MediaMTXConfig) string {
	// Detect device type
	if strings.HasPrefix(devicePath, "/dev/video") {
		// V4L2 device - build comprehensive FFmpeg command using codec config
		return fmt.Sprintf(
			"ffmpeg -f v4l2 -i %s -c:v libx264 -profile:v %s -level %s "+
				"-pix_fmt %s -preset %s -b:v %s -f rtsp rtsp://%s:%d/%s",
			devicePath,
			cfg.Codec.VideoProfile,
			cfg.Codec.VideoLevel,
			cfg.Codec.PixelFormat,
			cfg.Codec.Preset,
			cfg.Codec.Bitrate,
			cfg.Host, cfg.RTSPPort, streamName)
	} else if strings.HasPrefix(devicePath, "rtsp://") {
		// External RTSP source - use relay/proxy command
		return fmt.Sprintf(
			"ffmpeg -i %s -c copy -f rtsp rtsp://%s:%d/%s",
			devicePath, cfg.Host, cfg.RTSPPort, streamName)
	}

	// Fallback for unknown device types
	return fmt.Sprintf(
		"ffmpeg -i %s -c:v libx264 -preset %s -f rtsp rtsp://%s:%d/%s",
		devicePath, cfg.Codec.Preset, cfg.Host, cfg.RTSPPort, streamName)
}

// ParseSnapshotFilename parses a snapshot filename using the configured pattern
// This is the inverse of expandSnapshotPattern
func ParseSnapshotFilename(filename, pattern string) (device string, timestamp time.Time, err error) {
	// Handle the common pattern: %device_%timestamp.jpg
	if pattern == "%device_%timestamp.jpg" {
		// Remove extension first
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		// Split by underscore
		parts := strings.Split(nameWithoutExt, "_")
		if len(parts) < 2 {
			return "", time.Time{}, fmt.Errorf("filename %s does not match pattern %s", filename, pattern)
		}

		// Extract device (first part)
		device = parts[0]

		// Extract timestamp (last part)
		timestampStr := parts[len(parts)-1]
		timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("failed to parse timestamp from %s: %w", timestampStr, err)
		}
		timestamp = time.Unix(timestampInt, 0)

		return device, timestamp, nil
	}

	// For more complex patterns, we could implement regex-based parsing
	// For now, fallback to basic parsing
	device = "camera0" // Default
	if parts := strings.Split(filename, "_"); len(parts) > 0 {
		if strings.HasPrefix(parts[0], "camera") {
			device = parts[0]
		}
	}

	// Try to extract timestamp from filename if possible
	timestamp = time.Now() // Fallback to current time

	return device, timestamp, nil
}
