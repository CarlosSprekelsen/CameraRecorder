package config

import (
	"fmt"
	"strings"
)

// validateConfig validates the entire configuration.
func (cl *ConfigLoader) validateConfig(config *Config) error {
	var errors []string
	
	// Validate server configuration
	if err := cl.validateServerConfig(config.Server); err != nil {
		errors = append(errors, fmt.Sprintf("server: %v", err))
	}
	
	// Validate MediaMTX configuration
	if err := cl.validateMediaMTXConfig(config.MediaMTX); err != nil {
		errors = append(errors, fmt.Sprintf("mediamtx: %v", err))
	}
	
	// Validate camera configuration
	if err := cl.validateCameraConfig(config.Camera); err != nil {
		errors = append(errors, fmt.Sprintf("camera: %v", err))
	}
	
	// Validate logging configuration
	if err := cl.validateLoggingConfig(config.Logging); err != nil {
		errors = append(errors, fmt.Sprintf("logging: %v", err))
	}
	
	// Validate recording configuration
	if err := cl.validateRecordingConfig(config.Recording); err != nil {
		errors = append(errors, fmt.Sprintf("recording: %v", err))
	}
	
	// Validate snapshot configuration
	if err := cl.validateSnapshotConfig(config.Snapshots); err != nil {
		errors = append(errors, fmt.Sprintf("snapshots: %v", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}
	
	return nil
}

// validateServerConfig validates server configuration.
func (cl *ConfigLoader) validateServerConfig(config ServerConfig) error {
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", config.Port)
	}
	
	if config.MaxConnections < 1 {
		return fmt.Errorf("max_connections must be positive, got %d", config.MaxConnections)
	}
	
	if config.WebSocketPath == "" {
		return fmt.Errorf("websocket_path cannot be empty")
	}
	
	if !strings.HasPrefix(config.WebSocketPath, "/") {
		return fmt.Errorf("websocket_path must start with '/', got %s", config.WebSocketPath)
	}
	
	return nil
}

// validateMediaMTXConfig validates MediaMTX configuration.
func (cl *ConfigLoader) validateMediaMTXConfig(config MediaMTXConfig) error {
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
			return fmt.Errorf("%s must be between 1 and 65535, got %d", p.name, p.port)
		}
	}
	
	// Validate STANAG 4406 codec settings
	validProfiles := []string{"baseline", "main", "high"}
	if !contains(validProfiles, config.VideoProfile) {
		return fmt.Errorf("video_profile must be one of %v, got %s", validProfiles, config.VideoProfile)
	}
	
	validLevels := []string{"1.0", "1.1", "1.2", "1.3", "2.0", "2.1", "2.2", "3.0", "3.1", "3.2", "4.0", "4.1", "4.2", "5.0", "5.1", "5.2"}
	if !contains(validLevels, config.VideoLevel) {
		return fmt.Errorf("video_level must be one of %v, got %s", validLevels, config.VideoLevel)
	}
	
	validPixelFormats := []string{"yuv420p", "yuv422p", "yuv444p"}
	if !contains(validPixelFormats, config.PixelFormat) {
		return fmt.Errorf("pixel_format must be one of %v, got %s", validPixelFormats, config.PixelFormat)
	}
	
	validPresets := []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}
	if !contains(validPresets, config.Preset) {
		return fmt.Errorf("preset must be one of %v, got %s", validPresets, config.Preset)
	}
	
	// Validate health monitoring settings
	if config.HealthCheckInterval < 1 {
		return fmt.Errorf("health_check_interval must be at least 1 second, got %d", config.HealthCheckInterval)
	}
	
	if config.HealthFailureThreshold < 1 {
		return fmt.Errorf("health_failure_threshold must be at least 1, got %d", config.HealthFailureThreshold)
	}
	
	if config.HealthCircuitBreakerTimeout < 1 {
		return fmt.Errorf("health_circuit_breaker_timeout must be at least 1 second, got %d", config.HealthCircuitBreakerTimeout)
	}
	
	if config.HealthMaxBackoffInterval < 1 {
		return fmt.Errorf("health_max_backoff_interval must be at least 1 second, got %d", config.HealthMaxBackoffInterval)
	}
	
	if config.HealthRecoveryConfirmationThreshold < 1 {
		return fmt.Errorf("health_recovery_confirmation_threshold must be at least 1, got %d", config.HealthRecoveryConfirmationThreshold)
	}
	
	if config.BackoffBaseMultiplier <= 0 {
		return fmt.Errorf("backoff_base_multiplier must be positive, got %f", config.BackoffBaseMultiplier)
	}
	
	if len(config.BackoffJitterRange) != 2 {
		return fmt.Errorf("backoff_jitter_range must have exactly 2 values, got %d", len(config.BackoffJitterRange))
	}
	
	if config.BackoffJitterRange[0] >= config.BackoffJitterRange[1] {
		return fmt.Errorf("backoff_jitter_range first value must be less than second value, got [%f, %f]", 
			config.BackoffJitterRange[0], config.BackoffJitterRange[1])
	}
	
	if config.ProcessTerminationTimeout < 0 {
		return fmt.Errorf("process_termination_timeout must be non-negative, got %f", config.ProcessTerminationTimeout)
	}
	
	if config.ProcessKillTimeout < 0 {
		return fmt.Errorf("process_kill_timeout must be non-negative, got %f", config.ProcessKillTimeout)
	}
	
	// Validate stream readiness settings
	if err := cl.validateStreamReadinessConfig(config.StreamReadiness); err != nil {
		return fmt.Errorf("stream_readiness: %v", err)
	}
	
	return nil
}

