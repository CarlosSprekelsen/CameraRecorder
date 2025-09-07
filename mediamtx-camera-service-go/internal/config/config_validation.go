package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidateConfig performs comprehensive validation of the configuration.
func ValidateConfig(config *Config) error {
	var errors []error

	// Validate each configuration section
	if err := validateServerConfig(&config.Server); err != nil {
		errors = append(errors, err)
	}

	if err := validateMediaMTXConfig(&config.MediaMTX); err != nil {
		errors = append(errors, err)
	}

	if err := validateCameraConfig(&config.Camera); err != nil {
		errors = append(errors, err)
	}

	if err := validateLoggingConfig(&config.Logging); err != nil {
		errors = append(errors, err)
	}

	if err := validateRecordingConfig(&config.Recording); err != nil {
		errors = append(errors, err)
	}

	if err := validateSnapshotConfig(&config.Snapshots); err != nil {
		errors = append(errors, err)
	}

	if err := validateFFmpegConfig(&config.FFmpeg); err != nil {
		errors = append(errors, err)
	}

	if err := validateNotificationsConfig(&config.Notifications); err != nil {
		errors = append(errors, err)
	}

	if err := validatePerformanceConfig(&config.Performance); err != nil {
		errors = append(errors, err)
	}

	if err := validateRetentionPolicyConfig(&config.RetentionPolicy); err != nil {
		errors = append(errors, err)
	}

	if err := validateSecurityConfig(&config.Security); err != nil {
		errors = append(errors, err)
	}

	if err := validateServerDefaults(&config.ServerDefaults); err != nil {
		errors = append(errors, err)
	}

	if err := validateStorageConfig(&config.Storage); err != nil {
		errors = append(errors, err)
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return &ValidationError{
			Field:   "config",
			Message: fmt.Sprintf("configuration validation failed: %v", errors),
		}
	}

	return nil
}

// validateServerConfig validates server configuration.
func validateServerConfig(config *ServerConfig) error {
	if config.Host == "" {
		return &ValidationError{Field: "server.host", Message: "host cannot be empty"}
	}

	if config.Port <= 0 || config.Port > 65535 {
		return &ValidationError{Field: "server.port", Message: fmt.Sprintf("port must be between 1 and 65535, got %d", config.Port)}
	}

	if config.WebSocketPath == "" {
		return &ValidationError{Field: "server.websocket_path", Message: "websocket path cannot be empty"}
	}

	if !strings.HasPrefix(config.WebSocketPath, "/") {
		return &ValidationError{Field: "server.websocket_path", Message: "websocket path must start with '/'"}
	}

	if config.MaxConnections <= 0 {
		return &ValidationError{Field: "server.max_connections", Message: fmt.Sprintf("max connections must be positive, got %d", config.MaxConnections)}
	}

	if config.ReadTimeout <= 0 {
		return &ValidationError{Field: "server.read_timeout", Message: fmt.Sprintf("read timeout must be positive, got %v", config.ReadTimeout)}
	}

	if config.WriteTimeout <= 0 {
		return &ValidationError{Field: "server.write_timeout", Message: fmt.Sprintf("write timeout must be positive, got %v", config.WriteTimeout)}
	}

	if config.PingInterval <= 0 {
		return &ValidationError{Field: "server.ping_interval", Message: fmt.Sprintf("ping interval must be positive, got %v", config.PingInterval)}
	}

	if config.PongWait <= 0 {
		return &ValidationError{Field: "server.pong_wait", Message: fmt.Sprintf("pong wait must be positive, got %v", config.PongWait)}
	}

	if config.MaxMessageSize <= 0 {
		return &ValidationError{Field: "server.max_message_size", Message: fmt.Sprintf("max message size must be positive, got %d", config.MaxMessageSize)}
	}

	if config.ReadBufferSize <= 0 {
		return &ValidationError{Field: "server.read_buffer_size", Message: fmt.Sprintf("read buffer size must be positive, got %d", config.ReadBufferSize)}
	}

	if config.WriteBufferSize <= 0 {
		return &ValidationError{Field: "server.write_buffer_size", Message: fmt.Sprintf("write buffer size must be positive, got %d", config.WriteBufferSize)}
	}

	if config.ShutdownTimeout <= 0 {
		return &ValidationError{Field: "server.shutdown_timeout", Message: fmt.Sprintf("shutdown timeout must be positive, got %v", config.ShutdownTimeout)}
	}

	if config.ClientCleanupTimeout <= 0 {
		return &ValidationError{Field: "server.client_cleanup_timeout", Message: fmt.Sprintf("client cleanup timeout must be positive, got %v", config.ClientCleanupTimeout)}
	}

	return nil
}

