package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigManager manages configuration loading, validation, and hot reload functionality.
type ConfigManager struct {
	config          *Config
	configPath      string
	updateCallbacks []func(*Config)
	watcher         *fsnotify.Watcher
	watcherActive   int32 // Atomic: 0 = inactive, 1 = active
	watcherLock     sync.RWMutex
	lock            sync.RWMutex
	defaultConfig   *Config
	logger          *logging.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// CreateConfigManager creates a new configuration manager instance.
func CreateConfigManager() *ConfigManager {
	return &ConfigManager{
		updateCallbacks: make([]func(*Config), 0),
		defaultConfig:   getDefaultConfig(),
		logger:          logging.GetLogger("config-manager"),
		stopChan:        make(chan struct{}, 5), // Buffered to prevent deadlock during shutdown
	}
}

// LoadConfig loads configuration from YAML file with environment variable overrides and validation.
func (cm *ConfigManager) LoadConfig(configPath string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.logger.WithFields(logging.Fields{
		"config_path": configPath,
		"action":      "load_config",
	}).Info("Loading configuration")

	// REQ-CONFIG-001: Validate configuration files before loading
	// REQ-CONFIG-002: Fail fast on configuration errors
	// REQ-CONFIG-003: Early detection and clear error reporting
	if err := cm.validateConfigFile(configPath); err != nil {
		return fmt.Errorf("configuration validation failed: invalid configuration - %w", err)
	}

	// Set up Viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Set default values
	cm.setDefaults(v)

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("CAMERA_SERVICE")

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: invalid configuration - cannot read configuration file '%s': %w", configPath, err)
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// CRITICAL: Apply defaults after unmarshaling to prevent zero values from overriding defaults
	// This fixes the bug where incomplete YAML sections cause zero values to override Viper defaults
	cm.applyDefaultsAfterUnmarshal(&config)

	// REQ-CONFIG-001: Validate final configuration values after environment variable overrides
	// REQ-CONFIG-002: Fail fast on configuration errors
	// REQ-CONFIG-003: Early detection and clear error reporting
	if err := cm.validateFinalConfiguration(&config); err != nil {
		return fmt.Errorf("configuration validation failed: invalid configuration - %w", err)
	}

	// Validate configuration
	if err := ValidateConfig(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Store configuration
	oldConfig := cm.config
	cm.config = &config
	cm.configPath = configPath

	// Start file watching for hot reload (only if explicitly enabled)
	// Hot reload is disabled by default for tests and can be enabled via environment variable
	if os.Getenv("CAMERA_SERVICE_ENABLE_HOT_RELOAD") == "true" {
		if err := cm.startFileWatching(); err != nil {
			cm.logger.WithError(err).Warn("Failed to start file watching, hot reload disabled")
		}
	}

	// Notify callbacks
	cm.notifyConfigUpdated(oldConfig, &config)

	cm.logger.WithFields(logging.Fields{
		"config_path": configPath,
		"action":      "load_config",
		"status":      "success",
	}).Info("Configuration loaded successfully")

	return nil
}

// validateConfigFile validates the configuration file before loading
// REQ-CONFIG-001: The system SHALL validate configuration files before loading
// REQ-CONFIG-002: The system SHALL fail fast on configuration errors
// REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting
func (cm *ConfigManager) validateConfigFile(configPath string) error {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: '%s'", configPath)
	}

	// Read file content for validation
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("cannot read configuration file '%s': %w", configPath, err)
	}

	// Check for empty file
	if len(content) == 0 {
		return fmt.Errorf("configuration file is empty: '%s' - file must contain valid YAML configuration", configPath)
	}

	// Check for comments-only file
	lines := strings.Split(string(content), "\n")
	hasNonCommentContent := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip empty lines
		}
		if strings.HasPrefix(trimmedLine, "#") {
			continue // Skip comment lines
		}
		hasNonCommentContent = true
		break
	}

	if !hasNonCommentContent {
		return fmt.Errorf("configuration file contains only comments or is empty: '%s' - file must contain valid YAML configuration data", configPath)
	}

	return nil
}

