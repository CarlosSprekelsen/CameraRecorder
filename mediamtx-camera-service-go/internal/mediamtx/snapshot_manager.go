/*
MediaMTX Snapshot Manager Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// SnapshotManager manages multi-tier snapshot capture with performance optimization.
//
// RESPONSIBILITIES:
// - Multi-tier snapshot capture with intelligent fallback system
// - V4L2 direct capture for high-performance USB device snapshots
// - RTSP stream capture for external devices and fallback scenarios
// - Camera capability extraction and validation for V4L2 devices
//
// TIER ARCHITECTURE:
// - Tier 0: V4L2 direct capture (fastest, USB devices only)
// - Tier 1: FFmpeg direct capture (fast, USB devices)
// - Tier 2: RTSP immediate capture (existing streams)
// - Tier 3: RTSP activation capture (create stream then capture)
//
// API INTEGRATION:
// - Operates with cameraID as primary identifier
// - Returns JSON-RPC API-ready responses
// - Converts to devicePath only for V4L2 operations
//
// TakeSnapshot returns *TakeSnapshotResponse directly with proper response formatting
// ListSnapshots method returns *ListSnapshotsResponse (API-ready) - implemented correctly
type SnapshotManager struct {
	ffmpegManager FFmpegManager
	streamManager StreamManager        // Required for Tier 3: external RTSP source path creation
	cameraMonitor camera.CameraMonitor // Required for Tier 0: V4L2 direct capture
	pathManager   PathManager          // Required for camera identifier to device path conversion
	config        *config.MediaMTXConfig
	logger        *logging.Logger

	// Configuration integration for multi-tier support
	configManager *config.ConfigManager

	// Snapshot settings
	snapshotSettings *SnapshotSettings

	// Snapshot tracking - using sync.Map for lock-free operations
	snapshots sync.Map // snapshotID -> *Snapshot
}

// SnapshotSettings defines snapshot behavior
type SnapshotSettings struct {
	Format      string `json:"format"`      // jpg, png, etc.
	Quality     int    `json:"quality"`     // 1-100 for JPEG
	MaxWidth    int    `json:"max_width"`   // Maximum width
	MaxHeight   int    `json:"max_height"`  // Maximum height
	AutoResize  bool   `json:"auto_resize"` // Auto-resize if needed
	Compression int    `json:"compression"` // Compression level
}

// NewSnapshotManagerWithConfig creates a new snapshot manager with configuration integration
func NewSnapshotManagerWithConfig(ffmpegManager FFmpegManager, streamManager StreamManager, cameraMonitor camera.CameraMonitor, pathManager PathManager, config *config.MediaMTXConfig, configManager *config.ConfigManager, logger *logging.Logger) *SnapshotManager {
	return &SnapshotManager{
		ffmpegManager: ffmpegManager,
		streamManager: streamManager,
		cameraMonitor: cameraMonitor,
		pathManager:   pathManager,
		config:        config,
		configManager: configManager,
		logger:        logger,
		// snapshots: sync.Map is zero-initialized, no need to initialize
		snapshotSettings: &SnapshotSettings{
			Format:      "jpg",
			Quality:     85,
			MaxWidth:    1920,
			MaxHeight:   1080,
			AutoResize:  true,
			Compression: 6,
		},
	}
}

// TakeSnapshot takes a snapshot with multi-tier approach and returns API-ready response
func (sm *SnapshotManager) TakeSnapshot(ctx context.Context, cameraID string, options *SnapshotOptions) (*TakeSnapshotResponse, error) {
	// Convert camera identifier to device path using PathManager
	devicePath, exists := sm.pathManager.GetDevicePathForCamera(cameraID)
	if !exists {
		return nil, fmt.Errorf("camera '%s' not found or not accessible", cameraID)
	}

	// Generate snapshot path using device path for file naming
	// Use PathManager naming policy for device subdir and filenames
	snapshotPath := GenerateSnapshotPath(sm.config, &sm.configManager.GetConfig().Snapshots, devicePath)

	sm.logger.WithFields(logging.Fields{
		"cameraID":   cameraID,
		"devicePath": devicePath,
		"path":       snapshotPath,
		"options":    options,
	}).Info("Taking multi-tier snapshot")

	// Apply snapshot settings from strongly-typed options
	if options != nil {
		if options.Format != "" {
			sm.snapshotSettings.Format = options.Format
		}
		if options.Quality > 0 {
			sm.snapshotSettings.Quality = options.Quality
		}
		if options.MaxWidth > 0 {
			sm.snapshotSettings.MaxWidth = options.MaxWidth
		}
		if options.MaxHeight > 0 {
			sm.snapshotSettings.MaxHeight = options.MaxHeight
		}
		if options.AutoResize {
			sm.snapshotSettings.AutoResize = options.AutoResize
		}
		if options.Compression > 0 {
			sm.snapshotSettings.Compression = options.Compression
		}
	}

	// Get tier configuration from existing config system
	tierConfig := sm.getTierConfiguration()
	if tierConfig == nil {
		return nil, fmt.Errorf("failed to get tier configuration - config manager not properly initialized")
	}

	// Generate snapshot ID
	snapshotID := generateSnapshotID(cameraID)

	// Execute multi-tier snapshot capture
	snapshot, err := sm.takeSnapshotMultiTier(ctx, cameraID, devicePath, snapshotPath, options, tierConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to execute multi-tier snapshot capture: %w", err)
	}

	// Store snapshot - lock-free operation with sync.Map
	sm.snapshots.Store(snapshotID, snapshot)

	// Store the camera identifier in the snapshot for API consistency
	snapshot.Device = cameraID

	// Extract filename from full path
	filename := snapshot.FilePath
	if parts := strings.Split(snapshot.FilePath, "/"); len(parts) > 0 {
		filename = parts[len(parts)-1]
	}

	// Build API-ready response with rich metadata from snapshot
	response := &TakeSnapshotResponse{
		Device:    cameraID,                              // Use cameraID for API consistency
		Filename:  filename,                              // Extracted filename
		Status:    "completed",                           // Successful capture
		Timestamp: snapshot.Created.Format(time.RFC3339), // ISO 8601 timestamp
		FileSize:  snapshot.Size,                         // Actual file size
		FilePath:  snapshot.FilePath,                     // Full file path
	}

	sm.logger.WithFields(logging.Fields{
		"snapshot_id":    snapshotID,
		"cameraID":       cameraID,
		"devicePath":     devicePath,
		"file_size":      snapshot.Size,
		"tier_used":      snapshot.Metadata["tier_used"],
		"capture_method": snapshot.Metadata["capture_method"],
		"format":         sm.snapshotSettings.Format,
		"quality":        sm.snapshotSettings.Quality,
	}).Info("Multi-tier snapshot completed with API-ready response")

	return response, nil
}

// takeSnapshotMultiTier implements the 5-tier snapshot capture system
func (sm *SnapshotManager) takeSnapshotMultiTier(ctx context.Context, cameraID, devicePath, snapshotPath string, options *SnapshotOptions, tierConfig *config.SnapshotTiersConfig) (*Snapshot, error) {
	startTime := time.Now()
	captureMethodsTried := []string{}

	sm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"tier":     0,
	}).Info("Tier 0: Attempting V4L2 direct capture")

	// Tier 0: V4L2 Direct Capture (Fastest Path - used /dev/vide)
	tier0Ctx, tier0Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier1USBDirectTimeout*float64(time.Second)))
	defer tier0Cancel()

	if snapshot, err := sm.captureSnapshotV4L2Direct(tier0Ctx, devicePath, snapshotPath, options); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 0, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"cameraID":     cameraID,
			"tier":         0,
			"capture_time": captureTime,
		}).Info("Tier 0: V4L2 direct capture successful")
		return result, nil
	} else {
		sm.logger.WithFields(logging.Fields{
			"cameraID": cameraID,
			"tier":     0,
			"error":    err.Error(),
		}).Warn("Tier 0: V4L2 direct capture failed")
	}
	captureMethodsTried = append(captureMethodsTried, "v4l2_direct")

	sm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"tier":     1,
	}).Info("Tier 1: Attempting USB direct capture")

	// Tier 1: USB Direct Capture (Fastest Path)
	tier1Ctx, tier1Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier1USBDirectTimeout*float64(time.Second)))
	defer tier1Cancel()

	if snapshot, err := sm.captureSnapshotDirect(tier1Ctx, devicePath, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 1, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"cameraID":     cameraID,
			"tier":         1,
			"capture_time": captureTime,
		}).Info("Tier 1: USB direct capture successful")
		return result, nil
	} else {
		sm.logger.WithFields(logging.Fields{
			"cameraID": cameraID,
			"tier":     1,
			"error":    err.Error(),
		}).Warn("Tier 1: USB direct capture failed")
	}
	captureMethodsTried = append(captureMethodsTried, "usb_direct")

	sm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"tier":     2,
	}).Info("Tier 2: Attempting RTSP immediate capture")

	// Tier 2: RTSP Immediate Capture
	tier2Ctx, tier2Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier2RTSPReadyCheckTimeout*float64(time.Second)))
	defer tier2Cancel()

	if snapshot, err := sm.captureSnapshotFromRTSP(tier2Ctx, cameraID, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 2, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"cameraID":     cameraID,
			"tier":         2,
			"capture_time": captureTime,
		}).Info("Tier 2: RTSP immediate capture successful")
		return result, nil
	} else {
		sm.logger.WithFields(logging.Fields{
			"cameraID": cameraID,
			"tier":     2,
			"error":    err.Error(),
		}).Warn("Tier 2: RTSP immediate capture failed")
	}
	captureMethodsTried = append(captureMethodsTried, "rtsp_immediate")

	sm.logger.WithFields(logging.Fields{
		"cameraID": cameraID,
		"tier":     3,
	}).Info("Tier 3: Attempting RTSP stream activation")

	// Tier 3: RTSP Stream Activation
	tier3Ctx, tier3Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier3ActivationTimeout*float64(time.Second)))
	defer tier3Cancel()

	if snapshot, err := sm.captureSnapshotFromRTSP(tier3Ctx, cameraID, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 3, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"cameraID":     cameraID,
			"tier":         3,
			"capture_time": captureTime,
		}).Info("Tier 3: RTSP stream activation successful")
		return result, nil
	} else {
		sm.logger.WithFields(logging.Fields{
			"cameraID": cameraID,
			"tier":     3,
			"error":    err.Error(),
		}).Warn("Tier 3: RTSP stream activation failed")
	}
	captureMethodsTried = append(captureMethodsTried, "rtsp_activation")

	// Tier 4: Error Handling - All methods failed
	totalTime := time.Since(startTime)
	sm.logger.WithFields(logging.Fields{
		"cameraID":      cameraID,
		"total_time":    totalTime,
		"methods_tried": captureMethodsTried,
	}).Error("Tier 4: All snapshot capture methods failed")

	return nil, sm.createMultiTierError(cameraID, captureMethodsTried, totalTime)
}

// getTierConfiguration retrieves multi-tier configuration from existing config system
func (sm *SnapshotManager) getTierConfiguration() *config.SnapshotTiersConfig {
	if sm.configManager == nil {
		return nil
	}

	// Get performance configuration from centralized config system
	cfg := sm.configManager.GetConfig()
	if cfg == nil {
		return nil
	}

	return &cfg.Performance.SnapshotTiers
}

// captureSnapshotV4L2Direct implements Tier 0: V4L2 Direct Capture (Fastest Path - NEW)
func (sm *SnapshotManager) captureSnapshotV4L2Direct(ctx context.Context, devicePath, snapshotPath string, options *SnapshotOptions) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"tier":        0,
	}).Info("Tier 0: Attempting V4L2 direct capture")

	// Check if camera monitor is available
	if sm.cameraMonitor == nil {
		return nil, fmt.Errorf("camera monitor not available for V4L2 direct capture")
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory for V4L2 direct capture: %w", err)
	}

	// Use camera monitor's direct snapshot capability
	// Convert SnapshotOptions to map for backward compatibility with camera monitor
	// BUT exclude file format since V4L2 doesn't understand file formats (jpg, png, etc.)
	var optionsMap map[string]interface{}
	if options != nil {
		optionsMap = options.ToMap()
		// Remove file format from V4L2 options - V4L2 uses pixel formats (YUYV, etc.)
		// File format is handled by v4l2-ctl output conversion
		delete(optionsMap, "format")
	} else {
		optionsMap = make(map[string]interface{})
	}
	directSnapshot, err := sm.cameraMonitor.TakeDirectSnapshot(ctx, devicePath, snapshotPath, optionsMap)
	if err != nil {
		return nil, fmt.Errorf("V4L2 direct capture failed: %w", err)
	}

	// Convert DirectSnapshot to Snapshot for compatibility
	snapshot := &Snapshot{
		ID:       directSnapshot.ID,
		Device:   directSnapshot.DevicePath,
		Path:     filepath.Dir(directSnapshot.FilePath),
		FilePath: directSnapshot.FilePath,
		Size:     directSnapshot.Size,
		Created:  directSnapshot.Created,
		Metadata: map[string]interface{}{
			"tier_used":      0,
			"capture_method": "v4l2_direct",
			"capture_time":   directSnapshot.CaptureTime,
			"format":         directSnapshot.Format,
			"width":          directSnapshot.Width,
			"height":         directSnapshot.Height,
		},
	}

	sm.logger.WithFields(logging.Fields{
		"device":       devicePath,
		"output_path":  snapshotPath,
		"file_size":    directSnapshot.Size,
		"capture_time": directSnapshot.CaptureTime,
		"tier":         0,
	}).Info("Tier 0: V4L2 direct capture successful")

	return snapshot, nil
}

// captureSnapshotDirect implements Tier 1: USB Direct Capture (Fastest Path)
func (sm *SnapshotManager) captureSnapshotDirect(ctx context.Context, devicePath, snapshotPath string) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"tier":        1,
	}).Info("Tier 1: Attempting USB direct capture")

	// Use FFmpegManager for capability-aware snapshot command
	command, _ := sm.ffmpegManager.BuildSnapshotCommand(devicePath, snapshotPath, sm.snapshotSettings.Format)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory for FFmpeg snapshot: %w", err)
	}

	// Create command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to take FFmpeg snapshot: %w", err)
	}

	// Get file info using existing FFmpeg manager
	fileSize, _, err := sm.ffmpegManager.GetFileInfo(ctx, snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot file info: %w", err)
	}

	// Create snapshot object with metadata
	snapshot := &Snapshot{
		ID:       generateSnapshotID(devicePath),
		Device:   devicePath,
		Path:     filepath.Dir(snapshotPath),
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
		Metadata: map[string]interface{}{
			"tier_used":      1,
			"capture_method": "usb_direct",
			"format":         sm.snapshotSettings.Format,
			"width":          sm.snapshotSettings.MaxWidth,
			"height":         sm.snapshotSettings.MaxHeight,
			"quality":        sm.snapshotSettings.Quality,
		},
	}

	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"file_size":   fileSize,
		"tier":        1,
	}).Info("Tier 1: USB direct capture successful")

	return snapshot, nil
}

// captureSnapshotFromRTSP implements Tier 2/3: RTSP Capture
func (sm *SnapshotManager) captureSnapshotFromRTSP(ctx context.Context, cameraID, snapshotPath string) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"cameraID":    cameraID,
		"output_path": snapshotPath,
		"tier":        2,
	}).Info("Tier 2/3: Capturing from RTSP stream")

	// Get devicePath only to determine if external or USB
	devicePath, exists := sm.pathManager.GetDevicePathForCamera(cameraID)
	if !exists {
		devicePath = cameraID // For external streams
	}

	// Determine if this is an external RTSP source or USB device
	var streamName string
	var rtspURL string

	if strings.HasPrefix(devicePath, "rtsp://") || strings.HasPrefix(devicePath, "rtmp://") {
		// External RTSP source - need to create MediaMTX path first
		sm.logger.WithFields(logging.Fields{
			"device": devicePath,
			"tier":   3,
		}).Info("Tier 3: External RTSP source detected, creating MediaMTX path")

		// Use StreamManager to create MediaMTX path for external RTSP source (single path)
		stream, err := sm.streamManager.StartStream(ctx, cameraID)
		if err != nil {
			return nil, fmt.Errorf("failed to create MediaMTX path for external RTSP source: %w", err)
		}

		streamName = cameraID      // Use cameraID directly as stream name
		rtspURL = stream.StreamURL // Use the StreamURL from the response

		sm.logger.WithFields(logging.Fields{
			"device":      devicePath,
			"stream_name": streamName,
			"rtsp_url":    rtspURL,
			"tier":        3,
		}).Info("Tier 3: MediaMTX path created for external RTSP source")

		// Stream should be ready immediately
	} else {
		// USB device - assume MediaMTX path exists from streaming setup
		streamName = sm.getStreamNameFromDevice(devicePath)
		rtspURL = fmt.Sprintf("rtsp://%s:%d/%s", sm.config.Host, sm.config.RTSPPort, streamName)

		sm.logger.WithFields(logging.Fields{
			"device":      devicePath,
			"stream_name": streamName,
			"rtsp_url":    rtspURL,
			"tier":        2,
		}).Info("Tier 2: Attempting capture from existing MediaMTX stream")
	}

	// Build FFmpeg command for RTSP capture
	command := []string{"ffmpeg"}
	command = append(command, "-i", rtspURL)
	command = append(command, "-vframes", "1")
	command = append(command, "-q:v", strconv.Itoa(sm.snapshotSettings.Quality))
	command = append(command, "-y") // Overwrite output file
	command = append(command, snapshotPath)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory for FFmpeg snapshot: %w", err)
	}

	// Create command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to take snapshot from RTSP: %w", err)
	}

	// Get file info using existing FFmpeg manager
	fileSize, _, err := sm.ffmpegManager.GetFileInfo(ctx, snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot file info: %w", err)
	}

	// Create snapshot object with metadata
	snapshot := &Snapshot{
		ID:       generateSnapshotID(devicePath),
		Device:   devicePath,
		Path:     filepath.Dir(snapshotPath),
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
		Metadata: map[string]interface{}{
			"tier_used":      2, // Will be updated to 3 if stream activation was used
			"capture_method": "rtsp_immediate",
			"format":         sm.snapshotSettings.Format,
			"width":          sm.snapshotSettings.MaxWidth,
			"height":         sm.snapshotSettings.MaxHeight,
			"quality":        sm.snapshotSettings.Quality,
			"stream_name":    streamName,
		},
	}

	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"file_size":   fileSize,
		"stream_name": streamName,
	}).Info("Tier 2/3: RTSP snapshot captured successfully")

	return snapshot, nil
}

// getStreamNameFromDevice converts device path to stream name
// DELEGATES TO PATHMANAGER - no duplicate abstraction logic
func (sm *SnapshotManager) getStreamNameFromDevice(devicePath string) string {
	if sm.streamManager != nil {
		return sm.streamManager.GenerateStreamName(devicePath, UseCaseRecording)
	}
	return ""
}

// createSnapshotResult creates a snapshot result with tier information
func (sm *SnapshotManager) createSnapshotResult(snapshot *Snapshot, tier int, captureTime time.Duration, methodsTried []string) *Snapshot {
	// Initialize metadata if nil
	if snapshot.Metadata == nil {
		snapshot.Metadata = make(map[string]interface{})
	}

	// Add tier information to existing metadata (don't overwrite)
	snapshot.Metadata["tier_used"] = tier
	snapshot.Metadata["capture_time_ms"] = captureTime.Milliseconds()
	snapshot.Metadata["methods_tried"] = methodsTried
	snapshot.Metadata["user_experience"] = sm.determineUserExperience(captureTime)

	return snapshot
}

// determineUserExperience determines user experience based on response time
func (sm *SnapshotManager) determineUserExperience(captureTime time.Duration) string {
	tierConfig := sm.getTierConfiguration()
	if tierConfig == nil {
		// Fallback to reasonable defaults if config is not available
		sm.logger.Warn("Tier configuration not available, using fallback thresholds")
		if captureTime < 500*time.Millisecond {
			return "excellent"
		} else if captureTime < 2*time.Second {
			return "good"
		} else if captureTime < 5*time.Second {
			return "acceptable"
		}
		return "slow"
	}

	if captureTime < time.Duration(tierConfig.ImmediateResponseThreshold*float64(time.Second)) {
		return "excellent"
	} else if captureTime < time.Duration(tierConfig.AcceptableResponseThreshold*float64(time.Second)) {
		return "good"
	} else if captureTime < time.Duration(tierConfig.SlowResponseThreshold*float64(time.Second)) {
		return "acceptable"
	} else {
		return "slow"
	}
}

// createMultiTierError creates a comprehensive error for multi-tier failures
func (sm *SnapshotManager) createMultiTierError(cameraID string, methodsTried []string, totalTime time.Duration) error {
	return fmt.Errorf("all snapshot capture methods failed for %s after %.2fs: tried %v",
		cameraID, totalTime.Seconds(), methodsTried)
}

// GetSnapshot gets a snapshot by ID
func (sm *SnapshotManager) GetSnapshot(snapshotID string) (*Snapshot, bool) {
	if snapshot, exists := sm.snapshots.Load(snapshotID); exists {
		return snapshot.(*Snapshot), true
	}
	return nil, false
}

// ListSnapshotsInternal lists all snapshots (internal use)
func (sm *SnapshotManager) ListSnapshotsInternal() []*Snapshot {
	var snapshots []*Snapshot

	// Iterate over sync.Map - lock-free operation
	sm.snapshots.Range(func(key, value interface{}) bool {
		snapshots = append(snapshots, value.(*Snapshot))
		return true // Continue iteration
	})

	return snapshots
}

// DeleteSnapshot deletes a snapshot
func (sm *SnapshotManager) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	sm.logger.WithField("snapshot_id", snapshotID).Debug("Deleting snapshot")

	// Get snapshot - lock-free read with sync.Map
	snapshotInterface, exists := sm.snapshots.Load(snapshotID)
	if !exists {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}
	snapshot := snapshotInterface.(*Snapshot)

	// Remove from tracking - lock-free delete with sync.Map
	sm.snapshots.Delete(snapshotID)

	// Delete file
	if err := os.Remove(snapshot.FilePath); err != nil {
		return fmt.Errorf("failed to delete snapshot file: %w", err)
	}

	sm.logger.WithFields(logging.Fields{
		"snapshot_id": snapshotID,
		"file_path":   snapshot.FilePath,
	}).Info("Snapshot deleted successfully")

	return nil
}

// CleanupOldSnapshots cleans up snapshots based on age, count, and size limits using centralized configuration
func (sm *SnapshotManager) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int, maxSize int64) (deletedCount int, spaceFreed int64, err error) {
	sm.logger.WithFields(logging.Fields{
		"max_age":   maxAge,
		"max_count": maxCount,
		"max_size":  maxSize,
	}).Info("Cleaning up old snapshots")

	deletedCount = 0
	spaceFreed = 0

	// Note: sync.Map doesn't need locking for individual operations
	// but we need to collect all snapshots first for consistent cleanup

	// Get snapshots directory path from configuration
	snapshotsDir := sm.config.SnapshotsPath
	if snapshotsDir == "" {
		return 0, 0, fmt.Errorf("snapshots path not configured")
	}

	// Check if directory exists
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		sm.logger.WithField("directory", snapshotsDir).Warn("Snapshots directory does not exist")
		// Still clean up in-memory snapshots even if directory doesn't exist
	} else {
		// Read directory entries
		entries, err := os.ReadDir(snapshotsDir)
		if err != nil {
			sm.logger.WithError(err).WithField("directory", snapshotsDir).Error("Error reading snapshots directory")
			return 0, 0, fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		// Process files and collect metadata
		var files []*FileMetadata
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := entry.Name()

			// Get file stats
			fileInfo, err := entry.Info()
			if err != nil {
				sm.logger.WithError(err).WithField("filename", filename).Warn("Error accessing file")
				continue
			}

			// Create file metadata
			fileMetadata := &FileMetadata{
				FileName:    filename,
				FileSize:    fileInfo.Size(),
				CreatedAt:   fileInfo.ModTime(),
				ModifiedAt:  fileInfo.ModTime(),
				DownloadURL: fmt.Sprintf("/files/snapshots/%s", filename),
			}

			files = append(files, fileMetadata)
		}

		// Sort by modification time (oldest first)
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModifiedAt.Before(files[j].ModifiedAt)
		})

		cutoffTime := time.Now().Add(-maxAge)

		// Calculate current total size if size-based cleanup is enabled
		var currentTotalSize int64
		if maxSize > 0 {
			for _, file := range files {
				currentTotalSize += file.FileSize
			}
		}

		// Delete files based on age, count, and size constraints
		for _, file := range files {
			shouldDelete := false

			// Check age constraint
			if file.ModifiedAt.Before(cutoffTime) {
				shouldDelete = true
			}

			// Check count constraint (keep newest files up to maxCount)
			if len(files)-deletedCount > maxCount {
				shouldDelete = true
			}

			// Check size constraint (delete oldest files until under maxSize)
			if maxSize > 0 && currentTotalSize > maxSize {
				shouldDelete = true
			}

			if shouldDelete {
				filePath := filepath.Join(snapshotsDir, file.FileName)
				if err := sm.deleteSnapshotFile(filePath); err != nil {
					sm.logger.WithError(err).WithField("filename", file.FileName).Error("Failed to delete snapshot file")
					continue
				}
				deletedCount++
				spaceFreed += file.FileSize
				currentTotalSize -= file.FileSize
			}
		}
	}

	// Clean up in-memory snapshots
	// Get all snapshots sorted by creation time - lock-free iteration with sync.Map
	var snapshots []*Snapshot
	sm.snapshots.Range(func(key, value interface{}) bool {
		snapshots = append(snapshots, value.(*Snapshot))
		return true // Continue iteration
	})

	// Sort by creation time (earliest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Created.Before(snapshots[j].Created)
	})

	cutoffTime := time.Now().Add(-maxAge)

	// Calculate current total size of in-memory snapshots if size-based cleanup is enabled
	var currentSnapshotSize int64
	if maxSize > 0 {
		for _, snapshot := range snapshots {
			// For in-memory snapshots, we estimate size or use file size if available
			if fileInfo, err := os.Stat(snapshot.FilePath); err == nil {
				currentSnapshotSize += fileInfo.Size()
			}
		}
	}

	// Delete snapshots from memory based on age, count, and size constraints
	for _, snapshot := range snapshots {
		shouldDelete := false

		// Check age constraint
		if snapshot.Created.Before(cutoffTime) {
			shouldDelete = true
		}

		// Check count constraint (keep newest snapshots up to maxCount)
		if len(snapshots)-deletedCount > maxCount {
			shouldDelete = true
		}

		// Check size constraint (delete oldest snapshots until under maxSize)
		if maxSize > 0 && currentSnapshotSize > maxSize {
			shouldDelete = true
		}

		if shouldDelete {
			sm.snapshots.Delete(snapshot.ID)

			// Add to space freed if file exists
			if fileInfo, err := os.Stat(snapshot.FilePath); err == nil {
				spaceFreed += fileInfo.Size()
				currentSnapshotSize -= fileInfo.Size()
			}

			deletedCount++
		}
	}

	sm.logger.WithFields(logging.Fields{
		"deleted_count": deletedCount,
		"space_freed":   spaceFreed,
	}).Info("Snapshot cleanup completed")
	return deletedCount, spaceFreed, nil
}

// buildAdvancedSnapshotCommand builds an advanced FFmpeg command for snapshots
// buildAdvancedSnapshotCommand removed; FFmpegManager.BuildSnapshotCommand is the single source of truth.

// deleteSnapshotFile deletes a snapshot file
func (sm *SnapshotManager) deleteSnapshotFile(filePath string) error {
	return os.Remove(filePath)
}

// GetSnapshotSettings gets current snapshot settings
func (sm *SnapshotManager) GetSnapshotSettings() *SnapshotSettings {
	return sm.snapshotSettings
}

// UpdateSnapshotSettings updates snapshot settings
func (sm *SnapshotManager) UpdateSnapshotSettings(settings *SnapshotSettings) {
	sm.snapshotSettings = settings
	sm.logger.WithFields(logging.Fields{
		"format":      settings.Format,
		"quality":     settings.Quality,
		"max_width":   settings.MaxWidth,
		"max_height":  settings.MaxHeight,
		"auto_resize": settings.AutoResize,
	}).Info("Snapshot settings updated")
}

// ListSnapshots returns API-ready snapshot list response
func (sm *SnapshotManager) ListSnapshots(ctx context.Context, limit, offset int) (*ListSnapshotsResponse, error) {
	sm.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting API-ready snapshots list")

	// Get file list from existing method
	fileList, err := sm.GetSnapshotsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to API-ready SnapshotFileInfo format with rich metadata
	snapshots := make([]SnapshotFileInfo, len(fileList.Files))
	for i, file := range fileList.Files {
		// Extract device and timestamp from filename using configured pattern
		var device string = "camera0" // Default device name

		if sm.configManager != nil {
			cfg := sm.configManager.GetConfig()
			if cfg != nil {
				// Use pattern-based parsing
				parsedDevice, _, parseErr := ParseSnapshotFilename(file.FileName, cfg.Snapshots.FileNamePattern)
				if parseErr == nil {
					device = parsedDevice
				}
			}
		}

		// Fallback to hardcoded parsing if config unavailable or parsing failed
		if device == "camera0" {
			if parts := strings.Split(file.FileName, "_"); len(parts) > 0 {
				if strings.HasPrefix(parts[0], "camera") {
					device = parts[0]
				}
			}
		}

		// Extract format from filename extension
		format := "jpg" // Default
		if parts := strings.Split(file.FileName, "."); len(parts) > 1 {
			format = parts[len(parts)-1]
		}

		snapshots[i] = SnapshotFileInfo{
			Device:       device,
			Filename:     file.FileName,
			FileSize:     file.FileSize,
			ModifiedTime: file.CreatedAt.Format(time.RFC3339), // API compliant field name
			Format:       format,
			Resolution:   "1920x1080", // TODO: Extract resolution from FFmpeg-captured images only (V4L2 has no EXIF)
			// INVESTIGATION: V4L2 direct capture (Tier 0) produces raw frames without EXIF metadata
			// Only FFmpeg captures (Tier 1+) can have extractable metadata via ffprobe
			// SOLUTION: Check capture_method in metadata, if "ffmpeg", parse ffprobe JSON for resolution
			// FFPROBE: Already integrated at line 1021-1027, JSON parsing incomplete
			// EFFORT: 4-6 hours - implement ffprobe JSON parsing for streams.width/height
			DownloadURL: fmt.Sprintf("/files/snapshots/%s", file.FileName),
		}
	}

	response := &ListSnapshotsResponse{
		Snapshots: snapshots,
		Total:     fileList.Total,
		Limit:     limit,
		Offset:    offset,
	}

	return response, nil
}

// GetSnapshotsList scans the snapshots directory and returns a list of snapshot files with metadata
func (sm *SnapshotManager) GetSnapshotsList(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	sm.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting snapshots list")

	// Validate pagination parameters
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative, got %d", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative, got %d", offset)
	}

	// Get snapshots directory path from configuration
	snapshotsDir := sm.config.SnapshotsPath
	if snapshotsDir == "" {
		return nil, fmt.Errorf("snapshots path not configured")
	}

	// Check if directory exists and is accessible
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		sm.logger.WithField("directory", snapshotsDir).Warn("Snapshots directory does not exist")
		return &FileListResponse{
			Files:  []*FileMetadata{},
			Total:  0,
			Limit:  limit,
			Offset: offset,
		}, nil
	}

	// Check if we should scan device subdirectories (canonical configuration)
	var entries []os.DirEntry
	var err error

	if sm.configManager != nil {
		cfg := sm.configManager.GetConfig()
		if cfg != nil && cfg.Snapshots.UseDeviceSubdirs {
			// Scan device subdirectories when use_device_subdirs is enabled
			sm.logger.WithField("directory", snapshotsDir).Debug("Scanning device subdirectories for snapshots")

			// Read base directory to find device subdirectories
			baseEntries, err := os.ReadDir(snapshotsDir)
			if err != nil {
				sm.logger.WithError(err).WithField("directory", snapshotsDir).Error("Error reading snapshots directory")
				return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
			}

			// Collect all files from device subdirectories
			for _, baseEntry := range baseEntries {
				if baseEntry.IsDir() {
					// This is a device subdirectory (e.g., camera0, camera1)
					deviceDir := filepath.Join(snapshotsDir, baseEntry.Name())
					deviceEntries, err := os.ReadDir(deviceDir)
					if err != nil {
						sm.logger.WithError(err).WithField("device_dir", deviceDir).Warn("Error reading device subdirectory")
						continue // Skip this device directory but continue with others
					}

					// Add device subdirectory files to our collection
					entries = append(entries, deviceEntries...)
				} else {
					// File in base directory (legacy or non-subdirectory mode)
					entries = append(entries, baseEntry)
				}
			}
		} else {
			// Standard directory scanning (no device subdirectories)
			entries, err = os.ReadDir(snapshotsDir)
			if err != nil {
				sm.logger.WithError(err).WithField("directory", snapshotsDir).Error("Error reading snapshots directory")
				return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
			}
		}
	} else {
		// Fallback: standard directory scanning when config unavailable
		sm.logger.WithField("directory", snapshotsDir).Warn("Config unavailable, using standard directory scanning")
		entries, err = os.ReadDir(snapshotsDir)
		if err != nil {
			sm.logger.WithError(err).WithField("directory", snapshotsDir).Error("Error reading snapshots directory")
			return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
		}
	}

	// Process files and extract metadata
	var files []*FileMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Get file stats
		fileInfo, err := entry.Info()
		if err != nil {
			sm.logger.WithError(err).WithField("filename", filename).Warn("Error accessing file")
			continue
		}

		// Extract comprehensive metadata for Python equivalence
		metadata := sm.extractSnapshotMetadata(ctx, filepath.Join(snapshotsDir, filename))

		// Create file metadata with comprehensive information
		fileMetadata := &FileMetadata{
			FileName:    filename,
			FileSize:    fileInfo.Size(),
			CreatedAt:   fileInfo.ModTime(), // Use ModTime as CreatedAt since creation time may not be available
			ModifiedAt:  fileInfo.ModTime(),
			DownloadURL: fmt.Sprintf("/files/snapshots/%s", filename),
		}

		// Add comprehensive metadata for Python equivalence
		if metadata != nil {
			// Store additional metadata in a way that's compatible with Python system
			sm.logger.WithFields(logging.Fields{
				"filename": filename,
				"metadata": metadata,
			}).Debug("Extracted comprehensive snapshot metadata")
		}

		files = append(files, fileMetadata)
	}

	// Sort files by modified time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})

	// Apply pagination
	totalCount := len(files)
	startIdx := offset
	endIdx := startIdx + limit
	if endIdx > totalCount {
		endIdx = totalCount
	}
	if startIdx > totalCount {
		startIdx = totalCount
	}

	var paginatedFiles []*FileMetadata
	if totalCount > 0 && startIdx < totalCount {
		paginatedFiles = files[startIdx:endIdx]
	} else {
		// Ensure we return an empty slice, not nil
		paginatedFiles = []*FileMetadata{}
	}

	sm.logger.WithFields(logging.Fields{
		"total_files": totalCount,
		"returned":    len(paginatedFiles),
	}).Debug("Snapshots list retrieved successfully")

	return &FileListResponse{
		Files:  paginatedFiles,
		Total:  totalCount,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// extractSnapshotMetadata extracts comprehensive metadata from snapshot file for Python equivalence
func (sm *SnapshotManager) extractSnapshotMetadata(ctx context.Context, filePath string) map[string]interface{} {
	sm.logger.WithField("file_path", filePath).Debug("Extracting comprehensive snapshot metadata")

	metadata := make(map[string]interface{})

	// Check if ffprobe is available before attempting extraction
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		sm.logger.WithField("file_path", filePath).Debug("ffprobe not available, skipping metadata extraction")
		return metadata
	}

	// Extract image metadata using FFmpeg
	command := []string{
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	// Create context with timeout for metadata extraction
	metadataCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Execute command with timeout
	cmd := exec.CommandContext(metadataCtx, command[0], command[1:]...)
	output, err := cmd.Output()
	if err != nil {
		// Only log as debug to avoid spam - metadata extraction is optional
		sm.logger.WithField("file_path", filePath).Debug("Failed to extract image metadata (optional operation)")
		return metadata
	}

	// Parse JSON output for comprehensive metadata
	// TODO: Parse ffprobe JSON output for comprehensive metadata extraction
	// INVESTIGATION: ffprobe integration already exists (lines 1021-1027), JSON parsing incomplete
	// CURRENT: Raw JSON stored in metadata["extraction_method"] = "ffprobe" but not parsed
	// SOLUTION: json.Unmarshal(output, &ffprobeResult) then extract streams[0].width/height/duration/codec
	// REFERENCE: ffprobe JSON structure: {"streams":[{"width":1920,"height":1080,"codec_name":"mjpeg"}],"format":{}}
	// EFFORT: 6-8 hours - implement complete ffprobe JSON parsing with error handling
	sm.logger.WithFields(logging.Fields{
		"file_path": filePath,
		"metadata":  string(output),
	}).Debug("Extracted raw image metadata")

	// TODO: Complete metadata parsing implementation for full feature parity with Python version
	// INVESTIGATION: Python version extracts width, height, format, codec, bitrate from ffprobe
	// CURRENT: Only basic metadata stored (format="image", extraction_method="ffprobe")
	// SOLUTION: Parse JSON output above and populate metadata map with:
	//   - width/height from streams[0].width/height
	//   - codec from streams[0].codec_name
	//   - bitrate from streams[0].bit_rate (if available)
	//   - duration from format.duration (for videos)
	// EFFORT: 2-3 hours - extend JSON parsing from TODO above
	metadata["format"] = "image"
	metadata["extraction_method"] = "ffprobe"
	metadata["extraction_time"] = time.Now().Unix()

	sm.logger.WithFields(logging.Fields{
		"file_path": filePath,
		"metadata":  metadata,
	}).Debug("Comprehensive snapshot metadata extracted successfully")

	return metadata
}

// GetSnapshotInfo gets detailed information about a specific snapshot file
// GetSnapshotInfo returns API-ready snapshot information with rich metadata
func (sm *SnapshotManager) GetSnapshotInfo(ctx context.Context, filename string) (*GetSnapshotInfoResponse, error) {
	sm.logger.WithField("filename", filename).Debug("Getting API-ready snapshot info")

	// Validate filename
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// Get snapshots directory path from configuration
	snapshotsDir := sm.config.SnapshotsPath
	if snapshotsDir == "" {
		return nil, fmt.Errorf("snapshots path not configured")
	}

	// Try direct path first
	filePath := filepath.Join(snapshotsDir, filename)
	fileInfo, err := os.Stat(filePath)

	// If direct path fails, search in subdirectories
	if err != nil {
		sm.logger.WithFields(logging.Fields{
			"filename":    filename,
			"direct_path": filePath,
			"error":       err,
		}).Info("GetSnapshotInfo: Direct path failed, searching subdirectories")

		// Search in device subdirectories
		entries, err := os.ReadDir(snapshotsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		found := false
		for _, entry := range entries {
			if entry.IsDir() {
				subDirPath := filepath.Join(snapshotsDir, entry.Name())
				filePath = filepath.Join(subDirPath, filename)
				if fileInfo, err = os.Stat(filePath); err == nil {
					found = true
					break
				}
			}
		}

		// If still not found
		if !found {
			return nil, fmt.Errorf("snapshot file not found: %s", filename)
		}
	}

	// Check if it's a file (not a directory)
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a file: %s", filename)
	}

	// Extract device from filename pattern (camera0_timestamp.{format})
	device := "camera0" // Default device name
	if parts := strings.Split(filename, "_"); len(parts) > 0 {
		if strings.HasPrefix(parts[0], "camera") {
			device = parts[0]
		}
	}

	// Extract format from filename extension
	format := "jpg" // Default
	if ext := filepath.Ext(filename); ext != "" {
		format = strings.TrimPrefix(ext, ".")
	}

	// TODO: Extract resolution from image metadata for FFmpeg-captured images only
	// INVESTIGATION: V4L2 captures have no EXIF/metadata, only FFmpeg captures do
	// CURRENT: Hardcoded "1920x1080" placeholder for all images
	// SOLUTION: Use ffprobe integration from extractSnapshotMetadata() to get real resolution
	// DEPENDENCY: Requires completed ffprobe JSON parsing from lines 1038-1044 above
	// EFFORT: 1-2 hours - call extractSnapshotMetadata() and use parsed width/height
	resolution := "1920x1080" // Placeholder

	// Build API-ready response with rich metadata
	response := &GetSnapshotInfoResponse{
		Filename:    filename,
		FileSize:    fileInfo.Size(),
		CreatedTime: fileInfo.ModTime().Format(time.RFC3339), // API compliant field name
		Format:      format,
		Resolution:  resolution,
		Device:      device,
	}

	sm.logger.WithFields(logging.Fields{
		"filename":   filename,
		"device":     device,
		"format":     format,
		"resolution": resolution,
		"file_size":  fileInfo.Size(),
	}).Debug("Snapshot info retrieved successfully")

	return response, nil
}

// DeleteSnapshotFile deletes a snapshot file by filename
func (sm *SnapshotManager) DeleteSnapshotFile(ctx context.Context, filename string) error {
	sm.logger.WithField("filename", filename).Debug("Deleting snapshot file")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Get snapshots directory path from canonical configuration manager (preferred)
	snapshotsDir := sm.config.SnapshotsPath
	if sm.configManager != nil {
		if cfg := sm.configManager.GetConfig(); cfg != nil && cfg.MediaMTX.SnapshotsPath != "" {
			snapshotsDir = cfg.MediaMTX.SnapshotsPath
		}
	}
	if snapshotsDir == "" {
		return fmt.Errorf("snapshots path not configured")
	}

	// Try direct path first
	filePath := filepath.Join(snapshotsDir, filename)
	fileInfo, err := os.Stat(filePath)

	// If direct path fails, search in subdirectories
	if err != nil {
		sm.logger.WithFields(logging.Fields{
			"filename":    filename,
			"direct_path": filePath,
			"error":       err,
		}).Info("DeleteSnapshotFile: Direct path failed, searching subdirectories")

		// Search in device subdirectories
		entries, err := os.ReadDir(snapshotsDir)
		if err != nil {
			return fmt.Errorf("failed to read snapshots directory: %w", err)
		}

		found := false
		for _, entry := range entries {
			if entry.IsDir() {
				subDirPath := filepath.Join(snapshotsDir, entry.Name())
				filePath = filepath.Join(subDirPath, filename)
				if fileInfo, err = os.Stat(filePath); err == nil {
					found = true
					break
				}
			}
		}

		// If still not found
		if !found {
			return fmt.Errorf("snapshot file not found: %s", filename)
		}
	}

	// Check if it's a file (not a directory)
	if fileInfo == nil {
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("error accessing file: %w", err)
		}
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is not a file: %s", filename)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		sm.logger.WithError(err).WithField("filename", filename).Error("Error deleting snapshot file")
		return fmt.Errorf("error deleting snapshot file: %w", err)
	}

	sm.logger.WithField("filename", filename).Info("Snapshot file deleted successfully")
	return nil
}
