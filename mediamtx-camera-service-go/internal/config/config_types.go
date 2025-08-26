package config

import "time"

// ServerConfig represents WebSocket server configuration settings.
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	WebSocketPath  string        `mapstructure:"websocket_path"`
	MaxConnections int           `mapstructure:"max_connections"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	PingInterval   time.Duration `mapstructure:"ping_interval"`
	PongWait       time.Duration `mapstructure:"pong_wait"`
	MaxMessageSize int64         `mapstructure:"max_message_size"`
}

// CodecConfig represents STANAG 4406 codec configuration settings.
type CodecConfig struct {
	VideoProfile string `mapstructure:"video_profile"`
	VideoLevel   string `mapstructure:"video_level"`
	PixelFormat  string `mapstructure:"pixel_format"`
	Bitrate      string `mapstructure:"bitrate"`
	Preset       string `mapstructure:"preset"`
}

// StreamReadinessConfig represents stream readiness configuration.
type StreamReadinessConfig struct {
	Timeout                     float64 `mapstructure:"timeout"`
	RetryAttempts               int     `mapstructure:"retry_attempts"`
	RetryDelay                  float64 `mapstructure:"retry_delay"`
	CheckInterval               float64 `mapstructure:"check_interval"`
	EnableProgressNotifications bool    `mapstructure:"enable_progress_notifications"`
	GracefulFallback            bool    `mapstructure:"graceful_fallback"`
}

// MediaMTXConfig represents MediaMTX integration configuration.
type MediaMTXConfig struct {
	Host                                string                `mapstructure:"host"`
	APIPort                             int                   `mapstructure:"api_port"`
	RTSPPort                            int                   `mapstructure:"rtsp_port"`
	WebRTCPort                          int                   `mapstructure:"webrtc_port"`
	HLSPort                             int                   `mapstructure:"hls_port"`
	ConfigPath                          string                `mapstructure:"config_path"`
	RecordingsPath                      string                `mapstructure:"recordings_path"`
	SnapshotsPath                       string                `mapstructure:"snapshots_path"`
	Codec                               CodecConfig           `mapstructure:"codec"`
	HealthCheckInterval                 int                   `mapstructure:"health_check_interval"`
	HealthFailureThreshold              int                   `mapstructure:"health_failure_threshold"`
	HealthCircuitBreakerTimeout         int                   `mapstructure:"health_circuit_breaker_timeout"`
	HealthMaxBackoffInterval            int                   `mapstructure:"health_max_backoff_interval"`
	HealthRecoveryConfirmationThreshold int                   `mapstructure:"health_recovery_confirmation_threshold"`
	BackoffBaseMultiplier               float64               `mapstructure:"backoff_base_multiplier"`
	BackoffJitterRange                  []float64             `mapstructure:"backoff_jitter_range"`
	ProcessTerminationTimeout           float64               `mapstructure:"process_termination_timeout"`
	ProcessKillTimeout                  float64               `mapstructure:"process_kill_timeout"`
	StreamReadiness                     StreamReadinessConfig `mapstructure:"stream_readiness"`
}

// FFmpegSnapshotConfig represents FFmpeg snapshot configuration.
type FFmpegSnapshotConfig struct {
	ProcessCreationTimeout float64 `mapstructure:"process_creation_timeout"`
	ExecutionTimeout       float64 `mapstructure:"execution_timeout"`
	InternalTimeout        int     `mapstructure:"internal_timeout"`
	RetryAttempts          int     `mapstructure:"retry_attempts"`
	RetryDelay             float64 `mapstructure:"retry_delay"`
}

// FFmpegRecordingConfig represents FFmpeg recording configuration.
type FFmpegRecordingConfig struct {
	ProcessCreationTimeout float64 `mapstructure:"process_creation_timeout"`
	ExecutionTimeout       float64 `mapstructure:"execution_timeout"`
	InternalTimeout        int     `mapstructure:"internal_timeout"`
	RetryAttempts          int     `mapstructure:"retry_attempts"`
	RetryDelay             float64 `mapstructure:"retry_delay"`
}

// FFmpegConfig represents FFmpeg configuration for performance tuning.
type FFmpegConfig struct {
	Snapshot  FFmpegSnapshotConfig  `mapstructure:"snapshot"`
	Recording FFmpegRecordingConfig `mapstructure:"recording"`
}

// WebSocketNotificationConfig represents WebSocket notification configuration.
type WebSocketNotificationConfig struct {
	DeliveryTimeout float64 `mapstructure:"delivery_timeout"`
	RetryAttempts   int     `mapstructure:"retry_attempts"`
	RetryDelay      float64 `mapstructure:"retry_delay"`
	MaxQueueSize    int     `mapstructure:"max_queue_size"`
	CleanupInterval int     `mapstructure:"cleanup_interval"`
}

// RealTimeNotificationConfig represents real-time notification configuration.
type RealTimeNotificationConfig struct {
	CameraStatusInterval      float64 `mapstructure:"camera_status_interval"`
	RecordingProgressInterval float64 `mapstructure:"recording_progress_interval"`
	ConnectionHealthCheck     float64 `mapstructure:"connection_health_check"`
}

// NotificationsConfig represents notification configuration for real-time updates.
type NotificationsConfig struct {
	WebSocket WebSocketNotificationConfig `mapstructure:"websocket"`
	RealTime  RealTimeNotificationConfig  `mapstructure:"real_time"`
}

// ResponseTimeTargetsConfig represents response time targets configuration.
type ResponseTimeTargetsConfig struct {
	SnapshotCapture float64 `mapstructure:"snapshot_capture"`
	RecordingStart  float64 `mapstructure:"recording_start"`
	RecordingStop   float64 `mapstructure:"recording_stop"`
	FileListing     float64 `mapstructure:"file_listing"`
}

// SnapshotTiersConfig represents multi-tier snapshot capture configuration.
type SnapshotTiersConfig struct {
	Tier1USBDirectTimeout         float64 `mapstructure:"tier1_usb_direct_timeout"`
	Tier2RTSPReadyCheckTimeout    float64 `mapstructure:"tier2_rtsp_ready_check_timeout"`
	Tier3ActivationTimeout        float64 `mapstructure:"tier3_activation_timeout"`
	Tier3ActivationTriggerTimeout float64 `mapstructure:"tier3_activation_trigger_timeout"`
	TotalOperationTimeout         float64 `mapstructure:"total_operation_timeout"`
	ImmediateResponseThreshold    float64 `mapstructure:"immediate_response_threshold"`
	AcceptableResponseThreshold   float64 `mapstructure:"acceptable_response_threshold"`
	SlowResponseThreshold         float64 `mapstructure:"slow_response_threshold"`
}

// OptimizationConfig represents performance optimization configuration.
type OptimizationConfig struct {
	EnableCaching           bool `mapstructure:"enable_caching"`
	CacheTTL                int  `mapstructure:"cache_ttl"`
	MaxConcurrentOperations int  `mapstructure:"max_concurrent_operations"`
	ConnectionPoolSize      int  `mapstructure:"connection_pool_size"`
}

// PerformanceConfig represents performance tuning configuration.
type PerformanceConfig struct {
	ResponseTimeTargets ResponseTimeTargetsConfig `mapstructure:"response_time_targets"`
	SnapshotTiers       SnapshotTiersConfig       `mapstructure:"snapshot_tiers"`
	Optimization        OptimizationConfig        `mapstructure:"optimization"`
}

// CameraConfig represents camera discovery configuration.
type CameraConfig struct {
	PollInterval              float64 `mapstructure:"poll_interval"`
	DetectionTimeout          float64 `mapstructure:"detection_timeout"`
	DeviceRange               []int   `mapstructure:"device_range"`
	EnableCapabilityDetection bool    `mapstructure:"enable_capability_detection"`
	AutoStartStreams          bool    `mapstructure:"auto_start_streams"`
	CapabilityTimeout         float64 `mapstructure:"capability_timeout"`
	CapabilityRetryInterval   float64 `mapstructure:"capability_retry_interval"`
	CapabilityMaxRetries      int     `mapstructure:"capability_max_retries"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level          string `mapstructure:"level"`
	Format         string `mapstructure:"format"`
	FileEnabled    bool   `mapstructure:"file_enabled"`
	FilePath       string `mapstructure:"file_path"`
	MaxFileSize    int64  `mapstructure:"max_file_size"`
	BackupCount    int    `mapstructure:"backup_count"`
	ConsoleEnabled bool   `mapstructure:"console_enabled"`
}