// validateFinalConfiguration validates the final configuration values after environment variable overrides
// REQ-CONFIG-001: The system SHALL validate configuration files before loading
// REQ-CONFIG-002: The system SHALL fail fast on configuration errors
// REQ-CONFIG-003: Edge case handling SHALL mean early detection and clear error reporting
func (cm *ConfigManager) validateFinalConfiguration(config *Config) error {
	// Validate server configuration
	if strings.TrimSpace(config.Server.Host) == "" {
		return fmt.Errorf("server host cannot be empty or whitespace-only")
	}
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", config.Server.Port)
	}

	// Validate MediaMTX configuration
	if strings.TrimSpace(config.MediaMTX.Host) == "" {
		return fmt.Errorf("MediaMTX host cannot be empty or whitespace-only")
	}
	if config.MediaMTX.APIPort <= 0 || config.MediaMTX.APIPort > 65535 {
		return fmt.Errorf("MediaMTX API port must be between 1 and 65535, got %d", config.MediaMTX.APIPort)
	}
	if config.MediaMTX.RTSPPort <= 0 || config.MediaMTX.RTSPPort > 65535 {
		return fmt.Errorf("MediaMTX RTSP port must be between 1 and 65535, got %d", config.MediaMTX.RTSPPort)
	}
	if config.MediaMTX.WebRTCPort <= 0 || config.MediaMTX.WebRTCPort > 65535 {
		return fmt.Errorf("MediaMTX WebRTC port must be between 1 and 65535, got %d", config.MediaMTX.WebRTCPort)
	}
	if config.MediaMTX.HLSPort <= 0 || config.MediaMTX.HLSPort > 65535 {
		return fmt.Errorf("MediaMTX HLS port must be between 1 and 65535, got %d", config.MediaMTX.HLSPort)
	}
	if strings.TrimSpace(config.MediaMTX.ConfigPath) == "" {
		return fmt.Errorf("MediaMTX config path cannot be empty or whitespace-only")
	}
	if strings.TrimSpace(config.MediaMTX.RecordingsPath) == "" {
		return fmt.Errorf("MediaMTX recordings path cannot be empty or whitespace-only")
	}
	if strings.TrimSpace(config.MediaMTX.SnapshotsPath) == "" {
		return fmt.Errorf("MediaMTX snapshots path cannot be empty or whitespace-only")
	}

	// Validate camera configuration
	if config.Camera.PollInterval <= 0 {
		return fmt.Errorf("camera poll interval must be positive, got %f", config.Camera.PollInterval)
	}
	if config.Camera.DetectionTimeout <= 0 {
		return fmt.Errorf("camera detection timeout must be positive, got %f", config.Camera.DetectionTimeout)
	}
	if config.Camera.CapabilityTimeout <= 0 {
		return fmt.Errorf("camera capability timeout must be positive, got %f", config.Camera.CapabilityTimeout)
	}
	if config.Camera.CapabilityRetryInterval <= 0 {
		return fmt.Errorf("camera capability retry interval must be positive, got %f", config.Camera.CapabilityRetryInterval)
	}
	if config.Camera.CapabilityMaxRetries < 0 {
		return fmt.Errorf("camera capability max retries cannot be negative, got %d", config.Camera.CapabilityMaxRetries)
	}

	// Validate logging configuration
	validLogLevels := []string{"debug", "info", "warn", "warning", "error", "fatal", "panic"}
	levelFound := false
	for _, valid := range validLogLevels {
		if strings.ToLower(config.Logging.Level) == valid {
			levelFound = true
			break
		}
	}
	if !levelFound {
		return fmt.Errorf("logging level must be one of: %v, got %s", validLogLevels, config.Logging.Level)
	}
	if strings.TrimSpace(config.Logging.Format) == "" {
		return fmt.Errorf("logging format cannot be empty or whitespace-only")
	}
	if config.Logging.FileEnabled && strings.TrimSpace(config.Logging.FilePath) == "" {
		return fmt.Errorf("logging file path cannot be empty when file logging is enabled")
	}

	// Validate recording configuration
	if strings.TrimSpace(config.Recording.Format) == "" {
		return fmt.Errorf("recording format cannot be empty or whitespace-only")
	}
	if strings.TrimSpace(config.Recording.Quality) == "" {
		return fmt.Errorf("recording quality cannot be empty or whitespace-only")
	}
	if config.Recording.SegmentDuration < 0 {
		return fmt.Errorf("recording segment duration cannot be negative, got %d", config.Recording.SegmentDuration)
	}
	if config.Recording.MaxSegmentSize < 0 {
		return fmt.Errorf("recording max segment size cannot be negative, got %d", config.Recording.MaxSegmentSize)
	}
	if config.Recording.CleanupInterval < 0 {
		return fmt.Errorf("recording cleanup interval cannot be negative, got %d", config.Recording.CleanupInterval)
	}
	if config.Recording.MaxAge < 0 {
		return fmt.Errorf("recording max age cannot be negative, got %d", config.Recording.MaxAge)
	}
	if config.Recording.MaxSize <= 0 {
		return fmt.Errorf("recording max size must be positive, got %d", config.Recording.MaxSize)
	}
	if config.Recording.DefaultRotationSize < 0 {
		return fmt.Errorf("recording default rotation size cannot be negative, got %d", config.Recording.DefaultRotationSize)
	}
	if config.Recording.DefaultMaxDuration < 0 {
		return fmt.Errorf("recording default max duration cannot be negative, got %v", config.Recording.DefaultMaxDuration)
	}
	if config.Recording.DefaultRetentionDays < 0 {
		return fmt.Errorf("recording default retention days cannot be negative, got %d", config.Recording.DefaultRetentionDays)
	}

	// Validate snapshots configuration
	if strings.TrimSpace(config.Snapshots.Format) == "" {
		return fmt.Errorf("snapshots format cannot be empty or whitespace-only")
	}
	if config.Snapshots.Quality < 0 || config.Snapshots.Quality > 100 {
		return fmt.Errorf("snapshots quality must be between 0 and 100, got %d", config.Snapshots.Quality)
	}
	if config.Snapshots.MaxWidth <= 0 {
		return fmt.Errorf("snapshots max width must be positive, got %d", config.Snapshots.MaxWidth)
	}
	if config.Snapshots.MaxHeight <= 0 {
		return fmt.Errorf("snapshots max height must be positive, got %d", config.Snapshots.MaxHeight)
	}
	if config.Snapshots.CleanupInterval < 0 {
		return fmt.Errorf("snapshots cleanup interval cannot be negative, got %d", config.Snapshots.CleanupInterval)
	}
	if config.Snapshots.MaxAge < 0 {
		return fmt.Errorf("snapshots max age cannot be negative, got %d", config.Snapshots.MaxAge)
	}
	if config.Snapshots.MaxCount <= 0 {
		return fmt.Errorf("snapshots max count must be positive, got %d", config.Snapshots.MaxCount)
	}

	// Validate storage configuration
	if config.Storage.WarnPercent < 0 || config.Storage.WarnPercent > 100 {
		return fmt.Errorf("storage warn percent must be between 0 and 100, got %d", config.Storage.WarnPercent)
	}
	if config.Storage.BlockPercent < 0 || config.Storage.BlockPercent > 100 {
		return fmt.Errorf("storage block percent must be between 0 and 100, got %d", config.Storage.BlockPercent)
	}
	if config.Storage.WarnPercent >= config.Storage.BlockPercent {
		return fmt.Errorf("storage warn percent (%d) must be less than block percent (%d)", config.Storage.WarnPercent, config.Storage.BlockPercent)
	}
	if strings.TrimSpace(config.Storage.DefaultPath) == "" {
		return fmt.Errorf("storage default path cannot be empty or whitespace-only")
	}
	if strings.TrimSpace(config.Storage.FallbackPath) == "" {
		return fmt.Errorf("storage fallback path cannot be empty or whitespace-only")
	}

	return nil
}

