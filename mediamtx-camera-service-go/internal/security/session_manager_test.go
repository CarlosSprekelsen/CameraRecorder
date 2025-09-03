/*
Session Manager Unit Tests

Requirements Coverage:
- REQ-SEC-004: Session management and timeout

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"fmt"
	"sync"
	"testing"
	"time"

	""
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionManager_SessionLifecycle tests session lifecycle management
func TestSessionManager_SessionLifecycle(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	// Use shared security test environment
	env := NewJWTHandler("test_secret_key_for_unit_testing_only")(t)
	defer (t, env)

	// Test session creation using shared utility
	session := utils.CreateTestSession(t, env.SessionManager, "test_user", security.RoleAdmin)
	assert.Equal(t, "test_user", session.UserID)
	assert.Equal(t, security.RoleAdmin, session.Role)
	assert.NotEmpty(t, session.SessionID)

	// Test session validation
	validSession, err := env.SessionManager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)
	assert.Equal(t, session.SessionID, validSession.SessionID)

	// Test session count
	assert.Equal(t, 1, env.SessionManager.GetSessionCount())

	// Test session stats
	stats := env.SessionManager.GetSessionStats()
	assert.Equal(t, 1, stats["total_sessions"])
	roleCounts := stats["role_counts"].(map[string]int)
	assert.Equal(t, 1, roleCounts["admin"])
}

// TestSessionManager_Concurrency tests concurrent session operations
func TestSessionManager_Concurrency(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	// Use shared security test environment
	env := NewJWTHandler("test_secret_key_for_unit_testing_only")(t)
	defer (t, env)

	// Create multiple sessions concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user_%d", id)
			session := utils.CreateTestSession(t, env.SessionManager, userID, security.RoleViewer)

			// Validate session
			validSession, err := env.SessionManager.ValidateSession(session.SessionID)
			assert.NoError(t, err)
			assert.NotNil(t, validSession)

			// Update activity
			env.SessionManager.UpdateActivity(session.SessionID)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all sessions were created
	assert.Equal(t, 10, env.SessionManager.GetSessionCount())
}

// TestSessionManager_ExpiryHandling tests session expiry functionality
func TestSessionManager_ExpiryHandling(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	// Use shared security test environment with custom session timeout
	env := NewJWTHandler("test_secret_key_for_unit_testing_only")(t)
	defer (t, env)

	// Create a custom session manager with short timeout for this test
	shortTimeoutManager := NewSessionManager(1*time.Second, 1*time.Second)
	defer shortTimeoutManager.Stop()

	// Create session using the short timeout manager
	session, err := shortTimeoutManager.CreateSession("test_user", security.RoleViewer)
	require.NoError(t, err)

	// Session should be valid immediately
	validSession, err := shortTimeoutManager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)

	// Wait for session to expire
	time.Sleep(2 * time.Second)

	// Session should be expired
	expiredSession, err := shortTimeoutManager.ValidateSession(session.SessionID)
	assert.Error(t, err)
	assert.Nil(t, expiredSession)

	// Session count should be 0 after cleanup
	time.Sleep(2 * time.Second) // Wait for cleanup
	assert.Equal(t, 0, shortTimeoutManager.GetSessionCount())
}

func TestSessionManager_EdgeCases(t *testing.T) {
	// Use shared security test environment
	env := NewJWTHandler("test_secret_key_for_unit_testing_only")(t)
	defer (t, env)

	t.Run("get_session_by_user_id", func(t *testing.T) {

		// Create a session using shared utility
		session := utils.CreateTestSession(t, env.SessionManager, "test_user", security.RoleAdmin)

		// Get session by user ID
		foundSessions := env.SessionManager.GetSessionByUserID("test_user")
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, session.SessionID, foundSession.SessionID)
		assert.Equal(t, "test_user", foundSession.UserID)
		assert.Equal(t, security.RoleAdmin, foundSession.Role)

		// Test with non-existent user
		notFound := env.SessionManager.GetSessionByUserID("nonexistent_user")
		assert.Nil(t, notFound)
	})

	t.Run("invalidate_user_sessions", func(t *testing.T) {
		// Create multiple sessions for the same user
		utils.CreateTestSession(t, env.SessionManager, "multi_user", security.RoleViewer)
		utils.CreateTestSession(t, env.SessionManager, "multi_user", security.RoleOperator)

		// Create session for different user
		utils.CreateTestSession(t, env.SessionManager, "other_user", security.RoleAdmin)

		// Verify sessions exist
		assert.NotNil(t, env.SessionManager.GetSessionByUserID("multi_user"))
		assert.NotNil(t, env.SessionManager.GetSessionByUserID("other_user"))

		// Invalidate all sessions for multi_user
		err := env.SessionManager.InvalidateUserSessions("multi_user")
		require.NoError(t, err)

		// Verify multi_user sessions are gone
		assert.Nil(t, env.SessionManager.GetSessionByUserID("multi_user"))

		// Verify other_user session still exists
		assert.NotNil(t, env.SessionManager.GetSessionByUserID("other_user"))
	})

	t.Run("session_with_special_characters", func(t *testing.T) {
		specialUserID := "user@domain.com!@#$%^&*()"
		utils.CreateTestSession(t, env.SessionManager, specialUserID, security.RoleAdmin)

		foundSessions := env.SessionManager.GetSessionByUserID(specialUserID)
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, specialUserID, foundSession.UserID)
	})

	t.Run("concurrent_session_creation_same_user", func(t *testing.T) {
		var wg sync.WaitGroup
		sessions := make([]*security.Session, 10)
		errors := make([]error, 10)

		// Create 10 sessions concurrently for the same user
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				sessions[index], errors[index] = env.SessionManager.CreateSession("concurrent_user", security.RoleViewer)
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
		session := utils.CreateTestSession(t, env.SessionManager, "activity_user", security.RoleOperator)

		initialActivity := session.LastActivity

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Update activity
		env.SessionManager.UpdateActivity(session.SessionID)

		// Get updated session
		updatedSessions := env.SessionManager.GetSessionByUserID("activity_user")
		assert.NotNil(t, updatedSessions)
		assert.Len(t, updatedSessions, 1)
		updatedSession := updatedSessions[0]
		assert.True(t, updatedSession.LastActivity.After(initialActivity))
	})
}

func TestSessionManager_AdditionalEdgeCases(t *testing.T) {
	// Use shared security test environment
	env := NewJWTHandler("test_secret_key_for_unit_testing_only")(t)
	defer (t, env)

	t.Run("create_session_with_empty_user_id", func(t *testing.T) {
		_, err := env.SessionManager.CreateSession("", security.RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("create_session_with_invalid_role", func(t *testing.T) {
		_, err := env.SessionManager.CreateSession("test_user", security.Role(999))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("validate_session_with_empty_id", func(t *testing.T) {
		_, err := env.SessionManager.ValidateSession("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")
	})

	t.Run("update_activity_with_empty_id", func(t *testing.T) {
		// Should not panic or error
		env.SessionManager.UpdateActivity("")
	})

	t.Run("get_session_by_user_id_with_empty_id", func(t *testing.T) {
		sessions := env.SessionManager.GetSessionByUserID("")
		assert.Nil(t, sessions)
	})

	t.Run("invalidate_user_sessions_with_empty_id", func(t *testing.T) {
		err := env.SessionManager.InvalidateUserSessions("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("session_manager_with_very_short_timeout", func(t *testing.T) {
		// Create a custom session manager with very short timeout for this specific test
		shortTimeoutManager := NewSessionManager(1*time.Millisecond, 1*time.Millisecond)
		defer shortTimeoutManager.Stop()

		session, err := shortTimeoutManager.CreateSession("test_user", security.RoleViewer)
		require.NoError(t, err)
		assert.NotNil(t, session)

		// Wait for session to expire and cleanup to run
		time.Sleep(10 * time.Millisecond)

		// Session should be removed by cleanup
		_, err = shortTimeoutManager.ValidateSession(session.SessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

// Performance benchmarks for session manager
// Note: Benchmarks use individual managers for performance isolation
func BenchmarkSessionManager_CreateSession(b *testing.B) {
	manager := NewSessionManager(24*time.Hour, 5*time.Minute)
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
	manager := NewSessionManager(24*time.Hour, 5*time.Minute)
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
	manager := NewSessionManager(24*time.Hour, 5*time.Minute)
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
