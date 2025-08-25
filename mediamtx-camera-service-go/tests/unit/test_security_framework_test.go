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
	"strings"
	"sync"
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
	handler, err := security.NewJWTHandler("test_secret")
	require.NoError(t, err)

	// Performance test for token generation
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := handler.GenerateToken(fmt.Sprintf("user_%d", i), "viewer", 1)
		require.NoError(t, err)
	}
	duration := time.Since(start)

	// Should complete within reasonable time
	assert.Less(t, duration, 5*time.Second, "Token generation should be fast")
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

func TestPermissionChecker_EdgeCases(t *testing.T) {
	checker := security.NewPermissionChecker()

	t.Run("get_required_role", func(t *testing.T) {
		role := checker.GetRequiredRole("ping")
		assert.Equal(t, security.RoleViewer, role)

		role = checker.GetRequiredRole("take_snapshot")
		assert.Equal(t, security.RoleOperator, role)

		role = checker.GetRequiredRole("get_metrics")
		assert.Equal(t, security.RoleAdmin, role)

		role = checker.GetRequiredRole("nonexistent_method")
		assert.Equal(t, security.RoleAdmin, role) // Default to admin for unknown methods
	})

	t.Run("get_role_hierarchy", func(t *testing.T) {
		hierarchy := checker.GetRoleHierarchy()
		expected := map[string]int{
			"viewer":   1,
			"operator": 2,
			"admin":    3,
		}
		assert.Equal(t, expected, hierarchy)
	})

	t.Run("get_method_permissions", func(t *testing.T) {
		permissions := checker.GetMethodPermissions()
		assert.NotEmpty(t, permissions)
		assert.Contains(t, permissions, "ping")
		assert.Contains(t, permissions, "take_snapshot")
		assert.Contains(t, permissions, "get_metrics")
	})

	t.Run("add_method_permission", func(t *testing.T) {
		err := checker.AddMethodPermission("test_method", security.RoleViewer)
		require.NoError(t, err)
		assert.True(t, checker.HasPermission(security.RoleViewer, "test_method"))
		assert.True(t, checker.HasPermission(security.RoleOperator, "test_method"))
		assert.True(t, checker.HasPermission(security.RoleAdmin, "test_method"))
	})

	t.Run("remove_method_permission", func(t *testing.T) {
		// First add a permission
		err := checker.AddMethodPermission("temp_method", security.RoleOperator)
		require.NoError(t, err)
		assert.True(t, checker.HasPermission(security.RoleOperator, "temp_method"))

		// Then remove it
		err = checker.RemoveMethodPermission("temp_method")
		require.NoError(t, err)
		assert.False(t, checker.HasPermission(security.RoleViewer, "temp_method"))
		assert.False(t, checker.HasPermission(security.RoleOperator, "temp_method"))
		assert.False(t, checker.HasPermission(security.RoleAdmin, "temp_method"))
	})

	t.Run("validate_role_edge_cases", func(t *testing.T) {
		role, err := checker.ValidateRole("viewer")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleViewer, role)

		role, err = checker.ValidateRole("operator")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleOperator, role)

		role, err = checker.ValidateRole("admin")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleAdmin, role)

		invalidRole, err := checker.ValidateRole("invalid_role")
		assert.Error(t, err)
		assert.Equal(t, security.RoleViewer, invalidRole) // Returns default role

		emptyRole, err := checker.ValidateRole("")
		assert.Error(t, err)
		assert.Equal(t, security.RoleViewer, emptyRole) // Returns default role

		caseRole, err := checker.ValidateRole("VIEWER") // Should work due to ToLower conversion
		assert.NoError(t, err)
		assert.Equal(t, security.RoleViewer, caseRole) // Should return viewer role
	})
}