// startFileWatching starts watching the configuration file for changes.
func (cm *ConfigManager) startFileWatching() error {
	// Stop existing watcher if any
	cm.stopFileWatching()

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	cm.watcherLock.Lock()
	cm.watcher = watcher
	cm.watcherLock.Unlock()

	// Watch the directory containing the config file
	configDir := filepath.Dir(cm.configPath)
	if err := cm.watcher.Add(configDir); err != nil {
		cm.watcher.Close()
		cm.watcherLock.Lock()
		cm.watcher = nil
		cm.watcherLock.Unlock()
		return fmt.Errorf("failed to watch config directory %s: %w", configDir, err)
	}

	// Mark watcher as active atomically
	atomic.StoreInt32(&cm.watcherActive, 1)

	// Start watching goroutine
	cm.wg.Add(1)
	go cm.watchFileChanges()

	cm.logger.WithFields(logging.Fields{
		"config_path": cm.configPath,
		"watch_dir":   configDir,
	}).Info("File watching started for hot reload")

	return nil
}

// stopFileWatching stops the file watcher.
func (cm *ConfigManager) stopFileWatching() {
	// Mark watcher as inactive atomically
	atomic.StoreInt32(&cm.watcherActive, 0)

	cm.watcherLock.Lock()
	defer cm.watcherLock.Unlock()

	if cm.watcher != nil {
		// Close watcher safely
		if err := cm.watcher.Close(); err != nil {
			cm.logger.WithError(err).Warn("Error closing file watcher")
		}
		cm.watcher = nil
		cm.logger.Debug("File watcher stopped and cleaned up")
	}
}

// watchFileChanges watches for file changes and triggers configuration reload.
func (cm *ConfigManager) watchFileChanges() {
	defer cm.wg.Done()

	// Debounce timer to avoid multiple reloads
	var reloadTimer *time.Timer

	for {
		select {
		case <-cm.stopChan:
			return
		default:
			// Check if watcher is still active atomically
			if atomic.LoadInt32(&cm.watcherActive) == 0 {
				return
			}

			// Check if watcher is still valid before accessing its channels
			cm.watcherLock.RLock()
			if cm.watcher == nil {
				cm.watcherLock.RUnlock()
				return
			}
			events := cm.watcher.Events
			errors := cm.watcher.Errors
			cm.watcherLock.RUnlock()

			// Use a select with timeout to avoid blocking indefinitely
			select {
			case <-cm.stopChan:
				return
			case event, ok := <-events:
				if !ok {
					return
				}

				// Check if the changed file is our config file
				if event.Name == cm.configPath {
					cm.logger.WithFields(logging.Fields{
						"file":  event.Name,
						"event": event.Op.String(),
					}).Debug("Configuration file change detected")

					// Handle different event types
					switch event.Op {
					case fsnotify.Write, fsnotify.Create:
						// Debounce reload to avoid multiple rapid reloads
						if reloadTimer != nil {
							reloadTimer.Stop()
						}
						reloadTimer = time.AfterFunc(100*time.Millisecond, func() {
							cm.reloadConfiguration()
						})
					case fsnotify.Remove:
						cm.logger.Warn("Configuration file was removed, hot reload disabled")
						cm.stopFileWatching()
						return // Exit the goroutine when file is removed
					}
				}

			case err, ok := <-errors:
				if !ok {
					return
				}
				cm.logger.WithError(err).Error("File watcher error")
			case <-time.After(100 * time.Millisecond):
				// Continue loop to check stopChan and watcher validity
				continue
			}
		}
	}
}

// reloadConfiguration reloads the configuration file.
func (cm *ConfigManager) reloadConfiguration() {
	cm.logger.Info("Reloading configuration due to file change")

	// Check if file still exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		cm.logger.Warn("Configuration file no longer exists, stopping hot reload")
		cm.stopFileWatching()
		return
	}

	// Reload configuration
	if err := cm.LoadConfig(cm.configPath); err != nil {
		cm.logger.WithError(err).Error("Failed to reload configuration")
		return
	}

	cm.logger.Info("Configuration reloaded successfully")
}

// Stop stops the configuration manager and cleans up resources with context-aware cancellation.
func (cm *ConfigManager) Stop(ctx context.Context) error {
	cm.logger.Info("Stopping configuration manager")

	// Signal stop
	select {
	case <-cm.stopChan:
		// Already closed
	default:
		close(cm.stopChan)
	}

	// Stop file watching
	cm.stopFileWatching()

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		cm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		cm.logger.Warn("Configuration manager shutdown timeout")
		return ctx.Err()
	}

	cm.logger.Info("Configuration manager stopped")
	return nil
}

// GetConfig returns the current configuration.
func (cm *ConfigManager) GetConfig() *Config {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	if cm.config == nil {
		return cm.defaultConfig
	}
	return cm.config
}

// GetLogger returns the config manager's logger for level configuration.
func (cm *ConfigManager) GetLogger() *logging.Logger {
	return cm.logger
}

// SaveConfig saves the current configuration to the configuration file.
func (cm *ConfigManager) SaveConfig() error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if cm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	if cm.configPath == "" {
		return fmt.Errorf("no configuration file path set")
	}

	cm.logger.WithFields(logging.Fields{
		"config_path": cm.configPath,
		"action":      "save_config",
	}).Info("Saving configuration to file")

	// Create a new Viper instance for saving
	v := viper.New()
	v.SetConfigFile(cm.configPath)
	v.SetConfigType("yaml")

	// Set all configuration values
	cm.setConfigValues(v, cm.config)

	// Ensure the directory exists
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the configuration to file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.WithFields(logging.Fields{
		"config_path": cm.configPath,
		"action":      "save_config",
		"status":      "success",
	}).Info("Configuration saved successfully")

	return nil
}

