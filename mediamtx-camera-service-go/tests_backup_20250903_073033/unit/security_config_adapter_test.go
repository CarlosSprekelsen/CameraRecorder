package security

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/assert"
)

// Create test config instances
func createTestSecurityConfig() *config.SecurityConfig {
	return &config.SecurityConfig{
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		JWTSecretKey:      "test_secret",
		JWTExpiryHours:    24,
	}
}

func createTestLoggingConfig() *config.LoggingConfig {
	return &config.LoggingConfig{
		Level:          "info",
		Format:         "json",
		FileEnabled:    true,
		FilePath:       "/var/log/test",
		MaxFileSize:    100 * 1024 * 1024,
		BackupCount:    5,
		ConsoleEnabled: true,
	}
}

func TestNewConfigAdapter(t *testing.T) {
	securityConfig := createTestSecurityConfig()
	loggingConfig := createTestLoggingConfig()

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	assert.NotNil(t, adapter)
	assert.Equal(t, securityConfig, adapter.GetSecurityConfig())
	assert.Equal(t, loggingConfig, adapter.GetLoggingConfig())
}

func TestConfigAdapter_GetRateLimitRequests(t *testing.T) {
	securityConfig := &SecurityConfig{RateLimitRequests: 150}
	loggingConfig := &LoggingConfig{}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetRateLimitRequests()
	assert.Equal(t, 150, result)
}

func TestConfigAdapter_GetRateLimitRequests_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetRateLimitRequests()
	assert.Equal(t, 100, result) // Default fallback
}

func TestConfigAdapter_GetRateLimitWindow(t *testing.T) {
	securityConfig := &SecurityConfig{RateLimitWindow: 2 * time.Minute}
	loggingConfig := &LoggingConfig{}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetRateLimitWindow()
	assert.Equal(t, 2*time.Minute, result)
}

func TestConfigAdapter_GetRateLimitWindow_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetRateLimitWindow()
	assert.Equal(t, time.Minute, result) // Default fallback
}

func TestConfigAdapter_GetJWTSecretKey(t *testing.T) {
	securityConfig := &SecurityConfig{JWTSecretKey: "custom_secret"}
	loggingConfig := &LoggingConfig{}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetJWTSecretKey()
	assert.Equal(t, "custom_secret", result)
}

func TestConfigAdapter_GetJWTSecretKey_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetJWTSecretKey()
	assert.Equal(t, "", result) // Default fallback
}

func TestConfigAdapter_GetJWTExpiryHours(t *testing.T) {
	securityConfig := &SecurityConfig{JWTExpiryHours: 48}
	loggingConfig := &LoggingConfig{}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetJWTExpiryHours()
	assert.Equal(t, 48, result)
}

func TestConfigAdapter_GetJWTExpiryHours_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetJWTExpiryHours()
	assert.Equal(t, 24, result) // Default fallback
}

func TestConfigAdapter_GetLogLevel(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{Level: "debug"}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetLogLevel()
	assert.Equal(t, "debug", result)
}

func TestConfigAdapter_GetLogLevel_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetLogLevel()
	assert.Equal(t, "info", result) // Default fallback
}

func TestConfigAdapter_GetLogFormat(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{Format: "text"}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetLogFormat()
	assert.Equal(t, "text", result)
}

func TestConfigAdapter_GetLogFormat_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetLogFormat()
	assert.Equal(t, "json", result) // Default fallback
}

func TestConfigAdapter_IsFileLoggingEnabled(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{FileEnabled: true}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.IsFileLoggingEnabled()
	assert.True(t, result)
}

func TestConfigAdapter_IsFileLoggingEnabled_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.IsFileLoggingEnabled()
	assert.False(t, result) // Default fallback
}

func TestConfigAdapter_GetLogFilePath(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{FilePath: "/custom/log/path"}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetLogFilePath()
	assert.Equal(t, "/custom/log/path", result)
}

func TestConfigAdapter_GetLogFilePath_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetLogFilePath()
	assert.Equal(t, "/var/log/camera-service", result) // Default fallback
}

func TestConfigAdapter_GetMaxLogFileSize(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{MaxFileSize: 200 * 1024 * 1024}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetMaxLogFileSize()
	assert.Equal(t, int64(200*1024*1024), result)
}

func TestConfigAdapter_GetMaxLogFileSize_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetMaxLogFileSize()
	assert.Equal(t, int64(100*1024*1024), result) // Default fallback
}

func TestConfigAdapter_GetLogBackupCount(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{BackupCount: 10}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.GetLogBackupCount()
	assert.Equal(t, 10, result)
}

func TestConfigAdapter_GetLogBackupCount_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.GetLogBackupCount()
	assert.Equal(t, 5, result) // Default fallback
}

func TestConfigAdapter_IsConsoleLoggingEnabled(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{ConsoleEnabled: false}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	result := adapter.IsConsoleLoggingEnabled()
	assert.False(t, result)
}

func TestConfigAdapter_IsConsoleLoggingEnabled_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	result := adapter.IsConsoleLoggingEnabled()
	assert.True(t, result) // Default fallback
}

func TestConfigAdapter_CreateAuditLoggerConfig(t *testing.T) {
	securityConfig := &SecurityConfig{}
	loggingConfig := &LoggingConfig{
		FilePath:       "/var/log/test",
		MaxFileSize:    50 * 1024 * 1024,
		FileEnabled:    true,
		ConsoleEnabled: false,
		Level:          "warn",
	}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	config := adapter.CreateAuditLoggerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "/var/log/test/security", config["log_directory"])
	assert.Equal(t, int64(50*1024*1024), config["max_file_size"])
	assert.Equal(t, 30*24*time.Hour, config["max_file_age"])
	assert.Equal(t, time.Hour, config["rotation_interval"])
	assert.Equal(t, 1000, config["buffer_size"])
	assert.True(t, config["enable_file_logging"].(bool))
	assert.False(t, config["enable_console_logging"].(bool))
	assert.Equal(t, "warn", config["log_level"])
}

func TestConfigAdapter_CreateRateLimiterConfig(t *testing.T) {
	securityConfig := &SecurityConfig{
		RateLimitRequests: 200,
		RateLimitWindow:   30 * time.Second,
	}
	loggingConfig := &LoggingConfig{}

	adapter := NewConfigAdapter(securityConfig, loggingConfig)

	config := adapter.CreateRateLimiterConfig()

	assert.NotNil(t, config)

	// Check that all expected methods are present
	expectedMethods := []string{"ping", "get_camera_list", "start_recording", "take_snapshot", "authenticate"}
	for _, method := range expectedMethods {
		methodConfig, exists := config[method]
		assert.True(t, exists, "Method %s should be present", method)
		assert.NotNil(t, methodConfig, "RateLimitConfig should not be nil for method %s", method)
	}
}

func TestConfigAdapter_CreateRateLimiterConfig_NilConfig(t *testing.T) {
	adapter := NewConfigAdapter(nil, nil)

	config := adapter.CreateRateLimiterConfig()

	assert.NotNil(t, config)

	// Should use default values
	pingConfig, exists := config["ping"]
	assert.True(t, exists, "ping method should be present")
	assert.NotNil(t, pingConfig, "ping RateLimitConfig should not be nil")
}
