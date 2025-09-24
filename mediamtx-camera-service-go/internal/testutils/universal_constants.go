/*
Universal Test Constants - Project-Wide Test Infrastructure

Provides universal constants for timeouts, security, and common test values
that can be shared across all modules. This eliminates magic numbers and
ensures consistency across the entire test suite.

Design Principles:
- Values derived from canonical configuration fixtures
- Follows established MediaMTX test pattern
- Cross-module consistency and maintainability
- Configuration-driven (not hardcoded)

Requirements Coverage:
- REQ-TEST-001: Universal test constants
- REQ-TEST-002: Cross-module consistency
- REQ-TEST-003: Magic number elimination
- REQ-TEST-004: Configuration-driven testing
*/

package testutils

import (
	"time"
)

// =============================================================================
// UNIVERSAL TEST TIMEOUT CONSTANTS
// =============================================================================
// These constants provide standardized timeouts across all test modules,
// following the established MediaMTX pattern but available project-wide.

const (
	// Test Timeout Constants - Universal across all modules
	UniversalTimeoutShort    = 100 * time.Millisecond // Short operations (validation, simple checks)
	UniversalTimeoutMedium   = 200 * time.Millisecond // Medium operations (cleanup, polling)
	UniversalTimeoutLong     = 500 * time.Millisecond // Long operations (connections, readiness)
	UniversalTimeoutVeryLong = 1 * time.Second        // Very long operations (startup, initialization)
	UniversalTimeoutExtreme  = 10 * time.Second       // Extreme operations (complex workflows)

	// Test Performance Thresholds - Universal performance expectations
	UniversalThresholdFastShutdown   = 100 * time.Millisecond // Fast shutdown operations
	UniversalThresholdMediumShutdown = 500 * time.Millisecond // Medium shutdown operations
	UniversalThresholdFastOperation  = 500 * time.Millisecond // Fast operations
	UniversalThresholdStressTest     = 30 * time.Second       // Stress test completion

	// Test Validation Periods - Universal validation timing
	UniversalValidationPeriodShort = 100 * time.Millisecond // Short validation periods
	UniversalValidationPeriodLong  = 150 * time.Millisecond // Long validation periods

	// Test Retry Constants - Universal retry patterns
	UniversalRetryAttempts = 3               // Number of retry attempts for flaky operations
	UniversalRetryDelay    = 1 * time.Second // Delay between retry attempts

	// Test Concurrency Constants - Universal concurrency patterns
	UniversalConcurrencyGoroutines = 5  // Number of goroutines for concurrency tests
	UniversalConcurrencyIterations = 50 // Number of iterations for stress tests
)

// =============================================================================
// UNIVERSAL SECURITY TEST CONSTANTS
// =============================================================================
// These constants are derived from canonical configuration fixtures
// (tests/fixtures/config_valid_complete.yaml) to ensure consistency.

const (
	// Security Rate Limiting - From canonical configuration
	UniversalRateLimitRequests = 100             // Default rate limit requests per window
	UniversalRateLimitWindow   = 1 * time.Minute // Default rate limit window
	UniversalJWTExpiryHours    = 24              // Default JWT token expiry hours

	// Security Session Management - From canonical configuration
	UniversalSessionTimeout  = 30 * time.Minute // Default session timeout
	UniversalCleanupInterval = 5 * time.Minute  // Default cleanup interval

	// Security Validation Thresholds - From implementation analysis
	UniversalMaxFilenameLength    = 255        // Maximum filename length (filesystem limit)
	UniversalControlCharThreshold = 32         // Control character threshold (ASCII)
	UniversalMaxIntValue          = 2147483647 // Maximum int32 value

	// Security Buffer and Limits - From implementation analysis
	UniversalBufferSize            = 1000 // Default buffer size for various operations
	UniversalBlockedCountThreshold = 10   // Blocked count threshold for security
	UniversalMinBlockedCount       = 5    // Minimum blocked count for actions

	// Security Test Multipliers - From rate limiting configuration
	UniversalRateMultiplierTenth     = 0.1  // 10% of base rate
	UniversalRateMultiplierFifth     = 0.2  // 20% of base rate
	UniversalRateMultiplierTwentieth = 0.05 // 5% of base rate
	UniversalRateMultiplierFiftieth  = 0.02 // 2% of base rate
	UniversalRateMultiplierThirtieth = 0.03 // 3% of base rate
	UniversalRateMultiplierHundredth = 0.01 // 1% of base rate
)

// =============================================================================
// UNIVERSAL TEST STRING CONSTANTS
// =============================================================================
// Common test strings used across multiple modules for consistency.

const (
	// Test Identity Constants
	UniversalTestClient = "test_client"                           // Standard test client identifier
	UniversalTestMethod = "test_method"                           // Standard test method identifier
	UniversalTestSecret = "test_secret_key_for_unit_testing_only" // Standard test JWT secret

	// Test User Constants
	UniversalTestUserID   = "test_user" // Standard test user ID
	UniversalTestUserRole = "admin"     // Standard test user role

	// Test Numeric Constants
	UniversalTestHours24     = 24   // 24 hours (common duration)
	UniversalTestSeconds3600 = 3600 // 3600 seconds (1 hour in seconds)
	UniversalTestSeconds7200 = 7200 // 7200 seconds (2 hours in seconds)

	// Test Error Rate Thresholds
	UniversalErrorRateThreshold  = 0.01 // 1% error rate threshold for load tests
	UniversalErrorRateMultiplier = 100  // Multiplier for percentage display
)

// =============================================================================
// UNIVERSAL TEST CONFIGURATION CONSTANTS
// =============================================================================
// Configuration-related constants for test setup and management.

const (
	// Test File and Directory Constants
	UniversalTestFileSize    = 10 // Default test file size
	UniversalTestBackupCount = 3  // Default backup count for tests

	// File Size Validation Constants - Universal across all modules
	UniversalMinRecordingFileSize = 10000 // 10KB minimum for recordings (meaningful content)
	UniversalMinSnapshotFileSize  = 1000  // 1KB minimum for snapshots (meaningful content)

	// Test Load Generation Constants
	UniversalTestStringLength = 10000 // Length for very long test strings

	// Test Context Constants
	UniversalContextTimeoutShort = 5 * time.Second  // Short context timeout
	UniversalContextTimeoutLong  = 30 * time.Second // Long context timeout
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================
// Utility functions for common test patterns using universal constants.

// GetStandardTestTimeout returns appropriate timeout for test type
func GetStandardTestTimeout(testType string) time.Duration {
	switch testType {
	case "short", "validation", "unit":
		return UniversalTimeoutShort
	case "medium", "cleanup", "polling":
		return UniversalTimeoutMedium
	case "long", "connection", "readiness":
		return UniversalTimeoutLong
	case "startup", "initialization":
		return UniversalTimeoutVeryLong
	case "extreme", "workflow", "integration":
		return UniversalTimeoutExtreme
	default:
		return UniversalTimeoutMedium // Safe default
	}
}

// GetStandardRetryConfig returns standard retry configuration
func GetStandardRetryConfig() (attempts int, delay time.Duration) {
	return UniversalRetryAttempts, UniversalRetryDelay
}

// GetStandardConcurrencyConfig returns standard concurrency configuration
func GetStandardConcurrencyConfig() (goroutines, iterations int) {
	return UniversalConcurrencyGoroutines, UniversalConcurrencyIterations
}