// setConfigValues recursively sets configuration values in Viper
func (cm *ConfigManager) setConfigValues(v *viper.Viper, config *Config) {
	// Server configuration
	v.Set("server.host", config.Server.Host)
	v.Set("server.port", config.Server.Port)
	v.Set("server.websocket_path", config.Server.WebSocketPath)
	v.Set("server.max_connections", config.Server.MaxConnections)
	v.Set("server.read_timeout", config.Server.ReadTimeout)
	v.Set("server.write_timeout", config.Server.WriteTimeout)
	v.Set("server.ping_interval", config.Server.PingInterval)
	v.Set("server.pong_wait", config.Server.PongWait)
	v.Set("server.max_message_size", config.Server.MaxMessageSize)

	// MediaMTX configuration
	v.Set("mediamtx.host", config.MediaMTX.Host)
	v.Set("mediamtx.api_port", config.MediaMTX.APIPort)
	v.Set("mediamtx.rtsp_port", config.MediaMTX.RTSPPort)
	v.Set("mediamtx.webrtc_port", config.MediaMTX.WebRTCPort)
	v.Set("mediamtx.hls_port", config.MediaMTX.HLSPort)
	v.Set("mediamtx.config_path", config.MediaMTX.ConfigPath)
	v.Set("mediamtx.recordings_path", config.MediaMTX.RecordingsPath)
	v.Set("mediamtx.snapshots_path", config.MediaMTX.SnapshotsPath)
	v.Set("mediamtx.health_check_interval", config.MediaMTX.HealthCheckInterval)
	v.Set("mediamtx.health_failure_threshold", config.MediaMTX.HealthFailureThreshold)
	v.Set("mediamtx.health_circuit_breaker_timeout", config.MediaMTX.HealthCircuitBreakerTimeout)
	v.Set("mediamtx.health_max_backoff_interval", config.MediaMTX.HealthMaxBackoffInterval)
	v.Set("mediamtx.health_recovery_confirmation_threshold", config.MediaMTX.HealthRecoveryConfirmationThreshold)
	v.Set("mediamtx.backoff_base_multiplier", config.MediaMTX.BackoffBaseMultiplier)
	v.Set("mediamtx.backoff_jitter_range", config.MediaMTX.BackoffJitterRange)
	v.Set("mediamtx.process_termination_timeout", config.MediaMTX.ProcessTerminationTimeout)
	v.Set("mediamtx.process_kill_timeout", config.MediaMTX.ProcessKillTimeout)
	v.Set("mediamtx.health_check_timeout", config.MediaMTX.HealthCheckTimeout)

	// Camera configuration
	v.Set("camera.poll_interval", config.Camera.PollInterval)
	v.Set("camera.detection_timeout", config.Camera.DetectionTimeout)
	v.Set("camera.device_range", config.Camera.DeviceRange)
	v.Set("camera.enable_capability_detection", config.Camera.EnableCapabilityDetection)
	v.Set("camera.auto_start_streams", config.Camera.AutoStartStreams)
	v.Set("camera.capability_timeout", config.Camera.CapabilityTimeout)
	v.Set("camera.capability_retry_interval", config.Camera.CapabilityRetryInterval)
	v.Set("camera.capability_max_retries", config.Camera.CapabilityMaxRetries)

	// Logging configuration
	v.Set("logging.level", config.Logging.Level)
	v.Set("logging.format", config.Logging.Format)
	v.Set("logging.file_enabled", config.Logging.FileEnabled)
	v.Set("logging.file_path", config.Logging.FilePath)
	v.Set("logging.max_file_size", config.Logging.MaxFileSize)
	v.Set("logging.backup_count", config.Logging.BackupCount)
	v.Set("logging.console_enabled", config.Logging.ConsoleEnabled)

	// Recording configuration
	v.Set("recording.enabled", config.Recording.Enabled)
	v.Set("recording.format", config.Recording.Format)
	v.Set("recording.quality", config.Recording.Quality)
	v.Set("recording.segment_duration", config.Recording.SegmentDuration)
	v.Set("recording.max_segment_size", config.Recording.MaxSegmentSize)
	v.Set("recording.auto_cleanup", config.Recording.AutoCleanup)
	v.Set("recording.cleanup_interval", config.Recording.CleanupInterval)
	v.Set("recording.max_age", config.Recording.MaxAge)
	v.Set("recording.max_size", config.Recording.MaxSize)

	// Snapshots configuration
	v.Set("snapshots.enabled", config.Snapshots.Enabled)
	v.Set("snapshots.format", config.Snapshots.Format)
	v.Set("snapshots.quality", config.Snapshots.Quality)
	v.Set("snapshots.max_width", config.Snapshots.MaxWidth)
	v.Set("snapshots.max_height", config.Snapshots.MaxHeight)
	v.Set("snapshots.auto_cleanup", config.Snapshots.AutoCleanup)
	v.Set("snapshots.cleanup_interval", config.Snapshots.CleanupInterval)
	v.Set("snapshots.max_age", config.Snapshots.MaxAge)
	v.Set("snapshots.max_count", config.Snapshots.MaxCount)

	// Retention policy configuration
	v.Set("retention_policy.enabled", config.RetentionPolicy.Enabled)
	v.Set("retention_policy.type", config.RetentionPolicy.Type)
	v.Set("retention_policy.max_age_days", config.RetentionPolicy.MaxAgeDays)
	v.Set("retention_policy.max_size_gb", config.RetentionPolicy.MaxSizeGB)
	v.Set("retention_policy.auto_cleanup", config.RetentionPolicy.AutoCleanup)
}

