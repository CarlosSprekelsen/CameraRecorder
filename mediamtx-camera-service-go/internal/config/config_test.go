package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigLoader(t *testing.T) {
	loader := NewConfigLoader()
	assert.NotNil(t, loader)
	assert.NotNil(t, loader.viper)
	assert.NotNil(t, loader.logger)
}

func TestLoadConfigWithDefaults(t *testing.T) {
	loader := NewConfigLoader()
	
	// Test loading with non-existent file (should use defaults)
	config, err := loader.LoadConfig("non-existent-file.yaml")
	require.NoError(t, err)
	assert.NotNil(t, config)
	
	// Verify default values
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8002, config.Server.Port)
	assert.Equal(t, "/ws", config.Server.WebSocketPath)
	assert.Equal(t, 100, config.Server.MaxConnections)
	
	assert.Equal(t, "127.0.0.1", config.MediaMTX.Host)
	assert.Equal(t, 9997, config.MediaMTX.APIPort)
	assert.Equal(t, 8554, config.MediaMTX.RTSPPort)
	assert.Equal(t, 8889, config.MediaMTX.WebRTCPort)
	assert.Equal(t, 8888, config.MediaMTX.HLSPort)
	
	assert.Equal(t, 0.1, config.Camera.PollInterval)
	assert.Equal(t, 2.0, config.Camera.DetectionTimeout)
	assert.Equal(t, []int{0, 9}, config.Camera.DeviceRange)
	assert.True(t, config.Camera.EnableCapabilityDetection)
	assert.True(t, config.Camera.AutoStartStreams)
	
	assert.Equal(t, "INFO", config.Logging.Level)
	assert.True(t, config.Logging.FileEnabled)
	assert.True(t, config.Logging.ConsoleEnabled)
	
	assert.False(t, config.Recording.Enabled)
	// AutoRecord field removed from Python YAML structure
	assert.Equal(t, "fmp4", config.Recording.Format)
	assert.Equal(t, "high", config.Recording.Quality)
	
	assert.True(t, config.Snapshots.Enabled)
	assert.Equal(t, "jpeg", config.Snapshots.Format)
	assert.Equal(t, 90, config.Snapshots.Quality)
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempFile := createTempConfigFile(t)
	defer os.Remove(tempFile)
	
	loader := NewConfigLoader()
	config, err := loader.LoadConfig(tempFile)
	require.NoError(t, err)
	assert.NotNil(t, config)
	
	// Verify values from file
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, 9000, config.Server.Port)
	assert.Equal(t, "/test", config.Server.WebSocketPath)
	assert.Equal(t, 200, config.Server.MaxConnections)
	
	assert.Equal(t, "192.168.1.100", config.MediaMTX.Host)
	assert.Equal(t, 9998, config.MediaMTX.APIPort)
	
	assert.Equal(t, 0.5, config.Camera.PollInterval)
	assert.Equal(t, 5.0, config.Camera.DetectionTimeout)
	assert.Equal(t, []int{0, 1, 2}, config.Camera.DeviceRange)
	assert.False(t, config.Camera.EnableCapabilityDetection)
	assert.True(t, config.Camera.AutoStartStreams)
	
	assert.Equal(t, "DEBUG", config.Logging.Level)
	assert.True(t, config.Logging.FileEnabled)
	assert.False(t, config.Logging.ConsoleEnabled)
	
	assert.True(t, config.Recording.Enabled)
	// AutoRecord field removed from Python YAML structure
	assert.Equal(t, "mp4", config.Recording.Format)
	assert.Equal(t, "high", config.Recording.Quality)
	
	assert.False(t, config.Snapshots.Enabled)
	assert.Equal(t, "png", config.Snapshots.Format)
	assert.Equal(t, 95, config.Snapshots.Quality)
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "10.0.0.1")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "9001")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_HOST", "10.0.0.2")
	os.Setenv("CAMERA_SERVICE_CAMERA_POLL_INTERVAL", "0.2")
	os.Setenv("CAMERA_SERVICE_LOGGING_LEVEL", "ERROR")
	os.Setenv("CAMERA_SERVICE_RECORDING_ENABLED", "true")
	os.Setenv("CAMERA_SERVICE_SNAPSHOTS_ENABLED", "false")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_HOST")
		os.Unsetenv("CAMERA_SERVICE_CAMERA_POLL_INTERVAL")
		os.Unsetenv("CAMERA_SERVICE_LOGGING_LEVEL")
		os.Unsetenv("CAMERA_SERVICE_RECORDING_ENABLED")
		os.Unsetenv("CAMERA_SERVICE_SNAPSHOTS_ENABLED")
	}()
	
	config, err := NewConfigLoader().LoadConfig("non-existent-file.yaml")
	require.NoError(t, err)
	assert.NotNil(t, config)
	
	// Verify environment variable overrides
	assert.Equal(t, "10.0.0.1", config.Server.Host)
	assert.Equal(t, 9001, config.Server.Port)
	assert.Equal(t, "10.0.0.2", config.MediaMTX.Host)
	assert.Equal(t, 0.2, config.Camera.PollInterval)
	assert.Equal(t, "ERROR", config.Logging.Level)
	assert.True(t, config.Recording.Enabled)
	assert.False(t, config.Snapshots.Enabled)
}

