/*
JWT Handler Unit Tests

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTHandler_TokenGeneration tests JWT token generation functionality
func TestJWTHandler_TokenGeneration(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use test helper for consistent JWT handler creation
	jwtHandler := TestJWTHandler(t)

	tests := []struct {
		name        string
		userID      string
		role        string
		expiryHours int
		wantErr     bool
	}{
		{
			name:        "valid token generation",
			userID:      "test_user",
			role:        "admin",
			expiryHours: 24,
			wantErr:     false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			role:        "viewer",
			expiryHours: 24,
			wantErr:     true,
		},
		{
			name:        "invalid role",
			userID:      "test_user",
			role:        "invalid_role",
			expiryHours: 24,
			wantErr:     true,
		},
		{
			name:        "zero expiry hours",
			userID:      "test_user",
			role:        "operator",
			expiryHours: 0,
			wantErr:     false, // Should default to 24 hours
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtHandler.GenerateToken(tt.userID, tt.role, tt.expiryHours)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.True(t, len(token) > 100) // JWT tokens are typically long
			}
		})
	}
}

// =============================================================================
// COMPREHENSIVE JWT HANDLER TESTS FOR 90%+ COVERAGE
// =============================================================================

func TestJWTHandler_CheckRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		clientID    string
		requests    int
		expectLimit bool
	}{
		{"Within limit", "client1", 50, false},
		{"At limit", "client2", 100, false},
		{"Over limit", "client3", 101, true},
		{"Multiple clients", "client4", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a separate JWT handler for each test case to avoid interference
			jwtHandler := TestJWTHandler(t)

			// Make multiple requests to test rate limiting
			for i := 0; i < tt.requests; i++ {
				limited := jwtHandler.CheckRateLimit(tt.clientID)
				// i is 0-based, so i=100 means the 101st request
				// After 100 requests (i=99), the 101st request (i=100) should be rate limited
				// limited=true means request allowed, limited=false means request blocked (rate limited)
				if i >= 100 && tt.expectLimit {
					assert.False(t, limited, "Should be rate limited after 100 requests (limited=false)")
				} else {
					assert.True(t, limited, "Should not be rate limited (limited=true)")
				}
			}
		})
	}
}

func TestJWTHandler_RecordRequest(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	clientID := "test_client"

	// Record some requests
	jwtHandler.RecordRequest(clientID)
	jwtHandler.RecordRequest(clientID)
	jwtHandler.RecordRequest(clientID)

	// Check rate limit info
	info := jwtHandler.GetClientRateInfo(clientID)
	assert.NotNil(t, info)
	assert.Equal(t, clientID, info.ClientID)
	assert.Equal(t, int64(3), info.RequestCount)
}

func TestJWTHandler_GetClientRateInfo(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	clientID := "test_client"

	// Record some requests
	jwtHandler.RecordRequest(clientID)
	jwtHandler.RecordRequest(clientID)

	// Get rate info
	info := jwtHandler.GetClientRateInfo(clientID)
	assert.NotNil(t, info)
	assert.Equal(t, clientID, info.ClientID)
	assert.Equal(t, int64(2), info.RequestCount)
}

func TestJWTHandler_SetRateLimit(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	// Set custom rate limit
	jwtHandler.SetRateLimit(50, time.Minute)

	// Test with new limit
	clientID := "test_client"
	for i := 0; i < 51; i++ {
		limited := jwtHandler.CheckRateLimit(clientID)
		// limited=true means request allowed, limited=false means request blocked (rate limited)
		if i >= 50 {
			assert.False(t, limited, "Should be rate limited after 50 requests (limited=false)")
		} else {
			assert.True(t, limited, "Should not be rate limited (limited=true)")
		}
	}
}

func TestJWTHandler_CleanupExpiredClients(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	clientID := "test_client"

	// Record some requests
	jwtHandler.RecordRequest(clientID)

	// Cleanup expired clients (should not affect recent requests)
	jwtHandler.CleanupExpiredClients(1 * time.Minute)

	// Client should still exist
	info := jwtHandler.GetClientRateInfo(clientID)
	assert.NotNil(t, info)
	assert.Equal(t, clientID, info.ClientID)
}

func TestJWTHandler_IsTokenExpired(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{"Valid token", GenerateTestToken(t, jwtHandler, "user1", "admin"), false},
		{"Expired token", GenerateExpiredTestToken(t, jwtHandler, "user1", "admin"), true},
		{"Invalid token", "invalid.token.here", true},
		{"Empty token", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expired := jwtHandler.IsTokenExpired(tt.token)
			if tt.expectError {
				assert.True(t, expired, "Token should be expired or invalid")
			} else {
				assert.False(t, expired, "Token should not be expired")
			}
		})
	}
}

func TestJWTHandler_GetSecretKey(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	secretKey := jwtHandler.GetSecretKey()
	assert.NotEmpty(t, secretKey)
	assert.Equal(t, "test_secret_key_for_unit_testing_only", secretKey)
}

func TestJWTHandler_GetAlgorithm(t *testing.T) {
	t.Parallel()
	jwtHandler := TestJWTHandler(t)

	algorithm := jwtHandler.GetAlgorithm()
	assert.NotEmpty(t, algorithm)
	assert.Equal(t, "HS256", algorithm)
}

// TestJWTHandler_TokenValidation tests JWT token validation functionality
func TestJWTHandler_TokenValidation(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use test helper for consistent JWT handler creation
	jwtHandler := TestJWTHandler(t)

	// Generate a valid token using test helper
	token := GenerateTestToken(t, jwtHandler, "test_user", "admin")

	// Test valid token validation
	claims, err := jwtHandler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "test_user", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Greater(t, claims.EXP, claims.IAT)

	// Test invalid token
	invalidToken := "invalid.jwt.token"
	_, err = jwtHandler.ValidateToken(invalidToken)
	assert.Error(t, err)

	// Test empty token
	_, err = jwtHandler.ValidateToken("")
	assert.Error(t, err)

	// Test token with wrong secret
	logger := logging.NewLogger("test-wrong-handler")
	wrongHandler, err := NewJWTHandler("wrong_secret_key", logger)
	require.NoError(t, err)

	_, err = wrongHandler.ValidateToken(token)
	assert.Error(t, err)
}

// TestJWTHandler_ExpiryHandling tests JWT token expiry functionality
func TestJWTHandler_ExpiryHandling(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Create JWT handler directly for unit testing
	logger := logging.NewLogger("test-jwt-handler")
	jwtHandler, err := NewJWTHandler("test_secret_key_for_unit_testing_only", logger)
	require.NoError(t, err)

	// Test token with short expiry
	token, err := jwtHandler.GenerateToken("user@domain.com", "admin", 1)
	require.NoError(t, err)

	claims, err := jwtHandler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user@domain.com", claims.UserID)
	assert.Equal(t, "admin", claims.Role)

	// Test token with long user ID
	longUserID := "very_long_user_id_that_exceeds_normal_length_limits_and_should_still_work_properly"
	token, err = jwtHandler.GenerateToken(longUserID, "viewer", 1)
	require.NoError(t, err)

	claims, err = jwtHandler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, longUserID, claims.UserID)
	assert.Equal(t, "viewer", claims.Role)
}

// TestJWTHandler_ClaimsValidation tests JWT claims validation functionality
func TestJWTHandler_ClaimsValidation(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Create JWT handler directly for unit testing
	logger := logging.NewLogger("test-jwt-handler")
	jwtHandler, err := NewJWTHandler("test_secret_key_for_unit_testing_only", logger)
	require.NoError(t, err)

	// Generate token and validate claims
	token, err := jwtHandler.GenerateToken("test_user", "operator", 24)
	require.NoError(t, err)

	claims, err := jwtHandler.ValidateToken(token)
	require.NoError(t, err)

	// Verify all required claims are present
	assert.NotEmpty(t, claims.UserID)
	assert.NotEmpty(t, claims.Role)
	assert.NotZero(t, claims.IAT) // Issued at
	assert.NotZero(t, claims.EXP) // Expiration

	// Verify claim values
	assert.Equal(t, "test_user", claims.UserID)
	assert.Equal(t, "operator", claims.Role)

	// Verify timing claims
	now := time.Now().Unix()
	assert.LessOrEqual(t, claims.IAT, now)
	assert.Greater(t, claims.EXP, now)
}

// TestJWTHandler_ErrorHandling tests JWT error handling functionality
func TestJWTHandler_ErrorHandling(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Test invalid secret key
	logger := logging.NewLogger("test-invalid-handler")
	_, err := NewJWTHandler("", logger)
	assert.Error(t, err)

	// Test very long secret key
	longSecret := string(make([]byte, 1000))
	_, err = NewJWTHandler(longSecret, logger)
	assert.NoError(t, err) // Should handle long secrets

	// Test special characters in secret
	specialSecret := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	_, err = NewJWTHandler(specialSecret, logger)
	assert.NoError(t, err) // Should handle special characters
}