// AddUpdateCallback adds a callback function to be called when configuration is updated.
func (cm *ConfigManager) AddUpdateCallback(callback func(*Config)) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.updateCallbacks = append(cm.updateCallbacks, callback)
}

// RegisterLoggingConfigurationUpdates registers automatic logging configuration updates.
// This method sets up a callback that automatically updates the global logging configuration
// whenever the main configuration is reloaded, ensuring all loggers use the latest settings.
//
// This solution:
// - ✅ No circular dependencies (config calls logging, never reverse)
// - ✅ Production ready (uses existing config reload infrastructure)
// - ✅ Power friendly (no polling, uses existing file watching)
// - ✅ Leverages factory pattern (updates all loggers via ConfigureGlobalLogging)
// - ✅ Integrates smoothly (uses your existing callback architecture)
func (cm *ConfigManager) RegisterLoggingConfigurationUpdates() {
	cm.AddUpdateCallback(func(newConfig *Config) {
		if newConfig == nil {
			cm.logger.Warn("Skipping logging config update - invalid configuration")
			return
		}

		// Convert main config logging section to logging.LoggingConfig
		// Uses the same conversion pattern as main.go
		loggingConfig := &logging.LoggingConfig{
			Level:          newConfig.Logging.Level,
			Format:         newConfig.Logging.Format,
			FileEnabled:    newConfig.Logging.FileEnabled,
			FilePath:       newConfig.Logging.FilePath,
			MaxFileSize:    int(newConfig.Logging.MaxFileSize),
			BackupCount:    newConfig.Logging.BackupCount,
			ConsoleEnabled: newConfig.Logging.ConsoleEnabled,
		}

		// Update global logging configuration and all loggers via factory
		if err := logging.ConfigureGlobalLogging(loggingConfig); err != nil {
			cm.logger.WithError(err).Error("Failed to update logging configuration")
			return
		}

		cm.logger.WithFields(logging.Fields{
			"level":           loggingConfig.Level,
			"format":          loggingConfig.Format,
			"file_enabled":    loggingConfig.FileEnabled,
			"console_enabled": loggingConfig.ConsoleEnabled,
		}).Info("Logging configuration updated successfully")
	})
}