func TestSessionManager_EdgeCases(t *testing.T) {
	t.Run("get_session_by_user_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()

		// Create a session
		session, err := manager.CreateSession("test_user", security.RoleAdmin)
		require.NoError(t, err)
		assert.NotNil(t, session)

		// Get session by user ID
		foundSessions := manager.GetSessionByUserID("test_user")
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, session.SessionID, foundSession.SessionID)
		assert.Equal(t, "test_user", foundSession.UserID)
		assert.Equal(t, security.RoleAdmin, foundSession.Role)

		// Test with non-existent user
		notFound := manager.GetSessionByUserID("nonexistent_user")
		assert.Nil(t, notFound)
	})

	t.Run("invalidate_user_sessions", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()

		// Create multiple sessions for the same user
		_, err := manager.CreateSession("multi_user", security.RoleViewer)
		require.NoError(t, err)

		_, err = manager.CreateSession("multi_user", security.RoleOperator)
		require.NoError(t, err)

		// Create session for different user
		_, err = manager.CreateSession("other_user", security.RoleAdmin)
		require.NoError(t, err)

		// Verify sessions exist
		assert.NotNil(t, manager.GetSessionByUserID("multi_user"))
		assert.NotNil(t, manager.GetSessionByUserID("other_user"))

		// Invalidate all sessions for multi_user
		err = manager.InvalidateUserSessions("multi_user")
		require.NoError(t, err)

		// Verify multi_user sessions are gone
		assert.Nil(t, manager.GetSessionByUserID("multi_user"))

		// Verify other_user session still exists
		assert.NotNil(t, manager.GetSessionByUserID("other_user"))
	})

	t.Run("session_with_special_characters", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()

		specialUserID := "user@domain.com!@#$%^&*()"
		session, err := manager.CreateSession(specialUserID, security.RoleAdmin)
		require.NoError(t, err)
		assert.NotNil(t, session)

		foundSessions := manager.GetSessionByUserID(specialUserID)
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, specialUserID, foundSession.UserID)
	})

	t.Run("concurrent_session_creation_same_user", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()

		var wg sync.WaitGroup
		sessions := make([]*security.Session, 10)
		errors := make([]error, 10)

		// Create 10 sessions concurrently for the same user
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				sessions[index], errors[index] = manager.CreateSession("concurrent_user", security.RoleViewer)
			}(i)
		}

		wg.Wait()

		// All should succeed
		for i := 0; i < 10; i++ {
			assert.NoError(t, errors[i])
			assert.NotNil(t, sessions[i])
		}

		// All sessions should have unique IDs
		sessionIDs := make(map[string]bool)
		for _, session := range sessions {
			assert.False(t, sessionIDs[session.SessionID], "Session ID should be unique")
			sessionIDs[session.SessionID] = true
		}
	})

	t.Run("session_activity_tracking", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()

		session, err := manager.CreateSession("activity_user", security.RoleOperator)
		require.NoError(t, err)

		initialActivity := session.LastActivity

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Update activity
		manager.UpdateActivity(session.SessionID)

		// Get updated session
		updatedSessions := manager.GetSessionByUserID("activity_user")
		assert.NotNil(t, updatedSessions)
		assert.Len(t, updatedSessions, 1)
		updatedSession := updatedSessions[0]
		assert.True(t, updatedSession.LastActivity.After(initialActivity))
	})
}