// validateMediaMTXConfig validates MediaMTX configuration.
func validateMediaMTXConfig(config *MediaMTXConfig) error {
	if config.Host == "" {
		return &ValidationError{Field: "mediamtx.host", Message: "host cannot be empty"}
	}

	if config.APIPort <= 0 || config.APIPort > 65535 {
		return &ValidationError{Field: "mediamtx.api_port", Message: fmt.Sprintf("API port must be between 1 and 65535, got %d", config.APIPort)}
	}

	if config.RTSPPort <= 0 || config.RTSPPort > 65535 {
		return &ValidationError{Field: "mediamtx.rtsp_port", Message: fmt.Sprintf("RTSP port must be between 1 and 65535, got %d", config.RTSPPort)}
	}

	if config.WebRTCPort <= 0 || config.WebRTCPort > 65535 {
		return &ValidationError{Field: "mediamtx.webrtc_port", Message: fmt.Sprintf("WebRTC port must be between 1 and 65535, got %d", config.WebRTCPort)}
	}

	if config.HLSPort <= 0 || config.HLSPort > 65535 {
		return &ValidationError{Field: "mediamtx.hls_port", Message: fmt.Sprintf("HLS port must be between 1 and 65535, got %d", config.HLSPort)}
	}

	if config.ConfigPath == "" {
		return &ValidationError{Field: "mediamtx.config_path", Message: "config path cannot be empty"}
	}

	if config.RecordingsPath == "" {
		return &ValidationError{Field: "mediamtx.recordings_path", Message: "recordings path cannot be empty"}
	}

	if config.SnapshotsPath == "" {
		return &ValidationError{Field: "mediamtx.snapshots_path", Message: "snapshots path cannot be empty"}
	}

	// Validate file paths exist and are accessible
	if err := validateFilePath("mediamtx.config_path", config.ConfigPath); err != nil {
		return err
	}

	if err := validateFilePath("mediamtx.recordings_path", config.RecordingsPath); err != nil {
		return err
	}

	if err := validateFilePath("mediamtx.snapshots_path", config.SnapshotsPath); err != nil {
		return err
	}

	// Validate codec configuration
	if err := validateCodecConfig(&config.Codec); err != nil {
		return fmt.Errorf("failed to validate codec configuration: %w", err)
	}

	// Validate health monitoring configuration
	if config.HealthCheckInterval <= 0 {
		return &ValidationError{Field: "mediamtx.health_check_interval", Message: fmt.Sprintf("health check interval must be positive, got %d", config.HealthCheckInterval)}
	}

	if config.HealthFailureThreshold <= 0 {
		return &ValidationError{Field: "mediamtx.health_failure_threshold", Message: fmt.Sprintf("health failure threshold must be positive, got %d", config.HealthFailureThreshold)}
	}

	if config.HealthCircuitBreakerTimeout <= 0 {
		return &ValidationError{Field: "mediamtx.health_circuit_breaker_timeout", Message: fmt.Sprintf("circuit breaker timeout must be positive, got %d", config.HealthCircuitBreakerTimeout)}
	}

	if config.HealthMaxBackoffInterval <= 0 {
		return &ValidationError{Field: "mediamtx.health_max_backoff_interval", Message: fmt.Sprintf("max backoff interval must be positive, got %d", config.HealthMaxBackoffInterval)}
	}

	if config.HealthRecoveryConfirmationThreshold <= 0 {
		return &ValidationError{Field: "mediamtx.health_recovery_confirmation_threshold", Message: fmt.Sprintf("recovery confirmation threshold must be positive, got %d", config.HealthRecoveryConfirmationThreshold)}
	}

	if config.BackoffBaseMultiplier <= 0 {
		return &ValidationError{Field: "mediamtx.backoff_base_multiplier", Message: fmt.Sprintf("backoff base multiplier must be positive, got %f", config.BackoffBaseMultiplier)}
	}

	if len(config.BackoffJitterRange) != 2 {
		return &ValidationError{Field: "mediamtx.backoff_jitter_range", Message: "backoff jitter range must have exactly 2 elements"}
	}

	if config.BackoffJitterRange[0] >= config.BackoffJitterRange[1] {
		return &ValidationError{Field: "mediamtx.backoff_jitter_range", Message: "backoff jitter range min must be less than max"}
	}

	if config.ProcessTerminationTimeout <= 0 {
		return &ValidationError{Field: "mediamtx.process_termination_timeout", Message: fmt.Sprintf("process termination timeout must be positive, got %f", config.ProcessTerminationTimeout)}
	}

	if config.ProcessKillTimeout <= 0 {
		return &ValidationError{Field: "mediamtx.process_kill_timeout", Message: fmt.Sprintf("process kill timeout must be positive, got %f", config.ProcessKillTimeout)}
	}

	// Validate RTSP monitoring configuration
	if err := validateRTSPMonitoringConfig(&config.RTSPMonitoring); err != nil {
		return err
	}

	// Validate stream readiness configuration
	if err := validateStreamReadinessConfig(&config.StreamReadiness); err != nil {
		return fmt.Errorf("failed to validate stream readiness configuration: %w", err)
	}

	return nil
}

