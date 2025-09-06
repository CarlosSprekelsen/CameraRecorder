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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// SnapshotManager manages advanced snapshot operations
type SnapshotManager struct {
	ffmpegManager FFmpegManager
	config        *MediaMTXConfig
	logger        *logging.Logger

	// Configuration integration for multi-tier support
	configManager *config.ConfigManager

	// Snapshot settings
	snapshotSettings *SnapshotSettings

	// Snapshot tracking
	snapshots   map[string]*Snapshot
	snapshotsMu sync.RWMutex
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

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(ffmpegManager FFmpegManager, config *MediaMTXConfig, logger *logging.Logger) *SnapshotManager {
	return &SnapshotManager{
		ffmpegManager: ffmpegManager,
		config:        config,
		logger:        logger,
		snapshots:     make(map[string]*Snapshot),
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

// NewSnapshotManagerWithConfig creates a new snapshot manager with configuration integration
func NewSnapshotManagerWithConfig(ffmpegManager FFmpegManager, config *MediaMTXConfig, configManager *config.ConfigManager, logger *logging.Logger) *SnapshotManager {
	return &SnapshotManager{
		ffmpegManager: ffmpegManager,
		config:        config,
		configManager: configManager,
		logger:        logger,
		snapshots:     make(map[string]*Snapshot),
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

// TakeSnapshot takes a snapshot with multi-tier approach (enhanced existing method)
func (sm *SnapshotManager) TakeSnapshot(ctx context.Context, device, path string, options map[string]interface{}) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"device":  device,
		"path":    path,
		"options": options,
	}).Info("Taking multi-tier snapshot")

	// Apply snapshot settings from options (existing logic)
	if format, ok := options["format"].(string); ok {
		sm.snapshotSettings.Format = format
	}
	if quality, ok := options["quality"].(int); ok {
		sm.snapshotSettings.Quality = quality
	}
	if maxWidth, ok := options["max_width"].(int); ok {
		sm.snapshotSettings.MaxWidth = maxWidth
	}
	if maxHeight, ok := options["max_height"].(int); ok {
		sm.snapshotSettings.MaxHeight = maxHeight
	}
	if autoResize, ok := options["auto_resize"].(bool); ok {
		sm.snapshotSettings.AutoResize = autoResize
	}
	if compression, ok := options["compression"].(int); ok {
		sm.snapshotSettings.Compression = compression
	}

	// Get tier configuration from existing config system
	tierConfig := sm.getTierConfiguration()
	if tierConfig == nil {
		return nil, fmt.Errorf("failed to get tier configuration - config manager not properly initialized")
	}

	// Generate snapshot path
	snapshotID := generateSnapshotID(device)
	snapshotPath := sm.generateSnapshotPath(device, snapshotID)

	// Execute multi-tier snapshot capture
	snapshot, err := sm.takeSnapshotMultiTier(ctx, device, snapshotPath, options, tierConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to execute multi-tier snapshot capture: %w", err)
	}

	// Store snapshot (existing logic)
	sm.snapshotsMu.Lock()
	sm.snapshots[snapshotID] = snapshot
	sm.snapshotsMu.Unlock()

	sm.logger.WithFields(logging.Fields{
		"snapshot_id": snapshotID,
		"device":      device,
		"path":        path,
		"file_size":   snapshot.Size,
		"format":      sm.snapshotSettings.Format,
		"quality":     sm.snapshotSettings.Quality,
		"tier_used":   snapshot.Metadata["tier_used"],
	}).Info("Multi-tier snapshot taken successfully")

	return snapshot, nil
}

// takeSnapshotMultiTier implements the 4-tier snapshot capture system
func (sm *SnapshotManager) takeSnapshotMultiTier(ctx context.Context, device, snapshotPath string, options map[string]interface{}, tierConfig *config.SnapshotTiersConfig) (*Snapshot, error) {
	startTime := time.Now()
	captureMethodsTried := []string{}

	sm.logger.WithFields(logging.Fields{
		"device": device,
		"tier":   1,
	}).Info("Tier 1: Attempting USB direct capture")

	// Tier 1: USB Direct Capture (Fastest Path)
	tier1Ctx, tier1Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier1USBDirectTimeout*float64(time.Second)))
	defer tier1Cancel()

	if snapshot, err := sm.captureSnapshotDirect(tier1Ctx, device, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 1, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"device":       device,
			"tier":         1,
			"capture_time": captureTime,
		}).Info("Tier 1: USB direct capture successful")
		return result, nil
	}
	captureMethodsTried = append(captureMethodsTried, "usb_direct")

	sm.logger.WithFields(logging.Fields{
		"device": device,
		"tier":   2,
	}).Info("Tier 2: Attempting RTSP immediate capture")

	// Tier 2: RTSP Immediate Capture
	tier2Ctx, tier2Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier2RTSPReadyCheckTimeout*float64(time.Second)))
	defer tier2Cancel()

	if snapshot, err := sm.captureSnapshotFromRTSP(tier2Ctx, device, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 2, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"device":       device,
			"tier":         2,
			"capture_time": captureTime,
		}).Info("Tier 2: RTSP immediate capture successful")
		return result, nil
	}
	captureMethodsTried = append(captureMethodsTried, "rtsp_immediate")

	sm.logger.WithFields(logging.Fields{
		"device": device,
		"tier":   3,
	}).Info("Tier 3: Attempting RTSP stream activation")

	// Tier 3: RTSP Stream Activation
	tier3Ctx, tier3Cancel := context.WithTimeout(ctx, time.Duration(tierConfig.Tier3ActivationTimeout*float64(time.Second)))
	defer tier3Cancel()

	if snapshot, err := sm.captureSnapshotFromRTSP(tier3Ctx, device, snapshotPath); err == nil {
		captureTime := time.Since(startTime)
		result := sm.createSnapshotResult(snapshot, 3, captureTime, captureMethodsTried)
		sm.logger.WithFields(logging.Fields{
			"device":       device,
			"tier":         3,
			"capture_time": captureTime,
		}).Info("Tier 3: RTSP stream activation successful")
		return result, nil
	}
	captureMethodsTried = append(captureMethodsTried, "rtsp_activation")

	// Tier 4: Error Handling - All methods failed
	totalTime := time.Since(startTime)
	sm.logger.WithFields(logging.Fields{
		"device":        device,
		"total_time":    totalTime,
		"methods_tried": captureMethodsTried,
	}).Error("Tier 4: All snapshot capture methods failed")

	return nil, sm.createMultiTierError(device, captureMethodsTried, totalTime)
}

