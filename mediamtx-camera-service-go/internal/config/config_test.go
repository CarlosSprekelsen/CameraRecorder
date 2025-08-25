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
	
	assert.Equal(t, "localhost", config.MediaMTX.Host)
	assert.Equal(t, 9997, config.MediaMTX.APIPort)
	assert.Equal(t, 8554, config.MediaMTX.RTSPPort)
	assert.Equal(t, 8889, config.MediaMTX.WebRTCPort)
	assert.Equal(t, 8888, config.MediaMTX.HLSPort)
	
	assert.Equal(t, 0.1, config.Camera.PollInterval)
	assert.Equal(t, 1.0, config.Camera.DetectionTimeout)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, config.Camera.DeviceRange)
	assert.True(t, config.Camera.EnableCapabilityDetection)
	assert.False(t, config.Camera.AutoStartStreams)
	
	assert.Equal(t, "INFO", config.Logging.Level)
	assert.False(t, config.Logging.FileEnabled)
	assert.True(t, config.Logging.ConsoleEnabled)
	
	assert.False(t, config.Recording.Enabled)
	assert.False(t, config.Recording.AutoRecord)
	assert.Equal(t, "fmp4", config.Recording.Format)
	assert.Equal(t, "medium", config.Recording.Quality)
	
	assert.True(t, config.Snapshots.Enabled)
	assert.Equal(t, "jpeg", config.Snapshots.Format)
	assert.Equal(t, 85, config.Snapshots.Quality)
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
	assert.True(t, config.Recording.AutoRecord)
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
	
	loader := NewConfigLoader()
	config, err := loader.LoadConfig("non-existent-file.yaml")
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
	loader := NewConfigLoader()
	
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
	
	err := loader.validateConfig(config)
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
	
	err = loader.validateConfig(config)
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
	
	err = loader.validateConfig(config)
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
	
	err = loader.validateConfig(config)
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
