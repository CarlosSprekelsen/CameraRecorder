package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/google/uuid"
)

// Session represents a user session with authentication and activity tracking.
// Mirrors the Python session structure for compatibility.
type Session struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
}

// SessionManager manages user sessions with thread-safe operations.
// Implements session creation, validation, cleanup, and activity tracking.
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	logger   *logging.Logger
	// Configuration
	defaultSessionTimeout time.Duration
	cleanupInterval       time.Duration
	// Cleanup control
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewSessionManager creates a new session manager with custom configuration.
// This constructor initializes the session manager with automatic cleanup routines
// and ensures proper resource management through graceful shutdown mechanisms.
func NewSessionManager(sessionTimeout, cleanupInterval time.Duration) *SessionManager {
	manager := &SessionManager{
		sessions:              make(map[string]*Session),
		logger:                logging.GetLogger("session-manager"),
		defaultSessionTimeout: sessionTimeout,
		cleanupInterval:       cleanupInterval,
		stopChan:              make(chan struct{}, 5), // Buffered to prevent deadlock during shutdown
	}

	// Start automatic cleanup
	manager.startCleanup()

	manager.logger.WithFields(logging.Fields{
		"session_timeout":  sessionTimeout,
		"cleanup_interval": cleanupInterval,
	}).Info("Session manager initialized with custom configuration")

	return manager
}

// CreateSession creates a new session for the specified user.
// Returns the session and any error encountered during creation.
func (sm *SessionManager) CreateSession(userID string, role Role) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if role < RoleViewer || role > RoleAdmin {
		return nil, fmt.Errorf("invalid role: %d", role)
	}

	now := time.Now()
	sessionID := uuid.New().String()

	session := &Session{
		SessionID:    sessionID,
		UserID:       userID,
		Role:         role,
		CreatedAt:    now,
		ExpiresAt:    now.Add(sm.defaultSessionTimeout),
		LastActivity: now,
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	sm.logger.WithFields(logging.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"role":       role.String(),
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	}).Info("Session created successfully")

	return session, nil
}

// ValidateSession validates a session and returns the session if valid.
// This method performs comprehensive session validation including expiration checks,
// updates last activity timestamp, and automatically removes expired sessions.
// Returns the session if valid, error if invalid, expired, or not found.
func (sm *SessionManager) ValidateSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		sm.logger.WithField("session_id", sessionID).Debug("Session not found")
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		sm.logger.WithField("session_id", sessionID).Debug("Session has expired")
		// Remove expired session
		sm.removeSession(sessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// Update last activity
	sm.UpdateActivity(sessionID)

	sm.logger.WithFields(logging.Fields{
		"session_id": sessionID,
		"user_id":    session.UserID,
		"role":       session.Role.String(),
	}).Debug("Session validated successfully")

	return session, nil
}

// UpdateActivity updates the last activity timestamp for a session.
// This method is thread-safe and can be called frequently.
func (sm *SessionManager) UpdateActivity(sessionID string) {
	if sessionID == "" {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// RemoveSession removes a session from the manager.
// This method is thread-safe and can be called from cleanup routines.
func (sm *SessionManager) removeSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		sm.logger.WithField("session_id", sessionID).Debug("Session removed")
	}
}

// CleanupExpiredSessions removes all expired sessions from the manager.
// This method is called automatically by the cleanup routine.
func (sm *SessionManager) CleanupExpiredSessions() {
	now := time.Now()
	expiredSessions := make([]string, 0)

	sm.mu.RLock()
	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	sm.mu.RUnlock()

	// Remove expired sessions
	for _, sessionID := range expiredSessions {
		sm.removeSession(sessionID)
	}

	if len(expiredSessions) > 0 {
		sm.logger.WithField("expired_count", fmt.Sprintf("%d", len(expiredSessions))).Info("Cleaned up expired sessions")
	}
}

// startCleanup starts the automatic cleanup routine.
func (sm *SessionManager) startCleanup() {
	sm.cleanupTicker = time.NewTicker(sm.cleanupInterval)
	sm.wg.Add(1)

	go func() {
		defer sm.wg.Done()
		for {
			select {
			case <-sm.cleanupTicker.C:
				sm.CleanupExpiredSessions()
			case <-sm.stopChan:
				return
			}
		}
	}()
}

// Stop stops the session manager and cleanup routine with context-aware cancellation.
// This method should be called when shutting down the application.
func (sm *SessionManager) Stop(ctx context.Context) error {
	if sm.cleanupTicker != nil {
		sm.cleanupTicker.Stop()
	}

	// Signal stop
	select {
	case <-sm.stopChan:
		// Already closed
	default:
		close(sm.stopChan)
	}

	// Wait for cleanup goroutine to finish with timeout
	done := make(chan struct{})
	go func() {
		sm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		sm.logger.Warn("Session manager shutdown timeout")
		return ctx.Err()
	}

	sm.logger.Info("Session manager stopped")
	return nil
}

// GetSessionCount returns the current number of active sessions.
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// GetSessionStats returns statistics about the current sessions.
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := logging.Fields{
		"total_sessions": len(sm.sessions),
		"role_counts":    make(map[string]int),
	}

	// Count sessions by role
	for _, session := range sm.sessions {
		roleName := session.Role.String()
		stats["role_counts"].(map[string]int)[roleName]++
	}

	return stats
}

// GetSessionByUserID returns all sessions for a specific user.
// Useful for debugging and session management.
func (sm *SessionManager) GetSessionByUserID(userID string) []*Session {
	if userID == "" {
		return nil
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var userSessions []*Session
	for _, session := range sm.sessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions
}

// InvalidateUserSessions invalidates all sessions for a specific user.
// Useful when a user's permissions change or for security purposes.
func (sm *SessionManager) InvalidateUserSessions(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionsToRemove := make([]string, 0)
	for sessionID, session := range sm.sessions {
		if session.UserID == userID {
			sessionsToRemove = append(sessionsToRemove, sessionID)
		}
	}

	for _, sessionID := range sessionsToRemove {
		delete(sm.sessions, sessionID)
	}

	sm.logger.WithFields(logging.Fields{
		"user_id":              userID,
		"sessions_invalidated": len(sessionsToRemove),
	}).Info("User sessions invalidated")

	return nil
}
