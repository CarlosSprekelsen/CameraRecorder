//go:build unit
// +build unit

/*
JWT Handler Unit Tests

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security_test

import (
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTHandler_TokenGeneration tests JWT token generation functionality
func TestJWTHandler_TokenGeneration(t *testing.T) {
	// REQ-SEC-001: JWT token-based authentication for all API access

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
			handler, err := security.NewJWTHandler("test_secret_key")
			require.NoError(t, err)

			token, err := handler.GenerateToken(tt.userID, tt.role, tt.expiryHours)

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
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := security.NewJWTHandler("test_secret_key")
	require.NoError(t, err)

	// Generate a valid token
	token, err := handler.GenerateToken("test_user", "admin", 24)
	require.NoError(t, err)

	// Test valid token validation
	claims, err := handler.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "test_user", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Greater(t, claims.EXP, claims.IAT)

	// Test invalid token
	invalidToken := "invalid.jwt.token"
	claims, err = handler.ValidateToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)

	// Test empty token
	claims, err = handler.ValidateToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)

	// Test token with wrong secret
	wrongHandler, err := security.NewJWTHandler("wrong_secret_key")
	require.NoError(t, err)

	claims, err = wrongHandler.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestJWTHandler_ExpiryHandling tests JWT token expiry functionality
func TestJWTHandler_ExpiryHandling(t *testing.T) {
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := security.NewJWTHandler("test_secret_key")
	require.NoError(t, err)

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
	assert.False(t, handler.IsTokenExpired(tokenString))

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should be expired
	assert.True(t, handler.IsTokenExpired(tokenString))

	// Validation should fail for expired token
	claims, err := handler.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// Edge case tests for missing coverage
func TestJWTHandler_EdgeCases(t *testing.T) {
	handler, err := security.NewJWTHandler("test_secret")
	require.NoError(t, err)

	t.Run("get_secret_key", func(t *testing.T) {
		secret := handler.GetSecretKey()
		assert.Equal(t, "test_secret", secret)
	})

	t.Run("get_algorithm", func(t *testing.T) {
		algo := handler.GetAlgorithm()
		assert.Equal(t, "HS256", algo)
	})

	t.Run("token_with_special_characters", func(t *testing.T) {
		token, err := handler.GenerateToken("user@domain.com", "admin", 1)
		require.NoError(t, err)

		// Validate the token
		claims, err := handler.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user@domain.com", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("very_long_user_id", func(t *testing.T) {
		longUserID := strings.Repeat("a", 1000)
		token, err := handler.GenerateToken(longUserID, "viewer", 1)
		require.NoError(t, err)

		claims, err := handler.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, longUserID, claims.UserID)
	})
}

// Additional edge cases for higher coverage
func TestJWTHandler_AdditionalEdgeCases(t *testing.T) {
	t.Run("token_with_max_expiry", func(t *testing.T) {
		handler, err := security.NewJWTHandler("test_secret")
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
		handler, err := security.NewJWTHandler("test_secret")
		require.NoError(t, err)

		unicodeUserID := "user_测试_🎉_🚀"
		token, err := handler.GenerateToken(unicodeUserID, "viewer", 1)
		require.NoError(t, err)

		claims, err := handler.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, unicodeUserID, claims.UserID)
	})

	t.Run("validate_token_with_whitespace", func(t *testing.T) {
		handler, err := security.NewJWTHandler("test_secret")
		require.NoError(t, err)

		// Test with whitespace in token
		_, err = handler.ValidateToken("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})
}

// Performance benchmarks for JWT handler
func BenchmarkJWTHandler_TokenGeneration(b *testing.B) {
	handler, err := security.NewJWTHandler("test_secret")
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
	handler, err := security.NewJWTHandler("test_secret")
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
