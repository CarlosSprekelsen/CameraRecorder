//go:build unit
// +build unit

/*
Session Manager Unit Tests

Requirements Coverage:
- REQ-SEC-004: Session management and timeout

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// Performance benchmarks for session manager
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