// validateStreamReadinessConfig validates stream readiness configuration.
func (cl *ConfigLoader) validateStreamReadinessConfig(config StreamReadinessConfig) error {
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %f", config.Timeout)
	}
	
	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative, got %d", config.RetryAttempts)
	}
	
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative, got %f", config.RetryDelay)
	}
	
	if config.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive, got %f", config.CheckInterval)
	}
	
	return nil
}

// validateCameraConfig validates camera configuration.
func (cl *ConfigLoader) validateCameraConfig(config CameraConfig) error {
	if config.PollInterval < 0.01 {
		return fmt.Errorf("poll_interval must be at least 0.01 seconds, got %f", config.PollInterval)
	}
	
	if config.DetectionTimeout < 0.1 {
		return fmt.Errorf("detection_timeout must be at least 0.1 seconds, got %f", config.DetectionTimeout)
	}
	
	if len(config.DeviceRange) == 0 {
		return fmt.Errorf("device_range cannot be empty")
	}
	
	for i, device := range config.DeviceRange {
		if device < 0 {
			return fmt.Errorf("device_range[%d] must be non-negative, got %d", i, device)
		}
	}
	
	if config.CapabilityTimeout < 0 {
		return fmt.Errorf("capability_timeout must be non-negative, got %f", config.CapabilityTimeout)
	}
	
	if config.CapabilityRetryInterval < 0 {
		return fmt.Errorf("capability_retry_interval must be non-negative, got %f", config.CapabilityRetryInterval)
	}
	
	if config.CapabilityMaxRetries < 0 {
		return fmt.Errorf("capability_max_retries must be non-negative, got %d", config.CapabilityMaxRetries)
	}
	
	return nil
}

// validateLoggingConfig validates logging configuration.
func (cl *ConfigLoader) validateLoggingConfig(config LoggingConfig) error {
	validLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	if !contains(validLevels, config.Level) {
		return fmt.Errorf("invalid logging level: %s, must be one of %v", config.Level, validLevels)
	}
	
	if config.Format == "" {
		return fmt.Errorf("format cannot be empty")
	}
	
	if config.FileEnabled {
		if config.FilePath == "" {
			return fmt.Errorf("file_path cannot be empty when file logging is enabled")
		}
		
		if config.MaxFileSize < 1 {
			return fmt.Errorf("max_file_size must be at least 1 byte, got %d", config.MaxFileSize)
		}
		
		if config.BackupCount < 0 {
			return fmt.Errorf("backup_count must be non-negative, got %d", config.BackupCount)
		}
	}
	
	return nil
}

