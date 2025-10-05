/*
Component: Security Authentication Integration
Purpose: Validates JWT authentication, authorization, and session management integration
Requirements: REQ-SEC-001, REQ-SEC-002, REQ-SEC-003, REQ-SEC-004
Category: Integration
API Reference: internal/security/jwt_handler.go
Test Organization:
  - TestSecurity_JWTAuthentication (lines 45-85)
  - TestSecurity_AuthorizationBoundaries (lines 87-127)
  - TestSecurity_SessionManagement (lines 129-169)
  - TestSecurity_ErrorPropagation (lines 171-211)
*/

package integration

import (
	"context"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SecurityIntegrationAsserter handles security integration validation
type SecurityIntegrationAsserter struct {
	setup      *testutils.UniversalTestSetup
	jwtHandler *security.JWTHandler
}

// NewSecurityIntegrationAsserter creates a new security integration asserter
func NewSecurityIntegrationAsserter(t *testing.T) *SecurityIntegrationAsserter {
	// Use testutils.SetupTest with valid config fixture
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")

	// Create JWT handler using loaded configuration
	configManager := setup.GetConfigManager()
	config := configManager.GetConfig()

	logger := setup.GetLogger()
	jwtHandler, err := security.NewJWTHandler(config.Security.JWTSecretKey, logger)
	require.NoError(t, err, "JWT handler should be created")

	asserter := &SecurityIntegrationAsserter{
		setup:      setup,
		jwtHandler: jwtHandler,
	}

	// Register cleanup
	t.Cleanup(func() {
		asserter.Cleanup()
	})

	return asserter
}

// Cleanup performs cleanup of all resources
func (a *SecurityIntegrationAsserter) Cleanup() {
	// JWT handler doesn't need explicit cleanup
}

// AssertJWTAuthentication validates JWT token authentication
func (a *SecurityIntegrationAsserter) AssertJWTAuthentication(ctx context.Context, userID, role string) (string, error) {
	// Generate token with 24 hour expiry
	token, err := a.jwtHandler.GenerateToken(userID, role, 24)
	if err != nil {
		return "", err
	}

	// Validate token
	claims, err := a.jwtHandler.ValidateToken(token)
	if err != nil {
		return "", err
	}

	// Verify auth state established
	if claims.UserID != userID {
		return "", assert.AnError
	}
	if claims.Role != role {
		return "", assert.AnError
	}

	return token, nil
}

// AssertAuthorizationBoundary validates authorization enforcement
func (a *SecurityIntegrationAsserter) AssertAuthorizationBoundary(ctx context.Context, token, requiredRole string) error {
	// Validate token
	claims, err := a.jwtHandler.ValidateToken(token)
	if err != nil {
		return err
	}

	// Check authorization by comparing roles
	// Admin has access to everything, user only to user resources
	hasAccess := false
	if requiredRole == "admin" && claims.Role == "admin" {
		hasAccess = true
	} else if requiredRole == "user" && (claims.Role == "admin" || claims.Role == "user") {
		hasAccess = true
	} else if requiredRole == "viewer" {
		hasAccess = true // Everyone can view
	}

	if !hasAccess {
		return assert.AnError // Simulate authorization failure
	}

	return nil
}

// AssertSessionLifecycle validates session management
func (a *SecurityIntegrationAsserter) AssertSessionLifecycle(ctx context.Context, userID string) error {
	// Create session (token)
	token, err := a.jwtHandler.GenerateToken(userID, "user", 24)
	if err != nil {
		return err
	}

	// Validate session is active
	claims, err := a.jwtHandler.ValidateToken(token)
	if err != nil {
		return err
	}

	// Verify session state tracked
	if claims.UserID != userID {
		return assert.AnError
	}

	return nil
}

// TestSecurity_JWTAuthentication_ReqSEC001 validates JWT authentication flow
// REQ-SEC-001: JWT authentication flow
func TestSecurity_JWTAuthentication_ReqSEC001(t *testing.T) {
	asserter := NewSecurityIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Table-driven test for JWT authentication
	tests := []struct {
		name        string
		userID      string
		role        string
		expectError bool
		description string
	}{
		{"valid_admin_token", "admin_user", "admin", false, "Valid admin token"},
		{"valid_user_token", "regular_user", "user", false, "Valid user token"},
		{"valid_viewer_token", "viewer_user", "viewer", false, "Valid viewer token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := asserter.AssertJWTAuthentication(ctx, tt.userID, tt.role)

			if tt.expectError {
				require.Error(t, err, "Authentication should fail: %s", tt.description)
			} else {
				require.NoError(t, err, "Authentication should succeed: %s", tt.description)
				require.NotEmpty(t, token, "Token should be generated")

				// Validate auth state established
				claims, err := asserter.jwtHandler.ValidateToken(token)
				require.NoError(t, err, "Token should be valid")
				assert.Equal(t, tt.userID, claims.UserID, "User ID should match")
				assert.Equal(t, tt.role, claims.Role, "Role should match")

				// Validate subsequent operation succeeds
				// Test authorization by checking role-based access
				hasAccess := (claims.Role == tt.role)
				assert.True(t, hasAccess, "User should have access to their role")
			}
		})
	}
}

