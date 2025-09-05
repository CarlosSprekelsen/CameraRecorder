//go:build unit
// +build unit

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CONFIG ADAPTER TESTS FOR 90%+ COVERAGE
// =============================================================================

func TestNewConfigAdapter(t *testing.T) {
	t.Parallel()

	// Test with nil config
	adapter := NewConfigAdapter(nil)
	assert.NotNil(t, adapter)
	assert.Nil(t, adapter.config)

	// Test with mock config
	mockConfig := &MockSecurityConfig{}
	adapter = NewConfigAdapter(mockConfig)
	assert.NotNil(t, adapter)
	assert.Equal(t, mockConfig, adapter.config)
}

func TestConfigAdapter_GetSecurityConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   SecurityConfig
	}{
		{"Nil config", nil, nil},
		{"Mock config", &MockSecurityConfig{}, &MockSecurityConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetSecurityConfig()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetLoggingConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   LoggingConfig
	}{
		{"Nil config", nil, LoggingConfig{}},
		{"Mock config", &MockSecurityConfig{}, LoggingConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetLoggingConfig()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetRateLimitRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   int
	}{
		{"Nil config", nil, 100}, // Default value
		{"Mock config", &MockSecurityConfig{}, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetRateLimitRequests()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetRateLimitWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   string
	}{
		{"Nil config", nil, "1m"}, // Default value
		{"Mock config", &MockSecurityConfig{}, "1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetRateLimitWindow()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetJWTSecretKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   string
	}{
		{"Nil config", nil, "test_secret_key"}, // Default value
		{"Mock config", &MockSecurityConfig{}, "test_secret_key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetJWTSecretKey()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetJWTExpiryHours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   int
	}{
		{"Nil config", nil, 24}, // Default value
		{"Mock config", &MockSecurityConfig{}, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetJWTExpiryHours()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   string
	}{
		{"Nil config", nil, "info"}, // Default value
		{"Mock config", &MockSecurityConfig{}, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetLogLevel()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetLogFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   string
	}{
		{"Nil config", nil, "json"}, // Default value
		{"Mock config", &MockSecurityConfig{}, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetLogFormat()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_IsFileLoggingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   bool
	}{
		{"Nil config", nil, false}, // Default value
		{"Mock config", &MockSecurityConfig{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.IsFileLoggingEnabled()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetLogFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   string
	}{
		{"Nil config", nil, "/tmp/security.log"}, // Default value
		{"Mock config", &MockSecurityConfig{}, "/tmp/security.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetLogFilePath()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetMaxLogFileSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   int
	}{
		{"Nil config", nil, 100}, // Default value
		{"Mock config", &MockSecurityConfig{}, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetMaxLogFileSize()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_GetLogBackupCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   int
	}{
		{"Nil config", nil, 3}, // Default value
		{"Mock config", &MockSecurityConfig{}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.GetLogBackupCount()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_IsConsoleLoggingEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   bool
	}{
		{"Nil config", nil, true}, // Default value
		{"Mock config", &MockSecurityConfig{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.IsConsoleLoggingEnabled()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_CreateAuditLoggerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   AuditLoggerConfig
	}{
		{"Nil config", nil, AuditLoggerConfig{
			Level:      "info",
			Format:     "json",
			FileEnabled: false,
			FilePath:   "/tmp/security.log",
			MaxSize:    100,
			BackupCount: 3,
			ConsoleEnabled: true,
		}},
		{"Mock config", &MockSecurityConfig{}, AuditLoggerConfig{
			Level:      "info",
			Format:     "json",
			FileEnabled: false,
			FilePath:   "/tmp/security.log",
			MaxSize:    100,
			BackupCount: 3,
			ConsoleEnabled: true,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.CreateAuditLoggerConfig()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestConfigAdapter_CreateRateLimiterConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config SecurityConfig
		want   RateLimiterConfig
	}{
		{"Nil config", nil, RateLimiterConfig{
			Requests: 100,
			Window:   "1m",
		}},
		{"Mock config", &MockSecurityConfig{}, RateLimiterConfig{
			Requests: 100,
			Window:   "1m",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewConfigAdapter(tt.config)
			result := adapter.CreateRateLimiterConfig()
			assert.Equal(t, tt.want, result)
		})
	}
}

// =============================================================================
// MOCK SECURITY CONFIG FOR TESTING
// =============================================================================

type MockSecurityConfig struct{}

func (m *MockSecurityConfig) GetRateLimitRequests() int {
	return 100
}

func (m *MockSecurityConfig) GetRateLimitWindow() string {
	return "1m"
}

func (m *MockSecurityConfig) GetJWTSecretKey() string {
	return "test_secret_key"
}

func (m *MockSecurityConfig) GetJWTExpiryHours() int {
	return 24
}

func (m *MockSecurityConfig) GetLogLevel() string {
	return "info"
}

func (m *MockSecurityConfig) GetLogFormat() string {
	return "json"
}

func (m *MockSecurityConfig) IsFileLoggingEnabled() bool {
	return false
}

func (m *MockSecurityConfig) GetLogFilePath() string {
	return "/tmp/security.log"
}

func (m *MockSecurityConfig) GetMaxLogFileSize() int {
	return 100
}

func (m *MockSecurityConfig) GetLogBackupCount() int {
	return 3
}

func (m *MockSecurityConfig) IsConsoleLoggingEnabled() bool {
	return true
}

