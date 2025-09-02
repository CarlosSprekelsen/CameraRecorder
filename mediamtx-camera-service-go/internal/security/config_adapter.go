package security

import (
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

// ConfigAdapter bridges existing configuration with security middleware
type ConfigAdapter struct {
	securityConfig *config.SecurityConfig
	loggingConfig  *config.LoggingConfig
}

// NewConfigAdapter creates a new configuration adapter
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

// GetRateLimitRequests returns rate limit requests from existing config
func (ca *ConfigAdapter) GetRateLimitRequests() int {
	if ca.securityConfig != nil {
		return ca.securityConfig.RateLimitRequests
	}
	return 100 // Default fallback
}

// GetRateLimitWindow returns rate limit window from existing config
func (ca *ConfigAdapter) GetRateLimitWindow() time.Duration {
	if ca.securityConfig != nil {
		return ca.securityConfig.RateLimitWindow
	}
	return time.Minute // Default fallback
}

// GetJWTSecretKey returns JWT secret key from existing config
func (ca *ConfigAdapter) GetJWTSecretKey() string {
	if ca.securityConfig != nil {
		return ca.securityConfig.JWTSecretKey
	}
	return "" // Default fallback
}

// GetJWTExpiryHours returns JWT expiry hours from existing config
func (ca *ConfigAdapter) GetJWTExpiryHours() int {
	if ca.securityConfig != nil {
		return ca.securityConfig.JWTExpiryHours
	}
	return 24 // Default fallback
}

// GetLogLevel returns log level from existing config
func (ca *ConfigAdapter) GetLogLevel() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.Level
	}
	return "info" // Default fallback
}

// GetLogFormat returns log format from existing config
func (ca *ConfigAdapter) GetLogFormat() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.Format
	}
	return "json" // Default fallback
}

// IsFileLoggingEnabled returns whether file logging is enabled
func (ca *ConfigAdapter) IsFileLoggingEnabled() bool {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.FileEnabled
	}
	return false // Default fallback
}

// GetLogFilePath returns log file path from existing config
func (ca *ConfigAdapter) GetLogFilePath() string {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.FilePath
	}
	return "/var/log/camera-service" // Default fallback
}

// GetMaxLogFileSize returns max log file size from existing config
func (ca *ConfigAdapter) GetMaxLogFileSize() int64 {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.MaxFileSize
	}
	return 100 * 1024 * 1024 // 100 MB default fallback
}

// GetLogBackupCount returns log backup count from existing config
func (ca *ConfigAdapter) GetLogBackupCount() int {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.BackupCount
	}
	return 5 // Default fallback
}

// IsConsoleLoggingEnabled returns whether console logging is enabled
func (ca *ConfigAdapter) IsConsoleLoggingEnabled() bool {
	if ca.loggingConfig != nil {
		return ca.loggingConfig.ConsoleEnabled
	}
	return true // Default fallback
}

// CreateAuditLoggerConfig creates audit logger config from existing config
func (ca *ConfigAdapter) CreateAuditLoggerConfig() *AuditLoggerConfig {
	return &AuditLoggerConfig{
		LogDirectory:         ca.GetLogFilePath() + "/security",
		MaxFileSize:          ca.GetMaxLogFileSize(),
		MaxFileAge:           30 * 24 * time.Hour, // 30 days
		RotationInterval:     1 * time.Hour,
		BufferSize:           1000,
		EnableFileLogging:    ca.IsFileLoggingEnabled(),
		EnableConsoleLogging: ca.IsConsoleLoggingEnabled(),
		LogLevel:             ca.GetLogLevel(),
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
		"authenticate": {
			RequestsPerSecond: baseRate * 0.01, // 1% of base rate (prevent brute force)
			BurstSize:         int(baseRate * 0.03),
			WindowSize:        window,
		},
	}
}
