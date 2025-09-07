package config

import "time"

// ServerConfig represents WebSocket server configuration settings.
type ServerConfig struct {
	Host                 string        `mapstructure:"host"`
	Port                 int           `mapstructure:"port"`
	WebSocketPath        string        `mapstructure:"websocket_path"`
	MaxConnections       int           `mapstructure:"max_connections"`
	ReadTimeout          time.Duration `mapstructure:"read_timeout"`
	WriteTimeout         time.Duration `mapstructure:"write_timeout"`
	PingInterval         time.Duration `mapstructure:"ping_interval"`
	PongWait             time.Duration `mapstructure:"pong_wait"`
	MaxMessageSize       int64         `mapstructure:"max_message_size"`
	ReadBufferSize       int           `mapstructure:"read_buffer_size"`
	WriteBufferSize      int           `mapstructure:"write_buffer_size"`
	ShutdownTimeout      time.Duration `mapstructure:"shutdown_timeout"`
	ClientCleanupTimeout time.Duration `mapstructure:"client_cleanup_timeout"`
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

// SecurityConfig represents security configuration settings.
type SecurityConfig struct {
	RateLimitRequests int           `mapstructure:"rate_limit_requests"` // Default: 100 requests per window
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`   // Default: 1 minute
	JWTSecretKey      string        `mapstructure:"jwt_secret_key"`
	JWTExpiryHours    int           `mapstructure:"jwt_expiry_hours"` // Default: 24 hours
}

// StorageConfig represents storage configuration settings.
type StorageConfig struct {
	WarnPercent  int    `mapstructure:"warn_percent"`  // Default: 80% usage warning
	BlockPercent int    `mapstructure:"block_percent"` // Default: 90% usage block
	DefaultPath  string `mapstructure:"default_path"`  // Default: "/opt/camera-service/recordings"
	FallbackPath string `mapstructure:"fallback_path"` // Default: "/tmp/recordings"
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
	HealthCheckTimeout                  time.Duration         `mapstructure:"health_check_timeout"` // Default: 5 seconds
	// Health monitoring defaults
	HealthMonitorDefaults HealthMonitorDefaults `mapstructure:"health_monitor_defaults"`

	// RTSP Connection Monitoring Configuration
	RTSPMonitoring RTSPMonitoringConfig `mapstructure:"rtsp_monitoring"`
}

// HealthMonitorDefaults represents health monitoring default values
type HealthMonitorDefaults struct {
	CheckInterval   float64 `mapstructure:"check_interval"`    // Default: 5.0 seconds
	MaxBackoffDelay float64 `mapstructure:"max_backoff_delay"` // Default: 30.0 seconds
	ShutdownTimeout float64 `mapstructure:"shutdown_timeout"`  // Default: 30.0 seconds
}

// RTSPMonitoringConfig represents RTSP connection monitoring configuration
type RTSPMonitoringConfig struct {
	Enabled             bool    `mapstructure:"enabled"`               // Default: true
	CheckInterval       int     `mapstructure:"check_interval"`        // Default: 30 seconds
	ConnectionTimeout   int     `mapstructure:"connection_timeout"`    // Default: 10 seconds
	MaxConnections      int     `mapstructure:"max_connections"`       // Default: 50
	SessionTimeout      int     `mapstructure:"session_timeout"`       // Default: 300 seconds
	BandwidthThreshold  int64   `mapstructure:"bandwidth_threshold"`   // Default: 1000000 bytes/sec
	PacketLossThreshold float64 `mapstructure:"packet_loss_threshold"` // Default: 0.05 (5%)
	JitterThreshold     float64 `mapstructure:"jitter_threshold"`      // Default: 50.0 ms
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
	// Fallback defaults for when configuration is missing
	FallbackDefaults FFmpegFallbackDefaults `mapstructure:"fallback_defaults"`
}

// FFmpegFallbackDefaults represents fallback defaults for FFmpeg operations
type FFmpegFallbackDefaults struct {
	RetryDelay             float64 `mapstructure:"retry_delay"`              // Default: 1.0 second
	ProcessCreationTimeout float64 `mapstructure:"process_creation_timeout"` // Default: 10.0 seconds
	ExecutionTimeout       float64 `mapstructure:"execution_timeout"`        // Default: 30.0 seconds
	MaxBackoffDelay        float64 `mapstructure:"max_backoff_delay"`        // Default: 30.0 seconds
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
	Enabled              bool          `mapstructure:"enabled"`
	Format               string        `mapstructure:"format"`
	Quality              string        `mapstructure:"quality"`
	SegmentDuration      int           `mapstructure:"segment_duration"`
	MaxSegmentSize       int64         `mapstructure:"max_segment_size"`
	AutoCleanup          bool          `mapstructure:"auto_cleanup"`
	CleanupInterval      int           `mapstructure:"cleanup_interval"`
	MaxAge               int           `mapstructure:"max_age"`
	MaxSize              int64         `mapstructure:"max_size"`
	DefaultRotationSize  int64         `mapstructure:"default_rotation_size"`  // Default: 100MB
	DefaultMaxDuration   time.Duration `mapstructure:"default_max_duration"`   // Default: 24 hours
	DefaultRetentionDays int           `mapstructure:"default_retention_days"` // Default: 7 days
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

// RetentionPolicyConfig represents file retention policy configuration.
type RetentionPolicyConfig struct {
	Enabled     bool   `mapstructure:"enabled"`      // Whether retention policies are enabled
	Type        string `mapstructure:"type"`         // "age", "size", or "manual"
	MaxAgeDays  int    `mapstructure:"max_age_days"` // For age-based policy (default: 7)
	MaxSizeGB   int    `mapstructure:"max_size_gb"`  // For size-based policy (default: 1)
	AutoCleanup bool   `mapstructure:"auto_cleanup"` // Whether to automatically clean up files
}

// Config represents the complete service configuration.
type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	MediaMTX        MediaMTXConfig        `mapstructure:"mediamtx"`
	Camera          CameraConfig          `mapstructure:"camera"`
	Logging         LoggingConfig         `mapstructure:"logging"`
	Recording       RecordingConfig       `mapstructure:"recording"`
	Snapshots       SnapshotConfig        `mapstructure:"snapshots"`
	FFmpeg          FFmpegConfig          `mapstructure:"ffmpeg"`
	Notifications   NotificationsConfig   `mapstructure:"notifications"`
	Performance     PerformanceConfig     `mapstructure:"performance"`
	Security        SecurityConfig        `mapstructure:"security"`
	Storage         StorageConfig         `mapstructure:"storage"`
	RetentionPolicy RetentionPolicyConfig `mapstructure:"retention_policy"`
	// Server operation defaults
	ServerDefaults ServerDefaults `mapstructure:"server_defaults"`
}

// ServerDefaults represents server operation default values
type ServerDefaults struct {
	ShutdownTimeout     float64 `mapstructure:"shutdown_timeout"`      // Default: 30.0 seconds
	CameraMonitorTicker float64 `mapstructure:"camera_monitor_ticker"` // Default: 5.0 seconds
}