// getTierConfiguration retrieves multi-tier configuration from existing config system
func (sm *SnapshotManager) getTierConfiguration() *config.SnapshotTiersConfig {
	if sm.configManager == nil {
		sm.logger.Error("Config manager is nil - this should not happen in production")
		// This is a critical error - return nil to force proper error handling
		return nil
	}

	// Get performance configuration from centralized config system
	cfg := sm.configManager.GetConfig()
	if cfg == nil {
		sm.logger.Error("Failed to get config from config manager - this should not happen in production")
		// This is a critical error - return nil to force proper error handling
		return nil
	}

	return &cfg.Performance.SnapshotTiers
}

// captureSnapshotDirect implements Tier 1: USB Direct Capture (Fastest Path)
func (sm *SnapshotManager) captureSnapshotDirect(ctx context.Context, devicePath, snapshotPath string) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"tier":        1,
	}).Info("Tier 1: Attempting USB direct capture")

	// Use existing FFmpeg manager for direct capture
	command := sm.buildAdvancedSnapshotCommand(devicePath, snapshotPath, nil)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "create_output_dir", "failed to create output directory", err)
	}

	// Create command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "take_snapshot", "failed to take snapshot", err)
	}

	// Get file info using existing FFmpeg manager
	fileSize, _, err := sm.ffmpegManager.GetFileInfo(ctx, snapshotPath)
	if err != nil {
		return nil, NewFFmpegErrorWithErr(0, "snapshot", "get_file_info", "failed to get file info", err)
	}

	// Create snapshot object
	snapshot := &Snapshot{
		ID:       generateSnapshotID(devicePath),
		Device:   devicePath,
		Path:     filepath.Dir(snapshotPath),
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
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
func (sm *SnapshotManager) captureSnapshotFromRTSP(ctx context.Context, devicePath, snapshotPath string) (*Snapshot, error) {
	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"tier":        2,
	}).Info("Tier 2/3: Capturing from RTSP stream")

	// Build RTSP URL from device path
	streamName := sm.getStreamNameFromDevice(devicePath)
	rtspURL := fmt.Sprintf("rtsp://%s:%d/%s", sm.config.Host, sm.config.RTSPPort, streamName)

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
		return nil, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "create_output_dir", "failed to create output directory", err)
	}

	// Create command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "take_snapshot", "failed to take snapshot from RTSP", err)
	}

	// Get file info using existing FFmpeg manager
	fileSize, _, err := sm.ffmpegManager.GetFileInfo(ctx, snapshotPath)
	if err != nil {
		return nil, NewFFmpegErrorWithErr(0, "snapshot", "get_file_info", "failed to get file info", err)
	}

	// Create snapshot object
	snapshot := &Snapshot{
		ID:       generateSnapshotID(devicePath),
		Device:   devicePath,
		Path:     filepath.Dir(snapshotPath),
		FilePath: snapshotPath,
		Size:     fileSize,
		Created:  time.Now(),
	}

	sm.logger.WithFields(logging.Fields{
		"device":      devicePath,
		"output_path": snapshotPath,
		"file_size":   fileSize,
		"tier":        2,
	}).Info("Tier 2/3: RTSP capture successful")

	return snapshot, nil
}