// validateFilePath validates that a file path exists and is accessible
func validateFilePath(fieldName, path string) error {
	if path == "" {
		return &ValidationError{Field: fieldName, Message: "path cannot be empty"}
	}

	// Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("path contains invalid characters or traversal attempts: %s", path)}
	}

	// Check if path is absolute (recommended for security)
	if !filepath.IsAbs(cleanPath) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("path should be absolute for security: %s", path)}
	}

	// Check if path exists on filesystem
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("path does not exist: %s", path)}
	}

	// Check if path is accessible (readable)
	if _, err := os.Stat(cleanPath); err != nil {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("path is not accessible: %s - %v", path, err)}
	}

	return nil
}

// validateLogFilePath validates that a log file path is valid and its directory is accessible
func validateLogFilePath(fieldName, path string) error {
	if path == "" {
		return &ValidationError{Field: fieldName, Message: "log file path cannot be empty"}
	}

	// Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("log file path contains invalid characters or traversal attempts: %s", path)}
	}

	// Check if path is absolute (recommended for security)
	if !filepath.IsAbs(cleanPath) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("log file path should be absolute for security: %s", path)}
	}

	// Get the directory path for the log file
	logDir := filepath.Dir(cleanPath)

	// Check if the directory exists and is accessible
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("log directory does not exist: %s", logDir)}
	}

	// Check if the directory is accessible (writable for log files)
	if _, err := os.Stat(logDir); err != nil {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("log directory is not accessible: %s - %v", logDir, err)}
	}

	return nil
}

// validateCodecConfig validates codec configuration.
func validateCodecConfig(config *CodecConfig) error {
	validProfiles := []string{"baseline", "main", "high"}
	if !contains(validProfiles, config.VideoProfile) {
		return &ValidationError{Field: "mediamtx.codec.video_profile", Message: fmt.Sprintf("video profile must be one of %v, got %s", validProfiles, config.VideoProfile)}
	}

	validLevels := []string{"1.0", "1.1", "1.2", "1.3", "2.0", "2.1", "2.2", "3.0", "3.1", "3.2", "4.0", "4.1", "4.2", "5.0", "5.1", "5.2"}
	if !contains(validLevels, config.VideoLevel) {
		return &ValidationError{Field: "mediamtx.codec.video_level", Message: fmt.Sprintf("video level must be one of %v, got %s", validLevels, config.VideoLevel)}
	}

	validFormats := []string{"yuv420p", "yuv422p", "yuv444p"}
	if !contains(validFormats, config.PixelFormat) {
		return &ValidationError{Field: "mediamtx.codec.pixel_format", Message: fmt.Sprintf("pixel format must be one of %v, got %s", validFormats, config.PixelFormat)}
	}

	if config.Bitrate == "" {
		return &ValidationError{Field: "mediamtx.codec.bitrate", Message: "bitrate cannot be empty"}
	}

	validPresets := []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}
	if !contains(validPresets, config.Preset) {
		return &ValidationError{Field: "mediamtx.codec.preset", Message: fmt.Sprintf("preset must be one of %v, got %s", validPresets, config.Preset)}
	}

	return nil
}

// validateStreamReadinessConfig validates stream readiness configuration.
func validateStreamReadinessConfig(config *StreamReadinessConfig) error {
	if config.Timeout <= 0 {
		return &ValidationError{Field: "mediamtx.stream_readiness.timeout", Message: fmt.Sprintf("timeout must be positive, got %f", config.Timeout)}
	}

	if config.RetryAttempts < 0 {
		return &ValidationError{Field: "mediamtx.stream_readiness.retry_attempts", Message: fmt.Sprintf("retry attempts must be non-negative, got %d", config.RetryAttempts)}
	}

	if config.RetryDelay < 0 {
		return &ValidationError{Field: "mediamtx.stream_readiness.retry_delay", Message: fmt.Sprintf("retry delay must be non-negative, got %f", config.RetryDelay)}
	}

	if config.CheckInterval <= 0 {
		return &ValidationError{Field: "mediamtx.stream_readiness.check_interval", Message: fmt.Sprintf("check interval must be positive, got %f", config.CheckInterval)}
	}

	return nil
}