// validateRecordingConfig validates recording configuration.
func (cl *ConfigLoader) validateRecordingConfig(config RecordingConfig) error {
	validFormats := []string{"mp4", "fmp4", "mkv", "avi"}
	if !contains(validFormats, config.Format) {
		return fmt.Errorf("invalid recording format: %s, must be one of %v", config.Format, validFormats)
	}
	
	validQualities := []string{"low", "medium", "high"}
	if !contains(validQualities, config.Quality) {
		return fmt.Errorf("invalid recording quality: %s, must be one of %v", config.Quality, validQualities)
	}
	
	if config.SegmentDuration < 1 {
		return fmt.Errorf("segment_duration must be at least 1 second, got %d", config.SegmentDuration)
	}
	
	if config.MaxSegmentSize < 1 {
		return fmt.Errorf("max_segment_size must be at least 1 byte, got %d", config.MaxSegmentSize)
	}
	
	if config.CleanupInterval < 1 {
		return fmt.Errorf("cleanup_interval must be at least 1 second, got %d", config.CleanupInterval)
	}
	
	if config.MaxAge < 1 {
		return fmt.Errorf("max_age must be at least 1 second, got %d", config.MaxAge)
	}
	
	if config.MaxSize < 1 {
		return fmt.Errorf("max_size must be at least 1 byte, got %d", config.MaxSize)
	}
	
	if config.MaxDuration < 1 {
		return fmt.Errorf("max_duration must be at least 1 second, got %d", config.MaxDuration)
	}
	
	if config.CleanupAfterDays < 0 {
		return fmt.Errorf("cleanup_after_days must be non-negative, got %d", config.CleanupAfterDays)
	}
	
	if config.RotationMinutes < 1 {
		return fmt.Errorf("rotation_minutes must be at least 1 minute, got %d", config.RotationMinutes)
	}
	
	if config.StorageWarnPercent < 0 || config.StorageWarnPercent > 100 {
		return fmt.Errorf("storage_warn_percent must be between 0 and 100, got %d", config.StorageWarnPercent)
	}
	
	if config.StorageBlockPercent < 0 || config.StorageBlockPercent > 100 {
		return fmt.Errorf("storage_block_percent must be between 0 and 100, got %d", config.StorageBlockPercent)
	}
	
	if config.StorageWarnPercent >= config.StorageBlockPercent {
		return fmt.Errorf("storage_warn_percent (%d) must be less than storage_block_percent (%d)", 
			config.StorageWarnPercent, config.StorageBlockPercent)
	}
	
	return nil
}

// validateSnapshotConfig validates snapshot configuration.
func (cl *ConfigLoader) validateSnapshotConfig(config SnapshotConfig) error {
	validFormats := []string{"jpg", "jpeg", "png", "bmp"}
	if !contains(validFormats, config.Format) {
		return fmt.Errorf("invalid snapshot format: %s, must be one of %v", config.Format, validFormats)
	}
	
	if config.Quality < 1 || config.Quality > 100 {
		return fmt.Errorf("snapshot quality must be between 1 and 100, got %d", config.Quality)
	}
	
	if config.MaxWidth < 1 {
		return fmt.Errorf("max_width must be at least 1 pixel, got %d", config.MaxWidth)
	}
	
	if config.MaxHeight < 1 {
		return fmt.Errorf("max_height must be at least 1 pixel, got %d", config.MaxHeight)
	}
	
	if config.CleanupInterval < 1 {
		return fmt.Errorf("cleanup_interval must be at least 1 second, got %d", config.CleanupInterval)
	}
	
	if config.MaxAge < 1 {
		return fmt.Errorf("max_age must be at least 1 second, got %d", config.MaxAge)
	}
	
	if config.MaxCount < 1 {
		return fmt.Errorf("max_count must be at least 1, got %d", config.MaxCount)
	}
	
	if config.CleanupAfterDays < 0 {
		return fmt.Errorf("cleanup_after_days must be non-negative, got %d", config.CleanupAfterDays)
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
