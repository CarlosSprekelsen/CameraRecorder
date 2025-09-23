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
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"golang.org/x/sync/singleflight"
)

// pathManager manages MediaMTX path lifecycle and device abstraction services.
//
// RESPONSIBILITIES:
// - MediaMTX API integration using api_types.go
// - Device abstraction layer and mapping services
// - Camera capability extraction and API formatting
// - Path lifecycle management for both V4L2 and external streams
//
// ABSTRACTION SERVICES:
// - Provides cameraID to devicePath mapping services
// - Acts as middleware between Controller and hardware layers
// - Bridges CameraMonitor and JSON-RPC API requirements
//
// API INTEGRATION:
// - Returns JSON-RPC API-ready responses
// - Uses MediaMTX api_types.go for all operations
type pathManager struct {
	client        MediaMTXClient
	config        *config.MediaMTXConfig
	logger        *logging.Logger
	cameraMonitor camera.CameraMonitor

	// Camera-path mapping for abstraction layer
	cameraPaths  map[string]string // device path -> path name
	pathCameras  map[string]string // path name -> device path
	mappingMutex sync.RWMutex

	// Idempotency protection for concurrent path creation
	createGroup singleflight.Group

	// Per-path mutexes for serializing create→ready→patch operations
	pathMutexes map[string]*sync.Mutex
	pathMutexMu sync.RWMutex

	// Metrics tracking
	metrics *PathManagerMetrics
}

// PathManagerMetrics tracks path operation metrics
type PathManagerMetrics struct {
	PathReadyLatencyMs  int64 `json:"path_ready_latency_ms"` // Histogram of path ready latency
	PatchAttemptsTotal  int64 `json:"patch_attempts_total"`  // Counter of patch attempts
	DeviceEventsTotal   int64 `json:"device_events_total"`   // Counter of device events (add/remove/change)
	PathOperationsTotal int64 `json:"path_operations_total"` // Counter of total path operations
}

// NewPathManager creates a new MediaMTX path manager
func NewPathManager(client MediaMTXClient, config *config.MediaMTXConfig, logger *logging.Logger) PathManager {
	return &pathManager{
		client:      client,
		config:      config,
		logger:      logger,
		cameraPaths: make(map[string]string),
		pathCameras: make(map[string]string),
		pathMutexes: make(map[string]*sync.Mutex),
		metrics:     &PathManagerMetrics{},
	}
}

// NewPathManagerWithCamera creates a new MediaMTX path manager with camera monitor
func NewPathManagerWithCamera(client MediaMTXClient, config *config.MediaMTXConfig, cameraMonitor camera.CameraMonitor, logger *logging.Logger) PathManager {
	return &pathManager{
		client:        client,
		config:        config,
		logger:        logger,
		cameraMonitor: cameraMonitor,
		cameraPaths:   make(map[string]string),
		pathCameras:   make(map[string]string),
		pathMutexes:   make(map[string]*sync.Mutex),
		metrics:       &PathManagerMetrics{},
	}
}

// getPathMutex gets or creates a mutex for the given path name
func (pm *pathManager) getPathMutex(pathName string) *sync.Mutex {
	pm.pathMutexMu.RLock()
	mutex, exists := pm.pathMutexes[pathName]
	pm.pathMutexMu.RUnlock()

	if exists {
		return mutex
	}

	pm.pathMutexMu.Lock()
	defer pm.pathMutexMu.Unlock()

	// Double-check after acquiring write lock
	if mutex, exists := pm.pathMutexes[pathName]; exists {
		return mutex
	}

	// Create new mutex for this path
	mutex = &sync.Mutex{}
	pm.pathMutexes[pathName] = mutex
	return mutex
}

