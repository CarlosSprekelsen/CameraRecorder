package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ConfigManager manages configuration loading, validation, and hot reload functionality.
type ConfigManager struct {
	config          *Config
	configPath      string
	updateCallbacks []func(*Config)
	watcher         *fsnotify.Watcher
	watcherLock     sync.RWMutex
	lock            sync.RWMutex
	defaultConfig   *Config
	logger          *logrus.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewConfigManager creates a new configuration manager instance.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		updateCallbacks: make([]func(*Config), 0),
		defaultConfig:   getDefaultConfig(),
		logger:          logrus.New(),
		stopChan:        make(chan struct{}),
	}
}

// LoadConfig loads configuration from YAML file with environment variable overrides and validation.
func (cm *ConfigManager) LoadConfig(configPath string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.logger.WithFields(logrus.Fields{
		"config_path": configPath,
		"action":      "load_config",
	}).Info("Loading configuration")

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
		cm.logger.WithError(err).Warn("Failed to read config file, using defaults")
		// Continue with defaults
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
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

	cm.logger.WithFields(logrus.Fields{
		"config_path": configPath,
		"action":      "load_config",
		"status":      "success",
	}).Info("Configuration loaded successfully")

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

	// Start watching goroutine
	cm.wg.Add(1)
	go cm.watchFileChanges()

	cm.logger.WithFields(logrus.Fields{
		"config_path": cm.configPath,
		"watch_dir":   configDir,
	}).Info("File watching started for hot reload")

	return nil
}