// setDefaults sets default configuration values in Viper.
func (cm *ConfigManager) setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8002)
	v.SetDefault("server.websocket_path", "/ws")
	v.SetDefault("server.max_connections", 100)
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "1s")
	v.SetDefault("server.ping_interval", "30s")
	v.SetDefault("server.pong_wait", "60s")
	v.SetDefault("server.max_message_size", 1048576) // 1MB
	v.SetDefault("server.read_buffer_size", 1024)
	v.SetDefault("server.write_buffer_size", 1024)
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.client_cleanup_timeout", "10s")

	// MediaMTX defaults
	v.SetDefault("mediamtx.host", "127.0.0.1")
	v.SetDefault("mediamtx.api_port", 9997)
	v.SetDefault("mediamtx.rtsp_port", 8554)
	v.SetDefault("mediamtx.webrtc_port", 8889)
	v.SetDefault("mediamtx.hls_port", 8888)
	v.SetDefault("mediamtx.timeout", "10s")
	v.SetDefault("mediamtx.config_path", "/opt/camera-service/config/mediamtx.yml")
	v.SetDefault("mediamtx.recordings_path", "/opt/camera-service/recordings")
	v.SetDefault("mediamtx.snapshots_path", "/opt/camera-service/snapshots")

	// MediaMTX codec defaults - STANAG 4609 compliant
	v.SetDefault("mediamtx.codec.video_profile", "high422") // STANAG 4609 compliant with 4:2:2 support
	v.SetDefault("mediamtx.codec.video_level", "4.0")       // Tactical system standard
	v.SetDefault("mediamtx.codec.pixel_format", "yuv422p")  // 4:2:2 for tactical systems
	v.SetDefault("mediamtx.codec.bitrate", "2M")            // Increased for tactical quality
	v.SetDefault("mediamtx.codec.preset", "fast")           // Better quality for tactical use

	// MediaMTX health monitoring defaults
	v.SetDefault("mediamtx.health_check_interval", 30)
	v.SetDefault("mediamtx.health_failure_threshold", 10)
	v.SetDefault("mediamtx.health_circuit_breaker_timeout", 60)
	v.SetDefault("mediamtx.health_max_backoff_interval", 120)
	v.SetDefault("mediamtx.health_recovery_confirmation_threshold", 3)
	v.SetDefault("mediamtx.health_check_timeout", "5s")
	v.SetDefault("mediamtx.backoff_base_multiplier", 2.0)
	v.SetDefault("mediamtx.backoff_jitter_range", []float64{0.8, 1.2})
	v.SetDefault("mediamtx.process_termination_timeout", 3.0)
	v.SetDefault("mediamtx.process_kill_timeout", 2.0)

	// RTSP Connection Monitoring defaults
	v.SetDefault("mediamtx.rtsp_monitoring.enabled", true)
	v.SetDefault("mediamtx.rtsp_monitoring.check_interval", 30)
	v.SetDefault("mediamtx.rtsp_monitoring.connection_timeout", 10)
	v.SetDefault("mediamtx.rtsp_monitoring.max_connections", 50)
	v.SetDefault("mediamtx.rtsp_monitoring.session_timeout", 300)
	v.SetDefault("mediamtx.rtsp_monitoring.bandwidth_threshold", 1000000)
	v.SetDefault("mediamtx.rtsp_monitoring.packet_loss_threshold", 0.05)
	v.SetDefault("mediamtx.rtsp_monitoring.jitter_threshold", 50.0)

	// External Stream Discovery defaults
	v.SetDefault("mediamtx.external_discovery.enabled", true)
	v.SetDefault("mediamtx.external_discovery.scan_interval", 0)        // On-demand only
	v.SetDefault("mediamtx.external_discovery.scan_timeout", 30)        // 30 seconds max
	v.SetDefault("mediamtx.external_discovery.max_concurrent_scans", 5) // Limit concurrency
	v.SetDefault("mediamtx.external_discovery.enable_startup_scan", true)

	// Skydio-specific defaults (validated from official docs)
	v.SetDefault("mediamtx.external_discovery.skydio.enabled", true)
	v.SetDefault("mediamtx.external_discovery.skydio.network_ranges", []string{"192.168.42.0/24"})
	v.SetDefault("mediamtx.external_discovery.skydio.eo_port", 5554)
	v.SetDefault("mediamtx.external_discovery.skydio.ir_port", 6554)
	v.SetDefault("mediamtx.external_discovery.skydio.eo_stream_path", "/subject")
	v.SetDefault("mediamtx.external_discovery.skydio.ir_stream_path", "/infrared")
	v.SetDefault("mediamtx.external_discovery.skydio.enable_both_streams", true)
	v.SetDefault("mediamtx.external_discovery.skydio.known_ips", []string{"192.168.42.10"})

	// Generic UAV defaults (for other models)
	v.SetDefault("mediamtx.external_discovery.generic_uav.enabled", false)
	v.SetDefault("mediamtx.external_discovery.generic_uav.network_ranges", []string{})
	v.SetDefault("mediamtx.external_discovery.generic_uav.common_ports", []int{554, 8554})
	v.SetDefault("mediamtx.external_discovery.generic_uav.stream_paths", []string{"/stream", "/live", "/video"})
	v.SetDefault("mediamtx.external_discovery.generic_uav.known_ips", []string{})

	// MediaMTX stream readiness defaults
	v.SetDefault("mediamtx.stream_readiness.timeout", 15.0)
	v.SetDefault("mediamtx.stream_readiness.retry_attempts", 3)
	v.SetDefault("mediamtx.stream_readiness.retry_delay", 2.0)
	v.SetDefault("mediamtx.stream_readiness.check_interval", 0.5)
	v.SetDefault("mediamtx.stream_readiness.enable_progress_notifications", true)
	v.SetDefault("mediamtx.stream_readiness.graceful_fallback", true)

	// Performance fine-tuning defaults
	v.SetDefault("mediamtx.stream_readiness.controller_ticker_interval", 0.1)                             // 100ms for fast controller readiness
	v.SetDefault("mediamtx.stream_readiness.stream_manager_ticker_interval", 0.1)                         // 100ms for fast stream readiness
	v.SetDefault("mediamtx.stream_readiness.path_manager_retry_intervals", []float64{0.1, 0.2, 0.4, 0.8}) // Retry backoffs

	// FFmpeg defaults
	v.SetDefault("ffmpeg.snapshot.process_creation_timeout", 5.0)
	v.SetDefault("ffmpeg.snapshot.execution_timeout", 8.0)
	v.SetDefault("ffmpeg.snapshot.internal_timeout", 5000000)
	v.SetDefault("ffmpeg.snapshot.retry_attempts", 2)
	v.SetDefault("ffmpeg.snapshot.retry_delay", 1.0)

	v.SetDefault("ffmpeg.recording.process_creation_timeout", 10.0)
	v.SetDefault("ffmpeg.recording.execution_timeout", 15.0)
	v.SetDefault("ffmpeg.recording.internal_timeout", 10000000)
	v.SetDefault("ffmpeg.recording.retry_attempts", 3)
	v.SetDefault("ffmpeg.recording.retry_delay", 2.0)

	// Notifications defaults
	v.SetDefault("notifications.websocket.delivery_timeout", 5.0)
	v.SetDefault("notifications.websocket.retry_attempts", 3)
	v.SetDefault("notifications.websocket.retry_delay", 1.0)
	v.SetDefault("notifications.websocket.max_queue_size", 1000)
	v.SetDefault("notifications.websocket.cleanup_interval", 30)

	v.SetDefault("notifications.real_time.camera_status_interval", 1.0)
	v.SetDefault("notifications.real_time.recording_progress_interval", 0.5)
	v.SetDefault("notifications.real_time.connection_health_check", 10.0)

	// Performance defaults
	v.SetDefault("performance.response_time_targets.snapshot_capture", 2.0)
	v.SetDefault("performance.response_time_targets.recording_start", 2.0)
	v.SetDefault("performance.response_time_targets.recording_stop", 2.0)
	v.SetDefault("performance.response_time_targets.file_listing", 1.0)

	v.SetDefault("performance.snapshot_tiers.tier1_usb_direct_timeout", 2.0)
	v.SetDefault("performance.snapshot_tiers.tier2_rtsp_ready_check_timeout", 1.0)
	v.SetDefault("performance.snapshot_tiers.tier3_activation_timeout", 3.0)
	v.SetDefault("performance.snapshot_tiers.tier3_activation_trigger_timeout", 1.0)
	v.SetDefault("performance.snapshot_tiers.total_operation_timeout", 10.0)
	v.SetDefault("performance.snapshot_tiers.immediate_response_threshold", 0.5)
	v.SetDefault("performance.snapshot_tiers.acceptable_response_threshold", 2.0)
	v.SetDefault("performance.snapshot_tiers.slow_response_threshold", 5.0)

	v.SetDefault("performance.optimization.enable_caching", true)
	v.SetDefault("performance.optimization.cache_ttl", 300)
	v.SetDefault("performance.optimization.max_concurrent_operations", 5)
	v.SetDefault("performance.optimization.connection_pool_size", 10)

	// Monitoring thresholds defaults
	v.SetDefault("performance.monitoring_thresholds.memory_usage_percent", 90.0)
	v.SetDefault("performance.monitoring_thresholds.error_rate_percent", 5.0)
	v.SetDefault("performance.monitoring_thresholds.average_response_time_ms", 1000.0)
	v.SetDefault("performance.monitoring_thresholds.active_connections_limit", 900)
	v.SetDefault("performance.monitoring_thresholds.goroutines_limit", 1000)

	// Debounce configuration defaults
	v.SetDefault("performance.debounce.health_monitor_seconds", 15)
	v.SetDefault("performance.debounce.storage_monitor_seconds", 30)
	v.SetDefault("performance.debounce.performance_monitor_seconds", 45)

	// CORS security defaults
	v.SetDefault("security.cors_origins", []string{"http://localhost:3000", "https://localhost:3000"})
	v.SetDefault("security.cors_methods", []string{"GET", "POST", "OPTIONS"})
	v.SetDefault("security.cors_headers", []string{"Authorization", "Content-Type"})
	v.SetDefault("security.cors_credentials", false)

	// Camera defaults
	v.SetDefault("camera.poll_interval", 0.1)
	v.SetDefault("camera.detection_timeout", 2.0)
	v.SetDefault("camera.device_range", []int{0, 9})
	v.SetDefault("camera.enable_capability_detection", true)
	v.SetDefault("camera.auto_start_streams", true)
	v.SetDefault("camera.capability_timeout", 5.0)
	v.SetDefault("camera.capability_retry_interval", 1.0)
	v.SetDefault("camera.capability_max_retries", 3)

	// Retention policy defaults
	v.SetDefault("retention_policy.enabled", true)
	v.SetDefault("retention_policy.type", "age")
	v.SetDefault("retention_policy.max_age_days", 7)
	v.SetDefault("retention_policy.max_size_gb", 1)
	v.SetDefault("retention_policy.auto_cleanup", true)

	// Logging defaults - aligned with canonical configuration
	v.SetDefault("logging.level", "error") // Only critical errors by default
	v.SetDefault("logging.format", "json") // Structured logging for production
	v.SetDefault("logging.file_enabled", true)
	v.SetDefault("logging.file_path", "/var/log/camera-service.log")
	v.SetDefault("logging.max_file_size", 5242880) // 5MB for edge devices
	v.SetDefault("logging.backup_count", 3)
	v.SetDefault("logging.console_enabled", false) // Disabled for edge devices

	// Recording defaults
	v.SetDefault("recording.enabled", false)
	v.SetDefault("recording.format", "fmp4")
	v.SetDefault("recording.quality", "high")
	v.SetDefault("recording.segment_duration", 3600)
	v.SetDefault("recording.max_segment_size", 524288000)
	v.SetDefault("recording.auto_cleanup", true)
	v.SetDefault("recording.cleanup_interval", 86400)
	v.SetDefault("recording.max_age", 604800)
	v.SetDefault("recording.max_size", 10737418240)

	// Snapshots defaults
	v.SetDefault("snapshots.enabled", true)
	v.SetDefault("snapshots.format", "jpeg")
	v.SetDefault("snapshots.quality", 90)
	v.SetDefault("snapshots.max_width", 1920)
	v.SetDefault("snapshots.max_height", 1080)
	v.SetDefault("snapshots.auto_cleanup", true)
	v.SetDefault("snapshots.cleanup_interval", 3600)
	v.SetDefault("snapshots.max_age", 86400)
	v.SetDefault("snapshots.max_count", 1000)

	// Storage defaults
	v.SetDefault("storage.warn_percent", 80)
	v.SetDefault("storage.block_percent", 90)
	v.SetDefault("storage.default_path", "/opt/camera-service/recordings")
	v.SetDefault("storage.fallback_path", "/tmp/recordings")
}

