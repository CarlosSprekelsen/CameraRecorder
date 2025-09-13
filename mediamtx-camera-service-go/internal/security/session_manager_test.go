/*
Session Manager Unit Tests

Requirements Coverage:
- REQ-SEC-004: Session management and timeout

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionManager_SessionLifecycle tests session lifecycle management
func TestSessionManager_SessionLifecycle(t *testing.T) {
	// REQ-SEC-004: Session management and timeout

	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	// Test session creation using test helper
	session := CreateTestSession(t, env.SessionManager, "test_user", RoleAdmin)
	assert.Equal(t, "test_user", session.UserID)
	assert.Equal(t, RoleAdmin, session.Role)
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

	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	// Create multiple sessions concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user_%d", id)
			session := CreateTestSession(t, env.SessionManager, userID, RoleViewer)

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

	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	// Create a custom session manager with very short timeout for this test
	shortTimeoutManager := NewSessionManager(1*time.Millisecond, 1*time.Millisecond) // 1ms timeout
	defer shortTimeoutManager.Stop()

	// Create session using the short timeout manager
	session, err := shortTimeoutManager.CreateSession("test_user", RoleViewer)
	require.NoError(t, err)

	// Session should be valid immediately
	validSession, err := shortTimeoutManager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)

	// Wait just 2ms for session to actually expire (much faster than 2 seconds)
	time.Sleep(2 * time.Millisecond)

	// Manually trigger cleanup to remove expired sessions
	shortTimeoutManager.CleanupExpiredSessions()

	// Session should be expired
	expiredSession, err := shortTimeoutManager.ValidateSession(session.SessionID)
	assert.Error(t, err)
	assert.Nil(t, expiredSession)

	// Session count should be 0 after cleanup
	assert.Equal(t, 0, shortTimeoutManager.GetSessionCount())
}

func TestSessionManager_EdgeCases(t *testing.T) {
	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	t.Run("get_session_by_user_id", func(t *testing.T) {

		// Create a session using test helper
		session := CreateTestSession(t, env.SessionManager, "test_user", RoleAdmin)

		// Get session by user ID
		foundSessions := env.SessionManager.GetSessionByUserID("test_user")
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, session.SessionID, foundSession.SessionID)
		assert.Equal(t, "test_user", foundSession.UserID)
		assert.Equal(t, RoleAdmin, foundSession.Role)

		// Test with non-existent user
		notFound := env.SessionManager.GetSessionByUserID("nonexistent_user")
		assert.Nil(t, notFound)
	})

	t.Run("invalidate_user_sessions", func(t *testing.T) {
		// Create multiple sessions for the same user
		CreateTestSession(t, env.SessionManager, "multi_user", RoleViewer)
		CreateTestSession(t, env.SessionManager, "multi_user", RoleOperator)

		// Create session for different user
		CreateTestSession(t, env.SessionManager, "other_user", RoleAdmin)

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
		CreateTestSession(t, env.SessionManager, specialUserID, RoleAdmin)

		foundSessions := env.SessionManager.GetSessionByUserID(specialUserID)
		assert.NotNil(t, foundSessions)
		assert.Len(t, foundSessions, 1)
		foundSession := foundSessions[0]
		assert.Equal(t, specialUserID, foundSession.UserID)
	})

	t.Run("concurrent_session_creation_same_user", func(t *testing.T) {
		var wg sync.WaitGroup
		sessions := make([]*Session, 10)
		errors := make([]error, 10)

		// Create 10 sessions concurrently for the same user
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				sessions[index], errors[index] = env.SessionManager.CreateSession("concurrent_user", RoleViewer)
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
		session := CreateTestSession(t, env.SessionManager, "activity_user", RoleOperator)

		initialActivity := session.LastActivity

		// Update activity (no delay needed for unit test)
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
	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	t.Run("create_session_with_empty_user_id", func(t *testing.T) {
		_, err := env.SessionManager.CreateSession("", RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("create_session_with_invalid_role", func(t *testing.T) {
		_, err := env.SessionManager.CreateSession("test_user", Role(999))
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

		session, err := shortTimeoutManager.CreateSession("test_user", RoleViewer)
		require.NoError(t, err)
		assert.NotNil(t, session)

		// Wait 2ms for session to expire, then trigger cleanup
		time.Sleep(2 * time.Millisecond)
		shortTimeoutManager.CleanupExpiredSessions()

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
		_, err := manager.CreateSession(fmt.Sprintf("user_%d", i), RoleViewer)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionManager_ValidateSession(b *testing.B) {
	manager := NewSessionManager(24*time.Hour, 5*time.Minute)
	defer manager.Stop()

	session, err := manager.CreateSession("test_user", RoleAdmin)
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
			session, err := manager.CreateSession(fmt.Sprintf("user_%d", i), RoleViewer)
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

// TestSessionManager_ContextAwareShutdown tests the context-aware shutdown functionality
func TestSessionManager_ContextAwareShutdown(t *testing.T) {
	t.Run("graceful_shutdown_with_context", func(t *testing.T) {
		// Use security test environment from test helpers
		env := SetupTestSecurityEnvironment(t)
		defer TeardownTestSecurityEnvironment(t, env)

		// Test graceful shutdown with context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err := env.SessionManager.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Session manager should stop gracefully")
		assert.Less(t, elapsed, 1*time.Second, "Shutdown should be fast")
	})

	t.Run("shutdown_with_cancelled_context", func(t *testing.T) {
		// Use security test environment from test helpers
		env := SetupTestSecurityEnvironment(t)
		defer TeardownTestSecurityEnvironment(t, env)

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Stop should complete quickly since context is already cancelled
		start := time.Now()
		err := env.SessionManager.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Session manager should stop even with cancelled context")
		assert.Less(t, elapsed, 100*time.Millisecond, "Shutdown should be very fast with cancelled context")
	})

	t.Run("shutdown_timeout_handling", func(t *testing.T) {
		// Use security test environment from test helpers
		env := SetupTestSecurityEnvironment(t)
		defer TeardownTestSecurityEnvironment(t, env)

		// Use very short timeout to test timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(2 * time.Millisecond)

		start := time.Now()
		err := env.SessionManager.Stop(ctx)
		elapsed := time.Since(start)

		// Should timeout but not hang
		require.Error(t, err, "Should timeout with very short timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Error should indicate timeout")
		assert.Less(t, elapsed, 1*time.Second, "Should not hang indefinitely")
	})

	t.Run("double_stop_handling", func(t *testing.T) {
		// Use security test environment from test helpers
		env := SetupTestSecurityEnvironment(t)
		defer TeardownTestSecurityEnvironment(t, env)

		// Stop first time
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()
		err := env.SessionManager.Stop(ctx1)
		require.NoError(t, err, "First stop should succeed")

		// Stop second time should not error
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		err = env.SessionManager.Stop(ctx2)
		assert.NoError(t, err, "Second stop should not error")
	})
}