// CreatePath creates a new path with idempotency protection
func (pm *pathManager) CreatePath(ctx context.Context, name, source string, options *PathConf) error {
	// Track path operation metrics
	atomic.AddInt64(&pm.metrics.PathOperationsTotal, 1)

	// DEFENSIVE NORMALIZATION: Handle nil options PathConf
	// Contract: *PathConf params are optional; nil means "no options"
	// PathManager never mutates caller structs; always normalize before writes
	var opts *PathConf
	if options == nil {
		opts = &PathConf{}
	} else {
		// Clone to avoid mutating caller's PathConf
		opts = &PathConf{}
		*opts = *options // Shallow copy is sufficient for PathConf
	}

	// IMPORTANT: Avoid "publisher" source which creates runtime-only paths
	// Convert "publisher" to a concrete on-demand source
	if source == "publisher" {
		pm.logger.WithField("path_name", name).
			Warn("Converting 'publisher' source to on-demand configuration")

		// Check if this is for a camera device
		devicePath := GetDevicePathFromCameraIdentifier(name)
		if devicePath != "" && strings.HasPrefix(devicePath, "/dev/video") {
			// Create an on-demand FFmpeg command for the camera
			source = fmt.Sprintf(
				"ffmpeg -f v4l2 -i %s -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/%s",
				devicePath, name,
			)
			opts.RunOnDemand = source
			opts.RunOnDemandRestart = true
			opts.RunOnDemandStartTimeout = pm.config.RunOnDemandStartTimeout
			opts.RunOnDemandCloseAfter = pm.config.RunOnDemandCloseAfter
			// Clear source since we're using runOnDemand
			source = ""
		} else {
			// For non-camera paths, use a redirect or leave empty
			// Empty source with runOnDemand allows dynamic publisher connection
			source = ""
			if opts.RunOnDemand == "" {
				// Generate proper FFmpeg command using centralized configuration
				pathName := GetMediaMTXPathName(devicePath)
				opts.RunOnDemand = BuildFFmpegCommand(devicePath, pathName, pm.config)
				opts.RunOnDemandRestart = true
				opts.RunOnDemandStartTimeout = pm.config.RunOnDemandStartTimeout
				opts.RunOnDemandCloseAfter = pm.config.RunOnDemandCloseAfter
			}
		}
	}

	// Get device path for logging context
	devicePath := GetDevicePathFromCameraIdentifier(name)
	if devicePath == "" {
		devicePath = source // fallback to source if not a camera identifier
	}

	pm.logger.WithFields(logging.Fields{
		"camera_id":   name,
		"device_path": devicePath,
		"path_name":   name,
		"method":      "POST",
		"endpoint":    FormatConfigPathsAdd(name),
		"source":      source,
		"options":     options,
	}).Debug("Creating MediaMTX path")

	// Enhanced validation for better user experience and software resilience
	if err := pm.validatePathName(name); err != nil {
		return fmt.Errorf("invalid path name: %w", err)
	}

	// Validate camera device exists if this is a camera path
	if strings.HasPrefix(name, "camera") {
		exists, err := pm.ValidateCameraDevice(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to validate camera device %s: %w", name, err)
		}
		if !exists {
			return fmt.Errorf("camera %s not found: camera device does not exist", name)
		}
	}

	if err := pm.validateSource(source, opts); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// Use singleflight to prevent concurrent creation attempts for the same path
	result, err, _ := pm.createGroup.Do(name, func() (interface{}, error) {
		return nil, pm.createPathInternal(ctx, name, source, opts, devicePath)
	})

	if err != nil {
		// Check if this is a "path exists" error
		if pm.isAlreadyExistsError(err) {
			pm.logger.WithField("path_name", name).
				Info("Path already exists, checking if we can use it")

			// Try to get the existing path
			if existingPath, getErr := pm.GetPath(ctx, name); getErr == nil {
				pm.logger.WithFields(logging.Fields{
					"path_name": name,
					"source":    existingPath.Source,
				}).Info("Using existing path")

				// If we need to update the configuration, patch it
				if options != nil {
					if patchErr := pm.PatchPath(ctx, name, options); patchErr != nil {
						pm.logger.WithError(patchErr).
							Warn("Could not patch existing path, continuing anyway")
					}
				}

				return nil // Success - path exists and is usable
			}

			// Path exists but we can't get it, might be a runtime path
			// Try to work around it
			pm.logger.WithField("path_name", name).
				Warn("Path exists but not accessible, may be runtime path")

			// Return success anyway for idempotency
			return nil
		}

		return err
	}

	// Result is always nil for this operation
	_ = result
	return nil
}

