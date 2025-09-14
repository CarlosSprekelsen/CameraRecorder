//go:build unit && testhelpers
// +build unit,testhelpers

package security

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CONFIG ADAPTER TESTS FOR 100% COVERAGE
// =============================================================================

func TestNewConfigAdapter(t *testing.T) {
	t.Parallel()

	// Test with nil configs
	adapter := NewConfigAdapter(nil, nil)
	assert.NotNil(t, adapter)
	assert.Nil(t, adapter.securityConfig)
	assert.Nil(t, adapter.loggingConfig)

	// Test with actual configs
	securityConfig := &config.SecurityConfig{
		RateLimitRequests: 200,
		RateLimitWindow:   2 * time.Minute,
		JWTSecretKey:      "test_secret",
		JWTExpiryHours:    48,
	}
	loggingConfig := &config.LoggingConfig{
		Level:          "debug",
		Format:         "text",
		FileEnabled:    true,
		FilePath:       "/var/log/test",
		MaxFileSize:    200 * 1024 * 1024,
		BackupCount:    10,
		ConsoleEnabled: false,
	}

	adapter = NewConfigAdapter(securityConfig, loggingConfig)
	assert.NotNil(t, adapter)
	assert.Equal(t, securityConfig, adapter.securityConfig)
	assert.Equal(t, loggingConfig, adapter.loggingConfig)
}

func TestConfigAdapter_GetSecurityConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expectedConfig *config.SecurityConfig
	}{
		{"Nil security config", nil, nil, nil},
		{"Valid security config", &config.SecurityConfig{RateLimitRequests: 150}, nil, &config.SecurityConfig{RateLimitRequests: 150}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetSecurityConfig()
			assert.Equal(t, tt.expectedConfig, result)
		})
	}
}

func TestConfigAdapter_GetLoggingConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expectedConfig *config.LoggingConfig
	}{
		{"Nil logging config", nil, nil, nil},
		{"Valid logging config", nil, &config.LoggingConfig{Level: "warn"}, &config.LoggingConfig{Level: "warn"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetLoggingConfig()
			assert.Equal(t, tt.expectedConfig, result)
		})
	}
}