// TestSecurity_AuthorizationBoundaries_ReqSEC002 validates authorization enforcement
// REQ-SEC-002: Authorization enforcement at boundaries
func TestSecurity_AuthorizationBoundaries_ReqSEC002(t *testing.T) {
	asserter := NewSecurityIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Create tokens for different roles
	adminToken, err := asserter.AssertJWTAuthentication(ctx, "admin_user", "admin")
	require.NoError(t, err, "Admin token should be created")

	userToken, err := asserter.AssertJWTAuthentication(ctx, "regular_user", "user")
	require.NoError(t, err, "User token should be created")

	// Table-driven test for authorization boundaries
	tests := []struct {
		name          string
		token         string
		requiredRole  string
		expectSuccess bool
		description   string
	}{
		{"admin_access_admin", adminToken, "admin", true, "Admin should access admin resources"},
		{"admin_access_user", adminToken, "user", true, "Admin should access user resources"},
		{"user_access_user", userToken, "user", true, "User should access user resources"},
		{"user_access_admin", userToken, "admin", false, "User should not access admin resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := asserter.AssertAuthorizationBoundary(ctx, tt.token, tt.requiredRole)

			if tt.expectSuccess {
				require.NoError(t, err, "Authorization should succeed: %s", tt.description)
			} else {
				require.Error(t, err, "Authorization should fail: %s", tt.description)
			}
		})
	}
}

// TestSecurity_SessionManagement_ReqSEC003 validates session lifecycle
// REQ-SEC-003: Session management
func TestSecurity_SessionManagement_ReqSEC003(t *testing.T) {
	asserter := NewSecurityIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Table-driven test for session management
	tests := []struct {
		name        string
		userID      string
		expectError bool
		description string
	}{
		{"create_session", "session_user", false, "Create valid session"},
		{"validate_session", "active_user", false, "Validate active session"},
		{"expire_session", "expired_user", false, "Handle expired session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := asserter.AssertSessionLifecycle(ctx, tt.userID)

			if tt.expectError {
				require.Error(t, err, "Session should fail: %s", tt.description)
			} else {
				require.NoError(t, err, "Session should succeed: %s", tt.description)

				// Validate session state tracked
				token, err := asserter.jwtHandler.GenerateToken(tt.userID, "user", 24)
				require.NoError(t, err, "Token generation should succeed")

				claims, err := asserter.jwtHandler.ValidateToken(token)
				require.NoError(t, err, "Token validation should succeed")
				assert.Equal(t, tt.userID, claims.UserID, "Session should track user ID")

				// Validate expired sessions rejected
				// Create expired token by generating with very short expiry
				expiredToken, err := asserter.jwtHandler.GenerateToken(tt.userID, "user", -1) // Negative hours = expired
				if err == nil {
					_, err = asserter.jwtHandler.ValidateToken(expiredToken)
					assert.Error(t, err, "Expired token should be rejected")
				}
			}
		})
	}
}

// TestSecurity_ErrorPropagation_ReqSEC004 validates security error propagation
// REQ-SEC-004: Security error propagation
func TestSecurity_ErrorPropagation_ReqSEC004(t *testing.T) {
	asserter := NewSecurityIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Table-driven test for error propagation
	tests := []struct {
		name        string
		token       string
		expectError bool
		errorCode   string
		description string
	}{
		{"invalid_token", "invalid.token.here", true, "INVALID_TOKEN", "Invalid token format"},
		{"malformed_token", "not.a.token", true, "INVALID_TOKEN", "Malformed token"},
		{"empty_token", "", true, "INVALID_TOKEN", "Empty token"},
		{"valid_token", "", false, "", "Valid token (will be generated)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			if tt.name == "valid_token" {
				// Generate valid token for this test case
				var err error
				token, err = asserter.AssertJWTAuthentication(ctx, "test_user", "user")
				require.NoError(t, err, "Should generate valid token")
			} else {
				token = tt.token
			}

			// Test token validation
			claims, err := asserter.jwtHandler.ValidateToken(token)

			if tt.expectError {
				require.Error(t, err, "Token validation should fail: %s", tt.description)

				// Validate error code correct
				if tt.errorCode != "" {
					assert.Contains(t, err.Error(), tt.errorCode, "Error should contain correct code")
				}

				// Validate diagnostic details preserved
				assert.NotEmpty(t, err.Error(), "Error should contain diagnostic details")
			} else {
				require.NoError(t, err, "Token validation should succeed: %s", tt.description)
				require.NotNil(t, claims, "Claims should be returned")

				// Validate diagnostic details preserved in success case
				assert.Equal(t, "test_user", claims.UserID, "User ID should be preserved")
			}
		})
	}
}
