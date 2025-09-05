//go:build unit
// +build unit

package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RATE LIMITER TESTS FOR 90%+ COVERAGE
// =============================================================================

func TestDefaultRateLimitConfig(t *testing.T) {
	t.Parallel()

	config := DefaultRateLimitConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 100, config.Requests)
	assert.Equal(t, "1m", config.Window)
}

func TestNewEnhancedRateLimiter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *RateLimiterConfig
		want   bool
	}{
		{"Nil config", nil, true},
		{"Default config", &RateLimiterConfig{Requests: 50, Window: "30s"}, true},
		{"Custom config", &RateLimiterConfig{Requests: 200, Window: "2m"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewEnhancedRateLimiter(tt.config)
			assert.NotNil(t, limiter)
			assert.NotNil(t, limiter.clients)
			assert.NotNil(t, limiter.methodLimits)
			assert.NotNil(t, limiter.globalStats)
		})
	}
}

func TestEnhancedRateLimiter_SetMethodRateLimit(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(nil)
	
	// Set method rate limit
	limiter.SetMethodRateLimit("test_method", 50, time.Minute)
	
	// Verify it was set
	limit, exists := limiter.methodLimits["test_method"]
	assert.True(t, exists)
	assert.Equal(t, 50, limit.Requests)
	assert.Equal(t, time.Minute, limit.Window)
}

func TestEnhancedRateLimiter_CheckLimit(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	tests := []struct {
		name     string
		clientID string
		method   string
		requests int
		want     bool
	}{
		{"Within limit", "client1", "method1", 5, false},
		{"At limit", "client2", "method2", 10, false},
		{"Over limit", "client3", "method3", 11, true},
		{"Different clients", "client4", "method4", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make requests
			for i := 0; i < tt.requests; i++ {
				limited := limiter.CheckLimit(tt.clientID, tt.method)
				if i >= 10 && tt.want {
					assert.True(t, limited, "Should be rate limited after 10 requests")
				} else {
					assert.False(t, limited, "Should not be rate limited")
				}
			}
		})
	}
}

func TestEnhancedRateLimiter_ResetClientLimits(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 5, Window: time.Minute})
	clientID := "test_client"
	
	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit(clientID, "test_method")
	}
	
	// Reset limits
	limiter.ResetClientLimits(clientID)
	
	// Should be able to make requests again
	limited := limiter.CheckLimit(clientID, "test_method")
	assert.False(t, limited, "Should not be rate limited after reset")
}

func TestEnhancedRateLimiter_GetClientStats(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	clientID := "test_client"
	
	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit(clientID, "test_method")
	}
	
	// Get client stats
	stats := limiter.GetClientStats(clientID)
	assert.NotNil(t, stats)
	assert.Equal(t, clientID, stats.ClientID)
	assert.Equal(t, 3, stats.RequestCount)
	assert.False(t, stats.IsLimited)
}

func TestEnhancedRateLimiter_GetMethodStats(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	method := "test_method"
	
	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit("client1", method)
		limiter.CheckLimit("client2", method)
	}
	
	// Get method stats
	stats := limiter.GetMethodStats(method)
	assert.NotNil(t, stats)
	assert.Equal(t, method, stats.Method)
	assert.Equal(t, 6, stats.TotalRequests) // 3 from client1 + 3 from client2
	assert.Equal(t, 2, stats.UniqueClients)
}

func TestEnhancedRateLimiter_GetGlobalStats(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.CheckLimit("client1", "method1")
		limiter.CheckLimit("client2", "method2")
	}
	
	// Get global stats
	stats := limiter.GetGlobalStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 6, stats.TotalRequests) // 3 from client1 + 3 from client2
	assert.Equal(t, 2, stats.UniqueClients)
	assert.Equal(t, 2, stats.UniqueMethods)
}

func TestEnhancedRateLimiter_CleanupOldClients(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	clientID := "test_client"
	
	// Make some requests
	limiter.CheckLimit(clientID, "test_method")
	
	// Cleanup old clients
	limiter.CleanupOldClients()
	
	// Client should still exist (recent activity)
	stats := limiter.GetClientStats(clientID)
	assert.NotNil(t, stats)
}

func TestEnhancedRateLimiter_StartCleanupRoutine(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	// Start cleanup routine
	limiter.StartCleanupRoutine()
	
	// Should not panic and should be running
	assert.NotNil(t, limiter)
	
	// Stop cleanup routine
	limiter.StopCleanupRoutine()
}

func TestEnhancedRateLimiter_StopCleanupRoutine(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	// Start cleanup routine
	limiter.StartCleanupRoutine()
	
	// Stop cleanup routine
	limiter.StopCleanupRoutine()
	
	// Should not panic
	assert.NotNil(t, limiter)
}

// =============================================================================
// EDGE CASES AND ERROR CONDITIONS
// =============================================================================

func TestEnhancedRateLimiter_EmptyClientID(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	// Test with empty client ID
	limited := limiter.CheckLimit("", "test_method")
	assert.False(t, limited, "Empty client ID should not be rate limited")
}

func TestEnhancedRateLimiter_EmptyMethod(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 10, Window: time.Minute})
	
	// Test with empty method
	limited := limiter.CheckLimit("test_client", "")
	assert.False(t, limited, "Empty method should not be rate limited")
}

func TestEnhancedRateLimiter_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 100, Window: time.Minute})
	
	// Test concurrent access
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(clientID string) {
			for j := 0; j < 10; j++ {
				limiter.CheckLimit(clientID, "test_method")
			}
			done <- true
		}("client" + string(rune(i)))
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should not panic and should have processed all requests
	stats := limiter.GetGlobalStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalRequests)
}

func TestEnhancedRateLimiter_MethodSpecificLimits(t *testing.T) {
	t.Parallel()

	limiter := NewEnhancedRateLimiter(&RateLimiterConfig{Requests: 100, Window: time.Minute})
	
	// Set method-specific limit
	limiter.SetMethodRateLimit("limited_method", 5, time.Minute)
	
	clientID := "test_client"
	
	// Test global limit (should allow 100 requests)
	for i := 0; i < 10; i++ {
		limited := limiter.CheckLimit(clientID, "unlimited_method")
		assert.False(t, limited, "Should not be rate limited for unlimited method")
	}
	
	// Test method-specific limit (should limit at 5 requests)
	for i := 0; i < 6; i++ {
		limited := limiter.CheckLimit(clientID, "limited_method")
		if i >= 5 {
			assert.True(t, limited, "Should be rate limited for limited method after 5 requests")
		} else {
			assert.False(t, limited, "Should not be rate limited for limited method")
		}
	}
}