func TestConfigAdapter_GetRateLimitRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       int
	}{
		{"Nil config - use default", nil, nil, DefaultRateLimitRequests},
		{"Valid config", &config.SecurityConfig{RateLimitRequests: 250}, nil, 250},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetRateLimitRequests()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetRateLimitWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       time.Duration
	}{
		{"Nil config - use default", nil, nil, DefaultRateLimitWindow},
		{"Valid config", &config.SecurityConfig{RateLimitWindow: 5 * time.Minute}, nil, 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetRateLimitWindow()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetJWTSecretKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       string
	}{
		{"Nil config - empty string", nil, nil, ""},
		{"Valid config", &config.SecurityConfig{JWTSecretKey: "my_secret_key"}, nil, "my_secret_key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetJWTSecretKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetJWTExpiryHours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       int
	}{
		{"Nil config - use default", nil, nil, DefaultJWTExpiryHours},
		{"Valid config", &config.SecurityConfig{JWTExpiryHours: 72}, nil, 72},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetJWTExpiryHours()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       string
	}{
		{"Nil config - use default", nil, nil, DefaultLogLevel},
		{"Valid config", nil, &config.LoggingConfig{Level: "error"}, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetLogLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetLogFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       string
	}{
		{"Nil config - use default", nil, nil, DefaultLogFormat},
		{"Valid config", nil, &config.LoggingConfig{Format: "text"}, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetLogFormat()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_IsFileLoggingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       bool
	}{
		{"Nil config - use default", nil, nil, false},
		{"Valid config - enabled", nil, &config.LoggingConfig{FileEnabled: true}, true},
		{"Valid config - disabled", nil, &config.LoggingConfig{FileEnabled: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.IsFileLoggingEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetLogFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       string
	}{
		{"Nil config - use default", nil, nil, DefaultLogFilePath},
		{"Valid config", nil, &config.LoggingConfig{FilePath: "/custom/log/path"}, "/custom/log/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetLogFilePath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetMaxLogFileSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       int64
	}{
		{"Nil config - use default", nil, nil, DefaultMaxFileSize},
		{"Valid config", nil, &config.LoggingConfig{MaxFileSize: 500 * 1024 * 1024}, 500 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetMaxLogFileSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_GetLogBackupCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       int
	}{
		{"Nil config - use default", nil, nil, DefaultBackupCount},
		{"Valid config", nil, &config.LoggingConfig{BackupCount: 15}, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.GetLogBackupCount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_IsConsoleLoggingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expected       bool
	}{
		{"Nil config - use default", nil, nil, true},
		{"Valid config - enabled", nil, &config.LoggingConfig{ConsoleEnabled: true}, true},
		{"Valid config - disabled", nil, &config.LoggingConfig{ConsoleEnabled: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.IsConsoleLoggingEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigAdapter_CreateAuditLoggerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		securityConfig *config.SecurityConfig
		loggingConfig  *config.LoggingConfig
		expectedKeys   []string
	}{
		{
			name:           "Nil configs - use defaults",
			securityConfig: nil,
			loggingConfig:  nil,
			expectedKeys:   []string{"log_directory", "max_file_size", "max_file_age", "rotation_interval", "buffer_size", "enable_file_logging", "enable_console_logging", "log_level"},
		},
		{
			name:           "Valid configs",
			securityConfig: nil,
			loggingConfig: &config.LoggingConfig{
				Level:          "debug",
				FilePath:       "/custom/logs",
				MaxFileSize:    300 * 1024 * 1024,
				FileEnabled:    true,
				ConsoleEnabled: false,
			},
			expectedKeys: []string{"log_directory", "max_file_size", "max_file_age", "rotation_interval", "buffer_size", "enable_file_logging", "enable_console_logging", "log_level"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.CreateAuditLoggerConfig()

			// Check that all expected keys are present
			for _, key := range tt.expectedKeys {
				assert.Contains(t, result, key, "Expected key %s to be present in audit logger config", key)
			}

			// Check specific values based on config
			expectedLogDir := adapter.GetLogFilePath() + "/security"
			expectedMaxSize := adapter.GetMaxLogFileSize()
			assert.Equal(t, expectedLogDir, result["log_directory"])
			assert.Equal(t, expectedMaxSize, result["max_file_size"])
			assert.Equal(t, 30*24*time.Hour, result["max_file_age"])
			assert.Equal(t, 1*time.Hour, result["rotation_interval"])
			assert.Equal(t, 1000, result["buffer_size"])
		})
	}
}

func TestConfigAdapter_CreateRateLimiterConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		securityConfig  *config.SecurityConfig
		loggingConfig   *config.LoggingConfig
		expectedMethods []string
	}{
		{
			name:            "Nil configs - use defaults",
			securityConfig:  nil,
			loggingConfig:   nil,
			expectedMethods: []string{"ping", "get_camera_list", "start_recording", "take_snapshot", "start_streaming", "stop_streaming", "get_stream_url", "get_stream_status", "authenticate"},
		},
		{
			name: "Custom rate limits",
			securityConfig: &config.SecurityConfig{
				RateLimitRequests: 200,
				RateLimitWindow:   2 * time.Minute,
			},
			loggingConfig:   nil,
			expectedMethods: []string{"ping", "get_camera_list", "start_recording", "take_snapshot", "start_streaming", "stop_streaming", "get_stream_url", "get_stream_status", "authenticate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.securityConfig, tt.loggingConfig)
			result := adapter.CreateRateLimiterConfig()

			// Check that all expected methods are present
			for _, method := range tt.expectedMethods {
				assert.Contains(t, result, method, "Expected method %s to be present in rate limiter config", method)
				assert.NotNil(t, result[method], "Rate limit config for method %s should not be nil", method)
			}

			// Check that authenticate has the lowest rate (1% of base rate)
			authenticateConfig := result["authenticate"]
			assert.NotNil(t, authenticateConfig)
			expectedRate := float64(adapter.GetRateLimitRequests()) * 0.01 // 1% of base rate
			assert.Equal(t, expectedRate, authenticateConfig.RequestsPerSecond)
		})
	}
}

// =============================================================================
// INTERFACE COMPLIANCE TEST
// =============================================================================

func TestConfigAdapter_ImplementsSecurityConfigProvider(t *testing.T) {
	t.Parallel()

	// This test ensures ConfigAdapter implements SecurityConfigProvider interface
	var _ SecurityConfigProvider = (*ConfigAdapter)(nil)

	// Test that we can use it as the interface
	adapter := NewConfigAdapter(nil, nil)
	var provider SecurityConfigProvider = adapter

	// Test interface methods
	assert.Equal(t, DefaultRateLimitRequests, provider.GetRateLimitRequests())
	assert.Equal(t, DefaultRateLimitWindow, provider.GetRateLimitWindow())
	assert.Equal(t, "", provider.GetJWTSecretKey())
	assert.Equal(t, DefaultJWTExpiryHours, provider.GetJWTExpiryHours())
	assert.Equal(t, DefaultLogLevel, provider.GetLogLevel())
	assert.Equal(t, DefaultLogFormat, provider.GetLogFormat())
	assert.Equal(t, false, provider.IsFileLoggingEnabled())
	assert.Equal(t, DefaultLogFilePath, provider.GetLogFilePath())
	assert.Equal(t, int64(DefaultMaxFileSize), provider.GetMaxLogFileSize())
	assert.Equal(t, DefaultBackupCount, provider.GetLogBackupCount())
	assert.Equal(t, true, provider.IsConsoleLoggingEnabled())

	// Test config creation methods
	auditConfig := provider.CreateAuditLoggerConfig()
	assert.NotNil(t, auditConfig)
	assert.Contains(t, auditConfig, "log_directory")

	rateLimiterConfig := provider.CreateRateLimiterConfig()
	assert.NotNil(t, rateLimiterConfig)
	assert.Contains(t, rateLimiterConfig, "authenticate")
}
