package security

import (
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

// SecurityConfigProvider defines the interface for accessing security configuration
// This interface provides type-safe access to security configuration values
// and eliminates the need for interface{} usage in security components.
type SecurityConfigProvider interface {
	// Rate limiting configuration
	GetRateLimitRequests() int
	GetRateLimitWindow() time.Duration

	// JWT configuration
	GetJWTSecretKey() string
	GetJWTExpiryHours() int

	// Logging configuration
	GetLogLevel() string
	GetLogFormat() string
	IsFileLoggingEnabled() bool
	GetLogFilePath() string
	GetMaxLogFileSize() int64
	GetLogBackupCount() int
	IsConsoleLoggingEnabled() bool

	// Rate limiter configuration creation
	CreateRateLimiterConfig() map[string]*RateLimitConfig

	// Audit logger configuration creation
	CreateAuditLoggerConfig() map[string]interface{}
}

// ConfigAdapter bridges existing configuration with security middleware.
// It implements SecurityConfigProvider interface to provide type-safe access
// to security configuration values and eliminates hardcoded defaults.
type ConfigAdapter struct {
	securityConfig *config.SecurityConfig
	loggingConfig  *config.LoggingConfig
}

// Ensure ConfigAdapter implements SecurityConfigProvider interface
var _ SecurityConfigProvider = (*ConfigAdapter)(nil)

// Security configuration constants - centralized defaults
const (
	// Default rate limiting values
	DefaultRateLimitRequests = 100
	DefaultRateLimitWindow   = time.Minute

	// Default JWT configuration
	DefaultJWTExpiryHours = 24

	// Default logging configuration
	DefaultLogLevel    = "info"
	DefaultLogFormat   = "json"
	DefaultLogFilePath = "/var/log/camera-service"
	DefaultMaxFileSize = 100 * 1024 * 1024 // 100 MB
	DefaultBackupCount = 5
)

// NewConfigAdapter creates a new configuration adapter with centralized defaults.
// This constructor ensures that all security components receive consistent
// configuration values and eliminates hardcoded defaults throughout the module.
func NewConfigAdapter(securityConfig *config.SecurityConfig, loggingConfig *config.LoggingConfig) *ConfigAdapter {
	return &ConfigAdapter{
		securityConfig: securityConfig,
		loggingConfig:  loggingConfig,
	}
}

// GetSecurityConfig returns the security configuration
func (ca *ConfigAdapter) GetSecurityConfig() *config.SecurityConfig {
	return ca.securityConfig
}

// GetLoggingConfig returns the logging configuration
func (ca *ConfigAdapter) GetLoggingConfig() *config.LoggingConfig {
	return ca.loggingConfig
}

// GetRateLimitRequests returns rate limit requests from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetRateLimitRequests() int {
	if ca.securityConfig != nil {
		return ca.securityConfig.RateLimitRequests
	}
	return DefaultRateLimitRequests
}

// GetRateLimitWindow returns rate limit window from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetRateLimitWindow() time.Duration {
	if ca.securityConfig != nil {
		return ca.securityConfig.RateLimitWindow
	}
	return DefaultRateLimitWindow
}

// GetJWTSecretKey returns JWT secret key from existing config.
// Returns empty string if configuration is not available (this should be validated elsewhere).
func (ca *ConfigAdapter) GetJWTSecretKey() string {
	if ca.securityConfig != nil {
		return ca.securityConfig.JWTSecretKey
	}
	return "" // No default for security keys - must be explicitly configured
}

// GetJWTExpiryHours returns JWT expiry hours from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetJWTExpiryHours() int {
	if ca.securityConfig != nil {
		return ca.securityConfig.JWTExpiryHours
	}
	return DefaultJWTExpiryHours
}

// GetLogLevel returns log level from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetLogLevel() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.Level
	}
	return DefaultLogLevel
}

// GetLogFormat returns log format from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetLogFormat() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.Format
	}
	return DefaultLogFormat
}

// IsFileLoggingEnabled returns whether file logging is enabled.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) IsFileLoggingEnabled() bool {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.FileEnabled
	}
	return false // Default to console-only logging for security
}

// GetLogFilePath returns log file path from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetLogFilePath() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.FilePath
	}
	return DefaultLogFilePath
}

// GetMaxLogFileSize returns max log file size from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetMaxLogFileSize() int64 {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.MaxFileSize
	}
	return DefaultMaxFileSize
}

// GetLogBackupCount returns log backup count from existing config.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) GetLogBackupCount() int {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.BackupCount
	}
	return DefaultBackupCount
}

// IsConsoleLoggingEnabled returns whether console logging is enabled.
// Uses centralized default if configuration is not available.
func (ca *ConfigAdapter) IsConsoleLoggingEnabled() bool {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.ConsoleEnabled
	}
	return true // Default to console logging for security visibility
}

// CreateAuditLoggerConfig creates audit logger config from existing config
// Note: This now returns a map since we're using the existing logging module
func (ca *ConfigAdapter) CreateAuditLoggerConfig() map[string]interface{} {
	return map[string]interface{}{
		"log_directory":          ca.GetLogFilePath() + "/security",
		"max_file_size":          ca.GetMaxLogFileSize(),
		"max_file_age":           30 * 24 * time.Hour, // 30 days
		"rotation_interval":      1 * time.Hour,
		"buffer_size":            1000, // Default buffer size
		"enable_file_logging":    ca.IsFileLoggingEnabled(),
		"enable_console_logging": ca.IsConsoleLoggingEnabled(),
		"log_level":              ca.GetLogLevel(),
	}
}

// CreateRateLimiterConfig creates rate limiter config from existing config
func (ca *ConfigAdapter) CreateRateLimiterConfig() map[string]*RateLimitConfig {
	// Use existing config values or defaults
	baseRate := float64(ca.GetRateLimitRequests())
	window := ca.GetRateLimitWindow()

	return map[string]*RateLimitConfig{
		"ping": {
			RequestsPerSecond: baseRate * 0.1, // 10% of base rate
			BurstSize:         int(baseRate * 0.2),
			WindowSize:        window,
		},
		"get_camera_list": {
			RequestsPerSecond: baseRate * 0.05, // 5% of base rate
			BurstSize:         int(baseRate * 0.1),
			WindowSize:        window,
		},
		"start_recording": {
			RequestsPerSecond: baseRate * 0.02, // 2% of base rate
			BurstSize:         int(baseRate * 0.05),
			WindowSize:        window,
		},
		"take_snapshot": {
			RequestsPerSecond: baseRate * 0.03, // 3% of base rate
			BurstSize:         int(baseRate * 0.06),
			WindowSize:        window,
		},
		"start_streaming": {
			RequestsPerSecond: baseRate * 0.02, // 2% of base rate
			BurstSize:         int(baseRate * 0.05),
			WindowSize:        window,
		},
		"stop_streaming": {
			RequestsPerSecond: baseRate * 0.02, // 2% of base rate
			BurstSize:         int(baseRate * 0.05),
			WindowSize:        window,
		},
		"get_stream_url": {
			RequestsPerSecond: baseRate * 0.1, // 10% of base rate
			BurstSize:         int(baseRate * 0.2),
			WindowSize:        window,
		},
		"get_stream_status": {
			RequestsPerSecond: baseRate * 0.1, // 10% of base rate
			BurstSize:         int(baseRate * 0.2),
			WindowSize:        window,
		},
		"authenticate": {
			RequestsPerSecond: baseRate * 0.01, // 1% of base rate (prevent brute force)
			BurstSize:         int(baseRate * 0.03),
			WindowSize:        window,
		},
	}
}