// notifyConfigUpdated notifies all registered callbacks of configuration updates.
func (cm *ConfigManager) notifyConfigUpdated(oldConfig, newConfig *Config) {
	// Create error channel for callback panics
	panicChan := make(chan error, len(cm.updateCallbacks))

	// Create WaitGroup to track callback goroutines
	var callbackWg sync.WaitGroup

	for _, callback := range cm.updateCallbacks {
		callbackWg.Add(1)
		go func(cb func(*Config), config *Config) {
			defer callbackWg.Done()
			defer func() {
				// Recover from panics in goroutine and propagate as errors
				if r := recover(); r != nil {
					panicErr := fmt.Errorf("panic in config callback: %v", r)
					cm.logger.WithError(panicErr).Error("Config callback panic")

					// Propagate panic as error instead of swallowing it
					select {
					case panicChan <- panicErr:
					default:
						cm.logger.WithError(panicErr).Warn("Panic channel overflow, panic error dropped")
					}
				}
			}()
			cb(config)
		}(callback, newConfig)
	}

	// Process any panics that occurred in config callbacks
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		// Wait for all callbacks to complete
		callbackWg.Wait()
		// Close the panic channel to signal the processing goroutine to exit
		close(panicChan)
	}()

	// Process panic errors
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		for err := range panicChan {
			cm.logger.WithError(err).Warn("Config callback panic occurred")
			// Optionally increment error counters or trigger recovery mechanisms
		}
	}()
}

