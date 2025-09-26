/*
MediaMTX Configuration Integration

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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// ConfigIntegration provides integration between MediaMTX package and existing config system
type ConfigIntegration struct {
	configManager *config.ConfigManager
	ffmpegManager FFmpegManager
	logger        *logging.Logger
}

// VersionInfo represents build-time version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

// Build-time variables - these will be set via -ldflags during build
// These are package-level variables that will be injected by the build process
var (
	Version   = "1.0.0"   // Injected by -ldflags "-X github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx.Version=..."
	BuildDate = "unknown" // Injected by -ldflags "-X github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx.BuildDate=..."
	GitCommit = "unknown" // Injected by -ldflags "-X github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx.GitCommit=..."
)

// NewConfigIntegration creates a new configuration integration
func NewConfigIntegration(configManager *config.ConfigManager, ffmpegManager FFmpegManager, logger *logging.Logger) *ConfigIntegration {
	ci := &ConfigIntegration{
		configManager: configManager,
		ffmpegManager: ffmpegManager,
		logger:        logger,
	}

	// Log version information at startup
	versionInfo := ci.GetVersionInfo()
	logger.WithFields(logging.Fields{
		"version":    versionInfo.Version,
		"build_date": versionInfo.BuildDate,
		"git_commit": versionInfo.GitCommit,
	}).Info("ConfigIntegration initialized with version info")

	return ci
}

// GetMediaMTXConfig retrieves MediaMTX configuration from the existing config system
func (ci *ConfigIntegration) GetMediaMTXConfig() (*config.MediaMTXConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	// Convert existing config to MediaMTX config
	fmt.Printf("DEBUG: GetMediaMTXConfig - Source ControllerTickerInterval: %v\n", cfg.MediaMTX.StreamReadiness.ControllerTickerInterval)
	mediaMTXConfig := &config.MediaMTXConfig{
		// Core MediaMTX settings
		BaseURL:        fmt.Sprintf("http://%s:%d", cfg.MediaMTX.Host, cfg.MediaMTX.APIPort),
		HealthCheckURL: fmt.Sprintf("http://%s:%d%s", cfg.MediaMTX.Host, cfg.MediaMTX.APIPort, MediaMTXPathsList),
		Timeout:        time.Duration(cfg.MediaMTX.HealthCheckInterval) * time.Second,
		RetryAttempts:  3, // Default value
		RetryDelay:     time.Duration(cfg.MediaMTX.HealthMaxBackoffInterval) * time.Second,

		// Circuit breaker configuration
		CircuitBreaker: config.CircuitBreakerConfig{
			FailureThreshold: cfg.MediaMTX.HealthFailureThreshold,
			RecoveryTimeout:  time.Duration(cfg.MediaMTX.HealthCircuitBreakerTimeout) * time.Second,
			MaxFailures:      cfg.MediaMTX.HealthRecoveryConfirmationThreshold,
		},

		// Connection pool configuration
		ConnectionPool: config.ConnectionPoolConfig{
			MaxIdleConns:        100,              // Default value
			MaxIdleConnsPerHost: 10,               // Default value
			IdleConnTimeout:     90 * time.Second, // Default value
		},

		// Integration with existing config
		Host:                  cfg.MediaMTX.Host,
		APIPort:               cfg.MediaMTX.APIPort,
		RTSPPort:              cfg.MediaMTX.RTSPPort,
		WebRTCPort:            cfg.MediaMTX.WebRTCPort,
		HLSPort:               cfg.MediaMTX.HLSPort,
		ConfigPath:            cfg.MediaMTX.ConfigPath,
		RecordingsPath:        cfg.MediaMTX.RecordingsPath,
		SnapshotsPath:         cfg.MediaMTX.SnapshotsPath,
		OverrideMediaMTXPaths: cfg.MediaMTX.OverrideMediaMTXPaths,

		// Codec configuration
		Codec: cfg.MediaMTX.Codec,

		// FFmpeg and Performance Configuration
		FFmpeg:      cfg.MediaMTX.FFmpeg,
		Performance: cfg.MediaMTX.Performance,

		// Health and Circuit Breaker Configuration
		HealthCheckInterval:                 cfg.MediaMTX.HealthCheckInterval,
		HealthFailureThreshold:              cfg.MediaMTX.HealthFailureThreshold,
		HealthCircuitBreakerTimeout:         cfg.MediaMTX.HealthCircuitBreakerTimeout,
		HealthMaxBackoffInterval:            cfg.MediaMTX.HealthMaxBackoffInterval,
		HealthRecoveryConfirmationThreshold: cfg.MediaMTX.HealthRecoveryConfirmationThreshold,
		HealthCheckTimeout:                  cfg.MediaMTX.HealthCheckTimeout,
		BackoffBaseMultiplier:               cfg.MediaMTX.BackoffBaseMultiplier,
		BackoffJitterRange:                  cfg.MediaMTX.BackoffJitterRange,
		ProcessTerminationTimeout:           cfg.MediaMTX.ProcessTerminationTimeout,
		ProcessKillTimeout:                  cfg.MediaMTX.ProcessKillTimeout,

		// Stream Readiness Configuration
		StreamReadiness: cfg.MediaMTX.StreamReadiness,

		// Run on demand configuration
		RunOnDemandStartTimeout: cfg.MediaMTX.RunOnDemandStartTimeout,
		RunOnDemandCloseAfter:   cfg.MediaMTX.RunOnDemandCloseAfter,

		// Recording configuration
		RecordPartDuration:    cfg.MediaMTX.RecordPartDuration,
		RecordSegmentDuration: cfg.MediaMTX.RecordSegmentDuration,
		RecordDeleteAfter:     cfg.MediaMTX.RecordDeleteAfter,
	}

	ci.logger.WithFields(logging.Fields{
		"host":     mediaMTXConfig.Host,
		"api_port": mediaMTXConfig.APIPort,
		"base_url": mediaMTXConfig.BaseURL,
	}).Debug("MediaMTX configuration loaded from existing config system")

	return mediaMTXConfig, nil
}

// ValidateMediaMTXConfig validates MediaMTX configuration
func (ci *ConfigIntegration) ValidateMediaMTXConfig(mediaMTXConfig *config.MediaMTXConfig) error {
	if err := validateConfig(mediaMTXConfig); err != nil {
		return fmt.Errorf("MediaMTX config validation failed: %w", err)
	}

	// Additional validation specific to integration
	if mediaMTXConfig.Host == "" {
		return fmt.Errorf("MediaMTX host is required")
	}

	if mediaMTXConfig.APIPort <= 0 {
		return fmt.Errorf("MediaMTX API port must be positive")
	}

	if mediaMTXConfig.RecordingsPath == "" {
		return fmt.Errorf("MediaMTX recordings path is required")
	}

	if mediaMTXConfig.SnapshotsPath == "" {
		return fmt.Errorf("MediaMTX snapshots path is required")
	}

	ci.logger.Debug("MediaMTX configuration validation passed")
	return nil
}

// GetRecordingConfig retrieves recording configuration
func (ci *ConfigIntegration) GetRecordingConfig() (*config.RecordingConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.Recording, nil
}

// GetSnapshotConfig retrieves snapshot configuration
func (ci *ConfigIntegration) GetSnapshotConfig() (*config.SnapshotConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.Snapshots, nil
}

// GetFFmpegConfig retrieves FFmpeg configuration
func (ci *ConfigIntegration) GetFFmpegConfig() (*config.FFmpegConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.FFmpeg, nil
}

// GetCameraConfig retrieves camera configuration
func (ci *ConfigIntegration) GetCameraConfig() (*config.CameraConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.Camera, nil
}

// GetPerformanceConfig retrieves performance configuration
func (ci *ConfigIntegration) GetPerformanceConfig() (*config.PerformanceConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.Performance, nil
}

// GetExternalDiscoveryConfig retrieves external discovery configuration
func (ci *ConfigIntegration) GetExternalDiscoveryConfig() (*config.ExternalDiscoveryConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.ExternalDiscovery, nil
}

// GetConfig retrieves the full configuration
func (ci *ConfigIntegration) GetConfig() (*config.Config, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return cfg, nil
}

// UpdateMediaMTXConfig updates MediaMTX configuration in the existing config system
func (ci *ConfigIntegration) UpdateMediaMTXConfig(mediaMTXConfig *config.MediaMTXConfig) error {
	// Validate the new configuration
	if err := ci.ValidateMediaMTXConfig(mediaMTXConfig); err != nil {
		return fmt.Errorf("invalid MediaMTX configuration: %w", err)
	}

	// Get current config
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get current config: config is nil")
	}

	// Update MediaMTX configuration
	cfg.MediaMTX.Host = mediaMTXConfig.Host
	cfg.MediaMTX.APIPort = mediaMTXConfig.APIPort
	cfg.MediaMTX.RTSPPort = mediaMTXConfig.RTSPPort
	cfg.MediaMTX.WebRTCPort = mediaMTXConfig.WebRTCPort
	cfg.MediaMTX.HLSPort = mediaMTXConfig.HLSPort
	cfg.MediaMTX.ConfigPath = mediaMTXConfig.ConfigPath
	cfg.MediaMTX.RecordingsPath = mediaMTXConfig.RecordingsPath
	cfg.MediaMTX.SnapshotsPath = mediaMTXConfig.SnapshotsPath
	cfg.MediaMTX.HealthCheckInterval = mediaMTXConfig.HealthCheckInterval
	cfg.MediaMTX.HealthFailureThreshold = mediaMTXConfig.HealthFailureThreshold
	cfg.MediaMTX.HealthCircuitBreakerTimeout = mediaMTXConfig.HealthCircuitBreakerTimeout
	cfg.MediaMTX.HealthMaxBackoffInterval = mediaMTXConfig.HealthMaxBackoffInterval
	cfg.MediaMTX.HealthRecoveryConfirmationThreshold = mediaMTXConfig.HealthRecoveryConfirmationThreshold
	cfg.MediaMTX.BackoffBaseMultiplier = mediaMTXConfig.BackoffBaseMultiplier
	cfg.MediaMTX.BackoffJitterRange = mediaMTXConfig.BackoffJitterRange
	cfg.MediaMTX.ProcessTerminationTimeout = mediaMTXConfig.ProcessTerminationTimeout
	cfg.MediaMTX.ProcessKillTimeout = mediaMTXConfig.ProcessKillTimeout

	// Save configuration to file
	if err := ci.configManager.SaveConfig(); err != nil {
		ci.logger.WithError(err).Error("Failed to save configuration to file")
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	ci.logger.Info("MediaMTX configuration updated and saved to file")
	return nil
}

// GetVersionInfo returns build-time version information
func (ci *ConfigIntegration) GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   getVersionFromBuildOrEnv(),
		BuildDate: getBuildDateFromBuildOrEnv(),
		GitCommit: getGitCommitFromBuildOrEnv(),
	}
}

// Helper functions to access build-time variables with fallbacks
func getVersionFromBuildOrEnv() string {
	// Check environment variable first (for runtime override)
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	// Use build-time injected version
	return Version
}

func getBuildDateFromBuildOrEnv() string {
	if buildDate := os.Getenv("SERVICE_BUILD_DATE"); buildDate != "" {
		return buildDate
	}
	return BuildDate
}

func getGitCommitFromBuildOrEnv() string {
	if gitCommit := os.Getenv("SERVICE_GIT_COMMIT"); gitCommit != "" {
		return gitCommit
	}
	return GitCommit
}

// GetRetentionPolicy retrieves retention policy configuration
func (ci *ConfigIntegration) GetRetentionPolicy() (*config.RetentionPolicyConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	return &cfg.RetentionPolicy, nil
}

// GetCleanupLimits calculates cleanup limits based on configuration
func (ci *ConfigIntegration) GetCleanupLimits() (maxAge time.Duration, maxCount int, maxSize int64, err error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return 0, 0, 0, fmt.Errorf("failed to get config: config is nil")
	}

	// Use retention policy for age-based cleanup
	maxAge = time.Duration(cfg.RetentionPolicy.MaxAgeDays) * 24 * time.Hour

	// Use snapshot config for count-based cleanup (snapshots have count limits)
	maxCount = cfg.Snapshots.MaxCount

	// Use retention policy for size-based cleanup (convert GB to bytes)
	maxSize = int64(cfg.RetentionPolicy.MaxSizeGB) * 1024 * 1024 * 1024

	return maxAge, maxCount, maxSize, nil
}

// BuildSourceURL builds the appropriate source URL based on configuration and path type
func (ci *ConfigIntegration) BuildSourceURL(pathName string, pathSource *PathSource) (string, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return "", fmt.Errorf("failed to get config: config is nil")
	}

	// If path has a specific source, use it appropriately
	if pathSource != nil {
		if pathSource.Type == "rtspSource" {
			// For RTSP sources, use the provided ID as the source URL
			if strings.HasPrefix(pathSource.ID, "rtsp://") {
				return pathSource.ID, nil
			}
			// If ID is a device path, convert to device path
			return GetDevicePathFromCameraIdentifier(pathSource.ID), nil
		}
	}

	// Build RTSP URL using configuration
	return fmt.Sprintf("rtsp://%s:%d/%s", cfg.MediaMTX.Host, cfg.MediaMTX.RTSPPort, pathName), nil
}

// BuildPathConf creates a comprehensive PathConf for general path operations
// This variant accepts a formatResolver to enable dynamic pixel format selection
func (ci *ConfigIntegration) BuildPathConf(pathName string, pathSource *PathSource, enableRecording bool) (*PathConf, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	// Determine source URL and device path
	var devicePath string
	if pathSource != nil && pathSource.Type == "rtspSource" {
		devicePath = GetDevicePathFromCameraIdentifier(pathSource.ID)
	} else {
		devicePath = GetDevicePathFromCameraIdentifier(pathName)
	}

	pathConf := &PathConf{
		// Basic path configuration
		Name: pathName,

		// On-demand configuration
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: cfg.MediaMTX.RunOnDemandStartTimeout,
		SourceOnDemandCloseAfter:   cfg.MediaMTX.RunOnDemandCloseAfter,

		// FFmpeg command configuration using FFmpegManager
		RunOnDemand:        ci.buildPathCommand(devicePath, pathName),
		RunOnDemandRestart: true,
	}

	// Add recording configuration if requested
	if enableRecording {
		recordPath := GenerateRecordingPath(&cfg.MediaMTX, &cfg.Recording)
		pathConf.Record = cfg.Recording.Enabled
		pathConf.RecordFormat = cfg.Recording.RecordFormat
		pathConf.RecordPath = recordPath
		pathConf.RecordPartDuration = cfg.MediaMTX.RecordPartDuration
		pathConf.RecordSegmentDuration = cfg.MediaMTX.RecordSegmentDuration
		pathConf.RecordDeleteAfter = cfg.MediaMTX.RecordDeleteAfter
	}

	return pathConf, nil
}

// WatchConfigChanges watches for configuration changes and notifies the MediaMTX controller
func (ci *ConfigIntegration) WatchConfigChanges(controller MediaMTXController) error {
	// Note: SubscribeToChanges method doesn't exist in ConfigManager
	// Configuration watching would need to be implemented through the existing config system
	ci.logger.Debug("Configuration change watcher not implemented (requires ConfigManager enhancement)")
	return nil
}

// buildPathCommand builds FFmpeg command using injected FFmpegManager with fallback
func (ci *ConfigIntegration) buildPathCommand(devicePath, pathName string) string {
	runOnDemand, err := ci.ffmpegManager.BuildRunOnDemandCommand(devicePath, pathName)
	if err != nil {
		// Return a basic fallback command if FFmpegManager fails
		cfg := ci.configManager.GetConfig()
		return fmt.Sprintf("ffmpeg -f v4l2 -i %s -c:v libx264 -preset %s -f rtsp rtsp://%s:%d/%s",
			devicePath, cfg.MediaMTX.Codec.Preset, cfg.MediaMTX.Host, cfg.MediaMTX.RTSPPort, pathName)
	}
	return runOnDemand
}