// createPathInternal performs the actual path creation logic
func (pm *pathManager) createPathInternal(ctx context.Context, name, source string, options *PathConf, devicePath string) error {
	// Create path request using PathConf for configuration
	path := &PathConf{
		Name:   name,
		Source: source,
	}

	// Apply options to path from PathConf
	if options != nil {
		if options.SourceOnDemand {
			path.SourceOnDemand = options.SourceOnDemand
		}
		if options.SourceOnDemandStartTimeout != "" {
			path.SourceOnDemandStartTimeout = options.SourceOnDemandStartTimeout
		}
		if options.SourceOnDemandCloseAfter != "" {
			path.SourceOnDemandCloseAfter = options.SourceOnDemandCloseAfter
		}
		if options.RunOnDemand != "" {
			path.RunOnDemand = options.RunOnDemand
		}
		if options.RunOnDemandRestart {
			path.RunOnDemandRestart = options.RunOnDemandRestart
		}
		if options.RunOnDemandCloseAfter != "" {
			path.RunOnDemandCloseAfter = options.RunOnDemandCloseAfter
		}
		if options.RunOnDemandStartTimeout != "" {
			path.RunOnDemandStartTimeout = options.RunOnDemandStartTimeout
		}
	}

	// Marshal request - choose correct marshaling function based on path type
	var data []byte
	var err error

	if path.RunOnDemand != "" {
		// For on-demand streams (USB devices), use USB-specific marshaling
		data, err = marshalCreateUSBPathRequest(name, path.RunOnDemand)
	} else if strings.HasPrefix(source, "rtsp://") || strings.HasPrefix(source, "rtmp://") {
		// For external stream sources, use stream-specific marshaling
		data, err = marshalCreateStreamRequest(name, source)
	} else {
		// For direct sources, use standard path marshaling
		// Convert PathConf to Path for marshaling
		pathForMarshaling := &Path{
			Name:   path.Name,
			Source: nil, // Will be set by MediaMTX
		}
		data, err = marshalCreatePathRequest(pathForMarshaling)
	}

	if err != nil {
		return NewPathErrorWithErr(name, "create_path", "failed to marshal request", err)
	}

	// Send request - name must be in URL path per Swagger spec
	pm.logger.WithFields(logging.Fields{
		"name": name,
		"data": string(data),
		"url":  FormatConfigPathsAdd(name),
	}).Info("Sending CreatePath request to MediaMTX - FULL REQUEST")

	response, err := pm.client.Post(ctx, FormatConfigPathsAdd(name), data)
	pm.logger.WithFields(logging.Fields{
		"path_name": name,
		"error":     err,
		"error_nil": err == nil,
		"response":  response,
	}).Info("CreatePath API call completed - FULL RESPONSE")
	if err != nil {
		// Log the actual error for debugging with full context
		errorMsg := err.Error()
		pm.logger.WithError(err).WithFields(logging.Fields{
			"camera_id":     name,
			"device_path":   devicePath,
			"path_name":     name,
			"method":        "POST",
			"endpoint":      FormatConfigPathsAdd(name),
			"status":        "failed",
			"body":          truncateString(string(data), 200), // Trimmed body
			"error_type":    fmt.Sprintf("%T", err),
			"error_message": errorMsg,
		}).Error("CreatePath HTTP request failed - investigating idempotency")

		// Check if this is a "path exists" error (idempotent success)
		// Check both the error message and details for the specific error text
		isAlreadyExists := strings.Contains(errorMsg, "path already exists") ||
			strings.Contains(errorMsg, "already exists")

		// Also check the details field for MediaMTXError
		if mediaMTXErr, ok := err.(*MediaMTXError); ok {
			isAlreadyExists = isAlreadyExists ||
				strings.Contains(mediaMTXErr.Details, "path already exists") ||
				strings.Contains(mediaMTXErr.Details, "already exists")
		}

		if isAlreadyExists {
			pm.logger.WithFields(logging.Fields{
				"name":  name,
				"error": errorMsg,
			}).Info("MediaMTX path already exists, treating as success")
			return nil // Idempotent success - path exists, which is what we wanted
		}
		return NewPathErrorWithErr(name, "create_path", "failed to create path", err)
	}

	pm.logger.WithFields(logging.Fields{
		"camera_id":   name,
		"device_path": devicePath,
		"path_name":   name,
		"method":      "POST",
		"endpoint":    FormatConfigPathsAdd(name),
		"status":      "success",
	}).Info("MediaMTX path created successfully")
	return nil
}

