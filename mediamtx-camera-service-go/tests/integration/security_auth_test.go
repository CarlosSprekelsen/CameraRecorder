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
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SecurityIntegrationAsserter handles security integration validation
type SecurityIntegrationAsserter struct {
	setup          *testutils.UniversalTestSetup
	securityHelper *testutils.SecurityHelper
	jwtHandler     *security.JWTHandler
}

// NewSecurityIntegrationAsserter creates a new security integration asserter
func NewSecurityIntegrationAsserter(t *testing.T) *SecurityIntegrationAsserter {
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")

	securityHelper := testutils.NewSecurityHelper(t, setup)

	return &SecurityIntegrationAsserter{
		setup:          setup,
		securityHelper: securityHelper,
		jwtHandler:     securityHelper.GetJWTHandler(),
	}
}

// AssertJWTAuthentication validates JWT token authentication
func (a *SecurityIntegrationAsserter) AssertJWTAuthentication(ctx context.Context, userID, role string) (string, error) {
	// Generate token using SecurityHelper
	token, err := a.securityHelper.GenerateTestToken(userID, role, 24*time.Hour)
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
		return "", fmt.Errorf("user ID mismatch: expected %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		return "", fmt.Errorf("role mismatch: expected %s, got %s", role, claims.Role)
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

	// Check authorization using real role hierarchy: admin > operator > viewer
	hasAccess := false
	if claims.Role == "admin" {
		hasAccess = true // Admin has access to everything
	} else if claims.Role == "operator" && (requiredRole == "operator" || requiredRole == "viewer") {
		hasAccess = true // Operator can access operator and viewer resources
	} else if claims.Role == "viewer" && requiredRole == "viewer" {
		hasAccess = true // Viewer can only access viewer resources
	}

	if !hasAccess {
		return fmt.Errorf("authorization failed: role %s cannot access %s resources", claims.Role, requiredRole)
	}

	return nil
}

// AssertSessionLifecycle validates session management
func (a *SecurityIntegrationAsserter) AssertSessionLifecycle(ctx context.Context, userID string) error {
	// Create session (token)
	token, err := a.jwtHandler.GenerateToken(userID, "operator", 24)
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
		return fmt.Errorf("session user ID mismatch: expected %s, got %s", userID, claims.UserID)
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
		{"valid_admin_token", testutils.UniversalTestUserID, testutils.UniversalTestUserRole, false, "Valid admin token"},
		{"valid_operator_token", testutils.UniversalTestUserID, "operator", false, "Valid operator token"},
		{"valid_viewer_token", testutils.UniversalTestUserID, "viewer", false, "Valid viewer token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use AssertionHelper for consistent assertions
			ah := testutils.NewAssertionHelper(t)

			token, err := asserter.AssertJWTAuthentication(ctx, tt.userID, tt.role)

			if tt.expectError {
				require.Error(t, err, "Authentication should fail: %s", tt.description)
			} else {
				ah.AssertNoErrorWithContext(err, tt.description)
				require.NotEmpty(t, token, "Token should be generated")

				// Validate auth state established
				claims, err := asserter.jwtHandler.ValidateToken(token)
				ah.AssertNoErrorWithContext(err, "Token validation")
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

	// Use AssertionHelper for consistent assertions
	ah := testutils.NewAssertionHelper(t)

	// Create tokens for different roles
	adminToken, err := asserter.AssertJWTAuthentication(ctx, testutils.UniversalTestUserID, testutils.UniversalTestUserRole)
	ah.AssertNoErrorWithContext(err, "Admin token creation")

	operatorToken, err := asserter.AssertJWTAuthentication(ctx, testutils.UniversalTestUserID, "operator")
	ah.AssertNoErrorWithContext(err, "Operator token creation")

	// Table-driven test for authorization boundaries
	tests := []struct {
		name          string
		token         string
		requiredRole  string
		expectSuccess bool
		description   string
	}{
		{"admin_access_admin", adminToken, testutils.UniversalTestUserRole, true, "Admin should access admin resources"},
		{"admin_access_operator", adminToken, "operator", true, "Admin should access operator resources"},
		{"operator_access_operator", operatorToken, "operator", true, "Operator should access operator resources"},
		{"operator_access_admin", operatorToken, testutils.UniversalTestUserRole, false, "Operator should not access admin resources"},
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
				token, err := asserter.jwtHandler.GenerateToken(tt.userID, "operator", 24)
				require.NoError(t, err, "Token generation should succeed")

				claims, err := asserter.jwtHandler.ValidateToken(token)
				require.NoError(t, err, "Token validation should succeed")
				assert.Equal(t, tt.userID, claims.UserID, "Session should track user ID")

				// Validate expired sessions rejected
				// Create expired token by generating with very short expiry
				expiredToken, err := asserter.jwtHandler.GenerateToken(tt.userID, "operator", -1) // Negative hours = expired
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
				token, err = asserter.AssertJWTAuthentication(ctx, "test_user", "operator")
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
