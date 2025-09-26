/*
WebSocket Error Handling Integration Tests - Error Scenarios and Recovery

Tests comprehensive error handling scenarios including authentication failures,
invalid requests, network issues, and recovery patterns. Validates that the system
gracefully handles errors and provides meaningful error responses.

API Documentation Reference: docs/api/json_rpc_methods.md
Requirements Coverage:
- REQ-WS-003: Error handling and recovery
- REQ-API-003: Request/response message handling
- REQ-API-004: Error response standardization
- REQ-API-005: JSON-RPC 2.0 error compliance

Design Principles:
- Real components only (no mocks)
- Comprehensive error scenario coverage
- Recovery pattern validation
- Meaningful error messages
- JSON-RPC 2.0 error compliance
*/

package websocket

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// AUTHENTICATION ERROR TESTS
// ============================================================================

// TestErrorHandling_InvalidToken_Integration validates error handling
// for invalid JWT tokens
func TestErrorHandling_InvalidToken_Integration(t *testing.T) {
	// Create integration asserter with real components
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	// Progressive Readiness first
	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	// Test invalid token scenarios
	err = asserter.AssertInvalidTokenHandling()
	require.NoError(t, err, "Invalid token handling should work")

	t.Log("✅ Invalid token error handling validated")
}

// TestErrorHandling_ExpiredToken_Integration validates error handling
// for expired JWT tokens
func TestErrorHandling_ExpiredToken_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertExpiredTokenHandling()
	require.NoError(t, err, "Expired token handling should work")

	t.Log("✅ Expired token error handling validated")
}

// TestErrorHandling_MalformedToken_Integration validates error handling
// for malformed JWT tokens
func TestErrorHandling_MalformedToken_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertMalformedTokenHandling()
	require.NoError(t, err, "Malformed token handling should work")

	t.Log("✅ Malformed token error handling validated")
}

// ============================================================================
// NETWORK ERROR TESTS
// ============================================================================

// TestErrorHandling_NetworkRecovery_Integration validates network error
// recovery patterns
func TestErrorHandling_NetworkRecovery_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertNetworkErrorRecovery()
	require.NoError(t, err, "Network error recovery should work")

	t.Log("✅ Network error recovery validated")
}

// TestErrorHandling_ConnectionTimeout_Integration validates connection
// timeout handling
func TestErrorHandling_ConnectionTimeout_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertConnectionTimeoutHandling()
	require.NoError(t, err, "Connection timeout handling should work")

	t.Log("✅ Connection timeout handling validated")
}

// ============================================================================
// REQUEST ERROR TESTS
// ============================================================================

// TestErrorHandling_InvalidMethod_Integration validates error handling
// for invalid JSON-RPC methods
func TestErrorHandling_InvalidMethod_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertInvalidMethodHandling()
	require.NoError(t, err, "Invalid method handling should work")

	t.Log("✅ Invalid method error handling validated")
}

// TestErrorHandling_MalformedRequest_Integration validates error handling
// for malformed JSON-RPC requests
func TestErrorHandling_MalformedRequest_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertMalformedRequestHandling()
	require.NoError(t, err, "Malformed request handling should work")

	t.Log("✅ Malformed request error handling validated")
}

// TestErrorHandling_InvalidParameters_Integration validates error handling
// for invalid method parameters
func TestErrorHandling_InvalidParameters_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertInvalidParametersHandling()
	require.NoError(t, err, "Invalid parameters handling should work")

	t.Log("✅ Invalid parameters error handling validated")
}

// ============================================================================
// GRACEFUL DEGRADATION TESTS
// ============================================================================

// TestErrorHandling_GracefulDegradation_Integration validates graceful
// degradation under error conditions
func TestErrorHandling_GracefulDegradation_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertGracefulDegradation()
	require.NoError(t, err, "Graceful degradation should work")

	t.Log("✅ Graceful degradation validated")
}

// TestErrorHandling_ServiceUnavailable_Integration validates error handling
// when service is unavailable
func TestErrorHandling_ServiceUnavailable_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertServiceUnavailableHandling()
	require.NoError(t, err, "Service unavailable handling should work")

	t.Log("✅ Service unavailable error handling validated")
}

// ============================================================================
// COMPREHENSIVE ERROR SCENARIOS
// ============================================================================

// TestErrorHandling_ComprehensiveScenarios_Integration validates
// comprehensive error handling scenarios
func TestErrorHandling_ComprehensiveScenarios_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertComprehensiveErrorScenarios()
	require.NoError(t, err, "Comprehensive error scenarios should work")

	t.Log("✅ Comprehensive error scenarios validated")
}

// TestErrorHandling_ErrorRecovery_Integration validates error recovery
// patterns and resilience
func TestErrorHandling_ErrorRecovery_Integration(t *testing.T) {
	asserter := NewWebSocketIntegrationAsserter(t)
	defer asserter.EnhancedCleanup()

	err := asserter.AssertProgressiveReadiness()
	require.NoError(t, err, "Progressive Readiness should work")

	err = asserter.AssertErrorRecoveryPatterns()
	require.NoError(t, err, "Error recovery patterns should work")

	t.Log("✅ Error recovery patterns validated")
}
