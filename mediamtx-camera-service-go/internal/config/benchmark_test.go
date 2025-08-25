package config

import (
	"testing"
	"time"
)

// BenchmarkConfigLoading measures the performance of configuration loading
func BenchmarkConfigLoading(b *testing.B) {
	loader := NewConfigLoader()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadConfig("config/default.yaml")
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
	}
}

// BenchmarkConfigLoadingWithDefaults measures performance with default values
func BenchmarkConfigLoadingWithDefaults(b *testing.B) {
	loader := NewConfigLoader()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadConfig("non-existent-file.yaml")
		if err != nil {
			// Expected to fail, but should still load defaults
			config := &Config{}
			loader.setDefaults()
			_ = validateConfig(config)
		}
	}
}

// BenchmarkConfigValidation measures validation performance
func BenchmarkConfigValidation(b *testing.B) {
	loader := NewConfigLoader()
	config, err := loader.LoadConfig("config/default.yaml")
	if err != nil {
		b.Fatalf("Failed to load config for benchmark: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateConfig(config)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkHotReloadWatcher measures file watching performance
func BenchmarkHotReloadWatcher(b *testing.B) {
	callback := func(*Config) error {
		return nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher, _ := NewConfigWatcher("config/default.yaml", callback)
		watcher.Start()
		time.Sleep(1 * time.Millisecond) // Brief pause to simulate file watching
		watcher.Stop()
	}
}

// BenchmarkEnvironmentVariableOverrides measures env var binding performance
func BenchmarkEnvironmentVariableOverrides(b *testing.B) {
	loader := NewConfigLoader()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate environment variable binding
		loader.viper.BindEnv("server.host", "CAMERA_SERVER_HOST")
		loader.viper.BindEnv("server.port", "CAMERA_SERVER_PORT")
		loader.viper.BindEnv("mediamtx.host", "MEDIAMTX_HOST")
		loader.viper.BindEnv("mediamtx.api_port", "MEDIAMTX_API_PORT")
		loader.viper.BindEnv("camera.poll_interval", "CAMERA_POLL_INTERVAL")
		loader.viper.BindEnv("logging.level", "CAMERA_LOG_LEVEL")
	}
}

// BenchmarkStructInitialization measures struct creation performance
func BenchmarkStructInitialization(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &Config{
			Server: ServerConfig{
				Host: "127.0.0.1",
				Port: 8002,
			},
			MediaMTX: MediaMTXConfig{
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
			Camera: CameraConfig{
				PollInterval: 0.1,
				DetectionTimeout: 2.0,
				DeviceRange: []int{0, 9},
				EnableCapabilityDetection: true,
				AutoStartStreams: true,
				CapabilityTimeout: 5.0,
				CapabilityRetryInterval: 1.0,
				CapabilityMaxRetries: 3,
			},
			Logging: LoggingConfig{
				Level: "INFO",
				Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
				FileEnabled: true,
				FilePath: "/opt/camera-service/logs/camera-service.log",
				MaxFileSize: 10485760,
				BackupCount: 5,
				ConsoleEnabled: true,
			},
			Recording: RecordingConfig{
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
			Snapshots: SnapshotConfig{
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
			FFmpeg: FFmpegConfig{
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
			Notifications: NotificationsConfig{
				WebSocket: WebSocketNotificationConfig{
					DeliveryTimeout: 5.0,
					RetryAttempts: 3,
					RetryDelay: 1.0,
				},
				RealTime: RealTimeNotificationConfig{
					CameraStatusInterval: 30.0,
					RecordingProgressInterval: 5.0,
					ConnectionHealthCheck: 10.0,
				},
			},
			Performance: PerformanceConfig{
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
		}
		_ = config // Prevent compiler optimization
	}
}

// BenchmarkValidationFunctions measures individual validation function performance
func BenchmarkValidationFunctions(b *testing.B) {
	// Create valid configs for each validation function
	mediaMTXConfig := &MediaMTXConfig{
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
	}
	
	cameraConfig := &CameraConfig{
		PollInterval: 0.1,
		DetectionTimeout: 2.0,
		DeviceRange: []int{0, 9},
		EnableCapabilityDetection: true,
		AutoStartStreams: true,
		CapabilityTimeout: 5.0,
		CapabilityRetryInterval: 1.0,
		CapabilityMaxRetries: 3,
	}
	
	loggingConfig := &LoggingConfig{
		Level: "INFO",
		Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
		FileEnabled: true,
		FilePath: "/opt/camera-service/logs/camera-service.log",
		MaxFileSize: 10485760,
		BackupCount: 5,
		ConsoleEnabled: true,
	}
	
	recordingConfig := &RecordingConfig{
		Enabled: false,
		Format: "fmp4",
		Quality: "high",
		SegmentDuration: 3600,
		MaxSegmentSize: 524288000,
		AutoCleanup: true,
		CleanupInterval: 86400,
		MaxAge: 604800,
		MaxSize: 10737418240,
	}
	
	snapshotConfig := &SnapshotConfig{
		Enabled: true,
		Format: "jpeg",
		Quality: 90,
		MaxWidth: 1920,
		MaxHeight: 1080,
		AutoCleanup: true,
		CleanupInterval: 3600,
		MaxAge: 86400,
		MaxCount: 1000,
	}
	
	ffmpegConfig := &FFmpegConfig{
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
	}
	
	performanceConfig := &PerformanceConfig{
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
	}
	
	b.Run("MediaMTXValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateMediaMTXConfig(mediaMTXConfig)
			if err != nil {
				b.Fatalf("MediaMTX validation failed: %v", err)
			}
		}
	})
	
	b.Run("CameraValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateCameraConfig(cameraConfig)
			if err != nil {
				b.Fatalf("Camera validation failed: %v", err)
			}
		}
	})
	
	b.Run("LoggingValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateLoggingConfig(loggingConfig)
			if err != nil {
				b.Fatalf("Logging validation failed: %v", err)
			}
		}
	})
	
	b.Run("RecordingValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateRecordingConfig(recordingConfig)
			if err != nil {
				b.Fatalf("Recording validation failed: %v", err)
			}
		}
	})
	
	b.Run("SnapshotValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateSnapshotConfig(snapshotConfig)
			if err != nil {
				b.Fatalf("Snapshot validation failed: %v", err)
			}
		}
	})
	
	b.Run("FFmpegValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validateFFmpegConfig(ffmpegConfig)
			if err != nil {
				b.Fatalf("FFmpeg validation failed: %v", err)
			}
		}
	})
	
	b.Run("PerformanceValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := validatePerformanceConfig(performanceConfig)
			if err != nil {
				b.Fatalf("Performance validation failed: %v", err)
			}
		}
	})
}

// BenchmarkMemoryUsage measures memory allocation during config operations
func BenchmarkMemoryUsage(b *testing.B) {
	loader := NewConfigLoader()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		config, err := loader.LoadConfig("config/default.yaml")
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
		
		err = validateConfig(config)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
		
		_ = config // Prevent compiler optimization
	}
}