func TestConfigValidation(t *testing.T) {
	
	// Test invalid port
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 70000, // Invalid port
		},
		MediaMTX: MediaMTXConfig{
			Host:    "localhost",
			APIPort: 9997,
			RTSPPort: 8554,
			WebRTCPort: 8889,
			HLSPort: 8888,
		},
		Camera: CameraConfig{
			PollInterval: 0.1,
			DetectionTimeout: 1.0,
			DeviceRange: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		Logging: LoggingConfig{
			Level: "INFO",
			Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
		},
		Recording: RecordingConfig{
			Format: "fmp4",
			Quality: "medium",
		},
		Snapshots: SnapshotConfig{
			Format: "jpeg",
			Quality: 85,
		},
	}
	
	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port must be between 1 and 65535")
	
	// Test invalid logging level
	config = &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8002,
		},
		MediaMTX: MediaMTXConfig{
			Host:    "localhost",
			APIPort: 9997,
			RTSPPort: 8554,
			WebRTCPort: 8889,
			HLSPort: 8888,
		},
		Camera: CameraConfig{
			PollInterval: 0.1,
			DetectionTimeout: 1.0,
			DeviceRange: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		Logging: LoggingConfig{
			Level: "INVALID_LEVEL",
			Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
		},
		Recording: RecordingConfig{
			Format: "fmp4",
			Quality: "medium",
		},
		Snapshots: SnapshotConfig{
			Format: "jpeg",
			Quality: 85,
		},
	}
	
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid logging level")
	
	// Test invalid recording format
	config = &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8002,
		},
		MediaMTX: MediaMTXConfig{
			Host:    "localhost",
			APIPort: 9997,
			RTSPPort: 8554,
			WebRTCPort: 8889,
			HLSPort: 8888,
		},
		Camera: CameraConfig{
			PollInterval: 0.1,
			DetectionTimeout: 1.0,
			DeviceRange: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		Logging: LoggingConfig{
			Level: "INFO",
			Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
		},
		Recording: RecordingConfig{
			Format: "invalid_format",
			Quality: "medium",
		},
		Snapshots: SnapshotConfig{
			Format: "jpeg",
			Quality: 85,
		},
	}
	
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recording format")
	
	// Test invalid snapshot quality
	config = &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8002,
		},
		MediaMTX: MediaMTXConfig{
			Host:    "localhost",
			APIPort: 9997,
			RTSPPort: 8554,
			WebRTCPort: 8889,
			HLSPort: 8888,
		},
		Camera: CameraConfig{
			PollInterval: 0.1,
			DetectionTimeout: 1.0,
			DeviceRange: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		Logging: LoggingConfig{
			Level: "INFO",
			Format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s",
		},
		Recording: RecordingConfig{
			Format: "fmp4",
			Quality: "medium",
		},
		Snapshots: SnapshotConfig{
			Format: "jpeg",
			Quality: 150, // Invalid quality
		},
	}
	
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot quality must be between 1 and 100")
}