// stopFileWatching stops the file watcher.
func (cm *ConfigManager) stopFileWatching() {
	cm.watcherLock.Lock()
	defer cm.watcherLock.Unlock()

	if cm.watcher != nil {
		// Close watcher safely
		if err := cm.watcher.Close(); err != nil {
			cm.logger.WithError(err).Warn("Error closing file watcher")
		}
		cm.watcher = nil
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
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}

			// Check if the changed file is our config file
			if event.Name == cm.configPath {
				cm.logger.WithFields(logrus.Fields{
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

		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			cm.logger.WithError(err).Error("File watcher error")
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

// Stop stops the configuration manager and cleans up resources.
func (cm *ConfigManager) Stop() {
	cm.logger.Info("Stopping configuration manager")

	// Signal stop
	close(cm.stopChan)

	// Stop file watching
	cm.stopFileWatching()

	// Wait for goroutines to finish
	cm.wg.Wait()

	cm.logger.Info("Configuration manager stopped")
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

// AddUpdateCallback adds a callback function to be called when configuration is updated.
func (cm *ConfigManager) AddUpdateCallback(callback func(*Config)) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.updateCallbacks = append(cm.updateCallbacks, callback)
}

// setDefaults sets default configuration values in Viper.
func (cm *ConfigManager) setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8002)
	v.SetDefault("server.websocket_path", "/ws")
	v.SetDefault("server.max_connections", 100)

	// MediaMTX defaults
	v.SetDefault("mediamtx.host", "127.0.0.1")
	v.SetDefault("mediamtx.api_port", 9997)
	v.SetDefault("mediamtx.rtsp_port", 8554)
	v.SetDefault("mediamtx.webrtc_port", 8889)
	v.SetDefault("mediamtx.hls_port", 8888)
	v.SetDefault("mediamtx.config_path", "/opt/camera-service/config/mediamtx.yml")
	v.SetDefault("mediamtx.recordings_path", "/opt/camera-service/recordings")
	v.SetDefault("mediamtx.snapshots_path", "/opt/camera-service/snapshots")

	// MediaMTX codec defaults
	v.SetDefault("mediamtx.codec.video_profile", "baseline")
	v.SetDefault("mediamtx.codec.video_level", "3.0")
	v.SetDefault("mediamtx.codec.pixel_format", "yuv420p")
	v.SetDefault("mediamtx.codec.bitrate", "600k")
	v.SetDefault("mediamtx.codec.preset", "ultrafast")

	// MediaMTX health monitoring defaults
	v.SetDefault("mediamtx.health_check_interval", 30)
	v.SetDefault("mediamtx.health_failure_threshold", 10)
	v.SetDefault("mediamtx.health_circuit_breaker_timeout", 60)
	v.SetDefault("mediamtx.health_max_backoff_interval", 120)
	v.SetDefault("mediamtx.health_recovery_confirmation_threshold", 3)
	v.SetDefault("mediamtx.backoff_base_multiplier", 2.0)
	v.SetDefault("mediamtx.backoff_jitter_range", []float64{0.8, 1.2})
	v.SetDefault("mediamtx.process_termination_timeout", 3.0)
	v.SetDefault("mediamtx.process_kill_timeout", 2.0)

	// MediaMTX stream readiness defaults
	v.SetDefault("mediamtx.stream_readiness.timeout", 15.0)
	v.SetDefault("mediamtx.stream_readiness.retry_attempts", 3)
	v.SetDefault("mediamtx.stream_readiness.retry_delay", 2.0)
	v.SetDefault("mediamtx.stream_readiness.check_interval", 0.5)
	v.SetDefault("mediamtx.stream_readiness.enable_progress_notifications", true)
	v.SetDefault("mediamtx.stream_readiness.graceful_fallback", true)

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

	v.SetDefault("performance.snapshot_tiers.tier1_usb_direct_timeout", 0.5)
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

	// Camera defaults
	v.SetDefault("camera.poll_interval", 0.1)
	v.SetDefault("camera.detection_timeout", 2.0)
	v.SetDefault("camera.device_range", []int{0, 9})
	v.SetDefault("camera.enable_capability_detection", true)
	v.SetDefault("camera.auto_start_streams", true)
	v.SetDefault("camera.capability_timeout", 5.0)
	v.SetDefault("camera.capability_retry_interval", 1.0)
	v.SetDefault("camera.capability_max_retries", 3)

	// Logging defaults
	v.SetDefault("logging.level", "INFO")
	v.SetDefault("logging.format", "%(asctime)s - %(name)s - %(levelname)s - %(message)s")
	v.SetDefault("logging.file_enabled", true)
	v.SetDefault("logging.file_path", "/opt/camera-service/logs/camera-service.log")
	v.SetDefault("logging.max_file_size", 10485760)
	v.SetDefault("logging.backup_count", 5)
	v.SetDefault("logging.console_enabled", true)

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
}

// applyEnvironmentOverrides applies environment variable overrides to configuration.
func (cm *ConfigManager) applyEnvironmentOverrides(config *Config) {
	// Map of environment variable patterns to config paths
	envMappings := map[string]string{
		"CAMERA_SERVICE_SERVER_HOST":                        "server.host",
		"CAMERA_SERVICE_SERVER_PORT":                        "server.port",
		"CAMERA_SERVICE_SERVER_WEBSOCKET_PATH":              "server.websocket_path",
		"CAMERA_SERVICE_SERVER_MAX_CONNECTIONS":             "server.max_connections",
		"CAMERA_SERVICE_MEDIAMTX_HOST":                      "mediamtx.host",
		"CAMERA_SERVICE_MEDIAMTX_API_PORT":                  "mediamtx.api_port",
		"CAMERA_SERVICE_MEDIAMTX_RTSP_PORT":                 "mediamtx.rtsp_port",
		"CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT":               "mediamtx.webrtc_port",
		"CAMERA_SERVICE_MEDIAMTX_HLS_PORT":                  "mediamtx.hls_port",
		"CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH":               "mediamtx.config_path",
		"CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH":           "mediamtx.recordings_path",
		"CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH":            "mediamtx.snapshots_path",
		"CAMERA_SERVICE_CAMERA_POLL_INTERVAL":               "camera.poll_interval",
		"CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT":           "camera.detection_timeout",
		"CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION": "camera.enable_capability_detection",
		"CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS":          "camera.auto_start_streams",
		"CAMERA_SERVICE_LOGGING_LEVEL":                      "logging.level",
		"CAMERA_SERVICE_LOGGING_FORMAT":                     "logging.format",
		"CAMERA_SERVICE_LOGGING_FILE_ENABLED":               "logging.file_enabled",
		"CAMERA_SERVICE_LOGGING_FILE_PATH":                  "logging.file_path",
		"CAMERA_SERVICE_LOGGING_CONSOLE_ENABLED":            "logging.console_enabled",
		"CAMERA_SERVICE_RECORDING_ENABLED":                  "recording.enabled",
		"CAMERA_SERVICE_RECORDING_FORMAT":                   "recording.format",
		"CAMERA_SERVICE_RECORDING_QUALITY":                  "recording.quality",
		"CAMERA_SERVICE_SNAPSHOTS_ENABLED":                  "snapshots.enabled",
		"CAMERA_SERVICE_SNAPSHOTS_FORMAT":                   "snapshots.format",
		"CAMERA_SERVICE_SNAPSHOTS_QUALITY":                  "snapshots.quality",
	}

	for envVar, configPath := range envMappings {
		if value := os.Getenv(envVar); value != "" {
			cm.logger.WithFields(logrus.Fields{
				"env_var":     envVar,
				"config_path": configPath,
				"value":       value,
			}).Debug("Applying environment variable override")

			// Note: Viper handles the actual override during unmarshaling
		}
	}
}

// validateConfig validates the configuration and returns an error if invalid.
// This is a legacy method - use ValidateConfig() for comprehensive validation.
func (cm *ConfigManager) validateConfig(config *Config) error {
	return ValidateConfig(config)
}

// notifyConfigUpdated notifies all registered callbacks of configuration updates.
func (cm *ConfigManager) notifyConfigUpdated(oldConfig, newConfig *Config) {
	for _, callback := range cm.updateCallbacks {
		go func(cb func(*Config), config *Config) {
			defer func() {
				if r := recover(); r != nil {
					cm.logger.WithError(fmt.Errorf("panic in config callback: %v", r)).Error("Config callback panic")
				}
			}()
			cb(config)
		}(callback, newConfig)
	}
}

// getDefaultConfig returns a default configuration instance.
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           8002,
			WebSocketPath:  "/ws",
			MaxConnections: 100,
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
			Level:          "INFO",
			Format:         "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
			FileEnabled:    true,
			FilePath:       "/opt/camera-service/logs/camera-service.log",
			MaxFileSize:    10485760,
			BackupCount:    5,
			ConsoleEnabled: true,
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
	}
}

// Global configuration manager instance
var globalConfigManager *ConfigManager
var globalConfigManagerOnce sync.Once

// GetConfigManager returns the global configuration manager instance.
func GetConfigManager() *ConfigManager {
	globalConfigManagerOnce.Do(func() {
		globalConfigManager = NewConfigManager()
	})
	return globalConfigManager
}

// LoadConfig is a convenience function to load configuration using the global manager.
func LoadConfig(configPath string) error {
	return GetConfigManager().LoadConfig(configPath)
}

// GetConfig is a convenience function to get the current configuration.
func GetConfig() *Config {
	return GetConfigManager().GetConfig()
}