// validateCameraConfig validates camera configuration.
func validateCameraConfig(config *CameraConfig) error {
	if config.PollInterval < 0 {
		return &ValidationError{Field: "camera.poll_interval", Message: fmt.Sprintf("poll interval must be non-negative, got %f", config.PollInterval)}
	}

	if config.DetectionTimeout <= 0 {
		return &ValidationError{Field: "camera.detection_timeout", Message: fmt.Sprintf("detection timeout must be positive, got %f", config.DetectionTimeout)}
	}

	if len(config.DeviceRange) != 2 {
		return &ValidationError{Field: "camera.device_range", Message: "device range must have exactly 2 elements"}
	}

	if config.DeviceRange[0] < 0 || config.DeviceRange[1] < 0 {
		return &ValidationError{Field: "camera.device_range", Message: "device range values must be non-negative"}
	}

	if config.DeviceRange[0] > config.DeviceRange[1] {
		return &ValidationError{Field: "camera.device_range", Message: "device range min must be less than or equal to max"}
	}

	if config.CapabilityTimeout <= 0 {
		return &ValidationError{Field: "camera.capability_timeout", Message: fmt.Sprintf("capability timeout must be positive, got %f", config.CapabilityTimeout)}
	}

	if config.CapabilityRetryInterval < 0 {
		return &ValidationError{Field: "camera.capability_retry_interval", Message: fmt.Sprintf("capability retry interval must be non-negative, got %f", config.CapabilityRetryInterval)}
	}

	if config.CapabilityMaxRetries < 0 {
		return &ValidationError{Field: "camera.capability_max_retries", Message: fmt.Sprintf("capability max retries must be non-negative, got %d", config.CapabilityMaxRetries)}
	}

	return nil
}

// validateLoggingConfig validates logging configuration.
func validateLoggingConfig(config *LoggingConfig) error {
	validLevels := []string{"DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL"}
	if !contains(validLevels, strings.ToUpper(config.Level)) {
		return &ValidationError{Field: "logging.level", Message: fmt.Sprintf("log level must be one of %v, got %s", validLevels, config.Level)}
	}

	if config.Format == "" {
		return &ValidationError{Field: "logging.format", Message: "log format cannot be empty"}
	}

	if config.FileEnabled && config.FilePath == "" {
		return &ValidationError{Field: "logging.file_path", Message: "file path cannot be empty when file logging is enabled"}
	}

	// Validate log file path when file logging is enabled
	if config.FileEnabled && config.FilePath != "" {
		if err := validateLogFilePath("logging.file_path", config.FilePath); err != nil {
			return err
		}
	}

	if config.MaxFileSize <= 0 {
		return &ValidationError{Field: "logging.max_file_size", Message: fmt.Sprintf("max file size must be positive, got %d", config.MaxFileSize)}
	}

	if config.BackupCount < 0 {
		return &ValidationError{Field: "logging.backup_count", Message: fmt.Sprintf("backup count must be non-negative, got %d", config.BackupCount)}
	}

	return nil
}

// validateRecordingConfig validates recording configuration.
func validateRecordingConfig(config *RecordingConfig) error {
	validFormats := []string{"fmp4", "mp4", "mkv", "avi"}
	if !contains(validFormats, config.Format) {
		return &ValidationError{Field: "recording.format", Message: fmt.Sprintf("recording format must be one of %v, got %s", validFormats, config.Format)}
	}

	validQualities := []string{"low", "medium", "high"}
	if !contains(validQualities, config.Quality) {
		return &ValidationError{Field: "recording.quality", Message: fmt.Sprintf("recording quality must be one of %v, got %s", validQualities, config.Quality)}
	}

	if config.SegmentDuration <= 0 {
		return &ValidationError{Field: "recording.segment_duration", Message: fmt.Sprintf("segment duration must be positive, got %d", config.SegmentDuration)}
	}

	if config.MaxSegmentSize <= 0 {
		return &ValidationError{Field: "recording.max_segment_size", Message: fmt.Sprintf("max segment size must be positive, got %d", config.MaxSegmentSize)}
	}

	if config.CleanupInterval <= 0 {
		return &ValidationError{Field: "recording.cleanup_interval", Message: fmt.Sprintf("cleanup interval must be positive, got %d", config.CleanupInterval)}
	}

	if config.MaxAge <= 0 {
		return &ValidationError{Field: "recording.max_age", Message: fmt.Sprintf("max age must be positive, got %d", config.MaxAge)}
	}

	if config.MaxSize <= 0 {
		return &ValidationError{Field: "recording.max_size", Message: fmt.Sprintf("max size must be positive, got %d", config.MaxSize)}
	}

	if config.DefaultRotationSize < 0 {
		return &ValidationError{Field: "recording.default_rotation_size", Message: fmt.Sprintf("recording default rotation size cannot be negative, got %d", config.DefaultRotationSize)}
	}

	return nil
}

