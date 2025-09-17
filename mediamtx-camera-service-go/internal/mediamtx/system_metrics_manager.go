/*
System Metrics Manager Implementation

Provides system-wide resource monitoring and metrics collection for API responses.
Centralizes all system-level monitoring logic that was previously scattered in Controller.

Requirements Coverage:
- REQ-MTX-004: Health monitoring and system metrics
- REQ-API-001: JSON-RPC API compliance for metrics endpoints

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"golang.org/x/sys/unix"
)

// SystemMetricsManager manages system-wide resource monitoring and metrics collection.
//
// RESPONSIBILITIES:
// - System resource monitoring (CPU, memory, disk, goroutines)
// - Storage space calculations and health assessment
// - Cross-component metrics aggregation
// - API-ready response formatting for system metrics endpoints
//
// SCOPE:
// - Handles all system-level resource monitoring
// - Manages storage operations and file system monitoring
// - Aggregates metrics from multiple system components
// - Does NOT handle MediaMTX-specific connectivity (that's HealthMonitor)
//
// API INTEGRATION:
// - Returns JSON-RPC API-ready responses
// - Provides rich system metrics with proper formatting
type SystemMetricsManager struct {
	config            *config.Config
	recordingConfig   *config.RecordingConfig
	configIntegration *ConfigIntegration
	logger            *logging.Logger

	// Dependencies for metrics aggregation
	recordingManager *RecordingManager
	cameraMonitor    interface{ GetMetrics() map[string]interface{} }
	streamManager    StreamManager

	// System start time for uptime calculations
	startTime time.Time
}

// NewSystemMetricsManager creates a new system metrics manager
func NewSystemMetricsManager(
	config *config.Config,
	recordingConfig *config.RecordingConfig,
	configIntegration *ConfigIntegration,
	logger *logging.Logger,
) *SystemMetricsManager {
	return &SystemMetricsManager{
		config:            config,
		recordingConfig:   recordingConfig,
		configIntegration: configIntegration,
		logger:            logger,
		startTime:         time.Now(),
	}
}

// SetDependencies sets the required dependencies for metrics aggregation
func (sm *SystemMetricsManager) SetDependencies(
	recordingManager *RecordingManager,
	cameraMonitor interface{ GetMetrics() map[string]interface{} },
	streamManager StreamManager,
) {
	sm.recordingManager = recordingManager
	sm.cameraMonitor = cameraMonitor
	sm.streamManager = streamManager
}

// GetStorageInfoAPI returns storage information in API-ready format
func (sm *SystemMetricsManager) GetStorageInfoAPI(ctx context.Context) (*GetStorageInfoResponse, error) {
	sm.logger.Debug("Collecting system storage information")

	// Get recordings path from configuration
	recordingsPath := sm.recordingConfig.RecordingsPath
	if recordingsPath == "" {
		return nil, fmt.Errorf("recordings path not configured")
	}

	// Perform file system operations
	var st unix.Statfs_t
	if err := unix.Statfs(recordingsPath, &st); err != nil {
		return nil, fmt.Errorf("failed to get storage statistics: %w", err)
	}

	// Calculate storage metrics
	totalSpace := int64(st.Blocks * uint64(st.Bsize))
	freeSpace := int64(st.Bfree * uint64(st.Bsize))
	usedSpace := totalSpace - freeSpace
	availableSpace := int64(st.Bavail * uint64(st.Bsize)) // Available to non-root users

	usagePercent := 0.0
	if totalSpace > 0 {
		usagePercent = float64(usedSpace) / float64(totalSpace) * 100.0
	}

	// Get directory sizes (aggregate from managers to avoid FS walking)
	recordingsSize := int64(0)
	snapshotsSize := int64(0)

	// Get recordings size from RecordingManager
	if sm.recordingManager != nil {
		if recList, err := sm.recordingManager.GetRecordingsList(ctx, 100000, 0); err == nil {
			for _, file := range recList.Files {
				recordingsSize += file.FileSize
			}
		}
	}

	// Get snapshots size from configuration
	snapshotsPath := sm.config.Camera.SnapshotsPath
	if snapshotsPath != "" {
		if entries, err := os.ReadDir(snapshotsPath); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					if info, err := entry.Info(); err == nil {
						snapshotsSize += info.Size()
					}
				}
			}
		}
	}

	// Build API-ready response
	response := &GetStorageInfoResponse{
		TotalSpace:     totalSpace,
		UsedSpace:      usedSpace,
		AvailableSpace: availableSpace,
		UsagePercent:   usagePercent,
		RecordingsSize: recordingsSize,
		SnapshotsSize:  snapshotsSize,
	}

	sm.logger.WithFields(logging.Fields{
		"total_space_gb":     float64(totalSpace) / (1024 * 1024 * 1024),
		"used_space_gb":      float64(usedSpace) / (1024 * 1024 * 1024),
		"usage_percent":      usagePercent,
		"recordings_size_mb": float64(recordingsSize) / (1024 * 1024),
		"snapshots_size_mb":  float64(snapshotsSize) / (1024 * 1024),
	}).Debug("Storage information collected successfully")

	return response, nil
}

// GetSystemMetricsAPI returns system metrics in API-ready format
func (sm *SystemMetricsManager) GetSystemMetricsAPI(ctx context.Context) (*GetSystemMetricsResponse, error) {
	sm.logger.Debug("Collecting system performance metrics")

	// Collect system resource metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate CPU usage (simplified - could be enhanced with proper CPU monitoring)
	cpuUsage := 0.0 // TODO-IMPL: Implement proper CPU usage calculation

	// Calculate memory usage percentage
	memUsage := float64(memStats.Alloc) / float64(memStats.Sys) * 100.0

	// Get disk usage from storage info
	storageInfo, err := sm.GetStorageInfoAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info for system metrics: %w", err)
	}

	// Get goroutine count
	goroutineCount := runtime.NumGoroutine()

	// Build API-ready response
	response := &GetSystemMetricsResponse{
		CPUUsage:    cpuUsage,
		MemoryUsage: memUsage,
		DiskUsage:   storageInfo.UsagePercent,
		Goroutines:  goroutineCount,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	sm.logger.WithFields(logging.Fields{
		"cpu_usage":    cpuUsage,
		"memory_usage": memUsage,
		"disk_usage":   storageInfo.UsagePercent,
		"goroutines":   goroutineCount,
	}).Debug("System metrics collected successfully")

	return response, nil
}

// GetMetricsAPI aggregates metrics from all components and returns API-ready format
func (sm *SystemMetricsManager) GetMetricsAPI(ctx context.Context) (*GetMetricsResponse, error) {
	sm.logger.Debug("Aggregating comprehensive system metrics")

	// Get system metrics
	systemMetrics, err := sm.GetSystemMetricsAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Build system metrics map
	systemMetricsMap := map[string]interface{}{
		"cpu_usage":    systemMetrics.CPUUsage,
		"memory_usage": systemMetrics.MemoryUsage,
		"disk_usage":   systemMetrics.DiskUsage,
		"goroutines":   systemMetrics.Goroutines,
	}

	// Build camera metrics map
	cameraMetrics := make(map[string]interface{})
	if sm.cameraMonitor != nil {
		cameraMetrics = sm.cameraMonitor.GetMetrics()
	}

	// Build recording metrics map
	recordingMetrics := make(map[string]interface{})
	// TODO-IMPL: Add recording-specific metrics from RecordingManager

	// Build stream metrics map
	streamMetrics := make(map[string]interface{})
	if sm.streamManager != nil {
		if streams, err := sm.streamManager.ListStreams(ctx); err == nil {
			activeStreams := 0
			totalViewers := 0
			for _, stream := range streams.Streams {
				if stream.Status == "active" {
					activeStreams++
				}
				totalViewers += stream.Viewers
			}
			streamMetrics["active_streams"] = activeStreams
			streamMetrics["total_streams"] = len(streams.Streams)
			streamMetrics["total_viewers"] = totalViewers
		}
	}

	// Build API-ready response
	response := &GetMetricsResponse{
		Timestamp:        time.Now().Format(time.RFC3339),
		SystemMetrics:    systemMetricsMap,
		CameraMetrics:    cameraMetrics,
		RecordingMetrics: recordingMetrics,
		StreamMetrics:    streamMetrics,
	}

	return response, nil
}
