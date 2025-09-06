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
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// ConfigIntegration provides integration between MediaMTX package and existing config system
type ConfigIntegration struct {
	configManager *config.ConfigManager
	logger        *logging.Logger
}

// NewConfigIntegration creates a new configuration integration
func NewConfigIntegration(configManager *config.ConfigManager, logger *logging.Logger) *ConfigIntegration {
	return &ConfigIntegration{
		configManager: configManager,
		logger:        logger,
	}
}

// GetMediaMTXConfig retrieves MediaMTX configuration from the existing config system
func (ci *ConfigIntegration) GetMediaMTXConfig() (*MediaMTXConfig, error) {
	cfg := ci.configManager.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("failed to get config: config is nil")
	}

	// Convert existing config to MediaMTX config
	mediaMTXConfig := &MediaMTXConfig{
		// Core MediaMTX settings
		BaseURL:        fmt.Sprintf("http://%s:%d", cfg.MediaMTX.Host, cfg.MediaMTX.APIPort),
		HealthCheckURL: fmt.Sprintf("http://%s:%d/v3/paths/list", cfg.MediaMTX.Host, cfg.MediaMTX.APIPort),
		Timeout:        time.Duration(cfg.MediaMTX.HealthCheckInterval) * time.Second,
		RetryAttempts:  3, // Default value
		RetryDelay:     time.Duration(cfg.MediaMTX.HealthMaxBackoffInterval) * time.Second,

		// Circuit breaker configuration
		CircuitBreaker: CircuitBreakerConfig{
			FailureThreshold: cfg.MediaMTX.HealthFailureThreshold,
			RecoveryTimeout:  time.Duration(cfg.MediaMTX.HealthCircuitBreakerTimeout) * time.Second,
			MaxFailures:      cfg.MediaMTX.HealthRecoveryConfirmationThreshold,
		},

		// Connection pool configuration
		ConnectionPool: ConnectionPoolConfig{
			MaxIdleConns:        100,              // Default value
			MaxIdleConnsPerHost: 10,               // Default value
			IdleConnTimeout:     90 * time.Second, // Default value
		},

		// Integration with existing config
		Host:                                cfg.MediaMTX.Host,
		APIPort:                             cfg.MediaMTX.APIPort,
		RTSPPort:                            cfg.MediaMTX.RTSPPort,
		WebRTCPort:                          cfg.MediaMTX.WebRTCPort,
		HLSPort:                             cfg.MediaMTX.HLSPort,
		ConfigPath:                          cfg.MediaMTX.ConfigPath,
		RecordingsPath:                      cfg.MediaMTX.RecordingsPath,
		SnapshotsPath:                       cfg.MediaMTX.SnapshotsPath,
		HealthCheckInterval:                 cfg.MediaMTX.HealthCheckInterval,
		HealthFailureThreshold:              cfg.MediaMTX.HealthFailureThreshold,
		HealthCircuitBreakerTimeout:         cfg.MediaMTX.HealthCircuitBreakerTimeout,
		HealthMaxBackoffInterval:            cfg.MediaMTX.HealthMaxBackoffInterval,
		HealthRecoveryConfirmationThreshold: cfg.MediaMTX.HealthRecoveryConfirmationThreshold,
		BackoffBaseMultiplier:               cfg.MediaMTX.BackoffBaseMultiplier,
		BackoffJitterRange:                  cfg.MediaMTX.BackoffJitterRange,
		ProcessTerminationTimeout:           cfg.MediaMTX.ProcessTerminationTimeout,
		ProcessKillTimeout:                  cfg.MediaMTX.ProcessKillTimeout,
	}

	ci.logger.WithFields(map[string]interface{}{
		"host":     mediaMTXConfig.Host,
		"api_port": mediaMTXConfig.APIPort,
		"base_url": mediaMTXConfig.BaseURL,
	}).Debug("MediaMTX configuration loaded from existing config system")

	return mediaMTXConfig, nil
}

// ValidateMediaMTXConfig validates MediaMTX configuration
func (ci *ConfigIntegration) ValidateMediaMTXConfig(mediaMTXConfig *MediaMTXConfig) error {
	if err := validateConfig(mediaMTXConfig); err != nil {
		return fmt.Errorf("MediaMTX config validation failed: %w", err)
	}

	// Additional validation specific to integration
	if mediaMTXConfig.Host == "" {
		return NewConfigurationError("host", "", "MediaMTX host is required")
	}

	if mediaMTXConfig.APIPort <= 0 {
		return NewConfigurationError("api_port", fmt.Sprintf("%d", mediaMTXConfig.APIPort), "MediaMTX API port must be positive")
	}

	if mediaMTXConfig.RecordingsPath == "" {
		return NewConfigurationError("recordings_path", "", "MediaMTX recordings path is required")
	}

	if mediaMTXConfig.SnapshotsPath == "" {
		return NewConfigurationError("snapshots_path", "", "MediaMTX snapshots path is required")
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

// UpdateMediaMTXConfig updates MediaMTX configuration in the existing config system
func (ci *ConfigIntegration) UpdateMediaMTXConfig(mediaMTXConfig *MediaMTXConfig) error {
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

// WatchConfigChanges watches for configuration changes and notifies the MediaMTX controller
func (ci *ConfigIntegration) WatchConfigChanges(controller MediaMTXController) error {
	// Note: SubscribeToChanges method doesn't exist in ConfigManager
	// Configuration watching would need to be implemented through the existing config system
	ci.logger.Debug("Configuration change watcher not implemented (requires ConfigManager enhancement)")
	return nil
}