// RecordingConfig represents recording configuration.
type RecordingConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	Format          string `mapstructure:"format"`
	Quality         string `mapstructure:"quality"`
	SegmentDuration int    `mapstructure:"segment_duration"`
	MaxSegmentSize  int64  `mapstructure:"max_segment_size"`
	AutoCleanup     bool   `mapstructure:"auto_cleanup"`
	CleanupInterval int    `mapstructure:"cleanup_interval"`
	MaxAge          int    `mapstructure:"max_age"`
	MaxSize         int64  `mapstructure:"max_size"`
}

// SnapshotConfig represents snapshot configuration.
type SnapshotConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	Format          string `mapstructure:"format"`
	Quality         int    `mapstructure:"quality"`
	MaxWidth        int    `mapstructure:"max_width"`
	MaxHeight       int    `mapstructure:"max_height"`
	AutoCleanup     bool   `mapstructure:"auto_cleanup"`
	CleanupInterval int    `mapstructure:"cleanup_interval"`
	MaxAge          int    `mapstructure:"max_age"`
	MaxCount        int    `mapstructure:"max_count"`
}

// Config represents the complete service configuration.
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	MediaMTX      MediaMTXConfig      `mapstructure:"mediamtx"`
	Camera        CameraConfig        `mapstructure:"camera"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Recording     RecordingConfig     `mapstructure:"recording"`
	Snapshots     SnapshotConfig      `mapstructure:"snapshots"`
	FFmpeg        FFmpegConfig        `mapstructure:"ffmpeg"`
	Notifications NotificationsConfig `mapstructure:"notifications"`
	Performance   PerformanceConfig   `mapstructure:"performance"`
	HealthPort    *int                `mapstructure:"health_port"` // Optional health server port for testing
}