// validateSnapshotConfig validates snapshot configuration.
func validateSnapshotConfig(config *SnapshotConfig) error {
	validFormats := []string{"jpeg", "png", "bmp"}
	if !contains(validFormats, config.Format) {
		return &ValidationError{Field: "snapshots.format", Message: fmt.Sprintf("snapshot format must be one of %v, got %s", validFormats, config.Format)}
	}

	if config.Quality < 1 || config.Quality > 100 {
		return &ValidationError{Field: "snapshots.quality", Message: fmt.Sprintf("quality must be between 1 and 100, got %d", config.Quality)}
	}

	if config.MaxWidth <= 0 {
		return &ValidationError{Field: "snapshots.max_width", Message: fmt.Sprintf("max width must be positive, got %d", config.MaxWidth)}
	}

	if config.MaxHeight <= 0 {
		return &ValidationError{Field: "snapshots.max_height", Message: fmt.Sprintf("max height must be positive, got %d", config.MaxHeight)}
	}

	if config.CleanupInterval <= 0 {
		return &ValidationError{Field: "snapshots.cleanup_interval", Message: fmt.Sprintf("cleanup interval must be positive, got %d", config.CleanupInterval)}
	}

	if config.MaxAge <= 0 {
		return &ValidationError{Field: "snapshots.max_age", Message: fmt.Sprintf("max age must be positive, got %d", config.MaxAge)}
	}

	if config.MaxCount <= 0 {
		return &ValidationError{Field: "snapshots.max_count", Message: fmt.Sprintf("max count must be positive, got %d", config.MaxCount)}
	}

	return nil
}

// validateFFmpegConfig validates FFmpeg configuration.
func validateFFmpegConfig(config *FFmpegConfig) error {
	if err := validateFFmpegSnapshotConfig(&config.Snapshot); err != nil {
		return fmt.Errorf("failed to validate FFmpeg snapshot config: %w", err)
	}

	if err := validateFFmpegRecordingConfig(&config.Recording); err != nil {
		return fmt.Errorf("failed to validate FFmpeg recording config: %w", err)
	}

	return nil
}

// validateFFmpegSnapshotConfig validates FFmpeg snapshot configuration.
func validateFFmpegSnapshotConfig(config *FFmpegSnapshotConfig) error {
	if config.ProcessCreationTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.snapshot.process_creation_timeout", Message: fmt.Sprintf("process creation timeout must be positive, got %f", config.ProcessCreationTimeout)}
	}

	if config.ExecutionTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.snapshot.execution_timeout", Message: fmt.Sprintf("execution timeout must be positive, got %f", config.ExecutionTimeout)}
	}

	if config.InternalTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.snapshot.internal_timeout", Message: fmt.Sprintf("internal timeout must be positive, got %d", config.InternalTimeout)}
	}

	if config.RetryAttempts < 0 {
		return &ValidationError{Field: "ffmpeg.snapshot.retry_attempts", Message: fmt.Sprintf("retry attempts must be non-negative, got %d", config.RetryAttempts)}
	}

	if config.RetryDelay < 0 {
		return &ValidationError{Field: "ffmpeg.snapshot.retry_delay", Message: fmt.Sprintf("retry delay must be non-negative, got %f", config.RetryDelay)}
	}

	return nil
}

// validateFFmpegRecordingConfig validates FFmpeg recording configuration.
func validateFFmpegRecordingConfig(config *FFmpegRecordingConfig) error {
	if config.ProcessCreationTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.recording.process_creation_timeout", Message: fmt.Sprintf("process creation timeout must be positive, got %f", config.ProcessCreationTimeout)}
	}

	if config.ExecutionTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.recording.execution_timeout", Message: fmt.Sprintf("execution timeout must be positive, got %f", config.ExecutionTimeout)}
	}

	if config.InternalTimeout <= 0 {
		return &ValidationError{Field: "ffmpeg.recording.internal_timeout", Message: fmt.Sprintf("internal timeout must be positive, got %d", config.InternalTimeout)}
	}

	if config.RetryAttempts < 0 {
		return &ValidationError{Field: "ffmpeg.recording.retry_attempts", Message: fmt.Sprintf("retry attempts must be non-negative, got %d", config.RetryAttempts)}
	}

	if config.RetryDelay < 0 {
		return &ValidationError{Field: "ffmpeg.recording.retry_delay", Message: fmt.Sprintf("retry delay must be non-negative, got %f", config.RetryDelay)}
	}

	return nil
}

