package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidationEdgeCases_MediaMTX tests edge cases for MediaMTX validation
func TestValidationEdgeCases_MediaMTX(t *testing.T) {
	tests := []struct {
		name        string
		config      *MediaMTXConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid MediaMTX config",
			config: &MediaMTXConfig{
				Host:    "127.0.0.1",
				APIPort: 9997,
				RTSPPort: 8554,
				WebRTCPort: 8889,
				HLSPort: 8888,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
				HealthCheckInterval: 30,
				HealthFailureThreshold: 10,
				HealthCircuitBreakerTimeout: 60,
				HealthMaxBackoffInterval: 120,
				HealthRecoveryConfirmationThreshold: 3,
				BackoffBaseMultiplier: 2.0,
				BackoffJitterRange: []float64{0.8, 1.2},
				ProcessTerminationTimeout: 3.0,
				ProcessKillTimeout: 2.0,
				StreamReadiness: StreamReadinessConfig{
					Timeout: 15.0,
					RetryAttempts: 3,
					RetryDelay: 2.0,
					CheckInterval: 0.5,
					EnableProgressNotifications: true,
					GracefulFallback: true,
				},
			},
			expectValid: true,
		},
		{
			name: "invalid API port",
			config: &MediaMTXConfig{
				APIPort: 0,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "api_port must be between 1 and 65535",
		},
		{
			name: "invalid video profile",
			config: &MediaMTXConfig{
				APIPort: 9997,
				Codec: CodecConfig{
					VideoProfile: "invalid_profile",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "video_profile must be one of",
		},
		{
			name: "invalid video level",
			config: &MediaMTXConfig{
				APIPort: 9997,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "invalid_level",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "video_level must be one of",
		},
		{
			name: "invalid pixel format",
			config: &MediaMTXConfig{
				APIPort: 9997,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "invalid_format",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "pixel_format must be one of",
		},
		{
			name: "invalid preset",
			config: &MediaMTXConfig{
				APIPort: 9997,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "invalid_preset",
				},
			},
			expectValid: false,
			errorMsg:    "preset must be one of",
		},
		{
			name: "invalid health check interval",
			config: &MediaMTXConfig{
				APIPort: 9997,
				HealthCheckInterval: 0,
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "health_check_interval must be at least 1 second",
		},
		{
			name: "invalid backoff jitter range",
			config: &MediaMTXConfig{
				APIPort: 9997,
				BackoffJitterRange: []float64{1.2, 0.8}, // Wrong order
				Codec: CodecConfig{
					VideoProfile: "baseline",
					VideoLevel:   "3.0",
					PixelFormat:  "yuv420p",
					Bitrate:      "600k",
					Preset:       "ultrafast",
				},
			},
			expectValid: false,
			errorMsg:    "backoff_jitter_range first value must be less than second value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMediaMTXConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_StreamReadiness tests edge cases for stream readiness validation
func TestValidationEdgeCases_StreamReadiness(t *testing.T) {
	tests := []struct {
		name        string
		config      *StreamReadinessConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid stream readiness config",
			config: &StreamReadinessConfig{
				Timeout: 15.0,
				RetryAttempts: 3,
				RetryDelay: 2.0,
				CheckInterval: 0.5,
				EnableProgressNotifications: true,
				GracefulFallback: true,
			},
			expectValid: true,
		},
		{
			name: "zero timeout",
			config: &StreamReadinessConfig{
				Timeout: 0.0,
				RetryAttempts: 3,
				RetryDelay: 2.0,
				CheckInterval: 0.5,
			},
			expectValid: false,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "negative retry attempts",
			config: &StreamReadinessConfig{
				Timeout: 15.0,
				RetryAttempts: -1,
				RetryDelay: 2.0,
				CheckInterval: 0.5,
			},
			expectValid: false,
			errorMsg:    "retry_attempts must be non-negative",
		},
		{
			name: "negative retry delay",
			config: &StreamReadinessConfig{
				Timeout: 15.0,
				RetryAttempts: 3,
				RetryDelay: -1.0,
				CheckInterval: 0.5,
			},
			expectValid: false,
			errorMsg:    "retry_delay must be non-negative",
		},
		{
			name: "zero check interval",
			config: &StreamReadinessConfig{
				Timeout: 15.0,
				RetryAttempts: 3,
				RetryDelay: 2.0,
				CheckInterval: 0.0,
			},
			expectValid: false,
			errorMsg:    "check_interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStreamReadinessConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_Camera tests edge cases for camera validation
func TestValidationEdgeCases_Camera(t *testing.T) {
	tests := []struct {
		name        string
		config      *CameraConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid camera config",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, 9},
				EnableCapabilityDetection: true,
				AutoStartStreams: true,
				CapabilityTimeout: 5.0,
				CapabilityRetryInterval: 1.0,
				CapabilityMaxRetries: 3,
			},
			expectValid: true,
		},
		{
			name: "poll interval too small",
			config: &CameraConfig{
				PollInterval: 0.005, // Less than 0.01
				DeviceRange: []int{0, 9},
			},
			expectValid: false,
			errorMsg:    "poll_interval must be at least 0.01 seconds",
		},
		{
			name: "detection timeout too small",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 0.05, // Less than 0.1
				DeviceRange: []int{0, 9},
			},
			expectValid: false,
			errorMsg:    "detection_timeout must be at least 0.1 seconds",
		},
		{
			name: "empty device range",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{},
			},
			expectValid: false,
			errorMsg:    "device_range cannot be empty",
		},
		{
			name: "negative device in range",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, -1, 2},
			},
			expectValid: false,
			errorMsg:    "device_range[1] must be non-negative",
		},
		{
			name: "negative capability timeout",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, 9},
				CapabilityTimeout: -1.0,
			},
			expectValid: false,
			errorMsg:    "capability_timeout must be non-negative",
		},
		{
			name: "negative capability retry interval",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, 9},
				CapabilityRetryInterval: -1.0,
			},
			expectValid: false,
			errorMsg:    "capability_retry_interval must be non-negative",
		},
		{
			name: "negative capability max retries",
			config: &CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, 9},
				CapabilityMaxRetries: -1,
			},
			expectValid: false,
			errorMsg:    "capability_max_retries must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCameraConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_Logging tests edge cases for logging validation