// getDefaultConfig returns a default configuration instance.
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                 "0.0.0.0",
			Port:                 8002,
			WebSocketPath:        "/ws",
			MaxConnections:       100,
			ReadTimeout:          5 * time.Second, // Standard timeout
			WriteTimeout:         1 * time.Second,
			PingInterval:         30 * time.Second,
			PongWait:             60 * time.Second,
			MaxMessageSize:       1024 * 1024, // 1MB
			ReadBufferSize:       1024,
			WriteBufferSize:      1024,
			ShutdownTimeout:      30 * time.Second,
			ClientCleanupTimeout: 10 * time.Second,
		},
		MediaMTX: MediaMTXConfig{
			Host:                                "127.0.0.1",
			APIPort:                             9997,
			RTSPPort:                            8554,
			WebRTCPort:                          8889,
			HLSPort:                             8888,
			ConfigPath:                          "/opt/camera-service/config/mediamtx.yml",
			RecordingsPath:                      "/opt/camera-service/recordings",
			SnapshotsPath:                       "/opt/camera-service/snapshots",
			HealthCheckInterval:                 30,
			HealthFailureThreshold:              10,
			HealthCircuitBreakerTimeout:         60,
			HealthMaxBackoffInterval:            120,
			HealthRecoveryConfirmationThreshold: 3,
			BackoffBaseMultiplier:               2.0,
			BackoffJitterRange:                  []float64{0.8, 1.2},
			ProcessTerminationTimeout:           3.0,
			ProcessKillTimeout:                  2.0,
		},
		Camera: CameraConfig{
			PollInterval:              0.1,
			DetectionTimeout:          2.0,
			DeviceRange:               []int{0, 9},
			EnableCapabilityDetection: true,
			AutoStartStreams:          true,
			CapabilityTimeout:         5.0,
			CapabilityRetryInterval:   1.0,
			CapabilityMaxRetries:      3,
		},
		Logging: LoggingConfig{
			Level:          "error", // Only critical errors by default
			Format:         "json",  // Structured logging for production
			FileEnabled:    true,
			FilePath:       "/var/log/camera-service.log",
			MaxFileSize:    5242880, // 5MB for edge devices
			BackupCount:    3,
			ConsoleEnabled: false, // Disabled for edge devices
		},
		Recording: RecordingConfig{
			Enabled:         false,
			Format:          "fmp4",
			Quality:         "high",
			SegmentDuration: 3600,
			MaxSegmentSize:  524288000,
			AutoCleanup:     true,
			CleanupInterval: 86400,
			MaxAge:          604800,
			MaxSize:         10737418240,
		},
		Snapshots: SnapshotConfig{
			Enabled:         true,
			Format:          "jpeg",
			Quality:         90,
			MaxWidth:        1920,
			MaxHeight:       1080,
			AutoCleanup:     true,
			CleanupInterval: 3600,
			MaxAge:          86400,
			MaxCount:        1000,
		},
		APIKeyManagement: APIKeyManagementConfig{
			StoragePath:      "/etc/camera-service/api-keys.json",
			EncryptionKey:    "", // Must be set via environment variable
			BackupEnabled:    true,
			BackupPath:       "/var/backups/camera-service/keys",
			BackupInterval:   "24h",
			KeyLength:        32,
			KeyPrefix:        "csk_",
			KeyFormat:        "base64url",
			DefaultExpiry:    "90d",
			RotationEnabled:  false,
			RotationInterval: "30d",
			MaxKeysPerRole:   10,
			AuditLogging:     true,
			UsageTracking:    true,
			CLIEnabled:       true,
			AdminInterface:   false,
			AdminPort:        8004,
		},
		HTTPHealth: HTTPHealthConfig{
			Enabled:           true,
			Host:              "0.0.0.0",
			Port:              8003,
			ReadTimeout:       "5s",
			WriteTimeout:      "5s",
			IdleTimeout:       "30s",
			BasicEndpoint:     "/health",
			DetailedEndpoint:  "/health/detailed",
			ReadyEndpoint:     "/health/ready",
			LiveEndpoint:      "/health/live",
			ResponseFormat:    "json",
			IncludeVersion:    true,
			IncludeUptime:     true,
			IncludeComponents: true,
			MaxResponseTime:   "100ms",
			EnableMetrics:     false,
			InternalOnly:      true,
			AllowedIPs:        []string{},
		},
	}
}

// applyDefaultsAfterUnmarshal applies default values to configuration fields that are zero
// This fixes the critical bug where incomplete YAML sections cause zero values to override Viper defaults
func (cm *ConfigManager) applyDefaultsAfterUnmarshal(config *Config) {
	// Apply defaults to StreamReadinessConfig if any critical fields are zero
	if config.MediaMTX.StreamReadiness.ControllerTickerInterval <= 0 {
		config.MediaMTX.StreamReadiness.ControllerTickerInterval = 0.1 // 100ms default
	}
	if config.MediaMTX.StreamReadiness.StreamManagerTickerInterval <= 0 {
		config.MediaMTX.StreamReadiness.StreamManagerTickerInterval = 0.1 // 100ms default
	}
	if config.MediaMTX.StreamReadiness.Timeout <= 0 {
		config.MediaMTX.StreamReadiness.Timeout = 15.0 // 15 seconds default
	}
	if config.MediaMTX.StreamReadiness.RetryAttempts <= 0 {
		config.MediaMTX.StreamReadiness.RetryAttempts = 3 // 3 attempts default
	}
	if config.MediaMTX.StreamReadiness.RetryDelay <= 0 {
		config.MediaMTX.StreamReadiness.RetryDelay = 1.0 // 1 second default
	}
	if config.MediaMTX.StreamReadiness.CheckInterval <= 0 {
		config.MediaMTX.StreamReadiness.CheckInterval = 0.5 // 500ms default
	}
	if config.MediaMTX.StreamReadiness.MaxCheckInterval <= 0 {
		config.MediaMTX.StreamReadiness.MaxCheckInterval = 2.0 // 2 seconds default
	}
	if config.MediaMTX.StreamReadiness.InitialCheckInterval <= 0 {
		config.MediaMTX.StreamReadiness.InitialCheckInterval = 0.2 // 200ms default
	}
	if len(config.MediaMTX.StreamReadiness.PathManagerRetryIntervals) == 0 {
		config.MediaMTX.StreamReadiness.PathManagerRetryIntervals = []float64{0.1, 0.2, 0.4, 0.8} // Default retry backoffs
	}
}