// validateNotificationsConfig validates notifications configuration.
func validateNotificationsConfig(config *NotificationsConfig) error {
	if err := validateWebSocketNotificationConfig(&config.WebSocket); err != nil {
		return fmt.Errorf("failed to validate WebSocket notification config: %w", err)
	}

	if err := validateRealTimeNotificationConfig(&config.RealTime); err != nil {
		return fmt.Errorf("failed to validate real-time notification config: %w", err)
	}

	return nil
}

// validateWebSocketNotificationConfig validates WebSocket notification configuration.
func validateWebSocketNotificationConfig(config *WebSocketNotificationConfig) error {
	if config.DeliveryTimeout <= 0 {
		return &ValidationError{Field: "notifications.websocket.delivery_timeout", Message: fmt.Sprintf("delivery timeout must be positive, got %f", config.DeliveryTimeout)}
	}

	if config.RetryAttempts < 0 {
		return &ValidationError{Field: "notifications.websocket.retry_attempts", Message: fmt.Sprintf("retry attempts must be non-negative, got %d", config.RetryAttempts)}
	}

	if config.RetryDelay < 0 {
		return &ValidationError{Field: "notifications.websocket.retry_delay", Message: fmt.Sprintf("retry delay must be non-negative, got %f", config.RetryDelay)}
	}

	if config.MaxQueueSize <= 0 {
		return &ValidationError{Field: "notifications.websocket.max_queue_size", Message: fmt.Sprintf("max queue size must be positive, got %d", config.MaxQueueSize)}
	}

	if config.CleanupInterval <= 0 {
		return &ValidationError{Field: "notifications.websocket.cleanup_interval", Message: fmt.Sprintf("cleanup interval must be positive, got %d", config.CleanupInterval)}
	}

	return nil
}

// validateRealTimeNotificationConfig validates real-time notification configuration.
func validateRealTimeNotificationConfig(config *RealTimeNotificationConfig) error {
	if config.CameraStatusInterval <= 0 {
		return &ValidationError{Field: "notifications.real_time.camera_status_interval", Message: fmt.Sprintf("camera status interval must be positive, got %f", config.CameraStatusInterval)}
	}

	if config.RecordingProgressInterval <= 0 {
		return &ValidationError{Field: "notifications.real_time.recording_progress_interval", Message: fmt.Sprintf("recording progress interval must be positive, got %f", config.RecordingProgressInterval)}
	}

	if config.ConnectionHealthCheck <= 0 {
		return &ValidationError{Field: "notifications.real_time.connection_health_check", Message: fmt.Sprintf("connection health check must be positive, got %f", config.ConnectionHealthCheck)}
	}

	return nil
}

// validatePerformanceConfig validates performance configuration.
func validatePerformanceConfig(config *PerformanceConfig) error {
	if err := validateResponseTimeTargetsConfig(&config.ResponseTimeTargets); err != nil {
		return fmt.Errorf("failed to validate response time targets config: %w", err)
	}

	if err := validateSnapshotTiersConfig(&config.SnapshotTiers); err != nil {
		return fmt.Errorf("failed to validate snapshot tiers config: %w", err)
	}

	if err := validateOptimizationConfig(&config.Optimization); err != nil {
		return fmt.Errorf("failed to validate optimization config: %w", err)
	}

	return nil
}

// validateResponseTimeTargetsConfig validates response time targets configuration.
func validateResponseTimeTargetsConfig(config *ResponseTimeTargetsConfig) error {
	if config.SnapshotCapture <= 0 {
		return &ValidationError{Field: "performance.response_time_targets.snapshot_capture", Message: fmt.Sprintf("snapshot capture target must be positive, got %f", config.SnapshotCapture)}
	}

	if config.RecordingStart <= 0 {
		return &ValidationError{Field: "performance.response_time_targets.recording_start", Message: fmt.Sprintf("recording start target must be positive, got %f", config.RecordingStart)}
	}

	if config.RecordingStop <= 0 {
		return &ValidationError{Field: "performance.response_time_targets.recording_stop", Message: fmt.Sprintf("recording stop target must be positive, got %f", config.RecordingStop)}
	}

	if config.FileListing <= 0 {
		return &ValidationError{Field: "performance.response_time_targets.file_listing", Message: fmt.Sprintf("file listing target must be positive, got %f", config.FileListing)}
	}

	return nil
}