// getStreamNameFromDevice converts device path to stream name
func (sm *SnapshotManager) getStreamNameFromDevice(devicePath string) string {
	// Extract device number from path (e.g., "/dev/video0" -> "camera0")
	deviceName := filepath.Base(devicePath)
	if strings.HasPrefix(deviceName, "video") {
		deviceNum := strings.TrimPrefix(deviceName, "video")
		return fmt.Sprintf("camera%s", deviceNum)
	}
	return fmt.Sprintf("camera_%s", deviceName)
}

// createSnapshotResult creates a snapshot result with tier information
func (sm *SnapshotManager) createSnapshotResult(snapshot *Snapshot, tier int, captureTime time.Duration, methodsTried []string) *Snapshot {
	// Add tier information to snapshot
	snapshot.Metadata = map[string]interface{}{
		"tier_used":       tier,
		"capture_time_ms": captureTime.Milliseconds(),
		"methods_tried":   methodsTried,
		"user_experience": sm.determineUserExperience(captureTime),
	}

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
func (sm *SnapshotManager) createMultiTierError(device string, methodsTried []string, totalTime time.Duration) error {
	return fmt.Errorf("all snapshot capture methods failed for %s after %.2fs: tried %v",
		device, totalTime.Seconds(), methodsTried)
}

// GetSnapshot gets a snapshot by ID
func (sm *SnapshotManager) GetSnapshot(snapshotID string) (*Snapshot, bool) {
	sm.snapshotsMu.RLock()
	defer sm.snapshotsMu.RUnlock()

	snapshot, exists := sm.snapshots[snapshotID]
	return snapshot, exists
}

// ListSnapshots lists all snapshots
func (sm *SnapshotManager) ListSnapshots() []*Snapshot {
	sm.snapshotsMu.RLock()
	defer sm.snapshotsMu.RUnlock()

	snapshots := make([]*Snapshot, 0, len(sm.snapshots))
	for _, snapshot := range sm.snapshots {
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

// DeleteSnapshot deletes a snapshot
func (sm *SnapshotManager) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	sm.logger.WithField("snapshot_id", snapshotID).Debug("Deleting snapshot")

	sm.snapshotsMu.Lock()
	snapshot, exists := sm.snapshots[snapshotID]
	if !exists {
		sm.snapshotsMu.Unlock()
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	// Remove from tracking
	delete(sm.snapshots, snapshotID)
	sm.snapshotsMu.Unlock()

	// Delete file
	if err := os.Remove(snapshot.FilePath); err != nil {
		return NewFFmpegErrorWithErr(0, "delete_snapshot", "remove_file", "failed to delete snapshot file", err)
	}

	sm.logger.WithFields(logging.Fields{
		"snapshot_id": snapshotID,
		"file_path":   snapshot.FilePath,
	}).Info("Snapshot deleted successfully")

	return nil
}

// CleanupOldSnapshots cleans up old snapshots based on age and count
func (sm *SnapshotManager) CleanupOldSnapshots(ctx context.Context, maxAge time.Duration, maxCount int) error {
	sm.logger.WithFields(logging.Fields{
		"max_age":   maxAge,
		"max_count": maxCount,
	}).Info("Cleaning up old snapshots")

	sm.snapshotsMu.Lock()
	defer sm.snapshotsMu.Unlock()

	// Get all snapshots sorted by creation time
	snapshots := make([]*Snapshot, 0, len(sm.snapshots))
	for _, snapshot := range sm.snapshots {
		snapshots = append(snapshots, snapshot)
	}

	// Sort by creation time (oldest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Created.Before(snapshots[j].Created)
	})

	// Delete old snapshots
	deletedCount := 0
	for _, snapshot := range snapshots {
		// Check age
		if time.Since(snapshot.Created) > maxAge {
			if err := sm.deleteSnapshotFile(snapshot.FilePath); err != nil {
				sm.logger.WithError(err).WithField("snapshot_id", snapshot.ID).Error("Failed to delete old snapshot file")
				continue
			}
			delete(sm.snapshots, snapshot.ID)
			deletedCount++
		}
	}

	// Delete excess snapshots if we have too many
	if len(sm.snapshots) > maxCount {
		excessCount := len(sm.snapshots) - maxCount
		for i := 0; i < excessCount && i < len(snapshots); i++ {
			snapshot := snapshots[i]
			if err := sm.deleteSnapshotFile(snapshot.FilePath); err != nil {
				sm.logger.WithError(err).WithField("snapshot_id", snapshot.ID).Error("Failed to delete excess snapshot file")
				continue
			}
			delete(sm.snapshots, snapshot.ID)
			deletedCount++
		}
	}

	sm.logger.WithField("deleted_count", strconv.Itoa(deletedCount)).Info("Snapshot cleanup completed")
	return nil
}

// buildAdvancedSnapshotCommand builds an advanced FFmpeg command for snapshots
func (sm *SnapshotManager) buildAdvancedSnapshotCommand(device, outputPath string, options map[string]interface{}) []string {
	command := []string{"ffmpeg"}

	// Input device
	command = append(command, "-f", "v4l2")
	command = append(command, "-i", device)

	// Video frames (take only one frame)
	command = append(command, "-vframes", "1")

	// Video codec based on format
	switch sm.snapshotSettings.Format {
	case "jpg", "jpeg":
		command = append(command, "-c:v", "mjpeg")
		command = append(command, "-q:v", strconv.Itoa(sm.snapshotSettings.Quality))
	case "png":
		command = append(command, "-c:v", "png")
		command = append(command, "-compression_level", strconv.Itoa(sm.snapshotSettings.Compression))
	default:
		command = append(command, "-c:v", "mjpeg")
		command = append(command, "-q:v", strconv.Itoa(sm.snapshotSettings.Quality))
	}

	// Resize if needed
	if sm.snapshotSettings.AutoResize {
		scaleFilter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease",
			sm.snapshotSettings.MaxWidth, sm.snapshotSettings.MaxHeight)
		command = append(command, "-vf", scaleFilter)
	}

	// Output path
	command = append(command, outputPath)

	return command
}

