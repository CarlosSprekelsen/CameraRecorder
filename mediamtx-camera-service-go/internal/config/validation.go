package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

// validateConfig validates the complete configuration.
func validateConfig(config *Config) error {
	if err := validateServerConfig(&config.Server); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := validateMediaMTXConfig(&config.MediaMTX); err != nil {
		return fmt.Errorf("mediamtx config: %w", err)
	}
	if err := validateCameraConfig(&config.Camera); err != nil {
		return fmt.Errorf("camera config: %w", err)
	}
	if err := validateLoggingConfig(&config.Logging); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}
	if err := validateRecordingConfig(&config.Recording); err != nil {
		return fmt.Errorf("recording config: %w", err)
	}
	if err := validateSnapshotConfig(&config.Snapshots); err != nil {
		return fmt.Errorf("snapshots config: %w", err)
	}
	if err := validateFFmpegConfig(&config.FFmpeg); err != nil {
		return fmt.Errorf("ffmpeg config: %w", err)
	}
	if err := validatePerformanceConfig(&config.Performance); err != nil {
		return fmt.Errorf("performance config: %w", err)
	}

	// Cross-field validation
	if err := validateCrossFieldConstraints(config); err != nil {
		return fmt.Errorf("cross-field validation: %w", err)
	}

	return nil
}

// validateServerConfig validates server configuration.
func validateServerConfig(config *ServerConfig) error {
	if strings.TrimSpace(config.Host) == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	// Enhanced host validation
	if err := validateHost(config.Host); err != nil {
		return fmt.Errorf("invalid server host format: %w", err)
	}

	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if strings.TrimSpace(config.WebSocketPath) == "" {
		return fmt.Errorf("websocket path cannot be empty")
	}

	if config.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	return nil
}

// validateHost validates host format (hostname or IP address)
func validateHost(host string) error {
	// Check if it's a valid IP address
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Check if it's a valid hostname
	if len(host) > 253 {
		return fmt.Errorf("hostname too long")
	}

	// Basic hostname validation
	for _, part := range strings.Split(host, ".") {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("invalid hostname part")
		}
		// Check for valid characters (letters, digits, hyphens, but not starting/ending with hyphen)
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return fmt.Errorf("hostname part cannot start or end with hyphen")
		}
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
				(char >= '0' && char <= '9') || char == '-') {
				return fmt.Errorf("invalid character in hostname")
			}
		}
	}

	return nil
}

// validateMediaMTXConfig validates MediaMTX configuration.
func validateMediaMTXConfig(config *MediaMTXConfig) error {
	// Validate ports
	ports := []struct {
		name string
		port int
	}{
		{"api_port", config.APIPort},
		{"rtsp_port", config.RTSPPort},
		{"webrtc_port", config.WebRTCPort},
		{"hls_port", config.HLSPort},
	}
	
	for _, p := range ports {
		if p.port < 1 || p.port > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", p.name)
		}
	}

	// Validate STANAG 4406 codec settings
	validProfiles := []string{"baseline", "main", "high"}
	if !contains(validProfiles, config.Codec.VideoProfile) {
		return fmt.Errorf("video_profile must be one of %v", validProfiles)
	}
	
	validLevels := []string{"1.0", "1.1", "1.2", "1.3", "2.0", "2.1", "2.2", "3.0", "3.1", "3.2", "4.0", "4.1", "4.2", "5.0", "5.1", "5.2"}
	if !contains(validLevels, config.Codec.VideoLevel) {
		return fmt.Errorf("video_level must be one of %v", validLevels)
	}
	
	validPixelFormats := []string{"yuv420p", "yuv422p", "yuv444p"}
	if !contains(validPixelFormats, config.Codec.PixelFormat) {
		return fmt.Errorf("pixel_format must be one of %v", validPixelFormats)
	}
	
	validPresets := []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}
	if !contains(validPresets, config.Codec.Preset) {
		return fmt.Errorf("preset must be one of %v", validPresets)
	}

	// Validate health monitoring settings
	if config.HealthCheckInterval < 1 {
		return fmt.Errorf("health_check_interval must be at least 1 second")
	}
	
	if config.HealthFailureThreshold < 1 {
		return fmt.Errorf("health_failure_threshold must be at least 1")
	}
	
	if config.HealthCircuitBreakerTimeout < 1 {
		return fmt.Errorf("health_circuit_breaker_timeout must be at least 1 second")
	}
	
	if config.HealthMaxBackoffInterval < 1 {
		return fmt.Errorf("health_max_backoff_interval must be at least 1 second")
	}
	
	if config.HealthRecoveryConfirmationThreshold < 1 {
		return fmt.Errorf("health_recovery_confirmation_threshold must be at least 1")
	}
	
	if config.BackoffBaseMultiplier <= 0 {
		return fmt.Errorf("backoff_base_multiplier must be positive")
	}
	
	if len(config.BackoffJitterRange) != 2 {
		return fmt.Errorf("backoff_jitter_range must have exactly 2 values")
	}
	
	if config.BackoffJitterRange[0] >= config.BackoffJitterRange[1] {
		return fmt.Errorf("backoff_jitter_range first value must be less than second value")
	}
	
	if config.ProcessTerminationTimeout < 0 {
		return fmt.Errorf("process_termination_timeout must be non-negative")
	}
	
	if config.ProcessKillTimeout < 0 {
		return fmt.Errorf("process_kill_timeout must be non-negative")
	}

	// Validate stream readiness settings
	if err := validateStreamReadinessConfig(&config.StreamReadiness); err != nil {
		return fmt.Errorf("stream_readiness: %w", err)
	}

	return nil
}