func TestSecurity_ComprehensiveEdgeCases(t *testing.T) {
	t.Run("jwt_with_empty_secret", func(t *testing.T) {
		_, err := security.NewJWTHandler("")
		require.Error(t, err) // Should fail with empty secret
		assert.Contains(t, err.Error(), "secret key must be provided")
	})

	t.Run("session_manager_with_zero_timeout", func(t *testing.T) {
		manager := security.NewSessionManagerWithConfig(0, 1*time.Second)
		defer manager.Stop()

		// Should still work with zero timeout
		session, err := manager.CreateSession("test_user", security.RoleViewer)
		require.NoError(t, err)
		assert.NotNil(t, session)
	})

	t.Run("permission_checker_with_empty_method", func(t *testing.T) {
		checker := security.NewPermissionChecker()

		// Empty method should return false for all roles (as per implementation)
		assert.False(t, checker.HasPermission(security.RoleViewer, ""))
		assert.False(t, checker.HasPermission(security.RoleOperator, ""))
		assert.False(t, checker.HasPermission(security.RoleAdmin, ""))
	})

	t.Run("jwt_token_manipulation_detection", func(t *testing.T) {
		handler, err := security.NewJWTHandler("test_secret")
		require.NoError(t, err)

		// Generate valid token
		token, err := handler.GenerateToken("test_user", "admin", 1)
		require.NoError(t, err)

		// Tamper with token (add character)
		tamperedToken := token + "x"

		// Should fail validation
		_, err = handler.ValidateToken(tamperedToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})
}

// Performance benchmarks for security framework
func BenchmarkJWTHandler_TokenGeneration(b *testing.B) {
	handler, err := security.NewJWTHandler("test_secret")
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.GenerateToken(fmt.Sprintf("user_%d", i), "viewer", 24)
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

func BenchmarkPermissionChecker_HasPermission(b *testing.B) {
	checker := security.NewPermissionChecker()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.HasPermission(security.RoleViewer, "ping")
		checker.HasPermission(security.RoleOperator, "take_snapshot")
		checker.HasPermission(security.RoleAdmin, "get_metrics")
	}
}

func BenchmarkSessionManager_CreateSession(b *testing.B) {
	manager := security.NewSessionManager()
	defer manager.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.CreateSession(fmt.Sprintf("user_%d", i), security.RoleViewer)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionManager_ValidateSession(b *testing.B) {
	manager := security.NewSessionManager()
	defer manager.Stop()
	
	session, err := manager.CreateSession("test_user", security.RoleAdmin)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ValidateSession(session.SessionID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionManager_ConcurrentOperations(b *testing.B) {
	manager := security.NewSessionManager()
	defer manager.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Create session
			session, err := manager.CreateSession(fmt.Sprintf("user_%d", i), security.RoleViewer)
			if err != nil {
				b.Fatal(err)
			}
			
			// Validate session
			_, err = manager.ValidateSession(session.SessionID)
			if err != nil {
				b.Fatal(err)
			}
			
			// Update activity
			manager.UpdateActivity(session.SessionID)
			
			i++
		}
	})
}

func BenchmarkSecurity_EndToEnd(b *testing.B) {
	handler, err := security.NewJWTHandler("test_secret")
	if err != nil {
		b.Fatal(err)
	}
	
	checker := security.NewPermissionChecker()
	manager := security.NewSessionManager()
	defer manager.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate token
		token, err := handler.GenerateToken(fmt.Sprintf("user_%d", i), "operator", 24)
		if err != nil {
			b.Fatal(err)
		}
		
		// Validate token
		claims, err := handler.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
		
		// Create session
		session, err := manager.CreateSession(claims.UserID, security.RoleOperator)
		if err != nil {
			b.Fatal(err)
		}
		
		// Check permissions
		hasPermission := checker.HasPermission(security.RoleOperator, "take_snapshot")
		if !hasPermission {
			b.Fatal("Expected permission check to pass")
		}
		
		// Validate session
		_, err = manager.ValidateSession(session.SessionID)
		if err != nil {
			b.Fatal(err)
		}
	}
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

func TestPermissionChecker_AdditionalEdgeCases(t *testing.T) {
	t.Run("get_required_role_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()
		
		role := checker.GetRequiredRole("   ")
		assert.Equal(t, security.RoleAdmin, role) // Should default to admin
	})
	
	t.Run("add_method_permission_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()
		
		err := checker.AddMethodPermission("   ", security.RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})
	
	t.Run("remove_method_permission_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()
		
		err := checker.RemoveMethodPermission("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})
	
	t.Run("add_method_permission_with_invalid_role", func(t *testing.T) {
		checker := security.NewPermissionChecker()
		
		err := checker.AddMethodPermission("test_method", security.Role(999))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
}

func TestSessionManager_AdditionalEdgeCases(t *testing.T) {
	t.Run("create_session_with_empty_user_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		_, err := manager.CreateSession("", security.RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})
	
	t.Run("create_session_with_invalid_role", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		_, err := manager.CreateSession("test_user", security.Role(999))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
	
	t.Run("validate_session_with_empty_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		_, err := manager.ValidateSession("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")
	})
	
	t.Run("update_activity_with_empty_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		// Should not panic or error
		manager.UpdateActivity("")
	})
	
	t.Run("get_session_by_user_id_with_empty_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		sessions := manager.GetSessionByUserID("")
		assert.Nil(t, sessions)
	})
	
	t.Run("invalidate_user_sessions_with_empty_id", func(t *testing.T) {
		manager := security.NewSessionManager()
		defer manager.Stop()
		
		err := manager.InvalidateUserSessions("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})
	
	t.Run("session_manager_with_very_short_timeout", func(t *testing.T) {
		manager := security.NewSessionManagerWithConfig(1*time.Millisecond, 1*time.Millisecond)
		defer manager.Stop()
		
		session, err := manager.CreateSession("test_user", security.RoleViewer)
		require.NoError(t, err)
		assert.NotNil(t, session)
		
		// Wait for session to expire and cleanup to run
		time.Sleep(10 * time.Millisecond)
		
		// Session should be removed by cleanup
		_, err = manager.ValidateSession(session.SessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}