// validateSnapshotTiersConfig validates snapshot tiers configuration.
func validateSnapshotTiersConfig(config *SnapshotTiersConfig) error {
	if config.Tier1USBDirectTimeout <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.tier1_usb_direct_timeout", Message: fmt.Sprintf("tier1 USB direct timeout must be positive, got %f", config.Tier1USBDirectTimeout)}
	}

	if config.Tier2RTSPReadyCheckTimeout <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.tier2_rtsp_ready_check_timeout", Message: fmt.Sprintf("tier2 RTSP ready check timeout must be positive, got %f", config.Tier2RTSPReadyCheckTimeout)}
	}

	if config.Tier3ActivationTimeout <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.tier3_activation_timeout", Message: fmt.Sprintf("tier3 activation timeout must be positive, got %f", config.Tier3ActivationTimeout)}
	}

	if config.Tier3ActivationTriggerTimeout <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.tier3_activation_trigger_timeout", Message: fmt.Sprintf("tier3 activation trigger timeout must be positive, got %f", config.Tier3ActivationTriggerTimeout)}
	}

	if config.TotalOperationTimeout <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.total_operation_timeout", Message: fmt.Sprintf("total operation timeout must be positive, got %f", config.TotalOperationTimeout)}
	}

	if config.ImmediateResponseThreshold <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.immediate_response_threshold", Message: fmt.Sprintf("immediate response threshold must be positive, got %f", config.ImmediateResponseThreshold)}
	}

	if config.AcceptableResponseThreshold <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.acceptable_response_threshold", Message: fmt.Sprintf("acceptable response threshold must be positive, got %f", config.AcceptableResponseThreshold)}
	}

	if config.SlowResponseThreshold <= 0 {
		return &ValidationError{Field: "performance.snapshot_tiers.slow_response_threshold", Message: fmt.Sprintf("slow response threshold must be positive, got %f", config.SlowResponseThreshold)}
	}

	return nil
}

// validateOptimizationConfig validates optimization configuration.
func validateOptimizationConfig(config *OptimizationConfig) error {
	if config.CacheTTL <= 0 {
		return &ValidationError{Field: "performance.optimization.cache_ttl", Message: fmt.Sprintf("cache TTL must be positive, got %d", config.CacheTTL)}
	}

	if config.MaxConcurrentOperations <= 0 {
		return &ValidationError{Field: "performance.optimization.max_concurrent_operations", Message: fmt.Sprintf("max concurrent operations must be positive, got %d", config.MaxConcurrentOperations)}
	}

	if config.ConnectionPoolSize <= 0 {
		return &ValidationError{Field: "performance.optimization.connection_pool_size", Message: fmt.Sprintf("connection pool size must be positive, got %d", config.ConnectionPoolSize)}
	}

	return nil
}

// validateRetentionPolicyConfig validates retention policy configuration.
func validateRetentionPolicyConfig(config *RetentionPolicyConfig) error {
	// Validate policy type
	validTypes := []string{"age", "size", "manual"}
	if !contains(validTypes, config.Type) {
		return &ValidationError{Field: "retention_policy.type", Message: fmt.Sprintf("policy type must be one of %v, got %s", validTypes, config.Type)}
	}

	// Validate age-based policy parameters
	if config.Type == "age" {
		if config.MaxAgeDays < 0 {
			return &ValidationError{Field: "retention_policy.max_age_days", Message: fmt.Sprintf("max age days cannot be negative for age-based policy, got %d", config.MaxAgeDays)}
		}
		if config.MaxAgeDays > 0 && config.MaxAgeDays > 365 {
			return &ValidationError{Field: "retention_policy.max_age_days", Message: fmt.Sprintf("max age days cannot exceed 365 days, got %d", config.MaxAgeDays)}
		}
	}

	// Validate size-based policy parameters
	if config.Type == "size" {
		if config.MaxSizeGB <= 0 {
			return &ValidationError{Field: "retention_policy.max_size_gb", Message: fmt.Sprintf("max size GB must be positive for size-based policy, got %d", config.MaxSizeGB)}
		}
		if config.MaxSizeGB > 1000 {
			return &ValidationError{Field: "retention_policy.max_size_gb", Message: fmt.Sprintf("max size GB cannot exceed 1000 GB, got %d", config.MaxSizeGB)}
		}
	}

	return nil
}

// validateSecurityConfig validates security configuration.
func validateSecurityConfig(config *SecurityConfig) error {
	// Validate JWT secret key
	if strings.TrimSpace(config.JWTSecretKey) == "" {
		return &ValidationError{Field: "security.jwt_secret_key", Message: "JWT secret key cannot be empty"}
	}

	// Validate JWT expiry hours
	if config.JWTExpiryHours <= 0 {
		return &ValidationError{Field: "security.jwt_expiry_hours", Message: fmt.Sprintf("JWT expiry hours must be positive, got %d", config.JWTExpiryHours)}
	}

	// Validate rate limit requests
	if config.RateLimitRequests <= 0 {
		return &ValidationError{Field: "security.rate_limit_requests", Message: fmt.Sprintf("rate limit requests must be positive, got %d", config.RateLimitRequests)}
	}

	// Validate rate limit window
	if config.RateLimitWindow <= 0 {
		return &ValidationError{Field: "security.rate_limit_window", Message: fmt.Sprintf("rate limit window must be positive, got %v", config.RateLimitWindow)}
	}

	return nil
}