// PatchPath patches a path configuration with retry and jitter
func (pm *pathManager) PatchPath(ctx context.Context, name string, config *PathConf) error {
	// Track patch attempt metrics
	atomic.AddInt64(&pm.metrics.PatchAttemptsTotal, 1)

	// DEFENSIVE NORMALIZATION: Handle nil config PathConf
	// Treat nil as empty PathConf {} to prevent "null" PATCH body
	var patchConfig *PathConf
	if config == nil {
		patchConfig = &PathConf{}
	} else {
		patchConfig = config
	}

	// Get device path for logging context
	devicePath := GetDevicePathFromCameraIdentifier(name)
	if devicePath == "" {
		devicePath = name // fallback if not a camera identifier
	}

	pm.logger.WithFields(logging.Fields{
		"camera_id":   name,
		"device_path": devicePath,
		"path_name":   name,
		"method":      "PATCH",
		"endpoint":    FormatConfigPathsPatch(name),
		"config":      patchConfig,
	}).Debug("Patching MediaMTX path configuration")

	if err := pm.validatePathName(name); err != nil {
		return fmt.Errorf("invalid path name: %w", err)
	}

	data, err := json.Marshal(patchConfig)
	if err != nil {
		return NewPathErrorWithErr(name, "patch_path", "failed to marshal config", err)
	}

	// Retry with jitter: 100ms → 200ms → 400ms → 800ms (cap at ~2s total)
	backoffs := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}

	for attempt, backoff := range backoffs {
		pm.logger.WithFields(logging.Fields{
			"path_name":    name,
			"url":          FormatConfigPathsPatch(name),
			"json_payload": string(data),
			"config":       config,
			"attempt":      attempt + 1,
		}).Debug("Sending PATCH request to MediaMTX")

		err = pm.client.Patch(ctx, FormatConfigPathsPatch(name), data)
		if err == nil {
			pm.logger.WithFields(logging.Fields{
				"camera_id":   name,
				"device_path": devicePath,
				"path_name":   name,
				"method":      "PATCH",
				"endpoint":    FormatConfigPathsPatch(name),
				"status":      "success",
			}).Info("MediaMTX path configuration patched successfully")
			return nil
		}

		// Check if this is a retryable error
		errorMsg := err.Error()
		pm.logger.WithFields(logging.Fields{
			"path_name": name,
			"attempt":   attempt + 1,
			"error_msg": errorMsg,
		}).Debug("Checking if error is retryable")

		isRetryable := strings.Contains(errorMsg, "404") ||
			strings.Contains(errorMsg, "409") ||
			strings.Contains(errorMsg, "400") ||
			strings.Contains(errorMsg, "path not found") ||
			strings.Contains(errorMsg, "already exists") ||
			strings.Contains(errorMsg, "busy") ||
			strings.Contains(errorMsg, "bad request") ||
			strings.Contains(errorMsg, "invalid configuration")

		pm.logger.WithFields(logging.Fields{
			"path_name":    name,
			"attempt":      attempt + 1,
			"is_retryable": isRetryable,
			"max_attempts": len(backoffs),
		}).Debug("Retry decision")

		if !isRetryable || attempt == len(backoffs)-1 {
			// Log with comprehensive context including status and response body
			errorMsg := err.Error()
			pm.logger.WithError(err).WithFields(logging.Fields{
				"camera_id":   name,
				"device_path": devicePath,
				"path_name":   name,
				"method":      "PATCH",
				"endpoint":    FormatConfigPathsPatch(name),
				"status":      "failed",
				"body":        truncateString(string(data), 200), // Trimmed body
				"attempt":     attempt + 1,
				"error_msg":   errorMsg,
			}).Error("MediaMTX PATCH request failed")
			return NewPathErrorWithErr(name, "patch_path", "failed to patch path", err)
		}

		// Poll runtime path visibility between retries
		pm.logger.WithFields(logging.Fields{
			"path_name": name,
			"attempt":   attempt + 1,
			"backoff":   backoff,
		}).Debug("PATCH failed, checking runtime path visibility before retry")

		// Check runtime path visibility (not config)
		if _, runtimeErr := pm.GetPath(ctx, name); runtimeErr == nil {
			pm.logger.WithField("path_name", name).Debug("Path is visible in runtime, retrying PATCH")
		}

		pm.logger.WithFields(logging.Fields{
			"path_name": name,
			"attempt":   attempt + 1,
			"backoff":   backoff,
		}).Debug("Retrying PATCH after backoff")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return NewPathErrorWithErr(name, "patch_path", "failed to patch path after all retries", err)
}

