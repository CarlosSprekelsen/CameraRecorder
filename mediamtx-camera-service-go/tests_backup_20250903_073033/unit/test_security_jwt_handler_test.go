//go:build unit
// +build unit

/*
JWT Handler Unit Tests

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTHandler_TokenGeneration tests JWT token generation functionality
func TestJWTHandler_TokenGeneration(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

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
			// Use shared JWT handler instead of creating new one
			token, err := env.JWTHandler.GenerateToken(tt.userID, tt.role, tt.expiryHours)

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

// TestJWTHandler_TokenValidation tests JWT token validation functionality
func TestJWTHandler_TokenValidation(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

	// Generate a valid token using shared utility
	token := testtestutils.GenerateTestToken(t, env.JWTHandler, "test_user", "admin")

	// Test valid token validation using shared utility
	claims := testtestutils.ValidateTestToken(t, env.JWTHandler, token)
	assert.Equal(t, "test_user", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Greater(t, claims.EXP, claims.IAT)

	// Test invalid token
	invalidToken := "invalid.jwt.token"
	_, err := env.JWTHandler.ValidateToken(invalidToken)
	assert.Error(t, err)

	// Test empty token
	_, err = env.JWTHandler.ValidateToken("")
	assert.Error(t, err)

	// Test token with wrong secret
	wrongHandler, err := NewJWTHandler("wrong_secret_key")
	require.NoError(t, err)

	_, err = wrongHandler.ValidateToken(token)
	assert.Error(t, err)
}

// TestJWTHandler_ExpiryHandling tests JWT token expiry functionality
func TestJWTHandler_ExpiryHandling(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

	// Generate a token with very short expiry (1 second)
	// We need to create a token that expires in 1 second
	now := time.Now().Unix()
	expiresAt := now + 1 // 1 second from now

	// Create token manually with short expiry
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test_user",
		"role":    "viewer",
		"iat":     now,
		"exp":     expiresAt,
	})

	tokenString, err := token.SignedString([]byte("test_secret_key"))
	require.NoError(t, err)

	// Token should be valid immediately
	assert.False(t, env.JWTHandler.IsTokenExpired(tokenString))

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should be expired
	assert.True(t, env.JWTHandler.IsTokenExpired(tokenString))

	// Validation should fail for expired token
	_, err = env.JWTHandler.ValidateToken(tokenString)
	assert.Error(t, err)
}

// Edge case tests for missing coverage
func TestJWTHandler_EdgeCases(t *testing.T) {
	t.Parallel()
	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

	t.Run("get_secret_key", func(t *testing.T) {
		secret := env.JWTHandler.GetSecretKey()
		assert.NotEmpty(t, secret)
	})

	t.Run("get_algorithm", func(t *testing.T) {
		algo := env.JWTHandler.GetAlgorithm()
		assert.Equal(t, "HS256", algo)
	})

	t.Run("token_with_special_characters", func(t *testing.T) {
		token := testtestutils.GenerateTestTokenWithExpiry(t, env.JWTHandler, "user@domain.com", "admin", 1)

		// Validate the token using shared utility
		claims := testtestutils.ValidateTestToken(t, env.JWTHandler, token)
		assert.Equal(t, "user@domain.com", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("very_long_user_id", func(t *testing.T) {
		longUserID := strings.Repeat("a", 1000)
		token := testtestutils.GenerateTestTokenWithExpiry(t, env.JWTHandler, longUserID, "viewer", 1)

		claims := testtestutils.ValidateTestToken(t, env.JWTHandler, token)
		assert.Equal(t, longUserID, claims.UserID)
	})
}

// Additional edge cases for higher coverage
func TestJWTHandler_AdditionalEdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	t.Run("token_with_max_expiry", func(t *testing.T) {
		handler, err := NewJWTHandler("test_secret")
		require.NoError(t, err)

		// Test with maximum reasonable expiry (365 days)
		token, err := handler.GenerateToken("test_user", "admin", 24*365)
		require.NoError(t, err)

		claims, err := handler.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "test_user", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("token_with_unicode_characters", func(t *testing.T) {
		handler, err := NewJWTHandler("test_secret")
		require.NoError(t, err)

		unicodeUserID := "user_测试_🎉_🚀"
		token, err := handler.GenerateToken(unicodeUserID, "viewer", 1)
		require.NoError(t, err)

		claims, err := handler.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, unicodeUserID, claims.UserID)
	})

	t.Run("validate_token_with_whitespace", func(t *testing.T) {
		handler, err := NewJWTHandler("test_secret")
		require.NoError(t, err)

		// Test with whitespace in token
		_, err = handler.ValidateToken("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})
}

// TestJWTHandler_RateLimiting tests rate limiting functionality
func TestJWTHandler_RateLimiting(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

	// Set rate limit to 3 requests per 1 second window
	env.JWTHandler.SetRateLimit(3, time.Second)

	t.Run("check_rate_limit_first_request", func(t *testing.T) {
		// First request should be allowed
		allowed := env.JWTHandler.CheckRateLimit("client1")
		assert.True(t, allowed)
	})

	t.Run("check_rate_limit_within_limit", func(t *testing.T) {
		// Multiple requests within limit should be allowed
		for i := 0; i < 3; i++ {
			allowed := env.JWTHandler.CheckRateLimit("client2")
			assert.True(t, allowed)
		}
	})

	t.Run("check_rate_limit_exceeded", func(t *testing.T) {
		// Exceed rate limit
		for i := 0; i < 3; i++ {
			env.JWTHandler.CheckRateLimit("client3")
		}
		// Fourth request should be blocked
		allowed := env.JWTHandler.CheckRateLimit("client3")
		assert.False(t, allowed)
	})

	t.Run("check_rate_limit_window_reset", func(t *testing.T) {
		// Fill up the window
		for i := 0; i < 3; i++ {
			env.JWTHandler.CheckRateLimit("client4")
		}
		// Wait for window to reset
		time.Sleep(1100 * time.Millisecond)
		// Should be allowed again
		allowed := env.JWTHandler.CheckRateLimit("client4")
		assert.True(t, allowed)
	})
}

// TestJWTHandler_RecordRequest tests the RecordRequest function
func TestJWTHandler_RecordRequest(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Use shared security test environment
	env := testtestutils.SetupSecurityTestEnvironment(t)
	defer testtestutils.TeardownSecurityTestEnvironment(t, env)

	t.Run("record_request_new_client", func(t *testing.T) {
		// Record request for new client
		env.JWTHandler.RecordRequest("new_client")

		// Check rate info
		rateInfo := env.JWTHandler.GetClientRateInfo("new_client")
		assert.NotNil(t, rateInfo)
		assert.Equal(t, "new_client", rateInfo.ClientID)
		assert.Equal(t, int64(1), rateInfo.RequestCount)
	})

	t.Run("record_request_existing_client", func(t *testing.T) {
		// Record multiple requests
		env.JWTHandler.RecordRequest("existing_client")
		env.JWTHandler.RecordRequest("existing_client")
		env.JWTHandler.RecordRequest("existing_client")

		// Check rate info
		rateInfo := env.JWTHandler.GetClientRateInfo("existing_client")
		assert.NotNil(t, rateInfo)
		assert.Equal(t, int64(3), rateInfo.RequestCount)
	})

	t.Run("record_request_window_reset", func(t *testing.T) {
		// Set short window for testing
		env.JWTHandler.SetRateLimit(10, 100*time.Millisecond)

		// Record request
		env.JWTHandler.RecordRequest("window_client")

		// Wait for window to reset
		time.Sleep(150 * time.Millisecond)

		// Record another request
		env.JWTHandler.RecordRequest("window_client")

		// Should reset to 1
		rateInfo := env.JWTHandler.GetClientRateInfo("window_client")
		assert.NotNil(t, rateInfo)
		assert.Equal(t, int64(1), rateInfo.RequestCount)
	})
}

// TestJWTHandler_GetClientRateInfo tests the GetClientRateInfo function
func TestJWTHandler_GetClientRateInfo(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("get_client_rate_info_nonexistent", func(t *testing.T) {
		// Get info for non-existent client
		rateInfo := handler.GetClientRateInfo("nonexistent_client")
		assert.Nil(t, rateInfo)
	})

	t.Run("get_client_rate_info_existing", func(t *testing.T) {
		// Record some requests
		handler.RecordRequest("test_client")
		handler.RecordRequest("test_client")

		// Get rate info
		rateInfo := handler.GetClientRateInfo("test_client")
		assert.NotNil(t, rateInfo)
		assert.Equal(t, "test_client", rateInfo.ClientID)
		assert.Equal(t, int64(2), rateInfo.RequestCount)
		assert.NotZero(t, rateInfo.LastRequest)
		assert.NotZero(t, rateInfo.WindowStart)
	})

	t.Run("get_client_rate_info_copy", func(t *testing.T) {
		// Record request
		handler.RecordRequest("copy_client")

		// Get rate info
		rateInfo1 := handler.GetClientRateInfo("copy_client")
		rateInfo2 := handler.GetClientRateInfo("copy_client")

		// Should be different instances (copies)
		assert.NotSame(t, rateInfo1, rateInfo2)
		assert.Equal(t, rateInfo1.ClientID, rateInfo2.ClientID)
		assert.Equal(t, rateInfo1.RequestCount, rateInfo2.RequestCount)
	})
}

// TestJWTHandler_SetRateLimit tests the SetRateLimit function
func TestJWTHandler_SetRateLimit(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("set_rate_limit_default", func(t *testing.T) {
		// Set default rate limit
		handler.SetRateLimit(100, time.Minute)

		// Test that it's applied
		for i := 0; i < 100; i++ {
			allowed := handler.CheckRateLimit("default_client")
			assert.True(t, allowed)
		}
		// 101st request should be blocked
		allowed := handler.CheckRateLimit("default_client")
		assert.False(t, allowed)
	})

	t.Run("set_rate_limit_high", func(t *testing.T) {
		// Set high rate limit
		handler.SetRateLimit(1000, time.Hour)

		// Test that it's applied
		for i := 0; i < 1000; i++ {
			allowed := handler.CheckRateLimit("high_limit_client")
			assert.True(t, allowed)
		}
		// 1001st request should be blocked
		allowed := handler.CheckRateLimit("high_limit_client")
		assert.False(t, allowed)
	})

	t.Run("set_rate_limit_low", func(t *testing.T) {
		// Set low rate limit
		handler.SetRateLimit(1, time.Second)

		// First request should be allowed
		allowed := handler.CheckRateLimit("low_limit_client")
		assert.True(t, allowed)

		// Second request should be blocked
		allowed = handler.CheckRateLimit("low_limit_client")
		assert.False(t, allowed)
	})
}

// TestJWTHandler_CleanupExpiredClients tests the CleanupExpiredClients function
func TestJWTHandler_CleanupExpiredClients(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("cleanup_expired_clients", func(t *testing.T) {
		// Record requests for multiple clients
		handler.RecordRequest("active_client")
		handler.RecordRequest("expired_client1")
		handler.RecordRequest("expired_client2")

		// Wait for some clients to become inactive
		time.Sleep(100 * time.Millisecond)

		// Update activity for active client
		handler.RecordRequest("active_client")

		// Cleanup clients inactive for more than 50ms
		handler.CleanupExpiredClients(50 * time.Millisecond)

		// Active client should still exist
		rateInfo := handler.GetClientRateInfo("active_client")
		assert.NotNil(t, rateInfo)

		// Expired clients should be removed
		rateInfo = handler.GetClientRateInfo("expired_client1")
		assert.Nil(t, rateInfo)

		rateInfo = handler.GetClientRateInfo("expired_client2")
		assert.Nil(t, rateInfo)
	})

	t.Run("cleanup_no_expired_clients", func(t *testing.T) {
		// Record request for active client
		handler.RecordRequest("recent_client")

		// Cleanup with long inactive duration
		handler.CleanupExpiredClients(1 * time.Hour)

		// Client should still exist
		rateInfo := handler.GetClientRateInfo("recent_client")
		assert.NotNil(t, rateInfo)
	})
}

// TestJWTHandler_ValidateToken_EdgeCases tests edge cases for ValidateToken
func TestJWTHandler_ValidateToken_EdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("validate_token_missing_required_fields", func(t *testing.T) {
		// Create token missing required fields
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			// Missing role, iat, exp
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		claims, err := handler.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "missing required field")
	})

	t.Run("validate_token_invalid_role", func(t *testing.T) {
		// Create token with invalid role
		now := time.Now().Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			"role":    "invalid_role",
			"iat":     now,
			"exp":     now + 3600,
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		claims, err := handler.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("validate_token_invalid_timestamps", func(t *testing.T) {
		// Create token with invalid timestamp types
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			"role":    "admin",
			"iat":     "invalid_timestamp",
			"exp":     "invalid_timestamp",
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		claims, err := handler.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		// The actual error message may vary, so we check for any validation error
		assert.Contains(t, err.Error(), "Token used before issued")
	})

	t.Run("validate_token_wrong_signing_method", func(t *testing.T) {
		// Create a token with HS256 but manually change the header to RS256
		// This simulates an algorithm confusion attack
		now := time.Now().Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			"role":    "admin",
			"iat":     now,
			"exp":     now + 3600,
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		// Manually modify the header to change algorithm from HS256 to RS256
		parts := strings.Split(tokenString, ".")
		if len(parts) != 3 {
			t.Fatal("Invalid token format")
		}

		// Decode header
		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		require.NoError(t, err)

		var header map[string]interface{}
		err = json.Unmarshal(headerBytes, &header)
		require.NoError(t, err)

		// Change algorithm to RS256
		header["alg"] = "RS256"

		// Re-encode header
		newHeaderBytes, err := json.Marshal(header)
		require.NoError(t, err)
		newHeader := base64.RawURLEncoding.EncodeToString(newHeaderBytes)

		// Reconstruct token with modified header
		modifiedToken := newHeader + "." + parts[1] + "." + parts[2]

		claims, err := handler.ValidateToken(modifiedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		// With the fixed implementation, RS256 should be rejected with "unsupported signing method"
		assert.Contains(t, err.Error(), "unsupported signing method")
	})
}

// TestJWTHandler_IsTokenExpired_EdgeCases tests edge cases for IsTokenExpired
func TestJWTHandler_IsTokenExpired_EdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("is_token_expired_empty_token", func(t *testing.T) {
		// Test empty token
		expired := handler.IsTokenExpired("")
		assert.True(t, expired)

		// Test whitespace token
		expired = handler.IsTokenExpired("   ")
		assert.True(t, expired)
	})

	t.Run("is_token_expired_invalid_token", func(t *testing.T) {
		// Test invalid token format
		expired := handler.IsTokenExpired("invalid.token.format")
		assert.True(t, expired)
	})

	t.Run("is_token_expired_missing_exp", func(t *testing.T) {
		// Create token without exp field
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			"role":    "admin",
			"iat":     time.Now().Unix(),
			// Missing exp
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		expired := handler.IsTokenExpired(tokenString)
		assert.True(t, expired)
	})

	t.Run("is_token_expired_invalid_exp_type", func(t *testing.T) {
		// Create token with invalid exp type
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "test_user",
			"role":    "admin",
			"iat":     time.Now().Unix(),
			"exp":     "invalid_exp",
		})

		tokenString, err := token.SignedString([]byte("test_secret"))
		require.NoError(t, err)

		expired := handler.IsTokenExpired(tokenString)
		assert.True(t, expired)
	})
}

// Performance benchmarks for JWT handler
func BenchmarkJWTHandler_TokenGeneration(b *testing.B) {
	handler, err := NewJWTHandler("test_secret")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.GenerateToken("test_user", "viewer", 24)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTHandler_TokenValidation(b *testing.B) {
	handler, err := NewJWTHandler("test_secret")
	if err != nil {
		b.Fatal(err)
	}

	token, err := handler.GenerateToken("test_user", "admin", 24)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTHandler_RateLimitCheck(b *testing.B) {
	handler, err := NewJWTHandler("test_secret")
	if err != nil {
		b.Fatal(err)
	}

	handler.SetRateLimit(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.CheckRateLimit("benchmark_client")
	}
}
