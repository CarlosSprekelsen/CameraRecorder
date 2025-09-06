package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"golang.org/x/time/rate"
)

// RateLimitConfig defines rate limiting configuration for a method
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	WindowSize        time.Duration
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 100.0,
		BurstSize:         200,
		WindowSize:        time.Second,
	}
}

// MethodRateLimit defines rate limits for specific methods
type MethodRateLimit struct {
	Method string
	Config *RateLimitConfig
}

// ClientRateLimit tracks rate limiting for a specific client
type ClientRateLimit struct {
	Limiter      *rate.Limiter
	LastAccess   time.Time
	RequestCount int64
	BlockedCount int64
}

// EnhancedRateLimiter provides enhanced rate limiting with per-method limits and DDoS protection
type EnhancedRateLimiter struct {
	limits        map[string]*RateLimitConfig
	clientLimits  map[string]*ClientRateLimit
	globalLimiter *rate.Limiter
	mutex         sync.RWMutex
	logger        *logging.Logger
	config        interface{} // Will be typed based on existing config structure

	// DDoS protection
	maxRequestsPerMinute int
	blockedClients       map[string]time.Time
	blockDuration        time.Duration
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(logger *logging.Logger, config interface{}) *EnhancedRateLimiter {
	limiter := &EnhancedRateLimiter{
		limits:               make(map[string]*RateLimitConfig),
		clientLimits:         make(map[string]*ClientRateLimit),
		globalLimiter:        rate.NewLimiter(rate.Every(time.Second), 1000), // Global limit: 1000 req/sec
		logger:               logger,
		config:               config,
		maxRequestsPerMinute: 600, // 600 requests per minute per client
		blockedClients:       make(map[string]time.Time),
		blockDuration:        5 * time.Minute, // Block for 5 minutes
	}

	// Set default rate limits from configuration if available
	if config != nil {
		// Try to use config adapter if available
		if adapter, ok := config.(*ConfigAdapter); ok {
			limiter.setConfigBasedLimits(adapter)
		} else {
			// Fall back to default limits
			limiter.setDefaultLimits()
		}
	} else {
		// No config provided, use defaults
		limiter.setDefaultLimits()
	}

	return limiter
}

// setConfigBasedLimits sets rate limits from configuration adapter
func (erl *EnhancedRateLimiter) setConfigBasedLimits(adapter *ConfigAdapter) {
	// Get rate limits from existing configuration
	configLimits := adapter.CreateRateLimiterConfig()

	// Apply configured limits
	for method, config := range configLimits {
		erl.limits[method] = config
		erl.logger.WithFields(logging.Fields{
			"method":              method,
			"requests_per_second": config.RequestsPerSecond,
			"burst_size":          config.BurstSize,
			"action":              "config_rate_limit_applied",
			"component":           "security_rate_limiter",
		}).Info("Configuration-based rate limit applied")
	}
}

// setDefaultLimits sets default rate limits for common methods
func (erl *EnhancedRateLimiter) setDefaultLimits() {
	defaultLimits := map[string]*RateLimitConfig{
		"ping": {
			RequestsPerSecond: 10.0, // 10 pings per second
			BurstSize:         20,
			WindowSize:        time.Second,
		},
		"get_camera_list": {
			RequestsPerSecond: 5.0, // 5 requests per second
			BurstSize:         10,
			WindowSize:        time.Second,
		},
		"start_recording": {
			RequestsPerSecond: 2.0, // 2 recordings per second
			BurstSize:         5,
			WindowSize:        time.Second,
		},
		"take_snapshot": {
			RequestsPerSecond: 3.0, // 3 snapshots per second
			BurstSize:         6,
			WindowSize:        time.Second,
		},
		"start_streaming": {
			RequestsPerSecond: 2.0, // 2 streaming starts per second
			BurstSize:         5,
			WindowSize:        time.Second,
		},
		"stop_streaming": {
			RequestsPerSecond: 2.0, // 2 streaming stops per second
			BurstSize:         5,
			WindowSize:        time.Second,
		},
		"get_stream_url": {
			RequestsPerSecond: 10.0, // 10 stream URL requests per second
			BurstSize:         20,
			WindowSize:        time.Second,
		},
		"get_stream_status": {
			RequestsPerSecond: 10.0, // 10 stream status requests per second
			BurstSize:         20,
			WindowSize:        time.Second,
		},
		"authenticate": {
			RequestsPerSecond: 1.0, // 1 authentication per second (prevent brute force)
			BurstSize:         3,
			WindowSize:        time.Second,
		},
	}

	for method, config := range defaultLimits {
		erl.limits[method] = config
	}
}

// SetMethodRateLimit sets a custom rate limit for a specific method
func (erl *EnhancedRateLimiter) SetMethodRateLimit(method string, config *RateLimitConfig) {
	erl.mutex.Lock()
	defer erl.mutex.Unlock()

	erl.limits[method] = config
	erl.logger.WithFields(logging.Fields{
		"method":              method,
		"requests_per_second": config.RequestsPerSecond,
		"burst_size":          config.BurstSize,
		"action":              "rate_limit_set",
	}).Info("Method rate limit configured")
}

// CheckLimit checks if a client has exceeded rate limits for a method
func (erl *EnhancedRateLimiter) CheckLimit(method, clientID string) error {
	erl.mutex.Lock()
	defer erl.mutex.Unlock()

	// Check if client is blocked
	if blockTime, blocked := erl.blockedClients[clientID]; blocked {
		if time.Since(blockTime) < erl.blockDuration {
			erl.logger.WithFields(logging.Fields{
				"client_id": clientID,
				"method":    method,
				"action":    "rate_limit_blocked",
			}).Warn("Client blocked due to rate limit violations")
			return fmt.Errorf("client blocked due to rate limit violations")
		}
		// Unblock client after block duration
		delete(erl.blockedClients, clientID)
	}

	// Get or create client rate limiter
	clientLimit, exists := erl.clientLimits[clientID]
	if !exists {
		clientLimit = &ClientRateLimit{
			Limiter:      rate.NewLimiter(rate.Every(time.Second), 100), // Default: 100 req/sec
			LastAccess:   time.Now(),
			RequestCount: 0,
			BlockedCount: 0,
		}
		erl.clientLimits[clientID] = clientLimit
	}

	// Update client access time and request count
	clientLimit.LastAccess = time.Now()
	clientLimit.RequestCount++

	// Check global rate limit
	if !erl.globalLimiter.Allow() {
		erl.logger.WithFields(logging.Fields{
			"client_id": clientID,
			"method":    method,
			"action":    "global_rate_limit_exceeded",
		}).Warn("Global rate limit exceeded")
		return fmt.Errorf("global rate limit exceeded")
	}

	// Check method-specific rate limit
	if methodConfig, exists := erl.limits[method]; exists {
		// Create method-specific limiter for this client
		methodLimiter := rate.NewLimiter(rate.Every(time.Duration(float64(time.Second)/methodConfig.RequestsPerSecond)), methodConfig.BurstSize)

		if !methodLimiter.Allow() {
			clientLimit.BlockedCount++

			erl.logger.WithFields(logging.Fields{
				"client_id": clientID,
				"method":    method,
				"action":    "method_rate_limit_exceeded",
				"limit":     methodConfig.RequestsPerSecond,
			}).Warn("Method rate limit exceeded")

			// Check if client should be blocked
			if clientLimit.BlockedCount >= 10 { // Block after 10 violations
				erl.blockedClients[clientID] = time.Now()
				erl.logger.WithFields(logging.Fields{
					"client_id": clientID,
					"method":    method,
					"action":    "client_blocked",
					"duration":  erl.blockDuration,
				}).Warn("Client blocked due to repeated rate limit violations")
			}

			return fmt.Errorf("rate limit exceeded for method %s", method)
		}
	}

	// Check per-client rate limit (requests per minute)
	if clientLimit.RequestCount > int64(erl.maxRequestsPerMinute) {
		clientLimit.BlockedCount++

		erl.logger.WithFields(logging.Fields{
			"client_id": clientID,
			"method":    method,
			"action":    "client_rate_limit_exceeded",
			"requests":  clientLimit.RequestCount,
			"limit":     erl.maxRequestsPerMinute,
		}).Warn("Client rate limit exceeded")

		// Block client if they exceed limits repeatedly
		if clientLimit.BlockedCount >= 5 {
			erl.blockedClients[clientID] = time.Now()
			erl.logger.WithFields(logging.Fields{
				"client_id": clientID,
				"method":    method,
				"action":    "client_blocked",
				"duration":  erl.blockDuration,
			}).Warn("Client blocked due to excessive requests")
		}

		return fmt.Errorf("client rate limit exceeded")
	}

	return nil
}

// ResetClientLimits resets rate limiting for a specific client
func (erl *EnhancedRateLimiter) ResetClientLimits(clientID string) {
	erl.mutex.Lock()
	defer erl.mutex.Unlock()

	delete(erl.clientLimits, clientID)
	delete(erl.blockedClients, clientID)

	erl.logger.WithFields(logging.Fields{
		"client_id": clientID,
		"action":    "rate_limit_reset",
	}).Info("Client rate limits reset")
}

// GetClientStats returns rate limiting statistics for a client
func (erl *EnhancedRateLimiter) GetClientStats(clientID string) map[string]interface{} {
	erl.mutex.RLock()
	defer erl.mutex.RUnlock()

	clientLimit, exists := erl.clientLimits[clientID]
	if !exists {
		return map[string]interface{}{
			"client_id": clientID,
			"exists":    false,
		}
	}

	_, blocked := erl.blockedClients[clientID]

	return map[string]interface{}{
		"client_id":         clientID,
		"exists":            true,
		"request_count":     clientLimit.RequestCount,
		"blocked_count":     clientLimit.BlockedCount,
		"last_access":       clientLimit.LastAccess,
		"currently_blocked": blocked,
		"block_duration":    erl.blockDuration,
	}
}

// GetMethodStats returns rate limiting statistics for a method
func (erl *EnhancedRateLimiter) GetMethodStats(method string) map[string]interface{} {
	erl.mutex.RLock()
	defer erl.mutex.RUnlock()

	config, exists := erl.limits[method]
	if !exists {
		return map[string]interface{}{
			"method": method,
			"exists": false,
		}
	}

	return map[string]interface{}{
		"method":              method,
		"exists":              true,
		"requests_per_second": config.RequestsPerSecond,
		"burst_size":          config.BurstSize,
		"window_size":         config.WindowSize,
	}
}

// GetGlobalStats returns global rate limiting statistics
func (erl *EnhancedRateLimiter) GetGlobalStats() map[string]interface{} {
	erl.mutex.RLock()
	defer erl.mutex.RUnlock()

	return map[string]interface{}{
		"total_clients":           len(erl.clientLimits),
		"blocked_clients":         len(erl.blockedClients),
		"configured_methods":      len(erl.limits),
		"max_requests_per_minute": erl.maxRequestsPerMinute,
		"block_duration":          erl.blockDuration,
	}
}

// CleanupOldClients removes old client rate limit entries
func (erl *EnhancedRateLimiter) CleanupOldClients(maxAge time.Duration) {
	erl.mutex.Lock()
	defer erl.mutex.Unlock()

	now := time.Now()
	removed := 0

	for clientID, clientLimit := range erl.clientLimits {
		if now.Sub(clientLimit.LastAccess) > maxAge {
			delete(erl.clientLimits, clientID)
			removed++
		}
	}

	// Clean up blocked clients
	for clientID, blockTime := range erl.blockedClients {
		if now.Sub(blockTime) > erl.blockDuration {
			delete(erl.blockedClients, clientID)
		}
	}

	if removed > 0 {
		erl.logger.WithFields(logging.Fields{
			"removed_clients": removed,
			"action":          "cleanup_completed",
		}).Info("Old client rate limit entries cleaned up")
	}
}

// StartCleanupRoutine starts a background routine to clean up old client entries
func (erl *EnhancedRateLimiter) StartCleanupRoutine(interval, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			erl.CleanupOldClients(maxAge)
		}
	}()
}
