package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRoleBasics tests basic role functionality without complex dependencies
func TestRoleBasics(t *testing.T) {
	// Test role hierarchy
	assert.True(t, RoleAdmin >= RoleOperator)
	assert.True(t, RoleOperator >= RoleViewer)
	assert.True(t, RoleAdmin >= RoleViewer)

	// Test role string conversion
	assert.Equal(t, "viewer", RoleViewer.String())
	assert.Equal(t, "operator", RoleOperator.String())
	assert.Equal(t, "admin", RoleAdmin.String())
}

// TestPermissionCheckerBasics tests basic permission checker functionality
func TestPermissionCheckerBasics(t *testing.T) {
	checker := NewPermissionChecker()

	// Test basic permission checking
	assert.True(t, checker.HasPermission(RoleViewer, "ping"))
	assert.False(t, checker.HasPermission(RoleViewer, "take_snapshot"))
	assert.True(t, checker.HasPermission(RoleOperator, "take_snapshot"))
	assert.True(t, checker.HasPermission(RoleAdmin, "get_metrics"))

	// Test role validation
	role, err := checker.ValidateRole("admin")
	assert.NoError(t, err)
	assert.Equal(t, RoleAdmin, role)

	role, err = checker.ValidateRole("invalid_role")
	assert.Error(t, err)
	assert.Equal(t, RoleViewer, role) // Default fallback
}

// TestSessionManagerBasics tests basic session manager functionality
func TestSessionManagerBasics(t *testing.T) {
	// Use very short timeouts for fast test execution
	// 100ms session timeout, 50ms cleanup interval
	manager := NewSessionManager(100*time.Millisecond, 50*time.Millisecond)
	defer manager.Stop(context.Background())

	// Test session creation
	session, err := manager.CreateSession("test_user", RoleViewer)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test_user", session.UserID)
	assert.Equal(t, RoleViewer, session.Role)

	// Test session validation immediately (before cleanup can interfere)
	validSession, err := manager.ValidateSession(session.SessionID)
	assert.NoError(t, err)
	assert.NotNil(t, validSession)
	assert.Equal(t, session.SessionID, validSession.SessionID)
}

// TestInputValidatorBasics tests basic input validation functionality
func TestInputValidatorBasics(t *testing.T) {
	validator := NewInputValidator(nil, nil)

	// Test camera ID validation
	result := validator.ValidateCameraID("camera001")
	assert.False(t, result.HasErrors(), "Valid camera ID should pass")

	result = validator.ValidateCameraID("")
	assert.True(t, result.HasErrors(), "Empty camera ID should fail")

	// Test duration validation
	result, duration := validator.ValidateDuration("1m")
	assert.False(t, result.HasErrors(), "Valid duration should pass")
	assert.Greater(t, duration, time.Duration(0), "Duration should parse to positive value")
}