func TestConfigString(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8002,
		},
		MediaMTX: MediaMTXConfig{
			Host:    "localhost",
			APIPort: 9997,
		},
		Camera: CameraConfig{
			PollInterval: 0.1,
		},
		Logging: LoggingConfig{
			Level: "INFO",
		},
		Recording: RecordingConfig{
			Enabled: false,
		},
		Snapshots: SnapshotConfig{
			Enabled: true,
		},
	}
	
	str := config.String()
	assert.Contains(t, str, "Server: 127.0.0.1:8002")
	assert.Contains(t, str, "MediaMTX: localhost:9997")
	assert.Contains(t, str, "Camera: poll_interval=0.1")
	assert.Contains(t, str, "Logging: level=INFO")
	assert.Contains(t, str, "Recording: enabled=false")
	assert.Contains(t, str, "Snapshots: enabled=true")
}

// Helper function to create a temporary config file for testing
func createTempConfigFile(t *testing.T) string {
	content := `
server:
  host: "127.0.0.1"
  port: 9000
  websocket_path: "/test"
  max_connections: 200

mediamtx:
  host: "192.168.1.100"
  api_port: 9998
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/tmp/mediamtx.yml"
  recordings_path: "/tmp/recordings"
  snapshots_path: "/tmp/snapshots"

camera:
  poll_interval: 0.5
  detection_timeout: 5.0
  device_range: [0, 1, 2]
  enable_capability_detection: false
  auto_start_streams: true
  capability_timeout: 5.0
  capability_retry_interval: 1.0
  capability_max_retries: 3

logging:
  level: "DEBUG"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/tmp/camera-service.log"
  max_file_size: 10485760
  backup_count: 5
  console_enabled: false

recording:
  enabled: true
  auto_record: true
  format: "mp4"
  quality: "high"
  segment_duration: 3600
  max_segment_size: 524288000
  auto_cleanup: true
  cleanup_interval: 86400
  max_age: 604800
  max_size: 10737418240
  max_duration: 3600
  cleanup_after_days: 30
  rotation_minutes: 30
  storage_warn_percent: 80
  storage_block_percent: 90

snapshots:
  enabled: false
  format: "png"
  quality: 95
  max_width: 1920
  max_height: 1080
  auto_cleanup: true
  cleanup_interval: 3600
  max_age: 86400
  max_count: 1000
  cleanup_after_days: 7
`
	
	tempFile := t.TempDir() + "/test-config.yaml"
	err := os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(t, err)
	
	return tempFile
}