// validateServerDefaults validates server defaults configuration.
func validateServerDefaults(config *ServerDefaults) error {
	// Validate shutdown timeout
	if config.ShutdownTimeout < 0 {
		return &ValidationError{Field: "server_defaults.shutdown_timeout", Message: fmt.Sprintf("shutdown timeout cannot be negative, got %f", config.ShutdownTimeout)}
	}

	// Validate camera monitor ticker
	if config.CameraMonitorTicker < 0 {
		return &ValidationError{Field: "server_defaults.camera_monitor_ticker", Message: fmt.Sprintf("camera monitor ticker cannot be negative, got %f", config.CameraMonitorTicker)}
	}

	return nil
}

// validateStorageConfig validates storage configuration.
func validateStorageConfig(config *StorageConfig) error {
	// Validate warn and block percentages
	if config.WarnPercent < 0 || config.WarnPercent > 100 {
		return &ValidationError{Field: "storage.warn_percent", Message: fmt.Sprintf("warn percent must be between 0 and 100, got %d", config.WarnPercent)}
	}

	if config.BlockPercent < 0 || config.BlockPercent > 100 {
		return &ValidationError{Field: "storage.block_percent", Message: fmt.Sprintf("block percent must be between 0 and 100, got %d", config.BlockPercent)}
	}

	// Ensure warn percent is less than block percent
	if config.WarnPercent >= config.BlockPercent {
		return &ValidationError{Field: "storage.warn_percent", Message: fmt.Sprintf("warn percent (%d) must be less than block percent (%d)", config.WarnPercent, config.BlockPercent)}
	}

	// Validate storage paths
	if config.DefaultPath != "" {
		if err := validateStoragePath("storage.default_path", config.DefaultPath); err != nil {
			return err
		}
	}

	if config.FallbackPath != "" {
		if err := validateStoragePath("storage.fallback_path", config.FallbackPath); err != nil {
			return err
		}
	}

	return nil
}

// validateStoragePath validates that a storage path is valid, secure, and accessible
func validateStoragePath(fieldName, path string) error {
	if strings.TrimSpace(path) == "" {
		return &ValidationError{Field: fieldName, Message: "storage path cannot be empty"}
	}

	// Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("storage path contains invalid characters or traversal attempts: %s", path)}
	}

	// Check if path is absolute (recommended for security)
	if !filepath.IsAbs(cleanPath) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("storage path should be absolute for security: %s", path)}
	}

	// Check if path exists on filesystem
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("storage path does not exist: %s", path)}
	}

	// Check if path is accessible (readable and writable for storage)
	if _, err := os.Stat(cleanPath); err != nil {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("storage path is not accessible: %s - %v", path, err)}
	}

	return nil
}

// validateRTSPMonitoringConfig validates RTSP monitoring configuration
func validateRTSPMonitoringConfig(config *RTSPMonitoringConfig) error {
	if config.CheckInterval <= 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.check_interval", Message: fmt.Sprintf("check interval must be positive, got %d", config.CheckInterval)}
	}

	if config.ConnectionTimeout <= 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.connection_timeout", Message: fmt.Sprintf("connection timeout must be positive, got %d", config.ConnectionTimeout)}
	}

	if config.MaxConnections <= 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.max_connections", Message: fmt.Sprintf("max connections must be positive, got %d", config.MaxConnections)}
	}

	if config.SessionTimeout <= 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.session_timeout", Message: fmt.Sprintf("session timeout must be positive, got %d", config.SessionTimeout)}
	}

	if config.BandwidthThreshold <= 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.bandwidth_threshold", Message: fmt.Sprintf("bandwidth threshold must be positive, got %d", config.BandwidthThreshold)}
	}

	if config.PacketLossThreshold < 0 || config.PacketLossThreshold > 1 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.packet_loss_threshold", Message: fmt.Sprintf("packet loss threshold must be between 0 and 1, got %f", config.PacketLossThreshold)}
	}

	if config.JitterThreshold < 0 {
		return &ValidationError{Field: "mediamtx.rtsp_monitoring.jitter_threshold", Message: fmt.Sprintf("jitter threshold must be non-negative, got %f", config.JitterThreshold)}
	}

	return nil
}

// contains checks if a slice contains a specific value.
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
