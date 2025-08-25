package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ConfigLoader handles configuration loading using Viper.
type ConfigLoader struct {
	viper  *viper.Viper
	logger *logrus.Logger
}

// NewConfigLoader creates a new configuration loader.
func NewConfigLoader() *ConfigLoader {
	v := viper.New()
	
	// Set configuration file type
	v.SetConfigType("yaml")
	
	// Set environment variable prefix
	v.SetEnvPrefix("CAMERA_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Enable environment variable binding
	v.AutomaticEnv()
	
	return &ConfigLoader{
		viper:  v,
		logger: logrus.New(),
	}
}

// LoadConfig loads configuration from the specified file path.
func (cl *ConfigLoader) LoadConfig(configPath string) (*Config, error) {
	// Set configuration file path
	cl.viper.SetConfigFile(configPath)
	
	// Set default values (matching Python defaults)
	cl.setDefaults()
	
	// Read configuration file
	if err := cl.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			cl.logger.Warn("Configuration file not found, using defaults")
		} else {
			// Check if it's a file not found error
			if os.IsNotExist(err) {
				cl.logger.Warn("Configuration file not found, using defaults")
			} else {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}
	
	// Unmarshal into Config struct
	var config Config
	if err := cl.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	cl.logger.Info("Configuration loaded successfully")
	return &config, nil
}

// setDefaults sets all default configuration values matching Python defaults.
func (cl *ConfigLoader) setDefaults() {
	// Set server defaults
	cl.viper.SetDefault("server.host", "0.0.0.0")
	cl.viper.SetDefault("server.port", 8002)
	cl.viper.SetDefault("server.websocket_path", "/ws")
	cl.viper.SetDefault("server.max_connections", 100)

	// Set MediaMTX defaults
	cl.viper.SetDefault("mediamtx.host", "127.0.0.1")
	cl.viper.SetDefault("mediamtx.api_port", 9997)
	cl.viper.SetDefault("mediamtx.rtsp_port", 8554)
	cl.viper.SetDefault("mediamtx.webrtc_port", 8889)
	cl.viper.SetDefault("mediamtx.hls_port", 8888)
	cl.viper.SetDefault("mediamtx.config_path", "/opt/camera-service/config/mediamtx.yml")
	cl.viper.SetDefault("mediamtx.recordings_path", "/opt/camera-service/recordings")
	cl.viper.SetDefault("mediamtx.snapshots_path", "/opt/camera-service/snapshots")
	
	// Codec defaults
	cl.viper.SetDefault("mediamtx.codec.video_profile", "baseline")
	cl.viper.SetDefault("mediamtx.codec.video_level", "3.0")
	cl.viper.SetDefault("mediamtx.codec.pixel_format", "yuv420p")
	cl.viper.SetDefault("mediamtx.codec.bitrate", "600k")
	cl.viper.SetDefault("mediamtx.codec.preset", "ultrafast")
	
	// Health monitoring defaults
	cl.viper.SetDefault("mediamtx.health_check_interval", 30)
	cl.viper.SetDefault("mediamtx.health_failure_threshold", 10)
	cl.viper.SetDefault("mediamtx.health_circuit_breaker_timeout", 60)
	cl.viper.SetDefault("mediamtx.health_max_backoff_interval", 120)
	cl.viper.SetDefault("mediamtx.health_recovery_confirmation_threshold", 3)
	cl.viper.SetDefault("mediamtx.backoff_base_multiplier", 2.0)
	cl.viper.SetDefault("mediamtx.backoff_jitter_range", []float64{0.8, 1.2})
	cl.viper.SetDefault("mediamtx.process_termination_timeout", 3.0)
	cl.viper.SetDefault("mediamtx.process_kill_timeout", 2.0)
	
	// Stream readiness defaults
	cl.viper.SetDefault("mediamtx.stream_readiness.timeout", 15.0)
	cl.viper.SetDefault("mediamtx.stream_readiness.retry_attempts", 3)
	cl.viper.SetDefault("mediamtx.stream_readiness.retry_delay", 2.0)
	cl.viper.SetDefault("mediamtx.stream_readiness.check_interval", 0.5)
	cl.viper.SetDefault("mediamtx.stream_readiness.enable_progress_notifications", true)
	cl.viper.SetDefault("mediamtx.stream_readiness.graceful_fallback", true)

	// Set camera defaults
	cl.viper.SetDefault("camera.poll_interval", 0.1)
	cl.viper.SetDefault("camera.detection_timeout", 2.0)
	cl.viper.SetDefault("camera.device_range", []int{0, 9})
	cl.viper.SetDefault("camera.enable_capability_detection", true)
	cl.viper.SetDefault("camera.auto_start_streams", true)
	cl.viper.SetDefault("camera.capability_timeout", 5.0)
	cl.viper.SetDefault("camera.capability_retry_interval", 1.0)
	cl.viper.SetDefault("camera.capability_max_retries", 3)

	// Set logging defaults
	cl.viper.SetDefault("logging.level", "INFO")
	cl.viper.SetDefault("logging.format", "%(asctime)s - %(name)s - %(levelname)s - %(message)s")
	cl.viper.SetDefault("logging.file_enabled", true)
	cl.viper.SetDefault("logging.file_path", "/opt/camera-service/logs/camera-service.log")
	cl.viper.SetDefault("logging.max_file_size", 10485760)
	cl.viper.SetDefault("logging.backup_count", 5)
	cl.viper.SetDefault("logging.console_enabled", true)

	// Set recording defaults
	cl.viper.SetDefault("recording.enabled", false)
	cl.viper.SetDefault("recording.format", "fmp4")
	cl.viper.SetDefault("recording.quality", "high")
	cl.viper.SetDefault("recording.segment_duration", 3600)
	cl.viper.SetDefault("recording.max_segment_size", 524288000)
	cl.viper.SetDefault("recording.auto_cleanup", true)
	cl.viper.SetDefault("recording.cleanup_interval", 86400)
	cl.viper.SetDefault("recording.max_age", 604800)
	cl.viper.SetDefault("recording.max_size", 10737418240)

	// Set snapshots defaults
	cl.viper.SetDefault("snapshots.enabled", true)
	cl.viper.SetDefault("snapshots.format", "jpeg")
	cl.viper.SetDefault("snapshots.quality", 90)
	cl.viper.SetDefault("snapshots.max_width", 1920)
	cl.viper.SetDefault("snapshots.max_height", 1080)
	cl.viper.SetDefault("snapshots.auto_cleanup", true)
	cl.viper.SetDefault("snapshots.cleanup_interval", 3600)
	cl.viper.SetDefault("snapshots.max_age", 86400)
	cl.viper.SetDefault("snapshots.max_count", 1000)

	// Set FFmpeg defaults
	cl.viper.SetDefault("ffmpeg.snapshot.process_creation_timeout", 5.0)
	cl.viper.SetDefault("ffmpeg.snapshot.execution_timeout", 8.0)
	cl.viper.SetDefault("ffmpeg.snapshot.internal_timeout", 5000000)
	cl.viper.SetDefault("ffmpeg.snapshot.retry_attempts", 2)
	cl.viper.SetDefault("ffmpeg.snapshot.retry_delay", 1.0)
	
	cl.viper.SetDefault("ffmpeg.recording.process_creation_timeout", 10.0)
	cl.viper.SetDefault("ffmpeg.recording.execution_timeout", 15.0)
	cl.viper.SetDefault("ffmpeg.recording.internal_timeout", 10000000)
	cl.viper.SetDefault("ffmpeg.recording.retry_attempts", 3)
	cl.viper.SetDefault("ffmpeg.recording.retry_delay", 2.0)

	// Set notifications defaults
	cl.viper.SetDefault("notifications.websocket.delivery_timeout", 5.0)
	cl.viper.SetDefault("notifications.websocket.retry_attempts", 3)
	cl.viper.SetDefault("notifications.websocket.retry_delay", 1.0)
	cl.viper.SetDefault("notifications.websocket.max_queue_size", 1000)
	cl.viper.SetDefault("notifications.websocket.cleanup_interval", 30)
	
	cl.viper.SetDefault("notifications.real_time.camera_status_interval", 1.0)
	cl.viper.SetDefault("notifications.real_time.recording_progress_interval", 0.5)
	cl.viper.SetDefault("notifications.real_time.connection_health_check", 10.0)

	// Set performance defaults
	cl.viper.SetDefault("performance.response_time_targets.snapshot_capture", 2.0)
	cl.viper.SetDefault("performance.response_time_targets.recording_start", 2.0)
	cl.viper.SetDefault("performance.response_time_targets.recording_stop", 2.0)
	cl.viper.SetDefault("performance.response_time_targets.file_listing", 1.0)
	
	cl.viper.SetDefault("performance.snapshot_tiers.tier1_usb_direct_timeout", 0.5)
	cl.viper.SetDefault("performance.snapshot_tiers.tier2_rtsp_ready_check_timeout", 1.0)
	cl.viper.SetDefault("performance.snapshot_tiers.tier3_activation_timeout", 3.0)
	cl.viper.SetDefault("performance.snapshot_tiers.tier3_activation_trigger_timeout", 1.0)
	cl.viper.SetDefault("performance.snapshot_tiers.total_operation_timeout", 10.0)
	cl.viper.SetDefault("performance.snapshot_tiers.immediate_response_threshold", 0.5)
	cl.viper.SetDefault("performance.snapshot_tiers.acceptable_response_threshold", 2.0)
	cl.viper.SetDefault("performance.snapshot_tiers.slow_response_threshold", 5.0)
	
	cl.viper.SetDefault("performance.optimization.enable_caching", true)
	cl.viper.SetDefault("performance.optimization.cache_ttl", 300)
	cl.viper.SetDefault("performance.optimization.max_concurrent_operations", 5)
	cl.viper.SetDefault("performance.optimization.connection_pool_size", 10)
}

// GetViper returns the underlying Viper instance for advanced usage.
func (cl *ConfigLoader) GetViper() *viper.Viper {
	return cl.viper
}