// validateStreamReadinessConfig validates stream readiness configuration.
func validateStreamReadinessConfig(config *StreamReadinessConfig) error {
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}
	
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative")
	}
	
	if config.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	
	return nil
}

// validateCameraConfig validates camera configuration.
func validateCameraConfig(config *CameraConfig) error {
	if config.PollInterval < 0.01 {
		return fmt.Errorf("poll_interval must be at least 0.01 seconds")
	}
	
	if config.DetectionTimeout < 0.1 {
		return fmt.Errorf("detection_timeout must be at least 0.1 seconds")
	}
	
	if len(config.DeviceRange) == 0 {
		return fmt.Errorf("device_range cannot be empty")
	}
	
	for i, device := range config.DeviceRange {
		if device < 0 {
			return fmt.Errorf("device_range[%d] must be non-negative", i)
		}
	}
	
	if config.CapabilityTimeout < 0 {
		return fmt.Errorf("capability_timeout must be non-negative")
	}
	
	if config.CapabilityRetryInterval < 0 {
		return fmt.Errorf("capability_retry_interval must be non-negative")
	}
	
	if config.CapabilityMaxRetries < 0 {
		return fmt.Errorf("capability_max_retries must be non-negative")
	}

	return nil
}

// validateLoggingConfig validates logging configuration.
func validateLoggingConfig(config *LoggingConfig) error {
	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
	}
	if !validLevels[strings.ToLower(config.Level)] {
		return fmt.Errorf("invalid log level: %s", config.Level)
	}

	if config.Format == "" {
		return fmt.Errorf("format cannot be empty")
	}

	if config.FileEnabled {
		if strings.TrimSpace(config.FilePath) == "" {
			return fmt.Errorf("file_path cannot be empty when file logging is enabled")
		}
		
		if config.MaxFileSize < 1 {
			return fmt.Errorf("max_file_size must be at least 1 byte")
		}
		
		if config.BackupCount < 0 {
			return fmt.Errorf("backup_count must be non-negative")
		}
	}

	return nil
}

// validateRecordingConfig validates recording configuration.
func validateRecordingConfig(config *RecordingConfig) error {
	validFormats := []string{"mp4", "fmp4", "mkv", "avi"}
	if !contains(validFormats, config.Format) {
		return fmt.Errorf("invalid recording format: %s, must be one of %v", config.Format, validFormats)
	}
	
	validQualities := []string{"low", "medium", "high"}
	if !contains(validQualities, config.Quality) {
		return fmt.Errorf("invalid recording quality: %s, must be one of %v", config.Quality, validQualities)
	}
	
	if config.SegmentDuration < 1 {
		return fmt.Errorf("segment_duration must be at least 1 second")
	}
	
	if config.MaxSegmentSize < 1 {
		return fmt.Errorf("max_segment_size must be at least 1 byte")
	}
	
	if config.CleanupInterval < 1 {
		return fmt.Errorf("cleanup_interval must be at least 1 second")
	}
	
	if config.MaxAge < 1 {
		return fmt.Errorf("max_age must be at least 1 second")
	}
	
	if config.MaxSize < 1 {
		return fmt.Errorf("max_size must be at least 1 byte")
	}
	


	return nil
}

// validateSnapshotConfig validates snapshot configuration.
func validateSnapshotConfig(config *SnapshotConfig) error {
	validFormats := []string{"jpg", "jpeg", "png", "bmp"}
	if !contains(validFormats, config.Format) {
		return fmt.Errorf("invalid snapshot format: %s, must be one of %v", config.Format, validFormats)
	}
	
	if config.Quality < 1 || config.Quality > 100 {
		return fmt.Errorf("snapshot quality must be between 1 and 100")
	}
	
	if config.MaxWidth < 1 {
		return fmt.Errorf("max_width must be at least 1 pixel")
	}
	
	if config.MaxHeight < 1 {
		return fmt.Errorf("max_height must be at least 1 pixel")
	}
	
	if config.CleanupInterval < 1 {
		return fmt.Errorf("cleanup_interval must be at least 1 second")
	}
	
	if config.MaxAge < 1 {
		return fmt.Errorf("max_age must be at least 1 second")
	}
	
	if config.MaxCount < 1 {
		return fmt.Errorf("max_count must be at least 1")
	}
	


	return nil
}