func TestValidationEdgeCases_Logging(t *testing.T) {
	tests := []struct {
		name        string
		config      *LoggingConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid logging config",
			config: &LoggingConfig{
				Level: "INFO",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
				FileEnabled: true,
				FilePath: "/opt/camera-service/logs/camera-service.log",
				MaxFileSize: 10485760,
				BackupCount: 5,
				ConsoleEnabled: true,
			},
			expectValid: true,
		},
		{
			name: "invalid log level",
			config: &LoggingConfig{
				Level: "INVALID_LEVEL",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
			},
			expectValid: false,
			errorMsg:    "invalid log level",
		},
		{
			name: "empty format",
			config: &LoggingConfig{
				Level: "INFO",
				Format: "",
			},
			expectValid: false,
			errorMsg:    "format cannot be empty",
		},
		{
			name: "file enabled but no path",
			config: &LoggingConfig{
				Level: "INFO",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
				FileEnabled: true,
				FilePath: "",
			},
			expectValid: false,
			errorMsg:    "file_path cannot be empty when file logging is enabled",
		},
		{
			name: "zero max file size",
			config: &LoggingConfig{
				Level: "INFO",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
				FileEnabled: true,
				FilePath: "/tmp/test.log",
				MaxFileSize: 0,
			},
			expectValid: false,
			errorMsg:    "max_file_size must be at least 1 byte",
		},
		{
			name: "negative backup count",
			config: &LoggingConfig{
				Level: "INFO",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
				FileEnabled: true,
				FilePath: "/tmp/test.log",
				MaxFileSize: 10485760,
				BackupCount: -1,
			},
			expectValid: false,
			errorMsg:    "backup_count must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoggingConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_Recording tests edge cases for recording validation
func TestValidationEdgeCases_Recording(t *testing.T) {
	tests := []struct {
		name        string
		config      *RecordingConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid recording config",
			config: &RecordingConfig{
				Enabled: false,
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 3600,
				MaxSegmentSize: 524288000,
				AutoCleanup: true,
				CleanupInterval: 86400,
				MaxAge: 604800,
				MaxSize: 10737418240,
			},
			expectValid: true,
		},
		{
			name: "invalid recording format",
			config: &RecordingConfig{
				Format: "invalid_format",
				Quality: "high",
			},
			expectValid: false,
			errorMsg:    "invalid recording format",
		},
		{
			name: "invalid recording quality",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "invalid_quality",
			},
			expectValid: false,
			errorMsg:    "invalid recording quality",
		},
		{
			name: "zero segment duration",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 0,
			},
			expectValid: false,
			errorMsg:    "segment_duration must be at least 1 second",
		},
		{
			name: "zero max segment size",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 3600,
				MaxSegmentSize: 0,
			},
			expectValid: false,
			errorMsg:    "max_segment_size must be at least 1 byte",
		},
		{
			name: "zero cleanup interval",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 3600,
				MaxSegmentSize: 524288000,
				CleanupInterval: 0,
			},
			expectValid: false,
			errorMsg:    "cleanup_interval must be at least 1 second",
		},
		{
			name: "zero max age",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 3600,
				MaxSegmentSize: 524288000,
				CleanupInterval: 86400,
				MaxAge: 0,
			},
			expectValid: false,
			errorMsg:    "max_age must be at least 1 second",
		},
		{
			name: "zero max size",
			config: &RecordingConfig{
				Format: "fmp4",
				Quality: "high",
				SegmentDuration: 3600,
				MaxSegmentSize: 524288000,
				CleanupInterval: 86400,
				MaxAge: 604800,
				MaxSize: 0,
			},
			expectValid: false,
			errorMsg:    "max_size must be at least 1 byte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecordingConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_Snapshot tests edge cases for snapshot validation
func TestValidationEdgeCases_Snapshot(t *testing.T) {
	tests := []struct {
		name        string
		config      *SnapshotConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid snapshot config",
			config: &SnapshotConfig{
				Enabled: true,
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 1920,
				MaxHeight: 1080,
				AutoCleanup: true,
				CleanupInterval: 3600,
				MaxAge: 86400,
				MaxCount: 1000,
			},
			expectValid: true,
		},
		{
			name: "invalid snapshot format",
			config: &SnapshotConfig{
				Format: "invalid_format",
				Quality: 90,
			},
			expectValid: false,
			errorMsg:    "invalid snapshot format",
		},
		{
			name: "quality too low",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 0,
			},
			expectValid: false,
			errorMsg:    "snapshot quality must be between 1 and 100",
		},
		{
			name: "quality too high",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 101,
			},
			expectValid: false,
			errorMsg:    "snapshot quality must be between 1 and 100",
		},
		{
			name: "zero max width",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 0,
			},
			expectValid: false,
			errorMsg:    "max_width must be at least 1 pixel",
		},
		{
			name: "zero max height",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 1920,
				MaxHeight: 0,
			},
			expectValid: false,
			errorMsg:    "max_height must be at least 1 pixel",
		},
		{
			name: "zero cleanup interval",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 1920,
				MaxHeight: 1080,
				CleanupInterval: 0,
			},
			expectValid: false,
			errorMsg:    "cleanup_interval must be at least 1 second",
		},
		{
			name: "zero max age",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 1920,
				MaxHeight: 1080,
				CleanupInterval: 3600,
				MaxAge: 0,
			},
			expectValid: false,
			errorMsg:    "max_age must be at least 1 second",
		},
		{
			name: "zero max count",
			config: &SnapshotConfig{
				Format: "jpeg",
				Quality: 90,
				MaxWidth: 1920,
				MaxHeight: 1080,
				CleanupInterval: 3600,
				MaxAge: 86400,
				MaxCount: 0,
			},
			expectValid: false,
			errorMsg:    "max_count must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSnapshotConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_FFmpeg tests edge cases for FFmpeg validation
func TestValidationEdgeCases_FFmpeg(t *testing.T) {
	tests := []struct {
		name        string
		config      *FFmpegConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid FFmpeg config",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 5.0,
					ExecutionTimeout: 8.0,
					InternalTimeout: 5000000,
					RetryAttempts: 2,
					RetryDelay: 1.0,
				},
				Recording: FFmpegOperationConfig{
					ProcessCreationTimeout: 10.0,
					ExecutionTimeout: 15.0,
					InternalTimeout: 10000000,
					RetryAttempts: 3,
					RetryDelay: 2.0,
				},
			},
			expectValid: true,
		},
		{
			name: "zero process creation timeout",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 0.0,
					ExecutionTimeout: 8.0,
					InternalTimeout: 5000000,
					RetryAttempts: 2,
					RetryDelay: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "process_creation_timeout must be positive",
		},
		{
			name: "zero execution timeout",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 5.0,
					ExecutionTimeout: 0.0,
					InternalTimeout: 5000000,
					RetryAttempts: 2,
					RetryDelay: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "execution_timeout must be positive",
		},
		{
			name: "zero internal timeout",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 5.0,
					ExecutionTimeout: 8.0,
					InternalTimeout: 0,
					RetryAttempts: 2,
					RetryDelay: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "internal_timeout must be positive",
		},
		{
			name: "negative retry attempts",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 5.0,
					ExecutionTimeout: 8.0,
					InternalTimeout: 5000000,
					RetryAttempts: -1,
					RetryDelay: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "retry_attempts must be non-negative",
		},
		{
			name: "negative retry delay",
			config: &FFmpegConfig{
				Snapshot: FFmpegOperationConfig{
					ProcessCreationTimeout: 5.0,
					ExecutionTimeout: 8.0,
					InternalTimeout: 5000000,
					RetryAttempts: 2,
					RetryDelay: -1.0,
				},
			},
			expectValid: false,
			errorMsg:    "retry_delay must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFFmpegConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestValidationEdgeCases_Performance tests edge cases for performance validation
func TestValidationEdgeCases_Performance(t *testing.T) {
	tests := []struct {
		name        string
		config      *PerformanceConfig
		expectValid bool
		errorMsg    string
	}{
		{
			name: "valid performance config",
			config: &PerformanceConfig{
				ResponseTimeTargets: ResponseTimeTargets{
					SnapshotCapture: 2.0,
					RecordingStart: 2.0,
					RecordingStop: 2.0,
					FileListing: 1.0,
				},
				SnapshotTiers: SnapshotTiers{
					Tier1USBDirectTimeout: 0.5,
					Tier2RTSPReadyCheckTimeout: 1.0,
					Tier3ActivationTimeout: 3.0,
					Tier3ActivationTriggerTimeout: 1.0,
					TotalOperationTimeout: 10.0,
					ImmediateResponseThreshold: 0.5,
					AcceptableResponseThreshold: 2.0,
					SlowResponseThreshold: 5.0,
				},
				Optimization: OptimizationConfig{
					EnableCaching: true,
					CacheTTL: 300,
					MaxConcurrentOperations: 5,
					ConnectionPoolSize: 10,
				},
			},
			expectValid: true,
		},
		{
			name: "zero snapshot capture target",
			config: &PerformanceConfig{
				ResponseTimeTargets: ResponseTimeTargets{
					SnapshotCapture: 0.0,
					RecordingStart: 2.0,
					RecordingStop: 2.0,
					FileListing: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "snapshot_capture must be positive",
		},
		{
			name: "zero recording start target",
			config: &PerformanceConfig{
				ResponseTimeTargets: ResponseTimeTargets{
					SnapshotCapture: 2.0,
					RecordingStart: 0.0,
					RecordingStop: 2.0,
					FileListing: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "recording_start must be positive",
		},
		{
			name: "zero recording stop target",
			config: &PerformanceConfig{
				ResponseTimeTargets: ResponseTimeTargets{
					SnapshotCapture: 2.0,
					RecordingStart: 2.0,
					RecordingStop: 0.0,
					FileListing: 1.0,
				},
			},
			expectValid: false,
			errorMsg:    "recording_stop must be positive",
		},
		{
			name: "zero file listing target",
			config: &PerformanceConfig{
				ResponseTimeTargets: ResponseTimeTargets{
					SnapshotCapture: 2.0,
					RecordingStart: 2.0,
					RecordingStop: 2.0,
					FileListing: 0.0,
				},
			},
			expectValid: false,
			errorMsg:    "file_listing must be positive",
		},
		{
			name: "zero tier1 timeout",
			config: &PerformanceConfig{
				SnapshotTiers: SnapshotTiers{
					Tier1USBDirectTimeout: 0.0,
					Tier2RTSPReadyCheckTimeout: 1.0,
					Tier3ActivationTimeout: 3.0,
					Tier3ActivationTriggerTimeout: 1.0,
					TotalOperationTimeout: 10.0,
					ImmediateResponseThreshold: 0.5,
					AcceptableResponseThreshold: 2.0,
					SlowResponseThreshold: 5.0,
				},
			},
			expectValid: false,
			errorMsg:    "tier1_usb_direct_timeout must be positive",
		},
		{
			name: "zero max concurrent operations",
			config: &PerformanceConfig{
				Optimization: OptimizationConfig{
					EnableCaching: true,
					CacheTTL: 300,
					MaxConcurrentOperations: 0,
					ConnectionPoolSize: 10,
				},
			},
			expectValid: false,
			errorMsg:    "max_concurrent_operations must be positive",
		},
		{
			name: "zero connection pool size",
			config: &PerformanceConfig{
				Optimization: OptimizationConfig{
					EnableCaching: true,
					CacheTTL: 300,
					MaxConcurrentOperations: 5,
					ConnectionPoolSize: 0,
				},
			},
			expectValid: false,
			errorMsg:    "connection_pool_size must be positive",
		},
		{
			name: "caching enabled but zero TTL",
			config: &PerformanceConfig{
				Optimization: OptimizationConfig{
					EnableCaching: true,
					CacheTTL: 0,
					MaxConcurrentOperations: 5,
					ConnectionPoolSize: 10,
				},
			},
			expectValid: false,
			errorMsg:    "cache_ttl must be positive when caching is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePerformanceConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}
