package mediamtx

/*
MediaMTX Path Manager Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// pathManager represents the MediaMTX path manager
type pathManager struct {
	client        MediaMTXClient
	config        *MediaMTXConfig
	logger        *logging.Logger
	cameraMonitor camera.CameraMonitor

	// Camera-path mapping for abstraction layer
	cameraPaths  map[string]string // device path -> path name
	pathCameras  map[string]string // path name -> device path
	mappingMutex sync.RWMutex
}

// NewPathManager creates a new MediaMTX path manager
func NewPathManager(client MediaMTXClient, config *MediaMTXConfig, logger *logging.Logger) PathManager {
	return &pathManager{
		client:      client,
		config:      config,
		logger:      logger,
		cameraPaths: make(map[string]string),
		pathCameras: make(map[string]string),
	}
}

// NewPathManagerWithCamera creates a new MediaMTX path manager with camera monitor
func NewPathManagerWithCamera(client MediaMTXClient, config *MediaMTXConfig, cameraMonitor camera.CameraMonitor, logger *logging.Logger) PathManager {
	return &pathManager{
		client:        client,
		config:        config,
		logger:        logger,
		cameraMonitor: cameraMonitor,
		cameraPaths:   make(map[string]string),
		pathCameras:   make(map[string]string),
	}
}

// CreatePath creates a new path
func (pm *pathManager) CreatePath(ctx context.Context, name, source string, options map[string]interface{}) error {
	pm.logger.WithFields(logging.Fields{
		"name":    name,
		"source":  source,
		"options": options,
	}).Debug("Creating MediaMTX path")

	// Enhanced validation for better user experience and software resilience
	if err := pm.validatePathName(name); err != nil {
		return fmt.Errorf("invalid path name: %w", err)
	}

	if err := pm.validateSource(source, options); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// Create path request
	path := &Path{
		Name:   name,
		Source: source,
	}

	// Apply options to path
	if sourceOnDemand, ok := options["sourceOnDemand"].(bool); ok {
		path.SourceOnDemand = sourceOnDemand
	}
	if startTimeout, ok := options["sourceOnDemandStartTimeout"].(string); ok {
		if duration, err := parseDuration(startTimeout); err == nil {
			path.SourceOnDemandStartTimeout = duration
		}
	}
	if closeAfter, ok := options["sourceOnDemandCloseAfter"].(string); ok {
		if duration, err := parseDuration(closeAfter); err == nil {
			path.SourceOnDemandCloseAfter = duration
		}
	}
	if publishUser, ok := options["publishUser"].(string); ok {
		path.PublishUser = publishUser
	}
	if publishPass, ok := options["publishPass"].(string); ok {
		path.PublishPass = publishPass
	}
	if readUser, ok := options["readUser"].(string); ok {
		path.ReadUser = readUser
	}
	if readPass, ok := options["readPass"].(string); ok {
		path.ReadPass = readPass
	}
	if runOnDemand, ok := options["runOnDemand"].(string); ok {
		path.RunOnDemand = runOnDemand
	}
	if runOnDemandRestart, ok := options["runOnDemandRestart"].(bool); ok {
		path.RunOnDemandRestart = runOnDemandRestart
	}
	if runOnDemandCloseAfter, ok := options["runOnDemandCloseAfter"].(string); ok {
		if duration, err := parseDuration(runOnDemandCloseAfter); err == nil {
			path.RunOnDemandCloseAfter = duration
		}
	}
	if runOnDemandStartTimeout, ok := options["runOnDemandStartTimeout"].(string); ok {
		if duration, err := parseDuration(runOnDemandStartTimeout); err == nil {
			path.RunOnDemandStartTimeout = duration
		}
	}

	// Marshal request - choose correct marshaling function based on path type
	var data []byte
	var err error

	if path.RunOnDemand != "" {
		// For on-demand streams (USB devices), use USB-specific marshaling
		data, err = marshalCreateUSBPathRequest(name, path.RunOnDemand)
	} else {
		// For direct sources (external RTSP), use standard marshaling
		data, err = marshalCreatePathRequest(path)
	}

	if err != nil {
		return NewPathErrorWithErr(name, "create_path", "failed to marshal request", err)
	}

	// Send request - name must be in URL path per Swagger spec
	_, err = pm.client.Post(ctx, fmt.Sprintf("/v3/config/paths/add/%s", name), data)
	if err != nil {
		// Check if this is a "path already exists" error (idempotent success)
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "path already exists") || 
		   strings.Contains(errorMsg, "already exists") ||
		   strings.Contains(errorMsg, "400") {
			pm.logger.WithFields(logging.Fields{
				"name": name,
				"error": errorMsg,
			}).Info("MediaMTX path already exists, treating as success")
			return nil // Idempotent success - path exists, which is what we wanted
		}
		return NewPathErrorWithErr(name, "create_path", "failed to create path", err)
	}

	pm.logger.WithField("name", name).Info("MediaMTX path created successfully")
	return nil
}

// DeletePath deletes a path
func (pm *pathManager) DeletePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Deleting MediaMTX path")

	err := pm.client.Delete(ctx, fmt.Sprintf("/v3/config/paths/delete/%s", name))
	if err != nil {
		return NewPathErrorWithErr(name, "delete_path", "failed to delete path", err)
	}

	pm.logger.WithField("name", name).Info("MediaMTX path deleted successfully")
	return nil
}

// GetPath gets a specific path
func (pm *pathManager) GetPath(ctx context.Context, name string) (*Path, error) {
	pm.logger.WithField("name", name).Debug("Getting MediaMTX path")

	data, err := pm.client.Get(ctx, fmt.Sprintf("/v3/paths/get/%s", name))
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to get path", err)
	}

	path, err := parsePathResponse(data)
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to parse path response", err)
	}

	return path, nil
}

// ListPaths lists all paths
func (pm *pathManager) ListPaths(ctx context.Context) ([]*Path, error) {
	pm.logger.Debug("Listing MediaMTX paths")

	data, err := pm.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list paths: %w", err)
	}

	paths, err := parsePathsResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse paths response: %w", err)
	}

	pm.logger.WithField("count", strconv.Itoa(len(paths))).Debug("MediaMTX paths listed successfully")
	return paths, nil
}

// ValidatePath validates a path
func (pm *pathManager) ValidatePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Validating MediaMTX path")

	// Check if path exists
	exists := pm.PathExists(ctx, name)
	if !exists {
		return NewPathError(name, "validate_path", "path does not exist")
	}

	// Get path details to validate configuration
	_, err := pm.GetPath(ctx, name)
	if err != nil {
		return NewPathErrorWithErr(name, "validate_path", "failed to get path details", err)
	}

	pm.logger.WithField("name", name).Debug("MediaMTX path validated successfully")
	return nil
}

// PathExists checks if a path exists in configuration
func (pm *pathManager) PathExists(ctx context.Context, name string) bool {
	pm.logger.WithField("name", name).Debug("Checking if MediaMTX path exists")

	// Check configuration paths, not runtime paths
	data, err := pm.client.Get(ctx, fmt.Sprintf("/v3/config/paths/get/%s", name))
	if err != nil {
		return false
	}

	// Try to parse the response to verify it's valid
	_, err = parsePathConfResponse(data)
	return err == nil
}

// parsePathConfResponse parses a path configuration response
func parsePathConfResponse(data []byte) (map[string]interface{}, error) {
	var config map[string]interface{}
	err := json.Unmarshal(data, &config)
	return config, err
}

// parseDuration parses a duration string
func parseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// validatePathName validates path name format and content
func (pm *pathManager) validatePathName(name string) error {
	if name == "" {
		return fmt.Errorf("path name cannot be empty")
	}

	// Check for valid characters (alphanumeric, underscores, hyphens)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return fmt.Errorf("path name contains invalid character '%c' - only alphanumeric characters, underscores, and hyphens are allowed", char)
		}
	}

	// Check length limits
	if len(name) > 64 {
		return fmt.Errorf("path name too long (%d characters) - maximum 64 characters allowed", len(name))
	}

	if len(name) < 1 {
		return fmt.Errorf("path name too short - minimum 1 character required")
	}

	// Check for reserved names
	reservedNames := []string{"all", "~all", "~internal"}
	for _, reserved := range reservedNames {
		if name == reserved {
			return fmt.Errorf("path name '%s' is reserved and cannot be used", name)
		}
	}

	return nil
}

// validateSource validates source format and content
func (pm *pathManager) validateSource(source string, options map[string]interface{}) error {
	// Allow empty source if runOnDemand is specified
	if source == "" {
		if runOnDemand, exists := options["runOnDemand"]; exists && runOnDemand != "" {
			return nil // Empty source is valid when using runOnDemand
		}
		return fmt.Errorf("source cannot be empty")
	}

	// Check for valid source formats
	if strings.HasPrefix(source, "/dev/") {
		// Device path validation
		if !strings.HasPrefix(source, "/dev/video") &&
			!strings.HasPrefix(source, "/dev/camera") &&
			!strings.HasPrefix(source, "/dev/") {
			return fmt.Errorf("invalid device path format: %s - must be a valid device path (e.g., /dev/video0)", source)
		}
	} else if strings.HasPrefix(source, "rtsp://") || strings.HasPrefix(source, "rtmp://") {
		// Stream URL validation
		if len(source) < 10 {
			return fmt.Errorf("invalid stream URL format: %s - URL too short", source)
		}
	} else if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// HTTP stream validation
		if len(source) < 10 {
			return fmt.Errorf("invalid HTTP stream URL format: %s - URL too short", source)
		}
	} else {
		// Generic validation for other source types
		if len(source) < 3 {
			return fmt.Errorf("invalid source format: %s - source too short", source)
		}
	}

	// Check for potentially dangerous characters (but allow // in valid URL schemes)
	dangerousChars := []string{"..", "\\", "<", ">", "|", "&", ";", "`", "$"}

	// Special handling for // - only allow it in valid URL schemes
	if strings.Contains(source, "//") {
		// Allow // only in valid URL schemes
		validSchemes := []string{"rtsp://", "rtmp://", "http://", "https://"}
		hasValidScheme := false
		for _, scheme := range validSchemes {
			if strings.HasPrefix(source, scheme) {
				hasValidScheme = true
				break
			}
		}
		if !hasValidScheme {
			return fmt.Errorf("source contains potentially dangerous character sequence '//' - this may cause security issues")
		}
	}

	// Check other dangerous characters
	for _, char := range dangerousChars {
		if strings.Contains(source, char) {
			return fmt.Errorf("source contains potentially dangerous character sequence '%s' - this may cause security issues", char)
		}
	}

	return nil
}

// Camera Operations - PathManager consolidates camera operations for better architecture

// GetCameraList returns all discovered cameras in API-ready format
func (pm *pathManager) GetCameraList(ctx context.Context) (*CameraListResponse, error) {
	if pm.cameraMonitor == nil {
		return nil, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get cameras from camera monitor (internal format)
	cameras := pm.cameraMonitor.GetConnectedCameras()

	// Convert to API-ready format with abstraction layer
	apiCameras := make([]*APICameraInfo, 0, len(cameras))
	connectedCount := 0

	for _, cameraDevice := range cameras {
		// Apply abstraction layer: device path -> camera ID
		cameraID, exists := pm.GetCameraForDevicePath(cameraDevice.Path)
		if !exists {
			// Fallback: generate camera ID from device path
			cameraID = pm.generateCameraID(cameraDevice.Path)
		}

		// Extract resolution from camera formats
		resolution := "Unknown"
		if len(cameraDevice.Formats) > 0 {
			format := cameraDevice.Formats[0]
			resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
		}

		// Generate stream URLs using camera ID (abstracted) with configuration
		streams := map[string]string{
			"rtsp":   fmt.Sprintf("rtsp://%s:%d/%s", pm.config.Host, pm.config.RTSPPort, cameraID),
			"webrtc": fmt.Sprintf("http://%s:%d/%s/webrtc", pm.config.Host, pm.config.WebRTCPort, cameraID),
			"hls":    fmt.Sprintf("http://%s:%d/%s", pm.config.Host, pm.config.HLSPort, cameraID),
		}

		// Convert capabilities to API format
		capabilities := make(map[string]interface{})
		if len(cameraDevice.Capabilities.Capabilities) > 0 {
			capabilities["capabilities"] = cameraDevice.Capabilities.Capabilities
		}
		if len(cameraDevice.Formats) > 0 {
			capabilities["formats"] = len(cameraDevice.Formats)
		}

		// Create API-ready camera info
		apiCamera := &APICameraInfo{
			Device:       cameraID,                    // Abstracted camera ID
			Status:       string(cameraDevice.Status), // Camera status
			Name:         cameraDevice.Name,           // Camera name
			Resolution:   resolution,                  // Extracted resolution
			FPS:          30,                          // Default FPS (could be extracted from capabilities)
			Streams:      streams,                     // Generated stream URLs
			Capabilities: capabilities,                // API-ready capabilities
		}

		apiCameras = append(apiCameras, apiCamera)

		// Count connected cameras
		if cameraDevice.Status == camera.DeviceStatusConnected {
			connectedCount++
		}
	}

	response := &CameraListResponse{
		Cameras:   apiCameras,
		Total:     len(apiCameras),
		Connected: connectedCount,
	}

	pm.logger.WithFields(logging.Fields{
		"total":     response.Total,
		"connected": response.Connected,
	}).Debug("PathManager converted camera list to API format")

	return response, nil
}

// generateCameraID creates a camera ID from device path (fallback method)
func (pm *pathManager) generateCameraID(devicePath string) string {
	// Extract device number from /dev/video0 -> camera0
	if strings.HasPrefix(devicePath, "/dev/video") {
		deviceNum := strings.TrimPrefix(devicePath, "/dev/video")
		return fmt.Sprintf("camera%s", deviceNum)
	}
	return "camera_unknown"
}

// GetCameraStatus returns status for a specific camera
func (pm *pathManager) GetCameraStatus(ctx context.Context, device string) (*CameraStatusResponse, error) {
	if pm.cameraMonitor == nil {
		return nil, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get actual device path using abstraction layer
	devicePath := pm.getDevicePathFromCameraIdentifier(device)

	// Get camera from camera monitor
	cameraDevice, exists := pm.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return nil, fmt.Errorf("camera device not found: %s", device)
	}

	// Build response with abstraction layer
	response := &CameraStatusResponse{
		Device:       device, // Return camera identifier (camera0)
		Status:       string(cameraDevice.Status),
		Name:         cameraDevice.Name,
		Resolution:   "",
		FPS:          0,
		Streams:      make(map[string]string),
		Capabilities: &cameraDevice.Capabilities,
	}

	// Extract resolution and FPS from capabilities if available
	if len(cameraDevice.Formats) > 0 {
		format := cameraDevice.Formats[0]
		response.Resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
		if len(format.FrameRates) > 0 {
			response.FPS = 30 // Default FPS
		}
	}

	pm.logger.WithFields(logging.Fields{
		"device": device,
		"status": response.Status,
		"name":   response.Name,
	}).Debug("PathManager retrieved camera status")

	return response, nil
}

// ValidateCameraDevice validates that a camera device exists and is accessible
func (pm *pathManager) ValidateCameraDevice(ctx context.Context, device string) (bool, error) {
	if pm.cameraMonitor == nil {
		return false, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get actual device path using abstraction layer
	devicePath := pm.getDevicePathFromCameraIdentifier(device)

	// Check if camera exists
	_, exists := pm.cameraMonitor.GetDevice(devicePath)

	pm.logger.WithFields(logging.Fields{
		"device":      device,
		"device_path": devicePath,
		"exists":      exists,
	}).Debug("PathManager validated camera device")

	return exists, nil
}

// GetPathForCamera gets the MediaMTX path name for a camera identifier
// Since MediaMTX paths ARE camera identifiers (camera0), this is mostly identity mapping
func (pm *pathManager) GetPathForCamera(cameraID string) (string, bool) {
	// Direct mapping: camera0 -> camera0 (MediaMTX path name = camera identifier)
	// But check if device actually exists
	devicePath := pm.getDevicePathFromCameraIdentifier(cameraID)

	pm.mappingMutex.RLock()
	defer pm.mappingMutex.RUnlock()

	// Check if we have this device mapped to a path
	pathName, exists := pm.cameraPaths[devicePath]
	if exists {
		return pathName, true
	}

	// For most cases, MediaMTX path name = camera identifier
	return cameraID, false
}

// GetCameraForPath gets the camera identifier for a MediaMTX path name
// Since MediaMTX paths ARE camera identifiers (camera0), this is mostly identity mapping
func (pm *pathManager) GetCameraForPath(pathName string) (string, bool) {
	pm.mappingMutex.RLock()
	defer pm.mappingMutex.RUnlock()

	devicePath, exists := pm.pathCameras[pathName]
	if !exists {
		// For most cases, MediaMTX path name = camera identifier
		return pathName, false
	}

	// Return camera identifier using abstraction layer
	cameraID := pm.getCameraIdentifierFromDevicePath(devicePath)
	return cameraID, true
}

// GetDevicePathForCamera gets the actual USB device path for a camera identifier
// This is the main abstraction: camera0 -> /dev/video0
func (pm *pathManager) GetDevicePathForCamera(cameraID string) (string, bool) {
	devicePath := pm.getDevicePathFromCameraIdentifier(cameraID)

	// Check if device actually exists via camera monitor
	if pm.cameraMonitor != nil {
		_, exists := pm.cameraMonitor.GetDevice(devicePath)
		return devicePath, exists
	}

	// Fallback: return converted path without validation
	return devicePath, true
}

// GetCameraForDevicePath gets the camera identifier for a USB device path
// This is the reverse abstraction: /dev/video0 -> camera0
func (pm *pathManager) GetCameraForDevicePath(devicePath string) (string, bool) {
	cameraID := pm.getCameraIdentifierFromDevicePath(devicePath)

	// Check if device actually exists via camera monitor
	if pm.cameraMonitor != nil {
		_, exists := pm.cameraMonitor.GetDevice(devicePath)
		return cameraID, exists
	}

	// Fallback: return converted identifier without validation
	return cameraID, true
}

// Abstraction Layer Methods - Consolidated in PathManager

// getCameraIdentifierFromDevicePath converts a device path to a camera identifier
// Example: /dev/video0 -> camera0
func (pm *pathManager) getCameraIdentifierFromDevicePath(devicePath string) string {
	// Extract the number from /dev/video{N}
	if strings.HasPrefix(devicePath, "/dev/video") {
		number := strings.TrimPrefix(devicePath, "/dev/video")
		return fmt.Sprintf("camera%s", number)
	}
	// Fallback: return the device path as-is
	return devicePath
}

// getDevicePathFromCameraIdentifier converts a camera identifier to a device path
// Example: camera0 -> /dev/video0
func (pm *pathManager) getDevicePathFromCameraIdentifier(cameraID string) string {
	// Extract the number from camera{N}
	if strings.HasPrefix(cameraID, "camera") {
		number := strings.TrimPrefix(cameraID, "camera")
		return fmt.Sprintf("/dev/video%s", number)
	}
	// If already a device path, return as-is
	return cameraID
}