// ActivatePathPublisher performs deterministic RTSP activation to trigger MediaMTX publisher
// This is a protocol-based activation, not time-based waiting
func (pm *pathManager) ActivatePathPublisher(ctx context.Context, name string) error {
	pm.logger.WithFields(logging.Fields{
		"path_name": name,
	}).Debug("Activating MediaMTX publisher via RTSP handshake")

	// Generate RTSP URL for the path
	rtspURL := fmt.Sprintf("rtsp://%s:%d/%s", pm.config.Host, pm.config.RTSPPort, name)

	// Create a short-lived context for the RTSP handshake
	handshakeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform one-shot RTSP handshake to activate the publisher
	// This triggers MediaMTX to start the on-demand source

	// Use ffprobe for a quick RTSP connectivity check (one packet)
	// This is deterministic - either the handshake succeeds or fails
	cmd := exec.CommandContext(handshakeCtx, "ffprobe",
		"-v", "quiet",
		"-rtsp_transport", "tcp",
		"-timeout", "2000000", // 2 second timeout in microseconds
		"-show_entries", "format=duration",
		rtspURL)

	if err := cmd.Run(); err != nil {
		pm.logger.WithFields(logging.Fields{
			"path_name": name,
			"rtsp_url":  rtspURL,
			"error":     err.Error(),
		}).Debug("RTSP activation failed - publisher may not be ready yet")
		return NewPathError(name, "rtsp_activation", fmt.Sprintf("failed to activate publisher via RTSP: %v", err))
	}

	pm.logger.WithFields(logging.Fields{
		"path_name": name,
		"rtsp_url":  rtspURL,
	}).Debug("RTSP activation successful - publisher should be active")
	return nil
}

// DeletePath deletes a path with fallback for runtime paths
func (pm *pathManager) DeletePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Deleting MediaMTX path")

	// Try to delete via config API first
	err := pm.client.Delete(ctx, FormatConfigPathsDelete(name))
	if err != nil {
		// Check if this is a "not found" error
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "404") {
			pm.logger.WithField("name", name).
				Debug("Path not found in config, checking runtime")

			// Check if it exists as a runtime path
			if _, getErr := pm.GetPath(ctx, name); getErr == nil {
				pm.logger.WithField("name", name).
					Warn("Path exists in runtime but not in config - cannot delete runtime paths")

				// Runtime paths can't be deleted via API
				// Best we can do is disconnect any active connections
				// This is a MediaMTX limitation
				return nil // Return success for idempotency
			}

			// Path doesn't exist at all - return error for proper error handling
			pm.logger.WithField("name", name).Debug("Path does not exist")
			return NewPathErrorWithErr(name, "delete_path", "path not found", err)
		}

		return NewPathErrorWithErr(name, "delete_path", "failed to delete path", err)
	}

	pm.logger.WithField("name", name).Info("MediaMTX path deleted successfully")
	return nil
}

// GetPath gets a specific path (runtime status)
func (pm *pathManager) GetPath(ctx context.Context, name string) (*Path, error) {
	pm.logger.WithField("name", name).Debug("Getting MediaMTX path")

	data, err := pm.client.Get(ctx, FormatPathsGet(name))
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to get path", err)
	}

	path, err := parseStreamResponse(data)
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to parse stream response", err)
	}

	return path, nil
}

// ListPaths lists all path configurations
func (pm *pathManager) ListPaths(ctx context.Context) ([]*PathConf, error) {
	pm.logger.Debug("Listing MediaMTX path configurations")

	data, err := pm.client.Get(ctx, MediaMTXConfigPathsList)
	if err != nil {
		return nil, fmt.Errorf("failed to list path configurations: %w", err)
	}

	paths, err := parsePathConfListResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path configurations response: %w", err)
	}

	pm.logger.WithField("count", strconv.Itoa(len(paths))).Debug("MediaMTX path configurations listed successfully")
	return paths, nil
}

// GetRuntimePaths gets all runtime paths (active paths with status)
func (pm *pathManager) GetRuntimePaths(ctx context.Context) ([]*Path, error) {
	pm.logger.Debug("Getting MediaMTX runtime paths")

	data, err := pm.client.Get(ctx, MediaMTXPathsList)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime paths: %w", err)
	}

	paths, err := parsePathListResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime paths response: %w", err)
	}

	pm.logger.WithField("count", strconv.Itoa(len(paths))).Debug("MediaMTX runtime paths retrieved successfully")
	return paths, nil
}