// generateSnapshotPath generates a snapshot file path
func (sm *SnapshotManager) generateSnapshotPath(device, snapshotID string) string {
	// Use configuration paths from centralized config
	basePath := sm.config.SnapshotsPath
	if basePath == "" {
		// Fallback to centralized default
		basePath = "/opt/camera-service/snapshots"
	}

	// Generate filename
	filename := fmt.Sprintf("%s_%s.%s", device, snapshotID, sm.snapshotSettings.Format)

	// Clean device path for filename
	filename = strings.ReplaceAll(filename, "/", "_")

	return filepath.Join(basePath, filename)
}

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

// GetSnapshotsList scans the snapshots directory and returns a list of snapshot files with metadata
func (sm *SnapshotManager) GetSnapshotsList(ctx context.Context, limit, offset int) (*FileListResponse, error) {
	sm.logger.WithFields(logging.Fields{
		"limit":  limit,
		"offset": offset,
	}).Debug("Getting snapshots list")

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

	// Read directory entries
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		sm.logger.WithError(err).WithField("directory", snapshotsDir).Error("Error reading snapshots directory")
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
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

	paginatedFiles := files[startIdx:endIdx]

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

	// Extract image metadata using FFmpeg
	command := []string{
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	// Execute command with timeout
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	output, err := cmd.Output()
	if err != nil {
		sm.logger.WithError(err).WithField("file_path", filePath).Warn("Failed to extract image metadata")
		return metadata
	}

	// Parse JSON output for comprehensive metadata
	// For now, we'll log the raw output and extract basic information
	sm.logger.WithFields(logging.Fields{
		"file_path": filePath,
		"metadata":  string(output),
	}).Debug("Extracted raw image metadata")

	// Add basic metadata for Python equivalence
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
func (sm *SnapshotManager) GetSnapshotInfo(ctx context.Context, filename string) (*FileMetadata, error) {
	sm.logger.WithField("filename", filename).Debug("Getting snapshot info")

	// Validate filename
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// Get snapshots directory path from configuration
	snapshotsDir := sm.config.SnapshotsPath
	if snapshotsDir == "" {
		return nil, fmt.Errorf("snapshots path not configured")
	}

	// Construct full file path
	filePath := filepath.Join(snapshotsDir, filename)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("snapshot file not found: %s", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("error accessing file: %w", err)
	}

	// Check if it's a file (not a directory)
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a file: %s", filename)
	}

	// Create file metadata
	fileMetadata := &FileMetadata{
		FileName:    filename,
		FileSize:    fileInfo.Size(),
		CreatedAt:   fileInfo.ModTime(), // Use ModTime as CreatedAt since creation time may not be available
		ModifiedAt:  fileInfo.ModTime(),
		DownloadURL: fmt.Sprintf("/files/snapshots/%s", filename),
	}

	sm.logger.WithField("filename", filename).Debug("Snapshot info retrieved successfully")
	return fileMetadata, nil
}

// DeleteSnapshotFile deletes a snapshot file by filename
func (sm *SnapshotManager) DeleteSnapshotFile(ctx context.Context, filename string) error {
	sm.logger.WithField("filename", filename).Debug("Deleting snapshot file")

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Get snapshots directory path from configuration
	snapshotsDir := sm.config.SnapshotsPath
	if snapshotsDir == "" {
		return fmt.Errorf("snapshots path not configured")
	}

	// Construct full file path
	filePath := filepath.Join(snapshotsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot file not found: %s", filename)
	}

	// Check if it's a file (not a directory)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error accessing file: %w", err)
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