// Edge case tests for IV&V validation requirements
func TestPortBoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectValid bool
	}{
		{"minimum valid port", 1, true},
		{"maximum valid port", 65535, true},
		{"zero port", 0, false},
		{"negative port", -1, false},
		{"port too high", 65536, false},
		{"port way too high", 99999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{Port: tt.port, Host: "localhost", WebSocketPath: "/ws", MaxConnections: 100},
			}
			err := validateConfig(config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestInvalidYAMLHandling(t *testing.T) {
	// Test with malformed YAML
	loader := NewConfigLoader()
	
	// Create a temporary file with invalid YAML
	tempFile := t.TempDir() + "/invalid.yaml"
	err := os.WriteFile(tempFile, []byte(`
server:
  host: "localhost"
  port: invalid_port
  - invalid: yaml: structure
`), 0644)
	require.NoError(t, err)

	_, err = loader.LoadConfig(tempFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestEnvironmentVariableEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		envValue    string
		expectValid bool
	}{
		{"empty host", "CAMERA_SERVICE_SERVER_HOST", "", false},
		{"whitespace host", "CAMERA_SERVICE_SERVER_HOST", "   ", false},
		{"invalid host format", "CAMERA_SERVICE_SERVER_HOST", "invalid@host", false},
		{"valid host", "CAMERA_SERVICE_SERVER_HOST", "localhost", true},
		{"valid IP", "CAMERA_SERVICE_SERVER_HOST", "127.0.0.1", true},
		{"negative port", "CAMERA_SERVICE_SERVER_PORT", "-1", false},
		{"zero port", "CAMERA_SERVICE_SERVER_PORT", "0", false},
		{"valid port", "CAMERA_SERVICE_SERVER_PORT", "8080", true},
		{"port too high", "CAMERA_SERVICE_SERVER_PORT", "65536", false},
		{"non-numeric port", "CAMERA_SERVICE_SERVER_PORT", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tt.envVar, tt.envValue)
			defer os.Unsetenv(tt.envVar)

			loader := NewConfigLoader()
			config, err := loader.LoadConfig("non-existent-file.yaml")

			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		errorMsg    string
	}{
		{
			name: "empty server host",
			config: &Config{
				Server: ServerConfig{Host: "", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server host cannot be empty",
		},
		{
			name: "whitespace server host",
			config: &Config{
				Server: ServerConfig{Host: "   ", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server host cannot be empty",
		},
		{
			name: "invalid server host format",
			config: &Config{
				Server: ServerConfig{Host: "invalid@host", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "invalid server host format",
		},
		{
			name: "negative server port",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: -1, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server port must be between 1 and 65535",
		},
		{
			name: "zero server port",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 0, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server port must be between 1 and 65535",
		},
		{
			name: "server port too high",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 65536, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server port must be between 1 and 65535",
		},
		{
			name: "empty websocket path",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "websocket path cannot be empty",
		},
		{
			name: "negative max connections",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: -1},
			},
			expectValid: false,
			errorMsg:    "max connections must be positive",
		},
		{
			name: "zero max connections",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 0},
			},
			expectValid: false,
			errorMsg:    "max connections must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
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

func TestCrossFieldValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		errorMsg    string
	}{
		{
			name: "server port conflicts with MediaMTX API port",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				MediaMTX: MediaMTXConfig{APIPort: 8080},
			},
			expectValid: false,
			errorMsg:    "server port conflicts with MediaMTX API port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
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

func TestTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		errorMsg    string
	}{
		{
			name: "invalid video profile",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				MediaMTX: MediaMTXConfig{Codec: CodecConfig{VideoProfile: "invalid_profile"}},
			},
			expectValid: false,
			errorMsg:    "invalid video profile",
		},
		{
			name: "invalid video level",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				MediaMTX: MediaMTXConfig{Codec: CodecConfig{VideoLevel: "invalid_level"}},
			},
			expectValid: false,
			errorMsg:    "invalid video level",
		},
		{
			name: "invalid log level",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				Logging: LoggingConfig{Level: "invalid_level"},
			},
			expectValid: false,
			errorMsg:    "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
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

func TestBoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name: "minimum valid values",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 1, WebSocketPath: "/ws", MaxConnections: 1},
				Camera: CameraConfig{PollInterval: 0.01, DetectionTimeout: 0.1, DeviceRange: []int{0}},
			},
			expectValid: true,
		},
		{
			name: "maximum valid values",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 65535, WebSocketPath: "/ws", MaxConnections: 10000},
				Camera: CameraConfig{PollInterval: 1.0, DetectionTimeout: 10.0, DeviceRange: []int{10}},
			},
			expectValid: true,
		},
		{
			name: "exceed maximum values",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 65536, WebSocketPath: "/ws", MaxConnections: 10001},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEmptyArraysAndStrings(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name: "empty camera device range",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				Camera: CameraConfig{DeviceRange: []int{}},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestUnicodeAndSpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name: "unicode host name",
			config: &Config{
				Server: ServerConfig{Host: "localhost-测试", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestVeryLargeValues(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name: "very large port numbers",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 999999, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		errorMsg    string
	}{
		{
			name: "missing server host",
			config: &Config{
				Server: ServerConfig{Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
			},
			expectValid: false,
			errorMsg:    "server host cannot be empty",
		},
		{
			name: "missing camera device range",
			config: &Config{
				Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
				Camera: CameraConfig{},
			},
			expectValid: false,
			errorMsg:    "device_range cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
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

// Performance benchmarks for IV&V validation
func BenchmarkConfigLoadingLegacy(b *testing.B) {
	loader := NewConfigLoader()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadConfig("non-existent-file.yaml")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidationLegacy(b *testing.B) {
	config := &Config{
		Server: ServerConfig{Host: "localhost", Port: 8080, WebSocketPath: "/ws", MaxConnections: 100},
		MediaMTX: MediaMTXConfig{
			Codec: CodecConfig{
				VideoProfile: "baseline",
				VideoLevel:   "3.1",
				PixelFormat:  "yuv420p",
				Preset:       "medium",
			},
		},
		Camera: CameraConfig{
			PollInterval:     1.0,
			DetectionTimeout: 5.0,
			DeviceRange:      []int{0, 1, 2},
		},
		Logging: LoggingConfig{Level: "info", Format: "json"},
		Recording: RecordingConfig{
			Enabled:         true,
			Format:          "mp4",
			Quality:         "high",
			SegmentDuration: 60,
		},
		Snapshots: SnapshotConfig{
			Enabled:  true,
			Format:   "jpg",
			Quality:  90,
			MaxWidth: 1920,
		},
		FFmpeg: FFmpegConfig{
			Snapshot: FFmpegOperationConfig{
				ProcessCreationTimeout: 5.0,
				ExecutionTimeout:       30.0,
				InternalTimeout:        60,
			},
			Recording: FFmpegOperationConfig{
				ProcessCreationTimeout: 5.0,
				ExecutionTimeout:       30.0,
				InternalTimeout:        60,
			},
		},
		Performance: PerformanceConfig{
			ResponseTimeTargets: ResponseTimeTargets{
				SnapshotCapture: 1.0,
				RecordingStart:  2.0,
				RecordingStop:   1.0,
				FileListing:     0.5,
			},
			SnapshotTiers: SnapshotTiers{
				Tier1USBDirectTimeout:         0.1,
				Tier2RTSPReadyCheckTimeout:    0.5,
				Tier3ActivationTimeout:        1.0,
				Tier3ActivationTriggerTimeout: 0.5,
				TotalOperationTimeout:         5.0,
				ImmediateResponseThreshold:    0.1,
				AcceptableResponseThreshold:   1.0,
				SlowResponseThreshold:         3.0,
			},
			Optimization: OptimizationConfig{
				EnableCaching:           true,
				CacheTTL:                300,
				MaxConcurrentOperations: 10,
				ConnectionPoolSize:      5,
			},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateConfig(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEnvironmentVariableOverridesLegacy(b *testing.B) {
	// Set environment variables
	os.Setenv("CAMERA_SERVICE_SERVER_HOST", "localhost")
	os.Setenv("CAMERA_SERVICE_SERVER_PORT", "8080")
	os.Setenv("CAMERA_SERVICE_MEDIAMTX_API_PORT", "8081")
	defer func() {
		os.Unsetenv("CAMERA_SERVICE_SERVER_HOST")
		os.Unsetenv("CAMERA_SERVICE_SERVER_PORT")
		os.Unsetenv("CAMERA_SERVICE_MEDIAMTX_API_PORT")
	}()

	loader := NewConfigLoader()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadConfig("non-existent-file.yaml")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test GetViper method for 100% coverage
func TestGetViper(t *testing.T) {
	loader := NewConfigLoader()
	viper := loader.GetViper()
	assert.NotNil(t, viper)
}