// validateFFmpegConfig validates FFmpeg configuration.
func validateFFmpegConfig(config *FFmpegConfig) error {
	// Validate snapshot operation config
	if err := validateFFmpegOperationConfig(&config.Snapshot); err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	
	// Validate recording operation config
	if err := validateFFmpegOperationConfig(&config.Recording); err != nil {
		return fmt.Errorf("recording: %w", err)
	}

	return nil
}

// validateFFmpegOperationConfig validates FFmpeg operation configuration.
func validateFFmpegOperationConfig(config *FFmpegOperationConfig) error {
	if config.ProcessCreationTimeout <= 0 {
		return fmt.Errorf("process_creation_timeout must be positive")
	}
	
	if config.ExecutionTimeout <= 0 {
		return fmt.Errorf("execution_timeout must be positive")
	}
	
	if config.InternalTimeout <= 0 {
		return fmt.Errorf("internal_timeout must be positive")
	}
	
	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}
	
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative")
	}

	return nil
}

// validatePerformanceConfig validates performance configuration.
func validatePerformanceConfig(config *PerformanceConfig) error {
	// Validate response time targets
	if err := validateResponseTimeTargets(&config.ResponseTimeTargets); err != nil {
		return fmt.Errorf("response_time_targets: %w", err)
	}
	
	// Validate snapshot tiers
	if err := validateSnapshotTiers(&config.SnapshotTiers); err != nil {
		return fmt.Errorf("snapshot_tiers: %w", err)
	}
	
	// Validate optimization config
	if err := validateOptimizationConfig(&config.Optimization); err != nil {
		return fmt.Errorf("optimization: %w", err)
	}

	return nil
}

// validateResponseTimeTargets validates response time target configuration.
func validateResponseTimeTargets(config *ResponseTimeTargets) error {
	if config.SnapshotCapture <= 0 {
		return fmt.Errorf("snapshot_capture must be positive")
	}
	
	if config.RecordingStart <= 0 {
		return fmt.Errorf("recording_start must be positive")
	}
	
	if config.RecordingStop <= 0 {
		return fmt.Errorf("recording_stop must be positive")
	}
	
	if config.FileListing <= 0 {
		return fmt.Errorf("file_listing must be positive")
	}

	return nil
}

// validateSnapshotTiers validates snapshot tier configuration.
func validateSnapshotTiers(config *SnapshotTiers) error {
	if config.Tier1USBDirectTimeout <= 0 {
		return fmt.Errorf("tier1_usb_direct_timeout must be positive")
	}
	
	if config.Tier2RTSPReadyCheckTimeout <= 0 {
		return fmt.Errorf("tier2_rtsp_ready_check_timeout must be positive")
	}
	
	if config.Tier3ActivationTimeout <= 0 {
		return fmt.Errorf("tier3_activation_timeout must be positive")
	}
	
	if config.Tier3ActivationTriggerTimeout <= 0 {
		return fmt.Errorf("tier3_activation_trigger_timeout must be positive")
	}
	
	if config.TotalOperationTimeout <= 0 {
		return fmt.Errorf("total_operation_timeout must be positive")
	}
	
	if config.ImmediateResponseThreshold <= 0 {
		return fmt.Errorf("immediate_response_threshold must be positive")
	}
	
	if config.AcceptableResponseThreshold <= 0 {
		return fmt.Errorf("acceptable_response_threshold must be positive")
	}
	
	if config.SlowResponseThreshold <= 0 {
		return fmt.Errorf("slow_response_threshold must be positive")
	}

	return nil
}

// validateOptimizationConfig validates optimization configuration.
func validateOptimizationConfig(config *OptimizationConfig) error {
	if config.EnableCaching {
		if config.CacheTTL <= 0 {
			return fmt.Errorf("cache_ttl must be positive when caching is enabled")
		}
	}
	
	if config.MaxConcurrentOperations <= 0 {
		return fmt.Errorf("max_concurrent_operations must be positive")
	}
	
	if config.ConnectionPoolSize <= 0 {
		return fmt.Errorf("connection_pool_size must be positive")
	}

	return nil
}

// validatePath validates file system path
func validatePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check for absolute path (optional security measure)
	if filepath.IsAbs(path) {
		// Ensure it's not trying to access system directories
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, "/etc") || 
		   strings.HasPrefix(cleanPath, "/sys") || 
		   strings.HasPrefix(cleanPath, "/proc") {
			return fmt.Errorf("access to system directories not allowed")
		}
	}

	return nil
}

// validateCrossFieldConstraints validates relationships between different config sections
func validateCrossFieldConstraints(config *Config) error {
	// Check for port conflicts
	if config.Server.Port == config.MediaMTX.APIPort {
		return fmt.Errorf("server port conflicts with MediaMTX API port")
	}

	// Check for reasonable performance settings
	if config.Performance.Optimization.MaxConcurrentOperations > 100 {
		return fmt.Errorf("max concurrent operations too high for system stability")
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
