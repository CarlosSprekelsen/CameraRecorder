package security

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// RATE LIMITER TESTS FOR 90%+ COVERAGE
// =============================================================================

// TestSecurityConfig provides a test implementation of SecurityConfigProvider
type TestSecurityConfig struct{}

func (c *TestSecurityConfig) GetRateLimitRequests() int             { return 100 }
func (c *TestSecurityConfig) GetRateLimitWindow() time.Duration     { return time.Minute }
func (c *TestSecurityConfig) GetJWTSecretKey() string               { return "test_secret" }
func (c *TestSecurityConfig) GetJWTExpiryHours() int                { return 24 }
func (c *TestSecurityConfig) GetLogLevel() string                   { return "info" }
func (c *TestSecurityConfig) GetLogFilePath() string                { return "/tmp/test" }
func (c *TestSecurityConfig) GetMaxLogFileSize() int64              { return 10 }
func (c *TestSecurityConfig) GetMaxLogFileAge() time.Duration       { return 24 * time.Hour }
func (c *TestSecurityConfig) GetLogRotationInterval() time.Duration { return time.Hour }
func (c *TestSecurityConfig) GetLogBackupCount() int                { return 3 }
func (c *TestSecurityConfig) GetLogFormat() string                  { return "json" }
func (c *TestSecurityConfig) GetLogConsoleEnabled() bool            { return true }
func (c *TestSecurityConfig) IsFileLoggingEnabled() bool            { return true }
func (c *TestSecurityConfig) IsConsoleLoggingEnabled() bool         { return true }
func (c *TestSecurityConfig) CreateRateLimiterConfig() map[string]*RateLimitConfig {
	return map[string]*RateLimitConfig{
		"default": DefaultRateLimitConfig(),
	}
}
func (c *TestSecurityConfig) CreateAuditLoggerConfig() map[string]interface{} {
	return map[string]interface{}{
		"log_directory": "/tmp/test/security",
		"max_file_size": 10,
		"max_file_age":  24 * time.Hour,
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	t.Parallel()

	config := DefaultRateLimitConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 100.0, config.RequestsPerSecond)
	assert.Equal(t, 200, config.BurstSize)
	assert.Equal(t, time.Second, config.WindowSize)
}

func TestNewEnhancedRateLimiter(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()

	tests := []struct {
		name   string
		config SecurityConfigProvider
		want   bool
	}{
		{"Nil config", nil, true},
		{"Valid config", &TestSecurityConfig{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewEnhancedRateLimiter(logger, tt.config)
			assert.NotNil(t, limiter)
			assert.NotNil(t, limiter.limits)
			assert.NotNil(t, limiter.clientLimits)
			assert.NotNil(t, limiter.globalLimiter)
			assert.NotNil(t, limiter.logger)
		})
	}
}

func TestEnhancedRateLimiter_SetMethodRateLimit(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)

	// Set method rate limit
	config := &RateLimitConfig{
		RequestsPerSecond: 50.0,
		BurstSize:         100,
		WindowSize:        time.Minute,
	}
	limiter.SetMethodRateLimit("test_method", config)

	// Verify it was set
	limit, exists := limiter.limits["test_method"]
	assert.True(t, exists)
	assert.Equal(t, 50.0, limit.RequestsPerSecond)
	assert.Equal(t, time.Minute, limit.WindowSize)
}

func TestEnhancedRateLimiter_CheckLimit(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)

	// Test basic rate limiting
	method := "test_method"
	clientID := "test_client"

	// First request should succeed
	err := limiter.CheckLimit(method, clientID)
	assert.NoError(t, err, "First request should succeed")

	// Multiple requests should also succeed (within limits)
	for i := 0; i < 5; i++ {
		err := limiter.CheckLimit(method, clientID)
		assert.NoError(t, err, "Request %d should succeed", i+2)
	}
}

func TestEnhancedRateLimiter_ResetClientLimits(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)
	clientID := "test_client"
	method := "test_method"

	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit(method, clientID)
	}

	// Reset limits
	limiter.ResetClientLimits(clientID)

	// Should be able to make requests again
	err := limiter.CheckLimit(method, clientID)
	assert.NoError(t, err, "Should not be rate limited after reset")
}

func TestEnhancedRateLimiter_GetClientStats(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)
	clientID := "test_client"
	method := "test_method"

	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit(method, clientID)
	}

	// Get client stats
	stats := limiter.GetClientStats(clientID)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats["request_count"])
}

func TestEnhancedRateLimiter_GetMethodStats(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)
	method := "test_method"

	// Set method rate limit first
	config := &RateLimitConfig{
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		WindowSize:        time.Minute,
	}
	limiter.SetMethodRateLimit(method, config)

	// Get method stats
	stats := limiter.GetMethodStats(method)
	assert.NotNil(t, stats)
	assert.Equal(t, method, stats["method"])
}

func TestEnhancedRateLimiter_GetGlobalStats(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)

	// Get global stats
	stats := limiter.GetGlobalStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_clients")
	assert.Contains(t, stats, "blocked_clients")
	assert.Contains(t, stats, "configured_methods")
}

func TestEnhancedRateLimiter_CleanupOldClients(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)
	clientID := "test_client"
	method := "test_method"

	// Make a request to create client entry
	limiter.CheckLimit(method, clientID)

	// Wait a bit so the client becomes "old" enough to be cleaned up
	time.Sleep(2 * time.Millisecond)

	// Cleanup old clients (with very short max age)
	limiter.CleanupOldClients(1 * time.Millisecond)

	// Wait a bit for cleanup
	time.Sleep(2 * time.Millisecond)

	// Client should be cleaned up
	stats := limiter.GetClientStats(clientID)
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.False(t, stats["exists"].(bool), "Client should be cleaned up")
}

func TestEnhancedRateLimiter_StartCleanupRoutine(t *testing.T) {
	t.Parallel()

	logger := logging.GetLogger()
	limiter := NewEnhancedRateLimiter(logger, nil)

	// Start cleanup routine with very short interval
	limiter.StartCleanupRoutine(1*time.Millisecond, 1*time.Millisecond)

	// Wait a bit for routine to start
	time.Sleep(5 * time.Millisecond)

	// Stop the routine (this should not panic)
	// Note: The actual implementation might not have a stop method
	// This test mainly ensures StartCleanupRoutine doesn't panic
}
