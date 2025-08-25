//go:build unit
// +build unit

/*
Security Framework Unit Tests

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access
- REQ-SEC-002: API key validation for service-to-service communication
- REQ-SEC-003: Role-based access control for different user types
- REQ-SEC-004: Session management and timeout
- REQ-SEC-005: Security middleware validation

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security_test

import (
	"fmt"
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

// TestPermissionChecker_RoleHierarchy tests role hierarchy functionality
func TestPermissionChecker_RoleHierarchy(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	checker := security.NewPermissionChecker()

	// Test role hierarchy
	assert.True(t, security.RoleAdmin >= security.RoleOperator)
	assert.True(t, security.RoleOperator >= security.RoleViewer)
	assert.True(t, security.RoleAdmin >= security.RoleViewer)

	// Test role string conversion
	assert.Equal(t, "viewer", security.RoleViewer.String())
	assert.Equal(t, "operator", security.RoleOperator.String())
	assert.Equal(t, "admin", security.RoleAdmin.String())

	// Test role validation
	role, err := checker.ValidateRole("admin")
	assert.NoError(t, err)
	assert.Equal(t, security.RoleAdmin, role)

	role, err = checker.ValidateRole("invalid_role")
	assert.Error(t, err)
	assert.Equal(t, security.RoleViewer, role) // Default fallback
}

// TestPermissionChecker_MethodPermissions tests method permission checking
func TestPermissionChecker_MethodPermissions(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	checker := security.NewPermissionChecker()

	tests := []struct {
		name      string
		method    string
		userRole  security.Role
		hasAccess bool
	}{
		{
			name:      "viewer can access ping",
			method:    "ping",
			userRole:  security.RoleViewer,
			hasAccess: true,
		},
		{
			name:      "viewer cannot access take_snapshot",
			method:    "take_snapshot",
			userRole:  security.RoleViewer,
			hasAccess: false,
		},
		{
			name:      "operator can access take_snapshot",
			method:    "take_snapshot",
			userRole:  security.RoleOperator,
			hasAccess: true,
		},
		{
			name:      "admin can access get_metrics",
			method:    "get_metrics",
			userRole:  security.RoleAdmin,
			hasAccess: true,
		},
		{
			name:      "operator cannot access get_metrics",
			method:    "get_metrics",
			userRole:  security.RoleOperator,
			hasAccess: false,
		},
		{
			name:      "unknown method requires admin",
			method:    "unknown_method",
			userRole:  security.RoleAdmin,
			hasAccess: false, // Method not in permission matrix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPermission := checker.HasPermission(tt.userRole, tt.method)
			assert.Equal(t, tt.hasAccess, hasPermission)
		})
	}
}

// TestSessionManager_SessionLifecycle tests session lifecycle management
func TestSessionManager_SessionLifecycle(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	manager := security.NewSessionManager()
	defer manager.Stop()

	// Test session creation
	session, err := manager.CreateSession("test_user", security.RoleAdmin)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test_user", session.UserID)
	assert.Equal(t, security.RoleAdmin, session.Role)
	assert.NotEmpty(t, session.SessionID)

	// Test session validation
	validSession, err := manager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)
	assert.Equal(t, session.SessionID, validSession.SessionID)

	// Test session count
	assert.Equal(t, 1, manager.GetSessionCount())

	// Test session stats
	stats := manager.GetSessionStats()
	assert.Equal(t, 1, stats["total_sessions"])
	roleCounts := stats["role_counts"].(map[string]int)
	assert.Equal(t, 1, roleCounts["admin"])
}

// TestSessionManager_Concurrency tests concurrent session operations
func TestSessionManager_Concurrency(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	manager := security.NewSessionManager()
	defer manager.Stop()

	// Create multiple sessions concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user_%d", id)
			session, err := manager.CreateSession(userID, security.RoleViewer)
			assert.NoError(t, err)
			assert.NotNil(t, session)

			// Validate session
			validSession, err := manager.ValidateSession(session.SessionID)
			assert.NoError(t, err)
			assert.NotNil(t, validSession)

			// Update activity
			manager.UpdateActivity(session.SessionID)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all sessions were created
	assert.Equal(t, 10, manager.GetSessionCount())
}

// TestSessionManager_ExpiryHandling tests session expiry functionality
func TestSessionManager_ExpiryHandling(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	// Create manager with short session timeout
	manager := security.NewSessionManagerWithConfig(1*time.Second, 1*time.Second)
	defer manager.Stop()

	// Create session
	session, err := manager.CreateSession("test_user", security.RoleViewer)
	require.NoError(t, err)

	// Session should be valid immediately
	validSession, err := manager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)

	// Wait for session to expire
	time.Sleep(2 * time.Second)

	// Session should be expired
	expiredSession, err := manager.ValidateSession(session.SessionID)
	assert.Error(t, err)
	assert.Nil(t, expiredSession)

	// Session count should be 0 after cleanup
	time.Sleep(2 * time.Second) // Wait for cleanup
	assert.Equal(t, 0, manager.GetSessionCount())
}

// TestSecurityIntegration_EndToEnd tests end-to-end security flow
func TestSecurityIntegration_EndToEnd(t *testing.T) {
	// REQ-SEC-005: Security middleware validation

	// Create security components
	jwtHandler, err := security.NewJWTHandler("test_secret_key")
	require.NoError(t, err)

	permissionChecker := security.NewPermissionChecker()
	sessionManager := security.NewSessionManager()
	defer sessionManager.Stop()

	// Generate JWT token
	token, err := jwtHandler.GenerateToken("test_user", "operator", 24)
	require.NoError(t, err)

	// Validate JWT token
	claims, err := jwtHandler.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "test_user", claims.UserID)
	assert.Equal(t, "operator", claims.Role)

	// Create session
	session, err := sessionManager.CreateSession(claims.UserID, security.RoleOperator)
	require.NoError(t, err)

	// Test permission checking
	hasPermission := permissionChecker.HasPermission(session.Role, "take_snapshot")
	assert.True(t, hasPermission)

	hasPermission = permissionChecker.HasPermission(session.Role, "get_metrics")
	assert.False(t, hasPermission)

	// Validate session
	validSession, err := sessionManager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)
}

// TestSecurity_ErrorHandling tests error handling scenarios
func TestSecurity_ErrorHandling(t *testing.T) {
	// REQ-SEC-001: JWT token-based authentication for all API access

	// Test JWT handler with empty secret
	handler, err := security.NewJWTHandler("")
	assert.Error(t, err)
	assert.Nil(t, handler)

	// Test session manager with invalid role
	manager := security.NewSessionManager()
	defer manager.Stop()

	session, err := manager.CreateSession("test_user", 999) // Invalid role
	assert.Error(t, err)
	assert.Nil(t, session)

	// Test permission checker with empty method
	checker := security.NewPermissionChecker()
	hasPermission := checker.HasPermission(security.RoleAdmin, "")
	assert.False(t, hasPermission)
}

// TestSecurity_Performance tests performance characteristics
func TestSecurity_Performance(t *testing.T) {
	// REQ-SEC-001: JWT token-based authentication for all API access

	handler, err := security.NewJWTHandler("test_secret_key")
	require.NoError(t, err)

	// Performance test: generate many tokens quickly
	start := time.Now()
	for i := 0; i < 1000; i++ {
		token, err := handler.GenerateToken("test_user", "viewer", 24)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	}
	duration := time.Since(start)

	// Should complete within reasonable time (< 1 second for 1000 tokens)
	assert.Less(t, duration, time.Second, "Token generation should be fast")

	// Average time per token should be < 1ms
	avgTimePerToken := duration / 1000
	assert.Less(t, avgTimePerToken, time.Millisecond, "Average time per token should be < 1ms")
}