// ValidatePath validates a path
func (pm *pathManager) ValidatePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Validating MediaMTX path")

	// Get path details to validate configuration (this also checks existence)
	_, err := pm.GetPath(ctx, name)
	if err != nil {
		return NewPathErrorWithErr(name, "validate_path", "path does not exist or failed to get path details", err)
	}

	pm.logger.WithField("name", name).Debug("MediaMTX path validated successfully")
	return nil
}

// PathExists checks if a path exists in runtime (not config)
func (pm *pathManager) PathExists(ctx context.Context, name string) bool {
	pm.logger.WithField("name", name).Debug("Checking if MediaMTX path exists in runtime")

	// Use runtime endpoint to check path existence (not config)
	_, err := pm.GetPath(ctx, name)
	return err == nil
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
func (pm *pathManager) validateSource(source string, options *PathConf) error {
	// Allow empty source if runOnDemand is specified
	if source == "" {
		if options != nil && options.RunOnDemand != "" {
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
	apiCameras := make([]CameraInfo, 0, len(cameras))
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

		// Create API-ready camera info using CameraInfo from rpc_types.go
		apiCamera := CameraInfo{
			Device:     cameraID,                    // Abstracted camera ID
			Status:     string(cameraDevice.Status), // Camera status
			Name:       cameraDevice.Name,           // Camera name
			Resolution: resolution,                  // Extracted resolution
			FPS:        30,                          // Default FPS (could be extracted from capabilities)
			Streams:    streams,                     // Generated stream URLs
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
func (pm *pathManager) GetCameraStatus(ctx context.Context, device string) (*GetCameraStatusResponse, error) {
	if pm.cameraMonitor == nil {
		return nil, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get actual device path using abstraction layer
	devicePath := GetDevicePathFromCameraIdentifier(device)

	// Get camera from camera monitor
	cameraDevice, exists := pm.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return nil, fmt.Errorf("camera not found: %s", device)
	}

	// Build response with abstraction layer
	response := &GetCameraStatusResponse{
		Device:       device, // Return camera identifier (camera0)
		Status:       string(cameraDevice.Status),
		Name:         cameraDevice.Name,
		Resolution:   "",
		FPS:          0,
		Streams:      make(map[string]string),
		Capabilities: pm.convertV4L2CapabilitiesToMap(&cameraDevice.Capabilities),
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

// GetCameraCapabilities returns detailed capabilities for a specific camera
func (pm *pathManager) GetCameraCapabilities(ctx context.Context, device string) (*GetCameraCapabilitiesResponse, error) {
	if pm.cameraMonitor == nil {
		return nil, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get actual device path using abstraction layer
	devicePath := GetDevicePathFromCameraIdentifier(device)

	// Get camera from camera monitor
	cameraDevice, exists := pm.cameraMonitor.GetDevice(devicePath)
	if !exists {
		return nil, fmt.Errorf("camera not found: %s", device)
	}

	// Extract ALL formats from camera module (no more single format extraction)
	supportedFormats := make([]string, 0, len(cameraDevice.Formats))
	supportedResolutions := make([]string, 0, len(cameraDevice.Formats))
	fpsOptionsMap := make(map[int]bool)

	for _, format := range cameraDevice.Formats {
		// Collect pixel formats
		if format.PixelFormat != "" {
			supportedFormats = append(supportedFormats, format.PixelFormat)
		}

		// Collect resolutions
		if format.Width > 0 && format.Height > 0 {
			resolution := fmt.Sprintf("%dx%d", format.Width, format.Height)
			supportedResolutions = append(supportedResolutions, resolution)
		}

		// Collect all frame rates
		for _, fpsStr := range format.FrameRates {
			if fps, err := strconv.Atoi(strings.TrimSuffix(fpsStr, " fps")); err == nil {
				fpsOptionsMap[fps] = true
			}
		}
	}

	// Convert FPS map to sorted slice
	fpsOptions := make([]int, 0, len(fpsOptionsMap))
	for fps := range fpsOptionsMap {
		fpsOptions = append(fpsOptions, fps)
	}

	// Determine validation status based on camera state and capability detection
	validationStatus := "none"
	if cameraDevice.Status == "CONNECTED" {
		validationStatus = "confirmed"
	} else if cameraDevice.Status == "DISCONNECTED" {
		validationStatus = "disconnected"
	}

	// Build API-ready response using correct GetCameraCapabilitiesResponse fields
	response := &GetCameraCapabilitiesResponse{
		Device:           device,
		Formats:          supportedFormats,
		Resolutions:      supportedResolutions,
		FpsOptions:       fpsOptions,
		FrameRates:       fpsOptions, // Keep both for backward compatibility
		Capabilities:     cameraDevice.Capabilities.Capabilities,
		ValidationStatus: validationStatus, // Set validation status per API documentation
	}

	pm.logger.WithFields(logging.Fields{
		"device":            device,
		"formats_count":     len(supportedFormats),
		"resolutions_count": len(supportedResolutions),
		"fps_options_count": len(fpsOptions),
	}).Debug("PathManager retrieved camera capabilities")

	return response, nil
}

// ValidateCameraDevice validates that a camera device exists and is accessible
func (pm *pathManager) ValidateCameraDevice(ctx context.Context, device string) (bool, error) {
	if pm.cameraMonitor == nil {
		return false, fmt.Errorf("camera monitor not available in path manager")
	}

	// Get actual device path using abstraction layer
	devicePath := GetDevicePathFromCameraIdentifier(device)

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
	devicePath := GetDevicePathFromCameraIdentifier(cameraID)

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
	cameraID := GetMediaMTXPathName(devicePath)
	return cameraID, true
}

// GetDevicePathForCamera gets the actual USB device path for a camera identifier
// This is the main abstraction: camera0 -> /dev/video0
func (pm *pathManager) GetDevicePathForCamera(cameraID string) (string, bool) {
	devicePath := GetDevicePathFromCameraIdentifier(cameraID)

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
	cameraID := GetMediaMTXPathName(devicePath)

	// Check if device actually exists via camera monitor
	if pm.cameraMonitor != nil {
		_, exists := pm.cameraMonitor.GetDevice(devicePath)
		return cameraID, exists
	}

	// Fallback: return converted identifier without validation
	return cameraID, true
}

// GetMetrics returns the current path manager metrics
func (pm *pathManager) GetMetrics() *PathManagerMetrics {
	return &PathManagerMetrics{
		PathReadyLatencyMs:  atomic.LoadInt64(&pm.metrics.PathReadyLatencyMs),
		PatchAttemptsTotal:  atomic.LoadInt64(&pm.metrics.PatchAttemptsTotal),
		DeviceEventsTotal:   atomic.LoadInt64(&pm.metrics.DeviceEventsTotal),
		PathOperationsTotal: atomic.LoadInt64(&pm.metrics.PathOperationsTotal),
	}
}

// TrackDeviceEvent tracks device events for metrics
func (pm *pathManager) TrackDeviceEvent(eventType string) {
	atomic.AddInt64(&pm.metrics.DeviceEventsTotal, 1)
	pm.logger.WithFields(logging.Fields{
		"event_type":   eventType,
		"total_events": atomic.LoadInt64(&pm.metrics.DeviceEventsTotal),
	}).Debug("Device event tracked")
}

// isAlreadyExistsError checks if error indicates path exists
func (pm *pathManager) isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	return strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "path already exists") ||
		strings.Contains(errorMsg, "409") // Conflict status code
}

// truncateString truncates a string to the specified length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// convertV4L2CapabilitiesToMap converts V4L2Capabilities to map for API response
func (pm *pathManager) convertV4L2CapabilitiesToMap(caps *camera.V4L2Capabilities) map[string]interface{} {
	if caps == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	if len(caps.Capabilities) > 0 {
		result["capabilities"] = caps.Capabilities
	}
	if caps.DriverName != "" {
		result["driver_name"] = caps.DriverName
	}
	if caps.CardName != "" {
		result["card_name"] = caps.CardName
	}
	if caps.BusInfo != "" {
		result["bus_info"] = caps.BusInfo
	}
	if caps.Version != "" {
		result["version"] = caps.Version
	}
	if len(caps.DeviceCaps) > 0 {
		result["device_caps"] = caps.DeviceCaps
	}
	return result
}
